package game

import (
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// treeSet is the core tables plus five techs of the first age, cheap and
// quick, with the shapes no real tech has yet:
//
//	tree_base, tree_left, tree_right   need nothing
//	tree_join                          needs tree_base, and tree_left or tree_right
//	tree_method                        research takes half the time
func treeSet() *rules.Set {
	src := rules.FromConfig()
	tech := func(key, name string, def config.TechDef) config.TechDef {
		def.Key, def.Name, def.Age, def.Lane = key, name, src.Ages[0].Key, config.LaneKnowledge
		def.Cost, def.ResearchTicks = 10, 4
		return def
	}
	src.Techs = append(src.Techs,
		tech("tree_base", "Base", config.TechDef{}),
		tech("tree_left", "Left", config.TechDef{}),
		tech("tree_right", "Right", config.TechDef{}),
		tech("tree_join", "Join", config.TechDef{Prerequisites: []string{"tree_base"}, AnyOf: []string{"tree_left", "tree_right"}}),
		tech("tree_method", "Method", config.TechDef{Effects: []config.TechEffect{{Kind: config.EffectResearchTime, Value: 0.5}}}),
	)
	return rules.Compile(src)
}

// treeEngine is a new game on treeSet with knowledge to spend.
func treeEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := NewGameEngineWith(treeSet())
	ge.SeedRNG(3)
	ge.Resources.resources["knowledge"].Storage = 1e6
	setAmount(ge, "knowledge", 1000)
	return ge
}

// learn marks techs researched, as finishing them would.
func learn(ge *GameEngine, keys ...string) {
	for _, k := range keys {
		ge.Research.researched[k] = true
	}
	ge.Research.rebuildBonuses()
}

// researchOut runs the research in progress to its end.
func researchOut(ge *GameEngine) {
	for i := 0; i < 1_000_000 && ge.Research.currentTech != ""; i++ {
		ge.processResearch()
	}
}

// TestStartResearchWithAnEitherOrGroup: a tech with an either-or group
// starts once every prerequisite is researched and one key of the group is,
// whichever one. The refusal names what is missing, and the snapshot the UI
// reads agrees with the command.
func TestStartResearchWithAnEitherOrGroup(t *testing.T) {
	ge := treeEngine(t)
	join := func() TechState { return ge.GetState().Research.Techs["tree_join"] }

	if ts := join(); !reflect.DeepEqual(ts.Prerequisites, []string{"tree_base"}) || !reflect.DeepEqual(ts.AnyOf, []string{"tree_left", "tree_right"}) {
		t.Errorf("the snapshot lists prerequisites %v and either-or %v", ts.Prerequisites, ts.AnyOf)
	}
	err := ge.StartResearch("tree_join")
	if err == nil || err.Error() != "Join needs Base researched first." {
		t.Errorf("with nothing researched: %v", err)
	}
	// A key of the group does not stand in for a prerequisite.
	learn(ge, "tree_left")
	if err := ge.StartResearch("tree_join"); err == nil || err.Error() != "Join needs Base researched first." {
		t.Errorf("with one branch and no Base: %v", err)
	}
	if ts := join(); ts.PrereqsMet || ts.Available {
		t.Errorf("with one branch and no Base the snapshot says met %v, available %v", ts.PrereqsMet, ts.Available)
	}

	for _, branch := range []string{"tree_left", "tree_right"} {
		ge := treeEngine(t)
		learn(ge, "tree_base")
		err := ge.StartResearch("tree_join")
		if err == nil || err.Error() != "Join needs Left or Right researched first." {
			t.Errorf("with Base and no branch: %v", err)
		}
		if ts := ge.GetState().Research.Techs["tree_join"]; ts.PrereqsMet || ts.Available {
			t.Errorf("with Base and no branch the snapshot says met %v, available %v", ts.PrereqsMet, ts.Available)
		}
		learn(ge, branch)
		if ts := ge.GetState().Research.Techs["tree_join"]; !ts.PrereqsMet || !ts.Available {
			t.Errorf("with Base and %s the snapshot says met %v, available %v", branch, ts.PrereqsMet, ts.Available)
		}
		if err := ge.StartResearch("tree_join"); err != nil {
			t.Errorf("with Base and %s: %v", branch, err)
		}
		researchOut(ge)
		if !ge.Research.IsResearched("tree_join") {
			t.Errorf("Join did not finish after starting with %s", branch)
		}
	}
}

