package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// CommandResult is the return value of HandleCommand. The caller (Dashboard)
// logs Message if it is non-empty and Type != "success" (successes are
// ephemeral and only shown as toast/log entries by the engine itself).
// If OverlayName is set the Dashboard opens that named overlay panel.
type CommandResult struct {
	Message     string
	Type        string // "info", "success", "error", "warning"
	OverlayName string // non-empty → dashboard should open this overlay panel
}

// HandleCommand parses a raw command string and dispatches to the appropriate
// sub-handler. Single-letter shortcuts (g, b, r, a, u, s, t) are normalised
// to their full equivalents before switching. Commands that purely open an
// overlay return an empty Message with OverlayName set.
func HandleCommand(input string, engine *game.GameEngine) CommandResult {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return CommandResult{}
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "h", "?":
		return CommandResult{OverlayName: "help"}
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
	case "dismiss":
		return cmdDismiss(args, engine)
	case "sell":
		return cmdSell(args, engine)
	case "status", "s":
		return cmdStatus(engine)
	case "research", "res":
		if len(args) == 0 {
			return CommandResult{OverlayName: "techs"}
		}
		return cmdResearch(args, engine)
	case "expedition", "exp":
		return cmdExpedition(args, engine)
	case "campaign":
		return cmdCampaign(args, engine)
	case "trade", "t":
		return cmdTrade(args, engine)
	case "diplomacy", "dip":
		if len(args) == 0 {
			return CommandResult{OverlayName: "diplomacy"}
		}
		return cmdDiplomacy(args, engine)
	case "wonder":
		return cmdWonder(args, engine)
	case "prestige":
		return cmdPrestige(args, engine)
	case "festival":
		return cmdFestival(args, engine)
	case "blackmarket", "bm":
		return cmdBlackMarket(args, engine)
	case "rates":
		return cmdRates(engine)
	case "speed":
		return cmdSpeed(args, engine)
	case "upgrade":
		return cmdUpgrade(args, engine)
	case "advance":
		return cmdAdvance(engine)
	case "milestones", "ms":
		return CommandResult{OverlayName: "milestones"}
	case "techs":
		return CommandResult{OverlayName: "techs"}
	case "army":
		return CommandResult{OverlayName: "army"}
	case "stats":
		return CommandResult{OverlayName: "stats"}
	case "wonders":
		return CommandResult{OverlayName: "wonders"}
	case "workers":
		return CommandResult{OverlayName: "workers"}
	case "logs":
		return CommandResult{OverlayName: "logs"}
	case "epoch":
		return CommandResult{OverlayName: "epoch"}
	case "history":
		return CommandResult{OverlayName: "history"}
	case "buildings":
		return CommandResult{OverlayName: "buildings"}
	case "citymap":
		return CommandResult{OverlayName: "citymap"}
	case "map":
		// Alias for the city view — preserves existing muscle memory.
		return CommandResult{OverlayName: "map"}
	case "worldmap":
		return CommandResult{OverlayName: "worldmap"}
	case "catastrophe", "cat":
		return cmdCatastrophe(args, engine)
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
	case "account", "acct":
		return cmdAccount(args, engine)
	case "theme":
		return cmdTheme(args, engine)
	default:
		return CommandResult{
			Message: fmt.Sprintf("Unknown command: %s. Type 'help' for commands.", cmd),
			Type:    "error",
		}
	}
}

func cmdWonder(args []string, engine *game.GameEngine) CommandResult {
	state := engine.GetState()

	// Find the wonder for the current age
	var curWonder *wonderInfo
	for _, w := range getWonderList() {
		if w.ageKey == state.Age {
			wCopy := w
			curWonder = &wCopy
			break
		}
	}
	if curWonder == nil {
		return CommandResult{Message: "No wonder available this age.", Type: "error"}
	}

	bs := state.Buildings[curWonder.key]
	if bs.Count > 0 {
		return CommandResult{
			Message: fmt.Sprintf("[gold]★ %s[-] is already built!", curWonder.name),
			Type:    "info",
		}
	}

	// "wonder collect <resource> <amount|all>"
	if len(args) >= 3 && strings.ToLower(args[0]) == "collect" {
		resource := strings.ToLower(args[1])
		var amount float64
		if strings.ToLower(args[2]) == "all" {
			if rs, ok := state.Resources[resource]; ok {
				amount = rs.Amount
			} else {
				return CommandResult{Message: fmt.Sprintf("Unknown resource: %s", args[1]), Type: "error"}
			}
		} else {
			var err error
			amount, err = strconv.ParseFloat(args[2], 64)
			if err != nil || amount <= 0 {
				return CommandResult{Message: "Usage: wonder collect <resource> <amount|all>", Type: "error"}
			}
		}
		if err := engine.BankWonderResource(curWonder.key, resource, amount); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		newBS := engine.GetState().Buildings[curWonder.key]
		banked := newBS.WonderBank[resource]
		need := curWonder.def.BaseCost[resource]
		msg := fmt.Sprintf("Banked %.0f %s into %s (%s / %s)", amount, resource, curWonder.name, FormatNumber(banked), FormatNumber(need))
		if newBS.WonderBankFull {
			msg += fmt.Sprintf("\n[green]Bank full! Type 'build %s' to begin construction.[-]", curWonder.key)
		}
		return CommandResult{Message: msg, Type: "success"}
	}

	// Default: show bank status
	var sb strings.Builder
	fmt.Fprintf(&sb, "[gold::b]%s[-] — Wonder Bank\n\n", curWonder.name)

	costKeys := make([]string, 0, len(curWonder.def.BaseCost))
	for k := range curWonder.def.BaseCost {
		costKeys = append(costKeys, k)
	}
	sort.Strings(costKeys)

	for _, res := range costKeys {
		need := curWonder.def.BaseCost[res]
		banked := bs.WonderBank[res]
		pct := 0.0
		if need > 0 {
			pct = banked / need * 100
			if pct > 100 {
				pct = 100
			}
		}
		clr := "red"
		if pct >= 100 {
			clr = "green"
		} else if pct > 0 {
			clr = "yellow"
		}
		fmt.Fprintf(&sb, "  [%s]%s: %s / %s (%.0f%%)[-]\n", clr, res, FormatNumber(banked), FormatNumber(need), pct)
	}

	if bs.WonderBankFull {
		fmt.Fprintf(&sb, "\n[green]Bank full! Type 'build %s' to begin construction.[-]", curWonder.key)
	} else {
		fmt.Fprintf(&sb, "\n[gray]Use 'wonder collect <resource> <amount|all>' to bank resources.[-]")
	}
	return CommandResult{Message: sb.String(), Type: "info"}
}

