package mapview

import (
	"bytes"
	"image"
	"math"
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

// At whole-system zoom the display squeeze keeps the eight planets apart from
// each other and from Sol, wherever each sits on its orbit.
func TestPlanetsSitApartAtWholeSystemZoom(t *testing.T) {
	v := nearView(t)
	const w, h = 384, 236 // the real window's canvas
	// the zoom at which Pluto's orbit just fits the canvas height
	ppu := .94 * h / 2 / squeeze(v.ch[4].f, world.PlutoAU)
	at(v, math.Log10(w/ppu*v.ch[4].mx), 4, world.Vec{})
	if v.m != 4 {
		t.Fatalf("whole-system zoom is in frame %d, not the system", v.m)
	}
	p := v.penFor(image.NewRGBA(image.Rect(0, 0, w, h)), 4, v.snapped(w))
	sx, sy := p.at(world.Vec{})
	type pt struct {
		key  string
		x, y float64
	}
	var ps []pt
	for _, o := range planets(v) {
		if o.Key == "pluto" {
			continue
		}
		x, y := p.at(o.Pos)
		if x < 0 || y < 0 || x > w || y > h {
			t.Errorf("%s is off the canvas at %.0f,%.0f", o.Key, x, y)
		}
		if d := math.Hypot(x-sx, y-sy); d < 6 {
			t.Errorf("%s is %.1f px from Sol", o.Key, d)
		}
		ps = append(ps, pt{o.Key, x, y})
	}
	if len(ps) != 8 {
		t.Fatalf("%d planets", len(ps))
	}
	for i := range ps {
		for j := i + 1; j < len(ps); j++ {
			if d := math.Hypot(ps[i].x-ps[j].x, ps[i].y-ps[j].y); d < 6 {
				t.Errorf("%s and %s are %.1f px apart", ps[i].key, ps[j].key, d)
			}
		}
	}
	// the belts and Pluto lie between their neighbours, not on them
	rb := squeeze(p.f, world.BeltAU) * p.ppu
	for _, q := range ps {
		if d := math.Abs(math.Hypot(q.x-sx, q.y-sy) - rb); d < 6 {
			t.Errorf("%s is %.1f px from the asteroid belt", q.key, d)
		}
	}
}

// Picking, panning and drawing share one mapping: the squeeze inverts.
func TestSystemSqueezeInverts(t *testing.T) {
	f := &world.Frame{Kind: world.FrameSystem}
	for _, r := range []float64{.01, .39, 1, 5.2, 39.5, 140} {
		if got := unsqueeze(f, squeeze(f, r)); math.Abs(got-r) > 1e-9*r {
			t.Errorf("unsqueeze(squeeze(%v)) = %v", r, got)
		}
	}
	pt := world.Vec{X: 3, Y: -4}
	if got := undisp(f, disp(f, pt)); math.Hypot(got.X-3, got.Y+4) > 1e-9 {
		t.Errorf("undisp(disp) = %v", got)
	}
}

// The selection ring clears everything drawn of the object, Saturn's rings
// included, by two art pixels.
func TestSelectionRingClearsTheDrawnObject(t *testing.T) {
	for _, z := range []float64{sysZ, 12.6, 12} {
		v := sysView(t)
		at(v, z, 4, world.Vec{})
		img := v.Render(tw, th)
		p := v.penFor(img, 4, v.snapped(tw))
		for _, o := range planets(v) {
			if o.Body == nil || o.Body.Spec.Size == 0 && o.Key != "pluto" {
				continue
			}
			extent := 0
			if o.Body.Spec.Size != 0 {
				extent = v.bod.get(&v.pal, o, stepR(p.radius(o)), 0).ext
			}
			if got := ringR(p, o); got < float64(extent)+2 {
				t.Errorf("z %.1f %s: ring radius %.1f, drawn extent %d", z, o.Key, got, extent)
			}
		}
	}
}

// Every name that shows sits within labelGap of its own object's box, in
// every frame, at every zoom tried; the selected one always shows.
func TestLabelsStayBesideTheirObjects(t *testing.T) {
	for fi := 0; fi < 6; fi++ {
		v := nearView(t)
		if fi >= len(v.ch) {
			continue
		}
		v.SetFace(boxFace{})
		v.SetNames(func(k string) string { return k })
		lo, hi := v.ch[fi].lo, v.ch[fi].hi
		if fi == 0 {
			lo, hi = v.ch[0].bmin, v.ch[0].bmax
		}
		for k := 0; k <= 6; k++ {
			z := lo + (hi-lo)*float64(k)/6
			at(v, z, fi, centreOf(v, fi))
			img := v.Render(tw, th)
			shown, pts := v.choose(img, v.snapped(tw))
			for i, l := range shown {
				own := l.box
				if own.Empty() {
					own = boxOf(l.x, l.y, math.Max(l.r, 1))
				}
				r := image.Rectangle{Min: pts[i], Max: pts[i].Add(l.mask.Bounds().Size())}
				gap := max(own.Min.X-r.Max.X, r.Min.X-own.Max.X, own.Min.Y-r.Max.Y, r.Min.Y-own.Max.Y)
				if gap > labelGap && !l.chosen {
					t.Errorf("frame %d z %.2f: %q is %d art px from its object", fi, z, l.key, gap)
				}
			}
		}
	}
}

// The selected object's name shows even when crowded.
func TestSelectedNameAlwaysShows(t *testing.T) {
	v := sysView(t)
	v.SetFace(boxFace{})
	v.SetNames(func(k string) string { return k })
	v.sel = selection{fi: 4, key: "venus"}
	at(v, 14.1, 4, centreOf(v, 4)) // the whole system, planets crowding the canvas
	img := v.Render(tw, th)
	shown, _ := v.choose(img, v.snapped(tw))
	for _, l := range shown {
		if l.key == "venus" && l.chosen {
			return
		}
	}
	t.Error("the selected planet's name is missing")
}
