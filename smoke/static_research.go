package smoke

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The Research Covenant: what the tech tree's lock may ask of an age, from
// config alone. Each age's wonder needs its keystone tech, and the advance
// needs the wonder, so a keystone is a gate like any other and gets the same
// kind of check. Knowledge is no gate's requirement any more, so nothing but
// research forces a player to hold it.
//
//   - Entry: the first tech of an age (its cheapest) fits the knowledge
//     storage a player enters the age with, or does once one copy of the
//     age's own storage building stands (TechEntryStorageCopies; the
//     Renaissance, priced at the tree's full share, gets
//     TechEntryStorageCopiesFull).
//     "Enters with" is the least anyone can: what the gate into the age
//     forces (ladderForced: storage is never lost, and every storage
//     building raises every cap alike), or the priciest tech an earlier
//     wonder made them pay for, whichever is more. The storage ladder
//     already promises the age's first storage copy can be paid for from
//     there.
//   - Storage: every tech the wonder waits for fits, with
//     GateResourceMargin to spare, in the most knowledge storage buildable
//     in the age.
//   - Time: what the wonder waits for is affordable well inside the age's
//     target. The knowledge it costs, as a share of what the age makes in
//     its target time (config.AgeKnowledge), plus the time to research it,
//     plus the time to build the wonder, all as shares of the target, come
//     to at most 1. It is counted for a run that researched nothing an
//     earlier wonder did not ask for, so the spine techs of earlier ages a
//     keystone stands on are on this age's bill: the worst case.
//
// (That the keystone can be reached in its age at all is the Gate
// Covenant's tech_locked check.)
const (
	// TechEntryStorageCopies is how many copies of an age's storage building
	// the Entry rule allows before the age's first tech must fit. One: the
	// copy the storage ladder guarantees.
	TechEntryStorageCopies = 1
	// TechEntryStorageCopiesFull is the same allowance for an age whose
	// research is priced at the tree's full share
	// (config.ResearchBudgetShareDesign) while the tree is a third of its
	// size: the Renaissance, whose first tech then takes four Renaissance
	// Vaults. (Its old gate asked for 30M knowledge held at once: ten.)
	TechEntryStorageCopiesFull = 4
)

// entryCopiesAllowed is the Entry rule's allowance for age.
func entryCopiesAllowed(age string) int {
	if config.ResearchBudgetShareOf(age) >= config.ResearchBudgetShareDesign {
		return TechEntryStorageCopiesFull
	}
	return TechEntryStorageCopies
}

// ResearchRow is one age against the Research Covenant.
type ResearchRow struct {
	Age string `json:"age"`
	// First is the age's cheapest tech ("" for an age with none).
	First     string  `json:"first_tech,omitempty"`
	FirstCost float64 `json:"first_cost,omitempty"`
	// EntryStorage is the least knowledge storage anyone enters the age
	// with, and ForcedBy what forces it.
	EntryStorage float64 `json:"entry_storage"`
	ForcedBy     string  `json:"forced_by,omitempty"`
	// StorageCopies is the copies of the age's storage building it takes
	// before First fits: 0 when it fits on entry, -1 when it never does.
	StorageCopies int `json:"storage_copies"`
	// Keystone is the tech the age's wonder needs ("" for none), and Chain
	// what a run that researched nothing an earlier wonder did not ask for
	// still has to research for it, what the keystone stands on first.
	Keystone string   `json:"keystone,omitempty"`
	Wonder   string   `json:"wonder,omitempty"`
	Chain    []string `json:"chain,omitempty"`
	// ChainCost is the chain's knowledge, ChainMax its priciest tech, and
	// MaxStorage the most knowledge storage buildable in the age.
	ChainCost  float64 `json:"chain_cost,omitempty"`
	ChainMax   float64 `json:"chain_max,omitempty"`
	MaxStorage float64 `json:"max_storage,omitempty"`
	// The Time rule's three parts, as shares of the age: the chain's cost
	// over what the age makes, the chain's research time and the wonder's
	// build time over the target.
	KnowledgeShare float64 `json:"knowledge_share,omitempty"`
	ResearchShare  float64 `json:"research_share,omitempty"`
	BuildShare     float64 `json:"build_share,omitempty"`
}

// TimeShare is the share of the age's target the Time rule counts: the
// knowledge, the research and the wonder's construction.
func (r ResearchRow) TimeShare() float64 {
	return r.KnowledgeShare + r.ResearchShare + r.BuildShare
}

