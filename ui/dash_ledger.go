package ui

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// dash_ledger.go is the top of the dashboard under the status ribbon: the
// Ledger. From the top it holds a long bar to the next age with what it is
// waiting on and, in the full size, the resources a blocker is short of, and
// the four pillars of the gate as small meters. (The resources themselves are
// the Resources panel's, below.)
//
// What it says about the gate is the Advance panel's: the pillars, their
// percents, the share of the way and the year are read from the same
// game.AdvanceView (and the engine's held calendar), so the two cannot
// disagree.
//
// It is drawn in three sizes, by the terminal it is drawn on:
//
//	compact   under 100x30   the long bar, with what it waits on after it
//	standard  from 100x30    that, and the four pillars on one row
//	full      the rich layout  the bar and the waiting line apart (with what the
//	                           gate is short of), pillars with their notes
//
// The rows a size takes are fixed (ledgerHeight), so nothing under the Ledger
// moves when a number changes.

type ledgerTier int

const (
	ledgerCompact ledgerTier = iota
	ledgerStandard
	ledgerFull
)

const (
	// ledgerCompactRows: the long bar and what it waits on, on one row.
	ledgerCompactRows = 1
	// ledgerStandardRows: that, and the pillars on a row.
	ledgerStandardRows = 2
	// ledgerFullHeight: the bar, what it waits on, the pillars' names and
	// their meters.
	ledgerFullHeight = 4

	// ledgerStdCols and ledgerStdRows are the smallest terminal the standard
	// Ledger is drawn on. The full one is the rich layout's, from
	// ledgerFullRows tall: under that the rich layout keeps the standard
	// Ledger and the boxes under it keep the rows.
	ledgerStdCols  = 100
	ledgerStdRows  = 30
	ledgerFullRows = 39
)

// ledgerTierFor is the Ledger a w by h terminal gets.
func ledgerTierFor(w, h int) ledgerTier {
	switch {
	case dashRich(w, h) && h >= ledgerFullRows:
		return ledgerFull
	case w >= ledgerStdCols && h >= ledgerStdRows:
		return ledgerStandard
	}
	return ledgerCompact
}

// ledgerHeight is the height of the Ledger, under the status ribbon, on a
// w by h terminal. It changes with the size of the terminal and at no other
// time.
func ledgerHeight(w, h int) int {
	switch ledgerTierFor(w, h) {
	case ledgerFull:
		return ledgerFullHeight
	case ledgerStandard:
		return ledgerStandardRows
	}
	return ledgerCompactRows
}

// ledgerExtraRows is the rows the standard Ledger takes over the compact one
// (the mini map's smallest terminal is worked out from the standard size).
const ledgerExtraRows = ledgerStandardRows - ledgerCompactRows

// ===== What the Ledger knows =====

// ledgerPillar is one pillar as the Ledger draws it.
type ledgerPillar struct {
	Name  string
	Share float64
	// Pct is the percent the Advance panel writes for the pillar.
	Pct  int
	Met  bool
	Note string
	// Status is what ends the pillar's meter in the Advance panel: "✓ stands"
	// or its percent.
	Status string
}

// ledgerData is everything the Ledger draws, worked out once a refresh.
type ledgerData struct {
	v game.AdvanceView
	// Pct is the long bar's percent: the Advance panel's "of the way".
	Pct     int
	Pillars []ledgerPillar
	// Waits are the requirements still to meet, in the Advance panel's order;
	// Then is the wonder that comes after them.
	Waits []string
	Then  string
	// Need are the resources a blocker is short of; Short their names, in
	// the ruleset's order.
	Need  map[string]bool
	Short []string
}

// ledger is the dashboard's Ledger: the last thing it worked out.
type ledger struct {
	data  ledgerData
	ready bool

	// order is the resources in the ruleset's order, cached for the ruleset
	// it was read from.
	set   *rules.Set
	order []string
}

