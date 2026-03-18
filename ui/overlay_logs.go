package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// logsProvider generates the logs overlay text from the current game state.
// Shows the last ~50 game log entries, skipping debug entries, newest at the bottom.
func logsProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	sb.WriteString("[gold]═══ Game Log ═══[-]\n\n")

	logs := state.Log

	// Filter out debug entries and cap at last 50 visible entries
	var visible []game.LogEntry
	for _, entry := range logs {
		if entry.Type == "debug" {
			continue
		}
		visible = append(visible, entry)
	}

	if len(visible) == 0 {
		sb.WriteString("[gray]No game logs yet.[-]\n")
		return sb.String()
	}

	// Take last 50
	if len(visible) > 50 {
		visible = visible[len(visible)-50:]
	}

	for _, entry := range visible {
		color := "white"
		prefix := "   "
		switch entry.Type {
		case "success":
			color = "green"
			prefix = "[+]"
		case "warning":
			color = "yellow"
			prefix = "[!]"
		case "error":
			color = "red"
			prefix = "[X]"
		case "event":
			color = "gold"
			prefix = "[*]"
		case "info":
			color = "cyan"
			prefix = "[i]"
		}
		fmt.Fprintf(&sb, "[gray]T%-5d[-] [%s]%s %s[-]\n", entry.Tick, color, prefix, entry.Message)
	}

	// Engine state summary at the bottom
	sb.WriteString("\n[gold]═══ Engine State ═══[-]\n")
	fmt.Fprintf(&sb, " Tick: [cyan]%d[-]  Age: [cyan]%s[-]  Pop: [cyan]%d/%d[-]\n",
		state.Tick, state.AgeName, state.Workers.TotalPop, state.Workers.MaxPop)
	fmt.Fprintf(&sb, " Food drain: [yellow]%.2f/tick[-]  Idle: [yellow]%d[-]\n",
		state.Workers.FoodDrain, state.Workers.TotalIdle)
	if state.TickSpeedBonus > 0 {
		fmt.Fprintf(&sb, " Tick speed: [green]+%.0f%%[-] (interval: [cyan]%dms[-])\n",
			state.TickSpeedBonus*100, state.TickIntervalMs)
	} else {
		fmt.Fprintf(&sb, " Tick speed: [gray]base[-] (interval: [cyan]%dms[-])\n", state.TickIntervalMs)
	}

	// Active events (compact)
	if len(state.ActiveEvents) > 0 {
		sb.WriteString("\n [gold]Active Events:[-]\n")
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, "  [yellow]⚡ %s[-] (%d ticks)\n", evt.Name, evt.TicksLeft)
		}
	}

	// Build queue
	if len(state.BuildQueue) > 0 {
		sb.WriteString("\n [gold]Build Queue:[-]\n")
		for _, bq := range state.BuildQueue {
			pct := 0.0
			if bq.TotalTicks > 0 {
				pct = float64(bq.TotalTicks-bq.TicksLeft) / float64(bq.TotalTicks) * 100
			}
			bar := ProgressBar(float64(bq.TotalTicks-bq.TicksLeft), float64(bq.TotalTicks), 10)
			fmt.Fprintf(&sb, "  [cyan]%s[-] %s %.0f%%\n", bq.Name, bar, pct)
		}
	}

	// Research in progress
	if state.Research.CurrentTech != "" {
		sb.WriteString("\n [gold]Researching:[-]\n")
		done := state.Research.TotalTicks - state.Research.TicksLeft
		bar := ProgressBar(float64(done), float64(state.Research.TotalTicks), 10)
		fmt.Fprintf(&sb, "  [cyan]%s[-] %s (%d/%d ticks)\n",
			state.Research.CurrentTechName, bar, done, state.Research.TotalTicks)
	}

	// Worker assignments (compact)
	anyWorkers := false
	for _, vt := range state.Workers.Types {
		if vt.Unlocked && vt.Count > 0 {
			anyWorkers = true
			break
		}
	}
	if anyWorkers {
		sb.WriteString("\n [gold]Workers:[-]\n")
		// Stable domain order: sort by name
		type domainEntry struct {
			name        string
			count       int
			idle        int
			assignments map[string]int
		}
		var domains []domainEntry
		for _, vt := range state.Workers.Types {
			if !vt.Unlocked || vt.Count == 0 {
				continue
			}
			domains = append(domains, domainEntry{
				name:        vt.Name,
				count:       vt.Count,
				idle:        vt.IdleCount,
				assignments: vt.Assignments,
			})
		}
		// Sort by name for stable output
		for i := 0; i < len(domains); i++ {
			for j := i + 1; j < len(domains); j++ {
				if domains[i].name > domains[j].name {
					domains[i], domains[j] = domains[j], domains[i]
				}
			}
		}
		for _, d := range domains {
			fmt.Fprintf(&sb, "  [cyan]%s[-] x%d (idle: %d)\n", d.name, d.count, d.idle)
			for bldKey, count := range d.assignments {
				if count > 0 {
					fmt.Fprintf(&sb, "    → %s: %d\n", bldKey, count)
				}
			}
		}
	}

	return sb.String()
}
