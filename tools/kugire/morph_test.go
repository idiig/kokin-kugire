package kugire

import (
	"strings"
	"testing"
)

func TestParseToken(t *testing.T) {
	tests := []struct {
		raw  string
		want Token
	}{
		{
			// uninflected noun
			raw:  "年/名/とし",
			want: Token{Surface: "年", POS: "名", Reading: "とし"},
		},
		{
			// irregular verb in continuative form
			raw:  "き/カ変-用:来:く/き",
			want: Token{Surface: "き", POS: "カ変", Inflect: "用", Lemma: "来", Reading: "き"},
		},
		{
			// noun with subtype (place name)
			raw:  "みよしの/名-地名/みよしの",
			want: Token{Surface: "みよしの", POS: "名", Subtype: "地名", Reading: "みよしの"},
		},
		{
			// particle with subtype
			raw:  "つゝ/接助-反復/つつ",
			want: Token{Surface: "つゝ", POS: "接助", Subtype: "反復", Reading: "つつ"},
		},
		{
			// auxiliary with combined terminal/attributive form
			raw:  "らん/推-終体:らむ:らむ/らむ",
			want: Token{Surface: "らん", POS: "推", Inflect: "終体", Lemma: "らむ", Reading: "らむ"},
		},
		{
			// perfective auxiliary in attributive form
			raw:  "る/完-体:り:り/る",
			want: Token{Surface: "る", POS: "完", Inflect: "体", Lemma: "り", Reading: "る"},
		},
		{
			// terminal auxiliary (sentence-final)
			raw:  "けり/過-終:けり:けり/けり",
			want: Token{Surface: "けり", POS: "過", Inflect: "終", Lemma: "けり", Reading: "けり"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := parseToken(tt.raw)
			if err != nil {
				t.Fatalf("parseToken(%q) error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("parseToken(%q)\n got  %+v\n want %+v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseMorphLine(t *testing.T) {
	// poem 3 from morphological-annotation.txt
	line := "10003 春霞/名/はるがすみ たて/タ四-已:立つ:たつ/たて る/完-体:り:り/る や/係助/や いつこ/代/いづこ みよしの/名-地名/みよしの ゝ/格助/の 吉野/名-地名/よしの の/格助/の 山/名/やま に/格助/に 雪/名/ゆき は/係助/は ふり/ラ四-用:降る:ふる/ふり つゝ/接助-反復/つつ"

	id, tokens, err := parseMorphLine(line)
	if err != nil {
		t.Fatalf("parseMorphLine error: %v", err)
	}
	if id != 3 {
		t.Errorf("id: got %d, want 3 (10003 - 10000)", id)
	}
	if len(tokens) != 15 {
		t.Errorf("len(tokens): got %d, want 15", len(tokens))
	}
	if tokens[0].Surface != "春霞" {
		t.Errorf("tokens[0].Surface: got %q, want %q", tokens[0].Surface, "春霞")
	}
	if tokens[1].Inflect != "已" {
		t.Errorf("tokens[1].Inflect: got %q, want %q", tokens[1].Inflect, "已")
	}
}

func TestParseMorphLine_malformed(t *testing.T) {
	_, _, err := parseMorphLine("not a valid line")
	if err == nil {
		t.Error("expected error for malformed line, got nil")
	}
}

func TestLoadMorphData(t *testing.T) {
	// minimal two-poem input
	input := strings.NewReader(
		"10001 年/名/とし の/格助/の\n" +
			"10002 袖/名/そで ひち/タ四-用:漬つ:ひつ/ひち\n",
	)
	data, err := loadMorphData(input)
	if err != nil {
		t.Fatalf("loadMorphData error: %v", err)
	}
	if len(data) != 2 {
		t.Errorf("len(data): got %d, want 2", len(data))
	}
	// 10001 - 10000 = 1
	tokens, ok := data[1]
	if !ok {
		t.Fatal("poem 1 not found (key should be 1 = 10001 - 10000)")
	}
	if len(tokens) != 2 {
		t.Errorf("poem 1 tokens: got %d, want 2", len(tokens))
	}
	if tokens[0].Surface != "年" {
		t.Errorf("tokens[0].Surface: got %q, want %q", tokens[0].Surface, "年")
	}
}
