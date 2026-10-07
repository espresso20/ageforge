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
// Nothing is reported for a run the developer console has changed, and a
// forced catastrophe never counts.
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
	// BadgeEvBadge: a badge was earned. Subject: its family. Integrity
	// badges are not reported.
	BadgeEvBadge = "badge_earned"

	// The integrity events. They are reported whatever state the run is in.
	//
	// BadgeEvDevUnlocked: the developer console was unlocked.
	BadgeEvDevUnlocked = "dev_unlocked"
	// BadgeEvSaveModified: a save edited outside the game was loaded.
	BadgeEvSaveModified = "save_modified"
	// BadgeEvSaveElite: a save carrying the forge master's proof was loaded.
	BadgeEvSaveElite = "save_elite"
)

// BadgeRevealKind says when a locked badge's name and description may be
// shown. Until then it is a silhouette: the game (not the screen drawing
// it) withholds the text.
type BadgeRevealKind uint8

const (
	// BadgeVisible shows from the start.
	BadgeVisible BadgeRevealKind = iota
	// BadgeRevealAtAge shows once the account has reached the age Key.
	BadgeRevealAtAge
	// BadgeRevealNextAge shows once the age Key is the next one, or
	// reached: the Next Age goal names it already.
	BadgeRevealNextAge
	// BadgeRevealOnCounter shows once the lifetime counter Key is above 0:
	// a civilization met, a harbinger heard, an awakening seen.
	BadgeRevealOnCounter
	// BadgeSecret never shows until earned. A secret badge lists as a
	// silhouette with its Hint.
	BadgeSecret
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

// RevealUntilSeen hides a badge until the lifetime counter is above 0.
func RevealUntilSeen(counter string) BadgeReveal {
	return BadgeReveal{Kind: BadgeRevealOnCounter, Key: counter}
}

// BadgeReward is what a badge gives. Cosmetic only: a theme or a title.
type BadgeReward struct {
	Theme string
	Title string
}

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
}

// StaticProof is a proof from config by rule.
func StaticProof(rule string) BadgeProof { return BadgeProof{Kind: BadgeProofStatic, Rule: rule} }

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
	// badge the event's subject must equal it ("" matches any).
	Subject string
	Name    string
	Desc    string
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
	// Emblem is the symbol the badge case draws; "" takes the family's.
	Emblem string
	Reward BadgeReward
	Proof  BadgeProof
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
}

// BadgeEraTiers is the tier of a badge about an age, by its era's order:
// the Stone and Iron Eras are bronze, the Steel and Electric silver, the
// Digital and Neon gold, the Cosmic platinum.
var BadgeEraTiers = []BadgeTier{
	BadgeBronze, BadgeBronze, BadgeSilver, BadgeSilver, BadgeGold, BadgeGold, BadgePlatinum,
}

