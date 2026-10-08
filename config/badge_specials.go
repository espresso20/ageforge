package config

// badge_specials.go holds the hand-written badges: the fixed rows of the
// families that are too small or too odd for a template (the Last Passage,
// the map's looks), the specials, the integrity badges, and the rows held
// out of the catalog until they can be proven.

// A few conditions the rows below repeat.
func atLeast(fact string, v float64) BadgeCond {
	return BadgeCond{Fact: fact, Op: BadgeAtLeast, Value: v}
}
func never(fact string) BadgeCond { return BadgeCond{Fact: fact, Op: BadgeAtMost, Value: 0} }

var (
	secret = BadgeReveal{Kind: BadgeSecret}
	// The reveals of the badges that show once their subject has come up.
	onHarbinger = RevealUntilSeen(BadgeEvHarbingerMet)
	onWar       = RevealUntilSeen(BadgeEvWarStarted)
	onCiv       = RevealUntilSeen(BadgeEvCivMet)
)

// Badges returns the hand-written badges, in the order they list.
func Badges() []BadgeDef {
	out := []BadgeDef{
		// ----- F11, the Last Passage -----
		{
			Key: "lastpassage.endured", Family: "doom", Subject: "endured",
			Name: "Endured: The Last Passage", Desc: "Choose Endure at the Last Passage and prestige with what it leaves you.",
			Tier: BadgePlatinum, Scope: BadgeMoment, Event: BadgeEvLastPassage,
			Reveal: RevealUntilAge("interstellar_age"), Emblem: "hazard", Set: "doom.endured",
			Proof: Occurs("The Last Passage that comes through waits for the answer, and Endure is always one of the two."),
		},
		{
			Key: "lastpassage.succumbed", Family: "doom", Subject: "succumbed",
			Name: "Succumbed: The Last Passage", Desc: "Succumb at the Last Passage and carry its legacy into every run after.",
			Tier: BadgeLegendary, Scope: BadgeMoment, Event: BadgeEvLastPassage,
			Reveal: RevealUntilAge("interstellar_age"), Emblem: "ruin",
			Reward: BadgeReward{Title: "Cosmic Heir"},
			Proof:  Occurs("The Last Passage that comes through waits for the answer, and Succumb is always one of the two."),
		},
		{
			Key: "lastpassage.spared", Family: "doom", Subject: "spared",
			Name: "Spared: The Last Passage", Desc: "Prestige out of the Cosmic Era and have the Last Passage open on nothing.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvLastPassage,
			Reveal: RevealUntilAge("interstellar_age"), Emblem: "sun",
			Proof: Occurs("A prestige from the Cosmic Era opens the Last Passage, and a passage that was not invited does not always come through."),
		},

		// ----- F14, the map's looks -----
		mapLook("map.style.roguelike", "Dungeon Crawler", "Advance an age with the roguelike map style.", BadgeEvAdvancedStyle, "roguelike"),
		mapLook("map.style.skyline", "Skyline Watcher", "Advance an age with the skyline map style.", BadgeEvAdvancedStyle, "skyline"),
		mapLook("map.glyphs.ascii", "Plain Speaker", "Advance an age with the plain (ascii) map glyphs.", BadgeEvAdvancedGlyphs, "ascii"),
		mapLook("map.glyphs.unicode", "Box Drawer", "Advance an age with the unicode map glyphs.", BadgeEvAdvancedGlyphs, "unicode"),
		mapLook("map.glyphs.nerd", "Font Snob", "Advance an age with the Nerd Font map glyphs.", BadgeEvAdvancedGlyphs, "nerd"),

		// ----- flex -----
		{
			Key: "special.hut_hoarder", Family: "special", Subject: "hut",
			Name: "Hut Hoarder",
			// The prices quoted are checked against the cost curve
			// (TestHutHoarderQuotesRealPrices).
			Desc: "Have 60 huts standing before you leave the Primitive Age. The 60th costs 18,956 wood, 1,354 times the first.",
			Tier: BadgeGold, Rarity: BadgeEpic,
			Scope: BadgeRun, Event: BadgeEvBuildingBuilt, InAge: "primitive_age",
			Counter: "standing.hut", Threshold: 60,
			Emblem: "hut",
			Proof:  StaticProof(BadgeRuleCopies),
		},
		{
			Key: "special.zoning_board", Family: "special", Subject: "hut",
			Name: "Zoning Board",
			Desc: "Have 62 huts standing in the Primitive Age. There is no 63rd: the storage will not hold its price.",
			Hint: "Something about huts.",
			Tier: BadgePlatinum, Rarity: BadgeMythic,
			Scope: BadgeRun, Event: BadgeEvBuildingBuilt, InAge: "primitive_age",
			Counter: "standing.hut", Threshold: 62,
			Reveal: secret, Emblem: "hut",
			Reward: BadgeReward{Title: "Hut Magnate"},
			Proof:  StaticProof(BadgeRuleCeiling),
		},
		{
			Key: "special.every_nook", Family: "special", Subject: "stash",
			Name:  "Every Nook",
			Desc:  "Have 50 stashes standing before you leave the Primitive Age.",
			Tier:  BadgeSilver,
			Scope: BadgeRun, Event: BadgeEvBuildingBuilt, InAge: "primitive_age",
			Counter: "standing.stash", Threshold: 50,
			Emblem: "store",
			Proof:  StaticProof(BadgeRuleCeiling),
		},

		// ----- challenges -----
		{
			Key: "special.nothing_to_endure", Family: "special", Subject: "modern_age",
			Name: "Nothing to Endure", Desc: "Reach the Modern Age without choosing Endure once in the run.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvAgeReached,
			When:   []BadgeCond{never("run." + BadgeEvEndured)},
			Reveal: RevealUntilNextAge("modern_age"), Emblem: "sun",
			Proof: BotProof("greedy"),
		},
		{
			Key: "special.hands_on", Family: "special", Subject: "industrial_age",
			Name: "Hands On", Desc: "Reach the Industrial Age without the build plan starting anything.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvAgeReached,
			When:   []BadgeCond{never("run." + BadgeEvPlanStarted)},
			Reveal: RevealUntilNextAge("industrial_age"), Emblem: "engineer",
			Proof: Occurs("The build plan is a convenience: everything it starts can be started by hand, and nothing asks for it."),
		},
		{
			Key: "special.exact_change", Family: "special",
			Name: "Exact Change", Desc: "Raise five wonders in one run without overflow putting anything into a wonder's bank.",
			Tier: BadgeSilver, Scope: BadgeRun, Event: BadgeEvWonderRaised,
			Counter: "run." + BadgeEvWonderRaised, Threshold: 5,
			When:   []BadgeCond{never("run." + BadgeEvWonderOverflow)},
			Emblem: "wonder",
			Proof:  StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.family_heirlooms", Family: "special", Subject: "iron_age",
			Name: "Family Heirlooms", Desc: "Leave the Bronze Age without upgrading a single building in the run.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvAgeReached,
			When:   []BadgeCond{never("run." + BadgeEvBuildingUpgraded)},
			Reveal: RevealUntilAge("bronze_age"), Emblem: "housing",
			Proof: Occurs("Upgrading is a command the player gives; no advance asks for it."),
		},
		{
			Key: "special.express_lane", Family: "special", Subject: "modern_age",
			Name: "Express Lane", Desc: "Reach the Modern Age within four fifths of the combined pacing target of the twelve ages before it.",
			Tier: BadgePlatinum, Scope: BadgeMoment, Event: BadgeEvAgeReached,
			When:   []BadgeCond{{Fact: "ev.pace", Op: BadgeAtMost, Value: 0.8}},
			Reveal: RevealUntilNextAge("modern_age"), Emblem: "sun",
			Proof: BotProof("veteran"),
		},
		{
			Key: "special.early_riser", Family: "special", Subject: "iron_age",
			Name: "Early Riser", Desc: "Reach the Iron Age within the combined pacing target of the three ages before it.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvAgeReached,
			When:   []BadgeCond{{Fact: "ev.pace", Op: BadgeAtMost, Value: 1}},
			Reveal: RevealUntilNextAge("iron_age"), Emblem: "sun",
			Proof: BotProof("veteran"),
		},
		{
			Key: "special.fashionably_late", Family: "special",
			Name:  "Fashionably Late",
			Desc:  "Spend ten times the Primitive Age's pacing target in the Primitive Age.",
			Tier:  BadgeBronze,
			Scope: BadgeRun, Event: BadgeEvTick, InAge: "primitive_age",
			Pred: BadgePredAgeOverstay, Threshold: 10,
			Emblem: "sun",
			Proof:  StaticProof(BadgeRuleTime),
		},

		// ----- catastrophes and harbingers -----
		{
			Key: "special.sandbagged", Family: "special",
			Name: "Sandbagged", Desc: "Endure a catastrophe with Brace at level 2.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvEndured,
			When:   []BadgeCond{atLeast("ev.brace", 2)},
			Reveal: onHarbinger, Emblem: "hall",
			Proof: BotProof("harbinger"),
		},
		{
			Key: "special.bare_hands", Family: "special",
			Name: "Bare Hands", Desc: "Endure a catastrophe with no Brace and no garrison.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvEndured,
			When:   []BadgeCond{never("ev.brace"), never("ev.garrison")},
			Reveal: onHarbinger, Emblem: "hazard",
			Proof: Occurs("Brace is bought and soldiers are trained by choice; a run that does neither still gets the Endure choice."),
		},
		{
			Key: "special.standing_guard", Family: "special",
			Name: "Standing Guard", Desc: "Have the garrison save buildings during an Endure.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvEndured,
			When:   []BadgeCond{atLeast("ev.saved", 1)},
			Reveal: onHarbinger, Emblem: "military",
			Proof: BotProof("army"),
		},
		{
			Key: "special.asked_for_it", Family: "special",
			Name: "Asked For It", Desc: "Invite a catastrophe, then Endure it.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvEndured,
			When:   []BadgeCond{atLeast("ev.invited", 1)},
			Reveal: onHarbinger, Emblem: "hazard",
			Proof: Occurs("Invite makes the doom certain, and the doom that strikes offers Endure."),
		},
		{
			Key: "special.signed_the_guest_book", Family: "special",
			Name: "Signed the Guest Book", Desc: "Invite a catastrophe, then Succumb to it.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvSuccumbed,
			When:   []BadgeCond{atLeast("ev.invited", 1)},
			Reveal: onHarbinger, Emblem: "ruin",
			Proof: Occurs("Invite makes the doom certain, and the doom that strikes offers Succumb."),
		},
		{
			Key: "special.paid_in_full", Family: "special", Subject: "vindicated",
			Name: "Paid in Full", Desc: "Appease a harbinger twice and have the catastrophe come anyway.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{atLeast("ev.appease", 2)},
			Reveal: onHarbinger, Emblem: "faith",
			Proof: BotProof("harbinger"),
		},
		{
			Key: "special.tithe_to_a_liar", Family: "special", Subject: "discredited",
			Name: "Tithe to a Liar", Desc: "Appease a harbinger who turns out to have made the whole thing up.",
			Hint: "Something about a wasted offering.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{atLeast("ev.appease", 1)},
			Reveal: secret, Emblem: "faith",
			Proof: BotProof("harbinger"),
		},
		{
			Key: "special.self_fulfilling", Family: "special", Subject: "fulfilled",
			Name: "Self-Fulfilling", Desc: "Invite doom on a false prophet's word and get it anyway.",
			Hint: "Something about asking for trouble.",
			Tier: BadgePlatinum, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{atLeast("ev.false_prophet", 1)},
			Reveal: secret, Emblem: "harbinger",
			Proof: Occurs("A false prophet can be invited like any harbinger, and an invited doom always strikes, invented or not."),
		},
		{
			Key: "special.prepared_for_nothing", Family: "special", Subject: "spared",
			Name: "Prepared for Nothing", Desc: "Brace to level 2 and be spared.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{atLeast("ev.brace", 2)},
			Reveal: onHarbinger, Emblem: "hall",
			Proof: BotProof("harbinger"),
		},
		{
			Key: "special.full_chorus", Family: "special",
			Name: "Full Chorus", Desc: "Hear all three voices of one era's warning in a single thread.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{atLeast("ev.figures", 3)},
			Reveal: onHarbinger, Emblem: "harbinger",
			Proof: BotProof("greedy"),
		},
		{
			Key: "special.selective_hearing", Family: "special", Subject: "spared",
			Name: "Selective Hearing", Desc: "Ignore a harbinger completely and be spared.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
			When:   []BadgeCond{never("ev.appease"), never("ev.brace"), never("ev.invited")},
			Reveal: onHarbinger, Emblem: "harbinger",
			Proof: BotProof("greedy"),
		},

		// ----- the Last Passage and Cosmic Legacy -----
		{
			Key: "special.knock_knock", Family: "special",
			Name: "Knock Knock", Desc: "Invite the Last Passage.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvInvited,
			When:   []BadgeCond{atLeast("ev.last", 1)},
			Reveal: RevealUntilAge("interstellar_age"), Emblem: "hazard",
			Proof: Occurs("The Cosmic Era's harbingers speak for the Last Passage, and Invite is free."),
		},
		{
			Key: "special.walked_out_whole", Family: "special", Subject: "endured",
			Name: "Walked Out Whole", Desc: "Endure the Last Passage with Brace at level 2.",
			Tier: BadgePlatinum, Scope: BadgeMoment, Event: BadgeEvLastPassage,
			When:   []BadgeCond{atLeast("ev.brace", 2)},
			Reveal: RevealUntilAge("interstellar_age"), Emblem: "hall",
			Proof: BotProof("cosmic"),
		},
		{
			Key: "special.heirloom_universe", Family: "special",
			Name: "Heirloom Universe", Desc: "Prestige five times while holding Cosmic Legacy.",
			Tier: BadgePlatinum, Scope: BadgeLifetime, Counter: BadgeEvLegacyPrestige, Threshold: 5,
			Reveal: RevealUntilSeen(BadgeEvLastPassage + ".succumbed"), Emblem: "star",
			Proof: StaticProof(BadgeRuleLifetime),
		},

		// ----- army, diplomacy and trade -----
		{
			Key: "special.not_today", Family: "special",
			Name: "Not Today", Desc: "Have your garrison blunt 30 raids in one run.",
			Tier: BadgeSilver, Scope: BadgeRun, Event: BadgeEvRaidBlunted,
			Counter: "run." + BadgeEvRaidBlunted, Threshold: 30,
			Reveal: RevealUntilSeen(BadgeEvRaidBlunted), Emblem: "military",
			Proof: BotProof("army"),
		},
		{
			Key: "special.protection_money", Family: "special",
			Name: "Protection Money", Desc: "End a war by paying tribute.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvWarEnded,
			When:   []BadgeCond{atLeast("ev.tribute", 1)},
			Reveal: onWar, Emblem: "trade",
			Proof: Occurs("Tribute ends any war at once, for gold and culture."),
		},
		{
			Key: "special.cold_shoulder", Family: "special",
			Name: "Cold Shoulder", Desc: "Let a war burn out without paying anyone.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvWarEnded,
			When:   []BadgeCond{never("ev.tribute")},
			Reveal: onWar, Emblem: "war",
			Proof: Occurs("A war with no new provocation ends by itself after its quiet spell."),
		},
		{
			Key: "special.popular", Family: "special",
			Name: "Popular", Desc: "Be at war with two civilizations at once.",
			Hint: "Something about making enemies.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvWarStarted,
			When:   []BadgeCond{atLeast("ev.wars", 2)},
			Reveal: secret, Emblem: "war",
			Proof: Occurs("Each civilization goes to war on its own opinion and provocations; nothing stops two at once."),
		},
		{
			Key: "special.everybodys_friend", Family: "special",
			Name: "Everybody's Friend", Desc: "Be allied with every civilization you have met, at least four of them.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvCivAllied,
			When:   []BadgeCond{atLeast("ev.met", 4), never("ev.unallied")},
			Reveal: onCiv, Emblem: "diplomacy",
			Proof: Occurs("An alliance takes opinion, which gifts buy, and gold; four civilizations are met by the Industrial Age."),
		},
		{
			Key: "special.regifting", Family: "special",
			Name: "Regifting", Desc: "Receive workers on loan from a civilization.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvWorkersLent,
			Reveal: onCiv, Emblem: "civ",
			Proof: Occurs("A friendly civilization at peace lends workers by itself, on a roll it makes as the ages pass."),
		},
		{
			Key: "special.loyal_customer", Family: "special",
			Name: "Loyal Customer", Desc: "Take 10 deals from one civilization in one run.",
			Tier: BadgeSilver, Scope: BadgeRun, Event: BadgeEvDeal,
			Counter: "run." + BadgeEvDeal + ".*", Threshold: 10,
			Reveal: RevealUntilSeen(BadgeEvDeal), Emblem: "trade",
			Proof: StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.no_questions_asked", Family: "special",
			Name: "No Questions Asked", Desc: "Make a deal at the black market.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvBlackMarket,
			Reveal: RevealUntilAge("colonial_age"), Emblem: "trade",
			Proof: Occurs("The black market opens with its tech and takes any stake."),
		},

		// ----- idle, check-in and the build plan -----
		{
			Key: "special.while_you_were_out", Family: "special",
			Name: "While You Were Out", Desc: "Come back after the full offline allowance to find that the build plan started 10 or more items and has nothing left.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvReturned,
			When:   []BadgeCond{atLeast("ev.capped", 1), atLeast("ev.plan_started", 10), never("ev.plan_left")},
			Emblem: "engineer",
			Proof:  BotProof("idle"),
		},
		{
			Key: "special.maximum_leave", Family: "special",
			Name: "Maximum Leave", Desc: "Stay away past the offline allowance. The game stopped counting before you came back.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvReturned,
			When:   []BadgeCond{atLeast("ev.capped", 1)},
			Emblem: "housing",
			Proof:  Occurs("Time away past the allowance is capped, and the return says so."),
		},
		{
			Key: "special.regular", Family: "special",
			Name: "Regular", Desc: "Play on 7 different days within 10 days.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvDayPlayed,
			When:   []BadgeCond{atLeast("acct.days_within.10", 7)},
			Emblem: "sun",
			Proof:  StaticProof(BadgeRuleHabit),
		},
		{
			Key: "special.standing_appointment", Family: "special",
			Name: "Standing Appointment", Desc: "Play on 30 different days.",
			Tier: BadgeGold, Scope: BadgeLifetime, Counter: BadgeEvDayPlayed, Threshold: 30,
			Emblem: "sun",
			Proof:  StaticProof(BadgeRuleHabit),
		},
		{
			Key: "special.nothing_wasted", Family: "special",
			Name: "Nothing Wasted", Desc: "Let overflow finish banking a wonder while you are away.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvWonderBanked,
			When:   []BadgeCond{atLeast("ev.away", 1)},
			Emblem: "wonder",
			Proof:  BotProof("idle"),
		},
		{
			Key: "special.long_game", Family: "special",
			Name: "Long Game", Desc: "Have 20 items waiting in the build plan at once.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvPlanQueued,
			When:   []BadgeCond{atLeast("ev.size", 20)},
			Emblem: "engineer",
			Proof:  Occurs("The build plan holds 60 items, and adding one is a command."),
		},
		{
			Key: "special.absentee_landlord", Family: "special",
			Name: "Absentee Landlord", Desc: "Prestige in a run where you spent at least three times as long away as you did playing. Only credited time away counts.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvPrestige,
			Pred: BadgePredMostlyAway, Threshold: 3,
			Emblem: "housing",
			Proof:  BotProof("idle"),
		},

		// ----- milestones and the case itself -----
		{
			Key: "special.chain_reaction", Family: "special",
			Name: "Chain Reaction", Desc: "Complete a milestone chain.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvChain,
			Emblem: "star",
			Proof:  Occurs("Every milestone of every chain is proven completable by the milestone check."),
		},
		{
			Key: "special.six_for_six", Family: "special",
			Name: "Six for Six", Desc: "Complete all six milestone chains in one run.",
			Tier: BadgeLegendary, Scope: BadgeRun, Event: BadgeEvChain,
			Counter: "run." + BadgeEvChain, Threshold: 6,
			Emblem: "sprite.chains",
			Proof:  StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.box_ticker", Family: "special",
			Name: "Box Ticker", Desc: "Complete every milestone in one run.",
			Hint: "Something about finishing things.",
			Tier: BadgeLegendary, Scope: BadgeRun, Event: BadgeEvMilestone,
			Counter: "run." + BadgeEvMilestone, Measure: BadgeMeasureMilestones,
			Reveal: secret, Emblem: "sprite.boxes",
			Proof: StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.collector", Family: "special",
			Name: "Collector", Desc: "Earn 25 badges.",
			Tier: BadgeBronze, Scope: BadgeLifetime, Counter: BadgeEvBadge, Threshold: 25,
			Emblem: "star", Ladder: "Badges earned", Proof: DerivedProof(),
		},
		{
			Key: "special.curator", Family: "special",
			Name: "Curator", Desc: "Earn 100 badges.",
			Tier: BadgeSilver, Scope: BadgeLifetime, Counter: BadgeEvBadge, Threshold: 100,
			Emblem: "star", Ladder: "Badges earned", Proof: DerivedProof(),
		},
		{
			Key: "special.archivist", Family: "special",
			Name: "Archivist", Desc: "Earn 250 badges.",
			Tier: BadgeGold, Scope: BadgeLifetime, Counter: BadgeEvBadge, Threshold: 250,
			Emblem: "star", Ladder: "Badges earned", Proof: DerivedProof(),
		},
		{
			Key: "special.museum_piece", Family: "special",
			Name: "Museum Piece", Desc: "Earn 400 badges. Unlocks the Prismatic theme.",
			Tier: BadgeLegendary, Scope: BadgeLifetime, Counter: BadgeEvBadge, Threshold: 400,
			Emblem: "sprite.museum", Ladder: "Badges earned",
			Reward: BadgeReward{Theme: "prismatic"},
			Proof:  DerivedProof(),
		},
		{
			Key: "special.full_set", Family: "special", Subject: "lineage",
			Name: "Full Set", Desc: "Finish every rung of one lineage ladder.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvBadge,
			When:   []BadgeCond{atLeast("ev.top", 1)},
			Emblem: "star", Proof: DerivedProof(),
		},
		{
			Key: "special.twenty_two_wonders", Family: "special",
			Name: "Twenty-Two Wonders", Desc: "Raise every wonder at least once.",
			Tier: BadgeLegendary, Scope: BadgeLifetime, Counter: BadgeEvBadge + ".set.wonder", Measure: BadgeMeasureSet,
			Emblem: "sprite.wonders", Proof: DerivedProof(),
		},
		{
			Key: "special.rogues_gallery", Family: "special",
			Name: "Rogues' Gallery", Desc: "Hear out every harbinger there is.",
			Tier: BadgeLegendary, Scope: BadgeLifetime, Counter: BadgeEvBadge + ".set.harbinger.met", Measure: BadgeMeasureSet,
			Reveal: onHarbinger, Emblem: "sprite.rogues", Proof: DerivedProof(),
		},
		{
			Key: "special.small_world", Family: "special",
			Name: "Small World", Desc: "Meet every civilization there is.",
			Tier: BadgePlatinum, Scope: BadgeLifetime, Counter: BadgeEvBadge + ".set.civ.met", Measure: BadgeMeasureSet,
			Reveal: onCiv, Emblem: "civ", Proof: DerivedProof(),
		},
		{
			Key: "special.connoisseur_of_endings", Family: "special",
			Name: "Connoisseur of Endings", Desc: "Succumb to every era's catastrophe. Unlocks the Ashfall theme.",
			Tier: BadgeLegendary, Scope: BadgeLifetime, Counter: BadgeEvBadge + ".set.doom.succumbed", Measure: BadgeMeasureSet,
			Reveal: RevealUntilSeen(BadgeEvSuccumbed), Emblem: "sprite.endings",
			Reward: BadgeReward{Theme: "ashfall"},
			Proof:  DerivedProof(),
		},
		{
			Key: "special.unkillable", Family: "special",
			Name: "Unkillable", Desc: "Endure every era's catastrophe and the Last Passage.",
			Tier: BadgeLegendary, Scope: BadgeLifetime, Counter: BadgeEvBadge + ".set.doom.endured", Measure: BadgeMeasureSet,
			Reveal: RevealUntilSeen(BadgeEvEndured), Emblem: "sprite.undying",
			Reward: BadgeReward{Title: "The Undying"},
			Proof:  DerivedProof(),
		},

		// ----- odds and ends -----
		{
			Key: "special.portion_control", Family: "special",
			Name: "Portion Control", Desc: "Lose a worker to starvation.",
			Hint: "Something about dinner.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvStarved,
			Reveal: secret, Emblem: "lineage.food",
			Proof: Occurs("Workers starve when food runs out, and food runs out when more workers eat than farms feed."),
		},
		{
			Key: "special.grumbling", Family: "special",
			Name: "Grumbling", Desc: "Let morale fall far enough that the game warns you about it.",
			Hint: "Something about the mood.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvLowMorale,
			Reveal: secret, Emblem: "culture",
			Proof: Occurs("Morale falls with hunger, crowding and catastrophe, and the game warns below its low mark."),
		},
		{
			Key: "special.every_day_a_holiday", Family: "special",
			Name: "Every Day Is a Holiday", Desc: "Hold 10 festivals in one run.",
			Tier: BadgeSilver, Scope: BadgeRun, Event: BadgeEvFestival,
			Counter: "run." + BadgeEvFestival, Threshold: 10,
			Reveal: RevealUntilSeen(BadgeEvFestival), Emblem: "culture",
			Proof: StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.liquidation_sale", Family: "special",
			Name:  "Liquidation Sale",
			Desc:  "Sell 100 buildings in one run.",
			Hint:  "Something about a clearance.",
			Tier:  BadgeBronze,
			Scope: BadgeRun, Event: BadgeEvBuildingSold,
			Counter: "run." + BadgeEvBuildingSold, Threshold: 100,
			Reveal: secret, Emblem: "trade",
			Proof: StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.artisanal", Family: "special",
			Name: "Artisanal", Desc: "Gather by hand 1,000 times in one run.",
			Tier: BadgeBronze, Scope: BadgeRun, Event: BadgeEvGathered,
			Counter: "run." + BadgeEvGathered, Threshold: 1000,
			Emblem: "lineage.organic_extraction",
			Proof:  StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.forgetful", Family: "special",
			Name: "Forgetful", Desc: "Turn down an Ancient Memory.",
			Hint: "Something about saying no.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvMemoryDeclined,
			Reveal: secret, Emblem: "knowledge",
			Proof: Occurs("An Ancient Memory is offered after a prestige, and it can be declined."),
		},
		{
			Key: "special.deja_vu", Family: "special",
			Name: "Déjà Vu", Desc: "Accept an Ancient Memory.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvMemoryAccepted,
			Reveal: RevealUntilSeen(BadgeEvPrestige), Emblem: "knowledge",
			Proof: Occurs("An Ancient Memory is offered after a prestige, and it can be accepted."),
		},
		{
			Key: "special.belt_and_braces", Family: "special",
			Name: "Belt and Braces", Desc: "Export an account backup.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvExported,
			Emblem: "store",
			Proof:  Occurs("account export is a command."),
		},
		{
			Key: "special.written_down_somewhere", Family: "special",
			Name: "Written Down Somewhere", Desc: "Look at your recovery code.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvRecoveryShown,
			Emblem: "knowledge",
			Proof:  Occurs("account is a command, and it shows the code."),
		},
		{
			Key: "special.maxed_out", Family: "special",
			Name: "Maxed Out", Desc: "Buy an item of the legacy kit.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvKit,
			Reveal: RevealUntilSeen(BadgeEvPrestige), Emblem: "star",
			Proof: Occurs("The cheapest kit item costs the points of one prestige from the Medieval Age."),
		},
		{
			Key: "special.fully_invested", Family: "special",
			Name: "Fully Invested", Desc: "Own the whole legacy kit.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvKit,
			When:   []BadgeCond{never("ev.left")},
			Reveal: RevealUntilSeen(BadgeEvPrestige), Emblem: "star",
			Proof: Occurs("The whole kit costs less than the points of one prestige from the Modern Age."),
		},
		{
			Key: "special.lucky_break", Family: "special", Subject: "good_legendary",
			Name: "Lucky Break", Desc: "Have an era open with a legendary good event.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvEraEvent,
			Emblem: "star",
			Proof:  Occurs("Every era opens with an event, and a legendary good one is on the roll."),
		},
		{
			Key: "special.character_building", Family: "special", Subject: "bad_challenging",
			Name: "Character Building", Desc: "Weather three challenging era events in one run.",
			Tier: BadgeSilver, Scope: BadgeRun, Event: BadgeEvEraEvent,
			Counter: "run." + BadgeEvEraEvent + ".bad_challenging", Threshold: 3,
			Emblem: "hazard",
			Proof:  StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.everything_full", Family: "special",
			Name: "Everything Full", Desc: "Have every resource you can store at its storage limit at once, in the Industrial Age or later. A resource that is spent as it is made does not count.",
			Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvCensus,
			When:   []BadgeCond{atLeast("ev.full", 1), atLeast("ev.age", 8)},
			Reveal: RevealUntilNextAge("industrial_age"), Emblem: "store",
			Proof: Occurs("Every capped resource has a producer by its age, and a store that is left alone fills."),
		},
		{
			Key: "special.old_bones", Family: "special",
			Name: "Old Bones", Desc: "Start a run carrying ruins from a civilization that fell.",
			Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvRuinsCarried,
			Reveal: RevealUntilSeen(BadgeEvSuccumbed), Emblem: "ruin",
			Proof: Occurs("A run that ends leaves its buildings as ruins, and the next run starts among them."),
		},
		{
			Key: "special.we_are_not_alone", Family: "special",
			Name: "We Are Not Alone", Desc: "Inspect a visitor from somewhere else on the map.",
			Hint: "Something about the sky.",
			Tier: BadgeSilver, Scope: BadgeMoment, Event: BadgeEvVisitor,
			Reveal: secret, Emblem: "alien",
			Proof: Occurs("From the Space Age a visitor crosses the open map about once every ninety minutes, and now and then before it, with motion on or off. The cursor can stand on it."),
		},

		// ----- integrity: worth nothing, counted in nothing -----
		{
			Key: "special.hand_in_the_cookie_jar", Family: "special",
			Name:  "Hand in the Cookie Jar",
			Desc:  "Unlock the developer console.",
			Scope: BadgeMoment, Event: BadgeEvDevUnlocked,
			Reveal: secret,
			Emblem: "cookie_jar",
			Proof:  BadgeProof{Kind: BadgeProofIntegrity},
		},
		{
			Key: "special.touched_by_the_source", Family: "special",
			Name:  "Touched by the Source",
			Desc:  "Load a save that carries the forge master's proof.",
			Scope: BadgeMoment, Event: BadgeEvSaveElite,
			Reveal: secret,
			Emblem: "source",
			Reward: BadgeReward{Theme: "source"},
			Proof:  BadgeProof{Kind: BadgeProofIntegrity},
		},
		{
			Key: "special.creative_accounting", Family: "special",
			Name:  "Creative Accounting",
			Desc:  "Load a save that was edited outside the game.",
			Scope: BadgeMoment, Event: BadgeEvSaveModified,
			Reveal: secret,
			Emblem: "ledger",
			Reward: BadgeReward{Theme: "glitch"},
			Proof:  BadgeProof{Kind: BadgeProofIntegrity},
		},
	}
	return out
}

