package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// citytraffic.go is the Earth arc's own traffic and its ambient effects:
// cars streaming down the highways, the news helicopter circling the town,
// drone swarms, hovercars and crowds in the megacity's alleys, sky trains
// on every elevated line (maglevs once fusion powers the city) and
// climbers on the tether; and, frame by frame, headlights on the highways
// at night, steam from the megacity's vents and its acid rain. Everything
// rides lanes the scene laid once.

// cityMovers adds the city's own traffic to the scene's movers.
func (s *scene) cityMovers(add addFn) {
	cl := s.city
	if cl == nil {
		return
	}
	m := s.m
	busy := m.Activity.Traffic
	about := func(k mapmodel.Mover) bool { return k.Info().In(m.AgeIdx) }
	for li, ln := range cl.lines {
		lane := [][]mapmodel.Pt{ln.pts}
		switch ln.feat {
		case mapmodel.FeatHighway:
			if about(mapmodel.MoverCar) {
				for j := 0; j < 2+int(float64(busy*2)); j++ {
					add(mapmodel.MoverCar, lane, 100+li*8+j, false, 1)
				}
			}
			if about(mapmodel.MoverTruck) {
				add(mapmodel.MoverTruck, lane, 100+li, false, 2)
			}
		case mapmodel.FeatSkyRail:
			if about(mapmodel.MoverSkyTrain) {
				for j := 0; j < 2; j++ {
					add(mapmodel.MoverSkyTrain, lane, li*4+j, false, 6)
				}
			}
		case mapmodel.FeatMaglevLine:
			if about(mapmodel.MoverMaglev) {
				for j := 0; j < 1+b2i(busy > 0.5); j++ {
					add(mapmodel.MoverMaglev, lane, 10+li*4+j, false, 10)
				}
			}
		case mapmodel.FeatTether:
			if about(mapmodel.MoverClimber) {
				for j := 0; j < 3; j++ {
					add(mapmodel.MoverClimber, lane, j, true, 14)
				}
			}
		}
	}
	if about(mapmodel.MoverNewsHeli) { // circling the story
		for j := 0; j < 1+b2i(busy > 0.5); j++ {
			add(mapmodel.MoverNewsHeli, [][]mapmodel.Pt{s.loop(j)}, j, true, 0)
		}
	}
	if !cl.megacity() {
		return
	}
	lanes := s.alleyLanes()
	if len(lanes) == 0 {
		return
	}
	hover, swarms, crowds := 6+int(float64(busy*6)), 3, 10+int(float64(busy*8))
	if cl.fusion() {
		hover, swarms, crowds = 3+int(float64(busy*3)), 1, 5+int(float64(busy*3))
	}
	if about(mapmodel.MoverHovercar) {
		for j := 0; j < hover; j++ {
			add(mapmodel.MoverHovercar, [][]mapmodel.Pt{lanes[(j*3+1)%len(lanes)]}, 50+j, false, 2)
		}
	}
	if about(mapmodel.MoverDrone) { // a swarm keeps together, two cells apart
		for sw := 0; sw < swarms; sw++ {
			lane := [][]mapmodel.Pt{lanes[(sw*5+2)%len(lanes)]}
			lead := -1
			for d := 0; d < 4; d++ {
				n := len(s.movers)
				add(mapmodel.MoverDrone, lane, 60+sw*8+d, false, 3)
				if len(s.movers) == n {
					break
				}
				if lead < 0 {
					lead = n
				} else if p := s.movers[n].period(); p > 0 {
					s.movers[n].off = (s.movers[lead].off + 2*d) % p
				}
			}
		}
	}
	if about(mapmodel.MoverCrowd) { // milling to and fro on a short stretch
		for j := 0; j < crowds; j++ {
			l := lanes[j%len(lanes)]
			n := 8 + int(mapmodel.Hash(s.w.Seed, 361, int64(j))%6)
			if len(l) <= n+1 {
				continue
			}
			at := int(mapmodel.Hash(s.w.Seed, 362, int64(j)) % uint64(min(len(l)-n, 40)))
			add(mapmodel.MoverCrowd, [][]mapmodel.Pt{l[at : at+n]}, j, false, 4)
		}
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// loop is the news helicopter's circuit: an ellipse round a point near the
// square (the story it is covering), as a closed run of cells.
func (s *scene) loop(j int) []mapmodel.Pt {
	w := s.w
	cx := w.CX + int(mapmodel.Hash(w.Seed, 363, int64(j))%21) - 10
	cy := w.CY + int(mapmodel.Hash(w.Seed, 364, int64(j))%9) - 4
	rx, ry := 9.0+float64(5*j), 4.0+float64(2*j)
	var out []mapmodel.Pt
	for k := 0; k < 96; k++ {
		a := float64(k) / 96
		p := pt(cx+int(float64(mapmodel.Cos(a)*rx)), cy+int(float64(mapmodel.Sin(a)*ry)))
		if p.X < 0 || p.Y < 0 || p.X >= w.W || p.Y >= w.H {
			continue
		}
		if n := len(out); n == 0 || out[n-1] != p {
			out = append(out, p)
		}
	}
	return out
}

// alleyLanes are lanes through the megacity's alleys and the town's
// streets round the square (found once per scene).
func (s *scene) alleyLanes() [][]mapmodel.Pt {
	cl := s.city
	if cl.lanesDone {
		return cl.lanes
	}
	cl.lanesDone = true
	w := s.w
	pass := func(x, y int) bool {
		i := y*w.W + x
		switch s.cells[i].k {
		case kStreet, kRoad, kBridge, kPlaza:
			return true
		case kNone:
			u := cl.use[i]
			return u == uAlley || u == uTrash
		}
		return false
	}
	cl.lanes = s.lanesOn(w.CX-60, w.CY-21, w.CX+60, w.CY+21, pass, 16, 10, 360)
	return cl.lanes
}

// putFX draws one effect cell over the map, keeping its background, never
// over a building, a wonder, a civ or the square, nor in the fog.
func (v *view) putFX(cv *mapstyle.Canvas, p mapmodel.Pt, r rune, fg tcell.Color, attr tcell.AttrMask) {
	s := v.sc
	k := s.at(p.X, p.Y).k
	cx, cy, ok := v.g.cellOf(p)
	if !ok || s.seen(p) < 2 || k == kTile || k == kWonder || k == kSite || k == kCentre {
		return
	}
	_, cur := cv.Get(cx, cy)
	_, bg, _ := cur.Decompose()
	cv.Put(cx, cy, r, v.style(glyph{r: r, fg: fg, bg: bg, attr: attr}))
}

// drawCityFX draws the city's ambient effects for this frame: headlights
// streaming down the highways at night, steam from the megacity's vents
// and its acid rain. The data glow, the flickering screens and neon and
// the plasma pulse are drawn with their cells.
func (v *view) drawCityFX(cv *mapstyle.Canvas) {
	s := v.sc
	cl := s.city
	if cl == nil || v.pal.city == nil || v.compact {
		return
	}
	cp, f := v.pal.city, v.anim
	if s.m.Clock.Night > 0.5 {
		for li, ln := range cl.lines {
			if ln.feat != mapmodel.FeatHighway {
				continue
			}
			n := len(ln.pts)
			for j := 0; j < 6; j++ {
				ink, step := cp.fg[theme.CityHeadlight], f/2+j*n/6+li*5
				if j%2 == 1 { // tail lights, coming back in
					ink, step = cp.fg[theme.CityTaillight], n*8-f/2+j*n/6
				}
				v.putFX(cv, ln.pts[(step%n+n)%n], '·', ink, tcell.AttrBold)
			}
		}
	}
	for k, p := range cl.vents {
		ph := (f/3 + int(mapmodel.Hash(int64(k), 365)%9)) % 9
		if ph >= 4 {
			continue // the vent rests between breaths
		}
		v.putFX(cv, pt(p.X+ph/3, p.Y-1-ph/2), [4]rune{'░', '░', '·', '·'}[ph], cp.fg[theme.CitySteam], 0)
	}
	if cl.look.Rain {
		g := v.g
		drops := g.w * g.h / 60
		for k := 0; k < drops; k++ {
			y := int((mapmodel.Hash(int64(k), 366)%uint64(g.h+8) + uint64(f)) % uint64(g.h+2))
			x := int(mapmodel.Hash(int64(k), 367)%uint64(g.w+g.h)) - y/2
			if x < 0 || x >= g.w || y >= g.h {
				continue
			}
			p := pt(g.vx+x/g.cellW, g.vy+y)
			if g.zoom == zRegion {
				continue
			}
			v.putFX(cv, p, '╱', cp.fg[theme.CityAcid], 0)
		}
	}
}
