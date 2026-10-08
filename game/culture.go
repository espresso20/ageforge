package game

// Culture strength: what the roll that reads culture goes by (the tier of a
// good epoch event, at an advance into a new era).
//
// It is faith strength's rule (faith.go) for culture: the culture a town
// holds that its own culture buildings made, measured against what
// CultureFullSets moderate sets of culture buildings would have made over
// the same run. A moderate set is config.FlowCopies fully staffed copies of
// every culture building the age has. The same three running totals carry
// it (CultureSave), fed by the culture rate the engine works out every
// tick, and strength is devotion × share kept / CultureFullSets:
//
//   - devotion is Own / Moderate: how the town's culture buildings have done
//     against the moderate set (1 is a moderate town);
//   - the share kept is the culture held over Own + Other, the culture the
//     town's income made this run, at most 1. Spending culture (a festival,
//     a Black Market deal, a monument, an Appease) takes its share of the
//     strength; culture from outside the income (bought at the market, an
//     event's gift) can make up for culture spent and counts for nothing
//     past that.
//
// The tier it opens:
//
//   - A moderate town (devotion 1) reads 1/4.5 = 22%: Minor events only,
//     which is where every town stood when the tier read the fill of
//     culture's store. Culture is kept in the general store, and a moderate
//     town's whole run of it never filled 8% of the least store an era's
//     gate leaves a player holding (under 1% from the Modern Age on); three
//     and a half times the culture buildings reached 26% at best, once.
//     The one way over the old 40% was the market, which sells 3 culture a
//     gold: a few minutes of an age's gold filled the store to any mark.
//   - Major events open above CultureMajorAbove: a devotion over 1.8, about
//     twice the moderate set (twice reads 44%).
//   - The Legendary event opens above CultureLegendaryAbove: a devotion
//     over 3.375 (three and a half times reads 78%), and then on
//     cultureLegendaryChance of the rolls.
//   - Culture every town is handed (a wonder's, a tech's) counts for
//     nothing, as with faith, and bought culture does not either.
//   - It follows the run, not the age: nothing jumps at an advance, so the
//     roll reads what the player was shown before advancing.
//
// Culture is only read: nothing is clamped or taken.
const (
	CultureMajorAbove     = 0.40
	CultureLegendaryAbove = 0.75
	// CultureFullSets is how many moderate sets of culture buildings' worth
	// of culture reads as full strength: faith's FaithFullSets, so twice the
	// moderate set is over CultureMajorAbove and three and a half times it
	// over CultureLegendaryAbove.
	CultureFullSets = 4.5
	// cultureLegendaryChance is the share of good epoch rolls, with the
	// Legendary tier open, that draw from every tier.
	cultureLegendaryChance = 0.15
)

// CultureTier is the best tier of good epoch event a culture strength opens.
type CultureTier string

const (
	CultureTierMinor     CultureTier = "minor"     // strength up to CultureMajorAbove
	CultureTierMajor     CultureTier = "major"     // over CultureMajorAbove
	CultureTierLegendary CultureTier = "legendary" // over CultureLegendaryAbove
)

// CultureTierAt is the best tier culture strength strength opens. Pure.
func CultureTierAt(strength float64) CultureTier {
	switch {
	case strength > CultureLegendaryAbove:
		return CultureTierLegendary
	case strength > CultureMajorAbove:
		return CultureTierMajor
	}
	return CultureTierMinor
}

// CultureSave is the run's culture measure, persisted as
// GameSave.CultureMeasure: faith's three running totals (FaithSave), for
// culture.
type CultureSave = FaithSave

// CultureStrengthOf is the culture strength of a town holding held culture
// whose run totals are m, in [0,1]: CultureDevotionOf × CultureKeptOf /
// CultureFullSets. 0 until the town's culture buildings have made something
// and there is a moderate set to measure them against. Pure: the rule
// itself, for the engine and for the smoke suite's static check.
func CultureStrengthOf(held float64, m CultureSave) float64 {
	strength := float64(CultureDevotionOf(m)*CultureKeptOf(held, m)) / CultureFullSets
	if strength > 1 {
		return 1
	}
	return strength
}

// CultureDevotionOf is how the town's culture buildings compare with a
// moderate set over the run: Own / Moderate (0 with nothing to measure
// against). Pure.
func CultureDevotionOf(m CultureSave) float64 { return FaithDevotionOf(m) }

