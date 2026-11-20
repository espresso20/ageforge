package ui

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// trade_stop.go is the smugglers' stop on the Trade page: a card for the
// black market with what a run costs, the odds, whether one can be made now
// and a chip for each resource a win would pay in. The map draws it as a
// place of its own, off the roads.

// stopRows is the rows the stop takes on the map: its four rows of drawing
// and its name.
const stopRows = 5

// blackMarketSection is the page's section for the stop; it has none before
// the black market opens.
func blackMarketSection(state game.GameState) (deckSection, bool) {
	bm := state.BlackMarket
	if !bm.Available {
		return deckSection{}, false
	}
	c := deckCard{
		Title: rowid.Line("ui.trade_stop.blackMarketSection.02", "Black market"), Stop: true,
		Lines: []deckLine{
			{rowid.Line("ui.trade_stop.blackMarketSection.03", "Cost"), game.Amount(bm.Cost, "culture"), 0},
			{textfmt.Capitalize(game.ResourceName("culture")), textfmt.Number(bm.Culture), xGood},
			{rowid.Line("ui.trade_stop.blackMarketSection.04", "Odds"), rowid.Line("ui.trade_stop.blackMarketSection.05", "five in six"), 0},
		},
	}
	sec := deckSection{Title: rowid.Line("ui.trade_stop.blackMarketSection.01", "Smugglers"), MinW: 46}
	if refusal, locked := featureRefusal(state, config.FeatureBlackMarket); locked {
		c.Quiet, c.Why, c.Lines = true, refusal, nil
		sec.Cards = []deckCard{c}
		return sec, true
	}
	short := bm.Culture < bm.Cost
	switch {
	case !bm.Ready:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.06", "lying low"), xChipMil
		c.Meter = true
		if bm.CooldownTicks > 0 {
			c.Share = max(0, min(1, 1-float64(bm.CooldownLeft)/float64(bm.CooldownTicks)))
		}
		c.MeterText = rowid.Sprintf("ui.trade_stop.blackMarketSection.07", "%s left", formatTicks(bm.CooldownLeft, state))
	case short:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.08", "not enough culture"), xChipMil
		c.Lines[1].K = xBad
	default:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.09", "ready"), xChipCivic
	}
	for _, p := range bm.Payouts {
		c.Chips = append(c.Chips, deckChip{Word: game.Amount(p.Amount, p.Resource), Command: "blackmarket " + p.Resource, Off: !bm.Ready || short})
	}
	sec.Cards = []deckCard{c}
	return sec, true
}

// stopTop is the top row of the stop on a map whose land runs from row y0 to
// y1 and whose town's gate is on row ty: under the town when the land has the
// rows, else over it. ok is false when it has neither.
func stopTop(y0, y1, ty int) (top int, ok bool) {
	if top = ty + 4; top+stopRows-1 <= y1 {
		return top, true
	}
	if top = y0; top+stopRows <= ty-4 {
		return top, true
	}
	return 0, false
}

// drawRoadStop draws the stop on the map with its top row at top: a shack
// with a lit window and a few crates, and its name under it, where no road
// runs. It takes a hit for the card n.
func drawRoadStop(g *xgrid, top int, c *deckCard, picked bool, n int) {
	shack := [...]string{
		"  ▄▄▄▄▄▄▄▄   ",
		" ▐████████▌▪ ",
		" ▐█•▐▌████▌▪▪",
		"  ▀▀▀▀▀▀▀▀   ",
	}
	for i, l := range shack {
		for j, r := range []rune(l) {
			k := xHill
			switch r {
			case ' ':
				continue
			case '▄', '▀':
				k = xDim
			case '•':
				k = xGold
			case '▪':
				k = xFaint
			case '▐', '▌':
				if i == 2 && j >= 4 && j <= 5 {
					k = xBad // the door
				}
			}
			switch {
			case picked && (r == '▄' || r == '▀'):
				k = xAccent
			case c.Quiet && r != '▪':
				k = xFaint
			}
			g.set(3+j, top+i, r, k)
		}
	}
	name := fitText(c.Title, 13)
	nk := xLabel
	if c.Quiet {
		nk = xDim
	}
	nx := 3 + (13-runeLen(name))/2 - 1
	g.put(nx, top+4, " "+name+" ", nk)
	if picked {
		g.put(nx-1, top+4, "▸", xAccent)
	}
	g.hit(hitDeckCard, n, 2, top, 15, stopRows)
}

// drawStopEntry draws the stop's entry in the box at the right of the map,
// its text starting at column x on row, and returns the next row. Every stop
// shows its name and state; the picked one adds what a run costs, the odds,
// the wait, and its chips. The caller checks that the rows fitted.
func drawStopEntry(g *xgrid, c *deckCard, picked bool, x, row, inner int) int {
	tk := xBright
	switch {
	case picked:
		tk = xAccent
	case c.Quiet:
		tk = xDim
	}
	g.putFit(x+2, row, inner, c.Title, tk)
	row++
	if c.Chip != "" {
		g.put(x+2, row, " "+c.Chip+" ", c.ChipK)
		row++
	}
	if !picked {
		return row
	}
	lw := 0
	for _, l := range c.Lines {
		lw = max(lw, runeLen(l.Label))
	}
	for _, l := range c.Lines {
		k := l.K
		if k == 0 {
			k = xText
		}
		g.put(x+2, row, l.Label, xDim)
		g.putFit(x+3+lw, row, inner-lw-1, l.Value, k)
		row++
	}
	if c.Meter {
		mw := clamp(inner-1-runeLen(c.MeterText), 4, 14)
		advBar(g, x+2, row, mw, c.Share, xAccent)
		g.putFit(x+3+mw, row, inner-mw-1, c.MeterText, xText)
		row++
	}
	for _, l := range wrapCells(c.Why, inner) {
		g.put(x+2, row, l, xBad)
		row++
	}
	return row + g.deckChips(x+2, row, inner, c.Chips)
}
