package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The harbinger price check: every Appease and Brace price, both levels, must
// fit the most storage a player can have in the first age of its era
// (MaxStorage), since the price is the same in every age of the thread.
// Appease is priced on what the era's ages produce at their pacing targets
// (config.FlowIncome × config.AgeTargetTicks), so a longer curve raises it
// while faith and culture storage stay as typed; this is the check that
// catches a curve change outgrowing storage. Level 2 costs double level 1
// for both answers.

// HarbingerPriceProblem is one price no storage in its era's first age holds.
type HarbingerPriceProblem struct {
	Epoch      string  `json:"epoch"`
	Answer     string  `json:"answer"` // "appease 1", "appease 2", "brace 1", "brace 2"
	Resource   string  `json:"resource"`
	Price      float64 `json:"price"`
	Age        string  `json:"age"`
	MaxStorage float64 `json:"max_storage"`
}

// StaticHarbingerPrices checks every era's harbinger prices against the most
// storage buildable in its first age.
func StaticHarbingerPrices() []HarbingerPriceProblem {
	var out []HarbingerPriceProblem
	epochs := config.EpochByKey()
	for _, r := range HarbingerPrices() {
		first := epochs[r.Epoch].Ages[0]
		check := func(answer string, cost map[string]float64, level float64) {
			for _, res := range sortedKeys(cost) {
				price := float64(cost[res] * level)
				if m := MaxStorage(first, res); price > m {
					out = append(out, HarbingerPriceProblem{Epoch: r.Epoch, Answer: answer, Resource: res, Price: price, Age: first, MaxStorage: m})
				}
			}
		}
		check("appease 1", r.AppeaseL1, 1)
		check("appease 2", r.AppeaseL1, 2)
		check("brace 1", r.BraceL1, 1)
		check("brace 2", r.BraceL1, 2)
	}
	return out
}

// writeHarbingerPrices renders the check for the static scenario.
func writeHarbingerPrices(sb *strings.Builder, problems []HarbingerPriceProblem) {
	sb.WriteString("Every Appease and Brace price, both levels, against the most storage buildable in the first age of its era (the price is the same in every age of the thread).\n\n")
	if len(problems) == 0 {
		sb.WriteString("No problems.\n")
		return
	}
	sb.WriteString("| era | answer | price | max storage in |\n|---|---|---|---|\n")
	for _, p := range problems {
		fmt.Fprintf(sb, "| %s | %s | %s %s | %s (%s) |\n", p.Epoch, p.Answer, num(p.Price), p.Resource, num(p.MaxStorage), p.Age)
	}
}
