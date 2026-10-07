package game

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// The bonus truth guard.
//
// Every bonus the game promises a player is switched on in a real engine and
// the thing its text names is measured: a resource's rate, storage, housing,
// research time, a building's price, game speed, the Defense Rating, loot.
// TestBonusTruth fails when a promise does not show up at the size promised.
//
// How one promise is measured. Two readings of the same engine, the bonus
// off and then on, and nothing else changed between them. The bonus is one
// effect of one source (a tech, a milestone, a building, an event), switched
// on through the code the game itself runs when the player earns it. The
// difference between the readings is what the bonus delivered; every other
// meter must stay where it was, or the bonus leaks somewhere it never said.
//
// Where it is measured. Twice:
//
//   - clean: an engine in the age the bonus is earned, with buildings and
//     workers and no other bonus. The effect must be exactly what the text
//     says. A miss here is a broken effect: DEAD, WRONG SIZE or WRONG TARGET.
//   - typical: the pacing model's own typical player (config.TypicalIncome):
//     every tech up to the age and every earlier age's wonder, in the age
//     the bonus is earned and the ages after it. A bonus that works clean
//     and falls short here is CAPPED: a cap swallowed it.
//
// The soft cap. A bonus in a production pool (all production, a resource's
// own production) may fall short for one reason only: its pool is past
// +200%, where a point counts a quarter. That is the rule the game states,
// so it is checked, not excused: what the bonus delivers there must be
// exactly what the rule leaves of it (truthSoftCap, written out here), a
// quarter of its value once the pool is past the knee without it, and the
// note beside it must say that number. Anything else is CAPPED.
//
// BONUS_TRUTH_REPORT=<file> go test ./game -run 'TestBonusTruth$' writes the
// full table, with the age each bonus starts to count a quarter in.
//
// What a percentage promises. Bonuses of one kind add together: +30% and
// +20% make +50%, not +56%. So "+30% gold production" promises 30 points of
// the base rate, and that is what is measured: the change in the rate
// divided by what the buildings make before any bonus.
//
// Two more things fail the guard. A limit the game does not own up to: a
// bonus the typical player does not get in full must carry the game's own
// note (game/caps.go), and a bonus that is delivered must not. And
// a promise about a resource that is still locked in the age it is earned
// in, unless its text says it is for later ("once unlocked").
//
// Coverage is automatic: the promises are read from config, so a new tech,
// milestone, building or event is measured without touching this file. An
// effect type or target the guard does not know fails the test: add a meter
// for it here, in the same change that adds it to the engine.
//
// Accepted exceptions live in truthAccepted, each with its reason. Keep it
// short: an entry there is a promise the game does not keep.

// ----- the engines -----

type truthMode int

const (
	truthClean truthMode = iota
	truthTypical
)

func (m truthMode) String() string {
	if m == truthClean {
		return "clean"
	}
	return "typical"
}

// truthCopies is how many copies of every building a lab engine holds, fully
// staffed: config.FlowCopies, the pacing model's moderate investment.
const truthCopies = int(config.FlowCopies)

// truthSoldiers is the garrison the Defense Rating is read with.
const truthSoldiers = 1000

// newTruthEngine builds a lab engine in age. Clean: truthCopies of every
// building up to age, fully staffed, and nothing else. Typical adds what
// config.TypicalIncome assumes a player holds by then: every tech up to age
// and every earlier age's wonder. Milestones, monuments, events and morale
// are left out, as there, so the typical pools are a floor: a player who
// also holds milestone rewards reaches a cap sooner, never later.
func newTruthEngine(age string, mode truthMode) *GameEngine {
	ge := NewGameEngine()
	ge.SeedRNG(1)
	order := ageOrders()
	at := order[age]
	for _, a := range ageKeys() {
		if order[a] <= at {
			ge.applyAgeUnlocks(a)
		}
	}
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	ge.Workers.SetAge(age)
	ge.Prestige.NoteAgeEntered(age)
	pool := ge.Workers.domains["worker"]
	for _, key := range ge.Buildings.order {
		def := ge.Buildings.defs[key]
		built, ok := order[def.RequiredAge]
		if !ok || built > at {
			continue
		}
		switch def.Category {
		case "wonder":
			if mode == truthTypical && built < at {
				ge.Buildings.counts[key] = 1
			}
		case "monument":
		default:
			ge.Buildings.counts[key] = truthCopies
			if def.WorkerCapacity > 0 {
				n := truthCopies * def.WorkerCapacity
				pool.assignments[key] = n
				pool.count += n
			}
		}
	}
	if mode == truthTypical {
		for key, def := range ge.Research.defs {
			if order[def.Age] <= at {
				ge.Research.researched[key] = true
			}
		}
		ge.Research.rebuildBonuses()
	}
	ge.recalculateRates()
	ge.recalculateTickSpeed()
	// Half of every store: a grant has room, a loss has something to take.
	for _, key := range ge.Resources.order {
		r := ge.Resources.resources[key]
		r.Amount = r.Storage / 2
	}
	return ge
}

// truthLab hands out the lab engines, one per age and mode, built on first
// use. Probes share them: each switches its bonus on and off again.
type truthLab struct {
	engines map[string]*GameEngine
	// everyAge makes a bonus in a pool be measured in every age from the
	// one it is earned in, not only the first three and the last: slower,
	// and what the report needs to name the age a cap starts to bite in.
	everyAge bool
}

func newTruthLab() *truthLab {
	return &truthLab{engines: map[string]*GameEngine{}}
}

func (l *truthLab) engine(age string, mode truthMode) *GameEngine {
	k := age + "/" + mode.String()
	if ge, ok := l.engines[k]; ok {
		return ge
	}
	ge := newTruthEngine(age, mode)
	l.engines[k] = ge
	return ge
}

// firstMade is the first age from age on in which buildings make res ("" if
// none do): a bonus to a resource has nothing to raise before it. With
// income set, any income counts (a tech's or a wonder's flat output too), as
// the typical player has it.
func (l *truthLab) firstMade(res, age string, income bool) string {
	keys := ageKeys()
	for _, a := range keys[ageOrders()[age]:] {
		if income {
			ge := l.engine(a, truthTypical)
			ge.recalculateRates()
			if ge.Resources.resources[res].Rate > 0 {
				return a
			}
			continue
		}
		ge := l.engine(a, truthClean)
		ge.recalculateRates()
		if ge.Resources.resources[res].Breakdown.BuildingRate > 0 {
			return a
		}
	}
	return ""
}

// ----- the meters -----

// truthReading is everything the cheap meters read off an engine, by name:
// "rate:food", "made:food" (what buildings make before any bonus),
// "storage:food", "stock:food", "housing", "speed", "research", "cost",
// "defense", "morale", "workers".
type truthReading map[string]float64

const (
	truthRefTech     = "zz_truth_ref_tech"
	truthRefBuilding = "zz_truth_ref_building"
	truthRefTicks    = 1_000_000
	truthRefCost     = 1e12
	truthMoraleStart = 10.0
)

// readTruth reads every cheap meter. It leaves the engine as it found it.
func readTruth(t *testing.T, ge *GameEngine) truthReading {
	t.Helper()
	ge.recalculateRates()
	ge.recalculateTickSpeed()
	out := truthReading{}
	for _, key := range ge.Resources.order {
		r := ge.Resources.resources[key]
		out["rate:"+key] = r.Rate
		out["made:"+key] = r.Breakdown.BuildingRate
		out["storage:"+key] = r.Storage
		out["stock:"+key] = r.Amount
	}
	out["housing"] = float64(ge.popCapLocked())
	out["workers"] = float64(ge.Workers.TotalPop())
	// Game speed: how many base ticks pass per tick interval.
	out["speed"] = float64(BaseTickInterval) / float64(ge.tickIntervalLocked())
	out["research"] = truthResearchTime(t, ge)
	out["cost"] = truthBuildCost(ge)
	// Construction time: the share of its listed time a building takes.
	out["buildtime"] = float64(ge.buildTicksLocked(config.BuildingDef{BuildTicks: truthRefTicks})) / truthRefTicks
	for key, v := range truthMechanics(ge) {
		out["mech:"+key] = v
	}
	out["defense"] = truthDefense(ge)
	out["morale"] = truthMoraleLift(t, ge)
	return out
}

// truthMechanics reads every mechanic number a tech can move
// (config.Mechanics), each where the game uses it: what the market pays on a
// pair it trades at parity (as the fee that leaves), a route's time per
// run, a scouting expedition launched off a fixed roll, how long a set of
// deals lasts, what a gift costs, the share of its price a festival costs,
// the wait between festivals.
// Hand gathering and raid losses are read as the engine's own term: a real
// gather and a real raid spend the engine, and TestTechGatherAndRaids runs
// both. A number that cannot be read in this engine is left out. It leaves
// the engine as it found it.
func truthMechanics(ge *GameEngine) map[string]float64 {
	out := map[string]float64{
		config.MechanicGatherAmount:          ge.gatherBonus(),
		config.MechanicRaidLoss:              ge.raidLossFactor(),
		config.MechanicRouteTicks:            float64(ge.Trade.RunTicks(truthRefTicks)) / truthRefTicks,
		config.MechanicDealRefreshTicks:      float64(ge.Diplomacy.dealRefreshFor(ge.age)),
		config.MechanicGiftCost:              ge.Diplomacy.GiftPrice(),
		config.MechanicFestivalCooldownTicks: float64(ge.festivalCooldown()),
	}
	// A festival's price is a share of the culture store: read it against
	// what it would be with no tech, so only the techs' cut moves it.
	if full := math.Max(festivalMinCost, ge.Resources.GetStorage("culture")*festivalCostFraction); full > 0 {
		out[config.MechanicFestivalCost] = ge.festivalCost() / full
	}
	if priced := ge.rules.PricedResources(ge.age); len(priced) >= 2 {
		from, to := priced[0], priced[1]
		base, ok := ge.rules.MarketRate(from, to, ge.age)
		if paid, _ := ge.Trade.marketRate(from, to, ge.age); ok && base > 0 {
			out[config.MechanicMarketFee] = 1 - (1-config.ExchangeFee)*paid/base
		}
	}
	if defs := ge.Military.GetAvailableExpeditionsByCategory(ExpeditionScouting, ge.age, ageOrders()); len(defs) > 0 {
		held := ge.Military.activeByCat[ExpeditionScouting]
		delete(ge.Military.activeByCat, ExpeditionScouting)
		if err := ge.Military.LaunchExpedition(rand.New(rand.NewSource(1)), defs[0].Key, ge.age, ageOrders()); err == nil {
			out[config.MechanicExpeditionTicks] = float64(ge.Military.activeByCat[ExpeditionScouting].TicksLeft)
		}
		delete(ge.Military.activeByCat, ExpeditionScouting)
		if held != nil {
			ge.Military.activeByCat[ExpeditionScouting] = held
		}
	}
	return out
}

// truthLayerScale is what one point of the tech layer on res is worth in its
// final rate: what the resource's buildings and crews make after every
// pooled bonus (the layer's base), times what multiplies the rate after the
// layer (an ally's bonus, the Cosmic Legacy, Era Mastery). 0 when nothing
// makes res here. Every part is read off the breakdown, and none of them
// moves with the layer, so it reads the same whatever is researched.
func truthLayerScale(ge *GameEngine, res string) float64 {
	ge.recalculateRates()
	b := ge.Resources.resources[res].Breakdown
	base := b.BuildingRate + b.BonusRate + b.WorkerRate
	if base <= 0 {
		return 0
	}
	before := base + b.ResearchRate + b.EventRate
	ally := 1 + b.TradeRate/before
	legacy := 1 + b.LegacyRate/(before+b.TradeRate)
	return base * ally * legacy * ge.speedK()
}

