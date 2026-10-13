package ui

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// scholars_text.go is how the screens say research: a bar that fills to 100
// percent, the time it takes the scholars researching now, and the scholars
// themselves. Every figure is read off the state; the words are the engine's
// (game/scholars.go).

// needsScholars is what the engine says of a tech that no scholar is
// researching (game.startedResearchLine); needsScholarsShort is the same
// words for a list.
const (
	needsScholars      = "It needs scholars researching to go on."
	needsScholarsShort = "needs scholars researching"
)

// noScholars is what the engine says of a town without scholars
// (game.SetResearchShare), without "You have".
const noScholars = "no scholars yet: knowledge buildings come with them"

// researchFraction is how far the tech in progress has come, 0 to 1: its
// progress over its price, so it reads right with no scholar researching.
func researchFraction(state game.GameState) float64 {
	r := state.Research
	if r.CurrentTech == "" || r.Price <= 0 {
		return 0
	}
	return math.Min(math.Max(r.Progress/r.Price, 0), 1)
}

// researchPercent is researchFraction as a whole percent. It reads 100 only
// at the price.
func researchPercent(state game.GameState) int {
	f := researchFraction(state)
	p := int(math.Floor(f * 100))
	if f >= 1 {
		p = 100
	}
	return p
}

// researchBar is the bar of the tech in progress, drawn the way a build in
// progress is: width cells.
func researchBar(state game.GameState, width int) string {
	return ProgressBar(researchFraction(state), 1, width)
}

// researchLeft is what is left of the tech in progress: the time at the
// work the researching scholars do now, or that it needs scholars.
func researchLeft(state game.GameState) string {
	if state.Research.NeedsScholars || state.Research.TicksLeft <= 0 {
		return needsScholarsShort
	}
	return formatTicks(state.Research.TicksLeft, state) + " left"
}

// techTime is the time a tech not started takes the scholars researching
// now, or that it needs scholars.
func techTime(ts game.TechState, state game.GameState) string {
	if ts.TicksNeeded <= 0 {
		return needsScholarsShort
	}
	return formatTicks(ts.TicksNeeded, state)
}

// researchingLine is the tech in progress on one line: its name, the bar,
// the percent and what is left.
func researchingLine(state game.GameState, barW int) string {
	return fmt.Sprintf("%s %s %d%% (%s)", state.Research.CurrentTechName, researchBar(state, barW), researchPercent(state), researchLeft(state))
}

// scholarsLine is the scholars on one line: how many research, how many
// teach, and the bonus teaching gives workers now.
func scholarsLine(state game.GameState) string {
	s := state.Scholars
	if s.Total <= 0 {
		return "You have " + noScholars + "."
	}
	return fmt.Sprintf("%d research, %d teach. Teaching adds %s to what workers make.", s.Researching, s.Teaching, textfmt.SignedPercent(s.TeachingBonus))
}

// scholarsShort is scholarsLine in the fewest words, for a title bar.
func scholarsShort(state game.GameState) string {
	s := state.Scholars
	return fmt.Sprintf("Scholars %d research, %d teach, %s", s.Researching, s.Teaching, textfmt.SignedPercent(s.TeachingBonus))
}

// queueLines are the techs queued after the one in progress, in order, and
// the head start held for the next one. nil when there is neither.
func queueLines(state game.GameState) []string {
	r := state.Research
	var out []string
	for i, key := range r.Queue {
		name := key
		if ts, ok := r.Techs[key]; ok {
			name = ts.Name
		}
		out = append(out, fmt.Sprintf("%d. %s", i+1, name))
	}
	if r.HeadStart > 0 {
		out = append(out, headStartLine(state))
	}
	return out
}

// headStartLine says the head start held: a share of the next tech queued,
// else the time of work kept at the scholars' pace now.
func headStartLine(state game.GameState) string {
	r := state.Research
	for _, key := range r.Queue {
		if ts, ok := r.Techs[key]; ok && ts.Cost > 0 {
			share := math.Min(r.HeadStart/ts.Cost, 1) * 100
			if share < 1 {
				return fmt.Sprintf("Head start: under 1%% of %s", ts.Name)
			}
			return fmt.Sprintf("Head start: %d%% of %s", int(share), ts.Name)
		}
	}
	return "Head start: kept for the next tech"
}
