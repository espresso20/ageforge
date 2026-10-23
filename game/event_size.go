package game

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// event_size.go turns an event's sized effects (config/event_size.go) into
// amounts for the town the event happens to, and writes the line that
// states them.

// townIncome is what the town makes of res each tick before Era Mastery
// scales it: its rate with the timed events and what the people eat left
// out, so an event is sized on the town and not on the events already
// running. Never negative. Read from the last rates pass; under the lock.
func (ge *GameEngine) townIncome(res string) float64 {
	if res == config.ResearchTarget {
		// The work all the scholars would do, with no event running.
		return ge.researchIncome
	}
	r := ge.Resources.resources[res]
	if r == nil {
		return 0
	}
	// Prestige, like Era Mastery, is applied after timed event production.
	// Undo both final stages before sizing a new event, so its amount is
	// multiplied exactly once when the rates are next recalculated.
	made := r.Rate
	if k := ge.speedK(); k > 0 && k != 1 {
		made = made / k
	}
	made -= r.Breakdown.PrestigeRate
	made -= r.Breakdown.EventRate
	made -= r.Breakdown.FoodDrain // the drain is negative: this puts it back
	return max(made, 0)
}

// incomeOf is townIncome as the player sees it: scaled by Era Mastery's
// speed on known ground, as the rates on screen are. The prices that are a
// number of minutes of income (the festival, the Black Market, deals) are
// worked out on it. Under the lock.
func (ge *GameEngine) incomeOf(res string) float64 {
	income := ge.townIncome(res)
	if k := ge.speedK(); k > 0 {
		income = float64(income * k)
	}
	return income
}

// eventTown is what an event's size is read from for res in the current
// age: the stock, the town's income and the age's typical income. perTick
// asks for the units a rate is applied in (before Era Mastery scales the
// rates); otherwise they are the units the player sees.
func (ge *GameEngine) eventTown(res string, perTick bool) config.EventTown {
	t := config.EventTown{
		Stock:   ge.Resources.Get(res),
		Income:  ge.townIncome(res),
		Typical: ge.rules.TypicalIncome(res, ge.age),
	}
	if k := ge.speedK(); !perTick && k > 0 {
		t.Income = float64(t.Income * k)
		t.Typical = float64(t.Typical * k)
	}
	return t
}

// sizeEffects turns sized effects into the effects the engine applies:
// EventGain into an instant_resource of the amount, EventLoss into a
// steal_resource of it, EventRate into a production of it per tick. An
// effect on a resource the town has not unlocked is dropped (there is
// nothing to give it or to take), as is one that comes to nothing. Every
// other effect passes through. Under the lock.
func (ge *GameEngine) sizeEffects(effects []config.Effect) []config.Effect {
	out := make([]config.Effect, 0, len(effects))
	for _, eff := range effects {
		if !config.IsSizedEffect(eff.Type) {
			out = append(out, eff)
			continue
		}
		if eff.Target == config.ResearchTarget {
			// The work scholars do is sized on what they do, and an event's
			// change to it is a share of it (research_boost).
			switch eff.Type {
			case config.EventGain:
				if v := config.EventSize(eff, ge.eventTown(eff.Target, false)); v != 0 {
					out = append(out, config.Effect{Type: "instant_resource", Target: eff.Target, Value: v})
				}
			case config.EventRate:
				// A share of the scholars' work, as the text states it: it
				// needs no size, and it holds whenever scholars are at work.
				if eff.Value != 0 {
					out = append(out, config.Effect{Type: "research_boost", Value: eff.Value})
				}
			}
			continue
		}
		if !ge.Resources.IsUnlocked(eff.Target) {
			continue
		}
		var sized config.Effect
		switch eff.Type {
		case config.EventGain:
			sized = config.Effect{Type: "instant_resource", Target: eff.Target, Value: config.EventSize(eff, ge.eventTown(eff.Target, false))}
		case config.EventLoss:
			sized = config.Effect{Type: "steal_resource", Target: eff.Target, Value: config.EventSize(eff, ge.eventTown(eff.Target, false))}
		case config.EventRate:
			sized = config.Effect{Type: "production", Target: eff.Target, Value: config.EventSize(eff, ge.eventTown(eff.Target, true))}
		}
		if sized.Value != 0 {
			out = append(out, sized)
		}
	}
	return out
}

// eventOutcome is what a random event did when it fired.
type eventOutcome struct {
	// gained is what the event gave, by resource.
	gained map[string]float64
	// lost is what left the stores, after a garrison's share.
	lost    map[string]float64
	workers int
	// rates are the event's timed changes, per tick as the player sees
	// them.
	rates []config.Effect
	// research is what the event gave the scholars' work: "12% of Banking".
	research []string
}

// eventLine is a random event's log line: its own sentence, then what it
// did, with the real amounts. duration is how long its rates last.
func eventLine(sentence string, out eventOutcome, duration string) string {
	parts := []string{"[comment]" + strings.TrimSpace(sentence) + "[-]"}
	if lost := lossParts(out.lost, out.workers); len(lost) > 0 {
		parts = append(parts, vibe.Line("game.event_size.eventLine.01", "Lost ")+joinAnd(lost)+".")
	}
	gained := lossParts(out.gained, 0)
	gained = append(gained, out.research...)
	if len(gained) > 0 {
		parts = append(parts, vibe.Line("game.event_size.eventLine.02", "Gained ")+joinAnd(gained)+".")
	}
	if text := effectsText(out.rates); text != "" {
		parts = append(parts, textfmt.Capitalize(text)+vibe.Line("game.event_size.eventLine.03", " for ")+duration+".")
	}
	return strings.Join(parts, " ")
}
