package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_traffic.go is what moves in the sky scenes, as dense as the real
// state behind it (trafficCounts: routes, people at work, soldiers,
// wealth): climbers on the tether, shuttles rising from the planet,
// satellites along the limb and mining drones from the belt in the Space
// Age; generation ships trailing their engines across the Interstellar;
// starships jumping to warp and alien saucers in the Galactic; phase ships
// tunnelling through the Quantum; a mote of light in the Transcendent.
// Every drawing names the mover it stands for in the shared roster
// (mapmodel/movers.go) and shows only once an age has introduced it: the
// no-spoilers rule, as the ground's moverTags keep it.

// skyTpl is a sky vehicle drawn facing east: rows of runes, an ink mask
// (f frame, d dim frame, l light, g glow, a b c the accents, s bright, r
// rock, e echo, u the species colour) and the mover it stands for (or
// MoverNone for the skyline's own drawings: debris, patrol craft).
type skyTpl struct {
	rows, mask []string
	mover      mapmodel.Mover
	sym        mapmodel.Sym // the Nerd Font mark at its middle
	emit       bool         // self-lit
}

var (
	stClimber   = skyTpl{[]string{"◆"}, []string{"l"}, mapmodel.MoverClimber, mapmodel.SymClimber, true}
	stShuttle   = skyTpl{[]string{"▴"}, []string{"f"}, mapmodel.MoverShuttle, mapmodel.SymShuttle, false}
	stSatellite = skyTpl{[]string{"─✧─"}, []string{"dsd"}, mapmodel.MoverSatellite, mapmodel.SymSatellite, true}
	stDrone     = skyTpl{[]string{"▫"}, []string{"d"}, mapmodel.MoverMiningDrone, mapmodel.SymMiningDrone, false}
	stDroneFull = skyTpl{[]string{"▪"}, []string{"r"}, mapmodel.MoverMiningDrone, mapmodel.SymMiningDrone, false}
	stPod       = skyTpl{[]string{"▪"}, []string{"l"}, mapmodel.MoverMaglev, mapmodel.SymMaglev, true}
	stPatrol    = skyTpl{[]string{"▐█►"}, []string{"dfc"}, mapmodel.MoverNone, mapmodel.SymMilitary, false}
	stTaxi      = skyTpl{[]string{"═►"}, []string{"bf"}, mapmodel.MoverHovercar, mapmodel.SymHovercar, false}
	stDebris    = skyTpl{[]string{"·"}, []string{"r"}, mapmodel.MoverNone, mapmodel.SymNone, false}
	stGenShip   = skyTpl{[]string{" ▄▄▄ ▄▄▄ ▄▄▄ ▄  ", "▓█████████████►", " ▀▀▀ ▀▀▀ ▀▀▀ ▀  "},
		[]string{" fff fff fff f  ", "gffffffffffffff", " ddd ddd ddd d  "}, mapmodel.MoverGenShip, mapmodel.SymGenShip, false}
	stHauler   = skyTpl{[]string{"▐▀▀►"}, []string{"gffd"}, mapmodel.MoverShuttle, mapmodel.SymShuttle, false}
	stHabitat  = skyTpl{[]string{"◎"}, []string{"f"}, mapmodel.MoverHabitat, mapmodel.SymOrbital, false}
	stStarship = skyTpl{[]string{"══╗ ▄▄██▄▄", "  ╚═▀▀▀▀▀▀"}, []string{"bbd ffffff", "  dddddddd"},
		mapmodel.MoverStarship, mapmodel.SymStarship, false}
	stAlien     = skyTpl{[]string{"◄◉►"}, []string{"uuu"}, mapmodel.MoverAlienShip, mapmodel.SymUFO, true}
	stGalShuttl = skyTpl{[]string{"▄►"}, []string{"fb"}, mapmodel.MoverShuttle, mapmodel.SymShuttle, false}
	stPhase     = skyTpl{[]string{"◈"}, []string{"g"}, mapmodel.MoverPhaseShip, mapmodel.SymPhaseShip, true}
	stMote      = skyTpl{[]string{"∘"}, []string{"s"}, mapmodel.MoverMote, mapmodel.SymMote, true}
)

