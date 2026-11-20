package ui

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// trade_board.go is the Trade panel's board: one entry per route on the left
// and the goods table with the offers at the right (trade_side.go). A panel
// too small for it falls back to the plain card deck.

const (
	tradeMinW, tradeMinH = 247, 24
	tradeSideW           = 89
)

// The route columns, from the board's left edge.
const (
	tcState, tcName, tcSends, tcArrow, tcBrings, tcShare, tcRun = 4, 15, 35, 56, 58, 91, 133
)

// tradeFlow is one good a route moves, as the route runs now: per tick and
// per minute, and what the town itself makes of it in a minute.
type tradeFlow struct {
	key, good        string
	tick, min, make_ float64
	stock            float64
}

// tradeRoute is one route as the board shows it.
type tradeRoute struct {
	key, name, state, cmd, why string
	out, in                    []tradeFlow
	shares                     map[string][]game.PayShare
	base                       map[string]float64
	baseTicks, size, minCount  int
	run, scale                 float64
	speed                      float64
	shortOf                    string
	share                      string
	by, lift                   string
	war                        bool
}

// moving is true while the route moves goods, at all its flow or less.
func (r tradeRoute) moving() bool { return r.state == "running" || r.state == "short" }

func tradeGoodName(res string) string {
	if res == config.ResearchTarget {
		return game.Amount(0, res)
	}
	return game.ResourceName(res)
}

// tradeTownMake is what the town makes of a good in a minute, trade left out.
func tradeTownMake(st game.GameState, res string, tpm float64) float64 {
	rs := st.Resources[res]
	return float64((rs.Rate - rs.Breakdown.TradeRate - rs.Breakdown.TradeFlow) * tpm)
}

// tradeRoutes lists the routes the town has open or running, in the order
// they unlock, each with the state the board shows.
func tradeRoutes(st game.GameState) []tradeRoute {
	set := st.Ruleset()
	tpm := 60 / tickInterval(st).Seconds()
	_, locked := featureRefusal(st, config.FeatureTradeRoutes)
	active := map[string]game.ActiveRouteInfo{}
	for _, a := range st.Trade.ActiveRoutes {
		active[a.Key] = a
	}
	avail := map[string]game.TradeRouteInfo{}
	for _, v := range st.Trade.AvailableRoutes {
		avail[v.Key] = v
	}
	flows := func(m map[string]float64, k float64) (out []tradeFlow) {
		for _, res := range sortedKeysOf(m) {
			out = append(out, tradeFlow{key: res, good: tradeGoodName(res), tick: float64(m[res] * k), min: float64(m[res]*k) * tpm,
				make_: tradeTownMake(st, res, tpm), stock: st.Resources[res].Amount})
		}
		return out
	}
	idx := map[string]int{}
	var rs []tradeRoute
	for _, def := range set.TradeRoutes() {
		a, on := active[def.Key]
		v, off := avail[def.Key]
		if !on && !off {
			continue
		}
		idx[def.Key], _ = set.Index(def.MinAge)
		r := tradeRoute{key: def.Key, name: def.Name, base: def.Import, baseTicks: def.TicksPerRun, minCount: def.MinCount, shares: v.Shares,
			run: 1, scale: 1, speed: st.Trade.Speed}
		out, in := v.Export, v.Import
		if on {
			out, in, r.shares, r.run, r.scale = a.Export, a.Import, a.Shares, a.Run, a.Scale
		}
		// A route held or paused shows what it would move.
		k := max(r.scale, 1)
		if on && !a.Paused && !a.NoBuilding && !a.Disrupted {
			k = float64(r.run * r.scale)
		}
		r.out, r.in = flows(out, k), flows(in, k)
		r.size = int(r.scale*float64(def.MinCount) + 0.5)
		for _, f := range r.in {
			if rs := st.Resources[f.key]; f.key != config.ResearchTarget && r.share == "" && rs.Rate > 0 && f.tick > 0 {
				r.share = rowid.Sprintf("ui.trade_board.tradeRoutes.01", "%.0f%% of %s", math.Min(100, f.tick/rs.Rate*100), f.good)
			}
		}
		switch {
		case on && a.Disrupted:
			r.state, r.cmd = "suspended", "trade route stop "+def.Key
			r.why = rowid.Sprintf("ui.overlay_trade.tradePage.15", "Suspended: %s imports are blockaded", game.ResourceName(a.DisruptedBy))
			for _, fk := range sortedKeysOf(st.Diplomacy.Factions) {
				if f := st.Diplomacy.Factions[fk]; f.Discovered && (f.AtWar || f.Status == "embargo") && f.Specialty == a.DisruptedBy {
					r.by, r.war, r.lift = f.Name, f.AtWar, "diplomacy neutral "+fk
					if f.AtWar {
						r.lift = "diplomacy tribute " + fk
					}
				}
			}
		case on && a.Paused:
			r.state, r.cmd = "paused", "trade route start "+def.Key
		case on && a.NoBuilding:
			r.state, r.cmd = "suspended", "trade route stop "+def.Key
			bld := st.Buildings[def.RequiredBld]
			r.why = rowid.Sprintf("ui.trade_board.tradeRoutes.02", "Held: needs %s (have %s)", game.BuildingCount(def.MinCount, def.RequiredBld), textfmt.Int(bld.Count))
		case on && a.Run < 1:
			r.state, r.cmd = "short", "trade route stop "+def.Key
			low := math.Inf(1)
			for _, f := range r.out {
				if f.tick > 0 && f.stock/f.tick < low {
					low, r.shortOf = f.stock/f.tick, f.good
				}
			}
		case on:
			r.state, r.cmd = "running", "trade route stop "+def.Key
		case v.CanStart && !locked:
			r.state, r.cmd = "ready", "trade route start "+def.Key
		default:
			r.state = "locked"
			if !v.CanStart {
				bld := st.Buildings[def.RequiredBld]
				r.why = rowid.Sprintf("ui.overlay_trade.tradePage.23", "needs %s (have %s)", game.BuildingCount(def.MinCount, def.RequiredBld), textfmt.Int(max(bld.Count, bld.HeldWithLater)))
			} else {
				r.why, _ = featureRefusal(st, config.FeatureTradeRoutes)
			}
		}
		rs = append(rs, r)
	}
	sort.SliceStable(rs, func(i, j int) bool { return idx[rs[i].key] < idx[rs[j].key] })
	return rs
}

