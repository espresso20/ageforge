package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// The plan as a research queue: `plan research` plans what a tech still
// needs before it, takes any tech in sight, and a tech of a later age waits
// for the age without holding up this age's research.

// chainSet is the core tables plus a small tree in the first age whose two
// branches cost different amounts:
//
//	q_a, q_deep, q_left     need nothing (10, 20 and 25 knowledge)
//	q_right                 needs q_deep (10)
//	q_goal                  needs q_a, and q_left or q_right
//
// So the left branch adds 25 and the right one 30, or 10 once q_deep is
// held.
func chainSet() *rules.Set {
	src := rules.FromConfig()
	tech := func(key, name string, cost float64, def config.TechDef) config.TechDef {
		def.Key, def.Name, def.Age, def.Lane = key, name, src.Ages[0].Key, config.LaneKnowledge
		def.Cost, def.ResearchTicks = cost, 4
		return def
	}
	src.Techs = append(src.Techs,
		tech("q_a", "Alpha", 10, config.TechDef{}),
		tech("q_deep", "Deep", 20, config.TechDef{}),
		tech("q_left", "Left", 25, config.TechDef{}),
		tech("q_right", "Right", 10, config.TechDef{Prerequisites: []string{"q_deep"}}),
		tech("q_goal", "Goal", 10, config.TechDef{Prerequisites: []string{"q_a"}, AnyOf: []string{"q_left", "q_right"}}),
	)
	return rules.Compile(src)
}

func chainEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := NewGameEngineWith(chainSet())
	ge.SeedRNG(5)
	ge.Resources.resources["knowledge"].Storage = 1e6
	setAmount(ge, "knowledge", 1000)
	return ge
}

// researchKeys is the plan's research items, in order.
func researchKeys(ge *GameEngine) []string {
	var out []string
	for _, it := range ge.plan {
		if it.Kind == PlanResearch {
			out = append(out, it.Key)
		}
	}
	return out
}

// TestPlanResearchAddsWhatATechNeeds: one `plan research` plans the tech's
// missing prerequisites first, each after its own, and for an either-or
// group the branch already on its way, else the one that costs least to
// add. The plan then researches the chain in that order.
func TestPlanResearchAddsWhatATechNeeds(t *testing.T) {
	// Nothing held: Alpha, then the cheaper branch (Left, 25 against Deep
	// and Right, 30), then Goal.
	ge := chainEngine(t)
	chain, err := ge.PlanAddResearchChain("q_goal")
	if want := []string{"q_a", "q_left", "q_goal"}; err != nil || !reflect.DeepEqual(chain, want) || !reflect.DeepEqual(researchKeys(ge), want) {
		t.Fatalf("chain %v, plan %v, err %v; want %v", chain, researchKeys(ge), err, want)
	}
	for _, want := range chain {
		ge.runPlanTick()
		if ge.Research.currentTech != want {
			t.Fatalf("researching %q, want %s; plan %+v", ge.Research.currentTech, want, ge.plan)
		}
		researchOut(ge)
	}
	if len(ge.plan) != 0 || !ge.Research.IsResearched("q_goal") || ge.Research.IsResearched("q_right") {
		t.Errorf("after the chain: plan %+v, researched Goal %v, Right %v", ge.plan, ge.Research.IsResearched("q_goal"), ge.Research.IsResearched("q_right"))
	}

	// With Deep held the right branch adds 10, so it is the cheaper one.
	ge = chainEngine(t)
	learn(ge, "q_deep")
	if chain, _ := ge.PlanAddResearchChain("q_goal"); !reflect.DeepEqual(chain, []string{"q_a", "q_right", "q_goal"}) {
		t.Errorf("with Deep researched the chain is %v, want the right branch", chain)
	}

	// A branch already planned, in progress or researched is the one taken,
	// whatever it costs, and what is on its way is not planned twice.
	ge = chainEngine(t)
	if chain, _ := ge.PlanAddResearchChain("q_right"); !reflect.DeepEqual(chain, []string{"q_deep", "q_right"}) {
		t.Fatalf("planning Right added %v", chain)
	}
	if chain, _ := ge.PlanAddResearchChain("q_goal"); !reflect.DeepEqual(chain, []string{"q_a", "q_goal"}) {
		t.Errorf("with Right planned the chain is %v, want Alpha and Goal only", chain)
	}
	ge = chainEngine(t)
	learn(ge, "q_a")
	if err := ge.StartResearch("q_left"); err != nil {
		t.Fatal(err)
	}
	if chain, _ := ge.PlanAddResearchChain("q_goal"); !reflect.DeepEqual(chain, []string{"q_goal"}) {
		t.Errorf("with Alpha researched and Left in progress the chain is %v, want Goal alone", chain)
	}

	// What the UI reads from a snapshot is what the engine would add.
	ge = chainEngine(t)
	if got := PlanResearchChain(ge.GetState(), "q_goal"); !reflect.DeepEqual(got, []string{"q_a", "q_left", "q_goal"}) {
		t.Errorf("the snapshot's chain is %v", got)
	}
	_ = ge.PlanAddResearch("q_goal")
	if got := PlanResearchChain(ge.GetState(), "q_goal"); got != nil {
		t.Errorf("a planned tech still has a chain to add: %v", got)
	}
}

