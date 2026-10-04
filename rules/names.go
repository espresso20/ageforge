package rules

import (
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Kind says what a key names, for Name.
type Kind uint8

const (
	KindAge Kind = iota
	KindBuilding
	KindTech
	KindResource
	KindCiv
	KindRoute
	KindPrestigeUpgrade
	numKinds
)

func (s *Set) buildNames() {
	for k := range s.names {
		s.names[k] = map[string]string{}
	}
	for _, d := range s.ages {
		s.names[KindAge][d.Key] = d.Name
	}
	for _, d := range s.buildings {
		s.names[KindBuilding][d.Key] = d.Name
	}
	for _, d := range s.techs {
		s.names[KindTech][d.Key] = d.Name
	}
	for _, d := range s.resources {
		s.names[KindResource][d.Key] = d.Name
	}
	for _, d := range s.factions {
		s.names[KindCiv][d.Key] = d.Name
	}
	for _, d := range s.routes {
		s.names[KindRoute][d.Key] = d.Name
	}
	for _, d := range s.upgrades {
		s.names[KindPrestigeUpgrade][d.Key] = d.Name
	}
}

// Name returns the display name of key ("Industrial Age", "Lumber Mill").
// A key the set has no name for reads as the key itself, spaced and
// capitalized: "iron_ore" becomes "Iron ore".
func (s *Set) Name(kind Kind, key string) string {
	if kind < numKinds {
		if n := s.names[kind][key]; n != "" {
			return n
		}
	}
	return textfmt.Capitalize(strings.ReplaceAll(key, "_", " "))
}

// The quantities Counts reports, as the docs word them.
const (
	CountAges        = "ages"
	CountBuildings   = "buildings"
	CountTechs       = "technologies"
	CountMilestones  = "milestones"
	CountChains      = "milestone chains"
	CountResources   = "resources"
	CountEras        = "epochs"
	CountDomains     = "worker domains"
	CountShopItems   = "prestige shop items"
	CountLineages    = "lineages"
	CountWonders     = "wonders"
	CountCivs        = "civilizations"
	CountRoutes      = "trade routes"
	CountLineageBld  = "lineage buildings"
	CountStorage     = "storage buildings"
	CountMonuments   = "cultural monuments"
	CountStandalone  = "standalone buildings"
	lineageStorage   = "storage"
	lineageWonder    = "wonder"
	lineageMonuments = "monument"
)

func (s *Set) buildCounts() {
	c := map[string]int{
		CountAges: len(s.ages), CountBuildings: len(s.buildings), CountTechs: len(s.techs),
		CountMilestones: len(s.milestones), CountChains: len(s.chains), CountResources: len(s.resources),
		CountEras: len(s.eras), CountDomains: len(s.domains), CountShopItems: len(s.shop),
		CountCivs: len(s.factions), CountRoutes: len(s.routes),
	}
	lineages := map[string]bool{}
	for _, d := range s.buildings {
		switch {
		case d.Category == "wonder":
			c[CountWonders]++
		case d.Category == "storage":
			c[CountStorage]++
		case d.Category == "monument":
			c[CountMonuments]++
		case d.LineageKey != "" && d.LineageKey != lineageStorage && d.LineageKey != lineageWonder && d.LineageKey != lineageMonuments:
			c[CountLineageBld]++
			lineages[d.LineageKey] = true
		default:
			c[CountStandalone]++
		}
	}
	c[CountLineages] = len(lineages)
	s.counts = c
}

// Counts returns the numbers the docs quote about the game: how many ages,
// buildings, technologies and so on the set defines, keyed by the Count
// constants. The map is the caller's.
func (s *Set) Counts() map[string]int {
	out := make(map[string]int, len(s.counts))
	for k, v := range s.counts {
		out[k] = v
	}
	return out
}
