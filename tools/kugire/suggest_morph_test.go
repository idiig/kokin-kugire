package kugire

import "testing"

// ── classifyToken ────────────────────────────────────────────────

func TestClassifyToken(t *testing.T) {
	tests := []struct {
		tok  Token
		want string
	}{
		// K1: terminal inflection forms
		{Token{POS: "過", Inflect: "終", Lemma: "けり"}, "K1"},
		{Token{POS: "推", Inflect: "終体", Lemma: "む"}, "K1"},
		{Token{POS: "タ四", Inflect: "已", Lemma: "立つ"}, "K1"},
		// K1: final/binding particles
		{Token{POS: "終助"}, "K1"},
		{Token{POS: "係助"}, "K1"},
		// K1: imperative
		{Token{POS: "マ上一", Inflect: "命"}, "K1"},
		// K0: case particle
		{Token{POS: "格助"}, "K0"},
		// K0: conjunctive particle
		{Token{POS: "接助"}, "K0"},
		// K0: conjunctive particle with subtype
		{Token{POS: "接助", Subtype: "反復"}, "K0"},
		// K0: continuative form
		{Token{POS: "カ変", Inflect: "用", Lemma: "来"}, "K0"},
		// K0: attributive form (not terminal/attributive)
		{Token{POS: "完", Inflect: "体", Lemma: "り"}, "K0"},
		// K2: noun (体言止め)
		{Token{POS: "名"}, "K2"},
		{Token{POS: "名", Subtype: "地名"}, "K2"},
		// K2: default (pronoun, adverb, etc.)
		{Token{POS: "代"}, "K2"},
		{Token{POS: "副"}, "K2"},
	}

	for _, tt := range tests {
		got := classifyToken(tt.tok)
		if got != tt.want {
			t.Errorf("classifyToken(%+v) = %q, want %q", tt.tok, got, tt.want)
		}
	}
}

// ── alignTokensToSegments ─────────────────────────────────────────

func TestAlignTokensToSegments(t *testing.T) {
	// Alignment uses kana readings (token.Reading) matched against
	// normalized segment kana (ゝ/ゞ expanded, spaces removed).
	// Segment "ひとゝせを" normalizes to "ひととせを"; token reading "ひととせ"+"を" = "ひととせを".
	_ = []string{
		"年の内に",
		"春はきにけり",
		"ひとゝせを", // ゝ → repeated char: ひとゝせ = ひとと + せ = ひとと (wait: ひ-と-ゝ-せ, ゝ repeats と → ひととせ)
		"こそとやいはむ",
		"ことしとやいはむ",
	}
	tokens := []Token{
		{Surface: "年", Reading: "とし"},
		{Surface: "の", Reading: "の"},
		{Surface: "内", Reading: "うち"},
		{Surface: "に", Reading: "に"},
		{Surface: "春", Reading: "はる"},
		{Surface: "は", Reading: "は"},
		{Surface: "き", Reading: "き"},
		{Surface: "に", Reading: "に"},
		{Surface: "けり", Reading: "けり"},
		{Surface: "一とせ", Reading: "ひととせ"}, // surface differs from XML; use reading
		{Surface: "を", Reading: "を"},
		{Surface: "こそ", Reading: "こぞ"},
		{Surface: "と", Reading: "と"},
		{Surface: "や", Reading: "や"},
		{Surface: "いは", Reading: "いは"},
		{Surface: "む", Reading: "む"},
		{Surface: "ことし", Reading: "ことし"},
		{Surface: "と", Reading: "と"},
		{Surface: "や", Reading: "や"},
		{Surface: "いは", Reading: "いは"},
		{Surface: "む", Reading: "む"},
	}

	// Pass segment kana (normalizeKana applied inside); ゝ in seg 2 expands to と
	segKanas := []string{
		"としのうちに",
		"はるはきにけり",
		"ひとゝせを", // normalizes to ひとゝせを → ひととせを
		"こぞとやいはむ",
		"ことしとやいはむ",
	}
	groups, err := alignTokensToSegments(tokens, segKanas)
	if err != nil {
		t.Fatalf("alignTokensToSegments error: %v", err)
	}
	if len(groups) != 5 {
		t.Fatalf("len(groups): got %d, want 5", len(groups))
	}
	if len(groups[0]) != 4 {
		t.Errorf("seg 0 tokens: got %d, want 4", len(groups[0]))
	}
	if len(groups[1]) != 5 {
		t.Errorf("seg 1 tokens: got %d, want 5", len(groups[1]))
	}
	// seg 2: 一とせ + を → 2 tokens (surface ≠ seg, but kana matches)
	if len(groups[2]) != 2 {
		t.Errorf("seg 2 tokens: got %d, want 2", len(groups[2]))
	}
	last := groups[4][len(groups[4])-1]
	if last.Surface != "む" {
		t.Errorf("seg 4 last surface: got %q, want %q", last.Surface, "む")
	}
}

