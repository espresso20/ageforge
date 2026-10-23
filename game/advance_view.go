package game

import (
	"fmt"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// advance_view.go is the gate into the next age as the Advance panel shows
// it: four pillars (Knowledge, People, Industry, Monument), each a list of
// requirements with how far along it is, the march of ages with a calendar
// year, what the next age brings, and what else stands in the way.
//
// It is a read of a snapshot and decides nothing. What the gate asks for is
// still ProgressManager.CheckAdvancement: the building counts its next age
// lists and the leaving age's wonder. The pillars sort exactly those, plus
// the wonder's keystone tech (which the wonder cannot be built without), so
// "all four stand" and "CheckAdvancement says ready" are the same fact
// (advance_view_test.go holds the two together for every age).

// charterHint ends the advance command's refusals: where to see the whole
// gate.
const charterHint = "Type 'charter' to see everything the advance needs."

// GateRowKind says how a row is counted and written.
type GateRowKind int

const (
	// GateRowCount is a building: Have of Need copies.
	GateRowCount GateRowKind = iota
	// GateRowShare is a share in whole percent (Have of Need = 100): the
	// wonder's bank.
	GateRowShare
	// GateRowDone is one thing, done or not (Have 0 or 1 of Need 1): the
	// keystone tech, the wonder built.
	GateRowDone
)

// GateRow is one requirement of a pillar.
type GateRow struct {
	Kind GateRowKind
	// Key is the building or tech the row is about.
	Key   string
	Label string
	// Have and Need count the row in its kind's unit.
	Have, Need int
	Met        bool
	// Progress is how far along the row is, from 0 to 1 (1 once met).
	Progress float64
	// Note is a few plain words on where the row stands, for the rows that
	// are more than a count ("researching 62%", "needs Calendar first").
	Note string
}

// GatePillar is one of the four groups of requirements.
type GatePillar struct {
	// Key is a config pillar key (config.PillarKnowledge and the rest).
	Key  string
	Name string
	Rows []GateRow
	// Facts are lines of context about the pillar that the gate does not
	// ask for: how many scholars work, how many people live in the town.
	Facts []string
	// Met is true when every row is met, and for a pillar with no rows in
	// this age. Progress is the mean of the rows' progress (1 with none).
	Met      bool
	Progress float64
}

// MarchStop is one age on the march of ages.
type MarchStop struct {
	Key string
	// Name and Year are "" for an age the player cannot see named yet
	// (SightOf); the panel shows such an age as an unnamed stop.
	Name string
	Year string
}

// AdvanceBrings is what the next age adds, short.
type AdvanceBrings struct {
	// Buildings is how many kinds of building the next age unlocks, its
	// wonder not counted.
	Buildings int
	// Resources are the names of the resources it unlocks.
	Resources []string
	// People are the age's own names for its four kinds of person.
	People []string
	// Wonder is the next age's wonder.
	Wonder string
}

// AdvanceView is the next-age gate as four pillars.
type AdvanceView struct {
	Tick           int
	Age, AgeName   string
	Next, NextName string
	// Final is true in the last age: there is no gate.
	Final bool
	// Pillars are the four pillars in display order; nil in the final age.
	Pillars []GatePillar
	// Stand counts the pillars that are met.
	Stand int
	// Progress is the mean progress of the pillars that ask for something,
	// 0 to 1 (0 in the final age).
	Progress float64
	// AllStand is true when all four pillars are met. Ready also needs
	// nothing else in the way: it is what Enter on the panel would do.
	AllStand, Ready bool
	// Blockers stop the advance outright; Notices are things the player
	// should know before they advance.
	Blockers []string
	Notices  []string
	// March lists every age in order; Here is this age's place in it.
	March []MarchStop
	Here  int
	// YearFrom and YearTo are the start years of this age and the next.
	// Share is how far the calendar has come between them: Progress,
	// unless HoldCalendar has held it at a higher mark. Year and YearText
	// are the calendar year at Share.
	YearFrom, YearTo int
	Share            float64
	Year             int
	YearText         string
	// Brings is what the next age adds (zero in the final age).
	Brings AdvanceBrings
}

// AdvanceView works out the gate into the next age from the snapshot.
func (st GameState) AdvanceView() AdvanceView {
	set := st.Ruleset()
	keys := set.AgeKeys()
	here, _ := set.Index(st.Age)
	next := set.Next(st.Age)
	years := set.AgeYears()
	sight := SightOf(&st)

	v := AdvanceView{
		Tick: st.Tick, Age: st.Age, AgeName: set.Name(rules.KindAge, st.Age),
		Next: next, Final: next == "", Here: here,
	}
	if next != "" {
		v.NextName = set.Name(rules.KindAge, next)
	}
	for i, k := range keys {
		stop := MarchStop{Key: k}
		if sight.Age(k) {
			stop.Name = set.Name(rules.KindAge, k)
			stop.Year = ageYearText(set, k, years[i])
		}
		v.March = append(v.March, stop)
	}
	if here < len(years) {
		v.YearFrom, v.YearTo = years[here], years[here]
	}
	if v.Final {
		v.Blockers = append(v.Blockers, "This is the final age. There is no age beyond it.")
		v.Year, v.YearText = v.YearFrom, ageYearText(set, st.Age, v.YearFrom)
		return v
	}
	v.YearTo = years[here+1]

	v.Pillars = st.gatePillars(set, next)
	// The calendar moves with the pillars that have something to do: one
	// with nothing needed this age stands from the start, but it is no
	// progress, or an age would begin a quarter of the way through.
	asked := 0
	for _, p := range v.Pillars {
		if p.Met {
			v.Stand++
		}
		if len(p.Rows) > 0 {
			v.Progress += p.Progress
			asked++
		}
	}
	if asked > 0 {
		v.Progress = v.Progress / float64(asked)
	}
	v.AllStand = v.Stand == len(v.Pillars)
	v.setCalendar(v.Progress)

	if st.PendingCatastrophe != "" {
		name, _ := set.Catastrophe(st.PendingCatastrophe)
		v.Blockers = append(v.Blockers, fmt.Sprintf("%s is upon you. Type 'catastrophe' to choose Endure or Succumb before advancing.", name))
	}
	if h := st.Harbinger; h != nil && !h.LastPassage {
		// The same test the advance makes (fateBeforeAdvance), on what the
		// figure has said: a doom promised for this age, or any open doom
		// when the advance leaves the era. Nothing hidden is read.
		if h.When == WhenThisAge || set.EraOf(next) != set.EraOf(st.Age) {
			v.Notices = append(v.Notices, fmt.Sprintf("%s stands in your way. Type 'harbinger' to answer, or advance to meet it.", h.Name))
		}
	}
	v.Ready = v.AllStand && st.PendingCatastrophe == ""
	v.Brings = advanceBrings(set, next)
	return v
}

// setCalendar puts the calendar at share of the way to the next age.
func (v *AdvanceView) setCalendar(share float64) {
	v.Share = share
	v.Year = config.YearBetween(v.YearFrom, v.YearTo, share)
	v.YearText = config.YearLabel(v.Year)
}

// ageYearText is the start of age as the march writes it: a number of years,
// or the age's own word.
func ageYearText(set *rules.Set, age string, year int) string {
	if w := set.AgeYearLabel(age); w != "" {
		return w
	}
	return config.YearLabel(year)
}

// gatePillars builds the four pillars of the gate into next.
func (st GameState) gatePillars(set *rules.Set, next string) []GatePillar {
	var rows = map[string][]GateRow{}

	// The age's wonder: its keystone leads the Knowledge pillar; the bank
	// and the building of it are the Monument's.
	wonderKey := set.Wonder(st.Age)
	wonder := st.Buildings[wonderKey]
	built := wonderKey != "" && wonder.Count >= 1
	if tech := set.Keystone(st.Age); tech != "" {
		rows[config.PillarKnowledge] = append(rows[config.PillarKnowledge], st.keystoneRow(set, tech, built, wonder))
	}

	// The buildings the gate asks for.
	if age, ok := set.Age(next); ok {
		for key, need := range age.BuildingReqs {
			def, known := set.Building(key)
			pillar := config.PillarIndustry
			if known {
				pillar = config.GatePillar(def)
			}
			// A gate counts the building and the later rungs of its line.
			bs := st.Buildings[key]
			have := max(bs.Count, bs.HeldWithLater)
			progress := 1.0
			if have < need {
				progress = float64(have) / float64(need)
			}
			rows[pillar] = append(rows[pillar], GateRow{
				Kind: GateRowCount, Key: key, Label: set.Name(rules.KindBuilding, key),
				Have: have, Need: need, Met: have >= need, Progress: progress,
			})
		}
	}

	if wonderKey != "" {
		rows[config.PillarMonument] = st.monumentRows(set, wonderKey, built, wonder)
	}

	out := make([]GatePillar, 0, 4)
	for _, key := range config.GatePillars() {
		list := rows[key]
		sort.SliceStable(list, func(i, j int) bool {
			// The keystone and the wonder's rows keep their place; buildings
			// go by name.
			if list[i].Kind != GateRowCount || list[j].Kind != GateRowCount {
				return list[i].Kind != GateRowCount && list[j].Kind == GateRowCount
			}
			if list[i].Label != list[j].Label {
				return list[i].Label < list[j].Label
			}
			return list[i].Key < list[j].Key
		})
		p := GatePillar{Key: key, Name: config.PillarName(key), Rows: list, Met: true, Progress: 1}
		if len(list) > 0 {
			sum := 0.0
			for _, r := range list {
				p.Met = p.Met && r.Met
				sum += r.Progress
			}
			p.Progress = sum / float64(len(list))
		}
		p.Facts = st.pillarFacts(key)
		out = append(out, p)
	}
	return out
}

// keystoneRow is the age's keystone tech as a row: met when researched, and
// met without it once the wonder stands or the age is exempt from the lock (a
// save from before the lock), since then nothing waits for it.
func (st GameState) keystoneRow(set *rules.Set, tech string, built bool, wonder BuildingState) GateRow {
	ts := st.Research.Techs[tech]
	name := set.Name(rules.KindTech, tech)
	waived := !ts.Researched && wonder.NeedsTech == ""
	r := GateRow{Kind: GateRowDone, Key: tech, Label: "Keystone: " + name, Need: 1}
	switch {
	case ts.Researched:
		r.Have, r.Met, r.Progress, r.Note = 1, true, 1, "researched"
	case built || waived:
		r.Have, r.Met, r.Progress, r.Note = 1, true, 1, "not needed"
	case st.Research.CurrentTech == tech:
		if st.Research.Price > 0 {
			r.Progress = clampShare(st.Research.Progress / st.Research.Price)
		}
		r.Note = fmt.Sprintf("researching, %d%%", int(r.Progress*100))
	default:
		r.Note = "not started"
		for _, q := range st.Research.Queue {
			if q == tech {
				r.Note = "queued"
			}
		}
	}
	return r
}

// monumentRows are the wonder's bank and the wonder built.
func (st GameState) monumentRows(set *rules.Set, wonderKey string, built bool, wonder BuildingState) []GateRow {
	name := set.Name(rules.KindBuilding, wonderKey)
	def, _ := set.Building(wonderKey)

	share := 0.0
	if built || wonder.WonderBankFull {
		share = 1
	} else {
		share = WonderBankShare(set, st.Age, def.BaseCost, wonder.WonderBank)
	}
	pct := int(share * 100)
	if share < 1 {
		pct = min(pct, 99)
	}
	bank := GateRow{Kind: GateRowShare, Key: wonderKey, Label: name + " bank", Have: pct, Need: 100,
		Met: share >= 1, Progress: share}
	if bank.Met {
		bank.Have, bank.Note = 100, "full"
	}

	build := GateRow{Kind: GateRowDone, Key: wonderKey, Label: name, Need: 1}
	switch {
	case built:
		build.Have, build.Met, build.Progress, build.Note = 1, true, 1, "built"
	default:
		for _, q := range st.BuildQueue {
			if q.Name == name && q.TotalTicks > 0 {
				build.Progress = clampShare(1 - float64(q.TicksLeft)/float64(q.TotalTicks))
				build.Note = fmt.Sprintf("building, %d%%", int(build.Progress*100))
			}
		}
		if build.Note == "" {
			switch {
			case wonder.NeedsTech != "":
				build.Note = "needs " + set.Name(rules.KindTech, wonder.NeedsTech) + " first"
			case !bank.Met:
				build.Note = "bank it first"
			default:
				build.Note = "ready to build"
			}
		}
	}
	return []GateRow{bank, build}
}

// pillarFacts are the lines of context a pillar carries, in the age's own
// words for its people.
func (st GameState) pillarFacts(pillar string) []string {
	ws := st.Workers
	switch pillar {
	case config.PillarKnowledge:
		return []string{fmt.Sprintf("%s at work: %d", capFirst(st.Term(config.TermScholars).Many), ws.Scholars)}
	case config.PillarPeople:
		return []string{
			fmt.Sprintf("%s: %d", capFirst(st.Term(config.TermWorkers).Many), ws.Working()),
			fmt.Sprintf("%s: %d", capFirst(st.Term(config.TermReserve).Many), ws.Reserve()),
			fmt.Sprintf("Homes: %d", ws.MaxPop),
		}
	}
	return nil
}

// advanceBrings is what age next adds.
func advanceBrings(set *rules.Set, next string) AdvanceBrings {
	var b AdvanceBrings
	age, ok := set.Age(next)
	if !ok {
		return b
	}
	for _, key := range age.UnlockBuildings {
		if def, ok := set.Building(key); ok && def.Category == "wonder" {
			continue
		}
		b.Buildings++
	}
	for _, key := range age.UnlockResources {
		b.Resources = append(b.Resources, set.Name(rules.KindResource, key))
	}
	for _, term := range config.PeopleTerms() {
		b.People = append(b.People, capFirst(config.Term(next, term).Many))
	}
	if w := set.Wonder(next); w != "" {
		b.Wonder = set.Name(rules.KindBuilding, w)
	}
	return b
}

func clampShare(x float64) float64 {
	switch {
	case x != x || x < 0:
		return 0
	case x > 1:
		return 1
	}
	return x
}

// calendarMark is the furthest the town has been shown on its way to the
// next age: the age and tick it was seen at, and the share.
type calendarMark struct {
	age   string
	tick  int
	share float64
}

// resetCalendar forgets the mark: a new age, a loaded save.
func (ge *GameEngine) resetCalendar() {
	ge.calendarMu.Lock()
	ge.calendar = calendarMark{}
	ge.calendarMu.Unlock()
}

// HoldCalendar keeps the calendar from running backwards: a town that sells a
// building or loses some to a catastrophe has fewer requirements met than it
// had, but the year it was shown does not go back within the age. It raises
// the view's calendar to the furthest share the panel has shown for this age
// (and records this view's share if it is further). The pillars themselves
// are not touched: they show how the gate stands. The mark is derived from
// what was shown, held in memory only, and starts over in a new age, a new
// run and a loaded game.
func (ge *GameEngine) HoldCalendar(v AdvanceView) AdvanceView {
	if ge == nil || v.Final {
		return v
	}
	ge.calendarMu.Lock()
	defer ge.calendarMu.Unlock()
	m := ge.calendar
	if m.age != v.Age || v.Tick < m.tick {
		m = calendarMark{age: v.Age}
	}
	m.tick = v.Tick
	m.share = max(m.share, v.Share)
	ge.calendar = m
	v.setCalendar(m.share)
	return v
}
