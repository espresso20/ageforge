package game

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Small helpers for player-facing log lines and refusals written in package
// game: display names instead of config keys, readable amounts, and wall-clock
// durations instead of ticks. They mirror the UI's formatters (ui/widgets.go
// FormatNumber, ui/format_ticks.go formatTicks), which package game cannot
// import.

// resourceLabel is a resource's player-facing name in lower case ("dark
// matter" for dark_matter). An unknown key falls back to the key with spaces.
func resourceLabel(key string) string {
	if def, ok := config.ResourceByKey()[key]; ok && def.Name != "" {
		return strings.ToLower(def.Name)
	}
	return strings.ReplaceAll(key, "_", " ")
}

// ageLabel is an age's display name ("Industrial Age"), or the key when the
// age is unknown.
func ageLabel(key string) string {
	if def, ok := config.AgeByKey()[key]; ok && def.Name != "" {
		return def.Name
	}
	return key
}

// amountText prints an amount for a log line: "0.4", "12", "950", "12.5K",
// "3.1M".
func amountText(v float64) string {
	switch {
	case v >= 1000:
		return formatPlanAmount(v)
	case v >= 1 || v == 0:
		return fmt.Sprintf("%.0f", v)
	default:
		return fmt.Sprintf("%.1f", v)
	}
}

// amountsText lists amounts as "120 food, 40 gold", sorted by resource key,
// skipping zero and negative entries.
func amountsText(m map[string]float64) string {
	var parts []string
	for _, k := range sortedKeys(m) {
		if m[k] > 0 {
			parts = append(parts, amountText(m[k])+" "+resourceLabel(k))
		}
	}
	return strings.Join(parts, ", ")
}

// approxTicks renders a tick count at tick interval iv as an approximate
// wall-clock duration, the same way the UI does: "~45s", "~4m 44s", "~1h 12m".
func approxTicks(ticks int, iv time.Duration) string {
	if iv <= 0 {
		iv = BaseTickInterval
	}
	total := int((time.Duration(ticks) * iv).Round(time.Second) / time.Second)
	if total < 0 {
		total = 0
	}
	switch {
	case total < 60:
		return fmt.Sprintf("~%ds", total)
	case total < 3600:
		if s := total % 60; s != 0 {
			return fmt.Sprintf("~%dm %ds", total/60, s)
		}
		return fmt.Sprintf("~%dm", total/60)
	case total < 86400:
		if m := (total % 3600) / 60; m != 0 {
			return fmt.Sprintf("~%dh %dm", total/3600, m)
		}
		return fmt.Sprintf("~%dh", total/3600)
	default:
		if h := (total % 86400) / 3600; h != 0 {
			return fmt.Sprintf("~%dd %dh", total/86400, h)
		}
		return fmt.Sprintf("~%dd", total/86400)
	}
}

// joinAnd joins parts as "a", "a and b", "a, b and c".
func joinAnd(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
}

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
		parts = append(parts, amountText(resources[k])+" "+resourceLabel(k))
	}
	if workers > 0 {
		noun := "workers"
		if workers == 1 {
			noun = "worker"
		}
		parts = append(parts, fmt.Sprintf("%d %s", workers, noun))
	}
	return parts
}
