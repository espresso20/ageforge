package skyline

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// orbit.go is the sky arc on the skyline (mapmodel/sky.go): from the Space
// Age on the panorama leaves the ground and each age draws a scene of its
// own. The mechanics stay the panorama's: one district per age west to
// east, every lot a silhouette centred on its column and standing on the
// baseline in its depth row, the camera, the cursor and its Tab order, the
// ▼ markers, the flows tags, the minimap strip, the header and the status
// line. What changes per scene is the look: the backdrop, the baseline the
// lots stand on (a truss, a formation line, a docking ring, a time echo),
// the lots' sky forms (orbit_forms.go), the traffic
// (orbit_traffic.go), the palette (mapmodel.SkyPaletteOf, resolved here
// against the theme) and one ambient effect. The last scene keeps none of
// it: the Transcendent Age is the mandala alone (orbit_mandala.go).
//
// The sky keeps its layouts in slots of its own: the full view's in v.slay,
// the compact view's in v.sclay, and the ground skyline the Quantum Age's
// time echo needs in v.glay. Their keys are negative, so no sky layout is
// ever taken for a ground one.

// The sky scenes' own depths. Each signature feature has a depth no other
// scene uses, which is also how the tests check that no scene draws a later
// one's.
const (
	dTruss     uint8 = 14 // the Space Age truss
	dFormation uint8 = 16 // the Interstellar formation line
	dDockRing  uint8 = 17 // the Galactic docking ring
	dEchoLine  uint8 = 18 // the Quantum time echo's ground line
	dFrontier  uint8 = 21 // construction on the frontier
	dHaloQ     uint8 = 42 // the Quantum probability clouds
	dGate      uint8 = 46 // the Interstellar warp gate
	dHub       uint8 = 47 // the Galactic starbase hub
	dFractal   uint8 = 49 // the Quantum fractal landmark
	dTether    uint8 = 50 // the Space Age tether
	dFringe    uint8 = 53 // the Quantum interference fringes
	dEcho      uint8 = 58 // the Quantum time echo
	dTrail     uint8 = 74 // engine trails, warp streaks, afterimages
	dLane      uint8 = 76 // sky traffic between the far objects and the lots
	dColony    uint8 = 82 // Interstellar colony worlds
	dSystem    uint8 = 83 // Galactic star systems and their lanes
	dMoon      uint8 = 84 // the Space Age moon
	dOrrFront  uint8 = 91 // the Interstellar orrery: its planets on the near side,
	dOrrSun    uint8 = 92 // its sun,
	dOrrBack   uint8 = 93 // its planets on the far side
	dOrrery    uint8 = 94 // and its corona and orbits
	dMandala   uint8 = 87 // the Transcendent mandala
	dBelt      uint8 = 89 // the Space Age asteroid belt
	dPlanet    uint8 = 90 // the Space Age planet
	dSpiral    uint8 = 96 // the Galactic spiral
	dNebula    uint8 = 97 // the Interstellar nebula
	dGalNebula uint8 = 98 // the Galactic nebulae
)

// How an ink is lit.
const (
	iLit  = iota // a structure: hazed toward the void with depth, ink on a light page
	iEmit        // self-lit: hazed less, deepened on a light page so it still reads
	iBack        // the backdrop: the void's family, lifted toward a light page
	numInkModes
)

// orb is one sky frame: the ground scene's state (camera, layout, cursor,
// markers, chrome) plus the scene's resolved palette.
type orb struct {
	*scene
	sc      mapmodel.SkyScene
	pal     mapmodel.SkyPalette
	light   bool        // a light page
	page    tcell.Color // the page colour
	voidRaw tcell.Color // the scene's void on a dark page
	voidC   tcell.Color // the void as this page shows it
	ink     [mapmodel.NumSkyInks][numInkModes][numHaze]tcell.Color
	lc      [4 * 3 * int(numSlots) * numHaze]tcell.Color // lot colours (dim, tone, slot, depth), lazily
	lcOK    [4 * 3 * int(numSlots) * numHaze]bool
	present mapmodel.District // the newest district
	x0, x1  int               // world columns the baseline spans
	iriTab  []tcell.Color     // the Quantum ramp, resolved (orbit_quantum.go)
	iq      []tcell.Color     // the Quantum slot colours, lazily
	pearl   tcell.Color       // the Quantum's pale sheen
}

// skyGround is the baseline row of a sky scene of S rows: raised from the
// ground's S-3 so there is room below for the planet, a reflection or the
// void.
func skyGround(S int) int {
	return clampInt(int(float64(S)*0.58+0.5), 6, max(6, S-4))
}

// skyScale sizes sky forms to the scene: a little larger than the ground's
// scale for the same baseline, since the baseline sits higher.
func skyScale(groundY int) float64 {
	s := float64(groundY-1) / 30
	if s < 0.42 {
		s = 0.42
	}
	if s > 1.25 {
		s = 1.25
	}
	return s
}

