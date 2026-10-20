package config

// Mission kinds describe the three expedition lengths and lock rules.
const (
	MissionScout      = "scout"
	MissionExpedition = "expedition"
	MissionCampaign   = "campaign"
)

// MissionDef is a mission available throughout one epoch.
type MissionDef struct {
	Key         string
	Name        string
	Description string
	Era         string
	Kind        string
	MinMinutes  float64
	MaxMinutes  float64
	BaseSuccess float64
	Yield       float64
	CostMinutes float64
}

// missionTable is each mission's identity and words: the key, the name and
// the description, with the era and the kind. The tuning comes from the kind
// (missionKinds). Names and descriptions are placeholders for the owner's
// copy pass.
var missionTable = []MissionDef{
	{Key: "scouting_party", Name: "Scouting Party", Description: "The party scouts the nearby wilderness.", Era: "stone_era", Kind: MissionScout},
	{Key: "long_hunt", Name: "Long Hunt", Description: "The party hunts across the open plains.", Era: "stone_era", Kind: MissionExpedition},
	{Key: "raid_the_rival_clan", Name: "Raid the Rival Clan", Description: "The party raids a rival clan's settlement.", Era: "stone_era", Kind: MissionCampaign},
	{Key: "road_survey", Name: "Road Survey", Description: "The party surveys the roads beyond the city.", Era: "iron_era", Kind: MissionScout},
	{Key: "trade_caravan", Name: "Trade Caravan", Description: "The party escorts a caravan along a distant trade route.", Era: "iron_era", Kind: MissionExpedition},
	{Key: "siege_campaign", Name: "Siege Campaign", Description: "The army lays siege to a rival stronghold.", Era: "iron_era", Kind: MissionCampaign},
	{Key: "coastal_survey", Name: "Coastal Survey", Description: "The party surveys a distant coastline.", Era: "steel_era", Kind: MissionScout},
	{Key: "naval_expedition", Name: "Naval Expedition", Description: "The party sails to explore distant shores.", Era: "steel_era", Kind: MissionExpedition},
	{Key: "colonial_campaign", Name: "Colonial Campaign", Description: "The army campaigns across a faraway colony.", Era: "steel_era", Kind: MissionCampaign},
	{Key: "aerial_survey", Name: "Aerial Survey", Description: "The party surveys the land from above.", Era: "electric_era", Kind: MissionScout},
	{Key: "rail_expedition", Name: "Rail Expedition", Description: "The party follows the rail line into new territory.", Era: "electric_era", Kind: MissionExpedition},
	{Key: "industrial_war", Name: "Industrial War", Description: "The army advances on an industrial rival.", Era: "electric_era", Kind: MissionCampaign},
	{Key: "data_run", Name: "Data Run", Description: "The party traces data through a remote network.", Era: "digital_era", Kind: MissionScout},
	{Key: "satellite_survey", Name: "Satellite Survey", Description: "The party maps distant regions by satellite.", Era: "digital_era", Kind: MissionExpedition},
	{Key: "cyber_campaign", Name: "Cyber Campaign", Description: "The army launches a campaign across enemy networks.", Era: "digital_era", Kind: MissionCampaign},
	{Key: "street_recon", Name: "Street Recon", Description: "The party scouts the neon streets.", Era: "neon_era", Kind: MissionScout},
	{Key: "neon_heist", Name: "Neon Heist", Description: "The party breaks into a vault in the neon district.", Era: "neon_era", Kind: MissionExpedition},
	{Key: "fusion_plant_assault", Name: "Fusion Plant Assault", Description: "The army assaults a rival fusion plant.", Era: "neon_era", Kind: MissionCampaign},
	{Key: "orbital_survey", Name: "Orbital Survey", Description: "The party surveys a nearby planet from orbit.", Era: "cosmic_era", Kind: MissionScout},
	{Key: "deep_space_expedition", Name: "Deep Space Expedition", Description: "The party explores a distant star system.", Era: "cosmic_era", Kind: MissionExpedition},
	{Key: "galactic_conquest", Name: "Galactic Conquest", Description: "The army advances to conquer a galactic sector.", Era: "cosmic_era", Kind: MissionCampaign},
}

// missionKinds is what a kind of mission is: how long it is out, how likely
// it is to come back with a success, how many minutes of the town's income
// its haul is worth for each minute out, and what it costs in minutes of
// income (of soldier income, for a campaign).
var missionKinds = map[string]struct{ min, max, success, yield, cost float64 }{
	MissionScout:      {5, 30, .85, 2, 5},
	MissionExpedition: {30, 60, .75, 3, 15},
	MissionCampaign:   {60, 120, .65, 4, 10},
}

// Missions returns the era mission catalog in epoch and kind order.
func Missions() []MissionDef {
	out := make([]MissionDef, 0, len(missionTable))
	for _, m := range missionTable {
		k := missionKinds[m.Kind]
		m.MinMinutes, m.MaxMinutes, m.BaseSuccess, m.Yield, m.CostMinutes = k.min, k.max, k.success, k.yield, k.cost
		out = append(out, m)
	}
	return out
}
