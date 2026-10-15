package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// wallSizes are the two terminals the storage wall's screens are drawn at.
var wallSizes = [][2]int{{80, 24}, {120, 40}}

// wallStore is the Medieval Age's storage building (a Strongroom).
const wallStore = "keep"

// wallTown is a Medieval Age town (a few of every building, mixed stores)
// where counts overrides the number of some buildings: a store built 40
// times asks more for its next copy than the Medieval Age's stores can hold.
func wallTown(t *testing.T, counts map[string]int) (*Dashboard, *tview.Pages, *game.GameEngine) {
	t.Helper()
	d, pages, eng := stagedDashboard(t, "medieval_age")
	all := map[string]int{}
	for k, b := range eng.GetState().Buildings {
		all[k] = b.Count
	}
	for k, n := range counts {
		all[k] = n
	}
	eng.Buildings.LoadCounts(all)
	eng.StepTicks(1)
	return d, pages, eng
}

// flat is text on one line with its runs of spaces closed up, so a sentence
// the panel broke across rows can be looked for whole.
func flat(rows []string) string {
	return strings.Join(strings.Fields(strings.Join(rows, " ")), " ")
}

// refusalFor is what the engine says to a build of key that is over a store.
func refusalFor(t *testing.T, eng *game.GameEngine, key string) string {
	t.Helper()
	err := eng.BuildBuilding(key)
	if err == nil || !strings.Contains(err.Error(), "more than your") {
		t.Fatalf("build %s: want the over-the-store refusal, got %v", key, err)
	}
	return err.Error()
}

// panelPages draws the dashboard and turns the Buildings panel through its
// pages, returning each page's rows (checked to be whole lines), until the
// last one.
func panelPages(t *testing.T, d *Dashboard, pages *tview.Pages, w, h int) [][]string {
	t.Helper()
	var out [][]string
	d.economyTab.bldFirst = 0
	for i := 0; i < 40; i++ {
		cells := drawnDashboard(t, d, pages, w, h)
		rows := checkFitted(t, fmt.Sprintf("Buildings page %d at %dx%d", i+1, w, h), d.economyTab.buildingTV, cells, w)
		out = append(out, rows)
		before := d.economyTab.bldFirst
		d.economyTab.ScrollDown()
		if d.economyTab.bldFirst == before {
			break
		}
	}
	return out
}

// A building whose next copy costs more than a store holds is marked in the
// Buildings panel: the red cross it has for any building that cannot be
// bought, and under its cost the engine's own refusal, word for word, naming
// the storage building to build.
func TestBuildingsPanelMarksAnOverStoreCopy(t *testing.T) {
	restoreForge(t)
	for _, tc := range []struct {
		key    string
		count  int
		suffix string
	}{
		{"workshop", 80, "Build more storage first: Strongroom."},
	} {
		d, pages, eng := wallTown(t, map[string]int{tc.key: tc.count})
		st := eng.GetState()
		name := st.Buildings[tc.key].Name
		if res := overStoreRes(st, st.Buildings[tc.key].NextCost); res == "" {
			t.Fatalf("staging: the next %s (%v) fits the stores", name, st.Buildings[tc.key].NextCost)
		}
		want := refusalFor(t, eng, tc.key)
		if !strings.HasSuffix(want, tc.suffix) {
			t.Fatalf("the refusal for %s ends %q, want it to end %q", name, want, tc.suffix)
		}
		for _, sz := range wallSizes {
			w, h := sz[0], sz[1]
			found := false
			for _, rows := range panelPages(t, d, pages, w, h) {
				if !strings.Contains(flat(rows), want) {
					continue
				}
				found = true
				for _, r := range rows {
					if strings.Contains(r, " "+name+" x"+fmt.Sprint(tc.count)) && !strings.Contains(r, "✗") {
						t.Errorf("%dx%d: the %s is not marked with the cross: %q", w, h, name, r)
					}
				}
			}
			if !found {
				t.Errorf("%dx%d: no page of the Buildings panel says %q", w, h, want)
			}
		}
	}
	// A copy that fits is not marked.
	d2, pages2, _ := wallTown(t, nil)
	for _, sz := range wallSizes {
		for _, rows := range panelPages(t, d2, pages2, sz[0], sz[1]) {
			if strings.Contains(flat(rows), "more than your") {
				t.Errorf("%dx%d: a building that fits a store is marked:\n%s", sz[0], sz[1], strings.Join(rows, "\n"))
			}
		}
	}
}

