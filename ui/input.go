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
	case "trade", "t":
		return cmdTrade(args, engine)
	case "diplomacy", "dip":
		return cmdDiplomacy(args, engine)
	case "prestige":
		return cmdPrestige(args, engine)
	case "rates":
		return cmdRates(engine)
	case "speed":
		return cmdSpeed(args, engine)
	case "upgrade":
		return cmdUpgrade(args, engine)
	case "dump", "exportlogs":
		return cmdDump(args, engine)
	case "saves":
		return cmdSaveList()
	case "save":
		if len(args) > 0 && args[0] == "list" {
			return cmdSaveList()
		}
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
  [cyan]build[-] <building> [count|max] - Build structure(s) (default: 1)
  [cyan]recruit[-] <type> [count|max]  - Recruit villagers (default: 1)
  [cyan]assign[-] <type> <resource> [n|all]- Assign villagers to gather
  [cyan]unassign[-] <type> <resource> [n|all]- Unassign villagers
  [cyan]research[-] <tech_key>         - Research a technology
  [cyan]research[-] cancel             - Cancel current research
  [cyan]research[-] list               - List available techs
  [cyan]expedition[-] <key>            - Launch a military expedition
  [cyan]expedition[-] list             - List available expeditions
  [cyan]trade[-] <from> <to> <amount>  - Exchange resources
  [cyan]trade[-] list                  - Show exchange rates
  [cyan]trade[-] route list            - List trade routes
  [cyan]trade[-] route start <key>     - Start a trade route
  [cyan]trade[-] route stop <key>      - Stop a trade route
  [cyan]diplomacy[-]                   - Show faction status
  [cyan]diplomacy[-] ally <faction>    - Ally with faction (costs gold)
  [cyan]diplomacy[-] rival <faction>   - Declare rivalry
  [cyan]diplomacy[-] embargo <faction> - Embargo faction
  [cyan]diplomacy[-] gift <faction>    - Send gift (+15 opinion)
  [cyan]diplomacy[-] neutral <faction> - Reset to neutral
  [cyan]prestige[-]                    - View prestige status
  [cyan]prestige[-] confirm yes        - Reset game with prestige bonus
  [cyan]prestige[-] shop               - View prestige upgrades
  [cyan]prestige[-] buy <key>          - Buy a prestige upgrade
  [cyan]rates[-]                       - Show resource rate breakdown
  [cyan]status[-]                      - Show detailed status
  [cyan]upgrade[-]                     - List available building upgrades
  [cyan]upgrade[-] <building>          - Upgrade all of that building type
  [cyan]upgrade[-] all                 - Upgrade everything affordable
  [cyan]dump[-]                        - Export logs to file for debugging
  [cyan]save[-] [name]                 - Save game (default: autosave)
  [cyan]load[-] [name]                 - Load game (default: autosave)
  [cyan]saves[-]                       - List all save files
  [cyan]speed[-] [1.0|1.5|2.0|...]     - Set game speed (unlocks per wonder built)
  [cyan]help[-]                        - Show this help

