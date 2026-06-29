package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestDiplomacyProvider_RendersAllFactions verifies the diplomacy overlay renders
// every canonical faction (including undiscovered ones as locked teasers), shows
// status + opinion, and does not panic on a mixed-status snapshot.
func TestDiplomacyProvider_RendersAllFactions(t *testing.T) {
	// Build a snapshot covering each status branch plus an undiscovered faction.
	statuses := []string{"neutral", "friendly", "allied", "rival", "embargo"}
	factions := make(map[string]game.FactionInfo)
	defs := config.BaseFactions()
	for i, def := range defs {
		if i == len(defs)-1 {
			// Leave the last faction undiscovered to exercise the locked teaser path.
			factions[def.Key] = game.FactionInfo{Name: def.Name, Specialty: def.Specialty, Discovered: false}
			continue
		}
		factions[def.Key] = game.FactionInfo{
			Name:       def.Name,
			Specialty:  def.Specialty,
			Discovered: true,
			Opinion:    -100 + i*45, // spans negative → positive across factions
			Status:     statuses[i%len(statuses)],
			TradeBonus: def.TradeBonus,
			TradeCount: i * 3,
		}
	}

	state := game.GameState{Diplomacy: game.DiplomacyState{Factions: factions}}

	out := diplomacyProvider(state, 80)
	if out == "" {
		t.Fatal("diplomacyProvider returned empty output")
	}

	// Every faction name should appear (discovered or not).
	for _, def := range defs {
		if !strings.Contains(out, def.Name) {
			t.Errorf("overlay output missing faction %q", def.Name)
		}
	}
	// Status labels and section header should be present.
	for _, want := range []string{"Diplomacy", "Opinion", "Status", "allied", "Undiscovered"} {
		if !strings.Contains(out, want) {
			t.Errorf("overlay output missing expected token %q", want)
		}
	}
}

// TestDiplomacyProvider_EmptyState confirms the provider is safe on a zero-value
// state (nil factions map) and shows the onboarding hint.
func TestDiplomacyProvider_EmptyState(t *testing.T) {
	out := diplomacyProvider(game.GameState{}, 80)
	if out == "" {
		t.Fatal("diplomacyProvider returned empty output on zero state")
	}
	if !strings.Contains(out, "Colonial Age") {
		t.Errorf("empty-state overlay should show the onboarding hint, got:\n%s", out)
	}
}

// TestDiplomacyProvider_RendersExpandedCivData renders a snapshot exercising the
// new civilization fields — personality, backstory, war banner, and lent-worker
// status — across the full roster and confirms the panel includes them without
// panicking.
func TestDiplomacyProvider_RendersExpandedCivData(t *testing.T) {
	defs := config.BaseFactions()
	factions := make(map[string]game.FactionInfo)
	for i, def := range defs {
		info := game.FactionInfo{
			Name:        def.Name,
			Specialty:   def.Specialty,
			Personality: def.Personality,
			Backstory:   def.Backstory,
			Discovered:  true,
			Opinion:     20,
			Status:      "neutral",
		}
		switch i % 3 {
		case 0:
			info.AtWar = true
			info.Opinion = -90
			info.Status = "embargo"
		case 1:
			info.LentWorkers = 5
			info.LentPerm = true
			info.Status = "allied"
			info.Opinion = 85
		}
		factions[def.Key] = info
	}

	out := diplomacyProvider(game.GameState{Diplomacy: game.DiplomacyState{Factions: factions}}, 80)
	if out == "" {
		t.Fatal("diplomacyProvider returned empty output for expanded roster")
	}
	// Personality labels, a war banner, and lent-worker status should all surface.
	for _, want := range []string{"AT WAR", "on loan", "tribute"} {
		if !strings.Contains(out, want) {
			t.Errorf("expanded overlay missing expected token %q", want)
		}
	}
	// Each civ's personality string should render somewhere.
	for _, def := range defs {
		if !strings.Contains(out, def.Personality) {
			t.Errorf("overlay missing personality %q for civ %q", def.Personality, def.Name)
		}
	}
}

// TestDiplomacyThreshold covers the distance-to-next-tier indicator branches.
func TestDiplomacyThreshold(t *testing.T) {
	cases := []struct {
		status  string
		opinion int
		want    string
	}{
		{"neutral", 0, "to friendly"},
		{"friendly", 30, "to ally-eligible"},
		{"friendly", 55, "ally-eligible"},
		{"allied", 100, "maxed"},
		{"rival", -10, "decaying"},
		{"embargo", -50, "decaying"},
	}
	for _, c := range cases {
		got := diplomacyThreshold(c.status, c.opinion)
		if !strings.Contains(got, c.want) {
			t.Errorf("diplomacyThreshold(%q, %d) = %q, want it to contain %q",
				c.status, c.opinion, got, c.want)
		}
	}
}
