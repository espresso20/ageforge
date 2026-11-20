package ui

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
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
	st.Military.Active = []*game.ExpeditionSnapshot{st.Military.ActiveShort}
	st.Military.SoldiersAway = 20
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
		over := peopleText(st, peopleView{}, peopleBigW, peopleBigH)
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
	st.Military.Active = []*game.ExpeditionSnapshot{st.Military.ActiveShort}
	st.Military.SoldiersAway = 20
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
		p.handleKey(kit.KeyRight)
		if p.view.tab != want {
			t.Errorf("Right went to tab %d, want %d", p.view.tab, want)
		}
	}
	p.handleKey(kit.KeyLeft)
	if p.view.tab != tabScholars {
		t.Errorf("Left from the Overview went to tab %d, want the Scholars", p.view.tab)
	}
	p.view.tab, p.view.row = tabOverview, 0
	for i := 0; i < 9; i++ {
		p.handleKey(kit.KeyDown)
	}
	if p.view.row != 3 {
		t.Errorf("Down nine times on the Overview picked card %d, want the last (3)", p.view.row)
	}
	p.handleKey(kit.KeyRight)
	if p.view.row != 0 {
		t.Errorf("a new tab keeps row %d", p.view.row)
	}
	// With something typed the arrows belong to the bar.
	if p.routeKey(newKey(kit.KeyRight, 0, 0), "roster") {
		t.Error("the panel took an arrow while the prompt had text")
	}
	if !p.routeKey(newKey(kit.KeyRight, 0, 0), "") {
		t.Error("the panel left an arrow alone with an empty prompt")
	}
}

// Every tab draws at 80x24 and 100x30 with its frame whole: no row spills
// past the border, the tab row and the key bar are there, and the panel
// names the age's words.
func TestPeopleTabsFitAtEverySize(t *testing.T) {
	st := peopleTown(t)
	for _, size := range [][2]int{{150, 38}, {157, 43}, {230, 54}, {279, 71}} {
		w, h := size[0], size[1]
		for tab, tabName := range peopleTabs {
			name := tabName.String()
			for row := 0; row < 3; row++ {
				v := peopleView{tab: tab, row: row}
				g := renderPeopleGrid(st, v, w, h)
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
				if !strings.Contains(lines[1], "tabs") || strings.Contains(lines[h-2], "tabs") {
					t.Errorf("%s: the tabs hint is not once, at the end of the tab row: %q / %q", where, lines[1], lines[h-2])
				}
			}
		}
	}
}

// The Overview names the kinds, the three tiles and what needs doing.
func TestPeopleOverviewContent(t *testing.T) {
	restoreForge(t)
	_, _, eng := stagedDashboard(t, "bronze_age")
	text := peopleText(eng.GetState(), peopleView{}, peopleBigW, peopleBigH)
	wantAll(t, "the Overview", text, "Craftsfolk", "Villagers", "Spearmen", "Scribes", "At work", "Waiting", "Needs you", "Morale", "Food", "Homes", "Open Work")
}

// Below the smallest size the Overview is one centred line; at it, the
// picture and its words are whole.
func TestPeopleOverviewSizes(t *testing.T) {
	st := peopleTown(t)
	small, o := renderPeople(st, peopleView{}, nil, peopleMinW-1, peopleMinH)
	if o != nil || strings.Count(strings.TrimSpace(strings.ReplaceAll(small.String(), "│", "")), "small") != 1 {
		t.Errorf("below the smallest size the panel draws %v:\n%s", o, small.String())
	}
	for _, size := range [][2]int{{157, 43}, {157, 40}, {279, 68}, {279, 71}} {
		g, o := renderPeople(st, peopleView{}, nil, size[0], size[1])
		if o == nil {
			t.Fatalf("%dx%d: no Overview", size[0], size[1])
		}
		for _, wd := range o.words {
			if wd.x < 1 || wd.x+runeLen(wd.s) > size[0]-1 || wd.y < 3 || wd.y >= size[1]-2 {
				t.Errorf("%dx%d: %q at %d,%d is outside the picture", size[0], size[1], wd.s, wd.x, wd.y)
			}
		}
		if len(g.hits) < 4 {
			t.Errorf("%dx%d: %d card hits", size[0], size[1], len(g.hits))
		}
	}
}

// The Army tab lists the military buildings and the party out; the Scholars
// tab lists the desks and the research.
func TestPeopleArmyAndScholarsTabs(t *testing.T) {
	st := peopleTown(t)
	army := renderPeopleGrid(st, peopleView{tab: tabArmy}, 157, 43).String()
	wantAll(t, "the Army tab", army, "Military buildings", "Away", "Trade Trek to the Cedar Pass", "20 Spearmen", "Rank and defense")
	scholars := renderPeopleGrid(st, peopleView{tab: tabScholars}, 157, 43).String()
	wantAll(t, "the Scholars tab", scholars, "Desks", "Research", "Pick research", "scribes", "How to speed research")
}

