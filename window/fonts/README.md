# Window fonts

All fonts are under the SIL Open Font License; each has its `OFL-*.txt`.

- `JetBrainsMono-*.ttf`, `DepartureMono-Regular.otf`: the two faces the player picks (whole fonts).
- `NotoSans*.ttf`, `NotoEmoji-Regular.ttf`, `NotoSansJP-Regular.ttf`, `NotoSansExtra.ttf`: fallbacks for the runes the two faces lack. Each is **cut down to the runes the game prints**, listed in `subset/<name>.txt`. Block and box drawing (U+2500 to U+259F) are drawn as shapes in `glyph.go` and need no font.

`runes_test.go` fails when the game prints a rune that no font or shape covers. To add one: put it in the list of the first fallback that has it (order: Symbols2, Symbols, Math, Canadian Aboriginal, JP, Emoji, Extra), then re-cut.

## Re-cutting

```
pip install fonttools brotli
window/fonts/subset.sh <dir with the full source fonts>
```

The source dir needs these upstream files:

| File | From |
| --- | --- |
| `NotoSansSymbols2-Regular.ttf`, `NotoSansSymbols.ttf`, `NotoSansMath-Regular.ttf`, `NotoSansCanadianAboriginal.ttf` | the full copies in git at 47f69d12 (`git show 47f69d12:window/fonts/<name>`) |
| `NotoSansJP[wght].ttf` | https://github.com/google/fonts/raw/main/ofl/notosansjp/NotoSansJP%5Bwght%5D.ttf |
| `NotoEmoji[wght].ttf` (monochrome) | https://github.com/google/fonts/raw/main/ofl/notoemoji/NotoEmoji%5Bwght%5D.ttf |
| `NotoSans[wdth,wght].ttf` | https://github.com/google/fonts/raw/main/ofl/notosans/NotoSans%5Bwdth,wght%5D.ttf |
| `NotoSansHebrew[wdth,wght].ttf` | https://github.com/google/fonts/raw/main/ofl/notosanshebrew/NotoSansHebrew%5Bwdth,wght%5D.ttf |

The core of it, per font: `pyftsubset <full font> --text-file=subset/<name>.txt --no-hinting --layout-features= --notdef-outline --output-file=<name>.ttf`. The variable fonts are first set to regular weight with `fonttools varLib.instancer <font> wght=400 -o <static>`; `NotoSansExtra.ttf` is the Noto Sans and Noto Sans Hebrew cuts merged with `fonttools merge`.
