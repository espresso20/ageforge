package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// dash_ledger.go is the top of the dashboard under the status ribbon: the
// Ledger. From the top it holds a long bar to the next age with what it is
// waiting on, the four pillars of the gate as small meters, and a card for
// each resource that has something going on (name, amount, rate, a trend line
// of the last few minutes).
//
// What it says about the gate is the Advance panel's: the pillars, their
// percents, the share of the way and the year are read from the same
// game.AdvanceView (and the engine's held calendar), so the two cannot
// disagree. The trend lines are the Ledger's own memory (ledgerTrail): a few
// samples of each amount, kept here in the UI and never in the game or a save.
//
// It is drawn in three sizes, by the terminal it is drawn on:
//
//	compact   under 100x30   the long bar with what it waits on, one row of cards
//	standard  from 100x30    that, the four pillars on one row, a row of two-line cards
//	full      the rich layout  the bar and the waiting line apart, pillars with their
//	                           notes, and as many rows of cards as the height allows
//
// The rows a size takes are fixed (ledgerHeight), so nothing under the Ledger
// moves when a number changes or a resource turns up.

type ledgerTier int

const (
	ledgerCompact ledgerTier = iota
	ledgerStandard
	ledgerFull
)

const (
	// ledgerCompactRows is the Next Age strip's old height: two rows.
	ledgerCompactRows = 2
	// ledgerStandardRows: the bar and what it waits on, the pillars, a card (two lines).
	ledgerStandardRows = 4
	// ledgerFullHead: the bar, what it waits on, the pillars' names, their
	// meters and a blank row; the cards follow, two lines a row of them.
	ledgerFullHead = 5

	// ledgerStdCols and ledgerStdRows are the smallest terminal the standard
	// Ledger is drawn on. The full one is the rich layout's.
	ledgerStdCols = 100
	ledgerStdRows = 30
)

// ledgerTierFor is the Ledger a w by h terminal gets.
func ledgerTierFor(w, h int) ledgerTier {
	switch {
	case dashRich(w, h):
		return ledgerFull
	case w >= ledgerStdCols && h >= ledgerStdRows:
		return ledgerStandard
	}
	return ledgerCompact
}

// ledgerCardRows is how many rows of cards the full Ledger has on a terminal
// h rows tall.
func ledgerCardRows(h int) int {
	switch {
	case h >= 44:
		return 3
	case h >= 38:
		return 2
	}
	return 1
}

// ledgerCardShape is how the cards sit on a w cell wide Ledger: how many to
// a row and how wide each is, two cells of air between them. The full
// Ledger on a wide terminal writes the wide card, the rest the narrow one.
func ledgerCardShape(w int) (per, cw int) {
	if w >= ledgerWideCols {
		per = max(1, (w-2)/(ledgerWideMinW+2))
		return per, min(28, (w-2)/per-2)
	}
	per = max(1, (w-2)/(ledgerNarrowW+2))
	return per, min(25, (w-2)/per-2)
}

// ledgerHeight is the height of the Ledger, under the status ribbon, on a
// w by h terminal for an age with slots resources to show. The full Ledger
// has the rows of cards the slots call for, up to what the height allows,
// so the rows change when the age does and at no other time.
func ledgerHeight(w, h, slots int) int {
	switch ledgerTierFor(w, h) {
	case ledgerFull:
		per, _ := ledgerCardShape(w)
		rows := min(ledgerCardRows(h), max(1, (slots+per-1)/per))
		return ledgerFullHead + 2*rows
	case ledgerStandard:
		return ledgerStandardRows
	}
	return ledgerCompactRows
}

// ledgerExtraRows is the rows the standard Ledger takes over the compact one
// (the mini map's smallest terminal is worked out from the standard size).
const ledgerExtraRows = ledgerStandardRows - ledgerCompactRows

// ===== The trend lines =====

