package ui

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestThePeopleTabsPickLinesByWhatTheyAreAbout: a line is picked by the words
// it has. The colour of a line follows the same rule.
func TestThePeopleTabsPickLinesByWhatTheyAreAbout(t *testing.T) {
	st := game.GameState{Log: []game.LogEntry{
		{Message: "Recruited 3 workers (population 4/9).", Type: game.LogRoutine},
		{Message: "Built a Hut (you have 4).", Type: game.LogRoutine},
		{Message: "Squad of 3 pulled up, no cap.", Type: game.LogRoutine},
		{Message: "The soldiers lost the plot.", Type: "info"},
		{Message: "Two of the squad took an L.", Type: "warning"},
		{Message: "A debug line about workers.", Type: "debug"},
	}}
	messages := func(lines []game.LogEntry) []string {
		var out []string
		for _, l := range lines {
			out = append(out, l.Message)
		}
		return out
	}
	if got, want := messages(recentLog(st, peopleEvent, 10)), []string{
		"The soldiers lost the plot.", "Recruited 3 workers (population 4/9).",
	}; !reflect.DeepEqual(got, want) {
		t.Errorf("the People tab picks %q, want %q", got, want)
	}
	if got, want := messages(recentLog(st, armyEvent, 10)), []string{"The soldiers lost the plot."}; !reflect.DeepEqual(got, want) {
		t.Errorf("the Army tab picks %q, want %q", got, want)
	}
	if got, want := messages(recentLog(st, peopleEvent, 1)), []string{"The soldiers lost the plot."}; !reflect.DeepEqual(got, want) {
		t.Errorf("the newest line picked is %q, want %q", got, want)
	}
	for i, want := range []xk{xGood, xText, xText, xBad, xText, xText} {
		if got := logKind(st.Log[i]); got != want {
			t.Errorf("line %d (%q) is coloured %v, want %v", i, st.Log[i].Message, got, want)
		}
	}
}

// renderPeopleGrid is the panel's cells with no recorder.
func renderPeopleGrid(st game.GameState, v peopleView, w, h int) *xgrid {
	g, _ := renderPeople(st, v, nil, w, h)
	return g
}

// peopleText is every word of the panel: its cells and the Overview's scene.
func peopleText(st game.GameState, v peopleView, w, h int) string {
	g, o := renderPeople(st, v, nil, w, h)
	s := g.String()
	if o != nil {
		for _, wd := range o.s.words {
			s += "\n" + wd.s
		}
	}
	return s
}
