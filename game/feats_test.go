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
		through += rules.Core().CoinRefTicks(a, 1)
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

// wonderKeyOf is a wonder of age.
func wonderKeyOf(t *testing.T, age string) string {
	t.Helper()
	for _, b := range rules.Core().Buildings() {
		if b.Category == "wonder" && b.RequiredAge == age {
			return b.Key
		}
	}
	t.Fatalf("no wonder in %s", age)
	return ""
}

// atWar starts a war with a civilization the run has met.
func atWar(ge *GameEngine, key string) {
	ge.Diplomacy.DiscoverFaction(key)
	ge.Diplomacy.factions[key].AtWar = true
}

// The feats of the ten later ages strike on the event that meets them and not on its near miss.
func TestFeatsOfTheLaterAgesDecidedByAnEvent(t *testing.T) {
	wonder := func(age string) Event {
		return Event{Kind: config.BadgeEvWonderRaised, Subject: wonderKeyOf(t, age)}
	}
	lp := func(appease, brace int, invited bool) func(*GameEngine) {
		return func(ge *GameEngine) {
			ge.harbinger = &HarbingerSave{AppeaseLevel: appease, BraceLevel: brace, Invited: invited}
		}
	}
	allied := func(met, unallied float64) Event {
		return Event{Kind: config.BadgeEvCivAllied, Attrs: map[string]float64{"met": met, "unallied": unallied}}
	}
	cases := []struct {
		name, age string
		share     float64 // of the age's target spent in it
		set       func(ge *GameEngine)
		ev        Event
		want      bool
	}{
		{"medieval, a war seen through", "medieval_age", 0.1, func(ge *GameEngine) { ge.Stats.Deeds.WarBegan("f", featEntry) }, Event{Kind: config.BadgeEvWarEnded, Subject: "f"}, true},
		{"medieval, ended by tribute", "medieval_age", 0.1, func(ge *GameEngine) { ge.Stats.Deeds.WarBegan("f", featEntry); ge.Stats.Deeds.Tributes = 1 },
			Event{Kind: config.BadgeEvWarEnded, Subject: "f", Attrs: map[string]float64{"tribute": 1}}, false},
		{"medieval, an earlier tribute", "medieval_age", 0.1, func(ge *GameEngine) { ge.Stats.Deeds.WarBegan("f", featEntry); ge.Stats.Deeds.Tributes = 1 }, Event{Kind: config.BadgeEvWarEnded, Subject: "f"}, false},
		{"medieval, a war from before the age", "medieval_age", 0.1, nil, Event{Kind: config.BadgeEvWarEnded, Subject: "f"}, false},
		{"victorian, in the first sixth", "victorian_age", 1.0/6 - 0.01, nil, wonder("victorian_age"), true},
		{"victorian, just past a sixth", "victorian_age", 1.0/6 + 0.01, nil, wonder("victorian_age"), false},
		{"victorian, another age's wonder", "victorian_age", 0.05, nil, wonder("colonial_age"), false},
		{"scholar chain finished", "information_age", 0.1, nil, Event{Kind: config.BadgeEvMilestone, Subject: "tech_master"}, true},
		{"another milestone", "information_age", 0.1, nil, Event{Kind: config.BadgeEvMilestone, Subject: "tech_pioneer"}, false},
		{"digital, wonder at war", "digital_age", 0.1, func(ge *GameEngine) { atWar(ge, "shadow_syndicate") }, wonder("digital_age"), true},
		{"digital, wonder at peace", "digital_age", 0.1, nil, wonder("digital_age"), false},
		{"fusion, eight allies at peace", "fusion_age", 0.1, nil, allied(10, 2), true},
		{"fusion, seven allies", "fusion_age", 0.1, nil, allied(10, 3), false},
		{"fusion, a war running", "fusion_age", 0.1, func(ge *GameEngine) { atWar(ge, "shadow_syndicate") }, allied(10, 2), false},
		{"galactic, appease 2", "galactic_age", 0.1, lp(2, 0, false), Event{Kind: config.BadgeEvAppeased}, true},
		{"galactic, appease 1", "galactic_age", 0.1, lp(1, 0, false), Event{Kind: config.BadgeEvAppeased}, false},
		{"galactic, a doom's thread", "galactic_age", 0.1, func(ge *GameEngine) { ge.harbinger = &HarbingerSave{TargetEpoch: "cosmic_era", AppeaseLevel: 2} }, Event{Kind: config.BadgeEvAppeased}, false},
		{"quantum, brace second", "quantum_age", 0.1, lp(0, 2, true), Event{Kind: config.BadgeEvBraced}, true},
		{"quantum, invite second", "quantum_age", 0.1, lp(0, 2, true), Event{Kind: config.BadgeEvInvited}, true},
		{"quantum, braced only", "quantum_age", 0.1, lp(0, 2, false), Event{Kind: config.BadgeEvBraced}, false},
		{"quantum, invited at brace 1", "quantum_age", 0.1, lp(0, 1, true), Event{Kind: config.BadgeEvInvited}, false},
		{"transcendent, endured at brace 2", "transcendent_age", 0.1, nil, Event{Kind: config.BadgeEvLastPassage, Subject: "endured", Attrs: map[string]float64{"brace": 2}}, true},
		{"transcendent, endured at brace 1", "transcendent_age", 0.1, nil, Event{Kind: config.BadgeEvLastPassage, Subject: "endured", Attrs: map[string]float64{"brace": 1}}, false},
		{"transcendent, spared", "transcendent_age", 0.1, nil, Event{Kind: config.BadgeEvLastPassage, Subject: "spared", Attrs: map[string]float64{"brace": 2}}, false},
	}
	for _, c := range cases {
		ge, acct := featEngine(t, c.age, inAge(c.age, c.share))
		ge.mu.Lock()
		if c.set != nil {
			c.set(ge)
		}
		ge.report(c.ev)
		ge.mu.Unlock()
		wantFeat(t, c.name, acct, c.age, c.want)
	}
}

