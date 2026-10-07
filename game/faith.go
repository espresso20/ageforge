package game

// Faith strength: what the rolls that read faith go by (a fated doom's
// strike, the Last Passage's roll, the epoch roll at an advance).
//
// It is the faith a town holds that its own faith buildings made, measured
// against what FaithFullSets moderate sets of faith buildings would have
// made over the same run. A moderate set is config.FlowCopies fully staffed
// copies of every faith building the age has, the economy Appease is priced
// on. Under FaithMidAt of that is the low band, over FaithHighAbove the high
// one.
//
// Three running totals for the run carry it (FaithSave), each fed by the
// faith rate the engine works out every tick, so they move exactly as the
// store does: tick by tick at the town's own bonuses, morale and Era Mastery
// speed, and at OfflineEfficiency while the game is closed.
//
//   - Moderate: what a moderate set would have made in this town.
//   - Own: what the town's own faith buildings made.
//   - Other: the rest of the town's faith income: wonders, techs, events.
//
// Strength is devotion × share kept / FaithFullSets:
//
//   - devotion is Own / Moderate: how the town's faith buildings have done
//     against the moderate set (1 is a moderate town);
//   - the share kept is the faith held over Own + Other, the faith the
//     town's income made this run, at most 1. Spending or losing faith takes
//     its share of the strength; a windfall (an event's gift, loot) can make
//     up for faith spent but counts for nothing past that.
//
// Put the other way, the faith that counts is the faith held times Own /
// (Own + Other): the store is taken to be made up as the run's faith was.
//
// What that gives, in every age and at every moment a roll can fall:
//
//   - A moderate town (devotion 1) reads 1/4.5 = 22%, the low band: the odds
//     the game was tuned and measured at, when the bands read the fill of a
//     store faith could never fill and every town sat in the low band
//     (smoke's TestFaithNeverFilledTheGeneralStore keeps that arithmetic).
//   - The middle band starts at a devotion of 1.125 and the top band above
//     3.375: twice the moderate set reads 44%, three times 67%, and three
//     and a half times 78%.
//   - Faith every town is handed counts for nothing. Stonehenge makes 0.6
//     faith a tick where a moderate Bronze Age set makes 0.07, and every
//     player must build it to advance; the Sistine Chapel and Theology are
//     the same. A measure on all the faith held would hand a band to whoever
//     built the wonder, and one that set the wonders' faith against the
//     moderate town's would take ten times the faith buildings to reach the
//     top band in the Iron Era. Their faith is the player's to spend all the
//     same.
//   - It follows the run, not the age: nothing jumps at an advance, so the
//     epoch roll reads what the player was shown before advancing, and a
//     hoard carried in from earlier ages counts for exactly what the moderate
//     set's does.
//
// Faith is only read: nothing is clamped or taken.
const (
	FaithMidAt     = 0.25
	FaithHighAbove = 0.75
	// FaithFullSets is how many moderate sets of faith buildings' worth of
	// faith reads as full strength. Over 4, so a moderate town is under the
	// middle band with room to spare (an eighth more faith buildings); under
	// 4.67, so the top band takes less than three and a half times the
	// moderate set.
	FaithFullSets = 4.5
)

// FaithBand is the band a faith strength falls in: what every roll that
// reads faith goes by.
type FaithBand string

const (
	FaithBandLow  FaithBand = "low"  // strength under FaithMidAt
	FaithBandMid  FaithBand = "mid"  // from FaithMidAt to FaithHighAbove
	FaithBandHigh FaithBand = "high" // over FaithHighAbove
)

// FaithBandAt is the band of faith strength strength. Pure.
func FaithBandAt(strength float64) FaithBand {
	switch {
	case strength < FaithMidAt:
		return FaithBandLow
	case strength > FaithHighAbove:
		return FaithBandHigh
	}
	return FaithBandMid
}

// FaithSave is the run's faith measure, persisted as GameSave.FaithMeasure:
// the three running totals faith strength is read from. They start at zero
// with a run and only grow.
type FaithSave struct {
	// Moderate is what a moderate set of faith buildings would have made in
	// this town over the run.
	Moderate float64 `json:"moderate,omitempty"`
	// Own is what the town's own faith buildings have made.
	Own float64 `json:"own,omitempty"`
	// Other is what the rest of the town's faith income has made: wonders,
	// techs, events.
	Other float64 `json:"other,omitempty"`
}

// FaithStrengthOf is the faith strength of a town holding held faith whose
// run totals are m, in [0,1]: FaithDevotionOf × FaithKeptOf / FaithFullSets.
// 0 until the town's faith buildings have made something and there is a
// moderate set to measure them against. Pure: the rule itself, for the
// engine and for the smoke suite's static check.
func FaithStrengthOf(held float64, m FaithSave) float64 {
	strength := float64(FaithDevotionOf(m)*FaithKeptOf(held, m)) / FaithFullSets
	if strength > 1 {
		return 1
	}
	return strength
}