// BadgeFamilies returns the generated families.
//
// This is the seed catalog: Only keeps a few subjects of each family, enough
// to exercise every path. The full catalog drops the Only lines and adds
// rows.
func BadgeFamilies() []BadgeFamilyDef {
	return []BadgeFamilyDef{
		{
			Family: "age", Source: BadgeSourceAges,
			Only: []string{"stone_age", "iron_age", "modern_age"},
			Key:  "age.{key}", Name: "{name}", Desc: "Reach the {name}.",
			Names: map[string]string{
				"age.stone_age":  "Rock Solid",
				"age.iron_age":   "Age of Iron",
				"age.modern_age": "Into the Modern Age",
			},
			TierByEra: true,
			Tiers:     map[string]BadgeTier{"transcendent_age": BadgeLegendary},
			Scope:     BadgeMoment, Event: BadgeEvAgeReached,
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleGate),
			Aliases: map[string][]string{
				"age.iron_age":   {"reached_iron"},
				"age.modern_age": {"reached_modern"},
			},
		},
		{
			Family: "lineage", Source: BadgeSourceLineages,
			Only: []string{"housing"},
			Key:  "lineage.{key}.{n}", Name: "{name} {rung}",
			Desc:  "Build {count} {lname} buildings across all your runs. Sold and rebuilt copies count once.",
			Scope: BadgeLifetime, Counter: BadgeEvBuiltLineage + ".{key}",
			Rungs: []BadgeRung{
				{Name: "Hobbyist", Tier: BadgeBronze},
				{Name: "Contractor", Tier: BadgeSilver},
				{Name: "Magnate", Tier: BadgeGold},
				{Name: "Tycoon", Tier: BadgePlatinum},
				{Name: "Dynasty", Tier: BadgeLegendary},
			},
			Ladders:         map[string][]float64{"housing": {14, 57}},
			RevealBySubject: true,
			Proof:           StaticProof(BadgeRuleLifetime),
		},
		{
			Family: "ladder", Source: BadgeSourceNone,
			Key: "ladder.prestiges.{n}", Name: "{rung}",
			Desc:  "Prestige {count} times.",
			Scope: BadgeLifetime, Counter: BadgeEvPrestige,
			Rungs: []BadgeRung{
				{Name: "First Prestige", Tier: BadgeBronze},
				{Name: "Creature of Habit", Tier: BadgeSilver},
				{Name: "Serial Reincarnator", Tier: BadgeGold},
				{Name: "Eternal Return", Tier: BadgeLegendary},
			},
			Descs:   map[string]string{"ladder.prestiges.1": "Prestige for the first time."},
			Ladders: map[string][]float64{"": {1, 3, 10, 25}},
			Proof:   StaticProof(BadgeRuleLifetime),
			Aliases: map[string][]string{
				"ladder.prestiges.1": {"first_prestige"},
				"ladder.prestiges.3": {"prestige_x10"},
			},
		},
	}
}

// Badges returns the hand-written badges: the specials and the integrity
// badges. The seed catalog holds one or two of each kind.
func Badges() []BadgeDef {
	return []BadgeDef{
		{
			Key: "special.hut_hoarder", Family: "special", Subject: "hut",
			Name: "Hut Hoarder",
			Desc: "Have 60 huts standing before you leave the Primitive Age. Each one costs more than the last.",
			Tier: BadgeGold, Rarity: BadgeEpic,
			Scope: BadgeRun, Event: BadgeEvBuildingBuilt, InAge: "primitive_age",
			Counter: "standing.hut", Threshold: 60,
			Proof: StaticProof(BadgeRuleCopies),
		},
		{
			Key: "special.fashionably_late", Family: "special",
			Name:  "Fashionably Late",
			Desc:  "Spend ten times the Primitive Age's pacing target in the Primitive Age.",
			Tier:  BadgeBronze,
			Scope: BadgeRun, Event: BadgeEvTick, InAge: "primitive_age",
			Pred: BadgePredAgeOverstay, Threshold: 10,
			Proof: StaticProof(BadgeRuleTime),
		},
		{
			Key: "special.liquidation_sale", Family: "special",
			Name:  "Liquidation Sale",
			Desc:  "Sell 100 buildings in one run.",
			Hint:  "Something about a clearance.",
			Tier:  BadgeBronze,
			Scope: BadgeRun, Event: BadgeEvBuildingSold,
			Counter: "run." + BadgeEvBuildingSold, Threshold: 100,
			Reveal: BadgeReveal{Kind: BadgeSecret},
			Proof:  StaticProof(BadgeRuleRunCount),
		},
		{
			Key: "special.hand_in_the_cookie_jar", Family: "special",
			Name:  "Hand in the Cookie Jar",
			Desc:  "Unlock the developer console. A run it changes earns nothing.",
			Scope: BadgeMoment, Event: BadgeEvDevUnlocked,
			Reveal: BadgeReveal{Kind: BadgeSecret},
			Proof:  BadgeProof{Kind: BadgeProofIntegrity},
		},
		{
			Key: "special.creative_accounting", Family: "special",
			Name:  "Creative Accounting",
			Desc:  "Load a save that was edited outside the game.",
			Scope: BadgeMoment, Event: BadgeEvSaveModified,
			Reveal: BadgeReveal{Kind: BadgeSecret},
			Proof:  BadgeProof{Kind: BadgeProofIntegrity},
		},
	}
}
