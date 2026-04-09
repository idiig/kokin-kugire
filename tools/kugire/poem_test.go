package kugire

import (
	"os"
	"strings"
	"testing"
)

// ── KanekoSource ───────────────────────────────────────────────────

func TestKanekoSource(t *testing.T) {
	src := KanekoSource("/some/path/translation-kaneko.txt")
	if src.Code != "kaneko" {
		t.Errorf("Code = %q, want %q", src.Code, "kaneko")
	}
	if src.Path != "/some/path/translation-kaneko.txt" {
		t.Errorf("Path = %q", src.Path)
	}
	if src.Load == nil {
		t.Error("Load is nil")
	}
}

func TestKanekoSource_Load(t *testing.T) {
	src := KanekoSource("")
	input := "$$DATA|\n$A|00001\n$D|テスト訳文\n"
	data, err := src.Load(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data[1] != "テスト訳文" {
		t.Errorf("got %q, want %q", data[1], "テスト訳文")
	}
}

// ── LoadPoem ───────────────────────────────────────────────────────

func TestLoadPoem_realData(t *testing.T) {
	xmlPath := "../../data/kokinwakashu.xml"
	morphPath := "../../data/morphological-annotation.txt"
	kanekoPath := "../../data/translation-kaneko.txt"

	for _, p := range []string{xmlPath, morphPath, kanekoPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("data file not available: %s", p)
		}
	}

	d, err := LoadPoem(xmlPath, morphPath, 1, []TranslationSource{
		KanekoSource(kanekoPath),
	})
	if err != nil {
		t.Fatalf("LoadPoem: %v", err)
	}

	if d.ID != 1 {
		t.Errorf("ID = %d, want 1", d.ID)
	}
	if len(d.Segments) != 5 {
		t.Errorf("len(Segments) = %d, want 5", len(d.Segments))
	}
	if len(d.SegmentsKana) != 5 {
		t.Errorf("len(SegmentsKana) = %d, want 5", len(d.SegmentsKana))
	}
	if len(d.Tokens) == 0 {
		t.Error("Tokens is empty")
	}
	if d.Translations["kaneko"] == "" {
		t.Error("Translations[kaneko] is empty")
	}
}

func TestLoadPoem_multipleTranslators(t *testing.T) {
	xmlPath := "../../data/kokinwakashu.xml"
	morphPath := "../../data/morphological-annotation.txt"
	kanekoPath := "../../data/translation-kaneko.txt"

	for _, p := range []string{xmlPath, morphPath, kanekoPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("data file not available: %s", p)
		}
	}

	// Register kaneko twice under different codes to simulate two translators.
	d, err := LoadPoem(xmlPath, morphPath, 1, []TranslationSource{
		KanekoSource(kanekoPath),
		{Code: "kaneko2", Path: kanekoPath, Load: loadKanekoData},
	})
	if err != nil {
		t.Fatalf("LoadPoem: %v", err)
	}
	if d.Translations["kaneko"] == "" {
		t.Error("Translations[kaneko] missing")
	}
	if d.Translations["kaneko2"] == "" {
		t.Error("Translations[kaneko2] missing")
	}
}

func TestLoadPoem_missingTranslationFile(t *testing.T) {
	xmlPath := "../../data/kokinwakashu.xml"
	morphPath := "../../data/morphological-annotation.txt"

	for _, p := range []string{xmlPath, morphPath} {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("data file not available: %s", p)
		}
	}

	// Missing translation file should return an error.
	_, err := LoadPoem(xmlPath, morphPath, 1, []TranslationSource{
		KanekoSource("/nonexistent/path.txt"),
	})
	if err == nil {
		t.Fatal("expected error for missing translation file")
	}
}
