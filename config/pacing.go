package config

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Pacing: the target curve and the rules derived from it
// (the economy design's Laws 2 and 3).
//
// The whole economy is paced from one table, AgeTargets: how long a player
// should spend in each age at 1x. These rules turn it into numbers:
//
//   - The Payback Rule (normalizeProductionRates): a production building,
//     fully staffed, earns back the price of its first copy in
//     PaybackTicks(age), valued at the age's price parity. For construction
//     resources the rate typed into a lineage file is ignored.
//   - Market parity (MarketRate): the market trades any two construction
//     resources of the current age at parity less ExchangeFee, so trading
//     never beats building and no resource is stranded without a source.
//   - Wonder size (normalizeWonderCosts): each age's wonder, which the next
//     advance requires, costs WonderPriceUnits of its age.
//   - Time caps (normalizeBuildTicks, normalizeResearchTicks): nothing takes
//     longer to build or research than a fixed share of its age's target. A
//     tech's time is a share of its age's research cap, by its kind.
//   - Research budgets (normalizeResearchCosts): an age's techs share one
//     knowledge budget, a share of what the age makes in its target time,
//     split by kind.
//
// A construction resource of an age is one that appears in the first-copy
// price of at least one of that age's buildings (wonders aside), except the
// flow resources below. Its price level is the median of those prices; the
// parity of two resources is the ratio of their price levels.

// TickSeconds is the length of one tick at 1x. It must equal
// game.BaseTickInterval (a game test checks it).
const TickSeconds = 2.0

// AgeTargets is the time a player should spend in each age at 1x, entering
// it to entering the next: baseAgeTargets × AgeStretch, about a week to the
// Modern Age and the first prestige. The final age's entry only sizes its
// buildings.
var AgeTargets = stretchedTargets()

// PacingStretch is how much longer every age from the Bronze Age on runs than
// on the curve its clocks were typed for (baseAgeTargets, about three days to
// the Modern Age). AgeTargets multiplies baseAgeTargets by it, and every
// clock counted in ticks (the random-event delay, event and boon durations,
// raids, routes, expeditions, cooldowns) is multiplied by it through
// StretchTicks, so an age holds as many events, raids and routes as it did
// before, each lasting the same share of it. A future change to the curve is
// this one number, plus the texts the tests flag.
const PacingStretch = 2.6

// unstretchedAges keep their base length: the Primitive and Stone Ages are
// the first hour of the game, and stay one.
var unstretchedAges = map[string]bool{"primitive_age": true, "stone_age": true}

// AgeStretch is the factor age's clocks run at: 1 for the Primitive and
// Stone Ages (and an unknown age), PacingStretch from the Bronze Age on.
func AgeStretch(age string) float64 {
	if _, ok := baseAgeTargets[age]; !ok || unstretchedAges[age] {
		return 1
	}
	return PacingStretch
}

// StretchTicks re-times a clock typed in ticks for age: ticks × AgeStretch,
// rounded to the nearest tick. Anything that counts real time per age (a
// delay, a duration, a cooldown, a cadence) goes through it, so the same
// number of ticks covers the same share of a longer age.
func StretchTicks(age string, ticks int) int {
	return StretchTicksBy(ticks, AgeStretch(age))
}

// StretchTicksBy is StretchTicks with the age's factor given: ticks × s,
// rounded to the nearest tick, and ticks itself at a factor of 1.
func StretchTicksBy(ticks int, s float64) int {
	if s == 1 {
		return ticks
	}
	return int(math.Round(float64(float64(ticks) * s)))
}

// withDuration writes a stretched clock into a flavor text: "{dur}" becomes
// DurationText(ticks). Epoch events and awakenings quote their durations, and
// those follow the curve.
func withDuration(text string, ticks int) string {
	return strings.ReplaceAll(text, "{dur}", DurationText(ticks))
}

// stretchedTargets is baseAgeTargets × AgeStretch, to the second.
func stretchedTargets() map[string]time.Duration {
	out := make(map[string]time.Duration, len(baseAgeTargets))
	for age, d := range baseAgeTargets {
		secs := math.Round(float64(d.Seconds() * AgeStretch(age)))
		out[age] = time.Duration(secs) * time.Second
	}
	return out
}

// baseAgeTargets is the curve the game's tick clocks were written for, before
// the one-week stretch. Edit a target here; AgeTargets follows.
var baseAgeTargets = map[string]time.Duration{
	"primitive_age":    15 * time.Minute,
	"stone_age":        45 * time.Minute,
	"bronze_age":       90 * time.Minute,
	"iron_age":         150 * time.Minute,
	"classical_age":    210 * time.Minute,
	"medieval_age":     270 * time.Minute,
	"renaissance_age":  6 * time.Hour,
	"colonial_age":     7 * time.Hour,
	"industrial_age":   8 * time.Hour,
	"victorian_age":    9 * time.Hour,
	"electric_age":     10 * time.Hour,
	"atomic_age":       12 * time.Hour,
	"modern_age":       12 * time.Hour,
	"information_age":  14 * time.Hour,
	"digital_age":      16 * time.Hour,
	"cyberpunk_age":    18 * time.Hour,
	"fusion_age":       20 * time.Hour,
	"space_age":        22 * time.Hour,
	"interstellar_age": 24 * time.Hour,
	"galactic_age":     24 * time.Hour,
	"quantum_age":      24 * time.Hour,
	"transcendent_age": 24 * time.Hour,
}

