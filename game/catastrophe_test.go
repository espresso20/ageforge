package game

import (
	"math"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// --- helpers ---------------------------------------------------------------

// seqSource is a rand.Source that replays fixed Float64 values in order
// (cycling), so a test can force "bad roll, then escalate" and friends.
type seqSource struct {
	vals []float64
	i    int
}

func (s *seqSource) Int63() int64 {
	v := s.vals[s.i%len(s.vals)]
	s.i++
	return int64(v * (1 << 63))
}
func (s *seqSource) Seed(int64) {}

func riggedRNG(vals ...float64) *rand.Rand { return rand.New(&seqSource{vals: vals}) }

// badThenEscalate forces a bad epoch roll (0.99 ≥ any good chance) followed by
// an escalation roll that succeeds (0.01 < 0.30).
func badThenEscalate() *rand.Rand { return riggedRNG(0.99, 0.01) }

// pickBuildings returns n sorted non-wonder building keys with workers, each
// from a distinct WorkerDomain where possible.
func pickWorkerBuildings(t *testing.T, domains ...string) []string {
	t.Helper()
	byKey := config.BuildingByKey()
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []string
	for _, d := range domains {
		found := ""
		for _, k := range keys {
			def := byKey[k]
			if def.Category != "wonder" && def.WorkerCapacity > 0 && def.WorkerDomain == d {
				found = k
				break
			}
		}
		if found == "" {
			t.Fatalf("no worker building in domain %q", d)
		}
		out = append(out, found)
	}
	return out
}

func pickWonders(t *testing.T, n int) []string {
	t.Helper()
	var out []string
	for _, def := range config.BuildingByKey() {
		if def.Category == "wonder" {
			out = append(out, def.Key)
		}
	}
	sort.Strings(out)
	if len(out) < n {
		t.Fatalf("need %d wonders, have %d", n, len(out))
	}
	return out[:n]
}

// catEngine returns a seeded engine placed in the given age/epoch.
func catEngine(t *testing.T, age string, seed int64) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.SeedRNG(seed)
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	return ge
}

func nonWonderTotal(ge *GameEngine) int { return ge.Buildings.DestroyableCount() }

func setWorkers(ge *GameEngine, total int) { ge.Workers.domains["worker"].count = total }

// --- Endure -----------------------------------------------------------------

func TestEndure_DestroysTwentyPercentOfNonWondersOnly(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	b := pickWorkerBuildings(t, "food", "knowledge")
	ge.Buildings.counts[b[0]] = 10
	ge.Buildings.counts[b[1]] = 10
	wonders := pickWonders(t, 5)
	for _, w := range wonders {
		ge.Buildings.counts[w] = 1
	}
	ge.pendingCatastrophe = "iron_era"

	if err := ge.Endure(); err != nil {
		t.Fatalf("Endure: %v", err)
	}
	// 20 destroyable → floor(20/5) = 4 destroyed. The old code counted the 5
	// wonders too and destroyed 5.
	if got := nonWonderTotal(ge); got != 16 {
		t.Errorf("non-wonder buildings after Endure = %d, want 16", got)
	}
	for _, w := range wonders {
		if ge.Buildings.counts[w] != 1 {
			t.Errorf("wonder %s was touched by Endure", w)
		}
	}
}

func TestEndure_AtLeastOneBuildingDestroyed(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	b := pickWorkerBuildings(t, "food")
	ge.Buildings.counts[b[0]] = 3 // floor(3/5) = 0 → clamp to 1
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	if got := ge.Buildings.counts[b[0]]; got != 2 {
		t.Errorf("count = %d, want 2", got)
	}
}

func TestEndure_ReleasesWorkersOfDestroyedBuildings(t *testing.T) {
	ge := catEngine(t, "iron_age", 7)
	key := pickWorkerBuildings(t, "food")[0]
	capPer := config.BuildingByKey()[key].WorkerCapacity
	ge.Buildings.counts[key] = 20
	full := 20 * capPer
	setWorkers(ge, full)
	if !ge.Workers.Assign("worker", key, full) {
		t.Fatal("assign failed")
	}
	ge.pendingCatastrophe = "iron_era"

	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	// 4 destroyed → cap 16×c; release first, then the 25% loss scales the
	// remaining assignment: 16c - int(16c×0.25). Without the release the loss
	// would scale the stranded 20c instead (15c).
	remainingCap := 16 * capPer
	want := remainingCap - int(float64(remainingCap)*0.25)
	if got := ge.Workers.GetAssignedCount("worker", key); got != want {
		t.Errorf("assigned after Endure = %d, want %d", got, want)
	}
	if got := ge.Workers.GetAssignedCount("worker", key); got > ge.Buildings.counts[key]*capPer {
		t.Errorf("assigned %d exceeds capacity of %d buildings", got, ge.Buildings.counts[key])
	}
	wantPool := full - int(float64(full)*0.25)
	if got := ge.Workers.TotalPop(); got != wantPool {
		t.Errorf("pool after Endure = %d, want %d", got, wantPool)
	}
}

