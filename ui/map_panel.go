package ui

import (
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// map_panel.go is the Map panel: the active map style over the whole
// screen but the command bar, opened with map (aliases citymap and
// worldmap). The command bar keeps the keyboard while the panel is open, so
// commands work as they do anywhere: printable keys go to the prompt, and
// the map takes only the keys that print nothing (routeKey): the arrows,
// Tab and Shift-Tab, PgUp and PgDn, Home and End, and Enter on an empty
// prompt, which types the inspected command. Style, glyphs and the flows
// overlay are commands (map style, map glyphs, map flows). The style draws
// everything but the panel's bottom row, a key bar.

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
	flows bool // the flows overlay, in every style (map flows)
	hint  bool // show the icons hint in the key bar while open
	note  string

	// settings reads the current settings.
	settings func() mapSettings
	// hintShown records that the icons hint was shown.
	hintShown func()
	// stage puts a command in the prompt, unrun, for the player's Enter.
	stage func(cmd string)
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the map does not take to the command bar, and
	// the keyboard with it, when the panel itself has the focus.
	toPrompt func(ev *tcell.EventKey)
}

func newMapPanel(mv *mapViews) *mapPanel {
	p := &mapPanel{Box: tview.NewBox(), mv: mv, styles: styleSet{reg: mv.reg}, now: time.Now}
	p.start = p.now()
	p.set = defaultMapSettings(mv.reg)
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
	// The panel leaves the command bar's rows to the dashboard underneath
	// (a Flex does not clear the cells it leaves empty), so the prompt stays
	// in sight, and in use, while the map is open.
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
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

// current is the active style's view, with the panel's flows setting.
func (p *mapPanel) current() mapstyle.Style {
	st := p.styles.get(p.set.Style)
	st.SetOption(mapstyle.OptFlows, p.flows)
	return st
}

// setFlows turns the flows overlay on ("on"), off ("off") or over (anything
// else) and reports whether it is now on.
func (p *mapPanel) setFlows(mode string) bool {
	switch mode {
	case "on":
		p.flows = true
	case "off":
		p.flows = false
	default:
		p.flows = !p.flows
	}
	return p.flows
}

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

// drawBar is the key bar: the hint or note, what Enter does, then the
// setting commands and Esc. When it does not fit it drops what Enter does,
// then shortens the rest, so the hint and Esc always show.
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
	if p.prompt != nil && p.prompt() != "" {
		cmd = []barPart{{"Enter", 1}, {" runs the command below · ", 2}}
	} else if in, ok := st.Inspect(f); ok && in.Command != "" {
		cmd = []barPart{{"Enter", 1}, {" types ", 2}, {in.Command, 0}, {" · ", 2}}
	}
	full := []barPart{{"map style", 1}, {" " + styleTitle(p.mv.reg, p.set.Style) + " · ", 2},
		{"map glyphs", 1}, {" " + p.set.Tier.String() + " · ", 2}, {"Esc", 1}, {" close", 2}}
	short := []barPart{{"map style", 1}, {" · ", 2}, {"Esc", 1}, {" close", 2}}
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
		// A long reply is clipped so the short keys always show.
		if room := avail - width(short); len(lead) > 0 && width(lead) > room {
			lead = []barPart{{mapmodel.Clip(lead[0].text, max(0, room-2)), lead[0].role}, {"  ", 0}}
		}
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

// InputHandler is for when the panel itself has the focus (the command bar
// normally keeps it): the map's own keys work, and anything else goes to the
// command bar, which takes the keyboard back.
func (p *mapPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		prompt := ""
		if p.prompt != nil {
			prompt = p.prompt()
		}
		if !p.routeKey(ev, prompt) && p.toPrompt != nil {
			p.toPrompt(ev)
		}
	})
}

// routeKey decides who takes a key while the panel is open, given the
// command bar's text: the map (it applies the key and reports true) or the
// command bar (false). The map takes the keys that print nothing: the
// arrows, PgUp and PgDn, Home and End, and, while the prompt is empty, Tab,
// Shift-Tab and Enter. With something typed those three are the prompt's
// (completion, run), and every printable key always is.
func (p *mapPanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight,
		tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
	case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev)
	return true
}

// handleKey applies one map key and reports whether anything used it.
// Enter types the inspected command into the prompt; the rest are the
// style's. Esc is the dashboard's (it closes the panel).
func (p *mapPanel) handleKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyEnter {
		in, ok := p.current().Inspect(p.frame())
		if !ok || in.Command == "" {
			p.note = "Nothing to type here. Move the cursor onto a building or a civilization."
			return true
		}
		if p.stage != nil {
			p.stage(in.Command)
		}
		p.note = ""
		return true
	}
	used := p.current().HandleKey(ev, p.frame())
	if used {
		p.note = ""
	}
	return used
}

// barTags matches the color tags a reply may carry: the key bar prints
// plain text.
var barTags = regexp.MustCompile(`\[[a-zA-Z0-9#:,\-]*\]`)

// reply puts what a command typed with the panel open did on the key bar,
// since the log is behind the map: its first line, or that it ran.
func (p *mapPanel) reply(cmd string, res CommandResult) {
	msg, _, _ := strings.Cut(res.Message, "\n")
	msg = strings.TrimSpace(barTags.ReplaceAllString(msg, ""))
	if msg == "" {
		msg = "Ran " + cmd + "."
	}
	p.hint, p.note = false, msg
}
