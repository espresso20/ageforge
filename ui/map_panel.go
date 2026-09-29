package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// map_panel.go is the Map panel: the active map style full screen, opened
// with map (aliases citymap and worldmap). The style draws everything but
// the bottom row, a key bar with the panel's own keys (s style, g glyphs,
// Enter stage the inspected command, Esc close). Every other key goes to
// the style (its cursor, Tab through targets, zoom and toggles).

// mapAnimStep is the map animation frame length (about 8 frames a second).
const mapAnimStep = 125 * time.Millisecond

// mapIconsHint is the one-time hint the first Map panel open shows.
const mapIconsHint = "Want real icons on your map? Type icons."

// mapViews is shared by the Map panel and the mini map: the style
// registry, one model builder (the UI goroutine owns it) and the
// since-last-visit baseline.
type mapViews struct {
	reg     *mapstyle.Registry
	builder *mapmodel.Builder

	sinceKey [2]int64
	since    *mapmodel.Visit
}

func newMapViews(reg *mapstyle.Registry) *mapViews {
	return &mapViews{reg: reg, builder: mapmodel.NewBuilder(nil)}
}

// model returns the map model for a snapshot, measured from the load-time
// baseline (state.SessionStart). The builder reuses the last model while
// nothing the maps draw has changed.
func (mv *mapViews) model(state *game.GameState) *mapmodel.Model {
	if s := state.SessionStart; s == nil {
		mv.since, mv.sinceKey = nil, [2]int64{}
	} else if k := [2]int64{int64(s.Tick), s.SavedAt.UnixNano()}; mv.since == nil || k != mv.sinceKey {
		mv.since, mv.sinceKey = mapmodel.VisitFromSession(s), k
	}
	return mv.builder.Model(state, mv.since)
}

// styleSet keeps one live view per style, so switching away and back keeps
// a style's camera and cursor.
type styleSet struct {
	reg   *mapstyle.Registry
	views map[string]mapstyle.Style
}

func (ss *styleSet) get(name string) mapstyle.Style {
	if ss.views == nil {
		ss.views = map[string]mapstyle.Style{}
	}
	if v, ok := ss.views[name]; ok {
		return v
	}
	v, ok := ss.reg.New(name)
	if !ok {
		v, _ = ss.reg.New(ss.reg.Default())
	}
	ss.views[name] = v
	return v
}

// mapPanel is the Map overlay primitive.
type mapPanel struct {
	*tview.Box
	mv     *mapViews
	styles styleSet
	set    mapSettings
	model  *mapmodel.Model
	start  time.Time
	now    func() time.Time // the animation clock (tests pin it)

	world bool // the next open starts on the known world
	hint  bool // show the icons hint in the key bar while open
	note  string

	// settings reads the current settings; save persists a change.
	settings func() mapSettings
	save     func(mapSettings)
	// hintShown records that the icons hint was shown.
	hintShown func()
	// stage closes the panel and puts a command in the prompt.
	stage func(cmd string)
}

func newMapPanel(mv *mapViews) *mapPanel {
	p := &mapPanel{Box: tview.NewBox(), mv: mv, styles: styleSet{reg: mv.reg}, now: time.Now}
	p.start = p.now()
	p.set = mapSettings{Style: mv.reg.Default(), Tier: mapmodel.TierUnicode}
	return p
}

// open is the overlay's build: it runs on every Show. It picks up the
// settings, points the style at the settlement or the world, and decides
// whether this is the open that shows the icons hint.
func (p *mapPanel) open(state game.GameState) tview.Primitive {
	if p.settings != nil {
		p.set = p.settings()
	}
	p.note = ""
	p.current().SetOption(mapstyle.OptWorld, p.world)
	p.world = false
	p.hint = !p.set.HintShown && p.set.Tier != mapmodel.TierNerd
	if p.hint && p.hintShown != nil {
		p.hintShown()
		p.set.HintShown = true
	}
	p.update(state)
	return p
}

// update is the overlay's refresh (every UI tick while open).
func (p *mapPanel) update(state game.GameState) {
	if p.settings != nil {
		s := p.settings()
		s.HintShown = s.HintShown || p.set.HintShown
		p.set = s
	}
	p.model = p.mv.model(&state)
}

func (p *mapPanel) current() mapstyle.Style { return p.styles.get(p.set.Style) }

func (p *mapPanel) frame() mapstyle.Frame {
	return mapstyle.Frame{Model: p.model, Anim: int(p.now().Sub(p.start) / mapAnimStep), Tier: p.set.Tier}
}

