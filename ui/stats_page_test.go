package ui

import (
	"image"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/window/kit"
)

// gridLines is a grid's rows as text, trailing blanks trimmed.
func gridLines(g *xgrid) []string {
	lines := make([]string, g.h)
	for y := range lines {
		row := make([]rune, g.w)
		for x := range row {
			row[x] = g.c[y*g.w+x].r
			if row[x] == 0 {
				row[x] = ' '
			}
		}
		lines[y] = strings.TrimRight(string(row), " ")
	}
	return lines
}

// statsRowsText is a table's rows as text, a row to a line.
func statsRowsText(rows []statsRow) string {
	var sb strings.Builder
	for _, r := range rows {
		sb.WriteString(strings.TrimSpace(r.label + " " + r.value))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// statsPageText is the text of a part of the page in cells only, w wide.
func statsPageText(st game.GameState, part, w int) string {
	return strings.Join(gridLines(buildStats(st, part, w, 0, w >= statsWide, false, 0).g), "\n")
}

// statsFixture is a run of nAges ages, nRes gathered resources and the given
// play time, at one tick a second.
func statsFixture(t *testing.T, nAges, nRes int, play time.Duration) game.GameState {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	a, err := game.CreateAccount("Fixture")
	must(t, err)
	eng.SetAccount(a)
	st := eng.GetState()
	st.TickIntervalMs = 1000
	ages := config.Ages()
	st.Stats.AgeLog = nil
	tick := 0
	for i := 0; i < nAges; i++ {
		st.Stats.AgeLog = append(st.Stats.AgeLog, game.AgeEntry{Age: ages[i].Key, Tick: tick})
		tick += 200 + i*i*40 + (i%3)*500
	}
	st.Tick = tick + 300
	st.Stats.PlayTime = play
	st.Stats.TotalBuilt, st.Stats.TotalRecruited = 12*nAges, 7*nAges
	st.Research.TotalResearched, st.Military.CompletedCount = 9*nAges, nAges/2
	var keys []string
	for k := range st.Resources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	st.Stats.TotalGathered = map[string]float64{}
	for i := 0; i < nRes && i < len(keys); i++ {
		st.Stats.TotalGathered[keys[i]] = math.Pow(10, float64(2+i%9)+0.37*float64(i%4))
		rs := st.Resources[keys[i]]
		rs.Rate = 0.4 * float64(i+1)
		if i%7 == 3 {
			rs.Rate = -1.5
		}
		st.Resources[keys[i]] = rs
	}
	return st
}

// laidStats is a panel opened on st and laid out w by h, with pictures on
// when pics is set.
func laidStats(t *testing.T, st game.GameState, w, h int, pics bool) *statsPanel {
	t.Helper()
	if pics {
		pictureCell = func() (int, int) { return 12, 24 }
		t.Cleanup(func() { pictureCell = nil })
	} else {
		pictureCell = nil
	}
	p := newStatsPanel()
	p.open(st)
	p.Layout(image.Rect(0, 0, w, h))
	return p
}

func keyOf(p *statsPanel, c kit.KeyCode) { p.routeKey(&kit.Key{Code: c}, "") }

func TestStatsTabsSwitchByKeyAndClick(t *testing.T) {
	p := laidStats(t, statsFixture(t, 3, 4, time.Hour), 160, 40, true)
	keyOf(p, kit.KeyTab)
	if p.part != partLifetime {
		t.Fatalf("Tab: part %d", p.part)
	}
	keyOf(p, kit.KeyTab)
	if p.part != partRun {
		t.Fatalf("Tab again: part %d", p.part)
	}
	click := func(x int) {
		p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: x, Y: 1})
	}
	click(14)
	if p.part != partLifetime {
		t.Fatalf("click on the second tab: part %d", p.part)
	}
	click(4)
	if p.part != partRun {
		t.Fatalf("click on the first tab: part %d", p.part)
	}
	foot := strings.Join(gridLines(p.chrome), "\n")
	for _, k := range []string{"Up/Dn", "PgUp/PgDn", "Tab", "Esc"} {
		if !strings.Contains(foot, k) {
			t.Errorf("footer lacks %q", k)
		}
	}
}

