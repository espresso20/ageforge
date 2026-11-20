package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// nextState is a Medieval snapshot where only the gate to the next age is
// open: no choice, no ready age, a tech in progress, scholars at work, nobody
// waiting and room in the homes. A test turns one thing on.
func nextState(t *testing.T) game.GameState {
	t.Helper()
	_, _, eng := stagedDashboard(t, "medieval_age")
	st := eng.GetState()
	st.PendingCatastrophe, st.AgeReady = "", false
	st.Research.CurrentTech, st.Research.NeedsScholars = "tool_making", false
	st.Workers.TotalIdle, st.Workers.HeldBy = 0, ""
	return st
}

// fullyJobbed fills every job of every building.
func fullyJobbed(st game.GameState) game.GameState {
	b := map[string]game.BuildingState{}
	for k, v := range st.Buildings {
		v.WorkersAssigned = v.Count * v.WorkerCapacity
		b[k] = v
	}
	st.Buildings = b
	return st
}

func TestNextRuleNothingPressing(t *testing.T) {
	_, _, eng := stagedDashboard(t, "medieval_age")
	st := eng.GetState()
	st.PendingCatastrophe, st.AgeReady, st.NextAgeName = "", false, ""
	st.Research.CurrentTech = "tool_making"
	st.Age = st.Ruleset().AgeKeys()[len(st.Ruleset().AgeKeys())-1]
	got := computeNext(st)
	if got.rule != ruleNone || got.cmd != "" || got.card != "" || got.panel != "" || got.title == "" {
		t.Fatalf("a quiet town gives %+v, want a title only", got)
	}
}

func TestNextRuleChoice(t *testing.T) {
	st := nextState(t)
	st.PendingCatastrophe, st.AgeReady = "some_epoch", true
	if got := computeNext(st); got.rule != ruleChoice || got.cmd != "catastrophe" {
		t.Fatalf("a pending catastrophe gives %+v, want the choice first", got)
	}
	// The offer of a memory is a modal that can wait: not a step.
	st.PendingCatastrophe, st.AgeReady, st.PendingMemoryTech = "", false, "tool_making"
	if got := computeNext(st); got.rule == ruleChoice {
		t.Fatalf("a waiting memory offer gives %+v, want no choice", got)
	}
}

func TestNextRuleAdvance(t *testing.T) {
	st := nextState(t)
	st.AgeReady = true
	if got := computeNext(st); got.rule != ruleAdvance || got.cmd != "advance" || got.panel != "charter" {
		t.Fatalf("a ready age gives %+v, want advance", got)
	}
}

func TestNextRuleResearch(t *testing.T) {
	st := nextState(t)
	st.Research.CurrentTech = ""
	st.Research.Techs = map[string]game.TechState{
		"dear":  {Name: "Dear", Available: true, Cost: 9},
		"cheap": {Name: "Cheap", Available: true, Cost: 2},
		"shut":  {Name: "Shut", Cost: 1},
	}
	got := computeNext(st)
	if got.rule != ruleResearch || got.panel != "research" || got.cmd != "research cheap" {
		t.Fatalf("an idle lab gives %+v, want a tech that can start", got)
	}
}

func TestNearestStartableWalksBackFromTheKeystone(t *testing.T) {
	techs := map[string]game.TechState{
		"key":   {Prerequisites: []string{"mid", "other"}},
		"mid":   {Prerequisites: []string{"base"}},
		"other": {Prerequisites: []string{"done"}, Available: false},
		"base":  {Available: true},
		"done":  {Researched: true},
	}
	if got := nearestStartable(techs, "key"); got != "base" {
		t.Fatalf("nearest = %q, want base", got)
	}
	techs["key"] = game.TechState{Available: true}
	if got := nearestStartable(techs, "key"); got != "key" {
		t.Fatalf("a keystone that can start: %q", got)
	}
	if got := nearestStartable(map[string]game.TechState{"key": {Prerequisites: []string{"x"}}}, "key"); got != "" {
		t.Fatalf("nothing can start, got %q", got)
	}
}