// update works the Ledger out from a snapshot. eng holds the calendar (the
// year never runs back within an age); it may be nil.
func (l *ledger) update(state game.GameState, eng *game.GameEngine) {
	v := state.AdvanceView()
	if eng != nil {
		v = eng.HoldCalendar(v)
	}
	d := &l.data
	d.v = v
	d.Pct = sharePercent(v.Share, v.Share >= 1)
	d.Pillars = d.Pillars[:0]
	for _, p := range v.Pillars {
		status, _ := pillarStatus(p)
		d.Pillars = append(d.Pillars, ledgerPillar{Name: p.Name, Share: p.Progress, Pct: sharePercent(p.Progress, false), Met: p.Met, Note: pillarNote(p), Status: status})
	}
	l.blockers(state)
	l.ready = true
}

// resourceOrder is the ruleset's resources in its own order, then any other
// the snapshot holds by key.
func (l *ledger) resourceOrder(state game.GameState) []string {
	set := state.Ruleset()
	if set != l.set || l.order == nil {
		l.set, l.order = set, l.order[:0]
		seen := map[string]bool{}
		for _, def := range set.Resources() {
			l.order = append(l.order, def.Key)
			seen[def.Key] = true
		}
		rest := keysWhere(state.Resources, func(k string, _ game.ResourceState) bool { return !seen[k] })
		l.order = append(l.order, rest...)
	}
	return l.order
}

// blockers lists what the gate is waiting on and which resources a blocker is
// short of: for a building, the price of its next copy against the stock; for
// the wonder's bank, what is left to bank against the stock.
func (l *ledger) blockers(state game.GameState) {
	d := &l.data
	d.Waits, d.Then, d.Short = d.Waits[:0], "", d.Short[:0]
	if d.Need == nil {
		d.Need = map[string]bool{}
	}
	clear(d.Need)

	short := func(cost map[string]float64, banked map[string]float64) {
		for res, price := range cost {
			if price-banked[res] > state.Resources[res].Amount {
				d.Need[res] = true
			}
		}
	}

	var build *game.GateRow
	bankOpen := false
	for _, p := range d.v.Pillars {
		for i := range p.Rows {
			r := &p.Rows[i]
			if r.Met {
				continue
			}
			switch r.Kind {
			case game.GateRowCount:
				d.Waits = append(d.Waits, fmt.Sprintf("%s %d/%d", r.Label, r.Have, r.Need))
				short(buildingPrice(&state, r.Key, state.Buildings[r.Key]), nil)
			case game.GateRowAmount:
				d.Waits = append(d.Waits, fmt.Sprintf("%s %s/%s", r.Label, textfmt.Number(r.Value), textfmt.Number(r.Target)))
			case game.GateRowShare:
				bankOpen = true
				d.Waits = append(d.Waits, fmt.Sprintf("%s %d%%", r.Label, r.Have))
				if def, ok := state.Ruleset().Building(r.Key); ok {
					short(def.BaseCost, state.Buildings[r.Key].WonderBank)
				}
			default:
				if p.Key == config.PillarMonument {
					build = r
					continue
				}
				item := r.Label
				if r.Note != "" {
					item = fmt.Sprintf("%s (%s)", r.Label, r.Note)
				}
				d.Waits = append(d.Waits, item)
			}
		}
	}
	if build != nil {
		if bankOpen || len(d.Waits) > 0 {
			d.Then = build.Label
		} else {
			item := build.Label
			if build.Note != "" {
				item = fmt.Sprintf("%s (%s)", build.Label, build.Note)
			}
			d.Waits = append(d.Waits, item)
		}
	}

	// The names go in the ruleset's order, so the line does not change from
	// frame to frame.
	for _, key := range l.resourceOrder(state) {
		if d.Need[key] {
			d.Short = append(d.Short, state.Resources[key].Name)
		}
	}
}

// ===== Drawing =====

// lrow builds one row of tagged text and counts the cells it takes.
type lrow struct {
	sb strings.Builder
	w  int
}

