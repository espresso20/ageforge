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
	return l.fan9(t, well, -1, mint)
}

// FanLift paints copy k of the fan raised over a hard shadow with a brass lip:
// the copy's cells and what lies about them, to lay over Fan's picture.
func FanLift(l TrayLayout, t Tray, well, k int, mint *coins.Mint) (cells image.Rectangle, img *image.RGBA) {
	return l.fan9(t, well, k, mint)
}

// FanOrder is the copies' indices from the leftmost on screen to the
// rightmost, for a fan of n copies.
func FanOrder(n int) []int {
	var o []int
	for k := (n - 1) &^ 1; k > 0; k -= 2 {
		o = append(o, k)
	}
	if n > 0 {
		o = append(o, 0)
	}
	for k := 1; k < n; k += 2 {
		o = append(o, k)
	}
	return o
}

// fan9 paints the fan; sel >= 0 raises that copy and covers only its cells.
func (l TrayLayout) fan9(t Tray, well, sel int, mint *coins.Mint) (cells image.Rectangle, img *image.RGBA) {
	if well < 0 || well >= Wells || len(t.Stacks[well]) == 0 || l.fanSz < 48 {
		return
	}
	st := t.Stacks[well][:min(len(t.Stacks[well]), FanMax)]
	sz, pts, box := l.fanSz, l.fanAt(well, len(st)), image.Rectangle{}
	lift := max(4, int(8*l.s+0.5))
	for k, p := range pts {
		if sel < 0 || k == sel {
			box = box.Union(image.Rect(p.X-sz/2-lift-2, p.Y-sz/2-lift-2, p.X-sz/2+sz+l.pad+2, p.Y-sz/2+sz+l.pad+2))
		}
	}
	cells = image.Rect(box.Min.X/l.Cell.X, box.Min.Y/l.Cell.Y, (box.Max.X+l.Cell.X-1)/l.Cell.X, (box.Max.Y+l.Cell.Y-1)/l.Cell.Y).Intersect(l.fan)
	c := newCanvas(l.pix(cells), l.s)
	rp, kind := lining(t.Slot)
	c.surface(l.pix(cells), l.lining, rp, kind, 7)
	for k := len(st) - 1; k >= 0; k-- {
		im := get(mint, t.Money, t.Types[well], st[k].Level, st[k].Seed, sz)
		switch {
		case im == nil:
		case k != sel:
			c.cast(im, pts[k].X, pts[k].Y, l.pad-1, 0)
		default:
			x, y := pts[k].X-lift, pts[k].Y-lift
			c.cast(im, x, y, l.pad-1+lift, lift)
			c.tint = brass[4]
			for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				c.lay(im, x+d[0], y+d[1], 0, true)
			}
			c.lay(im, x, y, 0, false)
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
