package ui

import (
	"image/color"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/peopledash"
	"github.com/espresso20/ageforge/rowid"
)

// people_work.go is the Work tab's head band: everyone at work as a numeral,
// the jobs by area as blocks, and the one control that sets every area to auto.

var (
	txtAllAuto  = rowid.E("ui.people_work.words.01", "All auto")
	txtJobsOpen = rowid.E("ui.people_work.words.02", "open")
)

// workBandH is the rows the band takes on a panel h rows tall.
func workBandH(h int) int {
	if h >= peopleBigH {
		return 6
	}
	return 4
}

// A bandPart is one group's slots and how many are filled: a run of blocks.
type bandPart struct{ slots, filled int }

// band paints the head band of a tab: big as a numeral with sub under it, an
// optional chip beside it, the parts as runs of blocks in col, a note under
// the blocks and one control at the right.
func (c peopleBody) band(big, sub, chip, note string, col color.RGBA, parts []bandPart, ctl, ctlCmd string) {
	s, g := c.s, c.g
	h := workBandH(g.h)
	x, w := c.x+1, c.w-2
	s.rect(x-1, c.y, w+2, h, s.pal.Tile)
	scale := 3
	if h > 4 {
		scale = 5
	}
	nx := s.num(x+1, c.y+1, big, s.pal.Text, true, scale)
	s.tx(x+1, c.y+1+s.numH(true, scale), sub, s.pal.Dim, false)
	if chip != "" {
		s.chipAt(nx, c.y+1, chip, s.pal.Amber)
	}
	chipW := runeLen(ctl) + 2
	room := x + w - 1 - (nx + 14) - chipW - 2
	total := 0
	for _, r := range parts {
		total += peopledash.BlockCount(r.slots) + 1
	}
	if total > 0 && room > 4 {
		pitch := min(14, max(1, room*s.cw/total))
		at := 0
		for _, r := range parts {
			n := peopledash.BlockCount(r.slots)
			s.cv.Blocks(s.px(nx+14)+at*pitch, s.py(c.y+1)+2, 2*s.ch-4, n, peopledash.BlockCount(r.filled), col, s.pal.Tile, max(1, pitch-2), pitch)
			at += n + 1
		}
	}
	if ctl != "" {
		g.chip(x+w-chipW-1, c.y+1, ctl, keySteel, ctlCmd, "", "", false)
	}
	g.putFit(nx+14, c.y+h-2, x+w-chipW-(nx+14)-2, note, xDim)
}

// workBand is the Work tab's band: everyone at work, the jobs by area.
func (c peopleBody) workBand(rows []game.ShareRow) {
	slots, filled := 0, 0
	parts := make([]bandPart, len(rows))
	for i, r := range rows {
		slots += r.Slots
		filled += r.Workers
		parts[i] = bandPart{r.Slots, r.Workers}
	}
	chip, ctl := "", ""
	if open := slots - filled; open > 0 {
		chip = num(open) + " " + txtJobsOpen.String()
	}
	if len(rows) > 0 {
		ctl = txtAllAuto.String()
	}
	c.band(num(filled), rowid.Sprintf("ui.people_work.band.01", "of %s jobs", num(slots)), chip, recruitText(c.st), c.s.pal.Work, parts, ctl, command(verbRoster, "auto"))
}
