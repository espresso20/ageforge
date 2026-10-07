package game

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// pacing_clocks_test.go: every tick clock the one-week curve re-times
// (config.StretchTicks) runs at its base length in the Primitive and Stone
// Ages and config.PacingStretch times longer from the Bronze Age on, so an age
// holds as many events, raids, routes and expeditions as it did before.

// stretchedAge is an age past the stretch; baseAge one before it.
const (
	stretchedAge = "iron_age"
	baseAge      = "stone_age"
)

func ageOrderMap() map[string]int {
	m := map[string]int{}
	for i, a := range config.AgeOrder() {
		m[a] = i
	}
	return m
}

// TestEventClocksStretch: the delay to the next random event, a timed event's
// duration and an event's cooldown all follow the age's stretch.
func TestEventClocksStretch(t *testing.T) {
	order := ageOrderMap()
	for _, age := range []string{baseAge, stretchedAge} {
		lo, hi := config.StretchTicks(age, eventMinDelay), config.StretchTicks(age, eventMaxDelay)
		rng := rand.New(rand.NewSource(7))
		for i := 0; i < 200; i++ {
			if d := eventDelay(rng, age); d < lo || d > hi {
				t.Fatalf("%s: event delay %d outside [%d, %d]", age, d, lo, hi)
			}
		}

		// Duration: play until a timed event fires and check what it got.
		em := NewEventManager()
		rng = rand.New(rand.NewSource(3))
		checked := false
		for tick := 1; tick < 200000 && !checked; tick++ {
			fired, _ := em.Tick(rng, tick, age, order, config.EpochForAge(age))
			for _, def := range fired {
				if def.Duration <= 0 {
					continue
				}
				for _, ae := range em.active {
					if ae.Key == def.Key {
						if want := config.StretchTicks(age, def.Duration); ae.TicksLeft != want {
							t.Errorf("%s: %s lasts %d ticks, want %d", age, def.Key, ae.TicksLeft, want)
						}
						checked = true
					}
				}
			}
		}
		if !checked {
			t.Fatalf("%s: no timed event fired", age)
		}

		// Cooldown: bountiful_harvest (Primitive, cooldown 50) is barred until
		// its stretched cooldown has passed since it last fired.
		def := config.EventByKey()["bountiful_harvest"]
		em = NewEventManager()
		const fired = 1000
		em.lastFired[def.Key] = fired
		cd := config.StretchTicks(age, def.Cooldown)
		eligible := func(tick int) bool {
			for _, d := range em.getEligible(tick, age, order, "", config.EpochForAge(age)) {
				if d.Key == def.Key {
					return true
				}
			}
			return false
		}
		if eligible(fired + cd - 1) {
			t.Errorf("%s: %s eligible %d ticks after firing, cooldown %d", age, def.Key, cd-1, cd)
		}
		if !eligible(fired + cd) {
			t.Errorf("%s: %s not eligible after its %d-tick cooldown", age, def.Key, cd)
		}
	}
	if config.StretchTicks(baseAge, eventMinDelay) != eventMinDelay {
		t.Errorf("the Stone Age's event delay moved")
	}
}

