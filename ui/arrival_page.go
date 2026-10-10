package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// arrival_page.go lays out and draws the arrival screen: what a player
// sees on reaching a new age, and a new epoch. arrival.go runs it.
//
// The screen has two beats. First the moment: the age's name struck in hot
// iron over the player's own town, which is redrawn from the age left
// behind into the new one. An advance that also enters a new epoch opens
// with the era's name instead, larger, after a longer build-up, with the
// whole screen lit by the strike. Then the information: what the age opens,
// in a box over the town, in the words the old splash used
// (ageSplashLines).
//
// Like the main menu it is drawn into a grid from a scene that moves on a
// fixed script with a generator of its own, so a frame is the same every
// time it is drawn, and every colour comes from the theme (menuPalette).

// ---- the script ----

// The beats, in animation frames (eight a second).
const (
	// An age: the name heats, is struck, and the town turns while it glows.
	arrAgeStrike   = 5  // the strike
	arrAgeDissolve = 12 // frames the town takes to turn
	arrAgeFrames   = 26 // the whole of it: 3.25 seconds

	// An epoch's opening: two blows to build up, then the heavy one.
	arrEraBlow1  = 4
	arrEraBlow2  = 8
	arrEraStrike = 13
	arrEraFrames = 26 // then the age's own beats follow: 6.5 seconds in all
)

// arrivalFrames is how long the celebration is, in frames.
func arrivalFrames(epoch bool) int {
	if epoch {
		return arrEraFrames + arrAgeFrames
	}
	return arrAgeFrames
}

// ---- the layout ----

// ironPlace is a name set in iron and where it sits. Without rows (ok
// false) the name does not fit in iron at this size and is set as text.
type ironPlace struct {
	text string
	rows [menuWordH]string
	ok   bool
	mode ironMode
	x, y int
	w, h int
	// tail is the name's last word where only the words before it fit in
	// iron: it stands under them in capitals, on the block's last row.
	tail string
}

// placeIron picks the largest way a name fits in room cells by rows: the
// modes are tried in order with the whole name, then with all but its last
// word (which goes under it in capitals), and one row of spaced capitals is
// the last resort.
func placeIron(text string, modes []ironMode, room, rows int) ironPlace {
	p := ironPlace{text: text}
	try := func(head, tail string, extra int) bool {
		set, ok := pixelWord(head)
		if !ok {
			return false
		}
		for _, m := range modes {
			if w, h := m.size(len(set[0])); w <= room && h+extra <= rows {
				p.rows, p.ok, p.mode, p.w, p.h, p.tail = set, true, m, w, h+extra, tail
				return true
			}
		}
		return false
	}
	if try(text, "", 0) {
		return p
	}
	if i := strings.LastIndex(text, " "); i > 0 && try(text[:i], text[i+1:], 2) {
		return p
	}
	p.w, p.h = runeLen(spaced(strings.ToUpper(text))), 1
	if p.w > room {
		p.w = runeLen(strings.ToUpper(text))
	}
	return p
}

// arrivalLayout is where the celebration's parts sit at a size.
type arrivalLayout struct {
	w, h int
	// tiny: under the game's least terminal. The name is a line of text.
	tiny bool
	// age is the age's name, with the lines under it from row under; top is
	// the first row of the town.
	age        ironPlace
	under, top int
	// era is the era's name, for an advance into a new epoch.
	era ironPlace
	// hint is the row of the line that says how to go on.
	hint int
}

// arrivalView is what the screen says.
type arrivalView struct {
	age  string // "Bronze Age"
	desc string
	quip string
	// epoch: the advance entered a new epoch, named era with its icon.
	epoch         bool
	era, eraIcon  string
	epochHeading  string // the old splash's "New epoch" line, without its rules
	passed        string // "Ages passed: Stone Age, Bronze Age", "" for none
	prompt        string
	lines         []splashLine
	plain, motion bool
}

// underLine is a line under the age's name; lead marks the first, which
// is set in the page's strongest ink.
type underLine struct {
	text string
	lead bool
}

// underLines is the lines under the age's name: its description, its quip
// and the ages passed, each wrapped to the width.
func (v *arrivalView) underLines(w int) (out []underLine) {
	for i, s := range []string{v.desc, v.quip, v.passed} {
		if s == "" {
			continue
		}
		for _, line := range wrapWords(s, max(w, 8)) {
			out = append(out, underLine{text: line, lead: i == 0})
		}
	}
	return out
}

