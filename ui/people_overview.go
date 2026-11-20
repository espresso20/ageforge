package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/peopledash"
	"github.com/espresso20/ageforge/rowid"
)

// people_overview.go is the People panel's Overview: a band with everyone as
// blocks, a strip of what needs doing, one card for each group and a tile each
// for morale, food and homes. It reads the snapshot and the run's recorder;
// the picture is people_scene.go's, the buttons are cells on the grid.

const (
	peopleBigW, peopleBigH = 230, 54 // the smallest rectangle that gets the large layout
	peopleMinW, peopleMinH = 150, 38 // below this the panel says it is too small
	peopleTrendN           = 12      // samples a trend draws
)

// The words on the Overview.
var (
	txtAtWork      = rowid.E("ui.people_overview.words.01", "At work")
	txtWaiting     = rowid.E("ui.people_overview.words.02", "Waiting")
	txtSoldiers    = rowid.E("ui.people_overview.words.03", "Soldiers")
	txtScholars    = rowid.E("ui.people_overview.words.04", "Scholars")
	txtNeedsYou    = rowid.E("ui.people_overview.words.05", "Needs you")
	txtNothing     = rowid.E("ui.people_overview.words.06", "Nothing needs doing.")
	txtFoodTile    = rowid.E("ui.people_overview.words.08", "Food")
	txtHomesTile   = rowid.E("ui.people_overview.words.09", "Homes")
	txtOpenWork    = rowid.E("ui.people_overview.words.10", "Open Work")
	txtOpenArmy    = rowid.E("ui.people_overview.words.11", "Open Army")
	txtOpenScholar = rowid.E("ui.people_overview.words.12", "Open Scholars")
)

// moverWords name what moves morale, by the engine's key.
var moverWords = map[string]rowid.Entry{
	game.MoverStarving:  rowid.E("ui.people_overview.movers.01", "Hunger"),
	game.MoverScouting:  rowid.E("ui.people_overview.movers.02", "Too many scouting"),
	game.MoverWaiting:   rowid.E("ui.people_overview.movers.03", "Too many waiting"),
	game.MoverBuildings: rowid.E("ui.people_overview.movers.04", "Buildings"),
	game.MoverFaith:     rowid.E("ui.people_overview.movers.05", "Faith"),
	game.MoverSettling:  rowid.E("ui.people_overview.movers.06", "Settling toward 50%"),
}

// peopleBig says whether a w by h panel gets the large layout.
func peopleBig(w, h int) bool { return w >= peopleBigW && h >= peopleBigH }

// peopleOverview works the Overview out in a panel w by h: the picture and
// words, and the buttons and hits on g.
type peopleOverview struct {
	s    *pScene
	g    *xgrid
	st   game.GameState
	hist *game.RunHistory
	row  int // the picked card
	big  bool
	kind []peopleKind
}

// newOverview lays the Overview out below row y0 of g, above its key row.
func newOverview(g *xgrid, st game.GameState, hist *game.RunHistory, row, y0 int) *peopleOverview {
	o := &peopleOverview{s: newPScene(g.w, y0, g.h-2), g: g, st: st, hist: hist, row: row, big: peopleBig(g.w, g.h), kind: peopleKinds(st)}
	band, strip := 5, 2
	if o.big {
		band, strip = 8, 3
	}
	rest := g.h - 2 - y0 - band - strip - 3
	cards := rest * 45 / 100
	if o.big {
		cards = rest * 42 / 100
	}
	y := y0
	o.band(y, band)
	y += band + 1
	o.needs(y, strip)
	y += strip + 1
	o.cards(y, cards)
	y += cards + 1
	o.tiles(y, g.h-2-y)
	return o
}

// groupColour is a group's colour, by its place on the cards.
func (o *peopleOverview) groupColour(i int) color.RGBA {
	return [4]color.RGBA{o.s.pal.Work, o.s.pal.Waiting, o.s.pal.Soldiers, o.s.pal.Scholars}[i]
}

// span is the inside of the frame: columns 2 to w-3.
func (o *peopleOverview) span() (x, w int) { return 2, o.g.w - 4 }

