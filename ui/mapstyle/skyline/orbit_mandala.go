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

// breath is how bright ring k of n is at this frame (the core is -1): one
// slow wave of light rolling outward ring by ring.
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
// limit, the stars cleared from under it.
func (o *orb) drawMandala(g mapstyle.MandalaGeom, rings []mapmodel.MandalaRing, cx, cy, limit int) {
	void := o.c(mapmodel.InkVoid, iBack, 0)
	gold := o.c(mapmodel.InkAccent, iEmit, 0)
	white := o.c(mapmodel.InkLight, iEmit, 0)
	core := o.c(mapmodel.InkGlow, iEmit, 0)
	dim := o.c(mapmodel.InkFrameDim, iEmit, 0)
	ext := g.Extent + 1.5
	for y := max(0, cy-int(ext)-1); y <= min(limit-1, cy+int(ext)+1); y++ {
		for x := max(0, cx-int(2*ext)-1); x <= min(o.W-1, cx+int(2*ext)+1); x++ {
			dx, dy := float64(x-cx)/mapstyle.MandalaAspect, float64(y-cy)
			if float64(dx*dx)+float64(dy*dy) < float64(ext*ext) {
				o.fb.set(x, o.Y(y), ' ', void, void, dMandala)
			}
		}
	}
	off := len(rings) - g.Rings
	era := make([]tcell.Color, g.Rings)
	for k := range era {
		era[k] = o.resolve(mapstyle.EraLight(rings[k+off].Epoch), iEmit, 0)
	}
	crown := len(o.m.Crown()) > 0
	bc := o.breath(-1, len(rings))
	for _, mc := range g.Cells {
		x, y := cx+mc.DX, cy+mc.DY
		if y < 0 || y >= limit {
			continue
		}
		var ch rune
		var c tcell.Color
		switch mc.Part {
		case mapstyle.MdCore:
			ch, c = mapmodel.R(mapmodel.SymCore, o.tier), theme.Mix(white, core, 0.5+0.5*bc)
		case mapstyle.MdPetal:
			ch, c = mapmodel.R(mapmodel.SymPetal, o.tier), theme.Mix(theme.Mix(void, gold, 0.6), gold, bc)
			if !crown {
				ch, c = '◇', theme.Mix(void, dim, 0.7) // the crown not yet raised
			}
		case mapstyle.MdLine:
			ch, c = '·', theme.Mix(void, era[mc.Ring], 0.3+0.35*o.breath(mc.Ring+off, len(rings)))
		case mapstyle.MdBead:
			e := rings[mc.Ring+off].Epoch
			ch = mapmodel.R(mapmodel.EraSym(e), o.tier)
			c = theme.Mix(theme.Mix(void, era[mc.Ring], 0.7), theme.Mix(era[mc.Ring], white, 0.3), o.breath(mc.Ring+off, len(rings)))
		}
		o.fb.fg(x, o.Y(y), ch, c, dMandala)
	}
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
