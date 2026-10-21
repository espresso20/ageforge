package game

import (
	"strings"
	"testing"
)

// An advance moves nothing and offers nothing: it turns the buildings of the
// age left into older buildings, which the summary lists and the log names.
func TestAdvanceAge_TurnsBuildingsOlder(t *testing.T) {
	ge := NewGameEngine()
	ge.Buildings.UnlockBuilding("gathering_camp")
	ge.Buildings.counts["gathering_camp"] = 2

	if ge.Buildings.IsOlder("gathering_camp") {
		t.Fatal("a Gathering Camp is older in the Primitive Age")
	}
	ge.advanceAge("stone_age")

	target, ok := ge.Buildings.UpgradeTarget("gathering_camp")
	if !ok || target.Key != "forager_post" {
		t.Fatalf("gathering_camp upgrades to %q (%v) in the Stone Age, want forager_post", target.Key, ok)
	}
	if got := ge.Buildings.GetCount("gathering_camp"); got != 2 {
		t.Errorf("the advance moved copies: %d Gathering Camps, want 2", got)
	}
	if got := ge.Buildings.GetCount("forager_post"); got != 0 {
		t.Errorf("the advance made %d Forager Posts, want none", got)
	}
	var listed bool
	for _, tr := range ge.lastAgeAdvanceSummary.BuildingsTransformed {
		if tr.OldKey == "gathering_camp" && tr.NewKey == "forager_post" && tr.Count == 2 {
			listed = true
		}
	}
	if !listed {
		t.Errorf("the advance summary does not list the Gathering Camps: %+v", ge.lastAgeAdvanceSummary)
	}
	if len(ge.lastAgeAdvanceSummary.BuildingsLegacy) != 0 {
		t.Errorf("the advance flagged legacy buildings %v; nothing is flagged any more", ge.lastAgeAdvanceSummary.BuildingsLegacy)
	}
	if len(ge.GetAvailableUpgrades()) == 0 {
		t.Error("upgrade lists nothing after the advance")
	}
	var named bool
	for _, l := range ge.GetState().Log {
		if strings.Contains(l.Message, "Gathering Camps can upgrade to Forager Post") {
			named = true
		}
	}
	if !named {
		t.Error("the log does not name the Gathering Camps that can upgrade")
	}
}

// A building turns older again each time its line gets a newer rung, and one
// that is last in its line never does. A paused line keeps its last rung as
// current until it resumes.
func TestAdvanceAge_LinesNeverSwitchResource(t *testing.T) {
	ge := NewGameEngine()
	// Each step stays in the epoch the engine is in: an epoch crossing rolls a
	// random event that could cost a copy.
	advance := func(age string) {
		ge.currentEpoch = ge.rules.EraOf(age)
		ge.advanceAge(age)
	}
	to := func(key string) string {
		target, ok := ge.Buildings.UpgradeTarget(key)
		if !ok {
			return ""
		}
		return target.Key
	}
	for _, k := range []string{"wood_workshop", "sawmill", "crypto_exchange", "black_market", "nano_alloy_plant"} {
		ge.Buildings.UnlockBuilding(k)
		ge.Buildings.counts[k] = 1
	}
	advance("medieval_age")
	if got := to("wood_workshop"); got != "sawmill" {
		t.Errorf("wood_workshop upgrades to %q in the Medieval Age, want sawmill", got)
	}
	advance("renaissance_age")
	advance("colonial_age")
	if got := to("sawmill"); got != "" {
		t.Errorf("sawmill is the last of the wood line but upgrades to %s", got)
	}
	if ge.Buildings.IsOlder("sawmill") {
		t.Error("sawmill is the last of its line and must not be older")
	}
	advance("cyberpunk_age")
	if got := to("crypto_exchange"); got != "" {
		t.Errorf("crypto_exchange upgrades to %s in the Cyberpunk Age, which makes crypto, not gold", got)
	}
	advance("fusion_age")
	if got := to("crypto_exchange"); got != "energy_exchange" {
		t.Errorf("crypto_exchange upgrades to %q in the Fusion Age, want energy_exchange", got)
	}
	if got := to("black_market"); got != "" {
		t.Errorf("black_market is the last of the crypto line but upgrades to %s", got)
	}
	if got := to("nano_alloy_plant"); got != "" {
		t.Errorf("the titanium line is paused in the Fusion Age, so nano_alloy_plant is current, but it upgrades to %s", got)
	}
	advance("space_age")
	if got := to("nano_alloy_plant"); got != "orbital_refinery" {
		t.Errorf("nano_alloy_plant upgrades to %q in the Space Age, want orbital_refinery", got)
	}
}
