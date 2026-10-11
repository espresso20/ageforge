package smoke

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// botIn is a bot on a new game moved into age, with the given techs
// researched. Nothing is played: the tests below read what the bot would do
// from one snapshot.
func botIn(t *testing.T, age string, researched ...string) (*game.GameEngine, *Bot) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	ge := game.NewGameEngine()
	ge.SeedRNG(5)
	for _, a := range ge.Rules().AgeKeys()[1:] {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
		if a == age {
			break
		}
	}
	if len(researched) > 0 {
		ge.GrantTechsForTest(researched...)
	}
	return ge, NewBot(ge)
}

// finish ends the research in progress, which must be key, as its last
// tick would.
func finish(t *testing.T, ge *game.GameEngine, key string) {
	t.Helper()
	if got := ge.GetState().Research.CurrentTech; got != key {
		t.Fatalf("researching %q, want %s", got, key)
	}
	if err := ge.CancelResearch(); err != nil {
		t.Fatal(err)
	}
	ge.GrantTechsForTest(key)
}

// giveKnowledge sets what the bot holds of knowledge, under a cap that fits.
func giveKnowledge(ge *game.GameEngine, v float64) {
	ge.SetStockForTest("knowledge", v)
}

// TestBotResearchesTheKeystoneFirst: the bot's research order. In the Stone
// Age the Great Monolith waits for Stoneworking, which stands on Tool
// Making, so those two come first, in that order, and their knowledge is a
// target of the age. While the keystone is out of reach the bot saves for it
// and starts nothing else, however cheap. Once the keystone is researched
// the rest goes cheapest first.
func TestBotResearchesTheKeystoneFirst(t *testing.T) {
	ge, b := botIn(t, "stone_age")
	techs := ge.GetState().Research.Techs
	tool, stone, writing := techs["tool_making"].Cost, techs["stoneworking"].Cost, techs["primitive_writing"].Cost
	if !(writing < stone) {
		t.Fatalf("setup: Primitive Writing (%v) should be cheaper than Stoneworking (%v), or saving for the keystone is not tested", writing, stone)
	}

	p := b.newPlan(ge.GetState())
	if want := []string{"tool_making", "stoneworking"}; !reflect.DeepEqual(p.must, want) {
		t.Fatalf("what the Stone Age cannot be left without: %v, want %v", p.must, want)
	}
	if got := p.hold["knowledge"]; got != tool+stone {
		t.Errorf("the bot holds back %v knowledge, want the %v the two techs cost", got, tool+stone)
	}
	if p.target["knowledge"] < tool+stone || p.capNeed["knowledge"] < stone {
		t.Errorf("knowledge target %v and cap need %v, want at least %v and %v", p.target["knowledge"], p.capNeed["knowledge"], tool+stone, stone)
	}

	// Enough for Tool Making only: it starts, the first of the chain.
	giveKnowledge(ge, tool)
	b.research(b.newPlan(ge.GetState()))
	if got := ge.GetState().Research.CurrentTech; got != "tool_making" {
		t.Fatalf("with %v knowledge the bot is researching %q, want tool_making", tool, got)
	}
	finish(t, ge, "tool_making")

	// Enough for Primitive Writing but not for Stoneworking: it waits.
	giveKnowledge(ge, stone-1)
	p = b.newPlan(ge.GetState())
	if want := []string{"stoneworking"}; !reflect.DeepEqual(p.must, want) {
		t.Fatalf("after Tool Making: %v, want %v", p.must, want)
	}
	b.research(p)
	if got := ge.GetState().Research.CurrentTech; got != "" {
		t.Fatalf("one knowledge short of the keystone the bot started %q instead of saving", got)
	}
	// And the wonder bank does not take the knowledge it is saving. (The
	// Great Monolith costs none; the rule is the same for one that does.)
	if got := p.hold["knowledge"]; got != stone {
		t.Errorf("the bot holds back %v knowledge, want Stoneworking's %v", got, stone)
	}

	giveKnowledge(ge, stone)
	b.research(b.newPlan(ge.GetState()))
	if got := ge.GetState().Research.CurrentTech; got != "stoneworking" {
		t.Fatalf("with the keystone's price in hand the bot is researching %q", got)
	}
	// In progress it is paid for: nothing is held back for it.
	if p := b.newPlan(ge.GetState()); p.hold["knowledge"] != 0 || len(p.must) != 1 {
		t.Errorf("with the keystone in progress: holding %v, must %v", p.hold["knowledge"], p.must)
	}
	finish(t, ge, "stoneworking")

	// Nothing the age cannot be left without is left: the cheapest goes next.
	giveKnowledge(ge, 1e6)
	p = b.newPlan(ge.GetState())
	if len(p.must) != 0 || p.hold["knowledge"] != 0 {
		t.Fatalf("with the keystone researched: must %v, holding %v", p.must, p.hold["knowledge"])
	}
	b.research(p)
	if got := ge.GetState().Research.CurrentTech; got != "language" {
		t.Errorf("after the keystone the bot is researching %q, want the cheapest (language, a spine tech of the first age)", got)
	}
}