const (
	// PaybackDivisor and PaybackEpochExponent set the payback as a share of
	// the age's target: target × epochProgress^exponent / divisor. See
	// PaybackTicks.
	PaybackDivisor       = 16.0
	PaybackEpochExponent = 0.9
	// PaybackLateExponent is the exponent from PaybackLateFrom on (the
	// Victorian Age, the first of the Electric Era): the curve's second
	// segment. See PaybackTicks.
	PaybackLateExponent = 1.25
	PaybackLateFrom     = "victorian_age"
	// BuildTimeDivisor caps construction at target / BuildTimeDivisor, for
	// wonders too: an age's wonder must stand before the next advance.
	BuildTimeDivisor = 6.0
	// StorageBuildTimeDivisor is the cap for storage buildings, which queue
	// one copy at a time (MaxCount) and are bought many times an age.
	StorageBuildTimeDivisor = 48.0
	// ExchangeFee is what the market keeps on a trade at parity.
	ExchangeFee = 0.2
	// WonderPriceUnits is what an age's wonder costs in price units of its
	// age (see priceUnits), spread over its resources in their literal
	// proportions.
	WonderPriceUnits = 40.0
)

// flowResources are never priced for the rules above, even where a building
// costs them: their amounts drive other systems (food feeds workers, faith
// sets morale and catastrophe odds, culture fills its own caps and pays for
// festivals and monuments, soldiers are an army). Their producers keep their
// literal rates, and requirements on them are sized to those rates.
//
// Nanobots are one of them for the second reason alone: a build material
// whose amounts were sized by hand to its producers' typed rates (a Nanobot
// Vat is listed at 6,554 a tick, and a Cyberpunk Age building asks for
// 150K of them where it asks for trillions of everything else). Priced, the
// 150K made a price level of its own, so the payback rule cut the vat to 20
// a tick and the market traded one nanobot for 66M electricity at parity.
// The Modern Age's Nano Foundry, in an age that does not price them, kept
// its 80 a tick: half of what it made was sold, a fifth to a third on top
// of the Digital and Cyberpunk Ages' whole income, and an Ancient Cache's
// 40% of a store, in nanobots, paid for an age outright.
var flowResources = map[string]bool{"food": true, "faith": true, "culture": true, "soldiers": true, "nanobots": true}

// IsFlowResource reports whether res is exempt from the Payback Rule and
// market parity.
func IsFlowResource(res string) bool { return flowResources[res] }

// AgeTargetTicks is AgeTargets[age] in ticks at 1x (0 for an unknown age).
func AgeTargetTicks(age string) float64 {
	return AgeTargets[age].Seconds() / TickSeconds
}

// PaybackTicks is how long a fully staffed producer of age takes to earn back
// its first copy's price.
//
// The share of the target grows through the game: 1/16 in the Primitive Age,
// 1/9 in the Iron Age, 1/6 in the Renaissance, 2/9 in the Victorian, 1/3 in
// the Space Age. A Primitive player has nothing but what they build that
// age, so producers must pay back fast. Later, every age also runs on the
// previous ages' buildings, which keep producing forever, and those pile
// up; each new producer can add less or the age flies by. PaybackAdjust
// stretches it for the odd age the curve leaves too fast.
//
// The exponent was 1.25 while techs added into the bonus pools. With their
// bonuses in a small layer of their own, a building has to carry more of
// an age itself, and the more so the later the age: 0.9 gives a producer
// of the Bronze Age 1.2 times the output it had.
//
// The curve has two segments. 0.9 holds to the Industrial Age. From the
// Victorian Age on the exponent is 1.25 again (PaybackLateExponent): 0.35 of
// the target in the Victorian Age, 0.47 in the Modern, 0.67 in the Space
// Age, 0.75 in the Galactic. Research came off the critical path of every
// age from the Victorian on when the tree was finished, and a full run of
// the bot to a Quantum Age prestige (five seeds) played the ages on the
// 0.9 curve at 0.38 to 0.56 of their targets from the Modern Age on, the
// Information Age aside. The Victorian, Electric and Atomic Ages had been
// set one by one to 1.7, 1.45 and 1.75 times the 0.9 curve; the segment
// gives them 1.63, 1.67 and 1.71 and they have no entry any more.
func PaybackTicks(age string) float64 {
	return paybackTicks(age, AgePositions(AgeOrder()))
}

// paybackTicks is PaybackTicks with each age's position given, so a pass
// over every building reads the age order once.
func paybackTicks(age string, pos map[string]int) float64 {
	adj := 1.0
	if v, ok := PaybackAdjust[age]; ok {
		adj = v
	}
	exp := PaybackEpochExponent
	if i, ok := pos[age]; ok {
		if from, ok := pos[PaybackLateFrom]; ok && i >= from {
			exp = PaybackLateExponent
		}
	}
	return adj * AgeTargetTicks(age) * detmath.Pow(epochProgress(age, pos), exp) / PaybackDivisor
}