// TestPlanResearchRefusals: what the plan still says no to, and that a chain
// goes in whole or not at all.
func TestPlanResearchRefusals(t *testing.T) {
	ge := chainEngine(t)
	_ = ge.PlanAddResearch("q_goal")
	if err := ge.PlanAddResearch("q_goal"); err == nil || err.Error() != "Goal is already in the plan." {
		t.Errorf("planning a planned tech: %v", err)
	}
	learn(ge, "q_deep")
	if err := ge.PlanAddResearch("q_deep"); err == nil || err.Error() != "Can't plan Deep: already researched." {
		t.Errorf("planning a researched tech: %v", err)
	}

	// Research takes no room in the plan: a chain goes into a plan that is
	// full of builds, and a full plan still refuses the next build.
	ge = chainEngine(t)
	for planLoad(ge.plan) < MaxPlanItems {
		ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: "hut", Count: 1})
	}
	chain, err := ge.PlanAddResearchChain("q_goal")
	if err != nil || len(chain) != 3 || len(researchKeys(ge)) != 3 {
		t.Errorf("a chain of three into a plan full of builds: %v, %v", chain, err)
	}
	if planLoad(ge.plan) != MaxPlanItems || len(ge.plan) != MaxPlanItems+3 {
		t.Errorf("the plan holds %d items, %d of them counted; want %d and %d", len(ge.plan), planLoad(ge.plan), MaxPlanItems+3, MaxPlanItems)
	}
	if _, err := ge.PlanAddBuild("stash", 1); err == nil || err.Error() != "The plan is full (60 items, not counting research). Remove one with plan remove <n> first." {
		t.Errorf("a build into a full plan: %v", err)
	}
	if err := ge.PlanAddAdvance(); err == nil || !strings.HasPrefix(err.Error(), "The plan is full") {
		t.Errorf("an advance into a full plan: %v", err)
	}
	// The save keeps all of it, and a hand-made plan cannot hold more
	// research than the tree has techs.
	if got := loadPlanIn(ge.rules, ge.plan); len(got) != len(ge.plan) {
		t.Errorf("a plan of %d items loads as %d", len(ge.plan), len(got))
	}
	var flood []PlanItem
	for i := 0; i < ge.rules.TechCount()+50; i++ {
		flood = append(flood, PlanItem{Kind: PlanResearch, Key: "q_a", Count: 1})
	}
	if got := loadPlanIn(ge.rules, flood); len(got) != ge.rules.TechCount() {
		t.Errorf("a save with %d research items loads %d, want the tree's %d at most", len(flood), len(got), ge.rules.TechCount())
	}
}

