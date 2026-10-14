package game

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// scholarEngine is a Stone Age town with some Story Circles and Elders' Halls
// standing (and a few huts' worth of people who are not their staff).
func scholarEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(3)
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.age = "stone_age"
	ge.Workers.SetAge("stone_age")
	ge.Workers.UnlockType("worker")
	ge.Buildings.counts["story_circle"] = 2
	ge.Buildings.counts["elders_hall"] = 1
	ge.recalculateRates()
	return ge
}

// Scholars exist from the moment a knowledge building stands, take no
// population and need no assignment.
func TestScholarsComeWithTheBuilding(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	want := 2*2 + 1*2 // two Story Circles of 2 and an Elders' Hall of 2
	if st.Scholars.Total != want {
		t.Fatalf("scholars = %d, want %d", st.Scholars.Total, want)
	}
	if pop := st.Workers.TotalPop; pop != 0 {
		t.Errorf("the population is %d: scholars are not staffed from it", pop)
	}
	if bs := st.Buildings["story_circle"]; bs.Scholars != 2 || bs.ScholarCap != config.KnowledgeCap("story_circle") || bs.WorkerCapacity != 0 {
		t.Errorf("story circle state: scholars %d, cap %d (want %d), worker slots %d", bs.Scholars, bs.ScholarCap, config.KnowledgeCap("story_circle"), bs.WorkerCapacity)
	}
	if err := ge.AssignWorker("story_circle", 1); err == nil {
		t.Error("a worker was assigned to a knowledge building")
	}
	kinds := map[string]int{}
	for _, g := range st.Scholars.Kinds {
		kinds[g.Kind] = g.Count
	}
	if kinds["Shaman"] != 4 || kinds["Elder"] != 2 {
		t.Errorf("scholars by kind = %v, want 4 Shamans and 2 Elders (the classes of their buildings' ages)", kinds)
	}
}

// Research accrues to the one tech in progress and completes at its price;
// there is no wait but the work.
func TestResearchAccruesAndCompletes(t *testing.T) {
	ge := scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatalf("a tech with no price to afford: %v", err)
	}
	if err := ge.StartResearch("pottery"); err == nil {
		t.Error("two techs in progress at once")
	}
	st := ge.GetState()
	price := st.Research.Price
	perTick := st.Research.PerTick
	if price <= 0 || perTick <= 0 {
		t.Fatalf("price %v, work a tick %v", price, perTick)
	}
	want := int(math.Ceil(price / perTick))
	if st.Research.TicksLeft != want || st.Research.NeedsScholars {
		t.Fatalf("ticks left %d (needs scholars %v), want %d", st.Research.TicksLeft, st.Research.NeedsScholars, want)
	}
	// Half way: a bar of half.
	for i := 0; i < want/2; i++ {
		ge.mu.Lock()
		ge.processResearch()
		ge.mu.Unlock()
	}
	if got := ge.GetState().Research.Progress; math.Abs(got-perTick*float64(want/2)) > 1e-9*price {
		t.Errorf("progress %v after %d ticks, want %v", got, want/2, perTick*float64(want/2))
	}
	for i := want / 2; i < want; i++ {
		ge.mu.Lock()
		ge.processResearch()
		ge.mu.Unlock()
	}
	if !ge.Research.IsResearched("tool_making") || ge.Research.currentTech != "" {
		t.Errorf("after the full time tool_making researched %v, in progress %q", ge.Research.IsResearched("tool_making"), ge.Research.currentTech)
	}
}

// The research setting splits the scholars exactly; with nothing to research
// every scholar teaches; with no scholar researching a tech says it needs
// scholars.
func TestScholarPostsSplitExactly(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	if st.Scholars.Researching != 0 || st.Scholars.Teaching != st.Scholars.Total {
		t.Fatalf("with nothing to research, researching %d teaching %d of %d", st.Scholars.Researching, st.Scholars.Teaching, st.Scholars.Total)
	}
	if st.Research.PerTick != 0 {
		t.Errorf("idle scholars do %v work", st.Research.PerTick)
	}
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	full := ge.GetState().Research.PerTick
	for _, pct := range []int{100, 70, 50, 40, 0} {
		if _, err := ge.SetResearchShare(pct); err != nil {
			t.Fatal(err)
		}
		st = ge.GetState()
		n := st.Scholars.Total
		wantR := (n*pct + 50) / 100
		if st.Scholars.Researching != wantR || st.Scholars.Teaching != n-wantR || st.Scholars.Share != pct {
			t.Errorf("%d%%: researching %d teaching %d, want %d and %d", pct, st.Scholars.Researching, st.Scholars.Teaching, wantR, n-wantR)
		}
		if want := full * float64(wantR) / float64(n); math.Abs(st.Research.PerTick-want) > 1e-9*full {
			t.Errorf("%d%%: %v work a tick, want %v", pct, st.Research.PerTick, want)
		}
	}
	st = ge.GetState()
	if !st.Research.NeedsScholars || st.Research.TicksLeft != 0 {
		t.Errorf("with no scholar researching: needs scholars %v, ticks left %d", st.Research.NeedsScholars, st.Research.TicksLeft)
	}
	if _, err := ge.SetResearchShare(101); err == nil {
		t.Error("a share of 101 percent was accepted")
	}
}

