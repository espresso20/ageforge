package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// cultureTresholds are the culture milestone amounts in ascending order.
var cultureThresholds = []float64{
	500, 2500, 10000, 50000, 250000, 1_000_000, 5_000_000, 25_000_000, 100_000_000, 500_000_000, 1_000_000_000,
}

// cultureThresholdLabels are short effect labels for each threshold (parallel to cultureThresholds).
var cultureThresholdLabels = []string{
	"+5% knowledge rate",
	"+10% knowledge rate",
	"+15% knowledge rate, unlock wonder tier",
	"+20% knowledge rate",
	"+25% knowledge rate, culture events",
	"+30% knowledge rate",
	"+research speed",
	"+research speed",
	"+research speed",
	"+research speed",
	"+research speed",
}

// cultureProgressBar returns a 10-char wide bar using ▓/░ characters.
func cultureProgressBar(current, max float64) string {
	const width = 10
	if max <= 0 {
		return strings.Repeat("░", width)
	}
	ratio := current / max
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	return strings.Repeat("▓", filled) + strings.Repeat("░", empty)
}

// formatCultureRow builds the culture resource row string.
func formatCultureRow(rs game.ResourceState) string {
	amount := rs.Amount

	// Find the next threshold not yet reached.
	nextIdx := -1
	for i, t := range cultureThresholds {
		if amount < t {
			nextIdx = i
			break
		}
	}

	var midPart string
	if nextIdx < 0 {
		// Above all thresholds — Culture Mastered.
		midPart = fmt.Sprintf("[gold]✦ Culture Mastered[-]  %-8s", FormatNumber(amount))
	} else {
		threshold := cultureThresholds[nextIdx]
		bar := cultureProgressBar(amount, threshold)
		label := cultureThresholdLabels[nextIdx]
		// Wrap bar in literal [ ] so tview does not interpret the block chars as a color tag.
		midPart = "\u005b" + bar + "\u005d" + fmt.Sprintf("  %s / %s  [gray]%s[-]",
			FormatNumber(amount), FormatNumber(threshold), label)
	}

	return fmt.Sprintf(" %-12s %s %s\n", rs.Name, midPart, FormatRate(rs.Rate))
}

// faithBand describes a faith band with its label and epoch odds text.
type faithBand struct {
	label     string // tview-tagged label
	epochOdds string // e.g. "40% good"
}

// faithBandFor returns the appropriate faithBand based on pct (0.0–1.0).
func faithBandFor(amount, storage float64) faithBand {
	if storage <= 0 || amount == 0 {
		return faithBand{"[red]✝ No Faith[-]", "40% good"}
	}
	pct := amount / storage
	switch {
	case pct <= 0.25:
		return faithBand{"[gray]◈ Dim Faith[-]", "40% good"}
	case pct <= 0.50:
		return faithBand{"[white]◈ Low Faith[-]", "50% good"}
	case pct <= 0.75:
		return faithBand{"[yellow]◈ Faith[-]", "50% good"}
	case pct < 1.0:
		return faithBand{"[green]◈ Strong Faith[-]", "60% good"}
	default:
		return faithBand{"[gold]✦ Faith Full[-]", "60% good + prestige bonus"}
	}
}

