package ui

import (
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

// workBand paints the band at the top of the tab's body.
func (c peopleBody) workBand(rows []game.ShareRow) {
	s, g := c.s, c.g
	h := workBandH(g.h)
	x, w := c.x+1, c.w-2
	s.rect(x-1, c.y, w+2, h, s.pal.Tile)
	slots, filled := 0, 0
	for _, r := range rows {
		slots += r.Slots
		filled += r.Workers
	}
	scale := 3
	if h > 4 {
		scale = 5
	}
	nx := s.num(x+1, c.y+1, num(filled), s.pal.Text, true, scale)
	s.tx(x+1, c.y+1+s.numH(true, scale), rowid.Sprintf("ui.people_work.band.01", "of %s jobs", num(slots)), s.pal.Dim, false)
	if open := slots - filled; open > 0 {
		s.chipAt(nx, c.y+1, num(open)+" "+txtJobsOpen.String(), s.pal.Amber)
	}
	// The jobs by area, a gap between areas.
	chipW := runeLen(txtAllAuto.String()) + 2
	room := x + w - 1 - (nx + 14) - chipW - 2
	total := 0
	for _, r := range rows {
		total += peopledash.BlockCount(r.Slots) + 1
	}
	if total > 0 && room > 4 {
		pitch := min(14, max(2, room*s.cw/total))
		at := 0
		for _, r := range rows {
			n := peopledash.BlockCount(r.Slots)
			s.cv.Blocks(s.px(nx+14)+at*pitch, s.py(c.y+1)+2, 2*s.ch-4, n, peopledash.BlockCount(r.Workers), s.pal.Work, s.pal.Tile, max(1, pitch-2), pitch)
			at += n + 1
		}
	}
	if len(rows) > 0 {
		g.chip(x+w-chipW-1, c.y+1, txtAllAuto.String(), keySteel, command(verbRoster, "auto"), "", "", false)
	}
	g.putFit(nx+14, c.y+h-2, x+w-chipW-(nx+14)-2, recruitText(c.st), xDim)
}
