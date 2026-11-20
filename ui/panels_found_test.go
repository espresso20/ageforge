package ui

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// quiet is a town where only the start set is found: the first age, both
// features shut, nothing under way.
func quiet() game.GameState {
	return game.GameState{Age: ageKeys()[0], Features: map[string]game.FeatureState{
		config.FeatureTradeRoutes: {Live: true}, config.FeatureDiplomacy: {Live: true},
	}}
}

func TestPanelsReached(t *testing.T) {
	has := func(st game.GameState, name string) bool { return slices.Contains(panelsReached(st), name) }
	base := quiet()
	for _, n := range []string{"plan", "history", "trade", "factions", "expedition", "army", "harbinger", "epoch"} {
		if has(base, n) {
			t.Errorf("%s is found in a quiet town", n)
		}
	}
	for _, n := range startPanels {
		if !has(base, n) {
			t.Errorf("%s is not found from the start", n)
		}
	}
	for name, change := range map[string]func(*game.GameState){
		"plan":    func(s *game.GameState) { s.Age = ageKeys()[1] },
		"history": func(s *game.GameState) { s.Age = ageKeys()[1] },
		"trade": func(s *game.GameState) {
			s.Features[config.FeatureTradeRoutes] = game.FeatureState{Live: true, Open: true}
		},
		"factions": func(s *game.GameState) {
			s.Features[config.FeatureDiplomacy] = game.FeatureState{Live: true, Open: true}
		},
		"expedition": func(s *game.GameState) { s.Military.Board = []game.BoardSlotInfo{{}} },
		"army":       func(s *game.GameState) { s.Military.SoldierCap = 4 },
		"harbinger":  func(s *game.GameState) { s.Harbinger = &game.HarbingerView{} },
		"epoch":      func(s *game.GameState) { s.CatastropheOutlook.Possible = true },
	} {
		st := quiet()
		change(&st)
		if !has(st, name) {
			t.Errorf("%s is not found when its condition holds", name)
		}
	}
	for _, age := range ageKeys() {
		st := quiet()
		st.Age = age
		if has(st, "wonders") != (currentWonder(st) != nil) {
			t.Errorf("wonders and the age %s disagree", age)
		}
	}
}

// accountDash is a Medieval town on a made-up account in a temp root.
func accountDash(t *testing.T) (*Dashboard, *stage, *game.Account) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	acct, err := game.CreateAccount("Found Player")
	must(t, err)
	d, st, eng := medievalTown(t, 279, 74)
	eng.SetAccount(acct)
	d.panels = panelView{}
	d.refresh()
	return d, st, acct
}

func listText(d *Dashboard) string {
	return untag(strings.Join(panelRows(formWords, 120, d.panels), "\n"))
}

func TestNewPanelsAndTheirPlates(t *testing.T) {
	d, _, _ := accountDash(t)
	st := d.lastState
	for _, n := range startPanels {
		if d.panels.fresh(n) {
			t.Errorf("start panel %s is marked new", n)
		}
	}
	want := 0
	for _, n := range panelsReached(*st) {
		if !slices.Contains(startPanels, n) {
			want++
		}
	}
	if want == 0 || d.panels.newCount() != want {
		t.Fatalf("new panels = %d, want %d", d.panels.newCount(), want)
	}
	if list := listText(d); strings.Contains(list, "army") != slices.Contains(panelsReached(*st), "army") {
		t.Errorf("the list does not follow what was found: %q", list)
	}
	text := untag(nextBoxText(d.next, 62, 40, true))
	if !strings.Contains(text, fmt.Sprintf(" %d new ", want)) {
		t.Errorf("the rule row has no count: %q", text)
	}
	// A click, a typed name and Show me clear the plate.
	for _, how := range []func(){
		func() { d.runLine("plan") },
		func() { d.inputField.SetText("history"); d.submitInput() },
		func() { d.showMe("", "wonders") },
	} {
		before := d.panels.newCount()
		how()
		if d.panels.newCount() != before-1 {
			t.Errorf("opening did not clear a plate: %d to %d", before, d.panels.newCount())
		}
	}
}

// What an account found stays found on its next run.
func TestFoundPanelsKeptPerAccount(t *testing.T) {
	d, _, acct := accountDash(t)
	d.runLine("plan")
	found, opened := acct.PanelsSeen()
	if !slices.Contains(found, "plan") || !slices.Equal(opened, []string{"plan"}) {
		t.Fatalf("saved found %v opened %v", found, opened)
	}
	again, ok, err := game.LoadAccount()
	must(t, err)
	if !ok {
		t.Fatal("the account did not load")
	}
	f2, o2 := again.PanelsSeen()
	if !slices.Equal(found, f2) || !slices.Equal(opened, o2) {
		t.Errorf("a second run has found %v opened %v", f2, o2)
	}
	// A fresh dashboard (a later run) starts from them and shows no plate for plan.
	d2, _, eng2 := stagedDashboard(t, "primitive_age")
	eng2.SetAccount(again)
	d2.refresh()
	if !d2.panels.found["plan"] || d2.panels.fresh("plan") {
		t.Error("a later run lost plan")
	}
}

// A plate takes the cells beside a name, so the letters keep their columns.
func TestPlateKeepsColumns(t *testing.T) {
	plated := panelView{found: seenPanels().found, opened: map[string]bool{}}
	plain := seenPanels()
	for _, c := range []struct{ form, T int }{{formRich, 62}, {formGrouped, 100}, {formWords, 200}} {
		a := strings.Split(untag(strings.Join(panelRows(c.form, c.T, plated), "\n")), "\n")
		b := strings.Split(untag(strings.Join(panelRows(c.form, c.T, plain), "\n")), "\n")
		if len(a) != len(b) {
			t.Errorf("form %d: %d rows with plates, %d without", c.form, len(a), len(b))
			continue
		}
		for i := range a {
			for _, n := range panelNames() {
				if x, y := strings.Index(a[i], n), strings.Index(b[i], n); x != y {
					t.Errorf("form %d: %s moved from column %d to %d", c.form, n, y, x)
				}
			}
		}
	}
}

// seenPanels is a view with every panel found and opened.
func seenPanels() panelView {
	found, opened := map[string]bool{}, map[string]bool{}
	for _, n := range panelNames() {
		found[n], opened[n] = true, true
	}
	return panelView{found, opened}
}

func ageKeys() []string {
	var st game.GameState
	return st.Ruleset().AgeKeys()
}
