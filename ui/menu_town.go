package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// menu_town.go draws the player's own town behind the main menu: the map
// of their current game, by the game's own map renderer, in their glyph set
// and theme, at the age the game is in, moving as the Map panel's does.
//
// It is a picture of the save, not the game. The state comes from
// game.ViewSave, which restores the save into an engine of its own and
// throws the engine away: nothing is loaded into the engine in play, no
// tick runs, no time away is applied and nothing is written. The map draws
// its full view onto a screen of the menu's own; the rows that hold the
// scene (SceneInset) are copied to the menu, dimmed toward the page and
// faded at the top and the right so the menu over them reads.

// menuTown is the picture of one save.
type menuTown struct {
	model *mapmodel.Model
	style mapstyle.Style
	// builder lays the town out; it keeps the last model while nothing the
	// maps draw has changed (update).
	builder *mapmodel.Builder
	// name and age caption the picture: the town's name and its age's.
	name, age string

	// buf is the cells the map draws the picture into, before the menu
	// takes them over, faded, into its own.
	buf kit.Buffer
	// failed is set when the map could not draw the picture. The menu
	// then shows its first page; it never shows an error.
	failed bool
}

// newMenuTown builds the picture of a save's state. nil when there is
// nothing to draw.
func newMenuTown(st *game.GameState) (t *menuTown) {
	return newMenuTownOn(mapmodel.NewBuilder(nil), st, true)
}

// newMenuTownOn is newMenuTown on a builder the caller keeps, so that two
// pictures of one game (the arrival screen's town before and after) lay the
// land out once: the land is most of what a picture costs to make. follows
// is for the picture that will follow the game (update): it takes the
// builder's cache. The other is built aside and stays as it is.
func newMenuTownOn(builder *mapmodel.Builder, st *game.GameState, follows bool) (t *menuTown) {
	if st == nil || builder == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			t = nil // a state the maps cannot lay out: no picture
		}
	}()
	style := mapstyle.New()
	// The picture alone: no cursor, no legend, no overlay, no news.
	for _, o := range []mapstyle.Option{mapstyle.OptFlows, mapstyle.OptInspect, mapstyle.OptLegend, mapstyle.OptChanges, mapstyle.OptWorld} {
		style.SetOption(o, false)
	}
	var m *mapmodel.Model
	if follows {
		m = builder.Model(st, nil)
	} else {
		m = builder.Build(st, nil)
	}
	if m == nil {
		return nil
	}
	return &menuTown{model: m, style: style, builder: builder, name: m.TownName, age: m.AgeName}
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

// The least size the map draws its full view at (below it the map draws its
// compact view, which is not a picture of the town).
const (
	townMinW = 60
	townMinH = 16
)

// draw paints the picture into the w by h block of g at (x0, y0) and
// reports whether it did. flare is the strike's flash.
func (t *menuTown) draw(g *mGrid, pal *menuPalette, x0, y0, w, h int, f mapstyle.Frame, flare bool) (ok bool) {
	return t.paint(g, pal, x0, y0, w, h, false, f, flare)
}

// drawFoot is draw for a block that may have fewer rows than the map lays
// its full view out in: the picture is then laid out at the least height
// it takes and the block shows the foot of it. (The arrival screen's land
// is what is left under a name and its lines.)
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
	top, bottom := t.style.SceneInset(w, h)
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
	if fw != t.buf.W || fh != t.buf.H {
		t.buf.Resize(fw, fh)
	}
	f.Model = t.model
	t.style.Draw(kitCells{p: kit.NewPainter(&t.buf)}, mapstyle.Rect{W: fw, H: fh}, f)

	for y := 0; y < h; y++ {
		rowLevel := min(3, y)
		for x := 0; x < w; x++ {
			c := t.buf.At(x, y+top)
			r, attr := c.R, c.Attr
			fg, bg := kit.RGB(c.Fg.R, c.Fg.G, c.Fg.B), kit.RGB(c.Bg.R, c.Bg.G, c.Bg.B)
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
				if attr&kit.Bold != 0 && g.in(x0+x, y0+y) {
					g.c[(y0+y)*g.w+x0+x].bold = true
				}
				continue
			}
			g.set(x0+x, y0+y, r, theme.Mix(fg, pal.bg, k), theme.Mix(bg, pal.bg, k), attr&kit.Bold != 0)
		}
	}
	return true
}

// close lets go of the picture's cells.
func (t *menuTown) close() {
	if t != nil {
		t.buf = kit.Buffer{}
	}
}
