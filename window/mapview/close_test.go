package mapview

import (
	"bytes"
	"image"
	"math"
	"testing"

	"github.com/espresso20/ageforge/world"
)

const planetFi = 5

func homeView(t *testing.T) *View {
	t.Helper()
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	v.bod.wg.Wait()
	if len(v.ch) <= planetFi || v.ch[planetFi].f.Kind != world.FramePlanet {
		t.Fatalf("chain has no Planet frame at %d", planetFi)
	}
	return v
}

// planetZs are n zooms spread through the Planet frame's own stretch, with no
// crossfade on screen.
func planetZs(v *View, fi, n int) (zs []float64) {
	lo, hi := v.ch.zRange()
	var all []float64
	for z := hi; z >= lo; z -= .02 {
		if u, o, _ := v.ch.layers(z); u == o && v.ch.master(z) == fi {
			all = append(all, z)
		}
	}
	for k := 0; k < n && len(all) > 0; k++ {
		zs = append(zs, all[(2*k+1)*len(all)/(2*n)])
	}
	return zs
}

func planetObj(v *View, key string) *world.Object { return v.objAt(planetFi, key) }

func TestHomeSameSeedSamePixels(t *testing.T) {
	var shots [2]*image.RGBA
	for i := range shots {
		v := homeView(t)
		at(v, planetZs(v, planetFi, 3)[1], planetFi, world.Vec{})
		shots[i] = v.Render(tw, th)
	}
	if !bytes.Equal(shots[0].Pix, shots[1].Pix) {
		t.Fatal("two views of one seed drew different close-ups")
	}
}

// Wherever the Moon is on screen, Earth is a sphere of 12 art pixels or more.
func TestEarthIsBigWhereTheMoonShows(t *testing.T) {
	v := homeView(t)
	earth, moon := planetObj(v, world.SolNames[world.Home]), planetObj(v, "the_moon")
	if earth == nil || moon == nil {
		t.Fatal("the Planet frame lacks Earth or the Moon")
	}
	seen := 0
	for _, z := range planetZs(v, planetFi, 40) {
		at(v, z, planetFi, world.Vec{})
		p := v.penFor(&image.RGBA{Rect: image.Rect(0, 0, tw, th)}, planetFi, v.snapped(tw))
		x, y := p.at(moon.Pos)
		if x < 0 || y < 0 || x >= tw || y >= th {
			continue
		}
		seen++
		if d := 2*drawnR(p, earth) - 1; d < 12 {
			t.Errorf("z %.2f: the Moon is on screen and Earth is %.0f px across", z, d)
		}
	}
	if seen == 0 {
		t.Fatal("the Moon is never on screen in the Planet frame")
	}
}

// Stations, ships and the swarm keep off Earth's disc.
func TestTrafficKeepsOffEarth(t *testing.T) {
	v := homeView(t)
	earth := planetObj(v, world.SolNames[world.Home])
	full := v.ch[planetFi].f
	bare := *full
	bare.Objects = nil
	for _, o := range full.Objects {
		if o.Key == earth.Key {
			bare.Objects = append(bare.Objects, o)
		}
	}
	for _, z := range planetZs(v, planetFi, 6) {
		at(v, z, planetFi, world.Vec{})
		v.ch[planetFi].f = &bare
		a := v.Render(tw, th)
		a = cloneImg(a)
		v.ch[planetFi].f = full
		b := v.Render(tw, th)
		p := v.penFor(b, planetFi, v.snapped(tw))
		cx, cy := p.at(earth.Pos)
		r := drawnR(p, earth) - 1
		bad := 0
		for y := 0; y < th; y++ {
			for x := 0; x < tw; x++ {
				if math.Hypot(float64(x)-cx, float64(y)-cy) <= r && a.RGBAAt(x, y) != b.RGBAAt(x, y) {
					bad++
				}
			}
		}
		if bad > 0 {
			t.Errorf("z %.2f: %d pixels on Earth's disc differ once the traffic is drawn", z, bad)
		}
	}
}

func cloneImg(a *image.RGBA) *image.RGBA {
	b := image.NewRGBA(a.Rect)
	copy(b.Pix, a.Pix)
	return b
}

