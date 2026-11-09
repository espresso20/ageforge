package window

import "testing"

func TestProceduralDecision(t *testing.T) {
	for _, r := range []rune{'─', '│', '┌', '╔', '╭', '▀', '▄', '█', '░', '▒', '▓', '▌', '▘'} {
		if sh, ok := procedural(r, 16, 32); !ok || len(sh.tints)+len(sh.arcs)+len(sh.diags) == 0 {
			t.Errorf("%q should be drawn as shapes", r)
		}
	}
	for _, r := range []rune{'a', 'Z', ' ', '→', '♣', 'é', '·', '▸'} {
		if _, ok := procedural(r, 16, 32); ok {
			t.Errorf("%q should go to the font", r)
		}
	}
}

func TestFullBlockFillsCell(t *testing.T) {
	sh, _ := procedural('█', 10, 20)
	if len(sh.tints) != 1 || sh.tints[0].r.Dx() != 10 || sh.tints[0].r.Dy() != 20 || sh.tints[0].cover != 1 {
		t.Errorf("full block = %+v", sh)
	}
}

func TestGridSize(t *testing.T) {
	if c, r := gridSize(1600, 960, 16, 32); c != 100 || r != 30 {
		t.Errorf("got %d x %d", c, r)
	}
	if c, r := gridSize(1600, 960, 32, 64); c != 50 || r != 15 {
		t.Errorf("bigger font: got %d x %d", c, r)
	}
}