[gold]Shortcuts:[-] g=gather, b=build, r=recruit, a=assign, u=unassign, s=status, res=research, exp=expedition, t=trade, dip=diplomacy`
	return CommandResult{Message: help, Type: "info"}
}

func cmdUpgrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		// List available upgrades
		upgrades := engine.GetAvailableUpgrades()
		if len(upgrades) == 0 {
			return CommandResult{Message: "No building upgrades available right now.", Type: "info"}
		}
		var lines []string
		lines = append(lines, "[gold]Available Upgrades (25% of target cost):[-]")
		for _, u := range upgrades {
			affordable := "[red]✗[-]"
			if u.CanAfford {
				affordable = "[green]✓[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] → [cyan]%s[-] (%d available) - Cost: %s",
				affordable, u.FromKey, u.ToKey, u.Count, FormatCost(u.Cost)))
		}
		lines = append(lines, "\n  Type [cyan]upgrade <building>[-] or [cyan]upgrade all[-]")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	key := strings.ToLower(args[0])
	if key == "all" {
		result, err := engine.UpgradeAll()
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		total := 0
		for _, n := range result {
			total += n
		}
		return CommandResult{
			Message: fmt.Sprintf("Upgraded %d buildings total!", total),
			Type:    "success",
		}
	}

	n, err := engine.UpgradeBuilding(key)
	if err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Upgraded %d %s!", n, key),
		Type:    "success",
	}
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
		Type:    "info",
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

	// Check for count or "max"
	if len(args) >= 2 {
		countArg := strings.ToLower(args[1])
		count := 0
		if countArg == "max" {
			count = 10000 // BuildMultiple will stop when resources run out or max is hit
		} else if n, err := strconv.Atoi(countArg); err == nil && n > 0 {
			count = n
		}
		if count > 0 {
			built, err := engine.BuildMultiple(key, count)
			if err != nil {
				return CommandResult{Message: err.Error(), Type: "error"}
			}
			return CommandResult{
				Message: fmt.Sprintf("Built %d %s!", built, key),
				Type:    "success",
			}
		}
	}

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
		return CommandResult{Message: "Usage: recruit <worker|scholar> [count|max]", Type: "error"}
	}
	vType := strings.ToLower(args[0])

	// Check for "max"
	if len(args) >= 2 && strings.ToLower(args[1]) == "max" {
		recruited, err := engine.RecruitMax(vType)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Recruited %d %s(s)!", recruited, vType),
			Type:    "success",
		}
	}

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
		return CommandResult{Message: "Usage: assign <type> <resource> [count|all]", Type: "error"}
	}
	vType := strings.ToLower(args[0])
	resource := strings.ToLower(args[1])
	if len(args) >= 3 && strings.ToLower(args[2]) == "all" {
		n, err := engine.AssignAll(vType, resource)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Assigned all %d %s(s) to %s", n, vType, resource),
			Type:    "success",
		}
	}
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
		return CommandResult{Message: "Usage: unassign <type> <resource> [count|all]", Type: "error"}
	}
	vType := strings.ToLower(args[0])
	resource := strings.ToLower(args[1])
	if len(args) >= 3 && strings.ToLower(args[2]) == "all" {
		n, err := engine.UnassignAll(vType, resource)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Unassigned all %d %s(s) from %s", n, vType, resource),
			Type:    "success",
		}
	}
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
	return CommandResult{Message: fmt.Sprintf("Game saved as '%s'", name), Type: "info"}
}

func cmdLoad(args []string, engine *game.GameEngine) CommandResult {
	name := "autosave"
	if len(args) > 0 {
		name = args[0]
	}
	if err := engine.LoadGame(name); err != nil {
		return CommandResult{Message: fmt.Sprintf("Load failed: %v", err), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Game loaded from '%s'", name), Type: "info"}
}

func cmdRates(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Resource Rate Breakdown:[-]")

	for _, rs := range state.Resources {
		if !rs.Unlocked || (rs.Rate == 0 && rs.Breakdown == (game.RateBreakdown{})) {
			continue
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-]:  %s/tick", rs.Name, FormatRate(rs.Rate)))
		b := rs.Breakdown
		var parts []string
		if b.BuildingRate != 0 {
			parts = append(parts, fmt.Sprintf("Buildings: %+.2f", b.BuildingRate))
		}
		if b.VillagerRate != 0 {
			parts = append(parts, fmt.Sprintf("Villagers: %+.2f", b.VillagerRate))
		}
		if b.ResearchRate != 0 {
			parts = append(parts, fmt.Sprintf("Research: %+.2f", b.ResearchRate))
		}
		if b.EventRate != 0 {
			parts = append(parts, fmt.Sprintf("Events: %+.2f", b.EventRate))
		}
		if b.TradeRate != 0 {
			parts = append(parts, fmt.Sprintf("Trade: %+.2f", b.TradeRate))
		}
		if b.BonusRate != 0 {
			parts = append(parts, fmt.Sprintf("Bonuses: %+.2f", b.BonusRate))
		}
		if b.FoodDrain != 0 {
			parts = append(parts, fmt.Sprintf("Drain: %+.2f", b.FoodDrain))
		}
		if len(parts) > 0 {
			lines = append(lines, fmt.Sprintf("    %s", strings.Join(parts, "  ")))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, "  [gray]No active resource rates[-]")
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdSpeed(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		mult := engine.GetSpeedMultiplier()
		maxSpeed := engine.GetMaxSpeed()
		return CommandResult{
			Message: fmt.Sprintf("Current speed: [cyan]%.1fx[-] (max: [green]%.1fx[-], +0.5x per wonder built)", mult, maxSpeed),
			Type:    "info",
		}
	}
	n, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return CommandResult{Message: "Usage: speed <1.0|1.5|2.0|...>", Type: "error"}
	}
	if err := engine.SetSpeedMultiplier(n); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Game speed set to %.1fx", n),
		Type:    "success",
	}
}

func cmdSaveList() CommandResult {
	saves, err := game.ListSaveDetails()
	if err != nil {
		return CommandResult{Message: fmt.Sprintf("Failed to list saves: %v", err), Type: "error"}
	}
	if len(saves) == 0 {
		return CommandResult{Message: "No save files found.", Type: "info"}
	}
	var lines []string
	lines = append(lines, "[gold]Save Files:[-]")
	for _, s := range saves {
		age := s.Age
		if age == "" {
			age = "unknown"
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%-15s[-] %s  [gray](%s)[-]",
			s.Name, s.Timestamp.Format("2006-01-02 15:04:05"), age))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
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
		// Require "prestige confirm yes" to actually execute
		if len(args) >= 2 && strings.ToLower(args[1]) == "yes" {
			if err := engine.DoPrestige(); err != nil {
				return CommandResult{Message: err.Error(), Type: "error"}
			}
			return CommandResult{
				Message: "Prestige complete! Your empire has been reset with permanent bonuses.",
				Type:    "success",
			}
		}
		// Show warning
		state := engine.GetState()
		p := state.Prestige
		var lines []string
		lines = append(lines, "[yellow]⚠ PRESTIGE WARNING ⚠[-]")
		lines = append(lines, fmt.Sprintf("  You will earn [cyan]%d[-] prestige points.", p.PendingPoints))
		lines = append(lines, "  [red]ALL progress will be reset:[-] resources, buildings, villagers, research, military.")
		lines = append(lines, "  Only prestige points and upgrades are kept.")
		lines = append(lines, "")
		lines = append(lines, "  Type [cyan]prestige confirm yes[-] to proceed.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
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

func cmdTrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdTradeList(engine)
	}
	subcmd := strings.ToLower(args[0])

	if subcmd == "list" {
		return cmdTradeList(engine)
	}
	if subcmd == "route" {
		return cmdTradeRoute(args[1:], engine)
	}

	// Exchange: trade <from> <to> <amount>
	if len(args) < 3 {
		return CommandResult{Message: "Usage: trade <from> <to> <amount> or trade list / trade route list", Type: "error"}
	}
	from := strings.ToLower(args[0])
	to := strings.ToLower(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil || amount <= 0 {
		return CommandResult{Message: "Amount must be a positive number", Type: "error"}
	}

	got, err := engine.ExchangeResources(from, to, amount)
	if err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Exchanged %.0f %s → %.1f %s", amount, from, got, to),
		Type:    "success",
	}
}

func cmdTradeList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	trade := state.Trade
	var lines []string
	lines = append(lines, "[gold]Exchange Rates:[-]")

	if len(trade.ExchangeRates) == 0 {
		lines = append(lines, "  [gray]No exchange rates available (build a market first)[-]")
	} else {
		for _, info := range trade.ExchangeRates {
			pressureStr := ""
			if info.Pressure > 0.05 {
				pressureStr = fmt.Sprintf(" [red]↓%.0f%%[-]", info.Pressure*30)
			}
			lines = append(lines, fmt.Sprintf("  [cyan]%s → %s[-]: %.2f%s", info.From, info.To, info.Rate, pressureStr))
		}
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdTradeRoute(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 || strings.ToLower(args[0]) == "list" {
		return cmdTradeRouteList(engine)
	}
	subcmd := strings.ToLower(args[0])

	if len(args) < 2 {
		return CommandResult{Message: "Usage: trade route start|stop <route_key>", Type: "error"}
	}
	routeKey := strings.Join(args[1:], "_")

	switch subcmd {
	case "start":
		if err := engine.StartTradeRoute(routeKey); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Trade route started: %s", routeKey), Type: "success"}
	case "stop":
		if err := engine.StopTradeRoute(routeKey); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Trade route stopped: %s", routeKey), Type: "success"}
	default:
		return CommandResult{Message: "Usage: trade route start|stop <route_key>", Type: "error"}
	}
}

func cmdTradeRouteList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	trade := state.Trade
	var lines []string
	lines = append(lines, "[gold]Trade Routes:[-]")

	if len(trade.ActiveRoutes) > 0 {
		lines = append(lines, "\n[green]Active:[-]")
		for _, route := range trade.ActiveRoutes {
			lines = append(lines, fmt.Sprintf("  [cyan]%s[-] (%s) - %d ticks left, %d cycles done",
				route.Name, route.Key, route.TicksLeft, route.CyclesDone))
		}
	}

	if len(trade.AvailableRoutes) > 0 {
		lines = append(lines, "\n[yellow]Available:[-]")
		for _, route := range trade.AvailableRoutes {
			status := "[red]✗[-]"
			if route.CanStart {
				status = "[green]✓[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] - %s", status, route.Key, route.Name))
			lines = append(lines, fmt.Sprintf("    %s", route.Description))
		}
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		lines = append(lines, "  [gray]No trade routes available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

func cmdDiplomacy(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdDiplomacyStatus(engine)
	}
	subcmd := strings.ToLower(args[0])

	switch subcmd {
	case "ally":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy ally <faction_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SetDiplomaticStatus(factionKey, "allied"); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Allied with %s!", factionKey), Type: "success"}

	case "rival":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy rival <faction_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SetDiplomaticStatus(factionKey, "rival"); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Declared rivalry with %s!", factionKey), Type: "warning"}

	case "embargo":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy embargo <faction_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SetDiplomaticStatus(factionKey, "embargo"); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Embargoed %s!", factionKey), Type: "warning"}

	case "gift":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy gift <faction_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SendGift(factionKey); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Sent gift to %s (+15 opinion)", factionKey), Type: "success"}

	case "neutral":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy neutral <faction_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SetDiplomaticStatus(factionKey, "neutral"); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Reset %s to neutral", factionKey), Type: "info"}

	default:
		return CommandResult{Message: "Usage: diplomacy [ally|rival|embargo|gift|neutral] <faction_key>", Type: "error"}
	}
}

func cmdDiplomacyStatus(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	dip := state.Diplomacy
	var lines []string
	lines = append(lines, "[gold]Faction Status:[-]")

	if len(dip.Factions) == 0 {
		lines = append(lines, "  [gray]No factions discovered yet (reach Colonial Age)[-]")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	for key, f := range dip.Factions {
		if !f.Discovered {
			lines = append(lines, fmt.Sprintf("  [gray]%s [Undiscovered][-]", f.Name))
			continue
		}
		bonusStr := ""
		if f.Status == "allied" && f.TradeBonus > 0 {
			bonusStr = fmt.Sprintf("  [green]+%.0f%% %s[-]", f.TradeBonus*100, f.Specialty)
		}
		lines = append(lines, fmt.Sprintf("  [cyan]%s[-] (%s) [%s]  Opinion: %d%s  Trades: %d",
			f.Name, key, f.Status, f.Opinion, bonusStr, f.TradeCount))
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}
