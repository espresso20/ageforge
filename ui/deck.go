package ui

import (
	"image"

	"github.com/espresso20/ageforge/vibe"
)

// deck.go draws a deck: a panel of cards in sections. It is the look the
// list panels share (Trade, Army, Milestones): what a panel shows is a
// deckPage worked out from a snapshot, and this file lays it out and draws
// it. deck_panel.go is the panel that holds one.

// deckLine is a line inside a card or under a section's name: a quiet label
// and a value in K's colour (plain text when K is zero).
type deckLine struct {
	Label, Value string
	K            xk
}

// deckCard is one card.
type deckCard struct {
	Title string
	// Chip is the card's state, at the right of its top edge ("" for none).
	Chip  string
	ChipK xk
	// Desc is quiet text under the title, wrapped to two lines at most.
	Desc  string
	Lines []deckLine
	// Meter draws a bar filled to Share with MeterText after it.
	Meter     bool
	Share     float64
	MeterText string
	// Why says, in the bad colour, what stands in the card's way.
	Why string
	// Command is what Enter, or a click on the card's button, types and
	// runs; Action is the word on the button. A card with no command only
	// shows.
	Action, Command string
	// Quiet draws the title dim: a card that is done with, or out of reach.
	Quiet bool
}

// deckSection is a named run of cards.
type deckSection struct {
	Title string
	// Meta is a count or a note at the right end of the section's rule.
	Meta string
	// MinW is the least width a card of this section is given (0 for
	// deckCardW).
	MinW  int
	Notes []deckLine
	Cards []deckCard
}

// deckPage is everything a deck panel shows.
type deckPage struct {
	Title string
	// Head is a line of facts under the title; Notes are lines under that,
	// each wrapped and in its own colour (a warning in the bad one).
	Head     []xseg
	Notes    []deckLine
	Sections []deckSection
	// Hint is what the foot says may be typed.
	Hint string
}

const (
	// deckCardW is the least width of a card, deckCols the most columns.
	deckCardW = 38
	deckCols  = 4
	// deckChrome is the panel's rows that are not content: the top edge, a
	// blank row, the key line and the bottom edge.
	deckChrome = 4
)

const (
	hitDeckCard = iota // a card
	hitDeckAct         // a card's button
)

// deckLayout is where a drawn deck's parts are.
type deckLayout struct {
	// cards are each card's cells in the content, by its number; total is
	// the content's rows and viewH how many of them the panel shows.
	cards        []image.Rectangle
	total, viewH int
}

// card is the n-th card of the page, counting through its sections.
func (pg *deckPage) card(n int) (deckCard, bool) {
	for _, s := range pg.Sections {
		if n >= 0 && n < len(s.Cards) {
			return s.Cards[n], true
		}
		n -= len(s.Cards)
	}
	return deckCard{}, false
}

// noteLines is the page's notes wrapped to inner cells, each line with its
// colour (quiet when the note names none).
func (pg *deckPage) noteLines(inner int) []xseg {
	var out []xseg
	for _, n := range pg.Notes {
		k := n.K
		if k == 0 {
			k = xDim
		}
		for _, l := range deckWrap(n.Label+n.Value, inner) {
			out = append(out, xseg{l, k})
		}
	}
	return out
}

// count is how many cards the page has.
func (pg *deckPage) count() int {
	n := 0
	for _, s := range pg.Sections {
		n += len(s.Cards)
	}
	return n
}

// deckWrap wraps s to at most two lines of w cells, breaking at a space.
func deckWrap(s string, w int) []string {
	rs := []rune(s)
	if len(rs) <= w || w < 4 {
		return []string{fitText(s, max(w, 1))}
	}
	cut := w
	for cut > w/2 && rs[cut] != ' ' {
		cut--
	}
	rest := rs[cut:]
	for len(rest) > 0 && rest[0] == ' ' {
		rest = rest[1:]
	}
	return []string{string(rs[:cut]), fitText(string(rest), w)}
}

// height is the rows the card takes cw wide, its edges included.
func (c *deckCard) height(cw int) int {
	h := 2 + len(c.Lines)
	if c.Desc != "" {
		h += len(deckWrap(c.Desc, cw-4))
	}
	for _, on := range []bool{c.Meter, c.Why != "", c.Command != ""} {
		if on {
			h++
		}
	}
	return h
}

// deckColumns is how many cards go side by side in inner cells, and how wide
// each is.
func deckColumns(inner, minW int) (cols, cw int) {
	if minW <= 0 {
		minW = deckCardW
	}
	cols = clamp(inner/minW, 1, deckCols)
	return cols, (inner - (cols - 1)) / cols
}

