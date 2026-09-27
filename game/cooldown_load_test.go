package game

import "testing"

// TestLoadGame_KeepsCooldowns: the festival and black-market cooldowns were
// not saved, so a load made both available at once (the saveload scenario
// caught a loaded bot holding a festival the live game was still waiting
// on). Prestige resets them with the tick counter.
func TestLoadGame_KeepsCooldowns(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(1)
	ge.tick = 1000
	ge.festivalReadyTick = 1500
	ge.blackMarketReadyTick = 1200
	if err := ge.SaveGame("cooldowns"); err != nil {
		t.Fatal(err)
	}
	b := NewGameEngine()
	if err := b.LoadGame("cooldowns"); err != nil {
		t.Fatal(err)
	}
	if b.festivalReadyTick != 1500 || b.blackMarketReadyTick != 1200 {
		t.Errorf("after load: festival ready at %d, black market at %d; want 1500 and 1200", b.festivalReadyTick, b.blackMarketReadyTick)
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
