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

func TestLabelsClearEveryDrawnObject(t *testing.T) {
	f := boxFace{}
	mine := boxOf(100, 50, 3)
	other := boxOf(112, 50, 8) // a ringed neighbour right of it
	l := label{x: 100, y: 50, r: 3, box: mine, text: "me", mask: f.Mask("me"), prio: 10}
	shown, at := place([]label{l}, tw, th, mine, other)
	if len(shown) != 1 {
		t.Fatal("the label found no spot")
	}
	r := image.Rectangle{Min: at[0], Max: at[0].Add(shown[0].mask.Bounds().Size())}
	if r.Inset(-labelPad).Overlaps(other) {
		t.Errorf("label %v touches a neighbour's box %v", r, other)
	}
	if r.Inset(-labelPad).Overlaps(mine) && !r.Inset(-labelGap).Overlaps(mine) {
		t.Error("label sits on its own object")
	}
	// boxed in on every side: hidden, not forced through
	if s, _ := place([]label{l}, tw, th, mine, image.Rect(60, 20, 97, 80), image.Rect(103, 20, 140, 80),
		image.Rect(60, 20, 140, 47), image.Rect(60, 53, 140, 80)); len(s) != 0 {
		t.Error("a label with no clear spot was kept")
	}
}

func TestLabelStartsNextToItsOwnObject(t *testing.T) {
	f := boxFace{}
	own := boxOf(150, 80, 6)
	l := label{x: 150, y: 80, r: 6, box: own, text: "near", mask: f.Mask("near"), prio: 10}
	// right and left are blocked, so it goes above
	right, left := image.Rect(own.Max.X, 76, own.Max.X+40, 84), image.Rect(own.Min.X-40, 76, own.Min.X, 84)
	shown, at := place([]label{l}, tw, th, own, right, left)
	if len(shown) != 1 {
		t.Fatal("no spot found")
	}
	r := image.Rectangle{Min: at[0], Max: at[0].Add(shown[0].mask.Bounds().Size())}
	if !r.Inset(-labelGap - 1).Overlaps(own) { // at most labelGap clear pixels
		t.Errorf("label %v starts more than 2 art px from its object %v", r, own)
	}
	if r.Max.Y > own.Min.Y {
		t.Errorf("expected above the object, got %v", r)
	}
}

func TestFocusOnTheLMCDivesIntoTheGalaxyFrame(t *testing.T) {
	v := New()
	v.Set(build(t, "quantum_age"), 1, testPal) // the web is locked above the Local Group
	v.home()
	lmc := findObj(t, v.ch[1].f, "large_magellanic_cloud")
	x, y := screenOf(v, 1, lmc)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	settle(v)
	near := findObj(t, v.ch[2].f, "large_magellanic_cloud")
	if v.sel.fi != 2 || v.sel.key != near.Key {
		t.Fatalf("selection is %v, want the close-up LMC", v.sel)
	}
	d := v.ch.centreIn(2, v.m, near.Pos)
	if px := math.Hypot(d.X-v.off.X, d.Y-v.off.Y) * v.ch[v.m].mx / v.width() * tw; px > 2 {
		t.Errorf("the LMC is %.1f art px from the centre", px)
	}
	if across := 2 * near.Size * v.ch[2].mx / v.width() * tw; across < 40 {
		t.Errorf("the LMC is %.0f art px across, want at least 40", across)
	}
}

func TestHomeCentresTheGalaxyHighAndSolDeep(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	for _, c := range []struct {
		z   float64
		key string
		fi  int
	}{{v.ch[1].bmax, "disc", 1}, {(v.ch[1].bmin+v.ch[1].bmax)/2 - .1, "sol", 1}, {v.ch[1].bmin, "sol", 1}} {
		v.z, v.tz = c.z, c.z
		v.Update(0)
		v.Do(Cmd{Op: Home})
		p, _ := v.ch.objPos(c.fi, c.key)
		d := v.ch.centreIn(c.fi, v.m, p)
		if px := math.Hypot(d.X-v.off.X, d.Y-v.off.Y) * v.ch[v.m].mx / v.width() * tw; px > 2 {
			t.Errorf("z=%.2f: %s is %.1f art px from the centre after Home", c.z, c.key, px)
		}
	}
}

func TestWarmCoreStaysSmallAtEveryMip(t *testing.T) {
	v := New()
	v.Set(build(t, "galactic_age"), 1, testPal)
	o := findObj(t, v.ch[0].f, "milky_way")
	warm := map[[3]uint8]bool{}
	for _, c := range coreRamp(&testPal) {
		warm[[3]uint8{c.R, c.G, c.B}] = true
	}
	gm := geomOf(o)
	for _, rpx := range []float64{100, 40, 14, 8} {
		const w, h = 260, 260
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		fill(img, testPal.Bg)
		p := pen{img: img, pal: &testPal, tex: &v.tex, w: w, h: h}
		p.galaxy(o, w/2, h/2, rpx)
		var lit, hot, arm, far int
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				c := img.RGBAAt(x, y)
				if c == testPal.Bg {
					continue
				}
				u, vv := gm.face(float64(x-w/2), float64(y-h/2), rpx)
				r := math.Hypot(u, vv)
				lit++
				if warm[[3]uint8{c.R, c.G, c.B}] {
					hot++
					if r > .3 {
						far++
					}
				} else if r >= .2 && r <= .4 {
					arm++
				}
			}
		}
		if lit == 0 || float64(hot) > .2*float64(lit) {
			t.Errorf("rpx %.0f: %d of %d lit pixels are warm, want under 20%%", rpx, hot, lit)
		}
		if far > 0 {
			t.Errorf("rpx %.0f: %d warm pixels beyond 30%% of the radius", rpx, far)
		}
		if rpx >= 14 && arm == 0 {
			t.Errorf("rpx %.0f: no lit arm pixels at 20-40%% of the radius", rpx)
		}
	}
}
