package config

// Badges are the account's permanent record: earned once, kept across every
// run, never worth anything inside a run. Milestones are the other layer
// (per run, with gameplay rewards); both are judged from the same reports
// of what happened in the game (game/badges.go), and a badge never changes
// a run's numbers, so a run plays the same on any account.
//
// This file is data. A badge is one BadgeDef row; a family of badges made
// from a config table (one per age, a ladder per lineage) is one
// BadgeFamilyDef row, expanded when the ruleset is compiled
// (rules/badges.go). Adding a badge is adding a row here: the engine
// already reports the events and keeps the counters a row can name.
//
// Every badge states how it can be earned (Proof). The reachability guard
// (smoke/static_badges.go) checks each proof against the game's own numbers,
// so a badge that cannot be earned fails the build.

// BadgeTier is a badge's metal: its art and its points.
type BadgeTier uint8

const (
	// BadgeNoTier is an integrity badge's tier: no metal, no points.
	BadgeNoTier BadgeTier = iota
	BadgeBronze
	BadgeSilver
	BadgeGold
	BadgePlatinum
	BadgeLegendary
)

// Name is the tier in words ("bronze"); "" for no tier.
func (t BadgeTier) Name() string {
	switch t {
	case BadgeBronze:
		return "bronze"
	case BadgeSilver:
		return "silver"
	case BadgeGold:
		return "gold"
	case BadgePlatinum:
		return "platinum"
	case BadgeLegendary:
		return "legendary"
	}
	return ""
}

// Points is what a badge of the tier adds to the account's score.
func (t BadgeTier) Points() int {
	switch t {
	case BadgeBronze:
		return 5
	case BadgeSilver:
		return 10
	case BadgeGold:
		return 25
	case BadgePlatinum:
		return 50
	case BadgeLegendary:
		return 100
	}
	return 0
}

// BadgeRarity is a design label for how much play a badge takes. It is not
// measured from players.
type BadgeRarity uint8

const (
	// BadgeRarityByTier takes the rarity from the tier: bronze is common,
	// silver uncommon, gold rare, platinum epic, legendary legendary.
	BadgeRarityByTier BadgeRarity = iota
	BadgeCommon
	BadgeUncommon
	BadgeRare
	BadgeEpic
	BadgeMythic
)

// Name is the rarity in words ("common"); "" for BadgeRarityByTier.
func (r BadgeRarity) Name() string {
	switch r {
	case BadgeCommon:
		return "common"
	case BadgeUncommon:
		return "uncommon"
	case BadgeRare:
		return "rare"
	case BadgeEpic:
		return "epic"
	case BadgeMythic:
		return "legendary"
	}
	return ""
}

// BadgeScope says what a badge is judged on.
type BadgeScope uint8

const (
	// BadgeLifetime is earned when an account counter reaches Threshold.
	// Counters only grow, across every run.
	BadgeLifetime BadgeScope = iota
	// BadgeRun is earned when Event is reported and the run's own facts
	// meet the row: Counter at Threshold or more, InAge, When and Pred.
	BadgeRun
	// BadgeMoment is earned when Event is reported and the event itself
	// meets the row: Subject, InAge, When and Pred.
	BadgeMoment
)

