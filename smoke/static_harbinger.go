package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
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
// catches a curve change outgrowing storage. Level 2 is checked at its own
// price, the same again as level 1.
//
// The Appease rule check (StaticAppeaseRules) holds each level-1 Appease
// price to the warning it is priced on: a moderate economy
// (config.FlowIncome) must make it inside the thread's shortest warning, a
// fifth of the arrival age for a doom and two thirds of it for the Last
// Passage, and the Last Passage's must cost more than a doom's foretold in
// the same age.
//
// The Brace rule check (StaticBraceRules) holds every thread's level-1
// Brace, in every era, to real effort: it must take a moderate economy
// (config.TypicalIncome) at least BraceMinWarningShare of the thread's
// shortest warning, and in the final era (the Last Passage's, which guards
// the run's points, and its fated doom's) at least WarningBraceMinHours of
// income; that economy must make every resource it asks for inside the
// warning; and every one must be a material the moderate economy's own
// buildings make by the arrival age, not one the age buys at the market or
// only a wonder trickles out.

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
		check := func(answer string, cost map[string]float64) {
			if r.TargetEpoch == "" {
				answer += " (Last Passage)"
			}
			for _, res := range sortedKeys(cost) {
				if m := MaxStorage(r.Age, res); cost[res] > m {
					out = append(out, HarbingerPriceProblem{Epoch: r.Epoch, Answer: answer, Resource: res, Price: cost[res], Age: r.Age, MaxStorage: m})
				}
			}
		}
		check("appease 1", r.AppeaseL1)
		check("appease 2", r.AppeaseL2)
		check("brace 1", r.BraceL1)
		check("brace 2", r.BraceL2)
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

// WarningBraceMinHours is the least a level-1 Brace of the final era's two
// threads may cost: the hours at 1x a moderate economy needs to make the
// slowest of its resources. Both were once a share of the era's largest age
// requirement, which the Interstellar Age makes in under 7 ticks.
const WarningBraceMinHours = 10.0

// BraceMinWarningShare is the least a level-1 Brace of any thread may cost,
// as the share of the thread's shortest warning a moderate economy needs to
// make the slowest of its resources. An ordinary doom's is priced at a
// third of it; before that it was a share of what the era's advances ask,
// which later ages make in seconds (the Electric Era's: 2 seconds of the
// Atomic Age's income).
const BraceMinWarningShare = 0.3

// The rules a BraceRuleProblem can break.
const (
	// BraceRuleEffort: level 1 takes a moderate economy less than
	// BraceMinWarningShare of the warning, or in the final era less than
	// WarningBraceMinHours of income.
	BraceRuleEffort = "effort"
	// BraceRuleWarning: level 1 asks more of a resource than a moderate
	// economy makes in the thread's shortest warning.
	BraceRuleWarning = "warning"
	// BraceRuleMaterial: level 1 asks for a resource no building of the
	// moderate economy makes by the arrival age.
	BraceRuleMaterial = "material"
)

// BraceRuleProblem is one level-1 Brace price that breaks a pricing rule.
type BraceRuleProblem struct {
	Epoch string `json:"epoch"`
	// Age is the age the thread's harbinger arrives in.
	Age         string `json:"age"`
	LastPassage bool   `json:"last_passage"`
	Rule        string `json:"rule"`
	// Resource is the one the rule broke on: for BraceRuleEffort the slowest
	// to make ("" when Brace asks for nothing).
	Resource string  `json:"resource"`
	Price    float64 `json:"price"`
	// Limit is what the price is held to: what a moderate economy makes of
	// the resource in the least time the price may take (BraceRuleEffort,
	// at least), or in the shortest warning (BraceRuleWarning, at most); 0
	// for BraceRuleMaterial.
	Limit float64 `json:"limit"`
	// Hours is the least time the price may take, in hours at 1x
	// (BraceRuleEffort).
	Hours float64 `json:"hours,omitempty"`
}

// thread names the thread the problem is in.
func (p BraceRuleProblem) thread() string {
	if p.LastPassage {
		return fmt.Sprintf("the Last Passage foretold in %s", p.Age)
	}
	return fmt.Sprintf("a doom of %s foretold in %s", p.Epoch, p.Age)
}

// String says what is wrong in plain words.
func (p BraceRuleProblem) String() string {
	switch {
	case p.Rule == BraceRuleMaterial:
		return fmt.Sprintf("%s asks %s %s for Brace level 1, which no building of a moderate economy makes by then", p.thread(), num(p.Price), p.Resource)
	case p.Rule == BraceRuleWarning:
		return fmt.Sprintf("%s asks %s %s for Brace level 1, more than the %s a moderate economy makes in its shortest warning", p.thread(), num(p.Price), p.Resource, num(p.Limit))
	case p.Resource == "":
		return fmt.Sprintf("%s asks nothing for Brace level 1; it must cost at least %.3g hours of a moderate economy's income", p.thread(), p.Hours)
	}
	return fmt.Sprintf("%s asks %s %s for Brace level 1, its dearest part, under the %s a moderate economy makes in %.3g hours", p.thread(), num(p.Price), p.Resource, num(p.Limit), p.Hours)
}

