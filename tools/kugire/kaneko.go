package kugire

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// KanekoSource returns a TranslationSource for Kaneko's translation file.
func KanekoSource(path string) TranslationSource {
	return TranslationSource{Code: "kaneko", Path: path, Load: loadKanekoData}
}

// loadKanekoData parses translation-kaneko.txt (BBDB format) and returns
// a map from poem ID to the $D (modern Japanese prose) field.
// Lines before $$DATA| are header/metadata and are skipped.
func loadKanekoData(r io.Reader) (map[int]string, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	result := make(map[int]string)
	inData := false
	var currentID int
	var currentD string

	flush := func() {
		if currentID > 0 && currentD != "" {
			result[currentID] = currentD
		}
		currentID = 0
		currentD = ""
	}

	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")

		if !inData {
			if strings.HasPrefix(line, "$$DATA|") {
				inData = true
			}
			continue
		}

		if strings.HasPrefix(line, "$A|") {
			flush()
			idStr := strings.TrimPrefix(line, "$A|")
			rawID, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				return nil, fmt.Errorf("$A field %q: %w", idStr, err)
			}
			currentID = rawID % 10000
			continue
		}

		if strings.HasPrefix(line, "$D|") {
			currentD = strings.TrimPrefix(line, "$D|")
			continue
		}
	}
	flush()

	return result, nil
}
