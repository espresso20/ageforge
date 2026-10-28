#!/usr/bin/env bash
# Hot-path benchmark comparison, used by .github/workflows/bench.yml and
# runnable on a laptop (see docs/ci.md).
#
#   scripts/ci-bench.sh compare <base-ref> <out-dir>   benchmark <base-ref> and the working tree, back to back
#   scripts/ci-bench.sh single <out-dir>               benchmark the working tree only
#
# `compare` writes <out-dir>/base.txt, head.txt, benchstat.txt, summary.md and
# regressions.txt (empty when nothing got slower by more than the threshold).
# It always exits 0 when the benchmarks ran: a slow benchmark is a report, not
# a failure.
#
# Both trees are built first, then run alternately (base, head, base, head), so
# a noisy neighbor on a shared runner hits both sides instead of one.
#
# The benchmarks: the engine tick and the UI snapshot (game), one dashboard
# frame (ui), and one frame of the map at a ground age and a sky age.
set -euo pipefail

ROUNDS=${BENCH_ROUNDS:-2}        # alternations of base and head
COUNT=${BENCH_COUNT:-3}          # samples per benchmark per round
BENCHTIME=${BENCH_TIME:-200ms}   # per sample
THRESHOLD=${BENCH_THRESHOLD:-15} # percent slower that gets reported

# One suite per line: <package> | <-bench pattern>.
SUITES=(
	"./game|^(BenchmarkTick|BenchmarkGetState)$"
	"./ui|^BenchmarkDashboardFrame$"
	"./ui/mapstyle/procmap|^BenchmarkMapFrame$/^(medieval_age|galactic_age)$/^vibemode_off$"
)

# build_side <tree> <bin-dir>: compile each suite's test binary.
build_side() {
	local tree=$1 bin=$2 i=0 pkg
	mkdir -p "$bin"
	for suite in "${SUITES[@]}"; do
		pkg=${suite%%|*}
		(cd "$tree" && go test -c -o "$bin/suite$i.test" "$pkg")
		i=$((i + 1))
	done
}

# run_side <tree> <bin-dir> <out-file>: run each suite once, append the output.
run_side() {
	local tree=$1 bin=$2 out=$3 i=0 pkg pat
	for suite in "${SUITES[@]}"; do
		pkg=${suite%%|*}
		pat=${suite#*|}
		# A test binary runs in the package directory, as `go test` would.
		(cd "$tree/$pkg" && "$bin/suite$i.test" -test.run='^$' -test.bench="$pat" \
			-test.benchmem -test.benchtime="$BENCHTIME" -test.count="$COUNT") >>"$out"
		i=$((i + 1))
	done
}

# find_regressions <benchstat-output>: print "name  +NN.NN%" for each sec/op
# row that benchstat marks significant and that is slower by more than the
# threshold. Rows without a significant change show "~" and never match.
find_regressions() {
	awk -v limit="$THRESHOLD" '
		/sec\/op/ { metric = "sec/op"; next }
		/B\/op/ { metric = "B/op"; next }
		/allocs\/op/ { metric = "allocs/op"; next }
		NF == 0 { metric = ""; next }
		metric == "sec/op" {
			for (i = 2; i <= NF; i++) {
				if ($i ~ /^\+[0-9.]+%$/) {
					pct = substr($i, 2, length($i) - 2) + 0
					if (pct > limit) printf "%s  %s\n", $1, $i
				}
			}
		}
	' "$1"
}

write_summary() {
	local dir=$1 reg
	reg=$(<"$dir/regressions.txt")
	{
		if [ -n "$reg" ]; then
			echo "### Benchmarks: slower by more than ${THRESHOLD}%"
			echo
			echo '```'
			echo "$reg"
			echo '```'
		else
			echo "### Benchmarks: nothing slower by more than ${THRESHOLD}%"
		fi
		echo
		echo "<details><summary>benchstat, base against this change</summary>"
		echo
		echo '```'
		cat "$dir/benchstat.txt"
		echo '```'
		echo
		echo "</details>"
	} >"$dir/summary.md"
}

mode=${1:-}
case "$mode" in
compare)
	base_ref=${2:?usage: ci-bench.sh compare <base-ref> <out-dir>}
	out=${3:?usage: ci-bench.sh compare <base-ref> <out-dir>}
	root=$(git rev-parse --show-toplevel)
	mkdir -p "$out"
	out=$(cd "$out" && pwd)
	work=$(mktemp -d)
	trap 'git -C "$root" worktree remove --force "$work/base" >/dev/null 2>&1 || true; rm -rf "$work"' EXIT
	git -C "$root" worktree add --detach "$work/base" "$base_ref" >/dev/null
	build_side "$work/base" "$work/bin-base"
	build_side "$root" "$work/bin-head"
	: >"$out/base.txt"
	: >"$out/head.txt"
	for _ in $(seq "$ROUNDS"); do
		run_side "$work/base" "$work/bin-base" "$out/base.txt"
		run_side "$root" "$work/bin-head" "$out/head.txt"
	done
	(cd "$out" && benchstat base.txt head.txt) >"$out/benchstat.txt" 2>&1 || true
	find_regressions "$out/benchstat.txt" >"$out/regressions.txt"
	write_summary "$out"
	cat "$out/summary.md"
	;;
single)
	out=${2:?usage: ci-bench.sh single <out-dir>}
	root=$(git rev-parse --show-toplevel)
	mkdir -p "$out"
	out=$(cd "$out" && pwd)
	work=$(mktemp -d)
	trap 'rm -rf "$work"' EXIT
	build_side "$root" "$work/bin"
	: >"$out/head.txt"
	for _ in $(seq "$ROUNDS"); do
		run_side "$root" "$work/bin" "$out/head.txt"
	done
	(cd "$out" && benchstat head.txt) >"$out/benchstat.txt" 2>&1 || true
	{
		echo "### Benchmarks on this commit"
		echo
		echo '```'
		cat "$out/benchstat.txt"
		echo '```'
	} >"$out/summary.md"
	cat "$out/summary.md"
	;;
*)
	echo "usage: scripts/ci-bench.sh compare <base-ref> <out-dir> | single <out-dir>" >&2
	exit 2
	;;
esac
