package game

import (
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// logCount is how many log lines carry text.
func logCount(ge *GameEngine, text string) int {
	n := 0
	for _, e := range ge.GetLogs() {
		if strings.Contains(e.Message, text) {
			n++
		}
	}
	return n
}

// A tech whose progress already reaches its price (a save from the stuck
// state, say) completes on the next tick although no scholar researches.
func TestFullProgressCompletesOnATickWithNoWork(t *testing.T) {
	ge := scholarEngine(t)
	if _, err := ge.SetResearchShare(0); err != nil {
		t.Fatal(err)
	}
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Research.progress = ge.Research.defs["tool_making"].Cost
	ge.processResearch()
	ge.mu.Unlock()
	if !ge.GetState().Research.Techs["tool_making"].Researched {
		t.Error("a tech at its price stayed in progress")
	}
	if n := logCount(ge, "Research complete: Tool Making."); n != 1 {
		t.Errorf("logged %d lines", n)
	}
}

// The overflow of a finished tech is gone: work past its price is discarded,
// and the next tech the queue starts begins at zero progress.
func TestFinishedTechsOverflowIsGone(t *testing.T) {
	ge := scholarEngine(t)
	if _, err := ge.SetResearchShare(0); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("pottery"); err != nil {
		t.Fatal(err)
	}
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	price := ge.GetState().Research.Techs["tool_making"].Cost
	ge.mu.Lock()
	if done := ge.workResearch(price * 1.5); done != "" {
		ge.finishResearch(done)
	}
	ge.runPlanTick()
	ge.mu.Unlock()
	st := ge.GetState().Research
	if !st.Techs["tool_making"].Researched {
		t.Fatal("tool_making did not finish")
	}
	if st.CurrentTech == "" || st.CurrentTech == "tool_making" {
		t.Fatalf("the queue started %q, want the next tech", st.CurrentTech)
	}
	if st.Progress != 0 {
		t.Errorf("%s starts at progress %v, want 0: the overflow carried over", st.CurrentTech, st.Progress)
	}
}

// A gift with nothing in progress is lost, and the log says so. Nothing is
// held for the next tech started.
func TestGiftWithNothingInProgressIsLost(t *testing.T) {
	ge := scholarEngine(t)
	ge.mu.Lock()
	said := ge.grantResearch(1000)
	ge.mu.Unlock()
	if said != "" {
		t.Errorf("a gift with nothing to study said %q", said)
	}
	if n := logCount(ge, "Your scholars had nothing to study, so the insight was lost."); n != 1 {
		t.Errorf("the lost gift was logged %d times, want one line", n)
	}
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if p := ge.GetState().Research.Progress; p != 0 {
		t.Errorf("a tech started after a lost gift has progress %v, want 0", p)
	}
}

// A gift lands at most half the tech in progress. A gift of half or less lands
// whole with no cut line; a bigger one adds exactly half and the log says the
// scholars gained some ground.
func TestGiftLandsAtMostHalfATech(t *testing.T) {
	ge := scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	price := ge.GetState().Research.Techs["tool_making"].Cost
	ge.mu.Lock()
	said := ge.grantResearch(price * 0.9)
	ge.mu.Unlock()
	st := ge.GetState().Research
	if math.Abs(st.Progress-price/2) > 1e-9*price {
		t.Errorf("a gift of 0.9 prices left progress %v, want exactly half the price %v", st.Progress, price/2)
	}
	if said != "50% of Tool Making" {
		t.Errorf("the capped gift said %q, want the half it landed", said)
	}
	if n := logCount(ge, "Your scholars gained some ground on Tool Making."); n != 1 {
		t.Errorf("the cut gift logged the ground line %d times, want one", n)
	}
	if st.Techs["tool_making"].Researched {
		t.Error("a gift of 0.9 prices completed a tech that was at zero")
	}

	// A gift of a quarter fits and says nothing about a cut.
	ge = scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	said = ge.grantResearch(price / 4)
	ge.mu.Unlock()
	if math.Abs(ge.GetState().Research.Progress-price/4) > 1e-9*price {
		t.Errorf("a gift of a quarter left progress %v, want %v", ge.GetState().Research.Progress, price/4)
	}
	if said != "25% of Tool Making" || logCount(ge, "gained some ground") != 0 {
		t.Errorf("a gift that fits said %q and logged a cut", said)
	}

	// Progress already past half: a capped gift completes the tech.
	ge = scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Research.progress = price * 0.6
	said = ge.grantResearch(price * 0.9)
	ge.mu.Unlock()
	if !ge.GetState().Research.Techs["tool_making"].Researched {
		t.Error("a capped gift past half did not complete the tech")
	}
	if said != "enough to complete Tool Making" {
		t.Errorf("the completing gift said %q", said)
	}
	if n := logCount(ge, "Research complete: Tool Making."); n != 1 {
		t.Errorf("the completion was logged %d times, want one line", n)
	}
	if p := ge.GetState().Research.Progress; p != 0 {
		t.Errorf("after the completion progress is %v, want 0 (the rest is lost)", p)
	}

	// Progress below half: a capped gift stops short of the price.
	ge = scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Research.progress = price * 0.3
	ge.grantResearch(price * 0.9)
	ge.mu.Unlock()
	if st := ge.GetState().Research; st.Techs["tool_making"].Researched || math.Abs(st.Progress-price*0.8) > 1e-9*price {
		t.Errorf("a capped gift from 0.3 left progress %v researched %v, want 0.8 and unfinished", st.Progress, st.Techs["tool_making"].Researched)
	}
}

// grandDiscovery fires the Grand Discovery on ge and returns the techs whose
// research-done event it published, in order.
func grandDiscovery(ge *GameEngine) []string {
	var fired []string
	ge.Bus.Subscribe(EventResearchDone, func(e EventData) {
		if key, ok := e.Payload["tech"].(string); ok {
			fired = append(fired, key)
		}
	})
	ge.applyGoodEpochEvent(config.EpochEventByKey()["grand_discovery"])
	return fired
}

// startableNow is the expected Grand Discovery pick, worked out from the
// rules alone: the techs whose age is reached and whose prerequisites are met
// (every one, and one of the either-or group), not yet researched, cheapest
// first, ties by key, at most three.
func startableNow(ge *GameEngine) []string {
	ageOrder := ge.progress.GetAgeOrder()
	var want []string
	for _, key := range sortedKeys(ge.Research.defs) {
		d := ge.Research.defs[key]
		if ge.Research.IsResearched(key) || ageOrder[d.Age] > ageOrder[ge.age] || !d.PrereqsMet(ge.Research.IsResearched) {
			continue
		}
		want = append(want, key)
	}
	sort.SliceStable(want, func(i, j int) bool {
		return ge.Research.defs[want[i]].Cost < ge.Research.defs[want[j]].Cost
	})
	if len(want) > 3 {
		want = want[:3]
	}
	return want
}

// The Grand Discovery completes up to three techs the player could start now,
// cheapest first, each through finishResearch: the log, the research-done
// event and the completion. A tech whose prerequisites are not met is skipped
// however cheap it is.
func TestGrandDiscoveryCompletesStartableTechsCheapestFirst(t *testing.T) {
	ge := scholarEngine(t)
	// Pottery is cheaper than anything startable, but its prerequisites are
	// not met.
	ge.mu.Lock()
	pottery := ge.Research.defs["pottery"]
	pottery.Cost = 1
	ge.Research.defs["pottery"] = pottery
	want := startableNow(ge)
	ge.mu.Unlock()
	if len(want) != 3 {
		t.Fatalf("staging: %d techs can be started, want 3", len(want))
	}
	if want[0] == "pottery" || want[1] == "pottery" || want[2] == "pottery" {
		t.Fatalf("staging: pottery counts as startable: %v", want)
	}

	fired := grandDiscovery(ge)

	if strings.Join(fired, ",") != strings.Join(want, ",") {
		t.Errorf("research-done fired for %v, want %v", fired, want)
	}
	st := ge.GetState().Research
	for _, key := range want {
		if !st.Techs[key].Researched {
			t.Errorf("%s was not completed", key)
		}
		name := ge.Research.defs[key].Name
		if n := logCount(ge, "Research complete: "+name+"."); n != 1 {
			t.Errorf("the completion of %s was logged %d times, want one line", name, n)
		}
	}
	if st.Techs["pottery"].Researched {
		t.Error("the Grand Discovery completed pottery, whose prerequisites are not met")
	}
}

// An either-or group nothing meets keeps a tech from being startable. With
// one such tech among the startable, the Grand Discovery completes the rest,
// which is fewer than three.
func TestGrandDiscoveryHonorsEitherOrGroups(t *testing.T) {
	ge := scholarEngine(t)
	ge.mu.Lock()
	def := ge.Research.defs["tool_making"]
	def.AnyOf = []string{"no_such_tech_a", "no_such_tech_b"}
	ge.Research.defs["tool_making"] = def
	want := startableNow(ge)
	ge.mu.Unlock()
	if len(want) != 2 {
		t.Fatalf("staging: %d techs can be started with tool_making gated, want 2: %v", len(want), want)
	}

	fired := grandDiscovery(ge)

	if strings.Join(fired, ",") != strings.Join(want, ",") {
		t.Errorf("research-done fired for %v, want %v", fired, want)
	}
	if ge.GetState().Research.Techs["tool_making"].Researched {
		t.Error("the Grand Discovery completed tool_making, whose either-or group is not met")
	}
}

// Ancient Memory starts its tech at zero progress.
func TestAncientMemoryStartsAtZero(t *testing.T) {
	ge := scholarEngine(t)
	ge.mu.Lock()
	ge.pendingMemoryTech = "tool_making"
	ge.mu.Unlock()
	if err := ge.AcceptAncientMemory(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState().Research
	if st.CurrentTech != "tool_making" || st.Progress != 0 || st.Techs["tool_making"].Researched {
		t.Errorf("Ancient Memory: in progress %q at %v, researched %v; want tool_making at 0 and unfinished", st.CurrentTech, st.Progress, st.Techs["tool_making"].Researched)
	}
	if n := logCount(ge, "Your head start covered it"); n != 0 {
		t.Errorf("the head start line appeared %d times", n)
	}
}

// twinTown is a Stone Age town with scholars and two techs planned, so one
// finishing partway through an offline step has another to start.
func twinTown(t *testing.T) *GameEngine {
	t.Helper()
	ge := scholarEngine(t)
	for _, key := range []string{"language", "primitive_writing"} {
		if err := ge.PlanAddResearch(key); err != nil {
			t.Fatal(err)
		}
	}
	ge.mu.Lock()
	ge.runPlanTick() // the first tech starts
	ge.mu.Unlock()
	return ge
}

// An offline span credits research as live ticks would: a tech that finishes
// partway through an offline step hands the rest of the step to the next
// tech, at the rate that holds after the completion. The reference is the
// same ticks played one at a time at the offline efficiency.
func TestOfflineResearchMatchesLiveTicks(t *testing.T) {
	away, live := twinTown(t), twinTown(t)
	st := away.GetState()
	perTick := st.Research.PerTick
	first := st.Research.Techs[st.Research.CurrentTech]
	if perTick <= 0 || first.Cost <= 0 {
		t.Fatalf("staging: %v work a tick on %q", perTick, st.Research.CurrentTech)
	}
	// Ticks the first tech takes at the offline efficiency, then half as
	// many again for the second, so it ends mid-step and mid-second-tech.
	t1 := int(math.Ceil(first.Cost / (perTick * OfflineEfficiency)))
	if t1%OfflineStepTicks == 0 {
		t1++
	}
	total := t1 + t1/2
	if total/OfflineStepTicks == t1/OfflineStepTicks {
		t.Fatalf("staging: the first tech ends at tick %d inside the last step of %d", t1, total)
	}

	// Live: one tick at a time.
	live.mu.Lock()
	for i := 0; i < total; i++ {
		if live.Research.currentTech != "" {
			if done := live.workResearch(OfflineEfficiency * live.researchPerTick); done != "" {
				live.finishResearch(done)
			}
		}
		live.recalculateRates()
		var s planStarts
		live.runPlan(&s)
	}
	live.mu.Unlock()

	// Away: the same ticks as one span.
	away.mu.Lock()
	elapsed := time.Duration(total)*away.tickIntervalLocked() + away.tickIntervalLocked()/2
	away.applyOfflineProgress(elapsed)
	away.mu.Unlock()

	a, l := away.GetState().Research, live.GetState().Research
	if a.CurrentTech != l.CurrentTech || a.TotalResearched != l.TotalResearched {
		t.Fatalf("away: %q with %d done; live: %q with %d done", a.CurrentTech, a.TotalResearched, l.CurrentTech, l.TotalResearched)
	}
	if a.CurrentTech == "" || a.TotalResearched != live.GetState().Research.TotalResearched {
		t.Fatalf("staging: expected the second tech in progress, got %q", a.CurrentTech)
	}
	if math.Abs(a.Progress-l.Progress) > 1e-6*math.Max(l.Progress, 1) {
		t.Errorf("progress away %v, live %v", a.Progress, l.Progress)
	}
	if math.Abs(a.PerTick-l.PerTick) > 1e-9*math.Max(l.PerTick, 1) {
		t.Errorf("work a tick away %v, live %v", a.PerTick, l.PerTick)
	}
	if math.Abs(away.Stats.ResearchWork-live.Stats.ResearchWork) > 1e-6*live.Stats.ResearchWork {
		t.Errorf("work counted away %v, live %v", away.Stats.ResearchWork, live.Stats.ResearchWork)
	}
}
