package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// MilitaryTab displays military units, expeditions, and defense
type MilitaryTab struct {
	root       *tview.Flex
	overviewTV *tview.TextView
	expedTV    *tview.TextView
	lootTV     *tview.TextView
}

// NewMilitaryTab creates the military tab
func NewMilitaryTab() *MilitaryTab {
	t := &MilitaryTab{}

	t.overviewTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.overviewTV.SetBorder(true).SetTitle(" Military Overview ").SetTitleColor(ColorTitle)

	t.expedTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.expedTV.SetBorder(true).SetTitle(" Expeditions ").SetTitleColor(ColorTitle)

	t.lootTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.lootTV.SetBorder(true).SetTitle(" Loot History ").SetTitleColor(ColorTitle)

	// Left: overview + loot, Right: expeditions
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.overviewTV, 10, 0, false).
		AddItem(t.lootTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftPanel, 0, 1, false).
		AddItem(t.expedTV, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *MilitaryTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the military tab with current state
func (t *MilitaryTab) Refresh(state game.GameState) {
	t.refreshOverview(state)
	t.refreshExpeditions(state)
	t.refreshLoot(state)
}

func (t *MilitaryTab) refreshOverview(state game.GameState) {
	var sb strings.Builder
	mil := state.Military

	fmt.Fprintf(&sb, " [gold]Soldiers:[-]  %d\n", mil.SoldierCount)
	fmt.Fprintf(&sb, " [gold]Defense:[-]   %.1f\n", mil.DefenseRating)

	if mil.MilitaryBonus > 0 {
		fmt.Fprintf(&sb, " [green]Military Bonus: +%.0f%%[-]\n", mil.MilitaryBonus*100)
	}
	if mil.ExpeditionBonus > 0 {
		fmt.Fprintf(&sb, " [green]Expedition Bonus: +%.0f%%[-]\n", mil.ExpeditionBonus*100)
	}

	sb.WriteString("\n")

	if mil.ActiveExpedition != nil {
		exp := mil.ActiveExpedition
		fmt.Fprintf(&sb, " [yellow]Active:[-] %s\n", exp.Name)
		fmt.Fprintf(&sb, " [yellow]Soldiers deployed:[-] %d\n", exp.Soldiers)
		bar := ProgressBar(float64(exp.TicksLeft), float64(exp.TicksLeft+1), 20) // approximate
		fmt.Fprintf(&sb, " %s %d ticks remaining\n", bar, exp.TicksLeft)
	} else {
		sb.WriteString(" [gray]No active expedition[-]\n")
	}

	fmt.Fprintf(&sb, "\n [gray]Completed: %d expeditions[-]", mil.CompletedCount)

	t.overviewTV.SetText(sb.String())
}

func (t *MilitaryTab) refreshExpeditions(state game.GameState) {
	var sb strings.Builder
	mil := state.Military

	if len(mil.Expeditions) == 0 {
		sb.WriteString(" [gray]No expeditions available yet[-]\n")
		sb.WriteString(" [gray]Reach Bronze Age and recruit soldiers[-]\n")
		sb.WriteString(" [gray]to unlock expeditions.[-]\n")
	} else {
		for _, exp := range mil.Expeditions {
			var statusIcon string
			if exp.CanLaunch {
				statusIcon = "[green]▸[-]"
			} else {
				statusIcon = "[red]▸[-]"
			}

			diffColor := "green"
			if exp.Difficulty > 0.5 {
				diffColor = "red"
			} else if exp.Difficulty > 0.3 {
				diffColor = "yellow"
			}

			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", statusIcon, exp.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", exp.Description)
			fmt.Fprintf(&sb, "   Soldiers: %d  Duration: %d ticks  Difficulty: [%s]%.0f%%[-]\n",
				exp.SoldiersNeeded, exp.Duration, diffColor, exp.Difficulty*100)

			if exp.CanLaunch {
				fmt.Fprintf(&sb, "   [green]expedition %s[-]\n", exp.Key)
			} else if mil.ActiveExpedition != nil {
				sb.WriteString("   [gray]expedition in progress[-]\n")
			} else {
				fmt.Fprintf(&sb, "   [red]need %d soldiers[-]\n", exp.SoldiersNeeded)
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(" [gray]Commands: expedition <key>[-]\n")

	t.expedTV.SetText(sb.String())
}

func (t *MilitaryTab) refreshLoot(state game.GameState) {
	var sb strings.Builder
	mil := state.Military

	if len(mil.TotalLoot) == 0 {
		sb.WriteString(" [gray]No loot collected yet[-]\n")
		sb.WriteString(" [gray]Complete expeditions to earn rewards![-]\n")
	} else {
		sb.WriteString(" [gold]Total Loot Collected:[-]\n\n")
		keys := make([]string, 0, len(mil.TotalLoot))
		for k := range mil.TotalLoot {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			amount := mil.TotalLoot[key]
			fmt.Fprintf(&sb, " %-12s [green]%.0f[-]\n", key, amount)
		}
	}

	t.lootTV.SetText(sb.String())
}
