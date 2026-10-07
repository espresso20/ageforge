package rules

import (
	"math"
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The core set must say exactly what package config says, for every key
// config knows and for keys it does not. These tests are the proof that
// moving a lookup from config to the set changes no number.

var probeAges = append(config.AgeOrder(), "", "not_an_age")

// sameLevel compares two deal price levels. A flow resource's level sums a
// building's price over a map (config.FlowDealLevels), so two builds of the
// table can differ in the last bit; a set works its table out once and
// every engine on it reads the same one.
func sameLevel(a, b float64) bool {
	return a == b || math.Abs(a-b) <= 1e-12*math.Max(math.Abs(a), math.Abs(b))
}

func probeEras() []string {
	keys := []string{"", "not_an_era", config.LastPassageKey}
	for _, e := range config.Epochs() {
		keys = append(keys, e.Key)
	}
	return keys
}

func TestCoreMatchesConfigTables(t *testing.T) {
	s := Core()
	check := func(name string, got, want any) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s differs from config", name)
		}
	}
	check("Ages", s.Ages(), config.Ages())
	check("AgeKeys", s.AgeKeys(), config.AgeOrder())
	check("Eras", s.Eras(), config.Epochs())
	check("Buildings", s.Buildings(), config.BaseBuildings())
	check("BuildingMap", s.BuildingMap(), config.BuildingByKey())
	check("Techs", s.Techs(), config.Technologies())
	check("TechMap", s.TechMap(), config.TechByKey())
	check("TechLanes", s.TechLanes(), config.TechLanes())
	check("Resources", s.Resources(), config.BaseResources())
	check("ResourceMap", s.ResourceMap(), config.ResourceByKey())
	check("Milestones", s.Milestones(), config.Milestones())
	check("MilestoneChains", s.MilestoneChains(), config.MilestoneChains())
	check("MilestoneTitles", s.MilestoneTitles(), config.MilestoneTitles())
	check("Events", s.Events(), config.RandomEvents())
	check("EraEvents", s.EraEvents(), config.EpochExclusiveEvents())
	check("EventMap", s.EventMap(), config.EventByKey())
	check("GoodEraEvents", s.GoodEraEvents(), config.GoodEpochEvents())
	check("ChallengingEraEvents", s.ChallengingEraEvents(), config.ChallengingEpochEvents())
	check("Factions", s.Factions(), config.BaseFactions())
	check("TradeRoutes", s.TradeRoutes(), config.BaseTradeRoutes())
	check("ExchangeRates", s.ExchangeRates(), config.BaseExchangeRates())
	check("WorkerDomains", s.WorkerDomains(), config.WorkerDomains())
	check("PrestigeUpgrades", s.PrestigeUpgrades(), config.PrestigeUpgrades())
	check("ShopUpgrades", s.ShopUpgrades(), config.ActivePrestigeUpgrades())
	check("LegacyKit", s.LegacyKit(), config.LegacyKit())

	for key, want := range config.AgeByKey() {
		if got, ok := s.Age(key); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Age(%q) differs from config", key)
		}
	}
	for key, want := range config.MilestoneByKey() {
		if got, ok := s.Milestone(key); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Milestone(%q) differs from config", key)
		}
	}
	for key, want := range config.EpochEventByKey() {
		if got, ok := s.EraEvent(key); !ok || got != want {
			t.Errorf("EraEvent(%q) differs from config", key)
		}
	}
	for key, want := range config.FactionByKey() {
		if got, ok := s.Faction(key); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Faction(%q) differs from config", key)
		}
	}
	for key, want := range config.PrestigeUpgradeByKey() {
		if got, ok := s.PrestigeUpgrade(key); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("PrestigeUpgrade(%q) differs from config", key)
		}
	}
	for age, want := range config.TechsByAge() {
		if !reflect.DeepEqual(s.TechsOf(age), want) {
			t.Errorf("TechsOf(%q) differs from config", age)
		}
	}
	for _, miss := range []string{"", "nope"} {
		if _, ok := s.Age(miss); ok {
			t.Errorf("Age(%q) found", miss)
		}
		if _, ok := s.Building(miss); ok {
			t.Errorf("Building(%q) found", miss)
		}
		if _, ok := s.Tech(miss); ok {
			t.Errorf("Tech(%q) found", miss)
		}
		if _, ok := s.Resource(miss); ok {
			t.Errorf("Resource(%q) found", miss)
		}
	}
}

