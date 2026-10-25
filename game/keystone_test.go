package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// stoneAgeEngine is a new game moved into the Stone Age: the first age whose
// wonder, the Great Monolith, has a keystone (Stoneworking).
func stoneAgeEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(3)
	noErr(t, ge.EnterAgeForTest("stone_age"))
	if ge.Research.IsResearched("stoneworking") || !ge.Buildings.TechLocked("great_monolith") {
		t.Fatal("setup: the Great Monolith does not wait for Stoneworking")
	}
	return ge
}

// TestWonderWaitsForItsKeystone: an age's wonder needs its keystone tech,
// and only to be built. Its bank is open from the first tick of the age:
// deposits and overflow fill it while the tech is still to come. With the
// bank full, `build` is refused with the tech's name, the advance says the
// same, and once the tech is researched the wonder builds.
func TestWonderWaitsForItsKeystone(t *testing.T) {
	ge := stoneAgeEngine(t)
	if !logHas(ge, "★ Wonder available: Great Monolith. Bank its cost and research Stoneworking, its keystone, then build it.") {
		t.Error("entering the Stone Age did not say the Great Monolith needs Stoneworking")
	}
	bs := ge.GetState().Buildings["great_monolith"]
	if !bs.Unlocked || bs.NeedsTech != "stoneworking" || bs.CanBuild {
		t.Fatalf("the Great Monolith: unlocked %v, needs %q, can be built %v; want open for banking, waiting for stoneworking", bs.Unlocked, bs.NeedsTech, bs.CanBuild)
	}

	// A deposit goes in.
	setAmount(ge, "wood", 400)
	if got, err := ge.BankWonderResource("great_monolith", "wood", 300); err != nil || got != 300 {
		t.Fatalf("a deposit while the keystone is missing: banked %v, %v", got, err)
	}
	if got := ge.Buildings.wonderBanks["great_monolith"]["wood"]; got != 300 {
		t.Errorf("the bank holds %v wood, want the 300 deposited", got)
	}

	// A full bank says what is still missing, and building is refused.
	fillWonderBank(t, ge, "great_monolith")
	if !logHas(ge, "Great Monolith is fully banked. Research Stoneworking, then type 'build great_monolith' to start construction.") {
		t.Error("the full bank did not say to research Stoneworking first")
	}
	if bs := ge.GetState().Buildings["great_monolith"]; !bs.WonderBankFull || bs.CanBuild {
		t.Errorf("with a full bank and no keystone: bank full %v, can be built %v", bs.WonderBankFull, bs.CanBuild)
	}
	err := ge.BuildBuilding("great_monolith")
	if err == nil || err.Error() != "Great Monolith needs Stoneworking first. Research it to build here." {
		t.Errorf("building without the keystone: %v", err)
	}
	if n, err := ge.BuildMultiple("great_monolith", 1); err == nil || n != 0 {
		t.Errorf("building one through the batch path without the keystone: built %d, %v", n, err)
	}
	if ge.Buildings.Build("great_monolith", ge.Resources) {
		t.Error("the building manager built the wonder without its keystone")
	}
	err = ge.AdvanceAge()
	if err == nil || !strings.Contains(err.Error(), "research Stoneworking, its keystone") {
		t.Errorf("advancing without the wonder: %v", err)
	}
	if !ge.Buildings.IsWonderBankFull("great_monolith") {
		t.Error("a refused build emptied the bank")
	}

	// The keystone arrives through research, and says what it opened.
	learn(ge, "tool_making")
	noErr(t, ge.StartResearch("stoneworking"))
	researchOut(ge)
	if !logHas(ge, "With Stoneworking researched, Great Monolith can be built once its bank is full.") {
		t.Error("finishing the keystone did not say what it opened")
	}
	if bs := ge.GetState().Buildings["great_monolith"]; bs.NeedsTech != "" || !bs.CanBuild {
		t.Errorf("with the keystone: needs %q, can be built %v", bs.NeedsTech, bs.CanBuild)
	}
	if err := ge.BuildBuilding("great_monolith"); err != nil {
		t.Fatalf("building with the keystone researched: %v", err)
	}
}

