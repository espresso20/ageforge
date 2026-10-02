package ui

import "github.com/espresso20/ageforge/game"

// log_routing.go decides which log shows an entry. Both logs show every line
// a player can see, routine confirmations (a build started, workers
// assigned, a plan item added) included, so the main window always answers
// the command just typed. They differ in how a line reads: the main window's
// log (Dashboard.refreshLog) shows the message alone, in its category's
// color; the logs panel (logsProvider) puts the tick number and a category
// tag in front of every line and marks routine confirmations with a dot.
// game.LogRoutine documents the rule for choosing a category.

// mainLogShows reports whether the main window's log shows e: everything but
// debug lines, which only dumps carry.
func mainLogShows(e game.LogEntry) bool {
	return e.Type != "debug"
}

// logsPanelShows reports whether the logs panel shows e: everything but
// debug lines, which only dumps carry.
func logsPanelShows(e game.LogEntry) bool {
	return e.Type != "debug"
}
