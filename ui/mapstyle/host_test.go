package mapstyle

import (
	"image"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/world"
)

type fakeHost struct {
	shown  []image.Rectangle
	worlds []*world.World
	hidden int
	cmds   []MapCmd
	report func(MapState)
}

func (h *fakeHost) Show(r image.Rectangle, w *world.World, _ uint64, _ world.Palette) {
	h.shown = append(h.shown, r)
	h.worlds = append(h.worlds, w)
}
func (h *fakeHost) Watch(_ func(string) string, fn func(MapState)) { h.report = fn }
func (h *fakeHost) Hide()                                          { h.hidden++ }
func (h *fakeHost) Command(c MapCmd)                               { h.cmds = append(h.cmds, c) }

func hostFrame(t *testing.T, age string, h SpaceHost) (*view, Frame) {
	t.Helper()
	fixture.WithTheme(t, "forge")
	f := Frame{Model: modelFor(t, fixture.Options{Age: age, Seed: 7}), Space: h}
	v := New().(*view)
	v.scene(f)
	return v, f
}

// TestStarAgeSkyGoesToTheHost: a star age hands its world to the host once per
// draw, the same world each time, with transparent cells under it.
func TestStarAgeSkyGoesToTheHost(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	scr := drawFrame(v, f, 120, 40, false)
	defer scr.Fini()
	if len(h.shown) != 1 || h.worlds[0] == nil || h.worlds[0].Root == nil {
		t.Fatalf("Show called %d times, want once with a world", len(h.shown))
	}
	drawFrame(v, f, 120, 40, false).Fini()
	if len(h.shown) != 2 || h.worlds[0] != h.worlds[1] {
		t.Error("a second draw should show the same world pointer")
	}
	r := h.shown[0]
	_, _, sty, _ := scr.GetContent(r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2)
	if _, bg, _ := sty.Decompose(); bg != tcell.Color(theme.Transparent) {
		t.Errorf("a cell under the sky has ground %v, want transparent", bg)
	}
}

// TestMapKeysBecomeCommands: while the sky is hosted the map keys go to the host.
func TestMapKeysBecomeCommands(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "galactic_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	for _, ev := range []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModShift),
		tcell.NewEventKey(tcell.KeyPgDn, 0, 0),
		tcell.NewEventKey(tcell.KeyHome, 0, 0),
	} {
		if !v.HandleKey(ev, f) {
			t.Errorf("key %v not handled", ev.Key())
		}
	}
	want := []MapCmd{{Op: MapPanLeft, Fast: true}, {Op: MapZoomIn}, {Op: MapHome}}
	if len(h.cmds) != len(want) {
		t.Fatalf("got %d commands, want %d", len(h.cmds), len(want))
	}
	for i, c := range want {
		if h.cmds[i] != c {
			t.Errorf("command %d = %+v, want %+v", i, h.cmds[i], c)
		}
	}
}

// TestGroundAgeHidesTheHost: an age with no sky never shows, and draws cells.
func TestGroundAgeHidesTheHost(t *testing.T) {
	h := &fakeHost{}
	v, f := hostFrame(t, "stone_age", h)
	drawFrame(v, f, 120, 40, false).Fini()
	if len(h.shown) != 0 {
		t.Errorf("a ground age showed the sky %d times", len(h.shown))
	}
}

// TestNilHostDrawsCells: terminal mode draws the cosmos as cells of its own.
func TestNilHostDrawsCells(t *testing.T) {
	v, f := hostFrame(t, "galactic_age", nil)
	scr := drawFrame(v, f, 120, 40, false)
	defer scr.Fini()
	w, h := scr.Size()
	inked := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, sty, _ := scr.GetContent(x, y)
			if _, bg, _ := sty.Decompose(); bg == tcell.Color(theme.Transparent) {
				t.Fatalf("cell %d,%d has a transparent ground with no host", x, y)
			}
			if r != ' ' {
				inked++
			}
		}
	}
	if inked == 0 {
		t.Error("the cosmos drew no glyphs")
	}
}
