package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// history_panel.go is the History overlay: the game's metrics page. One chart
// of the run in play with three ranges, a column to pick the measures, a
// pointer that reads them out, and (history_runs.go) the filed runs. The
// chart is history_plot.go; the data is history_data.go.

var (
	txtHistoryTitle = rowid.E("ui.dashboard.NewDashboard.16", "Civilization History")
	txtThisAge      = rowid.E("ui.history_panel.words.01", "This age")
	txtAllRuns      = rowid.E("ui.history_panel.words.03", "All runs")
	txtLinear       = rowid.E("ui.history_panel.words.04", "Linear")
	txtLog          = rowid.E("ui.history_panel.words.05", "Log")
	txtPeople       = rowid.E("ui.history_panel.words.06", "People")
	txtProgress     = rowid.E("ui.history_panel.words.07", "Progress")
	txtRates        = rowid.E("ui.history_panel.words.08", "Rates")
	txtOwnScale     = rowid.E("ui.history_panel.words.09", "each measure on its own scale")
	txtStaysLinear  = rowid.E("ui.history_panel.words.10", "stays linear")
	txtKeyPointer   = rowid.E("ui.history_panel.words.11", "pointer")
	txtKeyMeasure   = rowid.E("ui.history_panel.words.12", "measure")
	txtKeyShow      = rowid.E("ui.history_panel.words.13", "show")
	txtKeyRange     = rowid.E("ui.history_panel.words.14", "range")
	txtKeyLog       = rowid.E("ui.history_panel.words.15", "log scale")
	txtKeyRun       = rowid.E("ui.history_panel.words.16", "run")
	txtKeyGhost     = rowid.E("ui.history_panel.words.17", "ghost")
)

// The chart's hit ids, on the panel's own grid.
const (
	hitHistMeasure = iota + 1
	hitHistRun
)

// historyPanel is the History overlay's primitive: the whole screen but the
// command bar. The keyboard stays with the command bar; the panel takes the
// keys that print nothing, and the digits and Space, while the bar is empty.
type historyPanel struct {
	barPanel
	kit.Base

	// src reads the engine: the run's recorder, the filed runs and the save's name.
	src func() (*game.RunHistory, game.RunArchiveView, string)
	// settings reads the glyph tier.
	settings func() mapSettings

	set            *rules.Set
	hist           *game.RunHistory
	arch           game.RunArchiveView
	name           string
	ms             []histMeasure
	rng            histRange
	logs           [histRanges]bool // the scale chosen for each range, while the panel is open
	slots          [histSlots]string
	row            int // the picker's cursor, an index into ms
	ptr            int // the tick under the pointer, -1 for the newest
	hover          image.Point
	ghost          string // the filed run drawn behind, by id
	cursor, scroll int    // the runs list's cursor and first row

	w, h   int
	chrome *xgrid
	v      *histView
	plain  bool
	cache  map[string]*image.RGBA
}

func newHistoryPanel() *historyPanel {
	p := &historyPanel{w: 100, h: 30, ptr: -1, hover: image.Pt(-1, -1), cache: map[string]*image.RGBA{}}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *historyPanel) open(state game.GameState) Primitive {
	p.set = state.Ruleset()
	p.rng, p.ptr, p.row, p.ghost, p.cursor, p.scroll = rangeRun, -1, 0, "", 0, 0
	p.logs = [histRanges]bool{rangeAll: true}
	p.slots = [histSlots]string{"population"}
	p.load()
	return overBar(p)
}

func (p *historyPanel) update(state game.GameState) {
	p.set = state.Ruleset()
	p.load()
}

// load reads the engine afresh and drops what no longer exists.
func (p *historyPanel) load() {
	if p.src != nil {
		p.hist, p.arch, p.name = p.src()
	}
	p.ms = histMeasures(p.hist, func(k string) string {
		return rowid.Sprintf("ui.history_panel.load.01", "%s rate", p.set.Name(rules.KindResource, k))
	})
	if p.runByID(p.ghost) == nil {
		p.ghost = ""
	}
	p.row = clamp(p.row, 0, len(p.ms)-1)
}

