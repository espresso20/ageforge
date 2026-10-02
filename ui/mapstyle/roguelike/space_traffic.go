package roguelike

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_traffic.go is the sky scenes' traffic, from the shared roster
// (mapmodel.MoversAt): climbers on the tether, mining drones off the belt,
// shuttles and satellites in the Space Age; generation ships trailing their
// engines in the Interstellar; starships jumping to warp and the alien
// saucers, ordinary traffic now, in the Galactic; ships that tunnel in the
// Quantum; a mote of light in the Transcendent. A scene lays the lanes once;
// a frame only indexes them. Nothing moves over a building, a civ or a
// wonder, and nothing shows before the age that introduces it.

// skyMoverKind is one kind of sky traffic.
type skyMoverKind uint8

const (
	smClimber skyMoverKind = iota
	smMiningDrone
	smShuttle
	smSatellite
	smGenShip
	smStarship
	smAlien
	smPhase
	smMote
	numSkyMoverKinds
)

// skyMoverKinds ties each kind to its roster entry.
var skyMoverKinds = [numSkyMoverKinds]mapmodel.Mover{mapmodel.MoverClimber, mapmodel.MoverMiningDrone,
	mapmodel.MoverShuttle, mapmodel.MoverSatellite, mapmodel.MoverGenShip, mapmodel.MoverStarship,
	mapmodel.MoverAlienShip, mapmodel.MoverPhaseShip, mapmodel.MoverMote}

// skyMover is one moving thing and the lane it runs.
type skyMover struct {
	kind  skyMoverKind
	info  mapmodel.MoverInfo
	lane  []mapmodel.Pt
	once  bool // one way, then rest out of sight; else there and back
	dwell int  // steps at rest at each end, or between runs
	off   int  // phase, in steps
	n     int  // its number among its kind (picks its inspect line)
	trail int  // cells of engine trail behind it
	alien int  // species, for alien ships
	warp  int  // a starship jumps to warp this many cells before its lane ends
}

// skyMoverAt is where a mover is this frame.
type skyMoverAt struct {
	mv   *skyMover
	i    int // index on the lane; -1 while it rests
	dir  int // +1 forward along the lane, -1 back
	warp bool
}

// addMover adds a mover of kind k when the age has it, with a phase from
// the seed.
func (s *skyScene) addMover(k skyMoverKind, lane []mapmodel.Pt, n int, once bool, dwell int) *skyMover {
	info := skyMoverKinds[k].Info()
	if len(lane) < 2 || !info.In(s.m.AgeIdx) {
		return nil
	}
	mv := skyMover{kind: k, info: info, lane: lane, once: once, dwell: dwell, n: n}
	if p := mv.period(); p > 0 {
		mv.off = int(mapmodel.Hash(s.b.seed, 610, int64(k), int64(n)) % uint64(p))
	}
	s.movers = append(s.movers, mv)
	return &s.movers[len(s.movers)-1]
}

// period is the mover's cycle in steps (a step is info.Pace frames).
func (mv *skyMover) period() int {
	if mv.once {
		return len(mv.lane) + mv.dwell
	}
	return 2 * (len(mv.lane) - 1 + mv.dwell)
}

// where is the mover's place at frame f.
func (mv *skyMover) where(f int) skyMoverAt {
	at := skyMoverAt{mv: mv, i: -1, dir: 1}
	per := mv.period()
	if per <= 0 {
		return at
	}
	t := (f/mv.info.Pace + mv.off) % per
	n := len(mv.lane)
	if mv.once {
		if t < n {
			at.i = t
			at.warp = mv.warp > 0 && t >= n-mv.warp
		}
		return at
	}
	span := n - 1
	switch {
	case t < mv.dwell:
		at.i = 0
	case t < mv.dwell+span:
		at.i = t - mv.dwell
	case t < 2*mv.dwell+span:
		at.i, at.dir = span, -1
	default:
		at.i, at.dir = span-(t-2*mv.dwell-span), -1
	}
	return at
}

// blocks reports a cell traffic never covers: a building, the hub, a civ,
// a wonder, a mark of the mandala.
func blocks(c skyCell) bool {
	switch c.k {
	case skUnit, skHub, skCiv, skWonder, skMark, skCore:
		return true
	}
	return false
}

// drawSkyTraffic draws every mover at this frame and notes where each is.
func (v *skyView) drawSkyTraffic(cv *mapstyle.Canvas) {
	s, f := v.sc, v.anim
	v.movers = v.movers[:0]
	for i := range s.movers {
		mv := &s.movers[i]
		at := mv.where(f)
		v.movers = append(v.movers, at)
		if at.i < 0 {
			continue
		}
		v.drawMover(cv, s, at)
	}
}

// put draws one traffic cell, keeping the cell's background, never over
// anything traffic must not cover.
func (v *skyView) putMover(cv *mapstyle.Canvas, s *skyScene, p mapmodel.Pt, r rune, ink mapmodel.SkyInk, lv uint8, bold bool) bool {
	c := s.at(p.X, p.Y)
	if blocks(c) {
		return false
	}
	cx, cy, ok := v.sg.cellOf(p)
	if !ok {
		return false
	}
	_, cur := cv.Get(cx, cy)
	_, bg, _ := cur.Decompose()
	cv.Put(cx, cy, r, v.style(look{r: r, ink: ink, lv: lv, bgc: bg, bold: bold}))
	return true
}

