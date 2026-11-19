package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// cultureThresholds defines the culture breakpoints at which rewards unlock.
// The bar displayed in the Economy panel measures progress towards the next threshold.
var cultureThresholds = []float64{
	500, 2500, 10000, 50000, 250000, config.Million, 5 * config.Million, 25 * config.Million, 100 * config.Million, 500 * config.Million, config.Billion,
}

// txtResearchSpeed is said the same way on several screens, so it has one id.
var txtResearchSpeed = rowid.E("ui.tab_economy.cultureThresholdLabels.07", "+research speed")

// cultureThresholdLabels are short effect labels for each threshold (parallel to cultureThresholds).
var cultureThresholdLabels = []rowid.Entry{
	rowid.E("ui.tab_economy.cultureThresholdLabels.01", "+5% research speed"),
	rowid.E("ui.tab_economy.cultureThresholdLabels.02", "+10% research speed"),
	rowid.E("ui.tab_economy.cultureThresholdLabels.03", "+15% research speed, unlock wonder tier"),
	rowid.E("ui.tab_economy.cultureThresholdLabels.04", "+20% research speed"),
	rowid.E("ui.tab_economy.cultureThresholdLabels.05", "+25% research speed, culture events"),
	rowid.E("ui.tab_economy.cultureThresholdLabels.06", "+30% research speed"),
	txtResearchSpeed,
	txtResearchSpeed,
	txtResearchSpeed,
	txtResearchSpeed,
	txtResearchSpeed,
}

// EconomyTab is the permanent background panel visible at all times on the Dashboard.
// It is split into three panes: resource summary (left-top), under construction
// queue (left-bottom), and the build cards (right), which dash_cards_box.go
// draws. The cards are scrollable via the wheel and PgUp/PgDn because the list
// can grow long in late ages.
//
// The two text panes write their text for the size they are drawn at
// (fitView), so no row is ever left for the terminal to wrap: see
// resources_box.go for the Resources box, and layoutConstruction below.
type EconomyTab struct {
	root           *nativeGroup
	leftCol        *econColumn
	resourceTV     *resourceView
	buildBox       *buildBox
	constructionTV *fitView

	state game.GameState
	// trend is the amounts of the last few minutes, for the Resources box's
	// trend lines; plain says whether to draw in ASCII; now is the clock.
	trend resHistory
	plain func() bool
	now   func() time.Time
	// resPage is the page of the Resources box on show (Ctrl+R turns it),
	// and resPages how many its rows took at the last draw.
	resPage, resPages int
	// resKeys is the resource each text line of the box belongs to ("" for
	// none), as of the last draw.
	resKeys []string
}

// txtResourcesTitle is said the same way on several screens, so it has one id.
var txtResourcesTitle = rowid.E("ui.tab_economy.NewEconomyTab.01", " Resources ")

// retitle writes the titles on the borders of the tab's three boxes, in the
// script now in force.
func (t *EconomyTab) retitle() {
	t.resourceTV.SetTitle(txtResourcesTitle.String())
	t.buildBox.title = rowid.Line("ui.tab_economy.NewEconomyTab.02", " Buildings ")
	t.buildBox.dirty = true
	t.constructionTV.SetTitle(rowid.Line("ui.tab_economy.NewEconomyTab.03", " Under construction "))
}

// NewEconomyTab constructs the economy tab widget tree: the left column
// (econColumn: resources, construction and, once the Dashboard adds it, the
// log) and the build cards, side by side (econLeftW shares the width).
// app is what the cards' buttons press through.
func NewEconomyTab(app *appLoop) *EconomyTab {
	t := &EconomyTab{resPages: 1}

	t.resourceTV = newResourceView(func(w, h int) string {
		rows := t.rows()
		lines, owner, pages := layoutResourceOwners(rows, w, h, t.resPage)
		t.resKeys = t.resKeys[:0]
		for _, o := range owner {
			key := ""
			if o >= 0 {
				key = rows[o].key
			}
			t.resKeys = append(t.resKeys, key)
		}
		t.resPages = pages
		t.resPage = ((t.resPage % pages) + pages) % pages
		return strings.Join(lines, "\n")
	}, func() bool { return t.plain != nil && t.plain() })
	t.resourceTV.SetBorder(true)
	t.resourceTV.tip, t.resourceTV.tipFor = &app.feel.tip, t.rateTip

	t.resourceTV.wheel = t.wheelResources
	t.buildBox = newBuildBox(func() bool { return t.plain != nil && t.plain() }, app)

	t.constructionTV = newFitView(func(w, h int) string {
		return strings.Join(layoutConstruction(t.state, w, h), "\n")
	})
	t.constructionTV.SetBorder(true)
	t.retitle()

	// Persistent tab chrome: enroll titles so a live theme switch restyles them.
	theme.Track(func() {
		t.resourceTV.SetTitleColor(theme.Of(theme.RoleLabel))
		t.constructionTV.SetTitleColor(theme.Of(theme.RoleHighlight))
	})

	t.leftCol = &econColumn{resources: t.resourceTV, construction: t.constructionTV,
		resourceRows: func() int { return resourceRowsNeed(t.rows()) }}
	t.leftCol.nativeGroup = newNativeGroup(t.leftCol.arrange)

	t.root = newNativeGroup(func(x, y, w, h int) {
		left := econLeftW(w)
		t.root.put(t.leftCol, x, y, left, h)
		t.root.put(t.buildBox, x+left, y, w-left, h)
	})

	return t
}

