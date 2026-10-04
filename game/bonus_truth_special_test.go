package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The promises that are not one effect of one config entry: Era Mastery's
// speed, the morale dial, ruins, the one-off parts of epoch events, the
// legacy kit and the retired perks. TestBonusTruth (bonus_truth_test.go)
// holds everything that is.

// near reports whether got is want within a part in a million.
func near(got, want float64) bool {
	return math.Abs(got-want) <= 1e-6*math.Max(1, math.Abs(want))
}

// TestBonusTruthEraMastery: an age with mastery m runs k = 1 + √m times as
// fast. Every resource's rate and every store are k times what they were,
// research and build times are divided by k (rounded up), and nothing else
// moves. An age six or more behind the record runs at least 4x (catch-up).
func TestBonusTruthEraMastery(t *testing.T) {
	for _, age := range []string{"stone_age", "industrial_age", "modern_age"} {
		for _, m := range []int{1, 4, 10} {
			ge := newTruthEngine(age, truthTypical)
			before := readTruth(t, ge)
			ge.Prestige.SetMastery(age, m)
			after := readTruth(t, ge)
			k := config.MasteryK(m)
			if got := ge.speedK(); got != k {
				t.Fatalf("%s at mastery %d: speed %v, want %v", age, m, got, k)
			}
			for _, meter := range sortedKeys(before) {
				want := before[meter]
				switch {
				case strings.HasPrefix(meter, "rate:"), strings.HasPrefix(meter, "storage:"):
					want *= k
				case meter == "research":
					want = math.Ceil(before[meter]*truthRefTicks/k) / truthRefTicks
				case meter == "morale":
					// Faith income lifts morale, up to a cap a tick: k times
					// the income may lift it more. Not a promise of mastery.
					continue
				}
				if !near(after[meter], want) {
					t.Errorf("%s at mastery %d (%sx): %s is %v, want %v (was %v)", age, m, SpeedText(k), meter, after[meter], want, before[meter])
				}
			}
			for _, key := range []string{"hut", "stash"} {
				def := ge.Buildings.defs[key]
				if got, want := ge.buildTicksLocked(def), max(1, int(math.Ceil(float64(def.BuildTicks)/k))); got != want {
					t.Errorf("%s at mastery %d: %s builds in %d ticks, want %d", age, m, key, got, want)
				}
			}
		}
	}
	// Catch-up: no mastery, a record six ages on.
	ge := newTruthEngine("primitive_age", truthClean)
	ge.Prestige.SetRecord(ageKeys()[config.CatchUpGap])
	if got := ge.speedK(); got != config.CatchUpK {
		t.Errorf("an age %d behind the record runs at %vx, want the catch-up %vx", config.CatchUpGap, got, config.CatchUpK)
	}
	ge.Prestige.SetRecord(ageKeys()[config.CatchUpGap-1])
	if got := ge.speedK(); got != 1 {
		t.Errorf("an age %d behind the record runs at %vx, want 1x", config.CatchUpGap-1, got)
	}
}

// TestBonusTruthMorale: morale multiplies what buildings make (the Morale
// wiki page): nothing at 50%, +20% at the cap, x0.5 at the 10% floor, and
// each wonder raises the cap by 5 points.
func TestBonusTruthMorale(t *testing.T) {
	ge := newTruthEngine("classical_age", truthClean)
	made := func() map[string]float64 {
		ge.recalculateRates()
		out := map[string]float64{}
		for _, key := range ge.Resources.order {
			if v := ge.Resources.resources[key].Breakdown.BuildingRate; v > 0 {
				out[key] = v
			}
		}
		return out
	}
	base := made()
	if len(base) == 0 {
		t.Fatal("no building output to read")
	}
	for _, c := range []struct {
		name   string
		morale float64
		want   float64
	}{
		{"neutral", moraleNeutral, 1},
		{"the cap", ge.moraleCap(), 1 + moraleMaxBonus},
		{"the floor", 0.10, moraleMinMult},
		{"halfway up", (moraleNeutral + ge.moraleCap()) / 2, 1 + moraleMaxBonus/2},
	} {
		ge.morale = c.morale
		for key, v := range made() {
			if !near(v/base[key], c.want) {
				t.Errorf("morale at %s (%v): %s is x%v of neutral, want x%v", c.name, c.morale, key, v/base[key], c.want)
			}
		}
	}
	ge.morale = moraleNeutral
	cap0 := ge.moraleCap()
	for key, def := range ge.Buildings.defs {
		if def.Category == "wonder" {
			ge.Buildings.counts[key] = 1
			break
		}
	}
	if got := ge.moraleCap(); !near(got, cap0+0.05) {
		t.Errorf("one wonder raises the morale cap from %v to %v, want +0.05", cap0, got)
	}
}