// truthCut reads a tech's cut of a price or a time off meter: the share
// that comes off, whatever was taken off before (the cuts multiply).
func truthCut(meter string, unit string) truthKind {
	return truthKind{Unit: unit, Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			if before[meter] <= 0 {
				return truthMeasured{Skip: "no " + meter + " to read"}
			}
			// The reference price and times round to whole units.
			return truthMeasured{Delivered: after[meter]/before[meter] - 1, Noise: 2 / (truthRefTicks * before[meter]), Allowed: meterIs(meter)}
		}}
}

// truthResearchTime is the share of its listed time a tech takes to research
// now (1 with no bonus), read by starting a reference tech through the real
// start and taking it back.
func truthResearchTime(t *testing.T, ge *GameEngine) float64 {
	t.Helper()
	rm := ge.Research
	cur, left, total := rm.currentTech, rm.ticksLeft, rm.totalTicks
	know := ge.Resources.resources["knowledge"].Amount
	rm.currentTech = ""
	rm.defs[truthRefTech] = config.TechDef{Key: truthRefTech, Name: "Reference", Age: "primitive_age", ResearchTicks: truthRefTicks}
	if err := ge.startResearchLocked(truthRefTech, true); err != nil {
		t.Fatalf("research meter: %v", err)
	}
	ticks := rm.totalTicks
	delete(rm.defs, truthRefTech)
	rm.currentTech, rm.ticksLeft, rm.totalTicks = cur, left, total
	ge.Resources.resources["knowledge"].Amount = know
	return float64(ticks) / truthRefTicks
}

// truthBuildCost is the share of its listed price a building costs now (1
// with no discount), read off a reference building's price.
func truthBuildCost(ge *GameEngine) float64 {
	ge.Buildings.defs[truthRefBuilding] = config.BuildingDef{Key: truthRefBuilding, BaseCost: map[string]float64{"wood": truthRefCost}, CostScale: 1}
	c := ge.Buildings.GetCost(truthRefBuilding)["wood"]
	delete(ge.Buildings.defs, truthRefBuilding)
	return c / truthRefCost
}

// truthDefense is the Defense Rating of a garrison of truthSoldiers, per
// point of its unboosted rating (1 with no bonus).
func truthDefense(ge *GameEngine) float64 {
	r := ge.Resources.resources["soldiers"]
	have := r.Amount
	r.Amount = truthSoldiers
	rating := ge.defenseRating()
	r.Amount = have
	return rating / (2 * truthSoldiers)
}

// truthMoraleLift is how far one tick moves morale. It is read high on the
// dial with the cap lifted out of the way (a thousand copies of a wonder
// raise it for the length of the tick), where neither the floor, the pivot
// nor the cap bends the reading.
func truthMoraleLift(t *testing.T, ge *GameEngine) float64 {
	t.Helper()
	wonder := ""
	for _, key := range ge.Buildings.order {
		def := ge.Buildings.defs[key]
		lifts := false
		for _, e := range def.Effects {
			lifts = lifts || e.Type == "morale"
		}
		if def.Category == "wonder" && !lifts {
			wonder = key
			break
		}
	}
	if wonder == "" {
		t.Fatal("morale meter: no wonder to raise the cap with")
	}
	m, warned, had := ge.morale, ge.lowMoraleWarned, ge.Buildings.counts[wonder]
	ge.Buildings.counts[wonder] = had + 1000
	ge.morale = truthMoraleStart
	ge.updateMoraleTick()
	after := ge.morale
	ge.Buildings.counts[wonder] = had
	if had == 0 {
		delete(ge.Buildings.counts, wonder)
	}
	ge.morale, ge.lowMoraleWarned = m, warned
	if after <= moraleNeutral+moraleDrift || after >= truthMoraleStart*2 {
		t.Fatalf("morale meter left its straight stretch: %v after one tick from %v", after, truthMoraleStart)
	}
	return after - truthMoraleStart
}

// truthCrewOutput is what workers add to their buildings, per resource and
// before any bonus: what the buildings make staffed less what they make
// empty. It is what a worker output bonus raises.
func truthCrewOutput(ge *GameEngine) map[string]float64 {
	pool := ge.Workers.domains["worker"]
	crew := pool.assignments
	read := func() map[string]float64 {
		ge.recalculateRates()
		out := map[string]float64{}
		for _, key := range ge.Resources.order {
			out[key] = ge.Resources.resources[key].Breakdown.BuildingRate
		}
		return out
	}
	staffed := read()
	pool.assignments = map[string]int{}
	empty := read()
	pool.assignments = crew
	ge.recalculateRates()
	out := map[string]float64{}
	for key, v := range staffed {
		if d := v - empty[key]; d > 0 {
			out[key] = d
		}
	}
	return out
}

// truthSureSource is a random source whose first draw is as high as a draw
// can be, so the success roll it feeds always passes; every later draw
// comes from an ordinary seeded stream.
type truthSureSource struct {
	drawn bool
	rest  rand.Source
}

func (s *truthSureSource) Int63() int64 {
	if !s.drawn {
		s.drawn = true
		return 1<<63 - 1025 // Float64 reads 1 - 2^-53
	}
	return s.rest.Int63()
}

func (s *truthSureSource) Seed(seed int64) { s.rest.Seed(seed) }

// truthLoot is what a successful expedition brings back per unit of its
// listed reward (1 with no bonus): an expedition of the engine's age is
// resolved through the real tick with a sure success roll. It spends the
// engine: call it on one made for the reading.
func truthLoot(t *testing.T, ge *GameEngine) float64 {
	t.Helper()
	defs := ge.Military.GetAvailableExpeditionsByCategory(ExpeditionScouting, ge.age, ageOrders())
	if len(defs) == 0 {
		t.Fatalf("loot meter: no expedition in %s", ge.age)
	}
	def := defs[0]
	ge.rng = rand.New(&truthSureSource{rest: rand.NewSource(1)}) // the success roll is the tick's first draw
	ge.Military.activeByCat[ExpeditionScouting] = &ActiveExpedition{Key: def.Key, Name: def.Name, TicksLeft: 1}
	ge.Military.totalLoot = map[string]float64{}
	ge.processExpeditions()
	res := sortedKeys(def.Rewards)[0]
	return ge.Military.totalLoot[res] / def.Rewards[res]
}

// truthRouteIncome is what a trade route brings in per unit of its listed
// import (1 with no bonus): a route of the engine's age runs one cycle
// through the real tick. It spends the engine.
func truthRouteIncome(t *testing.T, ge *GameEngine) (float64, string) {
	t.Helper()
	order := ageOrders()
	for _, def := range ge.Trade.routeList {
		if order[def.MinAge] > order[ge.age] || ge.Buildings.GetCount(def.RequiredBld) < def.MinCount {
			continue
		}
		for res, amount := range def.Export {
			r := ge.Resources.resources[res]
			r.Amount = math.Max(r.Amount, amount)
		}
		ge.Trade.activeRoutes = map[string]*ActiveRoute{def.Key: {Key: def.Key, TicksLeft: 1}}
		ge.Trade.totalImported = map[string]float64{}
		ge.processTrade()
		res := sortedKeys(def.Import)[0]
		return ge.Trade.totalImported[res] / def.Import[res], ""
	}
	return 0, "no trade route runs in " + ge.age
}

// truthOpinion is the opinion one diplomacy tick adds across two neutral
// civilizations, in all. It spends the engine.
func truthOpinion(ge *GameEngine) float64 {
	keys := sortedKeys(ge.Diplomacy.factionDefs)[:2]
	ge.Diplomacy.factions = map[string]*FactionState{}
	for _, k := range keys {
		ge.Diplomacy.factions[k] = &FactionState{Discovered: true, Status: "neutral"}
	}
	ge.tick = 1 // off every drift and decay beat
	ge.processDiplomacy()
	// The tick may meet more civilizations (the age fallback): they share
	// the embassy's opinion, so all of them are counted.
	total := 0.0
	for _, k := range sortedKeys(ge.Diplomacy.factions) {
		fs := ge.Diplomacy.factions[k]
		total += float64(fs.Opinion) + fs.OpinionAccum
	}
	return total
}

// ----- a promise -----

// truthPromise is one effect of one source: what the game says it does and
// how to switch it on.
type truthPromise struct {
	Source string // "tech", "milestone", "building", "wonder", ...
	Key    string // the source's config key
	Name   string // the source's display name
	Age    string // the age a player earns it in
	Eff    config.Effect
	// Text is the promise as the game words it.
	Text string
	// Kind picks the meter (truthKinds).
	Kind string
	// Count is how many copies of the source deliver the effect together
	// (truthCopies for a building; 1 otherwise).
	Count float64
	// Later marks a promise that is for when its resource exists, and says
	// so: a building whose text reads "once unlocked", a legacy carried into
	// every later run. Without it, a promise about a resource that is still
	// locked in the age it is earned in is a promise the game cannot keep.
	Later bool
	// Also lists the meters the source moves besides this promise's own,
	// because it is applied whole and its other promises ride along (an
	// epoch event with two rates). A name ending in ":" is a prefix.
	Also []string
	// Raid marks a loss a raid takes, which the techs' cut of raid losses
	// shrinks before it lands.
	Raid bool
	// wire prepares the states a probe needs in ge.
	wire func(ge *GameEngine) truthSwitch
}

// also reports whether meter is one the promise's source moves on the side.
func (p truthPromise) also(meter string) bool {
	for _, a := range p.Also {
		if meter == a || (strings.HasSuffix(a, ":") && strings.HasPrefix(meter, a)) {
			return true
		}
	}
	return false
}

// truthSwitch moves one engine between the two readings of a probe.
type truthSwitch struct {
	off     func() // the effect off, everything else as it stands
	on      func() // the effect on
	restore func() // the engine as the probe found it
}

func (p truthPromise) ID() string { return p.Source + " " + p.Key + ": " + p.Text }

// truthOutcome is one measurement of one promise.
type truthOutcome struct {
	Age       string
	Mode      truthMode
	Promised  float64
	Delivered float64
	// Noise is how far float rounding alone can move Delivered.
	Noise float64
	// Said is the note the game shows beside the bonus while it is on (""
	// when it shows none): what the panels and the log tell the player
	// about it (CapNote).
	Said string
	// Rule is what the soft cap leaves of the promise where it was read,
	// for a promise that moves a production pool: all of it while the pool
	// stays under +200%, a quarter of each point past it. Promised for any
	// other promise. Knee says the pool was past +200% with the bonus on,
	// and Past that it already was with the bonus off, so that every point
	// of the bonus counts a quarter.
	Rule       float64
	Knee, Past bool
	// Leaks are the meters that moved and were not part of the promise.
	Leaks []string
	// Skip is why nothing could be measured here ("" when measured).
	Skip string
}

// full reports whether the promise was delivered in full.
func (o truthOutcome) full() bool { return o.Skip == "" && o.is(o.Promised) }

// is reports whether Delivered is want, within the reading's noise.
func (o truthOutcome) is(want float64) bool {
	return math.Abs(o.Delivered-want) <= 3e-6*math.Max(1, math.Abs(want))+o.Noise
}

// truthSoftCap is what a production pool that has earned bonuses in all
// applies: all of it up to +200%, a quarter of every point past it. The
// rule is written out here on purpose, numbers and all, and not read from
// config or the engine: this is the promise the game makes the player.
func truthSoftCap(earned float64) float64 {
	if earned > 2.0 {
		return 2.0 + (earned-2.0)*0.25
	}
	return earned
}

// truthKneePool is the production pool a promise adds to ("" for a promise
// in no pool with a knee).
func truthKneePool(p truthPromise) string {
	target, ok := effectPool(p.Eff)
	if !ok || target == "gather_rate" {
		return ""
	}
	if target == "production_all" || strings.HasSuffix(target, "_rate") {
		return target
	}
	return ""
}

