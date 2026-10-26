package mapgen

import "testing"

func TestOrbitPlanDeterministic(t *testing.T) {
	a, b := NewOrbitPlan(7, testCatalog), NewOrbitPlan(7, testCatalog)
	if len(a.Lots) != len(b.Lots) || a.Lots[0] != b.Lots[0] || a.Lots[len(a.Lots)-1] != b.Lots[len(b.Lots)-1] {
		t.Fatal("same seed, different orbit plan")
	}
}

// TestOrbitPlanStations: stations stand clear of each other and of every
// ring but their own, each sits on a ring cell, and the rings are one piece
// with the elevator, so every station is reachable.
func TestOrbitPlanStations(t *testing.T) {
	p := NewOrbitPlan(7, testCatalog)
	if len(p.Lots) < 120 {
		t.Fatalf("only %d stations", len(p.Lots))
	}
	road := map[Pt]int{}
	for i, r := range p.Roads {
		for _, c := range r.Cells {
			road[c] = i
		}
	}
	taken := map[Pt]bool{}
	for _, l := range append(append([]Lot{}, p.Lots...), p.Plots...) {
		if l.Rect.W == 0 {
			continue
		}
		if road[l.Door] != l.Road {
			t.Fatalf("door %v is not on road %d", l.Door, l.Road)
		}
		l.Rect.cells(func(x, y int) {
			if taken[Pt{x, y}] {
				t.Fatalf("two stations cover (%d,%d)", x, y)
			}
			taken[Pt{x, y}] = true
			if i, onRoad := road[Pt{x, y}]; onRoad && i != l.Road && (i < 2 || ringOf(i) != ringOf(l.Road)) {
				t.Fatalf("a station on ring %d covers road %d at (%d,%d)", ringOf(l.Road), i, x, y)
			}
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
		t.Fatalf("%d road cells, %d reached from the hub", len(road), len(reached))
	}
}

// TestOrbitWondersHaveRoom: the wonders of the orbit ages each have a plot.
func TestOrbitWondersHaveRoom(t *testing.T) {
	p := NewOrbitPlan(7, testCatalog)
	for age := groundAges; age < len(p.Plots); age++ {
		if p.Plots[age].Rect.W == 0 {
			t.Errorf("no plot for the wonder of age %d", age)
		}
	}
}
