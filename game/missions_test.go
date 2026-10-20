package game

import (
	"math"
	"sort"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// missionReady parks a seeded engine in age with the mission techs learned, a
// distinct income of every resource and plenty of stock, so any mission of the
// era can be sent. Income of resource i (in key order) is 0.2 + 0.05*i per tick.
func missionReady(t *testing.T, seed int64, age string) *GameEngine {
	t.Helper()
	ge := autoExpeditionTestEngine(t, age, 0, 0, 1e9)
	ge.SeedRNG(seed)
	learn(ge, "exploration", "military_tactics")
	keys := make([]string, 0)
	for _, def := range config.BaseResources() {
		keys = append(keys, def.Key)
	}
	sort.Strings(keys)
	for i, res := range keys {
		setIncome(ge, res, 0.2+0.05*float64(i))
	}
	return ge
}

// eraFirstAges is the first age of every era, in order.
func eraFirstAges() []string {
	var out []string
	for _, era := range config.Epochs() {
		out = append(out, era.Ages[0])
	}
	return out
}

// TestMissionRolls_WithinRangesAndSeeded: a launch rolls its length, its
// chance and its haul inside the mission's ranges, and two engines with the
// same seed roll the same.
func TestMissionRolls_WithinRangesAndSeeded(t *testing.T) {
	kinds := []string{config.MissionScout, config.MissionExpedition, config.MissionCampaign}
	for _, age := range eraFirstAges() {
		for _, kind := range kinds {
			def := missionAt(NewGameEngine(), age, kind)
			lengths := map[float64]bool{}
			for seed := int64(1); seed <= 30; seed++ {
				launch := func() *ActiveExpedition {
					ge := missionReady(t, seed, age)
					setSoldiers(ge, 1e6)
					if err := ge.LaunchExpedition(def.Key); err != nil {
						t.Fatalf("seed %d: launch %s: %v", seed, def.Key, err)
					}
					return ge.Military.ActiveByCategory(def.Category)
				}
				a, b := launch(), launch()
				if a.LengthMinutes != b.LengthMinutes || a.SuccessChance != b.SuccessChance || a.HaulMinutes != b.HaulMinutes || a.TicksLeft != b.TicksLeft {
					t.Fatalf("seed %d: %s rolled %+v and then %+v from one seed", seed, def.Key, a, b)
				}
				lengths[a.LengthMinutes] = true

				if a.LengthMinutes < def.MinMinutes || a.LengthMinutes > def.MaxMinutes {
					t.Errorf("%s: length %v outside [%v, %v] minutes", def.Key, a.LengthMinutes, def.MinMinutes, def.MaxMinutes)
				}
				if want := int(math.Round(a.LengthMinutes * ticksPerMinute)); a.TicksLeft != want {
					t.Errorf("%s: %v minutes became %d ticks, want %d (no tech, no age stretch)", def.Key, a.LengthMinutes, a.TicksLeft, want)
				}
				if a.SuccessChance < 0.10 || a.SuccessChance > 0.95 {
					t.Errorf("%s: chance %v outside [0.10, 0.95]", def.Key, a.SuccessChance)
				}
				// The draw is +-0.10; a campaign's army term adds at most +-0.15.
				slack := 0.10
				if kind == config.MissionCampaign {
					slack = 0.25
				}
				if a.SuccessChance < def.BaseSuccess-slack-1e-9 || a.SuccessChance > def.BaseSuccess+slack+1e-9 {
					t.Errorf("%s: chance %v is further than %v from the base %v", def.Key, a.SuccessChance, slack, def.BaseSuccess)
				}
				lo, hi := def.MinMinutes*def.Yield*0.6, def.MaxMinutes*def.Yield*1.6
				if a.HaulMinutes < lo || a.HaulMinutes > hi {
					t.Errorf("%s: haul minutes %v outside [%v, %v]", def.Key, a.HaulMinutes, lo, hi)
				}
				if own := a.LengthMinutes * def.Yield; a.HaulMinutes < own*0.6-1e-9 || a.HaulMinutes > own*1.6+1e-9 {
					t.Errorf("%s: haul minutes %v outside [0.6, 1.6] x %v", def.Key, a.HaulMinutes, own)
				}
			}
			if len(lengths) < 20 {
				t.Errorf("%s: 30 seeds gave %d different lengths, want them to differ", def.Key, len(lengths))
			}
		}
	}
}

// TestMissionRolls_CampaignChanceFollowsPower: a campaign's chance moves
// with the army (0.10 for each unit of power over typical, within +-0.15)
// and a scout's does not.
func TestMissionRolls_CampaignChanceFollowsPower(t *testing.T) {
	roll := func(power float64, key string) float64 {
		mm := NewMilitaryManager()
		ge := missionReady(t, 3, "iron_age")
		if err := mm.LaunchMission(ge.gameRNG(), key, "iron_age", ageOrders(), power, 100, nil, nil, 3); err != nil {
			t.Fatal(err)
		}
		return mm.ActiveByCategory(mm.ExpeditionDefByKey(key).Category).SuccessChance
	}
	base := roll(100, "siege_campaign") // power = typical: no army term
	for _, c := range []struct{ power, term float64 }{{200, 0.10}, {0, -0.10}, {1000, 0.15}, {-5000, -0.15}} {
		got := roll(c.power, "siege_campaign") - base
		if math.Abs(got-c.term) > 1e-9 {
			t.Errorf("power %v moves the chance by %v, want %v", c.power, got, c.term)
		}
	}
	if a, b := roll(100, "road_survey"), roll(1000, "road_survey"); a != b {
		t.Errorf("a scout's chance moved with power: %v and %v", a, b)
	}
}

// chargedMost is the two resources the buildings of age charge most: the
// count of BaseCost entries across the age's buildings, ties by key.
func chargedMost(ge *GameEngine, age string) []string {
	counts := map[string]int{}
	info, _ := ge.rules.Age(age)
	for _, key := range info.UnlockBuildings {
		if b, ok := ge.rules.Building(key); ok {
			for res := range b.BaseCost {
				counts[res]++
			}
		}
	}
	var keys []string
	for res := range counts {
		keys = append(keys, res)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys[:2]
}

// TestMissionCost_IsMinutesOfIncomeInTheMostChargedMaterials: a scout costs
// 5 minutes and an expedition 15 minutes of the town's income, split in half
// over the two materials the age charges most, rounded up; a material with
// no income is left out.
func TestMissionCost_IsMinutesOfIncomeInTheMostChargedMaterials(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := missionReady(t, 1, age)
		top := chargedMost(ge, age)
		for _, c := range []struct {
			kind    string
			minutes float64
		}{{config.MissionScout, 5}, {config.MissionExpedition, 15}} {
			def := missionAt(ge, age, c.kind)
			cost := ge.MissionCost(def.Key)
			if len(cost) != 2 {
				t.Fatalf("%s in %s costs %v, want the two materials %v", def.Key, age, cost, top)
			}
			for _, res := range top {
				want := math.Ceil(c.minutes / 2 * ticksPerMinute * ge.townIncomeLocked(res))
				if cost[res] != want {
					t.Errorf("%s in %s costs %v %s, want %v (%v minutes of income %v, half)", def.Key, age, cost[res], res, want, c.minutes, ge.townIncomeLocked(res))
				}
			}
		}

		// No income of the first: it is left out, and the third does not fill in.
		setIncome(ge, top[0], 0)
		def := missionAt(ge, age, config.MissionScout)
		cost := ge.MissionCost(def.Key)
		if _, ok := cost[top[0]]; ok || len(cost) != 1 || cost[top[1]] == 0 {
			t.Errorf("%s in %s with no income of %s costs %v, want only %s", def.Key, age, top[0], cost, top[1])
		}
		setIncome(ge, top[1], 0)
		if cost := ge.MissionCost(def.Key); len(cost) != 0 {
			t.Errorf("%s in %s with no income at all costs %v, want nothing", def.Key, age, cost)
		}
	}
}

// TestMissionCost_CampaignIsTenMinutesOfSoldiers: a campaign costs the
// soldiers the town trains in ten minutes, rounded up, and at least one.
func TestMissionCost_CampaignIsTenMinutesOfSoldiers(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := missionReady(t, 1, age)
		def := missionAt(ge, age, config.MissionCampaign)
		setIncome(ge, "soldiers", 0.037)
		want := math.Ceil(10 * ticksPerMinute * ge.townIncomeLocked("soldiers"))
		cost := ge.MissionCost(def.Key)
		if cost["soldiers"] != want || len(cost) != 1 {
			t.Errorf("%s in %s costs %v, want %v soldiers and nothing else", def.Key, age, cost, want)
		}
		setIncome(ge, "soldiers", 0)
		if got := ge.MissionCost(def.Key)["soldiers"]; got != 1 {
			t.Errorf("%s in %s with no soldier income costs %v soldiers, want 1", def.Key, age, got)
		}
	}
}

