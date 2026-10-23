// Package boon rolls weighted random rewards (boons) and setbacks (maluses) and
// applies them through the Applier interface. It does not know who asks for a
// roll. Faction encounters use it.
//
// The package imports only config (effects, ages, resources), detmath, pkg/textfmt
// (number and duration wording) and the standard library. It must not import game:
// that would be an import cycle, and the roll engine is tested with a fake Applier
// (TestNoGameImport checks it). Everything about the outside world comes in through
// two types:
//
//   - Profile: the caller's roll-shaping (which Kinds are live, per-kind weight
//     multipliers, magnitude and rarity scale, the target specialty and the
//     current age). A faction's character and standing bias the table through it.
//   - Applier: the calls a consumer implements to carry out a boon (timed
//     effects, resource grants, worker loans, drains and worker losses).
//
// A caller uses it in three steps:
//
//	b := boon.RollBoon(profile, rng) // rng is the caller's seeded *rand.Rand
//	line := boon.Apply(b, applier)   // all boon→effect translation happens here
//	log(line)                        // the returned flavor line is player-facing
//
// RollBoon depends only on (profile, rng). It draws from the passed *rand.Rand in
// a fixed order, so a run's boon stream is reproducible from its seed. It never
// uses package-level rand.
package boon

import (
	"time"

	"github.com/espresso20/ageforge/detmath"
)

// Kind is the category of a boon. RateBuff through TempWorkers are the kinds in
// Catalog. StorageCap and TradeIncome are reserved. ResourceDrain and WorkerLoss
// are negative-only and appear only in MalusCatalog.
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

	// StorageCap is reserved. It has no catalog entry and no Applier method, and
	// Apply does nothing for it. Stock has no cap (ResourceManager.Add never
	// clips a credit), so there is nothing for it to raise.
	StorageCap
	// TradeIncome is reserved: it would be a timed bump to trade income. It has
	// no catalog entry and no Applier method, and Apply does nothing for it.
	TradeIncome

	// Negative-only kinds. The timed kinds above become maluses by carrying a
	// negative Magnitude. The two below have no positive counterpart and appear
	// only in MalusCatalog.

	// ResourceDrain removes a fraction of a resource's current stock, and
	// optionally carries a short production dip. InstantAmount holds the drain
	// fraction (0.10 = 10% of current stock). Magnitude and DurationTicks hold
	// the optional dip; both zero means a pure drain.
	ResourceDrain
	// WorkerLoss removes workers outright. InstantAmount holds the count
	// (rounded on apply). There is no duration.
	WorkerLoss
)

// Polarity says whether an entry helps or hurts. One engine rolls both tables: a
// malus is a boon with a negative sign, drawn from a separate catalog. The zero
// value is Positive.
type Polarity int

const (
	// Positive is a boon: the player gains something.
	Positive Polarity = iota
	// Negative is a malus: the player loses something (a negative-magnitude
	// timed effect, a resource drain or dead workers).
	Negative
)

// String returns a stable identifier for a Polarity.
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

// String returns a stable identifier for a Kind, free of flavor text. It is for
// logs and tests.
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

// TargetRule tells RollBoon how to pick a Def's target resource.
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
	// TargetRandomRare targets a random resource from rareResourceKeys that the
	// Profile's age has unlocked (wood if there is none).
	TargetRandomRare
)

// Boon is a rolled reward. It holds everything Apply needs, so a consumer can log
// it and Apply can dispatch it without looking at the catalog again.
//
// Field usage by Kind:
//   - RateBuff / AllProduction / TickSpeed: Magnitude (fraction) and DurationTicks.
//     Resource is set for RateBuff and empty otherwise.
//   - InstantResource: InstantAmount (lump) and Resource. DurationTicks is 0.
//   - TempWorkers: InstantAmount is the worker count (rounded on apply) and
//     DurationTicks the loan length. Resource is empty and Magnitude is 0.
//   - ResourceDrain (malus): InstantAmount is the drain fraction of current stock,
//     with Resource. Magnitude and DurationTicks optionally carry a short
//     production dip (both zero means no dip).
//   - WorkerLoss (malus): InstantAmount is the worker count lost. No duration.
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

