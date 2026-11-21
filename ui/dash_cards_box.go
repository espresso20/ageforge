package ui

import (
	"image"
	"image/color"
	"slices"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// dash_cards_box.go is the dashboard's Buildings box.
//
//	┌─ Buildings ─────────────────────────────────────────────────────┐
//	│ Medieval Age                                  Buy  x1  x10  max │
//	│ [ Gather Food ] [ Gather Wood ]                                 │  (the ages that gather)
//	│▏ Cathedral x1              short of 44K gold · in 1m 44s        │
//	│▏ +0.034 faith/s                                       [ Build ] │
//	│▏ 444K gold · 1.33M stone                                        │
//	│    · · ·  6 more later                                          │
//	└─────────────────────────────────────────────────────────────────┘
//
// Each building is a card of three rows: its name and count, what a copy
// makes in a second, its price. At the left a gauge says how much of the
// price is in hand; at the right
// is the Build button, the same column on every card. A price that does not
// fit the row goes on a row more. A card that cannot be built has its button
// off and says what stops it. The box lists the cards that can be built,
// then the near ones, and counts the rest on a quiet row a click opens.

const (
	// The kinds of thing a click picks (xhit.id).
	hitCard  = 60 // n is the card's place in the layout
	hitMore  = 61 // the quiet row
	hitOlder = 62 // the row that counts the older buildings

	// cardWords is where a card's words start: after the gauge and the mark
	// of the picked card.
	cardWords = 2
	// buildColWide is the Build button's cells with its Enter key, and
	// buildColNarrow without it: the key goes where the box has no room.
	buildColWide, buildColNarrow = 16, 9
	buildKeyFrom                 = 64 // inner width from which the picked card's button shows its key
	// gatherTallRows is the inner height from which the Gather buttons are
	// three rows tall; a shorter box has them one row.
	gatherTallRows = 22
	// gatherAges is how many of the first ages show the Gather buttons.
	gatherAges = 2
	// compactBelow is the inner width under which a card is stacked: its Build
	// button is a one-row key in the name row, so the words have the width.
	compactBelow = 40
	// compactKeyW is the cells of a stacked card's Build key.
	compactKeyW = 7
	// buyKeyBase is the first of the private runes the buy keys press.
	buyKeyBase = 0xE000
)

// buyKey is the key a click on a buy choice presses: a private rune, which
// the box takes before it reaches the app.
func buyKey(c buyChoice) kit.Key { return kit.Key{Code: kit.KeyRune, Rune: rune(buyKeyBase + int(c))} }

// buyOf is the buy choice a key stands for.
func buyOf(k kit.Key) (buyChoice, bool) {
	c := buyChoice(k.Rune - buyKeyBase)
	return c, k.Code == kit.KeyRune && c >= 0 && c < buyChoices
}

// placedCard is a card as laid out: where it is, how many rows it takes and
// the lines of its price.
type placedCard struct {
	card  buildCard
	y     int // grid row of its first row
	rows  int
	price [][]costPart
	tx    int // the column its words start at, in the box's cells
	build int // the index of its Build button in the box's keys
	// compact says the card is stacked (compactBelow). keyRow is the row of
	// its Build key, and reasonRow the row of its reason: the foot row, 0 when
	// the card has none.
	compact           bool
	keyRow, reasonRow int
	// small says the card's Build button is a one-row key: every card but the
	// picked one. build is -1 for a card that has no button.
	small bool
}

// buildBox is the Buildings box.
type buildBox struct {
	*nativeHost
	kit.Base
	title string
	// plain reads the glyph tier: the plain one folds the box to ASCII.
	plain func() bool
	// k presses the box's keys: the buy choices, the Gather buttons and the
	// Build buttons.
	k keyPresser

	st    game.GameState
	has   bool
	buy   buyChoice
	cards []buildCard
	// order is the keys of the cards in the order they are listed and shown
	// how many are before the quiet row. It is kept as it is while the pointer
	// is over the box, so a card never moves from under it, and sorted again
	// when the pointer is away or a building was built (resort, built).
	order  []string
	shown  int
	built  int
	resort bool
	// picked is the card the keyboard's Enter builds, expanded says the quiet
	// row has been opened and first is the first listed thing on show.
	picked   string
	expanded bool
	first    int
	// pickAt is the picked card's place in order: when the card leaves the
	// list the card now in that place is picked. reveal asks the next draw to
	// scroll the pick into view.
	pickAt int
	reveal bool
	// nextKey is the card of the next step, which carries a plate.
	nextKey string
	// older says the box lists the older buildings the town holds in place of
	// the ones to build, and olderN is how many kinds there are.
	older  bool
	olderN int
	// offer is the engine's upgrade offer for a building.
	offer func(key string) game.UpgradeOffer
	// detail is what the detail box says, det its keys and noteRows the cells of
	// its note; det is nil when the box has no room for it.
	detail     detailInfo
	det        *keySet
	noteRows   []image.Rectangle
	noteBeside bool

	// the layout
	grid  *xgrid
	w, h  int
	dirty bool
	// pickTip asks the next paint to aim the tooltip at the card the keys
	// just picked.
	pickTip  bool
	placed   []placedCard
	more     int // the grid row of the quiet row, -1 for none
	olderRow int // the grid row of the older buildings' row, -1 for none
	// strip and cardKeys are the keys of the strip and of the cards, for
	// painting each in its own way.
	strip, cardKeys *keySet
	// shownItems is how many listed things the box showed and items how many
	// there are, for the wheel.
	shownItems, items, hidden int

	// the pointer: the cell it is over in the box's cells, whether it is over
	// the box, and the card under it (-1 for none).
	over  image.Point
	in    bool
	hover int

	imgs     map[cardPic]*image.RGBA
	imgBytes int
	imgCell  [2]int
	// laidCell is the cell size the layout was made for.
	laidCell [2]int
}

func newBuildBox(plain func() bool, app *appLoop) *buildBox {
	b := &buildBox{plain: plain, more: -1, olderRow: -1, hover: -1}
	b.nativeHost = newNativeHost()
	b.k.wire(app)
	press := b.k.press
	b.k.press = func(k kit.Key) {
		if c, ok := buyOf(k); ok {
			b.setBuy(c)
			return
		}
		press(k)
	}
	return b
}

// update takes a new snapshot of the game: the cards are made again from it.
func (b *buildBox) update(st game.GameState) {
	b.st, b.has = st, true
	b.recompute()
}

// setBuy changes how many copies a Build button builds and prices.
func (b *buildBox) setBuy(c buyChoice) {
	if c == b.buy {
		return
	}
	b.buy = c
	b.recompute()
}

// recompute makes the cards and puts them in order, unless the order is held.
func (b *buildBox) recompute() {
	older := olderCards(b.st)
	b.olderN = len(older)
	if b.older && b.olderN == 0 {
		b.older = false
	}
	if b.older {
		b.cards = older
	} else {
		b.cards = buildCards(b.st, b.buy)
	}
	built := len(b.st.BuildQueue)
	for _, bs := range b.st.Buildings {
		built += bs.Count
	}
	have := make(map[string]bool, len(b.cards))
	for _, c := range b.cards {
		have[c.key] = true
	}
	if b.order == nil || b.resort || !b.in || built != b.built {
		b.order, b.shown = orderCards(b.cards)
		b.resort = false
	} else {
		// The order stays: what is gone leaves it and what is new comes last.
		var kept []string
		shown := 0
		seen := map[string]bool{}
		for i, k := range b.order {
			if have[k] {
				kept = append(kept, k)
				seen[k] = true
				if i < b.shown {
					shown++
				}
			}
		}
		for _, c := range b.cards {
			if !seen[c.key] {
				kept = append(kept, c.key)
			}
		}
		b.order, b.shown = kept, shown
	}
	b.settlePick()
	b.built = built
	b.refreshDetail()
	b.dirty = true
}

// settlePick keeps the same building picked when the list is sorted again:
// the pick is the card, not its row. When the card has left the list, the card
// now in its old place is picked (the last one when that is past the end). A
// pick that moved is scrolled into view the next time the box is drawn.
func (b *buildBox) settlePick() {
	if b.picked == "" {
		return
	}
	if at := slices.Index(b.order, b.picked); at >= 0 {
		b.reveal = b.reveal || at != b.pickAt
		b.pickAt = at
		return
	}
	if len(b.order) == 0 {
		b.picked = ""
		return
	}
	b.pickAt = min(b.pickAt, len(b.order)-1)
	b.picked, b.reveal = b.order[b.pickAt], true
}

// pick picks the card of key.
func (b *buildBox) pick(key string) {
	b.picked, b.pickAt, b.pickTip = key, max(slices.Index(b.order, key), 0), !b.in
	b.refreshDetail()
	b.dirty = true
}

// show picks the card of key and brings it into view, as a click on it would
// leave it, opening the quiet row when it is in it.
func (b *buildBox) show(key string) {
	b.closeOlder()
	if i := slices.Index(b.order, key); i >= b.shown {
		b.expanded = true
	}
	b.pick(key)
	b.reveal = true
}

// toggleOlder switches the box between the buildings to build and the older
// ones held.
func (b *buildBox) toggleOlder() {
	b.older = !b.older
	b.picked, b.pickAt, b.first, b.expanded, b.resort = "", 0, 0, false, true
	b.recompute()
}

// closeOlder returns the box to its build list, reporting whether it was
// showing the older buildings.
func (b *buildBox) closeOlder() bool {
	if !b.older {
		return false
	}
	b.toggleOlder()
	return true
}

// card is the card of key.
func (b *buildBox) card(key string) (buildCard, bool) {
	for _, c := range b.cards {
		if c.key == key {
			return c, true
		}
	}
	return buildCard{}, false
}

// pickedLine is the command the picked card's Build button carries, "" when
// no card is picked or it cannot be built.
func (b *buildBox) pickedLine() string {
	if c, ok := b.card(b.picked); ok && c.canBuild {
		return c.cmd
	}
	return ""
}

// inner is the cells inside the box's frame.
func (b *buildBox) inner() (w, h int) { return b.w - 2, b.h - 2 }

// picCells is the cell size the cards' pictures are drawn at, zero for none.
func (b *buildBox) picCells() (cw, ch int) {
	if b.plain != nil && b.plain() {
		return 0, 0
	}
	return pictureCells()
}

// ===== Layout =====

// Layout takes the box's rectangle and lays the cards out for it.
func (b *buildBox) Layout(r image.Rectangle) {
	b.Place(r)
	w, h := r.Dx(), r.Dy()
	if w <= 0 || h <= 0 {
		b.grid = nil
		return
	}
	if cw, ch := b.picCells(); b.grid == nil || w != b.w || h != b.h || b.dirty || [2]int{cw, ch} != b.laidCell {
		b.draw(w, h)
	}
}

// CanFocus: the box takes no keys.
func (b *buildBox) CanFocus() bool { return false }

// buildCol is the cells the Build column takes in an inner width.
func buildCol(inner int) int {
	if inner >= buildKeyFrom {
		return buildColWide
	}
	return buildColNarrow
}

// priceLines breaks a price into lines at most room cells wide, between its
// resources. A resource wider than the room ("49.8T dark matter crystals" in
// a narrow box) goes on over several lines at its spaces, so none of it is cut.
func priceLines(parts []costPart, room int) [][]costPart {
	var lines [][]costPart
	cur, used := []costPart(nil), 0
	flush := func() {
		if len(cur) > 0 {
			lines = append(lines, cur)
		}
		cur, used = nil, 0
	}
	for _, p := range parts {
		n := runeLen(p.text)
		if n > room {
			flush()
			for _, chunk := range wrapWords(p.text, max(room, 1)) {
				lines = append(lines, []costPart{{text: chunk, ok: p.ok}})
			}
			continue
		}
		if len(cur) > 0 && used+3+n > room {
			flush()
		}
		if len(cur) > 0 {
			used += 3
		}
		cur = append(cur, p)
		used += n
	}
	flush()
	if len(lines) == 0 {
		lines = [][]costPart{nil} // a building that costs nothing keeps its row
	}
	return lines
}

// draw lays the box out for w by h cells. The detail of the picked building
// goes under the cards when it costs no card its place; a box too short for
// both shows the cards.
func (b *buildBox) draw(w, h int) {
	b.w, b.h, b.dirty = w, h, false
	cw, ch := b.picCells()
	b.laidCell = [2]int{cw, ch}
	b.drawAt(w, h, 0)
	if !b.detail.ok || w-2 < detailMinW {
		return
	}
	base := b.shownItems
	b.drawAt(w, h, detailRows)
	// A list that already runs past the box keeps the detail while a few
	// cards still show beside it; a list that fits keeps every card.
	fits := base >= b.items
	if (fits && b.shownItems < base) || b.shownItems < min(base, detailKeepCards) {
		b.drawAt(w, h, 0)
	}
	if b.det != nil {
		b.growDetail(w, h)
	}
}

// growDetail gives the detail the rows it wants when the box has them to
// spare: the most rows that leave every card, and every row the list uses,
// where the detail's frame of detailRows left them.
func (b *buildBox) growDetail(w, h int) {
	first, shown, hidden, more, older := b.first, b.shownItems, b.hidden, b.more >= 0, b.olderRow >= 0
	for n := b.detailWant(w); n > detailRows; n-- {
		b.drawAt(w, h, n)
		if b.first == first && b.shownItems == shown && b.hidden == hidden && (b.more >= 0) == more && (b.olderRow >= 0) == older {
			return
		}
	}
	b.drawAt(w, h, detailRows)
}

// drawAt lays the box out with d rows at the bottom kept for the detail
// (none when 0): the frame, the strip, and the cards from the first one on
// show.
func (b *buildBox) drawAt(w, h, d int) {
	g := newXGrid(w, h)
	b.grid, b.placed, b.more, b.olderRow, b.det = g, nil, -1, -1, nil
	bot := h - 1 - d
	g.box(0, 0, w, h, xBorder, strings.TrimSpace(b.title), xTitle, false)
	iw, ih := w-2, h-2
	seen := b.shownItems
	b.strip, b.cardKeys, b.shownItems, b.items = nil, nil, 0, 0
	if !b.has || iw < 8 || ih < 1 {
		return
	}
	set := g.keyset()
	// The strip: the age and the buy choices; in the ages that gather, the
	// Gather buttons under them.
	stripFrom := len(set.keys)
	words := make([]string, buyChoices)
	buyW := 0
	for c := range words {
		words[c] = " " + buyWords[c] + " "
		buyW += runeLen(words[c]) + 1
	}
	buyW--
	label := txtBuy.String() + " "
	if iw < 36 {
		label = ""
	}
	bx := w - 2 - buyW - runeLen(label)
	g.putFit(2, 1, max(bx-3, 1), b.st.AgeName, xTitle)
	x := g.put(bx, 1, label, xDim)
	for c := buyChoice(0); c < buyChoices; c++ {
		r := image.Rect(x, 1, x+runeLen(words[c]), 2)
		g.put(x, 1, words[c], xKeycap)
		kind := keySteel
		if c == b.buy {
			kind = keyAccent
		}
		set.keys = append(set.keys, xkey{key: buyKey(c), r: r, cap: r, kind: kind})
		x = r.Max.X + 1
	}
	top := 3
	if b.st.CanGatherByHand() && b.gatherAge() {
		tall := ih >= gatherTallRows
		rowH := 1
		if tall {
			rowH = actionRows
		}
		gx, gy := 2, top
		for _, res := range []string{"food", "wood", "stone"} {
			rs, ok := b.st.Resources[res]
			if !ok || !rs.Unlocked {
				continue
			}
			name := rowid.Sprintf("ui.dash_cards_box.draw.01", "Gather %s", rs.Name)
			cmd := command("gather", res)
			word := " " + name + " "
			wide := runeLen(word)
			if tall {
				wide = actionW(name, "", signNone)
			}
			if gx > 2 && gx+wide > w-1 {
				gx, gy = 2, gy+rowH+min(1, rowH-1)
			}
			if tall {
				gx = g.actionButton(gx, gy, actionSpec{kind: keyPlain, label: name, cmd: cmd}) + 1
				continue
			}
			r := image.Rect(gx, gy, gx+wide, gy+1)
			g.put(gx, gy, word, xKeycap)
			set.keys = append(set.keys, xkey{r: r, cap: r, kind: keySteel, cmd: cmd})
			gx = r.Max.X + 1
		}
		if gx > 2 || gy > top {
			top = gy + rowH + 1
		}
	}
	b.strip = &keySet{keys: append([]xkey(nil), set.keys[stripFrom:]...), plain: g.plain, glyph: set.glyph}

	// The cards: the ones before the quiet row, and once it has been opened
	// the rest. The quiet row goes under the last card on show and counts the
	// cards that are not on show below it.
	items := b.order[:b.shown]
	laterHidden := len(b.order) - b.shown
	if b.expanded {
		items, laterHidden = b.order, 0
	}
	b.items = len(items)
	b.first = min(max(b.first, 0), max(len(items)-1, 0))
	if b.reveal {
		if i := slices.Index(items, b.picked); i >= 0 && (i < b.first || i >= b.first+max(seen, 1)) {
			b.first = i
		}
		b.reveal = false
	}
	bw := buildCol(iw)
	compact := iw < compactBelow
	tx := 1 + cardWords
	room := (1 + iw) - bw - 2 - tx // the words' cells
	if compact {
		room = iw - cardWords - 1
	}
	cardFrom := len(set.keys)
	y := top
	for i := b.first; i < len(items); i++ {
		c, ok := b.card(items[i])
		if !ok {
			continue
		}
		lines := priceLines(c.price, room)
		rows := 2 + len(lines)
		// A reason goes at the right of the name row when all of it fits
		// there, else on a row of its own under the price. A stacked card's
		// Build key sits at the right of the last price line when that leaves
		// room for it, else on a row of its own.
		keyRow, reasonRow := 0, 0
		if c.reason != "" {
			full := c.reason
			if c.wait != "" {
				full += " · " + c.wait
			}
			if compact || runeLen(c.name+" x"+textfmt.Int(c.owned))+2+runeLen(full) > room {
				reasonRow, rows = rows, rows+1
			}
		}
		if compact {
			last := 0
			for j, part := range lines[len(lines)-1] {
				last += runeLen(part.text) + 3*min(j, 1)
			}
			if last+1+compactKeyW <= room {
				keyRow = 2 + len(lines) - 1
			} else {
				keyRow, rows = rows, rows+1
			}
		}
		need := rows
		if len(items)-(i+1)+laterHidden > 0 {
			need++ // the quiet row stays under it
		}
		if b.olderN > 0 {
			need++ // and the row of the older buildings
		}
		if y+need > bot {
			// A box too short for even the first card shows what fits of its
			// price; otherwise the card waits for the next page.
			if b.shownItems > 0 || bot-y < 3 {
				break
			}
			cut := rows - (bot - y)
			lines = lines[:max(len(lines)-cut, 1)]
			rows, keyRow, reasonRow = 2+len(lines), 0, 0
			if compact {
				keyRow = rows - 1
			}
		}
		pc := placedCard{card: c, y: y, rows: rows, price: lines, tx: tx, compact: compact, keyRow: keyRow, reasonRow: reasonRow, build: -1}
		switch {
		case c.older:
		case compact:
			kx := 1 + iw - compactKeyW - 1
			word := " " + txtBuild.String() + " "
			r := image.Rect(kx, y+keyRow, kx+compactKeyW, y+keyRow+1)
			g.put(kx, y+keyRow, fitText(word, compactKeyW), xKeycap)
			set.keys = append(set.keys, xkey{r: r, cap: r, kind: keyAccent, cmd: c.cmd, off: !c.canBuild, why: c.reason})
		case c.key == b.picked:
			// The picked card has the one primary button.
			by := y + (rows-actionRows)/2
			bxx := 1 + iw - bw - 1
			spec := actionSpec{kind: keyPrimary, label: txtBuild.String(), cmd: c.cmd, off: !c.canBuild, why: c.reason}
			if c.canBuild && bw == buildColWide {
				spec.key = "Enter"
			}
			g.fill(bxx, by, bw, actionRows)
			at := bxx + (bw-actionW(spec.label, spec.key, signNone))/2
			set.action(at, by, spec, func(x, y int, text string) { g.put(x, y, text, xText) })
			last := &set.keys[len(set.keys)-1]
			last.r = image.Rect(bxx, by, bxx+bw, by+actionRows)
		default:
			// Every other card has a one-row key on its middle row, at the
			// right of the Build column.
			kx, ky := 1+iw-compactKeyW-1, y+(rows-actionRows)/2+1
			r := image.Rect(kx, ky, kx+compactKeyW, ky+1)
			g.put(kx, ky, fitText(" "+txtBuild.String()+" ", compactKeyW), xKeycap)
			set.keys = append(set.keys, xkey{r: r, cap: r, kind: keyAccent, cmd: c.cmd, off: !c.canBuild, why: c.reason})
			pc.small = true
		}
		if !c.older {
			pc.build = len(set.keys) - 1 - cardFrom
		}
		g.hit(hitCard, len(b.placed), 1, y, iw, rows)
		b.placed = append(b.placed, pc)
		y += rows
		b.shownItems++
	}
	b.hidden = len(items) - (b.first + b.shownItems) + laterHidden
	if b.hidden > 0 && y < bot {
		b.more = y
		g.put(tx, y, "· · ·", xDim)
		g.put(tx+6, y, rowid.Sprintf("ui.dash_cards_box.draw.02", "%s more later", textfmt.Int(b.hidden)), xDim)
		g.hit(hitMore, 0, 1, y, iw, 1)
		y++
	}
	if b.olderN > 0 && y < bot {
		b.olderRow = y
		look := xDim
		if b.older {
			look = xText
		}
		g.put(tx, y, "· · ·", xDim)
		g.put(tx+6, y, textfmt.Int(b.olderN)+" "+txtOlder.String(), look)
		g.hit(hitOlder, 0, 1, y, iw, 1)
	}
	b.cardKeys = &keySet{keys: append([]xkey(nil), set.keys[cardFrom:]...), plain: g.plain, glyph: set.glyph}
	if d > 0 {
		detFrom := len(set.keys)
		b.drawDetail(g, set, w, h, d)
		b.det = &keySet{keys: append([]xkey(nil), set.keys[detFrom:]...), plain: g.plain, glyph: set.glyph}
	}
	if b.plain != nil && b.plain() {
		g.fold()
		b.strip.plain, b.cardKeys.plain = true, true
		if b.det != nil {
			b.det.plain = true
		}
	}
}

// ===== Mouse =====

// cardAt is the card under the cell pt of the box, -1 for none.
func (b *buildBox) cardAt(pt image.Point) int {
	if b.grid == nil {
		return -1
	}
	if n, ok := b.grid.hitAt(hitCard, pt); ok {
		return n
	}
	return -1
}

// Mouse: a click on a Build button or a buy choice or a Gather button
// presses it; a click on a card picks it and one on the quiet row opens or
// closes it; the wheel turns the list. A move that changes what is lit asks
// for a repaint.
func (b *buildBox) Mouse(m kit.Mouse) bool {
	org := b.Bounds().Min
	was := b.hover
	if b.in = m.Kind != kit.MouseLeave; b.in {
		b.over = m.Pt().Sub(org)
	}
	b.hover = -1
	if b.in {
		b.hover = b.cardAt(b.over)
	}
	if m.Kind == kit.MouseMove {
		b.aimPrice(org)
	}
	if b.grid != nil && b.k.mouse(b.grid, org, m) {
		return true
	}
	switch {
	case m.Kind == kit.MouseWheel:
		// One listed thing a turn, however far the wheel went.
		if m.Wheel < 0 {
			return b.scroll(-1)
		}
		return b.scroll(1)
	case m.Kind == kit.MousePress && m.Button == kit.ButtonLeft && b.in && b.grid != nil:
		if _, ok := b.grid.hitAt(hitMore, b.over); ok {
			b.moreClicked()
			return true
		}
		if _, ok := b.grid.hitAt(hitOlder, b.over); ok {
			b.toggleOlder()
			return true
		}
		if n := b.cardAt(b.over); n >= 0 && n < len(b.placed) {
			b.pick(b.placed[n].card.key)
			return true
		}
	}
	return was != b.hover
}

// tipPrice is the target of the tooltip of one card's price.
type tipPrice struct {
	b   *buildBox
	key string
}

// priceRect is the cells of a card's price lines, in the box's cells.
func (b *buildBox) priceRect(pc placedCard) (image.Rectangle, bool) {
	w := 0
	for _, line := range pc.price {
		n := 0
		for j, part := range line {
			n += runeLen(part.text) + 3*min(j, 1)
		}
		w = max(w, n)
	}
	return image.Rect(pc.tx, pc.y+2, pc.tx+w, pc.y+2+len(pc.price)), len(pc.card.price) > 0
}

// aimPrice aims the tooltip at the price of the card under the pointer when
// the pointer is on its price lines.
func (b *buildBox) aimPrice(org image.Point) {
	if b.hover < 0 || b.hover >= len(b.placed) {
		return
	}
	pc := b.placed[b.hover]
	if r, ok := b.priceRect(pc); ok && b.over.In(r) {
		b.k.sense().tip.Aim(tipPrice{b, pc.card.key}, r.Add(org), priceTip(pc.card.price), false)
	}
}

// aimPicked aims the tooltip at the card the keys picked: its Build button
// when that is off and has a reason, else its price.
func (b *buildBox) aimPicked(org image.Point) {
	for _, pc := range b.placed {
		if pc.card.key != b.picked || pc.build < 0 {
			continue
		}
		if k := b.cardKeys.keys[pc.build]; k.off && k.why != "" {
			b.k.sense().tip.Aim(tipKey{&b.k, k.r}, k.r.Add(org), kit.TipText{Reason: k.why}, true)
		} else if r, ok := b.priceRect(pc); ok {
			b.k.sense().tip.Aim(tipPrice{b, pc.card.key}, r.Add(org), priceTip(pc.card.price), true)
		}
	}
}

// moreClicked is a click on the quiet row: it opens the rest of the cards;
// once they are open it turns a page toward the ones still below.
func (b *buildBox) moreClicked() {
	b.dirty = true
	if !b.expanded {
		b.expanded = true
		return
	}
	b.page(1)
}

// scroll turns the list by n listed things, down for a positive n, and
// opens the quiet row when it is turned past it. It reports whether
// anything changed.
func (b *buildBox) scroll(n int) bool {
	if n == 0 {
		return false
	}
	if n < 0 {
		if b.first == 0 {
			return false
		}
		b.first = max(b.first+n, 0)
		b.dirty = true
		return true
	}
	if b.first+b.shownItems >= b.items {
		if b.expanded || len(b.order) == b.shown {
			return false
		}
		b.expanded = true
		b.dirty = true
		return true
	}
	b.first = min(b.first+n, b.items-1)
	b.dirty = true
	return true
}

// page turns the list a page, down for a positive dir.
func (b *buildBox) page(dir int) {
	b.scroll(dir * max(b.shownItems, 1))
}

// PointerAt: the hand over a card, the quiet row and any key a click presses.
func (b *buildBox) PointerAt(at image.Point) kit.Pointer {
	org := b.Bounds().Min
	if b.grid == nil {
		return kit.PointerArrow
	}
	if b.k.pointer(b.grid, org, at) == kit.PointerHand {
		return kit.PointerHand
	}
	pt := at.Sub(org)
	_, more := b.grid.hitAt(hitMore, pt)
	return handIf(more || b.cardAt(pt) >= 0)
}

// ===== Paint =====

// cardLook is the colours a card is drawn in.
type cardLook struct{ fill, light, dark, text, dim color.RGBA }

// cardLookOf is the look of a card in a state.
func cardLookOf(pal *expPalette, picked, hover, afford bool) cardLook {
	g, tx := pal.tone(xText, true), pal.tone(xText, false)
	dim := pal.tone(xDim, false)
	acc := pal.tone(xAccent, false)
	switch {
	case picked:
		return cardLook{toward(g, acc, .2), toward(acc, keyWhite, .25), toward(acc, keyBlack, .45), tx, dim}
	case hover:
		return cardLook{toward(g, acc, .12), toward(acc, keyWhite, .2), toward(acc, keyBlack, .4), tx, dim}
	case !afford:
		return cardLook{toward(g, tx, .03), toward(g, tx, .14), toward(g, keyBlack, .25), dim, toward(dim, g, .3)}
	}
	return cardLook{toward(g, tx, .07), toward(g, tx, .3), toward(g, keyBlack, .4), tx, dim}
}

// cardPic is what a card's picture was drawn for.
type cardPic struct {
	w, rows, cw, ch      int
	look                 cardLook
	ground, accent, good color.RGBA
	gauge                int // the filled part of the gauge, in units
	picked, hover, can   bool
}

const cardPicMax = 8 << 20

// picture is the picture of a card: its slab, gauge and edge, in the state it
// is in. A picture that has been painted is never drawn on again, so each
// state keeps its own.
func (b *buildBox) picture(k cardPic) *image.RGBA {
	if b.imgCell != [2]int{k.cw, k.ch} || b.imgBytes > cardPicMax {
		b.imgs, b.imgCell, b.imgBytes = nil, [2]int{k.cw, k.ch}, 0
	}
	img, ok := b.imgs[k]
	if !ok {
		if b.imgs == nil {
			b.imgs = map[cardPic]*image.RGBA{}
		}
		img = drawCard(k)
		b.imgs[k] = img
		b.imgBytes += len(img.Pix)
	}
	return img
}

// drawCard draws a card's picture in units of a pixel of a 10 by 21 cell: a
// raised slab lit from the upper left, an accent ring when picked and a bar at
// the left when the pointer is on it, and the gauge.
func drawCard(k cardPic) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, k.w*k.cw, k.rows*k.ch))
	u := max(1, k.ch/21)
	a := art{img, u}
	W, H := (img.Rect.Dx()+u-1)/u, (img.Rect.Dy()+u-1)/u
	body := image.Rect(1, 0, W-1, H)
	a.fill(body, 1, k.look.fill)
	a.hline(body.Min.X+1, body.Max.X-1, body.Min.Y, k.look.light)
	a.hline(body.Min.X+1, body.Max.X-1, body.Max.Y-1, k.look.dark)
	for y := body.Min.Y + 1; y < body.Max.Y-1; y++ {
		a.px(body.Min.X, y, k.look.light)
		a.px(body.Max.X-1, y, k.look.dark)
	}
	if k.picked {
		for i := 0; i < 2; i++ {
			q := body.Inset(i)
			a.hline(q.Min.X+1, q.Max.X-1, q.Min.Y, k.accent)
			a.hline(q.Min.X+1, q.Max.X-1, q.Max.Y-1, k.accent)
			for y := q.Min.Y; y < q.Max.Y; y++ {
				a.px(q.Min.X, y, k.accent)
				a.px(q.Max.X-1, y, k.accent)
			}
		}
	}
	if k.hover {
		a.fill(image.Rect(2, 3, 5, H-3), 0, k.accent)
	}
	track := image.Rect(5, 8, 7, H-8)
	a.fill(track, 0, toward(k.ground, keyBlack, .3))
	fill := k.accent
	if k.can {
		fill = k.good
	}
	a.fill(image.Rect(5, track.Max.Y-k.gauge, 7, track.Max.Y), 0, fill)
	return img
}

