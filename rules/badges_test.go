package rules

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The seed catalog: the four account achievements of account.json version
// 1 as badges, and one or two more of each kind. A change here is a change
// to what accounts hold, so the keys are pinned.
func TestSeedCatalog(t *testing.T) {
	want := []string{
		"age.stone_age", "age.iron_age", "age.modern_age",
		"lineage.housing.1", "lineage.housing.2",
		"ladder.prestiges.1", "ladder.prestiges.2", "ladder.prestiges.3", "ladder.prestiges.4",
		"special.hut_hoarder", "special.fashionably_late", "special.liquidation_sale",
		"special.hand_in_the_cookie_jar", "special.touched_by_the_source", "special.creative_accounting",
	}
	var got []string
	for _, b := range Core().Badges() {
		got = append(got, b.Key)
	}
	if !slices.Equal(got, want) {
		t.Errorf("the catalog is\n %v\nwant\n %v", got, want)
	}
	for alias, key := range map[string]string{
		"first_prestige": "ladder.prestiges.1", "prestige_x10": "ladder.prestiges.3",
		"reached_iron": "age.iron_age", "reached_modern": "age.modern_age",
	} {
		if got, ok := Core().BadgeForAlias(alias); !ok || got != key {
			t.Errorf("the old achievement %s is the badge %q (found %v), want %s", alias, got, ok, key)
		}
	}
	if _, ok := Core().BadgeForAlias("no_such_achievement"); ok {
		t.Error("an unknown alias found a badge")
	}
	if problems := BadgeProblems(config.Badges(), config.BadgeFamilies(), Core()); len(problems) != 0 {
		t.Errorf("the badge tables have problems: %v", problems)
	}
}

// What a family makes of one row of its table.
func TestFamilyBadges(t *testing.T) {
	s := Core()

	iron, ok := s.Badge("age.iron_age")
	if !ok {
		t.Fatal("no Iron Age badge")
	}
	if iron.Name != "Age of Iron" || iron.Desc != "Reach the Iron Age." || iron.Subject != "iron_age" ||
		iron.Family != "age" || iron.Event != config.BadgeEvAgeReached || iron.Scope != config.BadgeMoment {
		t.Errorf("the Iron Age badge: %+v", iron)
	}
	if iron.Tier != config.BadgeBronze || iron.Reveal != config.RevealUntilNextAge("iron_age") {
		t.Errorf("the Iron Age badge's tier and reveal: %v %+v", iron.Tier, iron.Reveal)
	}
	if modern, _ := s.Badge("age.modern_age"); modern.Tier != config.BadgeGold {
		t.Errorf("the Modern Age is in the Digital Era, so its badge is gold; got %s", modern.Tier.Name())
	}

	rung, ok := s.Badge("lineage.housing.2")
	if !ok {
		t.Fatal("no second housing rung")
	}
	if rung.Name != "Housing Contractor" || rung.Tier != config.BadgeSilver || rung.Threshold != 57 ||
		rung.Counter != "built.lineage.housing" || rung.Scope != config.BadgeLifetime {
		t.Errorf("the second housing rung: %+v", rung)
	}
	if want := "Build 57 housing buildings across all your runs. Sold and rebuilt copies count once."; rung.Desc != want {
		t.Errorf("description %q, want %q", rung.Desc, want)
	}
	if rung.Reveal.Kind != config.BadgeVisible {
		t.Errorf("housing starts in the first age, so its ladder shows from the start: %+v", rung.Reveal)
	}
	if _, ok := s.Badge("lineage.housing.3"); ok {
		t.Error("a rung with no count in the ladder table was made")
	}

	first, _ := s.Badge("ladder.prestiges.1")
	if first.Name != "First Prestige" || first.Desc != "Prestige for the first time." || first.Subject != "" {
		t.Errorf("the first prestige rung: %+v", first)
	}
	if last, _ := s.Badge("ladder.prestiges.4"); last.Desc != "Prestige 25 times." || last.Tier != config.BadgeLegendary {
		t.Errorf("the last prestige rung: %+v", last)
	}
}

