package flavor

// harbinger.go is the shared plumbing for the nine harbinger Moments: the
// risk-tier kinds, the per-age pool helper, and the request shape the game
// side is expected to build. The prose lives in catalog_harb_*.go.
//
// # The feature, in one paragraph
//
// Through an epoch whose transition can roll a catastrophe, each age's
// harbinger (config.HarbingerFor) takes up the warning in turn. How frightened
// they are follows the risk tier. Before the Industrial Age they are vague;
// from the Industrial Age on their medium prints the odds, which the UI shows
// as a number, so the prose refers to "the figure" and never contains a digit.
// The player can appease, brace, invite the doom on purpose, or ignore it. In
// the Stone, Iron and Steel Eras the whole warning may be false, and after the
// transition the log says so.
//
// # The request shape
//
// Every harbinger Moment takes the same Request:
//
//	flavor.Request{
//		Moment:  flavor.HarbingerWarning,
//		Age:     age,          // the CURRENT age; selects the speaker's voice
//		Subject: h.Name,       // config.HarbingerDef.Name, e.g. "the Town Crier"
//		Kind:    flavor.TierMedium, // see Kind below
//	}
//
// Age does more work here than anywhere else in the package. The other Moments
// use it to pick an era bucket; these also carry per-age pools (tmpl.Ages),
// because the harbinger changes every age and a line quoting the newsboy's
// headline cannot fire under the civil-defence broadcast two ages later.
//
// Subject is the harbinger's display name, always a singular noun phrase that
// opens lowercase ("the Oracle", "your future self"), so it only ever sits
// mid-sentence. Leave it empty and the {subject} lines drop out; everything
// else still works.
//
// Kind matters to two Moments. HarbingerWarning takes the risk tier, whose
// strings are game.CatastropheTier's values, so the caller can pass
// string(outlook.Tier) straight through. HarbingerFulfilled takes
// KindFalseProphet when the invited harbinger had been lying. The rest ignore
// it. With no tier, HarbingerWarning falls back
// to a small set of tier-neutral reaction lines, which is what keeps a zero
// Request grammatical, but a caller should always send the tier: the warning
// is the one place pre-Industrial players learn how worried to be.
//
// # Output
//
// Same contract as the rest of the package, plus one rule of its own: no digit
// ever appears in a harbinger line, because the odds are printed by the UI and
// a second number in the prose would disagree with it sooner or later. No
// harbinger sentence uses the amount, count or resource placeholders, and
// TestHarbingerTemplatesUseNoFigures and TestHarbingerLinesAreClean hold the
// catalog and the output to it.

// Risk tiers, sent as Request.Kind on HarbingerWarning. Plain strings because
// Kind is a string everywhere else in the package; compared lowercased.
const (
	// TierNone is a harbinger who comes, looks round, and finds nothing to fear.
	TierNone = "none"
	// TierLow is a small risk: an uneasy harbinger.
	TierLow = "low"
	// TierMedium is a real risk worth paying to reduce.
	TierMedium = "medium"
	// TierHigh is a harbinger in open terror.
	TierHigh = "high"
)

// HarbingerTiers returns the four tiers in ascending order of risk.
func HarbingerTiers() []string {
	return []string{TierNone, TierLow, TierMedium, TierHigh}
}

// Tier sets, named so a pool reads as prose. Adjacent pairs exist because most
// reaction lines honestly fit two neighbouring tiers: people going back to
// work fits "nothing to fear" and "a little to fear" alike.
var (
	tNone    = []string{TierNone}
	tLow     = []string{TierLow}
	tMed     = []string{TierMedium}
	tHigh    = []string{TierHigh}
	tNoneLow = []string{TierNone, TierLow}
	tLowMed  = []string{TierLow, TierMedium}
	tMedHigh = []string{TierMedium, TierHigh}
)

// agePool is pool for sentences that belong to particular ages rather than to
// a whole era bucket: the speaker's own voice. Eras is derived from the ages,
// so the era-bleed lint holds a per-age line to its bucket's vocabulary.
func agePool(prefix string, ages []string, ss []skel) []tmpl {
	var eras []era
	seen := map[era]bool{}
	for _, a := range ages {
		e, ok := eraOf(a)
		if !ok {
			panic("flavor: agePool given unknown age " + a)
		}
		if !seen[e] {
			seen[e] = true
			eras = append(eras, e)
		}
	}
	out := pool(prefix, eras, ss)
	for i := range out {
		out[i].Ages = ages
	}
	return out
}

// ageSet is one per-age batch of sentences in a pool declaration.
type ageSet struct {
	ages []string
	ss   []skel
}

// agePools builds every per-age batch in declaration order, which is fixed, so
// the template set is stable for the determinism contract.
func agePools(prefix string, sets []ageSet) []tmpl {
	var out []tmpl
	for _, s := range sets {
		out = append(out, agePool(prefix+"_"+s.ages[0], s.ages, s.ss)...)
	}
	return out
}

// ages is shorthand for an age-key list in a pool declaration.
func ages(keys ...string) []string { return keys }

// harbingerBanks holds the few noun banks the harbinger sentences draw on.
// Each is era-bound through the pools that reference it.
var harbingerBanks = map[string][]string{
	// Things left at a shrine or a standing stone. Ancient bucket only.
	"harb_offering_ancient": {
		"a string of shells", "the best antler in the camp", "a pot of honey",
		"a whole smoked fish", "a knife nobody could spare", "a bowl of red ochre",
		"two good flints", "a pelt still soft from the scraping",
	},
	// What a feudal household gives the church to be prayed for.
	"harb_offering_feudal": {
		"a silver spoon", "a bolt of blue cloth", "a year's worth of candles",
		"a sheep and her lamb", "the second-best bed", "a psalter with a cracked board",
		"a purse of old pennies", "a cask of the good wine",
	},
	// What people carry down into the shelter, industrial bucket.
	"harb_shelter_industrial": {
		"a box of tinned peaches", "a wind-up radio", "the family photographs",
		"a sack of flour", "a first-aid tin", "a deck of cards with two missing",
		"three jerrycans of water", "a crate of candles",
	},
}
