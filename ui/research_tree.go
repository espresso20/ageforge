package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// research_tree.go draws the Research panel: the tech tree as a map.
//
// Everything here is a pure function of a game snapshot, a size and a view
// (what is selected, how far the map is scrolled, which zoom). buildTree
// lays the techs out on a virtual canvas, lanes across and ages down;
// renderTree cuts the view out of it and adds the bars round it. The panel
// (research_panel.go) keeps the view and paints the cells in theme colours.
//
// Nothing below names a lane, an age or a tech: lanes, kinds, emblems,
// prerequisites and what a tech opens all come from the ruleset, so a
// content change redraws the map by adding data.

// tStyle is how a cell is painted. The panel maps each to theme roles.
type tStyle uint8

const (
	tsText    tStyle = iota // body text
	tsDim                   // locked, later, hints
	tsBright                // ready to start
	tsHi                    // in progress, numbers
	tsLane                  // the cell's lane colour (researched)
	tsGold                  // the selected tech's chain, steps still to do
	tsGoldDim               // the chain's steps already done
	tsGood                  // names you hold
	tsSel                   // the selection
	tsChip                  // title and key bars
	tsChipKey               // a key on a bar
	tsChipDim               // quiet text on a bar
	tsBorder                // the card's frame
)

// tCell is one cell of the panel.
type tCell struct {
	r    rune
	st   tStyle
	lane int8 // index into the tree's lanes for tsLane, else -1
}

// tGrid is a block of cells.
type tGrid struct {
	w, h int
	c    []tCell
}

func newTGrid(w, h int) *tGrid {
	g := &tGrid{w: max(w, 0), h: max(h, 0)}
	g.c = make([]tCell, g.w*g.h)
	for i := range g.c {
		g.c[i] = tCell{r: ' ', lane: -1}
	}
	return g
}

