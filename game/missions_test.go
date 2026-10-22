package game

import (
	"math"
	"slices"
	"sort"
	"strings"
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

// tuningOf is the tuning an offer was rolled from.
func tuningOf(o MissionOffer) (config.MissionTuning, float64) {
	t, _ := config.MissionKindTuning(o.Kind())
	base := t.BaseSuccess
	if o.Tier == TierGrand && o.Grand == GrandCivil {
		base = config.CivilGrandSuccess
	}
	return t, base
}

// TestMissionRolls_WithinRangesAndSeeded: an offer rolls its length, its
// chance and its haul inside its tier's ranges (a Legendary one with its
// changes), and two engines with the same seed roll the same board.
func TestMissionRolls_WithinRangesAndSeeded(t *testing.T) {
	for _, age := range eraFirstAges() {
		lengths := map[float64]bool{}
		for seed := int64(1); seed <= 30; seed++ {
			board := func() []boardSlot {
				ge := missionReady(t, seed, age)
				ge.EnsureMissionBoard()
				return ge.Military.board.clone().Slots
			}
			a, b := board(), board()
			if len(a) != 9 {
				t.Fatalf("seed %d: %d slots, want 9", seed, len(a))
			}
			for i := range a {
				if *a[i].Offer != *b[i].Offer {
					t.Fatalf("seed %d slot %d: rolled %+v and then %+v from one seed", seed, i+1, *a[i].Offer, *b[i].Offer)
				}
				o := *a[i].Offer
				tuning, base := tuningOf(o)
				lengths[o.LengthMinutes] = true
				minLen, maxLen := tuning.MinMinutes, tuning.MaxMinutes
				lo, hi := minLen*tuning.Yield*0.6, maxLen*tuning.Yield*1.6
				loSuccess, hiSuccess := math.Max(0.10, base-0.1), math.Min(0.95, base+0.1)
				if o.Rare {
					minLen, maxLen = minLen*1.5, maxLen*1.5
					lo, hi = lo*4, hi*4
					loSuccess, hiSuccess = math.Max(0.10, loSuccess-0.15), math.Max(0.10, hiSuccess-0.15)
				}
				if o.LengthMinutes < minLen-1e-9 || o.LengthMinutes > maxLen+1e-9 {
					t.Errorf("%s: length %v outside [%v, %v] minutes", o.Title, o.LengthMinutes, minLen, maxLen)
				}
				if o.SuccessChance < loSuccess-1e-9 || o.SuccessChance > hiSuccess+1e-9 {
					t.Errorf("%s: chance %v outside [%v, %v]", o.Title, o.SuccessChance, loSuccess, hiSuccess)
				}
				if o.HaulMinutes < lo-1e-9 || o.HaulMinutes > hi+1e-9 {
					t.Errorf("%s: haul minutes %v outside [%v, %v]", o.Title, o.HaulMinutes, lo, hi)
				}
			}
		}
		if len(lengths) < 150 {
			t.Errorf("%s: 270 offers gave %d different lengths, want them to differ", age, len(lengths))
		}
	}
}

// TestMissionLaunch_TicksFollowTheRolledLength: a mission is out for its
// rolled length in ticks (no tech here, no age stretch), and sent with the
// odds it was offered with (a short mission is not changed by the army).
func TestMissionLaunch_TicksFollowTheRolledLength(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := missionReady(t, 4, age)
		setSoldiers(ge, 1e12)
		for _, c := range []struct{ tier, grand string }{{TierShort, ""}, {TierLong, ""}, {TierGrand, GrandCivil}} {
			n := plantOffer(ge, c.tier, c.grand)
			offer, _ := ge.Military.BoardOffer(n)
			if err := ge.LaunchExpedition(n); err != nil {
				t.Fatalf("%s %s: %v", age, c.tier, err)
			}
			a := ge.Military.ActiveByTier(c.tier)
			if want := int(math.Round(offer.LengthMinutes * ticksPerMinute)); a.TicksLeft != want {
				t.Errorf("%s %s: %v minutes became %d ticks, want %d", age, c.tier, offer.LengthMinutes, a.TicksLeft, want)
			}
			if a.SuccessChance != offer.SuccessChance || a.HaulMinutes != offer.HaulMinutes || a.Title != offer.Title {
				t.Errorf("%s %s: sent as %+v, offered as %+v", age, c.tier, a.MissionOffer, offer)
			}
		}
	}
}

