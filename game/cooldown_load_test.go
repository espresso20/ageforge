package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestLoadGame_KeepsCooldowns: the festival and black-market cooldowns were
// not saved, so a load made both available at once (the saveload scenario
// caught a loaded bot holding a festival the live game was still waiting
// on). Prestige resets them with the tick counter.
func TestLoadGame_KeepsCooldowns(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.tick = 1000
	ge.festivalReadyTick = 1250
	ge.blackMarketReadyTick = 1200
	if err := ge.SaveGame("cooldowns"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("cooldowns"); err != nil {
		t.Fatal(err)
	}
	if b.festivalReadyTick != 1250 || b.blackMarketReadyTick != 1200 {
		t.Errorf("after load: festival ready at %d, black market at %d; want 1250 and 1200", b.festivalReadyTick, b.blackMarketReadyTick)
	}
	if b.FestivalStatus().Ready {
		t.Error("a festival on cooldown is ready again after a load")
	}

	b.mu.Lock()
	b.age = "modern_age"
	b.completePrestige(prestigePlain)
	b.mu.Unlock()
	if b.festivalReadyTick != 0 || b.blackMarketReadyTick != 0 {
		t.Errorf("after prestige (tick back to 0): festival ready at %d, black market at %d; want 0", b.festivalReadyTick, b.blackMarketReadyTick)
	}
}

// TestSuccumb_CooldownsStartOver: the cooldowns are tick numbers, and a
// Succumb sends the tick counter back to 0 without having reset them: a
// festival cooldown with 780 ticks left at tick 5,001 had 5,781 left in the
// new run. After a Succumb neither cooldown is ever longer than a fresh one.
func TestSuccumb_CooldownsStartOver(t *testing.T) {
	for _, tick := range []int{0, 1, 779, 5001, 250000} {
		ge := bmEngine(100000)
		ge.SeedRNG(4)
		ge.currentEpoch = config.EpochForAge(ge.age)
		ge.tick = tick
		ge.blackMarketRand = blackMarketSeedWin()
		if _, _, err := ge.DoBlackMarket("gold"); err != nil {
			t.Fatalf("tick %d: black market: %v", tick, err)
		}
		if err := ge.DoFestival(); err != nil {
			t.Fatalf("tick %d: festival: %v", tick, err)
		}
		fs, bs := ge.FestivalStatus(), ge.BlackMarketStatus()
		if fs.CooldownLeft != fs.CooldownTicks || bs.CooldownLeft != bs.CooldownTicks || fs.CooldownTicks != 780 {
			t.Fatalf("tick %d: fresh cooldowns are %d of %d and %d of %d, want full ones (780 for the festival)", tick, fs.CooldownLeft, fs.CooldownTicks, bs.CooldownLeft, bs.CooldownTicks)
		}
		// The run's other timers start over too, as at a prestige.
		ge.autoExpeditionTicksLeft, ge.autoExpeditionStarved = 500, true
		ge.ageReady, ge.starvationTicks = true, 7
		ge.pendingCatastrophe = ge.currentEpoch
		if err := ge.Succumb(); err != nil {
			t.Fatalf("tick %d: Succumb: %v", tick, err)
		}
		if ge.autoExpeditionTicksLeft != 0 || ge.autoExpeditionStarved || ge.ageReady || ge.starvationTicks != 0 {
			t.Errorf("tick %d: after a Succumb the survey countdown is %d (starved %v), ready-to-advance %v, famine ticks %d; want a new run's",
				tick, ge.autoExpeditionTicksLeft, ge.autoExpeditionStarved, ge.ageReady, ge.starvationTicks)
		}
		fs, bs = ge.FestivalStatus(), ge.BlackMarketStatus()
		if fs.CooldownLeft > fs.CooldownTicks || bs.CooldownLeft > bs.CooldownTicks {
			t.Errorf("tick %d: after a Succumb the festival has %d ticks left (a fresh cooldown is %d) and the black market %d (fresh: %d)",
				tick, fs.CooldownLeft, fs.CooldownTicks, bs.CooldownLeft, bs.CooldownTicks)
		}
		if fs.CooldownLeft != 0 || bs.CooldownLeft != 0 || !fs.Ready {
			t.Errorf("tick %d: a new run starts with both cooldowns over; got festival %d, black market %d", tick, fs.CooldownLeft, bs.CooldownLeft)
		}
		// They stay in range as the new run goes on.
		ge.StepTicks(50)
		if fs = ge.FestivalStatus(); fs.CooldownLeft != 0 {
			t.Errorf("tick %d: 50 ticks into the new run the festival cooldown is %d", tick, fs.CooldownLeft)
		}
	}
}

// TestLoadGame_BoundsStuckCooldowns: a save written after a Succumb, before
// Succumb reset the cooldowns, carries one that ends thousands of ticks on.
// Loading it brings the cooldown back to a fresh one at most.
func TestLoadGame_BoundsStuckCooldowns(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(2)
	ge.tick = 40
	ge.festivalReadyTick = 5781
	ge.blackMarketReadyTick = 5625
	if err := ge.SaveGame("stuck"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("stuck"); err != nil {
		t.Fatal(err)
	}
	fs, bs := b.FestivalStatus(), b.BlackMarketStatus()
	if fs.CooldownLeft > fs.CooldownTicks || bs.CooldownLeft > bs.CooldownTicks {
		t.Errorf("after load the festival has %d ticks left (a fresh cooldown is %d) and the black market %d (fresh: %d)",
			fs.CooldownLeft, fs.CooldownTicks, bs.CooldownLeft, bs.CooldownTicks)
	}
	if b.cheaterBadge {
		t.Error("the load flagged an honest save")
	}
}
