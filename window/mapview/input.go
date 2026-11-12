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

// undisp inverts disp: the squeeze maps r to sqrt(r*R), so r = d*d/R.
func undisp(f *world.Frame, d world.Vec) world.Vec {
	if f.Kind != world.FrameRegion {
		return d
	}
	s := math.Hypot(d.X, d.Y) / regionSqueeze
	return world.Vec{X: d.X * s, Y: d.Y * s}
}

// snapped is the centre Render draws with: whole art pixels.
func (v *View) snapped(w int) world.Vec {
	ppu := float64(w) * v.ch[v.m].mx / v.width()
	return world.Vec{X: math.Round(v.off.X*ppu) / ppu, Y: math.Round(v.off.Y*ppu) / ppu}
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
	under, over, _ := v.ch.layers(v.z)
	off := v.snapped(w)
	best := float64(pickReach*pickReach) + 1e-9
	geom := &image.RGBA{Rect: image.Rect(0, 0, w, h)} // the pen only reads its size
	for _, i := range [...]int{under, over, v.m} {    // the master last, so a tie goes to it
		p := v.penFor(geom, i, off)
		for j := range v.ch[i].f.Objects {
			o := &v.ch[i].f.Objects[j]
			if !o.Pick {
				continue
			}
			ox, oy := p.at(o.Pos)
			if d := (ox-float64(x))*(ox-float64(x)) + (oy-float64(y))*(oy-float64(y)); d <= best {
				best, sel, ok = d, selection{fi: i, key: o.Key}, true
			}
		}
	}
	return sel, ok
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
	if o == nil {
		return
	}
	v.hold = hold{}
	v.focus = focusTo{on: true, fi: v.sel.fi, pos: o.Pos}
	v.tz -= focusSteps * zoomStep
	for i := v.sel.fi + 1; i < len(v.ch); i++ {
		if v.ch[i].f.Anchor == o.Key {
			v.tz = v.ch[i].lo
			break
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