// Draw draws the style over all but the last row, and the key bar on it.
func (p *mapPanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	st, f := p.current(), p.frame()
	if h < 3 {
		st.Draw(scr, mapstyle.Rect{X: x, Y: y, W: w, H: h}, f)
		return
	}
	st.Draw(scr, mapstyle.Rect{X: x, Y: y, W: w, H: h - 1}, f)
	p.drawBar(scr, mapstyle.Rect{X: x, Y: y + h - 1, W: w, H: 1}, st, f)
}

// barPart is one run of key-bar text in one style.
type barPart struct {
	text string
	role int // 0 text, 1 key, 2 dim
}

// drawBar is the key bar: the hint or note, the inspected command, then
// the keys. When it does not fit it drops the command, then shortens the
// keys, so the hint and the keys always show.
func (p *mapPanel) drawBar(scr tcell.Screen, r mapstyle.Rect, st mapstyle.Style, f mapstyle.Frame) {
	bg := theme.Color(theme.RoleChip)
	base := tcell.StyleDefault.Background(bg).Foreground(theme.Legible(theme.Color(theme.RoleText), bg, 4.5))
	styles := [3]tcell.Style{
		base,
		base.Foreground(theme.Legible(theme.Color(theme.RoleAccent), bg, 3)).Bold(true),
		base.Foreground(theme.Legible(theme.Color(theme.RoleDim), bg, 3)),
	}
	var lead, cmd []barPart
	if p.hint {
		lead = []barPart{{mapIconsHint, 1}, {"  ", 0}}
	} else if p.note != "" {
		lead = []barPart{{p.note, 1}, {"  ", 0}}
	}
	if in, ok := st.Inspect(f); ok && in.Command != "" {
		cmd = []barPart{{"Enter", 1}, {" type ", 2}, {in.Command, 0}, {" · ", 2}}
	}
	full := []barPart{{"s", 1}, {" style: " + styleTitle(p.mv.reg, p.set.Style) + " · ", 2},
		{"g", 1}, {" glyphs: " + p.set.Tier.String() + " · ", 2}, {"Esc", 1}, {" close", 2}}
	short := []barPart{{"s", 1}, {" style · ", 2}, {"g", 1}, {" glyphs · ", 2}, {"Esc", 1}}
	width := func(parts ...[]barPart) int {
		n := 0
		for _, ps := range parts {
			for _, pt := range ps {
				n += mapstyle.TextLen(pt.text)
			}
		}
		return n
	}
	var parts [][]barPart
	switch avail := r.W - 2; {
	case width(lead, cmd, full) <= avail:
		parts = [][]barPart{lead, cmd, full}
	case width(lead, full) <= avail:
		parts = [][]barPart{lead, full}
	default:
		parts = [][]barPart{lead, short}
	}
	cv := mapstyle.NewCanvas(scr, r, f.Tier, base)
	x := 1
	for _, ps := range parts {
		for _, pt := range ps {
			if x < r.W {
				x = cv.Text(x, 0, r.W-x, pt.text, styles[pt.role])
			}
		}
	}
}

// InputHandler routes keys: the panel's own first, then the style's.
func (p *mapPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		p.handleKey(ev)
	})
}

// handleKey applies one key and reports whether anything used it. Esc is
// the overlay's (it closes the panel before the key gets here).
func (p *mapPanel) handleKey(ev *tcell.EventKey) bool {
	switch {
	case ev.Key() == tcell.KeyEnter:
		in, ok := p.current().Inspect(p.frame())
		if !ok || in.Command == "" {
			p.note = "Nothing to type here. Move the cursor onto a building or a civilization."
			return true
		}
		if p.stage != nil {
			p.stage(in.Command)
		}
		return true
	case ev.Key() == tcell.KeyRune && ev.Rune() == 's':
		p.set.Style = nextStyle(p.mv.reg, p.set.Style)
		p.persist()
		p.hint, p.note = false, "Style: "+styleTitle(p.mv.reg, p.set.Style)
		return true
	case ev.Key() == tcell.KeyRune && ev.Rune() == 'g':
		p.set.Tier = nextTier(p.set.Tier)
		p.persist()
		p.hint, p.note = false, "Glyphs: "+p.set.Tier.String()
		if p.set.Tier == mapmodel.TierNerd {
			p.note += " (boxes? type icons)"
		}
		return true
	}
	used := p.current().HandleKey(ev, p.frame())
	if used {
		p.note = ""
	}
	return used
}

func (p *mapPanel) persist() {
	if p.save != nil {
		p.save(p.set)
	}
}
