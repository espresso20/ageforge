package config

import (
	"fmt"
	"strconv"
)

// A tech's promises as data a test can check. Everything a tech does (an
// effect, a building or a wonder it opens, a command it opens) is one check:
// a number of the game, read with the tech and without it, and how the two
// must differ. TechChecks derives them from the tech's effects, the
// buildings that name it and the feature locks, so the list cannot drift
// from what the tech does.
//
// A check has two written forms. The atom is the short name of the promise
// ("out:food:10", "cost:3", "feature:campaigns"). The check is the test
// ("rate(food) x1.10", "build_price x0.97", "allowed(campaigns) 0>1"): x
// multiplies the number, + or - adds to it, 0>1 goes from no to yes, 0>pos
// from zero to above zero.

// TechCheckOp is how the number must differ with the tech.
type TechCheckOp string

const (
	// CheckTimes: the number is multiplied by Value.
	CheckTimes TechCheckOp = "x"
	// CheckPlus: Value is added to the number.
	CheckPlus TechCheckOp = "+"
	// CheckOpens: the answer goes from no to yes.
	CheckOpens TechCheckOp = "0>1"
	// CheckStarts: the number goes from zero to above zero.
	CheckStarts TechCheckOp = "0>pos"
)

// The numbers a check reads.
const (
	// MeasureRate is what the resource Target makes per tick, before the
	// food drain. Target AllResources is every resource that makes anything.
	MeasureRate = "rate"
	// MeasureStorage is the storage of every resource.
	MeasureStorage = "storage"
	// MeasureHousing is housing. The game rounds it up to a whole person.
	MeasureHousing = "housing"
	// MeasureBuildPrice is what a building costs.
	MeasureBuildPrice = "build_price"
	// MeasureBuildTicks is how long construction takes.
	MeasureBuildTicks = "build_ticks"
	// MeasureResearchTicks is how long research takes.
	MeasureResearchTicks = "research_ticks"
	// MeasureGameSpeed, MeasureMilitaryPower and MeasureExpeditionReward are
	// the pools of those names.
	MeasureGameSpeed        = "tick_speed"
	MeasureMilitaryPower    = "military_power"
	MeasureExpeditionReward = "expedition_reward"
	// MeasureMechanic is the mechanic number Target (Mechanics).
	MeasureMechanic = "mechanic"
	// MeasureCanBuild is whether the building Target can be built, its
	// price aside.
	MeasureCanBuild = "can_build"
	// MeasureAllowed is whether the command behind the feature lock Target
	// is open.
	MeasureAllowed = "allowed"
)

// AllResources is the target of a rate check on every resource.
const AllResources = "all"

// TechCheck is one promise of a tech, as a test reads it.
type TechCheck struct {
	// Atom is the promise's short name: "out:food:10".
	Atom string
	// Measure is the number read (the Measure constants) and Target what it
	// is read of: a resource, a mechanic, a building, a feature lock; ""
	// where the number has no target.
	Measure string
	Target  string
	// Op and Value say how the number differs with the tech. Value is the
	// factor for CheckTimes, the amount for CheckPlus, and unused for the
	// other two.
	Op    TechCheckOp
	Value float64
}

// String is the check as the tree's tables write it: "rate(food) x1.10",
// "market_fee -0.03", "can_build(barracks) 0>1".
func (c TechCheck) String() string {
	what := c.Measure
	switch {
	case c.Measure == MeasureMechanic:
		what = c.Target
	case c.Target != "":
		what = c.Measure + "(" + c.Target + ")"
	}
	switch c.Op {
	case CheckTimes:
		return what + " x" + strconv.FormatFloat(c.Value, 'f', 2, 64)
	case CheckPlus:
		s := strconv.FormatFloat(c.Value, 'f', -1, 64)
		if c.Value >= 0 {
			s = "+" + s
		}
		return what + " " + s
	}
	return what + " " + string(c.Op)
}

// percentAtom prints a fraction as the whole-ish percentage an atom carries:
// 0.1 is "10", -0.03 is "3".
func percentAtom(v float64) string {
	return strconv.FormatFloat(absFloat(v)*100, 'f', -1, 64)
}

// check is the check for one effect of a tech.
func (e TechEffect) check() TechCheck {
	times := func(atom, measure, target string) TechCheck {
		return TechCheck{Atom: atom, Measure: measure, Target: target, Op: CheckTimes, Value: 1 + e.Value}
	}
	plus := func(atom, measure, target string) TechCheck {
		return TechCheck{Atom: atom, Measure: measure, Target: target, Op: CheckPlus, Value: e.Value}
	}
	pct := percentAtom(e.Value)
	switch e.Kind {
	case EffectOutput:
		return times("out:"+e.Target+":"+pct, MeasureRate, e.Target)
	case EffectAllOutput:
		return times("all:"+pct, MeasureRate, AllResources)
	case EffectFlatOutput:
		return TechCheck{Atom: "first:" + e.Target, Measure: MeasureRate, Target: e.Target, Op: CheckStarts}
	case EffectStorage:
		return times("store:"+pct, MeasureStorage, "")
	case EffectHousing:
		return times("house:"+pct, MeasureHousing, "")
	case EffectBuildCost:
		return times("cost:"+pct, MeasureBuildPrice, "")
	case EffectBuildTime:
		return times("time:"+pct, MeasureBuildTicks, "")
	case EffectResearchTime:
		return times("research:"+pct, MeasureResearchTicks, "")
	case EffectGameSpeed:
		return plus("speed:"+pct, MeasureGameSpeed, "")
	case EffectMilitaryPower:
		return plus("mil:"+pct, MeasureMilitaryPower, "")
	case EffectExpeditionReward:
		return plus("exp:"+pct, MeasureExpeditionReward, "")
	case EffectMechanic:
		atom := fmt.Sprintf("p:%s:%s", e.Target, strconv.FormatFloat(e.Value, 'f', -1, 64))
		if MechanicByKey()[e.Target].Multiplies {
			return times(atom, MeasureMechanic, e.Target)
		}
		return plus(atom, MeasureMechanic, e.Target)
	}
	return TechCheck{Atom: string(e.Kind)}
}

// TechChecks lists every promise of tech t as a check: its effects in order,
// then the buildings that wait for it (a wonder among them is the one it is
// the keystone of), then the commands it opens. buildings and locks are the
// tables to read those from.
func TechChecks(t TechDef, buildings []BuildingDef, locks []FeatureLockDef) []TechCheck {
	var out []TechCheck
	for _, e := range t.Effects {
		out = append(out, e.check())
	}
	for _, b := range buildings {
		if b.RequiredTech != t.Key {
			continue
		}
		atom := "unlock:" + b.Key
		if b.Category == "wonder" {
			atom = "keystone:" + b.Key
		}
		out = append(out, TechCheck{Atom: atom, Measure: MeasureCanBuild, Target: b.Key, Op: CheckOpens})
	}
	for _, l := range locks {
		if l.Tech == t.Key {
			out = append(out, TechCheck{Atom: "feature:" + l.Key, Measure: MeasureAllowed, Target: l.Key, Op: CheckOpens})
		}
	}
	return out
}

// AllTechChecks is TechChecks for every tech of the game, by tech key.
func AllTechChecks() map[string][]TechCheck {
	buildings := BaseBuildings()
	locks := FeatureLocks()
	out := make(map[string][]TechCheck)
	for _, t := range Technologies() {
		out[t.Key] = TechChecks(t, buildings, locks)
	}
	return out
}
