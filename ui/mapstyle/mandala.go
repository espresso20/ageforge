package mapstyle

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// mandala.go is the Transcendent Age's mandala as geometry both styles
// draw: a bright core (a star, when there is room), a crown of eight solid
// petals round it, and one ring per era the player passed through, the
// Stone Era innermost. A ring is a band of its era's glyph, cell after cell
// (only beads when the rings stand too close for a band), and every ring's
// beads sit on the same spokes, so the spokes show without a stroke of
// their own. A wave of light rolls out from the core ring by ring (Breath).
//
// The layout is strictly symmetric about the centre cell: each quadrant is
// worked out once and mirrored, so rounding can never make one side differ
// from the other. A terminal cell is about twice as tall as it is wide, so
// a ring of radius r rows spans 2r columns and comes out round.

// MandalaAspect is how many columns make the height of a row.
const MandalaAspect = 2.0

// MandalaPart is what a cell of the mandala is.
type MandalaPart uint8

const (
	MdCore  MandalaPart = iota + 1 // the core, the one centre cell
	MdRay                          // a ray of the core's star
	MdPetal                        // a petal of the crown (Tip: its outer point)
	MdLine                         // a ring's band, between its beads
	MdBead                         // a ring's bead, on a spoke
)

// MandalaCell is one cell of the mandala, as an offset from its centre.
type MandalaCell struct {
	DX, DY int
	Part   MandalaPart
	Ring   int  // the ring, 0 innermost (lines and beads)
	Idx    int  // a bead's or petal's number, clockwise from the top
	Tip    bool // a petal's outer point
}

// MandalaGeom is a mandala laid out for a space.
type MandalaGeom struct {
	Rings  int       // rings drawn
	Radii  []float64 // each ring's radius in rows
	Beads  []int     // each ring's beads
	Lines  bool      // the rings have dotted lines (there is room for them)
	Petals int       // petals in the crown (0: no room for a crown)
	PetalR float64   // the crown's radius in rows
	Extent float64   // the outermost ring's radius in rows
	Cells  []MandalaCell
}

// Spacing limits, in rows between neighbouring rings.
const (
	mdLineGap = 2.0  // a dotted line needs a clear row between rings
	mdBeadGap = 1.25 // beads alone need a row and a bit, or the rings jumble
	mdMaxGap  = 3.0  // a few rings keep together rather than fill the space
	// mdCrownMin is the smallest mandala (its outer radius in rows) with
	// room for the crown: a smaller one gives the space to the eras' rings.
	mdCrownMin = 8.0
	// mdGrandMin is the smallest with room for the grand crown: petals with
	// points and a star for a core.
	mdGrandMin = 16.0
)

