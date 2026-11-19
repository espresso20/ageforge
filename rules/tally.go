package rules

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/vibe"
)

// tally.go builds the lifetime tallies of a ruleset: the ledger's own, and
// the ones made from the game's tables.

// Tallies is every tally in the ledger's order: lineages, domains, resources,
// then the tallies over the account's counters.
func (s *Set) Tallies() config.Tallies {
	var out config.Tallies
	for _, sub := range s.badgeSubjects(config.BadgeSourceLineages) {
		if r, ok := config.TallyLineageRungs[sub.key]; ok {
			out = append(out, config.TallyDef{Key: "lineage." + sub.key, Name: sub.name + vibe.Line("rules.tally.Tallies.01", " builds"), Counter: config.BadgeEvBuiltLineage + "." + sub.key, Rungs: r})
		}
	}
	for _, sub := range s.badgeSubjects(config.BadgeSourceDomains) {
		if r, ok := config.TallyDomainRungs[sub.key]; ok {
			out = append(out, config.TallyDef{Key: "domain." + sub.key, Name: sub.name + vibe.Line("rules.tally.Tallies.02", " workers"), Counter: "staffed." + sub.key, Peak: true, Rungs: r})
		}
	}
	for _, sub := range s.badgeSubjects(config.BadgeSourceResources) {
		perRun, deep := s.RunProduction(sub.key)
		_, first := s.FirstProduction(sub.key)
		n := 4
		if deep {
			n--
		}
		if perRun > 0 && first > 0 {
			out = append(out, config.TallyDef{Key: "resource." + sub.key, Name: sub.name + vibe.Line("rules.tally.Tallies.03", " stores"), Counter: config.BadgeEvProduced + "." + sub.key,
				Rungs: evenLadder(first, perRun*config.BadgeMaxRuns, n)})
		}
	}
	var civs []string
	for _, sub := range s.badgeSubjects(config.BadgeSourceCivs) {
		civs = append(civs, sub.key)
	}
	out = append(out, config.TallyLedgers()...)
	return append(out, config.TallyCivs(civs, config.TallyHeeded)...)
}
