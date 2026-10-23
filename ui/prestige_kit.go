package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
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
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.prestigePointsLines.01", "\n  [green]You can prestige now for %s.[-]", textfmt.Count(p.PendingPoints, vibe.Line("ui.prestige_kit.prestigePointsLines.02", "point"), vibe.Line("ui.prestige_kit.prestigePointsLines.03", "points"))))
	default:
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.prestigePointsLines.04", "\n  [yellow]Reach %s to prestige; a prestige there pays %s.[-]",
			ageRef(state, game.PrestigeMinAge), textfmt.Count(state.Ruleset().DepthPoints(game.PrestigeMinAge), vibe.Line("ui.prestige_kit.prestigePointsLines.05", "point"), vibe.Line("ui.prestige_kit.prestigePointsLines.06", "points"))))
	}
	if p.NextAge != "" && p.CanPrestige {
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.prestigePointsLines.07", "  From %s it would pay %s (%s more).",
			ageRef(state, p.NextAge), textfmt.Count(p.NextAgePoints, vibe.Line("ui.prestige_kit.prestigePointsLines.08", "point"), vibe.Line("ui.prestige_kit.prestigePointsLines.09", "points")), textfmt.Int(p.NextAgePoints-p.PendingPoints)))
	}
	here, _ := state.Ruleset().Age(state.Age)
	run, _ := state.Ruleset().Age(game.PrestigeRunAge)
	if p.CanPrestige && here.Order < run.Order {
		lines = append(lines, vibe.Line("ui.prestige_kit.prestigePointsLines.10", "  [gray]This is an early taste: each era's ages pay 3 times what the era before paid, so going deeper pays far more per day.[-]"))
	}
	return lines
}

// carryoverStatusLines reports the three automatic carry-overs. They begin
// with the first prestige, so state level is the only fact needed here.
func carryoverStatusLines(state game.GameState) []string {
	if state.Prestige.Level == 0 {
		return []string{vibe.Line("ui.prestige_kit.carryoverStatusLines.01", "  Plan Template: [gray]not yet[-]"), vibe.Line("ui.prestige_kit.carryoverStatusLines.02", "  Roster: [gray]not yet[-]"), vibe.Line("ui.prestige_kit.carryoverStatusLines.03", "  Old Friends: [gray]not yet[-]")}
	}
	return []string{vibe.Line("ui.prestige_kit.carryoverStatusLines.04", "  Plan Template: [green]on[-]"), vibe.Line("ui.prestige_kit.carryoverStatusLines.05", "  Roster: [green]on[-]"), vibe.Line("ui.prestige_kit.carryoverStatusLines.06", "  Old Friends: [green]on[-]")}
}

// kitMemoryLines say what the kit remembers from your runs so far, for the
// shop.
func kitMemoryLines(state game.GameState) []string {
	k := state.Prestige.Kit
	if k.PlanItems+k.Factions+k.Shares == 0 {
		return []string{vibe.Line("ui.prestige_kit.kitMemoryLines.01", "  [gray]The kit remembers your plan, roster and the civilizations you meet. Nothing is remembered yet: it starts with your first prestige.[-]")}
	}
	lines := []string{vibe.Line("ui.prestige_kit.kitMemoryLines.02", "  [gold]What the kit remembers from your runs[-]")}
	if k.PlanItems > 0 {
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.kitMemoryLines.03", "  Plan: %s over %s.", textfmt.Count(k.PlanItems, vibe.Line("ui.prestige_kit.kitMemoryLines.04", "item"), vibe.Line("ui.prestige_kit.kitMemoryLines.05", "items")), textfmt.Count(k.PlanAges, vibe.Line("ui.prestige_kit.kitMemoryLines.06", "age"), vibe.Line("ui.prestige_kit.kitMemoryLines.07", "ages"))))
	}
	if k.Shares > 0 {
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.kitMemoryLines.08", "  Roster: %s.", textfmt.Count(k.Shares, vibe.Line("ui.prestige_kit.kitMemoryLines.09", "domain"), vibe.Line("ui.prestige_kit.kitMemoryLines.10", "domains"))))
	}
	if k.Factions > 0 {
		lines = append(lines, vibe.Sprintf("ui.prestige_kit.kitMemoryLines.11", "  Civilizations met: %d.", k.Factions))
	}
	return lines
}
