// Package boon is a standalone, reusable random-reward service. It rolls a
// weighted "boon" — a temporary or instant buff — and applies it through a
// narrow Applier interface, so ANY caller (faction encounters today; events,
// milestones, or prestige tomorrow) can draw boons without the boon logic
// knowing anything about them.
//
// Decoupling is the whole point. This package imports AT MOST config (for the
// resource-key / effect vocabulary) and the standard library. It does NOT
// import game — that would be an import cycle and would defeat the isolation
// that makes the roll engine unit-testable with a fake Applier. All knowledge
// of the outside world enters through two seams:
//
//   - Profile: the caller's roll-shaping (which Kinds are live, per-kind weight
//     multipliers, magnitude/rarity scale, the target specialty + current age).
//     This is how a faction's identity and standing will later bias the table
//     WITHOUT this package ever learning what a faction is.
//   - Applier: the minimal set of side-effecting calls a consumer implements,
//     each mapping to machinery the game already has (timed effects, resource
//     grants, worker loans).
//
// The flow a caller uses is three lines:
//
//	b := boon.RollBoon(profile, rng) // rng is the caller's seeded *rand.Rand
//	line := boon.Apply(b, applier)   // all boon→effect translation happens here
//	log(line)                        // the returned flavor line is player-facing
//
// RollBoon is pure given (profile, rng): it draws only from the passed
// *rand.Rand, in a fixed order, so a run's whole boon stream is reproducible
// from its seed. It never touches package-level rand.
package boon

import "math"

// Kind enumerates the categories of boon this engine can grant. The first five
// map to machinery the game already exposes and are catalogued below. The last
// two are reserved: they need engine plumbing that does not exist yet.
type Kind int

const (
	// RateBuff is a timed +Magnitude fraction on a single resource's rate,
	// carried as a "<resource>_rate" effect.
	RateBuff Kind = iota
	// AllProduction is a timed +Magnitude fraction on all production, carried
	// as a "production_all" effect.
	AllProduction
	// TickSpeed is a timed +Magnitude fraction on tick speed, carried as a
	// "tick_speed" effect.
	TickSpeed
	// InstantResource is a one-shot lump grant of a resource (no duration).
	InstantResource
	// TempWorkers lends N free workers for a duration.
	TempWorkers

	// StorageCap (reserved, NOT catalogued) would be a timed bump to a
	// resource's storage cap. It needs a storage-cap modifier pool the engine
	// does not have yet — there is no Applier path for it, so Apply no-ops it.
	// Add a Def and an Applier method once that plumbing lands.
	StorageCap
	// TradeIncome (reserved, NOT catalogued) would be a timed bump to trade
	// income. It needs a trade-income modifier the engine does not have yet;
	// same story as StorageCap — enum reserved, no catalog entry, Apply no-ops.
	TradeIncome

	// --- Negative-only kinds (the malus catalogue) ---------------------------
	// The timed kinds above double as maluses simply by carrying a NEGATIVE
	// Magnitude; the two below have no positive counterpart and exist only on
	// the malus table.

	// ResourceDrain removes a FRACTION of a resource's current stockpile, and
	// optionally carries a short production dip alongside it. InstantAmount
	// holds the drain fraction (0.10 = 10% of current); Magnitude/DurationTicks
	// hold the optional dip (both zero ⇒ pure drain, no dip).
	ResourceDrain
	// WorkerLoss removes N workers outright. InstantAmount holds the count
	// (rounded on apply); there is no duration — the dead do not come back.
	WorkerLoss
)

// Polarity says whether an entry HELPS or HURTS. It is what lets one engine roll
// both tables: a malus is just a boon with negative sign, drawn off a separate
// catalog. The zero value is Positive, so every pre-existing Def and Profile
// keeps its old meaning without touching it.
type Polarity int

const (
	// Positive is a boon: the player gains something.
	Positive Polarity = iota
	// Negative is a malus: the player loses something (a negative-magnitude
	// timed effect, a resource drain, or dead workers).
	Negative
)

