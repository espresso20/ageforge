package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/window/kit"
)

// rowWith is the cell of the first row inside box whose text has word, two
// cells into the word.
func rowWith(t *testing.T, st *stage, box Primitive, word string) image.Point {
	t.Helper()
	x, y, w, h := box.GetRect()
	for j := y; j < y+h; j++ {
		row := make([]rune, w)
		for i := range row {
			row[i] = st.buf.At(x+i, j).R
		}
		if i := strings.Index(string(row), word); i >= 0 {
			return image.Pt(x+len([]rune(string(row)[:i]))+2, j)
		}
	}
	t.Fatalf("no row of the box shows %q", word)
	return image.Point{}
}

// A click on a box of the dashboard opens the panel that is the box in full,
// as typing the panel's name does: the mini map opens the Map, the header's
// bar Advance, the People box People, the Log the full record and Under
// construction the build plan. The pointer is the hand over each.
func TestClicksOnTheDashboardsBoxes(t *testing.T) {
	d, pages := playtestTown(t)
	d.refresh()
	opens := func(line string) string {
		t.Helper()
		d.runForTest(line)
		name := d.overlayMgr.ActiveName()
		d.overlayMgr.Hide()
		if name == "" {
			t.Fatalf("staging: typing %s opened no panel", line)
		}
		return name
	}
	st := newStage(pages, playtestCols, playtestRows)
	if !d.body.miniShown {
		t.Fatal("staging: the mini map does not show on the playtest's screen")
	}
	for _, c := range []struct {
		what string
		box  Primitive
		want string
	}{
		{"the mini map", d.miniMap, opens("map")}, {"the header's bar", d.headerTV, opens("charter")},
		{"the People box", d.workerMiniTV, opens("people")}, {"the Log", d.logTV, opens("logs")},
		{"Under construction", d.economyTab.constructionTV, opens("plan")},
	} {
		st.paint()
		x, y, w, h := c.box.GetRect()
		at := image.Pt(x+w/2, y+h/2)
		st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
		if st.paint(); st.buf.Pointer != kit.PointerHand {
			t.Errorf("over %s the pointer's shape is %d, want the hand", c.what, st.buf.Pointer)
		}
		st.click(at.X, at.Y)
		if got := d.overlayMgr.ActiveName(); got != c.want {
			t.Errorf("a click on %s opened %q, want %q", c.what, got, c.want)
		}
		d.overlayMgr.Hide()
	}
}