// FaithDevotionOf is how the town's faith buildings compare with a moderate
// set over the run: Own / Moderate (0 with nothing to measure against).
// Pure.
func FaithDevotionOf(m FaithSave) float64 {
	if m.Moderate <= 0 || m.Own <= 0 {
		return 0
	}
	return m.Own / m.Moderate
}

// FaithKeptOf is the share of the faith the town's income has made this run
// that it still holds, capped at 1 (a windfall can put more in the store
// than the income made). 0 with no income yet. Pure.
func FaithKeptOf(held float64, m FaithSave) float64 {
	made := m.Own + m.Other
	if held <= 0 || made <= 0 {
		return 0
	}
	if held >= made {
		return 1
	}
	return held / made
}

// faithStrength returns the town's faith strength now, in [0,1]. Read-only.
func (ge *GameEngine) faithStrength() float64 {
	return FaithStrengthOf(ge.Resources.Get("faith"), ge.faithMeasure)
}

// faithRates are the two parts of the faith rate the measure follows, per
// tick: what the town's own faith buildings make and what a moderate set
// would make in their place. recalculateRates sets them with the rate.
type faithRates struct {
	own, moderate float64
}

// noteFaithRates works out faithRates from the pieces recalculateRates has
// in hand. ownBase is the town's faith production before bonuses from every
// building that is not a wonder (staffing counted, ruins included) and
// ownByWorkers the part of it its workers add; through carries a base
// production and its workers' part through every step the faith rate goes
// through after that (morale, the all-production and faith bonuses, the
// worker output bonus, trade, the Cosmic Legacy, Era Mastery). The moderate
// set is config.FlowCopies copies of every faith building of the age, fully
// staffed, through the same steps.
func (ge *GameEngine) noteFaithRates(ownBase, ownByWorkers float64, through func(base, byWorkers float64) float64) {
	moderateBase := ge.rules.FlowBuildingOutput("faith", ge.age)
	ge.faithRate = faithRates{
		own:      through(ownBase, ownByWorkers),
		moderate: through(moderateBase, float64(moderateBase*staffedShare)),
	}
}

// accrueFaith adds ticks worth of the current faith rates to the run's
// measure: one for a tick, OfflineEfficiency times the step for the offline
// catch-up, as the store itself is credited. Under the write lock, after the
// rates were applied.
func (ge *GameEngine) accrueFaith(ticks float64) {
	r := ge.Resources.resources["faith"]
	if r == nil || ticks <= 0 {
		return
	}
	ge.faithMeasure.Moderate += float64(ge.faithRate.moderate * ticks)
	ge.faithMeasure.Own += float64(ge.faithRate.own * ticks)
	if other := r.Rate - ge.faithRate.own; other > 0 {
		ge.faithMeasure.Other += float64(other * ticks)
	}
}

// clearFaithRun starts the faith measure over with a new run (a prestige, a
// Succumb, a reset). Under the write lock.
func (ge *GameEngine) clearFaithRun() {
	ge.faithMeasure = FaithSave{}
}

// faithSaveCopy is the measure for a save: nil until it holds something, so
// a new run's save keeps its bytes.
func (ge *GameEngine) faithSaveCopy() *FaithSave {
	if ge.faithMeasure == (FaithSave{}) {
		return nil
	}
	m := ge.faithMeasure
	return &m
}

// restoreFaithState loads the faith measure. A save written before the
// measure existed has none: its town is taken as a moderate one that has
// kept its faith (Own = Moderate = the faith it holds), which reads 1 /
// FaithFullSets, the low band every roll read before, and moves from there
// with what the town does. Under the write lock, after the resources are
// restored.
func (ge *GameEngine) restoreFaithState(save *GameSave) {
	ge.faithMeasure = FaithSave{}
	if save.FaithMeasure != nil {
		ge.faithMeasure = *save.FaithMeasure
		for _, v := range []*float64{&ge.faithMeasure.Moderate, &ge.faithMeasure.Own, &ge.faithMeasure.Other} {
			if !(*v >= 0) || *v > maxFaithMeasure { // a hand-edited save: NaN, negative or absurd
				*v = 0
			}
		}
		return
	}
	if save.FaithMeasured {
		return
	}
	if held := ge.Resources.Get("faith"); held > 0 {
		ge.faithMeasure = FaithSave{Moderate: held, Own: held}
	}
}

// maxFaithMeasure bounds a loaded measure: far past anything a run makes.
const maxFaithMeasure = 1e300

// SetFaithMeasureForTest puts ge's faith measure and its faith in place: a
// test hook for other packages (the UI's panels, the smoke suite's checks).
// The store is widened to hold the faith. Not reachable from play. Takes the
// write lock.
func (ge *GameEngine) SetFaithMeasureForTest(held float64, m FaithSave) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.setFaithMeasure(held, m)
}

// setFaithMeasure is the lock-free body of SetFaithMeasureForTest.
func (ge *GameEngine) setFaithMeasure(held float64, m FaithSave) {
	ge.Resources.UnlockResource("faith")
	if ge.Resources.GetStorage("faith") < held {
		ge.Resources.LoadStorage(map[string]float64{"faith": held})
	}
	ge.Resources.LoadAmounts(map[string]float64{"faith": held})
	ge.faithMeasure = m
}
