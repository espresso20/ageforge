package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Era Mastery in the UI (game/mastery.go): the speed of the age you are in,
// what the next prestige adds, and a table of the ages you know. Every view
// reads k from the snapshot (MasteryState) and names ages only up to the
// record, the deepest age ever entered (the no-spoiler rule).

// masteryTicks is ticks at the current age's speed: what a build or a
// research started now takes.
func masteryTicks(ticks int, state game.GameState) int {
	return game.MasteryTicks(ticks, state.Mastery.K)
}

// researchTicks is what a tech listed at ticks takes to research if started
// now: the research speed pool takes its share off, the techs' own cut of
// research time and Ancient Knowledge multiply what is left, then the
// current age's speed divides the rest (game.ResearchTicks, what the engine
// starts it with).
func researchTicks(ticks int, state game.GameState) int {
	return game.ResearchTicks(ticks, state.Pools["research_speed"].Earned, state.Research.ResearchTime, state.SuccumbResearchFactor, state.Mastery.K)
}

// masteryNowText is the current age's ground: "known ground, 2.4x faster
// (mastery 2)", "known ground, 4x faster (catching up)" or "new ground, 1x".
func masteryNowText(state game.GameState) string {
	m := state.Mastery
	switch {
	case m.K <= 1:
		return "new ground, 1x"
	case m.CatchUp:
		return fmt.Sprintf("known ground, %s faster (catching up to your record)", game.SpeedText(m.K))
	}
	return fmt.Sprintf("known ground, %s faster (mastery %d)", game.SpeedText(m.K), m.Level)
}

// masteryNextText is what a prestige now adds to mastery, or "" when it
// adds nothing.
func masteryNextText(state game.GameState) string {
	g := state.Mastery.NextGains
	switch len(g) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("the %s gains a mastery level", game.AgeName(g[0]))
	}
	return fmt.Sprintf("the %s to the %s gain a mastery level each", game.AgeName(g[0]), game.AgeName(g[len(g)-1]))
}

// masteryStatusLines are the prestige status lines for Era Mastery.
func masteryStatusLines(state game.GameState) []string {
	lines := []string{fmt.Sprintf("  Era Mastery: %s", masteryNowText(state))}
	if next := masteryNextText(state); next != "" {
		lines = append(lines, fmt.Sprintf("  Next prestige: %s.", next))
	}
	return lines
}

// writeMasterySection renders the ages you know, up to your record, with
// each one's mastery and speed.
func writeMasterySection(sb *strings.Builder, state game.GameState) {
	sb.WriteString("[gold]── Era Mastery ──[-]\n\n")
	m := state.Mastery
	fmt.Fprintf(sb, " This age: %s\n", masteryNowText(state))
	if m.Record == "" || (len(m.Ages) == 0 && m.K <= 1) {
		sb.WriteString(" [gray]An age runs faster once a run has completed it: 2x after one, up to 4.2x after ten.[-]\n")
		return
	}
	sb.WriteString("\n")
	for _, a := range state.Ruleset().AgeKeys() {
		k := m.Speeds[a]
		level := m.Ages[a]
		mark := "  "
		if a == state.Age {
			mark = "▸ "
		}
		speed := "1x"
		if k > 1 {
			speed = game.SpeedText(k)
		}
		note := ""
		if k > config.MasteryK(level) {
			note = " [gray](catch-up)[-]"
		}
		fmt.Fprintf(sb, " %s%-18s mastery %2d  %s%s\n", mark, game.AgeName(a), level, speed, note)
		if a == m.Record {
			break
		}
	}
	if next := masteryNextText(state); next != "" {
		fmt.Fprintf(sb, "\n [gray]Next prestige: %s.[-]\n", next)
	}
	fmt.Fprintf(sb, " [gray]Ages %s or more behind your record run at least %s.[-]\n",
		textfmt.Int(config.CatchUpGap), game.SpeedText(config.CatchUpK))
}
