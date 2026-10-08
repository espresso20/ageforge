package config

// badge_catalog.go is the badge catalog: the families made from the game's
// tables, the ladders over the account's counters, and the hand-written
// specials. badges.go has the types; rules/badges.go expands the families
// when the ruleset is compiled; smoke/static_badges.go proves each badge
// can be earned.
//
// A badge that cannot be shown to be earnable does not ship. The few whose
// only proof would be a way of playing no smoke bot plays yet are kept
// ready in BadgesHeldOut and are not in the ruleset.

// ----- the tables a family may be made from that live outside config -----

// BadgeExpeditionDef is an expedition as the catalog needs it. The
// expeditions themselves belong to the game package, which checks this
// table against its own (TestBadgeExpeditionsMatchTheGame).
type BadgeExpeditionDef struct{ Key, Name, MinAge string }

// BadgeExpeditions lists every expedition, in the game's order.
func BadgeExpeditions() []BadgeExpeditionDef {
	return []BadgeExpeditionDef{
		{"scout_party", "Scout Party", "primitive_age"},
		{"scout_ruins", "Scout Nearby Ruins", "bronze_age"},
		{"raid_bandits", "Raid Bandit Camp", "bronze_age"},
		{"trade_escort", "Trade Escort", "iron_age"},
		{"conquer_territory", "Conquer Territory", "iron_age"},
		{"siege_castle", "Siege Enemy Castle", "medieval_age"},
		{"naval_expedition", "Naval Expedition", "renaissance_age"},
		{"colonial_campaign", "Colonial Campaign", "industrial_age"},
		{"world_domination", "World Domination", "modern_age"},
		{"cyber_raid", "Cyber Raid", "information_age"},
		{"neon_heist", "Neon Heist", "cyberpunk_age"},
		{"fusion_assault", "Fusion Plant Assault", "fusion_age"},
		{"orbital_strike", "Orbital Strike", "space_age"},
		{"warp_invasion", "Warp Invasion", "interstellar_age"},
		{"galactic_conquest", "Galactic Conquest", "galactic_age"},
		{"quantum_incursion", "Quantum Incursion", "quantum_age"},
	}
}

// BadgeThemeDef is a theme as the catalog needs it: its key, its name and
// the badge that gives it ("" for a theme every account has). The themes
// themselves belong to the theme package; the UI checks this table against
// its registry (TestBadgeThemesMatchTheRegistry).
type BadgeThemeDef struct{ Key, Name, Badge string }

// BadgeThemes lists every theme, in the picker's order.
func BadgeThemes() []BadgeThemeDef {
	return []BadgeThemeDef{
		{"forge", "Forge", ""},
		{"daylight", "Daylight", ""},
		{"deuteranopia", "Deuteranopia-safe", ""},
		{"protanopia", "Protanopia-safe", ""},
		{"high_contrast", "High Contrast", ""},
		{"high_contrast_light", "High Contrast Light", ""},
		{"parchment", "Parchment", "age.renaissance_age"},
		{"bronze", "Bronze", "age.bronze_age"},
		{"cyberpunk", "Cyberpunk", "age.cyberpunk_age"},
		{"monochrome", "Monochrome", "age.information_age"},
		{"cosmic", "Cosmic", "age.galactic_age"},
		{"source", "Source", "special.touched_by_the_source"},
		{"glitch", "Glitch", "special.creative_accounting"},
		{"ashfall", "Ashfall", "special.connoisseur_of_endings"},
		{"ledger", "Ledger", "ladder.deals.3"},
		{"prismatic", "Prismatic", "special.museum_piece"},
	}
}

// The map styles and glyph sets a badge is worn for. They belong to the
// UI, which checks them against its own (TestBadgeMapLooksMatchTheGame).
var (
	BadgeMapStyles = []string{"roguelike", "skyline"}
	BadgeMapGlyphs = []string{"ascii", "unicode", "nerd"}
)

// BadgeTitleDef is a title an account may wear: the title, and what earns
// it. Exactly one of Badge and Set is set: the badge to hold, or the set
// to hold every badge of.
type BadgeTitleDef struct {
	Title string
	Badge string
	Set   string
}

