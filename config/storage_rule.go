package config

// What is left of the storage rule. Stock is uncapped (#268) and a storage
// building boosts production instead of raising a store, so nothing sizes a
// store any more and no check holds a price against one. Three facts about
// storage buildings survive, because the gates and the price tables are built
// on them:
//
//   - a gate asks for StorageRuleCopies copies of the storage building of the
//     age it leaves (GateStorageCopies, in storage_rule_test.go), so a town
//     that passes arrives with five of each;
//   - a storage building climbs at StorageRate a copy (Kittens Game's barn),
//     and its price is its only limit: it has no copy cap;
//   - an older age's storage building cannot be built in a later age.
//
// The size a store used to have lives on as the reference store
// (reference_store.go), for the rules that still read one.
const (
	StorageRuleCopies = 5
	// StorageRules is the version of the storage rules a save is written
	// under: 1 from the storage wall on (a price larger than a store cannot
	// be bought, overflow does not pay the plan). A save without it is from
	// before, and keeps its stock on load.
	StorageRules = 1
	// StorageRate is the rate a storage building's price climbs at for each
	// copy owned: Kittens Game's barn.
	StorageRate = 1.75
)

// AgeHasStorageBuilding reports whether the age has a storage building of its
// own. The last age has none.
func AgeHasStorageBuilding(defs []BuildingDef, age string) bool {
	for _, d := range defs {
		if d.Category == "storage" && d.RequiredAge == age {
			return true
		}
	}
	return false
}
