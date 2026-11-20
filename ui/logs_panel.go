package ui

import (
	"fmt"
	"image"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/logroll"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/window/kit"
)

// logs_panel.go is the Logs panel, "the roll and the slate": the log as a roll
// of paper (rows, filters, Find, following, folds), the rack of office stamps
// and the Detail lever that filter it, and the slate beside it
// (logs_slate.go). logs_paint.go draws it from pkg/logroll's pictures.

var (
	txtLogsTitle = rowid.E("ui.dashboard.NewDashboard.12", "Logs")
	txtLogsEmpty = rowid.E("ui.logs_panel.words.03", "Nothing is written here yet.")
	txtLogsNone  = rowid.E("ui.logs_panel.words.04", "No line says that.")
	txtLogsFind  = rowid.E("ui.logs_panel.words.05", "Find")
	txtLogsSmall = rowid.E("ui.logs_panel.words.07", "Make the window larger to read the log.")
	txtLogsBands = [...]rowid.Entry{rowid.E("ui.logs_panel.bands.01", "Last 10 minutes"), rowid.E("ui.logs_panel.bands.02", "Last hour"),
		rowid.E("ui.logs_panel.bands.03", "Earlier today"), rowid.E("ui.logs_panel.bands.04", "Yesterday")}
	txtLogsLevels = [...]rowid.Entry{rowid.E("ui.logs_panel.levels.01", "Headlines"), rowid.E("ui.logs_panel.levels.02", "Town"), rowid.E("ui.logs_panel.levels.03", "Everything")}
)

// logsKeys letter the ten stamps, in office order; the other letters are F, O,
// N and A.
const logsKeys = "BRPKSTDMWG"

type logRowKind uint8

const (
	rowLine logRowKind = iota
	rowBand
	rowSeam
	rowMember
	rowEarlier
)

// logRow is one row of the roll.
type logRow struct {
	kind         logRowKind
	e            game.LogEntry
	text, tally  string
	at           int64
	paid, gained []game.LogAmt
	key          string
	hits         [][2]int // runes of text Find marks
}

// logsRequest is how a `logs` command opens the panel.
type logsRequest struct {
	level   int
	office  game.Office
	all     bool
	find    string
	setFind bool
}

type logsPanel struct {
	barPanel
	kit.Base

	state game.GameState
	now   func() time.Time
	focus func() // takes the keyboard for Find

	level   int
	on      [11]bool
	find    string
	typing  bool
	rows    []logRow
	opened  map[string]bool
	top     int
	sel     int
	follow  bool
	seen    int
	req     logsRequest
	lay     logroll.Layout
	chrome  *xgrid
	cache   map[string]*image.RGBA
	hasData bool
}

func newLogsPanel() *logsPanel {
	p := &logsPanel{now: time.Now, cache: map[string]*image.RGBA{}}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// request is what the next open starts from.
func (p *logsPanel) request(r logsRequest) { p.req = r }

func (p *logsPanel) open(state game.GameState) Primitive {
	p.level, p.find, p.typing, p.follow, p.sel = game.LevelTown, "", false, true, 0
	p.opened = map[string]bool{}
	for o := range p.on {
		p.on[o] = o != int(game.OfficeGame)
	}
	if r := p.req; r.level > 0 || r.office > 0 || r.all || r.setFind {
		if r.level > 0 {
			p.level = r.level
		}
		if r.all {
			p.on = [11]bool{}
			for o := 1; o <= 10; o++ {
				p.on[o] = true
			}
		}
		if r.office > 0 {
			p.on = [11]bool{}
			p.on[r.office] = true
		}
		p.find = r.find
	}
	p.req = logsRequest{}
	p.update(state)
	return overBar(p)
}

func (p *logsPanel) update(state game.GameState) {
	p.state, p.hasData = state, true
	if p.follow {
		p.seen = len(state.Log)
	}
	p.build()
}

// words splits s into lowercase words of letters and digits.
func logWords(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

func (p *logsPanel) resName(k string) string {
	if k == "workers" {
		return k
	}
	return p.state.Ruleset().Name(rules.KindResource, k)
}

// amounts writes amounts as "12 wood, 3 stone" with sign in front of each.
func (p *logsPanel) amounts(a []game.LogAmt, sign string) string {
	parts := make([]string, len(a))
	for i, x := range a {
		parts[i] = sign + textfmt.Number(x.N) + " " + p.resName(x.Res)
	}
	return strings.Join(parts, ", ")
}

// matches tells whether every word of Find is a whole word of the line: its
// text, its kept members' text or what it paid and gained.
func (p *logsPanel) matches(e game.LogEntry) bool {
	hay := e.Message + " " + p.amounts(e.Paid, "") + " " + p.amounts(e.Gained, "")
	for _, m := range e.Members {
		hay += " " + m.Text
	}
	have := logWords(hay)
	for _, w := range logWords(p.find) {
		if !slices.Contains(have, w) {
			return false
		}
	}
	return true
}

// marks are the runes of text that are whole words of Find.
func (p *logsPanel) marks(text string) (out [][2]int) {
	want, rs := logWords(p.find), []rune(text)
	for i := 0; i < len(rs); {
		if !unicode.IsLetter(rs[i]) && !unicode.IsDigit(rs[i]) {
			i++
			continue
		}
		j := i
		for j < len(rs) && (unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j])) {
			j++
		}
		if slices.Contains(want, strings.ToLower(string(rs[i:j]))) {
			out = append(out, [2]int{i, j})
		}
		i = j
	}
	return out
}

