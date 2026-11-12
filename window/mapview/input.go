package mapview

import (
	"image"
	"math"

	"github.com/espresso20/ageforge/world"
)

// The mouse. Positions are art pixels on a canvas w by h, the same canvas
// Render fills, so what a click lands on is what was drawn.

const (
	pickReach  = 6   // art pixels a click may miss an object by
	focusSteps = 2.0 // zoom steps a double click goes in on an object with no frame of its own
)

// A selection is an object of a chain frame, by key so a rebuilt sky keeps it.
type selection struct {
	fi  int
	key string
}

// State is what the ui needs of the map: the frame it is centred in and the
// object selected, if any. Sel points into the world and is read-only.
type State struct {
	Frame    world.FrameKind
	SelFrame world.FrameKind
	Sel      *world.Object
	Lock     string // the age that opens Sel's frame, while the age has not reached it
}

// hold keeps the point under the cursor still while a wheel zoom eases: P is
// that point, in frame hm's units, and f where the cursor is, in widths from
// the canvas centre.
type hold struct {
	on     bool
	p      world.Vec
	hm     int
	fx, fy float64
}

// focusTo eases the centre onto an object after a double click.
type focusTo struct {
	on  bool
	fi  int
	pos world.Vec
}

// disp is a point's display position in frame f: the region squeezes distance.
func disp(f *world.Frame, p world.Vec) world.Vec { return regionPos(f, p) }

// undisp inverts disp.
func undisp(f *world.Frame, d world.Vec) world.Vec {
	r := math.Hypot(d.X, d.Y)
	if r == 0 || !squeezed(f) {
		return d
	}
	s := unsqueeze(f, r) / r
	return world.Vec{X: d.X * s, Y: d.Y * s}
}

// snapped is the centre Render draws with: whole art pixels.
func (v *View) snapped(w int) world.Vec {
	ppu := float64(w) * v.ch[v.m].mx / v.width()
	f := v.ch[v.m].f
	d := disp(f, v.off)
	return undisp(f, world.Vec{X: math.Round(d.X*ppu) / ppu, Y: math.Round(d.Y*ppu) / ppu})
}

// objAt finds the object of chain frame fi called key.
func (v *View) objAt(fi int, key string) *world.Object {
	if fi < 0 || fi >= len(v.ch) || key == "" {
		return nil
	}
	for i := range v.ch[fi].f.Objects {
		if o := &v.ch[fi].f.Objects[i]; o.Key == key {
			return o
		}
	}
	return nil
}

// state is the State now.
func (v *View) state() State {
	s := State{Frame: v.ch[v.m].f.Kind}
	if o := v.objAt(v.sel.fi, v.sel.key); o != nil {
		s.Sel, s.SelFrame = o, v.ch[v.sel.fi].f.Kind
		if v.sel.fi < v.open {
			s.Lock = world.OpensAt(s.SelFrame)
		}
	}
	return s
}

type stateKey struct {
	fi    int
	key   string
	frame world.FrameKind
	valid bool
}

// take reports the state when it changed since the last report.
func (v *View) take() (State, func(State), bool) {
	if v.ch == nil || v.onChange == nil {
		return State{}, nil, false
	}
	s := v.state()
	k := stateKey{frame: s.Frame, valid: true}
	if s.Sel != nil {
		k.fi, k.key = v.sel.fi, s.Sel.Key
	}
	if k == v.last {
		return State{}, nil, false
	}
	v.last = k
	return s, v.onChange, true
}

// flush hands a changed State to the callback, outside the lock.
func (v *View) flush() {
	v.mu.Lock()
	s, fn, ok := v.take()
	v.mu.Unlock()
	if ok {
		fn(s)
	}
}

// OnChange sets the callback for a change of master frame or selection. It
// runs on whichever goroutine caused the change and must not block.
func (v *View) OnChange(fn func(State)) {
	v.mu.Lock()
	v.onChange, v.last = fn, stateKey{}
	v.mu.Unlock()
}

