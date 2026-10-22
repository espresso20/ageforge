package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// expedition_view.go draws the Expeditions panel: a pure function of the
// game's state and what the player has picked, into a grid the size of the
// panel. The Ready Room is where a party is chosen and sent; the Expedition
// Report follows the parties out and lists what came home. The layout is
// worked out at 100 columns by 30 rows (the lab's) and shrinks to 80 by 21
// and a little less: the legend goes first, then the briefing's blank lines
// and fine print.

// expView is what the player has picked and the clock.
type expView struct {
	// report is the Expedition Report; false is the Ready Room.
	report bool
	// sel is the highlighted offer's key (expKey); "" picks the first.
	sel string
	// party is the highlighted party on the Report.
	party int
	// confirm is the send box.
	confirm bool
	// flash is a line for the key bar (a refusal, a launch), good or not.
	flash     string
	flashGood bool
	// marked is how many of the returns listed on the Report are new.
	marked int
	// frame is the animation frame; atRest holds everything still.
	frame int
}

// expLayout is where the pieces of the panel go on a w by h panel.
type expLayout struct {
	w, h  int
	small bool // too small to draw
	// outH is the height of the Out now box: 4 on a tall panel, 3 on a
	// short one, 1 (a plain line, no box) on the smallest.
	outH int
	// legendY is the legend row, -1 when the panel is too short for it.
	legendY int
	mainY   int // the map and the card
	mainH   int
	mapW    int
	cardX   int
	cardW   int
	outY    int
	footY   int
}

func layoutFor(w, h int) expLayout {
	l := expLayout{w: w, h: h, legendY: -1}
	if w < 60 || h < 14 {
		l.small = true
		return l
	}
	legend := 0
	switch {
	case h >= 30:
		l.outH, legend = 4, 1
	case h >= 22:
		l.outH = 3
	default:
		l.outH = 1
	}
	l.cardW = 35
	if w < 76 {
		l.cardW = 31
	}
	l.mapW = w - 3 - l.cardW
	l.cardX = 1 + l.mapW + 1
	l.mainY = 2
	l.mainH = h - 4 - l.outH - legend
	l.outY = l.mainY + l.mainH
	if legend > 0 {
		l.legendY = l.outY + l.outH
	}
	l.footY = h - 2
	return l
}

// expKey is the identity of an offer or of the mission sent from it.
func expKey(title string, expiry int) string { return fmt.Sprintf("%s|%d", title, expiry) }

// expPinItem is one pin on the map: an offer on the board, or a mission
// out (which stands at the place its offer did).
type expPinItem struct {
	key    string
	tier   string
	rare   bool
	offer  *game.ExpeditionInfo
	act    *game.ExpeditionSnapshot
	at     pt
	placed bool
}

func (it expPinItem) glyph() rune {
	switch {
	case it.rare:
		return '◈'
	case it.tier == game.TierLong:
		return '◇'
	case it.tier == game.TierGrand:
		if it.military() {
			return '▲'
		}
		return '■'
	}
	return '○'
}

func (it expPinItem) military() bool {
	if it.offer != nil {
		return it.offer.Grand == game.GrandMilitary
	}
	return it.act != nil && it.act.Grand == game.GrandMilitary
}

// expScene is the map with everything on it.
type expScene struct {
	terrain *expTerrain
	town    pt
	acts    []expPinItem
	offers  []expPinItem
}

// buildScene places every mission under way and every offer on a map w by h.
// Missions out are placed first, so an offer that is sent stands where it
// stood.
func buildScene(state game.GameState, w, h int) *expScene {
	sc := &expScene{terrain: buildTerrain(state.Seed, state.Age, w, h), town: townAt(w, h)}
	mil := state.Military
	var specs []pinSpec
	for _, a := range []*game.ExpeditionSnapshot{mil.ActiveShort, mil.ActiveLong, mil.ActiveGrand} {
		if a != nil {
			sc.acts = append(sc.acts, expPinItem{key: expKey(a.Name, a.ExpiresTick), tier: a.Tier, rare: a.Rare, act: a})
			specs = append(specs, pinSpec{pinHash(a.Name, a.ExpiresTick), a.Tier})
		}
	}
	for _, slot := range mil.Board {
		if slot.Offer != nil {
			o := *slot.Offer
			sc.offers = append(sc.offers, expPinItem{key: expKey(o.Name, o.ExpiresTick), tier: o.Tier, rare: o.Rare, offer: &o})
			specs = append(specs, pinSpec{pinHash(o.Name, o.ExpiresTick), o.Tier})
		}
	}
	at, ok := placePins(sc.terrain, specs)
	for i := range sc.acts {
		sc.acts[i].at, sc.acts[i].placed = at[i], ok[i]
	}
	for i := range sc.offers {
		j := len(sc.acts) + i
		sc.offers[i].at, sc.offers[i].placed = at[j], ok[j]
	}
	return sc
}