// PaybackAdjust multiplies the payback of the ages it lists: the curve is
// smooth and the ages aren't. The Renaissance is where gold income jumps (its
// exchanges buy the stone and steel the age can't make), and once the Storage
// Covenant raised its vault the smoke bot finished it in 0.57x of its target.
// Its producers repaid 1.3x slower, and the other half of that fix was a 30M
// knowledge requirement on its gate. That requirement went when the wonders
// got their keystones. Patronage's price stands in for it on a first run
// (0.94x of the target), but not on known ground, where the market is as
// much faster as everything else and buys the knowledge: there the age ran
// at 0.50x of its target ÷ k, on the edge of its band. So the payback takes
// the gate's half too: 1.7x (about 6.5 hours instead of 3.9). With the
// Stone and Iron Eras' new techs a run arrives stronger (a fifth more
// knowledge, cheaper and quicker building, more food and iron) and the age
// ran at 0.76x of its target on a first run, where it had run at 0.96x: 2.0x.
// Its storage is not a lever: the Renaissance Vault sits on the Storage
// Covenant's line.
//
// The Steel and Electric Eras' techs took research off the critical path of
// their ages: an age that held three or four techs made a run pay a quarter
// to a half of its knowledge for the wonder's keystone, and one that holds
// nine asks for a tenth. What paces those ages now is what their buildings
// make and what their gates ask, so both were set again, on a first run
// with the techs in (the smoke bot, three seeds, medians):
//
//   - Victorian: 0.57x of its target with its producers at 1.4x, 0.67x at
//     1.7x, and 0.78x at 1.7x with a gate of 12 of each building (ages.go;
//     it asked for 10).
//   - Electric: 0.56x on the curve alone before the techs, 0.67x at 1.45x.
//   - Atomic: 0.62x to 0.66x at 1.45x to 1.75x with a gate of 15, 0.63x at
//     1x with a gate of 18, and 0.77x at 1.75x with a gate of 18. It takes
//     both.
//   - Renaissance: 0.62x to 0.68x at 2.0x to 2.25x, and on known ground
//     0.46x of its target ÷ k, under its band. Its buildings are priced in
//     knowledge, so its gate is what slows it: with 12 of each building
//     where it asked for 8 it runs at 0.84x, and at 0.62x on known ground.
//     It keeps its 2.0x.
//   - Bronze: 0.63x on the curve, 0.66x at 1.1x.
//
// A larger gate lengthens its own age and shortens the next, which starts
// on more buildings: the Colonial Age went from 0.82x to 0.68x when the
// Renaissance's gate grew, and the Industrial Age from 0.84x to 0.64x when
// the Colonial Age's did, so that one was put back. A slower payback
// lengthens both.
//
// Six entries was a list that said the curve itself wanted reshaping from
// the Victorian Age on, and it has been (PaybackLateExponent): the
// Victorian, Electric and Atomic Ages sit on the curve now.
//
// The ages after the Atomic are not smooth either, and nothing run per pull
// request plays them: these are set against the deep run (five seeds to a
// Quantum Age prestige, medians), in two passes. On the 0.9 curve the
// Modern Age ran at 0.53x of its target, the Information Age at 0.84x (at
// 0.8x the curve), the Digital at 0.49x, the Cyberpunk at 0.45x (at 0.8x),
// the Fusion at 0.45x, the Space Age at 0.56x, the Interstellar at 0.51x
// and the Galactic at 0.38x. The late segment carries the Modern, Digital,
// Cyberpunk and Interstellar Ages (0.73x, 0.77x, 0.71x at 0.7x of it, and
// 0.80x on the first pass); the rest keep an entry, each a multiple of the
// new curve:
//
//   - Information, 0.45x: the one late age that ran near its target, on
//     the payback it keeps (it had 0.8x of a curve half as steep). It has
//     to buy its way into data and steel before it makes anything, it
//     starts on a Modern Age that now repays more slowly (1.03x on the
//     first pass), and it must not pass 1.2x. Its vault sits on the Storage
//     Covenant's line, so its producers cannot repay faster than they do
//     without a larger vault.
//   - Fusion, 2.2x: 0.66x at 1.5x. It answers little to its own payback: a
//     Fusion Reactor at this rate makes about what a Cyberpunk Age foundry
//     does, the requirement is the smallest of its era (ten of each), and
//     the Cyberpunk Age's producers go on making the electricity it builds
//     with. The Cyberpunk Age on the curve slows it too.
//   - Space, 1.4x: 0.70x at 0.85x.
//   - Galactic, 2.0x: 0.78x at 1.8x, with the largest bonuses in the game.
//   - Quantum and Transcendent, 0.5x: where they were. The bot prestiges on
//     entering the Quantum Age, so nothing measures them, and a price there
//     is fifty times a Galactic one: they wait for a run that plays them.
//
// On these entries the second pass read 0.69x in the Modern Age, 1.06x in
// the Information Age, 0.75x in the Digital, 0.86x in the Cyberpunk, 0.84x
// in the Fusion, 1.01x in the Space Age, 0.97x in the Interstellar and
// 0.82x in the Galactic, with the first run to the Modern Age at 5.47 days.
// A median of five seeds moves by 0.05 or so from one pass to the next in
// an age with two lengths (the Industrial, the Electric, the Information),
// so an age set near a line can read on either side of it: the Electric Age
// read 0.67x, 0.71x and 0.63x on three passes, its payback a seventh longer
// on the last two.
//
// What makes one age quick and its neighbour not is mostly how far its
// prices jump from the age before, which is what a run arrives able to
// pay, and how much of what it builds with the age before goes on making.
// An age answers to its own payback with about half the change (the Modern
// Age: 1.76 times the payback, 1.38 times the length) and passes about a
// third of it on to the next (the Information Age, its own payback
// unchanged, ran 1.23 times longer).
//
// The Industrial Age, 1.15x, is the last age before the late segment, and
// it has two lengths. A run that leaves the Colonial Age with the larger of the
// two stores the bot builds there (170M against 94M) plays it at 0.59x to
// 0.64x of its target, the other at 0.77x to 0.91x. Two seeds of five took
// the quick one before the late segment and three after, which put the
// median at 0.64x.
var PaybackAdjust = map[string]float64{
	"bronze_age":       1.1,
	"renaissance_age":  2.0,
	"industrial_age":   1.15,
	"information_age":  0.45,
	"fusion_age":       2.2,
	"space_age":        1.4,
	"galactic_age":     2.0,
	"quantum_age":      0.5,
	"transcendent_age": 0.5,
}

// epochProgress counts epochs of three ages each, continuously: 1 in the
// Primitive Age, 2 in the Iron Age, 3 in the Renaissance, 1/3 more per age.
// Continuous rather than by epoch so the first age of an epoch doesn't
// inherit buildings that paid back much faster than its own. pos is each
// age's position (AgePositions).
func epochProgress(age string, pos map[string]int) float64 {
	if i, ok := pos[age]; ok {
		return 1 + float64(i)/3
	}
	return 1
}

