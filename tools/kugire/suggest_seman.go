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

1. 対応関係: 句1〜2が「掬って遊んだ水が」、句3が「凍っているのを」、句4〜5が「風が吹き解かすだろうか」に対応。
2. 訳文の切れ目: 読点はあるが、全体が「水が…凍っているのを…解かすだろうか」という一続きの問いかけ。意味の一段落なし。
3. 句切れ位置: なし。

{"positions": []}

--- 例（二句切れ）---
和歌の句:
1: 雪の内に
2: 春はきにけり
3: うくひすの
4: こほれるなみた
5: 今やとくらむ

現代語訳: まだ雪がある内なのに、如何にも長閑な春は来たことであるわ、さては雪は勿論鴬の凍っている涙も、今は解けることだろうか。

1. 対応関係: 句1〜2が「雪がある内なのに春は来たことであるわ」、句3〜5が「鴬の涙も今は解けることだろうか」に対応。
2. 訳文の切れ目: 「来たことであるわ」で感嘆的に一段落し、「さては」以降が別の意味展開。
3. 句切れ位置: 句2の後。

{"positions": [2]}

--- 例（四句切れ）---
和歌の句:
1: 霞たち
2: このめもはるの
3: 雪ふれは
4: 花なきさとも
5: 花そちりける

現代語訳: 霞も立ち木の芽も萌え出す時も時とて、泡雪がちらちら降ると、これは奇妙、花のないこの里さえも花が散ったわ。

1. 対応関係: 句1〜2が「霞も立ち木の芽も萌え出す時」、句3が「泡雪が降ると」、句4が「花のないこの里さえも」、句5が「花が散ったわ」に対応。
2. 訳文の切れ目: 「これは奇妙」という驚きの感嘆の後、句5が結句として独立。切れ目は句4の後。
3. 句切れ位置: 句4の後。

{"positions": [4]}
`

// buildSemanPrompt constructs the Ollama prompt for kugire suggestion.
func buildSemanPrompt(d PoemData, translatorCode string) string {
	translation := d.Translations[translatorCode]
	var sb strings.Builder
	sb.WriteString("和歌の句切れを判定するために、次の手順に従ってください。\n")
	sb.WriteString("1. 対応関係の確認: 和歌の各句（1〜5）が現代語訳のどの部分に対応するかを確認する。\n")
	sb.WriteString("2. 訳文の切れ目の判定: 現代語訳において意味が一段落する切れ目（文の終わり・感嘆・意味の転換）を見つける。\n")
	sb.WriteString("3. 句切れ位置への対応付け: その切れ目が和歌の何番目の句の後に対応するかを特定する。\n")
	sb.WriteString("最後の句（5番）の後は含めないこと。JSON形式 {\"positions\": [...]} で答えてください。\n\n")
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