// band is everyone, as a number and as blocks: each group in its colour, the
// empty homes hollow.
func (o *peopleOverview) band(y, h int) {
	s, ws := o.s, o.st.Workers
	x, w := o.span()
	s.rect(x-1, y, w+2, h, s.pal.Tile)
	scale := 3
	if o.big {
		scale = 5
	}
	c := s.num(x+1, y+1, num(ws.TotalPop), s.pal.Text, true, scale)
	s.tx(x+1, y+1+s.numH(true, scale), rowid.Sprintf("ui.people_overview.band.01", "of %s homes", num(ws.MaxPop)), s.pal.Dim, false)
	room := x + w - 1 - (c + 2)
	counts := [4]int{o.kind[0].count, o.kind[1].count, o.kind[2].count, o.kind[3].count}
	type seg struct{ n, fill int }
	var segs []seg
	total := 0
	for _, n := range counts {
		segs = append(segs, seg{peopledash.BlockCount(n), peopledash.BlockCount(n)})
		total += segs[len(segs)-1].n
	}
	empty := peopledash.BlockCount(max(0, ws.MaxPop-ws.TotalPop))
	segs = append(segs, seg{empty, 0})
	total += empty
	lines := max(1, (h-2)/3)
	pitch := 14
	if per := max(1, room*s.cw/pitch); total > lines*per {
		pitch = max(3, room*s.cw*lines/total)
	}
	per := max(1, room*s.cw/pitch)
	i := 0
	for k, sg := range segs {
		col := s.pal.Homes
		if k < 4 {
			col = o.groupColour(k)
		}
		for j := 0; j < sg.n; j++ {
			ln, at := i/per, i%per
			if ln < lines {
				o.s.cv.Blocks(s.px(c+2)+at*pitch, s.py(y+1+ln*3)+2, 2*s.ch-4, 1, min(1, sg.fill-j), col, s.pal.Tile, max(1, pitch-2), pitch)
			}
			i++
		}
	}
}

// need is one thing on the strip.
type need struct {
	head, detail string
	chip         string
	cmd          string
}

// needsList is what wants the player, most pressing first: growth paused, open
// jobs, morale falling, food short.
func (o *peopleOverview) needsList() []need {
	st, ws := o.st, o.st.Workers
	var out []need
	if ws.HeldBy == game.HeldWork {
		out = append(out, need{rowid.Line("ui.people_overview.needs.01", "Growth is paused"), rowid.Line("ui.people_overview.needs.02", "As many are waiting as are at work."), rowid.Line("ui.people_overview.needs.03", "Buildings"), "buildings"})
	}
	if ws.HeldBy == game.HeldHousing || ws.HeldBy == game.HeldFood {
		detail := ""
		if l := heldByLines(ws, st.TickIntervalMs, 0); len(l) > 1 {
			detail = l[1]
		}
		out = append(out, need{rowid.Sprintf("ui.people_overview.needs.12", "Growth is held by %s", ws.HeldBy), detail, rowid.Line("ui.people_overview.needs.13", "Buildings"), "buildings"})
	}
	if slots, filled := o.jobs(); slots > filled && ws.Reserve() > 0 {
		out = append(out, need{rowid.Sprintf("ui.people_overview.needs.04", "%s open jobs", num(slots-filled)), rowid.Sprintf("ui.people_overview.needs.05", "%s waiting could fill them.", num(ws.Reserve())), rowid.Line("ui.people_overview.needs.06", "Fill jobs"), "roster auto"})
	}
	if per := o.perMinute(); per < 0 {
		out = append(out, need{rowid.Line("ui.people_overview.needs.07", "Morale is falling"), rowid.Sprintf("ui.people_overview.needs.08", "%.1f points a minute.", -per*100), "", ""})
	}
	if rs, ok := st.Resources["food"]; ok && rs.Rate < 0 {
		out = append(out, need{rowid.Line("ui.people_overview.needs.09", "Food is short"), rowid.Sprintf("ui.people_overview.needs.10", "%s food.", rateText(rs.Rate, st.TickIntervalMs)), rowid.Line("ui.people_overview.needs.11", "Buildings"), "buildings"})
	}
	return out
}

// jobs is the work slots and how many are filled.
func (o *peopleOverview) jobs() (slots, filled int) {
	for _, r := range game.ShareRows(o.st) {
		slots += r.Slots
		filled += r.Workers
	}
	return
}

// perMinute is what the movers add up to, in morale points a minute.
func (o *peopleOverview) perMinute() float64 {
	t := 0.0
	for _, m := range o.st.MoraleMovers {
		t += m.PerMinute
	}
	return t
}

