package game

import "strings"

// log_tags.go: what a log line is about.
//
// The log colours a loss red and a gain green. It finds the lines by the words
// they have ("lost", "recruited"): a panel asks the line what it is about
// (LogEntry.About). Nothing is stored with a line.

// LogTag says what a log line is about.
type LogTag uint8

const (
	TagBad  LogTag = 1 << iota // bad news: something or someone lost
	TagGood                    // good news: someone arrived or came home
)

// logWords are the words that mark a line written in plain English as being
// about a subject.
var logWords = map[LogTag][]string{
	TagBad:  {"starv", "died", "fell", "lost"},
	TagGood: {"moved in", "recruited", "joined", "home"},
}

// About reports whether the line is about subject (TagBad or TagGood), by the words the line has.
func (e LogEntry) About(subject LogTag) bool {
	m := strings.ToLower(e.Message)
	for _, w := range logWords[subject] {
		if strings.Contains(m, w) {
			return true
		}
	}
	return false
}