// tradeLook is a state's word, its colour in text and its colour as a plate.
func tradeLook(state string) (word string, k, plate xk) {
	switch state {
	case "running":
		return rowid.Line("ui.overlay_trade.tradePage.10", "running"), xGood, xChipShort
	case "ready":
		return rowid.Line("ui.overlay_trade.tradePage.20", "ready"), xAccent, xChipCivic
	case "short":
		return rowid.Line("ui.trade_board.tradeLook.01", "short"), xGold, xChipLong
	case "suspended":
		return rowid.Line("ui.overlay_trade.tradePage.14", "suspended"), xBad, xChipMil
	case "paused":
		return rowid.Line("ui.trade_board.tradeLook.02", "paused"), xDim, xKeyName
	}
	return rowid.Line("ui.overlay_trade.tradePage.22", "locked"), xDim, xKeyName
}

// plate draws a word as a one-row button that runs cmd; it returns the next x.
func (g *xgrid) plate(x, y int, word string, k xk, cmd string) int {
	end := g.put(x, y, " "+word+" ", k)
	if cmd != "" {
		set := g.keyset()
		set.keys = append(set.keys, xkey{r: image.Rect(x, y, end, y+1), cmd: cmd, kind: keyPlain})
	}
	return end
}

// tradeNetCell says what trade does to a good against what the town makes.
func tradeNetCell(make_, takes, brings float64) (string, xk) {
	net, k := brings-takes, xDim
	if takes == 0 && brings == 0 {
		return "·", k
	}
	var s string
	switch {
	case make_ <= 0:
		s = textfmt.Signed(net) + textfmt.PerMinute
	case math.Abs(net/make_*100) < 0.05:
		return rowid.Line("ui.trade_board.tradeNetCell.01", "under 0.1% of make"), k
	default:
		s = rowid.Sprintf("ui.trade_board.tradeNetCell.02", "%+.1f%% of make", net/make_*100)
	}
	if net > 0 {
		k = xGood
	} else if net < 0 {
		k = xBad
	}
	return s, k
}

