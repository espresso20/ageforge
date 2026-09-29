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

// TestEventTextIsSane verifies every event's player-facing LogMessage and its
// Description are non-empty and free of broken tview color tags / stray printf
// directives. LogMessage is what the engine writes to the log on trigger
// (engine.processEvents); it must stay clean. Humour is additive — this guards
// against a rewrite that accidentally dropped text or introduced a bad token.
func TestEventTextIsSane(t *testing.T) {
	for _, e := range allEventDefs() {
		if strings.TrimSpace(e.LogMessage) == "" {
			t.Errorf("event %q has an empty LogMessage", e.Key)
		}
		if strings.TrimSpace(e.Description) == "" {
			t.Errorf("event %q has an empty Description", e.Key)
		}
		// The base-event LogMessages are written verbatim to the log; none use
		// dynamic-color tags, so a square bracket signals a typo, and a % is a
		// leftover format directive.
		for label, s := range map[string]string{"LogMessage": e.LogMessage, "Description": e.Description} {
			if strings.ContainsAny(s, "[]") {
				t.Errorf("event %q %s contains a square bracket (tview tag risk): %q", e.Key, label, s)
			}
			if pctDirective.MatchString(s) {
				t.Errorf("event %q %s contains a %% directive: %q", e.Key, label, s)
			}
		}
	}
}

// pctDirective matches a leftover printf verb ("%s", "%d"); a percentage such
// as "15%" is fine. The full effect check lives in TestEventTextMatchesEffects.
var pctDirective = regexp.MustCompile(`%[a-zA-Z]`)

// TestMilestoneFlavorIsSane verifies every milestone and chain carries a
// non-empty Flavor and that the strings are safe to render. The engine wraps
// the quip as fmt.Sprintf("  [gray]%s[-]", flavor), so the quip itself must
// contain no square brackets or % directives.
func TestMilestoneFlavorIsSane(t *testing.T) {
	guard := func(label, key, s string) {
		if strings.TrimSpace(s) == "" {
			t.Errorf("%s %q has an empty Flavor", label, key)
		}
		if strings.ContainsAny(s, "[]") {
			t.Errorf("%s %q Flavor has a square bracket (tview tag risk): %q", label, key, s)
		}
		if strings.Contains(s, "%") {
			t.Errorf("%s %q Flavor has a %% directive: %q", label, key, s)
		}
	}
	ms := Milestones()
	if len(ms) == 0 {
		t.Fatal("no milestones")
	}
	for _, m := range ms {
		guard("milestone", m.Key, m.Flavor)
	}
	for _, c := range MilestoneChains() {
		guard("chain", c.Key, c.Flavor)
	}
}

// TestEveryMilestoneHasFlavor asserts full coverage: no milestone or chain was
// left without a quip during the humour pass.
func TestEveryMilestoneHasFlavor(t *testing.T) {
	missing := 0
	for _, m := range Milestones() {
		if m.Flavor == "" {
			t.Errorf("milestone %q (%s) missing Flavor", m.Key, m.Name)
			missing++
		}
	}
	for _, c := range MilestoneChains() {
		if c.Flavor == "" {
			t.Errorf("chain %q (%s) missing Flavor", c.Key, c.Name)
			missing++
		}
	}
	if missing > 0 {
		t.Logf("%d milestone/chain Flavor fields are empty", missing)
	}
}
