package citymap

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// citygen.go once held the count-driven city SYNTHESIZER (citymap v2): it grew a whole
// cityPlan — streets, blocks, count-scaled lots — that the isometric render path
// painted. The citymap v3 top-down rewrite (design-and-architecture/city-synthesis.md)
// replaced that whole pipeline with topdown.go's top-down engine (golden-spiral
// placement, fill-frame, the roof atlas), so the plan model + street/block/populate
// machinery is gone.
//
// What survives is the GATHER step — pure, deterministic, and still the input topdown.go
// builds its plan from: turn the player's built-building set into one classified
// builtBuilding per distinct type. topdown.go's generateTopPlan calls gatherBuildings
// and reads role/domain/category/tier/count off each entry.

// cityRole buckets a building by its function in the city fabric. Housing reads as
// residential dwellings; the civic set (faith / knowledge / culture / wonder /
// storage / diplomacy / research) reads as the labeled landmarks the city centers on;
// everything else is production. topdown.go uses the role to promote a hero and to
// classify districts.
type cityRole int

const (
	roleResidential cityRole = iota
	roleProduction
	roleLandmark
)

// builtBuilding is one distinct built type gathered for synthesis: its config key,
// display name, lineage/category (for color + role), tier, and instance count (drives
// the count-scaled roof population). topdown.go's generateTopPlan consumes a slice of
// these.
type builtBuilding struct {
	key      string
	name     string
	domain   string // lineageKey
	category string
	tier     int
	count    int
	role     cityRole
}

// classifyRole maps a building's category+lineage to its city role. Housing is
// residential; wonders/monuments/faith/knowledge/research/culture/diplomacy/storage
// read as civic landmarks the city centers on; everything else (the production
// lineages) is production. Pure data — safe on the render path.
func classifyRole(category, lineageKey string) cityRole {
	switch category {
	case "housing":
		return roleResidential
	case "wonder", "monument", "storage", "diplomacy", "research":
		return roleLandmark
	}
	switch lineageKey {
	case "faith", "knowledge", "culture_arts":
		return roleLandmark
	}
	return roleProduction
}

// gatherBuildings turns the built-building set into the synthesis input: one
// builtBuilding per distinct built type, classified into a city role, sorted for
// determinism (builtBuildingKeys returns sorted keys, so placement order is stable
// frame-to-frame). byKey is the pure config table (no locks). This is the single
// input to topdown.go's generateTopPlan.
func gatherBuildings(state game.GameState, byKey map[string]config.BuildingDef) []builtBuilding {
	keys, counts := builtBuildingKeys(state)
	out := make([]builtBuilding, 0, len(keys))
	for _, k := range keys {
		lineage, category, _ := buildingMeta(byKey, k)
		def := byKey[k]
		name := def.Name
		if name == "" {
			name = titleCaseKey(k)
		}
		domain := lineage
		// Special categories color by category (via lineageColor), so carry the
		// category as the domain when the lineage is empty/irrelevant.
		switch category {
		case "wonder", "monument", "storage", "diplomacy":
			domain = category
		}
		if domain == "" {
			domain = "misc"
		}
		out = append(out, builtBuilding{
			key:      k,
			name:     name,
			domain:   domain,
			category: category,
			tier:     def.LineageTier,
			count:    counts[k],
			role:     classifyRole(category, lineage),
		})
	}
	return out
}
