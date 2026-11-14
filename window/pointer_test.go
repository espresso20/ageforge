package window

import "testing"

func TestCellAtFindsTheCellUnderAPixel(t *testing.T) {
	// A 10 by 4 grid of 8x16 cells, centred in a window 6 pixels wider and
	// 4 taller than the grid.
	g := &Game{cw: 8, ch: 16, cols: 10, rows: 4, pw: 86, ph: 68}
	for _, c := range []struct {
		px, py, x, y int
		in           bool
	}{
		{3, 2, 0, 0, true},
		{10, 17, 0, 0, true},
		{11, 18, 1, 1, true},
		{82, 65, 9, 3, true},
		{2, 10, -1, 0, false}, // the margin left of the grid
		{83, 10, 10, 0, false},
		{40, 66, 4, 4, false},
	} {
		x, y, in := g.cellAt(c.px, c.py)
		if x != c.x || y != c.y || in != c.in {
			t.Errorf("pixel %d,%d is cell %d,%d (in %v), want %d,%d (in %v)", c.px, c.py, x, y, in, c.x, c.y, c.in)
		}
	}
	if _, _, in := (&Game{}).cellAt(5, 5); in {
		t.Error("before the first layout there is no grid to be in")
	}
}
