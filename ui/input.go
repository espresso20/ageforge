package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/ageforge/game"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Message string
	Type    string // "info", "success", "error"
}

// HandleCommand parses and executes a command string
func HandleCommand(input string, engine *game.GameEngine) CommandResult {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return CommandResult{}
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "h", "?":
		return cmdHelp(args)
	case "gather", "g":
		return cmdGather(args, engine)
	case "build", "b":
		return cmdBuild(args, engine)
	case "recruit", "r":
		return cmdRecruit(args, engine)
	case "assign", "a":
		return cmdAssign(args, engine)
	case "unassign", "u":
		return cmdUnassign(args, engine)
	case "status", "s":
		return cmdStatus(engine)
	case "research", "res":
		return cmdResearch(args, engine)
	case "expedition", "exp":
		return cmdExpedition(args, engine)
	case "prestige":
		return cmdPrestige(args, engine)
	case "dump", "exportlogs":
		return cmdDump(args, engine)
	case "save":
		return cmdSave(args, engine)
	case "load":
		return cmdLoad(args, engine)
	default:
		return CommandResult{
			Message: fmt.Sprintf("Unknown command: %s. Type 'help' for commands.", cmd),
			Type:    "error",
		}
	}
}

func cmdHelp(args []string) CommandResult {
	help := `[gold]Commands:[-]
  [cyan]gather[-] <food|wood|stone> [n] - Hand-gather resources (max 5)
  [cyan]build[-] <building>            - Build a structure
  [cyan]recruit[-] <type> [count]      - Recruit villagers (default: 1)
  [cyan]assign[-] <type> <resource> [n]- Assign villagers to gather
  [cyan]unassign[-] <type> <resource> [n]- Unassign villagers
  [cyan]research[-] <tech_key>         - Research a technology
  [cyan]research[-] cancel             - Cancel current research
  [cyan]research[-] list               - List available techs
  [cyan]expedition[-] <key>            - Launch a military expedition
  [cyan]expedition[-] list             - List available expeditions
  [cyan]prestige[-]                    - View prestige status
  [cyan]prestige[-] confirm            - Reset game with prestige bonus
  [cyan]prestige[-] shop               - View prestige upgrades
  [cyan]prestige[-] buy <key>          - Buy a prestige upgrade
  [cyan]status[-]                      - Show detailed status
  [cyan]dump[-]                        - Export logs to file for debugging
  [cyan]save[-] [name]                 - Save game (default: autosave)
  [cyan]load[-] [name]                 - Load game (default: autosave)
  [cyan]help[-]                        - Show this help

[gold]Shortcuts:[-] g=gather, b=build, r=recruit, a=assign, u=unassign, s=status, res=research, exp=expedition`
	return CommandResult{Message: help, Type: "info"}
}

func cmdDump(args []string, engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	logs := engine.GetLogs()

	// Create data/logs directory
	if err := os.MkdirAll("data/logs", 0755); err != nil {
		return CommandResult{Message: fmt.Sprintf("Failed to create logs directory: %v", err), Type: "error"}
	}

	// Generate timestamped filename
	ts := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("data/logs/dump_%s.log", ts)

	var sb strings.Builder

	// Header with engine state
	sb.WriteString("=== AgeForge Log Dump ===\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Tick: %d\n", state.Tick))
	sb.WriteString(fmt.Sprintf("Age: %s (%s)\n", state.AgeName, state.Age))
	sb.WriteString(fmt.Sprintf("Population: %d/%d (idle: %d, food drain: %.2f/tick)\n",
		state.Villagers.TotalPop, state.Villagers.MaxPop, state.Villagers.TotalIdle, state.Villagers.FoodDrain))
	sb.WriteString("\n--- Resources ---\n")
	for _, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %-12s %8.1f / %8.0f  rate: %+.3f/tick\n",
			rs.Name, rs.Amount, rs.Storage, rs.Rate))
	}
	sb.WriteString("\n--- Build Queue ---\n")
	if len(state.BuildQueue) == 0 {
		sb.WriteString("  (empty)\n")
	}
	for _, bq := range state.BuildQueue {
		sb.WriteString(fmt.Sprintf("  %s: %d/%d ticks\n", bq.Name, bq.TotalTicks-bq.TicksLeft, bq.TotalTicks))
	}
	sb.WriteString("\n--- Active Events ---\n")
	if len(state.ActiveEvents) == 0 {
		sb.WriteString("  (none)\n")
	}
	for _, evt := range state.ActiveEvents {
		sb.WriteString(fmt.Sprintf("  %s: %d ticks left\n", evt.Name, evt.TicksLeft))
	}
	if state.Research.CurrentTech != "" {
		sb.WriteString(fmt.Sprintf("\n--- Research ---\n  %s: %d/%d ticks\n",
			state.Research.CurrentTechName,
			state.Research.TotalTicks-state.Research.TicksLeft,
			state.Research.TotalTicks))
	}

	// All log entries
	sb.WriteString(fmt.Sprintf("\n=== Log Entries (%d) ===\n", len(logs)))
	for _, entry := range logs {
		sb.WriteString(fmt.Sprintf("T%-5d [%-7s] %s\n", entry.Tick, entry.Type, entry.Message))
	}

	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return CommandResult{Message: fmt.Sprintf("Failed to write dump: %v", err), Type: "error"}
	}

	return CommandResult{
		Message: fmt.Sprintf("Logs exported to %s (%d entries)", filename, len(logs)),
		Type:    "success",
	}
}

