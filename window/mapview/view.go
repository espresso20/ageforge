package mapview

import (
	"image"
	"math"
	"sync"
	"time"

	"github.com/espresso20/ageforge/world"
)

// Op is what a key asks the camera to do.
type Op uint8

const (
	PanLeft Op = iota
	PanRight
	PanUp
	PanDown
	ZoomIn
	ZoomOut
	Next // the next pickable object in the frame, centred
	Prev
	Home // back to the root's home object, fully zoomed out
)

// Cmd is one key for the camera. Fast (Shift) pans four times as far.
type Cmd struct {
	Op   Op
	Fast bool
}

const (
	zoomStep = 0.5                   // decades per key
	panStep  = 0.125                 // of the view width per key
	easeTau  = 50 * time.Millisecond // 150 ms to settle to within 5%
)

// ArtScale is the integer number of device pixels one art pixel covers on a
// map w device pixels wide: about 3 at 1200 wide, never fewer than 2.
func ArtScale(w int) int { return max(2, (w+175)/350) }

// View is the map's camera and canvas. Its methods are safe from the ui
// goroutine (Set, Clear, Do) and the draw thread (Update, Render) at once.
type View struct {
	mu       sync.Mutex
	w        *world.World
	key      uint64
	pal      world.Palette
	ch       chain
	z, tz    float64   // log10 of the view width in metres, now and wanted
	m        int       // master frame
	off      world.Vec // centre, in the master frame's units
	pick     int       // the object Next/Prev last centred
	sel      selection // the object selected, by click or Next/Prev
	hold     hold
	focus    focusTo
	face     Face
	names    func(string) string
	onChange func(State)
	last     stateKey
	dirty    bool
	canvas   *image.RGBA
	tex      texCache
}

func New() *View { return &View{dirty: true} }

// Set gives the view the sky. Nothing is rebuilt unless key changed.
func (v *View) Set(w *world.World, key uint64, pal world.Palette) {
	if w != nil && w.Root != nil {
		v.tex.warm(w.Root) // the galaxies' textures, made here and not on the draw path
	}
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.w == w && v.key == key {
		return
	}
	fresh := v.w == nil || v.w.Root == nil || w.Root == nil || v.w.Root.Kind != w.Root.Kind
	v.w, v.key, v.pal, v.dirty = w, key, pal, true
	if w == nil || w.Root == nil {
		v.ch = nil
		return
	}
	v.ch = newChain(w.Root)
	if fresh {
		v.home()
	} else {
		v.clamp()
	}
}

// Clear drops the sky, as when the map closes.
func (v *View) Clear() {
	v.mu.Lock()
	v.w, v.ch, v.canvas = nil, nil, nil
	v.mu.Unlock()
}

// Shown says whether there is a sky to draw.
func (v *View) Shown() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.ch != nil
}

// Zoom is the view's log10 width in metres, and the frame it is centred in.
func (v *View) Zoom() (z float64, master int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.z, v.m
}

func (v *View) home() {
	_, hi := v.ch.zRange()
	v.z, v.tz, v.m, v.pick, v.sel = hi, hi, 0, -1, selection{}
	v.off = world.Vec{}
	for _, k := range []string{"milky_way", "sol"} { // the home galaxy centred, so its disc is in the middle when it fills the view
		if p, ok := v.ch.objPos(0, k); ok {
			v.off = p
			break
		}
	}
	v.dirty = true
}

// Do applies a camera key.
func (v *View) Do(c Cmd) {
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil {
		return
	}
	v.hold, v.focus = hold{}, focusTo{}
	switch c.Op {
	case ZoomIn:
		v.tz -= zoomStep
	case ZoomOut:
		v.tz += zoomStep
	case Home:
		v.home()
	case Next, Prev:
		v.cycle(c.Op == Next)
	default:
		f := panStep * v.width() / v.ch[v.m].mx
		if c.Fast {
			f *= 4
		}
		d := [...]world.Vec{PanLeft: {X: -f}, PanRight: {X: f}, PanUp: {Y: -f}, PanDown: {Y: f}}[c.Op]
		v.off = world.Vec{X: v.off.X + d.X, Y: v.off.Y + d.Y}
	}
	v.clamp()
	v.dirty = true
}

func (v *View) width() float64 { return math.Pow(10, v.z) }

// cycle centres the next pickable object of the master frame.
func (v *View) cycle(fwd bool) {
	f := v.ch[v.m].f
	n, step := len(f.Objects), 1
	if !fwd {
		step = -1
	}
	for i := 1; i <= n; i++ {
		j := ((v.pick+i*step)%n + n) % n
		if o := &f.Objects[j]; o.Pick {
			v.pick, v.off, v.sel = j, o.Pos, selection{fi: v.m, key: o.Key}
			return
		}
	}
}

// clamp keeps the zoom inside the chain and the centre inside its frame.
func (v *View) clamp() {
	lo, hi := v.ch.zRange()
	v.tz = math.Max(lo, math.Min(hi, v.tz))
	v.z = math.Max(lo, math.Min(hi, v.z))
	r := v.ch[v.m].f.Radius * 1.1
	v.off = world.Vec{X: math.Max(-r, math.Min(r, v.off.X)), Y: math.Max(-r, math.Min(r, v.off.Y))}
}

// Update eases the zoom toward its target by dt and keeps the master frame
// current. It reports whether the picture changed.
func (v *View) Update(dt time.Duration) bool {
	defer v.flush()
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil {
		return false
	}
	k := 1 - math.Exp(-float64(dt)/float64(easeTau))
	if d := v.tz - v.z; math.Abs(d) > 1e-4 {
		v.z += d * k
		if math.Abs(v.tz-v.z) < 1e-3 {
			v.z = v.tz
		}
		v.dirty = true
	}
	if m := v.ch.master(v.z); m != v.m {
		v.off, v.m, v.dirty = v.ch.centreIn(v.m, m, v.off), m, true
	}
	if v.hold.on || v.focus.on {
		v.follow(k)
		v.clamp()
		v.dirty = true
	}
	return v.dirty
}

// Render draws the sky on a canvas of w by h art pixels and returns it (the
// same image until something changes). Nil when there is no sky.
func (v *View) Render(w, h int) *image.RGBA {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.ch == nil || w <= 0 || h <= 0 {
		return nil
	}
	if v.canvas == nil || v.canvas.Bounds().Dx() != w || v.canvas.Bounds().Dy() != h {
		v.canvas, v.dirty = image.NewRGBA(image.Rect(0, 0, w, h)), true
	}
	if !v.dirty {
		return v.canvas
	}
	v.dirty = false
	snapped := v.snapped(w) // whole art pixels, in the master frame
	under, over, t := v.ch.layers(v.z)
	v.paint(v.canvas, under, snapped)
	if over != under {
		top := image.NewRGBA(v.canvas.Rect)
		v.paint(top, over, snapped)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if Bayer(x, y) < t {
					i := v.canvas.PixOffset(x, y)
					copy(v.canvas.Pix[i:i+4], top.Pix[i:i+4])
				}
			}
		}
	}
	v.overlay(v.canvas, snapped)
	return v.canvas
}
