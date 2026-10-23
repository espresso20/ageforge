package mapmodel

import "github.com/espresso20/ageforge/vibe"

// vibe_text.go: the words of the map's tables, and VibeMode.
//
// The rosters (the city's features, the movers, the looks of the Earth arc)
// are tables made once, whose entries carry their row ids (vibe.E). Each
// roster's info struct holds the English a table was written in (Name,
// Title, Lines), which is what every reader gets while VibeMode is off: the
// copy is a plain value, as it always was. While VibeMode is on, the accessor
// that hands a struct out (Info, CityLookAt) first reads the entries again,
// so a switch needs no restart and no rebuild of the tables.

// plainLines is the English of a table's lines.
func plainLines(es []vibe.Entry) []string {
	if es == nil {
		return nil
	}
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Plain()
	}
	return out
}

// toldLines is a table's lines as they read now.
func toldLines(es []vibe.Entry) []string {
	if es == nil {
		return nil
	}
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.String()
	}
	return out
}