// skyFx is a vehicle's effect.
type skyFx uint8

const (
	fxNone    skyFx = iota
	fxExhaust       // a rocket's exhaust under it
	fxTrail         // a long engine trail behind it
	fxWarp          // the jump to warp: a long streak ahead and a flash
	fxPhase         // afterimages where it was
	fxGlow          // a soft halo
)

// skyVehicle is one placed sky vehicle in frame terms.
type skyVehicle struct {
	t     *skyTpl
	x, y  int // screen column of its left edge, scene row of its bottom row
	depth uint8
	west  bool
	seed  int
	fx    skyFx
	n     int            // the effect's length (trail cells, streak cells, afterimage step)
	hue   theme.SpaceHue // the species colour of an alien
}

// skyGeo is the geometry the sky traffic needs: the scene, the frame and
// the baseline.
type skyGeo struct {
	m                 *mapmodel.Model
	sc                mapmodel.SkyScene
	anim, cam, W, S   int
	gy                int
	limb              []float64 // the Space Age planet's limb, half rows
	tether            int
	present, frontier int // world columns of the present district's centre and the frontier
}

func (o *orb) geo() skyGeo {
	g := skyGeo{m: o.m, sc: o.sc, anim: o.anim, cam: o.cam, W: o.W, S: o.S, gy: o.groundY,
		present: o.present.Centre, frontier: o.m.Skyline.FrontierX}
	if o.sc == mapmodel.SkyOrbit {
		g.limb = limbOf(o.W, o.S, o.groundY)
		g.tether = o.tetherX()
	}
	return g
}

// limbRow is the scene row of the planet's limb at screen column x.
func (g skyGeo) limbRow(x int) int {
	if len(g.limb) == 0 {
		return g.S
	}
	return min(g.S, int(g.limb[clampInt(x, 0, len(g.limb)-1)]/2))
}

// wrap places a stream anchored to the world on a screen w wide: its left
// column at frame anim, moving speed columns a frame.
func wrapX(seed, anim, cam, w int, speed float64) int {
	pad := 14
	span := w + 2*pad
	pos := float64(hash(seed, 23)%uint64(span)) + float64(anim)*speed - float64(cam)
	x := int(math.Floor(pos)) % span
	if x < 0 {
		x += span
	}
	return x - pad
}