// rule fills in what the soft cap leaves of the promise, from what its pool
// had earned with the bonus off and on. A promise that moves no pool (a
// tech, an ally's bonus, the Cosmic Legacy) is owed in full.
func (o *truthOutcome) rule(off, on float64) {
	o.Rule = o.Promised
	if off == on {
		return
	}
	o.Rule = truthSoftCap(on) - truthSoftCap(off)
	o.Knee = math.Max(off, on) > 2.0+1e-9
	o.Past = math.Min(off, on) >= 2.0-1e-9
}

// truthClose reports whether got is want within a few parts in a million
// (the reference tech and price round to whole ticks and units).
func truthClose(got, want float64) bool {
	return math.Abs(got-want) <= 3e-6*math.Max(1, math.Abs(want))
}

// truthMoved reports whether a meter changed at all.
func truthMoved(before, after float64) bool {
	return math.Abs(after-before) > 1e-9*math.Max(1, math.Max(math.Abs(before), math.Abs(after)))
}

// ----- the kinds: what each effect promises and which meter reads it -----

// truthMeasured is what a meter read for one promise.
type truthMeasured struct {
	// Delivered is what the bonus did, in the promise's own unit.
	Delivered float64
	// Noise is how far float rounding alone can move Delivered: a +0.2/tick
	// read off a rate in the billions is only good to a few thousandths.
	Noise float64
	// Allowed says which meters the promise may move.
	Allowed func(meter string) bool
	// Skip is why nothing could be measured in this engine ("" when read).
	Skip string
}

// truthKind reads what a promise delivered.
type truthKind struct {
	// Unit says what Delivered and Promised count.
	Unit string
	// Pool marks a kind whose bonuses share a pool that a cap can hold: its
	// promises are measured in every age from the one they are earned in.
	Pool bool
	// measure reads the cheap meters' two readings. Set for the kinds they
	// cover.
	measure func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured
	// spend reads a meter that uses the engine up (a resolved expedition, a
	// trade run, a diplomacy tick): it is read on two fresh engines, the
	// effect off in one and on in the other. Set instead of measure.
	spend func(t *testing.T, p truthPromise, ge *GameEngine) (reading float64, skip string)
}

func meterIs(names ...string) func(string) bool {
	return func(m string) bool {
		for _, n := range names {
			if m == n {
				return true
			}
		}
		return false
	}
}

func meterHasPrefix(prefixes ...string) func(string) bool {
	return func(m string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(m, p) {
				return true
			}
		}
		return false
	}
}

func anyMeter(string) bool { return true }

// truthDiff is a meter's change between two readings and the float noise
// on it: a few units in the last place of the larger reading.
func truthDiff(before, after truthReading, meter string) (d, noise float64) {
	a, b := before[meter], after[meter]
	big := math.Max(math.Abs(a), math.Abs(b))
	return b - a, 8 * (math.Nextafter(big, math.Inf(1)) - big)
}

// truthRefResources are the resources whose rate reads the all-production
// multiplier directly: buildings make them, and no bonus of their own is in
// the way.
func truthRefResources(ge *GameEngine, r truthReading) []string {
	res := ge.buildResolver()
	var out []string
	for _, key := range ge.Resources.order {
		if r["made:"+key] <= 0 || len(res.Breakdown(key+"_rate")) > 0 {
			continue
		}
		out = append(out, key)
	}
	return out
}

// truthAllFactor is the all-production multiplier in force, read off a
// reference resource: what it yields per unit its buildings make.
func truthAllFactor(ge *GameEngine, r truthReading) (float64, bool) {
	refs := truthRefResources(ge, r)
	if len(refs) == 0 {
		return 0, false
	}
	ge.recalculateRates()
	b := ge.Resources.resources[refs[0]].Breakdown
	return (b.BuildingRate + b.BonusRate) / b.BuildingRate, true
}

// truthAcross reads one promise off several resources that must agree:
// each resource's rate change over its own base. layered says the promise
// sits in a pool, before the tech layer: the layer multiplies what it adds,
// as Era Mastery's k does, and the reading is taken per point of both.
func truthAcross(ge *GameEngine, before, after truthReading, base map[string]float64, layered bool) truthMeasured {
	m := truthMeasured{Allowed: meterHasPrefix("rate:")}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, key := range sortedKeys(base) {
		d, noise := truthDiff(before, after, "rate:"+key)
		scale := base[key] * ge.speedK()
		if layered {
			scale *= ge.Research.OutputFactor(key)
		}
		lo, hi = math.Min(lo, d/scale), math.Max(hi, d/scale)
		m.Noise = math.Max(m.Noise, noise/scale)
	}
	m.Delivered = lo
	if hi-lo > 3e-6*math.Max(1, math.Abs(hi))+2*m.Noise {
		m.Skip = fmt.Sprintf("resources disagree: %v to %v", lo, hi)
	}
	return m
}

