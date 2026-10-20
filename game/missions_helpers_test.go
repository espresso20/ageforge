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

// missionAt returns the mission of kind in the era age belongs to.
func missionAt(ge *GameEngine, age, kind string) *ExpeditionDef {
	era := ge.rules.EraOf(age)
	for _, def := range ge.Military.expeditions {
		if def.Era == era && def.Kind == kind {
			d := def
			return &d
		}
	}
	return nil
}

// stockFor sets the stock of every resource the mission costs to twice its
// current cost, so the launch can be paid. Caller must NOT hold the engine lock.
func stockFor(ge *GameEngine, key string) map[string]float64 {
	cost := ge.MissionCost(key)
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
