package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// newExpeditionTestEngine spins up a fresh engine in the Bronze Age, stocked
// with enough food/wood/soldiers to launch the early scouting/military
// expeditions used by these tests. scout_party (scouting) and raid_bandits
// (military) both unlock by the Bronze Age. The age jump uses the dev console
// (/age applies the age's unlocks); resource amounts are set via LoadAmounts so
// they bypass nominal storage caps.
func newExpeditionTestEngine(t *testing.T) *game.GameEngine {
	t.Helper()

	prevDev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prevDev })

	engine := game.NewGameEngine()
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

// TestCmdExpeditionNoArgsOpensPanel verifies bare `expedition` opens the
// Expeditions panel rather than launching anything.
func TestCmdExpeditionNoArgsOpensPanel(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition(nil, engine)
	if res.OverlayName != "expedition" {
		t.Errorf("cmdExpedition(nil) OverlayName = %q, want %q", res.OverlayName, "expedition")
	}
}

// TestCmdExpeditionMilitaryKeyRedirects verifies a MILITARY key passed to
// `expedition` returns the redirect-to-campaign info message and does NOT
// launch (no active scout, no active military).
func TestCmdExpeditionMilitaryKeyRedirects(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition([]string{"raid", "bandits"}, engine)
	if res.Type != "info" {
		t.Errorf("cmdExpedition(raid_bandits) Type = %q, want %q", res.Type, "info")
	}
	if !strings.Contains(res.Message, "campaign raid_bandits") {
		t.Errorf("cmdExpedition(raid_bandits) message = %q, want it to redirect to 'campaign raid_bandits'", res.Message)
	}
	st := engine.GetState()
	if st.Military.ActiveScout != nil || st.Military.ActiveMilitary != nil {
		t.Errorf("redirect should not launch anything; got scout=%v military=%v", st.Military.ActiveScout, st.Military.ActiveMilitary)
	}
}

// TestCmdExpeditionScoutingKeyLaunches verifies a SCOUTING key launches and
// sets the active scout.
func TestCmdExpeditionScoutingKeyLaunches(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdExpedition([]string{"scout", "party"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdExpedition(scout_party) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveScout == nil {
		t.Errorf("after launching scout_party, ActiveScout is nil; expected a running scouting expedition")
	}
}

// TestCmdCampaignScoutingKeyRedirects verifies a SCOUTING key passed to
// `campaign` returns the redirect-to-expedition info message and does NOT
// launch.
func TestCmdCampaignScoutingKeyRedirects(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdCampaign([]string{"scout", "party"}, engine)
	if res.Type != "info" {
		t.Errorf("cmdCampaign(scout_party) Type = %q, want %q", res.Type, "info")
	}
	if !strings.Contains(res.Message, "expedition scout_party") {
		t.Errorf("cmdCampaign(scout_party) message = %q, want it to redirect to 'expedition scout_party'", res.Message)
	}
	st := engine.GetState()
	if st.Military.ActiveScout != nil || st.Military.ActiveMilitary != nil {
		t.Errorf("redirect should not launch anything; got scout=%v military=%v", st.Military.ActiveScout, st.Military.ActiveMilitary)
	}
}

// TestCmdCampaignMilitaryKeyLaunches verifies a MILITARY key launches and sets
// the active military campaign.
func TestCmdCampaignMilitaryKeyLaunches(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdCampaign([]string{"raid", "bandits"}, engine)
	if res.Type != "success" {
		t.Fatalf("cmdCampaign(raid_bandits) Type = %q (msg %q), want %q", res.Type, res.Message, "success")
	}
	if st := engine.GetState(); st.Military.ActiveMilitary == nil {
		t.Errorf("after launching raid_bandits, ActiveMilitary is nil; expected a running campaign")
	}
}

// TestCmdCampaignListReturnsCampaigns verifies bare `campaign` returns the
// campaign list (info), naming a military campaign and not opening a panel.
func TestCmdCampaignListReturnsCampaigns(t *testing.T) {
	engine := newExpeditionTestEngine(t)

	res := cmdCampaign(nil, engine)
	if res.Type != "info" {
		t.Errorf("cmdCampaign(nil) Type = %q, want %q", res.Type, "info")
	}
	if res.OverlayName != "" {
		t.Errorf("cmdCampaign(nil) OverlayName = %q, want empty (list, not panel)", res.OverlayName)
	}
	if !strings.Contains(res.Message, "Campaigns") {
		t.Errorf("cmdCampaign(nil) message = %q, want a campaign list", res.Message)
	}
}