// SetNames sets how a label key reads: the ui owns the words.
func (v *View) SetNames(names func(string) string) {
	v.mu.Lock()
	v.names, v.dirty = names, true
	v.mu.Unlock()
}

// SetFace sets the pixel font labels are drawn in; nil draws none.
func (v *View) SetFace(f Face) {
	v.mu.Lock()
	v.face, v.dirty = f, true
	v.mu.Unlock()
}

// cursorPoint is the master-frame point under the cursor at x, y, and where
// the cursor is in widths from the canvas centre.
func (v *View) cursorPoint(x, y, w, h int) (p world.Vec, fx, fy float64) {
	fx, fy = (float64(x)-float64(w)/2)/float64(w), (float64(y)-float64(h)/2)/float64(w)
	f, s := v.ch[v.m].f, v.width()/v.ch[v.m].mx
	c := disp(f, v.off)
	return undisp(f, world.Vec{X: c.X + fx*s, Y: c.Y + fy*s}), fx, fy
}

// Wheel zooms by n steps (positive in) about the cursor at x, y: the point
// under it stays under it while the zoom eases.
func (v *View) Wheel(x, y, w, h int, n float64) {
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil || w <= 0 || h <= 0 || n == 0 {
		return
	}
	v.focus = focusTo{}
	p, fx, fy := v.cursorPoint(x, y, w, h)
	switch v.m { // the star or planet under the cursor is the one the zoom is going into
	case v.ch.idx(world.FrameNeighbourhood):
		v.nearSystem(p)
	case v.ch.idx(world.FrameSystem):
		v.nearTail(p)
	}
	v.hold = hold{on: true, p: p, hm: v.m, fx: fx, fy: fy}
	v.tz -= n * zoomStep
	v.clamp()
	v.dirty = true
}

// Pan moves the view by dx, dy art pixels on a canvas w wide: the sky follows
// the pointer. The picture snaps to whole art pixels when drawn.
func (v *View) Pan(dx, dy float64, w int) {
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil || w <= 0 {
		return
	}
	v.hold, v.focus = hold{}, focusTo{}
	f, s := v.ch[v.m].f, v.width()/v.ch[v.m].mx/float64(w) // frame units per art pixel
	c := disp(f, v.off)
	v.off = undisp(f, world.Vec{X: c.X - dx*s, Y: c.Y - dy*s})
	v.clamp()
	v.dirty = true
}

// near is the nearest pickable object to x, y within reach: from the frames
// on screen, the master winning a tie.
func (v *View) near(x, y, w, h int) (sel selection, ok bool) {
	under, over, _ := v.ch.layersAt(v.z, v.look())
	off := v.snapped(w)
	best, nearC := float64(pickReach*pickReach)+1e-9, 0.0
	geom := &image.RGBA{Rect: image.Rect(0, 0, w, h)} // the pen only reads its size
	for _, i := range [...]int{under, over, v.m} {    // the master last, so a tie goes to it
		p := v.penFor(geom, i, off)
		for j := range v.ch[i].f.Objects {
			o := &v.ch[i].f.Objects[j]
			if !o.Pick {
				continue
			}
			ox, oy := p.at(o.Pos)
			c := math.Hypot(ox-float64(x), oy-float64(y))
			d := math.Max(0, c-pickEdge(p, o))
			if d*d < best || (d*d == best && ok && c <= nearC) { // on two discs, the one whose centre is closer
				best, nearC, sel, ok = d*d, c, selection{fi: i, key: o.Key}, true
			}
		}
	}
	return sel, ok
}

// pickEdge is how far from its centre a click still hits o's sprite itself: a
// big sphere is hit anywhere on its disc, not only near its middle.
func pickEdge(p pen, o *world.Object) float64 {
	if p.f.Kind == world.FramePlanet || (p.f.Kind == world.FrameSystem && o.Kind == world.KindPlanet && o.Body != nil) {
		return drawnR(p, o) - 1
	}
	return 0
}

