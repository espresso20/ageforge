package mapgen

import "testing"

// spacePlans are the plans of the three space scales, with the ages whose
// wonders each holds a plot for.
var spacePlans = []struct {
	name string
	plan func(seed int64) *Plan
	ages []int
}{
	{"orbit", func(s int64) *Plan { return NewOrbitPlan(s, testCatalog) }, []int{17}},
	{"system", func(s int64) *Plan { return NewSystemPlan(s, testCatalog) }, []int{18}},
	{"stars", func(s int64) *Plan { return NewStarPlan(s, testCatalog) }, []int{19, 20, 21}},
}

func TestSpacePlansDeterministic(t *testing.T) {
	for _, sp := range spacePlans {
		a, b := sp.plan(7), sp.plan(7)
		if len(a.Lots) != len(b.Lots) || len(a.Roads) != len(b.Roads) || a.Lots[0] != b.Lots[0] || a.Lots[len(a.Lots)-1] != b.Lots[len(b.Lots)-1] {
			t.Errorf("%s: same seed, different plan", sp.name)
		}
	}
}

// TestSpacePlansHoldTogether: every road cell is in the world and joins the
// hub, every parcel's door is a cell of its road, no parcel covers another or
// a road that is not its own ring's, and each scale has room for the wonders
// of its ages.
func TestSpacePlansHoldTogether(t *testing.T) {
	for _, sp := range spacePlans {
		for _, seed := range []int64{4, 7, 23} {
			p := sp.plan(seed)
			if len(p.Lots) < 60 {
				t.Errorf("%s/%d: only %d parcels", sp.name, seed, len(p.Lots))
			}
			road := map[Pt]int{}
			for i, r := range p.Roads {
				for _, c := range r.Cells {
					if !inWorld(c.X, c.Y) {
						t.Fatalf("%s/%d: road %d leaves the world at %v", sp.name, seed, i, c)
					}
					road[c] = i
				}
			}
			taken := map[Pt]bool{}
			for _, l := range append(append([]Lot{}, p.Lots...), p.Plots...) {
				if l.Rect.W == 0 {
					continue
				}
				if p.Roads[l.Road].Cells[l.At] != l.Door {
					t.Fatalf("%s/%d: door %v is not cell %d of road %d", sp.name, seed, l.Door, l.At, l.Road)
				}
				l.Rect.cells(func(x, y int) {
					if taken[Pt{x, y}] {
						t.Fatalf("%s/%d: two parcels cover (%d,%d)", sp.name, seed, x, y)
					}
					taken[Pt{x, y}] = true
				})
			}
			reached := map[Pt]bool{p.Roads[0].Cells[0]: true}
			queue := []Pt{p.Roads[0].Cells[0]}
			for head := 0; head < len(queue); head++ {
				for _, d := range neighbours {
					n := Pt{queue[head].X + d[0], queue[head].Y + d[1]}
					if _, onRoad := road[n]; onRoad && !reached[n] {
						reached[n] = true
						queue = append(queue, n)
					}
				}
			}
			if len(reached) != len(road) {
				t.Errorf("%s/%d: %d road cells, %d reached from the hub", sp.name, seed, len(road), len(reached))
			}
			for _, age := range sp.ages {
				if p.Plots[age].Rect.W == 0 {
					t.Errorf("%s/%d: no plot for the wonder of age %d", sp.name, seed, age)
				}
			}
		}
	}
}

// TestRingParcelsStandOnTheirOwnRing: on the ring scales a parcel covers no
// road but its own ring.
func TestRingParcelsStandOnTheirOwnRing(t *testing.T) {
	for _, sp := range spacePlans[:2] {
		p := sp.plan(7)
		road := map[Pt]int{}
		for i, r := range p.Roads {
			for _, c := range r.Cells {
				road[c] = i
			}
		}
		for _, l := range p.Lots {
			l.Rect.cells(func(x, y int) {
				if i, onRoad := road[Pt{x, y}]; onRoad && (i < 2 || ringOf(i) != ringOf(l.Road)) {
					t.Fatalf("%s: a parcel on ring %d covers road %d at (%d,%d)", sp.name, ringOf(l.Road), i, x, y)
				}
			})
		}
	}
}

func TestHalfRingsAreClosedCurves(t *testing.T) {
	for _, ry := range []int{4, 7, 16} {
		east, west := halfRing(60, 40, ry, 1), halfRing(60, 40, ry, -1)
		for _, half := range [][]Pt{east, west} {
			for i := 1; i < len(half); i++ {
				if abs(half[i].X-half[i-1].X)+abs(half[i].Y-half[i-1].Y) != 1 {
					t.Fatalf("ry %d: cells %v and %v do not touch", ry, half[i-1], half[i])
				}
			}
			if half[0].Y != 40-ry || half[len(half)-1].Y != 40+ry {
				t.Errorf("ry %d: a half ring runs from row %d to %d, want %d to %d", ry, half[0].Y, half[len(half)-1].Y, 40-ry, 40+ry)
			}
		}
	}
}
