package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestVeteranKitCanned: the canned kit memory parses, spans the ages a
// veteran's run to the Modern Age passes, ends each age's part with its
// advance, and plans no research (the bot researches for itself).
func TestVeteranKitCanned(t *testing.T) {
	k, err := veteranKit()
	if err != nil {
		t.Fatal(err)
	}
	ages := map[string]int{}
	last := map[string]string{}
	for _, it := range k.Plan {
		ages[it.Age]++
		last[it.Age] = it.Kind
	}
	for _, a := range config.AgeOrder() {
		if a == game.PrestigeRunAge {
			break
		}
		if ages[a] == 0 {
			t.Errorf("the canned template has nothing for %s", a)
			continue
		}
		if last[a] != game.PlanAdvance {
			t.Errorf("the canned template's part for %s ends with %s, not its advance", a, last[a])
		}
		if ages[a] > game.MaxPlanItems {
			t.Errorf("the canned template holds %d items for %s, over the plan's %d", ages[a], a, game.MaxPlanItems)
		}
	}
	for _, it := range k.Plan {
		if it.Kind == game.PlanResearch {
			t.Errorf("the canned template plans research (%s in %s); the bot researches for itself", it.Key, it.Age)
		}
	}
	if len(k.Factions) == 0 {
		t.Error("the canned kit remembers no civilizations")
	}
}

// TestDepthRow: points per day by prestige age from first runs, each deeper
// one graded against DepthPaysMin times the one before.
func TestDepthRow(t *testing.T) {
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	// A first run whose ages each take a day until the Medieval Age and then
	// enough that the Modern Age comes on day toModern.
	run := func(toModern float64) *RunResult {
		r := &RunResult{}
		for _, a := range config.AgeOrder() {
			if a == "medieval_age" {
				break
			}
			r.Ages = append(r.Ages, AgeSplit{Cycle: 1, Age: a, Seconds: 86400 * 0.81 / 5})
		}
		mid := 0
		for _, a := range config.AgeOrder() {
			if order[a] >= order["medieval_age"] && order[a] < order["modern_age"] {
				mid++
			}
		}
		for _, a := range config.AgeOrder() {
			if order[a] >= order["medieval_age"] && order[a] < order["modern_age"] {
				r.Ages = append(r.Ages, AgeSplit{Cycle: 1, Age: a, Seconds: 86400 * (toModern - 0.81) / float64(mid)})
			}
		}
		r.Ages = append(r.Ages, AgeSplit{Cycle: 1, Age: "modern_age", Prestiged: true})
		return r
	}
	d := newDepth(Config{}, []*RunResult{run(5.3), run(5.3), run(5.3)}, order)
	if d == nil || len(d.Points) != 2 || d.Failed {
		t.Fatalf("depth row %+v, want Medieval and Modern, passing", d)
	}
	if p := d.Points[0]; p.Age != "medieval_age" || p.Points != 9 || p.PerDay < 11 || p.PerDay > 11.2 {
		t.Errorf("medieval point %+v, want 9 points at about 11.1 a day", p)
	}
	if p := d.Points[1]; p.Points != 120 || p.VsShallow < 2 {
		t.Errorf("modern point %+v, want 120 points, about twice the Medieval rate", p)
	}
	// A Modern Age that took 12 days pays less per day than the taste.
	if d := newDepth(Config{}, []*RunResult{run(12)}, order); d == nil || !d.Failed {
		t.Errorf("a 12-day first run should fail depth pays: %+v", d)
	}
	if newDepth(Config{Preset: PresetVeteran}, []*RunResult{run(5.3)}, order) != nil {
		t.Error("a preset's first run graded as a first run")
	}
}

// TestStaticDepth: the depth weights, the docs' table, the taste's share
// and the first kit price all hold.
func TestStaticDepth(t *testing.T) {
	if p := StaticDepth(); len(p) != 0 {
		t.Errorf("static depth problems: %v", p)
	}
}

// TestKitCarryProblems: a lost kit item, unremembered civilizations,
// dropped shares and a missing template slice are caught.
func TestKitCarryProblems(t *testing.T) {
	owned := func() map[string]game.PrestigeUpgradeState {
		m := map[string]game.PrestigeUpgradeState{}
		for _, k := range config.LegacyKit() {
			m[k] = game.PrestigeUpgradeState{Tier: 1, MaxTier: 1, Kit: true}
		}
		return m
	}
	before := game.GameState{Age: "modern_age"}
	before.Prestige.Upgrades = owned()
	before.Workers.Shares = map[string]float64{"food": 40}
	before.Diplomacy.Factions = map[string]game.FactionInfo{"riverlands_tribes": {Discovered: true}}
	good := game.GameState{Age: "primitive_age"}
	good.Prestige.Upgrades = owned()
	good.Prestige.Kit = game.LegacyKitState{Factions: 1, PlanByAge: map[string]int{"primitive_age": 3}, PlanAppliedAge: "primitive_age"}
	good.Workers.Shares = map[string]float64{"food": 40}
	if p := kitCarryProblems(before, good); len(p) != 0 {
		t.Errorf("a clean carry-over reported %v", p)
	}
	bad := good
	bad.Prestige.Upgrades = map[string]game.PrestigeUpgradeState{}
	bad.Prestige.Kit = game.LegacyKitState{PlanByAge: map[string]int{"primitive_age": 3}}
	bad.Workers.Shares = nil
	got := map[string]bool{}
	for _, p := range kitCarryProblems(before, bad) {
		got[p.check] = true
	}
	for _, c := range []string{"prestige_lost_kit", "kit_factions_memory", "kit_shares", "kit_template"} {
		if !got[c] {
			t.Errorf("a broken carry-over did not report %s (got %v)", c, got)
		}
	}
}
