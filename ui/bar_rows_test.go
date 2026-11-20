package ui

import "testing"

// The command bar is six rows on a window of 120x39 and up, else three; a
// panel gets the rows above it.
func TestBarRowsFollowTheWindow(t *testing.T) {
	for _, c := range []struct{ w, h, bar, panel int }{
		{279, 74, 6, 68}, {279, 71, 6, 65}, {279, 68, 6, 62},
		{157, 46, 6, 40}, {157, 43, 6, 37}, {157, 40, 6, 34},
		{120, 39, 6, 33}, {120, 38, 3, 35}, {120, 36, 3, 33},
		{119, 60, 3, 57}, {80, 24, 3, 21},
	} {
		if got := barRows(c.w, c.h); got != c.bar {
			t.Errorf("%dx%d: bar is %d rows, want %d", c.w, c.h, got, c.bar)
		}
		if got := c.h - barRows(c.w, c.h); got != c.panel {
			t.Errorf("%dx%d: panels get %d rows, want %d", c.w, c.h, got, c.panel)
		}
	}
}

// The dashboard's own input takes the size barRows gives, and the line is
// at double size only with six rows.
func TestTheCommandBoxTakesTheBarRows(t *testing.T) {
	restoreForge(t)
	d, pages, _ := stagedDashboard(t, "bronze_age")
	pages.SwitchToPage("dashboard")
	for _, sz := range [][2]int{{279, 74}, {157, 43}, {120, 36}, {100, 30}} {
		w, h := sz[0], sz[1]
		stageOf(d, pages, w, h)
		_, _, _, bh := d.inputField.GetRect()
		if want := barRows(w, h); bh != want {
			t.Errorf("%dx%d: the command box is %d rows, want %d", w, h, bh, want)
		}
		if big := d.inputField.in.Big; big != (barRows(w, h) == barBigRows) {
			t.Errorf("%dx%d: big line = %v", w, h, big)
		}
	}
}