// tradeJoin is a route's goods as text, as a rate a second or an amount a
// minute; sign is -1 for what a route takes.
func tradeJoin(fl []tradeFlow, minute bool, sign float64, tickMs int) string {
	var ps []string
	for _, f := range fl {
		switch {
		case f.key == config.ResearchTarget:
			ps = append(ps, game.Amount(0, f.key))
		case f.tick == 0:
		case minute:
			ps = append(ps, game.Amount(f.min, f.key))
		default:
			num, unit := rateFigure(float64(sign*f.tick), tickMs)
			ps = append(ps, num+" "+f.good+unit)
		}
	}
	return strings.Join(ps, ", ")
}

// tradeTimeCuts are the researched techs that shorten a run, by name.
func tradeTimeCuts(st game.GameState) (cuts []game.PayShare) {
	for _, def := range st.Ruleset().Techs() {
		for _, e := range def.Effects {
			if st.Research.Techs[def.Key].Researched && e.Kind == config.EffectMechanic && e.Target == config.MechanicRouteTicks {
				cuts = append(cuts, game.PayShare{Label: def.Name, Percent: e.Value * 100})
			}
		}
	}
	return cuts
}

type tradeExtra struct {
	label string
	rows  [][]xseg
}

// tradeWrap lays units of text out in rows from x0 to lim, a gap apart.
func tradeWrap(units [][]xseg, x0, lim, gap int) (rows [][]xseg) {
	x := x0
	for _, u := range units {
		w := 0
		for _, s := range u {
			w += runeLen(s.s)
		}
		if len(rows) == 0 || x+w > lim && x > x0 {
			rows, x = append(rows, nil), x0
		} else {
			rows[len(rows)-1] = append(rows[len(rows)-1], xseg{strings.Repeat(" ", gap), xText})
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], u...)
		x += w + gap
	}
	return rows
}

// tradeExtras is what the picked route says under itself.
func tradeExtras(st game.GameState, r tradeRoute, lim int) (ex []tradeExtra) {
	text := func(ss ...string) (u [][]xseg) {
		for _, s := range ss {
			u = append(u, []xseg{{s, xText}})
		}
		return u
	}
	label, goods := rowid.Line("ui.trade_board.tradeExtras.01", "amount"), 0
	tpm := 60 / tickInterval(st).Seconds()
	for _, f := range r.in {
		if sh := r.shares[f.key]; len(sh) > 0 && f.key != config.ResearchTarget && goods < 2 {
			goods++
			base := float64(r.base[f.key]*tpm) / float64(max(r.baseTicks, 1))
			ss, total := []string{rowid.Sprintf("ui.trade_board.tradeExtras.02", "%s at base", game.Amount(base, f.key))}, 100.0
			for _, p := range sh {
				ss = append(ss, textfmt.Signed(p.Percent)+"% "+p.Label)
				total += p.Percent
			}
			ss = append(ss, "= "+game.Amount(float64(base*total/100), f.key))
			ex, label = append(ex, tradeExtra{label, tradeWrap(text(ss...), tcSends, lim, 2)}), ""
		}
	}
	if cuts := tradeTimeCuts(st); len(cuts) > 0 {
		ss := []string{}
		for _, p := range cuts {
			ss = append(ss, textfmt.Signed(p.Percent)+"% "+p.Label)
		}
		ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.04", "time"), tradeWrap(text(append(ss, rowid.Sprintf("ui.trade_board.tradeExtras.12", "= flow x%.2f", r.speed))...), tcSends, lim, 2)})
	}
	if r.size > 0 && r.state != "ready" && r.state != "locked" {
		ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.11", "size"), [][]xseg{{{rowid.Sprintf("ui.trade_board.tradeExtras.13", "%d of %d buildings: flow x%.2f the gate's", r.size, r.minCount, r.scale), xText}}}})
	}
	var cells [][]xseg
	for i, fl := range [][]tradeFlow{r.out, r.in} {
		for _, f := range fl {
			if f.make_ <= 0 {
				continue
			}
			takes, brings := f.min, 0.0
			if i == 1 {
				takes, brings = 0, f.min
			}
			s, k := tradeNetCell(f.make_, takes, brings)
			cells = append(cells, []xseg{{f.good + " ", xText}, {s, k}})
		}
	}
	if len(cells) > 0 {
		ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.05", "effect"), tradeWrap(cells, tcSends, lim, 3)})
	}
	if r.state == "suspended" && r.by != "" {
		cause := rowid.Sprintf("ui.trade_board.tradeExtras.06", "embargo on %s", r.by)
		if r.war {
			cause = rowid.Sprintf("ui.trade_board.tradeExtras.07", "war with %s", r.by)
		}
		ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.08", "cause"), [][]xseg{{{cause, xText}}}})
		if r.lift != "" {
			ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.09", "to lift"), [][]xseg{{{r.lift, xRoute}}}})
		}
	}
	if r.cmd != "" {
		ex = append(ex, tradeExtra{rowid.Line("ui.trade_board.tradeExtras.10", "type"), [][]xseg{{{r.cmd, xRoute}}}})
	}
	return ex
}