// builds checks a step is a build step of a card that passes ok.
func builds(t *testing.T, got nextStep, rule string, ok func(game.BuildingState) bool, st game.GameState) {
	t.Helper()
	if got.rule != rule || got.card == "" || got.cmd != "build "+got.card || got.panel != "" {
		t.Fatalf("got %+v, want a %s build step on the main view", got, rule)
	}
	if !ok(st.Buildings[got.card]) {
		t.Fatalf("%s is the wrong building for %s", got.card, rule)
	}
}

func TestNextRuleScholars(t *testing.T) {
	st := nextState(t)
	st.Research.CurrentTech, st.Research.NeedsScholars = "", true
	st.Research.Techs = map[string]game.TechState{}
	got := computeNext(st)
	builds(t, got, ruleScholars, func(b game.BuildingState) bool { return b.Scholars > 0 }, st)
}

func TestNextRuleIdle(t *testing.T) {
	st := fullyJobbed(nextState(t))
	st.Workers.TotalIdle = 7
	builds(t, computeNext(st), ruleIdle, func(b game.BuildingState) bool { return b.WorkerCapacity > 0 }, st)
	// A job still open: nothing to build for them.
	b := st.Buildings
	st.Buildings = map[string]game.BuildingState{}
	for k, v := range b {
		if v.WorkerCapacity > 0 && v.Count > 0 {
			v.WorkersAssigned = 0
		}
		st.Buildings[k] = v
	}
	if got := computeNext(st); got.rule == ruleIdle {
		t.Fatalf("open jobs still give %+v", got)
	}
}

func TestNextRuleHomes(t *testing.T) {
	st := nextState(t)
	st.Workers.HeldBy = game.HeldHousing
	got := computeNext(st)
	if got.rule != ruleHomes || got.card == "" || !strings.HasPrefix(got.cmd, "build ") {
		t.Fatalf("full homes give %+v", got)
	}
}

