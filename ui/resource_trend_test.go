package ui

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

func trendRows(hist *resHistory) []resRow {
	st := game.GameState{Resources: map[string]game.ResourceState{
		"food":  {Name: "Food", Amount: 340000, Rate: 13.1, Unlocked: true},
		"wood":  {Name: "Wood", Amount: 3400, Rate: 4.2, Unlocked: true},
		"stone": {Name: "Stone", Amount: 0, Rate: 0, Unlocked: true},
	}}
	return resourceRowsFor(st, hist)
}

func rowNamed(rows []resRow, name string) resRow {
	for _, r := range rows {
		if r.name == name {
			return r
		}
	}
	return resRow{}
}

func count(s string, set string) int {
	n := 0
	for _, r := range s {
		if strings.ContainsRune(set, r) {
			n++
		}
	}
	return n
}

func TestMagnitudeBarPresentAndScaled(t *testing.T) {
	rows := trendRows(nil)
	f := fitResourceFormat(rows, 60)
	if f.bar != 35 {
		t.Fatalf("bar is %d cells, want %d", f.bar, 35)
	}
	lit := func(r resRow) float64 {
		bar := untag(magnitudeCells(r.fill, 12))
		whole := float64(count(bar, "█"))
		for i, p := range "▏▎▍▌▋▊▉" {
			whole += float64(count(bar, string(p))) * float64(i+1) / 8
		}
		return whole
	}
	food, wood, stone := lit(rowNamed(rows, "Food")), lit(rowNamed(rows, "Wood")), lit(rowNamed(rows, "Stone"))
	if food < 10.9 || food > 12 {
		t.Errorf("the largest amount fills %.2f cells, want about 11 to 12", food)
	}
	// 100 times smaller is two powers of ten down, each 12/log10(top) cells.
	want := 2 * 12 / math.Log10(1e6+1)
	if got := food - wood; math.Abs(got-want) > 0.3 {
		t.Errorf("100 times smaller is %.2f cells shorter, want about %.2f", got, want)
	}
	if stone != 0 || strings.ContainsAny(untag(magnitudeCells(0, 12)), "█▏▎▍▌▋▊▉") {
		t.Errorf("zero should show an empty bar, got %.2f cells", stone)
	}
	if line := untag(f.line(rowNamed(rows, "Food"))); !strings.ContainsAny(line, "█") || runeLen(line) > 60 {
		t.Errorf("row has no bar or is too wide: %q", line)
	}
}

func TestMagnitudeTop(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{0, 10}, {9, 10}, {10, 10}, {11, 100}, {340000, 1e6}, {1e6, 1e6}} {
		if got := magnitudeTop([]float64{c.in, 1}); got != c.want {
			t.Errorf("top of %v is %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNarrowDropsTrendThenBar(t *testing.T) {
	rows := trendRows(&resHistory{})
	for i := range rows {
		rows[i].trend = []float64{1, 2, 3}
	}
	wide := fitResourceFormat(rows, 60)
	if wide.trend == 0 || wide.bar == 0 {
		t.Fatalf("60 cells: trend %d bar %d", wide.trend, wide.bar)
	}
	sawNoTrend, sawNoBar := false, false
	for w := 60; w >= 20; w-- {
		f := fitResourceFormat(rows, w)
		if f.trend > 0 && f.bar == 0 {
			t.Errorf("width %d: a trend line without a bar", w)
		}
		sawNoTrend = sawNoTrend || (f.trend == 0 && f.bar > 0)
		sawNoBar = sawNoBar || f.bar == 0
		if sawNoBar && f.trend > 0 {
			t.Errorf("width %d: the trend came back after the bar went", w)
		}
		for _, r := range rows {
			line := untag(f.line(r))
			rate, _ := f.rateText(r)
			if runeLen(line) > w && w >= 24 {
				t.Errorf("width %d: row too wide: %q", w, line)
			}
			if !strings.Contains(line, rate) || !strings.Contains(line, r.amt) {
				t.Errorf("width %d: amount or rate cut: %q", w, line)
			}
		}
	}
	if !sawNoTrend || !sawNoBar {
		t.Errorf("never saw trend dropped alone (%v) and then bar dropped (%v)", sawNoTrend, sawNoBar)
	}
}

func TestRisingHistoryDrawsRisingLine(t *testing.T) {
	line := untag(trendCells([]float64{1, 2, 3, 4, 5, 6, 7, 8}, "green", resTrendCells))
	if line != "▁▂▃▄▅▆▇█" {
		t.Errorf("rising line is %q", line)
	}
	if got := untag(trendCells([]float64{8, 4, 1}, "red", resTrendCells)); got != "     █▄▁" {
		t.Errorf("falling line is %q", got)
	}
	if got := untag(trendCells([]float64{5}, "gray", resTrendCells)); strings.TrimSpace(got) != "" {
		t.Errorf("one point should be blank, got %q", got)
	}
}

func TestHistorySamplesEveryMinute(t *testing.T) {
	h := &resHistory{}
	st := func(tick int, a float64) game.GameState {
		return game.GameState{Tick: tick, Resources: map[string]game.ResourceState{"food": {Amount: a, Unlocked: true}}}
	}
	t0 := time.Unix(1000, 0)
	h.note(st(1, 1), t0)
	h.note(st(2, 2), t0.Add(30*time.Second))
	h.note(st(3, 3), t0.Add(61*time.Second))
	if got := h.points("food", 9); len(got) != 3 || got[0] != 1 || got[1] != 3 || got[2] != 9 {
		t.Errorf("points %v", got)
	}
	h.note(st(0, 5), t0.Add(62*time.Second)) // the tick went back: another game
	if got := h.points("food", 9); len(got) != 2 || got[0] != 5 {
		t.Errorf("a different game kept its history: %v", got)
	}
}