// AgePositions maps each age key in order to its position. Nothing here
// keeps it: BaseBuildings builds it once per call and hands it to the rules
// that need it, and an engine reads positions from its rules.Set.
func AgePositions(order []string) map[string]int {
	pos := make(map[string]int, len(order))
	for i, k := range order {
		pos[k] = i
	}
	return pos
}

// PriceLevelsByAge returns, per age, the median first-copy price of each
// construction resource across that age's buildings (wonders aside). Pure:
// rules.Compile builds a ruleset's price levels with it.
func PriceLevelsByAge(defs []BuildingDef) map[string]map[string]float64 {
	return priceLevels(defs)
}

func priceLevels(defs []BuildingDef) map[string]map[string]float64 {
	prices := map[string]map[string][]float64{}
	for _, d := range defs {
		if d.RequiredAge == "" || d.Category == "wonder" {
			continue
		}
		for res, c := range d.BaseCost {
			if c <= 0 || flowResources[res] {
				continue
			}
			if prices[d.RequiredAge] == nil {
				prices[d.RequiredAge] = map[string][]float64{}
			}
			prices[d.RequiredAge][res] = append(prices[d.RequiredAge][res], c)
		}
	}
	out := make(map[string]map[string]float64, len(prices))
	for age, byRes := range prices {
		out[age] = make(map[string]float64, len(byRes))
		for res, v := range byRes {
			out[age][res] = median(v)
		}
	}
	return out
}

func median(v []float64) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// roundRate rounds a derived rate to 3 significant figures. Unlike
// roundSignificant it never floors a small positive value at 1.
func roundRate(v float64) float64 {
	if v <= 0 {
		return v
	}
	mag := detmath.Pow(10, 3-math.Ceil(detmath.Log10(v)))
	return math.Round(v*mag) / mag
}

// priceUnits is a price measured in price levels of age: 1.0 is a building
// that costs the median of one resource. Resources without a level (flow
// resources) do not count.
func priceUnits(cost map[string]float64, levels map[string]float64) float64 {
	u := 0.0
	for res, c := range cost {
		if pl := levels[res]; pl > 0 {
			u += c / pl
		}
	}
	return u
}

// normalizeProductionRates applies the Payback Rule. For every non-wonder
// building, each production effect on a construction resource of its age is
// set to
//
//	rate = priceUnits(first copy) × priceLevel(output) / PaybackTicks(age) / n
//
// where n is the number of such outputs, so a building with two outputs
// splits its value between them. Effects on flow resources and on resources
// no building of the age costs keep their literal values. Wonders are left
// alone: they are a one-off, not an investment.
func normalizeProductionRates(defs []BuildingDef, pos map[string]int) []BuildingDef {
	levels := priceLevels(defs)
	for i := range defs {
		d := &defs[i]
		if d.Category == "wonder" || d.RequiredAge == "" {
			continue
		}
		lv := levels[d.RequiredAge]
		pb := paybackTicks(d.RequiredAge, pos)
		if pb <= 0 || len(lv) == 0 {
			continue
		}
		n := 0
		for _, e := range d.Effects {
			if e.Type == "production" && e.Value > 0 && lv[e.Target] > 0 {
				n++
			}
		}
		if n == 0 {
			continue
		}
		units := priceUnits(d.BaseCost, lv)
		// Effects is a fresh literal slice per call; copy anyway so no other
		// def can alias the rewritten values.
		effs := append([]Effect(nil), d.Effects...)
		for j, e := range effs {
			if e.Type == "production" && e.Value > 0 && lv[e.Target] > 0 {
				effs[j].Value = roundRate(units * lv[e.Target] / pb / float64(n))
			}
		}
		d.Effects = effs
	}
	return defs
}

// normalizeBuildTicks caps construction time at a share of the age's target,
// so a required building can never take longer to build than the age is
// meant to last.
func normalizeBuildTicks(defs []BuildingDef) []BuildingDef {
	for i := range defs {
		d := &defs[i]
		t := AgeTargetTicks(d.RequiredAge)
		if t <= 0 || d.BuildTicks <= 0 {
			continue
		}
		div := BuildTimeDivisor
		if d.Category == "storage" {
			div = StorageBuildTimeDivisor
		}
		if limit := int(t / div); d.BuildTicks > limit {
			d.BuildTicks = max(limit, 1)
		}
	}
	return defs
}

// normalizeWonderCosts sizes every wonder to WonderPriceUnits of its age:
// the parts of its price in construction resources are scaled together so
// they add up to that many price units, keeping their proportions. An age's
// wonder must be built before the next advance, so its price is part of the
// gate. Parts in flow resources, or in resources no building of the age
// costs, keep their literal values (they are sized by hand to what the age
// produces).
func normalizeWonderCosts(defs []BuildingDef) []BuildingDef {
	levels := priceLevels(defs)
	for i := range defs {
		d := &defs[i]
		if d.Category != "wonder" {
			continue
		}
		lv := levels[d.RequiredAge]
		units := priceUnits(d.BaseCost, lv)
		if units <= 0 {
			continue
		}
		k := WonderPriceUnits / units
		cost := make(map[string]float64, len(d.BaseCost))
		for res, c := range d.BaseCost {
			if lv[res] > 0 {
				c = roundSignificant(c*k, 2)
			}
			cost[res] = c
		}
		d.BaseCost = cost
	}
	return defs
}

// ResearchTimeDivisor sets an age's research cap: its target divided by
// this. No tech takes longer, so the techs an age offers fit in it (research
// runs one tech at a time).
const ResearchTimeDivisor = 8.0

// Research time by kind (the tech tree's first pacing rule): a tech's
// research time is this share of its age's research cap. The required techs
// are the quick ones, so the spine never holds an age up for long, and a
// capstone is the long one.
const (
	ResearchTimeSpine    = 0.4
	ResearchTimeKeystone = 0.5
	ResearchTimeOptional = 0.5
	ResearchTimeCapstone = 0.8
)