// mapLook is a badge for a map style or a glyph set worn at an advance.
func mapLook(key, name, desc, event, subject string) BadgeDef {
	return BadgeDef{
		Key: key, Family: "map", Subject: subject, Name: name, Desc: desc,
		Tier: BadgeBronze, Scope: BadgeMoment, Event: event,
		Emblem: "hall",
		Proof:  Occurs("The map's style and glyphs are settings, and every run advances."),
	}
}

// BadgeHeldOutDef is a badge kept out of the catalog, and why.
type BadgeHeldOutDef struct {
	Badge BadgeDef
	// Why says what proof it waits for.
	Why string
}

// BadgesHeldOut returns the badges that are written and not shipped: each
// is a row the catalog can take the day it can be proven. None of them is
// in the ruleset, so none can be earned, shown or counted.
//
// A badge is here for one of two reasons. Its only proof is a way of
// playing that no smoke bot plays (a pacifist, an autarky) and nothing in
// the game's tables can stand in for it. Or the game cannot produce what it
// asks for as it stands.
func BadgesHeldOut() []BadgeHeldOutDef {
	return []BadgeHeldOutDef{
		{
			Badge: BadgeDef{
				Key: "special.conscientious_objector", Family: "special",
				Name: "Conscientious Objector", Desc: "Prestige without training a single soldier in the run.",
				Tier: BadgeGold, Scope: BadgeMoment, Event: BadgeEvPrestige,
				Emblem: "military", Proof: BotProof("pacifist"),
			},
			Why: "Needs a pacifist bot: the early gates ask for military buildings, and whether a run can hold them unstaffed all the way to a prestige is a question of play, not of the tables. It also needs a count of soldiers trained in the run.",
		},
		{
			Badge: BadgeDef{
				Key: "special.autarky", Family: "special", Subject: "modern_age",
				Name: "Autarky", Desc: "Reach the Modern Age without the market, a trade route or a faction deal.",
				Tier: BadgePlatinum, Rarity: BadgeEpic, Scope: BadgeMoment, Event: BadgeEvAgeReached,
				Emblem: "trade", Proof: BotProof("autarky"),
			},
			Why: "Needs an autarky bot: the gate check proves every gate has a source, and the market is one of the sources it accepts, so the tables do not show the run can be made without it.",
		},
		{
			Badge: BadgeDef{
				Key: "special.stopped_clock", Family: "special", Subject: "vindicated",
				Name: "Stopped Clock", Desc: "A false prophet's catastrophe arrives, uninvited. The lie was accurate.",
				Hint: "Something about being right by accident.",
				Tier: BadgeLegendary, Scope: BadgeMoment, Event: BadgeEvHarbingerResolved,
				When:   []BadgeCond{atLeast("ev.false_prophet", 1), never("ev.invited")},
				Reveal: secret, Emblem: "sprite.clock", Proof: Occurs(""),
			},
			Why: "Cannot happen: a false prophet comes only to an era whose doom is not fated, and an unfated doom strikes only when it is invited.",
		},
		{
			Badge: BadgeDef{
				Key: "special.note_to_self", Family: "special", Subject: "future_self",
				Name: "Note to Self", Desc: "Appease while your future self is speaking.",
				Hint: "Something about a familiar voice.",
				Tier: BadgePlatinum, Rarity: BadgeEpic, Scope: BadgeMoment, Event: BadgeEvAppeased,
				Reveal: secret, Emblem: "harbinger", Proof: BotProof("harbinger"),
			},
			Why: "The harbinger bot has afforded Appease only for the figures of the first twelve ages; like the ten later Heeded badges, this one joins when a bot affords it in the Quantum Age.",
		},
		{
			Badge: BadgeDef{
				Key: "special.floor_it", Family: "special",
				Name: "Floor It", Desc: "Play at the highest speed your wonders allow.",
				Tier: BadgeBronze, Scope: BadgeMoment, Event: BadgeEvCensus,
				Emblem: "sun", Proof: Occurs(""),
			},
			Why: "Cannot happen: the game no longer has a speed the player sets.",
		},
	}
}
