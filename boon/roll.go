package boon

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// rareResourceKeys is the pool a Grand Cache draws from, limited to what the
// current age has unlocked. The order is fixed so the pick is deterministic.
var rareResourceKeys = []string{
	"gold", "uranium", "plasma", "titanium",
	"dark_matter", "antimatter", "quantum_flux", "crypto", "data",
}

// RollBoon picks a Def by weight under the profile (enabled kinds, weight
// multipliers, rarity and magnitude scale), then rolls its magnitude, duration,
// target and flavor line. It depends only on (profile, rng): every random draw
// comes from the passed *rand.Rand in a fixed order, so the same seed gives the
// same boon sequence. It never uses package-level rand.
//
// Profile.Polarity selects the table: Positive draws Catalog, Negative draws
// MalusCatalog. Weighting, ranges, targeting and flavor work the same for both.
//
// If the profile disables every kind in the table, RollBoon returns the zero
// Boon. Callers should leave at least one kind live.
func RollBoon(profile Profile, rng *rand.Rand) Boon {
	// Candidates keep the catalog's fixed order, so the weighted pick is
	// deterministic given rng's state.
	cat := tableFor(profile.Polarity)
	cands := make([]Def, 0, len(cat))
	weights := make([]float64, 0, len(cat))
	var total float64
	for _, d := range cat {
		if !profile.kindEnabled(d.Kind) {
			continue
		}
		w := profile.effectiveWeight(d)
		if w <= 0 {
			continue
		}
		cands = append(cands, d)
		weights = append(weights, w)
		total += w
	}
	if len(cands) == 0 || total <= 0 {
		return Boon{}
	}

	// Weighted pick. rng.Float64() is in [0,1), so roll < total. Starting from the
	// last candidate covers float rounding in the running sum.
	roll := rng.Float64() * total
	pick := cands[len(cands)-1]
	cum := 0.0
	for i, w := range weights {
		cum += w
		if roll < cum {
			pick = cands[i]
			break
		}
	}

	b := Boon{Kind: pick.Kind, Polarity: pick.Polarity, Name: pick.Name}
	switch pick.Kind {
	case InstantResource:
		// Scaled by age and by magnitude: a fixed lump is a fortune in the
		// primitive age and invisible in the quantum age, and standing and
		// strength should show in the size of a gift.
		b.InstantAmount = rollFloatRange(rng, pick.AmountMin, pick.AmountMax) * profile.instantScale()
	case TempWorkers, WorkerLoss:
		// Head-counts are discrete: ranged, never scaled.
		b.InstantAmount = float64(rollIntRange(rng, int(pick.AmountMin), int(pick.AmountMax)))
		b.DurationTicks = rollIntRange(rng, pick.DurMin, pick.DurMax)
	case ResourceDrain:
		// InstantAmount is the drain fraction of current stock. The optional
		// Mag and Dur pair is a short production dip.
		frac := rollFloatRange(rng, pick.AmountMin, pick.AmountMax) * profile.magScale()
		b.InstantAmount = math.Min(math.Max(frac, 0), maxDrainFraction)
		if pick.MagMin != 0 || pick.MagMax != 0 {
			b.Magnitude = rollFloatRange(rng, pick.MagMin, pick.MagMax) * profile.magScale()
			b.DurationTicks = rollIntRange(rng, pick.DurMin, pick.DurMax)
		}
	default: // RateBuff, AllProduction, TickSpeed
		b.Magnitude = rollFloatRange(rng, pick.MagMin, pick.MagMax) * profile.magScale()
		b.DurationTicks = rollIntRange(rng, pick.DurMin, pick.DurMax)
	}
	// The table's durations are typed for the base curve. From the Bronze Age on
	// the ages run config.PacingStretch times longer, and so do timed boons and
	// setbacks, so each covers the same share of an age. Expeditions, which bring
	// the encounters, stretch the same way, so the slot load described in
	// catalog.go does not change.
	b.DurationTicks = config.StretchTicks(profile.Age, b.DurationTicks)
	b.Resource = resolveTarget(pick, profile, rng)
	b.Flavor = rollFlavor(pick, b, rng, profile.tickInterval())
	return b
}

// tableFor returns the catalog that polarity p draws from.
func tableFor(p Polarity) []Def {
	if p == Negative {
		return MalusCatalog()
	}
	return Catalog()
}

// ageIndex is a profile age's position in config.AgeOrder. An empty or unknown
// age gives 0, the smallest instant-lump scaling.
func ageIndex(age string) int {
	if age == "" {
		return 0
	}
	for i, k := range config.AgeOrder() {
		if k == age {
			return i
		}
	}
	return 0
}

// instantScale is the multiplier on an InstantResource lump: instantGrowthPerAge
// raised to the profile's age index, times MagnitudeScale (so standing and
// strength show in the size of a gift), divided by the age's pacing stretch. A
// lump is tuned to a number of ticks of production, and from the Bronze Age on a
// tick makes 1/config.PacingStretch as much.
func (p Profile) instantScale() float64 {
	return float64(detmath.Pow(instantGrowthPerAge, float64(ageIndex(p.Age)))*p.magScale()) / config.AgeStretch(p.Age)
}

