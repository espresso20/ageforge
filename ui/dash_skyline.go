package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// dash_skyline.go is the skyline strip: the town's skyline, in a border,
// under the Buildings list on the dashboard. It is the skyline map style's
// compact view (the Map panel's skyline, squeezed into the strip), always in
// that style whatever the map style setting is, and read-only: no cursor and
// no keys. It moves with the Map panel's clock when the motion setting is on
// and holds still when it is off, and it is redrawn on the dashboard's own
// refresh. `map` still opens the full Map panel.
//
// Its height is half the Log box's, round down (stripHeight), shown only when
// that leaves stripMinInner rows or more inside the border and the terminal
// is stripMinCols by stripMinRows or larger. Otherwise the Buildings list
// has the whole column (mapDock).

const (
	// stripMinInner is the fewest rows inside the strip's border.
	stripMinInner = 6
	// stripMinCols and stripMinRows are the smallest terminal the strip is
	// shown on: under them the Buildings list needs the column.
	stripMinCols = 100
	stripMinRows = 30
	// stripMinBody is the rows the Buildings list keeps above the strip.
	stripMinBody = 12
)

// stripHeight is the strip's height, border included, for a Log box logRows
// tall: half of it, round down, or 0 when that leaves too few rows inside.
func stripHeight(logRows int) int {
	if h := logRows / 2; h-2 >= stripMinInner {
		return h
	}
	return 0
}

// stripFits says whether a w by h terminal is large enough for the strip.
func stripFits(w, h int) bool { return w >= stripMinCols && h >= stripMinRows }

// skylineStrip is the strip primitive (border included).
type skylineStrip struct {
	*tview.Box
	mv    *mapViews
	view  mapstyle.Style
	set   mapSettings
	model *mapmodel.Model
	start time.Time
	now   func() time.Time
}

func newSkylineStrip(mv *mapViews) *skylineStrip {
	s := &skylineStrip{Box: tview.NewBox(), mv: mv, now: time.Now}
	s.start = s.now()
	s.set = defaultMapSettings(mv.reg)
	if v, ok := mv.reg.New("skyline"); ok {
		// The picture alone: no cursor, no flows overlay, no legend.
		for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend} {
			v.SetOption(o, false)
		}
		s.view = v
	}
	return s
}

// update takes the settings and the snapshot of a refresh (UI goroutine).
func (s *skylineStrip) update(set mapSettings, state *game.GameState) {
	s.set = set
	s.model = s.mv.model(state)
}

// Draw draws the border and, inside it, the skyline.
func (s *skylineStrip) Draw(scr tcell.Screen) {
	x, y, w, h := s.GetRect()
	if w < 3 || h < 3 || s.view == nil {
		return
	}
	cv := &mapstyle.Canvas{Scr: scr, R: mapstyle.Rect{X: x, Y: y, W: w, H: h}, Tier: s.set.Tier}
	drawMapFrame(cv, w, h, " Skyline ")
	f := mapFrame(s.model, s.set, s.now().Sub(s.start))
	s.view.DrawCompact(scr, mapstyle.Rect{X: x + 1, Y: y + 1, W: w - 2, H: h - 2}, f)
}