// String returns a stable, flavor-free identifier for a Polarity.
func (p Polarity) String() string {
	switch p {
	case Positive:
		return "Positive"
	case Negative:
		return "Negative"
	default:
		return "Polarity(?)"
	}
}

// String returns a stable, plain identifier for a Kind. Used for logs and
// tests; deliberately free of flavor so it is safe to surface anywhere.
func (k Kind) String() string {
	switch k {
	case RateBuff:
		return "RateBuff"
	case AllProduction:
		return "AllProduction"
	case TickSpeed:
		return "TickSpeed"
	case InstantResource:
		return "InstantResource"
	case TempWorkers:
		return "TempWorkers"
	case StorageCap:
		return "StorageCap"
	case TradeIncome:
		return "TradeIncome"
	case ResourceDrain:
		return "ResourceDrain"
	case WorkerLoss:
		return "WorkerLoss"
	default:
		return "Kind(?)"
	}
}

// TargetRule tells RollBoon how to resolve a Def's concrete target resource.
type TargetRule int

const (
	// TargetNone means the boon has no resource target (production_all,
	// tick_speed, temp workers).
	TargetNone TargetRule = iota
	// TargetSpecialty targets the Profile-supplied specialty resource.
	TargetSpecialty
	// TargetRandomAge targets a random resource unlocked at or before the
	// Profile's current age.
	TargetRandomAge
	// TargetSpecificResource targets the Def's own Resource field.
	TargetSpecificResource
	// TargetRandomRare targets a random rare/knowledge resource that is also
	// age-appropriate.
	TargetRandomRare
)

// Boon is a concrete, rolled reward instance — fully self-describing so a
// consumer can log it and Apply can dispatch it without re-consulting the
// catalog.
//
// Field usage by Kind:
//   - RateBuff / AllProduction / TickSpeed: Magnitude (fraction) + DurationTicks.
//     Resource is set for RateBuff, empty otherwise.
//   - InstantResource: InstantAmount (lump) + Resource; DurationTicks is 0.
//   - TempWorkers: InstantAmount holds the worker COUNT (rounded on apply) and
//     DurationTicks the loan window; Resource is empty, Magnitude is 0.
//   - ResourceDrain (malus): InstantAmount holds the drain FRACTION of current
//     stock + Resource; Magnitude/DurationTicks optionally carry a short
//     production dip (zero ⇒ no dip).
//   - WorkerLoss (malus): InstantAmount holds the worker COUNT lost; no duration.
type Boon struct {
	Kind          Kind
	Polarity      Polarity // Positive = boon, Negative = malus (copied from the Def)
	Resource      string   // target resource key, where relevant ("" if none)
	Magnitude     float64  // fractional buff, e.g. 0.13 = +13%; NEGATIVE for a malus
	DurationTicks int      // buff/loan/debuff length in ticks; 0 = instant
	InstantAmount float64  // lump grant, worker count, or drain fraction (see above)
	Name          string   // catalog name; used as the injected effect's name
	Flavor        string   // finished, player-facing flavor line
}

// Def is a catalog entry: the tunable DATA describing one kind of boon and the
// ranges RollBoon draws from. Starting values live directly in the Def literals
// in catalog.go, so tuning is a data edit — no logic changes.
//
// Which range fields apply depends on Kind:
//   - RateBuff / AllProduction / TickSpeed: [MagMin,MagMax] + [DurMin,DurMax].
//     A malus entry simply uses a NEGATIVE magnitude range (MagMin ≤ MagMax
//     still holds: e.g. MagMin -0.15, MagMax -0.08).
//   - InstantResource: [AmountMin,AmountMax]; no duration.
//   - TempWorkers: [AmountMin,AmountMax] (worker count) + [DurMin,DurMax].
//   - ResourceDrain: [AmountMin,AmountMax] as a FRACTION of current stock
//     (0.05 = 5%), plus an OPTIONAL [MagMin,MagMax]+[DurMin,DurMax] production
//     dip. Leave the mag/dur ranges zero for a pure drain.
//   - WorkerLoss: [AmountMin,AmountMax] (worker count); no duration.
type Def struct {
	Kind      Kind
	Polarity  Polarity // Positive (default) or Negative; must match the catalog it lives in
	Name      string
	MagMin    float64    // magnitude range (percentage buffs)
	MagMax    float64    //
	DurMin    int        // duration range in ticks (timed boons)
	DurMax    int        //
	AmountMin float64    // lump/worker-count range (InstantResource, TempWorkers)
	AmountMax float64    //
	Weight    int        // base rarity weight (higher = more common)
	Target    TargetRule // how to resolve the concrete target resource
	Resource  string     // used only when Target == TargetSpecificResource
	Flavors   []string   // flavor-template pool; supports {pct} {res} {ticks} {amt} {n}
}