// resolveTarget picks the target resource for a Def under a profile. Only the
// random rules draw from rng, and each draws a fixed number of values, so the
// roll sequence stays deterministic.
func resolveTarget(d Def, p Profile, rng *rand.Rand) string {
	switch d.Target {
	case TargetSpecialty:
		if p.Specialty != "" {
			return p.Specialty
		}
		if pool := ageAppropriateResources(p.Age); len(pool) > 0 {
			return pool[0]
		}
		return "food"
	case TargetSpecificResource:
		return d.Resource
	case TargetRandomAge:
		pool := ageAppropriateResources(p.Age)
		if len(pool) == 0 {
			return "food"
		}
		return pool[rng.Intn(len(pool))]
	case TargetRandomRare:
		pool := rareAgeResources(p.Age)
		return pool[rng.Intn(len(pool))]
	default: // TargetNone
		return ""
	}
}

// ageAppropriateResources returns the keys of resources unlocked at or before age
// that have a BaseStorage above zero (ones a town can save up). An empty or
// unknown age gives every such resource, so a caller that omits Age still gets a
// usable pool. No resource has BaseStorage <= 0 today, so the filter excludes
// nothing. The order follows config.BaseResources, and the index map is used only
// for lookups, so the result is deterministic.
func ageAppropriateResources(age string) []string {
	order := config.AgeOrder()
	idx := make(map[string]int, len(order))
	for i, k := range order {
		idx[k] = i
	}
	cur, ok := idx[age]
	if !ok {
		cur = len(order) // unknown or empty age: everything is unlocked
	}
	var out []string
	for _, r := range config.BaseResources() {
		if r.BaseStorage <= 0 {
			continue // not a resource a town can save up
		}
		ri, ok := idx[r.Age]
		if !ok || ri <= cur {
			out = append(out, r.Key)
		}
	}
	return out
}

// rareAgeResources returns the rare pool limited to what age has unlocked, or
// wood if none of it is unlocked yet.
func rareAgeResources(age string) []string {
	avail := ageAppropriateResources(age)
	set := make(map[string]bool, len(avail))
	for _, k := range avail {
		set[k] = true
	}
	out := make([]string, 0, len(rareResourceKeys))
	for _, k := range rareResourceKeys {
		if set[k] {
			out = append(out, k)
		}
	}
	if len(out) == 0 {
		return []string{"wood"}
	}
	return out
}

// rollIntRange returns a value in [lo, hi] inclusive (lo if the range is empty).
func rollIntRange(rng *rand.Rand, lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + rng.Intn(hi-lo+1)
}

// rollFloatRange returns a value in [lo, hi) (lo if the range is empty).
func rollFloatRange(rng *rand.Rand, lo, hi float64) float64 {
	if hi <= lo {
		return lo
	}
	return lo + float64(rng.Float64()*(hi-lo))
}

// rollFlavor picks a template from the Def's pool and fills in its placeholders
// from the rolled boon. It returns the Def's name if the pool is empty.
//
// Placeholders:
//
//	{pct}   Magnitude as a percentage, absolute value. The template carries the
//	        direction in its wording ("+{pct}" for a boon, "falls {pct}" for a
//	        malus), so a bare "-12%" never lands in a sentence that already
//	        said "falls".
//	{res}   the target resource key, with underscores as spaces
//	{ticks} duration as wall-clock time at the profile's tick interval ("~50m");
//	        templates read "for {ticks}"
//	{amt}   InstantAmount rounded, K/M formatted (lump grants)
//	{n}     InstantAmount rounded, K/M formatted (head-counts)
//	{frac}  InstantAmount as a percentage (ResourceDrain's drain fraction)
func rollFlavor(d Def, b Boon, rng *rand.Rand, interval time.Duration) string {
	if len(d.Flavors) == 0 {
		return d.Name
	}
	tmpl := d.Flavors[rng.Intn(len(d.Flavors))]
	amount := textfmt.Number(math.Round(b.InstantAmount))
	rep := strings.NewReplacer(
		"{pct}", fmt.Sprintf("%d%%", int(math.Round(math.Abs(b.Magnitude)*100))),
		"{res}", resourceLabel(b.Resource),
		"{ticks}", textfmt.Ticks(b.DurationTicks, interval),
		"{amt}", amount,
		"{n}", amount,
		"{frac}", drainPercent(b),
	)
	return rep.Replace(tmpl)
}

// drainPercent formats a ResourceDrain's fraction as a percentage ("12%").
func drainPercent(b Boon) string {
	return fmt.Sprintf("%d%%", int(math.Round(math.Abs(b.InstantAmount)*100)))
}

// WithoutDip returns b with its production dip removed (Magnitude and
// DurationTicks zeroed). The game uses it when the malus capacity is full. A
// ResourceDrain that carried a dip gets a drain-only line, so the text never
// announces a slowdown that will not happen: "4% of the iron they carried is
// gone."
func (b Boon) WithoutDip() Boon {
	hadDip := b.Magnitude != 0 || b.DurationTicks != 0
	b.Magnitude, b.DurationTicks = 0, 0
	if b.Kind == ResourceDrain && hadDip {
		b.Flavor = fmt.Sprintf("%s of the %s they carried is gone.", drainPercent(b), resourceLabel(b.Resource))
	}
	return b
}

// resourceLabel formats a resource key for flavor text (quantum_flux becomes
// "quantum flux"). An empty key becomes "resources".
func resourceLabel(key string) string {
	if key == "" {
		return "resources"
	}
	return strings.ReplaceAll(key, "_", " ")
}