// CultureKeptOf is the share of the culture the town's income has made this
// run that it still holds, capped at 1. 0 with no income yet. Pure.
func CultureKeptOf(held float64, m CultureSave) float64 { return FaithKeptOf(held, m) }

// cultureStrength returns the town's culture strength now, in [0,1].
// Read-only.
func (ge *GameEngine) cultureStrength() float64 {
	return CultureStrengthOf(ge.Resources.Get("culture"), ge.cultureMeasure)
}

// noteCultureRates is noteFaithRates for culture: what the town's own
// culture buildings make per tick and what a moderate set would make in
// their place, each through the steps the culture rate goes through. Culture
// buildings take no crew, so the moderate set's worker part is only what
// its staffed buildings (none, on the core rules) would add.
func (ge *GameEngine) noteCultureRates(ownBase, ownByWorkers float64, through func(base, byWorkers float64) float64) {
	moderateBase := ge.rules.FlowBuildingOutput("culture", ge.age)
	ge.cultureRate = faithRates{
		own:      through(ownBase, ownByWorkers),
		moderate: through(moderateBase, float64(ge.rules.FlowStaffedOutput("culture", ge.age)*staffedShare)),
	}
}

// accrueCulture adds ticks worth of the current culture rates to the run's
// measure, as accrueFaith does for faith. Under the write lock, after the
// rates were applied.
func (ge *GameEngine) accrueCulture(ticks float64) {
	r := ge.Resources.resources["culture"]
	if r == nil || ticks <= 0 || !ge.Resources.unlocked["culture"] {
		return
	}
	ge.cultureMeasure.Moderate += float64(ge.cultureRate.moderate * ticks)
	ge.cultureMeasure.Own += float64(ge.cultureRate.own * ticks)
	if other := r.Rate - ge.cultureRate.own; other > 0 {
		ge.cultureMeasure.Other += float64(other * ticks)
	}
}

// clearCultureRun starts the culture measure over with a new run (a
// prestige, a Succumb, a reset). Under the write lock.
func (ge *GameEngine) clearCultureRun() {
	ge.cultureMeasure = CultureSave{}
}

// cultureSaveCopy is the measure for a save: nil until it holds something,
// so a new run's save keeps its bytes.
func (ge *GameEngine) cultureSaveCopy() *CultureSave {
	if ge.cultureMeasure == (CultureSave{}) {
		return nil
	}
	m := ge.cultureMeasure
	return &m
}

// restoreCultureState loads the culture measure. A save written before the
// measure existed has none: its town is taken as a moderate one that has
// kept its culture (Own = Moderate = the culture it holds), which reads 1 /
// CultureFullSets, Minor events only, and moves from there with what the
// town does. Under the write lock, after the resources are restored.
func (ge *GameEngine) restoreCultureState(save *GameSave) {
	ge.cultureMeasure = CultureSave{}
	if save.CultureMeasure != nil {
		ge.cultureMeasure = *save.CultureMeasure
		for _, v := range []*float64{&ge.cultureMeasure.Moderate, &ge.cultureMeasure.Own, &ge.cultureMeasure.Other} {
			if !(*v >= 0) || *v > maxFaithMeasure { // a hand-edited save: NaN, negative or absurd
				*v = 0
			}
		}
		return
	}
	if save.CultureMeasured {
		return
	}
	if held := ge.Resources.Get("culture"); held > 0 {
		ge.cultureMeasure = CultureSave{Moderate: held, Own: held}
	}
}

// SetCultureMeasureForTest puts ge's culture measure and its culture in
// place: a test hook for other packages. The store is widened to hold the
// culture. Not reachable from play. Takes the write lock.
func (ge *GameEngine) SetCultureMeasureForTest(held float64, m CultureSave) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.setCultureMeasure(held, m)
}

// setCultureMeasure is the lock-free body of SetCultureMeasureForTest.
func (ge *GameEngine) setCultureMeasure(held float64, m CultureSave) {
	ge.Resources.UnlockResource("culture")
	if ge.Resources.GetStorage("culture") < held {
		ge.Resources.LoadStorage(map[string]float64{"culture": held})
	}
	ge.Resources.LoadAmounts(map[string]float64{"culture": held})
	ge.cultureMeasure = m
}