// composeSky builds a sky frame into v.fb. It mirrors compose: the same
// camera, cursor and chrome, the scene's own layers.
func (v *view) composeSky(f mapstyle.Frame, W, H int, sc mapmodel.SkyScene) *scene {
	if sc == mapmodel.SkyMandala {
		return v.composeMandala(f, W, H) // no panorama: the mandala alone
	}
	m := f.Model
	v.lastW = W
	s := &scene{v: v, m: m, fb: &v.fb, tier: f.Tier, anim: f.Anim, W: W, H: H, S: H - 3, top: 1, sel: -1}
	s.groundY = skyGround(s.S)
	s.band = bandOf(m.AgeIdx)
	s.p, s.mp = v.palettes(m)
	if sc == mapmodel.SkyQuantum {
		v.groundBehind(m, s.groundY)
	}
	s.lay = v.skyLayoutFor(m, sc, s.groundY, false)
	s.vis = v.vis
	defer func() { v.vis = s.vis }()
	if v.follow {
		v.cam = presentCam(m, W)
	}
	if v.inspect {
		ts := v.targetsFor(m, W, v.cam)
		if i := v.resolve(m, ts, f.Anim); i >= 0 && v.reveal {
			v.revealTarget(m, ts[i], W)
		}
	}
	v.cam = clampInt(v.cam, 0, max(0, m.Skyline.Width-W))
	s.cam = v.cam
	s.ridge = ridgeItems(m, W, s.cam)
	if v.inspect && v.cur.kind == tLot {
		s.sel = s.lay.find(v.cur.key, v.cur.cp)
	}
	if ep := m.Catastrophe.Pending; ep != "" {
		e := m.Catalog.EpochIdx[ep]
		s.gridOut = e == 3 || e == 4 || e == 5
	}
	v.fb.reset(W, H)
	o := newOrb(s, sc)
	s.vis = s.lay.visible(s.cam-16, s.cam+s.W+16, s.vis)
	o.void()
	switch sc {
	case mapmodel.SkyOrbit:
		o.spaceBackdrop()
	case mapmodel.SkyDeep:
		o.deepBackdrop()
	case mapmodel.SkyGalaxy:
		o.galaxyBackdrop()
	case mapmodel.SkyQuantum:
		o.quantumBackdrop()
	}
	o.farObjects()
	o.lots()
	o.frontier()
	o.traffic()
	if sc == mapmodel.SkyOrbit || sc == mapmodel.SkyDeep {
		s.visitor()
	}
	if sc == mapmodel.SkyGalaxy {
		o.sparkles()
	}
	o.catastrophe()
	o.farCursorRows()
	s.markers()
	o.chrome()
	return s
}

// newOrb resolves the scene's palette against the theme for this frame.
func newOrb(s *scene, sc mapmodel.SkyScene) *orb {
	o := &orb{scene: s, sc: sc, pal: mapmodel.SkyPaletteOf(sc), light: s.p.light,
		page: theme.Color(theme.RoleBackground)}
	o.voidRaw = theme.SpaceColor(o.pal[mapmodel.InkVoid])
	o.voidC = o.voidRaw
	if o.light {
		o.voidC = theme.Mix(o.voidRaw, o.page, 0.84)
	}
	for k := mapmodel.SkyInk(0); k < mapmodel.NumSkyInks; k++ {
		c := theme.SpaceColor(o.pal[k])
		for mode := 0; mode < numInkModes; mode++ {
			for d := 0; d < numHaze; d++ {
				o.ink[k][mode][d] = o.resolve(c, mode, d)
			}
		}
	}
	o.pearl = o.resolve(theme.Mix(theme.SpaceColor(o.pal[mapmodel.InkStarBright]), theme.SpaceColor(o.pal[mapmodel.InkHaze]), 0.35), iEmit, 1)
	if n := len(s.m.Skyline.Districts); n > 0 {
		o.present = s.m.Skyline.Districts[n-1]
		o.x0 = s.m.Skyline.Districts[0].X0 - 6
	}
	o.x1 = s.m.Skyline.FrontierX + 4 + 4*7 + 6
	return o
}

// resolve turns a raw sky colour into what this frame draws: hazed by depth
// row toward the void, mapped onto a light page (the void lifted toward the
// page so a light theme never gets a black hole, structures kept as ink,
// lights deepened so they read) and folded to the theme's duotone.
func (o *orb) resolve(c tcell.Color, mode, depth int) tcell.Color {
	depth = clampInt(depth, 0, numHaze-1)
	k := hazeK[depth]
	if o.light {
		switch mode {
		case iBack:
			c = theme.Mix(c, o.page, 0.72)
		case iLit:
			if l := luma(c); l > 92 {
				c = theme.Shade(c, 92/l)
			}
		default:
			if l, lim := luma(c), luma(o.page)-96; l > lim && lim > 0 {
				c = theme.Shade(c, lim/l)
			}
		}
	}
	switch mode {
	case iLit:
		c = theme.Mix(c, o.voidC, k)
	case iEmit:
		c = theme.Mix(c, o.voidC, float64(k*0.55))
	}
	return o.p.final(c)
}

// c is ink k lit in mode at a depth row.
func (o *orb) c(k mapmodel.SkyInk, mode, depth int) tcell.Color {
	return o.ink[k][mode][clampInt(depth, 0, numHaze-1)]
}

// raw resolves an arbitrary sky hue (an alien's colour, an iridescent step).
func (o *orb) raw(h theme.SpaceHue, mode, depth int) tcell.Color {
	return o.resolve(theme.SpaceColor(h), mode, depth)
}