// The events the game reports (GameEngine.report). An event has a kind, the
// thing it happened to (its subject: an age, a building, an outcome) and an
// amount, 1 unless said otherwise. A lifetime counter is named after what it
// counts: the kind alone ("prestige") or the kind and a subject
// ("built.lineage.housing"). Only counters a badge names are kept.
//
// Nothing about the developer console changes any of this: a run a dev
// command has changed reports, counts and earns like any other.
//
// BadgeEvBuilt and BadgeEvBuiltLineage are reported only for the buildings
// a badge counts, so the run keeps build marks for those alone.
const (
	// BadgeEvTick is one game tick. Milestones are judged on it.
	BadgeEvTick = "tick"
	// BadgeEvAgeReached: the run entered an age. Subject: the age.
	BadgeEvAgeReached = "age_reached"
	// BadgeEvPrestige: a prestige completed. Subject: the age it was made from.
	BadgeEvPrestige = "prestige"
	// BadgeEvBuildingBuilt: one copy of a building finished. Subject: the
	// building. Every completion counts, a rebuilt copy included.
	BadgeEvBuildingBuilt = "building_built"
	// BadgeEvBuilt: a copy finished that took its building past the most
	// copies the run has built of it, net of sales (upgrades count for
	// neither). Subject: the building. Selling and rebuilding adds nothing.
	BadgeEvBuilt = "built"
	// BadgeEvBuiltLineage is BadgeEvBuilt by lineage. Subject: the lineage.
	BadgeEvBuiltLineage = "built.lineage"
	// BadgeEvWonderRaised: a wonder finished. Subject: the wonder.
	BadgeEvWonderRaised = "wonder_raised"
	// BadgeEvResearchDone: a tech finished. Subject: the tech.
	BadgeEvResearchDone = "research_done"
	// BadgeEvMilestone and BadgeEvChain: a milestone or a chain completed.
	// Subject: its key.
	BadgeEvMilestone = "milestone_completed"
	BadgeEvChain     = "chain_completed"
	// BadgeEvBuildingSold and BadgeEvBuildingUpgraded: copies sold or
	// upgraded. Subject: the building (the old tier). Amount: the copies.
	BadgeEvBuildingSold     = "building_sold"
	BadgeEvBuildingUpgraded = "building_upgraded"
	// BadgeEvGathered: one gather by hand. Subject: the resource.
	BadgeEvGathered = "gathered_by_hand"
	// BadgeEvStarved: a worker starved.
	BadgeEvStarved = "worker_starved"
	// BadgeEvEndured and BadgeEvSuccumbed: the answer to a catastrophe.
	// Subject: the era. Attributes on Endured: "brace" (the Brace level),
	// "saved" (buildings the garrison saved).
	BadgeEvEndured   = "catastrophe_endured"
	BadgeEvSuccumbed = "catastrophe_succumbed"
	// BadgeEvLastPassage: the Last Passage resolved. Subject: "endured",
	// "succumbed" or "spared".
	BadgeEvLastPassage = "last_passage"
	// BadgeEvHarbingerMet: a harbinger arrived. Subject: the figure.
	BadgeEvHarbingerMet = "harbinger_met"
	// BadgeEvHarbingerResolved: a thread got its verdict. Subject: the
	// verdict ("fulfilled", "vindicated", "spared", "discredited").
	BadgeEvHarbingerResolved = "harbinger_resolved"
	// BadgeEvAppeased, BadgeEvBraced, BadgeEvInvited: an answer bought from
	// a harbinger. Subject: the figure speaking.
	BadgeEvAppeased = "harbinger_appeased"
	BadgeEvBraced   = "harbinger_braced"
	BadgeEvInvited  = "harbinger_invited"
	// BadgeEvDeal: a faction deal taken. Subject: the civilization.
	BadgeEvDeal = "deal_taken"
	// BadgeEvExpedition: an expedition came back with a success. Subject:
	// the expedition.
	BadgeEvExpedition = "expedition_returned"
	// BadgeEvFestival and BadgeEvBlackMarket: a festival held, a black
	// market deal made (won or lost).
	BadgeEvFestival    = "festival_held"
	BadgeEvBlackMarket = "black_market_deal"
	// BadgeEvMemoryAccepted and BadgeEvMemoryDeclined: the answer to an
	// Ancient Memory.
	BadgeEvMemoryAccepted = "memory_accepted"
	BadgeEvMemoryDeclined = "memory_declined"
	// BadgeEvEraEvent: an era's entry event fired. Subject: its type
	// ("good_minor", "good_major", "good_legendary", "bad_challenging").
	BadgeEvEraEvent = "era_event"
	// BadgeEvAwakening: an awakening fired. Subject: the awakening.
	BadgeEvAwakening = "awakening"
	// BadgeEvMarketTrade: one exchange at the market. Subject: the resource
	// given.
	BadgeEvMarketTrade = "market_trade"
	// BadgeEvCivMet: a civilization was met for the first time this run.
	// Subject: the civilization.
	BadgeEvCivMet = "civ_met"
	// BadgeEvCivStatus: the player set a standing with a civilization.
	// Subject: the standing ("allied", "rival", "embargo", "neutral").
	// BadgeEvCivAllied is the same moment for an alliance, by civilization.
	BadgeEvCivStatus = "civ_status"
	BadgeEvCivAllied = "civ_allied"
	// BadgeEvRaidBlunted: the garrison kept part of a raid's losses.
	BadgeEvRaidBlunted = "raid_blunted"
	// BadgeEvLowMorale: morale fell below the point where the game warns.
	BadgeEvLowMorale = "morale_low"
	// BadgeEvUpgradeBought: a prestige shop item was bought. Subject: the
	// item.
	BadgeEvUpgradeBought = "upgrade_bought"
	// BadgeEvPlanStarted: the build plan started an item by itself.
	// Subject: the item's kind ("build", "research", "trade", "deal",
	// "advance").
	BadgeEvPlanStarted = "plan_started"
	// BadgeEvReturned: the game caught up on time away, on a load.
	// Attributes: "ticks" (the ticks credited), "capped" (1 when the time
	// away was past the offline allowance).
	BadgeEvReturned = "returned"
	// BadgeEvAwayTicks: the ticks credited for time away. Amount: the ticks,
	// so the run fact is the run's whole time away.
	BadgeEvAwayTicks = "away_ticks"
	// BadgeEvDayPlayed: a game was started or loaded on a calendar day the
	// account had not been played on before.
	BadgeEvDayPlayed = "day_played"
	// BadgeEvExported and BadgeEvRecoveryShown: the account was exported;
	// its recovery code was shown.
	BadgeEvExported      = "account_exported"
	BadgeEvRecoveryShown = "recovery_code_shown"
	// BadgeEvBadge: a badge was earned. Subject: its family. Attributes:
	// "top" (1 for the last rung of a ladder). A badge in a set also counts
	// "badge_earned.set.<set>". Integrity badges are not reported.
	BadgeEvBadge = "badge_earned"
	// BadgeEvProduced: what buildings, workers and techs produced of a
	// resource since the last report (a batch every few ticks, and one for
	// time away). Subject: the resource. Amount: how much. Trades, loot,
	// refunds and gifts are not production.
	BadgeEvProduced = "produced"
	// BadgeEvCensus: the state of the run, every BadgeCensusTicks ticks.
	// Attributes: "staffed.<domain>" (workers at work in a domain), "full"
	// (1 when every capped resource in use is at its cap), "age" (the
	// age's place in the order, from 0).
	BadgeEvCensus = "census"
	// BadgeEvWarStarted: a civilization went to war with the player.
	// Subject: the civilization. Attribute: "wars" (wars under way now).
	// BadgeEvWarEnded: a war ended. Attribute: "tribute" (1 when it was
	// bought off, 0 when it burned out).
	BadgeEvWarStarted = "war_started"
	BadgeEvWarEnded   = "war_ended"
	// BadgeEvGiftSent: a gift was sent. Subject: the civilization.
	BadgeEvGiftSent = "gift_sent"
	// BadgeEvWorkersLent: a civilization lent workers. Subject: the
	// civilization.
	BadgeEvWorkersLent = "workers_lent"
	// BadgeEvDoomNamed: a harbinger's warning named an era's doom, or the
	// doom struck. Subject: the era.
	BadgeEvDoomNamed = "doom_named"
	// BadgeEvEraLeft: the run left an era (entered the next, or made a
	// prestige from the last age). Subject: the era left. Attribute:
	// "pace" (the ticks spent in the era over its ages' pacing targets).
	BadgeEvEraLeft = "era_left"
	// BadgeEvPlanQueued: something was added to the build plan. Attribute:
	// "size" (items waiting now).
	BadgeEvPlanQueued = "plan_queued"
	// BadgeEvWonderOverflow: overflow fed a wonder's bank for the first
	// time in the run.
	BadgeEvWonderOverflow = "wonder_overflow"
	// BadgeEvWonderBanked: a wonder's bank was filled by overflow.
	// Attribute: "away" (1 when it filled during time away).
	BadgeEvWonderBanked = "wonder_banked"
	// BadgeEvRuinsCarried: a run began with ruins of a civilization that
	// fell.
	BadgeEvRuinsCarried = "ruins_carried"
	// BadgeEvLegacyPrestige: a prestige completed while Cosmic Legacy was
	// held.
	BadgeEvLegacyPrestige = "prestige_with_legacy"
	// BadgeEvKit: a legacy kit item was bought. Attribute: "left" (kit
	// items not owned yet).
	BadgeEvKit = "kit_bought"

	// The events about how the account looks. They are read off the
	// account's own settings when the run advances an age, so they are
	// never tallied in a run (BadgeSessionEvents).
	//
	// BadgeEvAdvancedTheme, BadgeEvAdvancedStyle, BadgeEvAdvancedGlyphs: the
	// theme, map style and glyph set worn at an advance. Subject: its key.
	BadgeEvAdvancedTheme  = "advanced_in_theme"
	BadgeEvAdvancedStyle  = "advanced_with_style"
	BadgeEvAdvancedGlyphs = "advanced_with_glyphs"
	// BadgeEvVisitor: a visitor on the map was inspected. Subject: its
	// kind.
	BadgeEvVisitor = "visitor_inspected"

	// The integrity events. Like the day played they are about the session,
	// not the run (BadgeSessionEvents).
	//
	// BadgeEvDevUnlocked: the developer console was unlocked.
	BadgeEvDevUnlocked = "dev_unlocked"
	// BadgeEvSaveModified: a save edited outside the game was loaded.
	BadgeEvSaveModified = "save_modified"
	// BadgeEvSaveElite: a save carrying the forge master's proof was loaded.
	BadgeEvSaveElite = "save_elite"
)