// skyTrafficFor lists the sky vehicles of a frame. It is a pure function of
// the geometry.
func skyTrafficFor(g skyGeo) []skyVehicle {
	m := g.m
	if m == nil || g.W <= 0 || g.gy < 4 {
		return nil
	}
	c := trafficCounts(m)
	var out []skyVehicle
	add := func(v skyVehicle) {
		if v.t.mover != mapmodel.MoverNone && !mapmodel.Introduced(v.t.mover, m.AgeIdx) {
			return // not before its age
		}
		out = append(out, v)
	}
	sp := func(seed int, lo, hi float64) float64 { return lo + (hi-lo)*hashf(seed, 13) }
	dir := func(seed int) bool { return hash(seed, 17)%2 == 0 }
	move := func(seed int, speed float64) (int, bool) {
		west := dir(seed)
		if west {
			speed = -speed
		}
		return wrapX(seed, g.anim, g.cam, g.W, speed), west
	}
	gy := g.gy
	switch g.sc {
	case mapmodel.SkyOrbit:
		// climbers on the tether, up and down
		if x := g.tether; x >= 0 && x < g.W {
			foot := g.limbRow(x)
			L := foot + 6
			for i := 0; i < clampInt(2+c.foot/5, 2, 4); i++ {
				p := (g.anim/3 + i*L/3 + int(hash(i, 3)%7)) % L
				y := foot - p // going up
				if i%2 == 1 {
					y = p - 5 // coming down
				}
				if y >= 0 && y < foot {
					add(skyVehicle{t: &stClimber, x: x, y: y, depth: dTether, seed: i})
				}
			}
		}
		// shuttles: one per running route (and one besides), climbing
		// from the planet to the truss's underside
		for i := 0; i < 1+min(c.route, 6); i++ {
			seed := 1100 + i
			x := wrapX(seed, 0, g.cam, g.W, 0)
			T := 90 + int(hash(seed, 5)%60)
			t := float64((g.anim+int(hash(seed, 7)%uint64(T)))%T) / float64(T)
			if t > 0.9 {
				continue // docked
			}
			from := g.limbRow(x) - 1
			to := gy + 2
			y := from + int(math.Round(float64(to-from)*t))
			drift := int(t * 4)
			if dir(seed) {
				drift = -drift
			}
			add(skyVehicle{t: &stShuttle, x: x + drift, y: y, depth: dLane2, seed: seed, fx: fxExhaust, n: 3})
		}
		// satellites along the limb
		for i := 0; i < 3; i++ {
			seed := 1200 + i
			x, west := move(seed, sp(seed, 0.12, 0.3))
			y := g.limbRow(x+1) - 2 - i%2
			if y > gy+1 {
				add(skyVehicle{t: &stSatellite, x: x, y: y, depth: dLane2, west: west, seed: seed})
			}
		}
		// mining drones between the belt and the station
		n := 1
		for _, l := range m.SkyLineages() {
			if l.Key == mapmodel.LinMines {
				n = clampInt(1+len(l.Units)/6, 1, 4)
			}
		}
		b1 := max(1, gy/6) + max(2, gy/7)
		for i := 0; i < n; i++ {
			seed := 1300 + i
			x := wrapX(seed, 0, g.cam, g.W, 0)
			T := 120 + int(hash(seed, 5)%80)
			ph := (g.anim + int(hash(seed, 7)%uint64(T))) % T
			t := float64(ph) / float64(T/2)
			tpl := &stDrone
			if t > 1 {
				t = 2 - t // home again, loaded
				tpl = &stDroneFull
			}
			y := b1 + int(math.Round(float64(gy-2-b1)*t))
			xx := x + int(math.Round(6*mapmodel.Sin(t/2)))
			add(skyVehicle{t: tpl, x: xx, y: y, depth: dLane1, seed: seed})
		}
		// pods running along the truss, as many as are at work
		for i := 0; i < min(6, c.foot/2); i++ {
			seed := 1400 + i
			x, west := move(seed, sp(seed, 0.3, 0.6))
			add(skyVehicle{t: &stPod, x: x, y: gy, depth: dTruss, west: west, seed: seed})
		}
		// patrol craft and private flyers among the modules
		for i := 0; i < c.army+c.war; i++ {
			seed := 1500 + i
			x, west := move(seed, sp(seed, 0.2, 0.4))
			add(skyVehicle{t: &stPatrol, x: x, y: gy - 3 - int(hash(seed, 3)%uint64(max(1, gy/3))), depth: dLane0, west: west, seed: seed})
		}
		for i := 0; i < min(6, c.private); i++ {
			seed := 1600 + i
			x, west := move(seed, sp(seed, 0.4, 0.8))
			add(skyVehicle{t: &stTaxi, x: x, y: gy - 2 - int(hash(seed, 3)%uint64(max(1, gy/2))), depth: dLane1, west: west, seed: seed})
		}
		// slow debris far up
		for i := 0; i < 2; i++ {
			seed := 1700 + i
			x, west := move(seed, sp(seed, 0.02, 0.06))
			add(skyVehicle{t: &stDebris, x: x, y: 1 + int(hash(seed, 3)%uint64(max(1, gy/3))), depth: dAir, west: west, seed: seed})
		}
	case mapmodel.SkyDeep:
		// generation ships, big and slow, trailing their engines
		gens := clampInt(1+c.route/3, 1, 3)
		for i := 0; i < gens; i++ {
			seed := 2100 + i
			x, west := move(seed, sp(seed, 0.12, 0.2))
			y := gy*3/10 + i*max(3, gy/7) + int(hash(seed, 3)%3)
			add(skyVehicle{t: &stGenShip, x: x, y: y, depth: dTrail - 1, west: west, seed: seed, fx: fxTrail,
				n: 20 + int(hash(seed, 9)%21)})
		}
		// haulers between the colonies, one per route
		for i := 0; i < min(c.route, 6); i++ {
			seed := 2200 + i
			x, west := move(seed, sp(seed, 0.25, 0.45))
			add(skyVehicle{t: &stHauler, x: x, y: gy - 2 - int(hash(seed, 3)%uint64(max(1, gy/3))), depth: [3]uint8{dLane0, dLane1, dLane2}[i%3],
				west: west, seed: seed})
		}
		// mining drones still working the outer rocks
		for i := 0; i < 2; i++ {
			seed := 2300 + i
			x := wrapX(seed, 0, g.cam, g.W, 0)
			T := 140 + int(hash(seed, 5)%60)
			ph := (g.anim + int(hash(seed, 7)%uint64(T))) % T
			t := float64(ph) / float64(T/2)
			tpl := &stDrone
			if t > 1 {
				t, tpl = 2-t, &stDroneFull
			}
			y := 2 + int(math.Round(float64(gy-4)*t))
			add(skyVehicle{t: tpl, x: x + int(8*t), y: y, depth: dLane1, seed: seed})
		}
		// orbital habitats turning far off
		for i := 0; i < 2; i++ {
			seed := 2400 + i
			x, west := move(seed, 0.03)
			add(skyVehicle{t: &stHabitat, x: x, y: 2 + int(hash(seed, 3)%uint64(max(1, gy/3))), depth: dLane, west: west, seed: seed})
		}
		for i := 0; i < c.army+c.war; i++ {
			seed := 2500 + i
			x, west := move(seed, sp(seed, 0.2, 0.35))
			add(skyVehicle{t: &stPatrol, x: x, y: gy - 3 - int(hash(seed, 3)%uint64(max(1, gy/3))), depth: dLane0, west: west, seed: seed})
		}
	case mapmodel.SkyGalaxy:
		// starships: they cruise a while, then jump to warp
		ships := clampInt(2+c.route/2, 2, 6)
		for i := 0; i < ships; i++ {
			seed := 3100 + i
			T := 200 + int(hash(seed, 5)%140)
			cyc := (g.anim + int(hash(seed, 7)%uint64(T))) / T
			ph := (g.anim + int(hash(seed, 7)%uint64(T))) % T
			cruise := T * 7 / 10
			if ph >= cruise+8 {
				continue // gone to warp
			}
			cs := seed*31 + cyc
			west := dir(cs)
			x0 := int(hash(cs, 23)%uint64(g.W+20)) - 10
			step := ph / 4
			if west {
				step = -step
			}
			y := 2 + int(hash(cs, 3)%uint64(max(1, gy-6)))
			v := skyVehicle{t: &stStarship, x: x0 + step, y: y, depth: [2]uint8{dLane2, dLane}[i%2], west: west, seed: seed}
			if ph >= cruise {
				v.fx, v.n = fxWarp, ph-cruise
			}
			add(v)
		}
		// alien saucers: ordinary traffic now, each species in its colour
		aliens := clampInt(2+m.DiscoveredCount()/4, 2, 4)
		for i := 0; i < aliens; i++ {
			seed := 3200 + i
			x, west := move(seed, sp(seed, 0.15, 0.35))
			y := 3 + int(hash(seed, 3)%uint64(max(1, gy-6))) + (g.anim/12+i)%2
			add(skyVehicle{t: &stAlien, x: x, y: y, depth: dLane2, west: west, seed: seed,
				hue: mapmodel.Aliens[i%len(mapmodel.Aliens)].Hue})
		}
		// shuttles between the decks
		for i := 0; i < clampInt(c.foot/3, 1, 5); i++ {
			seed := 3300 + i
			x, west := move(seed, sp(seed, 0.3, 0.6))
			add(skyVehicle{t: &stGalShuttl, x: x, y: gy - 2 - int(hash(seed, 3)%uint64(max(1, gy/2))),
				depth: [3]uint8{dLane0, dLane1, dLane2}[i%3], west: west, seed: seed})
		}
		for i := 0; i < 1; i++ {
			seed := 3400 + i
			x, west := move(seed, 0.03)
			add(skyVehicle{t: &stHabitat, x: x, y: 2 + int(hash(seed, 3)%uint64(max(1, gy/3))), depth: dLane, west: west, seed: seed})
		}
	case mapmodel.SkyQuantum:
		// phase ships: seen only in bursts, jumping cells at a time
		ships := clampInt(3+c.route/2, 3, 7)
		for i := 0; i < ships; i++ {
			seed := 4100 + i
			J := 10 + int(hash(seed, 3)%8)
			D := 5 + int(hash(seed, 5)%6)
			k := g.anim / J
			if g.anim%J >= 5 {
				continue // between possibilities
			}
			west := dir(seed)
			step := k * D
			if west {
				step = -step
			}
			span := g.W + 28
			x := int(hash(seed, 23)%uint64(span)) + step - g.cam/2
			x = ((x%span)+span)%span - 14
			y := 2 + int(hash(seed, 9)%uint64(max(1, gy-4)))
			add(skyVehicle{t: &stPhase, x: x, y: y, depth: [3]uint8{dLane0, dLane1, dLane2}[i%3], west: west, seed: seed, fx: fxPhase, n: D})
		}
	case mapmodel.SkyMandala:
		// a mote of light or two, drifting home to the mandala
		for i := 0; i < 1+min(1, c.foot/6); i++ {
			seed := 5100 + i
			T := 900 + int(hash(seed, 3)%400)
			t := float64((g.anim+int(hash(seed, 5)%uint64(T)))%T) / float64(T)
			sx := float64(int(hash(seed, 7) % uint64(max(1, g.W))))
			cx, cy := float64(g.W)/2, float64(gy-1)*0.47
			x := sx + (cx-sx)*t + 3*mapmodel.Sin(t*2)
			y := float64(gy-1) + (cy-float64(gy-1))*t
			add(skyVehicle{t: &stMote, x: int(math.Round(x)), y: int(math.Round(y)), depth: dLane, seed: seed, fx: fxGlow})
		}
	}
	return out
}