func TestCoreMatchesConfigPerAge(t *testing.T) {
	s := Core()
	order := config.AgeOrder()
	byKey := config.AgeByKey()
	resources := []string{"", "nope"}
	for _, r := range config.BaseResources() {
		resources = append(resources, r.Key)
	}
	for _, age := range probeAges {
		i, ok := s.Index(age)
		if def, known := byKey[age]; ok != known || (known && (i != def.Order || order[i] != age)) {
			t.Errorf("Index(%q) = %d, %v: not the age's place and its Order", age, i, ok)
		}
		wantNext := ""
		if ok && i+1 < len(order) {
			wantNext = order[i+1]
		}
		if got := s.Next(age); got != wantNext {
			t.Errorf("Next(%q) = %q, want %q", age, got, wantNext)
		}
		if got, want := s.EraOf(age), config.EpochForAge(age); got != want {
			t.Errorf("EraOf(%q) = %q, config %q", age, got, want)
		}
		if got, want := s.Target(age), config.AgeTargets[age]; got != want {
			t.Errorf("Target(%q) = %v, config %v", age, got, want)
		}
		if got, want := s.TargetTicks(age), config.AgeTargetTicks(age); got != want {
			t.Errorf("TargetTicks(%q) = %v, config %v", age, got, want)
		}
		for _, ticks := range []int{-3, 0, 1, 7, 30, 144, 1000, 99999} {
			if got, want := s.StretchTicks(age, ticks), config.StretchTicks(age, ticks); got != want {
				t.Errorf("StretchTicks(%q, %d) = %d, config %d", age, ticks, got, want)
			}
		}
		if got, want := s.DepthWeight(age), config.DepthWeight(age); got != want {
			t.Errorf("DepthWeight(%q) = %d, config %d", age, got, want)
		}
		if got, want := s.DepthPoints(age), config.DepthPoints(age); got != want {
			t.Errorf("DepthPoints(%q) = %d, config %d", age, got, want)
		}
		gotH, okH := s.Harbinger(age)
		wantH, okW := config.HarbingerFor(age)
		if okH != okW || gotH != wantH {
			t.Errorf("Harbinger(%q) differs from config", age)
		}
		gotA, okA := s.Awakening(age)
		wantA, okAW := config.AwakeningForAge(age)
		if okA != okAW || !reflect.DeepEqual(gotA, wantA) {
			t.Errorf("Awakening(%q) differs from config", age)
		}
		if got, want := s.AgeEntryCosts(age), config.AgeEntryCosts(age); !reflect.DeepEqual(got, want) {
			t.Errorf("AgeEntryCosts(%q) differs from config", age)
		}
		for _, d := range config.WorkerDomains() {
			got, okG := s.WorkerClass(d, age)
			want, okC := config.WorkerClassByDomainAndAge(d, age)
			if okG != okC || got != want {
				t.Errorf("WorkerClass(%q, %q) = %+v, config %+v", d, age, got, want)
			}
		}
		if _, ok := s.WorkerClass("not_a_domain", age); ok {
			t.Errorf("WorkerClass found a class for an unknown domain in %q", age)
		}
	}
	// The prices rebuild config's building table on every call, so they are
	// read once per age here and compared cell by cell.
	levels := config.PriceLevelsByAge(config.BaseBuildings())
	flowLevels := config.FlowDealLevels(config.BaseBuildings(), config.AgePositions(order))
	flow := config.Incomes(config.BaseBuildings(), config.Technologies(), order, config.IsFlowResource)
	typical := config.Incomes(config.BaseBuildings(), config.Technologies(), order, config.AnyResource)
	listed := config.ExchangeRateByKey()
	pos := config.AgePositions(order)
	for _, age := range probeAges {
		if !reflect.DeepEqual(s.PriceLevels(age), levels[age]) {
			t.Errorf("PriceLevels(%q) differs from config", age)
		}
		if got, want := s.PricedResources(age), config.PricedResourcesAt(levels[age]); len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
			t.Errorf("PricedResources(%q) = %v, config %v", age, got, want)
		}
		if got, want := s.MarketPairs(age), config.MarketPairsAt(age, config.BaseExchangeRates(), pos, levels[age]); !reflect.DeepEqual(got, want) {
			t.Errorf("MarketPairs(%q) differs from config", age)
		}
		for _, r := range resources {
			if got, want := s.FlowIncome(r, age), flow[age][r]; got != want {
				t.Errorf("FlowIncome(%q, %q) = %v, config %v", r, age, got, want)
			}
			if got, want := s.TypicalIncome(r, age), typical[age][r]; got != want {
				t.Errorf("TypicalIncome(%q, %q) = %v, config %v", r, age, got, want)
			}
			if got, want := s.DealPriceLevel(r, age), config.DealPriceLevelAt(r, levels[age], flowLevels[age]); !sameLevel(got, want) {
				t.Errorf("DealPriceLevel(%q, %q) = %v, config %v", r, age, got, want)
			}
			for _, to := range resources {
				got, okG := s.MarketRate(r, to, age)
				want, okW := config.MarketRateAt(r, to, listed, levels[age])
				if got != want || okG != okW {
					t.Errorf("MarketRate(%q, %q, %q) = %v, %v; config %v, %v", r, to, age, got, okG, want, okW)
				}
				got, okG = s.MarketOffers(r, to, age)
				want, okW = config.MarketOffersAt(r, to, age, listed, pos, levels[age])
				if got != want || okG != okW {
					t.Errorf("MarketOffers(%q, %q, %q) = %v, %v; config %v, %v", r, to, age, got, okG, want, okW)
				}
			}
		}
		for _, x := range config.BaseExchangeRates() {
			if got, want := s.ExchangeRate(x, age), config.ExchangeRateAt(x, levels[age]); got != want {
				t.Errorf("ExchangeRate(%s:%s, %q) = %v, config %v", x.From, x.To, age, got, want)
			}
		}
	}
}

