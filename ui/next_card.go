package ui

import (
	"strings"

	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
)

// next_card.go fills the Next box: the card for the next step, then the rule
// row and the list of panels under it. One function writes it for both
// layouts; the classic one gets the list alone, in words.

var (
	txtNextTitle = rowid.E("ui.next_card.txtNext.01", " Next ")
	txtPanels    = rowid.E("ui.next_card.txtPanels.01", " Panels ")
	txtNow       = rowid.E("ui.next_card.txtNow.01", "NOW")
	txtThen      = rowid.E("ui.next_card.txtThen.01", "THEN")
	txtShowMe    = rowid.E("ui.next_card.txtShowMe.01", " Show me ")
	txtTabKey    = rowid.E("ui.next_card.txtTabKey.01", " Tab ")
	txtEnterKey  = rowid.E("ui.next_card.txtEnterKey.01", " Enter ")
	regionShow   = "nextshow"
	regionThen   = "nextthen"
)

// The forms of the list, richest first.
const (
	formStack = iota // one column a group, stacked
	formRich         // two columns: Town and Progress, World and You
	formGrouped
	formWords
)

const (
	richColW      = 29 // a rich column: the name field, a blank, the job line
	richMinT      = 59 // the width the rich form needs
	groupedMinT   = 28
	groupTitleW   = 10
	tabMinT       = 30 // under this width the Tab plate and the command go
	panelRuleRows = 1
)

// isPanel reports whether name is a panel on the Panels list. A step that
// lives on the main view (a festival, say) names no panel and shows no word.
func isPanel(name string) bool {
	for _, g := range panelGroups {
		for _, p := range g.panels {
			if p.name == name {
				return true
			}
		}
	}
	return false
}

// cell is a piece of a row with the cells it takes.
type cell struct {
	s     string
	w     int
	fresh bool // the cell starts with a plate, which takes the cell before it
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func dimText(s string) string  { return theme.Paint(theme.RoleDim, s) }
func bodyText(s string) string { return theme.Paint(theme.RoleText, s) }

// panelRows is the found panels in a form, T cells wide, a pad cell at the
// left of each row.
func panelRows(form, T int, v panelView) []string {
	var groups [][]panelInfo
	var titles []string
	for _, g := range panelGroups {
		var ps []panelInfo
		for _, p := range g.panels {
			if v.has(p.name) {
				ps = append(ps, p)
			}
		}
		if len(ps) > 0 {
			groups, titles = append(groups, ps), append(titles, strings.ToUpper(g.title.String()))
		}
	}
	if form == formStack && T >= richMinT {
		return stackRows(groups, titles, v)
	}
	if form == formRich && T >= richMinT {
		return richRows(groups, titles, T, v)
	}
	// flow writes names two blanks apart (one in the narrowest box, which
	// has rows for no more), wrapped at room.
	gap := 2
	if T < 24 {
		gap = 1
	}
	// A new name sits on a plate that borrows the cell before it and the
	// cell after, so the letters keep their columns. lead is what comes
	// before a row's first name: a pad cell, or the group's heading.
	flow := func(names []string, room int, lead func(row int, fresh bool) string) []string {
		var rows []string
		var row string
		used, prev, count := 0, false, 0
		for _, n := range names {
			fresh := v.fresh(n)
			w := len(n)
			if fresh {
				w += 2
			}
			if count > 0 && used+gap+w > room {
				rows, row, used, count, prev = append(rows, row), "", 0, 0, false
			}
			shown := panelLink(n, bodyText(n))
			if fresh {
				shown = panelLink(n, newPlate(n))
			}
			if count == 0 {
				row, used = lead(len(rows), fresh), 0
			} else {
				used += gap
				row += strings.Repeat(" ", max(gap-b2i(fresh)-b2i(prev), 0))
			}
			row, used, prev, count = row+shown, used+w, fresh, count+1
		}
		return append(rows, row)
	}
	if form <= formGrouped && T >= groupedMinT {
		var out []string
		for i, ps := range groups {
			names := make([]string, len(ps))
			for j, p := range ps {
				names[j] = p.name
			}
			out = append(out, flow(names, T-groupTitleW, func(j int, fresh bool) string {
				if j == 0 {
					return " " + dimText(titles[i]+strings.Repeat(" ", groupTitleW-len(titles[i])-b2i(fresh)))
				}
				return strings.Repeat(" ", 1+groupTitleW-b2i(fresh))
			})...)
		}
		return out
	}
	var names []string
	for _, ps := range groups {
		for _, p := range ps {
			names = append(names, p.name)
		}
	}
	return flow(names, T, func(_ int, fresh bool) string { return strings.Repeat(" ", 1-b2i(fresh)) })
}

// stackRows is the list with one column a group, stacked in the order Town,
// World, Progress, You, a blank between groups: 25 rows. The name and what it
// is for sit on one row, as in richRows.
func stackRows(groups [][]panelInfo, titles []string, v panelView) []string {
	var out []string
	for _, i := range []int{0, 2, 1, 3} {
		if i >= len(groups) {
			continue
		}
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, " "+dimText(titles[i]))
		for _, p := range groups[i] {
			name, fresh := bodyText(p.name+strings.Repeat(" ", panelNameW-len(p.name)+1)), v.fresh(p.name)
			if fresh {
				name = newPlate(p.name) + strings.Repeat(" ", panelNameW-len(p.name))
			}
			out = append(out, strings.Repeat(" ", 1-b2i(fresh))+panelLink(p.name, name+dimText(p.job.String())))
		}
	}
	return out
}

