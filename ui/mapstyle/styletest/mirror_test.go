package styletest

import "testing"

func TestMirrorRune(t *testing.T) {
	for _, c := range []struct {
		r, lr, tb rune
	}{
		{'╱', '╲', '╲'},
		{'◄', '►', '◄'},
		{'◤', '◥', '◣'},
		{'▀', '▀', '▄'},
		{'▲', '▲', '▼'},
		{'●', '●', '●'},
	} {
		if got := MirrorRune(c.r, true); got != c.lr {
			t.Errorf("%q across the vertical axis: %q, want %q", c.r, got, c.lr)
		}
		if got := MirrorRune(c.r, false); got != c.tb {
			t.Errorf("%q across the horizontal axis: %q, want %q", c.r, got, c.tb)
		}
	}
	for _, table := range []map[rune]rune{mirrorLR, mirrorTB} {
		for a, b := range table {
			if table[b] != a {
				t.Errorf("%q mirrors to %q, which mirrors to %q", a, b, table[b])
			}
		}
	}
}