// selected is the index of the highlighted offer: the one the key names, else
// the first.
func (sc *expScene) selected(key string) int {
	for i, it := range sc.offers {
		if it.key == key {
			return i
		}
	}
	return 0
}

// pinCells are the offers' cells, for hopping.
func (sc *expScene) pinCells() []pt {
	cells := make([]pt, len(sc.offers))
	for i, it := range sc.offers {
		cells[i] = it.at
	}
	return cells
}

// dur is a span of ticks as wall-clock time.
func expDur(ticks int, state game.GameState) string {
	return humanizeDuration(time.Duration(max(ticks, 0)) * tickInterval(state))
}

// expBar is a bar n cells wide, p of it full.
func expBar(p float64, n int, full, empty string) string {
	k := int(math.Round(math.Max(0, math.Min(1, p)) * float64(n)))
	return strings.Repeat(full, k) + strings.Repeat(empty, n-k)
}

func oddsKind(p float64) xk {
	switch {
	case p >= .75:
		return xGood
	case p >= .60:
		return xGold
	}
	return xBad
}

func capFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[:1])) + string(r[1:])
}

// briefTitle is an offer's title without the mission's name and without
// the Legendary mark, which the card already says: "Count the herds near
// the Copper Hills".
func briefTitle(o game.ExpeditionInfo) string {
	title := o.Name
	if i := strings.Index(title, ": "); i >= 0 {
		title = title[i+2:]
	}
	return capFirst(title)
}

// tierLabel is the card's top line: the mission's name and its kind.
func tierLabel(o game.ExpeditionInfo) string {
	kind := "short"
	switch {
	case o.Tier == game.TierLong:
		kind = "long"
	case o.Grand == game.GrandMilitary:
		kind = "military"
	case o.Tier == game.TierGrand:
		kind = "civil"
	}
	label := o.Label
	if label == "" {
		label = capFirst(o.Tier)
	}
	return label + " · " + kind
}

// goingTo is "Scout to the Copper Hills".
func goingTo(label, place string) string {
	if place == "" {
		return label
	}
	return label + " to the " + place
}

// haveAtLeast reports whether the stock covers a cost.
func haveAtLeast(state game.GameState, res string, amount float64) bool {
	return state.Resources[res].Amount >= amount
}

// refusalText is the card's reason an offer cannot be sent.
func refusalText(o game.ExpeditionInfo) string {
	reason := o.LaunchBlockReason
	if reason == "" {
		reason = "cannot be sent now"
	}
	return capFirst(strings.TrimRight(reason, ".")) + "."
}

// nextBoardChange is the ticks until the board next changes: an offer
// expires or a taken one is replaced. -1 with nothing to wait for.
func nextBoardChange(state game.GameState) int {
	soonest := -1
	note := func(t int) {
		if soonest < 0 || t < soonest {
			soonest = t
		}
	}
	for _, slot := range state.Military.Board {
		if slot.Offer != nil {
			note(slot.Offer.ExpiresTicks)
		} else {
			note(slot.RefillTicks)
		}
	}
	return soonest
}

// societyLine is what the Geographic Society is doing.
func societyLine(state game.GameState) (string, xk) {
	auto := state.Military.AutoExpedition
	switch {
	case auto.Active && auto.Starved:
		return "Geographic Society: cannot pay the expedition cost.", xBad
	case auto.Active:
		return "Geographic Society: next party in " + formatTicks(auto.TicksLeft, state), xAccent
	}
	hint := "A later age brings it."
	if state.Age == "colonial_age" {
		hint = "The Industrial Age brings it."
	}
	return "A Geographic Society scouts automatically. " + hint, xDim
}

