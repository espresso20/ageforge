package config

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildKeySet collects all keys from a slice using a getter function
func buildKeySet[T any](items []T, getKey func(T) string) map[string]bool {
	m := make(map[string]bool)
	for _, item := range items {
		m[getKey(item)] = true
	}
	return m
}

// suggest finds the closest match to 'input' in 'valid' keys (edit distance <= 3)
func suggest(input string, valid map[string]bool) string {
	best, bestDist := "", 4
	for k := range valid {
		d := editDist(input, k)
		if d < bestDist {
			bestDist = d
			best = k
		}
	}
	return best
}

func suggestFromMap[T any](input string, valid map[string]T) string {
	m := make(map[string]bool, len(valid))
	for k := range valid {
		m[k] = true
	}
	return suggest(input, m)
}

func editDist(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = curr
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// hint builds a "did you mean X?" string, or empty if no close match
func hint(input string, valid map[string]bool) string {
	if s := suggest(input, valid); s != "" {
		return fmt.Sprintf(" (did you mean %q?)", s)
	}
	return ""
}

func hintFromMap[T any](input string, valid map[string]T) string {
	m := make(map[string]bool, len(valid))
	for k := range valid {
		m[k] = true
	}
	return hint(input, m)
}

// validList returns a sorted comma-separated list of keys, truncated if > max
func validList(keys map[string]bool, max int) string {
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	if len(sorted) > max {
		sorted = append(sorted[:max], fmt.Sprintf("... (%d more)", len(sorted)-max))
	}
	return strings.Join(sorted, ", ")
}

// ---------------------------------------------------------------------------
// Age config validation
// ---------------------------------------------------------------------------

func TestConfig_AgeKeysExist(t *testing.T) {
	ages := Ages()
	resourceKeys := ResourceByKey()
	buildingKeys := BuildingByKey()

	for _, age := range ages {
		for res := range age.ResourceReqs {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in age advancement requirements\n"+
					"  File:     config/ages.go\n"+
					"  Age:      %q (%s)\n"+
					"  Field:    ResourceReqs\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					age.Key, age.Name, res, hintFromMap(res, resourceKeys))
			}
		}
		for bld := range age.BuildingReqs {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("\n"+
					"  Bad building key in age advancement requirements\n"+
					"  File:     config/ages.go\n"+
					"  Age:      %q (%s)\n"+
					"  Field:    BuildingReqs\n"+
					"  Got:      %q  <-- this building doesn't exist\n"+
					"  Fix:      Check config/buildings.go for valid building keys%s\n",
					age.Key, age.Name, bld, hintFromMap(bld, buildingKeys))
			}
		}
		for _, bld := range age.UnlockBuildings {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("\n"+
					"  Bad building key in age unlock list\n"+
					"  File:     config/ages.go\n"+
					"  Age:      %q (%s)\n"+
					"  Field:    UnlockBuildings\n"+
					"  Got:      %q  <-- this building doesn't exist\n"+
					"  Fix:      Either add this building to config/buildings.go, or fix the typo%s\n",
					age.Key, age.Name, bld, hintFromMap(bld, buildingKeys))
			}
		}
		for _, res := range age.UnlockResources {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in age unlock list\n"+
					"  File:     config/ages.go\n"+
					"  Age:      %q (%s)\n"+
					"  Field:    UnlockResources\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Either add this resource to config/resources.go, or fix the typo%s\n",
					age.Key, age.Name, res, hintFromMap(res, resourceKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Building config validation
// ---------------------------------------------------------------------------

func TestConfig_BuildingKeysExist(t *testing.T) {
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	techKeys := TechByKey()
	resourceKeys := ResourceByKey()

	for _, bld := range BaseBuildings() {
		if bld.RequiredAge != "" && !ageKeys[bld.RequiredAge] {
			t.Errorf("\n"+
				"  Bad age key in building definition\n"+
				"  File:     config/buildings.go\n"+
				"  Building: %q (%s)\n"+
				"  Field:    RequiredAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				bld.Key, bld.Name, bld.RequiredAge, hint(bld.RequiredAge, ageKeys))
		}
		if bld.RequiredTech != "" {
			if _, ok := techKeys[bld.RequiredTech]; !ok {
				t.Errorf("\n"+
					"  Bad tech key in building definition\n"+
					"  File:     config/buildings.go\n"+
					"  Building: %q (%s)\n"+
					"  Field:    RequiredTech\n"+
					"  Got:      %q  <-- this technology doesn't exist\n"+
					"  Fix:      Check config/research.go for valid tech keys%s\n",
					bld.Key, bld.Name, bld.RequiredTech, hintFromMap(bld.RequiredTech, techKeys))
			}
		}
		for res := range bld.BaseCost {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in building cost\n"+
					"  File:     config/buildings.go\n"+
					"  Building: %q (%s)\n"+
					"  Field:    BaseCost\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					bld.Key, bld.Name, res, hintFromMap(res, resourceKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Tech config validation
// ---------------------------------------------------------------------------

func TestConfig_TechKeysExist(t *testing.T) {
	techKeys := TechByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, tech := range Technologies() {
		if tech.Age != "" && !ageKeys[tech.Age] {
			t.Errorf("\n"+
				"  Bad age key in technology definition\n"+
				"  File:     config/research.go\n"+
				"  Tech:     %q (%s)\n"+
				"  Field:    Age\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				tech.Key, tech.Name, tech.Age, hint(tech.Age, ageKeys))
		}
		for _, prereq := range tech.Prerequisites {
			if _, ok := techKeys[prereq]; !ok {
				t.Errorf("\n"+
					"  Bad prerequisite in technology definition\n"+
					"  File:     config/research.go\n"+
					"  Tech:     %q (%s)\n"+
					"  Field:    Prerequisites\n"+
					"  Got:      %q  <-- this technology doesn't exist\n"+
					"  Fix:      Check config/research.go for valid tech keys%s\n",
					tech.Key, tech.Name, prereq, hintFromMap(prereq, techKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Milestone config validation
// ---------------------------------------------------------------------------

func TestConfig_MilestoneKeysExist(t *testing.T) {
	milestoneKeys := MilestoneByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	resourceKeys := ResourceByKey()
	buildingKeys := BuildingByKey()
	techKeys := TechByKey()

	for _, ms := range Milestones() {
		if ms.MinAge != "" && !ageKeys[ms.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in milestone definition\n"+
				"  File:     config/milestones.go\n"+
				"  Milestone: %q (%s)\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				ms.Key, ms.Name, ms.MinAge, hint(ms.MinAge, ageKeys))
		}
		for res := range ms.MinResources {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in milestone requirements\n"+
					"  File:     config/milestones.go\n"+
					"  Milestone: %q (%s)\n"+
					"  Field:    MinResources\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					ms.Key, ms.Name, res, hintFromMap(res, resourceKeys))
			}
		}
		for bld := range ms.MinBuildings {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("\n"+
					"  Bad building key in milestone requirements\n"+
					"  File:     config/milestones.go\n"+
					"  Milestone: %q (%s)\n"+
					"  Field:    MinBuildings\n"+
					"  Got:      %q  <-- this building doesn't exist\n"+
					"  Fix:      Check config/buildings.go for valid building keys%s\n",
					ms.Key, ms.Name, bld, hintFromMap(bld, buildingKeys))
			}
		}
		for _, tech := range ms.RequiredTechs {
			if _, ok := techKeys[tech]; !ok {
				t.Errorf("\n"+
					"  Bad tech key in milestone requirements\n"+
					"  File:     config/milestones.go\n"+
					"  Milestone: %q (%s)\n"+
					"  Field:    RequiredTechs\n"+
					"  Got:      %q  <-- this technology doesn't exist\n"+
					"  Fix:      Check config/research.go for valid tech keys%s\n",
					ms.Key, ms.Name, tech, hintFromMap(tech, techKeys))
			}
		}
	}

	for _, chain := range MilestoneChains() {
		for _, mk := range chain.MilestoneKeys {
			if _, ok := milestoneKeys[mk]; !ok {
				t.Errorf("\n"+
					"  Bad milestone key in chain definition\n"+
					"  File:     config/milestones.go (MilestoneChains)\n"+
					"  Chain:    %q (%s)\n"+
					"  Field:    MilestoneKeys\n"+
					"  Got:      %q  <-- this milestone doesn't exist\n"+
					"  Fix:      Check the Milestones() list for valid milestone keys%s\n",
					chain.Key, chain.Name, mk, hintFromMap(mk, milestoneKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Trade config validation
// ---------------------------------------------------------------------------

func TestConfig_TradeKeysExist(t *testing.T) {
	resourceKeys := ResourceByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	buildingKeys := BuildingByKey()

	for _, rate := range BaseExchangeRates() {
		if _, ok := resourceKeys[rate.From]; !ok {
			t.Errorf("\n"+
				"  Bad resource key in exchange rate\n"+
				"  File:     config/trade.go (BaseExchangeRates)\n"+
				"  Rate:     %s -> %s\n"+
				"  Field:    From\n"+
				"  Got:      %q  <-- this resource doesn't exist\n"+
				"  Fix:      Check config/resources.go for valid resource keys%s\n",
				rate.From, rate.To, rate.From, hintFromMap(rate.From, resourceKeys))
		}
		if _, ok := resourceKeys[rate.To]; !ok {
			t.Errorf("\n"+
				"  Bad resource key in exchange rate\n"+
				"  File:     config/trade.go (BaseExchangeRates)\n"+
				"  Rate:     %s -> %s\n"+
				"  Field:    To\n"+
				"  Got:      %q  <-- this resource doesn't exist\n"+
				"  Fix:      Check config/resources.go for valid resource keys%s\n",
				rate.From, rate.To, rate.To, hintFromMap(rate.To, resourceKeys))
		}
		if rate.MinAge != "" && !ageKeys[rate.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in exchange rate\n"+
				"  File:     config/trade.go (BaseExchangeRates)\n"+
				"  Rate:     %s -> %s\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				rate.From, rate.To, rate.MinAge, hint(rate.MinAge, ageKeys))
		}
	}

	for _, route := range BaseTradeRoutes() {
		if route.MinAge != "" && !ageKeys[route.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in trade route\n"+
				"  File:     config/trade.go (BaseTradeRoutes)\n"+
				"  Route:    %q (%s)\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				route.Key, route.Name, route.MinAge, hint(route.MinAge, ageKeys))
		}
		if route.RequiredBld != "" {
			if _, ok := buildingKeys[route.RequiredBld]; !ok {
				t.Errorf("\n"+
					"  Bad building key in trade route\n"+
					"  File:     config/trade.go (BaseTradeRoutes)\n"+
					"  Route:    %q (%s)\n"+
					"  Field:    RequiredBld\n"+
					"  Got:      %q  <-- this building doesn't exist\n"+
					"  Fix:      Check config/buildings.go for valid building keys%s\n",
					route.Key, route.Name, route.RequiredBld, hintFromMap(route.RequiredBld, buildingKeys))
			}
		}
		for res := range route.Export {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in trade route exports\n"+
					"  File:     config/trade.go (BaseTradeRoutes)\n"+
					"  Route:    %q (%s)\n"+
					"  Field:    Export\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					route.Key, route.Name, res, hintFromMap(res, resourceKeys))
			}
		}
		for res := range route.Import {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in trade route imports\n"+
					"  File:     config/trade.go (BaseTradeRoutes)\n"+
					"  Route:    %q (%s)\n"+
					"  Field:    Import\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					route.Key, route.Name, res, hintFromMap(res, resourceKeys))
			}
		}
	}

	for _, faction := range BaseFactions() {
		if faction.MinAge != "" && !ageKeys[faction.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in faction definition\n"+
				"  File:     config/trade.go (BaseFactions)\n"+
				"  Faction:  %q (%s)\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				faction.Key, faction.Name, faction.MinAge, hint(faction.MinAge, ageKeys))
		}
		if faction.Specialty != "" {
			if _, ok := resourceKeys[faction.Specialty]; !ok {
				t.Errorf("\n"+
					"  Bad resource key in faction specialty\n"+
					"  File:     config/trade.go (BaseFactions)\n"+
					"  Faction:  %q (%s)\n"+
					"  Field:    Specialty\n"+
					"  Got:      %q  <-- this resource doesn't exist\n"+
					"  Fix:      Check config/resources.go for valid resource keys%s\n",
					faction.Key, faction.Name, faction.Specialty, hintFromMap(faction.Specialty, resourceKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Event config validation
// ---------------------------------------------------------------------------

func TestConfig_EventKeysExist(t *testing.T) {
	resourceKeys := ResourceByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, event := range RandomEvents() {
		if event.MinAge != "" && !ageKeys[event.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in event definition\n"+
				"  File:     config/events.go\n"+
				"  Event:    %q (%s)\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				event.Key, event.Name, event.MinAge, hint(event.MinAge, ageKeys))
		}
		for _, eff := range event.Effects {
			if _, ok := resourceKeys[eff.Target]; !ok {
				if !isSpecialTarget(eff.Target) {
					t.Errorf("\n"+
						"  Bad effect target in event definition\n"+
						"  File:     config/events.go\n"+
						"  Event:    %q (%s)\n"+
						"  Field:    Effects[].Target\n"+
						"  Got:      %q  <-- not a valid resource or bonus key\n"+
						"  Fix:      Use a resource key from config/resources.go, or a bonus key\n"+
						"            like production_all, gather_rate, tick_speed, etc.%s\n",
						event.Key, event.Name, eff.Target, hintFromMap(eff.Target, resourceKeys))
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Upgrade config validation
// ---------------------------------------------------------------------------

func TestConfig_UpgradeKeysExist(t *testing.T) {
	buildingKeys := BuildingByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, upg := range BuildingUpgrades() {
		if _, ok := buildingKeys[upg.From]; !ok {
			t.Errorf("\n"+
				"  Bad building key in upgrade definition\n"+
				"  File:     config/upgrades.go\n"+
				"  Upgrade:  %s -> %s\n"+
				"  Field:    From\n"+
				"  Got:      %q  <-- this building doesn't exist\n"+
				"  Fix:      Check config/buildings.go for valid building keys%s\n",
				upg.From, upg.To, upg.From, hintFromMap(upg.From, buildingKeys))
		}
		if _, ok := buildingKeys[upg.To]; !ok {
			t.Errorf("\n"+
				"  Bad building key in upgrade definition\n"+
				"  File:     config/upgrades.go\n"+
				"  Upgrade:  %s -> %s\n"+
				"  Field:    To\n"+
				"  Got:      %q  <-- this building doesn't exist\n"+
				"  Fix:      Check config/buildings.go for valid building keys%s\n",
				upg.From, upg.To, upg.To, hintFromMap(upg.To, buildingKeys))
		}
		if upg.MinAge != "" && !ageKeys[upg.MinAge] {
			t.Errorf("\n"+
				"  Bad age key in upgrade definition\n"+
				"  File:     config/upgrades.go\n"+
				"  Upgrade:  %s -> %s\n"+
				"  Field:    MinAge\n"+
				"  Got:      %q  <-- this age doesn't exist\n"+
				"  Fix:      Check config/ages.go for valid age keys%s\n",
				upg.From, upg.To, upg.MinAge, hint(upg.MinAge, ageKeys))
		}
	}
}

// ---------------------------------------------------------------------------
// Duplicate key detection
// ---------------------------------------------------------------------------

func TestConfig_NoDuplicateKeys(t *testing.T) {
	checkDupes := func(name, file string, items []string) {
		seen := make(map[string]bool)
		for _, k := range items {
			if seen[k] {
				t.Errorf("\n"+
					"  Duplicate %s key found\n"+
					"  File:     %s\n"+
					"  Key:      %q  <-- appears more than once\n"+
					"  Fix:      Remove or rename one of the duplicates\n",
					name, file, k)
			}
			seen[k] = true
		}
	}

	var resourceKeys, buildingKeys, techKeys, ageKeys, milestoneKeys, eventKeys []string
	for _, r := range BaseResources() {
		resourceKeys = append(resourceKeys, r.Key)
	}
	for _, b := range BaseBuildings() {
		buildingKeys = append(buildingKeys, b.Key)
	}
	for _, tc := range Technologies() {
		techKeys = append(techKeys, tc.Key)
	}
	for _, a := range Ages() {
		ageKeys = append(ageKeys, a.Key)
	}
	for _, m := range Milestones() {
		milestoneKeys = append(milestoneKeys, m.Key)
	}
	for _, e := range RandomEvents() {
		eventKeys = append(eventKeys, e.Key)
	}

	checkDupes("resource", "config/resources.go", resourceKeys)
	checkDupes("building", "config/buildings.go", buildingKeys)
	checkDupes("tech", "config/research.go", techKeys)
	checkDupes("age", "config/ages.go", ageKeys)
	checkDupes("milestone", "config/milestones.go", milestoneKeys)
	checkDupes("event", "config/events.go", eventKeys)
}

// ---------------------------------------------------------------------------
// Reachability checks
// ---------------------------------------------------------------------------

func TestConfig_AllUnlockedBuildingsAreReachable(t *testing.T) {
	buildingKeys := BuildingByKey()
	unlocked := make(map[string]bool)
	for _, age := range Ages() {
		for _, bld := range age.UnlockBuildings {
			unlocked[bld] = true
		}
	}
	for key := range buildingKeys {
		if !unlocked[key] {
			def := buildingKeys[key]
			t.Errorf("\n"+
				"  Orphaned building — never unlocked by any age\n"+
				"  File:     config/buildings.go + config/ages.go\n"+
				"  Building: %q (%s)\n"+
				"  Fix:      Add %q to the UnlockBuildings list of the appropriate\n"+
				"            age in config/ages.go (probably %q based on RequiredAge)\n",
				key, def.Name, key, def.RequiredAge)
		}
	}
}

func TestConfig_AllResourcesAreReachable(t *testing.T) {
	resourceKeys := ResourceByKey()
	unlocked := make(map[string]bool)
	for _, age := range Ages() {
		for _, res := range age.UnlockResources {
			unlocked[res] = true
		}
	}
	for key := range resourceKeys {
		if !unlocked[key] {
			def := resourceKeys[key]
			t.Errorf("\n"+
				"  Orphaned resource — never unlocked by any age\n"+
				"  File:     config/resources.go + config/ages.go\n"+
				"  Resource: %q (%s)\n"+
				"  Fix:      Add %q to the UnlockResources list of the appropriate\n"+
				"            age in config/ages.go (probably %q based on the resource's Age field)\n",
				key, def.Name, key, def.Age)
		}
	}
}

// ---------------------------------------------------------------------------
// Effect target validation
// ---------------------------------------------------------------------------

func TestConfig_EffectTargetsValid(t *testing.T) {
	resourceKeys := ResourceByKey()

	for _, bld := range BaseBuildings() {
		for _, eff := range bld.Effects {
			if _, ok := resourceKeys[eff.Target]; !ok && !isSpecialTarget(eff.Target) {
				t.Errorf("\n"+
					"  Bad effect target in building definition\n"+
					"  File:     config/buildings.go\n"+
					"  Building: %q (%s)\n"+
					"  Field:    Effects[].Target\n"+
					"  Got:      %q  <-- not a valid resource or bonus key\n"+
					"  Fix:      Use a resource key from config/resources.go, or a bonus key\n"+
					"            like production_all, gather_rate, tick_speed, military_power, etc.%s\n",
					bld.Key, bld.Name, eff.Target, hintFromMap(eff.Target, resourceKeys))
			}
		}
	}

	for _, tech := range Technologies() {
		for _, eff := range tech.Effects {
			if _, ok := resourceKeys[eff.Target]; !ok && !isSpecialTarget(eff.Target) {
				t.Errorf("\n"+
					"  Bad effect target in technology definition\n"+
					"  File:     config/research.go\n"+
					"  Tech:     %q (%s)\n"+
					"  Field:    Effects[].Target\n"+
					"  Got:      %q  <-- not a valid resource or bonus key\n"+
					"  Fix:      Use a resource key from config/resources.go, or a bonus key\n"+
					"            like production_all, gather_rate, tick_speed, military_power, etc.%s\n",
					tech.Key, tech.Name, eff.Target, hintFromMap(eff.Target, resourceKeys))
			}
		}
	}

	for _, ms := range Milestones() {
		for _, eff := range ms.Rewards {
			if _, ok := resourceKeys[eff.Target]; !ok && !isSpecialTarget(eff.Target) {
				t.Errorf("\n"+
					"  Bad reward target in milestone definition\n"+
					"  File:     config/milestones.go\n"+
					"  Milestone: %q (%s)\n"+
					"  Field:    Rewards[].Target\n"+
					"  Got:      %q  <-- not a valid resource or bonus key\n"+
					"  Fix:      Use a resource key from config/resources.go, or a bonus key\n"+
					"            like production_all, gather_rate, tick_speed, etc.%s\n",
					ms.Key, ms.Name, eff.Target, hintFromMap(eff.Target, resourceKeys))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Summary
// ---------------------------------------------------------------------------

func TestConfig_Summary(t *testing.T) {
	t.Logf("Config inventory: %d ages, %d resources, %d buildings, %d techs, %d milestones, %d events, %d trade routes, %d factions, %d upgrades",
		len(Ages()),
		len(BaseResources()),
		len(BaseBuildings()),
		len(Technologies()),
		len(Milestones()),
		len(RandomEvents()),
		len(BaseTradeRoutes()),
		len(BaseFactions()),
		len(BuildingUpgrades()),
	)
}

// ---------------------------------------------------------------------------
// isSpecialTarget — effect targets that aren't resource keys
// ---------------------------------------------------------------------------

func isSpecialTarget(target string) bool {
	specials := map[string]bool{
		"population": true, "military": true, "all": true,
		"production_all": true, "gather_rate": true, "expedition_reward": true,
		"knowledge_rate": true, "build_cost": true, "tick_speed": true,
		"storage": true, "trade_rate": true, "research_speed": true,
		"build_speed": true, "military_power": true,
		"food_rate": true, "gold_rate": true, "iron_rate": true,
		"stone_rate": true, "wood_rate": true, "coal_rate": true,
		"steel_rate": true, "oil_rate": true, "electricity_rate": true,
		"uranium_rate": true, "data_rate": true, "crypto_rate": true,
		"plasma_rate": true, "titanium_rate": true, "dark_matter_rate": true,
		"antimatter_rate": true, "quantum_flux_rate": true,
		"culture_rate": true, "faith_rate": true,
	}
	return specials[target]
}

func init() {
	_ = fmt.Sprintf
	_ = strings.Join
	_ = sort.Strings
}