func cmdAdvance(engine *game.GameEngine) CommandResult {
	if err := engine.AdvanceAge(); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	state := engine.GetState()
	return CommandResult{
		Message: fmt.Sprintf("Your civilization enters the [gold]%s[-]!", state.AgeName),
		Type:    "success",
	}
}

func cmdUpgrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		// List available upgrades
		upgrades := engine.GetAvailableUpgrades()
		if len(upgrades) == 0 {
			return CommandResult{Message: "No building upgrades available right now.", Type: "info"}
		}
		var lines []string
		lines = append(lines, "[gold]Available Upgrades (cost delta: new copy cost − 50% old refund):[-]")
		for _, u := range upgrades {
			affordable := "[red]✗[-]"
			if u.CanAfford {
				affordable = "[green]✓[-]"
			}
			lines = append(lines, fmt.Sprintf("  %s [cyan]%s[-] → [cyan]%s[-] (%d available) - Cost: %s",
				affordable, u.FromKey, u.ToKey, u.Count, FormatCost(u.Cost)))
		}
		lines = append(lines, "\n  Type [cyan]upgrade <building> [n|all][-]")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	building := strings.ToLower(args[0])
	all := false
	count := 0
	if len(args) >= 2 {
		if strings.ToLower(args[1]) == "all" {
			all = true
		} else {
			n, err := strconv.Atoi(args[1])
			if err != nil || n <= 0 {
				return CommandResult{Message: "Usage: upgrade <building> [count|all]", Type: "error"}
			}
			count = n
		}
	} else {
		all = true // default: upgrade all
	}

	if err := engine.UpgradeBuilding(building, count, all); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{Message: "", Type: "success"}
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
		state.Workers.TotalPop, state.Workers.MaxPop, state.Workers.TotalIdle, state.Workers.FoodDrain))
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

// gatherMaxYield is the per-use cap on hand-gathered resources.
const gatherMaxYield = 25.0

func cmdGather(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: gather <food|wood|stone> [amount] (max 25)", Type: "error"}
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
	if amount > gatherMaxYield {
		amount = gatherMaxYield
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

	// "build <key> [count|max]" — passing 10000 as the count is a sentinel
	// that tells BuildMultiple "keep building until you can't afford it or
	// hit the MaxCount limit". BuildMultiple returns the actual count built.
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
	if len(args) == 0 {
		if err := engine.RecruitWorker("worker", 1); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: "Recruited 1 worker!", Type: "success"}
	}

	arg := strings.ToLower(args[0])
	if arg == "max" {
		recruited, err := engine.RecruitMax("worker")
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Recruited %d workers!", recruited), Type: "success"}
	}

	n, err := strconv.Atoi(arg)
	if err != nil || n <= 0 {
		return CommandResult{
			Message: "Usage: recruit [count|max] — workers are recruited from available housing capacity and assigned to buildings.",
			Type:    "error",
		}
	}
	if err := engine.RecruitWorker("worker", n); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Recruited %d worker(s)!", n), Type: "success"}
}

func cmdAssign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: assign <building> [count|all]", Type: "error"}
	}
	building := strings.ToLower(args[0])
	if len(args) >= 2 && strings.ToLower(args[1]) == "all" {
		n, err := engine.AssignAll(building)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Assigned all %d workers to %s", n, building),
			Type:    "success",
		}
	}
	count := 1
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}
	if err := engine.AssignWorker(building, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Assigned %d worker(s) to %s", count, building),
		Type:    "success",
	}
}

