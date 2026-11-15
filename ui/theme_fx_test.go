package ui

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// An effect keeps off the cells a picture covers: they do not read as blank.
func TestThemeEffectsLeavePicturesAlone(t *testing.T) {
	b := kit.NewBuffer(20, 6)
	p := kit.NewPainter(b)
	p.Picture(image.Rect(4, 1, 11, 4), image.NewRGBA(image.Rect(0, 0, 70, 63)))
	cells := kitCells{p: p}
	if r, _, _, _ := cells.GetContent(5, 2); r == ' ' {
		t.Error("a cell under a picture reads as blank")
	}
	if r, _, _, _ := cells.GetContent(1, 2); r != ' ' {
		t.Errorf("a bare cell reads as %q", r)
	}
	bg := theme.Of(theme.RoleBackground)
	if fxEmpty(cells, 7, 2, 0, 20, bg) || fxEmpty(cells, 13, 2, 0, 20, bg) {
		t.Error("an effect finds room within its margin of a picture")
	}
}
