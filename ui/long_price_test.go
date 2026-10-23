package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// Prices got longer when four raw materials were priced: 24 buildings list
// five materials, and three Cyberpunk Age buildings write a cost line of more
// than 80 characters. These tests put the longest real prices on every screen
// that prints one (the Buildings panel, the build list, the Buildings and
// Build plan overlays, the wonder panel) at the two small sizes people play
// at, and fail on a price that is cut, broken inside a material, or pushed
// off the screen.

var longPriceSizes = [][2]int{{80, 24}, {100, 30}}

// longPriceCases are the longest real prices: the three Cyberpunk Age
// buildings with the longest cost lines, and a five-material Iron Age building.
var longPriceCases = []struct {
	age  string
	keys []string
}{
	{"iron_age", []string{"timber_yard"}},
	{"cyberpunk_age", []string{"neon_citadel", "cyber_hub", "dark_energy_tap"}},
}

// pricePartsOf are the materials of a price as every screen writes them
// ("6.00K iron ore"), in the order FormatCost lists them.
func pricePartsOf(cost map[string]float64) []string {
	return strings.Split(FormatCost(cost), ", ")
}

// partsNotWhole returns the parts of a price that rows do not carry whole. A
// part may continue on the next row at a space (a box narrower than
// "49.8T dark matter crystals" has to), so the rows are read as one run of
// words: a word cut in two, or a part left out, is not found.
func partsNotWhole(rows []string, parts []string) []string {
	run := flat(rows)
	var lost []string
	for _, p := range parts {
		if !strings.Contains(run, p) {
			lost = append(lost, p)
		}
	}
	return lost
}

// priceTown is a dashboard over a town of age with none of key built, every
// stock of the age's scale, so the first copy of key can be bought.
func priceTown(t *testing.T, age, key string) (*Dashboard, *tview.Pages, *game.GameEngine, game.BuildingState) {
	t.Helper()
	d, pages, eng := stagedDashboard(t, age)
	all := map[string]int{}
	for k, b := range eng.GetState().Buildings {
		all[k] = b.Count
	}
	all[key] = 0
	eng.Buildings.LoadCounts(all)
	eng.StepTicks(1)
	st := eng.GetState()
	for k, rs := range st.Resources {
		if rs.Unlocked {
			eng.SetStockForTest(k, eng.Rules().StockScale(age))
		}
	}
	eng.StepTicks(1)
	st = eng.GetState()
	b := st.Buildings[key]
	b.NextCost = buildingPrice(&st, key, b) // the price the screens show
	if !b.Unlocked || len(b.NextCost) == 0 {
		t.Fatalf("staging: %s is %+v in the %s", key, b, age)
	}
	return d, pages, eng, b
}

// TestTheLongestPriceIsOnTheList: the cases above hold the longest cost line
// of any building, so a longer one added later is not missed.
func TestTheLongestPriceIsOnTheList(t *testing.T) {
	eng := game.NewGameEngine()
	longest, longestKey := 0, ""
	for _, def := range eng.Rules().Buildings() {
		if n := len("   Cost: " + FormatCost(def.BaseCost)); n > longest {
			longest, longestKey = n, def.Key
		}
	}
	best := 0
	for _, c := range longPriceCases {
		for _, key := range c.keys {
			def, ok := eng.Rules().Building(key)
			if !ok {
				t.Fatalf("no building %q", key)
			}
			if n := len("   Cost: " + FormatCost(def.BaseCost)); n > best {
				best = n
			}
		}
	}
	if best != longest {
		t.Errorf("the longest cost line is %d characters (%s); the cases here reach %d: add it", longest, longestKey, best)
	}
	if longest < 80 {
		t.Errorf("the longest cost line is %d characters, expected 80 or more", longest)
	}
}

