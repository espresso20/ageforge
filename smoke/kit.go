package smoke

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Depth points and the legacy kit in the smoke suite (Pacing v2, PR 6).
//
//   - Config.Kit gives a run the veteran's legacy kit: every item bought, and
//     the canned memory in testdata/veteran_kit.json (applyKit). The veteran
//     check-in runs use it.
//   - The canned memory is a veteran's last run: the active bot on the
//     veteran preset, with every build it made (an upgrade as a build of the
//     new tier: the plan can't upgrade), every trade (as a trade item buying
//     what it got that age) and every advance written to the plan log, in
//     the order it made them (Config.DumpLegacy, -dump-legacy). Regenerate
//     it when content moves:
//
//     go run ./cmd/smoke -scenario veteran -tier full -seeds 1 -dump-legacy smoke/testdata/veteran_kit.json
//
//   - Depth pays (DepthRow): points per day of a first run that prestiges
//     from the Medieval Age, the Modern Age, and the Cyberpunk Age (a run
//     through the Digital Age), from the first cycle's age times. Each must
//     beat the one before by DepthPaysMin.
//   - Kit carry-over (kitCarryProblems, checkKitOnAdvance): with the kit
//     bought, a prestige keeps the shares and remembers the run, a new age
//     gets its template slice, and remembered civilizations are met at
//     their age.

//go:embed testdata/veteran_kit.json
var veteranKitJSON []byte

// veteranKit is the canned kit memory.
func veteranKit() (game.LegacyKit, error) {
	var k game.LegacyKit
	if err := json.Unmarshal(veteranKitJSON, &k); err != nil {
		return k, fmt.Errorf("testdata/veteran_kit.json: %w", err)
	}
	return k, nil
}

// applyKit gives ge the veteran's kit when on: every item bought and the
// canned memory, put to work as a purchase would.
func applyKit(ge *game.GameEngine, on bool) error {
	if !on {
		return nil
	}
	k, err := veteranKit()
	if err != nil {
		return err
	}
	ge.SetLegacyForTest(k, true)
	return nil
}