// TestBonusTruthRuins: a ruin makes half its building's listed rate with
// nobody in it (the Catastrophe wiki page).
func TestBonusTruthRuins(t *testing.T) {
	ge := newTruthEngine("iron_age", truthClean)
	for _, key := range []string{"gathering_camp", "wood_camp"} {
		def := ge.Buildings.defs[key]
		var eff config.Effect
		for _, e := range def.Effects {
			if e.Type == "production" {
				eff = e
			}
		}
		before := readTruth(t, ge)
		ge.Buildings.ruins[key] = 2
		after := readTruth(t, ge)
		delete(ge.Buildings.ruins, key)
		if got, want := after["made:"+eff.Target]-before["made:"+eff.Target], 2*eff.Value*0.5; !near(got, want) {
			t.Errorf("two %s ruins make %v %s/tick, want %v (half the listed rate each)", key, got, eff.Target, want)
		}
	}
}

// TestBonusTruthEpochEventsAtOnce: the one-off parts of the epoch events,
// as each one's own text states them. Their lasting parts (rates) are
// TestBonusTruth's; an event this test does not know fails it, so a new one
// is read here before it ships.
func TestBonusTruthEpochEventsAtOnce(t *testing.T) {
	type world struct {
		workers float64
		stock   map[string]float64
		store   map[string]float64
		techs   int
		counts  map[string]int
		built   int
	}
	look := func(ge *GameEngine) world {
		w := world{workers: float64(ge.Workers.TotalPop()), stock: map[string]float64{}, store: map[string]float64{}, techs: ge.Research.ResearchedCount(), counts: map[string]int{}}
		for _, key := range ge.Resources.order {
			w.stock[key] = ge.Resources.resources[key].Amount
			w.store[key] = ge.Resources.resources[key].Storage
		}
		for key, n := range ge.Buildings.counts {
			w.counts[key] = n
			w.built += n
		}
		return w
	}
	lost := func(t *testing.T, before, after world, res string, share float64) {
		t.Helper()
		if got, want := after.stock[res], before.stock[res]*(1-share); !near(got, want) {
			t.Errorf("%s left: %v, want %v (%v%% of %v lost)", res, got, want, share*100, before.stock[res])
		}
	}
	checks := map[string]func(t *testing.T, ge *GameEngine, before, after world){
		// "Your population grows by 15%."
		"population_surge": func(t *testing.T, ge *GameEngine, before, after world) {
			if got, want := after.workers-before.workers, math.Floor(before.workers*0.15); got != want {
				t.Errorf("workers gained: %v, want %v (15%% of %v)", got, want, before.workers)
			}
		},
		// "Every unlocked resource gains 40% of its storage."
		"ancient_cache": func(t *testing.T, ge *GameEngine, before, after world) {
			for _, key := range ge.Resources.order {
				if !ge.Resources.IsUnlocked(key) || before.store[key] <= 0 {
					continue
				}
				// The lab starts at half a store: 40% more fits.
				if got, want := after.stock[key]-before.stock[key], 0.40*before.store[key]; !near(got, want) {
					t.Errorf("%s gained %v, want %v (40%% of its storage)", key, got, want)
				}
			}
		},
		// "Up to 3 techs you can research now are completed for free."
		"grand_discovery": func(t *testing.T, ge *GameEngine, before, after world) {
			if got := after.techs - before.techs; got != 3 {
				t.Errorf("free techs: %d, want 3", got)
			}
		},
		// "You get 10 free copies of the building you have the most of."
		"architects_gift": func(t *testing.T, ge *GameEngine, before, after world) {
			if got := after.built - before.built; got != 10 {
				t.Errorf("free buildings: %d, want 10", got)
			}
		},
		// "Culture +30% and faith +20% of what you hold" (then the rates).
		"cultural_festival": func(t *testing.T, ge *GameEngine, before, after world) {
			for res, share := range map[string]float64{"culture": 0.30, "faith": 0.20} {
				if got, want := after.stock[res], before.stock[res]*(1+share); !near(got, want) {
					t.Errorf("%s: %v, want %v (+%v%% of %v)", res, got, want, share*100, before.stock[res])
				}
			}
		},
		// "Your trading partners vanish with 50% of your gold."
		"merchant_betrayal": func(t *testing.T, ge *GameEngine, before, after world) { lost(t, before, after, "gold", 0.50) },
		// "Up to 8 buildings burn down."
		"the_great_fire": func(t *testing.T, ge *GameEngine, before, after world) {
			if got := before.built - after.built; got != 8 {
				t.Errorf("buildings burned: %d, want 8", got)
			}
		},
		// "20% of your workers die."
		"epidemic": func(t *testing.T, ge *GameEngine, before, after world) {
			if got, want := before.workers-after.workers, math.Floor(before.workers*0.20); got != want {
				t.Errorf("workers lost: %v, want %v (20%% of %v)", got, want, before.workers)
			}
		},
		// "You lose 60% of your faith."
		"political_instability": func(t *testing.T, ge *GameEngine, before, after world) { lost(t, before, after, "faith", 0.60) },
		// "You lose 50% of your gold."
		"economic_crash": func(t *testing.T, ge *GameEngine, before, after world) { lost(t, before, after, "gold", 0.50) },
		// "Current research is canceled with no refund, you lose 80% of your knowledge."
		"the_dark_age": func(t *testing.T, ge *GameEngine, before, after world) {
			lost(t, before, after, "knowledge", 0.80)
			if ge.Research.currentTech != "" {
				t.Errorf("research on %s still runs", ge.Research.currentTech)
			}
		},
		// Rates only: TestBonusTruth reads them.
		"age_of_plenty": nil, "trade_winds": nil, "worker_innovation": nil, "peaceful_century": nil,
		"epoch_blessing": nil, "the_famine": nil, "resource_drought": nil,
	}
	run := func(defs []config.EpochEventDef, good bool) {
		for _, def := range defs {
			check, known := checks[def.Key]
			if !known {
				t.Errorf("epoch event %s: nothing here checks what its text promises at once (%q). Add a check, or list it as rates only.", def.Key, def.FlavorText)
				continue
			}
			if check == nil {
				continue
			}
			t.Run(def.Key, func(t *testing.T) {
				// The Classical Age of the typical player: culture and
				// faith exist, and techs of the age are left to discover.
				ge := newTruthEngine("classical_age", truthClean)
				ge.Research.currentTech, ge.Research.ticksLeft, ge.Research.totalTicks = "tool_making", 10, 10
				before := look(ge)
				if good {
					ge.applyGoodEpochEvent(def)
				} else {
					ge.applyChallengingEpochEvent(def, ge.currentEpoch)
				}
				check(t, ge, before, look(ge))
			})
		}
	}
	run(config.GoodEpochEvents(), true)
	run(config.ChallengingEpochEvents(), false)
}

