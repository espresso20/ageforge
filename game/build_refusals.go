package game

import (
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// build_refusals.go holds the engine's reasons for refusing a build, a sale
// or an upgrade as text a snapshot can give, so a screen can show a button's
// reason in the engine's own words. The engine's commands use the same
// functions, so the two never differ.

// underConstructionErr is the refusal for a unique building that is already
// being built; left is how long the copy has still to go.
func underConstructionErr(name, left string) error {
	return vibe.Errorf("game.engine.startBuildPaid.02", "%s is already under construction (%s left).", name, left)
}

// QueuedRefusal is the engine's refusal to queue another copy of key while one
// is under construction, "" when the building may be queued again. Only a
// unique building is refused: a capped one queues copies up to its limit.
func (s GameState) QueuedRefusal(key string) string {
	def, ok := s.Ruleset().Building(key)
	if !ok || def.MaxCount != 1 {
		return ""
	}
	for _, q := range s.BuildQueue {
		if q.Key == key {
			return underConstructionErr(def.Name, DurationText(q.TicksLeft, time.Duration(s.TickIntervalMs)*time.Millisecond)).Error()
		}
	}
	return ""
}

// sellAgeErr is the refusal to sell in the first age.
func sellAgeErr(set *rules.Set, age string) error {
	if age == "primitive_age" {
		return vibe.Errorf("game.engine.SellBuilding.02", "You cannot sell buildings in the %s.", set.Name(rules.KindAge, "primitive_age"))
	}
	return nil
}

// sellBlockErr is the refusal to sell a copy of def, of which the town holds
// count and some are under construction.
func sellBlockErr(def config.BuildingDef, count int, queued bool) error {
	switch {
	case def.Category == "wonder":
		return vibe.Errorf("game.engine.SellBuilding.03", "Wonders cannot be sold.")
	case def.Category == "storage":
		// Storage is permanent, like wonders (see isDestroyable). Once its age
		// has passed the age lock never lets it be rebuilt, and an age's first
		// storage copy can cost more than the storage left after a sale, so a
		// sold copy could leave no way to raise a cap ever again.
		return vibe.Errorf("game.engine.SellBuilding.04", "Storage cannot be sold: once its age has passed, it can never be rebuilt.")
	case def.Scholars > 0:
		// Knowledge buildings are the knowledge store, spared the same way: a
		// tech priced over the smaller store could never be started.
		return vibe.Errorf("game.engine.SellBuilding.05", "Knowledge buildings cannot be sold: once its age has passed, it can never be rebuilt.")
	case count == 0:
		return vibe.Errorf("game.engine.SellBuilding.06", "You have no %s to sell.", def.Name)
	case queued:
		return vibe.Errorf("game.engine.SellBuilding.07", "A %s is still under construction. Sell after it is finished.", def.Name)
	}
	return nil
}

// SellRefusal is why a copy of key cannot be sold now, in the engine's words,
// "" when it can.
func (s GameState) SellRefusal(key string) string {
	set := s.Ruleset()
	if err := sellAgeErr(set, s.Age); err != nil {
		return err.Error()
	}
	def, ok := set.Building(key)
	if !ok {
		return ""
	}
	if err := sellBlockErr(def, s.Buildings[key].Count, s.queuedOf(key) > 0); err != nil {
		return err.Error()
	}
	return ""
}

// noUpgradeErr is the refusal to upgrade a building that is not an older one.
func noUpgradeErr(name string) error {
	return vibe.Errorf("game.engine.upgradeBuildingLocked.01", "%s has no upgrade this age. Type 'upgrade' to list the ones that do.", name)
}

// noneToUpgradeErr is the refusal to upgrade a building the town does not hold.
func noneToUpgradeErr(name string) error {
	return vibe.Errorf("game.engine.upgradeBuildingLocked.02", "You have no %s to upgrade.", name)
}

// UpgradeOffer is what upgrading the next copy of an older building comes to:
// the building it becomes, its price, and why it cannot be done now ("" when
// it can).
type UpgradeOffer struct {
	FromKey, ToKey, ToName string
	Count                  int // copies held
	Cost                   map[string]float64
	Refusal                string
}

// UpgradeOffer is the offer to upgrade the next copy of key: for an older
// building the town holds, what it becomes and costs and why it cannot be
// done now; for any other building, only the engine's refusal. It takes the
// engine's read lock, so a screen asks once per refresh, for one building.
func (ge *GameEngine) UpgradeOffer(key string) UpgradeOffer {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	o := UpgradeOffer{FromKey: key}
	oldDef, ok := ge.rules.Building(key)
	if !ok {
		return o
	}
	if ge.Buildings.counts[key] == 0 && ge.Buildings.IsOlder(key) {
		o.Refusal = noneToUpgradeErr(oldDef.Name).Error()
		return o
	}
	target, ok := ge.Buildings.UpgradeTarget(key)
	if !ok || !ge.Buildings.IsOlder(key) {
		o.Refusal = noUpgradeErr(oldDef.Name).Error()
		return o
	}
	cost, err := ge.upgradeCheckLocked(key, oldDef, target)
	o.ToKey, o.ToName, o.Count, o.Cost = target.Key, target.Name, ge.Buildings.counts[key], cost
	if err != nil {
		o.Refusal = err.Error()
	}
	return o
}
