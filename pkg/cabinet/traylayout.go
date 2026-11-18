package cabinet

import (
	"image"
	"math"
)

// Wells is the number of type wells in a tray; index Wells means the
// legendary well.
const Wells = 5

const (
	strip  = 3  // rows of each neighbour's strip
	rail   = 4  // rows of this drawer's own rail
	labelW = 12 // cells under a well's name plate
)

// TrayLayout is where the open drawer's parts sit. Wells and Legend are in
// pixels, every other rectangle is in cells.
type TrayLayout struct {
	Cells, Cell       image.Point
	Wells             [Wells]image.Rectangle
	Legend            image.Rectangle
	Labels            [Wells + 1]image.Rectangle // one row under each well; the last is the legendary's
	Name, Count, Hint image.Rectangle            // one row each, on this drawer's rail
	Above, Below      image.Rectangle            // the neighbours' name plates
	Ticket            image.Rectangle            // two rows on the rail, for the selected copy's words
	s                 float64                    // pixels per design unit
	lining            image.Rectangle
	ctr               [Wells + 1]image.Point
	piece, grow, step int             // type piece size, rim of the cut, offset between copies
	dia               int             // diameter of the legendary well
	fan               image.Rectangle // cells of the lining below the arc, where Fan lays copies
	fanSz, pad        int             // size of a fanned copy; room for its shadow
}

// PlanTray places the tray for a panel of cols by rows cells. The wells sit
// in a shallow arc under the legendary well; every label sits on whole cells.
// It reports false when the panel is too small to hold them apart.
func PlanTray(cols, rows, cw, ch int) (TrayLayout, bool) {
	if cols < 1 || rows < 1 || cw < 1 || ch < 1 {
		return TrayLayout{}, false
	}
	w := cols * cw
	s0 := float64(w) / 1920
	sw, wall := max(6, int(math.Round(36*s0))), max(3, int(math.Round(16*s0)))
	lin := image.Rect(sw, strip*ch+wall, w-sw, (rows-strip-rail)*ch-wall)
	if lin.Dy() < 1 || lin.Dx() < 1 {
		return TrayLayout{}, false
	}
	s := math.Min(float64(lin.Dx())/1848, float64(lin.Dy())/706)
	u := func(v float64) int { return int(math.Round(v * s)) }
	l := TrayLayout{Cells: image.Pt(cols, rows), Cell: image.Pt(cw, ch), s: s, lining: lin,
		piece: u(150), grow: max(3, u(9)), step: max(2, u(4)), dia: u(280)}
	if l.piece < 48 {
		return TrayLayout{}, false
	}
	ox, oy := (lin.Min.X+lin.Max.X)/2, lin.Min.Y-u(5)
	reach := l.piece/2 + l.grow + 4*l.step + 1
	for i, dx := range [Wells]float64{-640, -340, 0, 340, 640} {
		t := dx / 640
		c := image.Pt(ox+u(dx), oy+u(520-150*t*t))
		l.ctr[i], l.Wells[i] = c, image.Rect(c.X-reach, c.Y-reach, c.X+reach, c.Y+reach)
	}
	l.ctr[Wells] = image.Pt(ox, oy+u(190))
	r := l.dia/2 + max(2, u(3)) + 1
	l.Legend = image.Rect(ox-r, l.ctr[Wells].Y-r, ox+r, l.ctr[Wells].Y+r)
	my := max(2, u(3))
	for i := range l.Labels {
		b := l.Wells[min(i, Wells-1)].Max.Y
		if i == Wells {
			b = l.Legend.Max.Y
		}
		row, x := (b+my+ch)/ch, l.ctr[i].X/cw-labelW/2
		l.Labels[i] = image.Rect(x, row, x+labelW, row+1)
		p := l.pix(l.Labels[i])
		if p.Max.Y+my > lin.Max.Y {
			return TrayLayout{}, false
		}
		for _, o := range append(l.Wells[:], l.Legend) {
			if p.Overlaps(o) {
				return TrayLayout{}, false
			}
		}
	}
	low := 0
	for _, r := range l.Labels {
		low = max(low, l.pix(r).Max.Y)
	}
	l.pad = my + 1
	l.fan = image.Rect((lin.Min.X+cw-1)/cw, (low+2*my+1+ch-1)/ch, lin.Max.X/cw, lin.Max.Y/ch)
	l.fanSz = min(l.piece, (l.pix(l.fan).Dy()-2*l.pad)*3/4) // leaves room for the sag
	x0, pr := (sw+cw-1)/cw+2, rows-strip-rail/2-1
	l.Name = image.Rect(x0, pr, x0+ageGlyphs, pr+1)
	l.Count = image.Rect(x0+ageGlyphs+3, pr, x0+ageGlyphs+3+freeGlyph, pr+1)
	l.Hint = image.Rect(l.Count.Max.X+3, pr, l.Count.Max.X+39, pr+1)
	l.Above = image.Rect(x0, 1, x0+ageGlyphs, 2)
	l.Below = image.Rect(x0, rows-2, x0+ageGlyphs, rows-1)
	right := cols - x0
	l.Ticket = image.Rect(max(right-36, l.Hint.Max.X+3), pr, right, pr+2)
	return l, l.Ticket.Dx() >= 24
}

// At gives the well under a cell, Wells for the legendary well, or -1.
func (l TrayLayout) At(cellX, cellY int) int {
	p := image.Pt(cellX*l.Cell.X+l.Cell.X/2, cellY*l.Cell.Y+l.Cell.Y/2)
	if p.In(l.Legend) {
		return Wells
	}
	for i, w := range l.Wells {
		if p.In(w) {
			return i
		}
	}
	return -1
}

// pix turns a rectangle of cells into pixels.
func (l TrayLayout) pix(r image.Rectangle) image.Rectangle {
	return image.Rect(r.Min.X*l.Cell.X, r.Min.Y*l.Cell.Y, r.Max.X*l.Cell.X, r.Max.Y*l.Cell.Y)
}
