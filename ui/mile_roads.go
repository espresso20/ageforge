package ui

import (
	"image"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// mile_roads.go draws the Milestones panel as roads: a road a category, a
// stone a milestone. The stones reached come first, then the one being
// walked to, then the ones ahead; the picked stone is shown in full under
// the roads. It draws the same page milestonesPage makes for a deck of
// cards; a window too small for the roads gets the cards.

const (
	// mileMinW and mileMinH are the smallest panel the roads are drawn in.
	mileMinW, mileMinH = 100, 26
	// mileLabelW is the column at the left with a road's name and count;
	// mileStep the least cells from one stone to the next, mileStepMax the
	// most.
	mileLabelW  = 26
	mileStep    = 12
	mileStepMax = 22
	// mileDetailH is the rows of the box that shows the picked stone.
	mileDetailH = 7
)

// mileKinds are the roads' colours, in turn.
var mileKinds = [...]xk{xWork, xLore, xWater, xBad, xGold, xGood, xArmy, xLegend}

// mileDone reports whether a milestone's card is one that is reached.
func mileDone(c *deckCard) bool { return c.ChipK == xChipShort }

// mileOrder is a road's stones in the order they stand along it: reached,
// then nearest to reached, then the rest as the page has them.
func mileOrder(cards []deckCard) []int {
	order := make([]int, len(cards))
	for i := range order {
		order[i] = i
	}
	rank := func(c *deckCard) (int, float64) {
		switch {
		case mileDone(c):
			return 0, 0
		case c.Meter && c.Share > 0:
			return 1, -c.Share
		}
		return 2, 0
	}
	sort.SliceStable(order, func(a, b int) bool {
		ra, sa := rank(&cards[order[a]])
		rb, sb := rank(&cards[order[b]])
		return ra < rb || ra == rb && sa < sb
	})
	return order
}

// renderMileRoads draws the page as the roads.
func renderMileRoads(pg *deckPage, st game.GameState, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout) {
	if w < mileMinW || h < mileMinH || pg.count() == 0 {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, pg.Title, xTitle, false)
	g.closeButton(w)
	g.segsFit(2, 1, w-4, pg.Head...)
	inner := w - 4
	trackW := inner - mileLabelW
	perRow := max(trackW/mileStep, 1)
	viewH := h - deckChrome - 2 - mileDetailH

	// Where each road goes: two rows a line of stones, a blank row after.
	tops := make([]int, len(pg.Sections))
	total := 0
	for si, s := range pg.Sections {
		tops[si] = total
		total += 2*max((len(s.Cards)+perRow-1)/perRow, 1) + 1
	}
	total = max(total-1, 1)
	lay := deckLayout{cards: make([]image.Rectangle, pg.count()), total: total, viewH: viewH}
	scroll = clamp(scroll, 0, max(0, total-viewH))

	tall := newXGrid(inner, total)
	n0 := 0
	var picked *deckCard
	pickedRoad, pickedK := "", xText
	for si := range pg.Sections {
		s := &pg.Sections[si]
		k, y := mileKinds[si%len(mileKinds)], tops[si]
		x := tall.putFit(0, y, mileLabelW-3-runeLen(s.Meta), s.Title, k)
		tall.put(x+1, y, s.Meta, xDim)
		var notes []string
		for _, l := range s.Notes {
			notes = append(notes, l.Label+l.Value)
		}
		tall.putFit(0, y+1, mileLabelW-2, strings.Join(notes, " · "), xDim)
		order := mileOrder(s.Cards)
		step := clamp(trackW/max(min(len(order), perRow), 1), mileStep, mileStepMax)
		for pos, ci := range order {
			c := &s.Cards[ci]
			sx, sy := mileLabelW+(pos%perRow)*step, y+2*(pos/perRow)
			// The road to the next stone on this line: walked as far as
			// the next stone is reached.
			if pos+1 < len(order) && (pos+1)%perRow != 0 {
				next := &s.Cards[order[pos+1]]
				walked := 0
				switch {
				case mileDone(next):
					walked = step
				case mileDone(c) && next.Meter:
					walked = int(next.Share * float64(step-1))
				}
				for dx := 1; dx < step; dx++ {
					if dx <= walked {
						tall.set(sx+dx, sy, '━', k)
					} else {
						tall.set(sx+dx, sy, '─', xFaint)
					}
				}
			}
			mark, mk, nk := '○', xFaint, xDim
			switch {
			case mileDone(c):
				mark, mk, nk = '●', k, xText
			case c.ChipK == xChipCivic || c.Meter && c.Share > 0:
				mark, mk, nk = '◉', xBright, xBright
			}
			if n0+ci == sel {
				picked, pickedRoad, pickedK = c, s.Title, k
				nk = xAccent
				if sx > 0 {
					tall.put(sx-1, sy, "[", xAccent)
				}
				tall.put(sx+1, sy, "]", xAccent)
			}
			tall.set(sx, sy, mark, mk)
			tall.put(sx, sy+1, fitText(c.Title, step-1), nk)
			r := image.Rect(sx, sy, min(sx+step-1, inner), sy+2)
			lay.cards[n0+ci] = r
			tall.hit(hitDeckCard, n0+ci, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		}
		n0 += len(s.Cards)
	}

	// The roads in view, then the picked stone in full under them.
	const top = 3
	for yy := 0; yy < viewH && scroll+yy < total; yy++ {
		for xx := 0; xx < inner; xx++ {
			g.store(2+xx, top+yy, tall.at(xx, scroll+yy))
		}
	}
	view := image.Rect(2, top, 2+inner, top+viewH)
	for _, ht := range tall.hits {
		if r := ht.r.Add(image.Pt(2, top-scroll)).Intersect(view); !r.Empty() {
			g.hits = append(g.hits, xhit{ht.id, ht.n, r})
		}
	}
	if total > viewH {
		mark, size := top+scroll*viewH/total, max(1, viewH*viewH/total)
		for i := 0; i < size && mark+i < top+viewH; i++ {
			g.set(w-1, mark+i, '┃', xAccent)
		}
	}
	if picked != nil {
		drawMileStone(g, picked, pickedRoad, pickedK, 2, h-3-mileDetailH, inner)
	}
	g.segsFit(2, h-3, inner, said...)
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, deckKeys(pg, sel, total > viewH), pg.Hint)
	return g, lay
}

// drawMileStone draws the picked milestone in full in a box at (x, y): its
// name and state on the edge with its road's name, then what it asks, how
// far along it is, what holds it and what it gives.
func drawMileStone(g *xgrid, c *deckCard, road string, k xk, x, y, w int) {
	g.box(x, y, w, mileDetailH, xAccent, c.Title, xBright, false)
	rx := x + w - 2 - runeLen(road) - 2
	g.put(rx, y, " "+road+" ", k)
	if c.Chip != "" {
		g.put(rx-runeLen(c.Chip)-3, y, " "+c.Chip+" ", c.ChipK)
	}
	row, end := y+1, y+mileDetailH-1
	put := func(parts ...xseg) {
		if row < end {
			g.segsFit(x+2, row, w-4, parts...)
			row++
		}
	}
	for _, l := range wrapCells(c.Desc, w-4) {
		put(xseg{l, xText})
	}
	if c.Meter {
		bw := clamp(w/3, 20, 60)
		if row < end {
			fill := xAccent
			if c.Share >= 1 {
				fill = xGood
			}
			advBar(g, x+2, row, bw, c.Share, fill)
			g.putFit(x+4+bw, row, w-6-bw, c.MeterText, xBright)
			row++
		}
	}
	if c.Why != "" {
		put(xseg{c.Why, xBad})
	}
	for _, l := range c.Lines {
		label := l.Label
		if label != "" {
			label += " "
		}
		put(xseg{label, xDim}, xseg{l.Value, max(l.K, xText)})
	}
}