// Click selects the nearest object within reach of x, y, or nothing.
func (v *View) Click(x, y, w, h int) {
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil || w <= 0 || h <= 0 {
		return
	}
	v.sel, _ = v.near(x, y, w, h)
	v.pick = -1
	if v.sel.fi == v.m {
		for j := range v.ch[v.m].f.Objects {
			if v.ch[v.m].f.Objects[j].Key == v.sel.key {
				v.pick = j
			}
		}
	}
	v.dirty = true
}

// Focus is a double click: select, then ease the centre onto the object and
// zoom to the band of the frame that hangs on it, or a step in when none does.
func (v *View) Focus(x, y, w, h int) {
	v.Click(x, y, w, h)
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	o := v.objAt(v.sel.fi, v.sel.key)
	if o == nil || v.sel.fi < v.open { // a locked frame's objects can be selected, not entered
		return
	}
	key := o.Key
	switch {
	case o.Kind == world.KindPlanet && v.sel.fi == v.ch.idx(world.FrameSystem):
		v.swapTail(key)
	case o.Kind == world.KindStar && v.sel.fi == v.ch.idx(world.FrameGalaxy):
		if !o.Held && key != "sol" { // seen from outside the territory: named, not entered
			return
		}
		v.swapPlace(key)
	case o.Kind == world.KindStar && v.sel.fi == v.ch.idx(world.FrameRegion):
		v.swapNear(key)
	case (o.Kind == world.KindStar || o.Kind == world.KindColony) && v.sel.fi == v.ch.idx(world.FrameNeighbourhood):
		key = o.Label // a colony stands at its star (its label is the star's key): the dive goes to that system
		v.swapSystem(key)
	}
	v.hold = hold{}
	v.focus = focusTo{on: true, fi: v.sel.fi, pos: o.Pos}
	v.tz -= focusSteps * zoomStep
	for i := v.sel.fi + 1; i < len(v.ch); i++ {
		if v.ch[i].f.Anchor == key {
			v.tz = v.ch[i].lo
			break
		}
	}
	if o.Kind == world.KindGalaxy { // a galaxy the deeper frame draws close up: dive to it, five of its widths across
		for i := v.sel.fi + 1; i < len(v.ch); i++ {
			for j := range v.ch[i].f.Objects {
				if c := &v.ch[i].f.Objects[j]; c.Key == o.Key && c.Kind == world.KindGalaxy {
					v.sel, v.pick = selection{fi: i, key: c.Key}, -1
					v.focus = focusTo{on: true, fi: i, pos: c.Pos}
					v.tz = math.Log10(10 * c.Size * v.ch[i].mx)
				}
			}
		}
	}
	if v.tz > v.z {
		v.tz = v.z - focusSteps*zoomStep
	}
	v.clamp()
	v.dirty = true
}

// follow keeps what the zoom and a focus are holding still, after the zoom
// moved and the master may have changed. k is the share of the way to go.
func (v *View) follow(k float64) {
	if v.hold.on {
		if v.hold.hm != v.m {
			v.hold.p, v.hold.hm = v.ch.centreIn(v.hold.hm, v.m, v.hold.p), v.m
		}
		f, s := v.ch[v.m].f, v.width()/v.ch[v.m].mx
		d := disp(f, v.hold.p)
		v.off = undisp(f, world.Vec{X: d.X - v.hold.fx*s, Y: d.Y - v.hold.fy*s})
		if v.z == v.tz {
			v.hold.on = false
		}
	}
	if v.focus.on {
		t := v.ch.centreIn(v.focus.fi, v.m, v.focus.pos)
		d := world.Vec{X: t.X - v.off.X, Y: t.Y - v.off.Y}
		if s := v.width() / v.ch[v.m].mx; math.Hypot(d.X, d.Y) < 1e-3*s {
			v.off, v.focus.on = t, false
		} else {
			v.off = world.Vec{X: v.off.X + d.X*k, Y: v.off.Y + d.Y*k}
		}
	}
}

