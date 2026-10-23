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

// A building whose next copy costs more than is held is marked like any
// building that is not affordable (the red cross), and nothing about
// storage: stock has no limit.
func TestBuildingsPanelHasNoStorageWall(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{"workshop": 80})
	st := eng.GetState()
	over := ""
	for res, cost := range st.Buildings["workshop"].NextCost {
		if cost > st.Resources[res].Amount {
			over = res
		}
	}
	if over == "" {
		t.Fatalf("staging: the town holds the next workshop (%v)", st.Buildings["workshop"].NextCost)
	}
	err := eng.BuildBuilding("workshop")
	if err == nil || strings.Contains(err.Error(), "storage") {
		t.Fatalf("build workshop = %v, want a plain refusal", err)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		for _, rows := range panelPages(t, d, pages, w, h) {
			text := flat(rows)
			for _, bad := range []string{"more than your", "storage holds", "Build more storage", "at its limit"} {
				if strings.Contains(text, bad) {
					t.Errorf("%dx%d: the Buildings panel says %q:\n%s", w, h, bad, strings.Join(rows, "\n"))
				}
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

// An item in the plan whose price is over a store does not wait on storage:
// the plan's note never speaks of one.
func TestPlanPanelNeverWaitsOnStorage(t *testing.T) {
	restoreForge(t)
	d, pages, eng := wallTown(t, map[string]int{"workshop": 80})
	if _, err := eng.PlanAddBuild("workshop", 1); err != nil {
		t.Fatal(err)
	}
	v := eng.GetState().Plan[0]
	if strings.Contains(v.Note, "storage") {
		t.Fatalf("the plan item is %+v, want a note that does not speak of storage", v)
	}
	list := untag(planListText(eng.GetState()))
	if strings.Contains(list, "storage") {
		t.Errorf("plan list speaks of storage:\n%s", list)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		d.overlayMgr.screenW = w
		if !d.overlayMgr.Show("plan", eng.GetState()) {
			t.Fatal("plan overlay not registered")
		}
		screen := renderText(t, pages, w, h)
		if strings.Contains(screen, "storage") || strings.Contains(screen, "banked toward") {
			t.Errorf("%dx%d: the plan panel speaks of storage or a bank:\n%s", w, h, screen)
		}
		d.overlayMgr.Hide()
	}
}

// The Ledger's waiting line lists the age's storage building like any other
// building count, and the refusal to advance points at it.
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
	if r := HandleCommand("advance", eng); r.Type != "error" || !strings.Contains(r.Message, "waiting on") {
		t.Errorf("advance with two stores short: %+v", r)
	}
	for _, sz := range wallSizes {
		w, h := sz[0], sz[1]
		cells := drawnDashboard(t, d, pages, w, h)
		rows := checkFitted(t, fmt.Sprintf("Ledger at %dx%d", w, h), d.ledgerTV, cells, w)
		got := flat(rows)
		if !strings.Contains(got, name+" 3/5") {
			t.Errorf("%dx%d: the Ledger does not list the %s still to build:\n%s", w, h, name, strings.Join(rows, "\n"))
		}
	}
}
