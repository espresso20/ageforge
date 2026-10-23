package ui

import "github.com/espresso20/ageforge/vibe"

// vibe_text.go: the tables of text the screens make once. A table at the top of
// a file cannot call vibe.Line when it is made, so each of its texts is a
// vibe.Entry that carries its row id and reads when it is used. These are the
// few helpers for using them.

// said is the strings a list of entries reads as now.
func said(entries []vibe.Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.String()
	}
	return out
}

// keyHelp is a key and what it does, as the Help panel lists a panel's keys.
// The key is the name of a key on the keyboard, and stays as it is.
type keyHelp struct {
	key  string
	what vibe.Entry
}

// helpPairs is a table of keyHelp as the pairs the Help panel lays out.
func helpPairs(table []keyHelp) [][2]string {
	out := make([][2]string, len(table))
	for i, k := range table {
		out[i] = [2]string{k.key, k.what.String()}
	}
	return out
}
