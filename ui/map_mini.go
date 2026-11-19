package ui

import (
	"image"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// map_mini.go is the dashboard's mini map: the map's compact view in a
// border, a box about square on screen under the Panels list in the right
// column (dash_layout.go), shown only when the column has room for it after
// the list, and the minimap setting turns it off. It keeps its own view of the map (separate from the Map
// panel's camera) and draws from the model the last refresh built, so a
// redraw between refreshes (typing at the prompt) costs only the compact
// draw. The compact view ends with the since-last-visit news.

// miniMinInnerH is the fewest rows of picture inside the border the mini
// map is worth showing with.
const miniMinInnerH = 6

// miniMap is the mini map primitive (border included).
type miniMap struct {
	*nativeHost
	kit.Base
	mv    *mapViews
	view  mapstyle.Style
	set   mapSettings
	model *mapmodel.Model
	start time.Time
	now   func() time.Time
	// onClick is what a click on the mini map does (the dashboard opens
	// the Map panel).
	onClick func()
}

func newMiniMap(mv *mapViews) *miniMap {
	m := &miniMap{mv: mv, view: mapstyle.New(), now: time.Now}
	m.start = m.now()
	m.set = defaultMapSettings()
	m.nativeHost = newNativeHost()
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
	bg := theme.Of(theme.RoleBackground)
	border := kit.Style{}.Background(bg).Foreground(theme.Legible(theme.Of(theme.RoleBorder), bg, 1.8))
	accent := kit.Style{}.Background(bg).Foreground(theme.Legible(theme.Of(theme.RoleAccent), bg, 3)).Bold(true)
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

func (m *miniMap) Layout(r image.Rectangle) { m.Place(r) }

// Mouse: a click on the mini map does what the dashboard set for it.
func (m *miniMap) Mouse(ev kit.Mouse) bool {
	if ev.Kind != kit.MousePress || ev.Button != kit.ButtonLeft || m.onClick == nil {
		return false
	}
	m.onClick()
	return true
}

// PointerAt: the hand, where a click does something.
func (m *miniMap) PointerAt(image.Point) kit.Pointer { return handIf(m.onClick != nil) }

// Paint draws the border and the compact view inside it. The map still
// writes cells through mapstyle.Cells, at the box's place on the screen: scr puts them on
// the painter.
func (m *miniMap) Paint(p *kit.Painter) {
	if m.nativeHost == nil {
		return
	}
	x, y, w, h := m.GetRect()
	if w < 3 || h < 3 {
		return
	}
	scr := cellsFor(p, m, m.Bounds())
	cv := &mapstyle.Canvas{Scr: scr, R: mapstyle.Rect{X: x, Y: y, W: w, H: h}, Tier: m.set.Tier}
	drawMapFrame(cv, w, h, rowid.Line("ui.map_mini.miniMap.Draw.01", " Map "))

	f := mapFrame(m.model, m.set, m.now().Sub(m.start))
	m.view.DrawCompact(scr, mapstyle.Rect{X: x + 1, Y: y + 1, W: w - 2, H: h - 2}, f)
}
