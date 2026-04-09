package kugire

import (
	"fmt"
	"os"
	"testing"
)

// TestShowPoem1 prints a human-readable view of poem 1 from real data files.
// Run with: go test -v -run TestShowPoem1
func TestShowPoem1(t *testing.T) {
	const id = 1

	// Segments from XML
	xmlFile, err := os.Open("../../data/kokinwakashu.xml")
	if err != nil {
		t.Skipf("data file not available: %v", err)
	}
	segs, kanas, err := loadSegments(xmlFile, id)
	xmlFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Tokens from morphological annotation
	morphFile, err := os.Open("../../data/morphological-annotation.txt")
	if err != nil {
		t.Skipf("data file not available: %v", err)
	}
	morphData, err := loadMorphData(morphFile)
	morphFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Translation from Kaneko
	kanekoFile, err := os.Open("../../data/translation-kaneko.txt")
	if err != nil {
		t.Skipf("data file not available: %v", err)
	}
	kanekoData, err := loadKanekoData(kanekoFile)
	kanekoFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("\n=== Poem %d ===\n\n", id)

	fmt.Println("[ Segments ]")
	for i, s := range segs {
		fmt.Printf("  %d  %s\n", i+1, s)
	}

	fmt.Println("\n[ Morphological tokens ]")
	fmt.Printf("  %-12s  %-8s  %-6s  %s\n", "Surface", "POS", "Inflect", "Lemma")
	fmt.Println("  " + "─────────────────────────────────────")
	for _, tok := range morphData[id] {
		fmt.Printf("  %-12s  %-8s  %-6s  %s\n", tok.Surface, tok.POS, tok.Inflect, tok.Lemma)
	}

	fmt.Println("\n[ Translation (Kaneko) ]")
	fmt.Printf("  %s\n", kanekoData[id])

	// SuggestMorph
	d := PoemData{
		ID:           id,
		Segments:     segs,
		SegmentsKana: kanas,
		Tokens:       morphData[id],
		Translations: map[string]string{"kaneko": kanekoData[id]},
	}
	positions := SuggestMorph(d)
	semanPositions, err := SuggestSeman(d, "kaneko")
	if err != nil {
		t.Logf("SuggestSeman: %v (Ollama not available?)", err)
	}

	allPositions := append(positions, semanPositions...)
	draft := RenderDraft(d, allPositions)
	fmt.Println("\n[ Draft (as written to /tmp/kokin-kugire-1.txt) ]")
	fmt.Println(draft)
}
