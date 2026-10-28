// Package wordlist reads the plain-text files that hold the game's authored
// word lists and sentences.
//
// A file is a series of lists. A line that starts with "==" opens a list, and
// every other line is one entry of it. Blank lines are skipped, and so is a
// line that starts with "#", which is a comment:
//
//	# A comment.
//	== name setting setting
//	first entry
//	second entry
//
// White space around a line is not part of the entry. The package knows
// nothing about what the entries mean: the game's packages read their own
// meaning into the entries and the heading, and the copy tool uses the
// recorded positions to put a rewritten entry back in the file.
package wordlist

import (
	"fmt"
	"strings"
)

const headMark = "=="

// Entry is one line of a list.
type Entry struct {
	Line  int    // the line number in the file, from 1
	Start int    // byte offset of the first character of Text in the file
	End   int    // byte offset just after the last character of Text
	Text  string // the line, without the white space around it
}

// List is one "==" heading and the entries under it.
type List struct {
	Head    []string // the words of the heading after "=="; the first is the list's name
	Line    int      // the line number of the heading
	Entries []Entry
}

// Name is the first word of the heading.
func (l List) Name() string { return l.Head[0] }

// Parse reads the lists of one file. It reports the first thing that is not a
// list: an entry before any heading, a heading with no name, two lists with
// the same heading.
func Parse(text string) ([]List, error) {
	var lists []List
	seen := map[string]bool{}
	offset := 0
	for index, raw := range strings.SplitAfter(text, "\n") {
		lineStart := offset
		offset += len(raw)
		line := strings.TrimSpace(raw)
		lineNumber := index + 1
		switch {
		case line == "" || line[0] == '#':
			continue
		case strings.HasPrefix(line, headMark):
			head := strings.Fields(strings.TrimPrefix(line, headMark))
			if len(head) == 0 {
				return nil, fmt.Errorf("line %d: a heading with no name", lineNumber)
			}
			key := strings.Join(head, " ")
			if seen[key] {
				return nil, fmt.Errorf("line %d: the list %q is already in this file", lineNumber, key)
			}
			seen[key] = true
			lists = append(lists, List{Head: head, Line: lineNumber})
		case len(lists) == 0:
			return nil, fmt.Errorf("line %d: an entry before the first heading", lineNumber)
		default:
			start := lineStart + strings.Index(raw, line)
			last := &lists[len(lists)-1]
			last.Entries = append(last.Entries, Entry{Line: lineNumber, Start: start, End: start + len(line), Text: line})
		}
	}
	return lists, nil
}