func (r *lrow) put(role theme.Role, s string) {
	if s == "" {
		return
	}
	r.sb.WriteString(theme.Tag(role))
	r.sb.WriteString(s)
	r.sb.WriteString("[-]")
	r.w += runeLen(s)
}

// chip puts s on the accent, as the resources a blocker is short of are.
func (r *lrow) chip(s string) {
	r.sb.WriteString(theme.TagFgBg(theme.RoleOnAccent, theme.RoleAccent))
	r.sb.WriteString(s)
	r.sb.WriteString(theme.Reset)
	r.w += runeLen(s)
}

func (r *lrow) raw(s string, cells int) {
	r.sb.WriteString(s)
	r.w += cells
}

func (r *lrow) pad(to int) {
	if r.w < to {
		r.sb.WriteString(strings.Repeat(" ", to-r.w))
		r.w = to
	}
}

func (r *lrow) bar(share float64, width int, fill theme.Role) {
	if width <= 0 {
		return
	}
	n := barFill(share, width)
	r.put(fill, strings.Repeat("█", n))
	r.put(theme.RoleDim, strings.Repeat("░", width-n))
}

func (r *lrow) String() string { return r.sb.String() }

// ledgerView is a Ledger drawn for one terminal.
type ledgerView struct {
	m     *ledgerData
	w     int
	plain bool
}

// ledgerLines writes the Ledger's rows (h at most) of a tier for a view w
// cells wide, in the glyphs of the tier.
func ledgerLines(m *ledgerData, tier ledgerTier, w, h int, glyphs mapmodel.GlyphTier) []string {
	lv := ledgerView{m: m, w: w, plain: glyphs == mapmodel.TierASCII}
	var lines []string
	switch tier {
	case ledgerFull:
		lines = lv.full()
	case ledgerStandard:
		lines = lv.standard()
	default:
		lines = lv.compact()
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	if lv.plain {
		for i := range lines {
			lines[i] = foldLedger(lines[i])
		}
	}
	return lines
}

// ledgerFold are the marks the plain glyph tier writes the Ledger's marks as.
var ledgerFold = map[rune]rune{
	'█': '#', '░': '-', '✓': '+', '·': '|', '…': '~',
}

// foldLedger writes s in ASCII.
func foldLedger(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return s
	}
	out := []rune(s)
	for i, r := range out {
		if r < 0x80 {
			continue
		}
		if f, ok := ledgerFold[r]; ok {
			out[i] = f
		} else {
			out[i] = mapmodel.Fold(r, mapmodel.TierASCII)
		}
	}
	return string(out)
}

// leadText is the long bar's label.
func (lv ledgerView) leadText() string {
	return vibe.Sprintf("ui.dash_ledger.leadText.01", "To the %s", lv.m.v.NextName)
}

// charterHint is the hint that opens the Advance panel, when it fits.
func (lv ledgerView) charterHint() string { return charterStripHint.String() }

// finalRow is the long bar's row in the last age.
func (lv ledgerView) finalRow() string {
	var r lrow
	r.put(theme.RoleAccent, " "+truncate(vibe.Line("ui.dash_ledger.finalRow.01", "You have reached the final age."), max(1, lv.w-2)))
	return r.String()
}

// readyText is what the waiting line says once every pillar stands.
func (lv ledgerView) readyText() (role theme.Role, text string) {
	if lv.m.v.Ready {
		return theme.RolePositive, vibe.Line("ui.dash_ledger.readyText.01", "Every pillar stands. Type advance to go on.")
	}
	return theme.RoleWarning, vibe.Line("ui.dash_ledger.readyText.02", "Every pillar stands. Answer the catastrophe first (type catastrophe).")
}

