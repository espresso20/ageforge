package skyline

import (
	"strings"
	"testing"
)

func segsText(parts []seg) string {
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(p.s)
	}
	return b.String()
}

// TestPanoramaHintsFit: the skyline's status line names its keys as fully
// as the width allows and never clips where the view is, from the game's
// narrowest terminal up. It used to say "map flows flows" and cut the place
// off at 100 columns.
func TestPanoramaHintsFit(t *testing.T) {
	const where = "zoom: close  information age district" // the longest the place can be
	for _, mode := range []hintMode{hintScroll, hintPan, hintFit} {
		for w := 80; w <= 200; w++ {
			parts := panoramaHints(w, where, mode)
			line := segsText(parts)
			if n := segsLen(parts); n > w {
				t.Errorf("mode %d at %d columns the status line is %d wide: %q", mode, w, n, line)
			}
			if !strings.HasSuffix(line, "│ "+where) {
				t.Errorf("mode %d at %d columns the status line does not end with where the view is: %q", mode, w, line)
			}
			if strings.Contains(line, "flows flows") {
				t.Errorf("mode %d at %d columns the status line repeats itself: %q", mode, w, line)
			}
			// the arrows' hint says what they do, and the far level, where they do nothing, does not name them
			switch {
			case mode == hintPan && w >= 100 && !strings.Contains(line, "◄► ↑↓ pan"):
				t.Errorf("at %d columns the close level's line does not name the pan keys: %q", w, line)
			case mode == hintFit && strings.Contains(line, "◄►"):
				t.Errorf("at %d columns the far level's line names arrows that do nothing: %q", w, line)
			case mode == hintScroll && strings.Contains(line, " pan"):
				t.Errorf("at %d columns the scroll line names a pan: %q", w, line)
			}
		}
	}
	if line := segsText(panoramaHints(120, "zoom: mid  bronze age district", hintScroll)); !strings.Contains(line, "map flows") || !strings.Contains(line, "Home End") {
		t.Errorf("120 columns has room for every key: %q", line)
	}
	if line := segsText(panoramaHints(140, "zoom: mid  bronze age district", hintScroll)); !strings.Contains(line, "map flows overlay") {
		t.Errorf("a wide terminal says what map flows is: %q", line)
	}
	for _, w := range []int{80, 100, 120} { // zoom and the district jump are on the line wherever they fit
		line := segsText(panoramaHints(w, "zoom: mid  bronze age district", hintScroll))
		if !strings.Contains(line, "+ - zoom") {
			t.Errorf("%d columns: the status line does not name the zoom keys: %q", w, line)
		}
	}
	if line := segsText(panoramaHints(100, "zoom: mid  bronze age district", hintScroll)); !strings.Contains(line, "[ ] district") {
		t.Errorf("100 columns has room to name the district keys: %q", line)
	}
}
