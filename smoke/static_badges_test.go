package smoke

import (
	"math"
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

// ladderFamily returns a pointer to the ladder family in src that the badge
// case calls name ("Wonders raised").
func ladderFamily(t *testing.T, src *rules.Source, name string) *config.BadgeFamilyDef {
	t.Helper()
	for i := range src.BadgeFamilies {
		if src.BadgeFamilies[i].Family == "ladder" && src.BadgeFamilies[i].Ladder == name {
			return &src.BadgeFamilies[i]
		}
	}
	t.Fatalf("no ladder %s", name)
	return nil
}

// TestFirstRungsComeEarly is the first-rung rule on the real catalog, with
// the table it rests on: every lifetime ladder's first rung is one
// occurrence, or a count an ordinary first run has by the time it reaches
// the Bronze Age, or the second age after the ladder's subject first
// appears. It was written after a first playthrough met "Produce 1B food"
// as the first food rung, with a million food made by the Bronze Age.
func TestFirstRungsComeEarly(t *testing.T) {
	first := StaticFirstRungs()
	if len(first) < 60 {
		t.Fatalf("only %d lifetime ladders were checked", len(first))
	}
	byName := map[string]FirstRung{}
	for _, r := range first {
		if r.Need > 1 && r.Need > r.Have {
			t.Errorf("%s: the first rung (%s) asks for %s, and an ordinary first run has %s by the %s", r.Ladder, r.Key, config.FormatAmount(r.Need), config.FormatAmount(r.Have), r.By)
		}
		for i := 1; i < len(r.Rungs); i++ {
			if r.Rungs[i] <= r.Rungs[i-1] {
				t.Errorf("%s: rung %d (%v) is not above rung %d (%v)", r.Ladder, i+1, r.Rungs[i], i, r.Rungs[i-1])
			}
		}
		if strings.HasPrefix(r.Key, "resource.") {
			byName[strings.TrimSuffix(strings.TrimPrefix(r.Key, "resource."), "."+r.Key[strings.LastIndex(r.Key, ".")+1:])] = r
		}
		t.Logf("%-22s first rung %-8s from the %-16s by the %-16s (a first run has %s)", r.Ladder, config.FormatAmount(r.Need), r.From, r.By, config.FormatAmount(r.Have))
	}

	// Food: a town passes the first rung well before it leaves the Stone
	// Age, a quarter of the way into what it makes there at the latest.
	set := rules.Core()
	food, ok := byName["food"]
	if !ok {
		t.Fatal("no food ladder")
	}
	if most := set.ProductionThrough("food", "primitive_age") + 0.25*(set.ProductionThrough("food", "stone_age")-set.ProductionThrough("food", "primitive_age")); food.Need > most {
		t.Errorf("the first food rung asks for %s; a town has made %s a quarter of the way through the Stone Age", config.FormatAmount(food.Need), config.FormatAmount(most))
	}
	if food.Need >= 1e6 {
		t.Errorf("the first food rung asks for %s: the playtest had made a million by the Bronze Age and was nowhere near the old one", config.FormatAmount(food.Need))
	}

	// Every resource ladder climbs in even multiplicative steps, from the
	// first age's production to twenty-five runs: no step is more than a
	// quarter longer than another (the rungs are rounded to two figures).
	if len(byName) != len(set.Resources()) {
		t.Errorf("%d resource ladders for %d resources", len(byName), len(set.Resources()))
	}
	badges := 0
	for _, res := range set.Resources() {
		r := byName[res.Key]
		badges += len(r.Rungs)
		if len(r.Rungs) < 3 {
			t.Errorf("%s has %d rungs", res.Key, len(r.Rungs))
			continue
		}
		low, high := math.Inf(1), 0.0
		for i := 1; i < len(r.Rungs); i++ {
			step := r.Rungs[i] / r.Rungs[i-1]
			low, high = math.Min(low, step), math.Max(high, step)
		}
		if high > 1.25*low {
			t.Errorf("%s: the rungs %v climb in steps from x%.3g to x%.3g, not evenly", res.Key, r.Rungs, low, high)
		}
		perRun, _ := set.RunProduction(res.Key)
		if top := r.Rungs[len(r.Rungs)-1]; top > config.BadgeMaxRuns*perRun || top < 0.9*config.BadgeMaxRuns*perRun {
			t.Errorf("%s: the top rung is %s, want %d runs of %s", res.Key, config.FormatAmount(top), config.BadgeMaxRuns, config.FormatAmount(perRun))
		}
	}
	if badges != 94 {
		t.Errorf("the resource ladders have %d badges, want the 94 the catalog had", badges)
	}
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
		{"a first rung a first run does not build in time", "lineage.housing.1", "first_rung", func(src *rules.Source) {
			family(t, src, "lineage").Ladders = map[string][]float64{"housing": {30, 57, 230, 570, 1400}}
		}},
		{"a first rung of two runs' wonders", "ladder.wonders.1", "first_rung", func(src *rules.Source) {
			ladderFamily(t, src, "Wonders raised").Ladders = map[string][]float64{"": {25, 100, 300}}
		}},
		{"a first rung of more badges than the first ages give", "special.collector", "first_rung", func(src *rules.Source) {
			special(t, src, "special.collector").Threshold = 25
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
		{"a subject the event never carries", "domain.food.1", "existence", func(src *rules.Source) {
			// The census has no subject: a row held to one waits for ever.
			family(t, src, "domain").AnySubject = false
		}},
		{"an age where the event names a building", "maximalist.stone_age", "existence", func(src *rules.Source) {
			family(t, src, "maximalist").AnySubject = false
		}},
		{"an age where the event names a tech", "techs.iron_age", "existence", func(src *rules.Source) {
			family(t, src, "techs").AnySubject = false
		}},
		{"a civilization that does not exist", "special.sale_probe", "existence", func(src *rules.Source) {
			src.Badges = append(src.Badges, config.BadgeDef{
				Key: "special.sale_probe", Family: "special", Name: "Probe", Desc: "Meet Atlantis.",
				Tier: config.BadgeBronze, Scope: config.BadgeMoment, Event: config.BadgeEvCivMet, Subject: "atlantis",
				Proof: config.Occurs("It does not."),
			})
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
