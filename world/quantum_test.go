package world

import "testing"

// A settled place keeps the version the seed and its id pick: the same each
// time, inside the versions it has, and the picks do vary from place to place.
func TestPickVersionIsStableForASeed(t *testing.T) {
	seen := map[int]bool{}
	for id := galStride; id < galStride+200; id++ {
		o := GenStar(7, id, Vec{}, true)
		if o.Doubt < 2 || o.Doubt > 3 {
			t.Fatalf("star %d has %d versions", id, o.Doubt)
		}
		n := PickVersion(7, &o)
		if n < 0 || n >= int(o.Doubt) || n != PickVersion(7, &o) {
			t.Fatalf("star %d picked version %d of %d", id, n, o.Doubt)
		}
		seen[n] = true
	}
	if len(seen) < 2 {
		t.Error("every place picked the same version")
	}
	if GenStar(7, 5000, Vec{}, true).Doubt != 0 {
		t.Error("a Milky Way star is uncertain")
	}
}
