package config

// The gate's four pillars. What stands between a town and the next age is
// shown as four groups (the Advance panel): Knowledge, People, Industry and
// Monument. This file is the one place that says which group a building
// belongs to; nothing here changes what a gate asks for (AgeDef.BuildingReqs
// and the leaving age's wonder are still the whole rule).

// The pillar keys, in the order the panel lists them.
const (
	PillarKnowledge = "knowledge"
	PillarPeople    = "people"
	PillarIndustry  = "industry"
	PillarMonument  = "monument"
)

// GatePillars lists the pillar keys in display order.
func GatePillars() []string {
	return []string{PillarKnowledge, PillarPeople, PillarIndustry, PillarMonument}
}

// PillarName is a pillar's name on screen.
func PillarName(pillar string) string {
	switch pillar {
	case PillarKnowledge:
		return "Knowledge"
	case PillarPeople:
		return "People"
	case PillarIndustry:
		return "Industry"
	case PillarMonument:
		return "Monument"
	}
	return pillar
}

// GatePillar is the pillar a building counts toward when a gate asks for it:
//
//   - Monument: a wonder (no gate asks for one by count; the age's wonder is
//     the Monument pillar's own row);
//   - People: housing, the food lineage (the buildings that feed a town) and
//     the faith lineage (shrines, temples and the rest, which are research
//     buildings by category but serve the town's people, not its scholars);
//   - Knowledge: the rest of the research category (the knowledge buildings
//     that bring scholars);
//   - Industry: everything else (producers, trade, military and storage).
//
// It reads only the building's category, lineage and worker domain, so a
// ruleset with other buildings sorts them the same way.
func GatePillar(d BuildingDef) string {
	switch {
	case d.Category == "wonder":
		return PillarMonument
	case d.Category == "housing", d.LineageKey == "food", d.WorkerDomain == "food",
		d.LineageKey == "faith", d.WorkerDomain == "faith":
		return PillarPeople
	case d.Category == "research":
		return PillarKnowledge
	}
	return PillarIndustry
}