func TestEndure_WorkerLossIsProportionalAcrossDomains(t *testing.T) {
	ge := catEngine(t, "iron_age", 3)
	b := pickWorkerBuildings(t, "food", "knowledge", "military")
	assigned := map[string]int{b[0]: 40, b[1]: 20, b[2]: 8}
	total := 100
	for _, k := range b {
		// Plenty of buildings so destroying 20% never cuts below the assignment.
		ge.Buildings.counts[k] = 100
	}
	setWorkers(ge, total)
	for k, n := range assigned {
		if !ge.Workers.Assign("worker", k, n) {
			t.Fatalf("assign %s failed", k)
		}
	}
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	for k, n := range assigned {
		want := n - int(float64(n)*0.25)
		if got := ge.Workers.GetAssignedCount("worker", k); got != want {
			t.Errorf("%s (%s) assigned = %d, want %d", k, config.BuildingByKey()[k].WorkerDomain, got, want)
		}
	}
	if got := ge.Workers.TotalPop(); got != 75 {
		t.Errorf("pool = %d, want 75", got)
	}
}

func TestEndure_ResourcesDebuffMoraleAndRecord(t *testing.T) {
	ge := catEngine(t, "iron_age", 1)
	ge.Buildings.counts[pickWorkerBuildings(t, "food")[0]] = 5
	ge.Resources.LoadStorage(map[string]float64{"wood": 1000, "food": 1000})
	ge.Resources.LoadAmounts(map[string]float64{"wood": 800, "food": 200})
	ge.morale = 0.60
	if err := ge.forceCatastrophe(); err != nil {
		t.Fatalf("force in iron era: %v", err)
	}
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	if got := ge.Resources.Get("wood"); math.Abs(got-120) > 1e-9 {
		t.Errorf("wood = %v, want 120 (15%%)", got)
	}
	if got := ge.Resources.Get("food"); math.Abs(got-30) > 1e-9 {
		t.Errorf("food = %v, want 30 (15%%)", got)
	}
	if math.Abs(ge.morale-0.50) > 1e-9 {
		t.Errorf("morale = %v, want 0.50", ge.morale)
	}
	found := false
	for _, ev := range ge.Events.GetActiveForSave() {
		if ev.Key == "endure_reconstruction" {
			found = true
			// 216 ticks on the base curve, stretched for the Iron Age.
			if ev.TicksLeft != config.StretchTicks("iron_age", 216) || len(ev.Effects) != 1 || ev.Effects[0].Type != "production_all" || ev.Effects[0].Value != -0.10 {
				t.Errorf("reconstruction debuff = %+v", ev)
			}
		}
	}
	if !found {
		t.Error("Reconstruction Effort debuff missing")
	}
	if ge.pendingCatastrophe != "" || !ge.survivedEpochs["iron_era"] {
		t.Error("Endure must clear pending and mark the epoch survived")
	}
	r := ge.epochEventHistory[len(ge.epochEventHistory)-1]
	if r.EventType != "catastrophe" || r.Outcome != CatastropheEndured {
		t.Errorf("record = %+v, want endured catastrophe", r)
	}
	st := ge.GetState()
	if st.CatastrophesEndured != 1 || st.CatastrophesSuccumbed != 0 {
		t.Errorf("counts endured=%d succumbed=%d, want 1/0", st.CatastrophesEndured, st.CatastrophesSuccumbed)
	}
}

// --- Succumb ----------------------------------------------------------------

func succumbIn(t *testing.T, ge *GameEngine, age string) {
	t.Helper()
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	ge.pendingCatastrophe = ge.currentEpoch
	if err := ge.Succumb(); err != nil {
		t.Fatalf("Succumb in %s: %v", age, err)
	}
}

func TestSuccumb_RuinsResetAndLegacy(t *testing.T) {
	ge := catEngine(t, "iron_age", 5)
	setPrestigeLevel(ge, 2)
	b := pickWorkerBuildings(t, "food", "knowledge")
	ge.Buildings.counts[b[0]] = 6
	ge.Buildings.counts[b[1]] = 6
	succumbIn(t, ge, "iron_age")

	if ge.age != "primitive_age" || ge.currentEpoch != "stone_era" {
		t.Errorf("age/epoch = %s/%s, want primitive/stone", ge.age, ge.currentEpoch)
	}
	if ge.Prestige.GetLevel() != 2 {
		t.Errorf("prestige level = %d, want 2 (kept)", ge.Prestige.GetLevel())
	}
	if got := ge.Buildings.RuinTotal(); got != SuccumbRuinCount {
		t.Errorf("ruins = %d, want %d", got, SuccumbRuinCount)
	}
	if ge.Buildings.counts[b[0]]+ge.Buildings.counts[b[1]] != 0 {
		t.Error("buildings survived the reset")
	}
	if !ge.legacyBonuses["iron_era"] {
		t.Error("legacy flag not set")
	}
	if got := ge.permanentBonuses["iron_rate"]; math.Abs(got-0.20) > 1e-9 {
		t.Errorf("iron_rate legacy = %v, want 0.20", got)
	}
	if _, stored := ge.permanentBonuses["research_speed"]; stored {
		t.Error("research bonus must be derived, not stored in permanentBonuses")
	}
	if got := ge.succumbResearchFactor(); got != 0.8 {
		t.Errorf("research time factor = %v, want 0.8", got)
	}
	if ge.pendingCatastrophe != "" {
		t.Error("Succumb must clear pending")
	}
	st := ge.GetState()
	if st.CatastrophesSuccumbed != 1 || st.CatastrophesEndured != 0 {
		t.Errorf("counts endured=%d succumbed=%d, want 0/1", st.CatastrophesEndured, st.CatastrophesSuccumbed)
	}
}