// TestPlanTakesAnEitherOrGroup: the plan refuses a tech until what it needs
// is researched, in progress or planned above it, where one key of an
// either-or group is enough, and then runs the chain in order.
func TestPlanTakesAnEitherOrGroup(t *testing.T) {
	ge := treeEngine(t)
	refusal := func() string {
		err := ge.PlanAddResearch("tree_join")
		if err == nil {
			t.Fatal("the plan took Join too early")
		}
		return err.Error()
	}
	if got, want := refusal(), "Can't plan Join: it needs Base first, which isn't researched or planned before it."; got != want {
		t.Errorf("no Base: %q, want %q", got, want)
	}
	if err := ge.PlanAddResearch("tree_base"); err != nil {
		t.Fatal(err)
	}
	if got, want := refusal(), "Can't plan Join: it needs Left or Right first, and none of them is researched or planned before it."; got != want {
		t.Errorf("no branch: %q, want %q", got, want)
	}
	for _, k := range []string{"tree_right", "tree_join"} {
		if err := ge.PlanAddResearch(k); err != nil {
			t.Fatalf("plan %s: %v", k, err)
		}
	}
	for _, want := range []string{"tree_base", "tree_right", "tree_join"} {
		ge.runPlanTick()
		if ge.Research.currentTech != want {
			t.Fatalf("researching %q, want %s; plan %+v", ge.Research.currentTech, want, ge.plan)
		}
		researchOut(ge)
	}
	if len(ge.plan) != 0 || !ge.Research.IsResearched("tree_join") || ge.Research.IsResearched("tree_left") {
		t.Errorf("after the chain: plan %+v, Join researched %v, Left researched %v", ge.plan, ge.Research.IsResearched("tree_join"), ge.Research.IsResearched("tree_left"))
	}

	// A branch already researched or in progress counts too.
	ge = treeEngine(t)
	learn(ge, "tree_base")
	if err := ge.StartResearch("tree_left"); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("tree_join"); err != nil {
		t.Errorf("with Left in progress the plan refused Join: %v", err)
	}
}

// TestPlanTechWaitsForWhatItNeeds: a tech in the plan whose need is no
// longer on its way is not dropped. It waits, it says what for, and while it
// waits it holds neither the research slot's turn nor its knowledge, so the
// missing tech can be planned below it and still go first.
func TestPlanTechWaitsForWhatItNeeds(t *testing.T) {
	ge := treeEngine(t)
	for _, k := range []string{"tree_base", "tree_left", "tree_join"} {
		if err := ge.PlanAddResearch(k); err != nil {
			t.Fatalf("plan %s: %v", k, err)
		}
	}
	// Take away the only branch the plan held.
	if _, err := ge.PlanRemove(2); err != nil {
		t.Fatal(err)
	}
	ge.runPlanTick()
	if ge.Research.currentTech != "tree_base" {
		t.Fatalf("researching %q, want tree_base", ge.Research.currentTech)
	}
	researchOut(ge)
	ge.runPlanTick()
	if len(ge.plan) != 1 || ge.plan[0].Key != "tree_join" || ge.Research.currentTech != "" {
		t.Fatalf("Join should wait: plan %+v, researching %q", ge.plan, ge.Research.currentTech)
	}
	if v := ge.planViews(); len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "needs Left or Right first" {
		t.Errorf("views = %+v, want Join blocked on Left or Right", v)
	}

	// The missing branch goes in below Join. With knowledge for one tech
	// only, it is the one that starts: Join held nothing back.
	if err := ge.PlanAddResearch("tree_right"); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "knowledge", 10)
	if v := ge.planViews(); len(v) != 2 || v[0].Note != "needs Left or Right first" || v[1].Status != PlanStatusReady {
		t.Errorf("views = %+v, want Join waiting and Right ready below it", v)
	}
	ge.runPlanTick()
	if ge.Research.currentTech != "tree_right" {
		t.Fatalf("researching %q, want tree_right, the branch planned below the waiting Join", ge.Research.currentTech)
	}
	// With its branch in progress Join is an ordinary queued tech again.
	setAmount(ge, "knowledge", 10)
	if v := ge.planViews(); len(v) != 1 || v[0].Note != "research slot busy" {
		t.Errorf("views = %+v, want Join waiting on the slot", v)
	}
	researchOut(ge)
	ge.runPlanTick()
	if ge.Research.currentTech != "tree_join" || len(ge.plan) != 0 {
		t.Errorf("after Right: researching %q, plan %+v", ge.Research.currentTech, ge.plan)
	}
	for _, l := range ge.log {
		if strings.HasPrefix(l.Message, "Plan: dropped") {
			t.Errorf("the plan dropped something: %s", l.Message)
		}
	}
}