func (g *tGrid) in(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

func (g *tGrid) at(x, y int) tCell {
	if !g.in(x, y) {
		return tCell{r: ' ', lane: -1}
	}
	return g.c[y*g.w+x]
}

func (g *tGrid) put(x, y int, r rune, st tStyle, lane int) {
	if g.in(x, y) {
		g.c[y*g.w+x] = tCell{r: r, st: st, lane: int8(lane)}
	}
}

// text writes s from (x, y), one cell a rune, and returns the next x.
func (g *tGrid) text(x, y int, s string, st tStyle, lane int) int {
	for _, r := range s {
		g.put(x, y, r, st, lane)
		x++
	}
	return x
}

// String is the grid as lines of text, trailing spaces kept.
func (g *tGrid) String() string {
	var sb strings.Builder
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			sb.WriteRune(g.c[y*g.w+x].r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// treeGeom is the size of things at one zoom and width.
type treeGeom struct {
	laneW   int  // columns a lane takes
	badgeW  int  // a badge's width (a pill's, in the far zoom)
	badgeH  int  // a badge's rows
	nameH   int  // rows for the name plate over a badge
	linkH   int  // connector rows under a badge
	gutterW int  // the age gutter on the left
	pill    bool // the far zoom: one-row pills with a letterhead
}

func (g treeGeom) pitch() int { return g.nameH + g.badgeH + g.linkH }

// treeGeomFor picks the geometry: big badges from 120 columns, small ones
// under it, pills in the far zoom.
func treeGeomFor(w int, far bool) treeGeom {
	switch {
	case far && w >= 120:
		return treeGeom{laneW: 11, badgeW: 10, badgeH: 1, linkH: 1, gutterW: 6, pill: true}
	case far:
		return treeGeom{laneW: 10, badgeW: 9, badgeH: 1, linkH: 1, gutterW: 6, pill: true}
	case w >= 120:
		return treeGeom{laneW: 18, badgeW: 9, badgeH: 5, nameH: 1, linkH: 2, gutterW: 10}
	}
	return treeGeom{laneW: 15, badgeW: 7, badgeH: 3, nameH: 1, linkH: 2, gutterW: 5}
}

// techMark is a tech's state on the map.
type techMark int

const (
	markDone    techMark = iota // researched
	markRunning                 // being researched
	markReady                   // can be started
	markLocked                  // its age is reached, something it needs is not
	markNext                    // in an age you have not reached, but may see
)

// stateWords says a mark in the status line's words.
func (m techMark) words() string {
	switch m {
	case markDone:
		return "researched"
	case markRunning:
		return "in progress"
	case markReady:
		return "you can start it"
	case markLocked:
		return "waits for what it needs"
	}
	return "in the next age"
}

// treeNode is one tech on the map.
type treeNode struct {
	key    string
	def    config.TechDef
	ts     game.TechState
	lane   int // index into treeModel.lanes
	row    int // row of techs, counted from the top of the map
	mark   techMark
	plan   int    // its place in the build plan, 1 for the first item; 0 when not planned
	opens  bool   // it opens a command or a building
	emblem rune   // the glyph in its badge
	cx     int    // virtual x of its emblem and its lines
	top    int    // virtual y of its first row (the name plate, or the pill)
	needs  []edge // what it needs, each a line into it
}

// edge is one line into a tech: from a tech it needs.
type edge struct {
	from string
	or   bool // one of an either-or group
}

// treeBand is one age on the map.
type treeBand struct {
	age, name   string
	row0, rows  int // its rows of techs
	have, total int
	here, next  bool // the age you are in; an age you have not reached
}

// treeModel is the laid-out tree.
type treeModel struct {
	geom  treeGeom
	lanes []config.TechLaneDef
	nodes []*treeNode
	by    map[string]*treeNode
	bands []treeBand
	w, h  int // the virtual canvas

	known, inReach, nextAge, later int // counts for the title
	laterAges                      int
	lines                          map[[2]int]lineCell
}

func (m *treeModel) bottom(n *treeNode) int { return n.top + m.geom.nameH + m.geom.badgeH - 1 }
func (m *treeModel) left(n *treeNode) int {
	if m.geom.pill {
		return n.cx - 1
	}
	return n.cx - m.geom.badgeW/2
}

// buildTree lays out every tech the player may see. ascii picks emblems
// the plain glyph tier can draw. sel is the selected tech, whose chain of
// prerequisites the lines light up ("" for none).
func buildTree(state game.GameState, geom treeGeom, ascii bool, sel string) *treeModel {
	set := state.Ruleset()
	sight := game.SightOf(&state)
	here, _ := set.Index(state.Age)
	m := &treeModel{geom: geom, by: map[string]*treeNode{}, lines: map[[2]int]lineCell{}}

	opens := map[string]bool{}
	for _, b := range set.Buildings() {
		// A keystone has its star; the mark is for a building it opens.
		if b.RequiredTech != "" && b.Category != "wonder" {
			opens[b.RequiredTech] = true
		}
	}
	planned := map[string]int{}
	for i, it := range state.Plan {
		if it.Kind == game.PlanResearch && planned[it.Key] == 0 {
			planned[it.Key] = i + 1
		}
	}
	laneAt := map[string]int{}
	laneOf := func(key string) int {
		if i, ok := laneAt[key]; ok {
			return i
		}
		def := config.TechLaneDef{Key: key, Name: strings.ReplaceAll(key, "_", " "), Emblem: "*"}
		for _, l := range set.TechLanes() {
			if l.Key == key {
				def = l
			}
		}
		laneAt[key] = len(m.lanes)
		m.lanes = append(m.lanes, def)
		return laneAt[key]
	}
	// A lane gets its column when its first tech comes into sight, and the
	// columns keep the lane table's order.
	var seen []config.TechDef
	for _, age := range set.AgeKeys() {
		techs := set.TechsOf(age)
		if !sight.Age(age) {
			m.later += len(techs)
			if len(techs) > 0 {
				m.laterAges++
			}
			continue
		}
		seen = append(seen, techs...)
	}
	for _, l := range set.TechLanes() {
		for _, t := range seen {
			if t.Lane == l.Key {
				laneOf(l.Key)
				break
			}
		}
	}

	row := 0
	for _, age := range set.AgeKeys() {
		techs := set.TechsOf(age)
		if !sight.Age(age) || len(techs) == 0 {
			continue
		}
		at, _ := set.Index(age)
		def, _ := set.Age(age)
		band := treeBand{age: age, name: def.Name, row0: row, total: len(techs), here: at == here, next: at > here}
		// A tech sits below the techs of its own age that it needs.
		inAge := map[string]config.TechDef{}
		for _, t := range techs {
			inAge[t.Key] = t
		}
		depth := map[string]int{}
		var depthOf func(key string, guard int) int
		depthOf = func(key string, guard int) int {
			if d, ok := depth[key]; ok || guard > len(techs) {
				return d
			}
			d := 0
			t := inAge[key]
			for _, p := range append(append([]string(nil), t.Prerequisites...), t.AnyOf...) {
				if _, ok := inAge[p]; ok {
					d = max(d, depthOf(p, guard+1)+1)
				}
			}
			depth[key] = d
			return d
		}
		taken := map[[2]int]bool{}
		for _, t := range techs {
			ts := state.Research.Techs[t.Key]
			n := &treeNode{key: t.Key, def: t, ts: ts, lane: laneOf(t.Lane), plan: planned[t.Key],
				opens: opens[t.Key] || len(set.FeaturesOpenedBy(t.Key)) > 0}
			r := depthOf(t.Key, 0)
			for taken[[2]int{r, n.lane}] {
				r++
			}
			taken[[2]int{r, n.lane}] = true
			n.row = row + r
			band.rows = max(band.rows, r+1)
			switch {
			case ts.Researched:
				n.mark = markDone
				band.have++
				m.known++
			case state.Research.CurrentTech == t.Key:
				n.mark = markRunning
			case band.next:
				n.mark = markNext
			case ts.PrereqsMet:
				n.mark = markReady
			default:
				n.mark = markLocked
			}
			if band.next {
				m.nextAge++
			} else {
				m.inReach++
			}
			n.emblem = techEmblem(t, m.lanes[n.lane], ascii)
			for _, p := range t.Prerequisites {
				n.needs = append(n.needs, edge{from: p})
			}
			for _, p := range t.AnyOf {
				n.needs = append(n.needs, edge{from: p, or: true})
			}
			m.nodes = append(m.nodes, n)
			m.by[t.Key] = n
		}
		row += band.rows
		m.bands = append(m.bands, band)
	}
	for _, n := range m.nodes {
		n.top = n.row * geom.pitch()
		if geom.pill {
			n.cx = n.lane*geom.laneW + 1
		} else {
			n.cx = n.lane*geom.laneW + geom.laneW/2
		}
	}
	sort.SliceStable(m.nodes, func(i, j int) bool {
		if m.nodes[i].row != m.nodes[j].row {
			return m.nodes[i].row < m.nodes[j].row
		}
		return m.nodes[i].lane < m.nodes[j].lane
	})
	m.w = len(m.lanes) * geom.laneW
	m.h = row*geom.pitch() + 1 // and the line that says what waits further on
	m.route(sel)
	return m
}

// techEmblem is the one glyph in a tech's badge: its own, or its lane's.
// The plain tier draws the first letter of the tech's letterhead instead.
func techEmblem(t config.TechDef, lane config.TechLaneDef, ascii bool) rune {
	if ascii {
		for _, r := range techCode(t) {
			return r
		}
		return '*'
	}
	for _, s := range []string{t.Emblem, lane.Emblem} {
		for _, r := range s {
			return r
		}
	}
	return '*'
}

// techCode is a tech's letterhead: its own, or one made from its name.
func techCode(t config.TechDef) string {
	if t.Code != "" {
		return t.Code
	}
	return config.TechCodeFor(t.Name)
}

// ----- lines -----

const (
	dirN = 1 << iota
	dirE
	dirS
	dirW
)

// lineCell is one cell of the connectors: which neighbours it joins.
type lineCell struct {
	bits   uint8
	dashed bool  // part of an either-or line only
	solid  bool  // part of a line the tech needs outright
	level  uint8 // 0 plain, 1 on the selection's chain and done, 2 on it and still to do
}

// route draws a line into every tech from each tech it needs. A line
// leaves the notch under a badge, runs along the first connector row and
// comes down on the name plate. A line that has further to go than the
// next row runs down the edge of the tech's lane, clear of the badges it
// passes. Lines into an age you have not reached are not drawn.
func (m *treeModel) route(sel string) {
	chain := map[string]bool{}
	var walk func(key string)
	walk = func(key string) {
		n := m.by[key]
		if n == nil || chain[key] {
			return
		}
		chain[key] = true
		for _, e := range n.needs {
			walk(e.from)
		}
	}
	if sel != "" {
		walk(sel)
	}
	g := m.geom
	for _, c := range m.nodes {
		if c.mark == markNext {
			continue
		}
		for _, e := range c.needs {
			p := m.by[e.from]
			if p == nil || p.row >= c.row {
				continue
			}
			level := uint8(0)
			if chain[c.key] {
				level = 2
				if p.mark == markDone {
					level = 1
				}
			}
			ya, yb := m.bottom(p)+1, c.top-1
			pts := [][2]int{{p.cx, ya}}
			switch {
			case c.row == p.row+1 && p.cx == c.cx:
			case c.row == p.row+1:
				pts = append(pts, [2]int{c.cx, ya})
			default:
				ch := c.lane * g.laneW
				if g.pill {
					ch += g.laneW - 1
				}
				pts = append(pts, [2]int{ch, ya}, [2]int{ch, yb})
			}
			pts = append(pts, [2]int{c.cx, yb})
			m.line(pts, e.or, level)
		}
	}
}

// line marks the cells of a path through pts, joined end to end, entered
// from above and left downwards.
func (m *treeModel) line(pts [][2]int, dashed bool, level uint8) {
	mark := func(x, y int, bit uint8) {
		c := m.lines[[2]int{x, y}]
		c.bits |= bit
		if dashed {
			c.dashed = true
		} else {
			c.solid = true
		}
		c.level = max(c.level, level)
		m.lines[[2]int{x, y}] = c
	}
	mark(pts[0][0], pts[0][1], dirN)
	for i := 1; i < len(pts); i++ {
		x, y := pts[i-1][0], pts[i-1][1]
		for x != pts[i][0] || y != pts[i][1] {
			nx, ny, out, in := x, y, uint8(0), uint8(0)
			switch {
			case x < pts[i][0]:
				nx, out, in = x+1, dirE, dirW
			case x > pts[i][0]:
				nx, out, in = x-1, dirW, dirE
			case y < pts[i][1]:
				ny, out, in = y+1, dirS, dirN
			default:
				ny, out, in = y-1, dirN, dirS
			}
			mark(x, y, out)
			mark(nx, ny, in)
			x, y = nx, ny
		}
	}
	last := pts[len(pts)-1]
	mark(last[0], last[1], dirS)
}

// lineRune is the box-drawing rune for a cell's joins. A straight run that
// only either-or lines use is dashed; corners and joins are always solid.
func lineRune(c lineCell) rune {
	dash := c.dashed && !c.solid
	switch c.bits {
	case dirN, dirS, dirN | dirS:
		if dash {
			return '┆'
		}
		return '│'
	case dirE, dirW, dirE | dirW:
		if dash {
			return '┄'
		}
		return '─'
	case dirN | dirE:
		return '╰'
	case dirN | dirW:
		return '╯'
	case dirS | dirE:
		return '╭'
	case dirS | dirW:
		return '╮'
	case dirN | dirS | dirE:
		return '├'
	case dirN | dirS | dirW:
		return '┤'
	case dirE | dirW | dirS:
		return '┬'
	case dirE | dirW | dirN:
		return '┴'
	}
	return '┼'
}

// ----- drawing the map -----

// frameRunes is a badge frame: the four corners, then the flat and the
// upright edge.
type frameRunes struct{ tl, tr, bl, br, h, v rune }

var (
	frameRound  = frameRunes{'╭', '╮', '╰', '╯', '─', '│'}
	frameDouble = frameRunes{'╔', '╗', '╚', '╝', '═', '║'}
)

// badgeRows is a tech's badge as rows of runes, badgeW wide: the frame says
// its kind (rounded for optional, double for the spine and keystones, half
// blocks for a capstone), the edges its state (dashed while it waits, a
// shaded block in an age not reached, a bar along the bottom while it is
// researched).
func badgeRows(g treeGeom, n *treeNode, progress float64) []string {
	big := g.badgeH == 5
	e := string(n.emblem)
	if n.mark == markNext {
		if big {
			return []string{" ░░░░░░░ ", "░░░░░░░░░", "░░░ " + e + " ░░░", "░░░░░░░░░", " ░░░▾░░░ "}
		}
		return []string{"░░░░░░░", "░░ " + e + " ░░", "░░░▾░░░"}
	}
	if n.def.Capstone {
		if big {
			return []string{" ▄▄▄▄▄▄▄ ", "▐▀     ▀▌", "▐   " + e + "   ▌", "▐▄     ▄▌", " ▀▀▀▾▀▀▀ "}
		}
		return []string{"▄▄▄▄▄▄▄", "▐  " + e + "  ▌", "▀▀▀▾▀▀▀"}
	}
	f := frameRound
	if n.ts.Kind == config.TechKeystone || n.ts.Kind == config.TechSpine {
		f = frameDouble
	}
	if n.mark == markLocked {
		f.h, f.v = '┄', '┆'
	}
	h := func(k int) string { return strings.Repeat(string(f.h), k) }
	foot := h(2) + "▾" + h(2)
	if n.mark == markRunning {
		done := int(math.Round(5 * progress))
		done = min(max(done, 0), 5)
		foot = strings.Repeat("▓", done) + strings.Repeat("░", 5-done)
	}
	if big {
		return []string{
			" " + string(f.tl) + h(5) + string(f.tr) + " ",
			string(f.tl) + string(f.br) + "     " + string(f.bl) + string(f.tr),
			string(f.v) + "   " + e + "   " + string(f.v),
			string(f.bl) + string(f.tr) + "     " + string(f.tl) + string(f.br),
			" " + string(f.bl) + foot + string(f.br) + " ",
		}
	}
	top := h(5)
	if n.ts.Kind == config.TechKeystone {
		top = "★" + h(4)
	}
	return []string{string(f.tl) + top + string(f.tr), string(f.v) + "  " + e + "  " + string(f.v), string(f.bl) + foot + string(f.br)}
}

// markStyle is the style a tech's badge and name take from its state.
func markStyle(m techMark) tStyle {
	switch m {
	case markDone:
		return tsLane
	case markRunning:
		return tsHi
	case markReady:
		return tsBright
	}
	return tsDim
}

// nameSpan is where a tech's name plate sits: its first column and the
// name, cut to the lane with an ellipsis when it is too long.
func (m *treeModel) nameSpan(n *treeNode) (x int, name string) {
	rs := []rune(n.def.Name)
	if most := m.geom.laneW - 1; len(rs) > most {
		rs = append(rs[:most-1], '…')
	}
	x = max(n.cx-len(rs)/2, n.lane*m.geom.laneW+1)
	return x, string(rs)
}

// progressOf is how far the research in progress has come, 0 to 1.
func progressOf(state game.GameState) float64 {
	if state.Research.TotalTicks <= 0 {
		return 0
	}
	return 1 - float64(state.Research.TicksLeft)/float64(state.Research.TotalTicks)
}

// canvas draws the whole tree on its virtual canvas: the lines first, then
// every tech over them.
func (m *treeModel) canvas(state game.GameState, sel string) *tGrid {
	g := newTGrid(m.w, m.h)
	geom := m.geom
	// The border of the ages you have not reached: a dotted rule.
	for _, b := range m.bands {
		if b.next && b.row0 > 0 {
			for x := 0; x < m.w; x++ {
				g.put(x, b.row0*geom.pitch()-1, '┄', tsDim, -1)
			}
			break
		}
	}
	type label struct{ x, y int }
	var ors []label
	for at, c := range m.lines {
		st := tsDim
		switch c.level {
		case 1:
			st = tsGoldDim
		case 2:
			st = tsGold
		}
		g.put(at[0], at[1], lineRune(c), st, -1)
		// A long dashed run says "or" where it starts.
		if c.dashed && !c.solid && c.bits == dirE|dirW {
			if l := m.lines[[2]int{at[0] - 1, at[1]}]; !(l.dashed && !l.solid && l.bits == dirE|dirW) {
				ors = append(ors, label{at[0], at[1]})
			}
		}
	}
	for _, l := range ors {
		run := 0
		for {
			c := m.lines[[2]int{l.x + run, l.y}]
			if !(c.dashed && !c.solid && c.bits == dirE|dirW) {
				break
			}
			run++
		}
		if run >= 6 {
			g.text(l.x+run/2-2, l.y, " or ", tsText, -1)
		}
	}
	for _, n := range m.nodes {
		st := markStyle(n.mark)
		left := m.left(n)
		if geom.pill {
			l, r := ' ', ' '
			switch n.mark {
			case markDone:
				l = '✓'
			case markRunning:
				l = '⟳'
			case markReady:
				l, r = '(', ')'
			case markLocked:
				l, r = '┆', '┆'
			case markNext:
				l, r = '░', '░'
			}
			if n.plan > 0 && n.mark != markDone && n.mark != markRunning {
				// In the plan: the plan's arrow leads, and what the tech
				// still waits for keeps its closing mark.
				l = '▸'
				if n.mark == markReady {
					r = ' '
				}
			}
			code := []rune(techCode(n.def))
			room := geom.badgeW - 3
			if n.ts.Kind == config.TechKeystone {
				room--
			}
			if len(code) > room {
				code = code[:room]
			}
			pst := st
			if n.key == sel {
				pst = tsSel
			}
			g.put(left, n.top, l, pst, n.lane)
			g.put(left+1, n.top, n.emblem, pst, n.lane)
			g.put(left+2, n.top, r, pst, n.lane)
			x := g.text(left+3, n.top, string(code), pst, n.lane)
			if n.ts.Kind == config.TechKeystone {
				g.put(x, n.top, '★', tsHi, -1)
			}
			continue
		}
		nx, name := m.nameSpan(n)
		nst := st
		if n.key == sel {
			nst = tsSel
		}
		g.text(nx, n.top, name, nst, n.lane)
		rows := badgeRows(geom, n, progressOf(state))
		for dy, row := range rows {
			g.text(left, n.top+geom.nameH+dy, row, st, n.lane)
		}
		y0, yl := n.top+geom.nameH, n.top+geom.nameH+geom.badgeH-1
		big := geom.badgeH == 5
		if big && n.ts.Kind == config.TechKeystone && n.mark != markNext {
			g.put(left, y0, '★', tsHi, -1)
		}
		mx := left + geom.badgeW
		if big {
			mx--
		}
		switch {
		case n.mark == markDone:
			g.put(mx, y0, '✓', tsLane, n.lane)
		case n.mark == markRunning:
			g.put(mx, y0, '⟳', tsHi, -1)
		case n.plan > 0:
			g.text(mx, y0, "▸"+fmt.Sprint(n.plan), tsHi, -1)
		}
		if n.opens {
			g.put(mx, yl, '◇', st, n.lane)
		}
	}
	return g
}

// ----- the panel around the map -----

// treeView is what the player is looking at.
type treeView struct {
	sel    string // the selected tech
	vx, vy int    // the map's scroll
	far    bool   // the far zoom
	card   bool   // the selected tech's card is open
	note   string // what the last action said, on the card
	ascii  bool   // the plain glyph tier
	prompt bool   // something is typed in the command bar
}

// mapRect is where the map sits in a panel of w by h: under the title and
// the lane row, over the status line and the key bar, right of the gutter
// and left of the scroll bar.
func mapRect(geom treeGeom, w, h int) (x, y, mw, mh int) {
	return geom.gutterW, 2, max(w-geom.gutterW-1, 0), max(h-4, 0)
}

// follow moves the view so the selected tech, the lines under it and a cell
// either side of it are on screen. It moves as little as it can.
func (m *treeModel) follow(v *treeView, w, h int) {
	_, _, mw, mh := mapRect(m.geom, w, h)
	if n := m.by[v.sel]; n != nil {
		x0, x1 := n.lane*m.geom.laneW, (n.lane+1)*m.geom.laneW-1
		y0, y1 := n.top, m.bottom(n)+m.geom.linkH
		if y1-y0 >= mh {
			y1 = y0 + mh - 1
		}
		if x1-x0 >= mw {
			x1 = x0 + mw - 1
		}
		v.vx = min(v.vx, x0)
		v.vx = max(v.vx, x1-mw+1)
		v.vy = min(v.vy, y0)
		v.vy = max(v.vy, y1-mh+1)
	}
	v.vx = max(min(v.vx, m.w-mw), 0)
	v.vy = max(min(v.vy, m.h-mh), 0)
}

// shortAge is an age's name for the gutter: capitals, without " Age", cut
// to room.
func shortAge(name string, room int) string {
	rs := []rune(strings.ToUpper(strings.TrimSuffix(name, " Age")))
	if len(rs) > room {
		rs = rs[:room]
	}
	return string(rs)
}

// renderTree draws the whole panel at w by h: the title bar, the lane row,
// the gutter, the map in view with its scroll bar, the status line and the
// key bar, and the selected tech's card over the map when it is open.
func renderTree(state game.GameState, m *treeModel, v treeView, w, h int) *tGrid {
	out := newTGrid(w, h)
	if w < 20 || h < 8 {
		out.text(0, 0, "Research: make the window larger.", tsText, -1)
		return out
	}
	geom := m.geom
	ox, oy, mw, mh := mapRect(geom, w, h)
	canvas := m.canvas(state, v.sel)
	for y := 0; y < mh; y++ {
		for x := 0; x < mw; x++ {
			c := canvas.at(v.vx+x, v.vy+y)
			if v.card && c.st != tsSel {
				c.st = tsDim
			}
			out.c[(oy+y)*w+ox+x] = c
		}
	}
	// What waits further on, on the canvas's last line.
	if fy := m.h - 1 - v.vy; fy >= 0 && fy < mh {
		out.text(ox, oy+fy, clipRunes(m.beyondLine(), mw), tsDim, -1)
	}

	// The title bar.
	for x := 0; x < w; x++ {
		out.put(x, 0, ' ', tsChip, -1)
		out.put(x, h-1, ' ', tsChip, -1)
	}
	know := state.Resources["knowledge"]
	right := fmt.Sprintf("Knowledge %s %s/tick ", FormatNumber(know.Amount), textfmt.RateValue(know.Rate))
	if w < 100 {
		right = fmt.Sprintf("%s %s/t ", FormatNumber(know.Amount), textfmt.RateValue(know.Rate))
	}
	title := fmt.Sprintf(" RESEARCH  %d of %d known", m.known, m.inReach)
	if w >= 100 {
		if m.nextAge > 0 {
			title += fmt.Sprintf(" · %d next age", m.nextAge)
		}
		if m.later > 0 {
			title += fmt.Sprintf(" · %d beyond", m.later)
		}
	} else {
		title = fmt.Sprintf(" RESEARCH %d/%d", m.known, m.inReach)
	}
	x := out.text(0, 0, clipRunes(title, w), tsChip, -1)
	if cur := state.Research.CurrentTech; cur != "" {
		pct := progressOf(state)
		done := int(math.Round(10 * pct))
		run := fmt.Sprintf("  ⟳ %s %s%s %d%% · %s left", state.Research.CurrentTechName,
			strings.Repeat("▓", done), strings.Repeat("░", 10-done), int(math.Round(100*pct)), formatTicks(state.Research.TicksLeft, state))
		if x+len([]rune(run))+len([]rune(right)) <= w {
			out.text(x, 0, run, tsChipKey, -1)
		}
	}
	if rx := w - len([]rune(right)); rx > x {
		out.text(rx, 0, right, tsChip, -1)
	}

	// The lane row: each lane in view over its column, and an arrow at
	// either end for the lanes off screen that way. A lane cut by the
	// right edge keeps a shortened name while there is room for one.
	offLeft, offRight, nearLeft := 0, 0, -1
	edge := ox + mw // one past the map's last column
	for i, l := range m.lanes {
		lx := ox + i*geom.laneW - v.vx + 1
		if lx < ox {
			offLeft, nearLeft = offLeft+1, i
			continue
		}
		room := min(geom.laneW-1, edge-lx)
		if i < len(m.lanes)-1 {
			// Keep clear of the corner the right arrow is drawn in.
			room = min(room, edge-4-lx)
		}
		label := ""
		if v.ascii {
			label = laneLabel(l.Name, room)
		} else if short := laneLabel(l.Name, room-2); short != "" {
			label = l.Emblem + " " + short
		}
		if label == "" {
			offRight++
			continue
		}
		out.text(lx, 1, label, tsLane, i)
	}
	if offLeft > 0 {
		// The nearest lane off to the left is named when the gutter has
		// room for it; a narrow gutter carries the count alone.
		s := "◂"
		if offLeft > 1 {
			s = fmt.Sprintf("◂%d", offLeft)
		}
		if name := laneLabel(m.lanes[nearLeft].Name, ox-2-len([]rune(s))); name != "" {
			s += " " + name
		} else {
			s = fmt.Sprintf("◂%d", offLeft)
		}
		out.text(0, 1, clipRunes(s, ox), tsHi, -1)
	}
	if offRight > 0 {
		s := fmt.Sprintf("%d▸", offRight)
		out.text(w-1-len([]rune(s)), 1, s, tsHi, -1)
	}

	// The gutter: each age in view named beside its first row on screen,
	// with how many of its techs you hold.
	pitch := geom.pitch()
	for _, b := range m.bands {
		top, bot := b.row0*pitch-v.vy, (b.row0+b.rows)*pitch-v.vy
		st := tsText
		if b.next {
			st = tsDim
		} else if b.here {
			st = tsBright
		}
		for y := max(top, 0); y < min(bot, mh); y++ {
			out.put(0, oy+y, '▌', st, -1)
		}
		y := max(top, 0)
		if y >= min(bot, mh) {
			continue
		}
		out.text(1, oy+y, shortAge(b.name, ox-2), st, -1)
		tally := fmt.Sprintf("%d/%d", b.have, b.total)
		switch {
		case b.here && ox >= 8:
			tally += " now"
		case b.next && ox >= 8:
			tally = "next"
		}
		if y+1 < min(bot, mh) && !geom.pill {
			out.text(1, oy+y+1, clipRunes(tally, ox-2), tsDim, -1)
		}
	}
	if v.vy > 0 {
		out.put(ox-1, oy, '▲', tsHi, -1)
	}
	if m.h-v.vy > mh {
		out.put(ox-1, oy+mh-1, '▼', tsHi, -1)
	}
	// The scroll bar: where the view sits on the map, top to bottom.
	if mh > 0 {
		a, z := 0, mh
		if m.h > mh {
			a = v.vy * mh / m.h
			z = max((v.vy+mh)*mh/m.h, a+1)
		}
		for y := 0; y < mh; y++ {
			r := '░'
			if y >= a && y < z {
				r = '█'
			}
			out.put(w-1, oy+y, r, tsDim, -1)
		}
	}

	// The status line and the key bar.
	out.text(0, h-2, clipRunes(m.statusLine(state, v), w), tsText, -1)
	keys := [][2]string{{"←↑↓→", "move"}, {"Tab", "next you can start"}, {"PgUp PgDn", "zoom"}, {"Home", "your age"}, {"Enter", "card"}, {"Esc", "close"}}
	if v.ascii {
		keys[0][0] = "Arrows"
	}
	if w < 100 {
		keys = [][2]string{keys[0], {"Tab", "next"}, keys[2], keys[4], keys[5]}
	}
	switch {
	case v.prompt:
		keys = [][2]string{{"Enter", "runs the command below"}, {"Esc", "close"}}
	case v.card:
		keys = [][2]string{{"Enter", "does what the card says"}, {"Esc", "closes the card"}}
	}
	x = 1
	for _, k := range keys {
		if x+len([]rune(k[0]))+1+len([]rune(k[1])) > w {
			break
		}
		x = out.text(x, h-1, k[0], tsChipKey, -1)
		x = out.text(x+1, h-1, k[1], tsChipDim, -1) + 2
	}
	if v.card {
		if n := m.by[v.sel]; n != nil {
			m.drawCard(out, state, n, v, ox, oy, mw, mh)
		}
	}
	return out
}

// clipRunes cuts s to at most n cells.
// laneLabel is a lane's name for the lane row: its first word in capitals,
// shortened with a full stop when the room is narrower than the word
// ("KNOWL.", "MATER."), and "" when not even four letters of it fit.
func laneLabel(name string, room int) string {
	name = strings.ToUpper(name)
	if j := strings.IndexAny(name, " &"); j > 0 {
		name = name[:j]
	}
	rs := []rune(name)
	if len(rs) <= room {
		return name
	}
	if room < 5 {
		return ""
	}
	rs = rs[:room-1]
	for len(rs) > 4 && strings.ContainsRune("AEIOU", rs[len(rs)-1]) {
		rs = rs[:len(rs)-1]
	}
	return string(rs) + "."
}

func clipRunes(s string, n int) string {
	rs := []rune(s)
	if n < 0 {
		n = 0
	}
	if len(rs) > n {
		rs = rs[:n]
	}
	return string(rs)
}

// beyondLine says what the map does not draw: the techs of later ages are
// counted, never named.
func (m *treeModel) beyondLine() string {
	if m.later == 0 {
		return ""
	}
	return fmt.Sprintf("┄┄┄ %s in later ages ┄┄┄", textfmt.Count(m.later, "more tech waits", "more techs wait"))
}

// costText is a tech's price and time as the status line says them.
func costText(state game.GameState, n *treeNode) string {
	return fmt.Sprintf("%s knowledge, %s", FormatNumber(n.ts.Cost), formatTicks(researchTicks(n.def.ResearchTicks, state), state))
}

// statusLine names the selected tech, its state, its price and what it
// opens.
func (m *treeModel) statusLine(state game.GameState, v treeView) string {
	n := m.by[v.sel]
	if n == nil {
		return " No tech in sight yet."
	}
	parts := []string{" ▸ " + n.def.Name, n.mark.words()}
	if n.ts.Kind == config.TechKeystone {
		parts = append(parts, "keystone")
	}
	if n.mark != markDone {
		parts = append(parts, costText(state, n))
	}
	if n.plan > 0 {
		parts = append(parts, fmt.Sprintf("item %d in your plan", n.plan))
	}
	if what := opensText(state, n.def); what != "" {
		parts = append(parts, what)
	}
	if !v.card {
		parts = append(parts, "Enter opens its card")
	}
	s := strings.Join(parts, " · ")
	if v.ascii {
		s = strings.ReplaceAll(strings.ReplaceAll(s, "▸", ">"), "·", "-")
	}
	return s
}

// opensText is what a tech opens, in a few words: commands, buildings, a
// wonder. "" when it opens nothing.
func opensText(state game.GameState, def config.TechDef) string {
	set := state.Ruleset()
	var out []string
	for _, l := range set.FeaturesOpenedBy(def.Key) {
		out = append(out, l.Opens)
	}
	var blds []string
	wonder := ""
	for _, b := range set.Buildings() {
		if b.RequiredTech != def.Key {
			continue
		}
		if b.Category == "wonder" {
			wonder = b.Name
		} else {
			blds = append(blds, "the "+b.Name)
		}
	}
	if len(blds) > 0 {
		sort.Strings(blds)
		out = append(out, "opens "+joinAnd(blds))
	}
	if wonder != "" {
		out = append(out, "the "+wonder+" needs it")
	}
	return strings.Join(out, ", ")
}

// joinAnd joins names as a list: "A", "A and B", "A, B and C".
func joinAnd(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// ----- the card -----

// Inline style switches in card text: a rune from here on is drawn in
// that style, until the next switch. They take no room.
const (
	swText   = '\x01'
	swBright = '\x02'
	swHi     = '\x03'
	swGood   = '\x04'
	swDim    = '\x05'
)

func swStyle(r rune) (tStyle, bool) {
	switch r {
	case swText:
		return tsText, true
	case swBright:
		return tsBright, true
	case swHi:
		return tsHi, true
	case swGood:
		return tsGood, true
	case swDim:
		return tsDim, true
	}
	return tsText, false
}

// cardAction is what Enter does on a tech's card.
type cardAction int

const (
	actNone  cardAction = iota // nothing to do: researched, or in progress
	actStart                   // start the research now
	actPlan                    // add it to the build plan
)

// cardActionFor decides the card's action: start a tech that can start
// now, plan one that cannot.
func cardActionFor(state game.GameState, n *treeNode) cardAction {
	switch {
	case n.mark == markDone || n.mark == markRunning:
		return actNone
	case n.plan > 0:
		return actNone
	case n.mark == markReady && state.Research.CurrentTech == "" && state.Resources["knowledge"].Amount >= n.ts.Cost:
		return actStart
	}
	return actPlan
}

// cardSentences is the card's text for a tech, a paragraph an entry: what
// it is and does, what it costs, where it stands in the tree, and what
// Enter does. It reads as sentences, and names no tech of an age the player
// may not see.
func (m *treeModel) cardSentences(state game.GameState, n *treeNode, note string) []string {
	set := state.Ruleset()
	sw := func(r rune, s string) string { return string(r) + s + string(swText) }
	var out []string

	// What it is, and what it does.
	// A bonus reads "+8% gold production" and is something the tech gives;
	// any other effect is a clause of its own ("buildings cost 3% less").
	var gives, with []string
	for _, e := range n.def.Effects {
		if text := e.Text(); strings.HasPrefix(text, "+") || strings.HasPrefix(text, "-") {
			gives = append(gives, text)
		} else {
			with = append(with, text)
		}
	}
	blurb := strings.TrimSpace(n.def.Description)
	if len(gives) > 0 {
		blurb += " " + sw(swBright, n.def.Name+" gives "+joinAnd(gives)+".")
	}
	if len(with) > 0 {
		blurb += " " + sw(swBright, "With it, "+joinAnd(with)+".")
	}
	for _, l := range set.FeaturesOpenedBy(n.key) {
		blurb += " " + sw(swBright, "It "+l.Opens+".")
	}
	var blds []string
	for _, b := range set.Buildings() {
		switch {
		case b.RequiredTech != n.key:
		case b.Category == "wonder":
			blurb += " " + sw(swBright, "It is this age's keystone: the "+b.Name+" cannot be built without it.")
		default:
			blds = append(blds, "the "+b.Name)
		}
	}
	if len(blds) > 0 {
		sort.Strings(blds)
		blurb += " " + sw(swBright, "It opens "+joinAnd(blds)+".")
	}
	out = append(out, strings.TrimSpace(blurb))

	// What it costs.
	know := state.Resources["knowledge"].Amount
	took := formatTicks(researchTicks(n.def.ResearchTicks, state), state)
	switch n.mark {
	case markDone:
		out = append(out, "You already hold this.")
	case markRunning:
		out = append(out, fmt.Sprintf("It is being researched: %s done, %s left.",
			sw(swHi, fmt.Sprintf("%d%%", int(math.Round(100*progressOf(state))))), sw(swHi, formatTicks(state.Research.TicksLeft, state))))
	default:
		out = append(out, fmt.Sprintf("It costs %s of your %s knowledge and takes %s.",
			sw(swHi, FormatNumber(n.ts.Cost)), sw(swHi, FormatNumber(know)), sw(swHi, took)))
	}

	// Where it stands.
	name := func(key string) string {
		if p := m.by[key]; p != nil && p.mark == markDone {
			return sw(swGood, p.def.Name)
		} else if p != nil {
			return sw(swDim, p.def.Name)
		}
		return "a tech of a later age"
	}
	var path string
	var all []string
	for _, k := range n.def.Prerequisites {
		all = append(all, name(k))
	}
	switch {
	case len(all) == 0 && len(n.def.AnyOf) == 0:
		path = "It needs nothing before it."
	case len(all) > 0:
		path = "It builds on " + joinAnd(all) + "."
	}
	if len(n.def.AnyOf) > 0 {
		var any, held []string
		for _, k := range n.def.AnyOf {
			any = append(any, name(k))
			if p := m.by[k]; p != nil && p.mark == markDone {
				held = append(held, p.def.Name)
			}
		}
		path = strings.TrimSpace(path + " It needs one of " + strings.Join(any, " or ") + ".")
		if len(held) > 0 {
			path += " You hold " + held[0] + "."
		}
	}
	later := 0
	for _, t := range set.Techs() {
		for _, k := range append(append([]string(nil), t.Prerequisites...), t.AnyOf...) {
			if k == n.key {
				later++
				break
			}
		}
	}
	switch later {
	case 0:
	case 1:
		path += " One later tech builds on it."
	default:
		path += fmt.Sprintf(" %s later techs build on it.", textfmt.Capitalize(numberWord(later)))
	}
	out = append(out, path)

	// What Enter does.
	var act string
	switch cardActionFor(state, n) {
	case actStart:
		act = sw(swHi, " Enter ") + "  starts the research."
	case actPlan:
		// What it still needs goes into the plan before it.
		var first []string
		if chain := game.PlanResearchChain(state, n.key); len(chain) > 1 {
			for _, k := range chain[:len(chain)-1] {
				def, _ := set.Tech(k)
				first = append(first, def.Name)
			}
		}
		act = sw(swHi, " Enter ") + "  adds it to your plan."
		if len(first) > 0 {
			act = sw(swHi, " Enter ") + "  adds it to your plan, after " + neededFirst(first) + "."
		}
		switch {
		case n.mark == markNext && len(first) > 0:
			act += " They run in that order, and it starts once you reach the " + m.bandOf(n).name + "."
		case n.mark == markNext:
			act += " It starts once you reach the " + m.bandOf(n).name + " and hold what it needs."
		case len(first) > 0:
			act += " They run in that order, one at a time."
		case n.mark == markLocked:
			act += " It starts once you hold what it needs."
		case state.Research.CurrentTech != "":
			act += fmt.Sprintf(" %s is running, so it starts in %s.", state.Research.CurrentTechName, formatTicks(state.Research.TicksLeft, state))
		default:
			act += " It starts when the knowledge is there."
		}
	default:
		switch {
		case n.plan > 0:
			act = fmt.Sprintf("It is item %d in your plan.", n.plan)
		case n.mark == markRunning:
			act = "Type research cancel to stop it; the knowledge it cost is not returned."
		}
	}
	if note != "" {
		act = strings.TrimSpace(sw(swHi, note) + " " + act)
	}
	if act != "" {
		out = append(out, act)
	}
	return out
}

// neededFirst words the techs a plan command adds before the one asked for,
// in the order they will run: "Fire Mastery, which it needs first", "the 2
// techs it needs first (Primitive Writing, then Mathematics)", or, for a
// long chain, "the 7 techs it needs first (Tool Making, Stoneworking,
// Bronze Working and 4 more)".
func neededFirst(names []string) string {
	switch n := len(names); {
	case n == 0:
		return ""
	case n == 1:
		return names[0] + ", which it needs first"
	case n <= 4:
		return fmt.Sprintf("the %d techs it needs first (%s, then %s)", n, strings.Join(names[:n-1], ", "), names[n-1])
	default:
		return fmt.Sprintf("the %d techs it needs first (%s and %d more)", n, strings.Join(names[:3], ", "), n-3)
	}
}

// numberWord writes a small count as a word.
func numberWord(n int) string {
	words := []string{"no", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return fmt.Sprint(n)
}

func (m *treeModel) bandOf(n *treeNode) treeBand {
	for _, b := range m.bands {
		if n.row >= b.row0 && n.row < b.row0+b.rows {
			return b
		}
	}
	return treeBand{}
}

// wrapStyled breaks styled text into lines of at most w cells, at spaces.
// Style switches take no room and carry over a break.
func wrapStyled(s string, w int) []string {
	var lines []string
	var line, word []rune
	lw, ww := 0, 0
	cur := swText
	lead := cur
	flush := func() {
		lines = append(lines, string(lead)+string(line))
		line, lw, lead = nil, 0, cur
	}
	emit := func() {
		if ww == 0 && len(word) == 0 {
			return
		}
		if lw > 0 && lw+1+ww > w {
			flush()
		}
		if lw > 0 {
			line = append(line, ' ')
			lw++
		}
		line = append(line, word...)
		lw += ww
		word, ww = nil, 0
	}
	for _, r := range s {
		if _, ok := swStyle(r); ok {
			word = append(word, r)
			cur = r
			continue
		}
		if r == ' ' {
			emit()
			continue
		}
		word = append(word, r)
		ww++
	}
	emit()
	if lw > 0 || len(lines) == 0 {
		flush()
	}
	return lines
}

// styledText writes text with inline style switches and returns the next x.
func (g *tGrid) styledText(x, y int, s string) int {
	st := tsText
	for _, r := range s {
		if ns, ok := swStyle(r); ok {
			st = ns
			continue
		}
		g.put(x, y, r, st, -1)
		x++
	}
	return x
}

// largeBadge is the card's badge, 17 by 9: a ring and the tech's emblem.
var largeBadge = []string{
	"    ▄▄█████▄▄    ",
	"  ▄█▓▒░░░░░▒▓█▄  ",
	" █▓░         ░▓█ ",
	"▐█▒           ▒█▌",
	"▐█▒           ▒█▌",
	"▐█▒           ▒█▌",
	" █▓░         ░▓█ ",
	"  ▀█▓▒░░░░░▒▓█▀  ",
	"    ▀▀█████▀▀    ",
}

// kindWords names a tech's kind for the card's caption.
func kindWords(n *treeNode) string {
	switch {
	case n.def.Capstone:
		return "capstone"
	case n.ts.Kind == config.TechKeystone:
		return "keystone"
	case n.ts.Kind == config.TechSpine:
		return "on the spine"
	}
	return "optional"
}

// drawCard draws the selected tech's card over the map area: the large
// badge with the lane and the kind under it, and the card's sentences
// beside it. Under 70 columns the badge is left out.
func (m *treeModel) drawCard(out *tGrid, state game.GameState, n *treeNode, v treeView, ox, oy, mw, mh int) {
	cw := min(mw-2, 82)
	if cw < 30 || mh < 8 {
		return
	}
	badge := cw >= 66 && mh >= 15
	tx, tw := 3, cw-6
	if badge {
		tx, tw = 24, cw-27
	}
	head := []string{string(swBright) + strings.ToUpper(n.def.Name), textfmt.Capitalize(n.mark.words()) + ".", ""}
	var body []string
	sent := m.cardSentences(state, n, v.note)
	for i, s := range sent {
		if i == len(sent)-1 && len(sent) > 3 {
			body = append(body, string(swDim)+strings.Repeat("─", tw))
		} else if i > 0 {
			body = append(body, "")
		}
		body = append(body, wrapStyled(s, tw)...)
	}
	lines := append(head, body...)
	ch := len(lines) + 2
	if badge {
		ch = max(ch, 15)
	}
	if ch > mh {
		ch = mh
		lines = lines[:max(ch-2, 0)]
	}
	x0, y0 := ox+(mw-cw)/2, oy+(mh-ch)/2
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			r := ' '
			switch {
			case y == 0 && x == 0:
				r = '╭'
			case y == 0 && x == cw-1:
				r = '╮'
			case y == ch-1 && x == 0:
				r = '╰'
			case y == ch-1 && x == cw-1:
				r = '╯'
			case y == 0 || y == ch-1:
				r = '─'
			case x == 0 || x == cw-1:
				r = '│'
			}
			out.put(x0+x, y0+y, r, tsBorder, -1)
		}
	}
	out.text(x0+2, y0, clipRunes(" "+n.def.Name+" ", cw-4), tsBright, -1)
	if esc := " Esc closes "; cw > len(esc)+len([]rune(n.def.Name))+8 {
		out.text(x0+cw-2-len([]rune(esc)), y0, esc, tsDim, -1)
	}
	if badge {
		ring := tsLane
		if n.def.Capstone {
			ring = tsHi
		}
		for dy, row := range largeBadge {
			out.text(x0+3, y0+2+dy, row, ring, n.lane)
		}
		out.put(x0+3+8, y0+2+4, n.emblem, tsBright, -1)
		if n.ts.Kind == config.TechKeystone {
			out.put(x0+3+8, y0+2+2, '★', tsHi, -1)
		}
		lane := m.lanes[n.lane].Name
		out.text(x0+3+max((17-len([]rune(lane)))/2, 0), y0+12, clipRunes(lane, 19), tsLane, n.lane)
		kind := kindWords(n)
		out.text(x0+3+max((17-len([]rune(kind)))/2, 0), y0+13, kind, tsDim, -1)
	}
	for i, l := range lines {
		out.styledText(x0+tx, y0+1+i, l)
	}
}

// ----- the plain glyph tier -----

// plainFold is what the plain tier draws for the marks and edges the map
// uses, where the map model's own fold would lose the difference between
// two of them (a dashed edge and a solid one, the filled part of a bar and
// the rest).
var plainFold = map[rune]rune{
	'✓': 'v', '★': '*', '▾': 'v', '⟳': '~', '▸': '>', '◂': '<', '◇': '+', '▲': '^', '▼': 'v',
	'┄': '.', '┆': ':', '▓': '#', '░': '.', '█': '#', '▒': '+', '▌': '|', '▐': '|', '▄': '_', '▀': '"',
	'…': '.', '·': '-', '←': '<', '→': '>', '↑': '^', '↓': 'v',
}

// foldPlain rewrites a grid for the plain glyph tier: nothing but ASCII.
func foldPlain(g *tGrid) {
	for i, c := range g.c {
		if c.r < 0x80 {
			continue
		}
		if r, ok := plainFold[c.r]; ok {
			g.c[i].r = r
		} else {
			g.c[i].r = mapmodel.Fold(c.r, mapmodel.TierASCII)
		}
	}
}
