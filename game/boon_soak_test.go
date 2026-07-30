package game

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Long-run robustness / soak coverage for the faction BOON system (package boon
// wired through faction_boon.go → encounters.go → the event/resolver machinery).
//
// The design of the boon path is "inject a fresh timed ActiveEvent per encounter"
// (see faction_boon.go boonApplier.InjectTimedEffects → EventManager.InjectEvent).
// InjectEvent APPENDS with no dedup by key, so a raw double-inject still STACKS.
// What bounds the system is no longer expiry alone: rollExpeditionEncounter now
// gates on maxConcurrentFactionBoons (and maxConcurrentFactionMaluses), so an
// encounter arriving at capacity grants nothing positive. On top of that,
// recalculateRates clamps the applied multiplier to productionCap, so even a
// pathological pool cannot exceed x3.0. tick_speed keeps its own clamp — the real
// MinTickInterval floor in getTickInterval.
//
// These tests drive the REAL path (rollExpeditionEncounter → applyFactionBoon →
// boon.RollBoon → boon.Apply → InjectEvent) at the fastest cadence the expedition
// system can actually sustain, across several prestige resets, with every faction
// discovered + allied (the worst-case magnitude/rarity profile). They assert the
// invariants a healthy long run must keep:
//
//   - no panic over millions of driven ticks;
//   - concurrent active boons stay bounded BY THE CAPACITY GATE, not by luck;
//   - the summed production_all / <res>_rate additive pools stay under a hard
//     ceiling, and the multiplier the engine actually applies never exceeds
//     productionCap;
//   - the tick interval never dips below MinTickInterval (code-enforced clamp);
//   - no NaN/Inf leaks into resource rates or the computed multipliers;
//   - prestige/Reset clears every active boon (no cross-run leak);
//   - the same seed + same driven sequence reproduces byte-identical end state.
//
// UPDATED (capacity pass): the ceilings below are now backed by REAL production
// clamps — maxConcurrentFactionBoons / maxConcurrentFactionMaluses in the
// encounter path and productionCap in recalculateRates. The previous run measured
// 238 concurrent boons and a x20.3 knowledge_rate multiplier against ceilings of
// 1024 / 50.0 that existed only to catch a runaway. Those ceilings are now set
// just above the enforced bound, so a regression that removes either clamp fails
// immediately rather than after it has already ruined balance.

const (
	// soakPrestigeCycles is how many prestige/Reset boundaries the soak crosses.
	soakPrestigeCycles = 3
	// soakTicksPerCycle is the driven-tick count per cycle. 3 x 400k = 1.2M ticks
	// of real expiry bookkeeping and ~240k encounter attempts — the "millions of
	// ticks / ~500 hours of play" long-run regime. Concurrency saturates within the
	// first ~3k ticks (the longest boon duration), so the rest re-confirms the
	// plateau holds rather than climbing.
	soakTicksPerCycle = 400_000
	// soakEncounterEveryTicks deliberately OVER-drives the encounter path: one
	// attempt every 5 ticks, against a real game where the shortest scouting run
	// takes 60-100 ticks and only a fraction of resolutions produce an encounter at
	// all (see game/boon_tuning_test.go, which measures ~27 encounters per 10k
	// ticks for a continuous explorer — roughly 70x slower than this).
	//
	// That is the point. A soak asks whether the bounds HOLD under load the game
	// cannot actually generate; it is not a cadence measurement and must not be
	// read as one. Automatic dispatch (the Geographic Society, see
	// game/auto_expedition.go) does raise SUSTAINED encounter pressure — an idler
	// now keeps a scouting party in the field indefinitely instead of only while
	// they are at the keyboard — but it does so at one resolution per ~100 ticks and
	// one encounter per ~640, i.e. ~35x below what this driver already sustains, and
	// through the same rollExpeditionEncounter gate. The bounds below are unaffected. This constant was once justified by a "shortest duration is ~10
	// ticks" claim taken from a legacy fixed Duration field on ExpeditionDef, which
	// the runtime ignored (the field has since been deleted) — the number was wrong,
	// the pressure it applies is still valid, so the driver stays and the reasoning
	// is corrected. Each attempt is
	// still gated by the real per-resolution encounter probability inside
	// rollExpeditionEncounter.
	soakEncounterEveryTicks = 5

	// soakSampleEvery / soakRecalcEvery keep the hot loop cheap: bound-check the
	// resolver pools every N ticks and run a full recalculateRates NaN sweep less
	// often. Both are frequent enough to catch a monotonic runaway.
	soakSampleEvery = 200
	soakRecalcEvery = 1000

	// --- Invariant ceilings (now BACKED by production clamps) -------------------
	// Concurrency is gated in rollExpeditionEncounter: a positive boon is refused
	// at maxConcurrentFactionBoons, a timed setback at maxConcurrentFactionMaluses.
	// Each encounter injects at most one event and the gate is checked BEFORE
	// injecting, so the true bound is boons+maluses. The +2 slack absorbs any
	// non-faction active event the driven path might produce without letting a
	// regression (removed gate ⇒ hundreds) slip through.
	maxConcurrentBoons = maxConcurrentFactionBoons + maxConcurrentFactionMaluses + 2

	// maxAdditivePool bounds the raw additive Σ for any production_all or
	// <res>_rate pool. With at most maxConcurrentFactionBoons (5, after the
	// measured tuning pass) concurrent boons at the worst-case allied str-5
	// magnitude (Enlightenment 0.25 x ~1.82 = 0.455 each) the realistic worst
	// case is ~2.3 — still under this ceiling, which is the reason the capacity
	// bump stopped at 5. 3.0 mirrors productionCap: past it the engine's clamp
	// is doing all the work, which is exactly the state this test exists to
	// prevent from creeping back.
	maxAdditivePool = 3.0
)

// allyAllFactionsDiscovered marks every roster faction discovered AND allied, the
// worst-case standing for magnitude (x1.40) and rarity. Mirrors how the diplomacy
// manager stores state internally (see allyFaction in active_effects_test.go).
// Caller must hold the write lock (or run before the engine ticks).
func allyAllFactionsDiscovered(ge *GameEngine) {
	for _, def := range config.BaseFactions() {
		ge.Diplomacy.factions[def.Key] = &FactionState{
			Discovered: true,
			Opinion:    100,
			Status:     "allied",
		}
	}
}

// lastAgeKey returns the final age in the canonical order — high enough that every
// faction's MinAge floor is met, so all 11 are eligible to be re-encountered.
func lastAgeKey() string {
	order := config.AgeOrder()
	return order[len(order)-1]
}

// hotRateTargets is the set of <res>_rate pools most likely to be piled on: every
// faction specialty plus knowledge (fed by Enlightenment from every faction).
func hotRateTargets() []string {
	seen := map[string]bool{}
	var out []string
	add := func(res string) {
		key := res + "_rate"
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	add("knowledge")
	for _, def := range config.BaseFactions() {
		if def.Specialty != "" {
			add(def.Specialty)
		}
	}
	return out
}

// soakObservations records the worst values seen across a driven run.
type soakObservations struct {
	maxConcurrent int
	maxPool       float64
	maxPoolTarget string
	minInterval   float64 // milliseconds; +Inf until first sample
}

func newSoakObservations() soakObservations {
	return soakObservations{minInterval: math.Inf(1)}
}

// driveCycle runs one prestige cycle's worth of driven ticks under the write lock,
// asserting bounds inline and folding worst values into obs. The engine must be
// pre-seeded, aged, allied, and event-suppressed by the caller (all under lock).
func driveCycle(t *testing.T, ge *GameEngine, ticks int, obs *soakObservations) {
	t.Helper()
	order := ge.progress.GetAgeOrder()
	rateTargets := hotRateTargets()

	for tick := 0; tick < ticks; tick++ {
		// Fastest sustainable encounter cadence, each still probability-gated inside
		// rollExpeditionEncounter (scouting+success = the most sociable pairing).
		if tick%soakEncounterEveryTicks == 0 {
			ge.rollExpeditionEncounter(ExpeditionScouting, true)
		}

		// Real per-tick expiry: this is the ONLY thing that plateaus concurrency.
		ge.Events.Tick(tick, ge.age, order, ge.currentEpoch)

		if tick%soakSampleEvery == 0 {
			r := ge.buildResolver()

			// Concurrency bound. Random events are suppressed, so every active
			// event here is a faction boon or a faction setback — the two things
			// the capacity gates bound.
			if n := len(ge.Events.active); n > obs.maxConcurrent {
				obs.maxConcurrent = n
			}
			if len(ge.Events.active) > maxConcurrentBoons {
				t.Fatalf("concurrent active faction effects = %d exceeds ceiling %d at tick %d "+
					"(capacity gate breached — rollExpeditionEncounter is granting past the cap)",
					len(ge.Events.active), maxConcurrentBoons, tick)
			}
			if n := ge.activeFactionBoonCount(); n > maxConcurrentFactionBoons {
				t.Fatalf("concurrent faction BOONS = %d exceeds cap %d at tick %d",
					n, maxConcurrentFactionBoons, tick)
			}
			if n := ge.activeFactionMalusCount(); n > maxConcurrentFactionMaluses {
				t.Fatalf("concurrent faction SETBACKS = %d exceeds cap %d at tick %d",
					n, maxConcurrentFactionMaluses, tick)
			}

			// Summed additive pools: production_all + every hot <res>_rate. Both
			// the raw Σ and the multiplier recalculateRates actually applies are
			// checked — the pool ceiling catches stacking, the factor assertion
			// pins the productionCap clamp itself.
			checkPool := func(target string) {
				sum := r.AddTotal(target)
				if math.IsNaN(sum) || math.IsInf(sum, 0) {
					t.Fatalf("additive pool %q is non-finite (%v) at tick %d", target, sum, tick)
				}
				if sum > obs.maxPool {
					obs.maxPool, obs.maxPoolTarget = sum, target
				}
				if sum > maxAdditivePool {
					t.Fatalf("additive pool %q = %.4f exceeds ceiling %.1f at tick %d "+
						"(boon capacity gate breached — encounters are stacking again)",
						target, sum, maxAdditivePool, tick)
				}
				// The applied multiplier, computed exactly as recalculateRates does.
				factor := clamp(1.0+sum, productionFloor, productionCap)
				if factor > productionCap+1e-9 || factor < productionFloor-1e-9 {
					t.Fatalf("applied factor for %q = %.6f outside [%.2f,%.2f] at tick %d "+
						"(productionCap/Floor clamp is not binding)",
						target, factor, productionFloor, productionCap, tick)
				}
			}
			checkPool("production_all")
			for _, target := range rateTargets {
				checkPool(target)
			}

			// tick_speed floor: recompute exactly as the engine does and assert the
			// clamp holds. Stacked TickSpeed boons must only saturate at the floor.
			ge.recalculateTickSpeed()
			interval := ge.getTickInterval()
			ms := float64(interval.Milliseconds())
			if ms < obs.minInterval {
				obs.minInterval = ms
			}
			if interval < MinTickInterval {
				t.Fatalf("tick interval %v dipped below MinTickInterval %v at tick %d "+
					"(tick_speed floor breached)", interval, MinTickInterval, tick)
			}
		}

		if tick%soakRecalcEvery == 0 {
			// Full rate recompute exercises the real multiply path (rate *= factor);
			// scan every resource rate for NaN/Inf.
			ge.recalculateRates()
			for _, def := range config.BaseResources() {
				rate := ge.Resources.GetRate(def.Key)
				if math.IsNaN(rate) || math.IsInf(rate, 0) {
					t.Fatalf("resource %q rate is non-finite (%v) at tick %d", def.Key, rate, tick)
				}
				amt := ge.Resources.Get(def.Key)
				if math.IsNaN(amt) || math.IsInf(amt, 0) {
					t.Fatalf("resource %q amount is non-finite (%v) at tick %d", def.Key, amt, tick)
				}
			}
		}
	}
}

// TestBoonSoak_BoundedAndDeterministic is the primary soak: several hundred
// thousand driven ticks per prestige cycle, worst-case allied standing, asserting
// every long-run invariant and proving Reset clears the boon pool between runs.
func TestBoonSoak_BoundedAndDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("soak test skipped under -short")
	}

	ge := NewGameEngine()
	obs := newSoakObservations()

	ge.mu.Lock()
	ge.SeedRNG(0x50A_C0FFEE_1234) // seed once; DoPrestige preserves it, stream persists
	ge.mu.Unlock()

	for cycle := 0; cycle < soakPrestigeCycles; cycle++ {
		ge.mu.Lock()
		ge.age = lastAgeKey()
		ge.currentEpoch = config.EpochForAge(ge.age)
		allyAllFactionsDiscovered(ge)
		ge.Events.nextEventTick = 1 << 40 // suppress random events: active list is boon-only

		// A fresh cycle starts with no active boons (guaranteed after DoPrestige
		// below; asserted here for the first cycle too).
		if n := len(ge.Events.active); n != 0 {
			ge.mu.Unlock()
			t.Fatalf("cycle %d started with %d active events, want 0 (boons leaked across prestige)", cycle, n)
		}

		driveCycle(t, ge, soakTicksPerCycle, &obs)

		// Prestige boundary: DoPrestige (the real prestige path — age is the last
		// age, so CanPrestige passes) must wipe every active boon while preserving
		// the run seed. It acquires the lock itself, so release first.
		ge.mu.Unlock()
		if t.Failed() {
			return
		}
		if err := ge.DoPrestige(); err != nil {
			t.Fatalf("cycle %d DoPrestige failed: %v", cycle, err)
		}

		ge.mu.Lock()
		if n := len(ge.Events.active); n != 0 {
			ge.mu.Unlock()
			t.Fatalf("DoPrestige left %d active boons (prestige did not clear the boon pool)", n)
		}
		ge.mu.Unlock()
	}

	t.Logf("soak over %d ticks across %d prestige cycles: maxConcurrentBoons=%d, "+
		"maxAdditivePool=%.4f on %q, minTickInterval=%.0fms (floor %dms)",
		soakPrestigeCycles*soakTicksPerCycle, soakPrestigeCycles,
		obs.maxConcurrent, obs.maxPool, obs.maxPoolTarget,
		obs.minInterval, MinTickInterval.Milliseconds())

	// The run must have actually exercised the system, else the bounds pass vacuously.
	if obs.maxConcurrent == 0 {
		t.Fatal("soak produced zero concurrent boons — driven path never injected anything")
	}
	if obs.maxPool <= 0 {
		t.Fatal("soak never accumulated a positive additive pool — nothing was stacked")
	}
}