func (p *historyPanel) runByID(id string) *game.RunRecord {
	for i := range p.arch.Runs {
		if p.arch.Runs[i].ID == id && kept(p.arch.Runs[i]) {
			return &p.arch.Runs[i]
		}
	}
	return nil
}

func (p *historyPanel) measure(key string) (histMeasure, bool) {
	for _, m := range p.ms {
		if m.Key == key {
			return m, true
		}
	}
	return histMeasure{}, false
}

// toggle shows a measure in the first free slot, or clears it when shown. With
// every slot taken it takes the last one over. A measure finished runs do not
// keep cannot be picked in All runs.
func (p *historyPanel) toggle(i int) {
	if i < 0 || i >= len(p.ms) || p.rng == rangeAll && p.ms[i].Past == "" {
		return
	}
	key := p.ms[i].Key
	if at := slices.Index(p.slots[:], key); at >= 0 {
		p.slots[at] = ""
	} else if free := slices.Index(p.slots[:], ""); free >= 0 {
		p.slots[free] = key
	} else {
		p.slots[histSlots-1] = key
	}
	p.row = i
	p.Layout(p.Bounds())
}

func (p *historyPanel) setRange(r histRange) {
	if r < 0 || r >= histRanges || r == p.rng {
		return
	}
	p.rng = r
	for i, k := range p.slots {
		if m, ok := p.measure(k); ok && r == rangeAll && m.Past == "" {
			p.slots[i] = ""
		}
	}
	p.Layout(p.Bounds())
}

func (p *historyPanel) setLog(on bool) {
	p.logs[p.rng] = on
	p.Layout(p.Bounds())
}

// movePointer steps the pointer to the neighbouring sample.
func (p *historyPanel) movePointer(d int) {
	if p.v == nil || len(p.v.ticks) == 0 {
		return
	}
	i := slices.Index(p.v.ticks, p.v.ptr)
	i = clamp(i+d, 0, len(p.v.ticks)-1)
	p.ptr = p.v.ticks[i]
	if i == len(p.v.ticks)-1 {
		p.ptr = -1
	}
	p.Layout(p.Bounds())
}

// build lays the chart out for the panel's size and state.
func (p *historyPanel) build() {
	all := p.rng == rangeAll
	src := histSource{hist: p.hist, runs: p.arch.Runs}
	keys := p.set.AgeKeys()
	if len(keys) > 0 {
		src.firstAge = keys[0]
	}
	v := &histView{geo: histGeometry(p.w, p.h, all, len(p.arch.Runs)), ink: newHistInk(), set: p.set, keys: keys, all: all, ptr: -1}
	p.v = v
	if p.hist == nil {
		return
	}
	v.lo, v.end = src.window(p.rng)
	v.hi = v.end
	v.spans, v.marks = src.spans(p.rng), slices.Clone(src.marks(p.rng))
	slices.SortStableFunc(v.marks, func(a, b game.HistoryMark) int { return a.Tick - b.Tick })
	if all {
		if g := p.runByID(p.ghost); g != nil {
			v.ghost, v.gEnd, v.gSpans = g, g.History.Run.End, src.pastSpans(*g)
			v.hi = max(v.hi, v.gEnd)
		}
	}
	for _, t := range src.live(p.rng, "population").T {
		v.ticks = append(v.ticks, t)
	}
	if len(v.ticks) > 0 {
		v.ptr = v.ticks[len(v.ticks)-1]
		if p.ptr >= 0 {
			v.ptr = v.nearest(p.ptr)
		}
	}
	for slot, key := range p.slots {
		m, ok := p.measure(key)
		if !ok {
			continue
		}
		l := histLine{m: m, slot: slot, cur: shownSeries(m, src.live(p.rng, m.Key))}
		if v.ghost != nil {
			l.gh = shownSeries(m, past(*v.ghost, m.Past))
		}
		l.lg = p.logs[p.rng] && m.Cap == 0
		l.lo, l.hi = scaleOf(l)
		v.logOn = v.logOn || l.lg
		v.lines = append(v.lines, l)
	}
}

// shownSeries is a series in the units the measure is shown in.
func shownSeries(m histMeasure, s histSeries) histSeries {
	out := histSeries{T: s.T, V: make([]float64, len(s.V))}
	for i, v := range s.V {
		out.V[i] = m.shown(v)
	}
	return out
}

