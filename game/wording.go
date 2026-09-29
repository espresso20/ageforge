package game

import (
	"sort"
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

// ageLabel is an age's display name ("Industrial Age").
func ageLabel(key string) string { return AgeName(key) }

// amountText prints an amount for a log line: "0.4", "12", "950", "12.5K".
func amountText(v float64) string { return textfmt.Number(v) }

// amountsText lists amounts as "120 food, 40 gold", sorted by resource key,
// skipping zero and negative entries.
func amountsText(m map[string]float64) string {
	var parts []string
	for _, k := range sortedKeys(m) {
		if m[k] > 0 {
			parts = append(parts, Amount(m[k], k))
		}
	}
	return strings.Join(parts, ", ")
}

// approxTicks renders a tick count at tick interval iv as an approximate
// wall-clock duration: "~45s", "~4m 44s", "~1h 12m".
func approxTicks(ticks int, iv time.Duration) string { return DurationText(ticks, iv) }

// joinAnd joins parts as "a", "a and b", "a, b and c".
func joinAnd(parts []string) string { return textfmt.List(parts) }

// lossParts lists resource amounts and a worker count for a log line, e.g.
// ["8 food", "3 workers"]. Resources are sorted by key.
func lossParts(resources map[string]float64, workers int) []string {
	keys := make([]string, 0, len(resources))
	for k, v := range resources {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, Amount(resources[k], k))
	}
	if workers > 0 {
		parts = append(parts, textfmt.Count(workers, "worker", "workers"))
	}
	return parts
}
