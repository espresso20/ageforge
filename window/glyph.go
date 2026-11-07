package window

import (
	"image"
	"math"
)

// glyph.go decides how one cell is painted. The block elements (U+2580 to
// U+259F) and box drawing (U+2500 to U+257F) are filled rectangles, so frames
// and the map meet cell to cell with no seams whatever the font; every other
// rune goes to a font.

// tint is a rectangle to fill and how much of the foreground it takes
// (1 is solid, less is a shade over the background).
type tint struct {
	r     image.Rectangle
	cover float64
}

// shape is what a procedural rune is made of: rectangles, plus the few strokes
// that are not (the rounded corners and the diagonals).
type shape struct {
	tints []tint
	arcs  []arc
	diags []diag
}

// arc is a quarter circle of a rounded corner: the stroke runs from (x0, y0) to
// (x1, y1) about the centre (cx, cy).
type arc struct{ cx, cy, rad, from, to, width float64 }

// diag is a diagonal stroke from (x0, y0) to (x1, y1).
type diag struct{ x0, y0, x1, y1, width float64 }

// isProcedural reports whether r is drawn by the window itself.
func isProcedural(r rune) bool { return r >= 0x2500 && r <= 0x259F }

// procedural returns the shape of r in a cw by ch cell, or false when r is not
// one of the runes the window draws itself.
func procedural(r rune, cw, ch int) (shape, bool) {
	switch {
	case r >= 0x2580 && r <= 0x259F:
		return blockShape(r, cw, ch), true
	case r >= 0x2500 && r <= 0x257F:
		return boxShape(r, cw, ch), true
	}
	return shape{}, false
}

func rect(x0, y0, x1, y1 int) image.Rectangle { return image.Rect(x0, y0, x1, y1) }

// blockShape draws U+2580 to U+259F. The eighth blocks round to whole pixels,
// so two cells that touch always meet on the same pixel row.
func blockShape(r rune, cw, ch int) shape {
	frac := func(n, total, eighths int) int { return n * eighths / 8 }
	one := func(rc image.Rectangle) shape { return shape{tints: []tint{{rc, 1}}} }
	switch {
	case r == 0x2580: // upper half
		return one(rect(0, 0, cw, ch/2))
	case r >= 0x2581 && r <= 0x2588: // lower 1/8 .. full
		n := int(r - 0x2580)
		return one(rect(0, ch-frac(ch, ch, n), cw, ch))
	case r >= 0x2589 && r <= 0x258F: // left 7/8 .. 1/8
		n := 8 - int(r-0x2588)
		return one(rect(0, 0, frac(cw, cw, n), ch))
	case r == 0x2590: // right half
		return one(rect(cw/2, 0, cw, ch))
	case r >= 0x2591 && r <= 0x2593: // light, medium, dark shade
		return shape{tints: []tint{{rect(0, 0, cw, ch), float64(r-0x2590) * 0.25}}}
	case r == 0x2594: // upper 1/8
		return one(rect(0, 0, cw, ch/8))
	case r == 0x2595: // right 1/8
		return one(rect(cw-cw/8, 0, cw, ch))
	}
	// 2596..259F quadrants: bits are upper-left, upper-right, lower-left, lower-right.
	quad := map[rune]uint8{
		0x2596: 0b0010, 0x2597: 0b0001, 0x2598: 0b1000, 0x2599: 0b1011, 0x259A: 0b1001,
		0x259B: 0b1110, 0x259C: 0b1101, 0x259D: 0b0100, 0x259E: 0b0110, 0x259F: 0b0111,
	}[r]
	mx, my := cw/2, ch/2
	var s shape
	if quad&0b1000 != 0 {
		s.tints = append(s.tints, tint{rect(0, 0, mx, my), 1})
	}
	if quad&0b0100 != 0 {
		s.tints = append(s.tints, tint{rect(mx, 0, cw, my), 1})
	}
	if quad&0b0010 != 0 {
		s.tints = append(s.tints, tint{rect(0, my, mx, ch), 1})
	}
	if quad&0b0001 != 0 {
		s.tints = append(s.tints, tint{rect(mx, my, cw, ch), 1})
	}
	return s
}

// A box rune is four arms (left, right, up, down), each with a weight.
const (
	wNone = iota
	wLight
	wHeavy
	wDouble
)