// histNice rounds v up to 1, 1.5, 2, 2.5, 3, 4, 5, 6, 8 or 10 of its power of ten.
func histNice(v float64) float64 {
	if v <= 0 {
		return 1
	}
	p := math.Pow(10, math.Floor(math.Log10(v)))
	for _, s := range []float64{1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10} {
		if v <= s*p*(1+1e-9) {
			return s * p
		}
	}
	return 10 * p
}

// scaleOf is a line's range: its fixed cap or the nice number above its
// largest value; on a log scale the power of ten under its smallest.
func scaleOf(l histLine) (lo, hi float64) {
	top, low := 0.0, math.Inf(1)
	for _, s := range []histSeries{l.cur, l.gh} {
		for _, v := range s.V {
			top = math.Max(top, v)
			if v > 0 {
				low = math.Min(low, v)
			}
		}
	}
	hi = histNice(top * 1.04)
	if l.m.Cap > 0 {
		hi = l.m.Cap
	}
	if l.lg {
		lo = hi / 1000
		if !math.IsInf(low, 1) {
			lo = math.Pow(10, math.Floor(math.Log10(low)))
		}
		if lo >= hi {
			lo = hi / 10
		}
	}
	return lo, hi
}

// text is a figure in its measure's units.
func (m histMeasure) text(v float64, unit bool) string {
	switch {
	case m.Rate:
		return rateAmount(v, int(game.BaseTickInterval.Milliseconds()))
	case m.Unit == "%":
		return fmt.Sprintf("%.0f%%", v)
	case m.Unit == "x":
		return fmt.Sprintf("%.2fx", v)
	case unit:
		return textfmt.Number(math.Round(v)) + m.Unit
	}
	return textfmt.Number(math.Round(v))
}

// Layout lays the frame and the chart out in r.
func (p *historyPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.w, p.h = r.Dx(), r.Dy()
	p.plain = pictureCell == nil || p.settings != nil && p.settings().Tier == mapmodel.TierASCII
	p.build()
	g := newXGrid(p.w, p.h)
	g.box(0, 0, p.w, p.h, xBorder, txtHistoryTitle.String(), xTitle, false)
	g.closeButton(p.w)
	x := 2
	for i, name := range []string{txtThisAge.String(), txtThisRun.String(), txtAllRuns.String()} {
		k := xDim
		if histRange(i) == p.rng {
			k = xSel
		}
		end := g.put(x, 1, " "+name+" ", k)
		g.tab(image.Rect(x, 1, end, 2), name, i, histRange(i) == p.rng)
		x = end + 1
	}
	lin, lg := " "+txtLinear.String()+" ", " "+txtLog.String()+" "
	x = p.w - 2 - runeLen(lin) - 1 - runeLen(lg)
	for i, name := range []string{lin, lg} {
		k := xDim
		if p.logs[p.rng] == (i == 1) {
			k = xSel
		}
		end := g.put(x, 1, name, k)
		g.tab(image.Rect(x, 1, end, 2), strings.TrimSpace(name), 3+i, p.logs[p.rng] == (i == 1))
		x = end + 1
	}
	p.layoutPicker(g)
	p.layoutRuns(g)
	keys := [][2]string{{"←→", txtKeyPointer.String()}, {"1-0", txtKeyMeasure.String()}}
	if p.rng != rangeAll {
		keys = append(keys, [2]string{"Space", txtKeyShow.String()})
	}
	keys = append(keys, [2]string{"Tab", txtKeyRange.String()}, [2]string{"-", txtKeyLog.String()})
	if p.rng == rangeAll {
		keys = append(keys, [2]string{"↑↓", txtKeyRun.String()}, [2]string{"Enter", txtKeyGhost.String()})
	}
	expKeyBar(g, expLayout{w: p.w, h: p.h, footY: p.v.geo.footY}, append(keys, escClose()), "")
	if p.plain {
		g.fold()
	}
	p.chrome = g
}

// pickRow is a row of the measure picker: a heading, or a measure.
type pickRow struct {
	head string
	idx  int
}

