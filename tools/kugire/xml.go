package kugire

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// LoadAllPoemIDs returns all poem IDs found in the TEI XML file at path,
// in document order, by collecting @n attributes of <l> elements.
func LoadAllPoemIDs(path string) ([]int, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.PreserveCData = true
	if err := doc.ReadFromFile(path); err != nil {
		return nil, fmt.Errorf("read xml %s: %w", path, err)
	}
	var ids []int
	for _, l := range doc.FindElements("//l") {
		attr := l.SelectAttrValue("n", "")
		if attr == "" {
			continue
		}
		n, err := strconv.Atoi(attr)
		if err != nil {
			continue
		}
		ids = append(ids, n)
	}
	return ids, nil
}

// loadSegments reads the TEI XML from r and returns the surface and kana strings
// for each segment of poem id.
// Kana is extracted from <w lemmaRef="#kana.lemma">; surface is the element text.
func loadSegments(r io.Reader, id int) (surfaces, kanas []string, err error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	doc := etree.NewDocument()
	doc.ReadSettings.PreserveCData = true
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, nil, fmt.Errorf("parsing XML: %w", err)
	}

	path := fmt.Sprintf("//l[@n='%d']", id)
	l := doc.FindElement(path)
	if l == nil {
		return nil, nil, fmt.Errorf("poem %d not found in XML", id)
	}

	for _, seg := range l.SelectElements("seg") {
		var surf, kana strings.Builder
		for _, w := range seg.SelectElements("w") {
			surf.WriteString(w.Text())
			ref := w.SelectAttrValue("lemmaRef", "")
			kana.WriteString(kanaFromLemmaRef(ref, w.Text()))
		}
		surfaces = append(surfaces, surf.String())
		kanas = append(kanas, kana.String())
	}
	return surfaces, kanas, nil
}

// kanaFromLemmaRef extracts the kana reading from a lemmaRef attribute.
// Format: "#kana.lemma" or "#kana.lemma #kana2.lemma2 ..."
// Falls back to the surface text if no lemmaRef is present.
func kanaFromLemmaRef(ref, surface string) string {
	if ref == "" {
		return surface
	}
	// take the first ref entry
	first := strings.Fields(ref)[0]
	first = strings.TrimPrefix(first, "#")
	dot := strings.IndexByte(first, '.')
	if dot < 0 {
		return surface
	}
	return first[:dot]
}
