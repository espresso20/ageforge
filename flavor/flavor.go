// Package flavor is a standalone, reusable procedural prose service. It turns a
// structured description of a game moment into ONE varied, in-world, player-facing
// sentence, so the surfaces that fire hundreds of times per run stop reading like
// the same string on a loop.
//
// It is a HYBRID generator, not a word-salad one. Sentence STRUCTURE is generated
// (a template is an ordered list of slots); the words that fill the slots are
// AUTHORED phrases, written to read like prose and composed so the grammar holds
// however they combine. There is no {adj} {noun} {verb} anywhere in here: a slot
// draws a written clause, and the template supplies the connective tissue and the
// punctuation.
//
// Decoupling is the whole point, and it mirrors package boon exactly. This package
// imports AT MOST config (for the resource vocabulary and the canonical age order)
// and the standard library. It does NOT import game — that would be an import
// cycle and would defeat the isolation that makes the generator unit-testable on
// its own. All knowledge of the outside world enters through one seam:
//
//   - Request: the caller's description of what happened (which Moment, in what
//     Tone, at what Age, about which Subject/Kind/Resource, with what numbers).
//     Adapters that translate an expedition, a faction, or a raid into a Request
//     live on the GAME side, exactly as factionProfile / boonApplier do for boon.
//
// The flow a caller uses is one line:
//
//	line := flavor.Line(flavor.Request{Moment: flavor.ExpeditionSuccess, ...}, ge.rng)
//
// Line/Generate are PURE given (req, rng): every random draw comes from the passed
// *rand.Rand in a FIXED order (template first, then each bank slot left to right),
// so the same seed reproduces the same prose stream. The package never seeds
// anything and never touches package-level math/rand.
//
// # Output contract
//
// Generated text is MARKUP-FREE: never a '[', a ']', or a '%'. The caller applies
// colour tags and prints the mechanical facts. Flavor rides ALONGSIDE mechanics —
// it never replaces the line that says what actually happened.
//
// # Voice
//
// Dry, occasionally absurdist, in-world. The joke is on the game world, never on
// the player, and never on the mechanic — nothing in here winks at ticks, rolls,
// caps, or slots. Matches config/building_flavor.go and config/log_flavor.go.
//
// # Every Request field is optional
//
// A ZERO Request (plus a registered Moment) still produces a grammatical, generic
// line. Templates declare which Request fields they NEED; a template whose needs
// are unmet is simply not eligible, and every Moment ships need-free templates.
// That is also how "0 = omit from the sentence" is enforced for Amount/Count/Ticks:
// there is no such thing as a rendered "0 food" here, because the templates that
// mention an amount cannot be picked without one.
package flavor

import (
	"math/rand"
	"sort"
)

// Moment enumerates the game situations this package can narrate. The zero value
// is deliberately MomentUnknown so an unset field can never silently pick a real
// moment; an unregistered Moment yields "" rather than panicking.
//
// The list is additive: adding a Moment plus its templates is a data edit.
type Moment int

const (
	// MomentUnknown is the zero value and is NOT registered. Generate returns an
	// empty Result for it.
	MomentUnknown Moment = iota
	// ExpeditionSuccess is an expedition that resolved in the player's favour.
	// Subject is the expedition's display NAME, which is a verb-led order title
	// ("Raid Bandit Camp") as often as a noun phrase ("Scout Party") — see the
	// subject-frame rules in catalog.go.
	ExpeditionSuccess
	// ExpeditionFailure is an expedition that resolved badly (partial loot).
	ExpeditionFailure
	// EncounterStandoff is contact with a civilization you are at war with that
	// did not turn violent: an outcome, not silence.
	EncounterStandoff
	// EncounterAtCapacity is an encounter that arrived while the court already
	// holds the maximum number of foreign favours.
	EncounterAtCapacity
	// WarRaid is a periodic raid by a civilization at war with you.
	WarRaid
)

// String returns a stable, flavor-free identifier for a Moment. Safe to log.
func (m Moment) String() string {
	switch m {
	case ExpeditionSuccess:
		return "ExpeditionSuccess"
	case ExpeditionFailure:
		return "ExpeditionFailure"
	case EncounterStandoff:
		return "EncounterStandoff"
	case EncounterAtCapacity:
		return "EncounterAtCapacity"
	case WarRaid:
		return "WarRaid"
	case MomentUnknown:
		return "MomentUnknown"
	default:
		return "Moment(?)"
	}
}

// Tone biases which templates are eligible. The zero value is Neutral, so a
// caller that does not care about tone gets the unconstrained pool.
//
// Tone-constrained templates are ADDITIVE: they widen a tone's pool rather than
// narrowing the others, because every Moment also ships templates with no tone
// constraint at all. Picking a tone therefore never makes output repetitive.
type Tone int

