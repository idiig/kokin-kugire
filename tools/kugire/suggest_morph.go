package kugire

import (
	"fmt"
	"strings"
)

// classifyToken returns "K1" (strong kugire), "K2" (weak/nominal), or "K0" (not kugire)
// for the last token of a segment, based on its POS and inflection.
// Rules ported from kugire_from_kokin.c classify_token().
func classifyToken(tok Token) string {
	pos := tok.POS
	inflect := tok.Inflect

	// Build a combined field matching the original C pos string, e.g. "過-終:けり:けり"
	// for pattern matching purposes.
	combined := pos
	if inflect != "" {
		combined = pos + "-" + inflect + ":"
	}

	// K1: imperative
	if strings.Contains(inflect, "命") {
		return "K1"
	}
	// K1: sentence-final particle
	if strings.Contains(pos, "終助") {
		return "K1"
	}
	// K1: binding particle (こそ・ぞ・なむ・や・か)
	if strings.Contains(pos, "係助") {
		return "K1"
	}
	// K1: terminal inflection forms
	if strings.Contains(combined, "-終:") {
		return "K1"
	}
	if strings.Contains(combined, "-終体:") {
		return "K1"
	}
	// K1: realis form (已然形, broad — covers こそ結び)
	if strings.Contains(combined, "-已:") {
		return "K1"
	}

	// K0: case particle
	if strings.Contains(pos, "格助") {
		return "K0"
	}
	// K0: conjunctive particle
	if strings.Contains(pos, "接助") {
		return "K0"
	}
	// K0: continuative form
	if strings.Contains(combined, "-用:") {
		return "K0"
	}
	// K0: attributive form (終体 already caught above)
	if strings.Contains(combined, "-体:") {
		return "K0"
	}

	// K2: noun (体言止め)
	if strings.HasPrefix(pos, "名") {
		return "K2"
	}

	// K2: default (pronoun, adverb, etc.)
	return "K2"
}

// normalizeKana expands ゝ/ゞ iteration marks and strips spaces.
// ゝ repeats the previous kana character; ゞ repeats with voicing.
func normalizeKana(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	for i, r := range runes {
		switch r {
		case 'ゝ':
			if i > 0 {
				out = append(out, runes[i-1])
			}
		case 'ゞ':
			if i > 0 {
				// naive voicing: not needed for matching since token readings
				// use the already-voiced form; just repeat for now
				out = append(out, runes[i-1])
			}
		case ' ', '　', '\t':
			// skip
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

// alignTokensToSegments groups tokens into one slice per segment.
// segKanas holds the kana reading for each segment (from XML lemmaRef);
// matching is done by concatenating token Readings against normalizeKana(segKana).
func alignTokensToSegments(tokens []Token, segKanas []string) ([][]Token, error) {
	groups := make([][]Token, len(segKanas))
	idx := 0
	for si, segKana := range segKanas {
		want := normalizeKana(segKana)
		var buf strings.Builder
		start := idx
		for idx < len(tokens) {
			buf.WriteString(tokens[idx].Reading)
			idx++
			if buf.String() == want {
				groups[si] = tokens[start:idx]
				break
			}
			if len([]rune(buf.String())) > len([]rune(want)) {
				return nil, fmt.Errorf("seg %d: kana overflow: built %q, want %q", si, buf.String(), want)
			}
		}
		if groups[si] == nil {
			return nil, fmt.Errorf("seg %d: could not align: built %q, want %q", si, buf.String(), want)
		}
	}
	return groups, nil
}

// certForToken maps a classified token to a certainty level.
// K1 tokens are split into "high" (clear terminal forms) vs "mid" (binding particles).
// K2 tokens (体言止め etc.) are "low".
func certForToken(tok Token, class string) string {
	switch class {
	case "K1":
		if strings.Contains(tok.POS, "係助") {
			return "mid"
		}
		return "high"
	case "K2":
		return "low"
	default:
		return ""
	}
}

// SuggestMorph returns kugire positions suggested by morphological analysis.
// It evaluates the last token of each segment (except the final one).
// K1 tokens (終止・命令・已然・終助詞) → returned with cert "high" or "mid".
// K2 tokens (体言止め等) → returned with cert "low".
func SuggestMorph(d PoemData) []KugirePos {
	groups, err := alignTokensToSegments(d.Tokens, d.SegmentsKana)
	if err != nil {
		return nil
	}

	var positions []KugirePos
	for i, group := range groups[:len(groups)-1] {
		last := group[len(group)-1]
		class := classifyToken(last)
		if class == "K0" {
			continue
		}
		cert := certForToken(last, class)
		positions = append(positions, KugirePos{AfterSeg: i, Source: "morph", Cert: cert})
	}
	return positions
}
