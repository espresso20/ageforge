#!/bin/bash
# Re-cuts the fallback fonts to the runes in subset/*.txt. Usage:
#   window/fonts/subset.sh <dir with the full source fonts>
# The source fonts are the upstream Noto files (see README.md for where to get
# them). Needs fonttools: pip install fonttools brotli
set -euo pipefail
src=${1:?usage: subset.sh <source font dir>}
here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
cut() { # cut <source file> <rune list> <output file>
	pyftsubset "$1" --text-file="$here/subset/$2" --no-hinting --layout-features= --notdef-outline --output-file="$3"
}
# The variable fonts first become one static instance at regular weight.
inst() { fonttools varLib.instancer "$1" "${@:3}" -o "$2" -q; }

for f in NotoSansSymbols2-Regular NotoSansSymbols NotoSansMath-Regular NotoSansCanadianAboriginal; do
	cut "$src/$f.ttf" "$f.txt" "$here/$f.ttf"
done
inst "$src/NotoSansJP[wght].ttf" "$tmp/jp.ttf" wght=400
cut "$tmp/jp.ttf" NotoSansJP-Regular.txt "$here/NotoSansJP-Regular.ttf"
inst "$src/NotoEmoji[wght].ttf" "$tmp/emoji.ttf" wght=400
cut "$tmp/emoji.ttf" NotoEmoji-Regular.txt "$here/NotoEmoji-Regular.ttf"
inst "$src/NotoSans[wdth,wght].ttf" "$tmp/sans.ttf" wght=400 wdth=100
inst "$src/NotoSansHebrew[wdth,wght].ttf" "$tmp/heb.ttf" wght=400 wdth=100
cut "$tmp/sans.ttf" NotoSans-Extra.txt "$tmp/sans-sub.ttf"
cut "$tmp/heb.ttf" NotoSansHebrew-Extra.txt "$tmp/heb-sub.ttf"
(cd "$tmp" && fonttools merge sans-sub.ttf heb-sub.ttf && mv merged.ttf "$here/NotoSansExtra.ttf")
