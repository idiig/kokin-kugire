package kugire

import "io"

// Token is a single morphologically analysed word from morphological-annotation.txt.
type Token struct {
	Surface string // original surface form, e.g. "き"
	POS     string // part of speech, e.g. "カ変", "名", "格助"
	Subtype string // non-inflecting subtype, e.g. "地名", "反復" (empty if inflecting)
	Inflect string // inflection form, e.g. "用", "終", "体", "未", "已" (empty if none)
	Lemma   string // dictionary lemma, e.g. "来" (empty if not present)
	Reading string // kana reading, e.g. "き"
}

// KugirePos marks a kugire position after a given segment index.
type KugirePos struct {
	// AfterSeg is the 0-based index of the segment after which the kugire falls.
	// A kugire after segment 1 (the 2nd line) has AfterSeg=1.
	AfterSeg int
	Source   string // "morph" | translator code, e.g. "kaneko"
	Cert     string // certainty: "high" | "mid" | "low" (empty = unspecified)
}

// TranslationSource pairs a translator code with its file path and loader function.
// Adding a new translator requires only registering a TranslationSource — no
// changes to LoadPoem or the suggestion pipeline.
type TranslationSource struct {
	Code string                                    // e.g. "kaneko", "kubota"
	Path string                                    // path to the translation file
	Load func(r io.Reader) (map[int]string, error) // format-specific parser
}

// PoemData holds all data for one poem needed by the kugire pipeline.
type PoemData struct {
	ID           int
	Segments     []string          // surface text of each of the 5 segments (from XML)
	SegmentsKana []string          // kana reading of each segment (from <w> lemmaRef)
	Tokens       []Token           // flat morphological token list
	Translations map[string]string // translator_code -> modern Japanese prose
}
