package ui

import (
	"fmt"
	"strings"
	"testing"
)

// TestTopBarDropsTheEraBadge: the header names the age you are in and not
// the era badge (the epoch's icon and name). Its bar keeps the next age.
// Checked on a real staged game at the common sizes.
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
			top := strings.Join(checkHeader(t, where, d, cells, w), "\n")
			status := rectText(cells, w, 0, 1, w, 1)[0]
			if !strings.Contains(status, st.AgeName) {
				t.Errorf("%s: the header does not show the age %q: %q", where, st.AgeName, status)
			}
			if strings.Contains(top, st.EpochName) || strings.Contains(top, st.EpochIcon) {
				t.Errorf("%s: the header still shows the era badge %q %q:\n%s", where, st.EpochIcon, st.EpochName, top)
			}
		}
	}
}
