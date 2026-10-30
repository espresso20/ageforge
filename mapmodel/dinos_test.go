package mapmodel

import "testing"

// dinoCount counts the dinosaurs that start in the first n frames, sampled
// at a few seeds so one lucky hash proves nothing.
func dinoCount(share float64, n int) int {
	total := 0
	for _, seed := range []int64{1, 7, 42, -3} {
		for f := 0; f < n; f += 8 {
			total += len(DinosAt(seed, share, f))
		}
	}
	return total
}

func TestDinoChanceFallsInAStraightLine(t *testing.T) {
	if DinoChance(0) != 1000 || DinoChance(dinoGateEnd) != 0 || DinoChance(1) != 0 || DinoChance(-1) != 1000 {
		t.Fatalf("ends: %d %d %d", DinoChance(0), DinoChance(dinoGateEnd), DinoChance(1))
	}
	if c := DinoChance(dinoGateEnd / 2); c < 495 || c > 505 {
		t.Errorf("halfway chance %d, want about 500", c)
	}
}

func TestDinosThinOutAsTheGateFills(t *testing.T) {
	n := 200 * framesPerHour / 100
	c0, c5, c9, c95 := dinoCount(0, n), dinoCount(0.5, n), dinoCount(0.9, n), dinoCount(0.95, n)
	if c0 == 0 {
		t.Fatal("no dinosaurs at share 0")
	}
	if !(c0 > c5 && c5 > c9 && c9 > 0) {
		t.Errorf("counts at 0, 0.5, 0.9: %d %d %d, want falling and not zero", c0, c5, c9)
	}
	if c95 != 0 || dinoCount(1, n) != 0 {
		t.Errorf("dinosaurs at share 0.95 and above: %d", c95)
	}
}

func TestDinoScheduleIsDeterministicAndWholeKinds(t *testing.T) {
	kinds := map[SightingKind]bool{}
	for f := 0; f < 40*framesPerHour; f += 4 {
		a, b := DinosAt(3, 0, f), DinosAt(3, 0, f)
		if len(a) != len(b) {
			t.Fatalf("frame %d differs between calls", f)
		}
		for i, s := range a {
			if s != b[i] {
				t.Fatalf("frame %d differs between calls", f)
			}
			kinds[s.Kind] = true
			if f < s.Start || f >= s.Start+s.Frames {
				t.Fatalf("frame %d reports a sighting that is not on", f)
			}
		}
	}
	if len(kinds) != 3 {
		t.Errorf("kinds seen %v, want all three", kinds)
	}
}

func TestModelDinosOnlyInPrimitive(t *testing.T) {
	cat := NewCatalog()
	for i, age := range cat.Ages {
		m := &Model{Seed: 1, Catalog: cat, AgeIdx: i}
		got := 0
		for f := 0; f < 20*framesPerHour; f += 8 {
			got += len(m.Dinos(f))
		}
		if (age == "primitive_age") != (got > 0) {
			t.Errorf("%s: %d dinosaur frames", age, got)
		}
	}
}

func TestAsteroidWindowAndCadence(t *testing.T) {
	cat := NewCatalog()
	s0 := AsteroidShare()
	if s0 < 0.03 || s0 > 0.06 {
		t.Fatalf("asteroid share %v, want near 0.044", s0)
	}
	prim := cat.AgeIdx["primitive_age"]
	for _, c := range []struct {
		share float64
		age   int
		want  bool
	}{{s0, prim, true}, {s0 + 0.009, prim, true}, {s0 - 0.009, prim, true}, {s0 + 0.011, prim, false}, {s0 - 0.011, prim, false}, {0, prim, false}, {0.5, prim, false}, {s0, prim + 1, false}} {
		m := &Model{Seed: 5, Catalog: cat, AgeIdx: c.age, GateShare: c.share}
		seen, gap, worst := false, 0, 0
		for f := 0; f < 40*asteroidSlot; f++ {
			if _, ok := m.AsteroidAt(f); ok {
				seen, gap = true, 0
			} else if gap++; gap > worst {
				worst = gap
			}
		}
		if seen != c.want {
			t.Errorf("share %v age %d: seen %v, want %v", c.share, c.age, seen, c.want)
		}
		if c.want && worst >= 2*asteroidSlot {
			t.Errorf("share %v: %d frames with no streak", c.share, worst)
		}
	}
}