// ResearchTimeShare is the share of its age's research cap a tech of kind
// takes (an unknown kind reads as optional).
func ResearchTimeShare(kind TechKind) float64 {
	switch kind {
	case TechSpine:
		return ResearchTimeSpine
	case TechKeystone:
		return ResearchTimeKeystone
	case TechCapstone:
		return ResearchTimeCapstone
	}
	return ResearchTimeOptional
}

// ResearchCapTicks is age's research cap in ticks at 1x: its target divided
// by ResearchTimeDivisor (0 for an unknown age).
func ResearchCapTicks(age string) float64 {
	return AgeTargetTicks(age) / ResearchTimeDivisor
}

// normalizeResearchTicks sets every tech's research time from its age and
// its kind: ResearchTimeShare of the age's research cap, to the nearest
// tick and one tick at least. The time typed on a tech is not read. A tech
// of an age with no target keeps what it has.
func normalizeResearchTicks(techs []TechDef, kinds map[string]TechKind) []TechDef {
	for i := range techs {
		t := &techs[i]
		limit := ResearchCapTicks(t.Age)
		if limit <= 0 {
			continue
		}
		share := ResearchTimeShare(kinds[t.Key])
		t.ResearchTicks = max(int(math.Round(float64(limit*share))), 1)
	}
	return techs
}

// Research cost by budget (the tech tree's second pacing rule). An age's
// techs share one knowledge budget, ResearchBudget, split by these weights:
// a tech costs budget × its weight ÷ the sum of the weights of its age's
// techs. So a new tech makes every tech of its age a little cheaper, and the
// age's total stays where it was. What a run must research is the cheap
// part: the spine, then the keystone.
const (
	ResearchCostSpine    = 0.6
	ResearchCostKeystone = 0.8
	ResearchCostOptional = 1.0
	ResearchCostCapstone = 1.6
)

// ResearchCostWeight is a tech's share of its age's budget, by kind (an
// unknown kind reads as optional).
func ResearchCostWeight(kind TechKind) float64 {
	switch kind {
	case TechSpine:
		return ResearchCostSpine
	case TechKeystone:
		return ResearchCostKeystone
	case TechCapstone:
		return ResearchCostCapstone
	}
	return ResearchCostOptional
}

// ResearchBudgetShare is the share of the knowledge an age makes in its
// target time that its techs cost together. It is the tuning knob of
// research prices: one number for every age but the two below.
//
// At 0.9 an age's research is spread along the whole of it: the last tech
// comes within reach near the end, so the slot has something to do all age
// and no long stretch goes by with nothing new (the smoke suite's QuietMax).
// It is also what lets a keystone carry the weight the knowledge gates did.
//
// The tree is drawn for about nine techs an age. Until the rest of them
// arrive an age holds three or four (two, in the Cosmic Era), so each costs
// two to three times what it will, and what a wonder waits for is a quarter
// to a half of what its age makes rather than a sixth. A lower share was
// tried and measured (0.3, the same prices as the finished tree): the
// Atomic Age then went 13 hours with nothing new, and on known ground the
// Renaissance ran at 0.43x of its target ÷ k. What the share may ask of an
// age is the smoke suite's Research Covenant (smoke/static_research.go).
const ResearchBudgetShare = 0.9

// researchBudgetShares are the ages that keep their own share. The Primitive
// Age has no keystone and is over in a quarter of an hour: its two techs
// take half of what it makes. The Transcendent Age is the last: its techs
// come early in it, at under a third.
var researchBudgetShares = map[string]float64{
	"primitive_age":    0.5,
	"transcendent_age": 0.3,
}

// ResearchBudgetShareOf is the share of its own knowledge age's techs cost:
// ResearchBudgetShare, or the age's own share where it has one.
func ResearchBudgetShareOf(age string) float64 {
	if v, ok := researchBudgetShares[age]; ok {
		return v
	}
	return ResearchBudgetShare
}

// KnowledgePerHour is the knowledge a well-played game makes in an hour of
// each age at 1x: what the knowledge buildings of the smoke suite's greedy
// bot produce, averaged over the age, on a first run (the progression
// report's knowledge_per_hour_1x, the median of its seeds; knowledge bought
// at the market is not in it). It is an input like AgeTargets, typed here
// and re-measured when the economy moves: research budgets are sized from
// it.
//
// The Renaissance to Atomic Ages were not measured again when their own
// techs arrived (the tree's second content batch), on purpose. The bot
// builds knowledge for the techs an age cannot be left without and takes
// the rest from what is over. With nine techs an age where there were
// three, the required ones cost a third of what they did, the bot built a
// third of the knowledge buildings, and the report's figure fell with them
// (the Colonial Age read 7.6M to 19M where it had read 27M). A number that
// follows its own prices down has no floor, so these stand as measured
// while research was the ages' critical path: what a town that wants every
// tech of its age makes.
//
// The Primitive to Atomic Ages were measured with the Stone and Iron Eras'
// new techs in the tree (the medians of three seeds of two Determinism runs
// of the change that added them, averaged: prices follow the number, and
// the bot's staffing follows the prices). Language and Map Making raise knowledge
// by 10% each in every age after theirs, so every age a first run plays
// read higher than before, by a tenth to a third, and all twelve were
// re-measured, not only the six ages that gained techs. The Electric and
// Atomic Ages still move the most between seeds (141M to 265M in the
// Atomic Age): the bot staffs its knowledge buildings differently as tech
// prices move, so those two are good to a third. A first run prestiges on
// entering the Modern Age, so nothing run per PR measures that age or any
// after it. The Modern to Galactic Ages are what the deep run measured
// (five seeds to a Quantum Age prestige on the finished tree, medians: 188M,
// 286M, 498M, 654M, 853M, 944M, 967M and 1.05B). They had stood at 275M to
// 489M since before the tree was finished, flat from the Fusion Age on,
// where a run makes twice that. The Quantum and Transcendent Ages carry the
// trend on: the bot prestiges on entering the Quantum Age and never plays
// them. On the curve's late segment the ages run longer and the bot staffs
// more research as its prices rise, so the same run read higher again from
// the Space Age on (1.14B, 1.57B and 1.70B in the Space, Interstellar and
// Galactic Ages); these were not moved a second time.
var KnowledgePerHour = map[string]float64{
	"primitive_age":    1.1e3,
	"stone_age":        5.05e3,
	"bronze_age":       13.6e3,
	"iron_age":         45e3,
	"classical_age":    150e3,
	"medieval_age":     1.77e6,
	"renaissance_age":  9.0e6,
	"colonial_age":     28e6,
	"industrial_age":   85e6,
	"victorian_age":    115e6,
	"electric_age":     220e6,
	"atomic_age":       258e6,
	"modern_age":       190e6,
	"information_age":  290e6,
	"digital_age":      500e6,
	"cyberpunk_age":    650e6,
	"fusion_age":       850e6,
	"space_age":        940e6,
	"interstellar_age": 970e6,
	"galactic_age":     1.05e9,
	"quantum_age":      1.1e9,
	"transcendent_age": 1.15e9,
}