// dumpLegacy writes what ge's kit remembers to path, as JSON.
func dumpLegacy(ge *game.GameEngine, path string) error {
	data, err := json.MarshalIndent(ge.LegacyForTest(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ===== Depth pays =====

// The prestige ages DepthRow compares, shallow to deep: an early taste, a
// full run, and a run through the Digital Age.
var depthAges = []string{"medieval_age", "modern_age", "cyberpunk_age"}

// DepthPaysMin is how many times more points per day each deeper prestige
// must pay than the one before.
const DepthPaysMin = 1.25

// DepthPoint is one prestige age's points per day on a first run.
type DepthPoint struct {
	Age string `json:"age"`
	// Points is what a prestige there pays; Secs the median 1x time of a
	// first run to reach it (from cycle 1's age times), over Samples runs.
	Points    int     `json:"points"`
	Secs      float64 `json:"seconds_1x"`
	Samples   int     `json:"samples"`
	PerDay    float64 `json:"points_per_day"`
	VsShallow float64 `json:"vs_previous,omitempty"`
}

// DepthRow is the "depth pays" check over a set's first runs.
type DepthRow struct {
	Points []DepthPoint `json:"points"`
	Failed bool         `json:"failed,omitempty"`
}

// firstRunTo is the 1x time run r's first cycle took to enter age, and
// whether it did (Succumb replays included).
func firstRunTo(r *RunResult, order map[string]int, age string) (float64, bool) {
	secs := 0.0
	for _, a := range r.Ages {
		if a.Cycle != 1 {
			continue
		}
		if order[a.Age] >= order[age] {
			return secs, true
		}
		secs += a.Seconds
	}
	return secs, false
}

// newDepth measures points per day at each depth age a set's first runs
// reached (new players only: a preset's first run is not a first run).
// nil when fewer than two depths were reached.
func newDepth(cfg Config, runs []*RunResult, order map[string]int) *DepthRow {
	if cfg.Preset != "" || cfg.Kit {
		return nil
	}
	row := &DepthRow{}
	for _, age := range depthAges {
		var got []float64
		for _, r := range runs {
			if secs, ok := firstRunTo(r, order, age); ok && secs > 0 {
				got = append(got, secs)
			}
		}
		if len(got) == 0 {
			continue
		}
		_, med, _ := spread(got)
		p := DepthPoint{Age: age, Points: config.DepthPoints(age), Secs: med, Samples: len(got)}
		p.PerDay = float64(p.Points) / (med / 86400)
		if n := len(row.Points); n > 0 && row.Points[n-1].PerDay > 0 {
			p.VsShallow = p.PerDay / row.Points[n-1].PerDay
			if p.VsShallow < DepthPaysMin {
				row.Failed = true
			}
		}
		row.Points = append(row.Points, p)
	}
	if len(row.Points) < 2 {
		return nil
	}
	return row
}

// writeDepth renders the depth-pays table.
func (s *Summary) writeDepth(sb *strings.Builder) {
	d := s.Depth
	if d == nil {
		return
	}
	verdict := VerdictOK
	if d.Failed {
		verdict = VerdictSlow
	}
	fmt.Fprintf(sb, "\nDepth pays (points per day of a first run that prestiges there; each must beat the one before by %gx): %s.\n\n", DepthPaysMin, verdictMark(verdict))
	sb.WriteString("| prestige from | points | first run to it (median) | runs | points per day | vs the one before |\n|---|---|---|---|---|---|\n")
	for _, p := range d.Points {
		vs := "-"
		if p.VsShallow > 0 {
			vs = fmt.Sprintf("%.2fx", p.VsShallow)
		}
		fmt.Fprintf(sb, "| %s | %d | %s | %d | %.1f | %s |\n", p.Age, p.Points, days(p.Secs), p.Samples, p.PerDay, vs)
	}
}

// ===== Kit carry-over =====

// kitOwned reports whether st has kit item key.
func kitOwned(st game.GameState, key string) bool { return st.Prestige.Upgrades[key].Tier > 0 }

// kitCarryProblems checks the legacy kit across a prestige: the kit stays
// bought, the run's research order and the civilizations met are
// remembered, Worker Shares keeps the shares, and Plan Template starts the
// new run's plan from the first age's slice.
func kitCarryProblems(before, after game.GameState) []problem {
	var out []problem
	for _, key := range config.LegacyKit() {
		if kitOwned(before, key) && !kitOwned(after, key) {
			out = append(out, problem{"prestige_lost_kit", fmt.Sprintf("the legacy kit's %s did not survive prestige", key)})
		}
	}
	bk, ak := before.Prestige.Kit, after.Prestige.Kit
	if ak.ResearchTechs < bk.ResearchTechs || ak.ResearchTechs < before.Research.TotalResearched {
		out = append(out, problem{"kit_research_memory",
			fmt.Sprintf("the remembered research order holds %d techs after a prestige (before: %d, the run researched %d)", ak.ResearchTechs, bk.ResearchTechs, before.Research.TotalResearched)})
	}
	met := 0
	for _, f := range before.Diplomacy.Factions {
		if f.Discovered {
			met++
		}
	}
	if ak.Factions < bk.Factions || ak.Factions < met {
		out = append(out, problem{"kit_factions_memory",
			fmt.Sprintf("%d civilizations remembered after a prestige (before: %d, the run met %d)", ak.Factions, bk.Factions, met)})
	}
	if kitOwned(before, config.LegacyWorkers) && len(before.Workers.Shares) > 0 && !sameShares(before.Workers.Shares, after.Workers.Shares) {
		out = append(out, problem{"kit_shares", fmt.Sprintf("Worker Shares owned, but the shares went %v -> %v at a prestige", before.Workers.Shares, after.Workers.Shares)})
	}
	if kitOwned(before, config.LegacyPlan) && ak.PlanByAge[after.Age] > 0 && ak.PlanAppliedAge != after.Age {
		out = append(out, problem{"kit_template", fmt.Sprintf("Plan Template owned with %d items for %s, but the new run's plan did not get them", ak.PlanByAge[after.Age], after.Age)})
	}
	return out
}

func sameShares(a, b map[string]float64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// checkKitOnAdvance checks the kit after the runner's own advance: the new
// age's template slice went into the plan, and every remembered
// civilization of this age or earlier has been met.
func (r *runner) checkKitOnAdvance(st game.GameState) {
	k := st.Prestige.Kit
	if kitOwned(st, config.LegacyPlan) && k.PlanByAge[st.Age] > 0 && k.PlanAppliedAge != st.Age {
		r.anomaly(KindInvariant, "kit_template", fmt.Sprintf("Plan Template owned with %d items for %s, but the advance did not add them", k.PlanByAge[st.Age], st.Age), st, true)
	}
	if !kitOwned(st, config.LegacyFactions) {
		return
	}
	ages := config.AgeByKey()
	var missing []string
	for _, key := range k.FactionKeys {
		def, ok := config.FactionByKey()[key]
		if !ok || ages[def.MinAge].Order > ages[st.Age].Order {
			continue
		}
		if f := st.Diplomacy.Factions[key]; !f.Discovered {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		r.anomaly(KindInvariant, "kit_old_friends", fmt.Sprintf("Old Friends owned, but %s not met again in %s", strings.Join(missing, ", "), st.Age), st, true)
	}
}
