package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// plainText strips tview color tags, for asserting on what the player reads.
func plainText(s string) string {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText(s)
	return tv.GetText(true)
}

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
	text := plainText(planPanelText(engine.GetState(), 0, "", false, false, 100))
	if !strings.Contains(text, "advance to the Stone Age") || !strings.Contains(text, "0 of 4 pillars") {
		t.Errorf("panel:\n%s", text)
	}
}

func TestPlanAutocomplete(t *testing.T) {
	engine := game.NewGameEngine()
	comp := NewAutoCompleter(engine)
	if got := strings.Join(comp("plan "), ","); got != "plan advance,plan build,plan clear,plan deal,plan down,plan list,plan optimize,plan remove,plan research,plan up,plan upgrade" {
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

func TestPlanPanel_KeysAndText(t *testing.T) {
	engine := game.NewGameEngine()
	var p planPanel
	if text := plainText(p.provider(engine.GetState(), 100)); !strings.Contains(text, "The plan is empty") {
		t.Errorf("empty panel:\n%s", text)
	}
	for _, c := range []string{"plan build hut 2", "plan build story_circle", "plan build stash"} {
		HandleCommand(c, engine)
	}
	text := plainText(p.provider(engine.GetState(), 100))
	wantAll(t, "panel", text, "1. Hut ×2", "ready", "blocked", "too little wood coming in", "3. Stash")
	key := func(k tcell.Key, r rune) {
		if !p.handleKey(tcell.NewEventKey(k, r, tcell.ModNone), engine) {
			t.Fatalf("key %v %q not consumed", k, r)
		}
	}
	key(tcell.KeyDown, 0)
	key(tcell.KeyRune, 'u') // story_circle to the top
	if st := engine.GetState(); st.Plan[0].Key != "story_circle" || p.sel != 0 {
		t.Errorf("after down, u: plan[0] = %s, sel %d", st.Plan[0].Key, p.sel)
	}
	key(tcell.KeyRune, 'x')
	if st := engine.GetState(); len(st.Plan) != 2 || st.Plan[0].Key != "hut" {
		t.Errorf("after x: %+v", st.Plan)
	}
	key(tcell.KeyRune, 'C')
	if len(engine.GetState().Plan) != 2 || !p.clearArmed {
		t.Error("one C cleared the plan or didn't arm")
	}
	key(tcell.KeyRune, 'C')
	if len(engine.GetState().Plan) != 0 {
		t.Error("two Cs didn't clear the plan")
	}
	if p.handleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone), engine) {
		t.Error("an unbound key was consumed")
	}
}

// The detail line of each status: what a waiting item is short of and how
// much of that is held back for items above it, the reason a blocked item
// waits, and what a ready one will pay. Nothing speaks of a bank.
func TestPlanPanel_DetailLines(t *testing.T) {
	st := game.GameState{Resources: map[string]game.ResourceState{"stone": {Amount: 900}, "wood": {Amount: 50}}}
	waiting := game.PlanItemView{Kind: game.PlanBuild, Key: "longhouse", Name: "Longhouse", Count: 1,
		Cost:   map[string]float64{"stone": 1000, "wood": 100},
		Status: game.PlanStatusWaiting, Progress: 0.6, Short: "stone"}
	if got := plainText(planItemDetail(waiting, st)); !strings.Contains(got, "stone 600 / 1K (300 held for items above)") {
		t.Errorf("waiting line = %q", got)
	}
	blocked := waiting
	blocked.Status, blocked.Note, blocked.Short = game.PlanStatusBlocked, "needs more stone storage", ""
	if got := plainText(planItemDetail(blocked, st)); got != "needs more stone storage" {
		t.Errorf("blocked line = %q", got)
	}
	ready := waiting
	ready.Status, ready.Short, ready.Progress = game.PlanStatusReady, "", 1
	if got := plainText(planItemDetail(ready, st)); got != "starts next tick for 1K stone, 100 wood" {
		t.Errorf("ready line = %q", got)
	}
	ready.Cost = nil
	if got := plainText(planItemDetail(ready, st)); got != "starts next tick" {
		t.Errorf("free ready line = %q", got)
	}
	st.Plan = []game.PlanItemView{blocked}
	if got := plainText(planListText(st)); !strings.HasSuffix(got, "blocked: needs more stone storage") {
		t.Errorf("plan list = %q", got)
	}
}

func upgradeCommandEngine(t *testing.T, count int) *game.GameEngine {
	t.Helper()
	engine := game.NewGameEngine()
	engine.SetTownForTest(map[string]int{"gathering_camp": count}, 0)
	must(t, engine.EnterAgeForTest("stone_age"))
	return engine
}

func TestPlanUpgradeCommandsAndPanel(t *testing.T) {
	engine := upgradeCommandEngine(t, 5)
	res := HandleCommand("plan upgrade gathering_camp 3", engine)
	if res.Type == "error" || res.Message != "Planned to upgrade 3 Gathering Camps to Forager Post. Each starts as soon as it can be paid." {
		t.Fatalf("plan upgrade count: %+v", res)
	}
	line := plainText(planPanelText(engine.GetState(), 0, "", false, false, 100))
	if !strings.Contains(line, "upgrade 3 Gathering Camp → Forager Post") {
		t.Fatalf("plan panel line missing: %s", line)
	}

	all := upgradeCommandEngine(t, 4)
	res = HandleCommand("plan upgrade gathering_camp all", all)
	if res.Type == "error" || res.Message != "Planned to upgrade 4 Gathering Camps to Forager Post (all you hold)." {
		t.Fatalf("plan upgrade all: %+v", res)
	}
}

// planItemDetail is the second part of an item's line: the progress bar and
// what it is short of, the reason it is blocked, or what it will pay.
func planItemDetail(v game.PlanItemView, state game.GameState) string {
	return strings.ReplaceAll(planItemDetailGlued(v, state), glue, " ")
}
