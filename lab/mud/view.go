package main

// view.go lays the walk out on a tcell screen: title bar, scene and map side
// by side, the room text, a line of transcript, the prompt and the key hints.
// It is also the compact sidebar form (DrawMini). Everything is drawn through
// theme roles, so a theme switch or a light theme needs nothing special.

import (
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// Mode is what the text area shows.
type Mode int

const (
	ModeRoom Mode = iota
	ModeExamine
	ModeSurvey
	ModeAway
	ModeWorld
	ModeHelp
)

// View is the walk panel's state.
type View struct {
	City       *City
	Prev       *City // the last check-in, for "while you were away"
	Here       PlaceKey
	Mode       Mode
	ExamineKey string
	Input      string
	Echo       string // the last command, as typed
	Reply      Para   // the one-line answer to it
	History    []Para // the transcript, oldest first
	Frame      int
	Scroll     int
	Visited    map[PlaceKey]bool
}

func NewView(c *City, prev *City) *View {
	v := &View{City: c, Prev: prev, Here: Square, Visited: map[PlaceKey]bool{Square: true}}
	if prev != nil {
		v.Mode = ModeAway
	}
	return v
}

func style(role theme.Role, bold bool) tcell.Style {
	st := tcell.StyleDefault.Foreground(theme.Color(role)).Background(theme.Color(theme.RoleBackground))
	if role == theme.RoleSelectionText {
		st = st.Background(theme.Color(theme.RoleSelection))
	}
	if bold {
		st = st.Bold(true)
	}
	return st
}

// put writes s at (x,y), clipped to maxX; returns the x after it.
func put(s tcell.Screen, x, y, maxX int, text string, role theme.Role, bold bool) int {
	st := style(role, bold)
	for _, r := range text {
		if x >= maxX {
			break
		}
		s.SetContent(x, y, r, nil, st)
		x++
	}
	return x
}

func blit(s tcell.Screen, g *Grid, ox, oy int) {
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.At(x, y)
			if c.R == 0 {
				continue
			}
			s.SetContent(ox+x, oy+y, c.R, nil, style(c.Role, c.Bold))
		}
	}
}

// box draws a light border with an optional title in the top edge.
func box(s tcell.Screen, x, y, w, h int, title string) {
	b := style(theme.RoleBorder, false)
	for i := x + 1; i < x+w-1; i++ {
		s.SetContent(i, y, '─', nil, b)
		s.SetContent(i, y+h-1, '─', nil, b)
	}
	for j := y + 1; j < y+h-1; j++ {
		s.SetContent(x, j, '│', nil, b)
		s.SetContent(x+w-1, j, '│', nil, b)
	}
	s.SetContent(x, y, '╭', nil, b)
	s.SetContent(x+w-1, y, '╮', nil, b)
	s.SetContent(x, y+h-1, '╰', nil, b)
	s.SetContent(x+w-1, y+h-1, '╯', nil, b)
	if title != "" {
		put(s, x+2, y, x+w-2, " "+title+" ", theme.RoleDim, false)
	}
}

// wrap lays out a paragraph of spans into lines of at most w cells.
func wrap(p Para, w int) [][]Span {
	type word struct {
		text string
		role theme.Role
		bold bool
		sp   bool // preceded by a space
	}
	var words []word
	for _, s := range p {
		parts := strings.Split(s.Text, " ")
		for i, t := range parts {
			if t == "" {
				if i > 0 && len(words) > 0 {
					// leading/trailing space inside a span
					words = append(words, word{text: "", role: s.Role, sp: true})
				}
				continue
			}
			sp := i > 0 || strings.HasPrefix(s.Text, " ")
			words = append(words, word{text: t, role: s.Role, bold: s.Bold, sp: sp && len(words) > 0})
		}
	}
	var lines [][]Span
	var cur []Span
	n := 0
	for _, wd := range words {
		l := utf8.RuneCountInString(wd.text)
		pre := 0
		if wd.sp && n > 0 {
			pre = 1
		}
		if n+pre+l > w && n > 0 {
			lines = append(lines, cur)
			cur, n, pre = nil, 0, 0
		}
		if pre == 1 {
			cur = append(cur, Span{Text: " ", Role: wd.role})
			n++
		}
		if wd.text != "" {
			cur = append(cur, Span{Text: wd.text, Role: wd.role, Bold: wd.bold})
			n += l
		}
	}
	if len(cur) > 0 {
		lines = append(lines, cur)
	}
	return lines
}

