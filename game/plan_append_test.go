package game

import (
	"slices"
	"testing"
)

// PlanAddBuild and PlanAddUpgrade add to the plan the same way: onto the last
// item when it is the same building and kind, otherwise as a new item, with the
// per-item cap and the full-plan limit. Each says its own wording when the last
// item is at the cap, so the two cannot share that line.
func TestPlanAppend_BuildAndUpgrade(t *testing.T) {
	const atCap = "That plan item already holds 1000, the most one item can."
	const planFull = "The plan is full (60 items, not counting research). Remove one with plan remove <n> first."

	type twin struct {
		name  string
		kind  string
		key   string
		setup func(t *testing.T) *GameEngine
		add   func(ge *GameEngine, n int) (int, error)
	}
	logged := func(ge *GameEngine, kind, key string) int {
		total := 0
		for _, e := range ge.planLog {
			if e.Kind == kind && e.Key == key && e.Age == ge.age {
				total += e.Count
			}
		}
		return total
	}
	twins := []twin{
		{"build", PlanBuild, "stash",
			func(t *testing.T) *GameEngine { return planTestEngine(t) },
			func(ge *GameEngine, n int) (int, error) { return ge.PlanAddBuild("stash", n) }},
		{"upgrade", PlanUpgrade, "gathering_camp",
			func(t *testing.T) *GameEngine {
				ge := upgradePlanEngine(t)
				ge.Buildings.counts["gathering_camp"] = 5000
				return ge
			},
			func(ge *GameEngine, n int) (int, error) { return ge.PlanAddUpgrade("gathering_camp", n) }},
	}
	for _, k := range twins {
		t.Run(k.name, func(t *testing.T) {
			ge := k.setup(t)
			if got, err := k.add(ge, 400); err != nil || got != 400 {
				t.Fatalf("first add = %d, %v; want 400", got, err)
			}
			// Onto the last item, as far as the cap lets it.
			if got, err := k.add(ge, 700); err != nil || got != 600 {
				t.Fatalf("merge past the cap = %d, %v; want 600", got, err)
			}
			if len(ge.plan) != 1 || ge.plan[0].Kind != k.kind || ge.plan[0].Key != k.key || ge.plan[0].Count != maxPlanCount {
				t.Fatalf("plan = %+v, want one item holding %d", ge.plan, maxPlanCount)
			}
			if got := logged(ge, k.kind, k.key); got != maxPlanCount {
				t.Errorf("the plan log holds %d, want %d", got, maxPlanCount)
			}
			// The last item is at the cap: nothing more goes onto it.
			if got, err := k.add(ge, 1); err == nil || err.Error() != atCap || got != 0 {
				t.Errorf("add onto a full item = %d, %v; want 0 and %q", got, err, atCap)
			}
			if len(ge.plan) != 1 || ge.plan[0].Count != maxPlanCount {
				t.Errorf("plan changed by the refused add: %+v", ge.plan)
			}

			// A different last item: the same building goes in as a new item.
			hut := PlanItem{Kind: PlanBuild, Key: "hut", Count: 1}
			ge.plan = append(ge.plan, hut)
			before := len(ge.plan)
			if got, err := k.add(ge, 2); err != nil || got != 2 {
				t.Fatalf("add after another item = %d, %v; want 2", got, err)
			}
			if last := ge.plan[len(ge.plan)-1]; len(ge.plan) != before+1 || last.Key != k.key || last.Kind != k.kind || last.Count != 2 {
				t.Errorf("plan = %+v, want a new last item of 2 %s", ge.plan, k.key)
			}

			// A full plan takes no new item, but still grows the last one.
			for planLoad(ge.plan) < MaxPlanItems {
				ge.plan = slices.Insert(ge.plan, 0, hut)
			}
			if got, err := k.add(ge, 3); err != nil || got != 3 || ge.plan[len(ge.plan)-1].Count != 5 {
				t.Errorf("merge into a full plan = %d, %v, last item %+v; want 3 and a count of 5", got, err, ge.plan[len(ge.plan)-1])
			}
			ge.plan[len(ge.plan)-1] = hut
			if got, err := k.add(ge, 1); err == nil || err.Error() != planFull || got != 0 {
				t.Errorf("new item into a full plan = %d, %v; want 0 and %q", got, err, planFull)
			}
		})
	}
}