func TestAlignTokensToSegments_mismatch(t *testing.T) {
	segs := []string{"abc"}
	tokens := []Token{{Surface: "x"}, {Surface: "y"}}
	_, err := alignTokensToSegments(tokens, segs)
	if err == nil {
		t.Error("expected error for surface mismatch, got nil")
	}
}

// ── SuggestMorph ─────────────────────────────────────────────────

func TestSuggestMorph(t *testing.T) {
	// Poem 1: 年の内に / 春はきにけり / ひとゝせを / こそとやいはむ / ことしとやいはむ
	// Segment-final tokens:
	//   seg 0: に  (格助) → K0  no kugire
	//   seg 1: けり (過-終) → K1  kugire
	//   seg 2: を  (＊助=格助) → K0  no kugire
	//   seg 3: む  (推-終体) → K1  kugire
	//   seg 4: む  (推-終体) → K1  (poem end, no boundary after)
	d := PoemData{
		ID: 1,
		Segments: []string{
			"年の内に",
			"春はきにけり",
			"ひとゝせを",
			"こそとやいはむ",
			"ことしとやいはむ",
		},
		SegmentsKana: []string{
			"としのうちに",
			"はるはきにけり",
			"ひととせを", // lemmaRef gives expanded kana (no ゝ)
			"こぞとやいはむ",
			"ことしとやいはむ",
		},
		Tokens: []Token{
			{Surface: "年", POS: "名", Reading: "とし"},
			{Surface: "の", POS: "格助", Reading: "の"},
			{Surface: "内", POS: "名", Reading: "うち"},
			{Surface: "に", POS: "格助", Reading: "に"},
			{Surface: "春", POS: "名", Reading: "はる"},
			{Surface: "は", POS: "係助", Reading: "は"},
			{Surface: "き", POS: "カ変", Inflect: "用", Lemma: "来", Reading: "き"},
			{Surface: "に", POS: "完", Inflect: "用", Lemma: "ぬ", Reading: "に"},
			{Surface: "けり", POS: "過", Inflect: "終", Lemma: "けり", Reading: "けり"},
			{Surface: "一とせ", POS: "名", Reading: "ひととせ"},
			{Surface: "を", POS: "＊助", Reading: "を"},
			{Surface: "こそ", POS: "名", Reading: "こぞ"},
			{Surface: "と", POS: "格助", Reading: "と"},
			{Surface: "や", POS: "係助", Reading: "や"},
			{Surface: "いは", POS: "ハ四", Inflect: "未", Lemma: "言ふ", Reading: "いは"},
			{Surface: "む", POS: "推", Inflect: "終体", Lemma: "む", Reading: "む"},
			{Surface: "ことし", POS: "名", Reading: "ことし"},
			{Surface: "と", POS: "格助", Reading: "と"},
			{Surface: "や", POS: "係助", Reading: "や"},
			{Surface: "いは", POS: "ハ四", Inflect: "未", Lemma: "言ふ", Reading: "いは"},
			{Surface: "む", POS: "推", Inflect: "終体", Lemma: "む", Reading: "む"},
		},
	}

	positions := SuggestMorph(d)

	// expect kugire after seg 1 (けり K1) and seg 3 (む K1)
	// seg 0 (に K0), seg 2 (を K0) → no kugire
	// seg 4 is end of poem — no position after it
	if len(positions) != 2 {
		t.Fatalf("len(positions): got %d, want 2 (after segs 1 and 3)", len(positions))
	}
	if positions[0].AfterSeg != 1 {
		t.Errorf("positions[0].AfterSeg: got %d, want 1", positions[0].AfterSeg)
	}
	if positions[1].AfterSeg != 3 {
		t.Errorf("positions[1].AfterSeg: got %d, want 3", positions[1].AfterSeg)
	}
	for _, p := range positions {
		if p.Source != "morph" {
			t.Errorf("Source: got %q, want %q", p.Source, "morph")
		}
	}
}
