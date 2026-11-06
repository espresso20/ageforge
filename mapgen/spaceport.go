package mapgen

import (
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The space port, the close-up of the Interstellar Age (the deepest zoom of
// the Space Age): a ring city in orbit
// round the home world, a megacity of trillions seen from outside, drawn in
// quadrant blocks (quad.go). Every building the player has is a module of the
// ring; ships dock at the dock arms and leave toward the edge of the picture.
// Everything about it (tilt, where Earth stands, the order of the modules)
// comes from the seed.

// Shapes of a module.
const (
	shBlock = iota // a hull with a lit top
	shSpire        // a mast
	shDish         // a bowl on a pole
	shDock         // an arm with a berth at each end
)

// module is what one building line looks like on the ring: its shape, size
// in pixels (across, up) and hull.
type module struct {
	shape, w, h int
	hull        Paint
}

// lineModule is the one table that turns buildings into ring modules. The key
// is the building's line; a bigger stock of a building grows its module.
var lineModule = map[string]module{
	mapmodel.LinHousing:   {shBlock, 5, 3, P(HSteel, 255)},
	mapmodel.LinFood:      {shBlock, 6, 2, P(HLife, 255)},
	mapmodel.LinWood:      {shBlock, 4, 2, P(HHydro, 255)},
	mapmodel.LinMines:     {shBlock, 4, 3, P(HRock, 255)},
	mapmodel.LinMetal:     {shBlock, 5, 3, P(HFire, 255)},
	mapmodel.LinEngineer:  {shSpire, 3, 5, P(HAmber, 255)},
	mapmodel.LinEnergy:    {shDish, 6, 2, P(HPanel, 255)},
	mapmodel.LinHarbor:    {shDock, 7, 2, P(HWhite, 255)},
	mapmodel.LinHacker:    {shDish, 3, 3, P(HNeon, 255)},
	mapmodel.LinKnowledge: {shSpire, 2, 6, P(HCyan, 255)},
	mapmodel.LinFaith:     {shSpire, 3, 7, P(HWhite, 255)},
	mapmodel.LinCulture:   {shDish, 5, 3, P(HGold, 255)},
	mapmodel.LinMonument:  {shSpire, 3, 8, P(HGold, 255)},
	mapmodel.LinTrade:     {shDock, 6, 2, P(HSand, 255)},
	mapmodel.LinStorage:   {shDock, 5, 2, P(HMetal, 255)},
	mapmodel.LinMilitary:  {shBlock, 6, 2, P(HRed, 255)},
	mapmodel.LinDiplomacy: {shDish, 4, 2, P(HGlass, 255)},
}

// gateBuilding is the wonder that raises the warp gate on the ring.
const gateBuilding = "warp_nexus"

// gateRaised says whether the warp gate stands: the Warp Nexus is built.
func gateRaised(ws []mapmodel.Wonder) bool {
	return slices.ContainsFunc(ws, func(w mapmodel.Wonder) bool { return w.Key == gateBuilding && w.Built })
}

// slots on the outer ring, and on the inner one that holds the overflow.
const outerSlots, innerSlots = 60, 42

// portMod is one module of the ring: a building of the player's, at pixel
// (x, y), on the near side of the ring or the far side.
type portMod struct {
	b     *mapmodel.Building
	m     module
	tier  int
	x, y  int
	front bool
}

// port is the laid-out ring city of a frame.
type port struct {
	cx, cy, R int // Earth's centre in pixels, and its radius in pixels
	rx, ry    int // the ring's radii in pixels
	rot       int // the tilt of the ring
	mods      []portMod
	gate      bool
	gateAt    int // the angle of the gate on the ring
}

// at is the point of the ring at angle a, k percent of its radius out, and
// whether it is on the near side.
func (pp *port) at(a, k int) (x, y int, front bool) {
	X, Y := pp.rx*k/100*cos4096(a)/1024, 2*pp.ry*k/100*sin4096(a)/1024
	c, s := cos4096(pp.rot), sin4096(pp.rot)
	yr := (X*s + Y*c) / 1024
	return pp.cx + (X*c-Y*s)/1024, pp.cy + yr/2, yr >= 0
}

// port lays the ring out from the seed and the model's buildings; it is the
// same on every frame.
func (f *frame) port() *port {
	if f.pp != nil {
		return f.pp
	}
	seed, c := f.m.Seed, f.plan.Centre
	pp := &port{cx: 2*c.X + 1 + roll(41, seed, 5300) - 20, cy: 2*c.Y + 1, R: 22 + roll(7, seed, 5301)}
	pp.rx = 66 + roll(20, seed, 5302)
	pp.ry = pp.rx / 2 * (45 + roll(36, seed, 5303)) / 100
	pp.rot = roll(600, seed, 5304) - 300
	f.pp = pp

	// the modules, one per building and then more of the larger stocks, until
	// the ring is full
	type want struct {
		b    *mapmodel.Building
		tier int
	}
	var wants []want
	for round := 0; round < 4; round++ {
		for _, b := range f.m.Buildings {
			parcels, tier := growthFor(b.Count)
			if _, ok := lineModule[b.Lineage]; ok && !b.Wonder && b.Count > 0 && round < parcels {
				wants = append(wants, want{b, tier})
			}
		}
	}
	pp.gate = false
	if gateRaised(f.m.Wonders) {
		pp.gate = true
		pp.gateAt = roll(outerSlots, seed, 5305)
	}
	// the slots in the seed's order: the outer ring, then the inner
	var slots []int
	for i := 0; i < outerSlots+innerSlots; i++ {
		if !pp.gate || i >= outerSlots || (i-pp.gateAt+outerSlots+1)%outerSlots > 2 {
			slots = append(slots, i)
		}
	}
	for _, half := range [][]int{slots[:len(slots)-innerSlots], slots[len(slots)-innerSlots:]} {
		slices.SortFunc(half, func(a, b int) int { return int(hash(seed, 5306, int64(a))%1024) - int(hash(seed, 5306, int64(b))%1024) })
	}
	// few buildings fill only the outer ring; more spill to the inner one
	for i, w := range wants {
		if i >= len(slots) {
			break
		}
		s, n, k := slots[i], outerSlots, 100
		if s >= outerSlots {
			s, n, k = s-outerSlots, innerSlots, 80
		}
		x, y, front := pp.at(s*turn/n, k)
		pp.mods = append(pp.mods, portMod{w.b, lineModule[w.b.Lineage], w.tier, x, y, front})
	}
	return pp
}

// paintSpacePort draws the ring city over the sky already in the frame.
func (f *frame) paintSpacePort() {
	seed, anim := f.m.Seed, f.anim
	pp := f.port()
	pix := NewPixels(f.g.W, f.g.H)
	shade := lighting(seed)
	earthX, earthY, earthR := pp.cx*pxUnit, pp.cy*pxUnit, pp.R*pxUnit

	ring := func(front bool) {
		n := 4 * (pp.rx + pp.ry)
		for _, k := range [2]int{100, 94} {
			for i := 0; i < n; i++ {
				if x, y, fr := pp.at(i*turn/n, k); fr == front {
					pix.Set(x, y, P(HSteel, 150+60*k/100), 0)
				}
			}
		}
		for i, m := range pp.mods {
			if m.front == front {
				own := pix.Owner(Thing{Kind: KindBuilding, Key: m.b.Key, Lineage: m.b.Lineage, Count: m.b.Count}, 5)
				pix.module(m, own)
				switch h := m.m.h + (m.tier+1)/2; m.m.shape { // a beacon on each spire, a light at each dock's tip
				case shSpire:
					pix.light(seed, anim, 10+i, m.x, m.y-h, HRed, own)
				case shDock:
					pix.light(seed, anim, 10+i, m.x-(m.m.w+(m.tier+1)/2)/2, m.y-1, HLife, own)
				}
			}
		}
	}

	// the Moon, out beyond the ring, in front of or behind the world
	ma := roll(turn, seed, 5307) + anim/8*16
	mx, my, moonFront := pp.at(ma, 135)
	moonR := (5 + roll(4, seed, 5308)) * pxUnit
	moon := func(front bool) {
		if moonFront == front {
			pix.moon(mx*pxUnit, my*pxUnit, moonR, 0, shade)
		}
	}

	moon(false)
	ring(false)
	pix.planet(seed, anim, &SolSpec.Planets[planetEarth], planetEarth, SolSpec.ThingOf(planetEarth), earthX, earthY, earthR, shade)
	ring(true)
	moon(true)

	// a swarm of small satellites winks in the dark round the ring
	for i := 0; i < 16; i++ {
		x, y, _ := pp.at(roll(turn, seed, 5310, int64(i))+anim/(20+i)*16, 112+roll(45, seed, 5311, int64(i)))
		pix.light(seed, anim, 200+i, x, y, HWhite, 0)
	}

	// the warp gate stands on the ring, a few lights round it
	if pp.gate {
		x, y, _ := pp.at(pp.gateAt*turn/outerSlots, 100)
		own := pix.Owner(Thing{Kind: KindBuilding, Key: "warp_gate"}, 7)
		pix.Ellipse(x*pxUnit, y*pxUnit, 5*pxUnit, 9*pxUnit, P(HCyan, 255), own, 0, 0)
		pix.Ellipse(x*pxUnit, y*pxUnit, 4*pxUnit, 7*pxUnit, P(HNeon, 220), own, 0, 0)
		pix.Set(x, y, P(HWhite, 255), own)
		pix.light(seed, anim, 1, x, y-9, HRed, own)
		pix.light(seed, anim, 2, x, y+9, HLife, own)
		pix.light(seed, anim, 3, x-5, y, HWhite, own)
		pix.light(seed, anim, 4, x+5, y, HWhite, own)
	}

	for _, m := range pp.mods {
		f.targets = append(f.targets, Pt{m.x / 2, m.y / 2})
	}
	pix.Resolve(f.g)
	pix.Release()
}

// module draws one module centred on its pixel.
func (p *Pixels) module(m portMod, own uint8) {
	w, h := m.m.w+(m.tier+1)/2, m.m.h+(m.tier+1)/2
	hull, trim := m.m.hull, P(lineHue(m.b.Lineage), 255)
	x, y := m.x, m.y
	switch m.m.shape {
	case shBlock:
		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				c := hull
				if dy == 0 {
					c = trim
				}
				p.Set(x-w/2+dx, y-h/2+dy, c, own)
			}
		}
	case shSpire:
		for dx := 0; dx < w; dx++ {
			p.Set(x-w/2+dx, y, hull, own)
		}
		for dy := 1; dy < h; dy++ {
			p.Set(x, y-dy, hull, own)
		}
		p.Set(x, y-h, trim, own)
	case shDish:
		for dx := 0; dx < w; dx++ {
			p.Set(x-w/2+dx, y, hull, own)
		}
		p.Set(x-w/2, y-1, hull, own)
		p.Set(x-w/2+w-1, y-1, hull, own)
		p.Set(x, y+1, trim, own)
	default:
		for dx := 0; dx < w; dx++ {
			p.Set(x-w/2+dx, y, hull, own)
		}
		for _, e := range [2]int{x - w/2, x - w/2 + w - 1} {
			p.Set(e, y-1, trim, own)
			p.Set(e, y+1, trim, own)
		}
	}
}
