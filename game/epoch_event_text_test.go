package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// epoch_event_text_test.go pins the effect text in config (epoch event
// FlavorText) to what the engine really applies. The effects live in
// applyGoodEpochEvent / applyChallengingEpochEvent, so the check has to run
// here: it applies each event to a fresh engine and measures what changed.

// TestEpochEventTextMatchesEffects: every timed effect, instant percentage
// and worker change an epoch event applies appears in its FlavorText with the
// real number and duration.
func TestEpochEventTextMatchesEffects(t *testing.T) {
	good := map[string]bool{}
	for _, ev := range config.GoodEpochEvents() {
		good[ev.Key] = true
	}
	all := append(config.GoodEpochEvents(), config.ChallengingEpochEvents()...)
	tracked := []string{"gold", "faith", "knowledge", "culture"}
	for _, ev := range all {
		ge := NewGameEngine()
		for _, r := range tracked {
			ge.Resources.UnlockResource(r)
			ge.Resources.AddStorage(r, 1e6)
			ge.Resources.Add(r, 1000)
		}
		ge.Workers.domains["worker"].count = 100
		before := map[string]float64{}
		for _, r := range tracked {
			before[r] = ge.Resources.Get(r)
		}
		popBefore := float64(ge.Workers.TotalPop())
		bonusBefore := ge.permanentBonuses["production_all"]

		if good[ev.Key] {
			ge.applyGoodEpochEvent(ev)
		} else {
			ge.applyChallengingEpochEvent(ev, "stone_era")
		}

		text := strings.ToLower(ev.FlavorText)
		need := func(frag string) {
			t.Helper()
			if !strings.Contains(text, strings.ToLower(frag)) {
				t.Errorf("%s: %q is missing %q", ev.Key, ev.FlavorText, frag)
			}
		}

		for _, ae := range ge.Events.active {
			need("for " + config.DurationText(ae.TicksLeft))
			for _, e := range ae.Effects {
				switch {
				case e.Type == "production" && ev.Key == "resource_drought":
					// The resource depends on the epoch; the text names the amount.
					need(" " + signedAmount(e.Value) + "/tick")
				case e.Type == "production":
					need(config.RateText(e.Target, e.Value))
				case e.Type == "production_all":
					need("all production +" + config.FormatPercent(e.Value))
				default:
					t.Errorf("%s: effect type %q has no wording check; add one", ev.Key, e.Type)
				}
			}
		}

		for _, r := range tracked {
			after := ge.Resources.Get(r)
			switch {
			case ev.Key == "ancient_cache":
				if r == "gold" {
					need(config.FormatPercent((after-before[r])/ge.Resources.GetStorage(r)) + " of its storage")
				}
			case after < before[r]:
				need(config.FormatPercent(1-after/before[r]) + " of your " + r)
			case after > before[r]:
				need(r + " +" + config.FormatPercent(after/before[r]-1))
			}
		}

		if pop := float64(ge.Workers.TotalPop()); pop != popBefore {
			pct := config.FormatPercent(math.Abs(pop/popBefore - 1))
			if pop > popBefore {
				need("grows by " + pct)
			} else {
				need(pct + " of your workers")
			}
		}
		if d := ge.permanentBonuses["production_all"] - bonusBefore; d != 0 {
			need("all production +" + config.FormatPercent(d) + ", permanently")
		}
	}
}

// signedAmount formats a signed per-tick amount like the text does: "-3".
func signedAmount(v float64) string {
	if v < 0 {
		return "-" + config.FormatAmount(v)
	}
	return "+" + config.FormatAmount(v)
}

// TestEpochAndAwakeningLogsDoNotRepeatText: the flavor text states each
// effect, so a timed-effect or permanent-bonus event logs its headline and
// nothing else. Resource Drought is the exception: its text cannot name the
// resource.
func TestEpochAndAwakeningLogsDoNotRepeatText(t *testing.T) {
	byKey := map[string]config.EpochEventDef{}
	for _, ev := range append(config.GoodEpochEvents(), config.ChallengingEpochEvents()...) {
		byKey[ev.Key] = ev
	}
	for _, key := range []string{"age_of_plenty", "ancient_cache", "trade_winds", "worker_innovation", "peaceful_century", "epoch_blessing", "the_famine"} {
		ge := NewGameEngine()
		before := len(ge.log)
		if ev := byKey[key]; strings.HasPrefix(ev.Type, "good") {
			ge.applyGoodEpochEvent(ev)
		} else {
			ge.applyChallengingEpochEvent(ev, "stone_era")
		}
		if got := len(ge.log) - before; got != 1 {
			t.Errorf("%s logged %d lines, want only the headline: %v", key, got, ge.log[before:])
		}
	}

	ge := NewGameEngine()
	before := len(ge.log)
	ge.applyChallengingEpochEvent(byKey["resource_drought"], "stone_era")
	if n := len(ge.log) - before; n != 2 || !strings.Contains(ge.log[len(ge.log)-1].Message, "/tick for ") {
		t.Errorf("resource_drought should add one line naming the resource; got %v", ge.log[before:])
	}

	for _, def := range config.Awakenings() {
		ge := NewGameEngine()
		before := len(ge.log)
		ge.fireAwakening(def.TriggerAge)
		if got := len(ge.log) - before; got != 1 {
			t.Errorf("awakening %s logged %d lines, want 1: %v", def.Key, got, ge.log[before:])
		}
	}
}
