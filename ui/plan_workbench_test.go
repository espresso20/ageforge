package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
)

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

func wbKey(t *testing.T, p *planPanel, engine *game.GameEngine, k tcell.Key, r rune) {
	t.Helper()
	if !p.handleKey(tcell.NewEventKey(k, r, tcell.ModNone), engine) {
		t.Fatalf("key %v %q not consumed", k, r)
	}
}

func wbType(t *testing.T, p *planPanel, engine *game.GameEngine, s string) {
	t.Helper()
	for _, r := range s {
		wbKey(t, p, engine, tcell.KeyRune, r)
	}
}

func TestWorkbench_WideAndNarrowLayouts(t *testing.T) {
	engine := workbenchEngine(t)
	p := planPanel{engine: engine}
	wide := plainText(p.provider(engine.GetState(), 124))
	wantAll(t, "wide", wide, "Build plan", "Plan has no finish time yet", "Resource", "Ready in", "│")
	narrow := plainText(p.provider(engine.GetState(), 90))
	wantAll(t, "narrow", narrow, "Build plan", "Resource")
	if strings.Contains(narrow, "│") {
		t.Errorf("narrow layout is side by side:\n%s", narrow)
	}
	for _, l := range strings.Split(wide, "\n") {
		if visibleLen(l) > 120 {
			t.Errorf("line wider than the panel (%d): %q", visibleLen(l), l)
		}
	}
	if n := strings.Count(wide, "U/D"); n != 0 {
		t.Errorf("the key legend is in the body (%d times); it belongs in the footer", n)
	}
}

func TestWorkbench_QueueScrollsWithTheSelection(t *testing.T) {
	engine := game.NewGameEngine()
	for _, k := range []string{"hut", "gathering_camp", "stash", "story_circle", "wood_camp"} {
		for i := 0; i < 8; i++ {
			// Alternate so items do not merge.
			HandleCommand("plan build "+k, engine)
			HandleCommand("plan research tool_making", engine)
			HandleCommand("plan build "+[]string{"hut", "stash"}[i%2], engine)
		}
	}
	p := planPanel{engine: engine, height: func() int { return 24 }}
	for range 30 {
		wbKey(t, &p, engine, tcell.KeyDown, 0)
	}
	text := plainText(p.provider(engine.GetState(), 124))
	if !strings.Contains(text, "▸") || !strings.Contains(text, fmt.Sprintf("%2d.", p.sel+1)) {
		t.Errorf("selection %d is not on screen:\n%s", p.sel+1, text)
	}
	before := p.sel
	wbKey(t, &p, engine, tcell.KeyPgUp, 0)
	if p.sel >= before {
		t.Errorf("PgUp left the selection at %d (was %d)", p.sel, before)
	}
}

func TestWorkbench_OptimizePreviewKeepsOrNot(t *testing.T) {
	engine := workbenchEngine(t)
	p := planPanel{engine: engine}
	prev, err := engine.PlanOptimizePreview()
	if err != nil {
		t.Fatal(err)
	}
	order := engine.GetState().Plan
	wbKey(t, &p, engine, tcell.KeyRune, 'o')
	text := plainText(p.provider(engine.GetState(), 124))
	if prev.Changed {
		wantAll(t, "popup", text, "Optimize preview", "Now", "After", "Finishes in", "saves", "keep it")
	} else {
		wantAll(t, "popup", text, "Nothing")
	}
	wbKey(t, &p, engine, tcell.KeyEscape, 0)
	if p.mode != planModeQueue || fmt.Sprint(engine.GetState().Plan[0]) != fmt.Sprint(order[0]) {
		t.Error("Esc applied or kept the popup")
	}
	if !prev.Changed {
		return
	}
	wbKey(t, &p, engine, tcell.KeyRune, 'o')
	wbKey(t, &p, engine, tcell.KeyEnter, 0)
	got := engine.GetState().Plan
	changed := false
	for i := range got {
		changed = changed || got[i].Key != order[i].Key || got[i].Count != order[i].Count
	}
	if !changed || !strings.Contains(p.note, "Kept") {
		t.Errorf("Enter did not keep the new order (note %q)", p.note)
	}
}