// BadgeEventKinds lists every event the game reports. A badge may be judged
// on, and a counter may count, only these.
func BadgeEventKinds() []string {
	return []string{
		BadgeEvTick, BadgeEvAgeReached, BadgeEvPrestige,
		BadgeEvBuildingBuilt, BadgeEvBuilt, BadgeEvBuiltLineage, BadgeEvWonderRaised,
		BadgeEvResearchDone, BadgeEvMilestone, BadgeEvChain,
		BadgeEvBuildingSold, BadgeEvBuildingUpgraded, BadgeEvGathered, BadgeEvStarved,
		BadgeEvEndured, BadgeEvSuccumbed, BadgeEvLastPassage,
		BadgeEvHarbingerMet, BadgeEvHarbingerResolved, BadgeEvAppeased, BadgeEvBraced, BadgeEvInvited,
		BadgeEvDeal, BadgeEvExpedition, BadgeEvFestival, BadgeEvBlackMarket,
		BadgeEvMemoryAccepted, BadgeEvMemoryDeclined, BadgeEvEraEvent, BadgeEvAwakening,
		BadgeEvMarketTrade, BadgeEvCivMet, BadgeEvCivStatus, BadgeEvCivAllied,
		BadgeEvRaidBlunted, BadgeEvLowMorale, BadgeEvUpgradeBought, BadgeEvPlanStarted,
		BadgeEvReturned, BadgeEvAwayTicks,
		BadgeEvDayPlayed, BadgeEvExported, BadgeEvRecoveryShown, BadgeEvBadge,
		BadgeEvDevUnlocked, BadgeEvSaveModified, BadgeEvSaveElite,
		BadgeEvProduced, BadgeEvCensus, BadgeEvWarStarted, BadgeEvWarEnded, BadgeEvGiftSent,
		BadgeEvWorkersLent, BadgeEvDoomNamed, BadgeEvEraLeft, BadgeEvPlanQueued,
		BadgeEvWonderOverflow, BadgeEvWonderBanked, BadgeEvRuinsCarried, BadgeEvLegacyPrestige, BadgeEvKit,
		BadgeEvAdvancedTheme, BadgeEvAdvancedStyle, BadgeEvAdvancedGlyphs, BadgeEvVisitor,
	}
}

