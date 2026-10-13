package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The screens and the knowledge store: knowledge has a store of its own that
// only knowledge buildings raise, and a tech priced over it cannot start.
// Each test reads the real state and the real screen.

// overStoreTech is a tech of the age that can be looked at and is priced
// over the knowledge store, with one that fits for contrast.
func overStoreTech(t *testing.T, st game.GameState) (over, fits string) {
	t.Helper()
	for _, k := range sortedKeysOf(st.Research.Techs) {
		ts := st.Research.Techs[k]
		switch {
		case !ts.Available:
		case ts.OverStore && over == "":
			over = k
		case !ts.OverStore && fits == "":
			fits = k
		}
	}
	if over == "" || fits == "" {
		t.Fatalf("staging: want an available tech over the store and one under it, got over=%q fits=%q", over, fits)
	}
	return over, fits
}

// knowledgeBuildingName is the name of the knowledge building the state says
// to build.
func knowledgeBuildingName(t *testing.T, st game.GameState) string {
	t.Helper()
	def, ok := st.Ruleset().Building(st.KnowledgeStore.NextBuilding)
	if !ok {
		t.Fatalf("the age has no knowledge building to name: %q", st.KnowledgeStore.NextBuilding)
	}
	return def.Name
}

// The over-the-store line names the building that raises the store: for
// knowledge the age's knowledge building, for every other resource the
// age's storage building.
func TestStoreWallNamesTheBuildingThatRaisesTheStore(t *testing.T) {
	ge := gameIn(t, "bronze_age")
	st := ge.GetState()
	know := knowledgeBuildingName(t, st)
	if got := storageBuildingName(st, "knowledge"); got != know {
		t.Errorf("knowledge: storageBuildingName = %q, want the knowledge building %q", got, know)
	}
	store := ""
	for _, d := range st.Ruleset().Buildings() {
		if d.Category == "storage" && d.RequiredAge == st.Age {
			store = d.Name
		}
	}
	if got := storageBuildingName(st, "food"); got != store || store == "" || store == know {
		t.Errorf("food: storageBuildingName = %q, want the storage building %q", got, store)
	}
	// A building whose next copy costs more knowledge than the store holds.
	sw := &storeWall{state: st}
	line := sw.line(game.BuildingState{Name: "Thing", Unlocked: true, NextCost: map[string]float64{"knowledge": st.KnowledgeStore.Store * 10}})
	if !strings.HasSuffix(line, "Build more storage first: "+know+".") || !strings.Contains(line, "more than your knowledge storage holds ("+FormatNumber(st.KnowledgeStore.Store)+")") {
		t.Errorf("the line for a knowledge price = %q", line)
	}
	// Another price in the same list still names the storage building.
	line = sw.line(game.BuildingState{Name: "Thing", Unlocked: true, NextCost: map[string]float64{"food": st.Resources["food"].Storage * 10}})
	if !strings.HasSuffix(line, "Build more storage first: "+store+".") {
		t.Errorf("the line for a food price = %q", line)
	}
}

