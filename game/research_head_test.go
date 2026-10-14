package game

import (
	"math"
	"strings"
	"testing"
	"time"
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

// A head start that covers a tech's whole price completes it at the start,
// even with no scholar researching, and the queue goes on to the next tech.
func TestHeadStartThatCoversThePriceCompletesAtOnce(t *testing.T) {
	for _, share := range []int{0, 100} {
		ge := scholarEngine(t)
		if _, err := ge.SetResearchShare(share); err != nil {
			t.Fatal(err)
		}
		price := ge.GetState().Research.Techs["tool_making"].Cost
		ge.mu.Lock()
		ge.Research.head = price * 1.5
		ge.mu.Unlock()
		if err := ge.PlanAddResearch("pottery"); err != nil {
			t.Fatal(err)
		}
		if err := ge.StartResearch("tool_making"); err != nil {
			t.Fatal(err)
		}
		st := ge.GetState()
		if !st.Research.Techs["tool_making"].Researched || st.Research.CurrentTech != "" {
			t.Fatalf("share %d: a head start of 1.5 prices left tool_making unfinished (in progress %q)", share, st.Research.CurrentTech)
		}
		if n := logCount(ge, "Research complete: Tool Making."); n != 1 {
			t.Errorf("share %d: the completion was logged %d times, want one line", share, n)
		}
		// The next tick starts the queued tech.
		ge.mu.Lock()
		ge.processResearch()
		ge.runPlanTick()
		ge.mu.Unlock()
		if got := ge.GetState().Research.CurrentTech; got == "" || got == "tool_making" {
			t.Errorf("share %d: the queue should have started its next tech, in progress %q", share, got)
		}
		if n := logCount(ge, "Research complete: Tool Making."); n != 1 {
			t.Errorf("share %d: tool_making completed again (%d lines)", share, n)
		}
	}
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

// What a head start holds never passes the price of the dearest tech that
// could be started; the state and the save show the cap, and a grant that
// overflows says part is lost.
func TestHeadStartIsBoundedWhileHeld(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	top := 0.0
	for _, ts := range st.Research.Techs {
		if ts.Available {
			top = math.Max(top, ts.Cost)
		}
	}
	if top <= 0 {
		t.Fatal("staging: no tech can be started")
	}
	ge.mu.Lock()
	said := ge.grantResearch(top * 10)
	ge.mu.Unlock()
	if !strings.Contains(said, "part is lost") {
		t.Errorf("an overflowing grant said %q", said)
	}
	if got := ge.GetState().Research.HeadStart; math.Abs(got-top) > 1e-9*top {
		t.Errorf("the state holds a head start of %v, want the cap %v", got, top)
	}
	if got := ge.buildSaveSnapshot().Research.HeadStart; math.Abs(got-top) > 1e-9*top {
		t.Errorf("the save holds a head start of %v, want the cap %v", got, top)
	}
	// A grant that fits is kept whole and says nothing about loss.
	ge.mu.Lock()
	ge.Research.head = 0
	said = ge.grantResearch(top / 4)
	ge.mu.Unlock()
	if strings.Contains(said, "lost") || math.Abs(ge.GetState().Research.HeadStart-top/4) > 1e-9*top {
		t.Errorf("a grant that fits said %q and left %v", said, ge.GetState().Research.HeadStart)
	}
	// Even a value that got in by another door is not shown or saved above the cap.
	ge.mu.Lock()
	ge.Research.head = top * 3
	ge.mu.Unlock()
	if got := ge.GetState().Research.HeadStart; got > top*(1+1e-9) {
		t.Errorf("state shows %v above the cap %v", got, top)
	}
	if got := ge.buildSaveSnapshot().Research.HeadStart; got > top*(1+1e-9) {
		t.Errorf("save holds %v above the cap %v", got, top)
	}
}

// Work past a completed tech's price is a head start too, and it fits the cap
// of what can be started then.
func TestOverflowOfACompletionFitsTheCap(t *testing.T) {
	ge := scholarEngine(t)
	if err := ge.StartResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.grantResearch(ge.Research.defs["tool_making"].Cost * 1e6)
	held := ge.Research.head
	ge.mu.Unlock()
	top := 0.0
	for _, ts := range ge.GetState().Research.Techs {
		if ts.Available {
			top = math.Max(top, ts.Cost)
		}
	}
	if held > top*(1+1e-9) {
		t.Errorf("after a completion the head start is %v, above the dearest startable tech %v", held, top)
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
