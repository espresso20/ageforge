package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// buildingsProvider generates the buildings overlay text from the current game
// state. It lists all unlocked buildings across all ages up to and including
// the current age, grouped by age in chronological order (oldest first).
// This is a read-only browser — players use the 'build' command to construct.
func buildingsProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	ageOrder := config.AgeOrder()
	ageByKey := config.AgeByKey()

	// Find index of the current age so we only show ages up to and including it.
	currentIdx := -1
	for i, ak := range ageOrder {
		if ak == state.Age {
			currentIdx = i
			break
		}
	}
	if currentIdx < 0 {
		// Fallback: show all ages if current age is unrecognised.
		currentIdx = len(ageOrder) - 1
	}

	// Count total unlocked buildings for the header.
	totalUnlocked := 0
	totalBuilt := 0
	for _, bs := range state.Buildings {
		if bs.Unlocked {
			totalUnlocked++
			if bs.Count > 0 {
				totalBuilt++
			}
		}
	}

	fmt.Fprintf(&sb, "[gold]═══ Buildings: %d built / %d unlocked ═══[-]\n", totalBuilt, totalUnlocked)
	sb.WriteString(" [gray]Browse all ages. Use 'build <key>' to construct (current age only).[-]\n\n")

	for i := 0; i <= currentIdx; i++ {
		ageKey := ageOrder[i]
		ageDef, ok := ageByKey[ageKey]
		if !ok {
			continue
		}

		// Collect buildings that belong to this age and are unlocked.
		type entry struct {
			key string
			bs  game.BuildingState
		}
		var entries []entry
		for key, bs := range state.Buildings {
			if bs.AgeKey == ageKey && bs.Unlocked {
				entries = append(entries, entry{key: key, bs: bs})
			}
		}
		if len(entries) == 0 {
			continue
		}

		// Sort by name for stable, readable output.
		sort.Slice(entries, func(a, b int) bool {
			return entries[a].bs.Name < entries[b].bs.Name
		})

		// Age header.
		isCurrent := (ageKey == state.Age)
		if isCurrent {
			fmt.Fprintf(&sb, "[gold]── %s (current) ──[-]\n", ageDef.Name)
		} else {
			fmt.Fprintf(&sb, "[gray]── %s ──[-]\n", ageDef.Name)
		}

		for _, e := range entries {
			bs := e.bs

			// Build status indicator.
			builtPart := ""
			if bs.Count > 0 {
				builtPart = fmt.Sprintf(" [green]x%d[-]", bs.Count)
			}

			// Legacy indicator.
			legacyPart := ""
			if bs.IsLegacy {
				legacyPart = " [gray](legacy)[-]"
			}

			// Worker info.
			workerPart := ""
			if bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
				workerPart = fmt.Sprintf(" [cyan]workers: %d/%d[-]", bs.WorkersAssigned, bs.WorkerCapacity*bs.Count)
				if bs.Count == 0 {
					workerPart = fmt.Sprintf(" [gray]workers: cap %d/bldg[-]", bs.WorkerCapacity)
				}
			}

			// Name line.
			nameColor := "yellow"
			if bs.Count > 0 {
				nameColor = "green"
			} else if bs.IsLegacy {
				nameColor = "gray"
			}
			fmt.Fprintf(&sb, " [%s]%s[-]%s%s%s\n", nameColor, bs.Name, builtPart, legacyPart, workerPart)

			// Cost and description on indent.
			if len(bs.NextCost) > 0 && !bs.IsLegacy {
				fmt.Fprintf(&sb, "   [gray]Cost: %s[-]\n", FormatCost(bs.NextCost))
			}
			if bs.Description != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", bs.Description)
			}
			if bs.Flavor != "" {
				fmt.Fprintf(&sb, "   [gray::i]%s[-:-:-]\n", bs.Flavor)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
