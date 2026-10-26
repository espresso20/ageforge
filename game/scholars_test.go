package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
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

// Scholars exist from the moment a knowledge building stands, are part of
// the population and need no assignment.
func TestScholarsComeWithTheBuilding(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	want := 2*2 + 1*2 // two Story Circles of 2 and an Elders' Hall of 2
	if st.Scholars.Total != want {
		t.Fatalf("scholars = %d, want %d", st.Scholars.Total, want)
	}
	// They are people (population.go): the town is its scholars, and the
	// workers' pool is empty.
	if pop, idle := st.Workers.TotalPop, st.Workers.TotalIdle; pop != want || idle != 0 {
		t.Errorf("the population is %d (%d in the reserve): want %d, all of them scholars", pop, idle, want)
	}
	if bs := st.Buildings["story_circle"]; bs.Scholars != 2 || bs.FullRateCopies != config.KnowledgeCap("story_circle") || bs.WorkerCapacity != 0 {
		t.Errorf("story circle state: scholars %d, copies at full rate %d (want %d), worker slots %d", bs.Scholars, bs.FullRateCopies, config.KnowledgeCap("story_circle"), bs.WorkerCapacity)
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
	noErr(t, ge.StartResearch("tool_making"))
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
		noErr(t, ge.StartResearch("tool_making"))
		if after := ge.GetState().Research.TotalTicks; after != got {
			t.Errorf("%d%%: once started it takes %d ticks, before it said %d", pct, after, got)
		}
		noErr(t, ge.CancelResearch())
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
	noErr(t, ge.StartResearch("tool_making"))
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
	noErr(t, ge.PlanAddResearch("tool_making"))
	noErr(t, ge.PlanAddResearch("fire_mastery"))
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

// A knowledge building builds past its cap: nothing refuses the copy, it costs
// the normal climb, and each copy past the cap researches at half rate.
func TestKnowledgeBuildingBuildsPastItsCapAtHalfRate(t *testing.T) {
	ge := scholarEngine(t)
	cap := config.KnowledgeCap("elders_hall")
	def := ge.Buildings.defs["elders_hall"]
	var per float64
	for _, e := range def.Effects {
		if e.Type == "research" {
			per = e.Value
		}
	}
	if per <= 0 || cap <= 1 {
		t.Fatalf("setup: the Elders' Hall researches %v a copy with a cap of %d", per, cap)
	}
	baseAt := func(copies int) float64 {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.Buildings.counts["elders_hall"] = copies
		ge.Buildings.counts["story_circle"] = 0
		ge.recalculateRates()
		return ge.researchBase
	}
	// Up to the cap every copy researches at full rate, exactly.
	if got, want := baseAt(cap), float64(per*float64(cap)); got != want {
		t.Errorf("%d copies at the cap research %v, want %v", cap, got, want)
	}
	// Past it each copy adds half.
	if got, want := baseAt(cap+4), float64(per*float64(float64(cap)+2)); math.Abs(got-want) > 1e-9*want {
		t.Errorf("%d copies research %v, want %v: %d at full rate and 4 at half", cap+4, got, want, cap)
	}
	// Nothing refuses a copy at the cap, and the copy costs the next step of the climb.
	baseAt(cap)
	if bs := ge.GetState().Buildings["elders_hall"]; bs.AtMaxCount {
		t.Errorf("at the cap the building reads as at its limit")
	}
	ge.mu.Lock()
	ge.Buildings.unlocked["elders_hall"] = true
	ge.mu.Unlock()
	ge.Resources.Add("wood", config.Trillion)
	ge.Resources.Add("stone", config.Trillion)
	want := ge.Buildings.GetCost("elders_hall")
	for res, base := range def.BaseCost {
		if exact := math.Floor(base * math.Pow(def.CostScale, float64(cap))); math.Abs(want[res]-exact) > 1 {
			t.Errorf("the copy past the cap costs %v %s, want the normal climb %v", want[res], res, exact)
		}
	}
	if bs := ge.GetState().Buildings["elders_hall"]; !bs.CanBuild {
		t.Error("a copy past the cap cannot be built")
	}
	before := ge.Resources.Get("wood")
	if err := ge.BuildBuilding("elders_hall"); err != nil {
		t.Fatalf("a copy past the cap was refused: %v", err)
	}
	if ge.Resources.Get("wood") >= before {
		t.Error("the copy past the cap cost no wood")
	}
	if n, err := ge.PlanAddBuild("elders_hall", 2); err != nil || n != 2 {
		t.Errorf("planning two copies past the cap added %d: %v", n, err)
	}
}

// Everything that gave knowledge gives work to the tech in progress, at most
// half a tech at a time. With none in progress it is lost.
func TestResearchGrantsLandOnTheTechInProgress(t *testing.T) {
	ge := scholarEngine(t)
	// None in progress: the gift is lost, and the log says so.
	ge.mu.Lock()
	said := ge.grantResearch(100)
	ge.mu.Unlock()
	if said != "" || ge.Research.currentTech != "" || ge.Research.progress != 0 {
		t.Errorf("a gift with nothing in progress said %q, in progress %q at %v", said, ge.Research.currentTech, ge.Research.progress)
	}
	if n := logCount(ge, "Your scholars had nothing to study, so the insight was lost."); n != 1 {
		t.Errorf("the lost gift was logged %d times, want one line", n)
	}
	noErr(t, ge.StartResearch("fire_mastery"))
	price := ge.Research.defs["fire_mastery"].Cost
	ge.Research.progress = 0
	// In progress: the share of the tech.
	said = ge.grantResearch(price / 4)
	if !strings.Contains(said, "25%") {
		t.Errorf("a quarter of the tech said %q", said)
	}
	if math.Abs(ge.Research.progress-price/4) > 1e-9 {
		t.Errorf("progress %v, want %v", ge.Research.progress, price/4)
	}
	// A gift of a whole tech lands half of it: three quarters, not finished.
	said = ge.grantResearch(price)
	if ge.Research.IsResearched("fire_mastery") || math.Abs(ge.Research.progress-price*0.75) > 1e-9*price {
		t.Errorf("a gift of a whole tech at a quarter: progress %v, researched %v, want 0.75 and unfinished", ge.Research.progress, ge.Research.IsResearched("fire_mastery"))
	}
	if !strings.Contains(said, "50%") {
		t.Errorf("a capped gift said %q", said)
	}
	// Past half, the next capped gift completes it.
	if said := ge.grantResearch(price); !strings.Contains(said, "complete") || !ge.Research.IsResearched("fire_mastery") {
		t.Errorf("a gift past half said %q, researched %v", said, ge.Research.IsResearched("fire_mastery"))
	}
}

// Enlightenment, the faction boon, says it speeds up research. Granted, it
// must raise the work the scholars do each tick by its size and nothing else
// (the time pool is a different pool), and it must end with the boon.
func TestEnlightenmentBoonRaisesTheScholarsWork(t *testing.T) {
	var def boon.Def
	for _, d := range boon.Catalog() {
		if d.Name == "Enlightenment" {
			def = d
		}
	}
	if def.Name == "" {
		t.Fatal("no Enlightenment boon in the catalog")
	}
	ge := scholarEngine(t)
	noErr(t, ge.StartResearch("tool_making"))
	before := ge.GetState().Research.PerTick
	if before <= 0 {
		t.Fatalf("staging: the scholars do %v work", before)
	}
	const size = 0.2
	ge.mu.Lock()
	ge.applyRolledFactionBoon(config.FactionDef{Name: "Test Folk", Key: "test_folk"},
		boon.Boon{Kind: boon.RateBuff, Polarity: boon.Positive, Name: def.Name, Resource: def.Resource, Magnitude: size, DurationTicks: 100})
	ge.recalculateRates()
	ge.mu.Unlock()
	after := ge.GetState().Research.PerTick
	if want := before * (1 + size); math.Abs(after-want) > 1e-9*want {
		t.Errorf("Enlightenment of +%.0f%% moved the scholars' work from %v to %v, want %v", size*100, before, after, want)
	}
}
