package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
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

// renderOnboarding draws the first-steps guide for a list w cells wide:
// gold header, cyan commands, white prose, each step wrapped under itself.
func renderOnboarding(w int) string {
	var sb strings.Builder
	sb.WriteString("\n [gold]─── Getting started ───[-]\n")
	sb.WriteString(" [white]First steps:[-]\n")
	for i, st := range onboardingSteps {
		var text strings.Builder
		for j, c := range st.Commands {
			if j > 0 {
				fmt.Fprintf(&text, " [white]%s[-] ", st.Join)
			}
			// A command stays whole on its line: its spaces do not break.
			fmt.Fprintf(&text, "[cyan]%s[-]", strings.ReplaceAll(c, " ", glue))
		}
		sep := ": "
		if strings.HasPrefix(st.Note, "then ") {
			sep = ", "
		}
		text.WriteString(sep + st.Note)
		sb.WriteString(hangingRow(fmt.Sprintf(" [white]%d.[-] ", i+1), text.String(), w))
	}
	sb.WriteString(hangingRow(" ", "[white]Type[-] [cyan]help[-] [white]for all commands.[-]", w))
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

// EconomyTab is the permanent background panel visible at all times on the Dashboard.
// It is split into three panes: resource summary (left-top), under construction
// queue (left-bottom), and the full building list (right). The building panel is
// scrollable via PgUp/PgDn because it can grow very long in late ages.
//
// Each pane writes its text for the size it is drawn at (fitView), so no row
// is ever left for the terminal to wrap: see resources_box.go for the
// Resources box, and layoutConstruction and buildingLines below.
type EconomyTab struct {
	root           *tview.Flex
	leftCol        *econColumn
	resourceTV     *fitView
	buildingTV     *fitView
	constructionTV *fitView

	state game.GameState
	// resPage is the page of the Resources box on show (Ctrl+R turns it),
	// and resPages how many its rows took at the last draw.
	resPage, resPages int
	// bldFirst is the first entry of the Buildings list on show (PgUp and
	// PgDn move it a page of whole entries), bldShown how many entries the
	// last draw showed and bldRows how many rows it had for them.
	bldFirst, bldShown, bldRows int
}

// NewEconomyTab constructs the economy tab widget tree: the left column
// (econColumn: resources, construction and, once the Dashboard adds it, the
// log) and the building list, side by side at 1:1.
func NewEconomyTab() *EconomyTab {
	t := &EconomyTab{resPages: 1}

	t.resourceTV = newFitView(func(w, h int) string {
		lines, pages := layoutResourceBox(resourceRows(t.state), w, h, t.resPage)
		t.resPages = pages
		t.resPage = ((t.resPage % pages) + pages) % pages
		return strings.Join(lines, "\n")
	})
	t.resourceTV.SetBorder(true).SetTitle(" Resources ")

	t.buildingTV = newFitView(func(w, h int) string {
		head, entries := buildingEntries(t.state, w)
		lines, first, shown := layoutBuildings(head, entries, w, h, t.bldFirst)
		t.bldFirst, t.bldShown, t.bldRows = first, shown, h
		return strings.Join(lines, "\n")
	})
	t.buildingTV.SetScrollable(false)
	t.buildingTV.SetBorder(true).SetTitle(" Buildings ")

	t.constructionTV = newFitView(func(w, h int) string {
		return strings.Join(layoutConstruction(t.state, w, h), "\n")
	})
	t.constructionTV.SetBorder(true).SetTitle(" Under construction ")

	// Persistent tab chrome: enroll titles so a live theme switch restyles them.
	theme.Track(func() {
		t.resourceTV.SetTitleColor(theme.Color(theme.RoleLabel))
		t.buildingTV.SetTitleColor(theme.Color(theme.RoleHighlight))
		t.constructionTV.SetTitleColor(theme.Color(theme.RoleHighlight))
	})

	t.leftCol = &econColumn{Box: tview.NewBox(), resources: t.resourceTV, construction: t.constructionTV,
		resourceRows: func() int { return resourceBoxNeeds(len(resourceRows(t.state))) }}

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.leftCol, 0, 1, false).
		AddItem(t.buildingTV, 0, 1, false)

	return t
}

