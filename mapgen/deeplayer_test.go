package mapgen

import "testing"

// A coarse window (step above 256, as the Galactic close-up of Sol uses) must
// build its dark ground once, not once per cell.
func TestDeepLayerBuildsOncePerWindow(t *testing.T) {
	for _, step := range []int{256, 512, 1000} {
		o := &Overview{Grid: newGrid(60, 20), step: step, X0: 7, Y0: 3}
		var d deepLayer
		d.cell(1, o, 0, 0)
		first := &d.cells[0]
		for y := 0; y < o.Grid.H; y++ {
			for x := 0; x < o.Grid.W; x++ {
				d.cell(1, o, x, y)
				if &d.cells[0] != first {
					t.Fatalf("step %d: the window was rebuilt at cell (%d,%d)", step, x, y)
				}
			}
		}
	}
}

// The wide layers are drawn once per window and again when the window moves.
func TestWideLayerKeyedOnTheWindow(t *testing.T) {
	var gen Generator
	s := &Scene{Seed: 3, Grid: newGrid(400, 200), wide: &gen.wide}
	s.OverviewArm(80, 30)
	first := &gen.wide.cells[0]
	s.OverviewArm(80, 30)
	if &gen.wide.cells[0] != first {
		t.Error("the same window was drawn twice")
	}
	s.OverviewVoid(80, 30)
	if &gen.wide.cells[0] == first {
		t.Error("the void reused the arm sector")
	}
}
