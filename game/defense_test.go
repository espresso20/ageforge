package game

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// giveSoldiers unlocks the soldiers resource and fills it with n soldiers
// (storage raised to fit).
func giveSoldiers(ge *GameEngine, n float64) {
	ge.Resources.UnlockResource("soldiers")
	ge.Resources.AddStorage("soldiers", n)
	ge.Resources.Add("soldiers", n)
}

// soldiersFor returns how many soldiers blunt exactly half the cap (defense ==
// threat) in age, with no military_power bonus.
func soldiersFor(age string) float64 {
	order := 0
	for i, a := range config.AgeOrder() {
		if a == age {
			order = i
		}
	}
	return config.AgeThreat(order) / 2
}

func eventDef(t *testing.T, key string) config.EventDef {
	t.Helper()
	for _, e := range append(config.RandomEvents(), config.EpochExclusiveEvents()...) {
		if e.Key == key {
			return e
		}
	}
	t.Fatalf("no event %q", key)
	return config.EventDef{}
}

// --- The curve --------------------------------------------------------------------

func TestDefenseMitigation_Curve(t *testing.T) {
	threat := config.AgeThreat(5)
	if got := config.DefenseMitigation(0, threat); got != 0 {
		t.Errorf("no defense blunts %v, want 0", got)
	}
	if got := config.DefenseMitigation(-10, threat); got != 0 {
		t.Errorf("negative defense blunts %v, want 0", got)
	}
	if got := config.DefenseMitigation(threat, threat); math.Abs(got-config.DefenseMitigationCap/2) > 1e-12 {
		t.Errorf("defense == threat blunts %v, want half the cap %v", got, config.DefenseMitigationCap/2)
	}
	prev := 0.0
	for d := 1.0; d < 1e15; d *= 1.7 {
		m := config.DefenseMitigation(d, threat)
		if m > config.DefenseMitigationCap {
			t.Fatalf("defense %g blunts %v, above the cap %v", d, m, config.DefenseMitigationCap)
		}
		if m < prev {
			t.Fatalf("not monotonic: defense %g blunts %v after %v", d, m, prev)
		}
		prev = m
	}
	if prev < config.DefenseMitigationCap*0.99 {
		t.Errorf("a vast army blunts %v, want it close to the cap", prev)
	}
}

func TestDefenseMitigation_ScaledByAgeThreat(t *testing.T) {
	ages := config.AgeOrder()
	for i := 1; i < len(ages); i++ {
		if config.AgeThreat(i) <= config.AgeThreat(i-1) {
			t.Fatalf("threat does not grow from %s to %s", ages[i-1], ages[i])
		}
	}
	// A garrison that holds its own in the Iron Age is a rounding error by the
	// Quantum Age: the same defense against a later age's threat blunts less.
	iron := 0
	quantum := 0
	for i, a := range ages {
		switch a {
		case "iron_age":
			iron = i
		case "quantum_age":
			quantum = i
		}
	}
	d := config.AgeThreat(iron)
	atIron := config.DefenseMitigation(d, config.AgeThreat(iron))
	atQuantum := config.DefenseMitigation(d, config.AgeThreat(quantum))
	if atIron < 0.2 {
		t.Errorf("an Iron Age garrison blunts %v of an Iron Age raid, want at least 0.2", atIron)
	}
	if atQuantum > 0.001 {
		t.Errorf("an Iron Age garrison blunts %v of a Quantum Age raid, want under 0.1%%", atQuantum)
	}
}

func TestRaidMitigation_NoSoldiersIsZero(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	if m := ge.raidMitigation(); m != 0 {
		t.Fatalf("no soldiers blunts %v, want 0", m)
	}
	giveSoldiers(ge, soldiersFor("iron_age"))
	if m := ge.raidMitigation(); math.Abs(m-config.DefenseMitigationCap/2) > 1e-12 {
		t.Fatalf("defense == threat blunts %v, want %v", m, config.DefenseMitigationCap/2)
	}
}

// --- Raid events ------------------------------------------------------------------

