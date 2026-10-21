package ui

import (
	"fmt"
	"strings"
	"testing"
)

// TestTopBarDropsTheEraBadge: the status bar, the second of the two top
// bars, names the age you are in and not the era badge (the epoch's icon
// and name). The Next Age bar below keeps the next age. Checked on a real
// staged game at the common sizes.
func TestTopBarDropsTheEraBadge(t *testing.T) {
	restoreForge(t)
	for _, age := range layoutAges {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		if st.EpochName == "" || st.EpochIcon == "" {
			t.Fatalf("%s: the staged game has no era badge to check", age)
		}
		for _, sz := range [][2]int{{80, 24}, {100, 30}, {120, 40}} {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", st.AgeName, w, h)
			cells := drawnDashboard(t, d, pages, w, h)
			_, sy, _, _ := d.statusTV.GetInnerRect()
			top := strings.Join(rectText(cells, w, 0, 0, w, sy+1), "\n")
			status := rectText(cells, w, 0, sy, w, 1)[0]
			if !strings.Contains(status, st.AgeName) {
				t.Errorf("%s: the status bar does not show the age %q: %q", where, st.AgeName, status)
			}
			if strings.Contains(top, st.EpochName) || strings.Contains(top, st.EpochIcon) {
				t.Errorf("%s: the top bars still show the era badge %q %q:\n%s", where, st.EpochIcon, st.EpochName, top)
			}
		}
	}
}