// BadgeCensusTicks is how often the run's state is reported (BadgeEvCensus).
const BadgeCensusTicks = 50

// BadgeProducedTicks is how often production is reported (BadgeEvProduced).
const BadgeProducedTicks = 10

// BadgeSessionEvents are the events that are about the session or the
// account rather than the run: the tick itself, a new calendar day, an
// export, the recovery code shown, a badge earned, the developer console
// unlocked, what loading a save found. They
// are judged against the account and are never tallied in a run's facts, so
// what a run says happened in it is the same on any account and any date. A
// badge cannot ask for one as a run fact ("run.day_played").
func BadgeSessionEvents() []string {
	return []string{
		BadgeEvTick, BadgeEvDayPlayed, BadgeEvExported, BadgeEvRecoveryShown, BadgeEvBadge,
		BadgeEvDevUnlocked, BadgeEvSaveModified, BadgeEvSaveElite,
		// Counted for the account, in batches or by the clock: a tally of
		// them in the run would only be bulk.
		BadgeEvProduced, BadgeEvCensus,
		// How the account looks, not what the run did.
		BadgeEvAdvancedTheme, BadgeEvAdvancedStyle, BadgeEvAdvancedGlyphs, BadgeEvVisitor,
	}
}

// BadgeRevealKind says when a locked badge's name and description may be
// shown. Until then it is a silhouette: the game (not the screen drawing
// it) withholds the text.
type BadgeRevealKind uint8

const (
	// BadgeVisible shows from the start.
	BadgeVisible BadgeRevealKind = iota
	// BadgeRevealAtAge shows once the account has reached the age Key.
	BadgeRevealAtAge
	// BadgeRevealNextAge shows once the age Key has been the account's next
	// age, or reached: the Next Age goal named it then.
	BadgeRevealNextAge
	// BadgeRevealOnCounter shows once the lifetime counter Key is above 0:
	// a civilization met, a harbinger heard, an awakening seen.
	BadgeRevealOnCounter
	// BadgeSecret never shows until earned. A secret badge lists as a
	// silhouette with its Hint.
	BadgeSecret
	// BadgeRevealOnBadge shows once the account holds the badge Key: a
	// theme's badge shows when the theme is unlocked.
	BadgeRevealOnBadge
)

// BadgeReveal is a badge's spoiler rule.
type BadgeReveal struct {
	Kind BadgeRevealKind
	Key  string
}

