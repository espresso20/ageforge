package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
)

// army_field.go draws the Army panel as a field: the defense set against the
// threat on one bar, the soldiers mustered post by post, and the campaigns
// as orders. It draws the same page armyPage makes for a deck of cards; a
// window too small for the field gets the cards.

const (
	// armyMinW and armyMinH are the smallest panel the field is drawn in.
	armyMinW, armyMinH = 100, 24
	// armyOverview is how many cards open the page: soldiers, posts,
	// defense and threat. They are drawn as the bar and the muster, and
	// are not things to pick.
	armyOverview = 4
	// musterMarks is the most marks a post's row of soldiers has.
	musterMarks = 60
)

// musterLegend says how many soldiers a mark of the muster stands for.
func musterLegend(unit int) string {
	return rowid.Sprintf("ui.army_field.musterLegend.01", "each ▲ is %d", unit)
}

// musterUnit is how many soldiers a mark stands for, so the largest post
// fits its row: 1, 2, 5, 10, 20, 50 and so on.
func musterUnit(most int) int {
	unit := 1
	for i := 0; most > unit*musterMarks; i++ {
		switch i % 3 {
		case 0:
			unit *= 2
		case 1:
			unit = unit / 2 * 5
		default:
			unit *= 2
		}
	}
	return unit
}

// renderArmyField draws the page as the field.
func renderArmyField(pg *deckPage, st game.GameState, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout) {
	if w < armyMinW || h < armyMinH || len(pg.Sections) == 0 || len(pg.Sections[0].Cards) != armyOverview {
		return renderDeck(pg, sel, scroll, w, h, said...)
	}
	mil := st.Military
	over := pg.Sections[0].Cards
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, pg.Title, xTitle, false)
	g.closeButton(w)
	inner := w - 4
	lay := deckLayout{cards: make([]image.Rectangle, pg.count()), total: h - deckChrome, viewH: h - deckChrome}

	// The balance: the defense from the left, the threat from the right, and
	// the line between them where the two meet.
	y := 2
	def, thr := over[2], over[3]
	side := max(runeLen(def.Title)+1+runeLen(def.Lines[0].Value), runeLen(thr.Title)+1+runeLen(thr.Lines[0].Value), 14)
	barX, barW := 2+side+2, inner-2*(side+2)
	g.segs(2, y, xseg{def.Title + " ", xGood}, xseg{def.Lines[0].Value, xBright})
	split := -1
	if sum := mil.DefenseRating + mil.Threat; sum > 0 {
		split = int(mil.DefenseRating/sum*float64(barW) + 0.5)
	}
	for i := 0; i < barW; i++ {
		switch {
		case split < 0:
			g.set(barX+i, y, '░', xFaint)
		case i < split:
			g.set(barX+i, y, '█', xGood)
		default:
			g.set(barX+i, y, '█', xBad)
		}
	}
	if split > 0 && split < barW {
		g.set(barX+split, y, '▌', xBright)
	}
	g.segs(barX+barW+2, y, xseg{thr.Title + " ", xBad}, xseg{thr.Lines[0].Value, xBright})
	if len(thr.Lines) > 1 {
		g.rput(w-3, y+1, thr.Lines[1].Value, xDim)
	}
	y += 2
	for _, n := range pg.Notes {
		for _, l := range wrapCells(n.Label+n.Value, inner) {
			g.put(2, y, l, max(n.K, xDim))
			y++
		}
	}
	y++

	// The muster: who stands at each post, a mark for so many soldiers.
	sol, posts := over[0], over[1]
	x := g.put(2, y, sol.Title, xTitle)
	for _, l := range sol.Lines {
		x = g.put(x+2, y, l.Label+l.Value, max(l.K, xText))
	}
	most := 0
	for _, p := range mil.Posts {
		most = max(most, p.Posts)
	}
	unit := musterUnit(most)
	if most > 0 {
		g.rput(w-3, y, musterLegend(unit), xDim)
	}
	y += 2
	if len(mil.Posts) == 0 {
		g.putFit(2, y, inner, posts.Desc, xDim)
		y++
	}
	lw := 0
	for _, l := range posts.Lines {
		lw = max(lw, runeLen(l.Label)+runeLen(l.Value)+1)
	}
	for i, p := range mil.Posts {
		if i >= len(posts.Lines) {
			break
		}
		l := posts.Lines[i]
		mx := g.segs(2, y, xseg{l.Label, xText}, xseg{l.Value, xBright})
		mx = max(mx, 2+lw) + 2
		for j := 0; j < (p.Posts+unit-1)/unit && mx < w-3; j++ {
			k := xFaint
			if j < (p.Posted+unit-1)/unit {
				k = xArmy
			}
			g.set(mx, y, '▲', k)
			mx++
			if j%10 == 9 {
				mx++
			}
		}
		y++
	}
	y++

	// The orders: the campaign that is out, then the ones that could go.
	n := armyOverview
	for _, s := range pg.Sections[1:] {
		x := g.put(2, y, s.Title+" ", xTitle)
		for ; x < w-2-runeLen(s.Meta)-1; x++ {
			g.set(x, y, '─', xFaint)
		}
		g.put(x+1, y, s.Meta, xDim)
		y++
		for _, l := range s.Notes {
			g.putFit(2, y, inner, l.Label+l.Value, max(l.K, xDim))
			y++
		}
		for ci := range s.Cards {
			y = drawArmyOrder(g, &s.Cards[ci], n, n == sel, y, w, &lay)
			n++
		}
		y++
	}
	if y > h-2 {
		return renderDeck(pg, sel, scroll, w, h, said...) // the orders do not fit the field
	}
	g.segsFit(2, h-3, inner, said...)
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, deckKeys(pg, sel, false), pg.Hint)
	return g, lay
}

