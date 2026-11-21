package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// dash_detail.go is the framed box under the Buildings box's cards, for the
// picked building:
//
//	┌───────────────────────────────────────────────────────────────────┐
//	│ Hut x12                                       Manor · 63.8K gold  │
//	│ A crude shelter of sticks and leaves. +10 housing.   ┌────┐ ┌────┐ │
//	│ [-] [+] workers: 3/6                                 │Sell│ │Upg.│ │
//	│ Four walls and the ambition of someday having more.  └────┘ └────┘ │
//	└───────────────────────────────────────────────────────────────────┘
//
// Its name and count, its description and flavour, its workers with a minus
// and a plus that run the assign commands, and Sell and Upgrade buttons that
// run theirs. A button that cannot run is off, and the line at the top right
// says why in the engine's words (the reason of the button the pointer is on,
// else the first that matters). The box gives way before any card is hidden.

const (
	// detailRows is the rows of the frame, and detailMinW the inner width
	// under which the box leaves the detail out.
	detailRows = 6
	detailMinW = 46
	// detailKeepCards is the cards a long list keeps on show beside the detail.
	detailKeepCards = 3
	// detailRoomW is the cells the text keeps beside the buttons.
	detailRoomW = 24
)

// The command words the detail's buttons carry.
const (
	verbSell     = "sell"
	verbUpgrade  = "upgrade"
	verbAssign   = "assign"
	verbUnassign = "unassign"
)

// detailInfo is what the detail says about the picked building.
type detailInfo struct {
	ok bool
	bs game.BuildingState
	// The commands, whether each can run and why not.
	sellCmd, upCmd string
	// sellAsk is the question before a sale.
	sellAsk        string
	sellOff, upOff bool
	sellWhy, upWhy string
	minus, plus    stepSide
	workers        string
	upLine         string // what an upgrade makes and costs, for an older building
}

// sellCount is how many copies the sell command for a buy choice sells.
func sellCount(buy buyChoice, owned int) int {
	switch {
	case buy == buyTen && owned > 0:
		return min(10, owned)
	case buy == buyTen:
		return 10
	case buy == buyMax && owned > 0:
		return owned
	}
	return 1
}

// sellAsk is the question before a sale of n copies of the building name.
func sellAsk(name string, n int) string {
	return rowid.Sprintf("ui.dash_detail.sellAsk.01", "Sell %d x %s?", n, name)
}

// sellCommand is the sell command for the amount chosen: one copy, ten, or
// every copy held.
func sellCommand(key string, buy buyChoice, owned int) string {
	if n := sellCount(buy, owned); n > 1 {
		return command(verbSell, key, strconv.Itoa(n))
	}
	return command(verbSell, key)
}

// upgradeCommand is the upgrade command for the amount chosen: without a count
// the engine upgrades every copy, so one copy is asked for as 1.
func upgradeCommand(key string, buy buyChoice) string {
	switch buy {
	case buyTen:
		return command(verbUpgrade, key, "10")
	case buyMax:
		return command(verbUpgrade, key, "all")
	}
	return command(verbUpgrade, key, "1")
}

// refreshDetail takes what the detail says from the snapshot and the engine's
// offer for the picked building.
func (b *buildBox) refreshDetail() {
	b.detail = detailInfo{}
	bs, ok := b.st.Buildings[b.picked]
	if !b.has || b.picked == "" || !ok {
		return
	}
	d := detailInfo{ok: true, bs: bs}
	d.sellWhy = b.st.SellRefusal(b.picked)
	d.sellOff = d.sellWhy != ""
	d.sellCmd = sellCommand(b.picked, b.buy, bs.Count)
	d.sellAsk = sellAsk(bs.Name, sellCount(b.buy, bs.Count))
	d.upCmd = upgradeCommand(b.picked, b.buy)
	d.upOff = true
	if b.offer != nil {
		o := b.offer(b.picked)
		d.upWhy = o.Refusal
		d.upOff = o.Refusal != "" || o.ToKey == ""
		if o.ToKey != "" {
			parts := []string{o.ToName}
			for _, res := range sortedKeysOf(o.Cost) {
				if o.Cost[res] > 0 {
					parts = append(parts, game.Amount(o.Cost[res], res))
				}
			}
			d.upLine = strings.Join(parts, " · ")
		}
	}
	if bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
		d.workers = strings.TrimSpace(kit.Strip(workersText(bs)))
		d.minus, d.plus = buildingStep(b.picked, bs, b.st.Workers.TotalIdle)
	}
	b.detail = d
}

