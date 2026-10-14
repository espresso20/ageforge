package game

import (
	"math"
	"strings"
	"testing"
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