// TestBotCountsATechOnlySourceAsRequired: in the Modern Age the Information
// Age asks for steel and no Modern building makes it: Steel Forging's flat
// output is its only source (Satellite Tech's data is the other). So the bot
// must research them too. Every gate now asks for the raw materials, so the
// Steel Forging chain and Satellite Tech are on the Modern wonder's keystone
// chain already; the second half takes the wonder and the required
// buildings away, so that what is left is only what the resources ask.
func TestBotCountsATechOnlySourceAsRequired(t *testing.T) {
	ge, b := botIn(t, "modern_age")
	p := b.newPlan(ge.GetState())
	if p.st.NextAgeResReqs["steel"] <= 0 || b.madeHere(p, "steel") {
		t.Fatalf("setup: the Information Age asks for %v steel, a Modern building makes it: %v", p.st.NextAgeResReqs["steel"], b.madeHere(p, "steel"))
	}
	for _, key := range []string{"steel_forging", "satellite_tech"} {
		if !b.unblocked(p, key, true) {
			t.Fatalf("setup: %s should be the only source of a resource the Information Age asks for", key)
		}
	}
	for _, key := range []string{"steel_forging", "satellite_tech"} {
		if !slices.Contains(p.must, key) {
			t.Errorf("the whole age cannot be left without %s: %v", key, p.must)
		}
	}
	// With nothing else required, the tech-only sources stand alone, each
	// after the chain it stands on (by the cost of the tech, cheapest first).
	p.st.CurrentAgeWonderKey = ""
	p.needBld = nil
	got := b.mustTechs(p)
	want := []string{"tool_making", "stoneworking", "bronze_working", "iron_smelting", "steel_forging"}
	if len(got) < len(want) || !reflect.DeepEqual(got[:len(want)], want) {
		t.Errorf("what only the resources ask for: %v, want it to begin %v", got, want)
	}
	if !slices.Contains(got, "satellite_tech") {
		t.Errorf("what only the resources ask for: %v, want Satellite Tech", got)
	}
	// Satellite Tech stands on the whole chain behind it (aviation and
	// rocketry among it), and nothing else is counted: the Information Age's
	// techs make nothing the Information Age asks for in the Modern Age.
	for _, key := range got {
		if key == "computers" || key == "internet" {
			t.Errorf("%s is counted as required for a resource", key)
		}
	}
}

// TestBotTakesAnOpenerBeforeTheCheapest: once nothing is required, a tech
// that opens a building of this age goes ahead of a cheaper tech that opens
// nothing. Civilian Reactors opens the Atomic Age's Nuclear Plant.
func TestBotTakesAnOpenerBeforeTheCheapest(t *testing.T) {
	ge, b := botIn(t, "atomic_age")
	p := b.newPlan(ge.GetState())
	order := b.restOrder(p)
	at := map[string]int{}
	for i, k := range order {
		at[k] = i
	}
	techs := p.st.Research.Techs
	if !(techs["fire_mastery"].Cost < techs["civilian_reactors"].Cost) {
		t.Fatal("setup: Fire Mastery should be cheaper than Civilian Reactors")
	}
	if !b.openers(p)["civilian_reactors"] {
		t.Fatalf("Civilian Reactors is not what a building of the Atomic Age waits for: %v", b.openers(p))
	}
	if at["civilian_reactors"] > at["fire_mastery"] {
		t.Errorf("Civilian Reactors comes after Fire Mastery in %v", order[:8])
	}
	if at["fire_mastery"] > at["radio"] {
		t.Errorf("among techs that open nothing, the dearer Radio comes before Fire Mastery")
	}
}

