package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

func returnLineCase(xp float64) string {
	return returnLine(game.MissionReturn{Label: "Scout", Place: "the Copper Hills", Success: true, Experience: xp})
}

// The Ready Room header names the army's rank and its experience toward the
// next one, at the smallest size the panel draws; the Army tab names the rank,
// the experience and what the rank gives; and a return line says how much
// experience a mission gave or took.
func TestArmyRankShown(t *testing.T) {
	engine, _ := expFixture(t)
	state := engine.GetState()
	state.Military.Experience, state.Military.Rank = 42, 2
	for _, size := range [][2]int{{80, 24}, {100, 30}, {160, 45}} {
		text := renderExpedition(state, expView{}, size[0], size[1]).String()
		if !strings.Contains(text, "Army: Veteran · 42/80") {
			t.Errorf("the Ready Room at %dx%d lacks the rank:\n%s", size[0], size[1], text)
		}
		if !strings.Contains(text, "Soldiers ") {
			t.Errorf("the Ready Room at %dx%d lost the soldiers line:\n%s", size[0], size[1], text)
		}
	}
	state.Military.Experience, state.Military.Rank = 250, 4
	if text := renderExpedition(state, expView{}, 80, 24).String(); !strings.Contains(text, "Army: Legendary · 250") || strings.Contains(text, "250/") {
		t.Errorf("the top rank has nothing to count toward:\n%s", text)
	}

	st := peopleTown(t)
	st.Military.Experience, st.Military.Rank = 12.5, 1
	for _, size := range [][2]int{{80, 24}, {100, 30}} {
		army := renderPeople(st, peopleView{tab: tabArmy}, size[0], size[1]).String()
		wantAll(t, fmt.Sprintf("the Army tab at %dx%d", size[0], size[1]), army, "Rank Seasoned", "12.5/30", "odds +2 points, hauls +5%")
	}
}

func TestReturnLineSaysTheExperience(t *testing.T) {
	r := returnLineCase(3)
	if !strings.Contains(r, "The army gained 3 experience.") {
		t.Errorf("a gain is not told: %q", r)
	}
	if r := returnLineCase(-1.5); !strings.Contains(r, "The army lost 1.5 experience.") {
		t.Errorf("a loss is not told: %q", r)
	}
	if r := returnLineCase(0); strings.Contains(r, "experience") {
		t.Errorf("no change is told: %q", r)
	}
}

// The Stone Era's board shows short offers only, and a Veteran army sees the
// long ones too, in the state the panels read.
func TestBoardShowsTheOpenTiers(t *testing.T) {
	engine := newExpeditionTestEngine(t)
	engine.OpenTiersForTest(1)
	state := engine.GetState()
	if len(state.Military.Board) != 3 {
		t.Fatalf("the Bronze Age board has %d slots, want 3", len(state.Military.Board))
	}
	for _, slot := range state.Military.Board {
		if slot.Tier != game.TierShort {
			t.Errorf("slot %d is a %s offer in the Stone Era", slot.Number, slot.Tier)
		}
	}
	text := renderExpedition(state, expView{}, 80, 24).String()
	if !strings.Contains(text, "Army: Green · 0/10") {
		t.Errorf("the header lacks the Green army:\n%s", text)
	}
}
