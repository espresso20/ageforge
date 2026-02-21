package config

import (
	"fmt"
	"testing"
)

// buildKeySet collects all keys from a slice using a getter function
func buildKeySet[T any](items []T, getKey func(T) string) map[string]bool {
	m := make(map[string]bool)
	for _, item := range items {
		m[getKey(item)] = true
	}
	return m
}

func TestConfig_AgeKeysExist(t *testing.T) {
	ages := Ages()
	ageKeys := buildKeySet(ages, func(a AgeDef) string { return a.Key })
	resourceKeys := ResourceByKey()
	buildingKeys := BuildingByKey()

	for _, age := range ages {
		// ResourceReqs must reference valid resources
		for res := range age.ResourceReqs {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("age %q ResourceReqs references unknown resource %q", age.Key, res)
			}
		}
		// BuildingReqs must reference valid buildings
		for bld := range age.BuildingReqs {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("age %q BuildingReqs references unknown building %q", age.Key, bld)
			}
		}
		// UnlockBuildings must reference valid buildings
		for _, bld := range age.UnlockBuildings {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("age %q UnlockBuildings references unknown building %q", age.Key, bld)
			}
		}
		// UnlockResources must reference valid resources
		for _, res := range age.UnlockResources {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("age %q UnlockResources references unknown resource %q", age.Key, res)
			}
		}
		_ = ageKeys // used below
	}
}

func TestConfig_BuildingKeysExist(t *testing.T) {
	buildings := BaseBuildings()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	techKeys := TechByKey()
	resourceKeys := ResourceByKey()

	for _, bld := range buildings {
		// RequiredAge must be valid
		if bld.RequiredAge != "" {
			if !ageKeys[bld.RequiredAge] {
				t.Errorf("building %q RequiredAge references unknown age %q", bld.Key, bld.RequiredAge)
			}
		}
		// RequiredTech must be valid
		if bld.RequiredTech != "" {
			if _, ok := techKeys[bld.RequiredTech]; !ok {
				t.Errorf("building %q RequiredTech references unknown tech %q", bld.Key, bld.RequiredTech)
			}
		}
		// BaseCost keys must be valid resources
		for res := range bld.BaseCost {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("building %q BaseCost references unknown resource %q", bld.Key, res)
			}
		}
	}
}

func TestConfig_TechKeysExist(t *testing.T) {
	techs := Technologies()
	techKeys := TechByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, tech := range techs {
		// Age must be valid
		if tech.Age != "" {
			if !ageKeys[tech.Age] {
				t.Errorf("tech %q references unknown age %q", tech.Key, tech.Age)
			}
		}
		// Prerequisites must be valid tech keys
		for _, prereq := range tech.Prerequisites {
			if _, ok := techKeys[prereq]; !ok {
				t.Errorf("tech %q prerequisite references unknown tech %q", tech.Key, prereq)
			}
		}
	}
}

func TestConfig_MilestoneKeysExist(t *testing.T) {
	milestones := Milestones()
	milestoneKeys := MilestoneByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	resourceKeys := ResourceByKey()
	buildingKeys := BuildingByKey()
	techKeys := TechByKey()

	for _, ms := range milestones {
		if ms.MinAge != "" {
			if !ageKeys[ms.MinAge] {
				t.Errorf("milestone %q MinAge references unknown age %q", ms.Key, ms.MinAge)
			}
		}
		for res := range ms.MinResources {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("milestone %q MinResources references unknown resource %q", ms.Key, res)
			}
		}
		for bld := range ms.MinBuildings {
			if _, ok := buildingKeys[bld]; !ok {
				t.Errorf("milestone %q MinBuildings references unknown building %q", ms.Key, bld)
			}
		}
		for _, tech := range ms.RequiredTechs {
			if _, ok := techKeys[tech]; !ok {
				t.Errorf("milestone %q RequiredTechs references unknown tech %q", ms.Key, tech)
			}
		}
	}

	// Chain milestone keys must exist
	for _, chain := range MilestoneChains() {
		for _, mk := range chain.MilestoneKeys {
			if _, ok := milestoneKeys[mk]; !ok {
				t.Errorf("chain %q references unknown milestone %q", chain.Key, mk)
			}
		}
	}
}

