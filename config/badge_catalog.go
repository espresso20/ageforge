package config

// badge_catalog.go is the badge catalog: the families made from the game's
// tables, the ladders over the account's counters, and the hand-written
// specials. badges.go has the types; rules/badges.go expands the families
// when the ruleset is compiled; smoke/static_badges.go proves each badge
// can be earned.
//
// A badge that cannot be shown to be earnable does not ship.

// ----- the tables a family may be made from that live outside config -----

// BadgeExpeditionDef is a badge subject for the missions: one era and one
// kind. Key is MissionBadgeKey, Name the era's name, MinAge the era's first
// age. The missions themselves belong to the game package, which checks this
// table against its own (TestBadgeExpeditionsMatchTheGame).
type BadgeExpeditionDef struct{ Key, Name, MinAge, Kind string }

// MissionBadgeKey is the badge subject of an era's mission of one kind.
func MissionBadgeKey(era, kind string) string { return "expedition_" + era + "_" + kind }

// MissionKindWord is the word a badge uses for a mission kind.
func MissionKindWord(kind string) string {
	switch kind {
	case MissionExpedition:
		return "long"
	case MissionCampaign:
		return "grand"
	}
	return "scout"
}

// BadgeExpeditions lists every era and kind a mission badge exists for, in
// the catalog's order.
func BadgeExpeditions() []BadgeExpeditionDef {
	var out []BadgeExpeditionDef
	for _, era := range Epochs() {
		if len(era.Ages) == 0 {
			continue
		}
		for _, kind := range MissionKinds() {
			out = append(out, BadgeExpeditionDef{MissionBadgeKey(era.Key, kind), era.Name, era.Ages[0], kind})
		}
	}
	return out
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
		{"ledger", "Ledger", "ledger.deals.3"},
		{"prismatic", "Prismatic", "special.museum_piece"},
	}
}

// The glyph sets a badge is worn for. They belong to the UI, which checks
// them against its own (TestBadgeMapLooksMatchTheGame).
var BadgeMapGlyphs = []string{"ascii", "unicode"}

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
		{Title: "Survivor", Badge: "ledger.endured.3"},
		{Title: "Harbinger Whisperer", Badge: "ledger.heeded.5"},
		{Title: "Merchant Prince", Badge: "ledger.deals.3"},
		{Title: "The Undying", Badge: "special.unkillable"},
		{Title: "Cosmic Heir", Badge: "lastpassage.succumbed"},
		{Title: "Hut Magnate", Badge: "special.zoning_board"},
	}
}

// ----- the generated families -----

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
			// The last age has a drawing of its own.
			Emblems: map[string]string{"transcendent_age": "sprite.transcend"},
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
				"primitive_age": 81, "stone_age": 135, "bronze_age": 136, "iron_age": 173, "classical_age": 175,
				"medieval_age": 132, "renaissance_age": 183, "colonial_age": 192, "industrial_age": 163,
				"victorian_age": 162, "electric_age": 203, "atomic_age": 236, "modern_age": 234,
				"information_age": 183, "digital_age": 198, "cyberpunk_age": 201, "fusion_age": 150,
				"space_age": 143, "interstellar_age": 240, "galactic_age": 249, "quantum_age": 259,
				"transcendent_age": 27,
			},
			RevealAtAge: true,
			// Judged as each building finishes: the event is about the
			// building, the badge about its age.
			AnySubject: true,
			Proof:      StaticProof(BadgeRuleAgeCopies),
			Emblem:     "hall",
		},
		// F5. Every tech of an age, in one run.
		{
			Family: "techs", Source: BadgeSourceAges,
			Key: "techs.{key}", Name: "{short} Syllabus", Desc: "{techs}",
			Tier:  BadgeSilver,
			Scope: BadgeRun, Event: BadgeEvResearchDone, Counter: "researched_age.{key}",
			Measure:     BadgeMeasureTechs,
			RevealAtAge: true,
			// Judged as each tech finishes: the event is about the tech, the
			// badge about its age.
			AnySubject: true,
			Proof:      StaticProof(BadgeRuleTechs),
			Emblem:     "knowledge",
		},
		// F10. Each harbinger heard out: the figures of every age.
		{
			Family: "harbinger", Source: BadgeSourceHarbingers,
			Key: "harbinger.{key}.met", Name: "{Name}", Desc: "Hear {name} out.",
			Names: map[string]string{"harbinger.future_self.met": "Your Future Self", "harbinger.unmade_self.met": "Your Unmade Self"},
			Tier:  BadgeBronze, Scope: BadgeMoment, Event: BadgeEvHarbingerMet,
			RevealBySubject: true, Emblem: "harbinger", Set: "harbinger.met",
			Proof: Occurs("An era's doom is fated by a roll at the era's first tick, and a fated doom sends the harbinger of the age its warning falls in. Each era rolls again on every run."),
		},
		// F11. Each doom, endured and succumbed to.
		{
			Family: "doom", Source: BadgeSourceDooms,
			Key: "doom.{key}.endured", Name: "Endured: {name}", Desc: "Choose Endure against {mid} and keep going with what is left.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvEndured,
			RevealBySubject: true, Emblem: "hazard",
			Proof: Occurs("A doom that strikes waits for the answer, and Endure is always one of the two."),
		},
		{
			Family: "doom", Source: BadgeSourceDooms,
			Key: "doom.{key}.succumbed", Name: "Succumbed: {name}", Desc: "Succumb to {mid} and start over with its legacy.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvSuccumbed,
			RevealBySubject: true, Emblem: "ruin",
			Proof: Occurs("A doom that strikes waits for the answer, and Succumb is always one of the two."),
		},
		// F12. Each era and kind of mission, come back from with a success.
		{
			Family: "expedition", Source: BadgeSourceExpeditions,
			Key: "expedition.{key}", Name: "{name} {Word}", Desc: "Come back from a successful {word} mission in the {name}.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvExpedition,
			RevealBySubject: true, Emblem: "military",
			Proof: Occurs("A mission of every era can be sent whenever its age and its lock are open, and every mission has a chance to succeed."),
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
	return fams
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
