package ui

import "github.com/espresso20/ageforge/game"

// log_routing.go decides which log shows an entry. Routine confirmations (a
// build started, workers assigned, a plan item added: lines whose effect a
// panel already shows) go only to the logs panel, so the main window's log
// keeps notable events, warnings and errors. game.LogRoutine documents the
// rule for choosing a category.

// mainLogShows reports whether the main window's log shows e: everything but
// routine confirmations and debug lines.
func mainLogShows(e game.LogEntry) bool {
	return e.Type != game.LogRoutine && e.Type != "debug"
}

// logsPanelShows reports whether the logs panel shows e: everything but
// debug lines, which only dumps carry.
func logsPanelShows(e game.LogEntry) bool {
	return e.Type != "debug"
}