// withStartingStock gives ge a starting-stock grant of food and wood, as the
// first shop's Starting Food and Starting Wood perks were (they are retired;
// the engine still pays whatever starting stock the shop holds).
func withStartingStock(ge *GameEngine, food, wood float64) {
	pm := ge.Prestige
	list := append([]config.PrestigeUpgradeDef(nil), pm.upgradeList...)
	for res, amount := range map[string]float64{"food": food, "wood": wood} {
		def := config.PrestigeUpgradeDef{Key: "zz_test_start_" + res, Name: "Test " + res, EffectKey: res,
			EffectType: "starting_resource", PerTier: amount, MaxTier: 1, Costs: []int{1}}
		list = append(list, def)
		pm.upgradeDefs[def.Key] = def
		pm.upgrades[def.Key] = 1
	}
	pm.upgradeList = list
}

// A Succumb sizes the stores before the starting stock lands, as a prestige
// does. The stock used to be added to stores still at their base size, so
// anything over 50 food and 50 wood was cut off, and the new run's first
// snapshot showed the base caps until the first tick.
func TestSuccumb_StartingStockLandsInSizedStores(t *testing.T) {
	// A run that fell in the Renaissance Age: the rebuild's Primitive Age
	// is six ages behind the record and catches up at 4x, stores included.
	fallen := func(seed int64) *GameEngine {
		ge := catEngine(t, "renaissance_age", seed)
		for _, a := range ageKeys() {
			ge.Prestige.NoteAgeEntered(a)
			if a == "renaissance_age" {
				break
			}
		}
		return ge
	}
	base := config.ResourceByKey()["food"].BaseStorage
	if base != 50 || config.ResourceByKey()["wood"].BaseStorage != 50 {
		t.Fatalf("base food and wood storage are %v and %v, want 50: adjust the test", base, config.ResourceByKey()["wood"].BaseStorage)
	}

	ge := fallen(21)
	succumbIn(t, ge, "renaissance_age")
	st := ge.GetState()
	if k := st.Mastery.K; k != config.CatchUpK {
		t.Fatalf("the rebuild starts at %vx, want %vx", k, config.CatchUpK)
	}
	for _, res := range []string{"food", "wood"} {
		if got, want := st.Resources[res].Storage, float64(base*config.CatchUpK); got != want {
			t.Errorf("%s storage %v in the first snapshot after a Succumb, want %v (sized before the first tick)", res, got, want)
		}
	}
	if food, wood := st.Resources["food"].Amount, st.Resources["wood"].Amount; food != 15 || wood != 12 {
		t.Errorf("a Succumb starts with %v food and %v wood, want 15 and 12", food, wood)
	}

	// With starting stock above the base stores: all of it lands, the same
	// amounts a prestige starts with.
	ge = fallen(22)
	withStartingStock(ge, 100, 90)
	succumbIn(t, ge, "renaissance_age")
	st = ge.GetState()
	if food, wood := st.Resources["food"].Amount, st.Resources["wood"].Amount; food != 115 || wood != 102 {
		t.Errorf("a Succumb with +100 food and +90 wood starts with %v food and %v wood, want 115 and 102 (it was cut off at %v)", food, wood, base)
	}
	pr := fallen(22)
	withStartingStock(pr, 100, 90)
	if err := pr.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	ps := pr.GetState()
	for _, res := range []string{"food", "wood"} {
		if st.Resources[res].Amount != ps.Resources[res].Amount || st.Resources[res].Storage != ps.Resources[res].Storage {
			t.Errorf("%s after a Succumb: %v of %v; after a prestige: %v of %v. They must start alike", res,
				st.Resources[res].Amount, st.Resources[res].Storage, ps.Resources[res].Amount, ps.Resources[res].Storage)
		}
	}
}

// twoEpochs is Ancient Knowledge after two epochs: x0.8 twice.
var twoEpochs = AncientKnowledgeFactor(2)

func TestSuccumb_ResearchBonusStacksAcrossDistinctEpochs(t *testing.T) {
	ge := catEngine(t, "iron_age", 9)
	succumbIn(t, ge, "iron_age")
	succumbIn(t, ge, "renaissance_age") // steel era
	if got := ge.succumbResearchFactor(); got != twoEpochs || math.Abs(got-0.64) > 1e-12 {
		t.Fatalf("after iron + steel: x%v, want x0.64", got)
	}
	succumbIn(t, ge, "classical_age") // iron era again: no new legacy
	if got := ge.succumbResearchFactor(); got != twoEpochs {
		t.Errorf("repeat epoch must not stack: x%v, want x0.64", got)
	}
	// It multiplies research time on its own: the research speed pool, which
	// is added up and taken off the listed time, holds none of it.
	if got := ge.combinedResearchSpeed(); got != 0 {
		t.Errorf("research speed pool = %v, want 0: Ancient Knowledge is not part of it", got)
	}
	for _, m := range ge.GetState().Modifiers {
		if m.Target == "research_speed" {
			t.Errorf("research_speed modifier %+v: Ancient Knowledge must not be in the pool", m)
		}
	}
	if got := ge.GetState().SuccumbResearchFactor; got != twoEpochs {
		t.Errorf("GameState.SuccumbResearchFactor = %v, want %v", got, twoEpochs)
	}
}

