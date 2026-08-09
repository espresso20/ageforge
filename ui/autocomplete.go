// Package ui provides all terminal user interface components for AgeForge,
// built on top of tview/tcell. The package is structured around a permanent
// Dashboard (economy background) with named overlay panels for research,
// military, trade, etc. All UI refreshes must happen inside app.QueueUpdateDraw.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// commands is the full list of command names for autocomplete.
// Includes both full names and single-letter/short aliases so they show up in tab completion.
var commands = []string{
	"gather", "g",
	"build", "b",
	"recruit", "r",
	"assign", "a",
	"unassign", "u",
	"dismiss",
	"sell",
	"research", "res",
	"expedition", "exp",
	"campaign",
	"trade", "t",
	"factions", "diplomacy", "dip",
	"catastrophe", "cat",
	"status", "s",
	"help", "h",
	"prestige",
	"festival",
	"blackmarket", "bm",
	"upgrade",
	"advance", "rates", "speed", "save", "saves", "load",
	"account",
	"theme",
	"wonder",
	"dump", "exportlogs",
	"milestones", "ms",
	"techs", "army", "stats", "wonders", "workers", "logs", "epoch", "history", "buildings",
	"citymap", "worldmap", "map",
}

// NewAutoCompleter returns an autocomplete function for the command input field.
// It uses live game state to provide context-aware suggestions.
func NewAutoCompleter(engine *game.GameEngine) func(string) []string {
	return func(currentText string) []string {
		text := strings.TrimLeft(currentText, " ")
		if text == "" {
			return nil
		}

		parts := strings.Fields(text)
		// Check if there's a trailing space (user finished typing a word)
		hasTrailingSpace := len(currentText) > 0 && currentText[len(currentText)-1] == ' '

		if len(parts) == 1 && !hasTrailingSpace {
			// Partial command name
			return filterPrefix(commands, parts[0], "")
		}

		cmd := strings.ToLower(parts[0])
		args := parts[1:]
		prefix := cmd + " "

		if hasTrailingSpace {
			// User typed "cmd arg1 " — all args are completed, suggest next with empty filter
			argPrefix := prefix
			for _, p := range args {
				argPrefix += p + " "
			}
			return suggestArg(cmd, args, "", argPrefix, engine)
		}

		// User is typing an argument — filter on last partial word
		partial := strings.ToLower(args[len(args)-1])
		completed := args[:len(args)-1]
		argPrefix := prefix
		for _, p := range completed {
			argPrefix += p + " "
		}
		return suggestArg(cmd, completed, partial, argPrefix, engine)
	}
}

