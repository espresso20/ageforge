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
	for _, want := range []string{"1. Hut ×3", "2. Stash", "research Tool Making", "research Fire Mastery"} {
		if !strings.Contains(list, want) {
			t.Errorf("plan list lacks %q:\n%s", want, list)
		}
	}
	if res := HandleCommand("plan remove 2", engine); res.Type == "error" || len(engine.GetState().Plan) != 3 {
		t.Errorf("plan remove 2: %+v", res)
	}
	if res := HandleCommand("plan clear", engine); res.Type == "error" || len(engine.GetState().Plan) != 0 {
		t.Errorf("plan clear: %+v", res)
	}
}

func TestPlanTradeAndAdvanceCommands(t *testing.T) {
	engine := game.NewGameEngine()
	if res := HandleCommand("plan advance", engine); res.Type == "error" {
		t.Fatalf("plan advance: %s", res.Message)
	}
	if res := HandleCommand("plan advance", engine); res.Type != "error" {
		t.Error("a second plan advance was accepted")
	}
	for _, c := range []string{"plan trade", "plan trade wood", "plan trade wood iron", "plan trade wood wood", "plan trade wood food -5", "plan advance now"} {
		if res := HandleCommand(c, engine); res.Type != "error" {
			t.Errorf("%q = %+v, want an error", c, res)
		}
	}
	text := plainText(planPanelText(engine.GetState(), 0, "", false, false))
	if !strings.Contains(text, "advance to the Stone Age") || !strings.Contains(text, "waiting for the requirements") {
		t.Errorf("panel:\n%s", text)
	}
}

func TestPlanAutocomplete(t *testing.T) {
	engine := game.NewGameEngine()
	comp := NewAutoCompleter(engine)
	if got := strings.Join(comp("plan "), ","); got != "plan advance,plan build,plan clear,plan deal,plan down,plan list,plan remove,plan research,plan trade,plan up" {
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
	for _, want := range []string{"1. Hut ×2", "ready", "blocked", "needs more wood storage", "3. Stash"} {
		if !strings.Contains(text, want) {
			t.Errorf("panel lacks %q:\n%s", want, text)
		}
	}
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