// LayMandala lays out n rings round a centre cell with ry rows above and
// below it and rx columns either side. When the space cannot hold n rings
// apart it holds as many as it can (Rings < n): the caller keeps the
// newest eras.
func LayMandala(n int, ry, rx float64) MandalaGeom {
	var g MandalaGeom
	R := math.Min(ry, rx/MandalaAspect) - 0.5
	if R < 0.5 {
		g.Cells = []MandalaCell{{Part: MdCore}}
		return g
	}
	r0 := math.Min(mdLineGap, R) // a clear cell round the core
	grand := R >= mdGrandMin
	switch {
	case grand:
		g.PetalR, g.Petals = 3, 8
		r0 = g.PetalR + 3
	case R >= mdCrownMin:
		g.PetalR, g.Petals = clampF(0.16*R, 1.5, 3), 8
		r0 = g.PetalR + 2
	}
	rings := max(0, min(n, 1+int((R-r0)/mdBeadGap)))
	gap := 0.0
	if rings > 1 {
		gap = math.Min(mdMaxGap, (R-r0)/float64(rings-1))
	}
	g.Rings = rings
	g.Lines = r0 >= mdLineGap && (rings <= 1 || gap >= mdLineGap)
	for k := 0; k < rings; k++ {
		g.Radii = append(g.Radii, r0+float64(gap*float64(k)))
	}
	if rings > 0 {
		g.Extent = g.Radii[rings-1]
	}
	g.Cells = append(g.Cells, MandalaCell{Part: MdCore})
	rays := [][2]int{{1, 0}} // the core's star: a short ray either side
	if grand {
		rays = [][2]int{{0, 1}, {1, 1}, {1, 0}, {2, 0}} // and up, down and the diagonals
	}
	for _, p := range rays {
		if g.Petals > 0 {
			g.Cells = append(g.Cells, mirrorAt(p, MdRay, 0, 0)...)
		}
	}
	for j := 0; g.Petals > 0 && j <= g.Petals/4; j++ {
		t := float64(j) / float64(g.Petals)
		base := spokeAt(g.PetalR, t)
		g.Cells = append(g.Cells, mirrorIdx(base, MdPetal, 0, j, g.Petals)...)
		if tip := spokeAt(g.PetalR+0.7, t); grand && tip != base {
			for _, c := range mirrorIdx(tip, MdPetal, 0, j, g.Petals) {
				c.Tip = true
				g.Cells = append(g.Cells, c)
			}
		}
	}
	spacing := 5.0 // cells between beads on a lined ring
	if !g.Lines {
		spacing = 3
	}
	prev := 8
	for k, r := range g.Radii {
		path := quadrantPath(r)
		b := prev
		for b*2 <= int(perimeter(r)/spacing) {
			b *= 2
		}
		prev = b
		g.Beads = append(g.Beads, b)
		bead := map[[2]int]int{}
		for i := 0; i <= b/4; i++ {
			bead[nearest(path, spokeAt(r, float64(i)/float64(b)))] = i
		}
		for _, p := range path {
			i, isBead := bead[p]
			switch {
			case isBead:
				g.Cells = append(g.Cells, mirrorIdx(p, MdBead, k, i, b)...)
			case g.Lines:
				g.Cells = append(g.Cells, mirrorAt(p, MdLine, k, 0)...)
			}
		}
	}
	return g
}

// spokeAt is the first-quadrant offset (columns right, rows up) of the
// point at radius r rows on the spoke t turns clockwise from the top,
// folded into the first quadrant.
func spokeAt(r, t float64) [2]int {
	// the float64 conversions keep a product from fusing with the add (an
	// FMA on arm64), so every machine rounds to the same cells
	x, y := float64(MandalaAspect*r*mapmodel.Sin(t)), float64(r*mapmodel.Cos(t))
	return [2]int{int(math.Floor(math.Abs(x) + 0.5)), int(math.Floor(math.Abs(y) + 0.5))}
}

// quadrantPath is a ring's first quadrant as a thin 8-connected line of
// offsets (columns right, rows up), from the top to the right.
func quadrantPath(r float64) [][2]int {
	steps := int(16*(MandalaAspect+1)*r) + 8
	var ps [][2]int
	for i := 0; i <= steps; i++ {
		p := spokeAt(r, 0.25*float64(i)/float64(steps))
		if n := len(ps); n > 0 && ps[n-1] == p {
			continue
		}
		ps = append(ps, p)
	}
	// thin it: drop a cell whose neighbours already touch (never an end,
	// where the mirrored quadrants join)
	for i := 1; i < len(ps)-1; {
		a, c := ps[i-1], ps[i+1]
		if absI(a[0]-c[0]) <= 1 && absI(a[1]-c[1]) <= 1 {
			ps = append(ps[:i], ps[i+1:]...)
			continue
		}
		i++
	}
	return ps
}

// perimeter is about how many cells a ring of radius r rows runs through.
func perimeter(r float64) float64 { return 4 * float64(len(quadrantPath(r))-1) }

