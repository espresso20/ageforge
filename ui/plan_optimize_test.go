package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// optimizeCommandEngine is a Primitive Age town with a crew of wood cutters and
// not much wood, so a plan has something to wait for.
func optimizeCommandEngine(t *testing.T) *game.GameEngine {
	t.Helper()
	engine := newCaseTestEngine(t, "")
	for i := 0; i < 3; i++ {
		if msg := game.DevExecCommand("/build wood_camp", engine); msg != "built wood_camp" {
			t.Fatalf("dev /build wood_camp: %q", msg)
		}
	}
	for _, c := range []string{"recruit 8", "assign wood_camp 6"} {
		if res := HandleCommand(c, engine); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	engine.SetStockForTest("wood", 150)
	engine.SetStockForTest("food", 150)
	return engine
}

func TestPlanOptimizeCommand(t *testing.T) {
	engine := optimizeCommandEngine(t)
	if res := HandleCommand("plan optimize", engine); res.Type != "error" || !strings.Contains(res.Message, "empty") {
		t.Errorf("plan optimize on an empty plan = %+v, want the empty-plan error", res)
	}
	for _, c := range []string{"plan build hut 10", "plan build wood_camp 10", "plan build gathering_camp 10"} {
		if res := HandleCommand(c, engine); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	res := HandleCommand("plan optimize", engine)
	if res.Type == "error" {
		t.Fatalf("plan optimize: %s", res.Message)
	}
	text := plainText(res.Message)
	wantAll(t, "reply", text, "Reordered the plan", "instead of", "  1. ")
	items := engine.GetState().Plan
	if len(items) <= 3 {
		t.Errorf("plan has %d items, want the three kinds mixed", len(items))
	}
	if res := HandleCommand("plan reorder", engine); res.Type == "error" || !strings.Contains(plainText(res.Message), "left it alone") {
		t.Errorf("plan reorder on an optimized plan = %+v, want it left alone", res)
	}
	if res := HandleCommand("plan optimize now", engine); res.Type != "error" || !strings.Contains(res.Message, "Usage: plan optimize") {
		t.Errorf("plan optimize now = %+v, want the usage", res)
	}
	HandleCommand("plan clear", engine)
	HandleCommand("plan research tool_making", engine)
	if res := HandleCommand("plan optimize", engine); res.Type == "error" || !strings.Contains(res.Message, "Nothing to reorder") {
		t.Errorf("plan optimize on research alone = %+v, want nothing to reorder", res)
	}
}

func TestPlanOptimizeText(t *testing.T) {
	order := make([]string, 15)
	for i := range order {
		order[i] = fmt.Sprintf("%d Huts", i+1)
	}
	res := game.PlanOptimizeResult{Changed: true, Order: order, Before: 3 * time.Hour, After: 2*time.Hour + 5*time.Minute, Beyond: 1500, Waiting: 4}
	text := planOptimizeText(res)
	wantAll(t, "text", text,
		"Reordered the plan",
		"about 2h 5m instead of 3h",
		"  1. 1 Huts",
		"  12. 12 Huts",
		"  and 3 more",
		"Only the first 2K copies were reordered; the other 1.5K keep their order after them.",
		"4 copies cannot start until the town makes what they cost.",
	)
	if strings.Contains(text, "13. ") {
		t.Errorf("text lists more than 12 items:\n%s", text)
	}
	kept := planOptimizeText(game.PlanOptimizeResult{Reason: game.PlanOptimizeBest, Before: 90 * time.Second, After: 90 * time.Second, Order: order})
	if !strings.Contains(kept, "left it alone") || !strings.Contains(kept, "about 1m 30s") || strings.Contains(kept, "  1. ") {
		t.Errorf("kept text = %q", kept)
	}
	if got := planOptimizeText(game.PlanOptimizeResult{Reason: game.PlanOptimizeNothing}); !strings.HasPrefix(got, "Nothing to reorder") {
		t.Errorf("nothing text = %q", got)
	}
	for _, s := range []string{text, kept} {
		if strings.ContainsAny(s, "—!") {
			t.Errorf("text has an em dash or an exclamation mark: %q", s)
		}
	}
}
