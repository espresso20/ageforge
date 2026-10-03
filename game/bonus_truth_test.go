package game

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The bonus truth guard.
//
// Every bonus the game promises a player is switched on in a real engine and
// the thing its text names is measured: a resource's rate, storage, housing,
// research time, a building's price, game speed, the Defense Rating, loot.
// The test fails when a promise does not show up at the size promised.
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
//     the bonus is earned and the two ages after it. A bonus that works
//     clean and falls short here is CAPPED: a cap swallowed it.
//
// What a percentage promises. Bonuses of one kind add together: +30% and
// +20% make +50%, not +56%. So "+30% gold production" promises 30 points of
// the base rate, and that is what is measured: the change in the rate
// divided by what the buildings make before any bonus.
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
}

func newTruthLab() *truthLab { return &truthLab{engines: map[string]*GameEngine{}} }

func (l *truthLab) engine(age string, mode truthMode) *GameEngine {
	k := age + "/" + mode.String()
	if ge, ok := l.engines[k]; ok {
		return ge
	}
	ge := newTruthEngine(age, mode)
	l.engines[k] = ge
	return ge
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
	out["defense"] = truthDefense(ge)
	out["morale"] = truthMoraleLift(t, ge)
	return out
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

// ----- a promise -----

// truthPromise is one effect of one source: what the game says it does and
// how to switch it on.
type truthPromise struct {
	// ID names the promise: "tech tool_making: +15% worker output".
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
	// wire prepares the three states a probe needs in ge.
	wire func(ge *GameEngine) truthSwitch
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
	// Leaks are the meters that moved and were not part of the promise.
	Leaks []string
	// Skip is why nothing could be measured here ("" when measured).
	Skip string
}

func (o truthOutcome) ok() bool {
	return o.Skip == "" && truthClose(o.Delivered, o.Promised) && len(o.Leaks) == 0
}

// truthClose reports whether got is want within a part in a million (the
// reference tech and price round to whole ticks and units).
func truthClose(got, want float64) bool {
	return math.Abs(got-want) <= 2e-6*math.Max(1, math.Abs(want))
}

// truthMoved reports whether a meter changed at all.
func truthMoved(before, after float64) bool {
	return math.Abs(after-before) > 1e-9*math.Max(1, math.Max(math.Abs(before), math.Abs(after)))
}

// ----- the kinds: what each effect promises and which meter reads it -----

// truthKind reads what a promise delivered from two readings.
type truthKind struct {
	// Unit says what Delivered and Promised count ("points of base output").
	Unit string
	// measure returns the delivered amount, in the promise's own unit, and
	// the meters the promise is allowed to move. A non-empty skip says why
	// it cannot be measured in this engine.
	measure func(p truthPromise, ge *GameEngine, before, after truthReading) (delivered float64, allowed func(meter string) bool, skip string)
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

// truthAllFactor is the all-production multiplier in force, read off the
// reference resources: what one of them yields per unit its buildings make,
// with flat income and upkeep taken out by differencing two of the engine's
// own breakdown lines.
func truthAllFactor(ge *GameEngine, r truthReading) (float64, bool) {
	refs := truthRefResources(ge, r)
	if len(refs) == 0 {
		return 0, false
	}
	key := refs[0]
	b := ge.Resources.resources[key].Breakdown
	return (b.BuildingRate + b.BonusRate) / b.BuildingRate, true
}

var truthKinds = map[string]truthKind{
	// "+X% all production": X points on every resource buildings make.
	"all_production": {
		Unit: "points of base output",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			refs := truthRefResources(ge, before)
			if len(refs) == 0 {
				return 0, nil, "no building output to read"
			}
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, key := range refs {
				d := (after["rate:"+key] - before["rate:"+key]) / before["made:"+key] / ge.speedK()
				lo, hi = math.Min(lo, d), math.Max(hi, d)
			}
			if !truthClose(lo, hi) {
				return lo, meterHasPrefix("rate:"), fmt.Sprintf("resources disagree: %v to %v", lo, hi)
			}
			return lo, meterHasPrefix("rate:"), ""
		},
	},
	// "+X% <resource> production": X points on that resource alone.
	"resource_production": {
		Unit: "points of base output",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			res := strings.TrimSuffix(p.Eff.Target, "_rate")
			made := before["made:"+res]
			if made <= 0 {
				return 0, nil, "no " + res + " is made here"
			}
			all, ok := truthAllFactor(ge, before)
			if !ok {
				return 0, nil, "no building output to read"
			}
			d := (after["rate:"+res] - before["rate:"+res]) / made / all / ge.speedK()
			return d, meterIs("rate:" + res), ""
		},
	},
	// "+V <resource>/tick" from a tech or an event: V more per tick.
	"flat_rate": {
		Unit: "per tick",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			res := p.Eff.Target
			return (after["rate:"+res] - before["rate:"+res]) / ge.speedK(), meterIs("rate:" + res), ""
		},
	},
	// "+V <resource>/tick (N workers)" on a building: each fully staffed
	// copy makes V before bonuses.
	"building_output": {
		Unit: "per tick per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			res := p.Eff.Target
			return (after["made:"+res] - before["made:"+res]) / p.Count, meterIs("rate:"+res, "made:"+res), ""
		},
	},
	// "+V storage for every resource" or for one.
	"storage": {
		Unit: "storage",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			if p.Eff.Target != "all" {
				key := "storage:" + p.Eff.Target
				return (after[key] - before[key]) / p.Count / ge.speedK(), meterIs(key), ""
			}
			lo, hi := math.Inf(1), math.Inf(-1)
			for _, key := range ge.Resources.order {
				d := (after["storage:"+key] - before["storage:"+key]) / p.Count / ge.speedK()
				lo, hi = math.Min(lo, d), math.Max(hi, d)
			}
			if !truthClose(lo, hi) {
				return lo, meterHasPrefix("storage:"), fmt.Sprintf("resources disagree: %v to %v", lo, hi)
			}
			return lo, meterHasPrefix("storage:"), ""
		},
	},
	// "+V housing".
	"housing": {
		Unit: "housing",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			return (after["housing"] - before["housing"]) / p.Count, meterIs("housing"), ""
		},
	},
	// "-X% building costs": X points off every price.
	"build_cost": {
		Unit: "points of the listed price",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			return after["cost"] - before["cost"], meterIs("cost"), ""
		},
	},
	// "+X% game speed".
	"game_speed": {
		Unit: "points of game speed",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			return after["speed"] - before["speed"], meterIs("speed"), ""
		},
	},
	// "+X% military power": X points on the Defense Rating.
	"military_power": {
		Unit: "points of Defense Rating",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			return after["defense"] - before["defense"], meterIs("defense"), ""
		},
	},
	// A worship or culture building's morale lift per tick.
	"morale": {
		Unit: "morale per tick per copy",
		measure: func(p truthPromise, ge *GameEngine, before, after truthReading) (float64, func(string) bool, string) {
			return (after["morale"] - before["morale"]) / p.Count, meterIs("morale"), ""
		},
	},
}