// pickRows are the picker's rows in order: each group's heading, then its measures.
func (p *historyPanel) pickRows() []pickRow {
	var rows []pickRow
	for gi, name := range []string{txtPeople.String(), txtProgress.String(), txtRates.String()} {
		rows = append(rows, pickRow{head: name, idx: -1})
		for i, m := range p.ms {
			if m.Group == histGroup(gi) {
				rows = append(rows, pickRow{idx: i})
			}
		}
	}
	return rows
}

// pickerWindow is the rows the picker shows: it scrolls to keep the cursor in sight.
func (p *historyPanel) pickerWindow() (rows []pickRow, first int) {
	rows = p.pickRows()
	n := max(1, p.v.geo.legendY-p.v.geo.y0)
	cur := slices.IndexFunc(rows, func(r pickRow) bool { return r.idx == p.row })
	first = clamp(cur-n+1, 0, max(0, len(rows)-n))
	return rows[first:min(len(rows), first+n)], first
}

// layoutPicker writes the picker column into the grid: headings, the keyed
// measures as buttons, the rest as names.
func (p *historyPanel) layoutPicker(g *xgrid) {
	rows, _ := p.pickerWindow()
	for j, r := range rows {
		y := p.v.geo.y0 + j
		if r.idx < 0 {
			g.put(2, y, r.head, xDim)
			continue
		}
		m := p.ms[r.idx]
		name := string([]rune(m.Label)[:min(runeLen(m.Label), 17)])
		if r.idx < 10 {
			g.keyButton(4, y, fmt.Sprint((r.idx+1)%10), name)
		} else {
			k := xText
			if p.rng == rangeAll && m.Past == "" {
				k = xFaint
			}
			g.put(4, y, name, k)
		}
		g.hit(hitHistMeasure, r.idx, 1, y, 28, 1)
		if r.idx == p.row {
			g.put(1, y, "▸", xAccent)
		}
	}
}