func TestRaidEvent_StealBluntedByGarrison(t *testing.T) {
	def := eventDef(t, "bandit_raid")
	if !def.Raid {
		t.Fatal("bandit_raid is not flagged Raid")
	}
	run := func(soldiers float64) (food, gold float64, ge *GameEngine) {
		ge = catEngine(t, "iron_age", 1)
		ge.Resources.AddStorage("food", 10000)
		ge.Resources.AddStorage("gold", 10000)
		ge.Resources.UnlockResource("gold")
		ge.Resources.Add("food", 1000)
		ge.Resources.Add("gold", 1000)
		if soldiers > 0 {
			giveSoldiers(ge, soldiers)
		}
		f, g := ge.Resources.Get("food"), ge.Resources.Get("gold")
		ge.applyEventEffects(def)
		return f - ge.Resources.Get("food"), g - ge.Resources.Get("gold"), ge
	}
	food0, gold0, ge0 := run(0)
	if ge0.Stats.Defense != nil {
		t.Errorf("no garrison, yet a defense tally: %+v", ge0.Stats.Defense)
	}
	food1, gold1, ge1 := run(soldiersFor("iron_age"))
	m := config.DefenseMitigationCap / 2
	for _, c := range []struct {
		res       string
		base, got float64
	}{{"food", food0, food1}, {"gold", gold0, gold1}} {
		if c.base <= 0 {
			t.Fatalf("undefended raid took no %s", c.res)
		}
		if want := c.base * (1 - m); math.Abs(c.got-want) > 1e-9 {
			t.Errorf("%s lost %v with the garrison, want %v (%v blunted of %v)", c.res, c.got, want, m, c.base)
		}
	}
	tally := ge1.Stats.Defense
	if tally == nil || tally.Raids != 1 || math.Abs(tally.Resources["food"]-food0*m) > 1e-9 {
		t.Errorf("defense tally = %+v, want 1 raid and %v food saved", tally, food0*m)
	}
	if !logHas(ge1, "Your garrison blunted about 22% of the raid") {
		t.Error("no log line saying what the garrison kept")
	}
}

func TestRaidEvent_WorkerLossBlunted(t *testing.T) {
	def := eventDef(t, "tribal_raid")
	run := func(soldiers float64) int {
		ge := catEngine(t, "iron_age", 1)
		setWorkers(ge, 100)
		if soldiers > 0 {
			giveSoldiers(ge, soldiers)
		}
		ge.applyEventEffects(def)
		return 100 - ge.Workers.TotalPop()
	}
	base, guarded := run(0), run(soldiersFor("iron_age"))
	if base != 10 {
		t.Fatalf("undefended tribal raid drove off %d of 100 workers, want 10", base)
	}
	if guarded >= base || guarded < 7 {
		t.Errorf("guarded tribal raid drove off %d workers, want fewer than %d and about %d", guarded, base, 8)
	}
}

func TestNonRaidEventIgnoresGarrison(t *testing.T) {
	def := eventDef(t, "earthquake")
	if def.Raid {
		t.Fatal("an earthquake is not a raid")
	}
	run := func(soldiers float64) float64 {
		ge := catEngine(t, "iron_age", 1)
		ge.Resources.AddStorage("wood", 10000)
		ge.Resources.Add("wood", 1000)
		if soldiers > 0 {
			giveSoldiers(ge, soldiers)
		}
		ge.applyEventEffects(def)
		return ge.Resources.Get("wood")
	}
	if a, b := run(0), run(1e6); a != b {
		t.Errorf("earthquake left %v wood without soldiers and %v with, want equal", a, b)
	}
}

// --- War raids --------------------------------------------------------------------

func TestWarRaid_BluntedByGarrison(t *testing.T) {
	faction := config.BaseFactions()[0].Key
	run := func(soldiers float64) (float64, *GameEngine) {
		ge := catEngine(t, "iron_age", 4)
		ge.Resources.UnlockResource("gold")
		ge.Resources.AddStorage("gold", 10000)
		ge.Resources.Add("gold", 1000)
		if soldiers > 0 {
			giveSoldiers(ge, soldiers)
		}
		ge.Diplomacy.pendingRaids = []RaidRequest{{FactionKey: faction, Resource: "gold", Amount: 200, Message: "raided"}}
		ge.applyWarRaids()
		return 1000 - ge.Resources.Get("gold"), ge
	}
	base, _ := run(0)
	if base != 200 {
		t.Fatalf("undefended war raid took %v gold, want 200", base)
	}
	got, ge := run(soldiersFor("iron_age"))
	if want := 200 * (1 - config.DefenseMitigationCap/2); math.Abs(got-want) > 1e-9 {
		t.Errorf("guarded war raid took %v gold, want %v", got, want)
	}
	if !logHas(ge, "Your garrison kept 45 gold") {
		t.Error("war raid log does not say what the garrison kept")
	}
}

// A raid bigger than the stock takes nothing; a garrison must not turn it into
// one that lands.
func TestWarRaid_GarrisonNeverMakesAMissLand(t *testing.T) {
	faction := config.BaseFactions()[0].Key
	ge := catEngine(t, "iron_age", 4)
	ge.Resources.UnlockResource("gold")
	ge.Resources.AddStorage("gold", 10000)
	ge.Resources.Add("gold", 150)
	giveSoldiers(ge, soldiersFor("iron_age"))
	ge.Diplomacy.pendingRaids = []RaidRequest{{FactionKey: faction, Resource: "gold", Amount: 200, Message: "raided"}}
	ge.applyWarRaids()
	if got := ge.Resources.Get("gold"); got != 150 {
		t.Errorf("gold after a raid bigger than the stock = %v, want 150 (untouched, as without an army)", got)
	}
}

