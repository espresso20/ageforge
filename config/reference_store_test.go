package config

import "testing"

// masterStoreSizes is what one copy of each storage building held when the
// stores were real (master at #268), pinned. The reference store is that, so
// the rules that still read a store size (the festival and Black Market
// prices, deals, the Brace ceiling, the Maximalist badges, event stock
// shares) behave as they did until they are redone.
var masterStoreSizes = map[string]float64{
	"stash": 130, "storage_pit": 1500, "warehouse": 8800, "granary": 32000,
	"classical_vault": 270000, "keep": 1.9e6, "renaissance_vault": 7.8e6,
	"colonial_warehouse": 3.7e7, "industrial_depot": 1.1e8, "victorian_vault": 1.8e9,
	"electric_warehouse": 2e10, "atomic_vault": 1.1e11, "modern_depot": 4.3e11,
	"info_vault": 3e12, "digital_archive": 9.4e12, "cyber_vault": 1.1e14,
	"fusion_vault": 4.7e14, "orbital_depot": 2.4e15, "stellar_vault": 8.5e16,
	"galactic_vault": 7.1e17, "quantum_vault": 6.2e19,
}

// TestReferenceStoreIsWhatTheStoresWere: the size a copy held, by building,
// is the one master had, and the reference store on leaving an age is the base
// store and five copies of every storage building up to it.
func TestReferenceStoreIsWhatTheStoresWere(t *testing.T) {
	defs := BaseBuildings()
	sizes, stores := ReferenceStoreSizes(defs, Ages(), ReferenceBaseStore())
	for key, want := range masterStoreSizes {
		if got := sizes[key]; got != want {
			t.Errorf("%s: reference store per copy %v, master held %v", key, got, want)
		}
	}
	if len(sizes) != len(masterStoreSizes) {
		t.Errorf("%d sizes, master had %d storage buildings", len(sizes), len(masterStoreSizes))
	}
	want := ReferenceBaseStore()
	for _, a := range Ages() {
		for _, d := range defs {
			if d.Category == "storage" && d.RequiredAge == a.Key {
				want += float64(StorageRuleCopies * sizes[d.Key])
			}
		}
		if got := stores[a.Key]; got != want {
			t.Errorf("%s: reference store %v, want the base and five of every storage building so far = %v", a.Key, got, want)
		}
	}
}

// TestReferenceStoreFollowsTheTypicalStore: the typical store of the Brace
// ceiling (TypicalStorages) is the reference store, resource by resource: a
// resource's own base in place of wood's.
func TestReferenceStoreFollowsTheTypicalStore(t *testing.T) {
	typical := TypicalStorages(BaseBuildings(), BaseResources(), AgeOrder())
	_, stores := ReferenceStoreSizes(BaseBuildings(), Ages(), ReferenceBaseStore())
	for _, age := range AgeOrder() {
		got := typical[age]["wood"]
		if want := stores[age]; got != want {
			t.Errorf("%s: typical wood store %v, reference store %v", age, got, want)
		}
	}
}
