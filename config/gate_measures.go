package config

import (
	"os"
	"sort"
	"strings"
)

// The measured gate (experimental). Today a gate asks for building counts and
// the age's wonder. The measured rule, picked at start-up with
// AGEFORGE_GATES=v2, keeps the wonder and asks for the same four pillars in
// other terms:
//
//   - People: the town's population reaching the population the reference
//     town has when it meets today's gate, whatever mix of homes gets it there;
//   - Industry: the income of the age's main materials together reaching the
//     reference town's income of them at that moment, whatever mix of makers
//     makes it;
//   - Knowledge: the age's keystone tech and one more key tech of the age;
//   - Monument: the wonder, banked and built, as before.
//
// The reference town is smoke/reference.go's ordinary player. package smoke
// reads it and writes the numbers into gate_measures_table.go (go test ./smoke
// -run TestGateMeasures -update-gates); a test holds the table to the model, so
// nothing here is a number written by hand. This file holds the two rules that
// choose what is measured, and the table's type.

// GatesEnv is the environment variable that picks the gate rule.
const GatesEnv = "AGEFORGE_GATES"

// gatesV2 is read once, when the program starts, and never again: a game does
// not change its gate while it runs.
var gatesV2 = ParseGateRule(os.Getenv(GatesEnv))

// ParseGateRule reports whether a value of GatesEnv picks the measured gate.
// Anything but "v2" (any case) is today's rule.
func ParseGateRule(value string) bool { return strings.EqualFold(strings.TrimSpace(value), "v2") }

// GatesV2 reports whether the program started with the measured gate. It is
// the value GatesEnv had when the program started.
func GatesV2() bool { return gatesV2 }

// GateMainMaterialsMax is how many materials an age's Industry measure adds up.
const GateMainMaterialsMax = 2

// GateMeasure is what the measured gate asks of a town leaving Age.
type GateMeasure struct {
	// Age is the age the town leaves; the gate leads into the next one.
	Age string
	// People is the population the town must reach, at least.
	People int
	// Materials are the age's main materials, as resource keys, in the order
	// they are named on screen. Income is the most a tick of all of them
	// together the town must reach, at least.
	Materials []string
	Income    float64
	// KeyTech is the tech asked for beside the age's keystone; "" when the
	// age has no tech to give.
	KeyTech string
}

// GateMeasures returns the measure of each age that has a gate, in age order.
// It is a copy of the table in gate_measures_table.go.
func GateMeasures() []GateMeasure {
	out := make([]GateMeasure, len(gateMeasureTable))
	for i, m := range gateMeasureTable {
		m.Materials = append([]string(nil), m.Materials...)
		out[i] = m
	}
	return out
}

// GateMainMaterials returns an age's main materials: of the resources that
// the age's makers produce, the (up to) GateMainMaterialsMax that the prices
// of the age's buildings use most. A price counts in the age's price units,
// the amount over the resource's level, so a resource that is dear to make
// weighs the same as one that is cheap. Buildings of the age are all its
// buildings but the wonder and the monuments (the Monument pillar's own); a
// maker is one with a production effect. Ties go to the resource key. levels
// are the age's price levels (PriceLevelsByAge).
func GateMainMaterials(age string, buildings []BuildingDef, levels map[string]float64) []string {
	made := map[string]bool{}
	weight := map[string]float64{}
	for _, d := range buildings {
		if d.RequiredAge != age || d.Category == "wonder" || d.Category == "monument" {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" {
				made[e.Target] = true
			}
		}
		for res, cost := range d.BaseCost {
			if l := levels[res]; l > 0 {
				weight[res] += cost / l
			}
		}
	}
	var mats []string
	for res := range made {
		if weight[res] > 0 {
			mats = append(mats, res)
		}
	}
	sort.Slice(mats, func(i, j int) bool {
		if weight[mats[i]] != weight[mats[j]] {
			return weight[mats[i]] > weight[mats[j]]
		}
		return mats[i] < mats[j]
	})
	if len(mats) > GateMainMaterialsMax {
		mats = mats[:GateMainMaterialsMax]
	}
	return mats
}

// GateKeyTech returns the key tech of an age, the one asked for beside its
// keystone: of the age's techs that the keystone does not already stand on,
// the one that most of the age's buildings need (a building names the tech it
// waits for). When no building needs any of them, the one most other techs
// are built on. Ties go to the cheaper tech, then to the key. "" when the age
// has no such tech. wonder is the age's wonder, whose tech is the keystone.
func GateKeyTech(age, wonder string, buildings []BuildingDef, techs []TechDef) string {
	byKey := make(map[string]TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	keystone := ""
	needs := map[string]int{}
	for _, d := range buildings {
		if d.Key == wonder {
			keystone = d.RequiredTech
		}
		if d.RequiredAge == age && d.RequiredTech != "" {
			needs[d.RequiredTech]++
		}
	}
	// The keystone's own prerequisites come with it.
	under := map[string]bool{}
	var mark func(key string)
	mark = func(key string) {
		for _, p := range byKey[key].Prerequisites {
			if !under[p] {
				under[p] = true
				mark(p)
			}
		}
	}
	if keystone != "" {
		mark(keystone)
	}
	leaned := map[string]int{}
	for _, t := range techs {
		for _, p := range t.Prerequisites {
			leaned[p]++
		}
		for _, p := range t.AnyOf {
			leaned[p]++
		}
	}
	best, found := TechDef{}, false
	better := func(a, b TechDef) bool {
		if needs[a.Key] != needs[b.Key] {
			return needs[a.Key] > needs[b.Key]
		}
		if leaned[a.Key] != leaned[b.Key] {
			return leaned[a.Key] > leaned[b.Key]
		}
		if a.Cost != b.Cost {
			return a.Cost < b.Cost
		}
		return a.Key < b.Key
	}
	for _, t := range techs {
		if t.Age != age || t.Key == keystone || under[t.Key] {
			continue
		}
		if !found || better(t, best) {
			best, found = t, true
		}
	}
	return best.Key
}
