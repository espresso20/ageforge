package game

import (
	"maps"
	"math"
	"slices"
)

// prestigeCueLocked estimates the time to the current age's gate. Callers hold
// ge.mu for writing; it uses only managers' lock-free internal methods.
func (ge *GameEngine) prestigeCueLocked() PrestigeCue {
	cue := PrestigeCue{Age: ge.age}
	if len(ge.Stats.AgeLog) == 0 {
		return cue
	}
	elapsed := ge.tick - ge.Stats.AgeLog[0].Tick
	if elapsed < 0 {
		return cue
	}
	// Sorted walks keep the float sums in one order, so identical games agree.
	need := make(map[string]float64)
	reqs := ge.progress.GetRequirementsForNext(ge.age)
	for _, key := range slices.Sorted(maps.Keys(reqs)) {
		wanted := reqs[key]
		missing := max(0, wanted-ge.Buildings.GetCount(key)-ge.Buildings.GetQueueCount(key, ge.buildQueue))
		if missing == 0 {
			continue
		}
		cost, ok := ge.Buildings.BuildBatchCost(key, missing, ge.buildQueue)
		if !ok {
			return cue
		}
		for _, res := range slices.Sorted(maps.Keys(cost)) {
			need[res] += cost[res]
		}
	}
	if wonder := ge.progress.WonderForAge(ge.age); wonder != "" && ge.Buildings.GetCount(wonder) < 1 {
		def := ge.Buildings.defs[wonder]
		bank := ge.Buildings.GetWonderBank(wonder)
		for _, res := range slices.Sorted(maps.Keys(def.BaseCost)) {
			need[res] += math.Max(0, def.BaseCost[res]-bank[res])
		}
	}

	remaining := 0.0
	for res, amount := range need {
		shortfall := math.Max(0, amount-ge.Resources.Get(res))
		if shortfall == 0 {
			continue
		}
		rate := ge.Resources.GetRate(res)
		if rate <= 0 {
			return cue
		}
		remaining = math.Max(remaining, shortfall/rate)
	}
	if keystone := ge.rules.Keystone(ge.age); keystone != "" && !ge.Research.IsResearched(keystone) {
		def, ok := ge.Research.defs[keystone]
		if !ok {
			return cue
		}
		var ticks int
		if ge.Research.currentTech == keystone {
			if ge.researchPerTick <= 0 {
				return cue
			}
			ticks = ge.researchTicksFor(def.Cost - ge.Research.progress)
		} else {
			if ge.researchIfStarted <= 0 {
				return cue
			}
			ticks = ge.researchTicksIfStarted(def.Cost)
		}
		remaining = math.Max(remaining, float64(ticks))
	}
	if remaining <= 0 {
		return cue
	}
	l := ge.Prestige.ProductionMultiplier()
	next := ge.Prestige.NextProductionMultiplier(ge.age)
	cue.Multiplier, cue.NextMultiplier = l, next
	if next <= l {
		return cue
	}
	fresh := (float64(elapsed) + remaining) * l / next
	cue.FinishTicks = int(math.Ceil(remaining))
	cue.FreshTicks = int(math.Ceil(fresh))
	cue.Seen = ge.prestigeCueSeenAge == ge.age
	cue.Ready = ge.Prestige.CanPrestige(ge.age, ge.progress.ageOrder()) && fresh < remaining
	return cue
}

// MarkPrestigeCueSeen records that the current age's cue was shown.
func (ge *GameEngine) MarkPrestigeCueSeen() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.prestigeCueSeenAge = ge.age
	ge.prestigeCue.Seen = true
}
