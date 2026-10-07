package game

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Same seed, same inputs, same run. Every gameplay roll comes off the engine's
// seeded RNG (and the quip stream seeded beside it), and loops that roll dice or
// emit log lines walk maps in sorted order, so two engines built from one seed
// and fed the same actions must agree to the last bit. A difference means some
// code path is reading global math/rand, the wall clock, or map order.

// determinismTicks is long enough for several random events, a few expedition
// resolutions (with their encounter rolls), lending windows and war raids.
const determinismTicks = 3000

// determinismScript is the fixed script determinismRun drives: what to build,
// and which expeditions to send.
type determinismScript struct {
	builds, scouting, military []string
}

// determinismRun builds a mid-game engine from seed and drives it through a
// fixed script, returning a fingerprint of everything the RNG can touch.
func determinismRun(t *testing.T, seed int64) string {
	t.Helper()
	ge, sc := determinismSetup(t, seed)
	transcript := sc.play(ge, 0, determinismTicks)
	return determinismFingerprint(ge) + transcript
}

// determinismSetup builds determinismRun's mid-game engine and its script.
func determinismSetup(t *testing.T, seed int64) (*GameEngine, determinismScript) {
	t.Helper()
	ge := newSeededEngine(seed)

	// Medieval Age with every earlier age's unlocks, so events, both expedition
	// categories, diplomacy and faction encounters are all live.
	target := "medieval_age"
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == target {
			break
		}
	}
	ge.advanceAge(target)
	ge.pendingCatastrophe = "" // don't let a transition roll freeze the run

	// A few of every unlocked building, so production, storage and morale are
	// sums over many building types (the float sums that must not follow map
	// order), and a workforce partly assigned so fill ratios are fractional.
	builds := slices.Sorted(maps.Keys(ge.Buildings.unlocked))
	for _, key := range builds {
		if ge.Buildings.defs[key].Category != "wonder" {
			ge.Buildings.counts[key] = 3
		}
	}
	for key := range ge.Resources.resources {
		ge.Resources.UnlockResource(key)
	}
	ge.recalculateRates()
	for _, r := range ge.Resources.resources {
		r.Amount = r.Storage / 2
	}
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 120, 1000)
	for i, key := range builds {
		if ge.Buildings.defs[key].WorkerCapacity > 0 {
			ge.Workers.Assign("worker", key, 1+i%3)
		}
	}

	// Meet every civ. Peaceful ones start friendly enough to lend workers; the
	// first aggressive one is at war, so raids and war encounters fire.
	atWar := false
	for _, def := range config.BaseFactions() {
		ge.Diplomacy.DiscoverFaction(def.Key)
		fs := ge.Diplomacy.factions[def.Key]
		switch {
		case def.Personality == "peaceful":
			fs.Opinion = 60
		case def.Personality == "aggressive" && !atWar:
			fs.AtWar, fs.Opinion, fs.LastProvocationTick = true, -80, 1<<30
			atWar = true
		}
	}

	// The script sends expeditions, takes deals and starts routes: hold the
	// techs those commands wait for.
	learn(ge, "the_wheel", "exploration", "envoys", "drama", "military_tactics", "navigation")

	var scouting, military []string
	for _, def := range ge.Military.GetAvailableExpeditions(ge.age, ge.progress.GetAgeOrder()) {
		if def.Category == ExpeditionScouting {
			scouting = append(scouting, def.Key)
		} else {
			military = append(military, def.Key)
		}
	}
	if len(builds) == 0 || len(scouting) == 0 || len(military) == 0 {
		t.Fatalf("scenario setup: %d buildings, %d scouting, %d military expeditions", len(builds), len(scouting), len(military))
	}
	return ge, determinismScript{builds: builds, scouting: scouting, military: military}
}

// play runs script iterations [from, to) on ge, one tick each, and returns
// the log transcript. Split at any iteration, two calls play exactly what one
// call over the whole range would.
func (sc determinismScript) play(ge *GameEngine, from, to int) string {
	builds, scouting, military := sc.builds, sc.scouting, sc.military
	// The engine keeps only the last MaxLogSize lines, so the log is sampled
	// every iteration: each pass records the lines stamped with this
	// iteration's ticks. Lines from one doTick are recorded twice (once in
	// the next pass too), which is harmless: the transcript only has to be a
	// deterministic function of the full log history.
	var transcript strings.Builder
	for i := from; i < to; i++ {
		t0 := ge.tick
		// The script. Errors are part of the run (a failed build is still
		// deterministic), so they are ignored rather than asserted.
		if i%7 == 0 {
			_ = ge.BuildBuilding(builds[(i/7)%len(builds)])
		}
		ge.mu.Lock()
		// Keep the colony fed and armed so the run does not collapse into
		// an empty, starving state that exercises nothing.
		if r := ge.Resources.resources["food"]; r.Amount < 500 {
			r.Amount = 500
		}
		if i%40 == 0 {
			ge.Resources.resources["soldiers"].Amount = 200
		}
		ge.mu.Unlock()
		if i%40 == 0 {
			_ = ge.LaunchExpedition(scouting[(i/40)%len(scouting)])
			_ = ge.LaunchExpedition(military[(i/40)%len(military)])
		}
		if i%90 == 0 {
			_ = ge.RecruitWorker("worker", 2)
		}
		if i%150 == 75 {
			// Take each civ's first open trade deal, in roster order.
			for _, def := range config.BaseFactions() {
				for n := 1; n <= 5; n++ {
					if _, err := ge.AcceptFactionDeal(def.Key, n); err == nil {
						break
					}
				}
			}
		}
		ge.mu.Lock()
		if ge.pendingCatastrophe != "" {
			ge.mu.Unlock()
			_ = ge.Endure()
		} else {
			ge.mu.Unlock()
		}
		ge.StepTicks(1)
		for _, l := range ge.log {
			if l.Tick >= t0 {
				fmt.Fprintf(&transcript, "log %d %s %s\n", l.Tick, l.Type, l.Message)
			}
		}
	}
	return transcript.String()
}

