package smoke

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The bot's research order. An age's wonder cannot be built without its
// keystone tech, and the advance needs the wonder, so the bot researches
// what the age cannot be left without before anything else:
//
//  1. the keystone of this age's wonder and everything it stands on, what it
//     stands on first;
//  2. a tech that unlocks something the bot wants to build: the tech behind
//     a building the next age requires, and a tech that is the age's only
//     source of a resource it asks for (Steel Forging, when the Renaissance
//     asks for steel and no Medieval building makes it);
//  3. then the rest, cheapest first, with a tech that makes a missing
//     resource or opens a building of this age ahead of the others.
//
// While a tech of the first two groups is still to start, the bot saves its
// knowledge for it: nothing of the third group starts, the knowledge those
// techs cost is a target of the age like any requirement (so the bot builds
// knowledge producers and storage for it, and buys knowledge at the market
// where it sells), and the wonder bank does not take it.

// techChain appends key and everything it still needs to out, what it needs
// first: every unresearched prerequisite, all the way down, and, for an
// either-or group nothing satisfies yet, the branch that costs the least
// knowledge to finish. A tech already in seen, researched or unknown adds
// nothing.
func techChain(set *rules.Set, techs map[string]game.TechState, key string, seen map[string]bool, out []string) []string {
	t, ok := techs[key]
	if !ok || t.Researched || seen[key] {
		return out
	}
	seen[key] = true
	def, _ := set.Tech(key)
	for _, pre := range def.Prerequisites {
		out = techChain(set, techs, pre, seen, out)
	}
	researched := func(k string) bool { return techs[k].Researched || seen[k] }
	if len(def.AnyOf) > 0 && !def.AnyOfMet(researched) {
		best, bestCost := "", math.Inf(1)
		for _, alt := range def.AnyOf {
			trial := make(map[string]bool, len(seen))
			for k := range seen {
				trial[k] = true
			}
			cost := 0.0
			for _, k := range techChain(set, techs, alt, trial, nil) {
				cost += techs[k].Cost
			}
			if _, known := techs[alt]; known && cost < bestCost {
				best, bestCost = alt, cost
			}
		}
		if best != "" {
			out = techChain(set, techs, best, seen, out)
		}
	}
	return append(out, key)
}

// unblocks reports whether tech key makes a resource the age still asks for
// and nothing produces.
func (b *Bot) unblocks(p *plan, key string) bool {
	return b.unblocked(p, key, false)
}

// unblocked is unblocks; with only set, the resource must also be one no
// building the bot can build in this age makes, so the tech is its only
// source.
func (b *Bot) unblocked(p *plan, key string, only bool) bool {
	def, _ := b.rules.Tech(key)
	for _, e := range def.Effects {
		if e.Kind != config.EffectFlatOutput || p.target[e.Target] <= p.amt[e.Target] || p.st.Resources[e.Target].Rate > 0 {
			continue
		}
		if !only || !b.madeHere(p, e.Target) {
			return true
		}
	}
	return false
}

// madeHere reports whether a building the bot can build in this age makes
// res, whatever it costs.
func (b *Bot) madeHere(p *plan, res string) bool {
	match := producer(res)
	for _, key := range sortedKeys(p.st.Buildings) {
		if !b.buildable(p, key) {
			continue
		}
		for _, e := range b.defs[key].Effects {
			if match(e) > 0 {
				return true
			}
		}
	}
	return false
}

// openers is the techs a building of this age still waits for.
func (b *Bot) openers(p *plan) map[string]bool {
	out := map[string]bool{}
	for bld, bs := range p.st.Buildings {
		if bs.NeedsTech != "" && b.defs[bld].RequiredAge == p.st.Age {
			out[bs.NeedsTech] = true
		}
	}
	return out
}

// mustTechs lists the techs the age cannot be left without that are not
// researched yet, in the order to research them (see the top of this file).
// It reads p.target and p.needBld, so newPlan calls it once those are set.
func (b *Bot) mustTechs(p *plan) []string {
	st := p.st
	techs := st.Research.Techs
	seen := map[string]bool{}
	var out []string
	if w := st.CurrentAgeWonderKey; w != "" {
		if k := st.Buildings[w].NeedsTech; k != "" {
			out = techChain(b.rules, techs, k, seen, out)
		}
	}
	for _, bld := range sortedKeys(p.needBld) {
		if k := st.Buildings[bld].NeedsTech; k != "" {
			out = techChain(b.rules, techs, k, seen, out)
		}
	}
	here := b.ageIdx[st.Age]
	for _, key := range b.byCost(techs) {
		if t := techs[key]; !t.Researched && b.ageIdx[t.Age] <= here && b.unblocked(p, key, true) {
			out = techChain(b.rules, techs, key, seen, out)
		}
	}
	return out
}

// restOrder is the techs in the order the bot takes them once nothing the
// age cannot be left without is waiting: one that makes a missing resource
// first, then one that opens a building of this age, then the cheapest.
func (b *Bot) restOrder(p *plan) []string {
	keys := b.byCost(p.st.Research.Techs)
	opens := b.openers(p)
	rank := make(map[string]int, len(keys))
	for _, key := range keys {
		switch {
		case b.unblocks(p, key):
			rank[key] = 0
		case opens[key]:
			rank[key] = 1
		default:
			rank[key] = 2
		}
	}
	sort.SliceStable(keys, func(i, j int) bool { return rank[keys[i]] < rank[keys[j]] })
	return keys
}

// byCost is the tech keys, cheapest first (by key among equals).
func (b *Bot) byCost(techs map[string]game.TechState) []string {
	keys := sortedKeys(techs)
	sort.SliceStable(keys, func(i, j int) bool { return techs[keys[i]].Cost < techs[keys[j]].Cost })
	return keys
}

// nextMust is the tech of p.must to start next: the first one that is open
// (its age reached, what it needs researched) and not the one in progress.
// "" when none can start.
func (p *plan) nextMust() string {
	for _, key := range p.must {
		if t := p.st.Research.Techs[key]; t.Available && key != p.st.Research.CurrentTech {
			return key
		}
	}
	return ""
}