const (
	// Neutral is the default register: flat, observational.
	Neutral Tone = iota
	// Grim is for losses with a body count behind them.
	Grim
	// Triumphant is for a win the world will over-celebrate.
	Triumphant
	// Wry is for the absurd ones — a polite war, an embarrassment of gifts.
	Wry
)

// String returns a stable, flavor-free identifier for a Tone.
func (t Tone) String() string {
	switch t {
	case Neutral:
		return "Neutral"
	case Grim:
		return "Grim"
	case Triumphant:
		return "Triumphant"
	case Wry:
		return "Wry"
	default:
		return "Tone(?)"
	}
}

// Request describes what happened. EVERY field is optional — the zero Request
// (with a registered Moment) produces a generic, grammatical line.
type Request struct {
	// Moment selects the template set. Unregistered ⇒ empty output.
	Moment Moment
	// Tone widens the eligible template set with tone-specific phrasings.
	Tone Tone
	// Age is a config age key (config.AgeOrder()). It bounds era-appropriate
	// fragments — a Galactic Age survey team does not have gate wardens. Empty or
	// unknown means "no era bound", the permissive default.
	Age string
	// Subject is the display name of the actor. What it means depends on the
	// Moment, and the templates are authored accordingly:
	//   - expedition moments: the expedition's ORDER TITLE, which may be verb-led
	//     ("Conquer Territory"). It is only ever used in title frames.
	//   - faction moments: a civilization's name, whose grammatical NUMBER is
	//     inconsistent across the roster ("Merchant Guild" vs "Void Reavers").
	//     It is only ever used as an object or with past-tense verbs.
	Subject string
	// Kind is a taxonomy hint — an expedition category ("scouting"/"military"), a
	// faction personality, a building lineage key. Matched case-insensitively.
	// Unknown or empty simply means "no kind-specific templates".
	Kind string
	// Resource is a config resource KEY (not a display name). This package owns
	// the key→label mapping AND the countability of every key, so it can emit
	// "your food stores" and "12 soldiers" without ever producing "foods" or
	// "a knowledge".
	Resource string
	// Amount is a resource quantity. Rounded; anything below 1 counts as absent
	// and no amount-bearing template will be picked.
	Amount float64
	// Count is a head-count (workers, ships, casualties). 0 counts as absent.
	Count int
	// Ticks is a duration. 0 counts as absent.
	Ticks int
}

// Result is one generated line plus the structured facts about how it was made.
//
// Template is a STABLE template identifier. It exists so machinery downstream can
// classify generated prose without string-matching it — see also Signatures, which
// gives a consumer that only has the finished TEXT (a log line, say) an enumerable
// set of literal fragments to match against.
type Result struct {
	// Text is the finished, markup-free sentence. Empty for an unknown Moment.
	Text string
	// Moment echoes the request's Moment.
	Moment Moment
	// Template is the stable id of the template used. Empty when Text is empty.
	Template string
}

// Line generates one sentence for req. It is Generate(req, rng).Text — the
// convenience form for the common case where the caller just wants prose.
func Line(req Request, rng *rand.Rand) string {
	return Generate(req, rng).Text
}

// Moments returns the registered Moments in a stable order. Anything not in this
// slice generates an empty Result.
func Moments() []Moment {
	return []Moment{
		ExpeditionSuccess,
		ExpeditionFailure,
		EncounterStandoff,
		EncounterAtCapacity,
		WarRaid,
	}
}

// Capacity reports how many DISTINCT lines a Moment can produce across its whole
// template set: the sum over templates of the product of its slot bank sizes.
//
// It ignores Request-dependent eligibility on purpose — it answers "how much prose
// is authored here", which is the number worth asserting a floor on in a test. The
// pool a specific Request actually draws from is necessarily smaller.
//
// Returns 0 for an unregistered Moment.
func Capacity(m Moment) int {
	total := 0
	for _, t := range templatesFor(m) {
		total += t.capacity()
	}
	return total
}

// Signatures returns the set of authored fragments guaranteed to appear VERBATIM
// in generated text for a Moment — its "anchor" banks.
//
// Every template of a Moment is required to draw one slot from an anchor bank
// (enforced by TestAnchorInvariant), anchor fragments carry no placeholders, and
// nothing in the render path transforms case or substitutes inside a fragment. So
// for any generated line L of Moment m, exactly one s in Signatures(m) satisfies
// strings.Contains(L, s).
//
// That is the seam a consumer holding only finished TEXT needs in order to classify
// it — game/boon_tuning_test.go uses it to keep bucketing encounter outcomes now
// that those lines are generated rather than drawn from a fixed slice. Prefer
// Result.Template when you have the Result; use this when you only have the string.
//
// The returned slice is a fresh copy in a stable (sorted) order.
func Signatures(m Moment) []string {
	set := anchorBanksFor(m)
	var out []string
	for _, name := range set {
		out = append(out, banks[name]...)
	}
	sort.Strings(out)
	return out
}
