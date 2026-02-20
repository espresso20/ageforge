package ui

import (
	"sort"
	"strings"

	"github.com/user/ageforge/game"
)

// commands is the full list of command names for autocomplete
var commands = []string{
	"gather", "build", "recruit", "assign", "unassign",
	"research", "expedition", "prestige",
	"rates", "status", "speed", "save", "saves", "load", "help", "quit",
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
		return filterPrefix(unlockedResourceKeys(state), partial, prefix)

	case "build", "b":
		return filterPrefix(unlockedBuildingKeys(state), partial, prefix)

	case "recruit", "r":
		return filterPrefix(unlockedVillagerTypes(state), partial, prefix)

	case "assign", "a":
		if len(completed) == 0 {
			return filterPrefix(unlockedVillagerTypes(state), partial, prefix)
		}
		if len(completed) == 1 {
			return filterPrefix(unlockedResourceKeys(state), partial, prefix)
		}
		return filterPrefix([]string{"all"}, partial, prefix)

	case "unassign", "u":
		if len(completed) == 0 {
			return filterPrefix(unlockedVillagerTypes(state), partial, prefix)
		}
		if len(completed) == 1 {
			vType := strings.ToLower(completed[0])
			return filterPrefix(assignedResources(state, vType), partial, prefix)
		}
		return filterPrefix([]string{"all"}, partial, prefix)

	case "research", "res":
		keys := availableTechKeys(state)
		keys = append(keys, "list", "cancel")
		return filterPrefix(keys, partial, prefix)

	case "expedition", "exp":
		keys := availableExpeditionKeys(state)
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

	case "speed":
		return filterPrefix([]string{"1", "2", "5", "10"}, partial, prefix)

	case "save":
		return filterPrefix(saveNames(), partial, prefix)

	case "load":
		return filterPrefix(saveNames(), partial, prefix)
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

func unlockedVillagerTypes(state game.GameState) []string {
	var keys []string
	for key, vt := range state.Villagers.Types {
		if vt.Unlocked {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func assignedResources(state game.GameState, vType string) []string {
	vt, ok := state.Villagers.Types[vType]
	if !ok {
		return nil
	}
	var keys []string
	for res, count := range vt.Assignments {
		if count > 0 {
			keys = append(keys, res)
		}
	}
	sort.Strings(keys)
	return keys
}

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

func availableExpeditionKeys(state game.GameState) []string {
	var keys []string
	for _, exp := range state.Military.Expeditions {
		keys = append(keys, exp.Key)
	}
	sort.Strings(keys)
	return keys
}

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

func saveNames() []string {
	saves, err := game.ListSaves()
	if err != nil {
		return nil
	}
	return saves
}
