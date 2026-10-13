package game

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
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

// dearTech is a tech of the engine's age with no prerequisite whose price is
// over the knowledge store.
func dearTech(t *testing.T, ge *GameEngine) config.TechDef {
	t.Helper()
	store := knowledgeStoreOf(ge)
	for _, def := range ge.rules.Techs() {
		if def.Age == ge.age && def.Cost > store && def.MissingPrereq(ge.Research.IsResearched) == "" && def.AnyOfMet(ge.Research.IsResearched) {
			return def
		}
	}
	t.Fatalf("setup: no open tech of %s costs more than the store %v", ge.age, store)
	return config.TechDef{}
}

// holdKnowledge puts amount in the store, over the cap if need be, as a save
// from before the knowledge store keeps it (the grace rule).
func holdKnowledge(ge *GameEngine, amount float64) {
	ge.mu.Lock()
	ge.Resources.resources["knowledge"].Amount = amount
	ge.Resources.markGraceAll()
	ge.mu.Unlock()
	ge.recalculateRates()
}

// A tech priced over the knowledge store cannot be started, by any path that
// starts research; the refusal is the one every store gives, naming the tech,
// the store and the knowledge building of the age. Stock held above the store
// (an old save) can still pay for it.
func TestKnowledgeStore_ATechOverTheStoreCannotBeStarted(t *testing.T) {
	paths := map[string]func(ge *GameEngine, key string) (string, error){
		"command": func(ge *GameEngine, key string) (string, error) {
			return "", ge.StartResearch(key)
		},
		"locked start": func(ge *GameEngine, key string) (string, error) {
			ge.mu.Lock()
			defer ge.mu.Unlock()
			return "", ge.startResearchLocked(key, true)
		},
		"plan": func(ge *GameEngine, key string) (string, error) {
			ge.mu.Lock()
			defer ge.mu.Unlock()
			ge.plan = []PlanItem{{Kind: PlanResearch, Key: key, Count: 1}}
			note := ""
			for _, v := range ge.planViews() {
				note = v.Note
			}
			ge.runPlanTick()
			if ge.Research.currentTech == "" {
				return note, fmt.Errorf("the plan did not start it")
			}
			return note, nil
		},
	}
	for name, path := range paths {
		t.Run(name+"/refused", func(t *testing.T) {
			ge := newSeededEngine(7)
			def := dearTech(t, ge)
			holdKnowledge(ge, ge.Resources.GetStorage("knowledge"))
			note, err := path(ge, def.Key)
			if err == nil {
				t.Fatalf("%s started %s, priced %v over the store", name, def.Key, def.Cost)
			}
			if ge.Research.currentTech != "" {
				t.Errorf("%s: research %q began anyway", name, ge.Research.currentTech)
			}
			if name == "plan" {
				if note != "needs more knowledge storage" {
					t.Errorf("the plan says %q", note)
				}
				return
			}
			store := ge.Resources.GetStorage("knowledge")
			bld, ok := ge.knowledgeBuildingOfAge()
			if !ok {
				t.Fatal("setup: the age has no knowledge building")
			}
			want := ge.storeTooSmallText(def.Name, "knowledge", def.Cost)
			if err.Error() != want {
				t.Errorf("refusal %q, want %q", err.Error(), want)
			}
			for _, part := range []string{def.Name, textfmt.Number(store), bld.Name} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("refusal %q lacks %q", err.Error(), part)
				}
			}
		})
		t.Run(name+"/paid from kept stock", func(t *testing.T) {
			ge := newSeededEngine(7)
			def := dearTech(t, ge)
			holdKnowledge(ge, def.Cost*3)
			if _, err := path(ge, def.Key); err != nil {
				t.Fatalf("%s refused a tech the held stock pays for: %v", name, err)
			}
			if ge.Research.currentTech != def.Key {
				t.Errorf("researching %q, want %q", ge.Research.currentTech, def.Key)
			}
			if got := ge.Resources.Get("knowledge"); got != def.Cost*2 {
				t.Errorf("knowledge left %v, want %v", got, def.Cost*2)
			}
		})
	}
}

// A building priced in knowledge meets the same wall through the common
// code (overStore in startBuildLocked): over the store and not held is
// refused, and stock kept above the store can pay.
func TestKnowledgeStore_ABuildingPricedInKnowledgeMeetsTheWall(t *testing.T) {
	priced := 0
	for _, def := range config.BaseBuildings() {
		if def.Category != "wonder" && def.BaseCost["knowledge"] > 0 {
			priced++
		}
	}
	if priced == 0 {
		t.Fatal("no building is priced in knowledge: the wall has nothing to meet")
	}
	price := func(ge *GameEngine) {
		d := ge.Buildings.defs["hut"]
		d.BaseCost = map[string]float64{"knowledge": 5 * config.KnowledgeBase}
		ge.Buildings.defs["hut"] = d
		ge.recalculateRates()
	}
	ge := newSeededEngine(7)
	price(ge)
	holdKnowledge(ge, ge.Resources.GetStorage("knowledge"))
	err := ge.BuildBuilding("hut")
	if err == nil || !strings.Contains(err.Error(), "more than your knowledge storage holds") {
		t.Errorf("a building priced over the knowledge store: %v", err)
	}
	ge = newSeededEngine(7)
	price(ge)
	holdKnowledge(ge, 20*config.KnowledgeBase)
	if err := ge.BuildBuilding("hut"); err != nil {
		t.Errorf("kept stock should pay for it: %v", err)
	}
}

// A wonder's bank takes knowledge past the store as it takes any material:
// the bank is not a store, so deposits add up to a price the store could
// never hold.
func TestKnowledgeStore_AWondersBankTakesKnowledgePastTheStore(t *testing.T) {
	ge := newSeededEngine(7)
	const wonder = "great_library"
	price := ge.Buildings.defs[wonder].BaseCost["knowledge"]
	store := knowledgeStoreOf(ge)
	if price <= 2*store {
		t.Fatalf("setup: the price %v is not well over the store %v", price, store)
	}
	ge.Buildings.UnlockBuilding(wonder)
	banked := 0.0
	for i := 0; i < 3; i++ {
		setResource(ge, "knowledge", store)
		ge.recalculateRates()
		got, err := ge.BankWonderMax(wonder, "knowledge")
		if err != nil {
			t.Fatalf("deposit %d: %v", i+1, err)
		}
		banked += got
	}
	if bank := ge.Buildings.GetWonderBank(wonder)["knowledge"]; bank != banked || bank <= store {
		t.Errorf("the bank holds %v after three deposits of a full store (%v), want %v and more than one store", bank, store, banked)
	}
}