// The four kinds have colours of their own in every theme.
func TestPeopleKindsHaveColours(t *testing.T) {
	p := newExpPalette()
	seen := map[kit.Style]xk{}
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
	for _, size := range [][2]int{{150, 38}, {157, 43}} {
		army := renderPeopleGrid(st, peopleView{tab: tabArmy}, size[0], size[1]).String()
		wantAll(t, fmt.Sprintf("the Army tab at %dx%d", size[0], size[1]), army, "27/27 posted", "32/32 posted")
	}
	// A building short of its posts says so in the same words.
	st.Military.Posts[0].Posted = 20
	if army := renderPeopleGrid(st, peopleView{tab: tabArmy}, 157, 43).String(); !strings.Contains(army, "20/27 posted") {
		t.Errorf("a short post is not shown:\n%s", army)
	}
	// The box reads the town's soldiers.
	st.Workers.Soldiers = 59
	st.Military.SoldierCount = 7
	if box := strings.Join(strings.Fields(plainText(peopleBoxText(st, 0, false))), " "); !strings.Contains(box, "Spearmen 59") {
		t.Errorf("the People box does not read the posts:\n%s", box)
	}
}

// TestPeoplePanelMouse: a click on a tab switches to it, a click on a card or
// a row picks it, and the wheel moves the pick, as the keys do.
func TestPeoplePanelMouse(t *testing.T) {
	st := peopleTown(t)
	p := newPeoplePanel()
	p.open(st)
	p.Layout(image.Rect(0, 0, peopleBigW, peopleBigH))
	press := func(pt image.Point) bool {
		return p.Mouse(kit.Mouse{Kind: kit.MousePress, Button: kit.ButtonLeft, X: pt.X, Y: pt.Y, Clicks: 1})
	}
	centre := func(id, n int) image.Point {
		for _, h := range p.grid.hits {
			if h.id == id && h.n == n {
				return h.r.Min.Add(image.Pt(1, 0))
			}
		}
		t.Fatalf("no hit %d/%d", id, n)
		return image.Point{}
	}
	if !press(centre(hitRow, 2)) || p.view.row != 2 {
		t.Errorf("a click on the third card picked row %d", p.view.row)
	}
	tabCell := func(n int) image.Point { return p.grid.tabs[n].r.Min.Add(image.Pt(1, 0)) }
	if !press(tabCell(tabArmy)) || p.view.tab != tabArmy || p.view.row != 0 {
		t.Errorf("a click on the Army tab: tab %d row %d", p.view.tab, p.view.row)
	}
	p.Layout(image.Rect(0, 0, peopleBigW, peopleBigH))
	if peopleRows(st, tabArmy) > 1 {
		if !press(centre(hitRow, 1)) || p.view.row != 1 {
			t.Errorf("a click on the second building picked row %d", p.view.row)
		}
	}
	if press(image.Pt(0, 0)) {
		t.Error("a click on the frame changed something")
	}
	if p.PointerAt(tabCell(tabArmy)) != kit.PointerHand || p.PointerAt(centre(hitRow, 0)) != kit.PointerHand || p.PointerAt(image.Pt(0, 0)) != kit.PointerArrow {
		t.Error("the hand should show over a tab and a row, and not over the frame")
	}
	p.view.row = 0
	if !p.Mouse(kit.Mouse{Kind: kit.MouseWheel, Wheel: 1}) || (peopleRows(st, tabArmy) > 1 && p.view.row != 1) {
		t.Errorf("the wheel left the pick on row %d", p.view.row)
	}
}

// T opens the research board from any tab; Enter closes and opens the picked
// work area's detail, on the Work tab only and never over typed text.
func TestPeopleResearchAndEnterKeys(t *testing.T) {
	engine, _ := expFixture(t)
	p := newPeoplePanel()
	p.open(engine.GetState())
	var ran []string
	p.buttons.run = func(c string) { ran = append(ran, c) }
	if !p.routeKey(newKey(kit.KeyRune, 't', 0), "") || len(ran) != 1 || ran[0] != "research" {
		t.Errorf("T ran %v, want research", ran)
	}
	if p.routeKey(newKey(kit.KeyRune, 't', 0), "x") {
		t.Error("T was taken with text in the prompt")
	}
	if p.routeKey(newKey(kit.KeyEnter, 0, 0), "") {
		t.Error("Enter was taken off the Work tab")
	}
	p.view.tab = tabWork
	if p.routeKey(newKey(kit.KeyEnter, 0, 0), "x") {
		t.Error("Enter was taken with text in the prompt")
	}
	if !p.routeKey(newKey(kit.KeyEnter, 0, 0), "") || !p.view.shut {
		t.Error("Enter did not close the detail")
	}
	p.pick(0)
	if p.view.shut {
		t.Error("picking an area left the detail closed")
	}
}
