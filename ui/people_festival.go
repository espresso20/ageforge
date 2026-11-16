package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// people_festival.go is the Festival card on the People overview: what a
// festival gives and costs, whether one can be held, and the button that
// holds one. The card reads the snapshot only.

// festivalCmd is the command line that holds a festival, confirmation included.
const festivalCmd = "festival confirm yes"

// festivalEvent is the key of the running festival among the active events.
const festivalEvent = "cultural_festival"

const (
	festivalCardMin = 30 // the narrowest the card is drawn
	festivalCardMax = 44
	festivalRowsMin = 6 // title row, three lines, one row for a chip, bottom edge
)

// The words the card and the header chip say.
var (
	txtFestival     = vibe.E("ui.people_festival.title.01", "Festival")
	txtFestivalHold = vibe.E("ui.people_festival.hold.01", "Hold festival")
	txtFestivalAsk  = vibe.E("ui.people_festival.ask.01", "Hold a festival?")
)

// festivalRows is the fewest rows the card needs: one more when the buff is
// capped and the card has to say so.
func festivalRows(st game.GameState) int {
	if st.Festival.CapNote != "" {
		return festivalRowsMin + 1
	}
	return festivalRowsMin
}

// festivalRunning is the ticks a running festival has left, and whether one is.
func festivalRunning(st game.GameState) (int, bool) {
	for _, e := range st.ActiveEvents {
		if e.Key == festivalEvent {
			return e.TicksLeft, true
		}
	}
	return 0, false
}

// festivalUnlocked reports that festivals can be held.
func festivalUnlocked(st game.GameState) bool {
	_, locked := featureRefusal(st, config.FeatureFestivals)
	return !locked
}

// festivalAsk is the question before a festival is held: it names the cost.
func festivalAsk(st game.GameState) string {
	return vibe.Sprintf("ui.people_festival.ask.02", "%s\n\nIt costs %s culture.", txtFestivalAsk.String(), textfmt.Number(st.Festival.Cost))
}

// festivalWidth is the width the card takes from a row c.w wide, 0 for none.
func (c peopleBody) festivalWidth() int {
	if !festivalUnlocked(c.st) {
		return 0
	}
	w := min(max(c.w/3, festivalCardMin), festivalCardMax)
	if c.w-w < 30 {
		return 0
	}
	return w
}

// festivalCard draws the card in the box at (x, y), w by h.
func (c peopleBody) festivalCard(x, y, w, h int) {
	g, st := c.g, c.st
	f := st.Festival
	g.box(x, y, w, h, xBorder, txtFestival.String(), xTitle, false)
	rx, ri := x+2, w-4
	row := y + 1
	g.putFit(rx, row, ri, vibe.Sprintf("ui.people_festival.card.01", "+%.0f%% production for %s", f.BuffPercent*100, formatTicks(f.BuffTicks, st)), xBright)
	row++
	// How much of the buff would count sits under the claim, before the price.
	if f.CapNote != "" {
		g.putFit(rx, row, ri, f.CapNote, xGold)
		row++
	}
	haveK := xGood
	if f.Culture < f.Cost {
		haveK = xBad
	}
	g.segsFit(rx, row, ri, xseg{vibe.Line("ui.people_festival.card.02", "Cost "), xDim}, xseg{textfmt.Number(f.Cost), xText},
		xseg{vibe.Line("ui.people_festival.card.03", "  have "), xDim}, xseg{textfmt.Number(f.Culture), haveK})
	row++
	left, running := festivalRunning(st)
	switch {
	case running:
		g.segsFit(rx, row, ri, xseg{vibe.Line("ui.people_festival.card.04", "Running "), xGood}, xseg{formatTicks(left, st), xText})
	case !f.Ready:
		g.segsFit(rx, row, ri, xseg{vibe.Line("ui.people_festival.card.05", "Cooling down "), xGold}, xseg{formatTicks(f.CooldownLeft, st), xText})
	default:
		g.putFit(rx, row, ri, vibe.Line("ui.people_festival.card.06", "Ready"), xGood)
	}
	word := txtFestivalHold.String()
	off := running || !f.Ready || f.Culture < f.Cost
	by := row + 1
	if y+h-1-by >= actionRows && actionW(word, "", signNone) <= ri {
		g.actionButton(rx, by, actionSpec{kind: keyPrimary, label: word, cmd: festivalCmd, ask: festivalAsk(st), yes: word, off: off})
		return
	}
	g.chip(rx, by, word, keyPrimary, festivalCmd, festivalAsk(st), word, off)
}

// festivalChip is the header's note on a running festival: the rise and the
// time left. It is empty while none runs.
func festivalChip(st game.GameState) string {
	left, ok := festivalRunning(st)
	if !ok {
		return ""
	}
	return strings.TrimSpace(vibe.Sprintf("ui.people_festival.chip.01", "%s +%.0f%% %s", txtFestival.String(), st.Festival.BuffPercent*100, formatTicks(left, st)))
}