// BadgeTitles returns the titles badges give, besides the titles of the
// score (BadgeScoreTitles).
func BadgeTitles() []BadgeTitleDef {
	return []BadgeTitleDef{
		{Title: "Survivor", Badge: "ladder.endured.3"},
		{Title: "Harbinger Whisperer", Set: "harbinger.heeded"},
		{Title: "Merchant Prince", Badge: "ladder.deals.3"},
		{Title: "The Undying", Badge: "special.unkillable"},
		{Title: "Cosmic Heir", Badge: "lastpassage.succumbed"},
		{Title: "Hut Magnate", Badge: "special.zoning_board"},
	}
}

// ----- the generated families -----

// The rungs of the families that are ladders.
var (
	lineageRungs = []BadgeRung{
		{Name: "Hobbyist", Tier: BadgeBronze}, {Name: "Contractor", Tier: BadgeSilver}, {Name: "Magnate", Tier: BadgeGold},
		{Name: "Tycoon", Tier: BadgePlatinum}, {Name: "Dynasty", Tier: BadgeLegendary},
	}
	resourceRungs = []BadgeRung{
		{Name: "Trickle", Tier: BadgeBronze}, {Name: "Stream", Tier: BadgeSilver},
		{Name: "River", Tier: BadgeGold}, {Name: "Flood", Tier: BadgeLegendary},
	}
	payrollRungs = []BadgeRung{{Name: "Payroll", Tier: BadgeSilver}, {Name: "Payroll II", Tier: BadgeGold}}
)

// mastery is the standing argument for a badge that asks for an age or an
// era inside its pacing target.
const mastery = "Era Mastery shortens every age an account has completed before: at mastery an age takes its pacing target divided by its mastery factor, so a returning account leaves an age well inside the target."