// renderExpedition draws the panel w by h.
func renderExpedition(state game.GameState, v expView, w, h int) *xgrid {
	g := newXGrid(w, h)
	l := layoutFor(w, h)
	v.frame = max(v.frame, 0)
	g.box(0, 0, w, h, xBorder, "Expeditions · "+state.AgeName, xTitle, false)
	if l.small {
		g.put(2, 2, fitText("The window is too small for the expeditions panel. Esc closes it.", w-4), xBad)
		return g
	}
	sc := buildScene(state, l.mapW-2, l.mainH-2)
	header(g, state, v, l)
	if v.report {
		drawReport(g, state, v, l, sc)
	} else {
		drawReady(g, state, v, l, sc)
	}
	if v.flash != "" {
		kind := xBad
		if v.flashGood {
			kind = xGood
		}
		g.fill(1, l.footY, w-2, 1)
		g.putFit(2, l.footY, w-4, v.flash, kind)
	}
	if v.confirm && !v.report {
		drawConfirm(g, state, v, sc)
	}
	return g
}

func header(g *xgrid, state game.GameState, v expView, l expLayout) {
	ready, report := xDim, xDim
	if v.report {
		report = xSel
	} else {
		ready = xSel
	}
	x := g.put(2, 1, " READY ROOM ", ready)
	x = g.put(x+1, 1, " EXPEDITION REPORT ", report)
	if mil := state.Military; !v.report && mil.ReturnsMade > mil.ReturnsSeen {
		g.put(x+1, 1, fmt.Sprintf("● %d new", mil.ReturnsMade-mil.ReturnsSeen), xLegend)
	}
	home := FormatNumber(float64(state.Military.SoldierCount))
	g.segs(l.w-3-len([]rune("Soldiers "+home+" at home"))+1, 1, xseg{"Soldiers ", xDim}, xseg{home, xText}, xseg{" at home", xDim})
}

// drawMap draws the terrain, the town, the routes and the pins inside the
// map box. hi is the highlighted pin (an offer's index in the Ready Room,
// a party's in the Report; -1 for none).
func drawMap(g *xgrid, state game.GameState, v expView, l expLayout, sc *expScene, title string, hi int) {
	g.box(1, l.mainY, l.mapW, l.mainH, xBorder, title, xTitle, false)
	ox, oy := 2, l.mainY+1
	t := sc.terrain
	for y := 0; y < t.h; y++ {
		for x := 0; x < t.w; x++ {
			g.set(ox+x, oy+y, t.r[y*t.w+x], t.k[y*t.w+x])
		}
	}
	abs := func(p pt) pt { return pt{ox + p.x, oy + p.y} }
	dot := func(i int) rune {
		if (i+v.frame)%3 == 0 {
			return '•'
		}
		return '·'
	}
	// Missions out: the route dotted faintly, the party at its progress, the
	// goal a diamond. In the Report the highlighted one's route lights up.
	for i, a := range sc.acts {
		if !a.placed {
			continue
		}
		route := routeCells(sc.town, a.at)
		lit := v.report && i == hi
		for j, c := range route[1 : len(route)-1] {
			p := abs(c)
			if lit {
				g.set(p.x, p.y, dot(j), xAccent)
			} else {
				g.set(p.x, p.y, '·', xFaint)
			}
		}
		if len(route) >= 3 {
			progress := 0.0
			if a.act.TotalTicks > 0 {
				progress = 1 - float64(a.act.TicksLeft)/float64(a.act.TotalTicks)
			}
			k := clamp(int(math.Round(progress*float64(len(route)-1))), 1, len(route)-2)
			p := abs(route[k])
			kind := xGood
			if lit && v.frame%2 == 1 {
				kind = xBright
			}
			g.set(p.x, p.y, '@', kind)
		}
		p := abs(a.at)
		kind := xGood
		if lit {
			kind = xTitle
		}
		g.set(p.x, p.y, '◆', kind)
	}
	if !v.report && hi >= 0 && hi < len(sc.offers) && sc.offers[hi].placed {
		route := routeCells(sc.town, sc.offers[hi].at)
		for j, c := range route[1:max(len(route)-1, 1)] {
			p := abs(c)
			g.set(p.x, p.y, dot(j), xAccent)
		}
	}
	tp := abs(sc.town)
	g.set(tp.x, tp.y, '▣', xBright)
	if !v.report {
		for i, o := range sc.offers {
			if !o.placed {
				continue
			}
			p := abs(o.at)
			kind := xTitle
			switch {
			case !o.offer.CanLaunch:
				kind = xDim
			case o.rare && v.frame%2 == 1:
				kind = xLegend2
			case o.rare:
				kind = xLegend
			case o.military():
				kind = xBad
			}
			if i == hi {
				switch {
				case v.frame%2 == 1 && o.rare:
					kind = xSelGold
				case v.frame%2 == 1:
					kind = xSel
				default:
					kind = xBright
				}
			}
			g.set(p.x, p.y, o.glyph(), kind)
		}
	}
	// The highlighted place's name beside its pin.
	var at pt
	var place string
	switch {
	case !v.report && hi >= 0 && hi < len(sc.offers) && sc.offers[hi].placed:
		at, place = abs(sc.offers[hi].at), sc.offers[hi].offer.Place
	case v.report && hi >= 0 && hi < len(sc.acts) && sc.acts[hi].placed:
		at, place = abs(sc.acts[hi].at), sc.acts[hi].act.Place
	}
	if place != "" {
		label := " " + place + " "
		lw := len([]rune(label))
		lx := at.x + 2
		switch {
		case lx+lw <= ox+t.w: // to the right of the pin
		case at.x-1-lw >= ox: // to the left
			lx = at.x - 1 - lw
		default: // neither fits: as much as the map holds
			label, lx = fitText(label, t.w), ox
		}
		g.put(lx, at.y, label, xLabel)
	}
}

