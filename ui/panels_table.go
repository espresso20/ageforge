package ui

import (
	"strings"

	"github.com/espresso20/ageforge/theme"
)

// panels_table.go is the Panels list as the rich layout shows it
// (dash_layout.go): the panels in groups by what they are for, each with a
// line on its job. The table below is the one place that wording lives.
// TestPanelsTableCoversEveryPanel holds it to sidebarPanels, so a panel
// added to the list has to be given a group and a line here.

// panelInfo is one panel: the command that opens it and what it is for.
type panelInfo struct{ name, job string }

// panelGroup is a heading and the panels under it.
type panelGroup struct {
	title  string
	panels []panelInfo
}

// panelGroups is the Panels list, grouped. The job lines are placeholders
// for the copy the game's words get; keep each under panelJobMax cells.
var panelGroups = []panelGroup{
	{"Town", []panelInfo{
		{"people", "workers and army"},
		{"plan", "queue builds ahead"},
		{"wonders", "the great works"},
		{"map", "your town, drawn"},
	}},
	{"Progress", []panelInfo{
		{"milestones", "goals and rewards"},
		{"research", "the tech tree"},
		{"charter", "what advance needs"},
		{"epoch", "eras and disasters"},
		{"history", "your story so far"},
	}},
	{"World", []panelInfo{
		{"expedition", "send parties out"},
		{"army", "soldiers and ranks"},
		{"trade", "routes and goods"},
		{"factions", "your neighbors"},
		{"harbinger", "the coming warning"},
	}},
	{"You", []panelInfo{
		{"badges", "your trophies"},
		{"stats", "numbers and rates"},
		{"logs", "the full record"},
		{"help", "how to play"},
	}},
}

const (
	// panelJobMax is the longest job line, in cells.
	panelJobMax = 18
	// panelNameW is the width the panel names are set in (the longest is
	// "milestones", "expedition" and so on at 10).
	panelNameW = 10
)

// panelsRichWidth is the inner width the rich list needs: a space, the name
// column, a space and the longest job line.
func panelsRichWidth() int { return 1 + panelNameW + 1 + panelJobMax }

// panelsRichRows is the rows the rich list needs inside the box: a heading
// for each group and a row for each panel.
func panelsRichRows() int {
	n := 0
	for _, g := range panelGroups {
		n += 1 + len(g.panels)
	}
	return n
}

// panelsRichText is the rich Panels list for a box w cells wide and h rows
// tall inside (w at least panelsRichWidth and h at least panelsRichRows):
// a dim heading for each group, then a row for each panel, its name and
// what it is for in the dim role, the open panel marked in its name. Rows
// to spare go to blank lines between the groups, the last pair first.
func panelsRichText(active string, w, h int) string {
	spare := h - panelsRichRows()
	var sb strings.Builder
	for i, g := range panelGroups {
		if i > 0 && spare >= len(panelGroups)-i {
			sb.WriteString("\n")
			spare--
		}
		sb.WriteString(" " + theme.Paint(theme.RoleDim, strings.ToUpper(g.title)) + "\n")
		for _, p := range g.panels {
			name := p.name + strings.Repeat(" ", max(0, panelNameW-len(p.name)))
			if p.name == active {
				sb.WriteString(theme.Selected(" "+name) + " ")
			} else {
				sb.WriteString(" [white]" + name + "[-] ")
			}
			sb.WriteString(theme.Paint(theme.RoleDim, fitText(p.job, max(0, w-1-panelNameW-1))) + "\n")
		}
	}
	return sb.String()
}
