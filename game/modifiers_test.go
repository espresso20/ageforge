package game

import (
	"math"
	"reflect"
	"testing"
)

// floatEq compares with a small epsilon for cases where binary floating point
// can't represent the decimal exactly.
func floatEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestTotal_AddOnly(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "a", Target: "production_all", Op: OpAdd, Value: 0.10})
	r.Add(Modifier{Source: "b", Target: "production_all", Op: OpAdd, Value: 0.10})
	r.Add(Modifier{Source: "c", Target: "production_all", Op: OpAdd, Value: 0.10})

	// (1 + 0.30) × 1 = 1.30
	if got := r.Total("production_all"); !floatEq(got, 1.30) {
		t.Fatalf("add-only total = %v, want 1.30", got)
	}
}

func TestTotal_MulOnly(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "boost", Target: "tick_speed", Op: OpMul, Value: 1.2})
	r.Add(Modifier{Source: "catastrophe", Target: "tick_speed", Op: OpMul, Value: 0.5})

	// 1.2 × 0.5 = 0.60 — exact in binary float.
	if got := r.Total("tick_speed"); got != 0.60 {
		t.Fatalf("mul-only total = %v, want exactly 0.60", got)
	}
}

func TestTotal_Mixed_OrderingMatters(t *testing.T) {
	r := NewResolver()
	// Adds sum to +0.30.
	r.Add(Modifier{Source: "r1", Target: "food_rate", Op: OpAdd, Value: 0.20})
	r.Add(Modifier{Source: "r2", Target: "food_rate", Op: OpAdd, Value: 0.10})
	// One independent multiplier of 0.8.
	r.Add(Modifier{Source: "morale", Target: "food_rate", Op: OpMul, Value: 0.8})

	got := r.Total("food_rate")

	// Correct ordering: (1 + 0.30) × 0.8 = 1.30 × 0.8 = 1.04.
	want := 1.04
	if !floatEq(got, want) {
		t.Fatalf("mixed total = %v, want %v ((1+ΣAdd)×ΠMul)", got, want)
	}

	// Guard against the WRONG folding where the mul is treated as an additive
	// percentage: (1 + 0.30 + 0.8) = 2.10. Make sure we are nowhere near it.
	wrongFold := 1 + 0.30 + 0.8
	if floatEq(got, wrongFold) {
		t.Fatalf("total %v matches the wrong (1+ΣAdd+mul) folding", got)
	}
}

func TestTotal_UnknownTarget(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "a", Target: "production_all", Op: OpAdd, Value: 0.5})

	if got := r.Total("does_not_exist"); got != 1.0 {
		t.Fatalf("unknown target total = %v, want 1.0", got)
	}
	if bd := r.Breakdown("does_not_exist"); len(bd) != 0 {
		t.Fatalf("unknown target breakdown len = %d, want 0", len(bd))
	}
}

func TestTotal_EmptyResolver(t *testing.T) {
	r := NewResolver()
	if got := r.Total("anything"); got != 1.0 {
		t.Fatalf("empty resolver total = %v, want 1.0", got)
	}
}

func TestBreakdown_ContentsAndAttribution(t *testing.T) {
	r := NewResolver()
	m1 := Modifier{Source: "research:masonry", Target: "production_all", Op: OpAdd, Value: 0.10}
	m2 := Modifier{Source: "wonder:colossus", Target: "production_all", Op: OpMul, Value: 1.18}
	r.Add(m1)
	r.Add(m2)

	bd := r.Breakdown("production_all")
	if len(bd) != 2 {
		t.Fatalf("breakdown len = %d, want 2", len(bd))
	}
	// Insertion order preserved, sources intact.
	if bd[0] != m1 || bd[1] != m2 {
		t.Fatalf("breakdown = %+v, want [%+v %+v] in order", bd, m1, m2)
	}
}

func TestBreakdown_ReturnsCopy(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "research:masonry", Target: "production_all", Op: OpAdd, Value: 0.10})

	bd := r.Breakdown("production_all")
	// Corrupt the returned slice's element.
	bd[0].Value = 999.0
	bd[0].Source = "tampered"

	// Resolver must be unaffected: Total still reflects the original 0.10.
	if got := r.Total("production_all"); !floatEq(got, 1.10) {
		t.Fatalf("after mutating returned breakdown, total = %v, want 1.10 (copy not honored)", got)
	}
	// And a fresh Breakdown must still show the original source.
	fresh := r.Breakdown("production_all")
	if fresh[0].Source != "research:masonry" || fresh[0].Value != 0.10 {
		t.Fatalf("resolver state corrupted by mutating returned slice: %+v", fresh[0])
	}
}

