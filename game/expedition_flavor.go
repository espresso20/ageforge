package game

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/flavor"
)

// Game → flavor glue: the adapters that turn engine state into a flavor.Request.
//
// This file is the exact analogue of faction_boon.go's factionProfile/boonApplier
// for the prose engine. Package flavor knows nothing about expeditions, factions,
// or raids — it takes a Moment and a handful of optional hints — so every piece of
// game knowledge (which category maps to which Kind, which resource is worth
// naming, what tone a personality earns) lives HERE.
//
// All three entry points draw from ge.rng, so the prose stream is reproducible
// from the run's persisted seed alongside the encounter and boon streams. They run
// under the engine write lock and touch only held state plus pure config lookups —
// no GetState, no lock-acquiring calls.

// flavorStream returns the engine's prose stream, building it on first use.
//
// Every generated line in the game goes through this one Stream, which suppresses
// a sentence it used in the last couple of dozen lines. Package flavor's Generate
// is memoryless by contract — that is what makes a seeded run reproducible — and
// a memoryless uniform draw repeats inside a screenful no matter how large the
// catalog is, so the de-duplication has to live with whoever owns the log. That
// is here.
func (ge *GameEngine) flavorStream() *flavor.Stream {
	if ge.prose == nil {
		ge.prose = flavor.NewStream()
	}
	return ge.prose
}

// expeditionFlavorOdds is the denominator of the "does this resolution get a
// quip" roll: roughly one expedition in this many gets a flavour line.
//
// Present but not stale, the same policy processBuildQueue applies to building
// completions — a line under every single resolution stops being a detail and
// becomes wallpaper, and the eye starts skipping the whole gray column. The
// encounter and raid lines are NOT gated: those moments are already rare (a few
// dozen encounters per ten thousand ticks) and a silent war raid reads as a
// missing log entry rather than as restraint.
const expeditionFlavorOdds = 3

// expeditionFlavorLine returns the cosmetic line that rides ALONGSIDE a resolved
// expedition's mechanical message ("<Name> succeeded. Loot: 60 food."). It never
// carries mechanical information — the functional line is emitted first and
// separately, exactly as the log-flavour layer does elsewhere.
//
// Returns "" when the odds roll declines this resolution, or if the generator has
// nothing for the moment (the latter guard is so a future Moment removal degrades
// to silence, not to a blank log entry).
func (ge *GameEngine) expeditionFlavorLine(res ExpeditionResult) string {
	if ge.rng == nil {
		ge.SeedRNG(newSeed())
	}
	// Drawn off ge.rng, not package rand, so the prose stream stays reproducible
	// from the run's persisted seed alongside the encounter and boon streams.
	if ge.rng.Intn(expeditionFlavorOdds) != 0 {
		return ""
	}
	moment := flavor.ExpeditionSuccess
	if !res.Success {
		moment = flavor.ExpeditionFailure
	}
	resource, amount := topReward(res.Rewards)
	return ge.flavorStream().Line(flavor.Request{
		Moment: moment,
		Tone:   expeditionTone(res.Category, res.Success),
		Age:    ge.age,
		// Subject is the expedition's display NAME, which is a verb-led order
		// title as often as a noun phrase ("Raid Bandit Camp" vs "Scout Party").
		// package flavor only ever uses it inside title frames, which is why it
		// is safe to hand over raw.
		Subject:  res.Name,
		Kind:     res.Category, // ExpeditionScouting / ExpeditionMilitary
		Resource: resource,
		Amount:   amount,
	}, ge.rng)
}

// expeditionTone maps an outcome to a register. A won campaign is worth
// celebrating; a scouting party that got away with it is worth a raised eyebrow;
// a campaign that went wrong cost people.
func expeditionTone(category string, success bool) flavor.Tone {
	switch {
	case success && category == ExpeditionMilitary:
		return flavor.Triumphant
	case success:
		return flavor.Wry
	case category == ExpeditionMilitary:
		return flavor.Grim
	default:
		return flavor.Neutral
	}
}

// topReward returns the single largest reward line — the one worth naming in a
// sentence — or ("", 0) when there is nothing to name.
//
// It iterates config.BaseResources() rather than the rewards map so the choice is
// DETERMINISTIC: Go map iteration order is randomised, and picking the resource
// off a map walk would desynchronise the prose stream from the seed. Ties break on
// config order.
func topReward(rewards map[string]float64) (string, float64) {
	if len(rewards) == 0 {
		return "", 0
	}
	best, bestAmt := "", 0.0
	for _, def := range config.BaseResources() {
		if amt, ok := rewards[def.Key]; ok && amt > bestAmt {
			best, bestAmt = def.Key, amt
		}
	}
	return best, bestAmt
}

// factionEncounterFlavor generates the line for an encounter outcome that gave the
// player nothing — a standoff with a civ at war, or an offer the court has no room
// for. Both were fixed three-line banks before this; they are now drawn off the
// same generator, and game/boon_tuning_test.go classifies them via
// flavor.Signatures rather than by matching those banks.
func (ge *GameEngine) factionEncounterFlavor(moment flavor.Moment, def config.FactionDef) string {
	if ge.rng == nil {
		ge.SeedRNG(newSeed())
	}
	return ge.flavorStream().Line(flavor.Request{
		Moment: moment,
		Tone:   factionTone(moment, def.Personality),
		Age:    ge.age,
		// Subject is a civilization name whose grammatical NUMBER is inconsistent
		// across the roster ("Merchant Guild" vs "Void Reavers"); package flavor
		// only uses it in object position or with past-tense verbs.
		Subject: def.Name,
		Kind:    def.Personality,
	}, ge.rng)
}

// raidLogLine composes the full player-facing war-raid entry: the mechanical line
// (who raided, what was lost, how much) followed by a generated sentence. The two
// halves stay separable on purpose — the mechanics never depend on the prose.
func (ge *GameEngine) raidLogLine(raid RaidRequest) string {
	if q := ge.raidFlavorLine(raid); q != "" {
		return raid.Message + " " + q
	}
	return raid.Message
}

// raidFlavorLine generates the prose half of a war-raid log entry. The mechanical
// half (who raided, what was lost, how much) is composed by raidMessage and is
// NOT repeated here — which is also why no Amount is passed: the number is already
// on the line.
func (ge *GameEngine) raidFlavorLine(raid RaidRequest) string {
	if ge.rng == nil {
		ge.SeedRNG(newSeed())
	}
	def, ok := config.FactionByKey()[raid.FactionKey]
	if !ok {
		return ""
	}
	return ge.flavorStream().Line(flavor.Request{
		Moment:   flavor.WarRaid,
		Tone:     factionTone(flavor.WarRaid, def.Personality),
		Age:      ge.age,
		Subject:  def.Name,
		Kind:     def.Personality,
		Resource: raid.Resource,
	}, ge.rng)
}

// factionTone maps a civilization's personality to a register per moment. An
// aggressive neighbour's raid reads grim; a mercantile one's polite non-war reads
// wry; a court drowning in gifts is wry by definition.
func factionTone(moment flavor.Moment, personality string) flavor.Tone {
	switch moment {
	case flavor.WarRaid:
		if personality == "aggressive" {
			return flavor.Grim
		}
		return flavor.Neutral
	case flavor.EncounterStandoff:
		if personality == "peaceful" || personality == "mercantile" {
			return flavor.Wry
		}
		return flavor.Neutral
	case flavor.EncounterAtCapacity:
		return flavor.Wry
	default:
		return flavor.Neutral
	}
}