func TestBuildingsPanelStashNextCopyFits(t *testing.T) {
	restoreForge(t)
	d, pages, eng := stagedDashboard(t, "primitive_age")
	all := map[string]int{}
	for k, b := range eng.GetState().Buildings {
		all[k] = b.Count
	}
	all["stash"] = 5
	eng.Buildings.LoadCounts(all)
	eng.StepTicks(1)
	st := eng.GetState()
	if got := st.Buildings["stash"].NextCost["wood"]; got != 574 {
		t.Fatalf("staging: the sixth Stash costs %v wood, want 574", got)
	}
	if store := st.Resources["wood"].Storage; store < st.Buildings["stash"].NextCost["wood"] {
		t.Fatalf("the sixth Stash costs %v wood but store holds %v", st.Buildings["stash"].NextCost["wood"], store)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		for _, rows := range panelPages(t, d, pages, w, h) {
			text := flat(rows)
			if strings.Contains(text, "Stash costs") && strings.Contains(text, "more than your") {
				t.Errorf("%dx%d: the panel marks the fitting Stash:\n%s", w, h, strings.Join(rows, "\n"))
			}
		}
	}
}

// Storage has no limit now: the Strongroom, built 40 times, shows neither MAX
// nor "Building limit reached." A building with a real limit still does.
func TestStorageBuildingShowsNoLimit(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{wallStore: 40, "grand_amphitheatre_monument": 1})
	st := eng.GetState()
	name := st.Buildings[wallStore].Name
	if b := st.Buildings[wallStore]; b.AtMaxCount || b.Count != 40 {
		t.Fatalf("staging: the %s is %+v", name, b)
	}
	if !st.Buildings["grand_amphitheatre_monument"].AtMaxCount {
		t.Fatal("staging: the monument is not at its limit")
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		maxes, limits, store := 0, 0, false
		for _, rows := range panelPages(t, d, pages, w, h) {
			for _, r := range rows {
				if strings.Contains(r, " "+name+" x40") {
					store = true
					if strings.Contains(r, "MAX") {
						t.Errorf("%dx%d: the %s shows MAX: %q", w, h, name, r)
					}
				}
				if strings.Contains(r, "MAX") {
					maxes++
				}
				if strings.Contains(r, "Building limit reached.") {
					limits++
				}
			}
		}
		if !store {
			t.Errorf("%dx%d: the %s is on no page of the panel", w, h, name)
		}
		if maxes != 1 || limits != 1 {
			t.Errorf("%dx%d: the monument at its limit shows MAX %d times and its limit line %d times, want once each", w, h, maxes, limits)
		}
	}
}

// The build list and overlay do not mark an owned storage building's next
// copy over-store: its store is floored to that queue-aware price.
func TestBuildListDoesNotMarkStorageNextCopy(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{wallStore: 40})
	st := eng.GetState()
	for res, cost := range st.Buildings[wallStore].NextCost {
		if st.Resources[res].Storage < cost {
			t.Fatalf("next %s costs %v %s but store holds %v", wallStore, cost, res, st.Resources[res].Storage)
		}
	}
	res := HandleCommand("build", eng)
	if strings.Contains(untag(res.Message), wallStore+" (40 built)") && strings.Contains(untag(res.Message), "more than your") {
		t.Errorf("`build` marks the fitting %s over-store:\n%s", wallStore, untag(res.Message))
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		d.overlayMgr.screenW = w
		if !d.overlayMgr.Show("buildings", eng.GetState()) {
			t.Fatal("buildings overlay not registered")
		}
		tv := d.overlayMgr.entries["buildings"].tv
		text := tv.GetText(true)
		if strings.Contains(text, "✗ "+wallStore) && strings.Contains(text, "more than your") {
			t.Errorf("%dx%d: the Buildings overlay marks the fitting %s:\n%s", w, h, wallStore, text)
		}
		tv.ScrollToEnd()
		screen := renderText(t, pages, w, h)
		if strings.Contains(screen, "✗ "+wallStore) && strings.Contains(screen, "more than your") {
			t.Errorf("%dx%d: the rendered Buildings overlay marks the fitting %s:\n%s", w, h, wallStore, screen)
		}
		d.overlayMgr.Hide()
	}
}

