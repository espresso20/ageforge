package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/rivo/tview"

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

// onboardingStep is one line of the first-steps guide: the commands to type,
// the word that joins them, and a short note on why.
type onboardingStep struct {
	Commands []string
	Join     string // "/", "and", "then"; unused for a single command
	Note     string
}

// onboardingSteps is the first-steps guide in the Buildings panel. It is data
// so onboarding_test.go can run every command through HandleCommand on a
// fresh game: a step that teaches a failing command must break the build.
var onboardingSteps = []onboardingStep{
	{Commands: []string{"gather wood 5", "gather food"}, Join: "/", Note: "collect by hand"},
	{Commands: []string{"build gathering_camp", "build wood_camp"}, Join: "and", Note: "food runs short first"},
	{Commands: []string{"build hut"}, Note: "shelter; raises your housing"},
	{Commands: []string{"workers"}, Note: "workers come on their own and staff the camps (staffed camps make 5x)"},
	{Commands: []string{"wonder collect all"}, Note: "then build the [gold]wonder[-] once its bank is full. You need it to advance."},
}

// onboardingCommands flattens onboardingSteps into the commands in the order a
// player would type them.
func onboardingCommands() []string {
	var out []string
	for _, st := range onboardingSteps {
		out = append(out, st.Commands...)
	}
	return out
}

// renderOnboarding draws the first-steps guide: gold header, cyan commands,
// white prose.
func renderOnboarding() string {
	var sb strings.Builder
	sb.WriteString("\n [gold]─── Getting started ───[-]\n")
	sb.WriteString(" [white]First steps:[-]\n")
	for i, st := range onboardingSteps {
		fmt.Fprintf(&sb, " [white]%d.[-] ", i+1)
		for j, c := range st.Commands {
			if j > 0 {
				fmt.Fprintf(&sb, " [white]%s[-] ", st.Join)
			}
			fmt.Fprintf(&sb, "[cyan]%s[-]", c)
		}
		sep := ": "
		if strings.HasPrefix(st.Note, "then ") {
			sep = ", "
		}
		fmt.Fprintf(&sb, "%s%s\n", sep, st.Note)
	}
	sb.WriteString(" [white]Type[-] [cyan]help[-] [white]for all commands.[-]\n")
	return sb.String()
}

// cultureThresholds defines the culture breakpoints at which rewards unlock.
// The bar displayed in the Economy panel measures progress towards the next threshold.
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
		midPart = fmt.Sprintf("[gold]✦ Culture mastered[-]  %-8s", FormatNumber(amount))
	} else {
		threshold := cultureThresholds[nextIdx]
		bar := cultureProgressBar(amount, threshold)
		label := cultureThresholdLabels[nextIdx]
		// Wrap bar in literal [ ] so tview does not interpret the block chars as a color tag.
		midPart = "\u005b" + bar + "\u005d" + fmt.Sprintf("  %s / %s  [gray]%s[-]",
			FormatNumber(amount), FormatNumber(threshold), label)
	}

	return fmt.Sprintf(" %-12s %s %s\n\n", rs.Name, midPart, FormatRateTick(rs.Rate))
}

// faithBand describes a faith strength with its label and epoch odds text.
type faithBand struct {
	label     string // tview-tagged label
	epochOdds string // e.g. "40% good"
}

// faithBandFor labels the faith strength in o. The odds are the engine's for
// the band the rolls read (o.FaithBand); the label only splits the bands
// finer.
func faithBandFor(o game.CatastropheOutlook) faithBand {
	odds := fmt.Sprintf("%.0f%% good", game.EpochGoodChanceIn(o.FaithBand)*100)
	strength := o.FaithStrength
	switch {
	case o.FaithFull > 0 && strength <= 0:
		return faithBand{"[red]✝ No faith[-]", odds}
	case o.FaithBand == game.FaithBandLow:
		return faithBand{"[gray]◈ Dim faith[-]", odds}
	case o.FaithBand == game.FaithBandHigh && strength >= 1:
		return faithBand{"[gold]✦ Faith full[-]", odds}
	case o.FaithBand == game.FaithBandHigh:
		return faithBand{"[green]◈ Strong faith[-]", odds}
	case strength <= 0.50:
		return faithBand{"[white]◈ Low faith[-]", odds}
	}
	return faithBand{"[yellow]◈ Faith[-]", odds}
}

// formatFaithRow builds the faith resource row string: the bar and the
// percentage are the faith strength the rolls read (o.FaithStrength, the
// faith held against what a moderate faith economy makes in three fifths of
// the age), not the fill of the store faith is kept in.
func formatFaithRow(rs game.ResourceState, o game.CatastropheOutlook) string {
	band := faithBandFor(o)
	pctStr := fmt.Sprintf("%.0f%%", o.FaithStrength*100)

	// Build the bar using the same cultureProgressBar helper (▓/░, width 10).
	bar := cultureProgressBar(o.FaithStrength, 1)
	barStr := "\u005b" + bar + "\u005d"

	midPart := fmt.Sprintf("%s  %s  %s  [gray](epoch: %s)[-]",
		barStr, band.label, pctStr, band.epochOdds)

	return fmt.Sprintf(" %-12s %s %s\n\n", rs.Name, midPart, FormatRateTick(rs.Rate))
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
	t.constructionTV.SetBorder(true).SetTitle(" Under construction ")

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

// WrapBuildings replaces the Buildings list in the layout with wrap(list),
// so the Dashboard can dock the mini map above it (mapDock).
func (t *EconomyTab) WrapBuildings(wrap func(list tview.Primitive) tview.Primitive) {
	t.root.RemoveItem(t.buildingTV)
	t.root.AddItem(wrap(t.buildingTV), 0, 1, false)
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

// resourceLegend explains the glyphs and amount colors in the resource rows,
// so neither carries meaning by color alone.
const resourceLegend = " [gray]Amount:[-] [green]rising[-] [gray]·[-] [red]falling[-] [gray]·[-] [yellow]90%+ full[-] [gray]·[-] [gold]◈[-] [gray]95%+ full ·[-] [red]▼[-] [gray]falling[-]\n"

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
			sb.WriteString(formatFaithRow(rs, state.CatastropheOutlook))
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
				rateColor, FormatRateTick(rs.Rate), bar, glyph)
		}
	}
	sb.WriteString(resourceLegend)
	t.resourceTV.SetText(safeTags(sb.String()))
}

func (t *EconomyTab) refreshBuildings(state game.GameState) {
	// Age ordering from the snapshot's ruleset, used to sort building groups
	// chronologically.
	set := state.Ruleset()
	ageIndex := set.Indexes()

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
		if def, ok := set.Age(ageKey); ok {
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
				fmt.Fprintf(&sb, "   [gold]↑ Upgrade available: %s. Type: upgrade %s[-]\n", newName, key)
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
		sb.WriteString(renderOnboarding())
	}

	t.buildingTV.SetText(safeTags(sb.String()))
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

	t.constructionTV.SetText(safeTags(sb.String()))
}
