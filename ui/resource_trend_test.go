package ui

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

func trendRows(plain bool, hist *resHistory) []resRow {
	st := game.GameState{Resources: map[string]game.ResourceState{
		"food":  {Name: "Food", Amount: 340000, Rate: 13.1, Unlocked: true},
		"wood":  {Name: "Wood", Amount: 3400, Rate: 4.2, Unlocked: true},
		"stone": {Name: "Stone", Amount: 0, Rate: 0, Unlocked: true},
	}}
	return resourceRowsFor(st, hist, plain)
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
	rows := trendRows(false, nil)
	f := fitResourceFormat(rows, 60)
	if f.bar != resBarMax {
		t.Fatalf("bar is %d cells, want %d", f.bar, resBarMax)
	}
	lit := func(r resRow) float64 {
		bar := untag(magnitudeCells(r.fill, 12, false))
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
	if stone != 0 || strings.ContainsAny(untag(magnitudeCells(0, 12, false)), "█▏▎▍▌▋▊▉") {
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

func TestASCIITierIsPlain(t *testing.T) {
	now := time.Unix(0, 0)
	h := &resHistory{}
	for i := 0; i < 4; i++ {
		h.note(game.GameState{Tick: i, Resources: map[string]game.ResourceState{"food": {Amount: float64(i * 10), Unlocked: true}}}, now.Add(time.Duration(i)*time.Minute))
	}
	rows := trendRows(true, h)
	f := fitResourceFormat(rows, 60)
	if f.trend == 0 {
		t.Fatal("no trend line at 60 cells")
	}
	r := rowNamed(rows, "Food")
	r.trend = []float64{1, 2, 3, 4, 5, 6, 7, 8}
	line := untag(f.line(r))
	tail := line[strings.Index(line, "/tick")+len("/tick"):]
	for _, c := range tail {
		if c > 0x7e {
			t.Fatalf("ASCII tier drew %q in %q", c, tail)
		}
	}
	if !strings.Contains(tail, "#") || !strings.Contains(tail, ".") || !strings.ContainsAny(tail, "_.-~^") {
		t.Errorf("want # . bar and _.-~^ trend, got %q", tail)
	}
}

func TestNarrowDropsTrendThenBar(t *testing.T) {
	rows := trendRows(false, &resHistory{})
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
			rate, _ := rateParts(r.rate, f.unit)
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
	line := untag(trendCells([]float64{1, 2, 3, 4, 5, 6, 7, 8}, false, "green"))
	if line != "▁▂▃▄▅▆▇█" {
		t.Errorf("rising line is %q", line)
	}
	if got := untag(trendCells([]float64{8, 4, 1}, false, "red")); got != "     █▄▁" {
		t.Errorf("falling line is %q", got)
	}
	if got := untag(trendCells([]float64{3, 3, 3}, true, "gray")); got != "     ---" {
		t.Errorf("flat ASCII line is %q", got)
	}
	if got := untag(trendCells([]float64{5}, false, "gray")); strings.TrimSpace(got) != "" {
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

// TestRenderResourceBars writes the box at a normal width in both tiers for a look.
func TestRenderResourceBars(t *testing.T) {
	dir := os.Getenv("RESOURCE_BARS_DIR")
	if dir == "" {
		t.Skip("set RESOURCE_BARS_DIR to write the renders")
	}
	for _, plain := range []bool{false, true} {
		h := &resHistory{}
		now := time.Unix(0, 0)
		for i := 0; i < 8; i++ {
			st := game.GameState{Tick: i, Resources: map[string]game.ResourceState{
				"food": {Amount: 340000 * float64(i+1) / 8, Unlocked: true},
				"wood": {Amount: 3400 + float64(i%3)*700, Unlocked: true},
			}}
			h.note(st, now.Add(time.Duration(i)*time.Minute))
		}
		rows := trendRows(plain, h)
		lines, _ := layoutResourceBox(rows, 56, 10, 0)
		name := "resource-bars-unicode.txt"
		if plain {
			name = "resource-bars-ascii.txt"
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(untag(strings.Join(lines, "\n"))+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestRenderResourceBarsOnScreen draws the real dashboard on a simulation
// screen and writes the Resources box as text, in both glyph tiers.
func TestRenderResourceBarsOnScreen(t *testing.T) {
	dir := os.Getenv("RESOURCE_BARS_DIR")
	if dir == "" {
		t.Skip("set RESOURCE_BARS_DIR to write the renders")
	}
	restoreForge(t)
	d, pages, eng := stagedDashboard(t, "medieval_age")
	st := eng.GetState()
	base := time.Unix(0, 0)
	for i := 0; i < 8; i++ {
		s := st
		s.Resources = map[string]game.ResourceState{}
		for k, rs := range st.Resources {
			rs.Amount = rs.Amount * (0.5 + float64(i)/14)
			s.Resources[k] = rs
		}
		s.Tick = st.Tick
		d.economyTab.trend.note(s, base.Add(time.Duration(i)*time.Minute))
	}
	d.economyTab.now = func() time.Time { return base.Add(7*time.Minute + time.Second) }
	for _, c := range []struct {
		plain bool
		w, h  int
		name  string
	}{
		{false, 120, 40, "resource-bars-screen-unicode.txt"}, {true, 120, 40, "resource-bars-screen-ascii.txt"},
		{false, 160, 45, "resource-bars-screen-unicode-160.txt"}, {true, 160, 45, "resource-bars-screen-ascii-160.txt"},
	} {
		plain, name := c.plain, c.name
		d.economyTab.plain = func() bool { return plain }
		cells := drawnDashboard(t, d, pages, c.w, c.h)
		rows := checkFitted(t, "render", d.economyTab.resourceTV, cells, c.w)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(strings.Join(rows, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
