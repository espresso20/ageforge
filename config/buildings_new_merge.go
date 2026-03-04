package config

// NewProductionBuildings returns all 13-lineage production, housing, research, and military
// buildings introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
//
// Lineages covered:
//  1. housing          — 21 tiers (primitive → quantum)
//  2. food             — 21 tiers (primitive → quantum)
//  3. organic_extraction — 21 tiers (primitive → quantum)
//  4. geological_extraction — 21 tiers (primitive → quantum)
//  5. knowledge        — 21 tiers (primitive → quantum)
//  6. faith            — 21 tiers (primitive → quantum)
//  7. military         — 21 tiers (primitive → quantum)
//  8. trade            — 19 tiers (bronze → quantum)
//  9. engineering      — 19 tiers (bronze → quantum)
// 10. culture_arts     — 17 tiers (classical → quantum)
// 11. metallurgy       — 18 tiers (iron → quantum)
// 12. energy           — 13 tiers (industrial → quantum)
// 13. hacker           — 8 tiers (information → quantum)
func NewProductionBuildings() []BuildingDef {
	out := newProductionBuildings()
	out = append(out, newProductionBuildings2()...)
	out = append(out, newProductionBuildings3()...)
	return out
}