// Def is a catalog entry: the tunable data for one kind of boon and the ranges
// RollBoon draws from. The values are the Def literals in catalog.go, so tuning
// is a data edit.
//
// Which range fields apply depends on Kind:
//   - RateBuff / AllProduction / TickSpeed: [MagMin,MagMax] and [DurMin,DurMax].
//     A malus entry uses a negative magnitude range (MagMin <= MagMax still
//     holds, e.g. MagMin -0.15, MagMax -0.08).
//   - InstantResource: [AmountMin,AmountMax]. No duration.
//   - TempWorkers: [AmountMin,AmountMax] (worker count) and [DurMin,DurMax].
//   - ResourceDrain: [AmountMin,AmountMax] as a fraction of current stock
//     (0.05 = 5%), plus an optional production dip from [MagMin,MagMax] and
//     [DurMin,DurMax]. Leave the dip ranges zero for a pure drain.
//   - WorkerLoss: [AmountMin,AmountMax] (worker count). No duration.
type Def struct {
	Kind      Kind
	Polarity  Polarity // Positive (default) or Negative; must match the catalog it lives in
	Name      string
	MagMin    float64    // lowest magnitude, as a fraction (negative for a malus)
	MagMax    float64    // highest magnitude
	DurMin    int        // shortest duration in ticks (timed boons)
	DurMax    int        // longest duration in ticks
	AmountMin float64    // lowest lump, worker count or drain fraction
	AmountMax float64    // highest lump, worker count or drain fraction
	Weight    int        // base rarity weight (higher = more common)
	Target    TargetRule // how to pick the target resource
	Resource  string     // used only when Target == TargetSpecificResource
	Flavors   []string   // flavor-template pool; placeholders are {pct} {res} {ticks} {amt} {n} {frac}
}

// Profile is the caller's roll-shaping. It biases the table without this package
// knowing why. A faction system fills Specialty and Age from the faction and sets
// WeightMult, MagnitudeScale and RarityScale from its character and your standing.
type Profile struct {
	// Specialty is the resource key a TargetSpecialty boon aims at. Empty falls
	// back to the first age-appropriate resource.
	Specialty string
	// Age is the current age key. It bounds the age-appropriate resource pools,
	// sets the instant-lump growth factor (see instantScale) and stretches timed
	// durations. Empty or unknown means every resource is unlocked, with the age
	// index at 0 (the smallest lump scaling).
	Age string
	// Polarity selects the table RollBoon draws from: Positive draws Catalog
	// (the boons), Negative draws MalusCatalog (the setbacks).
	Polarity Polarity
	// Enabled is a denylist: a Kind rolls unless Enabled[k] is false. A nil map
	// leaves every kind live.
	Enabled map[Kind]bool
	// WeightMult multiplies a Kind's base weight. A missing entry or nil map
	// means 1.0.
	WeightMult map[Kind]float64
	// MagnitudeScale scales rolled percentage magnitudes, the InstantResource
	// lump (on top of the age-growth factor) and the ResourceDrain fraction.
	// Worker counts (TempWorkers, WorkerLoss) are not scaled: a scaled count is a
	// cliff, not a curve. A value <= 0 is treated as 1.0.
	MagnitudeScale float64
	// RarityScale biases toward rarer entries. Effective weight is
	// baseWeight^(1/RarityScale). Above 1 the weight spread shrinks, so rare
	// (low-weight) boons get more of the probability. Below 1 common boons
	// dominate. A value <= 0 is treated as 1.0.
	RarityScale float64
	// TickInterval is the length of one tick, used to word durations as
	// wall-clock time in flavor lines. A value <= 0 means DefaultTickInterval.
	TickInterval time.Duration
}

// DefaultTickInterval is the game's base tick length (2s at 1x speed), used when
// Profile.TickInterval is unset.
const DefaultTickInterval = 2 * time.Second

// tickInterval returns the profile's tick length, or DefaultTickInterval if it is
// not positive.
func (p Profile) tickInterval() time.Duration {
	if p.TickInterval <= 0 {
		return DefaultTickInterval
	}
	return p.TickInterval
}

// DefaultProfile returns a neutral profile: every kind enabled, unit weights and
// unit scales. Callers set Specialty and Age and adjust the rest.
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

// kindWeightMult returns the per-kind weight multiplier (1.0 if unset).
func (p Profile) kindWeightMult(k Kind) float64 {
	if p.WeightMult == nil {
		return 1.0
	}
	if m, ok := p.WeightMult[k]; ok {
		return m
	}
	return 1.0
}

// magScale returns the magnitude scale, or 1.0 if it is not positive.
func (p Profile) magScale() float64 {
	if p.MagnitudeScale <= 0 {
		return 1.0
	}
	return p.MagnitudeScale
}

// rarityScale returns the rarity scale, or 1.0 if it is not positive.
func (p Profile) rarityScale() float64 {
	if p.RarityScale <= 0 {
		return 1.0
	}
	return p.RarityScale
}

// effectiveWeight is a Def's weight for the weighted pick under this profile: the
// base weight raised to 1/RarityScale, times the per-kind multiplier. It is 0 for
// a non-positive base weight.
func (p Profile) effectiveWeight(d Def) float64 {
	if d.Weight <= 0 {
		return 0
	}
	return detmath.Pow(float64(d.Weight), 1.0/p.rarityScale()) * p.kindWeightMult(d.Kind)
}
