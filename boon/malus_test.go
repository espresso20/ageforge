package boon

import (
	"math"
	"math/rand"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// malus_test.go covers the negative half of the engine (Polarity, MalusCatalog,
// the two malus-only Kinds and their Applier paths) plus the two roll-side
// robustness fixes that landed with it: no-op target exclusion and age-scaled
// instant lumps.

// --- polarity / table selection --------------------------------------------

// TestCatalogPolaritiesAreConsistent pins the invariant the table split rests
// on: every Catalog() entry is Positive and every MalusCatalog() entry is
// Negative. A Def in the wrong table would silently hand a gift to a war party.
func TestCatalogPolaritiesAreConsistent(t *testing.T) {
	for _, d := range Catalog() {
		if d.Polarity != Positive {
			t.Errorf("Catalog entry %q has Polarity %v, want Positive", d.Name, d.Polarity)
		}
	}
	for _, d := range MalusCatalog() {
		if d.Polarity != Negative {
			t.Errorf("MalusCatalog entry %q has Polarity %v, want Negative", d.Name, d.Polarity)
		}
	}
	if len(MalusCatalog()) == 0 {
		t.Fatal("MalusCatalog is empty")
	}
}

// TestNegativeProfileDrawsOnlyFromMalusCatalog proves Profile.Polarity really is
// the table selector: a Negative profile never yields a boon name, and over many
// rolls it reaches every malus entry.
func TestNegativeProfileDrawsOnlyFromMalusCatalog(t *testing.T) {
	malus := map[string]bool{}
	for _, d := range MalusCatalog() {
		malus[d.Name] = true
	}

	p := DefaultProfile()
	p.Polarity = Negative
	p.Age = "space_age"
	p.Specialty = "iron"

	seen := map[string]bool{}
	for i, b := range rollN(p, 4242, 3000) {
		if !malus[b.Name] {
			t.Fatalf("roll %d drew %q, which is not on the malus table", i, b.Name)
		}
		if b.Polarity != Negative {
			t.Fatalf("roll %d (%s) has Polarity %v, want Negative", i, b.Name, b.Polarity)
		}
		if b.Flavor == "" {
			t.Fatalf("roll %d (%s) has empty flavor", i, b.Name)
		}
		seen[b.Name] = true
	}
	for name := range malus {
		if !seen[name] {
			t.Errorf("malus %q never rolled over 3000 draws", name)
		}
	}
}

// TestMalusMagnitudesAreNegative checks the sign of every timed setback: a
// "buff" with a positive value would be a reward for losing.
func TestMalusMagnitudesAreNegative(t *testing.T) {
	p := DefaultProfile()
	p.Polarity = Negative
	p.Age = "atomic_age"

	for i, b := range rollN(p, 99, 2000) {
		switch b.Kind {
		case RateBuff, AllProduction, TickSpeed:
			if b.Magnitude >= 0 {
				t.Fatalf("roll %d (%s) is a timed malus with magnitude %.4f, want < 0", i, b.Name, b.Magnitude)
			}
		case ResourceDrain:
			if b.InstantAmount <= 0 || b.InstantAmount > maxDrainFraction+1e-9 {
				t.Fatalf("roll %d (%s) drain fraction %.4f outside (0,%.2f]", i, b.Name, b.InstantAmount, maxDrainFraction)
			}
			if b.DurationTicks > 0 && b.Magnitude >= 0 {
				t.Fatalf("roll %d (%s) carries a dip with magnitude %.4f, want < 0", i, b.Name, b.Magnitude)
			}
		case WorkerLoss:
			if b.InstantAmount < 1 {
				t.Fatalf("roll %d (%s) loses %.2f workers, want >= 1", i, b.Name, b.InstantAmount)
			}
		default:
			t.Fatalf("roll %d produced unexpected malus kind %v", i, b.Kind)
		}
	}
}

// TestMalusSeverityScalesWithMagnitudeScale: a harsher profile bites harder on
// both the timed magnitudes and the drain fraction. Monotonicity, not exact
// values.
func TestMalusSeverityScalesWithMagnitudeScale(t *testing.T) {
	mean := func(scale float64) (mag, drain float64) {
		p := DefaultProfile()
		p.Polarity = Negative
		p.Age = "atomic_age"
		p.MagnitudeScale = scale

		var magSum, magN, drainSum, drainN float64
		for _, b := range rollN(p, 7, 4000) {
			if b.Kind == RateBuff || b.Kind == AllProduction {
				magSum += math.Abs(b.Magnitude)
				magN++
			}
			if b.Kind == ResourceDrain {
				drainSum += b.InstantAmount
				drainN++
			}
		}
		return magSum / magN, drainSum / drainN
	}

	mildMag, mildDrain := mean(0.80)
	harshMag, harshDrain := mean(1.40)

	if harshMag <= mildMag {
		t.Errorf("timed malus magnitude did not scale: harsh %.4f <= mild %.4f", harshMag, mildMag)
	}
	if harshDrain <= mildDrain {
		t.Errorf("drain fraction did not scale: harsh %.4f <= mild %.4f", harshDrain, mildDrain)
	}
}

// TestDrainFractionIsClamped: even an absurd MagnitudeScale cannot make a
// setback take the whole stockpile.
func TestDrainFractionIsClamped(t *testing.T) {
	p := DefaultProfile()
	p.Polarity = Negative
	p.Age = "atomic_age"
	p.MagnitudeScale = 50 // pathological

	for i, b := range rollN(p, 3, 1500) {
		if b.Kind == ResourceDrain && b.InstantAmount > maxDrainFraction+1e-9 {
			t.Fatalf("roll %d drain fraction %.4f exceeds clamp %.2f", i, b.InstantAmount, maxDrainFraction)
		}
	}
}

// --- apply dispatch: malus kinds -------------------------------------------

func TestApplyResourceDrainRoutesToDrain(t *testing.T) {
	f := &fakeApplier{}
	// Pure drain (Spoiled Supplies shape): no dip, exactly one call.
	Apply(Boon{Kind: ResourceDrain, Polarity: Negative, Resource: "food", InstantAmount: 0.12, Name: "Spoiled Supplies", Flavor: "f"}, f)
	if f.calls() != 1 || len(f.drained) != 1 {
		t.Fatalf("pure ResourceDrain routed wrong: %+v", f)
	}
	if f.drained[0].resource != "food" || math.Abs(f.drained[0].fraction-0.12) > 1e-9 {
		t.Fatalf("ResourceDrain drained %+v, want food/0.12", f.drained[0])
	}
}

func TestApplyResourceDrainWithDipAlsoInjectsNegativeEffect(t *testing.T) {
	f := &fakeApplier{}
	// Lost Scouts shape: drain + a short NEGATIVE production dip.
	Apply(Boon{
		Kind: ResourceDrain, Polarity: Negative, Resource: "wood",
		InstantAmount: 0.04, Magnitude: -0.05, DurationTicks: 900,
		Name: "Lost Scouts", Flavor: "f",
	}, f)

	if len(f.drained) != 1 || len(f.injected) != 1 || f.calls() != 2 {
		t.Fatalf("compound ResourceDrain routed wrong: %+v", f)
	}
	eff := f.injected[0].effects
	if len(eff) != 1 || eff[0].Type != "production_all" {
		t.Fatalf("dip mapped to %+v, want a production_all effect", eff)
	}
	if eff[0].Value >= 0 {
		t.Fatalf("dip value %.4f is not negative — a malus must debuff", eff[0].Value)
	}
	if f.injected[0].ticks != 900 {
		t.Fatalf("dip ticks = %d, want 900", f.injected[0].ticks)
	}
}

func TestApplyWorkerLossRoutesToLoseWorkers(t *testing.T) {
	f := &fakeApplier{}
	Apply(Boon{Kind: WorkerLoss, Polarity: Negative, InstantAmount: 2.6, Name: "Dysentery", Flavor: "f"}, f)
	if f.calls() != 1 || len(f.lost) != 1 {
		t.Fatalf("WorkerLoss routed wrong: %+v", f)
	}
	if f.lost[0] != 3 { // 2.6 rounds to 3
		t.Fatalf("WorkerLoss lost %d workers, want 3 (rounded)", f.lost[0])
	}
}

// TestApplyMalusKindsIgnoreEmptyPayloads: a zero drain or a zero head-count must
// not reach the Applier at all — nothing happened, so nothing should be called.
func TestApplyMalusKindsIgnoreEmptyPayloads(t *testing.T) {
	f := &fakeApplier{}
	Apply(Boon{Kind: ResourceDrain, Resource: "", InstantAmount: 0, Flavor: "f"}, f)
	Apply(Boon{Kind: WorkerLoss, InstantAmount: 0, Flavor: "f"}, f)
	if f.calls() != 0 {
		t.Fatalf("empty malus payloads still called the Applier: %+v", f)
	}
}

// TestRollThenApplyMalusEndToEnd: every rolled malus applies through the fake
// and produces at least one real side effect.
func TestRollThenApplyMalusEndToEnd(t *testing.T) {
	p := DefaultProfile()
	p.Polarity = Negative
	p.Age = "space_age"
	p.Specialty = "iron"

	rng := rand.New(rand.NewSource(2026))
	for i := 0; i < 500; i++ {
		b := RollBoon(p, rng)
		f := &fakeApplier{}
		if line := Apply(b, f); line == "" {
			t.Fatalf("malus %d applied with empty flavor", i)
		}
		if f.calls() == 0 {
			t.Fatalf("malus %d (%s) made no Applier calls — the setback did nothing", i, b.Name)
		}
	}
}

// --- A2: random targets are never no-ops ------------------------------------

// TestRandomTargetsAreStorable guards the fix for silently vanishing rewards: a
// resource with BaseStorage <= 0 (today "soldiers") can hold neither a lump grant
// nor a meaningful rate buff, so it must never surface as a random target.
func TestRandomTargetsAreStorable(t *testing.T) {
	storage := map[string]float64{}
	for _, r := range config.BaseResources() {
		storage[r.Key] = r.BaseStorage
	}

	// Sanity: the fixture is only meaningful if such a resource actually exists.
	unstorable := 0
	for _, r := range config.BaseResources() {
		if r.BaseStorage <= 0 {
			unstorable++
		}
	}
	if unstorable == 0 {
		t.Skip("no unstorable resources in config — nothing to exclude")
	}

	for _, polarity := range []Polarity{Positive, Negative} {
		for _, age := range config.AgeOrder() {
			p := DefaultProfile()
			p.Polarity = polarity
			p.Age = age
			// No specialty: TargetSpecialty also falls back to the age pool.
			for i, b := range rollN(p, 555, 400) {
				if b.Resource == "" {
					continue
				}
				if s, ok := storage[b.Resource]; ok && s <= 0 {
					t.Fatalf("age %s polarity %v roll %d targeted unstorable resource %q "+
						"(BaseStorage %.0f — the effect would silently vanish)",
						age, polarity, i, b.Resource, s)
				}
			}
		}
	}
}

// TestAgeAppropriateResourcesNeverEmpty: filtering must not be able to starve the
// pool, at any age.
func TestAgeAppropriateResourcesNeverEmpty(t *testing.T) {
	for _, age := range append(config.AgeOrder(), "", "not_an_age") {
		if len(ageAppropriateResources(age)) == 0 {
			t.Errorf("ageAppropriateResources(%q) is empty", age)
		}
		if len(rareAgeResources(age)) == 0 {
			t.Errorf("rareAgeResources(%q) is empty", age)
		}
	}
}

// --- A3: instant lumps scale with age and standing --------------------------

// meanInstant rolls an InstantResource-only profile and returns the mean lump.
func meanInstant(t *testing.T, age string, scale float64) float64 {
	t.Helper()
	p := DefaultProfile()
	p.Age = age
	p.MagnitudeScale = scale
	p.Enabled = map[Kind]bool{
		RateBuff: false, AllProduction: false, TickSpeed: false, TempWorkers: false,
	}
	var sum float64
	const n = 2000
	for _, b := range rollN(p, 8080, n) {
		if b.Kind != InstantResource {
			t.Fatalf("kind filter leaked: got %v", b.Kind)
		}
		sum += b.InstantAmount
	}
	return sum / n
}

// TestInstantLumpsScaleWithAge asserts monotonic growth across the whole age
// order — a fixed 50–200 crate must not stay a rounding error into the late game.
func TestInstantLumpsScaleWithAge(t *testing.T) {
	order := config.AgeOrder()
	prev := meanInstant(t, order[0], 1.0)
	for _, age := range order[1:] {
		got := meanInstant(t, age, 1.0)
		if got <= prev {
			t.Fatalf("instant lump did not grow into %s: %.4f <= previous %.4f", age, got, prev)
		}
		prev = got
	}
}

// TestInstantLumpsScaleWithMagnitude asserts standing/strength is visible in the
// size of a gift — the second half of the A3 fix. A late-age allied str-5 profile
// must dwarf an early-age neutral one.
func TestInstantLumpsScaleWithMagnitude(t *testing.T) {
	const bronze = "bronze_age"
	neutral := meanInstant(t, bronze, 1.0)
	generous := meanInstant(t, bronze, 1.82) // ~allied (x1.40) x str-5 (x1.30)
	if generous <= neutral {
		t.Fatalf("MagnitudeScale did not affect instant lumps: %.4f <= %.4f", generous, neutral)
	}

	lateAllied := meanInstant(t, "quantum_age", 1.82)
	if lateAllied <= neutral*1000 {
		t.Fatalf("late-age allied lump %.2f is not decisively larger than a bronze-age neutral lump %.2f",
			lateAllied, neutral)
	}
}

// TestAgeIndexGuardsUnknownAges: an unrecognised or empty age resolves to index 0,
// the smallest (safest) scaling, rather than panicking or over-granting.
func TestAgeIndexGuardsUnknownAges(t *testing.T) {
	for _, age := range []string{"", "not_an_age", "AGE_OF_TYPOS"} {
		if got := ageIndex(age); got != 0 {
			t.Errorf("ageIndex(%q) = %d, want 0", age, got)
		}
	}
	if got := ageIndex(config.AgeOrder()[0]); got != 0 {
		t.Errorf("ageIndex(first age) = %d, want 0", got)
	}
}
