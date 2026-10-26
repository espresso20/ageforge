package mapgen

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

var testCatalog = mapmodel.NewCatalog()

func groundPlan(seed int64) (*Fields, *Plan) {
	f := newFields(seed)
	return f, NewGroundPlan(seed, f, testCatalog)
}

func TestPlanDeterministic(t *testing.T) {
	_, a := groundPlan(7)
	_, b := groundPlan(7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed, different plan")
	}
	if _, c := groundPlan(8); reflect.DeepEqual(a.Roads, c.Roads) {
		t.Fatal("two seeds, same roads")
	}
}

func TestPlanHasRoomForATown(t *testing.T) {
	for _, seed := range []int64{7, 23} {
		_, p := groundPlan(seed)
		if len(p.Lots) < 150 {
			t.Errorf("seed %d: only %d lots", seed, len(p.Lots))
		}
		placed := 0
		for _, plot := range p.Plots {
			if plot.Rect.W > 0 {
				placed++
			}
		}
		if placed < 10 {
			t.Errorf("seed %d: only %d wonder plots", seed, placed)
		}
	}
}

// TestEveryWonderHasAPlot: the wonder of every ground age has a plot, whatever
// the land.
func TestEveryWonderHasAPlot(t *testing.T) {
	for _, seed := range []int64{1, 2, 3, 4, 5, 7, 9, 11, 16, 19, 23, 27, 31, 36} {
		_, p := groundPlan(seed)
		for age := 0; age < groundAges; age++ {
			if p.Plots[age].Rect.W == 0 {
				t.Errorf("seed %d: no plot for the wonder of age %d", seed, age)
			}
		}
	}
}

// TestPlanNoOverlap: no two parcels or plots share a cell, none touches a
// road or water, and every one is dry buildable ground.
func TestPlanNoOverlap(t *testing.T) {
	for _, seed := range []int64{3, 7, 23} {
		f, p := groundPlan(seed)
		b := &planner{f: f}
		taken := map[Pt]string{}
		for _, r := range p.Roads {
			for _, c := range r.Cells {
				taken[c] = "road"
			}
		}
		claim := func(what string, r Rect) {
			r.cells(func(x, y int) {
				if prev, ok := taken[Pt{x, y}]; ok {
					t.Fatalf("seed %d: %s at (%d,%d) overlaps %s", seed, what, x, y, prev)
				}
				if !b.buildable(x, y) {
					t.Fatalf("seed %d: %s at (%d,%d) is on %s", seed, what, x, y, biomeNames[f.Biome[at(x, y)]])
				}
				taken[Pt{x, y}] = what
			})
		}
		for _, l := range p.Lots {
			claim("lot", l.Rect)
		}
		for _, l := range p.Plots {
			if l.Rect.W > 0 {
				claim("plot", l.Rect)
			}
		}
	}
}

// TestPlanEveryParcelReachable: walking the road cells from the plaza reaches
// a road cell beside every parcel's door.
func TestPlanEveryParcelReachable(t *testing.T) {
	for _, seed := range []int64{3, 7, 23} {
		_, p := groundPlan(seed)
		road := map[Pt]bool{}
		for _, r := range p.Roads {
			for _, c := range r.Cells {
				road[c] = true
			}
		}
		reached := map[Pt]bool{p.Centre: true}
		queue := []Pt{p.Centre}
		for head := 0; head < len(queue); head++ {
			for _, d := range neighbours {
				n := Pt{queue[head].X + d[0], queue[head].Y + d[1]}
				if road[n] && !reached[n] {
					reached[n] = true
					queue = append(queue, n)
				}
			}
		}
		parcels := append(append([]Lot{}, p.Lots...), p.Plots...)
		for _, l := range parcels {
			if l.Rect.W == 0 {
				continue
			}
			if !reached[l.Door] {
				t.Fatalf("seed %d: the door %v of a parcel is not reached by road", seed, l.Door)
			}
			if !beside(l.Rect, l.Door) {
				t.Fatalf("seed %d: the door %v does not touch its parcel %v", seed, l.Door, l.Rect)
			}
		}
	}
}

// beside reports whether cell c touches rectangle r (or lies in it).
func beside(r Rect, c Pt) bool {
	return c.X >= r.X-1 && c.X <= r.X+r.W && c.Y >= r.Y-1 && c.Y <= r.Y+r.H
}

// TestPlanOwnership: every parcel belongs to one line, and a line holds no
// more than its share.
func TestPlanOwnership(t *testing.T) {
	_, p := groundPlan(7)
	seen := 0
	for l, lots := range p.Order {
		seen += len(lots)
		for _, lot := range lots {
			if p.Owner[lot] != l {
				t.Fatalf("lot %d listed under line %d, owned by %d", lot, l, p.Owner[lot])
			}
		}
	}
	if seen != len(p.Lots) {
		t.Fatalf("%d lots handed out of %d", seen, len(p.Lots))
	}
}
