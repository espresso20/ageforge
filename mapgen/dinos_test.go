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

// TestPrimitiveAgePristineWilderness: verifies that Primitive Age renders cleanly
// without dinosaur or asteroid layers.
func TestPrimitiveAgePristineWilderness(t *testing.T) {
	var gen Generator
	for _, seed := range []int64{1, 3, 7} {
		m := dinoModel(seed, 0)
		if dinos := m.Dinos(100); len(dinos) != 0 {
			t.Fatalf("seed %d: expected 0 dinos, got %d", seed, len(dinos))
		}
		if _, ok := m.AsteroidAt(100); ok {
			t.Fatalf("seed %d: expected no asteroid", seed)
		}
		sc := gen.DrawAt(m, Moment{Anim: 100, Visit: -1})
		if sc == nil || sc.Grid == nil {
			t.Fatalf("seed %d: failed to draw primitive map", seed)
		}
	}
}

// TestDinoHuesEmpty: verifies DinoHues returns nil.
func TestDinoHuesEmpty(t *testing.T) {
	if hues := DinoHues(); len(hues) != 0 {
		t.Errorf("DinoHues() returned %v, want empty", hues)
	}
}

// TestPrimitivePictureIsDeterministic: the same seed, frame and share draw the
// same picture from a fresh generator.
func TestPrimitivePictureIsDeterministic(t *testing.T) {
	m := dinoModel(7, 0)
	f := 120
	a := (&Generator{}).DrawAt(m, Moment{Anim: f, Visit: f}).Grid
	b := (&Generator{}).DrawAt(m, Moment{Anim: f, Visit: f}).Grid
	if !reflect.DeepEqual(a.Cells, b.Cells) {
		t.Error("two draws of one frame differ")
	}
}