// Problems lists what the row breaks of the covenant, one line each (nil
// when it keeps it).
func (r ResearchRow) Problems() []string {
	var out []string
	if allowed := entryCopiesAllowed(r.Age); r.First != "" && (r.StorageCopies < 0 || r.StorageCopies > allowed) {
		need := fmt.Sprintf("%d copies of the age's storage building", r.StorageCopies)
		if r.StorageCopies < 0 {
			need = "more storage than the age can build"
		}
		out = append(out, fmt.Sprintf("%s: its first tech, %s, costs %s knowledge, and a player may enter the age with %s of knowledge storage (%s): it takes %s before the tech fits (at most %d allowed)",
			r.Age, r.First, num(r.FirstCost), num(r.EntryStorage), r.ForcedBy, need, allowed))
	}
	if r.Keystone == "" {
		return out
	}
	if r.ChainMax*GateResourceMargin > r.MaxStorage {
		out = append(out, fmt.Sprintf("%s: %s waits for a tech that costs %s knowledge, over 1/%g of the %s knowledge storage buildable in the age",
			r.Age, r.Wonder, num(r.ChainMax), GateResourceMargin, num(r.MaxStorage)))
	}
	if r.TimeShare() > 1 {
		out = append(out, fmt.Sprintf("%s: %s waits for %s (%s knowledge, %.0f%% of what the age makes in its target time); with %.0f%% of the target to research it and %.0f%% to build the wonder that is %.0f%% of the age, over 100%%",
			r.Age, r.Wonder, strings.Join(r.Chain, ", "), num(r.ChainCost), 100*r.KnowledgeShare, 100*r.ResearchShare, 100*r.BuildShare, 100*r.TimeShare()))
	}
	return out
}

// StaticResearch checks every age against the Research Covenant and returns
// one row per age, failing or not.
func StaticResearch() []ResearchRow {
	return staticResearch(config.Ages(), config.BuildingByKey(), config.Technologies(), config.AgeKnowledge)
}

// staticResearch is StaticResearch over the given tables; knowledge is what
// an age makes in its target time.
func staticResearch(ages []config.AgeDef, defs map[string]config.BuildingDef, techs []config.TechDef, knowledge func(age string) float64) []ResearchRow {
	byKey := make(map[string]config.TechDef, len(techs))
	for _, t := range techs {
		byKey[t.Key] = t
	}
	base := 0.0
	for _, r := range config.BaseResources() {
		if r.Key == "knowledge" {
			base = r.BaseStorage
		}
	}
	// The most a resource's base storage is over knowledge's: the gate's
	// biggest price sits under that resource's cap, and knowledge's cap is
	// lower by the difference.
	lead := 0.0
	for _, r := range config.BaseResources() {
		lead = math.Max(lead, r.BaseStorage-base)
	}

	have := map[string]bool{} // what a keystones-only run has researched so far
	held := 0.0               // the priciest tech it has paid for
	var out []ResearchRow
	for i, a := range ages {
		row := ResearchRow{Age: a.Key, EntryStorage: base, ForcedBy: "the starting store"}
		if i > 0 {
			if forced, by := ladderForced(ages[i-1], a, defs); forced-lead > row.EntryStorage {
				row.EntryStorage, row.ForcedBy = forced-lead, by
			}
		}
		if held > row.EntryStorage {
			row.EntryStorage, row.ForcedBy = held, "an earlier wonder's tech"
		}
		for _, t := range techs {
			if t.Age == a.Key && (row.First == "" || t.Cost < row.FirstCost) {
				row.First, row.FirstCost = t.Key, t.Cost
			}
		}
		if row.First != "" {
			row.StorageCopies = entryCopies(a.Key, row.FirstCost, row.EntryStorage, defs)
		}

		wonder := ageWonder(a, defs)
		if key := defs[wonder].RequiredTech; key != "" {
			row.Keystone, row.Wonder = key, wonder
			row.Chain = configChain(byKey, key, have, nil)
			ticks := 0
			for _, k := range row.Chain {
				t := byKey[k]
				row.ChainCost += t.Cost
				row.ChainMax = math.Max(row.ChainMax, t.Cost)
				ticks += t.ResearchTicks
			}
			held = math.Max(held, row.ChainMax)
			row.MaxStorage = maxStorageIn(defs, a.Key, "knowledge")
			if made := knowledge(a.Key); made > 0 {
				row.KnowledgeShare = row.ChainCost / made
			} else {
				row.KnowledgeShare = math.Inf(1)
			}
			if target := config.AgeTargetTicks(a.Key); target > 0 {
				row.ResearchShare = float64(ticks) / target
				row.BuildShare = float64(defs[wonder].BuildTicks) / target
			}
		}
		out = append(out, row)
	}
	return out
}

