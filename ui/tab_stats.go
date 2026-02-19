package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// StatsTab displays game statistics, milestones, and active events
type StatsTab struct {
	root       *tview.Flex
	statsTV    *tview.TextView
	milestoTV  *tview.TextView
	eventsTV   *tview.TextView
}

// NewStatsTab creates the stats tab
func NewStatsTab() *StatsTab {
	t := &StatsTab{}

	t.statsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.statsTV.SetBorder(true).SetTitle(" Statistics ").SetTitleColor(ColorTitle)

	t.milestoTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.milestoTV.SetBorder(true).SetTitle(" Milestones ").SetTitleColor(ColorTitle)

	t.eventsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.eventsTV.SetBorder(true).SetTitle(" Active Events ").SetTitleColor(ColorTitle)

	// Left: stats, Right: milestones + events
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.eventsTV, 8, 0, false).
		AddItem(t.milestoTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.statsTV, 0, 1, false).
		AddItem(rightPanel, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *StatsTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the stats tab
func (t *StatsTab) Refresh(state game.GameState) {
	t.refreshStats(state)
	t.refreshMilestones(state)
	t.refreshEvents(state)
}

func (t *StatsTab) refreshStats(state game.GameState) {
	var sb strings.Builder
	s := state.Stats
	p := state.Prestige

	// Prestige section (only show if level > 0)
	if p.Level > 0 {
		fmt.Fprintf(&sb, " [gold]Prestige Level:[-]    [cyan]%d[-]\n", p.Level)
		fmt.Fprintf(&sb, " [gold]Prestige Points:[-]   %d available / %d total\n", p.Available, p.TotalEarned)
		fmt.Fprintf(&sb, " [gold]Passive Bonus:[-]     [green]+%.0f%%[-] production\n", p.PassiveBonus*100)

		// Show purchased upgrades
		hasPurchased := false
		for _, key := range []string{
			"gather_boost", "storage_bonus", "research_speed", "military_power",
			"starting_food", "starting_wood", "population_cap", "expedition_loot",
		} {
			u, ok := p.Upgrades[key]
			if !ok || u.Tier == 0 {
				continue
			}
			if !hasPurchased {
				sb.WriteString(" [gold]Prestige Upgrades:[-]\n")
				hasPurchased = true
			}
			fmt.Fprintf(&sb, "   %s %d/%d — %s\n", u.Name, u.Tier, u.MaxTier, u.Effect)
		}
		sb.WriteString("\n")
	}

	fmt.Fprintf(&sb, " [gold]Play Time:[-]         %s\n", s.PlayTime.Truncate(1e9))
	fmt.Fprintf(&sb, " [gold]Total Ticks:[-]       %d\n", state.Tick)
	fmt.Fprintf(&sb, " [gold]Buildings Built:[-]    %d\n", s.TotalBuilt)
	fmt.Fprintf(&sb, " [gold]Villagers Recruited:[-] %d\n", s.TotalRecruited)
	fmt.Fprintf(&sb, " [gold]Techs Researched:[-]  %d\n", state.Research.TotalResearched)
	fmt.Fprintf(&sb, " [gold]Expeditions Done:[-]  %d\n", state.Military.CompletedCount)

	sb.WriteString("\n [gold]Ages Reached:[-]\n")
	for _, age := range s.AgesReached {
		fmt.Fprintf(&sb, "   [green]✓[-] %s\n", age)
	}

	sb.WriteString("\n [gold]Total Gathered:[-]\n")
	gKeys := make([]string, 0, len(s.TotalGathered))
	for k := range s.TotalGathered {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)
	for _, k := range gKeys {
		fmt.Fprintf(&sb, "   %-12s %s\n", k, FormatNumber(s.TotalGathered[k]))
	}

	t.statsTV.SetText(sb.String())
}

func (t *StatsTab) refreshMilestones(state game.GameState) {
	var sb strings.Builder
	ms := state.Milestones

	fmt.Fprintf(&sb, " [gold]Progress:[-] %d/%d\n\n", ms.CompletedCount, ms.TotalCount)

	// Sort milestone keys
	keys := make([]string, 0, len(ms.Milestones))
	for k := range ms.Milestones {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		m := ms.Milestones[key]
		if m.Completed {
			fmt.Fprintf(&sb, " [green]✓[-] [green]%s[-]\n", m.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", m.Description)
		} else {
			fmt.Fprintf(&sb, " [gray]○[-] [gray]%s[-]\n", m.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", m.Description)
		}
	}

	t.milestoTV.SetText(sb.String())
}

func (t *StatsTab) refreshEvents(state game.GameState) {
	var sb strings.Builder

	if len(state.ActiveEvents) == 0 {
		sb.WriteString(" [gray]No active events[-]\n")
		sb.WriteString(" [gray]Random events occur as you play[-]\n")
	} else {
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, " [yellow]⚡[-] [yellow]%s[-] (%d ticks left)\n", evt.Name, evt.TicksLeft)
		}
	}

	t.eventsTV.SetText(sb.String())
}
