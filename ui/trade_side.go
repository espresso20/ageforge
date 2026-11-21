package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// trade_side.go is the board's right column: the goods table, and under it
// the offers.

// tradeSide is what the side column leaves the panel: the cells of each
// offer card in the list's own rows, and the list's size.
type tradeSide struct {
	cards        []image.Rectangle
	total, viewH int
	scrolls      bool
	small        bool // the list has no room: the panel draws the deck
}

// tradeGoodsRows is the goods the routes touch, in the Resources box order,
// each with what the running routes take and bring in a minute.
func tradeGoodsRows(st game.GameState, routes []tradeRoute) (goods []tradeFlow, takes, brings map[string]float64) {
	takes, brings = map[string]float64{}, map[string]float64{}
	used := map[string]bool{}
	for _, r := range routes {
		for k, fl := range [][]tradeFlow{r.out, r.in} {
			for _, f := range fl {
				used[f.key] = true
				if r.state == "running" && k == 0 {
					takes[f.key] += f.min
				} else if r.state == "running" {
					brings[f.key] += f.min
				}
			}
		}
	}
	tpm := 60 / tickInterval(st).Seconds()
	var order []string
	for _, rd := range st.Ruleset().Resources() {
		order = append(order, rd.Key)
	}
	for _, res := range append(order, config.ResearchTarget) {
		if used[res] {
			goods = append(goods, tradeFlow{key: res, good: tradeGoodName(res), stock: st.Resources[res].Amount, make_: tradeTownMake(st, res, tpm)})
		}
	}
	return goods, takes, brings
}

// drawTradeRule draws a heading and the rule that runs on from it to x1.
func drawTradeRule(g *xgrid, x, y, x1 int, title string) int {
	x = g.put(x, y, title, xBright) + 1
	g.put(x, y, strings.Repeat("─", max(0, x1-x)), xDim)
	return x
}

// drawTradeSide draws the column from sx: the goods table from row 3, and
// what stands under it. pick is the offer card picked (negative when a route
// is); scroll is the list's first row; base is the deck index of the first
// offer card.
func drawTradeSide(g *xgrid, st game.GameState, routes []tradeRoute, sx, h, base, pick, scroll int) tradeSide {
	var side tradeSide
	sx++
	end := sx + tradeSideW
	g.put(sx-1, 4, strings.Repeat("─", tradeSideW+1), xDim)
	for _, hd := range []struct {
		x int // a negative x is the column the text ends on
		s string
	}{{sx, rowid.Line("ui.trade_side.drawTradeSide.01", "GOOD")}, {-(sx + 27), rowid.Line("ui.trade_side.drawTradeSide.02", "IN STORE")},
		{-(sx + 39), rowid.Line("ui.trade_side.drawTradeSide.03", "MAKES/MIN")}, {-(sx + 51), rowid.Line("ui.trade_side.drawTradeSide.04", "TAKES/MIN")},
		{-(sx + 64), rowid.Line("ui.trade_side.drawTradeSide.05", "BRINGS/MIN")}, {-(sx + 86), rowid.Line("ui.trade_side.drawTradeSide.06", "TRADE NET")}} {
		if hd.x < 0 {
			g.rput(-hd.x, 3, hd.s, xAccent)
		} else {
			g.put(hd.x, 3, hd.s, xAccent)
		}
	}
	goods, takes, brings := tradeGoodsRows(st, routes)
	y := 5
	for _, f := range goods {
		g.putFit(sx, y, 18, f.good, xText)
		if f.key == config.ResearchTarget {
			g.rput(sx+27, y, "·", xDim)
			g.rput(sx+39, y, "·", xDim)
		} else {
			g.rput(sx+27, y, textfmt.Number(f.stock), xText)
			g.rput(sx+39, y, textfmt.Number(f.make_), xText)
		}
		for i, v := range []float64{takes[f.key], brings[f.key]} {
			if v > 0 {
				g.rput(sx+51+13*i, y, textfmt.Number(v), xText)
			} else {
				g.rput(sx+51+13*i, y, "·", xDim)
			}
		}
		s, k := tradeNetCell(f.make_, takes[f.key], brings[f.key])
		if f.make_ == 0 && brings[f.key] > 0 && f.stock > 0 {
			s, k = rowid.Line("ui.trade_side.drawTradeSide.07", "only source"), xAccent
		}
		g.rput(end-3, y, s, k)
		y++
	}
	return drawTradeOffers(g, st, routes, side, sx, y+1, h, base, pick, scroll)
}