// TestPackageLookupsMatchTheirTables: the package-level config lookups, which
// now rebuild their tables on every call, still give what the tables hold.
func TestPackageLookupsMatchTheirTables(t *testing.T) {
	s := Core()
	for _, age := range []string{"primitive_age", "iron_age", "modern_age", "transcendent_age", "not_an_age"} {
		for _, r := range []string{"food", "faith", "wood", "iron", "steel", "nope"} {
			if got, want := config.FlowIncome(r, age), s.FlowIncome(r, age); got != want {
				t.Errorf("config.FlowIncome(%q, %q) = %v, set %v", r, age, got, want)
			}
			if got, want := config.TypicalIncome(r, age), s.TypicalIncome(r, age); got != want {
				t.Errorf("config.TypicalIncome(%q, %q) = %v, set %v", r, age, got, want)
			}
			if got, want := config.DealPriceLevel(r, age), s.DealPriceLevel(r, age); !sameLevel(got, want) {
				t.Errorf("config.DealPriceLevel(%q, %q) = %v, set %v", r, age, got, want)
			}
			got, okG := config.MarketRate(r, "gold", age)
			want, okW := s.MarketRate(r, "gold", age)
			if got != want || okG != okW {
				t.Errorf("config.MarketRate(%q, gold, %q) = %v, %v; set %v, %v", r, age, got, okG, want, okW)
			}
			got, okG = config.MarketOffers(r, "gold", age)
			want, okW = s.MarketOffers(r, "gold", age)
			if got != want || okG != okW {
				t.Errorf("config.MarketOffers(%q, gold, %q) = %v, %v; set %v, %v", r, age, got, okG, want, okW)
			}
		}
		if !reflect.DeepEqual(config.PriceLevels(age), s.PriceLevels(age)) {
			t.Errorf("config.PriceLevels(%q) differs from the set", age)
		}
		if got, want := config.PricedResources(age), s.PricedResources(age); len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
			t.Errorf("config.PricedResources(%q) = %v, set %v", age, got, want)
		}
		if !reflect.DeepEqual(config.MarketPairs(age), s.MarketPairs(age)) {
			t.Errorf("config.MarketPairs(%q) differs from the set", age)
		}
	}
}

