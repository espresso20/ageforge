package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// resources_box.go lays the dashboard's Resources box out for the size it
// is drawn at. Every resource is one row on one grid (name, amount, rate,
// marker; culture and faith add their progress toward something: a step and
// a bar), so the columns line up whatever the widest name or rate is, and a
// row never wraps: when the box is narrow it gives up the least important
// part first (the bar shrinks, then goes; then the spacing; then the step;
// then the rate's unit shortens; then long names). When the
// box is short the rows close up, and when even that is not enough the box
// shows a page of them and says how to turn it.

const (
	// resourcePageKey turns the Resources box's page when it has more rows
	// than it can show. Ctrl+R: R for resources, a key the prompt does not
	// use and every terminal passes on.
	resourcePageKey     = tcell.KeyCtrlR
	resourcePageKeyName = "Ctrl+R"

	resBarMax  = 12 // a bar is never wider than this
	resBarGood = 8  // the box gives up spacing before it lets the bar under this
	resBarMin  = 4  // under this a bar says nothing: it goes
	resNameCap = 14 // names are shortened to this before the bar is given up
	// resNameFloor is as far as a name is cut, past resNameCap, to keep a
	// bar of resBarMin: a two-letter unit makes the amounts wider by a cell
	// or two, and a long name gives that back before the bar goes.
	resNameFloor = 12
)

// resRow is one resource's row before it is laid out.
type resRow struct {
	name     string
	amt      string // the amount column
	amtColor string
	cap      string // the step column (culture's next step); "" for the rest
	// status stands where "/ step" does on a row that has no step to show
	// (faith's strength, culture past its last step): the longest of these
	// that fits, plain text, drawn in statusColor.
	status      []string
	statusColor string
	rate        float64
	fill        float64 // how far along the bar is, 0 to 1
	// progress draws a bar of progress toward something (faith strength,
	// culture's next step). A row without it has no bar.
	progress bool
	glyph    string // the marker after the bar: "▼" falling
	glyphCol string
	detail   string // a second line for a roomy box, tags and all
}