func cmdUnassign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: unassign <building> [count|all]", Type: "error"}
	}
	building := strings.ToLower(args[0])
	if len(args) >= 2 && strings.ToLower(args[1]) == "all" {
		n, err := engine.UnassignAll(building)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{
			Message: fmt.Sprintf("Unassigned all %d workers from %s", n, building),
			Type:    "success",
		}
	}
	count := 1
	if len(args) >= 2 {
		if n, err := strconv.Atoi(args[1]); err == nil && n > 0 {
			count = n
		}
	}
	if err := engine.UnassignWorker(building, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: fmt.Sprintf("Unassigned %d worker(s) from %s", count, building),
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
	v := state.Workers
	lines = append(lines, fmt.Sprintf("[gold]Population:[-] %d/%d (idle: %d, food drain: %.1f/tick)",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain))
	for _, vt := range v.Types {
		if !vt.Unlocked {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %-10s %d (idle: %d)", vt.Name, vt.Count, vt.IdleCount))
		for building, count := range vt.Assignments {
			if count > 0 {
				lines = append(lines, fmt.Sprintf("    → %s: %d", building, count))
			}
		}
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// shortAccountID returns a human-friendly short form of a 32-char hex account ID:
// the first 8 hex chars (enough to recognize an account at a glance). Falls back to
// the whole string when it's shorter than 8 chars.
func shortAccountID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// cmdAccount handles the `account` command family (accounts.md §9 Phase 4):
//
//	account                 → show the short ID + recovery code + honest "identity,
//	                          not progress" copy.
//	account recover <code>  → re-create the local identity from a recovery code.
//	                          Guarded: if the current account already holds unlocks,
//	                          require `account recover <code> confirm` to overwrite.
//
// It talks to engine.Account()/SetAccount() directly — no GameState snapshot needed.
func cmdAccount(args []string, engine *game.GameEngine) CommandResult {
	acct := engine.Account()

	if len(args) == 0 {
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("[gold]Account:[-] %s  (%s)", acct.Name(), shortAccountID(acct.AccountID)))
		lines = append(lines, fmt.Sprintf("[gold]Recovery code:[-] %s", acct.RecoveryCode()))
		lines = append(lines, "")
		lines = append(lines, "This code restores your IDENTITY (your account ID) across machines and")
		lines = append(lines, "reinstalls — NOT your earned progress. It is not a password: it proves")
		lines = append(lines, "nothing secret, only which account you are. Write it down.")
		lines = append(lines, "")
		lines = append(lines, "Restore identity on another machine with:  account recover <code>")
		lines = append(lines, "")
		lines = append(lines, "[gold]Progress (unlocks, stats, achievements)[-] is backed up SEPARATELY as a")
		lines = append(lines, "per-account file. Export and import both work now and are multi-account —")
		lines = append(lines, "each account is its own slot; an import adds/restores that account alongside")
		lines = append(lines, "your others. Switch between accounts in the [gold]Accounts[-] panel on the main menu.")
		lines = append(lines, "  account list                    → list your local accounts")
		lines = append(lines, "  account switch <name>           → switch to an existing local account")
		lines = append(lines, "  account export [path]           → write this account's progress backup")
		lines = append(lines, "  account backup                  → full snapshot (account.json + saves) to data/backups/")
		lines = append(lines, "  account import <path> [replace] → restore an account from a backup (merges by default)")
		lines = append(lines, "")
		lines = append(lines, "[red]Wipe Account[-] (permanently delete an account's identity + unlocks + stats)")
		lines = append(lines, "lives in the [gold]Accounts[-] panel on the main menu, behind a type-your-name confirm.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}

	sub := strings.ToLower(args[0])
	switch sub {
	case "list":
		// List every local account slot: name, short id, an active marker, and highest age.
		summaries := engine.ListAccounts()
		if len(summaries) == 0 {
			return CommandResult{Message: "No local accounts found.", Type: "info"}
		}
		var lines []string
		lines = append(lines, "[gold]Local accounts:[-]")
		for _, s := range summaries {
			marker := "  "
			if s.Active {
				marker = "[aqua]●[-] "
			}
			name := s.DisplayName
			if strings.TrimSpace(name) == "" {
				name = "(unnamed)"
			}
			age := s.HighestAge
			if age == "" {
				age = "—"
			}
			tampered := ""
			if s.Tampered {
				tampered = "  [red]⚠ modified[-]"
			}
			lines = append(lines, fmt.Sprintf("%s%s  [gray](%s)[-]  [gray]age:[-] %s%s",
				marker, name, shortAccountID(s.AccountID), age, tampered))
		}
		lines = append(lines, "")
		lines = append(lines, "Switch with:  account switch <name>   (or use the Accounts panel on the main menu)")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}

	case "switch":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: account switch <name>", Type: "error"}
		}
		// Resolve name→id via the shared derivation, then switch only if that slot exists.
		// A non-existent slot errors with guidance rather than minting an empty account
		// (use the Accounts panel — or `account import` — to create/restore one).
		name := strings.Join(args[1:], " ")
		id := game.AccountIDForName(name)
		if err := engine.SwitchAccount(id); err != nil {
			return CommandResult{
				Message: fmt.Sprintf("No account named %q — create it from the Accounts panel on the main menu.", strings.TrimSpace(name)),
				Type:    "error",
			}
		}
		// Re-resolve the active theme against the now-active account so the UI doesn't keep
		// the prior account's theme after the swap (theming.md §6).
		applyAccountTheme(engine)
		return CommandResult{
			Message: fmt.Sprintf("Now playing as %s (%s).", strings.TrimSpace(name), shortAccountID(id)),
			Type:    "success",
		}

	case "export":
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		blob, err := acct.ExportProgress()
		if err != nil {
			return CommandResult{Message: fmt.Sprintf("Export failed: %v", err), Type: "error"}
		}
		// Resolve the destination: explicit path arg, or a default that names the account so
		// multiple accounts' exports don't collide in a shared directory. The blob carries the
		// id regardless; the filename is just a convenience. Defaults beside the live account.
		var path string
		if len(args) >= 2 && args[1] != "" {
			path = args[1]
		} else {
			path = filepath.Join(game.DataDir(), fmt.Sprintf("account-%s-export.json", shortAccountID(acct.AccountID)))
		}
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return CommandResult{Message: fmt.Sprintf("Export failed: %v", err), Type: "error"}
			}
		}
		if err := os.WriteFile(path, blob, 0644); err != nil {
			return CommandResult{Message: fmt.Sprintf("Export failed: %v", err), Type: "error"}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("[gold]Progress exported:[-] %s", path))
		lines = append(lines, "This file is your PROGRESS backup (unlocks, stats, achievements). Keep it")
		lines = append(lines, "safe — it is separate from your recovery code, which carries only identity.")
		lines = append(lines, "Restore it with:  account import "+path)
		// Also take a FULL slot snapshot (account.json + saves/). A backup failure must not fail
		// the export — append a soft note instead.
		if backupPath, bErr := engine.BackupAccount(acct.AccountID); bErr == nil {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("[gold]Full backup (account.json + saves):[-] %s", backupPath))
		}
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "success"}

	case "backup":
		// A FULL snapshot of the ACTIVE account's slot (account.json + saves/) into
		// <root>/backups/. Distinct from export, which serializes only meta-progression.
		if acct == nil {
			return CommandResult{Message: "No account to back up.", Type: "warning"}
		}
		backupPath, err := engine.BackupAccount(acct.AccountID)
		if err != nil {
			return CommandResult{Message: fmt.Sprintf("Backup failed: %v", err), Type: "error"}
		}
		var lines []string
		lines = append(lines, fmt.Sprintf("[gold]Full backup saved:[-] %s", backupPath))
		lines = append(lines, "A complete snapshot of this account: account.json plus every save in")
		lines = append(lines, "its slot. Restore by copying the folder's contents back into")
		lines = append(lines, "data/accounts/<id>/. Only the 10 most recent backups per account are kept.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "success"}

	case "import":
		if acct == nil {
			return CommandResult{
				Message: "Accounts are unavailable (no account is loaded).",
				Type:    "warning",
			}
		}
		if len(args) < 2 {
			return CommandResult{
				Message: "Usage: account import <path> [replace]",
				Type:    "error",
			}
		}
		path := args[1]
		// merge by default; the `replace` token switches to wholesale replacement.
		merge := !(len(args) >= 3 && strings.EqualFold(args[2], "replace"))
		blob, err := os.ReadFile(path)
		if err != nil {
			return CommandResult{Message: fmt.Sprintf("Import failed: cannot read %s: %v", path, err), Type: "error"}
		}
		// An export is a single-account backup: it lands in its OWN account's slot (keyed by
		// the blob's account id), not the active account. Restoring a backup means making that
		// account current, so switch to it on success.
		imported, err := engine.ImportAccountExport(blob, merge)
		if err != nil {
			return CommandResult{Message: fmt.Sprintf("Import failed: %v", err), Type: "error"}
		}
		if err := engine.SwitchAccount(imported.AccountID); err != nil {
			return CommandResult{Message: fmt.Sprintf("Imported, but could not switch to the account: %v", err), Type: "error"}
		}
		mode := "merged"
		if !merge {
			mode = "replaced"
		}
		name := imported.DisplayName
		if name == "" {
			name = "(unnamed)"
		}
		themeCount := len(imported.UnlockedThemes())
		return CommandResult{
			Message: fmt.Sprintf("Imported account %q (%s) — now active — %d theme(s) unlocked, progress %s.",
				name, shortAccountID(imported.AccountID), themeCount, mode),
			Type: "success",
		}

	case "recover":
		if len(args) < 2 {
			return CommandResult{
				Message: "Usage: account recover <code>",
				Type:    "error",
			}
		}
		code := args[1]
		confirmed := len(args) >= 3 && strings.EqualFold(args[2], "confirm")

		// Overwrite guard: if the CURRENT account already has earned progress (unlocked
		// themes), recovering would replace the local identity and the code does NOT
		// carry that progress. Require an explicit confirm token before proceeding.
		if acct != nil && len(acct.UnlockedThemes()) > 0 && !confirmed {
			var lines []string
			lines = append(lines, "[red]Warning:[-] this account has unlocked progress on this machine.")
			lines = append(lines, "Recovering will REPLACE the current local identity. The recovery code")
			lines = append(lines, "carries identity only — your unlocks/stats are NOT carried by it and")
			lines = append(lines, "would no longer be attached to this identity. Export your progress first")
			lines = append(lines, "with `account export` (it writes a backup you can import later).")
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("To proceed anyway:  account recover %s confirm", code))
			return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
		}

		restored, err := game.ImportRecoveryCode(code)
		if err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		engine.SetAccount(restored)
		// Re-resolve the active theme against the now-installed account so the UI
		// doesn't keep the prior account's theme after an identity swap (theming.md
		// §6). A recovery code carries identity only, so a fresh restore resolves to
		// Forge; a restore of an account with a stored theme honors it.
		applyAccountTheme(engine)
		return CommandResult{
			Message: fmt.Sprintf("Identity restored: %s", shortAccountID(restored.AccountID)),
			Type:    "success",
		}

	case "wipe":
		// The destructive wipe lives behind the Accounts panel's type-your-name gate — we
		// deliberately do NOT wipe from a bare command. Direct the player there.
		return CommandResult{
			Message: "Wiping an account is permanent and lives in the Accounts panel on the main menu (press Esc to reach it, then 'Accounts', then 'w' on the account). It deletes that account's identity, theme unlocks, lifetime stats, and achievements — game saves are NOT affected.",
			Type:    "warning",
		}

	default:
		return CommandResult{
			Message: "Usage: account  |  account list  |  account switch <name>  |  account recover <code>  |  account export [path]  |  account backup  |  account import <path> [replace]  |  account wipe",
			Type:    "error",
		}
	}
}

