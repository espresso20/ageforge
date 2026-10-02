package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// pluralCases holds every building name in config that ends in s, x, ch, sh
// or y, plus a few regular names, with the plural a player should read.
var pluralCases = []struct{ name, plural string }{
	// Regular.
	{"Hut", "Huts"},
	{"Workshop", "Workshops"},

	// Already plural or invariant.
	{"Barracks", "Barracks"},
	{"Agricultural Works", "Agricultural Works"},
	{"Coal Works", "Coal Works"},
	{"Cosmic Organic Works", "Cosmic Organic Works"},
	{"Field Works", "Field Works"},
	{"Marble Works", "Marble Works"},
	{"Quantum Metal Works", "Quantum Metal Works"},
	{"Steam Works", "Steam Works"},
	{"Uranium Processing Works", "Uranium Processing Works"},
	{"Ironworks", "Ironworks"},
	{"Colonial Steelworks", "Colonial Steelworks"},
	{"Integrated Steelworks", "Integrated Steelworks"},
	{"Standing Stones", "Standing Stones"},

	// "X of Y" pluralizes X.
	{"Monument of Ages", "Monuments of Ages"},

	// Singular, ending in ss or us.
	{"Fortress", "Fortresses"},
	{"Agritech Campus", "Agritech Campuses"},
	{"Research Campus", "Research Campuses"},
	{"Transcendent Nexus", "Transcendent Nexuses"},
	{"Warp Nexus", "Warp Nexuses"},

	// x, ch, sh.
	{"Agricultural Complex", "Agricultural Complexes"},
	{"Bunker Complex", "Bunker Complexes"},
	{"Launch Complex", "Launch Complexes"},
	{"Nano Drill Complex", "Nano Drill Complexes"},
	{"Neural Art Complex", "Neural Art Complexes"},
	{"Smart Complex", "Smart Complexes"},
	{"Megaplex", "Megaplexes"},
	{"Church", "Churches"},
	{"Stash", "Stashes"},

	// Consonant + y.
	{"Granary", "Granaries"},
	{"Academy", "Academies"},
	{"Military Academy", "Military Academies"},
	{"Reality Academy", "Reality Academies"},
	{"Foundry", "Foundries"},
	{"Aerospace Foundry", "Aerospace Foundries"},
	{"Augmentation Foundry", "Augmentation Foundries"},
	{"Nano Foundry", "Nano Foundries"},
	{"Dark Matter Refinery", "Dark Matter Refineries"},
	{"Oil Refinery", "Oil Refineries"},
	{"Orbital Refinery", "Orbital Refineries"},
	{"Petroleum Refinery", "Petroleum Refineries"},
	{"Smart Refinery", "Smart Refineries"},
	{"Deep Space Observatory", "Deep Space Observatories"},
	{"Dyson Assembly", "Dyson Assemblies"},
	{"Embassy", "Embassies"},
	{"Grand Embassy", "Grand Embassies"},
	{"Library", "Libraries"},
	{"Eternal Library", "Eternal Libraries"},
	{"Great Library", "Great Libraries"},
	{"Monastery Library", "Monastery Libraries"},
	{"Geographic Society", "Geographic Societies"},
	{"Harbor Authority", "Harbor Authorities"},
	{"Quarry", "Quarries"},
	{"Marble Quarry", "Marble Quarries"},
	{"Neon Sanctuary", "Neon Sanctuaries"},
	{"Orbital Sanctuary", "Orbital Sanctuaries"},
	{"Physics Laboratory", "Physics Laboratories"},
	{"Smithy", "Smithies"},
	{"Steam Colliery", "Steam Collieries"},
	{"Stellar Metallurgy", "Stellar Metallurgies"},
	{"University", "Universities"},
	{"Void Monastery", "Void Monasteries"},
	{"Zero G Gallery", "Zero G Galleries"},

	// Vowel + y.
	{"Hydroponic Bay", "Hydroponic Bays"},
	{"Orbital Data Relay", "Orbital Data Relays"},
	{"Tokamak Array", "Tokamak Arrays"},
	{"Microgrid Array", "Microgrid Arrays"},
	{"Quantum Battery Array", "Quantum Battery Arrays"},
	{"Solar Collector Array", "Solar Collector Arrays"},
}

func TestPluralName(t *testing.T) {
	for _, tc := range pluralCases {
		if got := pluralName(2, tc.name); got != tc.plural {
			t.Errorf("pluralName(2, %q) = %q, want %q", tc.name, got, tc.plural)
		}
		if got := pluralName(0, tc.name); got != tc.plural {
			t.Errorf("pluralName(0, %q) = %q, want %q", tc.name, got, tc.plural)
		}
		if got := pluralName(1, tc.name); got != tc.name {
			t.Errorf("pluralName(1, %q) = %q, want it unchanged", tc.name, got)
		}
	}
}

// TestPluralNameCoversConfig keeps pluralCases complete: a building whose
// name ends in s, x, ch, sh or y has to say how it pluralizes.
func TestPluralNameCoversConfig(t *testing.T) {
	known := make(map[string]bool, len(pluralCases))
	for _, tc := range pluralCases {
		known[tc.name] = true
	}
	for _, d := range config.BaseBuildings() {
		for _, suffix := range []string{"s", "x", "ch", "sh", "y"} {
			if strings.HasSuffix(d.Name, suffix) && !known[d.Name] {
				t.Errorf("building %s (%q) ends in %q: add its plural to pluralCases", d.Key, d.Name, suffix)
				break
			}
		}
	}
}

func TestBuildingCountPlural(t *testing.T) {
	cases := []struct {
		n         int
		key, want string
	}{
		{1, "barracks", "1 Barracks"},
		{3, "barracks", "3 Barracks"},
		{1, "hut", "1 Hut"},
		{0, "hut", "0 Huts"},
		{2, "granary", "2 Granaries"},
		{4, "workshop", "4 Workshops"},
		{2, "fortress", "2 Fortresses"},
	}
	for _, tc := range cases {
		if got := BuildingCount(tc.n, tc.key); got != tc.want {
			t.Errorf("BuildingCount(%d, %q) = %q, want %q", tc.n, tc.key, got, tc.want)
		}
	}
}
