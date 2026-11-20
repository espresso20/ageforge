package ui

import (
	"fmt"
	"image"
	"strings"
	"unicode/utf8"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// ranks_panel.go is the Ranks screen: the whole screen but the command bar.
// Like the Map and Research panels it leaves the keyboard with the command
// bar and takes only the keys that print nothing: Up and Down pick a domain
// (or a badge on the lists), Enter opens the domain's badges, Tab cycles
// ranks / all badges / recent, Esc steps back and then closes. A click picks
// a row or a tab, as the keys do.

const (
	tabRanks = iota
	tabAll
	tabRecent
	numRankTabs
)

var rankTabNames = [numRankTabs]string{"Ranks", "All badges", "Recent"}

// badgeRequest is how a command asks the screen to open: on a tab, on a
// domain's badges, or on one badge. The zero value is the Ranks tab.
// badgeKeys are the screen's keys as the Help panel lists them.
var badgeKeys = []keyHelp{
	{"Up/Down", rowid.E("ui.ranks_panel.badgeKeys.01", "Pick a domain (a badge on the lists)")},
	{"Enter", rowid.E("ui.ranks_panel.badgeKeys.02", "The domain's badges (with something typed, Tab and Enter act on the prompt instead)")},
	{"Tab", rowid.E("ui.ranks_panel.badgeKeys.03", "Ranks, all badges, recent (Shift-Tab: back)")},
	{"PgUp/PgDn", rowid.E("ui.ranks_panel.badgeKeys.04", "A page up and down on the lists")},
	{"Home/End", rowid.E("ui.ranks_panel.badgeKeys.05", "The first and the last")},
	{"Esc", rowid.E("ui.ranks_panel.badgeKeys.06", "Back to Ranks, then close")},
}

type badgeRequest struct {
	tab int
	dom string
	sel string
}

type ranksView struct {
	tab  int
	dom  string // the picked domain on Ranks; the one the list is held to
	list bool   // the all-badges tab shows only dom's badges
	sel  string // the picked badge on a list tab
}

type rankHit struct {
	r  image.Rectangle
	id string
}

type ranksPanel struct {
	barPanel
	kit.Base
	model *ranksModel
	view  ranksView
	rows  []game.BadgeView // what the open list tab shows
	tabs  [numRankTabs]image.Rectangle
	hits  []rankHit
	w, h  int
}

func newRanksPanel() *ranksPanel {
	p := &ranksPanel{w: 120, h: 37}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *ranksPanel) open(state game.GameState) Primitive {
	p.update(state)
	return overBar(p)
}

// request sets what the next open shows.
func (p *ranksPanel) request(r badgeRequest) {
	p.view = ranksView{tab: r.tab, dom: r.dom, list: r.tab == tabAll && r.dom != "", sel: r.sel}
}

// update takes a new snapshot and keeps the pick on something the tab lists.
func (p *ranksPanel) update(state game.GameState) {
	var views []game.BadgeView
	if state.AccountStats != nil {
		views = state.AccountStats.Badges
	}
	p.model = buildRanks(views)
	p.settle()
}

// settle rebuilds the open list and keeps the pick in it.
func (p *ranksPanel) settle() {
	m := p.model
	if m.by[p.view.dom] == nil {
		p.view.dom = m.doms[0].key
	}
	p.rows = nil
	switch {
	case p.view.tab == tabAll && p.view.list:
		p.rows = m.list(p.view.dom)
	case p.view.tab == tabAll:
		p.rows = m.list("")
	case p.view.tab == tabRecent:
		p.rows = m.recent()
	default:
		return
	}
	if p.rowIndex(p.view.sel) < 0 && len(p.rows) > 0 {
		p.view.sel = p.rows[0].Key
	}
}

func (p *ranksPanel) rowIndex(key string) int {
	for i, v := range p.rows {
		if v.Key == key {
			return i
		}
	}
	return -1
}

func (p *ranksPanel) Layout(r image.Rectangle) {
	p.Place(r)
	p.w, p.h = r.Dx(), r.Dy()
}

// rankPaint paints in the panel's own cells.
type rankPaint struct {
	p *kit.Painter
	o image.Point
	w int
}

func (s rankPaint) text(x, y int, str string, col kit.Color, bold bool) int {
	s.p.Print(s.o.X+x, s.o.Y+y, max(0, s.w-x), str, kit.Style{}.Foreground(col).Bold(bold))
	return x + utf8.RuneCountInString(str)
}

func (s rankPaint) right(end, y int, str string, col kit.Color, bold bool) {
	s.text(end-utf8.RuneCountInString(str), y, str, col, bold)
}

func (s rankPaint) fill(x, y, w, h int, bg kit.Color) {
	s.p.Fill(image.Rect(s.o.X+x, s.o.Y+y, s.o.X+x+w, s.o.Y+y+h), bg)
}

func (s rankPaint) rule(x, y, w int, col kit.Color) {
	s.text(x, y, strings.Repeat("─", max(0, w)), col, false)
}

// bar is a thin bar of w cells, frac of it filled.
func (s rankPaint) bar(x, y, w int, frac float64, on, off kit.Color) {
	n := int(min(max(frac, 0), 1)*float64(w) + 0.5)
	for i := 0; i < w; i++ {
		col := off
		if i < n {
			col = on
		}
		s.p.Put(s.o.X+x+i, s.o.Y+y, '▀', kit.Style{}.Foreground(col))
	}
}

// Paint puts the screen on the kit's painter in theme colors.
func (p *ranksPanel) Paint(pt *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.model == nil {
		return
	}
	s := rankPaint{pt.Clip(r), r.Min, r.Dx()}
	bg := theme.Of(theme.RoleBackground)
	s.fill(0, 0, r.Dx(), r.Dy(), bg)
	txt, dim := kit.Role(theme.RoleText), kit.Role(theme.RoleDim)
	p.hits = p.hits[:0]
	m, w, h := p.model, r.Dx(), r.Dy()

	hw := paintHeading(s.p, r.Min.X+2, r.Min.Y, "Ranks", theme.RGBA(theme.RoleAccent)) - r.Min.X
	x := hw + 4
	for i, name := range rankTabNames {
		col, bold := dim, false
		if i == p.view.tab {
			col, bold = kit.Role(theme.RoleAccent), true
		}
		end := s.text(x, 1, name, col, bold)
		p.tabs[i] = image.Rect(x, 1, end, 2)
		x = end + 3
	}
	held := len(m.doms) * len(config.RankThresholds)
	s.text(2, 3, fmt.Sprintf("%d of %d badges earned   %s points   %d of %d ranks",
		m.earned, m.total, config.FormatAmount(float64(m.pts)), m.held, held), txt, false)
	if p.view.tab == tabAll && p.view.list {
		s.right(w-3, 3, "only "+m.by[p.view.dom].name, kit.Role(theme.RoleAccent), true)
	}

	if p.view.tab == tabRanks {
		p.paintRanks(s, w, h)
	} else {
		p.paintList(s, w, h)
	}
	s.rule(2, h-2, w-4, dim)
	hint := "↑ ↓ domain    Enter its badges    Tab ranks / all badges / recent    Esc close"
	if p.view.tab != tabRanks {
		hint = "↑ ↓ badge    PgUp PgDn page    Tab ranks / all badges / recent    Esc back"
	}
	s.text(2, h-1, hint, dim, false)
}

// domCol is a domain's colour on the background; a domain with no rank is
// dimmer.
func domCol(d *rankDom) kit.Color {
	bg := theme.Of(theme.RoleBackground)
	col := theme.Hue(d.hex, bg)
	if d.rank == 0 {
		col = theme.Mix(bg, col, 0.55)
	}
	return col
}

func (p *ranksPanel) paintRanks(s rankPaint, w, h int) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	listW, rx := w-4, w
	if w >= 125 {
		rx = w - 62
		listW = rx - 4
	}
	const y0, bx = 5, 32
	avail := h - y0 - 3
	pitch := 3
	if 4*len(m.doms) <= avail {
		pitch = 4
	}
	vis := max(1, avail/pitch)
	at := 0
	for i, d := range m.doms {
		if d.key == p.view.dom {
			at = i
		}
	}
	top := 0
	if len(m.doms) > vis {
		top = min(max(at-vis/2, 0), len(m.doms)-vis)
	}
	segW := min(max((listW-bx-1)/len(config.RankThresholds)-1, 2), 20)
	barEnd := bx + len(config.RankThresholds)*(segW+1) - 1
	for i := top; i < len(m.doms) && i < top+vis; i++ {
		d, y := m.doms[i], y0+(i-top)*pitch
		col := domCol(d)
		track := theme.Mix(bg, col, 0.25)
		sel := d.key == p.view.dom
		if sel {
			s.fill(1, y, listW, 3, theme.Mix(bg, col, 0.12))
			for j := 0; j < 3; j++ {
				s.text(1, y+j, "▌", col, false)
			}
		}
		s.text(4, y, d.name, col, sel)
		if d.rank > 0 {
			s.text(18, y, config.RankTitles[d.rank-1], bright, true)
		} else {
			s.text(18, y, "No rank yet", dim, false)
		}
		if n := d.toNext(); n > 0 {
			tail := fmt.Sprintf("%d to %s", n, config.RankTitles[d.rank])
			s.right(barEnd, y, tail, txt, sel)
			s.right(barEnd-utf8.RuneCountInString(tail)-3, y, fmt.Sprintf("%d points", d.pts), dim, false)
		}
		for k, th := range config.RankThresholds {
			lo := 0
			if k > 0 {
				lo = config.RankThresholds[k-1]
			}
			s.bar(bx+k*(segW+1), y+1, segW, float64(d.pts-lo)/float64(th-lo), col, track)
		}
		s.text(bx, y+2, fmt.Sprintf("%d of %d badges", d.earned, d.all), dim, false)
		p.hits = append(p.hits, rankHit{image.Rect(1, y, listW+1, y+3), d.key})
	}
	if rx < w {
		p.paintDetail(s, m.by[p.view.dom], rx, w-2-rx, h)
	}
}

