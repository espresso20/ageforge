package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// stoneAgeGame is a new game in the Stone Age, whose wonder, the Great
// Monolith, waits for Stoneworking.
func stoneAgeGame(t *testing.T) *game.GameEngine {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	if err := ge.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	return ge
}

// TestWonderViewsShowTheKeystone: wherever the game shows what a wonder
// needs, it shows the wonder's keystone tech and whether it is researched:
// the Wonders panel (the summary and the list), the `wonder` command, and
// the splash that announces the wonder. Before the research the line is red
// and gives the command; after it, green.
func TestWonderViewsShowTheKeystone(t *testing.T) {
	ge := stoneAgeGame(t)
	missing := "✗ Keystone: Stoneworking, not researched (research stoneworking)"
	done := "✓ Keystone: Stoneworking, researched"

	panel := wondersProvider(ge.GetState(), 120)
	if n := strings.Count(panel, missing); n != 2 {
		t.Errorf("the Wonders panel shows the missing keystone %d times, want it in the summary and in the list:\n%s", n, panel)
	}
	cmd := cmdWonder(nil, ge)
	if !strings.Contains(cmd.Message, missing) {
		t.Errorf("`wonder` does not show the missing keystone:\n%s", cmd.Message)
	}
	splash := buildAgeSplashText("stone_age", game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
	if want := "Bank its cost and research Stoneworking, its keystone, then build it."; !strings.Contains(splash, want) {
		t.Errorf("the Stone Age splash does not say %q:\n%s", want, splash)
	}
	if strings.Contains(splash, "Opened later by research:[-] Great Monolith") || strings.Contains(splash, "Great Monolith (Stoneworking)") {
		t.Errorf("the splash lists the wonder among the buildings a tech opens later:\n%s", splash)
	}
	// An age whose wonder needs no tech says what it always said.
	if s := buildAgeSplashText("bronze_age", game.AgeAdvanceSummary{}, false, game.EpochEventRecord{}); !strings.Contains(s, "[white]Bank its cost, then build it.[-]") {
		t.Errorf("the Bronze Age splash, whose wonder has no keystone:\n%s", s)
	}

	ge.GrantTechsForTest("tool_making", "stoneworking")
	panel = wondersProvider(ge.GetState(), 120)
	if strings.Contains(panel, "✗ Keystone") || strings.Count(panel, done) != 2 {
		t.Errorf("with Stoneworking researched the Wonders panel should show it done, twice:\n%s", panel)
	}
	if cmd := cmdWonder(nil, ge); !strings.Contains(cmd.Message, done) {
		t.Errorf("`wonder` does not show the keystone as researched:\n%s", cmd.Message)
	}
}

// TestFullBankSaysWhatIsStillMissing: a full bank tells the player to build,
// unless the keystone is still to research: then it says so, in every view.
func TestFullBankSaysWhatIsStillMissing(t *testing.T) {
	ge := stoneAgeGame(t)
	def, _ := ge.Rules().Building("great_monolith")
	for res, need := range def.BaseCost {
		ge.SetStockForTest(res, need)
		if _, err := ge.BankWonderMax("great_monolith", res); err != nil {
			t.Fatal(err)
		}
	}
	st := ge.GetState()
	if bs := st.Buildings["great_monolith"]; !bs.WonderBankFull || bs.CanBuild {
		t.Fatalf("setup: bank full %v, can build %v; want a full bank that cannot be built yet", bs.WonderBankFull, bs.CanBuild)
	}
	panel := wondersProvider(st, 120)
	for _, want := range []string{
		"✓ The bank is full. Research Stoneworking, then build it with: build great_monolith",
		"(bank full, waiting for its keystone)",
	} {
		if !strings.Contains(panel, want) {
			t.Errorf("the Wonders panel does not say %q:\n%s", want, panel)
		}
	}
	if strings.Contains(panel, "ready to build") {
		t.Errorf("the Wonders panel calls a wonder without its keystone ready to build:\n%s", panel)
	}
	if cmd := cmdWonder(nil, ge); !strings.Contains(cmd.Message, "Bank full. Research Stoneworking, then type 'build great_monolith' to start construction.") {
		t.Errorf("`wonder` with a full bank and no keystone:\n%s", cmd.Message)
	}

	ge.GrantTechsForTest("tool_making", "stoneworking")
	panel = wondersProvider(ge.GetState(), 120)
	for _, want := range []string{"✓ The bank is full. Build it with: build great_monolith", "(bank full, ready to build)"} {
		if !strings.Contains(panel, want) {
			t.Errorf("with the keystone researched the Wonders panel does not say %q:\n%s", want, panel)
		}
	}
	if cmd := cmdWonder(nil, ge); !strings.Contains(cmd.Message, "Bank full. Type 'build great_monolith' to start construction.") {
		t.Errorf("`wonder` with a full bank and the keystone:\n%s", cmd.Message)
	}
}