// Ancient Knowledge multiplies research time, x0.8 for each epoch succumbed
// in, so it never floors research. As +25% research speed per epoch, taken
// off the listed time, four epochs brought every tech to one tick.
func TestAncientKnowledgeNeverFloorsResearch(t *testing.T) {
	epochs := 0
	ge := NewGameEngine()
	ge.SeedRNG(3)
	for _, ep := range config.Epochs() {
		if config.CatastropheAllowed(ep.Key) {
			ge.legacyBonuses[ep.Key] = true
			epochs++
		}
	}
	if epochs != 6 {
		t.Fatalf("%d epochs a catastrophe can strike in, want 6", epochs)
	}
	factor := ge.succumbResearchFactor()
	if want := math.Pow(SuccumbResearchTimeFactor, 6); math.Abs(factor-want) > 1e-12 || factor < 0.26 {
		t.Fatalf("six epochs: x%v, want x%v", factor, want)
	}
	// Every tech in the game, started through the engine with all six.
	for _, key := range ge.Research.order {
		def := ge.Research.defs[key]
		ge.Research.currentTech = ""
		ge.Research.researched = map[string]bool{}
		ge.age = def.Age
		for _, pre := range def.Prerequisites {
			ge.Research.researched[pre] = true
		}
		ge.Resources.UnlockResource("knowledge")
		ge.Resources.AddStorage("knowledge", def.Cost)
		ge.Resources.Add("knowledge", def.Cost)
		if err := ge.startResearchLocked(key, true); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		got := ge.Research.totalTicks
		if want := int(float64(def.ResearchTicks) * factor); got != want || got <= 1 {
			t.Errorf("%s: %d ticks with six epochs, want %d of its %d (never one tick)", key, got, want, def.ResearchTicks)
		}
		if float64(got) < 0.26*float64(def.ResearchTicks)-1 {
			t.Errorf("%s: %d ticks is under 26%% of its %d", key, got, def.ResearchTicks)
		}
	}
	// Each further epoch takes the same share off what is left.
	prev := 1_000_000
	for n := 1; n <= 6; n++ {
		ticks := ResearchTicks(1_000_000, 0, AncientKnowledgeFactor(n), 1)
		if ratio := float64(ticks) / float64(prev); math.Abs(ratio-SuccumbResearchTimeFactor) > 1e-5 {
			t.Errorf("epoch %d: research time x%v of the epoch before, want x%v", n, ratio, SuccumbResearchTimeFactor)
		}
		prev = ticks
	}
	// On top of the research speed pool and Era Mastery it still multiplies.
	if got, want := ResearchTicks(1000, 0.5, AncientKnowledgeFactor(2), 2), 160; got != want {
		t.Errorf("1000 ticks at +50%% speed, two epochs, k 2: %d, want %d", got, want)
	}
	// A manager no engine has set leaves times alone.
	if got := ResearchTicks(1000, 0, 0, 0); got != 1000 {
		t.Errorf("unset factor changed 1000 ticks to %d", got)
	}
}

func TestSuccumb_ResearchBonusSurvivesPrestigeAndSaveLoad(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "iron_age", 11)
	succumbIn(t, ge, "iron_age")
	succumbIn(t, ge, "renaissance_age")

	ge.age = "modern_age"
	ge.currentEpoch = config.EpochForAge("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	if got := ge.succumbResearchFactor(); got != twoEpochs {
		t.Fatalf("after DoPrestige: x%v, want x%v", got, twoEpochs)
	}
	if err := ge.SaveGame("cat_research"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("cat_research"); err != nil {
		t.Fatal(err)
	}
	if got := ge2.succumbResearchFactor(); got != twoEpochs {
		t.Errorf("after load: x%v, want x%v", got, twoEpochs)
	}
	if got := ge2.GetState().SuccumbResearchFactor; got != twoEpochs {
		t.Errorf("after load, GameState.SuccumbResearchFactor = %v, want %v", got, twoEpochs)
	}
}