// truthEffectKind names the meter for an effect, by the effect's type and
// target alone. "" is an effect the guard has no meter for: the test fails
// on it, so nothing a player is promised goes unmeasured.
func truthEffectKind(source string, e config.Effect) string {
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
		if res, ok := strings.CutSuffix(e.Target, "_rate"); ok {
			if _, known := config.ResourceByKey()[res]; known {
				return "resource_production"
			}
		}
	case "production":
		if _, known := config.ResourceByKey()[e.Target]; known {
			if source == "building" || source == "wonder" {
				return "building_output"
			}
			return "flat_rate"
		}
	case "storage":
		if _, known := config.ResourceByKey()[e.Target]; known || e.Target == "all" {
			return "storage"
		}
	case "capacity":
		if e.Target == "population" {
			return "housing"
		}
		if e.Target == "military" {
			return "soldier_capacity"
		}
	case "morale":
		return "morale"
	case "opinion":
		return "opinion"
	case "trade_route_income":
		return "route_income"
	case "instant_resource":
		if _, known := config.ResourceByKey()[e.Target]; known {
			return "instant"
		}
	case "steal_resource":
		if _, known := config.ResourceByKey()[e.Target]; known {
			return "steal"
		}
	case "worker_loss":
		return "worker_loss"
	case "production_all":
		return "all_production"
	case "tick_speed":
		return "game_speed"
	}
	if res, ok := strings.CutSuffix(e.Type, "_rate"); ok {
		if _, known := config.ResourceByKey()[res]; known {
			return "resource_production"
		}
	}
	return ""
}

// ----- measuring -----