// swapSystem hangs the system of the star key under the neighbourhood, in
// place of the one there, when that system is out of view (so nothing pops).
// The star sits where the neighbourhood draws it, so the crossfade holds still.
func (v *View) swapSystem(key string) {
	si := v.ch.idx(world.FrameSystem)
	if si < 1 || v.ch[si].f.Anchor == key || v.ch.weightAt(si, v.z, v.look()) > 0 {
		return
	}
	if f := v.w.System(key); f != nil {
		v.ch, v.star, v.dirty = v.ch.withFrom(world.FrameNeighbourhood, f), key, true
		v.warmClose(f)
	}
}

// swapPlace hangs the region of star key (and the chain under it) under the
// galaxy, in place of the one there, when that region is out of view. Sol's is
// the home one the world built.
func (v *View) swapPlace(key string) {
	ri := v.ch.idx(world.FrameRegion)
	if ri < 1 || v.ch[ri].f.Anchor == key || v.ch.weightAt(ri, v.z, v.look()) > 0 {
		return
	}
	if f := v.w.Place(key); f != nil {
		v.ch, v.dirty = v.ch.withFrom(world.FrameGalaxy, f), true
		v.place, v.nbr, v.star = key, "", ""
		if key == "sol" {
			v.place = ""
		}
	}
}

// swapNear hangs the neighbourhood of star key (and the system under it) under
// the region, in place of the one there, when that is out of view.
func (v *View) swapNear(key string) {
	ni := v.ch.idx(world.FrameNeighbourhood)
	if ni < 1 || v.ch[ni].f.Anchor == key || v.ch.weightAt(ni, v.z, v.look()) > 0 {
		return
	}
	if f := v.w.Near(key); f != nil {
		v.ch, v.dirty = v.ch.withFrom(world.FrameRegion, f), true
		v.nbr, v.star = key, ""
		if key == "sol" {
			v.nbr = ""
		}
	}
}

// nearSystem swaps in the system of the star nearest p, a point in the
// neighbourhood's units, if one is close enough to be what the zoom is going into.
func (v *View) nearSystem(p world.Vec) {
	ni := v.ch.idx(world.FrameNeighbourhood)
	if ni < 0 {
		return
	}
	f, best, key := v.ch[ni].f, v.width()/v.ch[ni].mx*.12, ""
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindStar {
			if h := math.Hypot(o.Pos.X-p.X, o.Pos.Y-p.Y); h < best {
				best, key = h, o.Key
			}
		}
	}
	if key != "" {
		v.swapSystem(key)
	}
}

// warmClose makes the sprites of the first planet's close-up of system f off
// the draw path.
func (v *View) warmClose(f *world.Frame) {
	if len(f.Children) > 0 {
		v.bod.warm(v.pal, f.Children[0])
	}
}

// swapTail hangs the close-up of planet key under the system, in place of the
// last frame, when that frame is out of view (so nothing pops). The planet
// sits where the system draws it, so the crossfade into it holds still.
func (v *View) swapTail(key string) {
	n := len(v.ch)
	if n < 2 || v.ch[n-1].f.Kind != world.FramePlanet || v.ch[n-1].f.Anchor == key || v.ch.weightAt(n-1, v.z, v.look()) > 0 {
		return
	}
	for _, f := range v.ch[n-2].f.Children {
		if f.Anchor == key {
			v.ch, v.dirty = v.ch.withTail(f), true
			v.bod.warm(v.pal, f)
			return
		}
	}
}

// nearTail swaps in the close-up of the planet nearest p, a point in the
// system's units, as the squeezed system draws distance.
func (v *View) nearTail(p world.Vec) {
	f := v.ch[v.ch.idx(world.FrameSystem)].f
	pd, best, key := disp(f, p), math.Inf(1), ""
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindPlanet && o.Body != nil {
			d := disp(f, o.Pos)
			if h := math.Hypot(d.X-pd.X, d.Y-pd.Y); h < best {
				best, key = h, o.Key
			}
		}
	}
	if key != "" {
		v.swapTail(key)
	}
}
