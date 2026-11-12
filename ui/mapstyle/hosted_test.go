package mapstyle

import (
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/world"
)

// objectNamed finds an object of the hosted world by frame kind and key.
func objectNamed(t *testing.T, w *world.World, kind world.FrameKind, key string) *world.Object {
	t.Helper()
	var found *world.Object
	var walk func(f *world.Frame)
	walk = func(f *world.Frame) {
		if f.Kind == kind {
			for i := range f.Objects {
				if f.Objects[i].Key == key {
					found = &f.Objects[i]
				}
			}
		}
		for _, c := range f.Children {
			walk(c)
		}
	}
	walk(w.Root)
	if found == nil {
		t.Fatalf("no %q in a %v frame", key, kind)
	}
	return found
}

// TestHostStateDrivesTheScaleAndTheInspector: what the host reports is what the
// scale label and the inspect lines say, with no cell cursor involved.
func TestHostStateDrivesTheScaleAndTheInspector(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	if h.report == nil {
		t.Fatal("the view never asked the host to report")
	}
	if _, ok := v.Inspect(f); ok {
		t.Error("the inspector spoke with nothing selected")
	}
	home := v.zoomLabel()

	h.report(MapState{Frame: world.FrameLocalGroup})
	cosmos := v.zoomLabel()
	h.report(MapState{Frame: world.FrameGalaxy})
	overview := v.zoomLabel()
	h.report(MapState{Frame: world.FrameSystem})
	closeup := v.zoomLabel()
	if cosmos == overview || overview == closeup || cosmos == closeup {
		t.Errorf("the labels do not tell the scales apart: %q %q %q (was %q)", cosmos, overview, closeup, home)
	}

	venus := objectNamed(t, v.sky, world.FrameSystem, "venus")
	h.report(MapState{Frame: world.FrameSystem, SelFrame: world.FrameSystem, Sel: venus})
	in, ok := v.Inspect(f)
	if !ok || !strings.EqualFold(in.Title, SkyName("venus")) || in.Title == "" {
		t.Errorf("inspector says %+v (ok %v), want the planet's name %q", in, ok, SkyName("venus"))
	}

	andromeda := objectNamed(t, v.sky, world.FrameLocalGroup, "andromeda")
	h.report(MapState{Frame: world.FrameLocalGroup, SelFrame: world.FrameLocalGroup, Sel: andromeda})
	in, _ = v.Inspect(f)
	if in.Title != SkyName("andromeda") || len(in.Lines) == 0 {
		t.Errorf("inspector says %+v for Andromeda, want its name and a distance line", in)
	}
}

func TestSkyNameKnowsTheMapsObjects(t *testing.T) {
	for _, k := range []string{"sol", "milky_way", "andromeda", "alpha_centauri", "mars", "the_moon", "orion_nebula"} {
		if SkyName(k) == "" {
			t.Errorf("no name for %q", k)
		}
	}
	if SkyName("far_3") != "" {
		t.Error("an unnamed backdrop galaxy got a name")
	}
}

// The star map's scale label follows the band the window's camera is in: the
// Region reads as the overview, only the Neighbourhood as INNER, and the
// system and the home world as the close-up.
func TestScaleLabelFollowsTheBand(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	label := func(k world.FrameKind) string {
		h.report(MapState{Frame: k})
		return v.zoomLabel()
	}
	if label(world.FrameRegion) != label(world.FrameGalaxy) {
		t.Errorf("Region reads %q, want the Galaxy's %q", label(world.FrameRegion), label(world.FrameGalaxy))
	}
	inner := label(world.FrameNeighbourhood)
	for _, k := range []world.FrameKind{world.FrameLocalGroup, world.FrameGalaxy, world.FrameRegion, world.FrameSystem, world.FramePlanet} {
		if label(k) == inner {
			t.Errorf("frame %v reads %q, which belongs to the Neighbourhood alone", k, inner)
		}
	}
	if label(world.FrameSystem) != label(world.FramePlanet) {
		t.Errorf("System %q and Planet %q should share the close-up label", label(world.FrameSystem), label(world.FramePlanet))
	}
}

// A selected object of a locked frame reads as its name plus when it opens;
// the home world reads as itself, and the station has a name.
func TestLockedAndHomeInspectLines(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	star := objectNamed(t, v.sky, world.FrameNeighbourhood, "alpha_centauri")
	h.report(MapState{Frame: world.FrameNeighbourhood, SelFrame: world.FrameNeighbourhood, Sel: star, Lock: "interstellar_age"})
	in, ok := v.Inspect(f)
	if !ok || in.Title == "" || len(in.Lines) != 1 || !strings.Contains(in.Lines[0], "Interstellar") {
		t.Errorf("locked star inspects as %+v", in)
	}
	earth := objectNamed(t, v.sky, world.FrameSystem, "earth")
	h.report(MapState{Frame: world.FrameSystem, SelFrame: world.FrameSystem, Sel: earth})
	if in, _ = v.Inspect(f); strings.Contains(in.Title, "Velgate") || in.Title == "" {
		t.Errorf("Earth inspects as %q", in.Title)
	}
	if SkyName("space_station") == "" || SkyName("europa") == "" || SkyName("pluto") == "" {
		t.Error("a station, a moon or Pluto has no name")
	}
}

// A made-up star keeps its id in its key; the player reads the name alone, for
// the star, its planets and every star of the Region, with no two alike.
func TestMadeUpNamesShowNoIDs(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	digits := regexp.MustCompile(`\d{4,}`)
	region := v.sky.Root.Find(world.FrameRegion)
	seen := map[string]bool{}
	for _, o := range region.Objects {
		if !strings.HasPrefix(o.Key, "~") || o.Kind != world.KindStar {
			continue
		}
		n := SkyName(o.Label)
		if n == "" || digits.MatchString(n) || seen[n] {
			t.Errorf("star %q reads %q (empty, with an id, or a repeat)", o.Key, n)
		}
		seen[n] = true
		for _, p := range v.sky.System(o.Key).Objects {
			if n := SkyName(p.Label); digits.MatchString(n) {
				t.Errorf("%q of %q reads %q", p.Key, o.Key, n)
			}
		}
	}
	if len(seen) == 0 {
		t.Fatal("the Region holds no made-up stars")
	}
}
