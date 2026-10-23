package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestSteelAndElectricMechanics: the eight mechanic numbers the second
// content batch added, each where the game uses it, with its tech and
// without. A learn here goes through the same rebuild a finished research
// does, and recalculateRates hands the managers their terms.
func TestSteelAndElectricMechanics(t *testing.T) {
	with := func(techs ...string) *GameEngine {
		ge := newSeededEngine(11)
		learn(ge, techs...)
		ge.mu.Lock()
		ge.recalculateRates()
		ge.mu.Unlock()
		return ge
	}
	plain := with()

	// Architecture: a wonder builds in three quarters of its time, on top of
	// the cut of all construction; any other building keeps its time.
	arch := with("architecture")
	wonder := config.BuildingDef{BuildTicks: 4680, Category: "wonder"}
	house := config.BuildingDef{BuildTicks: 4680, Category: "housing"}
	if a, b := plain.buildTicksLocked(wonder), arch.buildTicksLocked(wonder); a != 4680 || b != 3510 {
		t.Errorf("a 4680 tick wonder takes %d ticks, %d with Architecture; want 4680 and 3510", a, b)
	}
	if got := arch.buildTicksLocked(house); got != 4680 {
		t.Errorf("a 4680 tick house takes %d ticks with Architecture, want 4680", got)
	}
	both := with("architecture", "the_wheel") // The Wheel: construction takes 5% less time
	if got := both.buildTicksLocked(wonder); got != 3334 {
		t.Errorf("a 4680 tick wonder takes %d ticks with Architecture and a 5%% cut of construction, want 3334 (three quarters, then 95%%)", got)
	}
	if got := WonderBuildTicks(4680, both.GetState().Research, 1); got != 3334 {
		t.Errorf("the panels' time for that wonder is %d ticks, want the engine's 3334", got)
	}
	if got := WonderBuildTicks(4680, plain.GetState().Research, 1); got != 4680 {
		t.Errorf("the panels' time for a wonder with no tech is %d ticks, want 4680", got)
	}

	// Baroque Arts: a festival's bonus lasts a quarter longer. The wait
	// between festivals is another number, and stays.
	baroque := with("baroque_arts")
	if a, b := plain.FestivalStatus(), baroque.FestivalStatus(); b.BuffTicks != a.BuffTicks*5/4 || b.CooldownTicks != a.CooldownTicks {
		t.Errorf("a festival lasts %d ticks, %d with Baroque Arts (cooldown %d and %d); want a quarter longer and the same wait",
			a.BuffTicks, b.BuffTicks, a.CooldownTicks, b.CooldownTicks)
	}

	// Embassies: a gift earns half as much again, rounded down to a point.
	emb := with("embassies")
	if a, b := plain.Diplomacy.GiftGain(), emb.Diplomacy.GiftGain(); a != GiftOpinion || b != 22 {
		t.Errorf("a gift earns %d opinion, %d with Embassies; want %d and 22", a, b, GiftOpinion)
	}
	if got := emb.GetState().Diplomacy.GiftGain; got != 22 {
		t.Errorf("the Factions panel reads %d opinion a gift with Embassies, want 22", got)
	}
	emb.mu.Lock()
	emb.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true}
	if _, err := emb.Diplomacy.SendGift("riverlands_tribes", config.Billion); err != nil {
		t.Fatal(err)
	}
	if got := emb.Diplomacy.factions["riverlands_tribes"].Opinion; got != 22 {
		t.Errorf("one gift with Embassies left opinion at %d, want 22", got)
	}
	emb.mu.Unlock()

	// Concert of Nations: an ally adds a quarter more to its specialty, and
	// the Factions panel lists the bonus with it.
	concert := with("concert_of_nations")
	for _, ge := range []*GameEngine{plain, concert} {
		ge.mu.Lock()
		ge.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true, Opinion: AllyOpinion, Status: "allied"}
		ge.mu.Unlock()
	}
	listed := plain.Diplomacy.factionDefs["riverlands_tribes"].TradeBonus
	if a, b := plain.Diplomacy.GetTradeBonus("food"), concert.Diplomacy.GetTradeBonus("food"); a != listed || math.Abs(b-float64(listed*1.25)) > 1e-12 {
		t.Errorf("the Riverlands add %v to food, %v with the Concert of Nations; want %v and a quarter more", a, b, listed)
	}
	if got := concert.GetState().Diplomacy.Factions["riverlands_tribes"].TradeBonus; math.Abs(got-float64(listed*1.25)) > 1e-12 {
		t.Errorf("the Factions panel lists the Riverlands' bonus as %v with the Concert of Nations, want %v", got, float64(listed*1.25))
	}

	// Interchangeable Parts: an upgrade costs 15% less, after its own 4% cut
	// of what the new copy costs. Building new is not an upgrade.
	parts := with("interchangeable_parts")
	for _, ge := range []*GameEngine{plain, parts} {
		ge.mu.Lock()
		ge.age = "industrial_age" // the Manor is an older building here
		ge.Buildings.counts["manor"] = 1
		ge.mu.Unlock()
	}
	full := plain.Buildings.UpgradeCopyCost("manor", 0, nil)
	cut := parts.Buildings.UpgradeCopyCost("manor", 0, nil)
	if len(cut) == 0 {
		t.Fatal("no price for a Manor's upgrade")
	}
	parts.mu.Lock()
	parts.Buildings.upgradeCost = 1
	uncut := parts.Buildings.UpgradeCopyCost("manor", 0, nil)
	parts.mu.Unlock()
	newCopy := parts.Buildings.CopyCost("tenement", 0, nil) // after the tech's own cut of building costs
	for _, res := range sortedKeys(full) {
		if want := math.Floor(float64(float64(newCopy[res]*config.UpgradeShare) * 0.85)); cut[res] != want || cut[res] >= full[res] {
			t.Errorf("upgrading a Manor costs %v %s with Interchangeable Parts (%v without the tech, %v with its cut of building costs alone), want %v",
				cut[res], res, full[res], uncut[res], want)
		}
	}

	// General Staff: a campaign takes 15% less time. A scouting expedition
	// has its own number, and keeps its time.
	staff := with("general_staff")
	launch := func(ge *GameEngine, key string) (ticks int, listedMin int) {
		t.Helper()
		ge.mu.Lock()
		ge.age = "victorian_age"
		ge.pushMechanics()
		ge.mu.Unlock()
		n := keyOffer(ge, key)
		ge.mu.Lock()
		defer ge.mu.Unlock()
		offer, err := ge.Military.BoardOffer(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ge.Military.LaunchOffer(offer, ge.tick, 0, nil, nil); err != nil {
			t.Fatal(err)
		}
		ticks = ge.Military.ActiveByTier(offer.Tier).TicksLeft
		ge.Military.setActive(offer.Tier, nil)
		_, infos := ge.boardViewLocked(0)
		for _, info := range infos {
			if info.Number == n {
				listedMin = info.DurationMin
			}
		}
		return ticks, listedMin
	}
	// victorian_age opens the Electric Era: its campaign and its scout.
	campaign, scouting := "industrial_war", "aerial_survey"
	a, aMin := launch(plain, campaign)
	b, bMin := launch(staff, campaign)
	if b != int(float64(float64(a)*0.85)) || bMin != int(float64(float64(aMin)*0.85)) {
		t.Errorf("the %s campaign takes %d ticks, %d with a General Staff (listed from %d and %d); want 15%% less of each", campaign, a, b, aMin, bMin)
	}
	if a, _ := launch(plain, scouting); true {
		if b, _ := launch(staff, scouting); a != b {
			t.Errorf("the %s expedition takes %d ticks, %d with a General Staff; want the same", scouting, a, b)
		}
	}

	// The Military-Industrial Complex: a campaign brings back a fifth more,
	// won or lost. Scouting does not. It also gives every military building a
	// tenth more posts.
	mic := with("military_industrial_complex")
	for _, ge := range []*GameEngine{plain, mic} {
		ge.mu.Lock()
		ge.Buildings.counts["garrison"] = 3
		ge.postAllSoldiers()
		ge.recalculateRates()
		ge.mu.Unlock()
	}
	base := plain.Buildings.defs["garrison"].Soldiers
	if a, b := plain.soldierTotal(), mic.soldierTotal(); a != 3*base || b != 3*int(math.Ceil(float64(base)*1.1)) {
		t.Errorf("three Garrisons post %d soldiers, %d with the Military-Industrial Complex; want %d and %d (a tenth more posts)", a, b, 3*base, 3*int(math.Ceil(float64(base)*1.1)))
	}
	// A mission is sent the way a player sends it, so its haul is sized on
	// the town's income and the campaign pay is read at the launch.
	loot := func(ge *GameEngine, key string) float64 {
		t.Helper()
		learn(ge, "exploration", "military_tactics")
		ge.mu.Lock()
		ge.age = "victorian_age"
		ge.pushMechanics()
		age, _ := ge.rules.Age(ge.age)
		for _, b := range age.UnlockBuildings {
			def, found := ge.rules.Building(b)
			if !found {
				continue
			}
			for _, effect := range def.Effects {
				if effect.Type == "production" {
					ge.Resources.UnlockResource(effect.Target)
					ge.Resources.SetRate(effect.Target, 1)
				}
			}
		}
		ge.mu.Unlock()
		setSoldiers(ge, 100)
		n := keyOffer(ge, key)
		stockFor(ge, n)
		if err := ge.LaunchExpedition(n); err != nil {
			t.Fatalf("send %s: %v", key, err)
		}
		ge.mu.Lock()
		defer ge.mu.Unlock()
		category := ExpeditionScouting
		if oldKinds[key] == config.MissionCampaign {
			category = ExpeditionMilitary
		}
		active := ge.Military.ActiveByCategory(category)
		active.TicksLeft = 1
		active.SuccessChance = 0.95 // the cap; the roll below is a 0.9
		active.Luck = 0.9
		res := ge.Military.Tick(0, 0)
		if len(res) != 1 || !res[0].Success {
			t.Fatalf("%s did not come home a success: %+v", key, res)
		}
		total := 0.0
		for _, v := range res[0].Rewards {
			total += v
		}
		return total
	}
	if a, b := loot(plain, campaign), loot(mic, campaign); a <= 0 || math.Abs(b-float64(a*1.20)) > 1e-6*a {
		t.Errorf("the %s campaign brings back %v, %v with the Military-Industrial Complex; want a fifth more", campaign, a, b)
	}
	if a, b := loot(plain, scouting), loot(mic, scouting); a != b {
		t.Errorf("the %s expedition brings back %v, %v with the Military-Industrial Complex; want the same", scouting, a, b)
	}
}
