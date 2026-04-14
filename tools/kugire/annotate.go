package kugire

import (
	"fmt"

	"github.com/beevik/etree"
)

// AnnotateXML writes kugire positions for poem id into the XML file at path.
// Only existing <k> elements with the given source are removed and replaced;
// <k> elements from other sources are preserved.
// <k> elements are appended at the end of <l>, after all <seg> children.
// @n is 1-based (AfterSeg + 1); @source is the evidence code.
func AnnotateXML(path string, id int, source string, positions []KugirePos) error {
	doc := etree.NewDocument()
	doc.ReadSettings.PreserveCData = true
	if err := doc.ReadFromFile(path); err != nil {
		return fmt.Errorf("read xml %s: %w", path, err)
	}

	l := doc.FindElement(fmt.Sprintf("//l[@n='%d']", id))
	if l == nil {
		return fmt.Errorf("poem %d not found in %s", id, path)
	}

	// Remove existing <k> children for this source only
	for _, k := range l.SelectElements("k") {
		if k.SelectAttrValue("source", "") == source {
			l.RemoveChild(k)
		}
	}

	// Append new <k> elements (sorted by AfterSeg for readability)
	for _, p := range positions {
		k := l.CreateElement("k")
		k.CreateAttr("n", fmt.Sprintf("%d", p.AfterSeg+1))
		k.CreateAttr("source", p.Source)
		if p.Cert != "" {
			k.CreateAttr("cert", p.Cert)
		}
	}

	doc.WriteSettings.CanonicalEndTags = false
	if err := doc.WriteToFile(path); err != nil {
		return fmt.Errorf("write xml %s: %w", path, err)
	}
	return nil
}
