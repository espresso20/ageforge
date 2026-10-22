package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// peopleTown is a Bronze Age town with buildings of every kind, soldiers and a
// party out, as a snapshot.
func peopleTown(t *testing.T) game.GameState {
	t.Helper()
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	st := eng.GetState()
	st.Workers.Soldiers = 108
	st.Workers.TotalPop += 108
	st.Military.SoldierCount = 108
	st.Military.ActiveShort = &game.ExpeditionSnapshot{Name: "Trade Trek to the Cedar Pass", Soldiers: 20, TicksLeft: 300}
	return st
}

// The box and the panel name each kind in the age's own word, from the
// lexicon, never a fixed one.
func TestPeopleNamesFollowTheAge(t *testing.T) {
	for _, age := range config.AgeOrder() {
		st := game.GameState{Age: age, AgeName: age}
		st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle, st.Workers.Scholars = 40, 50, 7, 3
		st.Workers.Soldiers = 9
		box := plainText(peopleBoxText(st, 0, true))
		over := renderPeople(st, peopleView{}, 100, 30).String()
		for _, key := range config.PeopleTerms() {
			name := config.Term(age, key).Many
			if !strings.Contains(box, name) {
				t.Errorf("%s: the People box does not use %q:\n%s", age, name, box)
			}
			if !strings.Contains(over, name) {
				t.Errorf("%s: the Overview does not use %q", age, name)
			}
		}
	}
	// And in the Bronze Age, the owner's example line.
	st := game.GameState{Age: "bronze_age"}
	st.Workers.TotalPop, st.Workers.MaxPop, st.Workers.TotalIdle, st.Workers.Scholars = 1061, 1113, 412, 30
	st.Workers.Soldiers = 108
	box := plainText(peopleBoxText(st, 0, false))
	for _, want := range []string{"Craftsfolk 511", "Villagers 412", "Spearmen 108", "Scribes 30"} {
		if !strings.Contains(strings.Join(strings.Fields(box), " "), want) {
			t.Errorf("the Bronze Age box lacks %q:\n%s", want, box)
		}
	}
}

// Soldiers are people now: the Resources box no longer lists them.
func TestSoldiersLeaveTheResourcesBox(t *testing.T) {
	st := peopleTown(t)
	if rs, ok := st.Resources["soldiers"]; !ok || !rs.Unlocked {
		t.Skip("the staged town has no soldiers resource unlocked")
	}
	for _, r := range resourceRows(st) {
		if strings.Contains(strings.ToLower(r.name), "soldier") {
			t.Errorf("the Resources box still lists %q", r.name)
		}
	}
	box := plainText(peopleBoxText(st, sidebarW-2, true))
	if !strings.Contains(box, "Spearmen") || !strings.Contains(box, "108") {
		t.Errorf("the People box does not list the soldiers:\n%s", box)
	}
}

// A party out shows beside the soldiers' count when there is room.
func TestPeopleBoxSoldiersOut(t *testing.T) {
	st := game.GameState{Age: "modern_age"}
	st.Workers.Soldiers = 108
	st.Military.ActiveShort = &game.ExpeditionSnapshot{Soldiers: 20}
	if line := plainText(peopleBoxText(st, 0, false)); !strings.Contains(line, "Soldiers 108 · 20 out") && !strings.Contains(strings.Join(strings.Fields(line), " "), "Soldiers 108 · 20 out") {
		t.Errorf("no soldiers out on the line:\n%s", line)
	}
	for _, line := range strings.Split(plainText(peopleBoxText(st, sidebarW-2, false)), "\n") {
		if runeLen(line) > sidebarW-2 {
			t.Errorf("row %q is wider than the box", line)
		}
	}
}