// void fills the scene rows with the void.
func (o *orb) void() {
	c := o.c(mapmodel.InkVoid, iBack, 0)
	for y := 0; y < o.S; y++ {
		for x := 0; x < o.W; x++ {
			o.fb.set(x, o.Y(y), ' ', c, c, dSky)
		}
	}
}

// stars scatters a hash-based starfield over rows [y0, y1), a few
// twinkling each frame; par is the parallax and density the share of
// cells lit.
func (o *orb) stars(y0, y1 int, density, par float64, salt int) {
	shift := int(float64(o.cam) * par)
	dim, bright := o.c(mapmodel.InkStar, iEmit, 0), o.c(mapmodel.InkStarBright, iEmit, 0)
	for y := max(0, y0); y < min(y1, o.S); y++ {
		for x := 0; x < o.W; x++ {
			sx := x + shift
			hv := hashf(sx, y, salt)
			if hv >= density {
				continue
			}
			u := hv / density
			ch, col := '·', theme.Mix(o.c(mapmodel.InkVoid, iBack, 0), dim, 0.35+0.65*hashf(sx, y, salt+1))
			switch {
			case u < 0.04:
				ch, col = '✦', bright
			case u < 0.12:
				ch, col = '+', theme.Mix(dim, bright, 0.5)
			case u < 0.3:
				ch, col = '∙', dim
			}
			switch hash(sx, y, o.anim/5) % 13 {
			case 0: // a twinkle at its peak
				col = bright
				if ch == '·' {
					ch = '∙'
				} else if ch == '+' {
					ch = '*'
				}
			case 1:
				col = theme.Mix(col, o.c(mapmodel.InkVoid, iBack, 0), 0.6)
			}
			o.fb.fg(x, o.Y(y), ch, col, dStar)
		}
	}
}

// ------------------------------------------------------------ far objects

// farRow is the scene row of ridge item i's far object.
func (o *orb) farRow(it ridgeItem) int {
	gy := o.groundY
	base := gy * 3 / 10
	if o.sc == mapmodel.SkyGalaxy {
		base = gy / 5
	}
	return clampInt(base+int(hash(it.i, 19)%uint64(max(1, gy/6))), 2, max(2, gy-6))
}

// farObjects draws the civs met as far objects at their ridge columns, each
// with a pennant in the colour of your standing and its name, and the
// harbinger as a hologram. The Galactic Age draws the civs as star systems
// joined by trade lanes instead.
func (o *orb) farObjects() {
	if o.sc == mapmodel.SkyGalaxy {
		o.starSystems()
	}
	for _, it := range o.ridge {
		if it.fac == nil {
			o.hologram(it)
			continue
		}
		if o.sc == mapmodel.SkyGalaxy {
			continue
		}
		o.farStation(it)
	}
}

// farStation is a civ far off: a small station, world or ring, by scene,
// with its pennant and name.
func (o *orb) farStation(it ridgeItem) {
	f := it.fac
	y := o.farRow(it)
	x := it.x
	d := dRidgeTown
	frame := o.c(mapmodel.InkFrame, iLit, 3)
	dim := o.c(mapmodel.InkFrameDim, iLit, 3)
	lit := o.c(mapmodel.InkLight, iEmit, 2)
	put := func(dx, dy int, ch rune, c tcell.Color) { o.fb.fg(x+dx, o.Y(y+dy), ch, c, d) }
	w := clampInt(2+f.Strength, 3, 7)
	switch o.sc {
	case mapmodel.SkyOrbit: // a smaller station: a module bar on a spine
		for dx := -w / 2; dx <= w/2; dx++ {
			put(dx, 0, '▀', frame)
		}
		put(-w/2-1, 0, '▐', dim)
		put(w/2+1, 0, '▌', dim)
		if hash(it.i, o.anim/40)%2 == 0 {
			put(0, 0, '▪', lit)
		}
	case mapmodel.SkyDeep: // a colony world with a ring
		sf := o.c([3]mapmodel.SkyInk{mapmodel.InkSurface, mapmodel.InkSurface2, mapmodel.InkRock}[it.i%3], iBack, 1)
		put(-1, 0, '▐', sf)
		o.fb.fill(x, o.Y(y), sf, d)
		put(1, 0, '▌', sf)
		put(-2, 0, '─', dim)
		put(2, 0, '─', dim)
	case mapmodel.SkyQuantum: // a ring that may or may not be there
		ic := o.iri(it.x+o.anim/4, iEmit, 2)
		if (o.anim/6+it.i)%5 != 0 {
			put(-1, 0, '(', ic)
			put(0, 0, '◇', ic)
			put(1, 0, ')', ic)
		}
	}
	// the pennant
	rc := o.mp.Fg[mapmodel.RelationClass(f.Relation)]
	put(0, -1, '│', dim)
	flag := '►'
	if f.Relation == mapmodel.RelWar {
		flag = mapmodel.R(mapmodel.SymWar, o.tier)
	} else if o.tier == mapmodel.TierNerd {
		flag = mapmodel.R(mapmodel.SymCiv, o.tier)
	}
	put(1, -1, flag, rc)
	o.labels = append(o.labels, label{x - textLen(f.Name)/2, y - 2, f.Name, rc})
}