// richRows is the two-column form: Town then Progress down the left, World
// then You down the right, a gutter of T-58 between.
func richRows(groups [][]panelInfo, titles []string, T int, v panelView) []string {
	column := func(idx ...int) []cell {
		var out []cell
		for _, i := range idx {
			if i >= len(groups) {
				continue
			}
			if len(out) > 0 {
				out = append(out, cell{})
			}
			out = append(out, cell{dimText(titles[i]), len(titles[i]), false})
			for _, p := range groups[i] {
				job := p.job.String()
				row := p.name + strings.Repeat(" ", panelNameW-len(p.name)+1) + job
				name, fresh := bodyText(p.name+strings.Repeat(" ", panelNameW-len(p.name)+1)), v.fresh(p.name)
				if fresh {
					name = newPlate(p.name) + strings.Repeat(" ", panelNameW-len(p.name))
				}
				out = append(out, cell{panelLink(p.name, name+dimText(job)), len(row), fresh})
			}
		}
		return out
	}
	// groups holds the found ones only, so a left column with no group is
	// the one that is empty and the right moves up into it.
	left, right := column(0, 1), column(2, 3)
	var out []string
	for i := 0; i < max(len(left), len(right)); i++ {
		l, r := cell{}, cell{}
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		row := strings.Repeat(" ", 1-b2i(l.fresh)) + l.s
		if r.w > 0 {
			row += strings.Repeat(" ", richColW-l.w+min(T-58, 6)-b2i(r.fresh)) + r.s
		}
		out = append(out, row)
	}
	return out
}

// cardCut is what a card keeps when rows are short: how many THEN items,
// whether the cost line, and how many reason rows.
type cardCut struct {
	then, whyRows int
	cost          bool
	cmdRow        bool // a command too long for the chip row goes on a row of its own
}

// fitWords cuts s to room cells at a word boundary (a space or an
// underscore), with an ellipsis, and never inside a word.
func fitWords(s string, room int) string {
	r := []rune(s)
	if len(r) <= room || room < 2 {
		return fitText(s, room)
	}
	cut := room - 1
	if r[cut] != ' ' && r[cut] != '_' {
		for cut > 0 && r[cut-1] != ' ' && r[cut-1] != '_' {
			cut--
		}
		if cut == 0 {
			cut = room - 1
		}
	}
	return strings.TrimRight(string(r[:cut]), " ") + "…"
}

