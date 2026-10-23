package config

import (
	"math"
	"strconv"
	"strings"
)

// goUnits are the names in units.go, indexed by thousands: goUnits[2] is
// Million, goUnits[11] is Decillion.
var goUnits = [...]string{2: "Million", 3: "Billion", 4: "Trillion", 5: "Quadrillion",
	6: "Quintillion", 7: "Sextillion", 8: "Septillion", 9: "Octillion", 10: "Nonillion", 11: "Decillion"}

// GoLiteral formats v as it should be written in Go source in this repo, for
// tools and test messages that print a number to be pasted into config:
// 75.4 * Million from a million up (Million alone for an exact unit), plain
// digits with an underscore every three digits from five digits up below that
// (22_700, 950_000, 4_500.5), and the exact value either way. A number past
// the unit names, or smaller than 0.0001, comes back as strconv prints it.
func GoLiteral(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) || v == 0 {
		return strconv.FormatFloat(v, 'g', -1, 64)
	}
	neg := ""
	if v < 0 {
		neg, v = "-", -v
	}
	// the shortest digits that read back as v: d.ddd e x
	s := strconv.FormatFloat(v, 'e', -1, 64)
	mant, expo, _ := strings.Cut(s, "e")
	x, _ := strconv.Atoi(expo)
	digits := strings.Replace(mant, ".", "", 1)
	point := x + 1 // digits before the decimal point; 75400000 has 8
	switch {
	case point >= 7:
		k := (point - 1) / 3 // thousands: 2 is a million
		if k >= len(goUnits) {
			return neg + strconv.FormatFloat(v, 'g', -1, 64)
		}
		before := point - 3*k // 1, 2 or 3 digits stay in front of the unit
		m := digits
		if len(m) < before {
			m += strings.Repeat("0", before-len(m))
		} else if len(m) > before {
			m = m[:before] + "." + m[before:]
		}
		if m == "1" {
			return neg + goUnits[k]
		}
		return neg + m + " * " + goUnits[k]
	case point >= 1:
		ip := digits
		if len(ip) < point {
			ip += strings.Repeat("0", point-len(ip))
		}
		out := ip[:point]
		if len(ip) >= 5 { // five digits in all, as 4_500.5 has
			out = groupDigits(out)
		}
		if len(ip) > point {
			out += "." + ip[point:]
		}
		return neg + out
	case point > -4: // down to 0.0001
		return neg + "0." + strings.Repeat("0", -point) + digits
	}
	return neg + strconv.FormatFloat(v, 'g', -1, 64)
}

// groupDigits puts an underscore every three digits of s.
func groupDigits(s string) string {
	first := len(s) % 3
	var b strings.Builder
	b.WriteString(s[:first])
	for i := first; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('_')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