var truthKinds = map[string]truthKind{
	// A tech's "+X% <resource> production": the tech layer. X of what the
	// resource's buildings and crews make after every pooled bonus and cap,
	// in every age: nothing caps the layer.
	"tech_output": {
		Unit: "of what the resource makes", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			res := p.Eff.Target
			scale := truthLayerScale(ge, res)
			if scale <= 0 {
				return truthMeasured{Skip: "no " + res + " is made here"}
			}
			d, noise := truthDiff(before, after, "rate:"+res)
			return truthMeasured{Delivered: d / scale, Noise: noise / scale, Allowed: meterIs("rate:" + res)}
		},
	},
	// A tech's "+X% all production": the same, on every resource.
	"tech_all_output": {
		Unit: "of what every resource makes", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			m := truthMeasured{Allowed: meterHasPrefix("rate:")}
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, key := range ge.Resources.order {
				scale := truthLayerScale(ge, key)
				if scale <= 0 {
					continue
				}
				d, noise := truthDiff(before, after, "rate:"+key)
				lo, hi = math.Min(lo, d/scale), math.Max(hi, d/scale)
				m.Noise = math.Max(m.Noise, noise/scale)
			}
			if math.IsInf(lo, 1) {
				return truthMeasured{Skip: "nothing is made here"}
			}
			m.Delivered = lo
			if hi-lo > 3e-6*math.Max(1, math.Abs(hi))+2*m.Noise {
				m.Skip = fmt.Sprintf("resources disagree: %v to %v", lo, hi)
			}
			return m
		},
	},
	// A tech's "+X% storage": X of every store, whatever else is researched
	// (the techs' storage bonuses add up).
	"tech_storage": {
		Unit: "of every store", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			ge.recalculateRates()
			held := 1 + ge.Research.Bonus(config.EffectStorage, "")
			m := truthMeasured{Allowed: meterHasPrefix("storage:", "stock:")}
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, key := range ge.Resources.order {
				base := ge.Resources.resources[key].Storage / held
				if base <= 0 {
					continue
				}
				d, noise := truthDiff(before, after, "storage:"+key)
				lo, hi = math.Min(lo, d/base), math.Max(hi, d/base)
				m.Noise = math.Max(m.Noise, noise/base)
			}
			m.Delivered = lo
			if hi-lo > 3e-6*math.Max(1, math.Abs(hi))+2*m.Noise {
				m.Skip = fmt.Sprintf("resources disagree: %v to %v", lo, hi)
			}
			return m
		},
	},
	// A tech's "+X% housing": X of housing, rounded up to a whole person
	// (game.TechHousing), so the reading is good to a person either way.
	"tech_housing": {
		Unit: "of housing", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			base := float64(ge.Buildings.GetPopCapacity() + int(ge.permanentBonuses["population"]+ge.Prestige.GetBonuses()["population"]))
			if base <= 0 {
				return truthMeasured{Skip: "no housing here"}
			}
			return truthMeasured{Delivered: (after["housing"] - before["housing"]) / base, Noise: 1 / base, Allowed: meterIs("housing")}
		},
	},
	// A tech's cut of building costs, of construction time, of research
	// time: X of it comes off.
	"tech_build_cost":    truthCut("cost", "of what buildings cost"),
	"tech_build_time":    truthCut("buildtime", "of the construction time"),
	"tech_research_time": truthCut("research", "of the research time"),
	// A tech's step on one number of a mechanic (config.Mechanics): the
	// number is multiplied, or added to, as the mechanic says.
	"tech_mechanic": {
		Unit: "of the mechanic's number", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			meter := "mech:" + p.Eff.Target
			def := config.MechanicByKey()[p.Eff.Target]
			a, ok := before[meter]
			if !ok {
				return truthMeasured{Skip: def.Name + " cannot be read here"}
			}
			if !def.Multiplies {
				return truthMeasured{Delivered: after[meter] - a, Allowed: meterIs(meter)}
			}
			if a <= 0 {
				return truthMeasured{Skip: def.Name + " reads zero here"}
			}
			// A number of ticks is whole: good to a tick either way.
			noise := 0.0
			if strings.HasSuffix(p.Eff.Target, "_ticks") && a > 1 {
				noise = 2 / a
			}
			return truthMeasured{Delivered: after[meter]/a - 1, Noise: noise, Allowed: meterIs(meter)}
		},
	},
	// "+X% all production": X points on every resource buildings make.
	"all_production": {
		Unit: "points of base output", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			base := map[string]float64{}
			for _, key := range truthRefResources(ge, before) {
				base[key] = before["made:"+key]
			}
			if len(base) == 0 {
				return truthMeasured{Skip: "no building output to read"}
			}
			return truthAcross(ge, before, after, base, true)
		},
	},
	// "+X% all production, after the caps" (the Cosmic Legacy): everything
	// a resource makes, bonuses and all, times 1 + X, in every age. Read as
	// the rise in each rate over what the resource made before it (the
	// food drain is not production and is left out).
	"final_production": {
		Unit: "of what every resource makes", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The probe has put the engine back: the bonus is off.
			ge.recalculateRates()
			k := ge.speedK()
			base := map[string]float64{}
			for _, key := range ge.Resources.order {
				r := ge.Resources.resources[key]
				if made := r.Rate/k - r.Breakdown.FoodDrain - r.Breakdown.LegacyRate; made > 0 {
					base[key] = made
				}
			}
			if len(base) == 0 {
				return truthMeasured{Skip: "nothing is made here"}
			}
			return truthAcross(ge, before, after, base, false)
		},
	},
	// "+X% <resource> production": X points on that resource alone.
	"resource_production": {
		Unit: "points of base output", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			res := truthRateTarget(p.Eff)
			made := before["made:"+res]
			if made <= 0 {
				return truthMeasured{Skip: "no " + res + " is made here"}
			}
			all, ok := truthAllFactor(ge, before)
			if !ok {
				return truthMeasured{Skip: "no building output to read"}
			}
			d, noise := truthDiff(before, after, "rate:"+res)
			// The tech layer multiplies what the pool adds, as k does.
			scale := made * all * ge.speedK() * ge.Research.OutputFactor(res)
			return truthMeasured{Delivered: d / scale, Noise: noise / scale, Allowed: meterIs("rate:" + res)}
		},
	},
	// "+X% worker output": X points on what workers add to their buildings.
	"worker_output": {
		Unit: "points of base worker output", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			crew := truthCrewOutput(ge)
			if len(crew) == 0 {
				return truthMeasured{Skip: "no staffed building to read"}
			}
			return truthAcross(ge, before, after, crew, true)
		},
	},
	// "+V <resource>/tick" from a tech or an event: V more per tick.
	"flat_rate": {
		Unit: "per tick",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			d, noise := truthDiff(before, after, "rate:"+p.Eff.Target)
			return truthMeasured{Delivered: d / ge.speedK(), Noise: noise, Allowed: meterIs("rate:" + p.Eff.Target)}
		},
	},
	// "+V <resource>/tick (N workers)" on a building: each fully staffed
	// copy makes V before bonuses.
	"building_output": {
		Unit: "per tick per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			res := p.Eff.Target
			d, noise := truthDiff(before, after, "made:"+res)
			return truthMeasured{Delivered: d / p.Count, Noise: noise / p.Count, Allowed: meterIs("rate:"+res, "made:"+res)}
		},
	},
	// "+V storage for every resource" or for one.
	"storage": {
		Unit: "storage per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The techs' storage bonus multiplies every store, as k does.
			scale := p.Count * ge.speedK() * (1 + ge.Research.Bonus(config.EffectStorage, ""))
			if p.Eff.Target != "all" {
				d, noise := truthDiff(before, after, "storage:"+p.Eff.Target)
				return truthMeasured{Delivered: d / scale, Noise: noise / scale, Allowed: meterIs("storage:" + p.Eff.Target)}
			}
			m := truthMeasured{Allowed: meterHasPrefix("storage:")}
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, key := range ge.Resources.order {
				d, noise := truthDiff(before, after, "storage:"+key)
				lo, hi = math.Min(lo, d/scale), math.Max(hi, d/scale)
				m.Noise = math.Max(m.Noise, noise/scale)
			}
			m.Delivered = lo
			if hi-lo > 3e-6*math.Max(1, math.Abs(hi))+2*m.Noise {
				m.Skip = fmt.Sprintf("resources disagree: %v to %v", lo, hi)
			}
			return m
		},
	},
	// "+V housing".
	"housing": {
		Unit: "housing per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The techs' housing bonus multiplies housing and rounds it up
			// to a whole person: with one, the reading is good to a person.
			h := ge.Research.Bonus(config.EffectHousing, "")
			if h == 0 {
				return truthMeasured{Delivered: (after["housing"] - before["housing"]) / p.Count, Allowed: meterIs("housing")}
			}
			scale := p.Count * (1 + h)
			return truthMeasured{Delivered: (after["housing"] - before["housing"]) / scale, Noise: 1 / scale, Allowed: meterIs("housing")}
		},
	},
	// "-X% building costs": X points off every price.
	"build_cost": {
		Unit: "points of the listed price", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The techs' cut multiplies what the pool leaves of the price.
			return truthMeasured{Delivered: (after["cost"] - before["cost"]) / ge.Research.Bonus(config.EffectBuildCost, ""), Allowed: meterIs("cost")}
		},
	},
	// "+X% game speed".
	"game_speed": {
		Unit: "points of game speed", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			return truthMeasured{Delivered: after["speed"] - before["speed"], Allowed: meterIs("speed")}
		},
	},
	// "+X% research speed": research takes X points less of its listed time
	// (site/docs/technologies.md: ticks = base x (1 - research speed)).
	"research_speed": {
		Unit: "points of the listed research time", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The techs' cut multiplies what the pool leaves of the time.
			return truthMeasured{Delivered: (before["research"] - after["research"]) / ge.Research.TimeFactor(), Noise: 2 / truthRefTicks, Allowed: meterIs("research")}
		},
	},
	// "research time x(1 - X)" (Ancient Knowledge): X of the research time
	// as it stood comes off, whatever was taken off it before.
	"research_time": {
		Unit: "of the research time left",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			// The reference tech rounds to whole ticks, before and after.
			noise := 2 / (truthRefTicks * before["research"])
			return truthMeasured{Delivered: 1 - after["research"]/before["research"], Noise: noise, Allowed: meterIs("research")}
		},
	},
	// "+X% military power": X points on the Defense Rating.
	"military_power": {
		Unit: "points of Defense Rating", Pool: true,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			return truthMeasured{Delivered: after["defense"] - before["defense"], Allowed: meterIs("defense")}
		},
	},
	// A worship or culture building's morale lift per tick.
	"morale": {
		Unit: "morale per tick per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			return truthMeasured{Delivered: (after["morale"] - before["morale"]) / p.Count, Allowed: meterIs("morale")}
		},
	},
	// "+V <resource>" at once.
	"instant": {
		Unit: "added to the store",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			d, noise := truthDiff(before, after, "stock:"+p.Eff.Target)
			return truthMeasured{Delivered: d, Noise: noise, Allowed: meterIs("stock:" + p.Eff.Target)}
		},
	},
	// "Up to V <resource>" lost at once (no garrison in the lab).
	"steal": {
		Unit: "taken from the store",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			key := "stock:" + p.Eff.Target
			if before[key] < p.Eff.Value {
				return truthMeasured{Skip: "the store holds less than the loss"}
			}
			d, noise := truthDiff(before, after, key)
			// A raid's take is cut by the techs' share before it lands
			// (config.MechanicRaidLoss); any other loss is taken whole.
			cut := 1.0
			if p.Raid {
				cut = ge.raidLossFactor()
			}
			return truthMeasured{Delivered: -d / cut, Noise: noise / cut, Allowed: meterIs(key)}
		},
	},
	// "X% of your workers" lost at once: that share, in whole workers.
	"worker_loss": {
		Unit: "share of workers",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			had := before["workers"]
			if had < 20 {
				return truthMeasured{Skip: "too few workers to read a share off"}
			}
			lost := had - after["workers"]
			// Whole workers leave: the share is right within one of them.
			return truthMeasured{Delivered: lost / had, Noise: 1 / had, Allowed: anyMeter}
		},
	},
	// "+N workers" lent, or "N workers" lost.
	"workers": {
		Unit: "workers",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			return truthMeasured{Delivered: after["workers"] - before["workers"], Allowed: anyMeter}
		},
	},
	// "X of your <resource>" lost at once, as a share of the store.
	"drain": {
		Unit: "share of the store",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			key := "stock:" + p.Eff.Target
			if before[key] <= 0 {
				return truthMeasured{Skip: "nothing in the store"}
			}
			return truthMeasured{Delivered: (before[key] - after[key]) / before[key], Allowed: meterIs(key)}
		},
	},
	// An ally's "+X% <specialty>": the whole rate times 1 + X, outside the
	// production cap (site/docs/factions.md, Allied Bonuses).
	"ally_bonus": {
		Unit: "share of the whole rate",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			res := p.Eff.Target
			gross := before["rate:"+res]
			if res == "food" {
				gross += ge.Workers.FoodDrain() * ge.speedK()
			}
			if gross <= 0 {
				return truthMeasured{Skip: "no " + res + " comes in here"}
			}
			d, noise := truthDiff(before, after, "rate:"+res)
			return truthMeasured{Delivered: d / gross, Noise: noise / gross, Allowed: meterIs("rate:" + res)}
		},
	},
	// "+X% expedition rewards": X points on what a success brings back.
	"expedition_reward": {
		Unit: "points of the listed loot", Pool: true,
		spend: func(t *testing.T, p truthPromise, ge *GameEngine) (float64, string) {
			return truthLoot(t, ge), ""
		},
	},
	// "+X% trade route income": X points on what a route brings in.
	"route_income": {
		Unit: "points of the listed import per copy",
		spend: func(t *testing.T, p truthPromise, ge *GameEngine) (float64, string) {
			v, skip := truthRouteIncome(t, ge)
			return v / p.Count, skip
		},
	},
	// "+V opinion/tick per worker", split across the civilizations that are
	// not hostile: V for every worker slot filled, in all.
	"opinion": {
		Unit: "opinion per tick per worker, in all",
		spend: func(t *testing.T, p truthPromise, ge *GameEngine) (float64, string) {
			def := ge.Buildings.defs[p.Key]
			return truthOpinion(ge) / p.Count / float64(def.WorkerCapacity), ""
		},
	},
}

// truthAlso is kind with the leak check off: for a source that is applied
// whole (an epoch event, a Succumb's flag, a setback with two halves), where
// the other parts move meters of their own. Only the promised one is read.
func truthAlso(kind string, also ...string) truthKind {
	base := truthKinds[kind]
	return truthKind{Unit: base.Unit, Pool: base.Pool,
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) truthMeasured {
			m := base.measure(p, ge, before, after)
			if allowed := m.Allowed; allowed != nil {
				m.Allowed = anyMeter
				if len(also) > 0 {
					m.Allowed = func(meter string) bool { return allowed(meter) || meterIs(also...)(meter) }
				}
			}
			return m
		}}
}

func init() {
	truthKinds["boon_all_production"] = truthAlso("all_production")
	truthKinds["boon_drain"] = truthAlso("drain")
	truthKinds["legacy_production"] = truthAlso("resource_production")
	truthKinds["legacy_research"] = truthAlso("research_time")
	// A festival is paid for in culture: the store moves with the rates.
	truthKinds["festival"] = truthAlso("all_production", "stock:culture")
}

// truthRateTarget is the resource a "<res>_rate" effect raises, whether the
// key sits in the effect's target (a tech, a milestone) or its type (a boon).
func truthRateTarget(e config.Effect) string {
	if res, ok := strings.CutSuffix(e.Target, "_rate"); ok {
		return res
	}
	if res, ok := strings.CutSuffix(e.Type, "_rate"); ok {
		return res
	}
	return e.Target
}

// truthEffectKind names the meter for an effect, by the effect's type and
// target alone. "" is an effect the guard has no meter for: the test fails
// on it, so nothing a player is promised goes unmeasured.
func truthEffectKind(source string, e config.Effect) string {
	isRes := func(key string) bool { _, ok := config.ResourceByKey()[key]; return ok }
	// The tech layer's kinds (TechEffect.Effect writes them "tech_<kind>").
	if kind, ok := strings.CutPrefix(e.Type, "tech_"); ok && source == "tech" {
		switch config.TechEffectKind(kind) {
		case config.EffectOutput:
			if isRes(e.Target) {
				return "tech_output"
			}
		case config.EffectAllOutput, config.EffectStorage, config.EffectHousing,
			config.EffectBuildCost, config.EffectBuildTime, config.EffectResearchTime:
			return e.Type
		case config.EffectMechanic:
			if _, ok := config.MechanicByKey()[e.Target]; ok {
				return "tech_mechanic"
			}
		}
		return ""
	}
	switch e.Type {
	case "bonus", "permanent_bonus":
		switch e.Target {
		case "production_all":
			return "all_production"
		case "gather_rate":
			return "worker_output"
		case "research_speed":
			return "research_speed"
		case "military_power":
			return "military_power"
		case "expedition_reward":
			return "expedition_reward"
		case "build_cost":
			return "build_cost"
		case "tick_speed":
			return "game_speed"
		}
		if res, ok := strings.CutSuffix(e.Target, "_rate"); ok && isRes(res) {
			return "resource_production"
		}
	case "production":
		if isRes(e.Target) {
			if source == "building" || source == "wonder" || source == "monument" {
				return "building_output"
			}
			return "flat_rate"
		}
	case "storage":
		if isRes(e.Target) || e.Target == "all" {
			return "storage"
		}
	case "capacity":
		switch e.Target {
		case "population":
			return "housing"
		}
	case "morale":
		return "morale"
	case "opinion":
		return "opinion"
	case "trade_route_income":
		return "route_income"
	case "instant_resource":
		if isRes(e.Target) {
			return "instant"
		}
	case "steal_resource":
		if isRes(e.Target) {
			return "steal"
		}
	case "worker_loss":
		return "worker_loss"
	case "production_all":
		return "all_production"
	case "tick_speed":
		return "game_speed"
	}
	if res, ok := strings.CutSuffix(e.Type, "_rate"); ok && isRes(res) {
		return "resource_production"
	}
	return ""
}

// ----- measuring -----

// truthSnapshot is the state a probe may disturb beyond its own switch.
type truthSnapshot struct {
	stock     map[string]float64
	permanent map[string]float64
	events    []ActiveEvent
	workers   int
	crew      map[string]int
	loans     []BoonWorkerLoan
	techs     map[string]bool
	counts    map[string]int
	legacy    map[string]bool
}