func TestWorkbench_AddFromThePanel(t *testing.T) {
	engine := game.NewGameEngine()
	HandleCommand("plan build hut 2", engine)
	HandleCommand("plan build stash", engine)
	p := planPanel{engine: engine}
	wbKey(t, &p, engine, tcell.KeyRune, 'a')
	wbType(t, &p, engine, "tool mak")
	wantAll(t, "box", plainText(p.provider(engine.GetState(), 124)), "Add: tool mak", "Tool Making")
	wbKey(t, &p, engine, tcell.KeyEnter, 0)
	plan := engine.GetState().Plan
	if len(plan) != 3 || plan[1].Key != "tool_making" || p.sel != 1 {
		t.Fatalf("research went to the wrong place (sel %d): %+v", p.sel, plan)
	}
	// Esc cancels.
	wbKey(t, &p, engine, tcell.KeyRune, 'a')
	wbType(t, &p, engine, "wood")
	wbKey(t, &p, engine, tcell.KeyEscape, 0)
	if p.mode != planModeQueue || len(engine.GetState().Plan) != 3 {
		t.Error("Esc did not cancel the add")
	}
	// A name that fits nothing says so.
	wbKey(t, &p, engine, tcell.KeyRune, 'a')
	wbType(t, &p, engine, "zzzz")
	wbKey(t, &p, engine, tcell.KeyEnter, 0)
	if !strings.Contains(p.note, "Nothing called") {
		t.Errorf("note = %q", p.note)
	}
}

func TestWorkbench_AddAsksWhichKindWhenAmbiguous(t *testing.T) {
	engine := game.NewGameEngine()
	st := engine.GetState()
	var input string
	var picks []planPick
	for _, w := range []string{"stone", "fire", "tool", "wood", "farm", "hunt", "fish", "storage", "camp", "hut", "pottery", "fur"} {
		got := planPicks(st, engine, w)
		kinds := map[string]bool{}
		for _, pk := range got {
			kinds[pk.kind] = true
		}
		if len(got) > 1 {
			input, picks = w, got
			break
		}
	}
	if input == "" {
		t.Skip("no name in the fixture fits two things")
	}
	p := planPanel{engine: engine}
	wbKey(t, &p, engine, tcell.KeyRune, 'a')
	wbType(t, &p, engine, input)
	wbKey(t, &p, engine, tcell.KeyEnter, 0)
	if p.mode != planModePick {
		t.Fatalf("mode = %d, want the choice", p.mode)
	}
	wantAll(t, "choice", plainText(p.provider(engine.GetState(), 124)), "Which one?")
	wbKey(t, &p, engine, tcell.KeyDown, 0)
	wbKey(t, &p, engine, tcell.KeyEnter, 0)
	if plan := engine.GetState().Plan; len(plan) == 0 || plan[0].Key != picks[1].key {
		t.Errorf("plan = %+v, want the second choice %+v", plan, picks[1])
	}
}

func TestWorkbench_PlusMinusChangeTheCount(t *testing.T) {
	engine := game.NewGameEngine()
	HandleCommand("plan build hut 3", engine)
	HandleCommand("plan research tool_making", engine)
	p := planPanel{engine: engine}
	wbKey(t, &p, engine, tcell.KeyRune, '+')
	wbKey(t, &p, engine, tcell.KeyRune, '+')
	wbKey(t, &p, engine, tcell.KeyRune, '-')
	if got := engine.GetState().Plan[0].Count; got != 4 {
		t.Errorf("count = %d, want 4", got)
	}
	for range 5 {
		wbKey(t, &p, engine, tcell.KeyRune, '-')
	}
	if got := engine.GetState().Plan[0].Count; got != 1 || p.noteGood {
		t.Errorf("count = %d (note %q), want it to stop at 1", got, p.note)
	}
	wbKey(t, &p, engine, tcell.KeyDown, 0)
	wbKey(t, &p, engine, tcell.KeyRune, '+')
	if p.noteGood {
		t.Errorf("+ on a research item: %q", p.note)
	}
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
