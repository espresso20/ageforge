package config

import (
	"regexp"
	"strings"
	"testing"
)

// allEventDefs returns the base random events plus the epoch-exclusive events —
// the full set whose player-facing text the humour pass (9b) rewrote.
func allEventDefs() []EventDef {
	out := append([]EventDef{}, RandomEvents()...)
	out = append(out, EpochExclusiveEvents()...)
	return out
}

// TestEventTextIsSane verifies every event's player-facing LogMessage is
// non-empty. LogMessage is what the engine writes to the log on trigger
// (engine.processEvents). TestConfigPlayerTextStyle holds the text to the
// style rules, except for the raid events, whose text it leaves to the Army
// PR: those are held here to the tview-tag and printf rules.
func TestEventTextIsSane(t *testing.T) {
	for _, e := range allEventDefs() {
		if strings.TrimSpace(e.LogMessage) == "" {
			t.Errorf("event %q has an empty LogMessage", e.Key)
		}
		if !raidEvents[e.Key] {
			continue
		}
		if strings.ContainsAny(e.LogMessage, "[]") {
			t.Errorf("event %q LogMessage contains a square bracket (tview tag risk): %q", e.Key, e.LogMessage)
		}
		if pctDirective.MatchString(e.LogMessage) {
			t.Errorf("event %q LogMessage contains a %% directive: %q", e.Key, e.LogMessage)
		}
	}
}

// pctDirective matches a leftover printf verb ("%s", "%d"); a percentage such
// as "15%" is fine. The full effect check lives in TestEventTextMatchesEffects.
var pctDirective = regexp.MustCompile(`%[a-zA-Z]`)

// TestMilestoneFlavorIsSane verifies every milestone and chain carries a
// non-empty Flavor (TestConfigPlayerTextStyle skips an empty one).
func TestMilestoneFlavorIsSane(t *testing.T) {
	ms := Milestones()
	if len(ms) == 0 {
		t.Fatal("no milestones")
	}
	for _, m := range ms {
		if strings.TrimSpace(m.Flavor) == "" {
			t.Errorf("milestone %q has an empty Flavor", m.Key)
		}
	}
	for _, c := range MilestoneChains() {
		if strings.TrimSpace(c.Flavor) == "" {
			t.Errorf("chain %q has an empty Flavor", c.Key)
		}
	}
}
