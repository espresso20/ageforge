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
)

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
type Boon struct {
	Kind          Kind
	Resource      string  // target resource key, where relevant ("" if none)
	Magnitude     float64 // fractional buff, e.g. 0.13 = +13% (0 for instant/worker kinds)
	DurationTicks int     // buff/loan length in ticks; 0 = instant
	InstantAmount float64 // lump grant (InstantResource) or worker count (TempWorkers)
	Name          string  // catalog name; used as the injected effect's name
	Flavor        string  // finished, player-facing flavor line
}

// Def is a catalog entry: the tunable DATA describing one kind of boon and the
// ranges RollBoon draws from. Starting values live directly in the Def literals
// in catalog.go, so tuning is a data edit — no logic changes.
//
// Which range fields apply depends on Kind:
//   - RateBuff / AllProduction / TickSpeed: [MagMin,MagMax] + [DurMin,DurMax].
//   - InstantResource: [AmountMin,AmountMax]; no duration.
//   - TempWorkers: [AmountMin,AmountMax] (worker count) + [DurMin,DurMax].
type Def struct {
	Kind      Kind
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
	// pools. Empty/unknown means "everything unlocked" (a permissive default).
	Age string
	// Enabled is a denylist: a Kind rolls UNLESS Enabled[k] == false. A nil map
	// (the default) leaves every kind live.
	Enabled map[Kind]bool
	// WeightMult multiplies a Kind's base weight. Absent/nil ⇒ 1.0.
	WeightMult map[Kind]float64
	// MagnitudeScale scales rolled percentage magnitudes (Magnitude field only;
	// instant lump sizes and worker counts keep their Def ranges in v1). <= 0 is
	// treated as 1.0.
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