// TestArmyTerm_FollowsPower: a military grand mission's chance moves with
// the army (0.10 for each unit of power over typical, within +-0.15).
func TestArmyTerm_FollowsPower(t *testing.T) {
	for _, c := range []struct{ power, term float64 }{{100, 0}, {200, 0.10}, {0, -0.10}, {1000, 0.15}, {-5000, -0.15}} {
		if got := armyTerm(c.power, 100); math.Abs(got-c.term) > 1e-9 {
			t.Errorf("power %v moves the chance by %v, want %v", c.power, got, c.term)
		}
	}
	if got := armyTerm(1000, 0); got != 0 {
		t.Errorf("no typical power moves the chance by %v", got)
	}
	// A civil grand mission and a short one are not changed by the army.
	ge := missionReady(t, 3, "iron_age")
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for _, o := range []MissionOffer{
		{Tier: TierShort, SuccessChance: .8}, {Tier: TierLong, SuccessChance: .7}, {Tier: TierGrand, Grand: GrandCivil, SuccessChance: .6},
	} {
		if got := ge.offerSuccessLocked(o); got != o.SuccessChance {
			t.Errorf("%s offer %v became %v with the army", o.Tier, o.SuccessChance, got)
		}
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

// TestMissionCost_IsMinutesOfIncomeInTheMostChargedMaterials: a short
// mission costs 5 minutes, a long one 15 and a civil grand one 45 minutes of
// the town's income, split in half over the two materials the age charges
// most, rounded up; a material with no income is left out.
func TestMissionCost_IsMinutesOfIncomeInTheMostChargedMaterials(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := missionReady(t, 1, age)
		top := chargedMost(ge, age)
		for _, c := range []struct {
			tier, grand string
			minutes     float64
		}{{TierShort, "", 5}, {TierLong, "", 15}, {TierGrand, GrandCivil, 45}} {
			n := plantOffer(ge, c.tier, c.grand)
			cost := ge.MissionCost(n)
			// Every mission also takes its party (TestEveryOfferCostsItsParty).
			if cost["soldiers"] < 1 {
				t.Errorf("%s %s in %s costs no party: %v", c.tier, c.grand, age, cost)
			}
			delete(cost, "soldiers")
			if len(cost) != 2 {
				t.Fatalf("%s %s in %s costs %v, want the two materials %v", c.tier, c.grand, age, cost, top)
			}
			for _, res := range top {
				want := math.Ceil(c.minutes / 2 * ticksPerMinute * ge.townIncomeLocked(res))
				if cost[res] != want {
					t.Errorf("%s %s in %s costs %v %s, want %v (%v minutes of income %v, half)", c.tier, c.grand, age, cost[res], res, want, c.minutes, ge.townIncomeLocked(res))
				}
			}
		}

		// No income of the first: it is left out, and the third does not fill in.
		setIncome(ge, top[0], 0)
		n := plantOffer(ge, TierShort, "")
		cost := ge.MissionCost(n)
		delete(cost, "soldiers")
		if _, ok := cost[top[0]]; ok || len(cost) != 1 || cost[top[1]] == 0 {
			t.Errorf("a short mission in %s with no income of %s costs %v, want only %s", age, top[0], cost, top[1])
		}
		setIncome(ge, top[1], 0)
		cost = ge.MissionCost(n)
		delete(cost, "soldiers")
		if len(cost) != 0 {
			t.Errorf("a short mission in %s with no income at all costs %v besides its party, want nothing", age, cost)
		}
	}
}

// TestMissionCost_MilitaryGrandIsItsParty: a military grand mission costs its
// party (30% of the age's reference army) and nothing else, whatever the town
// makes of soldiers.
func TestMissionCost_MilitaryGrandIsItsParty(t *testing.T) {
	for i, age := range eraFirstAges() {
		ge := missionReady(t, 1, age)
		n := plantOffer(ge, TierGrand, GrandMilitary)
		order, _ := ge.progress.AgeIndex(age)
		want := float64(config.PartySizeOf(config.PartyMilitary, config.RefArmy(order)))
		setIncome(ge, "soldiers", 0.037)
		cost := ge.MissionCost(n)
		if cost["soldiers"] != want || len(cost) != 1 {
			t.Errorf("era %d: a military grand mission in %s costs %v, want %v soldiers and nothing else", i, age, cost, want)
		}
		setIncome(ge, "soldiers", 0)
		if got := ge.MissionCost(n)["soldiers"]; got != want {
			t.Errorf("a military grand mission in %s with no soldier income costs %v soldiers, want %v", age, got, want)
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
// success brings it with the expedition bonus (and a military grand
// mission's pay), a failure 30% of it.
func TestMissionHaul_FormulaOnSuccessAndFailure(t *testing.T) {
	for _, age := range eraFirstAges() {
		for _, c := range []struct{ tier, grand string }{{TierShort, ""}, {TierLong, ""}, {TierGrand, GrandCivil}, {TierGrand, GrandMilitary}} {
			for _, win := range []bool{true, false} {
				ge := missionReady(t, 5, age)
				setSoldiers(ge, 1e12)
				ge.mu.Lock()
				ge.Military.campaignPay = 1.5
				ge.mu.Unlock()
				n := plantOffer(ge, c.tier, c.grand)
				want := haulKeys(ge, age)
				if len(want) == 0 || len(want) > 2 {
					t.Fatalf("%s: the age hauls %v, want one or two resources", age, want)
				}
				if err := ge.LaunchExpedition(n); err != nil {
					t.Fatalf("%s %s: %v", c.tier, c.grand, err)
				}
				active := ge.Military.ActiveByTier(c.tier)
				if len(active.BaseRewards) != len(want) {
					t.Fatalf("%s %s hauls %v, want %v", c.tier, c.grand, active.BaseRewards, want)
				}
				mult := failedLootShare
				if win {
					mult = 1 + ge.expeditionReward()
				}
				if c.grand == GrandMilitary {
					mult *= 1.5
				}
				for _, res := range want {
					base := ge.townIncomeLocked(res) * ticksPerMinute * active.HaulMinutes / 2
					if math.Abs(active.BaseRewards[res]-base) > 1e-9*base {
						t.Errorf("%s %s hauls %v %s, want %v", c.tier, c.grand, active.BaseRewards[res], res, base)
					}
				}

				// Come home: a chance over 1 always wins, under 0 always loses.
				ge.mu.Lock()
				active.TicksLeft = 1
				active.SuccessChance = map[bool]float64{true: 2, false: -1}[win]
				results := ge.Military.Tick(0, ge.expeditionReward())
				ge.mu.Unlock()
				if len(results) != 1 || results[0].Success != win {
					t.Fatalf("%s %s came home as %+v, want success=%v", c.tier, c.grand, results, win)
				}
				for _, res := range want {
					base := active.BaseRewards[res]
					if got := results[0].Rewards[res]; math.Abs(got-base*mult) > 1e-9*base*mult {
						t.Errorf("%s %s (win=%v) brought %v %s, want %v (%v x %v)", c.tier, c.grand, win, got, res, base*mult, base, mult)
					}
				}
			}
		}
	}
}

// TestCampaign_LeavesThePostsFull: a military grand mission sends a party of
// at most the soldiers at home, and takes none from the posts: the army is the
// same before the launch, while the party is out, and after a win or a loss.
// (Parties come out of the posts in a later change.)
func TestCampaign_LeavesThePostsFull(t *testing.T) {
	for _, c := range []struct {
		income float64 // soldiers/tick: cost = ceil(10 min of it)
		held   float64
		sent   int
		win    bool
	}{
		{0.02, 50, 6, true},
		{0.0147, 50, 5, true},
		{0.0033, 50, 1, true},
		{0.02, 50, 6, false},
		{0.5, 4, 0, true}, // the cost (150) is more than the 4 at home: refused, nothing sent
	} {
		ge := missionReady(t, 2, "iron_age")
		setSoldiers(ge, c.held)
		setIncome(ge, "soldiers", c.income)
		n := plantOffer(ge, TierGrand, GrandMilitary)
		cost := int(ge.MissionCost(n)["soldiers"])
		if c.held > float64(cost) {
			if cost != c.sent {
				t.Fatalf("income %v: the mission costs %d soldiers, test wants %d", c.income, cost, c.sent)
			}
		}
		if c.held >= float64(cost) {
			if err := ge.LaunchExpedition(n); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := ge.LaunchExpedition(n); err == nil {
				t.Fatalf("income %v: launched with %v soldiers at home for a cost of %d", c.income, c.held, cost)
			}
			continue
		}
		if active := ge.Military.ActiveByCategory(ExpeditionMilitary); active == nil || active.Soldiers != c.sent {
			t.Fatalf("income %v: the party is %+v, want %d soldiers", c.income, active, c.sent)
		}
		if got := ge.Resources.Get("soldiers"); got != c.held {
			t.Fatalf("soldiers after the launch = %v, want all %v still at their posts", got, c.held)
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
		if got := ge.Resources.Get("soldiers"); got != c.held {
			t.Errorf("sent %d, win=%v: %v soldiers at home, want %v", c.sent, c.win, got, c.held)
		}
	}
}

// TestMissionBoard_TitlesComeFromTheAgesWords: every offer of an age is
// built from that age's words: a tier name, an objective and a place of the
// age, and a twist of the age for its description; a Legendary offer has the
// word in front. A different age has different words.
func TestMissionBoard_TitlesComeFromTheAgesWords(t *testing.T) {
	for _, age := range config.Ages() {
		ge := missionReady(t, 11, age.Key)
		ge.EnsureMissionBoard()
		words, _ := config.MissionWordsFor(age.Key)
		for i, slot := range ge.Military.board.Slots {
			o := slot.Offer
			title := strings.TrimPrefix(o.Title, config.RareTitleWord+" ")
			if o.Rare != (title != o.Title) {
				t.Errorf("%s: %q is rare=%v", age.Key, o.Title, o.Rare)
			}
			which := config.WordsShort
			switch slotTier(i) {
			case TierLong:
				which = config.WordsLong
			case TierGrand:
				which = config.WordsGrandMilitary
				if o.Grand == GrandCivil {
					which = config.WordsGrandCivil
				}
			}
			names, objs := []string{words.Tiers[which]}, words.ObjectivesFor(which)
			ok := false
			for _, obj := range objs {
				for _, place := range words.Places {
					ok = ok || title == names[0]+": "+obj+" "+place
				}
			}
			if !ok {
				t.Errorf("%s slot %d: %q is not built from the age's words", age.Key, i+1, o.Title)
			}
			if !slices.Contains(words.Twists[:], o.Description) {
				t.Errorf("%s slot %d: description %q is not one of the age's twists", age.Key, i+1, o.Description)
			}
		}
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

// TestAutoExpedition_SendsTheBestHaulPerMinute: the Geographic Society sends
// the offer with the most haul minutes for each minute of length among those
// it can pay for, in every era; one it cannot pay for is passed over.
func TestAutoExpedition_SendsTheBestHaulPerMinute(t *testing.T) {
	for _, age := range eraFirstAges() {
		ge := autoExpeditionTestEngine(t, age, 1, 8, 1e9)
		learn(ge, "military_tactics")
		ge.EnsureMissionBoard()
		setSoldiers(ge, 1e12)
		var wantN int
		var bestRate float64
		for i, slot := range ge.Military.board.Slots {
			if slot.Offer.Tier == TierGrand {
				continue // the Society never sends a grand offer
			}
			if rate := slot.Offer.HaulMinutes / slot.Offer.LengthMinutes; rate > bestRate {
				wantN, bestRate = i+1, rate
			}
		}
		if got := ge.pickAutoOffer(); got != wantN {
			t.Fatalf("%s: the picker chose offer %d, want %d (rate %v)", age, got, wantN, bestRate)
		}
		want, _ := ge.Military.BoardOffer(wantN)
		if launched, _ := runAutoTicks(ge, 1); launched != 1 {
			t.Fatalf("%s: %d dispatches on the first tick, want 1", age, launched)
		}
		if active := ge.Military.ActiveByTier(want.Tier); active == nil || active.Title != want.Title {
			t.Errorf("%s: the society sent %+v, want %s", age, active, want.Title)
		}

		// With no soldiers, a military offer is passed over.
		ge2 := autoExpeditionTestEngine(t, age, 1, 8, 1e9)
		learn(ge2, "military_tactics")
		ge2.EnsureMissionBoard()
		setSoldiers(ge2, 0)
		if n := ge2.pickAutoOffer(); n != 0 {
			o, _ := ge2.Military.BoardOffer(n)
			if o.Military() {
				t.Errorf("%s: the society chose %s with no soldiers to send", age, o.Title)
			}
		}
	}
}

// TestMissionRefusals_OnlyTheBoardsOffersCanBeSent: an offer number outside
// 1 to 9 is refused, so is a slot waiting for its next offer, and a refused
// launch takes nothing.
func TestMissionRefusals_OnlyTheBoardsOffersCanBeSent(t *testing.T) {
	ge := missionReady(t, 1, "iron_age")
	for _, n := range []int{0, 10, -3} {
		if err := ge.LaunchExpedition(n); err == nil || !strings.Contains(err.Error(), "no offer") {
			t.Errorf("offer %d: %v, want a refusal that there is no such offer", n, err)
		}
	}
	n := plantOffer(ge, TierShort, "")
	ge.mu.Lock()
	ge.Military.board.Slots[n-1].Offer = nil
	ge.mu.Unlock()
	if err := ge.LaunchExpedition(n); err == nil || !strings.Contains(err.Error(), "taken") {
		t.Errorf("an empty slot: %v, want a refusal that the offer is taken", err)
	}
	if ge.Military.HasActive() {
		t.Error("a refused launch sent something")
	}
}

// TestMissionRolls_SurviveASave: a mission under way keeps its rolled length,
// chance, haul, cost and soldiers through a save and a load, and comes home
// with the haul it left with.
func TestMissionRolls_SurviveASave(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := missionReady(t, 7, "iron_age")
	setSoldiers(ge, 5000)
	if err := ge.LaunchExpedition(plantOffer(ge, TierGrand, GrandMilitary)); err != nil {
		t.Fatal(err)
	}
	before := *ge.Military.ActiveByCategory(ExpeditionMilitary)
	again := saveAndLoad(t, ge, "mission_rolls")
	after := again.Military.ActiveByCategory(ExpeditionMilitary)
	if after == nil {
		t.Fatal("the campaign is gone after the load")
	}
	if after.Title != before.Title || after.Soldiers != before.Soldiers || after.TicksLeft != before.TicksLeft ||
		after.LengthMinutes != before.LengthMinutes || after.SuccessChance != before.SuccessChance || after.HaulMinutes != before.HaulMinutes {
		t.Errorf("after the load %+v, want %+v", *after, before)
	}
	for name, pair := range map[string][2]map[string]float64{
		"base": {before.BaseRewards, after.BaseRewards}, "success": {before.SuccessRewards, after.SuccessRewards},
		"failure": {before.FailureRewards, after.FailureRewards}, "cost": {before.Cost, after.Cost},
	} {
		if len(pair[0]) == 0 && name != "cost" {
			t.Errorf("the %s haul was empty before the save", name)
		}
		if len(pair[0]) != len(pair[1]) {
			t.Errorf("the %s amounts %v became %v", name, pair[0], pair[1])
			continue
		}
		for res, v := range pair[0] {
			if pair[1][res] != v {
				t.Errorf("the %s amount of %s was %v, is %v", name, res, v, pair[1][res])
			}
		}
	}
}
