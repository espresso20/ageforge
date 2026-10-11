package config

// buildingsMonuments returns the Cultural Monuments — a culture sink.
//
// Culture accumulates endlessly (9+ buildings produce it) but its only sink
// used to be the one-time prestige/epoch gates, so billions sat idle late game.
// Monuments turn a LARGE lump of culture into a PERMANENT production bonus.
//
// Mechanism (construction-cost, NOT upkeep): there is no per-tick upkeep system
// in AgeForge, so Monuments are modelled as normal buildable structures whose
// BaseCost is paid almost entirely in culture. They are Category "monument" so
// the engine's static-bonus pass (getStaticBonuses) folds their "bonus" effects
// into the production_all multiplier — the same path wonders use — without
// routing them through the wonder-bank flow (that gate keys on Category
// "wonder"). Each is MaxCount 1 and gives a small, permanent +production_all.
//
// Balance intent: each monument's culture cost is roughly the culture a player
// has banked by the time its age unlocks, so building one is a real drain but
// never trivialises the prestige gates (which remain the primary long-term
// sink). Bonuses are deliberately modest (+1% → +5%, +11% total if all built).
func buildingsMonuments() []BuildingDef {
	return []BuildingDef{
		// Classical Age — first real culture stockpiles appear. Small primer.
		{
			Name: "Cultural Obelisk", Key: "cultural_obelisk", Category: "monument",
			BaseCost:  map[string]float64{"culture": 710, "stone": 17000},
			CostScale: 1.15,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.01},
			},
			RequiredAge: "classical_age",
			MaxCount:    1,
			BuildTicks:  2000,
			Description: "A carved obelisk celebrating your people's golden age.",
			LineageKey:  "monument",
		},
		// Medieval Age — culture flows steadily from great halls and cathedrals.
		{
			Name: "Grand Amphitheater", Key: "grand_amphitheatre_monument", Category: "monument",
			BaseCost:  map[string]float64{"culture": 7100, "stone": 110000, "gold": 34000},
			CostScale: 1.15,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.02},
			},
			RequiredAge: "medieval_age",
			MaxCount:    1,
			BuildTicks:  3510,
			Description: "A monumental arena for the games and pageants of an age.",
			LineageKey:  "monument",
		},
		// Industrial Age — mass media and museums pour out culture; bigger sink.
		{
			Name: "Eternal Library", Key: "eternal_library_monument", Category: "monument",
			BaseCost:  map[string]float64{"culture": 140000, "steel": 170000, "gold": 570000},
			CostScale: 1.15,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.03},
			},
			RequiredAge: "industrial_age",
			MaxCount:    1,
			BuildTicks:  6240,
			Description: "A vast archive preserving the knowledge of every age.",
			LineageKey:  "monument",
		},
		// Modern Age — late-game culture runs to the billions; the heavy sink.
		{
			Name: "Monument of Ages", Key: "monument_of_ages", Category: "monument",
			BaseCost:  map[string]float64{"culture": 7100000, "gold": 1.4e07},
			CostScale: 1.15,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.05},
			},
			RequiredAge:  "modern_age",
			RequiredTech: "television",
			MaxCount:     1,
			BuildTicks:   9360,
			Description:  "A timeless edifice commemorating the whole span of your civilization.",
			LineageKey:   "monument",
		},
	}
}
