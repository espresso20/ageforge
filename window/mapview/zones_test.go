package mapview

import (
	"math"
	"testing"
	"time"

	"github.com/espresso20/ageforge/world"
)

// galStar is the first generated star of the galaxy frame the people hold (or not).
func galStar(t *testing.T, f *world.Frame, held bool) *world.Object {
	t.Helper()
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindStar && o.Key != "sol" && o.Held == held {
			return o
		}
	}
	t.Fatalf("no star with held=%v", held)
	return nil
}

func galactic(t *testing.T) (*View, int, *world.Object) {
	t.Helper()
	v := New()
	v.Set(build(t, world.AgeGalactic), 1, testPal)
	v.bod.wg.Wait()
	gi := v.ch.idx(world.FrameGalaxy)
	if gi != v.open || gi < 1 {
		t.Fatalf("galaxy at %d, open %d", gi, v.open)
	}
	if _, ok := zFor(v, gi); !ok {
		t.Fatal("no zoom inside the galaxy")
	}
	return v, gi, galStar(t, v.ch[gi].f, true)
}

func TestGalacticHeldStarDivesToSystemThenPlanet(t *testing.T) {
	v, gi, star := galactic(t)
	z, _ := zFor(v, gi)
	at(v, z, gi, star.Pos)
	x, y := screenOf(v, gi, star)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	ri, si := v.ch.idx(world.FrameRegion), v.ch.idx(world.FrameSystem)
	if v.place != star.Key || v.ch[ri].f.Anchor != star.Key || v.ch[si].f.Anchor != star.Key || v.tz >= z {
		t.Fatalf("place %q, region %q, system %q, tz %.2f from %.2f: no dive", v.place, v.ch[ri].f.Anchor, v.ch[si].f.Anchor, v.tz, z)
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
	v.Set(build(t, world.AgeGalactic), 1, testPal) // a rebuild keeps the place the player is in
}

func TestGalacticStarOutsideTheTerritoryDoesNotDive(t *testing.T) {
	v, gi, _ := galactic(t)
	out := galStar(t, v.ch[gi].f, false)
	z, _ := zFor(v, gi)
	at(v, z, gi, out.Pos)
	x, y := screenOf(v, gi, out)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	if st := v.state(); st.Sel == nil || st.Sel.Key != out.Key {
		t.Fatalf("selected %v, want %s", st.Sel, out.Key)
	}
	if v.place != "" || v.tz != z || v.focus.on {
		t.Errorf("an outside star was entered: place %q, tz %.2f -> %.2f", v.place, z, v.tz)
	}
}

func TestGalacticLocalGroupIsLockedUntilQuantum(t *testing.T) {
	v, _, _ := galactic(t)
	if v.open != 1 || v.ch[0].f.Kind != world.FrameLocalGroup {
		t.Fatalf("open %d, first frame %v: the Local Group should sit locked above", v.open, v.ch[0].f.Kind)
	}
	g := firstStar(t, v.ch[0].f)
	v.sel = selection{fi: 0, key: g.Key}
	if st := v.state(); st.Lock != world.AgeQuantum {
		t.Errorf("lock %q, want %s", st.Lock, world.AgeQuantum)
	}
}

func TestInterstellarGalaxyStaysLocked(t *testing.T) {
	if w := build(t, world.AgeInterstellar); w.Root.Find(world.FrameGalaxy) != nil {
		t.Error("the Interstellar world holds a Galaxy frame")
	}
}

func TestKeyboardZoomStopsAtAMadeUpSystem(t *testing.T) {
	v, gi, star := galactic(t)
	z, _ := zFor(v, gi)
	at(v, z, gi, star.Pos)
	x, y := screenOf(v, gi, star)
	v.Focus(int(math.Round(x)), int(math.Round(y)), tw, th)
	si := v.ch.idx(world.FrameSystem)
	at(v, v.ch[si].bmin+(v.ch[si].bmax-v.ch[si].bmin)*.5, si, world.Vec{})
	for i := 0; i < 40; i++ {
		v.Do(Cmd{Op: ZoomIn})
		v.Update(time.Second)
	}
	if v.ch.weightAt(len(v.ch)-1, v.z, v.look()) > 0 {
		t.Errorf("zoom %.2f went past the system into a planet close-up", v.z)
	}
}