func TestCoreMatchesConfigPerEra(t *testing.T) {
	s := Core()
	byKey := config.EpochByKey()
	ages := config.AgeByKey()
	for _, era := range probeEras() {
		got, ok := s.Era(era)
		want, known := byKey[era]
		if ok != known || !reflect.DeepEqual(got, want) {
			t.Errorf("Era(%q) differs from config", era)
		}
		gotNext, okN := s.NextEra(era)
		wantNext, okW := config.NextEpoch(era)
		if okN != okW || !reflect.DeepEqual(gotNext, wantNext) {
			t.Errorf("NextEra(%q) differs from config", era)
		}
		if got, want := s.IsFinalEra(era), config.IsFinalEpoch(era); got != want {
			t.Errorf("IsFinalEra(%q) = %v, config %v", era, got, want)
		}
		if got, want := s.CatastropheAllowed(era), config.CatastropheAllowed(era); got != want {
			t.Errorf("CatastropheAllowed(%q) = %v, config %v", era, got, want)
		}
		if got, want := s.FateAllowed(era), config.FateAllowed(era); got != want {
			t.Errorf("FateAllowed(%q) = %v, config %v", era, got, want)
		}
		gn, gf := s.Catastrophe(era)
		wn, wf := config.CatastropheInfo(era)
		if gn != wn || gf != wf {
			t.Errorf("Catastrophe(%q) = %q, config %q", era, gn, wn)
		}
		if got, want := s.LegacyBonus(era), config.LegacyBonusForEpoch(era); len(got) != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
			t.Errorf("LegacyBonus(%q) = %v, config %v", era, got, want)
		}
		first, okF := s.EraFirstAge(era)
		if known && len(want.Ages) > 0 {
			if !okF || first != ages[want.Ages[0]].Order {
				t.Errorf("EraFirstAge(%q) = %d, %v; want the Order of %s", era, first, okF, want.Ages[0])
			}
		} else if okF {
			t.Errorf("EraFirstAge(%q) found", era)
		}
	}
	gn, gf := s.LastPassage()
	wn, wf := config.LastPassageInfo()
	if gn != wn || gf != wf {
		t.Errorf("LastPassage = %q, config %q", gn, wn)
	}
}

