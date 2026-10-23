package config

// StorageBoostPercent is what each standing copy of a storage building adds
// to all production, in percent. Stock has no cap, so a storage building has
// nothing to store; it speeds the town up instead. The building carries it as
// a "bonus" effect on production_all, and the engine sums every copy into the
// production_all pool under the source "storage" (game.storageModifiers), so
// it counts in the same bonus pool as a wonder and shows in its breakdown.
// Older ages' copies keep counting: a storage building is never sold or lost.
const StorageBoostPercent = 1

// StorageBoost is what one copy of a storage building adds to the
// production_all pool, as a fraction (0.01).
func StorageBoost() float64 { return StorageBoostPercent / 100.0 }
