package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
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
// "route_ticks -0.15", "can_build(barracks) 0>1".
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

// Check reports what is wrong with e ("" when nothing is): a kind that is
// not one of the constants, a target the kind does not take, a resource or a
// mechanic that does not exist, a value of zero, or a cut so deep it leaves
// nothing (a time or a price multiplied by zero or less). isResource says
// which keys are resources.
func (e TechEffect) Check(isResource func(string) bool) string {
	known := false
	for _, k := range TechEffectKinds() {
		known = known || k == e.Kind
	}
	switch {
	case !known:
		return fmt.Sprintf("unknown effect kind %q", e.Kind)
	case e.Value == 0:
		return fmt.Sprintf("%s effect with no value", e.Kind)
	case e.Kind == EffectMechanic:
		if _, ok := MechanicByKey()[e.Target]; !ok {
			return fmt.Sprintf("%s names %q, which is not a mechanic", e.Kind, e.Target)
		}
	case !e.Kind.TakesResource():
		if e.Target != "" {
			return fmt.Sprintf("%s takes no target, got %q", e.Kind, e.Target)
		}
	case e.Kind == EffectOutput && e.Target == ResearchTarget:
		// The work scholars do: not a resource, but a bonus to it is allowed.
	case !isResource(e.Target):
		return fmt.Sprintf("%s names %q, which is not a resource", e.Kind, e.Target)
	}
	if e.Key().Multiplies() && e.Value <= -1 {
		return fmt.Sprintf("%s of %v leaves nothing", e.Kind, e.Value)
	}
	return ""
}

// TechTreeProblems checks the rules the tree keeps and returns what breaks
// them, sorted, one line each (nil when nothing does):
//
//   - every key in Prerequisites and AnyOf is a tech, named once, and not
//     the tech itself;
//   - an either-or group has at least two keys;
//   - the graph has no cycle, counting an either-or key as an edge like any
//     other: a tech may never be its own ancestor, whichever branch is
//     taken;
//   - no tech needs a tech of a later age;
//   - every tech has a lane that exists, a code of two to TechCodeMax
//     capital letters that no other tech has, and an emblem of one glyph;
//   - a keystone opens in its wonder's age;
//   - no capstone is a keystone or sits on the spine: a capstone is never
//     required. (That the spine needs only the spine follows from how
//     TechKinds builds it.)
//
// ages is the age keys in order.
func TechTreeProblems(techs []TechDef, buildings []BuildingDef, ages []string) []string {
	var out []string
	bad := func(format string, args ...any) { out = append(out, fmt.Sprintf(format, args...)) }

	byKey := indexBy(techs, func(t TechDef) string { return t.Key })
	ageAt := AgePositions(ages)
	lanes := TechLaneByKey()

	codes := map[string]string{}
	for _, t := range techs {
		// What it needs.
		named := map[string]bool{}
		for _, list := range [][]string{t.Prerequisites, t.AnyOf} {
			for _, p := range list {
				pre, ok := byKey[p]
				switch {
				case p == t.Key:
					bad("%s needs itself", t.Key)
				case !ok:
					bad("%s needs %s, which is not a tech", t.Key, p)
				case named[p]:
					bad("%s names %s twice in what it needs", t.Key, p)
				case ageAt[pre.Age] > ageAt[t.Age]:
					bad("%s (%s) needs %s, a tech of a later age (%s)", t.Key, t.Age, p, pre.Age)
				}
				named[p] = true
			}
		}
		if len(t.AnyOf) == 1 {
			bad("%s has an either-or group of one (%s): write it as a prerequisite", t.Key, t.AnyOf[0])
		}

		// Lane, code, emblem.
		if _, ok := lanes[t.Lane]; !ok {
			bad("%s has no lane, or one that does not exist (%q)", t.Key, t.Lane)
		}
		switch other, taken := codes[t.Code]; {
		case !validTechCode(t.Code):
			bad("%s has the code %q: a code is 2 to %d capital letters", t.Key, t.Code, TechCodeMax)
		case taken:
			bad("%s and %s share the code %s: give one of them its own", other, t.Key, t.Code)
		default:
			codes[t.Code] = t.Key
		}
		if utf8.RuneCountInString(t.Emblem) != 1 {
			bad("%s has the emblem %q: an emblem is one glyph", t.Key, t.Emblem)
		}
	}

	// Cycles: a depth-first walk over every edge, either-or keys included.
	const (
		unseen = iota
		open
		done
	)
	state := make(map[string]int, len(techs))
	var path []string
	var walk func(key string)
	walk = func(key string) {
		t, ok := byKey[key]
		if !ok || state[key] == done {
			return
		}
		if state[key] == open {
			for i, k := range path {
				if k == key {
					bad("a cycle: %s", strings.Join(append(append([]string(nil), path[i:]...), key), " needs "))
					break
				}
			}
			return
		}
		state[key] = open
		path = append(path, key)
		for _, p := range t.Prerequisites {
			walk(p)
		}
		for _, p := range t.AnyOf {
			walk(p)
		}
		path = path[:len(path)-1]
		state[key] = done
	}
	for _, t := range techs {
		walk(t.Key)
	}

	// Keystones, the spine and capstones.
	for _, b := range buildings {
		if b.Category != "wonder" || b.RequiredTech == "" {
			continue
		}
		if t, ok := byKey[b.RequiredTech]; ok && t.Age != b.RequiredAge {
			bad("%s is the keystone of %s (%s) but opens in %s: a keystone opens in its wonder's age", t.Key, b.Key, b.RequiredAge, t.Age)
		}
	}
	kinds := TechKinds(techs, buildings)
	for _, t := range techs {
		switch {
		case !t.Capstone:
		case kinds[t.Key] == TechKeystone:
			bad("%s is flagged a capstone but a wonder requires it: a capstone is never required", t.Key)
		case kinds[t.Key] == TechSpine:
			bad("%s is flagged a capstone but a keystone stands on it: a capstone is never required", t.Key)
		}
	}

	sort.Strings(out)
	return out
}

// validTechCode reports whether code is 2 to TechCodeMax capital letters.
func validTechCode(code string) bool {
	if len(code) < 2 || len(code) > TechCodeMax {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