func (p *ranksPanel) paintDetail(s rankPaint, d *rankDom, rx, rw, h int) {
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	col := domCol(d)
	s.fill(rx-2, 5, rw+3, h-8, theme.Mix(bg, col, 0.08))
	ink, _ := col.RGBA()
	x := paintHeading(s.p, s.o.X+rx, s.o.Y+5, d.name, ink) - s.o.X + 2
	if d.rank > 0 {
		s.text(x, 5, fmt.Sprintf("rank %d  %s", d.rank, config.RankTitles[d.rank-1]), bright, true)
	} else {
		s.text(x, 5, "No rank yet", dim, false)
	}
	s.text(x, 6, fmt.Sprintf("%d points", d.pts), dim, false)
	s.right(rx+rw, 5, fmt.Sprintf("%d of %d badges", d.earned, d.all), txt, false)
	if n := d.toNext(); n > 0 {
		s.right(rx+rw, 6, fmt.Sprintf("%d points to %s", n, config.RankTitles[d.rank]), col, true)
	}
	s.rule(rx, 7, rw, col)
	y := 9
	if d.rank == 0 && d.opens != "" {
		s.text(rx, y, "Its workers arrive in the "+d.opens+".", dim, false)
		y += 2
	}
	s.text(rx, y, "NEAREST BADGES", dim, false)
	y += 2
	for _, v := range d.near {
		if y+3 > h-14 {
			break
		}
		s.text(rx, y, "□", dim, false)
		s.text(rx+2, y, v.Name, bright, true)
		s.right(rx+rw, y, fmt.Sprintf("%s  +%d", v.Tier, v.Points), txt, false)
		s.text(rx+2, y+1, truncate(v.Desc, rw-2), txt, false)
		pct := v.Progress / v.Target
		s.bar(rx+2, y+2, rw-8, pct, col, theme.Mix(bg, col, 0.25))
		s.right(rx+rw, y+2, fmt.Sprintf("%d%%", int(min(pct, 1)*100)), txt, false)
		y += 4
	}
	s.text(rx, y, "EARNED", dim, false)
	y++
	for t := config.BadgeBronze; t <= config.BadgeLegendary; t++ {
		if d.total[t] == 0 {
			continue
		}
		x := s.text(rx, y, strings.ToUpper(t.Name()), txt, true)
		s.text(x+2, y, fmt.Sprintf("%d points each", t.Points()), dim, false)
		s.right(rx+rw, y, fmt.Sprintf("%d of %d", d.have[t], d.total[t]), txt, false)
		y++
	}
	y += 2
	if y+4 <= h-3 {
		s.text(rx, y, "WHAT A RANK GIVES", dim, false)
		for i, g := range [][2]string{
			{"Each rank", fmt.Sprintf("+%d%% output from the domain's workers", int(config.RankBonus*100))},
			{"Rank 3 " + config.RankTitles[2], "the domain's title beside the town name"},
			{"Rank 6 " + config.RankTitles[5], "a color for the theme list"},
		} {
			s.text(rx, y+1+i, g[0], bright, true)
			s.text(rx+21, y+1+i, g[1], txt, false)
		}
	}
}