type arms [4]uint8 // left, right, up, down

const (
	aLeft = iota
	aRight
	aUp
	aDown
)

// boxArms is the arms of every straight box rune. Dashed lines map to their
// solid kind here and are cut up in boxShape.
var boxArms = func() map[rune]arms {
	m := map[rune]arms{}
	set := func(first rune, list ...arms) {
		for i, a := range list {
			m[first+rune(i)] = a
		}
	}
	// 2500 to 2503: the plain lines.
	set(0x2500, arms{1, 1, 0, 0}, arms{2, 2, 0, 0}, arms{0, 0, 1, 1}, arms{0, 0, 2, 2})
	// 250C to 254B: corners, tees and crosses, light and heavy.
	set(0x250C,
		arms{0, 1, 0, 1}, arms{0, 2, 0, 1}, arms{0, 1, 0, 2}, arms{0, 2, 0, 2},
		arms{1, 0, 0, 1}, arms{2, 0, 0, 1}, arms{1, 0, 0, 2}, arms{2, 0, 0, 2},
		arms{0, 1, 1, 0}, arms{0, 2, 1, 0}, arms{0, 1, 2, 0}, arms{0, 2, 2, 0},
		arms{1, 0, 1, 0}, arms{2, 0, 1, 0}, arms{1, 0, 2, 0}, arms{2, 0, 2, 0},
		arms{0, 1, 1, 1}, arms{0, 2, 1, 1}, arms{0, 1, 2, 1}, arms{0, 1, 1, 2},
		arms{0, 1, 2, 2}, arms{0, 2, 2, 1}, arms{0, 2, 1, 2}, arms{0, 2, 2, 2},
		arms{1, 0, 1, 1}, arms{2, 0, 1, 1}, arms{1, 0, 2, 1}, arms{1, 0, 1, 2},
		arms{1, 0, 2, 2}, arms{2, 0, 2, 1}, arms{2, 0, 1, 2}, arms{2, 0, 2, 2},
		arms{1, 1, 0, 1}, arms{2, 1, 0, 1}, arms{1, 2, 0, 1}, arms{2, 2, 0, 1},
		arms{1, 1, 0, 2}, arms{2, 1, 0, 2}, arms{1, 2, 0, 2}, arms{2, 2, 0, 2},
		arms{1, 1, 1, 0}, arms{2, 1, 1, 0}, arms{1, 2, 1, 0}, arms{2, 2, 1, 0},
		arms{1, 1, 2, 0}, arms{2, 1, 2, 0}, arms{1, 2, 2, 0}, arms{2, 2, 2, 0},
		arms{1, 1, 1, 1}, arms{2, 1, 1, 1}, arms{1, 2, 1, 1}, arms{2, 2, 1, 1},
		arms{1, 1, 2, 1}, arms{1, 1, 1, 2}, arms{1, 1, 2, 2}, arms{2, 1, 2, 1},
		arms{1, 2, 2, 1}, arms{2, 1, 1, 2}, arms{1, 2, 1, 2}, arms{2, 2, 2, 1},
		arms{2, 2, 1, 2}, arms{2, 1, 2, 2}, arms{1, 2, 2, 2}, arms{2, 2, 2, 2},
	)
	// 2550 to 256C: the double lines and their mixes.
	set(0x2550,
		arms{3, 3, 0, 0}, arms{0, 0, 3, 3},
		arms{0, 3, 0, 1}, arms{0, 1, 0, 3}, arms{0, 3, 0, 3},
		arms{3, 0, 0, 1}, arms{1, 0, 0, 3}, arms{3, 0, 0, 3},
		arms{0, 3, 1, 0}, arms{0, 1, 3, 0}, arms{0, 3, 3, 0},
		arms{3, 0, 1, 0}, arms{1, 0, 3, 0}, arms{3, 0, 3, 0},
		arms{0, 3, 1, 1}, arms{0, 1, 3, 3}, arms{0, 3, 3, 3},
		arms{3, 0, 1, 1}, arms{1, 0, 3, 3}, arms{3, 0, 3, 3},
		arms{3, 3, 0, 1}, arms{1, 1, 0, 3}, arms{3, 3, 0, 3},
		arms{3, 3, 1, 0}, arms{1, 1, 3, 0}, arms{3, 3, 3, 0},
		arms{3, 3, 1, 1}, arms{1, 1, 3, 3}, arms{3, 3, 3, 3},
	)
	// 256D to 2570: rounded corners are drawn as arcs (see boxShape), these
	// are the arms they join.
	set(0x256D, arms{0, 1, 0, 1}, arms{1, 0, 0, 1}, arms{1, 0, 1, 0}, arms{0, 1, 1, 0})
	// 2574 to 257F: half lines and the light/heavy mixes.
	set(0x2574,
		arms{1, 0, 0, 0}, arms{0, 0, 1, 0}, arms{0, 1, 0, 0}, arms{0, 0, 0, 1},
		arms{2, 0, 0, 0}, arms{0, 0, 2, 0}, arms{0, 2, 0, 0}, arms{0, 0, 0, 2},
		arms{1, 2, 0, 0}, arms{0, 0, 1, 2}, arms{2, 1, 0, 0}, arms{0, 0, 2, 1},
	)
	// Dashes: 2504 to 250B and 254C to 254F.
	for r := rune(0x2504); r <= 0x250B; r++ {
		heavy := uint8(1 + (r-0x2504)%2)
		if (r-0x2504)/2%2 == 0 {
			m[r] = arms{heavy, heavy, 0, 0}
		} else {
			m[r] = arms{0, 0, heavy, heavy}
		}
	}
	set(0x254C, arms{1, 1, 0, 0}, arms{2, 2, 0, 0}, arms{0, 0, 1, 1}, arms{0, 0, 2, 2})
	return m
}()

