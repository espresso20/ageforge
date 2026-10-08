// Package mapstyle is the contract between the map model and the map
// styles. There is one model (package mapmodel) and several thin renderers
// on top of it; a style draws a Frame, handles its own keys and says what
// its cursor is on, in the command vocabulary. The Registry lists the
// styles a player can pick (a "map style" setting), so a new style plugs in
// with one entry.
package mapstyle

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
)

// Rect is a screen rectangle.
type Rect struct{ X, Y, W, H int }

// Frame is what a style draws: the model, the animation frame and the
// glyph tier. Everything a frame shows is a function of these, so the same
// frame draws the same cells.
type Frame struct {
	Model *mapmodel.Model
	// Anim counts animation frames (the maps animate at about 8 per
	// second, independent of the game tick). Ticking life keys off it.
	Anim int
	// Clock counts the same frames, and keeps counting while the motion
	// setting holds Anim still: what happens in the world (the rare visitor
	// arriving and leaving) keys off it, so it happens with motion off too.
	// 0 means "the same as Anim", for a frame built without one.
	Clock int
	Tier  mapmodel.GlyphTier
}

// VisitFrame is the frame the rare visitor (mapmodel.SightingAt) is drawn
// and inspected at, or -1 when none is out. While the map moves it is the
// clock's frame. With the motion setting off the visitor still comes and
// goes on the clock, and stands still at the middle of its visit while it
// is here: the world goes on, the picture holds.
func (f Frame) VisitFrame() int {
	if f.Clock == 0 || f.Clock == f.Anim {
		return f.Anim
	}
	if f.Model == nil {
		return -1
	}
	if sg, ok := f.Model.SightingAt(f.Clock); ok {
		return sg.Start + sg.Frames/2
	}
	return -1
}

// Inspection is what a style's cursor is on.
type Inspection struct {
	Title string   // "Guildhall ×2"
	Lines []string // details, one per line
	// Command is a whole command the player can type for it ("build
	// guildhall", "diplomacy gift ironhold_clans"), "" when none fits.
	Command string
	// Kind says what sort of thing it is, for code that reacts to what the
	// player looked at: KindMover, KindAlien, or "" for everything else.
	Kind string
}

// Inspection kinds.
const (
	// KindMover is something passing through: a cart, a train, a plane.
	KindMover = "mover"
	// KindAlien is the rare visitor (mapmodel.SightingAt). Spotting one is
	// meant to earn a secret badge.
	KindAlien = "alien"
)

// Style is one way of drawing the map.
type Style interface {
	// Name is the registry key ("roguelike").
	Name() string
	// Draw renders the full view into r, writing every cell of r and
	// nothing outside it.
	Draw(scr tcell.Screen, r Rect, f Frame)
	// DrawCompact renders the glanceable mini view (a 40x15 sidebar panel)
	// into r, with the same guarantees.
	DrawCompact(scr tcell.Screen, r Rect, f Frame)
	// HandleKey applies a key to the view (cursor, zoom, toggles) and
	// reports whether it used it.
	HandleKey(ev *tcell.EventKey, f Frame) bool
	// Inspect reports what the cursor is on; ok is false when the style has
	// no cursor out.
	Inspect(f Frame) (in Inspection, ok bool)
	// SetOption turns a shared option on or off. Styles ignore options they
	// do not support.
	SetOption(o Option, on bool)
}

// Scenic is implemented by a style whose full view can be shown as a
// picture, without the bars round it. SceneInset reports how many rows of
// the full view at w by h are above and below the scene itself: the
// header, the status lines, the key bar. The main menu draws a save's map
// behind itself and leaves those rows out.
type Scenic interface {
	SceneInset(w, h int) (top, bottom int)
}

// CompactNews is implemented by a style whose compact view prints the
// since-last-visit news itself, so the mini map's frame does not repeat it.
type CompactNews interface {
	CompactShowsNews() bool
}

// Option is a view option several styles share, so a setting or a key
// binding can drive any of them the same way.
type Option uint8

const (
	// OptFlows overlays the flows summary: full stores, understaffed
	// buildings, idle workers.
	OptFlows Option = iota
	// OptInspect puts the cursor out.
	OptInspect
	// OptLegend shows the legend when there is room.
	OptLegend
	// OptChanges highlights what is new since the last visit.
	OptChanges
	// OptWorld opens the view on the known world rather than the
	// settlement (roguelike: the region zoom). Off returns to the
	// settlement. Styles with one view ignore it.
	OptWorld
)

// Entry describes one style for the registry.
type Entry struct {
	Name  string // key for "map style <name>"
	Title string // display name
	Blurb string // one line for a picker
	New   func() Style
}

// Registry is the set of styles a player can pick from. It holds no global
// state: build one with NewRegistry (ui/mapstyle/all has the standard set).
type Registry struct {
	entries []Entry
}

// NewRegistry makes a registry; the first entry is the default.
func NewRegistry(entries ...Entry) *Registry {
	return &Registry{entries: append([]Entry(nil), entries...)}
}

// Entries lists the styles in order.
func (r *Registry) Entries() []Entry { return append([]Entry(nil), r.entries...) }

// Names lists the style keys in order.
func (r *Registry) Names() []string {
	out := make([]string, len(r.entries))
	for i, e := range r.entries {
		out[i] = e.Name
	}
	return out
}

// Default is the first style's key.
func (r *Registry) Default() string {
	if len(r.entries) == 0 {
		return ""
	}
	return r.entries[0].Name
}

// New makes a fresh view of the named style.
func (r *Registry) New(name string) (Style, bool) {
	for _, e := range r.entries {
		if e.Name == name {
			return e.New(), true
		}
	}
	return nil, false
}