// hologram is the harbinger: a flickering figure far off, facing the
// station.
func (o *orb) hologram(it ridgeItem) {
	h := o.m.Harbinger
	if h == nil {
		return
	}
	x, y := it.x, o.farRow(it)+1
	ink := o.p.hue(theme.SkyHarbingerHolo, mEmit, 1)
	if hash(o.anim/2, 3)%5 == 0 {
		ink = theme.Mix(ink, o.fb.showAt(clampInt(x, 0, max(0, o.W-1)), o.Y(max(0, y))), 0.6)
	}
	d := dHarbinger
	o.fb.fg(x, o.Y(y-2), '▄', ink, d)
	o.fb.fg(x-1, o.Y(y-1), '▐', ink, d)
	o.fb.fill(x, o.Y(y-1), ink, d)
	o.fb.fg(x+1, o.Y(y-1), '▌', ink, d)
	o.fb.fg(x-1, o.Y(y), '▐', ink, d)
	o.fb.fg(x, o.Y(y), '▒', ink, d)
	o.fb.fg(x+1, o.Y(y), '▌', ink, d)
	for dy := 0; dy < 3; dy++ {
		o.fb.fg(x+2, o.Y(y-dy), '│', ink, d)
	}
	o.fb.fg(x, o.Y(y+1), '▀', theme.Mix(ink, o.c(mapmodel.InkVoid, iBack, 0), 0.6), d)
	if o.tier == mapmodel.TierNerd {
		o.fb.fg(x, o.Y(y-3), mapmodel.R(mapmodel.SymHarbinger, o.tier), o.mp.Fg[mapmodel.CDanger], d)
	}
	o.labels = append(o.labels, label{x - textLen(h.Name)/2, y - 4, h.Name, o.mp.Fg[mapmodel.CIdle]})
}

// farCursorRows makes the cursor's ▼ land just above each far object (the
// ground cursor puts it six rows above the ridge top at the item's column).
func (o *orb) farCursorRows() {
	tops := make([]float64, o.W)
	for i := range tops {
		tops[i] = float64(o.groundY - 2)
	}
	for _, it := range o.ridge {
		if it.x < 0 || it.x >= o.W {
			continue
		}
		top := o.farRow(it) - 2 // over a civ's name
		if it.fac == nil {
			top-- // over the harbinger's name, which sits a row higher
		}
		tops[it.x] = float64(top + 5)
	}
	o.farTops = tops
}

// ------------------------------------------------------------ lots

// lotCol is a lot's colour for a sprite slot: the scene's ink for the slot
// (orbit_forms.go), varied by the lot's tone and dimmed by its age.
func (o *orb) lotCol(lv *lotView, sl slot, depth int) tcell.Color {
	dim, tone := clampInt(lv.fam, 0, 3), clampInt(lv.vr, 0, 2)
	depth = clampInt(depth, 0, numHaze-1)
	i := ((dim*3+tone)*int(numSlots)+int(sl))*numHaze + depth
	if o.lcOK[i] {
		return o.lc[i]
	}
	sp := skySlots[o.sc][sl]
	if sp.ink == 0 && sp.k == 0 {
		sp = skySlots[o.sc][sWall]
	}
	c := theme.Shade(theme.SpaceColor(o.pal[sp.ink]), sp.k)
	if sp.mode != iEmit {
		switch tone {
		case 1:
			c = theme.Mix(c, theme.SpaceColor(o.pal[mapmodel.InkFrameDim]), 0.16)
		case 2:
			c = theme.Mix(c, theme.SpaceColor(o.pal[mapmodel.InkAccent]), 0.07)
		}
	}
	if dim > 0 { // older sections: weathered, a step toward the void
		a := [4]float64{0, 0.14, 0.26, 0.42}[dim]
		if sp.mode == iEmit {
			a *= 0.5
		}
		c = theme.Mix(c, o.voidRaw, a)
	}
	c = o.resolve(c, int(sp.mode), depth)
	o.lc[i], o.lcOK[i] = c, true
	return c
}

// lots draws the visible lots in their sky forms, back rows first: windows
// lit by staffing, beacons blinking, and on producers glowing vents where a
// chimney would smoke on the ground (there is no smoke in a vacuum).
func (o *orb) lots() {
	if o.sc == mapmodel.SkyQuantum {
		o.quantumLots()
		return
	}
	white := o.c(mapmodel.InkStarBright, iEmit, 0)
	for _, li := range o.vis {
		lv := &o.lay.lots[li]
		o.drawSprite(lv, lv.spr, func(sx, sy int, sl slot, hz int) tcell.Color {
			return o.lotCol(lv, sl, hz)
		}, white, li == o.sel)
		o.lotLights(lv, lv.spr)
	}
}

