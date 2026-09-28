package detmath

import (
	"math"
	"runtime"
	"testing"
)

// sweep calls fn with n positive normal floats spread over [lo, hi)
// geometrically, plus a jitter so the low mantissa bits vary.
func sweep(n int, lo, hi float64, fn func(x float64)) {
	step := math.Exp((math.Log(hi) - math.Log(lo)) / float64(n))
	x := lo
	for i := 0; i < n; i++ {
		fn(x)
		fn(math.Nextafter(x, math.Inf(1)))
		x = float64(x*step) + float64(x*1e-13*float64(i%7))
	}
}

// The detmath results are pinned bit for bit: this test passing on an
// architecture is what "the same on every architecture" means. CI runs it
// on amd64 and arm64 (the Go and Determinism workflows).
func TestGoldenBits(t *testing.T) {
	cases := []struct {
		name string
		got  float64
		want uint64
	}{
		{"Log(2)", Log(2), 0x3fe62e42fefa39ef},
		{"Log(0.37)", Log(0.37), 0xbfefd0ea24bf89b8},
		{"Log(1e300)", Log(1e300), 0x4085963447f87fb5},
		{"Log10(1000)", Log10(1000), 0x4008000000000000},
		{"Log10(46.3)", Log10(46.3), 0x3ffaa63840d42e9b},
		{"Exp(1)", Exp(1), 0x4005bf0a8b145769},
		{"Exp(-3.25)", Exp(-3.25), 0x3fa3da368521902d},
		{"Exp(700)", Exp(700), 0x7f0d945df4f8ec8e},
		// The boon rarity bias: base weights 10, 5, 2 at 1/1.5 and 1/1.6.
		{"Pow(10,1/1.5)", Pow(10, 1/1.5), 0x401290fca9c761f7},
		{"Pow(10,1/1.6)", Pow(10, 1/1.6), 0x4010de2c14fa8852},
		{"Pow(5,1/1.5)", Pow(5, 1/1.5), 0x400764636974629c},
		{"Pow(5,1/1.6)", Pow(5, 1/1.6), 0x4005dff9fc52fc15},
		{"Pow(2,1/1.5)", Pow(2, 1/1.5), 0x3ff965fea53d6e3c},
		{"Pow(2,1/1.6)", Pow(2, 1/1.6), 0x3ff8ace5422aa0dc},
		{"Pow(1.15,37)", Pow(1.15, 37), 0x406603fcf77ab881},
		{"Pow(4/3,0.75)", Pow(4.0/3, 0.75), 0x3ff3da57e4f1f790},
	}
	for _, c := range cases {
		if got := math.Float64bits(c.got); got != c.want {
			t.Errorf("%s = %v (%#x), want %v (%#x)", c.name, c.got, got, math.Float64frombits(c.want), c.want)
		}
	}
}

// Log and Log10 are the algorithm amd64's assembly math.Log runs, in the same
// order, so on amd64 they must match the standard library exactly. This is
// what keeps the Log10 call sites' results unchanged on amd64.
func TestLogMatchesStdlibOnAMD64(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skipf("math.Log is only assembly (and FMA-free) on amd64, not %s", runtime.GOARCH)
	}
	bad := 0
	sweep(200000, 1e-300, 1e300, func(x float64) {
		if g, w := Log(x), math.Log(x); g != w && bad < 10 {
			bad++
			t.Errorf("Log(%v) = %v, math.Log = %v", x, g, w)
		}
		if g, w := Log10(x), math.Log10(x); g != w && bad < 10 {
			bad++
			t.Errorf("Log10(%v) = %v, math.Log10 = %v", x, g, w)
		}
	})
}

// With an integer exponent math.Pow only multiplies, so it is exact to the
// same bits everywhere, and Pow must agree with it.
func TestPowIntegerMatchesStdlib(t *testing.T) {
	for _, b := range []float64{0.5, 0.98, 1.1, 1.13, 1.15, 1.5, 2, 3.7, 10, 123.456} {
		for n := -40; n <= 400; n++ {
			if g, w := Pow(b, float64(n)), math.Pow(b, float64(n)); math.Float64bits(g) != math.Float64bits(w) {
				t.Fatalf("Pow(%v, %d) = %v, math.Pow = %v", b, n, g, w)
			}
		}
	}
}

// Accuracy: Log within 1 ulp of the standard library, Exp within 2 (amd64's
// assembly Exp is a different algorithm), Pow within a few (exp(y*log x)
// carries log's rounding error times |y*log x|, as math.Pow's own fractional
// path does).
func TestCloseToStdlib(t *testing.T) {
	ulps := func(a, b float64) uint64 {
		x, y := math.Float64bits(a), math.Float64bits(b)
		if x > y {
			return x - y
		}
		return y - x
	}
	sweep(100000, 1e-300, 1e300, func(x float64) {
		if d := ulps(Log(x), math.Log(x)); d > 1 {
			t.Fatalf("Log(%v): %d ulps from math.Log", x, d)
		}
	})
	sweep(100000, 1e-6, 700, func(x float64) {
		for _, v := range []float64{x, -x} {
			if d := ulps(Exp(v), math.Exp(v)); d > 2 {
				t.Fatalf("Exp(%v): %d ulps from math.Exp", v, d)
			}
		}
	})
	sweep(20000, 1e-3, 1e3, func(x float64) {
		for _, y := range []float64{1 / 1.5, 1 / 1.6, 0.75, 1.3, -2.5} {
			if d := ulps(Pow(x, y), math.Pow(x, y)); d > 8 {
				t.Fatalf("Pow(%v, %v): %d ulps from math.Pow", x, y, d)
			}
		}
	})
}

func TestSpecialCases(t *testing.T) {
	inf, nan := math.Inf(1), math.NaN()
	for _, x := range []float64{0, -1, inf, nan, 1} {
		if g, w := Log(x), math.Log(x); !(g == w || math.IsNaN(g) && math.IsNaN(w)) {
			t.Errorf("Log(%v) = %v, want %v", x, g, w)
		}
	}
	for _, x := range []float64{-inf, inf, nan, 0, 1e-10, 710, -746} {
		if g, w := Exp(x), math.Exp(x); !(g == w || math.IsNaN(g) && math.IsNaN(w)) {
			t.Errorf("Exp(%v) = %v, want %v", x, g, w)
		}
	}
	for _, p := range [][2]float64{{0, -1}, {-0.0, -3}, {-2, 0.5}, {-8, 1.0 / 3}, {2, inf}, {0.5, inf}, {-inf, 3}, {inf, -1}, {nan, 0}, {1, nan}} {
		if g, w := Pow(p[0], p[1]), math.Pow(p[0], p[1]); !(g == w || math.IsNaN(g) && math.IsNaN(w)) {
			t.Errorf("Pow(%v, %v) = %v, want %v", p[0], p[1], g, w)
		}
	}
}