// expKeyBar writes the key hints along the foot, as many as fit. Each is a
// key and what it does; hint is a last note, dropped first.
func expKeyBar(g *xgrid, l expLayout, keys [][2]string, hint string) {
	x := 2
	room := l.w - 2
	for i, kv := range keys {
		need := len([]rune(kv[0])) + 1 + len([]rune(kv[1]))
		if i > 0 {
			need += 3
		}
		if x+need > room {
			return
		}
		if i > 0 {
			x = g.put(x, l.footY, " · ", xFaint)
		}
		x = g.put(x, l.footY, kv[0], xAccent)
		x = g.put(x, l.footY, " "+kv[1], xDim)
	}
	if hint != "" {
		const lead = " · or type "
		if x+len([]rune(lead))+len([]rune(hint)) <= room {
			x = g.put(x, l.footY, lead, xFaint)
			g.put(x, l.footY, hint, xText)
		}
	}
}

// drawOut draws the Out now box (or line): each mission under way, where to
// and how long it has left.
func drawOut(g *xgrid, state game.GameState, l expLayout, sc *expScene, title string, lines []xline) {
	if l.outH == 1 {
		return
	}
	g.box(1, l.outY, l.w-2, l.outH, xBorder, title, xTitle, false)
	for i := 0; i < l.outH-2 && i < len(lines); i++ {
		x := 3
		for _, p := range lines[i] {
			x = g.put(x, l.outY+1+i, p.s, p.k)
		}
	}
}

// xline is one line of segments.
type xline []xseg

// outNow is the Out now text, wrapped to room cells.
func outNow(state game.GameState, room int) []xline {
	mil := state.Military
	var parts []string
	for _, a := range []*game.ExpeditionSnapshot{mil.ActiveShort, mil.ActiveLong, mil.ActiveGrand} {
		if a != nil {
			parts = append(parts, fmt.Sprintf("%s, back in %s", goingTo(a.Label, a.Place), expDur(a.TicksLeft, state)))
		}
	}
	if len(parts) == 0 {
		return []xline{{{"Nobody is out.", xFaint}}}
	}
	var out []xline
	cur := ""
	for _, p := range parts {
		if cur != "" && len([]rune(cur))+3+len([]rune(p)) > room {
			out = append(out, xline{{cur, xText}})
			cur = p
		} else if cur == "" {
			cur = p
		} else {
			cur += " · " + p
		}
	}
	return append(out, xline{{cur, xText}})
}