// TestLongPricesFitTheBuildingsPanel: in the Buildings panel on the dashboard
// every word of the price is there, none cut in two, and no row is cut.
func TestLongPricesFitTheBuildingsPanel(t *testing.T) {
	restoreForge(t)
	for _, c := range longPriceCases {
		for _, key := range c.keys {
			d, pages, _, b := priceTown(t, c.age, key)
			parts := pricePartsOf(b.NextCost)
			for _, sz := range longPriceSizes {
				w, h := sz[0], sz[1]
				var rows []string
				seen := false
				for _, page := range panelPages(t, d, pages, w, h) {
					if strings.Contains(flat(page), b.Name) {
						seen = true
						rows = append(rows, page...)
					}
				}
				if !seen {
					t.Errorf("%s at %dx%d: the Buildings panel never shows %s", c.age, w, h, b.Name)
					continue
				}
				if lost := partsNotWhole(rows, parts); len(lost) > 0 {
					t.Errorf("%s at %dx%d: the Buildings panel does not show %v whole on a row:\n%s", key, w, h, lost, strings.Join(rows, "\n"))
				}
			}
		}
	}
}

// screenWith scrolls an overlay's text view until a row shows marker, with
// the rows that follow it, and returns the screen drawn there.
func screenWith(t *testing.T, d *Dashboard, pages *tview.Pages, name, marker string, w, h int) string {
	t.Helper()
	tv := d.overlayMgr.entries[name].tv
	text := strings.Split(tv.GetText(true), "\n")
	start := 0
	for i, l := range text {
		if strings.Contains(l, marker) {
			start = i
			break
		}
	}
	// Scroll to the row, then on until the marker is near the top of the box
	// (or the text runs out, and the last page has it lower down).
	last := ""
	for row := start; row < start+400; row++ {
		tv.ScrollTo(row, 0)
		screen := renderText(t, pages, w, h)
		lines := strings.Split(screen, "\n")
		for i, l := range lines {
			if strings.Contains(l, marker) {
				last = screen
				if i <= 3 {
					return screen
				}
			}
		}
	}
	if last == "" {
		t.Fatalf("%s overlay at %dx%d: never showed %q", name, w, h, marker)
	}
	return last
}

// TestLongPricesFitTheBuildOverlay: the Buildings overlay writes the price
// under the name with each material whole on a row, at both sizes.
func TestLongPricesFitTheBuildOverlay(t *testing.T) {
	restoreForge(t)
	for _, c := range longPriceCases {
		for _, key := range c.keys {
			d, pages, eng, b := priceTown(t, c.age, key)
			parts := pricePartsOf(b.NextCost)
			for _, sz := range longPriceSizes {
				w, h := sz[0], sz[1]
				d.overlayMgr.screenW = w
				if !d.overlayMgr.Show("buildings", eng.GetState()) {
					t.Fatal("buildings overlay not registered")
				}
				screen := screenWith(t, d, pages, "buildings", "("+key+")", w, h)
				rows := strings.Split(screen, "\n")
				if lost := partsNotWhole(rows, parts); len(lost) > 0 {
					t.Errorf("%s at %dx%d: the Buildings overlay does not show %v whole on a row:\n%s", key, w, h, lost, screen)
				}
				d.overlayMgr.Hide()
			}
		}
	}
}

// TestLongPricesFitTheBuildPlan: the plan prints a ready item's whole price
// ("starts next tick for ..."), and a waiting item names the material it is
// short of with its amounts; both keep their words whole at both sizes.
func TestLongPricesFitTheBuildPlan(t *testing.T) {
	restoreForge(t)
	for _, c := range longPriceCases {
		for _, key := range c.keys {
			d, pages, eng, b := priceTown(t, c.age, key)
			if _, err := eng.PlanAddBuild(key, 1); err != nil {
				t.Fatal(err)
			}
			eng.StepTicks(0)
			v := eng.GetState().Plan[0]
			if v.Status != game.PlanStatusReady {
				t.Fatalf("staging: the plan item for %s is %s (%s)", key, v.Status, v.Note)
			}
			// The plan prints the item's own price (a wonder's is what its
			// bank lacks, which is its full price).
			parts := pricePartsOf(v.Cost)
			if len(parts) != len(b.NextCost) {
				t.Fatalf("staging: the plan prices %s in %d materials, the building in %d", key, len(parts), len(b.NextCost))
			}
			for _, sz := range longPriceSizes {
				w, h := sz[0], sz[1]
				d.overlayMgr.screenW = w
				if !d.overlayMgr.Show("plan", eng.GetState()) {
					t.Fatal("plan overlay not registered")
				}
				screen := renderText(t, pages, w, h)
				if lost := partsNotWhole(strings.Split(screen, "\n"), parts); len(lost) > 0 {
					t.Errorf("%s at %dx%d: the plan does not show %v whole on a row:\n%s", key, w, h, lost, screen)
				}
				d.overlayMgr.Hide()
			}
		}
	}
}