// formatFaithRow builds the faith resource row string.
func formatFaithRow(rs game.ResourceState) string {
	amount := rs.Amount
	storage := rs.Storage

	band := faithBandFor(amount, storage)

	var pctStr string
	if storage > 0 {
		pct := amount / storage * 100
		pctStr = fmt.Sprintf("%.0f%%", pct)
	} else {
		pctStr = "0%"
	}

	// Build the bar using the same cultureProgressBar helper (▓/░, width 10).
	bar := cultureProgressBar(amount, storage)
	barStr := "\u005b" + bar + "\u005d"

	midPart := fmt.Sprintf("%s  %s  %s  [gray](epoch: %s)[-]",
		barStr, band.label, pctStr, band.epochOdds)

	return fmt.Sprintf(" %-12s %s %s\n", rs.Name, midPart, FormatRate(rs.Rate))
}

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
		switch key {
		case "culture":
			sb.WriteString(formatCultureRow(rs))
		case "faith":
			sb.WriteString(formatFaithRow(rs))
		default:
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
			if bs.WorkerCapacity > 0 {
				totalCap := bs.Count * bs.WorkerCapacity
				bar := workerAssignBar(bs.WorkersAssigned, totalCap)
				barStr := "\u005b" + bar + "\u005d"
				domainLabel := domainToLabel[bs.WorkerDomain]
				if domainLabel == "" {
					domainLabel = bs.WorkerDomain
				}
				fmt.Fprintf(&sb, "   [green]Workers:[-] %d / %d %s  %s\n",
					bs.WorkersAssigned, totalCap, domainLabel, barStr)
			}
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

// legacyKeyToDomain maps legacy VillagerState.Types keys to their domain strings.
var legacyKeyToDomain = map[string]string{
	"worker":    "food",
	"shaman":    "faith",
	"scholar":   "knowledge",
	"soldier":   "military",
	"merchant":  "trade",
	"engineer":  "engineering",
	"hacker":    "hacker",
	"astronaut": "astronaut",
}

// domainToLabel maps domain strings to friendly display labels.
var domainToLabel = map[string]string{
	"food":        "Food",
	"faith":       "Faith",
	"knowledge":   "Knowledge",
	"military":    "Military",
	"trade":       "Trade",
	"engineering": "Engineering",
	"hacker":      "Hacker",
	"astronaut":   "Astronaut",
}

// workerAssignBar returns a 10-char ▓/░ bar for assigned/capacity.
func workerAssignBar(assigned, capacity int) string {
	const width = 10
	if capacity <= 0 {
		return strings.Repeat("░", width)
	}
	ratio := float64(assigned) / float64(capacity)
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filled := int(ratio * float64(width))
	empty := width - filled
	return strings.Repeat("▓", filled) + strings.Repeat("░", empty)
}

func (t *EconomyTab) refreshVillagers(state game.GameState) {
	var sb strings.Builder
	v := state.Villagers

	fmt.Fprintf(&sb, " [gold]Total:[-] %d/%d  [gold]Idle:[-] %d  [gold]Food:[-] %.1f/tick\n\n",
		v.TotalPop, v.MaxPop, v.TotalIdle, v.FoodDrain)

	// Collect unlocked domain keys in a stable order matching legacyKeyToDomain insertion order.
	// We define a fixed ordering so the display is deterministic.
	legacyOrder := []string{"worker", "shaman", "scholar", "soldier", "merchant", "engineer", "hacker", "astronaut"}

	for _, legacyKey := range legacyOrder {
		vt, ok := v.Types[legacyKey]
		if !ok || !vt.Unlocked {
			continue
		}

		domain := legacyKeyToDomain[legacyKey]
		label := domainToLabel[domain]
		if label == "" {
			label = vt.Name
		}

		// Determine class name from config, falling back gracefully.
		className := ""
		if wcd, found := config.WorkerClassByDomainAndAge(domain, state.Age); found {
			className = wcd.ClassName
		}

		if className != "" {
			fmt.Fprintf(&sb, " [yellow]── %s Workers (%s) ──[-]\n", label, className)
		} else {
			fmt.Fprintf(&sb, " [yellow]── %s Workers ──[-]\n", label)
		}

		fmt.Fprintf(&sb, "  Total: %d  Idle: %d  Food drain: %.1f/t\n", vt.Count, vt.IdleCount, v.FoodDrain)

		// Per-building assignment lines: only buildings this domain has workers in.
		bKeys := make([]string, 0, len(vt.Assignments))
		for bk := range vt.Assignments {
			bKeys = append(bKeys, bk)
		}
		sort.Strings(bKeys)

		for _, bKey := range bKeys {
			assigned := vt.Assignments[bKey]
			bs, bsOk := state.Buildings[bKey]
			if !bsOk {
				continue
			}
			totalCap := bs.Count * bs.WorkerCapacity
			if totalCap <= 0 {
				continue
			}
			bar := workerAssignBar(assigned, totalCap)
			barStr := "\u005b" + bar + "\u005d"
			fmt.Fprintf(&sb, "  %-20s %3d / %-3d %s\n", bs.Name, assigned, totalCap, barStr)
		}
		sb.WriteString("\n")
	}
	t.villagerTV.SetText(sb.String())
}
