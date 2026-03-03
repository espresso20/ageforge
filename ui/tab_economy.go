package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// EconomyTab displays resources, buildings, and villager management
type EconomyTab struct {
	root       *tview.Flex
	resourceTV *tview.TextView
	buildingTV *tview.TextView
	villagerTV *tview.TextView
}

// NewEconomyTab creates the economy tab
func NewEconomyTab() *EconomyTab {
	t := &EconomyTab{}

	t.resourceTV = tview.NewTextView().SetDynamicColors(true)
	t.resourceTV.SetBorder(true).SetTitle(" Resources ").SetTitleColor(ColorResource)

	t.buildingTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.buildingTV.SetBorder(true).SetTitle(" Buildings ").SetTitleColor(ColorBuilding)

	t.villagerTV = tview.NewTextView().SetDynamicColors(true)
	t.villagerTV.SetBorder(true).SetTitle(" Population ").SetTitleColor(ColorVillager)

	// Left: resources + villagers, Right: buildings
	leftCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.resourceTV, 0, 3, false).
		AddItem(t.villagerTV, 0, 2, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftCol, 0, 1, false).
		AddItem(t.buildingTV, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *EconomyTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the economy tab with current game state
func (t *EconomyTab) Refresh(state game.GameState) {
	t.refreshResources(state)
	t.refreshBuildings(state)
	t.refreshVillagers(state)
}

func (t *EconomyTab) refreshResources(state game.GameState) {
	var sb strings.Builder

	keys := make([]string, 0)
	for k, rs := range state.Resources {
		if rs.Unlocked {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		rs := state.Resources[key]
		bar := ProgressBar(rs.Amount, rs.Storage, 12)
		amtColor := "white"
		if rs.Storage > 0 && rs.Amount >= rs.Storage*0.9 {
			amtColor = "yellow"
		} else if rs.Rate > 0 {
			amtColor = "green"
		} else if rs.Rate < 0 {
			amtColor = "red"
		}
		fmt.Fprintf(&sb, " %-12s [%s]%6s[-] / %-6s %s %s\n",
			rs.Name, amtColor, FormatNumber(rs.Amount), FormatNumber(rs.Storage),
			FormatRate(rs.Rate), bar)
	}
	t.resourceTV.SetText(sb.String())
}

func (t *EconomyTab) refreshBuildings(state game.GameState) {
	// Build age ordering from config
	ageOrder := config.AgeOrder()
	ageIndex := make(map[string]int, len(ageOrder))
	for i, k := range ageOrder {
		ageIndex[k] = i
	}
	ageByKey := config.AgeByKey()

	// Group unlocked buildings by their age key
	byAge := make(map[string][]string)
	for key, bs := range state.Buildings {
		if bs.Unlocked {
			byAge[bs.AgeKey] = append(byAge[bs.AgeKey], key)
		}
	}

	// Sort groups: current age first, then descending (most recent prior age first)
	groupKeys := make([]string, 0, len(byAge))
	for k := range byAge {
		groupKeys = append(groupKeys, k)
	}
	sort.Slice(groupKeys, func(i, j int) bool {
		if groupKeys[i] == state.Age {
			return true
		}
		if groupKeys[j] == state.Age {
			return false
		}
		return ageIndex[groupKeys[i]] > ageIndex[groupKeys[j]]
	})

	var sb strings.Builder
	for _, ageKey := range groupKeys {
		keys := byAge[ageKey]
		sort.Strings(keys)

		ageName := ageKey
		if def, ok := ageByKey[ageKey]; ok {
			ageName = def.Name
		}
		headerColor := "gray"
		if ageKey == state.Age {
			headerColor = "gold"
		}
		fmt.Fprintf(&sb, " [%s]── %s ──[-]\n", headerColor, ageName)

		for _, key := range keys {
			bs := state.Buildings[key]
			var icon string
			switch {
			case bs.AtMaxCount:
				icon = "[yellow]MAX[-]"
			case bs.CanBuild:
				icon = "[green]✓[-]"
			default:
				icon = "[red]✗[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-] [gray]x%d[-]\n", icon, bs.Name, bs.Count)
			if bs.AtMaxCount {
				fmt.Fprintf(&sb, "   [yellow]Building limit reached.[-]\n")
			} else {
				fmt.Fprintf(&sb, "   Cost: %s\n", FormatCost(bs.NextCost))
			}
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", bs.Description)
		}
		sb.WriteString("\n")
	}

	if len(state.BuildQueue) > 0 {
		sb.WriteString(" [gold]Under Construction:[-]\n")
		for _, item := range state.BuildQueue {
			bar := ProgressBar(float64(item.TotalTicks-item.TicksLeft), float64(item.TotalTicks), 10)
			eta := FormatETA(state.TickIntervalMs * item.TicksLeft)
			fmt.Fprintf(&sb, "   [yellow]%s[-] %s %d ticks (%s)\n", item.Name, bar, item.TicksLeft, eta)
		}
	}

	if sb.Len() == 0 {
		sb.WriteString(" [gray]No buildings unlocked yet[-]")
	}
	t.buildingTV.SetText(sb.String())
}

// ScrollUp scrolls the buildings panel up
func (t *EconomyTab) ScrollUp() {
	row, col := t.buildingTV.GetScrollOffset()
	t.buildingTV.ScrollTo(row-10, col)
}

// ScrollDown scrolls the buildings panel down
func (t *EconomyTab) ScrollDown() {
	row, col := t.buildingTV.GetScrollOffset()
	t.buildingTV.ScrollTo(row+10, col)
}

func (t *EconomyTab) refreshVillagers(state game.GameState) {
	var sb strings.Builder
	v := state.Villagers

	fmt.Fprintf(&sb, " [gold]Total:[-] %d/%d  [gold]Idle:[-] %d  [gold]Food:[-] %.1f/tick\n\n",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain)

	vtKeys := make([]string, 0)
	for k, vt := range v.Types {
		if vt.Unlocked {
			vtKeys = append(vtKeys, k)
		}
	}
	sort.Strings(vtKeys)

	for _, vtKey := range vtKeys {
		vt := v.Types[vtKey]
		fmt.Fprintf(&sb, " [mediumpurple]%s[-] x%d (idle: %d)\n", vt.Name, vt.Count, vt.IdleCount)
		aKeys := make([]string, 0, len(vt.Assignments))
		for res := range vt.Assignments {
			aKeys = append(aKeys, res)
		}
		sort.Strings(aKeys)
		for _, res := range aKeys {
			if vt.Assignments[res] > 0 {
				fmt.Fprintf(&sb, "   → %s: %d\n", res, vt.Assignments[res])
			}
		}
	}
	t.villagerTV.SetText(sb.String())
}