// TestLongPricesFitTheWonderPanel: a wonder's bank is a line a material; the
// longest of them (the Cyberpunk Age's) keeps each line whole.
func TestLongPricesFitTheWonderPanel(t *testing.T) {
	restoreForge(t)
	for _, age := range []string{"iron_age", "cyberpunk_age"} {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		wonder := st.CurrentAgeWonderKey
		def, _ := st.Ruleset().Building(wonder)
		for _, sz := range longPriceSizes {
			w, h := sz[0], sz[1]
			d.overlayMgr.screenW = w
			if !d.overlayMgr.Show("wonders", st) {
				t.Fatal("wonders overlay not registered")
			}
			screen := screenWith(t, d, pages, "wonders", "○ "+def.Name, w, h)
			rows := strings.Split(screen, "\n")
			for res, need := range def.BaseCost {
				line := game.ResourceName(res)
				found := false
				for _, r := range rows {
					if strings.Contains(r, line) && strings.Contains(r, FormatNumber(need)) {
						found = true
					}
				}
				if !found {
					t.Errorf("%s at %dx%d: the wonder bank has no whole row for %s / %s:\n%s", wonder, w, h, line, FormatNumber(need), screen)
				}
			}
			d.overlayMgr.Hide()
		}
	}
}

// tallRows is every row tv writes its text on when it is drawn lw cells wide
// and tall enough to show all of it: the box's own wrapping, read whole.
func tallRows(t *testing.T, tv *tview.TextView, lw int) []string {
	t.Helper()
	const tall = 2000
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	defer sim.Fini()
	sim.SetSize(lw+2, tall)
	tv.SetRect(0, 0, lw+2, tall)
	tv.ScrollToBeginning()
	tv.Draw(sim)
	sim.Show()
	cells, cw, ch := sim.GetContents()
	var rows []string
	for y := 1; y < ch-1; y++ {
		var b strings.Builder
		for x := 1; x < cw-1; x++ {
			if rs := cells[y*cw+x].Runes; len(rs) > 0 {
				b.WriteRune(rs[0])
			} else {
				b.WriteByte(' ')
			}
		}
		rows = append(rows, strings.TrimRight(b.String(), " "))
	}
	return rows
}

// TestLongPricesFitTheBuildList: `build` prints a line a building, price
// included, into the log. The log wraps at its box; no word of a price is
// broken at the two sizes.
func TestLongPricesFitTheBuildList(t *testing.T) {
	restoreForge(t)
	for _, c := range longPriceCases {
		for _, key := range c.keys {
			d, pages, eng, b := priceTown(t, c.age, key)
			res := HandleCommand("build", eng)
			parts := pricePartsOf(b.NextCost)
			for _, sz := range longPriceSizes {
				w, h := sz[0], sz[1]
				drawnDashboard(t, d, pages, w, h) // gives the log its size
				_, _, lw, _ := d.logTV.GetInnerRect()
				d.logTV.SetText(safeTags(res.Message))
				rows := tallRows(t, d.logTV, lw)
				at := -1
				for i, r := range rows {
					if strings.Contains(r, key) {
						at = i
					}
				}
				if at < 0 {
					t.Fatalf("%s at %dx%d: the build list does not list it", key, w, h)
				}
				if lost := partsNotWhole(rows[at:min(len(rows), at+10)], parts); len(lost) > 0 {
					t.Errorf("%s at %dx%d: the build list breaks %v across rows of the %d-cell log:\n%s", key, w, h, lost, lw, strings.Join(rows[at:min(len(rows), at+8)], "\n"))
				}
			}
		}
	}
}