// waitText is the line of what the gate waits on, in at most avail cells:
// as many items as fit, then how many more there are. It returns the tagged
// text and its width.
func (lv ledgerView) waitText(avail int) (string, int) {
	m := lv.m
	if m.v.AllStand {
		role, text := lv.readyText()
		for _, t := range []string{text, truncate(text, avail)} {
			if runeLen(t) <= avail {
				var r lrow
				r.put(role, t)
				return r.String(), r.w
			}
		}
		return "", 0
	}
	items := m.Waits
	if m.Then != "" {
		items = append(append([]string(nil), items...), vibe.Sprintf("ui.dash_ledger.waitText.01", "then %s", m.Then))
	}
	if len(items) == 0 {
		return "", 0
	}
	const sep = " · "
	sepW := runeLen(sep)
	for _, prefix := range []string{vibe.Line("ui.dash_ledger.waitText.02", "Waiting on: "), vibe.Line("ui.dash_ledger.waitText.03", "Needs: ")} {
		pw := runeLen(prefix)
		// The most items that fit, with the count of the rest after them.
		for k := len(items); k >= 1; k-- {
			w := pw
			for i := 0; i < k; i++ {
				if i > 0 {
					w += sepW
				}
				w += runeLen(items[i])
			}
			tail := ""
			if k < len(items) {
				tail = vibe.Sprintf("ui.dash_ledger.waitText.04", "+%d more", len(items)-k)
				w += sepW + runeLen(tail)
			}
			if w > avail {
				continue
			}
			var r lrow
			r.put(theme.RoleDim, prefix)
			for i := 0; i < k; i++ {
				if i > 0 {
					r.put(theme.RoleDim, sep)
				}
				r.put(theme.RoleWarning, items[i])
			}
			if tail != "" {
				r.put(theme.RoleDim, sep)
				r.put(theme.RoleDim, tail)
			}
			return r.String(), r.w
		}
	}
	return "", 0
}

// shortText is the resources a blocker is short of.
func (lv ledgerView) shortText(avail int) (string, int) {
	names := lv.m.Short
	if len(names) == 0 || lv.m.v.AllStand {
		return "", 0
	}
	lead := vibe.Line("ui.dash_ledger.shortText.01", "Short of: ")
	var r lrow
	r.put(theme.RoleDim, lead)
	for i, n := range names {
		sep := ""
		if i > 0 {
			sep = " · "
		}
		if r.w+runeLen(sep)+runeLen(n) > avail {
			break
		}
		if sep != "" {
			r.put(theme.RoleDim, sep)
		}
		r.chip(n)
	}
	if r.w == runeLen(lead) {
		return "", 0
	}
	return r.String(), r.w
}

// barRow is the long bar: its label, the bar, its percent and, in the
// compact and standard Ledgers (inline), what the gate waits on after it.
func (lv ledgerView) barRow(inline bool) string {
	m := lv.m
	if m.v.Final {
		return lv.finalRow()
	}
	lead := lv.leadText()
	fill := theme.RoleAccent
	if m.v.AllStand {
		fill = theme.RolePositive
	}
	// build lays the row out with a bar of bw cells and the room that is left
	// for the waiting text.
	build := func(bw int, lead string) (string, int) {
		var r lrow
		r.raw(" ", 1)
		r.put(theme.RoleLabel, lead)
		r.raw(" ", 1)
		r.bar(m.v.Share, bw, fill)
		r.raw(" ", 1)
		r.put(theme.RoleBright, fmt.Sprintf("%3d%%", m.Pct))
		if inline {
			text, tw := lv.waitText(lv.w - r.w - 2)
			if tw > 0 {
				r.raw("  ", 2)
				r.raw(text, tw)
			}
		}
		return r.String(), r.w
	}
	bw := clamp(lv.w/8, 6, 16)
	hint := ""
	if !inline {
		// The long bar of the full Ledger takes the room it has, up to a point,
		// and leaves room for the hint that opens the Advance panel.
		hint = lv.charterHint()
		bw = clamp(lv.w-1-runeLen(lead)-1-1-4-runeLen(hint)-2, 12, 72)
	}
	for ; bw >= 4; bw-- {
		line, w := build(bw, lead)
		if w > lv.w {
			continue
		}
		if hint != "" && w+2+runeLen(hint) <= lv.w {
			line += "  [dim]" + hint + "[-]"
		}
		return line
	}
	line, _ := build(4, truncate(lead, max(4, lv.w-12)))
	return line
}

