package presence

import (
	"fmt"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// fieldMax is the longest text Discord takes in one field, in characters.
const fieldMax = 128

// Provider returns the function a Runner calls for the current activity. It is nil while
// the player has Discord status off, the main menu's line while no game is in play, and
// the run's age, year, town and epoch while one is. started is when the process began.
// Nothing here runs inside a bus handler: the Runner calls it from its own goroutine.
func Provider(engine *game.GameEngine, started time.Time) func() *Activity {
	return func() *Activity {
		if acct := engine.Account(); acct != nil && !acct.DiscordOn() {
			return nil
		}
		if !engine.Running() {
			return MenuActivity(started)
		}
		return FromState(engine.GetState(), started)
	}
}

// MenuActivity is the line for the main menu.
func MenuActivity(started time.Time) *Activity {
	return base(started, vibe.Line("presence.MenuActivity.01", "In the main menu"), "")
}

// FromState is the line for a run: the age and its year, then the town and its epoch,
// with the prestige level after them when there is room.
func FromState(st game.GameState, started time.Time) *Activity {
	view := st.AdvanceView()
	details := fmt.Sprintf("%s · %s", view.AgeName, view.YearText)
	if view.YearText == "" {
		details = view.AgeName
	}
	epoch := st.EpochName
	if def, ok := st.Ruleset().Era(st.EpochKey); ok && def.Name != "" {
		epoch = def.Name
	}
	town := mapmodel.TownNameOf(st.Seed)
	state := fmt.Sprintf("%s · %s", town, epoch)
	if epoch == "" {
		state = town
	}
	if st.Prestige.Level > 0 {
		with := state + vibe.Sprintf("presence.FromState.03", " · Prestige %d", st.Prestige.Level)
		if len([]rune(with)) <= fieldMax {
			state = with
		}
	}
	return base(started, details, state)
}

func base(started time.Time, details, state string) *Activity {
	return &Activity{
		Details:    clip(details),
		State:      clip(state),
		LargeImage: "icon",
		LargeText:  "AgeForge",
		Start:      started,
	}
}

// clip cuts s to Discord's field length.
func clip(s string) string {
	r := []rune(s)
	if len(r) <= fieldMax {
		return s
	}
	return string(r[:fieldMax-1]) + "…"
}