func TestLoad_MigratesStoredSuccumbResearchBonus(t *testing.T) {
	ge := NewGameEngine()
	ge.legacyBonuses = map[string]bool{"iron_era": true, "steel_era": true}
	// An old save: the last Succumb wrote +0.25 into permanentBonuses.
	ge.permanentBonuses["research_speed"] = 0.25
	ge.restoreCatastropheState(&GameSave{})
	if _, ok := ge.permanentBonuses["research_speed"]; ok {
		t.Errorf("stored research bonus not stripped: %v", ge.permanentBonuses["research_speed"])
	}
	if got := ge.combinedResearchSpeed(); got != 0 {
		t.Errorf("research speed pool = %v, want 0: the stored copy is gone", got)
	}
	if got := ge.succumbResearchFactor(); got != twoEpochs {
		t.Errorf("Ancient Knowledge = x%v, want x%v (derived from the legacy flags)", got, twoEpochs)
	}

	// A new-format save is trusted as-is.
	ge2 := NewGameEngine()
	ge2.permanentBonuses["research_speed"] = 0.15
	ge2.restoreCatastropheState(&GameSave{SuccumbResearchDerived: true})
	if got := ge2.permanentBonuses["research_speed"]; got != 0.15 {
		t.Errorf("new-format save research_speed = %v, want 0.15", got)
	}
}

func TestLoad_ReconstructsOutcomesOnOldRecords(t *testing.T) {
	ge := NewGameEngine()
	ge.pendingCatastrophe = "steel_era"
	ge.survivedEpochs = map[string]bool{"iron_era": true}
	ge.legacyBonuses = map[string]bool{"electric_era": true}
	ge.epochEventHistory = []EpochEventRecord{
		{EpochKey: "electric_era", EventType: "catastrophe"},
		{EpochKey: "iron_era", EventType: "catastrophe"},
		{EpochKey: "steel_era", EventType: "catastrophe"},
		{EpochKey: "digital_era", EventType: "catastrophe"},
		{EpochKey: "iron_era", EventType: "good_minor"},
	}
	ge.restoreCatastropheState(&GameSave{})
	want := []string{CatastropheSuccumbed, CatastropheEndured, CatastrophePending, "", ""}
	for i, r := range ge.epochEventHistory {
		if r.Outcome != want[i] {
			t.Errorf("record %d (%s %s) outcome = %q, want %q", i, r.EpochKey, r.EventType, r.Outcome, want[i])
		}
	}
}

// --- Pending blocks progress --------------------------------------------------

func TestPendingBlocksAdvanceUntilResolved(t *testing.T) {
	ge := catEngine(t, "iron_age", 2)
	ge.Buildings.counts[pickWorkerBuildings(t, "food")[0]] = 5
	ge.pendingCatastrophe = "iron_era"
	ge.ageReady = true
	err := ge.AdvanceAge()
	if err == nil || !strings.Contains(err.Error(), "catastrophe") {
		t.Fatalf("AdvanceAge while pending: err = %v, want a catastrophe block", err)
	}
	if ge.age != "iron_age" {
		t.Fatalf("age moved while blocked: %s", ge.age)
	}
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	ge.ageReady = true
	if err := ge.AdvanceAge(); err != nil {
		t.Fatalf("AdvanceAge after Endure: %v", err)
	}
	if ge.age != "classical_age" {
		t.Errorf("age = %s, want classical_age", ge.age)
	}
}

func TestPendingBlocksPrestigeUntilResolved(t *testing.T) {
	ge := catEngine(t, "modern_age", 2)
	ge.pendingCatastrophe = "digital_era"
	err := ge.DoPrestige()
	if err == nil || !strings.Contains(err.Error(), "catastrophe") {
		t.Fatalf("DoPrestige while pending: err = %v, want a catastrophe block", err)
	}
	if ge.Prestige.GetLevel() != 0 {
		t.Fatal("prestige happened while blocked")
	}
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige after Endure: %v", err)
	}
}

// --- Iron-epoch gate, overwrite, once per epoch --------------------------------

func TestNoCatastropheBeforeIronEpoch(t *testing.T) {
	if config.CatastropheAllowed("stone_era") || !config.CatastropheAllowed("iron_era") || !config.CatastropheAllowed("cosmic_era") {
		t.Fatal("gate: stone must be closed, iron and later open")
	}
	if config.FateAllowed("stone_era") || !config.FateAllowed("iron_era") || !config.FateAllowed("neon_era") || !config.FateAllowed("cosmic_era") {
		t.Fatal("fates: none in the Stone Era, every era from the Iron Era to the Cosmic Era")
	}
	ge := catEngine(t, "bronze_age", 1)
	if err := ge.forceCatastrophe(); err == nil {
		t.Error("forcing a catastrophe in the Stone Era must be refused")
	}
	// Nothing is ever fated in the Stone Era, whatever the roll.
	g := catEngine(t, "primitive_age", 5)
	for i := 0; i < 300; i++ {
		g.rollFate()
		if g.fate == nil || g.fate.EpochKey != "stone_era" || g.fate.Fated {
			t.Fatalf("roll %d: stone era fate = %+v", i, g.fate)
		}
	}
	// A transition never brings a catastrophe: a bad roll entering the Iron
	// Era is a challenging event, and the era's doom is only fated.
	var toast []EventData
	ge.Bus.Subscribe(EventEpochEventFired, func(e EventData) { toast = append(toast, e) })
	ge.rng = badThenEscalate()
	ge.age = "iron_age"
	ge.currentEpoch = "iron_era"
	ge.rollEpochEvent("iron_era", "bronze_age")
	if ge.pendingCatastrophe != "" {
		t.Fatalf("iron era transition roll: pending = %q", ge.pendingCatastrophe)
	}
	last := ge.epochEventHistory[len(ge.epochEventHistory)-1]
	if last.EventType != "bad_challenging" || len(toast) != 1 || toast[0].Payload["event_type"] != "bad_challenging" {
		t.Errorf("iron era bad roll = %+v, events %+v; want one challenging event", last, toast)
	}
}

