package config

import (
	"strings"
	"testing"
)

// pacing_clocks_test.go: the clocks config re-times for the one-week curve
// (StretchTicks), and the texts that quote them.

// TestEpochEventsStretched: epoch events fire on entering an era from the
// Iron Era on, every one of which runs at the full stretch, so each timed
// event lasts its typed length × PacingStretch and its text says so.
func TestEpochEventsStretched(t *testing.T) {
	for _, ep := range Epochs()[1:] {
		if AgeStretch(ep.Ages[0]) != AgeStretch(epochEventAge) {
			t.Errorf("%s opens at stretch %v, but epoch events are timed for %v", ep.Key, AgeStretch(ep.Ages[0]), AgeStretch(epochEventAge))
		}
	}
	for _, ev := range append(GoodEpochEvents(), ChallengingEpochEvents()...) {
		if strings.Contains(ev.FlavorText, "{dur}") {
			t.Errorf("%s: unfilled duration in %q", ev.Key, ev.FlavorText)
		}
		if ev.Duration == 0 {
			continue
		}
		if !strings.Contains(ev.FlavorText, "for "+DurationText(ev.Duration)) {
			t.Errorf("%s: %q does not say %q", ev.Key, ev.FlavorText, DurationText(ev.Duration))
		}
	}
}

// TestAwakeningsStretched: an awakening lasts its typed length stretched for
// its trigger age (the Stone Age's stays as typed), and its text quotes it.
func TestAwakeningsStretched(t *testing.T) {
	for _, a := range Awakenings() {
		if strings.Contains(a.FlavorText, "{dur}") || !strings.Contains(a.FlavorText, "for "+DurationText(a.Duration)+".") {
			t.Errorf("%s: %q should end its effect with \"for %s.\"", a.Key, a.FlavorText, DurationText(a.Duration))
		}
	}
}

// TestSurvivorMilestonesStretched: the play-time milestones count 2.6x the
// ticks they did (about 14h 26m and 3 days); TestMilestoneDescriptionNumbers
// holds their descriptions to it.
func TestSurvivorMilestonesStretched(t *testing.T) {
	want := map[string]int{"survivor": 26000, "enduring_civilization": 130000}
	for _, m := range Milestones() {
		if w, ok := want[m.Key]; ok && m.MinTick != w {
			t.Errorf("%s: MinTick %d, want %d", m.Key, m.MinTick, w)
		}
	}
}

// epochEventAge is the age epoch events are timed for. They fire on entering
// an era, and the first era they can fire in, the Iron Era, already runs at
// PacingStretch, so every one does.
const epochEventAge = "iron_age"
