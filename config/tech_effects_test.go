package config

import (
	"math"
	"strings"
	"testing"
)

// TestTechEffectKindsStack pins how each kind stacks and what holds it: the
// kinds that cut a time or a price multiply and have a floor, the rest add
// and have none, and only game speed, military power and expedition rewards
// join a pool the rest of the game's bonuses share.
func TestTechEffectKindsStack(t *testing.T) {
	inf := math.Inf(1)
	want := map[TechEffectKind]struct {
		multiplies bool
		lo, hi     float64
		pool       string
	}{
		EffectOutput:           {false, -inf, inf, ""},
		EffectAllOutput:        {false, -inf, inf, ""},
		EffectFlatOutput:       {false, -inf, inf, ""},
		EffectStorage:          {false, -inf, inf, ""},
		EffectHousing:          {false, -inf, inf, ""},
		EffectBuildCost:        {true, BuildCostFloor, 1, ""},
		EffectBuildTime:        {true, BuildTimeFloor, 1, ""},
		EffectResearchTime:     {true, ResearchTimeFloor, 1, ""},
		EffectGameSpeed:        {false, -inf, inf, "tick_speed"},
		EffectMilitaryPower:    {false, -inf, inf, "military_power"},
		EffectExpeditionReward: {false, -inf, inf, "expedition_reward"},
	}
	for _, kind := range TechEffectKinds() {
		if kind == EffectMechanic {
			continue // TestMechanics
		}
		w, ok := want[kind]
		if !ok {
			t.Errorf("kind %q is not in this test: say how it stacks", kind)
			continue
		}
		k := TechEffectKey{Kind: kind}
		if kind.TakesResource() {
			k.Target = "gold"
		}
		lo, hi := k.Limits()
		pool, _ := k.Pool()
		if k.Multiplies() != w.multiplies || lo != w.lo || hi != w.hi || pool != w.pool {
			t.Errorf("%s: multiplies %v, limits %v to %v, pool %q; want %v, %v to %v, %q", kind, k.Multiplies(), lo, hi, pool, w.multiplies, w.lo, w.hi, w.pool)
		}
	}
	// Two +5% add up to +10%. Two 3% cuts leave 0.97 x 0.97.
	add := TechEffectKey{Kind: EffectOutput, Target: "gold"}
	if got := add.Fold(add.Fold(add.Start(), 0.05), 0.05); got != 0.1 {
		t.Errorf("two +5%% output bonuses fold to %v, want 0.1", got)
	}
	cut := TechEffectKey{Kind: EffectBuildCost}
	if got, want := cut.Fold(cut.Fold(cut.Start(), -0.03), -0.03), 0.97*0.97; math.Abs(got-want) > 1e-15 {
		t.Errorf("two 3%% cost cuts fold to %v, want %v", got, want)
	}
	// A floor holds however many cuts stack.
	term := cut.Start()
	for i := 0; i < 200; i++ {
		term = cut.Fold(term, -0.03)
	}
	if got := cut.Clamp(term); got != BuildCostFloor {
		t.Errorf("200 cost cuts clamp to %v, want the floor %v", got, BuildCostFloor)
	}
}

// TestTechEffectText: every kind reads in player words.
func TestTechEffectText(t *testing.T) {
	for _, tc := range []struct {
		e    TechEffect
		want string
	}{
		{TechEffect{Kind: EffectOutput, Target: "food", Value: 0.10}, "+10% food production"},
		{TechEffect{Kind: EffectAllOutput, Value: 0.05}, "+5% all production"},
		{TechEffect{Kind: EffectFlatOutput, Target: "steel", Value: 0.25}, "+0.25 steel/tick"},
		{TechEffect{Kind: EffectStorage, Value: 0.10}, "+10% storage"},
		{TechEffect{Kind: EffectHousing, Value: 0.05}, "+5% housing"},
		{TechEffect{Kind: EffectBuildCost, Value: -0.03}, "buildings cost 3% less"},
		{TechEffect{Kind: EffectBuildTime, Value: -0.08}, "construction takes 8% less time"},
		{TechEffect{Kind: EffectResearchTime, Value: -0.03}, "research takes 3% less time"},
		{TechEffect{Kind: EffectGameSpeed, Value: 0.05}, "+5% game speed"},
		{TechEffect{Kind: EffectMilitaryPower, Value: 0.15}, "+15% military power"},
		{TechEffect{Kind: EffectExpeditionReward, Value: 0.10}, "+10% expedition rewards"},
		{TechEffect{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.03}, "market fee 3 points lower"},
		{TechEffect{Kind: EffectMechanic, Target: MechanicRouteTicks, Value: -0.15}, "trade routes take 15% less time"},
		{TechEffect{Kind: EffectMechanic, Target: MechanicGatherAmount, Value: 2}, "gathering by hand brings 2 more"},
	} {
		if got := tc.e.Text(); got != tc.want {
			t.Errorf("%+v reads %q, want %q", tc.e, got, tc.want)
		}
	}
}

