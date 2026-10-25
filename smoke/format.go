package smoke

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

func bad(v float64) bool { return math.IsNaN(v) || math.IsInf(v, 0) }

// num formats a number compactly: whole below ten thousand, and from there
// the game's own format (textfmt.Number: 12.5K, 3.4M, 1.56T, 8.9Q), so the
// report reads as the game does. It used to stop at T and printed the Cosmic
// Era's quadrillions as 8900T in scientific form.
func num(v float64) string {
	switch {
	case bad(v):
		return fmt.Sprint(v)
	case math.Abs(v) >= 10_000:
		return textfmt.Number(v)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

// cell keeps a table cell on one line.
func cell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", "\\|")
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