func TestCoreMatchesConfigPerBuilding(t *testing.T) {
	s := Core()
	ages := config.AgeByKey()
	wonders := map[string]string{}
	for key, want := range config.BuildingByKey() {
		got, ok := s.Building(key)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Building(%q) differs from config", key)
		}
		if want.Category == "wonder" {
			for _, a := range ages {
				for _, u := range a.UnlockBuildings {
					if u == key && wonders[a.Key] == "" {
						wonders[a.Key] = key
					}
				}
			}
		}
		// The age after the building's own is the one the engine asks about
		// (config rebuilds every building per call, so one age each).
		age := s.Next(want.RequiredAge)
		next, okN := s.NextTier(want.LineageKey, want.LineageTier, age)
		wantNext := config.BuildingNextTierForAge(want.LineageKey, want.LineageTier, age)
		if okN != (wantNext != nil) || (okN && !reflect.DeepEqual(next, *wantNext)) {
			t.Errorf("NextTier(%q, %d, %q) differs from config", want.LineageKey, want.LineageTier, age)
		}
	}
	for _, age := range probeAges {
		if got := s.Wonder(age); got != wonders[age] {
			t.Errorf("Wonder(%q) = %q, want %q", age, got, wonders[age])
		}
	}
	for key, want := range config.TechByKey() {
		if got, ok := s.Tech(key); !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("Tech(%q) differs from config", key)
		}
	}
	for key, want := range config.ResourceByKey() {
		if got, ok := s.Resource(key); !ok || got != want {
			t.Errorf("Resource(%q) differs from config", key)
		}
		if got, want := s.ResourceLabel(key), config.ResourceLabel(key); got != want {
			t.Errorf("ResourceLabel(%q) = %q, config %q", key, got, want)
		}
	}
	if got, want := s.ResourceLabel("not_a_resource"), config.ResourceLabel("not_a_resource"); got != want {
		t.Errorf("ResourceLabel of an unknown key = %q, config %q", got, want)
	}
}

func TestNames(t *testing.T) {
	s := Core()
	for _, tc := range []struct {
		kind Kind
		key  string
		want string
	}{
		{KindAge, "iron_age", config.AgeByKey()["iron_age"].Name},
		{KindBuilding, "hut", config.BuildingByKey()["hut"].Name},
		{KindResource, "iron_ore", config.ResourceByKey()["iron_ore"].Name},
		{KindCiv, "merchant_guild", config.FactionByKey()["merchant_guild"].Name},
		{KindPrestigeUpgrade, config.LegacyPlan, config.PrestigeUpgradeByKey()[config.LegacyPlan].Name},
		{KindAge, "lost_age", "Lost age"},
		{KindBuilding, "", ""},
		{numKinds, "any_key", "Any key"},
	} {
		if got := s.Name(tc.kind, tc.key); got != tc.want {
			t.Errorf("Name(%d, %q) = %q, want %q", tc.kind, tc.key, got, tc.want)
		}
	}
	for _, d := range config.Technologies() {
		if got := s.Name(KindTech, d.Key); got != d.Name {
			t.Errorf("Name(tech %q) = %q, want %q", d.Key, got, d.Name)
		}
	}
	for _, d := range config.BaseTradeRoutes() {
		if got := s.Name(KindRoute, d.Key); got != d.Name {
			t.Errorf("Name(route %q) = %q, want %q", d.Key, got, d.Name)
		}
	}
}

func TestCounts(t *testing.T) {
	c := Core().Counts()
	for q, want := range map[string]int{
		CountAges: len(config.Ages()), CountBuildings: len(config.BaseBuildings()), CountTechs: len(config.Technologies()),
		CountMilestones: len(config.Milestones()), CountChains: len(config.MilestoneChains()),
		CountResources: len(config.BaseResources()), CountEras: len(config.Epochs()),
		CountDomains: len(config.WorkerDomains()), CountShopItems: len(config.ActivePrestigeUpgrades()),
		CountCivs: len(config.BaseFactions()), CountRoutes: len(config.BaseTradeRoutes()),
	} {
		if c[q] != want {
			t.Errorf("Counts[%q] = %d, want %d", q, c[q], want)
		}
	}
	if sum := c[CountWonders] + c[CountStorage] + c[CountMonuments] + c[CountLineageBld] + c[CountStandalone]; sum != c[CountBuildings] {
		t.Errorf("the building groups add up to %d of %d buildings", sum, c[CountBuildings])
	}
	if c[CountWonders] != c[CountAges] {
		t.Errorf("%d wonders for %d ages", c[CountWonders], c[CountAges])
	}
	c[CountAges] = -1
	if Core().Counts()[CountAges] == -1 {
		t.Error("Counts handed out the set's own map")
	}
}

