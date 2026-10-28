package game

import (
	"strings"
	"time"

	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Short names for the player-facing text helpers, kept so the Army code
// reads the same. Each one delegates to the shared helpers (names.go and
// pkg/textfmt) so there is one number format and one set of display names.

// resourceLabel is a resource's player-facing name in lower case ("dark
// matter" for dark_matter).
func resourceLabel(key string) string { return ResourceName(key) }

// amountText prints an amount for a log line: "0.4", "12", "950", "12.5K".
func amountText(v float64) string { return textfmt.Number(v) }

// amountsText lists amounts as "120 food, 40 gold", sorted by resource key,
// skipping zero and negative entries.
func amountsText(m map[string]float64) string {
	return strings.Join(amountParts(m, isPositive), ", ")
}

func isPositive(v float64) bool { return v > 0 }

// amountParts lists the entries of m that keep accepts as "120 food", sorted
// by resource key.
func amountParts(m map[string]float64, keep func(float64) bool) []string {
	var parts []string
	for _, key := range sortedKeys(m) {
		if keep(m[key]) {
			parts = append(parts, Amount(m[key], key))
		}
	}
	return parts
}

// approxTicks renders a tick count at tick interval iv as an approximate
// wall-clock duration: "~45s", "~4m 44s", "~1h 12m".
func approxTicks(ticks int, iv time.Duration) string { return DurationText(ticks, iv) }

// joinAnd joins parts as "a", "a and b", "a, b and c".
func joinAnd(parts []string) string { return textfmt.List(parts) }

// lossParts lists resource amounts and a worker count for a log line, e.g.
// ["8 food", "3 workers"]. Resources are sorted by key.
func lossParts(resources map[string]float64, workers int) []string {
	parts := amountParts(resources, isPositive)
	if workers > 0 {
		parts = append(parts, countWorkersRealOne(workers))
	}
	return parts
}