// probe measures p in ge: one reading with the effect off, one with it on.
func (l *truthLab) probe(t *testing.T, p truthPromise, ge *GameEngine, age string, mode truthMode) truthOutcome {
	t.Helper()
	out := truthOutcome{Age: age, Mode: mode, Promised: p.Eff.Value}
	kind, ok := truthKinds[p.Kind]
	if !ok {
		out.Skip = "no meter for " + p.Kind
		return out
	}
	stock := map[string]float64{}
	for _, key := range ge.Resources.order {
		stock[key] = ge.Resources.resources[key].Amount
	}
	sw := p.wire(ge)
	sw.off()
	before := readTruth(t, ge)
	sw.on()
	after := readTruth(t, ge)
	sw.restore()
	for _, key := range ge.Resources.order {
		ge.Resources.resources[key].Amount = stock[key]
	}
	delivered, allowed, skip := kind.measure(p, ge, before, after)
	out.Delivered, out.Skip = delivered, skip
	if allowed == nil {
		return out
	}
	for _, m := range sortedKeys(after) {
		if strings.HasPrefix(m, "stock:") {
			continue // a store that shrank clips its stock; that is storage's own doing
		}
		if truthMoved(before[m], after[m]) && !allowed(m) {
			out.Leaks = append(out.Leaks, fmt.Sprintf("%s %+.4g", m, after[m]-before[m]))
		}
	}
	return out
}

// truthAgesFrom is age and the two ages after it.
func truthAgesFrom(age string) []string {
	keys := ageKeys()
	i := ageOrders()[age]
	return keys[i:min(i+3, len(keys))]
}

// ----- the inventory -----

// truthTechText words a tech effect the way the Research panel does.
func truthBonusText(e config.Effect) string {
	return fmt.Sprintf("%+.0f%% %s", e.Value*100, EffectTargetName(e.Target))
}

func truthEffectText(e config.Effect) string {
	switch e.Type {
	case "bonus", "permanent_bonus":
		return truthBonusText(e)
	case "production":
		return fmt.Sprintf("%+g %s/tick", e.Value, ResourceName(e.Target))
	case "storage":
		if e.Target == "all" {
			return fmt.Sprintf("%+g storage for every resource", e.Value)
		}
		return fmt.Sprintf("%+g %s storage", e.Value, ResourceName(e.Target))
	case "capacity":
		if e.Target == "population" {
			return fmt.Sprintf("%+g housing", e.Value)
		}
	case "instant_resource":
		return fmt.Sprintf("%+g %s at once", e.Value, ResourceName(e.Target))
	case "morale":
		return fmt.Sprintf("%+g morale points a tick", e.Value*100)
	case "production_all":
		return fmt.Sprintf("%+.0f%% all production", e.Value*100)
	case "tick_speed":
		return fmt.Sprintf("%+.0f%% game speed", e.Value*100)
	}
	return fmt.Sprintf("%s %s %+g", e.Type, e.Target, e.Value)
}

// truthTechPromises is one promise per effect of every tech. The tech is
// researched either way; the off reading empties its effects and the on
// reading holds the one effect alone, so effects of one tech never mix.
func truthTechPromises() []truthPromise {
	var out []truthPromise
	for _, def := range config.Technologies() {
		for _, eff := range def.Effects {
			key := def.Key
			out = append(out, truthPromise{
				Source: "tech", Key: def.Key, Name: def.Name, Age: def.Age, Eff: eff, Count: 1,
				Text: truthEffectText(eff), Kind: truthEffectKind("tech", eff),
				wire: func(ge *GameEngine) truthSwitch {
					rm := ge.Research
					saved, was := rm.defs[key], rm.researched[key]
					set := func(effects []config.Effect) func() {
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
						on:  set([]config.Effect{eff}),
						restore: func() {
							rm.defs[key] = saved
							if !was {
								delete(rm.researched, key)
							}
							rm.rebuildBonuses()
						},
					}
				},
			})
		}
	}
	return out
}

func TestBonusTruthExplore(t *testing.T) {
	lab := newTruthLab()
	var rows []string
	for _, p := range truthTechPromises() {
		for _, mode := range []truthMode{truthClean, truthTypical} {
			ages := []string{p.Age}
			if mode == truthTypical {
				ages = truthAgesFrom(p.Age)
			}
			for _, age := range ages {
				ge := lab.engine(age, mode)
				o := lab.probe(t, p, ge, age, mode)
				rows = append(rows, fmt.Sprintf("%-60s %-8s %-16s promised %-8.4g delivered %-10.4g leaks=%v skip=%q", p.ID(), mode, age, o.Promised, o.Delivered, o.Leaks, o.Skip))
			}
		}
	}
	sort.Strings(rows)
	if path := os.Getenv("BONUS_TRUTH_DUMP"); path != "" {
		_ = os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0o644)
	}
}