func cmdGather(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: gather <food|wood|stone> [amount] (max 5)", Type: "error"}
	}
	resource := strings.ToLower(args[0])
	if resource != "food" && resource != "wood" && resource != "stone" {
		return CommandResult{Message: "You can only hand-gather food, wood, or stone.", Type: "error"}
	}
	amount := 3.0
	if len(args) >= 2 {
		if n, err := strconv.ParseFloat(args[1], 64); err == nil && n > 0 {
			amount = n
		}
	}
	if amount > 5 {
		amount = 5
	}
	actual, err := engine.GatherResource(resource, amount)
	if err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Gathered %.0f %s (total: %.0f)", amount, resource, actual),
		Type:    "success",
	}
}

func cmdBuild(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		// Show available buildings
		state := engine.GetState()
		var lines []string
		lines = append(lines, "[gold]Available buildings:[-]")
		for key, b := range state.Buildings {
			if !b.Unlocked {
				continue
			}
			affordable := ""
			if b.CanBuild {
				affordable = "[green]✓[-]"
			} else {
				affordable = "[red]✗[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] (%d built) - Cost: %s %s",
				affordable, key, b.Count, FormatCost(b.NextCost), b.Description))
		}
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}
	key := strings.ToLower(args[0])
	if err := engine.BuildBuilding(key); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Built %s!", key),
		Type:    "success",
	}
}

func cmdRecruit(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: recruit <worker|scholar> [count]", Type: "error"}
	}
	vType := strings.ToLower(args[0])
	count := 1
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}
	if err := engine.RecruitVillager(vType, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Recruited %d %s(s)!", count, vType),
		Type:    "success",
	}
}

func cmdAssign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 2 {
		return CommandResult{Message: "Usage: assign <type> <resource> [count]", Type: "error"}
	}
	vType := strings.ToLower(args[0])
	resource := strings.ToLower(args[1])
	count := 1
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
			count = n
		}
	}
	if err := engine.AssignVillager(vType, resource, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Assigned %d %s(s) to %s", count, vType, resource),
		Type:    "success",
	}
}

func cmdUnassign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 2 {
		return CommandResult{Message: "Usage: unassign <type> <resource> [count]", Type: "error"}
	}
	vType := strings.ToLower(args[0])
	resource := strings.ToLower(args[1])
	count := 1
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
			count = n
		}
	}
	if err := engine.UnassignVillager(vType, resource, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Unassigned %d %s(s) from %s", count, vType, resource),
		Type:    "success",
	}
}

