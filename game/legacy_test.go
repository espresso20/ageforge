package game

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// loadFixture copies a committed save into the test's save directory and
// loads it.
func loadFixture(t *testing.T, path, name string) *GameEngine {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath(name), data, 0o644); err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	if err := ge.LoadGame(name); err != nil {
		t.Fatal(err)
	}
	return ge
}

// TestPrestigeFromMedieval: prestige opens at the Medieval Age and pays its
// depth points, 9; the Classical Age is refused.
func TestPrestigeFromMedieval(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(7)
	_ = ge.SummonHarbingerForTest("classical_age")
	if err := ge.DoPrestige(); err == nil {
		t.Fatal("prestige from the Classical Age was accepted")
	}
	_ = ge.SummonHarbingerForTest("medieval_age")
	st := ge.GetState().Prestige
	if !st.CanPrestige || st.PendingPoints != 9 || st.NextAge != "renaissance_age" || st.NextAgePoints != 12 {
		t.Errorf("in the Medieval Age: can %v, pending %d, next %q pays %d; want true, 9, renaissance_age, 12",
			st.CanPrestige, st.PendingPoints, st.NextAge, st.NextAgePoints)
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if p := ge.GetState().Prestige; p.Level != 1 || p.Available != 9 || p.TotalEarned != 9 {
		t.Errorf("after a Medieval prestige: level %d, available %d, total %d; want 1, 9, 9", p.Level, p.Available, p.TotalEarned)
	}
}

// TestPlanLogRecordsThePlanAsWritten: the plan log tags each item with the
// age it was written in, merges a building's copies, takes back the copies
// a removed item never started, skips deals, and becomes the template at
// prestige.
func TestPlanLogRecordsThePlanAsWritten(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(9)
	ge.mu.Lock()
	ge.Resources.Add("wood", 1000)
	ge.Resources.Add("food", 1000)
	ge.mu.Unlock()
	if _, err := ge.PlanAddBuild("hut", 3); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 2); err != nil { // not merged into the last item: research sits between
		t.Fatal(err)
	}
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	// Remove the second hut item before it starts anything: its 2 copies go.
	if _, err := ge.PlanRemove(3); err != nil {
		t.Fatal(err)
	}
	// A started copy stays written: start one hut, then clear the plan.
	ge.mu.Lock()
	ge.plan[0].Count--
	ge.plan[0].Started++
	ge.mu.Unlock()
	ge.PlanClear()

	ge.mu.Lock()
	got := clonePlanTemplate(ge.planLog)
	ge.mu.Unlock()
	want := []PlanTemplateItem{{Age: "primitive_age", Kind: PlanBuild, Key: "hut", Count: 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plan log after the edits: %+v, want %+v", got, want)
	}

	// Now write a plan that stays, and prestige: the log becomes the
	// template, ages included, and starts over.
	if _, err := ge.PlanAddBuild("hut", 4); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	_ = ge.SummonHarbingerForTest("medieval_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	kit := ge.LegacyForTest()
	want = []PlanTemplateItem{
		{Age: "primitive_age", Kind: PlanBuild, Key: "hut", Count: 5},
		{Age: "primitive_age", Kind: PlanAdvance},
	}
	if !reflect.DeepEqual(kit.Plan, want) {
		t.Errorf("template after prestige: %+v, want %+v", kit.Plan, want)
	}
	ge.mu.Lock()
	n := len(ge.planLog)
	ge.mu.Unlock()
	if n != 2 {
		t.Errorf("the new run's plan log holds %d items, want the remembered first-age slice", n)
	}
}

// TestPlanTemplateChainsAcrossAdvance: with Plan Template bought, a new run
// starts with the first age's slice in the plan, and each advance adds the
// next age's slice, the advance item included, so the plan chains ages.
func TestPlanTemplateChainsAcrossAdvance(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(11)
	ge.SetLegacyForTest(LegacyKit{Plan: []PlanTemplateItem{
		{Age: "primitive_age", Kind: PlanBuild, Key: "hut", Count: 2},
		{Age: "primitive_age", Kind: PlanAdvance},
		{Age: "stone_age", Kind: PlanBuild, Key: "stone_pit", Count: 3},
		{Age: "stone_age", Kind: PlanAdvance},
	}}, true)
	ge.mu.Lock()
	ge.plan = nil // as if this run's copy was built: the log keeps it as written
	ge.mu.Unlock()
	_ = ge.SummonHarbingerForTest("medieval_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	kinds := func() []string {
		var out []string
		for _, it := range ge.GetState().Plan {
			out = append(out, it.Kind+":"+it.Key)
		}
		return out
	}
	if got, want := kinds(), []string{"build:hut", "advance:"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("the new run's plan: %v, want %v", got, want)
	}
	ge.mu.Lock()
	ge.plan = nil // the Primitive Age's items are spent
	ge.advanceAge("stone_age")
	ge.mu.Unlock()
	if got, want := kinds(), []string{"build:stone_pit", "advance:"}; !reflect.DeepEqual(got, want) {
		t.Errorf("after the advance: %v, want %v", got, want)
	}
	// The re-applied slices are written again, so the template carries on.
	ge.mu.Lock()
	logged := templateAgeCount(ge.planLog, "primitive_age") + templateAgeCount(ge.planLog, "stone_age")
	ge.mu.Unlock()
	if logged != 4 {
		t.Errorf("the run's plan log holds %d template items, want 4", logged)
	}
}

// TestPlanTemplateWithinPlanCap: a slice bigger than the plan's room fills
// it to MaxPlanItems and says how many waited out. The slice's tech goes in
// on top: research takes no room in the plan.
func TestPlanTemplateWithinPlanCap(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(12)
	var plan []PlanTemplateItem
	for range MaxPlanItems {
		plan = append(plan, PlanTemplateItem{Age: "primitive_age", Kind: PlanBuild, Key: "hut", Count: 1},
			PlanTemplateItem{Age: "primitive_age", Kind: PlanResearch, Key: "tool_making"})
	}
	ge.mu.Lock()
	ge.plan = nil
	for i := range MaxPlanItems - 1 {
		ge.plan = append(ge.plan, PlanItem{Kind: PlanDeal, Key: "riverlands_tribes", Deal: i + 1, Count: 1})
	}
	ge.mu.Unlock()
	ge.SetLegacyForTest(LegacyKit{Plan: plan}, true)
	if st := ge.GetState(); PlanLoad(st.Plan) != MaxPlanItems || len(st.Plan) != MaxPlanItems+1 {
		t.Errorf("the plan holds %d items, %d of them counted; want the cap of %d and one tech on top", len(st.Plan), PlanLoad(st.Plan), MaxPlanItems)
	}
	found := false
	for _, l := range ge.GetLogs() {
		if strings.HasPrefix(l.Message, "Plan Template: added ") && strings.Contains(l.Message, "The plan is full (60 items, not counting research)") {
			found = true
		}
	}
	if !found {
		t.Error("no Plan Template line saying the plan was full")
	}
}

// TestPlanTemplateCarriesPlannedResearch: the techs a player planned are
// their own research path, so they come along with the template: one that
// started (and left the plan) and one still waiting are both planned again
// in the next run, in the age they were planned in. Nothing else is, and
// nothing starts researching by itself.
func TestPlanTemplateCarriesPlannedResearch(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(13)
	for _, key := range []string{"tool_making", "fire_mastery"} {
		if err := ge.PlanAddResearch(key); err != nil {
			t.Fatal(err)
		}
	}
	ge.mu.Lock()
	var s planStarts
	ge.runPlan(&s) // Tool Making starts and leaves the plan; Fire Mastery waits behind it
	started, left := ge.Research.currentTech, len(ge.plan)
	ge.mu.Unlock()
	if started != "tool_making" || left != 1 {
		t.Fatalf("after the plan ran: researching %q with %d items left; want tool_making, 1", started, left)
	}
	_ = ge.SummonHarbingerForTest("medieval_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	var planned []string
	for _, it := range ge.GetState().Plan {
		if it.Kind == PlanResearch {
			planned = append(planned, it.Key)
		}
	}
	if want := []string{"tool_making", "fire_mastery"}; !reflect.DeepEqual(planned, want) {
		t.Errorf("the next run's plan holds the techs %v, want %v (the player's own plan)", planned, want)
	}
	// With no knowledge nothing starts, and the kit starts nothing of its
	// own when there is: only the plan's first tech takes the slot.
	st := ge.GetState()
	if st.Research.CurrentTech != "" || st.Research.TotalResearched != 0 {
		t.Errorf("the new run researches %q with %d done before the player did anything", st.Research.CurrentTech, st.Research.TotalResearched)
	}
	ge.mu.Lock()
	ge.plan = nil // the player drops the plan: the research slot stays idle
	ge.mu.Unlock()
	ge.StepTicks(5)
	if cur := ge.GetState().Research.CurrentTech; cur != "" {
		t.Errorf("with an empty plan and knowledge to spare, %q started by itself", cur)
	}
}

// TestOldFriendsMetAtTheirAge: with Old Friends bought, a remembered
// civilization is met as soon as the age reaches its own, at neutral
// opinion, with no expedition; without it, nothing happens.
func TestOldFriendsMetAtTheirAge(t *testing.T) {
	isolateAccountDir(t)
	for _, own := range []bool{true, false} {
		ge := newSeededEngine(15)
		ge.SetLegacyForTest(LegacyKit{Factions: []string{"riverlands_tribes", "ironhold_clans"}}, false)
		ge.mu.Lock()
		if own {
			ge.Prestige.level = 1
		}
		ge.advanceAge("stone_age")
		early := ge.Diplomacy.IsDiscovered("riverlands_tribes")
		ge.advanceAge("bronze_age")
		met := ge.Diplomacy.IsDiscovered("riverlands_tribes")
		later := ge.Diplomacy.IsDiscovered("ironhold_clans")
		fs, _ := ge.Diplomacy.StateOf("riverlands_tribes")
		ge.mu.Unlock()
		if early || later {
			t.Errorf("own %v: met before their age (riverlands at Stone %v, ironhold at Bronze %v)", own, early, later)
		}
		if met != own {
			t.Errorf("own %v: Riverlands Tribes met at the Bronze Age = %v", own, met)
		}
		if own && (fs.Opinion != 0 || fs.Status != "neutral") {
			t.Errorf("an old friend met at opinion %d, status %q; want 0, neutral", fs.Opinion, fs.Status)
		}
	}
}

// TestWorkerSharesCarryOver: the first prestige makes remembered shares free.
func TestWorkerSharesCarryOver(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(16)
	if _, err := ge.SetWorkerShare("faith", 40); err != nil {
		t.Fatal(err)
	}
	_ = ge.SummonHarbingerForTest("medieval_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if got := ge.WorkerShares(); got["faith"] != 40 {
		t.Errorf("shares after first prestige %v, want faith 40", got)
	}
}

// TestLegacyKitIsFreeAfterFirstPrestige applies remembered automation at once.
func TestLegacyKitIsFreeAfterFirstPrestige(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(17)
	if _, err := ge.PlanAddBuild("hut", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.SetWorkerShare("food", 50); err != nil {
		t.Fatal(err)
	}
	_ = ge.SummonHarbingerForTest("modern_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	if p := ge.GetState().Plan; len(p) != 1 || p[0].Key != "hut" || p[0].Count != 2 {
		t.Errorf("plan after first prestige: %+v, want 2 huts", p)
	}
	if got := ge.WorkerShares(); got["food"] != 50 {
		t.Errorf("shares after buying Worker Shares: %v, want food 50", got)
	}
}

// TestSuccumbReappliesTemplate: a Succumb folds the run's plan into the
// template and starts the rebuild with the first age's slice.
func TestSuccumbReappliesTemplate(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(18)
	ge.mu.Lock()
	ge.Prestige.level = 1
	ge.mu.Unlock()
	if _, err := ge.PlanAddBuild("hut", 3); err != nil {
		t.Fatal(err)
	}
	_ = ge.SummonHarbingerForTest("iron_age")
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatal(err)
	}
	if err := ge.Succumb(); err != nil {
		t.Fatal(err)
	}
	if p := ge.GetState().Plan; len(p) != 1 || p[0].Key != "hut" || p[0].Count != 3 {
		t.Errorf("plan after Succumb: %+v, want 3 huts", p)
	}
}

// TestLegacySaveRoundTrip: the shop version, the kit's memory (a planned
// tech in the template included) and the run's plan log survive a save and
// load unchanged.
func TestLegacySaveRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(19)
	kit := LegacyKit{
		Plan:     []PlanTemplateItem{{Age: "stone_age", Kind: PlanBuild, Key: "stone_pit", Count: 4}, {Age: "stone_age", Kind: PlanResearch, Key: "pottery"}, {Age: "stone_age", Kind: PlanAdvance}},
		Factions: []string{"riverlands_tribes"},
		Shares:   map[string]float64{"faith": 30},
	}
	ge.SetLegacyForTest(kit, true)
	if _, err := ge.PlanAddBuild("hut", 2); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	log := clonePlanTemplate(ge.planLog)
	ge.mu.Unlock()
	if err := ge.SaveGame("kit"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("kit"); err != nil {
		t.Fatal(err)
	}
	if loaded.cheaterBadge {
		t.Fatal("the save failed its own signature check")
	}
	if got := loaded.LegacyForTest(); !reflect.DeepEqual(got, kit) {
		t.Errorf("kit after load: %+v, want %+v", got, kit)
	}
	loaded.mu.RLock()
	gotLog := clonePlanTemplate(loaded.planLog)
	owned := loaded.Prestige.Owns(config.LegacyFactions)
	loaded.mu.RUnlock()
	if !reflect.DeepEqual(gotLog, log) || len(log) == 0 || !owned {
		t.Errorf("after load: plan log %+v (want %+v), owns Old Friends %v", gotLog, log, owned)
	}
}

// TestNewSaveWritesNoEmptyKitFields: a new game writes its shop version and
// nothing else of the kit, so saves without kit memory stay small.
func TestNewSaveWritesNoEmptyKitFields(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(20)
	if err := ge.SaveGame("fresh"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("fresh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"legacy_plan"`, `"legacy_factions"`, `"legacy_shares"`, `"plan_log"`} {
		if strings.Contains(string(data), field) {
			t.Errorf("a new game's save writes %s with nothing in it", field)
		}
	}
	// Research Memory was cut before release: no save ever carries its
	// fields (a remembered research order, a run's completion order).
	for _, field := range []string{`"legacy_research"`, `"order"`} {
		if strings.Contains(string(data), field) {
			t.Errorf("a save writes %s, a Research Memory field", field)
		}
	}
	if strings.Contains(string(data), `"shop_version"`) {
		t.Error("a new game's save writes the retired shop version")
	}
}

// writeShopSave writes ge's state as a signed save that mutate has changed:
// a stand-in for a build from the days of the prestige shop, which wrote
// tiers and a shop version that no build writes now.
func writeShopSave(t *testing.T, ge *GameEngine, name string, mutate func(*GameSave)) {
	t.Helper()
	ge.mu.RLock()
	gs := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	mutate(&gs)
	gs.Signature = signSave(gs)
	data, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath(name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestLegacyShopSavesKeepTheirCarryOverAndPoints verifies old signed save
// fields still load, while level rather than a former purchase turns on the
// three free carry-overs.
func TestLegacyShopSavesKeepTheirCarryOverAndPoints(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(31)
	ge.mu.Lock()
	ge.Prestige.LoadState(1, 120, 120)
	ge.Prestige.LoadLegacy([]PlanTemplateItem{{Age: "primitive_age", Kind: PlanBuild, Key: "hut", Count: 1}}, []string{"riverlands_tribes"}, map[string]float64{"food": 40})
	ge.mu.Unlock()
	writeShopSave(t, ge, "old-shop", func(gs *GameSave) {
		gs.Prestige.Upgrades = map[string]int{config.LegacyPlan: 1, "gather_boost": 3}
		gs.Prestige.ShopVersion = 2
	})
	if saved := readSaveFromDisk(t, "old-shop"); saved.Prestige.Upgrades[config.LegacyPlan] != 1 || saved.Prestige.ShopVersion != 2 {
		t.Fatalf("setup: the old save lost its shop state: %+v", saved.Prestige)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("old-shop"); err != nil {
		t.Fatal(err)
	}
	if loaded.cheaterBadge {
		t.Fatal("the old signed save failed its signature check")
	}
	p := loaded.GetState().Prestige
	if p.TotalEarned != 120 || p.Available != 120 || p.ProductionMultiplier <= 1 {
		t.Errorf("unspent old points loaded as %+v", p)
	}
	for _, key := range config.LegacyKit() {
		if !loaded.Prestige.Owns(key) {
			t.Errorf("first-prestige old save does not carry %s", key)
		}
	}
	if err := loaded.SaveGame("old-shop-rewritten"); err != nil {
		t.Fatal(err)
	}
	if saved := readSaveFromDisk(t, "old-shop-rewritten"); saved.Prestige.Upgrades != nil || saved.Prestige.ShopVersion != 0 {
		t.Errorf("rewritten old save keeps shop state: %+v", saved.Prestige)
	}

	never := newSeededEngine(32)
	writeShopSave(t, never, "never-prestiged", func(gs *GameSave) {
		gs.Prestige.Upgrades = map[string]int{config.LegacyPlan: 1, config.LegacyWorkers: 1, config.LegacyFactions: 1}
		gs.Prestige.ShopVersion = 2
	})
	loaded = NewGameEngine()
	if err := loaded.LoadGame("never-prestiged"); err != nil {
		t.Fatal(err)
	}
	p = loaded.GetState().Prestige
	if p.ProductionMultiplier != 1 {
		t.Errorf("never-prestiged multiplier = %v, want 1", p.ProductionMultiplier)
	}
	for _, key := range config.LegacyKit() {
		if loaded.Prestige.Owns(key) {
			t.Errorf("never-prestiged save carries %s", key)
		}
	}
}

// TestPrestigeRecordedByAge: the account counts each prestige under the age
// it came from, so a taste and a full run can be told apart.
func TestPrestigeRecordedByAge(t *testing.T) {
	isolateAccountDir(t)
	acct := &Account{AccountID: "abc"}
	acct.RecordPrestigeFrom("medieval_age")
	acct.RecordPrestigeFrom("medieval_age")
	acct.RecordPrestigeFrom("modern_age")
	acct.RecordPrestigeFrom("")
	st, _ := acct.LifetimeStats()
	if st.TotalPrestiges != 4 || st.PrestigesByAge["medieval_age"] != 2 || st.PrestigesByAge["modern_age"] != 1 || len(st.PrestigesByAge) != 2 {
		t.Errorf("stats %+v; want 4 prestiges, 2 from the Medieval Age, 1 from the Modern Age", st)
	}
	st.PrestigesByAge["medieval_age"] = 99
	if again, _ := acct.LifetimeStats(); again.PrestigesByAge["medieval_age"] != 2 {
		t.Error("LifetimeStats returned the live map")
	}
	data, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"prestiges_by_age"`) {
		t.Errorf("account stats JSON %s has no prestiges_by_age", data)
	}
}
