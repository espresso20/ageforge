// Package detmath is the math library for simulation code: Log, Log10, Exp
// and Pow that return the same bits on every architecture and CPU.
//
// The standard library's versions do not. math.Log is assembly on amd64 and
// Go elsewhere; math.Exp is assembly on amd64 (with a separate FMA path
// picked at run time from the CPU's features) and on arm64; and wherever they
// are Go, the compiler may fuse a*b+c into one FMA instruction on arm64 and
// round differently. math.Pow with a fractional exponent goes through Exp and
// Log. Any of these can move the last bit of a result, and in a simulation
// one bit is enough for a run to drift apart.
//
// These are the Go standard library's pure-Go algorithms (FreeBSD's msun,
// via Go's math package; BSD licence, see below) with every product that
// feeds an addition rounded by an explicit float64() conversion, so none can
// be fused. The results:
//
//   - Log and Log10 equal math.Log and math.Log10 on amd64 bit for bit.
//   - Pow with an integer exponent equals math.Pow everywhere: that path is
//     repeated multiplication.
//   - Exp, and Pow with a fractional exponent, can differ from the standard
//     library's in the last bit, on any architecture.
//
// Sqrt, Floor, Ceil, Trunc, Round, Abs, Mod, Min and Max are exact (IEEE 754
// or pure bit manipulation) and the same everywhere; call package math for
// those. See "Float rules for simulation code" in CONTRIBUTING.md.
//
// Portions copyright 2009 The Go Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found at
// https://go.dev/LICENSE.
package detmath

import "math"

const (
	ln2Hi = 6.93147180369123816490e-01 // 0x3fe62e42fee00000
	ln2Lo = 1.90821492927058770002e-10 // 0x3dea39ef35793c76
	log2e = 1.44269504088896338700e+00
)

// Log returns the natural logarithm of x. Special cases as math.Log.
func Log(x float64) float64 {
	const (
		L1 = 6.666666666666735130e-01 // 0x3FE5555555555593
		L2 = 3.999999999940941908e-01 // 0x3FD999999997FA04
		L3 = 2.857142874366239149e-01 // 0x3FD2492494229359
		L4 = 2.222219843214978396e-01 // 0x3FCC71C51D8E78AF
		L5 = 1.818357216161805012e-01 // 0x3FC7466496CB03DE
		L6 = 1.531383769920937332e-01 // 0x3FC39A09D078C69F
		L7 = 1.479819860511658591e-01 // 0x3FC2F112DF3E5244
	)
	switch {
	case math.IsNaN(x) || math.IsInf(x, 1):
		return x
	case x < 0:
		return math.NaN()
	case x == 0:
		return math.Inf(-1)
	}

	f1, ki := math.Frexp(x)
	if f1 < math.Sqrt2/2 {
		f1 *= 2
		ki--
	}
	f := f1 - 1
	k := float64(ki)

	s := f / (2 + f)
	s2 := s * s
	s4 := s2 * s2
	t1 := float64(s2 * (L1 + float64(s4*(L3+float64(s4*(L5+float64(s4*L7)))))))
	t2 := float64(s4 * (L2 + float64(s4*(L4+float64(s4*L6)))))
	R := t1 + t2
	hfsq := float64(0.5 * f * f)
	return float64(k*ln2Hi) - ((hfsq - (float64(s*(hfsq+R)) + float64(k*ln2Lo))) - f)
}

// Log10 returns the decimal logarithm of x. Special cases as math.Log10.
func Log10(x float64) float64 {
	return float64(Log(x) * (1 / math.Ln10))
}

// Exp returns e**x. Special cases as math.Exp.
func Exp(x float64) float64 {
	const (
		overflow  = 7.09782712893383973096e+02
		underflow = -7.45133219101941108420e+02
		nearZero  = 1.0 / (1 << 28) // 2**-28
	)
	switch {
	case math.IsNaN(x):
		return x
	case x > overflow: // handles +Inf
		return math.Inf(1)
	case x < underflow: // handles -Inf
		return 0
	case -nearZero < x && x < nearZero:
		return 1 + x
	}

	// Reduce: x = hi - lo = k*ln2 + r, |r| <= 0.5*ln2.
	var k int
	switch {
	case x < 0:
		k = int(float64(log2e*x) - 0.5)
	case x > 0:
		k = int(float64(log2e*x) + 0.5)
	}
	hi := x - float64(float64(k)*ln2Hi)
	lo := float64(float64(k) * ln2Lo)
	return expmulti(hi, lo, k)
}

