package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestBadgesAreReachable is the Badge Covenant on the real catalog: every
// badge must carry a proof that holds against the game's own numbers, so a
// badge that cannot be earned fails the build. The rules are listed at the
// top of static_badges.go. A failure says which number breaks.
func TestBadgesAreReachable(t *testing.T) {
	problems, reach := StaticBadges()
	for _, p := range problems {
		t.Errorf("[%s] %s", p.Kind, p.Why)
	}
	if len(reach) == 0 {
		t.Fatal("the catalog has no badges")
	}
	for _, r := range reach {
		if r.Limit > 0 {
			t.Logf("%-34s %-17s asks %-6s of %-8s (%s)", r.Key, r.Proof, config.FormatAmount(r.Need), config.FormatAmount(r.Limit), r.Basis)
		} else {
			t.Logf("%-34s %-17s", r.Key, r.Proof)
		}
	}
}

// brokenBadges runs the covenant on the real catalog with change applied
// to its tables, and returns the kinds of problem found for key ("" for a
// table problem).
func brokenBadges(t *testing.T, key string, change func(src *rules.Source)) []string {
	t.Helper()
	src := rules.FromConfig()
	change(&src)
	set := rules.Compile(src)
	problems, _ := staticBadges(set, src.Badges, src.BadgeFamilies, config.BuildingByKey(), game.PrestigeRunAge)
	var kinds []string
	for _, p := range problems {
		if p.Key == key {
			kinds = append(kinds, p.Kind)
		}
	}
	return kinds
}

// special returns a pointer to a hand-written badge in src by key.
func special(t *testing.T, src *rules.Source, key string) *config.BadgeDef {
	t.Helper()
	for i := range src.Badges {
		if src.Badges[i].Key == key {
			return &src.Badges[i]
		}
	}
	t.Fatalf("no hand-written badge %s", key)
	return nil
}

// family returns a pointer to a family in src by its name.
func family(t *testing.T, src *rules.Source, name string) *config.BadgeFamilyDef {
	t.Helper()
	for i := range src.BadgeFamilies {
		if src.BadgeFamilies[i].Family == name {
			return &src.BadgeFamilies[i]
		}
	}
	t.Fatalf("no badge family %s", name)
	return nil
}

