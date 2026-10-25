package config

import (
	"math"
	"strings"
	"testing"
)

// TestStorageBuildingsBoostProduction: a storage building holds nothing (stock
// has no cap); each copy adds StorageBoostPercent percent to all production,
// and that is all it does. Its price climbs at the storage rate with no copy
// limit: the price is the limit.
func TestStorageBuildingsBoostProduction(t *testing.T) {
	seen := 0
	for _, d := range BaseBuildings() {
		if d.Category != "storage" {
			continue
		}
		seen++
		if d.CostScale != StorageRate {
			t.Errorf("%s climbs at %g a copy, the storage rate is %g", d.Key, d.CostScale, StorageRate)
		}
		if d.MaxCount != 0 {
			t.Errorf("%s has a copy limit of %d: the price curve is the limit on storage", d.Key, d.MaxCount)
		}
		if len(d.Effects) != 1 || d.Effects[0].Type != "bonus" || d.Effects[0].Target != "production_all" || d.Effects[0].Value != StorageBoost() {
			t.Errorf("%s effects = %v, want one bonus of %g%% on production_all", d.Key, d.Effects, float64(StorageBoostPercent))
		}
		if !strings.Contains(d.Description, "+1% all production") {
			t.Errorf("%s says %q: its text must say what it does", d.Key, d.Description)
		}
	}
	if seen != 21 {
		t.Fatalf("%d storage buildings in the table, want 21", seen)
	}
}

// TestGatesAskForFiveStorageCopies: every gate asks for five copies of the
// storage building of the age it leaves, and for no other storage building.
// A town that passes arrives with five of each age's, so the boost they give
// is the same for everyone who is let through.
func TestGatesAskForFiveStorageCopies(t *testing.T) {
	defs := BaseBuildings()
	storage := map[string]bool{}
	for _, d := range defs {
		if d.Category == "storage" {
			storage[d.Key] = true
		}
	}
	want := GateStorageCopies(defs, Ages())
	asked := 0
	for _, a := range Ages() {
		for key, n := range want[a.Key] {
			asked++
			if a.BuildingReqs[key] != n {
				t.Errorf("the gate into the %s asks for %d %s, the rule gives %d", a.Name, a.BuildingReqs[key], key, n)
			}
		}
		for key, n := range a.BuildingReqs {
			if storage[key] && want[a.Key][key] != n {
				t.Errorf("the gate into the %s asks for %d %s, which the rule does not give", a.Name, n, key)
			}
		}
	}
	if asked != len(storage) {
		t.Errorf("%d gates ask for storage, want one for each of the %d storage buildings", asked, len(storage))
	}
}

// storageCopyIncomeShare is the most of the age's typical income, over one
// pacing target of the age, that any one price part of a storage copy the
// gate asks for may be. The stores used to be the bound on these prices; with
// none, the bound is what the age makes: every copy the gate asks for, the
// first Stash included, is payable from the age's own income inside one
// target of the age.
const storageCopyIncomeShare = 1.0

// TestStorageCopiesAreAffordableFromTheAgesIncome: the intent of the old
// "first copy fits the store" rule that survives without a store. Each of the
// copies a gate asks for (StorageRuleCopies) of an age's storage building has
// every price part inside what the age's typical income (Incomes: a moderate
// town's) makes in one pacing target of the age. Without this an age could
// open with a storage building nobody can pay for.
func TestStorageCopiesAreAffordableFromTheAgesIncome(t *testing.T) {
	defs := BaseBuildings()
	income := Incomes(defs, Technologies(), AgeOrder(), AnyResource)
	checked := 0
	for _, d := range defs {
		if d.Category != "storage" {
			continue
		}
		checked++
		target := LatePriceTargetTicks(d.RequiredAge)
		for res, base := range d.BaseCost {
			per := income[d.RequiredAge][res]
			if per <= 0 {
				t.Errorf("%s costs %s, which the %s makes none of", d.Key, res, d.RequiredAge)
				continue
			}
			limit := storageCopyIncomeShare * per * target
			for k := 0; k < StorageRuleCopies; k++ {
				if price := base * math.Pow(d.CostScale, float64(k)); price > limit {
					t.Errorf("%s copy %d costs %s %s, more than the %s the %s makes in one target", d.Key, k+1, FormatRateValue(price), res, FormatRateValue(limit), d.RequiredAge)
				}
			}
		}
	}
	if checked != 21 {
		t.Fatalf("%d storage buildings checked, want 21", checked)
	}
}

// GateStorageCopies is what the gate out of each age asks for in storage
// buildings, by the age the gate leads into: StorageRuleCopies copies of the
// storage building of the age left. An age with no storage building of its
// own asks for none.
func GateStorageCopies(defs []BuildingDef, ages []AgeDef) map[string]map[string]int {
	out := map[string]map[string]int{}
	for i := 0; i+1 < len(ages); i++ {
		for _, d := range defs {
			if d.Category == "storage" && d.RequiredAge == ages[i].Key {
				if out[ages[i+1].Key] == nil {
					out[ages[i+1].Key] = map[string]int{}
				}
				out[ages[i+1].Key][d.Key] = StorageRuleCopies
			}
		}
	}
	return out
}
