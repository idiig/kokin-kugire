package kugire

import (
	"fmt"

	"github.com/beevik/etree"
)

// AnnotateXML writes kugire positions for poem id into the XML file at path.
// Existing <k> children of the target <l> are removed and replaced.
// <k> elements are appended at the end of <l>, after all <seg> children.
// @n is 1-based (AfterSeg + 1); @source is the evidence code.
func AnnotateXML(path string, id int, positions []KugirePos) error {
	doc := etree.NewDocument()
	doc.ReadSettings.PreserveCData = true
	if err := doc.ReadFromFile(path); err != nil {
		return fmt.Errorf("read xml %s: %w", path, err)
	}

	l := doc.FindElement(fmt.Sprintf("//l[@n='%d']", id))
	if l == nil {
		return fmt.Errorf("poem %d not found in %s", id, path)
	}

	// Remove existing <k> children
	for _, k := range l.SelectElements("k") {
		l.RemoveChild(k)
	}

	// Append new <k> elements (sorted by AfterSeg for readability)
	for _, p := range positions {
		k := l.CreateElement("k")
		k.CreateAttr("n", fmt.Sprintf("%d", p.AfterSeg+1))
		k.CreateAttr("source", p.Source)
	}

	doc.WriteSettings.CanonicalEndTags = false
	if err := doc.WriteToFile(path); err != nil {
		return fmt.Errorf("write xml %s: %w", path, err)
	}
	return nil
}