// routeKey takes a key if it is one of the panel's. With something typed in
// the prompt every key belongs to the prompt.
func (p *historyPanel) routeKey(ev *kit.Key, prompt string) bool {
	if prompt != "" {
		return false
	}
	switch ev.Code {
	case kit.KeyLeft:
		p.movePointer(-1)
	case kit.KeyRight:
		p.movePointer(1)
	case kit.KeyHome:
		p.ptr = -1
		if p.v != nil && len(p.v.ticks) > 0 {
			p.ptr = p.v.ticks[0]
		}
		p.Layout(p.Bounds())
	case kit.KeyEnd:
		p.ptr = -1
		p.Layout(p.Bounds())
	case kit.KeyTab:
		p.setRange((p.rng + 1) % histRanges)
	case kit.KeyBacktab:
		p.setRange((p.rng + histRanges - 1) % histRanges)
	case kit.KeyUp, kit.KeyDown:
		d := 1
		if ev.Code == kit.KeyUp {
			d = -1
		}
		if p.rng == rangeAll {
			p.runKey(d)
		} else {
			p.row = clamp(p.row+d, 0, len(p.ms)-1)
			p.Layout(p.Bounds())
		}
	case kit.KeyEnter:
		if p.rng != rangeAll {
			return false
		}
		p.pickRun()
	case kit.KeyRune:
		switch {
		case ev.Rune == ' ' && p.rng != rangeAll:
			p.toggle(p.row)
		case ev.Rune == '-':
			p.setLog(!p.logs[p.rng])
		case ev.Rune >= '0' && ev.Rune <= '9':
			p.toggle(int((ev.Rune - '1' + 10) % 10))
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// Mouse: the buttons and tabs, and the pointer following the mouse over the chart.
func (p *historyPanel) Mouse(m kit.Mouse) bool {
	switch {
	case p.chrome == nil:
		return false
	case p.buttons.mouse(p.chrome, p.Bounds().Min, m):
		return true
	case m.Kind == kit.MouseLeave:
		p.hover = image.Pt(-1, -1)
		return false
	case m.Kind == kit.MouseMove:
		p.hover = m.Pt().Sub(p.Bounds().Min)
		p.followMouse()
		return false
	case m.Kind == kit.MouseWheel && p.rng == rangeAll:
		p.scrollRuns(m.Wheel * histRunWheel)
		return true
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft:
		return false
	}
	pt := m.Pt().Sub(p.Bounds().Min)
	if t, ok := p.chrome.tabAt(pt); ok {
		if t < 3 {
			p.setRange(histRange(t))
		} else {
			p.setLog(t == 4)
		}
		return true
	}
	if n, ok := p.chrome.hitAt(hitHistRun, pt); ok {
		p.cursor = n
		p.pickRun()
		return true
	}
	if n, ok := p.chrome.hitAt(hitHistMeasure, pt); ok {
		p.row = n
		p.toggle(n)
		return true
	}
	return false
}

// followMouse moves the pointer to the sample under the mouse, when it is over the chart.
func (p *historyPanel) followMouse() {
	if p.v == nil {
		return
	}
	g := p.v.geo
	if len(p.v.ticks) == 0 || p.hover.Y < g.y0 || p.hover.Y >= g.axisY || p.hover.X < g.px0 || p.hover.X >= g.px0+g.pcols {
		return
	}
	t := p.v.nearest(p.v.tickAt(p.hover.X-g.px0, g.pcols))
	if t != p.v.ptr {
		p.ptr = t
		p.Layout(p.Bounds())
	}
}

// PointerAt: the hand over a tab, a key or a measure.
func (p *historyPanel) PointerAt(at image.Point) kit.Pointer {
	if p.chrome == nil {
		return kit.PointerArrow
	}
	pt := at.Sub(p.Bounds().Min)
	_, tab := p.chrome.tabAt(pt)
	_, key := p.chrome.keyAt(pt)
	_, row := p.chrome.hitAt(hitHistMeasure, pt)
	return handIf(tab || key || row)
}

// text puts a coloured word on the panel: over the chart's picture when
// there is one.
func (p *historyPanel) text(p2 *kit.Painter, org image.Point, x, y int, s string, col color.RGBA, bold bool) {
	st := kit.Style{}.Foreground(theme.RGB(col.R, col.G, col.B)).Bold(bold)
	if p.plain {
		p2.Print(org.X+x, org.Y+y, runeLen(s), kit.Escape(s), st)
		return
	}
	p2.Label(org.X+x, org.Y+y, kit.Escape(s), st)
}

// Paint puts the frame, then the chart, its words and the pointer's readout.
func (p *historyPanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.chrome == nil || p.v == nil {
		return
	}
	pal := newExpPalette()
	paintExpeditionKit(p2, p.chrome, r.Min, pal)
	v := p.v
	g, c, org := v.geo, v.ink, r.Min
	if g.pcols > 0 && g.picRows() > 3 {
		cw, ch := cellPixels()
		if p.plain {
			v.cells(p2, org.Add(image.Pt(g.px0, g.y0)))
			p.shapes(p2, org, 0, 0)
		} else {
			bounds := image.Rect(g.px0, g.y0, g.px0+g.pcols, g.axisY).Add(org)
			p2.Picture(bounds, p.picture(cw, ch))
			p.shapes(p2, org, cw, ch)
		}
		p.words(p2, org)
	}
	if p.rng == rangeAll && g.listTop > 0 {
		p.runs(p2, org)
	}
	p.buttons.paint(p2, p.chrome, r.Min, pal)
	_ = c
}

// picture is the chart's image, kept while nothing it draws has changed.
func (p *historyPanel) picture(cw, ch int) *image.RGBA {
	v := p.v
	key := fmt.Sprint(v.geo, v.lo, v.hi, v.end, v.ptr, v.ink, cw, ch, len(v.marks), len(v.spans), p.ghost, v.gEnd, v.logOn)
	for _, l := range v.lines {
		key += fmt.Sprint(l.slot, l.m.Key, l.lg, l.lo, l.hi, len(l.cur.V), l.cur.V[max(0, len(l.cur.V)-1):], len(l.gh.V))
	}
	img := p.cache[key]
	if img == nil {
		if len(p.cache) > 40 {
			clear(p.cache)
		}
		img = v.picture(cw, ch)
		p.cache[key] = img
	}
	return img
}

// shapes puts each slot's shape beside its measure in the picker.
func (p *historyPanel) shapes(p2 *kit.Painter, org image.Point, cw, ch int) {
	v := p.v
	rows, _ := p.pickerWindow()
	for j, r := range rows {
		if r.idx < 0 {
			continue
		}
		if slot := slices.Index(p.slots[:], p.ms[r.idx].Key); slot >= 0 {
			if p.plain {
				p.text(p2, org, 2, v.geo.y0+j, string([]rune("●■▲")[slot]), v.ink.slot(slot), false)
				continue
			}
			p2.Picture(image.Rect(2, v.geo.y0+j, 3, v.geo.y0+j+1).Add(org), shapeImg(slot, v.ink.slot(slot), v.ink.ground))
		}
	}
}

// words writes the age names, the axes, the legend and the readout.
func (p *historyPanel) words(p2 *kit.Painter, org image.Point) {
	v := p.v
	g, c := v.geo, v.ink
	for _, t := range v.tags() {
		col := toward(c.ageColor(t.idx), c.bright, .3)
		p.text(p2, org, g.px0+t.x, g.y0+1, t.name, col, t.now)
	}
	if len(v.lines) > 0 {
		l, rows := v.lines[0], []int{g.plotTop(), (g.plotTop() + g.axisY - 1) / 2, g.axisY - 1}
		fr := []float64{1, .5, 0}
		if l.lg {
			rows, fr = []int{g.plotTop(), g.axisY - 1}, []float64{1, 0}
		}
		for i, y := range rows {
			val := l.hi * fr[i]
			if l.lg {
				val = l.lo * math.Pow(l.hi/l.lo, fr[i])
			}
			s := l.m.text(val, false)
			p.text(p2, org, g.px0-1-runeLen(s), y, s, c.dim, false)
		}
	}
	p.axis(p2, org)
	p.legend(p2, org)
	p.readout(p2, org)
}

// axis is the time axis under the chart: round times, none twice, none on the
// pointer's time or past the edge.
func (p *historyPanel) axis(p2 *kit.Painter, org image.Point) {
	v := p.v
	g, c := v.geo, v.ink
	chip, chipX := "", 0
	if v.ptr >= 0 {
		chip = histTime(v.ptr)
		chipX = clamp(g.px0+v.cellX(v.ptr)-runeLen(chip)/2, g.px0, g.px0+g.pcols-runeLen(chip))
	}
	last := ""
	for _, t := range v.axisTicks() {
		s, x := histTime(t), g.px0+v.cellX(t)
		if s == last || x+runeLen(s) > g.w-2 || chip != "" && x < chipX+runeLen(chip)+1 && x+runeLen(s)+1 > chipX {
			continue
		}
		p.text(p2, org, x, g.axisY, s, c.dim, false)
		last = s
	}
	if chip != "" {
		p.text(p2, org, chipX, g.axisY, chip, c.bright, true)
	}
}

// legend names each shown measure with its true range.
func (p *historyPanel) legend(p2 *kit.Painter, org image.Point) {
	v := p.v
	g, c := v.geo, v.ink
	x := g.px0
	for _, l := range v.lines {
		var rng string
		switch {
		case l.lg:
			rng = rowid.Sprintf("ui.history_panel.legend.01", "log, %s to %s", l.m.text(l.lo, false), l.m.text(l.hi, true))
		case l.m.Cap > 0 && p.logs[p.rng]:
			rng = rowid.Sprintf("ui.history_panel.legend.02", "0 to %s, %s", l.m.text(l.hi, true), txtStaysLinear.String())
		default:
			rng = rowid.Sprintf("ui.history_panel.legend.03", "0 to %s", l.m.text(l.hi, true))
		}
		s := l.m.Label + "  " + rng
		if x+2+runeLen(s) > g.w-2 {
			break
		}
		if !p.plain {
			p2.Picture(image.Rect(x, g.legendY, x+1, g.legendY+1).Add(org), shapeImg(l.slot, c.slot(l.slot), c.ground))
		} else {
			p.text(p2, org, x, g.legendY, string([]rune("●■▲")[l.slot]), c.slot(l.slot), false)
		}
		p.text(p2, org, x+2, g.legendY, s, c.slot(l.slot), false)
		x += 2 + runeLen(s) + 3
	}
	if len(v.lines) > 1 && !v.all {
		s := txtOwnScale.String()
		if x+runeLen(s) <= g.w-2 {
			p.text(p2, org, g.w-2-runeLen(s), g.legendY, s, c.dim, false)
		}
	}
}

// readout is the box beside the pointer: the time, the age, and one row per
// shown measure.
func (p *historyPanel) readout(p2 *kit.Painter, org image.Point) {
	v := p.v
	g, c := v.geo, v.ink
	if v.ptr < 0 || len(v.lines) == 0 {
		return
	}
	bw, bh := 36, 3+len(v.lines)
	two := v.all && v.ghost != nil
	if two {
		bw = 50
	}
	if bw > g.pcols-2 || bh > g.plotRows()-1 {
		return
	}
	px := g.px0 + v.cellX(v.ptr)
	bx := px + 2
	if px-g.px0 > g.pcols/2 {
		bx = px - bw - 1
	}
	bx = clamp(bx, g.px0+1, g.px0+g.pcols-bw-1)
	by := g.plotTop() + 1
	if low := g.axisY - bh - 1; p.covers(bx, by, bw, bh) > p.covers(bx, low, bw, bh) {
		by = low
	}
	box := image.Rect(bx, by, bx+bw, by+bh).Add(org)
	surf := theme.RGB(c.surface.R, c.surface.G, c.surface.B)
	p2.Fill(box, surf)
	p2.Box(box, kit.BorderLine, kit.Style{}.Foreground(theme.RGB(c.dim.R, c.dim.G, c.dim.B)).Background(surf))
	put := func(x, y int, s string, col color.RGBA, bold bool) {
		p2.Print(org.X+x, org.Y+y, runeLen(s), kit.Escape(s), kit.Style{}.Foreground(theme.RGB(col.R, col.G, col.B)).Background(surf).Bold(bold))
	}
	head := rowid.Sprintf("ui.history_panel.readout.01", "At %s", histTime(v.ptr))
	put(bx+2, by+1, head, c.bright, true)
	if !v.all && len(v.spans) > 0 {
		age := v.spans[len(v.spans)-1]
		for _, sp := range v.spans {
			if v.ptr < sp.End {
				age = sp
				break
			}
		}
		name := v.ageName(age.Key)
		put(bx+bw-2-runeLen(name), by+1, name, c.ageColor(v.ageIdx(age.Key)), false)
	}
	for i, l := range v.lines {
		y := by + 2 + i
		val := "-"
		if x, ok := l.cur.at(v.ptr); ok {
			val = l.m.text(x, true)
		}
		if !p.plain {
			p2.Picture(image.Rect(bx+2, y, bx+3, y+1).Add(org), shapeImg(l.slot, c.slot(l.slot), c.surface))
		} else {
			put(bx+2, y, string([]rune("●■▲")[l.slot]), c.slot(l.slot), false)
		}
		put(bx+4, y, string([]rune(l.m.Label)[:min(runeLen(l.m.Label), 16)]), c.slot(l.slot), false)
		put(bx+bw-2-runeLen(val)-btoi(two)*14, y, val, c.bright, false)
		if two {
			gv := "-"
			if x, ok := l.gh.at(v.ptr); ok {
				gv = l.m.text(x, true)
			}
			put(bx+bw-2-runeLen(gv), y, gv, c.dim, false)
		}
	}
	if two {
		put(bx+bw-2-runeLen(txtKeyGhost.String()), by+1, txtKeyGhost.String(), c.dim, false)
	}
}

// covers counts the lines' points under a box of the chart: the readout goes
// where it hides fewer.
func (p *historyPanel) covers(bx, by, bw, bh int) int {
	v := p.v
	g, n := v.geo, 0
	for _, l := range v.lines {
		for x := bx; x < bx+bw; x++ {
			if val, ok := l.cur.at(v.tickAt(x-g.px0, g.pcols)); ok {
				y := g.plotTop() + 1 + int((1-l.frac(val))*float64(g.plotRows()-2))
				if y >= by && y < by+bh {
					n++
				}
			}
		}
	}
	return n
}