// layoutDeck works out where each card goes in a content inner cells wide,
// and how many rows the content has.
func layoutDeck(pg *deckPage, inner int) (cards []image.Rectangle, total int) {
	y := 0
	if len(pg.Head) > 0 {
		y += 2
	}
	if notes := pg.noteLines(inner); len(notes) > 0 {
		y += len(notes) + 1
	}
	for _, s := range pg.Sections {
		y += 1 + len(s.Notes)
		cols, cw := deckColumns(inner, s.MinW)
		for i := 0; i < len(s.Cards); i += cols {
			row := s.Cards[i:min(i+cols, len(s.Cards))]
			ch := 0
			for j := range row {
				ch = max(ch, row[j].height(cw))
			}
			for j := range row {
				cards = append(cards, image.Rect(j*(cw+1), y, j*(cw+1)+cw, y+ch))
			}
			y += ch
		}
		y++
	}
	return cards, max(y-1, 0)
}

// renderDeck draws the page w by h with card sel picked and the content
// scrolled down by scroll rows (held in range). said, when not empty, is
// written on the row over the keys: the last command the panel ran and what
// came of it.
func renderDeck(pg *deckPage, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout) {
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, pg.Title, xTitle, false)
	g.closeButton(w)
	inner, viewH := w-4, h-deckChrome
	if inner < 12 || viewH < 3 {
		return g, deckLayout{}
	}
	cards, total := layoutDeck(pg, inner)
	lay := deckLayout{cards: cards, total: total, viewH: viewH}
	scroll = clamp(scroll, 0, max(0, total-viewH))

	// The content is drawn whole on a grid of its own and the rows in view
	// are copied in, so only the copy knows about the scroll.
	tall := newXGrid(inner, max(total, 1))
	y, n := 0, 0
	if len(pg.Head) > 0 {
		tall.segsFit(0, y, inner, pg.Head...)
		y += 2
	}
	notes := pg.noteLines(inner)
	for _, l := range notes {
		tall.put(0, y, l.s, l.k)
		y++
	}
	if len(notes) > 0 {
		y++
	}
	for _, s := range pg.Sections {
		meta := ""
		if s.Meta != "" {
			meta = " " + s.Meta
		}
		x := tall.put(0, y, fitText(s.Title, inner-len([]rune(meta))-3)+" ", xTitle)
		for ; x < inner-len([]rune(meta)); x++ {
			tall.set(x, y, '─', xFaint)
		}
		tall.put(x, y, meta, xDim)
		y++
		for _, l := range s.Notes {
			k := l.K
			if k == 0 {
				k = xDim
			}
			tall.putFit(0, y, inner, l.Label+l.Value, k)
			y++
		}
		for i := range s.Cards {
			drawDeckCard(tall, &s.Cards[i], cards[n], n == sel, n)
			y = max(y, cards[n].Max.Y)
			n++
		}
		y++
	}
	for yy := 0; yy < viewH && scroll+yy < total; yy++ {
		for xx := 0; xx < inner; xx++ {
			g.store(2+xx, 1+yy, tall.at(xx, scroll+yy))
		}
	}
	view := image.Rect(2, 1, 2+inner, 1+viewH)
	for _, ht := range tall.hits {
		if r := ht.r.Add(image.Pt(2, 1-scroll)).Intersect(view); !r.Empty() {
			g.hits = append(g.hits, xhit{ht.id, ht.n, r})
		}
	}
	for _, k := range tall.keys {
		if r := k.r.Add(image.Pt(2, 1-scroll)).Intersect(view); !r.Empty() {
			g.keys = append(g.keys, xkey{k.key, r})
		}
	}
	if total > viewH {
		// The scroll mark, on the right edge: where the view is in the whole.
		top, size := 1+scroll*viewH/total, max(1, viewH*viewH/total)
		for i := 0; i < size && top+i <= viewH; i++ {
			g.set(w-1, top+i, '┃', xAccent)
		}
	}
	g.segsFit(2, h-3, inner, said...)
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, deckKeys(pg, sel, total > viewH), pg.Hint)
	return g, lay
}

