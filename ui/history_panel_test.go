package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

func histKey(p *historyPanel, r rune) { p.routeKey(&kit.Key{Code: kit.KeyRune, Rune: r}, "") }

func TestHistoryRangesSpanSamples(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(600, 6, 1), nil)
	p.setRange(rangeRun)
	run := p.v
	p.setRange(rangeAge)
	age := p.v
	if len(age.ticks) == 0 || len(age.ticks) >= len(run.ticks) {
		t.Fatalf("age samples %d, run samples %d", len(age.ticks), len(run.ticks))
	}
	if age.lo <= run.lo || age.end != run.end {
		t.Fatalf("age span %d..%d, run span %d..%d", age.lo, age.end, run.lo, run.end)
	}
	if len(age.spans) != 1 || len(run.spans) < 2 {
		t.Fatalf("age spans %d, run spans %d", len(age.spans), len(run.spans))
	}
}

func TestHistoryLogDefaultOnAllRunsOnly(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), []game.RunRecord{fixtureRecord("a", game.RunPrestige, 300, 4, 1)})
	for r, want := range map[histRange]bool{rangeAge: false, rangeRun: false, rangeAll: true} {
		p.setRange(r)
		if p.v.logOn != want {
			t.Errorf("range %d: log %v, want %v", r, p.v.logOn, want)
		}
	}
	p.setRange(rangeAll)
	p.setLog(false)
	p.setRange(rangeRun)
	p.setRange(rangeAll)
	if p.v.logOn {
		t.Error("the choice for All runs was not kept")
	}
}

func TestHistorySlotsKeepColourAndSign(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), nil)
	histKey(p, '2')
	histKey(p, '3')
	if p.slots != [histSlots]string{"population", "morale", "faith"} {
		t.Fatalf("slots %v", p.slots)
	}
	histKey(p, '1')
	histKey(p, '4')
	if p.slots != [histSlots]string{"buildings", "morale", "faith"} {
		t.Fatalf("slots %v", p.slots)
	}
	for _, l := range p.v.lines {
		if l.m.Key == "buildings" && l.slot != 0 {
			t.Errorf("buildings took slot %d", l.slot)
		}
	}
}

func TestHistoryPointerKeyAndHoverAgree(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(40, 4, 1), nil)
	for range 5 {
		p.routeKey(&kit.Key{Code: kit.KeyLeft}, "")
	}
	byKey := p.v.ptr
	if byKey == p.v.ticks[len(p.v.ticks)-1] {
		t.Fatal("the pointer did not move")
	}
	g := p.v.geo
	x := p.v.cellX(byKey) + g.px0 + p.Bounds().Min.X
	p.routeKey(&kit.Key{Code: kit.KeyEnd}, "")
	p.Mouse(kit.Mouse{Kind: kit.MouseMove, X: x, Y: g.y0 + 3})
	if p.v.ptr != byKey {
		t.Fatalf("hover read %d, key read %d", p.v.ptr, byKey)
	}
}

func TestHistoryEdgeSizes(t *testing.T) {
	for _, n := range []int{0, 1, 5, 4000} {
		run := game.NewRunHistory()
		if n > 0 {
			run = fixtureRun(n, 6, 1)
		}
		for _, w := range []int{100, 120, 160} {
			p := histFixturePanel(w, 40, run, nil)
			for r := range histRange(histRanges) {
				p.setRange(r)
				histPaint(p)
				p.routeKey(&kit.Key{Code: kit.KeyLeft}, "")
				p.routeKey(&kit.Key{Code: kit.KeyHome}, "")
			}
		}
	}
	var _ = image.Pt
}

func histRunsFixture(n int) []game.RunRecord {
	var out []game.RunRecord
	for i := range n {
		out = append(out, fixtureRecord("r"+string(rune('a'+i)), game.RunPrestige, 100+20*i, 3+i%5, 1))
	}
	return out
}

func TestHistoryPickingTheGhost(t *testing.T) {
	runs := histRunsFixture(4)
	runs[1].History = nil
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), runs)
	p.setRange(rangeAll)
	p.routeKey(&kit.Key{Code: kit.KeyEnter}, "")
	if p.ghost != "ra" || p.v.ghost == nil {
		t.Fatalf("ghost %q after Enter on the first row", p.ghost)
	}
	p.routeKey(&kit.Key{Code: kit.KeyDown}, "")
	p.routeKey(&kit.Key{Code: kit.KeyEnter}, "")
	if p.ghost != "ra" {
		t.Fatalf("a run with no kept history was picked: %q", p.ghost)
	}
	p.routeKey(&kit.Key{Code: kit.KeyDown}, "")
	hit, _ := p.chrome.hitAt(hitHistRun, image.Pt(5, p.v.geo.listTop+2+3))
	if hit != 3 {
		t.Fatalf("row 3 hit %d", hit)
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 5, Y: p.v.geo.listTop + 2 + 3})
	if p.ghost != "rd" {
		t.Fatalf("click picked %q", p.ghost)
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: 5, Y: p.v.geo.listTop + 2 + 1})
	if p.ghost != "rd" {
		t.Fatalf("click on an unkept run changed the ghost to %q", p.ghost)
	}
	if got := p.v.lines[0].gh; len(got.V) == 0 {
		t.Fatal("the ghost has no line")
	}
}