// soakSignature captures a driven run's end state: discovered roster, active boons
// (key/duration/effects), and every resource rate — enough that two identical seeds
// diverge visibly if determinism breaks.
func soakSignature(ge *GameEngine) string {
	var b strings.Builder

	var disc []string
	for _, def := range config.BaseFactions() {
		if ge.Diplomacy.IsDiscovered(def.Key) {
			disc = append(disc, def.Key)
		}
	}
	sort.Strings(disc)
	b.WriteString("disc:")
	b.WriteString(strings.Join(disc, ","))

	// Active boons, sorted for a stable signature (append order is irrelevant to
	// the resolver, which sums them).
	var evs []string
	for _, ae := range ge.Events.active {
		var sb strings.Builder
		sb.WriteString(ae.Key)
		sb.WriteByte('@')
		sb.WriteString(strconv.Itoa(ae.TicksLeft))
		for _, eff := range ae.Effects {
			sb.WriteByte(' ')
			sb.WriteString(eff.Type)
			sb.WriteByte('=')
			sb.WriteString(ftoa(eff.Value))
		}
		evs = append(evs, sb.String())
	}
	sort.Strings(evs)
	b.WriteString("|boons:")
	b.WriteString(strings.Join(evs, ";"))

	b.WriteString("|rates:")
	for _, def := range config.BaseResources() {
		b.WriteString(def.Key)
		b.WriteByte('=')
		b.WriteString(ftoa(ge.Resources.GetRate(def.Key)))
		b.WriteByte(',')
	}
	return b.String()
}

