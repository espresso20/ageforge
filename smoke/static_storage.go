package smoke

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
)

// The Storage Covenant: storage is a wall, and the wall is where the rule
// puts it. A store sized to hours of income (the covenant's first form) held
// too much in some ages and too little in others; since the storage rule
// (config/storage_rule.go) a store is sized by what the age's buildings cost,
// and the covenant checks the tables against that rule.
//
// It is written out again here, from the tables alone, and does not call
// config.StorageRuleSizes or read any of config/storage_rule.go's constants:
// a check that shared the rule's code would agree with it whatever it said.
// The numbers below are the rule as the owner wrote it down; if the rule
// moves, they move with it, in the same change.
//
// The reference town holds storageCopies copies of every building of the age
// (or the gate's count, where the gate out of the age asks for more). For
// every age with a storage building:
//
//   - the storage building climbs at storageClimb a copy and has no copy
//     limit: the price is the limit;
//   - five copies of it, on top of the earlier ages' stores, hold the dearest
//     price any building of the age asks of the reference town, in any one
//     resource, and a quarter more (storageRoom). Wonders and monuments are
//     left out: a wonder is paid into its bank and may cost more than a store;
//   - each copy up to the fifth fits under the store the copies before it
//     give, with storageMargin to spare;
//   - a town arrives in the next age able to pay the first copy of that age's
//     storage building and of its housing, with storageMargin to spare (an
//     age with no storage building of its own is held to what it asks
//     instead: its dearest price and a quarter, and the first copy of each of
//     its buildings);
//   - the gate out of the age asks for storageCopies copies of its storage
//     building, so every town that passes arrives with the reference store;
//   - the gate asks for gateShare of the store each material needs (the
//     dearest price the age's buildings ask in it, and a quarter more), in
//     every material the age's buildings cost but food, faith, culture and
//     knowledge, to two figures;
//   - the wall: in the first earlyWallAges ages that have storage, the store
//     walls at exactly five copies (a sixth is priced over the store); in
//     the later ages a player reaches at most lateReach times the reference
//     store (the climb is 1.75, but the dearest buildings of the age are not
//     the storage building).
const (
	storageCopies  = 5
	storageRoom    = 1.25
	storageMargin  = 1.05
	storageClimb   = 1.75
	gateShare      = 0.25
	earlyWallAges  = 4
	lateReach      = 2.8
	gateRoundSlack = 0.051 // two figures: the largest relative rounding is 5%
)

// gateFreeResources are the resources no gate asks for: the flow resources
// (food, faith, culture, soldiers, nanobots), which have no price to size a
// store by, and knowledge, which a tech is paid in.
var gateFreeResources = map[string]bool{"food": true, "faith": true, "culture": true, "soldiers": true, "nanobots": true, "knowledge": true}

// StorageProblem is a place the tables part from the storage rule.
type StorageProblem struct {
	Age  string `json:"age"`
	Rule string `json:"rule"` // curve, hold, climb, arrive, gate_copies, gate_amount, wall
	Why  string `json:"why"`
}

func (p StorageProblem) String() string { return fmt.Sprintf("%s: %s", p.Age, p.Why) }

// StorageRow is one age's storage as the tables give it.
type StorageRow struct {
	Age      string  `json:"age"`
	Building string  `json:"building"`
	PerCopy  float64 `json:"per_copy"`
	// Dearest is the dearest price the age asks of the reference town; Store
	// the reference store on leaving the age (five copies on the earlier
	// stores); Reached the store a player reaches by buying every storage
	// building to its wall, and Copies the copies of this age's building that
	// bought.
	Dearest float64 `json:"dearest_ask"`
	Store   float64 `json:"reference_store"`
	Reached float64 `json:"reached_store"`
	Copies  int     `json:"copies"`
}

// Ratio is the store a player reaches over the reference store.
func (r StorageRow) Ratio() float64 { return r.Reached / r.Store }

// StaticStorage checks every age against the storage rule and returns what
// breaks it, with one row per age that has a storage building.
func StaticStorage() ([]StorageProblem, []StorageRow) {
	return staticStorage(config.BuildingByKey(), config.Ages())
}

// storageCost is the price of copy k (1 for the first) of d in res.
func storageCost(d config.BuildingDef, res string, k int) float64 {
	return math.Max(1, math.Floor(float64(d.BaseCost[res]*detmath.Pow(d.CostScale, float64(k-1)))))
}

// storageDearest is the dearest single price of copy k of d, in any resource.
func storageDearest(d config.BuildingDef, k int) float64 {
	top := 0.0
	for _, res := range sortedKeys(d.BaseCost) {
		top = math.Max(top, storageCost(d, res, k))
	}
	return top
}

