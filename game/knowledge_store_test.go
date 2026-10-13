package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The knowledge store (knowledge_store.go): the base plus the knowledge
// buildings' shares, and nothing else.

func knowledgeStoreOf(ge *GameEngine) float64 {
	ge.recalculateRates()
	return ge.Resources.GetStorage("knowledge")
}

func TestKnowledgeStore_StorageBuildingDoesNotRaiseIt(t *testing.T) {
	ge := newSeededEngine(7)
	base := knowledgeStoreOf(ge)
	if base != config.KnowledgeBase {
		t.Fatalf("a town with no building holds %v knowledge, want the base %v", base, config.KnowledgeBase)
	}
	wood := ge.Resources.GetStorage("wood")
	for _, key := range []string{"stash", "storage_pit", "warehouse"} {
		ge.Buildings.counts[key] = 10
	}
	if got := knowledgeStoreOf(ge); got != base {
		t.Errorf("storage buildings moved the knowledge store: %v -> %v", base, got)
	}
	if ge.Resources.GetStorage("wood") <= wood {
		t.Fatal("setup: the storage buildings raised no other store")
	}
}

func TestKnowledgeStore_KnowledgeBuildingRaisesItByItsFigure(t *testing.T) {
	ge := newSeededEngine(7)
	base := knowledgeStoreOf(ge)
	n := 0
	for _, def := range ge.rules.Buildings() {
		if !config.IsKnowledgeBuilding(def) {
			continue
		}
		n++
		ge.Buildings.counts[def.Key] = 0
		if def.KnowledgeStore <= 0 {
			t.Errorf("%s makes knowledge and holds no store", def.Key)
			continue
		}
		before := knowledgeStoreOf(ge)
		ge.Buildings.counts[def.Key] = 3
		want := before + float64(3*def.KnowledgeStore)
		if got := knowledgeStoreOf(ge); math.Abs(got-want) > 1e-9*want {
			t.Errorf("3 %s: the store is %v, want %v (+3 x %v)", def.Key, got, want, def.KnowledgeStore)
		}
		ge.Buildings.counts[def.Key] = 0
	}
	if n == 0 {
		t.Fatal("no knowledge building found")
	}
	if got := knowledgeStoreOf(ge); got != base {
		t.Errorf("the store did not return to the base: %v", got)
	}
}

func TestKnowledgeStore_TheGatesCopiesHoldWhatTheAgeIsPricedFor(t *testing.T) {
	ge := newSeededEngine(7)
	ages := ge.rules.Ages()
	held := config.KnowledgeStoreHeld(ge.rules.Buildings(), ages)
	for i, a := range ages {
		if i+1 < len(ages) {
			for _, def := range ge.rules.Buildings() {
				if def.RequiredAge == a.Key && config.IsKnowledgeBuilding(def) {
					ge.Buildings.counts[def.Key] = ages[i+1].BuildingReqs[def.Key]
				}
			}
		}
		got, want := knowledgeStoreOf(ge), held[a.Key]
		if math.Abs(got-want) > 1e-9*want {
			t.Errorf("%s: a town with the gates' copies holds %v knowledge, want %v", a.Key, got, want)
		}
	}
}

func TestKnowledgeStore_MasteryAndTechsLeaveItAlone(t *testing.T) {
	ge := newSeededEngine(7)
	ge.Buildings.counts["scriptorium"] = 5
	base := knowledgeStoreOf(ge)
	general := ge.Resources.GetStorage("wood")

	ge.Prestige.SetMastery(ge.age, 9)
	if ge.speedK() <= 1 {
		t.Fatal("setup: mastery gave no speed")
	}
	if got := knowledgeStoreOf(ge); got != base {
		t.Errorf("mastery moved the knowledge store: %v -> %v", base, got)
	}
	if ge.Resources.GetStorage("wood") <= general {
		t.Error("setup: mastery did not raise the general stores")
	}

	learn(ge, "pottery", "masonry") // +10% storage each
	if got := knowledgeStoreOf(ge); got != base {
		t.Errorf("a tech's storage bonus moved the knowledge store: %v -> %v", base, got)
	}
}

// A bonus that names knowledge storage by itself would still apply; none
// exists. If one is ever added, this test says so and the store takes it
// (specific, in recalculateRates).
func TestKnowledgeStore_NoBonusNamesIt(t *testing.T) {
	for _, def := range config.BaseBuildings() {
		for _, e := range def.Effects {
			if e.Type == "storage" && e.Target == "knowledge" {
				t.Errorf("%s names knowledge storage", def.Key)
			}
		}
	}
	for _, m := range config.Milestones() {
		for _, e := range m.Rewards {
			if e.Target == "knowledge" && e.Type != "instant_resource" {
				t.Errorf("milestone %s: %s names knowledge", m.Key, e.Type)
			}
		}
	}
	for _, u := range config.PrestigeUpgrades() {
		if u.EffectKey == "knowledge" {
			t.Errorf("prestige upgrade %s names knowledge storage", u.Key)
		}
	}
}

// An age advance trims every stock but faith's and knowledge's. Knowledge's
// store is the limit now, and a tech is bought from it, so what the player
// saved comes through the advance whole.
func TestKnowledgeStore_AnAdvanceDoesNotTrimKnowledge(t *testing.T) {
	ge := NewGameEngine()
	ge.Buildings.counts["story_circle"] = 4
	ge.recalculateRates()
	store := ge.Resources.GetStorage("knowledge")
	ge.Resources.LoadAmounts(map[string]float64{"knowledge": store, "wood": 1e6})

	ge.advanceAge("stone_age")

	if got := ge.Resources.Get("knowledge"); got != store {
		t.Errorf("knowledge after the advance is %v, want the full store %v", got, store)
	}
	if got := ge.Resources.Get("wood"); got >= 1e6 {
		t.Errorf("setup: the advance trimmed no other stock (wood %v)", got)
	}
}