// inkOf is a colour as a bold or plain ink.
func inkOf(c color.RGBA, bold bool) kit.Style {
	return kit.Style{}.Foreground(theme.RGB(c.R, c.G, c.B)).Bold(bold)
}

// Paint puts the box on the painter in the theme's colours.
func (b *buildBox) Paint(p *kit.Painter) {
	r := b.Bounds()
	if r.Empty() || b.grid == nil {
		return
	}
	p = p.Clip(r)
	pal := newExpPalette()
	paintExpeditionKit(p, b.grid, r.Min, pal)
	plain := b.plain != nil && b.plain()
	cw, ch := 0, 0
	if pictureCell != nil && !plain {
		cw, ch = pictureCell()
	}
	if b.pickTip {
		b.pickTip = false
		b.aimPicked(r.Min)
	}
	hot, isHot := b.k.lit(b.cardKeys)
	for i := range b.placed {
		b.paintCard(p, pal, r.Min, i, cw, ch, plain, hot, isHot)
	}
	b.k.paint(p, b.strip, r.Min, pal)
	b.paintDetail(p, pal, r.Min, cw > 0 && ch > 0, plain)
}

// paintCard paints card i: its ground, its picture, its words and its Build
// button.
func (b *buildBox) paintCard(p *kit.Painter, pal *expPalette, org image.Point, i, cw, ch int, plain bool, hot xkey, isHot bool) {
	pc := b.placed[i]
	c := pc.card
	iw, _ := b.inner()
	rect := image.Rect(org.X+1, org.Y+pc.y, org.X+1+iw, org.Y+pc.y+pc.rows)
	picked := c.key == b.picked
	look := cardLookOf(pal, picked, b.hover == i, c.canBuild)
	ground := pal.tone(xText, true)
	fill := theme.RGB(look.fill.R, look.fill.G, look.fill.B)
	pix := cw > 0 && ch > 0
	p.Fill(rect, fill)
	if pix {
		units := max(1, ch/21)
		track := (pc.rows*ch+units-1)/units - 16
		p.Picture(rect, b.picture(cardPic{w: iw, rows: pc.rows, cw: cw, ch: ch, look: look, ground: ground,
			accent: pal.tone(xAccent, false), good: pal.tone(xGood, false), gauge: int(float64(track)*c.frac + .5),
			picked: picked, hover: b.hover == i, can: c.canBuild}))
	}
	if !pix {
		b.paintCells(p, pal, rect, look, c, picked, pc.compact)
	}
	lab := func(x, y int, s string, st kit.Style) int {
		if plain {
			s = foldToASCII(s)
		}
		return p.Label(rect.Min.X+x, rect.Min.Y+y, kit.Escape(s), st)
	}
	bw := buildCol(iw)
	tx, room := cardWords, iw-bw-2-cardWords
	if pc.compact {
		room = iw - cardWords - 1
	}
	count := " x" + textfmt.Int(c.owned)
	n := lab(tx, 0, fitText(c.name, max(room-runeLen(count), 1)), inkOf(look.text, true))
	lab(tx+n, 0, count, inkOf(look.dim, false))
	from := tx + n + runeLen(count)
	if c.key == b.nextKey && b.nextKey != "" {
		plate := txtNextStep.String()
		if at := from + 1; at+runeLen(plate) <= tx+room {
			// Label keeps the cell's own background, so the plate fills its cells first.
			p.Fill(image.Rect(rect.Min.X+at, rect.Min.Y, rect.Min.X+at+runeLen(plate), rect.Min.Y+1), theme.Of(theme.RoleLabel))
			lab(at, 0, plate, kit.Style{}.Foreground(theme.Of(theme.RoleBackground)).Bold(true))
			from = at + runeLen(plate)
		}
	}
	if c.reason != "" {
		row, to := 0, tx+room
		if pc.reasonRow > 0 {
			row, from = pc.reasonRow, tx
		}
		b.paintReason(lab, pal, look, c, row, from, to)
	}
	lab(tx, 1, fitText(c.out, room), inkOf(toward(look.text, ground, .15), false))
	for li, line := range pc.price {
		x := tx
		for pi, part := range line {
			if pi > 0 {
				x += lab(x, 2+li, " · ", inkOf(look.dim, false))
			}
			fg, bold := look.text, true
			if !part.ok {
				fg, bold = toward(look.dim, ground, .1), false
			}
			x += lab(x, 2+li, fitText(part.text, max(tx+room-x, 1)), inkOf(fg, bold))
		}
	}
	// The Build button, on the card's own ground.
	cp := *pal
	cp.styles[xText] = cp.styles[xText].Background(fill)
	if pc.build < 0 {
		return
	}
	key := b.cardKeys.keys[pc.build]
	st := b.stateOf(key, hot, isHot)
	if pc.compact || pc.small {
		b.k.paintKey(p, b.cardKeys, org, &cp, key, st, pix)
		return
	}
	b.k.paintAction(p, b.cardKeys, org, &cp, key, st, pix)
}

