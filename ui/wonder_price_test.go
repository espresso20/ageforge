package ui

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// A wonder is paid through its bank at the ruleset's price. The engine's
// NextCost for a wonder is scaled by the build-cost factor the techs give
// every other building, so it is lower than what the bank charges (the Neon
// Citadel read 15.5T crypto in the Buildings panel and 19.2T in its bank and
// the plan). Every screen shows the price the bank charges.
func TestEveryScreenShowsTheBanksPriceForAWonder(t *testing.T) {
	restoreForge(t)
	scaledSeen := false
	for _, age := range []string{"iron_age", "cyberpunk_age"} {
		d, pages, eng := stagedDashboard(t, age)
		st := eng.GetState()
		wonders := 0
		for _, key := range sortedKeysOf(st.Buildings) {
			b := st.Buildings[key]
			if b.Category != "wonder" || b.AgeKey != age {
				continue
			}
			wonders++
			def, _ := st.Ruleset().Building(key)
			// Where the town has techs that cut build costs, the engine's
			// number differs from the bank's: that is what this guards.
			for res, c := range def.BaseCost {
				if math.Abs(b.NextCost[res]-c) > 1e-6*c {
					scaledSeen = true
				}
			}
			parts := pricePartsOf(def.BaseCost)

			// The Buildings panel and the build list.
			var rows []string
			for _, sz := range longPriceSizes {
				rows = rows[:0]
				for _, page := range panelPages(t, d, pages, sz[0], sz[1]) {
					if strings.Contains(flat(page), b.Name) {
						rows = append(rows, page...)
					}
				}
				if lost := partsNotWhole(rows, parts); len(lost) > 0 {
					t.Errorf("%s at %dx%d: the Buildings panel does not show the bank's price for %v", key, sz[0], sz[1], lost)
				}
			}
			if list := untag(HandleCommand("build", eng).Message); !strings.Contains(strings.Join(strings.Fields(list), " "), strings.Join(parts, ", ")) {
				t.Errorf("%s: the build list does not show the bank's price %q", key, strings.Join(parts, ", "))
			}

			// The Buildings overlay.
			d.overlayMgr.screenW = 100
			if !d.overlayMgr.Show("buildings", eng.GetState()) {
				t.Fatal("buildings overlay not registered")
			}
			screen := screenWith(t, d, pages, "buildings", "("+key+")", 100, 30)
			if lost := partsNotWhole(strings.Split(screen, "\n"), parts); len(lost) > 0 {
				t.Errorf("%s: the Buildings overlay does not show the bank's price for %v", key, lost)
			}
			d.overlayMgr.Hide()

			// The wonder panel: a line a part, against the same price.
			text := untag(wondersProvider(eng.GetState(), 100))
			for res, c := range def.BaseCost {
				if want := game.ResourceName(res) + ": 0 / " + textfmt.Number(c); !strings.Contains(text, want) {
					// The panel lists only the age's current wonder in full.
					if key == st.CurrentAgeWonderKey {
						t.Errorf("%s: the wonder panel has no line %q", key, want)
					}
				}
			}

			// The plan asks the bank's price.
			pst := eng.GetState()
			for _, v := range pst.Plan {
				if v.Key == key {
					t.Fatalf("staging: %s is already planned", key)
				}
			}
			if _, err := eng.PlanAddBuild(key, 1); err != nil {
				t.Fatal(err)
			}
			var item game.PlanItemView
			for _, v := range eng.GetState().Plan {
				if v.Key == key {
					item = v
				}
			}
			if len(item.Cost) != len(def.BaseCost) {
				t.Fatalf("%s: the plan prices it in %v, the bank in %v", key, item.Cost, def.BaseCost)
			}
			for res, c := range def.BaseCost {
				if math.Abs(item.Cost[res]-c) > 1e-6*c {
					t.Errorf("%s: the plan asks %v %s, the bank %v", key, item.Cost[res], res, c)
				}
			}
			if _, err := eng.PlanRemove(1); err != nil {
				t.Fatal(err)
			}
		}
		if wonders == 0 {
			t.Fatalf("no wonder in the %s", age)
		}
	}
	if !scaledSeen {
		t.Error("no wonder had an engine price different from the bank's, so this checked nothing")
	}
}