// drawUnder draws the lines under the age's name, the first n of them.
func (L arrivalLayout) drawUnder(g *mGrid, pal *menuPalette, v *arrivalView, n int) {
	for i, line := range v.underLines(L.w - 6) {
		if i >= n {
			break
		}
		ink := pal.dim
		if line.lead {
			ink = pal.ink
		}
		g.textOn(centred(L.w, runeLen(line.text)+2), L.under+i, " "+line.text+" ", ink, pal.bg, false)
	}
}

// arrivalLayoutFor lays the celebration out at w by h.
func arrivalLayoutFor(w, h int, v *arrivalView) arrivalLayout {
	L := arrivalLayout{w: w, h: h, hint: h - 2}
	room := w - 4
	if w < menuMinW || h < menuMinH {
		L.tiny = true
		L.age = ironPlace{text: v.age, w: min(runeLen(v.age), max(w, 0)), h: 1}
		L.era = ironPlace{text: v.era, w: min(runeLen(v.era), max(w, 0)), h: 1}
		L.age.x, L.age.y = centred(w, L.age.w), max(h/3, 0)
		L.era.x, L.era.y = centred(w, L.era.w), max(h/3, 0)
		L.under, L.top = L.age.y+2, h
		return L
	}
	// The age: as wide as it goes on one line. The plain glyph set has no
	// half cells.
	modes := []ironMode{{sx: 2, sy: 1}, {sx: 1, sy: 1}}
	if !v.plain {
		modes = append(modes, ironMode{sy: 1, half: true})
	}
	L.age = placeIron(v.age, modes, room, menuWordH+2)
	L.age.x, L.age.y = centred(w, L.age.w), 2
	if h >= 36 {
		L.age.y = 3
	}
	L.under = L.age.y + L.age.h + 1
	L.top = L.under + len(v.underLines(w-6)) + 1

	// The era: twice the height where it fits, and never smaller than the
	// age would be.
	big := []ironMode{{sx: 2, sy: 2}, {sx: 1, sy: 2}}
	if !v.plain {
		big = append(big, ironMode{sy: 2, half: true})
	}
	L.era = placeIron(v.era, append(big, modes...), room, h-8)
	L.era.x, L.era.y = centred(w, L.era.w), max((h-L.era.h)/2-1, 3)
	return L
}

// ---- the scene ----

// arrivalScene is the state of what moves: the clock, the heat, the
// strike, the flash, the sparks, the stars and how far the town has turned.
type arrivalScene struct {
	L     arrivalLayout
	epoch bool
	rnd   menuRand
	frame int
	t     float64
	// strike flares the name; flash lights the whole screen (an epoch's
	// heavy blow); heat is how far the name has come up from cold, 0 to 1.
	strike, flash, heat float64
	// turn is how far the town has turned from the old age to the new,
	// 0 to 1; noise is each cell's place in that turn.
	turn   float64
	noise  []float64
	sparks menuSparks
	stars  []menuStar
}

// newArrivalScene makes the scene at its first frame.
func newArrivalScene(L arrivalLayout, epoch bool) *arrivalScene {
	sc := &arrivalScene{L: L, epoch: epoch, rnd: menuRand{s: 0x2545f4914f6cdd1d}}
	if L.tiny {
		sc.turn, sc.heat = 1, 1
		return sc
	}
	sc.stars = newMenuStars(&sc.rnd, L.w, L.top)
	sc.noise = make([]float64, L.w*L.h)
	for i := range sc.noise {
		sc.noise[i] = sc.rnd.f()
	}
	return sc
}

// inEra reports whether the scene is in an epoch's opening: the era's
// name has the screen.
func (sc *arrivalScene) inEra() bool { return sc.epoch && sc.frame < arrEraFrames }

// name is the name on screen now.
func (sc *arrivalScene) name() *ironPlace {
	if sc.inEra() {
		return &sc.L.era
	}
	return &sc.L.age
}

// burst throws n sparks off the name on screen.
func (sc *arrivalScene) burst(n int, lift float64) {
	p := sc.name()
	sc.sparks.emit(&sc.rnd, float64(p.x)+float64(p.w)/2, float64(p.y+p.h), n, float64(p.w)*0.95, lift)
}

