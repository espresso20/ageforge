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
// Dry, in-world, concrete, and mostly FLAT: at least three lines in ten carry no
// irony at all, because comedy needs a straight man and a pool where every line
// strains for a wry observation is exhausting. The joke is on the game world,
// never on the player, and never on the mechanic — nothing in here winks at ticks,
// rolls, caps, or slots. Matches config/building_flavor.go.
//
// Sentence LENGTH is deliberately uneven — four words next to thirty-eight — and
// that unevenness is a hard requirement rather than a stylistic preference. Two
// earlier versions of this catalog were rejected as machine-written, and the thing
// they had in common was a narrow band of sentence lengths: one sat at 12-25 words
// and the other at 6-14, and both read as generated. Uniform rhythm is the tell,
// independent of how good any individual line is. TestBurstiness asserts the
// distribution and a standard deviation floor; TestNoAITells bans the
// false-contrast and withheld-payload shapes; TestRegisterQuotas holds the mix of
// flat, wry, joking and non-narrative lines.
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
	"sync"
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
	// HarbingerArrival is the age's harbinger turning up before an epoch
	// transition. Subject is the harbinger's roster Name ("the Town Crier").
	// See harbinger.go for the request shape all seven harbinger Moments share.
	HarbingerArrival
	// HarbingerWarning is the warning itself: the cry, the headline, the leaked
	// memo, and the settlement's reaction to it. Kind is the risk tier, one of
	// TierNone, TierLow, TierMedium, TierHigh.
	HarbingerWarning
	// HarbingerAppeased follows the player paying to lower the odds.
	HarbingerAppeased
	// HarbingerBraced follows the player paying to soften the blow.
	HarbingerBraced
	// HarbingerVindicated is a harbinger who warned, and the catastrophe came.
	HarbingerVindicated
	// HarbingerSpared is a real harbinger who warned, and the roll went the
	// player's way.
	HarbingerSpared
	// HarbingerDiscredited is a false prophet found out after the transition.
	// Only reachable before the Industrial Age; see config.HarbingerDef.
	HarbingerDiscredited
	// HarbingerInvited is the settlement reacting to its leader inviting the
	// catastrophe on purpose.
	HarbingerInvited
	// HarbingerFulfilled is an invited catastrophe arriving. Kind is
	// KindFalseProphet when the harbinger who was invited had been lying.
	HarbingerFulfilled
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
	case HarbingerArrival:
		return "HarbingerArrival"
	case HarbingerWarning:
		return "HarbingerWarning"
	case HarbingerAppeased:
		return "HarbingerAppeased"
	case HarbingerBraced:
		return "HarbingerBraced"
	case HarbingerVindicated:
		return "HarbingerVindicated"
	case HarbingerSpared:
		return "HarbingerSpared"
	case HarbingerDiscredited:
		return "HarbingerDiscredited"
	case HarbingerInvited:
		return "HarbingerInvited"
	case HarbingerFulfilled:
		return "HarbingerFulfilled"
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
	// slot is the noun phrase drawn into the sentence's slot, if it had one. It is
	// unexported because it is not part of the contract — it exists so a Stream can
	// notice that two different sentences two lines apart both mentioned the long
	// rope, which skeleton identity cannot see.
	slot string
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
// A Stream suppresses on TWO axes, because skeleton identity alone missed a real
// complaint: two ADJACENT lines that were different sentences about the same
// thing (the dog, twice; the count, twice). Sentences carry a Topic from a closed
// vocabulary, and a topic that fired in the last few lines is redrawn too. The
// topic window is deliberately short — a catalog only has thirty topics, so a
// long one would starve the pool — while the skeleton window stays a screenful.
type Stream struct {
	recent []string
	at     int
	topics []string
	tat    int
	slots  []string
	sat    int
}

const (
	// streamMemory is how many recent skeletons a Stream will avoid reusing.
	// Sized to a screenful of log, which is the window a repeat is noticed in.
	streamMemory = 32
	// topicMemory is how many recent SUBJECTS a Stream will avoid returning to.
	// Short on purpose: adjacency is what reads badly, and the topic vocabulary
	// is small enough that a long window would just exhaust it.
	topicMemory = 4
	// slotMemory is how many recent NOUN-PHRASE DRAWS a Stream will avoid
	// repeating, so two different sentences a few lines apart do not both happen
	// to be about the long rope.
	slotMemory = 6
	// streamRetries is how many times a Stream will redraw before giving up on
	// the soft (topic, slot) filters. A skeleton repeat is then ruled out by a
	// direct pick from the unseen part of the pool; only a pool narrower than
	// streamMemory can still repeat, and it degrades to plain Generate rather
	// than looping.
	streamRetries = 12
)

// NewStream returns an empty Stream.
func NewStream() *Stream {
	return &Stream{
		recent: make([]string, streamMemory),
		topics: make([]string, topicMemory),
		slots:  make([]string, slotMemory),
	}
}