// BadgeFamilies returns the generated families, in the order their tabs
// and their badges list.
func BadgeFamilies() []BadgeFamilyDef {
	fams := []BadgeFamilyDef{
		// F1. An age reached. Five of them give the age's theme.
		{
			Family: "age", Source: BadgeSourceAges, Except: []string{"primitive_age"},
			Key: "age.{key}", Name: "{name}", Desc: "Reach the {name}.",
			Names: map[string]string{
				"age.stone_age":        "Rock Solid",
				"age.iron_age":         "Age of Iron",
				"age.renaissance_age":  "Born Again, Slightly",
				"age.modern_age":       "Into the Modern Age",
				"age.transcendent_age": "Nothing Left to Prove",
			},
			Descs: map[string]string{
				"age.bronze_age":      "Reach the Bronze Age. Unlocks the Bronze theme.",
				"age.renaissance_age": "Reach the Renaissance Age. Unlocks the Parchment theme.",
				"age.information_age": "Reach the Information Age. Unlocks the Monochrome theme.",
				"age.cyberpunk_age":   "Reach the Cyberpunk Age. Unlocks the Cyberpunk theme.",
				"age.galactic_age":    "Reach the Galactic Age. Unlocks the Cosmic theme.",
			},
			TierByEra: true,
			Tiers:     map[string]BadgeTier{"transcendent_age": BadgeLegendary},
			Scope:     BadgeMoment, Event: BadgeEvAgeReached,
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleGate),
			Emblem:          "centre.{era}",
			Aliases: map[string][]string{
				"age.iron_age":   {"reached_iron"},
				"age.modern_age": {"reached_modern"},
			},
			Rewards: map[string]BadgeReward{
				"age.bronze_age":      {Theme: "bronze"},
				"age.renaissance_age": {Theme: "parchment"},
				"age.information_age": {Theme: "monochrome"},
				"age.cyberpunk_age":   {Theme: "cyberpunk"},
				"age.galactic_age":    {Theme: "cosmic"},
			},
		},
		// F2. An era left inside its pacing target. The first four are the
		// eras the veteran preset plays on every pull request; the last
		// three rest on the same mastery.
		swift([]string{"stone_era", "iron_era", "steel_era", "electric_era"}, BotProof("veteran")),
		swift([]string{"digital_era", "neon_era", "cosmic_era"}, Occurs(mastery)),
		// F3. A wonder raised.
		{
			Family: "wonder", Source: BadgeSourceWonders,
			Key: "wonder.{key}", Name: "{name}", Desc: "Raise the {name}.",
			TierByEra: true,
			Tiers:     map[string]BadgeTier{"singularity_core": BadgeLegendary},
			Scope:     BadgeMoment, Event: BadgeEvWonderRaised,
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleWonder),
			Emblem:          "wonder", Set: "wonder",
		},
		// F4. Most of what an age's storage can hold of its buildings,
		// standing at once (appendix C of the design: three quarters).
		{
			Family: "maximalist", Source: BadgeSourceAges,
			Key: "maximalist.{key}", Name: "{short} Maximalist",
			Desc:  "Have {count} of the {name}'s buildings standing before you leave it. Wonders do not count.",
			Descs: map[string]string{"maximalist.transcendent_age": "Have {count} of the Transcendent Age's buildings standing before you prestige. Wonders do not count."},
			Tier:  BadgeGold, Rarity: BadgeEpic,
			Scope: BadgeRun, Event: BadgeEvBuildingBuilt, InAge: "{key}", Counter: "standing_age.{key}",
			Thresholds: map[string]float64{
				"primitive_age": 229, "stone_age": 272, "bronze_age": 269, "iron_age": 297, "classical_age": 236,
				"medieval_age": 214, "renaissance_age": 246, "colonial_age": 326, "industrial_age": 362,
				"victorian_age": 290, "electric_age": 260, "atomic_age": 257, "modern_age": 276,
				"information_age": 311, "digital_age": 282, "cyberpunk_age": 272, "fusion_age": 217,
				"space_age": 227, "interstellar_age": 245, "galactic_age": 273, "quantum_age": 259,
				"transcendent_age": 27,
			},
			RevealAtAge: true,
			Proof:       StaticProof(BadgeRuleAgeCopies),
			Emblem:      "hall",
		},
		// F5. Every tech of an age, in one run.
		{
			Family: "techs", Source: BadgeSourceAges,
			Key: "techs.{key}", Name: "{short} Syllabus", Desc: "{techs}",
			Tier:  BadgeSilver,
			Scope: BadgeRun, Event: BadgeEvResearchDone, Counter: "researched_age.{key}",
			Measure:     BadgeMeasureTechs,
			RevealAtAge: true,
			Proof:       StaticProof(BadgeRuleTechs),
			Emblem:      "knowledge",
		},
		// F6. A lineage's buildings across every run (appendix A: a
		// quarter of a run, one, four, ten and twenty-five runs).
		{
			Family: "lineage", Source: BadgeSourceLineages,
			Key: "lineage.{key}.{n}", Name: "{name} {rung}",
			Desc:  "Build {count} {lname} buildings across all your runs. Sold and rebuilt copies count once.",
			Scope: BadgeLifetime, Counter: BadgeEvBuiltLineage + ".{key}",
			Rungs: lineageRungs,
			Ladders: map[string][]float64{
				"culture_arts":          {6, 22, 88, 220, 550},
				"energy":                {5, 10, 40, 100, 250},
				"engineering":           {7, 26, 100, 260, 650},
				"faith":                 {10, 38, 150, 380, 950},
				"food":                  {10, 38, 150, 380, 950},
				"geological_extraction": {10, 40, 160, 400, 1000},
				"hacker":                {5, 17, 68, 170, 430},
				"harbor":                {3, 5, 20, 50, 130},
				"housing":               {14, 57, 230, 570, 1400},
				"knowledge":             {10, 38, 150, 380, 950},
				"metallurgy":            {6, 23, 92, 230, 580},
				"military":              {9, 35, 140, 350, 880},
				"organic_extraction":    {10, 39, 160, 390, 980},
				"storage":               {8, 33, 130, 330, 810},
				"trade":                 {6, 24, 96, 240, 600},
			},
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleLifetime),
			Emblem:          "lineage.{key}", Ladder: "{name}",
		},
		// F7. Workers at work in a domain at once (appendix D: a tenth and
		// a fifth of the slots a run's buildings hold).
		{
			Family: "domain", Source: BadgeSourceDomains,
			Key: "domain.{key}.{n}", Name: "{name} {rung}",
			Desc:  "Have {count} {lname} workers at work at the same time.",
			Scope: BadgeRun, Event: BadgeEvCensus, Counter: "ev.staffed.{key}",
			Rungs: payrollRungs,
			Ladders: map[string][]float64{
				"food": {270, 530}, "lumber": {270, 540}, "masonry": {270, 550}, "knowledge": {160, 320},
				"faith": {160, 310}, "military": {270, 540}, "trade": {170, 340}, "engineering": {180, 360},
				"metallurgy": {150, 290}, "energy": {76, 150}, "hacker": {260, 520},
			},
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleStaffing),
			Ladder:          "{name}",
			Emblem:          "lineage.{key}",
			Emblems:         map[string]string{"lumber": "lineage.organic_extraction", "masonry": "lineage.geological_extraction"},
		},
		// F8. A resource produced across every run (appendix B: a quarter
		// of a run, one, five and twenty-five runs; a resource that comes
		// after a run's last age starts at one).
		{
			Family: "resource", Source: BadgeSourceResources,
			Key: "resource.{key}.{n}", Name: "{name} {rung}",
			Desc:  "Produce {count} {lname} across all your runs. Trades, loot and gifts are not production.",
			Scope: BadgeLifetime, Counter: BadgeEvProduced + ".{key}",
			Rungs:   resourceRungs,
			Measure: BadgeMeasureProduction, Runs: []float64{0.25, 1, 5, 25},
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleLifetime),
			Emblem:          "store", Ladder: "{name}",
		},
		// F9. Each civilization: met, allied, a regular, and at peace again.
		{
			Family: "civ", Source: BadgeSourceCivs,
			Key: "civ.{key}.met", Name: "{name}: First Contact", Desc: "Meet the {name}.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvCivMet,
			RevealBySubject: true, Emblem: "civ", Set: "civ.met",
			Proof: Occurs("A civilization is met when the run reaches its age, or sooner through an expedition."),
		},
		{
			Family: "civ", Source: BadgeSourceCivs,
			Key: "civ.{key}.allied", Name: "{name}: Allies", Desc: "Ally with the {name}.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvCivAllied,
			RevealBySubject: true, Emblem: "diplomacy",
			Proof: Occurs("An alliance is offered by any civilization whose opinion has been raised with gifts, for a price in gold."),
		},
		{
			Family: "civ", Source: BadgeSourceCivs,
			Key: "civ.{key}.regular", Name: "{name}: Regular", Desc: "Take 5 deals from the {name} across your runs.",
			Tier: BadgeSilver, Scope: BadgeLifetime, Counter: BadgeEvDeal + ".{key}", Threshold: 5,
			RevealBySubject: true, Emblem: "trade",
			Proof: StaticProof(BadgeRuleLifetime),
		},
		{
			Family: "civ", Source: BadgeSourceCivs,
			Key: "civ.{key}.peace", Name: "{name}: Peace Terms", Desc: "End a war with the {name}, by tribute or by waiting it out.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvWarEnded,
			RevealBySubject: true, Emblem: "war",
			Proof: Occurs("Any civilization goes to war once its opinion is low enough and it has been provoked twice (an embargo, a raided route), and every war ends: by tribute, or by itself after a quiet spell."),
		},
		// F10. Each harbinger heard out, and the ones the harbinger bot
		// has afforded to appease: the figures of the first twelve ages.
		{
			Family: "harbinger", Source: BadgeSourceHarbingers,
			Key: "harbinger.{key}.met", Name: "{Name}", Desc: "Hear {name} out.",
			Names: map[string]string{"harbinger.future_self.met": "Your Future Self", "harbinger.unmade_self.met": "Your Unmade Self"},
			Tier:  BadgeBronze, Scope: BadgeMoment, Event: BadgeEvHarbingerMet,
			RevealBySubject: true, Emblem: "harbinger", Set: "harbinger.met",
			Proof: Occurs("An era's doom is fated by a roll at the era's first tick, and a fated doom sends the harbinger of the age its warning falls in. Each era rolls again on every run."),
		},
		{
			Family: "harbinger", Source: BadgeSourceHarbingers,
			Only: []string{"wild_man", "hermit", "soothsayer", "desert_prophet", "oracle", "town_crier",
				"court_astrologer", "pamphleteer", "newsboy", "soapbox_doomsayer", "telegraph", "civil_defence_broadcast"},
			Key: "harbinger.{key}.heeded", Name: "Heeded {name}", Desc: "Appease while {name} is speaking.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvAppeased,
			RevealBySubject: true, Emblem: "faith", Set: "harbinger.heeded",
			Proof: BotProof("harbinger"),
		},
		// F11. Each doom, endured and succumbed to.
		{
			Family: "doom", Source: BadgeSourceDooms,
			Key: "doom.{key}.endured", Name: "Endured: {name}", Desc: "Choose Endure against {mid} and keep going with what is left.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvEndured,
			RevealBySubject: true, Emblem: "hazard", Set: "doom.endured",
			Proof: Occurs("A doom that strikes waits for the answer, and Endure is always one of the two."),
		},
		{
			Family: "doom", Source: BadgeSourceDooms,
			Key: "doom.{key}.succumbed", Name: "Succumbed: {name}", Desc: "Succumb to {mid} and start over with its legacy.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvSuccumbed,
			RevealBySubject: true, Emblem: "ruin", Set: "doom.succumbed",
			Proof: Occurs("A doom that strikes waits for the answer, and Succumb is always one of the two."),
		},
		// F12. Each expedition, come back from with a success.
		{
			Family: "expedition", Source: BadgeSourceExpeditions,
			Key: "expedition.{key}", Name: "{name}", Desc: "Come back from a successful {name}.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvExpedition,
			RevealBySubject: true, Emblem: "military",
			Proof: Occurs("An expedition can be sent whenever its age and its soldiers are there, and every expedition has a chance to succeed that soldiers raise."),
		},
		// F13. Each theme, worn at an advance.
		{
			Family: "theme", Source: BadgeSourceThemes,
			Key: "theme.{key}", Name: "Dressed as {name}", Desc: "Advance an age while wearing the {name} theme.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvAdvancedTheme,
			RevealBySubject: true, Emblem: "monument",
			Proof: Occurs("A theme the account has can be worn at any time, and every run advances."),
		},
		// F15. Each awakening, seen.
		{
			Family: "awakening", Source: BadgeSourceAwakenings,
			Key: "awakening.{key}", Name: "{name}", Desc: "Witness {mid}.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvAwakening,
			RevealBySubject: true, Emblem: "sun",
			Proof: Occurs("An awakening fires by itself in its age, once the run has what it asks for there."),
		},
	}
	return append(fams, badgeLadders()...)
}

