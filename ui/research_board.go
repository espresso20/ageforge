package ui

import (
	"fmt"
	"image"
	"sort"
	"strings"
	"unicode"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// research_board.go is the Research board: the whole screen but the command
// bar, set out as a table. The tech in work and the queue sit on top, the
// techs that can start are rows with their time, cost and what they open,
// the techs still out of reach follow as a list, and the selected tech has
// its own column with the two buttons. L holds the board to one lane and T
// opens the whole tree (tree_panel.go). Like the Map it leaves the keyboard
// with the command bar and takes only the keys that print nothing; the three
// letters count only as capitals, since commands are typed in lower case.

// boardKeys are the board's keys as the Help panel lists them.
var boardKeys = []keyHelp{
	{"Up/Down", rowid.E("ui.research_board.boardKeys.01", "Pick a tech; the rows that can start come first, then the ones still out of reach")},
	{"Enter", rowid.E("ui.research_board.boardKeys.02", "Research the picked tech, or queue it when one is in work (with something typed, Tab and Enter act on the prompt instead)")},
	{"Q", rowid.E("ui.research_board.boardKeys.03", "Queue the picked tech, with the techs it needs (a capital letter, on an empty prompt)")},
	{"L", rowid.E("ui.research_board.boardKeys.04", "Hold the board to the picked tech's lane, and again to show every lane (a capital letter)")},
	{"T", rowid.E("ui.research_board.boardKeys.05", "Open the whole tree on the picked tech (a capital letter)")},
	{"Esc", rowid.E("ui.research_board.boardKeys.06", "Show every lane again, then close")},
}

// boardTech is a tech in sight with its state in this snapshot.
type boardTech struct {
	def  config.TechDef
	ts   game.TechState
	secs int // the time it takes the scholars researching, 0 with none
}

// boardModel is the board's snapshot of the research: the techs the player
// may see, which can start, which wait, and what leads to what.
type boardModel struct {
	state         game.GameState
	byKey         map[string]*boardTech
	techs         []*boardTech
	ready, coming []*boardTech
	queued        map[string]bool
	leads         map[string][]string
	lanes         []config.TechLaneDef
	laneOf        map[string]config.TechLaneDef
	tick          float64 // seconds a tick takes
	age           int     // the current age's place
}

func buildBoard(state game.GameState) *boardModel {
	set := state.Ruleset()
	sight := game.SightOf(&state)
	rs := state.Research
	m := &boardModel{
		state: state, byKey: map[string]*boardTech{}, queued: map[string]bool{},
		leads: map[string][]string{}, lanes: set.TechLanes(), laneOf: map[string]config.TechLaneDef{},
		tick: tickInterval(state).Seconds(),
	}
	m.age, _ = set.Index(state.Age)
	for _, l := range m.lanes {
		m.laneOf[l.Key] = l
	}
	for _, k := range rs.Queue {
		m.queued[k] = true
	}
	for _, d := range set.Techs() {
		if !sight.Age(d.Age) {
			continue
		}
		t := &boardTech{def: d, ts: rs.Techs[d.Key]}
		t.secs = int(float64(t.ts.TicksNeeded)*m.tick + 0.5)
		m.techs = append(m.techs, t)
		m.byKey[d.Key] = t
	}
	for _, t := range m.techs {
		for _, k := range t.needs() {
			m.leads[k] = append(m.leads[k], t.def.Key)
		}
		switch {
		case t.ts.Researched, t.def.Key == rs.CurrentTech, m.queued[t.def.Key]:
		case t.ts.Available:
			m.ready = append(m.ready, t)
		default:
			m.coming = append(m.coming, t)
		}
	}
	sort.SliceStable(m.ready, func(i, j int) bool { return m.ready[i].secs < m.ready[j].secs })
	sort.SliceStable(m.coming, func(i, j int) bool { return m.coming[i].def.Cost < m.coming[j].def.Cost })
	return m
}

// needs are the keys a tech waits on: its prerequisites and its either-or
// group.
func (t *boardTech) needs() []string {
	return append(append([]string(nil), t.def.Prerequisites...), t.def.AnyOf...)
}

// laneShort is a lane's name in a word: the first one.
func (m *boardModel) laneShort(lane string) string {
	f := strings.Fields(m.laneOf[lane].Name)
	if len(f) == 0 {
		return lane
	}
	return f[0]
}

func (m *boardModel) ageName(key string) string {
	if a, ok := m.state.Ruleset().Age(key); ok {
		return a.Name
	}
	return key
}

// effects is what a tech gives: what it opens, then its own effects.
func (m *boardModel) effects(t *boardTech) []string {
	out := techGates(m.state, t.def)
	for _, e := range t.def.Effects {
		if s := e.Text(); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		out = append(out, "no effect listed")
	}
	return out
}

// missing are the names of what a tech still waits on, without the one it is
// being asked about: techs not researched, and its age when it is ahead.
func (m *boardModel) missing(t *boardTech, except string) []string {
	var out []string
	for _, k := range t.needs() {
		if o := m.byKey[k]; o != nil && k != except && !o.ts.Researched {
			out = append(out, o.def.Name)
		}
	}
	if i, _ := m.state.Ruleset().Index(t.def.Age); i > m.age {
		out = append(out, "the "+m.ageName(t.def.Age))
	}
	return out
}

// hm is a time in seconds as hours and minutes, "1:30".
func hm(secs int) string {
	mins := (secs + 59) / 60
	return fmt.Sprintf("%d:%02d", mins/60, mins%60)
}

// clock is a time in seconds as it counts down: "41:12", "1:10:00".
func clock(secs int) string {
	if secs >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", secs/3600, secs/60%60, secs%60)
	}
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

func timeOf(secs int) string {
	if secs <= 0 {
		return "-:--"
	}
	return hm(secs)
}

// packLines sets the items on at most n lines of w cells: one a line when
// they fit that way, else two on a line. What does not fit is counted at the
// end of the last line.
func packLines(items []string, w, n int) []string {
	var lines []string
	if len(items) <= n {
		for _, it := range items {
			lines = append(lines, fitText(it, w))
		}
		return lines
	}
	for i, it := range items {
		it = fitText(it, w)
		if k := len(lines) - 1; k >= 0 && runeLen(lines[k])+3+runeLen(it) <= w {
			lines[k] += "   " + it
			continue
		}
		if len(lines) == n {
			more := fmt.Sprintf("   +%d more", len(items)-i)
			lines[n-1] = fitText(lines[n-1], w-len(more)) + more
			break
		}
		lines = append(lines, it)
	}
	return lines
}

// listLines sets the names, comma separated, on at most n lines of w cells
// and counts what does not fit at the end of the last line.
func listLines(names []string, w, n int) []string {
	var lines []string
	cur := ""
	for i, s := range names {
		last := i == len(names)-1
		if !last {
			s += ","
		}
		add := s
		if cur != "" {
			add = cur + " " + s
		}
		room := w
		if len(lines) == n-1 && !last {
			room -= 4
		}
		if runeLen(add) <= room {
			cur = add
			continue
		}
		if len(lines) == n-1 {
			return append(lines, strings.TrimRight(cur, ",")+fmt.Sprintf(" +%d", len(names)-i))
		}
		lines = append(lines, cur)
		cur = s
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// ----- the panel -----

type boardView struct {
	sel        string // the picked tech
	lane       string // the lane the board is held to, "" for all
	note       string // what the game answered to the last Enter or Q
	rtop, ctop int    // the first row shown of each list
}

const (
	hitTech = iota
	hitResearch
	hitQueue
	hitLane
)

type boardHit struct {
	r    image.Rectangle
	kind int
	id   string
}

type researchPanel struct {
	barPanel
	kit.Base
	engine *game.GameEngine
	// toTree opens the whole tree on a tech.
	toTree        func(key string)
	model         *boardModel
	ready, coming []*boardTech // what the board shows, held to the lane
	view          boardView
	hits          []boardHit
	w, h          int
}

func newResearchPanel() *researchPanel {
	p := &researchPanel{w: 120, h: 37}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *researchPanel) open(state game.GameState) Primitive {
	p.update(state)
	return overBar(p)
}

// update takes a new snapshot and keeps the pick on a tech the board lists.
func (p *researchPanel) update(state game.GameState) {
	p.model = buildBoard(state)
	p.settle()
}

func (p *researchPanel) settle() {
	m := p.model
	held := func(l []*boardTech) (out []*boardTech) {
		for _, t := range l {
			if p.view.lane == "" || t.def.Lane == p.view.lane {
				out = append(out, t)
			}
		}
		return out
	}
	p.ready, p.coming = held(m.ready), held(m.coming)
	if p.index() < 0 {
		p.view.sel = ""
		if all := p.all(); len(all) > 0 {
			p.view.sel = all[0].def.Key
		}
	}
	p.follow()
}

// all is the techs a pick can land on, in order: the ones that can start,
// then the rest.
func (p *researchPanel) all() []*boardTech {
	return append(append([]*boardTech(nil), p.ready...), p.coming...)
}

func (p *researchPanel) index() int {
	for i, t := range p.all() {
		if t.def.Key == p.view.sel {
			return i
		}
	}
	return -1
}

func (p *researchPanel) selected() *boardTech {
	if p.model == nil {
		return nil
	}
	return p.model.byKey[p.view.sel]
}

// boardGeo is where the parts fall for a panel of w by h.
type boardGeo struct {
	cTech, cLane, cCost, cUnl, cStat, cLead int // cLead is 0 when there is no room for the column
	bw                                      int // width of the cost bar
	rx, rw, tEnd                            int
	detail                                  bool
	yHead, y0, maxReady                     int
	strips, by                              int
	lanes                                   bool
}

func (p *researchPanel) geometry() boardGeo {
	w, h := p.w, p.h
	g := boardGeo{cTech: 15, yHead: 9, y0: 12}
	g.cLane, g.cCost, g.cUnl, g.bw = g.cTech+32, g.cTech+45, g.cTech+61, 13
	if w < 130 {
		// A narrow window drops the leads column and shrinks the rest.
		g.cLane, g.cCost, g.cUnl, g.bw = g.cTech+24, g.cTech+36, g.cTech+47, 8
	}
	g.detail = w >= 150
	g.tEnd = w - 4
	if g.detail {
		g.rw = min(max(w*21/100, 40), 70)
		g.rx = w - 3 - g.rw
		g.tEnd = g.rx - 5
	}
	span := g.tEnd - g.cUnl
	g.cStat, g.cLead = g.cUnl+span*35/100, g.cUnl+span*53/100
	if w < 130 {
		g.cStat, g.cLead = g.tEnd-14, 0
	}
	g.maxReady = min(7, max(2, (h-24)/6))
	g.strips = h - 29
	g.by = g.strips - 5
	g.lanes = g.detail && h >= 60
	if !g.lanes {
		g.by = h - 9
	}
	return g
}

// comingRows is how many rows of the coming list fit under the rows shown
// of the ready list.
func (p *researchPanel) comingRows(g boardGeo) int {
	return max(p.h-4-(g.y0+min(len(p.ready), g.maxReady)*3+2), 1)
}

// follow scrolls the lists to keep the pick in view.
func (p *researchPanel) follow() {
	g := p.geometry()
	i := p.index()
	v := &p.view
	v.rtop = max(min(v.rtop, len(p.ready)-g.maxReady), 0)
	v.ctop = max(min(v.ctop, len(p.coming)-p.comingRows(g)), 0)
	switch {
	case i < 0:
	case i < len(p.ready):
		v.rtop = max(min(v.rtop, i), i-g.maxReady+1)
	default:
		i -= len(p.ready)
		rows := p.comingRows(g)
		v.ctop = max(min(v.ctop, i), i-rows+1)
	}
}

func (p *researchPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if w, h := r.Dx(), r.Dy(); w > 0 && h > 0 {
		p.w, p.h = w, h
		if p.model != nil {
			p.follow()
		}
	}
}

// ----- keys and mouse -----

// routeKey takes a key if it is one of the board's and reports whether it
// did. With something typed in the prompt, Tab and Enter belong to the
// prompt, and the three letters need an empty one.
func (p *researchPanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome, kit.KeyEnd:
		p.handleKey(ev.Code)
	case kit.KeyEnter:
		if prompt != "" {
			return false
		}
		p.handleKey(ev.Code)
	case kit.KeyRune:
		if prompt != "" || ev.Mod&(kit.ModCtrl|kit.ModAlt) != 0 || !unicode.IsUpper(ev.Rune) {
			return false
		}
		switch ev.Rune {
		case 'Q':
			p.act(true)
		case 'L':
			p.holdLane(p.view.lane == "")
		case 'T':
			p.openTree()
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// closeLane shows every lane again and reports whether the board was held
// to one: Esc does that before it closes.
func (p *researchPanel) closeLane() bool {
	if p.view.lane == "" {
		return false
	}
	p.holdLane(false)
	return true
}

// holdLane holds the board to the picked tech's lane, or lets it go.
func (p *researchPanel) holdLane(on bool) {
	p.view.lane = ""
	if t := p.selected(); on && t != nil {
		p.view.lane = t.def.Lane
	}
	p.settle()
}

func (p *researchPanel) openTree() {
	if p.toTree != nil {
		p.toTree(p.view.sel)
	}
}

func (p *researchPanel) handleKey(key kit.KeyCode) {
	if p.model == nil {
		return
	}
	switch key {
	case kit.KeyUp:
		p.move(-1)
	case kit.KeyDown:
		p.move(1)
	case kit.KeyPgUp:
		p.move(-p.comingRows(p.geometry()))
	case kit.KeyPgDn:
		p.move(p.comingRows(p.geometry()))
	case kit.KeyHome:
		p.move(-1 << 20)
	case kit.KeyEnd:
		p.move(1 << 20)
	case kit.KeyEnter:
		p.act(false)
	}
}

func (p *researchPanel) move(d int) {
	all := p.all()
	if len(all) > 0 {
		p.pick(all[clamp(p.index()+d, 0, len(all)-1)].def.Key)
	}
}

func (p *researchPanel) pick(key string) {
	p.view.sel, p.view.note = key, ""
	p.follow()
}

// act starts the picked tech, or queues it: Enter starts one that can start
// while nothing is in work and queues any other; queue always queues. What
// the game answers goes in the note.
func (p *researchPanel) act(queue bool) {
	t := p.selected()
	if t == nil || p.engine == nil {
		return
	}
	var err error
	if queue || !t.ts.Available || p.model.state.Research.CurrentTech != "" {
		err = p.engine.PlanAddResearch(t.def.Key)
	} else {
		err = p.engine.StartResearch(t.def.Key)
	}
	p.view.note = ""
	if err != nil {
		p.view.note = err.Error()
	}
	p.update(p.engine.GetState())
}

func (p *researchPanel) Mouse(m kit.Mouse) bool {
	switch {
	case m.Kind == kit.MouseWheel:
		key := kit.KeyDown
		if m.Wheel < 0 {
			key = kit.KeyUp
		}
		for i := max(m.Wheel, -m.Wheel); i > 0; i-- {
			p.handleKey(key)
		}
		return m.Wheel != 0
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft || p.model == nil:
		return false
	}
	pt := m.Pt().Sub(p.Bounds().Min)
	for _, h := range p.hits {
		if !pt.In(h.r) {
			continue
		}
		switch h.kind {
		case hitTech:
			p.pick(h.id)
			if m.Clicks == 2 {
				p.act(false)
			}
		case hitResearch:
			p.act(false)
		case hitQueue:
			p.act(true)
		case hitLane:
			if p.view.lane == h.id {
				p.view.lane = ""
			} else {
				p.view.lane = h.id
			}
			p.settle()
		}
		return true
	}
	return false
}

// PointerAt: the hand over a row or a button a click would act on.
func (p *researchPanel) PointerAt(at image.Point) kit.Pointer {
	pt := at.Sub(p.Bounds().Min)
	for _, h := range p.hits {
		if pt.In(h.r) {
			return kit.PointerHand
		}
	}
	return kit.PointerArrow
}

// ----- painting -----

func (p *researchPanel) Paint(pt *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.model == nil {
		return
	}
	s := rankPaint{pt.Clip(r), r.Min, r.Dx()}
	m, w, h := p.model, r.Dx(), r.Dy()
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	faint := faintCol()
	s.fill(0, 0, w, h, bg)
	p.hits = p.hits[:0]
	g := p.geometry()
	laneCol := func(lane string) kit.Color { return theme.Hue(m.laneOf[lane].Hue, bg) }
	dimmed := func(c kit.Color, k float64) kit.Color { return theme.Mix(bg, c, k) }
	head := func(x, y int, str string, c kit.Color) int {
		return paintHeading(s.p, s.o.X+x, s.o.Y+y, str, inkOf(c)) - s.o.X
	}

	view := "the board"
	if p.view.lane != "" {
		view = "one lane: " + m.laneOf[p.view.lane].Name
	}
	s.text(2, 0, "RESEARCH", kit.Role(theme.RoleAccent), true)
	s.text(12, 0, view, dim, false)
	rs := m.state.Research
	if sc := m.state.Scholars; m.tick > 0 {
		s.right(w-11, 0, fmt.Sprintf("%d scholars researching   %s work a second", sc.Researching, config.FormatAmount(rs.PerTick/m.tick)), txt, false)
	}
	s.fill(w-8, 0, 5, 1, theme.Of(theme.RoleAccent))
	s.text(w-8, 0, " Esc ", kit.Role(theme.RoleOnAccent), false)

	p.paintTop(s, w, head)

	sel := p.selected()
	var sc kit.Color = kit.Role(theme.RoleAccent)
	if sel != nil {
		sc = laneCol(sel.def.Lane)
	}
	i := p.index()

	// The tint under the picked row and the detail column goes down first.
	tint := theme.Mix(bg, sc, 0.10)
	if g.detail {
		s.fill(g.rx-2, g.yHead-1, w-1-(g.rx-2), max(h-3-(g.yHead-1)-1, 0), tint)
	}
	shown := p.ready[min(p.view.rtop, len(p.ready)):min(p.view.rtop+g.maxReady, len(p.ready))]
	if i >= 0 && i < len(p.ready) && i-p.view.rtop < len(shown) {
		sy := g.y0 + (i-p.view.rtop)*3
		s.fill(1, sy, g.tEnd+1, 2, tint)
		for x := 1; x <= g.tEnd+1; x++ {
			s.p.Put(s.o.X+x, s.o.Y+sy-1, '▄', kit.Style{}.Foreground(tint))
			s.p.Put(s.o.X+x, s.o.Y+sy+2, '▀', kit.Style{}.Foreground(tint))
		}
		for j := 0; j < 2; j++ {
			s.p.Put(s.o.X+1, s.o.Y+sy+j, '▌', kit.Style{}.Foreground(sc))
		}
	}
	for _, c := range []struct {
		x int
		s string
	}{{3, "TIME"}, {g.cTech, "TECH"}, {g.cLane, "LANE"}, {g.cCost, "COST"}, {g.cUnl, "UNLOCKS"}, {g.cStat, "STATUS"}, {g.cLead, "LEADS TO"}} {
		if c.x == 0 {
			continue
		}
		s.text(c.x, g.yHead, c.s, dim, false)
	}
	divider := func(y int, label, right string) {
		s.rule(2, y, g.tEnd-1, faint)
		s.text(4, y, " "+label+" ", txt, true)
		s.right(g.tEnd+1, y, " "+right+" ", dim, false)
	}
	divider(g.yHead+1, "READY NOW", fmt.Sprintf("%d can start", len(p.ready)))
	if len(p.ready) == 0 {
		s.text(g.cTech, g.y0, "nothing can start yet", dim, false)
	}

	maxCost := 1.0
	for _, t := range p.ready {
		maxCost = max(maxCost, t.def.Cost)
	}
	wait := rs.TicksLeft
	for _, k := range rs.Queue {
		if q := m.byKey[k]; q != nil {
			wait += q.ts.TicksNeeded
		}
	}
	for n, t := range shown {
		y := g.y0 + n*3
		on := t.def.Key == p.view.sel
		lc := laneCol(t.def.Lane)
		tc, lead := txt, dim
		if on {
			tc, lead = bright, txt
		}
		p.hits = append(p.hits, boardHit{image.Rect(1, y-1, g.tEnd+2, y+2), hitTech, t.def.Key})
		head(3, y, timeOf(t.secs), tc)
		head(g.cTech, y, clipRunes(t.def.Name, (g.cLane-2-g.cTech)/2), lc)
		s.text(g.cLane, y, m.laneShort(t.def.Lane), lc, false)
		s.text(g.cCost, y, config.FormatAmount(t.def.Cost), bright, true)
		s.bar(g.cCost, y+1, g.bw, t.def.Cost/maxCost, lc, dimmed(lc, 0.2), '▀')
		for j, ln := range packLines(m.effects(t), g.cStat-3-g.cUnl, 2) {
			s.text(g.cUnl, y+j, ln, kit.Role(theme.RoleEffect), false)
		}
		switch {
		case t.secs == 0:
			s.text(g.cStat, y, "NO SCHOLARS", kit.Role(theme.RoleWarning), true)
			s.text(g.cStat, y+1, "none are researching", dim, false)
		default:
			s.text(g.cStat, y, "READY", kit.Role(theme.RolePositive), true)
			if rs.CurrentTech == "" {
				s.text(g.cStat, y+1, "starts now", dim, false)
			} else {
				s.text(g.cStat, y+1, "starts in "+clock(int(float64(wait)*m.tick)), dim, false)
			}
		}
		var on2 []string
		if g.cLead == 0 {
			continue
		}
		for _, k := range m.leads[t.def.Key] {
			on2 = append(on2, m.byKey[k].def.Name)
		}
		if len(on2) == 0 {
			s.text(g.cLead, y, "nothing further", dim, false)
		}
		for j, ln := range listLines(on2, g.tEnd+1-g.cLead, 2) {
			s.text(g.cLead, y+j, ln, lead, false)
		}
	}

	// The coming list: a plain line each, the picked tech lit where it waits.
	cy := g.y0 + len(shown)*3
	rows := p.comingRows(g)
	top := min(p.view.ctop, len(p.coming))
	coming := p.coming[top:min(top+rows, len(p.coming))]
	divider(cy, "COMING UP", fmt.Sprintf("%d of %d", top+len(coming), len(p.coming)))
	for n, t := range coming {
		y := cy + 2 + n
		on := t.def.Key == p.view.sel
		lc := dimmed(laneCol(t.def.Lane), 0.75)
		tc := dim
		if on {
			s.fill(1, y, g.tEnd+1, 1, tint)
			s.p.Put(s.o.X+1, s.o.Y+y, '▌', kit.Style{}.Foreground(sc))
			lc, tc = laneCol(t.def.Lane), txt
		}
		p.hits = append(p.hits, boardHit{image.Rect(1, y, g.tEnd+2, y+1), hitTech, t.def.Key})
		s.text(3, y, timeOf(t.secs), tc, false)
		s.text(g.cTech, y, fitText(t.def.Name, g.cLane-2-g.cTech), lc, on)
		s.text(g.cLane, y, m.laneShort(t.def.Lane), lc, false)
		s.text(g.cCost, y, config.FormatAmount(t.def.Cost), tc, false)
		s.text(g.cUnl, y, fitText(m.effects(t)[0], g.cStat-3-g.cUnl), tc, false)
		x := s.text(g.cStat, y, "needs", dim, false) + 1
		needs := p.needsOf(t)
		for j, n := range needs {
			room, wd := g.tEnd+1-x, runeLen(n.s)
			if j > 0 {
				wd += 2
			}
			if j < len(needs)-1 {
				room -= 4
			}
			if wd > room {
				s.text(x, y, fmt.Sprintf(" +%d", len(needs)-j), dim, false)
				break
			}
			if j > 0 {
				x = s.text(x, y, ",", dim, false) + 1
			}
			x = s.text(x, y, n.s, n.col, n.bold)
		}
	}

	if g.detail && sel != nil {
		p.paintDetail(s, g, sel, sc, head)
	}
	s.rule(2, h-3, w-4, faint)
	s.text(2, h-2, "↑ ↓ row    Enter research    Q queue    L one lane    T the whole tree    Esc close", dim, false)
}

type boardNeed struct {
	s    string
	col  kit.Color
	bold bool
}

// needsOf is what a coming tech waits on, as it reads on its line: the picked
// tech lit, a queued one marked.
func (p *researchPanel) needsOf(t *boardTech) []boardNeed {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	dim := kit.Role(theme.RoleDim)
	var out []boardNeed
	for _, k := range t.needs() {
		o := m.byKey[k]
		switch {
		case o == nil || o.ts.Researched:
		case k == p.view.sel:
			out = append(out, boardNeed{o.def.Name, theme.Hue(m.laneOf[o.def.Lane].Hue, bg), true})
		case m.queued[k] || k == m.state.Research.CurrentTech:
			out = append(out, boardNeed{o.def.Name + " (queued)", kit.Role(theme.RoleText), false})
		default:
			out = append(out, boardNeed{o.def.Name, dim, false})
		}
	}
	if i, _ := m.state.Ruleset().Index(t.def.Age); i > m.age {
		out = append(out, boardNeed{"the " + m.ageName(t.def.Age), dim, false})
	}
	return out
}

// paintTop is the tech in work with its clock, and the queue as one timeline.
func (p *researchPanel) paintTop(s rankPaint, w int, head func(x, y int, str string, c kit.Color) int) {
	m := p.model
	rs := m.state.Research
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	s.text(3, 2, "NOW RESEARCHING", dim, false)
	cur := m.byKey[rs.CurrentTech]
	if cur == nil {
		s.text(3, 3, "nothing is in work", txt, false)
		s.text(3, 4, "Enter on a tech starts it", dim, false)
		return
	}
	lc := theme.Hue(m.laneOf[cur.def.Lane].Hue, bg)
	fx := head(3, 3, clipRunes(cur.def.Name, 20), lc) + 2
	x := s.text(fx, 3, strings.ToUpper(m.laneOf[cur.def.Lane].Name), lc, true)
	s.text(x+3, 3, m.ageName(cur.def.Age), dim, false)
	s.text(fx, 4, fitText(strings.Join(m.effects(cur), "   "), w-fx-30), kit.Role(theme.RoleEffect), false)

	left, total := int(float64(rs.TicksLeft)*m.tick+0.5), int(float64(rs.TotalTicks)*m.tick+0.5)
	frac := 0.0
	if rs.Price > 0 {
		frac = rs.Progress / rs.Price
	}
	s.right(w-3, 2, "TIME LEFT", dim, false)
	if rs.NeedsScholars {
		s.right(w-3, 3, "no scholar is researching", kit.Role(theme.RoleWarning), false)
	} else {
		str := clock(left)
		head(w-3-headWidth(str), 3, str, bright)
	}

	// The queue as one timeline: each tech as wide as the time it takes.
	items := []*boardTech{cur}
	for _, k := range rs.Queue {
		if q := m.byKey[k]; q != nil {
			items = append(items, q)
		}
	}
	sum := 0
	for _, t := range items {
		sum += max(t.secs, 1)
	}
	tw := w - 6 - (len(items) - 1)
	x, wait := 3, left
	for i, t := range items {
		c := theme.Hue(m.laneOf[t.def.Lane].Hue, bg)
		bw := int(float64(tw)*float64(max(t.secs, 1))/float64(sum) + 0.5)
		if i == len(items)-1 {
			bw = w - 3 - x
		}
		if i == 0 {
			s.bar(x, 6, bw, frac, c, theme.Mix(bg, c, 0.28), '▄')
			lx := s.text(x, 7, fmt.Sprintf("%d%%", int(frac*100+0.5)), bright, true)
			s.text(lx+2, 7, fmt.Sprintf("%s of %s done", clock(total-left), clock(total)), dim, false)
		} else {
			s.bar(x, 6, bw, 0, c, theme.Mix(bg, c, 0.28), '▄')
			lab := "NEXT"
			if i > 1 {
				lab = "THEN"
			}
			lx := s.text(x, 7, lab, dim, false)
			lx = s.text(lx+2, 7, fitText(t.def.Name, max(bw-lx-2, 1)), c, true)
			lx = s.text(lx+2, 7, hm(t.secs), txt, false)
			s.text(lx+2, 7, "starts in "+hm(wait), dim, false)
			wait += t.secs
		}
		x += bw + 1
	}
	s.right(w-3, 7, "the queue ends in "+hm(wait), dim, false)
}

// paintDetail is the picked tech's column: its numbers, what it gives, its
// place in the tree, the two buttons and the ten lanes.
func (p *researchPanel) paintDetail(s rankPaint, g boardGeo, t *boardTech, sc kit.Color,
	head func(x, y int, str string, c kit.Color) int) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	faint, eff := faintCol(), kit.Role(theme.RoleEffect)
	rx, rw, y := g.rx, g.rw, g.yHead
	nx := head(rx, y, clipRunes(t.def.Name, (rw-14)/2), sc) + 2
	s.text(nx, y, strings.ToUpper(m.laneOf[t.def.Lane].Name), sc, true)
	s.text(nx, y+1, m.ageName(t.def.Age), dim, false)
	s.rule(rx, y+2, rw, theme.Mix(bg, sc, 0.35))
	y += 4
	rs := m.state.Research
	state := "after the queue"
	switch {
	case t.ts.Researched:
		state = "researched"
	case t.def.Key == rs.CurrentTech:
		state = "in work"
	case m.queued[t.def.Key]:
		state = "queued"
	case rs.CurrentTech == "":
		state = "now"
	}
	for _, kv := range [][3]string{
		{"COST", config.FormatAmount(t.def.Cost), "of work for the scholars"},
		{"TIME", timeOf(t.secs), fmt.Sprintf("with %d scholars", m.state.Scholars.Researching)},
		{"STARTS", state, ""},
	} {
		kx := s.text(rx, y, kv[0], dim, false)
		vx := rx + 22 - runeLen(kv[1])
		for x := kx + 1; x < vx-1; x += 2 {
			s.text(x, y, "·", faint, false)
		}
		s.text(vx, y, kv[1], bright, true)
		s.text(rx+24, y, kv[2], txt, false)
		y++
	}
	y++
	s.text(rx, y, "UNLOCKS", dim, false)
	y++
	for _, e := range m.effects(t) {
		for _, ln := range wrapText(e, rw) {
			s.text(rx, y, ln, eff, false)
			y++
		}
	}
	y++
	for _, ln := range wrapText(t.def.Description, rw) {
		s.text(rx, y, ln, dim, false)
		y++
	}

	// Where it sits: what it needs above, what it leads to below.
	y++
	s.text(rx, y, "IN THE TREE", dim, false)
	y += 2
	spine := theme.Mix(bg, sc, 0.55)
	treeEnd := g.by - 3
	mark := func(x, y int, r rune) { s.p.Put(s.o.X+x, s.o.Y+y, r, kit.Style{}.Foreground(spine)) }
	pres := t.needs()
	for i, k := range pres {
		o := m.byKey[k]
		if o == nil {
			continue
		}
		mark(rx, y, map[bool]rune{true: '┌', false: '├'}[i == 0])
		mark(rx+1, y, '─')
		if o.ts.Researched {
			s.text(rx+3, y, "✓", kit.Role(theme.RolePositive), false)
		} else {
			s.text(rx+3, y, "□", dim, false)
		}
		s.text(rx+5, y, o.def.Name, txt, false)
		s.right(rx+rw, y, m.laneShort(o.def.Lane), theme.Hue(m.laneOf[o.def.Lane].Hue, bg), false)
		y++
	}
	s.text(rx, y, "♦", bright, false)
	s.text(rx+3, y, t.def.Name, sc, true)
	s.right(rx+rw, y, "you are here", dim, false)
	y++
	leads := m.leads[t.def.Key]
	for i, k := range leads[:min(len(leads), 6)] {
		if y > treeEnd {
			break
		}
		o := m.byKey[k]
		last := i == min(len(leads), 6)-1
		mark(rx, y, map[bool]rune{true: '└', false: '├'}[last])
		mark(rx+1, y, '─')
		mark(rx+2, y, '►')
		s.text(rx+5, y, fitText(o.def.Name, 20), txt, false)
		s.text(rx+27, y, m.ageName(o.def.Age), dim, false)
		s.right(rx+rw, y, m.laneShort(o.def.Lane), theme.Hue(m.laneOf[o.def.Lane].Hue, bg), false)
		y++
		if also := m.missing(o, t.def.Key); len(also) > 0 && y <= treeEnd {
			if !last {
				mark(rx, y, '│')
			}
			s.text(rx+5, y, fitText("also needs "+strings.Join(also, ", "), rw-5), dim, false)
			y++
		}
	}

	// The buttons keep their place above the lanes.
	b1, b2 := "  Research  Enter  ", "  Queue  Q  "
	by := g.by
	s.fill(rx, by-1, runeLen(b1), 3, theme.Of(theme.RoleAccent))
	s.text(rx, by, b1, kit.Role(theme.RoleOnAccent), true)
	b2x := rx + runeLen(b1) + 2
	s.fill(b2x, by-1, runeLen(b2), 3, theme.Of(theme.RoleChip))
	s.text(b2x, by, b2, txt, false)
	p.hits = append(p.hits,
		boardHit{image.Rect(rx, by-1, rx+runeLen(b1), by+2), hitResearch, ""},
		boardHit{image.Rect(b2x, by-1, b2x+runeLen(b2), by+2), hitQueue, ""})
	if p.view.note != "" {
		s.text(rx, by+3, fitText(p.view.note, rw), kit.Role(theme.RoleWarning), false)
	}
	if g.lanes {
		s.rule(rx, g.strips-2, rw, theme.Mix(bg, sc, 0.35))
		p.paintLanes(s, g, t)
	}
}

// paintLanes is the ten lanes, two rows each: a row of marks, one a tech, and
// under it the next tech of the lane that is not yet in hand.
func (p *researchPanel) paintLanes(s rankPaint, g boardGeo, sel *boardTech) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	rx, rw, y := g.rx, g.rw, g.strips
	have := 0
	for _, t := range m.techs {
		if t.ts.Researched {
			have++
		}
	}
	s.text(rx, y, "THE TEN LANES", dim, false)
	s.right(rx+rw, y, fmt.Sprintf("%d of %d", have, len(m.techs)), dim, false)
	y += 2
	cur := m.state.Research.CurrentTech
	for _, l := range m.lanes {
		var lane []*boardTech
		for _, t := range m.techs {
			if t.def.Lane == l.Key {
				lane = append(lane, t)
			}
		}
		if len(lane) == 0 {
			continue
		}
		set := m.state.Ruleset()
		sort.SliceStable(lane, func(i, j int) bool {
			ai, _ := set.Index(lane[i].def.Age)
			aj, _ := set.Index(lane[j].def.Age)
			if ai != aj {
				return ai < aj
			}
			return lane[i].def.Cost < lane[j].def.Cost
		})
		lc := theme.Hue(l.Hue, bg)
		n, next := 0, -1
		for i, t := range lane {
			r, col := '·', theme.Mix(bg, lc, 0.6)
			switch {
			case t.def.Key == sel.def.Key:
				r, col = '♦', bright
			case t.def.Key == cur:
				r, col = '►', bright
			case t.ts.Researched:
				r, col = '■', lc
			case m.queued[t.def.Key]:
				r, col = '◊', lc
			}
			if t.ts.Researched {
				n++
			}
			if x := 13 + i; x < rw-9 {
				s.p.Put(s.o.X+rx+x, s.o.Y+y, r, kit.Style{}.Foreground(col))
			}
			if t.ts.Researched || m.queued[t.def.Key] || t.def.Key == cur {
				continue
			}
			if next < 0 || t.ts.Available && !lane[next].ts.Available ||
				t.ts.Available == lane[next].ts.Available && t.def.Cost < lane[next].def.Cost {
				next = i
			}
		}
		nameCol := lc
		if n == 0 {
			nameCol = theme.Mix(bg, lc, 0.55)
		}
		p.hits = append(p.hits, boardHit{image.Rect(rx, y, rx+rw, y+2), hitLane, l.Key})
		s.text(rx, y, m.laneShort(l.Key), nameCol, p.view.lane == l.Key)
		s.right(rx+rw, y, fmt.Sprintf("%d of %d", n, len(lane)), dim, false)
		y++
		if next >= 0 {
			t := lane[next]
			c := theme.Mix(bg, nameCol, 0.7)
			nx := s.text(rx+13, y, fitText(t.def.Name, 22), c, false)
			nx = s.text(nx+2, y, timeOf(t.secs), dim, false) + 2
			if t.ts.Available {
				s.text(nx, y, "ready", kit.Role(theme.RolePositive), false)
			} else {
				miss := m.missing(t, "")
				txtNeed := "needs " + miss[0]
				if len(miss) > 1 {
					txtNeed += fmt.Sprintf(" +%d", len(miss)-1)
				}
				s.text(nx, y, fitText(txtNeed, rx+rw-nx), theme.Mix(bg, dim, 0.75), false)
			}
		}
		y++
	}
	y++
	lx := rx
	for _, k := range []struct {
		r rune
		s string
		c kit.Color
	}{{'■', "done", txt}, {'►', "in work", bright}, {'◊', "queued", txt}, {'♦', "selected", bright}, {'·', "open", txt}} {
		s.text(lx, y, string(k.r), k.c, false)
		lx = s.text(lx+2, y, k.s, dim, false) + 3
	}
}
