package ui

import (
	"image"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

func planTestPanel(t *testing.T, size image.Point) (*planPanel, *game.GameEngine) {
	t.Helper()
	pictureCell = func() (int, int) { return 7, 15 }
	t.Cleanup(func() { pictureCell = nil })
	engine := workbenchEngine(t)
	p := newPlanPanel()
	p.engine = engine
	p.buttons.press = func(k kit.Key) { p.routeKey(&k, "") }
	p.open(engine.GetState())
	p.Layout(image.Rectangle{Max: size})
	if !p.ok {
		t.Fatalf("no board at %v", size)
	}
	return p, engine
}

func planKeyPress(p *planPanel, r rune) bool {
	return p.routeKey(&kit.Key{Code: kit.KeyRune, Rune: r}, "")
}

func planOrder(e *game.GameEngine) (keys []string) {
	for _, v := range e.GetState().Plan {
		keys = append(keys, v.Key)
	}
	return keys
}

func TestPlanPanelKeysReachTheEngine(t *testing.T) {
	p, e := planTestPanel(t, image.Pt(279, 67))
	first := planOrder(e)[0]
	if !p.routeKey(&kit.Key{Code: kit.KeyDown}, "") || p.sel != 1 {
		t.Fatalf("down: sel %d", p.sel)
	}
	if planKeyPress(p, 'p') || !e.GetState().PlanPaused {
		// letters are the panel's: p must be taken and pause the plan
	}
	if !e.GetState().PlanPaused {
		t.Fatal("p did not pause")
	}
	planKeyPress(p, 'p')
	if e.GetState().PlanPaused {
		t.Fatal("p did not resume")
	}
	planKeyPress(p, 't')
	if got := planOrder(e)[0]; got == first || p.sel != 0 {
		t.Fatalf("t: order %v sel %d", planOrder(e), p.sel)
	}
	planKeyPress(p, 'b')
	if p.sel != len(planOrder(e))-1 {
		t.Fatalf("b: sel %d", p.sel)
	}
	before := len(planOrder(e))
	planKeyPress(p, 'x')
	if len(planOrder(e)) != before-1 {
		t.Fatal("x did not remove")
	}
	p.selectTo(0)
	for _, r := range "3" {
		planKeyPress(p, r)
	}
	p.routeKey(&kit.Key{Code: kit.KeyEnter}, "")
	if got := e.GetState().Plan[0].Count; got != 3 {
		t.Fatalf("count box: %d", got)
	}
	planKeyPress(p, '+')
	if got := e.GetState().Plan[0].Count; got != 4 {
		t.Fatalf("+: %d", got)
	}
	if p.routeKey(&kit.Key{Code: kit.KeyRune, Rune: 'p'}, "x") {
		t.Fatal("a letter was taken while the command bar held text")
	}
	planKeyPress(p, 'c')
	if len(planOrder(e)) == 0 {
		t.Fatal("one c cleared the plan")
	}
	planKeyPress(p, 'c')
	if len(planOrder(e)) != 0 {
		t.Fatal("c twice did not clear")
	}
}

func TestPlanPanelOptimizeKeepsOrDiscards(t *testing.T) {
	p, e := planTestPanel(t, image.Pt(279, 67))
	was := planOrder(e)
	planKeyPress(p, 'o')
	if p.mode != planPreview {
		t.Skip("nothing to optimize in this plan")
	}
	if !p.closeFirst() || p.mode != planQueue {
		t.Fatal("Esc did not close the preview")
	}
	if got := planOrder(e); len(got) != len(was) {
		t.Fatalf("discard changed the plan: %v", got)
	}
	planKeyPress(p, 'o')
	p.routeKey(&kit.Key{Code: kit.KeyEnter}, "")
	if p.mode != planQueue {
		t.Fatal("Enter did not leave the preview")
	}
}

func TestPlanPanelClickAndDrag(t *testing.T) {
	p, e := planTestPanel(t, image.Pt(279, 67))
	var rows []int
	var ys []int
	for _, s := range p.drawn.Slots {
		if s.Item >= 0 {
			rows, ys = append(rows, s.Item), append(ys, s.Y)
		}
	}
	if len(rows) < 3 {
		t.Fatalf("only %d rows", len(rows))
	}
	x := p.l.ListX + 10
	p.Mouse(kit.Mouse{Kind: kit.MousePress, X: x, Y: ys[1], Button: kit.ButtonLeft})
	if p.sel != rows[1] {
		t.Fatalf("click: sel %d", p.sel)
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, X: x, Y: ys[0], Button: kit.ButtonLeft})
	moved := planOrder(e)[0]
	p.Mouse(kit.Mouse{Kind: kit.MouseMove, X: x, Y: ys[2]})
	if !p.drag.on {
		t.Fatal("no drag started")
	}
	p.Mouse(kit.Mouse{Kind: kit.MouseRelease, X: x, Y: ys[2], Button: kit.ButtonLeft})
	if planOrder(e)[0] == moved {
		t.Fatalf("drag did not move the first item: %v", planOrder(e))
	}
	// A tray entry adds below the selection.
	n := len(planOrder(e))
	er, ok := p.l.EntryRect(0)
	if !ok || len(p.shown) == 0 {
		t.Fatal("tray has no first entry")
	}
	p.Mouse(kit.Mouse{Kind: kit.MousePress, X: er.Min.X + 5, Y: er.Min.Y, Button: kit.ButtonLeft})
	if len(planOrder(e)) < n {
		t.Fatalf("entry click lost items: %v", planOrder(e))
	}
}

func TestPlanPanelUnchangedRefreshKeepsPictures(t *testing.T) {
	p, e := planTestPanel(t, image.Pt(279, 67))
	var before []*image.RGBA
	for _, q := range p.pics {
		before = append(before, q.img)
	}
	p.update(e.GetState())
	if len(p.pics) != len(before) {
		t.Fatalf("pictures %d, then %d", len(before), len(p.pics))
	}
	for i, q := range p.pics {
		if q.img != before[i] {
			t.Fatalf("picture %d was drawn again", i)
		}
	}
}

func TestPlanPanelIsFast(t *testing.T) {
	p, e := planTestPanel(t, image.Pt(279, 67))
	st := e.GetState()
	start := time.Now()
	for range 20 {
		p.st = st
		p.relayout()
	}
	if per := time.Since(start) / 20; per > 8*time.Millisecond {
		t.Errorf("a refresh takes %v, over 8ms", per)
	}
	start = time.Now()
	_ = p.l.Board()
	if d := time.Since(start); d > 60*time.Millisecond {
		t.Errorf("the board takes %v, over 60ms", d)
	}
}