// determinismFingerprint renders the RNG-reachable state as text. Floats are
// printed as their bit patterns: "close enough" is not the property under test.
func determinismFingerprint(ge *GameEngine) string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	var b strings.Builder
	fmt.Fprintf(&b, "tick %d age %s pop %d pending %q\n", ge.tick, ge.age, ge.Workers.TotalPop(), ge.pendingCatastrophe)
	for _, k := range slices.Sorted(maps.Keys(ge.Resources.resources)) {
		r := ge.Resources.resources[k]
		fmt.Fprintf(&b, "res %s %x %x\n", k, math.Float64bits(r.Amount), math.Float64bits(r.Rate))
	}
	for _, k := range slices.Sorted(maps.Keys(ge.Buildings.counts)) {
		fmt.Fprintf(&b, "bld %s %d\n", k, ge.Buildings.counts[k])
	}
	for _, q := range ge.buildQueue {
		fmt.Fprintf(&b, "queue %s %d\n", q.BuildingKey, q.TicksLeft)
	}
	for _, ae := range ge.Events.active {
		fmt.Fprintf(&b, "event %s %d\n", ae.Key, ae.TicksLeft)
	}
	fmt.Fprintf(&b, "next event %d streaks %d/%d\n", ge.Events.nextEventTick, ge.Events.goodStreak, ge.Events.badStreak)
	for _, k := range slices.Sorted(maps.Keys(ge.Diplomacy.factions)) {
		fs := ge.Diplomacy.factions[k]
		fmt.Fprintf(&b, "civ %s op %d war %v status %s deals %d/%d/%s\n", k, fs.Opinion, fs.AtWar, fs.Status, fs.DealRound, fs.DealTicks, fs.DealsFor)
		for _, d := range fs.Deals {
			fmt.Fprintf(&b, "  deal %d %s %s %x %s %x %d %v\n", d.ID, d.Kind, d.Give, math.Float64bits(d.GiveAmt), d.Get, math.Float64bits(d.GetAmt), d.Standing, d.Taken)
		}
	}
	for _, lb := range ge.Diplomacy.lentBatches {
		fmt.Fprintf(&b, "lent %s %d %d %v\n", lb.FactionKey, lb.Count, lb.ReturnTick, lb.Permanent)
	}
	fmt.Fprintf(&b, "expeditions completed %d\n", ge.Military.completedCount)
	for _, k := range slices.Sorted(maps.Keys(ge.Military.totalLoot)) {
		fmt.Fprintf(&b, "loot %s %x\n", k, math.Float64bits(ge.Military.totalLoot[k]))
	}
	return b.String()
}

func TestDeterminism_SameSeedSameRun(t *testing.T) {
	isolateAccountDir(t)
	a := determinismRun(t, 42)
	b := determinismRun(t, 42)
	if a != b {
		t.Fatalf("same seed, same inputs, different runs:\n%s", firstDiff(a, b))
	}
	// Guard against a vacuous pass: the script must actually have exercised
	// the random systems it exists to cover.
	for _, want := range []string{"Expedition resolved", "Event triggered"} {
		if !strings.Contains(a, want) {
			t.Errorf("run never logged %q; the scenario no longer reaches that system", want)
		}
	}
}

func TestDeterminism_DifferentSeedsDiverge(t *testing.T) {
	isolateAccountDir(t)
	if determinismRun(t, 42) == determinismRun(t, 43) {
		t.Fatal("seeds 42 and 43 produced identical runs; the seed is not reaching the RNG")
	}
}

// firstDiff returns the first differing line of two fingerprints with a little
// context.
func firstDiff(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return fmt.Sprintf("line %d:\n  run 1: %s\n  run 2: %s", i+1, al[i], bl[i])
		}
	}
	return fmt.Sprintf("one fingerprint is a prefix of the other (%d vs %d lines)", len(al), len(bl))
}
