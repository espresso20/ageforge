package mapgen

import (
	"slices"
	"testing"
)

// TestGalaxyShapesVary: twenty seeds give at least twelve different galaxies,
// with different arm counts, and some have a bar and some do not.
func TestGalaxyShapesVary(t *testing.T) {
	shapes := map[GalaxyShape]bool{}
	arms, bars := map[int]bool{}, map[bool]bool{}
	for seed := int64(1); seed <= 20; seed++ {
		s := pickShape(seed)
		shapes[s] = true
		arms[s.Arms], bars[s.Bar > 0] = true, true
		if s.Arms < 2 || s.Arms > 4 || s.Faint < 0 || s.Faint > 2 || s.Home >= s.Arms {
			t.Errorf("seed %d: arms %d faint %d home %d out of range", seed, s.Arms, s.Faint, s.Home)
		}
	}
	if len(shapes) < 12 {
		t.Errorf("20 seeds gave %d different galaxies, want at least 12", len(shapes))
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

// TestStarNodesFollowTheHomeArm: every system the town can use lies on the
// home arm, in the world, with none on another.
func TestStarNodesFollowTheHomeArm(t *testing.T) {
	for _, seed := range []int64{4, 7, 23} {
		p := NewStarPlan(seed, testCatalog)
		if len(p.Nodes) != starNodes+1 {
			t.Fatalf("seed %d: %d systems", seed, len(p.Nodes))
		}
		seen := map[Pt]bool{}
		for _, n := range p.Nodes {
			if !inWorld(n.X, n.Y) || seen[n] {
				t.Errorf("seed %d: system %v is off the world or doubled", seed, n)
			}
			seen[n] = true
		}
		for _, n := range p.Nodes[1:] {
			near := false
			for _, h := range p.Galaxy.Home {
				if abs(h.X-n.X) <= 3 && abs(h.Y-n.Y) <= 2 {
					near = true
					break
				}
			}
			if !near {
				t.Errorf("seed %d: system %v is not on the home arm", seed, n)
			}
		}
	}
}

// TestStarMapBoundsHoldTheGalaxy: the overview frames the whole galaxy, not
// just the few systems the town has laid.
func TestStarMapBoundsHoldTheGalaxy(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("galactic_age", 4, noon)), 0)
	if sc.Bounds.W < 2*galRX || sc.Bounds.H < 2*galRYBase*70/100 {
		t.Errorf("bounds %+v do not hold the galaxy", sc.Bounds)
	}
}
