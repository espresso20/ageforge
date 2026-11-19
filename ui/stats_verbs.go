package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// stats_verbs.go is what the Stats page does by click: Prestige, a danger
// button under the Prestige table that asks first and then runs the command a
// player would type.

// prestigeCmd is the command line that prestiges, confirmation included.
const prestigeCmd = "prestige confirm yes"

var (
	txtPrestigeVerb = rowid.E("ui.stats_verbs.prestige.01", "Prestige")
	txtPrestigeAsk  = rowid.E("ui.stats_verbs.ask.01", "Prestige and begin a new run?\n\nYour points, carry-overs and Era Mastery are kept.")
)

// prestigeButton draws Prestige at (x, y), off until a prestige is open, and
// returns the next y.
func (pg *statsPage) prestigeButton(st game.GameState, x, y int) int {
	word := txtPrestigeVerb.String()
	pg.g.actionButton(x, y, actionSpec{kind: keyDanger, label: word, cmd: prestigeCmd, ask: txtPrestigeAsk.String(), yes: word, off: !st.Prestige.CanPrestige})
	return y + actionRows
}

// pageOrg is where the page's first cell sits on the screen.
func (p *statsPanel) pageOrg() image.Point {
	return p.Bounds().Min.Add(image.Pt(2, 3-p.scroll))
}

// pageMouse gives the page's buttons the mouse: a press or a move over the
// part of the page in view.
func (p *statsPanel) pageMouse(m kit.Mouse) bool {
	if p.page == nil || m.Kind == kit.MouseWheel {
		return false
	}
	if p.pagePt(m) == image.Pt(-1, -1) {
		m.Kind = kit.MouseLeave
	}
	return p.pageBtn.mouse(p.page.g, p.pageOrg(), m)
}

// pagePointer is the hand over a page button a click would press.
func (p *statsPanel) pagePointer(at image.Point) kit.Pointer {
	if p.page == nil || at.Y < p.Bounds().Min.Y+3 || at.Y >= p.Bounds().Min.Y+3+p.viewRows() {
		return kit.PointerArrow
	}
	return p.pageBtn.pointer(p.page.g, p.pageOrg(), at)
}