// Profile is the caller's roll-shaping. It biases the table without this
// package knowing why — a faction system fills Specialty/Age from the faction
// and can lift WeightMult / MagnitudeScale / RarityScale as standing grows.
type Profile struct {
	// Specialty is the resource key a TargetSpecialty boon aims at. Empty falls
	// back to the first age-appropriate resource.
	Specialty string
	// Age is the current age key; it bounds the "age-appropriate" resource
	// pools AND sets the instant-lump age-growth factor (see instantScale).
	// Empty/unknown means "everything unlocked" at age index 0 — a permissive
	// pool with the most conservative lump scaling.
	Age string
	// Polarity selects WHICH TABLE RollBoon draws from: Positive ⇒ Catalog()
	// (the boons), Negative ⇒ MalusCatalog() (the setbacks). The zero value is
	// Positive, so an untouched Profile behaves exactly as it always has.
	Polarity Polarity
	// Enabled is a denylist: a Kind rolls UNLESS Enabled[k] == false. A nil map
	// (the default) leaves every kind live.
	Enabled map[Kind]bool
	// WeightMult multiplies a Kind's base weight. Absent/nil ⇒ 1.0.
	WeightMult map[Kind]float64
	// MagnitudeScale scales rolled percentage magnitudes, the InstantResource
	// lump (on top of the age-growth factor), and the ResourceDrain fraction.
	// Worker COUNTS (TempWorkers, WorkerLoss) stay unscaled — they are discrete
	// and a scaled count is a cliff, not a curve. <= 0 is treated as 1.0.
	MagnitudeScale float64
	// RarityScale biases toward rarer entries. Effective weight is
	// baseWeight^(1/RarityScale): > 1 compresses the weight spread so rare
	// (low-weight) boons get relatively more probability; < 1 lets common boons
	// dominate. <= 0 is treated as 1.0.
	RarityScale float64
}

// DefaultProfile returns a neutral profile: every kind enabled, unit weights,
// unit scales. Callers set Specialty/Age and lift the knobs as they see fit.
func DefaultProfile() Profile {
	return Profile{
		MagnitudeScale: 1.0,
		RarityScale:    1.0,
	}
}

// kindEnabled reports whether a Kind may roll under this profile.
func (p Profile) kindEnabled(k Kind) bool {
	if p.Enabled == nil {
		return true
	}
	v, ok := p.Enabled[k]
	return !ok || v
}

// kindWeightMult returns the per-kind weight multiplier (default 1.0).
func (p Profile) kindWeightMult(k Kind) float64 {
	if p.WeightMult == nil {
		return 1.0
	}
	if m, ok := p.WeightMult[k]; ok {
		return m
	}
	return 1.0
}

// magScale returns the magnitude scale, guarding non-positive values.
func (p Profile) magScale() float64 {
	if p.MagnitudeScale <= 0 {
		return 1.0
	}
	return p.MagnitudeScale
}

// rarityScale returns the rarity scale, guarding non-positive values.
func (p Profile) rarityScale() float64 {
	if p.RarityScale <= 0 {
		return 1.0
	}
	return p.RarityScale
}

// effectiveWeight is a Def's post-profile weight for the weighted pick. It folds
// in the rarity bias (an exponent on the base weight) and the per-kind
// multiplier. Returns 0 for a non-positive base weight.
func (p Profile) effectiveWeight(d Def) float64 {
	if d.Weight <= 0 {
		return 0
	}
	return math.Pow(float64(d.Weight), 1.0/p.rarityScale()) * p.kindWeightMult(d.Kind)
}
