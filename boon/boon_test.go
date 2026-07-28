package boon

import (
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// --- helpers ---------------------------------------------------------------

func rollN(p Profile, seed int64, n int) []Boon {
	rng := rand.New(rand.NewSource(seed))
	out := make([]Boon, n)
	for i := range out {
		out[i] = RollBoon(p, rng)
	}
	return out
}

func defsByName() map[string]Def {
	m := make(map[string]Def)
	for _, d := range Catalog() {
		m[d.Name] = d
	}
	return m
}

// fakeApplier records every call so tests can assert dispatch and mapping in
// complete isolation from the game engine.
type fakeApplier struct {
	injected []struct {
		effects []config.Effect
		ticks   int
		name    string
	}
	granted []struct {
		resource string
		amount   float64
	}
	workers []struct {
		count int
		ticks int
	}
	drained []struct {
		resource string
		fraction float64
	}
	lost []int
}

func (f *fakeApplier) InjectTimedEffects(effects []config.Effect, ticks int, name string) {
	f.injected = append(f.injected, struct {
		effects []config.Effect
		ticks   int
		name    string
	}{effects, ticks, name})
}

func (f *fakeApplier) GrantResource(resource string, amount float64) {
	f.granted = append(f.granted, struct {
		resource string
		amount   float64
	}{resource, amount})
}

func (f *fakeApplier) GrantTempWorkers(count, ticks int) {
	f.workers = append(f.workers, struct {
		count int
		ticks int
	}{count, ticks})
}

func (f *fakeApplier) DrainResource(resource string, fraction float64) {
	f.drained = append(f.drained, struct {
		resource string
		fraction float64
	}{resource, fraction})
}

func (f *fakeApplier) LoseWorkers(count int) { f.lost = append(f.lost, count) }

// calls is the total number of Applier calls recorded.
func (f *fakeApplier) calls() int {
	return len(f.injected) + len(f.granted) + len(f.workers) + len(f.drained) + len(f.lost)
}

// --- determinism -----------------------------------------------------------

func TestRollBoonDeterministic(t *testing.T) {
	p := DefaultProfile()
	p.Specialty = "iron"
	p.Age = "iron_age"

	a := rollN(p, 20260726, 300)
	b := rollN(p, 20260726, 300)

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed + profile produced different boon sequences")
	}

	// A different seed should (essentially always) diverge, proving the stream
	// actually tracks the rng rather than being constant.
	c := rollN(p, 99999, 300)
	if reflect.DeepEqual(a, c) {
		t.Fatalf("different seeds produced identical sequences — rng not driving rolls")
	}
}

// --- weighting: disabled kind never rolls ----------------------------------

func TestDisabledKindNeverRolls(t *testing.T) {
	p := DefaultProfile()
	p.Age = "iron_age"
	p.Specialty = "iron"
	p.Enabled = map[Kind]bool{RateBuff: false}

	for i, b := range rollN(p, 7, 2000) {
		if b.Kind == RateBuff {
			t.Fatalf("roll %d produced a disabled RateBuff boon: %+v", i, b)
		}
	}
}

// --- weighting: up-weighted kind dominates ---------------------------------

func TestUpweightedKindDominates(t *testing.T) {
	p := DefaultProfile()
	p.Age = "iron_age"
	p.WeightMult = map[Kind]float64{TickSpeed: 100}

	const n = 2000
	tick := 0
	for _, b := range rollN(p, 42, n) {
		if b.Kind == TickSpeed {
			tick++
		}
	}
	if tick < n*4/5 { // expect ~90%+; assert a safe > 80%
		t.Fatalf("up-weighted TickSpeed only won %d/%d rolls; expected dominance", tick, n)
	}
}

// --- ranges: rolled values respect the Def bounds (× MagnitudeScale) --------