// topicIndex maps a skeleton id to its authored Topic. Built once: ids are unique
// across Moments (TestTemplatesAreWellFormed), so one flat map covers the catalog.
var topicIndex = sync.OnceValue(func() map[string]string {
	out := map[string]string{}
	for _, m := range Moments() {
		for _, t := range templatesFor(m) {
			out[t.ID] = t.Topic
		}
	}
	return out
})

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
	if s.topics == nil {
		s.topics = make([]string, topicMemory)
	}
	if s.slots == nil {
		s.slots = make([]string, slotMemory)
	}
	topics := topicIndex()
	// The eligible pool is computed once per line, not once per retry. Each
	// draw below is exactly what Generate would make (one Intn over the same
	// pool, then render), so the stream is unchanged; the rescans were most
	// of the cost for a pool only just wider than the window.
	pool := eligible(req)
	if len(pool) == 0 {
		return Result{Moment: req.Moment}
	}
	var res Result
	var topic string
	for i := 0; i <= streamRetries; i++ {
		res = render(pool[rng.Intn(len(pool))], req, rng)
		topic = topics[res.Template]
		if !s.seen(res.Template) && !s.sameTopicRecently(topic) && !s.sameSlotRecently(res.slot) {
			break
		}
	}
	// The topic and slot rings are soft preferences; the skeleton ring is a
	// promise. If every retry still landed on a sentence from the last
	// streamMemory lines, pick directly from the ones that are not, so a pool
	// wider than the window can never repeat inside it. Measured before this
	// existed: about one repeat per ten thousand lines, which is rare but is
	// exactly the "didn't I just read that" moment the Stream is for.
	//
	// The direct pick still prefers a sentence whose topic is not in the recent
	// ring when the unseen part of the pool offers one. For a pool much wider
	// than the window this almost never matters; for one only just wider (the
	// harbinger Moments, which fire a handful of times a run and are sized
	// accordingly) the unseen remainder is two or three sentences, and picking
	// blind among them put two same-topic lines side by side about one time in
	// thirty. Still one rng draw, so the draw count is unchanged.
	if res.Template != "" && s.seen(res.Template) {
		if fresh := s.unseen(pool); len(fresh) > 0 {
			if calm := s.offTopic(fresh, topics); len(calm) > 0 {
				fresh = calm
			}
			res = render(fresh[rng.Intn(len(fresh))], req, rng)
			topic = topics[res.Template]
		}
	}
	if res.Template != "" {
		s.recent[s.at] = res.Template
		s.at = (s.at + 1) % len(s.recent)
		if topic != "" {
			s.topics[s.tat] = topic
			s.tat = (s.tat + 1) % len(s.topics)
		}
		if res.slot != "" {
			s.slots[s.sat] = res.slot
			s.sat = (s.sat + 1) % len(s.slots)
		}
	}
	return res
}

// sameSlotRecently reports whether this noun-phrase draw is in the recent slot
// ring. A slotless sentence (empty draw) never suppresses.
func (s *Stream) sameSlotRecently(slot string) bool {
	if slot == "" {
		return false
	}
	for _, x := range s.slots {
		if x == slot {
			return true
		}
	}
	return false
}

// unseen returns the templates in pool that are not in the recent skeleton ring,
// preserving catalog order so the pick stays deterministic given the rng.
func (s *Stream) unseen(pool []tmpl) []tmpl {
	out := make([]tmpl, 0, len(pool))
	for _, t := range pool {
		if !s.seen(t.ID) {
			out = append(out, t)
		}
	}
	return out
}

// offTopic returns the templates in pool whose topic is not in the recent
// topic ring, preserving order.
func (s *Stream) offTopic(pool []tmpl, topics map[string]string) []tmpl {
	out := make([]tmpl, 0, len(pool))
	for _, t := range pool {
		if !s.sameTopicRecently(topics[t.ID]) {
			out = append(out, t)
		}
	}
	return out
}

// seen reports whether id is in the recent skeleton ring.
func (s *Stream) seen(id string) bool {
	for _, x := range s.recent {
		if x == id {
			return true
		}
	}
	return false
}

// sameTopicRecently reports whether topic is in the recent topic ring. An untagged
// sentence (empty topic) never suppresses, so a partially tagged catalog degrades
// to plain skeleton suppression rather than to nothing.
func (s *Stream) sameTopicRecently(topic string) bool {
	if topic == "" {
		return false
	}
	for _, x := range s.topics {
		if x == topic {
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
		HarbingerArrival,
		HarbingerWarning,
		HarbingerAppeased,
		HarbingerBraced,
		HarbingerVindicated,
		HarbingerSpared,
		HarbingerDiscredited,
		HarbingerInvited,
		HarbingerFulfilled,
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
