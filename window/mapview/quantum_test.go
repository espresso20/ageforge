package mapview

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/world"
)

func quantum(t *testing.T) *View {
	t.Helper()
	v := New()
	v.Set(build(t, world.AgeQuantum), 1, testPal)
	v.bod.wg.Wait()
	if v.open != 1 || v.ch[1].f.Kind != world.FrameLocalGroup {
		t.Fatalf("open %d, first frame %v: the Local Group should be open under the locked web", v.open, v.ch[0].f.Kind)
	}
	return v
}

// dive double-clicks object o of chain frame fi from a zoom inside it.
func dive(t *testing.T, v *View, fi int, o *world.Object) {
	t.Helper()
	z, ok := zFor(v, fi)
	if !ok {
		t.Fatalf("no zoom inside frame %d", fi)
	}
	at(v, z, fi, o.Pos)
	x, y := screenOf(v, fi, o)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
}

func TestQuantumGalaxyDivesToSystemThenPlanet(t *testing.T) {
	v := quantum(t)
	and := v.objAt(1, "andromeda")
	dive(t, v, 1, and)
	if v.gal != "andromeda" || v.ch[2].f.Anchor != "andromeda" {
		t.Fatalf("gal %q, frame 1 %q: Andromeda was not entered", v.gal, v.ch[2].f.Anchor)
	}
	var star *world.Object
	for i := range v.ch[2].f.Objects {
		if o := &v.ch[2].f.Objects[i]; o.Kind == world.KindStar && o.Key != v.ch[3].f.Anchor { // not the one already hung
			star = o
			break
		}
	}
	if star == nil || star.Doubt == 0 {
		t.Fatalf("Andromeda holds no uncertain star: %v", star)
	}
	dive(t, v, 2, star)
	ri, si := v.ch.idx(world.FrameRegion), v.ch.idx(world.FrameSystem)
	if v.place != star.Key || v.ch[ri].f.Anchor != star.Key || v.ch[si].f.Anchor != star.Key {
		t.Fatalf("place %q, region %q, system %q: no dive into %s", v.place, v.ch[ri].f.Anchor, v.ch[si].f.Anchor, star.Key)
	}
	var planet *world.Object
	for i := range v.ch[si].f.Objects {
		if o := &v.ch[si].f.Objects[i]; o.Kind == world.KindPlanet {
			planet = o
		}
	}
	if planet == nil {
		t.Fatal("the entered system holds no planet")
	}
	at(v, v.ch[si].bmin+(v.ch[si].bmax-v.ch[si].bmin)*.5, si, world.Vec{})
	v.swapTail(planet.Key)
	if tail := v.ch[len(v.ch)-1].f; tail.Kind != world.FramePlanet || tail.Anchor != planet.Key {
		t.Fatalf("tail %v %q, want the planet %s", tail.Kind, tail.Anchor, planet.Key)
	}
	v.Set(build(t, world.AgeQuantum), 1, testPal) // a rebuild keeps the galaxy the player is in
	if v.gal != "andromeda" || v.ch[2].f.Anchor != "andromeda" {
		t.Errorf("a rebuild left gal %q", v.gal)
	}
}

func TestQuantumGalaxiesKeepUniqueKeys(t *testing.T) {
	w := build(t, world.AgeQuantum)
	seen := map[string]string{}
	n := 0
	for _, lg := range w.Root.Find(world.FrameLocalGroup).Objects {
		if lg.Doubt == 0 {
			continue
		}
		n++
		g := lg.Key
		f := w.Galaxy(g)
		if f == nil {
			t.Fatalf("no frame for %s", g)
		}
		for _, o := range f.Objects {
			if o.Kind != world.KindStar {
				continue
			}
			if at, dup := seen[o.Key]; dup {
				t.Fatalf("star %s is in both %s and %s", o.Key, at, g)
			}
			seen[o.Key] = g
		}
	}
	if n < 3 {
		t.Errorf("only %d galaxies can be entered", n)
	}
}

func TestUncertainPlaceDrawsDifferentlyOnTwoTicks(t *testing.T) {
	v := quantum(t)
	and := v.objAt(1, "andromeda")
	z, _ := zFor(v, 1)
	at(v, z, 1, and.Pos)
	shot := func() []byte {
		v.dirty = true
		return append([]byte(nil), v.Render(tw, th).Pix...)
	}
	a := shot()
	v.Update(1100 * time.Millisecond) // one swap later
	b := shot()
	if string(a) == string(b) {
		t.Fatal("an uncertain galaxy drew the same on two clock ticks")
	}
	v.Update(1100 * time.Millisecond)
	v.Update(1100 * time.Millisecond)
	if len(a) == 0 || len(v.pending) != 0 {
		t.Fatal("looking settled something")
	}
}

func TestSettledPlaceStopsFlickering(t *testing.T) {
	v := quantum(t)
	and := v.objAt(1, "andromeda")
	z, _ := zFor(v, 1)
	at(v, z, 1, and.Pos)
	x, y := screenOf(v, 1, and)
	v.Click(int(math.Round(x)), int(math.Round(y)), tw, th)
	if got := v.TakeSettled(); len(got) != 1 || got[0] != v.sel.key {
		t.Fatalf("a click settled %v, selected %q", got, v.sel.key)
	}
	picked := v.objAt(v.sel.fi, v.sel.key)
	if !v.isSettled(picked) {
		t.Fatalf("%s is not settled after the click", picked.Key)
	}
	shot := func() []byte {
		v.dirty = true
		return append([]byte(nil), v.Render(tw, th).Pix...)
	}
	a := shot()
	v.Update(1100 * time.Millisecond)
	r := image.Rect(int(x)-4, int(y)-4, int(x)+5, int(y)+5)
	b := shot()
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			i := (py*tw + px) * 4
			if string(a[i:i+4]) != string(b[i:i+4]) {
				t.Fatalf("a settled place still changed at %d,%d", px, py)
			}
		}
	}
	if got := v.TakeSettled(); len(got) != 0 {
		t.Errorf("settling twice reported %v", got)
	}
}

func TestSettledSurvivesTheSaveLoad(t *testing.T) {
	w := build(t, world.AgeQuantum)
	lg := w.Root.Find(world.FrameLocalGroup)
	o := &lg.Objects[0]
	for i := range lg.Objects {
		if lg.Objects[i].Key == "andromeda" {
			o = &lg.Objects[i]
		}
	}
	v := New()
	v.Set(w, 1, testPal)
	if v.isSettled(o) {
		t.Fatal("settled before anyone looked")
	}
	w.Settled = map[string]bool{"andromeda": true} // what a loaded save hands the model
	if !v.isSettled(o) {
		t.Fatal("a place in the save is not settled")
	}
}

func TestLocalGroupStaysLockedAtGalactic(t *testing.T) {
	w := build(t, world.AgeGalactic)
	for _, o := range w.Root.Find(world.FrameLocalGroup).Objects {
		if o.Doubt > 0 || (o.Kind == world.KindGalaxy && o.Key != "milky_way" && o.Pick && o.Shape != nil) {
			t.Errorf("%s is open at Galactic", o.Key)
		}
	}
	if w.Galaxy("andromeda") != nil {
		t.Error("Andromeda has a frame at Galactic")
	}
}