// econLeftMax is the left column's width in a wide layout: the Buildings box
// has the rest.
const econLeftMax = 46

// econLeftW is the left column's width in an economy tab w cells wide: half of
// it while that is under econLeftMax, as the two boxes always shared a narrow
// terminal, and econLeftMax from there up.
func econLeftW(w int) int { return min(econLeftMax, w/2) }

// AddToLeftColumn puts the Dashboard's log panel at the foot of the left column.
func (t *EconomyTab) AddToLeftColumn(item Primitive) {
	t.leftCol.log = item
}

// AddPeopleBox hands the left column the People box, which it draws in the
// right half of the Under construction row when rich says the terminal gets
// the rich layout (dash_layout.go); height is the box's height in that layout.
func (t *EconomyTab) AddPeopleBox(people Primitive, rich func() bool, height func() int) {
	t.leftCol.people, t.leftCol.rich, t.leftCol.peopleHeight = people, rich, height
}

// Root returns the root primitive
func (t *EconomyTab) Root() Primitive {
	return t.root
}

// Refresh takes a new snapshot: the three panes write themselves from it at
// their next draw.
func (t *EconomyTab) Refresh(state game.GameState) {
	t.state = state
	if t.now == nil {
		t.now = time.Now
	}
	t.trend.note(state, t.now())
	t.resourceTV.changed()
	t.buildBox.update(state)
	t.constructionTV.changed()
}

// rows is the Resources box's rows for the snapshot on show.
func (t *EconomyTab) rows() []resRow {
	return resourceRowsFor(t.state, &t.trend, t.plain != nil && t.plain())
}