// needs is the strip: a triangle, the headline and its detail, and a chip when
// one command settles it.
func (o *peopleOverview) needs(y, h int) {
	s := o.s
	x, w := o.span()
	s.rect(x-1, y, w+2, h, s.pal.Tile)
	s.tx(x+1, y+h/2, txtNeedsYou.String(), s.pal.Dim, true)
	items := o.needsList()
	if len(items) == 0 {
		s.tx(x+16, y+h/2, txtNothing.String(), s.pal.Dim, false)
		return
	}
	col := (w - 16) / 3
	for i, n := range items[:min(3, len(items))] {
		cx := x + 16 + i*col
		s.cv.Warn(s.px(cx)+9, s.py(y)+(h*s.ch-15)/2, s.pal.Amber)
		tx := cx + 4
		room := col - 5
		if n.chip != "" {
			room -= runeLen(n.chip) + 3
			o.g.chip(cx+col-runeLen(n.chip)-3, y+h/2, n.chip, keySteel, n.cmd, "", "", false)
		}
		s.tx(tx, y+(h-1)/2, fitText(n.head, room), s.pal.Text, true)
		if h > 2 {
			s.tx(tx, y+(h-1)/2+1, fitText(n.detail, room), s.pal.Dim, false)
		}
	}
}

// cards is one card for each group.
func (o *peopleOverview) cards(y, h int) {
	s, ws := o.s, o.st.Workers
	x, w := o.span()
	slots, filled := o.jobs()
	names := [4]string{txtAtWork.String(), txtWaiting.String(), txtSoldiers.String(), txtScholars.String()}
	subs := [4]string{
		rowid.Sprintf("ui.people_overview.cards.01", "of %s jobs", num(slots)),
		rowid.Sprintf("ui.people_overview.cards.02", "of %s people", num(ws.TotalPop)),
		rowid.Sprintf("ui.people_overview.cards.03", "%s out on missions", num(o.kind[2].out)),
		rowid.Sprintf("ui.people_overview.cards.04", "%s researching, %s teaching", num(o.st.Scholars.Researching), num(o.st.Scholars.Teaching)),
	}
	posts := max(o.kind[2].posts, o.kind[2].count)
	links := [4]struct {
		word string
		tab  int
	}{{txtOpenWork.String(), tabWork}, {txtOpenWork.String(), tabWork}, {txtOpenArmy.String(), tabArmy}, {txtOpenScholar.String(), tabScholars}}
	numScale, pctScale, bodyRow := 3, 3, 5
	if o.big {
		numScale, bodyRow = 5, 9
	}
	cw := (w - 3) / 4
	for i, k := range o.kind {
		cx := x + i*(cw+1)
		if i == 3 {
			cw = x + w - cx
		}
		col := o.groupColour(i)
		s.tile(cx, y, cw, h, peopledash.Mix(s.pal.Tile, col, .18), col)
		if i == o.row {
			s.frame(cx, y, cw, h, col)
		}
		o.g.hit(hitRow, i, cx, y, cw, h)
		s.num(cx+2, y+1, num(k.count), s.pal.Text, true, numScale)
		share := 0
		if ws.TotalPop > 0 {
			share = int(math.Round(float64(k.count) * 100 / float64(ws.TotalPop)))
		}
		pt := fmt.Sprintf("%d%%", share)
		pw := (peopledash.NumW(pt, o.big, pctScale) + s.cw - 1) / s.cw
		s.num(cx+cw-pw-2, y+1, pt, s.pal.Dim, o.big, pctScale)
		nr := y + 1 + s.numH(true, numScale)
		// A short card drops the count line so the blocks keep clear of the words.
		step, tr := 3, nr
		if o.big {
			s.tx(cx+2, tr, fitText(subs[i], cw-4), s.pal.Dim, false)
			tr++
		} else {
			step = 2
		}
		s.tx(cx+2, tr, fitText(names[i], cw-4), s.pal.Text, true)
		s.tx(cx+2, tr+1, fitText(k.name, cw-4), s.pal.Dim, false)
		// The blocks: jobs filled of jobs, posts held of posts, the rest all filled.
		n, fill := peopledash.BlockCount(k.count), peopledash.BlockCount(k.count)
		switch i {
		case 0:
			n, fill = peopledash.BlockCount(slots), peopledash.BlockCount(filled)
		case 2:
			n = peopledash.BlockCount(posts)
		}
		by, bh := tr+3, max(2, y+h-3-(tr+3))
		if !o.big {
			by, bh = tr+2, max(2, y+h-3-(tr+2))
		}
		if o.big {
			by = y + bodyRow
			bh = max(2, y+h-3-by)
		}
		per := max(1, (cw-4)*s.cw/14)
		rows := max(1, bh/step)
		pitch := 14
		if n > rows*per {
			pitch = max(3, (cw-4)*s.cw*rows/n)
			per = max(1, (cw-4)*s.cw/pitch)
		}
		for j := 0; j < n; j++ {
			ln, at := j/per, j%per
			if ln < rows {
				f := 0
				if j < fill {
					f = 1
				}
				s.cv.Blocks(s.px(cx+2)+at*pitch, s.py(by+ln*step)+2, step*s.ch-4, 1, f, col, s.pal.Tile, max(1, pitch-2), pitch)
			}
		}
		o.g.peopleLink(cx+2, y+h-2, cw-4, links[i].word, links[i].tab)
	}
}

