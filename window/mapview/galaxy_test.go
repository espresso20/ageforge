package mapview

import (
	"crypto/sha256"
	"image"
	"math"
	"testing"

	"github.com/espresso20/ageforge/world"
)

func findObj(t *testing.T, f *world.Frame, key string) *world.Object {
	t.Helper()
	for i := range f.Objects {
		if f.Objects[i].Key == key {
			return &f.Objects[i]
		}
	}
	t.Fatalf("frame %s has no %q", f.Key, key)
	return nil
}

func texHash(g *galTex) [32]byte {
	h := sha256.New()
	for _, m := range g.mips {
		h.Write(m)
	}
	var s [32]byte
	copy(s[:], h.Sum(nil))
	return s
}

func TestTextureIsDeterministic(t *testing.T) {
	w := build(t, "galactic_age")
	o := findObj(t, w.Root, "milky_way")
	k, _ := texKeyOf(o)
	a, b := buildTex(k), buildTex(k)
	if texHash(a) != texHash(b) {
		t.Error("the same key built two different textures")
	}
	k.seed++
	if texHash(buildTex(k)) == texHash(a) {
		t.Error("a different seed built the same texture")
	}
}

func TestCameraMovesDoNotRebuildTextures(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	n := v.tex.builds.Load()
	if n == 0 {
		t.Fatal("Set built no textures")
	}
	lo, hi := v.ch.zRange()
	for k := 0; k <= 12; k++ {
		v.z = hi - (hi-lo)*float64(k)/12
		v.tz = v.z
		v.off.X += 3
		v.Update(0)
		v.dirty = true
		v.Render(300, 170)
	}
	if got := v.tex.builds.Load(); got != n {
		t.Errorf("camera moves rebuilt textures: %d builds, was %d", got, n)
	}
}

// lit reads the texture's light under sol, for a galaxy drawn at pos.
func lit(o *world.Object, sol world.Vec, g *galTex) float64 {
	dx, dy := sol.X-o.Pos.X, sol.Y-o.Pos.Y
	u, v := geomOf(o).face(dx, dy, o.Size)
	best := 0.0
	for _, d := range [][2]float64{{0, 0}, {.012, 0}, {-.012, 0}, {0, .012}, {0, -.012}} {
		best = math.Max(best, g.at(0, u+d[0], v+d[1])[chL])
	}
	return best
}

func TestSolSitsOnALitArmAtBothBands(t *testing.T) {
	{
		v := New()
		v.Set(build(t, "galactic_age"), 1, testPal)
		for l, key := range []string{"milky_way", "disc"} {
			o := findObj(t, v.ch[l].f, key)
			sol := findObj(t, v.ch[l].f, "sol")
			if got := lit(o, sol.Pos, v.tex.get(o)); got < .25 {
				t.Errorf("band %d: Sol is on a dark pixel (light %.2f)", l, got)
			}
		}
	}
}

func TestDiscIsCentredAtTheTopOfTheGalaxyBand(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	v.home()
	v.z, v.tz = v.ch[1].lo, v.ch[1].lo
	v.Update(0)
	disc := findObj(t, v.ch[1].f, "disc")
	d := v.ch.centreIn(1, v.m, disc.Pos)
	const w = 300
	px := math.Hypot(d.X-v.off.X, d.Y-v.off.Y) * v.ch[v.m].mx / v.width() * w
	if px > 2 {
		t.Errorf("disc centre is %.1f art px from the view centre, want within 2", px)
	}
}

func TestLabelsKeepTwoArtPixelsApart(t *testing.T) {
	f := boxFace{}
	// "b" would sit one pixel clear of "a" without the pad
	a := label{x: 10, y: 50, r: 2, text: "aaaa", mask: f.Mask("aaaa"), prio: 90}
	b := label{x: 10, y: 50 + 9 + 1, r: 2, text: "bbbb", mask: f.Mask("bbbb"), prio: 10}
	shown, at := place([]label{a, b}, tw, th)
	for i := range shown {
		for j := i + 1; j < len(shown); j++ {
			ra := image.Rectangle{Min: at[i], Max: at[i].Add(shown[i].mask.Bounds().Size())}
			rb := image.Rectangle{Min: at[j], Max: at[j].Add(shown[j].mask.Bounds().Size())}
			if ra.Inset(-2).Overlaps(rb) {
				t.Errorf("%s and %s are closer than 2 art px", shown[i].text, shown[j].text)
			}
		}
	}
}