// step moves the scene on one animation frame, and plays whatever the
// script has at that frame.
func (sc *arrivalScene) step() {
	sc.frame++
	if sc.L.tiny {
		return
	}
	f := sc.frame
	if sc.epoch {
		switch {
		case f == arrEraBlow1:
			sc.strike = 0.4
			sc.burst(14, 0.6)
		case f == arrEraBlow2:
			sc.strike = 0.62
			sc.burst(26, 0.75)
		case f == arrEraStrike:
			// The heavy blow: the name at white heat, the whole screen lit,
			// sparks the width of it.
			sc.strike, sc.flash = 1.3, 1
			sc.burst(90, 1.1)
			sc.sparks.emit(&sc.rnd, float64(sc.L.w)/2, float64(sc.L.h)*0.6, 70, float64(sc.L.w), 1.2)
		case f == arrEraFrames:
			// The era gives way to the age: its sparks go with it.
			sc.strike, sc.heat, sc.sparks.a = 0, 0, nil
		}
		if f < arrEraFrames {
			sc.heat = min(1, float64(f)/float64(arrEraStrike))
		}
		f -= arrEraFrames
	}
	if f >= 0 {
		switch {
		case f < arrAgeStrike:
			sc.heat = float64(f) / float64(arrAgeStrike)
		case f == arrAgeStrike:
			sc.heat, sc.strike = 1, 1
			sc.burst(60, 0.85)
		}
		if f > arrAgeStrike && sc.turn < 1 {
			sc.turn = min(1, float64(f-arrAgeStrike)/float64(arrAgeDissolve))
		}
	}
	for i := 0; i < menuSub; i++ {
		sc.t += 0.071 * menuPace
		sc.strike *= 0.938
		sc.flash *= 0.80
		sc.sparks.step(&sc.rnd)
		// Embers off the town, once there is a fire to rise from.
		if sc.heat >= 1 && !sc.inEra() && sc.rnd.f() < 0.5*menuPace {
			sc.sparks.emit(&sc.rnd, sc.rnd.f()*float64(sc.L.w), float64(sc.L.top+1)+sc.rnd.f()*3, 1, 2, 0.5)
		}
	}
}

// settle puts the scene where the celebration ends, for the information
// to be drawn over: the age's name glowing, the town turned. It is where
// the script arrives by itself; a key that moves on early jumps here.
func (sc *arrivalScene) settle() {
	if sc.L.tiny {
		return
	}
	if sc.inEra() {
		sc.sparks.a = nil // the era's sparks go with its name
		sc.strike = 0
	}
	sc.frame = max(sc.frame, arrivalFrames(sc.epoch))
	sc.heat, sc.turn, sc.flash = 1, 1, 0
}

// rest puts the scene at the frame a still screen holds (the motion setting
// off): the name struck and glowing, sparks in the air, the town turned.
// For an epoch it is the era's name, with the screen still lit.
func (sc *arrivalScene) rest() {
	if sc.L.tiny {
		return
	}
	sc.frame = arrivalFrames(sc.epoch)
	if sc.epoch {
		sc.frame = arrEraStrike + 3
	}
	sc.heat, sc.turn, sc.strike, sc.t = 1, 1, 0.55, 4
	sc.sparks.a = nil
	if sc.epoch {
		sc.flash = 0.3
		sc.burst(70, 1.0)
	} else {
		sc.burst(46, 0.85)
	}
	for i := 0; i < 9; i++ {
		sc.sparks.step(&sc.rnd)
	}
}

// ---- drawing ----

// heatOf is how hot a pixel of the name is now. Before the strike the
// iron is coming up from cold; after it, it glows as the menu's wordmark
// does.
func (sc *arrivalScene) heatOf(p *ironPlace) func(px, py int) float64 {
	w := len(p.rows[0])
	cold := 0.62 * (1 - sc.heat)
	return func(px, py int) float64 {
		return ironHeat(px, py, w, sc.t, sc.strike, 0.5) - cold
	}
}

// drawName draws a name in iron, or as a line of capitals where it does
// not fit in iron.
func (sc *arrivalScene) drawName(g *mGrid, pal *menuPalette, p *ironPlace, plain bool) {
	if p.ok {
		heat := sc.heatOf(p)
		drawIron(g, pal, p.rows, p.x, p.y, p.mode, heat, plain)
		if p.tail != "" {
			tail := spaced(strings.ToUpper(p.tail))
			g.text(p.x+(p.w-runeLen(tail))/2, p.y+p.h-1, tail, rampAt(&pal.heat, heat(len(p.rows[0])/2, menuWordH/2)), true)
		}
		return
	}
	text := strings.ToUpper(p.text)
	if runeLen(spaced(text)) <= p.w {
		text = spaced(text)
	}
	g.text(p.x, p.y, truncate(text, max(g.w-p.x, 0)), rampAt(&pal.heat, 0.5+0.5*sc.heat), true)
}

