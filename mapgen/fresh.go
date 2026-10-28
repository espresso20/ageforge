package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// What is new since the last visit is lit on the ground it stands on, and a
// district can be brought into focus by dimming the buildings of the other
// lines. Both are passes over the town as it was drawn.

// freshTint is the ground a new building or wonder stands on.
var freshTint = P(HFresh, 70)

// focusDim is how far the buildings outside the district in focus are dimmed.
const focusDim = 150

// paintFresh lights the parcels of the building types that grew since the
// last visit, and the plots of the wonders finished since.
func (f *frame) paintFresh() {
	if !f.fresh {
		return
	}
	light := func(r Rect) {
		r.cells(func(x, y int) { f.g.Cells[at(x, y)].Bg = freshTint })
	}
	for _, pl := range f.placed {
		if !pl.ruin && pl.b.Delta > 0 {
			light(f.plan.Lots[pl.lot].Rect)
		}
	}
	for _, ws := range f.wonders() {
		if ws.w.Delta {
			light(ws.plot.Rect)
		}
	}
}

// paintFocus dims the buildings of every line but the district in focus.
func (f *frame) paintFocus() {
	if f.district == "" {
		return
	}
	for i, h := range f.g.Hit {
		th := f.g.Things[h]
		if (th.Kind == KindBuilding && th.Lineage != f.district) || th.Kind == KindAmbient {
			f.g.Cells[i].Dim = dimmer(f.g.Cells[i].Dim, focusDim)
		}
	}
}

// District is one building line's part of the town: the box round its
// buildings and the building nearest the middle of it.
type District struct {
	Line  string
	Box   Rect
	Heart Pt
}

// districts lists the building lines with something standing, in the order of
// mapmodel.LineageOrder. The targets are the buildings and wonders, nearest
// the square first.
func (f *frame) districts() []District {
	var out []District
	for _, line := range mapmodel.LineageOrder {
		var box Rect
		for _, pl := range f.placed {
			if pl.ruin || pl.b.Lineage != line {
				continue
			}
			r := f.plan.Lots[pl.lot].Rect
			if box.W == 0 {
				box = r
				continue
			}
			x0, y0 := min(box.X, r.X), min(box.Y, r.Y)
			x1, y1 := max(box.X+box.W, r.X+r.W), max(box.Y+box.H, r.Y+r.H)
			box = Rect{x0, y0, x1 - x0, y1 - y0}
		}
		if box.W == 0 {
			continue
		}
		middle := Pt{box.X + box.W/2, box.Y + box.H/2}
		heart, gap := Pt{}, 1<<60
		for _, t := range f.targets {
			th := f.g.ThingAt(t.X, t.Y)
			if d := (Rect{t.X, t.Y, 1, 1}).dist(middle); th.Kind == KindBuilding && th.Lineage == line && d < gap {
				heart, gap = t, d
			}
		}
		if gap < 1<<60 {
			out = append(out, District{line, box, heart})
		}
	}
	return out
}
