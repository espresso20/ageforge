package game

import (
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// Tests for the game → flavor glue in expedition_flavor.go. Package flavor tests
// the generator in isolation; these pin what the ENGINE does with it: every call
// site produces a line in every age, the prose comes off the engine's seeded rng,
// the war raid is logged exactly once, and none of it races with a UI reader.

// flavorTestEngine returns an engine seeded for reproducible prose, at an age.
func flavorTestEngine(seed int64, age string) *GameEngine {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.SeedRNG(seed)
	ge.age = age
	ge.mu.Unlock()
	return ge
}

// cleanFlavor reports whether a generated line is fit to drop into a log entry:
// non-empty, markup-free (the caller owns the colour tags), placeholder-free.
func cleanFlavor(s string) bool {
	return s != "" && !strings.ContainsAny(s, "[]%{}~") && s == strings.TrimSpace(s)
}

// TestFlavorCallSitesCoverEveryAge drives all three adapters in all 22 ages with
// the live expedition roster and the live faction roster. The expedition line is
// gated to roughly one resolution in three, so it is retried until the gate
// opens; everything else must produce a line on the first call.
func TestFlavorCallSitesCoverEveryAge(t *testing.T) {
	for _, age := range config.AgeOrder() {
		ge := flavorTestEngine(int64(len(age)), age)
		ge.mu.Lock()
		for _, def := range ge.Military.expeditions {
			for _, success := range []bool{true, false} {
				res := ExpeditionResult{
					Key: def.Key, Name: def.Name, Category: def.Category, Success: success,
					Rewards: def.Rewards,
				}
				line := ""
				for i := 0; i < 60 && line == ""; i++ {
					line = ge.expeditionFlavorLine(res)
				}
				if !cleanFlavor(line) {
					t.Errorf("%s: expedition %s success=%v produced %q", age, def.Key, success, line)
				}
			}
		}
		for _, def := range config.BaseFactions() {
			for _, m := range []flavor.Moment{flavor.EncounterStandoff, flavor.EncounterAtCapacity} {
				if line := ge.factionEncounterFlavor(m, def); !cleanFlavor(line) {
					t.Errorf("%s: %v for %s produced %q", age, m, def.Key, line)
				}
			}
			raid := RaidRequest{FactionKey: def.Key, Resource: def.Specialty, Amount: 100,
				Message: raidMessage(def, 100, def.Specialty)}
			line := ge.raidLogLine(raid)
			if !strings.HasPrefix(line, raid.Message+" ") {
				t.Errorf("%s: raid line for %s does not lead with the mechanical message: %q", age, def.Key, line)
				continue
			}
			if tail := strings.TrimPrefix(line, raid.Message+" "); !cleanFlavor(tail) {
				t.Errorf("%s: raid flavour for %s is %q", age, def.Key, tail)
			}
		}
		ge.mu.Unlock()
	}
}

// TestRaidLineUnknownFactionFallsBack pins the guard in raidFlavorLine: a raid
// from a key config does not know still logs its mechanical line, just bare.
func TestRaidLineUnknownFactionFallsBack(t *testing.T) {
	ge := flavorTestEngine(1, "medieval_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()
	raid := RaidRequest{FactionKey: "no_such_civ", Resource: "gold", Amount: 10, Message: "mechanical"}
	if got := ge.raidLogLine(raid); got != "mechanical" {
		t.Fatalf("raidLogLine for an unknown faction = %q, want the bare mechanical message", got)
	}
}

// callSiteRun makes a fixed sequence of calls through every adapter and returns
// the prose. Used to compare engines seeded the same and differently.
func callSiteRun(seed int64) []string {
	ge := flavorTestEngine(seed, "iron_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()
	var out []string
	defs := config.BaseFactions()
	for i := 0; i < 120; i++ {
		exp := ge.Military.expeditions[i%len(ge.Military.expeditions)]
		out = append(out, ge.expeditionFlavorLine(ExpeditionResult{
			Key: exp.Key, Name: exp.Name, Category: exp.Category, Success: i%3 != 0, Rewards: exp.Rewards,
		}))
		def := defs[i%len(defs)]
		out = append(out, ge.factionEncounterFlavor(flavor.EncounterStandoff, def))
		out = append(out, ge.raidLogLine(RaidRequest{FactionKey: def.Key, Resource: "gold", Amount: 50, Message: "m"}))
	}
	return out
}

// TestFlavorFollowsTheEngineSeed is the determinism contract at the engine
// level: the prose stream comes off ge.rng, so two engines with the same seed say
// the same things in the same order, and a different seed says different things.
// This is what would break if an adapter reached for math/rand's global source.
func TestFlavorFollowsTheEngineSeed(t *testing.T) {
	a, b, c := callSiteRun(20260926), callSiteRun(20260926), callSiteRun(20260927)
	differ := 0
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("call %d differs between engines with the same seed:\n  %q\n  %q", i, a[i], b[i])
		}
		if a[i] != c[i] {
			differ++
		}
	}
	if differ < len(a)/2 {
		t.Fatalf("different seeds agreed on %d of %d calls; the prose is not following ge.rng", len(a)-differ, len(a))
	}
}

// TestExpeditionFlavorOdds pins the "present but not stale" gate: about one
// resolution in expeditionFlavorOdds gets a flavour line.
func TestExpeditionFlavorOdds(t *testing.T) {
	ge := flavorTestEngine(5, "classical_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()
	exp := ge.Military.expeditions[0]
	const n = 6000
	got := 0
	for i := 0; i < n; i++ {
		if ge.expeditionFlavorLine(ExpeditionResult{Key: exp.Key, Name: exp.Name, Category: exp.Category, Success: true}) != "" {
			got++
		}
	}
	want := 1.0 / expeditionFlavorOdds
	if share := float64(got) / n; share < want-0.03 || share > want+0.03 {
		t.Fatalf("%.1f%% of resolutions got a flavour line; want about %.1f%%", share*100, want*100)
	}
}

// TestTopRewardIsDeterministic pins topReward: largest amount wins, ties break on
// config resource order (never on map iteration order), and nothing to name
// returns ("", 0).
func TestTopRewardIsDeterministic(t *testing.T) {
	if k, a := topReward(nil); k != "" || a != 0 {
		t.Fatalf("topReward(nil) = %q, %v", k, a)
	}
	if k, _ := topReward(map[string]float64{"gold": 5, "food": 900, "wood": 3}); k != "food" {
		t.Fatalf("largest reward: got %q, want food", k)
	}
	order := config.BaseResources()
	first, second := order[0].Key, order[1].Key
	for i := 0; i < 200; i++ {
		if k, _ := topReward(map[string]float64{second: 10, first: 10}); k != first {
			t.Fatalf("tie broke to %q on iteration %d; want config order (%q)", k, i, first)
		}
	}
}

// TestWarRaidIsLoggedOnce is the regression test for the war-raid repeat this
// branch fixed: the raid used to be announced by DiplomacyManager AND by the
// engine, and every announcement reprinted the civ's whole backstory. One raid,
// one entry, no backstory, flavour appended.
func TestWarRaidIsLoggedOnce(t *testing.T) {
	ge := flavorTestEngine(11, "medieval_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def := config.FactionByKey()["ironhold_clans"]
	ge.Diplomacy.factions[def.Key] = &FactionState{Discovered: true, Opinion: -100, Status: "rival", AtWar: true}
	ge.tick = 4 * config.StretchTicks(ge.age, warRaidInterval) // a raid tick
	ge.Diplomacy.factions[def.Key].LastProvocationTick = ge.tick
	before := len(ge.log)
	ge.processDiplomacy()

	raids := 0
	for _, e := range ge.log[before:] {
		if !strings.Contains(e.Message, "raided you") {
			continue
		}
		raids++
		if def.Backstory != "" && strings.Contains(e.Message, def.Backstory) {
			t.Errorf("raid entry reprints the backstory: %q", e.Message)
		}
		if !strings.Contains(e.Message, "[-] ") {
			t.Errorf("raid entry has no flavour after the mechanical half: %q", e.Message)
		}
	}
	if raids != 1 {
		t.Fatalf("one raid produced %d log entries; want exactly 1", raids)
	}
}

// TestFlavorIsRaceFreeUnderTheEngineLock runs the real tick path (which reaches
// every adapter through processExpeditions, rollExpeditionEncounter and
// processDiplomacy) on one goroutine while others read GetState, the way the UI
// does. Run with -race: the Stream is not concurrency-safe by design, so this is
// the test that proves it is only ever touched under the write lock.
func TestFlavorIsRaceFreeUnderTheEngineLock(t *testing.T) {
	ge := flavorTestEngine(99, "medieval_age")
	ge.mu.Lock()
	for _, def := range config.BaseFactions() {
		ge.Diplomacy.factions[def.Key] = &FactionState{Discovered: true, Opinion: -100, Status: "rival", AtWar: true}
	}
	ge.mu.Unlock()

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = ge.GetState()
				}
			}
		}()
	}

	rng := rand.New(rand.NewSource(1))
	defs := config.BaseFactions()
	for i := 0; i < 400; i++ {
		ge.doTick()
		ge.mu.Lock()
		for _, fs := range ge.Diplomacy.factions {
			fs.LastProvocationTick = ge.tick // keep every war alive
		}
		ge.rollExpeditionEncounter(ExpeditionScouting, rng.Intn(2) == 0)
		exp := ge.Military.expeditions[i%len(ge.Military.expeditions)]
		_ = ge.expeditionFlavorLine(ExpeditionResult{Key: exp.Key, Name: exp.Name, Category: exp.Category, Success: i%2 == 0, Rewards: exp.Rewards})
		_ = ge.factionEncounterFlavor(flavor.EncounterAtCapacity, defs[i%len(defs)])
		ge.mu.Unlock()
	}
	close(stop)
	wg.Wait()

	ge.mu.Lock()
	defer ge.mu.Unlock()
	raids := 0
	for _, e := range ge.log {
		if strings.Contains(e.Message, "raided you") {
			raids++
		}
	}
	if raids == 0 {
		t.Fatal("400 ticks at war with every civ produced no raid entries; the raid path was not exercised")
	}
}