func TestRollNeverOverwritesPending(t *testing.T) {
	ge := catEngine(t, "renaissance_age", 1)
	ge.pendingCatastrophe = "iron_era"
	ge.rng = badThenEscalate()
	ge.rollEpochEvent("steel_era", ge.age)
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("pending overwritten: %q", ge.pendingCatastrophe)
	}
	for _, r := range ge.epochEventHistory {
		if r.EpochKey == "steel_era" && r.EventType == "catastrophe" {
			t.Error("steel era got a catastrophe record while iron was pending")
		}
	}
	if err := ge.forceCatastrophe(); err == nil {
		t.Error("forcing while pending must be refused")
	}
}

// The dev console's /catastrophe is the only direct trigger, and it respects
// the Iron gate. Without dev mode it does nothing.
func TestDevCatastropheCommandRespectsGate(t *testing.T) {
	prev := DevModeActive
	t.Cleanup(func() { DevModeActive = prev })

	DevModeActive = false
	ge := catEngine(t, "iron_age", 1)
	if out := DevConsoleCommand("/catastrophe", ge); out != "" || ge.pendingCatastrophe != "" {
		t.Fatalf("dev command ran without dev mode: %q, pending=%q", out, ge.pendingCatastrophe)
	}

	DevModeActive = true
	stone := catEngine(t, "bronze_age", 1)
	if out := DevConsoleCommand("/catastrophe", stone); !strings.Contains(out, "refused") || stone.pendingCatastrophe != "" {
		t.Errorf("stone era: %q, pending=%q; want refused", out, stone.pendingCatastrophe)
	}
	if out := DevConsoleCommand("/catastrophe", ge); ge.pendingCatastrophe != "iron_era" {
		t.Errorf("iron era: %q, pending=%q; want iron_era", out, ge.pendingCatastrophe)
	}
	// Other dev commands still reach DevExecCommand.
	if out := DevConsoleCommand("/ages", ge); !strings.Contains(out, "iron_age") {
		t.Errorf("/ages via DevConsoleCommand = %q", out)
	}
}

// --- Harbinger seam: invite ---------------------------------------------------

// The invite flag only arms the Last Passage now: an era's invite lives on its
// fate (fate_test.go). Outside the final epoch it moves nothing, and Succumb
// and prestige clear it.
func TestInviteFlagArmsOnlyTheLastPassage(t *testing.T) {
	ge := catEngine(t, "renaissance_age", 1)
	ge.harbingerTickCheck()
	before := ge.catastropheOutlook()
	ge.inviteCatastrophe()
	if o := ge.catastropheOutlook(); o != before || o.Probability != 0 {
		t.Errorf("invite flag in a quiet era moved the outlook: %+v, was %+v", o, before)
	}
	lp := catEngine(t, "quantum_age", 1)
	lp.inviteCatastrophe()
	if o := lp.catastropheOutlook(); !o.Possible || o.Probability != 1 || o.Tier != CatastropheTierHigh {
		t.Errorf("invited Last Passage outlook = %+v, want certain/high", o)
	}
	// Succumb and prestige start a new run: the invite does not carry over.
	succumbIn(t, ge, "iron_age")
	if ge.catastropheInvited {
		t.Error("invite survived Succumb")
	}
}

// --- Ruin cap -------------------------------------------------------------------

func TestRuinCap_DropsLowestValueFirst(t *testing.T) {
	bm := NewBuildingManager()
	byKey := config.BuildingByKey()
	var early, late string
	// Pick a primitive-age producer and a much later one.
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		def := byKey[k]
		if def.Category == "wonder" {
			continue
		}
		if early == "" && def.RequiredAge == "primitive_age" {
			early = k
		}
		if late == "" && def.RequiredAge == "industrial_age" {
			late = k
		}
	}
	if early == "" || late == "" {
		t.Fatal("could not pick ruin keys")
	}
	bm.LoadRuins(map[string]int{early: 20, late: 10})
	if got := bm.RuinTotal(); got != MaxRuins {
		t.Fatalf("ruin total = %d, want %d", got, MaxRuins)
	}
	if bm.ruins[late] != 10 || bm.ruins[early] != MaxRuins-10 {
		t.Errorf("ruins = %v, want the late ones kept and early trimmed", bm.ruins)
	}
}

func TestRuinCap_RepeatedSuccumbsNeverExceedCap(t *testing.T) {
	ge := catEngine(t, "iron_age", 6)
	key := pickWorkerBuildings(t, "food")[0]
	for i := 0; i < 5; i++ {
		ge.Buildings.counts[key] = 12
		succumbIn(t, ge, "iron_age")
		if got := ge.Buildings.RuinTotal(); got > MaxRuins {
			t.Fatalf("after %d succumbs ruins = %d > cap %d", i+1, got, MaxRuins)
		}
	}
	if got := ge.Buildings.RuinTotal(); got != MaxRuins {
		t.Errorf("ruins after 5 succumbs = %d, want %d", got, MaxRuins)
	}
}