func drawReady(g *xgrid, state game.GameState, v expView, l expLayout, sc *expScene) {
	hi := -1
	if len(sc.offers) > 0 {
		hi = sc.selected(v.sel)
	}
	drawMap(g, state, v, l, sc, "The lands of the "+state.AgeName, hi)
	drawBriefing(g, state, v, l, sc, hi)

	room := l.w - 6
	lines := outNow(state, room)
	if l.outH == 1 {
		g.put(2, l.outY, "Out now ", xDim)
		if len(lines) > 0 {
			x := 10
			for _, p := range lines[0] {
				x = g.putFit(x, l.outY, l.w-3-x, p.s, p.k)
			}
		}
	} else {
		drawOut(g, state, l, sc, "Out now", lines)
		if change := nextBoardChange(state); change >= 0 {
			text := " board changes in " + expDur(change, state) + " "
			g.put(l.w-3-len([]rune(text)), l.outY, text, xDim)
		}
	}
	if l.legendY >= 0 {
		expLegend(g, state, l)
	}
	expKeyBar(g, l, [][2]string{{"←↑↓→", "hop between pins"}, {"Enter", "send"}, {"Tab", "report"}, {"Esc", "close"}}, "expedition <name>")
}

// expLegend is the key to the pins, in the age's own names for the missions.
func expLegend(g *xgrid, state game.GameState, l expLayout) {
	names := [4]string{"Short", "Long", "Military", "Civil"}
	if words, ok := state.Ruleset().MissionWords(state.Age); ok {
		names = [4]string{words.Tiers[config.WordsShort], words.Tiers[config.WordsLong], words.Tiers[config.WordsGrandMilitary], words.Tiers[config.WordsGrandCivil]}
	}
	x := 2
	for _, e := range []struct {
		glyph string
		k     xk
		text  string
	}{
		{"▣", xBright, "town"}, {"○", xTitle, names[0]}, {"◇", xTitle, names[1]}, {"▲", xBad, names[2]},
		{"■", xTitle, names[3]}, {"◈", xLegend, "Legendary"}, {"@", xGood, "a party out"},
	} {
		x = g.put(x, l.legendY, e.glyph, e.k)
		x = g.put(x, l.legendY, " "+e.text+"  ", xDim)
	}
}

// cardLine is a line of the briefing and how readily it goes when the card
// is short of rows: the lowest priority goes first, and 0 never goes.
type cardLine struct {
	parts  []xseg
	italic bool
	prio   int
}

