package game

import "testing"

// pluralCases has a row or two for each rule pluralName implements, with the
// plural a player should read.
var pluralCases = []struct{ in, want string }{
	// Regular: add s.
	{"Hut", "Huts"},
	{"Workshop", "Workshops"},

	// Already plural or invariant: ends in s, but not ss or us.
	{"Barracks", "Barracks"},
	{"Standing Stones", "Standing Stones"},
	{"Ironworks", "Ironworks"},

	// "X of Y" pluralizes X, by the same rules.
	{"Monument of Ages", "Monuments of Ages"},
	{"Academy of Arts", "Academies of Arts"},

	// Singular, ending in ss or us: es.
	{"Fortress", "Fortresses"},
	{"Research Campus", "Research Campuses"},
	{"Warp Nexus", "Warp Nexuses"},

	// x, ch, sh: es.
	{"Launch Complex", "Launch Complexes"},
	{"Megaplex", "Megaplexes"},
	{"Church", "Churches"},
	{"Stash", "Stashes"},

	// Consonant + y: ies.
	{"Granary", "Granaries"},
	{"Dark Matter Refinery", "Dark Matter Refineries"},

	// Vowel + y: s.
	{"Hydroponic Bay", "Hydroponic Bays"},
	{"Solar Collector Array", "Solar Collector Arrays"},
	{"Alloy", "Alloys"},

	// A lone y has no letter before it to look at.
	{"y", "ys"},
}

func TestPluralName(t *testing.T) {
	for _, tc := range pluralCases {
		if got := pluralName(2, tc.in); got != tc.want {
			t.Errorf("pluralName(2, %q) = %q, want %q", tc.in, got, tc.want)
		}
		if got := pluralName(0, tc.in); got != tc.want {
			t.Errorf("pluralName(0, %q) = %q, want %q", tc.in, got, tc.want)
		}
		for _, n := range []int{1, -1} {
			if got := pluralName(n, tc.in); got != tc.in {
				t.Errorf("pluralName(%d, %q) = %q, want it unchanged", n, tc.in, got)
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