// --- Determinism ------------------------------------------------------------------

func endureOutcome(t *testing.T, seed int64) map[string]int {
	t.Helper()
	ge := catEngine(t, "iron_age", seed)
	byKey := config.BuildingByKey()
	keys := make([]string, 0)
	for k, def := range byKey {
		if def.Category != "wonder" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys[:12] {
		ge.Buildings.counts[k] = 15
	}
	ge.pendingCatastrophe = "iron_era"
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	out := make(map[string]int)
	for _, k := range keys[:12] {
		out[k] = ge.Buildings.counts[k]
	}
	return out
}

func epochRollOutcome(t *testing.T, seed int64) []string {
	t.Helper()
	ge := catEngine(t, "stone_age", seed)
	var got []string
	for _, ep := range config.Epochs()[1:] {
		ge.pendingCatastrophe = "" // resolve instantly so every roll runs
		ge.currentEpoch = ep.Key
		ge.rollEpochEvent(ep.Key, ge.age)
		r := ge.epochEventHistory[len(ge.epochEventHistory)-1]
		got = append(got, r.EventKey)
	}
	return got
}

func TestCatastropheRandomnessIsSeeded(t *testing.T) {
	a, b := endureOutcome(t, 42), endureOutcome(t, 42)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("same seed, different destroyed set:\n%v\n%v", a, b)
	}
	differs := false
	for s := int64(43); s < 48; s++ {
		if !reflect.DeepEqual(a, endureOutcome(t, s)) {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("five other seeds all destroyed the same set as seed 42")
	}

	r1, r2 := epochRollOutcome(t, 42), epochRollOutcome(t, 42)
	if !reflect.DeepEqual(r1, r2) {
		t.Errorf("same seed, different epoch rolls:\n%v\n%v", r1, r2)
	}
	differs = false
	for s := int64(43); s < 53; s++ {
		if !reflect.DeepEqual(r1, epochRollOutcome(t, s)) {
			differs = true
			break
		}
	}
	if !differs {
		t.Error("ten other seeds all rolled the same epoch events as seed 42")
	}

	// Ruins too.
	ruins := func(seed int64) map[string]int {
		ge := catEngine(t, "iron_age", seed)
		for i, k := range pickWorkerBuildings(t, "food", "knowledge", "military") {
			ge.Buildings.counts[k] = 5 + i
		}
		ge.Buildings.GenerateRuins(ge.gameRNG(), 8)
		return ge.Buildings.GetAllRuins()
	}
	if !reflect.DeepEqual(ruins(99), ruins(99)) {
		t.Error("same seed, different ruins")
	}
}

// --- Outlook ------------------------------------------------------------------------

// setFaithStanding gives ge the faith that reads as share of full standing
// in the age it is in (FaithStandingFullIn), with a store that holds it.
func setFaithStanding(ge *GameEngine, share float64) {
	amount := share * FaithStandingFullIn(ge.rules, ge.age)
	ge.Resources.LoadStorage(map[string]float64{"faith": math.Max(amount, ge.Resources.GetStorage("faith"))})
	ge.Resources.LoadAmounts(map[string]float64{"faith": amount})
}

// The outlook is what the player can know. In a quiet era it reads the same
// whether a doom is fated or not; once a harbinger warns, it says what the
// warning says (the strike chance by faith band); the final epoch keeps the
// Last Passage's odds.
func TestCatastropheOutlook(t *testing.T) {
	cases := []struct {
		name     string
		standing float64
		wantP    float64
		wantTier CatastropheTier
		wantBand FaithBand
	}{
		{"no faith", 0, 0.90, CatastropheTierHigh, FaithBandLow},
		{"low faith", 0.10, 0.90, CatastropheTierHigh, FaithBandLow},
		{"mid faith", 0.50, 0.75, CatastropheTierMedium, FaithBandMid},
		{"high faith", 0.90, 0.60, CatastropheTierLow, FaithBandHigh},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			quiet := catEngine(t, "classical_age", 1)
			fated := catEngine(t, "classical_age", 1)
			if err := quiet.ForceQuietFateForTest("iron_era"); err != nil {
				t.Fatal(err)
			}
			if err := fated.ForceFateForTest("iron_era", 20000); err != nil {
				t.Fatal(err)
			}
			for _, ge := range []*GameEngine{quiet, fated} {
				setFaithStanding(ge, c.standing)
			}
			oq, of := quiet.CatastropheOutlook(), fated.CatastropheOutlook()
			if oq != of {
				t.Fatalf("the outlook tells a fated era from a quiet one:\nquiet %+v\nfated %+v", oq, of)
			}
			if !oq.Possible || oq.Warned || oq.Probability != 0 || oq.Tier != CatastropheTierNone || oq.NextEpochKey != "steel_era" ||
				math.Abs(oq.FaithStanding-c.standing) > 1e-9 || oq.FaithBand != c.wantBand ||
				oq.FaithFull != FaithStandingFullIn(quiet.rules, "classical_age") {
				t.Errorf("quiet outlook = %+v, want possible, unwarned, no odds, standing %v (%s)", oq, c.standing, c.wantBand)
			}
			// A harbinger comes: the outlook says what its warning says.
			if err := fated.SummonHarbingerForTest("classical_age"); err != nil {
				t.Fatal(err)
			}
			setFaithStanding(fated, c.standing)
			o := fated.CatastropheOutlook()
			if !o.Possible || !o.Warned || math.Abs(o.Probability-c.wantP) > 1e-9 || o.Tier != c.wantTier {
				t.Errorf("warned outlook = %+v, want p=%v tier=%s", o, c.wantP, c.wantTier)
			}
			if st := fated.GetState(); st.CatastropheOutlook != o {
				t.Errorf("GetState outlook %+v != %+v", st.CatastropheOutlook, o)
			}
		})
	}

	// Nothing can strike in the Stone Era.
	if o := catEngine(t, "bronze_age", 1).CatastropheOutlook(); o.Possible || o.Warned {
		t.Errorf("stone era outlook = %+v, want not possible", o)
	}
	// The final epoch's passage is prestige: the Last Passage, same odds.
	ge := catEngine(t, "quantum_age", 1)
	if o := ge.CatastropheOutlook(); !o.Possible || o.Passage != PassagePrestige || o.NextEpochKey != "" || o.Probability != 0.18 {
		t.Errorf("final epoch outlook = %+v, want the Last Passage at 18%%", o)
	}
	ge.pendingLastPassage = true
	if o := ge.CatastropheOutlook(); o.Possible || o.Passage != PassagePrestige || o.Probability != 0 {
		t.Errorf("final epoch outlook with the Last Passage pending = %+v, want not possible", o)
	}
	// An era whose doom has been lifted has nothing more to bring.
	spared := catEngine(t, "classical_age", 2)
	if err := spared.SummonHarbingerForTest("classical_age"); err != nil {
		t.Fatal(err)
	}
	spared.rng = riggedRNG(0.999)
	spared.fate.StrikeTick = spared.tick
	spared.harbingerTickCheck()
	if spared.fate.Resolved != FateSpared {
		t.Fatalf("setup: the doom was not spared: %+v", spared.fate)
	}
	if o := spared.CatastropheOutlook(); o.Possible || o.Warned {
		t.Errorf("outlook after the era's doom was spared = %+v, want not possible", o)
	}
}