// light lights every cell of the page by k, 0 to 1: the flash of an
// epoch's heavy blow.
func light(g *mGrid, pal *menuPalette, k float64) {
	if k <= 0.01 {
		return
	}
	glow, hot := rampAt(&pal.fire, 0.62), rampAt(&pal.heat, 1)
	for i := range g.c {
		c := &g.c[i]
		bg := pal.bg
		if c.hasBg {
			bg = c.bg
		}
		c.bg, c.hasBg = theme.Mix(bg, glow, 0.6*k), true
		c.fg = theme.Mix(c.fg, hot, 0.45*k)
	}
}

// drawTowns draws the town in the land of the page: the old age's, the new
// age's, or part of each while it turns. A town that is missing (no state
// to draw it from) is left as sky.
func (sc *arrivalScene) drawTowns(g *mGrid, pal *menuPalette, old, cur *menuTown, mf mapstyle.Frame, flare bool) {
	L := sc.L
	lh := L.h - L.top
	switch {
	case lh <= 0:
	case sc.turn >= 1:
		cur.draw(g, pal, 0, L.top, L.w, lh, mf, flare)
	case sc.turn <= 0:
		old.draw(g, pal, 0, L.top, L.w, lh, mf, flare)
	default:
		// Each cell has its place in the turn (noise): it shows the old
		// town until the turn reaches it, and the new one after.
		a, b := newMGrid(L.w, lh), newMGrid(L.w, lh)
		old.draw(a, pal, 0, 0, L.w, lh, mf, flare)
		cur.draw(b, pal, 0, 0, L.w, lh, mf, flare)
		for y := 0; y < lh; y++ {
			for x := 0; x < L.w; x++ {
				src := a
				if sc.noise[(L.top+y)*L.w+x] < sc.turn {
					src = b
				}
				if c := src.c[y*L.w+x]; c.r != ' ' || c.hasBg {
					g.c[(L.top+y)*g.w+x] = c
				}
			}
		}
	}
}

// dim takes the cells of a block k of the way to the page: the town under
// an era's name stands back.
func dim(g *mGrid, pal *menuPalette, x, y, w, h int, k float64) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if !g.in(i, j) {
				continue
			}
			c := &g.c[j*g.w+i]
			c.fg = theme.Mix(c.fg, pal.bg, k)
			if c.hasBg {
				c.bg = theme.Mix(c.bg, pal.bg, k)
			}
		}
	}
}

// renderMoment draws the celebration's frame: the era's name or the age's.
func (sc *arrivalScene) renderMoment(pal *menuPalette, v *arrivalView, old, cur *menuTown, mf mapstyle.Frame) *mGrid {
	L := sc.L
	g := newMGrid(L.w, L.h)
	if L.tiny {
		sc.drawName(g, pal, sc.name(), v.plain)
		if y := sc.name().y + 2; y < g.h && runeLen(v.prompt) <= g.w {
			g.text(centred(g.w, runeLen(v.prompt)), y, v.prompt, pal.dim, false)
		}
		return g
	}
	drawMenuStars(g, pal, sc.stars, sc.t)
	flare := sc.strike > 0.45
	if sc.inEra() {
		// The era's name has the screen: the town it found stands back,
		// and clear of the name where the name reaches down into it.
		p := &L.era
		if lh := L.h - L.top; lh > 0 {
			land := newMGrid(L.w, lh)
			old.draw(land, pal, 0, 0, L.w, lh, mf, flare)
			dim(land, pal, 0, 0, L.w, lh, 0.55)
			for y := max(p.y+p.h+1-L.top, 0); y < lh; y++ {
				for x := 0; x < L.w; x++ {
					if c := land.c[y*L.w+x]; c.r != ' ' || c.hasBg {
						g.c[(L.top+y)*g.w+x] = c
					}
				}
			}
		}
		light(g, pal, sc.flash)
		sc.sparks.draw(g, pal)
		sc.drawName(g, pal, p, v.plain)
		if head := v.epochHeading; head != "" && runeLen(head) <= L.w-4 && p.y >= 2 {
			g.text(centred(L.w, runeLen(head)), p.y-2, head, pal.accent, true)
		}
		// A still screen holds this frame, so it names the age too.
		if !v.motion && p.y+p.h+1 < L.hint {
			line := truncate(v.age, L.w-4)
			g.text(centred(L.w, runeLen(line)), p.y+p.h+1, line, pal.ink, true)
		}
	} else {
		sc.drawTowns(g, pal, old, cur, mf, flare)
		sc.sparks.draw(g, pal)
		p := &L.age
		sc.drawName(g, pal, p, v.plain)
		if v.epoch {
			// Which era this age opens, over its name.
			era := strings.TrimSpace(v.eraIcon + " " + v.era)
			if v.plain {
				era = plainSafe(era)
			}
			if runeLen(era) <= L.w-4 {
				g.text(centred(L.w, runeLen(era)), p.y-2, era, pal.accent, true)
			}
		}
		L.drawUnder(g, pal, v, L.h)
	}
	if hint := " " + v.prompt + " "; runeLen(hint) <= L.w-4 {
		g.textOn(L.w-2-runeLen(hint), L.hint, hint, pal.groundDim, pal.ground, false)
	}
	return g
}

