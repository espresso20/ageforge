package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/game"
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
	sellCmd, upCmd    string
	sellOff, upOff    bool
	sellWhy, upWhy    string
	minusCmd, plusCmd string
	minusOff, plusOff bool
	workers           string
	upLine            string // what an upgrade makes and costs, for an older building
}

// sellCommand is the sell command for the amount chosen: one copy, ten, or
// every copy held.
func sellCommand(key string, buy buyChoice, owned int) string {
	switch {
	case buy == buyTen:
		return command(verbSell, key, "10")
	case buy == buyMax && owned > 0:
		return command(verbSell, key, strconv.Itoa(owned))
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
		d.minusCmd = command(verbUnassign, b.picked)
		d.plusCmd = command(verbAssign, b.picked)
		d.minusOff = bs.WorkersAssigned == 0
		d.plusOff = bs.Count == 0 || bs.WorkersAssigned >= bs.WorkerCapacity*bs.Count || b.st.Workers.TotalIdle == 0
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

// drawDetail draws the frame, the texts and the keys of the detail at the
// bottom of the box and returns the cells of its note.
func (b *buildBox) drawDetail(g *xgrid, set *keySet, w, h int) {
	d := b.detail
	fy := h - 1 - detailRows
	g.box(1, fy, w-2, detailRows, xAccent, "", xTitle, false)
	ix, iy, iwid := 3, fy+1, w-6
	sellW, upW := actionW(txtSell.String(), "", signNone), actionW(txtUpgrade.String(), "", signNone)
	ux := ix + iwid - upW
	sx := ux - 1 - sellW
	tw := sx - 1 - ix
	for _, a := range []struct {
		x   int
		txt string
		cmd string
		off bool
	}{{sx, txtSell.String(), d.sellCmd, d.sellOff}, {ux, txtUpgrade.String(), d.upCmd, d.upOff}} {
		g.fill(a.x, iy+1, actionW(a.txt, "", signNone), actionRows)
		set.action(a.x, iy+1, actionSpec{kind: keyPlain, label: a.txt, cmd: a.cmd, off: a.off}, func(x, y int, text string) { g.put(x, y, text, xText) })
	}
	x := g.put(ix, iy, fitText(d.bs.Name, max(tw-6, 1)), xTitle)
	g.put(x, iy, " x"+strconv.Itoa(d.bs.Count), xDim)
	b.note = image.Rect(ix, iy, ix+iwid, iy+1)

	desc := wrapWords(buildingText(d.bs), tw)
	rows := 3
	if d.workers != "" {
		rows--
	}
	if d.bs.Flavor != "" {
		rows--
	}
	nd := max(min(len(desc), rows), 1)
	if len(desc) > nd {
		desc[nd-1] = fitText(strings.Join(desc[nd-1:], " "), tw)
	}
	y := iy + 1
	for i := 0; i < nd && i < len(desc); i++ {
		g.put(ix, y, desc[i], xText)
		y++
	}
	if d.workers != "" && y < iy+4 {
		for i, k := range []struct {
			sign, cmd string
			off       bool
		}{{"-", d.minusCmd, d.minusOff}, {"+", d.plusCmd, d.plusOff}} {
			r := image.Rect(ix+i*4, y, ix+i*4+3, y+1)
			g.put(r.Min.X, y, " "+k.sign+" ", xKeycap)
			set.keys = append(set.keys, xkey{r: r, cap: r, kind: keySteel, cmd: k.cmd, off: k.off})
		}
		g.putFit(ix+8, y, tw-8, d.workers, xText)
		y++
	}
	if d.bs.Flavor != "" && y < iy+4 {
		g.putFit(ix, y, tw, d.bs.Flavor, xDim)
	}
}

// stateOf is the look of a key: off, held down, under the pointer, or at rest.
func (b *buildBox) stateOf(key, hot xkey, isHot bool) keyState {
	switch {
	case key.off:
		return keyOff
	case b.k.press != nil && b.k.sunkButton(key):
		return keyPressed
	case b.k.press != nil && isHot && hot.r == key.r:
		return keyHover
	}
	return keyRest
}

// paintDetail paints the detail's keys and its note.
func (b *buildBox) paintDetail(p *kit.Painter, pal *expPalette, org image.Point, pix, plain bool) {
	if b.det == nil {
		return
	}
	hot, isHot := b.k.lit(b.det)
	for _, key := range b.det.keys {
		st := b.stateOf(key, hot, isHot)
		if key.r.Dy() == actionRows {
			b.k.paintAction(p, b.det, org, pal, key, st, pix)
		} else {
			b.k.paintKey(p, b.det, org, pal, key, st, pix)
		}
	}
	text, bad := b.detailNote(hot, isHot)
	if text == "" {
		return
	}
	ink := inkOf(pal.tone(xDim, false), false)
	if bad {
		ink = inkOf(pal.tone(xBad, false), false)
	}
	room := b.note.Dx() - runeLen(b.detail.bs.Name+" x"+strconv.Itoa(b.detail.bs.Count)) - 2
	text = fitText(text, max(room, 1))
	if plain {
		text = foldToASCII(text)
	}
	p.Label(org.X+b.note.Max.X-runeLen(text), org.Y+b.note.Min.Y, kit.Escape(text), ink)
}
