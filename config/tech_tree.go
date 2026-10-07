package config

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// The tech tree's shape: the lanes techs sit in, the kind each tech is, the
// letterhead and emblem each one carries, and the rules its prerequisite
// graph keeps (TechTreeProblems).

// Tech lane keys. A lane is a column of the tree, about one of the game's
// systems.
const (
	LaneFaith       = "faith"       // faith, culture, morale, festivals, monuments, harbingers
	LaneKnowledge   = "knowledge"   // knowledge buildings, research time
	LaneTrade       = "trade"       // market, trade routes, harbors, expeditions, civilizations, deals
	LaneAgriculture = "agriculture" // food, housing, population
	LaneCraft       = "craft"       // construction, build cost and time, storage, game speed
	LaneMaterials   = "materials"   // wood, stone, metal and fuel
	LaneMilitary    = "military"    // military power, soldiers, campaigns, raids
	LaneEnergy      = "energy"      // coal, electricity, plasma
	LaneSpace       = "space"       // flight, rockets, orbit, the stars
	LaneComputing   = "computing"   // data, crypto
)

// TechTreeVersion is the tech tree's rules as a save knows them. It goes up
// when the tree starts locking something it did not lock before, and a save
// written under an older version (or before the tree: 0) then gets one age
// of grace from the new locks (the game's GameSave.TreeVersion).
//
//	1: each age's wonder needs its keystone tech (BuildingDef.RequiredTech).
//	2: commands wait for a tech (FeatureLocks).
//	3: Stonehenge needs its keystone (Calendar), four more commands wait for
//	   a tech that the tree now holds (trade routes, expeditions past the
//	   Scout Party, diplomacy, festivals), and seven buildings of the Stone
//	   and Iron Eras wait for one.
const TechTreeVersion = 3

// TechLaneDef is one lane of the tech tree.
type TechLaneDef struct {
	Key  string
	Name string
	// Emblem is the lane's glyph, one cell wide. A tech with no emblem of
	// its own carries it.
	Emblem string
	// Hue is the lane's identity colour on the tree, as "#rrggbb". The
	// panels pass it through the theme's contrast rule before drawing.
	Hue string
}

// TechLanes returns the lanes in the order the tree draws them, left to
// right.
func TechLanes() []TechLaneDef {
	return []TechLaneDef{
		{Key: LaneFaith, Name: "Faith & Culture", Emblem: "Ω", Hue: "#c9a0dc"},
		{Key: LaneKnowledge, Name: "Knowledge", Emblem: "§", Hue: "#6fb3e0"},
		{Key: LaneTrade, Name: "Trade & Exploration", Emblem: "$", Hue: "#e0b84f"},
		{Key: LaneAgriculture, Name: "Agriculture", Emblem: "♣", Hue: "#7fc97f"},
		{Key: LaneCraft, Name: "Craft & Engineering", Emblem: "⚒", Hue: "#d9925a"},
		{Key: LaneMaterials, Name: "Materials", Emblem: "♦", Hue: "#b0b7c3"},
		{Key: LaneMilitary, Name: "Military", Emblem: "†", Hue: "#e06c6c"},
		{Key: LaneEnergy, Name: "Energy", Emblem: "☼", Hue: "#f2d94e"},
		{Key: LaneSpace, Name: "Flight & Space", Emblem: "↑", Hue: "#8f9cf0"},
		{Key: LaneComputing, Name: "Computing", Emblem: "λ", Hue: "#4fd1c5"},
	}
}

// TechLaneByKey returns a map of lane key -> TechLaneDef.
func TechLaneByKey() map[string]TechLaneDef {
	m := make(map[string]TechLaneDef)
	for _, l := range TechLanes() {
		m[l.Key] = l
	}
	return m
}

// TechCodeMax is the most letters a tech's code may have: the width of the
// narrowest letterhead the tree draws.
const TechCodeMax = 5

