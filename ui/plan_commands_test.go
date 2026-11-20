package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The plan commands that match the panel's controls reach the engine.
func TestPlanBoardCommands(t *testing.T) {
	engine := game.NewGameEngine()
	for _, c := range []string{"plan build hut 3", "plan build stash", "plan research fire_mastery", "plan build wood_camp"} {
		if res := HandleCommand(c, engine); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	keys := func() []string {
		var out []string
		for _, v := range engine.GetState().Plan {
			out = append(out, v.Key)
		}
		return out
	}
	for _, c := range []struct {
		cmd  string
		want []string
	}{
		{"plan top 3", []string{"fire_mastery", "hut", "stash", "wood_camp"}},
		{"plan bottom 1", []string{"stash", "fire_mastery", "wood_camp", "hut"}},
		{"plan move 4 2", []string{"hut", "wood_camp", "stash", "fire_mastery"}},
		{"plan move 1 4", []string{"stash", "fire_mastery", "wood_camp", "hut"}},
	} {
		cmd, want := c.cmd, c.want
		if res := HandleCommand(cmd, engine); res.Type == "error" {
			t.Fatalf("%s: %s", cmd, res.Message)
		}
		got := keys()
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("after %q: %v, want %v", cmd, got, want)
			}
		}
		HandleCommand("plan clear", engine)
		for _, c := range []string{"plan build hut 3", "plan build stash", "plan research fire_mastery", "plan build wood_camp"} {
			HandleCommand(c, engine)
		}
	}
	if res := HandleCommand("plan count 1 7", engine); res.Type == "error" || engine.GetState().Plan[0].Count != 7 {
		t.Errorf("plan count 1 7: %+v", res)
	}
	if res := HandleCommand("plan pause", engine); res.Type == "error" || !engine.GetState().PlanPaused {
		t.Errorf("plan pause: %+v", res)
	}
	wantAll(t, "plan list", plainText(HandleCommand("plan list", engine).Message), "Paused")
	if res := HandleCommand("plan resume", engine); res.Type == "error" || engine.GetState().PlanPaused {
		t.Errorf("plan resume: %+v", res)
	}
	for _, c := range []string{"plan count", "plan count 1", "plan count x 2", "plan top", "plan move 1", "plan pause now", "plan top 9"} {
		if res := HandleCommand(c, engine); res.Type != "error" {
			t.Errorf("%q = %+v, want an error", c, res)
		}
	}
}
