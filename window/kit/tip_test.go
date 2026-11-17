package kit

import (
	"image"
	"strings"
	"testing"
	"time"
)

type tipClock struct{ t time.Time }

func (c *tipClock) now() time.Time { return c.t }

func newTip() (*Tip, *tipClock) {
	c := &tipClock{time.Unix(1000, 0)}
	return &Tip{Now: c.now}, c
}

var tipTarget = image.Rect(10, 5, 16, 6)

func TestTipShowsOnlyAfterDelay(t *testing.T) {
	for _, byKey := range []bool{false, true} {
		tip, c := newTip()
		tip.Aim("a", tipTarget, TipText{Lines: []string{"x"}}, byKey)
		if tip.Visible() || !tip.Waiting() {
			t.Fatalf("byKey=%v: shown at once", byKey)
		}
		c.t = c.t.Add(TipDelay - time.Millisecond)
		if tip.Visible() {
			t.Fatalf("byKey=%v: shown early", byKey)
		}
		c.t = c.t.Add(time.Millisecond)
		if !tip.Visible() || tip.Waiting() {
			t.Fatalf("byKey=%v: not shown after the delay", byKey)
		}
	}
}

func TestTipPointerRestartsKeyKeepsTheDelay(t *testing.T) {
	tip, c := newTip()
	tip.Aim("a", tipTarget, TipText{}, false)
	c.t = c.t.Add(300 * time.Millisecond)
	tip.Aim("a", tipTarget, TipText{}, false) // moved within the target
	c.t = c.t.Add(300 * time.Millisecond)
	if tip.Visible() {
		t.Fatal("a move within the target did not restart the rest")
	}
	tip, c = newTip()
	tip.Aim("a", tipTarget, TipText{}, true)
	c.t = c.t.Add(300 * time.Millisecond)
	tip.Aim("a", tipTarget, TipText{}, true)
	c.t = c.t.Add(100 * time.Millisecond)
	if !tip.Visible() {
		t.Fatal("the same pick restarted the delay")
	}
}

func TestTipGoes(t *testing.T) {
	show := func() *Tip {
		tip, c := newTip()
		tip.Aim("a", tipTarget, TipText{Lines: []string{"x"}}, false)
		c.t = c.t.Add(TipDelay)
		if !tip.Visible() {
			t.Fatal("not shown")
		}
		return tip
	}
	tip := show()
	tip.Move(Mouse{Kind: MouseMove, X: 11, Y: 5})
	if !tip.Visible() {
		t.Error("a move within the target hid it")
	}
	for name, m := range map[string]Mouse{
		"move off": {Kind: MouseMove, X: 40, Y: 20},
		"click":    {Kind: MousePress, X: 11, Y: 5},
		"leave":    {Kind: MouseLeave, X: 11, Y: 5},
	} {
		tip := show()
		tip.Move(m)
		if tip.Visible() || tip.Waiting() {
			t.Errorf("%s did not clear it", name)
		}
	}
	tip = show()
	tip.Clear() // a key or a change of screen
	if tip.Visible() {
		t.Error("Clear left it showing")
	}
}

func TestTipStaysInsideAndOffTheTarget(t *testing.T) {
	area := image.Rect(0, 0, 80, 24)
	long := TipText{Lines: []string{strings.Repeat("word ", 20), "b", "c", "d", "e", "f", "g"}, Reason: "short"}
	for name, target := range map[string]image.Rectangle{
		"middle": image.Rect(30, 10, 36, 11),
		"right":  image.Rect(74, 5, 80, 6),
		"bottom": image.Rect(30, 21, 36, 22),
		"corner": image.Rect(74, 21, 80, 22),
		"left":   image.Rect(0, 0, 4, 1),
	} {
		for _, pics := range []bool{true, false} {
			tip, c := newTip()
			if pics {
				tip.Cell = func() (int, int) { return 7, 15 }
			}
			tip.Aim("a", target, long, false)
			c.t = c.t.Add(TipDelay)
			r := tip.Rect(area)
			if r.Empty() || !r.In(area) {
				t.Errorf("%s pics=%v: %v outside %v", name, pics, r, area)
			}
			if r.Overlaps(target) {
				t.Errorf("%s pics=%v: %v covers the target %v", name, pics, r, target)
			}
			if r.Dx() > TipMaxCols || r.Dy() > TipMaxRows {
				t.Errorf("%s pics=%v: %v is over %dx%d", name, pics, r, TipMaxCols, TipMaxRows)
			}
		}
	}
	// The bottom rows (the command bar) are left out of the area: the plate
	// goes above a target that has no room below.
	bar := image.Rect(0, 0, 80, 21)
	tip, c := newTip()
	tip.Aim("a", image.Rect(30, 19, 36, 20), TipText{Lines: []string{"a", "b", "c"}}, false)
	c.t = c.t.Add(TipDelay)
	if r := tip.Rect(bar); !r.In(bar) || r.Max.Y > 19 {
		t.Errorf("plate %v reaches the bar or the target", r)
	}
}

func TestTipPaintsAndCuts(t *testing.T) {
	for _, pics := range []bool{true, false} {
		tip, c := newTip()
		if pics {
			tip.Cell = func() (int, int) { return 7, 15 }
		}
		tip.Aim("a", tipTarget, TipText{Lines: []string{"first", strings.Repeat("x", 60), "3", "4", "5", "6"}, Reason: "nope"}, false)
		c.t = c.t.Add(TipDelay)
		b := NewBuffer(80, 24)
		tip.Paint(NewPainter(b), image.Rect(0, 0, 80, 24))
		r := tip.Rect(image.Rect(0, 0, 80, 24))
		var rows []string
		for y := r.Min.Y; y < r.Max.Y; y++ {
			var s strings.Builder
			for x := r.Min.X; x < r.Max.X; x++ {
				s.WriteRune(b.At(x, y).R)
			}
			rows = append(rows, s.String())
		}
		all := strings.Join(rows, "\n")
		if !strings.Contains(all, "first") || !strings.Contains(all, "! nope") {
			t.Errorf("pics=%v: missing text in\n%s", pics, all)
		}
		if !strings.Contains(all, "…") && !strings.Contains(all, "...") {
			t.Errorf("pics=%v: long text not cut in\n%s", pics, all)
		}
		if !pics && !strings.HasPrefix(rows[0], "+-") {
			t.Errorf("plain plate has no ASCII border:\n%s", all)
		}
		if pics && len(b.Pics) != 1 {
			t.Errorf("want one picture, got %d", len(b.Pics))
		}
	}
}
