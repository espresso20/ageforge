package game

import (
	"testing"
)

// heard subscribes to the record of a run's ending and returns where it
// lands.
func heard(ge *GameEngine) *[]RunEnding {
	var got []RunEnding
	ge.Bus.Subscribe(EventRunEnded, func(e EventData) {
		if end, ok := e.Payload["ending"].(RunEnding); ok {
			got = append(got, end)
		}
	})
	return &got
}

// lineKinds is the kinds of an ending's lines, in order.
func lineKinds(end RunEnding) (out []string) {
	for _, l := range end.Lines {
		out = append(out, l.Kind)
	}
	return out
}

func hasKind(end RunEnding, kind string) bool {
	for _, l := range end.Lines {
		if l.Kind == kind {
			return true
		}
	}
	return false
}

// TestRunEndingRecord: every way a run ends publishes one record, after the
// reset, that says how it ended, where, and what it earned, and carries the
// lines the game wrote about it, each marked by what it is. The lines are
// the log's own: nothing is in the record that the log does not hold, and
// the log is as it was before there was a record.
func TestRunEndingRecord(t *testing.T) {
	check := func(name string, ge *GameEngine, got *[]RunEnding, kind, age string, level int) RunEnding {
		t.Helper()
		if len(*got) != 1 {
			t.Fatalf("%s: %d records", name, len(*got))
		}
		end := (*got)[0]
		if end.Kind != kind || end.Age != age || end.Level != level || ge.age != "primitive_age" || ge.tick != 0 {
			t.Errorf("%s: the record says %s from %s at level %d (the new run: %s, tick %d)", name, end.Kind, end.Age, end.Level, ge.age, ge.tick)
		}
		if end.Badges != -1 {
			t.Errorf("%s: %d badges counted with no account", name, end.Badges)
		}
		// Every line of the record is a line of the new run's log, in order.
		at := 0
		for _, l := range end.Lines {
			found := false
			for ; at < len(ge.log); at++ {
				if ge.log[at].Message == l.Text && ge.log[at].Type == l.Level {
					found, at = true, at+1
					break
				}
			}
			if !found {
				t.Errorf("%s: a %s line of the record is not in the log, or out of its order: %q", name, l.Kind, l.Text)
			}
		}
		return end
	}

	// A prestige from the Medieval Age: the early taste.
	ge := catEngine(t, "medieval_age", 5)
	got := heard(ge)
	want := ge.GetState().Prestige.PendingPoints
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	end := check("a prestige", ge, got, RunEndPrestige, "medieval_age", 1)
	if end.Points != want || end.Full != want || want <= 0 || !end.Prestige() || end.Came() || end.Catastrophe != "" {
		t.Errorf("a prestige: %d of %d points, want %d", end.Points, end.Full, want)
	}
	for _, kind := range []string{EndLineComplete, EndLineEarly, EndLineMastery} {
		if !hasKind(end, kind) {
			t.Errorf("a prestige: no %s line (%v)", kind, lineKinds(end))
		}
	}
	if hasKind(end, EndLineVerdict) {
		t.Error("a plain prestige has a verdict")
	}

	// The Last Passage, endured: the verdict first, part of the points.
	ge = lpEngine(t, "transcendent_age", 6)
	got = heard(ge)
	full := ge.GetState().Prestige.PendingPoints
	makeLastPassagePending(t, ge)
	if len(*got) != 0 {
		t.Fatal("a prestige that waits on the Last Passage has not ended the run")
	}
	if err := ge.EndureLastPassage(); err != nil {
		t.Fatal(err)
	}
	end = check("the Last Passage endured", ge, got, RunEndEndured, "transcendent_age", 1)
	if end.Full != full || end.Points <= 0 || end.Points >= full || !end.Came() || end.Catastrophe == "" {
		t.Errorf("the Last Passage endured: %d of %d points, %q", end.Points, end.Full, end.Catastrophe)
	}
	if kinds := lineKinds(end); len(kinds) < 3 || kinds[0] != EndLineVerdict {
		t.Errorf("the Last Passage endured: the verdict does not come first (%v)", kinds)
	}

	// Succumbed: no points, and the Cosmic Legacy.
	ge = lpEngine(t, "transcendent_age", 6)
	got = heard(ge)
	makeLastPassagePending(t, ge)
	if err := ge.SuccumbLastPassage(); err != nil {
		t.Fatal(err)
	}
	end = check("the Last Passage succumbed to", ge, got, RunEndSuccumbed, "transcendent_age", 1)
	if end.Points != 0 || end.Full <= 0 || !hasKind(end, EndLineLegacy) || lineKinds(end)[0] != EndLineVerdict {
		t.Errorf("the Last Passage succumbed to: %d of %d points (%v)", end.Points, end.Full, lineKinds(end))
	}

	// A fall to an era's catastrophe: no prestige, the level as it was.
	ge = catEngine(t, "iron_age", 5)
	setPrestigeLevel(ge, 2)
	got = heard(ge)
	succumbIn(t, ge, "iron_age")
	end = check("a fall", ge, got, RunEndFallen, "iron_age", 2)
	if end.Prestige() || !end.Came() || end.Catastrophe == "" || end.Points != 0 || hasKind(end, EndLineComplete) {
		t.Errorf("a fall: %+v", end)
	}
	if kinds := lineKinds(end); kinds[0] != EndLineVerdict || !hasKind(end, EndLineLegacy) || !hasKind(end, EndLineKnowledge) {
		t.Errorf("a fall: %v", kinds)
	}
}