func takeTruthSnapshot(ge *GameEngine) truthSnapshot {
	pool := ge.Workers.domains["worker"]
	s := truthSnapshot{
		stock:     map[string]float64{},
		permanent: map[string]float64{},
		events:    append([]ActiveEvent(nil), ge.Events.active...),
		workers:   pool.count,
		crew:      map[string]int{},
		loans:     append([]BoonWorkerLoan(nil), ge.Diplomacy.boonLoans...),
		techs:     map[string]bool{},
		counts:    map[string]int{},
		legacy:    map[string]bool{},
	}
	for k, v := range ge.legacyBonuses {
		s.legacy[k] = v
	}
	for _, key := range ge.Resources.order {
		s.stock[key] = ge.Resources.resources[key].Amount
	}
	for k, v := range ge.permanentBonuses {
		s.permanent[k] = v
	}
	for k, v := range pool.assignments {
		s.crew[k] = v
	}
	for k, v := range ge.Research.researched {
		s.techs[k] = v
	}
	for k, v := range ge.Buildings.counts {
		s.counts[k] = v
	}
	return s
}

func (s truthSnapshot) put(ge *GameEngine) {
	pool := ge.Workers.domains["worker"]
	ge.permanentBonuses = map[string]float64{}
	for k, v := range s.permanent {
		ge.permanentBonuses[k] = v
	}
	ge.Events.active = append([]ActiveEvent(nil), s.events...)
	pool.count = s.workers
	pool.assignments = map[string]int{}
	for k, v := range s.crew {
		pool.assignments[k] = v
	}
	ge.Diplomacy.boonLoans = append([]BoonWorkerLoan(nil), s.loans...)
	ge.Research.researched = map[string]bool{}
	for k, v := range s.techs {
		ge.Research.researched[k] = v
	}
	ge.Research.rebuildBonuses()
	ge.Buildings.counts = map[string]int{}
	for k, v := range s.counts {
		ge.Buildings.counts[k] = v
	}
	ge.legacyBonuses = map[string]bool{}
	for k, v := range s.legacy {
		ge.legacyBonuses[k] = v
	}
	ge.recalculateRates() // storage first: stock must fit before it goes back
	for key, v := range s.stock {
		ge.Resources.resources[key].Amount = v
	}
	ge.recalculateRates()
	ge.recalculateTickSpeed()
}

// probe measures p in age and mode.
func (l *truthLab) probe(t *testing.T, p truthPromise, age string, mode truthMode) truthOutcome {
	t.Helper()
	out := truthOutcome{Age: age, Mode: mode, Promised: p.Eff.Value, Rule: p.Eff.Value}
	kind, ok := truthKinds[p.Kind]
	if !ok {
		out.Skip = "no meter for " + p.Kind
		return out
	}
	if kind.spend != nil {
		read := func(on bool) (float64, string) {
			ge := newTruthEngine(age, mode)
			sw := p.wire(ge)
			if sw.off(); on {
				sw.on()
			}
			ge.recalculateRates()
			ge.recalculateTickSpeed()
			return kind.spend(t, p, ge)
		}
		before, skip := read(false)
		after, _ := read(true)
		out.Delivered, out.Skip = after-before, skip
		// What the game says about it, and whether any cheap meter moves:
		// none is part of a promise read this way.
		ge := l.engine(age, mode)
		snap := takeTruthSnapshot(ge)
		sw := p.wire(ge)
		sw.off()
		off := readTruth(t, ge)
		sw.on()
		on := readTruth(t, ge)
		out.Said = ge.capNoteLocked(p.Eff, true)
		sw.restore()
		snap.put(ge)
		for _, m := range sortedKeys(on) {
			if truthMoved(off[m], on[m]) {
				out.Leaks = append(out.Leaks, fmt.Sprintf("%s %+.4g", m, on[m]-off[m]))
			}
		}
		return out
	}
	ge := l.engine(age, mode)
	if res := truthPaidIn(p); res != "" && !ge.Resources.IsUnlocked(res) {
		out.Skip = res + " is locked in " + age
		return out
	}
	snap := takeTruthSnapshot(ge)
	sw := p.wire(ge)
	pool := truthKneePool(p)
	sw.off()
	before := readTruth(t, ge)
	earnedOff := ge.buildResolver().AddTotal(pool)
	sw.on()
	after := readTruth(t, ge)
	earnedOn := ge.buildResolver().AddTotal(pool)
	out.Said = ge.capNoteLocked(p.Eff, true)
	if pool != "" {
		out.rule(earnedOff, earnedOn)
	}
	sw.restore()
	snap.put(ge)
	m := kind.measure(p, ge, before, after)
	out.Delivered, out.Noise, out.Skip = m.Delivered, m.Noise, m.Skip
	allowed := m.Allowed
	if allowed == nil {
		return out
	}
	faith := truthMoved(before["rate:faith"], after["rate:faith"])
	for _, m := range sortedKeys(after) {
		if !truthMoved(before[m], after[m]) || allowed(m) || p.also(m) {
			continue
		}
		if m == "morale" && faith {
			continue // faith income lifts morale (site/docs/faith.md): the rate's own doing
		}
		if strings.HasPrefix(m, "stock:") && p.Kind == "storage" {
			continue // a store that shrinks clips its stock
		}
		out.Leaks = append(out.Leaks, fmt.Sprintf("%s %+.4g", m, after[m]-before[m]))
	}
	return out
}

// ----- the inventory -----

func truthPercent(v float64) string { return fmt.Sprintf("%+g%%", math.Round(v*1000)/10) }

// truthEffectText words an effect the way the game does.
func truthEffectText(e config.Effect) string {
	switch e.Type {
	case "bonus", "permanent_bonus":
		return truthPercent(e.Value) + " " + EffectTargetName(e.Target)
	case "production":
		return fmt.Sprintf("%+g %s/tick", e.Value, ResourceName(e.Target))
	case "storage":
		switch e.Target {
		case "all":
			return fmt.Sprintf("%+g storage for every resource", e.Value)
		case "soldiers":
			return fmt.Sprintf("%+g soldier storage", e.Value)
		}
		return fmt.Sprintf("%+g %s storage", e.Value, ResourceName(e.Target))
	case "capacity":
		if e.Target == "population" {
			return fmt.Sprintf("%+g housing", e.Value)
		}
		return fmt.Sprintf("capacity %s %+g", e.Target, e.Value)
	case "instant_resource":
		return fmt.Sprintf("%+g %s at once", e.Value, ResourceName(e.Target))
	case "steal_resource":
		return fmt.Sprintf("up to %g %s lost", e.Value, ResourceName(e.Target))
	case "worker_loss":
		return fmt.Sprintf("%g%% of your workers lost", e.Value*100)
	case "morale":
		return fmt.Sprintf("%+g morale points a tick", math.Round(e.Value*1e6)/1e4)
	case "opinion":
		return fmt.Sprintf("%+g opinion/tick per worker", e.Value)
	case "trade_route_income":
		return truthPercent(e.Value) + " trade route income"
	case "production_all":
		return truthPercent(e.Value) + " all production"
	case "tick_speed":
		return truthPercent(e.Value) + " game speed"
	}
	if res, ok := strings.CutSuffix(e.Type, "_rate"); ok {
		return truthPercent(e.Value) + " " + ResourceName(res) + " production"
	}
	return fmt.Sprintf("%s %s %+g", e.Type, e.Target, e.Value)
}

// truthTechPromises is one promise per effect of every tech. The tech is
// researched either way; the off reading empties its effects and the on
// reading holds the one effect alone, so effects of one tech never mix.
func truthTechPromises() []truthPromise { return truthTechPromisesOf(config.Technologies()) }

// truthTechPromisesOf is truthTechPromises for the given techs. A tech the
// engine does not hold yet (a test's made-up one) joins it for the probe.
func truthTechPromisesOf(defs []config.TechDef) []truthPromise {
	var out []truthPromise
	for _, def := range defs {
		for _, typed := range def.Effects {
			// A tech's effect is typed (a kind and a target). The meters
			// read the general Effect every other source is written in.
			eff := typed.Effect()
			key := def.Key
			out = append(out, truthPromise{
				Source: "tech", Key: def.Key, Name: def.Name, Age: def.Age, Eff: eff, Count: 1,
				Text: typed.Text(), Kind: truthEffectKind("tech", eff),
				wire: func(ge *GameEngine) truthSwitch {
					rm := ge.Research
					saved, known := rm.defs[key]
					order := rm.order
					if !known {
						saved = def
						rm.order = append(append([]string(nil), order...), key)
					}
					set := func(effects []config.TechEffect) func() {
						return func() {
							d := saved
							d.Effects = effects
							rm.defs[key] = d
							rm.researched[key] = true
							rm.rebuildBonuses()
						}
					}
					return truthSwitch{
						off: set(nil),
						on:  set([]config.TechEffect{typed}),
						restore: func() {
							rm.defs[key] = saved
							if !known {
								delete(rm.defs, key)
								rm.order = order
							}
						},
					}
				},
			})
		}
	}
	return out
}

// truthMilestoneAge is the first age a milestone's conditions can be met
// in: the latest of its own MinAge and the ages of the buildings, techs,
// wonders and resources it names, of the housing its population needs and
// of the tech count it asks for. Counters that only play fills (structures
// built, soldiers trained) add no age of their own.
func truthMilestoneAge(l *truthLab, def config.MilestoneDef) string {
	order := ageOrders()
	keys := ageKeys()
	at := 0
	later := func(age string) {
		if i, ok := order[age]; ok && i > at {
			at = i
		}
	}
	later(def.MinAge)
	blds := config.BuildingByKey()
	for key := range def.MinBuildings {
		later(blds[key].RequiredAge)
	}
	if len(def.MinBuildingSum.Keys) > 0 {
		first := len(keys) - 1
		for _, key := range def.MinBuildingSum.Keys {
			first = min(first, order[blds[key].RequiredAge])
		}
		later(keys[first])
	}
	techs := config.TechByKey()
	for _, key := range def.RequiredTechs {
		later(techs[key].Age)
	}
	for res := range def.MinResources {
		later(config.ResourceByKey()[res].Age)
	}
	if def.MinSoldiersTrained > 0 {
		later(config.ResourceByKey()["soldiers"].Age)
	}
	if def.MinTechCount > 0 || def.MinWonders > 0 || def.MinPopulation > 0 {
		for i, a := range keys {
			n, w := 0, 0
			for _, t := range techs {
				if order[t.Age] <= i {
					n++
				}
			}
			for _, b := range blds {
				if b.Category == "wonder" && order[b.RequiredAge] <= i {
					w++
				}
			}
			if n >= def.MinTechCount && w >= def.MinWonders && l.engine(a, truthClean).popCapLocked() >= def.MinPopulation {
				later(a)
				break
			}
		}
	}
	return keys[at]
}

// truthMilestonePromises is one promise per reward of every milestone,
// granted through the engine's own reward step.
func truthMilestonePromises(l *truthLab) []truthPromise {
	var out []truthPromise
	for _, def := range config.Milestones() {
		age := truthMilestoneAge(l, def)
		for _, eff := range def.Rewards {
			out = append(out, truthPromise{
				Source: "milestone", Key: def.Key, Name: def.Name, Age: age, Eff: eff, Count: 1,
				Text: milestoneRewardParts([]config.Effect{eff}), Kind: truthEffectKind("milestone", eff),
				wire: func(ge *GameEngine) truthSwitch {
					return truthSwitch{
						off:     func() {},
						on:      func() { ge.applyMilestoneRewards([]config.Effect{eff}) },
						restore: func() {},
					}
				},
			})
		}
	}
	return out
}