// AgeKnowledge is the knowledge age makes in its target time:
// KnowledgePerHour × the target in hours (0 for an unknown age).
func AgeKnowledge(age string) float64 {
	return float64(KnowledgePerHour[age] * AgeTargets[age].Hours())
}

// ResearchBudget is what age's techs cost together, rounding aside: its
// share (ResearchBudgetShareOf) of AgeKnowledge.
func ResearchBudget(age string) float64 {
	return float64(AgeKnowledge(age) * ResearchBudgetShareOf(age))
}

// normalizeResearchCosts prices every tech from its age's budget: budget ×
// the tech's weight ÷ the weights of every tech of that age, to three
// significant figures and a whole number of knowledge, 1 at least. The cost
// typed on a tech is not read. A tech of an age with no budget keeps what it
// has.
func normalizeResearchCosts(techs []TechDef, kinds map[string]TechKind) []TechDef {
	// Summed in table order, so the float sum is the same on every run.
	weights := map[string]float64{}
	for _, t := range techs {
		weights[t.Age] += ResearchCostWeight(kinds[t.Key])
	}
	for i := range techs {
		t := &techs[i]
		budget, sum := ResearchBudget(t.Age), weights[t.Age]
		if budget <= 0 || sum <= 0 {
			continue
		}
		share := ResearchCostWeight(kinds[t.Key]) / sum
		t.Cost = math.Max(math.Round(roundSignificant(float64(budget*share), 3)), 1)
	}
	return techs
}

// FormatRateValue prints a per-tick rate with 3 significant figures and a
// K/M/B/T/Q suffix from a thousand up (0.711, 44.8, 3.34K, 2.72M). It is
// textfmt.Number, the one number format the game uses.
func FormatRateValue(v float64) string {
	return textfmt.Number(v)
}

// The lookups below (PriceLevels, ExchangeRate, MarketRate, MarketPairs,
// FlowIncome, TypicalIncome, and DealPriceLevel, PricedResources and
// MarketOffers in deals.go) describe the tables this package defines and
// keep nothing between calls: each one rebuilds the building table, about a
// millisecond. They are for tests and tools. An engine reads the same
// numbers from its rules.Set, which works them out once with the pure
// functions beside them (PriceLevelsByAge, ExchangeRateAt, MarketRateAt,
// MarketPairsAt, Incomes, FlowDealLevels, MarketOffersAt).

// PriceLevels returns the median first-copy price of each construction
// resource in age (nil for an age with no buildings).
func PriceLevels(age string) map[string]float64 {
	return priceLevels(BaseBuildings())[age]
}

// exchangeLevels is every age's price levels, rebuilt on every call.
func exchangeLevels() map[string]map[string]float64 {
	return priceLevels(BaseBuildings())
}

// ExchangeRate is the market rate from def.From to def.To in age: parity
// minus ExchangeFee when both are construction resources of age, the
// literal BaseRate otherwise.
func ExchangeRate(def ExchangeRateDef, age string) float64 {
	return ExchangeRateAt(def, PriceLevels(age))
}

// ExchangeRateAt is ExchangeRate against an age's price levels.
func ExchangeRateAt(def ExchangeRateDef, lv map[string]float64) float64 {
	from, to := lv[def.From], lv[def.To]
	if from <= 0 || to <= 0 {
		return def.BaseRate
	}
	return roundRate(to / from * (1 - ExchangeFee))
}

// MarketRate is the rate the market pays for one from in age, and whether
// the pair trades at all. Any two construction resources of age trade at
// parity less ExchangeFee; the listed pairs (BaseExchangeRates) trade at
// ExchangeRate. The listed pairs keep their historical behavior of not
// checking MinAge here; MarketPairs is what the UI and the bot offer.
func MarketRate(from, to, age string) (float64, bool) {
	return MarketRateAt(from, to, ExchangeRateByKey(), PriceLevels(age))
}