// nearest is the cell of path nearest offset p, measured round (a column
// is half a row).
func nearest(path [][2]int, p [2]int) [2]int {
	best, bd := path[0], math.Inf(1)
	for _, q := range path {
		dx, dy := float64(q[0]-p[0])/MandalaAspect, float64(q[1]-p[1])
		if d := float64(dx*dx) + float64(dy*dy); d < bd {
			best, bd = q, d
		}
	}
	return best
}

// mirrorAt is a first-quadrant offset in all four quadrants (rows up become
// rows down on screen), each image once.
func mirrorAt(p [2]int, part MandalaPart, ring, idx int) []MandalaCell {
	var out []MandalaCell
	for _, s := range [4][2]int{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}} {
		c := MandalaCell{DX: s[0] * p[0], DY: s[1] * p[1], Part: part, Ring: ring, Idx: idx}
		dup := false
		for _, o := range out {
			dup = dup || o.DX == c.DX && o.DY == c.DY
		}
		if !dup {
			out = append(out, c)
		}
	}
	return out
}

// mirrorIdx is the i-th of b points spaced round a ring (one in the first
// quadrant) and its mirror images, numbered clockwise from the top.
func mirrorIdx(p [2]int, part MandalaPart, ring, i, b int) []MandalaCell {
	out := mirrorAt(p, part, ring, i)
	for k := range out {
		c := &out[k]
		switch {
		case c.DX >= 0 && c.DY <= 0:
			c.Idx = i
		case c.DX >= 0:
			c.Idx = b/2 - i
		case c.DY > 0:
			c.Idx = b/2 + i
		default:
			c.Idx = (b - i) % b
		}
	}
	return out
}

// MandalaPulse is how many frames one wave of light takes to roll from the
// core out past the last ring: six seconds at the maps' eight frames a
// second.
const MandalaPulse = 48

// Breath is how brightly the wave of light lights ring k of n at frame anim,
// 0 to 1 (the core and the crown are ring -1): a crest about a ring wide
// rolling outward from the core, ring by ring, once every MandalaPulse
// frames, with nothing ahead of it or behind it. It starts and ends out of
// sight, so the wave never jumps.
func Breath(anim, k, n int) float64 {
	span := float64(max(0, n)) + 3.9
	f := float64(((anim % MandalaPulse) + MandalaPulse) % MandalaPulse)
	w := float64(span*f)/MandalaPulse - 2.7
	d := (float64(k) - w) / 1.2
	if float64(d*d) >= 1 {
		return 0
	}
	e := 1 - float64(d*d)
	return float64(e * e)
}

// PetalTip is the glyph of a petal's point at offset (dx, dy): a solid
// triangle pointing away from the core.
func PetalTip(dx, dy int) rune {
	switch {
	case dx == 0 && dy < 0:
		return '▲'
	case dx == 0:
		return '▼'
	case dy == 0 && dx > 0:
		return '►'
	case dy == 0:
		return '◄'
	case dx > 0 && dy < 0:
		return '◥'
	case dx > 0:
		return '◢'
	case dy > 0:
		return '◣'
	}
	return '◤'
}

// RayRune is the stroke of the core's ray at offset (dx, dy).
func RayRune(dx, dy int) rune {
	switch {
	case dx == 0:
		return '┃'
	case dy == 0:
		return '━'
	case dx*dy < 0:
		return '╱'
	}
	return '╲'
}

// EraLight is the colour of an era's ring before a style resolves it
// against its theme: the era's own hue, lifted a little toward the
// mandala's light so every era reads on the indigo.
func EraLight(epoch int) tcell.Color {
	lift := [7]float64{0.15, 0.25, 0.15, 0.1, 0.1, 0.15, 0.35} // the grey Iron and the violet Cosmic need more to stand off the indigo
	return theme.Mix(theme.MapHueColor(theme.EpochHue(epoch)), theme.SpaceColor(theme.SpaceLight), lift[max(0, min(epoch, 6))])
}

func clampF(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func absI(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
