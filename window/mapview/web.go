package mapview

import (
	"math"

	"github.com/espresso20/ageforge/world"
)

// web.go: the cosmic web the Transcendent Age opens. A strand is a bent
// filament drawn as short dithered segments (thicker the stronger its line
// glows); a cluster is a disc with a few members round it.

const strandSteps = 14

// strandPts is the strand's curve on screen, from its quadratic Path.
func (p *pen) strandPts(o *world.Object) (xs, ys [strandSteps + 1]float64) {
	ax, ay := p.at(o.Path[0])
	cx, cy := p.at(o.Path[1])
	bx, by := p.at(o.Path[2])
	for i := 0; i <= strandSteps; i++ {
		t := float64(i) / strandSteps
		u := 1 - t
		xs[i], ys[i] = u*u*ax+2*u*t*cx+t*t*bx, u*u*ay+2*u*t*cy+t*t*by
	}
	return xs, ys
}

func (p *pen) strand(o *world.Object) {
	if len(o.Path) < 3 {
		return
	}
	xs, ys := p.strandPts(o)
	lanes := 1 + int(o.Weight)/110 // a faint thread is one pixel wide, the strongest line three
	for i := 0; i < strandSteps; i++ {
		for k := 0; k < lanes; k++ {
			d := float64(k) - float64(lanes-1)/2
			p.line(xs[i], ys[i]+d, xs[i+1], ys[i+1]+d, o.Tone, o.Weight)
		}
	}
}

func (p *pen) webCluster(o *world.Object, x, y, r float64) {
	r = math.Max(1.5, r)
	p.disc(x, y, r, o.Tone, o.Weight)
	h := hashKey(o.Key) ^ uint32(o.Seed)
	for i := 0; i < 5; i++ { // the members round it
		h = h*1664525 + 1013904223
		a, d := float64(h>>8&1023)/1024*2*math.Pi, r*(1.4+float64(h>>20&7)/6)
		p.disc(x+d*math.Cos(a), y+d*math.Sin(a), 1, o.Tone, o.Weight/2)
	}
}

// strandDist is how far (screen pixels) x, y is from the strand's curve.
func strandDist(p pen, o *world.Object, x, y float64) float64 {
	if len(o.Path) < 3 {
		return math.Inf(1)
	}
	xs, ys := p.strandPts(o)
	best := math.Inf(1)
	for i := 0; i < strandSteps; i++ {
		ax, ay, bx, by := xs[i], ys[i], xs[i+1], ys[i+1]
		dx, dy := bx-ax, by-ay
		t := 0.0
		if l := dx*dx + dy*dy; l > 0 {
			t = math.Max(0, math.Min(1, ((x-ax)*dx+(y-ay)*dy)/l))
		}
		best = math.Min(best, math.Hypot(ax+t*dx-x, ay+t*dy-y))
	}
	return best
}