// truthChainPromises is the game speed boost of every milestone chain.
func truthChainPromises(l *truthLab) []truthPromise {
	var out []truthPromise
	byKey := config.MilestoneByKey()
	for _, chain := range config.MilestoneChains() {
		age := ageKeys()[0]
		for _, k := range chain.MilestoneKeys {
			if a := truthMilestoneAge(l, byKey[k]); ageOrders()[a] > ageOrders()[age] {
				age = a
			}
		}
		eff := config.Effect{Type: "tick_speed", Target: "tick_speed", Value: chain.BoostValue}
		out = append(out, truthPromise{
			Source: "chain", Key: chain.Key, Name: chain.Name, Age: age, Eff: eff, Count: 1,
			Text: "Game speed " + truthPercent(chain.BoostValue) + " for a while", Kind: "game_speed",
			wire: func(ge *GameEngine) truthSwitch {
				return truthSwitch{
					off:     func() {},
					on:      func() { ge.startChainBoost(chain, chain.BoostDuration) },
					restore: func() {},
				}
			},
		})
	}
	return out
}

// truthBuildingPromises is one promise per effect of every building,
// wonders and monuments included. The building stands either way (as many
// copies as the lab holds, or one of a wonder or monument); the off reading
// empties its effects and the on reading holds the one effect alone.
func truthBuildingPromises() []truthPromise {
	var out []truthPromise
	for _, def := range config.BaseBuildings() {
		source := "building"
		if def.Category == "wonder" || def.Category == "monument" {
			source = def.Category
		}
		for _, eff := range def.Effects {
			key := def.Key
			count := float64(truthCopies)
			if source != "building" {
				count = 1
			}
			// An output the building's own text marks "once unlocked".
			text, later := truthEffectText(eff), false
			if eff.Type == "production" && config.MakesBeforeUnlock(def, eff.Target) && strings.Contains(def.Description, "/tick "+config.OnceUnlocked) {
				text, later = text+" "+config.OnceUnlocked, true
			}
			out = append(out, truthPromise{
				Source: source, Key: def.Key, Name: def.Name, Age: def.RequiredAge, Eff: eff, Count: count,
				Text: text, Kind: truthEffectKind(source, eff), Later: later,
				wire: func(ge *GameEngine) truthSwitch {
					bm := ge.Buildings
					saved := bm.defs[key]
					set := func(effects []config.Effect) func() {
						return func() {
							d := saved
							d.Effects = effects
							bm.defs[key] = d
							bm.counts[key] = int(count)
						}
					}
					return truthSwitch{
						off:     set(nil),
						on:      set([]config.Effect{eff}),
						restore: func() { bm.defs[key] = saved },
					}
				},
			})
		}
	}
	return out
}

// truthEventAge is the first age an event can fire in.
func truthEventAge(def config.EventDef) string {
	age := def.MinAge
	if ep, ok := config.EpochByKey()[def.EpochKey]; ok && len(ep.Ages) > 0 {
		if ageOrders()[ep.Ages[0]] > ageOrders()[age] {
			age = ep.Ages[0]
		}
	}
	if _, ok := ageOrders()[age]; !ok {
		age = ageKeys()[0]
	}
	return age
}

// truthEventSwitch fires one effect the way a triggered event does: a
// timed effect joins the active events for the event's duration, an instant
// one is applied on the spot (EventManager.Tick, applyEventEffects). An
// instant event's timed effect is never applied, and reads as nothing.
func truthEventSwitch(ge *GameEngine, def config.EventDef, eff config.Effect) truthSwitch {
	return truthSwitch{
		off: func() { ge.Resources.resources["soldiers"].Amount = 0 }, // no garrison: a raid takes its full share
		on: func() {
			one := def
			one.Effects = []config.Effect{eff}
			if one.Duration > 0 {
				ge.Events.active = append(ge.Events.active, ActiveEvent{Key: one.Key, Name: one.Name, TicksLeft: one.Duration, Effects: one.Effects})
			}
			ge.applyEventEffects(one)
		},
		restore: func() {},
	}
}

// truthEventPromises is one promise per effect of every random and
// epoch-exclusive event.
func truthEventPromises() []truthPromise {
	var out []truthPromise
	add := func(source string, defs []config.EventDef) {
		for _, def := range defs {
			for _, eff := range def.Effects {
				out = append(out, truthPromise{
					Source: source, Key: def.Key, Name: def.Name, Age: truthEventAge(def), Eff: eff, Count: 1,
					Text: truthEffectText(eff), Kind: truthEffectKind(source, eff), Raid: def.Raid,
					wire: func(ge *GameEngine) truthSwitch { return truthEventSwitch(ge, def, eff) },
				})
			}
		}
	}
	add("event", config.RandomEvents())
	add("era event", config.EpochExclusiveEvents())
	return out
}

// truthAwakeningPromises is one promise per effect of every awakening,
// injected as fireAwakening injects it.
func truthAwakeningPromises() []truthPromise {
	var out []truthPromise
	for _, def := range config.Awakenings() {
		for _, eff := range def.Effects {
			out = append(out, truthPromise{
				Source: "awakening", Key: def.Key, Name: def.Name, Age: def.TriggerAge, Eff: eff, Count: 1,
				Text: truthEffectText(eff), Kind: truthEffectKind("awakening", eff),
				wire: func(ge *GameEngine) truthSwitch {
					return truthSwitch{
						off: func() {},
						on: func() {
							ge.Events.InjectEvent(ActiveEvent{Key: def.Key, Name: def.Name, TicksLeft: def.Duration, Effects: []config.Effect{eff}})
						},
						restore: func() {},
					}
				},
			})
		}
	}
	return out
}

var (
	truthAllRe  = regexp.MustCompile(`[Aa]ll production ([+-]\d+)%`)
	truthFlatRe = regexp.MustCompile(`([A-Za-z][a-z ]*?) ([+-][\d.]+)/tick`)
)

// truthEpochEventPromises reads the rate promises out of every epoch
// event's own text ("All production +100%", "Gold +5/tick") and measures
// them with the whole event applied through the engine's own step. The
// events are code, not data, so the text is the only statement of what they
// do.
func truthEpochEventPromises(t *testing.T) []truthPromise {
	var out []truthPromise
	labels := map[string]string{}
	for _, r := range config.BaseResources() {
		labels[strings.ToLower(r.Name)] = r.Key
	}
	// Epoch events fire on entering an era, from the Iron Era on: the first
	// age of each. One that pays in a resource waits for an era the
	// resource is unlocked in (rollGoodEpochEvent holds the Cultural
	// Festival back until culture exists).
	var entries []string
	for _, ep := range config.Epochs() {
		if config.CatastropheAllowed(ep.Key) {
			entries = append(entries, ep.Ages[0])
		}
	}
	entryFor := func(res string) string {
		for _, age := range entries {
			if ageOrders()[config.ResourceByKey()[res].Age] <= ageOrders()[age] {
				return age
			}
		}
		return entries[len(entries)-1]
	}
	add := func(defs []config.EpochEventDef, good bool) {
		for _, def := range defs {
			age := entries[0]
			wire := func(ge *GameEngine) truthSwitch {
				return truthSwitch{
					off: func() {},
					on: func() {
						// The event is applied whole. Its one-off parts (workers
						// gained or lost, buildings burned or given, stock, free
						// techs) are put back, so the rates read only what the
						// text says lasts: the timed and permanent effects.
						keep := takeTruthSnapshot(ge)
						if good {
							ge.applyGoodEpochEvent(def)
						} else {
							ge.applyChallengingEpochEvent(def, ge.currentEpoch)
						}
						keep.events = append([]ActiveEvent(nil), ge.Events.active...)
						keep.permanent = map[string]float64{}
						for k, v := range ge.permanentBonuses {
							keep.permanent[k] = v
						}
						keep.put(ge)
					},
					restore: func() {},
				}
			}
			first := len(out)
			for _, m := range truthAllRe.FindAllStringSubmatch(def.FlavorText, -1) {
				v, _ := strconv.ParseFloat(m[1], 64)
				out = append(out, truthPromise{
					Source: "epoch event", Key: def.Key, Name: def.Name, Age: age, Count: 1,
					Eff:  config.Effect{Type: "production_all", Value: v / 100},
					Text: m[0], Kind: "all_production", wire: wire,
				})
			}
			for _, m := range truthFlatRe.FindAllStringSubmatch(def.FlavorText, -1) {
				v, _ := strconv.ParseFloat(m[2], 64)
				// The words before the number end in the resource's name:
				// "then culture", "and faith", "Gold".
				words := strings.Fields(strings.ToLower(m[1]))
				name, res, ok := strings.Join(words, " "), "", false
				for i := range words {
					if res, ok = labels[strings.Join(words[i:], " ")]; ok {
						break
					}
				}
				if name == "its production" {
					// "The epoch's main building material": the era's own.
					res, ok = config.EpochByKey()[config.EpochForAge(age)].PrimaryResource, true
				}
				if ok && def.Key == "cultural_festival" {
					age = entryFor("culture")
				}
				if !ok {
					t.Errorf("epoch event %s promises %q, and %q is no resource: word it as \"<Resource> +N/tick\"", def.Key, m[0], name)
					continue
				}
				out = append(out, truthPromise{
					Source: "epoch event", Key: def.Key, Name: def.Name, Age: age, Count: 1,
					Eff:  config.Effect{Type: "production", Target: res, Value: v},
					Text: m[0], Kind: "flat_rate", wire: wire,
				})
			}
			// The promises of one event are measured together, in one age,
			// and ride along with each other.
			for i := first; i < len(out); i++ {
				out[i].Age = age
			}
			var meters []string
			for _, p := range out[first:] {
				if p.Kind == "all_production" {
					meters = append(meters, "rate:")
				} else {
					meters = append(meters, "rate:"+p.Eff.Target)
				}
			}
			for i := first; i < len(out); i++ {
				out[i].Also = meters
			}
		}
	}
	add(config.GoodEpochEvents(), true)
	add(config.ChallengingEpochEvents(), false)
	return out
}

// truthBoonPromises is every boon and setback in the catalog at the top of
// its range, applied through the applier a civilization's gift uses.
func truthBoonPromises() []truthPromise {
	var out []truthPromise
	add := func(defs []boon.Def) {
		for _, def := range defs {
			b := boon.Boon{Kind: def.Kind, Polarity: def.Polarity, Name: def.Name, Magnitude: def.MagMax, DurationTicks: def.DurMax, InstantAmount: def.AmountMax}
			if def.Polarity == boon.Negative {
				b.Magnitude = def.MagMin
			}
			// A boon aims at the civilization's specialty or a resource of
			// the age; the lab aims at one a Bronze Age player makes.
			b.Resource = "food"
			if def.Resource != "" {
				b.Resource = def.Resource
			}
			age := "bronze_age" // the first civilization is met there
			var effs []config.Effect
			switch def.Kind {
			case boon.RateBuff:
				effs = append(effs, config.Effect{Type: b.Resource + "_rate", Target: b.Resource, Value: b.Magnitude})
			case boon.AllProduction:
				effs = append(effs, config.Effect{Type: "production_all", Target: "production_all", Value: b.Magnitude})
			case boon.TickSpeed:
				effs = append(effs, config.Effect{Type: "tick_speed", Target: "tick_speed", Value: b.Magnitude})
			case boon.InstantResource:
				effs = append(effs, config.Effect{Type: "instant_resource", Target: b.Resource, Value: b.InstantAmount})
			case boon.TempWorkers:
				effs = append(effs, config.Effect{Type: "boon_workers", Value: math.Round(b.InstantAmount)})
			case boon.WorkerLoss:
				effs = append(effs, config.Effect{Type: "boon_workers", Value: -math.Round(b.InstantAmount)})
			case boon.ResourceDrain:
				effs = append(effs, config.Effect{Type: "boon_drain", Target: b.Resource, Value: b.InstantAmount})
				if b.Magnitude != 0 {
					effs = append(effs, config.Effect{Type: "production_all", Target: "production_all", Value: b.Magnitude})
				}
			}
			for _, eff := range effs {
				kind := truthEffectKind("boon", eff)
				text := truthEffectText(eff)
				switch eff.Type {
				case "boon_workers":
					kind, text = "workers", fmt.Sprintf("%+g workers", eff.Value)
				case "boon_drain":
					kind, text = "boon_drain", fmt.Sprintf("%g%% of your %s lost", eff.Value*100, ResourceName(eff.Target))
				case "production_all":
					if def.Kind == boon.ResourceDrain {
						kind = "boon_all_production"
					}
				}
				out = append(out, truthPromise{
					Source: "boon", Key: strings.ReplaceAll(strings.ToLower(def.Name), " ", "_"), Name: def.Name, Age: age, Eff: eff, Count: 1,
					Text: text, Kind: kind,
					wire: func(ge *GameEngine) truthSwitch {
						return truthSwitch{
							off: func() {},
							on: func() {
								boon.Apply(b, boonApplier{ge: ge, name: "Lab", key: "lab", malus: def.Polarity == boon.Negative})
							},
							restore: func() {},
						}
					},
				})
			}
		}
	}
	add(boon.Catalog())
	add(boon.MalusCatalog())
	return out
}