func TestStatsScrollStopsAtBothEnds(t *testing.T) {
	p := laidStats(t, statsFixture(t, 22, 24, 100*time.Hour), 100, 30, true)
	top := p.page.used - p.viewRows()
	if top <= p.viewRows() {
		t.Fatalf("fixture does not scroll: used %d view %d", p.page.used, p.viewRows())
	}
	keyOf(p, kit.KeyUp)
	if p.scroll != 0 {
		t.Fatalf("scrolled above the top: %d", p.scroll)
	}
	keyOf(p, kit.KeyDown)
	keyOf(p, kit.KeyPgDn)
	if p.scroll != 1+p.viewRows() {
		t.Fatalf("PgDn moved to %d, want %d", p.scroll, 1+p.viewRows())
	}
	keyOf(p, kit.KeyPgUp)
	if p.scroll != 1 {
		t.Fatalf("PgUp moved to %d", p.scroll)
	}
	p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: 1000})
	if p.scroll != top {
		t.Fatalf("scroll past the end: %d, last %d", p.scroll, top)
	}
	keyOf(p, kit.KeyDown)
	if p.scroll != top {
		t.Fatalf("Down at the end moved to %d", p.scroll)
	}
}

func TestStatsPictureCountsAndPlainCells(t *testing.T) {
	st := statsFixture(t, 6, 12, 5*time.Hour)
	p := laidStats(t, st, 160, 40, true)
	if got, want := len(p.page.pics), 6+6+12; got != want {
		t.Fatalf("pictures: %d, want %d (tiles, blocks, bars)", got, want)
	}
	// the frame's own pictures are the same on both parts, so the parts differ
	// by exactly the page's pictures, and the plate of the Prestige button
	// that only the first part holds
	run := len(paintAlone(overBar(p), 160, 40).Pics)
	keyOf(p, kit.KeyTab)
	life := len(paintAlone(overBar(p), 160, 40).Pics)
	if got, want := run-life, 24-len(p.page.pics)+1; got != want {
		t.Errorf("painted pictures differ by %d, want %d", got, want)
	}
	pl := laidStats(t, st, 160, 40, false)
	if len(pl.page.pics) != 0 {
		t.Errorf("pictures off: %d pictures", len(pl.page.pics))
	}
	for _, part := range []int{partRun, partLifetime} {
		pl.part = part
		pl.Layout(pl.Bounds())
		b := paintAlone(overBar(pl), 160, 40)
		if len(b.Pics) != 0 {
			t.Errorf("part %d: pictures in the plain tier", part)
		}
		for i, c := range b.Cells {
			if c.R > 127 {
				t.Fatalf("part %d: cell %d,%d holds %q", part, i%b.W, i/b.W, c.R)
			}
		}
	}
}

func TestStatsNothingOverlapsAndAgeNamesAreWhole(t *testing.T) {
	st := statsFixture(t, 22, 24, 100*time.Hour)
	for _, w := range []int{100, 160} {
		p := laidStats(t, st, w, 40, true)
		var rects []image.Rectangle
		for _, tl := range p.page.tiles {
			rects = append(rects, tl.r)
		}
		for _, b := range p.page.blocks {
			rects = append(rects, b.r)
		}
		for _, b := range p.page.bars {
			rects = append(rects, b.track)
		}
		for i := range rects {
			if rects[i].Max.X > w-4 {
				t.Errorf("w %d: %v leaves the page", w, rects[i])
			}
			for j := i + 1; j < len(rects); j++ {
				if rects[i].Overlaps(rects[j]) {
					t.Errorf("w %d: %v overlaps %v", w, rects[i], rects[j])
				}
			}
		}
		for _, b := range p.page.blocks {
			whole := false
			for _, l := range p.page.labels {
				if !image.Pt(l.x, l.y).In(b.r) {
					continue
				}
				if l.x+runeLen(l.s) > b.r.Max.X {
					t.Errorf("w %d: %q runs out of its block", w, l.s)
				}
				if strings.HasPrefix(b.name, l.s) && l.s != b.name && l.y == b.r.Min.Y+(b.r.Dy()-2)/2 {
					t.Errorf("w %d: %q is cut from %q", w, l.s, b.name)
				}
				whole = whole || l.s == b.name
			}
			if whole != b.named {
				t.Errorf("w %d: %q painted %v, named %v", w, b.name, whole, b.named)
			}
		}
	}
}