// TestResearchBonusesAreKeyedByKindAndTarget: gold output, gold storage and
// gold a tick are three things, summed apart, and the pools the rest of the
// game reads hold the fractions only.
func TestResearchBonusesAreKeyedByKindAndTarget(t *testing.T) {
	ge := NewGameEngine()
	// Banking: +50% gold, +100 gold storage. Alchemy: +15% knowledge, +0.1
	// gold a tick. Pottery: +25 storage for every resource. Feudalism: +5
	// housing.
	learn(ge, "banking", "alchemy", "pottery", "feudalism")
	rm := ge.Research
	for _, c := range []struct {
		kind   config.TechEffectKind
		target string
		want   float64
	}{
		{config.EffectOutput, "gold", 0.5},
		{config.EffectFlatStorage, "gold", 100},
		{config.EffectFlatOutput, "gold", 0.1},
		{config.EffectOutput, "knowledge", 0.15},
		{config.EffectFlatStorage, config.AllResources, 25},
		{config.EffectFlatHousing, "", 5},
		{config.EffectAllOutput, "", 0},
		{config.EffectFlatStorage, "knowledge", 0},
	} {
		if got := rm.Bonus(c.kind, c.target); got != c.want {
			t.Errorf("Bonus(%s, %q) = %v, want %v", c.kind, c.target, got, c.want)
		}
	}
	pools := map[string]float64{"gold_rate": 0.5, "knowledge_rate": 0.15}
	if got := rm.GetBonuses(); !reflect.DeepEqual(got, pools) {
		t.Errorf("pools = %v, want %v", got, pools)
	}
	mods := map[string]float64{}
	for _, m := range rm.Modifiers() {
		if m.Source != "research" || m.Op != OpAdd {
			t.Errorf("modifier %+v is not an added research bonus", m)
		}
		mods[m.Target] = m.Value
	}
	if !reflect.DeepEqual(mods, pools) {
		t.Errorf("modifiers = %v, want one per pool: %v", mods, pools)
	}
	rs := ge.GetState().Research
	if !reflect.DeepEqual(rs.Bonuses, pools) ||
		!reflect.DeepEqual(rs.Flat, map[string]float64{"gold": 0.1}) ||
		!reflect.DeepEqual(rs.Storage, map[string]float64{"gold": 100, "all": 25}) || rs.Housing != 5 {
		t.Errorf("the snapshot holds bonuses %v, flat %v, storage %v, housing %v", rs.Bonuses, rs.Flat, rs.Storage, rs.Housing)
	}
}

// TestResearchTimeIsAKindATechCanCarry: no tech shortens research today, but
// the kind is there and the engine applies it: with a tech that takes half
// off research time, the next research takes half as long.
func TestResearchTimeIsAKindATechCanCarry(t *testing.T) {
	ge := treeEngine(t)
	if err := ge.StartResearch("tree_base"); err != nil {
		t.Fatal(err)
	}
	full := ge.Research.totalTicks
	researchOut(ge)
	learn(ge, "tree_method")
	if got := ge.Research.Bonus(config.EffectResearchTime, ""); got != 0.5 {
		t.Fatalf("research time bonus = %v, want 0.5", got)
	}
	if err := ge.StartResearch("tree_left"); err != nil {
		t.Fatal(err)
	}
	if got := ge.Research.totalTicks; full != 4 || got != 2 {
		t.Errorf("research took %d ticks before the tech and %d after, want 4 and 2", full, got)
	}
}
