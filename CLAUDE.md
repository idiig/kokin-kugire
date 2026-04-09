# kokin-kugire

## Project Overview

Adding a kugire (句切れ) annotation layer to the Kokinwakashu (古今和歌集) TEI XML data. Kugire are phrase/clause breaks within poems — ambiguous judgements that may be grammatical (derivable from morphological analysis) or semantic (derivable from expert explanatory translations). The annotation schema supports multiple competing interpretations, each tagged with its evidence source.

## Language Policy

- **Conversation**: Japanese
- **Documentation (docs/, README, CLAUDE.md)**: English
- **Code and commit messages**: English
- **XML content**: Japanese (original text)

## Data Sources

- `data/kokinwakashu.xml` — Poem text (TEI XML)
- `data/morphological-annotation.txt` — Morphological annotations; one poem per line, space-separated tokens in `surface/POS[inflection]/reading` format; poem ID = `10NNN` for poem NNN
- `data/translation-kaneko.txt` — Kaneko's explanatory translation (BBDB format); field `$D` = modern Japanese prose; poem ID = `00NNN` (last 4 digits)

## Tech Stack

- **Data format**: TEI XML (P5)
- **Programming language**: Go
- **Editor**: `nano` (kugire review in tmux)
- **Environment**: Nix flake (`nix develop`); `go`, `nano`, `tmux`, `ollama` provided by devShell
- **LLM assistance**: Semantic kugire suggestion via local Ollama (`qwen2.5`); grammatical suggestion via morphological rules

## Project Structure

```
data/
  kokinwakashu.xml              — Poem text (TEI XML, source)
  morphological-annotation.txt  — Morphological annotations (source)
  translation-kaneko.txt        — Kaneko's explanatory translation (source)
  kokin-kugire.xml              — Annotated output (kugire tags added to poem XML)

tools/kugire/cache/
  kugire-{id}-{source}.txt      — Per-poem draft files (tracked in git)

tools/kugire/
  types.go              — Token, KugirePos, PoemData, TranslationSource
  poem.go               — LoadPoem (composes all data sources)
  morph.go              — loadMorphData, parseMorphLine, parseToken
  kaneko.go             — loadKanekoData, KanekoSource
  xml.go                — loadSegments, kanaFromLemmaRef, LoadAllPoemIDs
  suggest_morph.go      — SuggestMorph (K1-only grammatical suggestion)
  suggest_seman.go      — SuggestSeman (Ollama LLM semantic suggestion)
  draft.go              — RenderDraft, ParseDraft
  diff.go               — DiffKugire, KugireDiff
  annotate.go           — AnnotateXML
  cmd/kugire-tag/       — Batch suggestion for all poems
  cmd/kugire-review/    — Interactive per-poem review

.claude/
  rules/      — Technical rules (git-workflow)
docs/         — Project documentation (English)
```

## Session Initialization

Run `nix develop` in the project root. The devShell provides `go`, `nano`, `tmux`, and `ollama`; the shellHook starts `ollama serve` automatically if not already running.

First-time setup (once): `ollama pull qwen2.5`

Tool commands that Claude Code runs are prefixed with `nix develop --command bash -c "..."`.
Commands (`kugire-tag`, `kugire-review`) are run from `tools/kugire/` with explicit data paths:

```bash
cd tools/kugire
./kugire-tag --source morph \
  --xml ../../data/kokinwakashu.xml \
  --morph ../../data/morphological-annotation.txt \
  --kaneko ../../data/translation-kaneko.txt \
  --output ../../data/kokin-kugire.xml
```

## Kugire Annotation

Two types of kugire, each with its own evidence source:

- **Grammatical** (`morph`) — derived from morphological analysis; `SuggestMorph` returns K1 positions only
- **Semantic** — tagged with the translator's code (e.g. `kaneko`); `SuggestSeman` calls Ollama with a few-shot prompt

The schema allows multiple competing kugire annotations on the same poem (different translators may disagree). Do not collapse them into a single interpretation.

