package flavor

// catalog.go is the REGISTRY half of the data layer: it wires each Moment to its
// authored sentences and merges the noun banks those sentences draw from. The
// prose itself lives one file per Moment, next to the banks it uses.
//
// # What an author writes
//
// One finished SENTENCE at a time. Six to fourteen words. It may carry at most
// one slot, and the slot is a noun phrase of a single tight category. It never
// gets welded to another sentence by the machine — see skeleton.go for why that
// is a deliberate hole in the model rather than a missing feature.
//
// # The five rules the prose is held to
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
//  3. MOSTLY FLAT. Roughly one line in five reaches for a joke. Comedy needs a
//     straight man, and a catalog where every line strains for a wry observation
//     is exhausting and makes none of them land. The majority simply report
//     something specific and let the world speak.
//
//  4. VARY THE BEAT. Not every line is about the return. Rotate the subject
//     matter: what was found, what broke, who did something, what the town did,
//     what the map says, a rumour, a count, an animal, the weather, a name. A
//     Moment wants a spread of EVENTS, not a spread of phrasings for one event.
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
func templatesFor(m Moment) []tmpl {
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
