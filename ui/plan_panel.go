package ui

import (
	"image"
	"reflect"
	"strconv"
	"unicode"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/planboard"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// plan_panel.go is the Plan overlay: the drafting table. The queue is the
// list on the left with its bars on a time ruler; the selected item has a job
// sheet, the advance a gate card, and a tray adds items. The pictures are
// pkg/planboard's, the words and the keys are here (plan_panel_paint.go draws
// them). The keyboard stays with the command bar: the panel takes its letters
// while the bar is empty, or all keys while its name or number box is open.

// What the tray's tabs hold, in order.
var planTabs = [5]string{game.PlanBuild, game.PlanUpgrade, game.PlanResearch, game.PlanAdvance, game.PlanExpedition}

type planMode uint8

const (
	planQueue planMode = iota
	planNameBox
	planCountBox
	planPreview
)

// A planEntry is a line of the tray: something the plan can take.
type planEntry struct {
	kind, key, label, price string
	n                       int
}

// memo keeps the last picture made for a key, so an unchanged panel hands
// the window the same picture and nothing is uploaded again.
type memo[K, V any] struct {
	k   K
	v   V
	set bool
}

func (m *memo[K, V]) get(k K, make func() V) V {
	if !m.set || !reflect.DeepEqual(m.k, k) {
		m.k, m.v, m.set = k, make(), true
	}
	return m.v
}

type planPanel struct {
	barPanel
	kit.Base

	engine *game.GameEngine

	st    game.GameState
	est   game.PlanEstimate
	opt   game.PlanOptimizeResult // the preview, while planPreview
	old   []game.PlanItemEstimate // the estimate before it
	shown []planEntry             // the tray's list: this tab, filtered by the name typed
	all   []planEntry

	sel, first, tab, trayOff int
	mode                     planMode
	input, note              string
	good, armed              bool
	drag                     struct {
		down, on    bool
		from, to, y int
	}

	l     planboard.Layout
	ok    bool
	keys  *xgrid
	words []planWord
	pics  []planPic
	drawn planboard.Drawn

	boardM memo[planboard.Layout, *image.RGBA]
	schedM memo[planSchedKey, planboard.Drawn]
	picM   [4]memo[any, *image.RGBA]
}

func newPlanPanel() *planPanel {
	p := &planPanel{}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *planPanel) open(state game.GameState) Primitive {
	p.sel, p.first, p.tab, p.trayOff = 0, 0, 0, 0
	p.mode, p.input, p.note, p.armed = planQueue, "", "", false
	p.update(state)
	return overBar(p)
}

// update takes the engine's state and lays the panel out again. It runs on
// the UI goroutine, outside the engine's lock.
func (p *planPanel) update(state game.GameState) {
	p.st = state
	if p.engine != nil {
		p.est = p.engine.PlanEstimate()
	}
	p.entries()
	p.relayout()
}

func (p *planPanel) n() int {
	if p.mode == planPreview {
		return len(p.opt.Rows)
	}
	return len(p.st.Plan)
}

// view is the selected plan item; false when there is none.
func (p *planPanel) view() (game.PlanItemView, bool) {
	if p.mode == planPreview || p.sel < 0 || p.sel >= len(p.st.Plan) {
		return game.PlanItemView{}, false
	}
	return p.st.Plan[p.sel], true
}

// closeFirst takes back what Esc closes before the panel: the preview, a box,
// the Clear question.
func (p *planPanel) closeFirst() bool {
	had := p.mode != planQueue || p.armed
	p.mode, p.input, p.armed = planQueue, "", false
	if had {
		p.relayout()
	}
	return had
}

// run presses a command line for the panel: the line goes through the
// command bar, whose reply is the feedback, and the panel takes the state again.
func (p *planPanel) run(line string) {
	p.note, p.armed = "", false
	if p.buttons.run == nil {
		return
	}
	p.buttons.run(line)
	if p.engine != nil {
		p.update(p.engine.GetState())
	}
}

// reply puts what a command run over the panel did in its note.
func (p *planPanel) reply(_ string, res CommandResult) {
	msg, bad := replyLine(res)
	p.note, p.good = msg, !bad
}

// at is the 1-based number the plan commands give the selected item.
func (p *planPanel) at() string { return strconv.Itoa(p.sel + 1) }

func (p *planPanel) selectTo(i int) {
	if n := p.n(); n > 0 {
		p.sel = clamp(i, 0, n-1)
		p.relayout()
	}
}

// move shifts the selected item by d places; it stays selected.
func (p *planPanel) move(d int) {
	to := clamp(p.sel+d, 0, max(0, len(p.st.Plan)-1))
	if to == p.sel {
		return
	}
	p.run("plan move " + p.at() + " " + strconv.Itoa(to+1))
	p.sel = to
}

func (p *planPanel) moveTo(top bool) {
	d := len(p.st.Plan) - 1 - p.sel
	if top {
		d = -p.sel
	}
	p.move(d)
}

func (p *planPanel) remove() { p.run("plan remove " + p.at()) }

func (p *planPanel) setCount(count int) {
	p.run("plan count " + p.at() + " " + strconv.Itoa(max(1, count)))
}

func (p *planPanel) pause() {
	if p.st.PlanPaused {
		p.run("plan resume")
		return
	}
	p.run("plan pause")
}

// clear asks first: the second press empties the plan.
func (p *planPanel) clear() {
	if !p.armed && len(p.st.Plan) > 1 {
		p.armed = true
		p.relayout()
		return
	}
	p.run("plan clear")
	p.sel = 0
}

// optimize shows the better order; Enter keeps it, Esc leaves the plan as it is.
func (p *planPanel) optimize() {
	res, err := p.engine.PlanOptimizePreview()
	if err != nil || !res.Changed {
		p.note, p.good = planNoChange(res.Reason), err == nil
		if err != nil {
			p.note = err.Error()
		}
		p.update(p.st)
		return
	}
	p.old, p.opt, p.mode, p.note = p.est.Items, res, planPreview, ""
	p.relayout()
}

func (p *planPanel) keep() {
	p.mode = planQueue
	p.run("plan optimize")
}

// setTab shows tray tab t.
func (p *planPanel) setTab(t int) {
	p.tab, p.trayOff = (t+len(planTabs))%len(planTabs), 0
	p.entries()
	p.relayout()
}

// planLine is the command that plans entry e.
func planLine(e planEntry) string {
	switch e.kind {
	case game.PlanBuild:
		return "plan build " + e.key + " 1"
	case game.PlanUpgrade:
		return "plan upgrade " + e.key + " 1"
	case game.PlanResearch:
		return "plan research " + e.key
	case game.PlanAdvance:
		return "plan advance"
	case game.PlanExpedition:
		return "plan expedition " + e.key
	}
	return ""
}

// add puts entry e in the plan, just below the selection.
func (p *planPanel) add(e planEntry) {
	line := planLine(e)
	if line == "" || p.engine == nil {
		return
	}
	before := len(p.st.Plan)
	p.run(line)
	// New items land at the end; take them up to just below the selection.
	after := len(p.engine.GetState().Plan)
	at := p.sel + 2
	if before == 0 {
		at = 1
	}
	for i := 0; i < after-before; i++ {
		from := before + 1 + i
		if at+i != from {
			p.run("plan move " + strconv.Itoa(from) + " " + strconv.Itoa(at+i))
		}
	}
	p.sel = max(0, min(at-1, after-1))
	if after == before {
		p.sel = max(0, after-1) // merged into the last item
	}
	p.mode, p.input = planQueue, ""
	p.relayout()
}

// routeKey takes a key if it is one of the panel's.
func (p *planPanel) routeKey(ev *kit.Key, prompt string) bool {
	if p.mode == planNameBox || p.mode == planCountBox {
		return p.boxKey(ev)
	}
	if prompt != "" {
		return false
	}
	if p.mode == planPreview {
		switch ev.Code {
		case kit.KeyEnter:
			p.keep()
		case kit.KeyUp, kit.KeyDown:
			p.selectTo(p.sel + int(ev.Code-kit.KeyUp)*2 - 1)
		default:
			return false
		}
		return true
	}
	page := max(1, len(p.drawn.Slots)-2)
	switch ev.Code {
	case kit.KeyUp:
		p.selectTo(p.sel - 1)
	case kit.KeyDown:
		p.selectTo(p.sel + 1)
	case kit.KeyPgUp:
		p.selectTo(p.sel - page)
	case kit.KeyPgDn:
		p.selectTo(p.sel + page)
	case kit.KeyHome:
		p.selectTo(0)
	case kit.KeyEnd:
		p.selectTo(p.n() - 1)
	case kit.KeyTab:
		p.setTab(p.tab + 1)
	case kit.KeyBacktab:
		p.setTab(p.tab - 1)
	case kit.KeyDelete:
		p.remove()
	case kit.KeyEnter:
		return false
	case kit.KeyRune:
		return p.letter(unicode.ToLower(ev.Rune))
	default:
		return false
	}
	return true
}

// letter is a key that prints: the panel's own, or a digit that opens the number box.
func (p *planPanel) letter(r rune) bool {
	_, has := p.view()
	if !has && r != 'a' && r != 'p' && r != 'o' && r != 'c' {
		return false
	}
	switch {
	case r == 'p':
		p.pause()
	case r == 'o':
		p.optimize()
	case r == 'c':
		p.clear()
	case r == 'a':
		p.mode, p.input = planNameBox, ""
		p.relayout()
	case r == 't', r == 'b':
		p.moveTo(r == 't')
	case r == 'u':
		p.move(-1)
	case r == 'd':
		p.move(1)
	case r == 'x':
		p.remove()
	case r == '+' || r == '=':
		p.setCount(p.st.Plan[p.sel].Count + 1)
	case r == '-':
		p.setCount(p.st.Plan[p.sel].Count - 1)
	case r == 'n' || r >= '0' && r <= '9':
		p.mode, p.input = planCountBox, ""
		if r != 'n' {
			p.input = string(r)
		}
		p.relayout()
	default:
		return false
	}
	return true
}

// boxKey is a key while the name box or the number box is open: it is the box's.
func (p *planPanel) boxKey(ev *kit.Key) bool {
	switch ev.Code {
	case kit.KeyEsc:
		p.closeFirst()
	case kit.KeyBackspace:
		if r := []rune(p.input); len(r) > 0 {
			p.input = string(r[:len(r)-1])
		}
	case kit.KeyEnter:
		p.commitBox()
		return true
	case kit.KeyTab:
		p.setTab(p.tab + 1)
	case kit.KeyRune:
		if p.mode == planNameBox || ev.Rune >= '0' && ev.Rune <= '9' && len(p.input) < 4 {
			p.input += string(ev.Rune)
		}
	default:
		return false
	}
	p.relayout()
	return true
}

func (p *planPanel) commitBox() {
	if p.mode == planCountBox {
		n := 0
		for _, r := range p.input {
			n = n*10 + int(r-'0')
		}
		p.mode, p.input = planQueue, ""
		if n > 0 {
			p.setCount(n)
		}
		return
	}
	if len(p.shown) == 0 {
		p.note, p.good = rowid.Sprintf("ui.plan_panel.commitBox.01", "Nothing called %q fits the plan.", p.input), false
		p.relayout()
		return
	}
	p.add(p.shown[0])
}

// ---- mouse ----

// rowAt is the item whose rows hold pt (panel cells), -1 when none, n past the last.
func (p *planPanel) rowAt(pt image.Point) int {
	if pt.X < p.l.ListX || pt.X > p.l.Lanes.Max.X || pt.Y < p.l.ListY || pt.Y >= p.l.SchedEnd {
		return -1
	}
	at := -1
	for _, s := range p.drawn.Slots {
		if s.Item >= 0 && pt.Y >= s.Y && pt.Y < s.Y+p.l.RowH {
			return s.Item
		}
		if s.Item >= 0 && pt.Y >= s.Y {
			at = s.Item
		}
	}
	if at >= 0 && p.drawn.More == 0 {
		return p.n()
	}
	return -1
}

// Mouse: the buttons, then a click or a drag on the list, the tray's tabs and entries.
func (p *planPanel) Mouse(m kit.Mouse) bool {
	org := p.Bounds().Min
	if !p.ok || p.keys == nil || p.buttons.mouse(p.keys, org, m) {
		return p.ok
	}
	pt := m.Pt().Sub(org)
	switch m.Kind {
	case kit.MouseWheel:
		if pt.In(p.l.Tray) {
			p.trayOff = max(0, p.trayOff+m.Wheel*2)
		} else {
			p.first = clamp(p.first+m.Wheel, 0, max(0, p.n()-1))
		}
		p.relayout()
		return true
	case kit.MouseMove:
		if p.drag.down {
			i := p.rowAt(pt)
			if i >= 0 && i != p.drag.to || p.drag.on && pt.Y != p.drag.y {
				p.drag.on, p.drag.y = true, pt.Y
				if i >= 0 {
					p.drag.to = i
				}
				p.relayout()
			}
			return true
		}
		return false
	case kit.MouseRelease:
		return p.drop()
	case kit.MousePress:
		if m.Button != kit.ButtonLeft {
			return false
		}
	default:
		return false
	}
	if i := p.rowAt(pt); i >= 0 && i < p.n() {
		p.drag.down, p.drag.on, p.drag.from, p.drag.to = p.mode == planQueue, false, i, i
		p.selectTo(i)
		return true
	}
	for t := range planTabs {
		if pt.In(p.l.TabRect(t)) {
			p.setTab(t)
			return true
		}
	}
	for k := 0; ; k++ {
		r, ok := p.l.EntryRect(k)
		if !ok {
			return false
		}
		if pt.In(r) && p.trayOff+k < len(p.shown) {
			p.add(p.shown[p.trayOff+k])
			return true
		}
	}
}

// drop ends a drag: the item goes above the row it was let go over.
func (p *planPanel) drop() bool {
	d := p.drag
	p.drag.down, p.drag.on = false, false
	if !d.on {
		return false
	}
	to := d.to
	if to > d.from {
		to--
	}
	if to != d.from {
		p.sel = d.from
		p.move(to - d.from)
		return true
	}
	p.relayout()
	return true
}

// PointerAt: the hand over a row, a tab, an entry or a key.
func (p *planPanel) PointerAt(at image.Point) kit.Pointer {
	if !p.ok || p.keys == nil {
		return kit.PointerArrow
	}
	org := p.Bounds().Min
	pt := at.Sub(org)
	if p.buttons.pointer(p.keys, org, at) == kit.PointerHand {
		return kit.PointerHand
	}
	hand := p.mode == planQueue && p.rowAt(pt) >= 0 && p.rowAt(pt) < p.n()
	for t := range planTabs {
		hand = hand || pt.In(p.l.TabRect(t))
	}
	for k := 0; k < p.l.EntryCount() && !hand; k++ {
		r, _ := p.l.EntryRect(k)
		hand = pt.In(r) && p.trayOff+k < len(p.shown)
	}
	return handIf(hand)
}
