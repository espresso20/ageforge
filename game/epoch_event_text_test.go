package game

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
		// A town that makes something of everything, so every rate has a
		// production to be a share of.
		ge := newTruthEngine("classical_age", truthClean)
		for _, r := range tracked {
			ge.Resources.UnlockResource(r)
			ge.Resources.AddStorage(r, 1e9)
			ge.Resources.Add(r, 1000-ge.Resources.Get(r))
		}
		ge.Workers.domains["worker"].count = 100
		before := map[string]float64{}
		for _, r := range tracked {
			before[r] = ge.Resources.Get(r)
		}
		popBefore := float64(ge.Workers.TotalPop())
		bonusBefore := ge.permanentBonuses["production_all"]
		era, _ := ge.rules.Era(ge.currentEpoch)
		logged := len(ge.log)

		if good[ev.Key] {
			ge.applyGoodEpochEvent(ev)
		} else {
			ge.applyChallengingEpochEvent(ev, ge.currentEpoch)
		}

		text := strings.ToLower(ev.FlavorText)
		need := func(frag string) {
			t.Helper()
			if !strings.Contains(text, strings.ToLower(frag)) {
				t.Errorf("%s: %q is missing %q", ev.Key, ev.FlavorText, frag)
			}
		}

		// Each timed rate is a share of the town's production: the text
		// states the share, the engine applies what it comes to here, and
		// the line under the headline states that amount.
		rates := map[string]float64{}
		for _, r := range ev.Rates {
			label, res := strings.ToLower(ResourceName(r.Target)), r.Target
			if res == "" {
				label, res = "its", era.PrimaryResource
			}
			need(label + " production " + textfmt.SignedPercent(r.Value))
			sized := r
			sized.Target = res
			rates[res] = config.EventSize(sized, ge.eventTown(res, true))
			if rates[res] == 0 {
				t.Errorf("%s: the test town makes no %s, so the rate reads as nothing", ev.Key, res)
			}
		}
		applied := 0
		for _, ae := range ge.Events.active[len(ge.Events.active)-min(1, len(ge.Events.active)):] {
			if ae.Name != ev.Name {
				continue
			}
			need("for " + config.DurationText(ae.TicksLeft))
			for _, e := range ae.Effects {
				switch {
				case e.Type == "production":
					applied++
					if want, ok := rates[e.Target]; !ok || math.Abs(e.Value-want) > 1e-9*math.Abs(want) {
						t.Errorf("%s: %s changes by %v a tick, want %v (the share its text states)", ev.Key, e.Target, e.Value, want)
					}
					line := ""
					for _, entry := range ge.log[logged:] {
						if strings.Contains(strings.ToLower(entry.Message), strings.ToLower(ResourceName(e.Target))+" "+textfmt.Rate(e.Value)) {
							line = entry.Message
						}
					}
					if line == "" {
						t.Errorf("%s: no log line states the amount, %s %s: %v", ev.Key, ResourceName(e.Target), textfmt.Rate(e.Value), ge.log[logged:])
					}
				case e.Type == "production_all":
					need("all production +" + config.FormatPercent(e.Value))
				default:
					t.Errorf("%s: effect type %q has no wording check; add one", ev.Key, e.Type)
				}
			}
		}
		if applied != len(ev.Rates) {
			t.Errorf("%s: %d of its %d rates were applied", ev.Key, applied, len(ev.Rates))
		}

		for _, r := range tracked {
			after := ge.Resources.Get(r)
			switch {
			case ev.Key == "ancient_cache":
				// Minutes of the town's own income: the text states how
				// many, read here off the gold the store took.
				if r == "gold" {
					town := ge.eventTown(r, false)
					perMinute := config.EventSize(config.Effect{Type: config.EventGain, Target: r, Value: 1}, town)
					if perMinute <= 0 || after <= before[r] {
						t.Errorf("%s: the test town gained no gold, so the minutes read as nothing", ev.Key)
						break
					}
					need(fmt.Sprintf("%g minutes of what your town makes", math.Round((after-before[r])/perMinute*1e6)/1e6))
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

// TestEpochAndAwakeningLogsDoNotRepeatText: the flavor text states each
// effect, so an event with a percentage or a permanent bonus logs its
// headline and nothing else. An event with a timed rate adds one line: its
// text states the rate as a share of the town's production, and the line
// states the amount that comes to (and, for Resource Drought, the resource).
func TestEpochAndAwakeningLogsDoNotRepeatText(t *testing.T) {
	byKey := map[string]config.EpochEventDef{}
	for _, ev := range append(config.GoodEpochEvents(), config.ChallengingEpochEvents()...) {
		byKey[ev.Key] = ev
	}
	apply := func(key string) []LogEntry {
		ge := newTruthEngine("classical_age", truthClean)
		before := len(ge.log)
		if ev := byKey[key]; strings.HasPrefix(ev.Type, "good") {
			ge.applyGoodEpochEvent(ev)
		} else {
			ge.applyChallengingEpochEvent(ev, ge.currentEpoch)
		}
		return ge.log[before:]
	}
	for _, key := range []string{"age_of_plenty", "ancient_cache", "worker_innovation", "peaceful_century", "epoch_blessing"} {
		if got := apply(key); len(got) != 1 {
			t.Errorf("%s logged %d lines, want only the headline: %v", key, len(got), got)
		}
	}
	for _, key := range []string{"trade_winds", "the_famine", "resource_drought"} {
		got := apply(key)
		if len(got) != 2 || !strings.Contains(got[len(got)-1].Message, "/tick for ") {
			t.Errorf("%s should add one line with the amount its rate comes to; got %v", key, got)
		}
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