// MarketRateAt is MarketRate against the listed pairs (keyed "from:to") and
// an age's price levels.
//
// A listed pair is closed in an age that prices what it buys and not what
// it takes (a flow resource aside, which no age prices). The typed rate of
// a listed pair is for goods the age does not build with; what an age does
// build with it sells for its other construction resources, at parity.
// Gold is the case: the Information Age's hubs make it by the hundreds of
// millions, no age after builds with it, and at the typed 0.15 data and
// 0.04 crypto a gold it bought the Digital and Cyberpunk Ages about as
// much data again as they made.
func MarketRateAt(from, to string, listed map[string]ExchangeRateDef, lv map[string]float64) (float64, bool) {
	if from == to {
		return 0, false
	}
	if def, ok := listed[from+":"+to]; ok {
		if listedPairClosed(def, lv) {
			return 0, false
		}
		return ExchangeRateAt(def, lv), true
	}
	if lv[from] > 0 && lv[to] > 0 {
		return ExchangeRateAt(ExchangeRateDef{From: from, To: to}, lv), true
	}
	return 0, false
}

// listedPairClosed reports whether a listed pair does not trade in an age
// with price levels lv: the age prices what the pair buys and not what it
// takes, and what it takes is not a flow resource (see MarketRateAt).
func listedPairClosed(def ExchangeRateDef, lv map[string]float64) bool {
	return lv[def.To] > 0 && lv[def.From] <= 0 && !flowResources[def.From]
}

// MarketPairs lists what the market offers in age: every listed pair whose
// MinAge has been reached and that the age has not closed (MarketRateAt),
// plus every ordered pair of the age's construction resources, each with
// its rate for age as BaseRate. Sorted by from, then to.
func MarketPairs(age string) []ExchangeRateDef {
	return MarketPairsAt(age, BaseExchangeRates(), AgePositions(AgeOrder()), PriceLevels(age))
}

