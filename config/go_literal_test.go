package config

import (
	"math"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

func TestGoLiteral(t *testing.T) {
	for _, c := range []struct {
		v    float64
		want string
	}{
		{75.4 * Million, "75.4 * Million"},
		{1.52 * Million, "1.52 * Million"},
		{2.5 * Quadrillion, "2.5 * Quadrillion"},
		{Million, "Million"},
		{Sextillion, "Sextillion"},
		{Decillion, "Decillion"},
		{999.999999 * Million, "999.999999 * Million"},
		{3 * Billion, "3 * Billion"},
		{-3 * Billion, "-3 * Billion"},
		{344.4363246 * Quintillion, "344.4363246 * Quintillion"},
		{22700, "22_700"},
		{950000, "950_000"},
		{999_999, "999_999"},
		{4500.5, "4_500.5"},
		{9999, "9999"},
		{7500, "7500"},
		{0.75, "0.75"},
		{0.0001, "0.0001"},
		{-12.5, "-12.5"},
		{0, "0"},
		{1e-9, "1e-09"},
		{1e300, "1e+300"},
		{1000 * Decillion, "1e+36"},
		{math.Inf(1), "+Inf"},
	} {
		if got := GoLiteral(c.v); got != c.want {
			t.Errorf("GoLiteral(%v) = %q, want %q", c.v, got, c.want)
		}
	}
}

// TestGoLiteralReadsBackExactly: what GoLiteral prints is the same float64
// when Go reads it as source, where "m * Unit" is worked out exactly.
func TestGoLiteralReadsBackExactly(t *testing.T) {
	for k := 2; k < len(goUnits); k++ {
		want, _ := strconv.ParseFloat("1e"+strconv.Itoa(3*k), 64)
		if got := map[string]float64{"Million": Million, "Billion": Billion, "Trillion": Trillion, "Quadrillion": Quadrillion,
			"Quintillion": Quintillion, "Sextillion": Sextillion, "Septillion": Septillion, "Octillion": Octillion,
			"Nonillion": Nonillion, "Decillion": Decillion}[goUnits[k]]; got != want {
			t.Errorf("%s = %v, want %v", goUnits[k], got, want)
		}
	}
	pow := func(k int) *big.Rat {
		return new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(3*k)), nil))
	}
	read := func(lit string) float64 { // the exact value of the source text, as a float64
		lit = strings.ReplaceAll(lit, "_", "")
		m, unit, ok := strings.Cut(lit, " * ")
		if !ok && strings.IndexFunc(lit, unicode.IsLetter) >= 0 {
			m, unit = "1", lit // a bare unit name
		}
		r, ok := new(big.Rat).SetString(m)
		if !ok {
			t.Fatalf("cannot read %q", lit)
		}
		for k := 2; k < len(goUnits); k++ {
			if goUnits[k] == unit {
				r.Mul(r, pow(k))
			}
		}
		f, _ := r.Float64()
		return f
	}
	for _, v := range []float64{Million, Billion, Quadrillion, Decillion, 2.5 * Quadrillion, 75.4 * Million, 999.999999 * Million} {
		if got := read(GoLiteral(v)); got != v {
			t.Errorf("GoLiteral(%v) = %q reads back as %v", v, GoLiteral(v), got)
		}
	}
	r := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < 5000; i++ {
		v := float64(r.IntN(9_999_999)+1) / math.Pow10(r.IntN(8)) * math.Pow10(r.IntN(36)-3)
		if v < 1e-4 {
			continue
		}
		lit := GoLiteral(v)
		if strings.Contains(lit, "e") {
			continue
		}
		if got := read(lit); got != v {
			t.Fatalf("GoLiteral(%v) = %q reads back as %v", v, lit, got)
		}
	}
}
