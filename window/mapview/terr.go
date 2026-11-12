package mapview

import (
	"math"

	"github.com/espresso20/ageforge/world"
)

// territory tints the sectors the people hold over the disc: a soft dither of
// the player's gold, and a brighter, denser edge where the territory ends. The
// raster it reads is the plan's own, built once per plan.
func (p *pen) territory(o *world.Object, x, y, rpx float64) {
	if rpx < 4 {
		return
	}
	mask := o.Terr.Mask()
	n := int(math.Sqrt(float64(len(mask))))
	gm := geomOf(o)
	hx, hy := rpx*math.Hypot(gm.cos, gm.sin*gm.q), rpx*gm.ys*math.Hypot(gm.sin, gm.cos*gm.q)
	xi, yi := int(math.Round(x)), int(math.Round(y))
	gold := p.pal.Mix(world.ToneGold, 255)
	for yy := max(0, yi-int(hy)-1); yy <= min(p.h-1, yi+int(hy)+1); yy++ {
		for xx := max(0, xi-int(hx)-1); xx <= min(p.w-1, xi+int(hx)+1); xx++ {
			wx, wy := xx-xi, yy-yi
			u, v := gm.face(float64(wx), float64(wy), rpx)
			if u*u+v*v >= 1 {
				continue
			}
			c := mask[int((v+1)/2*float64(n))*n+int((u+1)/2*float64(n))]
			if c == 0 {
				continue
			}
			b := Bayer(wx, wy)
			k := 0
			switch {
			case c == 2 && b < .7:
				k = 200
			case c == 1 && b < .4:
				k = 95
			}
			if k > 0 {
				p.px(xx, yy, mixRGB(p.img.RGBAAt(xx, yy), gold, k))
			}
		}
	}
}