// Every table a family can be made from gives one badge per row, with the
// reveal rule that fits the table: the full catalog is rows of these.
func TestFamilySources(t *testing.T) {
	src := FromConfig()
	lineages := map[string]bool{}
	for _, b := range src.Buildings {
		if b.LineageKey != "" && b.LineageKey != lineageWonder && b.LineageKey != lineageMonuments {
			lineages[b.LineageKey] = true
		}
	}
	cases := []struct {
		name   string
		source config.BadgeSource
		want   int
		reveal config.BadgeRevealKind // of the last badge made
	}{
		{"ages", config.BadgeSourceAges, len(src.Ages), config.BadgeRevealNextAge},
		{"eras", config.BadgeSourceEras, len(src.Eras), config.BadgeRevealAtAge},
		{"wonders", config.BadgeSourceWonders, len(src.Ages), config.BadgeRevealAtAge},
		{"lineages", config.BadgeSourceLineages, len(lineages), config.BadgeRevealAtAge},
		{"resources", config.BadgeSourceResources, len(src.Resources), config.BadgeRevealAtAge},
		{"civs", config.BadgeSourceCivs, len(src.Factions), config.BadgeRevealOnCounter},
		{"harbingers", config.BadgeSourceHarbingers, len(src.Harbingers), config.BadgeRevealOnCounter},
		{"awakenings", config.BadgeSourceAwakenings, len(src.Awakenings), config.BadgeRevealOnCounter},
		{"none", config.BadgeSourceNone, 1, config.BadgeVisible},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fam := config.BadgeFamilyDef{
				Family: "probe", Source: tc.source, Key: "probe.{key}", Name: "{name}", Desc: "About {lname}.",
				Tier: config.BadgeSilver, Scope: config.BadgeMoment, Event: config.BadgeEvAgeReached,
				RevealBySubject: true, Proof: config.StaticProof(config.BadgeRuleGate),
			}
			src := FromConfig()
			src.Badges, src.BadgeFamilies = nil, []config.BadgeFamilyDef{fam}
			got := Compile(src).Badges()
			if len(got) != tc.want {
				t.Fatalf("%d badges, want one per row: %d", len(got), tc.want)
			}
			seen := map[string]bool{}
			for _, b := range got {
				if seen[b.Key] {
					t.Errorf("the key %s was made twice", b.Key)
				}
				seen[b.Key] = true
				if tc.source != config.BadgeSourceNone && (b.Subject == "" || b.Name == "") {
					t.Errorf("a badge with no subject or name: %+v", b)
				}
			}
			if last := got[len(got)-1]; last.Reveal.Kind != tc.reveal {
				t.Errorf("the last badge is revealed by rule %d, want %d: %+v", last.Reveal.Kind, tc.reveal, last.Reveal)
			}
			// Something of the first age is never hidden behind an age.
			if first := got[0]; first.Reveal.Kind == config.BadgeRevealAtAge || first.Reveal.Kind == config.BadgeRevealNextAge {
				if pos, _ := Core().Index(first.Reveal.Key); pos == 0 {
					t.Errorf("the first badge hides behind the first age: %+v", first.Reveal)
				}
			}
		})
	}
}

// Only and Except choose rows; a family that chooses none, or a row its
// table does not have, is a problem the guard reports.
func TestFamilyOnlyAndExcept(t *testing.T) {
	fam := config.BadgeFamilyDef{
		Family: "probe", Source: config.BadgeSourceAges, Key: "probe.{key}", Name: "{name}", Desc: "Reach the {name}.",
		Scope: config.BadgeMoment, Event: config.BadgeEvAgeReached,
	}
	count := func(f config.BadgeFamilyDef) (int, []string) {
		src := FromConfig()
		src.Badges, src.BadgeFamilies = nil, []config.BadgeFamilyDef{f}
		set := Compile(src)
		return len(set.Badges()), BadgeProblems(nil, src.BadgeFamilies, set)
	}
	all, problems := count(fam)
	if all != Core().NumAges() || len(problems) != 0 {
		t.Fatalf("the whole table: %d badges, problems %v", all, problems)
	}
	fam.Except = []string{"primitive_age"}
	if n, _ := count(fam); n != all-1 {
		t.Errorf("Except dropped %d rows, want 1", all-n)
	}
	fam.Except, fam.Only = nil, []string{"iron_age", "atlantis_age"}
	n, problems := count(fam)
	if n != 1 || len(problems) != 1 {
		t.Errorf("Only with an unknown row: %d badges, problems %v", n, problems)
	}
	fam.Only = []string{"atlantis_age"}
	if n, problems := count(fam); n != 0 || len(problems) < 1 {
		t.Errorf("a family that keeps no row: %d badges, problems %v", n, problems)
	}
}

// The core set's digest must not change when badges are read from it: a
// Set never changes.
func TestBadgesAreCopies(t *testing.T) {
	s := Core()
	before := s.Digest()
	list := s.Badges()
	list[0].Name = "changed"
	list[0].Aliases = append(list[0].Aliases, "changed")
	if b, _ := s.Badge(list[0].Key); b.Name == "changed" {
		t.Error("writing to a returned badge changed the set")
	}
	if s.Digest() != before {
		t.Error("reading badges changed the set's digest")
	}
}
