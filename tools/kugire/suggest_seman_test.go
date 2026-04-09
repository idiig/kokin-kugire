package kugire

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── buildSemanPrompt ───────────────────────────────────────────────

func TestBuildSemanPrompt_containsSegments(t *testing.T) {
	d := PoemData{
		ID: 1,
		Segments: []string{
			"年の内に", "春はきにけり", "ひとゝせを", "こそとやいはむ", "ことしとやいはむ",
		},
		Translations: map[string]string{
			"kaneko": "年内に思い掛けず春は来たことであるわ、さてはこの同じ一年の内の昨日までを、去年と言おうか、それとも今年と言おうか。",
		},
	}
	got := buildSemanPrompt(d, "kaneko")

	for _, seg := range d.Segments {
		if !strings.Contains(got, seg) {
			t.Errorf("prompt missing segment %q", seg)
		}
	}
	if !strings.Contains(got, d.Translations["kaneko"]) {
		t.Errorf("prompt missing translation")
	}
}

func TestBuildSemanPrompt_containsFewShot(t *testing.T) {
	d := PoemData{
		ID:       1,
		Segments: make([]string, 5),
		Translations: map[string]string{
			"kaneko": "test translation",
		},
	}
	got := buildSemanPrompt(d, "kaneko")
	// few-shot block must be present
	if !strings.Contains(got, "例") {
		t.Errorf("prompt missing few-shot block")
	}
	if !strings.Contains(got, "positions") {
		t.Errorf("prompt missing JSON format hint")
	}
}

// ── parseSemanResponse ─────────────────────────────────────────────

func TestParseSemanResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"basic", `{"positions": [1, 2, 3]}`, []int{1, 2, 3}, false},
		{"single", `{"positions": [2]}`, []int{2}, false},
		{"empty list", `{"positions": []}`, []int{}, false},
		{"invalid json", `not json`, nil, true},
		{"wrong key", `{"result": [1]}`, []int{}, false}, // unknown key → empty
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSemanResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ── SuggestSeman (mock Ollama) ─────────────────────────────────────

func TestSuggestSeman_mockOllama(t *testing.T) {
	// Mock Ollama returns positions [2, 4] (1-based)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"response": `{"positions": [2, 4]}`}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	orig := ollamaEndpoint
	ollamaEndpoint = server.URL + "/api/generate"
	defer func() { ollamaEndpoint = orig }()

	d := PoemData{
		ID:       1,
		Segments: make([]string, 5),
		Translations: map[string]string{
			"kaneko": "テスト訳文。別の節。",
		},
	}
	positions, err := SuggestSeman(d, "kaneko")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1-based [2,4] → 0-based AfterSeg [1,3]
	want := []KugirePos{
		{AfterSeg: 1, Source: "kaneko"},
		{AfterSeg: 3, Source: "kaneko"},
	}
	if len(positions) != len(want) {
		t.Fatalf("len: got %d, want %d\n  got: %v", len(positions), len(want), positions)
	}
	for i, p := range positions {
		if p != want[i] {
			t.Errorf("[%d] got %+v, want %+v", i, p, want[i])
		}
	}
}

func TestSuggestSeman_lastSegIgnored(t *testing.T) {
	// LLM returns 5 (last seg) — must be filtered out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"response": `{"positions": [2, 5]}`}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	orig := ollamaEndpoint
	ollamaEndpoint = server.URL + "/api/generate"
	defer func() { ollamaEndpoint = orig }()

	d := PoemData{
		ID:       1,
		Segments: make([]string, 5),
		Translations: map[string]string{
			"kaneko": "テスト。",
		},
	}
	positions, err := SuggestSeman(d, "kaneko")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(positions) != 1 || positions[0].AfterSeg != 1 {
		t.Errorf("got %v, want [{AfterSeg:1 Source:kaneko}]", positions)
	}
}

func TestSuggestSeman_missingTranslation(t *testing.T) {
	d := PoemData{
		ID:           1,
		Segments:     make([]string, 5),
		Translations: map[string]string{},
	}
	positions, err := SuggestSeman(d, "kaneko")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if positions != nil {
		t.Errorf("expected nil, got %v", positions)
	}
}

func TestSuggestSeman_ollamaDown(t *testing.T) {
	orig := ollamaEndpoint
	ollamaEndpoint = "http://127.0.0.1:1" // nothing listening
	defer func() { ollamaEndpoint = orig }()

	d := PoemData{
		ID:       1,
		Segments: make([]string, 5),
		Translations: map[string]string{
			"kaneko": "テスト。",
		},
	}
	_, err := SuggestSeman(d, "kaneko")
	if err == nil {
		t.Fatal("expected error when Ollama is down")
	}
}