// TestCompileKeepsNothingOfItsSource: a set is built from copies, so a source
// changed after Compile, or compiled twice, cannot reach into a set.
func TestCompileKeepsNothingOfItsSource(t *testing.T) {
	src := FromConfig()
	a := Compile(src)
	first := src.Ages[0].Key
	src.Ages[0].Key = "changed"
	src.Buildings = src.Buildings[:1]
	src.Techs[0].Name = "changed"
	delete(src.Targets, first)
	src.WorkerDomains[0] = "changed"
	if a.AgeKeys()[0] != first || len(a.Buildings()) != len(config.BaseBuildings()) ||
		a.Techs()[0].Name == "changed" || a.Target(first) != config.AgeTargets[first] || a.WorkerDomains()[0] == "changed" {
		t.Error("a change to the source after Compile reached the set")
	}
	if _, ok := config.AgeTargets[first]; !ok {
		t.Fatal("deleting from a Source's targets reached config.AgeTargets")
	}
	// Two compiles of the same tables are the same set. The flow levels are
	// left out: see sameLevel.
	x, y := *Compile(FromConfig()), *Core()
	x.flowLevels, y.flowLevels = nil, nil
	if !reflect.DeepEqual(x, y) {
		t.Error("two compiles of the same tables differ")
	}
}

// TestSetHandsOutCopies: every slice and map a set returns is the caller's
// own, as package config's were. A caller may sort it, write to it or grow
// it without reaching the set.
func TestSetHandsOutCopies(t *testing.T) {
	s := Compile(FromConfig())
	before := s.Digest()
	age := s.AgeKeys()[3]
	era := s.EraOf(age)
	// Scribble over everything a method returns: zero every slice element
	// and add a key to every map.
	for name, got := range map[string]any{
		"Ages": s.Ages(), "AgeKeys": s.AgeKeys(), "Indexes": s.Indexes(), "Eras": s.Eras(),
		"LegacyBonus": s.LegacyBonus(era),
		"Buildings":   s.Buildings(), "BuildingMap": s.BuildingMap(), "AgeEntryCosts": s.AgeEntryCosts(age),
		"Techs": s.Techs(), "TechMap": s.TechMap(), "TechsOf": s.TechsOf(age), "TechLanes": s.TechLanes(),
		"Resources": s.Resources(), "ResourceMap": s.ResourceMap(),
		"Milestones": s.Milestones(), "MilestoneChains": s.MilestoneChains(), "MilestoneTitles": s.MilestoneTitles(),
		"Events": s.Events(), "EraEvents": s.EraEvents(), "EventMap": s.EventMap(),
		"GoodEraEvents": s.GoodEraEvents(), "ChallengingEraEvents": s.ChallengingEraEvents(),
		"Factions": s.Factions(), "TradeRoutes": s.TradeRoutes(),
		"ExchangeRates": s.ExchangeRates(),
		"WorkerDomains": s.WorkerDomains(), "PrestigeUpgrades": s.PrestigeUpgrades(),
		"ShopUpgrades": s.ShopUpgrades(), "LegacyKit": s.LegacyKit(),
		"PriceLevels": s.PriceLevels(age), "PricedResources": s.PricedResources(age),
		"MarketPairs": s.MarketPairs(age), "Counts": s.Counts(),
	} {
		v := reflect.ValueOf(got)
		if v.Len() == 0 {
			t.Errorf("%s is empty: nothing to check", name)
			continue
		}
		switch v.Kind() {
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				v.Index(i).SetZero()
			}
		case reflect.Map:
			v.SetMapIndex(reflect.ValueOf("scratch"), reflect.Zero(v.Type().Elem()))
			for _, k := range v.MapKeys() {
				v.SetMapIndex(k, reflect.Zero(v.Type().Elem()))
			}
		default:
			t.Fatalf("%s returns a %s", name, v.Kind())
		}
		if got := s.Digest(); got != before {
			t.Fatalf("%s handed out the set's own %s: writing to it changed the set", name, v.Kind())
		}
	}
	if s.Digest() != s.Digest() {
		t.Error("a set's digest is not stable")
	}
}