func cmdSave(args []string, engine *game.GameEngine) CommandResult {
	// With a name: branch a new save off the current run (BranchSave switches the
	// active slot + sets the parent, so autosave follows the branch). Without a
	// name: overwrite the active slot. The dashboard UI intercepts a bare `save`
	// to pop the Overwrite/Branch modal, but this path stays sane for tests and
	// other callers.
	if len(args) > 0 {
		name := args[0]
		if err := engine.BranchSave(name); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Branched a new save '%s' — autosave now follows it", name), Type: "info"}
	}
	active := engine.ActiveSaveName()
	if err := engine.SaveGame(active); err != nil {
		return CommandResult{Message: fmt.Sprintf("Save failed: %v", err), Type: "error"}
	}
	return CommandResult{Message: fmt.Sprintf("Saved to '%s'", active), Type: "info"}
}

func cmdLoad(args []string, engine *game.GameEngine) CommandResult {
	// No name: do NOT silently load autosave. Guide the player to the browser.
	// The dashboard intercepts a bare `load` to open the Load Game tree; this
	// fallback keeps other callers (and tests) from loading a slot by surprise.
	if len(args) == 0 {
		return CommandResult{Message: "Type 'load <name>' to load a specific save, or open Load Game from the menu (or press Esc) to browse your save tree.", Type: "info"}
	}
	name := args[0]
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
		if b.WorkerRate != 0 {
			parts = append(parts, fmt.Sprintf("Workers: %+.2f", b.WorkerRate))
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

// cmdTheme handles the `theme` command's text paths:
//   - `theme`        → directs the player to the picker (the interactive picker
//     page is opened by the dashboard intercept; this fallback covers the
//     command-table path and tests, which can't open an interactive page).
//   - `theme list`   → lists every theme (name, key, active marker, accessible note).
//   - `theme <key>`  → switches directly to a theme by key if it exists, else an
//     error listing the valid keys.
//
// theme.SetActive applies the name-remap + restyle; the dashboard redraws on its
// next tick, so the switch shows up live without an explicit Draw here.
//
// engine bridges to the account layer (theming.md §6): on a successful switch we
// persist the new active theme account-wide so a CLI switch survives saves, and the
// theme is gated through themeAvailable so locked flavor themes (Phase 3) refuse.
// engine/Account() are nil-guarded — the game must run accountless.
func cmdTheme(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{
			Message: "Usage: theme list | theme <key>. Type `theme` from the menu (or open it) for the live picker.",
			Type:    "info",
		}
	}
	if strings.ToLower(args[0]) == "list" {
		return cmdThemeList(themeAccount(engine))
	}

	key := strings.ToLower(args[0])
	t, ok := theme.ByKey(key)
	if !ok {
		return CommandResult{
			Message: fmt.Sprintf("Unknown theme: %s. Valid: %s", key, strings.Join(themeKeys(), ", ")),
			Type:    "error",
		}
	}
	// Unlock gate (theming.md §4/§5): refuse a known-but-locked theme. No theme is
	// locked today (all shipped themes are Accessible/Forge); this is the Phase-3 seam.
	if !themeAvailable(themeAccount(engine), t) {
		return CommandResult{Message: themeUnavailableMsg, Type: "error"}
	}
	if err := theme.SetActive(t.Key); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	theme.Restyle()
	// Persist account-wide so a CLI switch survives saves (theming.md §6). Nil-guarded;
	// the Save error is non-fatal — the theme is already applied, so we report success
	// regardless rather than rolling back a working visual change over a write hiccup.
	if acct := themeAccount(engine); acct != nil {
		_ = acct.SetActiveTheme(t.Key)
	}
	return CommandResult{Message: fmt.Sprintf("Theme set: %s", t.Name), Type: "success"}
}

// themeAccount returns engine's account, or nil when accountless. Mirrors the
// picker's account() guard so the command + picker share one nil-safe accessor.
func themeAccount(engine *game.GameEngine) *game.Account {
	if engine == nil {
		return nil
	}
	return engine.Account()
}

// cmdThemeList renders the `theme list` output: one line per theme with its name,
// key, an active marker, and a status note. Unlocked themes show "(accessible)" when
// applicable; a LOCKED flavor theme shows "🔒 <unlock hint>" instead, so the player
// sees exactly how to earn it (theming.md §5/§7), e.g. "monochrome  🔒 Reach the
// Information Age".
//
// acct may be nil (accountless play / tests): with no account only the always-
// available set (Accessible + Forge) is unlocked, so the flavor themes correctly
// render as locked with their hints.
func cmdThemeList(acct *game.Account) CommandResult {
	activeKey := theme.Active().Key
	var lines []string
	lines = append(lines, "[gold]Themes:[-]")
	for _, t := range theme.All() {
		marker := "  "
		if t.Key == activeKey {
			marker = "[gold]●[-] "
		}
		note := ""
		switch {
		case !themeAvailable(acct, t):
			// Locked flavor theme: show the unlock condition rather than nothing.
			hint := t.UnlockHint
			if hint == "" {
				hint = "unlock via a milestone"
			}
			note = fmt.Sprintf("  [red]🔒[-] [gray]%s[-]", hint)
		case t.Accessible:
			note = "  [cyan](accessible)[-]"
		}
		lines = append(lines, fmt.Sprintf("%s[white]%-18s[-] [gray]%s[-]%s",
			marker, t.Name, t.Key, note))
	}
	lines = append(lines, "[gray]Use `theme <key>` to switch.[-]")
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// themeKeys returns every registered theme key in display order, for usage/error
// messages and autocomplete.
func themeKeys() []string {
	all := theme.All()
	keys := make([]string, 0, len(all))
	for _, t := range all {
		keys = append(keys, t.Key)
	}
	return keys
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

	// Support multi-word keys entered with spaces by joining all remaining args
	// with underscores (e.g. "research bronze working" → "bronze_working").
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

// cmdExpedition handles the `expedition`/`exp` command — the civilian SCOUTING
// surface. No args opens the Expeditions panel; "list" prints the scouting list;
// a key launches a scouting expedition. Military keys are redirected to the
// `campaign` command rather than launched here.
func cmdExpedition(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{OverlayName: "expedition"}
	}
	subcmd := strings.ToLower(args[0])
	if subcmd == "list" {
		return cmdScoutingList(engine)
	}

	expKey := strings.Join(args, "_")

	// Reject military keys here — they belong to `campaign`.
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil && def.Category != game.ExpeditionScouting {
		return CommandResult{
			Message: fmt.Sprintf("%s is a military campaign — wage it with 'campaign %s'.", def.Name, expKey),
			Type:    "info",
		}
	}

	if err := engine.LaunchExpedition(expKey); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	name := expKey
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil {
		name = def.Name
	}
	return CommandResult{
		Message: fmt.Sprintf("Expedition launched: %s!", name),
		Type:    "success",
	}
}

// cmdCampaign handles the `campaign` command — the MILITARY surface. No args or
// "list" prints the campaign list; a key wages a military campaign (costs
// soldiers). Scouting keys are redirected to the `expedition` command.
func cmdCampaign(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return cmdCampaignList(engine)
	}
	subcmd := strings.ToLower(args[0])
	if subcmd == "list" {
		return cmdCampaignList(engine)
	}

	expKey := strings.Join(args, "_")

	// Reject scouting keys here — they belong to `expedition`.
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil && def.Category != game.ExpeditionMilitary {
		return CommandResult{
			Message: fmt.Sprintf("%s is a scouting expedition — send it with 'expedition %s'.", def.Name, expKey),
			Type:    "info",
		}
	}

	if err := engine.LaunchExpedition(expKey); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	name := expKey
	if def := engine.Military.ExpeditionDefByKey(expKey); def != nil {
		name = def.Name
	}
	return CommandResult{
		Message: fmt.Sprintf("Campaign launched: %s!", name),
		Type:    "success",
	}
}

// cmdFestival handles the `festival` culture-sink command. Bare `festival`
// shows status (cost, cooldown, what it does). `festival confirm yes` spends a
// lump of culture to inject a temporary empire-wide production buff. Mirrors the
// prestige confirm UX so the muscle memory carries over.
func cmdFestival(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdFestivalStatus(engine)
	}
	subcmd := strings.ToLower(args[0])

	switch subcmd {
	case "confirm":
		if len(args) >= 2 && strings.ToLower(args[1]) == "yes" {
			if err := engine.DoFestival(); err != nil {
				return CommandResult{Message: err.Error(), Type: "error"}
			}
			st := engine.FestivalStatus()
			return CommandResult{
				Message: fmt.Sprintf("Festival underway! Spent %.0f culture. +%.0f%% to all production for %d ticks.",
					st.Cost, st.BuffPercent*100, st.BuffTicks),
				Type: "success",
			}
		}
		// Show the confirm prompt with the live cost.
		st := engine.FestivalStatus()
		if !st.Ready {
			return CommandResult{
				Message: fmt.Sprintf("[yellow]Festival on cooldown[-] — %d ticks until the next one can be held.", st.CooldownLeft),
				Type:    "warning",
			}
		}
		var lines []string
		lines = append(lines, "[gold]Hold a Cultural Festival?[-]")
		lines = append(lines, fmt.Sprintf("  Cost: [cyan]%.0f culture[-] (you have %.0f)", st.Cost, st.Culture))
		lines = append(lines, fmt.Sprintf("  Effect: [green]+%.0f%%[-] to all production for [cyan]%d[-] ticks.", st.BuffPercent*100, st.BuffTicks))
		lines = append(lines, fmt.Sprintf("  Cooldown afterward: [cyan]%d[-] ticks.", st.CooldownTicks))
		if st.Culture < st.Cost {
			lines = append(lines, "")
			lines = append(lines, "  [red]Not enough culture.[-]")
		}
		lines = append(lines, "")
		lines = append(lines, "  Type [cyan]festival confirm yes[-] to celebrate.")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "warning"}
	default:
		return CommandResult{Message: "Usage: festival [confirm yes]", Type: "error"}
	}
}

