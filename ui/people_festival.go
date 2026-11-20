package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// people_festival.go is the Festival card on the People overview: what a
// festival gives and costs, whether one can be held, and the button that
// holds one. The card reads the snapshot only.

// festivalCmd is the command line that holds a festival, confirmation included.
const festivalCmd = "festival confirm yes"

// festivalEvent is the key of the running festival among the active events.
const festivalEvent = "cultural_festival"

// The words the card and the header chip say.
var (
	txtFestival     = rowid.E("ui.people_festival.title.01", "Festival")
	txtFestivalHold = rowid.E("ui.people_festival.hold.01", "Hold festival")
	txtFestivalAsk  = rowid.E("ui.people_festival.ask.01", "Hold a festival?")
)

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
	q := rowid.Sprintf("ui.people_festival.ask.02", "%s\n\nIt costs %s culture.", txtFestivalAsk.String(), textfmt.Number(st.Festival.Cost))
	if st.Festival.CapNote != "" {
		q += "\n" + st.Festival.CapNote
	}
	return q
}

// festivalChip is the header's note on a running festival: the rise and the
// time left. It is empty while none runs.
func festivalChip(st game.GameState) string {
	left, ok := festivalRunning(st)
	if !ok {
		return ""
	}
	return strings.TrimSpace(rowid.Sprintf("ui.people_festival.chip.01", "%s +%.0f%% %s", txtFestival.String(), st.Festival.BuffPercent*100, formatTicks(left, st)))
}