// truthAllyPromises is every civilization's allied bonus.
func truthAllyPromises() []truthPromise {
	var out []truthPromise
	for _, def := range config.BaseFactions() {
		if def.TradeBonus == 0 {
			continue
		}
		eff := config.Effect{Type: "ally", Target: def.Specialty, Value: def.TradeBonus}
		out = append(out, truthPromise{
			Source: "ally", Key: def.Key, Name: def.Name, Age: def.MinAge, Eff: eff, Count: 1,
			Text: truthPercent(def.TradeBonus) + " " + ResourceName(def.Specialty) + " production while allied", Kind: "ally_bonus",
			wire: func(ge *GameEngine) truthSwitch {
				saved, had := ge.Diplomacy.factions[def.Key]
				gone := func() { delete(ge.Diplomacy.factions, def.Key) }
				return truthSwitch{
					off: gone,
					on: func() {
						ge.Diplomacy.factions[def.Key] = &FactionState{Discovered: true, Opinion: AllyOpinion, Status: "allied"}
					},
					restore: func() {
						if gone(); had {
							ge.Diplomacy.factions[def.Key] = saved
						}
					},
				}
			},
		})
	}
	return out
}

// truthLegacyPromises is what a Succumb leaves behind: each epoch's legacy
// bonus, resource by resource, and Ancient Knowledge, epoch after epoch.
func truthLegacyPromises() []truthPromise {
	var out []truthPromise
	var fallen []string // the epochs a catastrophe can strike in, in order
	for _, ep := range config.Epochs() {
		flag := func(ge *GameEngine) truthSwitch {
			return truthSwitch{
				off: func() { delete(ge.legacyBonuses, ep.Key) },
				on: func() {
					ge.legacyBonuses[ep.Key] = true
					ge.reapplyLegacyBonuses()
				},
				restore: func() {},
			}
		}
		bonuses := config.LegacyBonusForEpoch(ep.Key)
		for _, res := range sortedKeys(bonuses) {
			eff := config.Effect{Type: "permanent_bonus", Target: res + "_rate", Value: bonuses[res]}
			out = append(out, truthPromise{
				// A legacy is carried into every later run, from its first age.
				Source: "legacy", Key: ep.Key, Name: ep.Name + " legacy", Age: ageKeys()[0], Eff: eff, Count: 1,
				Text: truthEffectText(eff), Kind: "legacy_production", wire: flag, Later: true,
			})
		}
		if !config.CatastropheAllowed(ep.Key) {
			continue // no catastrophe strikes here: its legacy can't be earned
		}
		// Ancient Knowledge: research time x0.8 for each epoch succumbed
		// in, on top of the ones before it. The nth is measured with the
		// first n-1 already held, as a player collects them. It multiplies
		// what is left, so it sits in no pool and no cap holds it.
		before := append([]string(nil), fallen...)
		fallen = append(fallen, ep.Key)
		eff := config.Effect{Type: "ancient_knowledge", Target: "research_time", Value: 1 - SuccumbResearchTimeFactor}
		out = append(out, truthPromise{
			Source: "ancient knowledge", Key: ep.Key,
			Name: fmt.Sprintf("Ancient Knowledge, epoch %d (%s)", len(fallen), ep.Name), Age: ageKeys()[0], Eff: eff, Count: 1,
			Text: "research time " + ResearchFactorText(SuccumbResearchTimeFactor), Kind: "legacy_research",
			wire: func(ge *GameEngine) truthSwitch {
				return truthSwitch{
					off: func() {
						ge.legacyBonuses = map[string]bool{}
						for _, k := range before {
							ge.legacyBonuses[k] = true
						}
					},
					on:      func() { ge.legacyBonuses[ep.Key] = true },
					restore: func() {},
				}
			},
		})
	}
	return out
}

// truthOtherPromises is the handful of one-off sources: the Cosmic Legacy,
// the festival and Endure's Reconstruction Effort.
func truthOtherPromises() []truthPromise {
	// The Cosmic Legacy is no part of the all-production pool: it multiplies
	// production after the caps (cosmicLegacyFactor), so its effect names no
	// pool and no "capped" note can stand beside it.
	cosmic := config.Effect{Type: "cosmic_legacy", Target: "production", Value: CosmicLegacyProductionBonus}
	festival := config.Effect{Type: "production_all", Target: "production_all", Value: festivalBuffPercent}
	rebuild := config.Effect{Type: "production_all", Target: "production_all", Value: endureDebuffProduction}
	return []truthPromise{
		{
			// A catastrophe can strike from the Iron Era on.
			Source: "endure", Key: "reconstruction", Name: "Reconstruction Effort (Endure)", Age: "iron_age", Eff: rebuild, Count: 1,
			Text: truthEffectText(rebuild) + " for a while", Kind: "all_production",
			wire: func(ge *GameEngine) truthSwitch {
				return truthSwitch{off: func() {}, on: func() { ge.startReconstruction() }, restore: func() {}}
			},
		},
		{
			Source: "cosmic legacy", Key: "cosmic_legacy", Name: "Cosmic Legacy", Age: ageKeys()[0], Eff: cosmic, Count: 1,
			Text: truthPercent(cosmic.Value) + " all production, after the caps, permanently", Kind: "final_production",
			wire: func(ge *GameEngine) truthSwitch {
				had := ge.cosmicLegacy
				return truthSwitch{
					off:     func() { ge.cosmicLegacy = false },
					on:      func() { ge.cosmicLegacy = true },
					restore: func() { ge.cosmicLegacy = had },
				}
			},
		},
		{
			// Culture, the festival's price, arrives in the Classical Age.
			Source: "festival", Key: "festival", Name: "Cultural Festival", Age: config.ResourceByKey()["culture"].Age, Eff: festival, Count: 1,
			Text: truthEffectText(festival) + " for a while", Kind: "festival",
			wire: func(ge *GameEngine) truthSwitch {
				ready := ge.festivalReadyTick
				return truthSwitch{
					off: func() {},
					on: func() {
						c := ge.Resources.resources["culture"]
						c.Amount = math.Max(c.Amount, ge.festivalCost())
						ge.festivalReadyTick = 0
						if err := ge.DoFestival(); err != nil {
							panic(err)
						}
					},
					restore: func() { ge.festivalReadyTick = ready },
				}
			},
		},
	}
}

// truthPromises is every promise the guard measures.
func truthPromises(t *testing.T, l *truthLab) []truthPromise {
	var out []truthPromise
	out = append(out, truthTechPromises()...)
	out = append(out, truthMilestonePromises(l)...)
	out = append(out, truthChainPromises(l)...)
	out = append(out, truthBuildingPromises()...)
	out = append(out, truthEventPromises()...)
	out = append(out, truthAwakeningPromises()...)
	out = append(out, truthEpochEventPromises(t)...)
	out = append(out, truthBoonPromises()...)
	out = append(out, truthAllyPromises()...)
	out = append(out, truthLegacyPromises()...)
	out = append(out, truthOtherPromises()...)
	return out
}

// ----- the verdict -----

// The classes a promise can land in.
const (
	truthOK          = "OK"
	truthDead        = "DEAD"
	truthCapped      = "CAPPED"
	truthWrongSize   = "WRONG SIZE"
	truthWrongTarget = "WRONG TARGET"
	truthUnmeasured  = "NOT MEASURED"
	// truthUnsaid is a cap the game does not own up to: the bonus falls
	// short and nothing beside it says "capped", or it says "capped" and
	// the bonus is all there.
	truthUnsaid = "CAP NOT SHOWN"
	// truthLocked is a promise about a resource the player cannot have yet
	// in the age the promise is earned in, with no word of that in its text:
	// the engine applies no rate to a locked resource.
	truthLocked = "LOCKED RESOURCE"
)

// truthVerdict is what the guard found for one promise.
type truthVerdict struct {
	P     truthPromise
	Class string
	// MeasuredAge is the age the clean reading was taken in: the promise's
	// own, or the first one after it with something for the bonus to raise.
	MeasuredAge string
	Clean       truthOutcome
	// Typical is the typical reading in every age measured, in order.
	Typical []truthOutcome
	// ShortFrom is the first age the typical player gets less than promised
	// ("" when never), and GoneFrom the first it gets nothing at all.
	ShortFrom, GoneFrom string
	// QuarterFrom is the first age the bonus counts less than promised
	// because its pool is past +200%, by exactly what the soft cap says (""
	// when never). Quarters counts the readings taken with the pool past
	// +200% before the bonus, where it must deliver a quarter of its value.
	QuarterFrom string
	Quarters    int
	// Unsaid is where what the game says about a cap and what the bonus
	// delivers part ways ("" when they never do).
	Unsaid string
	Note   string
}

// truthPoolResource is the resource whose own pool a promise sits in ("" for
// a promise in no resource's pool).
func truthPoolResource(p truthPromise) string {
	switch p.Kind {
	case "resource_production", "legacy_production", "ally_bonus":
		return truthRateTarget(p.Eff)
	case "tech_output":
		return p.Eff.Target
	}
	return ""
}

// truthPaidIn is the resource a promise pays in or takes from ("" when it
// is not about one resource's rate or stock).
func truthPaidIn(p truthPromise) string {
	switch p.Kind {
	case "flat_rate", "building_output", "instant", "steal", "drain", "boon_drain":
		return p.Eff.Target
	}
	return truthPoolResource(p)
}