// TestBadgeGuardCatchesTheUnreachable breaks the catalog one way at a
// time and checks the covenant says so: a guard that passes everything
// proves nothing.
func TestBadgeGuardCatchesTheUnreachable(t *testing.T) {
	const hut, late, sale, jar = "special.hut_hoarder", "special.fashionably_late", "special.liquidation_sale", "special.hand_in_the_cookie_jar"
	cases := []struct {
		name   string
		key    string
		kind   string
		change func(src *rules.Source)
	}{
		{"more copies than can stand", hut, "copies", func(src *rules.Source) {
			special(t, src, hut).Threshold = 5000
		}},
		{"every copy that can stand", hut, "copies", func(src *rules.Source) {
			// The ceiling itself is over the share a copies badge may ask.
			m := newMilestoneModel(config.BuildingByKey(), game.PrestigeRunAge)
			special(t, src, hut).Threshold = float64(m.ceiling("hut").n)
		}},
		{"copies counted in another age", hut, "copies", func(src *rules.Source) {
			special(t, src, hut).InAge = "stone_age"
		}},
		{"copies of something that is not a building", hut, "copies", func(src *rules.Source) {
			b := special(t, src, hut)
			b.Subject, b.Counter = "hovel", "standing.hovel"
		}},
		{"a ladder past 25 runs", "lineage.housing.2", "lifetime", func(src *rules.Source) {
			family(t, src, "lineage").Ladders = map[string][]float64{"housing": {14, 1e6}}
		}},
		{"a ladder of prestiges past 25 runs", "ladder.prestiges.4", "lifetime", func(src *rules.Source) {
			family(t, src, "ladder").Ladders = map[string][]float64{"": {1, 3, 10, 26}}
		}},
		{"a counter no event adds to", "ladder.prestiges.1", "existence", func(src *rules.Source) {
			family(t, src, "ladder").Counter = "high_fives"
		}},
		{"a counter the check has no run for", "ladder.prestiges.1", "lifetime", func(src *rules.Source) {
			family(t, src, "ladder").Counter = config.BadgeEvGiftSent
		}},
		{"an age no advance reaches", "age.primitive_age", "gate", func(src *rules.Source) {
			f := family(t, src, "age")
			f.Except, f.Only = nil, []string{"primitive_age"}
		}},
		{"an event the game does not report", sale, "existence", func(src *rules.Source) {
			special(t, src, sale).Event = "building_juggled"
		}},
		{"a predicate the engine does not have", late, "existence", func(src *rules.Source) {
			special(t, src, late).Pred = "age_nap"
		}},
		{"a time threshold not read as a multiple of the target", late, "time", func(src *rules.Source) {
			b := special(t, src, late)
			b.Pred, b.Counter = "", "run."+config.BadgeEvStarved
		}},
		{"more sales than a run can make", sale, "run_count", func(src *rules.Source) {
			special(t, src, sale).Threshold = 1e7
		}},
		{"a run count the check cannot bound", sale, "run_count", func(src *rules.Source) {
			special(t, src, sale).Counter = "run." + config.BadgeEvGiftSent
		}},
		{"no proof", sale, "proof", func(src *rules.Source) {
			special(t, src, sale).Proof = config.BadgeProof{}
		}},
		{"a bot style no bot plays", sale, "bot", func(src *rules.Source) {
			special(t, src, sale).Proof = config.BadgeProof{Kind: config.BadgeProofBot, Rule: "speedrunner"}
		}},
		{"a later age named in a visible description", late, "spoiler", func(src *rules.Source) {
			special(t, src, late).Desc = "Spend ten times the Primitive Age's pacing target there, long before the Modern Age."
		}},
		{"a later age named in a secret badge's hint", sale, "spoiler", func(src *rules.Source) {
			special(t, src, sale).Hint = "Something about the Iron Age."
		}},
		{"an exclamation mark", sale, "text", func(src *rules.Source) {
			special(t, src, sale).Desc = "Sell 100 buildings in one run!"
		}},
		{"a dash", late, "text", func(src *rules.Source) {
			special(t, src, late).Name = "Fashionably Late — Very"
		}},
		{"a secret badge with no hint", sale, "existence", func(src *rules.Source) {
			special(t, src, sale).Hint = ""
		}},
		{"a run fact about the session", sale, "existence", func(src *rules.Source) {
			special(t, src, sale).When = []config.BadgeCond{{Fact: "run." + config.BadgeEvDayPlayed, Op: config.BadgeAtLeast, Value: 1}}
		}},
		{"a theme that does not exist", late, "existence", func(src *rules.Source) {
			special(t, src, late).Reward = config.BadgeReward{Theme: "plaid"}
		}},
		{"an integrity badge with a tier", jar, "integrity", func(src *rules.Source) {
			special(t, src, jar).Tier = config.BadgeGold
		}},
		{"an integrity badge not on the list", sale, "integrity", func(src *rules.Source) {
			special(t, src, sale).Proof = config.BadgeProof{Kind: config.BadgeProofIntegrity}
		}},
		{"a listed integrity badge that is not one", jar, "integrity", func(src *rules.Source) {
			special(t, src, jar).Proof = config.StaticProof(config.BadgeRuleGate)
		}},
		{"an old achievement no badge keeps", "prestige_x10", "alias", func(src *rules.Source) {
			family(t, src, "ladder").Aliases = map[string][]string{"ladder.prestiges.1": {"first_prestige"}}
		}},
		{"a family that makes no badges", "", "table", func(src *rules.Source) {
			family(t, src, "lineage").Only = []string{"basket_weaving"}
		}},
		{"two rungs at one count", "ladder.prestiges.2", "table", func(src *rules.Source) {
			family(t, src, "ladder").Ladders = map[string][]float64{"": {1, 1, 10, 25}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kinds := brokenBadges(t, tc.key, tc.change)
			found := false
			for _, k := range kinds {
				found = found || k == tc.kind
			}
			if !found {
				t.Errorf("the covenant did not report a %q problem for %q; it reported %v", tc.kind, tc.key, kinds)
			}
		})
	}
}

// TestBadgeProofsNameTheNumber: a failure reads as a sentence with the
// numbers in it, like the Milestone Covenant's.
func TestBadgeProofsNameTheNumber(t *testing.T) {
	src := rules.FromConfig()
	for i := range src.Badges {
		if src.Badges[i].Key == "special.hut_hoarder" {
			src.Badges[i].Threshold = 5000
		}
	}
	problems, _ := staticBadges(rules.Compile(src), src.Badges, src.BadgeFamilies, config.BuildingByKey(), game.PrestigeRunAge)
	if len(problems) != 1 {
		t.Fatalf("want one problem, got %+v", problems)
	}
	for _, want := range []string{"Hut Hoarder", "5,000", "Huts", "Primitive Age", "wood", "97%"} {
		if !strings.Contains(problems[0].Why, want) {
			t.Errorf("the failure does not mention %q: %s", want, problems[0].Why)
		}
	}
}

// TestHutHoarderQuotesRealPrices: Hut Hoarder's description quotes the price
// of the 60th hut and how many times the first that is. Both come from the
// cost curve, so a change to the curve must change the text.
func TestHutHoarderQuotesRealPrices(t *testing.T) {
	def, ok := rules.Core().Badge("special.hut_hoarder")
	if !ok {
		t.Fatal("no Hut Hoarder badge")
	}
	hut := config.BuildingByKey()[def.Subject]
	first := lastCopyPrice(hut, "wood", 1)
	last := lastCopyPrice(hut, "wood", int(def.Threshold))
	price := grouped(int(last))
	times := grouped(int(last / first))
	want := "The " + grouped(int(def.Threshold)) + "th costs " + price + " wood, " + times + " times the first."
	if !strings.Contains(def.Desc, want) {
		t.Errorf("Hut Hoarder says %q, but the cost curve says %q", def.Desc, want)
	}
}
