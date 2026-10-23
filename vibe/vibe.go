// Package vibe is VibeMode: a switch that turns the game's words into a
// second script.
//
// The script is the sheets in words/vibe, a line per row id (the ids the
// copy export gives every line of the game). The switch is one flag for the
// whole process, set from the account's setting. Two things read the script:
//
//   - the ruleset overlay (package rules) swaps the text of every definition
//     that has a row, when the ruleset is built, and
//   - Line swaps one string of code in place, for a line the code writes
//     itself (a log line, a title).
//
// A line with no row in the script reads as it always did, and nothing but
// words ever changes: no number, no key, no save.
//
// vibe imports nothing of the game, so any package may read it.
package vibe

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/espresso20/ageforge/words"
)

var on atomic.Bool

// On reports whether VibeMode is on.
func On() bool { return on.Load() }

// Set turns VibeMode on or off for the process. Whoever calls it with a
// changed value owns reloading the text that was built under the old one
// (the engine rebinds its ruleset; the dashboard redraws).
func Set(v bool) { on.Store(v) }

// table is the script, read once from the embedded sheets.
type table struct {
	text     map[string]string
	ids      []string
	problems []string
}

// embedded is the script in the binary, read once.
var embedded = sync.OnceValue(func() *table {
	rows, problems := words.VibeRows()
	t := &table{text: make(map[string]string, len(rows)), problems: problems}
	for _, r := range rows {
		t.text[r.ID] = r.Text
		t.ids = append(t.ids, r.ID)
	}
	sort.Strings(t.ids)
	return t
})

// loaned is a script a test put in place of the embedded one.
var loaned atomic.Pointer[table]

// script is the script in force: the embedded one, or a test's.
func script() *table {
	if t := loaned.Load(); t != nil {
		return t
	}
	return embedded()
}

// UseScriptForTest puts a script of the given rows in place of the embedded
// one until the returned function is called. It is for tests of the machinery,
// which must not depend on the lines the game ships with.
func UseScriptForTest(rows map[string]string) (restore func()) {
	t := &table{text: make(map[string]string, len(rows))}
	for id, s := range rows {
		t.text[id] = s
		t.ids = append(t.ids, id)
	}
	sort.Strings(t.ids)
	prev := loaned.Swap(t)
	return func() { loaned.Store(prev) }
}

// Text is the script's line for a row id, whether or not VibeMode is on. It
// reports false for a row the script has no line for.
func Text(id string) (string, bool) {
	s, ok := script().text[id]
	return s, ok
}

// Line is the text for a string the code writes itself: the script's line
// for the row id while VibeMode is on and the script has one, else plain.
// id is the row id the copy export gives the string, and plain the string.
func Line(id, plain string) string {
	if !on.Load() {
		return plain
	}
	if s, ok := script().text[id]; ok {
		return s
	}
	return plain
}

// IDs lists the row ids the script has a line for, in order.
func IDs() []string { return append([]string(nil), script().ids...) }

// Problems lists what was wrong with the embedded sheets when they were
// read: a sheet that would not parse, a row id given twice. A build with any
// is caught by the tests; the game carries on with the rows that were fine.
func Problems() []string { return append([]string(nil), script().problems...) }
