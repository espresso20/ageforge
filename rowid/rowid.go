// Package rowid carries the row id the copy export gives each string the game
// writes itself, so the export and the import find the same text by the same
// id. The wrappers return the plain string.
//
// rowid imports nothing of the game, so any package may use it.
package rowid

import (
	"fmt"
)

// TEMPORARY (removed in the switch PR): callers of the removed switch still compile.
func On() bool                   { return false }
func Set(bool)                   {}
func Text(string) (string, bool) { return "", false }

// Line is the string the code writes itself. id is the row id the copy export
// gives it, and plain the string.
func Line(id, plain string) string { return plain }

// Sprintf is fmt.Sprintf for a format the code writes itself (a log line, a
// refusal). id is the row id the copy export gives the format.
//
// It is a wrapper vet reads as it reads fmt.Sprintf, so a format and its
// arguments that do not agree are still caught at the call.
func Sprintf(id, format string, a ...any) string { return fmt.Sprintf(format, a...) }

// Errorf is fmt.Errorf for a refusal or an error the code writes itself, on the
// same terms as Sprintf.
func Errorf(id, format string, a ...any) error { return fmt.Errorf(format, a...) }

// Entry is a string of a table the code makes once (a list of key hints, the
// headings of a panel): it carries its row id and its plain text.
//
//	var panelKeys = []rowid.Entry{rowid.E("ui.panel.panelKeys.01", "Close the panel")}
//	... panelKeys[0].String()
type Entry struct {
	id, plain string
}

// E is the text of a table entry: the row id and the plain string.
func E(id, plain string) Entry { return Entry{id: id, plain: plain} }

// Same is the entry of a string that has no row, to stand in a table beside
// entries that do (a mark that is no word: "NVT", "24H").
func Same(s string) Entry { return Entry{plain: s} }

// String is the entry's text.
func (t Entry) String() string { return t.plain }

// Plain is the text in the game's own English.
func (t Entry) Plain() string { return t.plain }
