package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// setResource unlocks a resource, gives it headroom, zeroes its production rate,
// and sets its amount to EXACTLY amount (overriding any starting amount the engine
// seeds). Caller must NOT hold the engine lock.
func setResource(ge *GameEngine, key string, amount float64) {
	ge.mu.Lock()
	ge.Resources.UnlockResource(key)
	r, ok := ge.Resources.resources[key]
	if ok {
		r.Storage = amount + 1000
		r.Amount = amount
		r.Rate = 0
	}
	ge.mu.Unlock()
}

func setSoldiers(ge *GameEngine, amount float64) {
	ge.mu.Lock()
	ge.garrisonForTest, ge.garrisonForTestSet = amount, true
	ge.syncSoldiers()
	ge.mu.Unlock()
}

func setAge(ge *GameEngine, age string) {
	ge.mu.Lock()
	ge.age = age
	ge.mu.Unlock()
}

// TestLaunchExpedition_TakesThePartyOutOfThePosts verifies a campaign's launch
// takes its party (a share of the age's reference army) out of the soldiers at
// home: they leave their posts, so the army at home is that much smaller, and
// they are still people.
func TestLaunchExpedition_TakesThePartyOutOfThePosts(t *testing.T) {
	ge := NewGameEngine()
	learn(ge, "military_tactics") // the tech that opens campaigns
	setAge(ge, "iron_age")

	party := config.PartySizeOf(config.PartyMilitary, config.RefArmy(3))
	if party != 129 {
		t.Fatalf("the Iron Age campaign party is %d, want 129 (30%% of the reference army of 430)", party)
	}
	setSoldiers(ge, float64(party+71))
	need := int(ge.MissionCost(keyOffer(ge, "siege_campaign"))["soldiers"])
	if need != party {
		t.Fatalf("campaign cost = %d soldiers, want its party of %d", need, party)
	}

	if err := ge.LaunchExpedition(keyOffer(ge, "siege_campaign")); err != nil {
		t.Fatalf("LaunchExpedition(siege_campaign) returned error: %v", err)
	}

	if got := float64(ge.soldiersHome()); got != 71 {
		t.Errorf("soldiers at home after launch = %v, want 71 (200 less the party of %d)", got, party)
	}
	if active := ge.Military.ActiveByTier(TierGrand); active == nil || active.Soldiers != need || active.Party != need {
		t.Errorf("the party is %+v, want %d soldiers", active, need)
	}
}

// TestLaunchExpedition_DeductsCostResources verifies a scout's cost is taken
// from the two materials the age charges most: half of its five minutes of
// the town's income in each, rounded up.
func TestLaunchExpedition_DeductsCostResources(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "stone_age")

	setSoldiers(ge, 100)
	setResource(ge, "stone", 1000)
	setResource(ge, "wood", 1000)
	setIncome(ge, "stone", 0.5)
	setIncome(ge, "wood", 0.5)
	want := math.Ceil(2.5 * ticksPerMinute * 0.5)

	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) returned error: %v", err)
	}

	if stone := ge.Resources.Get("stone"); stone != 1000-want {
		t.Errorf("stone after scouting_party = %v, want %v", stone, 1000-want)
	}
	if wood := ge.Resources.Get("wood"); wood != 1000-want {
		t.Errorf("wood after scouting_party = %v, want %v", wood, 1000-want)
	}
}

// TestLaunchExpedition_InsufficientSoldiers_NoDeduction verifies a launch with
// too few soldiers errors and deducts nothing.
func TestLaunchExpedition_InsufficientSoldiers_NoDeduction(t *testing.T) {
	ge := NewGameEngine()
	learn(ge, "military_tactics") // the tech that opens campaigns
	setAge(ge, "iron_age")

	// siege_campaign takes a party of 129 here; give only 4.
	setSoldiers(ge, 4)

	err := ge.LaunchExpedition(keyOffer(ge, "siege_campaign"))
	if err == nil {
		t.Fatalf("expected error launching with insufficient soldiers, got nil")
	}
	if got := float64(ge.soldiersHome()); got != 4 {
		t.Errorf("soldiers after failed launch = %v, want 4 (unchanged)", got)
	}
	if ge.Military.HasActive() {
		t.Errorf("expedition should not be active after a failed launch")
	}
}

// TestLaunchExpedition_InsufficientCost_NoDeduction verifies a launch that
// can afford one cost resource but not the other errors and deducts NOTHING.
func TestLaunchExpedition_InsufficientCost_NoDeduction(t *testing.T) {
	ge := NewGameEngine()
	setAge(ge, "stone_age")

	// scouting_party costs stone and wood. Give enough stone, too little wood.
	setResource(ge, "stone", 1000)
	setResource(ge, "wood", 10)
	setIncome(ge, "stone", 0.5)
	setIncome(ge, "wood", 0.5)

	err := ge.LaunchExpedition(keyOffer(ge, "scouting_party"))
	if err == nil {
		t.Fatalf("expected error launching scouting_party with insufficient wood, got nil")
	}
	if stone := ge.Resources.Get("stone"); stone != 1000 {
		t.Errorf("stone after failed launch = %v, want 1000 (unchanged — no partial charge)", stone)
	}
	if wood := ge.Resources.Get("wood"); wood != 10 {
		t.Errorf("wood after failed launch = %v, want 10 (unchanged)", wood)
	}
	if ge.Military.HasActive() {
		t.Errorf("expedition should not be active after a failed launch")
	}
}

