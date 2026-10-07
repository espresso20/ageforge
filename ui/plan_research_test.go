package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
)

// The plan as a research queue, as the player meets it: the command says
// what it added, the tree shows each planned tech's place, and the card's
// second Enter plans a locked tech with what it needs.

// planKeys is the plan's research items in the snapshot, in order.
func planKeys(st game.GameState) []string {
	var out []string
	for _, v := range st.Plan {
		if v.Kind == game.PlanResearch {
			out = append(out, v.Key)
		}
	}
	return out
}

// TestPlanResearchCommandAddsTheChain: `plan research` plans what the tech
// still needs before it and says so in order, takes a tech of the next age,
// and answers a tech of an age still hidden like a key it does not know.
func TestPlanResearchCommandAddsTheChain(t *testing.T) {
	engine := game.NewGameEngine()
	res := HandleCommand("plan research pottery", engine)
	if want := "Planned research: Pottery, after Fire Mastery, which it needs first. Techs start one at a time, in plan order. Pottery waits for the Stone Age."; res.Type == "error" || res.Message != want {
		t.Errorf("plan research pottery:\n got %q\nwant %q", res.Message, want)
	}
	if got := strings.Join(planKeys(engine.GetState()), ","); got != "fire_mastery,pottery" {
		t.Errorf("the plan holds %s", got)
	}
	// What is already on its way is not added twice, and a tech that needs
	// nothing more reads as before.
	res = HandleCommand("plan research animal husbandry", engine)
	if want := "Planned research: Animal Husbandry. Techs start one at a time, in plan order. Animal Husbandry waits for the Stone Age."; res.Message != want {
		t.Errorf("plan research animal husbandry:\n got %q\nwant %q", res.Message, want)
	}
	res = HandleCommand("plan research tool_making", engine)
	if want := "Planned research: Tool Making. Techs start one at a time, in plan order."; res.Message != want {
		t.Errorf("plan research tool_making:\n got %q\nwant %q", res.Message, want)
	}
	list := plainText(HandleCommand("plan list", engine).Message)
	for _, want := range []string{"research Fire Mastery", "research Pottery", "waits for the Stone Age"} {
		if !strings.Contains(list, want) {
			t.Errorf("plan list lacks %q:\n%s", want, list)
		}
	}

	// An age still hidden: no name, no suggestion from it, nothing planned.
	for _, c := range []string{"plan research philosophy", "plan research philosophi", "plan research mathematics"} {
		res := HandleCommand(c, engine)
		if res.Type != "error" || !strings.HasPrefix(res.Message, "No tech '") || strings.Contains(res.Message, "Philosophy") || strings.Contains(res.Message, "'philosophy'?") {
			t.Errorf("%s: %+v, want a refusal that names nothing hidden", c, res)
		}
	}
	if got := len(planKeys(engine.GetState())); got != 4 {
		t.Errorf("the plan holds %d techs after the refusals, want 4", got)
	}

	// The prompt offers what the plan takes: the next age's techs too, in
	// reach order, and nothing from an age still hidden.
	offers := strings.Join(NewAutoCompleter(engine)("plan research "), ",")
	if !strings.Contains(offers, "plan research stoneworking") || strings.Contains(offers, "philosophy") || strings.Contains(offers, "pottery") {
		t.Errorf("plan research offers %s; want the next age's unplanned techs and nothing hidden or planned", offers)
	}
}

// TestPlannedTechsShowTheirPlaceOnTheMap: every tech of a planned chain
// carries its place in the plan on the tree, at both zooms, and the card of
// a planned tech says which item it is.
func TestPlannedTechsShowTheirPlaceOnTheMap(t *testing.T) {
	ge := classicalGame(t) // Philosophy running, Road Building planned first
	if res := HandleCommand("plan research imperial_legions", ge); res.Type == "error" {
		t.Fatal(res.Message)
	}
	st := ge.GetState()
	if got := strings.Join(planKeys(st), ","); got != "road_building,military_tactics,siege_warfare,imperial_legions" {
		t.Fatalf("the plan holds %s", got)
	}
	m := buildTree(st, treeGeomFor(120, false), false, "imperial_legions")
	g := m.canvas(st, "imperial_legions")
	for key, place := range map[string]string{"road_building": "▸1", "military_tactics": "▸2", "siege_warfare": "▸3", "imperial_legions": "▸4"} {
		n := m.by[key]
		if n == nil {
			t.Fatalf("%s is not on the map", key)
		}
		x, y := m.left(n)+m.geom.badgeW-1, n.top+m.geom.nameH
		if got := string([]rune{g.at(x, y).r, g.at(x+1, y).r}); got != place {
			t.Errorf("%s shows %q at its top right corner, want %s", key, got, place)
		}
		if card := cardOf(st, key); !strings.Contains(card, "It is item "+place[len("▸"):]+" in your plan.") {
			t.Errorf("%s's card does not say its place:\n%s", key, card)
		}
	}
	// Zoomed out, the plan's arrow leads the emblem, and a tech that still
	// waits keeps its closing mark.
	far := buildTree(st, treeGeomFor(120, true), false, "imperial_legions")
	fg := far.canvas(st, "imperial_legions")
	for key, closing := range map[string]rune{"road_building": ' ', "military_tactics": ' ', "siege_warfare": '┆', "imperial_legions": '┆'} {
		n := far.by[key]
		if l, r := fg.at(far.left(n), n.top).r, fg.at(far.left(n)+2, n.top).r; l != '▸' || r != closing {
			t.Errorf("zoomed out, %s reads %q %q round its emblem, want the plan's arrow and %q", key, l, r, closing)
		}
	}
	if n := far.by["mathematics"]; fg.at(far.left(n), n.top).r != '✓' {
		t.Error("zoomed out, a researched tech lost its tick")
	}
}