const (
	// ledgerTrailLen is the samples kept of each amount, ledgerSampleTicks the
	// ticks between two samples: with a 2 second tick, four minutes.
	ledgerTrailLen    = 24
	ledgerSampleTicks = 5
	// ledgerSteady is the change over a trail, as a share of the larger end,
	// under which an amount is called steady.
	ledgerSteady = 0.01
)

// ledgerRing keeps the last ledgerTrailLen samples of one amount.
type ledgerRing struct {
	v    [ledgerTrailLen]float64
	n    int // samples held, up to ledgerTrailLen
	head int // where the next goes
}

func (r *ledgerRing) push(x float64) {
	r.v[r.head] = x
	r.head = (r.head + 1) % ledgerTrailLen
	if r.n < ledgerTrailLen {
		r.n++
	}
}

// appendTo adds the samples to out, oldest first.
func (r *ledgerRing) appendTo(out []float64) []float64 {
	start := (r.head - r.n + ledgerTrailLen) % ledgerTrailLen
	for i := 0; i < r.n; i++ {
		out = append(out, r.v[(start+i)%ledgerTrailLen])
	}
	return out
}

// ledgerTrail is the Ledger's memory of the resources' amounts. It samples
// by the game's tick, so a faster game fills it faster, and starts over when
// the tick goes back (a new game, a new run) or reset is called (a load).
type ledgerTrail struct {
	rings   map[string]*ledgerRing
	last    int
	started bool
}

func (t *ledgerTrail) reset() {
	clear(t.rings)
	t.started = false
}

// observe takes a sample of every unlocked resource when ledgerSampleTicks
// have passed since the last.
func (t *ledgerTrail) observe(state game.GameState) {
	if t.rings == nil {
		t.rings = map[string]*ledgerRing{}
	}
	if t.started && state.Tick < t.last {
		t.reset()
	}
	if t.started && state.Tick-t.last < ledgerSampleTicks {
		return
	}
	t.started, t.last = true, state.Tick
	for key, rs := range state.Resources {
		if !rs.Unlocked {
			continue
		}
		r := t.rings[key]
		if r == nil {
			r = &ledgerRing{}
			t.rings[key] = r
		}
		r.push(rs.Amount)
	}
}

// ledgerTrend says which way an amount has gone over its trail: 1 rising,
// -1 falling, 0 steady. A stock that is going down now is falling, whatever
// the trail says of the minutes before; with fewer than two samples the rate
// says.
func ledgerTrend(trail []float64, rate float64) int {
	if rate < 0 {
		return -1
	}
	if len(trail) < 2 {
		switch {
		case rate > 0:
			return 1
		case rate < 0:
			return -1
		}
		return 0
	}
	first, last := trail[0], trail[len(trail)-1]
	scale := math.Max(math.Abs(first), math.Abs(last))
	if scale == 0 || math.Abs(last-first) < scale*ledgerSteady {
		return 0
	}
	if last > first {
		return 1
	}
	return -1
}

// ledgerSparkLevels are the eight heights of a trend line, lowest first, and
// their plain ASCII twins.
const (
	ledgerSparkUnicode = "▁▂▃▄▅▆▇█"
	ledgerSparkASCII   = "_.-:=+*#"
)

