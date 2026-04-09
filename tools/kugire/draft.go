package kugire

import (
	"fmt"
	"regexp"
	"strings"
)

var tagRe = regexp.MustCompile(`\[K:([^\]]+)\]`)

// RenderDraft produces the nano draft for a poem.
// It begins with comment lines containing reference information
// (poem ID, translations, morphological hints), followed by the
// 5 content lines — one per segment — with [K:<source>] tags appended
// where a kugire is suggested. Tags on the final segment are omitted.
//
// Comment lines (starting with #) are ignored by ParseDraft.
func RenderDraft(d PoemData, positions []KugirePos) string {
	var sb strings.Builder

	// ── header ──────────────────────────────────────────────────
	fmt.Fprintf(&sb, "# Poem %d\n", d.ID)
	sb.WriteString("# Add/remove [K:<source>] tags. Lines starting with # are ignored.\n")
	sb.WriteString("#\n")

	// ── translations ────────────────────────────────────────────
	for code, text := range d.Translations {
		fmt.Fprintf(&sb, "# %s: %s\n", code, text)
	}
	sb.WriteString("#\n")

	// ── morphological hints ──────────────────────────────────────
	// Show the last token of each segment (except the final) with its class.
	if len(d.Tokens) > 0 && len(d.SegmentsKana) > 0 {
		sb.WriteString("# morph:\n")
		groups, err := alignTokensToSegments(d.Tokens, d.SegmentsKana)
		if err == nil {
			for i, group := range groups {
				last := group[len(group)-1]
				class := classifyToken(last)
				inflectStr := last.Inflect
				if inflectStr == "" {
					inflectStr = "-"
				}
				marker := ""
				if i < len(d.Segments)-1 && class == "K1" {
					marker = " ← K1"
				}
				fmt.Fprintf(&sb, "#   seg %d: %s  末: %s (%s-%s) %s%s\n",
					i+1, d.Segments[i], last.Surface, last.POS, inflectStr, class, marker)
			}
		}
		sb.WriteString("#\n")
	}

	// ── content lines ────────────────────────────────────────────
	tags := make(map[int][]string)
	for _, p := range positions {
		if p.AfterSeg < len(d.Segments)-1 {
			tags[p.AfterSeg] = append(tags[p.AfterSeg], p.Source)
		}
	}
	for i, seg := range d.Segments {
		sb.WriteString(seg)
		for _, src := range tags[i] {
			fmt.Fprintf(&sb, " [K:%s]", src)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ParseDraft parses an edited nano draft and returns the kugire positions.
// Lines beginning with # and blank lines are skipped.
// Content lines are counted 0-based as segment indices.
// Tags on the last content line (segment 4) are ignored.
func ParseDraft(raw string) ([]KugirePos, error) {
	var positions []KugirePos
	lineIdx := 0

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		for _, m := range tagRe.FindAllStringSubmatch(trimmed, -1) {
			source := m[1]
			// lineIdx is 0-based segment index; skip tags on last segment
			if lineIdx < 4 {
				positions = append(positions, KugirePos{AfterSeg: lineIdx, Source: source})
			}
		}
		lineIdx++
	}
	return positions, nil
}