// drawSprite draws one lot's sprite with colour fn, windows lit by
// staffing, the flows overlay's dimming and the cursor's highlight.
func (o *orb) drawSprite(lv *lotView, spr *sprite, col func(sx, sy int, sl slot, hz int) tcell.Color,
	white tcell.Color, sel bool) {
	x0 := lv.ml.X - spr.w/2 - o.cam
	y0 := o.groundY - spr.h
	hz := lv.ml.Row
	dim := o.v.flows && flagged(lv)
	warm := o.c(mapmodel.InkLight, iEmit, hz)
	pale := theme.Mix(warm, white, 0.45)
	for sy := 0; sy < spr.h; sy++ {
		Y := o.Y(y0 + sy)
		for sx := 0; sx < spr.w; sx++ {
			c := spr.c[sy*spr.w+sx]
			X := x0 + sx
			if c.ch == 0 || X < 0 || X >= o.W {
				continue
			}
			fg := col(sx, sy, c.fg, hz)
			if c.fg == sWin && o.windowLit(lv, sx, sy) {
				fg = warm
				if hash(int(lv.ml.Seed), sx, sy)%4 == 0 {
					fg = pale
				}
			}
			if dim {
				fg = theme.Shade(fg, 0.55)
			}
			if sel {
				fg = theme.Mix(fg, white, 0.3)
			}
			if c.bg == sNone {
				if c.ch == '█' {
					o.fb.fill(X, Y, fg, lv.depth)
				} else {
					o.fb.fg(X, Y, c.ch, fg, lv.depth)
				}
				continue
			}
			bg := col(sx, sy, c.bg, hz)
			if c.bg == sWin && o.windowLit(lv, sx, sy) {
				bg = warm
			}
			if dim {
				bg = theme.Shade(bg, 0.55)
			}
			if sel {
				bg = theme.Mix(bg, white, 0.3)
			}
			o.fb.set(X, Y, c.ch, fg, bg, lv.depth)
		}
	}
}

// lotLights blinks a lot's beacons and lights its vents: a producer's
// vents glow while it is staffed and go dark when it idles.
func (o *orb) lotLights(lv *lotView, spr *sprite) {
	x0 := lv.ml.X - spr.w/2 - o.cam
	y0 := o.groundY - spr.h
	hz := lv.ml.Row
	for _, b := range spr.beacons {
		on := (o.anim/3+int(hash(int(lv.ml.Seed), b.x)%7))%8 < 3
		c := o.c(mapmodel.InkAccent3, iEmit, hz)
		if !on {
			c = theme.Mix(c, o.c(mapmodel.InkVoid, iBack, 0), 0.55)
		}
		o.fb.fg(x0+b.x, o.Y(y0+b.y), '•', c, lv.depth)
	}
	for i, e := range spr.smoke {
		busy := lv.staff
		if !lv.producer || busy < 0.05 {
			continue // an idle works shows no glow: that is the point
		}
		ph := 0.5 + 0.5*mapmodel.Sin(float64(o.anim+i*7)/24+hashf(int(lv.ml.Seed), i))
		c := o.c(mapmodel.InkGlow, iEmit, hz)
		if o.sc == mapmodel.SkyOrbit {
			c = o.c(mapmodel.InkLight, iEmit, hz)
		}
		c = theme.Mix(o.c(mapmodel.InkVoid, iBack, 0), c, 0.35+0.65*clamp01(busy)*(0.55+0.45*ph))
		ch := '▪'
		if busy > 0.6 && ph > 0.7 {
			ch = '■'
		}
		o.fb.fg(x0+e.x, o.Y(y0+e.y), ch, c, lv.depth)
	}
	for _, fl := range spr.flame {
		k := int(hash(int(lv.ml.Seed)+fl.x, o.anim/2) % 3)
		c := o.c([3]mapmodel.SkyInk{mapmodel.InkLight, mapmodel.InkGlow, mapmodel.InkAccent}[k], iEmit, hz)
		o.fb.fg(x0+fl.x, o.Y(y0+fl.y), [3]rune{'•', '∙', '•'}[k], c, lv.depth)
	}
}

// ------------------------------------------------------------ frontier

// frontier is construction: one frame going up per queued build (four at
// most), past the newest district, in the scene's own way: a scaffold that
// fills as the build progresses, with welding sparks.
func (o *orb) frontier() {
	fx := o.m.Skyline.FrontierX - o.cam + 4
	gy := o.groundY
	scaf := o.c(mapmodel.InkFrameDim, iLit, 0)
	fill := o.c(mapmodel.InkFrame, iLit, 0)
	spark := o.c(mapmodel.InkLight, iEmit, 0)
	if o.sc == mapmodel.SkyQuantum {
		scaf, fill, spark = o.iri(fx+o.anim/4, iLit, 0), o.iri(fx+o.anim/4+3, iLit, 0), o.iri(o.anim/2, iEmit, 0)
	}
	for i := range o.m.Queue {
		if i >= 4 {
			break
		}
		x := fx + i*7
		if x < -8 || x > o.W+2 {
			continue
		}
		p := o.m.Queue[i].Progress
		h := min(3+int(hash(i, 5)%3), max(2, gy-2))
		put := func(dx, y int, r rune, c tcell.Color) { o.fb.fg(x+dx, o.Y(y), r, c, dFrontier) }
		for y := gy - h; y < gy; y++ {
			put(0, y, '┊', scaf)
			put(4, y, '┊', scaf)
		}
		for dx := 0; dx <= 4; dx++ {
			put(dx, gy-h-1, '┄', scaf)
		}
		built := int(p*float64(h) + 0.5)
		for y := gy - built; y < gy; y++ {
			for dx := 1; dx <= 3; dx++ {
				put(dx, y, '▓', fill)
			}
		}
		if (o.anim/2+i*3)%5 < 2 {
			put(1+(o.anim/3+i)%3, gy-built-1, '✦', spark)
		}
	}
}

// ------------------------------------------------------------ catastrophe