// --- Endure -----------------------------------------------------------------------

func TestComputeEndure_NoGarrisonIsBraceOnly(t *testing.T) {
	for brace := 0; brace <= HarbingerMaxBrace; brace++ {
		o := computeEndure(137, brace, 0)
		if o.DestroyPct != float64(braceDestroyPct[brace]) || o.KeepFrac != braceKeepFrac[brace] {
			t.Errorf("brace %d without garrison: %v%% fall, %v kept; want %d%%, %v", brace, o.DestroyPct, o.KeepFrac, braceDestroyPct[brace], braceKeepFrac[brace])
		}
		if want := 137 * braceDestroyPct[brace] / 100; o.DestroyCount != want || o.BuildingsSaved != 0 || o.Capped {
			t.Errorf("brace %d without garrison: %+v, want %d destroyed and nothing saved", brace, o, want)
		}
	}
}

func TestComputeEndure_BraceThenGarrison(t *testing.T) {
	// Brace 1 (15% fall, 30% kept), then a garrison blunting 20% of the rest.
	o := computeEndure(100, 1, 0.20)
	if math.Abs(o.DestroyPct-12) > 1e-9 {
		t.Errorf("destroy = %v%%, want 12%% (15%% x 0.8)", o.DestroyPct)
	}
	if want := 0.30 + 0.70*0.20; math.Abs(o.KeepFrac-want) > 1e-9 {
		t.Errorf("keep = %v, want %v", o.KeepFrac, want)
	}
	if o.DestroyCount != 12 || o.BuildingsSaved != 3 || o.Capped {
		t.Errorf("counts = %+v, want 12 destroyed, 3 saved, uncapped", o)
	}
}

func TestComputeEndure_CombinedCap(t *testing.T) {
	floorDestroy := float64(braceDestroyPct[0]) * (1 - config.EndureReductionCap)
	floorLoss := (1 - braceKeepFrac[0]) * (1 - config.EndureReductionCap)
	for brace := 0; brace <= HarbingerMaxBrace; brace++ {
		for _, g := range []float64{0.05, 0.2, 0.3, config.DefenseMitigationCap} {
			o := computeEndure(1000, brace, g)
			if o.DestroyPct < floorDestroy-1e-9 {
				t.Errorf("brace %d garrison %v: %v%% fall, below the combined floor %v%%", brace, g, o.DestroyPct, floorDestroy)
			}
			if loss := 1 - o.KeepFrac; loss < floorLoss-1e-9 {
				t.Errorf("brace %d garrison %v: %v of stock lost, below the combined floor %v", brace, g, loss, floorLoss)
			}
			if o.DestroyPct > float64(braceDestroyPct[brace]) || o.KeepFrac < braceKeepFrac[brace] {
				t.Errorf("brace %d garrison %v: the garrison made it worse: %+v", brace, g, o)
			}
			if saved := float64(o.BuildingsSaved) / float64(braceDestroyPct[brace]*10); saved > g+1e-9 {
				t.Errorf("brace %d garrison %v: saved %v of the braced buildings, more than the garrison's share", brace, g, saved)
			}
		}
	}
	// Brace 2 (10% fall, 45% kept) with a maxed garrison hits the cap on both.
	o := computeEndure(100, 2, config.DefenseMitigationCap)
	if !o.Capped || math.Abs(o.DestroyPct-8) > 1e-9 || o.DestroyCount != 8 || o.BuildingsSaved != 2 {
		t.Errorf("brace 2 + max garrison = %+v, want capped at 8%% (8 of 100 fall, 2 saved)", o)
	}
	if want := 1 - 0.85*0.4; math.Abs(o.KeepFrac-want) > 1e-9 {
		t.Errorf("brace 2 + max garrison keeps %v, want the cap's %v", o.KeepFrac, want)
	}
}