// waitRow is the full Ledger's second row: what the gate waits on, and the
// resources a blocker is short of.
func (lv ledgerView) waitRow() string {
	if lv.m.v.Final {
		return ""
	}
	avail := lv.w - 2
	short, sw := lv.shortText(min(40, avail/3))
	if sw > 0 {
		avail -= sw + 3
	}
	text, tw := lv.waitText(avail)
	var r lrow
	r.raw(" ", 1)
	r.raw(text, tw)
	if sw > 0 {
		r.pad(1 + tw + 3)
		r.raw(short, sw)
	}
	if tw == 0 && sw == 0 {
		return ""
	}
	return r.String()
}

// pillarsRow is the standard Ledger's one row of pillars: name, meter, percent.
func (lv ledgerView) pillarsRow() string {
	m := lv.m
	if len(m.Pillars) == 0 {
		return ""
	}
	pw := (lv.w - 2) / len(m.Pillars)
	if pw-2-9-1-1-3 < 4 {
		return ""
	}
	var r lrow
	r.raw(" ", 1)
	for i, p := range m.Pillars {
		start := 1 + i*pw
		r.pad(start)
		role := theme.RoleAccent
		if p.Met {
			role = theme.RolePositive
		}
		r.put(role, padRight(shortName(p.Name, 9), 9))
		r.raw(" ", 1)
		bw := max(4, pw-2-9-1-1-3)
		r.bar(p.Share, bw, role)
		r.raw(" ", 1)
		r.put(role, padRight(lv.pillarEnd(p, false), 3))
	}
	return r.String()
}

// pillarEnd is what ends a pillar's meter: the Advance panel's words when
// there is room for them, else a check when it stands or its percent.
func (lv ledgerView) pillarEnd(p ledgerPillar, long bool) string {
	if long || !p.Met {
		return p.Status
	}
	return "✓"
}

// pillarsTwoRows is the full Ledger's pillars: a row of names and notes, a row
// of meters.
func (lv ledgerView) pillarsTwoRows() (names, meters string) {
	m := lv.m
	if len(m.Pillars) == 0 {
		return "", ""
	}
	pw := (lv.w - 2) / len(m.Pillars)
	if pw-2-8-1 < 4 {
		return "", ""
	}
	var a, b lrow
	a.raw(" ", 1)
	b.raw(" ", 1)
	for i, p := range m.Pillars {
		start := 1 + i*pw
		a.pad(start)
		b.pad(start)
		role := theme.RoleAccent
		if p.Met {
			role = theme.RolePositive
		}
		inner := pw - 2
		name := strings.ToUpper(shortName(p.Name, 12))
		a.put(role, name)
		if room := inner - runeLen(name) - 1; room >= 6 {
			a.raw(" ", 1)
			a.put(theme.RoleDim, truncate(p.Note, room))
		}
		status := lv.pillarEnd(p, true)
		sw := max(runeLen(status), 8)
		bw := max(4, inner-sw-1)
		b.bar(p.Share, bw, role)
		b.raw(" ", 1)
		b.put(role, status)
	}
	return a.String(), b.String()
}

// ===== The three sizes =====

func (lv ledgerView) compact() []string {
	return []string{lv.barRow(true)}
}

func (lv ledgerView) standard() []string {
	return []string{lv.barRow(true), lv.pillarsRow()}
}

func (lv ledgerView) full() []string {
	names, meters := lv.pillarsTwoRows()
	return []string{lv.barRow(false), lv.waitRow(), names, meters}
}

// ===== The box =====

// ledgerBox is the Ledger's primitive. Its rows are strings of colour tags,
// as a fitView holds, and are written for the size it is drawn at in the same
// way (the render function is handed the width and the rows); but it draws them
// a rune at a time. The Ledger is a thousand cells of single-width marks that
// change twice a second, and a TextView takes a good part of a frame to parse
// and lay out that much.
type ledgerBox struct {
	*tview.Box
	render func(w, h int) []string
	lines  []string
	w, h   int
	stale  bool
	// drawn: the box has been drawn, so it has a real size.
	drawn bool
}

