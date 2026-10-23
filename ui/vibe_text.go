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