func cmdStatus(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string

	lines = append(lines, fmt.Sprintf("[gold]Age:[-] %s  [gold]Tick:[-] %d", state.AgeName, state.Tick))
	lines = append(lines, "")

	// Resources
	lines = append(lines, "[gold]Resources:[-]")
	for _, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		bar := ProgressBar(rs.Amount, rs.Storage, 15)
		lines = append(lines, fmt.Sprintf("  %-10s %s/%s %s %s",
			rs.Name, FormatNumber(rs.Amount), FormatNumber(rs.Storage), FormatRate(rs.Rate), bar))
	}
	lines = append(lines, "")

	// Population
	v := state.Villagers
	lines = append(lines, fmt.Sprintf("[gold]Population:[-] %d/%d (idle: %d, food drain: %.1f/tick)",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain))
	for _, vt := range v.Types {
		if !vt.Unlocked {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %-10s %d (idle: %d)", vt.Name, vt.Count, vt.IdleCount))
		for res, count := range vt.Assignments {
			if count > 0 {
				lines = append(lines, fmt.Sprintf("    → %s: %d", res, count))
			}
		}
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdSave(args []string, engine *game.GameEngine) CommandResult {
	name := "autosave"
	if len(args) > 0 {
		name = args[0]
	}
	if err := engine.SaveGame(name); err != nil {
		return CommandResult{Message: fmt.Sprintf("Save failed: %v", err), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Game saved as '%s'", name), Type: "success"}
}

func cmdLoad(args []string, engine *game.GameEngine) CommandResult {
	name := "autosave"
	if len(args) > 0 {
		name = args[0]
	}
	if err := engine.LoadGame(name); err != nil {
		return CommandResult{Message: fmt.Sprintf("Load failed: %v", err), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Game loaded from '%s'", name), Type: "success"}
}

func cmdResearch(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return cmdResearchList(engine)
	}
	subcmd := strings.ToLower(args[0])

	if subcmd == "list" {
		return cmdResearchList(engine)
	}
	if subcmd == "cancel" {
		if err := engine.CancelResearch(); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: "Research cancelled.", Type: "warning"}
	}

	// Try to start research
	// Support multi-word keys by joining with underscore
	techKey := strings.Join(args, "_")
	if err := engine.StartResearch(techKey); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Started researching %s!", techKey),
		Type:    "success",
	}
}

func cmdResearchList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Technologies:[-]")

	for key, ts := range state.Research.Techs {
		if !ts.Available {
			continue
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-] - %s (%.0f knowledge)", key, ts.Name, ts.Cost))
	}

	if state.Research.CurrentTech != "" {
		lines = append(lines, fmt.Sprintf("\n[yellow]Currently researching: %s (%d ticks left)[-]",
			state.Research.CurrentTechName, state.Research.TicksLeft))
	}

	if len(lines) == 1 {
		lines = append(lines, "  [gray]No technologies available to research[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdExpedition(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return cmdExpeditionList(engine)
	}
	subcmd := strings.ToLower(args[0])
	if subcmd == "list" {
		return cmdExpeditionList(engine)
	}

	// Launch expedition
	expKey := strings.Join(args, "_")
	if err := engine.LaunchExpedition(expKey); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Expedition launched: %s!", expKey),
		Type:    "success",
	}
}

func cmdPrestige(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdPrestigeStatus(engine)
	}
	subcmd := strings.ToLower(args[0])

	switch subcmd {
	case "confirm":
		if err := engine.DoPrestige(); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: "Prestige complete! Your empire has been reset with permanent bonuses.",
			Type:    "success",
		}
	case "shop":
		return cmdPrestigeShop(engine)
	case "buy":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: prestige buy <upgrade_key>", Type: "error"}
		}
		key := strings.Join(args[1:], "_")
		if err := engine.BuyPrestigeUpgrade(key); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Purchased prestige upgrade: %s!", key),
			Type:    "success",
		}
	default:
		return CommandResult{Message: "Usage: prestige [confirm|shop|buy <key>]", Type: "error"}
	}
}

func cmdPrestigeStatus(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	p := state.Prestige
	var lines []string

	lines = append(lines, "[gold]Prestige Status[-]")
	lines = append(lines, fmt.Sprintf("  Level: [cyan]%d[-]", p.Level))
	lines = append(lines, fmt.Sprintf("  Points: [cyan]%d[-] available / [cyan]%d[-] total earned", p.Available, p.TotalEarned))

	if p.PassiveBonus > 0 {
		lines = append(lines, fmt.Sprintf("  Passive Bonus: [green]+%.0f%%[-] production", p.PassiveBonus*100))
	}

	if p.CanPrestige {
		lines = append(lines, fmt.Sprintf("\n  [green]You can prestige now for %d points![-]", p.PendingPoints))
		lines = append(lines, "  Type [cyan]prestige confirm[-] to reset with bonuses.")
	} else {
		lines = append(lines, fmt.Sprintf("\n  [yellow]Reach Medieval Age to prestige (would earn %d pts)[-]", p.PendingPoints))
	}

	lines = append(lines, "\n  Type [cyan]prestige shop[-] to view upgrades.")
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdPrestigeShop(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	p := state.Prestige
	var lines []string

	lines = append(lines, fmt.Sprintf("[gold]Prestige Shop[-] (available: [cyan]%d[-] pts)", p.Available))
	lines = append(lines, "")

	for _, key := range []string{
		"gather_boost", "storage_bonus", "research_speed", "military_power",
		"starting_food", "starting_wood", "population_cap", "expedition_loot",
		"tick_speed",
	} {
		u, ok := p.Upgrades[key]
		if !ok {
			continue
		}
		tierStr := fmt.Sprintf("%d/%d", u.Tier, u.MaxTier)
		costStr := "[gray]MAXED[-]"
		if u.NextCost > 0 {
			costStr = fmt.Sprintf("[cyan]%d pts[-]", u.NextCost)
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-] [%s] %s - %s (Next: %s)",
			key, tierStr, u.Name, u.Description, costStr))
	}

	lines = append(lines, "\n  Type [cyan]prestige buy <key>[-] to purchase.")
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdExpeditionList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Expeditions:[-]")

	for _, exp := range state.Military.Expeditions {
		canStr := "[red]✗[-]"
		if exp.CanLaunch {
			canStr = "[green]✓[-]"
		}
		lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] - %s (%d soldiers, %d ticks)",
			canStr, exp.Key, exp.Name, exp.SoldiersNeeded, exp.Duration))
	}

	if state.Military.ActiveExpedition != nil {
		lines = append(lines, fmt.Sprintf("\n[yellow]Active: %s (%d ticks left)[-]",
			state.Military.ActiveExpedition.Name, state.Military.ActiveExpedition.TicksLeft))
	}

	if len(state.Military.Expeditions) == 0 {
		lines = append(lines, "  [gray]No expeditions available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}