// cmdFestivalStatus renders the bare `festival` status panel.
func cmdFestivalStatus(engine *game.GameEngine) CommandResult {
	st := engine.FestivalStatus()
	var lines []string
	lines = append(lines, "[gold]Cultural Festival[-]")
	lines = append(lines, "  Spend a lump of culture for a temporary empire-wide production boost.")
	lines = append(lines, fmt.Sprintf("  Cost: [cyan]%.0f culture[-]  (you have %.0f)", st.Cost, st.Culture))
	lines = append(lines, fmt.Sprintf("  Effect: [green]+%.0f%%[-] to all production for [cyan]%d[-] ticks.", st.BuffPercent*100, st.BuffTicks))
	if st.Ready {
		if st.Culture >= st.Cost {
			lines = append(lines, "  Status: [green]ready[-] — type [cyan]festival confirm yes[-].")
		} else {
			lines = append(lines, "  Status: [yellow]not enough culture yet.[-]")
		}
	} else {
		lines = append(lines, fmt.Sprintf("  Status: [yellow]on cooldown[-] — %d ticks remaining.", st.CooldownLeft))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdBlackMarket handles the `blackmarket` / `trade black` command. Bare form
// shows the status panel (cost, odds, cooldown). `blackmarket <resource>` spends
// a lump of culture on a high-risk/high-reward deal that may pay out a large
// amount of the chosen resource — or vanish with the culture.
func cmdBlackMarket(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return cmdBlackMarketStatus(engine)
	}
	resource := strings.ToLower(args[0])

	won, gain, err := engine.DoBlackMarket(resource)
	if err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	if won {
		return CommandResult{
			Message: fmt.Sprintf("[green]The deal paid off![-] Smugglers delivered %.1f %s.", gain, resource),
			Type:    "success",
		}
	}
	return CommandResult{
		Message: fmt.Sprintf("[red]The deal went bad.[-] The culture is gone and no %s arrived.", resource),
		Type:    "warning",
	}
}

// cmdBlackMarketStatus renders the bare `blackmarket` status panel.
func cmdBlackMarketStatus(engine *game.GameEngine) CommandResult {
	st := engine.BlackMarketStatus()
	var lines []string
	lines = append(lines, "[gold]Black Market[-]")
	if !st.Available {
		lines = append(lines, "  [gray]Smuggling networks open in the Colonial Age.[-]")
		return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
	}
	lines = append(lines, "  Spend culture on a high-risk smuggling deal for a chance at a big resource haul.")
	lines = append(lines, fmt.Sprintf("  Cost: [cyan]%.0f culture[-] per deal  (you have %.0f)", st.Cost, st.Culture))
	lines = append(lines, fmt.Sprintf("  Odds: [green]%.0f%%[-] payout at [green]%.1fx[-] value, else the culture is lost.", st.WinChance*100, st.WinMult))
	if st.Ready {
		if st.Culture >= st.Cost {
			lines = append(lines, "  Status: [green]ready[-] — type [cyan]blackmarket <resource>[-] (e.g. blackmarket gold).")
		} else {
			lines = append(lines, "  Status: [yellow]not enough culture yet.[-]")
		}
	} else {
		lines = append(lines, fmt.Sprintf("  Status: [yellow]lying low[-] — %d ticks until the next deal.", st.CooldownLeft))
	}
	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
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
		lines = append(lines, "  [red]ALL progress will be reset:[-] resources, buildings, workers, research, military.")
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

// cmdScoutingList prints only the available SCOUTING expeditions, with the
// active scout (if any) as a footer. This backs `expedition list`.
func cmdScoutingList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Expeditions:[-]")

	appendExpeditionGroup(&lines, "Scouting", state.Military.Expeditions, game.ExpeditionScouting)

	if state.Military.ActiveScout != nil {
		lines = append(lines, fmt.Sprintf("\n[yellow]Active expedition: %s (%d ticks left)[-]",
			state.Military.ActiveScout.Name, state.Military.ActiveScout.TicksLeft))
	}

	if !hasCategory(state.Military.Expeditions, game.ExpeditionScouting) {
		lines = append(lines, "  [gray]No expeditions available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// cmdCampaignList prints only the available MILITARY campaigns, with the active
// campaign (if any) as a footer. This backs `campaign` and `campaign list`.
func cmdCampaignList(engine *game.GameEngine) CommandResult {
	state := engine.GetState()
	var lines []string
	lines = append(lines, "[gold]Available Campaigns:[-]")

	appendExpeditionGroup(&lines, "Campaigns", state.Military.Expeditions, game.ExpeditionMilitary)

	if state.Military.ActiveMilitary != nil {
		lines = append(lines, fmt.Sprintf("\n[yellow]Active campaign: %s (%d ticks left)[-]",
			state.Military.ActiveMilitary.Name, state.Military.ActiveMilitary.TicksLeft))
	}

	if !hasCategory(state.Military.Expeditions, game.ExpeditionMilitary) {
		lines = append(lines, "  [gray]No campaigns available yet[-]")
	}

	return CommandResult{Message: strings.Join(lines, "\n"), Type: "info"}
}

// hasCategory reports whether any expedition in exps matches the given category.
func hasCategory(exps []game.ExpeditionInfo, category string) bool {
	for _, exp := range exps {
		if exp.Category == category {
			return true
		}
	}
	return false
}

// appendExpeditionGroup appends the subset of exps matching category to lines,
// under a labeled header (e.g. "Scouting"). The header is omitted when no
// expedition matches, so empty subsections produce no output.
func appendExpeditionGroup(lines *[]string, label string, exps []game.ExpeditionInfo, category string) {
	first := true
	for _, exp := range exps {
		if exp.Category != category {
			continue
		}
		if first {
			*lines = append(*lines, fmt.Sprintf("[yellow]%s:[-]", label))
			first = false
		}
		canStr := "[red]✗[-]"
		if exp.CanLaunch {
			canStr = "[green]✓[-]"
		}
		// Soldier-free scouting expeditions omit the soldier prefix; military
		// campaigns lead with their soldier requirement.
		var reqParts []string
		if exp.SoldiersNeeded > 0 {
			reqParts = append(reqParts, fmt.Sprintf("%d soldiers", exp.SoldiersNeeded))
		}
		if cost := formatExpeditionCost(exp.Cost); cost != "" {
			reqParts = append(reqParts, cost)
		}
		reqParts = append(reqParts, fmt.Sprintf("%d ticks", exp.Duration))
		reqs := strings.Join(reqParts, ", ")
		line := fmt.Sprintf("  %s [cyan]%s[-] - %s (%s)", canStr, exp.Key, exp.Name, reqs)
		if !exp.CanLaunch && exp.LaunchBlockReason != "" {
			line += fmt.Sprintf(" [red]— %s[-]", exp.LaunchBlockReason)
		}
		*lines = append(*lines, line)
	}
}

func cmdTrade(args []string, engine *game.GameEngine) CommandResult {
	if len(args) == 0 {
		return CommandResult{OverlayName: "trade"}
	}
	subcmd := strings.ToLower(args[0])

	if subcmd == "list" {
		return cmdTradeList(engine)
	}
	if subcmd == "route" {
		return cmdTradeRoute(args[1:], engine)
	}
	if subcmd == "black" {
		// `trade black [resource]` is an alias for the black-market command.
		return cmdBlackMarket(args[1:], engine)
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

	case "tribute":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy tribute <civ_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.SendTribute(factionKey); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Tribute paid to %s — peace restored.", factionKey), Type: "success"}

	case "raid":
		if len(args) < 2 {
			return CommandResult{Message: "Usage: diplomacy raid <civ_key>", Type: "error"}
		}
		factionKey := strings.Join(args[1:], "_")
		if err := engine.RaidCivRoute(factionKey); err != nil {
			return CommandResult{Message: err.Error(), Type: "error"}
		}
		return CommandResult{Message: fmt.Sprintf("Raided %s's trade route.", factionKey), Type: "warning"}

	default:
		return CommandResult{Message: "Usage: diplomacy [ally|rival|embargo|gift|neutral|tribute|raid] <civ_key>", Type: "error"}
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

func cmdSell(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: sell <building> [count]", Type: "error"}
	}
	building := args[0]
	count := 1
	if len(args) >= 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n <= 0 {
			return CommandResult{Message: "Usage: sell <building> [count]", Type: "error"}
		}
		count = n
	}
	if err := engine.SellBuilding(building, count); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{Message: "", Type: "success"}
}

func cmdDismiss(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 {
		return CommandResult{Message: "Usage: dismiss <building> [count|all]", Type: "error"}
	}
	building := args[0]
	all := false
	count := 1
	if len(args) >= 2 {
		if strings.ToLower(args[1]) == "all" {
			all = true
		} else {
			n, err := strconv.Atoi(args[1])
			if err != nil || n <= 0 {
				return CommandResult{Message: "Usage: dismiss <building> [count|all]", Type: "error"}
			}
			count = n
		}
	}
	if err := engine.DismissWorkers(building, count, all); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{Message: "", Type: "success"}
}

func cmdCatastrophe(args []string, engine *game.GameEngine) CommandResult {
	if len(args) < 1 || strings.ToLower(args[0]) != "invoke" {
		return CommandResult{
			Message: "Usage: catastrophe invoke — voluntarily trigger the epoch catastrophe\n" +
				"  [red]Warning: this will open the Endure / Succumb / Defer modal.[-]",
			Type: "info",
		}
	}
	if err := engine.InvokeCatastrophe(); err != nil {
		return CommandResult{Message: err.Error(), Type: "error"}
	}
	return CommandResult{
		Message: "[red]Catastrophe invoked! A choice awaits...[-]",
		Type:    "warning",
	}
}
