package ui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// lockedMarker opens the name line of a building a tech still locks: the
// game's lock, as its locked themes wear it.
const lockedMarker = "🔒"

// lockedHintLines are the words a locked building's entry says in place of
// its flavor. A building always gets the same one (lockedHintLine).
var lockedHintLines = []string{
	"Locked. Your scholars are still arguing about it.",
	"Locked. Someone has to read a book first. Not you, obviously.",
	"Locked. The blueprints exist. Nobody can read them yet.",
	"Locked. Research it, or keep staring. Staring has a poor track record.",
	"Locked. Your scholars need a nudge. A firm, encouraging nudge.",
	"Locked. Theory first, bricks later.",
	"Locked. The idea is out there. Send a scholar to fetch it.",
	"Locked. Needs knowledge, which needs scholars, which need snacks.",
	"Locked. Your builders asked how. Nobody had an answer.",
	"Locked. Crack a book. Or crack the whip at the library.",
	"Locked. Science has not happened to this one yet.",
	"Locked. It is on the syllabus. Your town skipped class.",
}

// lockedHintLine is the word for building key: the same every time it is
// drawn, and not the same for every building.
func lockedHintLine(key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	return lockedHintLines[h.Sum32()%uint32(len(lockedHintLines))]
}

// lockedNeedsLine says what a locked building waits for: the tech, and the
// research still to come before it, nearest first.
func lockedNeedsLine(bs game.BuildingState) string {
	if len(bs.LockTechPath) == 0 {
		return fmt.Sprintf("Needs %s (research)", bs.LockTech)
	}
	// LockTechPath is nearest first; the line reads in research order, so
	// the first tech named is the one to research first.
	order := make([]string, len(bs.LockTechPath))
	for i, name := range bs.LockTechPath {
		order[len(order)-1-i] = name
	}
	return fmt.Sprintf("Needs %s (research), after %s", bs.LockTech, strings.Join(order, ", then "))
}

// lockedBuildingEntry is a building its age has unlocked and a tech still
// locks, as the Economy tab's Buildings box lists it: greyed out, the
// research it needs, its cost and what it does, and a line from the
// scholars in place of its flavor. Every line is broken to w cells.
func lockedBuildingEntry(state game.GameState, key string, bs game.BuildingState, w int) string {
	var sb strings.Builder
	count := fmt.Sprintf("x%d", bs.Count)
	name := truncate(bs.Name, max(4, w-(1+2+1+1+runeLen(count))))
	fmt.Fprintf(&sb, " [gray]%s %s %s[-]\n", lockedMarker, name, count)
	sb.WriteString(hangingRow("   ", "[gray]"+lockedNeedsLine(bs)+"[-]", w))
	sb.WriteString("[gray]" + costRow("   Cost: ", buildingPrice(&state, key, bs), w) + "[-]")
	sb.WriteString(hangingRow("   ", "[gray]"+glueRates(buildingText(bs))+"[-]", w))
	sb.WriteString(hangingRow("   ", "[gray::i]"+lockedHintLine(key)+"[-:-:-]", w))
	return sb.String()
}