// judge measures one promise and classes it.
func (l *truthLab) judge(t *testing.T, p truthPromise) truthVerdict {
	t.Helper()
	v := truthVerdict{P: p, MeasuredAge: p.Age}
	kind, ok := truthKinds[p.Kind]
	if !ok {
		v.Class, v.Note = truthUnmeasured, "no meter for this effect"
		return v
	}
	// A resource that is still locked gathers nothing, whatever its rate
	// reads. A promise that says it is for later is measured from the age
	// its resource unlocks in; any other is one the game cannot keep.
	if res := truthPaidIn(p); res != "" {
		unlock := config.ResourceByKey()[res].Age
		if ageOrders()[unlock] > ageOrders()[p.Age] {
			if !p.Later {
				v.Class, v.Note = truthLocked, ResourceName(res)+" is locked until "+AgeName(unlock)+", and nothing in the text says so"
				return v
			}
			v.MeasuredAge = unlock
			if p.Source != "legacy" {
				v.Note = "nothing until " + AgeName(unlock) + ": " + ResourceName(res) + " is locked before it, and the description says \"" + config.OnceUnlocked + "\""
			}
		}
	}
	// An ally's bonus multiplies the whole rate, flat income included, and
	// sits outside every pool: its exact reading is the typical player's.
	exact := truthClean
	if p.Kind == "ally_bonus" {
		exact = truthTypical
	}
	// A bonus to a resource has nothing to raise before the resource comes in.
	if res := truthPoolResource(p); res != "" {
		first := l.firstMade(res, v.MeasuredAge, exact == truthTypical)
		if first == "" {
			v.Class, v.Note = truthDead, "nothing makes "+ResourceName(res)+" from "+AgeName(p.Age)+" on"
			return v
		}
		if first != v.MeasuredAge && p.Source != "legacy" {
			v.Note = "nothing to raise until " + AgeName(first) + ": nothing makes " + ResourceName(res) + " before it"
		}
		v.MeasuredAge = first
	}
	v.Clean = l.probe(t, p, v.MeasuredAge, exact)
	switch c := v.Clean; {
	case c.Skip != "":
		v.Class, v.Note = truthUnmeasured, c.Skip
		return v
	case !c.full() && c.Said != "":
		// Short with nothing else in the way, and the game says a cap holds
		// it: the bonus alone runs past its pool's limit (or, for Ancient
		// Knowledge, the epochs before it already fill the pool).
		v.Class, v.ShortFrom = truthCapped, v.MeasuredAge
		if c.is(0) {
			v.GoneFrom = v.MeasuredAge
		}
		v.Typical = []truthOutcome{c}
		return v
	case len(c.Leaks) > 0 && c.is(0) && !c.full():
		v.Class, v.Note = truthWrongTarget, "moves "+strings.Join(c.Leaks, ", ")+" and not what it names"
		return v
	case c.is(0) && !c.full():
		v.Class = truthDead
		return v
	case !c.full():
		v.Class = truthWrongSize
		return v
	case len(c.Leaks) > 0:
		v.Class, v.Note = truthWrongTarget, "also moves "+strings.Join(c.Leaks, ", ")
		return v
	}
	// It works. Does the typical player get all of it? In the age it is
	// earned in and the two after; a bonus in a pool also in the last age,
	// where every pool is at its fullest, or in every age when the report
	// wants to know where exactly a cap starts to bite.
	keys := ageKeys()
	ages := keys[ageOrders()[v.MeasuredAge]:]
	if n := len(ages); n > 3 && !(kind.Pool && l.everyAge) {
		ages = ages[:3:3]
		if kind.Pool {
			ages = append(ages, keys[len(keys)-1])
		}
	}
	v.Class = truthOK
	for _, age := range ages {
		o := l.probe(t, p, age, truthTypical)
		if o.Skip != "" {
			continue
		}
		v.Typical = append(v.Typical, o)
		if (o.Said != "") == o.full() && v.Unsaid == "" {
			v.Unsaid = fmt.Sprintf("in %s it delivers %s of %s and the game says %q", AgeName(age), fmtG(o.Delivered), fmtG(o.Promised), o.Said)
		}
		switch {
		case o.full():
		case o.Knee && o.is(o.Rule) && !o.is(0) && (!o.Past || o.is(0.25*o.Promised)):
			// Past +200%, and worth exactly what the soft cap says: a
			// quarter of its value when the pool was past the knee without
			// it. The note beside it must say that number.
			if v.QuarterFrom == "" {
				v.QuarterFrom = age
			}
			if o.Past {
				v.Quarters++
			}
			if want := "counts a quarter past +200%: " + pointsText(o.Rule) + " now"; o.Said != want && v.Unsaid == "" {
				v.Unsaid = fmt.Sprintf("in %s it delivers %s of %s, as the soft cap says, and the game says %q, not %q", AgeName(age), fmtG(o.Delivered), fmtG(o.Promised), o.Said, want)
			}
		default:
			v.Class = truthCapped
			if v.ShortFrom == "" {
				v.ShortFrom = age
			}
			if v.GoneFrom == "" && o.is(0) {
				v.GoneFrom = age
			}
		}
		if len(o.Leaks) > 0 && v.Class == truthOK {
			v.Class, v.Note = truthWrongTarget, "in "+AgeName(age)+" also moves "+strings.Join(o.Leaks, ", ")
		}
	}
	return v
}

// truthPool names the pool a capped promise sits in, as truthAccepted keys
// it: the meter kind, and the resource for a resource's own pool.
func truthPool(p truthPromise) string {
	if res := truthPoolResource(p); res != "" {
		return "resource_production:" + res
	}
	switch {
	case strings.HasSuffix(p.Kind, "all_production"), p.Kind == "festival":
		return "all_production"
	case strings.HasSuffix(p.Kind, "research"), p.Kind == "research_speed":
		return "research_speed"
	}
	return p.Kind
}

// truthAccepted is the allow-list: the promises the guard knows the game
// does not keep, and why that is accepted for now. Keys are "CAPPED/<pool>"
// for a cap, "<class>/<source> <key>" for anything else. Every entry must
// still be needed: the test fails on one nothing uses, so a fixed promise
// takes its excuse with it.
var truthAccepted = map[string]string{
	"LOCKED RESOURCE/ally stellar_federation": "the Stellar Federation is met in the Space Age and its specialty, dark matter, unlocks in the Interstellar Age: " +
		"an alliance made early pays nothing for one age. Moving the civilization or the resource is a design call.",
}

// TestBonusTruth is the guard: every promise measured, every miss a failure
// unless truthAccepted carries it.
//
// The all-production pool used to be on that list: the engine applied +200%
// of it at most, and the typical player's wonders alone pass that in the
// last age. The soft cap took the excuse away and left a check in its
// place: every bonus read with its pool past +200% must move the rate by a
// quarter of its value, and the guard fails if it took no such reading.
func TestBonusTruth(t *testing.T) {
	lab := newTruthLab()
	lab.everyAge = os.Getenv("BONUS_TRUTH_REPORT") != ""
	promises := truthPromises(t, lab)
	used := map[string]bool{}
	var verdicts []truthVerdict
	quarters := map[string]int{} // readings past the knee, by source
	for _, p := range promises {
		if p.Kind == "" {
			key := truthUnmeasured + "/" + p.Eff.Type + ":" + p.Eff.Target
			if _, ok := truthAccepted[key]; ok {
				used[key] = true
				verdicts = append(verdicts, truthVerdict{P: p, Class: truthUnmeasured, MeasuredAge: p.Age, Note: truthAccepted[key]})
				continue
			}
			t.Errorf("%s: no meter for effect type %q, target %q. An effect nothing measures can be silently ignored by the engine: give it a meter in truthEffectKind and truthKinds.", p.ID(), p.Eff.Type, p.Eff.Target)
			continue
		}
		v := lab.judge(t, p)
		verdicts = append(verdicts, v)
		quarters[p.Source] += v.Quarters
		if v.Unsaid != "" {
			t.Errorf("%s: %s: %s. Every list of bonuses must say so beside one a limit holds back (\"capped\", or what it counts past +200%%), and only beside those (game/caps.go).", p.ID(), truthUnsaid, v.Unsaid)
		}
		if v.Class == truthOK {
			continue
		}
		// A cap is accepted for its whole pool; anything else, one source
		// at a time.
		key := v.Class + "/" + p.Source + " " + p.Key
		if v.Class == truthCapped {
			key = v.Class + "/" + truthPool(p)
		}
		if _, ok := truthAccepted[key]; ok {
			used[key] = true
			continue
		}
		switch v.Class {
		case truthCapped:
			t.Errorf("%s: CAPPED. Earned in %s it delivers %s to the typical player from %s on (promised %s %s; the soft cap leaves %s of it there). A cap swallows it; if that is accepted, add %q to truthAccepted with the reason.",
				p.ID(), AgeName(p.Age), truthShort(v), AgeName(v.ShortFrom), fmtG(p.Eff.Value), truthKinds[p.Kind].Unit, truthRule(v), key)
		default:
			t.Errorf("%s: %s in %s. Promised %s %s, delivered %s. %s",
				p.ID(), v.Class, AgeName(v.MeasuredAge), fmtG(v.Clean.Promised), truthKinds[p.Kind].Unit, fmtG(v.Clean.Delivered), v.Note)
		}
	}
	for key := range truthAccepted {
		if !used[key] {
			t.Errorf("truthAccepted[%q] excuses nothing any more: remove it", key)
		}
	}
	// The soft cap's check must have had something to check: milestone
	// rewards, wonders, monuments and the timed bonuses (a festival, an era
	// event such as the power surge, a civilization's boon) each read with
	// the all-production pool past +200%, each worth a quarter there.
	t.Logf("readings with the pool past +200%%, by source: %v", quarters)
	for _, source := range []string{"milestone", "wonder", "monument", "festival", "epoch event", "boon"} {
		if quarters[source] == 0 {
			t.Errorf("no %s bonus was read with its pool past +200%%: the check that a bonus past the knee moves the rate by a quarter of its value proved nothing for them", source)
		}
	}
	if path := os.Getenv("BONUS_TRUTH_REPORT"); path != "" {
		if err := os.WriteFile(path, []byte(truthReport(verdicts)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func fmtG(v float64) string { return strconv.FormatFloat(math.Round(v*1e6)/1e6, 'g', -1, 64) }

// truthShort says how much of a capped promise the typical player gets in
// the first age it falls short.
func truthShort(v truthVerdict) string {
	for _, o := range v.Typical {
		if o.Age == v.ShortFrom {
			return fmtG(o.Delivered)
		}
	}
	return "less"
}

// truthRule says what the soft cap leaves of a short promise in the first
// age it falls short.
func truthRule(v truthVerdict) string {
	for _, o := range v.Typical {
		if o.Age == v.ShortFrom {
			return fmtG(o.Rule)
		}
	}
	return "all"
}

// truthReport is the full table, one line per promise, grouped by class.
func truthReport(verdicts []truthVerdict) string {
	var sb strings.Builder
	counts := map[string]int{}
	for _, v := range verdicts {
		counts[v.Class]++
	}
	fmt.Fprintf(&sb, "%d promises measured.\n\n", len(verdicts))
	for _, class := range []string{truthOK, truthCapped, truthDead, truthWrongSize, truthWrongTarget, truthLocked, truthUnmeasured} {
		fmt.Fprintf(&sb, "- %s: %d\n", class, counts[class])
	}
	for _, class := range []string{truthDead, truthWrongSize, truthWrongTarget, truthLocked, truthUnmeasured, truthCapped, truthOK} {
		var rows []string
		for _, v := range verdicts {
			if v.Class != class {
				continue
			}
			row := fmt.Sprintf("| %s | %s | %s | %s |", v.P.Source, v.P.Name, v.P.Text, AgeName(v.P.Age))
			switch class {
			case truthCapped:
				gone := "never fully"
				if v.GoneFrom != "" {
					gone = AgeName(v.GoneFrom)
				}
				row += fmt.Sprintf(" %s | %s |", AgeName(v.ShortFrom), gone)
			case truthOK:
				note := v.Note
				if v.QuarterFrom != "" {
					note = strings.TrimSpace(note + " counts a quarter past +200%, from " + AgeName(v.QuarterFrom) + " on for the typical player")
				}
				row += " " + note + " |"
			default:
				row += fmt.Sprintf(" promised %s, delivered %s. %s |", fmtG(v.Clean.Promised), fmtG(v.Clean.Delivered), v.Note)
			}
			rows = append(rows, row)
		}
		if len(rows) == 0 {
			continue
		}
		sort.Strings(rows)
		fmt.Fprintf(&sb, "\n## %s (%d)\n\n", class, len(rows))
		switch class {
		case truthCapped:
			sb.WriteString("| source | name | promise | earned in | short from | nothing from |\n|---|---|---|---|---|---|\n")
		case truthOK:
			sb.WriteString("| source | name | promise | earned in | note |\n|---|---|---|---|---|\n")
		default:
			sb.WriteString("| source | name | promise | earned in | what happens |\n|---|---|---|---|---|\n")
		}
		sb.WriteString(strings.Join(rows, "\n"))
		sb.WriteString("\n")
	}
	return sb.String()
}
