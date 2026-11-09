package mapstyle

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

type fakeStage struct {
	placed  []Placement
	cleared int
}

func (s *fakeStage) Place(_ string, p Placement) { s.placed = append(s.placed, p) }
func (s *fakeStage) Clear(string)                { s.cleared++ }

func cosmosFrame(t *testing.T, st Stage) (*view, Frame) {
	t.Helper()
	fixture.WithTheme(t, "forge")
	f := Frame{Model: modelFor(t, fixture.Options{Age: "galactic_age", Seed: 7}), Stage: st, Slot: "map"}
	v := New().(*view)
	v.scene(f)
	v.zoomTo(v.scene(f), zCosmos)
	return v, f
}

// TestGalacticVoidIsPlacedOnTheStage: the star-age cosmos goes to the stage with
// transparent cells under it; the inner zoom stays cells and clears the slot.
func TestGalacticVoidIsPlacedOnTheStage(t *testing.T) {
	st := &fakeStage{}
	v, f := cosmosFrame(t, st)
	scr := drawFrame(v, f, 120, 40, false)
	defer scr.Fini()
	if len(st.placed) != 1 || st.placed[0].Build == nil || st.placed[0].Cells.Empty() {
		t.Fatalf("cosmos placed %d images, want one with a build", len(st.placed))
	}
	if img := st.placed[0].Build(8, 16); img == nil || img.Bounds().Dx() == 0 {
		t.Fatal("the build made no image")
	}
	r := st.placed[0].Cells
	_, _, sty, _ := scr.GetContent(r.Min.X+r.Dx()/2, r.Min.Y+r.Dy()/2)
	if _, bg, _ := sty.Decompose(); bg != tcell.Color(theme.Transparent) {
		t.Errorf("a cell under the image has ground %v, want transparent", bg)
	}
	v.zoomTo(v.scene(f), zInner)
	st.placed = nil
	drawFrame(v, f, 120, 40, false).Fini()
	if len(st.placed) != 0 || st.cleared == 0 {
		t.Errorf("inner zoom placed %d images and cleared %d times, want 0 and some", len(st.placed), st.cleared)
	}
}

// TestNilStageDrawsCells: terminal mode (no stage) draws the cosmos as cells
// of its own: stars and no transparent ground anywhere.
func TestNilStageDrawsCells(t *testing.T) {
	v, f := cosmosFrame(t, nil)
	scr := drawFrame(v, f, 120, 40, false)
	defer scr.Fini()
	w, h := scr.Size()
	inked := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, sty, _ := scr.GetContent(x, y)
			if _, bg, _ := sty.Decompose(); bg == tcell.Color(theme.Transparent) {
				t.Fatalf("cell %d,%d has a transparent ground with no stage", x, y)
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
