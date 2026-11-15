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

// TestBrailleDotsAreDrawn: the window draws the braille dots itself (no font
// it ships has them), each dot its own eighth of the cell, and the eight
// together fill it. The blank one draws nothing.
func TestBrailleDotsAreDrawn(t *testing.T) {
	if sh, ok := procedural(0x2800, 10, 21); !ok || len(sh.tints) != 0 {
		t.Errorf("the blank braille cell = %+v, %v; want a shape with nothing in it", sh, ok)
	}
	// Dot 1 is the top left, dot 8 the bottom right.
	for r, want := range map[rune][4]int{0x2801: {0, 0, 5, 5}, 0x2880: {5, 15, 10, 21}} {
		sh, _ := procedural(r, 10, 21)
		if len(sh.tints) != 1 || sh.tints[0].r != rect(want[0], want[1], want[2], want[3]) || sh.tints[0].cover != 1 {
			t.Errorf("U+%04X = %+v, want one solid dot at %v", r, sh.tints, want)
		}
	}
	sh, _ := procedural(0x28FF, 10, 21)
	area := 0
	for _, d := range sh.tints {
		area += d.r.Dx() * d.r.Dy()
	}
	if len(sh.tints) != 8 || area != 10*21 {
		t.Errorf("all eight dots cover %d pixels in %d parts, want the whole 10 by 21 cell in 8", area, len(sh.tints))
	}
	if !isProcedural(0x2800) || !isProcedural(0x28FF) || isProcedural(0x2900) {
		t.Error("the braille block is U+2800 to U+28FF, and the window draws all of it")
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