// A tech priced over the knowledge store is not startable anywhere it
// appears: the command refuses it, the map and its card say so in the
// Buildings panel's mark and sentence, the list marks it, completion does
// not offer it and the plan holds it as blocked.
func TestOverStoreTechIsNotStartableAnywhere(t *testing.T) {
	restoreForge(t)
	ge := gameIn(t, "stone_age")
	ge.SetStockForTest("knowledge", 100) // some, but not what the tech costs
	st := ge.GetState()
	over, fits := overStoreTech(t, st)
	ts := st.Research.Techs[over]
	know := knowledgeBuildingName(t, st)
	want := fmt.Sprintf("%s costs %s knowledge, more than your knowledge storage holds (%s). Build more storage first: %s.",
		ts.Name, FormatNumber(ts.Cost), FormatNumber(st.KnowledgeStore.Store), know)
	if got := techStoreLine(st, ts); got != want {
		t.Fatalf("techStoreLine = %q, want %q", got, want)
	}

	// The command.
	err := ge.StartResearch(over)
	if err == nil || err.Error() != want {
		t.Errorf("research %s: error %v, want the engine's refusal %q", over, err, want)
	}

	// The research list: the cross, and the sentence under the tech.
	list := cmdResearchList(ge).Message
	for _, line := range strings.Split(list, "\n") {
		has := strings.Contains(line, "["+"cyan]"+over+"[-]")
		if has && !strings.Contains(line, overMark) {
			t.Errorf("`research list` does not mark %s: %q", over, line)
		}
		if strings.Contains(line, "["+"cyan]"+fits+"[-]") && strings.Contains(line, overMark) {
			t.Errorf("`research list` marks %s, which fits the store: %q", fits, line)
		}
	}
	if !strings.Contains(list, want) {
		t.Errorf("`research list` lacks the sentence %q:\n%s", want, list)
	}

	// Completion and affordability.
	if techAffordable(st, ts) {
		t.Errorf("%s is priced over the store and is offered as affordable", over)
	}
	keys := availableTechKeys(st)
	if containsWord(keys, over) || !containsWord(keys, fits) {
		t.Errorf("`research` offers %v: want %s left out and %s in", keys, over, fits)
	}

	// The map and the card, at both sizes.
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		m, _, g := drawTree(st, sz[0], sz[1], treeView{sel: over})
		n := m.by[over]
		if n.mark == markReady {
			t.Errorf("%dx%d: %s is drawn as a tech you can start", sz[0], sz[1], over)
		}
		status := m.statusLine(st, treeView{sel: over})
		if !strings.Contains(status, overMark+" ") || strings.Contains(status, "you can start it") || !strings.Contains(status, "waits for what it needs") {
			t.Errorf("%dx%d: the status line of %s = %q", sz[0], sz[1], over, status)
		}
		if !strings.Contains(g.String(), overMark) {
			t.Errorf("%dx%d: the map does not show the cross for %s:\n%s", sz[0], sz[1], over, g.String())
		}
		_, _, g2 := drawTree(st, sz[0], sz[1], treeView{sel: fits})
		if strings.Contains(strings.SplitAfter(g2.String(), "\n")[len(strings.SplitAfter(g2.String(), "\n"))-1], overMark) {
			t.Errorf("%dx%d: a tech that fits is crossed", sz[0], sz[1])
		}
	}
	m := buildTree(st, treeGeomFor(120, false), false, over)
	if cardActionFor(st, m.by[over]) == actStart {
		t.Errorf("the card of %s offers to start it", over)
	}
	card := cardOf(st, over)
	if !strings.Contains(card, want) {
		t.Errorf("the card of %s lacks %q:\n%s", over, want, card)
	}
	if strings.Contains(card, "It starts when the knowledge is there") {
		t.Errorf("the card of %s says it starts when the knowledge is there:\n%s", over, card)
	}
	if fitsCard := cardOf(st, fits); strings.Contains(fitsCard, "more than your") {
		t.Errorf("the card of %s, which fits, says it is over the store:\n%s", fits, fitsCard)
	}

	// The plan holds it, blocked, in the words the plan already has.
	if err := ge.PlanAddResearch(over); err != nil {
		t.Fatalf("plan research %s: %v", over, err)
	}
	var p planPanel
	if text := plainText(p.provider(ge.GetState(), 100)); !strings.Contains(text, "blocked") || !strings.Contains(text, "needs more knowledge storage") {
		t.Errorf("the plan does not hold %s as blocked:\n%s", over, text)
	}
}

// Stock an old save holds above the store can still pay: no cross then.
func TestHeldKnowledgePaysForATechOverTheStore(t *testing.T) {
	ge := gameIn(t, "stone_age")
	over, _ := overStoreTech(t, ge.GetState())
	cost := ge.GetState().Research.Techs[over].Cost
	ge.SetStockForTest("knowledge", cost*2)
	st := ge.GetState()
	if st.Research.Techs[over].OverStore {
		t.Skip("the engine does not keep stock above the store here")
	}
	if strings.Contains(cmdResearchList(ge).Message, overMark) {
		t.Errorf("a tech the held knowledge can pay is crossed")
	}
}

// The tree's header says the knowledge held over the store, in the form the
// Resources box uses.
func TestResearchHeaderShowsTheKnowledgeStore(t *testing.T) {
	ge := gameIn(t, "stone_age")
	ge.SetStockForTest("knowledge", 100)
	st := ge.GetState()
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		_, _, g := drawTree(st, sz[0], sz[1], treeView{sel: ""})
		top := strings.SplitN(g.String(), "\n", 2)[0]
		held, store := FormatNumber(st.Resources["knowledge"].Amount), FormatNumber(st.KnowledgeStore.Store)
		want := held + "/" + store
		if sz[0] >= 100 {
			want = held + " / " + store
		}
		if !strings.Contains(top, want) {
			t.Errorf("%dx%d: the header %q lacks %q", sz[0], sz[1], strings.TrimSpace(top), want)
		}
	}
}

