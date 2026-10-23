package skyline

import (
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// orbit_mandala.go is the Transcendent Age. The civilization has left
// corporeal form, and the skyline leaves the corporeal world with it: no
// city, no ground for one to stand on, no far civs, no traffic. What is
// left is the mandala both styles draw (mapstyle.LayMandala), alone on the
// indigo and as large as the pane: a core burning gold in a star of rays, a
// crown of eight solid petals for the age's own buildings, and one ring for
// every era the player passed through, the Stone Era innermost, each a
// band of its era's glyph in its era's colour with dark space between
// them. It is strictly symmetric about the core, and a crest of light
// rolls out from the core ring by ring.
//
// It is still the player's history, and the cursor still reads it: Tab
// steps from the core through the crown and out ring by ring, and every
// building type still standing has a cell of its era's ring (its age's
// petal, in the crown) that says what it is and the command for it.

// mandalaLayout is the mandala laid out for a pane: its centre cell
// (column, scene row) and its geometry.
type mandalaLayout struct {
	cx, cy  int
	g       mapstyle.MandalaGeom
	n, w, s int // what it was laid out for: rings, columns, scene rows
}

// mandalaFor lays the mandala of n rings out in the middle of a pane of W
// columns and S scene rows, as large as the pane holds (cached: a frame
// only draws it).
func (v *view) mandalaFor(n, W, S int) *mandalaLayout {
	if l := v.md; l != nil && l.n == n && l.w == W && l.s == S {
		return l
	}
	cx, cy := W/2, (S-1)/2
	v.md = &mandalaLayout{cx: cx, cy: cy, n: n, w: W, s: S,
		g: mapstyle.LayMandala(n, float64(min(cy, S-1-cy)), float64(min(cx, W-1-cx)))}
	return v.md
}

// composeMandala builds the Transcendent Age's frame into v.fb: the void,
// a few stars, the mandala, the cursor on it, and the header and status
// line. There is no panorama, so no camera and no minimap strip: the
// mandala has every row between the two lines of chrome. It has two zoom
// levels: the whole mandala, and the close one, a window half as wide and
// half as tall on its middle at twice the size, which the arrows pan (there
// is nothing wider than the mandala for the far levels to show).
func (v *view) composeMandala(f mapstyle.Frame, W, H int) *scene {
	m := f.Model
	v.sw, v.worldW = W, m.Skyline.Width
	z := zMid
	if v.zoom == zClose {
		z = zClose
	}
	v.lastW = W
	world := &v.fb
	if z != zMid {
		world = &v.lfb
	}
	S := max(0, H-2)
	s := &scene{v: v, m: m, fb: world, tier: f.Tier, anim: f.Anim, visit: f.VisitFrame(), W: W, H: H, S: S, top: 1, sel: -1,
		zm: z, lw: W, ls: S, sw: W, sh: H, sS: S}
	s.groundY = s.S // no ground: it is sky to the last row
	s.p, s.mp = v.palettes(m)
	s.cam = v.cam // left as it was: there is nothing to scroll
	rings := m.Mandala()
	l := v.mandalaFor(len(rings), W, s.S)
	if v.inspect {
		v.resolve(m, v.targetsFor(m, W, v.cam), f.VisitFrame())
		if v.reveal && z == zClose { // the window goes to the cursor, when it is not in it
			v.revealMandala(m, l, rings, S, W)
		}
		v.reveal = false // nothing else scrolls
	}
	s.pan, s.offX = v.mandalaWindow(z, S, W)
	world.reset(W, H)
	o := newOrb(s, mapmodel.SkyMandala)
	o.void()
	o.mandalaStars()
	o.drawMandala(l.g, rings, l.cx, l.cy, s.S)
	o.catastrophe()
	if v.inspect {
		o.mandalaCursor(l, rings)
	}
	o.scene = s.present()
	o.panBar()
	o.chrome()
	return o.scene
}

// mandalaWindow is the close level's window over the mandala frame: how many
// rows it is raised by and how many columns in it starts. It starts on the
// middle, and the pan the keys keep is relative to that.
func (v *view) mandalaWindow(z zoom, S, W int) (pan, offX int) {
	win := (W + 1) / 2
	hiX := W - win
	hiY := S - 1 - (S-1)/2
	v.panLo, v.panHi = -hiY/2, hiY-hiY/2
	v.panXLo, v.panXHi = -hiX/2, hiX-hiX/2
	if z != zClose {
		return 0, 0
	}
	v.pan = clampInt(v.pan, v.panLo, v.panHi)
	v.panX = clampInt(v.panX, v.panXLo, v.panXHi)
	return hiY/2 + v.pan, hiX/2 + v.panX
}

// revealMandala moves the close window to the cursor's cell when it is not
// in view (relative to the middle, which is where the window starts).
func (v *view) revealMandala(m *mapmodel.Model, l *mandalaLayout, rings []mapmodel.MandalaRing, S, W int) {
	dx, dy, ok := mandalaSpot(&l.g, m, rings, v.cur)
	if !ok {
		return
	}
	win := (W + 1) / 2
	hiX, hiY := W-win, S-1-(S-1)/2
	pan, offX := hiY/2+v.pan, hiX/2+v.panX
	cx, cy := l.cx+dx, l.cy+dy
	top, bottom := S-1-pan-(S-1)/2, S-1-pan
	if cx < offX+2 || cx >= offX+win-2 {
		offX = clampInt(cx-win/2, 0, hiX)
	}
	if cy < top+1 || cy > bottom-1 {
		pan = clampInt(S-1-cy-(S-1)/4, 0, hiY)
	}
	v.pan, v.panX = pan-hiY/2, offX-hiX/2
}

// mandalaStars scatters a few faint stars over the pane: points of light
// only, a few of them brightening for a moment. No star changes its glyph.
func (o *orb) mandalaStars() {
	void := o.c(mapmodel.InkVoid, iBack, 0)
	dim, bright := o.c(mapmodel.InkStar, iEmit, 0), o.c(mapmodel.InkStarBright, iEmit, 0)
	for y := 0; y < o.S; y++ {
		for x := 0; x < o.W; x++ {
			hv := hashf(x, y, 419)
			if hv >= 0.01 {
				continue
			}
			ch, col := '·', theme.Mix(void, dim, 0.4+0.6*hashf(x, y, 420))
			if hv < 0.002 {
				ch, col = '∙', dim
			}
			if hash(x, y, o.anim/6)%11 == 0 {
				col = bright
			}
			o.fb.fg(x, o.Y(y), ch, col, dStar)
		}
	}
}

// mandalaKey is HandleKey in the Transcendent Age. There is no panorama to
// scroll, so the camera stays where it was (and follows the present again
// after a prestige): Tab and the arrows put the cursor out and move it,
// round a ring and from ring to ring; Home takes it to the core; Esc puts
// it away. + and - (PgDn and PgUp) close in on the middle of the mandala
// and back; while it is close in and the cursor is away the arrows pan the
// window over it (Shift for half a window) and Home brings it back to the
// middle first.
func (v *view) mandalaKey(ev *tcell.EventKey, f mapstyle.Frame) bool {
	m, key := f.Model, ev.Key()
	shift := ev.Modifiers()&tcell.ModShift != 0
	move := func(dx, drow, tab int) {
		if !v.inspect { // the first key only puts the cursor out
			v.inspect, dx, drow, tab = true, 0, 0, 0
		}
		v.step(m, f.VisitFrame(), dx, drow, tab)
	}
	closeUp := v.zoom == zClose
	panning := closeUp && !v.inspect && (v.panHi > v.panLo || v.panXHi > v.panXLo)
	pan := func(dx, dy int) {
		wy, wx := max(1, (v.panHi-v.panLo+1)/2), max(1, (v.panXHi-v.panXLo+1)/2)
		if !shift {
			wy, wx = panStep, panStep
		}
		v.pan = clampInt(v.pan+dy*wy, v.panLo, v.panHi)
		v.panX = clampInt(v.panX+dx*wx, v.panXLo, v.panXHi)
	}
	switch {
	case key == tcell.KeyPgDn || key == tcell.KeyRune && (ev.Rune() == '+' || ev.Rune() == '='):
		v.setMandalaZoom(zClose)
	case key == tcell.KeyPgUp || key == tcell.KeyRune && (ev.Rune() == '-' || ev.Rune() == '_'):
		v.setMandalaZoom(zMid)
	case key == tcell.KeyTab || key == tcell.KeyBacktab:
		d := 1
		if key == tcell.KeyBacktab || shift {
			d = -1
		}
		move(0, 0, d)
	case key == tcell.KeyLeft:
		if panning {
			pan(-1, 0)
			return true
		}
		move(-1, 0, 0)
	case key == tcell.KeyRight:
		if panning {
			pan(1, 0)
			return true
		}
		move(1, 0, 0)
	case key == tcell.KeyUp:
		if panning {
			pan(0, 1)
			return true
		}
		move(0, 1, 0)
	case key == tcell.KeyDown:
		if panning {
			pan(0, -1)
			return true
		}
		move(0, -1, 0)
	case key == tcell.KeyHome:
		if closeUp && (v.pan != 0 || v.panX != 0) { // the window comes back to the middle first
			v.pan, v.panX = 0, 0
			return true
		}
		v.inspect, v.cur = true, target{kind: tCore}
	case key == tcell.KeyEscape:
		if !v.inspect {
			return false
		}
		v.inspect = false
	default:
		return false
	}
	return true
}

// setMandalaZoom closes in on the mandala (zClose) or goes back to the whole
// of it (anything else): the window starts on the middle.
func (v *view) setMandalaZoom(z zoom) {
	if z != zClose {
		z = zMid
	}
	if v.zoom == z || v.zoom != zClose && z == zMid {
		v.zoom = z // a level the mandala does not have is the whole of it
		return
	}
	v.zoom = z
	v.pan, v.panX = 0, 0
	v.reveal = v.inspect
}

// breath is how brightly the wave of light lights ring k of n at this frame
// (the core is -1): a crest rolling outward ring by ring.
func (o *orb) breath(k, n int) float64 { return mapstyle.Breath(o.anim, k, n) }

// mdTurn is a whole turn round a ring, in the units a target's place on
// its ring is counted in.
const mdTurn = 3600

// mandalaTargets lists what the cursor can stand on in Tab order: the
// core, the crown's building types, then each ring outward with the types
// still standing from its era, clockwise from the top (a ring with nothing
// left standing is a target itself). A target's row is its ring (the core
// 0, the crown 1) and its place the way round, so the arrows step round a
// ring and from ring to ring.
func mandalaTargets(m *mapmodel.Model) []tgt {
	out := []tgt{{target: target{kind: tCore}}}
	crown := m.Crown()
	for j, mk := range crown {
		out = append(out, tgt{target{kind: tMark, key: mk.Key}, 1, j * mdTurn / len(crown)})
	}
	for k, r := range m.Mandala() {
		if len(r.Marks) == 0 {
			out = append(out, tgt{target{kind: tRing, cp: r.Epoch}, k + 2, 0})
		}
		for j, mk := range r.Marks {
			out = append(out, tgt{target{kind: tMark, key: mk.Key}, k + 2, j * mdTurn / len(r.Marks)})
		}
	}
	return out
}

// mandalaInspection says what a target of the mandala is: the core, a ring
// with nothing left standing, or a building type (what the cursor says of
// it anywhere, under the ring or the crown it shines on).
func mandalaInspection(m *mapmodel.Model, t tgt) (in inspectionData, ok bool) {
	switch t.kind {
	case tCore:
		return inspectionData{title: vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.01", "The core"), cmd: m.SquareCommand(),
			lines: []string{vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.02", "you, no longer bound to form"), vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.03", "every age you lived through, in light")}}, true
	case tRing:
		if t.cp < 0 || t.cp >= len(m.Catalog.EpochName) {
			return in, false
		}
		return inspectionData{title: vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.04", "Ring of the ") + m.Catalog.EpochName[t.cp], cmd: mapmodel.CmdStatus,
			lines: []string{vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.05", "nothing of it still stands"), vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.06", "the ring remembers")}}, true
	case tMark:
		in = lotInspection(m, t.key, 0)
		if in.title == "" {
			return in, false
		}
		where, marks := vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.07", "a petal of the crown"), m.Crown()
		if rings := m.Mandala(); t.row >= 2 && t.row-2 < len(rings) {
			if r := rings[t.row-2]; r.Epoch < len(m.Catalog.EpochName) {
				where, marks = vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.08", "a light on the ring of the ")+m.Catalog.EpochName[r.Epoch], r.Marks
			}
		}
		for j, mk := range marks { // which of the ring's lights it is
			if mk.Key == t.key {
				where += " (" + strconv.Itoa(j+1) + vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaInspection.09", " of ") + strconv.Itoa(len(marks)) + ")"
			}
		}
		in.lines = append([]string{where}, in.lines...)
		return in, true
	}
	return in, false
}

