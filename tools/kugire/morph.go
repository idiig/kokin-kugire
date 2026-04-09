package kugire

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseToken parses one token field from morphological-annotation.txt.
// Format: surface/POS[-{subtype|infltype[:lemma:dictform]}]/reading
func parseToken(raw string) (Token, error) {
	parts := strings.SplitN(raw, "/", 3)
	if len(parts) != 3 {
		return Token{}, fmt.Errorf("token %q: expected 3 /-separated fields, got %d", raw, len(parts))
	}
	surface, posField, reading := parts[0], parts[1], parts[2]

	tok := Token{Surface: surface, Reading: reading}

	dash := strings.IndexByte(posField, '-')
	if dash < 0 {
		// e.g. "名", "代", "格助" — no inflection or subtype
		tok.POS = posField
		return tok, nil
	}

	tok.POS = posField[:dash]
	rest := posField[dash+1:]

	colon := strings.IndexByte(rest, ':')
	if colon < 0 {
		// e.g. "名-地名", "接助-反復" — subtype, not an inflection
		tok.Subtype = rest
		return tok, nil
	}

	// e.g. "用:来:く", "終:けり:けり", "終体:らむ:らむ"
	colonParts := strings.SplitN(rest, ":", 3)
	tok.Inflect = colonParts[0]
	if len(colonParts) >= 2 {
		tok.Lemma = colonParts[1]
	}
	// colonParts[2] is the dict-form; we don't store it separately
	return tok, nil
}

// parseMorphLine parses one line of morphological-annotation.txt.
// Returns the poem number (1-based) and the token list.
func parseMorphLine(line string) (int, []Token, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, nil, fmt.Errorf("empty line")
	}
	sp := strings.IndexByte(line, ' ')
	if sp < 0 {
		return 0, nil, fmt.Errorf("no space in line: %q", line)
	}
	idStr := line[:sp]
	rawID, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, nil, fmt.Errorf("line id %q: %w", idStr, err)
	}
	// IDs are 1NNNN for Kokinwakashu (book 1 of Hachidaishu).
	// Poem number = rawID - 10000.
	poemID := rawID - 10000

	fields := strings.Fields(line[sp+1:])
	tokens := make([]Token, 0, len(fields))
	for _, f := range fields {
		tok, err := parseToken(f)
		if err != nil {
			return 0, nil, err
		}
		tokens = append(tokens, tok)
	}
	return poemID, tokens, nil
}

// loadMorphData reads all lines from r and returns a map from poem ID to tokens.
func loadMorphData(r io.Reader) (map[int][]Token, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]Token)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		id, tokens, err := parseMorphLine(line)
		if err != nil {
			return nil, fmt.Errorf("parseMorphLine: %w", err)
		}
		result[id] = tokens
	}
	return result, nil
}
