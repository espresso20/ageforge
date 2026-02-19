package ui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// LogsTab shows game logs with a toggle between game-only and verbose mode
type LogsTab struct {
	root    *tview.Flex
	content *tview.TextView
	header  *tview.TextView
	verbose bool
}

// NewLogsTab creates the game logs tab
func NewLogsTab() *LogsTab {
	t := &LogsTab{}

	t.header = tview.NewTextView().
		SetDynamicColors(true)
	t.header.SetText(t.headerText())

	t.content = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(500)
	t.content.SetBorder(true).SetTitle(" Game Logs ").SetTitleColor(ColorTitle)

	t.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.header, 1, 0, false).
		AddItem(t.content, 0, 1, false)

	return t
}

// ToggleVerbose switches between game-only and verbose mode
func (t *LogsTab) ToggleVerbose() {
	t.verbose = !t.verbose
	t.header.SetText(t.headerText())
}

// IsVerbose returns the current verbose state
func (t *LogsTab) IsVerbose() bool {
	return t.verbose
}

func (t *LogsTab) headerText() string {
	mode := "[green]Game Logs[-]"
	if t.verbose {
		mode = "[yellow]Verbose (Debug)[-]"
	}
	return fmt.Sprintf(" Mode: %s  |  [gray]Press 'v' to toggle[-]", mode)
}

// ScrollUp scrolls content up
func (t *LogsTab) ScrollUp() {
	row, col := t.content.GetScrollOffset()
	t.content.ScrollTo(row-10, col)
}

// ScrollDown scrolls content down
func (t *LogsTab) ScrollDown() {
	row, col := t.content.GetScrollOffset()
	t.content.ScrollTo(row+10, col)
}

// Root returns the root primitive
func (t *LogsTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the logs tab with current game state
func (t *LogsTab) Refresh(state game.GameState) {
	var sb strings.Builder

	if t.verbose {
		t.renderVerbose(&sb, state)
	} else {
		t.renderGameLogs(&sb, state)
	}

	t.content.SetText(sb.String())
	t.content.ScrollToEnd()
}

// renderGameLogs shows only game activity logs
func (t *LogsTab) renderGameLogs(sb *strings.Builder, state game.GameState) {
	if len(state.Log) == 0 {
		sb.WriteString("[gray]No game logs yet.[-]")
		return
	}

	for _, entry := range state.Log {
		color := "white"
		prefix := ""
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
		case "debug":
			continue // skip debug in game-only mode
		}
		fmt.Fprintf(sb, "[gray]T%-5d[-] [%s]%s %s[-]\n", entry.Tick, color, prefix, entry.Message)
	}
}

// renderVerbose shows all logs including debug/engine info
func (t *LogsTab) renderVerbose(sb *strings.Builder, state game.GameState) {
	// Engine state header
	sb.WriteString("[gold]═══ Engine State ═══[-]\n")
	fmt.Fprintf(sb, " Tick: [cyan]%d[-]  Age: [cyan]%s[-]  Pop: [cyan]%d/%d[-]\n",
		state.Tick, state.AgeName, state.Villagers.TotalPop, state.Villagers.MaxPop)
	fmt.Fprintf(sb, " Food drain: [yellow]%.2f/tick[-]  Idle: [yellow]%d[-]\n",
		state.Villagers.FoodDrain, state.Villagers.TotalIdle)
	if state.TickSpeedBonus > 0 {
		fmt.Fprintf(sb, " Tick speed: [green]+%.0f%%[-] (interval: [cyan]%dms[-])\n",
			state.TickSpeedBonus*100, state.TickIntervalMs)
	} else {
		fmt.Fprintf(sb, " Tick speed: [gray]base[-] (interval: [cyan]%dms[-])\n", state.TickIntervalMs)
	}
	sb.WriteString("\n")

	// Resource rates
	sb.WriteString("[gold]═══ Resource Rates ═══[-]\n")
	for _, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		rateColor := "gray"
		if rs.Rate > 0 {
			rateColor = "green"
		} else if rs.Rate < 0 {
			rateColor = "red"
		}
		fmt.Fprintf(sb, " %-12s %8.1f / %8.0f  [%s]%+.3f/tick[-]\n",
			rs.Name, rs.Amount, rs.Storage, rateColor, rs.Rate)
	}
	sb.WriteString("\n")

	// Build queue
	if len(state.BuildQueue) > 0 {
		sb.WriteString("[gold]═══ Build Queue ═══[-]\n")
		for _, bq := range state.BuildQueue {
			pct := float64(bq.TotalTicks-bq.TicksLeft) / float64(bq.TotalTicks) * 100
			fmt.Fprintf(sb, " [cyan]%s[-] — %d/%d ticks (%.0f%%)\n",
				bq.Name, bq.TotalTicks-bq.TicksLeft, bq.TotalTicks, pct)
		}
		sb.WriteString("\n")
	}

	// Active events
	if len(state.ActiveEvents) > 0 {
		sb.WriteString("[gold]═══ Active Events ═══[-]\n")
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(sb, " [yellow]%s[-] — %d ticks left\n", evt.Name, evt.TicksLeft)
		}
		sb.WriteString("\n")
	}

	// Research
	if state.Research.CurrentTech != "" {
		sb.WriteString("[gold]═══ Research ═══[-]\n")
		fmt.Fprintf(sb, " [cyan]%s[-] — %d/%d ticks\n\n",
			state.Research.CurrentTechName,
			state.Research.TotalTicks-state.Research.TicksLeft,
			state.Research.TotalTicks)
	}

	// Villager assignments
	sb.WriteString("[gold]═══ Villager Assignments ═══[-]\n")
	for _, vt := range state.Villagers.Types {
		if !vt.Unlocked || vt.Count == 0 {
			continue
		}
		fmt.Fprintf(sb, " [cyan]%s[-] x%d (idle: %d)\n", vt.Name, vt.Count, vt.IdleCount)
		for res, count := range vt.Assignments {
			if count > 0 {
				fmt.Fprintf(sb, "   → %s: %d assigned\n", res, count)
			}
		}
	}
	sb.WriteString("\n")

	// Full log with debug entries
	sb.WriteString("[gold]═══ Full Log ═══[-]\n")
	if len(state.Log) == 0 {
		sb.WriteString("[gray]No logs yet.[-]")
		return
	}

	for _, entry := range state.Log {
		color := "white"
		prefix := ""
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
		case "debug":
			color = "gray"
			prefix = "[D]"
		}
		fmt.Fprintf(sb, "[gray]T%-5d[-] [%s]%s %s[-]\n", entry.Tick, color, prefix, entry.Message)
	}
}
