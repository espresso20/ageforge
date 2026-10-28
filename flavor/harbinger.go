package flavor

// harbinger.go is the shared plumbing for the nine harbinger Moments: the
// risk-tier kinds and the request shape the game side is expected to build.
// The prose lives in text/harb_*.txt.
//
// # The feature, in one paragraph
//
// When a doom is fated in an era, a harbinger comes some while before it
// strikes, and each age's harbinger (config.HarbingerFor) takes up the warning
// in turn until it does. How frightened they are follows the risk tier. Before
// the Industrial Age they are vague; from the Industrial Age on their medium
// prints the odds, which the UI shows as a number, so the prose refers to "the
// figure" and never contains a digit. The player can appease, brace, invite
// the doom on purpose, or ignore it. In the Stone, Iron and Steel Eras the
// whole warning may be false, and once its foretold window passes the log
// says so.
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

// KindFalseProphet is the Request.Kind for HarbingerFulfilled and
// HarbingerVindicated when the warning had been invented.
const KindFalseProphet = "false_prophet"

// HarbingerTiers returns the four tiers in ascending order of risk.
func HarbingerTiers() []string {
	return []string{TierNone, TierLow, TierMedium, TierHigh}
}
