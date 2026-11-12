package world

import "testing"

func TestZonesGrowOnlyWithLitSystems(t *testing.T) {
	g := NewGalaxy(7)
	prev := NewZones(7, g.Shape, solLy(g), 0)
	if !prev.Held[prev.Sol] {
		t.Fatal("Sol's sector is not held at nothing lit")
	}
	for lit := 1; lit <= 4*ZoneTotal+8; lit++ {
		z := NewZones(7, g.Shape, solLy(g), lit)
		for s := range z.Held {
			if prev.Held[s] && !z.Held[s] {
				t.Fatalf("lit %d lost sector %d", lit, s)
			}
		}
		prev = z
	}
	for s, h := range prev.Held {
		if !h {
			t.Errorf("sector %d still free with every system lit", s)
		}
	}
}

func TestZonesSameSeedSameSectors(t *testing.T) {
	g := NewGalaxy(7)
	a, b := NewZones(7, g.Shape, solLy(g), 9), NewZones(7, g.Shape, solLy(g), 9)
	if a.Held != b.Held || a.Names != b.Names {
		t.Error("the same seed made different sectors")
	}
	oa, ob := a.objects(7), b.objects(7)
	for i := range oa {
		if oa[i].Key != ob[i].Key || oa[i].Pos != ob[i].Pos {
			t.Fatalf("object %d differs", i)
		}
	}
	if c := NewZones(8, NewGalaxy(8).Shape, solLy(NewGalaxy(8)), 9); c.Names == a.Names {
		t.Error("another seed made the same names")
	}
}

func TestZoneStarsLieInTheirSectorAndRebuildFromKey(t *testing.T) {
	g := NewGalaxy(7)
	z := NewZones(7, g.Shape, solLy(g), 12)
	n := 0
	for _, o := range z.objects(7) {
		if o.Kind != KindStar {
			continue
		}
		n++
		if s := z.ZoneAt(o.Pos); s < 0 || z.Held[s] != o.Held {
			t.Errorf("%s: held %v in sector %d", o.Key, o.Held, s)
		}
		if id, ok := StarID(o.Key); !ok || StarKey(7, id) != o.Key {
			t.Errorf("%s does not rebuild from its key", o.Key)
		}
	}
	if n != ZoneTotal*zoneStars {
		t.Errorf("%d stars, want %d", n, ZoneTotal*zoneStars)
	}
}
