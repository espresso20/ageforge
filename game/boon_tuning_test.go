package game

import (
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Measurement harness for the faction-encounter BOON loop, and the one gate that
// rests on it: TestBoonTuning_AutomationSitsBetweenCasualAndActive.
//
// boon_soak_test.go asks "is the system BOUNDED?". This file asks whether the
// engagement gradient holds: how often a hands-on player, a casual player and the
// Geographic Society meet a faction, and how much of the time each holds a boon.
//
// What it simulates, per scenario:
//   - the real expedition cadence for that age — the SHORTEST available
//     expedition per category (the spam-optimal choice a continuous explorer
//     actually makes), re-launched the moment it resolves (or after a gap, for
//     the casual scenario), with each launch's length rolled from the mission
//     kind's [MinMinutes, MaxMinutes] range exactly as the runtime rolls it;
//   - the real success roll (successRoll > DifficultyBase, no military bonus);
//   - the real encounter path: rollExpeditionEncounter → applyFactionBoon /
//     applyFactionMalus → boon.RollBoon → boon.Apply → InjectEvent;
//   - real per-tick expiry via EventManager.Tick.
//
// The economy is deliberately NOT simulated (no building production, no
// expedition rewards, no launch costs): stockpiles are pre-filled and rates left
// alone so the boon engine is the only thing moving.

const (
	// tuningTicks is the driven-tick budget per scenario. The real cadence is slow
	// (two categories, each re-launching a ~60-160 tick expedition), so 2M ticks is
	// a few thousand expedition resolutions per scenario: enough for the rates the
	// gate reads to be stable. Costs a second.
	tuningTicks = 2_000_000
	// tuningStockFill is the amount every resource is pre-loaded with (and the
	// storage cap raised to) so a ResourceDrain has something to bite and an
	// InstantResource grant is not silently clamped away.
	tuningStockFill = config.Billion
	// tuningStartPop is the starting worker head-count, big enough that repeated
	// WorkerLoss maluses cannot floor the population.
	tuningStartPop = 5000
)

// --- scenario definition ----------------------------------------------------

// tuningStanding is the diplomatic state a scenario installs for one faction.
type tuningStanding struct {
	status string
	atWar  bool
}

// tuningScenario describes one representative style of play.
type tuningScenario struct {
	name string
	age  string
	// seed seeds the engine and the expedition driver. Each scenario keeps the seed
	// it had when the harness ran more scenarios (0xB00_1234 + its old index), so
	// the numbers the gate reads did not move when the unused ones were deleted.
	seed int64
	// standings keys factions by roster key. Any eligible faction absent from
	// the map is left UNDISCOVERED (so the run opens with discovery encounters).
	standings map[string]tuningStanding
	// scoutGap / milGap are the idle ticks between one expedition resolving and
	// the next launching. 0 = back-to-back. Negative = that category is unused.
	scoutGap int
	milGap   int
	// autoSocieties / autoWorkers describe an AUTOMATED scenario: a player who
	// launches nothing by hand and instead has this many Geographic Societies with
	// this many workers assigned across them (see game/auto_expedition.go).
	// autoSocieties > 0 makes the scouting driver run on the shipped auto-dispatch
	// cadence instead of scoutGap, so the tuning constants are what is measured.
	autoSocieties int
	autoWorkers   int
}

func tuningScenarios() []tuningScenario {
	mid := map[string]tuningStanding{
		"riverlands_tribes":  {status: "allied"},
		"ironhold_clans":     {status: "friendly"},
		"merchant_guild":     {status: "neutral"},
		"artisan_league":     {status: "friendly"},
		"atomic_directorate": {status: "neutral"},
	}

	return []tuningScenario{
		{
			name:      "B mid-game (atomic, 5 civs, mixed standing)",
			age:       "atomic_age",
			seed:      0xB00_1234 + 1,
			standings: mid,
			scoutGap:  0,
			milGap:    0,
		},
		{
			name:      "D casual (atomic, 5 civs, scout every ~1000 ticks)",
			age:       "atomic_age",
			seed:      0xB00_1234 + 3,
			standings: mid,
			scoutGap:  1000,
			milGap:    -1,
		},
		// F/F2 are the AUTOMATED idler: nothing launched by hand at all, every
		// expedition dispatched by the Geographic Society. Same age and standings as
		// B/D so they can be read straight off against hands-on and casual play.
		//
		// This is the measurement the automation feature has to pass. F must land
		// BETWEEN D (casual) and B (hands-on), at roughly 50-70% of B's encounter
		// rate — see TestBoonTuning_AutomationSitsBetweenCasualAndActive, which
		// asserts it. milGap is -1 on both: automation never touches the military
		// category, which is a large part of why it stays behind hands-on play.
		{
			name:          "F automated idler (atomic, 6 societies fully staffed)",
			age:           "atomic_age",
			seed:          0xB00_1234 + 5,
			standings:     mid,
			scoutGap:      0,
			milGap:        -1,
			autoSocieties: 6,
			autoWorkers:   48,
		},
		{
			name:          "F2 automated idler, part-invested (atomic, 2 societies, half staffed)",
			age:           "atomic_age",
			seed:          0xB00_1234 + 6,
			standings:     mid,
			scoutGap:      0,
			milGap:        -1,
			autoSocieties: 2,
			autoWorkers:   8,
		},
	}
}

// --- measured output --------------------------------------------------------

// tuningStats is what the gate reads from one scenario run: how many encounters
// fired and on how many ticks at least one faction boon was live.
type tuningStats struct {
	ticks        int
	encounters   int
	ticksAnyBoon int
}

func (s *tuningStats) per10k(n int) float64 {
	if s.ticks == 0 {
		return 0
	}
	return float64(n) * 10000 / float64(s.ticks)
}

func (s *tuningStats) pctOfTicks(n int) float64 {
	if s.ticks == 0 {
		return 0
	}
	return float64(n) * 100 / float64(s.ticks)
}

// --- the driver -------------------------------------------------------------

// expDriver models one expedition category's launch/resolve loop, INCLUDING the
// randomized per-launch duration the real game rolls.
type expDriver struct {
	durMin     int // randomized range low end, as typed (base curve)
	durMax     int
	difficulty float64
	gap        int
	ticksLeft  int
	cooldown   int
	category   string
	live       bool
	// autoInterval > 0 puts the driver on the Geographic Society's automatic
	// cadence instead of a fixed gap (see idleGap). lastDur is the duration of the
	// launch currently in flight, which the automatic cadence needs because its
	// countdown runs from the LAUNCH, not from the return.
	autoInterval int
	lastDur      int
}

// harnessDef is a kind of mission as the harness models it: the range of ticks
// an offer of the kind rolls its length over, and how hard it is. The board
// rolls those when an offer is made; a player who spams the shortest offer
// sees the low end of the range.
type harnessDef struct {
	DurationMin    int
	DurationMax    int
	DifficultyBase float64
}

// harnessDefs are the kinds of mission a category sends: short and long
// missions are scouting, the grand one military.
func harnessDefs(category string) []harnessDef {
	var out []harnessDef
	for _, kind := range config.MissionKinds() {
		cat := ExpeditionScouting
		if kind == config.MissionCampaign {
			cat = ExpeditionMilitary
		}
		if cat != category {
			continue
		}
		t, _ := config.MissionKindTuning(kind)
		out = append(out, harnessDef{minutesToTicks(t.MinMinutes), minutesToTicks(t.MaxMinutes), 1 - t.BaseSuccess})
	}
	return out
}

// rollDuration draws one launch's active duration with the SAME rule
// LaunchExpedition uses — a uniform value in [DurationMin, DurationMax],
// not stretched by the age — but off the harness's seeded rng so a scenario stays
// reproducible. Mirrors the runtime's guards: a def without a real range pins at
// its floor, and the result is floored at 1 so a malformed def cannot spin the
// driver.
func (d *expDriver) rollDuration(rng *rand.Rand) int {
	ticks := d.durMin
	if d.durMax > d.durMin {
		ticks = d.durMin + rng.Intn(d.durMax-d.durMin+1)
	}
	if ticks < 1 {
		ticks = 1
	}
	return ticks // a mission's length is rolled in minutes and no age stretches it
}

// newExpDriver picks the SHORTEST expedition available in a category — the
// spam-optimal choice, and therefore the honest worst case for encounter
// cadence. "Shortest" is compared on the low end of the randomized duration
// range. Returns live=false when the category has nothing available or the
// scenario disabled it (gap < 0).
func newExpDriver(category string, gap int, rng *rand.Rand) expDriver {
	if gap < 0 {
		return expDriver{category: category}
	}
	defs := harnessDefs(category)
	if len(defs) == 0 {
		return expDriver{category: category}
	}
	best := defs[0]
	for _, d := range defs[1:] {
		if d.DurationMin < best.DurationMin {
			best = d
		}
	}
	d := expDriver{
		durMin:     best.DurationMin,
		durMax:     best.DurationMax,
		difficulty: best.DifficultyBase,
		gap:        gap,
		category:   category,
		live:       true,
	}
	d.relaunch(rng)
	return d
}

// newAutoScoutDriver builds the scouting driver for an AUTOMATED scenario. Unlike
// newExpDriver it does not guess at the cadence or the expedition: it asks the
// shipped code for both — pickAutoOffer for which party goes out and
// autoExpeditionIntervalIn for how often — so what the harness reports is the
// behaviour of the live tuning constants, not a re-implementation of them.
// Caller must hold the engine lock.
func newAutoScoutDriver(ge *GameEngine, rng *rand.Rand) expDriver {
	ge.ensureMissionBoardLocked()
	n := ge.pickAutoOffer()
	if n == 0 {
		return expDriver{category: ExpeditionScouting}
	}
	offer, _ := ge.Military.BoardOffer(n)
	count, fill := ge.autoExpeditionInvestment()
	interval := autoExpeditionIntervalIn(ge.age, count, fill)
	if interval <= 0 {
		return expDriver{category: ExpeditionScouting}
	}
	// The party is the tier the society picked, its length rolled over the
	// tier's range at every launch (a new offer is made each time).
	tuning, _ := config.MissionKindTuning(offer.Kind())
	d := expDriver{
		durMin:       minutesToTicks(tuning.MinMinutes),
		durMax:       minutesToTicks(tuning.MaxMinutes),
		difficulty:   1 - offer.SuccessChance,
		category:     offer.Category(),
		autoInterval: interval,
		live:         true,
	}
	d.relaunch(rng)
	return d
}

// relaunch puts a fresh expedition in flight, remembering its rolled duration.
func (d *expDriver) relaunch(rng *rand.Rand) {
	d.lastDur = d.rollDuration(rng)
	d.ticksLeft = d.lastDur
}

// idleGap is how long the driver waits after a resolution before launching again.
//
// Manual play waits the scenario's fixed gap. AUTOMATED play instead reproduces
// the real dispatcher's cycle of max(interval, duration): the society's countdown
// runs from the launch and the party can only leave when the single scouting slot
// is free, so whatever is left of the countdown when the last party comes home is
// the idle time — and nothing at all if the trip outlasted the countdown.
func (d *expDriver) idleGap() int {
	if d.autoInterval > 0 {
		return d.autoInterval - d.lastDur
	}
	return d.gap
}

// step advances one tick. resolved reports whether the expedition finished this
// tick; success carries its outcome.
func (d *expDriver) step(rng *rand.Rand) (resolved, success bool) {
	if !d.live {
		return false, false
	}
	if d.ticksLeft > 0 {
		d.ticksLeft--
		if d.ticksLeft > 0 {
			return false, false
		}
		// Same success rule as MilitaryManager.tickCategory with no military
		// bonus: successRoll > DifficultyBase (floored at 0.05).
		diff := math.Max(d.difficulty, 0.05)
		d.cooldown = d.idleGap()
		if d.cooldown <= 0 {
			d.relaunch(rng)
		}
		return true, rng.Float64() > diff
	}
	d.cooldown--
	if d.cooldown <= 0 {
		d.relaunch(rng)
	}
	return false, false
}

// prepareTuningEngine builds an engine parked at a scenario's age with its
// standings installed, stockpiles filled, a worker pool seeded, and random
// events suppressed. Caller must hold no lock; the returned engine is left
// UNLOCKED.
func prepareTuningEngine(sc tuningScenario) *GameEngine {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()

	ge.SeedRNG(sc.seed)
	ge.age = sc.age
	ge.currentEpoch = config.EpochForAge(ge.age)
	ge.Events.nextEventTick = 1 << 40 // suppress random events: active list is faction-only

	for key, st := range sc.standings {
		ge.Diplomacy.factions[key] = &FactionState{
			Discovered: true,
			Opinion:    100,
			Status:     st.status,
			AtWar:      st.atWar,
		}
	}

	for _, def := range config.BaseResources() {
		ge.Resources.UnlockResource(def.Key)
		if r, ok := ge.Resources.resources[def.Key]; ok {
			r.Amount = tuningStockFill
			r.Rate = 0
		}
	}
	ge.Workers.AddLentWorkers(tuningStartPop)
	// Every mission takes a party: the scenario's town has an army at home
	// that covers any number of them.
	ge.garrisonForTest, ge.garrisonForTestSet = config.Trillion, true
	ge.syncSoldiers()

	// Automated scenarios stand up the Geographic Societies and staff them, so the
	// shipped investment/cadence functions see a real installation to read.
	if sc.autoSocieties > 0 {
		ge.Buildings.LoadCounts(map[string]int{autoExpeditionBuildingKey: sc.autoSocieties})
		if sc.autoWorkers > 0 && !ge.Workers.Assign(autoExpeditionBuildingKey, sc.autoWorkers) {
			panic("tuning scenario could not staff its Geographic Societies")
		}
	}

	return ge
}

// runTuningScenario drives one scenario for tuningTicks and returns what the gate
// reads: the encounter count and the ticks on which a faction boon was live. An
// expedition resolution that yields any player-facing line is one encounter.
func runTuningScenario(sc tuningScenario) *tuningStats {
	ge := prepareTuningEngine(sc)
	stats := &tuningStats{ticks: tuningTicks}

	expRNG := rand.New(rand.NewSource(sc.seed ^ 0x5EED))

	ge.mu.Lock()
	defer ge.mu.Unlock()

	order := ge.progress.GetAgeOrder()
	drivers := []*expDriver{}
	if sc.autoSocieties > 0 {
		// Automated: the scouting driver runs on the shipped dispatch cadence and
		// the player launches nothing by hand.
		if d := newAutoScoutDriver(ge, expRNG); d.live {
			drivers = append(drivers, &d)
		}
	} else if d := newExpDriver(ExpeditionScouting, sc.scoutGap, expRNG); d.live {
		drivers = append(drivers, &d)
	}
	if d := newExpDriver(ExpeditionMilitary, sc.milGap, expRNG); d.live {
		drivers = append(drivers, &d)
	}

	for tick := 0; tick < tuningTicks; tick++ {
		for _, d := range drivers {
			resolved, success := d.step(expRNG)
			if !resolved {
				continue
			}
			if len(ge.rollExpeditionEncounter(d.category, success)) > 0 {
				stats.encounters++
			}
		}

		ge.Events.Tick(ge.gameRNG(), tick, ge.age, order, ge.currentEpoch)

		if ge.activeFactionBoonCount() > 0 {
			stats.ticksAnyBoon++
		}
	}

	return stats
}

// --- the gate ----------------------------------------------------------------

// TestBoonTuning_AutomationSitsBetweenCasualAndActive is the GATE on the
// engagement gradient — the one assertion in this file that is not a loose sanity
// check.
//
// The tuning pass that preceded automation deliberately spread play across a wide
// band: a player exploring continuously meets someone ~26 times per 10k ticks and
// holds a boon 97.5% of the time; a player who remembers to send a scout every
// ~1000 ticks meets someone ~1.4 times and holds a boon 15.8% of the time. The
// Geographic Society must land BETWEEN those, and at full investment must sit near
// 50-70% of hands-on play. If automation ever matched back-to-back manual
// dispatching, playing by hand would be pointless and that whole band would
// collapse into a single point.
//
// The lever is autoExpeditionMinInterval (game/auto_expedition.go). If this test
// fails, that constant is what to move — up to slow automation down, down to speed
// it up — and then re-read the GRADIENT block this test logs.
func TestBoonTuning_AutomationSitsBetweenCasualAndActive(t *testing.T) {
	if testing.Short() {
		t.Skip("needs the full tick budget for stable rates")
	}

	// Same age and standings across all four so the only variable is HOW the
	// expeditions get dispatched.
	byName := map[string]*tuningStats{}
	for _, sc := range tuningScenarios() {
		key := strings.Fields(sc.name)[0]
		byName[key] = runTuningScenario(sc)
	}

	active, ok := byName["B"]
	if !ok {
		t.Fatal("scenario B (hands-on, atomic) is missing — the gate has no reference point")
	}
	casual, ok := byName["D"]
	if !ok {
		t.Fatal("scenario D (casual, atomic) is missing")
	}
	auto, ok := byName["F"]
	if !ok {
		t.Fatal("scenario F (automated idler, atomic) is missing")
	}
	partial, ok := byName["F2"]
	if !ok {
		t.Fatal("scenario F2 (part-invested automation, atomic) is missing")
	}

	rate := func(s *tuningStats) float64 { return s.per10k(s.encounters) }
	uptime := func(s *tuningStats) float64 { return s.pctOfTicks(s.ticksAnyBoon) }

	share := rate(auto) / rate(active) * 100
	t.Logf("\nGRADIENT (atomic age, identical standings)\n"+
		"  casual  (D)              %6.1f enc/10k   %5.1f%% uptime\n"+
		"  auto part-invested (F2)  %6.1f enc/10k   %5.1f%% uptime\n"+
		"  auto FULL (F)            %6.1f enc/10k   %5.1f%% uptime   <- %.1f%% of hands-on\n"+
		"  hands-on (B)             %6.1f enc/10k   %5.1f%% uptime\n",
		rate(casual), uptime(casual),
		rate(partial), uptime(partial),
		rate(auto), uptime(auto), share,
		rate(active), uptime(active))

	// --- 1. Full automation lands strictly between casual and hands-on. ---
	if rate(auto) <= rate(casual) {
		t.Errorf("full automation (%.1f enc/10k) is no better than casual play (%.1f) — "+
			"the Geographic Society is not worth building", rate(auto), rate(casual))
	}
	if rate(auto) >= rate(active) {
		t.Errorf("full automation (%.1f enc/10k) matches or beats hands-on play (%.1f) — "+
			"raise autoExpeditionMinInterval", rate(auto), rate(active))
	}
	if uptime(auto) <= uptime(casual) || uptime(auto) >= uptime(active) {
		t.Errorf("automation boon uptime %.1f%% is outside the casual..hands-on band (%.1f%%..%.1f%%)",
			uptime(auto), uptime(casual), uptime(active))
	}

	// --- 2. And it lands in the intended 50-70% window. A little slack either
	// side: the expedition duration roll and the encounter roll both float. ---
	const wantLo, wantHi = 48.0, 72.0
	if share < wantLo || share > wantHi {
		t.Errorf("full automation runs at %.1f%% of the hands-on encounter rate, want %.0f-%.0f%% — "+
			"move autoExpeditionMinInterval (currently %d ticks; the share moves inversely with it)",
			share, wantLo, wantHi, autoExpeditionMinInterval)
	}

	// --- 3. Investment is a real gradient of its own: a part-invested society is
	// better than nothing and worse than a full installation. ---
	if rate(partial) <= rate(casual) || rate(partial) >= rate(auto) {
		t.Errorf("part-invested automation (%.1f enc/10k) is not between casual (%.1f) and "+
			"full investment (%.1f) — the count/staffing terms are not doing their job",
			rate(partial), rate(casual), rate(auto))
	}
}
