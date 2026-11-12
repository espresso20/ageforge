package mapview

import (
	"math"

	"github.com/espresso20/ageforge/world"
)

// metres per one of a frame's own units.
var unitM = [...]float64{world.UnitKly: 9.4607e18, world.UnitLy: 9.4607e15, world.UnitAU: 1.496e11, world.UnitPlanetRadii: 6.371e6}

// A layer is one frame of the chain with what the camera needs of it.
type layer struct {
	f          *world.Frame
	mx         float64   // zoom metres per own unit (schematic scale included)
	anchor     world.Vec // the frame's origin, in the parent's units
	reach      float64   // how far from its origin the camera may centre, own units
	lo, hi     float64   // log10 of the view width at which the frame is fully in (lo) and out (hi)
	bmin, bmax float64   // log10 of the frame's own band
}

// A chain is the world's frames from the root down, each hung on the last.
type chain []layer

func (c chain) objPos(l int, key string) (world.Vec, bool) {
	for i := range c[l].f.Objects {
		if o := &c[l].f.Objects[i]; o.Key == key {
			return o.Pos, true
		}
	}
	return world.Vec{}, false
}

// newChain walks the first child at every level.
//
// The Local Group is schematic: it draws the Milky Way at 540 kly across
// where the galaxy frame is 100 kly, so every frame below it is scaled by that
// ratio and the zoom stays continuous. Sol is one point in every frame (the
// world puts it on the home arm and scales the marker from it), so no frame
// needs a special anchor.
func newChain(root *world.Frame) chain {
	var c chain
	k := 1.0
	for f := root; f != nil; {
		l := layer{f: f, mx: unitM[f.Unit] * k * math.Max(1, f.Inflate), reach: f.Radius * 1.1}
		for _, o := range f.Objects { // a satellite galaxy past the disc is still worth centring on
			if o.Kind == world.KindGalaxy {
				l.reach = math.Max(l.reach, math.Hypot(o.Pos.X, o.Pos.Y)+o.Size)
			}
		}
		if n := len(c); n > 0 {
			p := c[n-1]
			l.anchor = f.Pos
			if pos, ok := c.objPos(n-1, f.Anchor); ok && f.Anchor != "" {
				l.anchor = pos
			}
			if p.f.Schem && k == 1 { // the first frame under a schematic one sets the scale
				k = schemRatio(c, l.f)
				l.mx = unitM[f.Unit] * k * math.Max(1, f.Inflate)
			}
		}
		c = append(c, l)
		if len(f.Children) == 0 {
			break
		}
		f = f.Children[0]
	}
	c.bounds()
	return c
}

// bounds sets every layer's own band and the handoffs between them.
func (c chain) bounds() {
	for i := range c {
		c[i].bmin = math.Log10(c[i].f.Band.Min * c[i].mx)
		c[i].bmax = math.Log10(c[i].f.Band.Max * c[i].mx)
	}
	for i := 1; i < len(c); i++ { // the handoff spans whichever way the bands overlap or gap
		a, b := math.Min(c[i-1].bmin, c[i].bmax), math.Max(c[i-1].bmin, c[i].bmax)
		c[i].lo, c[i].hi = a, math.Max(b, a+1e-3)
	}
}

// withSystem is the chain with the system under the neighbourhood (and the
// close-up under that) swapped for sys, another star's, hung on its star.
func (c chain) withSystem(sys *world.Frame) chain {
	ni := c.idx(world.FrameNeighbourhood)
	out := append(chain(nil), c[:ni+1]...)
	for f := sys; f != nil; {
		out = append(out, out.layerFor(f, c[ni].mx/(unitM[c[ni].f.Unit]*math.Max(1, c[ni].f.Inflate))))
		if len(f.Children) == 0 {
			break
		}
		f = f.Children[0]
	}
	out.bounds()
	return out
}

// withTail is the chain with its last frame (a planet's close-up) swapped for
// f, another planet's, hung on the system above it.
func (c chain) withTail(f *world.Frame) chain {
	n := len(c) - 1
	sys := c[n-1]
	out := append(chain(nil), c[:n]...)
	out = append(out, out.layerFor(f, sys.mx/(unitM[sys.f.Unit]*math.Max(1, sys.f.Inflate)))) // k: the schematic scale, if any
	out.bounds()
	return out
}

// layerFor is the layer of frame f hung on the last frame of c, k the scale a
// schematic frame above sets.
func (c chain) layerFor(f *world.Frame, k float64) layer {
	l := layer{f: f, mx: unitM[f.Unit] * k * math.Max(1, f.Inflate), reach: f.Radius * 1.1, anchor: f.Pos}
	if pos, ok := c.objPos(len(c)-1, f.Anchor); ok && f.Anchor != "" {
		l.anchor = pos
	}
	return l
}

