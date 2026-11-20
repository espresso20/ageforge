package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

func plainText(s string) string { return kit.Strip(s) }

func TestPlanCommands(t *testing.T) {
	engine := game.NewGameEngine()
	if res := HandleCommand("plan", engine); res.OverlayName != "plan" {
		t.Fatalf("bare plan = %+v, want the Plan panel", res)
	}
	for _, c := range []string{"plan build hut 3", "plan build stash", "plan research tool_making", "plan research fire_mastery"} {
		if res := HandleCommand(c, engine); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	st := engine.GetState()
	if len(st.Plan) != 4 || st.Plan[0].Count != 3 {
		t.Fatalf("plan = %+v", st.Plan)
	}
	for _, c := range []string{"plan build", "plan build hut 0", "plan build nope", "plan research", "plan remove", "plan remove 9", "plan up x", "plan frobnicate", "plan research tool_making"} {
		if res := HandleCommand(c, engine); res.Type != "error" {
			t.Errorf("%q = %+v, want an error", c, res)
		}
	}
	if res := HandleCommand("plan down 1", engine); res.Type == "error" || engine.GetState().Plan[1].Key != "hut" {
		t.Errorf("plan down 1: %+v", res)
	}
	if res := HandleCommand("plan up 2", engine); res.Type == "error" || engine.GetState().Plan[0].Key != "hut" {
		t.Errorf("plan up 2: %+v", res)
	}
	list := plainText(HandleCommand("plan list", engine).Message)
	wantAll(t, "plan list", list, "1. Hut ×3", "2. Stash", "research Tool Making", "research Fire Mastery")
	if res := HandleCommand("plan remove 2", engine); res.Type == "error" || len(engine.GetState().Plan) != 3 {
		t.Errorf("plan remove 2: %+v", res)
	}
	if res := HandleCommand("plan clear", engine); res.Type == "error" || len(engine.GetState().Plan) != 0 {
		t.Errorf("plan clear: %+v", res)
	}
}

func TestPlanAdvanceCommand(t *testing.T) {
	engine := game.NewGameEngine()
	if res := HandleCommand("plan advance", engine); res.Type == "error" {
		t.Fatalf("plan advance: %s", res.Message)
	}
	if res := HandleCommand("plan advance", engine); res.Type != "error" {
		t.Error("a second plan advance was accepted")
	}
	if res := HandleCommand("plan advance now", engine); res.Type != "error" {
		t.Errorf("plan advance now = %+v, want an error", res)
	}
	text := plainText(planListText(engine.GetState()))
	if !strings.Contains(text, "advance to the Stone Age") || !strings.Contains(text, "0 of 4 pillars") {
		t.Errorf("panel:\n%s", text)
	}
}

func TestPlanAutocomplete(t *testing.T) {
	engine := game.NewGameEngine()
	comp := NewAutoCompleter(engine)
	if got := strings.Join(comp("plan "), ","); got != "plan advance,plan bottom,plan build,plan clear,plan count,plan down,plan expedition,plan list,plan move,plan optimize,plan pause,plan remove,plan research,plan resume,plan top,plan up,plan upgrade" {
		t.Errorf("plan subcommands = %s", got)
	}
	builds := strings.Join(comp("plan build "), ",")
	if !strings.Contains(builds, "plan build hut") || !strings.Contains(builds, "plan build longhouse") || strings.Contains(builds, "warehouse") {
		t.Errorf("plan build offers %s; want this age's buildings and the next age's only", builds)
	}
	if got := comp("plan research tool"); len(got) != 1 || got[0] != "plan research tool_making" {
		t.Errorf("plan research tool = %v", got)
	}
	HandleCommand("plan build hut 2", engine)
	HandleCommand("plan build stash", engine)
	if got := strings.Join(comp("plan remove "), ","); got != "plan remove 1,plan remove 2" {
		t.Errorf("plan remove offers %s", got)
	}
}

func upgradeCommandEngine(t *testing.T, count int) *game.GameEngine {
	t.Helper()
	engine := game.NewGameEngine()
	engine.SetTownForTest(map[string]int{"gathering_camp": count}, 0)
	must(t, engine.EnterAgeForTest("stone_age"))
	return engine
}

func TestPlanUpgradeCommands(t *testing.T) {
	engine := upgradeCommandEngine(t, 5)
	res := HandleCommand("plan upgrade gathering_camp 3", engine)
	if res.Type == "error" || res.Message != "Planned to upgrade 3 Gathering Camps to Forager Post. Each starts as soon as it can be paid." {
		t.Fatalf("plan upgrade count: %+v", res)
	}
	all := upgradeCommandEngine(t, 4)
	res = HandleCommand("plan upgrade gathering_camp all", all)
	if res.Type == "error" || res.Message != "Planned to upgrade 4 Gathering Camps to Forager Post (all you hold)." {
		t.Fatalf("plan upgrade all: %+v", res)
	}
}

// workbenchEngine is a town with a plan that has something to optimize.
func workbenchEngine(t *testing.T) *game.GameEngine {
	t.Helper()
	engine := game.NewGameEngine()
	for _, c := range []string{"plan build hut 6", "plan build gathering_camp 8", "plan build hut 4", "plan research tool_making", "plan advance"} {
		if res := HandleCommand(c, engine); res.Type == "error" {
			t.Fatalf("%s: %s", c, res.Message)
		}
	}
	return engine
}

func TestPlanExpeditionCommand(t *testing.T) {
	engine := game.NewGameEngine()
	if res := HandleCommand("plan expedition", engine); res.Type != "error" {
		t.Errorf("bare plan expedition = %+v, want usage", res)
	}
	if res := HandleCommand("plan expedition nowhere at all", engine); res.Type != "error" {
		t.Errorf("an unknown mission = %+v, want an error", res)
	}
}

func TestPlanList_ShowsWhenEachItemStarts(t *testing.T) {
	engine := workbenchEngine(t)
	list := plainText(HandleCommand("plan list", engine).Message)
	wantAll(t, "plan list", list, "starts now")
}

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
