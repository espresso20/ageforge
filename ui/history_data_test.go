package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
)

// recorded is a recorder holding n samples at ticks 10, 20, ... of a
// population that equals the tick, with an age mark at tick 400 and a wonder
// mark at 600.
func recorded(n int) *game.RunHistory {
	h := game.NewRunHistory()
	for i := 1; i <= n; i++ {
		h.Record(10*i, map[string]float64{"population": float64(10 * i), "food_rate": 1})
	}
	h.Mark(400, game.MarkAge, "bronze_age", 0)
	h.Mark(600, game.MarkWonder, "pyramids", 0)
	return h
}

func TestHistoryRangesCutTheRecorder(t *testing.T) {
	src := histSource{hist: recorded(100), firstAge: "stone_age"}

	age := src.live(rangeAge, "population")
	if len(age.T) == 0 || age.T[0] < 400 || age.T[len(age.T)-1] != 1000 {
		t.Fatalf("This age should start at the last age mark: ticks %v", age.T)
	}
	run := src.live(rangeRun, "population")
	if len(run.T) != 100 || run.T[0] != 10 {
		t.Fatalf("This run should be the whole recorder: %d samples from %v", len(run.T), run.T[:1])
	}
	if got := src.live(rangeAll, "population"); len(got.T) != len(run.T) {
		t.Errorf("All runs draws the run in play whole: %d samples, want %d", len(got.T), len(run.T))
	}

	if m := src.marks(rangeAge); len(m) != 2 || m[0].Kind != game.MarkAge {
		t.Errorf("This age keeps the marks from its own start on: %v", m)
	}
	hist := recorded(100)
	hist.Mark(200, game.MarkWonder, "early", 0)
	if m := (histSource{hist: hist}).marks(rangeAge); len(m) != 2 {
		t.Errorf("a mark before the age began is not in This age: %v", m)
	}
	if m := src.marks(rangeRun); len(m) != 2 {
		t.Errorf("This run keeps every mark: %v", m)
	}
}

func TestHistorySpans(t *testing.T) {
	src := histSource{hist: recorded(100), firstAge: "stone_age"}
	got := src.spans(rangeRun)
	want := []histSpan{{"stone_age", 0, 400}, {"bronze_age", 400, 1000}}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("run spans = %v, want %v", got, want)
	}
	got = src.spans(rangeAge)
	if len(got) != 1 || got[0] != (histSpan{"bronze_age", 400, 1000}) {
		t.Errorf("age span = %v", got)
	}
}

func TestHistoryPastLinesUpFromItsOwnStart(t *testing.T) {
	old := recorded(50)
	rec := game.RunRecord{ID: "a", History: &game.RunHistory{Run: old.Run, Marks: old.Marks}}
	s := past(rec, "population")
	if len(s.T) != 50 || s.T[0] != 10 || s.T[49] != 500 {
		t.Errorf("a finished run reads from its own tick 0: %v", s.T[:2])
	}
	if got := past(rec, ""); len(got.T) != 0 {
		t.Errorf("a measure finished runs do not keep reads empty: %v", got)
	}
	if got := past(game.RunRecord{ID: "b"}, "population"); len(got.T) != 0 || kept(game.RunRecord{ID: "b"}) {
		t.Errorf("a run with no kept history reads empty")
	}
	src := histSource{hist: old, firstAge: "stone_age"}
	if sp := src.pastSpans(rec); len(sp) != 2 || sp[1].End != 500 {
		t.Errorf("past spans = %v", sp)
	}
}

func TestHistoryEmptyAndSingleSample(t *testing.T) {
	for _, h := range []*game.RunHistory{nil, game.NewRunHistory(), recorded(1)} {
		src := histSource{hist: h, firstAge: "stone_age"}
		for r := rangeAge; r < histRanges; r++ {
			src.live(r, "population")
			src.marks(r)
			src.spans(r)
			src.window(r)
		}
	}
}

func TestResourceRatesSkipFoodAndSort(t *testing.T) {
	h := game.NewRunHistory()
	h.Record(10, map[string]float64{"population": 1, "rate.wood": 1, "rate.food": 1, "rate.coal": 2})
	got := resourceRates(h)
	if len(got) != 2 || got[0] != "rate.coal" || got[1] != "rate.wood" {
		t.Errorf("resourceRates = %v", got)
	}
}