// TestASecondSet: a set compiled from other tables answers from its own, and
// the core set does not move.
func TestASecondSet(t *testing.T) {
	src := FromConfig()
	last := src.Ages[len(src.Ages)-1].Key
	src.Ages = src.Ages[:len(src.Ages)-1]
	src.Techs = append(src.Techs, config.TechDef{Key: "second_set_tech", Name: "Second Set Tech", Age: src.Ages[0].Key})
	s := Compile(src)
	if _, ok := s.Index(last); ok {
		t.Errorf("the shorter set still places %s", last)
	}
	if got := s.Next(src.Ages[len(src.Ages)-1].Key); got != "" {
		t.Errorf("the shorter set's last age is followed by %q", got)
	}
	if _, ok := s.Tech("second_set_tech"); !ok || s.Name(KindTech, "second_set_tech") != "Second Set Tech" {
		t.Error("the second set does not know its own tech")
	}
	if got := s.Counts()[CountAges]; got != len(config.Ages())-1 {
		t.Errorf("the shorter set counts %d ages", got)
	}
	if _, ok := Core().Tech("second_set_tech"); ok {
		t.Error("the core set learned the second set's tech")
	}
	// A tech's kind comes from the set's own wonders: give this set's Stone
	// Age wonder a keystone, and it and what it stands on are required here
	// and nowhere else.
	keyed := FromConfig()
	var wonder string
	for i, b := range keyed.Buildings {
		if b.Category == "wonder" && b.RequiredAge == "stone_age" {
			keyed.Buildings[i].RequiredTech, wonder = "pottery", b.Key
		}
	}
	k := Compile(keyed)
	if wonder == "" || k.TechKind("pottery") != config.TechKeystone || k.TechKind("fire_mastery") != config.TechSpine || k.TechKind("chronometry") != config.TechOptional {
		t.Errorf("with %q needing pottery: pottery is %q, fire_mastery %q, chronometry %q; want keystone, spine, optional",
			wonder, k.TechKind("pottery"), k.TechKind("fire_mastery"), k.TechKind("chronometry"))
	}
	if got := Core().TechKind("pottery"); got != config.TechOptional {
		t.Errorf("the core set calls pottery %q: none of its own wonders requires it", got)
	}
	// The same goes for which wonder a tech is the keystone of, and which
	// tech an age's wonder waits for.
	if k.KeystoneOf("pottery") != wonder || k.Keystone("stone_age") != "pottery" || k.KeystoneOf("fire_mastery") != "" {
		t.Errorf("with %q needing pottery: pottery is the keystone of %q, the Stone Age's keystone is %q, fire_mastery is the keystone of %q",
			wonder, k.KeystoneOf("pottery"), k.Keystone("stone_age"), k.KeystoneOf("fire_mastery"))
	}
	if got := Core().KeystoneOf("pottery"); got != "" {
		t.Errorf("the core set calls pottery the keystone of %q", got)
	}
	if Core().Keystone("stone_age") != "stoneworking" || Core().KeystoneOf("stoneworking") != wonder || Core().Keystone("primitive_age") != "" || Core().Keystone("no_such_age") != "" {
		t.Errorf("the core set: the Stone Age's keystone is %q, Stoneworking is the keystone of %q, the Primitive Age's is %q",
			Core().Keystone("stone_age"), Core().KeystoneOf("stoneworking"), Core().Keystone("primitive_age"))
	}
	if got := Core().TechKind("no_such_tech"); got != "" {
		t.Errorf("an unknown tech has the kind %q, want none", got)
	}
	if _, ok := Core().Index(last); !ok {
		t.Errorf("the core set lost %s", last)
	}
}