func TestTargets_SortedAndDeduped(t *testing.T) {
	r := NewResolver()
	// Insert out of order, with multiple mods sharing a target.
	r.Add(Modifier{Source: "x", Target: "tick_speed", Op: OpAdd, Value: 0.1})
	r.Add(Modifier{Source: "y", Target: "production_all", Op: OpAdd, Value: 0.1})
	r.Add(Modifier{Source: "z", Target: "production_all", Op: OpMul, Value: 1.2})
	r.Add(Modifier{Source: "w", Target: "food_rate", Op: OpAdd, Value: 0.1})

	got := r.Targets()
	want := []string{"food_rate", "production_all", "tick_speed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Targets() = %v, want %v (sorted, deduped)", got, want)
	}
}

func TestTargets_Empty(t *testing.T) {
	r := NewResolver()
	if got := r.Targets(); len(got) != 0 {
		t.Fatalf("Targets() on empty resolver = %v, want empty", got)
	}
}

func TestTotal_NoOpsDoNotChangeResult(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "real", Target: "gather_rate", Op: OpAdd, Value: 0.25})
	r.Add(Modifier{Source: "noop_add", Target: "gather_rate", Op: OpAdd, Value: 0.0})
	r.Add(Modifier{Source: "noop_mul", Target: "gather_rate", Op: OpMul, Value: 1.0})

	if got := r.Total("gather_rate"); !floatEq(got, 1.25) {
		t.Fatalf("total with no-ops = %v, want 1.25", got)
	}
	// Add stores everything, so the no-ops are still present in the Breakdown.
	if bd := r.Breakdown("gather_rate"); len(bd) != 3 {
		t.Fatalf("breakdown len = %d, want 3 (no-ops retained)", len(bd))
	}
}

func TestAddAll(t *testing.T) {
	r := NewResolver()
	r.AddAll([]Modifier{
		{Source: "a", Target: "production_all", Op: OpAdd, Value: 0.10},
		{Source: "b", Target: "production_all", Op: OpAdd, Value: 0.20},
		{Source: "c", Target: "tick_speed", Op: OpMul, Value: 1.5},
	})

	if got := r.Total("production_all"); !floatEq(got, 1.30) {
		t.Fatalf("AddAll production_all total = %v, want 1.30", got)
	}
	if got := r.Total("tick_speed"); !floatEq(got, 1.5) {
		t.Fatalf("AddAll tick_speed total = %v, want 1.5", got)
	}
}

// TestAll_RoundTrips verifies All() is the inverse of AddAll: flattening a
// resolver and rebuilding from the dump reproduces every Total. This is the
// path GetState → GameState.Modifiers → UI relies on.
func TestAll_RoundTrips(t *testing.T) {
	src := []Modifier{
		{Source: "research", Target: "production_all", Op: OpAdd, Value: 0.10},
		{Source: "morale", Target: "production_all", Op: OpMul, Value: 1.18},
		{Source: "prestige", Target: "tick_speed", Op: OpAdd, Value: 0.05},
		{Source: "event:peace", Target: "tick_speed", Op: OpAdd, Value: 0.01},
	}
	orig := NewResolver()
	orig.AddAll(src)

	flat := orig.All()
	if len(flat) != len(src) {
		t.Fatalf("All() len = %d, want %d", len(flat), len(src))
	}

	rebuilt := NewResolver()
	rebuilt.AddAll(flat)
	for _, target := range orig.Targets() {
		if a, b := orig.Total(target), rebuilt.Total(target); !floatEq(a, b) {
			t.Errorf("Total(%q): orig %v != rebuilt %v", target, a, b)
		}
	}

	// Empty resolver yields a non-nil, zero-length slice (safe for AddAll).
	if got := NewResolver().All(); got == nil || len(got) != 0 {
		t.Errorf("empty All() = %v, want non-nil empty slice", got)
	}
}

// TestAddTotal_SumsOpAddIgnoresOpMul verifies AddTotal pools only OpAdd values
// and never folds in an OpMul factor, and that an unknown target is 0.
func TestAddTotal_SumsOpAddIgnoresOpMul(t *testing.T) {
	r := NewResolver()
	r.Add(Modifier{Source: "research", Target: "production_all", Op: OpAdd, Value: 0.25})
	r.Add(Modifier{Source: "prestige", Target: "production_all", Op: OpAdd, Value: 0.10})
	r.Add(Modifier{Source: "morale", Target: "production_all", Op: OpMul, Value: 1.18})
	r.Add(Modifier{Source: "wonders", Target: "gold_rate", Op: OpAdd, Value: 0.40})

	if got, want := r.AddTotal("production_all"), 0.35; math.Abs(got-want) > 1e-9 {
		t.Errorf("production_all AddTotal = %.12f, want %.12f (OpMul must be ignored)", got, want)
	}
	if got, want := r.AddTotal("gold_rate"), 0.40; math.Abs(got-want) > 1e-9 {
		t.Errorf("gold_rate AddTotal = %.12f, want %.12f", got, want)
	}
	if got := r.AddTotal("nonexistent_target"); got != 0 {
		t.Errorf("unknown target AddTotal = %.12f, want 0", got)
	}
}
