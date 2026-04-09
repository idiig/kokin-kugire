Batch-generate kugire suggestions for all 1111 poems with the given source.

Usage: /kugire-tag <source>   (source = morph | kaneko)

Run this bash command from the project root:

```bash
nix develop --command bash -c "
  cd tools/kugire && ./kugire-tag --source $ARGUMENTS \
    --xml ../../data/kokinwakashu.xml \
    --morph ../../data/morphological-annotation.txt \
    --kaneko ../../data/translation-kaneko.txt \
    --output ../../data/kokin-kugire.xml \
    2>/tmp/kugire-$ARGUMENTS.log
"
```

If the binary is not built yet, build first:
```bash
nix develop --command bash -c "cd tools/kugire && go build ./cmd/kugire-tag/"
```

After the command completes, report:
- Number of cache files: `ls tools/kugire/cache/kugire-*-$ARGUMENTS.txt | wc -l`
- Number of <k> elements in XML: `grep -c '<k ' data/kokin-kugire.xml`

**Important**: Run only one instance at a time. If data/kokin-kugire.xml is missing or corrupted, restore with:
```bash
cp data/kokinwakashu.xml data/kokin-kugire.xml
```
