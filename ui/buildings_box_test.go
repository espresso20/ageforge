package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The playtest's screen and town: a MacBook at a large font, where the
// middle column is about 60 columns by 45 rows, and a Primitive Age town of
// 160 people with 32 huts, 20 stashes and 15 gathering camps. The mini map
// took the top third of that column and the Buildings list was cut off
// mid-entry at the bottom ("Stash x20" showed its name and nothing else).
//
// The playtest's terminal is big enough for the rich layout, whose Panels
// column is at least richRightMin wide, so the screen is the 120 columns of the two
// middle columns and that.
const (
	playtestCols = 2*60 + richRightMin
	// The row over the boxes there is the header, at its full height.
	playtestRows = 45 + headerStdRows + barBigRows
)

// playtestTown is a dashboard on a town like the playtest's.
func playtestTown(t *testing.T) (*Dashboard, *nativePages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	eng.SetStockForTest("food", 10000)
	eng.SetStockForTest("wood", 10000)
	eng.SetTownForTest(map[string]int{
		"hut": 32, "stash": 20, "gathering_camp": 15, "wood_camp": 12, "story_circle": 6, "shrine": 4, "sacred_grove": 2,
	}, 160)
	d, pages := newTestDashboard(eng)
	d.refresh()
	return d, pages
}

// checkCards holds the Buildings box to what it laid out: the cards sit inside
// the box, one under another, nothing overlaps, every card's Build button is
// inside the box, and every part of every price is on the screen whole.
func checkCards(t *testing.T, where string, d *Dashboard, cells []simCell, screenW int) {
	t.Helper()
	b := d.economyTab.buildBox
	x, y, w, h := b.GetRect()
	if w < 10 || h < 5 {
		t.Errorf("%s: the Buildings box is %dx%d", where, w, h)
		return
	}
	if len(b.placed) == 0 {
		t.Errorf("%s: the Buildings box shows no card", where)
		return
	}
	iw, ih := w-2, h-2
	edge := 1
	for i, pc := range b.placed {
		if pc.y < edge || pc.y+pc.rows > 1+ih {
			t.Errorf("%s: card %q is at rows %d-%d of a box with %d inside", where, pc.card.name, pc.y, pc.y+pc.rows, ih)
		}
		edge = pc.y + pc.rows
		key := b.cardKeys.keys[pc.build]
		if r := key.r; r.Min.X < 1 || r.Max.X > 1+iw || r.Min.Y < pc.y || r.Max.Y > pc.y+pc.rows {
			t.Errorf("%s: card %q has its Build button at %v, outside its card (rows %d-%d, %d wide)", where, pc.card.name, r, pc.y, pc.y+pc.rows, iw)
		}
		if i > 0 && key.r.Min.X != b.cardKeys.keys[b.placed[0].build].r.Min.X {
			t.Errorf("%s: the Build buttons are not in one column", where)
		}
		got := strings.Join(rectText(cells, screenW, x+1, y+pc.y, iw, pc.rows), "\n")
		for _, line := range pc.price {
			for _, part := range line {
				if !strings.Contains(got, part.text) {
					t.Errorf("%s: card %q lost the price part %q:\n%s", where, pc.card.name, part.text, got)
				}
			}
		}
		if !strings.Contains(got, pc.card.name) {
			t.Errorf("%s: card %q does not show its name:\n%s", where, pc.card.name, got)
		}
	}
}
