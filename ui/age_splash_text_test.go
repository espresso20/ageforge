package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// buildAgeSplashText joins the arrival's lines into the full text of an age
// advance. Only the tests read the text whole; the screen lays the lines out
// (ageSplashLines).
func buildAgeSplashText(newAge string, summary game.AgeAdvanceSummary,
	epochChanged bool, epochEvent game.EpochEventRecord) string {
	lines := ageSplashLines(newAge, summary, epochChanged, epochEvent)
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.text
	}
	return strings.Join(out, "\n")
}

// splashCase is one age advance to build the splash text for.
type splashCase struct {
	age     string
	summary game.AgeAdvanceSummary
	epoch   bool
	event   game.EpochEventRecord
}

// splashCases is every age with and without upgrades on offer, and every
// age that opens an epoch with each kind of event.
func splashCases() []splashCase {
	var out []splashCase
	full := game.AgeAdvanceSummary{
		BuildingsTransformed: []game.BuildingTransform{
			{OldName: "Elders' Hall", NewName: "Scriptorium", Count: 5},
			{OldKey: "long_house", NewKey: "house", Count: 15},
			{OldName: "Nuclear Extraction Plant", NewName: "Uranium Processing Works", Count: 120},
		},
	}
	events := config.EpochEventByKey()
	for _, a := range config.Ages() {
		out = append(out, splashCase{age: a.Key})
		out = append(out, splashCase{age: a.Key, summary: full})
		era := config.EpochForAge(a.Key)
		ep := config.EpochByKey()[era]
		if len(ep.Ages) == 0 || ep.Ages[0] != a.Key {
			continue
		}
		for _, kind := range []string{"good_minor", "good_major", "good_legendary", "bad_challenging", "catastrophe"} {
			rec := game.EpochEventRecord{EpochKey: era, EpochName: ep.Name, EventKey: "none", EventName: "An event", EventType: kind}
			for key, ev := range events {
				if ev.Type == kind {
					rec.EventKey, rec.EventName = key, ev.Name
					break
				}
			}
			out = append(out, splashCase{age: a.Key, summary: full, epoch: true, event: rec})
		}
	}
	return out
}

// TestAgeSplashTextIsWhatItWas: for every age, with and without upgrades,
// and with every kind of epoch event, each line of the arrival's text is one
// line (no line break inside it) and is blank exactly when it is of the
// blank kind.
func TestAgeSplashTextIsWhatItWas(t *testing.T) {
	cases := splashCases()
	if len(cases) < 2*len(config.Ages())+5 {
		t.Fatalf("only %d cases", len(cases))
	}
	for _, c := range cases {
		for _, l := range ageSplashLines(c.age, c.summary, c.epoch, c.event) {
			if strings.Contains(l.text, "\n") {
				t.Errorf("%s: a line holds a line break: %q", c.age, l.text)
			}
			if (l.kind == skBlank) != (l.text == "") {
				t.Errorf("%s: line %q is of kind %d", c.age, l.text, l.kind)
			}
		}
	}
}
