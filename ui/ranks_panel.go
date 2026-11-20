package ui

import (
	"fmt"
	"image"
	"image/color"
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
	keys  *xgrid // the buttons drawn at the last paint
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
	p.model = buildRanks(state.Ruleset(), views)
	p.settle()
}

// settle rebuilds the open list and keeps the pick in it.
func (p *ranksPanel) settle() {
	m := p.model
	if m.by[p.view.dom] == nil {
		p.view.dom = m.doms[0].key
		if near := m.closest(); len(near) > 0 {
			p.view.dom = near[0].key
		}
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
func (s rankPaint) bar(x, y, w int, frac float64, on, off kit.Color, r rune) {
	n := int(min(max(frac, 0), 1)*float64(w) + 0.5)
	for i := 0; i < w; i++ {
		col := off
		if i < n {
			col = on
		}
		s.p.Put(s.o.X+x+i, s.o.Y+y, r, kit.Style{}.Foreground(col))
	}
}

// Paint puts the screen on the kit's painter in theme colors.
func (p *ranksPanel) Paint(pt *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.model == nil {
		return
	}
	s := rankPaint{pt.Clip(r), r.Min, r.Dx()}
	s.fill(0, 0, r.Dx(), r.Dy(), theme.Of(theme.RoleBackground))
	txt, dim := kit.Role(theme.RoleText), kit.Role(theme.RoleDim)
	p.hits = p.hits[:0]
	g := newXGrid(r.Dx(), r.Dy())
	p.keys = g
	m, w, h := p.model, r.Dx(), r.Dy()

	s.text(2, 0, "BADGES", kit.Role(theme.RoleAccent), true)
	x := 10
	for i, name := range rankTabNames {
		col, bold := dim, false
		if i == p.view.tab {
			col, bold = kit.Role(theme.RoleBright), true
		}
		end := s.text(x, 0, strings.ToLower(name), col, bold)
		p.tabs[i] = image.Rect(x, 0, end, 1)
		x = end + 3
	}
	g.closeButton(w)

	if p.view.tab == tabRanks {
		p.paintRanks(s, w, h)
	} else {
		held := len(m.doms) * len(m.thr)
		s.text(2, 3, fmt.Sprintf("%d of %d badges earned   %s points   %d of %d ranks",
			m.earned, m.total, config.FormatAmount(float64(m.pts)), m.held, held), txt, false)
		if p.view.tab == tabAll && p.view.list {
			s.right(w-3, 3, "only "+m.by[p.view.dom].name, kit.Role(theme.RoleAccent), true)
		}
		p.paintList(s, w, h)
	}
	s.rule(2, h-3, w-4, faintCol())
	foot := [][2]string{{"↑ ↓", "domain"}, {"Enter", "open the list"}, {"Tab", "ranks / all badges / recent"}, {"Esc", "close"}}
	if p.view.tab != tabRanks {
		foot = [][2]string{{"↑ ↓", "badge"}, {"PgUp PgDn", "page"}, {"Tab", "ranks / all badges / recent"}, {"Esc", "back"}}
	}
	s.footKeys(g, h-2, foot, dim)
	p.buttons.paint(pt, g, r.Min, newExpPalette())
}

// footKeys is a row of key buttons at y: each a cap and its word.
func (s rankPaint) footKeys(g *xgrid, y int, keys [][2]string, word kit.Color) {
	x := 2
	for _, k := range keys {
		end := g.keyButton(x, y, k[0], k[1])
		s.text(x+runeLen(k[0])+3, y, k[1], word, false)
		x = end + 1
	}
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

// pixelInk is a color as the pixels of a picture take it.
func pixelInk(c kit.Color) color.RGBA {
	v, _ := c.RGBA()
	return v
}

// tierCol is a tier's color; legendary takes the accent.
func tierCol(t config.BadgeTier) kit.Color {
	var h theme.BadgeHue
	switch t {
	case config.BadgeBronze:
		h = theme.BadgeBronze
	case config.BadgeSilver:
		h = theme.BadgeSilver
	case config.BadgeGold:
		h = theme.BadgeGold
	case config.BadgePlatinum:
		h = theme.BadgePlatinum
	default:
		return kit.Role(theme.RoleAccent)
	}
	return theme.Legible(theme.BadgeHueColor(h, theme.IsLight()), theme.Of(theme.RoleBackground), 3)
}

var rankTiers = [...]config.BadgeTier{config.BadgeBronze, config.BadgeSilver, config.BadgeGold,
	config.BadgePlatinum, config.BadgeLegendary}

// ranksGeo is where the Ranks tab's columns fall for a panel of w by h.
type ranksGeo struct {
	w, h         int
	bx, segW     int  // the bars start here, each segW cells wide
	bxEnd        int  // the last cell of the six bars
	detail       bool // the selected domain's column is shown
	rx, rw       int  // its left edge and width
	y0, pitch    int  // the first row and the rows a domain takes
	rowsEnd, top int  // the row under the last domain; the first one shown
	band         bool // the bottom band fits
}

func (p *ranksPanel) geometry(w, h int) ranksGeo {
	m := p.model
	g := ranksGeo{w: w, h: h, bx: 34, y0: 9}
	g.detail = w >= 140
	right := w - 8
	if g.detail {
		g.rw = min(max((w*30+50)/100, 50), 90)
		g.rx = w - 3 - g.rw
		right = g.rx - 9
	}
	g.segW = max((right-g.bx+1)/len(m.thr)-1, 2)
	g.bxEnd = g.bx + len(m.thr)*(g.segW+1) - 1
	div := 0
	if m.doms[0].rank > 0 && m.doms[len(m.doms)-1].rank == 0 {
		div = 2
	}
	avail := h - 3 - g.y0
	g.pitch = 3
	if 4*len(m.doms)+div <= avail+1 {
		g.pitch = 4
	}
	g.rowsEnd = g.y0 + g.pitch*len(m.doms) + div
	g.band = g.bxEnd >= 144 && g.rowsEnd+14 <= h-3 && g.rowsEnd-g.pitch+g.pitch <= h
	return g
}

func (p *ranksPanel) paintRanks(s rankPaint, w, h int) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	faint := faintCol()
	g := p.geometry(w, h)
	sel := m.by[p.view.dom]
	selCol := domCol(sel)
	zero := kit.Style{}

	// The tint under the selected row and its detail column goes down first.
	tint := theme.Mix(bg, selCol, 0.10)
	if g.detail {
		s.fill(g.rx-2, 5, w-1-(g.rx-2), max(h-9, 0), tint)
	}

	// Where each domain's row falls: the ranked first, a divider, then the
	// rest. With more domains than rows the list scrolls to keep the pick.
	type slot struct {
		d *rankDom
		y int
	}
	var slots []slot
	y := 0
	divAt := -1
	for i, d := range m.doms {
		if i > 0 && d.rank == 0 && m.doms[i-1].rank > 0 {
			divAt = y
			y += 2
		}
		slots = append(slots, slot{d, y})
		y += g.pitch
	}
	avail := h - 3 - g.y0
	off := 0
	if y > avail+1 {
		for _, sl := range slots {
			if sl.d == sel {
				off = min(max(sl.y-avail/2, 0), max(y-avail, 0))
			}
		}
	}
	rowTop := g.y0 - off

	// The top band: three figures, one grammar.
	block := func(x int, label string, fig string, col kit.Color, l1 string, c1 kit.Color, l2 string) int {
		s.text(x, 2, label, dim, false)
		fx := paintHeading(s.p, s.o.X+x, s.o.Y+3, fig, pixelInk(col)) - s.o.X + 2
		s.text(fx, 3, l1, c1, true)
		s.text(fx, 4, l2, txt, false)
		return fx + max(utf8.RuneCountInString(l1), utf8.RuneCountInString(l2)) + 4
	}
	ex := block(4, "BADGES EARNED", fmt.Sprint(m.earned), bright, fmt.Sprintf("of %d badges", m.total), bright,
		fmt.Sprintf("%d still to earn", m.total-m.earned))
	s.text(ex, 3, config.FormatAmount(float64(m.pts))+" points", dim, false)
	s.text(ex, 4, fmt.Sprintf("%d of %d ranks", m.held, len(m.doms)*len(m.thr)), dim, false)
	x2 := max((w*35+50)/100, 56)
	if top := m.top(); top != nil {
		tc := domCol(top)
		ex = block(x2, "HIGHEST RANK", m.titles[top.rank-1], tc, "of "+top.name, tc,
			fmt.Sprintf("rank %d of %d", top.rank, len(m.titles)))
		s.text(ex, 3, fmt.Sprintf("%d points", top.pts), dim, false)
		if n := top.toNext(); n > 0 {
			s.text(ex, 4, fmt.Sprintf("%d to %s", n, m.titles[top.rank]), dim, false)
		}
	} else {
		s.text(x2, 2, "HIGHEST RANK", dim, false)
		s.text(x2, 3, "none yet", dim, false)
		ex = x2 + 12
	}
	if near := m.closest(); g.detail && len(near) > 0 && ex+2 <= g.rx {
		c0 := near[0]
		cc := domCol(c0)
		cx := block(g.rx, "NEXT PROMOTION", fmt.Sprint(c0.toNext()), cc, c0.name, cc,
			"points to "+m.titles[c0.rank])
		_ = cx
		for i := 1; i < 3 && i < len(near); i++ {
			c := near[i]
			fx := g.rx + headWidth(fmt.Sprint(c0.toNext())) + 2
			xx := s.text(fx+max(utf8.RuneCountInString(c0.name), utf8.RuneCountInString("points to "+m.titles[c0.rank]))+4, 2+i, c.name, domCol(c), false)
			s.text(xx+1, 2+i, fmt.Sprintf("%d to %s", c.toNext(), m.titles[c.rank]), dim, false)
		}
	}

	// The ruler: six ranks, the points each starts at on the rule.
	for k, name := range m.titles {
		x0 := g.bx + k*(g.segW+1)
		end := x0 + g.segW
		col := dim
		if k < sel.rank {
			col = txt
		}
		label := fmt.Sprintf("%d %s", k+1, strings.ToUpper(name))
		if utf8.RuneCountInString(label) > g.segW {
			label = fmt.Sprint(k + 1)
		}
		s.right(end, 6, label, col, k+1 == sel.rank)
		s.rule(x0, 7, g.segW, faint)
		s.right(end-1, 7, fmt.Sprintf(" %d ", m.thr[k]), dim, false)
		tick := "┬"
		if k == len(m.titles)-1 {
			tick = "┐"
		}
		s.text(end, 7, tick, dim, false)
	}
	s.text(4, 6, "DOMAIN", dim, false)
	s.right(31, 6, "RANK", dim, false)

	fits := func(y int) bool { return y >= g.y0 && y+3 <= h-3 }
	if divAt >= 0 && fits(rowTop+divAt) {
		dy := rowTop + divAt
		w2 := g.bxEnd + 6
		s.rule(2, dy, w2-2, faint)
		s.text(4, dy, " NO RANK YET ", dim, false)
	}
	for _, sl := range slots {
		d, y := sl.d, rowTop+sl.y
		if !fits(y) {
			continue
		}
		ranked := d.rank > 0
		col := domCol(d)
		track := theme.Mix(bg, col, 0.20)
		if !ranked {
			track = theme.Mix(bg, col, 0.14)
		}
		if d == sel {
			s.fill(1, y, g.bxEnd+6, 3, tint)
			if g.pitch == 4 {
				for x := 1; x <= g.bxEnd+6; x++ {
					s.p.Put(s.o.X+x, s.o.Y+y-1, '▄', zero.Foreground(tint))
					s.p.Put(s.o.X+x, s.o.Y+y+3, '▀', zero.Foreground(tint))
				}
			}
			for j := 0; j < 3; j++ {
				s.text(1, y+j, "▌", selCol, false)
			}
		}
		paintHeading(s.p, s.o.X+4, s.o.Y+y, d.name, pixelInk(col))
		if ranked {
			n := fmt.Sprint(d.rank)
			paintHeading(s.p, s.o.X+31-headWidth(n), s.o.Y+y, n, pixelInk(bright))
			s.text(g.bx, y, m.titles[d.rank-1], bright, true)
		} else {
			s.text(30, y, "-", dim, false)
			s.text(g.bx, y, "No rank yet", dim, false)
		}
		if n := d.toNext(); n > 0 {
			right := fmt.Sprintf("%d to %s", n, m.titles[d.rank])
			rc := txt
			switch {
			case d == sel:
				rc = selCol
			case !ranked:
				rc = dim
			}
			s.right(g.bxEnd, y, right, rc, d == sel)
			s.right(g.bxEnd-utf8.RuneCountInString(right)-3, y, fmt.Sprintf("%d points", d.pts), dim, false)
		}
		for k, th := range m.thr {
			lo := 0
			if k > 0 {
				lo = m.thr[k-1]
			}
			s.bar(g.bx+k*(g.segW+1), y+1, g.segW, float64(d.pts-lo)/float64(th-lo), col, track, '▀')
		}
		s.text(g.bx, y+2, fmt.Sprintf("%d of %d badges", d.earned, d.all), dim, false)
		switch {
		case !ranked && d.opens != "":
			s.right(g.bxEnd, y+2, "its workers arrive in the "+d.opens, dim, false)
		case len(d.near) > 0:
			v := d.near[0]
			pc := fmt.Sprintf("%d%%", pct(v))
			s.right(g.bxEnd, y+2, pc, txt, false)
			nx := g.bxEnd - len(pc) - 2
			s.right(nx, y+2, v.Name, txt, false)
			s.right(nx-utf8.RuneCountInString(v.Name)-2, y+2, "next", dim, false)
		}
		p.hits = append(p.hits, rankHit{image.Rect(1, y, g.bxEnd+7, y+3), d.key})
	}

	if g.band {
		p.paintBand(s, g, g.rowsEnd)
	}
	if g.detail {
		p.paintDetail(s, sel, g)
	}
}

// paintBand is the bottom band: what came in lately, what a rank is for, and
// the tiers.
func (p *ranksPanel) paintBand(s rankPaint, g ranksGeo, by int) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	s.rule(2, by, g.rx-6, faintCol())
	s.text(4, by+1, "LATELY", dim, false)
	n := min(12, g.h-3-(by+2))
	for i, v := range m.recent() {
		if i >= n {
			break
		}
		ly := by + 2 + i
		col := dim
		dom := ""
		if d := m.by[v.Domain]; d != nil {
			col, dom = domCol(d), d.name
		}
		s.text(4, ly, "✓", col, false)
		s.text(6, ly, truncate(v.Name, 27), txt, false)
		s.text(34, ly, dom, col, false)
		s.text(48, ly, fmt.Sprintf("+%d", v.Points), txt, false)
		s.text(54, ly, v.Tier, tierCol(v.Level), false)
		s.text(66, ly, m.when(v.At), dim, false)
	}
	const lx = 84
	s.text(lx, by+1, "WHAT A RANK GIVES", dim, false)
	for i, gv := range [][2]string{
		{"Each rank", fmt.Sprintf("+%d%% output from the domain's workers", int(config.RankBonus*100))},
		{"Rank 3 " + m.titles[2], "the domain's title beside your town name"},
		{"Rank 6 " + m.titles[5], "a theme color (not in yet)"},
	} {
		s.text(lx, by+2+i, gv[0], bright, true)
		s.text(lx+21, by+2+i, gv[1], txt, false)
	}
	s.text(lx, by+7, "POINTS BY TIER", dim, false)
	s.right(g.bxEnd, by+7, "earned, every domain", dim, false)
	for i, t := range rankTiers {
		ty := by + 8 + i
		tc := tierCol(t)
		s.text(lx, ty, strings.ToUpper(t.Name()), tc, true)
		s.text(lx+11, ty, fmt.Sprintf("%d points each", t.Points()), dim, false)
		frac := 0.0
		if m.tierAll[t] > 0 {
			frac = float64(m.tierHave[t]) / float64(m.tierAll[t])
		}
		s.bar(lx+28, ty, g.bxEnd-(lx+28)-13, frac, tc, theme.Mix(bg, tc, 0.20), '▄')
		s.right(g.bxEnd, ty, fmt.Sprintf("%d of %d", m.tierHave[t], m.tierAll[t]), txt, false)
	}
}

// nameRuns joins names with a middle dot, wrapping at whole names.
func nameRuns(names []string, n int) []string {
	var out []string
	cur := ""
	for _, s := range names {
		switch {
		case cur == "":
			cur = s
		case utf8.RuneCountInString(cur)+3+utf8.RuneCountInString(s) <= n:
			cur += " · " + s
		default:
			out = append(out, cur)
			cur = s
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// dots sets a run of names with its separators in the faint color.
func (s rankPaint) dots(x, y int, line string, col, faint kit.Color) {
	for i, part := range strings.Split(line, " · ") {
		if i > 0 {
			x = s.text(x, y, " · ", faint, false)
		}
		x = s.text(x, y, part, col, false)
	}
}

func (p *ranksPanel) paintDetail(s rankPaint, d *rankDom, g ranksGeo) {
	m := p.model
	bg := theme.Of(theme.RoleBackground)
	txt, dim, bright := kit.Role(theme.RoleText), kit.Role(theme.RoleDim), kit.Role(theme.RoleBright)
	col := domCol(d)
	rx, rw, h := g.rx, g.rw, g.h
	line := theme.Mix(bg, col, 0.35)
	x := paintHeading(s.p, s.o.X+rx, s.o.Y+6, d.name, pixelInk(col)) - s.o.X + 2
	if d.rank > 0 {
		s.text(x, 6, fmt.Sprintf("rank %d  %s", d.rank, m.titles[d.rank-1]), bright, true)
	} else {
		s.text(x, 6, "No rank yet", dim, false)
	}
	s.text(x, 7, fmt.Sprintf("%d points", d.pts), dim, false)
	s.right(rx+rw, 6, fmt.Sprintf("%d of %d badges", d.earned, d.all), txt, false)
	if n := d.toNext(); n > 0 {
		s.right(rx+rw, 7, fmt.Sprintf("%d points to %s", n, m.titles[d.rank]), col, true)
	}
	s.rule(rx, 8, rw, line)

	// The earned names by tier are measured first: the nearest badges get the rest.
	type run struct {
		tier      config.BadgeTier
		have, all int
		lines     []string
	}
	var runs []run
	earnedRows := 2
	for _, t := range rankTiers {
		if d.total[t] == 0 {
			continue
		}
		r := run{t, d.have[t], d.total[t], nameRuns(d.names[t], rw-2)}
		runs = append(runs, r)
		earnedRows += 2 + max(1, len(r.lines))
	}
	earnedRows--
	foot := h - 7

	ry := 10
	if d.rank == 0 && d.opens != "" {
		s.text(rx, ry, "Its workers arrive in the "+d.opens+".", dim, false)
		ry += 2
	}
	s.text(rx, ry, "NEAREST BADGES", dim, false)
	ry += 2
	for _, v := range d.near {
		lines := wrapCells(v.Desc, rw-2)
		if len(lines) > 2 {
			lines = lines[:2]
		}
		if ry+len(lines)+3+earnedRows > foot-1 {
			break
		}
		s.text(rx, ry, "□", dim, false)
		s.text(rx+2, ry, v.Name, bright, true)
		s.right(rx+rw, ry, fmt.Sprintf("%s  +%d", v.Tier, v.Points), tierCol(v.Level), false)
		for j, ln := range lines {
			s.text(rx+2, ry+1+j, ln, txt, false)
		}
		by := ry + 1 + len(lines)
		s.bar(rx+2, by, rw-8, float64(pct(v))/100, col, theme.Mix(bg, col, 0.22), '▄')
		s.right(rx+rw, by, fmt.Sprintf("%d%%", pct(v)), txt, false)
		ry = by + 2
	}

	s.text(rx, ry, "EARNED", dim, false)
	ry += 2
	for _, r := range runs {
		if ry >= foot-1 {
			break
		}
		tc := tierCol(r.tier)
		x := s.text(rx, ry, strings.ToUpper(r.tier.Name()), tc, true)
		s.text(x+2, ry, fmt.Sprintf("%d points each", r.tier.Points()), dim, false)
		s.right(rx+rw, ry, fmt.Sprintf("%d of %d", r.have, r.all), txt, false)
		ry++
		if len(r.lines) == 0 {
			s.text(rx+2, ry, "none yet", faintCol(), false)
			ry += 2
			continue
		}
		for _, ln := range r.lines {
			if ry < foot-1 {
				s.dots(rx+2, ry, ln, txt, faintCol())
			}
			ry++
		}
		ry++
	}
	s.rule(rx, foot, rw, line)
	s.text(rx, foot+1, fmt.Sprintf("%d still to earn in %s", d.all-d.earned, d.name), txt, false)
	s.right(rx+rw, foot+1, "Enter  the whole list", dim, false)
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
	if p.keys != nil && p.buttons.mouse(p.keys, p.Bounds().Min, m) {
		return true
	}
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
	if p.keys != nil && p.buttons.pointer(p.keys, p.Bounds().Min, at) == kit.PointerHand {
		return kit.PointerHand
	}
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

// faintCol is the color of rules and separators: the dim text, mostly faded
// into the background.
func faintCol() kit.Color {
	return theme.Mix(theme.Of(theme.RoleBackground), kit.Role(theme.RoleDim), 0.7)
}
