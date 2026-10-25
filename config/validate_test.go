package config

import (
	"fmt"
	"maps"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// checkKey reports, on one line naming the file, the owner and the field, a key
// that is not in the valid set. It says whether the key is there.
func checkKey[V any](t *testing.T, file, owner, field, key string, valid map[string]V) bool {
	t.Helper()
	if _, ok := valid[key]; ok {
		return true
	}
	t.Errorf("%s: %s field %s: unknown key %q", file, owner, field, key)
	return false
}

// keysOf is the key of every row, in order, duplicates kept.
func keysOf[T any](rows []T, key func(T) string) []string {
	keys := make([]string, 0, len(rows))
	for _, r := range rows {
		keys = append(keys, key(r))
	}
	return keys
}

// ---------------------------------------------------------------------------
// Age config validation
// ---------------------------------------------------------------------------

func TestConfig_AgeKeysExist(t *testing.T) {
	resourceKeys := ResourceByKey()
	buildingKeys := BuildingByKey()

	for _, age := range Ages() {
		owner := fmt.Sprintf("age %q", age.Key)
		for bld := range age.BuildingReqs {
			checkKey(t, "ages.go", owner, "BuildingReqs", bld, buildingKeys)
		}
		for _, bld := range age.UnlockBuildings {
			checkKey(t, "ages.go", owner, "UnlockBuildings", bld, buildingKeys)
		}
		for _, res := range age.UnlockResources {
			checkKey(t, "ages.go", owner, "UnlockResources", res, resourceKeys)
		}
	}
}

// ---------------------------------------------------------------------------
// Building config validation
// ---------------------------------------------------------------------------

func TestConfig_BuildingKeysExist(t *testing.T) {
	ageKeys := AgeByKey()
	techKeys := TechByKey()
	resourceKeys := ResourceByKey()

	for _, bld := range BaseBuildings() {
		owner := fmt.Sprintf("building %q", bld.Key)
		if bld.RequiredAge != "" {
			checkKey(t, "buildings.go", owner, "RequiredAge", bld.RequiredAge, ageKeys)
		}
		if bld.RequiredTech != "" && checkKey(t, "buildings.go", owner, "RequiredTech", bld.RequiredTech, techKeys) {
			// A building can only be built in its own age (the age
			// lock), so a tech from any other age would open it too
			// early (never, in practice) or too late (never at all).
			if tech := techKeys[bld.RequiredTech]; tech.Age != bld.RequiredAge {
				t.Errorf("building %q needs tech %q from %s, but is built in %s: the tech must open in the building's age",
					bld.Key, bld.RequiredTech, tech.Age, bld.RequiredAge)
			}
		}
		for res := range bld.BaseCost {
			checkKey(t, "buildings.go", owner, "BaseCost", res, resourceKeys)
		}
	}
}

// ---------------------------------------------------------------------------
// Tech config validation
// ---------------------------------------------------------------------------

func TestConfig_TechKeysExist(t *testing.T) {
	techKeys := TechByKey()
	ageKeys := AgeByKey()

	for _, tech := range Technologies() {
		owner := fmt.Sprintf("tech %q", tech.Key)
		if tech.Age != "" {
			checkKey(t, "research.go", owner, "Age", tech.Age, ageKeys)
		}
		for _, prereq := range tech.Prerequisites {
			checkKey(t, "research.go", owner, "Prerequisites", prereq, techKeys)
		}
	}
}

// ---------------------------------------------------------------------------
// Milestone config validation
// ---------------------------------------------------------------------------

func TestConfig_MilestoneKeysExist(t *testing.T) {
	milestoneKeys := MilestoneByKey()
	ageKeys := AgeByKey()
	resourceKeys := withResearchTarget(ResourceByKey())
	buildingKeys := BuildingByKey()
	techKeys := TechByKey()

	for _, ms := range Milestones() {
		owner := fmt.Sprintf("milestone %q", ms.Key)
		if ms.MinAge != "" {
			checkKey(t, "milestones.go", owner, "MinAge", ms.MinAge, ageKeys)
		}
		for res := range ms.MinResources {
			checkKey(t, "milestones.go", owner, "MinResources", res, resourceKeys)
		}
		for bld := range ms.MinBuildings {
			checkKey(t, "milestones.go", owner, "MinBuildings", bld, buildingKeys)
		}
		for _, bld := range ms.MinBuildingSum.Keys {
			checkKey(t, "milestones.go", owner, "MinBuildingSum.Keys", bld, buildingKeys)
		}
		if sum := ms.MinBuildingSum; (sum.Count > 0) != (len(sum.Keys) > 0) {
			t.Errorf("%s: MinBuildingSum needs both keys and a count, got %d key(s) and count %d", ms.Key, len(sum.Keys), sum.Count)
		}
		for _, tech := range ms.RequiredTechs {
			checkKey(t, "milestones.go", owner, "RequiredTechs", tech, techKeys)
		}
	}

	for _, chain := range MilestoneChains() {
		owner := fmt.Sprintf("chain %q", chain.Key)
		for _, mk := range chain.MilestoneKeys {
			checkKey(t, "milestones.go", owner, "MilestoneKeys", mk, milestoneKeys)
		}
	}
}

// ---------------------------------------------------------------------------
// Trade config validation
// ---------------------------------------------------------------------------

func TestConfig_TradeKeysExist(t *testing.T) {
	resourceKeys := withResearchTarget(ResourceByKey())
	ageKeys := AgeByKey()
	buildingKeys := BuildingByKey()

	for _, route := range BaseTradeRoutes() {
		owner := fmt.Sprintf("route %q", route.Key)
		if route.MinAge != "" {
			checkKey(t, "trade.go", owner, "MinAge", route.MinAge, ageKeys)
		}
		if route.RequiredBld != "" {
			checkKey(t, "trade.go", owner, "RequiredBld", route.RequiredBld, buildingKeys)
		}
		for res := range route.Export {
			checkKey(t, "trade.go", owner, "Export", res, resourceKeys)
		}
		for res := range route.Import {
			checkKey(t, "trade.go", owner, "Import", res, resourceKeys)
		}
	}

	for _, faction := range BaseFactions() {
		owner := fmt.Sprintf("faction %q", faction.Key)
		if faction.MinAge != "" {
			checkKey(t, "trade.go", owner, "MinAge", faction.MinAge, ageKeys)
		}
		if faction.Specialty != "" {
			checkKey(t, "trade.go", owner, "Specialty", faction.Specialty, resourceKeys)
		}
		// Every civilization must declare a valid personality and a non-empty backstory.
		if !ValidPersonalities[faction.Personality] {
			t.Errorf("\n"+
				"  Bad personality in civilization definition\n"+
				"  File:     config/trade.go (BaseFactions)\n"+
				"  Civ:      %q (%s)\n"+
				"  Field:    Personality\n"+
				"  Got:      %q  <-- must be aggressive|peaceful|mercantile|isolationist\n",
				faction.Key, faction.Name, faction.Personality)
		}
		if faction.Backstory == "" {
			t.Errorf("\n"+
				"  Missing backstory in civilization definition\n"+
				"  File:     config/trade.go (BaseFactions)\n"+
				"  Civ:      %q (%s)\n"+
				"  Field:    Backstory must be non-empty (shown on first contact)\n",
				faction.Key, faction.Name)
		}
		if faction.Strength < 1 || faction.Strength > 5 {
			t.Errorf("\n"+
				"  Bad strength in civilization definition\n"+
				"  File:     config/trade.go (BaseFactions)\n"+
				"  Civ:      %q (%s)\n"+
				"  Field:    Strength = %d  <-- must be 1..5\n",
				faction.Key, faction.Name, faction.Strength)
		}
	}
}

// ---------------------------------------------------------------------------
// Event config validation
// ---------------------------------------------------------------------------

func TestConfig_EventKeysExist(t *testing.T) {
	targets := effectTargets()
	ageKeys := AgeByKey()

	for _, event := range RandomEvents() {
		owner := fmt.Sprintf("event %q", event.Key)
		if event.MinAge != "" {
			checkKey(t, "events.go", owner, "MinAge", event.MinAge, ageKeys)
		}
		for _, eff := range event.Effects {
			// worker_loss effects operate on workers directly — no resource target needed.
			if eff.Type == "worker_loss" {
				continue
			}
			checkKey(t, "events.go", owner, "Effects[].Target", eff.Target, targets)
		}
	}
}

// ---------------------------------------------------------------------------
// Duplicate key detection
// ---------------------------------------------------------------------------

func TestConfig_NoDuplicateKeys(t *testing.T) {
	checkDupes := func(name, file string, keys []string) {
		seen := make(map[string]bool, len(keys))
		for _, k := range keys {
			if seen[k] {
				t.Errorf("%s: duplicate %s key %q", file, name, k)
			}
			seen[k] = true
		}
	}

	checkDupes("resource", "resources.go", keysOf(BaseResources(), func(r ResourceDef) string { return r.Key }))
	checkDupes("building", "buildings.go", keysOf(BaseBuildings(), func(b BuildingDef) string { return b.Key }))
	checkDupes("tech", "research.go", keysOf(Technologies(), func(tc TechDef) string { return tc.Key }))
	checkDupes("age", "ages.go", keysOf(Ages(), func(a AgeDef) string { return a.Key }))
	checkDupes("milestone", "milestones.go", keysOf(Milestones(), func(m MilestoneDef) string { return m.Key }))
	checkDupes("event", "events.go", keysOf(RandomEvents(), func(e EventDef) string { return e.Key }))
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
	resourceKeys := withResearchTarget(ResourceByKey())
	targets := effectTargets()

	for _, bld := range BaseBuildings() {
		owner := fmt.Sprintf("building %q", bld.Key)
		for _, eff := range bld.Effects {
			// "morale" effects act on the civilization's morale dial directly —
			// they carry no resource Target, so skip target validation.
			if eff.Type == "morale" {
				continue
			}
			// "research" effects are the scholars' work: they carry no target
			// either, and a knowledge building (one that comes with scholars)
			// or a wonder is what carries one.
			if eff.Type == "research" {
				if eff.Target != "" || eff.Value <= 0 || (bld.Scholars <= 0 && bld.Category != "wonder") {
					t.Errorf("%s: research effect %+v needs no target, a positive value and scholars (or a wonder)", bld.Key, eff)
				}
				continue
			}
			checkKey(t, "buildings.go", owner, "Effects[].Target", eff.Target, targets)
		}
	}

	// A tech's effect is typed: a kind from config/tech_effects.go, and a
	// resource for the kinds that name one.
	isResource := func(key string) bool { _, ok := resourceKeys[key]; return ok }
	for _, tech := range Technologies() {
		for _, eff := range tech.Effects {
			if problem := eff.Check(isResource); problem != "" {
				t.Errorf("research.go: tech %q field Effects[]: %s", tech.Key, problem)
			}
		}
	}

	for _, ms := range Milestones() {
		owner := fmt.Sprintf("milestone %q", ms.Key)
		for _, eff := range ms.Rewards {
			checkKey(t, "milestones.go", owner, "Rewards[].Target", eff.Target, targets)
		}
	}
}

// ---------------------------------------------------------------------------
// specialTargets — effect targets that aren't resource keys
// ---------------------------------------------------------------------------

var specialTargets = map[string]bool{
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
	"opinion": true, "trade_route_income": true,
}

// effectTargets is every key an effect may name: a resource (and the work
// scholars do) or one of the special targets.
func effectTargets() map[string]bool {
	out := maps.Clone(specialTargets)
	for k := range withResearchTarget(ResourceByKey()) {
		out[k] = true
	}
	return out
}

// withResearchTarget adds the work scholars do to a set of resource keys: it is
// not a resource, but it is a valid target for a gain, a rate, a trade
// route's import and a milestone's reward and requirement.
func withResearchTarget(m map[string]ResourceDef) map[string]ResourceDef {
	out := maps.Clone(m)
	out[ResearchTarget] = ResourceDef{Key: ResearchTarget, Name: "Research"}
	return out
}
