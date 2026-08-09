package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// earlyGameAges lists the age keys during which the "Getting Started" onboarding
// block renders in the Buildings panel. Re-homed here out of the startup log so the
// otherwise-empty early-game panel does useful work; it auto-retires once the
// player advances past these ages. Tunable: trim to just {"primitive_age"} for a
// shorter intro, or extend it if later ages still warrant hand-holding.
var earlyGameAges = []string{"primitive_age", "stone_age"}

// isEarlyGame reports whether the given age key is one of the early-game ages that
// should show the onboarding block.
func isEarlyGame(ageKey string) bool {
	for _, k := range earlyGameAges {
		if k == ageKey {
			return true
		}
	}
	return false
}

// cultureThresholds defines the culture breakpoints at which rewards unlock.
// The bar displayed in the economy tab measures progress towards the next threshold.
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
// Filled portion uses BarFillColor (Accent role), empty uses BarEmptyColor (Dim role).
func cultureProgressBar(current, max float64) string {
	const width = 10
	if max <= 0 {
		return BarEmptyColor() + strings.Repeat("░", width) + "[-]"
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
	return BarFillColor() + strings.Repeat("▓", filled) + BarEmptyColor() + strings.Repeat("░", empty) + "[-]"
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

	return fmt.Sprintf(" %-12s %s %s\n\n", rs.Name, midPart, FormatRate(rs.Rate))
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

	return fmt.Sprintf(" %-12s %s %s\n\n", rs.Name, midPart, FormatRate(rs.Rate))
}

// EconomyTab is the permanent background panel visible at all times on the Dashboard.
// It is split into three panes: resource summary (left-top), under construction
// queue (left-bottom), and the full building list (right). The building panel is
// scrollable via PgUp/PgDn because it can grow very long in late ages.
type EconomyTab struct {
	root           *tview.Flex
	leftCol        *tview.Flex
	resourceTV     *tview.TextView
	buildingTV     *tview.TextView
	constructionTV *tview.TextView
}

// NewEconomyTab constructs the economy tab widget tree. The left column stacks
// resources / construction at a 3:1 height ratio (the Dashboard injects the log
// below at weight 2 → 3:1:2). The building panel takes a 1:1 column ratio against
// the left column (50:50) — narrowed from the old 4:5 so the building panel is
// tighter (descriptions word-wrap cleanly) and the left column (resources /
// construction / log) gets more room, improving log readability.
func NewEconomyTab() *EconomyTab {
	t := &EconomyTab{}

	t.resourceTV = tview.NewTextView().SetDynamicColors(true)
	t.resourceTV.SetBorder(true).SetTitle(" Resources ")

	t.buildingTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWordWrap(true)
	t.buildingTV.SetBorder(true).SetTitle(" Buildings ")

	t.constructionTV = tview.NewTextView().SetDynamicColors(true)
	t.constructionTV.SetBorder(true).SetTitle(" Under Construction ")

	// Persistent tab chrome: enroll titles so a live theme switch restyles them.
	theme.Track(func() {
		t.resourceTV.SetTitleColor(theme.Color(theme.RoleLabel))
		t.buildingTV.SetTitleColor(theme.Color(theme.RoleHighlight))
		t.constructionTV.SetTitleColor(theme.Color(theme.RoleHighlight))
	})

	// Left: resources + under construction (compact); the log is injected below
	// (Dashboard.AddToLeftColumn). Column weights resolve to 3:1:2
	// resources:construction:log — trimmed resources to give the log more headspace.
	t.leftCol = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.resourceTV, 0, 3, false).
		AddItem(t.constructionTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.leftCol, 0, 1, false).
		AddItem(t.buildingTV, 0, 1, false)

	return t
}

// AddToLeftColumn injects an additional item into the left column flex container.
// This allows the Dashboard to append the log panel below the construction queue.
func (t *EconomyTab) AddToLeftColumn(item tview.Primitive, fixedSize, proportion int) {
	t.leftCol.AddItem(item, fixedSize, proportion, false)
}

