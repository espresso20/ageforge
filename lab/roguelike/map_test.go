package main

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

func testView(t *testing.T, age string, seed int64) MapView {
	t.Helper()
	v, err := loadView("states", age, seed)
	if err != nil {
		t.Skipf("no cached state for %s: %v", age, err)
	}
	return v
}

func tilesByKey(s *Scene) map[string][]Pt {
	m := map[string][]Pt{}
	for _, b := range s.Blds {
		m[b.V.Key] = append([]Pt(nil), b.Tiles...)
	}
	return m
}

// Building more of the current age's buildings must never move anything
// already on the map.
func TestGrowthNeverMovesNeighbours(t *testing.T) {
	for _, age := range []string{"primitive_age", "medieval_age", "cyberpunk_age"} {
		mv := testView(t, age, 7)
		w := NewWorld(7, worldW, worldH)
		p := NewPlan(w)
		before := tilesByKey(Build(w, p, mv, false))

		// grow the newest tier's largest building across a tile boundary
		top, pick := -1, -1
		for i, b := range mv.Buildings {
			if b.Category == "wonder" || b.Count == 0 {
				continue
			}
			if b.Tier > top || b.Tier == top && tilesForLineage(b.Lineage, b.Count) > tilesForLineage(mv.Buildings[pick].Lineage, mv.Buildings[pick].Count) {
				top, pick = b.Tier, i
			}
		}
		grown := mv
		grown.Buildings = append([]BuildingView(nil), mv.Buildings...)
		b := &grown.Buildings[pick]
		n0 := tilesForLineage(b.Lineage, b.Count)
		for tilesForLineage(b.Lineage, b.Count) == n0 && b.Count < 100000 {
			b.Count++
		}
		after := tilesByKey(Build(w, p, grown, false))
		for k, tiles := range before {
			got := after[k]
			if k == b.Key {
				got = got[:len(tiles)]
			}
			if !reflect.DeepEqual(tiles, got) {
				t.Errorf("%s: growing %s moved %s", age, b.Key, k)
			}
		}
		if len(after[b.Key]) <= len(before[b.Key]) {
			t.Errorf("%s: %s did not grow on the map", age, b.Key)
		}
	}
}

// Same seed, same map; another seed, another map.
func TestDeterministicAndPersonal(t *testing.T) {
	mv := testView(t, "medieval_age", 7)
	w := NewWorld(7, worldW, worldH)
	a := tilesByKey(Build(w, NewPlan(w), mv, false))
	b := tilesByKey(Build(w, NewPlan(NewWorld(7, worldW, worldH)), mv, false))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed produced a different layout")
	}
	w2 := NewWorld(11, worldW, worldH)
	c := tilesByKey(Build(w2, NewPlan(w2), mv, false))
	if reflect.DeepEqual(a, c) {
		t.Fatal("a different seed produced the same layout")
	}
}

// Every zoom at every size draws without panicking, including tiny ones.
func TestDrawAnySize(t *testing.T) {
	_ = theme.SetActive("forge")
	defer theme.SetActive("forge")
	for _, age := range []string{"primitive_age", "galactic_age"} {
		mv := testView(t, age, 7)
		w := NewWorld(7, worldW, worldH)
		v := NewView(Build(w, NewPlan(w), mv, true))
		for _, sz := range [][2]int{{1, 1}, {12, 5}, {40, 15}, {80, 24}, {100, 30}, {160, 48}, {240, 70}} {
			for z := ZRegion; z <= ZDistrict; z++ {
				v.Zoom = z
				for f := 0; f < 3; f++ {
					v.Frame = f
					fr := snap(v, sz[0], sz[1], false)
					if fr.W != sz[0] || fr.H != sz[1] {
						t.Fatalf("%s %v: got %dx%d", age, sz, fr.W, fr.H)
					}
				}
			}
			snap(v, sz[0], sz[1], true)
		}
		// the cursor can go anywhere
		for _, c := range [][2]int{{0, 0}, {w.W - 1, w.H - 1}, {w.W / 2, 0}} {
			v.CurX, v.CurY = c[0], c[1]
			v.Zoom = ZSettlement
			snap(v, 80, 24, false)
			v.NextBuilding(1)
		}
	}
}