// A knowledge building shows what a copy adds to the knowledge store, in the
// form a storage building gives its figure, on the Buildings panel, in the
// buildings panel overlay and in the build list; no other building does.
func TestKnowledgeBuildingsShowTheirStoreFigure(t *testing.T) {
	ge := gameIn(t, "bronze_age")
	st := ge.GetState()
	know := st.KnowledgeStore.NextBuilding
	each := st.Buildings[know].KnowledgeStore
	if each <= 0 {
		t.Fatalf("%s carries no store figure in the state", know)
	}
	figure := "+" + config.FormatAmount(each) + " knowledge storage."
	for name, text := range map[string]string{
		"Buildings panel":   buildingLines(st, 80),
		"buildings overlay": buildingsProvider(st, 80),
		"build list":        cmdBuild(nil, ge).Message,
	} {
		plain := plainText(text)
		if !strings.Contains(flat([]string{plain}), figure) {
			t.Errorf("%s lacks %q for %s", name, figure, know)
		}
		if n := strings.Count(flat([]string{plain}), "knowledge storage."); n > countKnowledgeBuildings(st) {
			t.Errorf("%s shows the knowledge store figure %d times, for %d knowledge buildings unlocked", name, n, countKnowledgeBuildings(st))
		}
	}
}

// countKnowledgeBuildings is the knowledge buildings the lists show: the
// unlocked ones carrying a store figure.
func countKnowledgeBuildings(st game.GameState) int {
	n := 0
	for _, b := range st.Buildings {
		if b.Unlocked && b.KnowledgeStore > 0 && !b.IsLegacy {
			n++
		}
	}
	return n
}

var everyResourceClaim = regexp.MustCompile(`storage for every resource( but knowledge)?`)

// Nothing on a screen says a bonus to every store raises knowledge: a
// storage building, a tech's or a wonder's bonus and the Research bonus line
// all leave it out.
func TestNoScreenSaysAStoreRaisesKnowledge(t *testing.T) {
	ge := gameIn(t, "medieval_age")
	st := ge.GetState()
	screens := map[string]string{
		"Buildings panel":   buildingLines(st, 100),
		"buildings overlay": buildingsProvider(st, 100),
		"build list":        cmdBuild(nil, ge).Message,
		"wonders overlay":   wondersAllText(st),
	}
	checked := 0
	for name, text := range screens {
		for _, m := range everyResourceClaim.FindAllString(plainText(text), -1) {
			checked++
			if m != everyStore {
				t.Errorf("%s says %q: knowledge is not raised by it", name, m)
			}
		}
	}
	if checked == 0 {
		t.Errorf("staging: no screen carries a bonus to every store")
	}
	if got := formatTechEffect(config.Effect{Type: "storage", Target: "all", Value: 500}); got != "+500 "+everyStore {
		t.Errorf("formatTechEffect(storage, all) = %q", got)
	}
	if got := formatTechEffect(config.Effect{Type: "storage", Target: "knowledge", Value: 500}); got != "+500 knowledge storage" {
		t.Errorf("formatTechEffect(storage, knowledge) = %q: a bonus that names knowledge still says so", got)
	}
}

// wondersAllText is the wonders overlay's text for a state, at a width that
// shows its descriptions.
func wondersAllText(st game.GameState) string {
	var sb strings.Builder
	for _, d := range st.Ruleset().Buildings() {
		if d.Category == "wonder" {
			sb.WriteString(storageClaim(d.Description) + "\n")
			for _, e := range d.Effects {
				sb.WriteString(formatEffect(e) + "\n")
			}
		}
	}
	return sb.String()
}

// The catastrophe modal says knowledge buildings are spared.
func TestCatastropheModalSparesKnowledgeBuildings(t *testing.T) {
	l := buildCatastropheModalLayout("steel_era", false, 0, game.DefaultEndureOutcome(), game.GameState{})
	if text := flat(strings.Split(plainText(l.endure), "\n")); !strings.Contains(text, "(wonders, storage and knowledge buildings are spared)") {
		t.Errorf("the Endure lines do not say what is spared:\n%s", l.endure)
	}
}

