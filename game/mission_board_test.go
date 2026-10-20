package game

import (
	"encoding/json"
	"math/rand"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Tests for the mission board (mission_board.go): its shape, its clocks, its
// rare offers and its saves. The rolls inside the offers are checked in
// missions_test.go.

// boardEngine is a seeded engine parked in age with the board made.
func boardEngine(t *testing.T, seed int64, age string) *GameEngine {
	t.Helper()
	ge := missionReady(t, seed, age)
	ge.EnsureMissionBoard()
	return ge
}

// stepBoard runs n ticks of the board's clock and the missions: the tick
// counter moves and processExpeditions does the rest, as doTick does.
func stepBoard(ge *GameEngine, n int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for i := 0; i < n; i++ {
		ge.tick++
		ge.processExpeditions()
	}
}

// boardJSON is the board as saved: a fingerprint of every offer and counter.
func boardJSON(t *testing.T, ge *GameEngine) string {
	t.Helper()
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	data, err := json.Marshal(ge.Military.board)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestBoard_ShapeAndNumbering: nine offers, three of each tier, numbered
// short then long then grand; a grand offer is military or civil and no other
// offer has a grand kind; every offer carries an expiry an hour on.
func TestBoard_ShapeAndNumbering(t *testing.T) {
	grands := map[string]int{}
	for seed := int64(1); seed <= 40; seed++ {
		ge := boardEngine(t, seed, "iron_age")
		state := ge.GetState().Military
		if len(state.Board) != 9 || len(state.Expeditions) != 9 {
			t.Fatalf("seed %d: %d slots, %d offers, want 9 and 9", seed, len(state.Board), len(state.Expeditions))
		}
		for i, slot := range state.Board {
			wantTier := []string{TierShort, TierLong, TierGrand}[i/3]
			if slot.Number != i+1 || slot.Tier != wantTier || slot.Offer == nil {
				t.Fatalf("seed %d: slot %d is %+v, want number %d, tier %s and an offer", seed, i, slot, i+1, wantTier)
			}
			o := slot.Offer
			if o.Number != i+1 || o.Key != string(rune('1'+i)) || o.Tier != wantTier {
				t.Errorf("seed %d: offer %d reads %+v", seed, i+1, *o)
			}
			if (o.Tier == TierGrand) != (o.Grand != "") {
				t.Errorf("seed %d: offer %d is tier %s with grand kind %q", seed, i+1, o.Tier, o.Grand)
			}
			if o.Grand != "" {
				grands[o.Grand]++
			}
			if want := minutesToTicks(config.BoardExpireMinutes); o.ExpiresTicks != want {
				t.Errorf("seed %d: offer %d expires in %d ticks, want %d", seed, i+1, o.ExpiresTicks, want)
			}
			if o.Description == "" || o.Name == "" {
				t.Errorf("seed %d: offer %d has no title or description", seed, i+1)
			}
		}
	}
	// 120 grand offers, military or civil by a fair coin: both must turn up,
	// and neither may take nearly all.
	if grands[GrandMilitary] < 30 || grands[GrandCivil] < 30 {
		t.Errorf("120 grand offers split %v, want both kinds in good number", grands)
	}
}

// TestBoard_TakenOfferIsRefilledAfterFiveMinutes: a launched offer leaves its
// slot empty, and exactly five minutes later a new offer of the same tier is
// in it; the other slots are untouched.
func TestBoard_TakenOfferIsRefilledAfterFiveMinutes(t *testing.T) {
	ge := boardEngine(t, 3, "iron_age")
	learn(ge, "exploration", "military_tactics")
	setSoldiers(ge, 1e6)
	before := ge.Military.board.clone()

	const n = 5 // the second long offer
	if err := ge.LaunchExpedition(n); err != nil {
		t.Fatal(err)
	}
	if slot := ge.Military.board.Slots[n-1]; slot.Offer != nil {
		t.Fatalf("a taken offer is still on the board: %+v", *slot.Offer)
	}
	wait := minutesToTicks(config.BoardRefillMinutes)
	if wait != 150 {
		t.Fatalf("five minutes are %d ticks, the test assumes 150", wait)
	}

	stepBoard(ge, wait-1)
	if ge.Military.board.Slots[n-1].Offer != nil {
		t.Fatal("the slot was refilled before five minutes")
	}
	stepBoard(ge, 1)
	got := ge.Military.board.Slots[n-1].Offer
	if got == nil {
		t.Fatal("the slot is still empty after five minutes")
	}
	if got.Tier != TierLong || got.Age != "iron_age" || got.ExpiresTick != ge.tick+minutesToTicks(config.BoardExpireMinutes) {
		t.Errorf("the new offer is %+v, want a long offer of the Iron Age that expires in an hour", *got)
	}
	if got.Title == before.Slots[n-1].Offer.Title && got.LengthMinutes == before.Slots[n-1].Offer.LengthMinutes {
		t.Error("the new offer is the one that was taken")
	}
	for i := range before.Slots {
		if i != n-1 && *before.Slots[i].Offer != *ge.Military.board.Slots[i].Offer {
			t.Errorf("slot %d changed while another was refilled", i+1)
		}
	}
}

// TestBoard_UntakenOfferExpiresAfterAnHour: an offer nobody takes is replaced
// at once, in the tick its hour is up, and not before; the replacement is a
// new offer of the same tier with a fresh hour.
func TestBoard_UntakenOfferExpiresAfterAnHour(t *testing.T) {
	ge := boardEngine(t, 8, "classical_age")
	hour := minutesToTicks(config.BoardExpireMinutes)
	if hour != 1800 {
		t.Fatalf("an hour is %d ticks, the test assumes 1800", hour)
	}
	first := *ge.Military.board.Slots[0].Offer
	stepBoard(ge, hour-1)
	if got := *ge.Military.board.Slots[0].Offer; got != first {
		t.Fatal("an offer was replaced before its hour was up")
	}
	stepBoard(ge, 1)
	got := ge.Military.board.Slots[0].Offer
	if got == nil || *got == first {
		t.Fatalf("the offer is still %+v after an hour", got)
	}
	if got.Tier != TierShort || got.ExpiresTick != ge.tick+hour {
		t.Errorf("the replacement is %+v, want a short offer with a fresh hour", *got)
	}
	// All nine were made on the same tick, so all nine went at once.
	for i, slot := range ge.Military.board.Slots {
		if slot.Offer == nil || slot.Offer.ExpiresTick != ge.tick+hour {
			t.Errorf("slot %d was not replaced with the others: %+v", i+1, slot.Offer)
		}
	}
	// A mission under way is not an offer: it does not expire.
}

// TestBoard_RegeneratedOnAnAgeAdvance: a new age gets a whole new board of
// that age's words, made when the age begins, and the mission under way in
// the old age carries on.
func TestBoard_RegeneratedOnAnAgeAdvance(t *testing.T) {
	ge := boardEngine(t, 4, "stone_age")
	if err := ge.LaunchExpedition(1); err != nil {
		t.Fatal(err)
	}
	old := ge.Military.ActiveByTier(TierShort).Title
	if err := ge.EnterAgeForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	words, _ := config.MissionWordsFor("bronze_age")
	if ge.Military.board.Age != "bronze_age" {
		t.Fatalf("the board is for %q after the advance", ge.Military.board.Age)
	}
	for i, slot := range ge.Military.board.Slots {
		if slot.Offer == nil || slot.Offer.Age != "bronze_age" {
			t.Fatalf("slot %d after the advance: %+v", i+1, slot.Offer)
		}
		tier := strings.TrimPrefix(slot.Offer.Title, config.RareTitleWord+" ")
		found := false
		for _, name := range words.Tiers {
			found = found || strings.HasPrefix(tier, name+": ")
		}
		if !found {
			t.Errorf("slot %d reads %q, which is no tier of the Bronze Age", i+1, slot.Offer.Title)
		}
	}
	if got := ge.Military.ActiveByTier(TierShort); got == nil || got.Title != old {
		t.Errorf("the mission under way is %+v, want %q to carry on", got, old)
	}
	// An age set from outside (a test, the console) is caught on the next tick.
	setAge(ge, "iron_age")
	stepBoard(ge, 1)
	if ge.Military.board.Age != "iron_age" {
		t.Errorf("the board is for %q, want iron_age", ge.Military.board.Age)
	}
}

// TestBoard_OneMissionOfEachTierAtATime: a short, a long and a grand mission
// can be out together; a second of a tier is refused until the first is home.
func TestBoard_OneMissionOfEachTierAtATime(t *testing.T) {
	ge := boardEngine(t, 6, "iron_age")
	learn(ge, "exploration", "military_tactics")
	setSoldiers(ge, 1e6)
	for _, tier := range []string{TierShort, TierLong, TierGrand} {
		first := 1 + 3*map[string]int{TierShort: 0, TierLong: 1, TierGrand: 2}[tier]
		if err := ge.LaunchExpedition(first); err != nil {
			t.Fatalf("the first %s mission: %v", tier, err)
		}
		if err := ge.LaunchExpedition(first + 1); err == nil || !strings.Contains(err.Error(), "already under way") {
			t.Errorf("a second %s mission: %v, want the slot's refusal", tier, err)
		}
	}
	for _, tier := range boardTiers {
		if ge.Military.ActiveByTier(tier) == nil {
			t.Errorf("no %s mission is out", tier)
		}
	}
	// The refused launches took nothing: the offers they named are still up.
	for _, n := range []int{2, 5, 8} {
		if _, err := ge.Military.BoardOffer(n); err != nil {
			t.Errorf("offer %d was taken by a refused launch: %v", n, err)
		}
	}
	// One comes home; its tier is free again.
	ge.mu.Lock()
	ge.Military.ActiveByTier(TierLong).TicksLeft = 1
	ge.processExpeditions()
	ge.mu.Unlock()
	if ge.Military.ActiveByTier(TierLong) != nil {
		t.Fatal("the long mission did not come home")
	}
	if err := ge.LaunchExpedition(6); err != nil {
		t.Errorf("a long mission once the first is home: %v", err)
	}
}

// TestBoard_CivilGrandCostsMaterialsAndMilitaryGrandSoldiers: a civil grand
// offer costs materials and no soldiers, and sends none; a military grand
// offer costs soldiers and sends them.
func TestBoard_CivilGrandCostsMaterialsAndMilitaryGrandSoldiers(t *testing.T) {
	ge := missionReady(t, 2, "iron_age")
	setSoldiers(ge, 40)
	setIncome(ge, "soldiers", 0.02)
	civil := plantOffer(ge, TierGrand, GrandCivil)
	cost := ge.MissionCost(civil)
	if _, soldiers := cost["soldiers"]; soldiers || len(cost) != 2 {
		t.Errorf("a civil grand mission costs %v, want two materials and no soldiers", cost)
	}
	if err := ge.LaunchExpedition(civil); err != nil {
		t.Fatal(err)
	}
	if active := ge.Military.ActiveByTier(TierGrand); active.Soldiers != 0 || active.Military() {
		t.Errorf("the civil mission is out as %+v, want no soldiers and no army term", active)
	}
	if got := ge.Resources.Get("soldiers"); got != 40 {
		t.Errorf("a civil mission took %v soldiers", 40-got)
	}
	// Its odds are its own: the army does not move them.
	ge.mu.Lock()
	ge.Military.ActiveByTier(TierGrand).TicksLeft = 1
	ge.processExpeditions()
	ge.mu.Unlock()

	military := plantOffer(ge, TierGrand, GrandMilitary)
	cost = ge.MissionCost(military)
	if cost["soldiers"] < 1 || len(cost) != 1 {
		t.Errorf("a military grand mission costs %v, want soldiers and nothing else", cost)
	}
	if err := ge.LaunchExpedition(military); err != nil {
		t.Fatal(err)
	}
	if active := ge.Military.ActiveByTier(TierGrand); active.Soldiers != int(cost["soldiers"]) || !active.Military() {
		t.Errorf("the military mission is out as %+v, want %v soldiers", active, cost["soldiers"])
	}
}

// rareRun rolls n offers' rarity off rng, one after the other, on a new
// board.
func rareRun(rng *rand.Rand, n int) []bool {
	var b MissionBoard
	out := make([]bool, n)
	for i := range out {
		out[i] = b.rollRare(rng)
	}
	return out
}

// TestRare_GapFloorAndShare: over a hundred thousand offers off a seeded
// stream, no two Legendary offers are within ten of each other, no run of
// ordinary offers is longer than forty, and about 4% are Legendary (5% is the
// chance, less the ten offers after each one that cannot be).
func TestRare_GapFloorAndShare(t *testing.T) {
	const n = 100000
	run := rareRun(rand.New(rand.NewSource(20261020)), n)
	last, rares, longest := -1, 0, 0
	for i, rare := range run {
		if !rare {
			continue
		}
		rares++
		if last >= 0 {
			between := i - last - 1
			if between < config.RareOfferGap {
				t.Fatalf("Legendary offers %d and %d have only %d between them, want at least %d", last, i, between, config.RareOfferGap)
			}
			longest = max(longest, between)
		}
		last = i
	}
	if longest > config.RareOfferFloor {
		t.Errorf("%d ordinary offers in a row, the next must be Legendary after %d", longest, config.RareOfferFloor)
	}
	if share := float64(rares) / n; share < .03 || share > .05 {
		t.Errorf("%.2f%% of offers are Legendary, want about 4%% (the chance is 5%%)", share*100)
	}
	t.Logf("%d Legendary offers in %d (%.2f%%), longest wait %d", rares, n, float64(rares)/n*100, longest)

	// The rules bind when the dice do not: a stream that never rolls low still
	// gets one every forty-first offer, and one that always rolls low gets
	// one as soon as the gap allows.
	never := rareRun(rand.New(&fixedSource{v: 0.99}), 200)
	always := rareRun(rand.New(&fixedSource{v: 0}), 200)
	for i := range never {
		if want := i%(config.RareOfferFloor+1) == config.RareOfferFloor; never[i] != want {
			t.Fatalf("with no luck, offer %d is Legendary=%v, want %v (every %dst)", i, never[i], want, config.RareOfferFloor+1)
		}
		if want := i%(config.RareOfferGap+1) == 0; always[i] != want {
			t.Fatalf("with every roll low, offer %d is Legendary=%v, want %v (every %dth)", i, always[i], want, config.RareOfferGap+1)
		}
	}
}

// fixedSource is a random source whose Float64 is always about v.
type fixedSource struct{ v float64 }

func (s *fixedSource) Int63() int64 { return int64(s.v * (1 << 63)) }
func (s *fixedSource) Seed(int64)   {}

// TestRare_OfferModifiers: a Legendary offer is 1.5 times as long, 15 points
// less likely (never under 10%) and pays four times the haul, and its title
// starts with the word. Offers are rolled on the same stream twice, once
// made rare and once not, to see the change and nothing else.
func TestRare_OfferModifiers(t *testing.T) {
	ironWords, _ := config.MissionWordsFor("iron_age")
	for _, tier := range boardTiers {
		var ordinary, rare *MissionOffer
		for seed := int64(1); seed < 4000 && (ordinary == nil || rare == nil); seed++ {
			var b MissionBoard
			o := b.makeOffer(rand.New(rand.NewSource(seed)), ironWords, tier, 0)
			if o.Grand == GrandCivil {
				continue
			}
			if o.Rare && rare == nil {
				rare = o
			}
			if !o.Rare && ordinary == nil {
				ordinary = o
			}
		}
		if rare == nil || ordinary == nil {
			t.Fatalf("%s: no Legendary and ordinary pair found", tier)
		}
		tuning, base := tuningOf(*rare)
		if !strings.HasPrefix(rare.Title, config.RareTitleWord+" ") || strings.HasPrefix(ordinary.Title, config.RareTitleWord) {
			t.Errorf("%s: titles %q and %q", tier, rare.Title, ordinary.Title)
		}
		if rare.LengthMinutes < tuning.MinMinutes*1.5-1e-9 || rare.LengthMinutes > tuning.MaxMinutes*1.5+1e-9 {
			t.Errorf("%s: a Legendary length of %v is outside [%v, %v] (1.5 x the tier's range)", tier, rare.LengthMinutes, tuning.MinMinutes*1.5, tuning.MaxMinutes*1.5)
		}
		if rare.SuccessChance < config.RareSuccessFloor || rare.SuccessChance > base+.1-config.RareSuccessCut+1e-9 {
			t.Errorf("%s: a Legendary chance of %v is outside [%v, %v]", tier, rare.SuccessChance, config.RareSuccessFloor, base+.1-config.RareSuccessCut)
		}
		if rare.HaulMinutes < tuning.MinMinutes*tuning.Yield*0.6*4-1e-9 || rare.HaulMinutes > tuning.MaxMinutes*tuning.Yield*1.6*4+1e-9 {
			t.Errorf("%s: a Legendary haul of %v minutes is outside four times the tier's range", tier, rare.HaulMinutes)
		}
	}

	// The same draws, forced both ways, differ by exactly the changes.
	roll := func(forceRare bool) MissionOffer {
		ironWords, _ := config.MissionWordsFor("iron_age")
		var b MissionBoard
		if !forceRare {
			b.Made, b.SinceRare = 20, 3 // a Legendary offer was made lately: this one cannot be
		} else {
			b.Made, b.SinceRare = 60, config.RareOfferFloor // forty ordinary in a row: this one must be
		}
		// The rarity draw is the first (a long offer draws nothing before it);
		// the remaining draws are the same stream in both runs.
		src := rand.New(rand.NewSource(77))
		return *b.makeOffer(src, ironWords, TierLong, 100)
	}
	plain, legend := roll(false), roll(true)
	if plain.Rare || !legend.Rare {
		t.Fatalf("rarity forced wrong: %v and %v", plain.Rare, legend.Rare)
	}
	if legend.Title != config.RareTitleWord+" "+plain.Title || legend.Description != plain.Description {
		t.Errorf("titles %q and %q", plain.Title, legend.Title)
	}
	tuning, _ := config.MissionKindTuning(config.MissionExpedition)
	if want := plain.LengthMinutes * 1.5; absDiff(legend.LengthMinutes, want) > 1e-9 {
		t.Errorf("length %v, want %v (1.5 x %v)", legend.LengthMinutes, want, plain.LengthMinutes)
	}
	if want := max(.10, plain.SuccessChance-.15); absDiff(legend.SuccessChance, want) > 1e-9 {
		t.Errorf("chance %v, want %v (%v less 0.15, not under 0.10)", legend.SuccessChance, want, plain.SuccessChance)
	}
	if want := plain.HaulMinutes * 4; absDiff(legend.HaulMinutes, want) > 1e-9 {
		t.Errorf("haul %v, want %v (4 x %v)", legend.HaulMinutes, want, plain.HaulMinutes)
	}
	// May pass the tier's longest.
	if legend.LengthMinutes <= tuning.MaxMinutes && plain.LengthMinutes > tuning.MaxMinutes/1.5 {
		t.Errorf("a long Legendary offer of %v minutes should pass the tier's %v when its ordinary one was %v", legend.LengthMinutes, tuning.MaxMinutes, plain.LengthMinutes)
	}
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

// TestBoard_LegendaryRulesHoldOnTheLiveBoard: on the engine's own board,
// left to expire for four hundred hours, the Legendary offers keep the rules
// in the order the offers were made (slot order, an hour at a time) and the
// share is in range.
func TestBoard_LegendaryRulesHoldOnTheLiveBoard(t *testing.T) {
	ge := boardEngine(t, 99, "iron_age")
	hour := minutesToTicks(config.BoardExpireMinutes)
	var flags []bool
	for step := 0; step < 400; step++ {
		for _, slot := range ge.Military.board.Slots {
			flags = append(flags, slot.Offer.Rare)
		}
		stepBoard(ge, hour)
	}
	if len(flags) != 3600 {
		t.Fatalf("%d offers seen, want 3600", len(flags))
	}
	last, rares := -1, 0
	for i, rare := range flags {
		if !rare {
			continue
		}
		rares++
		if last >= 0 && i-last-1 < config.RareOfferGap {
			t.Fatalf("Legendary offers %d and %d are %d apart", last, i, i-last)
		}
		if i-last-1 > config.RareOfferFloor {
			t.Fatalf("%d ordinary offers before Legendary offer %d", i-last-1, i)
		}
		last = i
	}
	if share := float64(rares) / float64(len(flags)); share < .025 || share > .055 {
		t.Errorf("%.2f%% of the board's offers were Legendary", share*100)
	}
	if got := ge.Military.board.Made; got != 3600+9 {
		t.Errorf("the board counts %d offers made, want %d", got, 3609)
	}
}

// TestBoard_SameSeedSameBoardOverTime: two engines with one seed make the
// same offers for a long game, taken or not, and a different seed does not.
// Nothing in the generator may depend on a map's order.
func TestBoard_SameSeedSameBoardOverTime(t *testing.T) {
	run := func(seed int64) string {
		ge := boardEngine(t, seed, "classical_age")
		learn(ge, "exploration", "military_tactics")
		setSoldiers(ge, 1e6)
		var sb strings.Builder
		for step := 0; step < 200; step++ {
			stepBoard(ge, 60)
			if step%3 == 0 {
				_ = ge.LaunchExpedition(1 + step%9)
			}
			sb.WriteString(boardJSON(t, ge))
		}
		return sb.String()
	}
	a, b, c := run(31), run(31), run(32)
	if a != b {
		t.Error("two engines with one seed made different boards")
	}
	if a == c {
		t.Error("two seeds made the same boards")
	}
}

// TestBoard_SaveRoundTrip: the board, its counters and a mission under way
// survive a save and a load, and the loaded game goes on making the offers
// the saved one would have.
func TestBoard_SaveRoundTrip(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := boardEngine(t, 7, "iron_age")
	learn(ge, "exploration", "military_tactics")
	setSoldiers(ge, 50)
	setIncome(ge, "soldiers", 0.02)
	if err := ge.LaunchExpedition(8); err != nil {
		if err := ge.LaunchExpedition(7); err != nil {
			t.Fatal(err)
		}
	}
	if err := ge.LaunchExpedition(2); err != nil {
		t.Fatal(err)
	}
	stepBoard(ge, 400)
	before := boardJSON(t, ge)
	if !strings.Contains(before, `"made":`) {
		t.Fatalf("the board carries no counters: %s", before)
	}
	activeBefore := ge.Military.ActiveForSave()

	again := saveAndLoad(t, ge, "board_round_trip")
	if got := boardJSON(t, again); got != before {
		t.Errorf("the board changed through the save:\n before %s\n after  %s", before, got)
	}
	activeAfter := again.Military.ActiveForSave()
	jb, _ := json.Marshal(activeBefore)
	ja, _ := json.Marshal(activeAfter)
	if string(jb) != string(ja) || len(activeAfter) != 2 {
		t.Errorf("the missions under way changed through the save:\n before %s\n after  %s", jb, ja)
	}

	// Both go on for a game's hour and make the same offers.
	stepBoard(ge, 3600)
	stepBoard(again, 3600)
	if a, b := boardJSON(t, ge), boardJSON(t, again); a != b {
		t.Errorf("the loaded game makes other offers than the saved one:\n saved  %s\n loaded %s", a, b)
	}
}

// TestBoard_LegendaryBadge: coming back from a successful Legendary mission
// earns its badge, an ordinary success does not, a failed Legendary mission
// does not, and the Expeditions ladder counts a Legendary success once.
func TestBoard_LegendaryBadge(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	ge.SeedRNG(5)
	learn(ge, "exploration", "military_tactics")
	setAge(ge, "iron_age")
	setSoldiers(ge, 100)
	send := func(rare, win bool) {
		n := plantOffer(ge, TierShort, "")
		ge.mu.Lock()
		ge.Military.board.Slots[n-1].Offer.Rare = rare
		ge.mu.Unlock()
		if err := ge.LaunchExpedition(n); err != nil {
			t.Fatal(err)
		}
		ge.mu.Lock()
		active := ge.Military.ActiveByTier(TierShort)
		active.TicksLeft = 1
		active.SuccessChance = map[bool]float64{true: 2, false: -1}[win]
		ge.processExpeditions()
		ge.mu.Unlock()
	}
	const legend = "special.once_in_a_legend"
	badge, ok := ge.rules.Badge(legend)
	if !ok || badge.Desc != "Come back from a successful Legendary mission." {
		t.Fatalf("the badge is %+v, %v", badge, ok)
	}
	send(false, true)
	wantBadges(t, acct, "an ordinary success", []string{"expedition.expedition_iron_era_scout"}, legend)
	send(true, false)
	wantBadges(t, acct, "a failed Legendary mission", nil, legend)
	send(true, true)
	wantBadges(t, acct, "a Legendary success", []string{legend})
}