// TestPlannedWonderWaitsForItsKeystone: a wonder in the build plan waits
// for its keystone like any building behind a tech, says so, and starts by
// itself once the tech is researched.
func TestPlannedWonderWaitsForItsKeystone(t *testing.T) {
	ge := stoneAgeEngine(t)
	if _, err := ge.PlanAddBuild("great_monolith", 1); err != nil {
		t.Fatalf("planning the wonder before its keystone: %v", err)
	}
	fillWonderBank(t, ge, "great_monolith")
	ge.runPlanTick()
	if v := ge.planViews(); len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "waits for Stoneworking" {
		t.Fatalf("the planned wonder: %+v, want blocked, waiting for Stoneworking", v)
	}
	if queued(ge, "great_monolith") != 0 {
		t.Fatal("the plan started the wonder without its keystone")
	}
	learn(ge, "stoneworking")
	ge.runPlanTick()
	if queued(ge, "great_monolith") != 1 || len(ge.plan) != 0 {
		t.Errorf("with the keystone: %d under construction, %d plan items left", queued(ge, "great_monolith"), len(ge.plan))
	}
}

// TestWondersWithoutAKeystone: the Sacred Grove builds as it always did:
// nothing blocks the first age. It is the only wonder with no keystone.
func TestWondersWithoutAKeystone(t *testing.T) {
	ge := newSeededEngine(3)
	bare := 0
	for _, def := range ge.rules.Buildings() {
		if def.Category == "wonder" && def.RequiredTech == "" {
			bare++
		}
	}
	if bare != 1 {
		t.Errorf("%d wonders have no keystone, want the Sacred Grove alone", bare)
	}
	const wonder = "sacred_grove" // a new game starts in the Primitive Age
	if bs := ge.GetState().Buildings[wonder]; !bs.Unlocked || bs.NeedsTech != "" {
		t.Errorf("%s: unlocked %v, needs %q; want open with no tech", wonder, bs.Unlocked, bs.NeedsTech)
	}
	fillWonderBank(t, ge, wonder)
	if err := ge.BuildBuilding(wonder); err != nil {
		t.Errorf("building %s with no tech researched: %v", wonder, err)
	}
}

// TestResearchStateNamesKeystones: the snapshot the panels read carries
// each tech's kind and, for a keystone, the wonder it opens. Nothing else
// is a keystone of anything.
func TestResearchStateNamesKeystones(t *testing.T) {
	ge := NewGameEngine()
	techs := ge.GetState().Research.Techs
	for key, want := range map[string]struct {
		kind   config.TechKind
		wonder string
	}{
		"stoneworking": {config.TechKeystone, "great_monolith"},
		"mathematics":  {config.TechKeystone, "colosseum"},
		"tool_making":  {config.TechSpine, ""},
		"fire_mastery": {config.TechOptional, ""},
	} {
		if got := techs[key]; got.Kind != want.kind || got.KeystoneOf != want.wonder {
			t.Errorf("%s: kind %q, keystone of %q; want %q and %q", key, got.Kind, got.KeystoneOf, want.kind, want.wonder)
		}
	}
	keystones := 0
	for key, ts := range techs {
		if (ts.Kind == config.TechKeystone) != (ts.KeystoneOf != "") {
			t.Errorf("%s is %s but the keystone of %q", key, ts.Kind, ts.KeystoneOf)
		}
		if ts.KeystoneOf != "" {
			keystones++
			if def := ge.Buildings.defs[ts.KeystoneOf]; def.RequiredTech != key || def.RequiredAge != ts.Age {
				t.Errorf("%s (%s) is listed as the keystone of %s, which requires %q in %s", key, ts.Age, ts.KeystoneOf, def.RequiredTech, def.RequiredAge)
			}
		}
	}
	if keystones != 21 {
		t.Errorf("%d keystones in the snapshot, want 21", keystones)
	}
	if got := ge.rules.Keystone("stone_age"); got != "stoneworking" {
		t.Errorf("the Stone Age's keystone is %q", got)
	}
	if got := ge.rules.Keystone("bronze_age"); got != "calendar" {
		t.Errorf("the Bronze Age's keystone is %q", got)
	}
	if got := ge.rules.Keystone("primitive_age"); got != "" {
		t.Errorf("the Primitive Age has a keystone: %q", got)
	}
}
