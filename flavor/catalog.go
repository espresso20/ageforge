package flavor

import "sync"

// catalog.go is the REGISTRY half of the data layer: it wires each Moment to its
// authored sentences and merges the noun banks those sentences draw from. The
// prose itself lives one file per Moment, next to the banks it uses.
//
// # What an author writes
//
// One finished SENTENCE at a time, of whatever length that sentence wants to be —
// three words or forty. It may carry at most one slot, and the slot is a noun
// phrase of a single tight category. It never gets welded to another sentence by
// the machine — see skeleton.go for why that is a deliberate hole in the model
// rather than a missing feature.
//
// # The eight rules the prose is held to
//
//  0. VARY THE LENGTH, ON PURPOSE. This is the rule two previous versions of this
//     catalog broke, and it is the one that got them rejected. About one line in
//     five is three to six words; about one in four runs past seventeen; one in ten
//     runs past thirty-three and is genuinely subordinated, stacking circumstance
//     the way a chronicle or a quartermaster's book actually runs. A catalog whose
//     lines are all the same LENGTH reads as machine-written no matter how well any
//     one of them is written. TestBurstiness asserts the distribution and a
//     standard-deviation floor of seven words.
//
//  1. ORTHOGONAL TO THE MECHANICS. The line above this one in the log already
//     says "Scout Party succeeded! Gained loot." The flavour must not say it
//     again in a wry voice. It supplies a DETAIL: something someone did, something
//     that came back, something that broke, a consequence, an image. The
//     vocabulary of outcome-restating — pays out, closes out in profit, the
//     accounting, goes into the record as a good one, the venture concludes —
//     is banned, and TestNoOutcomeRestating fails the build on it.
//
//  2. CONCRETE OVER OBSERVATIONAL. Name things. "Two did not come back. The rest
//     brought maps." beats any sentence about how the undertaking concluded. The
//     abstract crutches — the venture, the undertaking, the party as a bare
//     subject, the record, the accounting — are banned for the same reason.
//
//  3. MOSTLY FLAT, AND TAGGED SO. At most one line in five reaches for a joke and
//     at least three in ten carry no irony whatsoever — a circumstantial fact,
//     reported and then stopped. Comedy needs a straight man, and a catalog where
//     every line strains for a wry observation is exhausting and makes none of
//     them land. Every sentence declares its Reg, so the mix is structural rather
//     than aspirational (TestRegisterQuotas).
//
//  3b. NOT ALL NARRATION. At least one line in seven is a ledger entry, a grumble,
//     an overheard fragment or a posted notice, declared with Form. A log that is
//     nothing but narration has one voice, and one voice is the failure mode.
//
//  3c. NO FALSE CONTRAST. "X came back. Y did not." / "not X, but Y" / "isn't X,
//     it's Y" / a fragment followed by a verdict / a trailing "and that is the
//     whole problem". Every one of these builds to a weighted final clause, they
//     are a recognised signature of machine prose, and they are banned by lint
//     (TestNoAITells). They feel like good writing, which is exactly why the last
//     two versions of this file were saturated with them.
//
//  4. VARY THE BEAT. Not every line is about the return. Rotate the subject
//     matter: what was found, what broke, who did something, what the town did,
//     what the map says, a rumour, a count, an animal, the weather, a name. A
//     Moment wants a spread of EVENTS, not a spread of phrasings for one event.
//     Every sentence declares a Topic from a closed vocabulary, and Stream refuses
//     a topic it used in the last four lines — because two ADJACENT lines about
//     the dog read badly even when they are different sentences, and skeleton
//     identity cannot see that.
//
//  5. ERA-APPROPRIATE. Concrete writing is inherently era-bound — that is the
//     cost of rule 2 and it is worth paying. A sentence naming a physical thing
//     goes in an era-gated pool; only sentences whose imagery genuinely lands in
//     the Stone Age and the Quantum Age alike go in the ungated pool.
//     TestNoEraBleedInAuthoredText enforces both directions.
//
// # Voice
//
// Dry, in-world, specific. The joke is on the game world — its clerks, its
// councils, its logistics — never on the player and never on the mechanic.
// Nothing here winks at ticks, rolls, slots, or caps. config/building_flavor.go
// sets the bar.

// --- era gating shorthands ---------------------------------------------------
//
// Named so a pool declaration reads as prose. nil means ungated: the sentence
// fires in every age and must therefore contain no era-coded noun at all.

var (
	erasAny        []era
	erasAncient    = []era{eraAncient}
	erasFeudal     = []era{eraFeudal}
	erasIndustrial = []era{eraIndustrial}
	erasDigital    = []era{eraDigital}
	erasCosmic     = []era{eraCosmic}

	// Spans, for imagery that genuinely covers two neighbouring buckets — a road
	// and a river crossing belong to every age that has ground under it.
	erasEarly = []era{eraAncient, eraFeudal}
	erasMid   = []era{eraFeudal, eraIndustrial}
	erasLate  = []era{eraDigital, eraCosmic}
)

// eraBucket pairs an era with the suffix its pool ids use. The order is fixed so
// generated template sets are stable, which the determinism contract needs.
type eraBucket struct {
	era    era
	suffix string
}

var eraBuckets = []eraBucket{
	{eraAncient, "ancient"},
	{eraFeudal, "feudal"},
	{eraIndustrial, "industrial"},
	{eraDigital, "digital"},
	{eraCosmic, "cosmic"},
}

// banks holds every noun-phrase bank, merged from the per-Moment files. Bank
// fragments are always lowercase mid-sentence noun phrases with no terminal
// punctuation: they drop into a slot, they never open a sentence.
var banks = mergeBanks(
	expSuccessBanks,
	expFailBanks,
	encStandoffBanks,
	encCapacityBanks,
	warRaidBanks,
)

// mergeBanks folds the bank tables into one map. A key collision is an authoring
// mistake that would silently discard a whole bank, so it panics at init rather
// than shipping a half-empty pool.
func mergeBanks(tables ...map[string][]string) map[string][]string {
	out := make(map[string][]string)
	for _, t := range tables {
		for name, bank := range t {
			if _, dup := out[name]; dup {
				panic("flavor: duplicate bank name " + name)
			}
			out[name] = bank
		}
	}
	return out
}

// templatesFor returns a Moment's skeleton set in a stable order. An unregistered
// Moment returns nil, which is what makes Generate yield an empty Result for it.
//
// The set is built once and shared: callers must treat the returned slice as
// read-only. Rebuilding every pool on every Generate call was the single largest
// cost in a draw, and the catalog is immutable after init anyway.
func templatesFor(m Moment) []tmpl {
	return templateCache()[m]
}

// templateCache builds every registered Moment's skeleton set exactly once. A
// sync.OnceValue rather than an init-time var so the cost is paid on first use,
// and so concurrent first calls (two engines in one test binary) are race-free.
var templateCache = sync.OnceValue(func() map[Moment][]tmpl {
	out := make(map[Moment][]tmpl, len(Moments()))
	for _, m := range Moments() {
		out[m] = buildTemplates(m)
	}
	return out
})

// buildTemplates assembles one Moment's skeleton set from its per-era pools.
func buildTemplates(m Moment) []tmpl {
	switch m {
	case ExpeditionSuccess:
		return expSuccessTemplates()
	case ExpeditionFailure:
		return expFailTemplates()
	case EncounterStandoff:
		return encStandoffTemplates()
	case EncounterAtCapacity:
		return encCapacityTemplates()
	case WarRaid:
		return warRaidTemplates()
	default:
		return nil
	}
}
