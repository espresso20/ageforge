package config

import (
	"fmt"
	"strings"
)

// tally.go is the lifetime tallies: one coin each, struck one level finer at
// each rung the account's count reaches. The ledger tallies are listed here;
// the ones made from the game's tables (lineages, resources) are built by
// rules.Set.Tallies from the figures below and the tables.

// CoinTally is the coin type of a tally. Its stack is named by the tally's
// key and holds one copy, whose Figure is the count when the level last rose.
const CoinTally = "tally"

// TallyDef is one lifetime tally.
type TallyDef struct {
	Key, Name string
	// Counter is the account counter that is counted. With Subjects, the
	// count is how many of Counter+"."+subject have reached Min; with Peak,
	// the counter is the most a run has had at once. With no Counter, each
	// subject is a whole counter name and Mins its own minimum.
	Counter  string
	Subjects []string
	Min      float64
	Mins     []float64
	Peak     bool
	Rungs    []float64
}

// MinOf is the count subject i must reach to be counted.
func (t TallyDef) MinOf(i int) float64 {
	if i < len(t.Mins) {
		return t.Mins[i]
	}
	return t.Min
}

// Level is the level a count reaches, 0 below the first rung: the rungs
// reached, spread over the six levels (halves up).
func (t TallyDef) Level(v float64) int {
	rungs, r := t.Rungs, 0
	for _, t := range rungs {
		if v >= t {
			r++
		}
	}
	if r == 0 {
		return 0
	}
	if len(rungs) < 2 {
		return 1
	}
	return 1 + ((r-1)*2*(CoinLevels-1)/(len(rungs)-1)+1)/2
}

// Tallies is a list of tallies.
type Tallies []TallyDef

// TallyLedgers is the tallies over the account's own counters, in the
// ledger's order.
func TallyLedgers() Tallies {
	t := func(key, name, counter string, rungs ...float64) TallyDef {
		return TallyDef{Key: "ledger." + key, Name: name, Counter: counter, Rungs: rungs}
	}
	return Tallies{
		t("prestiges", "Prestiges", BadgeEvPrestige, 1, 3, 10, 25),
		t("endured", "Dooms endured", BadgeEvEndured, 1, 5, 15, 30),
		t("succumbed", "Dooms succumbed to", BadgeEvSuccumbed, 1, 5, 15),
		t("threads", "Warnings resolved", BadgeEvHarbingerResolved, 1, 10, 40, 100),
		t("appeased", "Appeasements", BadgeEvAppeased, 1, 10, 50),
		t("braced", "Braces", BadgeEvBraced, 1, 10, 50),
		t("invited", "Invitations", BadgeEvInvited, 1, 5, 20),
		t("exposed", "False prophets", BadgeEvHarbingerResolved+".discredited", 1, 3, 5),
		t("wonders", "Wonders raised", BadgeEvWonderRaised, 2, 24, 300),
		t("techs", "Techs researched", BadgeEvResearchDone, 5, 77, 1200),
		t("milestones", "Milestones", BadgeEvMilestone, 5, 70, 1000),
		t("chains", "Milestone chains", BadgeEvChain, 1, 10, 30),
		t("expeditions", "Expeditions", BadgeEvExpedition, 1, 20, 100, 500),
		t("deals", "Faction deals", BadgeEvDeal, 1, 10, 100),
		t("raids", "Raids blunted", BadgeEvRaidBlunted, 1, 22, 500),
		t("returns", "Returns", BadgeEvReturned, 1, 25, 100),
		t("plan", "Plan items started", BadgeEvPlanStarted, 29, 500, 5000),
		t("upgrades", "Buildings upgraded", BadgeEvBuildingUpgraded, 20, 630, 20000),
		t("festivals", "Festivals", BadgeEvFestival, 1, 10, 50),
		t("blackmarket", "Black market deals", BadgeEvBlackMarket, 1, 10, 50),
		t("memories", "Ancient Memories", BadgeEvMemoryAccepted, 1, 5, 12),
		t("hardships", "Hard times weathered", BadgeEvEraEvent+".bad_challenging", 1, 10, 30),
	}
}

// The tallies that count a table's subjects: how many civilizations were met
// (Counter + "." + civ at least once), how many first twelve harbingers were
// appeased.
func TallyCivs(civs, harbingers []string) Tallies {
	var parts []string
	var mins []float64
	for _, c := range civs {
		for _, p := range []struct {
			ev  string
			min float64
		}{{BadgeEvCivMet, 1}, {BadgeEvCivAllied, 1}, {BadgeEvDeal, 5}, {BadgeEvWarEnded, 1}} {
			parts, mins = append(parts, p.ev+"."+c), append(mins, p.min)
		}
	}
	return Tallies{
		{Key: "ledger.civs", Name: "Civilization firsts", Subjects: parts, Mins: mins, Rungs: []float64{1, 5, 11, 22, 33, 44}},
		{Key: "ledger.heeded", Name: "Harbingers heeded", Counter: BadgeEvAppeased, Subjects: harbingers, Min: 1, Rungs: []float64{1, 3, 6, 9, 12}},
	}
}

// TallyHeeded is the harbingers whose appeasement counts as heeding.
var TallyHeeded = []string{"wild_man", "hermit", "soothsayer", "desert_prophet", "oracle", "town_crier",
	"court_astrologer", "pamphleteer", "newsboy", "soapbox_doomsayer", "telegraph", "civil_defence_broadcast"}

// TallyLineageRungs is each lineage's rungs: copies built across all runs.
var TallyLineageRungs = map[string][]float64{
	"culture_arts": {310, 770, 1900, 4800, 12_000}, "energy": {200, 510, 1300, 3200, 8000},
	"engineering": {310, 770, 1900, 4800, 12_000}, "faith": {380, 960, 2400, 6000, 15_000},
	"food": {380, 960, 2400, 6000, 15_000}, "geological_extraction": {410, 1000, 2600, 6400, 16_000},
	"hacker": {140, 350, 860, 2200, 5400}, "harbor": {79, 200, 500, 1200, 3100},
	"housing": {610, 1500, 3800, 9600, 24_000}, "knowledge": {130, 330, 830, 2100, 5200},
	"metallurgy": {280, 700, 1800, 4400, 11_000}, "military": {380, 960, 2400, 6000, 15_000},
	"organic_extraction": {380, 960, 2400, 6000, 15_000}, "storage": {110, 280, 690, 1700, 4300},
	"trade": {280, 700, 1800, 4400, 11_000},
}

// TallyDomainRungs is each worker domain's rungs: workers at work at once.
var TallyDomainRungs = map[string][]float64{
	"food": {270, 450}, "lumber": {270, 468}, "masonry": {270, 477}, "knowledge": {64, 110},
	"faith": {160, 265}, "military": {11900, 19600}, "trade": {170, 237}, "engineering": {180, 305},
	"metallurgy": {150, 248}, "energy": {76, 137}, "hacker": {260, 520},
}

// Gift reads a gift's badge field ("ledger.deals.3": a tally's key and the
// rung that gives it), and says the level that rung is struck at.
func (ts Tallies) Gift(field string) (key string, level int, ok bool) {
	i := strings.LastIndexByte(field, '.')
	if i < 0 {
		return
	}
	k, n := field[:i], 0
	if _, err := fmt.Sscan(field[i+1:], &n); err != nil {
		return
	}
	for _, t := range ts {
		if t.Key == k && n >= 1 && n <= len(t.Rungs) {
			return t.Key, t.Level(t.Rungs[n-1]), true
		}
	}
	return
}