// mandalaSpot is the cell of the layout (as an offset from the core) that
// stands for target t: the core, the top of an empty ring, or a building
// type's place among its ring's cells or the crown's petals, spread evenly
// round. ok is false for a ring the pane had no room for.
func mandalaSpot(g *mapstyle.MandalaGeom, m *mapmodel.Model, rings []mapmodel.MandalaRing, t target) (dx, dy int, ok bool) {
	spot := func(k, j, n int) (int, int, bool) {
		if k -= len(rings) - g.Rings; k < 0 || k >= g.Rings {
			return 0, 0, false
		}
		cells := g.RingCells(k)
		if len(cells) == 0 {
			return 0, 0, false
		}
		c := cells[j*len(cells)/max(1, n)]
		return c.DX, c.DY, true
	}
	switch t.kind {
	case tCore:
		return 0, 0, true
	case tRing:
		for k, r := range rings {
			if r.Epoch == t.cp {
				return spot(k, 0, 1)
			}
		}
	case tMark:
		crown := m.Crown()
		for j, mk := range crown {
			if mk.Key != t.key {
				continue
			}
			ps := g.PetalCells()
			if len(ps) == 0 {
				return 0, 0, true // no room for a crown: it is with the core
			}
			p := ps[j*len(ps)/len(crown)]
			return p.DX, p.DY, true
		}
		for k, r := range rings {
			for j, mk := range r.Marks {
				if mk.Key == t.key {
					return spot(k, j, len(r.Marks))
				}
			}
		}
	}
	return 0, 0, false
}