// suggestArg returns suggestions for the argument position of a command.
// completed contains fully-typed argument words, partial is what's being typed,
// and prefix is the string to prepend to each suggestion.
func suggestArg(cmd string, completed []string, partial string, prefix string, engine *game.GameEngine) []string {
	state := engine.GetState()

	switch cmd {
	case "gather", "g":
		// cmdGather only accepts food, wood, stone — not all resource keys
		return filterPrefix([]string{"food", "wood", "stone"}, partial, prefix)

	case "build", "b":
		if len(completed) == 0 {
			return filterPrefix(unlockedBuildingKeys(state), partial, prefix)
		}
		if len(completed) == 1 {
			return filterPrefix([]string{"max"}, partial, prefix)
		}

	case "recruit", "r":
		if len(completed) == 0 {
			return filterPrefix([]string{"max"}, partial, prefix)
		}

	case "assign", "a":
		if len(completed) == 0 {
			return filterPrefix(workerBuildingKeys(state), partial, prefix)
		}
		return filterPrefix([]string{"all"}, partial, prefix)

	case "unassign", "u":
		if len(completed) == 0 {
			return filterPrefix(assignedBuildingKeysAll(state), partial, prefix)
		}
		return filterPrefix([]string{"all"}, partial, prefix)

	case "dismiss":
		if len(completed) == 0 {
			return filterPrefix(assignedBuildingKeysAll(state), partial, prefix)
		}
		return filterPrefix([]string{"all"}, partial, prefix)

	case "sell":
		if len(completed) == 0 {
			return filterPrefix(builtBuildingKeys(state), partial, prefix)
		}

	case "research", "res":
		keys := availableTechKeys(state)
		keys = append(keys, "list", "cancel")
		return filterPrefix(keys, partial, prefix)

	case "expedition", "exp":
		keys := expeditionKeysByCategory(state, game.ExpeditionScouting)
		keys = append(keys, "list")
		return filterPrefix(keys, partial, prefix)

	case "campaign":
		keys := expeditionKeysByCategory(state, game.ExpeditionMilitary)
		keys = append(keys, "list")
		return filterPrefix(keys, partial, prefix)

	case "prestige":
		if len(completed) == 0 {
			return filterPrefix([]string{"confirm", "shop", "buy"}, partial, prefix)
		}
		if strings.ToLower(completed[0]) == "buy" {
			return filterPrefix(prestigeUpgradeKeys(state), partial, prefix)
		}
		if strings.ToLower(completed[0]) == "confirm" {
			return filterPrefix([]string{"yes"}, partial, prefix)
		}

	case "festival":
		if len(completed) == 0 {
			return filterPrefix([]string{"confirm"}, partial, prefix)
		}
		if strings.ToLower(completed[0]) == "confirm" {
			return filterPrefix([]string{"yes"}, partial, prefix)
		}

	case "blackmarket", "bm":
		if len(completed) == 0 {
			return filterPrefix(unlockedResourceKeys(state), partial, prefix)
		}

	case "trade", "t":
		if len(completed) == 0 {
			// First arg: "list", "route", "black", or resource name for exchange
			keys := unlockedResourceKeys(state)
			keys = append(keys, "list", "route", "black")
			return filterPrefix(keys, partial, prefix)
		}
		if strings.ToLower(completed[0]) == "black" {
			if len(completed) == 1 {
				return filterPrefix(unlockedResourceKeys(state), partial, prefix)
			}
		}
		if strings.ToLower(completed[0]) == "route" {
			if len(completed) == 1 {
				return filterPrefix([]string{"list", "start", "stop"}, partial, prefix)
			}
			if len(completed) == 2 {
				sub := strings.ToLower(completed[1])
				if sub == "start" {
					return filterPrefix(availableTradeRouteKeys(state), partial, prefix)
				}
				if sub == "stop" {
					return filterPrefix(activeTradeRouteKeys(state), partial, prefix)
				}
			}
		}
		if len(completed) == 1 {
			// Second arg for exchange: target resource
			return filterPrefix(unlockedResourceKeys(state), partial, prefix)
		}

	case "diplomacy", "dip":
		if len(completed) == 0 {
			return filterPrefix([]string{"ally", "rival", "embargo", "gift", "neutral"}, partial, prefix)
		}
		if len(completed) == 1 {
			return filterPrefix(discoveredFactionKeys(state), partial, prefix)
		}

	case "upgrade":
		if len(completed) == 0 {
			return filterPrefix(upgradeableBuildingKeys(engine), partial, prefix)
		}
		if len(completed) == 1 {
			return filterPrefix([]string{"all"}, partial, prefix)
		}

	case "speed":
		return filterPrefix(availableSpeedOptions(engine), partial, prefix)

	case "save":
		// "save list" shows save files; "save [name]" saves to named slot
		names := saveNames()
		names = append([]string{"list"}, names...)
		return filterPrefix(names, partial, prefix)

	case "load":
		return filterPrefix(saveNames(), partial, prefix)

	case "theme":
		// "theme list" plus a key per registered theme. Only the first arg is
		// completable; `theme <key>` takes no further args.
		if len(completed) == 0 {
			return filterPrefix(append([]string{"list"}, themeKeys()...), partial, prefix)
		}

	case "account", "acct":
		// Subcommands. `switch <name>` completes to the local account display names; the
		// rest take args that aren't enumerable here (recover code, export/import paths).
		if len(completed) == 0 {
			return filterPrefix([]string{"list", "switch", "recover", "export", "backup", "import"}, partial, prefix)
		}
		if strings.ToLower(completed[0]) == "switch" {
			return filterPrefix(localAccountNames(engine), partial, prefix)
		}

	case "catastrophe", "cat":
		return filterPrefix([]string{"invoke"}, partial, prefix)

	case "wonder":
		if len(completed) == 0 {
			// Only subcommand is "collect"
			return filterPrefix([]string{"collect"}, partial, prefix)
		}
		if len(completed) == 1 && strings.ToLower(completed[0]) == "collect" {
			// 2nd arg: an unlocked resource key
			return filterPrefix(unlockedResourceKeys(state), partial, prefix)
		}
		if len(completed) == 2 && strings.ToLower(completed[0]) == "collect" {
			// 3rd arg: "all" or common numeric amounts
			return filterPrefix([]string{"all", "100", "1000"}, partial, prefix)
		}
	}

	return nil
}

// filterPrefix filters candidates by a prefix and prepends the given prefix string to each match.
func filterPrefix(candidates []string, partial string, prefix string) []string {
	if len(candidates) == 0 {
		return nil
	}
	partial = strings.ToLower(partial)
	var results []string
	for _, c := range candidates {
		if strings.HasPrefix(strings.ToLower(c), partial) {
			results = append(results, prefix+c)
		}
	}
	sort.Strings(results)
	return results
}

