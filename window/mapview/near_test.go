package mapview

import (
	"crypto/sha256"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/espresso20/ageforge/world"
)

func nearObj(t *testing.T, f *world.Frame, key string) *world.Object {
	t.Helper()
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Key == key && o.Kind == world.KindStar {
			return o
		}
	}
	t.Fatalf("no star %s", key)
	return nil
}

// interstellar is a view at the Interstellar Age, zoomed inside the neighbourhood.
func interstellar(t *testing.T) (*View, int) {
	t.Helper()
	v := New()
	v.Set(build(t, world.AgeInterstellar), 1, testPal)
	v.bod.wg.Wait()
	ni := v.ch.idx(world.FrameNeighbourhood)
	if ni != v.open || v.open < 1 {
		t.Fatalf("neighbourhood at %d, open %d", ni, v.open)
	}
	z, ok := zFor(v, ni)
	if !ok {
		t.Fatal("no zoom inside the neighbourhood")
	}
	at(v, z, ni, world.Vec{})
	return v, ni
}

func TestInterstellarFocusDivesIntoAStarSystem(t *testing.T) {
	v, ni := interstellar(t)
	star := nearObj(t, v.ch[ni].f, "alpha_centauri")
	z := v.z
	x, y := screenOf(v, ni, star)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	si := v.ch.idx(world.FrameSystem)
	if v.star != "alpha_centauri" || v.ch[si].f.Anchor != "alpha_centauri" || v.tz >= z {
		t.Fatalf("star %q, system %q, tz %.2f from %.2f: no dive", v.star, v.ch[si].f.Anchor, v.tz, z)
	}
	settle(v)
	if v.ch[v.m].f.Kind != world.FrameSystem {
		t.Fatalf("settled in %v, want the system", v.ch[v.m].f.Kind)
	}
	var planet *world.Object
	for i := range v.ch[si].f.Objects {
		if o := &v.ch[si].f.Objects[i]; o.Kind == world.KindPlanet {
			planet = o
		}
	}
	if planet == nil {
		t.Fatal("alpha centauri holds no planet")
	}
	at(v, v.ch[si].bmin+(v.ch[si].bmax-v.ch[si].bmin)*.5, si, world.Vec{})
	v.swapTail(planet.Key)
	if tail := v.ch[len(v.ch)-1].f; tail.Kind != world.FramePlanet || tail.Anchor != planet.Key {
		t.Fatalf("tail %v %q, want the planet %s", tail.Kind, tail.Anchor, planet.Key)
	}
}

func TestSpaceAgeFocusOnANearStarDoesNotDive(t *testing.T) {
	v := New()
	v.Set(build(t, world.AgeSpace), 1, testPal)
	z, _ := zFor(v, 0)
	at(v, z, 0, world.Vec{})
	star := nearObj(t, v.ch[0].f, "alpha_centauri")
	x, y := screenOf(v, 0, star)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	if v.star != "" || v.focus.on || v.tz != z {
		t.Errorf("locked star dived: star %q, focus %v, tz %.2f -> %.2f", v.star, v.focus.on, z, v.tz)
	}
}

func TestSystemSurvivesARebuild(t *testing.T) {
	v, _ := interstellar(t)
	v.swapSystem("sirius")
	w := build(t, world.AgeInterstellar)
	v.Set(w, 2, testPal)
	if v.star != "sirius" || v.ch[v.ch.idx(world.FrameSystem)].f.Anchor != "sirius" {
		t.Errorf("system after a rebuild: %q", v.star)
	}
}

func TestColonyMarkSitsBesideItsStar(t *testing.T) {
	v, ni := interstellar(t)
	f := v.ch[ni].f
	n := 0
	for i := range f.Objects {
		c := &f.Objects[i]
		if c.Kind != world.KindColony {
			continue
		}
		n++
		star := nearObj(t, f, c.Label)
		if star.Pos != c.Pos {
			t.Errorf("colony %s at %v, its star %s at %v", c.Key, c.Pos, star.Key, star.Pos)
		}
		img := image.NewRGBA(image.Rect(0, 0, 80, 80))
		fill(img, testPal.Bg)
		p := pen{img: img, pal: &testPal, f: f, w: 80, h: 80}
		p.colonyMark(40, 40, star)
		lit, rad := 0, starRadius(star)+1
		for y := 0; y < 80; y++ {
			for x := 0; x < 80; x++ {
				if img.RGBAAt(x, y) == testPal.Bg {
					continue
				}
				lit++
				if math.Hypot(float64(x-40), float64(y-40)) <= rad {
					t.Errorf("colony mark covers the star at %d,%d", x, y)
				}
			}
		}
		if lit == 0 {
			t.Errorf("colony mark at %s draws nothing", star.Key)
		}
	}
	if n == 0 {
		t.Skip("the fixture has no colonies")
	}
}

func TestLockedLabelsDimWithTheirFrame(t *testing.T) {
	grey := luma(mixRGB(testPal.Bg, testPal.Tones[world.ToneWhite], lockDim))
	for _, c := range []struct {
		age    string
		locked bool
	}{{world.AgeSpace, true}, {world.AgeInterstellar, false}} {
		v := New()
		v.Set(build(t, c.age), 1, testPal)
		v.face, v.names = boxFace{}, func(k string) string { return k }
		fi := v.ch.idx(world.FrameNeighbourhood)
		z, _ := zFor(v, fi)
		at(v, z, fi, world.Vec{})
		img := image.NewRGBA(image.Rect(0, 0, tw, th))
		fill(img, testPal.Bg)
		v.overlay(img, v.off)
		lit, bright := 0, 0
		for p := 0; p < len(img.Pix); p += 4 {
			c := color.RGBA{R: img.Pix[p], G: img.Pix[p+1], B: img.Pix[p+2], A: 255}
			if c == testPal.Bg {
				continue
			}
			lit++
			if luma(c) > grey {
				bright++
			}
		}
		if lit == 0 {
			t.Fatalf("%s: no labels drawn", c.age)
		}
		if (bright == 0) != c.locked {
			t.Errorf("%s: %d bright label pixels of %d; locked %v", c.age, bright, lit, c.locked)
		}
	}
}

func TestGeneratedPlanetCloseUpsDrawAndRepeat(t *testing.T) {
	for _, star := range []string{"alpha_centauri", "epsilon_eridani", "tau_ceti"} {
		var sums [2][32]byte
		for i := range sums {
			v, _ := interstellar(t)
			v.swapSystem(star)
			si := v.ch.idx(world.FrameSystem)
			if v.ch[si].f.Anchor != star || len(v.ch[si].f.Children) == 0 {
				t.Fatalf("%s: no system or no planets", star)
			}
			v.bod.wg.Wait()
			tail := &v.ch[len(v.ch)-1]
			z := tail.lo + (tail.hi-tail.lo)*.4
			v.z, v.tz = z, z
			v.m = v.ch.master(z)
			v.off = v.ch.centreIn(len(v.ch)-1, v.m, tail.f.Objects[0].Pos)
			v.Update(0)
			img := v.Render(tw, th)
			sums[i] = sha256.Sum256(img.Pix)
			if i == 0 && nonBg(img.Pix) < 200 {
				t.Errorf("%s: close-up of %s is nearly empty", star, tail.f.Anchor)
			}
		}
		if sums[0] != sums[1] {
			t.Errorf("%s renders differently twice", star)
		}
	}
}
