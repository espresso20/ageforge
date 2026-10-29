package mapgen

import (
	"cmp"

	"github.com/espresso20/ageforge/mapmodel"
)

// The star map is the last scale: a seeded spiral galaxy (galaxy.go) and, on
// it, the people's territory (territory.go): a sector of systems for every
// building line along the home arm, the clusters of the civilizations met on
// the other arms, the lanes that join their hubs and the rift a harbinger
// opens at the rim. The plan only fixes where things can stand; what is built
// decides how much of it lights.

// NewStarPlan lays out the galaxy, the place of every sector and rival and
// the rift. It holds no parcels and no wonder plots: counts, not ground, draw
// the star map.
func NewStarPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newPlanner(seed)
	p := b.plan
	p.Scale, p.Centre = Stars, Pt{worldW / 2, worldH / 2}
	p.Roads = []Road{{Cells: []Pt{p.Centre}, Parent: -1}}
	b.claimRoad(p.Centre, 0)
	p.Galaxy = NewGalaxy(seed)
	p.Plots = make([]Lot, 22)
	p.layTerritory(seed)
	return p
}

// staircase is the path from a to b, not counting a, one step at a time along
// whichever axis has further to go (a row counts double, being the taller).
func staircase(a, b Pt) []Pt {
	var cells []Pt
	for a != b {
		if abs(b.X-a.X) >= 2*abs(b.Y-a.Y) {
			a.X += cmp.Compare(b.X, a.X)
		} else {
			a.Y += cmp.Compare(b.Y, a.Y)
		}
		cells = append(cells, a)
	}
	return cells
}