// TestCardPlansALockedTechWithItsChain: on the card of a tech that waits for
// others, the second Enter plans it with what it needs, and the card then
// says where it stands.
func TestCardPlansALockedTechWithItsChain(t *testing.T) {
	ge := classicalGame(t)
	p := newResearchPanel()
	p.engine = ge
	p.w, p.h = 120, 37
	p.request("close", "imperial_legions")
	p.update(ge.GetState())
	before := strings.Join(renderLines(p), "\n")
	if !strings.Contains(before, "adds it to your plan, after the 2 techs it needs") {
		t.Fatalf("the card does not say what Enter will add:\n%s", before)
	}
	if !p.routeKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), "") {
		t.Fatal("Enter on the card was not taken")
	}
	if got := strings.Join(planKeys(ge.GetState()), ","); got != "road_building,military_tactics,siege_warfare,imperial_legions" {
		t.Fatalf("after Enter the plan holds %s", got)
	}
	if after := strings.Join(renderLines(p), "\n"); !strings.Contains(after, "It is item 4 in your plan.") {
		t.Errorf("the card does not say where the tech stands now:\n%s", after)
	}
}

// TestPlanningAcrossKnownAges: on known ground the ages reached before are
// in sight, so one command queues a chain across them. Each tech says which
// age it waits for, this age's tech planned below them is not held up, and
// the tree words an age further on than the next one as what it is.
func TestPlanningAcrossKnownAges(t *testing.T) {
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	ge.Stats.AgesReached = append(ge.Stats.AgesReached, "stone_age", "bronze_age", "iron_age", "classical_age")
	ge.SetStockForTest("knowledge", 500)
	res := HandleCommand("plan research philosophy", ge)
	if want := "Planned research: Philosophy, after the 3 techs it needs first (Language, Primitive Writing, then Mathematics). Techs start one at a time, in plan order. Philosophy waits for the Classical Age."; res.Message != want {
		t.Errorf("plan research philosophy:\n got %q\nwant %q", res.Message, want)
	}
	if res := HandleCommand("plan research theology", ge); res.Type != "error" || strings.Contains(res.Message, "Theology") {
		t.Errorf("a tech past the known ages: %+v", res)
	}
	if res := HandleCommand("plan research imperial legions", ge); !strings.Contains(res.Message, "after the 6 techs it needs first (Tool Making, Stoneworking, Bronze Working and 3 more)") {
		t.Errorf("a long chain: %q", res.Message)
	}
	if res := HandleCommand("plan research fire_mastery", ge); res.Type == "error" {
		t.Fatal(res.Message)
	}
	st := ge.GetState()
	list := plainText(planListText(st))
	for _, want := range []string{"research Mathematics  blocked: waits for the Iron Age", "research Philosophy  blocked: waits for the Classical Age", "research Language  ready"} {
		if !strings.Contains(list, want) {
			t.Errorf("plan list lacks %q:\n%s", want, list)
		}
	}
	// Fire Mastery, this age's, planned last: it waits only for the slot's
	// turn behind Language and Tool Making, not for the later ages above it.
	last := st.Plan[len(st.Plan)-1]
	if last.Key != "fire_mastery" || last.Note != "after the research above it" {
		t.Errorf("the last item is %s (%q), want Fire Mastery behind Tool Making", last.Key, last.Note)
	}

	m, _, g := drawTree(st, 120, 37, treeView{sel: "philosophy"})
	lines := strings.Split(g.String(), "\n")
	if !strings.Contains(lines[0], "ahead") || strings.Contains(lines[0], "next age") {
		t.Errorf("the title counts the known ages ahead as the next age: %q", lines[0])
	}
	if got := m.statusLine(st, treeView{sel: "philosophy"}); !strings.Contains(got, "waits for the Classical Age") || !strings.Contains(got, "item 4 in your plan") {
		t.Errorf("the status line for Philosophy: %q", got)
	}
	if got := m.statusLine(st, treeView{sel: "pottery"}); !strings.Contains(got, "in the next age") {
		t.Errorf("the status line for a tech of the next age: %q", got)
	}
	nexts := 0
	for _, b := range m.bands {
		if b.ahead == 1 {
			nexts++
		}
	}
	if nexts != 1 || strings.Count(g.String(), "▌IRON") == 0 {
		t.Errorf("%d bands read as the next age, want one; the Iron Age should be on the map", nexts)
	}
}

// renderLines is the panel as drawn, a string a row.
func renderLines(p *researchPanel) []string {
	g := renderTree(p.state, p.model, p.view, p.w, p.h)
	return strings.Split(g.String(), "\n")
}