// RevealUntilAge hides a badge until the account has reached age.
func RevealUntilAge(age string) BadgeReveal { return BadgeReveal{Kind: BadgeRevealAtAge, Key: age} }

// RevealUntilNextAge hides a badge until age is next, or reached.
func RevealUntilNextAge(age string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealNextAge, Key: age}
}

// RevealUntilCivMet hides a badge until the civilization has been met.
func RevealUntilCivMet(civ string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealOnCounter, Key: BadgeEvCivMet + "." + civ}
}

// RevealUntilHarbingerMet hides a badge until the figure has arrived once.
func RevealUntilHarbingerMet(figure string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealOnCounter, Key: BadgeEvHarbingerMet + "." + figure}
}

// RevealUntilDoomNamed hides a badge until a warning has named the era's
// doom, or it has struck.
func RevealUntilDoomNamed(era string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealOnCounter, Key: BadgeEvDoomNamed + "." + era}
}

// RevealUntilBadge hides a badge until the account holds the badge key.
func RevealUntilBadge(key string) BadgeReveal { return BadgeReveal{Kind: BadgeRevealOnBadge, Key: key} }

// RevealUntilSeen hides a badge until the lifetime counter is above 0.
func RevealUntilSeen(counter string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealOnCounter, Key: counter}
}

// BadgeReward is what a badge gives. Cosmetic only: a theme or a title.
type BadgeReward struct {
	// Theme is a theme key. Earning the badge unlocks the theme on the
	// account.
	Theme string
	// Title is a title the badge case may show.
	Title string
}

// BadgeScoreTitle is a title the account's badge score earns.
type BadgeScoreTitle struct {
	// Points is the least score that holds the title.
	Points int
	Title  string
}

// BadgeScoreTitles returns the score titles, lowest first. The first one
// asks for nothing, so every account holds a title.
func BadgeScoreTitles() []BadgeScoreTitle {
	return []BadgeScoreTitle{
		{0, "Settler"},
		{250, "Headman"},
		{1000, "Magistrate"},
		{3000, "Sovereign"},
		{6000, "Paragon"},
		{10000, "Eternal"},
	}
}

// BadgeCompleteTitle is the title of an account that holds every badge
// that counts toward completion.
const BadgeCompleteTitle = "Completionist"

// BadgeProofKind is how a badge is known to be earnable.
type BadgeProofKind uint8

const (
	// BadgeNoProof is the zero value. A badge without a proof fails the
	// guard.
	BadgeNoProof BadgeProofKind = iota
	// BadgeProofStatic is proven from config alone, by the rule in Rule.
	BadgeProofStatic
	// BadgeProofBot is proven by play: a smoke bot of the style in Rule
	// must earn it.
	BadgeProofBot
	// BadgeProofDerived follows from other badges (earn 25 badges, finish
	// a ladder): it is reachable when they are.
	BadgeProofDerived
	// BadgeProofIntegrity marks an integrity badge. It is exempt from the
	// guard, worth no points and left out of completion.
	BadgeProofIntegrity
)

// The static proof rules (BadgeProof.Rule with BadgeProofStatic).
const (
	// BadgeRuleGate: the subject is an age, and every age gate is proven
	// reachable by the Gate Covenant.
	BadgeRuleGate = "gate"
	// BadgeRuleCopies: Threshold copies of the building Subject stand in
	// its own age. The last copy must fit the most storage buildable
	// there, and the count may be at most BadgeCopiesShare of the ceiling.
	BadgeRuleCopies = "copies"
	// BadgeRuleLifetime: a lifetime count. One run must add to the counter,
	// and Threshold may take at most BadgeMaxRuns runs.
	BadgeRuleLifetime = "lifetime"
	// BadgeRuleTime: a time threshold, stated as a multiple of an age's
	// pacing target (never in hours).
	BadgeRuleTime = "time"
	// BadgeRuleRunCount: Threshold of something a run can do at least that
	// often. The guard knows the limit for each event it can prove.
	BadgeRuleRunCount = "run_count"
	// BadgeRuleCeiling: every copy of the building Subject that can stand
	// in its own age, or fewer: the last one the storage buildable there
	// (or its max count) allows. Unlike BadgeRuleCopies it may ask for the
	// very last copy.
	BadgeRuleCeiling = "ceiling"
	// BadgeRuleAgeCopies: Threshold buildings of the age InAge standing
	// together, at most BadgeAgeShare of what the age's storage allows of
	// all of them.
	BadgeRuleAgeCopies = "age_copies"
	// BadgeRuleWonder: the subject is the wonder of an age, which the
	// advance from that age asks for (or, in the last age, one its storage
	// can pay for).
	BadgeRuleWonder = "wonder"
	// BadgeRuleTechs: every tech of the age Subject, counted from the tech
	// tree: Threshold is that count, and no tech shuts another out.
	BadgeRuleTechs = "techs"
	// BadgeRuleStaffing: Threshold workers at work in the domain Subject
	// at once, at most BadgeStaffShare of the slots a run's buildings of
	// that domain can hold, and no more than the housing a run can build.
	BadgeRuleStaffing = "staffing"
	// BadgeRuleHabit: a count of calendar days or of visits. It measures a
	// habit, not pacing, so it is not held to a number of runs.
	BadgeRuleHabit = "habit"
	// BadgeRuleOccurs: the thing asked for is something the game does, and
	// the row's Why says, in words, why a player can bring it about. The
	// guard checks what it can from the tables: that the subject exists
	// (a civilization, a harbinger, an era a doom can strike in, an
	// expedition, an awakening, a theme) and that the event is reported.
	BadgeRuleOccurs = "occurs"
)

