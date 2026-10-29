package mapgen

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

// TestGalaxyShapesVary: twenty seeds give at least twelve different galaxies,
// with different arm counts, and some have a bar and some do not.
func TestGalaxyShapesVary(t *testing.T) {
	shapes := map[GalaxyShape]bool{}
	arms, bars := map[int]bool{}, map[bool]bool{}
	quads, many, barred, flatMin, flatMax := map[int]bool{}, 0, 0, 100, 0
	for seed := int64(1); seed <= 20; seed++ {
		s := pickShape(seed)
		shapes[s] = true
		arms[s.Arms], bars[s.Bar > 0] = true, true
		quads[s.Rot*4/turn] = true
		if s.Arms >= 3 {
			many++
		}
		if s.Bar > 0 {
			barred++
		}
		flatMin, flatMax = min(flatMin, s.Flat), max(flatMax, s.Flat)
		if s.Arms < 2 || s.Arms > 4 || s.Faint < 0 || s.Faint > 2 || s.Home >= s.Arms {
			t.Errorf("seed %d: arms %d faint %d home %d out of range", seed, s.Arms, s.Faint, s.Home)
		}
	}
	if len(shapes) < 12 {
		t.Errorf("20 seeds gave %d different galaxies, want at least 12", len(shapes))
	}
	if len(quads) < 3 || many < 6 || many > 14 || barred < 3 || barred > 12 || flatMax-flatMin < 30 {
		t.Errorf("rotations in %d quadrants, %d of 20 with 3 or 4 arms, %d barred, flattening %d to %d: too alike", len(quads), many, barred, flatMin, flatMax)
	}
	if len(arms) < 3 || len(bars) < 2 {
		t.Errorf("arm counts %v, bar or not %v: too little variety", arms, bars)
	}
}

// TestGalaxyStable: one seed is the same galaxy every time.
func TestGalaxyStable(t *testing.T) {
	a, b := NewGalaxy(7), NewGalaxy(7)
	if a.Shape != b.Shape || !slices.Equal(a.Dens, b.Dens) || !slices.Equal(a.Sky, b.Sky) || !slices.Equal(a.Home, b.Home) {
		t.Error("the same seed drew two galaxies")
	}
	if c := NewGalaxy(8); slices.Equal(a.Sky, c.Sky) {
		t.Error("two seeds drew the same sky")
	}
}

// TestGalaxyFillsTheMap: the arms reach most of the way across the world and
// the sky has no text on it, only star glyphs.
func TestGalaxyFillsTheMap(t *testing.T) {
	const stars = " ·∙+✦✧◌"
	for _, seed := range []int64{4, 7, 23, 31, 58} {
		g := NewGalaxy(seed)
		minX, maxX, minY, maxY, lit := worldW, 0, worldH, 0, 0
		for i, d := range g.Dens {
			if d > 40 {
				x, y := i%worldW, i/worldW
				minX, maxX, minY, maxY = min(minX, x), max(maxX, x), min(minY, y), max(maxY, y)
				lit++
			}
		}
		if maxX-minX < worldW/2 || maxY-minY < worldH/3 || lit < worldW*worldH/12 {
			t.Errorf("seed %d: the galaxy spans %dx%d with %d lit cells", seed, maxX-minX, maxY-minY, lit)
		}
		for i, c := range g.Sky {
			if !slices.Contains([]rune(stars), c.R) {
				t.Fatalf("seed %d: cell %d shows %q, not a star", seed, i, c.R)
			}
		}
	}
}