// UPDATED (boon capacity/malus pass): instant lump ranges are no longer raw Def
// values — they are multiplied by Profile.instantScale() (age growth ×
// MagnitudeScale), so the expected bounds scale with them. The assertion is
// otherwise unchanged: rolled values must sit inside the Def range, transformed.
func TestRolledValuesWithinDefRanges(t *testing.T) {
	const eps = 1e-9
	catalog := defsByName()

	for _, scale := range []float64{1.0, 1.5, 0.5} {
		p := DefaultProfile()
		p.Age = "space_age" // wide resource pool
		p.Specialty = "iron"
		p.MagnitudeScale = scale
		instScale := p.instantScale()

		for i, b := range rollN(p, 123, 3000) {
			d, ok := catalog[b.Name]
			if !ok {
				t.Fatalf("roll %d has unknown name %q", i, b.Name)
			}
			switch b.Kind {
			case RateBuff, AllProduction, TickSpeed:
				lo := d.MagMin*scale - eps
				hi := d.MagMax*scale + eps
				if b.Magnitude < lo || b.Magnitude > hi {
					t.Fatalf("%s magnitude %.6f outside [%.6f,%.6f] at scale %.2f", b.Name, b.Magnitude, lo, hi, scale)
				}
				if b.DurationTicks < d.DurMin || b.DurationTicks > d.DurMax {
					t.Fatalf("%s duration %d outside [%d,%d]", b.Name, b.DurationTicks, d.DurMin, d.DurMax)
				}
			case InstantResource:
				lo := d.AmountMin*instScale - eps
				hi := d.AmountMax*instScale + eps
				if b.InstantAmount < lo || b.InstantAmount > hi {
					t.Fatalf("%s amount %.4f outside [%.4f,%.4f] at instant scale %.4f",
						b.Name, b.InstantAmount, lo, hi, instScale)
				}
				if b.DurationTicks != 0 {
					t.Fatalf("%s is instant but has duration %d", b.Name, b.DurationTicks)
				}
			case TempWorkers:
				if b.InstantAmount < d.AmountMin-eps || b.InstantAmount > d.AmountMax+eps {
					t.Fatalf("%s worker count %.4f outside [%.2f,%.2f]", b.Name, b.InstantAmount, d.AmountMin, d.AmountMax)
				}
				if b.DurationTicks < d.DurMin || b.DurationTicks > d.DurMax {
					t.Fatalf("%s duration %d outside [%d,%d]", b.Name, b.DurationTicks, d.DurMin, d.DurMax)
				}
			}
			if b.Flavor == "" {
				t.Fatalf("roll %d (%s) has empty flavor", i, b.Name)
			}
		}
	}
}

// --- apply dispatch: each kind routes to the right method, correctly mapped --

func TestApplyRateBuffMapsToResourceRate(t *testing.T) {
	f := &fakeApplier{}
	line := Apply(Boon{Kind: RateBuff, Resource: "food", Magnitude: 0.13, DurationTicks: 4000, Name: "Test", Flavor: "flavor"}, f)

	if len(f.injected) != 1 || len(f.granted) != 0 || len(f.workers) != 0 {
		t.Fatalf("RateBuff routed wrong: %+v", f)
	}
	eff := f.injected[0].effects
	if len(eff) != 1 || eff[0].Type != "food_rate" || eff[0].Value != 0.13 {
		t.Fatalf("RateBuff mapped to wrong effect: %+v", eff)
	}
	if f.injected[0].ticks != 4000 || f.injected[0].name != "Test" {
		t.Fatalf("RateBuff ticks/name wrong: %+v", f.injected[0])
	}
	if line != "flavor" {
		t.Fatalf("Apply returned %q, want the boon flavor", line)
	}
}