const (
	// BadgeAgeShare is the most of an age's whole building ceiling an age
	// maximalist badge may ask for.
	BadgeAgeShare = 0.75
	// BadgeStaffShare is the most of a domain's worker slots a staffing
	// badge may ask for.
	BadgeStaffShare = 0.25
)

const (
	// BadgeCopiesShare is the most of a building's own-age ceiling a
	// copies badge may ask for.
	BadgeCopiesShare = 0.97
	// BadgeRunShare is how much of a lineage's ceiling an ordinary run
	// builds: the unit lifetime build ladders are measured in.
	BadgeRunShare = 0.10
	// BadgeMaxRuns is the most runs of ordinary play the top rung of a
	// ladder may take.
	BadgeMaxRuns = 25
)

// BadgeProof is a badge's claim that it can be earned.
type BadgeProof struct {
	Kind BadgeProofKind
	// Rule is the static rule, or the bot style.
	Rule string
	// Why is the argument in words, for a proof the guard cannot compute
	// (BadgeRuleOccurs asks for one).
	Why string
}

// StaticProof is a proof from config by rule.
func StaticProof(rule string) BadgeProof { return BadgeProof{Kind: BadgeProofStatic, Rule: rule} }

// Occurs is the proof that something happens in the game, with the reason
// a player can bring it about.
func Occurs(why string) BadgeProof {
	return BadgeProof{Kind: BadgeProofStatic, Rule: BadgeRuleOccurs, Why: why}
}

// BotProof is a proof by play: a smoke bot of the style earns it.
func BotProof(style string) BadgeProof { return BadgeProof{Kind: BadgeProofBot, Rule: style} }

// DerivedProof is the proof of a badge that follows from other badges.
func DerivedProof() BadgeProof { return BadgeProof{Kind: BadgeProofDerived} }

// BadgeMeasure says how a threshold is filled in when the ruleset is
// compiled, for a number that belongs to the game's tables and not to the
// badge.
type BadgeMeasure uint8

const (
	// BadgeMeasureNone: the threshold is as written.
	BadgeMeasureNone BadgeMeasure = iota
	// BadgeMeasureProduction: a family's rungs are counts of the subject
	// produced. The first rung is what an ordinary town makes of it in the
	// first age it makes any (its typical income times that age's pacing
	// target), so a first run earns it there or soon after. The top rung
	// is TopRuns times what one run produces: the sum, over the run's
	// ages, of the typical income times the age's pacing target. The rungs
	// between climb in even multiplicative steps. A run is the ages before
	// the one a full prestige is made from; a resource that comes later is
	// measured over its first two ages.
	BadgeMeasureProduction
	// BadgeMeasureTechs: the threshold is the number of techs the subject
	// age has.
	BadgeMeasureTechs
	// BadgeMeasureMilestones: the threshold is the number of milestones.
	BadgeMeasureMilestones
	// BadgeMeasureSet: the threshold is the number of badges in the set
	// the counter names ("badge_earned.set.<set>").
	BadgeMeasureSet
)

// BadgeOp compares a fact with a value.
type BadgeOp uint8

const (
	// BadgeAtLeast holds when the fact is Value or more.
	BadgeAtLeast BadgeOp = iota
	// BadgeAtMost holds when the fact is Value or less. A fact the run
	// did not track from its first tick fails it: "never" cannot be shown.
	BadgeAtMost
)

// BadgeCond is one more condition on a badge. Fact names what is read:
//
//	run.<event>             how often the event was reported this run
//	run.<event>.<subject>   the same, for one subject
//	standing.<building>     copies of the building standing now
//	ev.<attribute>          an attribute of the event being judged
//	life.<counter>          a lifetime counter of the account
type BadgeCond struct {
	Fact  string
	Op    BadgeOp
	Value float64
}

