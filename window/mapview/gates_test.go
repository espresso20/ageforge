package mapview

import (
	"crypto/sha256"
	"image/color"
	"math"
	"testing"

	"github.com/espresso20/ageforge/world"
)

func nonBg(pix []uint8) int {
	n := 0
	for p := 0; p < len(pix); p += 4 {
		if (color.RGBA{R: pix[p], G: pix[p+1], B: pix[p+2], A: 255}) != testPal.Bg {
			n++
		}
	}
	return n
}

func TestGroundAgeCannotLeaveThePlanetFrame(t *testing.T) {
	root := build(t, "space_age").Root.Find(world.FramePlanet)
	v := New()
	v.Set(&world.World{Age: "bronze_age", Root: root}, 1, testPal)
	lo, hi := v.ch.zRange()
	if v.open != 0 || len(v.ch) != 1 || hi != v.ch[0].bmax || lo >= hi {
		t.Fatalf("ground age chain: open %d, %d frames, range %v..%v", v.open, len(v.ch), lo, hi)
	}
	for i := 0; i < 20; i++ {
		v.Wheel(150, 85, tw, th, -3)
	}
	settle(v)
	if v.z > hi+1e-9 || v.ch[v.m].f.Kind != world.FramePlanet {
		t.Errorf("zoomed out to %.2f in %v, past %.2f", v.z, v.ch[v.m].f.Kind, hi)
	}
}

func firstStar(t *testing.T, f *world.Frame) *world.Object {
	t.Helper()
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Pick && o.Key != "sol" {
			return o
		}
	}
	t.Fatal("no pickable star in the locked frame")
	return nil
}

func TestLockedStarSelectsButDoesNotDive(t *testing.T) {
	v := New()
	v.Set(build(t, "space_age"), 1, testPal)
	z, ok := zFor(v, 0)
	if !ok {
		t.Fatal("no zoom inside the locked frame")
	}
	star := firstStar(t, v.ch[0].f)
	at(v, z, 0, star.Pos)
	x, y := screenOf(v, 0, star)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	st := v.state()
	if st.Sel == nil || st.Sel.Pos != star.Pos || st.Lock != "interstellar_age" {
		t.Fatalf("selected %v, lock %q; want %s and interstellar_age", st.Sel, st.Lock, star.Key)
	}
	if v.focus.on || v.tz != z {
		t.Errorf("focus on a locked star moved the camera: focus %v, tz %.2f -> %.2f", v.focus.on, z, v.tz)
	}
}

func TestTabSkipsLockedObjects(t *testing.T) {
	v := New()
	v.Set(build(t, "space_age"), 1, testPal)
	z, _ := zFor(v, 0)
	at(v, z, 0, centreOf(v, 0))
	off, sel := v.off, v.sel
	v.Do(Cmd{Op: Next})
	v.Do(Cmd{Op: Prev})
	if v.off != off || v.sel != sel {
		t.Errorf("Tab moved within the locked frame: off %v -> %v", off, v.off)
	}
}

func deepView(t *testing.T, key string) (*View, *world.Frame) {
	t.Helper()
	v := New()
	v.Set(build(t, "space_age"), 1, testPal)
	v.bod.wg.Wait()
	z, _ := zFor(v, 1)
	at(v, z, 1, centreOf(v, 1))
	v.swapTail(key)
	tail := &v.ch[len(v.ch)-1]
	if tail.f.Anchor != key {
		t.Fatalf("tail is %q, want %q", tail.f.Anchor, key)
	}
	v.bod.wg.Wait()
	z = tail.lo + (tail.hi-tail.lo)*.4
	v.z, v.tz = z, z
	v.m = v.ch.master(z)
	o := tail.f.Objects[0]
	v.off = v.ch.centreIn(len(v.ch)-1, v.m, o.Pos)
	v.Update(0)
	return v, tail.f
}

func TestEverySolPlanetHasACloseUp(t *testing.T) {
	for _, k := range append(append([]string{}, world.SolNames[:]...), "pluto") {
		var sums [2][32]byte
		for i := range sums {
			v, _ := deepView(t, k)
			img := v.Render(tw, th)
			sums[i] = sha256.Sum256(img.Pix)
			if i == 0 && nonBg(img.Pix) < 200 {
				t.Errorf("%s close-up is nearly empty", k)
			}
		}
		if sums[0] != sums[1] {
			t.Errorf("%s close-up renders differently twice", k)
		}
	}
}

func TestJupiterDrawsItsFourMoons(t *testing.T) {
	v, f := deepView(t, "jupiter")
	n := 0
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Pick && o.Key != "jupiter" {
			n++
		}
	}
	if n != 4 {
		t.Fatalf("Jupiter's frame holds %d named moons, want 4", n)
	}
	img := v.Render(tw, th)
	if nonBg(img.Pix) < 500 {
		t.Error("Jupiter's close-up is nearly empty")
	}
}
