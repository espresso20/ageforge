package game

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestMasteryKTable: k = 1 + √m, the frontier at 1, catch-up raises an age
// six or more behind the record to 4 unless its own k is higher.
func TestMasteryKTable(t *testing.T) {
	want := []float64{1, 2, 1 + math.Sqrt(2), 1 + math.Sqrt(3), 3, 1 + math.Sqrt(5), 1 + math.Sqrt(6), 1 + math.Sqrt(7), 1 + math.Sqrt(8), 4, 1 + math.Sqrt(10)}
	for m, w := range want {
		if got := config.MasteryK(m); got != w {
			t.Errorf("MasteryK(%d) = %v, want %v", m, got, w)
		}
	}
	if config.MasteryK(-3) != 1 || config.MasteryK(99) != want[10] {
		t.Errorf("MasteryK is not clamped to 0..%d", config.MasteryCap)
	}
	for _, c := range []struct {
		m, behind int
		want      float64
	}{
		{0, -1, 1}, {0, 0, 1}, {0, 5, 1}, {0, 6, 4}, {1, 6, 4}, {9, 6, 4}, {10, 6, want[10]}, {10, 0, want[10]}, {4, 20, 4},
	} {
		if got := config.AgeSpeed(c.m, c.behind); got != c.want {
			t.Errorf("AgeSpeed(m=%d, behind=%d) = %v, want %v", c.m, c.behind, got, c.want)
		}
	}
	if config.CatchUpK != 4 || config.CatchUpGap != 6 || config.MasteryCap != 10 {
		t.Errorf("the mastery constants moved: cap %d, gap %d, catch-up %v", config.MasteryCap, config.CatchUpGap, config.CatchUpK)
	}
}

// TestMasteryTicks: times divide by k, rounded up, never below a tick.
func TestMasteryTicks(t *testing.T) {
	for _, c := range []struct {
		ticks int
		k     float64
		want  int
	}{
		{100, 1, 100}, {100, 0, 100}, {101, 2, 51}, {100, 2, 50}, {1, 4, 1}, {3, 4, 1}, {100, config.MasteryK(10), 25}, {0, 2, 0},
	} {
		if got := MasteryTicks(c.ticks, c.k); got != c.want {
			t.Errorf("MasteryTicks(%d, %v) = %d, want %d", c.ticks, c.k, got, c.want)
		}
	}
	if SpeedText(2) != "2x" || SpeedText(1+math.Sqrt(2)) != "2.4x" || SpeedText(config.MasteryK(10)) != "4.2x" {
		t.Errorf("SpeedText: %s %s %s", SpeedText(2), SpeedText(1+math.Sqrt(2)), SpeedText(config.MasteryK(10)))
	}
}

// masteryTwins are two engines in the same state, one with mastery m in the
// current age (record there too, so no catch-up).
func masteryTwins(t *testing.T, m int) (base, known *GameEngine) {
	t.Helper()
	base, known = newSeededEngine(7), newSeededEngine(7)
	for _, ge := range []*GameEngine{base, known} {
		ge.StepTicks(60)
		// A few of every unlocked building, so there is output to scale.
		ge.mu.Lock()
		for _, key := range sortedKeys(ge.Buildings.defs) {
			if def := ge.Buildings.defs[key]; ge.Buildings.IsUnlocked(key) && def.Category != "wonder" {
				ge.Buildings.counts[key] = 2
			}
		}
		ge.mu.Unlock()
	}
	known.SetMasteryForTest(map[string]int{known.age: m}, known.age)
	base.SetMasteryForTest(nil, base.age)
	return base, known
}