// TestTechEffectCheck: the check that guards config catches each way a
// typed effect can be wrong.
func TestTechEffectCheck(t *testing.T) {
	isRes := func(k string) bool { return k == "gold" }
	for _, tc := range []struct {
		name string
		e    TechEffect
		want string // a piece of the problem; "" for a valid effect
	}{
		{"valid output", TechEffect{Kind: EffectOutput, Target: "gold", Value: 0.1}, ""},
		{"valid all output", TechEffect{Kind: EffectAllOutput, Value: 0.05}, ""},
		{"valid mechanic", TechEffect{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.03}, ""},
		{"valid cut", TechEffect{Kind: EffectBuildTime, Value: -0.08}, ""},
		{"unknown kind", TechEffect{Kind: "gather_rate", Value: 0.1}, "unknown effect kind"},
		{"the old worker output kind", TechEffect{Kind: "worker_output", Value: 0.1}, "unknown effect kind"},
		{"no value", TechEffect{Kind: EffectOutput, Target: "gold"}, "no value"},
		{"target on a kind that takes none", TechEffect{Kind: EffectStorage, Target: "gold", Value: 0.1}, "takes no target"},
		{"not a resource", TechEffect{Kind: EffectOutput, Target: "mana", Value: 0.1}, "not a resource"},
		{"output of everything", TechEffect{Kind: EffectOutput, Target: AllResources, Value: 0.1}, "not a resource"},
		{"not a mechanic", TechEffect{Kind: EffectMechanic, Target: "wonder_ticks", Value: -0.1}, "not a mechanic"},
		{"a cut that leaves nothing", TechEffect{Kind: EffectBuildCost, Value: -1}, "leaves nothing"},
	} {
		got := tc.e.Check(isRes)
		if (tc.want == "") != (got == "") || !strings.Contains(got, tc.want) {
			t.Errorf("%s: Check = %q, want %q", tc.name, got, tc.want)
		}
	}
	// Every effect in the game passes.
	res := ResourceByKey()
	for _, tech := range Technologies() {
		for _, e := range tech.Effects {
			if p := e.Check(func(k string) bool { _, ok := res[k]; return ok }); p != "" {
				t.Errorf("%s: %s", tech.Key, p)
			}
		}
	}
}

// TestMechanics: the table of mechanic numbers is sound, and every number in
// it is one a tech of the game moves (a number joins the table with the
// tech that moves it).
func TestMechanics(t *testing.T) {
	seen := map[string]bool{}
	used := map[string]bool{}
	for _, tech := range Technologies() {
		for _, e := range tech.Effects {
			if e.Kind == EffectMechanic {
				used[e.Target] = true
			}
		}
	}
	for _, m := range Mechanics() {
		k := TechEffectKey{Kind: EffectMechanic, Target: m.Key}
		switch {
		case m.Key == "" || m.Name == "" || seen[m.Key]:
			t.Errorf("mechanic %q (%q) is unnamed or listed twice", m.Key, m.Name)
		case strings.Count(m.Text, "%s") != 1:
			t.Errorf("%s: text %q needs exactly one %%s", m.Key, m.Text)
		case !(m.Min < m.Max):
			t.Errorf("%s: limits %v to %v", m.Key, m.Min, m.Max)
		case k.Start() < m.Min || k.Start() > m.Max:
			t.Errorf("%s: with no tech its term is %v, outside %v to %v", m.Key, k.Start(), m.Min, m.Max)
		case k.Multiplies() != m.Multiplies:
			t.Errorf("%s: the key multiplies %v, the table says %v", m.Key, k.Multiplies(), m.Multiplies)
		case !used[m.Key]:
			t.Errorf("%s: no tech moves it; a number joins the table with its first tech", m.Key)
		}
		seen[m.Key] = true
	}
	fee := MechanicByKey()[MechanicMarketFee]
	if got := fee.Apply(ExchangeFee, -0.03); math.Abs(got-0.17) > 1e-12 {
		t.Errorf("a fee of %v with 3 points off is %v, want 0.17", ExchangeFee, got)
	}
	route := MechanicByKey()[MechanicRouteTicks]
	if got := route.Apply(100, 0.85); got != 85 {
		t.Errorf("a 100 tick route at x0.85 is %v, want 85", got)
	}
}

