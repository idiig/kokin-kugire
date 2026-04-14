package kugire

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ollamaEndpoint is the Ollama API URL. Override in tests.
var ollamaEndpoint = "http://localhost:11434/api/generate"

const ollamaModel = "qwen2.5:32b"

//go:embed prompt-seman.txt
var semanPromptPrefix string

// buildSemanPrompt constructs the Ollama prompt for kugire suggestion.
func buildSemanPrompt(d PoemData, translatorCode string) string {
	translation := d.Translations[translatorCode]
	var sb strings.Builder
	sb.WriteString(semanPromptPrefix)
	sb.WriteString("和歌の句:\n")
	for i, seg := range d.Segments {
		fmt.Fprintf(&sb, "%d: %s\n", i+1, seg)
	}
	fmt.Fprintf(&sb, "\n現代語訳: %s\n\n", translation)
	return sb.String()
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature"`
}

type ollamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Format  string        `json:"format"`
	Options ollamaOptions `json:"options"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

type semanResult struct {
	Alignment string `json:"alignment"`
	Breaks    string `json:"breaks"`
	Positions []int  `json:"positions"`
}

// parseSemanResponse parses the JSON returned by the LLM.
func parseSemanResponse(raw string) (semanResult, error) {
	var result semanResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return semanResult{}, fmt.Errorf("parse LLM response: %w", err)
	}
	return result, nil
}

// SuggestSeman returns kugire positions and chain-of-thought reasoning inferred
// from an expert translation using a local Ollama LLM.
func SuggestSeman(d PoemData, translatorCode string) ([]KugirePos, SemanReasoning, error) {
	translation, ok := d.Translations[translatorCode]
	if !ok || translation == "" {
		return nil, SemanReasoning{}, nil
	}

	prompt := buildSemanPrompt(d, translatorCode)

	reqBody, err := json.Marshal(ollamaRequest{
		Model:   ollamaModel,
		Prompt:  prompt,
		Stream:  false,
		Format:  "json",
		Options: ollamaOptions{Temperature: 0},
	})
	if err != nil {
		return nil, SemanReasoning{}, err
	}

	resp, err := http.Post(ollamaEndpoint, "application/json", bytes.NewReader(reqBody)) //nolint:noctx
	if err != nil {
		return nil, SemanReasoning{}, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, SemanReasoning{}, fmt.Errorf("read ollama response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, SemanReasoning{}, fmt.Errorf("parse ollama response: %w", err)
	}

	parsed, err := parseSemanResponse(ollamaResp.Response)
	if err != nil {
		return nil, SemanReasoning{}, err
	}

	reasoning := SemanReasoning{Alignment: parsed.Alignment, Breaks: parsed.Breaks}

	n := len(d.Segments)
	var positions []KugirePos
	for _, pos := range parsed.Positions {
		afterSeg := pos - 1 // 1-based → 0-based
		if afterSeg >= 0 && afterSeg < n-1 {
			positions = append(positions, KugirePos{AfterSeg: afterSeg, Source: translatorCode})
			break // only one position
		}
	}
	return positions, reasoning, nil
}
