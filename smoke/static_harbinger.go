package smoke

import (
	"fmt"
	"strings"
)

// The harbinger price check: every Appease and Brace price, both levels, must
// fit the most storage a player can have in the age the harbinger arrives in
// (MaxStorage), since a thread's price is set then and stays the same in
// every age it lives through, where storage is only larger. Every age a
// harbinger can arrive in is checked (HarbingerPrices). Appease is priced on
// what the age produces at its pacing target (a doom's thread: the shortest
// warning's share of config.FlowIncome × config.AgeTargetTicks; the Last
// Passage's: a share of its era's), so a longer curve raises it while faith
// and culture storage stay as typed; this is the check that catches a curve
// change outgrowing storage. Level 2 costs double level 1 for both answers.

// HarbingerPriceProblem is one price no storage in its arrival age holds.
type HarbingerPriceProblem struct {
	Epoch      string  `json:"epoch"`
	Answer     string  `json:"answer"` // "appease 1", "appease 2", "brace 1", "brace 2"
	Resource   string  `json:"resource"`
	Price      float64 `json:"price"`
	Age        string  `json:"age"`
	MaxStorage float64 `json:"max_storage"`
}

// StaticHarbingerPrices checks every thread's harbinger prices against the
// most storage buildable in the age its harbinger arrives in.
func StaticHarbingerPrices() []HarbingerPriceProblem {
	var out []HarbingerPriceProblem
	for _, r := range HarbingerPrices() {
		check := func(answer string, cost map[string]float64, level float64) {
			if r.TargetEpoch == "" {
				answer += " (Last Passage)"
			}
			for _, res := range sortedKeys(cost) {
				price := float64(cost[res] * level)
				if m := MaxStorage(r.Age, res); price > m {
					out = append(out, HarbingerPriceProblem{Epoch: r.Epoch, Answer: answer, Resource: res, Price: price, Age: r.Age, MaxStorage: m})
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
	sb.WriteString("Every Appease and Brace price, both levels, against the most storage buildable in the age the harbinger arrives in, for every age one can arrive in (a thread's price is set then and is the same in every age it lives through).\n\n")
	if len(problems) == 0 {
		sb.WriteString("No problems.\n")
		return
	}
	sb.WriteString("| era | answer | price | max storage in |\n|---|---|---|---|\n")
	for _, p := range problems {
		fmt.Fprintf(sb, "| %s | %s | %s %s | %s (%s) |\n", p.Epoch, p.Answer, num(p.Price), p.Resource, num(p.MaxStorage), p.Age)
	}
}
