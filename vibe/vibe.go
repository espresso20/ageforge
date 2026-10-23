// Package vibe is VibeMode: a switch that turns the game's words into a
// second script.
//
// The script is the sheets in words/vibe, a line per row id (the ids the
// copy export gives every line of the game). The switch is one flag for the
// whole process, set from the account's setting. Three things read the script:
//
//   - the ruleset overlay (package rules) swaps the text of every definition
//     that has a row, when the ruleset is built,
//   - Line swaps one string of code in place, for a line the code writes
//     itself (a title, a label, a word), and Sprintf and Errorf do the same
//     for a format (a log line, a refusal), before the arguments go in, and
//   - E is Line for an entry of a table the code makes once: the entry keeps
//     the id and reads when it is used.
//
// Each of them takes the row id the copy export gives the string, and the
// string. The id is part of the source, so it stays with the string when
// the code around it moves (cmd/words vibe-wire writes it, vibe-lint checks
// it).
//
// A line with no row in the script reads as it always did, and nothing but
// words ever changes: no number, no key, no save.
//
// vibe imports nothing of the game, so any package may read it.
package vibe

import (
	"fmt"
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

// rewriter is a function a test put in place of the script for the texts the
// code writes itself.
var rewriter atomic.Pointer[func(id, plain string) string]

// UseRewriterForTest makes every text the code writes itself (Line, Sprintf,
// Errorf) read as f(id, plain) while VibeMode is on, whatever the script
// holds, until the returned function is called. It is for tests that must
// change every wired text at once, without knowing which they are.
func UseRewriterForTest(f func(id, plain string) string) (restore func()) {
	prev := rewriter.Swap(&f)
	return func() { rewriter.Store(prev) }
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
	return lineFor(id, plain)
}

// Sprintf is fmt.Sprintf for a format the code writes itself (a log line, a
// refusal): the script's line for the row id, if VibeMode is on and the script
// has one, else format, and then the arguments are put in. id is the row id
// the copy export gives the format.
//
// It is a wrapper vet reads as it reads fmt.Sprintf, so a format and its
// arguments that do not agree are still caught at the call. The script's line
// must carry the same verbs in the same order (vibe-lint holds it to that).
func Sprintf(id, format string, a ...any) string {
	if !on.Load() {
		return fmt.Sprintf(format, a...)
	}
	return fmt.Sprintf(lineFor(id, format), a...)
}

// Errorf is fmt.Errorf for a refusal or an error the code writes itself, on
// the same terms as Sprintf. The script's line may carry %w where the format
// does, and wraps in the same place.
func Errorf(id, format string, a ...any) error {
	if !on.Load() {
		return fmt.Errorf(format, a...)
	}
	return fmt.Errorf(lineFor(id, format), a...)
}

// Entry is a string of a table the code makes once (a list of key hints, the
// headings of a panel): it carries its row id and its plain text, and reads
// as Line would when it is used, not when the table was made. So a table at
// the top of a file can follow the switch, with nothing to rebuild.
//
//	var panelKeys = []vibe.Entry{vibe.E("ui.panel.panelKeys.01", "Close the panel")}
//	... panelKeys[0].String()
//
// Like Line, it takes the row id the copy export gives the string.
type Entry struct {
	id, plain string
}

// E is the text of a table entry: the row id and the plain string.
func E(id, plain string) Entry { return Entry{id: id, plain: plain} }

// String is the text as Line would give it now: the script's line for the row
// while VibeMode is on and the script has one, else plain. Being a Stringer,
// an Entry prints as its line with %s and %v.
func (t Entry) String() string { return Line(t.id, t.plain) }

// Plain is the text in the game's own English, whatever the switch says. It is
// for the few places that must match or measure the English.
func (t Entry) Plain() string { return t.plain }

// lineFor is the script's line for a row id, else plain. VibeMode is on.
func lineFor(id, plain string) string {
	if f := rewriter.Load(); f != nil {
		return (*f)(id, plain)
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
