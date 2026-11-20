package ui

import "github.com/espresso20/ageforge/rowid"

// panels_table.go is the Panels list as the rich layout shows it
// (dash_layout.go): the panels in groups by what they are for, each with a
// line on its job. The table below is the one place that wording lives.
// TestPanelsTableCoversEveryPanel holds it to the panels' own names, so a
// panel has to be given a group and a line here.

// panelInfo is one panel: the command that opens it and what it is for.
type panelInfo struct {
	name string
	job  rowid.Entry
}

// panelGroup is a heading and the panels under it.
type panelGroup struct {
	title  rowid.Entry
	panels []panelInfo
}

// panelGroups is the Panels list, grouped. The job lines are placeholders
// for the copy the game's words get; keep each under panelJobMax cells.
var panelGroups = []panelGroup{
	{rowid.E("ui.panels_table.panelGroups.01", "Town"), []panelInfo{
		{"people", rowid.E("ui.panels_table.panelGroups.02", "workers and army")},
		{"plan", rowid.E("ui.panels_table.panelGroups.03", "queue builds ahead")},
		{"wonders", rowid.E("ui.panels_table.panelGroups.04", "the great works")},
		{"map", rowid.E("ui.panels_table.panelGroups.05", "your town, drawn")},
	}},
	{rowid.E("ui.panels_table.panelGroups.06", "Progress"), []panelInfo{
		{"milestones", rowid.E("ui.panels_table.panelGroups.07", "goals and rewards")},
		{"research", rowid.E("ui.panels_table.panelGroups.08", "the tech tree")},
		{"charter", rowid.E("ui.panels_table.panelGroups.09", "what advance needs")},
		{"epoch", rowid.E("ui.panels_table.panelGroups.10", "eras and disasters")},
		{"history", rowid.E("ui.panels_table.panelGroups.11", "your story so far")},
	}},
	{rowid.E("ui.panels_table.panelGroups.12", "World"), []panelInfo{
		{"expedition", rowid.E("ui.panels_table.panelGroups.13", "send parties out")},
		{"army", rowid.E("ui.panels_table.panelGroups.14", "soldiers and ranks")},
		{"trade", rowid.E("ui.panels_table.panelGroups.15", "routes and goods")},
		{"factions", rowid.E("ui.panels_table.panelGroups.16", "your neighbors")},
		{"harbinger", rowid.E("ui.panels_table.panelGroups.17", "the coming warning")},
	}},
	{rowid.E("ui.panels_table.panelGroups.18", "You"), []panelInfo{
		{"badges", rowid.E("ui.panels_table.panelGroups.19", "your trophies")},
		{"stats", rowid.E("ui.panels_table.panelGroups.20", "numbers and rates")},
		{"logs", rowid.E("ui.panels_table.panelGroups.21", "the full record")},
		{"help", rowid.E("ui.panels_table.panelGroups.22", "how to play")},
	}},
}

const (
	// panelJobMax is the longest job line, in cells.
	panelJobMax = 18
	// panelNameW is the width the panel names are set in (the longest is
	// "milestones", "expedition" and so on at 10).
	panelNameW = 10
)
