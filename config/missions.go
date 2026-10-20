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

// Missions returns the era mission catalog in epoch and kind order.
func Missions() []MissionDef {
	var out []MissionDef
	for _, era := range Epochs() {
		for _, item := range []struct {
			name, description, kind        string
			min, max, success, yield, cost float64
		}{
			{"", "", MissionScout, 5, 30, .85, 2, 5},
			{"", "", MissionExpedition, 30, 60, .75, 3, 15},
			{"", "", MissionCampaign, 60, 120, .65, 4, 10},
		} {
			name, description := missionCopy(era.Key, item.kind)
			out = append(out, MissionDef{
				Key: missionKey(name), Name: name, Description: description,
				Era: era.Key, Kind: item.kind, MinMinutes: item.min,
				MaxMinutes: item.max, BaseSuccess: item.success,
				Yield: item.yield, CostMinutes: item.cost,
			})
		}
	}
	return out
}

func missionCopy(era, kind string) (string, string) {
	copy := map[string][6]string{
		"stone_era":    {"Scouting Party", "The party scouts the nearby wilderness.", "Long Hunt", "The party hunts across the open plains.", "Raid the Rival Clan", "The party raids a rival clan's settlement."},
		"iron_era":     {"Road Survey", "The party surveys the roads beyond the city.", "Trade Caravan", "The party escorts a caravan along a distant trade route.", "Siege Campaign", "The army lays siege to a rival stronghold."},
		"steel_era":    {"Coastal Survey", "The party surveys a distant coastline.", "Naval Expedition", "The party sails to explore distant shores.", "Colonial Campaign", "The army campaigns across a faraway colony."},
		"electric_era": {"Aerial Survey", "The party surveys the land from above.", "Rail Expedition", "The party follows the rail line into new territory.", "Industrial War", "The army advances on an industrial rival."},
		"digital_era":  {"Data Run", "The party traces data through a remote network.", "Satellite Survey", "The party maps distant regions by satellite.", "Cyber Campaign", "The army launches a campaign across enemy networks."},
		"neon_era":     {"Street Recon", "The party scouts the neon streets.", "Neon Heist", "The party breaks into a vault in the neon district.", "Fusion Plant Assault", "The army assaults a rival fusion plant."},
		"cosmic_era":   {"Orbital Survey", "The party surveys a nearby planet from orbit.", "Deep Space Expedition", "The party explores a distant star system.", "Galactic Conquest", "The army advances to conquer a galactic sector."},
	}
	values := copy[era]
	index := map[string]int{MissionScout: 0, MissionExpedition: 1, MissionCampaign: 2}[kind] * 2
	return values[index], values[index+1]
}

func missionKey(name string) string {
	key := make([]byte, 0, len(name))
	underscore := false
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			key = append(key, c)
			underscore = false
		} else if len(key) > 0 && !underscore {
			key = append(key, '_')
			underscore = true
		}
	}
	for len(key) > 0 && key[len(key)-1] == '_' {
		key = key[:len(key)-1]
	}
	return string(key)
}