// AddToLeftColumn puts the Dashboard's log panel at the foot of the left column.
func (t *EconomyTab) AddToLeftColumn(item tview.Primitive) {
	t.leftCol.log = item
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

// Refresh takes a new snapshot: the three panes write themselves from it at
// their next draw.
func (t *EconomyTab) Refresh(state game.GameState) {
	t.state = state
	t.resourceTV.changed()
	t.buildingTV.changed()
	t.constructionTV.changed()
}

// NextResourcePage turns the Resources box to its next page (round to the
// first after the last) and reports whether it has more than one.
func (t *EconomyTab) NextResourcePage() bool {
	if t.resPages < 2 {
		return false
	}
	t.resPage = (t.resPage + 1) % t.resPages
	t.resourceTV.changed()
	return true
}

// econColumn is the dashboard's left column: the Resources box, Under
// construction and the log, stacked. It shares its height out 3:1:2, as the
// Flex it replaces did, and gives the Resources box more when its rows do
// not fit that share, as far as a floor for the two boxes under it: late
// ages hold many more resources than the first ones.
type econColumn struct {
	*tview.Box
	resources, construction tview.Primitive
	log                     tview.Primitive
	// resourceRows is how many rows the Resources box needs inside to show
	// everything.
	resourceRows func() int
}

// econHeights shares h rows out between the Resources box, Under
// construction and the log. want is the height the Resources box needs
// (border included) to show every row and its legend.
func econHeights(h, want int) (resources, construction, log int) {
	// The least the two under it keep when the Resources box needs the
	// room: a line of construction and three of log.
	const minConstruction, minLog = 3, 5
	// Half the column is the Resources box's, as it always was; more when
	// its rows do not fit that.
	resources = h * 3 / 6
	if most := h - minConstruction - minLog; want > resources && most > resources {
		resources = min(want, most)
	}
	rest := h - resources
	construction = min(rest, max(minConstruction, rest/3))
	return resources, construction, rest - construction
}

func (c *econColumn) Draw(screen tcell.Screen) {
	x, y, w, h := c.GetRect()
	rh, ch, lh := econHeights(h, c.resourceRows()+2)
	if c.log == nil {
		ch, lh = h-rh, 0
	}
	c.resources.SetRect(x, y, w, rh)
	c.resources.Draw(screen)
	c.construction.SetRect(x, y+rh, w, ch)
	c.construction.Draw(screen)
	if c.log != nil {
		c.log.SetRect(x, y+rh+ch, w, lh)
		c.log.Draw(screen)
	}
}

// layoutConstruction lays the Under construction box out for an inner area
// w cells wide and h rows tall: one line per building under way (copies of
// one building share a line), its bar and the time left. A narrow box
// shortens the bar, then drops it, then drops the word "left", then cuts
// the name; a short one ends with how many more are under way.
func layoutConstruction(state game.GameState, w, h int) []string {
	if w < 1 || h < 1 {
		return nil
	}
	if len(state.BuildQueue) == 0 {
		for _, s := range []string{" (nothing under construction)", " (nothing building)", " (none)"} {
			if runeLen(s) <= w {
				return []string{theme.Paint(theme.RoleDim, s)}
			}
		}
		return []string{""}
	}
	// Group queue items by name
	type queueGroup struct {
		label      string
		minTicks   int // fewest ticks left among the group (furthest along)
		totalTicks int
		count      int
	}
	groups := make(map[string]*queueGroup)
	var order []*queueGroup
	for _, item := range state.BuildQueue {
		if g, ok := groups[item.Name]; ok {
			g.count++
			if item.TicksLeft < g.minTicks {
				g.minTicks, g.totalTicks = item.TicksLeft, item.TotalTicks
			}
		} else {
			g = &queueGroup{label: item.Name, count: 1, minTicks: item.TicksLeft, totalTicks: item.TotalTicks}
			groups[item.Name] = g
			order = append(order, g)
		}
	}
	labelW, timeW := 0, 0
	times := make([]string, len(order))
	for i, g := range order {
		if g.count > 1 {
			g.label = fmt.Sprintf("%s x%d", g.label, g.count)
		}
		times[i] = formatTicks(g.minTicks, state)
		labelW, timeW = max(labelW, runeLen(g.label)), max(timeW, runeLen(times[i]))
	}
	const barMax, barMin = 20, 6
	left := " left"
	bar := min(barMax, w-(1+labelW+1+1+timeW+runeLen(left)))
	if bar < barMin {
		left = ""
		if bar = min(barMax, w-(1+labelW+1+1+timeW)); bar < barMin {
			bar = 0
			if 1+labelW+2+timeW+len(" left") <= w {
				left = " left"
			}
			if over := 1 + labelW + 2 + timeW - w; over > 0 {
				labelW = max(3, labelW-over)
			}
		}
	}
	lines := make([]string, 0, len(order))
	for i, g := range order {
		label := truncate(g.label, labelW)
		line := " [yellow]" + label + "[-]" + strings.Repeat(" ", labelW-runeLen(label))
		if bar > 0 {
			filled := 0
			if g.totalTicks > 0 {
				ratio := float64(g.totalTicks-g.minTicks) / float64(g.totalTicks)
				filled = min(bar, max(0, int(math.Round(float64(bar)*ratio))))
			}
			line += " " + BarFillColor() + strings.Repeat("█", filled) + BarEmptyColor() + strings.Repeat("░", bar-filled) + "[-] "
		} else {
			line += "  "
		}
		line += "[gray]" + strings.Repeat(" ", timeW-runeLen(times[i])) + times[i] + left + "[-]"
		lines = append(lines, line)
	}
	switch {
	case len(lines) <= h:
	case h == 1:
		// One line for several buildings: how many, and when the first is done.
		soonest := 0
		for i, g := range order {
			if g.minTicks < order[soonest].minTicks {
				soonest = i
			}
		}
		for _, s := range []string{
			fmt.Sprintf(" %d under way, the first in %s", len(state.BuildQueue), times[soonest]),
			fmt.Sprintf(" %d under way, %s", len(state.BuildQueue), times[soonest]),
			fmt.Sprintf(" %d under way", len(state.BuildQueue)),
		} {
			if lines = []string{theme.Paint(theme.RoleHighlight, s)}; runeLen(s) <= w {
				break
			}
		}
	default:
		more := len(lines) - (h - 1)
		lines = append(lines[:h-1], theme.Paint(theme.RoleDim, truncate(fmt.Sprintf(" +%d more under way", more), w)))
	}
	return lines
}

// buildingLines writes the whole Buildings list for a list w cells wide,
// every entry of it (the box itself shows a page of whole entries:
// layoutBuildings).
func buildingLines(state game.GameState, w int) string {
	head, entries := buildingEntries(state, w)
	var lines []string
	if head != "" {
		lines = append(lines, head)
	}
	for _, e := range entries {
		lines = append(lines, e...)
	}
	return strings.Join(lines, "\n")
}

// buildingEntries writes the Buildings list for a list w cells wide as its
// heading and its entries: this age's buildings, each with its cost, what
// it does, its flavor line and its worker slots, every line broken on
// purpose and indented under its own start (a cost breaks between its
// parts, never inside one). The first-steps guide, while it shows, is the
// last entry. An entry is the lines of one building: the box shows it
// whole or not at all.
func buildingEntries(state game.GameState, w int) (head string, entries [][]string) {
	if w < 8 {
		return "", nil
	}
	set := state.Ruleset()
	entry := func(text string) {
		entries = append(entries, strings.Split(strings.TrimRight(text, "\n"), "\n"))
	}

	// Only the current age's buildings.
	keys := make([]string, 0)
	for key, bs := range state.Buildings {
		if bs.Unlocked && bs.AgeKey == state.Age {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	if len(keys) > 0 {
		ageName := state.Age
		if def, ok := set.Age(state.Age); ok {
			ageName = def.Name
		}
		head = fmt.Sprintf(" [gold]── %s ──[-]", truncate(ageName, w-7))
		wall := &storeWall{state: state}
		for _, key := range keys {
			var sb strings.Builder
			bs := state.Buildings[key]
			icon, iconW := "[red]✗[-]", 1
			switch {
			case bs.AtMaxCount:
				icon, iconW = "[yellow]MAX[-]", 3
			case bs.CanBuild:
				icon = "[green]✓[-]"
			}
			countColor := "gray"
			if bs.Count > 0 {
				countColor = "gold"
			}
			count := fmt.Sprintf("x%d", bs.Count)
			name := truncate(bs.Name, max(4, w-(1+iconW+1+1+runeLen(count))))
			fmt.Fprintf(&sb, " %s [gold::b]%s[-:-:-] [%s]%s[-]\n", icon, name, countColor, count)
			if bs.AtMaxCount {
				sb.WriteString(hangingRow("   ", "[yellow]Building limit reached.[-]", w))
			} else {
				sb.WriteString(costRow("   Cost: ", bs.NextCost, w))
				if line := wall.line(bs); line != "" {
					sb.WriteString(hangingRow("   ", "[yellow]"+line+"[-]", w))
				}
			}
			sb.WriteString(hangingRow("   ", "[gray]"+glueRates(bs.Description)+"[-]", w))
			if bs.Flavor != "" {
				sb.WriteString(hangingRow("   ", "[gray::i]"+bs.Flavor+"[-:-:-]", w))
			}
			if bs.WorkerCapacity > 0 {
				totalCap := bs.Count * bs.WorkerCapacity
				domainLabel := domainToLabel[bs.WorkerDomain]
				if domainLabel == "" {
					domainLabel = textfmt.Capitalize(bs.WorkerDomain)
				}
				// Slots filled over slots, the domain, and a bar in brackets
				// when there is room for one worth drawing. A narrow list
				// closes the count up, then leaves the domain out.
				const lead = 12 // "   Workers: "
				text := fmt.Sprintf("%d / %d %s", bs.WorkersAssigned, totalCap, domainLabel)
				if lead+runeLen(text) > w {
					text = fmt.Sprintf("%d/%d %s", bs.WorkersAssigned, totalCap, domainLabel)
				}
				if lead+runeLen(text) > w {
					text = truncate(fmt.Sprintf("%d/%d", bs.WorkersAssigned, totalCap), max(1, w-lead))
				}
				line := "   [green]Workers:[-] " + text
				if cells := min(10, w-lead-runeLen(text)-4); cells >= 4 {
					line += "  [" + workerAssignBar(bs.WorkersAssigned, totalCap, cells) + "]"
				}
				sb.WriteString(line + "\n")
			}
			// Pending player-driven upgrade indicator
			if bs.PendingUpgrade != "" {
				newName := state.Buildings[bs.PendingUpgrade].Name
				if newName == "" {
					newName = bs.PendingUpgrade
				}
				sb.WriteString(hangingRow("   ", fmt.Sprintf("[gold]↑ Upgrade available: %s. Type: upgrade %s[-]", newName, key), w))
			}
			entry(sb.String())
		}
	}

	if len(entries) == 0 {
		head = " [gray]No buildings unlocked yet[-]"
	}

	// Early-game onboarding: fill the otherwise-empty lower Buildings space with a
	// first-steps guide (re-homed out of the startup log). Auto-retires once the
	// player advances past the early ages. Colors are deliberately legible — gold
	// header + cyan command names + white prose — not muddy gray.
	// Auto-retire the onboarding block once the player is past the early ages OR has
	// logged 30+ minutes of play. PlayTime is wall-clock (time.Since(GameStarted)),
	// so it's robust against tick-speed bonuses that a raw tick count would inflate.
	if isEarlyGame(state.Age) && state.Stats.PlayTime < 30*time.Minute {
		entry(renderOnboarding(w))
	}
	return head, entries
}

// layoutBuildings lays the Buildings list out in a box w cells wide and h
// rows high, from entry first on: the heading, then as many whole entries
// as fit. An entry is never cut: one that does not fit whole is left for
// the next page, and the box says how many are above and below and which
// keys turn the page. It returns the lines, the first entry it really
// started from (first is brought into range) and how many it showed. A box
// too short for even one entry shows the top of that entry, as much as
// fits: the one case an entry is cut.
func layoutBuildings(head string, entries [][]string, w, h, first int) (lines []string, from, shown int) {
	if head != "" {
		lines = append(lines, head)
	}
	total := 0
	for _, e := range entries {
		total += len(e)
	}
	if len(lines)+total <= h || h <= 0 {
		// It all fits (or the box has no size yet): the whole list.
		for _, e := range entries {
			lines = append(lines, e...)
		}
		return lines, 0, len(entries)
	}
	from = min(max(first, 0), len(entries)-1)
	room := h - len(lines)
	if from > 0 {
		room-- // the line that says what is above
	}
	end := from
	for end < len(entries) {
		need := len(entries[end])
		if end < len(entries)-1 {
			need++ // the line that says what is below
		}
		if need > room {
			break
		}
		room -= len(entries[end])
		end++
	}
	if from > 0 {
		lines = append(lines, theme.Paint(theme.RoleDim, truncate(fmt.Sprintf(" ↑ %d more above: PgUp", from), w)))
	}
	if end == from {
		// Not even one entry fits whole: show the top of it.
		cut := max(h-len(lines)-1, 0)
		lines = append(lines, entries[from][:min(cut, len(entries[from]))]...)
		end = from + 1
	} else {
		for _, e := range entries[from:end] {
			lines = append(lines, e...)
		}
	}
	if below := len(entries) - end; below > 0 {
		lines = append(lines, theme.Paint(theme.RoleDim, truncate(fmt.Sprintf(" ↓ %d more below: PgDn", below), w)))
	}
	if len(lines) > h {
		lines = lines[:h] // a box of a row or two has no room for the notes
	}
	return lines, from, end - from
}

// costRow writes a cost under lead ("   Cost: "), wrapped to w cells with
// each further line starting under the first part. A part ("1.12K iron")
// is never broken.
func costRow(lead string, cost map[string]float64, w int) string {
	parts := strings.Split(FormatCost(cost), ", ")
	indent := runeLen(lead)
	var lines []string
	cur := ""
	for i, p := range parts {
		if i < len(parts)-1 {
			p += ","
		}
		switch {
		case cur == "":
			cur = p
		case indent+runeLen(cur)+1+runeLen(p) <= w:
			cur += " " + p
		default:
			lines = append(lines, cur)
			cur = p
		}
	}
	lines = append(lines, cur)
	return lead + strings.Join(lines, "\n"+strings.Repeat(" ", indent)) + "\n"
}

// ScrollUp turns the Buildings list back a page of whole entries.
func (t *EconomyTab) ScrollUp() {
	if t.bldFirst <= 0 {
		return
	}
	// The page that ends just above the one on show: as many entries as
	// fit in the box, counted back from it.
	_, _, w, _ := t.buildingTV.GetInnerRect()
	_, entries := buildingEntries(t.state, w)
	first := min(t.bldFirst, len(entries))
	room := t.bldRows - 3 // the heading and the two lines that say what is above and below
	for first > 0 && len(entries[first-1]) <= room {
		room -= len(entries[first-1])
		first--
	}
	if first == t.bldFirst {
		first--
	}
	t.bldFirst = max(first, 0)
	t.buildingTV.changed()
}

// ScrollDown turns the Buildings list on a page of whole entries.
func (t *EconomyTab) ScrollDown() {
	_, _, w, _ := t.buildingTV.GetInnerRect()
	_, entries := buildingEntries(t.state, w)
	if t.bldFirst+t.bldShown >= len(entries) {
		return // the last entry is on show
	}
	t.bldFirst += max(t.bldShown, 1)
	t.buildingTV.changed()
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

// workerAssignBar returns a ▓/░ bar width cells wide for assigned/capacity.
// Filled portion uses BarFillColor (Accent role), empty uses BarEmptyColor (Dim role).
func workerAssignBar(assigned, capacity, width int) string {
	if capacity <= 0 {
		return BarEmptyColor() + strings.Repeat("░", width) + "[-]"
	}
	return progressCells(float64(assigned)/float64(capacity), width)
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
	if ratio < 0 {
		ratio = 0
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