// TestDiplomacyClocksStretch: war raids, the war's quiet-ending, rival decay
// and a worker loan's length all follow the age's stretch.
func TestDiplomacyClocksStretch(t *testing.T) {
	raidEvery := config.StretchTicks(stretchedAge, warRaidInterval)
	if raidEvery != 104 {
		t.Fatalf("iron-age raid cadence = %d, want 104 (40 x 2.6)", raidEvery)
	}
	key := "void_reavers"
	dm := NewDiplomacyManager()
	dm.factions[key] = &FactionState{Discovered: true, Status: "neutral", Opinion: -90, AtWar: true}
	_ = dm.processWar(warRaidInterval, stretchedAge)
	if n := len(dm.TakePendingRaids()); n != 0 {
		t.Errorf("a raid fired on the base-curve tick %d in the %s", warRaidInterval, stretchedAge)
	}
	_ = dm.processWar(raidEvery, stretchedAge)
	if n := len(dm.TakePendingRaids()); n != 1 {
		t.Errorf("%d raids on the stretched raid tick %d, want 1", n, raidEvery)
	}

	// The war burns out after the stretched quiet spell, not the base one.
	cooldown := config.StretchTicks(stretchedAge, warCooldownTicks)
	fs := &FactionState{Discovered: true, Status: "neutral", Opinion: -90, AtWar: true, LastProvocationTick: 0}
	dm = NewDiplomacyManager()
	dm.factions["shadow_syndicate"] = fs
	_ = dm.processWar(warCooldownTicks, stretchedAge)
	if !fs.AtWar {
		t.Errorf("the war ended after the base %d quiet ticks in the %s", warCooldownTicks, stretchedAge)
	}
	_ = dm.processWar(cooldown, stretchedAge)
	if fs.AtWar {
		t.Errorf("the war outlived its %d-tick quiet spell", cooldown)
	}

	// Rival decay: -5 on the stretched cadence only, measured against a
	// neutral twin (personality drift moves both alike).
	rivalDM, twinDM := NewDiplomacyManager(), NewDiplomacyManager()
	rival := &FactionState{Discovered: true, Status: "rival", Opinion: 0}
	twin := &FactionState{Discovered: true, Status: "neutral", Opinion: 0}
	rivalDM.factions["ironhold_clans"], twinDM.factions["ironhold_clans"] = rival, twin
	for _, tick := range []int{rivalDecayInterval, config.StretchTicks(stretchedAge, rivalDecayInterval)} {
		rivalDM.Tick(rand.New(rand.NewSource(1)), stretchedAge, ageOrderMap(), tick, false)
		twinDM.Tick(rand.New(rand.NewSource(1)), stretchedAge, ageOrderMap(), tick, false)
		want := 0
		if tick != rivalDecayInterval {
			want = -5
		}
		if got := rival.Opinion - twin.Opinion; got != want {
			t.Errorf("tick %d: rival is %d against its neutral twin, want %d", tick, got, want)
		}
	}

	// A loan lasts the stretched time, and the message says so.
	var peaceful string
	for _, def := range config.BaseFactions() {
		if def.Personality == "peaceful" {
			peaceful = def.Key
			break
		}
	}
	dm = NewDiplomacyManager()
	dm.factions[peaceful] = &FactionState{Discovered: true, Status: "friendly", Opinion: 60}
	drift := config.StretchTicks(stretchedAge, driftInterval)
	rng := rand.New(rand.NewSource(5))
	var msgs []string
	tick := drift
	for ; tick < drift*500 && !dm.hasLentBatch(peaceful); tick += drift {
		msgs = dm.processLending(rng, tick, stretchedAge)
	}
	if !dm.hasLentBatch(peaceful) {
		t.Fatal("no loan in 500 seeded windows")
	}
	lent := dm.lentBatches[0]
	lendTicks := config.StretchTicks(stretchedAge, lendDurationTicks)
	if got := lent.ReturnTick - (tick - drift); got != lendTicks {
		t.Errorf("loan lasts %d ticks, want %d", got, lendTicks)
	}
	if want := DurationText(lendTicks, BaseTickInterval); len(msgs) == 0 || !strings.Contains(msgs[len(msgs)-1], want) {
		t.Errorf("loan message %q does not say %q", msgs, want)
	}
}

// TestDealRefreshStretch: offers rotate on the stretched hour.
func TestDealRefreshStretch(t *testing.T) {
	if got := dealRefreshFor(baseAge); got != dealRefreshTicks {
		t.Errorf("stone refresh = %d, want %d", got, dealRefreshTicks)
	}
	if got := dealRefreshFor(stretchedAge); got != 4680 {
		t.Errorf("iron refresh = %d, want 4680 (1800 x 2.6)", got)
	}
}

// TestExpeditionClocksStretch: a launch rolls in the stretched range, the
// panel shows that range, and the Stone Age keeps the typed one.
func TestExpeditionClocksStretch(t *testing.T) {
	for _, c := range []struct{ age, key string }{{baseAge, "scout_party"}, {stretchedAge, "trade_escort"}} {
		ge := newSeededEngine(9)
		learn(ge, "military_tactics") // the tech that opens campaigns
		setAge(ge, c.age)
		setSoldiers(ge, 50)
		setResource(ge, "food", 1000)
		setResource(ge, "wood", 1000)
		def := ge.Military.ExpeditionDefByKey(c.key)
		lo, hi := config.StretchTicks(c.age, def.DurationMin), config.StretchTicks(c.age, def.DurationMax)
		if err := ge.LaunchExpedition(c.key); err != nil {
			t.Fatalf("%s: launch %s: %v", c.age, c.key, err)
		}
		active := ge.Military.ActiveByCategory(def.Category)
		if active == nil || active.TicksLeft < lo || active.TicksLeft > hi {
			t.Errorf("%s: %s rolled %+v, want ticks in [%d, %d]", c.age, c.key, active, lo, hi)
		}
		for _, info := range ge.GetState().Military.Expeditions {
			if info.Key == c.key && (info.DurationMin != lo || info.DurationMax != hi) {
				t.Errorf("%s: panel shows %s as %d-%d ticks, want %d-%d", c.age, c.key, info.DurationMin, info.DurationMax, lo, hi)
			}
		}
	}
}

// TestAutoExpeditionStretch: the Geographic Society's cadence, floor
// included, stretches with the age.
func TestAutoExpeditionStretch(t *testing.T) {
	if got := autoExpeditionIntervalIn("industrial_age", 1, 0); got != config.StretchTicks("industrial_age", autoExpeditionBaseInterval) || got != 2340 {
		t.Errorf("one unstaffed society, industrial = %d, want 2340 (900 x 2.6)", got)
	}
	if got := autoExpeditionIntervalIn("industrial_age", 20, 1); got != 260 {
		t.Errorf("the floor, industrial = %d, want 260 (100 x 2.6)", got)
	}
	if got := autoExpeditionIntervalIn(baseAge, 1, 0); got != autoExpeditionBaseInterval {
		t.Errorf("one unstaffed society, stone = %d, want %d", got, autoExpeditionBaseInterval)
	}
}

