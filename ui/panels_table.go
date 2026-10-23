package ui

import (
	"strings"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// panels_table.go is the Panels list as the rich layout shows it
// (dash_layout.go): the panels in groups by what they are for, each with a
// line on its job. The table below is the one place that wording lives.
// TestPanelsTableCoversEveryPanel holds it to sidebarPanels, so a panel
// added to the list has to be given a group and a line here.

// panelInfo is one panel: the command that opens it and what it is for.
type panelInfo struct {
	name string
	job  vibe.Entry
}

// panelGroup is a heading and the panels under it.
type panelGroup struct {
	title  vibe.Entry
	panels []panelInfo
}

// panelGroups is the Panels list, grouped. The job lines are placeholders
// for the copy the game's words get; keep each under panelJobMax cells.
var panelGroups = []panelGroup{
	{vibe.E("ui.panels_table.panelGroups.01", "Town"), []panelInfo{
		{"people", vibe.E("ui.panels_table.panelGroups.02", "workers and army")},
		{"plan", vibe.E("ui.panels_table.panelGroups.03", "queue builds ahead")},
		{"wonders", vibe.E("ui.panels_table.panelGroups.04", "the great works")},
		{"map", vibe.E("ui.panels_table.panelGroups.05", "your town, drawn")},
	}},
	{vibe.E("ui.panels_table.panelGroups.06", "Progress"), []panelInfo{
		{"milestones", vibe.E("ui.panels_table.panelGroups.07", "goals and rewards")},
		{"research", vibe.E("ui.panels_table.panelGroups.08", "the tech tree")},
		{"charter", vibe.E("ui.panels_table.panelGroups.09", "what advance needs")},
		{"epoch", vibe.E("ui.panels_table.panelGroups.10", "eras and disasters")},
		{"history", vibe.E("ui.panels_table.panelGroups.11", "your story so far")},
	}},
	{vibe.E("ui.panels_table.panelGroups.12", "World"), []panelInfo{
		{"expedition", vibe.E("ui.panels_table.panelGroups.13", "send parties out")},
		{"army", vibe.E("ui.panels_table.panelGroups.14", "soldiers and ranks")},
		{"trade", vibe.E("ui.panels_table.panelGroups.15", "routes and goods")},
		{"factions", vibe.E("ui.panels_table.panelGroups.16", "your neighbors")},
		{"harbinger", vibe.E("ui.panels_table.panelGroups.17", "the coming warning")},
	}},
	{vibe.E("ui.panels_table.panelGroups.18", "You"), []panelInfo{
		{"badges", vibe.E("ui.panels_table.panelGroups.19", "your trophies")},
		{"stats", vibe.E("ui.panels_table.panelGroups.20", "numbers and rates")},
		{"logs", vibe.E("ui.panels_table.panelGroups.21", "the full record")},
		{"help", vibe.E("ui.panels_table.panelGroups.22", "how to play")},
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
		sb.WriteString(" " + theme.Paint(theme.RoleDim, strings.ToUpper(g.title.String())) + "\n")
		for _, p := range g.panels {
			name := p.name + strings.Repeat(" ", max(0, panelNameW-len(p.name)))
			if p.name == active {
				sb.WriteString(theme.Selected(" "+name) + " ")
			} else {
				sb.WriteString(" [white]" + name + "[-] ")
			}
			sb.WriteString(theme.Paint(theme.RoleDim, fitText(p.job.String(), max(0, w-1-panelNameW-1))) + "\n")
		}
	}
	return sb.String()
}
