package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// visitor.go draws the rare visitor the shared schedule calls for
// (mapmodel.SightingAt): a saucer crossing the sky or hovering over the
// town, or a small figure walking a street, for ten to thirty seconds. It
// never claims a legend row, never shows in the mini view and never covers
// a building. Inspecting it says something fun and reports
// mapstyle.KindAlien.

// visit is where the visitor is this frame.
type visit struct {
	at     [3]mapmodel.Pt // a saucer's rim, dome, rim; a walker uses at[1]
	saucer bool
	waving bool
	s      mapmodel.Sighting
}

var (
	saucerLines = []string{"Not one of ours.", "It hums, then it's gone.", "Lights in the sky. Nobody will believe you."}
	walkerLines = []string{"It waves, then it's gone.", "A small visitor, far from home.", "It seems lost, and fine with that."}
)

// visitNow is the visitor this frame, if one is out.
func (v *view) visitNow() (visit, bool) {
	s := v.sc
	if s == nil || v.g.zoom == zRegion {
		return visit{}, false
	}
	sg, ok := s.m.SightingAt(v.anim)
	if !ok {
		return visit{}, false
	}
	switch sg.Kind {
	case mapmodel.SightWalker:
		if vi, ok := v.walkerVisit(sg); ok {
			return vi, true
		}
		return v.hoverVisit(sg) // no streets to walk: it hovers instead
	case mapmodel.SightHover:
		return v.hoverVisit(sg)
	}
	return v.flybyVisit(sg), true
}

func saucerAt(sg mapmodel.Sighting, x, y int) visit {
	return visit{at: [3]mapmodel.Pt{pt(x-1, y), pt(x, y), pt(x+1, y)}, saucer: true, s: sg}
}

// flybyVisit crosses the view from one side to the other, drifting a few
// rows, so it shows wherever the player is looking.
func (v *view) flybyVisit(sg mapmodel.Sighting) visit {
	g := v.g
	tw := max(1, g.w/g.cellW)
	ph := sg.Phase(v.anim)
	x := g.vx - 2 + int(float64(float64(tw+3)*ph))
	if sg.Roll%2 == 1 {
		x = g.vx + tw + 1 - int(float64(float64(tw+3)*ph))
	}
	y0 := g.vy + g.h*int(2+sg.Roll>>4%5)/10
	drift := int(sg.Roll>>8%7) - 3
	return saucerAt(sg, x, y0+int(float64(float64(drift)*ph)))
}

// hoverVisit drops in over the town near the square, hangs there with a
// slow wobble, then shoots off upward.
func (v *view) hoverVisit(sg mapmodel.Sighting) (visit, bool) {
	s, g := v.sc, v.g
	w := s.w
	tx := w.CX - 18 + int(sg.Roll>>4%37)
	ty := w.CY - 7 + int(sg.Roll>>12%15)
	open := func(x, y int) bool {
		for dx := -1; dx <= 1; dx++ {
			if k := s.at(x+dx, y).k; !s.in(x+dx, y) || k == kTile || k == kWonder || k == kSite || k == kCentre {
				return false
			}
		}
		return true
	}
	found := false
	for rad := 0; rad <= 6 && !found; rad++ {
		for dy := -rad; dy <= rad && !found; dy++ {
			for dx := -2 * rad; dx <= 2*rad && !found; dx++ {
				if open(tx+dx, ty+dy) {
					tx, ty, found = tx+dx, ty+dy, true
				}
			}
		}
	}
	if !found {
		return visit{}, false
	}
	ph := sg.Phase(v.anim)
	top := min(ty, g.vy) - 2
	switch {
	case ph < 0.15:
		return saucerAt(sg, tx, top+int(float64(float64(ty-top)*ph/0.15))), true
	case ph < 0.85:
		return saucerAt(sg, tx+[4]int{0, 1, 0, -1}[(v.anim/10)%4], ty), true
	}
	return saucerAt(sg, tx, ty-int(float64(float64(ty-top)*(ph-0.85)/0.15))), true
}

// walkerVisit walks a street at a stroll, stops halfway to wave, and walks
// on until it is gone.
func (v *view) walkerVisit(sg mapmodel.Sighting) (visit, bool) {
	lanes := v.sc.streetLanes()
	if len(lanes) == 0 {
		return visit{}, false
	}
	lane := lanes[sg.Roll%uint64(len(lanes))]
	if len(lane) < 4 {
		return visit{}, false
	}
	t := v.anim - sg.Start
	stop, pause := sg.Frames*9/20, sg.Frames*3/20 // stand and wave for a while, halfway along
	waving := t >= stop && t < stop+pause
	if t >= stop+pause {
		t -= pause
	} else if waving {
		t = stop
	}
	start := int(sg.Roll >> 20 % uint64(len(lane)))
	i := start + t/walkFrames
	if sg.Roll>>40%2 == 1 { // it walks the lane the other way
		i = start - t/walkFrames
	}
	i = ((i % (2*len(lane) - 2)) + 2*len(lane) - 2) % (2*len(lane) - 2)
	i = min(i, 2*len(lane)-2-i)
	at := lane[i]
	return visit{at: [3]mapmodel.Pt{lane[max(0, i-1)], at, lane[min(len(lane)-1, i+1)]}, waving: waving, s: sg}, true
}

// drawVisitor draws the visitor, if one is out, in the fresh-news green.
func (v *view) drawVisitor(cv *mapstyle.Canvas) {
	vi, ok := v.visitNow()
	if !ok {
		return
	}
	st := func(cx, cy int) tcell.Style {
		_, cur := cv.Get(cx, cy)
		_, bg, _ := cur.Decompose()
		return v.style(glyph{c: mapmodel.CFresh, bg: bg, attr: tcell.AttrBold})
	}
	put := func(p mapmodel.Pt, dx int, r rune) {
		cx, cy, ok := v.g.cellOf(p)
		if k := v.sc.at(p.X, p.Y).k; !ok || k == kTile || k == kWonder || k == kSite || k == kCentre {
			return
		}
		if cx += dx; cx >= v.g.x && cx < v.g.x+v.g.w {
			cv.Put(cx, cy, r, st(cx, cy))
		}
	}
	if !vi.saucer {
		put(vi.at[1], 0, mapmodel.R(mapmodel.SymAlien, v.tier))
		if vi.waving && (v.anim/3)%2 == 0 {
			put(pt(vi.at[1].X, vi.at[1].Y-1), 0, '˚')
		}
		return
	}
	// the saucer: rims either side of the dome, cell by cell on screen (at
	// the district zoom a tile is two cells wide)
	put(vi.at[1], -1, '◄')
	put(vi.at[1], 0, mapmodel.R(mapmodel.SymUFO, v.tier))
	put(vi.at[1], 1, '►')
}

// visitorAt is the inspection for the visitor when the cursor is on it.
func (v *view) visitorAt(p mapmodel.Pt) (insp, bool) {
	vi, ok := v.visitNow()
	if !ok || p != vi.at[0] && p != vi.at[1] && p != vi.at[2] {
		return insp{}, false
	}
	if vi.saucer {
		return insp{Title: "Unknown craft", Lines: []string{saucerLines[vi.s.Roll>>24%uint64(len(saucerLines))]},
			Kind: mapstyle.KindAlien}, true
	}
	return insp{Title: "Visitor", Lines: []string{walkerLines[vi.s.Roll>>24%uint64(len(walkerLines))]},
		Kind: mapstyle.KindAlien}, true
}