// Body is the text area's paragraphs for the current mode.
func (v *View) Body() []Para {
	c := v.City
	p := c.Places[v.Here]
	switch v.Mode {
	case ModeExamine:
		if out := c.Examine(p, v.ExamineKey); out != nil {
			return out
		}
	case ModeSurvey:
		return c.Survey()
	case ModeWorld:
		return c.World()
	case ModeHelp:
		return helpParas(c)
	case ModeAway:
		if v.Prev != nil {
			d, ch := c.Diff(v.Prev)
			out := []Para{{{Text: "While you were away", Role: theme.RoleAccent, Bold: true}, {Text: " (" + d + ")", Role: theme.RoleDim}}}
			if len(ch) == 0 {
				out = append(out, plain("Nothing much. The place ran itself.", theme.RoleText))
			}
			for _, x := range ch {
				out = append(out, x.Text)
			}
			out = append(out, plain("Places marked * on the map have changed. Walk over and look.", theme.RoleDim))
			return out
		}
	}
	out := []Para{c.Describe(p, v.Prev), {}}
	if len(p.Holdings) > 0 || (p.Key != Square && p.Key != Gate) {
		out = append(out, c.HereLine(p))
	}
	out = append(out, c.Ledger(p)...)
	if a := c.AlsoLine(p); a != nil {
		out = append(out, a)
	}
	if v.Prev != nil {
		if since := v.sinceLine(p); since != nil {
			out = append(out, since)
		}
	}
	out = append(out, c.ExitsLine(p))
	return out
}

// sinceLine is the per-room part of the idle diff.
func (v *View) sinceLine(p *Place) Para {
	d, ch := v.City.Diff(v.Prev)
	var parts []Span
	for _, x := range ch {
		if x.Place == p.Key {
			parts = append(parts, x.Text...)
			parts = append(parts, Span{Text: "  ", Role: theme.RoleDim})
		}
	}
	if len(parts) == 0 {
		return nil
	}
	return append(Para{{Text: "Since you last looked (" + d + "): ", Role: theme.RoleLabel, Bold: true}}, parts...)
}

func helpParas(c *City) []Para {
	rows := [][2]string{
		{"look, l", "describe where you are"},
		{"go <dir|place>", c.DirWord(North) + ", " + c.DirWord(East) + ", ... or any place name: walks the road there"},
		{"n e s w", "a step in that direction (arrow keys too)"},
		{"visit <building>", "walk to the building's place and examine it (visit smithy)"},
		{"examine <building>, x", "a building's card, in this room"},
		{"survey", "every place at once, from a height"},
		{"out", "from the gate: the wide world and the civilizations in it"},
		{"away", "what changed since your last check-in"},
		{"back, q", "leave the walk (in the game: back to the dashboard)"},
	}
	out := []Para{{{Text: "Walking the city", Role: theme.RoleAccent, Bold: true}}}
	for _, r := range rows {
		out = append(out, Para{{Text: padRight(r[0], 24), Role: theme.RoleHighlight}, {Text: r[1], Role: theme.RoleText}})
	}
	return out
}

