package cabinet

import (
	"image"

	"github.com/espresso20/ageforge/pkg/coins"
)

// FanMax is the most copies Fan shows.
const FanMax = 9

// Fan paints the copies of well's stack slid out across the lining below the
// arc, finest nearest the well, each over its own hard shadow, along one
// smooth curve towards the side with more room. The picture covers whole
// cells of the lining, to lay over Open's picture; it is empty where the stack
// is empty or the panel leaves no room for copies.
func Fan(l TrayLayout, t Tray, well int, mint *coins.Mint) (cells image.Rectangle, img *image.RGBA) {
	if well < 0 || well >= Wells || len(t.Stacks[well]) == 0 || l.fanSz < 48 {
		return
	}
	st := t.Stacks[well][:min(len(t.Stacks[well]), FanMax)]
	sz, pts, box := l.fanSz, l.fanAt(well, len(st)), image.Rectangle{}
	for _, p := range pts {
		box = box.Union(image.Rect(p.X-sz/2, p.Y-sz/2, p.X-sz/2+sz+l.pad, p.Y-sz/2+sz+l.pad))
	}
	cells = image.Rect(box.Min.X/l.Cell.X, box.Min.Y/l.Cell.Y, (box.Max.X+l.Cell.X-1)/l.Cell.X, (box.Max.Y+l.Cell.Y-1)/l.Cell.Y).Intersect(l.fan)
	c := newCanvas(l.pix(cells), l.s)
	rp, kind := lining(t.Slot)
	c.surface(l.pix(cells), l.lining, rp, kind, 7)
	for k := len(st) - 1; k >= 0; k-- {
		if im := get(mint, t.Money, t.Types[well], st[k].Level, st[k].Seed, sz); im != nil {
			c.cast(im, pts[k].X, pts[k].Y, l.pad-1, 0)
		}
	}
	return cells, c.img
}

// fanAt is the centres of n copies fanned from well: the finest in the middle,
// centred on the well's x (kept a copy's width inside the lining), the rest
// alternating outwards. The row sags in the middle by a whole number of pixels
// a step squared, so its second differences keep one sign.
func (l TrayLayout) fanAt(well, n int) []image.Point {
	z, sz := l.pix(l.fan), l.fanSz
	reach := max(1, n/2) // the farthest copy, in steps
	step := max(1, min(sz*3/4, (z.Dx()/2-sz-l.pad)/max(1, reach)))
	span := reach * step
	cx := max(z.Min.X+span+sz, min(l.ctr[well].X, z.Max.X-span-sz-l.pad))
	slack := z.Dy() - 2*l.pad - sz
	q := 0
	if slack >= reach*reach {
		q = max(1, min(slack, sz/3)/(reach*reach))
	}
	pts := make([]image.Point, n)
	for k := range pts {
		j := (k + 1) / 2
		pts[k] = image.Pt(cx+(2*(k%2)-1)*j*step*btoi(k > 0), z.Min.Y+l.pad+sz/2+(slack-q*reach*reach)/2+q*(reach*reach-j*j))
	}
	return pts
}
