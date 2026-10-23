package smoke

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The lineage badge scale (config.BadgeFamilies, family "lineage"): how a
// lineage's top rung is derived from the reference model, so the ladders can
// be derived again after a balance change.
//
// The dedicated player is the reference lingering player (RefLinger). He buys
// new copies and never upgrades (an upgrade is not a built copy), and he does
// not stop at a knowledge building's full-rate cap (past it a copy still
// counts as built and researches at half rate). He plays the four-run plan,
// each run starting from nothing: a run that ends on entering the Renaissance
// Age, one on entering the Digital Age, one on entering the Fusion Age, then
// a full clear, and full clears after that, 25 runs in all
// (config.BadgeMaxRuns). A lineage's total is the copies of it he builds
// over those runs. The top rung is that total to two significant figures; each
// rung below is lineageRungRatio of the one above, also to two figures.

// lineageRungRatio is how much of a rung the rung below it asks for.
const lineageRungRatio = 0.4

// lineageRungCount is how many rungs a lineage ladder has.
const lineageRungCount = 5

// lineagePlanWalls are the ages the first runs of the plan end on entering;
// the run after them is a full clear.
var lineagePlanWalls = []string{"renaissance_age", "digital_age", "fusion_age"}

// round2 rounds x to two significant figures.
func round2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	f := math.Pow(10, math.Floor(math.Log10(x))-1)
	return math.Round(x/f) * f
}

// lingerWalk sends the player through every age, staying linger times as long
// as it took to open the gate, and returns each age's stay as it ended.
func lingerWalk(t *refTables, linger float64) (out []refRun) {
	counts := map[string]int{}
	for i := range t.ages {
		p := refPlay{wonder: true, rate: 1, speed: 1, horizon: 40 * config.Million}
		p.leave = func(gate, now float64) bool { return now >= linger*gate }
		run := t.stay(i, counts, p)
		counts = run.counts
		out = append(out, run)
	}
	return out
}

// lineageTotals is the copies the dedicated player builds of each lineage
// over the badge horizon, by lineage, and what a full clear builds.
func lineageTotals() (total, clear map[string]float64) {
	tb := newRefTables()
	tb.upgrades = false
	for i := range tb.byAge { // past the knowledge cap
		for j := range tb.byAge[i] {
			if tb.byAge[i][j].MaxCount == 0 {
				tb.byAge[i][j].FullRateCopies = 0
			}
		}
	}
	runs := lingerWalk(tb, RefLinger)
	// built[l][i] is the copies of lineage l bought in age i (no upgrades, so
	// a building's count only grows).
	built := map[string][]float64{}
	prev := map[string]int{}
	for i, r := range runs {
		for k, n := range r.counts {
			l := tb.defs[k].LineageKey
			if l == "" {
				continue
			}
			if built[l] == nil {
				built[l] = make([]float64, len(runs))
			}
			built[l][i] += float64(n - prev[k])
		}
		prev = r.counts
	}
	total, clear = map[string]float64{}, map[string]float64{}
	for l, per := range built {
		through := func(last int) (s float64) {
			for i := 0; i <= last; i++ {
				s += per[i]
			}
			return s
		}
		for _, wall := range lineagePlanEnds(tb) {
			total[l] += through(wall - 1)
		}
		clear[l] = through(len(per) - 1)
		runsLeft := config.BadgeMaxRuns - len(lineagePlanWalls)
		total[l] += float64(runsLeft) * clear[l]
	}
	return total, clear
}

// lineagePlanEnds are the places of the ages the plan's first runs end on.
func lineagePlanEnds(t *refTables) []int {
	var out []int
	for _, w := range lineagePlanWalls {
		out = append(out, t.idx[w])
	}
	return out
}

// lineageLadder is a lineage's rungs, lowest first, from its total.
func lineageLadder(total float64) []float64 {
	top := round2(total)
	rungs := make([]float64, lineageRungCount)
	for n := range rungs {
		rungs[n] = round2(top * math.Pow(lineageRungRatio, float64(lineageRungCount-1-n)))
	}
	return rungs
}

// TestLineageDerivation prints the lineage ladders the reference model gives,
// in the form config/badge_catalog.go holds them. Run it after a balance
// change that moves the gates or prices:
//
//	LINEAGE_DERIVE=1 go test ./smoke -run TestLineageDerivation -v
func TestLineageDerivation(t *testing.T) {
	if os.Getenv("LINEAGE_DERIVE") != "1" {
		t.Skip("set LINEAGE_DERIVE=1 to print the derived lineage ladders")
	}
	total, clear := lineageTotals()
	var keys []string
	for l := range total {
		keys = append(keys, l)
	}
	sort.Strings(keys)
	for _, l := range keys {
		var parts []string
		for _, r := range lineageLadder(total[l]) {
			parts = append(parts, config.GoLiteral(r))
		}
		fmt.Printf("\t\t\t\t%-24s {%s}, // a full clear builds %.0f, %d runs build %.0f, %.1fx the old top rung\n",
			fmt.Sprintf("%q:", l), strings.Join(parts, ", "), clear[l], config.BadgeMaxRuns, total[l], total[l]/lineageLegendaryBefore[l])
	}
}