// people and workers open the panel; Left and Right go round the tabs; Up
// and Down pick a row and stop at the ends.
func TestPeoplePanelOpensAndSwitchesTabs(t *testing.T) {
	engine := newExpeditionTestEngine(t)
	for _, cmd := range []string{"people", "workers"} {
		if res := HandleCommand(cmd, engine); res.OverlayName != "people" {
			t.Errorf("%s = %+v, want the People panel", cmd, res)
		}
	}
	p := newPeoplePanel()
	p.open(engine.GetState())
	if p.view.tab != tabOverview {
		t.Fatalf("opened on tab %d", p.view.tab)
	}
	order := []int{tabWork, tabArmy, tabScholars, tabOverview}
	for _, want := range order {
		p.handleKey(tcell.KeyRight)
		if p.view.tab != want {
			t.Errorf("Right went to tab %d, want %d", p.view.tab, want)
		}
	}
	p.handleKey(tcell.KeyLeft)
	if p.view.tab != tabScholars {
		t.Errorf("Left from the Overview went to tab %d, want the Scholars", p.view.tab)
	}
	p.view.tab, p.view.row = tabOverview, 0
	for i := 0; i < 9; i++ {
		p.handleKey(tcell.KeyDown)
	}
	if p.view.row != 3 {
		t.Errorf("Down nine times on the Overview picked card %d, want the last (3)", p.view.row)
	}
	p.handleKey(tcell.KeyRight)
	if p.view.row != 0 {
		t.Errorf("a new tab keeps row %d", p.view.row)
	}
	// With something typed the arrows belong to the bar.
	if p.routeKey(tcell.NewEventKey(tcell.KeyRight, 0, 0), "roster") {
		t.Error("the panel took an arrow while the prompt had text")
	}
	if !p.routeKey(tcell.NewEventKey(tcell.KeyRight, 0, 0), "") {
		t.Error("the panel left an arrow alone with an empty prompt")
	}
}

// Every tab draws at 80x24 and 100x30 with its frame whole: no row spills
// past the border, the tab row and the key bar are there, and the panel
// names the age's words.
func TestPeopleTabsFitAt80x24And100x30(t *testing.T) {
	st := peopleTown(t)
	for _, size := range [][2]int{{80, 24}, {100, 30}, {80, 21}} {
		w, h := size[0], size[1]
		for tab, name := range peopleTabs {
			for row := 0; row < 3; row++ {
				v := peopleView{tab: tab, row: row}
				g := renderPeople(st, v, w, h)
				where := fmt.Sprintf("%s row %d at %dx%d", name, row, w, h)
				// The frame is whole.
				for y := 1; y < h-1; y++ {
					if l, r := g.at(0, y).r, g.at(w-1, y).r; l != '│' || r != '│' {
						t.Errorf("%s: row %d has %q and %q on the frame:\n%s", where, y, l, r, g.String())
						break
					}
				}
				lines := g.Lines()
				if !strings.Contains(lines[1], name) {
					t.Errorf("%s: the tab row lacks the tab: %q", where, lines[1])
				}
				if !strings.Contains(lines[h-2], "Esc") && tab == tabOverview {
					t.Errorf("%s: the key bar lacks Esc: %q", where, lines[h-2])
				}
				if !strings.Contains(lines[h-2], "tabs") {
					t.Errorf("%s: the key bar lacks the tabs hint: %q", where, lines[h-2])
				}
			}
		}
	}
}

