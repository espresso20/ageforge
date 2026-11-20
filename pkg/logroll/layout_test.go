package logroll

import "testing"

func TestPlanRects(t *testing.T) {
	for _, s := range [][2]int{{279, 71}, {279, 68}, {157, 43}, {157, 40}, {230, 54}, {150, 38}} {
		l := Plan(s[0], s[1])
		if l.Mode == Placeholder {
			t.Fatalf("%v: placeholder", s)
		}
		whole := Rect{0, 0, l.Cols, l.Rows}
		top := map[string]Rect{"title": l.Title, "rack": l.Rack, "lever": l.Lever,
			"roll": l.Roll, "rail": l.Rail, "slate": l.Slate, "hint": l.Hint}
		for a, ra := range top {
			if !whole.Holds(ra) {
				t.Errorf("%v: %s outside the panel: %+v", s, a, ra)
			}
			for b, rb := range top {
				if a < b && ra.Overlaps(rb) {
					t.Errorf("%v: %s overlaps %s: %+v %+v", s, a, b, ra, rb)
				}
			}
		}
		in := func(name string, p Rect, c ...Rect) {
			for _, r := range c {
				if !p.Holds(r) {
					t.Errorf("%v: %s does not hold %+v (%+v)", s, name, r, p)
				}
			}
		}
		in("title", l.Title, l.Find, l.Keys)
		in("roll", l.Roll, l.Head)
		in("rack", l.Rack, l.Stamps[:]...)
		in("lever", l.Lever, l.Stops[:]...)
		in("slate", l.Slate, l.SlateBody, l.Ledge)
		in("roll", l.Roll, l.Body)
		if l.SlateBody.Overlaps(l.Ledge) {
			t.Errorf("%v: slate body overlaps ledge", s)
		}
		for i, a := range l.Stamps {
			for j, b := range l.Stamps {
				if i < j && a.Overlaps(b) {
					t.Errorf("%v: stamps %d and %d overlap", s, i, j)
				}
			}
		}
		cols := []Rect{l.Clock, l.Stamp, l.Text, l.Paid, l.Gained, l.Times}
		for i, a := range cols {
			if a.W > 0 && (a.X < l.Roll.X || a.Right() > l.Roll.Right()) {
				t.Errorf("%v: column %d leaves the roll", s, i)
			}
			for j, b := range cols {
				if i < j && a.W > 0 && b.W > 0 && a.X < b.Right() && b.X < a.Right() {
					t.Errorf("%v: columns %d and %d overlap", s, i, j)
				}
			}
		}
		if l.Text.W < 20 {
			t.Errorf("%v: text column only %d wide", s, l.Text.W)
		}
	}
}

func TestPlanModes(t *testing.T) {
	for _, c := range []struct {
		w, h int
		m    Mode
	}{{229, 70, Compact}, {230, 53, Compact}, {230, 54, Full}, {150, 38, Compact},
		{149, 60, Placeholder}, {200, 37, Placeholder}} {
		if got := Plan(c.w, c.h).Mode; got != c.m {
			t.Errorf("%dx%d: mode %d, want %d", c.w, c.h, got, c.m)
		}
	}
}

func TestPlanNoFixedRows(t *testing.T) {
	a, b := Plan(279, 71), Plan(279, 68)
	if a.Body.H-b.Body.H != 3 || a.Lever.H-b.Lever.H != 3 || a.Slate.H-b.Slate.H != 3 {
		t.Errorf("three more rows should reach roll, lever and slate: %d %d %d",
			a.Body.H-b.Body.H, a.Lever.H-b.Lever.H, a.Slate.H-b.Slate.H)
	}
	if a.Rack != b.Rack || a.Slate.X != b.Slate.X {
		t.Error("the rack and the columns must not move with the rows")
	}
}