// catastrophe is the pending catastrophe's overlay: the solar event's
// glitch bands, the reality fracture's cracks in space. No roof fires in a
// vacuum.
func (o *orb) catastrophe() {
	if o.m.Catastrophe.Pending == "" {
		return
	}
	switch o.m.Catalog.EpochIdx[o.m.Catastrophe.Pending] {
	case 4, 5:
		o.glitch()
	case 6:
		o.skyCracks()
		o.glitch()
	default:
		o.skyCracks()
	}
}

// ------------------------------------------------------------ chrome

// chrome is the header, the minimap strip and the status line, in the
// skyline's chrome: the header shows the clock without the weather (there
// is none up here) and the legend names what the sky scenes draw. The
// mandala has no panorama for a strip to show.
func (o *orb) chrome() {
	bg := chromeBg()
	for x := 0; x < o.W; x++ {
		o.fb.set(x, 0, ' ', bg, bg, dTop)
		o.fb.set(x, o.H-1, ' ', bg, bg, dTop)
	}
	o.header(bg)
	if o.sc != mapmodel.SkyMandala {
		o.minimap()
	}
	o.status(bg)
}

func (o *orb) header(bg tcell.Color) {
	m := o.m
	acc := role(theme.RoleAccent, bg)
	x := 0
	for _, r := range "░▒▓" {
		o.fb.set(x, 0, r, acc, bg, dTop)
		x++
	}
	x = o.fb.text(x, 0, clip(" "+strings.ToUpper(m.AgeName)+" ", o.W-x), theme.Legible(bg, acc, 4.5), acc, dTop)
	for _, r := range "▓▒░" {
		o.fb.set(x, 0, r, acc, bg, dTop)
		x++
	}
	parts := []seg{
		{" " + m.EpochName, theme.RoleLabel},
		{"  day ", theme.RoleDim}, {strconv.Itoa(m.Clock.Day), theme.RoleHighlight},
		{"  " + m.Clock.String(), theme.RoleText},
		{"  pop ", theme.RoleDim}, {strconv.Itoa(m.Workers.Pop) + "/" + strconv.Itoa(m.Workers.MaxPop), theme.RoleHighlight},
		{"  bld ", theme.RoleDim}, {strconv.Itoa(m.TotalBuildings()), theme.RoleHighlight},
	}
	if n := m.Recap.NewCount; o.v.changes && n > 0 {
		parts = append(parts, seg{"  ▼ " + strconv.Itoa(n) + " new since your last visit", theme.RolePositive})
	}
	if n := idleProducers(m); n > 0 {
		parts = append(parts, seg{"  ▪ " + strconv.Itoa(n) + " idle", theme.RoleWarning})
	}
	if o.v.flows && m.Workers.Idle > 0 {
		parts = append(parts, seg{"  " + strconv.Itoa(m.Workers.Idle) + " idle workers", theme.RoleWarning})
	}
	if h := m.Harbinger; h != nil {
		parts = append(parts, seg{"  " + string(mapmodel.R(mapmodel.SymHarbinger, o.tier)) + " " + h.Name + " is watching", theme.RoleWarning})
	}
	if c := m.Catastrophe; c.Pending != "" {
		parts = append(parts, seg{"  " + string(mapmodel.R(mapmodel.SymFire, o.tier)) + " " + c.PendingName, theme.RoleNegative})
	}
	o.segs(x, 0, bg, parts)
}

// skyLegend is what the legend says each sky scene draws.
var skyLegend = [mapmodel.NumSkyScenes]string{
	mapmodel.SkyOrbit:   "city lights below ",
	mapmodel.SkyDeep:    "colony beacons ",
	mapmodel.SkyGalaxy:  "star systems ",
	mapmodel.SkyQuantum: "echoes ",
}

var skyLegendWhat = [mapmodel.NumSkyScenes]string{
	mapmodel.SkyOrbit:   "your districts  ",
	mapmodel.SkyDeep:    "your lineages  ",
	mapmodel.SkyGalaxy:  "civs met  ",
	mapmodel.SkyQuantum: "the town you were  ",
}