// TestTechHeadroom is the headroom test: with every tech of the game
// researched, no term has reached a floor or a cap, so the last tech taken
// still does its full step. A tech that breaks this fails the build.
func TestTechHeadroom(t *testing.T) {
	terms := TechTerms(Technologies())
	if len(terms) == 0 {
		t.Fatal("no tech has an effect")
	}
	for k, term := range terms {
		lo, hi := k.Limits()
		if term <= lo || term >= hi && k.Start() != hi {
			t.Errorf("%s %s: every tech together comes to %v, at or past its limit (%v to %v)", k.Kind, k.Target, term, lo, hi)
		}
		if k.Clamp(term) != term {
			t.Errorf("%s %s: every tech together comes to %v, which its limits change to %v", k.Kind, k.Target, term, k.Clamp(term))
		}
	}
	// The sums of today's tree, pinned so a change to them is a decision.
	near := func(got, want float64) bool { return math.Abs(got-want) < 1e-9 }
	for _, tc := range []struct {
		key  TechEffectKey
		want float64
	}{
		{TechEffectKey{Kind: EffectAllOutput}, 0.10},
		{TechEffectKey{Kind: EffectOutput, Target: "food"}, 0.94},
		{TechEffectKey{Kind: EffectOutput, Target: "knowledge"}, 0.85},
		{TechEffectKey{Kind: EffectOutput, Target: "gold"}, 0.31},
		{TechEffectKey{Kind: EffectOutput, Target: "faith"}, 0.28},
		{TechEffectKey{Kind: EffectOutput, Target: "culture"}, 0.27},
		{TechEffectKey{Kind: EffectOutput, Target: "steel"}, 0.27},
		{TechEffectKey{Kind: EffectOutput, Target: "coal"}, 0.23},
		{TechEffectKey{Kind: EffectStorage}, 0.41},
		{TechEffectKey{Kind: EffectHousing}, 0.37},
		{TechEffectKey{Kind: EffectBuildCost}, 0.97 * 0.97 * 0.96 * 0.97 * 0.98 * 0.96 * 0.98},
		{TechEffectKey{Kind: EffectBuildTime}, 0.92 * 0.95 * 0.94 * 0.95 * 0.92 * 0.95 * 0.95},
		{TechEffectKey{Kind: EffectResearchTime}, 0.97 * 0.97 * 0.97 * 0.94 * 0.96 * 0.97 * 0.94},
		{TechEffectKey{Kind: EffectGameSpeed}, 0.30},
		{TechEffectKey{Kind: EffectMilitaryPower}, 1.41},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicMarketFee}, -0.10},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicRouteTicks}, 0.85 * 0.85},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicRouteIncome}, 1.10},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicRaidLoss}, 0.81 * 0.85},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicMoraleCap}, 0.05},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicExpeditionTicks}, 0.90 * 0.90},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicWonderBuildTicks}, 0.75},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicFestivalTicks}, 1.25},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicGiftOpinion}, 1.50},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicAllianceBonus}, 1.25},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicUpgradeCost}, 0.85},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicCampaignTicks}, 0.85},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicSoldierStorage}, 1.20},
		{TechEffectKey{Kind: EffectMechanic, Target: MechanicCampaignReward}, 1.20},
	} {
		if got := terms[tc.key]; !near(got, tc.want) {
			t.Errorf("%s %s: every tech together comes to %v, want %v", tc.key.Kind, tc.key.Target, got, tc.want)
		}
	}
}