// Root returns the root primitive
func (t *EconomyTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the economy tab with current game state
func (t *EconomyTab) Refresh(state game.GameState) {
	t.refreshResources(state)
	t.refreshBuildings(state)
	t.refreshUnderConstruction(state)
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
			amtColor := "white"
			if rs.Storage > 0 && rs.Amount >= rs.Storage*0.9 {
				amtColor = "yellow"
			} else if rs.Rate > 0 {
				amtColor = "green"
			} else if rs.Rate < 0 {
				amtColor = "red"
			}
			rateColor := "green"
			if rs.Rate < 0 {
				rateColor = "red"
			} else if rs.Rate == 0 {
				rateColor = "gray"
			}
			bar := resourceBar(rs.Amount, rs.Storage, 12)
			glyph := ""
			if rs.Storage > 0 && rs.Amount/rs.Storage >= 0.95 {
				glyph = " [gold]◈[-]"
			} else if rs.Rate < 0 {
				glyph = " [red]▼[-]"
			}
			fmt.Fprintf(&sb, " %-14s [%s]%6s[-] [gray]/[-] [gray]%-6s[-]  [%s]%-8s[-]  %s%s\n\n",
				rs.Name, amtColor, FormatNumber(rs.Amount), FormatNumber(rs.Storage),
				rateColor, FormatRate(rs.Rate), bar, glyph)
		}
	}
	t.resourceTV.SetText(sb.String())
}

