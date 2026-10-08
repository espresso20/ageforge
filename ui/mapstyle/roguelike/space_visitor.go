package roguelike

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_visitor.go is the rare visitor in the sky scenes (the Space and
// Interstellar Ages; from the Galactic Age the aliens are traffic and the
// model schedules none): a saucer crossing the view, or dropping in over
// the hub to hover a while and shoot off. Like the town's, it never takes a
// legend row, never shows in the mini view and never covers a building.

// skyVisit is where the saucer's dome is this frame, if a visit is on.
func (v *skyView) skyVisit() (mapmodel.Pt, mapmodel.Sighting, bool) {
	s := v.sc
	if s == nil || v.compact {
		return mapmodel.Pt{}, mapmodel.Sighting{}, false
	}
	sg, ok := s.m.SightingAt(v.visit)
	if !ok {
		return mapmodel.Pt{}, sg, false
	}
	g := v.sg
	tw := max(1, g.w/g.cellW)
	ph := sg.Phase(v.visit)
	if sg.Kind == mapmodel.SightFlyby {
		x := g.vx - 2 + int(float64(tw+3)*ph)
		if sg.Roll%2 == 1 {
			x = g.vx + tw + 1 - int(float64(tw+3)*ph)
		}
		y := g.vy + g.h*int(1+sg.Roll>>4%4)/10 + int(float64(int(sg.Roll>>8%7)-3)*ph)
		return pt(x, y), sg, true
	}
	// a hover (and a walker, who has no streets up here): over the hub
	tx := s.hub.X - 14 + int(sg.Roll>>4%29)
	ty := s.hub.Y - 6 + int(sg.Roll>>12%5)
	top := min(ty, g.vy) - 2
	switch {
	case ph < 0.15:
		return pt(tx, top+int(float64(ty-top)*ph/0.15)), sg, true
	case ph < 0.85:
		return pt(tx+[4]int{0, 1, 0, -1}[(v.visit/10)%4], ty), sg, true
	}
	return pt(tx, ty-int(float64(ty-top)*(ph-0.85)/0.15)), sg, true
}

// drawSkyVisitor draws the saucer, in the fresh-news green, cell by cell on
// screen (a plane cell is two cells wide at the close zoom).
func (v *skyView) drawSkyVisitor(cv *mapstyle.Canvas) {
	p, _, ok := v.skyVisit()
	if !ok || blocks(v.sc.at(p.X, p.Y)) {
		return
	}
	cx, cy, in := v.sg.cellOf(p)
	if !in {
		return
	}
	for dx, r := range map[int]rune{-1: '◄', 0: mapmodel.R(mapmodel.SymUFO, v.tier), 1: '►'} {
		x := cx + dx
		if x < v.sg.x || x >= v.sg.x+v.sg.w {
			continue
		}
		_, cur := cv.Get(x, cy)
		_, bg, _ := cur.Decompose()
		cv.Put(x, cy, r, v.style(look{r: r, ink: inkFresh, lv: 2, bgc: bg, bold: true}))
	}
}

// skyVisitorAt is the inspection for the saucer when the cursor is on it.
func (v *skyView) skyVisitorAt(p mapmodel.Pt) (mapstyle.Inspection, bool) {
	q, sg, ok := v.skyVisit()
	if !ok || p.Y != q.Y || abs(p.X-q.X) > 1 {
		return mapstyle.Inspection{}, false
	}
	return mapstyle.Inspection{Title: "Unknown craft", Lines: []string{saucerLines[sg.Roll>>24%uint64(len(saucerLines))]},
		Kind: mapstyle.KindAlien}, true
}
