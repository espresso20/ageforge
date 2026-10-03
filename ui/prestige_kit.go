package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Depth points and the legacy kit on the prestige screens (Pacing v2, PR 6):
// what a prestige pays now against one age deeper, and the kit the shop
// sells with what it remembers.

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
			ageRef(state, game.PrestigeMinAge), textfmt.Count(config.DepthPoints(game.PrestigeMinAge), "point", "points")))
	}
	if p.NextAge != "" && p.CanPrestige {
		lines = append(lines, fmt.Sprintf("  From %s it would pay %s (%s more).",
			ageRef(state, p.NextAge), textfmt.Count(p.NextAgePoints, "point", "points"), textfmt.Int(p.NextAgePoints-p.PendingPoints)))
	}
	if order := config.AgeByKey(); p.CanPrestige && order[state.Age].Order < order[game.PrestigeRunAge].Order {
		lines = append(lines, "  [gray]This is an early taste: each era's ages pay 3 times what the era before paid, so going deeper pays far more per day.[-]")
	}
	return lines
}

// kitOwned counts the kit items bought.
func kitOwned(p game.PrestigeState) int {
	n := 0
	for _, key := range config.LegacyKit() {
		if p.Upgrades[key].Tier > 0 {
			n++
		}
	}
	return n
}

// kitStatusLine is the one-line kit summary for `prestige`.
func kitStatusLine(state game.GameState) string {
	p := state.Prestige
	return fmt.Sprintf("  Legacy kit: %d of %d owned.", kitOwned(p), len(config.LegacyKit()))
}

// kitMemoryLines say what the kit remembers from your runs so far, for the
// shop.
func kitMemoryLines(state game.GameState) []string {
	k := state.Prestige.Kit
	if k.PlanItems+k.Factions+k.Shares == 0 {
		return []string{"  [gray]The kit remembers your plan, worker shares and the civilizations you meet. Nothing is remembered yet: it starts with your first prestige.[-]"}
	}
	lines := []string{"  [gold]What the kit remembers from your runs[-]"}
	if k.PlanItems > 0 {
		lines = append(lines, fmt.Sprintf("  Plan: %s over %s.", textfmt.Count(k.PlanItems, "item", "items"), textfmt.Count(k.PlanAges, "age", "ages")))
	}
	if k.Shares > 0 {
		lines = append(lines, fmt.Sprintf("  Worker shares: %s.", textfmt.Count(k.Shares, "domain", "domains")))
	}
	if k.Factions > 0 {
		lines = append(lines, fmt.Sprintf("  Civilizations met: %d.", k.Factions))
	}
	return lines
}

// prestigeShopLines are the shop's items (the retired perks left out) and
// what the kit remembers.
func prestigeShopLines(state game.GameState) []string {
	p := state.Prestige
	lines := []string{
		fmt.Sprintf("[gold]Prestige shop[-] (you have [cyan]%s[-])", textfmt.Count(p.Available, "point", "points")),
		"",
		"  [gold]The legacy kit[-]: automation that carries over every prestige.",
	}
	for _, def := range config.ActivePrestigeUpgrades() {
		u, ok := p.Upgrades[def.Key]
		if !ok {
			continue
		}
		status := "[green]owned[-]"
		if u.Tier < u.MaxTier {
			status = "[cyan]" + textfmt.Count(u.NextCost, "point", "points") + "[-]"
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-] %s (%s): %s.", def.Key, u.Name, status, u.Description))
	}
	lines = append(lines, "")
	lines = append(lines, kitMemoryLines(state)...)
	lines = append(lines, "\n  Type [cyan]prestige buy <item>[-] to buy one.")
	return lines
}

// kitStatsLines are the owned kit items for the stats overlay ("" lines when
// none are owned).
func kitStatsLines(state game.GameState) string {
	p := state.Prestige
	var sb strings.Builder
	for _, key := range config.LegacyKit() {
		if u := p.Upgrades[key]; u.Tier > 0 {
			if sb.Len() == 0 {
				sb.WriteString("\n [gold]Legacy kit:[-]\n")
			}
			fmt.Fprintf(&sb, "  %s [green]owned[-]\n", u.Name)
		}
	}
	return sb.String()
}