// A tech not in progress shows the time it would take the scholars the share
// setting puts on research, even while every scholar teaches for want of a
// tech. With no scholar or a share of 0 it shows no time.
func TestTechTimeInAnIdleTownCountsTheScholarsWhoWouldResearch(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	if st.Scholars.Researching != 0 || st.Research.PerTick != 0 {
		t.Fatalf("staging: the town should be idle: %+v", st.Scholars)
	}
	n := st.Scholars.Total
	cost := st.Research.Techs["tool_making"].Cost
	if n <= 0 || cost <= 0 {
		t.Fatalf("staging: %d scholars, tool_making costs %v", n, cost)
	}
	ge.mu.RLock()
	full := ge.researchFull
	ge.mu.RUnlock()
	for _, pct := range []int{100, 50} {
		if _, err := ge.SetResearchShare(pct); err != nil {
			t.Fatal(err)
		}
		st = ge.GetState()
		if st.Scholars.Researching != 0 {
			t.Fatalf("%d%%: the idle town should still teach, %d research", pct, st.Scholars.Researching)
		}
		research := (n*pct + 50) / 100
		want := int(math.Ceil(cost / (full * float64(research) / float64(n))))
		got := st.Research.Techs["tool_making"].TicksNeeded
		if got <= 0 || got != want {
			t.Errorf("%d%%: idle tool_making takes %d ticks, want %d", pct, got, want)
		}
		// Starting it changes nothing about the time.
		if err := ge.StartResearch("tool_making"); err != nil {
			t.Fatal(err)
		}
		if after := ge.GetState().Research.TotalTicks; after != got {
			t.Errorf("%d%%: once started it takes %d ticks, before it said %d", pct, after, got)
		}
		if err := ge.CancelResearch(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ge.SetResearchShare(0); err != nil {
		t.Fatal(err)
	}
	if got := ge.GetState().Research.Techs["tool_making"].TicksNeeded; got != 0 {
		t.Errorf("with a share of 0 a tech shows %d ticks", got)
	}
	empty := newSeededEngine(3)
	if got := empty.GetState().Research.Techs["tool_making"].TicksNeeded; got != 0 {
		t.Errorf("with no scholars a tech shows %d ticks", got)
	}
}

// Teaching raises what workers make, by the age's step in proportion to the
// share of the scholars that teach, and it is a source of the worker output
// pool.
func TestTeachingIsDelivered(t *testing.T) {
	ge := scholarEngine(t)
	if got, want := ge.teachingBonus(), config.TeachingBonus(1); got != want {
		t.Fatalf("every scholar teaching in the Stone Age: %v, want %v", got, want)
	}
	found := false
	for _, m := range ge.buildResolver().mods["gather_rate"] {
		if m.Source == scholarTeachSource && m.Op == OpAdd {
			found = true
			if m.Value != config.TeachingBonus(1) {
				t.Errorf("the source gives %v, want %v", m.Value, config.TeachingBonus(1))
			}
		}
	}
	if !found {
		t.Error("teaching is no source of the worker output pool")
	}
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if got := ge.teachingBonus(); got != 0 {
		t.Errorf("with every scholar researching teaching gives %v", got)
	}
	if _, err := ge.SetResearchShare(50); err != nil {
		t.Fatal(err)
	}
	n := ge.scholarTotal()
	_, teach := ge.scholarPosts(n)
	if got, want := ge.teachingBonus(), config.TeachingBonus(1)*float64(teach)/float64(n); math.Abs(got-want) > 1e-12 {
		t.Errorf("half the scholars teaching give %v, want %v", got, want)
	}
}

// A queue moves to the next tech by itself, and a tech planned for an age not
// reached does not keep the scholars from teaching.
func TestQueueAdvances(t *testing.T) {
	ge := scholarEngine(t)
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("fire_mastery"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.runPlanTick()
	ge.recalculateRates()
	ge.mu.Unlock()
	first := ge.Research.currentTech
	if first == "" {
		t.Fatal("the plan started nothing")
	}
	ge.mu.Lock()
	ge.finishResearch(ge.workResearch(ge.Research.defs[first].Cost))
	ge.runPlanTick()
	ge.mu.Unlock()
	if ge.Research.currentTech == "" || ge.Research.currentTech == first {
		t.Errorf("the queue did not move on: researching %q after %q", ge.Research.currentTech, first)
	}
}

// A knowledge building cannot be built past its cap.
func TestKnowledgeBuildingStopsAtItsCap(t *testing.T) {
	ge := scholarEngine(t)
	cap := config.KnowledgeCap("elders_hall")
	ge.mu.Lock()
	ge.Buildings.counts["elders_hall"] = cap
	ge.mu.Unlock()
	if bs := ge.GetState().Buildings["elders_hall"]; !bs.AtMaxCount || bs.CanBuild {
		t.Errorf("at its cap: at max %v, can build %v", bs.AtMaxCount, bs.CanBuild)
	}
	ge.Resources.Add("wood", 1e6)
	ge.Resources.Add("stone", 1e6)
	if err := ge.BuildBuilding("elders_hall"); err == nil {
		t.Error("a copy was built past the cap")
	}
	if _, err := ge.PlanAddBuild("elders_hall", 1); err == nil {
		t.Error("a copy past the cap was planned")
	}
}

// Everything that gave knowledge gives work to the tech in progress, or a head
// start when none is.
func TestResearchGrantsLandOnTheTechInProgress(t *testing.T) {
	ge := scholarEngine(t)
	// None in progress: a head start, kept.
	if said := ge.grantResearch(100); !strings.Contains(said, "head start") {
		t.Errorf("a gift with nothing in progress said %q", said)
	}
	if ge.Research.head != 100 {
		t.Fatalf("head start %v, want 100", ge.Research.head)
	}
	// It starts the next tech, never more than its price.
	ge.Research.head = 1e12
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	price := ge.Research.defs["tool_making"].Cost
	if ge.Research.progress != price || ge.Research.head != 0 {
		t.Errorf("a huge head start gave progress %v (price %v), head %v after it", ge.Research.progress, price, ge.Research.head)
	}
	ge.Research.progress = 0
	// In progress: the share of the tech.
	said := ge.grantResearch(price / 4)
	if !strings.Contains(said, "25%") {
		t.Errorf("a quarter of the tech said %q", said)
	}
	if math.Abs(ge.Research.progress-price/4) > 1e-9 {
		t.Errorf("progress %v, want %v", ge.Research.progress, price/4)
	}
	// Enough to finish it.
	if said := ge.grantResearch(price); !strings.Contains(said, "complete") || !ge.Research.IsResearched("tool_making") {
		t.Errorf("a gift of a whole tech said %q, researched %v", said, ge.Research.IsResearched("tool_making"))
	}
}

// Old saves. Each is the old game's, signed by it.

const scholarsBusyFixture = "testdata/pre_scholars_medieval.json"
const scholarsIdleFixture = "testdata/pre_scholars_idle.json"

func scholarsFixtureView(t *testing.T, path string) *GameEngine {
	t.Helper()
	return scholarsFixtureLoad(t, path, true)
}

// scholarsFixtureLoad loads a fixture as a player does (view false: the
// notices are said, and the time away is played) or only looks at it.
func scholarsFixtureLoad(t *testing.T, path string, view bool) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("scholars_old"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	if err := ge.loadSave("scholars_old", view); err != nil {
		t.Fatal(err)
	}
	return ge
}

func TestScholarsSave_IsTheOldGamesAndSigned(t *testing.T) {
	for _, path := range []string{scholarsBusyFixture, scholarsIdleFixture, "testdata/pre_knowledge_store_medieval.json"} {
		save := loadSaveFile(t, path)
		if save.ScholarRules != 0 {
			t.Errorf("%s carries scholar_rules %d: it is not from before the scholars", path, save.ScholarRules)
		}
		if ok, _ := verifySave(&save); !ok || save.Signature == "" {
			t.Errorf("%s is not signed and valid as the old game wrote it", path)
		}
	}
}

// An old save loads with no tamper flag, keeps every copy even above the new
// cap, returns its assigned workers to the pool, completes the tech that was
// paid for and waiting, and turns the knowledge it held into a head start for
// the next tech. The log says so once.
func TestScholarsSave_LoadsClean(t *testing.T) {
	save := loadSaveFile(t, scholarsBusyFixture)
	ge := scholarsFixtureView(t, scholarsBusyFixture)
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}
	for key, n := range save.Buildings {
		if got := ge.Buildings.GetCount(key); got != n {
			t.Errorf("%s: %d copies after the load, the save has %d", key, got, n)
		}
	}
	if n := ge.Buildings.GetCount("monastery_library"); n <= config.KnowledgeCap("monastery_library") {
		t.Fatalf("setup: %d Monastery Libraries is not above the cap %d", n, config.KnowledgeCap("monastery_library"))
	}
	if got := ge.Workers.GetAssignedCount("worker", "monastery_library") + ge.Workers.GetAssignedCount("worker", "library"); got != 0 {
		t.Errorf("%d workers still assigned to knowledge buildings", got)
	}
	if want, got := save.Workers["worker"].Count, ge.Workers.TotalPop(); got != want || ge.Workers.IdleCount("worker") != want {
		t.Errorf("population %d, idle %d, want all %d idle", got, ge.Workers.IdleCount("worker"), want)
	}
	if !ge.Research.IsResearched(save.Research.CurrentTech) || ge.Research.currentTech != "" {
		t.Errorf("the tech that was paid for (%s) researched %v, in progress %q: it completes on load", save.Research.CurrentTech, ge.Research.IsResearched(save.Research.CurrentTech), ge.Research.currentTech)
	}
	if got := ge.Research.head; got != save.Resources["knowledge"] {
		t.Errorf("head start %v, the save held %v knowledge", got, save.Resources["knowledge"])
	}
	if _, ok := ge.Resources.resources["knowledge"]; ok {
		t.Error("the town holds a knowledge resource")
	}
	// The head start goes to the next tech, never more than its price.
	if err := ge.StartResearch("steel_forging"); err != nil {
		t.Fatal(err)
	}
	price := ge.Research.defs["steel_forging"].Cost
	if ge.Research.progress != min(save.Resources["knowledge"], price) {
		t.Errorf("the next tech started with %v of %v, want %v", ge.Research.progress, price, min(save.Resources["knowledge"], price))
	}
	// A save written after carries no old fields and no notice.
	if err := ge.SaveGame("scholars_new"); err != nil {
		t.Fatal(err)
	}
	again := loadSaveFile(t, savePath("scholars_new"))
	if again.ScholarRules != scholarRules || again.Research.TicksLeft != 0 || again.Research.TotalTicks != 0 {
		t.Errorf("the new save: scholar rules %d, ticks %d/%d", again.ScholarRules, again.Research.TicksLeft, again.Research.TotalTicks)
	}
	if _, held := again.Resources["knowledge"]; held {
		t.Error("the new save holds knowledge")
	}
	back := NewGameEngine()
	if err := back.loadSave("scholars_new", true); err != nil {
		t.Fatal(err)
	}
	if back.cheaterBadge {
		t.Error("reloading the new save raised the tamper flag")
	}
	if back.Research.currentTech != "steel_forging" || back.Research.progress != ge.Research.progress {
		t.Errorf("reloaded research: %q at %v, want steel_forging at %v", back.Research.currentTech, back.Research.progress, ge.Research.progress)
	}
}

// Played, not only viewed, an old save says once that knowledge is no longer
// stored; its next save never says it again.
func TestScholarsSave_SaysItOnce(t *testing.T) {
	ge := scholarsFixtureLoad(t, scholarsBusyFixture, false)
	if ge.cheaterBadge {
		t.Fatal("tamper flag")
	}
	if n := countLines(ge, scholarsNotice); n != 1 {
		t.Fatalf("the notice was said %d times, want once", n)
	}
	if err := ge.SaveGame("scholars_new"); err != nil {
		t.Fatal(err)
	}
	back := NewGameEngine()
	if err := back.loadSave("scholars_new", false); err != nil {
		t.Fatal(err)
	}
	if back.cheaterBadge || countLines(back, scholarsNotice) != 0 {
		t.Errorf("reloading the new save: tamper flag %v, notice %d times", back.cheaterBadge, countLines(back, scholarsNotice))
	}
}

// A save with nothing in progress and nothing planned keeps its knowledge as
// a head start and has all its scholars teaching.
func TestScholarsSave_IdleTownTeaches(t *testing.T) {
	ge := scholarsFixtureView(t, scholarsIdleFixture)
	if ge.cheaterBadge {
		t.Fatal("tamper flag on the idle fixture")
	}
	st := ge.GetState()
	if st.Scholars.Total == 0 || st.Scholars.Researching != 0 || st.Scholars.Teaching != st.Scholars.Total {
		t.Errorf("scholars %d, researching %d, teaching %d: all should teach", st.Scholars.Total, st.Scholars.Researching, st.Scholars.Teaching)
	}
	if st.Research.HeadStart <= 0 {
		t.Error("the knowledge it held did not become a head start")
	}
}

// The store's own old fixture (a save from the game that kept knowledge in a
// store) loads clean too.
func TestScholarsSave_FromTheKnowledgeStoreLoadsClean(t *testing.T) {
	ge := scholarsFixtureLoad(t, "testdata/pre_knowledge_store_medieval.json", false)
	if ge.cheaterBadge {
		t.Fatal("tamper flag on the knowledge store fixture")
	}
	if ge.Research.head <= 0 || countLines(ge, scholarsNotice) != 1 {
		t.Errorf("head start %v, notice said %d times", ge.Research.head, countLines(ge, scholarsNotice))
	}
}
