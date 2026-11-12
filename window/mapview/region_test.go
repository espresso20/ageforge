package mapview

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/world"
)

func nearView(t *testing.T) *View {
	t.Helper()
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	return v
}

func lum(c color.RGBA) int { return 3*int(c.R) + 6*int(c.G) + int(c.B) }

func TestRegionSameSeedSamePixels(t *testing.T) {
	var shots [2]*image.RGBA
	for i := range shots {
		v := nearView(t)
		z, _ := zFor(v, 2)
		at(v, z, 2, centreOf(v, 2))
		shots[i] = v.Render(tw, th)
	}
	if !bytes.Equal(shots[0].Pix, shots[1].Pix) {
		t.Fatal("two views of one seed drew different pixels")
	}
}

func TestNearArtIsNotRebuiltByCameraMoves(t *testing.T) {
	v := nearView(t)
	z, _ := zFor(v, 2)
	at(v, z, 2, centreOf(v, 2))
	v.Render(tw, th)
	art, tex := v.art.builds.Load(), v.tex.builds.Load()
	if art == 0 {
		t.Fatal("nothing was cached for the arm")
	}
	for i := 0; i < 8; i++ {
		v.Pan(7, -3, tw)
		v.Wheel(tw/2, th/2, tw, th, 0.3)
		v.Render(tw, th)
		at(v, z-0.1*float64(i), 2, world.Vec{X: float64(i) * 90, Y: 40})
		v.Render(tw, th)
	}
	if a, x := v.art.builds.Load(), v.tex.builds.Load(); a != art || x != tex {
		t.Fatalf("camera moves rebuilt the caches: arm %d to %d, galaxies %d to %d", art, a, tex, x)
	}
}

// The galaxy frame is out of the picture across the whole of the Region's
// band: its weight there is exactly zero, so none of its light reaches the canvas.
func TestGalaxyIsGoneInTheRegionBand(t *testing.T) {
	v := nearView(t)
	c := v.ch
	for z := c[2].lo - 0.01; z > c[3].hi+0.01; z -= 0.05 {
		if u, o, w := c.layers(z); u < 2 || o < 2 || w != 1 && u != 2 {
			t.Fatalf("z=%.2f draws frames %d and %d (weight %.2f): the galaxy shows in the Region band", z, u, o, w)
		}
	}
}

func blankPen(v *View, fi int, o *world.Object, r float64) (pen, *image.RGBA, float64, float64) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	fill(img, v.pal.Bg)
	b := v.penFor(img, fi, v.snapped(200))
	b.ppu = r / o.Size
	b.c = o.Pos
	b.cx, b.cy = 100, 100
	x, y := b.at(o.Pos)
	return b, img, x, y
}

func objOf(v *View, key string) *world.Object { return v.objAt(2, key) }

// pieces counts the separate lit blobs on img (8-connected).
func pieces(img *image.RGBA, lit func(color.RGBA) bool) int {
	b := img.Bounds()
	seen := make([]bool, b.Dx()*b.Dy())
	n := 0
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if seen[y*b.Dx()+x] || !lit(img.RGBAAt(x, y)) {
				continue
			}
			n++
			stack := [][2]int{{x, y}}
			seen[y*b.Dx()+x] = true
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						q := [2]int{p[0] + dx, p[1] + dy}
						if q[0] < 0 || q[1] < 0 || q[0] >= b.Dx() || q[1] >= b.Dy() || seen[q[1]*b.Dx()+q[0]] || !lit(img.RGBAAt(q[0], q[1])) {
							continue
						}
						seen[q[1]*b.Dx()+q[0]] = true
						stack = append(stack, q)
					}
				}
			}
		}
	}
	return n
}

func TestRegionFormsAreNotFilledEllipses(t *testing.T) {
	v := nearView(t)
	notBg := func(c color.RGBA) bool { return c != v.pal.Bg }
	forms := map[string]string{"cluster": "pleiades", "emission": "orion_nebula", "remnant": "crab_nebula"}
	for name, key := range forms {
		o := objOf(v, key)
		p, img, _, _ := blankPen(v, 2, o, 70)
		p.object(o)
		lit, inside := 0, 0
		for y := 0; y < 200; y++ {
			for x := 0; x < 200; x++ {
				if (x-100)*(x-100)+(y-100)*(y-100) <= 70*70 {
					inside++
					if notBg(img.RGBAAt(x, y)) {
						lit++
					}
				}
			}
		}
		if lit == 0 {
			t.Fatalf("%s drew nothing", name)
		}
		if name == "cluster" {
			if n := pieces(img, notBg); n < 8 {
				t.Errorf("cluster is %d lit pieces, want several", n)
			}
			continue
		}
		if share := float64(lit) / float64(inside); share > .85 || share < .1 {
			t.Errorf("%s lights %.0f%% of its disc: a filled ellipse or a smudge", name, 100*share)
		}
		seen := map[color.RGBA]bool{}
		for y := 0; y < 200; y++ {
			for x := 0; x < 200; x++ {
				seen[img.RGBAAt(x, y)] = true
			}
		}
		if len(seen) < 4 {
			t.Errorf("%s uses %d colours", name, len(seen))
		}
	}
}

func TestDarkNebulaIsDarkerThanTheSky(t *testing.T) {
	v := nearView(t)
	o := objOf(v, "horsehead_nebula")
	p, img, _, _ := blankPen(v, 2, o, 70)
	sky := p.pal.Mix(world.ToneArm, 60) // a lit background, as the arm's dust makes
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2] = sky.R, sky.G, sky.B
	}
	p.object(o)
	darker := 0
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			if lum(img.RGBAAt(x, y)) < lum(p.pal.Bg) {
				darker++
			}
		}
	}
	if darker < 50 {
		t.Errorf("only %d pixels are darker than the void", darker)
	}
}

func TestArmDrawsNoFlatWash(t *testing.T) {
	v := nearView(t)
	z, _ := zFor(v, 2)
	at(v, z, 2, centreOf(v, 2))
	img := v.Render(tw, th)
	seen := map[color.RGBA]int{}
	for i := 0; i < len(img.Pix); i += 4 {
		seen[color.RGBA{R: img.Pix[i], G: img.Pix[i+1], B: img.Pix[i+2], A: 255}]++
	}
	if seen[v.pal.Bg] < tw*th/3 {
		t.Errorf("the void is %d of %d pixels: the arm's backdrop is a wash", seen[v.pal.Bg], tw*th)
	}
}

func TestClickNearSolPicksSol(t *testing.T) {
	for _, fi := range []int{2, 3} {
		for _, d := range [][2]float64{{0, 0}, {5, 0}, {0, -5}, {-4, 4}} {
			v := nearView(t)
			z, _ := zFor(v, fi)
			at(v, z, fi, centreOf(v, fi))
			sol := objAtKey(v, fi, "sol")
			x, y := screenOf(v, fi, sol)
			v.Click(int(x+d[0]), int(y+d[1]), tw, th)
			if v.sel.key != "sol" {
				t.Errorf("frame %d: a click %.0f,%.0f art px off Sol picked %q", fi, d[0], d[1], v.sel.key)
			}
		}
	}
}

func objAtKey(v *View, fi int, key string) *world.Object { return v.objAt(fi, key) }
