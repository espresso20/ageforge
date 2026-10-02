package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_mandala.go is the Transcendent Age: beyond matter, geometry. White
// and gold light on deep indigo. Over the baseline, anchored to the view,
// a mandala built from the civilization's history breathes slowly: one ring
// per era, the Stone Era innermost and the Cosmic Era outermost, each ring
// carrying a mark for every building type raised in its ages (in the
// symbol it had then), spokes of light between the rings, and at the heart
// the Transcendent Age's own buildings round a bright core. The lots are
// pure light, pillars, diamonds and circles on a thin mirror line with
// their faint reflection below it. Nothing moves but a mote of light or
// two, drifting home.

// mandalaBackdrop draws everything behind the lots.
func (o *orb) mandalaBackdrop() {
	o.stars(0, o.groundY, 0.004, 0, 419)
	o.mandala()
	o.mirror()
}

// mandalaGeom is the mandala's centre (column, row) and outer radius in
// rows; a row is two columns wide, so the rings are round.
func (o *orb) mandalaGeom() (float64, float64, float64) {
	gy := o.groundY
	ry := math.Min(float64(gy-1)*0.5, float64(o.W)/4.2)
	return float64(o.W) / 2, float64(gy-1) * 0.5, math.Max(2, ry)
}

// breath is how bright ring k is at this frame: a slow pulse (about twelve
// seconds) that runs outward ring by ring.
func (o *orb) breath(k int) float64 {
	return 0.5 + 0.5*mapmodel.Sin(float64(o.anim-7*k)/96)
}

// mandala draws the rings of the eras as thin lines of light in half-block
// pixels, spokes between every other pair, each era's marks spread round
// its ring, and the crown at the heart round a bright core.
func (o *orb) mandala() {
	cx, cy, R := o.mandalaGeom()
	rings := o.m.Mandala()
	if len(rings) == 0 {
		return
	}
	void := o.c(mapmodel.InkVoid, iBack, 0)
	dim := o.c(mapmodel.InkFrameDim, iEmit, 0)
	gold := o.c(mapmodel.InkAccent, iEmit, 0)
	white := o.c(mapmodel.InkLight, iEmit, 0)
	core := o.c(mapmodel.InkGlow, iEmit, 0)
	n := len(rings)
	Rp := 2 * R // the outer radius in pixels (a column is as wide as half a row is tall)
	rad := func(k int) float64 { return Rp * (0.24 + 0.76*float64(k+1)/float64(n)) }
	rot := func(k int) float64 {
		dir := 1.0
		if k%2 == 1 {
			dir = -1
		}
		return dir * float64(o.anim) / float64(2400+400*k)
	}
	ink := make([]tcell.Color, n)
	for k := range ink {
		c := white
		if k%2 == 0 {
			c = gold
		}
		ink[k] = theme.Mix(theme.Mix(void, dim, 0.5), c, 0.12+0.42*o.breath(k))
	}
	cpy := 2*cy + 1
	o.pixels(int(cx-Rp)-1, int(cx+Rp)+1, int((cpy-Rp)/2)-1, min(o.groundY-1, int((cpy+Rp)/2)+1), dMandala,
		func(x, py int) (tcell.Color, bool) {
			dx, dy := float64(x)+0.5-cx, float64(py)+0.5-cpy
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			if d > Rp+0.6 {
				return 0, false
			}
			for k := 0; k < n; k++ {
				r := rad(k)
				if math.Abs(d-r) < 0.55 {
					return ink[k], true
				}
				if k%2 == 0 && k < n-1 && d > r+0.6 && d < rad(k+1)-0.6 {
					// spokes out to the next ring, turning with it
					a := turns(dx, dy) - rot(k)*0.5
					f := a*12 - math.Floor(a*12)
					if f < 0.06 || f > 0.94 {
						return theme.Mix(void, ink[k], 0.6), true
					}
				}
			}
			return 0, false
		})
	at := func(r, t float64) (int, int) {
		return int(math.Round(cx + r*mapmodel.Cos(t))), int(math.Round((cpy - r*mapmodel.Sin(t) - 0.5) / 2))
	}
	put := func(x, y int, ch rune, c tcell.Color) {
		if y >= 0 && y < o.groundY {
			o.fb.fg(x, o.Y(y), ch, c, dMandala)
		}
	}
	// the eras' marks round their rings
	for k := 0; k < n; k++ {
		b := o.breath(k)
		marks := rings[k].Marks
		for i, mk := range marks {
			x, y := at(rad(k), float64(i)/float64(len(marks))+rot(k))
			c := white
			if k%2 == 0 {
				c = gold
			}
			if mk.Wonder {
				c = theme.Mix(gold, core, 0.4)
			}
			put(x, y, mapmodel.R(mk.Sym, o.tier), theme.Mix(theme.Mix(void, c, 0.65), c, b))
		}
	}
	// petals round the heart, the crown and the core
	r0 := rad(0) * 0.62
	b := o.breath(-2)
	for j := 0; j < 8; j++ {
		x, y := at(r0, float64(j)/8-float64(o.anim)/3000)
		put(x, y, '◇', theme.Mix(theme.Mix(void, gold, 0.5), gold, b))
	}
	crown := o.m.Crown()
	for i, mk := range crown {
		x, y := at(r0*0.55, float64(i)/float64(max(1, len(crown)))+float64(o.anim)/2000)
		put(x, y, mapmodel.R(mk.Sym, o.tier), theme.Mix(white, core, b))
	}
	ix, iy := int(math.Round(cx)), int(math.Floor(cy))
	put(ix, iy, '✦', theme.Mix(white, core, 0.5+0.5*b))
	halo := theme.Mix(void, white, 0.3+0.35*b)
	put(ix-1, iy, '·', halo)
	put(ix+1, iy, '·', halo)
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
// a slow ripple.
func (o *orb) reflect() {
	gy := o.groundY
	void := o.c(mapmodel.InkVoid, iBack, 0)
	depth := min(o.S-1-gy, gy-1)
	for k := 1; k <= depth; k++ {
		a := 0.62 + 0.3*float64(k)/float64(max(1, depth))
		dx := int(math.Round(0.8 * mapmodel.Sin(float64(k)/4+float64(o.anim)/40)))
		for x := 0; x < o.W; x++ {
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