// A war's start is recorded where wars are told, and a tribute where it is paid.
func TestWarDeedsAreCountedInTheAge(t *testing.T) {
	ge, acct := featEngine(t, "medieval_age", inAge("medieval_age", 0.1))
	ge.mu.Lock()
	ge.Diplomacy.warEvents = []warEvent{{faction: "f", started: true, wars: 1}}
	ge.noteWars(false)
	began := ge.Stats.Deeds.Wars["f"]
	ge.Diplomacy.warEvents = []warEvent{{faction: "f"}}
	ge.noteWars(false)
	ge.mu.Unlock()
	if began != ge.tick {
		t.Errorf("war began at %d, want %d", began, ge.tick)
	}
	wantFeat(t, "medieval, war told start to end", acct, "medieval_age", true)
	ge.mu.Lock()
	ge.advanceAge("renaissance_age")
	d := ge.Stats.Deeds
	ge.mu.Unlock()
	if d.Wars != nil || d.Tributes != 0 {
		t.Errorf("deeds not zeroed on entering an age: %+v", d)
	}
}

// Colonial: nine routes with an unbroken run, judged on the tick.
func TestColonialFeatNeedsNineRoutes(t *testing.T) {
	for _, c := range []struct {
		n    int
		want bool
	}{{9, true}, {8, false}} {
		ge, acct := featEngine(t, "colonial_age", inAge("colonial_age", 0.1))
		ge.mu.Lock()
		ge.Stats.Deeds.RouteClean = map[string]int{}
		for i := 0; i < c.n; i++ {
			ge.Stats.Deeds.RouteClean[string(rune('a'+i))] = featEntry
		}
		ge.report(Event{Kind: config.BadgeEvTick})
		ge.mu.Unlock()
		wantFeat(t, "colonial routes", acct, "colonial_age", c.want)
	}
	// The age has nine routes to run: every route open by the Colonial Age.
	n := 0
	for _, r := range config.BaseTradeRoutes() {
		if p, ok := rules.Core().Index(r.MinAge); ok {
			if c, _ := rules.Core().Index("colonial_age"); p <= c {
				n++
			}
		}
	}
	if n != int(rules.Core().FeatMark("colonial_age")) {
		t.Errorf("%d routes open by the Colonial Age, the feat asks for %d", n, int(rules.Core().FeatMark("colonial_age")))
	}
}

// Industrial: a third of the posted workers in energy jobs, on the census tick.
func TestIndustrialFeatNeedsAThirdInEnergy(t *testing.T) {
	if def, ok := rules.Core().Building("coal_plant"); !ok || def.WorkerDomain != "energy" || def.WorkerCapacity == 0 {
		t.Fatalf("the Industrial Age cannot staff energy: %+v", def)
	}
	for _, c := range []struct {
		energy, other int
		tick          int
		want          bool
	}{{34, 66, 0, true}, {33, 67, 0, false}, {1, 2, 0, true}, {34, 66, 1, false}} {
		ge, acct := featEngine(t, "industrial_age", 0)
		ge.mu.Lock()
		ge.tick = inAge("industrial_age", 0.1)
		ge.tick += config.BadgeCensusTicks - ge.tick%config.BadgeCensusTicks + c.tick
		ge.Workers.assignments = map[string]int{"coal_plant": c.energy, "farm": c.other}
		ge.report(Event{Kind: config.BadgeEvTick})
		ge.mu.Unlock()
		wantFeat(t, "industrial share", acct, "industrial_age", c.want)
	}
}

// Transcendent: endure is reported inside the prestige, while the run is still in its age.
func TestTranscendentFeatIsStruckByAPrestigeThatEndures(t *testing.T) {
	for _, c := range []struct {
		brace int
		want  bool
	}{{2, true}, {1, false}} {
		ge, acct := featEngine(t, "transcendent_age", inAge("transcendent_age", 0.1))
		ge.mu.Lock()
		ge.harbinger = &HarbingerSave{BraceLevel: c.brace, Invited: true}
		ge.completePrestige(lastPassageEndured)
		ge.mu.Unlock()
		wantFeat(t, "transcendent endure", acct, "transcendent_age", c.want)
	}
}