// TestMasteryTwinEngines: at m = 1 (k = 2) against m = 0 in the same age,
// every rate is exactly doubled, every store doubled, and build and
// research times halved (rounded up; research after the speed step).
func TestMasteryTwinEngines(t *testing.T) {
	base, known := masteryTwins(t, 1)
	if k := known.GetState().Mastery.K; k != 2 {
		t.Fatalf("mastery 1 runs at %v, want 2", k)
	}
	bs, ks := base.GetState(), known.GetState()
	rated := 0
	for _, key := range sortedKeys(bs.Resources) {
		b, k := bs.Resources[key], ks.Resources[key]
		if k.Rate != 2*b.Rate {
			t.Errorf("%s rate %v at k = 2 against %v at k = 1, want exactly double", key, k.Rate, b.Rate)
		}
		if b.Rate != 0 {
			rated++
			if want := b.Rate; k.Breakdown.MasteryRate != want {
				t.Errorf("%s mastery breakdown line %v, want %v", key, k.Breakdown.MasteryRate, want)
			}
		}
		if k.Storage != 2*b.Storage {
			t.Errorf("%s storage %v at k = 2 against %v, want double", key, k.Storage, b.Storage)
		}
	}
	if rated == 0 {
		t.Fatal("no resource has a rate: the twins prove nothing")
	}

	// Build times: an odd BuildTicks halves rounded up.
	for _, key := range sortedKeys(base.Buildings.defs) {
		def := base.Buildings.defs[key]
		if def.BuildTicks > 0 {
			if got, want := known.buildTicksLocked(def), (def.BuildTicks+1)/2; got != want {
				t.Errorf("%s builds in %d ticks at k = 2, want %d (base %d)", key, got, want, def.BuildTicks)
			}
			if base.buildTicksLocked(def) != def.BuildTicks {
				t.Errorf("%s: the frontier changed its build time", key)
			}
		}
	}

	// Research: the same tech on both, after the same speed step.
	tech := firstPrimitiveTech(t)
	for _, ge := range []*GameEngine{base, known} {
		ge.mu.Lock()
		ge.age = tech.Age
		ge.Resources.UnlockResource("knowledge")
		ge.Resources.resources["knowledge"].Storage = tech.Cost * 4
		ge.Resources.resources["knowledge"].Amount = tech.Cost * 2
		err := ge.startResearchLocked(tech.Key, true)
		ge.mu.Unlock()
		if err != nil {
			t.Fatalf("research %s: %v", tech.Key, err)
		}
	}
	if b, k := base.Research.totalTicks, known.Research.totalTicks; k != MasteryTicks(b, 2) || k >= b {
		t.Errorf("research takes %d ticks at k = 2 against %d at k = 1, want %d", k, b, MasteryTicks(b, 2))
	}
}

