package mapview

import (
	"crypto/sha256"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/world"
)

var testPal = func() world.Palette {
	p := world.Palette{Bg: color.RGBA{R: 4, G: 4, B: 12, A: 255}}
	for i := range p.Tones {
		p.Tones[i] = color.RGBA{R: uint8(40 + i*12), G: uint8(255 - i*9), B: uint8(90 + i*7), A: 255}
	}
	return p
}()

func build(t *testing.T, age string) *world.World {
	t.Helper()
	st := fixture.State(fixture.Options{Age: age, Seed: 7, Tick: 43200})
	m := mapmodel.NewBuilder(mapmodel.NewCatalog()).Build(&st, nil)
	return world.Build(m, 7)
}

func TestChainBandsAreContinuous(t *testing.T) {
	c := newChain(build(t, "galactic_age").Root)
	if len(c) != 6 {
		t.Fatalf("chain has %d frames, want 6", len(c))
	}
	for i := 1; i < len(c); i++ {
		if c[i].lo > c[i].hi {
			t.Errorf("frame %d handoff runs backward", i)
		}
		// no zoom is left with nothing deeper than the frame above it
		if c.weight(i, c[i].hi+0.1) != 0 || c.weight(i, c[i].lo-0.1) != 1 {
			t.Errorf("frame %d weight is not 0 above and 1 below its handoff", i)
		}
	}
	lo, hi := c.zRange()
	if lo >= hi {
		t.Errorf("zoom range %v..%v", lo, hi)
	}
	for z := lo; z <= hi; z += 0.05 { // every zoom shows something, never two layers apart
		u, o, w := c.layers(z)
		if o-u > 1 || w < 0 || w > 1 {
			t.Fatalf("z=%.2f layers %d,%d weight %v", z, u, o, w)
		}
	}
}

func TestSolAnchorIsTheLocalGroupMarker(t *testing.T) {
	w := build(t, "galactic_age")
	c := newChain(w.Root)
	sol, _ := c.objPos(0, "sol")
	region := c[2] // Orion arm: anchored on Sol in the galaxy
	// back in local-group units the Orion arm's origin is the marker
	back := c.centreIn(2, 0, world.Vec{})
	if math.Abs(back.X-sol.X) > 1e-6 || math.Abs(back.Y-sol.Y) > 1e-6 {
		t.Errorf("region origin at %v in local-group units, want the marker %v", back, sol)
	}
	_ = region
}

func TestCentreConversionRoundTrips(t *testing.T) {
	c := newChain(build(t, "galactic_age").Root)
	v := world.Vec{X: 0.3, Y: -0.2}
	for i := range c {
		for j := max(0, i-1); j <= min(len(c)-1, i+1); j++ { // far jumps lose digits, which is why the camera keeps a master
			got := c.centreIn(j, i, c.centreIn(i, j, v))
			if math.Abs(got.X-v.X) > 1e-6 || math.Abs(got.Y-v.Y) > 1e-6 {
				t.Errorf("%d->%d->%d: %v, want %v", i, j, i, got, v)
			}
		}
	}
}

func TestZoomClampsAndEases(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	lo, hi := v.ch.zRange()
	for i := 0; i < 60; i++ {
		v.Do(Cmd{Op: ZoomIn})
	}
	for i := 0; i < 20; i++ {
		v.Update(50 * time.Millisecond)
	}
	if z, m := v.Zoom(); math.Abs(z-lo) > 1e-6 || m != 5 {
		t.Errorf("zoomed all the way in: z=%v master=%d, want %v and 5", z, m, lo)
	}
	for i := 0; i < 80; i++ {
		v.Do(Cmd{Op: ZoomOut})
	}
	v.Update(time.Hour)
	if z, m := v.Zoom(); math.Abs(z-hi) > 1e-6 || m != 0 {
		t.Errorf("zoomed all the way out: z=%v master=%d, want %v and 0", z, m, hi)
	}
	v.Do(Cmd{Op: ZoomIn})
	v.Update(10 * time.Millisecond)
	if z, _ := v.Zoom(); z >= hi || z <= hi-zoomStep {
		t.Errorf("one step eases part of the way: z=%v from %v", z, hi)
	}
}

func TestArtScale(t *testing.T) {
	for w, want := range map[int]int{300: 2, 700: 2, 1200: 3, 1750: 5} {
		if got := ArtScale(w); got != want {
			t.Errorf("ArtScale(%d) = %d, want %d", w, got, want)
		}
	}
}

func TestPanSnapsToArtPixels(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	v.Do(Cmd{Op: PanLeft})
	a := v.Render(300, 170)
	b := v.Render(300, 170)
	if a != b {
		t.Error("an unchanged view should return the cached canvas")
	}
}

func TestCrossfadeDithers(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	mid := (v.ch[1].lo + v.ch[1].hi) / 2
	v.z, v.tz = mid, mid
	v.m = v.ch.master(mid)
	under, over, w := v.ch.layers(mid)
	if over != 1 || under != 0 || math.Abs(w-0.5) > 1e-6 {
		t.Fatalf("mid handoff layers %d,%d weight %v", under, over, w)
	}
	// the dither mask shows about half of each layer
	n := 0
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			if Bayer(x, y) < w {
				n++
			}
		}
	}
	if n < 700 || n > 900 {
		t.Errorf("%d of 1600 pixels took the child at weight 0.5", n)
	}
}

// TestRenderIsDeterministic: the same world and camera give the same pixels,
// at every scale of the chain.
func TestRenderIsDeterministic(t *testing.T) {
	for _, age := range []string{"galactic_age", "space_age"} {
		w := build(t, age)
		lo, hi := newChain(w.Root).zRange()
		for k := 0; k <= 6; k++ {
			z := hi - (hi-lo)*float64(k)/6
			var sums [2][32]byte
			for i := range sums {
				v := New()
				v.Set(w, 1, testPal)
				v.z, v.tz = z, z
				v.m = v.ch.master(z)
				v.off = v.ch.centreIn(len(v.ch)-1, v.m, world.Vec{}) // looking at the home world
				v.Update(0)
				img := v.Render(300, 170)
				sums[i] = sha256.Sum256(img.Pix)
				bg, other := testPal.Bg, 0
				for p := 0; p < len(img.Pix); p += 4 {
					if (color.RGBA{R: img.Pix[p], G: img.Pix[p+1], B: img.Pix[p+2], A: 255}) != bg {
						other++
					}
				}
				if i == 0 && other == 0 {
					t.Errorf("%s at step %d (z=%.1f) drew nothing", age, k, z)
				}
			}
			if sums[0] != sums[1] {
				t.Errorf("%s step %d renders differently twice", age, k)
			}
		}
	}
}

func TestNextCentresAPickableObject(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	v.Do(Cmd{Op: Next})
	v.Do(Cmd{Op: Next})
	f := v.ch[v.m].f
	ok := false
	for _, o := range f.Objects {
		ok = ok || (o.Pick && o.Pos == v.off)
	}
	if !ok {
		t.Errorf("centre %v is not a pickable object of %s", v.off, f.Key)
	}
}

func TestSystemRootedAgeHasNoLocalGroup(t *testing.T) {
	v := New()
	v.Set(build(t, "space_age"), 1, testPal)
	if v.ch[0].f.Kind != world.FrameSystem || len(v.ch) != 2 {
		t.Errorf("space age chain starts at %v with %d frames, want the system and the planet", v.ch[0].f.Kind, len(v.ch))
	}
	v.Clear()
	if v.Shown() || v.Render(10, 10) != nil {
		t.Error("a cleared view still draws")
	}
}
