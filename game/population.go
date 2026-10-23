package game

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// Population. "People" is the parent of every kind of person in the town:
// workers (they have a job), the reserve (they live in the homes and have no
// job yet), scholars and soldiers.
//
//   - The population is the workers and the reserve (the WorkerManager's
//     pool) plus the scholars. The housing caps it.
//   - A knowledge building comes with its scholars (config.BuildingDef.Scholars
//     a copy). They are people from the moment the copy stands: they move in
//     with the building (no recruiting step), live in the homes and eat as
//     much as anyone. If the housing is full they still move in, and the
//     reserve stops growing until there is room again. A sold or destroyed
//     copy takes its scholars with it.
//   - A military building comes with its soldiers (config.BuildingDef.Soldiers
//     a copy) the same way: people from the moment the copy stands, in the
//     homes, eating (soldiers.go).
//   - So the room left for the pool is the housing less the population, and
//     the pool's own cap is the housing less the scholars and the soldiers
//     (workerCapLocked).

// popLocked is the population: workers, the reserve, scholars and soldiers.
// Caller holds the lock.
func (ge *GameEngine) popLocked() int {
	return ge.Workers.TotalPop() + ge.scholarTotal() + ge.soldierTotal()
}

// workerCapLocked is the most workers and reserve the housing holds with the
// scholars and the soldiers in it (negative when they alone fill more than the
// homes). Caller holds the lock.
func (ge *GameEngine) workerCapLocked() int {
	return ge.popCapLocked() - ge.scholarTotal() - ge.soldierTotal()
}

// foodDrainLocked is the food everyone eats a tick: each worker, each in the
// reserve, each scholar and each soldier pays the food class's cost for the
// age. Caller holds the lock.
func (ge *GameEngine) foodDrainLocked() float64 {
	drain := ge.Workers.FoodDrain()
	if s := ge.scholarTotal() + ge.soldierTotal(); s > 0 {
		drain += float64(float64(s) * ge.Workers.FoodCostPerWorker())
	}
	return drain
}

// Term is the word key has in the snapshot's age: "workers", "reserve",
// "soldiers", "scholars" (config.Term). The UI names people through it, so
// every age has its own name for them.
func (st GameState) Term(key string) config.Word { return config.Term(st.Age, key) }

// logScholarsArrived says, for a knowledge building that has just been built,
// that its scholars have moved in. Caller holds the lock.
func (ge *GameEngine) logScholarsArrived(def config.BuildingDef) {
	if def.Scholars <= 0 {
		return
	}
	w := config.Term(ge.age, config.TermScholars)
	ge.addLogAbout(LogRoutine, TagPeople|TagGood, vibe.Sprintf("game.population.logScholarsArrived.01", "%s moved in with the %s.", textfmt.Count(def.Scholars, strings.ToLower(w.One), strings.ToLower(w.Many)), def.Name))
}