// An item in the plan whose price is over a store waits on storage, in the
// plan's own words, and does not read as about to start; the panel no longer
// says overflow is banked toward the plan.
func TestPlanPanelWaitsOnStorage(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{"workshop": 80})
	if _, err := eng.PlanAddBuild("workshop", 1); err != nil {
		t.Fatal(err)
	}
	v := eng.GetState().Plan[0]
	if v.Status != game.PlanStatusBlocked || !strings.HasPrefix(v.Note, "needs more ") || !strings.HasSuffix(v.Note, " storage") {
		t.Fatalf("staging: the plan item is %+v", v)
	}
	want := planNote(v)
	if strings.Contains(want, "_") {
		t.Errorf("the note shows a resource key: %q", want)
	}
	list := untag(planListText(eng.GetState()))
	if !strings.Contains(list, "blocked: "+want) {
		t.Errorf("plan list: want %q in\n%s", "blocked: "+want, list)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		d.overlayMgr.screenW = w
		if !d.overlayMgr.Show("plan", eng.GetState()) {
			t.Fatal("plan overlay not registered")
		}
		screen := renderText(t, pages, w, h)
		if !strings.Contains(screen, "blocked") || !strings.Contains(screen, want) {
			t.Errorf("%dx%d: the plan item does not read %q:\n%s", w, h, want, screen)
		}
		for _, bad := range []string{"starts next tick", "ready", "banked toward"} {
			if strings.Contains(screen, bad) {
				t.Errorf("%dx%d: the plan panel says %q of an item over a store:\n%s", w, h, bad, screen)
			}
		}
		d.overlayMgr.Hide()
	}
}

// The Next Age row lists the age's storage building like any other building
// count, and the refusal to advance points at it.
func TestNextAgeRowListsTheStorageRequirement(t *testing.T) {
	restoreForge(t)
	d, pages, eng := stagedDashboard(t, "medieval_age")
	st := eng.GetState()
	name := st.Buildings[wallStore].Name
	if st.NextAgeBldReqs[wallStore] != 5 {
		t.Fatalf("staging: the next age asks %v, want 5 of the %s", st.NextAgeBldReqs, name)
	}
	// Everything the age asks is there but two of the storage and the wonder.
	all := map[string]int{}
	for k, b := range st.Buildings {
		all[k] = b.Count
	}
	for k, n := range st.NextAgeBldReqs {
		all[k] = n
	}
	all[wallStore] = 3
	all[st.CurrentAgeWonderKey] = 1
	eng.Buildings.LoadCounts(all)
	eng.StepTicks(1)
	if r := HandleCommand("advance", eng); r.Type != "error" || !strings.Contains(r.Message, "Next Age bar") {
		t.Errorf("advance with two stores short: %+v", r)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		cells := drawnDashboard(t, d, pages, w, h)
		rows := checkFitted(t, fmt.Sprintf("Next Age at %dx%d", w, h), d.ageTV, cells, w)
		got := flat(rows)
		if !strings.Contains(got, "✗ "+name+" 3/5") {
			t.Errorf("%dx%d: the Next Age row does not list the %s still to build:\n%s", w, h, name, strings.Join(rows, "\n"))
		}
	}
}