// expmulti returns e**r × 2**k where r = hi - lo and |r| <= ln2/2.
func expmulti(hi, lo float64, k int) float64 {
	const (
		P1 = 1.66666666666666657415e-01  // 0x3FC55555; 0x55555555
		P2 = -2.77777777770155933842e-03 // 0xBF66C16C; 0x16BEBD93
		P3 = 6.61375632143793436117e-05  // 0x3F11566A; 0xAF25DE2C
		P4 = -1.65339022054652515390e-06 // 0xBEBBBD41; 0xC5D26BF1
		P5 = 4.13813679705723846039e-08  // 0x3E663769; 0x72BEA4D0
	)
	r := hi - lo
	t := r * r
	c := r - float64(t*(P1+float64(t*(P2+float64(t*(P3+float64(t*(P4+float64(t*P5)))))))))
	y := 1 - ((lo - float64(r*c)/(2-c)) - hi)
	return math.Ldexp(y, k)
}

// Pow returns x**y. Special cases as math.Pow. With an integer y it is
// repeated multiplication and equals math.Pow on every architecture.
func Pow(x, y float64) float64 {
	switch {
	case y == 0 || x == 1:
		return 1
	case y == 1:
		return x
	case math.IsNaN(x) || math.IsNaN(y):
		return math.NaN()
	case x == 0:
		switch {
		case y < 0:
			if math.Signbit(x) && isOddInt(y) {
				return math.Inf(-1)
			}
			return math.Inf(1)
		case y > 0:
			if math.Signbit(x) && isOddInt(y) {
				return x
			}
			return 0
		}
	case math.IsInf(y, 0):
		switch {
		case x == -1:
			return 1
		case (math.Abs(x) < 1) == math.IsInf(y, 1):
			return 0
		default:
			return math.Inf(1)
		}
	case math.IsInf(x, 0):
		if math.IsInf(x, -1) {
			return Pow(1/x, -y) // Pow(-0, -y)
		}
		switch {
		case y < 0:
			return 0
		case y > 0:
			return math.Inf(1)
		}
	case y == 0.5:
		return math.Sqrt(x)
	case y == -0.5:
		return 1 / math.Sqrt(x)
	}

	yi, yf := math.Modf(math.Abs(y))
	if yf != 0 && x < 0 {
		return math.NaN()
	}
	if yi >= 1<<63 {
		// yi is a large even int that will lead to overflow (or underflow to
		// 0) for all x except -1 (x == 1 was handled earlier).
		switch {
		case x == -1:
			return 1
		case (math.Abs(x) < 1) == (y > 0):
			return 0
		default:
			return math.Inf(1)
		}
	}

	// ans = a1 * 2**ae (= 1 for now).
	a1 := 1.0
	ae := 0

	// ans *= x**yf
	if yf != 0 {
		if yf > 0.5 {
			yf--
			yi++
		}
		a1 = Exp(yf * Log(x))
	}

	// ans *= x**yi by repeated squaring, tracking the exponent separately
	// to avoid overflow and underflow in the intermediates.
	x1, xe := math.Frexp(x)
	for i := int64(yi); i != 0; i >>= 1 {
		if xe < -1<<12 || 1<<12 < xe {
			// catastrophic overflow; let Ldexp handle it.
			ae += xe
			break
		}
		if i&1 == 1 {
			a1 *= x1
			ae += xe
		}
		x1 = float64(x1 * x1) // else x1 += x1 below fuses with this product
		xe <<= 1
		if x1 < .5 {
			x1 += x1
			xe--
		}
	}

	// ans = a1 * 2**ae; if y < 0 { ans = 1 / ans }, with the inversion
	// done before Ldexp so the exponent stays in range.
	if y < 0 {
		a1 = 1 / a1
		ae = -ae
	}
	return math.Ldexp(a1, ae)
}

func isOddInt(x float64) bool {
	if math.Abs(x) >= (1 << 53) {
		// Every float this large is an even integer.
		return false
	}
	xi, xf := math.Modf(x)
	return xf == 0 && int64(xi)&1 == 1
}