// MarketPairsAt is MarketPairs against the listed pairs, each age's
// position and age's price levels. It leaves listed as it found it.
func MarketPairsAt(age string, listed []ExchangeRateDef, pos map[string]int, lv map[string]float64) []ExchangeRateDef {
	seen := map[string]bool{}
	var out []ExchangeRateDef
	for _, def := range listed {
		if pos[def.MinAge] > pos[age] || listedPairClosed(def, lv) {
			continue
		}
		def.BaseRate = ExchangeRateAt(def, lv)
		seen[def.From+":"+def.To] = true
		out = append(out, def)
	}
	res := make([]string, 0, len(lv))
	for r := range lv {
		res = append(res, r)
	}
	sort.Strings(res)
	for _, from := range res {
		for _, to := range res {
			if from == to || seen[from+":"+to] {
				continue
			}
			def := ExchangeRateDef{From: from, To: to, MinAge: age}
			def.BaseRate = ExchangeRateAt(def, lv)
			out = append(out, def)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

// FlowCopies is the "reasonable production level" of a flow resource: how
// many copies of each of its producers a player who invests moderately in it
// keeps, fully staffed. Flow producers keep hand-set rates (Law 3), so
// anything sized to what an age produces of a flow resource (harbinger
// Appease prices, the Gate Covenant's flow check) is sized to FlowIncome.
const FlowCopies = 5.0

// ProductionAllHeld is the all-production pool a well-played game holds in
// each age, as earned, before the soft cap (ProductionSoftCap): the "+X% all
// production" of its milestones, wonders and monuments added up (0.85 is
// +85%). It is an input like KnowledgePerHour, typed here. An age it leaves
// out holds none.
//
// Until techs got a layer of their own the model read this pool off the
// techs: their all-production bonuses filled it from the Industrial Age on
// and reached +200% in the Electric Age. Techs no longer join the pool.
// Milestones and wonders fill it instead, later: the Industrial to Atomic
// Ages are what the smoke suite's greedy bot held when it left each age on
// a first run (the progression report's production_all_earned, the median
// of its seeds), about two thirds of what a player with every milestone
// can hold. Nothing run per PR goes past the Atomic Age: every age from the
// Modern on is that share of what can be held there. From the Digital Age
// on the share is past the knee, +200%, where the pools once stopped: the
// model holds what is earned and IncomeFactor passes it through the soft
// cap, as the engine does, so +220% counts as +205%.
// The static caps report (smoke.StaticCaps) lists, age by age, the pool of
// a player with every milestone, wonder and monument, and a smoke test
// fails if a number here is above it.
var ProductionAllHeld = map[string]float64{
	"industrial_age":   0.85,
	"victorian_age":    0.95,
	"electric_age":     1.3,
	"atomic_age":       1.45,
	"modern_age":       1.7,
	"information_age":  1.85,
	"digital_age":      2.2,
	"cyberpunk_age":    2.2,
	"fusion_age":       2.35,
	"space_age":        2.45,
	"interstellar_age": 2.6,
	"galactic_age":     3.25,
	"quantum_age":      3.75,
	"transcendent_age": 4.4,
}

// FlowIncome is what a player who invests moderately in the flow resource res
// makes per tick in age at 1x: FlowCopies fully staffed copies of every
// non-wonder producer of res from the Primitive Age up to and including age,
// plus the output of every earlier age's wonder (each advance requires its
// age's wonder, so they stand), plus the flat output of every tech up to
// age, multiplied twice over. First by the all-production pool held by then
// (ProductionAllHeld), through the soft cap as the engine applies it (from
// the Digital Age on the pool is past the knee: output triples, and a
// quarter of the rest comes on top). Then by
// the tech layer, which no cap holds: 1 + the bonus every tech up to age
// gives that resource + the bonus they give all production. Monuments, the
// per-resource bonuses of milestones and wonders, morale and worker upkeep
// are left out. 0 for an unknown age or a resource nothing produces by
// then.
func FlowIncome(res, age string) float64 {
	return Incomes(BaseBuildings(), Technologies(), AgeOrder(), IsFlowResource)[age][res]
}

// TypicalIncome is FlowIncome's "what the age produces" for any resource,
// construction resources included: FlowCopies fully staffed copies of every
// producer of res up to age, every earlier wonder and every tech up to age,
// times the all-production pool held by then, times the tech layer. It is
// what the Storage Covenant (StorageHold) sizes storage against.
func TypicalIncome(res, age string) float64 {
	return Incomes(BaseBuildings(), Technologies(), AgeOrder(), AnyResource)[age][res]
}

// AnyResource accepts every resource: Incomes' filter for TypicalIncome.
func AnyResource(string) bool { return true }

// The Storage Covenant (the economy design's Law 1): the most storage buildable in an
// age must hold at least StorageHold(age) hours of the age's TypicalIncome at
// 1x, for every construction resource of the age. Typical income is a
// moderate investment (five copies of each producer); the smoke bot ends an
// age making one to three times it.
//
//   - StorageHoldHours, 4.5 hours from the Bronze Age on: about three hours
//     of a well-built economy, so a player who checks in every few hours
//     loses little at a cap. Longer absences lean on the build plan, which
//     overflow now pays, and on wonder overflow, so storage keeps its
//     pressure. Pacing v2's away-proofing raised it from 1.5 hours; on the
//     one-week curve most ages already held about 3.9, and the ages that
//     didn't reach 4.5 had their storage per copy raised (Bronze the most).
//   - EarlyStorageHoldHours, 1.5 hours for the Primitive and Stone Ages: the
//     first hour of the game, which keeps its pace (the ages AgeStretch
//     leaves at 1x), where storage is meant to fill fast and be built often.
const (
	StorageHoldHours      = 4.5
	EarlyStorageHoldHours = 1.5
)

// StorageHold is the Storage Covenant's hours for age: EarlyStorageHoldHours
// for the Primitive and Stone Ages, StorageHoldHours from the Bronze Age on.
func StorageHold(age string) float64 {
	if unstretchedAges[age] {
		return EarlyStorageHoldHours
	}
	return StorageHoldHours
}

// BuildingOutputs is the part of Incomes that the moderate economy's own
// buildings make, before any bonus: FlowCopies fully staffed copies of every
// non-wonder producer of each resource include accepts, from the Primitive
// Age up to and including each age, as age -> resource -> output per tick.
// Wonders, techs and the production_all bonus are left out: every town has
// those whatever it builds, so this is the part that measures what a town
// put into a resource (the engine's faith strength reads it). Pure.
func BuildingOutputs(defs []BuildingDef, order []string, include func(string) bool) map[string]map[string]float64 {
	idx := make(map[string]int, len(order))
	for i, a := range order {
		idx[a] = i
	}
	out := make(map[string]map[string]float64, len(order))
	for i, age := range order {
		made := map[string]float64{}
		for _, d := range defs {
			if j, ok := idx[d.RequiredAge]; !ok || j > i || d.Category == "wonder" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "production" && e.Value > 0 && include(e.Target) {
					made[e.Target] += float64(FlowCopies * e.Value)
				}
			}
		}
		out[age] = made
	}
	return out
}

// StaffedOutputs is the part of BuildingOutputs that comes from buildings
// with worker slots: the same sum over the producers that take a crew, as
// age -> resource -> output per tick. Workers add staffing's share of it,
// and a worker output bonus counts on that share alone, so a measure that
// runs a moderate set through the engine's steps needs to know how much of
// the set is staffed (all of faith's buildings, none of culture's). Pure.
func StaffedOutputs(defs []BuildingDef, order []string, include func(string) bool) map[string]map[string]float64 {
	staffed := make([]BuildingDef, 0, len(defs))
	for _, d := range defs {
		if d.WorkerCapacity > 0 {
			staffed = append(staffed, d)
		}
	}
	return BuildingOutputs(staffed, order, include)
}

// IncomeFactor is what Incomes multiplies res's output by in age: the
// all-production pool held by then (ProductionAllHeld), through the soft
// cap as the engine applies it (ProductionSoftCap: in full up to +200%, a
// quarter of every point past it), times the tech layer, which no cap
// holds: 1 + what every tech up to age adds to res + what they add to all
// production. pos is each age's position (AgePositions).
func IncomeFactor(techs []TechDef, pos map[string]int, age, res string) float64 {
	layer := 1.0
	for _, t := range techs {
		if j, ok := pos[t.Age]; !ok || j > pos[age] {
			continue
		}
		for _, e := range t.Effects {
			if e.Kind == EffectAllOutput || e.Kind == EffectOutput && e.Target == res {
				layer += e.Value
			}
		}
	}
	return float64((1 + ProductionSoftCap().Applied(ProductionAllHeld[age])) * layer)
}

// Incomes is FlowIncome's formula for every age in order and every resource
// include accepts, as age -> resource -> income per tick: IsFlowResource
// gives FlowIncome's table, AnyResource TypicalIncome's. Pure.
func Incomes(defs []BuildingDef, techs []TechDef, order []string, include func(string) bool) map[string]map[string]float64 {
	idx := make(map[string]int, len(order))
	for i, a := range order {
		idx[a] = i
	}
	out := make(map[string]map[string]float64, len(order))
	for i, age := range order {
		inc := map[string]float64{}
		for _, d := range defs {
			j, ok := idx[d.RequiredAge]
			if !ok || j > i {
				continue
			}
			for _, e := range d.Effects {
				if e.Type != "production" || e.Value <= 0 || !include(e.Target) {
					continue
				}
				switch {
				case d.Category != "wonder":
					inc[e.Target] += float64(FlowCopies * e.Value)
				case j < i:
					inc[e.Target] += e.Value
				}
			}
		}
		for _, t := range techs {
			if j, ok := idx[t.Age]; !ok || j > i {
				continue
			}
			for _, e := range t.Effects {
				if e.Kind == EffectFlatOutput && e.Value > 0 && include(e.Target) {
					inc[e.Target] += e.Value
				}
			}
		}
		for k := range inc {
			inc[k] = float64(inc[k] * IncomeFactor(techs, idx, age, k))
		}
		out[age] = inc
	}
	return out
}
