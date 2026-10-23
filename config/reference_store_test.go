package config

import (
	"math"
	"testing"
)

// referenceStoreSizes is what one copy of each storage building held in the
// reference store, pinned. They follow the first-copy prices (the store was
// sized to hold five copies and the dearest ask of the age), so they moved
// when new copies took NewCopyPriceFactor 1.43 (they were 130, 1.5K, 8.8K ... at
// 1.29). The rules that still read a store size (the festival and Black Market
// prices, deals, the Brace ceiling, the Maximalist badges, event stock shares)
// are measured on them.
var referenceStoreSizes = map[string]float64{
	"stash": 150, "storage_pit": 1700, "warehouse": 9700, "granary": 35000, "classical_vault": 290000, "keep": 2.1e+06, "renaissance_vault": 8.6e+06, "colonial_warehouse": 4.1e+07, "industrial_depot": 1.2e+08, "victorian_vault": 2e+09, "electric_warehouse": 2.2e+10, "atomic_vault": 1.2e+11, "modern_depot": 4.8e+11, "info_vault": 3.3e+12, "digital_archive": 1.1e+13, "cyber_vault": 1.2e+14, "fusion_vault": 5.3e+14, "orbital_depot": 2.6e+15, "stellar_vault": 9.4e+16, "galactic_vault": 7.9e+17, "quantum_vault": 6.8e+19,
}

// TestReferenceStoreSizesArePinned: the size a copy held, by building,
// is the one master had, and the reference store on leaving an age is the base
// store and five copies of every storage building up to it.
func TestReferenceStoreSizesArePinned(t *testing.T) {
	defs := BaseBuildings()
	sizes, stores := ReferenceStoreSizes(defs, Ages(), ReferenceBaseStore())
	for key, want := range referenceStoreSizes {
		if got := sizes[key]; got != want {
			t.Errorf("%s: reference store per copy %v, pinned %v", key, got, want)
		}
	}
	if len(sizes) != len(referenceStoreSizes) {
		t.Errorf("%d sizes, master had %d storage buildings", len(sizes), len(referenceStoreSizes))
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
		if want := stores[age]; math.Abs(got-want) > 1e-9*want {
			t.Errorf("%s: typical wood store %v, reference store %v", age, got, want)
		}
	}
}
