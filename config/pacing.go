package config

import (
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Pacing: the target curve and the rules derived from it
// (design-and-architecture/economy.md, Laws 2 and 3).
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
//     longer to build or research than a fixed share of its age's target.
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
	s := AgeStretch(age)
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
	PaybackEpochExponent = 1.25
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
var flowResources = map[string]bool{"food": true, "faith": true, "culture": true, "soldiers": true}

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
// 1/7 in the Iron Age, 1/4 in the Renaissance, 1/3 in the Victorian, 2/3 in
// the Space Age. A Primitive player has nothing but what they build that
// age, so producers must pay back fast. Later, every age also runs on the
// previous ages' buildings, which keep producing forever, and those pile
// up; each new producer can add less or the age flies by. PaybackAdjust
// stretches it for the odd age the curve leaves too fast.
func PaybackTicks(age string) float64 {
	adj := 1.0
	if v, ok := PaybackAdjust[age]; ok {
		adj = v
	}
	return adj * AgeTargetTicks(age) * detmath.Pow(epochProgress(age), PaybackEpochExponent) / PaybackDivisor
}

// PaybackAdjust multiplies the payback of the ages it lists: the curve is
// smooth and the ages aren't. The Renaissance is where gold income jumps (its
// exchanges buy the stone and steel the age can't make), and once the Storage
// Covenant raised its vault the smoke bot finished it in 0.57x of its target.
// Its producers repay 1.3x slower (about 1.9 hours instead of 1.5); the other
// half of that fix is its gate's knowledge requirement (config/ages.go).
// Its storage is not a lever: the Renaissance Vault sits on the Storage
// Covenant's line. Keep this list short; a second entry means the curve
// itself wants changing.
var PaybackAdjust = map[string]float64{
	"renaissance_age": 1.3,
}

// epochProgress counts epochs of three ages each, continuously: 1 in the
// Primitive Age, 2 in the Iron Age, 3 in the Renaissance, 1/3 more per age.
// Continuous rather than by epoch so the first age of an epoch doesn't
// inherit buildings that paid back much faster than its own.
func epochProgress(age string) float64 {
	if i, ok := ageIndex()[age]; ok {
		return 1 + float64(i)/3
	}
	return 1
}

var (
	ageIndexOnce sync.Once
	ageIndexMap  map[string]int
)

// ageIndex caches each age's position. BaseBuildings derives every payback
// from it and the engine calls BuildingByKey on hot paths, so rebuilding
// Ages() per building would be expensive. Read-only.
func ageIndex() map[string]int {
	ageIndexOnce.Do(func() {
		ageIndexMap = map[string]int{}
		for i, k := range AgeOrder() {
			ageIndexMap[k] = i
		}
	})
	return ageIndexMap
}

// priceLevels returns, per age, the median first-copy price of each
// construction resource across that age's buildings (wonders aside).
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
func normalizeProductionRates(defs []BuildingDef) []BuildingDef {
	levels := priceLevels(defs)
	for i := range defs {
		d := &defs[i]
		if d.Category == "wonder" || d.RequiredAge == "" {
			continue
		}
		lv := levels[d.RequiredAge]
		pb := PaybackTicks(d.RequiredAge)
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

// ResearchTimeDivisor caps a tech's research time at its age's target
// divided by this, so the handful of techs an age offers fit in it (research
// runs one tech at a time).
const ResearchTimeDivisor = 8.0

// normalizeResearchTicks applies the research-time cap.
func normalizeResearchTicks(techs []TechDef) []TechDef {
	for i := range techs {
		t := &techs[i]
		if limit := int(AgeTargetTicks(t.Age) / ResearchTimeDivisor); limit > 0 && t.ResearchTicks > limit {
			t.ResearchTicks = limit
		}
	}
	return techs
}

// FormatRateValue prints a per-tick rate with 3 significant figures and a
// K/M/B/T/Q suffix from a thousand up (0.711, 44.8, 3.34K, 2.72M). It is
// textfmt.Number, the one number format the game uses.
func FormatRateValue(v float64) string {
	return textfmt.Number(v)
}

// PriceLevels returns the median first-copy price of each construction
// resource in age (nil for an age with no buildings).
func PriceLevels(age string) map[string]float64 {
	return priceLevels(BaseBuildings())[age]
}

// ExchangeRate is the market rate from def.From to def.To in age: parity
// minus ExchangeFee when both are construction resources of age, the
// literal BaseRate otherwise.
func ExchangeRate(def ExchangeRateDef, age string) float64 {
	return exchangeRate(def, exchangeLevels()[age])
}

func exchangeRate(def ExchangeRateDef, lv map[string]float64) float64 {
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
	if from == to {
		return 0, false
	}
	if def, ok := exchangeByKey()[from+":"+to]; ok {
		return ExchangeRate(def, age), true
	}
	lv := exchangeLevels()[age]
	if lv[from] > 0 && lv[to] > 0 {
		return exchangeRate(ExchangeRateDef{From: from, To: to}, lv), true
	}
	return 0, false
}

// MarketPairs lists what the market offers in age: every listed pair whose
// MinAge has been reached, plus every ordered pair of the age's construction
// resources, each with its rate for age as BaseRate. Sorted by from, then to.
func MarketPairs(age string) []ExchangeRateDef {
	order := ageIndex()
	seen := map[string]bool{}
	var out []ExchangeRateDef
	for _, def := range BaseExchangeRates() {
		if order[def.MinAge] > order[age] {
			continue
		}
		def.BaseRate = ExchangeRate(def, age)
		seen[def.From+":"+def.To] = true
		out = append(out, def)
	}
	lv := exchangeLevels()[age]
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
			def.BaseRate = exchangeRate(def, lv)
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

var (
	exchangeByKeyOnce sync.Once
	exchangeByKeyMap  map[string]ExchangeRateDef
)

func exchangeByKey() map[string]ExchangeRateDef {
	exchangeByKeyOnce.Do(func() { exchangeByKeyMap = ExchangeRateByKey() })
	return exchangeByKeyMap
}

var (
	exchangeLevelsOnce sync.Once
	exchangeLevelsMap  map[string]map[string]float64
)

// exchangeLevels caches the price levels for the market, which reads them on
// every state snapshot. Config is static, so computing them once is safe;
// the map is shared and must be treated as read-only.
func exchangeLevels() map[string]map[string]float64 {
	exchangeLevelsOnce.Do(func() { exchangeLevelsMap = priceLevels(BaseBuildings()) })
	return exchangeLevelsMap
}

// FlowCopies is the "reasonable production level" of a flow resource: how
// many copies of each of its producers a player who invests moderately in it
// keeps, fully staffed. Flow producers keep hand-set rates (Law 3), so
// anything sized to what an age produces of a flow resource (harbinger
// Appease prices, the Gate Covenant's flow check) is sized to FlowIncome.
const FlowCopies = 5.0

// ProductionAllCap is the most the production_all bonus can multiply output
// by. It must equal the engine's productionCap (a game test checks).
const ProductionAllCap = 3.0

// FlowIncome is what a player who invests moderately in the flow resource res
// makes per tick in age at 1x: FlowCopies fully staffed copies of every
// non-wonder producer of res from the Primitive Age up to and including age,
// plus the output of every earlier age's wonder (each advance requires its
// age's wonder, so they stand), plus the flat output of every tech up to
// age, all multiplied by the production_all bonus held by then: every tech up
// to age and every earlier age's wonder, capped at ProductionAllCap as the
// engine caps it (from the Electric Age on the cap is reached, and output
// triples). Monuments, milestones, morale and worker upkeep are left out. 0
// for an unknown age or a resource nothing produces by then.
func FlowIncome(res, age string) float64 {
	return flowIncomes()[age][res]
}

var (
	flowIncomeOnce sync.Once
	flowIncomeMap  map[string]map[string]float64
)

// flowIncomes caches FlowIncome for every age and flow resource. The harbinger
// prices read it on every state snapshot. Read-only.
func flowIncomes() map[string]map[string]float64 {
	flowIncomeOnce.Do(func() {
		flowIncomeMap = computeFlowIncomes(BaseBuildings(), Technologies(), AgeOrder())
	})
	return flowIncomeMap
}

// TypicalIncome is FlowIncome's "what the age produces" for any resource,
// construction resources included: FlowCopies fully staffed copies of every
// producer of res up to age, every earlier wonder and every tech up to age,
// times the production_all bonus held by then. It is what the Storage
// Covenant (StorageHoldHours) sizes storage against.
func TypicalIncome(res, age string) float64 {
	return typicalIncomes()[age][res]
}

var (
	typicalIncomeOnce sync.Once
	typicalIncomeMap  map[string]map[string]float64
)

func typicalIncomes() map[string]map[string]float64 {
	typicalIncomeOnce.Do(func() {
		typicalIncomeMap = computeIncomes(BaseBuildings(), Technologies(), AgeOrder(), func(string) bool { return true })
	})
	return typicalIncomeMap
}

// StorageHoldHours is the Storage Covenant (economy.md, Law 1): the most
// storage buildable in an age must hold at least this many hours of the age's
// TypicalIncome at 1x, for every construction resource of the age. Typical
// income is a moderate investment (five copies of each producer); the smoke
// bot ends an age making one to three times it, so 1.5 hours of it is about
// an hour of a well-built economy: a player who checks in hourly loses
// nothing at a cap, and one who checks in less often leans on the build plan
// and wonder overflow rather than on storage, which keeps its pressure.
const StorageHoldHours = 1.5

func computeFlowIncomes(defs []BuildingDef, techs []TechDef, order []string) map[string]map[string]float64 {
	return computeIncomes(defs, techs, order, IsFlowResource)
}

// computeIncomes is FlowIncome's formula for every resource include accepts.
func computeIncomes(defs []BuildingDef, techs []TechDef, order []string, include func(string) bool) map[string]map[string]float64 {
	idx := make(map[string]int, len(order))
	for i, a := range order {
		idx[a] = i
	}
	out := make(map[string]map[string]float64, len(order))
	for i, age := range order {
		inc := map[string]float64{}
		bonus := 0.0 // production_all
		for _, d := range defs {
			j, ok := idx[d.RequiredAge]
			if !ok || j > i {
				continue
			}
			for _, e := range d.Effects {
				if d.Category == "wonder" && j < i && e.Type == "bonus" && e.Target == "production_all" {
					bonus += e.Value
				}
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
				if e.Type == "production" && e.Value > 0 && include(e.Target) {
					inc[e.Target] += e.Value
				}
				if e.Type == "bonus" && e.Target == "production_all" {
					bonus += e.Value
				}
			}
		}
		mult := math.Min(1+bonus, ProductionAllCap)
		for k := range inc {
			inc[k] *= mult
		}
		out[age] = inc
	}
	return out
}