// deckKeys is the foot's keys: moving, the picked card's action when it has
// one, paging when there is more than fits, and closing.
func deckKeys(pg *deckPage, sel int, pages bool) [][2]string {
	var keys [][2]string
	if pg.count() > 1 {
		keys = append(keys, [2]string{"←↑↓→", vibe.Line("ui.deck.deckKeys.01", "move")})
	}
	if c, ok := pg.card(sel); ok && c.Command != "" {
		keys = append(keys, [2]string{"Enter", c.Action})
	}
	if pages {
		keys = append(keys, [2]string{"PgUp PgDn", vibe.Line("ui.deck.deckKeys.04", "page")})
	}
	return append(keys, [2]string{"Esc", "close"})
}

// drawDeckCard draws card n in r. The picked card has a bright edge and its
// button shows the Enter key; every card with a command shows the command
// beside its button, so what a click does is also what to type.
func drawDeckCard(g *xgrid, c *deckCard, r image.Rectangle, picked bool, n int) {
	x, y, cw := r.Min.X, r.Min.Y, r.Dx()
	edge, tk := xFaint, xBright
	if picked {
		edge = xAccent
	}
	if c.Quiet {
		tk = xDim
	}
	g.box(x, y, cw, r.Dy(), edge, "", xTitle, false)
	g.hit(hitDeckCard, n, x, y, cw, r.Dy())
	if picked {
		g.put(x+1, y, "▸", xAccent)
	}
	chip := len([]rune(c.Chip))
	g.put(x+2, y, " "+fitText(c.Title, max(cw-8-chip, 1))+" ", tk)
	if c.Chip != "" && cw > chip+12 {
		g.put(x+cw-3-chip, y, " "+c.Chip+" ", c.ChipK)
	}
	y++
	if c.Desc != "" {
		for _, l := range deckWrap(c.Desc, cw-4) {
			g.put(x+2, y, l, xDim)
			y++
		}
	}
	lw := 0
	for _, l := range c.Lines {
		lw = max(lw, len([]rune(l.Label)))
	}
	if lw > 0 {
		lw++
	}
	for _, l := range c.Lines {
		k := l.K
		if k == 0 {
			k = xText
		}
		g.put(x+2, y, l.Label, xDim)
		g.putFit(x+2+lw, y, cw-4-lw, l.Value, k)
		y++
	}
	if c.Meter {
		mw := clamp(cw-6-len([]rune(c.MeterText)), 6, 24)
		k := xAccent
		if c.Share >= 1 {
			k = xGood
		}
		advBar(g, x+2, y, mw, c.Share, k)
		g.putFit(x+3+mw, y, cw-5-mw, c.MeterText, xText)
		y++
	}
	if c.Why != "" {
		g.putFit(x+2, y, cw-4, c.Why, xBad)
		y++
	}
	if c.Command == "" {
		return
	}
	end, ck := x+2, xFaint
	if picked {
		end, ck = g.keyButton(x+2, y, "Enter", c.Action), xDim
	} else {
		end = g.put(x+2, y, " "+c.Action+" ", xKeyName)
	}
	g.hit(hitDeckAct, n, x+2, y, end-x-2, 1)
	g.putFit(end+1, y, x+cw-3-end, c.Command, ck)
}

// deckStep is the card a key moves the pick to from sel: Left and Right by
// number, Up and Down to the nearest card in the row above or below.
func (l *deckLayout) step(sel, dx, dy int) int {
	if len(l.cards) == 0 {
		return 0
	}
	sel = clamp(sel, 0, len(l.cards)-1)
	if dx != 0 {
		return clamp(sel+dx, 0, len(l.cards)-1)
	}
	here := l.cards[sel]
	best, bestRow, bestOff := sel, 0, 0
	for i, r := range l.cards {
		row := (r.Min.Y - here.Min.Y) * dy
		if row <= 0 {
			continue
		}
		off := r.Min.X - here.Min.X
		if off < 0 {
			off = -off
		}
		if best == sel || row < bestRow || row == bestRow && off < bestOff {
			best, bestRow, bestOff = i, row, off
		}
	}
	return best
}

// reveal is the scroll that keeps card sel in view, changing scroll as
// little as it can.
func (l *deckLayout) reveal(sel, scroll int) int {
	if sel < 0 || sel >= len(l.cards) || l.viewH <= 0 {
		return 0
	}
	r := l.cards[sel]
	if r.Max.Y > scroll+l.viewH {
		scroll = r.Max.Y - l.viewH
	}
	if r.Min.Y < scroll {
		scroll = r.Min.Y
	}
	// The first card of the page brings the head above it into view too.
	if sel == 0 {
		scroll = 0
	}
	return clamp(scroll, 0, max(0, l.total-l.viewH))
}