func TestApplyAllProductionAndTickSpeedTypes(t *testing.T) {
	cases := []struct {
		kind Kind
		want string
	}{
		{AllProduction, "production_all"},
		{TickSpeed, "tick_speed"},
	}
	for _, c := range cases {
		f := &fakeApplier{}
		Apply(Boon{Kind: c.kind, Magnitude: 0.1, DurationTicks: 2500, Name: "X", Flavor: "f"}, f)
		if len(f.injected) != 1 {
			t.Fatalf("%v did not inject exactly one effect set", c.kind)
		}
		eff := f.injected[0].effects
		if len(eff) != 1 || eff[0].Type != c.want || eff[0].Value != 0.1 {
			t.Fatalf("%v mapped to %+v, want Type %q", c.kind, eff, c.want)
		}
	}
}

func TestApplyInstantResourceRoutesToGrant(t *testing.T) {
	f := &fakeApplier{}
	Apply(Boon{Kind: InstantResource, Resource: "gold", InstantAmount: 750, Name: "Grand Cache", Flavor: "f"}, f)
	if len(f.granted) != 1 || len(f.injected) != 0 || len(f.workers) != 0 {
		t.Fatalf("InstantResource routed wrong: %+v", f)
	}
	if f.granted[0].resource != "gold" || f.granted[0].amount != 750 {
		t.Fatalf("InstantResource granted %+v, want gold/750", f.granted[0])
	}
}

func TestApplyTempWorkersRoutesToWorkerGrant(t *testing.T) {
	f := &fakeApplier{}
	// InstantAmount carries the worker count; rounded on apply.
	Apply(Boon{Kind: TempWorkers, InstantAmount: 5.4, DurationTicks: 3000, Name: "Extra Hands", Flavor: "f"}, f)
	if len(f.workers) != 1 || len(f.injected) != 0 || len(f.granted) != 0 {
		t.Fatalf("TempWorkers routed wrong: %+v", f)
	}
	if f.workers[0].count != 5 || f.workers[0].ticks != 3000 {
		t.Fatalf("TempWorkers granted %+v, want count 5 / ticks 3000", f.workers[0])
	}
}

func TestApplyReservedKindsAreNoOps(t *testing.T) {
	for _, k := range []Kind{StorageCap, TradeIncome} {
		f := &fakeApplier{}
		line := Apply(Boon{Kind: k, Flavor: "reserved"}, f)
		if f.calls() != 0 {
			t.Fatalf("reserved kind %v triggered an Applier call: %+v", k, f)
		}
		if line != "reserved" {
			t.Fatalf("reserved kind %v dropped its flavor line", k)
		}
	}
}

// --- rolled boons apply through the fake end-to-end ------------------------

func TestRollThenApplyEndToEnd(t *testing.T) {
	p := DefaultProfile()
	p.Age = "space_age"
	p.Specialty = "iron"

	rng := rand.New(rand.NewSource(2026))
	for i := 0; i < 500; i++ {
		b := RollBoon(p, rng)
		f := &fakeApplier{}
		line := Apply(b, f)
		if line == "" {
			t.Fatalf("roll %d applied with empty flavor", i)
		}
		calls := f.calls()
		if calls != 1 {
			t.Fatalf("roll %d (%s) made %d Applier calls, want exactly 1", i, b.Kind, calls)
		}
	}
}

// --- decoupling guard: the package must never import game ------------------

func TestNoGameImport(t *testing.T) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			if strings.Contains(imp.Path.Value, "/ageforge/game") {
				t.Errorf("%s imports %s — boon must not depend on the game package", name, imp.Path.Value)
			}
		}
	}
}

// Sanity: DefaultProfile keeps every catalogued kind reachable.
func TestDefaultProfileRollsMultipleKinds(t *testing.T) {
	p := DefaultProfile()
	p.Age = "space_age"
	p.Specialty = "iron"
	seen := map[Kind]bool{}
	for _, b := range rollN(p, 1, 3000) {
		seen[b.Kind] = true
	}
	// The five catalogued kinds should all appear over many neutral rolls.
	for _, k := range []Kind{RateBuff, AllProduction, TickSpeed, InstantResource, TempWorkers} {
		if !seen[k] {
			t.Errorf("kind %v never rolled under DefaultProfile", k)
		}
	}
}