// TestBonusTruthPrestigeShop: every item the shop sells is a legacy kit
// item whose promise has a test of its own, and the retired perks of the
// first shop give nothing at any tier.
func TestBonusTruthPrestigeShop(t *testing.T) {
	// Each kit item's promise is a behavior, not a rate: these tests hold it.
	held := map[string]string{
		config.LegacyPlan:     "TestPlanTemplateChainsAcrossAdvance, TestPlanTemplateCarriesPlannedResearch (legacy_test.go); the smoke suite's prestige scenario",
		config.LegacyWorkers:  "TestWorkerSharesCarryOver (legacy_test.go); the smoke suite's prestige scenario",
		config.LegacyFactions: "TestOldFriendsMetAtTheirAge (legacy_test.go); the smoke suite's prestige scenario",
	}
	pm := NewPrestigeManager()
	for _, def := range config.PrestigeUpgrades() {
		if def.Retired {
			pm.upgrades[def.Key] = def.MaxTier
			continue
		}
		if def.EffectType != "legacy" {
			t.Errorf("shop item %s has effect type %q: the bonus truth guard measures no shop bonus. Give it a meter in bonus_truth_test.go.", def.Key, def.EffectType)
			continue
		}
		if held[def.Key] == "" {
			t.Errorf("legacy kit item %s (%q) has no test named here that holds its promise", def.Key, def.Description)
		}
	}
	if b := pm.GetBonuses(); len(b) != 0 {
		t.Errorf("retired perks at their top tier still give bonuses: %v", b)
	}
	if r := pm.GetStartingResources(); len(r) != 0 {
		t.Errorf("retired perks at their top tier still give starting resources: %v", r)
	}
}