// samples are a trend's values: the recorder's newest, with the last one
// replaced by the figure now.
func (o *peopleOverview) samples(key string, live float64) []float64 {
	if o.hist == nil {
		return nil
	}
	col := o.hist.Recent.Cols[key]
	n := min(o.hist.Recent.N, len(col))
	if n < 2 {
		return nil
	}
	col = col[max(0, n-peopleTrendN):n]
	v := make([]float64, len(col))
	for i, f := range col {
		v[i] = float64(f)
	}
	v[len(v)-1] = live
	return v
}

// tiles are the morale, food and homes tiles.
func (o *peopleOverview) tiles(y, h int) {
	x, w := o.span()
	mw, fw := w*42/100, w*29/100
	hw := w - mw - fw - 2
	o.morale(x, y, mw, h)
	o.food(x+mw+1, y, fw, h)
	o.homes(x+mw+fw+2, y, hw, h)
}

// tile paints the ground of a tile with its label.
func (o *peopleOverview) tile(x, y, w, h int, col color.RGBA, label string) {
	o.s.tile(x, y, w, h, o.s.pal.Tile, col)
	o.s.tx(x+2, y+1, label, o.s.pal.Text, true)
}

func (o *peopleOverview) morale(x, y, w, h int) {
	s, st := o.s, o.st
	zone := s.pal.MoraleZone(st.Morale)
	o.tile(x, y, w, h, zone, txtMorale.String())
	floor, top := 0.10, st.MoraleCap
	if top <= floor {
		top = 1
	}
	r := clamp((h-4)*s.ch-70, 40, 170)
	cx, cy := s.px(x+2)+r+12, s.py(y+2)+r+14
	s.cv.Dial(cx, cy, r, st.Morale, floor, top)
	rc := x + 2 + (2*r+30)/s.cw + 2
	rw := x + w - 2 - rc
	if rw < 14 {
		return
	}
	per := o.perMinute()
	word, down, moving := rowid.Line("ui.people_overview.morale.01", "Steady"), false, true
	switch {
	case per < -0.0005:
		word, down = rowid.Line("ui.people_overview.morale.02", "Falling"), true
	case per > 0.0005:
		word = rowid.Line("ui.people_overview.morale.03", "Rising")
	default:
		moving = false
	}
	if moving {
		s.cv.Arrow(s.px(rc)+9, s.py(y+1)+2, down, zone)
		s.tx(rc+4, y+1, word, s.pal.Text, true)
	} else {
		s.tx(rc, y+1, word, s.pal.Text, true)
	}
	sign := "+"
	if per < 0 {
		sign = "-"
	}
	c := s.num(rc, y+3, fmt.Sprintf("%s%.1f", sign, math.Abs(per)*100), zone, false, 4)
	s.tx(c, y+4, rowid.Line("ui.people_overview.morale.04", "points a minute"), s.pal.Dim, false)
	row := y + 6
	if th := 4; o.big && y+h-row > th+8 {
		if v := o.samples("morale", st.Morale); v != nil {
			s.cv.Trend(s.px(rc), s.py(row), rw*s.cw, th*s.ch, v, floor, top, zone, s.pal.Tile, .4, .5)
		}
		row += th + 1
	}
	for _, m := range st.MoraleMovers {
		if row >= y+h-3 {
			break
		}
		name := moverWords[m.Key].String()
		s.tx(rc, row, fitText(name, rw-9), s.pal.Dim, false)
		s.txR(rc+rw, row, fmt.Sprintf("%+.1f", m.PerMinute*100), s.pal.Text, false)
		row++
	}
	if d := computeMoraleBand(st.Morale, st.MoraleMultiplier).DeltaLabel; d != "" {
		s.tx(rc, y+h-2, fitText(rowid.Sprintf("ui.people_overview.morale.05", "All production %s", d), rw), s.pal.Text, true)
	}
}