// dashCount is how many dashes a dashed rune has (0 for a solid one).
func dashCount(r rune) int {
	switch {
	case r >= 0x2504 && r <= 0x2507, r >= 0x2508 && r <= 0x250B:
		if r <= 0x2507 {
			return 3
		}
		return 4
	case r >= 0x254C && r <= 0x254F:
		return 2
	}
	return 0
}

// boxShape draws U+2500 to U+257F.
func boxShape(r rune, cw, ch int) shape {
	var s shape
	if r >= 0x2571 && r <= 0x2573 {
		w := math.Max(1, math.Round(float64(cw)/8))
		if r != 0x2571 { // the one that runs from upper left
			s.diags = append(s.diags, diag{0, 0, float64(cw), float64(ch), w})
		}
		if r != 0x2572 {
			s.diags = append(s.diags, diag{float64(cw), 0, 0, float64(ch), w})
		}
		return s
	}
	if r >= 0x256D && r <= 0x2570 {
		return roundedCorner(r, cw, ch)
	}
	a, ok := boxArms[r]
	if !ok {
		return s
	}
	tl := max(1, int(math.Round(float64(cw)/8)))
	th := 2 * tl
	d := tl // a double line's lines sit this far either side of the centre
	mx, my := cw/2, ch/2

	thick := func(w uint8) int {
		if w == wHeavy {
			return th
		}
		return tl
	}
	// reach is how far an arm of weight w reaches across the centre, to either side.
	reach := func(w uint8) int {
		switch w {
		case wDouble:
			return d + tl/2 + tl%2
		case wNone:
			return 0
		}
		return (thick(w) + 1) / 2
	}
	var add func(x0, y0, x1, y1 int)
	add = func(x0, y0, x1, y1 int) {
		if x1 > x0 && y1 > y0 {
			s.tints = append(s.tints, tint{rect(x0, y0, x1, y1), 1})
		}
	}
	// armLine adds one line of an arm: o is the offset across the arm, t its
	// thickness, e how far from the centre it starts (negative is past the centre).
	armLine := func(dir, o, t, e int) {
		switch dir {
		case aLeft:
			add(0, my+o-t/2, mx-e, my+o-t/2+t)
		case aRight:
			add(mx+e, my+o-t/2, cw, my+o-t/2+t)
		case aUp:
			add(mx+o-t/2, 0, mx+o-t/2+t, my-e)
		case aDown:
			add(mx+o-t/2, my+e, mx+o-t/2+t, ch)
		}
	}
	if n := dashCount(r); n > 0 {
		horizontal := a[aLeft] != wNone
		t := thick(max(a[aLeft], a[aUp]))
		line := rect(0, my-t/2, cw, my-t/2+t)
		if !horizontal {
			line = rect(mx-t/2, 0, mx-t/2+t, ch)
		}
		s.tints = dash([]tint{{line, 1}}, n, horizontal)
		return s
	}
	for dir := 0; dir < 4; dir++ {
		w := a[dir]
		if w == wNone {
			continue
		}
		opp := dir ^ 1
		var p [2]uint8 // the arms across this one: side -1 and side +1
		if dir < 2 {
			p = [2]uint8{a[aUp], a[aDown]}
		} else {
			p = [2]uint8{a[aLeft], a[aRight]}
		}
		through := a[opp] != wNone
		if w != wDouble {
			e := 0
			switch {
			case p[0] == wDouble || p[1] == wDouble:
				if !through {
					e = d
				}
			default:
				e = -max(reach(p[0]), reach(p[1]))
			}
			armLine(dir, 0, thick(w), e)
			continue
		}
		for si, side := range [2]int{-1, 1} {
			ps, po := p[si], p[1-si]
			e := 0
			switch {
			case through:
				if ps == wDouble {
					e = d
				}
			case ps == wNone && po == wNone:
				e = 0
			case ps == wDouble:
				e = d
			case po == wDouble:
				e = -(d + tl/2 + tl%2)
			default:
				e = -max(reach(ps), reach(po))
			}
			armLine(dir, side*d, tl, e)
		}
	}
	return s
}

