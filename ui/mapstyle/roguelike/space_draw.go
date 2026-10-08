package roguelike

import (
	"math"
	"sort"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_draw.go draws a sky scene: the palette resolved against the theme,
// the camera over the plane, the cells with their effects, the traffic,
// the labels and the cursor. Everything heavy was laid out when the scene
// was built; a frame indexes it and runs a few hashes per cell.

// Inks past the scene palette: the relation colours, the state colours,
// the alien species and the Quantum Age's iridescent ramp.
const (
	inkRel    = mapmodel.NumSkyInks // + mapmodel.Relation
	inkFresh  = inkRel + 6
	inkIdle   = inkFresh + 1
	inkDanger = inkIdle + 1
	inkAlien  = inkDanger + 1 // + species
	inkIri    = inkAlien + 3  // + stop
	inkEra    = inkIri + 8    // + epoch: a mandala ring's era colour
	numInks   = inkEra + 7
)

// skyPal is a sky scene's palette on the active theme: each ink at four
// brightness levels, legible on the scene's void, and as a background.
type skyPal struct {
	key   string
	sky   mapmodel.SkyScene
	void  tcell.Color
	lv    [numInks][4]tcell.Color
	bgc   [numInks]tcell.Color
	bgs   [numInks]tcell.Color // soft: a faint tint of the void
	fresh tcell.Color
	flow  tcell.Color
	memo  map[[2]tcell.Color]tcell.Color
}

// palette returns the scene's palette, rebuilt when the theme or the scene
// changes. It also readies the ground view's palette, whose classes give
// the chrome, the relations and the state colours.
func (v *skyView) palette(sky mapmodel.SkyScene, epoch, age int) *skyPal {
	gp := v.g.palette(epoch, age)
	key := theme.Active().Key
	if v.sp != nil && v.sp.key == key && v.sp.sky == sky {
		return v.sp
	}
	v.sp = newSkyPal(sky, key, gp)
	return v.sp
}

func newSkyPal(sky mapmodel.SkyScene, key string, gp *pal) *skyPal {
	sp := &skyPal{key: key, sky: sky, memo: map[[2]tcell.Color]tcell.Color{}}
	bg := theme.Color(theme.RoleBackground)
	light := theme.IsLight()
	duo := theme.Active().Duotone
	lo, hi := bg, theme.Color(theme.RoleText)
	if luma(lo) > luma(hi) {
		lo, hi = hi, lo
	}
	fold := func(c tcell.Color) tcell.Color {
		if duo {
			return theme.Mix(lo, hi, luma(c)/255)
		}
		return c
	}
	pl := mapmodel.SkyPaletteOf(sky)
	voidHue := theme.SpaceColor(pl[mapmodel.InkVoid])
	switch {
	case duo:
		sp.void = bg
	case light:
		sp.void = theme.Mix(bg, voidHue, 0.10)
	default:
		k := 0.82
		if sky == mapmodel.SkyMandala {
			k = 0.94 // the indigo is half the look
		}
		sp.void = theme.Mix(bg, voidHue, k)
	}
	raw := func(i mapmodel.SkyInk) tcell.Color {
		switch {
		case i < mapmodel.NumSkyInks:
			return theme.SpaceColor(pl[i])
		case i < inkFresh:
			return gp.Fg[mapmodel.RelationClass(mapmodel.Relation(i-inkRel))]
		case i == inkFresh:
			return gp.Fg[mapmodel.CFresh]
		case i == inkIdle:
			return gp.Fg[mapmodel.CIdle]
		case i == inkDanger:
			return gp.Fg[mapmodel.CDanger]
		case i < inkIri:
			return theme.SpaceColor(mapmodel.Aliens[int(i-inkAlien)%len(mapmodel.Aliens)].Hue)
		case i >= inkEra:
			return mapstyle.EraLight(int(i - inkEra))
		}
		return theme.SpaceColor(mapmodel.SkyIridescent[int(i-inkIri)%len(mapmodel.SkyIridescent)])
	}
	peak := theme.Color(theme.RoleBright)
	for i := mapmodel.SkyInk(0); i < numInks; i++ {
		c := fold(raw(i))
		sp.lv[i][2] = theme.Legible(c, sp.void, 3)
		sp.lv[i][1] = theme.Legible(theme.Mix(sp.void, c, 0.62), sp.void, 1.9)
		sp.lv[i][0] = theme.Legible(theme.Mix(sp.void, c, 0.38), sp.void, 1.3)
		sp.lv[i][3] = theme.Legible(theme.Mix(c, peak, 0.35), sp.void, 4)
		k := 0.85
		if light || duo {
			k = 0.28
		}
		sp.bgc[i] = theme.Mix(sp.void, c, k)
		sp.bgs[i] = theme.Mix(sp.void, c, k*0.32)
	}
	sp.fresh = theme.Mix(sp.void, gp.Fg[mapmodel.CFresh], 0.30)
	sp.flow = theme.Mix(sp.void, gp.Fg[mapmodel.CIdle], 0.26)
	return sp
}

func luma(c tcell.Color) float64 {
	r, g, b := c.RGB()
	if r < 0 {
		return 0
	}
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

// on is fg made legible on bg (memoised).
func (sp *skyPal) on(fg, bg tcell.Color) tcell.Color {
	if bg == sp.void {
		return fg
	}
	k := [2]tcell.Color{fg, bg}
	c, ok := sp.memo[k]
	if !ok {
		c = theme.Legible(fg, bg, 3)
		sp.memo[k] = c
	}
	return c
}

// skyGeom is the map area's geometry for one frame.
type skyGeom struct {
	x, y, w, h int // the map area on the canvas
	cellW      int // cells per plane cell: 1, or 2 at the close zoom
	vx, vy     int // the plane cell under the area's top-left cell
}

// cellOf maps a plane cell to its canvas cell, if it is in the map area.
func (g *skyGeom) cellOf(p mapmodel.Pt) (int, int, bool) {
	cx, cy := (p.X-g.vx)*g.cellW, p.Y-g.vy
	if cx < 0 || cy < 0 || cx >= g.w || cy >= g.h {
		return 0, 0, false
	}
	return g.x + cx, g.y + cy, true
}

// fit places the camera: the whole plane centred when it fits, else
// following the cursor, scrolling only when it nears an edge.
func (v *skyView) fit(x, y, w, h, zoom int) skyGeom {
	g := skyGeom{x: x, y: y, w: w, h: h, cellW: 1}
	if zoom == zDistrict {
		g.cellW = 2
	}
	tw, th := max(1, w/g.cellW), h
	axis := func(view, size, cur int, c *int, key int, ck *int) int {
		if view >= size {
			return -(view - size) / 2
		}
		if key != *ck || v.snap || cur < *c || cur >= *c+view {
			*ck = key
			*c = cur - view/2
		}
		mg := min(8, view/4)
		*c = clamp(*c, cur-view+mg+1, cur-mg)
		*c = clamp(*c, 0, size-view)
		return *c
	}
	c := &v.cam[zoom]
	k := &v.camKey[zoom]
	g.vx = axis(tw, skyW, v.cur.X, &c.X, tw, &k[0])
	g.vy = axis(th, skyH, v.cur.Y, &c.Y, th, &k[1])
	return g
}

// begin readies the per-frame state and returns the scene.
func (v *skyView) begin(f mapstyle.Frame) *skyScene {
	v.anim, v.visit, v.clock, v.tier = f.Anim, f.VisitFrame(), f.Clock, f.Tier
	if len(v.seen) != int(numSkyLg) {
		v.seen = make([]skyLgEntry, numSkyLg)
	} else {
		clear(v.seen)
	}
	s := v.sceneFor(f.Model)
	ep, age := 5, -1
	sky := mapmodel.SkyOrbit
	if s != nil {
		ep, sky, age = s.m.Epoch, s.sky, s.m.AgeIdx
	}
	v.palette(sky, ep, age)
	return s
}

// look is a cell resolved for one frame.
type look struct {
	r    rune
	ink  mapmodel.SkyInk
	lv   uint8
	bg   mapmodel.SkyInk
	bgc  tcell.Color // overrides bg
	bold bool
	soft bool
}

// resolve turns a scene cell into what this frame shows.
func (v *skyView) resolve(s *skyScene, c skyCell, x, y int) look {
	l := look{r: c.r, ink: c.ink, lv: c.lv, bg: c.bg, bold: c.bold, soft: c.soft}
	if c.sym != mapmodel.SymNone {
		l.r = mapmodel.R(c.sym, v.tier)
	}
	f := v.anim
	switch c.fx {
	case fxTwinkle:
		switch (f/3 + int(c.ph)) % 41 {
		case 0:
			l.lv, l.bold = 3, true
		case 1:
			l.r, l.lv, l.bold = '*', 3, true
		case 2:
			l.lv = 3
		case 20, 21:
			l.lv = 0
		}
	case fxBlink:
		if (f/4+int(c.ph))%6 >= 3 {
			l.lv = uint8(max(0, int(l.lv)-2))
		}
	case fxCloud:
		if s.b.sky == mapmodel.SkyOrbit {
			l = v.cloud(s, c, l, x, y)
		}
	case fxSwirl:
		ph := (int(c.ph) + 24 - f/3%24) % 12 // arms turning round the gate's heart
		l.r = [12]rune{'▒', '░', '░', '·', ' ', ' ', '·', '∙', '░', '·', ' ', '·'}[ph]
		if ph >= 6 {
			l.ink = mapmodel.InkCloud
		}
	case fxSpark:
		if (f+int(c.ph))%5 < 2 {
			l.r, l.ink, l.lv, l.bold = '✦', mapmodel.InkStarBright, 3, true
		}
	case fxFlicker:
		if superposed(x, y, f) {
			l.r = c.alt
		}
	case fxFringe:
		l = v.fringe(c, l, f)
	case fxBreathe:
		l = v.breathe(s, c, l, x, y)
	case fxIri:
		l.ink = inkIri + mapmodel.SkyInk((x+2*y+f/4)%8)
	case fxSparkle:
		ph := (f/3 + int(c.ph)) % 23
		if ph < 3 {
			l.r, l.lv, l.bold = [3]rune{'˙', '✧', '·'}[ph], 3, true
		}
	case fxWisp:
		if (x/3+f/40+int(c.ph))%7 == 0 {
			l.lv = uint8(max(0, int(l.lv)-1))
		}
	}
	if s.sky == mapmodel.SkyQuantum {
		switch c.k {
		case skFrame, skHub, skSlot:
			l.ink = inkIri + mapmodel.SkyInk((x+2*y+f/4)%8) // iridescence runs along every structure
		case skUnit:
			// each lineage its own hue, all of them turning slowly
			lin := lineageIdx(s.m.Town.Tiles[c.ref].Lineage)
			l.ink = inkIri + mapmodel.SkyInk((lin+f/16)%8)
		}
	}
	return l
}

// superposed is the Quantum Age's stable pattern: bands sweeping slowly
// across the plane, each cell showing its other state while a band covers
// it.
func superposed(x, y, f int) bool { return ((x+2*y)/3+f/5)%9 < 2 }

// style makes a tcell style of a look.
func (v *skyView) style(l look) tcell.Style {
	sp := v.sp
	fg := sp.lv[l.ink][min(3, int(l.lv))]
	bg := sp.void
	switch {
	case l.bgc != 0:
		bg = l.bgc
	case l.bg != mapmodel.InkVoid && l.soft:
		bg = sp.bgs[l.bg]
	case l.bg != mapmodel.InkVoid:
		bg = sp.bgc[l.bg]
	}
	if bg != sp.void {
		fg = sp.on(fg, bg)
	}
	st := tcell.StyleDefault.Foreground(fg).Background(bg)
	if l.bold {
		st = st.Bold(true)
	}
	return st
}

// drawSkyMap draws the map area: the cells, the traffic, the visitor, the
// labels, then the cursor.
func (v *skyView) drawSkyMap(cv *mapstyle.Canvas, x, y, w, h int) {
	s := v.sc
	if w <= 0 || h <= 0 || s == nil {
		return
	}
	v.sg = v.fit(x, y, w, h, v.g.zoom)
	v.snap = false
	v.drawCells(cv, v.sg)
	v.drawOverlay(cv)
	v.drawSkyTraffic(cv)
	if !v.compact {
		v.drawSkyVisitor(cv)
		v.drawSkyLabels(cv)
	}
	if cx, cy, ok := v.sg.cellOf(v.cur); ok && v.g.inspect && !v.compact {
		for i := 0; i < v.sg.cellW && cx+i < v.sg.x+v.sg.w; i++ {
			r, _ := cv.Get(cx+i, cy)
			cv.Put(cx+i, cy, r, v.g.pal.Cursor.Bold(true))
		}
	}
}

// drawCells draws the plane cells under the map area, registering what
// shows in the legend.
func (v *skyView) drawCells(cv *mapstyle.Canvas, g skyGeom) {
	s := v.sc
	n := g.w * g.h
	if cap(v.occ) < n {
		v.occ = make([]bool, n)
	}
	v.occ = v.occ[:n]
	clear(v.occ)
	for i := 0; i < n; i++ {
		cx, cy := i%g.w, i/g.w
		px, py := g.vx+cx/g.cellW, g.vy+cy
		c := s.at(px, py)
		var l look
		if g.cellW == 2 && cx%2 == 1 {
			l = v.rightHalf(s, c, px, py)
		} else {
			l = v.resolve(s, c, px, py)
			v.legendCell(s, c, l)
		}
		l = v.stateOf(s, c, l, px, py)
		v.occ[i] = c.salience() >= 40
		cv.Put(g.x+cx, g.y+cy, l.r, v.style(l))
	}
}

// rightHalf is the second cell of a plane cell at the close zoom: lines
// run on, planets and fields fill, everything else leaves a gap.
func (v *skyView) rightHalf(s *skyScene, c skyCell, x, y int) look {
	l := v.resolve(s, c, x, y)
	east := s.at(x+1, y)
	switch c.k {
	case skFrame, skRing, skTether, skSlot:
		switch l.r {
		case '═', '╔', '╚', '╠', '╦', '╩', '╬':
			if joins(east) {
				l.r = '═'
				return l
			}
		case '─', '┌', '└', '├', '┬', '┴', '┼', '╭', '╰':
			if joins(east) {
				l.r = '─'
				return l
			}
		case '▤', '▓', '░', '·', '∙':
			if east.k == c.k {
				return l
			}
		}
	case skUnit:
		if b := s.m.Building(s.m.Town.Tiles[c.ref].Key); b != nil && joins(east) {
			l.r = '═'
			l.ink, l.lv = mapmodel.InkFrameDim, 1
			return l
		}
	case skPlanet, skField, skMoon, skColony, skNebula, skSpiral, skCloud:
		return l
	}
	return look{r: ' ', ink: l.ink, bg: l.bg, bgc: l.bgc, soft: l.soft}
}

// joins reports whether a cell is part of a structure a line runs into.
func joins(c skyCell) bool {
	return c.k == skFrame || c.k == skRing || c.k == skUnit || c.k == skSlot || c.k == skHub
}

// stateOf marks a unit by its building's state: legacy and understaffed
// units dim, new ones on the fresh background, flagged ones on the flows
// background (with the warning at the type's first unit).
func (v *skyView) stateOf(s *skyScene, c skyCell, l look, x, y int) look {
	if c.k != skUnit {
		return l
	}
	tt := &s.m.Town.Tiles[c.ref]
	b := s.m.Building(tt.Key)
	switch {
	case tt.Ruin || b == nil:
		l.r, l.ink, l.lv = mapmodel.R(mapmodel.SymRuin, v.tier), mapmodel.InkRockDark, 1
		v.reg(slRuin, l)
		return l
	case tt.Legacy:
		l.lv = 1
		v.reg(slLegacy, l)
	case b.Understaffed():
		l.lv = 1
		v.reg(slUnder, l)
	}
	switch {
	case v.g.flows && flagged(b):
		l.bgc = v.sp.flow
		if tt.Anchor && s.anchor[tt.Key] == pt(x, y) {
			l.r, l.ink, l.lv, l.bold = mapmodel.R(mapmodel.SymWarning, v.tier), inkIdle, 2, true
		}
		v.reg(slFlagged, look{r: mapmodel.R(mapmodel.SymWarning, v.tier), ink: inkIdle, lv: 2, bgc: v.sp.flow})
	case v.g.changes && tt.Fresh:
		l.bgc = v.sp.fresh
		v.reg(slFresh, look{r: ' ', bgc: v.sp.fresh})
	}
	return l
}

// cloud is the planet's surface at (x, y) this frame: clouds drift over
// the night side, one column every few seconds.
func (v *skyView) cloud(s *skyScene, c skyCell, l look, x, y int) look {
	if c.k != skPlanet {
		return l
	}
	off := v.anim / 40
	n := mapmodel.Noise(s.b.seed+77, float64(x+off)/7, float64(y)/1.6)
	switch {
	case n > 0.74:
		l.r, l.ink, l.lv = '▒', mapmodel.InkCloud, 1
	case n > 0.66:
		l.r, l.ink, l.lv = '░', mapmodel.InkCloud, 1
	}
	return l
}

// fringe is a probability cloud cell this frame: the fringes slide slowly
// across it.
func (v *skyView) fringe(c skyCell, l look, f int) look {
	ph := (int(c.ph) + f/3) % 32
	switch {
	case ph < 1:
		l.r, l.ink, l.lv = '▒', mapmodel.InkCloud2, 2
	case ph < 3:
		l.r, l.ink, l.lv = '░', mapmodel.InkCloud, 2
	case ph < 5:
		l.r, l.ink, l.lv = '·', mapmodel.InkCloud, 1
	default:
		l.r = ' '
	}
	return l
}

// drawSkyLabels names things on the plane: civs at both zooms; at the
// close zoom also the building types nearest the cursor and the wonders.
func (v *skyView) drawSkyLabels(cv *mapstyle.Canvas) {
	s, g := v.sc, &v.sg
	for _, lb := range s.labels {
		if lb.rel == 0 || lb.far && g.cellW == 1 {
			continue
		}
		v.label(cv, lb.p, lb.text, v.style(look{ink: inkRel + mapmodel.SkyInk(lb.rel-1), lv: 2}))
	}
	v.names = 0
	if g.cellW != 2 {
		return
	}
	type named struct {
		p    mapmodel.Pt
		name string
	}
	var ns []named
	for k, p := range s.anchor {
		if b := s.m.Building(k); b != nil {
			if _, _, ok := g.cellOf(p); ok {
				ns = append(ns, named{p, b.Name})
			}
		}
	}
	d := func(p mapmodel.Pt) int { return abs(p.X-v.cur.X) + abs(p.Y-v.cur.Y) }
	sort.SliceStable(ns, func(i, j int) bool {
		if di, dj := d(ns[i].p), d(ns[j].p); di != dj {
			return di < dj
		}
		return ns[i].name < ns[j].name
	})
	budget := districtLabelBudget(g.w, g.h)
	for _, n := range ns {
		if v.names == budget {
			break
		}
		if v.label(cv, n.p, n.name, v.style(look{ink: mapmodel.InkStarBright, lv: 2})) {
			v.names++
		}
	}
	for _, lb := range s.labels {
		if lb.rel == 0 && lb.far {
			v.label(cv, lb.p, lb.text, v.style(look{ink: lb.ink, lv: 2}))
		}
	}
}

// label writes text beside a plane cell in the first free spot round it.
func (v *skyView) label(cv *mapstyle.Canvas, p mapmodel.Pt, text string, st tcell.Style) bool {
	g := &v.sg
	cx, cy, ok := g.cellOf(p)
	if !ok {
		return false
	}
	n := mapstyle.TextLen(text)
	for _, o := range [6][2]int{{2 * g.cellW, 0}, {-n - 1, 0}, {-n / 2, -1}, {-n / 2, 1}, {2, 1}, {2, -1}} {
		if v.freeAt(cx+o[0], cy+o[1], n) {
			cv.Text(cx+o[0], cy+o[1], n, text, st)
			return true
		}
	}
	return false
}

// freeAt claims n cells from (x, y) for a label when they and a one-cell
// margin hold nothing salient.
func (v *skyView) freeAt(x, y, n int) bool {
	g := &v.sg
	if y < g.y || y >= g.y+g.h || x < g.x || x+n > g.x+g.w {
		return false
	}
	row := v.occ[(y-g.y)*g.w : (y-g.y+1)*g.w]
	for i := max(x-1, g.x); i <= min(x+n, g.x+g.w-1); i++ {
		if row[i-g.x] {
			return false
		}
	}
	for i := x; i < x+n; i++ {
		row[i-g.x] = true
	}
	return true
}

// drawOverlay draws what moves without being traffic: the orrery's planets
// round the home sun.
func (v *skyView) drawOverlay(cv *mapstyle.Canvas) {
	if v.sc.sky == mapmodel.SkyOrbit {
		for k := 0; k < orbDebris; k++ {
			if p, ok := v.debrisAt(k); ok && v.putMover(cv, v.sc, p, '·', mapmodel.InkFrameDim, 2, false) {
				v.reg(slDebris, look{r: '·', ink: mapmodel.InkFrameDim, lv: 2})
			}
		}
		return
	}
	if v.sc.sky != mapmodel.SkyDeep {
		return
	}
	for k := range deepOrbits {
		p := v.orreryPlanet(k)
		r, ink := deepPlanets[k].r, deepPlanets[k].ink
		if v.putMover(cv, v.sc, p, r, ink, 3, k == 2) && k == 2 {
			v.reg(slPlanet, look{r: r, ink: ink, lv: 3, bold: true})
		}
	}
}

// deepPlanets are the orrery's planets, innermost first; the third is the
// homeworld.
var deepPlanets = [4]struct {
	r      rune
	ink    mapmodel.SkyInk
	period int
}{{'∙', mapmodel.InkRock, 240}, {'•', mapmodel.InkAccent3, 400}, {'●', mapmodel.InkSurface, 640}, {'◦', mapmodel.InkMoon, 960}}

// orreryPlanet is where planet k of the orrery is this frame.
func (v *skyView) orreryPlanet(k int) mapmodel.Pt {
	o := deepOrbits[k]
	t := mapmodel.HashF(v.sc.b.seed, 560, int64(k)) + float64(v.anim)/float64(deepPlanets[k].period)
	return pt(int(math.Round(float64(deepSunX)+o[0]*mapmodel.Sin(t))), int(math.Round(float64(deepSunY)-o[1]*mapmodel.Cos(t))))
}

// orbDebris is how many specks of debris drift round the Space Age's
// planet.
const orbDebris = 7

// debrisAt is where speck k of the debris is this frame: on a low orbit
// over the limb, each at its own slow pace (a cell every second or two).
func (v *skyView) debrisAt(k int) (mapmodel.Pt, bool) {
	p := &v.sc.b.planet
	seed := v.sc.b.seed
	pace := 8 + int(mapmodel.Hash(seed, 680, int64(k))%9)
	span := skyW + 40
	x := (int(mapmodel.Hash(seed, 681, int64(k))%uint64(span)) + v.anim/pace) % span
	if k%2 == 1 {
		x = span - 1 - x
	}
	x -= 20
	if x < 0 || x >= skyW {
		return mapmodel.Pt{}, false
	}
	return pt(x, int(math.Round(p.limb(x)-7-float64(k%3)))), true
}