// TestSectorsFollowTheHomeArm: every sector of the town has its hub on the
// home arm, in the world, and no two systems anywhere share a cell or sit side
// by side.
func TestSectorsFollowTheHomeArm(t *testing.T) {
	for _, seed := range []int64{4, 7, 23, 58, 101} {
		p := NewStarPlan(seed, testCatalog)
		if len(p.Sectors) != len(mapmodel.LineageOrder) || len(p.Rivals) != rivalSlots {
			t.Fatalf("seed %d: %d sectors and %d rival places", seed, len(p.Sectors), len(p.Rivals))
		}
		seen := map[Pt]bool{}
		for _, sec := range append(slices.Clone(p.Sectors), p.Rivals...) {
			for _, c := range sec.Cells {
				if !inWorld(c.X, c.Y) || seen[c] || seen[Pt{c.X - 1, c.Y}] || seen[Pt{c.X + 1, c.Y}] {
					t.Errorf("seed %d: system %v is off the world, doubled or touching another", seed, c)
				}
				seen[c] = true
			}
		}
		for _, sec := range p.Sectors {
			near := false
			for _, h := range p.Galaxy.Home {
				if abs(h.X-sec.Hub.X) <= 3 && abs(h.Y-sec.Hub.Y) <= 3 {
					near = true
					break
				}
			}
			if !near {
				t.Errorf("seed %d: hub %v is not on the home arm", seed, sec.Hub)
			}
		}
	}
}

// TestSystemsGrowWithTheCount: more buildings light more systems, steeply
// enough that a few hundred fill a sector and a handful do not, and the
// biggest line stays inside its sector.
func TestSystemsGrowWithTheCount(t *testing.T) {
	last := 0
	for _, n := range []int{0, 1, 5, 20, 100, 500, 2000} {
		got := systemsFor(n, sectorMax)
		if got < last || got > sectorMax || n > 0 && got < 1 {
			t.Errorf("%d buildings light %d systems after %d", n, got, last)
		}
		last = got
	}
	if systemsFor(200, sectorMax) < 3*systemsFor(15, sectorMax) {
		t.Errorf("growth is too flat: 15 lights %d, 200 lights %d", systemsFor(15, sectorMax), systemsFor(200, sectorMax))
	}
}

// TestStarMapBoundsHoldTheTerritory: the overview frames the galaxy's body
// and every sector, rival and rift cell, and nothing like the whole world.
func TestStarMapBoundsHoldTheTerritory(t *testing.T) {
	for _, seed := range []int64{4, 23, 58, 101} {
		sc := (&Generator{}).Draw(model(state("galactic_age", seed, noon)), 0)
		b := sc.Bounds
		in := func(p Pt) bool { return p.X >= b.X && p.X < b.X+b.W && p.Y >= b.Y && p.Y < b.Y+b.H }
		p := NewStarPlan(seed, model(state("galactic_age", seed, noon)).Catalog)
		for _, list := range [][]Sector{p.Sectors, p.Rivals} {
			for _, sec := range list {
				for _, c := range sec.Cells {
					if !in(c) {
						t.Errorf("seed %d: bounds %+v miss the system %v", seed, b, c)
					}
				}
			}
		}
		hm := model(state("galactic_age", seed, noon))
		hm.Harbinger = &mapmodel.Harbinger{Key: "x"}
		hb := (&Generator{}).Draw(hm, 0).Bounds
		for _, c := range p.Rift {
			if c.X < hb.X || c.X >= hb.X+hb.W || c.Y < hb.Y || c.Y >= hb.Y+hb.H {
				t.Errorf("seed %d: bounds %+v miss the rift cell %v", seed, hb, c)
			}
		}
		if b.W < worldW*5/10 || b.H < worldH/3 || b.W > worldW*9/10 {
			t.Errorf("seed %d: bounds %+v are the wrong size for a galaxy", seed, b)
		}
	}
}

// TestTheCloseSkyIsPlainDarkSpace: the sky for a close look has no core and no
// bright stars, and thins the backdrop to a fraction of the overview's.
func TestTheCloseSkyIsPlainDarkSpace(t *testing.T) {
	for _, seed := range []int64{4, 23, 58} {
		g := NewGalaxy(seed)
		stars, deep := 0, 0
		for i, c := range g.Deep {
			if c.Bg.Hue == HCore || c.R == '✦' || c.R == '✧' || c.R == '+' || c.R == '◌' {
				t.Fatalf("seed %d: the close sky holds %q on %v", seed, c.R, c.Bg.Hue)
			}
			if g.Sky[i].R != ' ' {
				stars++
			}
			if c.R != ' ' {
				deep++
			}
		}
		if deep*2 > stars {
			t.Errorf("seed %d: %d backdrop stars close up against %d in the overview", seed, deep, stars)
		}
	}
}