// TestFestivalAndBlackMarketStretch: the festival's buff and cooldown and the
// black market's cooldown stretch with the age, in play and in the status the
// commands show.
func TestFestivalAndBlackMarketStretch(t *testing.T) {
	ge := NewGameEngine()
	ge.age = stretchedAge
	ge.Resources.UnlockResource("culture")
	ge.Resources.LoadAmounts(map[string]float64{"culture": 1_000_000})
	buff, cd := config.StretchTicks(stretchedAge, festivalBuffTicks), config.StretchTicks(stretchedAge, festivalCooldownTicks)
	if st := ge.FestivalStatus(); st.BuffTicks != buff || st.CooldownTicks != cd {
		t.Errorf("festival status %d/%d ticks, want %d/%d", st.BuffTicks, st.CooldownTicks, buff, cd)
	}
	if err := ge.DoFestival(); err != nil {
		t.Fatal(err)
	}
	for _, ae := range ge.Events.active {
		if ae.Key == "cultural_festival" && ae.TicksLeft != buff {
			t.Errorf("festival buff lasts %d ticks, want %d", ae.TicksLeft, buff)
		}
	}
	if ge.festivalReadyTick != ge.tick+cd {
		t.Errorf("next festival at tick %d, want %d", ge.festivalReadyTick, ge.tick+cd)
	}
	if !strings.Contains(lastLog(ge), DurationText(buff, BaseTickInterval)) {
		t.Errorf("festival log %q does not quote the stretched buff", lastLog(ge))
	}

	bm := bmEngine(50000)
	bm.blackMarketRand = blackMarketSeedWin()
	bcd := config.StretchTicks("colonial_age", blackMarketCooldownTicks)
	if st := bm.BlackMarketStatus(); st.CooldownTicks != bcd {
		t.Errorf("black market status cooldown %d, want %d", st.CooldownTicks, bcd)
	}
	if _, _, err := bm.DoBlackMarket("gold"); err != nil {
		t.Fatal(err)
	}
	if bm.blackMarketReadyTick != bm.tick+bcd {
		t.Errorf("black market ready at %d, want %d", bm.blackMarketReadyTick, bm.tick+bcd)
	}
}

// TestChainBoostStretch: a chain's speed boost lasts its typed length times
// the stretch of the age it lands in, and the bus event carries that length
// for the toast.
func TestChainBoostStretch(t *testing.T) {
	ge := NewGameEngine()
	ge.age = stretchedAge
	chain := config.MilestoneChainByKey()["military_chain"]
	for _, k := range chain.MilestoneKeys {
		ge.Milestones.completed[k] = true
	}
	var payload int
	ge.Bus.Subscribe(EventChainCompleted, func(e EventData) {
		if k, _ := e.Payload["key"].(string); k == chain.Key {
			payload, _ = e.Payload["boost_ticks"].(int)
		}
	})
	ge.checkMilestones()
	want := config.StretchTicks(stretchedAge, chain.BoostDuration)
	found := false
	for _, ae := range ge.Events.active {
		if ae.Key == chain.Key+"_boost" {
			found = true
			if ae.TicksLeft != want {
				t.Errorf("chain boost lasts %d ticks, want %d", ae.TicksLeft, want)
			}
		}
	}
	if !found {
		t.Fatal("the chain boost was not injected")
	}
	if payload != want {
		t.Errorf("chain event boost_ticks = %d, want %d", payload, want)
	}
}

// TestEndureDebuffStretch: Endure's Reconstruction Effort lasts its typed 216
// ticks stretched for the age the doom strikes in, and the log line says so.
func TestEndureDebuffStretch(t *testing.T) {
	if got := EndureDebuffTicksIn(baseAge); got != endureDebuffTicks {
		t.Errorf("stone debuff = %d, want %d", got, endureDebuffTicks)
	}
	if got := EndureDebuffTicksIn(stretchedAge); got != 562 {
		t.Errorf("iron debuff = %d, want 562 (216 x 2.6)", got)
	}
	ge := catEngine(t, stretchedAge, 1)
	if err := ge.forceCatastrophe(); err != nil {
		t.Fatal(err)
	}
	if err := ge.Endure(); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, ae := range ge.Events.active {
		if ae.Key == "endure_reconstruction" {
			found = true
			if ae.TicksLeft != 562 {
				t.Errorf("reconstruction lasts %d ticks, want 562", ae.TicksLeft)
			}
		}
	}
	if !found {
		t.Fatal("no Reconstruction Effort after Endure")
	}
	want := approxTicks(562, ge.tickIntervalLocked())
	quoted := false
	for _, l := range ge.log {
		if strings.Contains(l.Message, "Reconstruction: all production") && strings.Contains(l.Message, want) {
			quoted = true
		}
	}
	if !quoted {
		t.Errorf("no reconstruction log line quoting %s", want)
	}
}

// lastLog is the newest log message.
func lastLog(ge *GameEngine) string {
	if len(ge.log) == 0 {
		return ""
	}
	return ge.log[len(ge.log)-1].Message
}