func TestStatsEarlyPageFillsTheView(t *testing.T) {
	p := laidStats(t, statsFixture(t, 1, 3, 7*time.Minute), 160, 40, true)
	if p.page.used*4 < p.viewRows()*3 {
		t.Fatalf("early page takes %d of %d rows", p.page.used, p.viewRows())
	}
}

func TestStatsRatesAndBars(t *testing.T) {
	st := statsFixture(t, 6, 12, 5*time.Hour)
	p := laidStats(t, st, 160, 40, true)
	want := map[string]int{}
	for k, v := range st.Stats.TotalGathered {
		if v <= 0 {
			continue
		}
		rv, unit := ratePer(st.Resources[k].Rate, st.TickIntervalMs)
		s := textfmtRate(rv, unit)
		want[s]++
	}
	got := map[string]int{}
	bars := p.page.bars
	for i, b := range bars {
		got[b.rate]++
		if i > 0 && (b.value > bars[i-1].value || b.frac > bars[i-1].frac) {
			t.Errorf("bar %q is longer than the larger total before it", b.name)
		}
	}
	for s, n := range want {
		if got[s] != n {
			t.Errorf("rate %q: shown %d times, want %d", s, got[s], n)
		}
	}
	// a decade apart: one tick's width of the track
	ds := statsFixture(t, 1, 0, time.Minute)
	ds.Stats.TotalGathered = map[string]float64{"food": 100, "wood": 1e3, "stone": 1e4, "gold": 1e5}
	q := laidStats(t, ds, 160, 40, true)
	if len(q.page.bars) != 4 || len(q.page.ticks) < 2 {
		t.Fatalf("bars %d ticks %d", len(q.page.bars), len(q.page.ticks))
	}
	step := q.page.ticks[1] - q.page.ticks[0]
	for i := 1; i < 4; i++ {
		if d := q.page.bars[i-1].frac - q.page.bars[i].frac; math.Abs(d-step) > 1e-9 {
			t.Errorf("bars %d and %d are %.4f apart, a decade is %.4f", i-1, i, d, step)
		}
	}
}

func textfmtRate(rv float64, unit string) string {
	s := textfmt.Number(math.Abs(rv)) + unit
	switch {
	case rv < 0:
		return "-" + s
	case rv > 0:
		return "+" + s
	}
	return s
}

func TestStatsHoverReadouts(t *testing.T) {
	p := laidStats(t, statsFixture(t, 6, 12, 5*time.Hour), 160, 40, true)
	b, bar := p.page.blocks[1], p.page.bars[2]
	at := func(r image.Rectangle) { p.hover = image.Pt(r.Min.X, r.Min.Y) }
	at(b.r)
	if got, _ := p.readout(); got != b.readout || !strings.Contains(got, b.name) || !strings.Contains(got, b.time) {
		t.Errorf("block readout %q", got)
	}
	at(bar.track)
	if got, _ := p.readout(); got != bar.readout || !strings.Contains(got, commas(bar.value)) {
		t.Errorf("bar readout %q", got)
	}
	p.hover = image.Pt(-1, -1)
	if got, _ := p.readout(); got != "" {
		t.Errorf("readout off the page: %q", got)
	}
}

// screenRow is row y of a painted buffer as text.
func screenRow(b *kit.Buffer, y int) string {
	var sb strings.Builder
	for x := 0; x < b.W; x++ {
		sb.WriteRune(max(' ', b.Cells[y*b.W+x].R))
	}
	return sb.String()
}

func TestStatsReadoutSitsOnItsHeadingRow(t *testing.T) {
	p := laidStats(t, statsFixture(t, 6, 12, 5*time.Hour), 160, 40, true)
	b, bar := p.page.blocks[1], p.page.bars[2]
	for _, c := range []struct {
		at   image.Rectangle
		want string
		head image.Rectangle
	}{{b.r, b.readout, p.page.heads[0]}, {bar.track, bar.readout, p.page.heads[1]}} {
		p.hover = c.at.Min
		buf := paintAlone(overBar(p), 160, 40)
		row := screenRow(buf, 3+c.head.Min.Y)
		if !strings.Contains(row, c.want) || !strings.HasSuffix(strings.TrimRight(string([]rune(row)[:2+c.head.Max.X]), " "), c.want) {
			t.Errorf("heading row %q lacks %q at its right end", row, c.want)
		}
		if strings.Contains(screenRow(buf, 2), c.want) {
			t.Errorf("readout %q is still above the tiles", c.want)
		}
	}
}

