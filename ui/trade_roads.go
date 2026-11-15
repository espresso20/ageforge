package ui

import (
	"image"
	"sort"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// trade_roads.go draws the Trade panel as a map: the town at the left, a
// road from its gate to the end of each route, a caravan on every road that
// is running, and at the right what the routes move and what bears on them.
// It draws the same page tradePage makes for a deck of cards; a window too
// small for the map gets the cards.

const (
	// roadMinW and roadMinH are the smallest panel the map is drawn in.
	roadMinW, roadMinH = 118, 22
	// roadGateX is the column of the town's gate, where the roads start.
	roadGateX = 17
)

// roadLook is how a route's road and its end are drawn, read from the state
// chip of its card.
type roadLook struct {
	k          xk   // the road's colour
	line, vert rune // the road's strokes
	mark       rune // the mark at its end
	caravan    bool // a caravan is on the road
}

func roadLookOf(c *deckCard) roadLook {
	switch c.ChipK {
	case xChipShort:
		return roadLook{xGood, '━', '┃', '◆', true}
	case xChipMil:
		return roadLook{xBad, '━', '┃', '◆', true}
	case xChipCivic:
		return roadLook{xDim, '─', '│', '◆', false}
	}
	return roadLook{xFaint, '┄', '┆', '◇', false}
}

// renderTradeRoads draws the page as the map of roads. The cards that are
// routes (they have what goes out and what comes in) are the roads, in the
// page's order; the cards after them go in the box at the right.
func renderTradeRoads(pg *deckPage, st game.GameState, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout) {
	var cards []*deckCard
	for si := range pg.Sections {
		for ci := range pg.Sections[si].Cards {
			cards = append(cards, &pg.Sections[si].Cards[ci])
		}
	}
	routes := 0
	for routes < len(cards) && len(cards[routes].Lines) >= 2 {
		routes++
	}
	y0, y1 := 3, h-5
	avail := y1 - y0 + 1
	per := 3
	if routes*per > avail {
		per = 2
	}
	if routes == 0 || routes*per > avail || w < roadMinW || h < roadMinH {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, pg.Title, xTitle, false)
	g.closeButton(w)
	g.segsFit(2, 1, w-4, pg.Head...)
	lay := deckLayout{cards: make([]image.Rectangle, len(cards)), total: h - deckChrome, viewH: h - deckChrome}

	booksW := clamp(w/5, 28, 36)
	labelW := clamp(w*36/100, 42, 56)
	booksX := w - 2 - booksW
	labelX := booksX - 1 - labelW

	// The rows of the roads' ends: spread down the map, kept together.
	gap := min((avail-routes*per)/(routes+1), 3)
	start := y0 + (avail-routes*per-(routes-1)*gap)/2
	rows := make([]int, routes)
	for i := range rows {
		rows[i] = start + i*(per+gap)
	}
	ty := clamp((rows[0]+rows[routes-1])/2, y0+4, y1-2)

	// The land: a dusting of ground, so the roads cross somewhere.
	for y := y0; y <= y1; y++ {
		for x := 2; x < labelX-2; x++ {
			switch v := (x*7919 + y*104729 + x*y*31) % 97; v {
			case 0:
				g.set(x, y, '∩', xFaint)
			case 1, 2:
				g.set(x, y, '"', xFaint)
			case 3:
				g.set(x, y, '.', xFaint)
			}
		}
	}
	drawRoadTown(g, ty, mapmodel.TownNameOf(st.Seed))

	// Where each road leaves the trunk: the road that ends farthest from the
	// gate's row turns off first, so no road crosses another.
	bend := make([]int, routes)
	var above, below []int
	for i, dy := range rows {
		switch {
		case dy < ty:
			above = append(above, i)
		case dy > ty:
			below = append(below, i)
		}
	}
	sort.Slice(below, func(a, b int) bool { return rows[below[a]] > rows[below[b]] })
	for n, i := range above { // already farthest first
		bend[i] = roadGateX + 2 + 2*n
	}
	for n, i := range below {
		bend[i] = roadGateX + 2 + 2*n
	}
	lastBend := roadGateX + 2 + 2*max(len(above), len(below))
	for x := roadGateX; x <= lastBend; x++ {
		g.set(x, ty, '─', xDim)
	}

	for i := 0; i < routes; i++ {
		c, dy := cards[i], rows[i]
		look := roadLookOf(c)
		k := look.k
		if i == sel {
			k = xAccent
		}
		from := lastBend
		switch {
		case dy < ty:
			from = bend[i]
			for y := dy + 1; y < ty; y++ {
				g.set(from, y, look.vert, k)
			}
			g.set(from, ty, '┴', xDim)
			g.set(from, dy, '╭', k)
		case dy > ty:
			from = bend[i]
			for y := ty + 1; y < dy; y++ {
				g.set(from, y, look.vert, k)
			}
			g.set(from, ty, '┬', xDim)
			g.set(from, dy, '╰', k)
		}
		for x := from + 1; x < labelX-1; x++ {
			g.set(x, dy, look.line, k)
		}
		// The caravan: out for the first half of a trip with what it sells,
		// home for the second with what it bought. One that is held waits
		// at the start of its road.
		if look.caravan {
			span := labelX - 4 - (from + 2)
			p, load := c.Share*2, c.Lines[0].Value
			if p > 1 {
				p, load = 2-p, c.Lines[1].Value
			}
			cx := from + 2 + int(p*float64(span))
			g.put(cx, dy, "◆◆", xBright)
			if lx := clamp(cx-runeLen(load)/2, from+2, labelX-3-runeLen(load)); dy-1 >= y0 && lx > from {
				g.put(lx, dy-1, " "+load+" ", xLabel)
			}
		}
		// The end of the road: the route's name and state, what it trades,
		// and under that its button when it is picked, else what it waits on.
		nk := xBright
		if c.Quiet {
			nk = xDim
		}
		g.set(labelX, dy, look.mark, k)
		if i == sel {
			g.put(labelX-2, dy, "▸", xAccent)
		}
		x := g.putFit(labelX+2, dy, labelW-6-runeLen(c.Chip), c.Title, nk)
		if c.Chip != "" {
			g.put(x+1, dy, " "+c.Chip+" ", c.ChipK)
		}
		goods := []xseg{{c.Lines[0].Value, xText}, {" → ", xDim}, {c.Lines[1].Value, max(c.Lines[1].K, xText)}}
		under, uk := c.MeterText, xDim
		if c.Why != "" {
			under, uk = c.Why, xBad
		}
		last := dy + per - 1
		if i == sel && c.Command != "" {
			ex := g.keyButton(labelX+2, last, "Enter", c.Action)
			g.hit(hitDeckAct, i, labelX+2, last, ex-labelX-2, 1)
			g.putFit(ex+1, last, labelX+labelW-ex-1, c.Command, xDim)
			if per == 3 {
				g.segsFit(labelX+2, dy+1, labelW-2, goods...)
			}
		} else {
			g.segsFit(labelX+2, dy+1, labelW-2, goods...)
			if per == 3 {
				g.putFit(labelX+2, last, labelW-2, under, uk)
			}
		}
		lay.cards[i] = image.Rect(labelX-2, dy, labelX+labelW, dy+per)
		g.hit(hitDeckCard, i, labelX-2, dy, labelW+2, per)
	}

	if !drawRoadBooks(g, pg, cards, routes, sel, booksX, y0, booksW, avail, &lay) {
		return renderDeck(pg, sel, scroll, w, h, said...) // the box cannot hold all it has to say
	}
	g.segsFit(2, h-3, w-4, said...)
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, deckKeys(pg, sel, false), pg.Hint)
	return g, lay
}

