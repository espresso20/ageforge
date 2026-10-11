package config

import "testing"

// TestStorageTablesFollowTheRule recomputes every storage building's size
// from the storage rule (storage_rule.go) and holds the tables to it. A size
// typed by hand that the rule does not give fails here; so does a price or a
// gate count that moved without the sizes being worked out again.
func TestStorageTablesFollowTheRule(t *testing.T) {
	defs := BaseBuildings()
	ages := Ages()
	base := 0.0
	for _, r := range BaseResources() {
		if r.Key == "wood" {
			base = r.BaseStorage
		}
	}
	sizes, _ := StorageRuleSizes(defs, ages, base)
	seen := 0
	for _, d := range defs {
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
		got := 0.0
		for _, e := range d.Effects {
			if e.Type == "storage" && e.Target == "all" {
				got = e.Value
			}
		}
		if want := sizes[d.Key]; got != want {
			t.Errorf("%s holds %s a copy, the storage rule gives %s", d.Key, FormatRateValue(got), FormatRateValue(want))
		}
	}
	if seen != len(sizes) || seen == 0 {
		t.Fatalf("%d storage buildings in the table, %d sized by the rule", seen, len(sizes))
	}
}

// TestGatesAskForTheReferenceStore: every gate asks for five copies of the
// storage building of the age it leaves, and for no other storage building.
// The store those copies give is the one the storage rule sized to pay for
// the next age's first storage building; a town let through with less could
// never build storage again (the age lock forbids the older building).
func TestGatesAskForTheReferenceStore(t *testing.T) {
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

// TestFirstStorageCopyFitsTheStoreArrivedWith: in every age the first copy of
// the storage building fits under the reference store of the age before it
// (the first Stash under what a new game holds), and each copy up to the
// fifth under the store the copies before it give. Without this an age could
// open with nothing buyable.
func TestFirstStorageCopyFitsTheStoreArrivedWith(t *testing.T) {
	defs := BaseBuildings()
	store := 0.0
	for _, r := range BaseResources() {
		if r.Key == "wood" {
			store = r.BaseStorage
		}
	}
	for _, a := range Ages() {
		for _, d := range defs {
			if d.RequiredAge != a.Key || d.Category != "storage" {
				continue
			}
			size := 0.0
			for _, e := range d.Effects {
				if e.Type == "storage" && e.Target == "all" {
					size = e.Value
				}
			}
			for k := 0; k < StorageRuleCopies; k++ {
				if price := maxPrice(d, k); price > store {
					t.Errorf("%s: copy %d of the %s costs %s, the store by then holds %s", a.Name, k+1, d.Name, FormatRateValue(price), FormatRateValue(store))
				}
				store += size
			}
		}
	}
}
