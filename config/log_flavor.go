package config

import "math/rand"

// log_flavor.go — humour & personality layer (pass 9b), the LOG half.
//
// A small pool of flavour variants per "notable moment" the engine logs
// (a building finishing, an age turning over, a breakthrough, a famine, a
// catastrophe survived). The engine appends one of these as a dim trailing
// line so the same event reads a little differently each time, instead of
// the identical string forever.
//
// This is PURELY cosmetic — it never carries mechanical information. The
// functional log line (counts, names, effects) is always emitted separately
// and first; the flavour is additive. Picks are random per call, so a player
// queuing twenty buildings sees variety rather than the same quip twenty times.
//
// Keep these short (one line), tasteful, and aimed at the game world, never
// the player. They are intentionally age-agnostic: a single line has to land
// in the Stone Age and the Quantum Age alike, so they lean on the universal
// comedy of bureaucracy, pride, exhaustion, and mild existential unease.

// Log-flavour moment keys. Stable identifiers — the engine references these.
const (
	LogFlavorBuildingComplete    = "building_complete"
	LogFlavorAgeAdvance          = "age_advance"
	LogFlavorResearchDone        = "research_done"
	LogFlavorStarvation          = "starvation"
	LogFlavorStarvationEnded     = "starvation_ended"
	LogFlavorCatastropheSurvived = "catastrophe_survived"
)

// logFlavorPools maps a moment key to its pool of interchangeable quips.
var logFlavorPools = map[string][]string{
	LogFlavorBuildingComplete: {
		"The foreman declares it finished and immediately starts complaining about the next one.",
		"It is standing. It is, against the odds, standing.",
		"The ribbon-cutting was brief; nobody could find scissors.",
		"Built on time and only slightly over budget, which counts as a triumph.",
		"The workers admire their handiwork, then quietly wonder where the next paycheck is.",
		"A new building. The mapmakers sigh and reach for a fresh sheet.",
		"It matches the others. We have decided this is intentional.",
		"Construction complete. The scaffolding will linger for a suspiciously long time.",
	},
	LogFlavorAgeAdvance: {
		"The historians scramble to invent a flattering name for what just happened.",
		"An entire era, gone in a heartbeat. The old folks will not let anyone forget it.",
		"A new age dawns, and with it a whole new set of things to worry about.",
		"The calendar-makers are thrilled. Everyone else is mildly disoriented.",
		"Progress marches on, trampling the comfortable old ways as it goes.",
		"Future generations will study this moment, then misquote it.",
	},
	LogFlavorResearchDone: {
		"The researchers celebrate, then realise this opens up six new mysteries.",
		"Knowledge advances. Somewhere, a stubborn assumption quietly dies.",
		"The scholars are pleased, which for scholars means slightly less frowning.",
		"It works in theory and, remarkably, also in practice.",
		"A breakthrough. The grant committee will pretend they expected it all along.",
		"Discovery achieved. The notes are illegible but the result is sound.",
	},
	LogFlavorStarvation: {
		"The cooks are doing remarkable things with very little, and the very little is running out.",
		"Belts have been tightened. There are no more notches left to tighten.",
		"Morale surveys have been postponed on account of everyone being hungry.",
		"The granary echoes when you shout into it. People have been shouting into it.",
	},
	LogFlavorStarvationEnded: {
		"The cooks weep with relief. So, frankly, does everyone else.",
		"Full bellies return, and with them the luxury of complaining about other things.",
		"The famine is over. The collective sigh could be heard three valleys away.",
		"Food again. The crisis committee disbands, vowing never to speak of it.",
	},
	LogFlavorCatastropheSurvived: {
		"Battered, smoke-stained, and still here. Take the win.",
		"The civilization endures, mostly out of sheer stubbornness.",
		"Survivors emerge, dust themselves off, and start arguing about rebuilding plans.",
		"It was close. It is always close. Somehow it is always also fine.",
	},
}

// PickLogFlavor returns a random flavour line for the given moment, or ""
// if the moment is unknown or its pool is empty. Callers append the result
// as a separate, dim, cosmetic log line — never in place of functional text.
func PickLogFlavor(moment string) string {
	pool := logFlavorPools[moment]
	if len(pool) == 0 {
		return ""
	}
	return pool[rand.Intn(len(pool))]
}

// LogFlavorMoments returns the set of registered moment keys (sorted-order
// not guaranteed). Exposed for tests that assert every moment has a sane pool.
func LogFlavorMoments() []string {
	out := make([]string, 0, len(logFlavorPools))
	for k := range logFlavorPools {
		out = append(out, k)
	}
	return out
}