// TestFlatOutputIsAFirstSource: a flat output is for a tech that is the
// first source of a resource, so it stays rare. Today two techs have one.
func TestFlatOutputIsAFirstSource(t *testing.T) {
	got := map[string]string{}
	for _, tech := range Technologies() {
		for _, e := range tech.Effects {
			if e.Kind == EffectFlatOutput {
				got[tech.Key] = e.Target
			}
		}
	}
	want := map[string]string{"steel_forging": "steel", "satellite_tech": "data"}
	if len(got) != len(want) {
		t.Errorf("techs with a flat output: %v, want %v", got, want)
	}
	for k, res := range want {
		if got[k] != res {
			t.Errorf("%s makes %q flat, want %q", k, got[k], res)
		}
	}
}

// TestFeatureLocks: the nine locks, which of them are live today and which
// wait for their tech, and what a refusal says.
func TestFeatureLocks(t *testing.T) {
	techs := TechByKey()
	isTech := func(k string) bool { _, ok := techs[k]; return ok }
	locks := FeatureLocks()
	seen := map[string]bool{}
	for _, l := range locks {
		if l.Key == "" || l.Name == "" || l.Tech == "" || l.Needs == "" || l.Then == "" || l.Opens == "" || seen[l.Key] {
			t.Errorf("lock %+v is incomplete or listed twice", l)
		}
		seen[l.Key] = true
	}
	keys := func(ls []FeatureLockDef) string {
		var out []string
		for _, l := range ls {
			out = append(out, l.Key)
		}
		return strings.Join(out, " ")
	}
	live, waiting := LiveFeatureLocks(locks, isTech)
	if got, want := keys(live), "trade_routes campaigns expeditions diplomacy festivals naval_expedition black_market route_rail_freight"; got != want {
		t.Errorf("live locks: %s; want %s", got, want)
	}
	if got, want := keys(waiting), "route_warp_commerce"; got != want {
		t.Errorf("locks waiting for their tech: %s; want %s", got, want)
	}
	if got, want := FeatureLockByKey()[FeatureCampaigns].Refusal("Military Tactics"), "Campaigns need Military Tactics first. Research it to send one."; got != want {
		t.Errorf("refusal %q, want %q", got, want)
	}
	if got, want := FeatureLockByKey()[FeatureBlackMarket].Refusal("Mercantilism"), "The black market needs Mercantilism first. Research it to deal there."; got != want {
		t.Errorf("refusal %q, want %q", got, want)
	}
	// A live lock's tech is no later than the command it opens: the routes
	// it names exist from the tech's age or later.
	ages := AgePositions(AgeOrder())
	for _, r := range BaseTradeRoutes() {
		f := RouteFeature(r.Key)
		if f == "" {
			continue
		}
		if tech, ok := techs[FeatureLockByKey()[f].Tech]; ok && ages[tech.Age] > ages[r.MinAge] {
			t.Errorf("%s opens in the %s, before the %s that unlocks it can be researched (%s)", r.Name, r.MinAge, tech.Name, tech.Age)
		}
	}
	if got := FeaturesOpenedBy("military_tactics"); len(got) != 1 || got[0].Key != FeatureCampaigns {
		t.Errorf("Military Tactics opens %v, want campaigns", got)
	}
}

