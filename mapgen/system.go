package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The solar system is the next scale out: the sun in the middle, the planets
// on their orbits, the lanes between them. Parcels are marks along the
// orbits, in the order the people reached them; the planets stand on the
// orbits among them.

var systemSpec = ringSpec{
	cx: worldW / 2, cy: worldH / 2, hub: 2, rings: []int{4, 6, 8, 10, 12, 14, 16, 18},
	lotW: 1, lotH: 1, step: 3, spread: [2]int{26, 9},
}

// homeRing is the orbit of the home world, the third from the sun.
const homeRing = 2

// NewSystemPlan lays out the lanes, a planet on each, the wonder's ground and
// the marks along the orbits.
func NewSystemPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newRingPlan(seed, systemSpec, System)
	p := b.plan
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

// paintSystem draws the sun.
func (f *frame) paintSystem() {
	c := f.plan.Centre
	w, h := f.st.Hub.size()
	sun := Thing{Kind: KindTerrain, Key: "sun"}
	for dy, row := range f.st.Hub {
		for dx, r := range []rune(row) {
			if r != ' ' {
				f.g.draw(c.X-w/2+dx, c.Y-h/2+dy, r, f.ink(r, HNone))
				f.g.mark(c.X-w/2+dx, c.Y-h/2+dy, sun)
			}
		}
	}
}

// paintBodies draws what stands on the lanes of a space scale over them: the
// planets of the solar system, the systems of the star map.
func (f *frame) paintBodies() {
	switch f.st.Scale {
	case System:
		for i, b := range f.plan.Bodies {
			if i < len(f.st.Bodies) {
				f.g.draw(b.X, b.Y, f.st.Bodies[i].Glyph, P(f.st.Bodies[i].Hue, 255))
				key := "planet"
				if i == homeRing {
					key = "home"
				}
				f.g.mark(b.X, b.Y, Thing{Kind: KindTerrain, Key: key})
			}
		}
	case Stars:
		for i, n := range f.plan.Nodes {
			glyph, ink, key := '✦', P(HStar, 235), "star"
			if i == 0 {
				glyph, ink, key = []rune(f.st.Hub[0])[0], f.ink([]rune(f.st.Hub[0])[0], HNone), "home"
			}
			f.g.draw(n.X, n.Y, glyph, ink)
			f.g.mark(n.X, n.Y, Thing{Kind: KindTerrain, Key: key})
		}
	}
}