// ---- the information ----

// infoSeg is a run of text in one ink.
type infoSeg struct {
	text string
	ink  tcell.Color
	bold bool
}

// infoLine is a line of the information: its runs, to be centred.
type infoLine struct {
	segs []infoSeg
	kind splashKind
}

func (l infoLine) width() (n int) {
	for _, s := range l.segs {
		n += runeLen(s.text)
	}
	return n
}

func (l infoLine) plain() string { return plainOf(l.segs) }

// tagRun matches a tview colour tag: a foreground, and after colons a
// background and attributes.
var tagRun = regexp.MustCompile(`\[([a-zA-Z#][a-zA-Z0-9#]*|-)?(?::([a-zA-Z#][a-zA-Z0-9#]*|-)?)?(?::([a-zA-Z-]*))?\]`)

// parseSplashLine turns a line of the splash's markup into runs of ink on
// ground. A colour is a name the theme owns (the role names and the old
// aliases); anything else in brackets is text.
func parseSplashLine(text string, pal *menuPalette, ground tcell.Color) []infoSeg {
	names := theme.TagNames()
	base := theme.Legible(pal.ink, ground, 7)
	ink, bold := base, false
	var out []infoSeg
	emit := func(s string) {
		if s != "" {
			out = append(out, infoSeg{text: s, ink: ink, bold: bold})
		}
	}
	at := 0
	for _, m := range tagRun.FindAllStringSubmatchIndex(text, -1) {
		sub := func(i int) (string, bool) {
			if m[2*i] < 0 {
				return "", false
			}
			return text[m[2*i]:m[2*i+1]], true
		}
		fg, hasFg := sub(1)
		attrs, hasAttrs := sub(3)
		role, known := names[fg]
		if m[1]-m[0] == 2 || hasFg && fg != "-" && !known {
			continue // "[]" or a word in brackets: text, not a tag
		}
		emit(text[at:m[0]])
		at = m[1]
		switch {
		case hasFg && fg == "-":
			ink = base
		case hasFg:
			ink = theme.Legible(theme.Color(role), ground, 3)
		}
		if hasAttrs {
			bold = strings.Contains(attrs, "b")
		}
	}
	emit(text[at:])
	return out
}

// wrapInfo breaks a line at spaces into lines of at most w cells, each run
// keeping its ink.
func wrapInfo(l infoLine, w int) []infoLine {
	if l.width() <= w || w < 8 {
		return []infoLine{l}
	}
	type cell struct {
		r    rune
		ink  tcell.Color
		bold bool
	}
	var cells []cell
	for _, s := range l.segs {
		for _, r := range s.text {
			cells = append(cells, cell{r, s.ink, s.bold})
		}
	}
	build := func(cs []cell) infoLine {
		for len(cs) > 0 && cs[0].r == ' ' {
			cs = cs[1:]
		}
		for len(cs) > 0 && cs[len(cs)-1].r == ' ' {
			cs = cs[:len(cs)-1]
		}
		out := infoLine{kind: l.kind}
		for _, c := range cs {
			if n := len(out.segs); n > 0 && out.segs[n-1].ink == c.ink && out.segs[n-1].bold == c.bold {
				out.segs[n-1].text += string(c.r)
			} else {
				out.segs = append(out.segs, infoSeg{string(c.r), c.ink, c.bold})
			}
		}
		return out
	}
	var out []infoLine
	for len(cells) > w {
		cut := w
		for cut > 0 && cells[cut].r != ' ' {
			cut--
		}
		if cut == 0 {
			cut = w // a word longer than the line: cut it
		}
		out = append(out, build(cells[:cut]))
		cells = cells[cut:]
		for len(cells) > 0 && cells[0].r == ' ' {
			cells = cells[1:]
		}
	}
	if len(cells) > 0 {
		out = append(out, build(cells))
	}
	return out
}

