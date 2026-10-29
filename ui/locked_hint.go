package ui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// lockedMarker opens the name line of a building a tech still locks: the
// game's lock, as its locked themes wear it.
const lockedMarker = "🔒"

// lockedHintLines are the words a locked building's entry says in place of
// its flavor. A building always gets the same one (lockedHintLine).
var lockedHintLines = []vibe.Entry{
	vibe.E("ui.locked_hint.lockedHintLines.01", "Locked. Your scholars are still arguing about it."),
	vibe.E("ui.locked_hint.lockedHintLines.02", "Locked. Someone has to read a book first. Not you, obviously."),
	vibe.E("ui.locked_hint.lockedHintLines.03", "Locked. The blueprints exist. Nobody can read them yet."),
	vibe.E("ui.locked_hint.lockedHintLines.04", "Locked. Research it, or keep staring. Staring has a poor track record."),
	vibe.E("ui.locked_hint.lockedHintLines.05", "Locked. Your scholars need a nudge. A firm, encouraging nudge."),
	vibe.E("ui.locked_hint.lockedHintLines.06", "Locked. Theory first, bricks later."),
	vibe.E("ui.locked_hint.lockedHintLines.07", "Locked. The idea is out there. Send a scholar to fetch it."),
	vibe.E("ui.locked_hint.lockedHintLines.08", "Locked. Needs knowledge, which needs scholars, which need snacks."),
	vibe.E("ui.locked_hint.lockedHintLines.09", "Locked. Your builders asked how. Nobody had an answer."),
	vibe.E("ui.locked_hint.lockedHintLines.10", "Locked. Crack a book. Or crack the whip at the library."),
	vibe.E("ui.locked_hint.lockedHintLines.11", "Locked. Science has not happened to this one yet."),
	vibe.E("ui.locked_hint.lockedHintLines.12", "Locked. It is on the syllabus. Your town skipped class."),
}

// lockedHintLine is the word for building key: the same every time it is
// drawn, and not the same for every building.
func lockedHintLine(key string) string {
	h := fnv.New32a()
	h.Write([]byte(key))
	return lockedHintLines[h.Sum32()%uint32(len(lockedHintLines))].String()
}

// joinNames joins up to 3 names with commas and "or"/"and": "A", "A or B", "A, B or C".
// If there are more than 3, it returns the first 3 joined and the count of remaining.
func joinNames(names []string) (string, int) {
	if len(names) == 0 {
		return "", 0
	}
	if len(names) == 1 {
		return names[0], 0
	}
	if len(names) == 2 {
		return names[0] + " or " + names[1], 0
	}
	if len(names) == 3 {
		return names[0] + ", " + names[1] + " or " + names[2], 0
	}
	// More than 3: show the first 3
	return names[0] + ", " + names[1] + " or " + names[2], len(names) - 3
}

// lockedNeedsLine says what a locked building waits for: the tech, and the
// research still to come before it.
func lockedNeedsLine(bs game.BuildingState) string {
	if len(bs.LockTechPath) == 0 {
		return vibe.Sprintf("ui.locked_hint.lockedNeedsLine.01", "Needs %s (research)", bs.LockTech)
	}
	if len(bs.LockTechPath) == 1 {
		return vibe.Sprintf("ui.locked_hint.lockedNeedsLine.02", "Needs %s (research), after %s", bs.LockTech, bs.LockTechPath[0])
	}
	// Path has 2 or more techs. Tell the player what they can start with.
	if len(bs.LockTechNext) > 0 {
		names, more := joinNames(bs.LockTechNext)
		line := vibe.Sprintf("ui.locked_hint.lockedNeedsLine.03", "Needs %s (research), %d techs before it. Start with %s", bs.LockTech, len(bs.LockTechPath), names)
		if more > 0 {
			line += vibe.Sprintf("ui.locked_hint.lockedNeedsLine.04", " and %d more", more)
		}
		return line
	}
	// No next techs (shouldn't happen, but handle it)
	return vibe.Sprintf("ui.locked_hint.lockedNeedsLine.05", "Needs %s (research), %d techs before it", bs.LockTech, len(bs.LockTechPath))
}

// txtIndentedCost is said the same way on several screens, so it has one id.
var txtIndentedCost = vibe.E("ui.locked_hint.lockedBuildingEntry.01", "   Cost: ")

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
	sb.WriteString("[gray]" + costRow(txtIndentedCost.String(), buildingPrice(&state, key, bs), w) + "[-]")
	sb.WriteString(hangingRow("   ", "[gray]"+glueRates(buildingText(bs))+"[-]", w))
	sb.WriteString(hangingRow("   ", "[gray::i]"+lockedHintLine(key)+"[-:-:-]", w))
	return sb.String()
}
