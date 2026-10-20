package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// newExpeditionTestEngine spins up a fresh engine in the Bronze Age, stocked
// with enough food, wood and soldiers to launch any offer on its mission
// board. The age jump uses the dev console (/age applies the age's unlocks);
// resource amounts are set via LoadAmounts so they bypass nominal storage
// caps.
func newExpeditionTestEngine(t *testing.T) *game.GameEngine {
	t.Helper()

	prevDev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prevDev })

	engine := game.NewGameEngine()
	engine.SeedRNG(11)
	engine.GrantTechsForTest("military_tactics", "exploration") // the techs that open campaigns and expeditions
	if msg := game.DevExecCommand("/age bronze_age", engine); msg != "jumped to bronze_age" {
		t.Fatalf("dev /age bronze_age failed: %q", msg)
	}
	engine.Resources.LoadAmounts(map[string]float64{
		"food":     500,
		"wood":     500,
		"soldiers": 50,
	})
	return engine
}

// TestCmdExpeditionNoArgsPrintsTheBoard verifies bare `expedition` prints the
// mission board as plain lines (nine numbered offers) rather than opening a
// panel or launching anything.
func TestCmdExpeditionNoArgsPrintsTheBoard(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition(nil, engine)
	if res.Type != "info" || res.OverlayName != "" {
		t.Errorf("cmdExpedition(nil) = type %q overlay %q, want a printed board", res.Type, res.OverlayName)
	}
	if !strings.Contains(res.Message, "Mission board") {
		t.Errorf("the board has no heading:\n%s", res.Message)
	}
	for n := 1; n <= 9; n++ {
		if want := string(rune('0'+n)) + ". "; !strings.Contains(res.Message, want) {
			t.Errorf("the board does not list offer %d:\n%s", n, res.Message)
		}
	}
	for _, heading := range []string{"Short:", "Long:", "Grand:"} {
		if !strings.Contains(res.Message, heading) {
			t.Errorf("the board has no %s section:\n%s", heading, res.Message)
		}
	}
	st := engine.GetState()
	if st.Military.ActiveShort != nil || st.Military.ActiveLong != nil || st.Military.ActiveGrand != nil {
		t.Error("printing the board launched something")
	}
	if list := cmdExpedition([]string{"list"}, engine); list.Message != res.Message {
		t.Error("`expedition list` is not the board")
	}
}

// TestCmdExpeditionNumberLaunches verifies `expedition <n>` sends offer n and
// fills its tier's slot; a second offer of the same tier is refused while one
// is out, and another tier is still free.
func TestCmdExpeditionNumberLaunches(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition([]string{"1"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdExpedition(1) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveShort == nil {
		t.Errorf("after sending offer 1, ActiveShort is nil; expected a short mission under way")
	}
	if res := cmdExpedition([]string{"2"}, engine); res.Type != "error" || !strings.Contains(res.Message, "already under way") {
		t.Errorf("a second short mission: type %q, %q; want the slot's refusal", res.Type, res.Message)
	}
	if res := cmdExpedition([]string{"4"}, engine); res.Type != "success" {
		t.Errorf("a long mission beside the short one: type %q, %q", res.Type, res.Message)
	}
	if st := engine.GetState(); st.Military.ActiveLong == nil || st.Military.ActiveShort == nil {
		t.Error("the short and the long mission should both be out")
	}
}

// TestCmdCampaignIsTheSameCommand verifies `campaign` prints the same board
// and sends the same offers as `expedition`.
func TestCmdCampaignIsTheSameCommand(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	if a, b := cmdCampaign(nil, engine), cmdExpedition(nil, engine); a.Message != b.Message || a.Type != "info" {
		t.Errorf("campaign prints %q, expedition %q", a.Message, b.Message)
	}
	res := cmdCampaign([]string{"7"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdCampaign(7) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveGrand == nil || st.Military.ActiveMilitary == nil {
		t.Errorf("after sending offer 7, no grand mission is out: %+v", st.Military)
	}
}

// TestCmdExpeditionBadArguments verifies a word, an offer number off the
// board and a number with more after it are refused, and launch nothing.
func TestCmdExpeditionBadArguments(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	for _, args := range [][]string{{"scouting", "party"}, {"scouting_party"}, {"12"}, {"0"}, {"1", "2"}} {
		res := cmdExpedition(args, engine)
		if res.Type != "error" {
			t.Errorf("cmdExpedition(%v) Type = %q (msg %q), want an error", args, res.Type, res.Message)
		}
	}
	st := engine.GetState()
	if st.Military.ActiveShort != nil || st.Military.ActiveLong != nil || st.Military.ActiveGrand != nil {
		t.Error("a refused command launched something")
	}
}

// TestOfferLineMarksLegendaryOffers verifies a Legendary offer is marked in
// the printed board and an ordinary one is not.
func TestOfferLineMarksLegendaryOffers(t *testing.T) {
	engine := newExpeditionTestEngine(t)
	state := engine.GetState()
	info := game.ExpeditionInfo{Number: 3, Name: "Legendary Hunt: track the herd across the Dry Creek", Rare: true, CanLaunch: true, SuccessChance: .6, DurationMin: 900}
	if line := offerLine(info, state); !strings.Contains(line, "Legendary") || !strings.Contains(line, "★") || !strings.Contains(line, "3. ") || !strings.Contains(line, "60% odds") {
		t.Errorf("a Legendary offer is not marked or lacks its numbers: %q", line)
	}
	info.Rare, info.Name = false, "Hunt: track the herd across the Dry Creek"
	if line := offerLine(info, state); strings.Contains(line, "★") {
		t.Errorf("an ordinary offer is starred: %q", line)
	}
}