// resourceRows is the box's rows for a snapshot: every unlocked resource, in
// key order.
func resourceRows(state game.GameState) []resRow {
	keys := make([]string, 0, len(state.Resources))
	for k, rs := range state.Resources {
		if rs.Unlocked {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	rows := make([]resRow, 0, len(keys))
	for _, key := range keys {
		rs := state.Resources[key]
		switch key {
		case "culture":
			rows = append(rows, cultureRow(rs))
		case "faith":
			rows = append(rows, faithRow(rs, state.CatastropheOutlook))
		default:
			rows = append(rows, storeRow(rs))
		}
	}
	return rows
}

// storeRow is an ordinary resource: what is held and which way it is going.
// Stock has no ceiling, so there is no cap and no bar.
func storeRow(rs game.ResourceState) resRow {
	r := resRow{name: rs.Name, amt: FormatNumber(rs.Amount), amtColor: "white", rate: rs.Rate}
	switch {
	case rs.Rate > 0:
		r.amtColor = "green"
	case rs.Rate < 0:
		r.amtColor = "red"
		r.glyph, r.glyphCol = "▼", "red"
	}
	return r
}

// cultureRow is culture: what is held over the next step it earns
// something at, the bar the way there, and what that step gives on the
// detail line.
func cultureRow(rs game.ResourceState) resRow {
	r := resRow{name: rs.Name, amt: FormatNumber(rs.Amount), amtColor: "white", rate: rs.Rate, progress: true}
	for i, step := range cultureThresholds {
		if rs.Amount < step {
			r.cap = FormatNumber(step)
			r.fill = rs.Amount / step
			r.detail = vibe.Line("ui.resources_box.cultureRow.01", "[gray]at ") + FormatNumber(step) + ": " + cultureThresholdLabels[i].String() + "[-]"
			return r
		}
	}
	r.status, r.statusColor, r.fill = []string{vibe.Line("ui.resources_box.cultureRow.02", "✦ Culture mastered"), vibe.Line("ui.resources_box.cultureRow.03", "✦ mastered"), "✦"}, "gold", 1
	return r
}

// faithBand is how the faith row names a faith strength: its color, its
// label in full, the label without the word the row's name already says,
// its mark alone, and the odds of a good epoch event in the band the rolls
// read.
type faithBand struct {
	color, label, short, mark string
	epochOdds                 string // e.g. "40% good"
}

// faithBandFor labels the faith strength in o for a town holding held faith.
// The odds are the engine's for the band the rolls read (o.FaithBand); the
// label only splits the bands finer.
func faithBandFor(held float64, o game.CatastropheOutlook) faithBand {
	odds := vibe.Sprintf("ui.resources_box.faithBandFor.01", "%.0f%% good", game.EpochGoodChanceIn(o.FaithBand)*100)
	strength := o.FaithStrength
	switch {
	case held <= 0:
		return faithBand{"red", vibe.Line("ui.resources_box.faithBandFor.02", "✝ No faith"), vibe.Line("ui.resources_box.faithBandFor.03", "✝ None"), "✝", odds}
	case o.FaithBand == game.FaithBandLow:
		return faithBand{"gray", vibe.Line("ui.resources_box.faithBandFor.04", "◈ Dim faith"), vibe.Line("ui.resources_box.faithBandFor.05", "◈ Dim"), "◈", odds}
	case o.FaithBand == game.FaithBandHigh && strength >= 1:
		return faithBand{"gold", vibe.Line("ui.resources_box.faithBandFor.06", "✦ Faith full"), vibe.Line("ui.resources_box.faithBandFor.07", "✦ Full"), "✦", odds}
	case o.FaithBand == game.FaithBandHigh:
		return faithBand{"green", vibe.Line("ui.resources_box.faithBandFor.08", "◈ Strong faith"), vibe.Line("ui.resources_box.faithBandFor.09", "◈ Strong"), "◈", odds}
	case strength <= 0.50:
		return faithBand{"white", vibe.Line("ui.resources_box.faithBandFor.10", "◈ Low faith"), vibe.Line("ui.resources_box.faithBandFor.11", "◈ Low"), "◈", odds}
	}
	return faithBand{"yellow", vibe.Line("ui.resources_box.faithBandFor.12", "◈ Faith"), vibe.Line("ui.resources_box.faithBandFor.13", "◈ Faith"), "◈", odds}
}

// faithRow is faith: the amount column and the bar are the faith strength
// the rolls read (o.FaithStrength: the faith the town's own faith buildings
// made, against four and a half moderate sets'), not the fill of the store
// faith is kept in. The band stands where a step would, and the detail line
// says it in full with the odds it gives.
func faithRow(rs game.ResourceState, o game.CatastropheOutlook) resRow {
	band := faithBandFor(rs.Amount, o)
	return resRow{
		name: rs.Name, amt: fmt.Sprintf("%.0f%%", o.FaithStrength*100), amtColor: "white",
		status: []string{band.label, band.short, band.mark}, statusColor: band.color,
		rate: rs.Rate, fill: o.FaithStrength, progress: true,
		detail: vibe.Sprintf("ui.resources_box.faithRow.01", "[%s]%s[-] [gray](epoch: %s)[-]", band.color, band.label, band.epochOdds),
	}
}

// resFormat is the grid the rows are drawn on.
type resFormat struct {
	nameW, amtW, capW, rateW int
	gap                      int    // cells between columns
	sep                      int    // cells between amount and step: 3 is " / ", 1 is "/", 0 is no step column
	bar                      int    // cells of bar, 0 for none
	unit                     string // the rate's unit: "/tick", or "/t" when even the step had to go
}

// midW is the cells between the amount and the rate columns' gaps: the
// separator and the step.
func (f resFormat) midW() int {
	if f.sep == 0 {
		return 0
	}
	return f.sep + f.capW
}

// width is the cells a row takes.
func (f resFormat) width() int {
	w := 1 + f.nameW + f.gap + f.amtW + f.midW() + f.gap + f.rateW + 2 // 2: the marker and the space before it
	if f.bar > 0 {
		w += f.gap + f.bar
	}
	return w
}

// fitResourceFormat picks the grid for rows in a box w cells wide: the
// fullest one that fits.
func fitResourceFormat(rows []resRow, w int) resFormat {
	f := resFormat{unit: "/tick", gap: 2, sep: 3}
	longest := 0
	needMid, needBar := false, false // does any row have a step or status, a bar?
	for _, r := range rows {
		needMid = needMid || r.cap != "" || len(r.status) > 0
		needBar = needBar || r.progress
		longest = max(longest, runeLen(r.name))
		f.amtW = max(f.amtW, runeLen(r.amt))
		f.capW = max(f.capW, runeLen(r.cap))
		text, _ := rateParts(r.rate, f.unit)
		f.rateW = max(f.rateW, runeLen(text))
	}
	// spacings from roomy to tight; barIn is the widest bar w leaves.
	spacings := [][2]int{{2, 3}, {1, 3}, {1, 1}}
	if !needMid {
		spacings = [][2]int{{2, 0}, {1, 0}}
	}
	barIn := func(f resFormat) int {
		f.bar = 0
		return min(resBarMax, w-f.width()-f.gap)
	}
	try := func(nameW, least int) bool {
		f.nameW = nameW
		for _, s := range spacings {
			f.gap, f.sep = s[0], s[1]
			if !needBar {
				f.bar = 0
				if f.width() <= w {
					return true
				}
				continue
			}
			if b := barIn(f); b >= least {
				f.bar = b
				return true
			}
		}
		return false
	}
	// With a good bar: names in full, else shortened. Then any bar worth
	// drawing.
	capped := min(longest, resNameCap)
	for _, least := range []int{resBarGood, resBarMin} {
		if try(longest, least) || (capped < longest && try(capped, least)) {
			return f
		}
	}
	for n := capped - 1; n >= resNameFloor && n < longest; n-- {
		if try(n, resBarMin) {
			return f
		}
	}
	// No bar: the roomiest spacing that fits, then without the step, then
	// with the short unit, then with the names cut to what is left.
	f.bar, f.nameW = 0, capped
	for _, s := range spacings {
		if f.gap, f.sep = s[0], s[1]; f.width() <= w {
			return f
		}
	}
	f.gap, f.sep = 1, 0
	if f.width() <= w {
		return f
	}
	f.unit, f.rateW = "/t", 0
	for _, r := range rows {
		text, _ := rateParts(r.rate, f.unit)
		f.rateW = max(f.rateW, runeLen(text))
	}
	if over := f.width() - w; over > 0 {
		f.nameW = max(3, f.nameW-over)
	}
	return f
}

// line draws r on the grid: one line, f.width() cells at most, tags and all.
func (f resFormat) line(r resRow) string {
	var b strings.Builder
	gap := strings.Repeat(" ", f.gap)
	name := shortName(r.name, f.nameW)
	b.WriteString(" " + name + strings.Repeat(" ", f.nameW-runeLen(name)) + gap)
	fmt.Fprintf(&b, "[%s]%s%s[-]", r.amtColor, strings.Repeat(" ", max(0, f.amtW-runeLen(r.amt))), r.amt)
	if mid := f.midW(); mid > 0 {
		switch {
		case len(r.status) > 0:
			// The longest status that fits, a space before it when there is room.
			text := ""
			for _, s := range r.status {
				if runeLen(s) < mid {
					text = " " + s
					break
				}
				if runeLen(s) == mid {
					text = s
					break
				}
			}
			fmt.Fprintf(&b, "[%s]%s[-]%s", r.statusColor, text, strings.Repeat(" ", mid-runeLen(text)))
		case r.cap == "":
			b.WriteString(strings.Repeat(" ", mid))
		case f.sep == 3:
			fmt.Fprintf(&b, "[gray] / %s[-]%s", r.cap, strings.Repeat(" ", f.capW-runeLen(r.cap)))
		default:
			fmt.Fprintf(&b, "[gray]/%s[-]%s", r.cap, strings.Repeat(" ", f.capW-runeLen(r.cap)))
		}
	}
	rate, rateColor := rateParts(r.rate, f.unit)
	fmt.Fprintf(&b, "%s%s[%s]%s[-]", gap, strings.Repeat(" ", max(0, f.rateW-runeLen(rate))), rateColor, rate)
	if f.bar > 0 {
		b.WriteString(gap)
		if r.progress {
			b.WriteString(progressCells(r.fill, f.bar))
		} else {
			b.WriteString(strings.Repeat(" ", f.bar))
		}
	}
	if r.glyph != "" {
		fmt.Fprintf(&b, " [%s]%s[-]", r.glyphCol, r.glyph)
	}
	return b.String()
}

// progressCells is a bar of width cells, ratio of it filled, in the accent
// color: progress toward something.
func progressCells(ratio float64, width int) string {
	ratio = min(1, max(0, ratio))
	filled := int(ratio * float64(width))
	return BarFillColor() + strings.Repeat("▓", filled) + BarEmptyColor() + strings.Repeat("░", width-filled) + "[-]"
}

// shortName fits a name in n cells: leading words give way to their
// initials ("Dark Matter Crystals" is "D. M. Crystals" at 14), and what is
// still too long is cut with an ellipsis.
func shortName(name string, n int) string {
	if runeLen(name) <= n {
		return name
	}
	words := strings.Fields(name)
	for i := 0; i < len(words)-1; i++ {
		words[i] = string([]rune(words[i])[:1]) + "."
		if s := strings.Join(words, " "); runeLen(s) <= n {
			return s
		}
	}
	return truncate(strings.Join(words, " "), n)
}

// resLegends are the legend under the rows, from full to brief: what the
// amount's color and the marker mean, so neither carries meaning by color
// alone. The box prints the fullest that fits.
var resLegends = []vibe.Entry{
	vibe.E("ui.resources_box.resLegends.01", " [gray]Amount:[-] [green]rising[-] [gray]·[-] [red]falling[-] [gray]·[-] [red]▼[-] [gray]falling[-]"),
	vibe.E("ui.resources_box.resLegends.02", " [green]rising[-] [gray]·[-] [red]▼[-] [gray]falling[-]"),
	vibe.E("ui.resources_box.resLegends.03", " [red]▼[-] [gray]falling[-]"),
}

func resLegend(w int) string {
	for _, e := range resLegends {
		if l := e.String(); visibleLen(l) <= w {
			return l
		}
	}
	return ""
}

// resMoreLine is the last line of a box that shows a page of its rows:
// which rows these are, and the key that turns the page.
func resMoreLine(first, last, total, w int) string {
	rng := vibe.Sprintf("ui.resources_box.resMoreLine.01", "%d-%d of %d", first, last, total)
	for _, l := range []string{
		" " + theme.Paint(theme.RoleDim, rng+" ·") + " " + theme.Keycap(resourcePageKeyName) + theme.Paint(theme.RoleDim, vibe.Line("ui.resources_box.resMoreLine.02", " next page")),
		" " + theme.Paint(theme.RoleDim, rng) + " " + theme.Keycap(resourcePageKeyName),
		" " + theme.Paint(theme.RoleDim, fmt.Sprintf("%d-%d/%d", first, last, total)) + " " + theme.Paint(theme.RoleAccent, resourcePageKeyName),
	} {
		if visibleLen(l) <= w {
			return l
		}
	}
	return " " + theme.Paint(theme.RoleAccent, resourcePageKeyName)
}

// resourceBoxNeeds is the least rows a box must have inside to show n
// resources without paging: the rows, then the legend.
func resourceBoxNeeds(n int) int { return n + 1 }

// resourceRowsNeed is the rows inside the box that show rows without paging.
func resourceRowsNeed(rows []resRow) int {
	return resourceBoxNeeds(len(rows))
}

// layoutResourceBox lays rows out for a box w cells wide and h rows tall
// inside. It returns the lines to draw (none wider than w, never more than
// h) and how many pages the rows take: 1 when they all fit. page is the
// page to show, counted from 0 and taken round the number of pages.
func layoutResourceBox(rows []resRow, w, h, page int) (lines []string, pages int) {
	if w < 1 || h < 1 {
		return nil, 1
	}
	f := fitResourceFormat(rows, w)
	n := len(rows)
	legend := resLegend(w)
	var details []int // rows with a detail line that fits
	for i, r := range rows {
		if r.detail != "" && visibleLen(r.detail)+3 <= w {
			details = append(details, i)
		}
	}
	hasDetail := func(i int) bool {
		for _, d := range details {
			if d == i {
				return true
			}
		}
		return false
	}
	build := func(airy, withDetails, blank, withLegend bool) []string {
		var out []string
		for i, r := range rows {
			out = append(out, f.line(r))
			if withDetails && hasDetail(i) {
				out = append(out, "   "+r.detail)
			}
			if airy && i < n-1 {
				out = append(out, "")
			}
		}
		if withLegend && legend != "" {
			if blank {
				out = append(out, "")
			}
			out = append(out, legend)
		}
		return out
	}
	// From roomy to bare: a line between rows, the detail lines, the legend.
	for _, v := range [][4]bool{
		{true, true, true, true},
		{false, true, true, true},
		{false, false, true, true},
		{false, false, false, true},
		{false, false, false, false},
	} {
		if out := build(v[0], v[1], v[2], v[3]); len(out) <= h {
			return out, 1
		}
	}
	// A page of rows, and the line that says so.
	if h == 1 {
		return []string{resMoreLine(1, n, n, w)}, 1
	}
	size := h - 1
	room := max(1, size)
	pages = max(1, (n+room-1)/room)
	page = ((page % pages) + pages) % pages
	first := page * room
	last := min(n, first+room)
	for _, r := range rows[first:last] {
		lines = append(lines, f.line(r))
	}
	for len(lines) < size {
		lines = append(lines, "")
	}
	return append(lines, resMoreLine(first+1, last, n, w)), pages
}
