package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Orbit is the first space scale: the home planet seen from straight above, a
// small disc in the middle with its land and sea drawn from the world, and
// rings of orbit round it with stations standing on them, joined by one space
// elevator. Everything else is the engine as it was: the same parcels, the
// same stable placement, the same roads, joins, traffic and light.

var orbitSpec = ringSpec{
	cx: worldW / 2, cy: worldH / 2, hub: 4, rings: []int{7, 10, 13, 16},
	lotW: 5, lotH: 3, step: 6, spread: [2]int{30, 10},
}

// the home planet's disc: its radii in cells
const discRX, discRY = 8, 4

// NewOrbitPlan lays out the elevator, the rings, the wonder's plot and the
// station parcels, and hands the parcels to the building lines.
func NewOrbitPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newRingPlan(seed, orbitSpec, Orbit)
	b.plan.Fixed = []int{1}
	b.plotBesideRing(17, len(orbitSpec.rings)-1)
	b.cutRingLots(seed, cat, orbitSpec)
	return b.plan
}

// rim is 1000 on the edge of the ellipse of radii rx, ry cells round
// (cx, cy), less inside it and more outside.
func rim(x, y, cx, cy, rx, ry int) int {
	dx, dy := x-cx, y-cy
	return (dx*dx*ry*ry + dy*dy*rx*rx) * 1000 / (rx * rx * ry * ry)
}

// paintHomePlanet draws the disc. The whole world is squeezed onto it, so
// the land is the land the settlement stands on, with its coast and its
// forests; cloud drifts over it; the limb and the night side are dark, and on
// the night side the land shows the lights of its cities.
func (f *frame) paintHomePlanet() {
	c, seed := f.plan.Centre, f.m.Seed
	for y := c.Y - discRY; y <= c.Y+discRY; y++ {
		for x := c.X - discRX; x <= c.X+discRX; x++ {
			r := rim(x, y, c.X, c.Y, discRX, discRY)
			if r > 1000 {
				continue
			}
			wx := worldW/2 + (x-c.X)*(worldW/2)/discRX
			wy := worldH/2 + (y-c.Y)*(worldH/2)/discRY
			biome := f.fields.Biome[at(clamp(wx, 0, worldW-1), clamp(wy, 0, worldH-1))]
			cell := f.st.Land[biome].at(x, y, 0)
			if fbm(seed+9, x*256/6+f.anim*5, y*256/3, 3) > 620 {
				cell = f.st.Cloud.at(x, y, 0)
			}
			night := (x - c.X + discRX) * 150 / (2 * discRX) // the light comes from the left
			limb := clamp((r-700)*256/300, 0, 200)
			cell.Dim = dimmer(uint8(clamp(night, 0, 255)), uint8(limb))
			city := !biome.Water() && cell.Dim > 90 && chance(9, seed, 16, int64(x), int64(y))
			if city {
				cell.R = '∙'
				f.lights = append(f.lights, Pt{x, y})
			}
			f.g.paint(x, y, cell)
			f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "planet"})
		}
	}
}

// relicSpot is where the wonder of a ground age shines on the home planet.
func (f *frame) orbitRelic(age, try int) Pt {
	c := f.plan.Centre
	return Pt{c.X - discRX + 1 + roll(2*discRX-1, f.m.Seed, 17, int64(age), int64(try)),
		c.Y - discRY + 1 + roll(2*discRY-1, f.m.Seed, 18, int64(age), int64(try))}
}
