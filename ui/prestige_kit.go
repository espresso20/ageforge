package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
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
		lines = append(lines, rowid.Sprintf("ui.prestige_kit.prestigePointsLines.01", "\n  [green]You can prestige now for %s.[-]", textfmt.Count(p.PendingPoints, rowid.Line("ui.prestige_kit.prestigePointsLines.02", "point"), rowid.Line("ui.prestige_kit.prestigePointsLines.03", "points"))))
	default:
		lines = append(lines, rowid.Sprintf("ui.prestige_kit.prestigePointsLines.04", "\n  [yellow]Reach %s to prestige; a prestige there pays %s.[-]",
			ageRef(state, game.PrestigeMinAge), textfmt.Count(state.Ruleset().DepthPoints(game.PrestigeMinAge), rowid.Line("ui.prestige_kit.prestigePointsLines.05", "point"), rowid.Line("ui.prestige_kit.prestigePointsLines.06", "points"))))
	}
	if p.NextAge != "" && p.CanPrestige {
		lines = append(lines, rowid.Sprintf("ui.prestige_kit.prestigePointsLines.07", "  From %s it would pay %s (%s more).",
			ageRef(state, p.NextAge), textfmt.Count(p.NextAgePoints, rowid.Line("ui.prestige_kit.prestigePointsLines.08", "point"), rowid.Line("ui.prestige_kit.prestigePointsLines.09", "points")), textfmt.Int(p.NextAgePoints-p.PendingPoints)))
	}
	here, _ := state.Ruleset().Age(state.Age)
	run, _ := state.Ruleset().Age(game.PrestigeRunAge)
	if p.CanPrestige && here.Order < run.Order {
		lines = append(lines, rowid.Line("ui.prestige_kit.prestigePointsLines.10", "  [gray]This is an early taste: each era's ages pay 3 times what the era before paid, so going deeper pays far more per day.[-]"))
	}
	return lines
}

// carryoverStatusLines reports the three automatic carry-overs. They begin
// with the first prestige, so state level is the only fact needed here.
func carryoverStatusLines(state game.GameState) []string {
	if state.Prestige.Level == 0 {
		return []string{rowid.Line("ui.prestige_kit.carryoverStatusLines.01", "  Plan Template: [gray]not yet[-]"), rowid.Line("ui.prestige_kit.carryoverStatusLines.02", "  Roster: [gray]not yet[-]"), rowid.Line("ui.prestige_kit.carryoverStatusLines.03", "  Old Friends: [gray]not yet[-]")}
	}
	return []string{rowid.Line("ui.prestige_kit.carryoverStatusLines.04", "  Plan Template: [green]on[-]"), rowid.Line("ui.prestige_kit.carryoverStatusLines.05", "  Roster: [green]on[-]"), rowid.Line("ui.prestige_kit.carryoverStatusLines.06", "  Old Friends: [green]on[-]")}
}