// firstPrimitiveTech is a tech with no prerequisites from the first age.
func firstPrimitiveTech(t *testing.T) config.TechDef {
	t.Helper()
	var keys []string
	defs := config.TechByKey()
	for k, d := range defs {
		if d.Age == "primitive_age" && len(d.Prerequisites) == 0 {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		t.Fatal("no primitive tech without prerequisites")
	}
	sort.Strings(keys)
	return defs[keys[0]]
}

// TestMasteryCommitAtPrestige: a prestige raises every age below the run's
// furthest by one, up to the cap; the age prestiged from and later ones are
// untouched. Succumb commits nothing; Reset clears everything.
func TestMasteryCommitAtPrestige(t *testing.T) {
	pm := NewPrestigeManager()
	pm.NoteAgeEntered("modern_age")
	for i := 1; i <= 12; i++ {
		pm.NoteAgeEntered("modern_age")
		gained := pm.CommitRun()
		if i <= config.MasteryCap && len(gained) != 12 {
			t.Fatalf("prestige %d raised %d ages, want 12 (Primitive to Atomic)", i, len(gained))
		}
		if i > config.MasteryCap && len(gained) != 0 {
			t.Fatalf("prestige %d raised %v past the cap", i, gained)
		}
		for _, a := range ageKeys() {
			want := 0
			if ageOrders()[a] < ageOrders()["modern_age"] {
				want = min(i, config.MasteryCap)
			}
			if pm.Mastery(a) != want {
				t.Fatalf("after %d prestiges %s has mastery %d, want %d", i, a, pm.Mastery(a), want)
			}
		}
	}
	if pm.Record() != "modern_age" || pm.RunFurthest() != "primitive_age" {
		t.Errorf("record %q, run furthest %q after the prestiges", pm.Record(), pm.RunFurthest())
	}

	// On the engine: a prestige from the Modern Age commits; a Succumb does not.
	ge := newSeededEngine(3)
	if err := ge.SummonHarbingerForTest("modern_age"); err != nil {
		t.Log(err) // places the game; a missing harbinger is fine
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if m := ge.Prestige.Mastery("atomic_age"); m != 1 {
		t.Fatalf("the Atomic Age has mastery %d after a Modern prestige, want 1", m)
	}
	if k := ge.GetState().Mastery.K; k != 4 {
		t.Errorf("the new run's Primitive Age runs at %v; with the record 12 ages ahead it catches up at 4", k)
	}
	before := ge.Prestige.masterySave()
	_ = ge.SummonHarbingerForTest("iron_age")
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatal(err)
	}
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	if after := ge.Prestige.masterySave(); !sameMastery(before, after) {
		t.Errorf("Succumb changed mastery: %v -> %v", before, after)
	}
	ge.Reset()
	if len(ge.Prestige.masterySave()) != 0 || ge.Prestige.Record() != "primitive_age" || ge.GetState().Mastery.K != 1 {
		t.Errorf("Reset kept mastery %v, record %q", ge.Prestige.masterySave(), ge.Prestige.Record())
	}
}

func sameMastery(a, b map[string]int) bool {
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

// TestMasteryPassiveRetired: a prestige with no upgrades bought leaves the
// new run's production and tick speed untouched by the old passive.
func TestMasteryPassiveRetired(t *testing.T) {
	ge := newSeededEngine(5)
	_ = ge.SummonHarbingerForTest("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	if st.TickSpeedBonus != 0 || st.Prestige.PassiveBonus != 0 {
		t.Errorf("after a prestige: tick speed %v, passive %v; the passive retired into Era Mastery", st.TickSpeedBonus, st.Prestige.PassiveBonus)
	}
	for _, m := range st.Modifiers {
		if m.Source == "prestige" {
			t.Errorf("a prestige modifier with no upgrades bought: %+v", m)
		}
	}
}

// TestMasteryGraceRule: when k drops, stock above the new cap survives,
// stops growing, can be spent, loses its grace once under the cap, and the
// grace survives a save and load.
func TestMasteryGraceRule(t *testing.T) {
	isolateAccountDir(t)
	_, ge := masteryTwins(t, 1)
	ge.mu.Lock()
	ge.recalculateRates()
	r := ge.Resources.resources["wood"]
	high := r.Storage // the k = 2 cap
	r.Amount = high
	ge.Prestige.SetMastery(ge.age, 0) // k drops to 1 (as entering new ground does)
	ge.recalculateRates()
	cap1 := r.Storage
	ge.mu.Unlock()
	if cap1 >= high || r.Amount != high || !ge.Resources.Graced("wood") {
		t.Fatalf("after k dropped: wood %v (cap %v, was %v), graced %v; want the stock kept and graced", r.Amount, cap1, high, ge.Resources.Graced("wood"))
	}
	if !ge.GetState().Resources["wood"].OverCapGrace {
		t.Error("the snapshot does not show the grace")
	}
	ge.StepTicks(20)
	if r := ge.Resources.resources["wood"]; r.Amount > high {
		t.Errorf("graced wood grew to %v past %v", r.Amount, high)
	}

	// Save and load keep it.
	if err := ge.SaveGame("grace"); err != nil {
		t.Fatal(err)
	}
	if g := readSaveFromDisk(t, "grace").OverCapGrace; !g["wood"] {
		t.Errorf("the save's grace set is %v", g)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("grace"); err != nil {
		t.Fatal(err)
	}
	lr := loaded.Resources.resources["wood"]
	if !loaded.Resources.Graced("wood") || lr.Amount <= lr.Storage {
		t.Fatalf("after load: wood %v, cap %v, graced %v", lr.Amount, lr.Storage, loaded.Resources.Graced("wood"))
	}

	// Spending works; under the cap the grace ends and the cap holds again.
	loaded.mu.Lock()
	if !loaded.Resources.Remove("wood", lr.Amount-lr.Storage/2) {
		t.Fatal("could not spend graced wood")
	}
	loaded.recalculateRates()
	graced := loaded.Resources.Graced("wood")
	loaded.Resources.Add("wood", lr.Storage*3)
	loaded.mu.Unlock()
	if graced {
		t.Error("wood under its cap is still graced")
	}
	if lr.Amount != lr.Storage {
		t.Errorf("wood %v after a big add, want the cap %v", lr.Amount, lr.Storage)
	}
}

// TestMasteryFateWindow: in a mastered era the fate's window and the
// harbinger's lead divide by k, so the strike falls inside the era as it is
// played, not at its last transition.
func TestMasteryFateWindow(t *testing.T) {
	ge := newSeededEngine(11)
	all := map[string]int{}
	for _, a := range ageKeys() {
		all[a] = config.MasteryCap
	}
	ge.SetMasteryForTest(all, "transcendent_age")
	k := config.MasteryK(config.MasteryCap)
	for _, a := range config.EpochByKey()["iron_era"].Ages {
		if got, want := ge.expectedAgeTicks(a), baseAgeTicks(a)/k; got != want {
			t.Errorf("%s expected %v ticks, want %v (÷ %v)", a, got, want, k)
		}
	}
	era := ge.expectedEraTicks("iron_era")
	if math.Abs(era-baseEraTicks("iron_era")/k) > 1 {
		t.Errorf("the Iron Era is expected to last %.0f ticks, want %.0f (÷ %.2f)", era, baseEraTicks("iron_era")/k, k)
	}
	// Rolled fates land inside the mastered era.
	for seed := int64(1); seed <= 40; seed++ {
		g := newSeededEngine(seed)
		g.SetMasteryForTest(all, "transcendent_age")
		g.mu.Lock()
		g.currentEpoch = "iron_era"
		g.age = "iron_age"
		g.rollFate()
		fs := *g.fate
		g.mu.Unlock()
		if want := int(math.Round(era)); fs.Window != want {
			t.Fatalf("seed %d: the window is %d ticks, want the mastered era's %d", seed, fs.Window, want)
		}
		if fs.Fated && float64(fs.StrikeTick-fs.EntryTick) > era {
			t.Errorf("seed %d: the strike falls %d ticks into an era of %.0f", seed, fs.StrikeTick-fs.EntryTick, era)
		}
		lead := fs.LeadFrac * g.expectedAgeTicks("iron_age")
		if fs.Fated && lead > float64(harbingerLeadMax*baseAgeTicks("iron_age")/k)+1 {
			t.Errorf("seed %d: the harbinger's lead %.0f ticks is not divided by k", seed, lead)
		}
	}
}

// oldSave writes ge's state as a build from before Era Mastery wrote and
// signed it: no mastery fields at all.
func oldSave(t *testing.T, ge *GameEngine, name string) {
	t.Helper()
	ge.mu.RLock()
	old := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	old.Prestige.Mastery, old.Prestige.Furthest, old.Prestige.RunFurthest, old.Prestige.MasterySeeded = nil, "", "", false
	old.OverCapGrace = nil
	// Nor the legacy kit (Pacing v2, PR 6): the first shop, no kit memory.
	old.Prestige.ShopVersion = 0
	old.Prestige.LegacyPlan, old.Prestige.LegacyResearch, old.Prestige.LegacyFactions, old.Prestige.LegacyShares = nil, nil, nil, nil
	old.PlanLog, old.Research.Order = nil, nil
	old.Signature = signSave(old, saveHMACKey)
	data, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"mastery"`, `"furthest"`, `"run_furthest"`, `"mastery_seeded"`, `"over_cap_grace"`,
		`"shop_version"`, `"legacy_plan"`, `"legacy_research"`, `"legacy_factions"`, `"legacy_shares"`, `"plan_log"`} {
		if strings.Contains(string(data), field) {
			t.Fatalf("an old save still writes %s: the field must be omitempty", field)
		}
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath(name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func masteryLines(ge *GameEngine) []string {
	var out []string
	for _, l := range ge.log {
		if strings.HasPrefix(l.Message, "Era Mastery: ages you have completed") {
			out = append(out, l.Message)
		}
	}
	return out
}

// TestMasterySeedingLevel5: a signed level-5 save from before Era Mastery,
// mid-run in the Industrial Age, still verifies, and gets mastery 5 from
// the Primitive to the Atomic Age, the Modern Age as its record and the
// Industrial Age as the run's furthest, once, with one log line. Its next
// save carries the marker, so loading it again changes nothing.
func TestMasterySeedingLevel5(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(21)
	ge.StepTicks(30)
	ge.mu.Lock()
	ge.Prestige.LoadState(5, 86, 3, map[string]int{"gather_boost": 3, "storage_bonus": 3, "starting_food": 5})
	ge.age = "industrial_age"
	ge.currentEpoch = config.EpochForAge("industrial_age")
	ge.Stats.AgesReached = []string{"primitive_age", "stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age", "renaissance_age", "colonial_age", "industrial_age"}
	ge.mu.Unlock()
	oldSave(t, ge, "level5")

	loaded := NewGameEngine()
	if err := loaded.LoadGame("level5"); err != nil {
		t.Fatal(err)
	}
	if loaded.cheaterBadge {
		t.Fatal("the old signed save failed its signature check")
	}
	for _, a := range ageKeys() {
		want := 0
		if ageOrders()[a] < ageOrders()["modern_age"] {
			want = 5
		}
		if m := loaded.Prestige.Mastery(a); m != want {
			t.Errorf("%s mastery %d after seeding, want %d", a, m, want)
		}
	}
	if r, f := loaded.Prestige.Record(), loaded.Prestige.RunFurthest(); r != "modern_age" || f != "industrial_age" {
		t.Errorf("record %q, run furthest %q; want modern_age, industrial_age", r, f)
	}
	if k := loaded.GetState().Mastery.K; k != config.MasteryK(5) {
		t.Errorf("the Industrial Age runs at %v, want %v", k, config.MasteryK(5))
	}
	lines := masteryLines(loaded)
	want := "Era Mastery: ages you have completed now run faster. Primitive to Atomic: mastery 5 (3.2x)."
	if len(lines) != 1 || lines[0] != want {
		t.Errorf("seeding logged %q, want once %q", lines, want)
	}

	if err := loaded.SaveGame("level5"); err != nil {
		t.Fatal(err)
	}
	if s := readSaveFromDisk(t, "level5"); !s.Prestige.MasterySeeded || s.Prestige.Mastery["atomic_age"] != 5 || s.Prestige.Furthest != "modern_age" {
		t.Errorf("the next save carries seeded %v, atomic %d, furthest %q", s.Prestige.MasterySeeded, s.Prestige.Mastery["atomic_age"], s.Prestige.Furthest)
	}
	again := NewGameEngine()
	if err := again.LoadGame("level5"); err != nil {
		t.Fatal(err)
	}
	if n := len(masteryLines(again)); n != 0 || again.Prestige.Mastery("atomic_age") != 5 || again.cheaterBadge {
		t.Errorf("second load: %d seeding lines, atomic mastery %d, tamper %v", n, again.Prestige.Mastery("atomic_age"), again.cheaterBadge)
	}
}

// TestMasterySeedingLevel0: a level-0 save from before Era Mastery gains no
// mastery and hears nothing; it only learns its record and run's furthest.
func TestMasterySeedingLevel0(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(22)
	ge.StepTicks(30)
	oldSave(t, ge, "level0")
	loaded := NewGameEngine()
	if err := loaded.LoadGame("level0"); err != nil {
		t.Fatal(err)
	}
	if len(loaded.Prestige.masterySave()) != 0 || len(masteryLines(loaded)) != 0 || loaded.cheaterBadge {
		t.Errorf("level 0: mastery %v, lines %v, tamper %v", loaded.Prestige.masterySave(), masteryLines(loaded), loaded.cheaterBadge)
	}
	if loaded.Prestige.Record() != loaded.age || !loaded.Prestige.masterySeeded {
		t.Errorf("level 0: record %q (age %q), seeded %v", loaded.Prestige.Record(), loaded.age, loaded.Prestige.masterySeeded)
	}
}

// TestMasterySaveRoundTrip: mastery, the record, the run's furthest age and
// the seeding marker survive a save and load, and the loaded engine runs at
// the same speed.
func TestMasterySaveRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(23)
	ge.StepTicks(10)
	ge.SetMasteryForTest(map[string]int{"primitive_age": 7, "stone_age": 2, "bronze_age": 1}, "iron_age")
	if err := ge.SaveGame("round"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("round"); err != nil {
		t.Fatal(err)
	}
	a, b := ge.GetState().Mastery, loaded.GetState().Mastery
	if !sameMastery(a.Ages, b.Ages) || a.Record != b.Record || a.RunFurthest != b.RunFurthest || a.K != b.K || !loaded.Prestige.masterySeeded {
		t.Errorf("round trip: %+v -> %+v", a, b)
	}
	if b.K != config.MasteryK(7) {
		t.Errorf("the Primitive Age runs at %v after load, want %v", b.K, config.MasteryK(7))
	}
}