// roundedCorner is a light corner with the bend as a quarter circle.
func roundedCorner(r rune, cw, ch int) shape {
	tl := max(1, int(math.Round(float64(cw)/8)))
	mx, my := cw/2, ch/2
	rad := cw / 2
	var s shape
	add := func(x0, y0, x1, y1 int) { s.tints = append(s.tints, tint{rect(x0, y0, x1, y1), 1}) }
	w := float64(tl)
	switch r {
	case 0x256D: // down and right
		add(mx-tl/2, my+rad, mx-tl/2+tl, ch)
		add(mx+rad, my-tl/2, cw, my-tl/2+tl)
		s.arcs = append(s.arcs, arc{float64(mx + rad), float64(my + rad), float64(rad), math.Pi, 1.5 * math.Pi, w})
	case 0x256E: // down and left
		add(mx-tl/2, my+rad, mx-tl/2+tl, ch)
		add(0, my-tl/2, mx-rad, my-tl/2+tl)
		s.arcs = append(s.arcs, arc{float64(mx - rad), float64(my + rad), float64(rad), 1.5 * math.Pi, 2 * math.Pi, w})
	case 0x256F: // up and left
		add(mx-tl/2, 0, mx-tl/2+tl, my-rad)
		add(0, my-tl/2, mx-rad, my-tl/2+tl)
		s.arcs = append(s.arcs, arc{float64(mx - rad), float64(my - rad), float64(rad), 0, 0.5 * math.Pi, w})
	case 0x2570: // up and right
		add(mx-tl/2, 0, mx-tl/2+tl, my-rad)
		add(mx+rad, my-tl/2, cw, my-tl/2+tl)
		s.arcs = append(s.arcs, arc{float64(mx + rad), float64(my - rad), float64(rad), 0.5 * math.Pi, math.Pi, w})
	}
	return s
}

// dash cuts the solid line in lines into n dashes with gaps between.
func dash(lines []tint, n int, horizontal bool) []tint {
	var out []tint
	for _, t := range lines {
		span := t.r.Dx()
		if !horizontal {
			span = t.r.Dy()
		}
		// n dashes and n gaps, the first gap half-width at each end.
		step := float64(span) / float64(n)
		for i := 0; i < n; i++ {
			a := int(math.Round(float64(i)*step + step*0.15))
			b := int(math.Round(float64(i)*step + step*0.85))
			if b <= a {
				b = a + 1
			}
			var rc image.Rectangle
			if horizontal {
				rc = rect(t.r.Min.X+a, t.r.Min.Y, t.r.Min.X+b, t.r.Max.Y)
			} else {
				rc = rect(t.r.Min.X, t.r.Min.Y+a, t.r.Max.X, t.r.Min.Y+b)
			}
			out = append(out, tint{rc, 1})
		}
	}
	return out
}
