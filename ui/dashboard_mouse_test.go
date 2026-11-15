package ui

import (
	"image"
	"strings"
	"testing"
	"time"

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

// A click on a building in the dashboard's Buildings list writes its build
// command in the prompt and builds nothing; a double click builds one, as
// Enter on that line would, and leaves the prompt empty. A building at its
// limit is no click target. The pointer is the hand over a building.
func TestAClickOnABuildingWritesItsCommand(t *testing.T) {
	d, pages := playtestTown(t)
	d.runForTest("minimap off")
	d.refresh()
	st := newStage(pages, playtestCols, playtestRows)
	now := time.Unix(1_700_000_000, 0)
	st.root.Now = func() time.Time { return now }
	// A build goes into the queue; it stands a while later.
	camps := func() int {
		st := d.engine.GetState()
		n := st.Buildings["gathering_camp"].Count
		return n + len(st.BuildQueue) // nothing else is built in this test
	}
	before := camps()
	at := rowWith(t, st, d.economyTab.buildingTV, "Gathering Camp")

	st.mouse(kit.Mouse{Kind: kit.MouseMove, X: at.X, Y: at.Y})
	if st.paint(); st.buf.Pointer != kit.PointerHand {
		t.Errorf("over a building the pointer's shape is %d, want the hand", st.buf.Pointer)
	}
	st.click(at.X, at.Y)
	if got := d.inputField.GetText(); got != "build gathering_camp" || !d.inputField.HasFocus() || camps() != before {
		t.Errorf("a click: the prompt reads %q (has the keyboard: %v) and there are %d camps built or under way, want the command written, the keyboard there and still %d", got, d.inputField.HasFocus(), camps(), before)
	}
	// The row under it (the cost) is the same building's.
	now = now.Add(time.Minute)
	d.inputField.SetText("")
	st.click(at.X, at.Y+1)
	if got := d.inputField.GetText(); got != "build gathering_camp" {
		t.Errorf("a click on the building's cost line wrote %q", got)
	}
	// A double click builds one.
	now = now.Add(time.Minute)
	d.inputField.SetText("")
	st.click(at.X, at.Y)
	st.click(at.X, at.Y)
	if got := d.inputField.GetText(); got != "" || camps() != before+1 {
		t.Errorf("a double click: the prompt reads %q and there are %d camps built or under way, want it empty and %d", got, camps(), before+1)
	}
}

// A click on the mini map opens the Map panel and a click on the Ledger the
// Advance panel, as typing map and charter does. The pointer is the hand
// over both.
func TestClicksOnTheMiniMapAndTheLedger(t *testing.T) {
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
	wantMap, wantAdvance := opens("map"), opens("charter")
	st := newStage(pages, playtestCols, playtestRows)
	if !d.mapDock.shown {
		t.Fatal("staging: the mini map does not show on the playtest's screen")
	}
	for _, c := range []struct {
		what string
		box  Primitive
		want string
	}{{"the mini map", d.miniMap, wantMap}, {"the Ledger", d.ledgerTV, wantAdvance}} {
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