// TechCodeFor makes a tech's code from its name, for a tech that sets none:
// the first word that is not an article, letters only, in capitals, cut to
// TechCodeMax. "Printing Press" is PRINT, "The Wheel" is WHEEL, "Zero-G
// Manufacturing" is ZERO. It reads the name alone, so adding a tech never
// changes another's code; two names that give the same code fail the tree's
// check, and one of them then sets its own.
func TechCodeFor(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == ' ' || r == '-' })
	for _, w := range words {
		switch strings.ToLower(w) {
		case "the", "a", "an", "of":
			continue
		}
		var code []rune
		for _, r := range strings.ToUpper(w) {
			if r >= 'A' && r <= 'Z' && len(code) < TechCodeMax {
				code = append(code, r)
			}
		}
		if len(code) > 0 {
			return string(code)
		}
	}
	return ""
}

// fillTechArt gives every tech that left them out a code (from its name) and
// an emblem (its lane's glyph).
func fillTechArt(techs []TechDef) []TechDef {
	lanes := TechLaneByKey()
	for i := range techs {
		t := &techs[i]
		if t.Code == "" {
			t.Code = TechCodeFor(t.Name)
		}
		if t.Emblem == "" {
			t.Emblem = lanes[t.Lane].Emblem
		}
	}
	return techs
}

// TechKind is a tech's place in the tree's structure.
type TechKind string

const (
	// TechKeystone is a tech a wonder requires: the wonder cannot be built
	// without it, and the next age needs the wonder.
	TechKeystone TechKind = "keystone"
	// TechSpine is a tech a keystone stands on: one of its prerequisites,
	// all the way down. Keystones and the spine are the only techs a run
	// has to research.
	TechSpine TechKind = "spine"
	// TechCapstone is a tech flagged as the end of its lane for an era.
	TechCapstone TechKind = "capstone"
	// TechOptional is every other tech.
	TechOptional TechKind = "optional"
)

// TechKinds works out every tech's kind. Nothing but the capstone flag is
// written by hand: a keystone is a tech some wonder names as its
// RequiredTech, and the spine is everything a keystone needs through
// Prerequisites, all the way down. An either-or group (AnyOf) puts nothing
// on the spine: either member will do, so neither is required.
//
// A capstone that a keystone also stands on reads as spine here, and
// TechTreeProblems reports it.
func TechKinds(techs []TechDef, buildings []BuildingDef) map[string]TechKind {
	byKey := make(map[string]TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	kinds := make(map[string]TechKind, len(techs))
	for _, t := range techs {
		kinds[t.Key] = TechOptional
		if t.Capstone {
			kinds[t.Key] = TechCapstone
		}
	}
	var onSpine func(key string)
	onSpine = func(key string) {
		t, ok := byKey[key]
		if !ok || kinds[key] == TechSpine || kinds[key] == TechKeystone {
			return
		}
		kinds[key] = TechSpine
		for _, p := range t.Prerequisites {
			onSpine(p)
		}
	}
	for _, key := range keystoneTechs(buildings) {
		if _, ok := byKey[key]; !ok {
			continue
		}
		onSpine(key)
		kinds[key] = TechKeystone
	}
	return kinds
}

// keystoneTechs lists the techs wonders require, in building order, each
// once.
func keystoneTechs(buildings []BuildingDef) []string {
	var out []string
	seen := map[string]bool{}
	for _, b := range buildings {
		if b.Category == "wonder" && b.RequiredTech != "" && !seen[b.RequiredTech] {
			seen[b.RequiredTech] = true
			out = append(out, b.RequiredTech)
		}
	}
	return out
}

// PrereqsMet reports whether the tech's prerequisites are met: have says
// yes for every key in Prerequisites, and for at least one key of AnyOf when
// the tech has such a group.
func (t TechDef) PrereqsMet(have func(key string) bool) bool {
	return t.MissingPrereq(have) == "" && t.AnyOfMet(have)
}

// MissingPrereq returns the first key of Prerequisites have says no to, or
// "" when it says yes to all of them.
func (t TechDef) MissingPrereq(have func(key string) bool) string {
	for _, p := range t.Prerequisites {
		if !have(p) {
			return p
		}
	}
	return ""
}

// AnyOfMet reports whether the tech's either-or group is satisfied: it has
// none, or have says yes to one of its keys.
func (t TechDef) AnyOfMet(have func(key string) bool) bool {
	if len(t.AnyOf) == 0 {
		return true
	}
	for _, p := range t.AnyOf {
		if have(p) {
			return true
		}
	}
	return false
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

	byKey := make(map[string]TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	ageAt := make(map[string]int, len(ages))
	for i, a := range ages {
		ageAt[a] = i
	}
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