// TestBonusTruthUnknownEffect: an effect type or target the guard has no
// meter for is refused, not skipped. TestBonusTruth fails on any promise
// whose kind is empty.
func TestBonusTruthUnknownEffect(t *testing.T) {
	for _, e := range []config.Effect{
		{Type: "bonus", Target: "mystery_rate", Value: 0.1},
		{Type: "bonus", Target: "crafting_speed", Value: 0.1},
		{Type: "permanent_bonus", Target: "luck", Value: 0.1},
		{Type: "production", Target: "mana", Value: 1},
		{Type: "storage", Target: "mana", Value: 1},
		{Type: "capacity", Target: "military", Value: 10},
		{Type: "unlock", Target: "farm"},
		{Type: "mana_rate", Value: 0.1},
		{Type: "aura", Target: "all", Value: 1},
	} {
		for _, source := range []string{"tech", "milestone", "building", "event"} {
			if kind := truthEffectKind(source, e); kind != "" {
				t.Errorf("%s effect %+v gets meter %q: an unknown effect must have none, so the guard fails on it", source, e, kind)
			}
		}
	}
	// And every meter an effect can be given exists.
	lab := newTruthLab()
	for _, p := range truthPromises(t, lab) {
		if _, ok := truthKinds[p.Kind]; !ok && p.Kind != "" {
			t.Errorf("%s: kind %q has no meter in truthKinds", p.ID(), p.Kind)
		}
	}
}

// TestBonusTruthWorkerOutput: a worker output bonus raises what workers add
// to a building, and only that: an empty building gains nothing, a full one
// gains the bonus on the 80% its crew adds, and the gain does not grow with
// the other production bonuses.
func TestBonusTruthWorkerOutput(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	const key = "gathering_camp"
	def := ge.Buildings.defs[key]
	listed := def.Effects[0].Value
	pool := ge.Workers.domains["worker"]
	rate := func(crew int, gather, all float64) float64 {
		ge.Buildings.counts[key] = 1
		pool.count, pool.assignments = crew, map[string]int{key: crew}
		ge.permanentBonuses = map[string]float64{"gather_rate": gather, "production_all": all}
		ge.recalculateRates()
		return ge.Resources.resources["food"].Rate + ge.Workers.FoodDrain()
	}
	full := def.WorkerCapacity
	if got, want := rate(0, 0.5, 0)-rate(0, 0, 0), 0.0; !near(got, want) {
		t.Errorf("an empty building gains %v food/tick from +50%% worker output, want none", got)
	}
	if got, want := rate(full, 0.5, 0)-rate(full, 0, 0), 0.5*staffedShare*listed; !near(got, want) {
		t.Errorf("a full building gains %v food/tick from +50%% worker output, want %v (50%% of the %v its crew adds)", got, want, staffedShare*listed)
	}
	if got, want := rate(full, 0.5, 1)-rate(full, 0, 1), 0.5*staffedShare*listed; !near(got, want) {
		t.Errorf("with +100%% all production a full building gains %v food/tick from +50%% worker output, want the same %v", got, want)
	}
	if got, want := rate(full, -5, 0), (unstaffedShare+staffedShare*productionFloor)*listed; !near(got, want) {
		t.Errorf("a worker output penalty past the floor leaves %v food/tick, want %v (worker output at 10%%)", got, want)
	}
}

// TestBonusTruthGrantsSayWhatFit: an event's or a milestone's text states
// its full grant ("+250 food"). Storage takes only what fits, so when it
// cuts the grant short the log says what the player got; with room it says
// nothing.
func TestBonusTruthGrantsSayWhatFit(t *testing.T) {
	lastLine := func(ge *GameEngine) string { return ge.log[len(ge.log)-1].Message }
	harvest := config.EventByKey()["bountiful_harvest"] // +250 food
	reward := []config.Effect{{Type: "instant_resource", Target: "food", Value: 250}}

	// A new game holds 50 food at most, and starts with 25.
	ge := NewGameEngine()
	ge.applyEventEffects(harvest)
	if got, want := lastLine(ge), "  → Storage was nearly full: only 25 food (of 250) fit."; got != want {
		t.Errorf("a clipped event grant logs %q, want %q", got, want)
	}
	ge.applyMilestoneRewards(reward)
	if got, want := lastLine(ge), "  → Storage was nearly full: only 0 food (of 250) fit."; got != want {
		t.Errorf("a clipped milestone grant logs %q, want %q", got, want)
	}

	// With room, the text is the whole truth: no extra line.
	roomy := newTruthEngine("stone_age", truthClean)
	lines := len(roomy.log)
	roomy.applyEventEffects(harvest)
	roomy.applyMilestoneRewards(reward)
	for _, e := range roomy.log[lines:] {
		if strings.Contains(e.Message, "Storage was nearly full") {
			t.Errorf("a grant that fit logged %q", e.Message)
		}
	}
}