func TestConfig_TradeKeysExist(t *testing.T) {
	resourceKeys := ResourceByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })
	buildingKeys := BuildingByKey()

	for _, rate := range BaseExchangeRates() {
		if _, ok := resourceKeys[rate.From]; !ok {
			t.Errorf("exchange rate From references unknown resource %q", rate.From)
		}
		if _, ok := resourceKeys[rate.To]; !ok {
			t.Errorf("exchange rate To references unknown resource %q", rate.To)
		}
		if rate.MinAge != "" && !ageKeys[rate.MinAge] {
			t.Errorf("exchange rate %s→%s MinAge references unknown age %q", rate.From, rate.To, rate.MinAge)
		}
	}

	for _, route := range BaseTradeRoutes() {
		if route.MinAge != "" && !ageKeys[route.MinAge] {
			t.Errorf("trade route %q MinAge references unknown age %q", route.Key, route.MinAge)
		}
		if route.RequiredBld != "" {
			if _, ok := buildingKeys[route.RequiredBld]; !ok {
				t.Errorf("trade route %q RequiredBld references unknown building %q", route.Key, route.RequiredBld)
			}
		}
		for res := range route.Export {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("trade route %q Export references unknown resource %q", route.Key, res)
			}
		}
		for res := range route.Import {
			if _, ok := resourceKeys[res]; !ok {
				t.Errorf("trade route %q Import references unknown resource %q", route.Key, res)
			}
		}
	}

	for _, faction := range BaseFactions() {
		if faction.MinAge != "" && !ageKeys[faction.MinAge] {
			t.Errorf("faction %q MinAge references unknown age %q", faction.Key, faction.MinAge)
		}
		if faction.Specialty != "" {
			if _, ok := resourceKeys[faction.Specialty]; !ok {
				t.Errorf("faction %q Specialty references unknown resource %q", faction.Key, faction.Specialty)
			}
		}
	}
}

func TestConfig_EventKeysExist(t *testing.T) {
	resourceKeys := ResourceByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, event := range RandomEvents() {
		if event.MinAge != "" && !ageKeys[event.MinAge] {
			t.Errorf("event %q MinAge references unknown age %q", event.Key, event.MinAge)
		}
		for _, eff := range event.Effects {
			// Effect targets can be resource keys or special bonus keys
			if _, ok := resourceKeys[eff.Target]; !ok {
				if !isSpecialTarget(eff.Target) {
					t.Errorf("event %q effect target references unknown key %q", event.Key, eff.Target)
				}
			}
		}
	}
}

func TestConfig_UpgradeKeysExist(t *testing.T) {
	buildingKeys := BuildingByKey()
	ageKeys := buildKeySet(Ages(), func(a AgeDef) string { return a.Key })

	for _, upg := range BuildingUpgrades() {
		if _, ok := buildingKeys[upg.From]; !ok {
			t.Errorf("upgrade From references unknown building %q", upg.From)
		}
		if _, ok := buildingKeys[upg.To]; !ok {
			t.Errorf("upgrade To references unknown building %q", upg.To)
		}
		if upg.MinAge != "" && !ageKeys[upg.MinAge] {
			t.Errorf("upgrade %s→%s MinAge references unknown age %q", upg.From, upg.To, upg.MinAge)
		}
	}
}