// unlockedResourceKeys returns all resource keys that are currently unlocked,
// sorted alphabetically. Used to populate autocomplete for gather/trade commands.
func unlockedResourceKeys(state game.GameState) []string {
	var keys []string
	for key, rs := range state.Resources {
		if rs.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// unlockedBuildingKeys returns all building keys that are currently unlocked,
// sorted alphabetically. Used for the "build" command autocomplete.
func unlockedBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// unlockedDomainKeys returns all worker domain keys that are currently unlocked,
// sourced from state.Workers.Types which is keyed by domain key (e.g. "food", "knowledge").
func unlockedDomainKeys(state game.GameState) []string {
	var keys []string
	for key, vt := range state.Workers.Types {
		if vt.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// recruitCompletions returns the current class names (in lowercase) for all unlocked
// worker domains, sourced from state.Workers.Types[domain].Name.
// Falls back to the domain key if the class name is empty.
func recruitCompletions(state game.GameState) []string {
	var names []string
	for domain, vt := range state.Workers.Types {
		if !vt.Unlocked {
			continue
		}
		name := strings.ToLower(vt.Name)
		if name == "" {
			name = domain
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// builtBuildingKeys returns all building keys where at least one copy has been built,
// sorted alphabetically. Used for the "sell" command autocomplete.
func builtBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Count > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// workerBuildingKeys returns all unlocked building keys that accept workers
// (i.e. have a WorkerDomain and WorkerCapacity > 0).
func workerBuildingKeys(state game.GameState) []string {
	var keys []string
	for key, bs := range state.Buildings {
		if bs.Unlocked && bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// assignedBuildingKeysAll returns all building keys across all domains that currently
// have at least one worker assigned.
func assignedBuildingKeysAll(state game.GameState) []string {
	seen := map[string]bool{}
	for _, vt := range state.Workers.Types {
		for buildingKey, count := range vt.Assignments {
			if count > 0 {
				seen[buildingKey] = true
			}
		}
	}
	var keys []string
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// availableTechKeys returns tech keys that are currently available to research
// (unlocked but not yet started or completed).
func availableTechKeys(state game.GameState) []string {
	var keys []string
	for key, ts := range state.Research.Techs {
		if ts.Available {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// expeditionKeysByCategory returns the visible expedition keys for a single
// category (game.ExpeditionScouting or game.ExpeditionMilitary), sorted. Used so
// `expedition <key>` completes to scouting keys and `campaign <key>` to military
// keys. This includes locked entries — the engine enforces eligibility on launch.
func expeditionKeysByCategory(state game.GameState, category string) []string {
	var keys []string
	for _, exp := range state.Military.Expeditions {
		if exp.Category == category {
			keys = append(keys, exp.Key)
		}
	}
	sort.Strings(keys)
	return keys
}

// prestigeUpgradeKeys returns upgrade keys that still have tiers remaining
// (NextCost > 0 means the upgrade is not yet maxed out).
func prestigeUpgradeKeys(state game.GameState) []string {
	var keys []string
	for key, u := range state.Prestige.Upgrades {
		if u.NextCost > 0 { // not maxed
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// availableTradeRouteKeys returns route keys from Trade.AvailableRoutes.
// These are routes the player can start (not currently active).
func availableTradeRouteKeys(state game.GameState) []string {
	var keys []string
	for _, route := range state.Trade.AvailableRoutes {
		keys = append(keys, route.Key)
	}
	sort.Strings(keys)
	return keys
}

// activeTradeRouteKeys returns route keys from Trade.ActiveRoutes.
// These are routes that are currently running and can be stopped.
func activeTradeRouteKeys(state game.GameState) []string {
	var keys []string
	for _, route := range state.Trade.ActiveRoutes {
		keys = append(keys, route.Key)
	}
	sort.Strings(keys)
	return keys
}

// discoveredFactionKeys returns faction keys that have been discovered
// (factions are hidden until the player reaches a certain age or event).
func discoveredFactionKeys(state game.GameState) []string {
	var keys []string
	for key, f := range state.Diplomacy.Factions {
		if f.Discovered {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// availableSpeedOptions returns a list of valid speed multiplier strings
// from 1.0 up to the current max (increments of 0.5). Max speed increases
// by 0.5x for each wonder the player has built.
func availableSpeedOptions(engine *game.GameEngine) []string {
	maxSpeed := engine.GetMaxSpeed()
	var options []string
	for s := 1.0; s <= maxSpeed; s += 0.5 {
		options = append(options, fmt.Sprintf("%.1f", s))
	}
	return options
}

// upgradeableBuildingKeys returns the building keys that have a pending player-driven upgrade.
func upgradeableBuildingKeys(engine *game.GameEngine) []string {
	upgrades := engine.GetAvailableUpgrades()
	var keys []string
	for _, u := range upgrades {
		keys = append(keys, u.FromKey)
	}
	sort.Strings(keys)
	return keys
}

func saveNames() []string {
	saves, err := game.ListSaves()
	if err != nil {
		return nil
	}
	return saves
}

// localAccountNames returns the display names of every local account slot, for
// `account switch <name>` completion. Unnamed slots are skipped (they can't be switched to by
// name). Read-only — ListAccounts never mutates the active account.
func localAccountNames(engine *game.GameEngine) []string {
	var names []string
	for _, s := range engine.ListAccounts() {
		if strings.TrimSpace(s.DisplayName) != "" {
			names = append(names, s.DisplayName)
		}
	}
	sort.Strings(names)
	return names
}
