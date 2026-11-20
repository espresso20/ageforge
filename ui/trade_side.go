package ui

import (
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
// is); scroll is the list's first row.
func drawTradeSide(g *xgrid, st game.GameState, routes []tradeRoute, sx, h, pick, scroll int) tradeSide {
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
	return side
}