// briefing builds the card's lines for an offer.
func briefing(state game.GameState, o game.ExpeditionInfo, cw int) (body, foot []cardLine) {
	add := func(prio int, parts ...xseg) { body = append(body, cardLine{parts: parts, prio: prio}) }
	add(0, xseg{fitText(tierLabel(o), cw), xAccent})
	for _, ln := range limitLines(wrapText(briefTitle(o), cw), 2, cw) {
		add(0, xseg{ln, xBright})
	}
	for i, ln := range limitLines(wrapText(o.Description, cw), 2, cw) {
		body = append(body, cardLine{parts: []xseg{{ln, xGold}}, italic: true, prio: 5 - i})
	}
	add(1)
	if o.Tier == game.TierGrand && o.Grand == game.GrandMilitary {
		kind := xText
		if state.Military.SoldierCount < o.SoldiersNeeded {
			kind = xBad
		}
		add(0, xseg{"Party  ", xDim}, xseg{fmt.Sprintf("%s %s", FormatNumber(float64(o.SoldiersNeeded)), pluralize("soldier", o.SoldiersNeeded)), kind})
	}
	add(0, xseg{"Odds   ", xDim}, xseg{expBar(o.SuccessChance, 10, "█", "░"), oddsKind(o.SuccessChance)},
		xseg{fmt.Sprintf(" %.0f%%", o.SuccessChance*100), oddsKind(o.SuccessChance)})
	add(0, xseg{"Out    ", xDim}, xseg{expDur(o.DurationMin, state) + " and back", xText})
	add(3, xseg{"Ends   ", xDim}, xseg{"in " + expDur(o.ExpiresTicks, state), xText})
	add(0, xseg{"Haul   ", xDim}, xseg{fmt.Sprintf("about %.0f min of income", o.HaulMinutes), xGood})
	var haul []string
	for _, k := range sortedKeysOf(o.ExpectedHaulMin) {
		haul = append(haul, "+"+FormatNumber(o.ExpectedHaulMin[k])+" "+resourceDisplayName(state, k))
	}
	for i := 0; i < len(haul); i += 2 {
		row := haul[i]
		if i+1 < len(haul) {
			row = fmt.Sprintf("%-13s%s", haul[i], haul[i+1])
		}
		prio := 0
		if i > 0 {
			prio = 6
		}
		add(prio, xseg{"       ", xDim}, xseg{fitText(row, cw-7), xGood})
	}
	costKeys := sortedKeysOf(o.Cost)
	for i, k := range costKeys {
		label := "       "
		if i == 0 {
			label = "Cost   "
		}
		kind := xGood
		if !haveAtLeast(state, k, o.Cost[k]) {
			kind = xBad
		}
		add(0, xseg{label, xDim}, xseg{fitText(FormatNumber(o.Cost[k])+" "+resourceDisplayName(state, k), cw-7), kind})
	}
	if len(costKeys) == 0 && !(o.Tier == game.TierGrand && o.Grand == game.GrandMilitary) {
		add(0, xseg{"Cost   ", xDim}, xseg{"free", xGood})
	}

	fail := "If it fails, the cost is spent and a third of the haul is kept."
	if o.Rare {
		fail = "Legendary: it pays more, and the road is harder. A third of the haul is kept on a failure."
	}
	for i, ln := range limitLines(wrapText(fail, cw), 2, cw) {
		foot = append(foot, cardLine{parts: []xseg{{ln, xDim}}, prio: 3 - i})
	}
	if o.CanLaunch {
		foot = append(foot, cardLine{parts: []xseg{{" Enter ", xSel}, {" send the party", xText}}})
	} else {
		for _, ln := range limitLines(wrapText(refusalText(o), cw), 2, cw) {
			foot = append(foot, cardLine{parts: []xseg{{ln, xBad}}})
		}
	}
	return body, foot
}

// limitLines keeps the first n of lines, the last ending in an ellipsis if
// others were cut, within w cells.
func limitLines(lines []string, n, w int) []string {
	if len(lines) > n {
		lines = lines[:n]
		lines[n-1] = fitText(lines[n-1]+"…", w)
	}
	return lines
}

// fitCard drops the cheapest lines until body and foot fit in rows rows.
func fitCard(body, foot []cardLine, rows int) ([]cardLine, []cardLine) {
	for len(body)+len(foot) > rows {
		best, inFoot, bestPrio := -1, false, 1<<30
		for i, c := range body {
			if c.prio > 0 && c.prio < bestPrio {
				best, inFoot, bestPrio = i, false, c.prio
			}
		}
		for i, c := range foot {
			if c.prio > 0 && c.prio < bestPrio {
				best, inFoot, bestPrio = i, true, c.prio
			}
		}
		switch {
		case best < 0:
			// Nothing left to drop: the card is cut at the bottom.
			if len(body) > 0 {
				body = body[:len(body)-1]
			} else {
				foot = foot[:len(foot)-1]
			}
		case inFoot:
			foot = append(foot[:best:best], foot[best+1:]...)
		default:
			body = append(body[:best:best], body[best+1:]...)
		}
	}
	return body, foot
}