// tradeCiv is a civilization met, with the offers it has open.
type tradeCiv struct {
	key, name, status, why, note string
	opinion, refresh             int
	war, embargo                 bool
	suspends                     []string
	offers                       []game.DealInfo
}

// tradeCivs lists the civilizations met, in the ruleset's order. Taken offers
// are left out.
func tradeCivs(st game.GameState, routes []tradeRoute) (civs []tradeCiv) {
	for _, fd := range st.Ruleset().Factions() {
		f := st.Diplomacy.Factions[fd.Key]
		if !f.Discovered {
			continue
		}
		c := tradeCiv{key: fd.Key, name: f.Name, status: f.Status, opinion: f.Opinion, refresh: f.DealRefreshIn, war: f.AtWar, embargo: f.Status == "embargo"}
		if f.AtWar {
			c.status = rowid.Line("ui.trade_side.tradeCivs.01", "war")
		}
		for _, d := range f.Deals {
			if !d.Taken {
				c.offers = append(c.offers, d)
			}
		}
		switch {
		case f.DealsBlocked != "":
			c.why = rowid.Sprintf("ui.trade_side.tradeCivs.02", "no offers: %s", f.DealsBlocked)
		case len(c.offers) == 0:
			c.why = rowid.Line("ui.trade_side.tradeCivs.03", "nothing to offer right now")
		}
		if c.note = c.why; c.note == "" {
			c.note = rowid.Sprintf("ui.trade_side.tradeCivs.04", "new offers in %s", formatTicks(f.DealRefreshIn, st))
		}
		for _, r := range routes {
			if r.state == "suspended" && r.by == f.Name {
				c.suspends = append(c.suspends, r.name)
			}
		}
		civs = append(civs, c)
	}
	return civs
}

// tradeActiveGifts is the Active gifts block: each running boost with its
// time left, then each earned know-how.
func tradeActiveGifts(st game.GameState) []deckLine {
	ls := []deckLine{{Value: rowid.Line("ui.trade_side.tradeActiveGifts.01", "Active gifts"), K: xAccent}}
	for _, b := range st.Diplomacy.Boosts {
		ls = append(ls, deckLine{Value: game.DealGift(game.DealBoost, b.Res, float64(b.TicksLeft), textfmt.Number, func(t int) string { return formatTicks(t, st) }), K: xGood})
	}
	for _, fd := range st.Ruleset().Factions() {
		if f := st.Diplomacy.Factions[fd.Key]; f.Knowhow {
			ls = append(ls, deckLine{Value: rowid.Sprintf("ui.trade_side.tradeActiveGifts.02", "%s (%s)", game.DealGift(game.DealKnowhow, f.Specialty, 0, textfmt.Number, nil), f.Name), K: xGood})
		}
	}
	if len(ls) == 1 {
		return nil
	}
	return ls
}

// tradePayable is whether the town holds what the offer asks for.
func tradePayable(st game.GameState, d game.DealInfo) bool {
	return st.Resources[d.Give].Amount >= d.GiveAmt
}

// tradeOfferCounts is the offers open and how many of them can be paid now.
func tradeOfferCounts(st game.GameState, civs []tradeCiv) (open, pay int) {
	for _, c := range civs {
		for _, d := range c.offers {
			open++
			if tradePayable(st, d) {
				pay++
			}
		}
	}
	return open, pay
}

// tradeOfferGive and tradeOfferGet are the two sides of an offer as text.
func tradeOfferGive(d game.DealInfo) string {
	return textfmt.Number(d.GiveAmt) + " " + game.ResourceName(d.Give)
}

func tradeOfferGet(st game.GameState, d game.DealInfo) string {
	return game.DealGift(d.Kind, d.Res, d.Amt, textfmt.Number, func(t int) string { return formatTicks(t, st) })
}

// tradeOfferCmd is the command that accepts an offer.
func tradeOfferCmd(civ string, d game.DealInfo) string {
	return fmt.Sprintf("diplomacy accept %s %d", civ, d.Num)
}

// tradeRow is one row of the offers list: a civilization ('c'), an offer
// ('o'), a line under the picked offer ('x'), or a line at the foot ('f').
type tradeRow struct {
	kind     byte
	civ, off int
	segs     []xseg
}

