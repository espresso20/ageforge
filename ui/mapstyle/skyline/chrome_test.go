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
	const where = "information age district" // the longest district name
	for w := 80; w <= 200; w++ {
		parts := panoramaHints(w, where)
		line := segsText(parts)
		if n := segsLen(parts); n > w {
			t.Errorf("at %d columns the status line is %d wide: %q", w, n, line)
		}
		if !strings.HasSuffix(line, "│ "+where) {
			t.Errorf("at %d columns the status line does not end with where the view is: %q", w, line)
		}
		if strings.Contains(line, "flows flows") {
			t.Errorf("at %d columns the status line repeats itself: %q", w, line)
		}
	}
	if line := segsText(panoramaHints(100, "bronze age district")); !strings.Contains(line, "map flows") || !strings.Contains(line, "Home End") {
		t.Errorf("100 columns has room for every key: %q", line)
	}
	if line := segsText(panoramaHints(140, "bronze age district")); !strings.Contains(line, "map flows overlay") {
		t.Errorf("a wide terminal says what map flows is: %q", line)
	}
}