// drawBriefing draws the card for the highlighted offer.
func drawBriefing(g *xgrid, state game.GameState, v expView, l expLayout, sc *expScene, hi int) {
	x, y, w, h := l.cardX, l.mainY, l.cardW, l.mainH
	cw := w - 2
	if hi < 0 {
		g.box(x, y, w, h, xBorder, "Briefing", xTitle, false)
		for i, ln := range wrapText("No offers on the board. New ones come in a few minutes.", cw) {
			g.put(x+1, y+1+i, ln, xDim)
		}
		return
	}
	o := *sc.offers[hi].offer
	border, title, tk := xBorder, "Briefing", xTitle
	if o.Rare {
		title, tk = "◈ Legendary", xLegend
		border = xLegend
		if v.frame%2 == 1 {
			border, tk = xLegend2, xLegend2
		}
	}
	g.box(x, y, w, h, border, title, tk, false)
	body, foot := briefing(state, o, cw)
	body, foot = fitCard(body, foot, h-2)
	row := y + 1
	for _, c := range body {
		drawCardLine(g, x+1, row, cw, c)
		row++
	}
	row = y + h - 1 - len(foot)
	for _, c := range foot {
		drawCardLine(g, x+1, row, cw, c)
		row++
	}
}

func drawCardLine(g *xgrid, x, y, room int, c cardLine) {
	end := x + room
	for _, p := range c.parts {
		if x >= end {
			return
		}
		s := fitText(p.s, end-x)
		if c.italic {
			x = g.putIt(x, y, s, p.k)
		} else {
			x = g.put(x, y, s, p.k)
		}
	}
}

// drawConfirm is the box that asks before a party goes.
func drawConfirm(g *xgrid, state game.GameState, v expView, sc *expScene) {
	if len(sc.offers) == 0 {
		return
	}
	o := *sc.offers[sc.selected(v.sel)].offer
	w := min(58, g.w-4)
	h := 10
	x, y := (g.w-w)/2, max(1, (g.h-h)/2)
	g.fill(x, y, w, h)
	label := o.Label
	if label == "" {
		label = "party"
	}
	g.box(x, y, w, h, xAccent, "Send this "+label+"?", xTitle, true)
	row := y + 2
	for _, ln := range limitLines(wrapText(briefTitle(o), w-4), 2, w-4) {
		g.put(x+2, row, ln, xBright)
		row++
	}
	row = y + 4
	if o.Tier == game.TierGrand && o.Grand == game.GrandMilitary {
		g.segs(x+2, row, xseg{"Party ", xDim}, xseg{fmt.Sprintf("%s %s", FormatNumber(float64(o.SoldiersNeeded)), pluralize("soldier", o.SoldiersNeeded)), xText},
			xseg{fmt.Sprintf("  (%s stay home)", FormatNumber(float64(max(0, state.Military.SoldierCount-o.SoldiersNeeded)))), xDim})
		row++
	}
	costKeys := sortedKeysOf(o.Cost)
	if len(costKeys) > 0 {
		cx := g.put(x+2, row, "Cost  ", xDim)
		for i, k := range costKeys {
			kind := xGood
			if !haveAtLeast(state, k, o.Cost[k]) {
				kind = xBad
			}
			text := FormatNumber(o.Cost[k]) + " " + resourceDisplayName(state, k)
			if i > 0 {
				cx = g.put(cx, row, "  ", xDim)
			}
			if cx+len([]rune(text)) > x+w-2 {
				break
			}
			cx = g.put(cx, row, text, kind)
		}
		row++
	}
	g.segs(x+2, row, xseg{"Odds  ", xDim}, xseg{fmt.Sprintf("%.0f%%", o.SuccessChance*100), oddsKind(o.SuccessChance)},
		xseg{"   Back in  ", xDim}, xseg{expDur(o.DurationMin, state), xText},
		xseg{"   Haul  ", xDim}, xseg{fmt.Sprintf("about %.0f min", o.HaulMinutes), xGood})
	g.segs(x+2, y+h-2, xseg{" Enter ", xSel}, xseg{" send   ", xText}, xseg{" Esc ", xLabel}, xseg{" back", xText})
}

// returnLine is how a mission that came home is told.
func returnLine(r game.MissionReturn) string {
	name := goingTo(r.Label, r.Place)
	if r.Place == "" && r.Title != "" {
		name = r.Title
	}
	switch {
	case r.Success && r.Haul != "":
		return name + " came back: +" + r.Haul + "."
	case r.Success:
		return name + " came back."
	case r.Haul != "":
		return name + " failed. It kept " + r.Haul + "."
	}
	return name + " failed."
}

