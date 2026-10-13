//go:build knowledgefixture

package game

import (
	"os"
	"testing"
)

// The save the knowledge store has to load. It is a Medieval Age town as the
// game wrote it while knowledge sat in the general store: 20 Keeps (every
// store raised by them), 2 Monastery Libraries, knowledge full to that old
// cap, and a tech in the plan whose price is larger than the knowledge store
// the new rule gives the town (240 + 2 libraries) but not larger than what
// the town holds.
//
// This file builds only with -tags knowledgefixture. Run it on a checkout of
// the game from before the change, so the bytes are the old game's, written
// and signed by it:
//
//	UPDATE_FIXTURES=1 go test -tags knowledgefixture ./game -run TestWriteKnowledgeFixture
//
// then commit testdata/pre_knowledge_store_medieval.json.
func TestWriteKnowledgeFixture(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre knowledge store save")
	}
	isolateAccountDir(t)
	ge := newSeededEngine(94)
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
	ge.Buildings.counts["monastery_library"] = 2
	ge.mu.Unlock()
	ge.recalculateRates()

	// The cheapest Medieval tech that is open to research and costs more
	// than the new knowledge store (240 + 2 x 330000 = 660240).
	var pick string
	var cost float64
	ge.mu.Lock()
	for _, tech := range ge.rules.Techs() {
		if tech.Age != age || tech.Cost <= 660240 || tech.MissingPrereq(ge.Research.IsResearched) != "" || !tech.AnyOfMet(ge.Research.IsResearched) {
			continue
		}
		if pick == "" || tech.Cost < cost {
			pick, cost = tech.Key, tech.Cost
		}
	}
	ge.mu.Unlock()
	if pick == "" {
		t.Fatal("setup: no open Medieval tech costs more than the new store")
	}
	ge.mu.Lock()
	ge.Resources.Add("knowledge", 1e15)
	held := ge.Resources.Get("knowledge")
	ge.plan = []PlanItem{{Kind: PlanResearch, Key: pick}}
	ge.mu.Unlock()
	t.Logf("tech %s costs %v; knowledge held %v", pick, cost, held)
	if held < cost*2 {
		t.Fatalf("setup: the old cap %v is not far over the tech's price %v", held, cost)
	}
	if err := ge.SaveGame("pre_knowledge_store_medieval"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("pre_knowledge_store_medieval"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/pre_knowledge_store_medieval.json", data, 0o644); err != nil {
		t.Fatal(err)
	}
}
