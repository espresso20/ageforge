package mapview

import (
	"bytes"
	"testing"

	"github.com/espresso20/ageforge/world"
)

const sysZ = 13.31 // the whole of Sol's system in view

func sysView(t *testing.T) *View {
	t.Helper()
	v := nearView(t)
	at(v, sysZ, 4, world.Vec{})
	return v
}

func TestOrrerySameSeedSamePixels(t *testing.T) {
	a, b := sysView(t).Render(tw, th), sysView(t).Render(tw, th)
	if !bytes.Equal(a.Pix, b.Pix) {
		t.Fatal("two views of one seed drew different systems")
	}
}

func planets(v *View) (ps []*world.Object) {
	for i := range v.ch[4].f.Objects {
		if o := &v.ch[4].f.Objects[i]; o.Kind == world.KindPlanet {
			ps = append(ps, o)
		}
	}
	return ps
}

// The side of a planet that faces its star is the brighter one.
func TestPlanetsAreLitFromTheStar(t *testing.T) {
	v := sysView(t)
	half := func(s *sprite, left bool) (sum, n int) {
		for y := 0; y < 2*s.ext+1; y++ {
			for x := 0; x < 2*s.ext+1; x++ {
				if s.img.Pix[s.img.PixOffset(x, y)+3] != 0 && (x < s.ext) == left && x != s.ext {
					sum += luma(s.img.RGBAAt(x, y))
					n++
				}
			}
		}
		return
	}
	for _, o := range planets(v) {
		for _, c := range []struct{ dir, r int }{{0, 9}, {8, 9}, {0, 5}} { // sun to the right (dir 0), to the left (8)
			s := buildBody(&v.pal, o, c.r, c.dir)
			ls, ln := half(s, true)
			rs, rn := half(s, false)
			l, r := ls/max(1, ln), rs/max(1, rn)
			if (c.dir == 0 && r <= l) || (c.dir == 8 && l <= r) {
				t.Errorf("%s r=%d dir=%d: near half %d, far half %d", o.Key, c.r, c.dir, max(l, r), min(l, r))
			}
		}
	}
}

// Every planet has at least three lit art pixels, at the size it draws and at
// the smallest the map ever draws.
func TestEveryPlanetShows(t *testing.T) {
	v := sysView(t)
	img := v.Render(tw, th)
	p := v.penFor(img, 4, v.snapped(tw))
	floor := lum(v.pal.Mix(world.ToneWhite, 40))
	for _, o := range planets(v) {
		for _, r := range []int{1, stepR(p.radius(o))} {
			s, lit := buildBody(&v.pal, o, r, 3), 0
			for y := 0; y < 2*s.ext+1; y++ {
				for x := 0; x < 2*s.ext+1; x++ {
					if s.img.Pix[s.img.PixOffset(x, y)+3] != 0 && lum(s.img.RGBAAt(x, y)) > floor {
						lit++
					}
				}
			}
			if lit < 3 {
				t.Errorf("%s at r=%d: %d lit pixels", o.Key, r, lit)
			}
		}
	}
}

// Moving the camera at one scale reuses the sprites.
func TestCameraMovesDoNotRebuildSprites(t *testing.T) {
	v := sysView(t)
	v.Render(tw, th)
	built := v.bod.builds.Load()
	if built == 0 {
		t.Fatal("no sprites built for a system in view")
	}
	for _, c := range []Cmd{{Op: PanLeft}, {Op: PanUp}, {Op: PanRight, Fast: true}, {Op: PanDown}} {
		v.Do(c)
		v.Render(tw, th)
	}
	v.tz = sysZ + .01
	v.Update(1 << 30)
	v.Render(tw, th)
	if n := v.bod.builds.Load(); n != built {
		t.Errorf("sprites built went from %d to %d while the camera moved", built, n)
	}
}
