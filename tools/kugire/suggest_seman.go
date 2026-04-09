package kugire

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ollamaEndpoint is the Ollama API URL. Override in tests.
var ollamaEndpoint = "http://localhost:11434/api/generate"

const ollamaModel = "qwen2.5"

// semanFewShot is the few-shot block included in every prompt.
// Examples are drawn from real Kokinwakashu poems with Kaneko's translation.
const semanFewShot = `--- 例（句切れなし）---
和歌の句:
1: 袖ひちて
2: むすひし水の
3: こほれるを
4: 春立けふの
5: 風やとくらむ

現代語訳: 夏の暑い頃袖がびしょ濡れになるまで掬って遊んだ水が、この間の冬の寒さに凍っているのを、立春の今日の長閑な風が、再び元の水に吹き解かすだろうか。

訳は読点で区切られているが意味は一続き。句切れなし。

{"positions": []}

--- 例（二句切れ）---
和歌の句:
1: 雪の内に
2: 春はきにけり
3: うくひすの
4: こほれるなみた
5: 今やとくらむ

現代語訳: まだ雪がある内なのに、如何にも長閑な春は来たことであるわ、さては雪は勿論鴬の凍っている涙も、今は解けることだろうか。

「春は来たことであるわ」で感嘆的に切れ、意味が一段落する。

{"positions": [2]}

--- 例（四句切れ）---
和歌の句:
1: 霞たち
2: このめもはるの
3: 雪ふれは
4: 花なきさとも
5: 花そちりける

現代語訳: 霞も立ち木の芽も萌え出す時も時とて、泡雪がちらちら降ると、これは奇妙、花のないこの里さえも花が散ったわ。

「これは奇妙」という驚きの間（ま）の後に第5句が結句として切り立つ。

{"positions": [4]}
`

// buildSemanPrompt constructs the Ollama prompt for kugire suggestion.
func buildSemanPrompt(d PoemData, translatorCode string) string {
	translation := d.Translations[translatorCode]
	var sb strings.Builder
	sb.WriteString("和歌の句と現代語訳を見て、訳文の意味の区切りが何番目の句の後に対応するかを判断してください。\n")
	sb.WriteString("最後の句の後は含めないこと。JSON形式 {\"positions\": [...]} で答えてください。\n\n")
	sb.WriteString(semanFewShot)
	sb.WriteString("--- 問題 ---\n")
	sb.WriteString("和歌の句:\n")
	for i, seg := range d.Segments {
		fmt.Fprintf(&sb, "%d: %s\n", i+1, seg)
	}
	fmt.Fprintf(&sb, "\n現代語訳: %s\n\n", translation)
	return sb.String()
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

type semanResult struct {
	Positions []int `json:"positions"`
}

// parseSemanResponse parses the JSON returned by the LLM.
func parseSemanResponse(raw string) ([]int, error) {
	var result semanResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}
	return result.Positions, nil
}

// SuggestSeman returns kugire positions inferred from an expert translation
// using a local Ollama LLM.
func SuggestSeman(d PoemData, translatorCode string) ([]KugirePos, error) {
	translation, ok := d.Translations[translatorCode]
	if !ok || translation == "" {
		return nil, nil
	}

	prompt := buildSemanPrompt(d, translatorCode)

	reqBody, err := json.Marshal(ollamaRequest{
		Model:  ollamaModel,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(ollamaEndpoint, "application/json", bytes.NewReader(reqBody)) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ollama response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("parse ollama response: %w", err)
	}

	positions, err := parseSemanResponse(ollamaResp.Response)
	if err != nil {
		return nil, err
	}

	n := len(d.Segments)
	var result []KugirePos
	for _, pos := range positions {
		afterSeg := pos - 1 // 1-based → 0-based
		if afterSeg >= 0 && afterSeg < n-1 {
			result = append(result, KugirePos{AfterSeg: afterSeg, Source: translatorCode})
		}
	}
	return result, nil
}
