package game

import "github.com/espresso20/ageforge/config"

// The army's experience. Every mission that comes home gives it some (or
// takes some, on a failure); the rank it reaches is read at the moment a
// mission is sent, a board is made or an offer rolled (config/army.go has
// the numbers). Experience belongs to the run: prestige and a fallen
// civilization start a new military manager, so the army is Green again.

// Experience is the army's experience.
func (mm *MilitaryManager) Experience() float64 { return mm.experience }

// Rank is the army's rank index, 0 (Green) to 4 (Legendary).
func (mm *MilitaryManager) Rank() int { return config.ArmyRankOf(mm.experience) }

// rareFactor is the rank's factor on the chance of a Legendary offer.
func (mm *MilitaryManager) rareFactor() float64 { return config.ArmyRareFactor(mm.Rank()) }

// TiersOpen is how many tiers of the board are open in age for this army:
// the era's, and one more from Veteran (config.MissionTiersOpen).
func (mm *MilitaryManager) TiersOpen(age string) int {
	order := 0
	if era, ok := mm.rules.Era(mm.rules.EraOf(age)); ok {
		order = era.Order
	}
	return max(mm.minTiers, config.MissionTiersOpen(order, mm.Rank()))
}

// gainExperience puts what a returned mission gave the army (negative for a
// loss) to its experience, never below 0. It returns what the experience
// really changed by, and the rank before and after.
func (mm *MilitaryManager) gainExperience(delta float64) (change float64, before, after int) {
	before = mm.Rank()
	next := max(0, mm.experience+delta)
	change = next - mm.experience
	mm.experience = next
	return change, before, mm.Rank()
}