func (p *ranksPanel) paintList(s rankPaint, w, h int) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	nameW := min(max(w-70, 24), 48)
	xd := 4 + nameW + 2
	xt, xp := xd+14, xd+25
	s.text(4, 5, "BADGE", dim, false)
	s.text(xd, 5, "DOMAIN", dim, false)
	s.text(xt, 5, "TIER", dim, false)
	s.text(xp, 5, "POINTS", dim, false)
	if p.view.tab == tabRecent {
		s.right(w-3, 5, "EARNED", dim, false)
	} else {
		s.right(w-3, 5, "PROGRESS", dim, false)
	}
	if len(p.rows) == 0 {
		s.text(4, 7, "Nothing here yet.", dim, false)
		return
	}
	vis := p.listRows()
	at := max(p.rowIndex(p.view.sel), 0)
	top := 0
	if len(p.rows) > vis {
		top = min(max(at-vis/2, 0), len(p.rows)-vis)
	}
	for i := top; i < len(p.rows) && i < top+vis; i++ {
		v, y := p.rows[i], 6+i-top
		d := m.by[v.Domain]
		col := dim
		if d != nil {
			col = domCol(d)
		}
		if v.Key == p.view.sel {
			s.fill(1, y, w-2, 1, theme.Mix(bg, col, 0.12))
			s.text(1, y, "▌", col, false)
		}
		mark, name, ink := "□", v.Name, txt
		if v.Earned {
			mark, ink = "✓", bright
		}
		if v.Hidden {
			name, ink = game.BadgeHiddenName, dim
		}
		s.text(2, y, mark, col, false)
		s.text(4, y, truncate(name, nameW), ink, v.Earned)
		if d != nil {
			s.text(xd, y, d.name, col, false)
		}
		s.text(xt, y, v.Tier, dim, false)
		s.text(xp, y, fmt.Sprintf("+%d", v.Points), txt, false)
		tail := ""
		switch {
		case p.view.tab == tabRecent && !v.At.IsZero():
			tail = v.At.Format("2 Jan 2006")
		case !v.Earned && v.Target > 0:
			tail = fmt.Sprintf("%d%%", int(min(v.Progress/v.Target, 1)*100))
		}
		s.right(w-3, y, tail, dim, false)
		p.hits = append(p.hits, rankHit{image.Rect(1, y, w-1, y+1), v.Key})
	}
}

