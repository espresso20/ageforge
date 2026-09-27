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
			if ev.TicksLeft != 216 || len(ev.Effects) != 1 || ev.Effects[0].Type != "production_all" || ev.Effects[0].Value != -0.10 {
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
	if got := ge.succumbResearchBonus(); got != 0.25 {
		t.Errorf("research bonus = %v, want 0.25", got)
	}
	if ge.pendingCatastrophe != "" {
		t.Error("Succumb must clear pending")
	}
	st := ge.GetState()
	if st.CatastrophesSuccumbed != 1 || st.CatastrophesEndured != 0 {
		t.Errorf("counts endured=%d succumbed=%d, want 0/1", st.CatastrophesEndured, st.CatastrophesSuccumbed)
	}
}

func legacyModValue(mods []Modifier) float64 {
	v := 0.0
	for _, m := range mods {
		if m.Source == "legacy" && m.Target == "research_speed" {
			v += m.Value
		}
	}
	return v
}

func TestSuccumb_ResearchBonusStacksAcrossDistinctEpochs(t *testing.T) {
	ge := catEngine(t, "iron_age", 9)
	succumbIn(t, ge, "iron_age")
	succumbIn(t, ge, "renaissance_age") // steel era
	if got := ge.succumbResearchBonus(); got != 0.50 {
		t.Fatalf("after iron + steel: %v, want 0.50", got)
	}
	succumbIn(t, ge, "classical_age") // iron era again: no new legacy
	if got := ge.succumbResearchBonus(); got != 0.50 {
		t.Errorf("repeat epoch must not stack: %v, want 0.50", got)
	}
	if got := ge.combinedResearchSpeed(); math.Abs(got-0.50) > 1e-9 {
		t.Errorf("combinedResearchSpeed = %v, want 0.50", got)
	}
	// Flows through the resolver → Active Multipliers.
	if got := legacyModValue(ge.GetState().Modifiers); got != 0.50 {
		t.Errorf("resolver legacy research_speed = %v, want 0.50", got)
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
	if got := ge.succumbResearchBonus(); got != 0.50 {
		t.Fatalf("after DoPrestige: %v, want 0.50", got)
	}
	if err := ge.SaveGame("cat_research"); err != nil {
		t.Fatal(err)
	}
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("cat_research"); err != nil {
		t.Fatal(err)
	}
	if got := ge2.succumbResearchBonus(); got != 0.50 {
		t.Errorf("after load: %v, want 0.50", got)
	}
	if got := legacyModValue(ge2.GetState().Modifiers); got != 0.50 {
		t.Errorf("after load, resolver legacy = %v, want 0.50", got)
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
	if got := ge.combinedResearchSpeed(); got != 0.50 {
		t.Errorf("combined research speed = %v, want 0.50 (derived)", got)
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
	ge := catEngine(t, "bronze_age", 1)
	if err := ge.forceCatastrophe(); err == nil {
		t.Error("forcing a catastrophe in the Stone Era must be refused")
	}
	// A forced bad+escalate roll on a Stone Era transition stays a challenging event.
	ge.rng = badThenEscalate()
	ge.rollEpochEvent("stone_era")
	if ge.pendingCatastrophe != "" {
		t.Errorf("stone era roll produced a catastrophe: %q", ge.pendingCatastrophe)
	}
	last := ge.epochEventHistory[len(ge.epochEventHistory)-1]
	if last.EventType != "bad_challenging" {
		t.Errorf("stone era bad roll = %q, want bad_challenging", last.EventType)
	}

	// The same forced roll entering the Iron Era is a catastrophe, with a toast event.
	var toast []EventData
	ge.Bus.Subscribe(EventEpochEventFired, func(e EventData) { toast = append(toast, e) })
	ge.rng = badThenEscalate()
	ge.age = "iron_age"
	ge.currentEpoch = "iron_era"
	ge.rollEpochEvent("iron_era")
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("iron era forced roll: pending = %q", ge.pendingCatastrophe)
	}
	if len(toast) != 1 || toast[0].Payload["event_type"] != "catastrophe" {
		t.Errorf("catastrophe bus event = %+v", toast)
	}
}

func TestRollNeverOverwritesPending(t *testing.T) {
	ge := catEngine(t, "renaissance_age", 1)
	ge.pendingCatastrophe = "iron_era"
	ge.rng = badThenEscalate()
	ge.rollEpochEvent("steel_era")
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

func TestInvitedCatastropheStrikesAtNextAllowedTransition(t *testing.T) {
	ge := catEngine(t, "bronze_age", 1)
	ge.inviteCatastrophe()
	// Rigged to roll a good event: the invite must override it.
	ge.rng = riggedRNG(0.0)
	ge.currentEpoch = "iron_era"
	ge.rollEpochEvent("iron_era")
	if ge.pendingCatastrophe != "iron_era" {
		t.Fatalf("invited catastrophe did not strike: pending=%q", ge.pendingCatastrophe)
	}
	if ge.catastropheInvited {
		t.Error("invite not consumed")
	}
	last := ge.epochEventHistory[len(ge.epochEventHistory)-1]
	if last.EventType != "catastrophe" || !strings.Contains(last.EventName, "invited") {
		t.Errorf("record = %+v", last)
	}
}

func TestInviteWaitsWhileGatedOrPending(t *testing.T) {
	ge := catEngine(t, "stone_age", 1)
	ge.inviteCatastrophe()
	ge.rng = riggedRNG(0.0)
	ge.rollEpochEvent("stone_era") // gated
	if ge.pendingCatastrophe != "" || !ge.catastropheInvited {
		t.Fatalf("gated transition: pending=%q invited=%v; want none and kept", ge.pendingCatastrophe, ge.catastropheInvited)
	}
	ge.pendingCatastrophe = "iron_era"
	ge.rollEpochEvent("steel_era") // something already pending
	if ge.pendingCatastrophe != "iron_era" || !ge.catastropheInvited {
		t.Fatalf("pending transition: pending=%q invited=%v; want iron_era and kept", ge.pendingCatastrophe, ge.catastropheInvited)
	}
	// The outlook reports a certain catastrophe while invited.
	ge.pendingCatastrophe = ""
	ge.currentEpoch = "steel_era" // next: electric, not rolled yet
	if o := ge.catastropheOutlook(); !o.Possible || o.Probability != 1 || o.Tier != CatastropheTierHigh {
		t.Errorf("invited outlook = %+v, want certain/high", o)
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
		ge.rollEpochEvent(ep.Key)
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

func setFaith(ge *GameEngine, amount, storage float64) {
	ge.Resources.LoadStorage(map[string]float64{"faith": storage})
	ge.Resources.LoadAmounts(map[string]float64{"faith": amount})
}

func TestCatastropheOutlook(t *testing.T) {
	cases := []struct {
		name          string
		faith, store  float64
		wantP         float64
		wantTier      CatastropheTier
		wantFaithFill float64
	}{
		{"no faith storage", 0, 0, 0.15, CatastropheTierMedium, 0},
		{"low faith", 10, 100, 0.18, CatastropheTierHigh, 0.10},
		{"mid faith", 50, 100, 0.15, CatastropheTierMedium, 0.50},
		{"high faith", 90, 100, 0.12, CatastropheTierLow, 0.90},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ge := catEngine(t, "bronze_age", 1) // stone era; next transition → iron
			setFaith(ge, c.faith, c.store)
			o := ge.CatastropheOutlook()
			if o.NextEpochKey != "iron_era" || !o.Possible {
				t.Fatalf("outlook = %+v, want possible, next iron_era", o)
			}
			if math.Abs(o.Probability-c.wantP) > 1e-9 || o.Tier != c.wantTier || math.Abs(o.FaithFill-c.wantFaithFill) > 1e-9 {
				t.Errorf("outlook = %+v, want p=%v tier=%s fill=%v", o, c.wantP, c.wantTier, c.wantFaithFill)
			}
			if st := ge.GetState(); st.CatastropheOutlook != o {
				t.Errorf("GetState outlook %+v != %+v", st.CatastropheOutlook, o)
			}
		})
	}

	ge := catEngine(t, "quantum_age", 1)
	if o := ge.CatastropheOutlook(); o.Possible || o.NextEpochKey != "" || o.Tier != CatastropheTierNone || o.Probability != 0 {
		t.Errorf("final epoch outlook = %+v, want none", o)
	}
	ge = catEngine(t, "medieval_age", 1)
	ge.epochEventFired["steel_era"] = true
	if o := ge.CatastropheOutlook(); o.Possible || o.NextEpochKey != "steel_era" {
		t.Errorf("already-rolled next epoch outlook = %+v, want not possible", o)
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
