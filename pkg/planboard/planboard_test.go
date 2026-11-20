package planboard

import (
	"image"
	"os"
	"reflect"
	"testing"
	"time"
)

var sizes = [][2]int{{279, 71}, {279, 68}, {157, 43}, {157, 40}}

func TestLayoutHasNoOverlap(t *testing.T) {
	for _, sz := range sizes {
		l, ok := NewLayout(sz[0], sz[1], 7, 15)
		if !ok {
			t.Fatalf("%v: no layout", sz)
		}
		list := image.Rect(l.ListX, l.ListY, l.ListX+l.ListW, l.SchedEnd)
		job, gate := l.Cards(10)
		parts := map[string]image.Rectangle{
			"list": list, "lanes": l.Lanes, "right": l.Right, "waits": l.Waits, "tray": l.Tray,
		}
		parts["ruler"] = l.Ruler
		if l.Compact {
			parts["gate"], parts["job"] = image.Rectangle{}, image.Rectangle{}
		} else {
			parts["job"], parts["gate"] = job, gate
			delete(parts, "right")
		}
		names := make([]string, 0, len(parts))
		for k := range parts {
			names = append(names, k)
		}
		for _, a := range names {
			if !parts[a].Empty() && !parts[a].In(l.Inner) {
				t.Errorf("%v: %s %v outside %v", sz, a, parts[a], l.Inner)
			}
			for _, b := range names {
				if a >= b || parts[a].Empty() || parts[b].Empty() {
					continue
				}
				if (a == "lanes" && b == "ruler") || (a == "ruler" && b == "lanes") {
					continue
				}
				if parts[a].Overlaps(parts[b]) {
					t.Errorf("%v: %s %v overlaps %s %v", sz, a, parts[a], b, parts[b])
				}
			}
		}
		if l.Park.Max.X > l.Lanes.Max.X || l.Park.Min.X <= l.Lanes.Min.X {
			t.Errorf("%v: park %v outside lanes %v", sz, l.Park, l.Lanes)
		}
		if l.Hint < l.Inner.Max.Y {
			t.Errorf("%v: hint row %d inside the inner cells", sz, l.Hint)
		}
	}
}

func TestSmallRectHasNoLayout(t *testing.T) {
	for _, sz := range [][2]int{{149, 60}, {200, 37}, {68, 20}} {
		if _, ok := NewLayout(sz[0], sz[1], 7, 15); ok {
			t.Errorf("%v: a layout below the minimum", sz)
		}
	}
}

func TestBarsNeverMoveLeftAsStartGrows(t *testing.T) {
	l, _ := NewLayout(279, 71, 7, 15)
	marks := []time.Duration{0, time.Hour, 5 * time.Hour, 30 * time.Hour, 200 * time.Hour}
	ax := newAxis(marks, 200*time.Hour, l.Lanes.Min.X+1, l.RulerCells, 7)
	prev := -1
	for d := time.Duration(0); d <= 200*time.Hour; d += 7 * time.Minute {
		x := ax.x(d)
		if x < prev {
			t.Fatalf("x(%v)=%d is left of %d", d, x, prev)
		}
		prev = x
	}
	if ax.end() < prev {
		t.Errorf("axis ends at %d before the last bar at %d", ax.end(), prev)
	}
}

func TestRulerUsesTheLadderInOrder(t *testing.T) {
	l, _ := NewLayout(279, 71, 7, 15)
	d := l.Schedule(Sched{Rows: []Row{{Seed: "a", Sign: "build", End: 3 * time.Hour}}, Sel: -1, Drop: -1})
	if len(d.Ticks) < 2 || d.Ticks[0].Name != LadderNames[0] {
		t.Fatalf("ticks %v", d.Ticks)
	}
	for i := 1; i < len(d.Ticks); i++ {
		if d.Ticks[i].Cell <= d.Ticks[i-1].Cell || d.Ticks[i].Name != LadderNames[i] {
			t.Errorf("tick %d: %v after %v", i, d.Ticks[i], d.Ticks[i-1])
		}
	}
}

func TestSlotsPlaceHeadersAndReserveMore(t *testing.T) {
	l, _ := NewLayout(279, 71, 7, 15)
	s, more := l.Slots(0, 5, []int{1})
	want := []Slot{{-1, 6, 0}, {0, 7, 0}, {1, 9, 0}, {-1, 11, 1}, {2, 12, 1}, {3, 14, 1}, {4, 16, 1}}
	if !reflect.DeepEqual(s, want) || more != 0 {
		t.Errorf("slots %v more %d", s, more)
	}
	s, more = l.Slots(0, 200, nil)
	if more == 0 || len(s) == 0 {
		t.Fatal("200 items fit")
	}
	last := s[len(s)-1]
	if last.Y+l.RowH > l.SchedEnd-1 {
		t.Errorf("last item at %d leaves no row for the count", last.Y)
	}
	if more != 200-(len(s)-1) {
		t.Errorf("more %d with %d shown", more, len(s)-1)
	}
}

func rows(seedOf func(int) string) []Row {
	r := make([]Row, 4)
	for i := range r {
		r[i] = Row{Seed: seedOf(i), Sign: "build", Start: time.Duration(i) * time.Hour, End: time.Duration(i+1) * time.Hour}
	}
	return r
}

func TestDecorationFollowsTheItemNotTheRow(t *testing.T) {
	l, _ := NewLayout(279, 71, 7, 15)
	sched := func(r []Row, first int) Sched { return Sched{Rows: r, First: first, Sel: -1, Drop: -1} }
	all := rows(func(i int) string { return "k" + string(rune('a'+i)) })
	a := l.Schedule(sched(all, 0))
	b := l.Schedule(sched(all[1:], 0)) // the same items, one row higher
	var ay, by int
	for _, s := range a.Slots {
		if s.Item == 1 {
			ay = s.Y
		}
	}
	for _, s := range b.Slots {
		if s.Item == 0 {
			by = s.Y
		}
	}
	ch, cw := l.CH, l.CW
	// the sign's block of the same item, in rows 2 and 1 of the two pictures
	crop := func(d Drawn, y int) []byte {
		x0 := (l.ListX + 3) * cw
		var out []byte
		for py := y * ch; py < (y+l.RowH)*ch; py++ {
			for px := x0; px < x0+3*cw; px++ {
				p := d.Img.RGBAAt(px-d.Rect.Min.X*cw, py-d.Rect.Min.Y*ch)
				out = append(out, p.R, p.G, p.B)
			}
		}
		return out
	}
	if !reflect.DeepEqual(crop(a, ay), crop(b, by)) {
		t.Error("an item's sign changed with its row")
	}
}

func TestScheduleIsSteadyAndFast(t *testing.T) {
	l, _ := NewLayout(279, 74, 7, 15)
	r := rows(func(i int) string { return "k" + string(rune('a'+i)) })
	r[1].Mode, r[1].Cause = Parked, "income"
	r[2] = Row{Seed: "adv", Sign: "advance", Gate: true}
	s := Sched{Rows: r, Sel: 0, Drop: 2, FinishW: 14, UntimedW: 9}
	a, b := l.Schedule(s), l.Schedule(s)
	if !reflect.DeepEqual(a.Img.Pix, b.Img.Pix) || a.GateX < 0 {
		t.Error("two draws differ, or the gate is missing")
	}
	t0 := time.Now()
	l.Board()
	// Shared CI runners are too noisy for a wall-clock limit.
	if d := time.Since(t0); os.Getenv("CI") == "" && d > 60*time.Millisecond {
		t.Errorf("board took %v", d)
	}
}