// listRows is how many badges a list tab shows at once.
func (p *ranksPanel) listRows() int { return max(1, p.h-8) }

// Mouse: a click on a tab opens it, a click on a row picks it (a double click
// on a domain opens its badges, as Enter does), the wheel moves the pick.
func (p *ranksPanel) Mouse(m kit.Mouse) bool {
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
	for i, r := range p.tabs {
		if pt.In(r) {
			p.setTab(i)
			return true
		}
	}
	for _, h := range p.hits {
		if pt.In(h.r) {
			p.pick(h.id)
			if m.Clicks == 2 {
				p.handleKey(kit.KeyEnter)
			}
			return true
		}
	}
	return false
}

// PointerAt: the hand over a tab or a row a click would pick.
func (p *ranksPanel) PointerAt(at image.Point) kit.Pointer {
	pt := at.Sub(p.Bounds().Min)
	for i, r := range p.tabs {
		if pt.In(r) {
			return handIf(i != p.view.tab)
		}
	}
	for _, h := range p.hits {
		if pt.In(h.r) {
			return kit.PointerHand
		}
	}
	return kit.PointerArrow
}

func (p *ranksPanel) pick(id string) {
	if p.view.tab == tabRanks {
		p.view.dom = id
	} else {
		p.view.sel = id
	}
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt, Tab and Enter belong to the prompt.
func (p *ranksPanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome, kit.KeyEnd:
	case kit.KeyTab, kit.KeyBacktab, kit.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Code)
	return true
}

