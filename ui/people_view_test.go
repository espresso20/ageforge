package ui

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestThePeopleTabsPickLinesByWhatTheyAreAbout: a line written in plain
// English is picked by its words, as it always was; a line written under
// VibeMode's script is picked by the tags it was written with, so the script's
// words do not move it out of a panel (or into one). The colour of a line
// follows the same rule.
func TestThePeopleTabsPickLinesByWhatTheyAreAbout(t *testing.T) {
	st := game.GameState{Log: []game.LogEntry{
		{Message: "Recruited 3 workers (population 4/9).", Type: game.LogRoutine, Tags: game.TagPlain},
		{Message: "Built a Hut (you have 4).", Type: game.LogRoutine, Tags: game.TagPlain | game.TagPeople}, // tags are not read for a plain line
		{Message: "Squad of 3 pulled up, no cap.", Type: game.LogRoutine, Tags: game.TagPeople | game.TagGood},
		{Message: "The soldiers lost the plot.", Type: "info"}, // under a script, with no tag: words do not count
		{Message: "Two of the squad took an L.", Type: "warning", Tags: game.TagPeople | game.TagArmy | game.TagBad},
		{Message: "A debug line about workers.", Type: "debug", Tags: game.TagPeople},
	}}
	messages := func(lines []game.LogEntry) []string {
		var out []string
		for _, l := range lines {
			out = append(out, l.Message)
		}
		return out
	}
	if got, want := messages(recentLog(st, peopleEvent, 10)), []string{
		"Two of the squad took an L.", "Squad of 3 pulled up, no cap.", "Recruited 3 workers (population 4/9).",
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("the People tab picks %q, want %q", got, want)
	}
	if got, want := messages(recentLog(st, armyEvent, 10)), []string{"Two of the squad took an L."}; !reflect.DeepEqual(got, want) {
		t.Errorf("the Army tab picks %q, want %q", got, want)
	}
	if got, want := messages(recentLog(st, peopleEvent, 1)), []string{"Two of the squad took an L."}; !reflect.DeepEqual(got, want) {
		t.Errorf("the newest line picked is %q, want %q", got, want)
	}
	for i, want := range []xk{xGood, xText, xGood, xText, xBad, xText} {
		if got := logKind(st.Log[i]); got != want {
			t.Errorf("line %d (%q) is coloured %v, want %v", i, st.Log[i].Message, got, want)
		}
	}
}
