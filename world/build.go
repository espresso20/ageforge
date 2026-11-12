package world

import (
	"github.com/espresso20/ageforge/mapmodel"
)

// The ages that have a sky. Ground ages build an empty World.
const (
	AgeSpace        = "space_age"        // the home planet and its ring
	AgeInterstellar = "interstellar_age" // the solar system
	AgeGalactic     = "galactic_age"     // the star map
	AgeQuantum      = "quantum_age"
	AgeTranscendent = "transcendent_age"
)

// SpaceAges lists them in order.
var SpaceAges = []string{AgeSpace, AgeInterstellar, AgeGalactic, AgeQuantum, AgeTranscendent}

// ageOf is the age key of a model.
func ageOf(m *mapmodel.Model) string {
	if m.Catalog != nil && m.AgeIdx >= 0 && m.AgeIdx < len(m.Catalog.Ages) {
		return m.Catalog.Ages[m.AgeIdx]
	}
	return m.Age
}

// StarAge says whether the age shows the galaxy (the star map and its zoom
// ladder).
func StarAge(age string) bool {
	return age == AgeGalactic || age == AgeQuantum || age == AgeTranscendent
}

// Build lays out the sky of one save: seed is the save's seed (m.Seed), the
// snapshot says what the people have built. Ground ages have no sky and give
// a World with no Root. The same inputs give the same World.
func Build(m *mapmodel.Model, seed int64) *World {
	age := ageOf(m)
	w := &World{Seed: seed, Age: age, Key: ChangeKey(m, seed)}
	switch {
	case StarAge(age):
		g := NewGalaxy(seed)
		t := LayTerritory(seed, g)
		cols := Colonies(m, t)
		gal := galaxyFrame(m, seed, g, t)
		lg := localGroup(seed, g.Shape, solLy(g))
		region := regionFrame(seed, len(cols))
		near := neighbourhoodFrame(seed, cols)
		sys := systemFrame(m, seed, true)
		near.Children = []*Frame{sys}
		region.Children = []*Frame{near}
		gal.Children = []*Frame{region}
		lg.Children = []*Frame{gal}
		w.Root = lg
	case age == AgeSpace || age == AgeInterstellar:
		w.Root = systemFrame(m, seed, false)
	}
	return w
}

// ChangeKey hashes everything Build reads from the game state and the seed,
// and nothing else (not the tick), so a renderer rebuilds only when the sky
// would differ.
func ChangeKey(m *mapmodel.Model, seed int64) uint64 {
	age := ageOf(m)
	if !StarAge(age) && age != AgeSpace && age != AgeInterstellar {
		return mapmodel.Hash(seed, 0)
	}
	h := []int64{seed, mapmodel.HashStr(age), boolI(gateRaised(m.Wonders)), boolI(m.Harbinger != nil)}
	for _, l := range m.Lineages {
		if _, ok := lineIndex[l.Key]; !ok || l.Count <= 0 {
			continue
		}
		h = append(h, mapmodel.HashStr(l.Key), int64(l.Count), int64(tierOf(m, l)))
	}
	for i := range m.Factions {
		if c := &m.Factions[i]; c.Discovered {
			h = append(h, mapmodel.HashStr(c.Key), int64(c.Strength))
		}
	}
	return mapmodel.Hash(h...)
}

func boolI(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func gateRaised(ws []mapmodel.Wonder) bool {
	for _, w := range ws {
		if w.Key == GateWonder && w.Built {
			return true
		}
	}
	return false
}

// tierOf is 0..3: the rank of the best building of the line standing, spread
// over the line's kinds.
func tierOf(m *mapmodel.Model, l *mapmodel.Lineage) int {
	kinds := 0
	if m.Catalog != nil {
		kinds = len(m.Catalog.ByLineage[l.Key])
	}
	if kinds > 1 && len(l.Buildings) > 0 {
		return clamp(l.Buildings[len(l.Buildings)-1].Rank*3/(kinds-1), 0, 3)
	}
	return 0
}

// Colonies is the building lines with something standing, in lineage order,
// with the systems each lights in its sector.
func Colonies(m *mapmodel.Model, t Territory) []Colony {
	var out []Colony
	for _, l := range m.Lineages {
		i, ok := lineIndex[l.Key]
		if !ok || l.Count <= 0 || i >= len(t.Sectors) {
			continue
		}
		out = append(out, Colony{Line: l.Key, Count: l.Count, Systems: t.Sectors[i].Lit(l.Count), Tier: tierOf(m, l)})
	}
	return out
}
