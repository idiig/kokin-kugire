package kugire

import (
	"fmt"
	"os"
)

// LoadPoem assembles a PoemData for the given poem ID from the XML file,
// the morphological annotation file, and zero or more translation sources.
//
// Adding a new translator requires only passing an additional TranslationSource
// — no changes to this function are needed.
func LoadPoem(xmlPath, morphPath string, id int, sources []TranslationSource) (PoemData, error) {
	// Segments from XML
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return PoemData{}, fmt.Errorf("open xml: %w", err)
	}
	defer xmlFile.Close()

	segs, kanas, err := loadSegments(xmlFile, id)
	if err != nil {
		return PoemData{}, fmt.Errorf("load segments (poem %d): %w", id, err)
	}

	// Morphological tokens
	morphFile, err := os.Open(morphPath)
	if err != nil {
		return PoemData{}, fmt.Errorf("open morph: %w", err)
	}
	defer morphFile.Close()

	morphData, err := loadMorphData(morphFile)
	if err != nil {
		return PoemData{}, fmt.Errorf("load morph data: %w", err)
	}

	// Translations — one file per source
	translations := make(map[string]string)
	for _, src := range sources {
		f, err := os.Open(src.Path)
		if err != nil {
			return PoemData{}, fmt.Errorf("open translation %q: %w", src.Code, err)
		}
		data, err := src.Load(f)
		f.Close()
		if err != nil {
			return PoemData{}, fmt.Errorf("load translation %q: %w", src.Code, err)
		}
		translations[src.Code] = data[id]
	}

	return PoemData{
		ID:           id,
		Segments:     segs,
		SegmentsKana: kanas,
		Tokens:       morphData[id],
		Translations: translations,
	}, nil
}
