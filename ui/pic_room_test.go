package ui

import (
	"image"
	"testing"
)

// roomCells are the cells (width, height in pixels) the window's font sizes
// make: the smooth font at 10, 11, 16, 24 and 32 points on a display of scale
// 1 and of scale 2, and the pixel font at every multiple the settings allow on
// each.
var roomCells = func() map[string][2]int {
	m := map[string][2]int{}
	for _, c := range []struct {
		name   string
		cw, ch int
	}{
		{"mono10", 6, 13}, {"mono11", 7, 15}, {"mono16", 10, 21}, {"mono24", 14, 32}, {"mono32", 19, 42},
		{"mono10x2", 12, 26}, {"mono11x2", 13, 29}, {"mono16x2", 19, 42}, {"mono24x2", 29, 63}, {"mono32x2", 38, 84},
		{"pixel1", 7, 14}, {"pixel2", 14, 28}, {"pixel3", 21, 42},
		{"pixel4x2", 28, 56}, {"pixel5x2", 35, 70}, {"pixel6x2", 42, 84},
	} {
		m[c.name] = [2]int{c.cw, c.ch}
	}
	return m
}()

// lastDiffX is the right edge of what differs between two pictures.
func lastDiffX(a, b *image.RGBA) int {
	last := -1
	for y := a.Rect.Min.Y; y < a.Rect.Max.Y; y++ {
		for x := a.Rect.Min.X; x < a.Rect.Max.X; x++ {
			if a.RGBAAt(x, y) != b.RGBAAt(x, y) {
				last = max(last, x)
			}
		}
	}
	return last + 1
}

func TestPicRoomRoundsUp(t *testing.T) {
	for _, c := range []struct{ right, gap, cw, want int }{{47, 1, 7, 7}, {48, 1, 7, 7}, {49, 1, 7, 8}, {10, 0, 10, 1}, {11, 0, 10, 2}, {5, 0, 0, 0}} {
		if got := picRoom(c.right, c.gap, c.cw); got != c.want {
			t.Errorf("picRoom(%d, %d, %d) = %d, want %d", c.right, c.gap, c.cw, got, c.want)
		}
	}
}

// A ladder step is drawn to the whole cells its layout reserves, at every cell
// size: it has no width in art pixels that could run past its cells into the
// text beside it.
func TestCellFilledPicturesStayInTheirCells(t *testing.T) {
	in := newStatsInk(newExpPalette())
	const cols, rows = 7, 5
	for name, c := range roomCells {
		cw, ch := c[0], c[1]
		w, h := cols*cw, rows*ch
		img := statsStepPic(cols, rows, cw, ch, 1, 4, true, false, false, in.face, in)
		if img.Rect.Dx() != w || img.Rect.Dy() != h {
			t.Errorf("%s step: %dx%d, want %dx%d", name, img.Rect.Dx(), img.Rect.Dy(), w, h)
		}
	}
}
