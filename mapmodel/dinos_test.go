package mapmodel

import "testing"

// onMap is the average number of dinosaurs of kind on the map, sampled once a
// second over about two hours on several seeds.
func onMap(kind SightingKind, share float64) float64 {
	total, samples := 0, 0
	for _, seed := range []int64{1, 7, 42, -3} {
		for f := 0; f < 2*framesPerHour; f += 8 {
			samples++
			for _, s := range DinosAt(seed, share, f) {
				if s.Kind == kind {
					total++
				}
			}
		}
	}
	return float64(total) / float64(samples)
}

func TestDinoFactorFallsInAStraightLine(t *testing.T) {
	if DinoFactor(0) != 1 || DinoFactor(dinoGateEnd) != 0 || DinoFactor(1) != 0 || DinoFactor(-1) != 1 {
		t.Fatalf("ends: %v %v %v", DinoFactor(0), DinoFactor(dinoGateEnd), DinoFactor(1))
	}
	if c := DinoFactor(dinoGateEnd / 2); c < 0.499 || c > 0.501 {
		t.Errorf("halfway factor %v, want 0.5", c)
	}
	// slots are the share-0 count scaled and rounded down
	for _, c := range []struct {
		kind  int
		share float64
		want  int
	}{{0, 0, 2}, {1, 0, 3}, {0, 0.5, 0}, {1, 0.5, 1}, {0, 0.9, 0}, {1, 0.9, 0}, {1, 0.95, 0}, {0, 0.94, 0}} {
		if got := DinoSlots(c.kind, c.share); got != c.want {
			t.Errorf("slots of kind %d at share %v: %d, want %d", c.kind, c.share, got, c.want)
		}
	}
}

// TestDinosRoamAndThinOut: at share 0 about two sauropods and two to three
// raptors are on the map at any moment and a pterodactyl crosses about every
// minute; the counts fall as the gate fills and none is left at 0.95.
func TestDinosRoamAndThinOut(t *testing.T) {
	sauro, raptor := onMap(SightSauropod, 0), onMap(SightRaptor, 0)
	if sauro < 1.5 || sauro > 2.01 {
		t.Errorf("sauropods on the map at share 0: %.2f, want about 2", sauro)
	}
	if raptor < 2 || raptor > 3.01 {
		t.Errorf("raptors on the map at share 0: %.2f, want 2 to 3", raptor)
	}
	// flybys: one visit in each 75 s period, so 48 an hour
	flybys := 0
	for f := 0; f < framesPerHour; f++ {
		for _, s := range DinosAt(3, 0, f) {
			if s.Kind == SightPterodactyl && s.Start == f {
				flybys++
			}
		}
	}
	if flybys < 40 || flybys > 60 {
		t.Errorf("%d flybys in an hour at share 0, want one every 60 to 90 seconds", flybys)
	}
	for _, k := range []SightingKind{SightSauropod, SightRaptor, SightPterodactyl} {
		c0, c5, c9, c95 := onMap(k, 0), onMap(k, 0.5), onMap(k, 0.9), onMap(k, 0.95)
		if !(c0 > c5 && c5 >= c9) || c0 <= c9 {
			t.Errorf("kind %d on the map at share 0, 0.5, 0.9: %.2f %.2f %.2f, want falling", k, c0, c5, c9)
		}
		if c95 != 0 || onMap(k, 1) != 0 {
			t.Errorf("kind %d on the map at share 0.95 and above: %.2f", k, c95)
		}
	}
	if s5 := onMap(SightSauropod, 0.5); s5 != 0 {
		t.Errorf("sauropods at share 0.5: %.2f; two slots scaled by 0.47 round down to none", s5)
	}
	if r5 := onMap(SightRaptor, 0.5); r5 < 0.5 || r5 > 1.01 {
		t.Errorf("raptors at share 0.5: %.2f, want about one", r5)
	}
}

func TestDinoScheduleIsDeterministicAndWholeKinds(t *testing.T) {
	kinds := map[SightingKind]bool{}
	for f := 0; f < 10*framesPerHour; f += 16 {
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
		for f := 0; f < 2*framesPerHour; f += 8 {
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
