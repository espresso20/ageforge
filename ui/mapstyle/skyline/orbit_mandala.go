package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
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
// mandala has every row between the two lines of chrome.
func (v *view) composeMandala(f mapstyle.Frame, W, H int) *scene {
	m := f.Model
	v.lastW = W
	s := &scene{v: v, m: m, fb: &v.fb, tier: f.Tier, anim: f.Anim, W: W, H: H, S: max(0, H-2), top: 1, sel: -1}
	s.groundY = s.S // no ground: it is sky to the last row
	s.band = bandOf(m.AgeIdx)
	s.p, s.mp = v.palettes(m)
	v.cam = clampInt(v.cam, 0, max(0, m.Skyline.Width-W))
	s.cam = v.cam
	if v.inspect {
		v.resolve(m, v.targetsFor(m, W, v.cam), f.Anim)
		v.reveal = false // nothing scrolls
	}
	v.fb.reset(W, H)
	o := newOrb(s, mapmodel.SkyMandala)
	o.void()
	o.mandalaStars()
	rings := m.Mandala()
	l := v.mandalaFor(len(rings), W, s.S)
	o.drawMandala(l.g, rings, l.cx, l.cy, s.S)
	o.catastrophe()
	if v.inspect {
		o.mandalaCursor(l, rings)
	}
	o.chrome()
	return s
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
		return inspectionData{title: "The core", cmd: m.SquareCommand(),
			lines: []string{"you, no longer bound to form", "every age you lived through, in light"}}, true
	case tRing:
		if t.cp < 0 || t.cp >= len(m.Catalog.EpochName) {
			return in, false
		}
		return inspectionData{title: "Ring of the " + m.Catalog.EpochName[t.cp], cmd: mapmodel.CmdStatus,
			lines: []string{"nothing of it still stands", "the ring remembers"}}, true
	case tMark:
		in = lotInspection(m, t.key, 0)
		if in.title == "" {
			return in, false
		}
		where := "a petal of the crown"
		if rings := m.Mandala(); t.row >= 2 && t.row-2 < len(rings) {
			if e := rings[t.row-2].Epoch; e < len(m.Catalog.EpochName) {
				where = "a light on the ring of the " + m.Catalog.EpochName[e]
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
	o.fb.set(x, y, ch, o.mp.Bg, o.mp.Fg[mapmodel.CAccent], dTop)
}

// mandalaLegend is the legend line: what the mandala is made of.
func (o *orb) mandalaLegend() []seg {
	glyphs := ""
	for _, r := range o.m.Mandala() {
		glyphs += string(mapmodel.R(mapmodel.EraSym(r.Epoch), o.tier)) + " "
	}
	return []seg{
		{" " + string(mapmodel.R(mapmodel.SymCore, o.tier)) + " ", theme.RoleText}, {"the core: you, beyond form  ", theme.RoleDim},
		{string(mapmodel.R(mapmodel.SymPetal, o.tier)) + " ", theme.RoleText}, {"the crown: this age's works  ", theme.RoleDim},
		{glyphs, theme.RoleText}, {"a ring for each era you lived, oldest innermost  ", theme.RoleDim},
		{"the wave: ", theme.RoleText}, {"light, rolling outward", theme.RoleDim},
	}
}

// mandalaHints is the status line at rest: the keys, and where you are.
func (o *orb) mandalaHints() []seg {
	return []seg{{" Tab ", theme.RoleAccent}, {"inspect  ", theme.RoleDim}, {"◄► ", theme.RoleAccent},
		{"round a ring  ", theme.RoleDim}, {"▲▼ ", theme.RoleAccent}, {"ring to ring  ", theme.RoleDim},
		{"map flows ", theme.RoleAccent}, {"flows  ", theme.RoleDim}, {"│ ", theme.RoleDim},
		{"beyond form, in light", theme.RoleLabel}}
}

// drawMandala draws layout g of rings centred on (cx, cy), above row
// limit: the void under it cleared of stars, the core's gold glow, the
// rings' bands (in their era's glyph where there is room, else in
// half-block pixels, two to a cell, so neighbouring rings keep a dark gap
// between them), the beads, the crown and the core.
func (o *orb) drawMandala(g mapstyle.MandalaGeom, rings []mapmodel.MandalaRing, cx, cy, limit int) {
	void := o.c(mapmodel.InkVoid, iBack, 0)
	gold := o.c(mapmodel.InkAccent, iEmit, 0)
	white := o.c(mapmodel.InkLight, iEmit, 0)
	core := o.c(mapmodel.InkGlow, iEmit, 0)
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
				o.fb.set(x, o.Y(y), ' ', void, theme.Mix(void, gold, 0.8), dMandala) // the core burns gold
			case inDisc(x, y, glow):
				o.fb.set(x, o.Y(y), ' ', void, theme.Mix(void, gold, 0.25), dMandala)
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