// mandalaCursor marks the cell the cursor's target stands on: its glyph in
// the cursor's colours.
func (o *orb) mandalaCursor(l *mandalaLayout, rings []mapmodel.MandalaRing) {
	dx, dy, ok := mandalaSpot(&l.g, o.m, rings, o.v.cur)
	if !ok || l.cy+dy < 0 || l.cy+dy >= o.S {
		return
	}
	x, y := l.cx+dx, o.Y(l.cy+dy)
	c := o.fb.at(x, y)
	if c == nil {
		return
	}
	ch := c.ch
	if ch == ' ' || ch == 0 || ch == '█' {
		ch = mapmodel.R(mapmodel.SymPetal, o.tier)
	}
	cfg, cbg, _ := o.mp.CursorStyle(o.anim).Decompose()
	o.fb.set(x, y, ch, cfg, cbg, dTop)
}

// mandalaLegend is the legend line: what the mandala is made of.
func (o *orb) mandalaLegend() []seg {
	glyphs := ""
	for _, r := range o.m.Mandala() {
		glyphs += string(mapmodel.R(mapmodel.EraSym(r.Epoch), o.tier)) + " "
	}
	return []seg{
		{" " + string(mapmodel.R(mapmodel.SymCore, o.tier)) + " ", theme.RoleText}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaLegend.01", "the core: you, beyond form  "), theme.RoleDim},
		{string(mapmodel.R(mapmodel.SymPetal, o.tier)) + " ", theme.RoleText}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaLegend.02", "the crown: this age's works  "), theme.RoleDim},
		{glyphs, theme.RoleText}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaLegend.03", "a ring for each era you lived, oldest innermost  "), theme.RoleDim},
		{vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaLegend.04", "the wave: "), theme.RoleText}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaLegend.05", "light, rolling outward"), theme.RoleDim},
	}
}