func tradeExtraRows(ex []tradeExtra) (n int) {
	for _, e := range ex {
		n += len(e.rows)
	}
	return n
}

// drawTradeEntry draws one route at row y, its extras if it is picked, and
// returns the row after it.
func drawTradeEntry(g *xgrid, st game.GameState, r tradeRoute, ex []tradeExtra, y, end, n int, picked bool) int {
	word, k, _ := tradeLook(r.state)
	rows := 2 + tradeExtraRows(ex)
	if picked {
		g.band(2, y, end-2, rows)
	}
	g.hit(hitDeckCard, n, 2, y, end-2, rows)
	g.put(2, y, "▌", k)
	g.put(tcState, y, word, k)
	nk := xBright
	if r.state == "locked" {
		nk = xDim
	}
	g.putFit(tcName, y, tcSends-tcName-1, r.name, nk)
	g.putFit(tcSends, y, tcArrow-tcSends-1, tradeJoin(r.out, false, -1, st.TickIntervalMs), xText)
	g.put(tcArrow, y, "►", xDim)
	g.putFit(tcBrings, y, tcShare-tcBrings-1, tradeJoin(r.in, false, 1, st.TickIntervalMs), xText)
	if r.moving() {
		g.putFit(tcShare, y, tcRun-tcShare-8, r.share, xText)
		g.rput(tcRun, y, fmt.Sprintf("%.0f%%", r.run*100), k)
	}
	if r.cmd != "" {
		word, pk := rowid.Line("ui.trade_board.drawTradeEntry.01", "stop"), xKeyName
		if r.state == "ready" || r.state == "paused" {
			word, pk = rowid.Line("ui.trade_board.drawTradeEntry.02", "start"), xChipCivic
		}
		bx := end - 1 - runeLen(word) - 2
		g.plate(bx, y, word, pk, r.cmd)
		if picked {
			g.put(bx-8, y, " Enter ", xKeycap)
		}
	}
	ms := st.TickIntervalMs
	switch r.state {
	case "running", "ready":
		g.put(tcName, y+1, rowid.Line("ui.trade_board.drawTradeEntry.03", "a minute"), xDim)
		g.putFit(tcSends, y+1, tcBrings-tcSends-3, tradeJoin(r.out, true, -1, ms), xDim)
		g.putFit(tcBrings, y+1, tcShare-tcBrings-1, tradeJoin(r.in, true, 1, ms), xDim)
	case "short":
		g.putFit(tcName, y+1, end-tcName-2, rowid.Sprintf("ui.trade_board.drawTradeEntry.04", "short of %s: the route runs at %.0f%% of its flow", r.shortOf, r.run*100), xGold)
	case "paused":
		g.putFit(tcName, y+1, end-tcName-2, rowid.Line("ui.trade_board.drawTradeEntry.05", "Paused: it moves nothing until started"), xDim)
	default:
		g.putFit(tcName, y+1, end-tcName-2, r.why, k)
	}
	y += 2
	for _, e := range ex {
		for i, row := range e.rows {
			if i == 0 {
				g.put(tcName, y, e.label, xDim)
			}
			g.segsFit(tcSends, y, end-tcSends-14, row...)
			y++
		}
	}
	return y
}

