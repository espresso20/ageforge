package game

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// level5Upgrades is the prestige audit's level-5 save: five Modern Age
// prestiges bought cheapest first, 83 points spent, 3 left (the plan's
// worked example).
var level5Upgrades = map[string]int{
	"gather_boost": 3, "storage_bonus": 3, "research_speed": 3, "military_power": 3,
	"starting_food": 5, "starting_wood": 4, "population_cap": 3, "expedition_loot": 3,
	"tick_speed": 0,
}

const level5Fixture = "testdata/level5_shop_v1.json"

// TestWriteLevel5Fixture regenerates the signed level-5 fixture: a save from
// before Era Mastery and the legacy kit (no mastery, no shop version, no kit
// memory), mid-run in the Industrial Age. Run with UPDATE_FIXTURES=1; the committed bytes are what
// TestShopRefundLevel5Fixture loads, so a renamed or removed save field
// breaks its signature check.
func TestWriteLevel5Fixture(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite " + level5Fixture)
	}
	isolateAccountDir(t)
	ge := newSeededEngine(21)
	ge.StepTicks(30)
	ge.mu.Lock()
	ge.age = "industrial_age"
	ge.currentEpoch = config.EpochForAge("industrial_age")
	ge.Stats.AgesReached = []string{"primitive_age", "stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age", "renaissance_age", "colonial_age", "industrial_age"}
	ge.mu.Unlock()
	ge.mu.Lock()
	ge.Prestige.LoadState(5, 86, 3, level5Upgrades)
	ge.mu.Unlock()
	oldSave(t, ge, "level5") // from before Era Mastery and the legacy kit
	data, err := os.ReadFile(savePath("level5"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(level5Fixture, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

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

func refundLines(ge *GameEngine) []string {
	var out []string
	for _, l := range ge.GetLogs() {
		if strings.HasPrefix(l.Message, "The prestige shop changed.") {
			out = append(out, l.Message)
		}
	}
	return out
}

// TestShopRefundLevel5Fixture: the plan's worked example. A signed level-5
// save from the first shop still verifies; its 83 spent and 3 unspent old
// points become 600 new ones (5 prestiges at 120, more than 86 × 4.44 =
// 382), Available and the lifetime total both; the nine old tiers go to 0
// and keep their keys; the shop version becomes 2, with one log line. A
// second load refunds nothing, and the retired perks can't be bought.
func TestShopRefundLevel5Fixture(t *testing.T) {
	isolateAccountDir(t)
	ge := loadFixture(t, level5Fixture, "level5")
	if ge.cheaterBadge {
		t.Fatal("the signed level-5 fixture failed its signature check")
	}
	st := ge.GetState().Prestige
	if st.Available != 600 || st.TotalEarned != 600 || st.Level != 5 {
		t.Errorf("after the refund: available %d, total %d, level %d; want 600, 600, 5", st.Available, st.TotalEarned, st.Level)
	}
	if st.ShopVersion != config.PrestigeShopVersion {
		t.Errorf("shop version %d, want %d", st.ShopVersion, config.PrestigeShopVersion)
	}
	for key := range level5Upgrades {
		u, ok := st.Upgrades[key]
		if !ok || u.Tier != 0 || !u.Retired {
			t.Errorf("%s after the refund: %+v (present %v); want tier 0, retired", key, u, ok)
		}
		if tier, held := ge.Prestige.upgrades[key]; !held || tier != 0 {
			t.Errorf("%s: the upgrades map holds %d (key kept %v); want the key kept at 0", key, tier, held)
		}
	}
	want := "The prestige shop changed. Your old perks were refunded as 600 points (5 prestiges at 120 each)."
	if lines := refundLines(ge); len(lines) != 1 || lines[0] != want {
		t.Errorf("refund lines %q, want once %q", lines, want)
	}
	// Mastery seeding ran too (Pacing v2, PR 5): both one-time steps.
	if m := ge.Prestige.Mastery("atomic_age"); m != 5 {
		t.Errorf("Atomic Age mastery %d after seeding, want 5", m)
	}

	if err := ge.BuyPrestigeUpgrade("gather_boost"); err == nil || !strings.Contains(err.Error(), "old prestige shop") {
		t.Errorf("buying a retired perk: %v, want a refusal naming the old shop", err)
	}
	if err := ge.SaveGame("level5"); err != nil {
		t.Fatal(err)
	}
	if s := readSaveFromDisk(t, "level5"); s.Prestige.ShopVersion != config.PrestigeShopVersion || s.Prestige.Available != 600 {
		t.Errorf("the next save carries shop version %d, available %d", s.Prestige.ShopVersion, s.Prestige.Available)
	}
	again := NewGameEngine()
	if err := again.LoadGame("level5"); err != nil {
		t.Fatal(err)
	}
	if p := again.GetState().Prestige; p.Available != 600 || p.TotalEarned != 600 || len(refundLines(again)) != 0 || again.cheaterBadge {
		t.Errorf("second load: available %d, total %d, %d refund lines, tamper %v; want 600, 600, none, false",
			p.Available, p.TotalEarned, len(refundLines(again)), again.cheaterBadge)
	}
}

// TestShopRefundSynthetic: the same refund through the save writer the
// smoke suite uses, and a level-0 save with nothing to refund is marked
// silently.
func TestShopRefundSynthetic(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(3)
	if err := ge.WriteShopV1SaveForTest("synthetic", 5, 86, 3, level5Upgrades); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("synthetic"); err != nil {
		t.Fatal(err)
	}
	if p := loaded.GetState().Prestige; p.Available != 600 || loaded.cheaterBadge {
		t.Errorf("synthetic level-5 save: available %d, tamper %v; want 600, false", p.Available, loaded.cheaterBadge)
	}

	fresh := newSeededEngine(4)
	if err := fresh.WriteShopV1SaveForTest("level0", 0, 0, 0, nil); err != nil {
		t.Fatal(err)
	}
	zero := NewGameEngine()
	if err := zero.LoadGame("level0"); err != nil {
		t.Fatal(err)
	}
	if p := zero.GetState().Prestige; p.Available != 0 || p.ShopVersion != config.PrestigeShopVersion || len(refundLines(zero)) != 0 {
		t.Errorf("level-0 save: available %d, shop version %d, %d refund lines; want 0, %d, none",
			p.Available, p.ShopVersion, len(refundLines(zero)), config.PrestigeShopVersion)
	}
}

// TestShopRefundRule: each past prestige counts as a Modern Age run (120
// points), or the old points at 120 per 27 when that is more.
func TestShopRefundRule(t *testing.T) {
	maxed := map[string]int{}
	spent := 0
	for _, def := range config.PrestigeUpgrades() {
		if def.Retired {
			maxed[def.Key] = def.MaxTier
			for _, c := range def.Costs {
				spent += c
			}
		}
	}
	cases := []struct {
		name              string
		level, available  int
		upgrades          map[string]int
		wantOld, wantBack int
	}{
		{"level 0", 0, 0, nil, 0, 0},
		{"level 5 (worked example)", 5, 3, level5Upgrades, 86, 600},
		{"level 1, 27 points unspent", 1, 27, nil, 27, 120},
		{"level 1, 60 points (rate wins)", 1, 60, nil, 60, 266},
		{"level 38, maxed shop", 38, 0, maxed, spent, 4560},
	}
	for _, c := range cases {
		old, back := config.ShopRefund(c.level, c.available, c.upgrades)
		if old != c.wantOld || back != c.wantBack {
			t.Errorf("%s: old %d, refund %d; want %d, %d", c.name, old, back, c.wantOld, c.wantBack)
		}
	}
	if spent != 277 {
		t.Errorf("a maxed first shop cost %d points, want 277 (the frozen costs moved)", spent)
	}
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
	if n != 0 {
		t.Errorf("the new run's plan log holds %d items, want 0", n)
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
// it to MaxPlanItems and says how many waited out.
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
		ge.plan = append(ge.plan, PlanItem{Kind: PlanTrade, Key: "wood", To: "food", Count: 1, Amount: float64(i + 1)})
	}
	ge.mu.Unlock()
	ge.SetLegacyForTest(LegacyKit{Plan: plan}, true)
	if n := len(ge.GetState().Plan); n != MaxPlanItems {
		t.Errorf("the plan holds %d items, want the cap of %d", n, MaxPlanItems)
	}
	found := false
	for _, l := range ge.GetLogs() {
		if strings.HasPrefix(l.Message, "Plan Template: added ") && strings.Contains(l.Message, "The plan is full (60 items)") {
			found = true
		}
	}
	if !found {
		t.Error("no Plan Template line saying the plan was full")
	}
}

// TestResearchMemoryReplaysInOrder: with Research Memory bought, an idle
// research slot takes the first remembered tech that is available and
// affordable, in the remembered order; a tech of a later age waits in
// place; the plan's own research items go first.
func TestResearchMemoryReplaysInOrder(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(13)
	ge.SetLegacyForTest(LegacyKit{Research: []string{"stoneworking", "fire_mastery", "tool_making"}}, false)
	ge.mu.Lock()
	ge.Prestige.upgrades[config.LegacyResearch] = 1
	ge.Resources.AddStorage("knowledge", 100000)
	ge.Resources.Add("knowledge", 100000)
	ge.mu.Unlock()
	// stoneworking is a Stone Age tech: it waits; fire_mastery needs
	// tool_making: it waits; tool_making starts.
	if got := ge.GetState().Prestige.Kit.ResearchNext; got != "tool_making" {
		t.Fatalf("next remembered tech %q, want tool_making", got)
	}
	ge.mu.Lock()
	var s planStarts
	ge.runPlan(&s)
	cur := ge.Research.currentTech
	ge.mu.Unlock()
	if cur != "tool_making" || len(s.remembered) != 1 {
		t.Fatalf("research after the replay: %q (remembered %v), want tool_making", cur, s.remembered)
	}
	ge.mu.Lock()
	ge.Research.ticksLeft = 1
	ge.processResearch()
	ge.mu.Unlock()
	// A plan research item of this age keeps the slot for the plan.
	if err := ge.PlanAddResearch("fire_mastery"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Resources.Remove("knowledge", ge.Resources.Get("knowledge")) // the plan's item can't start
	s = planStarts{}
	ge.runPlan(&s)
	cur = ge.Research.currentTech
	ge.mu.Unlock()
	if cur != "" || len(s.remembered) != 0 {
		t.Errorf("with a plan research item waiting, memory started %q", cur)
	}
	if got := ge.GetState().Research.TotalResearched; got != 1 {
		t.Errorf("researched %d techs, want 1", got)
	}
}

// TestResearchOrderRecordsGrandDiscovery: the run's order holds techs in
// the order they finished, Grand Discovery's free techs included, and it
// becomes the remembered order at prestige, ahead of older techs.
func TestResearchOrderRecordsGrandDiscovery(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(14)
	ge.SetLegacyForTest(LegacyKit{Research: []string{"pottery", "tool_making"}}, false)
	ge.mu.Lock()
	ge.Research.currentTech, ge.Research.ticksLeft = "tool_making", 1
	ge.processResearch()
	free := ge.Research.ForceCompleteN(1, "primitive_age", ge.progress.GetAgeOrder())
	order := ge.Research.Order()
	ge.mu.Unlock()
	if want := append([]string{"tool_making"}, free...); !reflect.DeepEqual(order, want) {
		t.Fatalf("run order %v, want %v", order, want)
	}
	_ = ge.SummonHarbingerForTest("medieval_age")
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	want := append(append([]string(nil), order...), "pottery")
	if got := ge.LegacyForTest().Research; !reflect.DeepEqual(got, want) {
		t.Errorf("remembered order %v, want %v", got, want)
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
			ge.Prestige.upgrades[config.LegacyFactions] = 1
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

// TestWorkerSharesCarryOver: with Worker Shares bought, the shares set in a
// run are set again in the next; without it they go back to auto.
func TestWorkerSharesCarryOver(t *testing.T) {
	isolateAccountDir(t)
	for _, own := range []bool{true, false} {
		ge := newSeededEngine(16)
		if _, err := ge.SetWorkerShare("knowledge", 40); err != nil {
			t.Fatal(err)
		}
		if own {
			ge.mu.Lock()
			ge.Prestige.upgrades[config.LegacyWorkers] = 1
			ge.mu.Unlock()
		}
		_ = ge.SummonHarbingerForTest("medieval_age")
		if err := ge.DoPrestige(); err != nil {
			t.Fatal(err)
		}
		got := ge.WorkerShares()
		if own && got["knowledge"] != 40 {
			t.Errorf("owned: shares after prestige %v, want knowledge 40", got)
		}
		if !own && len(got) != 0 {
			t.Errorf("not owned: shares after prestige %v, want none (auto)", got)
		}
	}
}

// TestKitBoughtAfterPrestigeWorksNow: the kit remembers whether or not it
// is bought, so Plan Template and Worker Shares bought in the new run's
// first age apply at once.
func TestKitBoughtAfterPrestigeWorksNow(t *testing.T) {
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
	if n := len(ge.GetState().Plan); n != 0 {
		t.Fatalf("without the kit the new run's plan holds %d items", n)
	}
	for _, key := range []string{config.LegacyPlan, config.LegacyWorkers} {
		if err := ge.BuyPrestigeUpgrade(key); err != nil {
			t.Fatal(err)
		}
	}
	if p := ge.GetState().Plan; len(p) != 1 || p[0].Key != "hut" || p[0].Count != 2 {
		t.Errorf("plan after buying the template: %+v, want 2 huts", p)
	}
	if got := ge.WorkerShares(); got["food"] != 50 {
		t.Errorf("shares after buying Worker Shares: %v, want food 50", got)
	}
	if err := ge.BuyPrestigeUpgrade(config.LegacyPlan); err == nil {
		t.Error("a second Plan Template was sold")
	}
}

// TestSuccumbReappliesTemplate: a Succumb folds the run's plan into the
// template and starts the rebuild with the first age's slice.
func TestSuccumbReappliesTemplate(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(18)
	ge.mu.Lock()
	ge.Prestige.upgrades[config.LegacyPlan] = 1
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

// TestLegacySaveRoundTrip: the shop version, the kit's memory, the run's
// plan log and research order survive a save and load unchanged.
func TestLegacySaveRoundTrip(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(19)
	kit := LegacyKit{
		Plan:     []PlanTemplateItem{{Age: "stone_age", Kind: PlanBuild, Key: "stone_pit", Count: 4}, {Age: "stone_age", Kind: PlanTrade, Key: "wood", To: "stone", Amount: 50}, {Age: "stone_age", Kind: PlanAdvance}},
		Research: []string{"tool_making", "fire_mastery"},
		Factions: []string{"riverlands_tribes"},
		Shares:   map[string]float64{"knowledge": 30},
	}
	ge.SetLegacyForTest(kit, true)
	if _, err := ge.PlanAddBuild("hut", 2); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Research.currentTech, ge.Research.ticksLeft = "tool_making", 1
	ge.processResearch()
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
	gotLog, gotOrder, ver := clonePlanTemplate(loaded.planLog), loaded.Research.Order(), loaded.Prestige.shopVersion
	owned := loaded.Prestige.Owns(config.LegacyFactions)
	loaded.mu.RUnlock()
	if !reflect.DeepEqual(gotLog, log) || !reflect.DeepEqual(gotOrder, []string{"tool_making"}) || ver != config.PrestigeShopVersion || !owned {
		t.Errorf("after load: plan log %+v (want %+v), order %v, shop version %d, owns Old Friends %v", gotLog, log, gotOrder, ver, owned)
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
	for _, field := range []string{`"legacy_plan"`, `"legacy_research"`, `"legacy_factions"`, `"legacy_shares"`, `"plan_log"`} {
		if strings.Contains(string(data), field) {
			t.Errorf("a new game's save writes %s with nothing in it", field)
		}
	}
	if !strings.Contains(string(data), `"shop_version": 2`) {
		t.Error("a new game's save does not write its shop version")
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
	acct.RecordPrestige()
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

// TestPlanLogTradeRemoval: a trade item removed after it bought something
// stays written as what it bought; one that bought nothing goes.
func TestPlanLogTradeRemoval(t *testing.T) {
	ge := newSeededEngine(22)
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.planLog = []PlanTemplateItem{
		{Age: "primitive_age", Kind: PlanTrade, Key: "wood", To: "food", Amount: 500},
		{Age: "primitive_age", Kind: PlanTrade, Key: "food", To: "wood"},
	}
	ge.unlogPlanItemLocked(PlanItem{Kind: PlanTrade, Key: "wood", To: "food", Amount: 380, Got: 120})
	ge.unlogPlanItemLocked(PlanItem{Kind: PlanTrade, Key: "food", To: "wood"})
	want := []PlanTemplateItem{{Age: "primitive_age", Kind: PlanTrade, Key: "wood", To: "food", Amount: 120}}
	if !reflect.DeepEqual(ge.planLog, want) {
		t.Errorf("plan log after removing the trades: %+v, want %+v", ge.planLog, want)
	}
}