// mandalaHints is the status line at rest: the keys, and where you are.
func (o *orb) mandalaHints() []seg {
	key, dim := theme.RoleAccent, theme.RoleDim
	tail := []seg{{"│ ", dim}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.01", "beyond form, in light"), theme.RoleLabel}}
	return fitHints(o.W, tail,
		[]seg{{" Tab ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.03", "inspect  "), dim}, {"◄► ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.04", "round a ring  "), dim}, {"↑↓ ", key},
			{vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.05", "ring to ring  "), dim}, {"Home ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.07", "the core  "), dim}, {"map flows ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.08", "overlay  "), dim}},
		[]seg{{" Tab ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.10", "inspect  "), dim}, {"◄► ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.11", "round a ring  "), dim}, {"↑↓ ", key},
			{vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.12", "ring to ring  "), dim}, {"Home ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.14", "the core  "), dim}},
		[]seg{{" Tab ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.16", "inspect  "), dim}, {"◄► ↑↓ ", key}, {vibe.Line("ui.mapstyle.skyline.orbit_mandala.mandalaHints.17", "move  "), dim}},
		[]seg{{" Tab ◄► ↑↓  ", key}},
	)
}

// drawMandala draws layout g of rings centred on (cx, cy), above row
// limit: the void under it cleared of stars, the core's gold glow, the
// rings' bands (in their era's glyph where there is room, else in
// half-block pixels, two to a cell, so neighbouring rings keep a dark gap
// between them), the beads, the crown and the core.
func (o *orb) drawMandala(g mapstyle.MandalaGeom, rings []mapmodel.MandalaRing, cx, cy, limit int) {
	void := o.c(mapmodel.InkVoid, iBack, 0)
	// the light, made to stand off the void on every page (a light theme's
	// void is pale, and its light is ink)
	gold := theme.Legible(o.c(mapmodel.InkAccent, iEmit, 0), void, 3)
	white := theme.Legible(o.c(mapmodel.InkLight, iEmit, 0), void, 3)
	burn := theme.Mix(void, gold, 0.55) // the core's cell burns gold
	core := theme.Legible(o.c(mapmodel.InkGlow, iEmit, 0), burn, 3)
	inDisc := func(x, y int, r float64) bool {
		dx, dy := float64(x-cx)/mapstyle.MandalaAspect, float64(y-cy)
		return float64(dx*dx)+float64(dy*dy) < float64(r*r)
	}
	total := 0 // everything built in this last age
	for _, mk := range o.m.Crown() {
		total += mk.Count
	}
	crown := total > 0
	glow := 0.0 // the core's light spreads as the crown grows
	if crown && g.Petals > 0 {
		glow = math.Min(g.PetalR+1, 1.2+float64(0.45*mapmodel.Log2(1+float64(total))))
	}
	ext := g.Extent + 1.5
	for y := max(0, cy-int(ext)-1); y <= min(limit-1, cy+int(ext)+1); y++ {
		for x := max(0, cx-int(2*ext)-1); x <= min(o.W-1, cx+int(2*ext)+1); x++ {
			switch {
			case x == cx && y == cy:
				o.fb.set(x, o.Y(y), ' ', void, burn, dMandala)
			case inDisc(x, y, glow): // fading out from the core
				dx, dy := float64(x-cx)/mapstyle.MandalaAspect, float64(y-cy)
				fade := 1 - math.Sqrt(float64(dx*dx)+float64(dy*dy))/glow
				o.fb.set(x, o.Y(y), ' ', void, theme.Mix(void, gold, 0.06+float64(0.3*fade)), dMandala)
			case inDisc(x, y, ext):
				o.fb.set(x, o.Y(y), ' ', void, void, dMandala)
			}
		}
	}
	n := len(rings)
	off := n - g.Rings
	ring := make([]tcell.Color, g.Rings) // each ring's colour this frame: its era's, flashing white in the wave
	for k := range ring {
		era := theme.Legible(o.resolve(mapstyle.EraLight(rings[k+off].Epoch), iEmit, 0), void, 3)
		ring[k] = theme.Mix(era, white, 0.8*o.breath(k+off, n))
	}
	if !g.Lines && g.Rings > 0 && o.tier != mapmodel.TierASCII {
		o.mandalaPixels(g, ring, cx, cy, limit)
	}
	petal := gold
	if !crown {
		petal = white // the crown still to be raised
	}
	bc := o.breath(-1, n)
	for _, mc := range g.Cells {
		x, y := cx+mc.DX, cy+mc.DY
		if y < 0 || y >= limit {
			continue
		}
		var ch rune
		var c tcell.Color
		switch mc.Part {
		case mapstyle.MdCore:
			ch, c = mapmodel.R(mapmodel.SymCore, o.tier), core
		case mapstyle.MdRay:
			ch, c = mapstyle.RayRune(mc.DX, mc.DY), theme.Mix(gold, white, bc)
		case mapstyle.MdPetal:
			ch, c = mapmodel.R(mapmodel.SymPetal, o.tier), theme.Mix(petal, white, bc)
			if mc.Tip {
				ch = mapstyle.PetalTip(mc.DX, mc.DY)
			}
		case mapstyle.MdLine:
			ch, c = mapmodel.R(mapmodel.EraSym(rings[mc.Ring+off].Epoch), o.tier), ring[mc.Ring]
		case mapstyle.MdBead:
			ch, c = mapmodel.R(mapmodel.EraSym(rings[mc.Ring+off].Epoch), o.tier), theme.Mix(ring[mc.Ring], white, 0.3)
		}
		o.fb.fg(x, o.Y(y), ch, c, dMandala)
	}
}