// The selection ring is drawn outside a small object, never over it.
func TestRingLeavesSmallObjectsAlone(t *testing.T) {
	cases := []struct {
		fi  int
		key string
		z   func(v *View) float64
	}{
		{4, "mercury", func(v *View) float64 { return sysZ }},
		{planetFi, "the_moon", func(v *View) float64 { return planetZs(v, planetFi, 3)[0] }},
	}
	v := homeView(t)
	for _, o := range v.ch[planetFi].f.Objects {
		if o.Kind == world.KindStation || o.Kind == world.KindShip {
			cases = append(cases, struct {
				fi  int
				key string
				z   func(v *View) float64
			}{planetFi, o.Key, func(v *View) float64 { return planetZs(v, planetFi, 3)[1] }})
			break
		}
	}
	for _, c := range cases {
		o := v.objAt(c.fi, c.key)
		if o == nil {
			t.Errorf("%s/%s missing", v.ch[c.fi].f.Key, c.key)
			continue
		}
		v.sel = selection{}
		at(v, c.z(v), c.fi, o.Pos)
		bare := cloneImg(v.Render(tw, th))
		v.sel = selection{fi: c.fi, key: c.key}
		v.dirty = true
		ringed := v.Render(tw, th)
		p := v.penFor(ringed, c.fi, v.snapped(tw))
		x, y := p.at(o.Pos)
		r := int(math.Ceil(drawnR(p, o)))
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				px, py := int(math.Round(x))+dx, int(math.Round(y))+dy
				if bare.RGBAAt(px, py) != ringed.RGBAAt(px, py) {
					t.Errorf("%s: selecting it changed its own pixel at %d,%d", c.key, dx, dy)
				}
			}
		}
	}
}

// A click on any System planet or Planet-frame body picks it, at three zooms.
func TestPickHitsEveryBody(t *testing.T) {
	v := homeView(t)
	for _, fi := range []int{4, planetFi} {
		for _, z := range planetZs(v, fi, 3) {
			for i := range v.ch[fi].f.Objects {
				o := &v.ch[fi].f.Objects[i]
				if !o.Pick || (fi == 4 && o.Kind != world.KindPlanet) {
					continue
				}
				at(v, z, fi, o.Pos)
				x, y := screenOf(v, fi, o)
				got, ok := v.near(int(math.Round(x)), int(math.Round(y)), tw, th)
				if !ok || got.key != o.Key {
					t.Errorf("%s z %.2f: a click on %s picked %q (hit %v)", v.ch[fi].f.Key, z, o.Key, got.key, ok)
				}
			}
		}
	}
}

// Earth picked in the System stays marked once the Planet frame takes over.
func TestSelectionFollowsEarthDown(t *testing.T) {
	v := homeView(t)
	home := world.SolNames[world.Home]
	v.sel = selection{fi: 4, key: home}
	at(v, planetZs(v, planetFi, 3)[1], planetFi, world.Vec{})
	if got := v.selDrawn(); got != planetFi {
		t.Errorf("a System selection draws in frame %d inside the Planet frame", got)
	}
	at(v, sysZ, 4, world.Vec{})
	if got := v.selDrawn(); got != 4 {
		t.Errorf("a System selection draws in frame %d at system zoom", got)
	}
}

// Pans and zooms inside the Planet frame reuse the sprites.
func TestHomeCameraMovesDoNotRebuildSprites(t *testing.T) {
	v := homeView(t)
	zs := planetZs(v, planetFi, 3)
	at(v, zs[1], planetFi, world.Vec{})
	v.Render(tw, th)
	v.bod.wg.Wait()
	built := v.bod.builds.Load()
	for _, c := range []Cmd{{Op: PanLeft}, {Op: PanUp}, {Op: PanRight, Fast: true}, {Op: PanDown}} {
		v.Do(c)
		v.Render(tw, th)
	}
	v.Render(tw, th)
	v.bod.wg.Wait()
	if n := v.bod.builds.Load(); n != built {
		t.Errorf("sprites built went from %d to %d while the camera panned", built, n)
	}
}