func (o *peopleOverview) food(x, y, w, h int) {
	s, st := o.s, o.st
	rs := st.Resources["food"]
	o.tile(x, y, w, h, s.pal.Work, txtFoodTile.String())
	col := s.pal.Green
	if rs.Rate < 0 {
		col = s.pal.Red
	}
	figure, unit := rateFigure(rs.Rate, st.TickIntervalMs)
	scale := 3
	if o.big {
		scale = 5
	}
	c := s.num(x+2, y+3, figure, col, true, scale)
	s.tx(c, y+3+s.numH(true, scale)-1, unit, s.pal.Dim, false)
	drain := st.Workers.FoodDrain
	made := rs.Rate + drain
	frac := 0.0
	if made > 0 {
		frac = drain / made
	}
	row := y + 4 + s.numH(true, scale)
	s.bar(x+2, row, w-4, frac, col)
	row += 2
	s.tx(x+2, row, fitText(rowid.Sprintf("ui.people_overview.food.01", "Everyone eats %s", rateAmount(drain, st.TickIntervalMs)), w-4), s.pal.Text, false)
	if per := drain / math.Max(1, float64(st.Workers.TotalPop)); per > 0 {
		if grown := game.FoodBeforeMastery(rs.Rate, rs.Breakdown); grown > 0 && rs.Rate >= 0 {
			s.tx(x+2, row+1, fitText(rowid.Sprintf("ui.people_overview.food.02", "Would feed %s", num(int(grown/per))), w-4), s.pal.Dim, false)
		}
	}
	if th := y + h - 2 - (row + 3); th >= 3 {
		if v := o.samples("food_rate", rs.Rate); v != nil {
			lo, hi := v[0], v[0]
			for _, f := range v {
				lo, hi = math.Min(lo, f), math.Max(hi, f)
			}
			s.cv.Trend(s.px(x+2), s.py(row+3), (w-4)*s.cw, th*s.ch, v, lo-(hi-lo)*.1-1e-9, hi+(hi-lo)*.1+1e-9, col, s.pal.Tile, 0)
		}
	}
}

func (o *peopleOverview) homes(x, y, w, h int) {
	s, ws := o.s, o.st.Workers
	o.tile(x, y, w, h, s.pal.Homes, txtHomesTile.String())
	free := max(0, ws.MaxPop-ws.TotalPop)
	scale := 3
	if o.big {
		scale = 5
	}
	c := s.num(x+2, y+3, num(free), s.pal.Text, true, scale)
	s.tx(c, y+3+s.numH(true, scale)-1, rowid.Line("ui.people_overview.homes.01", "empty"), s.pal.Dim, false)
	row := y + 4 + s.numH(true, scale)
	s.bar(x+2, row, w-4, float64(ws.TotalPop)/math.Max(1, float64(ws.MaxPop)), s.pal.Homes)
	row += 2
	cx := x + 2
	for _, ch := range []struct{ word, cmd string }{{"+1", "recruit 1"}, {"+10", "recruit 10"}, {rowid.Line("ui.people_overview.homes.02", "Max"), "recruit max"}} {
		cx = o.g.chip(cx, row, ch.word, keyPrimary, ch.cmd, "", "", free == 0) + 1
	}
	auto, on := rowid.Line("ui.people_overview.homes.03", "Auto growth on"), "workers auto-recruit off"
	if !ws.AutoRecruit {
		auto, on = rowid.Line("ui.people_overview.homes.04", "Auto growth off"), "workers auto-recruit on"
	}
	o.g.chip(cx+1, row, auto, keySteel, on, "", "", false)
	s.tx(x+2, row+2, fitText(recruitText(o.st), w-4), s.pal.Dim, false)
}
