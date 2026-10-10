package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// menu_town.go draws the player's own town behind the main menu: the map
// of their current game, by the game's own map renderer, in their map
// style, glyph set and theme, at the age the game is in, moving as the Map
// panel's does.
//
// It is a picture of the save, not the game. The state comes from
// game.ViewSave, which restores the save into an engine of its own and
// throws the engine away: nothing is loaded into the engine in play, no
// tick runs, no time away is applied and nothing is written. The style
// draws its full view onto a screen of the menu's own; the rows that hold
// the scene (mapstyle.Scenic) are copied to the menu, dimmed toward the
// page and faded at the top and the right so the menu over them reads.

// menuTown is the picture of one save.
type menuTown struct {
	model *mapmodel.Model
	style mapstyle.Style
	// builder lays the town out; it keeps the last model while nothing the
	// maps draw has changed (update).
	builder *mapmodel.Builder
	// name and age caption the picture: the town's name and its age's.
	name, age string

	sim    tcell.SimulationScreen
	scr    tcell.Screen
	sw, sh int
	// failed is set when the style could not draw the picture. The menu
	// then shows its first page; it never shows an error.
	failed bool
}

// newMenuTown builds the picture of a save's state in a style. nil when
// there is nothing to draw.
func newMenuTown(st *game.GameState, reg *mapstyle.Registry, styleName string) (t *menuTown) {
	if st == nil || reg == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			t = nil // a state the maps cannot lay out: no picture
		}
	}()
	style, ok := reg.New(styleName)
	if !ok {
		if style, ok = reg.New(reg.Default()); !ok {
			return nil
		}
	}
	// The picture alone: no cursor, no legend, no overlay, no news.
	for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend, mapstyle.OptChanges, mapstyle.OptWorld} {
		style.SetOption(o, false)
	}
	builder := mapmodel.NewBuilder(nil)
	m := builder.Model(st, nil)
	if m == nil {
		return nil
	}
	t = &menuTown{model: m, style: style, builder: builder, age: m.AgeName}
	if m.Town.World != nil {
		t.name = m.Town.World.Name
	}
	return t
}

// update draws the picture from a newer state of the same game from here
// on (the arrival screen stays up while the game runs under it).
func (t *menuTown) update(st *game.GameState) {
	if t == nil || t.failed || st == nil {
		return
	}
	defer func() {
		if recover() != nil {
			t.failed = true
		}
	}()
	if m := t.builder.Model(st, nil); m != nil {
		t.model = m
	}
}

// townFade is how far a cell of the picture is mixed toward the page, by
// level: the picture's top rows and right columns rise through the first
// three, the body is the fourth, and the last is the flare of a strike.
var townFade = [5]float64{0.80, 0.62, 0.46, 0.34, 0.12}

// The least size a style draws its full view at (below it a style draws
// its compact view, which is not a picture of the town).
const (
	townMinW = 60
	townMinH = 16
)

// draw paints the picture into the w by h block of g at (x0, y0) and
// reports whether it did. flare is the strike's flash.
func (t *menuTown) draw(g *mGrid, pal *menuPalette, x0, y0, w, h int, f mapstyle.Frame, flare bool) (ok bool) {
	return t.paint(g, pal, x0, y0, w, h, false, f, flare)
}

// drawFoot is draw for a block that may have fewer rows than a style lays
// its full view out in: the picture is then laid out at the least height
// it takes and the block shows the foot of it, so a skyline keeps its
// ground and a map the town at its middle. (The arrival screen's land is
// what is left under a name and its lines.)
func (t *menuTown) drawFoot(g *mGrid, pal *menuPalette, x0, y0, w, h int, f mapstyle.Frame, flare bool) (ok bool) {
	return t.paint(g, pal, x0, y0, w, h, true, f, flare)
}

func (t *menuTown) paint(g *mGrid, pal *menuPalette, x0, y0, w, h int, foot bool, f mapstyle.Frame, flare bool) (ok bool) {
	if t == nil || t.failed || w <= 0 || h <= 0 {
		return false
	}
	defer func() {
		if recover() != nil {
			t.failed, ok = true, false
		}
	}()
	top, bottom := 0, 0
	if sc, scenic := t.style.(mapstyle.Scenic); scenic {
		top, bottom = sc.SceneInset(w, h)
	}
	fw, fh := w, h+top+bottom
	if foot && fh < townMinH {
		// Rows of the picture above the block: it is laid out taller than
		// the block and the block shows its foot.
		top += townMinH - fh
		fh = townMinH
	}
	if fw < townMinW || fh < townMinH {
		return false
	}
	if t.sim == nil {
		sim := tcell.NewSimulationScreen("UTF-8")
		if err := sim.Init(); err != nil {
			t.failed = true
			return false
		}
		t.sim, t.scr = sim, theme.WrapScreen(sim)
	}
	if fw != t.sw || fh != t.sh {
		t.sim.SetSize(fw, fh)
		t.sw, t.sh = fw, fh
	}
	f.Model = t.model
	t.style.Draw(t.scr, mapstyle.Rect{W: fw, H: fh}, f)

	for y := 0; y < h; y++ {
		rowLevel := min(3, y)
		for x := 0; x < w; x++ {
			r, _, st, _ := t.sim.GetContent(x, y+top)
			fg, bg, attr := st.Decompose()
			if (r == ' ' || r == 0) && bg == pal.bg {
				continue // open ground: the sky and the sparks show through
			}
			level := min(rowLevel, (w-1-x)/3)
			if flare {
				level = 4
			}
			k := townFade[level]
			if bg == pal.bg {
				g.put(x0+x, y0+y, r, theme.Mix(fg, pal.bg, k))
				if attr&tcell.AttrBold != 0 && g.in(x0+x, y0+y) {
					g.c[(y0+y)*g.w+x0+x].bold = true
				}
				continue
			}
			g.set(x0+x, y0+y, r, theme.Mix(fg, pal.bg, k), theme.Mix(bg, pal.bg, k), attr&tcell.AttrBold != 0)
		}
	}
	return true
}

// close lets go of the picture's screen.
func (t *menuTown) close() {
	if t != nil && t.sim != nil {
		t.sim.Fini()
		t.sim, t.scr = nil, nil
		t.sw, t.sh = 0, 0
	}
}