// traffic draws the frame's sky vehicles and their effects.
func (o *orb) traffic() {
	if o.v.noTraffic {
		return
	}
	for _, v := range skyTrafficFor(o.geo()) {
		o.drawVehicle(v)
	}
}

// inkOf is a template mask letter's ink.
func inkOf(mk byte) mapmodel.SkyInk {
	switch mk {
	case 'd':
		return mapmodel.InkFrameDim
	case 'l':
		return mapmodel.InkLight
	case 'g':
		return mapmodel.InkGlow
	case 'a':
		return mapmodel.InkAccent
	case 'b':
		return mapmodel.InkAccent2
	case 'c':
		return mapmodel.InkAccent3
	case 's':
		return mapmodel.InkStarBright
	case 'r':
		return mapmodel.InkRock
	case 'e':
		return mapmodel.InkEcho
	}
	return mapmodel.InkFrame
}

func (o *orb) drawVehicle(v skyVehicle) {
	n := len(v.t.rows)
	w := 0
	for _, r := range v.t.rows {
		w = max(w, len([]rune(r)))
	}
	hz := 0
	if v.depth >= dAir {
		hz = 3
	} else if v.depth >= dLane2 {
		hz = 1
	}
	col := func(mk byte, x int) tcell.Color {
		if o.sc == mapmodel.SkyQuantum {
			return o.iri(x/2+o.anim/2+v.seed, iEmit, hz)
		}
		if mk == 'u' {
			return o.raw(v.hue, iEmit, 0)
		}
		mode := iLit
		if v.t.emit || mk == 'l' || mk == 'g' || mk == 'b' || mk == 'c' || mk == 's' {
			mode = iEmit
		}
		return o.c(inkOf(mk), mode, hz)
	}
	markX, markY := v.x+w/2, v.y
	for r, row := range v.t.rows {
		rs, ms := []rune(row), []rune(v.t.mask[r])
		for i := range rs {
			j, ch := i, rs[i]
			if v.west {
				j = len(rs) - 1 - i
				if m, ok := mirrorRunes[ch]; ok {
					ch = m
				}
			}
			if ch == ' ' {
				continue
			}
			mk := byte('f')
			if i < len(ms) {
				mk = byte(ms[i])
			}
			x, y := v.x+j, v.y-(n-1-r)
			switch {
			case o.tier == mapmodel.TierNerd && x == markX && y == markY && v.t.sym != mapmodel.SymNone:
				ch = mapmodel.R(v.t.sym, o.tier)
			case ch == '◉':
				ch = mapmodel.R(mapmodel.SymUFO, o.tier)
			case len(rs) == 1 && v.t.sym != mapmodel.SymNone && v.t != &stPod && v.t != &stDroneFull:
				ch = mapmodel.R(v.t.sym, o.tier)
			}
			o.fb.fg(x, o.Y(y), ch, col(mk, x), v.depth)
		}
	}
	void := o.c(mapmodel.InkVoid, iBack, 0)
	switch v.fx {
	case fxExhaust:
		flame := o.c(mapmodel.InkLight, iEmit, hz)
		glow := o.c(mapmodel.InkGlow, iEmit, hz)
		for k := 1; k <= v.n; k++ {
			c := theme.Mix(flame, glow, float64(k)/float64(v.n))
			c = theme.Mix(c, void, float64(k-1)/float64(v.n+1))
			o.fb.fg(v.x, o.Y(v.y+k), [3]rune{'▓', '▒', '░'}[min(2, k-1)], c, v.depth)
		}
	case fxTrail:
		head := o.c(mapmodel.InkGlow, iEmit, 1)
		mid := o.c(mapmodel.InkEcho, iEmit, 1)
		dx := -1
		sx := v.x - 1
		if v.west {
			dx, sx = 1, v.x+w
		}
		y := v.y - 1
		for k := 0; k < v.n; k++ {
			t := float64(k) / float64(v.n)
			var c tcell.Color
			if t < 0.4 {
				c = theme.Mix(head, mid, t/0.4)
			} else {
				c = theme.Mix(mid, void, (t-0.4)/0.6)
			}
			ch := '━'
			switch {
			case t > 0.7:
				if k%2 == 1 {
					continue
				}
				ch = '·'
			case t > 0.4:
				ch = '─'
			}
			o.fb.fg(sx+k*dx, o.Y(y), ch, c, dTrail)
		}
	case fxWarp:
		streak := o.c(mapmodel.InkGlow, iEmit, 0)
		white := o.c(mapmodel.InkStarBright, iEmit, 0)
		L := 6 + v.n*7
		dx := 1
		sx := v.x + w
		if v.west {
			dx, sx = -1, v.x-1
		}
		for k := 0; k < L; k++ {
			c := theme.Mix(white, streak, float64(k)/float64(L))
			o.fb.fg(sx+k*dx, o.Y(v.y), '━', c, dTrail)
			if k > L/3 {
				o.fb.fg(sx+k*dx, o.Y(v.y-1), '─', theme.Mix(streak, void, 0.5), dTrail)
			}
		}
		if v.n < 3 {
			fx := sx + L*dx
			o.fb.fg(fx, o.Y(v.y), '✦', white, dTrail-1)
			o.fb.fg(fx, o.Y(v.y-1), '·', white, dTrail-1)
			o.fb.fg(fx, o.Y(v.y+1), '·', white, dTrail-1)
		}
	case fxPhase:
		for k := 1; k <= 2; k++ {
			x := v.x - k*v.n
			if v.west {
				x = v.x + k*v.n
			}
			c := theme.Mix(o.iri(x/2+o.anim/2+v.seed, iEmit, 1), void, 0.35*float64(k)+0.2)
			o.fb.fg(x, o.Y(v.y), '◇', c, dTrail)
		}
	case fxGlow:
		h := theme.Mix(o.c(mapmodel.InkLight, iEmit, 0), void, 0.6)
		if (o.anim/6)%2 == 0 {
			o.fb.fg(v.x-1, o.Y(v.y), '·', h, v.depth)
			o.fb.fg(v.x+1, o.Y(v.y), '·', h, v.depth)
		} else {
			o.fb.fg(v.x, o.Y(v.y-1), '˙', h, v.depth)
		}
	}
}