func newLedgerBox(render func(w, h int) []string) *ledgerBox {
	return &ledgerBox{Box: tview.NewBox(), render: render, stale: true}
}

// changed says what the box shows has changed, and writes it again, as
// fitView.changed does.
func (b *ledgerBox) changed() {
	b.stale = true
	if b.drawn {
		b.layout()
		return
	}
	const unbounded = 1 << 16
	b.lines = b.render(unbounded, unbounded)
}

// layout writes the rows for the box's current size if they are out of date.
func (b *ledgerBox) layout() {
	_, _, w, h := b.GetInnerRect()
	if b.stale || w != b.w || h != b.h {
		b.lines = b.render(w, h)
		b.w, b.h, b.stale = w, h, false
	}
}

// ledgerTagRoles are the names of the colour tags the Ledger writes, with
// the role each stands for.
var ledgerTagRoles = sync.OnceValue(func() map[string]theme.Role { return theme.TagNames() })

// ledgerTag reads the tag that opens at s[0] ('['): its length in bytes and
// the colours it sets. ok is false when what follows is not one of the
// Ledger's tags, which then print as they are.
func ledgerTag(s string) (n int, fg, bg tcell.Color, setFg, setBg, ok bool) {
	end := strings.IndexByte(s, ']')
	if end < 2 || end > 24 {
		return
	}
	body := s[1:end]
	fgName, bgName, _ := strings.Cut(body, ":")
	bgName, _, _ = strings.Cut(bgName, ":")
	// The reset tags are "-" and "-:-:-"; a name is a role's.
	known := func(name string) (tcell.Color, bool, bool) {
		switch name {
		case "":
			return 0, false, true
		case "-":
			return tcell.ColorDefault, true, true
		}
		role, found := ledgerTagRoles()[name]
		if !found {
			return 0, false, false
		}
		return theme.Ref(role), true, true
	}
	var okFg, okBg bool
	if fg, setFg, okFg = known(fgName); !okFg {
		return
	}
	if bg, setBg, okBg = known(bgName); !okBg {
		return
	}
	return end + 1, fg, bg, setFg, setBg, true
}

// ledgerRuneWidth is the cells a rune takes: one for the marks the Ledger
// draws, and what the terminal's rules say for any other.
func ledgerRuneWidth(r rune) int {
	switch {
	case r < 0x1100, r >= 0x2580 && r <= 0x25FF, r == 0x2713:
		return 1
	}
	return uniseg.StringWidth(string(r))
}

// Draw paints the box and, rune by rune, its rows.
func (b *ledgerBox) Draw(screen tcell.Screen) {
	if !b.drawn {
		b.drawn, b.stale = true, true
	}
	b.layout()
	b.DrawForSubclass(screen, b)
	x, y, w, h := b.GetInnerRect()
	baseFg, baseBg := tview.Styles.PrimaryTextColor, b.GetBackgroundColor()
	base := tcell.StyleDefault.Foreground(baseFg).Background(baseBg)
	for row, line := range b.lines {
		if row >= h {
			break
		}
		style, col := base, 0
		for i := 0; i < len(line) && col < w; {
			if line[i] == '[' {
				if n, fg, bg, setFg, setBg, ok := ledgerTag(line[i:]); ok {
					if setFg && fg == tcell.ColorDefault {
						style = style.Foreground(baseFg)
					} else if setFg {
						style = style.Foreground(fg)
					}
					if setBg && bg == tcell.ColorDefault {
						style = style.Background(baseBg)
					} else if setBg {
						style = style.Background(bg)
					}
					i += n
					continue
				}
			}
			r, size := utf8.DecodeRuneInString(line[i:])
			i += size
			if r != ' ' || style != base {
				screen.SetContent(x+col, y+row, r, nil, style)
			}
			col += max(1, ledgerRuneWidth(r))
		}
	}
}