// closeBack steps back to the Ranks tab and reports whether it was off it:
// Esc steps back first, then closes the panel.
func (p *ranksPanel) closeBack() bool {
	if p.view.tab == tabRanks {
		return false
	}
	p.setTab(tabRanks)
	return true
}

func (p *ranksPanel) handleKey(key kit.KeyCode) {
	if p.model == nil {
		return
	}
	switch key {
	case kit.KeyUp:
		p.move(-1)
	case kit.KeyDown:
		p.move(1)
	case kit.KeyPgUp:
		p.move(-p.listRows())
	case kit.KeyPgDn:
		p.move(p.listRows())
	case kit.KeyHome:
		p.move(-1 << 20)
	case kit.KeyEnd:
		p.move(1 << 20)
	case kit.KeyTab:
		p.setTab((p.view.tab + 1) % numRankTabs)
	case kit.KeyBacktab:
		p.setTab((p.view.tab + numRankTabs - 1) % numRankTabs)
	case kit.KeyEnter:
		if p.view.tab == tabRanks {
			p.view.tab, p.view.list, p.view.sel = tabAll, true, ""
			p.settle()
		}
	}
}

// move picks the domain or the badge n places on, stopping at the ends.
func (p *ranksPanel) move(n int) {
	if p.view.tab == tabRanks {
		at := 0
		for i, d := range p.model.doms {
			if d.key == p.view.dom {
				at = i
			}
		}
		p.view.dom = p.model.doms[clamp(at+n, 0, len(p.model.doms)-1)].key
		return
	}
	if len(p.rows) > 0 {
		p.view.sel = p.rows[clamp(max(p.rowIndex(p.view.sel), 0)+n, 0, len(p.rows)-1)].Key
	}
}

// setTab opens a tab from its top, with every domain's badges listed.
func (p *ranksPanel) setTab(t int) {
	p.view.tab, p.view.list, p.view.sel = t, false, ""
	p.settle()
}

// findBadge reads a word as a badge in sight: its name, the start of it, or a
// part of it, in that order. A withheld badge has no name to find.
func findBadge(views []game.BadgeView, word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return "", false
	}
	best, rank := "", 3
	for _, v := range views {
		if v.Hidden {
			continue
		}
		name := strings.ToLower(v.Name)
		r := 3
		switch {
		case name == word:
			r = 0
		case strings.HasPrefix(name, word):
			r = 1
		case strings.Contains(name, word):
			r = 2
		}
		if r < rank {
			best, rank = v.Key, r
		}
	}
	return best, best != ""
}
