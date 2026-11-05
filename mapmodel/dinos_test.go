package mapmodel

import "testing"

func TestNoDinosInPrimitiveOrAnyAge(t *testing.T) {
	cat := NewCatalog()
	for i, age := range cat.Ages {
		m := &Model{Seed: 1, Catalog: cat, AgeIdx: i}
		for f := 0; f < 2*framesPerHour; f += 8 {
			if got := m.Dinos(f); len(got) > 0 {
				t.Fatalf("%s: expected no dinos, got %d", age, len(got))
			}
			if _, ok := m.AsteroidAt(f); ok {
				t.Fatalf("%s: expected no asteroid", age)
			}
		}
	}
}

func TestDinoFactorAndSlotsZero(t *testing.T) {
	for _, share := range []float64{-1, 0, 0.5, 0.95, 1, 2} {
		if got := DinoFactor(share); got != 0 {
			t.Errorf("DinoFactor(%v) = %v, want 0", share, got)
		}
		for k := 0; k < 4; k++ {
			if got := DinoSlots(k, share); got != 0 {
				t.Errorf("DinoSlots(%d, %v) = %d, want 0", k, share, got)
			}
		}
	}
}

func TestDinosAtEmpty(t *testing.T) {
	for _, seed := range []int64{1, 7, 42, -3} {
		for f := 0; f < 1000; f += 8 {
			if got := DinosAt(seed, 0, f); len(got) != 0 {
				t.Fatalf("seed %d frame %d: DinosAt returned non-empty", seed, f)
			}
		}
	}
}
