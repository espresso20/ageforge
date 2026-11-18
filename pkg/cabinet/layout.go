// Package cabinet paints the coin cabinet: a walnut case of 25 drawer fronts,
// each with its money as the pull. It draws pictures from plain data and
// leaves every word to the caller, which writes cell text on the flat plates.
package cabinet

import (
	"image"
	"image/color"
	"math"
)

// Cols, Rows and Slots give the grid of drawers.
const (
	Cols, Rows, Slots = 5, 5, 25
	ageGlyphs         = 18 // cells under the name plate of an age drawer
	freeGlyph         = 12 // cells under the name plate of an ageless drawer
	minFront          = 22 // narrowest front, in cells
	minPull           = 24 // smallest pull, in pixels
)

// Layout is where everything sits. Fronts are in pixels; the rest are in cells.
type Layout struct {
	Cells, Cell image.Point            // panel size in cells; one cell in pixels
	Fronts      [Slots]image.Rectangle // pixels, on whole cells
	Plates      [Slots]image.Rectangle // cells: one row, the drawer's name
	Title, Hint image.Rectangle        // cells: the totals plate on the cornice, the hint plate on the plinth
	s           float64                // pixels per design unit
	gap         image.Point            // cells between fronts
	cornice     int                    // rows of cornice
	base        int                    // first row of the plinth
	stile       [2]int                 // cells of the left and right sides of the case
}

// Plan picks the largest five by five grid that fits the panel, keeping the
// proportions of the case: whatever cells are left go to the sides and the
// plinth. It reports false when the panel is too small to draw.
func Plan(cols, rows, cw, ch int) (Layout, bool) {
	if cols < 1 || rows < 1 || cw < 1 || ch < 1 {
		return Layout{}, false
	}
	s0 := math.Min(float64(cols*cw)/1920, float64(rows*ch)/1056)
	cellsOf := func(v float64, size, lo int) int { return max(lo, int(math.Round(v*s0/float64(size)))) }
	up := func(v, size int) int { return (v + size - 1) / size }
	gx, gy := cellsOf(21, cw, up(6, cw)), cellsOf(21, ch, up(6, ch))
	lx := cellsOf(48, cw, 1)
	title := max(1, up(int(math.Round(13*s0)), ch)) // the plate's row
	cornice := title + 1 + max(1, up(int(math.Round(21*s0)), ch))
	fw := (cols - 2*lx - 4*gx) / Cols
	fh := (rows - cornice - (Rows+1)*gy - cellsOf(80, ch, 3)) / Rows
	if fw < minFront || fh < 4 {
		return Layout{}, false
	}
	l := Layout{Cells: image.Pt(cols, rows), Cell: image.Pt(cw, ch), gap: image.Pt(gx, gy), cornice: cornice}
	l.s = float64(fw*cw) / 348
	left := cols - 2*lx - 4*gx - Cols*fw
	l.stile = [2]int{lx + left/2, lx + left - left/2}
	for i := 0; i < Slots; i++ {
		x0 := l.stile[0] + (i%Cols)*(fw+gx)
		y0 := cornice + gy + (i/Cols)*(fh+gy)
		l.Fronts[i] = image.Rect(x0*cw, y0*ch, (x0+fw)*cw, (y0+fh)*ch)
		pw := ageGlyphs
		if i >= 22 {
			pw = freeGlyph
		}
		px := x0 + (fw-pw)/2
		l.Plates[i] = image.Rect(px, y0+fh-2, px+pw, y0+fh-1)
	}
	l.base = cornice + (Rows+1)*gy + Rows*fh
	tw, hw := min(40, cols-8), min(44, cols-8)
	l.Title = image.Rect((cols-tw)/2, title, (cols+tw)/2, title+1)
	l.Hint = image.Rect((cols-hw)/2, l.base+1, (cols+hw)/2, l.base+2)
	if _, _, d := l.pull(0); d < minPull {
		return Layout{}, false
	}
	return l, true
}

// At gives the drawer under a cell, or -1.
func (l Layout) At(cellX, cellY int) int {
	p := image.Pt(cellX*l.Cell.X, cellY*l.Cell.Y)
	for i, f := range l.Fronts {
		if p.In(f) {
			return i
		}
	}
	return -1
}

// pix turns a rectangle of cells into pixels.
func (l Layout) pix(r image.Rectangle) image.Rectangle {
	return image.Rect(r.Min.X*l.Cell.X, r.Min.Y*l.Cell.Y, r.Max.X*l.Cell.X, r.Max.Y*l.Cell.Y)
}

// pull gives the centre and diameter of drawer i's pull, in the room above
// its name plate.
func (l Layout) pull(i int) (cx, cy, dia int) {
	f := l.Fronts[i]
	pad := max(3, int(math.Round(6*l.s)))
	room := l.pix(l.Plates[i]).Min.Y - f.Min.Y - 2*pad
	dia = min(int(math.Round(98*l.s)), room)
	return (f.Min.X + f.Max.X) / 2, f.Min.Y + pad + room/2, dia
}

// PlateFace is the flat colour under plate words.
func PlateFace(dim bool) color.RGBA {
	if dim {
		return hx(0x584622)
	}
	return brass[2]
}

// PlateInk is the colour for words on a plate.
func PlateInk(dim bool) color.RGBA {
	if dim {
		return hx(0x2a2110)
	}
	return hx(0x16100a)
}
