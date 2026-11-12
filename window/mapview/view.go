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
	open     int       // the first chain frame the age lets the player enter; those above it are locked
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
	top      *image.RGBA // the deeper frame of a crossfade, kept between frames
	tex      texCache
	art      artCache
	bod      bodyCache
	star     string        // the near star whose system hangs under the neighbourhood ("" Sol's)
	clock    time.Duration // the frame clock the lights blink on
	tick     int           // half seconds of it
}

func New() *View { return &View{dirty: true} }

// Set gives the view the sky. Nothing is rebuilt unless key changed.
func (v *View) Set(w *world.World, key uint64, pal world.Palette) {
	if w != nil && w.Root != nil {
		v.tex.warm(w.Root) // the galaxies' textures, made here and not on the draw path
		v.art.warm(w.Root.Find(world.FrameRegion), w.Seed)
		if home := w.Root.Find(world.FramePlanet); home != nil {
			v.bod.warm(pal, home)
		}
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
	if f := w.System(v.star); v.star != "" && f != nil && v.ch.idx(world.FrameSystem) > 0 {
		v.ch = v.ch.withSystem(f) // the system the player was in outlives a rebuild
	} else {
		v.star = ""
	}
	v.open = max(0, int(world.GateOf(w.Age).Open)-int(w.Root.Kind))
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

// home is the whole of the outermost open frame in view, centred.
func (v *View) home() {
	z := v.ch[0].bmax
	if v.open > 0 {
		z = v.ch[v.open].lo
	}
	v.z, v.tz, v.m = z, z, v.ch.masterAt(z, look{})
	v.centre()
}

// centre puts the Milky Way in the middle at the Local Group and in the upper
// half of the Galaxy band (so the whole disc is in view), Sol anywhere deeper.
func (v *View) centre() {
	key := "sol"
	if v.ch[0].f.Kind == world.FrameLocalGroup && (len(v.ch) < 2 || v.z >= (v.ch[1].bmin+v.ch[1].bmax)/2) {
		key = "milky_way"
	}
	p, _ := v.ch.objPos(0, key)
	v.pick, v.sel, v.off = -1, selection{}, v.ch.centreIn(0, v.m, p)
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
		v.centre()
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

// look is the camera's centre, for the frames that hang on Sol.
func (v *View) look() look { return look{m: v.m, off: v.off, on: true} }

func (v *View) width() float64 { return math.Pow(10, v.z) }

// cycle centres the next pickable object of the master frame.
func (v *View) cycle(fwd bool) {
	if v.m < v.open { // a locked frame has nothing to step to
		return
	}
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
	r := v.ch[v.m].reach
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
	v.clock += dt
	if t := int(v.clock / (500 * time.Millisecond)); t != v.tick {
		v.tick = t
		if v.z < v.ch[len(v.ch)-1].hi { // only a close-up has lights
			v.dirty = true
		}
	}
	if d := v.tz - v.z; math.Abs(d) > 1e-4 {
		v.z += d * k
		if math.Abs(v.tz-v.z) < 1e-3 {
			v.z = v.tz
		}
		v.dirty = true
	}
	if m := v.ch.masterAt(v.z, v.look()); m != v.m {
		v.off, v.m, v.dirty = v.ch.centreIn(v.m, m, v.off), m, true
	}
	switch v.m { // the star or planet nearest the centre is the one a zoom would enter
	case v.ch.idx(world.FrameNeighbourhood):
		v.nearSystem(v.off)
	case v.ch.idx(world.FrameSystem):
		v.nearTail(v.off)
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
	under, over, t := v.ch.layersAt(v.z, v.look())
	v.paint(v.canvas, under, snapped)
	if over != under {
		if v.top == nil || v.top.Rect != v.canvas.Rect {
			v.top = image.NewRGBA(v.canvas.Rect)
		}
		top := v.top
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
