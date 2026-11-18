package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

const featEntry = 100 // the tick the test run entered its age

// featEngine is a run with an account, in age since tick featEntry, at tick.
func featEngine(t *testing.T, age string, at int) (*GameEngine, *Account) {
	t.Helper()
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	ge.mu.Lock()
	ge.age, ge.tick = age, at
	ge.Stats.AgeLog = []AgeEntry{{Age: age, Tick: featEntry}}
	ge.mu.Unlock()
	return ge, acct
}

// inAge is the tick at which the run has spent share of age's target in it.
func inAge(age string, share float64) int {
	return featEntry + int(share*config.AgeTargetTicks(age))
}

func wantFeat(t *testing.T, name string, acct *Account, age string, want bool) {
	t.Helper()
	s := acct.CoinStackOf(config.CoinFeat, age)
	if got := len(s.Copies) == 1; got != want {
		t.Errorf("%s: feat coin struck %v, want %v (%+v)", name, got, want, s)
	} else if got && (s.Copies[0].Level != config.CoinLegendary || s.Copies[0].Seed == 0) {
		t.Errorf("%s: copy %+v", name, s.Copies[0])
	}
}

// The feats an event decides strike on the event that meets them and not before.
func TestFeatsDecidedByAnEvent(t *testing.T) {
	keystone := rules.Core().Keystone("iron_age")
	renaissance := rules.Core().TechsOf("renaissance_age")
	cases := []struct {
		name, age string
		set       func(ge *GameEngine)
		ev        Event
		want      bool
	}{
		{"31 huts", "primitive_age", func(ge *GameEngine) { ge.Buildings.counts["hut"] = 31 }, Event{Kind: config.BadgeEvBuildingBuilt, Subject: "hut"}, true},
		{"30 huts", "primitive_age", func(ge *GameEngine) { ge.Buildings.counts["hut"] = 30 }, Event{Kind: config.BadgeEvBuildingBuilt, Subject: "hut"}, false},
		{"keystone in time", "iron_age", func(ge *GameEngine) { ge.Stats.Deeds.KeystoneTick = featEntry + 650 }, Event{Kind: config.BadgeEvResearchDone, Subject: keystone}, true},
		{"keystone late", "iron_age", func(ge *GameEngine) { ge.Stats.Deeds.KeystoneTick = featEntry + 651 }, Event{Kind: config.BadgeEvResearchDone, Subject: keystone}, false},
		{"keystone in time at 2x", "iron_age", func(ge *GameEngine) {
			ge.Prestige.speeds = map[string]float64{"iron_age": 2}
			ge.Stats.Deeds.KeystoneTick = featEntry + 325
		}, Event{Kind: config.BadgeEvResearchDone, Subject: keystone}, true},
		{"keystone late at 2x", "iron_age", func(ge *GameEngine) {
			ge.Prestige.speeds = map[string]float64{"iron_age": 2}
			ge.Stats.Deeds.KeystoneTick = featEntry + 326
		}, Event{Kind: config.BadgeEvResearchDone, Subject: keystone}, false},
		{"eight festivals", "classical_age", func(ge *GameEngine) { ge.Stats.Deeds.Festivals = 8 }, Event{Kind: config.BadgeEvFestival}, true},
		{"seven festivals", "classical_age", func(ge *GameEngine) { ge.Stats.Deeds.Festivals = 7 }, Event{Kind: config.BadgeEvFestival}, false},
		{"two festivals at 4x", "classical_age", func(ge *GameEngine) {
			ge.Prestige.speeds = map[string]float64{"classical_age": 4}
			ge.Stats.Deeds.Festivals = 2
		}, Event{Kind: config.BadgeEvFestival}, true},
		{"one festival at 40x", "classical_age", func(ge *GameEngine) {
			ge.Prestige.speeds = map[string]float64{"classical_age": 40}
			ge.Stats.Deeds.Festivals = 1
		}, Event{Kind: config.BadgeEvFestival}, true},
		{"every Renaissance tech", "renaissance_age", func(ge *GameEngine) {
			for _, tc := range renaissance {
				ge.Research.researched[tc.Key] = true
			}
		}, Event{Kind: config.BadgeEvResearchDone, Subject: renaissance[0].Key}, true},
		{"one tech short", "renaissance_age", func(ge *GameEngine) {
			for _, tc := range renaissance[1:] {
				ge.Research.researched[tc.Key] = true
			}
		}, Event{Kind: config.BadgeEvResearchDone, Subject: renaissance[1].Key}, false},
	}
	for _, c := range cases {
		ge, acct := featEngine(t, c.age, inAge(c.age, 0.1))
		ge.mu.Lock()
		c.set(ge)
		ge.report(c.ev)
		ge.mu.Unlock()
		wantFeat(t, c.name, acct, c.age, c.want)
	}
	// Another age's feat is not judged: festivals in the Iron Age strike nothing.
	ge, acct := featEngine(t, "iron_age", featEntry)
	ge.mu.Lock()
	ge.Stats.Deeds.Festivals = 9
	ge.report(Event{Kind: config.BadgeEvFestival})
	ge.mu.Unlock()
	wantFeat(t, "festivals out of age", acct, "classical_age", false)
	wantFeat(t, "festivals out of age", acct, "iron_age", false)
}