// --- Save/load with pending ----------------------------------------------------------

func TestSaveLoadKeepsPendingAndItStillBlocks(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "iron_age", 8)
	if err := ge.forceCatastrophe(); err != nil {
		t.Fatal(err)
	}
	if err := ge.SaveGame("cat_pending"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	loaded := 0
	ge2.Bus.Subscribe(EventGameLoaded, func(EventData) { loaded++ })
	if err := ge2.LoadGame("cat_pending"); err != nil {
		t.Fatal(err)
	}
	if loaded != 1 {
		t.Errorf("EventGameLoaded published %d times, want 1", loaded)
	}
	if ge2.GetState().PendingCatastrophe != "iron_era" {
		t.Fatalf("pending lost across save/load: %q", ge2.pendingCatastrophe)
	}
	ge2.mu.Lock()
	ge2.ageReady = true
	ge2.mu.Unlock()
	if err := ge2.AdvanceAge(); err == nil {
		t.Error("loaded pending catastrophe must still block advancing")
	}
	r := ge2.epochEventHistory[len(ge2.epochEventHistory)-1]
	if r.Outcome != CatastrophePending {
		t.Errorf("record outcome after load = %q, want pending", r.Outcome)
	}
	if err := ge2.Endure(); err != nil {
		t.Fatal(err)
	}
}

// Saves written by early builds of this change carry catastrophe_fired. The
// field is ignored on load but must not break the signature check.
func TestLoadIgnoresDeprecatedCatastropheFired(t *testing.T) {
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	ge := catEngine(t, "iron_age", 3)
	snap := ge.buildSaveSnapshot()
	snap.CatastropheFired = map[string]bool{"iron_era": true}
	snap.Signature = signSave(snap, saveHMACKey)
	if err := os.MkdirAll(saveDirectory(), 0755); err != nil {
		t.Fatal(err)
	}
	writeRawSave(t, "cat_deprecated", snap)

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("cat_deprecated"); err != nil {
		t.Fatal(err)
	}
	if ge2.cheaterBadge {
		t.Error("save with catastrophe_fired failed signature verification")
	}
	if ge2.buildSaveSnapshot().CatastropheFired != nil {
		t.Error("catastrophe_fired must not be written back")
	}
}

func TestCountCatastropheOutcomes(t *testing.T) {
	h := []string{
		"Tick 5 — Endured The Great Plague (Iron Era). 3 buildings lost.",
		"Tick 9 — Succumbed to The World War (Steel Era). Civilization reset. Legacy bonus earned.",
		"Tick 2 — Succumbed to The World War (Steel Era). Civilization reset. Legacy bonus already held.",
	}
	e, s := countCatastropheOutcomes(h)
	if e != 1 || s != 2 {
		t.Errorf("endured=%d succumbed=%d, want 1/2", e, s)
	}
}