// StaticBraceRules checks every thread's level-1 Brace price, for every
// age its harbinger can arrive in, on the core ruleset's incomes (the ones
// HarbingerPrices is priced on).
func StaticBraceRules() []BraceRuleProblem {
	return braceRuleProblems(HarbingerPrices(), rules.Core().TypicalIncome, rules.Core().BuildingOutput)
}

// braceMinTicks is the least a row's level-1 Brace may take a moderate
// economy, in ticks at 1x: BraceMinWarningShare of its warning, and in the
// final era WarningBraceMinHours at least.
func braceMinTicks(r PriceRow) float64 {
	ticks := float64(BraceMinWarningShare * r.WarningTicks)
	if r.FinalEra {
		ticks = math.Max(ticks, WarningBraceMinHours*3600/config.TickSeconds)
	}
	return ticks
}

// braceRuleProblems is StaticBraceRules over rows, income (the per-tick
// income of a resource in an age at a moderate economy) and built (what
// that economy's own buildings make of it by then): the check's
// broken-number tests feed in altered prices.
func braceRuleProblems(rows []PriceRow, income, built func(res, age string) float64) []BraceRuleProblem {
	var out []BraceRuleProblem
	for _, r := range rows {
		minTicks := braceMinTicks(r)
		hours := minTicks * config.TickSeconds / 3600
		add := func(rule, res string, price, limit float64) {
			p := BraceRuleProblem{Epoch: r.Epoch, Age: r.Age, LastPassage: r.TargetEpoch == "", Rule: rule, Resource: res, Price: price, Limit: limit}
			if rule == BraceRuleEffort {
				p.Hours = hours
			}
			out = append(out, p)
		}
		// The price takes as long as its slowest resource: that one must take
		// the time. A resource no building makes breaks the material rule.
		slowest, slowestTicks := "", 0.0
		for _, res := range sortedKeys(r.BraceL1) {
			if built(res, r.Age) <= 0 {
				add(BraceRuleMaterial, res, r.BraceL1[res], 0)
				continue
			}
			rate := income(res, r.Age)
			if made := float64(rate * r.WarningTicks); r.BraceL1[res] > made {
				add(BraceRuleWarning, res, r.BraceL1[res], made)
			}
			if rate <= 0 {
				continue
			}
			if ticks := r.BraceL1[res] / rate; slowest == "" || ticks > slowestTicks {
				slowest, slowestTicks = res, ticks
			}
		}
		switch {
		case len(r.BraceL1) == 0:
			add(BraceRuleEffort, "", 0, 0)
		case slowest != "" && slowestTicks < minTicks:
			add(BraceRuleEffort, slowest, r.BraceL1[slowest], float64(income(slowest, r.Age)*minTicks))
		}
	}
	return out
}

// writeHarbingerPrices renders the three checks for the static scenario.
func writeHarbingerPrices(sb *strings.Builder, problems []HarbingerPriceProblem, rules []AppeaseRuleProblem, brace []BraceRuleProblem) {
	sb.WriteString("Every Appease and Brace price, both levels, against the most storage buildable in the age the harbinger arrives in, for every age one can arrive in (a thread's price is set then and is the same in every age it lives through).\n\n")
	if len(problems) == 0 {
		sb.WriteString("No problems.\n")
	} else {
		sb.WriteString("| era | answer | price | max storage in |\n|---|---|---|---|\n")
		for _, p := range problems {
			fmt.Fprintf(sb, "| %s | %s | %s %s | %s (%s) |\n", p.Epoch, p.Answer, num(p.Price), p.Resource, num(p.MaxStorage), p.Age)
		}
	}
	sb.WriteString("\nEvery level-1 Appease price against the warning it is priced on: a moderate economy (config.FlowIncome) must make it inside the thread's shortest warning (a fifth of the arrival age for a doom, two thirds of it for the Last Passage), and the Last Passage's must cost more than a doom's foretold in the same age.\n\n")
	if len(rules) == 0 {
		sb.WriteString("No problems.\n")
	} else {
		sb.WriteString("| thread | rule | price | held to |\n|---|---|---|---|\n")
		for _, p := range rules {
			held := "at most " + num(p.Limit)
			if p.Rule == AppeaseRuleOrdinary {
				held = "more than " + num(p.Limit)
			}
			fmt.Fprintf(sb, "| %s | %s | %s %s | %s |\n", p.thread(), p.Rule, num(p.Price), p.Resource, held)
		}
	}
	fmt.Fprintf(sb, "\nEvery level-1 Brace price, for every age the harbinger can arrive in: it must take a moderate economy (config.TypicalIncome) at least %g of the thread's shortest warning (and in the final era, for the Last Passage and its fated doom, at least %g hours of income), that economy must make each resource it asks for inside the warning, and each must be a material its own buildings make by then.\n\n", BraceMinWarningShare, WarningBraceMinHours)
	if len(brace) == 0 {
		sb.WriteString("No problems.\n")
		return
	}
	sb.WriteString("| thread | rule | price | held to |\n|---|---|---|---|\n")
	for _, p := range brace {
		held := "at most " + num(p.Limit)
		switch p.Rule {
		case BraceRuleEffort:
			held = "at least " + num(p.Limit)
		case BraceRuleMaterial:
			held = "a material the age makes"
		}
		fmt.Fprintf(sb, "| %s | %s | %s %s | %s |\n", p.thread(), p.Rule, num(p.Price), p.Resource, held)
	}
}
