package kugire

import (
	"strings"
	"testing"
)

const sampleXML = `<?xml version="1.0" encoding="UTF-8"?>
<TEI xmlns="http://www.tei-c.org/ns/1.0">
  <text>
    <body>
      <l n="1" xml:id="n1">
        <seg><w lemmaRef="#とし.年">年</w><w lemmaRef="#の.の">の</w><w lemmaRef="#うち.内">内</w><w lemmaRef="#に.に">に</w></seg>
        <seg><w lemmaRef="#はる.春">春</w><w lemmaRef="#は.は">は</w><w lemmaRef="#き.来">き</w></seg>
        <seg><w lemmaRef="#ひととせ.一年">ひとゝせ</w><w lemmaRef="#を.を">を</w></seg>
        <seg><w lemmaRef="#こぞ.去年">こそ</w><w lemmaRef="#と.と">と</w><w lemmaRef="#や.や">や</w></seg>
        <seg><w lemmaRef="#ことし.今年">ことし</w><w lemmaRef="#と.と">と</w><w lemmaRef="#や.や">や</w></seg>
      </l>
      <l n="2" xml:id="n2">
        <seg><w>袖</w><w>ひち</w><w>て</w></seg>
        <seg><w>むすひ</w><w>し</w></seg>
        <seg><w>水</w><w>の</w></seg>
        <seg><w>こほれ</w><w>る</w><w>を</w></seg>
        <seg><w>春</w><w>たつ</w><w>けふ</w></seg>
      </l>
    </body>
  </text>
</TEI>`

func TestLoadSegments(t *testing.T) {
	surfs, kanas, err := loadSegments(strings.NewReader(sampleXML), 1)
	if err != nil {
		t.Fatalf("loadSegments error: %v", err)
	}
	if len(surfs) != 5 {
		t.Fatalf("len(surfs): got %d, want 5", len(surfs))
	}
	if surfs[0] != "年の内に" {
		t.Errorf("surf 0: got %q, want %q", surfs[0], "年の内に")
	}
	if surfs[1] != "春はき" {
		t.Errorf("surf 1: got %q, want %q", surfs[1], "春はき")
	}
	if surfs[2] != "ひとゝせを" {
		t.Errorf("surf 2: got %q, want %q", surfs[2], "ひとゝせを")
	}
	// kana from lemmaRef: #とし.年 → とし
	if kanas[0] != "としのうちに" {
		t.Errorf("kana 0: got %q, want %q", kanas[0], "としのうちに")
	}
	// seg 2: <w lemmaRef="#ひととせ.一年">ひとゝせ</w><w lemmaRef="#を.を">を</w>
	// lemmaRef gives already-expanded kana: #ひととせ.一年 → ひととせ (no ゝ)
	if kanas[2] != "ひととせを" {
		t.Errorf("kana 2: got %q, want %q", kanas[2], "ひととせを")
	}
}

func TestLoadSegments_notFound(t *testing.T) {
	_, _, err := loadSegments(strings.NewReader(sampleXML), 99)
	if err == nil {
		t.Error("expected error for missing poem, got nil")
	}
}

func TestLoadSegments_poem2(t *testing.T) {
	surfs, _, err := loadSegments(strings.NewReader(sampleXML), 2)
	if err != nil {
		t.Fatalf("loadSegments error: %v", err)
	}
	if len(surfs) != 5 {
		t.Fatalf("len(surfs): got %d, want 5", len(surfs))
	}
	if surfs[0] != "袖ひちて" {
		t.Errorf("surf 0: got %q, want %q", surfs[0], "袖ひちて")
	}
}