// TestScout_LaunchableInItsEra_NotCarriedToTheNextAge verifies a short
// mission is launchable with its party at home in every age of the Stone Era, and
// that an offer made in one age is gone when the next begins: a launch after
// the advance sends an offer of the new age, built from its words.
func TestScout_LaunchableInItsEra_NotCarriedToTheNextAge(t *testing.T) {
	for _, age := range config.Epochs()[0].Ages {
		ge := NewGameEngine()
		setAge(ge, age)
		setSoldiers(ge, 1000)
		setResource(ge, "stone", 1000)
		setResource(ge, "wood", 1000)
		setIncome(ge, "stone", 0.5)
		setIncome(ge, "wood", 0.5)

		if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
			t.Fatalf("a short mission should be launchable in %s, got: %v", age, err)
		}
	}

	ge2 := NewGameEngine()
	setAge(ge2, "stone_age")
	setSoldiers(ge2, 1e6)
	setResource(ge2, "stone", 1000)
	setResource(ge2, "wood", 1000)
	setResource(ge2, "iron", 1000)
	setIncome(ge2, "stone", 0.5)
	setIncome(ge2, "wood", 0.5)
	setIncome(ge2, "iron", 0.5)
	n := keyOffer(ge2, "scouting_party")
	setAge(ge2, "iron_age")
	if err := ge2.LaunchExpedition(n); err != nil {
		t.Fatalf("offer %d should be sent from the new age's board: %v", n, err)
	}
	words, _ := config.MissionWordsFor("iron_age")
	if got := ge2.Military.ActiveByTier(TierShort); got == nil || got.Age != "iron_age" || !strings.HasPrefix(strings.TrimPrefix(got.Title, config.RareTitleWord+" "), words.Tiers[config.WordsShort]+": ") {
		t.Errorf("the old age's offer was sent in the Iron Age: %+v", got)
	}
}

// TestExpedition_ResolvesWithoutWorkerLoss verifies the worker-loss hack is
// gone: resolving an expedition, won or lost, does not touch the worker pool.
// Soldiers are spent at launch only.
//
// A resolution can also roll a faction encounter (rollExpeditionEncounter), a
// separate system that moves workers on purpose in both directions: a boon can
// deliver a work gang, a setback can bury a few. When this test built an
// unseeded engine, that roll made it fail at random (CI: before=5 after=8, a
// work gang). It now plays a fixed set of seeds and judges only resolutions
// where no encounter fired. Encounter worker effects have their own tests
// (encounters_test.go, faction_malus_test.go).
func TestExpedition_ResolvesWithoutWorkerLoss(t *testing.T) {
	var clean, won, lost int
	for seed := int64(1); seed <= 40; seed++ {
		ge := newSeededEngine(seed)
		learn(ge, "military_tactics") // the tech that opens campaigns
		setAge(ge, "iron_age")
		setSoldiers(ge, 500)

		ge.mu.Lock()
		ge.Workers.UnlockType("worker")
		ge.Workers.Recruit("worker", 5, 100)
		popBefore := ge.Workers.TotalPop()
		ge.mu.Unlock()

		if err := ge.LaunchExpedition(keyOffer(ge, "siege_campaign")); err != nil {
			t.Fatalf("seed %d: LaunchExpedition error: %v", seed, err)
		}
		// a grand offer's length is rolled when it is made (60 to 120 minutes), so
		// budget for the longest; the loop exits once it resolves.
		for i := 0; i < int(240*ticksPerMinute) && ge.Military.HasActive(); i++ {
			ge.mu.Lock()
			ge.processExpeditions()
			ge.mu.Unlock()
		}
		if ge.Military.HasActive() {
			t.Fatalf("seed %d: expedition did not resolve within tick budget", seed)
		}

		// An encounter always discovers a civ here: none are known at the start
		// and the Iron Age has undiscovered ones in range. Skip those seeds.
		if len(ge.Diplomacy.factions) > 0 {
			continue
		}
		clean++
		switch {
		case logContains(ge, "Test grand military succeeded. Loot:"):
			won++
		case logContains(ge, "Test grand military failed. You kept 30% of the loot:"):
			lost++
		}
		if popAfter := ge.Workers.TotalPop(); popAfter != popBefore {
			t.Errorf("seed %d: worker pop changed by expedition resolution: before=%d after=%d (should be unchanged)", seed, popBefore, popAfter)
		}
	}
	// The seeds are fixed, so these counts are too. If a tuning change moves
	// them below the floor, widen the seed range rather than loosening the check.
	t.Logf("%d encounter-free resolutions: %d won, %d lost", clean, won, lost)
	if clean < 20 || won == 0 || lost == 0 {
		t.Fatalf("seed sweep too thin: %d encounter-free resolutions (%d won, %d lost); need >= 20 with both outcomes", clean, won, lost)
	}
}

// logContains reports whether any log line contains s.
func logContains(ge *GameEngine, s string) bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	for _, l := range ge.log {
		if strings.Contains(l.Message, s) {
			return true
		}
	}
	return false
}