// paintReason writes what stops a card: the reason, then the wait, as much as
// there is room for, at the right of the name row from column from, or from
// column from on row when the card gives it a row of its own.
func (b *buildBox) paintReason(lab func(x, y int, s string, st kit.Style) int, pal *expPalette, look cardLook, c buildCard, row, from, to int) {
	bad := pal.tone(xBad, false)
	room := to - from
	if row == 0 {
		room -= 2
	}
	full := c.reason
	if c.wait != "" {
		full += " · " + c.wait
	}
	x := to - runeLen(full)
	if row > 0 {
		x = from
	}
	switch {
	case runeLen(full) <= room:
		n := lab(x, row, c.reason, inkOf(bad, false))
		if c.wait != "" {
			lab(x+n, row, " · "+c.wait, inkOf(look.dim, false))
		}
	case runeLen(c.reason) <= room:
		x = to - runeLen(c.reason)
		if row > 0 {
			x = from
		}
		lab(x, row, c.reason, inkOf(bad, false))
	case room >= 6:
		lab(from+(to-from-room), row, fitText(c.reason, room), inkOf(bad, false))
	}
}

// paintCells draws a card's left edge in cells, for a window with no
// pictures: the gauge in the first column and the picked mark.
func (b *buildBox) paintCells(p *kit.Painter, pal *expPalette, rect image.Rectangle, look cardLook, c buildCard, picked, compact bool) {
	acc := pal.tone(xAccent, false)
	plain := b.plain != nil && b.plain()
	fill := theme.RGB(look.fill.R, look.fill.G, look.fill.B)
	lit := int(c.frac*float64(rect.Dy()) + .5)
	for row := 0; row < rect.Dy(); row++ {
		g, col := '│', look.dim
		if rect.Dy()-row <= lit {
			g, col = '┃', acc
			if c.canBuild {
				col = pal.tone(xGood, false)
			}
		}
		if plain {
			g = '|'
			if col != look.dim {
				g = '#'
			}
		}
		p.Put(rect.Min.X, rect.Min.Y+row, g, inkOf(col, false))
	}
	if picked && !compact {
		mark := '▸'
		if plain {
			mark = '>'
		}
		p.Put(rect.Min.X+1, rect.Min.Y+rect.Dy()/2, mark, inkOf(acc, true).Background(fill))
	}
}

// lightGround reports whether a ground colour is a light one.
func lightGround(c color.RGBA) bool {
	return 0.299*float64(c.R)+0.587*float64(c.G)+0.114*float64(c.B) > 140
}

// foldToASCII is s in ASCII, for the plain glyph tier.
func foldToASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if f, ok := foldExpedition[r]; ok {
			return f
		}
		return r
	}, s)
}

// gatherAge says whether the age is one of the first that show the Gather
// buttons.
func (b *buildBox) gatherAge() bool {
	i, ok := b.st.Ruleset().Index(b.st.Age)
	return ok && i < gatherAges
}