// entryCopies is how many copies of age's storage building it takes before
// cost of knowledge fits a store of entry: 0 when it fits already, -1 when
// the age's storage never gets there.
func entryCopies(age string, cost, entry float64, defs map[string]config.BuildingDef) int {
	if cost <= entry {
		return 0
	}
	per, most := 0.0, 0
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if d.Category != "storage" || d.RequiredAge != age {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "storage" && (e.Target == "all" || e.Target == "knowledge") && e.Value > per {
				per, most = e.Value, d.MaxCount
			}
		}
	}
	if per <= 0 {
		return -1
	}
	n := int(math.Ceil((cost - entry) / per))
	if most > 0 && n > most {
		return -1
	}
	return n
}

// configChain appends key and every tech it still needs to out, what it
// needs first, and marks them in have: every prerequisite not in have, all
// the way down, and for an either-or group have does not satisfy, the
// branch that costs the least to finish.
func configChain(techs map[string]config.TechDef, key string, have map[string]bool, out []string) []string {
	t, ok := techs[key]
	if !ok || have[key] {
		return out
	}
	have[key] = true
	for _, pre := range t.Prerequisites {
		out = configChain(techs, pre, have, out)
	}
	if len(t.AnyOf) > 0 && !t.AnyOfMet(func(k string) bool { return have[k] }) {
		best, bestCost := "", math.Inf(1)
		for _, alt := range t.AnyOf {
			trial := make(map[string]bool, len(have))
			for k := range have {
				trial[k] = true
			}
			cost := 0.0
			for _, k := range configChain(techs, alt, trial, nil) {
				cost += techs[k].Cost
			}
			if _, known := techs[alt]; known && cost < bestCost {
				best, bestCost = alt, cost
			}
		}
		if best != "" {
			out = configChain(techs, best, have, out)
		}
	}
	return append(out, key)
}

// writeResearch renders the Research Covenant check.
func writeResearch(sb *strings.Builder, rows []ResearchRow) {
	fmt.Fprintf(sb, "Each age's wonder needs its keystone tech (the Research Covenant). The first tech of an age must fit the knowledge storage a player enters the age with, or do so once %d of the age's own storage building stands (%d in an age priced at the tree's full share: the Renaissance). Every tech the wonder waits for must fit, with %gx to spare, in the most knowledge storage buildable in the age. And what the wonder waits for must be affordable well inside the age: its knowledge as a share of what the age makes in its target time (%.0f%% of that is the age's whole research budget), plus its research time and the wonder's build time as shares of the target, may not pass 100%%. The chain is counted for a run that researched nothing an earlier wonder did not ask for. `go test ./smoke` fails on any row marked ✗.\n\n",
		TechEntryStorageCopies, TechEntryStorageCopiesFull, GateResourceMargin, 100*config.ResearchBudgetShare)
	sb.WriteString("| age | first tech | costs | enters with | storage copies first | the wonder waits for | costs | of the age's knowledge | + research | + build | of the age | |\n|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		mark := "✓"
		if len(r.Problems()) > 0 {
			mark = "✗"
		}
		first, firstCost, copies := "-", "-", "-"
		if r.First != "" {
			first, firstCost, copies = r.First, num(r.FirstCost), fmt.Sprint(r.StorageCopies)
			if r.StorageCopies < 0 {
				copies = "never"
			}
		}
		if r.Keystone == "" {
			fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | - | - | - | - | - | - | %s |\n", r.Age, first, firstCost, num(r.EntryStorage), copies, mark)
			continue
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %.0f%% | %.0f%% | %.0f%% | %.0f%% | %s |\n", r.Age, first, firstCost, num(r.EntryStorage), copies,
			strings.Join(r.Chain, ", "), num(r.ChainCost), 100*r.KnowledgeShare, 100*r.ResearchShare, 100*r.BuildShare, 100*r.TimeShare(), mark)
	}
}