// drawMover draws one mover and claims its legend row.
func (v *skyView) drawMover(cv *mapstyle.Canvas, s *skyScene, at skyMoverAt) {
	mv := at.mv
	p := mv.lane[at.i]
	head := mapmodel.R(mv.info.Sym, v.tier)
	ink, lv := mapmodel.InkStarBright, uint8(3)
	east := at.dir > 0 && laneEast(mv.lane) || at.dir < 0 && !laneEast(mv.lane)
	switch mv.kind {
	case smClimber:
		ink = mapmodel.InkLight
	case smMiningDrone:
		ink = mapmodel.InkRock
	case smShuttle:
		ink = mapmodel.InkFrame
		for k, r := range []rune{'░', '·'} {
			if j := at.i - 1 - k; j >= 0 {
				v.putMover(cv, s, mv.lane[j], r, mapmodel.InkLight, 2, false)
			}
		}
	case smSatellite:
		ink = mapmodel.InkStarBright
	case smGenShip:
		ink = mapmodel.InkMoon
		if v.tier != mapmodel.TierNerd && !east {
			head = '◄'
		}
		for k := 1; k <= mv.trail; k++ {
			j := at.i - k
			if j < 0 {
				break
			}
			r, tink, tlv := '─', mapmodel.InkGlow, uint8(2)
			switch {
			case k <= 3:
				r, tlv = '═', 3
			case k > mv.trail*2/3:
				r, tink, tlv = '·', mapmodel.InkEcho, 1
			case k > mv.trail/3:
				tink, tlv = mapmodel.InkEcho, 2
			}
			v.putMover(cv, s, mv.lane[j], r, tink, tlv, false)
		}
	case smStarship:
		ink = mapmodel.InkFrame
		if at.warp {
			for k := 1; k <= 10; k++ {
				if j := at.i - k; j >= 0 {
					tlv := uint8(3)
					if k > 5 {
						tlv = 1
					}
					v.putMover(cv, s, mv.lane[j], '─', mapmodel.InkGlow, tlv, false)
				}
			}
			if at.i >= len(mv.lane)-2 {
				head, ink = '✦', mapmodel.InkGlow
			}
		} else if j := at.i - 1; j >= 0 {
			v.putMover(cv, s, mv.lane[j], '═', mapmodel.InkAccent, 2, false)
		}
	case smAlien:
		ink = inkAlien + mapmodel.SkyInk(mv.alien%len(mapmodel.Aliens))
		v.putMover(cv, s, pt(p.X-1, p.Y), '◄', ink, 2, false)
		v.putMover(cv, s, pt(p.X+1, p.Y), '►', ink, 2, false)
	case smPhase:
		ink = inkIri + mapmodel.SkyInk((p.X+v.anim/4)%8)
		if (v.anim/4+mv.n)%3 == 2 {
			head = 0 // between possibilities
		}
		for _, back := range []int{3, 6} {
			if j := at.i - back; j >= 0 {
				v.putMover(cv, s, mv.lane[j], '·', mapmodel.InkEcho, 2, false)
			}
		}
	case smMote:
		ink, lv = mapmodel.InkStarBright, 3
	}
	if head != 0 {
		v.putMover(cv, s, p, head, ink, lv, true)
	}
	reg := head
	if reg == 0 || mv.kind == smGenShip || mv.kind == smStarship {
		reg = mapmodel.R(mv.info.Sym, v.tier)
	}
	v.reg(slMover+skyLgID(mv.kind), look{r: reg, ink: ink, lv: 3, bold: true})
}

// laneEast reports a lane that runs west to east overall.
func laneEast(l []mapmodel.Pt) bool { return len(l) > 1 && l[len(l)-1].X >= l[0].X }

// moverAt is the mover on p this frame, topmost first.
func (v *skyView) moverAt(p mapmodel.Pt) (mapstyle.Inspection, bool) {
	s := v.sc
	if s == nil || blocks(s.at(p.X, p.Y)) {
		return mapstyle.Inspection{}, false
	}
	for i := len(s.movers) - 1; i >= 0; i-- {
		mv := &s.movers[i]
		at := mv.where(v.anim)
		if at.i < 0 {
			continue
		}
		q := mv.lane[at.i]
		hit := q == p
		if mv.kind == smAlien {
			hit = hit || pt(q.X-1, q.Y) == p || pt(q.X+1, q.Y) == p
		}
		if mv.kind == smPhase && (v.anim/4+mv.n)%3 == 2 {
			hit = false
		}
		if !hit {
			continue
		}
		in := mapstyle.Inspection{Title: mv.info.Title, Lines: []string{mv.info.Line(mv.n)}, Kind: mapstyle.KindMover}
		if mv.kind == smAlien {
			a := mapmodel.Aliens[mv.alien%len(mapmodel.Aliens)]
			in.Lines = []string{a.Lines[int(mapmodel.Hash(int64(mv.n), 7)%uint64(len(a.Lines)))]}
		}
		return in, true
	}
	return mapstyle.Inspection{}, false
}
