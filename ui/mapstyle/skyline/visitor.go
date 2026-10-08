package skyline

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// visitor.go is the rare visitor (mapmodel/visitors.go): now and then a
// saucer crosses the sky or hovers over the town, and is gone. It flies
// high, behind the ridge, its towns and every building, and it is in no
// legend or key: the point is the surprise. The cursor can stand on it
// (↑ from the ridge) but Tab never goes there.

// saucerW is the saucer's width: a wing, the dome, a wing.
const saucerW = 3

// saucerAt places the saucer at animation frame anim on a view w wide: its
// left column and scene row, and whether a visit is showing at all (the
// saucer may be off an edge, coming or going, or gone up off the top). It
// is anchored to the screen, like the sun. A flyby crosses the view once
// on a row from the visit's roll, bobbing a row. A hover (a walker's ship
// too) glides in to a point above the city, hovers with a slow bob, then
// zips off the way it came or straight up.
func saucerAt(m *mapmodel.Model, anim, w int) (x, y int, ok bool) {
	sg, ok := m.SightingAt(anim)
	if !ok || w <= saucerW {
		return 0, 0, false
	}
	p, r := sg.Phase(anim), sg.Roll
	edge := -saucerW // where it comes in: off the west edge, or the east
	if r&1 == 1 {
		edge = w
	}
	if sg.Kind == mapmodel.SightFlyby {
		d := int(p * float64(w+saucerW))
		if edge == w {
			d = -d
		}
		return edge + d, 1 + int((r>>8)%5) + (anim/12)%2, true
	}
	hx, hy := w/4+int((r>>16)%uint64(max(1, w/2))), 2+int((r>>24)%4)
	switch {
	case p < 0.2: // gliding in, slowing as it comes
		q := 1 - float64((1-p/0.2)*(1-p/0.2))
		return edge + int(q*float64(hx-edge)), hy - 2 + int(float64(q*2)+0.5), true
	case p < 0.8: // hovering
		return hx, hy + (anim/16)%2, true
	}
	q := (p - 0.8) / 0.2
	if (r>>32)&1 == 0 { // off the way it came, faster and faster
		return hx + int(q*q*float64(edge-hx)), hy, true
	}
	return hx, hy - int(q*float64(hy+2)), true // straight up
}

// saucerInView reports whether the saucer's dome is on a view w wide at
// frame anim (where the cursor can stand on it).
func saucerInView(m *mapmodel.Model, anim, w int) bool {
	x, y, ok := saucerAt(m, anim, w)
	return ok && y >= 0 && x+1 >= 0 && x+1 < w
}

// visitor draws the saucer while a visit is showing, at the high-aircraft
// depth: in front of the clouds, behind the ridge, its towns and every
// building. Its glow is made legible on whatever sky it crosses.
func (s *scene) visitor() {
	x, y, ok := saucerAt(s.m, s.visit, s.W)
	if !ok || y < 0 || y >= s.groundY {
		return
	}
	for i, ch := range [saucerW]rune{'◄', mapmodel.R(mapmodel.SymUFO, s.tier), '►'} {
		h := theme.SkyNeonCyan
		if i == 1 {
			h = theme.SkyNavLight
		}
		if c := s.fb.at(x+i, s.Y(y)); c != nil {
			s.fb.fg(x+i, s.Y(y), ch, theme.Legible(s.p.hue(h, mEmit, 0), c.show(), 3), dAir)
		}
	}
}
