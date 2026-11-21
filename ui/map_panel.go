package ui

import (
	"image"
	"regexp"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// map_panel.go is the Map panel: the map over the whole screen but the
// command bar, opened with map (aliases citymap and worldmap). The command
// bar keeps the keyboard while the panel is open, so commands work as they
// do anywhere: printable keys go to the prompt, and the map takes only the
// keys that print nothing (routeKey): the arrows, Tab and Shift-Tab, PgUp
// and PgDn, Home and End, and Enter on an empty prompt, which types the
// inspected command. On an empty prompt it also takes + - = to zoom. Glyphs
// and the flows overlay are commands (map glyphs, map flows). The map draws
// everything but the panel's bottom row, a key bar.

// mapAnimStep is the map animation frame length (about 8 frames a second).
const mapAnimStep = 125 * time.Millisecond

// animFrame is the animation frame after since: the clock's when motion is
// on, and the first frame, held, when the motion setting is off.
func animFrame(motion bool, since time.Duration) int {
	if !motion {
		return 0
	}
	return worldFrame(since)
}

// worldFrame is the map's clock after since, in animation frames. It runs
// whatever the motion setting says: the picture can hold still, but what
// happens in the world (a visitor arriving) still happens.
func worldFrame(since time.Duration) int { return int(since / mapAnimStep) }

// mapFrame is the frame a map draws after since.
func mapFrame(m *mapmodel.Model, set mapSettings, since time.Duration) mapstyle.Frame {
	return mapstyle.Frame{Model: m, Anim: animFrame(set.Motion, since), Clock: worldFrame(since), Tier: set.Tier}
}

// mapViews is shared by the Map panel and the mini map: one model builder
// (the UI goroutine owns it) and the since-last-visit baseline.
type mapViews struct {
	builder *mapmodel.Builder

	sinceKey [2]int64
	since    *mapmodel.Visit
}

func newMapViews() *mapViews {
	return &mapViews{builder: mapmodel.NewBuilder(nil)}
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

// mapPanel is the Map overlay primitive.
type mapPanel struct {
	barPanel
	kit.Base
	mv    *mapViews
	view  mapstyle.Style
	set   mapSettings
	model *mapmodel.Model
	start time.Time
	now   func() time.Time // the animation clock (tests pin it)

	world bool // the next open starts on the known world
	flows bool // the flows overlay (map flows)
	note  string

	// settings reads the current settings.
	settings func() mapSettings
	// space is the window's own sky, when there is a window: the panel hands
	// it the world instead of drawing space in cells.
	space mapstyle.SpaceHost
	// stage puts a command in the prompt, unrun, for the player's Enter.
	stage func(cmd string)
	// spotted is told when the cursor is on the rare visitor, once a
	// visit; onVisitor is whether it was there at the last look.
	spotted func(kind string)
	// settle is told which uncertain places the player settled on the map.
	settle    func(ids []string)
	onVisitor bool
	// keys are the key prompts of the bar as last drawn, and row the bar's
	// text, which they read their caps from.
	keys keySet
	row  []rune
}

func newMapPanel(mv *mapViews) *mapPanel {
	p := &mapPanel{mv: mv, view: mapstyle.New(), now: time.Now}
	p.barPanel = newBarPanel(p.routeKey)
	p.start = p.now()
	p.set = defaultMapSettings()
	return p
}

// open is the overlay's build: it runs on every Show. It picks up the
// settings and points the map at the settlement or the world.
func (p *mapPanel) open(state game.GameState) Primitive {
	if p.settings != nil {
		p.set = p.settings()
	}
	p.note = ""
	p.current().SetOption(mapstyle.OptWorld, p.world)
	p.world = false
	p.update(state)
	return overBar(p)
}

// update is the overlay's refresh (every UI tick while open).
func (p *mapPanel) update(state game.GameState) {
	if p.settings != nil {
		p.set = p.settings()
	}
	p.model = p.mv.model(&state)
	p.lookForVisitor()
	if p.settle != nil && p.space != nil {
		if ids := p.space.TakeSettled(); len(ids) > 0 {
			p.settle(ids)
		}
	}
}

// lookForVisitor tells the game when the cursor stands on the rare visitor
// (mapstyle.KindAlien), once each time it comes to stand there. It runs on
// every refresh while the panel is open and after every key, never from
// Draw: the game takes its own lock to hear it.
func (p *mapPanel) lookForVisitor() {
	in, ok := p.current().Inspect(p.frame())
	on := ok && in.Kind == mapstyle.KindAlien
	if on && !p.onVisitor && p.spotted != nil {
		p.spotted(in.Kind)
	}
	p.onVisitor = on
}

// current is the map view, with the panel's flows setting.
func (p *mapPanel) current() mapstyle.Style {
	p.view.SetOption(mapstyle.OptFlows, p.flows)
	return p.view
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
	f := mapFrame(p.model, p.set, p.now().Sub(p.start))
	f.Space = p.space
	return f
}

// hide tells the window's sky the panel is gone.
func (p *mapPanel) hide() {
	if p.space != nil {
		p.space.Hide()
	}
}

func (p *mapPanel) Layout(r image.Rectangle) { p.Place(r) }

// wireMapPanel makes the key prompts of the map's bar pressable.
func (d *Dashboard) wireMapPanel() { d.mapPanel.buttons.wire(d.app) }

// Paint draws the map over all but the last row, and the key bar on it. The
// map still writes cells through mapstyle.Cells, at the panel's place on the screen: scr
// puts them on the painter.
func (p *mapPanel) Paint(kp *kit.Painter) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 || p.view == nil {
		return
	}
	scr := cellsFor(kp, p, p.Bounds())
	st, f := p.current(), p.frame()
	if h < 3 {
		st.Draw(scr, mapstyle.Rect{X: x, Y: y, W: w, H: h}, f)
		return
	}
	st.Draw(scr, mapstyle.Rect{X: x, Y: y, W: w, H: h - 1}, f)
	p.drawBar(scr, mapstyle.Rect{X: x, Y: y + h - 1, W: w, H: 1}, st, f)
	pal := newExpPalette()
	pal.styles[xText] = pal.styles[xText].Background(theme.Of(theme.RoleChip))
	p.buttons.paint(kp, &p.keys, p.Bounds().Min, pal)
}

// Mouse: a click on a key prompt of the bar presses its key.
func (p *mapPanel) Mouse(m kit.Mouse) bool { return p.buttons.mouse(&p.keys, p.Bounds().Min, m) }

// PointerAt: the hand over a key prompt of the bar.
func (p *mapPanel) PointerAt(at image.Point) kit.Pointer {
	return p.buttons.pointer(&p.keys, p.Bounds().Min, at)
}

// barPart is one run of key-bar text in one style.
type barPart struct {
	text string
	role int // 0 text, 1 key, 2 dim
	// word, when set, makes the part a key prompt: text is the key.
	word string
}

// drawBar is the key bar: the hint or note, what Enter does, then the
// setting commands and Esc. When it does not fit it drops what Enter does,
// then shortens the rest, so the hint and Esc always show.
func (p *mapPanel) drawBar(scr mapstyle.Cells, r mapstyle.Rect, st mapstyle.Style, f mapstyle.Frame) {
	bg := theme.Of(theme.RoleChip)
	base := kit.Style{}.Background(bg).Foreground(theme.Legible(theme.Of(theme.RoleText), bg, 4.5))
	styles := [3]kit.Style{
		base,
		base.Foreground(theme.Legible(theme.Of(theme.RoleAccent), bg, 3)).Bold(true),
		base.Foreground(theme.Legible(theme.Of(theme.RoleDim), bg, 3)),
	}
	var lead, cmd []barPart
	if p.note != "" {
		lead = []barPart{{p.note, 1, ""}, {"  ", 0, ""}}
	}
	if p.prompt != nil && p.prompt() != "" {
		word, sep := rowid.Line("ui.map_panel.drawBar.02", " runs the command below · "), ""
		if i := strings.LastIndex(word, "·"); i >= 0 {
			word, sep = word[:i], word[i:]
		}
		cmd = []barPart{{"Enter", 1, strings.TrimSpace(word)}, {sep, 2, ""}}
	} else if in, ok := st.Inspect(f); ok && in.Command != "" {
		cmd = []barPart{{"Enter", 1, strings.TrimSpace(rowid.Line("ui.map_panel.drawBar.04", " types "))}, {in.Command, 0, ""}, {" · ", 2, ""}}
	}
	full := []barPart{{"map glyphs", 1, ""}, {" " + p.set.Tier.String() + " · ", 2, ""}, {"Esc", 1, strings.TrimSpace(rowid.Line("ui.map_panel.drawBar.06", " close"))}}
	short := []barPart{{"map glyphs", 1, ""}, {" · ", 2, ""}, {"Esc", 1, strings.TrimSpace(rowid.Line("ui.map_panel.drawBar.08", " close"))}}
	width := func(parts ...[]barPart) int {
		n := 0
		for _, ps := range parts {
			for _, pt := range ps {
				if pt.word != "" {
					n += keyButtonW(pt.text, pt.word)
				} else {
					n += mapstyle.TextLen(pt.text)
				}
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
			lead = []barPart{{mapmodel.Clip(lead[0].text, max(0, room-2)), lead[0].role, ""}, {"  ", 0, ""}}
		}
		parts = [][]barPart{lead, short}
	}
	cv := mapstyle.NewCanvas(scr, r, f.Tier, base)
	p.row = make([]rune, r.W)
	// The close button sits in the panel's top right cell, as in a framed
	// panel; its glyphs come from the top row, the bar's from the bottom one.
	cx, closeRunes := r.W-closeButtonCols, []rune(" × ")
	p.keys = keySet{glyph: func(x, y int) rune {
		if y == 0 && x >= cx && x-cx < len(closeRunes) {
			return closeRunes[x-cx]
		}
		return p.row[x]
	}}
	p.keys.closeButton(r.W, func(int, int, string, bool) {})
	// The key prompts are recorded on the bar's row, which is the panel's last.
	_, _, _, ph := p.GetRect()
	// put writes text at x on the bar and keeps it in row, where the key
	// prompts read their caps from.
	put := func(x int, text string, st kit.Style) int {
		end := cv.Text(x, 0, max(r.W-x, 0), text, st)
		for i, c := range []rune(text) {
			if x+i < min(end, len(p.row)) {
				p.row[x+i] = c
			}
		}
		return end
	}
	write := func(x, _ int, text string, cap bool) {
		st := styles[2]
		if cap {
			st = styles[1]
		}
		put(x, text, st)
	}
	x := 1
	for _, ps := range parts {
		for _, pt := range ps {
			switch {
			case x >= r.W:
			case pt.word != "":
				x = p.keys.keyButton(x, ph-1, pt.text, pt.word, write)
			default:
				x = put(x, pt.text, styles[pt.role])
			}
		}
	}
}

// routeKey decides who takes a key while the panel is open, given the
// command bar's text: the map (it applies the key and reports true) or the
// command bar (false). The map takes the keys that print nothing: the
// arrows, PgUp and PgDn, Home and End, and, while the prompt is empty, Tab,
// Shift-Tab and Enter. With something typed those three are the prompt's
// (completion, run), and every printable key always is, but for the zoom
// keys + - = _ on an empty prompt (no command begins with one). + and =
// zoom in and - and _ zoom out, as PgDn and PgUp do.
func (p *mapPanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight,
		kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome, kit.KeyEnd:
	case kit.KeyTab, kit.KeyBacktab, kit.KeyEnter:
		if prompt != "" {
			return false
		}
	case kit.KeyRune:
		// + and = are PgDn, - and _ are PgUp, so the map zooms with no
		// modifier and no Page keys (a MacBook has none).
		k, zoomKey := zoomRuneKeys[ev.Rune]
		if !zoomKey || prompt != "" || ev.Mod&(kit.ModCtrl|kit.ModAlt) != 0 {
			return false
		}
		return p.handleKey(newKey(k, 0, 0))
	default:
		return false
	}
	p.handleKey(ev)
	return true
}

// zoomRuneKeys are the page keys the zoom characters stand for.
var zoomRuneKeys = map[rune]kit.KeyCode{'+': kit.KeyPgDn, '=': kit.KeyPgDn, '-': kit.KeyPgUp, '_': kit.KeyPgUp}

// handleKey applies one map key and reports whether anything used it.
// Enter types the inspected command into the prompt; the rest are the
// map's. Esc is the dashboard's (it closes the panel).
func (p *mapPanel) handleKey(ev *kit.Key) bool {
	if ev.Code == kit.KeyEnter {
		in, ok := p.current().Inspect(p.frame())
		if !ok || in.Command == "" {
			p.note = rowid.Line("ui.map_panel.handleKey.01", "Nothing to type here. Move the cursor onto a building or a civilization.")
			return true
		}
		if p.stage != nil {
			p.stage(in.Command)
		}
		p.note = ""
		return true
	}
	used := p.current().HandleKey(*ev, p.frame())
	if used {
		p.note = ""
		p.lookForVisitor()
	}
	return used
}

// barTags matches the color tags a reply may carry: the key bar prints
// plain text.
var barTags = regexp.MustCompile(`\[[a-zA-Z0-9#:,\-]*\]`)

// replyLine is the first line of a command's answer without its colour tags,
// and whether the command failed.
func replyLine(res CommandResult) (msg string, bad bool) {
	msg, _, _ = strings.Cut(res.Message, "\n")
	return strings.TrimSpace(barTags.ReplaceAllString(msg, "")), res.Type == "error" || res.Type == "warning"
}

// reply puts what a command typed with the panel open did on the key bar,
// since the log is behind the map: its first line, or that it ran.
func (p *mapPanel) reply(cmd string, res CommandResult) {
	msg, _ := replyLine(res)
	if msg == "" {
		msg = rowid.Line("ui.map_panel.reply.01", "Ran ") + cmd + "."
	}
	p.note = msg
}
