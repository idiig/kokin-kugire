package kugire

import (
	"fmt"
	"regexp"
	"strings"
)

var tagRe = regexp.MustCompile(`\[K:([^\]]+)\]`)

// formatTag renders a [K:...] tag for a KugirePos.
// Format: [K:source:cert] if cert is set, [K:source] otherwise.
func formatTag(p KugirePos) string {
	if p.Cert != "" {
		return fmt.Sprintf("[K:%s:%s]", p.Source, p.Cert)
	}
	return fmt.Sprintf("[K:%s]", p.Source)
}

// RenderDraft produces the nano draft for a poem.
// It begins with comment lines containing reference information
// (poem ID, translations, morphological hints), followed by the
// 5 content lines — one per segment — with [K:<source>] tags appended
// where a kugire is suggested. Tags on the final segment are omitted.
//
// reasoning is an optional map from translator code to SemanReasoning;
// when present, alignment and breaks are rendered as comment lines.
// Comment lines (starting with #) are ignored by ParseDraft.
func RenderDraft(d PoemData, positions []KugirePos, reasoning map[string]SemanReasoning) string {
	var sb strings.Builder

	// ── header ──────────────────────────────────────────────────
	fmt.Fprintf(&sb, "# Poem %d\n", d.ID)
	sb.WriteString("# Add/remove [K:<source>] tags. Lines starting with # are ignored.\n")
	sb.WriteString("#\n")

	// ── translations ────────────────────────────────────────────
	for code, text := range d.Translations {
		fmt.Fprintf(&sb, "# %s: %s\n", code, text)
		if r, ok := reasoning[code]; ok {
			fmt.Fprintf(&sb, "#   alignment: %s\n", r.Alignment)
			fmt.Fprintf(&sb, "#   breaks: %s\n", r.Breaks)
		}
	}
	sb.WriteString("#\n")

	// ── morphological hints ──────────────────────────────────────
	// Show the last token of each segment with its class and certainty.
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
				if i < len(d.Segments)-1 && class != "K0" {
					cert := certForToken(last, class)
					marker = fmt.Sprintf(" ← %s [%s]", class, cert)
				}
				fmt.Fprintf(&sb, "#   seg %d: %s  末: %s (%s-%s) %s%s\n",
					i+1, d.Segments[i], last.Surface, last.POS, inflectStr, class, marker)
			}
		}
		sb.WriteString("#\n")
	}

	// ── content lines ────────────────────────────────────────────
	tags := make(map[int][]KugirePos)
	for _, p := range positions {
		if p.AfterSeg < len(d.Segments)-1 {
			tags[p.AfterSeg] = append(tags[p.AfterSeg], p)
		}
	}
	for i, seg := range d.Segments {
		sb.WriteString(seg)
		for _, p := range tags[i] {
			fmt.Fprintf(&sb, " %s", formatTag(p))
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
			// m[1] is "source" or "source:cert"
			parts := strings.SplitN(m[1], ":", 2)
			source := parts[0]
			cert := ""
			if len(parts) == 2 {
				cert = parts[1]
			}
			// lineIdx is 0-based segment index; skip tags on last segment
			if lineIdx < 4 {
				positions = append(positions, KugirePos{AfterSeg: lineIdx, Source: source, Cert: cert})
			}
		}
		lineIdx++
	}
	return positions, nil
}