// ftoa formats a float at fixed precision so signatures are byte-stable.
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 6, 64) }

// TestBoonSoak_Deterministic drives an identical seeded encounter sequence through
// two independent engines, crossing a real prestige (DoPrestige) boundary midway —
// which preserves the run seed — and asserts byte-identical end state. Determinism
// must survive tens of thousands of rolls AND a prestige reset.
//
// NOTE: this deliberately uses DoPrestige, NOT Reset(): Reset() is a full wipe that
// re-seeds with a fresh time-based seed (newSeed = time.Now().UnixNano()), so
// determinism intentionally does NOT hold across a wipe. See the findings report.
func TestBoonSoak_Deterministic(t *testing.T) {
	const seed int64 = 0xD37E_8B00_7A11
	const ticks = 60_000

	run := func() string {
		ge := NewGameEngine()
		obs := newSoakObservations()
		ge.mu.Lock()
		ge.SeedRNG(seed)
		ge.age = lastAgeKey()
		ge.currentEpoch = config.EpochForAge(ge.age)
		allyAllFactionsDiscovered(ge)
		ge.Events.nextEventTick = 1 << 40
		driveCycle(t, ge, ticks, &obs)
		ge.mu.Unlock()

		// Prestige boundary — age is the last age so CanPrestige passes; DoPrestige
		// preserves the run seed, so the stream continues identically across it.
		if err := ge.DoPrestige(); err != nil {
			t.Fatalf("DoPrestige failed: %v", err)
		}

		ge.mu.Lock()
		ge.age = lastAgeKey()
		ge.currentEpoch = config.EpochForAge(ge.age)
		allyAllFactionsDiscovered(ge)
		ge.Events.nextEventTick = 1 << 40
		driveCycle(t, ge, ticks, &obs)
		sig := soakSignature(ge)
		ge.mu.Unlock()
		return sig
	}

	a := run()
	b := run()
	if a != b {
		t.Fatalf("seeded soak diverged across identical runs:\n a=%s\n b=%s", a, b)
	}
	if !strings.Contains(a, "faction_boon_") {
		t.Fatalf("determinism run produced no faction boons — signature is trivial:\n%s", a)
	}
}

