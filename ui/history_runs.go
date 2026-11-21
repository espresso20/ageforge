package ui

import (
	"image"
	"image/color"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

var (
	txtRunCol      = rowid.E("ui.history_runs.words.01", "Run")
	txtEndCol      = rowid.E("ui.history_runs.words.02", "End")
	txtHighestRun  = rowid.E("ui.history_runs.words.03", "Highest age")
	txtLengthCol   = rowid.E("ui.history_runs.words.04", "Length")
	txtPaceCol     = rowid.E("ui.history_runs.words.05", "Pace")
	txtInPlay      = rowid.E("ui.history_runs.words.06", "In play")
	txtNoRuns      = rowid.E("ui.history_runs.words.07", "No finished runs yet")
	txtWorkersCol  = rowid.E("ui.history_runs.words.08", "Workers")
	histRunWheel   = 3
	histCountsFrom = 120 // the panel width from which the list shows its four counts
)

// runCols are the list's column offsets from the panel's left margin.
type runCols struct{ name, end, age, length, counts, pace, stop int }

func histRunCols(w int) runCols {
	c := runCols{name: 6, end: 22, age: 25, length: 41, stop: w - 5}
	c.pace = 51
	if w >= histCountsFrom {
		c.counts, c.pace = 51, 100
	}
	return c
}

// runRow is one line of the list, the run in play or a filed run.
type runRow struct {
	name     string
	end      string // the outcome, "" for the run in play
	age      string
	length   int
	counts   [4]float64
	spans    []histSpan
	kept     bool
	best     bool
	playing  bool
	filedIdx int
}

// bestRun is the filed run that reached the highest age, the shortest among equals.
func (p *historyPanel) bestRun() int {
	best := -1
	for i, r := range p.arch.Runs {
		a := slices.Index(p.v.keys, r.Age)
		if best < 0 {
			best = i
			continue
		}
		b := slices.Index(p.v.keys, p.arch.Runs[best].Age)
		if a > b || a == b && r.Tick < p.arch.Runs[best].Tick {
			best = i
		}
	}
	return best
}

// runRows is the list's rows: the run in play, then the filed runs newest first.
func (p *historyPanel) runRows() []runRow {
	src := histSource{hist: p.hist, runs: p.arch.Runs}
	if keys := p.v.keys; len(keys) > 0 {
		src.firstAge = keys[0]
	}
	cur := runRow{name: p.name, playing: true, kept: true, length: p.v.end, filedIdx: -1, spans: src.spans(rangeRun)}
	if n := len(cur.spans); n > 0 {
		cur.age = cur.spans[n-1].Key
	}
	for i, col := range []string{"buildings", "population", "techs", "expeditions"} {
		if s := src.live(rangeRun, col); len(s.V) > 0 {
			cur.counts[i] = s.V[len(s.V)-1]
		}
	}
	rows := []runRow{cur}
	best := p.bestRun()
	for i, r := range p.arch.Runs {
		rows = append(rows, runRow{
			name: r.Run, end: r.Outcome, age: r.Age, length: r.Tick, spans: src.pastSpans(r), kept: kept(r), best: i == best, filedIdx: i,
			counts: [4]float64{float64(r.Built), float64(r.Recruited), float64(r.Researched), float64(r.Expeditions)},
		})
	}
	return rows
}

// runKey moves the list's cursor.
func (p *historyPanel) runKey(d int) {
	if len(p.arch.Runs) == 0 {
		return
	}
	p.cursor = clamp(p.cursor+d, 0, len(p.arch.Runs)-1)
	p.Layout(p.Bounds())
}

// pickRun draws the run under the cursor behind the current one, or lets go of
// it when it already is. A run with no kept history cannot be picked.
func (p *historyPanel) pickRun() {
	if p.cursor < 0 || p.cursor >= len(p.arch.Runs) || !kept(p.arch.Runs[p.cursor]) {
		return
	}
	id := p.arch.Runs[p.cursor].ID
	if p.ghost == id {
		id = ""
	}
	p.ghost = id
	p.Layout(p.Bounds())
}

// scrollRuns moves the list by rows, keeping the cursor on the list.
func (p *historyPanel) scrollRuns(d int) {
	p.scroll += d
	p.Layout(p.Bounds())
}

// runWindow is the first filed run shown, kept so the cursor is in sight.
func (p *historyPanel) runWindow() int {
	rows := p.v.geo.listRows
	p.scroll = clamp(p.scroll, 0, max(0, len(p.arch.Runs)-rows))
	if p.cursor < p.scroll {
		p.scroll = p.cursor
	} else if p.cursor >= p.scroll+rows {
		p.scroll = p.cursor - rows + 1
	}
	return p.scroll
}

// layoutRuns makes each filed run's row a hit.
func (p *historyPanel) layoutRuns(g *xgrid) {
	gm := p.v.geo
	if p.rng != rangeAll {
		return
	}
	first := p.runWindow()
	for i := range gm.listRows {
		if first+i < len(p.arch.Runs) {
			g.hit(hitHistRun, first+i, 1, gm.listTop+2+i, p.w-5, 1)
		}
	}
}

// fill paints cells in a flat colour.
func histFill(p2 *kit.Painter, org image.Point, r image.Rectangle, c color.RGBA) {
	p2.Fill(r.Add(org), theme.RGB(c.R, c.G, c.B))
}

// sign puts a 9 by 9 sign on two cells, or its letter when pictures are off.
func (p *historyPanel) sign(p2 *kit.Painter, org image.Point, x, y int, kind string, col, ground color.RGBA) {
	if p.plain {
		p.text(p2, org, x, y, histGlyph[kind], col, false)
		return
	}
	cw, ch := cellPixels()
	k := max(1, ch/12)
	img := image.NewRGBA(image.Rect(0, 0, 2*cw, ch))
	a := art{img, k}
	a.fill(image.Rect(0, 0, 2*cw/k, ch/k), 0, ground)
	histDrawGlyph(a, (2*cw/k-9)/2, (ch/k-9)/2, kind, col)
	p2.Picture(image.Rect(x, y, x+2, y+1).Add(org), img)
}

// runs paints the list under the chart: headings, the run in play, the filed runs.
func (p *historyPanel) runs(p2 *kit.Painter, org image.Point) {
	v := p.v
	g, c := v.geo, v.ink
	cl := histRunCols(g.w)
	x0 := 2
	heads := []struct {
		x int
		s string
	}{{cl.name, txtRunCol.String()}, {cl.end, txtEndCol.String()}, {cl.age + 3, txtHighestRun.String()}, {cl.length, txtLengthCol.String()}, {cl.pace, txtPaceCol.String()}}
	for _, h := range heads {
		p.text(p2, org, x0+h.x, g.listTop, h.s, c.dim, false)
	}
	if cl.counts > 0 {
		for i, s := range []string{txtBuildings.String(), txtWorkersCol.String(), txtTechsShort.String(), txtExpeditions.String()} {
			p.text(p2, org, x0+cl.counts+i*12+11-runeLen(s), g.listTop, s, c.dim, false)
		}
	}
	rows := p.runRows()
	longest := 1
	for _, r := range rows {
		if n := len(r.spans); n > 0 {
			longest = max(longest, r.spans[n-1].End)
		}
	}
	p.runLine(p2, org, rows[0], rows[0], g.listTop+1, longest)
	if len(rows) == 1 {
		p.text(p2, org, x0+cl.name, g.listTop+2, txtNoRuns.String(), c.dim, false)
		return
	}
	first := p.runWindow()
	for i := range g.listRows {
		if first+i < len(p.arch.Runs) {
			p.runLine(p2, org, rows[1+first+i], rows[1+first+i], g.listTop+2+i, longest)
		}
	}
	p2.Bar(org.X+g.w-2, org.Y+g.listTop+2, g.listRows, first, g.listRows, len(p.arch.Runs))
}

// runLine paints one row of the list.
func (p *historyPanel) runLine(p2 *kit.Painter, org image.Point, _, r runRow, y, longest int) {
	v := p.v
	g, c := v.geo, v.ink
	cl := histRunCols(g.w)
	x0 := 2
	bg := c.ground
	isGhost := !r.playing && p.ghost != "" && p.arch.Runs[r.filedIdx].ID == p.ghost
	switch {
	case r.playing:
		bg = toward(c.ground, c.accent, .10)
	case isGhost:
		bg = toward(c.ground, c.label, .14)
	case p.rng == rangeAll && r.filedIdx == p.cursor:
		bg = toward(c.ground, c.dim, .12)
	}
	histFill(p2, org, image.Rect(x0, y, g.w-4, y+1), bg)
	text := c.text
	if !r.kept {
		text = toward(c.ground, c.dim, .85)
	}
	switch {
	case r.playing:
		p.sign(p2, org, x0, y, "end-play", c.accent, bg)
	case r.best:
		p.sign(p2, org, x0, y, "crown", c.highlight, bg)
	}
	switch {
	case isGhost:
		p.sign(p2, org, x0+2, y, "ghost", c.label, bg)
	case !r.kept:
		p.sign(p2, org, x0+2, y, "ghost", toward(c.ground, c.dim, .45), bg)
		p.text(p2, org, x0+4, y, "×", c.negative, false)
	}
	if !r.playing {
		ec := c.dim
		switch r.end {
		case game.RunPrestige:
			ec = c.positive
		case game.RunSuccumbed:
			ec = c.negative
		}
		if !r.kept {
			ec = toward(c.ground, ec, .6)
		}
		p.sign(p2, org, x0+cl.end, y, "end-"+r.end, ec, bg)
	}
	ai := v.ageIdx(r.age)
	ac := c.ageColor(ai)
	if !r.kept {
		ac = toward(c.ground, ac, .55)
	}
	histFill(p2, org, image.Rect(x0+cl.age, y, x0+cl.age+1, y+1), ac)
	name, nc, bold, tag := r.name, text, false, ""
	room := cl.end - cl.name - 1
	if r.playing {
		// the run in play has no outcome sign, so its name runs to the age swatch
		nc, bold = c.bright, true
		tag = " " + txtInPlay.String()
		room = cl.age - cl.name - 1
	}
	p.text(p2, org, x0+cl.name, y, fitText(name, room-runeLen(tag))+tag, nc, bold)
	p.text(p2, org, x0+cl.age+3, y, histFit(v.ageName(r.age), cl.length-cl.age-4), ac, false)
	p.text(p2, org, x0+cl.length, y, histTime(r.length), text, false)
	if cl.counts > 0 {
		for i, n := range r.counts {
			s := textfmt.Number(n)
			p.text(p2, org, x0+cl.counts+i*12+11-runeLen(s), y, s, text, false)
		}
	}
	p.pace(p2, org, r, y, x0+cl.pace, cl.stop-cl.pace, longest, bg)
}

// histFit is an age name cut to room cells: whole, then without " Age", then nothing.
func histFit(s string, room int) string {
	if runeLen(s) <= room {
		return s
	}
	if t := strings.TrimSuffix(s, " Age"); runeLen(t) <= room {
		return t
	}
	return ""
}

// pace is a run's ages in proportion to how long each took, on one scale for the list.
func (p *historyPanel) pace(p2 *kit.Painter, org image.Point, r runRow, y, x, w, longest int, bg color.RGBA) {
	v := p.v
	c := v.ink
	if !r.kept || len(r.spans) == 0 || w <= 0 {
		p.text(p2, org, x, y, strings.Repeat("·", max(0, w)), toward(c.ground, c.dim, .3), false)
		return
	}
	for i, s := range r.spans {
		a, b := x+s.Start*w/longest, x+s.End*w/longest
		if b <= a {
			b = a + 1
		}
		col := c.ageColor(v.ageIdx(s.Key))
		if i > 0 {
			col = toward(col, bg, 0)
		}
		histFill(p2, org, image.Rect(a, y, min(b, x+w), y+1), col)
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