### Adding a New Translator

1. Write `loadXxxData(r io.Reader) (map[int]string, error)` in a new file
2. Export a `XxxSource(path string) TranslationSource` constructor
3. Pass the source to `LoadPoem` and add a `case "xxx":` to the source switch
   in both `cmd/kugire-tag/main.go` and `cmd/kugire-review/main.go`

No changes to `LoadPoem`, `SuggestSeman`, or the few-shot prompt are needed.

## Kugire Workflow (Two-Phase)

**Phase 1 — Batch tagging** (`kugire-tag`): Generate suggestions for all poems
and write cache files + XML annotations automatically. Run from `tools/kugire/`:

```bash
./kugire-tag --source morph   [--force] \
  --xml ../../data/kokinwakashu.xml \
  --morph ../../data/morphological-annotation.txt \
  --kaneko ../../data/translation-kaneko.txt \
  --output ../../data/kokin-kugire.xml
```

`morph` is fast (no Ollama). `kaneko` calls Ollama per poem (~30s/poem, several hours
for all 1111). `--force` overwrites existing cache.

**Run only one instance at a time** — concurrent writes corrupt the XML.
If `data/kokin-kugire.xml` is missing or corrupted, restore it with:
```bash
cp data/kokinwakashu.xml data/kokin-kugire.xml
```

**Phase 2 — Interactive review** (`kugire-review`): Open the draft for a
specific poem in nano; writes confirmed edits back to XML. Run from `tools/kugire/`:

```bash
./kugire-review --source morph 1   # cache hit → opens nano directly
./kugire-review --source kaneko 42 # cache miss → suggest → cache → nano
  [same --xml / --morph / --kaneko / --output flags as above]
```

Cache files (`tools/kugire/cache/kugire-{id}-{source}.txt`) are tracked in git
and are the persistent draft record. `AnnotateXML` is called after every nano session.

## Ollama Integration

- Managed by `flake.nix`; `ollama serve` starts automatically via `shellHook` on `nix develop`
- Endpoint: `http://localhost:11434/api/generate` (package var `ollamaEndpoint`; override in tests)
- Model: `qwen2.5` (pull once with `ollama pull qwen2.5`; ~4.7 GB)
- Prompt includes 3 few-shot examples drawn from real Kokinwakashu poems with Kaneko's translation
- Few-shot is translator-agnostic (uses generic 「現代語訳:」label); works for any future translator

## XML Annotation Schema

Kugire positions are written as `<k>` elements appended at the end of `<l>`, after all `<seg>` children:

```xml
<l n="1" xml:id="n1">
  <seg>年の内に</seg>
  <seg>春はきにけり</seg>
  ...
  <k n="2" source="morph"/>
  <k n="2" source="kaneko"/>
  <k n="4" source="morph"/>
</l>
```

- `@n` — 1-based segment number after which the kugire falls (`AfterSeg + 1`)
- `@source` — evidence code (`morph`, `kaneko`, etc.)
- Multiple `<k>` at the same `@n` represent competing interpretations
- `<k>` is a temporary custom tag; may be migrated to `<caesura>` (TEI) later
- `AnnotateXML` removes all existing `<k>` for a poem before writing new ones

## Important Constraints

- Do not modify original text content in XML; only add structural annotations
- Preserve TEI namespace (`http://www.tei-c.org/ns/1.0`) in all XML processing
- Do not call `Indent()` on documents with mixed content (text + elements) —
  it inserts spurious blank lines around text nodes

## Claude Code Self-Maintenance

When general rules, common workflows, or reusable skills emerge during a conversation, update the relevant CLAUDE configuration:

- **Project-level instructions** → `CLAUDE.md` (language policy, constraints, conventions)
- **Technical rules** → `.claude/rules/<topic>.md` (git workflow, coding style, XML processing)
- **Cross-project learnings** → `~/.claude/projects/.../memory/MEMORY.md` (auto memory)