// TestBoon_StackingIsCapacityBounded REPLACES the former
// TestBoon_StackingIsUnbounded_NoDedup, which asserted (as a passing observation)
// that nothing bounded boon stacking but expiry. That is no longer the design:
// rollExpeditionEncounter refuses a positive boon at maxConcurrentFactionBoons.
//
// The low-level machinery is unchanged and still asserted here — InjectEvent
// APPENDS, so a raw double-inject really does double the pool; the fix is a gate
// in the encounter path, not dedup in the event manager. What changed is the
// second half: driving the REAL encounter path can no longer exceed the cap, and
// past it the additive pool stops growing.
func TestBoon_StackingIsCapacityBounded(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "medieval_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def := config.FactionByKey()["riverlands_tribes"] // food specialist
	applier := boonApplier{ge: ge, name: def.Name, key: def.Key}

	const pct = 0.15
	inject := func() {
		applier.InjectTimedEffects([]config.Effect{{
			Type: "food_rate", Target: "food", Value: pct,
		}}, 5000, "Specialty Windfall")
	}

	// Raw machinery: still append-not-replace. The capacity gate lives one level
	// up, so a direct applier call is deliberately NOT gated.
	inject()
	if got := ge.buildResolver().AddTotal("food_rate"); math.Abs(got-pct) > 1e-9 {
		t.Fatalf("after one inject food_rate pool = %.4f, want %.4f", got, pct)
	}
	inject()
	if got := ge.buildResolver().AddTotal("food_rate"); math.Abs(got-2*pct) > 1e-9 {
		t.Fatalf("after two injects food_rate pool = %.4f, want %.4f "+
			"(InjectEvent should append, not replace)", got, 2*pct)
	}
	if n := len(ge.Events.active); n != 2 {
		t.Fatalf("two injects produced %d active events, want 2 (InjectEvent should append, not replace)", n)
	}

	// Now the REAL path, from a clean slate: no cadence of encounters can push the
	// live boon count past the cap.
	ge.Events.active = nil
	ge.Events.nextEventTick = 1 << 40
	ge.age = lastAgeKey()
	ge.currentEpoch = config.EpochForAge(ge.age)
	allyAllFactionsDiscovered(ge)
	ge.SeedRNG(0xCA9AC17)

	for i := 0; i < 5000; i++ {
		ge.rollExpeditionEncounter(ExpeditionScouting, true)
		if n := ge.activeFactionBoonCount(); n > maxConcurrentFactionBoons {
			t.Fatalf("encounter %d pushed concurrent faction boons to %d, cap is %d",
				i, n, maxConcurrentFactionBoons)
		}
	}
	if got := ge.activeFactionBoonCount(); got == 0 {
		t.Fatal("5000 encounters produced no faction boons — the driven path is inert")
	}
}