// drawArmyOrder draws card n as an order from row y: its name, state and
// terms; what it is about, or its march when it is out; and its button when
// it is picked, else what stands in its way. It returns the next free row.
func drawArmyOrder(g *xgrid, c *deckCard, n int, picked bool, y, w int, lay *deckLayout) int {
	inner, top := w-4, y
	mk, nk := xDim, xBright
	switch {
	case picked:
		mk = xAccent
		g.put(1, y, "▸", xAccent)
	case c.Meter:
		mk = xLegend
	}
	if c.Quiet {
		nk = xDim
	}
	x := g.put(2, y, "⚔ ", mk)
	x = g.putFit(x, y, inner/3, c.Title, nk)
	if c.Chip != "" {
		x = g.put(x+1, y, " "+c.Chip+" ", c.ChipK)
	}
	x += 2
	for i, l := range c.Lines {
		if i > 0 {
			x = g.put(x, y, " · ", xFaint)
		}
		if l.Label != "" {
			x = g.putFit(x, y, w-2-x, l.Label+" ", xDim)
		}
		x = g.putFit(x, y, w-2-x, l.Value, max(l.K, xText))
	}
	y++
	if c.Meter {
		// The march: how far through the campaign the soldiers are.
		const march = 40
		done := int(c.Share*march + 0.5)
		for i := 0; i < march; i++ {
			r, k := '━', xLegend
			if i >= done {
				r, k = '·', xFaint
			}
			g.set(4+i, y, r, k)
		}
		g.put(4+clamp(done-1, 0, march-2), y, "◆◆", xBright)
		g.putFit(4+march+2, y, inner-march-4, c.MeterText, xText)
		y++
	} else if c.Desc != "" {
		g.putFit(4, y, inner-2, c.Desc, xDim)
		y++
	}
	switch {
	case picked && c.Command != "":
		y += g.deckAction(4, y, c, true, false)
	case c.Why != "":
		g.putFit(4, y, inner-2, c.Why, xBad)
		y++
	}
	lay.cards[n] = image.Rect(2, top, w-2, y)
	g.hit(hitDeckCard, n, 2, top, inner, y-top)
	return y + 1
}
