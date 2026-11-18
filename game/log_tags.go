package game

import "strings"

// log_tags.go: what a log line is about.
//
// The People and Army tabs list the log lines about their subject, and the
// log colours a loss red and a gain green. They used to find those lines by
// reading their words ("recruited", "lost", "soldier"). A line now carries tags, set where it is written
// (addLogAbout), and a panel asks the line what it is about (LogEntry.About).

// LogTag says what a log line is about.
type LogTag uint8

const (
	TagPeople LogTag = 1 << iota // about the people of the town: recruited, assigned, moved in, starving, scholars, soldiers
	TagArmy                      // about the army: soldiers, campaigns, raids, the garrison
	TagBad                       // bad news: something or someone lost
	TagGood                      // good news: someone arrived or came home
)

// logWords are the words that mark a line written in plain English as being
// about a subject.
var logWords = map[LogTag][]string{
	TagPeople: {"roster:", "recruited", "moved in", "starved", "worker", "villager", "population",
		"scholar", "soldier", "reserve", "homes", "campaign", "expedition returned", "came home"},
	TagArmy: {"soldier", "campaign", "garrison", "defense", "raid", "army"},
	TagBad:  {"starv", "died", "fell", "lost"},
	TagGood: {"moved in", "recruited", "joined", "home"},
}

// About reports whether the line is about subject (one of TagPeople, TagArmy,
// TagBad, TagGood), by the words the line has.
func (e LogEntry) About(subject LogTag) bool {
	m := strings.ToLower(e.Message)
	for _, w := range logWords[subject] {
		if strings.Contains(m, w) {
			return true
		}
	}
	return false
}