// swift is the family of an era left inside its pacing target, for a few
// eras and one proof.
func swift(eras []string, proof BadgeProof) BadgeFamilyDef {
	return BadgeFamilyDef{
		Family: "swift", Source: BadgeSourceEras, Only: eras,
		Key: "swift.{key}", Name: "{name}", Desc: "Leave the {name} within its pacing target.",
		Names: map[string]string{
			"swift.stone_era":    "Out of the Cave Early",
			"swift.iron_era":     "Quick Study",
			"swift.steel_era":    "Ahead of Schedule",
			"swift.electric_era": "Fast Current",
			"swift.digital_era":  "Low Latency",
			"swift.neon_era":     "Skipped the Ads",
			"swift.cosmic_era":   "Speed of Light, Roughly",
		},
		Descs: map[string]string{"swift.cosmic_era": "Prestige out of the last age of the Cosmic Era within the era's pacing target."},
		Tier:  BadgeGold, Scope: BadgeMoment, Event: BadgeEvEraLeft,
		When:            []BadgeCond{{Fact: "ev.pace", Op: BadgeAtMost, Value: 1}},
		RevealBySubject: true,
		Proof:           proof,
		Emblem:          "sun",
	}
}

// ----- the ladders over the account's counters -----

// ladder is a ladder over one counter that belongs to no table. key is its
// place in the badge keys ("ladder.<key>.<n>"), name what the case calls
// it, noun the words of its description ("Endure {count} catastrophes."),
// first the description of its first rung when the count is one.
func ladder(key, name, counter, desc, first string, counts []float64, rungs []string, emblem string, reveal BadgeReveal, proof BadgeProof) BadgeFamilyDef {
	tiers := []BadgeTier{BadgeBronze, BadgeSilver, BadgeGold, BadgeLegendary}
	f := BadgeFamilyDef{
		Family: "ladder", Source: BadgeSourceNone,
		Key: "ladder." + key + ".{n}", Name: "{rung}", Desc: desc,
		Scope: BadgeLifetime, Counter: counter,
		Ladders: map[string][]float64{"": counts},
		Reveal:  reveal, Proof: proof,
		Emblem: emblem, Ladder: name,
	}
	for i, r := range rungs {
		f.Rungs = append(f.Rungs, BadgeRung{Name: r, Tier: tiers[i]})
	}
	if first != "" {
		f.Descs = map[string]string{"ladder." + key + ".1": first}
	}
	return f
}