// The Overview lists the kinds, the homes and what has happened lately.
func TestPeopleOverviewContent(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	eng.AddLog("success", "3 scribes moved in with the Scriptorium.")
	eng.AddLog("info", "Something about trade that is not about people.")
	st := eng.GetState()
	text := renderPeople(st, peopleView{}, 100, 30).String()
	for _, want := range []string{"Your people", "Homes ", "Craftsfolk", "Villagers", "Spearmen", "Scribes", "Food ", "Lately", "3 scribes moved in with the Scriptorium."} {
		if !strings.Contains(text, want) {
			t.Errorf("the Overview lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "not about people") {
		t.Errorf("Lately lists a line that is not about people:\n%s", text)
	}
}

// The Army tab lists the military buildings and the party out; the Scholars
// tab lists the desks and the research.
func TestPeopleArmyAndScholarsTabs(t *testing.T) {
	st := peopleTown(t)
	army := renderPeople(st, peopleView{tab: tabArmy}, 100, 30).String()
	for _, want := range []string{"Military buildings", "Away", "Trade Trek to the Cedar Pass", "20 Spearmen", "Defense"} {
		if !strings.Contains(army, want) {
			t.Errorf("the Army tab lacks %q:\n%s", want, army)
		}
	}
	scholars := renderPeople(st, peopleView{tab: tabScholars}, 100, 30).String()
	for _, want := range []string{"Desks", "Research", "Studying", "scribes", "How to speed research"} {
		if !strings.Contains(scholars, want) {
			t.Errorf("the Scholars tab lacks %q:\n%s", want, scholars)
		}
	}
}

// The stacked bar gives each kind its share of the homes and leaves the empty
// homes dim.
func TestPeopleStackBar(t *testing.T) {
	g := newXGrid(40, 1)
	stackBar(g, 0, 0, 40, []stackPart{{20, xWork}, {10, xRes}, {5, xLore}}, 70)
	count := map[xk]int{}
	for x := 0; x < 40; x++ {
		count[g.at(x, 0).k]++
	}
	if count[xWork] != 11 || count[xRes] != 6 || count[xLore] != 3 || count[xFaint] != 20 {
		t.Errorf("bar cells by kind: %v", count)
	}
	// Past the homes the bar fills and each kind keeps its share of the people.
	g = newXGrid(30, 1)
	stackBar(g, 0, 0, 30, []stackPart{{50, xWork}, {50, xRes}}, 60)
	if g.at(29, 0).k != xRes || g.at(0, 0).k != xWork || g.at(14, 0).k != xWork || g.at(15, 0).k != xRes {
		t.Errorf("an overfull bar: %q", g.Lines()[0])
	}
}

// The four kinds have colours of their own in every theme.
func TestPeopleKindsHaveColours(t *testing.T) {
	p := newExpPalette()
	seen := map[tcell.Style]xk{}
	for _, k := range []xk{xWork, xRes, xArmy, xLore} {
		if other, dup := seen[p.styles[k]]; dup {
			t.Errorf("kinds %d and %d share a style", k, other)
		}
		seen[p.styles[k]] = k
	}
}

// The Army tab shows each military building's soldiers as posted over posts,
// from the state, at 80x24 and 100x30; the People box's soldier line reads the
// posts (the town's soldiers), not the stock.
func TestPeopleArmyTabShowsPosts(t *testing.T) {
	st := peopleTown(t)
	st.Military.Posts = []game.SoldierPost{
		{Key: "war_camp", Name: "War Camp", Count: 3, PerCopy: 9, Posted: 27, Posts: 27},
		{Key: "barracks", Name: "Barracks", Count: 2, PerCopy: 16, Posted: 32, Posts: 32},
	}
	for _, key := range []string{"war_camp", "barracks"} {
		b := st.Buildings[key]
		b.Category = "military"
		b.Count = map[string]int{"war_camp": 3, "barracks": 2}[key]
		st.Buildings[key] = b
	}
	for _, size := range [][2]int{{80, 24}, {100, 30}} {
		army := renderPeople(st, peopleView{tab: tabArmy}, size[0], size[1]).String()
		for _, want := range []string{"27/27 posted", "32/32 posted"} {
			if !strings.Contains(army, want) {
				t.Errorf("the Army tab at %dx%d lacks %q:\n%s", size[0], size[1], want, army)
			}
		}
	}
	// A building short of its posts says so in the same words.
	st.Military.Posts[0].Posted = 20
	if army := renderPeople(st, peopleView{tab: tabArmy}, 100, 30).String(); !strings.Contains(army, "20/27 posted") {
		t.Errorf("a short post is not shown:\n%s", army)
	}
	// The box reads the town's soldiers.
	st.Workers.Soldiers = 59
	st.Military.SoldierCount = 7
	if box := strings.Join(strings.Fields(plainText(peopleBoxText(st, 0, false))), " "); !strings.Contains(box, "Spearmen 59") {
		t.Errorf("the People box does not read the posts:\n%s", box)
	}
}