func storageSize(d config.BuildingDef) float64 {
	per := 0.0
	for _, e := range d.Effects {
		if e.Type == "storage" && e.Target == "all" {
			per += e.Value
		}
	}
	return per
}

func staticStorage(defs map[string]config.BuildingDef, ages []config.AgeDef) ([]StorageProblem, []StorageRow) {
	var problems []StorageProblem
	n := len(ages)
	idx := make(map[string]int, n)
	for i, a := range ages {
		idx[a.Key] = i
	}
	// What the age builds, apart from wonders and monuments.
	byAge := make([][]config.BuildingDef, n)
	for _, k := range sortedKeys(defs) {
		d := defs[k]
		if i, ok := idx[d.RequiredAge]; ok && d.Category != "wonder" && d.Category != "monument" {
			byAge[i] = append(byAge[i], d)
		}
	}
	storageOf := func(i int) (config.BuildingDef, bool) {
		if i < 0 || i >= n {
			return config.BuildingDef{}, false
		}
		for _, d := range byAge[i] {
			if d.Category == "storage" {
				return d, true
			}
		}
		return config.BuildingDef{}, false
	}
	// How many copies the reference town holds.
	count := func(d config.BuildingDef, i int) int {
		if i+1 < n {
			return max(storageCopies, ages[i+1].BuildingReqs[d.Key])
		}
		return storageCopies
	}
	// The dearest price the age asks of the reference town, per resource.
	asks := func(i int) map[string]float64 {
		out := map[string]float64{}
		if i < 0 || i >= n {
			return out
		}
		for _, d := range byAge[i] {
			if d.Category == "storage" {
				continue
			}
			for _, res := range sortedKeys(d.BaseCost) {
				out[res] = math.Max(out[res], storageCost(d, res, count(d, i)))
			}
		}
		return out
	}
	dearest := func(i int) float64 {
		top := 0.0
		for _, v := range asks(i) {
			top = math.Max(top, v)
		}
		return top
	}
	// Every store is wood's: the rule sizes one store for every resource, and
	// a resource with a lower base holds up to 20 less before any building.
	base := 0.0
	for _, r := range config.BaseResources() {
		if r.Key == "wood" {
			base = r.BaseStorage
		}
	}

	// The reference store on leaving each age.
	store := make([]float64, n)
	before := make([]float64, n) // and on arriving in it
	prev := base
	for i := range ages {
		before[i] = prev
		if b, ok := storageOf(i); ok {
			prev += float64(storageCopies * storageSize(b))
		}
		store[i] = prev
	}

	var rows []StorageRow
	withStorage := 0
	for i, age := range ages {
		b, ok := storageOf(i)
		if !ok {
			continue
		}
		withStorage++
		fail := func(rule, format string, args ...any) {
			problems = append(problems, StorageProblem{Age: age.Key, Rule: rule, Why: fmt.Sprintf(format, args...)})
		}
		size := storageSize(b)
		if b.CostScale != storageClimb || b.MaxCount != 0 {
			fail("curve", "%s climbs at %g a copy with a copy limit of %d; storage climbs at %g and has no copy limit (the price is the limit)", b.Key, b.CostScale, b.MaxCount, storageClimb)
		}
		if need := storageRoom * dearest(i); store[i] < need {
			fail("hold", "five %s at %s a copy, on the earlier stores, hold %s; the dearest price the age asks of the reference town is %s and the store must hold it and a quarter more (%s)",
				b.Key, num(size), num(store[i]), num(dearest(i)), num(need))
		}
		for k := 2; k <= storageCopies; k++ {
			if p, held := storageDearest(b, k), before[i]+float64(float64(k-1)*size); storageMargin*p > held {
				fail("climb", "copy #%d of %s costs %s, over the %s the copies before it hold (with %gx to spare)", k, b.Key, num(p), num(held), storageMargin)
			}
		}
		if i+1 < n {
			nb, has := storageOf(i + 1)
			switch {
			case has:
				if p := storageDearest(nb, 1); storageMargin*p > store[i] {
					fail("arrive", "the first %s costs %s, over the %s a town leaving %s holds (with %gx to spare)", nb.Key, num(p), num(store[i]), age.Key, storageMargin)
				}
				cheapest, house := math.Inf(1), ""
				for _, d := range byAge[i+1] {
					if p := storageDearest(d, 1); d.Category == "housing" && p < cheapest {
						cheapest, house = p, d.Key
					}
				}
				if house != "" && storageMargin*cheapest > store[i] {
					fail("arrive", "the first %s costs %s, over the %s a town leaving %s holds (with %gx to spare)", house, num(cheapest), num(store[i]), age.Key, storageMargin)
				}
			default:
				if need := storageRoom * dearest(i+1); store[i] < need {
					fail("arrive", "%s has no storage building and asks %s of the reference town; the store leaving %s (%s) must hold it and a quarter more", ages[i+1].Key, num(dearest(i+1)), age.Key, num(store[i]))
				}
				for _, d := range byAge[i+1] {
					if p := storageDearest(d, 1); storageMargin*p > store[i] {
						fail("arrive", "the first %s costs %s, over the %s a town leaving %s holds (with %gx to spare)", d.Key, num(p), num(store[i]), age.Key, storageMargin)
					}
				}
			}
			if got := ages[i+1].BuildingReqs[b.Key]; got != storageCopies {
				fail("gate_copies", "the gate into %s asks for %d %s, want %d: whoever passes it must arrive with the reference store", ages[i+1].Key, got, b.Key, storageCopies)
			}
		}
		// The gate's resource requirements, below, are checked for every age
		// that has a gate, with or without a storage building.

		stores, copies := walledCopies(defs, age.Key)
		reached := stores["wood"]
		row := StorageRow{Age: age.Key, Building: b.Key, PerCopy: size, Dearest: dearest(i), Store: store[i], Reached: reached, Copies: copies[b.Key]}
		rows = append(rows, row)
		if withStorage <= earlyWallAges {
			if row.Copies != storageCopies {
				fail("wall", "%s walls at %d copies, want exactly %d in the first %d ages (a sixth is priced over the store)", b.Key, row.Copies, storageCopies, earlyWallAges)
			}
		} else if row.Ratio() > lateReach+1e-9 {
			fail("wall", "a player reaches %s in %s, %.2f times the reference store (%s); at most %g times", num(row.Reached), age.Key, row.Ratio(), num(row.Store), lateReach)
		}
	}

	// What each gate asks to hold.
	for i := 0; i+1 < n; i++ {
		to := ages[i+1]
		cost := map[string]bool{}
		for _, d := range byAge[i] {
			if d.Category == "storage" {
				continue
			}
			for res := range d.BaseCost {
				cost[res] = true
			}
		}
		ask := asks(i)
		for _, res := range sortedKeys(to.ResourceReqs) {
			have := to.ResourceReqs[res]
			switch {
			case gateFreeResources[res]:
				problems = append(problems, StorageProblem{Age: ages[i].Key, Rule: "gate_amount", Why: fmt.Sprintf("the gate into %s asks for %s %s; no gate asks for %s", to.Key, num(have), res, res)})
			case !cost[res]:
				problems = append(problems, StorageProblem{Age: ages[i].Key, Rule: "gate_amount", Why: fmt.Sprintf("the gate into %s asks for %s %s, which no building of %s costs", to.Key, num(have), res, ages[i].Key)})
			default:
				want := float64(gateShare * storageRoom * ask[res])
				if math.Abs(have-want) > float64(gateRoundSlack*want) {
					problems = append(problems, StorageProblem{Age: ages[i].Key, Rule: "gate_amount", Why: fmt.Sprintf("the gate into %s asks for %s %s, want %s: a quarter of the store the dearest price (%s) needs, with a quarter more", to.Key, num(have), res, num(want), num(ask[res]))})
				}
			}
		}
		var missing []string
		for res := range cost {
			if _, asked := to.ResourceReqs[res]; !asked && !gateFreeResources[res] {
				missing = append(missing, res)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			problems = append(problems, StorageProblem{Age: ages[i].Key, Rule: "gate_amount", Why: fmt.Sprintf("the gate into %s does not ask for %s, which the buildings of %s cost", to.Key, strings.Join(missing, ", "), ages[i].Key)})
		}
	}
	return problems, rows
}

// writeStorage renders the Storage Covenant check.
func writeStorage(sb *strings.Builder, problems []StorageProblem, rows []StorageRow) {
	fmt.Fprintf(sb, "Storage is a wall, sized by the storage rule (config/storage_rule.go) and checked here from the tables alone: five copies of an age's storage building, on the earlier stores, hold the dearest price the age asks of the reference town and a quarter more; each copy up to the fifth fits under the copies before it; a town arrives able to pay the next age's first storage building and housing; each gate asks for five copies of the storage building of the age it leaves and %g of the store each material needs; the first %d ages wall at exactly five copies and the later ones reach at most %g times the reference store. `go test ./smoke` fails on any row below.\n\n", gateShare, earlyWallAges, lateReach)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		sb.WriteString("| age | rule | |\n|---|---|---|\n")
		for _, p := range problems {
			fmt.Fprintf(sb, "| %s | %s | %s |\n", p.Age, p.Rule, p.Why)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| age | storage building | per copy | dearest ask | reference store | player reaches | x reference | copies at the wall |\n|---|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %.2fx | %d |\n", r.Age, r.Building, num(r.PerCopy), num(r.Dearest), num(r.Store), num(r.Reached), r.Ratio(), r.Copies)
	}
}