func drawReport(g *xgrid, state game.GameState, v expView, l expLayout, sc *expScene) {
	hi := -1
	if n := len(sc.acts); n > 0 {
		hi = clamp(v.party, 0, n-1)
	}
	drawMap(g, state, v, l, sc, "Parties in the field", hi)

	x, y, w, h := l.cardX, l.mainY, l.cardW, l.mainH
	cw := w - 2
	g.box(x, y, w, h, xBorder, fmt.Sprintf("Out · %d", len(sc.acts)), xTitle, false)
	// With no legend row to hold it, the Society's line takes the card's foot.
	soc, socKind := societyLine(state)
	foot := 0
	if l.legendY < 0 {
		foot = 2
		for i, ln := range limitLines(wrapText(soc, cw), foot, cw) {
			g.put(x+1, y+h-1-foot+i, ln, socKind)
		}
	}
	if len(sc.acts) == 0 {
		offers := len(sc.offers)
		for i, ln := range wrapText(fmt.Sprintf("Nobody is out. The Ready Room has %d %s waiting.", offers, pluralize("offer", offers)), cw) {
			g.put(x+1, y+1+i, ln, xDim)
		}
	} else {
		area := h - 2 - foot
		per := 4
		visible := max(1, (area+1)/per)
		first := clamp(hi-visible+1, 0, max(0, len(sc.acts)-visible))
		row := y + 1
		for i := first; i < len(sc.acts) && i < first+visible; i++ {
			a := sc.acts[i].act
			head := "  "
			kind := xBright
			if i == hi {
				head, kind = "▸ ", xSel
			}
			g.put(x+1, row, fmt.Sprintf("%-*s", cw, fitText(head+goingTo(a.Label, a.Place), cw)), kind)
			progress := 0.0
			if a.TotalTicks > 0 {
				progress = 1 - float64(a.TicksLeft)/float64(a.TotalTicks)
			}
			g.segsFit(x+3, row+1, cw-2, xseg{expBar(progress, 14, "▰", "▱"), xGood}, xseg{" back in " + expDur(a.TicksLeft, state), xText})
			parts := []xseg{}
			if a.Soldiers > 0 {
				parts = append(parts, xseg{fmt.Sprintf("%s %s", FormatNumber(float64(a.Soldiers)), pluralize("soldier", a.Soldiers)), xText}, xseg{" · ", xFaint})
			}
			parts = append(parts, xseg{fmt.Sprintf("%.0f%% odds", a.SuccessChance*100), oddsKind(a.SuccessChance)},
				xseg{fmt.Sprintf(" · haul %.0f min", a.HaulMinutes), xDim})
			g.segsFit(x+3, row+2, cw-2, parts...)
			row += per
		}
	}

	// What came home, newest first, the new ones marked.
	rets := state.Military.Returns
	if l.outH == 1 {
		g.put(2, l.outY, "Came home ", xDim)
		if len(rets) == 0 {
			g.put(12, l.outY, "Nothing yet.", xFaint)
		} else {
			mark := "  "
			if v.marked > 0 {
				mark = "● "
			}
			kind := xBad
			if rets[0].Success {
				kind = xGood
			}
			x := g.put(12, l.outY, mark, xLegend)
			g.putFit(x, l.outY, l.w-3-x, returnLine(rets[0]), kind)
		}
	} else {
		g.box(1, l.outY, l.w-2, l.outH, xBorder, "Came home", xTitle, false)
		if len(rets) == 0 {
			g.put(3, l.outY+1, "Nothing has come home yet.", xFaint)
		}
		for i := 0; i < l.outH-2 && i < len(rets); i++ {
			mark := "  "
			if i < v.marked {
				mark = "● "
			}
			kind := xBad
			if rets[i].Success {
				kind = xGood
			}
			x := g.put(3, l.outY+1+i, mark, xLegend)
			g.putFit(x, l.outY+1+i, l.w-5-x, returnLine(rets[i]), kind)
		}
	}
	if l.legendY >= 0 {
		g.putFit(2, l.legendY, l.w-4, soc, socKind)
	}
	keys := [][2]string{{"↑↓", "choose a party"}, {"Tab", "ready room"}, {"Esc", "close"}}
	expKeyBar(g, l, keys, "")
}