// tradePickedLines is what the picked offer says under itself.
func tradePickedLines(st game.GameState, civ tradeCiv, d game.DealInfo) (rows [][]xseg) {
	tickSec := tickInterval(st).Seconds()
	make_ := tradeTownMake(st, d.Give, 60/tickSec)
	store := textfmt.Number(st.Resources[d.Give].Amount)
	line := []xseg{{rowid.Sprintf("ui.trade_side.tradePickedLines.01", "in store %s %s · ", store, game.ResourceName(d.Give)), xText}}
	if make_ <= 0 {
		line = append(line, xseg{rowid.Line("ui.trade_side.tradePickedLines.02", "the town makes none"), xBad})
	} else {
		line = append(line, xseg{rowid.Sprintf("ui.trade_side.tradePickedLines.03", "the town makes %s a minute", textfmt.Number(make_)), xText})
		line = append(line, xseg{rowid.Sprintf("ui.trade_side.tradePickedLines.04", ": this lot is %s of it", formatTicks(int(d.GiveAmt/make_*60/tickSec+0.5), st)), xText})
	}
	return append([][]xseg{line}, []xseg{{rowid.Line("ui.trade_side.tradePickedLines.05", "type "), xDim}, {tradeOfferCmd(civ.key, d), xRoute}})
}

// drawTradeOffers draws the offers under the goods table from row y: a
// heading, the column heads, and a list that scrolls when it is taller than
// the room down to the last content row (h-4).
func drawTradeOffers(g *xgrid, st game.GameState, routes []tradeRoute, side tradeSide, sx, y, h, base, pick, scroll int) tradeSide {
	end := sx + tradeSideW
	civs := tradeCivs(st, routes)
	open, pay := tradeOfferCounts(st, civs)
	top := y + 3
	H := h - 3 - top
	if H < 5 {
		side.small = true
		return side
	}
	x := g.put(sx, y, rowid.Line("ui.trade_side.drawTradeOffers.01", "Offers"), xAccent) + 2
	x = g.put(x, y, rowid.Sprintf("ui.trade_side.drawTradeOffers.02", "%d open · %d you can pay now", open, pay), xDim) + 1
	word := rowid.Line("ui.trade_side.drawTradeOffers.03", "open Factions")
	px := end - 2 - runeLen(word) - 2
	g.put(x, y, strings.Repeat("─", max(0, px-x-1)), xDim)
	g.chip(px, y, chipSpec{kind: keySteel, label: word, cmd: "factions"})
	giveX, getX := sx+2, sx+36
	for _, hd := range []struct {
		x int
		s string
	}{{giveX, rowid.Line("ui.trade_side.drawTradeOffers.05", "YOU GIVE")}, {getX, rowid.Line("ui.trade_side.drawTradeOffers.06", "YOU GET")}} {
		g.put(hd.x, y+1, hd.s, xAccent)
	}
	g.rput(sx+68, y+1, rowid.Line("ui.trade_side.drawTradeOffers.07", "IN STORE"), xAccent)

	var rows []tradeRow
	oi := 0
	for ci, c := range civs {
		rows = append(rows, tradeRow{kind: 'c', civ: ci, off: -1})
		for k, d := range c.offers {
			rows = append(rows, tradeRow{kind: 'o', civ: ci, off: k})
			if oi == pick {
				for _, l := range tradePickedLines(st, c, d) {
					rows = append(rows, tradeRow{kind: 'x', civ: ci, off: k, segs: l})
				}
			}
			oi++
		}
	}
	foot := tradeActiveGifts(st)
	foot = append(foot, allyTradeNotes(st.Diplomacy)...)
	foot = append(foot, deckLine{Value: rowid.Line("ui.trade_side.drawTradeOffers.08", "each route that runs: +1 opinion a minute with every civilization not at war"), K: xDim})
	rows = append(rows, tradeRow{kind: 'f'})
	for _, l := range foot {
		rows = append(rows, tradeRow{kind: 'f', segs: []xseg{{l.Value, l.K}}})
	}
	N := len(rows)
	side.scrolls = N > H
	side.total, side.viewH = N, H
	lo, hi := 0, N
	if side.scrolls {
		scroll = max(0, min(scroll, N-H))
		lo, hi = scroll, scroll+H
		if lo > 0 {
			lo++
		}
		if hi < N {
			hi--
			if rows[hi-1].kind == 'c' && len(civs[rows[hi-1].civ].offers) > 0 {
				hi--
			}
		}
	}
	// A list that starts on an offer starts on its civilization instead.
	if lo > 0 && lo < N && (rows[lo].kind == 'o' || rows[lo].kind == 'x') {
		rows[lo] = tradeRow{kind: 'c', civ: rows[lo].civ, off: -1}
	}
	count := func(rs []tradeRow) (n, p int) {
		for _, r := range rs {
			if r.kind == 'o' {
				n++
				if tradePayable(st, civs[r.civ].offers[r.off]) {
					p++
				}
			}
		}
		return n, p
	}
	cy := top
	if lo > 0 {
		cy++
	}
	oi = 0
	for i, r := range rows {
		if r.kind == 'o' {
			side.cards = append(side.cards, image.Rect(0, i-2, 1, i+5))
		}
		if i < lo || i >= hi {
			continue
		}
		y := cy + i - lo
		switch r.kind {
		case 'c':
			tradeDrawCiv(g, civs[r.civ], sx, y, end)
		case 'o':
			d := civs[r.civ].offers[r.off]
			ok := tradePayable(st, d)
			tradeDrawOffer(g, st, civs[r.civ], d, ok, sx, y, [2]int{giveX, getX}, pick == oi)
			g.hit(hitDeckCard, base+oi, sx, y, tradeSideW, 1)
		case 'x':
			g.segsFit(giveX, y, end-giveX-1, r.segs...)
		case 'f':
			g.segsFit(sx, y, tradeSideW-1, r.segs...)
		}
		if r.kind == 'o' {
			oi++
		}
	}
	if lo > 0 {
		n, p := count(rows[:lo])
		tradeDrawMore(g, sx, top, end, n, p, true)
	}
	if hi < N {
		n, p := count(rows[hi:])
		tradeDrawMore(g, sx, top+H-1, end, n, p, false)
	}
	return side
}

