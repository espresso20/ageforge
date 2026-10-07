package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// resources_box.go lays the dashboard's Resources box out for the size it
// is drawn at. Every resource is one row on one grid (name, amount over cap,
// rate, bar, marker), so the columns line up whatever the widest name or
// rate is, and a row never wraps: when the box is narrow it gives up the
// least important part first (the bar shrinks, then goes; then the spacing;
// then the cap; then the rate's unit shortens; then long names). When the
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
)

// resRow is one resource's row before it is laid out.
type resRow struct {
	name     string
	amt      string // the amount column
	amtColor string
	cap      string // the cap column
	// status stands where "/ cap" does on a row that has no cap to show
	// (faith's strength, culture past its last step): the longest of these
	// that fits, plain text, drawn in statusColor.
	status      []string
	statusColor string
	rate        float64
	fill        float64 // how full the bar is, 0 to 1
	// progress draws the bar as progress toward something (faith strength,
	// culture's next step) instead of as a store filling up.
	progress bool
	glyph    string // the marker after the bar: "◈" nearly full, "▼" falling
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

// storeRow is an ordinary resource: what is held over what the store holds.
func storeRow(rs game.ResourceState) resRow {
	r := resRow{name: rs.Name, amt: FormatNumber(rs.Amount), amtColor: "white", cap: FormatNumber(rs.Storage), rate: rs.Rate}
	if rs.Storage > 0 {
		r.fill = rs.Amount / rs.Storage
	}
	switch {
	case rs.Storage > 0 && rs.Amount >= rs.Storage*0.9:
		r.amtColor = "yellow"
	case rs.Rate > 0:
		r.amtColor = "green"
	case rs.Rate < 0:
		r.amtColor = "red"
	}
	switch {
	case rs.Storage > 0 && rs.Amount/rs.Storage >= 0.95:
		r.glyph, r.glyphCol = "◈", "gold"
	case rs.Rate < 0:
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
			r.detail = "[gray]at " + FormatNumber(step) + ": " + cultureThresholdLabels[i] + "[-]"
			return r
		}
	}
	r.status, r.statusColor, r.fill = []string{"✦ Culture mastered", "✦ mastered", "✦"}, "gold", 1
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
	odds := fmt.Sprintf("%.0f%% good", game.EpochGoodChanceIn(o.FaithBand)*100)
	strength := o.FaithStrength
	switch {
	case held <= 0:
		return faithBand{"red", "✝ No faith", "✝ None", "✝", odds}
	case o.FaithBand == game.FaithBandLow:
		return faithBand{"gray", "◈ Dim faith", "◈ Dim", "◈", odds}
	case o.FaithBand == game.FaithBandHigh && strength >= 1:
		return faithBand{"gold", "✦ Faith full", "✦ Full", "✦", odds}
	case o.FaithBand == game.FaithBandHigh:
		return faithBand{"green", "◈ Strong faith", "◈ Strong", "◈", odds}
	case strength <= 0.50:
		return faithBand{"white", "◈ Low faith", "◈ Low", "◈", odds}
	}
	return faithBand{"yellow", "◈ Faith", "◈ Faith", "◈", odds}
}

// faithRow is faith: the amount column and the bar are the faith strength
// the rolls read (o.FaithStrength: the faith the town's own faith buildings
// made, against four and a half moderate sets'), not the fill of the store
// faith is kept in. The band stands where a cap would, and the detail line
// says it in full with the odds it gives.
func faithRow(rs game.ResourceState, o game.CatastropheOutlook) resRow {
	band := faithBandFor(rs.Amount, o)
	return resRow{
		name: rs.Name, amt: fmt.Sprintf("%.0f%%", o.FaithStrength*100), amtColor: "white",
		status: []string{band.label, band.short, band.mark}, statusColor: band.color,
		rate: rs.Rate, fill: o.FaithStrength, progress: true,
		detail: fmt.Sprintf("[%s]%s[-] [gray](epoch: %s)[-]", band.color, band.label, band.epochOdds),
	}
}

// resFormat is the grid the rows are drawn on.
type resFormat struct {
	nameW, amtW, capW, rateW int
	gap                      int    // cells between columns
	sep                      int    // cells between amount and cap: 3 is " / ", 1 is "/", 0 is no cap column
	bar                      int    // cells of bar, 0 for none
	unit                     string // the rate's unit: "/tick", or "/t" when even the cap had to go
}

// midW is the cells between the amount and the rate columns' gaps: the
// separator and the cap.
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
	for _, r := range rows {
		longest = max(longest, runeLen(r.name))
		f.amtW = max(f.amtW, runeLen(r.amt))
		f.capW = max(f.capW, runeLen(r.cap))
		text, _ := rateParts(r.rate, f.unit)
		f.rateW = max(f.rateW, runeLen(text))
	}
	// spacings from roomy to tight; barIn is the widest bar w leaves.
	spacings := [][2]int{{2, 3}, {1, 3}, {1, 1}}
	barIn := func(f resFormat) int {
		f.bar = 0
		return min(resBarMax, w-f.width()-f.gap)
	}
	try := func(nameW, least int) bool {
		f.nameW = nameW
		for _, s := range spacings {
			f.gap, f.sep = s[0], s[1]
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
	// No bar: the roomiest spacing that fits, then without the cap, then
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
			b.WriteString(resourceBar(r.fill, 1, f.bar))
		}
	}
	if r.glyph != "" {
		fmt.Fprintf(&b, " [%s]%s[-]", r.glyphCol, r.glyph)
	}
	return b.String()
}

// progressCells is a bar of width cells, ratio of it filled, in the accent
// color: progress toward something, where resourceBar is a store filling.
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
// amount's color and the two markers mean, so neither carries meaning by
// color alone. The box prints the fullest that fits.
var resLegends = []string{
	" [gray]Amount:[-] [green]rising[-] [gray]·[-] [red]falling[-] [gray]·[-] [yellow]90%+ full[-] [gray]·[-] [gold]◈[-] [gray]95%+ full ·[-] [red]▼[-] [gray]falling[-]",
	" [green]rising[-] [gray]·[-] [yellow]90%+ full[-] [gray]·[-] [gold]◈[-] [gray]95%+ full ·[-] [red]▼[-] [gray]falling[-]",
	" [gold]◈[-] [gray]95%+ full ·[-] [red]▼[-] [gray]falling[-]",
	" [gold]◈[-] [gray]full[-] [red]▼[-] [gray]falling[-]",
}

func resLegend(w int) string {
	for _, l := range resLegends {
		if visibleLen(l) <= w {
			return l
		}
	}
	return ""
}

// resMoreLine is the last line of a box that shows a page of its rows:
// which rows these are, and the key that turns the page.
func resMoreLine(first, last, total, w int) string {
	rng := fmt.Sprintf("%d-%d of %d", first, last, total)
	for _, l := range []string{
		" " + theme.Paint(theme.RoleDim, rng+" ·") + " " + theme.Keycap(resourcePageKeyName) + theme.Paint(theme.RoleDim, " next page"),
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
	pages = (n + size - 1) / size
	page = ((page % pages) + pages) % pages
	first := page * size
	last := min(n, first+size)
	for _, r := range rows[first:last] {
		lines = append(lines, f.line(r))
	}
	for len(lines) < size {
		lines = append(lines, "")
	}
	return append(lines, resMoreLine(first+1, last, n, w)), pages
}
