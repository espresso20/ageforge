package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The depth check (Pacing v2, PR 6), from config alone:
//
//   - every age weighs 3^epoch (1 per Stone Era age up to 729 per Cosmic
//     Era age), and a prestige pays the weights of the ages before it, with
//     no divisor: the table site/docs/prestige.md quotes;
//   - a second early taste right after a full run gains under
//     DepthTasteMax of what that run paid (9 of 120): resetting early again
//     is never worth it;
//   - prestige opens at the Medieval Age, and a taste there pays for the
//     first kit item.

// DepthTasteMax is the most a Medieval Age prestige may add, as a share of
// a Modern Age run's points.
const DepthTasteMax = 0.10

// depthDocTable is the points table the docs quote.
var depthDocTable = []struct {
	age    string
	points int
}{
	{"medieval_age", 9}, {"modern_age", 120}, {"information_age", 201},
	{"cyberpunk_age", 363}, {"interstellar_age", 1092},
}

// StaticDepth checks the depth weights, the docs' table, the taste's share
// and the kit's first price. It returns what breaks.
func StaticDepth() []string {
	var out []string
	for _, ep := range config.Epochs() {
		w := 1
		for i := 0; i < ep.Order; i++ {
			w *= 3
		}
		for _, a := range ep.Ages {
			if got := config.DepthWeight(a); got != w {
				out = append(out, fmt.Sprintf("%s (%s) weighs %d, want 3^%d = %d", a, ep.Key, got, ep.Order, w))
			}
		}
	}
	for _, row := range depthDocTable {
		if got := config.DepthPoints(row.age); got != row.points {
			out = append(out, fmt.Sprintf("a prestige from %s pays %d, the docs say %d", row.age, got, row.points))
		}
	}
	taste, run := config.DepthPoints(game.PrestigeMinAge), config.DepthPoints(game.PrestigeRunAge)
	if run <= 0 || float64(taste) >= float64(DepthTasteMax*float64(run)) {
		out = append(out, fmt.Sprintf("a second taste (%s, %d points) after a full run (%s, %d) gains %.1f%%, want under %.0f%%",
			game.PrestigeMinAge, taste, game.PrestigeRunAge, run, float64(taste*100)/float64(max(run, 1)), DepthTasteMax*100))
	}
	if first := config.PrestigeUpgradeByKey()[config.LegacyKit()[0]]; len(first.Costs) == 0 || first.Costs[0] > taste {
		out = append(out, fmt.Sprintf("the first kit item (%s) costs more than a taste pays (%d)", first.Key, taste))
	}
	return out
}

// writeDepthStatic renders the weights and the points table.
func writeDepthStatic(sb *strings.Builder, problems []string) {
	sb.WriteString("Each completed age pays 3^epoch; a prestige pays the ages before it, with no divisor.\n\n| epoch | weight per age |\n|---|---|\n")
	for _, ep := range config.Epochs() {
		fmt.Fprintf(sb, "| %s | %d |\n", ep.Key, config.DepthWeight(ep.Ages[0]))
	}
	sb.WriteString("\n| prestige from | points |\n|---|---|\n")
	for _, row := range depthDocTable {
		fmt.Fprintf(sb, "| %s | %d |\n", row.age, config.DepthPoints(row.age))
	}
	taste, run := config.DepthPoints(game.PrestigeMinAge), config.DepthPoints(game.PrestigeRunAge)
	fmt.Fprintf(sb, "\nA second taste after a full run adds %d of %d points (%.1f%%, limit %.0f%%).\n", taste, run, float64(taste*100)/float64(max(run, 1)), DepthTasteMax*100)
	for _, p := range problems {
		fmt.Fprintf(sb, "- %s\n", p)
	}
}