func (o *orb) status(bg tcell.Color) {
	y := o.H - 1
	m := o.m
	if o.v.inspect {
		if in, ok := o.v.inspection(m, o.anim); ok {
			lines := in.lines
			if o.v.cur.kind == tLot {
				lines = o.skyLines(in.lines)
			}
			// the command outlasts the details: drop them from the end
			// until the line fits, keeping at least the first
			fits := func(n int) bool {
				w := textLen(" ▲ ") + textLen(in.title)
				for _, l := range lines[:n] {
					w += 2 + textLen(l)
				}
				if in.cmd != "" {
					w += textLen("  type: ") + textLen(in.cmd)
				}
				return w <= o.W
			}
			for len(lines) > 1 && !fits(len(lines)) {
				lines = lines[:len(lines)-1]
			}
			parts := []seg{{" ▲ ", theme.RoleAccent}, {in.title, theme.RoleText}}
			for _, l := range lines {
				parts = append(parts, seg{"  " + l, theme.RoleDim})
			}
			if in.cmd != "" {
				parts = append(parts, seg{"  type: ", theme.RoleDim}, seg{in.cmd, theme.RoleLabel})
			}
			o.segs(0, y, bg, parts)
			return
		}
	}
	if o.v.flows {
		w := m.Flows.Worst
		if w == "" {
			w = "nothing is stuck"
		}
		parts := []seg{{" flows: ", theme.RoleAccent}, {w, theme.RoleWarning}}
		if len(m.Flows.Full) > 0 {
			names := make([]string, 0, len(m.Flows.Full))
			for _, st := range m.Flows.Full {
				names = append(names, strings.ToLower(st.Name))
			}
			parts = append(parts, seg{"  full: ", theme.RoleDim}, seg{strings.Join(names, ", "), theme.RoleWarning})
		}
		parts = append(parts, seg{"  type: ", theme.RoleDim}, seg{m.IdleCommand(), theme.RoleLabel})
		o.segs(0, y, bg, parts)
		return
	}
	if o.sc == mapmodel.SkyMandala { // no city to explain, no panorama to scroll
		parts := o.mandalaHints()
		if o.v.legend {
			parts = o.mandalaLegend()
		}
		o.segs(0, y, bg, parts)
		return
	}
	if o.v.legend {
		parts := []seg{{" lit windows ", theme.RoleText}, {"staffed  ", theme.RoleDim},
			{"glowing vents ", theme.RoleText}, {"producing  ", theme.RoleDim}, {"▼ ", theme.RolePositive}, {"new  ", theme.RoleDim},
			{"IDLE LOW FULL ", theme.RoleWarning}, {"flows  ", theme.RoleDim},
			{skyLegend[o.sc], theme.RoleText}, {skyLegendWhat[o.sc], theme.RoleDim}}
		if o.sc == mapmodel.SkyGalaxy {
			parts = append(parts, seg{string([]rune{'◄', mapmodel.R(mapmodel.SymUFO, o.tier), '►'}) + " ", theme.RoleText},
				seg{"alien traffic  ", theme.RoleDim})
		}
		parts = append(parts, seg{"pennants ", theme.RoleText}, seg{"your standing with each civ", theme.RoleDim})
		o.segs(0, y, bg, parts)
		return
	}
	where := "the void"
	if di := m.Skyline.DistrictAt(o.cam + o.W/2); di >= 0 {
		where = strings.ToLower(m.Catalog.AgeNames[m.Skyline.Districts[di].Age]) + " " + skyWhere[o.sc]
	}
	o.segs(0, y, bg, panoramaHints(o.W, where))
}

// skyWhere names a district in each scene's terms.
var skyWhere = [mapmodel.NumSkyScenes]string{
	mapmodel.SkyOrbit: "section", mapmodel.SkyDeep: "fleet", mapmodel.SkyGalaxy: "ring",
	mapmodel.SkyQuantum: "probability",
}

// skyLines words a lot's inspect lines in the scene's terms: the lineage
// becomes what it is up here (mapmodel.SkyPartOf), so both styles agree.
func (o *orb) skyLines(lines []string) []string {
	key := o.v.cur.key
	def := o.m.Catalog.Defs[key]
	if def == nil || def.Wonder || len(lines) == 0 {
		return lines
	}
	part := mapmodel.SkyPartOf(o.sc, def.Lineage)
	out := append([]string(nil), lines...)
	if lin := mapmodel.LineageNames[def.Lineage]; lin != "" && part.Name != lin {
		if strings.HasPrefix(out[0], lin+" · ") {
			out[0] = part.Name + " (" + lin + ") · " + strings.TrimPrefix(out[0], lin+" · ")
		}
	}
	return out
}

// ------------------------------------------------------------ layout

// skyLayKey marks a sky layout in the layout slot: negative, so layoutFor
// (whose ground rows are positive) never takes it for its own.
func skyLayKey(sc mapmodel.SkyScene, groundY int, compact bool) int {
	k := int(sc)*1000 + groundY + 1
	if compact {
		k += 500
	}
	return -k
}

