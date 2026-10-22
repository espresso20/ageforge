package game

import "github.com/espresso20/ageforge/config"

// setIncome gives res an income of perTick per tick, as townIncome reads it,
// and leaves its stock alone. The missions size their costs and hauls from
// it. Caller must NOT hold the engine lock.
func setIncome(ge *GameEngine, res string, perTick float64) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Resources.UnlockResource(res)
	k := ge.speedK()
	if k <= 0 {
		k = 1
	}
	if r := ge.Resources.resources[res]; r != nil {
		r.Rate = perTick * k
		r.Breakdown.PrestigeRate, r.Breakdown.EventRate, r.Breakdown.FoodDrain = 0, 0, 0
	}
}

// ticksPerMinute is how many engine ticks make a minute of game time.
const ticksPerMinute = 60 / config.TickSeconds

// plantOffer puts a plain offer of tier (and grand kind, for a grand tier) in
// the first slot of that tier on the board, and returns its number. The offer
// is the middle of its kind's ranges, with the kind's base odds, so a test
// can rely on it and a launch can be paid. The board is made first if it is
// not. Caller must NOT hold the engine lock.
func plantOffer(ge *GameEngine, tier, grand string) int {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.ensureMissionBoardLocked()
	first := 0
	for i := range boardTiers {
		if boardTiers[i] == tier {
			first = i * config.BoardOffersPerTier
		}
	}
	offer := testOffer(ge, tier, grand)
	ge.Military.board.Slots[first] = boardSlot{Offer: &offer}
	return first + 1
}

// testOffer is a plain offer of tier (and grand kind): the middle of its
// kind's ranges, with the kind's base odds, made in the engine's age. It
// touches nothing, so it can be built under the engine lock.
func testOffer(ge *GameEngine, tier, grand string) MissionOffer {
	offer := MissionOffer{Tier: tier, Grand: grand, Age: ge.age, Place: "the Test Hills", Description: "A test offer.", ExpiresTick: ge.tick + 1<<30}
	tuning, _ := config.MissionKindTuning(offer.Kind())
	offer.Title = "Test " + tier + " " + grand
	offer.LengthMinutes = (tuning.MinMinutes + tuning.MaxMinutes) / 2
	offer.SuccessChance = tuning.BaseSuccess
	if grand == GrandCivil {
		offer.SuccessChance = config.CivilGrandSuccess
	}
	offer.HaulMinutes = offer.LengthMinutes * tuning.Yield * 1.1
	return offer
}

// oldKinds says which kind of mission each of the fixed era missions the
// board replaced was, so a test written against one can send the same kind.
var oldKinds = map[string]string{
	"scouting_party": config.MissionScout, "road_survey": config.MissionScout, "coastal_survey": config.MissionScout,
	"aerial_survey": config.MissionScout, "data_run": config.MissionScout, "street_recon": config.MissionScout,
	"orbital_survey": config.MissionScout,
	"long_hunt":      config.MissionExpedition, "trade_caravan": config.MissionExpedition, "naval_expedition": config.MissionExpedition,
	"rail_expedition": config.MissionExpedition, "satellite_survey": config.MissionExpedition, "neon_heist": config.MissionExpedition,
	"deep_space_expedition": config.MissionExpedition,
	"raid_the_rival_clan":   config.MissionCampaign, "siege_campaign": config.MissionCampaign, "colonial_campaign": config.MissionCampaign,
	"industrial_war": config.MissionCampaign, "cyber_campaign": config.MissionCampaign, "fusion_plant_assault": config.MissionCampaign,
	"galactic_conquest": config.MissionCampaign,
}

// keyOffer plants an offer of the kind a fixed era mission was (a scout is a
// short offer, an expedition a long one and a campaign a grand military one)
// and returns its number. Caller must NOT hold the engine lock.
func keyOffer(ge *GameEngine, key string) int {
	switch oldKinds[key] {
	case config.MissionExpedition:
		return plantOffer(ge, TierLong, "")
	case config.MissionCampaign:
		return plantOffer(ge, TierGrand, GrandMilitary)
	}
	return plantOffer(ge, TierShort, "")
}

// ActiveByCategory is the mission under way for a category, for the tests:
// the short mission (else the long one) for ExpeditionScouting, the grand one
// for ExpeditionMilitary.
func (mm *MilitaryManager) ActiveByCategory(category string) *ActiveExpedition {
	if category == ExpeditionMilitary {
		return mm.active[TierGrand]
	}
	if a := mm.active[TierShort]; a != nil {
		return a
	}
	return mm.active[TierLong]
}

// clearActive empties every tier's slot.
func (mm *MilitaryManager) clearActive() {
	for _, tier := range boardTiers {
		mm.active[tier] = nil
	}
}

// stockFor sets the stock of every resource offer n costs to twice its
// current cost, so the launch can be paid. Caller must NOT hold the engine lock.
func stockFor(ge *GameEngine, n int) map[string]float64 {
	cost := ge.MissionCost(n)
	for res, amount := range cost {
		setIncomeKeep := ge.townIncomeLocked(res)
		setResource(ge, res, 2*amount)
		setIncome(ge, res, setIncomeKeep)
	}
	return cost
}

// townIncomeLocked reads townIncome taking the read lock.
func (ge *GameEngine) townIncomeLocked(res string) float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.townIncome(res)
}

// stoneEraReady puts the engine in age (an age of the Stone Era) with
// Military Tactics learned, an income and stock of stone and wood, and
// soldiers with an income, so a scout and a campaign can both be paid.
// Caller must NOT hold the engine lock.
func stoneEraReady(ge *GameEngine, age string) {
	learn(ge, "military_tactics") // the tech that opens campaigns
	setAge(ge, age)
	setResource(ge, "stone", 1000)
	setResource(ge, "wood", 1000)
	setSoldiers(ge, 20)
	setIncome(ge, "stone", 0.5)
	setIncome(ge, "wood", 0.5)
	setIncome(ge, "soldiers", 0.01)
}