func TestStatsBarsTakeTheHeightTheyHave(t *testing.T) {
	for n, pitch := range map[int]int{3: 3, 8: 2, 20: 1} {
		p := laidStats(t, statsFixture(t, 6, n, 5*time.Hour), 160, 40, true)
		bars := p.page.bars
		if len(bars) != n {
			t.Fatalf("%d bars, want %d", len(bars), n)
		}
		if got := bars[1].track.Min.Y - bars[0].track.Min.Y; got != pitch {
			t.Errorf("%d bars: pitch %d, want %d", n, got, pitch)
		}
		if got := bars[0].track.Dy(); got != max(1, pitch-1) {
			t.Errorf("%d bars: %d rows thick at pitch %d", n, got, pitch)
		}
	}
}

func TestStatsSmallSectionsGoUnderTheShorterColumn(t *testing.T) {
	col := func(st game.GameState, head string) int {
		p := laidStats(t, st, 160, 40, true)
		for _, l := range gridLines(p.page.g) {
			if x := strings.Index(l, head); x >= 0 {
				return x
			}
		}
		t.Fatalf("no heading %q", head)
		return 0
	}
	for _, head := range []string{txtMultsHd.String(), txtResearchHd.String()} {
		if x := col(statsFixture(t, 1, 3, time.Minute), head); x != 0 {
			t.Errorf("%q with three bars is at column %d, want the left one", head, x)
		}
		if x := col(statsFixture(t, 22, 24, 100*time.Hour), head); x == 0 {
			t.Errorf("%q with 24 bars is in the left column, want the right one", head)
		}
	}
	if x := col(statsFixture(t, 1, 3, time.Minute), txtPrestigeHd.String()); x == 0 {
		t.Error("Prestige and epoch opens the left column, want the right one")
	}
	for _, c := range []struct{ ages, res int }{{1, 3}, {6, 12}} {
		p := laidStats(t, statsFixture(t, c.ages, c.res, time.Hour), 160, 40, true)
		if d := p.page.ends[0] - p.page.ends[1]; d > 6 || d < -6 {
			t.Errorf("%d ages, %d resources: columns end at %v", c.ages, c.res, p.page.ends)
		}
	}
}

func TestStatsTilesShareOneScaleAndClearTheirSigns(t *testing.T) {
	in := newStatsInk(newExpPalette())
	for _, w := range []int{100, 120, 160} {
		for _, st := range []game.GameState{statsFixture(t, 22, 24, 190*time.Hour), statsFixture(t, 1, 3, 5*time.Hour+12*time.Minute)} {
			p := laidStats(t, st, w, 40, true)
			var heights []int
			for _, pc := range p.page.pics {
				if !strings.HasPrefix(pc.key, "tile") {
					continue
				}
				img := pc.draw(12, 24, false, in)
				box := image.Rectangle{}
				for y := 0; y < img.Rect.Dy(); y++ {
					for x := 0; x < img.Rect.Dx(); x++ {
						if img.RGBAAt(x, y) == in.face {
							box = box.Union(image.Rect(x, y, x+1, y+1))
						}
					}
				}
				heights = append(heights, box.Dy())
				if k := max(1, 24/21); box.Min.X-40*k < 12 {
					t.Errorf("w %d: numeral starts %d px after the sign, want a cell", w, box.Min.X-40*k)
				}
			}
			for _, h := range heights {
				if h != heights[0] {
					t.Errorf("w %d: numeral heights differ: %v", w, heights)
					break
				}
			}
		}
	}
}

func TestStatsDropsTheTicksCounter(t *testing.T) {
	st := statsFixture(t, 6, 12, 5*time.Hour)
	for _, pics := range []bool{true, false} {
		p := laidStats(t, st, 160, 40, pics)
		for part := range 2 {
			p.part = part
			p.Layout(p.Bounds())
			var all []string
			all = append(all, gridLines(p.page.g)...)
			for _, l := range p.page.labels {
				all = append(all, l.s)
			}
			if txt := strings.ToLower(strings.Join(all, "\n")); strings.Contains(txt, "tick") {
				t.Errorf("pics %v part %d: %q", pics, part, txt)
			}
		}
	}
}
