package mapgen

import (
	"reflect"
	"testing"
)

func TestGenSystemIsStable(t *testing.T) {
	a, b := GenSystem(42, 3, 2), GenSystem(42, 3, 2)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("the same save and star gave two systems")
	}
	same := 0
	for k := 1; k < 20; k++ {
		if reflect.DeepEqual(GenSystem(42, k, 0), GenSystem(42, k+1, 0)) {
			same++
		}
		if reflect.DeepEqual(GenSystem(42, k, 0), GenSystem(43, k, 0)) {
			same++
		}
	}
	if same > 0 {
		t.Errorf("%d neighbouring stars or saves came out alike", same)
	}
}

func TestGenSystemRanges(t *testing.T) {
	kinds := map[Hue]bool{}
	for seed := int64(1); seed <= 60; seed++ {
		for k := 1; k <= 30; k++ {
			sp := GenSystem(seed, k, 0)
			kinds[sp.Star.Mid] = true
			if n := len(sp.Planets); n < 1 || n > 10 {
				t.Fatalf("seed %d star %d: %d planets", seed, k, n)
			}
			if len(sp.Belts) > 2 {
				t.Fatalf("%d belts", len(sp.Belts))
			}
			last := 0
			for _, p := range sp.Planets {
				if p.Size < 18 || p.Size > 71 || p.Moons < 0 || p.Moons > 20 || p.Orbit <= last || p.Orbit > 98 {
					t.Fatalf("seed %d star %d: bad planet %+v after orbit %d", seed, k, p, last)
				}
				if p.Orbit < sp.Star.Size/5 {
					t.Fatalf("a planet inside its star: %+v", p)
				}
				last = p.Orbit
			}
		}
	}
	if len(kinds) < 3 {
		t.Errorf("only %d kinds of star in 1800 systems", len(kinds))
	}
}

func TestSolRecipe(t *testing.T) {
	if len(SolSpec.Planets) != 8 || !SolSpec.Planets[5].Ring || !SolSpec.Planets[planetEarth].Home {
		t.Fatal("Sol is not eight planets with Saturn ringed and the home world third")
	}
	rings := 0
	for _, p := range SolSpec.Planets {
		if p.Ring {
			rings++
		}
	}
	if rings != 1 {
		t.Errorf("%d ringed planets in Sol", rings)
	}
}

func TestNoStationsBeforeInterstellar(t *testing.T) {
	for seed := int64(1); seed <= 40; seed++ {
		for k := 1; k <= 20; k++ {
			if n := GenSystem(seed, k, 0).StationCount(); n != 0 {
				t.Fatalf("%d stations before the interstellar ages", n)
			}
			a, b := GenSystem(seed, k, 1).StationCount(), GenSystem(seed, k, 9).StationCount()
			if a < 1 || b < a || b > maxStations {
				t.Fatalf("stations at progress 1 and 9: %d, %d", a, b)
			}
		}
	}
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 0)
	if sc.Progress != 0 {
		t.Logf("progress at the start of the age: %d", sc.Progress)
	}
}

func TestStarSitesAreStable(t *testing.T) {
	a, b := starSites(5, 12), starSites(5, 12)
	if !reflect.DeepEqual(a, b) || len(a) < 4 {
		t.Fatalf("sites %v / %v", a, b)
	}
}

// TestSystemOverviewSameSeedSameFrame: a generated system, and the cosmos with
// its stars, draw the same twice and differ between stars.
func TestSystemOverviewSameSeedSameFrame(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 0)
	for _, mode := range []SystemViewMode{SystemViewCosmos, SystemViewFull, SystemViewInner} {
		a, b := sc.SystemOverview(160, 50, mode, 3), sc.SystemOverview(160, 50, mode, 3)
		if !reflect.DeepEqual(a.Grid.Cells, b.Grid.Cells) {
			t.Errorf("mode %d: two frames of one system differ", mode)
		}
	}
	if reflect.DeepEqual(sc.SystemOverview(160, 50, SystemViewFull, 1).Grid.Cells, sc.SystemOverview(160, 50, SystemViewFull, 2).Grid.Cells) {
		t.Error("two stars drew the same system")
	}
	g := sc.OverviewCosmos(160, 50).Grid
	stars := map[int]bool{}
	for _, h := range g.Hit {
		if th := g.Things[h]; th.Key == "star" && th.Count == 0 {
			stars[th.Ref] = true
		}
	}
	if len(stars) < 3 {
		t.Errorf("the cosmos shows %d stars to zoom into", len(stars))
	}
}