// drawRoadTown draws the town at the left of the map, its gate on row ty:
// a wall, a keep and roofs, and the town's name under it.
func drawRoadTown(g *xgrid, ty int, name string) {
	town := [...]string{
		"     ▲       ",
		"  ∩  █  ∩ ∩  ",
		" ∩∩∩ █ ∩∩∩∩∩ ",
		"▐███████████▌",
		"▐█ ∩∩ ∩ ∩∩ █▌",
		" ▀▀▀▀▀▀▀▀▀▀▀ ",
	}
	for i, l := range town {
		for j, r := range []rune(l) {
			k := xHill
			switch r {
			case ' ':
				continue
			case '∩':
				k = xGold
			case '▀':
				k = xDim
			case '▲':
				k = xAccent
			}
			if j == 5 && i > 0 && i < 3 {
				k = xTitle // the keep
			}
			g.set(3+j, ty-4+i, r, k)
		}
	}
	name = fitText(name, 13)
	g.put(3+(13-runeLen(name))/2-1, ty+2, " "+name+" ", xLabel)
}

// drawRoadBooks draws the box at the right of the map: what the running
// routes take out and bring in, the page's notes, and the cards that are
// not routes, each with its button. It reports whether all of it fitted.
func drawRoadBooks(g *xgrid, pg *deckPage, cards []*deckCard, routes, sel, x, y, w, h int, lay *deckLayout) bool {
	g.box(x, y, w, h, xFaint, "", xTitle, false)
	inner, row, end := w-4, y+1, y+h-1
	put := func(k xk, s string) {
		if row < end {
			g.putFit(x+2, row, inner, s, k)
		}
		row++
	}
	// What goes out and what comes in, a line a running route.
	for side := 0; side < 2; side++ {
		head := false
		for _, c := range cards[:routes] {
			if !roadLookOf(c).caravan {
				continue
			}
			if !head {
				put(xDim, c.Lines[side].Label)
				head = true
			}
			if row < end {
				k := max(c.Lines[side].K, xText)
				nx := g.putFit(x+3, row, inner-1, c.Lines[side].Value, k)
				g.putFit(nx+2, row, x+2+inner-nx-2, c.Title, xDim)
			}
			row++
		}
		if head {
			row++
		}
	}
	for _, n := range pg.Notes {
		for _, l := range wrapCells(n.Label+n.Value, inner) {
			put(max(n.K, xDim), l)
		}
	}
	if len(pg.Notes) > 0 {
		row++
	}
	for i := routes; i < len(cards); i++ {
		c := cards[i]
		top := row
		tk := xBright
		if i == sel {
			tk = xAccent
		}
		put(tk, c.Title)
		for _, l := range wrapCells(c.Desc, inner) {
			put(xDim, l)
		}
		if c.Command != "" {
			if row < end {
				bx := x + 2
				if i == sel {
					bx = g.keyButton(x+2, row, "Enter", c.Action)
				} else {
					bx = g.put(x+2, row, " "+c.Action+" ", xKeyName)
				}
				g.hit(hitDeckAct, i, x+2, row, bx-x-2, 1)
				g.putFit(bx+1, row, x+2+inner-bx-1, c.Command, xDim)
			}
			row++
		}
		if row > end {
			return false
		}
		lay.cards[i] = image.Rect(x+1, top, x+w-1, row)
		g.hit(hitDeckCard, i, x+1, top, w-2, row-top)
		row++
	}
	return row <= end+1
}