// skyLayoutFor dresses the model's lots in the scene's sky forms, cached
// per (model, scene, baseline) like layoutFor: the full view's in v.slay,
// the compact view's in v.sclay. Placement is the model's: every form is
// centred on its lot's column.
func (v *view) skyLayoutFor(m *mapmodel.Model, sc mapmodel.SkyScene, groundY int, compact bool) *layout {
	key := skyLayKey(sc, groundY, compact)
	slotp := &v.slay
	if compact {
		slotp = &v.sclay
	}
	if l := *slotp; l != nil && l.m == m && l.groundY == key {
		return l
	}
	if v.sprites == nil || len(v.sprites) > 6000 {
		v.sprites = map[spriteKey]*sprite{}
	}
	lay := &layout{m: m, groundY: key, scale: skyScale(groundY), width: max(m.Skyline.Width, 1)}
	maxH := max(2, groundY-2)
	present := m.AgeIdx
	keystone := keystoneWonder[sc]
	if n := len(m.Skyline.Districts); n > 0 {
		present = m.Skyline.Districts[n-1].Age
	}
	for i := range m.Skyline.Lots {
		ml := &m.Skyline.Lots[i]
		def := m.Catalog.Defs[ml.Key]
		if def == nil {
			continue
		}
		b := m.Building(ml.Key)
		lv := lotView{ml: ml, b: b, vr: int(hash(int(ml.Seed), 7) % 3), staff: staffOf(m, b), depth: rowDepth(ml.Row)}
		// older sections are weathered: a step dimmer for every six ages
		// back, legacy tiers a step more
		lv.fam = clampInt((present-ml.Age)/6, 0, 2)
		if ml.Legacy {
			lv.fam = min(3, lv.fam+1)
		}
		count := 1
		if b != nil {
			count = b.Count
		}
		fm := skyFormFor(sc, def, ml, ml.Key == keystone && ml.Age == present)
		lv.producer = b != nil && b.Rate > 0
		h := float64(fm.height * lay.scale)
		if !ml.Wonder {
			h *= 0.82 + 0.36*hashf(int(ml.Seed), 11)
			h *= 1 + 0.05*mapmodel.Log2(1+float64(count))
			h *= 0.78 + 0.22*clamp01(1-float64(present-ml.Age)/18) // the old core is smaller
			if ml.Row > 0 {
				h *= 1.1
			}
		} else {
			lv.depth = dWonder
		}
		lift := 0 // the Interstellar's fleet floats in formation, the far rows higher
		if sc == mapmodel.SkyDeep && !ml.Wonder {
			lift = ml.Row*2 + int(hash(int(ml.Seed), 13)%2)
		}
		sk := spriteKey{fm.key, ml.Seed, clampInt(int(h+0.5), 2, maxH)}
		if lift > 0 {
			sk.key += "^" + strconv.Itoa(lift)
		}
		if ml.Wonder {
			sk.seed = uint64(hash(len(ml.Key), int(ml.Key[0])))
		}
		spr := v.sprites[sk]
		if spr == nil {
			spr = fm.fn(newRnd(ml.Seed), sk.h)
			if !ml.Wonder && hash(int(ml.Seed), 5)%2 == 0 {
				spr = spr.mirror()
			}
			if spr.h > maxH+2 {
				spr = spr.cropTop(maxH + 2)
			}
			if lift > 0 {
				spr = spr.padFoot(lift)
			}
			v.sprites[sk] = spr
		}
		lv.spr = spr
		lv.x0 = ml.X - spr.w/2
		lay.maxW = max(lay.maxW, spr.w)
		lay.lots = append(lay.lots, lv)
	}
	sort.SliceStable(lay.lots, func(i, j int) bool {
		a, b := &lay.lots[i], &lay.lots[j]
		if a.depth != b.depth {
			return a.depth > b.depth
		}
		if a.spr.h != b.spr.h {
			return a.spr.h > b.spr.h
		}
		return a.ml.Seed < b.ml.Seed
	})
	lay.index = make(map[lotID]int, len(lay.lots))
	for i := range lay.lots {
		ml := lay.lots[i].ml
		lay.index[lotID{ml.Key, ml.Copy}] = i
		if ml.New {
			lay.news = append(lay.news, ml.X)
		}
	}
	lay.byX = make([]int32, len(lay.lots))
	for i := range lay.byX {
		lay.byX[i] = int32(i)
	}
	sort.SliceStable(lay.byX, func(i, j int) bool { return lay.lots[lay.byX[i]].x0 < lay.lots[lay.byX[j]].x0 })
	lay.prof = make([]int, lay.width)
	lay.tallest = make([]int32, lay.width)
	for i := range lay.tallest {
		lay.tallest[i] = -1
	}
	for i := range lay.lots {
		l := &lay.lots[i]
		for x := 0; x < l.spr.w; x++ {
			wx := l.x0 + x
			if wx < 0 || wx >= lay.width {
				continue
			}
			if t := l.spr.topAt(x); t >= 0 {
				if h := l.spr.h - t; h > lay.prof[wx] {
					lay.prof[wx] = h
					lay.tallest[wx] = int32(i)
				}
			}
		}
	}
	*slotp = lay
	return lay
}

// groundBehind keeps the model's ground layout at the sky's baseline in
// v.glay: the Quantum Age's time echo and the forms its lots flicker to.
func (v *view) groundBehind(m *mapmodel.Model, groundY int) *layout {
	if v.glay != nil && v.glay.m == m && v.glay.groundY == groundY {
		return v.glay
	}
	keep := v.lay // layoutFor caches in the ground slot; keep it as it was
	v.lay = nil
	v.glay = v.layoutFor(m, groundY)
	v.lay = keep
	return v.glay
}

// padFoot adds n empty rows under a sprite: it floats that high over the
// baseline, while the cursor, the markers and the profile still find it.
func (s *sprite) padFoot(n int) *sprite {
	o := newSprite(s.w, s.h+n)
	copy(o.c, s.c)
	o.smoke, o.beacons, o.blades, o.flame = s.smoke, s.beacons, s.blades, s.flame
	return o
}

// cropTop drops rows off the top of a sprite taller than h.
func (s *sprite) cropTop(h int) *sprite {
	if s.h <= h {
		return s
	}
	cut := s.h - h
	o := newSprite(s.w, h)
	copy(o.c, s.c[cut*s.w:])
	sh := func(ps []pt) []pt {
		out := make([]pt, 0, len(ps))
		for _, p := range ps {
			if p.y-cut >= 0 {
				out = append(out, pt{p.x, p.y - cut})
			}
		}
		return out
	}
	o.smoke, o.beacons, o.blades, o.flame = sh(s.smoke), sh(s.beacons), sh(s.blades), sh(s.flame)
	return o
}
