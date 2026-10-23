package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// buildingsProvider generates the buildings overlay text from the current game
// state. It lists all unlocked buildings across all ages up to and including
// the current age, grouped by age in chronological order (oldest first).
// This is a read-only browser: players use the build command to construct.
// Each row shows the key the build command takes next to the name.
func buildingsProvider(state game.GameState, screenW int) string {
	var sb strings.Builder

	set := state.Ruleset()
	ageOrder := set.AgeKeys()

	// Find index of the current age so we only show ages up to and including it.
	currentIdx := -1
	for i, ak := range ageOrder {
		if ak == state.Age {
			currentIdx = i
			break
		}
	}
	if currentIdx < 0 {
		// Fallback: show all ages if current age is unrecognized.
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

	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_buildings.buildingsProvider.01", "[gold]═══ Buildings: %d built, %d unlocked ═══[-]\n", totalBuilt, totalUnlocked))
	sb.WriteString(vibe.Line("ui.overlay_buildings.buildingsProvider.02", " [gray]Every age so far. Build with: build <key> (current age only).[-]\n\n"))

	for i := 0; i <= currentIdx; i++ {
		ageKey := ageOrder[i]
		ageDef, ok := set.Age(ageKey)
		if !ok {
			continue
		}

		// Collect buildings that belong to this age and are unlocked, and in
		// the current age the ones a tech still locks (listed after them).
		type entry struct {
			key string
			bs  game.BuildingState
		}
		var entries, locked []entry
		for key, bs := range state.Buildings {
			switch {
			case bs.AgeKey == ageKey && bs.Unlocked:
				entries = append(entries, entry{key: key, bs: bs})
			case bs.AgeKey == ageKey && ageKey == state.Age && !bs.Unlocked && bs.LockTech != "":
				locked = append(locked, entry{key: key, bs: bs})
			}
		}
		if len(entries)+len(locked) == 0 {
			continue
		}

		// Sort by name for stable, readable output.
		byName := func(es []entry) {
			sort.Slice(es, func(a, b int) bool {
				return es[a].bs.Name < es[b].bs.Name
			})
		}
		byName(entries)
		byName(locked)
		entries = append(entries, locked...)

		// Age header.
		isCurrent := (ageKey == state.Age)
		if isCurrent {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_buildings.buildingsProvider.03", "[gold]── %s (current) ──[-]\n", ageDef.Name))
		} else {
			fmt.Fprintf(&sb, "[gray]── %s ──[-]\n", ageDef.Name)
		}

		for _, e := range entries {
			bs := e.bs

			// A building a tech still locks: greyed out, the research it
			// needs, its cost and what it does, a word from the scholars.
			if !bs.Unlocked {
				width := overlayTextWidth(screenW)
				fmt.Fprintf(&sb, " [gray]%s %s (%s)[-]\n", lockedMarker, bs.Name, e.key)
				sb.WriteString(hangingRow("   ", "[gray]"+lockedNeedsLine(bs)+"[-]", width))
				if price := buildingPrice(&state, e.key, bs); len(price) > 0 {
					sb.WriteString("[gray]" + costRow(vibe.Line("ui.overlay_buildings.buildingsProvider.04", "   Cost: "), price, width) + "[-]")
				}
				if text := buildingText(bs); text != "" {
					fmt.Fprintf(&sb, "   [gray]%s[-]\n", text)
				}
				sb.WriteString(hangingRow("   ", "[gray::i]"+lockedHintLine(e.key)+"[-:-:-]", width))
				continue
			}

			// Build status indicator.
			builtPart := ""
			if bs.Count > 0 {
				builtPart = fmt.Sprintf(" [green]×%d[-]", bs.Count)
			}

			// Legacy indicator.
			legacyPart := ""
			if bs.IsLegacy {
				legacyPart = vibe.Line("ui.overlay_buildings.buildingsProvider.05", " [gray](legacy)[-]")
			}

			// Worker info.
			workerPart := ""
			if bs.WorkerDomain != "" && bs.WorkerCapacity > 0 {
				workerPart = vibe.Sprintf("ui.overlay_buildings.buildingsProvider.06", " [cyan]workers: %d/%d[-]", bs.WorkersAssigned, bs.WorkerCapacity*bs.Count)
				if bs.Count == 0 {
					workerPart = vibe.Sprintf("ui.overlay_buildings.buildingsProvider.07", " [gray]workers: up to %d each[-]", bs.WorkerCapacity)
				}
			}

			// Name line.
			nameColor := "yellow"
			if bs.Count > 0 {
				nameColor = "green"
			} else if bs.IsLegacy {
				nameColor = "gray"
			}
			fmt.Fprintf(&sb, " [%s]%s[-] [gray](%s)[-]%s%s%s\n", nameColor, bs.Name, e.key, builtPart, legacyPart, workerPart)

			// Cost and description on indent.
			if price := buildingPrice(&state, e.key, bs); len(price) > 0 && !bs.IsLegacy {
				sb.WriteString("[gray]" + costRow(vibe.Line("ui.overlay_buildings.buildingsProvider.08", "   Cost: "), price, overlayTextWidth(screenW)) + "[-]")
			}
			if text := buildingText(bs); text != "" {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", text)
			}
			if limit := atLimit(bs); limit != "" {
				fmt.Fprintf(&sb, "   [yellow]%s[-]\n", limit)
			}
			if bs.Flavor != "" {
				fmt.Fprintf(&sb, "   [comment::i]%s[-:-:-]\n", bs.Flavor)
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
