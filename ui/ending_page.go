package ui

import (
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// ending_page.go draws the film of a run's ending (ending.go runs it). A
// run ends in a prestige or in a fall to a catastrophe, and either way the
// town is given up and a new one begins at a campfire. The film tells it in
// beats:
//
//   - the ending: the town that was built, the whole screen, burning down
//     from its edges until one ember is left;
//   - the reckoning: in the dark, what the run earned, a line at a time,
//     each landing hot with a few sparks off the ember. A verdict (the Last
//     Passage's, a catastrophe's) comes first and stands alone a while;
//   - the strike, for a prestige only: three blows build to the heaviest in
//     the game, the word for it in iron as large as the screen holds, a star
//     for every prestige so far, every cell lit;
//   - the beginning: out of the light, the first fire of the new run on its
//     land, under the forged name of the game.
//
// A fall is the same film in a lower voice: the town goes cold instead of
// burning, nothing sparks, and there is no strike.
//
// Like the arrival screen it is drawn into a grid from a scene on a fixed
// script with a generator of its own, in colours that all come from the
// theme (menuPalette), and every word on it is one the game already says.

// ---- the script ----

// The film's beats.
const (
	beatEnding = iota
	beatReckoning
	beatStrike
	beatBeginning
)

// The script, in animation frames (eight a second).
const (
	// The ending: the town stands a moment, burns, and the tail is the
	// ember's alone.
	endBurnFrames = 28
	endBurnHold   = 4
	endBurnTail   = 4

	// The reckoning: a lead before its first line, a step between two
	// lines (a shorter one when there are many), a hold after a verdict,
	// and a tail after the last line.
	endLead        = 2
	endStep        = 3
	endStepFast    = 2
	endManyLines   = 7
	endVerdictHold = 8
	endTail        = 5
	endCountUp     = 4 // frames a number takes to count up

	// The strike: three blows build up, then the heavy one.
	endBlow1        = 2
	endBlow2        = 5
	endBlow3        = 8
	endBlow         = 11
	endStrikeFrames = 24

	// The beginning.
	endDawnFrames = 20

	// With the motion setting off each beat is one still frame, and this is
	// its turn.
	endStillFrames = 20
)

// endWord is the word for the moment, set in iron at the strike.
const endWord = "Prestige"

// endSpan is a beat's place in the film.
type endSpan struct {
	beat, from, frames int
}

func (s endSpan) end() int { return s.from + s.frames }

// ---- what it says ----

// The kinds of line in the reckoning.
const (
	elVerdict = iota // what the catastrophe or the Last Passage did
	elAge            // the age the run reached
	elCount          // a number that counts up: points, badges
	elText           // a line the game wrote
)

// endLine is a line of the reckoning.
type endLine struct {
	kind int
	// text is the line as the game wrote it, tags and all, and level the
	// log category it wrote it in.
	text, level string
	// count and say are a counting line's number and how it reads at n.
	count int
	say   func(n int) string
}

// endView is what the film says.
type endView struct {
	end      game.RunEnding
	prestige bool
	age      string // the age the run ended in, by name
	voice    string // the run's closing line, "" for none
	complete string // the game's line for the prestige, set under the word
	captions []string
	lines    []endLine
	prompt   string
	plain    bool
	motion   bool
}

// endLevelTag is the colour a log category's lines take in the reckoning:
// the log's own, but for the plain lines, which are the page's ink (a
// screen of them in the log's cyan would read as the log).
func endLevelTag(level string) string {
	switch level {
	case "success":
		return "[green]"
	case "warning":
		return "[yellow]"
	case "event":
		return "[gold]"
	case "error":
		return "[red]"
	}
	return ""
}

// endViewFor works out what the film says from the record of the ending.
func endViewFor(end game.RunEnding, pal *menuPalette) endView {
	v := endView{end: end, prestige: end.Prestige(), age: game.AgeName(end.Age)}
	bare := func(text string) string {
		return strings.TrimSpace(plainOf(parseSplashLine(text, pal, pal.bg)))
	}
	for _, l := range ageSplashLines(end.Age, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{}) {
		if l.kind == skPrompt {
			v.prompt = bare(l.text)
		}
	}
	// First what came for the run, alone; then what the run reached and
	// earned; then everything else the game reports, in its order.
	var rest []endLine
	for _, l := range end.Lines {
		switch l.Kind {
		case game.EndLineVerdict:
			v.lines = append(v.lines, endLine{kind: elVerdict, text: strings.TrimSpace(l.Text), level: l.Level})
		case game.EndLineVoice:
			v.voice = bare(l.Text)
		case game.EndLineComplete:
			v.complete = bare(l.Text)
		default:
			rest = append(rest, endLine{kind: elText, text: strings.TrimSpace(l.Text), level: l.Level})
		}
	}
	v.lines = append(v.lines, endLine{kind: elAge, text: v.age})
	if v.prestige && (end.Points > 0 || end.Kind != game.RunEndSuccumbed) {
		v.lines = append(v.lines, endLine{kind: elCount, count: end.Points, say: func(n int) string {
			return textfmt.Count(n, "prestige point", "prestige points")
		}})
	}
	if end.Badges > 0 {
		v.lines = append(v.lines, endLine{kind: elCount, count: end.Badges, say: func(n int) string {
			return textfmt.Count(n, "badge", "badges") + " earned"
		}})
	}
	v.lines = append(v.lines, rest...)
	return v
}

// ---- the layout ----

// endLayout is where the film's parts sit at a size.
type endLayout struct {
	w, h int
	// tiny: under the game's least terminal. Each beat is a line of text.
	tiny bool
	hint int
	// The ending: the age's name on row 1, the run's closing line under
	// it, the land from top, and the last ember.
	voice  []string
	top    int
	ex, ey int
	// The reckoning: its lines are wrapped to lineW and stand in the rows
	// from l0 up to l1.
	lineW, l0, l1 int
	// The strike: the word, the row of the level's stars and the first row
	// of the game's line under it.
	word          ironPlace
	stars, capRow int
	caption       []string
	// The beginning: the wordmark (wsx is 0 for none), the row of the
	// town's caption, and the first row of the new land.
	wx, wy, wsx int
	dawnCap     int
	dawnTop     int
}

// endLayoutFor lays the film out at w by h.
func endLayoutFor(w, h int, v *endView) endLayout {
	L := endLayout{w: w, h: h, hint: h - 2}
	if w < menuMinW || h < menuMinH {
		L.tiny = true
		return L
	}
	L.voice = wrapWords(v.voice, w-8)
	if len(L.voice) > 2 {
		L.voice = L.voice[:2]
	}
	if v.voice == "" {
		L.voice = nil
	}
	L.top = 3 + len(L.voice)
	if len(L.voice) > 0 {
		L.top++
	}
	L.ex, L.ey = w/2, h-4

	L.lineW = min(w-8, 84)
	L.l0, L.l1 = 2, h-5

	// The word: as large as the screen holds, which is larger than an
	// era's name gets. The plain glyph set has no half cells.
	modes := []ironMode{{sx: 3, sy: 3}, {sx: 2, sy: 3}, {sx: 2, sy: 2}, {sx: 1, sy: 3}, {sx: 1, sy: 2}}
	if !v.plain {
		modes = append(modes, ironMode{sy: 2, half: true})
	}
	modes = append(modes, ironMode{sx: 1, sy: 1})
	L.caption = wrapWords(v.complete, w-8)
	if len(L.caption) > 2 {
		L.caption = L.caption[:2]
	}
	L.word = placeIron(endWord, modes, w-4, h-8-len(L.caption))
	L.word.x = centred(w, L.word.w)
	L.word.y = max(2, (h-L.word.h-3-len(L.caption))/2)
	L.stars = L.word.y + L.word.h + 1
	L.capRow = L.stars + 2

	// The beginning.
	if v.prestige {
		L.wsx = 1
		if w >= menuWordW*2+10 {
			L.wsx = 2
		}
		L.wx, L.wy = centred(w, menuWordW*L.wsx), 2
		L.dawnCap = L.wy + menuWordH + 1
	} else {
		L.dawnCap = 2
	}
	L.dawnTop = L.dawnCap + 2
	return L
}

// ---- the reckoning's rows ----

// endReck is the reckoning set at a width: every line wrapped into rows,
// and the frame of the beat each line lands at.
type endReck struct {
	rows   []infoLine // every row, top to bottom
	of     []int      // the line each row belongs to
	first  []int      // each line's first row
	at     []int      // the frame each line lands at
	silent []bool     // a verdict: it lands without a spark
	frames int        // how long the beat is
}

// endLineSegs is a line's runs of ink at count n (for a counting line).
func endLineSegs(l endLine, n int, pal *menuPalette, plain bool) []infoSeg {
	var segs []infoSeg
	switch l.kind {
	case elAge:
		segs = []infoSeg{{text: l.text, ink: pal.ink, bold: true}}
	case elCount:
		segs = []infoSeg{{text: l.say(n), ink: pal.accent, bold: true}}
	default:
		segs = parseSplashLine(endLevelTag(l.level)+l.text, pal, pal.bg)
		if l.kind == elVerdict {
			for i := range segs {
				segs[i].bold = true
			}
		}
	}
	if plain {
		for i := range segs {
			segs[i].text = plainSafe(segs[i].text)
		}
	}
	return segs
}

// endReckFor sets the reckoning at a width.
func endReckFor(v *endView, pal *menuPalette, w int) endReck {
	var r endReck
	step := endStep
	if len(v.lines) > endManyLines {
		step = endStepFast
	}
	t := endLead
	for i, l := range v.lines {
		r.first = append(r.first, len(r.rows))
		r.at = append(r.at, t)
		r.silent = append(r.silent, l.kind == elVerdict)
		for _, row := range wrapInfo(infoLine{segs: endLineSegs(l, l.count, pal, v.plain)}, w) {
			r.rows = append(r.rows, row)
			r.of = append(r.of, i)
		}
		// A verdict stands alone a while before the next thing is said.
		if l.kind == elVerdict && (i+1 == len(v.lines) || v.lines[i+1].kind != elVerdict) {
			t += endVerdictHold
		} else {
			t += step
		}
	}
	r.frames = t - step + endTail
	if len(v.lines) == 0 {
		r.frames = endLead + endTail
	}
	return r
}

// ---- the scene ----

// endScene is the state of what moves in the film.
type endScene struct {
	L        endLayout
	prestige bool
	level    int
	spans    []endSpan
	rnd      menuRand
	frame    int
	t        float64
	sparks   menuSparks
	stars    []menuStar
	// burnAt is when each cell of the land burns, 0 to 1 through the
	// ending: the edges first, the ember's place last.
	burnAt []float64
	// The strike: how hot the word is, the blow's flare on it, the light
	// on the whole screen, the ring going out from the blow and how far
	// the blow has knocked the word sideways.
	heat, strike, flash, ring float64
	shake                     int
	reck                      endReck
	// landed is how many lines of the reckoning have landed.
	landed int
}

// endSpans is the film's beats for a reckoning of n frames: with motion,
// the script; without, a still frame each.
func endSpans(prestige, motion bool, reck int) []endSpan {
	beats := []endSpan{{beat: beatEnding, frames: endBurnFrames}, {beat: beatReckoning, frames: reck}}
	if prestige {
		beats = append(beats, endSpan{beat: beatStrike, frames: endStrikeFrames})
	}
	beats = append(beats, endSpan{beat: beatBeginning, frames: endDawnFrames})
	at := 0
	for i := range beats {
		if !motion {
			beats[i].frames = endStillFrames
		}
		beats[i].from = at
		at += beats[i].frames
	}
	return beats
}

// newEndScene makes the scene at its first frame.
func newEndScene(L endLayout, v *endView, pal *menuPalette) *endScene {
	sc := &endScene{L: L, prestige: v.prestige, level: v.end.Level, rnd: menuRand{s: 0x9e3779b97f4a7c15}, ring: -1}
	sc.reck = endReckFor(v, pal, max(L.lineW, 8))
	sc.spans = endSpans(v.prestige, v.motion, sc.reck.frames)
	if L.tiny {
		return sc
	}
	sc.stars = newMenuStars(&sc.rnd, L.w, L.h)
	// The fire comes in from the edges to the ember: a cell's turn is by
	// how far it is from the ember, with some chance in it.
	lh := L.h - L.top
	sc.burnAt = make([]float64, L.w*max(lh, 0))
	far := math.Hypot(float64(L.w)/4, float64(lh))
	for y := 0; y < lh; y++ {
		for x := 0; x < L.w; x++ {
			d := math.Hypot(float64(x-L.ex)/2, float64(L.top+y-L.ey)) / far
			sc.burnAt[y*L.w+x] = 0.62*(1-math.Min(d, 1)) + 0.30*sc.rnd.f()
		}
	}
	return sc
}

// total is the film's length in frames.
func (sc *endScene) total() int { return sc.spans[len(sc.spans)-1].end() }

// span is the beat the film is in at a frame, and the frame within it.
func (sc *endScene) span(frame int) (endSpan, int) {
	for _, s := range sc.spans {
		if frame < s.end() {
			return s, frame - s.from
		}
	}
	last := sc.spans[len(sc.spans)-1]
	return last, last.frames - 1
}

// burn is how far the ending's fire has come at frame f of the beat.
func burn(f int) float64 {
	return 1.25 * math.Max(0, math.Min(1, float64(f-endBurnHold)/float64(endBurnFrames-endBurnHold-endBurnTail)))
}

// enter sets a beat's own state at its start. What is in the air stays
// there from one beat to the next, but for the beginning, which comes out
// of the light.
func (sc *endScene) enter(s endSpan) {
	sc.heat, sc.strike, sc.flash, sc.ring, sc.shake, sc.landed = 0, 0, 0, -1, 0, 0
	switch s.beat {
	case beatReckoning:
		// A verdict is heard in still air.
		if len(sc.reck.silent) > 0 && sc.reck.silent[0] {
			sc.sparks.a = nil
		}
	case beatBeginning:
		sc.sparks.a = nil
		if sc.prestige {
			sc.flash, sc.strike = 0.9, 1
		}
	}
}

// seek puts the scene at the start of the beat a frame is in, as a key
// that skips there does: nothing in the air.
func (sc *endScene) seek(frame int) {
	s, _ := sc.span(frame)
	sc.frame = s.from
	sc.sparks.a = nil
	sc.enter(s)
}

// step moves the scene on one animation frame, and plays whatever the
// script has at that frame.
func (sc *endScene) step() {
	sc.frame++
	if sc.L.tiny {
		return
	}
	L := sc.L
	s, f := sc.span(sc.frame)
	if f == 0 {
		sc.enter(s)
	}
	switch s.beat {
	case beatEnding:
		// Sparks off the cells that are burning now.
		if p := burn(f); sc.prestige && len(sc.burnAt) > 0 {
			lh := L.h - L.top
			for i := 0; i < 40; i++ {
				x, y := int(sc.rnd.f()*float64(L.w)), int(sc.rnd.f()*float64(lh))
				if d := p - sc.burnAt[y*L.w+x]; d >= 0 && d < 0.12 {
					sc.sparks.emit(&sc.rnd, float64(x), float64(L.top+y), 1, 1, 0.6)
				}
			}
		}
	case beatReckoning:
		// A line lands: the ember flares and throws a few sparks. A
		// verdict lands in silence.
		for sc.landed < len(sc.reck.at) && sc.reck.at[sc.landed] <= f {
			if sc.prestige && !sc.reck.silent[sc.landed] {
				sc.strike = 0.9
				sc.sparks.emit(&sc.rnd, float64(L.ex), float64(L.ey), 5+min(sc.level, 10), 5, 0.7)
			}
			sc.landed++
		}
	case beatStrike:
		p := sc.L.word
		burst := func(n int, lift float64) {
			sc.sparks.emit(&sc.rnd, float64(p.x)+float64(p.w)/2, float64(p.y+p.h), n, float64(p.w)*0.95, lift)
		}
		switch f {
		case endBlow1:
			sc.strike, sc.flash = 0.45, 0.14
			burst(20, 0.6)
		case endBlow2:
			sc.strike, sc.flash = 0.65, 0.24
			burst(34, 0.75)
		case endBlow3:
			sc.strike, sc.flash = 0.85, 0.36
			burst(48, 0.9)
		case endBlow:
			// The heaviest blow in the game: the word at white heat and
			// knocked sideways, the whole screen lit, a ring going out
			// from it, and more sparks with every prestige.
			sc.strike, sc.flash, sc.ring, sc.shake = 1.7, 1.5, 0, 2
			more := 16 * min(sc.level, 10)
			burst(120+more/2, 1.2)
			sc.sparks.emit(&sc.rnd, float64(L.w)/2, float64(L.h)*0.62, 120+more/2, float64(L.w), 1.3)
		case endBlow + 1:
			sc.shake = -1
		case endBlow + 2:
			sc.shake = 0
		}
		sc.heat = math.Min(1, float64(f)/float64(endBlow))
		if sc.ring >= 0 {
			sc.ring += 3.2
		}
	case beatBeginning:
		// Embers off the new fire.
		if sc.rnd.f() < 0.8 {
			sc.sparks.emit(&sc.rnd, float64(L.w)/2+(sc.rnd.f()-0.5)*float64(L.w)*0.5, float64(L.dawnTop+2)+sc.rnd.f()*4, 1, 2, 0.5)
		}
	}
	for i := 0; i < menuSub; i++ {
		sc.t += 0.071 * menuPace
		sc.strike *= 0.938
		sc.flash *= 0.84
		sc.sparks.step(&sc.rnd)
	}
}

// pose puts the scene at the still frame of a beat (the motion setting
// off): the town with the fire at its edges, the reckoning all told, the
// word struck with the screen still lit, the new fire burning.
func (sc *endScene) pose(s endSpan) {
	sc.seek(s.from)
	sc.t = 4
	if sc.L.tiny {
		return
	}
	p := sc.L.word
	switch s.beat {
	case beatReckoning:
		sc.landed = len(sc.reck.at)
	case beatStrike:
		// A generator of its own: the still is the same every time.
		rnd := menuRand{s: 0x2545f4914f6cdd1d}
		sc.heat, sc.strike, sc.flash = 1, 0.7, 0.45
		sc.sparks.emit(&rnd, float64(p.x)+float64(p.w)/2, float64(p.y+p.h), 90, float64(p.w)*0.95, 1.1)
		sc.sparks.emit(&rnd, float64(sc.L.w)/2, float64(sc.L.h)*0.62, 60, float64(sc.L.w), 1.2)
		for i := 0; i < 9; i++ {
			sc.sparks.step(&rnd)
		}
	case beatBeginning:
		sc.flash, sc.strike = 0, 0.35
	}
}

// ---- drawing ----

// endHint draws the line that says how to go on, bottom right.
func (sc *endScene) endHint(g *mGrid, pal *menuPalette, v *endView) {
	if hint := " " + v.prompt + " "; v.prompt != "" && runeLen(hint) <= sc.L.w-4 {
		g.textOn(sc.L.w-2-runeLen(hint), sc.L.hint, hint, pal.groundDim, pal.ground, false)
	}
}

// emberGlyph is the last point of light.
func emberGlyph(plain bool) rune {
	if plain {
		return '*'
	}
	return '•'
}

// centreText writes a line centred on a row, in ink that reads on what is
// behind it.
func centreText(g *mGrid, pal *menuPalette, y int, s string, ink tcell.Color, bold bool) {
	x := centred(g.w, runeLen(s))
	readable(g, pal, x, y, g.text(x, y, s, ink, bold))
}

// render draws the film's frame. old and cur are the town that was and the
// town that is; either may be missing.
func (sc *endScene) render(pal *menuPalette, v *endView, old, cur *menuTown, mf mapstyle.Frame, still bool) *mGrid {
	L := sc.L
	g := newMGrid(L.w, L.h)
	s, f := sc.span(sc.frame)
	if L.tiny {
		line := v.age
		switch s.beat {
		case beatReckoning:
			line = v.complete
			if line == "" && len(v.lines) > 0 {
				line = plainOf(endLineSegs(v.lines[0], v.lines[0].count, pal, v.plain))
			}
		case beatStrike:
			line = strings.ToUpper(endWord)
		case beatBeginning:
			line = fitOption(v.captions, L.w)
		}
		line = truncate(line, max(L.w, 0))
		g.text(centred(L.w, runeLen(line)), max(L.h/3, 0), line, pal.ink, true)
		return g
	}
	drawMenuStars(g, pal, sc.stars, sc.t)
	switch s.beat {
	case beatEnding:
		sc.drawEnding(g, pal, v, old, mf, f, still)
	case beatReckoning:
		sc.drawReckoning(g, pal, v, f, still)
	case beatStrike:
		sc.drawStrike(g, pal, v)
	case beatBeginning:
		sc.drawBeginning(g, pal, v, cur, mf, f, still)
	}
	sc.endHint(g, pal, v)
	return g
}

// drawEnding: the town that was built, burning down to one ember (or, for
// a fall, going cold and dark).
func (sc *endScene) drawEnding(g *mGrid, pal *menuPalette, v *endView, old *menuTown, mf mapstyle.Frame, f int, still bool) {
	L := sc.L
	p := burn(f)
	if still {
		p = 0.42 // the fire well in from the edges, the town still standing
	}
	lh := L.h - L.top
	land := newMGrid(L.w, max(lh, 1))
	old.drawFoot(land, pal, 0, 0, L.w, lh, mf, false)
	low := rampAt(&pal.fire, 0.45)
	for y := 0; y < lh; y++ {
		for x := 0; x < L.w; x++ {
			c := land.c[y*L.w+x]
			if c.r == ' ' && !c.hasBg {
				continue
			}
			d := p - sc.burnAt[y*L.w+x]
			switch {
			case d < -0.1: // standing
			case d < 0: // the heat reaches it
				if v.prestige {
					c.fg = theme.Mix(c.fg, low, 0.5)
				}
			case d < 0.12: // burning (or, in a fall, going grey)
				c.fg, c.bold = rampAt(&pal.heat, 0.95-2.6*d), true
				if !v.prestige {
					c.fg, c.bold = pal.dim, false
				}
				c.hasBg = false
			case d < 0.3: // an ember of it
				c.r, c.fg, c.bold, c.hasBg = '·', theme.Mix(low, pal.bg, (d-0.12)/0.18*0.7), false, false
				if !v.prestige {
					c.fg = theme.Mix(pal.faint, pal.bg, (d-0.12)/0.18)
				}
			default:
				continue
			}
			g.c[(L.top+y)*g.w+x] = c
		}
	}
	sc.sparks.draw(g, pal)
	// What is ending: the age, and the run's last line in its voice.
	name := spaced(strings.ToUpper(v.age))
	if runeLen(name) > L.w-4 {
		name = strings.ToUpper(v.age)
	}
	g.textOn(centred(L.w, runeLen(name)+2), 1, " "+name+" ", pal.ink, pal.bg, true)
	for i, line := range L.voice {
		g.textOn(centred(L.w, runeLen(line)+2), 3+i, " "+line+" ", pal.dim, pal.bg, false)
	}
	sc.drawEmber(g, pal, v, math.Max(0, math.Min(1, (p-0.85)/0.3)))
}

// drawEmber draws the last point of light, k of the way lit.
func (sc *endScene) drawEmber(g *mGrid, pal *menuPalette, v *endView, k float64) {
	if k <= 0 {
		return
	}
	glow := 0.62 + 0.14*math.Sin(sc.t*3.1) + 0.5*sc.strike
	ink := rampAt(&pal.heat, glow)
	if !v.prestige {
		ink = pal.dim
	}
	g.set(sc.L.ex, sc.L.ey, emberGlyph(v.plain), theme.Mix(pal.bg, ink, k), pal.bg, true)
}

// drawReckoning: in the dark, what the run earned, a line at a time.
func (sc *endScene) drawReckoning(g *mGrid, pal *menuPalette, v *endView, f int, still bool) {
	L, r := sc.L, &sc.reck
	sc.drawEmber(g, pal, v, 1)
	sc.sparks.draw(g, pal)
	// The rows that have landed, and where the block stands: centred when
	// it all fits, and when it does not, moved up as lines land so that
	// the newest is always on the screen.
	shown := 0
	if sc.landed > 0 {
		shown = len(r.rows)
		if sc.landed < len(r.first) {
			shown = r.first[sc.landed]
		}
	}
	room := L.l1 - L.l0
	y0, off := L.l0+(room-len(r.rows))/2, 0
	if len(r.rows) > room {
		y0, off = L.l0, max(0, shown-room)
	}
	hot := rampAt(&pal.heat, 1)
	for i := off; i < shown; i++ {
		row, line := r.rows[i], v.lines[r.of[i]]
		age := f - r.at[r.of[i]]
		if still {
			age = 99
		}
		if line.kind == elCount && age < endCountUp {
			n := line.count * (age + 1) / endCountUp
			row = infoLine{segs: endLineSegs(line, n, pal, v.plain)}
		}
		// It lands hot and cools to its ink; a verdict does not glow.
		k := 0.0
		if v.prestige && line.kind != elVerdict {
			k = math.Max(0, 1-float64(age)/3)
		}
		x := centred(L.w, row.width())
		for _, seg := range row.segs {
			x += g.text(x, y0+i-off, seg.text, theme.Mix(seg.ink, hot, k), seg.bold || k > 0)
		}
	}
}

// endLightGone is the light below which the screen is dark again, and the
// theme's own weather shows in it.
const endLightGone = 0.05

// lightUp lights the whole screen by the blow's flash.
func (sc *endScene) lightUp(g *mGrid, pal *menuPalette) {
	if sc.flash > endLightGone {
		light(g, pal, sc.flash)
	}
}

// drawStrike: the word for it in iron, every cell lit.
func (sc *endScene) drawStrike(g *mGrid, pal *menuPalette, v *endView) {
	L, p := sc.L, sc.L.word
	sc.lightUp(g, pal)
	if sc.ring >= 0 {
		// The ring: a band of light going out from the middle of the word.
		cx, cy := float64(p.x)+float64(p.w)/2, float64(p.y)+float64(p.h)/2
		hot := rampAt(&pal.heat, 1)
		for y := 0; y < L.h; y++ {
			for x := 0; x < L.w; x++ {
				if d := math.Hypot((float64(x)-cx)/2, float64(y)-cy) - sc.ring; d > -1.6 && d < 1.6 {
					c := &g.c[y*g.w+x]
					bg := pal.bg
					if c.hasBg {
						bg = c.bg
					}
					c.bg, c.hasBg = theme.Mix(bg, hot, 0.45*(1-math.Abs(d)/1.6)), true
				}
			}
		}
	}
	sc.sparks.draw(g, pal)
	p.x += sc.shake
	cold := 0.62 * (1 - sc.heat)
	pw := 1
	if p.ok {
		pw = len(p.rows[0])
	}
	heat := func(px, py int) float64 { return ironHeat(px, py, pw, sc.t, sc.strike, 0.5) - cold }
	if p.ok {
		drawIron(g, pal, p.rows, p.x, p.y, p.mode, heat, v.plain)
	} else {
		centreText(g, pal, p.y, spaced(strings.ToUpper(endWord)), rampAt(&pal.heat, 0.5+0.5*sc.heat), true)
	}
	if sc.heat < 1 {
		return // the level and the game's line come with the heavy blow
	}
	// A star for every prestige so far, as many as the row holds.
	if n := min(sc.level, (L.w-8)/2); n > 0 && L.stars < L.hint-1 {
		star := "★"
		if v.plain {
			star = "*"
		}
		centreText(g, pal, L.stars, strings.TrimSpace(strings.Repeat(star+" ", n)), rampAt(&pal.heat, 0.8), true)
	}
	for i, line := range L.caption {
		if L.capRow+i < L.hint {
			centreText(g, pal, L.capRow+i, line, pal.ink, true)
		}
	}
}

// drawBeginning: out of the light, the first fire of the new run.
func (sc *endScene) drawBeginning(g *mGrid, pal *menuPalette, v *endView, cur *menuTown, mf mapstyle.Frame, f int, still bool) {
	L := sc.L
	lh := L.h - L.dawnTop
	cur.drawFoot(g, pal, 0, L.dawnTop, L.w, lh, mf, sc.flash > 0.25)
	if !v.prestige && !still {
		// A fall's dawn comes up out of the dark.
		dim(g, pal, 0, L.dawnTop, L.w, lh, math.Max(0, 1-float64(f+1)/10))
	}
	sc.lightUp(g, pal)
	sc.sparks.draw(g, pal)
	if L.wsx > 0 {
		drawWordmark(g, pal, L.wx, L.wy, L.wsx, sc.t, sc.strike, 0.5, v.plain)
	}
	if capt := fitOption(v.captions, L.w-4); capt != "" {
		if v.plain {
			capt = plainSafe(capt)
		}
		centreText(g, pal, L.dawnCap, capt, pal.ink, true)
	}
}