// ScrollUp and ScrollDown turn the build cards a page.
func (t *EconomyTab) ScrollUp()   { t.buildBox.page(-1) }
func (t *EconomyTab) ScrollDown() { t.buildBox.page(1) }

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
//
// In the rich layout (dash_layout.go) the Under construction row is shared
// by two boxes, Under construction on the left and the People box on the
// right, and is as tall as the People box.
type econColumn struct {
	*nativeGroup
	resources, construction Primitive
	log                     Primitive
	// people is the People box, drawn in the right half of the Under
	// construction row when rich says the terminal gets the rich layout.
	people Primitive
	// rich reports the layout; peopleHeight is the People box's height,
	// border included, in it (for a terminal that tall).
	rich         func() bool
	peopleHeight func() int
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

// econHeightsRich is econHeights for the rich layout, where the row under
// the Resources box is the People box's height, people rows. The Resources
// box gives way before the People box or the log do, down to a floor.
func econHeightsRich(h, want, people int) (resources, construction, log int) {
	const minLog, minResources = 5, 6
	construction = max(3, people)
	resources = h * 3 / 6
	most := max(minResources, h-construction-minLog)
	if want > resources {
		resources = want
	}
	resources = min(resources, most)
	if construction > h-resources {
		construction = h - resources
	}
	return resources, construction, h - resources - construction
}

// heights shares a column h rows tall out between its boxes, for the layout
// it is drawn in.
func (c *econColumn) heights(h int) (resources, construction, log int) {
	want := c.resourceRows() + 2
	if c.log == nil {
		resources, _, _ = econHeights(h, want)
		return resources, h - resources, 0
	}
	if c.richLayout() {
		return econHeightsRich(h, want, c.peopleHeight())
	}
	return econHeights(h, want)
}

// richLayout reports whether the column holds the People box.
func (c *econColumn) richLayout() bool {
	return c.people != nil && c.rich != nil && c.rich()
}

func (c *econColumn) arrange(x, y, w, h int) {
	rh, ch, lh := c.heights(h)
	c.put(c.resources, x, y, w, rh)
	if c.log != nil && c.richLayout() {
		// Under construction has the wider half when the width is odd: its
		// title is 22 cells with the border.
		half := w - w/2
		c.put(c.construction, x, y+rh, half, ch)
		c.put(c.people, x+half, y+rh, w-half, ch)
	} else {
		c.put(c.construction, x, y+rh, w, ch)
	}
	if c.log != nil {
		c.put(c.log, x, y+rh+ch, w, lh)
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
		for _, s := range []string{rowid.Line("ui.tab_economy.layoutConstruction.01", " (nothing under construction)"), rowid.Line("ui.tab_economy.layoutConstruction.02", " (nothing building)"), rowid.Line("ui.tab_economy.layoutConstruction.03", " (none)")} {
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
			rowid.Sprintf("ui.tab_economy.layoutConstruction.07", " %d under way, the first in %s", len(state.BuildQueue), times[soonest]),
			rowid.Sprintf("ui.tab_economy.layoutConstruction.08", " %d under way, %s", len(state.BuildQueue), times[soonest]),
			rowid.Sprintf("ui.tab_economy.layoutConstruction.09", " %d under way", len(state.BuildQueue)),
		} {
			if lines = []string{theme.Paint(theme.RoleHighlight, s)}; runeLen(s) <= w {
				break
			}
		}
	default:
		more := len(lines) - (h - 1)
		lines = append(lines[:h-1], theme.Paint(theme.RoleDim, truncate(rowid.Sprintf("ui.tab_economy.layoutConstruction.10", " +%d more under way", more), w)))
	}
	return lines
}

// buildingPrice is the price a building screen shows for the next copy of
// key. A wonder is paid through its bank at the price in the ruleset; the
// build-cost factor that scales every other building's price does not reach
// it, so the engine's NextCost for a wonder is not what it charges. The bank,
// the wonder panel and the plan all use the ruleset's price, and so does this.
func buildingPrice(state *game.GameState, key string, bs game.BuildingState) map[string]float64 {
	if bs.Category == "wonder" {
		if def, ok := state.Ruleset().Building(key); ok {
			return def.BaseCost
		}
	}
	return bs.NextCost
}

// costRow writes a cost under lead ("   Cost: "), wrapped to w cells with
// each further line starting under the first part. A part ("1.12K iron")
// is kept whole on a line; one that is wider than the room left of the lead
// ("49.8T dark matter crystals" in a narrow box) goes on over several lines
// at its spaces, so nothing of it is cut.
func costRow(lead string, cost map[string]float64, w int) string {
	parts := strings.Split(FormatCost(cost), ", ")
	indent := runeLen(lead)
	var lines []string
	cur := ""
	// begin starts a line with p.
	begin := func(p string) {
		if runeLen(p) <= w-indent {
			cur = p
			return
		}
		chunks := wrapWords(p, w-indent)
		lines = append(lines, chunks[:len(chunks)-1]...)
		cur = chunks[len(chunks)-1]
	}
	for i, p := range parts {
		if i < len(parts)-1 {
			p += ","
		}
		switch {
		case cur == "":
			begin(p)
		case indent+runeLen(cur)+1+runeLen(p) <= w:
			cur += " " + p
		default:
			lines = append(lines, cur)
			begin(p)
		}
	}
	lines = append(lines, cur)
	return lead + strings.Join(lines, "\n"+strings.Repeat(" ", indent)) + "\n"
}

// wheelResources turns the Resources box a page for a turn of the wheel: on
// for a turn toward the player, back for one away, stopping at the first and
// the last page. It reports whether the page changed.
func (t *EconomyTab) wheelResources(rows int) bool {
	page := t.resPage
	switch {
	case rows > 0 && page < t.resPages-1:
		page++
	case rows < 0 && page > 0:
		page--
	}
	if page == t.resPage {
		return false
	}
	t.resPage = page
	t.resourceTV.changed()
	return true
}

// tipRate is the tooltip on a resource's line in the Resources box.
type tipRate struct{ key string }

// rateTip is the tooltip for line i of the Resources box: what the resource's
// rate is made of, the largest parts first, in at most the plate's rows.
func (t *EconomyTab) rateTip(i int) (any, kit.TipText, bool) {
	if i < 0 || i >= len(t.resKeys) || t.resKeys[i] == "" {
		return nil, kit.TipText{}, false
	}
	key := t.resKeys[i]
	rs := t.state.Resources[key]
	src := rateSources(rs.Breakdown, t.state.TickIntervalMs)
	if len(src) == 0 {
		return nil, kit.TipText{}, false
	}
	sort.SliceStable(src, func(a, b int) bool { return math.Abs(src[a].size) > math.Abs(src[b].size) })
	lines := []string{rs.Name + ": " + rateText(rs.Rate, t.state.TickIntervalMs)}
	shown := src
	if len(src) > kit.TipMaxRows-1 {
		shown = src[:kit.TipMaxRows-2]
	}
	for _, s := range shown {
		lines = append(lines, "  "+s.text)
	}
	if more := len(src) - len(shown); more > 0 {
		lines = append(lines, rowid.Sprintf("ui.tab_economy.rateTip.01", "  and %d more", more))
	}
	return tipRate{key}, kit.TipText{Lines: lines, Wide: true}, true
}
