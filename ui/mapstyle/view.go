package mapstyle

// view.go is the map view: the picture the procedural engine (package
// mapgen) draws, its cells coloured in the player's theme, with the map's own
// header, news, inspector and words round it. It is the Style the Map panel,
// the mini map and the main menu draw through.

import (
	"slices"
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// The engine's zoom levels: the whole town squeezed to fit, one world cell to
// a screen cell, and the same close-up framed on one building line's district
// with the rest of the town dimmed.
const (
	zOverview = 0
	zCloseUp  = 1
	zDistrict = 2
)

var zoomNames = [3]vibe.Entry{
	zOverview: vibe.E("ui.mapstyle.roguelike.view.zoomNames.01", "REGION"),
	zCloseUp:  vibe.E("ui.mapstyle.roguelike.view.zoomNames.02", "SETTLEMENT"),
	zDistrict: vibe.E("ui.mapstyle.roguelike.view.zoomNames.03", "DISTRICT"),
}

// A panel at least this big shows the close-up; smaller ones (80 by 24, the
// mini map) show the whole town.
const closeUpMinW, closeUpMinH = 100, 30

// shared is the one generator every view draws from, so the Map panel and the
// mini map build a seed's land and plan once.
var shared struct {
	sync.Mutex
	gen mapgen.Generator
}

// view is one open map.
type view struct {
	pal  *cellPalette // the engine's cell colours
	ink  *inks        // the chrome's colours
	tier mapmodel.GlyphTier

	zoom     int  // the player's zoom, once chosen
	zoomSet  bool // else the panel's size chooses
	overview bool // what the last draw showed

	cur      mapgen.Pt
	seed     int64
	placed   bool
	recentre bool
	cam      mapgen.Pt
	camKey   [2]int
	inspect  bool
	legend   bool
	flows    bool // the flows overlay
	changes  bool // the news of what is new since the last visit

	district string // the building line in focus at the district zoom
	// what is new since the last visit stays lit for freshFrames after the map
	// first shows it: freshKey says which news that was, freshFrom when
	freshKey  uint64
	freshFrom int

	mark   mapgen.Pt // where the Map panel's cursor stands, for the mini map
	marked bool

	// the last overview drawn and the scene it came from, so the cursor can
	// ask what a cell is
	ov    *mapgen.Overview
	sc    *mapgen.Scene
	scKey sceneKey
}

// sceneKey says which frame a scene was drawn for.
type sceneKey struct {
	m        *mapmodel.Model
	anim     int
	visit    int
	fresh    bool
	district string
}

// New makes a map view drawn by the engine.
func New() Style {
	return &view{inspect: true, legend: true, changes: true}
}

// scene is the world drawn for a frame, reused while the frame is the same.
func (v *view) scene(f Frame) *mapgen.Scene {
	if f.Model == nil {
		return nil
	}
	if k := (sceneKey{f.Model, f.Anim, f.VisitFrame(), v.freshLit(f), v.focus()}); v.sc == nil || k != v.scKey {
		shared.Lock()
		v.sc, v.scKey = shared.gen.DrawAt(f.Model, mapgen.Moment{Anim: f.Anim, Visit: k.visit, Fresh: k.fresh, District: k.district}), k
		shared.Unlock()
	}
	if line := v.focus(); line != "" && !slices.ContainsFunc(v.sc.Districts, func(d mapgen.District) bool { return d.Line == line }) {
		v.zoom, v.district = zCloseUp, "" // the district is gone: back to the close-up
		return v.scene(f)
	}
	if !v.placed || v.seed != f.Model.Seed {
		v.seed, v.placed, v.recentre = f.Model.Seed, true, true
		v.cur = v.sc.Centre
	}
	return v.sc
}

func (v *view) SetOption(o Option, on bool) {
	switch o {
	case OptWorld:
		if on {
			v.zoom, v.zoomSet = zOverview, true
		} else if v.zoomSet && v.zoom == zOverview {
			v.zoomSet = false // back to what the size chooses
		}
	case OptInspect:
		v.inspect = on
	case OptLegend:
		v.legend = on
	case OptFlows:
		v.flows = on
	case OptChanges:
		v.changes = on
	}
}

// wantsOverview says whether a panel r shows the whole town.
func (v *view) wantsOverview(r Rect) bool {
	if v.zoomSet {
		return v.zoom == zOverview
	}
	return r.W < closeUpMinW || r.H < closeUpMinH
}

func (v *view) SceneInset(w, h int) (top, bottom int) { return 1, 4 }

func (v *view) Pointer() (mapmodel.Pt, bool) {
	return mapmodel.Pt{X: v.cur.X, Y: v.cur.Y}, v.placed && v.inspect
}

func (v *view) Point(p mapmodel.Pt, ok bool) { v.mark, v.marked = mapgen.Pt{X: p.X, Y: p.Y}, ok }

// blit puts a grid on the canvas at (x0, y0): the one thing the screen does
// with it.
func (v *view) blit(cv *Canvas, x0, y0 int, g *mapgen.Grid) {
	v.pal = cellPaletteFor(v.pal, g)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.Cells[y*g.W+x]
			r := c.R
			if r == 0 {
				r = ' '
			}
			cv.Put(x0+x, y0+y, r, v.pal.style(c))
		}
	}
}

// Blit draws a grid, as the screen layer would, onto a screen inside r. The
// lab pages use it to show the engine's pictures in a theme without a map
// panel round them.
func Blit(scr tcell.Screen, r Rect, g *mapgen.Grid, tier mapmodel.GlyphTier) {
	v := &view{}
	v.blit(NewCanvas(scr, r, tier, tcell.StyleDefault), 0, 0, g)
}