// TestTechChainTakesTheCheaperBranch: the chain behind a tech lists what it
// stands on first, skips what is researched, and for an either-or group
// nothing satisfies takes the branch that costs less to finish. A group one
// member of which is researched adds nothing.
func TestTechChainTakesTheCheaperBranch(t *testing.T) {
	src := rules.FromConfig()
	age := src.Ages[1].Key
	src.Techs = append(src.Techs,
		config.TechDef{Key: "zz_root", Age: age},
		config.TechDef{Key: "zz_cheap", Age: age},
		config.TechDef{Key: "zz_dear_base", Age: age},
		config.TechDef{Key: "zz_dear", Age: age, Prerequisites: []string{"zz_dear_base"}},
		config.TechDef{Key: "zz_join", Age: age, Prerequisites: []string{"zz_root"}, AnyOf: []string{"zz_dear", "zz_cheap"}},
		config.TechDef{Key: "zz_top", Age: age, Prerequisites: []string{"zz_join"}},
	)
	set := rules.Compile(src)
	state := func(researched ...string) map[string]game.TechState {
		techs := map[string]game.TechState{
			"zz_root": {Cost: 10}, "zz_cheap": {Cost: 50}, "zz_dear_base": {Cost: 30}, "zz_dear": {Cost: 30},
			"zz_join": {Cost: 5}, "zz_top": {Cost: 5},
		}
		for _, k := range researched {
			ts := techs[k]
			ts.Researched = true
			techs[k] = ts
		}
		return techs
	}
	for _, c := range []struct {
		researched []string
		want       []string
		why        string
	}{
		{nil, []string{"zz_root", "zz_cheap", "zz_join", "zz_top"}, "the cheap branch costs 50, the dear one 60"},
		{[]string{"zz_dear_base"}, []string{"zz_root", "zz_dear", "zz_join", "zz_top"}, "with its base researched the dear branch costs 30"},
		{[]string{"zz_dear"}, []string{"zz_root", "zz_join", "zz_top"}, "one member of the group is researched"},
		{[]string{"zz_root", "zz_cheap", "zz_join"}, []string{"zz_top"}, "everything below is researched"},
		{[]string{"zz_top"}, nil, "the tech itself is researched"},
	} {
		if got := techChain(set, state(c.researched...), "zz_top", map[string]bool{}, nil); !reflect.DeepEqual(got, c.want) {
			t.Errorf("with %v researched the chain is %v, want %v (%s)", c.researched, got, c.want, c.why)
		}
	}
	if got := techChain(set, state(), "no_such_tech", map[string]bool{}, nil); got != nil {
		t.Errorf("an unknown tech has the chain %v", got)
	}
}

// TestBlockersNameTheKeystone: the line a soft-lock or a timeout prints says
// which tech the wonder waits for, what it costs and the knowledge cap, and
// an age whose keystone no reachable storage can hold is an invariant
// violation, not a slow age.
func TestBlockersNameTheKeystone(t *testing.T) {
	ge, _ := botIn(t, "stone_age")
	// Stores a player has by the Stone Age, so the age's storage can be built.
	ge.SetStockForTest("wood", 2000)
	ge.SetStockForTest("stone", 2000)
	st := ge.GetState()
	if got := keystoneChain(st); !reflect.DeepEqual(got, []string{"tool_making", "stoneworking"}) {
		t.Fatalf("the Great Monolith waits for %v", got)
	}
	line := Blockers(st)
	for _, want := range []string{"wonder great_monolith", "needs tool_making (", "needs stoneworking ("} {
		if !strings.Contains(line, want) {
			t.Errorf("blockers %q do not name %q", line, want)
		}
	}
	// Reachable storage is far above Stoneworking's price in the real game.
	for _, p := range storageProblems(st, st.Ruleset().BuildingMap()) {
		if p.check == "keystone_over_storage" {
			t.Errorf("a new Stone Age game: %s", p.msg)
		}
	}
	// A keystone priced over anything the age can store is caught.
	ts := st.Research.Techs["stoneworking"]
	ts.Cost = 1e30
	st.Research.Techs["stoneworking"] = ts
	found := false
	for _, p := range storageProblems(st, st.Ruleset().BuildingMap()) {
		found = found || p.check == "keystone_over_storage"
	}
	if !found {
		t.Error("a keystone no storage can hold was not flagged")
	}
	ge.GrantTechsForTest("tool_making", "stoneworking")
	if st := ge.GetState(); keystoneChain(st) != nil || strings.Contains(Blockers(st), "needs ") {
		t.Errorf("with the keystone researched the wonder still waits for %v", keystoneChain(st))
	}
}