// mandalaPixels draws the rings as thin bands of half-block pixels (a
// pixel is half a row tall and a column wide, so they come out round),
// leaving the cells of the layout itself to its glyphs. Rings stand at
// least two pixels apart, so no cell ever holds two of them.
func (o *orb) mandalaPixels(g mapstyle.MandalaGeom, ring []tcell.Color, cx, cy, limit int) {
	taken := map[[2]int]bool{}
	for _, mc := range g.Cells {
		taken[[2]int{mc.DX, mc.DY}] = true
	}
	reach := int(g.Extent) + 1
	for dy := -reach; dy <= reach; dy++ {
		y := cy + dy
		if y < 0 || y >= limit {
			continue
		}
		for dx := -2 * reach; dx <= 2*reach; dx++ {
			if taken[[2]int{dx, dy}] {
				continue
			}
			up, lo := pixelRing(g, dx, dy, 0), pixelRing(g, dx, dy, 1)
			switch {
			case up >= 0 && lo >= 0:
				o.fb.fg(cx+dx, o.Y(y), '█', ring[up], dMandala)
			case up >= 0:
				o.fb.fg(cx+dx, o.Y(y), '▀', ring[up], dMandala)
			case lo >= 0:
				o.fb.fg(cx+dx, o.Y(y), '▄', ring[lo], dMandala)
			}
		}
	}
}

// pixelRing is the ring lit in the upper (h 0) or lower (h 1) half of the
// cell at (dx, dy) from the core, or -1: a ring is a band a pixel wide.
func pixelRing(g mapstyle.MandalaGeom, dx, dy, h int) int {
	px, py := float64(dx), float64(2*dy+h)-0.5
	d := math.Sqrt(float64(px*px) + float64(py*py))
	for k, r := range g.Radii {
		if math.Abs(d-2*r) < 0.55 {
			return k
		}
	}
	return -1
}
