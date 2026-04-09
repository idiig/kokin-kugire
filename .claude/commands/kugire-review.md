Open the kugire draft for a specific poem in nano and write confirmed positions back to XML.

Usage: /kugire-review <poem-id> <source>   (source = morph | kaneko)

Parse arguments: first token = poem ID (N), second token = source (SOURCE).

## Step 1 — Run review

```bash
cd /Users/idg/Documents/kokin-kugire && nix develop --command bash -c "
  cd tools/kugire && echo '' | ./kugire-review --source SOURCE \
    --xml ../../data/kokinwakashu.xml \
    --morph ../../data/morphological-annotation.txt \
    --kaneko ../../data/translation-kaneko.txt \
    --output ../../data/kokin-kugire.xml \
    N
" 2>&1
```

If the binary is not built yet, build first:
```bash
nix develop --command bash -c "cd tools/kugire && go build ./cmd/kugire-review/"
```

Tell the user: "nano opened in tmux window kugire-N. Edit the draft and press Enter in this window when done."

After the command completes, show:
- Cache file: `cat tools/kugire/cache/kugire-N-SOURCE.txt`
- XML result: `grep -o '<k[^/]*/>' data/kokin-kugire.xml` filtered for poem N

## Step 2 — Commit

Commit the updated cache file and XML without a Claude co-author line:

```bash
git add tools/kugire/cache/kugire-N-SOURCE.txt data/kokin-kugire.xml
git commit -m "review(SOURCE): poem N"
```

## Step 3 — Continue

Ask the user:
> "Continue to poem N+1 with source SOURCE? (y/n)"

If yes, run /kugire-review N+1 SOURCE.
