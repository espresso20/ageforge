// Package flavor is a standalone, reusable procedural prose service. It turns a
// structured description of a game moment into ONE varied, in-world, player-facing
// sentence, so the surfaces that fire hundreds of times per run stop reading like
// the same string on a loop.
//
// It is a SELECTION engine, not a composition one, and that distinction is the
// whole design. The unit of authorship is a finished SENTENCE somebody wrote and
// read back — a "skeleton" — which may carry at most one slot, and that slot only
// ever swaps a NOUN PHRASE from a bank of the same tight category. Nothing here
// joins two independently drawn clauses, because a generator with no semantic
// model cannot know that "the quartermaster looks at what returned" and "the
// quartermaster says nothing" are the same person twice. That is not a gap to be
// filled in later; it is the failure mode this package was rewritten to remove.
//
// The number that follows from it: repetition is governed by how many separate
// things a Moment can SAY, which is DistinctSkeletons, and not by the product of
// its bank sizes, which is Capacity. Reach for Capacity only when sizing storage.
//
// For anything a player reads as a RUN of lines, use a Stream rather than Line —
// Generate is memoryless, and memoryless uniform sampling repeats inside a
// screenful no matter how big the catalog gets.
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
// *rand.Rand in a FIXED order (skeleton first, then its slot if it has one), so the
// same seed reproduces the same prose stream. The package never seeds anything and
// never touches package-level math/rand.
//
// # Output contract
//
// Generated text is MARKUP-FREE: never a '[', a ']', or a '%'. The caller applies
// colour tags and prints the mechanical facts. Flavor rides ALONGSIDE mechanics —
// it never replaces the line that says what actually happened.
//
// Output is also ORTHOGONAL to it. The caller has already printed "Scout Party
// succeeded! Gained loot."; a flavour line that says the venture paid out is
// padding. These sentences supply a detail instead — what came back, who did not,
// what broke, what the town did about it.
//
// # Voice
//
// Dry, in-world, concrete, six to fourteen words, and mostly FLAT: roughly one
// line in five reaches for a joke, because comedy needs a straight man and a pool
// where every line strains for a wry observation is exhausting. The joke is on the
// game world, never on the player, and never on the mechanic — nothing in here
// winks at ticks, rolls, caps, or slots. Matches config/building_flavor.go.
//
// # Every Request field is optional
//
// A ZERO Request (plus a registered Moment) still produces a grammatical, generic
// line. Skeletons declare which Request fields they NEED; one whose needs are unmet
// is simply not eligible, and every Moment ships need-free sentences. That is also
// how "0 = omit from the sentence" is enforced for Amount/Count/Ticks: there is no
// such thing as a rendered "0 food" here, because the sentences that mention an
// amount cannot be picked without one.
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

// --- Stream: the anti-repeat wrapper a real log wants ----------------------

// Stream is a caller-owned Generate that avoids repeating a sentence it used
// recently. Use it for anything a player reads as a RUN of lines; use Line or
// Generate directly for one-offs and for tests that need purity.
//
// The reason this is a separate type rather than behaviour inside Generate is
// the determinism contract. Generate is pure given (req, rng) — the same seed
// reproduces the same prose, which is what lets a run's persisted seed replay its
// whole log — and recent-history suppression is by definition stateful. So the
// state lives with the CALLER, which also gets the semantics right: one Stream
// per log, deduplicating across every Moment that writes into it, rather than a
// package-global that would couple unrelated callers together.
//
// A Stream is deterministic given its own history and the rng, so a seeded replay
// through the same Stream reproduces exactly. It is not safe for concurrent use;
// the engine calls it under its write lock.
//
// Why it is needed at all, given a catalog of a couple of hundred sentences per
// Moment: a uniform draw of 30 samples from a pool of N produces about
// 30·29/(2N) repeated pairs, which is five repeats in thirty lines at N = 81 and
// still one at N = 400. Repetition inside a screenful is a property of memoryless
// sampling, not of catalog size, and no amount of extra authoring fixes it.
type Stream struct {
	recent []string
	at     int
}