// plainSafe drops from s every glyph the plain glyph set has no shape for
// (an era's icon, say), with the blank after it, so that it is left out
// and not drawn as a question mark.
func plainSafe(s string) string {
	var sb strings.Builder
	skip := false
	for _, r := range s {
		if skip && r == ' ' {
			skip = false
			continue
		}
		skip = false
		if r > 0x7e && mapmodel.Fold(r, mapmodel.TierASCII) == '?' {
			skip = true
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// plainOf is a line's text without its inks.
func plainOf(segs []infoSeg) string {
	var sb strings.Builder
	for _, s := range segs {
		sb.WriteString(s.text)
	}
	return sb.String()
}

// infoLines lays the splash's lines out for a box cw cells wide.
//
// head keeps the heading (the age's name, description and quip) in the
// text; without it the name stands over the box in iron. tight is how hard
// the text is packed, 0 to 4: the buildings now legacy on one line, then
// the upgrades in two columns, then no blank lines, then no rules round
// the heading.
func infoLines(v *arrivalView, pal *menuPalette, cw int, head bool, tight int) []infoLine {
	ground := pal.ground
	strong, soft := theme.Legible(pal.ink, ground, 7), theme.Legible(pal.dim, ground, 3)
	var out []infoLine
	add := func(kind splashKind, segs ...infoSeg) {
		out = append(out, wrapInfo(infoLine{segs: segs, kind: kind}, cw)...)
	}
	// The upgrades and the legacy buildings are gathered and set together.
	var upgrades, legacy []string
	flush := func() {
		if n := len(upgrades); n > 0 {
			rowW := runeLen(upgrades[0])
			if tight >= 2 && n > 1 && 2*rowW+4 <= cw {
				half := (n + 1) / 2
				for i := 0; i < half; i++ {
					row := upgrades[i] + strings.Repeat(" ", rowW+4)
					if i+half < n {
						row = upgrades[i] + "    " + upgrades[i+half]
					}
					add(skUpgradeRow, infoSeg{text: row, ink: strong})
				}
			} else {
				for _, row := range upgrades {
					add(skUpgradeRow, infoSeg{text: row, ink: strong})
				}
			}
			upgrades = nil
		}
		if len(legacy) > 0 {
			if tight >= 1 {
				add(skLegacy, infoSeg{text: strings.Join(legacy, ", "), ink: soft})
			} else {
				for _, name := range legacy {
					add(skLegacy, infoSeg{text: name, ink: soft})
				}
			}
			legacy = nil
		}
	}
	passed := v.passed == ""
	for _, l := range v.lines {
		segs := parseSplashLine(l.text, pal, ground)
		if v.plain {
			for i := range segs {
				segs[i].text = plainSafe(segs[i].text)
			}
		}
		switch l.kind {
		case skUpgradeRow:
			upgrades = append(upgrades, plainOf(segs))
			continue
		case skLegacy:
			legacy = append(legacy, strings.TrimSpace(plainOf(segs)))
			continue
		}
		flush()
		switch l.kind {
		case skRule, skTitle, skDesc, skQuip:
			if !head || l.kind == skRule && tight >= 4 {
				continue
			}
		case skBlank:
			if tight < 3 && len(out) > 0 && out[len(out)-1].kind != skBlank {
				out = append(out, infoLine{kind: skBlank})
			}
			continue
		default:
			// The ages passed on the way come first of what the age brought.
			if !passed {
				passed = true
				add(skBody, infoSeg{text: v.passed, ink: soft})
			}
		}
		// The text is centred: a line's own leading blanks are dropped.
		if len(segs) > 0 {
			segs[0].text = strings.TrimLeft(segs[0].text, " ")
		}
		add(l.kind, segs...)
	}
	flush()
	for len(out) > 0 && out[0].kind == skBlank {
		out = out[1:]
	}
	return out
}

// infoTightest is the hardest infoLines packs the text.
const infoTightest = 4

// infoLayout is the information as it fits a screen: the pages of lines,
// the box they sit in, and whether the age's name stands over it in iron.
type infoLayout struct {
	pages      [][]infoLine
	x, y, w, h int
	named      bool
}

// arrivalInfoFor fits the information to the screen. It is tried under
// the name in iron first, packed tighter step by step; then in a box of
// its own the height of the screen; and when even that does not hold it,
// it is set over as many pages as it needs. Nothing is left out.
func arrivalInfoFor(L arrivalLayout, v *arrivalView, pal *menuPalette) infoLayout {
	cw := min(L.w-8, 78)
	if L.tiny {
		cw = max(L.w-2, 1)
	}
	place := func(lines []infoLine, top, rows int, named bool) infoLayout {
		wide := 0
		for _, l := range lines {
			wide = max(wide, l.width())
		}
		out := infoLayout{pages: [][]infoLine{lines}, named: named}
		out.w, out.h = min(max(wide, 30)+6, L.w), len(lines)+2
		out.x, out.y = centred(L.w, out.w), top+max((rows-out.h)/2, 0)
		return out
	}
	if !L.tiny {
		for tight := 0; tight <= infoTightest; tight++ {
			if lines := infoLines(v, pal, cw, false, tight); len(lines)+2 <= L.h-L.top-1 {
				return place(lines, L.top, L.h-L.top-1, true)
			}
		}
	}
	var lines []infoLine
	for tight := 0; tight <= infoTightest; tight++ {
		if lines = infoLines(v, pal, cw, true, tight); len(lines)+2 <= L.h {
			return place(lines, 0, L.h, false)
		}
	}
	// Pages. The text is set with its blank lines again, and a page ends
	// at one where it can, so that what belongs together stays together.
	lines = infoLines(v, pal, cw, true, 2)
	per := max(L.h-2, 1)
	out := infoLayout{}
	wide := 0
	for len(lines) > 0 {
		for len(lines) > 0 && lines[0].kind == skBlank {
			lines = lines[1:]
		}
		n := min(per, len(lines))
		if n < len(lines) {
			for cut := n; cut > per/2; cut-- {
				if lines[cut].kind == skBlank {
					n = cut
					break
				}
			}
		}
		page := lines[:n]
		lines = lines[n:]
		for len(page) > 0 && page[len(page)-1].kind == skBlank {
			page = page[:len(page)-1]
		}
		if len(page) > 0 {
			out.pages = append(out.pages, page)
		}
		for _, l := range page {
			wide = max(wide, l.width())
		}
	}
	out.w, out.h = min(max(wide, 30)+6, L.w), min(per+2, L.h)
	out.x, out.y = centred(L.w, out.w), max((L.h-out.h)/2, 0)
	return out
}

// renderInfo draws the information's page over the scene: the town in the
// new age, the name over it where it fits, and the box.
func (sc *arrivalScene) renderInfo(pal *menuPalette, v *arrivalView, info infoLayout, page int, cur *menuTown, mf mapstyle.Frame) *mGrid {
	L := sc.L
	g := newMGrid(L.w, L.h)
	if !L.tiny {
		drawMenuStars(g, pal, sc.stars, sc.t)
		cur.draw(g, pal, 0, L.top, L.w, L.h-L.top, mf, false)
		sc.sparks.draw(g, pal)
		if info.named {
			sc.drawName(g, pal, &L.age, v.plain)
			L.drawUnder(g, pal, v, L.top-L.under-1)
		}
	}
	if len(info.pages) == 0 {
		return g
	}
	page = min(max(page, 0), len(info.pages)-1)
	lines := info.pages[page]
	g.fill(info.x, info.y, info.w, info.h, pal.groundInk, pal.ground)
	g.box(info.x, info.y, info.w, info.h, false, pal.groundDim)
	y := info.y + 1 + max((info.h-2-len(lines))/2, 0)
	for _, l := range lines {
		x := info.x + centred(info.w, l.width())
		for _, s := range l.segs {
			x += g.text(x, y, s.text, s.ink, s.bold)
		}
		y++
	}
	if len(info.pages) > 1 {
		mark := fmt.Sprintf(" %d/%d ", page+1, len(info.pages))
		g.text(info.x+info.w-runeLen(mark)-2, info.y+info.h-1, mark, pal.groundDim, false)
	}
	return g
}
