package kugire

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/beevik/etree"
)

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
}

func kTagsFor(t *testing.T, path string, id int) []string {
	t.Helper()
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(path); err != nil {
		t.Fatalf("read xml: %v", err)
	}
	l := doc.FindElement(fmt.Sprintf("//l[@n='%d']", id))
	if l == nil {
		t.Fatalf("poem %d not found", id)
	}
	var tags []string
	for _, k := range l.SelectElements("k") {
		n := k.SelectAttrValue("n", "")
		src := k.SelectAttrValue("source", "")
		tags = append(tags, n+":"+src)
	}
	return tags
}

// ── AnnotateXML ────────────────────────────────────────────────────

func TestAnnotateXML_writesKTags(t *testing.T) {
	src := "../../data/kokinwakashu.xml"
	if _, err := os.Stat(src); err != nil {
		t.Skip("data file not available")
	}

	tmp := filepath.Join(t.TempDir(), "kokin-kugire.xml")
	copyFile(t, src, tmp)

	positions := []KugirePos{
		{AfterSeg: 1, Source: "morph"},
		{AfterSeg: 1, Source: "kaneko"},
		{AfterSeg: 3, Source: "morph"},
	}
	if err := AnnotateXML(tmp, 1, positions); err != nil {
		t.Fatalf("AnnotateXML: %v", err)
	}

	got := kTagsFor(t, tmp, 1)
	want := []string{"2:morph", "2:kaneko", "4:morph"}
	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d\n  got: %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAnnotateXML_replacesExisting(t *testing.T) {
	src := "../../data/kokinwakashu.xml"
	if _, err := os.Stat(src); err != nil {
		t.Skip("data file not available")
	}

	tmp := filepath.Join(t.TempDir(), "kokin-kugire.xml")
	copyFile(t, src, tmp)

	// First annotation
	if err := AnnotateXML(tmp, 1, []KugirePos{{AfterSeg: 1, Source: "morph"}}); err != nil {
		t.Fatalf("first AnnotateXML: %v", err)
	}
	// Second annotation — replaces first
	if err := AnnotateXML(tmp, 1, []KugirePos{{AfterSeg: 3, Source: "kaneko"}}); err != nil {
		t.Fatalf("second AnnotateXML: %v", err)
	}

	got := kTagsFor(t, tmp, 1)
	want := []string{"4:kaneko"}
	if len(got) != len(want) {
		t.Fatalf("len: got %d, want %d\n  got: %v", len(got), len(want), got)
	}
	if got[0] != want[0] {
		t.Errorf("got %q, want %q", got[0], want[0])
	}
}

func TestAnnotateXML_emptyPositions(t *testing.T) {
	src := "../../data/kokinwakashu.xml"
	if _, err := os.Stat(src); err != nil {
		t.Skip("data file not available")
	}

	tmp := filepath.Join(t.TempDir(), "kokin-kugire.xml")
	copyFile(t, src, tmp)

	// Add then clear
	if err := AnnotateXML(tmp, 1, []KugirePos{{AfterSeg: 1, Source: "morph"}}); err != nil {
		t.Fatalf("AnnotateXML: %v", err)
	}
	if err := AnnotateXML(tmp, 1, nil); err != nil {
		t.Fatalf("AnnotateXML clear: %v", err)
	}

	got := kTagsFor(t, tmp, 1)
	if len(got) != 0 {
		t.Errorf("expected no <k> tags, got %v", got)
	}
}

func TestAnnotateXML_preservesOtherPoems(t *testing.T) {
	src := "../../data/kokinwakashu.xml"
	if _, err := os.Stat(src); err != nil {
		t.Skip("data file not available")
	}

	tmp := filepath.Join(t.TempDir(), "kokin-kugire.xml")
	copyFile(t, src, tmp)

	// Annotate poem 1 and poem 2 separately
	if err := AnnotateXML(tmp, 1, []KugirePos{{AfterSeg: 1, Source: "morph"}}); err != nil {
		t.Fatalf("poem 1: %v", err)
	}
	if err := AnnotateXML(tmp, 2, []KugirePos{{AfterSeg: 2, Source: "kaneko"}}); err != nil {
		t.Fatalf("poem 2: %v", err)
	}
	// Re-annotate poem 1 — poem 2 must be untouched
	if err := AnnotateXML(tmp, 1, []KugirePos{{AfterSeg: 3, Source: "morph"}}); err != nil {
		t.Fatalf("poem 1 re-annotate: %v", err)
	}

	if got := kTagsFor(t, tmp, 1); len(got) != 1 || got[0] != "4:morph" {
		t.Errorf("poem 1: got %v, want [4:morph]", got)
	}
	if got := kTagsFor(t, tmp, 2); len(got) != 1 || got[0] != "3:kaneko" {
		t.Errorf("poem 2: got %v, want [3:kaneko]", got)
	}
}
