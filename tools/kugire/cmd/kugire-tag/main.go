// Command kugire-tag batch-generates kugire suggestions for all poems and
// writes cache files + XML annotations automatically.
//
// Usage:
//
//	kugire-tag --source <morph|kaneko> [flags]
//	kugire-tag --source morph --force   # regenerate even if cache exists
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	kugire "kokin-kugire/tools/kugire"
)

func main() {
	xmlPath := flag.String("xml", "data/kokinwakashu.xml", "poem TEI XML (source)")
	morphPath := flag.String("morph", "data/morphological-annotation.txt", "morphological annotation")
	kanekoPath := flag.String("kaneko", "data/translation-kaneko.txt", "Kaneko translation")
	outputPath := flag.String("output", "data/kokin-kugire.xml", "annotated output XML")
	sourceName := flag.String("source", "", "kugire source: morph or kaneko (required)")
	force := flag.Bool("force", false, "overwrite existing cache files")
	flag.Parse()

	if *sourceName == "" {
		fmt.Fprintf(os.Stderr, "Error: --source is required (morph or kaneko)\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	switch *sourceName {
	case "morph", "kaneko":
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown source %q (use morph or kaneko)\n", *sourceName)
		os.Exit(1)
	}

	ids, err := kugire.LoadAllPoemIDs(*xmlPath)
	if err != nil {
		log.Fatalf("load poem IDs: %v", err)
	}

	if err := os.MkdirAll("cache", 0755); err != nil {
		log.Fatalf("mkdir cache: %v", err)
	}
	if err := ensureOutput(*outputPath, *xmlPath); err != nil {
		log.Fatalf("ensure output xml: %v", err)
	}

	total := len(ids)
	sources := []kugire.TranslationSource{
		kugire.KanekoSource(*kanekoPath),
	}

	for i, id := range ids {
		fmt.Fprintf(os.Stderr, "poem %d/%d (id=%d)\n", i+1, total, id)

		draftPath := fmt.Sprintf("cache/kugire-%d-%s.txt", id, *sourceName)

		// Skip if cache exists and not forcing.
		if !*force {
			if _, err := os.Stat(draftPath); err == nil {
				// Cache exists; still annotate XML from the cached draft.
				content, err := os.ReadFile(draftPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  read cache: %v (skipping)\n", err)
					continue
				}
				confirmed, err := kugire.ParseDraft(string(content))
				if err != nil {
					fmt.Fprintf(os.Stderr, "  parse draft: %v (skipping)\n", err)
					continue
				}
				if err := kugire.AnnotateXML(*outputPath, id, confirmed); err != nil {
					fmt.Fprintf(os.Stderr, "  annotate xml: %v (skipping)\n", err)
				}
				continue
			}
		}

		d, err := kugire.LoadPoem(*xmlPath, *morphPath, id, sources)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  load poem: %v (skipping)\n", err)
			continue
		}

		var positions []kugire.KugirePos
		switch *sourceName {
		case "morph":
			positions = kugire.SuggestMorph(d)
		case "kaneko":
			positions, err = kugire.SuggestSeman(d, "kaneko")
			if err != nil {
				fmt.Fprintf(os.Stderr, "  SuggestSeman: %v (skipping)\n", err)
				continue
			}
		}

		draft := kugire.RenderDraft(d, positions)
		if err := os.WriteFile(draftPath, []byte(draft), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  write cache: %v (skipping)\n", err)
			continue
		}

		if err := kugire.AnnotateXML(*outputPath, id, positions); err != nil {
			fmt.Fprintf(os.Stderr, "  annotate xml: %v\n", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Done. Processed %d poems.\n", total)
}

// ensureOutput copies sourcePath to outputPath if outputPath does not yet exist.
func ensureOutput(outputPath, sourcePath string) error {
	if _, err := os.Stat(outputPath); err == nil {
		return nil
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source xml: %w", err)
	}
	return os.WriteFile(outputPath, data, 0644)
}
