package smoke

import (
	"strings"
	"testing"
)

func TestFirstDiff(t *testing.T) {
	type inner struct{ V []float64 }
	type s struct {
		M map[string]float64
		P *inner
		T []string
	}
	mk := func() s {
		return s{M: map[string]float64{"a": 1, "b": 0.1 + 0.2}, P: &inner{V: []float64{1, 2}}, T: nil}
	}
	a, b := mk(), mk()
	if d := firstDiff(a, b, nil); d != "" {
		t.Fatalf("equal values differ: %s", d)
	}
	b.P.V[1] = 2 + 1e-15
	if d := firstDiff(a, b, nil); !strings.HasPrefix(d, "P.V[1]") {
		t.Errorf("a one-ulp change should be found at P.V[1], got %q", d)
	}
	b = mk()
	b.T = []string{}
	if d := firstDiff(a, b, nil); d != "" {
		t.Errorf("nil and empty slices should match, got %q", d)
	}
	b.M["c"] = 3
	if d := firstDiff(a, b, func(p string) bool { return p == "M" }); d != "" {
		t.Errorf("skipped path still reported: %q", d)
	}
}
