package main

import "strings"

// vibe_map.go: what vibe-wire decides about the map sheet that the syntax
// alone does not say. The rest of the sheet is decided like the screens: a text
// in a function is wired as a vibe call, and a text in a table made once becomes
// an entry that carries its id (vibe.E), or vibe.Same where the table holds a
// string that has no row (see vibe/vibe.go).

const (
	leftNeverShown = "never shown: a name in a table that no screen draws (the code and its tests find a row by it)"
	leftArt        = "art: a row of a drawing (a mask of palette letters or a pattern of marks), not words"
)

// neverShown are the tables of the map whose names nothing draws.
var neverShown = []string{
	"ui.mapstyle.skyline.roster.rosterEntries.", // the traffic roster: the name only labels an entry for the tests
}

func init() {
	leftAloneFor(leftArt,
		"ui.mapstyle.skyline.roster.tpCar.01", "ui.mapstyle.skyline.roster.tpHoverCar.01", // the mask of a car: aa is a palette letter, twice
		"ui.mapstyle.skyline.roster.trSteamEngine.01", // #OO#: the pattern of a smoking engine
	)
}

// mapReason says why a row of the map sheet is not wired, when its place says
// so by itself.
func mapReason(r *row) string {
	if r.area != areaMap {
		return ""
	}
	for _, p := range neverShown {
		if strings.HasPrefix(r.id, p) {
			return leftNeverShown
		}
	}
	return ""
}