// TestPlanResearchTakesAnyTechInSight: the plan takes a tech of any age the
// player may see named. A tech of an age still hidden is answered like a
// key that does not exist, naming nothing hidden; on known ground it goes
// in with its whole chain.
func TestPlanResearchTakesAnyTechInSight(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Resources.resources["knowledge"].Storage = 1e9
	// The next age is in sight on a first run: Pottery brings Fire Mastery.
	if chain, err := ge.PlanAddResearchChain("pottery"); err != nil || !reflect.DeepEqual(chain, []string{"fire_mastery", "pottery"}) {
		t.Fatalf("Pottery from the Primitive Age: %v, %v", chain, err)
	}
	// The Classical Age is not. The refusal reads like a mistyped key and
	// suggests nothing from a hidden age.
	err := ge.PlanAddResearch("philosophy")
	if err == nil || err.Error() != "No tech 'philosophy' in sight. Type research to see the tree." {
		t.Fatalf("a tech of a hidden age: %v", err)
	}
	if err := ge.PlanAddResearch("philosophi"); err == nil || strings.Contains(err.Error(), "philosophy") {
		t.Errorf("a near miss on a hidden tech names it: %v", err)
	}
	if err := ge.PlanAddResearch("potery"); err == nil || !strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("a near miss on a tech in sight gets no suggestion: %v", err)
	}
	for _, name := range []string{"Philosophy", "Mathematics", "Classical"} {
		if strings.Contains(err.Error(), name) {
			t.Errorf("the refusal names %s: %v", name, err)
		}
	}

	// Known ground: an age reached before is in sight, and one command
	// queues the chain across the ages between.
	ge.Stats.AgesReached = append(ge.Stats.AgesReached, "stone_age", "bronze_age", "iron_age", "classical_age")
	chain, err := ge.PlanAddResearchChain("philosophy")
	if want := []string{"language", "primitive_writing", "mathematics", "philosophy"}; err != nil || !reflect.DeepEqual(chain, want) {
		t.Fatalf("Philosophy on known ground: %v, %v; want %v", chain, err, want)
	}
	notes := map[string]string{}
	for _, v := range ge.planViews() {
		notes[v.Key] = v.Note
	}
	if notes["philosophy"] != "waits for the Classical Age" || notes["mathematics"] != "waits for the Iron Age" || notes["pottery"] != "waits for the Stone Age" {
		t.Errorf("what the later techs wait for: %v", notes)
	}
}

// TestPlanTechOfALaterAgeWaitsItsTurn: a tech of an age not reached yet
// stays in the plan and waits. It holds neither the research slot's turn
// nor its knowledge, so this age's techs planned below it run, and it
// starts in plan order once its age comes.
func TestPlanTechOfALaterAgeWaitsItsTurn(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Resources.resources["knowledge"].Storage = 1e9
	for _, k := range []string{"pottery", "tool_making"} { // Fire Mastery, Pottery (Stone Age), Tool Making
		if err := ge.PlanAddResearch(k); err != nil {
			t.Fatalf("plan %s: %v", k, err)
		}
	}
	price := func(k string) float64 { d, _ := ge.rules.Tech(k); return d.Cost }
	// Knowledge for the two techs of this age and no more: Pottery, waiting
	// for its age above Tool Making, holds none of it back.
	setAmount(ge, "knowledge", price("fire_mastery")+price("tool_making"))
	for _, want := range []string{"fire_mastery", "tool_making"} {
		ge.runPlanTick()
		if ge.Research.currentTech != want {
			t.Fatalf("researching %q, want %s; views %+v", ge.Research.currentTech, want, ge.planViews())
		}
		researchOut(ge)
	}
	ge.runPlanTick()
	if got := researchKeys(ge); !reflect.DeepEqual(got, []string{"pottery"}) || ge.Research.currentTech != "" {
		t.Fatalf("Pottery should wait for its age: plan %v, researching %q", got, ge.Research.currentTech)
	}
	if v := ge.planViews(); v[0].Status != PlanStatusBlocked || v[0].Note != "waits for the Stone Age" {
		t.Errorf("view = %+v, want Pottery waiting for the Stone Age", v[0])
	}
	for _, l := range ge.log {
		if strings.HasPrefix(l.Message, "Plan: dropped") {
			t.Errorf("the plan dropped something: %s", l.Message)
		}
	}
	// Its age comes: it starts.
	if err := ge.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "knowledge", price("pottery"))
	ge.runPlanTick()
	if ge.Research.currentTech != "pottery" {
		t.Errorf("in the Stone Age: researching %q, want pottery; views %+v", ge.Research.currentTech, ge.planViews())
	}
}