// tradeNextAge is the next age that opens routes, and how many it opens.
func tradeNextAge(st game.GameState) (age string, n int) {
	set := st.Ruleset()
	cur, _ := set.Index(st.Age)
	next := -1
	for _, def := range set.TradeRoutes() {
		i, _ := set.Index(def.MinAge)
		if i > cur && (next < 0 || i < next) {
			next, age, n = i, set.Name(rules.KindAge, def.MinAge), 0
		}
		if i == next {
			n++
		}
	}
	return age, n
}

// tradeMarketRows lays the black market out at x0 from row y: its rule, its
// state, its terms and what a win pays. It returns the row after it; with
// draw false it only counts.
func tradeMarketRows(g *xgrid, c deckCard, x0, y, w int, draw bool) int {
	y0 := y
	k := xDim
	if c.Chip == rowid.Line("ui.overlay_trade.tradePage.20", "ready") {
		k = xAccent
	}
	if draw {
		x := g.put(x0, y, c.Title, xBright)
		g.put(x+1, y, strings.Repeat("─", max(0, x0+w-x-1)), xDim)
		g.put(x0, y+1, c.Chip, k)
	}
	y += 2
	for _, l := range []string{c.MeterText, c.Why} {
		if l != "" {
			if draw {
				g.putFit(x0, y, w, l, xText)
			}
			y++
		}
	}
	for _, l := range c.Lines {
		if draw {
			g.putFit(x0, y, w, l.Label+": "+l.Value, xText)
		}
		y++
	}
	if len(c.Chips) > 0 {
		if draw {
			g.put(x0, y, rowid.Line("ui.trade_board.tradeMarketRows.01", "a win pays one of:"), xDim)
		}
		words := make([]string, len(c.Chips))
		for i, ch := range c.Chips {
			words[i] = ch.Word
		}
		for _, l := range deckWrap(strings.Join(words, " · "), w) {
			y++
			if draw {
				g.putFit(x0, y, w, l, xText)
			}
		}
		y++
	}
	return y - y0
}

