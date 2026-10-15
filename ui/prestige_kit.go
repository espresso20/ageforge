package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Depth points and carry-overs on the prestige screens.

// prestigePointsLines say what a prestige pays now and what one more age
// adds, for `prestige`.
func prestigePointsLines(state game.GameState) []string {
	p := state.Prestige
	var lines []string
	switch {
	case state.LastPassage.Pending:
		return nil
	case p.CanPrestige:
		lines = append(lines, fmt.Sprintf("\n  [green]You can prestige now for %s.[-]", textfmt.Count(p.PendingPoints, "point", "points")))
	default:
		lines = append(lines, fmt.Sprintf("\n  [yellow]Reach %s to prestige; a prestige there pays %s.[-]",
			ageRef(state, game.PrestigeMinAge), textfmt.Count(state.Ruleset().DepthPoints(game.PrestigeMinAge), "point", "points")))
	}
	if p.NextAge != "" && p.CanPrestige {
		lines = append(lines, fmt.Sprintf("  From %s it would pay %s (%s more).",
			ageRef(state, p.NextAge), textfmt.Count(p.NextAgePoints, "point", "points"), textfmt.Int(p.NextAgePoints-p.PendingPoints)))
	}
	here, _ := state.Ruleset().Age(state.Age)
	run, _ := state.Ruleset().Age(game.PrestigeRunAge)
	if p.CanPrestige && here.Order < run.Order {
		lines = append(lines, "  [gray]This is an early taste: each era's ages pay 3 times what the era before paid, so going deeper pays far more per day.[-]")
	}
	return lines
}

// carryoverStatusLines reports the three automatic carry-overs. They begin
// with the first prestige, so state level is the only fact needed here.
func carryoverStatusLines(state game.GameState) []string {
	if state.Prestige.Level == 0 {
		return []string{"  Plan Template: [gray]not yet[-]", "  Roster: [gray]not yet[-]", "  Old Friends: [gray]not yet[-]"}
	}
	return []string{"  Plan Template: [green]on[-]", "  Roster: [green]on[-]", "  Old Friends: [green]on[-]"}
}

// kitMemoryLines say what the kit remembers from your runs so far, for the
// shop.
func kitMemoryLines(state game.GameState) []string {
	k := state.Prestige.Kit
	if k.PlanItems+k.Factions+k.Shares == 0 {
		return []string{"  [gray]The kit remembers your plan, roster and the civilizations you meet. Nothing is remembered yet: it starts with your first prestige.[-]"}
	}
	lines := []string{"  [gold]What the kit remembers from your runs[-]"}
	if k.PlanItems > 0 {
		lines = append(lines, fmt.Sprintf("  Plan: %s over %s.", textfmt.Count(k.PlanItems, "item", "items"), textfmt.Count(k.PlanAges, "age", "ages")))
	}
	if k.Shares > 0 {
		lines = append(lines, fmt.Sprintf("  Roster: %s.", textfmt.Count(k.Shares, "domain", "domains")))
	}
	if k.Factions > 0 {
		lines = append(lines, fmt.Sprintf("  Civilizations met: %d.", k.Factions))
	}
	return lines
}