// idx is the position in the chain of the frame of kind k, or -1: the chain
// starts at the outermost frame its age shows, not always at the Local Group.
func (c chain) idx(k world.FrameKind) int {
	if i := int(k) - int(c[0].f.Kind); i >= 0 && i < len(c) {
		return i
	}
	return -1
}

// schemRatio is how much bigger the schematic frame draws the child than it is.
func schemRatio(c chain, child *world.Frame) float64 {
	p := c[len(c)-1]
	if o := p.anchorObj(child.Anchor); o != nil && child.Radius > 0 && o.Size > 0 {
		return o.Size * p.mx / (child.Radius * unitM[child.Unit])
	}
	return 1
}

func (l layer) anchorObj(key string) *world.Object {
	for i := range l.f.Objects {
		if l.f.Objects[i].Key == key {
			return &l.f.Objects[i]
		}
	}
	return nil
}

// A look is where the camera centres: master frame m, offset off. Frames below
// the Region hang on Sol, so they fade out when the centre is well away from
// it; the zero look is "at the frame's own origin" and leaves weights to zoom.
type look struct {
	m   int
	off world.Vec
	on  bool
}

// local is how far into view frame i is, by where the camera looks: 1 while
// the centre is within the frame's reach of its origin, fading to 0 a little
// past it, so a Region object can be zoomed into without the frames round Sol
// taking the view. Out past the arm's reach there is no Region object to look
// at, and the frames round Sol take over again.
func (c chain) local(i int, k look) float64 {
	if c[i].f.Kind < world.FrameNeighbourhood || !k.on {
		return 1
	}
	smooth := func(x float64) float64 { x = math.Max(0, math.Min(1, x)); return x * x * (3 - 2*x) }
	far := 0.0 // with no Region in the chain there is nothing else to look at
	if ri := c.idx(world.FrameRegion); ri >= 0 {
		r := c.centreIn(k.m, ri, k.off)
		far = smooth((math.Hypot(r.X, r.Y)/c[ri].reach - 1) / .5)
	}
	o := c.centreIn(k.m, i, k.off)
	return 1 - (1-far)*smooth((math.Hypot(o.X, o.Y)/c[i].reach-1)/.6)
}

// weight is how far into view frame i is at zoom z: 0 zoomed out past its
// handoff, 1 zoomed in past it. The root is always 1.
func (c chain) weight(i int, z float64) float64 { return c.weightAt(i, z, look{}) }

func (c chain) weightAt(i int, z float64, k look) float64 {
	if i == 0 {
		return 1
	}
	t := (c[i].hi - z) / (c[i].hi - c[i].lo)
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t) * c.local(i, k)
}

// layers is the two frames to draw at z: the deeper one that has begun to
// show, and the one under it while the deeper is still dithering in.
func (c chain) layers(z float64) (under, over int, w float64) { return c.layersAt(z, look{}) }

func (c chain) layersAt(z float64, k look) (under, over int, w float64) {
	over = 0
	for i := range c {
		if c.weightAt(i, z, k) > 0 {
			over = i
		}
	}
	w = c.weightAt(over, z, k)
	if w >= 1 || over == 0 {
		return over, over, 1
	}
	return over - 1, over, w
}

// master is the deepest frame at least half in view: the camera's centre is
// kept in its units, where it has the most precision.
func (c chain) master(z float64) int { return c.masterAt(z, look{}) }

func (c chain) masterAt(z float64, k look) int {
	m := 0
	for i := range c {
		if c.weightAt(i, z, k) >= 0.5 {
			m = i
		}
	}
	return m
}

// zRange is the zoom limits: all of the root, down to the deepest frame's floor.
func (c chain) zRange() (lo, hi float64) {
	return c[len(c)-1].bmin, c[0].bmax
}

// centreIn converts a centre given in frame m's units to frame i's units.
func (c chain) centreIn(m, i int, v world.Vec) world.Vec {
	for j := m; j > i; j-- { // out to the parent: scale, then offset by the anchor
		s := c[j].mx / c[j-1].mx
		v = world.Vec{X: c[j].anchor.X + v.X*s, Y: c[j].anchor.Y + v.Y*s}
	}
	for j := m; j < i; j++ { // in to the child
		s := c[j+1].mx / c[j].mx
		v = world.Vec{X: (v.X - c[j+1].anchor.X) / s, Y: (v.Y - c[j+1].anchor.Y) / s}
	}
	return v
}
