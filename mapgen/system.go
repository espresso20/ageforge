package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The plan of the solar system: the lanes of the orbits, with parcels as marks
// along them in the order the people reached them. The system itself is drawn
// by the orrery (orrery.go).

var systemSpec = ringSpec{
	cx: SystemW / 2, cy: SystemH / 2, hub: 7, rings: []int{14, 22, 31, 41, 59, 68, 76, 83},
	lotW: 3, lotH: 1, step: 6, spread: [2]int{30, 10},
}

// homeRing is the orbit of the home world, the third from the sun.
const homeRing = 2

// NewSystemPlan lays out the lanes, a planet on each, the wonder's ground and
// the marks along the orbits.
func NewSystemPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newRingPlan(seed, systemSpec, System)
	p := b.plan
	fixed := make([]int, 0, len(p.Roads)-1)
	for i := 1; i < len(p.Roads); i++ {
		fixed = append(fixed, i)
	}
	p.Fixed = fixed
	for ring := range systemSpec.rings {
		cells := p.Roads[2+2*ring+roll(2, seed, 30, int64(ring))].Cells
		p.Bodies = append(p.Bodies, cells[len(cells)/8+roll(len(cells)*3/4, seed, 31, int64(ring))])
	}
	for _, body := range p.Bodies {
		b.take(Rect{body.X, body.Y, 1, 1})
	}
	b.plotBesideRing(18, len(systemSpec.rings)-1)
	b.cutRingLots(seed, cat, systemSpec)
	return p
}

// paintBodies draws what stands on the lanes of a space scale over them: the
// planets of the solar system, the systems of the star map.
func (f *frame) paintBodies() {
	switch f.st.Scale {
	case Orbit:
		f.paintOrbitStructures()
	case Stars:
		f.paintTerritory()
		n, glyph := f.plan.Centre, []rune(f.st.Hub[0])[0]
		f.g.draw(n.X, n.Y, glyph, f.ink(glyph, HNone))
		f.g.mark(n.X, n.Y, Thing{Kind: KindTerrain, Key: "home"})
	}
}
