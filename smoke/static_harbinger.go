package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/rules"
)

// The harbinger price check: every Appease and Brace price, both levels, must
// fit the most storage a player can have in the age the harbinger arrives in
// (MaxStorage), since a thread's price is set then and stays the same in
// every age it lives through, where storage is only larger. Every age a
// harbinger can arrive in is checked (HarbingerPrices), for a doom's thread
// and for the Last Passage's. Appease is priced on what the age produces at
// its pacing target (the share of config.FlowIncome × config.AgeTargetTicks
// that the thread's shortest warning makes), so a longer curve raises it
// while faith and culture storage stay as typed; this is the check that
// catches a curve change outgrowing storage. Level 2 costs double level 1 for
// both answers.
//
// The Appease rule check (StaticAppeaseRules) holds each level-1 Appease
// price to the warning it is priced on: a moderate economy
// (config.FlowIncome) must make it inside the thread's shortest warning, a
// fifth of the arrival age for a doom and the whole of it for the Last
// Passage, and the Last Passage's must cost more than a doom's foretold in
// the same age.

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

// The rules an AppeaseRuleProblem can break.
const (
	// AppeaseRuleWarning: level 1 costs more than a moderate economy makes
	// in the thread's shortest warning.
	AppeaseRuleWarning = "warning"
	// AppeaseRuleOrdinary: the Last Passage's level 1 does not cost more than
	// a doom's foretold in the same age.
	AppeaseRuleOrdinary = "ordinary"
)

// AppeaseRuleProblem is one level-1 Appease price that breaks a pricing rule.
type AppeaseRuleProblem struct {
	Epoch string `json:"epoch"`
	// Age is the age the thread's harbinger arrives in.
	Age         string  `json:"age"`
	LastPassage bool    `json:"last_passage"`
	Rule        string  `json:"rule"`
	Resource    string  `json:"resource"`
	Price       float64 `json:"price"`
	// Limit is what the price is held to: what a moderate economy makes in
	// the shortest warning (AppeaseRuleWarning, at most), or a doom's price
	// in the same age (AppeaseRuleOrdinary, more than).
	Limit float64 `json:"limit"`
}

// thread names the thread the problem is in.
func (p AppeaseRuleProblem) thread() string {
	if p.LastPassage {
		return fmt.Sprintf("the Last Passage foretold in %s", p.Age)
	}
	return fmt.Sprintf("a doom of %s foretold in %s", p.Epoch, p.Age)
}

// String says what is wrong in plain words.
func (p AppeaseRuleProblem) String() string {
	if p.Rule == AppeaseRuleOrdinary {
		return fmt.Sprintf("%s asks %s %s for Appease level 1, no more than a doom's %s there", p.thread(), num(p.Price), p.Resource, num(p.Limit))
	}
	return fmt.Sprintf("%s asks %s %s for Appease level 1, more than the %s a moderate economy makes in its shortest warning", p.thread(), num(p.Price), p.Resource, num(p.Limit))
}

// StaticAppeaseRules checks every thread's level-1 Appease price against the
// warning it is priced on, for every age its harbinger can arrive in, on the
// core ruleset's incomes (the ones HarbingerPrices is priced on).
func StaticAppeaseRules() []AppeaseRuleProblem {
	return appeaseRuleProblems(HarbingerPrices(), rules.Core().FlowIncome)
}

// appeaseRuleProblems is StaticAppeaseRules over rows and income (the
// per-tick income of a resource in an age at a moderate economy): the check's
// broken-number tests feed in altered prices.
func appeaseRuleProblems(rows []PriceRow, income func(res, age string) float64) []AppeaseRuleProblem {
	// A doom's level-1 price by the age its harbinger arrives in.
	doom := map[string]map[string]float64{}
	for _, r := range rows {
		if r.TargetEpoch != "" {
			doom[r.Age] = r.AppeaseL1
		}
	}
	var out []AppeaseRuleProblem
	for _, r := range rows {
		lastPassage := r.TargetEpoch == ""
		add := func(rule, res string, price, limit float64) {
			out = append(out, AppeaseRuleProblem{Epoch: r.Epoch, Age: r.Age, LastPassage: lastPassage, Rule: rule, Resource: res, Price: price, Limit: limit})
		}
		for _, res := range sortedKeys(r.AppeaseL1) {
			if made := float64(income(res, r.Age) * r.WarningTicks); r.AppeaseL1[res] > made {
				add(AppeaseRuleWarning, res, r.AppeaseL1[res], made)
			}
		}
		if !lastPassage {
			continue
		}
		// Every resource a doom's Appease asks for there, the Last Passage's
		// asks more of (one it left out costs 0, and is caught).
		for _, res := range sortedKeys(doom[r.Age]) {
			if ordinary := doom[r.Age][res]; r.AppeaseL1[res] <= ordinary {
				add(AppeaseRuleOrdinary, res, r.AppeaseL1[res], ordinary)
			}
		}
	}
	return out
}

// writeHarbingerPrices renders both checks for the static scenario.
func writeHarbingerPrices(sb *strings.Builder, problems []HarbingerPriceProblem, rules []AppeaseRuleProblem) {
	sb.WriteString("Every Appease and Brace price, both levels, against the most storage buildable in the age the harbinger arrives in, for every age one can arrive in (a thread's price is set then and is the same in every age it lives through).\n\n")
	if len(problems) == 0 {
		sb.WriteString("No problems.\n")
	} else {
		sb.WriteString("| era | answer | price | max storage in |\n|---|---|---|---|\n")
		for _, p := range problems {
			fmt.Fprintf(sb, "| %s | %s | %s %s | %s (%s) |\n", p.Epoch, p.Answer, num(p.Price), p.Resource, num(p.MaxStorage), p.Age)
		}
	}
	sb.WriteString("\nEvery level-1 Appease price against the warning it is priced on: a moderate economy (config.FlowIncome) must make it inside the thread's shortest warning (a fifth of the arrival age for a doom, the whole of it for the Last Passage), and the Last Passage's must cost more than a doom's foretold in the same age.\n\n")
	if len(rules) == 0 {
		sb.WriteString("No problems.\n")
		return
	}
	sb.WriteString("| thread | rule | price | held to |\n|---|---|---|---|\n")
	for _, p := range rules {
		held := "at most " + num(p.Limit)
		if p.Rule == AppeaseRuleOrdinary {
			held = "more than " + num(p.Limit)
		}
		fmt.Fprintf(sb, "| %s | %s | %s %s | %s |\n", p.thread(), p.Rule, num(p.Price), p.Resource, held)
	}
}