// when says a time of day with its day: "today 14:05", "yesterday 21:30",
// "Mon 08:00" or a date.
func (p *logsPanel) when(at int64) string {
	t, now := time.Unix(at, 0), p.now()
	day := func(d time.Time) string { return d.Format("2006-01-02") }
	switch {
	case day(t) == day(now):
		return "today " + t.Format("15:04")
	case day(t) == day(now.AddDate(0, 0, -1)):
		return "yesterday " + t.Format("15:04")
	case now.Sub(t) < 6*24*time.Hour:
		return t.Format("Mon 15:04")
	}
	return t.Format("2 Jan 15:04")
}

// band is the heading of the stretch of the roll a line falls in.
func (p *logsPanel) band(at int64) string {
	t, now := time.Unix(at, 0), p.now()
	switch age := now.Sub(t); {
	case at == 0 || age < 10*time.Minute:
		return txtLogsBands[0].String()
	case age < time.Hour:
		return txtLogsBands[1].String()
	case t.YearDay() == now.YearDay() && t.Year() == now.Year():
		return txtLogsBands[2].String()
	case t.YearDay() == now.AddDate(0, 0, -1).YearDay() && t.Year() == now.AddDate(0, 0, -1).Year():
		return txtLogsBands[3].String()
	}
	return t.Format("Monday 2 January")
}

// build makes the roll's rows from the log and the filters.
// build makes the rows. A stamp hides its office but never a headline.
func (p *logsPanel) build() {
	rows, band, bi := p.rows[:0], "", 0
	var counts [11]int
	for _, e := range p.state.Log {
		if e.Seam {
			if p.find == "" && len(rows) > 0 {
				rows = append(rows, logRow{kind: rowSeam, e: e, text: rowid.Sprintf("ui.logs_panel.words.01", "Sitting ended %s, %d lines", p.when(e.At), e.Count)})
				band = ""
			}
			continue
		}
		if e.Level > p.level || e.Office < 1 || e.Office > 10 || !p.on[e.Office] && e.Level != game.LevelHeadline || p.find != "" && !p.matches(e) {
			continue
		}
		if b := p.band(e.At); b != band {
			band, counts = b, [11]int{}
			rows, bi = append(rows, logRow{kind: rowBand, text: b}), len(rows)
		}
		counts[e.Office]++
		rows[bi].tally = p.tallyText(counts)
		rows = append(rows, logRow{kind: rowLine, e: e, text: e.Message, at: e.At, paid: e.Paid, gained: e.Gained, key: fmt.Sprint(e.At, e.Tick, e.ID, e.Office)})
		if e.Count > 1 && p.opened[rows[len(rows)-1].key] {
			if n := e.Earlier(); n > 0 {
				rows = append(rows, logRow{kind: rowEarlier, e: e, text: rowid.Sprintf("ui.logs_panel.words.06", "%d earlier lines", n)})
			}
			for _, m := range e.Members {
				rows = append(rows, logRow{kind: rowMember, e: e, text: m.Text, at: m.At, paid: m.Paid, gained: m.Gained})
			}
		}
	}
	for i := range rows {
		if rows[i].kind != rowBand && rows[i].kind != rowSeam {
			rows[i].hits = p.marks(rows[i].text)
		}
	}
	p.rows = rows
	p.clamp()
}