// TestTechChecks: every tech's promises are there as data, in the atom form
// and the check form, and no tech promises nothing.
func TestTechChecks(t *testing.T) {
	all := AllTechChecks()
	if len(all) != len(Technologies()) {
		t.Fatalf("%d techs have checks, of %d", len(all), len(Technologies()))
	}
	for _, tech := range Technologies() {
		checks := all[tech.Key]
		if len(checks) == 0 {
			t.Errorf("%s promises nothing: no effect, no building, no command", tech.Key)
		}
		for _, c := range checks {
			if c.Atom == "" || c.Measure == "" || c.Op == "" {
				t.Errorf("%s: incomplete check %+v", tech.Key, c)
			}
		}
	}
	form := func(key string) string {
		var atoms, checks []string
		for _, c := range all[key] {
			atoms = append(atoms, c.Atom)
			checks = append(checks, c.String())
		}
		return strings.Join(atoms, "; ") + " | " + strings.Join(checks, "; ")
	}
	for key, want := range map[string]string{
		"tool_making":                 "out:food:10; out:wood:10; p:gather_amount:2 | rate(food) x1.10; rate(wood) x1.10; gather_amount +2",
		"stoneworking":                "out:stone:10; keystone:great_monolith | rate(stone) x1.10; can_build(great_monolith) 0>1",
		"currency":                    "out:gold:10; p:market_fee:-0.03 | rate(gold) x1.10; market_fee -0.03",
		"military_tactics":            "mil:15; unlock:barracks; feature:campaigns | military_power +0.15; can_build(barracks) 0>1; allowed(campaigns) 0>1",
		"calendar":                    "out:faith:10; unlock:altar; keystone:stonehenge | rate(faith) x1.10; can_build(altar) 0>1; can_build(stonehenge) 0>1",
		"the_wheel":                   "time:5; feature:trade_routes | build_ticks x0.95; allowed(trade_routes) 0>1",
		"boatbuilding":                "out:food:5; p:route_income:0.1 | rate(food) x1.05; route_income x1.10",
		"priesthood":                  "p:morale_cap:0.05 | morale_cap +0.05",
		"exploration":                 "feature:expeditions | allowed(expeditions) 0>1",
		"envoys":                      "feature:diplomacy | allowed(diplomacy) 0>1",
		"drama":                       "feature:festivals | allowed(festivals) 0>1",
		"scholasticism":               "research:6 | research_ticks x0.94",
		"guilds":                      "cost:4; time:8 | build_price x0.96; build_ticks x0.92",
		"fortification":               "p:raid_loss:-0.15; mil:10 | raid_loss x0.85; military_power +0.1",
		"steel_forging":               "first:steel; out:iron:8 | rate(steel) 0>pos; rate(iron) x1.08",
		"road_building":               "out:gold:8; p:route_ticks:-0.15 | rate(gold) x1.08; route_ticks x0.85",
		"railroads":                   "p:route_ticks:-0.15; feature:route_rail_freight | route_ticks x0.85; allowed(route_rail_freight) 0>1",
		"mass_production":             "time:8; keystone:eiffel_tower | build_ticks x0.92; can_build(eiffel_tower) 0>1",
		"industrialization":           "all:5; keystone:crystal_palace | rate(all) x1.05; can_build(crystal_palace) 0>1",
		"mercantilism":                "unlock:harbor; feature:black_market | can_build(harbor) 0>1; allowed(black_market) 0>1",
		"architecture":                "cost:3; p:wonder_build_ticks:-0.25 | build_price x0.97; wonder_build_ticks x0.75",
		"baroque_arts":                "p:festival_ticks:0.25 | festival_ticks x1.25",
		"embassies":                   "p:gift_opinion:0.5; unlock:embassy | gift_opinion x1.50; can_build(embassy) 0>1",
		"steam_power":                 "out:steel:6; out:coal:6; unlock:coal_plant | rate(steel) x1.06; rate(coal) x1.06; can_build(coal_plant) 0>1",
		"general_staff":               "p:campaign_ticks:-0.15 | campaign_ticks x0.85",
		"aviation":                    "p:expedition_ticks:-0.1 | expedition_ticks x0.90",
		"wire_transfers":              "p:market_fee:-0.02 | market_fee -0.02",
		"concert_of_nations":          "p:alliance_bonus:0.25; unlock:grand_embassy | alliance_bonus x1.25; can_build(grand_embassy) 0>1",
		"interchangeable_parts":       "cost:4; p:upgrade_cost:-0.15 | build_price x0.96; upgrade_cost x0.85",
		"geographic_societies":        "unlock:geographic_society | can_build(geographic_society) 0>1",
		"power_distribution":          "out:electricity:5; unlock:power_generator; keystone:hoover_dam | rate(electricity) x1.05; can_build(power_generator) 0>1; can_build(hoover_dam) 0>1",
		"military_industrial_complex": "p:soldier_storage:0.2; p:campaign_reward:0.2 | soldier_storage x1.20; campaign_reward x1.20",
		"holography":                  "unlock:holographic_theater | can_build(holographic_theater) 0>1",
		"pottery":                     "store:10 | storage x1.10",
		"fire_mastery":                "out:food:10; house:5 | rate(food) x1.10; housing x1.05",
		"civil_engineering":           "cost:3 | build_price x0.97",
		"printing_press":              "out:knowledge:6; research:3 | rate(knowledge) x1.06; research_ticks x0.97",
		"chronometry":                 "speed:5 | tick_speed +0.05",
		"navigation":                  "exp:10; feature:naval_expedition | expedition_reward +0.1; allowed(naval_expedition) 0>1",
	} {
		if got := form(key); got != want {
			t.Errorf("%s:\n got %s\nwant %s", key, got, want)
		}
	}
}
