package config

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Pacing: the target curve, and the lookups worked out from the tables.
//
// The numbers of the economy are written in the tables themselves: a
// building's first price, rate and output per copy (the buildings_*.go
// files), a tech's price and research time (research.go), a gate's amounts
// and counts (ages.go), and how long each age is meant to last (AgeTargets,
// below). Nothing rewrites them when the game loads; a golden record of all
// of them (testdata/golden_tables.json) fails the build when one moves.
//
// What is still worked out, from those tables and never the other way round:
//
//   - Price levels (PriceLevels): the median first-copy price of a resource
//     among an age's buildings.
//   - Market parity (MarketRate): the market trades any two construction
//     resources of the current age at the ratio of their price levels, less
//     ExchangeFee.
//   - The reference town (Incomes, TypicalStorages): what five staffed copies
//     of everything make and hold.
//
// A construction resource of an age is one that appears in the first-copy
// price of at least one of that age's buildings (wonders aside), except the
// flow resources below. Its price level is the median of those prices; the
// parity of two resources is the ratio of their price levels.

// TickSeconds is the length of one tick at 1x. It must equal
// game.BaseTickInterval (a game test checks it).
const TickSeconds = 2.0

// PacingStretch is how much longer every age from the Bronze Age on runs than
// on the curve the game's tick clocks were first typed for (about three days
// to the Modern Age). AgeTargets already holds the stretched lengths. Every
// clock counted in ticks while the game runs (the random-event delay, event and boon durations,
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
	if _, ok := AgeTargets[age]; !ok || unstretchedAges[age] {
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

// AgeTargets is the time a player is meant to spend in each age at 1x, from
// entering it to entering the next: about a week to the Modern Age and the
// first prestige. The numbers are written here as they are used. The final
// age's entry only sizes its buildings.
var AgeTargets = map[string]time.Duration{
	"primitive_age":    15 * time.Minute,
	"stone_age":        45 * time.Minute,
	"bronze_age":       234 * time.Minute,
	"iron_age":         390 * time.Minute,
	"classical_age":    546 * time.Minute,
	"medieval_age":     702 * time.Minute,
	"renaissance_age":  936 * time.Minute,
	"colonial_age":     1092 * time.Minute,
	"industrial_age":   1248 * time.Minute,
	"victorian_age":    1404 * time.Minute,
	"electric_age":     26 * time.Hour,
	"atomic_age":       1872 * time.Minute,
	"modern_age":       1872 * time.Minute,
	"information_age":  2184 * time.Minute,
	"digital_age":      2496 * time.Minute,
	"cyberpunk_age":    2808 * time.Minute,
	"fusion_age":       52 * time.Hour,
	"space_age":        3432 * time.Minute,
	"interstellar_age": 3744 * time.Minute,
	"galactic_age":     3744 * time.Minute,
	"quantum_age":      3744 * time.Minute,
	"transcendent_age": 3744 * time.Minute,
}

const (
	// ExchangeFee is what the market keeps on a trade at parity.
	ExchangeFee = 0.2
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

// IsFlowResource reports whether res is one of the flow resources, which have no price level and
// market parity.
func IsFlowResource(res string) bool { return flowResources[res] }

// AgeTargetTicks is AgeTargets[age] in ticks at 1x (0 for an unknown age).
func AgeTargetTicks(age string) float64 {
	return AgeTargets[age].Seconds() / TickSeconds
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
	keys := make([]string, 0, len(cost))
	for res := range cost {
		keys = append(keys, res)
	}
	sort.Strings(keys) // summed in key order: a float sum over a map's own order differs in its last digit from run to run
	u := 0.0
	for _, res := range keys {
		if pl := levels[res]; pl > 0 {
			u += cost[res] / pl
		}
	}
	return u
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
// +85%). It is typed here. An age it leaves
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

// TypicalStorages is what a moderate builder's store holds of each resource
// in each age, as age -> resource -> storage: the resource's base storage
// and FlowCopies copies of every storage building from the Primitive Age up
// to and including the age (the ones that store everything, and any that
// store that resource alone). Techs' storage bonus is left out: it is what
// a town holds for building its vaults, the count the moderate economy
// keeps of everything. Pure.
func TypicalStorages(defs []BuildingDef, resources []ResourceDef, order []string) map[string]map[string]float64 {
	idx := make(map[string]int, len(order))
	for i, a := range order {
		idx[a] = i
	}
	out := make(map[string]map[string]float64, len(order))
	for i, age := range order {
		all := 0.0
		own := map[string]float64{}
		for _, d := range defs {
			if j, ok := idx[d.RequiredAge]; !ok || j > i || d.Category != "storage" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type != "storage" || e.Value <= 0 {
					continue
				}
				if e.Target == "all" {
					all += float64(FlowCopies * e.Value)
				} else {
					own[e.Target] += float64(FlowCopies * e.Value)
				}
			}
		}
		held := make(map[string]float64, len(resources))
		for _, r := range resources {
			held[r.Key] = r.BaseStorage + all + own[r.Key]
		}
		out[age] = held
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
