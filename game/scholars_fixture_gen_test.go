//go:build scholarsfixture

package game

import (
	"os"
	"testing"
)

// The saves the scholars change has to load. Each is a Medieval Age town as
// the game wrote it while knowledge was a stock, knowledge buildings were
// staffed from the population and a tech was paid for in knowledge when it
// started.
//
//   - pre_scholars_medieval.json: knowledge held, workers assigned to the
//     knowledge buildings (more copies of the Monastery Library than the new
//     cap), a tech paid for and waiting out its research time, and a second
//     tech in the plan.
//   - pre_scholars_idle.json: the same town with no tech in progress and
//     nothing in the plan.
//
// This file builds only with -tags scholarsfixture. Run it on a checkout of the
// game from before the change, so the bytes are the old game's, written and
// signed by it:
//
//	UPDATE_FIXTURES=1 go test -tags scholarsfixture ./game -run TestWriteScholarsFixtures
func TestWriteScholarsFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre scholars saves")
	}
	for _, name := range []string{"pre_scholars_medieval", "pre_scholars_idle"} {
		writeScholarsFixture(t, name, name == "pre_scholars_medieval")
	}
}

func writeScholarsFixture(t *testing.T, name string, busy bool) {
	t.Helper()
	isolateAccountDir(t)
	ge := newSeededEngine(95)
	const age = "medieval_age"
	ge.mu.Lock()
	ge.Stats.AgesReached = nil
	for _, a := range ageKeys() {
		ge.Stats.AgesReached = append(ge.Stats.AgesReached, a)
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
		if w := ge.progress.WonderForAge(a); w != "" {
			ge.Buildings.counts[w] = 1
		}
	}
	ge.age = age
	ge.currentEpoch = ge.rules.EraOf(age)
	ge.Workers.SetAge(age)
	before := map[string]bool{}
	for _, a := range ageKeys() {
		before[a] = true
		if a == age {
			break
		}
	}
	for _, tech := range ge.rules.Techs() {
		if before[tech.Age] && tech.Age != age {
			ge.Research.researched[tech.Key] = true
		}
	}
	ge.Research.rebuildBonuses()
	ge.Buildings.counts["keep"] = 20
	ge.Buildings.counts["house"] = 200
	ge.Buildings.counts["library"] = 6
	ge.Buildings.counts["monastery_library"] = 12
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 300, 100000)
	ge.Workers.Assign("worker", "monastery_library", 30)
	ge.Workers.Assign("worker", "library", 12)
	ge.mu.Unlock()
	ge.recalculateRates()

	// Two open Medieval techs, cheapest first.
	var picks []string
	ge.mu.Lock()
	cheap := map[string]float64{}
	for _, tech := range ge.rules.Techs() {
		if tech.Age != age || tech.MissingPrereq(ge.Research.IsResearched) != "" || !tech.AnyOfMet(ge.Research.IsResearched) {
			continue
		}
		cheap[tech.Key] = tech.Cost
		picks = append(picks, tech.Key)
	}
	ge.mu.Unlock()
	if len(picks) < 2 {
		t.Fatal("setup: fewer than two open Medieval techs")
	}
	// Pick the cheapest two.
	for i := range picks {
		for j := i + 1; j < len(picks); j++ {
			if cheap[picks[j]] < cheap[picks[i]] {
				picks[i], picks[j] = picks[j], picks[i]
			}
		}
	}
	ge.mu.Lock()
	ge.Resources.Add("knowledge", 1e15)
	if busy {
		if err := ge.startResearchLocked(picks[0], true); err != nil {
			ge.mu.Unlock()
			t.Fatal(err)
		}
		ge.Research.ticksLeft = ge.Research.totalTicks / 2
		ge.plan = []PlanItem{{Kind: PlanResearch, Key: picks[1], Count: 1}}
	}
	ge.mu.Unlock()
	t.Logf("%s: in progress %v, queued %v, knowledge held %v", name, picks[0], picks[1], ge.Resources.Get("knowledge"))
	if err := ge.SaveGame(name); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/"+name+".json", data, 0o644); err != nil {
		t.Fatal(err)
	}
}
