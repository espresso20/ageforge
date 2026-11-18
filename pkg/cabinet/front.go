package cabinet

import (
	"image"
	"sync"

	"github.com/espresso20/ageforge/pkg/coins"
)

// Drawer is what one drawer front shows.
type Drawer struct {
	Money       string // age key of the pull's money; "" for an ageless drawer (ring pull)
	Reached     bool   // false: recess with the outline of the finest piece, plate dimmed
	Best        int    // finest level held, 1..coins.Levels; 0 for none
	Legend      bool   // the legendary is held: it is the pull
	Total, Held int    // studs, and how many are lit
}

func drawerAt(ds []Drawer, i int) Drawer {
	if i < len(ds) {
		return ds[i]
	}
	return Drawer{}
}

// pull paints the money as the drawer's pull. Where the mint draws nothing
// the empty recess or well stays.
func (c *canvas) pull(cx, cy, dia int, d Drawer, mint *coins.Mint) {
	n := dia * 88 / 98
	piece := func(level, size int) *image.RGBA {
		if mint == nil {
			return nil
		}
		return mint.Piece(d.Money, "age", level, 0, size)
	}
	switch {
	case d.Money == "":
		c.ring(cx, cy, dia, cx+cy)
	case !d.Reached || (d.Best < 1 && !d.Legend):
		c.disc(cx, cy, dia, false)
		if im := piece(coins.Levels, n); im != nil {
			c.outline(cx, cy, im)
		}
	case d.Legend:
		c.disc(cx, cy, dia, true)
		if im := piece(coins.Legendary, n+2); im != nil {
			c.lay(im, cx, cy, 0, false)
		}
	default:
		im := piece(min(d.Best, coins.Levels), n)
		if im == nil {
			c.disc(cx, cy, dia, false)
			return
		}
		if covered(im) < 0.4 { // a thin piece needs the recess to read against the wood
			c.disc(cx, cy, dia, false)
		}
		c.lay(im, cx, cy, max(1, c.sc(3)), true)
		c.lay(im, cx, cy, 0, false)
	}
}

// drawer paints drawer i's front moved by off. The plate stays on its cells.
func (c *canvas) drawer(l Layout, d Drawer, i int, off image.Point, mint *coins.Mint) {
	f := l.Fronts[i].Add(off)
	seed := i*7919 + 13
	lift := 0.0
	if off != (image.Point{}) {
		lift = 0.14 // a pulled front catches more light
	}
	c.board(f, 100+i*37, []float64{0, 0.16, 0, -0.14, 0, 0.1, -0.1}[i%7]+lift, false)
	c.edge(f, false)
	c.edge(f.Inset(1), false)
	c.edge(f.Inset(max(3, c.sc(5))), true)
	cx, cy, dia := l.pull(i)
	c.pull(cx+off.X, cy+off.Y, dia, d, mint)
	pl := l.pix(l.Plates[i])
	c.plate(pl, d.Money != "" && !d.Reached, seed)
	if d.Money == "" {
		return
	}
	lo := pl.Max.Y + max(2, c.sc(3)) + max(1, c.sc(3))
	hi := f.Max.Y - max(3, c.sc(5)) - off.Y
	c.studs(cx, (lo+hi)/2, d.Total, d.Held, f.Dx()-2*c.sc(24), seed)
	if d.Reached && d.Total > 0 && d.Held == d.Total {
		c.seal(f.Max.X-c.sc(26), f.Min.Y+c.sc(26), seed)
	}
}

// Front paints the whole cabinet from the front, cols*cellW by rows*cellH
// pixels. The drawers are cut in parallel; each writes only its own front.
func Front(l Layout, ds []Drawer, mint *coins.Mint) *image.RGBA {
	c := newCanvas(image.Rectangle{Max: image.Pt(l.Cells.X*l.Cell.X, l.Cells.Y*l.Cell.Y)}, l.s)
	c.carcass(l)
	var wg sync.WaitGroup
	for i := 0; i < Slots; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.drawer(l, drawerAt(ds, i), i, image.Point{}, mint)
		}()
	}
	wg.Wait()
	c.cornice(l)
	c.plinth(l)
	return c.img
}

// Patch is drawer i pulled out over its shadow, with the carcass behind it: a
// picture covering whole cells, to lay over Front's picture at the same place.
func Patch(l Layout, ds []Drawer, i int, mint *coins.Mint) (cells image.Rectangle, img *image.RGBA) {
	f := l.Fronts[i]
	cells = image.Rect(f.Min.X/l.Cell.X-1, f.Min.Y/l.Cell.Y-1, f.Max.X/l.Cell.X+1, f.Max.Y/l.Cell.Y+1)
	c := newCanvas(l.pix(cells), l.s)
	c.carcass(l)
	room := min(l.Cell.X, l.Cell.Y) - 1
	po, sh := min(max(4, c.sc(10)), room), min(max(3, c.sc(7)), room)
	c.fill(f.Add(image.Pt(sh, sh)), black)
	c.sides(f.Add(image.Pt(-po, -po)), po)
	c.drawer(l, drawerAt(ds, i), i, image.Pt(-po, -po), mint)
	return cells, c.img
}

// sides paints the boards of a pulled drawer in perspective, below and to the
// right of its front, each line a step darker as it recedes.
func (c *canvas) sides(r image.Rectangle, n int) {
	for k := 1; k <= n; k++ {
		c.fill(image.Rect(r.Min.X+k, r.Max.Y-1+k, r.Max.X+k, r.Max.Y+k), walnut[4-k*3/(n+1)])
		c.fill(image.Rect(r.Max.X-1+k, r.Min.Y+k, r.Max.X+k, r.Max.Y-1+k), walnut[1-k/(n+1)])
	}
}