func (t *EconomyTab) refreshBuildings(state game.GameState) {
	// Build age ordering from config — used to sort building groups chronologically.
	ageOrder := config.AgeOrder()
	ageIndex := make(map[string]int, len(ageOrder))
	for i, k := range ageOrder {
		ageIndex[k] = i
	}
	ageByKey := config.AgeByKey()

	// Group unlocked buildings by their age key — only show current age
	byAge := make(map[string][]string)
	for key, bs := range state.Buildings {
		if bs.Unlocked && bs.AgeKey == state.Age {
			byAge[bs.AgeKey] = append(byAge[bs.AgeKey], key)
		}
	}

	// Sort groups: current age first (most relevant), then descending age order
	// so the player sees their newest buildings before ancient ones.
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
			countColor := "gray"
			if bs.Count > 0 {
				countColor = "gold"
			}
			fmt.Fprintf(&sb, " %s [gold::b]%s[-] [%s]x%d[-]\n", icon, bs.Name, countColor, bs.Count)
			if bs.AtMaxCount {
				fmt.Fprintf(&sb, "   [yellow]Building limit reached.[-]\n")
			} else {
				fmt.Fprintf(&sb, "   Cost: %s\n", FormatCost(bs.NextCost))
			}
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", bs.Description)
			if bs.Flavor != "" {
				fmt.Fprintf(&sb, "   [gray::i]%s[-:-:-]\n", bs.Flavor)
			}
			if bs.WorkerCapacity > 0 {
				totalCap := bs.Count * bs.WorkerCapacity
				bar := workerAssignBar(bs.WorkersAssigned, totalCap)
				// NOTE: Wrapping the bar in literal U+005B/U+005D (square brackets) prevents
				// tview from interpreting the block-fill characters inside as a color tag.
				barStr := "\u005b" + bar + "\u005d"
				domainLabel := domainToLabel[bs.WorkerDomain]
				if domainLabel == "" {
					domainLabel = bs.WorkerDomain
				}
				fmt.Fprintf(&sb, "   [green]Workers:[-] %d / %d %s  %s\n",
					bs.WorkersAssigned, totalCap, domainLabel, barStr)
			}
			// Pending player-driven upgrade indicator
			if bs.PendingUpgrade != "" {
				newBS := state.Buildings[bs.PendingUpgrade]
				newName := newBS.Name
				if newName == "" {
					newName = bs.PendingUpgrade
				}
				fmt.Fprintf(&sb, "   [gold]↑ Upgrade available → %s  type: upgrade %s[-]\n", newName, key)
			}
		}
		sb.WriteString("\n")
	}

	if sb.Len() == 0 {
		sb.WriteString(" [gray]No buildings unlocked yet[-]")
	}

	// Early-game onboarding: fill the otherwise-empty lower Buildings space with a
	// first-steps guide (re-homed out of the startup log). Auto-retires once the
	// player advances past the early ages. Colors are deliberately legible — gold
	// header + cyan command names + white prose — not muddy gray.
	// Auto-retire the onboarding block once the player is past the early ages OR has
	// logged 30+ minutes of play. PlayTime is wall-clock (time.Since(GameStarted)),
	// so it's robust against tick-speed bonuses that a raw tick count would inflate.
	if isEarlyGame(state.Age) && state.Stats.PlayTime < 30*time.Minute {
		sb.WriteString("\n [gold]─── Getting Started ───[-]\n")
		sb.WriteString(" [white]New here? Walk the first steps:[-]\n")
		sb.WriteString(" [white]1.[-] [cyan]gather wood 5[-] [white]/[-] [cyan]gather food[-] — collect by hand\n")
		sb.WriteString(" [white]2.[-] [cyan]build hut[-] — shelter; raises your population cap\n")
		sb.WriteString(" [white]3.[-] [cyan]recruit worker[-] — bring in your first worker\n")
		sb.WriteString(" [white]4.[-] [cyan]assign gathering_camp 3[-] — put workers to work\n")
		sb.WriteString(" [white]5.[-] [cyan]build[-] the age's [gold]wonder[-] — unlocks a speed bonus\n")
		sb.WriteString(" [white]Type[-] [cyan]help[-] [white]for all commands.[-]\n")
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
// Filled portion uses BarFillColor (Accent role), empty uses BarEmptyColor (Dim role).
func workerAssignBar(assigned, capacity int) string {
	const width = 10
	if capacity <= 0 {
		return BarEmptyColor() + strings.Repeat("░", width) + "[-]"
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
	return BarFillColor() + strings.Repeat("▓", filled) + BarEmptyColor() + strings.Repeat("░", empty) + "[-]"
}

// resourceBar returns a width-char bar using █/░ characters with color based on fill level.
// ratio >= 0.95 → gold (at cap), >= 0.60 → green, >= 0.30 → yellow, < 0.30 → red.
func resourceBar(amount, storage float64, width int) string {
	if storage <= 0 {
		return strings.Repeat("░", width)
	}
	ratio := amount / storage
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	empty := width - filled

	var fillColor string
	switch {
	case ratio >= 0.95:
		fillColor = "gold"
	case ratio >= 0.60:
		fillColor = "green"
	case ratio >= 0.30:
		fillColor = "yellow"
	default:
		fillColor = "red"
	}

	bar := ""
	if filled > 0 {
		bar += fmt.Sprintf("[%s]%s[-]", fillColor, strings.Repeat("█", filled))
	}
	if empty > 0 {
		bar += fmt.Sprintf("[gray]%s[-]", strings.Repeat("░", empty))
	}
	return bar
}

func (t *EconomyTab) refreshUnderConstruction(state game.GameState) {
	var sb strings.Builder

	if len(state.BuildQueue) == 0 {
		sb.WriteString(" [gray](nothing under construction)[-]\n")
	} else {
		// Group queue items by name
		type queueGroup struct {
			name       string
			count      int
			minTicks   int // fewest ticks left among the group (furthest along)
			totalTicks int
		}
		groupMap := make(map[string]*queueGroup)
		groupOrder := make([]string, 0)
		for _, item := range state.BuildQueue {
			if g, ok := groupMap[item.Name]; ok {
				g.count++
				if item.TicksLeft < g.minTicks {
					g.minTicks = item.TicksLeft
					g.totalTicks = item.TotalTicks
				}
			} else {
				groupMap[item.Name] = &queueGroup{
					name:       item.Name,
					count:      1,
					minTicks:   item.TicksLeft,
					totalTicks: item.TotalTicks,
				}
				groupOrder = append(groupOrder, item.Name)
			}
		}
		const barWidth = 20
		for _, name := range groupOrder {
			g := groupMap[name]
			label := g.name
			if g.count > 1 {
				label = fmt.Sprintf("%s x%d", g.name, g.count)
			}
			filled := 0
			if g.totalTicks > 0 {
				ratio := float64(g.totalTicks-g.minTicks) / float64(g.totalTicks)
				filled = int(math.Round(float64(barWidth) * ratio))
				if filled < 0 {
					filled = 0
				}
				if filled > barWidth {
					filled = barWidth
				}
			}
			bar := BarFillColor() + strings.Repeat("█", filled) + BarEmptyColor() + strings.Repeat("░", barWidth-filled) + "[-]"
			fmt.Fprintf(&sb, " [yellow]%-22s[-] %s [gray]%s left[-]\n", label, bar, formatTicks(g.minTicks, state))
		}
	}

	t.constructionTV.SetText(sb.String())
}
