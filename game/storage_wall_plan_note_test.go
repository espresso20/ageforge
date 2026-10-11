package game

import "testing"

// The plan names a store by its display name, not its key: "dark matter",
// not "dark_matter".
func TestWall_PlanNoteNamesTheResourceNotTheKey(t *testing.T) {
	ge := newSeededEngine(1)
	const key = "protein_synthesizer" // priced in dark matter, an Interstellar Age building
	def := ge.Buildings.defs[key]
	ge.age = def.RequiredAge
	if def.RequiredTech != "" {
		ge.Research.researched[def.RequiredTech] = true
	}
	ge.plan = []PlanItem{{Kind: PlanBuild, Key: key, Count: 1}}
	cost, _ := ge.Buildings.BuildBatchCost(key, 1, nil)
	if res := ge.overStore(cost); res != "dark_matter" {
		t.Fatalf("setup: first store over is %q, want dark_matter", res)
	}
	v := ge.planViews()
	if len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "needs more dark matter storage" {
		t.Errorf("plan view = %+v, want blocked: needs more dark matter storage", v)
	}
}