func TestNextRuleGate(t *testing.T) {
	st := nextState(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	st = eng.GetState()
	st.AgeReady, st.PendingCatastrophe = false, ""
	st.Research.CurrentTech, st.Research.NeedsScholars = "x", false
	st.Workers.TotalIdle, st.Workers.HeldBy = 0, ""
	got := computeNext(st)
	if got.rule != ruleGate || got.cmd == "" {
		t.Fatalf("an open gate gives %+v, want its first unmet row", got)
	}
	if got.card != "" {
		if got.cmd != "build "+got.card || got.panel != "" {
			t.Fatalf("a building row gives %+v", got)
		}
	} else if got.panel != "research" && got.panel != "charter" {
		t.Fatalf("a row without a building opens %q", got.panel)
	}
}

func TestNextRuleDry(t *testing.T) {
	st := nextState(t)
	var res string
	for _, k := range sortedKeysOf(st.Resources) {
		n := nextCtx{st: withRate(st, k, 1, -1), set: st.Ruleset()}
		n.cards = buildCards(n.st, buyOne)
		if got, ok := n.dry(); ok {
			res = k
			if got.rule != ruleDry || got.card == "" || got.cmd != "build "+got.card {
				t.Fatalf("got %+v", got)
			}
			break
		}
	}
	if res == "" {
		t.Fatal("no falling resource gave the dry rule")
	}
	// A slow fall, longer than the limit, is not dry.
	n := nextCtx{st: withRate(st, res, nextDrySeconds*100, -1), set: st.Ruleset()}
	n.cards = buildCards(n.st, buyOne)
	if got, ok := n.dry(); ok {
		t.Fatalf("a slow fall gave %+v", got)
	}
}

// withRate sets a resource's stock and its rate a tick.
func withRate(st game.GameState, res string, amount, rate float64) game.GameState {
	m := map[string]game.ResourceState{}
	for k, v := range st.Resources {
		m[k] = v
	}
	r := m[res]
	r.Amount, r.Rate = amount, rate
	m[res] = r
	st.Resources = m
	return st
}

// A step is followed by the next matches in rule order, then the passive
// items in fixed order, three in all.
func TestNextThenOrder(t *testing.T) {
	st := nextState(t)
	st.PendingCatastrophe, st.AgeReady = "e", true
	st.Workers.HeldBy = game.HeldHousing
	got := computeNext(st)
	if got.rule != ruleChoice || len(got.then) != nextThen {
		t.Fatalf("got %+v", got)
	}
	if got.then[0].panel != "charter" || got.then[1].card == "" {
		t.Fatalf("then = %+v, want advance first, then the homes", got.then)
	}
	// The passive items follow in a fixed order.
	st = nextState(t)
	st.Research.CurrentTech, st.Research.CurrentTechName = "tool_making", "Tool Making"
	st.Plan = []game.PlanItemView{{Name: "Hut"}}
	st.Harbinger = &game.HarbingerView{Name: "A Figure"}
	n := nextCtx{st: st}
	var panels []string
	for _, it := range n.passive() {
		panels = append(panels, it.panel)
	}
	order := []string{"research", "charter", "milestones", "plan", "festival", "harbinger"}
	last := -1
	for _, p := range panels {
		i := slices.Index(order, p)
		if i <= last {
			t.Fatalf("passive items %v are out of order", panels)
		}
		last = i
	}
	if !slices.Contains(panels, "research") || !slices.Contains(panels, "plan") || !slices.Contains(panels, "harbinger") {
		t.Fatalf("passive items %v miss one", panels)
	}
}

// Tab on an empty line of the main view writes the step's command; with text
// it completes as before; with no command, or inside a panel, it does not.
func TestTabTakesTheNextStep(t *testing.T) {
	restoreForge(t)
	d, _, _ := stagedDashboard(t, "medieval_age")
	d.refresh()
	want := d.next.cmd
	if want == "" {
		t.Skipf("the town has no step: %+v", d.next)
	}
	keyTo(d.inputField, tk(kit.KeyTab, 0, 0))
	if got := d.inputField.GetText(); got != want {
		t.Fatalf("Tab wrote %q, want %q", got, want)
	}
	d.inputField.SetText("res")
	keyTo(d.inputField, tk(kit.KeyTab, 0, 0))
	if got := d.inputField.GetText(); got == want || !strings.HasPrefix(got, "res") {
		t.Fatalf("Tab with text wrote %q, want a completion of res", got)
	}
	d.inputField.SetText("")
	d.next.cmd = ""
	keyTo(d.inputField, tk(kit.KeyTab, 0, 0))
	if got := d.inputField.GetText(); got != "" {
		t.Fatalf("Tab with no command wrote %q", got)
	}
	d.next.cmd = want
	d.runLine("charter")
	if d.takeNext() {
		t.Fatal("Tab took the step inside a panel")
	}
}

// Show me picks a building's card and brings it into view, and opens a
// panel for any other step. It builds nothing.
func TestShowMe(t *testing.T) {
	restoreForge(t)
	d, st, eng := stagedDashboard(t, "medieval_age")
	_ = st
	d.refresh()
	before := eng.GetState().Stats.TotalBuilt
	card := ""
	for _, c := range d.economyTab.buildBox.cards {
		card = c.key
		break
	}
	d.showMe(card, "")
	box := d.economyTab.buildBox
	if box.picked != card || !box.reveal {
		t.Fatalf("picked %q reveal %v, want %q in view", box.picked, box.reveal, card)
	}
	d.showMe("", "charter")
	if got := d.overlayMgr.ActiveName(); got != "charter" {
		t.Fatalf("Show me on a panel opened %q", got)
	}
	if eng.GetState().Stats.TotalBuilt != before {
		t.Fatal("Show me built something")
	}
}

// The step's card carries the plate.
func TestNextStepPlate(t *testing.T) {
	d, st, _ := medievalTown(t, 279, 74)
	d.refresh()
	key := d.next.card
	if key == "" {
		t.Skipf("the town's step is not a building: %+v", d.next)
	}
	st.paint()
	if !strings.Contains(drawPrimitive(t, st.p, 279, 74).Text(), strings.TrimSpace(txtNextStep.String())) {
		t.Fatal("the next step's card has no plate on the screen")
	}
}