// ledgerSpark draws trail as width cells of heights, the newest at the right
// and the cells before the first sample blank, scaled between the trail's own
// lowest and highest. A flat trail sits at the middle.
func ledgerSpark(trail []float64, width int, plain bool) string {
	levels := []rune(ledgerSparkUnicode)
	if plain {
		levels = []rune(ledgerSparkASCII)
	}
	if len(trail) > width {
		trail = trail[len(trail)-width:]
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range trail {
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	out := make([]rune, 0, width)
	for i := len(trail); i < width; i++ {
		out = append(out, ' ')
	}
	for _, v := range trail {
		level := 3
		if hi > lo {
			level = clamp(int((v-lo)/(hi-lo)*float64(len(levels)-1)+0.5), 0, len(levels)-1)
		}
		out = append(out, levels[level])
	}
	return string(out)
}

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

// ledgerCard is one resource's card.
type ledgerCard struct {
	Key, Name    string
	Amount, Rate float64
	Trail        []float64
	Trend        int
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
	// Need are the resources a blocker is short of, and Worst the one that
	// will take longest to find.
	Need  map[string]bool
	Worst string
	// Short are the names of Need, in the order of the cards.
	Short []string
	Cards []ledgerCard
}

// ledger is the dashboard's Ledger: its memory and the last thing it worked out.
type ledger struct {
	trail ledgerTrail
	data  ledgerData
	ready bool

	// slots is how many resources the age has to show (the ones that have
	// come by it), which sets the rows of cards (ledgerHeight).
	slots int

	// order is the resources in the ruleset's order, cached for the ruleset
	// it was read from; slotsAge is the age slots was counted for.
	set      *rules.Set
	order    []string
	slotsAge string
	buf      []float64
}

// update works the Ledger out from a snapshot. eng holds the calendar (the
// year never runs back within an age); it may be nil.
func (l *ledger) update(state game.GameState, eng *game.GameEngine) {
	l.trail.observe(state)
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
	l.cardsFor(state)
	l.countSlots(state)
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
		var rest []string
		for k := range state.Resources {
			if !seen[k] {
				rest = append(rest, k)
			}
		}
		sort.Strings(rest)
		l.order = append(l.order, rest...)
	}
	return l.order
}

// countSlots counts the resources that have come by the age: the places the
// Ledger keeps for cards. It is the same all through an age.
func (l *ledger) countSlots(state game.GameState) {
	set := state.Ruleset()
	if l.slotsAge == state.Age && l.slots > 0 && set == l.set {
		return
	}
	l.slotsAge, l.slots = state.Age, 0
	here, _ := set.Index(state.Age)
	for _, def := range set.Resources() {
		if i, ok := set.Index(def.Age); ok && i <= here {
			l.slots++
		}
	}
}

// blockers lists what the gate is waiting on and which resources a blocker is
// short of: for a building, the price of its next copy against the stock; for
// the wonder's bank, what is left to bank against the stock.
func (l *ledger) blockers(state game.GameState) {
	d := &l.data
	d.Waits, d.Then, d.Worst, d.Short = d.Waits[:0], "", "", d.Short[:0]
	if d.Need == nil {
		d.Need = map[string]bool{}
	}
	clear(d.Need)
	scores := map[string]float64{}

	short := func(cost map[string]float64, banked map[string]float64) {
		for res, price := range cost {
			left := price - banked[res]
			rs := state.Resources[res]
			if left <= rs.Amount {
				continue
			}
			// How many ticks the stock needs to cover what is left.
			score := math.MaxFloat64
			if rs.Rate > 0 {
				score = (left - rs.Amount) / rs.Rate
			}
			d.Need[res] = true
			if score > scores[res] {
				scores[res] = score
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

	// The resource that will take longest to find; ties go to the one the
	// ruleset lists first, so the choice does not change from frame to frame.
	best := -1.0
	for _, key := range l.resourceOrder(state) {
		if !d.Need[key] {
			continue
		}
		d.Short = append(d.Short, state.Resources[key].Name)
		if scores[key] > best {
			best, d.Worst = scores[key], key
		}
	}
}

// cardsFor makes a card of each resource that has something going on, in the
// ruleset's order. A resource with nothing held, no rate or not unlocked has
// none.
func (l *ledger) cardsFor(state game.GameState) {
	d := &l.data
	d.Cards, l.buf = d.Cards[:0], l.buf[:0]
	for _, key := range l.resourceOrder(state) {
		rs, ok := state.Resources[key]
		// A resource with nothing held and no rate has no card, unless the
		// gate is short of it: that is the one to look at.
		if !ok || !rs.Unlocked || (rs.Amount <= 0 && rs.Rate == 0 && !d.Need[key]) {
			continue
		}
		var trail []float64
		if r := l.trail.rings[key]; r != nil {
			// The samples, then the amount now, so the line ends where the stock is.
			at := len(l.buf)
			l.buf = append(r.appendTo(l.buf), rs.Amount)
			trail = l.buf[at:len(l.buf):len(l.buf)]
		}
		d.Cards = append(d.Cards, ledgerCard{Key: key, Name: rs.Name, Amount: rs.Amount, Rate: rs.Rate, Trail: trail, Trend: ledgerTrend(trail, rs.Rate)})
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

// chip puts s on the accent, as the bottleneck is marked.
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
		lines = lv.full(max(1, (h-ledgerFullHead)/2))
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
	'█': '#', '░': '-', '▲': '^', '▼': 'v', '▬': '=', '▶': '>', '✓': '+', '·': '|', '…': '~',
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

// ===== The cards =====

const (
	ledgerAmtW = 6 // the widest amount: "123Qi", "11.8Qi"
	// The cards come in three shapes, by the room: inline (the compact Ledger's
	// one-row strip), narrow (two lines) and wide (two lines, with the trend in
	// words).
	ledgerInlineW  = 6 + 1 + ledgerAmtW + 1 + 7 + 1 + 1
	ledgerNarrowW  = 17
	ledgerWideMinW = 26
	ledgerWideCols = 142 // the full Ledger writes wide cards from this many columns
)

// trendText is the trend as a mark and a word, padded to its column.
func trendText(trend int) string {
	switch {
	case trend > 0:
		return padRight(vibe.Line("ui.dash_ledger.trendText.01", "▲ rising"), 9)
	case trend < 0:
		return padRight(vibe.Line("ui.dash_ledger.trendText.02", "▼ falling"), 9)
	}
	return padRight(vibe.Line("ui.dash_ledger.trendText.03", "▬ steady"), 9)
}

func trendGlyph(trend int) (string, theme.Role) {
	switch {
	case trend > 0:
		return "▲", theme.RolePositive
	case trend < 0:
		return "▼", theme.RoleNegative
	}
	return "▬", theme.RoleDim
}

// rateCell is a rate as text in width cells, right aligned, with the unit
// when it fits.
func rateCell(rate float64, width int) (string, theme.Role) {
	text, color := rateParts(rate, "/t")
	if runeLen(text) > width {
		text, _ = rateParts(rate, "")
	}
	role := theme.RolePositive
	switch color {
	case "red":
		role = theme.RoleNegative
	case "gray":
		role = theme.RoleDim
	}
	if n := runeLen(text); n < width {
		text = strings.Repeat(" ", width-n) + text
	}
	return truncate(text, width), role
}

// amountCell is an amount right aligned in width cells.
func amountCell(amount float64, width int) string {
	text := FormatNumber(amount)
	if n := runeLen(text); n < width {
		text = strings.Repeat(" ", width-n) + text
	}
	return truncate(text, width)
}

// nameCell puts a card's mark and name in width cells, in the style its
// place calls for: on the accent for the resource that holds the gate up
// longest, in the warning colour for the others a blocker is short of.
func (lv ledgerView) nameCell(r *lrow, c ledgerCard, width int) {
	name := padRight(shortName(c.Name, width-1), width-1)
	mark := " "
	if lv.m.Worst == c.Key {
		mark = "▶"
	}
	switch {
	case lv.m.Worst == c.Key:
		r.chip(mark + name)
	case lv.m.Need[c.Key]:
		r.put(theme.RoleWarning, mark+name)
	default:
		r.put(theme.RoleLabel, mark+name)
	}
}

// cardLines writes one card in two lines at x offsets decided by the caller;
// it returns the two lines, both exactly cw cells wide.
func (lv ledgerView) cardLines(c ledgerCard, cw int) (l1, l2 string) {
	var a, b lrow
	wide := cw >= ledgerWideMinW
	arrow, arrowRole := trendGlyph(c.Trend)
	if wide {
		nameW := cw - 17 // the mark and name; then the amount (6) and the rate (9), a space before each
		lv.nameCell(&a, c, nameW)
		a.raw(" ", 1)
		a.put(theme.RoleBright, amountCell(c.Amount, ledgerAmtW))
		a.raw(" ", 1)
		rate, role := rateCell(c.Rate, 9)
		a.put(role, rate)
		sw := min(12, cw-11)
		b.raw(" ", 1)
		b.put(theme.RoleLabel, ledgerSpark(c.Trail, sw, lv.plain))
		b.raw(" ", 1)
		b.put(arrowRole, trendText(c.Trend))
	} else {
		nameW := cw - 1 - ledgerAmtW // the mark and name; then a space and the amount
		lv.nameCell(&a, c, nameW)
		a.raw(" ", 1)
		a.put(theme.RoleBright, amountCell(c.Amount, ledgerAmtW))
		sw := cw - 1 - 1 - 7 - 1 - 1
		b.raw(" ", 1)
		b.put(theme.RoleLabel, ledgerSpark(c.Trail, sw, lv.plain))
		b.raw(" ", 1)
		rate, role := rateCell(c.Rate, 7)
		b.put(role, rate)
		b.raw(" ", 1)
		b.put(arrowRole, arrow)
	}
	a.pad(cw)
	b.pad(cw)
	return a.String(), b.String()
}

// pickCards chooses the cards that fit slots places. When there are more
// than that, the last place is the count of the rest.
func (lv ledgerView) pickCards(slots int) (cards []ledgerCard, more int) {
	if len(lv.m.Cards) <= slots {
		return lv.m.Cards, 0
	}
	cards = lv.chooseCards(max(0, slots-1))
	return cards, len(lv.m.Cards) - len(cards)
}

// chooseCards keeps keep of the cards: the ones a blocker is short of first,
// then the ones going down, then the newest resources. They keep the
// ruleset's order, so a card does not move unless the gate's needs or a rate
// do.
func (lv ledgerView) chooseCards(keep int) []ledgerCard {
	all := lv.m.Cards
	chosen := make([]bool, len(all))
	left := keep
	take := func(i int) {
		if left > 0 && !chosen[i] {
			chosen[i] = true
			left--
		}
	}
	for i, c := range all {
		if lv.m.Need[c.Key] {
			take(i)
		}
	}
	for i, c := range all {
		if c.Rate < 0 {
			take(i)
		}
	}
	for i := len(all) - 1; i >= 0; i-- {
		take(i)
	}
	cards := make([]ledgerCard, 0, keep)
	for i, c := range all {
		if chosen[i] {
			cards = append(cards, c)
		}
	}
	return cards
}

// moreText is the count of the resources there is no card for.
func moreText(n int, short bool) string {
	if short {
		return fmt.Sprintf("+%d", n)
	}
	return vibe.Sprintf("ui.dash_ledger.moreText.01", "+%d more", n)
}

// cardRows lays cards out in rows of per cards, each cw cells wide at the
// pitch of the row, two lines a row; rows is how many rows there are room for.
func (lv ledgerView) cardRows(rows, per, cw int) []string {
	if per < 1 || rows < 1 || cw < ledgerNarrowW-3 {
		return nil
	}
	cards, more := lv.pickCards(rows * per)
	pitch := (lv.w - 2) / per
	out := make([]string, 0, rows*2)
	for row := 0; row < rows; row++ {
		var a, b lrow
		a.raw(" ", 1)
		b.raw(" ", 1)
		for col := 0; col < per; col++ {
			i := row*per + col
			a.pad(1 + col*pitch)
			b.pad(1 + col*pitch)
			switch {
			case i < len(cards):
				l1, l2 := lv.cardLines(cards[i], cw)
				a.raw(l1, cw)
				b.raw(l2, cw)
			case i == rows*per-1 && more > 0:
				a.put(theme.RoleDim, moreText(more, false))
			}
		}
		out = append(out, strings.TrimRight(a.String(), " "), strings.TrimRight(b.String(), " "))
	}
	return out
}

// inlineStrip is the compact Ledger's one row of cards: name, amount, rate
// and which way it is going, a card to ledgerInlineW cells.
func (lv ledgerView) inlineStrip() string {
	all := lv.m.Cards
	if len(all) == 0 {
		return ""
	}
	const gap = 2
	width := func(n int) int { return 1 + n*ledgerInlineW + (n-1)*gap }
	slots := 0
	for width(slots+1) <= lv.w-1 {
		slots++
	}
	if slots == 0 {
		return ""
	}
	cards, more := all, 0
	if len(all) > slots {
		// The count of the rest takes room too: in words if they fit, else a number.
		short := false
		for slots > 1 && width(slots)+gap+runeLen(moreText(len(all)-slots, short)) > lv.w-1 {
			if !short {
				short = true
				continue
			}
			slots--
		}
		cards = lv.chooseCards(slots)
		more = len(all) - len(cards)
	}
	var r lrow
	r.raw(" ", 1)
	for i, c := range cards {
		if i > 0 {
			r.raw(strings.Repeat(" ", gap), gap)
		}
		lv.inlineCard(&r, c)
	}
	if more > 0 {
		for _, short := range []bool{false, true} {
			if t := moreText(more, short); r.w+gap+runeLen(t) <= lv.w-1 {
				r.raw(strings.Repeat(" ", gap), gap)
				r.put(theme.RoleDim, t)
				break
			}
		}
	}
	return r.String()
}

func (lv ledgerView) inlineCard(r *lrow, c ledgerCard) {
	name := padRight(shortName(c.Name, 6), 6)
	switch {
	case lv.m.Worst == c.Key:
		r.chip(name)
	case lv.m.Need[c.Key]:
		r.put(theme.RoleWarning, name)
	default:
		r.put(theme.RoleLabel, name)
	}
	r.raw(" ", 1)
	r.put(theme.RoleBright, amountCell(c.Amount, ledgerAmtW))
	r.raw(" ", 1)
	rate, role := rateCell(c.Rate, 7)
	r.put(role, rate)
	r.raw(" ", 1)
	arrow, arole := trendGlyph(c.Trend)
	r.put(arole, arrow)
}

// ===== The three sizes =====

func (lv ledgerView) compact() []string {
	return []string{lv.barRow(true), lv.inlineStrip()}
}

func (lv ledgerView) standard() []string {
	lines := []string{lv.barRow(true), lv.pillarsRow()}
	per, cw := ledgerCardShape(lv.w)
	return append(lines, lv.cardRows(1, per, cw)...)
}

func (lv ledgerView) full(rows int) []string {
	names, meters := lv.pillarsTwoRows()
	lines := []string{lv.barRow(false), lv.waitRow(), names, meters, ""}
	per, cw := ledgerCardShape(lv.w)
	return append(lines, lv.cardRows(rows, per, cw)...)
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

// GetText is the rows as text, one to a line, with the colour tags left in
// unless stripTags.
func (b *ledgerBox) GetText(stripTags bool) string {
	lines := make([]string, len(b.lines))
	for i, l := range b.lines {
		if stripTags {
			l = stripLedgerTags(l)
		}
		lines[i] = l
	}
	return strings.Join(lines, "\n")
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

// stripLedgerTags is s without the tags the Ledger writes.
func stripLedgerTags(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '[' {
			if n, _, _, _, _, ok := ledgerTag(s[i:]); ok {
				i += n
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		sb.WriteRune(r)
		i += size
	}
	return sb.String()
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
