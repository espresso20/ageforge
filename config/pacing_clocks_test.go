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
	want := map[string]int{
		"age_of_plenty": 562, "trade_winds": 374, "cultural_festival": 374, "peaceful_century": 749,
		"the_famine": 312, "merchant_betrayal": 187, "epidemic": 468, "resource_drought": 234,
		"political_instability": 156, "economic_crash": 562, "the_dark_age": 374,
	}
	for _, ev := range append(GoodEpochEvents(), ChallengingEpochEvents()...) {
		if strings.Contains(ev.FlavorText, "{dur}") {
			t.Errorf("%s: unfilled duration in %q", ev.Key, ev.FlavorText)
		}
		w, timed := want[ev.Key]
		if !timed {
			if ev.Duration != 0 {
				t.Errorf("%s: lasts %d ticks; add it to this table", ev.Key, ev.Duration)
			}
			continue
		}
		if ev.Duration != w {
			t.Errorf("%s: lasts %d ticks, want %d", ev.Key, ev.Duration, w)
		}
		if !strings.Contains(ev.FlavorText, "for "+DurationText(ev.Duration)) {
			t.Errorf("%s: %q does not say %q", ev.Key, ev.FlavorText, DurationText(ev.Duration))
		}
	}
}

// TestAwakeningsStretched: an awakening lasts its typed length stretched for
// its trigger age (the Stone Age's stays as typed), and its text quotes it.
func TestAwakeningsStretched(t *testing.T) {
	want := map[string]int{
		"awakening_pottery_mastery": 250, "awakening_metallurgy": 1300, "awakening_steam_breakthrough": 520,
		"awakening_electrification": 780, "awakening_information_age": 780, "awakening_cybernetic": 650,
		"awakening_first_contact": 1040,
	}
	for _, a := range Awakenings() {
		if a.Duration != want[a.Key] {
			t.Errorf("%s: lasts %d ticks, want %d", a.Key, a.Duration, want[a.Key])
		}
		if strings.Contains(a.FlavorText, "{dur}") || !strings.Contains(a.FlavorText, "for "+DurationText(a.Duration)+".") {
			t.Errorf("%s: %q should end its effect with \"for %s.\"", a.Key, a.FlavorText, DurationText(a.Duration))
		}
	}
}

// TestTradeRoutesStretched: a route's run is stretched for its MinAge, which
// is right in every age it runs only if no later age stretches differently:
// every route opens from the Bronze Age on.
func TestTradeRoutesStretched(t *testing.T) {
	order := AgeOrder()
	idx := map[string]int{}
	for i, a := range order {
		idx[a] = i
	}
	for _, r := range BaseTradeRoutes() {
		for _, later := range order[idx[r.MinAge]:] {
			if AgeStretch(later) != AgeStretch(r.MinAge) {
				t.Errorf("%s opens in %s (stretch %v) but runs on into %s (stretch %v): stretch it in the engine instead",
					r.Key, r.MinAge, AgeStretch(r.MinAge), later, AgeStretch(later))
				break
			}
		}
	}
	if r := TradeRouteByKey()["local_barter"]; r.TicksPerRun != 26 {
		t.Errorf("local_barter runs every %d ticks, want 26 (10 x 2.6)", r.TicksPerRun)
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
