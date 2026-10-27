package game

import (
	"math"
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
		err := ge.startResearchLocked(tech.Key, true)
		ge.mu.Unlock()
		if err != nil {
			t.Fatalf("research %s: %v", tech.Key, err)
		}
	}
	if b, k := base.researchPerTick, known.researchPerTick; b <= 0 || math.Abs(k-b*2) > 1e-9*b {
		t.Errorf("the scholars do %v work a tick at k = 2 against %v at k = 1, want twice", k, b)
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
// untouched. Reset clears everything. (A Succumb commits too:
// TestSuccumbCommitsMastery.)
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

	// On the engine: a prestige from the Modern Age commits.
	ge := newSeededEngine(3)
	if err := ge.SummonHarbingerForTest("modern_age"); err != nil {
		t.Log(err) // places the game; a missing harbinger is fine
	}
	noErr(t, ge.DoPrestige())
	if m := ge.Prestige.Mastery("atomic_age"); m != 1 {
		t.Fatalf("the Atomic Age has mastery %d after a Modern prestige, want 1", m)
	}
	if k := ge.GetState().Mastery.K; k != 4 {
		t.Errorf("the new run's Primitive Age runs at %v; with the record 12 ages ahead it catches up at 4", k)
	}
	ge.Reset()
	if len(ge.Prestige.masterySave()) != 0 || ge.Prestige.Record() != "primitive_age" || ge.GetState().Mastery.K != 1 {
		t.Errorf("Reset kept mastery %v, record %q", ge.Prestige.masterySave(), ge.Prestige.Record())
	}
}

// masteryThrough is the mastery map of a player whose ages before upTo hold
// level m each.
func masteryThrough(upTo string, m int) map[string]int {
	out := map[string]int{}
	for _, a := range ageKeys() {
		if a == upTo {
			break
		}
		out[a] = m
	}
	return out
}

