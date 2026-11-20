package ui

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// planShotGame is a made-up account in the Medieval Age with a mixed plan.
func planShotGame(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	for _, a := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		must(t, ge.EnterAgeForTest(a))
	}
	ge.GrantTechsForTest()
	st := ge.GetState()
	counts, n := map[string]int{}, 0
	for _, k := range sortedKeysOf(st.Buildings) {
		if b := st.Buildings[k]; b.Unlocked && b.Category != "wonder" {
			counts[k] = 2 + n%4
			n++
		}
	}
	ge.Buildings.LoadCounts(counts)
	ge.StepTicks(2)
	good, n := ge.Rules().StockScale(ge.GetState().Age), 0
	for _, k := range sortedKeysOf(ge.GetState().Resources) {
		if ge.GetState().Resources[k].Unlocked {
			ge.SetStockForTest(k, good*[]float64{.02, .3, 1.5, .1, .6}[n%5])
			n++
		}
	}
	ge.StepTicks(1)
	keys := plannableBuildingKeys(ge.GetState())
	for i, k := range keys {
		if i < 7 {
			HandleCommand("plan build "+k+" 3", ge)
		}
	}
	for i, k := range plannableTechKeys(ge.GetState()) {
		if i < 2 {
			HandleCommand("plan research "+k, ge)
		}
	}
	HandleCommand("plan advance", ge)
	for i, k := range keys {
		if i >= 7 && i < 10 {
			HandleCommand("plan build "+k+" 2", ge)
		}
	}
	return ge
}

// TestPlanPictures writes the panel's pictures into the directory PLAN_SHOT_DIR names.
func TestPlanPictures(t *testing.T) {
	dir := os.Getenv("PLAN_SHOT_DIR")
	if dir == "" {
		t.Skip("PLAN_SHOT_DIR not set")
	}
	pictureCell = func() (int, int) { return 7, 15 }
	defer func() { pictureCell = nil }()
	save := func(name string, p *planPanel, w, h int) {
		f, err := os.Create(filepath.Join(dir, "2026-11-20-plan-"+name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, ScreenImagePixel(paintAlone(overBar(p), w, h+6))); err != nil {
			t.Fatal(err)
		}
	}
	open := func(ge *game.GameEngine, w, h int) *planPanel {
		p := newPlanPanel()
		p.engine = ge
		p.open(ge.GetState())
		p.Layout(image.Rect(0, 0, w, h))
		return p
	}
	ge := planShotGame(t)
	p := open(ge, 279, 68)
	p.selectTo(2)
	save("main", p, 279, 68)
	p.drag.on, p.drag.down, p.drag.from, p.drag.to, p.drag.y = true, true, 2, 5, 18
	p.relayout()
	save("drag", p, 279, 68)
	p.drag.on, p.drag.down = false, false
	oe := optimizeCommandEngine(t) // a small town whose plan has a faster order
	for _, c := range []string{"plan build hut 10", "plan build wood_camp 10", "plan build gathering_camp 10"} {
		HandleCommand(c, oe)
	}
	p = open(oe, 279, 68)
	p.optimize()
	if p.mode != planPreview {
		t.Fatal("no preview to picture")
	}
	save("optimize", p, 279, 68)
	ge.PlanSetPaused(true)
	p = open(ge, 279, 68)
	p.selectTo(1)
	save("paused", p, 279, 68)
	ge.PlanSetPaused(false)
	q := open(ge, 157, 38)
	q.selectTo(1)
	save("compact", q, 157, 38)
}