// The feats decided on leaving strike as the age is left, from the age's own deeds.
func TestFeatsDecidedOnLeavingTheAge(t *testing.T) {
	var through float64 // the summed reference times through Modern
	for _, a := range rules.Core().AgeKeys() {
		through += config.CoinRefTicks(a, 1)
		if a == "modern_age" {
			break
		}
	}
	plant, _ := rules.Core().Building("nuclear_plant")
	need := 3 * plant.BaseCost["electricity"]
	warp := func(from int) func(*GameEngine) {
		return func(ge *GameEngine) {
			ge.Stats.Deeds.RouteClean = map[string]int{"warp_commerce": from}
		}
	}
	faith := func(own, held float64) func(*GameEngine) {
		return func(ge *GameEngine) {
			ge.faithMeasure = FaithSave{Moderate: 1, Own: own}
			setResource(ge, "faith", held)
		}
	}
	ally := func(status string) func(*GameEngine) {
		return func(ge *GameEngine) {
			ge.Diplomacy.DiscoverFaction("shadow_syndicate")
			ge.Diplomacy.factions["shadow_syndicate"].Status = status
		}
	}
	cases := []struct {
		name, age string
		share     float64 // of the age's target spent in it (0 leaves the tick at the entry)
		set       func(ge *GameEngine)
		want      bool
	}{
		{"stone, under half", "stone_age", 0.49, nil, true},
		{"stone, at half", "stone_age", 0.5, nil, false},
		{"bronze, no upgrade", "bronze_age", 0.59, nil, true},
		{"bronze, one upgrade", "bronze_age", 0.59, func(ge *GameEngine) { ge.Stats.Deeds.Upgrades = 1 }, false},
		{"bronze, too slow", "bronze_age", 0.61, nil, false},
		{"electric, three plants", "electric_age", 0.1, func(ge *GameEngine) { setResource(ge, "electricity", need) }, true},
		{"electric, short", "electric_age", 0.1, func(ge *GameEngine) { setResource(ge, "electricity", need*0.99) }, false},
		{"atomic, no pacts", "atomic_age", 0.85, nil, true},
		{"atomic, a pact", "atomic_age", 0.85, func(ge *GameEngine) { ge.Stats.Deeds.Pacts = 1 }, false},
		{"atomic, too slow", "atomic_age", 0.87, nil, false},
		{"cyberpunk, allied", "cyberpunk_age", 0.1, ally("allied"), true},
		{"cyberpunk, friendly", "cyberpunk_age", 0.1, ally("friendly"), false},
		{"space, top faith band", "space_age", 0.1, faith(4, 4), true},
		{"space, middle band", "space_age", 0.1, faith(2, 2), false},
		{"interstellar, no route", "interstellar_age", 0.1, nil, false},
	}
	for _, c := range cases {
		ge, acct := featEngine(t, c.age, inAge(c.age, c.share))
		if c.set != nil {
			c.set(ge)
		}
		ge.mu.Lock()
		ge.judgeFeatLeaving(c.age)
		ge.mu.Unlock()
		wantFeat(t, c.name, acct, c.age, c.want)
	}
	// Modern counts the whole run so far against the summed targets.
	for _, m := range []struct {
		share float64
		want  bool
	}{{0.75, true}, {0.77, false}} {
		ge, acct := featEngine(t, "modern_age", int(m.share*through))
		ge.mu.Lock()
		ge.judgeFeatLeaving("modern_age")
		ge.mu.Unlock()
		wantFeat(t, "modern", acct, "modern_age", m.want)
	}
	// Interstellar: an unbroken run of half the time in the age, and no less.
	at := inAge("interstellar_age", 0.1)
	spent := at - featEntry
	for _, w := range []struct {
		name string
		from int
		want bool
	}{{"half the stay", at - spent/2, true}, {"a little under", at - spent/2 + 2, false}} {
		ge, acct := featEngine(t, "interstellar_age", at)
		warp(w.from)(ge)
		ge.mu.Lock()
		ge.judgeFeatLeaving("interstellar_age")
		ge.mu.Unlock()
		wantFeat(t, "interstellar, "+w.name, acct, "interstellar_age", w.want)
	}
}

// Leaving an age through advanceAge judges it with the deeds the age made, and
// the feat is held once for the run.
func TestFeatIsStruckOnAdvanceAndOnlyOnce(t *testing.T) {
	ge, acct := featEngine(t, "stone_age", inAge("stone_age", 0.3))
	advanceTo(ge, "bronze_age")
	wantFeat(t, "stone, advanced", acct, "stone_age", true)
	first := acct.CoinStackOf(config.CoinFeat, "stone_age").Copies[0]
	ge.mu.Lock()
	ge.age = "stone_age"
	ge.judgeFeatLeaving("stone_age")
	ge.mu.Unlock()
	if s := acct.CoinStackOf(config.CoinFeat, "stone_age"); s.Count != 1 || s.Copies[0] != first {
		t.Errorf("a second judging changed the stack: %+v", s)
	}
}
