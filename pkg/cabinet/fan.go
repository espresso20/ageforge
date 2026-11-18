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

// fanAt is the centres of n copies fanned from well: a row from the well's
// own place towards the middle (to the right from the middle), sagging by a
// whole number of pixels a copy squared so the second differences keep one sign.
func (l TrayLayout) fanAt(well, n int) []image.Point {
	z, sz := l.pix(l.fan), l.fanSz
	half, mid := sz/2, (l.lining.Min.X+l.lining.Max.X)/2
	x0, dir := max(z.Min.X+l.pad+half, min(l.ctr[well].X, z.Max.X-l.pad-half)), 1
	room := z.Max.X - l.pad - half - x0
	if x0 > mid {
		dir, room = -1, x0-z.Min.X-l.pad-half
	}
	step := max(1, min(sz*3/4, room/max(1, n-1)))
	slack := z.Dy() - 2*l.pad - sz
	q := slack / 64
	pts := make([]image.Point, n)
	for k := range pts {
		pts[k] = image.Pt(x0+dir*k*step, z.Min.Y+l.pad+half+(slack-64*q)/2+q*k*k)
	}
	return pts
}