// The named predicates (BadgeDef.Pred), for conditions a BadgeCond cannot
// say. Each is a function in game/badge_preds.go.
const (
	// BadgePredAgeOverstay: the run has spent Threshold times the pacing
	// target of the age it is in, in that age.
	BadgePredAgeOverstay = "age_overstay"
	// BadgePredMostlyAway: the run's credited time away is Threshold times
	// its time played, or more.
	BadgePredMostlyAway = "mostly_away"
)

// BadgeDef is one badge.
type BadgeDef struct {
	// Key is permanent: it is what the account file stores.
	// "age.iron_age", "lineage.housing.3", "special.hut_hoarder".
	Key string
	// Family groups badges on one tab: "age", "lineage", "ladder",
	// "special".
	Family string
	// Subject is the config key the badge is about. For a Run or Moment
	// badge the event's subject must equal it ("" matches any), unless
	// AnySubject is set.
	Subject string
	// AnySubject: the badge is about Subject, and the event it is judged on
	// is about something else (an age's buildings are counted as each
	// building of it finishes; the census has no subject at all). The
	// event's subject is not held to Subject; the row's counter and
	// conditions name what is read.
	AnySubject bool
	Name       string
	Desc       string
	// Hint is the one line a secret badge shows until it is earned.
	Hint   string
	Tier   BadgeTier
	Rarity BadgeRarity
	Scope  BadgeScope
	// Counter and Threshold: for a Lifetime badge, the account counter and
	// the count that earns it. For a Run badge, a fact (as in BadgeCond)
	// and its least value; "" asks for nothing.
	Counter   string
	Threshold float64
	// Event is the report a Run or Moment badge is judged on.
	Event string
	// InAge, when set, is the age the run must be in.
	InAge string
	When  []BadgeCond
	Pred  string
	// Reveal is the spoiler rule. Text shown before the reveal may name no
	// age later than the one it reveals at; the guard checks it.
	Reveal BadgeReveal
	// Emblem names the symbol the badge case draws: a name from the
	// case's emblem table ("hut", "star"), "lineage.<key>" for a
	// lineage's map symbol, or "centre.<n>" for the town centre of era n.
	// "" takes the family's, and a star without one.
	Emblem string
	// Ladder is the name of the ladder the badge is a rung of ("Housing"),
	// "" for a badge on none. The rungs of a ladder share a Counter and
	// are told apart by their Threshold.
	Ladder string
	// Set names a set the badge belongs to ("civ.met"). Earning it adds
	// to the counter "badge_earned.set.<Set>", which a badge for the whole
	// set counts.
	Set string
	// Measure fills Threshold from the game's tables when the ruleset is
	// compiled.
	Measure BadgeMeasure
	Reward  BadgeReward
	Proof   BadgeProof
	// Aliases are the keys this badge had in older account files (the four
	// account achievements). An account holding one holds the badge.
	Aliases []string
}

// Points is what the badge adds to the score: its tier's, and nothing for
// an integrity badge.
func (d BadgeDef) Points() int {
	if d.Integrity() {
		return 0
	}
	return d.Tier.Points()
}

// Integrity reports whether d is an integrity badge.
func (d BadgeDef) Integrity() bool { return d.Proof.Kind == BadgeProofIntegrity }

// RarityName is the badge's rarity in words, from its tier when it sets
// none. "" for an integrity badge without one.
func (d BadgeDef) RarityName() string {
	if d.Rarity != BadgeRarityByTier {
		return d.Rarity.Name()
	}
	switch d.Tier {
	case BadgeBronze:
		return BadgeCommon.Name()
	case BadgeSilver:
		return BadgeUncommon.Name()
	case BadgeGold:
		return BadgeRare.Name()
	case BadgePlatinum:
		return BadgeEpic.Name()
	case BadgeLegendary:
		return BadgeMythic.Name()
	}
	return ""
}

// BadgeSource is the config table a family takes its subjects from.
type BadgeSource uint8

