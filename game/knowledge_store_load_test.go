package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A save from before knowledge had a store of its own, loaded by the game
// after it. The fixture (testdata/pre_knowledge_store_medieval.json) was
// written and signed by the old rules (knowledge_fixture_gen_test.go): a
// Medieval town with 20 Keeps and 2 Monastery Libraries, knowledge full to the
// old cap (about 19.7 million) where the new store is 660,240, and the
// Guilds tech (1.08 million) in the plan.

const knowledgeStoreFixture = "testdata/pre_knowledge_store_medieval.json"

const knowledgeStoreLine = "Knowledge now has its own store, raised by knowledge buildings. What you hold above it stays until you spend it."

func knowledgeFixtureView(t *testing.T) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	data, err := os.ReadFile(knowledgeStoreFixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("kstore"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	if err := ge.loadSave("kstore", true); err != nil {
		t.Fatal(err)
	}
	return ge
}

func TestKnowledgeSave_FixtureIsTheOldGamesAndSigned(t *testing.T) {
	save := loadSaveFile(t, knowledgeStoreFixture)
	if save.KnowledgeRules != 0 {
		t.Fatalf("the fixture carries knowledge_rules %d: it is not from before the store", save.KnowledgeRules)
	}
	if ok, _ := verifySave(&save); !ok || save.Signature == "" {
		t.Fatal("the fixture is not signed and valid as the old game wrote it")
	}
}

func TestKnowledgeSave_LoadsWithNoTamperFlagAndKeepsAllItsKnowledge(t *testing.T) {
	save := loadSaveFile(t, knowledgeStoreFixture)
	ge := knowledgeFixtureView(t)
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}
	held := save.Resources["knowledge"]
	store := ge.Resources.GetStorage("knowledge")
	if store != config.KnowledgeBase+2*330000 {
		t.Fatalf("setup: the new store is %v, want the base and two Monastery Libraries", store)
	}
	if held < 10*store {
		t.Fatalf("setup: the save holds %v, not far over the new store %v", held, store)
	}
	if got := ge.Resources.Get("knowledge"); got != held {
		t.Errorf("knowledge after the load is %v, the save held %v", got, held)
	}
	if !ge.Resources.Graced("knowledge") {
		t.Error("the stock over the store is not kept by the grace rule")
	}
	for key, n := range save.Buildings {
		if got := ge.Buildings.GetCount(key); got != n {
			t.Errorf("%s: %d copies after the load, the save has %d", key, got, n)
		}
	}
}

func TestKnowledgeSave_GainsNothingWhileOverAndPaysForATech(t *testing.T) {
	ge := knowledgeFixtureView(t)
	if len(ge.plan) != 1 || ge.plan[0].Kind != PlanResearch {
		t.Fatalf("setup: the plan came back as %v", ge.plan)
	}
	key := ge.plan[0].Key
	def, _ := ge.rules.Tech(key)
	if def.Cost <= ge.Resources.GetStorage("knowledge") {
		t.Fatalf("setup: %s costs %v, not over the store", key, def.Cost)
	}
	held := ge.Resources.Get("knowledge")

	// No plan, so nothing spends it: the stock is flat while over the store.
	plan := ge.plan
	ge.plan = nil
	if ge.Resources.GetRate("knowledge") <= 0 {
		t.Fatal("setup: the town makes no knowledge, so gaining none proves nothing")
	}
	ge.StepTicks(20)
	if got := ge.Resources.Get("knowledge"); got != held {
		t.Errorf("knowledge moved from %v to %v while over its store", held, got)
	}

	// The queued tech is paid from the kept stock.
	ge.plan = plan
	ge.StepTicks(1)
	if ge.Research.currentTech != key {
		t.Fatalf("the plan did not start %s (researching %q)", key, ge.Research.currentTech)
	}
	if got, want := ge.Resources.Get("knowledge"), held-def.Cost; got != want {
		t.Errorf("knowledge after paying for %s is %v, want %v", key, got, want)
	}
}

func TestKnowledgeSave_TheCommandPaysFromKeptStockToo(t *testing.T) {
	ge := knowledgeFixtureView(t)
	ge.plan = nil
	if err := ge.StartResearch("guilds"); err != nil {
		t.Fatalf("research guilds: %v", err)
	}
}

func TestKnowledgeSave_NoticeOnceAndSavedInTheNewShape(t *testing.T) {
	isolateAccountDir(t)
	count := func(ge *GameEngine) int {
		n := 0
		for _, l := range ge.GetLogs() {
			if l.Message == knowledgeStoreLine {
				n++
			}
		}
		return n
	}
	if knowledgeStoreNotice != knowledgeStoreLine {
		t.Fatalf("the notice reads %q", knowledgeStoreNotice)
	}
	ge := loadFixture(t, knowledgeStoreFixture, "kstore")
	if ge.cheaterBadge {
		t.Fatal("the old game's signature no longer verifies")
	}
	if n := count(ge); n != 1 {
		t.Fatalf("the line appears %d times on the first load, want 1", n)
	}
	again := saveAndLoad(t, ge, "kstore_again")
	if n := count(again); n != 0 {
		t.Errorf("the line appears %d times on the second load, want 0", n)
	}
	data, err := os.ReadFile(savePath("kstore_again"))
	if err != nil {
		t.Fatal(err)
	}
	var saved GameSave
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.KnowledgeRules != knowledgeRules {
		t.Errorf("the re-saved game carries knowledge_rules %d, want %d", saved.KnowledgeRules, knowledgeRules)
	}
	if again.cheaterBadge || saved.Signature == "" {
		t.Errorf("the re-saved game is not signed and valid (cheater %v, signature %q)", again.cheaterBadge, saved.Signature)
	}
	// A game started now says nothing of the kind.
	fresh := NewGameEngine()
	for _, l := range fresh.GetLogs() {
		if strings.Contains(l.Message, "own store") {
			t.Errorf("a new game says %q", l.Message)
		}
	}
}
