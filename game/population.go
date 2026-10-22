package game

import "github.com/espresso20/ageforge/config"

// Population. "People" is the parent of every kind of person in the town:
// workers (they have a job), the reserve (they live in the homes and have no
// job yet), scholars, and, in a later change, soldiers.
//
//   - The population is the workers and the reserve (the WorkerManager's
//     pool) plus the scholars. The housing caps it.
//   - A knowledge building comes with its scholars (config.BuildingDef.Scholars
//     a copy). They are people from the moment the copy stands: they move in
//     with the building (no recruiting step), live in the homes and eat as
//     much as anyone. If the housing is full they still move in, and the
//     reserve stops growing until there is room again. A sold or destroyed
//     copy takes its scholars with it.
//   - So the room left for the pool is the housing less the population, and
//     the pool's own cap is the housing less the scholars (workerCapLocked).

// popLocked is the population: workers, the reserve and scholars. Caller
// holds the lock.
func (ge *GameEngine) popLocked() int { return ge.Workers.TotalPop() + ge.scholarTotal() }

// workerCapLocked is the most workers and reserve the housing holds with the
// scholars in it (negative when the scholars alone fill more than the homes).
// Caller holds the lock.
func (ge *GameEngine) workerCapLocked() int { return ge.popCapLocked() - ge.scholarTotal() }

// foodDrainLocked is the food everyone eats a tick: each worker, each in the
// reserve and each scholar pays the food class's cost for the age. Caller
// holds the lock.
func (ge *GameEngine) foodDrainLocked() float64 {
	drain := ge.Workers.FoodDrain()
	if s := ge.scholarTotal(); s > 0 {
		drain += float64(s) * ge.Workers.FoodCostPerWorker()
	}
	return drain
}

// Term is the word key has in the snapshot's age: "workers", "reserve",
// "soldiers", "scholars" (config.Term). The UI names people through it, so
// every age has its own name for them.
func (st GameState) Term(key string) config.Word { return config.Term(st.Age, key) }