const (
	// BadgeSourceNone is a family with one unnamed subject: a ladder over
	// a counter that belongs to no table ("ladder.prestiges").
	BadgeSourceNone BadgeSource = iota
	// BadgeSourceAges: one subject per age. {name} is "Iron Age".
	BadgeSourceAges
	// BadgeSourceEras: one per era.
	BadgeSourceEras
	// BadgeSourceWonders: one per age's wonder, revealed with its age.
	BadgeSourceWonders
	// BadgeSourceLineages: one per building lineage (storage included,
	// wonders and monuments not), revealed with its first building's age.
	BadgeSourceLineages
	// BadgeSourceResources: one per resource, revealed with its age.
	BadgeSourceResources
	// BadgeSourceCivs: one per civilization, revealed when met.
	BadgeSourceCivs
	// BadgeSourceHarbingers: one per harbinger figure, revealed when met.
	BadgeSourceHarbingers
	// BadgeSourceAwakenings: one per awakening, revealed when it fires.
	BadgeSourceAwakenings
	// BadgeSourceDooms: one per era a doom can strike in. {name} is the
	// catastrophe ("The Great Plague"). Revealed when a warning names it.
	BadgeSourceDooms
	// BadgeSourceDomains: one per worker domain a building can be staffed
	// in, revealed with its first building's age.
	BadgeSourceDomains
	// BadgeSourceExpeditions: one per expedition (BadgeExpeditions),
	// revealed with its age.
	BadgeSourceExpeditions
	// BadgeSourceThemes: one per theme (BadgeThemes). A theme a badge
	// gives shows when that badge is held.
	BadgeSourceThemes
)

// BadgeRung is one step of a ladder.
type BadgeRung struct {
	// Name is the rung's word in a badge's name ("Hobbyist").
	Name string
	Tier BadgeTier
}

// BadgeFamilyDef makes one badge per subject of a config table, or one per
// subject and rung for a ladder. The text fields are templates:
//
//	{key}   the subject's key           {name}  its display name
//	{lname} the name in lower case      {rung}  the rung's name
//	{n}     the rung's number, from 1   {count} the threshold, written out
//	{era}   the order of the subject's era, from 0
//	{Name}  the name with its first letter in upper case
//	{techs} "all 6", "both" or "the one", for the subject age's techs
type BadgeFamilyDef struct {
	// Family is the badges' Family.
	Family string
	Source BadgeSource
	// Only keeps the listed subjects; empty keeps every subject the source
	// has. Except drops subjects.
	Only   []string
	Except []string
	// Key, Name and Desc are the templates for each badge. Names and Descs
	// override Name and Desc for a badge, by its key.
	Key   string
	Name  string
	Desc  string
	Names map[string]string
	Descs map[string]string
	// Tier is every badge's tier. TierByEra, when set, takes it from the
	// subject's era instead (BadgeEraTiers); Tiers overrides either for a
	// subject. A ladder takes its tiers from Rungs.
	Tier      BadgeTier
	TierByEra bool
	Tiers     map[string]BadgeTier
	Rarity    BadgeRarity
	Scope     BadgeScope
	// Event, Counter and InAge are templates too ("built.lineage.{key}").
	Event   string
	Counter string
	InAge   string
	// Threshold is every badge's threshold. A ladder takes its own from
	// Ladders: subject (or "" for BadgeSourceNone) to one count per rung,
	// which may be fewer than Rungs.
	Threshold float64
	Rungs     []BadgeRung
	Ladders   map[string][]float64
	// Reveal is the rule for every badge. RevealBySubject, when set, takes
	// it from the subject instead: its age for an age (the next age shows),
	// a wonder, a lineage or a resource; when met for a civilization or a
	// harbinger; when it fires for an awakening.
	Reveal          BadgeReveal
	RevealBySubject bool
	Proof           BadgeProof
	// Aliases and Rewards are by badge key.
	Aliases map[string][]string
	Rewards map[string]BadgeReward
	// Emblem is every badge's emblem and Ladder the name of each subject's
	// ladder, both templates ("lineage.{key}", "{name}"). Emblems
	// overrides Emblem for a subject.
	Emblem  string
	Emblems map[string]string
	Ladder  string
	// Set is the set every badge of the family joins.
	Set string
	// Measure and TopRuns fill the thresholds from the game's tables:
	// TopRuns is how many runs the top rung takes (BadgeMeasureProduction).
	Measure BadgeMeasure
	TopRuns float64
	// Hint is every badge's hint, for a secret family.
	Hint string
	// When is every badge's extra conditions.
	When []BadgeCond
	// Thresholds is each subject's threshold, for a family that is not a
	// ladder and whose number belongs to the subject (an age's buildings).
	Thresholds map[string]float64
	// RevealAtAge reveals each badge when the subject's own age is
	// reached, whatever its table's usual rule (an age's badge otherwise
	// shows one age early).
	RevealAtAge bool
	// AnySubject is BadgeDef.AnySubject for every badge of the family.
	AnySubject bool
}

// BadgeEraTiers is the tier of a badge about an age, by its era's order:
// the Stone and Iron Eras are bronze, the Steel and Electric silver, the
// Digital and Neon gold, the Cosmic platinum.
var BadgeEraTiers = []BadgeTier{
	BadgeBronze, BadgeBronze, BadgeSilver, BadgeSilver, BadgeGold, BadgeGold, BadgePlatinum,
}
