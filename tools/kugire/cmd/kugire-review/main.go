// Command kugire-review opens the kugire draft for a single poem in nano and
// writes confirmed positions back to the output XML.
//
// If a cache file already exists for the given poem+source, it is opened
// directly (no suggestion step). Otherwise the suggestion is run first and
// the result is cached before opening nano.
//
// Usage:
//
//	kugire-review --source <morph|kaneko> [flags] <poem-id>
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	kugire "kokin-kugire/tools/kugire"
)

func main() {
	xmlPath := flag.String("xml", "data/kokinwakashu.xml", "poem TEI XML (source)")
	morphPath := flag.String("morph", "data/morphological-annotation.txt", "morphological annotation")
	kanekoPath := flag.String("kaneko", "data/translation-kaneko.txt", "Kaneko translation")
	outputPath := flag.String("output", "data/kokin-kugire.xml", "annotated output XML")
	sourceName := flag.String("source", "", "kugire source: morph or kaneko (required)")
	flag.Parse()

	if *sourceName == "" {
		fmt.Fprintf(os.Stderr, "Error: --source is required (morph or kaneko)\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: kugire-review --source <morph|kaneko> [flags] <poem-id>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	id, err := strconv.Atoi(flag.Arg(0))
	if err != nil || id <= 0 {
		log.Fatalf("invalid poem id: %q", flag.Arg(0))
	}

	draftPath := fmt.Sprintf("cache/kugire-%d-%s.txt", id, *sourceName)

	// Cache-first: if draft already exists, open it directly.
	if _, err := os.Stat(draftPath); os.IsNotExist(err) {
		sources := []kugire.TranslationSource{
			kugire.KanekoSource(*kanekoPath),
		}
		d, err := kugire.LoadPoem(*xmlPath, *morphPath, id, sources)
		if err != nil {
			log.Fatalf("load poem %d: %v", id, err)
		}

		var positions []kugire.KugirePos
		reasoning := map[string]kugire.SemanReasoning{}
		switch *sourceName {
		case "morph":
			positions = kugire.SuggestMorph(d)
			fmt.Printf("morph: %d suggestion(s)\n", len(positions))
		case "kaneko":
			fmt.Println("Calling Ollama (qwen2.5)…")
			var r kugire.SemanReasoning
			positions, r, err = kugire.SuggestSeman(d, "kaneko")
			if err != nil {
				log.Fatalf("SuggestSeman: %v", err)
			}
			reasoning["kaneko"] = r
			fmt.Printf("kaneko: %d suggestion(s)\n", len(positions))
		default:
			log.Fatalf("unknown source: %q (use morph or kaneko)", *sourceName)
		}

		draft := kugire.RenderDraft(d, positions, reasoning)
		if err := os.MkdirAll("cache", 0755); err != nil {
			log.Fatalf("mkdir cache: %v", err)
		}
		if err := os.WriteFile(draftPath, []byte(draft), 0644); err != nil {
			log.Fatalf("write draft: %v", err)
		}
	} else {
		fmt.Printf("Using existing cache: %s\n", draftPath)
	}

	openInNano(id, draftPath)

	content, err := os.ReadFile(draftPath)
	if err != nil {
		log.Fatalf("read draft: %v", err)
	}
	confirmed, err := kugire.ParseDraft(string(content))
	if err != nil {
		log.Fatalf("parse draft: %v", err)
	}

	if err := ensureOutput(*outputPath, *xmlPath); err != nil {
		log.Fatalf("ensure output xml: %v", err)
	}
	if err := kugire.AnnotateXML(*outputPath, id, *sourceName, confirmed); err != nil {
		log.Fatalf("annotate xml: %v", err)
	}
	fmt.Printf("Written %d kugire position(s) for poem %d to %s\n",
		len(confirmed), id, *outputPath)
}

// openInNano opens the draft in nano via a new tmux window named kugire-N.
// Falls back to a manual prompt if tmux is unavailable.
func openInNano(n int, path string) {
	windowName := fmt.Sprintf("kugire-%d", n)
	cmd := exec.Command("tmux", "new-window", "-n", windowName,
		fmt.Sprintf("nano %s", path))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Could not open tmux window: %v\n", err)
		fmt.Printf("Open manually:  nano %s\n\n", path)
	} else {
		fmt.Printf("nano opened in tmux window (%s).\n", windowName)
	}
	fmt.Print("Press Enter when done editing… ")
	fmt.Scanln()
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
