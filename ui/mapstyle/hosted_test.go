package mapstyle

import (
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