// detailNote is the line at the top right of the detail: the reason of the
// button the pointer is on, else what an upgrade makes and costs, else why a
// sale cannot be made.
func (b *buildBox) detailNote(hot xkey, isHot bool) (text string, bad bool) {
	d := b.detail
	if isHot && hot.off {
		switch hot.cmd {
		case d.sellCmd:
			return d.sellWhy, true
		case d.upCmd:
			return d.upWhy, true
		}
	}
	switch {
	case d.upLine != "" && !d.upOff:
		return d.upLine, false
	case d.upOff && d.upWhy != "":
		return d.upWhy, true
	case d.sellOff:
		return d.sellWhy, true
	case d.upLine != "":
		return d.upLine, false
	}
	return "", false
}

// detailWidths is the inner width of the detail for a box w cells wide and
// the width of its text where the buttons stand beside it.
func detailWidths(w int) (iwid, tw int) {
	iwid = w - 6
	sellW, upW := chipW(txtSell.String(), ""), chipW(txtUpgrade.String(), "")
	return iwid, iwid - upW - 1 - sellW - 1
}

// detailWrap breaks s at its spaces into rows, the row at index i of the
// detail's text rows holding at most widthAt(i) cells; from is the index of
// the first row it fills.
func detailWrap(s string, from int, widthAt func(int) int) []string {
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.ReplaceAll(word, glue, " ")
	}
	return wrapLinesAt(words, func(i int) int { return max(widthAt(from+i), 1) }, visibleLen)
}

// detailPlan is how the detail's text lies in its rows: the title on the
// first, then the notice (on the title's row when it fits there), the
// description, the workers and the flavour line. Rows beside the buttons are
// tw cells wide, the rest iwid.
type detailPlan struct {
	iwid, tw int
	beside   bool // the notice shares the title's row
	noteN    int  // rows the notice keeps below the title
	desc     []string
	flav     []string
	workers  bool
}

// widthAt is the cells free on the detail's text row i.
func (p detailPlan) widthAt(i int) int {
	if i >= 1 && i <= chipRows {
		return p.tw
	}
	return p.iwid
}