// The Resources box sets knowledge apart from the goods a town stores: its
// row leads the box on every page, with its amount over its own store, and
// no store of goods comes before it.
func TestResourcesBoxSetsKnowledgeApart(t *testing.T) {
	restoreForge(t)
	for _, age := range layoutAges {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		if !st.Resources["knowledge"].Unlocked {
			t.Fatalf("%s: knowledge is not unlocked", age)
		}
		if st.Resources["knowledge"].Storage != st.KnowledgeStore.Store {
			t.Fatalf("%s: the knowledge resource's cap %v is not the knowledge store %v", age, st.Resources["knowledge"].Storage, st.KnowledgeStore.Store)
		}
		for _, sz := range [][2]int{{80, 24}, {100, 30}} {
			w, h := sz[0], sz[1]
			where := fmt.Sprintf("%s at %dx%d", age, w, h)
			pagesN := 1
			for p := 0; p < pagesN; p++ {
				cells := drawnDashboard(t, d, pages, w, h)
				pagesN = d.economyTab.resPages
				x, y, bw, bh := d.economyTab.resourceTV.GetInnerRect()
				rows := rectText(cells, w, x, y, bw, bh)
				if len(rows) == 0 || !strings.HasPrefix(strings.TrimSpace(rows[0]), "Knowl") {
					t.Errorf("%s, page %d: the box does not lead with knowledge: %q", where, p+1, rows)
					continue
				}
				// Its amount over its own store: on its row, or, in a box too
				// narrow for a cap column, under the amount.
				store := "/ " + FormatNumber(st.KnowledgeStore.Store)
				if bw < 40 {
					store = "/" + FormatNumber(st.KnowledgeStore.Store)
				}
				if got := rows[0] + " " + rows[1]; !strings.Contains(got, store) {
					t.Errorf("%s, page %d: the knowledge row lacks its store %q: %q", where, p+1, store, rows[:2])
				}
				for _, r := range rows[1:] {
					if strings.HasPrefix(strings.TrimSpace(r), "Knowl") {
						t.Errorf("%s, page %d: knowledge shows twice: %q", where, p+1, rows)
					}
				}
				if pagesN > 1 {
					d.economyTab.NextResourcePage()
				}
			}
		}
	}
}

// layoutResourceBox on its own: knowledge leads every page, a rule under it
// when the box has room, and the page count is of the goods under it.
func TestResourceBoxKeepsAnApartRowOnEveryPage(t *testing.T) {
	var rows []resRow
	know := storeRow(game.ResourceState{Name: "Knowledge", Amount: 90, Storage: 240, Rate: 1})
	know.apart = true
	rows = append(rows, know)
	for i := 0; i < 20; i++ {
		rows = append(rows, storeRow(game.ResourceState{Name: fmt.Sprintf("Good %02d", i), Amount: 10, Storage: 100, Rate: 1}))
	}
	// Roomy: all on one page, the rule under knowledge.
	lines, pages := layoutResourceBox(rows, 60, 30, 0)
	if pages != 1 || !strings.Contains(untag(lines[0]), "Knowledge") || !strings.Contains(untag(lines[1]), "───") {
		t.Errorf("roomy box: pages=%d lines=%q", pages, lines[:3])
	}
	// Tight: a page of goods under knowledge, the rule only when it fits.
	for _, h := range []int{5, 6, 8, 12} {
		seen := 0
		lines, pages = layoutResourceBox(rows, 60, h, 0)
		if pages < 2 {
			t.Fatalf("h=%d: want several pages, got %d", h, pages)
		}
		for p := 0; p < pages; p++ {
			lines, _ = layoutResourceBox(rows, 60, h, p)
			if len(lines) != h {
				t.Errorf("h=%d page %d: %d lines", h, p, len(lines))
			}
			if !strings.Contains(untag(lines[0]), "Knowledge") {
				t.Errorf("h=%d page %d does not lead with knowledge: %q", h, p, lines)
			}
			for _, l := range lines[1:] {
				if strings.Contains(untag(l), "Good ") {
					seen++
				}
			}
		}
		if seen != 20 {
			t.Errorf("h=%d: the pages show %d goods, want each of 20 once", h, seen)
		}
		if last := untag(lines[len(lines)-1]); !strings.Contains(last, "of 20") {
			t.Errorf("h=%d: the page line counts %q, want the 20 goods", h, last)
		}
	}
}
