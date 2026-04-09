package kugire

import (
	"strings"
	"testing"
)

// ── RenderDraft ───────────────────────────────────────────────────

func TestRenderDraft(t *testing.T) {
	d := PoemData{
		ID: 1,
		Segments: []string{
			"年の内に",
			"春はきにけり",
			"ひとゝせを",
			"こそとやいはむ",
			"ことしとやいはむ",
		},
		Tokens: []Token{
			{Surface: "年", POS: "名"},
			{Surface: "の", POS: "格助"},
			{Surface: "内", POS: "名"},
			{Surface: "に", POS: "格助"},
			{Surface: "春", POS: "名"},
			{Surface: "は", POS: "係助"},
			{Surface: "き", POS: "カ変", Inflect: "用", Lemma: "来"},
			{Surface: "に", POS: "完", Inflect: "用", Lemma: "ぬ"},
			{Surface: "けり", POS: "過", Inflect: "終", Lemma: "けり"},
			{Surface: "一とせ", POS: "名", Reading: "ひととせ"},
			{Surface: "を", POS: "＊助"},
			{Surface: "こそ", POS: "名", Reading: "こぞ"},
			{Surface: "と", POS: "格助"},
			{Surface: "や", POS: "係助"},
			{Surface: "いは", POS: "ハ四", Inflect: "未", Lemma: "言ふ"},
			{Surface: "む", POS: "推", Inflect: "終体", Lemma: "む"},
			{Surface: "ことし", POS: "名"},
			{Surface: "と", POS: "格助"},
			{Surface: "や", POS: "係助"},
			{Surface: "いは", POS: "ハ四", Inflect: "未", Lemma: "言ふ"},
			{Surface: "む", POS: "推", Inflect: "終体", Lemma: "む"},
		},
		SegmentsKana: []string{
			"としのうちに",
			"はるはきにけり",
			"ひととせを",
			"こぞとやいはむ",
			"ことしとやいはむ",
		},
		Translations: map[string]string{
			"kaneko": "年内に思い掛けず春は来たことであるわ、さてはこの同じ一年の内の昨日までを、去年と言おうか、それとも今年と言おうか。",
		},
	}
	positions := []KugirePos{
		{AfterSeg: 1, Source: "morph"},
		{AfterSeg: 1, Source: "kaneko"},
		{AfterSeg: 2, Source: "kaneko"},
		{AfterSeg: 3, Source: "morph"},
		{AfterSeg: 3, Source: "kaneko"},
	}

	got := RenderDraft(d, positions)

	// header comment must be present
	if !strings.Contains(got, "# Poem 1") {
		t.Errorf("missing '# Poem 1' header:\n%s", got)
	}
	// translation comment must be present
	if !strings.Contains(got, "# kaneko:") {
		t.Errorf("missing kaneko translation comment:\n%s", got)
	}
	// morph hint for seg 1 (けり, 過-終, K1)
	if !strings.Contains(got, "けり") {
		t.Errorf("missing morph hint for けり:\n%s", got)
	}

	// content lines: extract non-comment lines
	var contentLines []string
	for _, l := range strings.Split(got, "\n") {
		if l != "" && !strings.HasPrefix(l, "#") {
			contentLines = append(contentLines, l)
		}
	}
	if len(contentLines) != 5 {
		t.Fatalf("content line count: got %d, want 5\n%s", len(contentLines), got)
	}
	if contentLines[0] != "年の内に" {
		t.Errorf("content line 0: got %q, want %q", contentLines[0], "年の内に")
	}
	if !strings.Contains(contentLines[1], "[K:morph]") {
		t.Errorf("content line 1 missing [K:morph]: %q", contentLines[1])
	}
	if !strings.Contains(contentLines[1], "[K:kaneko]") {
		t.Errorf("content line 1 missing [K:kaneko]: %q", contentLines[1])
	}
	if strings.Contains(contentLines[4], "[K:") {
		t.Errorf("last content line should have no tag: %q", contentLines[4])
	}
}

func TestRenderDraft_noPositions(t *testing.T) {
	d := PoemData{
		ID:       1,
		Segments: []string{"a", "b", "c", "d", "e"},
	}
	got := RenderDraft(d, nil)
	// count non-comment, non-empty lines
	var contentLines []string
	for _, l := range strings.Split(got, "\n") {
		if l != "" && !strings.HasPrefix(l, "#") {
			contentLines = append(contentLines, l)
		}
	}
	if len(contentLines) != 5 {
		t.Fatalf("content line count: got %d, want 5", len(contentLines))
	}
	for _, l := range contentLines {
		if strings.Contains(l, "[K:") {
			t.Errorf("unexpected tag in line %q", l)
		}
	}
}

// ── ParseDraft ────────────────────────────────────────────────────

func TestParseDraft(t *testing.T) {
	draft := "年の内に\n春はきにけり [K:morph] [K:kaneko]\nひとゝせを [K:kaneko]\nこそとやいはむ [K:morph]\nことしとやいはむ\n"

	positions, err := ParseDraft(draft)
	if err != nil {
		t.Fatalf("ParseDraft error: %v", err)
	}

	want := []KugirePos{
		{AfterSeg: 1, Source: "morph"},
		{AfterSeg: 1, Source: "kaneko"},
		{AfterSeg: 2, Source: "kaneko"},
		{AfterSeg: 3, Source: "morph"},
	}
	if len(positions) != len(want) {
		t.Fatalf("len: got %d, want %d\n  got: %v", len(positions), len(want), positions)
	}
	for i, p := range positions {
		if p != want[i] {
			t.Errorf("positions[%d]: got %+v, want %+v", i, p, want[i])
		}
	}
}

func TestParseDraft_lastLineTagIgnored(t *testing.T) {
	// Tag on last line (seg 4) must be silently ignored — no boundary after final seg
	draft := "a\nb\nc\nd\ne [K:morph]\n"
	positions, err := ParseDraft(draft)
	if err != nil {
		t.Fatalf("ParseDraft error: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("expected no positions, got %v", positions)
	}
}

func TestParseDraft_emptyAndCommentLines(t *testing.T) {
	// Empty lines and # comments are skipped; content lines are counted
	draft := "# poem 1\n年の内に\n春はきにけり [K:morph]\n\nひとゝせを\nこそとやいはむ\nことしとやいはむ\n"
	positions, err := ParseDraft(draft)
	if err != nil {
		t.Fatalf("ParseDraft error: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("len: got %d, want 1", len(positions))
	}
	if positions[0].AfterSeg != 1 {
		t.Errorf("AfterSeg: got %d, want 1", positions[0].AfterSeg)
	}
}

func TestParseDraft_roundtrip(t *testing.T) {
	d := PoemData{
		ID: 1,
		Segments: []string{
			"年の内に",
			"春はきにけり",
			"ひとゝせを",
			"こそとやいはむ",
			"ことしとやいはむ",
		},
	}
	original := []KugirePos{
		{AfterSeg: 1, Source: "morph"},
		{AfterSeg: 2, Source: "kaneko"},
		{AfterSeg: 3, Source: "morph"},
	}

	draft := RenderDraft(d, original)
	parsed, err := ParseDraft(draft)
	if err != nil {
		t.Fatalf("ParseDraft error: %v", err)
	}
	if len(parsed) != len(original) {
		t.Fatalf("roundtrip len: got %d, want %d", len(parsed), len(original))
	}
	for i, p := range parsed {
		if p != original[i] {
			t.Errorf("position[%d]: got %+v, want %+v", i, p, original[i])
		}
	}
}
