package game

import (
	"slices"
	"testing"
	"time"
)

// The History panel draws each sample at its index and each age marker at its
// tick (ui/overlay_history.go, ui/braille.go). The two agree only while the
// samples are evenly spaced in ticks. The helpers below mirror those two
// mappings: the ui package cannot be imported from here, since it imports game.

// historyTickColumn is the panel's column for a tick (ui.AgeMarkerCols).
func historyTickColumn(tick, first, last, width int) int {
	return (tick - first) * (width - 1) / (last - first)
}

// historyIndexColumn is the panel's column for sample i of n (ui.BrailleLine),
// which spreads the samples over twice the width in braille dots.
func historyIndexColumn(i, n, width int) int {
	return i * (2*width - 1) / (n - 1) / 2
}

// requireNewRunHistory checks that the history holds only the run ge is now
// playing: nothing left over from the run before, the samples climbing from
// the first one played, and an age marker taken in this run still in place.
// The caller has just started the new run.
func requireNewRunHistory(t *testing.T, ge *GameEngine) {
	t.Helper()
	if n, m := len(ge.History.Samples), len(ge.History.AgeMarkers); n != 0 || m != 0 {
		t.Fatalf("the new run starts with %d samples and %d age markers, want none", n, m)
	}
	ge.StepTicks(100)
	ge.advanceAge("stone_age")
	ge.StepTicks(50)

	h := ge.History
	if len(h.Samples) == 0 {
		t.Fatal("the new run has no samples")
	}
	for i := 1; i < len(h.Samples); i++ {
		if h.Samples[i].Tick <= h.Samples[i-1].Tick {
			t.Fatalf("samples out of order: tick %d after tick %d", h.Samples[i].Tick, h.Samples[i-1].Tick)
		}
	}
	first, last := h.Samples[0].Tick, h.Samples[len(h.Samples)-1].Tick
	if first != historySampleInterval || last != ge.tick {
		t.Errorf("the header reads ticks %d - %d, want %d - %d (the new run)", first, last, historySampleInterval, ge.tick)
	}
	want := AgeMarker{Tick: 100, AgeName: "stone_age"}
	if !slices.Contains(h.AgeMarkers, want) {
		t.Errorf("the new run's age marker is gone: markers %+v, want %+v among them", h.AgeMarkers, want)
	}
}

// A run's history is its own. A prestige, a fall and a new game each start the
// history over: the old run's samples and age markers are gone, and the new
// run's age marker is kept.
func TestHistoryStartsOverWithEachRun(t *testing.T) {
	t.Run("prestige", func(t *testing.T) {
		ge := catEngine(t, "medieval_age", 11)
		ge.StepTicks(400)
		ge.advanceAge("modern_age")
		ge.StepTicks(200)
		if err := ge.DoPrestige(); err != nil {
			t.Fatalf("DoPrestige: %v", err)
		}
		requireNewRunHistory(t, ge)
	})
	t.Run("succumb", func(t *testing.T) {
		ge := catEngine(t, "bronze_age", 5)
		ge.StepTicks(400)
		ge.advanceAge("iron_age")
		ge.StepTicks(200)
		succumbIn(t, ge, "iron_age")
		requireNewRunHistory(t, ge)
	})
	t.Run("new game", func(t *testing.T) {
		ge := newSeededEngine(3)
		ge.StepTicks(400)
		ge.advanceAge("stone_age")
		ge.StepTicks(200)
		ge.Reset()
		requireNewRunHistory(t, ge)
	})
}

// Offline catch-up passes ticks with no live tick, and it used to take no
// samples for them: the samples bunched up and an age marker placed after the
// gap sat in the wrong column. Here the age advances inside the time away (the
// plan's advance item goes as soon as its gate is met), and the samples and the
// marker must still agree.
func TestHistoryMarkersStayInColumnAfterTimeAway(t *testing.T) {
	ge := newSeededEngine(1)
	ge.StepTicks(1500)
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	for k, n := range ge.progress.GetRequirementsForNext(ge.age) {
		ge.Buildings.counts[k] = n
	}
	ge.Buildings.counts[ge.progress.WonderForAge(ge.age)] = 1

	ge.SimulateOffline(time.Duration(1500) * ge.getTickInterval())
	if ge.age != "stone_age" {
		t.Fatalf("the plan did not advance during the time away: age is %s", ge.age)
	}
	ge.StepTicks(300)

	h := ge.History
	n := len(h.Samples)
	for i := 1; i < n; i++ {
		if d := h.Samples[i].Tick - h.Samples[i-1].Tick; d != historySampleInterval {
			t.Fatalf("samples at ticks %d and %d are %d apart, want %d: the time away left a gap",
				h.Samples[i-1].Tick, h.Samples[i].Tick, d, historySampleInterval)
		}
	}
	const width = 60
	first, last := h.Samples[0].Tick, h.Samples[n-1].Tick
	found := false
	for _, m := range h.AgeMarkers {
		if m.AgeName != "stone_age" {
			continue
		}
		found = true
		if m.Tick <= 1500 || m.Tick > 3000 {
			t.Fatalf("the age advance is at tick %d, want inside the time away (1500, 3000]", m.Tick)
		}
		at := -1
		for i, s := range h.Samples {
			if s.Tick == m.Tick-m.Tick%historySampleInterval {
				at = i
			}
		}
		if at < 0 {
			t.Fatalf("no sample at the marker's tick %d (window %d - %d)", m.Tick, first, last)
		}
		col := historyTickColumn(m.Tick, first, last, width)
		want := historyIndexColumn(at, n, width)
		if d := col - want; d < -1 || d > 1 {
			t.Errorf("age marker at tick %d draws in column %d, but its sample (index %d) draws in column %d",
				m.Tick, col, at, want)
		}
	}
	if !found {
		t.Fatalf("no stone age marker in %+v", h.AgeMarkers)
	}
}
