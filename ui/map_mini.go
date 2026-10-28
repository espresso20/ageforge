package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/procmap"
	"github.com/espresso20/ageforge/vibe"
)

// map_mini.go is the dashboard's mini map: the map's compact view in a
// border, docked above the Buildings list. It is short (at most
// miniMaxInnerH rows inside the border, and never more than a quarter of
// the column) so the list keeps most of the room, and the minimap setting
// turns it off. It keeps its own view of the map (separate from the Map
// panel's camera) and draws from the model the last refresh built, so a
// redraw between refreshes (typing at the prompt) costs only the compact
// draw. The compact view ends with the since-last-visit news.

const (
	miniMaxInnerH = 9
	miniMinInnerH = 6
	// miniMinW is the narrowest column the mini map shows in, border
	// included.
	miniMinW = 42
	// dockMinBody is the rows the Buildings list keeps under the mini map;
	// with less room the mini map hides and the list gets it all.
	dockMinBody = 18

	// miniMapMinCols and miniMapMinRows are the smallest terminal the
	// dashboard shows the mini map on: 105 by 34. Beside the right column
	// the screen is split in two, the odd column going to the right half,
	// where the mini map is, and that half must be miniMinW wide. Under the
	// rows the dashboard keeps for itself (the Ledger is its standard size
	// by then, ledgerExtraRows taller than the smallest) the dock needs the
	// shortest map, its border and dockMinBody rows of list. The wiki and
	// the minimap command quote them (TestMiniMapAppearsWhereTheDocsSay).
	miniMapMinCols = 2*miniMinW - 1 + sidebarW
	miniMapMinRows = dashHeaderRows + ledgerExtraRows + promptRows + miniMinInnerH + 2 + dockMinBody
)

// miniHeight is the mini map's height, border included, in a w x h dock:
// a quarter of the column at most, capped at miniMaxInnerH rows inside,
// and 0 (hidden) when that leaves it under miniMinInnerH rows or the
// Buildings list under dockMinBody.
func miniHeight(w, h int) int {
	inner := min(miniMaxInnerH, (h-2)/4)
	if w < miniMinW || inner < miniMinInnerH || h-inner-2 < dockMinBody {
		return 0
	}
	return inner + 2
}

// miniMap is the mini map primitive (border included).
type miniMap struct {
	*tview.Box
	mv    *mapViews
	view  mapstyle.Style
	set   mapSettings
	model *mapmodel.Model
	start time.Time
	now   func() time.Time
}

func newMiniMap(mv *mapViews) *miniMap {
	m := &miniMap{Box: tview.NewBox(), mv: mv, view: procmap.New(), now: time.Now}
	m.start = m.now()
	m.set = defaultMapSettings()
	return m
}

// update takes the settings and the snapshot of a refresh (UI goroutine).
func (m *miniMap) update(set mapSettings, state *game.GameState) {
	m.set = set
	m.model = m.mv.model(state)
}

// pointAt marks the Map panel's cursor on the mini map, as the same
// highlighted cell. The two keep a view each: the panel's has the cursor.
func (m *miniMap) pointAt(p *mapPanel) {
	if p == nil {
		return
	}
	m.view.Point(p.view.Pointer())
}

// drawMapFrame draws the border of a map box w by h cells at the canvas's
// origin, with title in the top border.
func drawMapFrame(cv *mapstyle.Canvas, w, h int, title string) {
	bg := theme.Color(theme.RoleBackground)
	border := tcell.StyleDefault.Background(bg).Foreground(theme.Legible(theme.Color(theme.RoleBorder), bg, 1.8))
	accent := tcell.StyleDefault.Background(bg).Foreground(theme.Legible(theme.Color(theme.RoleAccent), bg, 3)).Bold(true)
	for i := 1; i < w-1; i++ {
		cv.Put(i, 0, '─', border)
		cv.Put(i, h-1, '─', border)
	}
	for j := 1; j < h-1; j++ {
		cv.Put(0, j, '│', border)
		cv.Put(w-1, j, '│', border)
	}
	cv.Put(0, 0, '┌', border)
	cv.Put(w-1, 0, '┐', border)
	cv.Put(0, h-1, '└', border)
	cv.Put(w-1, h-1, '┘', border)
	cv.Text(1, 0, w-2, title, accent)
}

// Draw draws the border and the compact view inside it.
func (m *miniMap) Draw(scr tcell.Screen) {
	x, y, w, h := m.GetRect()
	if w < 3 || h < 3 {
		return
	}
	cv := &mapstyle.Canvas{Scr: scr, R: mapstyle.Rect{X: x, Y: y, W: w, H: h}, Tier: m.set.Tier}
	drawMapFrame(cv, w, h, vibe.Line("ui.map_mini.miniMap.Draw.01", " Map "))

	f := mapFrame(m.model, m.set, m.now().Sub(m.start))
	m.view.DrawCompact(scr, mapstyle.Rect{X: x + 1, Y: y + 1, W: w - 2, H: h - 2}, f)
}

// mapDock stacks the mini map above a body (the Buildings list) and lays
// them out itself on every draw: the mini map shows only when it is on and
// the dock has room for it and dockMinBody rows of body, so on a small
// terminal (80x24) it hides and the body gets the whole column, with no
// flicker between refreshes.
type mapDock struct {
	*tview.Box
	mini *miniMap
	body tview.Primitive
	// off is the minimap setting turned off: the body gets the column. The
	// dashboard sets this from the account's setting on every refresh.
	off bool
	// shown reports whether the last draw showed the mini map; drawn that
	// there was a draw at all (before it, the dock has no real size).
	shown, drawn bool
}

func newMapDock(mini *miniMap, body tview.Primitive) *mapDock {
	return &mapDock{Box: tview.NewBox(), mini: mini, body: body}
}

// wantsModel reports whether the mini map needs a model: it is on and
// showing, or nothing has been laid out yet.
func (d *mapDock) wantsModel() bool {
	_, _, w, h := d.GetRect()
	return !d.off && (!d.drawn || miniHeight(w, h) > 0)
}

func (d *mapDock) Draw(scr tcell.Screen) {
	x, y, w, h := d.GetRect()
	mh := miniHeight(w, h)
	if d.off || d.mini.model == nil {
		mh = 0 // off, or no refresh has built the map yet
	}
	d.shown, d.drawn = mh > 0, true
	if d.shown {
		d.mini.SetRect(x, y, w, mh)
		d.mini.Draw(scr)
	}
	d.body.SetRect(x, y+mh, w, h-mh)
	d.body.Draw(scr)
}

// Focus and input go to the body; the mini map takes no keys.
func (d *mapDock) Focus(delegate func(p tview.Primitive)) { delegate(d.body) }

func (d *mapDock) HasFocus() bool { return d.body.HasFocus() }

func (d *mapDock) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return d.body.MouseHandler()
}