// haulKeys is the two resources a mission should haul in age: of the
// resources the age's buildings produce and the town makes, the two with the
// highest income per unit of price level.
func haulKeys(ge *GameEngine, age string) []string {
	levels := ge.rules.PriceLevels(age)
	info, _ := ge.rules.Age(age)
	score := map[string]float64{}
	for _, key := range info.UnlockBuildings {
		b, ok := ge.rules.Building(key)
		if !ok {
			continue
		}
		for _, e := range b.Effects {
			if e.Type == "production" && ge.townIncomeLocked(e.Target) > 0 && levels[e.Target] > 0 {
				score[e.Target] = ge.townIncomeLocked(e.Target) / levels[e.Target]
			}
		}
	}
	var keys []string
	for res := range score {
		keys = append(keys, res)
	}
	sort.Slice(keys, func(i, j int) bool {
		if score[keys[i]] != score[keys[j]] {
			return score[keys[i]] > score[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > 2 {
		keys = keys[:2]
	}
	return keys
}

// TestMissionHaul_FormulaOnSuccessAndFailure: the haul is half of
// (income x ticks per minute x haul minutes) in each of two resources (an
// age that makes only one priced material hauls that one, still at half); a
// success brings it with the expedition bonus (and a campaign's pay), a
// failure 30% of it.
func TestMissionHaul_FormulaOnSuccessAndFailure(t *testing.T) {
	for _, age := range eraFirstAges() {
		for _, kind := range []string{config.MissionScout, config.MissionExpedition, config.MissionCampaign} {
			for _, win := range []bool{true, false} {
				ge := missionReady(t, 5, age)
				setSoldiers(ge, 1e6)
				ge.mu.Lock()
				ge.Military.campaignPay = 1.5
				ge.mu.Unlock()
				def := missionAt(ge, age, kind)
				want := haulKeys(ge, age)
				if len(want) == 0 || len(want) > 2 {
					t.Fatalf("%s: the age hauls %v, want one or two resources", age, want)
				}
				if err := ge.LaunchExpedition(def.Key); err != nil {
					t.Fatalf("%s: %v", def.Key, err)
				}
				active := ge.Military.ActiveByCategory(def.Category)
				if len(active.BaseRewards) != len(want) {
					t.Fatalf("%s hauls %v, want %v", def.Key, active.BaseRewards, want)
				}
				mult := 1.0
				if win {
					mult = 1 + ge.expeditionReward()
					if kind == config.MissionCampaign {
						mult *= 1.5
					}
				} else {
					mult = failedLootShare
					if kind == config.MissionCampaign {
						mult *= 1.5
					}
				}
				for _, res := range want {
					base := ge.townIncomeLocked(res) * ticksPerMinute * active.HaulMinutes / 2
					if math.Abs(active.BaseRewards[res]-base) > 1e-9*base {
						t.Errorf("%s hauls %v %s, want %v", def.Key, active.BaseRewards[res], res, base)
					}
				}

				// Come home: a chance over 1 always wins, under 0 always loses.
				ge.mu.Lock()
				active.TicksLeft = 1
				active.SuccessChance = map[bool]float64{true: 2, false: -1}[win]
				results := ge.Military.Tick(ge.gameRNG(), 0, ge.expeditionReward())
				ge.mu.Unlock()
				if len(results) != 1 || results[0].Success != win {
					t.Fatalf("%s came home as %+v, want success=%v", def.Key, results, win)
				}
				for _, res := range want {
					base := active.BaseRewards[res]
					if got := results[0].Rewards[res]; math.Abs(got-base*mult) > 1e-9*base*mult {
						t.Errorf("%s (win=%v) brought %v %s, want %v (%v x %v)", def.Key, win, got, res, base*mult, base, mult)
					}
				}
			}
		}
	}
}

// TestCampaign_HalfTheSoldiersComeBackOnAWin: a campaign that wins returns
// half the soldiers it sent, rounded down; one that loses returns none; a
// scout or an expedition never took any.
func TestCampaign_HalfTheSoldiersComeBackOnAWin(t *testing.T) {
	for _, c := range []struct {
		income float64 // soldiers/tick: cost = ceil(10 min of it)
		sent   int
		win    bool
		back   int
	}{
		{0.02, 6, true, 3},
		{0.0147, 5, true, 2}, // odd: rounded down
		{0.0033, 1, true, 0}, // one soldier: none come back
		{0.02, 6, false, 0},
	} {
		ge := missionReady(t, 2, "iron_age")
		setSoldiers(ge, 50)
		setIncome(ge, "soldiers", c.income)
		if got := int(ge.MissionCost("siege_campaign")["soldiers"]); got != c.sent {
			t.Fatalf("income %v: the campaign costs %d soldiers, test wants %d", c.income, got, c.sent)
		}
		if err := ge.LaunchExpedition("siege_campaign"); err != nil {
			t.Fatal(err)
		}
		if got := ge.Resources.Get("soldiers"); got != float64(50-c.sent) {
			t.Fatalf("soldiers after the launch = %v, want %d", got, 50-c.sent)
		}
		ge.mu.Lock()
		active := ge.Military.ActiveByCategory(ExpeditionMilitary)
		active.TicksLeft = 1
		active.SuccessChance = map[bool]float64{true: 2, false: -1}[c.win]
		ge.processExpeditions()
		ge.mu.Unlock()
		if ge.Military.HasActive() {
			t.Fatal("the campaign did not come home")
		}
		if got, want := ge.Resources.Get("soldiers"), float64(50-c.sent+c.back); got != want {
			t.Errorf("sent %d, win=%v: %v soldiers at home, want %v (%d back)", c.sent, c.win, got, want, c.back)
		}
	}
}

// TestMissionEras_OnlyTheCurrentErasMissionsAreOffered: each age offers the
// three missions of its own era and no other; a Neon mission is not offered
// in the Digital Era, is in the Fusion Age, and cannot be sent from the
// Digital Era.
func TestMissionEras_OnlyTheCurrentErasMissionsAreOffered(t *testing.T) {
	ge := NewGameEngine()
	for _, era := range config.Epochs() {
		for _, age := range era.Ages {
			got := ge.Military.GetAvailableExpeditions(age, ageOrders())
			if len(got) != 3 {
				t.Errorf("%s offers %d missions, want 3", age, len(got))
			}
			for _, def := range got {
				if def.Era != era.Key {
					t.Errorf("%s offers %s of the %s", age, def.Key, def.Era)
				}
			}
		}
	}
	has := func(age, key string) bool {
		for _, def := range ge.Military.GetAvailableExpeditions(age, ageOrders()) {
			if def.Key == key {
				return true
			}
		}
		return false
	}
	for _, key := range []string{"street_recon", "neon_heist", "fusion_plant_assault"} {
		if has("digital_age", key) {
			t.Errorf("the Digital Era offers %s", key)
		}
		if !has("fusion_age", key) {
			t.Errorf("the Fusion Age does not offer %s", key)
		}
	}
	digital := missionReady(t, 1, "digital_age")
	before := digital.Resources.Get("data")
	if err := digital.LaunchExpedition("neon_heist"); err == nil {
		t.Error("a Neon Era mission went out from the Digital Era")
	}
	if digital.Military.HasActive() || digital.Resources.Get("data") != before {
		t.Error("a refused launch took something")
	}
}

// TestEncounterChance_DoubledAndCapped: a mission resolution meets people
// twice as often as the base chances say, never above 0.6.
func TestEncounterChance_DoubledAndCapped(t *testing.T) {
	for _, c := range []struct {
		category string
		success  bool
		want     float64
	}{
		{ExpeditionScouting, true, 2 * encounterChanceScoutSuccess},
		{ExpeditionScouting, false, 2 * encounterChanceScoutFail},
		{ExpeditionMilitary, true, 2 * encounterChanceMilitarySuccess},
		{ExpeditionMilitary, false, 2 * encounterChanceMilitaryFail},
	} {
		if got := encounterChance(c.category, c.success); math.Abs(got-c.want) > 1e-12 || got > 0.6 {
			t.Errorf("%s success=%v: chance %v, want %v", c.category, c.success, got, c.want)
		}
	}
	for base, want := range map[float64]float64{0.1: 0.2, 0.3: 0.6, 0.31: 0.6, 0.9: 0.6, 0: 0} {
		if got := missionEncounterChance(base); math.Abs(got-want) > 1e-12 {
			t.Errorf("base %v doubles to %v, want %v (capped at 0.6)", base, got, want)
		}
	}
}

// TestAutoExpedition_SendsTheEraScout: the Geographic Society sends the
// scout of the era it stands in, whatever else it could afford, in every
// era.
func TestAutoExpedition_SendsTheEraScout(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := autoExpeditionTestEngine(t, age, 1, 8, 1e9)
		learn(ge, "military_tactics")
		want := missionAt(ge, age, config.MissionScout)
		def := ge.pickAutoScoutExpedition(ge.progress.ageOrder())
		if def == nil || def.Key != want.Key {
			t.Fatalf("%s: the picker chose %+v, want %s", age, def, want.Key)
		}
		if launched, _ := runAutoTicks(ge, 1); launched != 1 {
			t.Fatalf("%s: %d dispatches on the first tick, want 1", age, launched)
		}
		if active := ge.Military.ActiveByCategory(ExpeditionScouting); active == nil || active.Key != want.Key {
			t.Errorf("%s: the society sent %+v, want %s", age, active, want.Key)
		}
		if ge.Military.ActiveByCategory(ExpeditionMilitary) != nil {
			t.Errorf("%s: the society sent a campaign", age)
		}
	}
}
