package mapgen

import (
	"reflect"
	"testing"
)

func TestFieldsDeterministic(t *testing.T) {
	a, b := newFields(7), newFields(7)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed, different land")
	}
	if reflect.DeepEqual(a.Height, newFields(8).Height) {
		t.Fatal("two seeds, same land")
	}
}

func TestFieldsHaveSeaLandAndRivers(t *testing.T) {
	for _, seed := range []int64{3, 7, 11, 19} {
		f := newFields(seed)
		count := map[Biome]int{}
		for _, b := range f.Biome {
			count[b]++
		}
		for _, b := range []Biome{Deep, Sand, Grass, River} {
			if count[b] == 0 {
				t.Errorf("seed %d: no %s", seed, biomeNames[b])
			}
		}
	}
}

// TestRiversNeverDeadEnd: every river cell joins the sea through river cells.
func TestRiversNeverDeadEnd(t *testing.T) {
	for _, seed := range []int64{3, 7, 11, 19} {
		f := newFields(seed)
		reached := make([]bool, len(f.Biome))
		var queue []int
		for i, h := range f.Height {
			if h < seaLevel {
				reached[i] = true
				queue = append(queue, i)
			}
		}
		for head := 0; head < len(queue); head++ {
			c := queue[head]
			for _, d := range neighbours {
				x, y := c%worldW+d[0], c/worldW+d[1]
				if inWorld(x, y) && !reached[at(x, y)] && f.Biome[at(x, y)] == River {
					reached[at(x, y)] = true
					queue = append(queue, at(x, y))
				}
			}
		}
		for i, b := range f.Biome {
			if b == River && !reached[i] {
				t.Fatalf("seed %d: river cell (%d,%d) does not reach the sea", seed, i%worldW, i/worldW)
			}
		}
	}
}
