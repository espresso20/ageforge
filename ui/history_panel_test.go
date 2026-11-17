package ui

import (
	"image"
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