func TestConfig_NoDuplicateKeys(t *testing.T) {
	// Check for duplicate keys within each config type
	checkDupes := func(name string, items []string) {
		seen := make(map[string]bool)
		for _, k := range items {
			if seen[k] {
				t.Errorf("duplicate %s key: %q", name, k)
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
	for _, t := range Technologies() {
		techKeys = append(techKeys, t.Key)
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

	checkDupes("resource", resourceKeys)
	checkDupes("building", buildingKeys)
	checkDupes("tech", techKeys)
	checkDupes("age", ageKeys)
	checkDupes("milestone", milestoneKeys)
	checkDupes("event", eventKeys)
}

func TestConfig_AllUnlockedBuildingsAreReachable(t *testing.T) {
	// Every building should be unlocked by some age
	buildingKeys := BuildingByKey()
	unlocked := make(map[string]bool)
	for _, age := range Ages() {
		for _, bld := range age.UnlockBuildings {
			unlocked[bld] = true
		}
	}
	for key := range buildingKeys {
		if !unlocked[key] {
			t.Errorf("building %q is never unlocked by any age", key)
		}
	}
}

func TestConfig_AllResourcesAreReachable(t *testing.T) {
	// Every resource should be unlocked by some age
	resourceKeys := ResourceByKey()
	unlocked := make(map[string]bool)
	for _, age := range Ages() {
		for _, res := range age.UnlockResources {
			unlocked[res] = true
		}
	}
	for key := range resourceKeys {
		if !unlocked[key] {
			t.Errorf("resource %q is never unlocked by any age", key)
		}
	}
}

// isSpecialTarget returns true for effect targets that aren't resource keys
func isSpecialTarget(target string) bool {
	specials := map[string]bool{
		"population":       true,
		"military":         true,
		"all":              true,
		"production_all":   true,
		"gather_rate":      true,
		"expedition_reward": true,
		"knowledge_rate":   true,
		"build_cost":       true,
		"tick_speed":       true,
		"storage":          true,
		"trade_rate":       true,
		"research_speed":   true,
		"build_speed":      true,
		"military_power":   true,
		"food_rate":        true,
		"gold_rate":        true,
		"iron_rate":        true,
		"stone_rate":       true,
		"wood_rate":        true,
		"coal_rate":        true,
		"steel_rate":       true,
		"oil_rate":         true,
		"electricity_rate": true,
		"uranium_rate":     true,
		"data_rate":        true,
		"crypto_rate":      true,
		"plasma_rate":      true,
		"titanium_rate":    true,
		"dark_matter_rate": true,
		"antimatter_rate":  true,
		"quantum_flux_rate": true,
		"culture_rate":     true,
		"faith_rate":       true,
	}
	return specials[target]
}

// Smoke test: print summary of all config counts
func TestConfig_Summary(t *testing.T) {
	t.Logf("Config summary: %d ages, %d resources, %d buildings, %d techs, %d milestones, %d events, %d trade routes, %d factions, %d upgrades",
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

// Verify effect targets on buildings and techs
func TestConfig_EffectTargetsValid(t *testing.T) {
	resourceKeys := ResourceByKey()

	for _, bld := range BaseBuildings() {
		for _, eff := range bld.Effects {
			if _, ok := resourceKeys[eff.Target]; !ok {
				if !isSpecialTarget(eff.Target) {
					t.Errorf("building %q effect target references unknown key %q", bld.Key, eff.Target)
				}
			}
		}
	}

	for _, tech := range Technologies() {
		for _, eff := range tech.Effects {
			if _, ok := resourceKeys[eff.Target]; !ok {
				if !isSpecialTarget(eff.Target) {
					t.Errorf("tech %q effect target references unknown key %q (add to isSpecialTarget if intentional)", tech.Key, eff.Target)
				}
			}
		}
	}

	for _, ms := range Milestones() {
		for _, eff := range ms.Rewards {
			if _, ok := resourceKeys[eff.Target]; !ok {
				if !isSpecialTarget(eff.Target) {
					t.Errorf("milestone %q reward target references unknown key %q", ms.Key, eff.Target)
				}
			}
		}
	}
}

func init() {
	// Ensure fmt is used
	_ = fmt.Sprintf
}
