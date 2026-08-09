package boon

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// rareResourceKeys is the curated pool a Grand Cache draws from, intersected
// with what the current age has unlocked. Ordered for deterministic selection;
// tunable data like everything else.
var rareResourceKeys = []string{
	"knowledge", "gold", "uranium", "plasma", "titanium",
	"dark_matter", "antimatter", "quantum_flux", "crypto", "data",
}

// RollBoon weighted-picks a Def honoring the profile (enabled kinds, weight
// multipliers, rarity/magnitude scale), then rolls a concrete magnitude,
// duration, target, and flavor line. It is PURE given (profile, rng): every
// random draw comes from the passed *rand.Rand in a fixed order, so the same
// seed reproduces the same boon sequence. It never uses package-level rand.
//
// Profile.Polarity selects the TABLE: Positive draws Catalog(), Negative draws
// MalusCatalog(). Everything downstream (weighting, ranges, targeting, flavor)
// is identical — a malus is a boon with the sign flipped, which is exactly why
// one engine can roll both.
//
// If the profile disables every catalogued kind, RollBoon returns the zero Boon
// rather than panicking — callers are expected to leave at least one kind live.
func RollBoon(profile Profile, rng *rand.Rand) Boon {
	// Build the enabled candidate list from the catalog's fixed order, so the
	// weighted pick is deterministic given rng's state.
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

	// Weighted pick. rng.Float64() ∈ [0,1) so roll < total always; the trailing
	// assignment is belt-and-suspenders against float rounding.
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
		// Age-scaled AND magnitude-scaled: a fixed lump is a fortune in the
		// primitive age and invisible in the quantum age, and standing/strength
		// should be legible in the size of a gift.
		b.InstantAmount = rollFloatRange(rng, pick.AmountMin, pick.AmountMax) * profile.instantScale()
	case TempWorkers, WorkerLoss:
		// Discrete head-counts: never scaled, only ranged.
		b.InstantAmount = float64(rollIntRange(rng, int(pick.AmountMin), int(pick.AmountMax)))
		b.DurationTicks = rollIntRange(rng, pick.DurMin, pick.DurMax)
	case ResourceDrain:
		// InstantAmount carries the drain FRACTION of current stock; the
		// optional Mag/Dur pair carries a short production dip alongside it.
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
	b.Resource = resolveTarget(pick, profile, rng)
	b.Flavor = rollFlavor(pick, b, rng)
	return b
}

// tableFor returns the catalog a polarity draws from.
func tableFor(p Polarity) []Def {
	if p == Negative {
		return MalusCatalog()
	}
	return Catalog()
}

// ageIndex is a profile age's position in the canonical age order. Empty or
// unknown ages resolve to 0 — the most conservative (smallest) instant scaling,
// which is the safe direction to be wrong in.
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

// instantScale is the multiplier applied to an InstantResource lump: the per-age
// growth factor compounded over the profile's age position, times the profile's
// MagnitudeScale so standing and strength show up in the size of a gift.
func (p Profile) instantScale() float64 {
	return math.Pow(instantGrowthPerAge, float64(ageIndex(p.Age))) * p.magScale()
}

// resolveTarget picks the concrete target resource for a Def under a profile.
// Only the random rules consume rng, and each rule's draw count is fixed, so
// the overall roll sequence stays deterministic.
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

// ageAppropriateResources returns resource keys unlocked at or before age that
// are also STORABLE. An empty or unknown age yields everything unlocked — a
// permissive default so a caller that omits Age still gets a usable pool.
// Deterministic: config slices are ordered and the index map is used only for
// lookups.
//
// The BaseStorage > 0 filter is not cosmetic. A resource with no storage (today
// that is "soldiers", which is spent rather than stockpiled) would silently
// swallow a grant — ResourceManager.Add clamps to Storage, so the lump lands as
// zero — and a "<res>_rate" buff on it is equally invisible. Excluding them
// keeps the random target pools free of no-ops.
func ageAppropriateResources(age string) []string {
	order := config.AgeOrder()
	idx := make(map[string]int, len(order))
	for i, k := range order {
		idx[k] = i
	}
	cur, ok := idx[age]
	if !ok {
		cur = len(order) // unknown/empty age ⇒ treat everything as unlocked
	}
	var out []string
	for _, r := range config.BaseResources() {
		if r.BaseStorage <= 0 {
			continue // unstorable: a grant or rate buff on it is a no-op
		}
		ri, ok := idx[r.Age]
		if !ok || ri <= cur {
			out = append(out, r.Key)
		}
	}
	return out
}

// rareAgeResources returns the rare pool intersected with what age has unlocked,
// falling back to knowledge when nothing rare is available yet.
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
		return []string{"knowledge"}
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
	return lo + rng.Float64()*(hi-lo)
}

// rollFlavor picks a template from the Def's pool and fills its placeholders
// from the rolled boon. Every catalog entry ships at least one template, so the
// result is always non-empty (falling back to the Def name if a pool is empty).
//
// Placeholders:
//
//	{pct}   Magnitude as a percentage, ABSOLUTE — the template carries the
//	        direction in its wording ("+{pct}" for a boon, "falls {pct}" for a
//	        malus), so a bare "-12%" never leaks into a sentence that already
//	        said "falls".
//	{res}   prettified target resource key
//	{ticks} duration in ticks
//	{amt}   InstantAmount rounded (lump grants)
//	{n}     InstantAmount rounded (head-counts: workers gained or lost)
//	{frac}  InstantAmount as a percentage (ResourceDrain's drain fraction)
func rollFlavor(d Def, b Boon, rng *rand.Rand) string {
	if len(d.Flavors) == 0 {
		return d.Name
	}
	tmpl := d.Flavors[rng.Intn(len(d.Flavors))]
	amount := fmt.Sprintf("%d", int(math.Round(b.InstantAmount)))
	rep := strings.NewReplacer(
		"{pct}", fmt.Sprintf("%d%%", int(math.Round(math.Abs(b.Magnitude)*100))),
		"{res}", resourceLabel(b.Resource),
		"{ticks}", fmt.Sprintf("%d", b.DurationTicks),
		"{amt}", amount,
		"{n}", amount,
		"{frac}", fmt.Sprintf("%d%%", int(math.Round(math.Abs(b.InstantAmount)*100))),
	)
	return rep.Replace(tmpl)
}

// resourceLabel prettifies a resource key for flavor text (quantum_flux →
// "quantum flux"); "" becomes a generic "resources".
func resourceLabel(key string) string {
	if key == "" {
		return "resources"
	}
	return strings.ReplaceAll(key, "_", " ")
}