// TestSuccumbCommitsMastery: a Succumb ends the run as a prestige does, so
// every age the run completed gains a level through the same commit, the
// rebuild starts on known ground, and the prestige after it counts only what
// the rebuilt run completed.
func TestSuccumbCommitsMastery(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := newSeededEngine(5)
	// A first run that reaches the Renaissance Age and falls there.
	for _, a := range ageKeys() {
		ge.Prestige.NoteAgeEntered(a)
		if a == "renaissance_age" {
			break
		}
	}
	ge.age, ge.currentEpoch = "renaissance_age", config.EpochForAge("renaissance_age")
	if g := ge.GetState().Mastery.NextGains; len(g) != 6 || g[0] != "primitive_age" || g[5] != "medieval_age" {
		t.Fatalf("before the fall the run would raise %v, want Primitive to Medieval", g)
	}
	ge.pendingCatastrophe = ge.currentEpoch
	noErr(t, ge.Succumb())
	if got, want := ge.Prestige.masterySave(), masteryThrough("renaissance_age", 1); !sameMastery(got, want) {
		t.Fatalf("after the Succumb mastery is %v, want %v: the six ages the run completed, one level each", got, want)
	}
	if ge.Prestige.Mastery("renaissance_age") != 0 {
		t.Error("the age the run fell in was not completed and must gain nothing")
	}
	if ge.Prestige.RunFurthest() != "primitive_age" || ge.Prestige.Record() != "renaissance_age" {
		t.Errorf("after the Succumb the run's furthest is %q and the record %q, want primitive_age and renaissance_age", ge.Prestige.RunFurthest(), ge.Prestige.Record())
	}
	if ge.Prestige.GetLevel() != 0 || ge.Prestige.Snapshot().TotalEarned != 0 {
		t.Errorf("a Succumb is not a prestige: level %d, points %d", ge.Prestige.GetLevel(), ge.Prestige.Snapshot().TotalEarned)
	}
	// The rebuild starts on known ground: the Primitive Age is six ages
	// behind the record, so it catches up at 4x.
	st := ge.GetState()
	if st.Mastery.K != config.CatchUpK || len(st.Mastery.NextGains) != 0 {
		t.Errorf("the rebuild starts at %vx with next gains %v, want %vx and none", st.Mastery.K, st.Mastery.NextGains, config.CatchUpK)
	}
	if indexOfLog(ge, "Era Mastery: the Primitive Age to the Medieval Age gained a mastery level each") < 0 ||
		indexOfLog(ge, "Known ground: the Primitive Age runs 4x faster") < 0 {
		t.Errorf("the log does not say what the fall earned:\n%s", strings.Join(logMessages(ge), "\n"))
	}

	// The rebuild reaches the Medieval Age and takes the early prestige. It
	// raises the five ages that run completed, and nothing above them: the
	// Medieval Age, completed once, before the fall, stays at 1.
	for _, a := range ageKeys() {
		ge.Prestige.NoteAgeEntered(a)
		if a == "medieval_age" {
			break
		}
	}
	ge.age, ge.currentEpoch = "medieval_age", config.EpochForAge("medieval_age")
	ge.Prestige.floor = "" // Keep this focused on per-run mastery accounting.
	noErr(t, ge.DoPrestige())
	want := masteryThrough("medieval_age", 2)
	want["medieval_age"] = 1
	if got := ge.Prestige.masterySave(); !sameMastery(got, want) {
		t.Fatalf("after the Succumb and a Medieval prestige mastery is %v, want %v: an age is counted once per run that completed it", got, want)
	}

	// It is saved like any mastery.
	ge2 := saveAndLoad(t, ge, "fallen")
	if got := ge2.Prestige.masterySave(); !sameMastery(got, want) || ge2.cheaterBadge {
		t.Errorf("after a save and load mastery is %v (tampered: %v), want %v", got, ge2.cheaterBadge, want)
	}

	// A Succumb never passes the cap.
	top := newSeededEngine(6)
	for _, a := range ageKeys() {
		top.Prestige.SetMastery(a, config.MasteryCap)
	}
	top.Prestige.NoteAgeEntered("iron_age")
	top.age, top.currentEpoch = "iron_age", config.EpochForAge("iron_age")
	top.pendingCatastrophe = top.currentEpoch
	noErr(t, top.Succumb())
	for _, a := range ageKeys() {
		if m := top.Prestige.Mastery(a); m != config.MasteryCap {
			t.Errorf("%s has mastery %d after a Succumb at the cap, want %d", a, m, config.MasteryCap)
		}
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
	noErr(t, ge.DoPrestige())
	st := ge.GetState()
	if st.TickSpeedBonus != 0 {
		t.Errorf("after a prestige: tick speed %v; the passive retired into Era Mastery", st.TickSpeedBonus)
	}
	if st.Prestige.ProductionMultiplier <= 1 {
		t.Errorf("prestige production multiplier = %v, want above 1", st.Prestige.ProductionMultiplier)
	}
}

// TestMasteryDropKeepsStock: when k drops, stock has no ceiling to be cut to:
// nothing is lost, and a credit still adds to it whole.
func TestMasteryDropKeepsStock(t *testing.T) {
	isolateAccountDir(t)
	_, ge := masteryTwins(t, 1)
	ge.mu.Lock()
	ge.recalculateRates()
	r := ge.Resources.resources["wood"]
	high := config.Million // a stock held at k = 2
	r.Amount = high
	ge.Prestige.SetMastery(ge.age, 0) // k drops to 1 (as entering new ground does)
	ge.recalculateRates()
	ge.mu.Unlock()
	if r.Amount != high {
		t.Fatalf("after k dropped: wood %v, was %v; want the stock kept", r.Amount, high)
	}
	ge.mu.Lock()
	ge.Resources.Add("wood", 3*config.Million)
	ge.mu.Unlock()
	if r := ge.Resources.resources["wood"]; r.Amount != high+3*config.Million {
		t.Errorf("wood %v after a big add, want %v: nothing clips a credit", r.Amount, high+3*config.Million)
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

// TestMasterySaveRoundTrip: mastery, the record and the run's furthest age
// survive a save and load, and the loaded engine runs at
// the same speed.
func TestMasterySaveRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(23)
	ge.StepTicks(10)
	ge.SetMasteryForTest(map[string]int{"primitive_age": 7, "stone_age": 2, "bronze_age": 1}, "iron_age")
	loaded := saveAndLoad(t, ge, "round")
	a, b := ge.GetState().Mastery, loaded.GetState().Mastery
	if !sameMastery(a.Ages, b.Ages) || a.Record != b.Record || a.RunFurthest != b.RunFurthest || a.K != b.K {
		t.Errorf("round trip: %+v -> %+v", a, b)
	}
	if b.K != config.MasteryK(7) {
		t.Errorf("the Primitive Age runs at %v after load, want %v", b.K, config.MasteryK(7))
	}
}