// Endure end to end: Brace and garrison combined, the log says what the army
// saved, and a garrison-free Endure is exactly the braced one.
func TestEndure_BraceAndGarrison(t *testing.T) {
	setup := func(soldiers float64) *GameEngine {
		ge := catEngine(t, "iron_age", 11)
		b := pickWorkerBuildings(t, "food", "knowledge")
		ge.Buildings.counts[b[0]] = 50
		ge.Buildings.counts[b[1]] = 50
		ge.Resources.UnlockResource("gold")
		ge.Resources.AddStorage("gold", 100000)
		ge.Resources.Add("gold", 10000)
		if soldiers > 0 {
			giveSoldiers(ge, soldiers)
		}
		ge.pendingBraceLevel = 1
		ge.pendingCatastrophe = "iron_era"
		return ge
	}
	plain := setup(0)
	if err := plain.Endure(); err != nil {
		t.Fatal(err)
	}
	if got := nonWonderTotal(plain); got != 85 {
		t.Errorf("braced Endure left %d of 100 buildings, want 85", got)
	}
	if got := plain.Resources.Get("gold"); math.Abs(got-3000) > 1e-6 {
		t.Errorf("braced Endure left %v gold, want 3000 (30%%)", got)
	}

	guarded := setup(soldiersFor("iron_age"))
	want := guarded.GetState().PendingEndure
	if want.Garrison <= 0 || want.BuildingsSaved == 0 {
		t.Fatalf("preview shows no garrison: %+v", want)
	}
	if err := guarded.Endure(); err != nil {
		t.Fatal(err)
	}
	if got := nonWonderTotal(guarded); got != 100-want.DestroyCount {
		t.Errorf("guarded Endure left %d buildings, preview promised %d", got, 100-want.DestroyCount)
	}
	if got, w := guarded.Resources.Get("gold"), 10000*want.KeepFrac; math.Abs(got-w) > 1e-6 || got <= 3000 {
		t.Errorf("guarded Endure left %v gold, want %v (more than the braced 3000)", got, w)
	}
	if !logHas(guarded, "Your garrison held the line") {
		t.Error("Endure log does not say what the garrison saved")
	}
	if tally := guarded.Stats.Defense; tally == nil || tally.Buildings != want.BuildingsSaved || tally.Resources["gold"] <= 0 {
		t.Errorf("tally after Endure = %+v, want %d buildings and some gold", tally, want.BuildingsSaved)
	}
}

// Seeded rolls: the same seed and garrison destroy the same buildings.
func TestEndure_GarrisonDeterministic(t *testing.T) {
	run := func() map[string]int {
		ge := catEngine(t, "iron_age", 42)
		keys := pickWorkerBuildings(t, "food", "knowledge")
		for _, k := range keys {
			ge.Buildings.counts[k] = 20
		}
		giveSoldiers(ge, soldiersFor("iron_age"))
		ge.pendingCatastrophe = "iron_era"
		if err := ge.Endure(); err != nil {
			t.Fatal(err)
		}
		out := map[string]int{}
		for _, k := range keys {
			out[k] = ge.Buildings.counts[k]
		}
		return out
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Errorf("same seed, different Endure: %v vs %v", a, b)
	}
}

// --- Previews ---------------------------------------------------------------------

func TestHarbingerViewCountsGarrison(t *testing.T) {
	ge := catEngine(t, "bronze_age", 5)
	if err := ge.SummonHarbingerForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	v := ge.GetState().Harbinger
	if v == nil || v.GarrisonPct != 0 || v.EndureDestroyPct != 20 || v.EndureKeepPct != 15 {
		t.Fatalf("no garrison: view = %+v, want 20%% / 15%% and no garrison", v)
	}
	// The catastrophe would strike in the Iron Age (the target epoch's first
	// age): a garrison that matches that threat blunts half the cap.
	giveSoldiers(ge, soldiersFor("iron_age"))
	v = ge.GetState().Harbinger
	if v.GarrisonPct != 23 || v.EndureDestroyPct >= 20 || v.EndureKeepPct <= 15 {
		t.Errorf("garrison: view = %d%% garrison, %d%% fall, %d%% kept; want 23%%, under 20%%, over 15%%", v.GarrisonPct, v.EndureDestroyPct, v.EndureKeepPct)
	}
}

// --- Save/load --------------------------------------------------------------------

func TestDefenseTally_SaveLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	// Without a garrison the save carries no defense field at all.
	plain := catEngine(t, "iron_age", 2)
	b, err := json.Marshal(plain.buildSaveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"defense"`) {
		t.Error("a save with no garrison writes a defense field")
	}

	ge := catEngine(t, "iron_age", 2)
	giveSoldiers(ge, soldiersFor("iron_age"))
	ge.Resources.AddStorage("food", 10000)
	ge.Resources.Add("food", 1000)
	ge.applyEventEffects(eventDef(t, "bandit_raid"))
	want := ge.Stats.Defense.clone()
	if want == nil || want.Raids != 1 {
		t.Fatalf("tally before save = %+v", want)
	}
	if err := ge.SaveGame("defense"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("defense"); err != nil {
		t.Fatal(err)
	}
	if ge2.cheaterBadge {
		t.Error("save with a defense tally failed signature verification")
	}
	if !reflect.DeepEqual(ge2.Stats.Defense, want) {
		t.Errorf("tally after load = %+v, want %+v", ge2.Stats.Defense, want)
	}
	if saved := ge2.GetState().Military.Saved; !reflect.DeepEqual(saved, want) {
		t.Errorf("Army panel tally after load = %+v, want %+v", saved, want)
	}
}