// tallyText names the three offices with most lines in a band.
func (p *logsPanel) tallyText(counts [11]int) string {
	var offs []int
	for o := 1; o <= 10; o++ {
		if counts[o] > 0 {
			offs = append(offs, o)
		}
	}
	slices.SortStableFunc(offs, func(a, b int) int { return counts[b] - counts[a] })
	parts := make([]string, 0, 3)
	for _, o := range offs[:min(3, len(offs))] {
		parts = append(parts, fmt.Sprintf("%s %d", game.Office(o).Name(), counts[o]))
	}
	return strings.Join(parts, " · ")
}

// height is how many rows of the roll show at once.
func (p *logsPanel) height() int { return max(1, p.lay.Body.H) }

// clamp keeps the cursor and the window inside the rows and pins the window to
// the newest row while following.
func (p *logsPanel) clamp() {
	n, h := len(p.rows), p.height()
	p.sel = clamp(p.sel, 0, max(0, n-1))
	if p.follow {
		p.sel = max(0, n-1)
	}
	p.top = clamp(p.top, 0, max(0, n-h))
	if p.follow {
		p.top = max(0, n-h)
	}
}

// moveTo puts the cursor on row i, scrolling to keep it in sight; the last row
// takes the window back to following.
func (p *logsPanel) moveTo(i int) {
	n, h := len(p.rows), p.height()
	p.sel = clamp(i, 0, max(0, n-1))
	p.follow = p.sel >= n-1
	if p.sel < p.top {
		p.top = p.sel
	} else if p.sel >= p.top+h {
		p.top = p.sel - h + 1
	}
	p.clamp()
}

func (p *logsPanel) scrollBy(d int) {
	p.top = clamp(p.top+d, 0, max(0, len(p.rows)-p.height()))
	p.follow = p.top >= len(p.rows)-p.height()
	p.sel = clamp(p.sel, p.top, min(len(p.rows)-1, p.top+p.height()-1))
	p.clamp()
}

func (p *logsPanel) toggleFold() {
	if p.sel < 0 || p.sel >= len(p.rows) || p.rows[p.sel].kind != rowLine || p.rows[p.sel].e.Count < 2 {
		return
	}
	k := p.rows[p.sel].key
	p.opened[k] = !p.opened[k]
	p.build()
}

func (p *logsPanel) setLevel(l int) {
	p.level = clamp(l, game.LevelHeadline, game.LevelEverything)
	p.build()
}

func (p *logsPanel) toggleOffice(o int) {
	if o >= 1 && o <= 10 {
		p.on[o] = !p.on[o]
		p.build()
	}
}

// startFind puts the keyboard on the Find slip.
func (p *logsPanel) startFind() {
	p.typing = true
	if p.focus != nil {
		p.focus()
	}
}