// tradeDrawMore is the indicator row at an end of a scrolled list.
func tradeDrawMore(g *xgrid, sx, y, end, n, pay int, above bool) {
	where, key := rowid.Line("ui.trade_side.tradeDrawMore.01", "below"), "PgDn"
	if above {
		where, key = rowid.Line("ui.trade_side.tradeDrawMore.02", "above"), "PgUp"
	}
	s := rowid.Sprintf("ui.trade_side.tradeDrawMore.03", "more %s", where)
	if n > 0 {
		s = rowid.Sprintf("ui.trade_side.tradeDrawMore.04", "%d more %s %s", n, textfmt.Plural(n, rowid.Line("ui.trade_side.tradeDrawMore.05", "offer"), rowid.Line("ui.trade_side.tradeDrawMore.06", "offers")), where)
	}
	x := g.put(sx, y, s, xDim)
	if pay > 0 {
		g.put(x, y, rowid.Sprintf("ui.trade_side.tradeDrawMore.07", " · %d you can pay", pay), xAccent)
	}
	g.chip(end-2-runeLen(key)-2, y, chipSpec{kind: keySteel, label: key})
}

// tradeDrawCiv draws a civilization's line: its standing, what it will not do
// and the routes it suspends.
func tradeDrawCiv(g *xgrid, c tradeCiv, sx, y, end int) {
	g.putFit(sx, y, 24, c.name, xBright)
	k := xDim
	switch {
	case c.status == "allied":
		k = xGood
	case c.war || c.embargo:
		k = xBad
	}
	g.put(sx+27, y, c.status, k)
	x := g.put(sx+43, y, rowid.Sprintf("ui.trade_side.tradeDrawCiv.01", "opinion %d", c.opinion), xDim) + 2
	nx := sx + 87 - runeLen(c.note)
	g.put(nx, y, c.note, xDim)
	if len(c.suspends) > 0 && nx-x-1 >= 12 {
		g.putFit(x, y, nx-x-1, rowid.Sprintf("ui.trade_side.tradeDrawCiv.02", "suspends %s", strings.Join(c.suspends, ", ")), xBad)
	}
}

// tradeDrawOffer draws one offer's line; cols are the columns of its kind, its
// give and its get. The picked one is banded and shows the Enter key.
func tradeDrawOffer(g *xgrid, st game.GameState, c tradeCiv, d game.DealInfo, ok bool, sx, y int, cols [2]int, picked bool) {
	if picked {
		g.band(sx-1, y, tradeSideW+1, 1)
	}
	k, kind := xBad, keySteel
	if ok {
		k, kind = xText, keyAccent
	}
	x := g.put(cols[0], y, tradeOfferGive(d), k)
	g.put(max(x+1, cols[1]-3), y, "►", xDim)
	g.put(cols[1], y, tradeOfferGet(st, d), xText)
	g.rput(sx+68, y, textfmt.Number(st.Resources[d.Give].Amount), k)
	if picked {
		g.put(sx+71, y, rowid.Line("ui.trade_side.tradeDrawOffer.01", "Enter"), xDim)
	}
	g.chip(sx+79, y, chipSpec{kind: kind, label: rowid.Line("ui.trade_side.tradeDrawOffer.02", "accept"), cmd: tradeOfferCmd(c.key, d)})
}