// cardRows is the card as rows, T cells wide, with a pad cell at the left.
func cardRows(s nextStep, T int, cut cardCut) []string {
	x := func(parts ...string) string { return " " + strings.Join(parts, "") }
	rows := []string{x(dimText(txtNow.String())), x(theme.Paint(theme.RoleAccent, "►"), " ", theme.Tag(theme.RoleBright)+fitText(s.title, T-2)+"[-]")}
	if cut.whyRows > 0 {
		why := wrapText(strings.Join(s.why, " "), T-2)
		for _, l := range why[:min(len(why), cut.whyRows)] {
			rows = append(rows, x("  ", bodyText(l)))
		}
	}
	if cut.cost && s.cost != "" {
		role := theme.RoleWarning
		if s.enough {
			role = theme.RolePositive
		}
		room := max(T-4, 1)
		lines := wrapItems(strings.Split(s.cost, ", "), room)
		var have []string
		if s.have != "" {
			have = wrapItems(strings.Split(s.have, ", "), room)
		}
		for i, l := range lines {
			row := x("  ", bodyText(l))
			if i == len(lines)-1 && len(have) == 1 && runeLen(l)+2+runeLen(have[0]) <= room {
				row += "  " + theme.Paint(role, have[0])
				have = nil
			}
			rows = append(rows, row)
		}
		for _, l := range have {
			rows = append(rows, x("  ", theme.Paint(role, l)))
		}
	}
	if s.cmd != "" {
		var below string
		btn := clickRegion(regionShow, theme.TagFgBgAttr(theme.RoleText, theme.RoleChip, "b")+txtShowMe.String()+theme.Reset)
		if T >= tabMinT {
			key := txtTabKey.String()
			if s.onLine {
				key = txtEnterKey.String()
			}
			btn += "   " + theme.TagFgBgAttr(theme.RoleOnAccent, theme.RoleAccent, "b") + key + theme.Reset + " "
			if room := T - runeLen(txtShowMe.String()) - runeLen(key) - 4; runeLen(s.cmd) > room && cut.cmdRow {
				below = x("  ", dimText(fitWords(s.cmd, T-2)))
			} else {
				btn += dimText(fitWords(s.cmd, room))
			}
		}
		rows = append(rows, "", x(btn))
		if below != "" {
			rows = append(rows, below)
		}
	}
	if cut.then > 0 {
		rows = append(rows, "", x(dimText(txtThen.String())))
		for i, it := range s.then[:cut.then] {
			if i > 0 {
				rows = append(rows, "")
			}
			title, word := fitText(it.title, T), it.panel
			line := bodyText(title)
			if word != "" && isPanel(word) && runeLen(title)+2+len(word) <= T {
				line += strings.Repeat(" ", T-runeLen(title)-len(word)) + dimText(word)
			}
			link := func(s string) string { return clickRegion(regionThen+string(rune('0'+i)), s) }
			rows = append(rows, link(x(line)), link(x("  ", dimText(fitText(it.detail, T-2)))))
		}
	}
	return rows
}

// nextBoxText fills the Next box, w by h cells inside. The card sits at the
// top and the rule row and list at the bottom, with a blank between. When
// rows are short the card gives way in a fixed order: the THEN items from the
// last, then the label, the list a form down, the cost, the reasons, and at
// last the card. The classic layout (rich false) has the list alone, in words.
func nextBoxText(s nextStep, w, h int, rich bool) string {
	T := w - 2
	if !rich {
		return strings.Join(panelRows(formWords, w-1, s.view), "\n")
	}
	rule := " " + theme.Paint(theme.RoleBorder, "─") + theme.Paint(theme.RoleAccent, txtPanels.String()) +
		theme.Paint(theme.RoleBorder, strings.Repeat("─", max(w-2-runeLen(txtPanels.String()), 0)))
	type step struct {
		cut  cardCut
		form int
	}
	var steps []step
	for _, f := range []int{formStack, formRich} {
		for t := len(s.then); t >= 0; t-- {
			steps = append(steps, step{cardCut{then: t, cost: true, whyRows: 2, cmdRow: true}, f}, step{cardCut{then: t, cost: true, whyRows: 2}, f})
		}
	}
	for _, f := range []int{formGrouped, formWords} {
		steps = append(steps, step{cardCut{cost: true, whyRows: 2}, f})
	}
	steps = append(steps, step{cardCut{whyRows: 2}, formWords}, step{cardCut{whyRows: 1}, formWords}, step{cardCut{}, formWords})
	if s.title != "" {
		for _, st := range steps {
			card, list := cardRows(s, T, st.cut), panelRows(st.form, T, s.view)
			if used := len(card) + 1 + panelRuleRows + len(list); used <= h {
				rows := append(append(append(card, ""), rule), list...)
				return strings.Join(append(rows, make([]string, h-used)...), "\n")
			}
		}
	}
	for f := formRich; f < formWords; f++ {
		if list := panelRows(f, T, s.view); len(list) <= h {
			return strings.Join(list, "\n")
		}
	}
	return strings.Join(panelRows(formWords, T, s.view), "\n")
}
