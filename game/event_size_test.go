package game

import (
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// event_size_test.go holds the sizes of events (config/event_size.go) to
// the towns they happen to: every random event, era event and era
// transition event, in every one of the first eight ages, for a moderate
// town and for a maxed-out one. And the log line to what was applied.

// eventSizeAges is how many ages from the first the band test covers.
const eventSizeAges = 8

// maxedCopies is how many copies of an uncapped building the maxed-out town
// holds.
const maxedCopies = 100

// newMaxedEngine builds a town that has built everything: every building up
// to age at its max count (maxedCopies of one with none), fully staffed,
// every tech up to age, every earlier wonder, and every store full.
func newMaxedEngine(age string) *GameEngine {
	ge := newTruthEngine(age, truthTypical)
	order := ageOrders()
	pool := ge.Workers.domains["worker"]
	for _, key := range ge.Buildings.order {
		def := ge.Buildings.defs[key]
		if ge.Buildings.counts[key] == 0 || def.Category == "wonder" {
			continue
		}
		if order[def.RequiredAge] > order[age] {
			continue
		}
		n := maxedCopies
		if def.MaxCount > 0 {
			n = def.MaxCount
		}
		had := ge.Buildings.counts[key]
		ge.Buildings.counts[key] = n
		if def.WorkerCapacity > 0 {
			pool.assignments[key] = n * def.WorkerCapacity
			pool.count += (n - had) * def.WorkerCapacity
		}
	}
	ge.recalculateRates()
	ge.recalculateTickSpeed()
	for _, key := range ge.Resources.order {
		r := ge.Resources.resources[key]
		r.Amount = r.Storage
	}
	return ge
}

// eventsIn lists the random and era events that can fire in age, by key.
func eventsIn(age string) []config.EventDef {
	order := ageOrders()
	era := config.EpochForAge(age)
	var out []config.EventDef
	for _, def := range append(config.RandomEvents(), config.EpochExclusiveEvents()...) {
		if order[def.MinAge] > order[age] || (def.EpochKey != "" && def.EpochKey != era) {
			continue
		}
		out = append(out, def)
	}
	return out
}

// TestEventSizesStayInBand: in each of the first eight ages, every sized
// effect of every event that can fire there comes to an amount inside its
// band, in a moderate town and in a maxed-out one. A band exists for every
// one of them: the age has a typical income of the resource. The era
// transition events' rates are held to the same, in the ages an era opens
// in.
func TestEventSizesStayInBand(t *testing.T) {
	ages := ageKeys()[:eventSizeAges]
	for _, age := range ages {
		towns := []struct {
			name string
			ge   *GameEngine
		}{
			{"a moderate town", newTruthEngine(age, truthTypical)},
			{"a maxed-out town", newMaxedEngine(age)},
		}
		type sized struct {
			from string
			eff  config.Effect
		}
		var effects []sized
		for _, def := range eventsIn(age) {
			for _, eff := range def.Effects {
				if config.IsSizedEffect(eff.Type) {
					effects = append(effects, sized{"event " + def.Key, eff})
				}
			}
		}
		era, _ := config.EpochByKey()[config.EpochForAge(age)]
		if len(era.Ages) > 0 && era.Ages[0] == age && config.CatastropheAllowed(era.Key) {
			for _, ev := range append(config.GoodEpochEvents(), config.ChallengingEpochEvents()...) {
				for _, eff := range ev.Rates {
					if eff.Target == "" {
						eff.Target = era.PrimaryResource
					}
					effects = append(effects, sized{"era transition event " + ev.Key, eff})
				}
			}
		}
		if len(effects) == 0 {
			t.Errorf("%s: no event can fire", age)
		}
		for _, town := range towns {
			ge := town.ge
			for _, s := range effects {
				res := s.eff.Target
				if !ge.Resources.IsUnlocked(res) {
					// The Cultural Festival waits for culture
					// (rollGoodEpochEvent); nothing else may name a
					// resource the age has not unlocked.
					if s.from != "era transition event cultural_festival" {
						t.Errorf("%s, %s: %s is sized on %s, which is locked there", age, town.name, s.from, res)
					}
					continue
				}
				perTick := s.eff.Type == config.EventRate
				in := ge.eventTown(res, perTick)
				floor, ceil := config.EventBand(s.eff, in.Typical)
				if ceil <= 0 {
					t.Errorf("%s: %s is sized on %s, and the age has no typical income of it: the size has no band", age, s.from, res)
					continue
				}
				got := config.EventSize(s.eff, in)
				mag := got
				if mag < 0 {
					mag = -mag
				}
				what := "%s, %s: %s %s %s comes to %s, outside its band of %s to %s (stock %s, income %s, typical %s)"
				bad := mag > ceil*(1+1e-12) || mag < floor*(1-1e-12)
				if s.eff.Type == config.EventLoss {
					// A thin store loses a quarter of itself, under the floor.
					quarter := config.EventLossMostShare * in.Stock
					bad = mag > ceil*(1+1e-12) || mag > quarter*(1+1e-12) || (mag < floor*(1-1e-12) && mag < quarter*(1-1e-12))
				}
				if bad || mag <= 0 {
					t.Errorf(what, age, town.name, s.from, s.eff.Type, res, textfmt.Number(got), textfmt.Number(floor), textfmt.Number(ceil),
						textfmt.Number(in.Stock), textfmt.Number(in.Income), textfmt.Number(in.Typical))
				}
			}
		}
		// And the two towns are not the same town: the maxed-out one makes
		// and holds more of the age's first resource.
		mod, top := towns[0].ge, towns[1].ge
		if top.townIncome("food") <= mod.townIncome("food") || top.Resources.Get("food") <= mod.Resources.Get("food") {
			t.Errorf("%s: the maxed-out town makes %v food and holds %v; the moderate one %v and %v", age,
				top.townIncome("food"), top.Resources.Get("food"), mod.townIncome("food"), mod.Resources.Get("food"))
		}
	}
}

var eventDigits = regexp.MustCompile(`\d`)

// TestEventLineStatesWhatHappened: an event's log line is its own sentence
// and then what the event did, with the amounts that were applied: every
// gain and loss as it changed the stores, every worker lost, every rate as
// the rates pass applies it, and how long a timed event lasts. Nothing else
// in the line is a number.
func TestEventLineStatesWhatHappened(t *testing.T) {
	for _, def := range append(config.RandomEvents(), config.EpochExclusiveEvents()...) {
		ge := newTruthEngine(truthEventAge(def), truthTypical)
		ge.Resources.resources["soldiers"].Amount = 0 // no garrison: the losses land whole
		setWorkers(ge, 200)
		// Half of every store, in a store with room for any gain.
		before := map[string]float64{}
		for _, key := range ge.Resources.order {
			r := ge.Resources.resources[key]
			r.Storage = r.Storage * 1000
			before[key] = r.Amount
		}
		popBefore := ge.Workers.TotalPop()
		logged := len(ge.log)
		ge.Events.active = nil
		if def.Duration > 0 {
			ge.Events.active = append(ge.Events.active, ActiveEvent{Key: def.Key, Name: def.Name, TicksLeft: ge.rules.StretchTicks(ge.age, def.Duration), Effects: def.Effects})
		}
		ge.fireEvent(def)
		if len(ge.log) <= logged {
			t.Errorf("%s logged nothing", def.Key)
			continue
		}
		// The event's line is the one that carries its own sentence (the
		// debug lines around it are for dumps).
		line := ""
		for _, e := range ge.log[logged:] {
			if strings.HasPrefix(e.Message, def.LogMessage) {
				line = e.Message
			}
		}
		if line == "" {
			t.Errorf("%s: no line starts with the event's own sentence: %v", def.Key, ge.log[logged:])
			continue
		}
		rest := strings.ToLower(strings.TrimPrefix(line, def.LogMessage))
		take := func(frag, what string) {
			t.Helper()
			frag = strings.ToLower(frag)
			if !strings.Contains(rest, frag) {
				t.Errorf("%s: the line does not state %s (%q): %q", def.Key, what, frag, line)
				return
			}
			rest = strings.Replace(rest, frag, "", 1)
		}
		changed := 0
		for _, key := range ge.Resources.order {
			switch d := ge.Resources.Get(key) - before[key]; {
			case d > 0:
				changed++
				take(Amount(d, key), "the "+key+" gained")
			case d < 0:
				changed++
				take(Amount(-d, key), "the "+key+" lost")
			}
		}
		if lost := popBefore - ge.Workers.TotalPop(); lost > 0 {
			changed++
			take(textfmt.Count(lost, "worker", "workers"), "the workers lost")
		}
		for _, ae := range ge.Events.active {
			if ae.Key != def.Key {
				continue
			}
			for _, eff := range ae.Effects {
				if eff.Type != "production" {
					continue
				}
				changed++
				take(ResourceName(eff.Target)+" "+textfmt.Rate(eff.Value*ge.speedK()), "the rate of "+eff.Target)
				// The rate is what the rates pass will apply.
				off := ge.Resources.resources[eff.Target].Rate
				ge.recalculateRates()
				if got, want := ge.Resources.resources[eff.Target].Rate-off, eff.Value*ge.speedK(); got == 0 || got/want < 0.999 || got/want > 1.001 {
					// (off was read with the event already active or not,
					// depending on the last pass: compare against a pass
					// without it.)
					saved := ge.Events.active
					ge.Events.active = nil
					ge.recalculateRates()
					without := ge.Resources.resources[eff.Target].Rate
					ge.Events.active = saved
					ge.recalculateRates()
					if got := ge.Resources.resources[eff.Target].Rate - without; got/want < 0.999 || got/want > 1.001 {
						t.Errorf("%s: the line says %s %s, and the rate moves by %v", def.Key, eff.Target, textfmt.Rate(want), got)
					}
				}
			}
			take("for "+ge.durationLocked(ae.TicksLeft), "how long it lasts")
		}
		if changed == 0 {
			t.Errorf("%s did nothing in a moderate %s town: %q", def.Key, AgeName(ge.age), line)
		}
		if eventDigits.MatchString(rest) {
			t.Errorf("%s: the line quotes a number nothing backs (left over: %q): %q", def.Key, rest, line)
		}
		for _, e := range ge.log[logged:] {
			if strings.Contains(e.Message, "You lost") {
				t.Errorf("%s: a second line repeats the losses: %q", def.Key, e.Message)
			}
		}
	}
}

// TestEventSizeFollowsTheTown is the playtest's own case: a Beast Stampede
// in a Stone Age town with ten thousand food and wood took 20 food and 30
// wood. It now takes a share of what the town holds, so it is felt there,
// and the same share of a small town's stores does not ruin it.
func TestEventSizeFollowsTheTown(t *testing.T) {
	def := config.EventByKey()["beast_stampede"]
	stampede := func(stock float64) (food, wood float64, line string) {
		ge := newTruthEngine("stone_age", truthClean)
		ge.Resources.resources["soldiers"].Amount = 0
		for _, res := range []string{"food", "wood"} {
			r := ge.Resources.resources[res]
			r.Storage, r.Amount = 1e9, stock
		}
		logged := len(ge.log)
		ge.fireEvent(def)
		for _, e := range ge.log[logged:] {
			if strings.HasPrefix(e.Message, def.LogMessage) {
				line = e.Message
			}
		}
		return stock - ge.Resources.Get("food"), stock - ge.Resources.Get("wood"), line
	}
	food, wood, line := stampede(10000)
	if food != 800 || wood != 1000 {
		t.Errorf("a stampede through a town with 10K of each took %v food and %v wood, want 800 and 1,000 (8%% and 10%%)", food, wood)
	}
	if want := def.LogMessage + " Lost 800 food and 1K wood."; line != want {
		t.Errorf("the line reads %q, want %q", line, want)
	}
	// A small town loses the same share, or the token floor, and never more
	// than a quarter of what it holds.
	food, wood, _ = stampede(300)
	if food <= 0 || wood <= 0 || food > 75 || wood > 75 {
		t.Errorf("a stampede through a town with 300 of each took %v food and %v wood, want something and at most a quarter (75)", food, wood)
	}
}

// TestEventSizeSurvivesASaveAndLoad: an event is sized from the town as it
// stands when it fires, not from the rates the last tick left behind. A
// building that finished after that tick's rates pass is counted, so a game
// saved and loaded just before an event (whose rates the load worked out
// afresh) meets the same event as the game that played on.
func TestEventSizeSurvivesASaveAndLoad(t *testing.T) {
	isolateAccountDir(t)
	a := NewGameEngine()
	a.SeedRNG(7)
	a.SetTownForTest(map[string]int{"hut": 20, "stash": 10, "gathering_camp": 6}, 30)
	a.mu.Lock()
	a.Workers.Assign("worker", "gathering_camp", 12)
	a.mu.Unlock()
	a.StepTicks(20)
	// Four more camps finish after the tick's rates pass: the rates in
	// hand are now a tick out of date.
	a.mu.Lock()
	a.Buildings.counts["gathering_camp"] += 4
	stale := a.Resources.resources["food"].Rate
	a.mu.Unlock()
	if err := a.SaveGame("sized"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("sized"); err != nil {
		t.Fatal(err)
	}
	if fresh := b.Resources.resources["food"].Rate; fresh == stale {
		t.Fatalf("the loaded game's food rate equals the stale one (%v): the test needs them to differ", fresh)
	}
	harvest := config.EventByKey()["bountiful_harvest"]
	gained := func(ge *GameEngine) float64 {
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.permanentBonuses["food"] += 1e6 // room for the harvest, from the event's own rates pass
		before := ge.Resources.Get("food")
		ge.fireEvent(harvest)
		return ge.Resources.Get("food") - before
	}
	if ga, gb := gained(a), gained(b); ga != gb || ga <= 0 {
		t.Errorf("the game that played on gained %v food from the harvest and the one saved and loaded %v", ga, gb)
	}
}