// TestBonusTruthAncientMemory: a tech from an ancient cache researches at
// half speed, twice its tick count, after research speed has had its say.
func TestBonusTruthAncientMemory(t *testing.T) {
	const key = "stoneworking"
	for _, speed := range []float64{0, 0.25} {
		rm := NewResearchManager()
		if err := rm.StartMemoryResearch(key, speed); err != nil {
			t.Fatal(err)
		}
		listed := rm.defs[key].ResearchTicks
		if want := 2 * int(float64(listed)*(1-speed)); rm.totalTicks != want {
			t.Errorf("an ancient memory at +%v%% research speed takes %d ticks, want %d (twice %d less the bonus)", speed*100, rm.totalTicks, want, listed)
		}
	}
}

// TestBonusTruthEventDurations: a timed event's log line says how long the
// event really lasts in the age it fired in. From the Bronze Age on every
// event runs config.PacingStretch times its typed duration; the log used to
// quote the typed one ("for ~20s" for a drought that ran 52 seconds).
func TestBonusTruthEventDurations(t *testing.T) {
	drought := config.EventByKey()["drought"] // food -0.5/tick, typed for 10 ticks
	for _, c := range []struct {
		age   string
		ticks int
		want  string
	}{
		{"stone_age", drought.Duration, "for ~20s."},
		{"iron_age", config.StretchTicks("iron_age", drought.Duration), "for ~52s."},
	} {
		ge := NewGameEngine()
		ge.SeedRNG(1)
		ge.age, ge.currentEpoch = c.age, "no epoch: universal events only"
		ge.tick = 100000
		ge.Events.defs = []config.EventDef{drought}
		ge.Events.nextEventTick = 0
		ge.processEvents()
		if len(ge.Events.active) != 1 || ge.Events.active[0].TicksLeft != c.ticks {
			t.Fatalf("%s: active events %+v, want a Drought for %d ticks", c.age, ge.Events.active, c.ticks)
		}
		line := ""
		for _, e := range ge.log {
			if strings.Contains(e.Message, "Food -0.5/tick") {
				line = e.Message
			}
		}
		if !strings.HasSuffix(line, c.want) || strings.Contains(line, "{dur}") {
			t.Errorf("%s: the log says %q, want it to end %q", c.age, line, c.want)
		}
	}
	// And no event text carries a duration of its own.
	for _, def := range append(config.RandomEvents(), config.EpochExclusiveEvents()...) {
		if strings.Contains(def.LogMessage, "for ~") {
			t.Errorf("event %s quotes its own duration (%q): say \"for {dur}\"", def.Key, def.LogMessage)
		}
	}
}

// TestBonusTruthCulturalFestivalWaitsForCulture: the Cultural Festival epoch
// event pays in culture, which is locked until the Classical Age. Entering
// the Iron Era it is never rolled; from the Renaissance Age it is.
func TestBonusTruthCulturalFestivalWaitsForCulture(t *testing.T) {
	rolled := func(age string) int {
		ge := newTruthEngine(age, truthClean)
		n := 0
		for seed := int64(1); seed <= 80; seed++ {
			ge.SeedRNG(seed)
			ge.epochEventHistory = nil
			ge.rollGoodEpochEvent()
			if len(ge.epochEventHistory) == 1 && ge.epochEventHistory[0].EventKey == "cultural_festival" {
				n++
			}
		}
		return n
	}
	if n := rolled("iron_age"); n != 0 {
		t.Errorf("the Cultural Festival was rolled %d times of 80 in the Iron Age, where culture is locked", n)
	}
	if n := rolled("renaissance_age"); n == 0 {
		t.Error("the Cultural Festival was never rolled in 80 tries in the Renaissance Age")
	}
}
