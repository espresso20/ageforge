package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// orbit_mandala.go is the Transcendent Age: beyond matter, geometry. White
// and gold light on deep indigo. Over the baseline, anchored to the view,
// stands the mandala both styles draw (mapstyle.LayMandala): a bright core,
// a crown of eight petals for the age's own buildings, and one ring for
// every era the player passed through, the Stone Era innermost, each strung
// with its era's glyph in its era's muted colour, every ring's beads on the
// same spokes. It is strictly symmetric about the core, on clear indigo.
// The lots are pure light, pillars, diamonds and circles on a thin mirror
// line, and the mirror holds a faint reflection of all of it. The light
// breathes outward ring by ring; a mote of light or two drifts home.

// mandalaBackdrop draws everything behind the lots.
func (o *orb) mandalaBackdrop() {
	o.stars(0, o.groundY, 0.004, 0, 419)
	o.mandala()
	o.mirror()
}

// mandalaGeom is the mandala's centre cell (column, scene row) and the
// rows and columns it may reach round it: the sky over the mirror line, in
// the middle of the view (the far objects keep out of it: clearOfMandala).
func (o *orb) mandalaGeom() (cx, cy int, ry, rx float64) {
	gy := o.groundY
	cx, cy = o.W/2, (gy-1)/2
	return cx, cy, float64(min(cy, gy-1-cy)), float64(min(cx, o.W-1-cx, mandalaZone(o.W)-3))
}

// breath is how brightly the wave of light lights ring k of n at this frame
// (the core is -1): a crest rolling outward ring by ring.
func (o *orb) breath(k, n int) float64 { return mapstyle.Breath(o.anim, k, n) }

// mandala draws the mandala: the sky under it cleared of stars, then each
// ring's dotted line and beads, the crown and the core.
func (o *orb) mandala() {
	rings := o.m.Mandala()
	cx, cy, ry, rx := o.mandalaGeom()
	g := mapstyle.LayMandala(len(rings), ry, rx)
	o.drawMandala(g, rings, cx, cy, o.groundY)
}

// drawMandala draws layout g of rings centred on (cx, cy), above row
// limit: the sky under it cleared of stars, the core's gold glow, the
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
	crown := len(o.m.Crown()) > 0
	ext := g.Extent + 1.5
	for y := max(0, cy-int(ext)-1); y <= min(limit-1, cy+int(ext)+1); y++ {
		for x := max(0, cx-int(2*ext)-1); x <= min(o.W-1, cx+int(2*ext)+1); x++ {
			switch {
			case x == cx && y == cy:
				o.fb.set(x, o.Y(y), ' ', void, theme.Mix(void, gold, 0.8), dMandala) // the core burns gold
			case crown && g.Petals > 0 && inDisc(x, y, 1.9):
				o.fb.set(x, o.Y(y), ' ', void, theme.Mix(void, gold, 0.25), dMandala) // and glows with the crown
			case inDisc(x, y, ext):
				o.fb.set(x, o.Y(y), ' ', void, void, dMandala)
			}
		}
	}
	n := len(rings)
	off := n - g.Rings
	ring := make([]tcell.Color, g.Rings) // each ring's colour this frame: its era's, flashing white in the wave
	for k := range ring {
		ring[k] = theme.Mix(o.resolve(mapstyle.EraLight(rings[k+off].Epoch), iEmit, 0), white, 0.8*o.breath(k+off, n))
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
			ch, c = mapmodel.R(mapmodel.EraSym(rings[mc.Ring+off].Epoch), o.tier), theme.Mix(ring[mc.Ring], void, 0.15)
		case mapstyle.MdBead:
			ch, c = mapmodel.R(mapmodel.EraSym(rings[mc.Ring+off].Epoch), o.tier), theme.Mix(ring[mc.Ring], white, 0.2)
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

// mirror is the thin line of light the lots stand on.
func (o *orb) mirror() {
	gy := o.groundY
	gold := o.c(mapmodel.InkAccent, iEmit, 1)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	glint := o.c(mapmodel.InkGlow, iEmit, 0)
	for x := 0; x < o.W; x++ {
		wx := x + o.cam
		if wx < o.x0 || wx > o.x1 {
			continue
		}
		c := theme.Mix(void, gold, 0.6)
		if d := (wx - o.anim/3) % 40; d >= 0 && d < 3 {
			c = theme.Mix(gold, glint, 0.6)
		}
		o.fb.fg(x, o.Y(gy), '─', c, dMirror)
	}
}

// reflectRunes turns a glyph upside down for the reflection.
var reflectRunes = map[rune]rune{'▀': '▄', '▄': '▀', '◢': '◥', '◥': '◢', '◣': '◤', '◤': '◣', '▲': '▼', '▼': '▲',
	'╱': '╲', '╲': '╱', '┴': '┬', '┬': '┴', '▖': '▘', '▘': '▖', '▗': '▝', '▝': '▗', '└': '┌', '┌': '└',
	'┘': '┐', '┐': '┘', '╩': '╦', '╦': '╩', '╚': '╔', '╔': '╚', '╝': '╗', '╗': '╝'}

// reflect mirrors the lots under the mirror line, faint and fading, with
// a slow ripple, and the mandala behind them, fainter and still (a ripple
// would set its rings wobbling).
func (o *orb) reflect() {
	gy := o.groundY
	void := o.c(mapmodel.InkVoid, iBack, 0)
	depth := min(o.S-1-gy, gy-1)
	for k := 1; k <= depth; k++ {
		a := 0.62 + 0.3*float64(k)/float64(max(1, depth))
		dx := int(math.Round(0.8 * mapmodel.Sin(float64(k)/4+float64(o.anim)/40)))
		for x := 0; x < o.W; x++ {
			if src := o.fb.at(x, o.Y(gy-k)); src != nil && src.d == dMandala && src.ch != ' ' {
				ch := src.ch
				if r, ok := reflectRunes[ch]; ok {
					ch = r
				}
				o.fb.fg(x, o.Y(gy+k), ch, theme.Mix(src.fg, void, math.Min(0.92, a+0.12)), dReflect)
				continue
			}
			src := o.fb.at(clampInt(x+dx, 0, o.W-1), o.Y(gy-k))
			if src == nil || src.d < dWonder || src.d > dRow2 {
				continue
			}
			ch := src.ch
			if r, ok := reflectRunes[ch]; ok {
				ch = r
			}
			fg, bg := theme.Mix(src.fg, void, a), theme.Mix(src.bg, void, a)
			if ch == '█' {
				o.fb.set(x, o.Y(gy+k), ' ', fg, fg, dReflect)
				continue
			}
			o.fb.set(x, o.Y(gy+k), ch, fg, bg, dReflect)
		}
	}
}