func TestHistoryReadoutColumns(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), histRunsFixture(2))
	p.setRange(rangeAll)
	p.routeKey(&kit.Key{Code: kit.KeyEnter}, "")
	lines := histPaint(p).Lines()
	n := 0
	for _, l := range lines {
		if strings.Contains(l, txtKeyGhost.String()) && strings.Contains(l, "At ") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("readout head with a ghost column found %d times", n)
	}
}

func TestHistoryEmptyAccount(t *testing.T) {
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), nil)
	p.setRange(rangeAll)
	lines := histPaint(p).Lines()
	play, none := 0, 0
	for _, l := range lines {
		if strings.Contains(l, txtInPlay.String()) {
			play++
		}
		if strings.Contains(l, txtNoRuns.String()) {
			none++
		}
	}
	if play != 1 || none != 1 {
		t.Fatalf("in-play rows %d, placeholder lines %d", play, none)
	}
}

func TestHistoryCountsHiddenBelow120(t *testing.T) {
	for w, want := range map[int]bool{100: false, 119: false, 120: true, 160: true} {
		p := histFixturePanel(w, 40, fixtureRun(300, 4, 1), histRunsFixture(3))
		p.setRange(rangeAll)
		got := false
		for _, l := range histPaint(p).Lines() {
			if strings.Contains(l, txtExpeditions.String()) && strings.Contains(l, txtRunCol.String()) {
				got = true
			}
		}
		if got != want {
			t.Errorf("width %d: counts shown %v, want %v", w, got, want)
		}
	}
}

func TestHistoryNothingOutsideThePanel(t *testing.T) {
	for _, w := range []int{100, 120} {
		p := histFixturePanel(w, 40, fixtureRun(300, 4, 1), histRunsFixture(14))
		p.setRange(rangeAll)
		b := kit.NewBuffer(w+8, 40)
		pt := kit.NewPainter(b)
		p.Paint(pt)
		for y := range 40 {
			for x := w; x < w+8; x++ {
				if c := b.At(x, y); c.R != ' ' || c.Pic != 0 {
					t.Fatalf("width %d: a cell outside the panel at %d,%d", w, x, y)
				}
			}
		}
	}
}

// histCells gives the panel picture cells for the test.
func histCells(t *testing.T) {
	was := pictureCell
	pictureCell = func() (int, int) { return 12, 24 }
	t.Cleanup(func() { pictureCell = was })
}

func histRowText(b *kit.Buffer, y, w int) string {
	return bufText(b, image.Rect(0, y, w, y+1))
}

// A long name of the run in play is cut with an ellipsis and keeps its tag.
func TestHistoryRunNameCutWithEllipsis(t *testing.T) {
	histCells(t)
	p := histFixturePanel(160, 40, fixtureRun(300, 4, 1), nil)
	p.name = "Sanctuary of Frostvale"
	p.setRange(rangeAll)
	p.Layout(p.Bounds())
	row := histRowText(histPaint(p), p.v.geo.listTop+1, p.w)
	if !strings.Contains(row, "…") || !strings.Contains(row, txtInPlay.String()) {
		t.Errorf("row %q lacks the ellipsis or the tag", row)
	}
	p.name = "Hearth"
	p.Layout(p.Bounds())
	row = histRowText(histPaint(p), p.v.geo.listTop+1, p.w)
	if strings.Contains(row, "…") || !strings.Contains(row, "Hearth "+txtInPlay.String()) {
		t.Errorf("short name altered: %q", row)
	}
}

// At 100 columns the All runs key bar keeps every key, Enter included.
func TestHistoryKeyBarFitsAt100(t *testing.T) {
	histCells(t)
	p := histFixturePanel(100, 40, fixtureRun(300, 4, 1), []game.RunRecord{fixtureRecord("a", game.RunPrestige, 300, 4, 1)})
	p.setRange(rangeAll)
	p.Layout(p.Bounds())
	row := histRowText(histPaint(p), p.v.geo.footY, p.w)
	for _, w := range []string{txtKeyPointer.String(), txtKeyMeasure.String(), txtKeyRange.String(), txtKeyLog.String(), txtKeyRun.String(), txtKeyGhost.String()} {
		if !strings.Contains(row, w) {
			t.Errorf("key bar %q lacks %q", row, w)
		}
	}
}

// bufText is the text of the rectangle r of b, a line per row.
func bufText(b *kit.Buffer, r image.Rectangle) string {
	var sb strings.Builder
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			sb.WriteRune(b.At(x, y).R)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