const (
	// streamMemory is how many recent skeletons a Stream will avoid reusing.
	// Sized to a screenful of log, which is the window a repeat is noticed in.
	streamMemory = 32
	// streamRetries is how many times a Stream will redraw before accepting a
	// repeat. It always accepts eventually, so a narrow eligible pool degrades to
	// plain Generate instead of looping.
	streamRetries = 5
)

// NewStream returns an empty Stream.
func NewStream() *Stream { return &Stream{recent: make([]string, streamMemory)} }

// Line is Stream.Generate(req, rng).Text.
func (s *Stream) Line(req Request, rng *rand.Rand) string {
	return s.Generate(req, rng).Text
}

// Generate produces one line for req, redrawing up to streamRetries times if the
// skeleton is one of the last streamMemory used. A nil Stream falls back to the
// plain generator, so a caller that has not built one still gets prose.
func (s *Stream) Generate(req Request, rng *rand.Rand) Result {
	if s == nil {
		return Generate(req, rng)
	}
	if s.recent == nil {
		s.recent = make([]string, streamMemory)
	}
	var res Result
	for i := 0; i <= streamRetries; i++ {
		res = Generate(req, rng)
		if res.Template == "" || !s.seen(res.Template) {
			break
		}
	}
	if res.Template != "" {
		s.recent[s.at] = res.Template
		s.at = (s.at + 1) % len(s.recent)
	}
	return res
}

// seen reports whether id is in the recent ring.
func (s *Stream) seen(id string) bool {
	for _, x := range s.recent {
		if x == id {
			return true
		}
	}
	return false
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

// Capacity reports how many distinct STRINGS a Moment can produce: the sum over
// skeletons of its slot bank size (1 for a slotless sentence).
//
// It is kept for compatibility and for rough sizing, and it is NOT the quality
// bar. It never was a good one: a Moment whose lines all draw a noun from a bank
// of forty reports a capacity of forty per skeleton while a player sees the same
// sentence over and over with a different noun in it. Perceived repetition is
// governed by DistinctSkeletons — how many separate things a Moment can SAY — and
// that is the number tests assert a floor on.
//
// Returns 0 for an unregistered Moment.
func Capacity(m Moment) int {
	total := 0
	for _, t := range templatesFor(m) {
		total += t.capacity()
	}
	return total
}

// DistinctSkeletons reports how many separate authored sentences a Moment has.
//
// This is the repetition metric. A skeleton is one sentence a person wrote; a
// slot inside it swaps a noun, which varies the detail but not the beat, so a
// Moment with six skeletons and a thousand nouns still reads as six lines on a
// loop. The floor that matters is per ERA-ELIGIBLE POOL — the sentences an age in
// one era bucket can actually reach — because that is the pool a real run draws
// from; see TestSkeletonFloors.
//
// Returns 0 for an unregistered Moment.
func DistinctSkeletons(m Moment) int {
	return len(templatesFor(m))
}

// Signatures returns the set of authored fragments guaranteed to appear VERBATIM
// in generated text for a Moment — one per skeleton.
//
// Every skeleton opens with authored literal text, and nothing in the render path
// transforms case or rewrites the inside of a literal, so a skeleton's leading run
// (everything before its slot or its first placeholder) reaches the finished line
// character for character. So for any generated line L of Moment m, exactly one s
// in Signatures(m) satisfies strings.Contains(L, s) — enforced end to end by
// TestSignaturesClassifyEveryLine.
//
// That is the seam a consumer holding only finished TEXT needs in order to classify
// it — game/boon_tuning_test.go uses it to keep bucketing encounter outcomes now
// that those lines are generated rather than drawn from a fixed slice. Prefer
// Result.Template when you have the Result; use this when you only have the string.
//
// The returned slice is a fresh copy in a stable (sorted) order.
func Signatures(m Moment) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range templatesFor(m) {
		s := anchorOf(t)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