// renderTradeBoard draws the Trade panel: the routes at the left, the goods
// table and the offers at the right. Too small a panel gets the card deck.
func renderTradeBoard(pg *deckPage, st game.GameState, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout) {
	routes := tradeRoutes(st)
	end := w - tradeSideW - 7 // the board's right edge
	if len(routes) == 0 || w < tradeMinW || h < tradeMinH {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}
	var market *deckCard
	if sec, ok := blackMarketSection(st); ok {
		market = &sec.Cards[0]
	}
	var ex []tradeExtra
	if sel < len(routes) {
		ex = tradeExtras(st, routes[sel], end-16)
	}
	// The left side does not scroll: what does not fit with no blank row
	// between routes sends the panel to the deck.
	age, more := tradeNextAge(st)
	running, tail, mh := 0, 0, 0
	for _, r := range routes {
		if r.moving() {
			running++
		}
	}
	if running == 0 {
		tail += 3
	}
	if more > 0 {
		tail += 2
	}
	if market != nil {
		mh = tradeMarketRows(nil, *market, 0, 0, end-17, false)
		tail += 2 + mh
	}
	used := 4 + len(routes)*2 + tradeExtraRows(ex) + tail // the last row, with no gaps
	gap := 0
	if used+len(routes)-1 <= h-4 {
		gap = 1
	}
	if used > h-4 {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}

	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, pg.Title, xTitle, false)
	g.closeButton(w)
	lay := deckLayout{total: h - deckChrome, viewH: h - deckChrome}

	// The count of each state, and a button that starts every ready route.
	count := map[string]int{}
	for _, r := range routes {
		count[r.state]++
	}
	x := g.put(2, 1, rowid.Line("ui.trade_board.renderTradeBoard.01", "Routes"), xBright) + 2
	for _, s := range []string{"running", "ready", "short", "paused", "suspended", "locked"} {
		if count[s] > 0 {
			if x > 10 {
				x = g.put(x, 1, " · ", xDim)
			}
			word, k, _ := tradeLook(s)
			x = g.put(x, 1, fmt.Sprintf("%d %s", count[s], word), k)
		}
	}
	if n := count["ready"]; n >= 2 {
		word := rowid.Sprintf("ui.trade_board.renderTradeBoard.02", "start all ready (%d)", n)
		g.plate(end-1-runeLen(word)-2, 1, word, xChipCivic, "trade route start all")
	}
	g.put(2, 4, strings.Repeat("─", end-2), xDim)
	for _, hd := range []struct {
		x int
		s string
	}{{tcState, rowid.Line("ui.trade_board.renderTradeBoard.03", "STATE")}, {tcName, rowid.Line("ui.trade_board.renderTradeBoard.04", "ROUTE")},
		{tcSends, rowid.Line("ui.trade_board.renderTradeBoard.05", "SENDS")}, {tcBrings, rowid.Line("ui.trade_board.renderTradeBoard.06", "BRINGS")},
		{tcShare, rowid.Line("ui.trade_board.renderTradeBoard.07", "SHARE OF INCOME")}} {
		g.put(hd.x, 3, hd.s, xAccent)
	}
	g.rput(tcRun, 3, rowid.Line("ui.trade_board.renderTradeBoard.09", "RUN"), xAccent)
	// The whole of trade a second: what the routes take and bring in all.
	var takes, brings float64
	for _, r := range routes {
		for i, fl := range [][]tradeFlow{r.out, r.in} {
			for _, f := range fl {
				if r.moving() && f.key != config.ResearchTarget && i == 0 {
					takes += f.tick
				} else if r.moving() && f.key != config.ResearchTarget {
					brings += f.tick
				}
			}
		}
	}
	if takes+brings > 0 {
		g.put(2, 2, rowid.Sprintf("ui.trade_board.renderTradeBoard.14", "Trade in %s, out %s (goods a tick, all routes)", rateAmount(brings, st.TickIntervalMs), rateAmount(takes, st.TickIntervalMs)), xDim)
	}

	y := 5
	for i, r := range routes {
		var rex []tradeExtra
		if i == sel {
			rex = ex
		}
		y = drawTradeEntry(g, st, r, rex, y, end, i, i == sel)
		lay.cards = append(lay.cards, image.Rect(0, -10000+2*i, 1, -9999+2*i))
		if i < len(routes)-1 {
			y += gap
		}
	}
	y--
	if running == 0 {
		y += 3
		g.put(tcName, y, rowid.Line("ui.trade_board.renderTradeBoard.10", "A route swaps one good for another on its own, all the time, until you pause it."), xDim)
	}
	if more > 0 {
		y += 2
		g.put(tcName, y, rowid.Sprintf("ui.trade_board.renderTradeBoard.11", "%s: %d more %s", age, more, textfmt.Plural(more, rowid.Line("ui.trade_board.renderTradeBoard.12", "route"), rowid.Line("ui.trade_board.renderTradeBoard.13", "routes"))), xDim)
	}
	if market != nil {
		y += 3
		mi := len(routes)
		if sel == mi {
			g.band(tcName-1, y, end-tcName, mh)
		}
		tradeMarketRows(g, *market, tcName, y, end-17, true)
		g.hit(hitDeckCard, mi, tcName-1, y, end-tcName, mh)
		lay.cards = append(lay.cards, image.Rect(0, -10000+2*mi, 1, -9999+2*mi))
	}

	base := len(lay.cards)
	side := drawTradeSide(g, st, routes, w-tradeSideW-2, h, base, sel-base, scroll)
	if side.small {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}
	lay.cards = append(lay.cards, side.cards...)
	if side.viewH > 0 {
		lay.total, lay.viewH = side.total, side.viewH
	}

	g.segsFit(2, h-3, w-4, said...)
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, deckKeys(pg, sel, side.scrolls), pickedHint(pg, sel))
	return g, lay
}

// pickedHint is the command the picked card types, for the key row.
func pickedHint(pg *deckPage, sel int) string {
	if c, ok := pg.card(sel); ok {
		return c.Command
	}
	return pg.Hint
}