// routeKey takes the keys that print nothing, the digits and Space while the
// command bar is empty.
func (p *logsPanel) routeKey(ev *kit.Key, prompt string) bool {
	if prompt != "" {
		return false
	}
	h := p.height()
	switch ev.Code {
	case kit.KeyUp:
		p.moveTo(p.sel - 1)
	case kit.KeyDown:
		p.moveTo(p.sel + 1)
	case kit.KeyPgUp:
		p.moveTo(p.sel - h + 1)
	case kit.KeyPgDn:
		p.moveTo(p.sel + h - 1)
	case kit.KeyHome:
		p.moveTo(0)
	case kit.KeyEnd:
		p.moveTo(len(p.rows) - 1)
	case kit.KeyEnter:
		p.toggleFold()
	case kit.KeyRune:
		switch r := ev.Rune; {
		case r >= '1' && r <= '3':
			p.setLevel(int(r - '0'))
		case r == ' ':
			p.toggleFold()
		case r == 'j':
			p.moveTo(p.sel + 1)
		case r == 'k':
			p.moveTo(p.sel - 1)
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// Key is for when the panel has the keyboard: besides routeKey's keys, the
// letters. They are the panel's only then, so the command bar can be typed in.
func (p *logsPanel) Key(k kit.Key) bool {
	if p.typing {
		switch {
		case k.Code == kit.KeyEnter, k.Code == kit.KeyEsc:
			p.typing = false
		case k.Code == kit.KeyBackspace:
			if r := []rune(p.find); len(r) > 0 {
				p.find = string(r[:len(r)-1])
			}
		case k.Code == kit.KeyRune:
			p.find += string(k.Rune)
		}
		p.build()
		return true
	}
	if k.Code == kit.KeyRune {
		c := strings.ToUpper(string(k.Rune))
		switch i := strings.Index(logsKeys, c); {
		case i >= 0:
			p.toggleOffice(i + 1)
			return true
		case c == "F":
			p.startFind()
			return true
		case c == "O":
			p.toggleFold()
			return true
		case c == "N":
			p.moveTo(len(p.rows) - 1)
			return true
		case c == "A":
			p.on = [11]bool{}
			for o := 1; o <= 10; o++ {
				p.on[o] = true
			}
			p.build()
			return true
		}
	}
	return p.barPanel.Key(k)
}

// Mouse: the stamps, the lever, the Find slip, the rail, the rows, the
// "new lines" plate and the wheel.
func (p *logsPanel) Mouse(m kit.Mouse) bool {
	if p.chrome == nil {
		return false
	}
	org := p.Bounds().Min
	if p.buttons.mouse(p.chrome, p.hintOrigin(), m) {
		return true
	}
	if m.Kind == kit.MouseWheel {
		p.scrollBy(m.Wheel * 3)
		return true
	}
	if m.Kind != kit.MousePress || m.Button != kit.ButtonLeft {
		return false
	}
	pt, l := m.Pt().Sub(org), p.lay
	switch {
	case l.Find.In(pt.X, pt.Y):
		p.startFind()
	case l.Rail.In(pt.X, pt.Y):
		p.top, p.follow = (pt.Y-l.Rail.Y)*len(p.rows)/max(1, l.Rail.H)-p.height()/2, false
		p.scrollBy(0)
	case p.plate().In(pt.X, pt.Y) && p.unseen() > 0:
		p.moveTo(len(p.rows) - 1)
	case l.Body.In(pt.X, pt.Y):
		if i := p.rowAt(pt.Y); i >= 0 {
			p.moveTo(i)
			p.follow = false
			p.toggleFold()
		}
	default:
		for i, s := range l.Stamps {
			if s.In(pt.X, pt.Y) && l.Mode != logroll.Placeholder {
				p.toggleOffice(i + 1)
				return true
			}
		}
		for i, s := range l.Stops {
			if s.In(pt.X, pt.Y) && l.Mode != logroll.Placeholder {
				p.setLevel(i + 1)
				return true
			}
		}
		return false
	}
	return true
}

// rowAt is the row drawn at panel row y, or -1.
func (p *logsPanel) rowAt(y int) int {
	n, h := len(p.rows), p.height()
	i := p.top + y - p.lay.Body.Y - max(0, h-n)
	if i < p.top || i >= min(n, p.top+h) {
		return -1
	}
	return i
}

// unseen is how many lines were written since the window left the newest.
func (p *logsPanel) unseen() int {
	if p.follow {
		return 0
	}
	return max(0, len(p.state.Log)-p.seen)
}

// plate is the "new lines" plate, at the foot of the roll.
func (p *logsPanel) plate() logroll.Rect {
	b := p.lay.Body
	return logroll.Rect{X: b.Right() - 18, Y: b.Bottom() - 1, W: 17, H: 1}
}

func (p *logsPanel) PointerAt(at image.Point) kit.Pointer {
	if p.chrome == nil {
		return kit.PointerArrow
	}
	if _, key := p.chrome.keyAt(at.Sub(p.hintOrigin())); key {
		return kit.PointerHand
	}
	pt, l := at.Sub(p.Bounds().Min), p.lay
	acts := l.Find.In(pt.X, pt.Y) || l.Rail.In(pt.X, pt.Y) || p.plate().In(pt.X, pt.Y) && p.unseen() > 0
	if i := p.rowAt(pt.Y); l.Body.In(pt.X, pt.Y) && i >= 0 {
		acts = acts || p.rows[i].e.Count > 1 && p.rows[i].kind == rowLine
	}
	for _, s := range append(l.Stamps[:], l.Stops[:]...) {
		acts = acts || s.In(pt.X, pt.Y) && l.Mode != logroll.Placeholder
	}
	return handIf(acts)
}

// hintOrigin is where the hint row's grid starts on the screen.
func (p *logsPanel) hintOrigin() image.Point {
	return p.Bounds().Min.Add(image.Pt(0, p.lay.Hint.Y))
}
