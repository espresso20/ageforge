package mapgen

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

// dinoModel is a Primitive Age model at a gate share.
func dinoModel(seed int64, share float64) *mapmodel.Model {
	m := model(state("primitive_age", seed, noon))
	m.GateShare = share
	return m
}

// extras lists the cells a frame draws differently from the same frame with
// nothing about (the gate full), with the ground's thing at each.
func extras(gen *Generator, m *mapmodel.Model, anim int) (cells []Pt, kinds []Kind, g *Grid) {
	bare := *m
	bare.GateShare = 1
	a := gen.DrawAt(m, Moment{Anim: anim, Visit: -1}).Grid
	b := gen.DrawAt(&bare, Moment{Anim: anim, Visit: -1}).Grid
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			p := Pt{i % a.W, i / a.W}
			cells = append(cells, p)
			kinds = append(kinds, b.ThingAt(p.X, p.Y).Kind)
		}
	}
	return cells, kinds, a
}

// dinoFrame finds a frame at which the model has a dinosaur of kind on.
func dinoFrame(m *mapmodel.Model, kind mapmodel.SightingKind) (mapmodel.Sighting, bool) {
	for f := 0; f < 60*8*3600; f += 8 {
		for _, s := range m.Dinos(f) {
			if s.Kind == kind {
				return s, true
			}
		}
	}
	return mapmodel.Sighting{}, false
}

// TestDinosDrawOnOpenGroundOnly: each dinosaur shows up on the map, inside
// it, and never over a building, a wonder or scenery.
func TestDinosDrawOnOpenGroundOnly(t *testing.T) {
	var gen Generator
	for _, seed := range []int64{1, 3, 7} {
		m := dinoModel(seed, 0)
		for _, kind := range []mapmodel.SightingKind{mapmodel.SightSauropod, mapmodel.SightRaptor, mapmodel.SightPterodactyl} {
			sg, ok := dinoFrame(m, kind)
			if !ok {
				t.Fatalf("seed %d: no sighting of kind %d", seed, kind)
			}
			drawn := 0
			for pct := 0; pct < 100; pct += 5 {
				cells, kinds, _ := extras(&gen, m, sg.Start+sg.Frames*pct/100)
				drawn = max(drawn, len(cells))
				for i, p := range cells {
					if !inWorld(p.X, p.Y) {
						t.Errorf("seed %d kind %d: cell %v outside the map", seed, kind, p)
					}
					switch kinds[i] {
					case KindBuilding, KindWonder, KindAmbient:
						t.Errorf("seed %d kind %d: drawn over a %v at %v", seed, kind, kinds[i], p)
					}
				}
			}
			if drawn == 0 {
				t.Errorf("seed %d kind %d: never drawn through its visit", seed, kind)
			}
		}
	}
}

// TestNoDinosOutsidePrimitive: another age's picture is the same at share 0
// and with the gate full.
func TestNoDinosOutsidePrimitive(t *testing.T) {
	var gen Generator
	for _, age := range []string{"stone_age", "bronze_age", "modern_age"} {
		m := model(state(age, 3, noon))
		for _, share := range []float64{0, 0.04} {
			m.GateShare = share
			for f := 0; f < 3*mapmodel.SightingBlock; f += 977 {
				if cells, _, _ := extras(&gen, m, f); len(cells) != 0 {
					t.Fatalf("%s share %v frame %d: %d cells differ", age, share, f, len(cells))
				}
			}
		}
	}
}

// TestDinoSpritesAreAsciiFoldable: the ASCII tier folds every dinosaur
// glyph to plain ASCII.
func TestDinoSpritesAreAsciiFoldable(t *testing.T) {
	for _, s := range []sprite{sauropod, raptor, pterodactyl[0], pterodactyl[1], sauropod.west(), raptor.west()} {
		for _, row := range s.Rows {
			for _, r := range row {
				if f := mapmodel.Fold(r, mapmodel.TierASCII); f > 0x7e || f < ' ' || f == '?' {
					t.Errorf("%q folds to %q", r, f)
				}
			}
		}
	}
	if w := sauropod.width(); w < 5 || w > 7 || len(sauropod.Rows) != 2 {
		t.Errorf("sauropod is %d wide, %d rows", w, len(sauropod.Rows))
	}
	if w := raptor.width(); w < 2 || w > 3 {
		t.Errorf("raptor is %d wide", w)
	}
	if reflect.DeepEqual(sauropod.west().Rows, sauropod.Rows) {
		t.Error("the sauropod does not turn")
	}
}

// TestAsteroidDrawsInItsWindowOnly: the streak is on the map at some frame
// of every slot inside the window, and never outside it.
func TestAsteroidDrawsInItsWindowOnly(t *testing.T) {
	var gen Generator
	s0 := mapmodel.AsteroidShare()
	for _, c := range []struct {
		share float64
		want  bool
	}{{s0, true}, {s0 + 0.008, true}, {0, false}, {s0 + 0.02, false}, {0.5, false}} {
		m := dinoModel(3, c.share)
		// dinosaurs may be about too: the streak is told by its fire colours
		seen := 0
		for slot := 0; slot < 4; slot++ {
			got := false
			for f := slot * 80; f < (slot+1)*80; f++ {
				if _, ok := m.AsteroidAt(f); !ok {
					continue
				}
				cells, _, g := extras(&gen, m, f)
				for _, p := range cells {
					switch g.Cells[at(p.X, p.Y)].Fg.Hue {
					case HAmber, HFire, HRed:
						got = true
					}
				}
			}
			if got {
				seen++
			}
		}
		if (seen > 0) != c.want || (c.want && seen != 4) {
			t.Errorf("share %v: streak drawn in %d of 4 slots, want %v", c.share, seen, c.want)
		}
	}
}

// TestDinoPictureIsDeterministic: the same seed, frame and share draw the
// same picture from a fresh generator.
func TestDinoPictureIsDeterministic(t *testing.T) {
	m := dinoModel(7, 0)
	sg, ok := dinoFrame(m, mapmodel.SightSauropod)
	if !ok {
		t.Fatal("no sauropod")
	}
	f := sg.Start + sg.Frames/2
	a := (&Generator{}).DrawAt(m, Moment{Anim: f, Visit: f}).Grid
	b := (&Generator{}).DrawAt(m, Moment{Anim: f, Visit: f}).Grid
	if !reflect.DeepEqual(a.Cells, b.Cells) {
		t.Error("two draws of one frame differ")
	}
}