// badgeLadders returns the twenty-three ladders. The numbers are the
// design's: each rung is a count of runs of ordinary play (a ladder's last
// rung at most twenty-five), which the guard checks against what one run
// adds to the counter.
func badgeLadders() []BadgeFamilyDef {
	visible := BadgeReveal{}
	warned := RevealUntilSeen(BadgeEvDoomNamed)
	heard := RevealUntilSeen(BadgeEvHarbingerMet)
	life := StaticProof(BadgeRuleLifetime)
	out := []BadgeFamilyDef{
		ladder("prestiges", "Prestiges", BadgeEvPrestige, "Prestige {count} times.", "Prestige for the first time.",
			[]float64{1, 3, 10, 25}, []string{"First Prestige", "Creature of Habit", "Serial Reincarnator", "Eternal Return"}, "star", visible, life),
		ladder("endured", "Dooms endured", BadgeEvEndured, "Endure {count} catastrophes.", "Endure a catastrophe.",
			[]float64{1, 5, 15, 30}, []string{"Still Here", "Hard to Kill", "Weatherproof", "Load-Bearing"}, "hazard", warned, life),
		ladder("succumbed", "Dooms succumbed to", BadgeEvSuccumbed, "Succumb to {count} catastrophes.", "Succumb to a catastrophe.",
			[]float64{1, 5, 15}, []string{"Fresh Start", "Serial Quitter", "Professional Ruin"}, "ruin", warned, life),
		ladder("threads", "Warnings resolved", BadgeEvHarbingerResolved, "See {count} harbinger warnings through to their end.", "See a harbinger's warning through to its end.",
			[]float64{1, 10, 40, 100}, []string{"Heard One Out", "Regular Listener", "Open Door Policy", "Complaints Department"}, "harbinger", heard, life),
		ladder("appeased", "Appeasements", BadgeEvAppeased, "Buy {count} appeasements from harbingers.", "Appease a harbinger.",
			[]float64{1, 10, 50}, []string{"Small Offering", "Standing Order", "Patron of Prophets"}, "faith", heard, life),
		ladder("braced", "Braces", BadgeEvBraced, "Brace {count} times against a doom.", "Brace against a doom.",
			[]float64{1, 10, 50}, []string{"Sandbags", "Storm Shutters", "Fortress Mentality"}, "hall", heard, life),
		ladder("invited", "Invitations", BadgeEvInvited, "Invite {count} dooms.", "Invite a doom.",
			[]float64{1, 5, 20}, []string{"Come On Then", "Open Invitation", "Glutton for Punishment"}, "hazard", heard, life),
		ladder("exposed", "False prophets", BadgeEvHarbingerResolved+"."+"discredited", "See {count} false prophets exposed.", "See a false prophet exposed.",
			[]float64{1, 3, 5}, []string{"Called the Bluff", "Fact Checker", "Professional Skeptic"}, "harbinger", BadgeReveal{Kind: BadgeSecret}, life),
		ladder("wonders", "Wonders raised", BadgeEvWonderRaised, "Raise {count} wonders across all your runs.", "",
			[]float64{25, 100, 300}, []string{"Sightseer", "Monument Habit", "Wonder Fatigue"}, "wonder", visible, life),
		ladder("techs", "Techs researched", BadgeEvResearchDone, "Research {count} techs across all your runs.", "",
			[]float64{100, 500, 1200}, []string{"Curious", "Well Read", "Know-It-All"}, "knowledge", visible, life),
		ladder("milestones", "Milestones", BadgeEvMilestone, "Complete {count} milestones across all your runs.", "",
			[]float64{100, 400, 1000}, []string{"Goal Oriented", "Overachiever", "Checklist Enthusiast"}, "star", visible, life),
		ladder("chains", "Milestone chains", BadgeEvChain, "Complete {count} milestone chains across all your runs.", "Complete your first milestone chain.",
			[]float64{1, 10, 30}, []string{"Linked", "Chain Smoker", "Unbroken"}, "star", visible, life),
		ladder("expeditions", "Expeditions", BadgeEvExpedition, "Come back from {count} successful expeditions.", "",
			[]float64{10, 50, 200, 500}, []string{"Day Tripper", "Seasoned Traveler", "Frequent Flyer", "Never Home"}, "military", visible, life),
		ladder("deals", "Faction deals", BadgeEvDeal, "Take {count} deals from civilizations.", "",
			[]float64{5, 25, 100}, []string{"Deal Maker", "Preferred Customer", "Closer"}, "trade", RevealUntilSeen(BadgeEvCivMet), life),
		ladder("raids", "Raids blunted", BadgeEvRaidBlunted, "Have your garrison blunt {count} raids.", "",
			[]float64{10, 100, 500}, []string{"Held the Gate", "Wall of Shields", "Not on My Watch"}, "military", RevealUntilSeen(BadgeEvRaidBlunted), life),
		ladder("returns", "Returns", BadgeEvReturned, "Come back to a game that kept going {count} times.", "Come back to a game that kept going without you.",
			[]float64{1, 25, 100}, []string{"Welcome Back", "Frequent Guest", "Part of the Furniture"}, "housing", visible, StaticProof(BadgeRuleHabit)),
		ladder("plan", "Plan items started", BadgeEvPlanStarted, "Have the build plan start {count} items.", "",
			[]float64{50, 500, 5000}, []string{"Delegator", "Middle Management", "Autopilot"}, "engineer", visible, life),
		ladder("upgrades", "Buildings upgraded", BadgeEvBuildingUpgraded, "Upgrade {count} buildings across all your runs.", "",
			[]float64{250, 2500, 20000}, []string{"Renovator", "Property Developer", "Urban Renewal"}, "engineer", visible, life),
		ladder("festivals", "Festivals", BadgeEvFestival, "Hold {count} festivals.", "Hold a festival.",
			[]float64{1, 10, 50}, []string{"Party Starter", "Social Calendar", "Festival Circuit"}, "culture", RevealUntilSeen(BadgeEvFestival), life),
		ladder("blackmarket", "Black market deals", BadgeEvBlackMarket, "Make {count} deals at the black market.", "Make a deal at the black market.",
			[]float64{1, 10, 50}, []string{"Back Alley", "Known Associate", "Fence"}, "trade", RevealUntilSeen(BadgeEvBlackMarket), life),
		ladder("memories", "Ancient Memories", BadgeEvMemoryAccepted, "Accept {count} Ancient Memories.", "Accept an Ancient Memory.",
			[]float64{1, 5, 12}, []string{"Faint Recollection", "Old Soul", "Total Recall"}, "knowledge", RevealUntilSeen(BadgeEvPrestige), life),
		ladder("hardships", "Hard times weathered", BadgeEvEraEvent+".bad_challenging", "Weather {count} challenging era events.", "Weather a challenging era event.",
			[]float64{1, 10, 30}, []string{"Rough Patch", "Thick Skin", "Calloused"}, "hazard", visible, life),
		ladder("trades", "Market trades", BadgeEvMarketTrade, "Make {count} trades at the market.", "",
			[]float64{50, 500, 5000}, []string{"Haggler", "Day Trader", "Market Maker"}, "trade", RevealUntilSeen(BadgeEvMarketTrade), life),
	}
	for i := range out {
		switch out[i].Ladder {
		case "False prophets":
			out[i].Hint = "Something about a liar."
		case "Dooms endured":
			out[i].Rewards = map[string]BadgeReward{"ladder.endured.3": {Title: "Survivor"}}
		case "Faction deals":
			out[i].Rewards = map[string]BadgeReward{"ladder.deals.3": {Theme: "ledger", Title: "Merchant Prince"}}
			out[i].Descs = map[string]string{"ladder.deals.3": "Take 100 deals from civilizations. Unlocks the Ledger theme."}
		case "Prestiges":
			out[i].Aliases = map[string][]string{
				"ladder.prestiges.1": {"first_prestige"},
				"ladder.prestiges.3": {"prestige_x10"},
			}
		}
	}
	return out
}
