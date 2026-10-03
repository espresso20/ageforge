#!/usr/bin/env bash
# Cross-machine determinism check: one seed must play the same run on every
# OS and CPU architecture (see CONTRIBUTING.md, "Float rules for simulation
# code").
#
#   scripts/determinism.sh fingerprint <dir>     play the fixed runs, write <dir>/report.json and <dir>/config.sha256
#   scripts/determinism.sh compare <dir> <dir>...  fail unless every fingerprint matches the first
#   scripts/determinism.sh local                 on Apple Silicon: fingerprint arm64 and amd64 (Rosetta), compare
#
# GOARCH in the environment picks the architecture for `fingerprint` (an
# amd64 binary runs under Rosetta on an Apple Silicon Mac).
set -euo pipefail

# The fixed runs: three seeds of the progression bot through a prestige (at
# the Modern Age) and one seed of the Era Mastery veteran preset, with a state
# digest every 2000 ticks. About 1-3 minutes.
SEED_BASE=1
SEEDS=3
STOP_AGE=digital_age
DIGEST_EVERY=2000

fingerprint() {
	local out=$1
	rm -rf "$out"
	mkdir -p "$out"
	go run ./cmd/export_config --out "$out/config" >/dev/null
	(cd "$out/config" && find . -type f | LC_ALL=C sort | xargs shasum -a 256) >"$out/config.sha256"
	go run ./cmd/smoke -tier fast -scenario progression -seed-base "$SEED_BASE" -seeds "$SEEDS" \
		-stop-age "$STOP_AGE" -digest-every "$DIGEST_EVERY" -pacing report -out "$out/smoke" >/dev/null
	# One seed of the Era Mastery veteran preset, so k-scaled play (rates,
	# storage, build and research times at k = 4.16) replays everywhere too.
	go run ./cmd/smoke -tier fast -scenario progression -preset veteran -seed-base "$SEED_BASE" -seeds 1 \
		-stop-age "$STOP_AGE" -digest-every "$DIGEST_EVERY" -pacing report -out "$out/veteran" >/dev/null
	# Everything in the report but wall-clock times must match.
	jq -S --slurpfile v "$out/veteran/report.json" \
		'.scenarios += ($v[0].scenarios | map(.name = "veteran-preset")) | del(.. | .wall_ms?, .started?)' \
		"$out/smoke/report.json" >"$out/report.json"
	echo "fingerprint $(go env GOOS)/$(go env GOARCH): $(cat "$out/config.sha256" "$out/report.json" | shasum -a 256 | cut -c1-16)"
	runs "$out" | sed 's/^/  /'
}

# runs prints one line per seed: seed, ticks, final digest.
runs() {
	jq -r '.scenarios[].progression[]?.runs[] | "seed \(.seed): \(.ticks) ticks, \(.final_age), state \(.state_digest), map \(.map_digest // "none")"' "$1/report.json"
}

# firstSplit prints, per seed, the first digest-trail entry where b differs
# from a: the two runs parted in the ticks before it.
firstSplit() {
	jq -rn --slurpfile a "$1/report.json" --slurpfile b "$2/report.json" '
		def runs(r): [r[0].scenarios[].progression[]?.runs[]];
		[runs($a), runs($b)] | transpose[] | select(.[0] != null and .[1] != null)
		| . as [$x, $y]
		| ([range(0; [($x.digests // []), ($y.digests // [])] | map(length) | min)]
			| map(select($x.digests[.].digest != $y.digests[.].digest)) | first) as $i
		| if $i == null then
			(if $x.state_digest != $y.state_digest then "  seed \($x.seed): trails agree, final states differ (after tick \($x.digests[-1].tick // 0))" else empty end)
		  else
			"  seed \($x.seed): first differs at tick \($x.digests[$i].tick) (last match: tick \(if $i > 0 then $x.digests[$i-1].tick else "none" end))"
		  end'
}

compare() {
	local ref=$1
	shift
	local fail=0
	for d in "$@"; do
		if ! cmp -s "$ref/config.sha256" "$d/config.sha256"; then
			echo "MISMATCH: config differs between $ref and $d:"
			diff "$ref/config.sha256" "$d/config.sha256" || true
			fail=1
		fi
		if ! cmp -s "$ref/report.json" "$d/report.json"; then
			echo "MISMATCH: runs differ between $ref and $d"
			echo "$ref:"
			runs "$ref" | sed 's/^/  /'
			echo "$d:"
			runs "$d" | sed 's/^/  /'
			firstSplit "$ref" "$d"
			diff "$ref/report.json" "$d/report.json" | head -40 || true
			fail=1
		fi
	done
	if [ "$fail" -ne 0 ]; then
		return 1
	fi
	echo "OK: $# fingerprint(s) match $ref"
	runs "$ref" | sed 's/^/  /'
}

case "${1:-}" in
fingerprint)
	fingerprint "${2:?usage: $0 fingerprint <dir>}"
	;;
compare)
	[ $# -ge 3 ] || {
		echo "usage: $0 compare <dir> <dir>..." >&2
		exit 2
	}
	shift
	compare "$@"
	;;
local)
	base=${2:-determinism}
	(export GOARCH=arm64 && fingerprint "$base/arm64")
	(export GOARCH=amd64 && fingerprint "$base/amd64")
	compare "$base/arm64" "$base/amd64"
	;;
*)
	sed -n '2,11p' "$0" >&2
	exit 2
	;;
esac