// noteTexts are the lines the note can show: what an upgrade makes and costs,
// and the reasons of the two buttons.
func (d detailInfo) noteTexts() []string {
	var out []string
	for _, s := range []string{d.upLine, d.upWhy, d.sellWhy} {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// planDetail lays the picked building's detail out for a box w cells wide.
// The note keeps the rows its longest text needs, so that the layout does not
// move when the pointer changes the text.
func (b *buildBox) planDetail(w int) detailPlan {
	d := b.detail
	p := detailPlan{workers: d.workers != ""}
	p.iwid, p.tw = detailWidths(w)
	title := runeLen(d.bs.Name + " x" + strconv.Itoa(d.bs.Count))
	p.beside = true
	for _, t := range d.noteTexts() {
		if title+2+runeLen(t) > p.iwid {
			p.beside = false
		}
	}
	if !p.beside {
		for _, t := range d.noteTexts() {
			p.noteN = max(p.noteN, len(detailWrap(t, 1, p.widthAt)))
		}
	}
	at := 1 + p.noteN
	p.desc = detailWrap(buildingText(d.bs), at, p.widthAt)
	at += len(p.desc)
	if p.workers {
		at++
	}
	if d.bs.Flavor != "" {
		p.flav = detailWrap(d.bs.Flavor, at, p.widthAt)
	}
	return p
}

// rows is the text rows the plan wants.
func (p detailPlan) rows() int {
	n := 1 + p.noteN + len(p.desc) + len(p.flav)
	if p.workers {
		n++
	}
	return n
}

// fit gives way to rows text rows: the flavour line goes first, then the
// description is held to two rows with an ellipsis on the second. The title,
// the notice and the workers stay.
func (p detailPlan) fit(rows int) detailPlan {
	if p.rows() <= rows {
		return p
	}
	p.flav = nil
	if p.rows() <= rows {
		return p
	}
	keep := max(min(rows-(p.rows()-len(p.desc)), 2), 0)
	if keep < len(p.desc) {
		if keep > 0 {
			at := 1 + p.noteN + keep - 1
			p.desc[keep-1] = fitText(strings.Join(p.desc[keep-1:], " "), p.widthAt(at))
		}
		p.desc = p.desc[:keep]
	}
	return p
}

// detailWant is the frame rows the detail would take to say everything.
func (b *buildBox) detailWant(w int) int {
	return max(detailRows, 2+b.planDetail(w).rows())
}

// drawDetail draws the frame, the texts and the keys of the detail at the
// bottom of the box, d rows high, and notes the cells of its note.
func (b *buildBox) drawDetail(g *xgrid, set *keySet, w, h, d int) {
	det := b.detail
	fy := h - 1 - d
	g.box(1, fy, w-2, d, xAccent, "", xTitle, false)
	ix, iy := 3, fy+1
	p := b.planDetail(w).fit(d - 2)
	iwid, tw := p.iwid, p.tw
	sellW, upW := chipW(txtSell.String(), ""), chipW(txtUpgrade.String(), "")
	ux := ix + iwid - upW
	sx := ux - 1 - sellW
	for _, a := range []struct {
		x   int
		txt string
		cmd string
		ask string
		yes string
		off bool
	}{{sx, txtSell.String(), det.sellCmd, det.sellAsk, txtSell.String(), det.sellOff}, {ux, txtUpgrade.String(), det.upCmd, "", "", det.upOff}} {
		set.chip(a.x, iy+1, chipSpec{kind: keySteel, label: a.txt, cmd: a.cmd, ask: a.ask, yes: a.yes, off: a.off}, func(x, y int, text string) { g.put(x, y, text, xKeycap) })
	}
	x := g.put(ix, iy, fitText(det.bs.Name, max(tw-6, 1)), xTitle)
	g.put(x, iy, " x"+strconv.Itoa(det.bs.Count), xDim)
	b.noteRows = b.noteRows[:0]
	if p.beside {
		b.noteRows = append(b.noteRows, image.Rect(ix, iy, ix+iwid, iy+1))
	}
	for i := 1; i <= p.noteN; i++ {
		b.noteRows = append(b.noteRows, image.Rect(ix, iy+i, ix+p.widthAt(i), iy+i+1))
	}
	b.noteBeside = p.beside
	row := 1 + p.noteN
	for _, s := range p.desc {
		g.put(ix, iy+row, s, xText)
		row++
	}
	if p.workers {
		g.stepper(ix, iy+row, det.minus, det.plus)
		g.putFit(ix+8, iy+row, p.widthAt(row)-8, det.workers, xText)
		row++
	}
	for _, s := range p.flav {
		g.put(ix, iy+row, s, xDim)
		row++
	}
}

// stateOf is the look of a key: off, held down, under the pointer, or at rest.
func (b *buildBox) stateOf(key, hot xkey, isHot bool) keyState {
	switch {
	case key.off:
		return keyOff
	case b.k.press != nil && b.k.sunkButton(key):
		return keyPressed
	case b.k.press != nil && isHot && hot.r == key.r, key.picked:
		return keyHover
	}
	return keyRest
}

// paintDetail paints the detail's keys and its note.
func (b *buildBox) paintDetail(p *kit.Painter, pal *expPalette, org image.Point, pix bool) {
	if b.det == nil {
		return
	}
	hot, isHot := b.k.lit(b.det)
	for _, key := range b.det.keys {
		st := b.stateOf(key, hot, isHot)
		b.k.paintKey(p, b.det, org, pal, key, st, pix)
	}
	text, bad := b.detailNote(hot, isHot)
	if text == "" {
		return
	}
	ink := inkOf(pal.tone(xDim, false), false)
	if bad {
		ink = inkOf(pal.tone(xBad, false), false)
	}
	if b.noteBeside {
		r := b.noteRows[0]
		text = fitText(text, max(r.Dx()-runeLen(b.detail.bs.Name+" x"+strconv.Itoa(b.detail.bs.Count))-2, 1))
		p.Label(org.X+r.Max.X-runeLen(text), org.Y+r.Min.Y, kit.Escape(text), ink)
		return
	}
	rows := detailWrap(text, 0, func(i int) int { return b.noteRows[min(i, len(b.noteRows)-1)].Dx() })
	for i := 0; i < len(rows) && i < len(b.noteRows); i++ {
		p.Label(org.X+b.noteRows[i].Min.X, org.Y+b.noteRows[i].Min.Y, kit.Escape(rows[i]), ink)
	}
}