// Draw renders the full walk view into the whole screen.
func (v *View) Draw(s tcell.Screen) {
	W, H := s.Size()
	if W < 60 || H < 18 {
		// Too small for the room view: the compact map is the whole panel.
		v.DrawMini(s)
		return
	}
	s.Fill(' ', style(theme.RoleText, false))
	c := v.City
	p := c.Places[v.Here]

	// Title bar.
	x := put(s, 1, 0, W, glyphFor(v.Here)+" ", theme.RoleAccent, true)
	x = put(s, x, 0, W, capFirst(p.Name), theme.RoleAccent, true)
	x = put(s, x+2, 0, W, c.St.AgeName+" · "+c.St.EpochName, theme.RoleDim, false)
	right := c.TimeWord()
	if w := c.Weather(); w != "clear" && w != "quiet" {
		right += " · " + w
	}
	put(s, W-1-utf8.RuneCountInString(right), 0, W, right, theme.RoleLabel, false)

	// Top panes: scene | map. The map picks the widest labels that fit half.
	top := 1
	paneH := 9
	if H >= 34 {
		paneH = 11
	}
	if H < 28 {
		paneH = 7
	}
	if H < 20 {
		paneH = 5
	}
	mo := MapOpts{Gap: 2, Here: v.Here, Frame: v.Frame, Changed: c.ChangedPlaces(v.Prev)}
	for mo.LabelW = 7; mo.LabelW > 3; mo.LabelW-- {
		if mw, _ := c.mapSize(mo); mw+4 <= W*11/20 {
			break
		}
	}
	mw, mh := c.mapSize(mo)
	mapBoxW := mw + 4
	if mh+2 > paneH {
		paneH = mh + 2
	}
	sceneW := W - mapBoxW
	box(s, 0, top, sceneW, paneH, "")
	g := drawVignette(c, p, c.Scene(v.Frame), sceneW-2, paneH-2)
	blit(s, g, 1, top+1)
	box(s, sceneW, top, mapBoxW, paneH, "map")
	blit(s, c.drawMap(mo), sceneW+2, top+1+(paneH-2-mh)/2)

	// Text area.
	textTop := top + paneH + 1
	textBottom := H - 4
	tw := W - 4
	var lines [][]Span
	for _, para := range v.Body() {
		if len(para) == 0 {
			lines = append(lines, nil)
			continue
		}
		lines = append(lines, wrap(para, tw)...)
	}
	region := textBottom - textTop + 1
	// The transcript sits under the room text, newest at the bottom, like a
	// MUD client's scrollback; it gets whatever room the text leaves (at
	// least the latest reply).
	var tr [][]Span
	for _, h := range v.History {
		tr = append(tr, wrap(h, tw)...)
	}
	keep := 0
	if len(tr) > 0 {
		keep = max(2, region-len(lines)-1)
		keep = min(keep, len(tr), region/2+1)
	}
	avail := region
	if keep > 0 {
		avail = region - keep - 1
		for i := 0; i < keep; i++ {
			l := tr[len(tr)-keep+i]
			xx := 2
			for _, sp := range l {
				role := sp.Role
				if i < keep-2 {
					role = theme.RoleDim // older lines fade
				}
				xx = put(s, xx, textBottom-keep+1+i, W-2, sp.Text, role, sp.Bold && i >= keep-2)
			}
		}
	}
	if v.Scroll > len(lines)-avail {
		v.Scroll = max(0, len(lines)-avail)
	}
	for i := 0; i < avail && v.Scroll+i < len(lines); i++ {
		xx := 2
		for _, sp := range lines[v.Scroll+i] {
			xx = put(s, xx, textTop+i, W-2, sp.Text, sp.Role, sp.Bold)
		}
	}
	if len(lines)-v.Scroll > avail {
		put(s, W-4, textBottom, W, " ▼", theme.RoleDim, false)
	}

	// Transcript, prompt, hints.
	sep := style(theme.RoleBorder, false)
	for i := 0; i < W; i++ {
		s.SetContent(i, H-3, '─', nil, sep)
	}
	xx := put(s, 1, H-2, W, "> ", theme.RoleAccent, true)
	xx = put(s, xx, H-2, W, v.Input, theme.RoleBright, false)
	put(s, xx, H-2, W, "█", theme.RoleAccent, false)
	hints := []string{"←↑→↓ walk", "l look", "x examine", "survey", "out", "? help", "q back"}
	if W < 90 {
		hints = []string{"←↑→↓ walk", "l look", "survey", "? help", "q back"}
	}
	xx = 1
	for i, h := range hints {
		if i > 0 {
			xx = put(s, xx, H-1, W, "  ", theme.RoleDim, false)
		}
		k, rest, _ := strings.Cut(h, " ")
		xx = put(s, xx, H-1, W, k, theme.RoleHighlight, false)
		xx = put(s, xx, H-1, W, " "+rest, theme.RoleDim, false)
	}
}

// DrawMini renders the compact sidebar form: the map with counts, where you
// are, and the latest change. Meant for a dashboard corner (about 40×15).
func (v *View) DrawMini(s tcell.Screen) {
	W, H := s.Size()
	s.Fill(' ', style(theme.RoleText, false))
	c := v.City
	box(s, 0, 0, W, H, "")
	put(s, 2, 0, W-2, " "+capFirst(c.Places[Square].Name)+" ", theme.RoleAccent, true)
	right := " " + c.TimeWord() + " "
	put(s, W-2-utf8.RuneCountInString(right), 0, W-2, right, theme.RoleDim, false)
	mo := MapOpts{Gap: 1, Counts: true, Here: v.Here, Frame: v.Frame, Changed: c.ChangedPlaces(v.Prev)}
	for mo.LabelW = 7; mo.LabelW > 3; mo.LabelW-- {
		if mw, _ := c.mapSize(mo); mw <= W-4 {
			break
		}
	}
	mw, mh := c.mapSize(mo)
	blit(s, c.drawMap(mo), (W-mw)/2, 1)
	y := 1 + mh
	if y < H-2 {
		// The newest thing worth saying.
		var line Para
		if v.Prev != nil {
			_, ch := c.Diff(v.Prev)
			if len(ch) > 0 {
				line = append(Para{{Text: "* ", Role: theme.RoleHighlight}}, ch[0].Text...)
			}
		}
		if line == nil {
			if h := c.St.Harbinger; h != nil {
				line = Para{{Text: "? ", Role: theme.RoleWarning}, {Text: capFirst(h.Name) + " is in " + c.Places[Square].Name + ".", Role: theme.RoleText}}
			} else {
				line = Para{{Text: "@ ", Role: theme.RoleBright}, {Text: "You are in " + c.Places[v.Here].Name + ".", Role: theme.RoleText}}
			}
		}
		for i, l := range wrap(line, W-4) {
			if y+i >= H-1 {
				break
			}
			xx := 2
			for _, sp := range l {
				xx = put(s, xx, y+i, W-2, sp.Text, sp.Role, sp.Bold)
			}
		}
	}
	put(s, 2, H-1, W-2, " walk ↵ ", theme.RoleDim, false)
}
