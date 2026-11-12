package mapview

import (
	"image/color"
	"math"

	"github.com/espresso20/ageforge/world"
)

// The Planet band is the home world's close-up: the planet as a lit sphere,
// the Moon on a dim orbit and the people's stations, swarm and ships as
// small sprites on low orbits. The frame draws its unit larger than true (see
// world.HomeInflate) and bends distance (squeeze), so the planet is a body to
// look at while the Moon is still in view, and nothing of the traffic lands
// on its disc.

// The art pixels per planet radius the zoom can reach in this band: the cache
// warms the sprite sizes in between.
const homePPUMin, homePPUMax = 10, 125

// homeDir is the light's bucket for the whole close-up: the star is where the
// frame's planet's own orbit points the other way, so the sphere is lit from the same
// side as the system band draws it.
func homeDir(f *world.Frame) int {
	ph := 0
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindPlanet && o.Body != nil {
			ph = o.Body.Phase
			break
		}
	}
	return ((ph+world.Turn/lightBuckets/2)/(world.Turn/lightBuckets) + lightBuckets/2) % lightBuckets
}

// homeScale is the whole-number size of the stations: they swell a little as
// the planet fills the view.
func homeScale(ppu float64) int { return 1 + min(2, int(ppu/30)) }

// A craft is a station's shape: a bitmap ('#' hull, '+' panel, 'o' a light
// that blinks) or a ring of radius ring pixels, with spokes or panels.
type craft struct {
	art           []string
	ring          int
	spokes, solar bool
}

var (
	stationArt = [4]craft{
		{ring: 4, spokes: true},
		{art: []string{"+++.###.+++", "+++-#o#-+++", "+++.###.+++"}},
		{art: []string{"...#...", "..###..", ".#####.", "o#####o", ".#####.", "..###..", "...#..."}},
		{art: []string{"...#o#...", "+++###+++", "+++#o#+++", "...###...", "...#o#..."}},
	}
	megaArt   = [3]craft{{ring: 6}, {ring: 6, spokes: true}, {ring: 6, solar: true}}
	moduleArt = craft{art: []string{"#####", "#o.o#", "#####"}}
	swarmArt  = craft{art: []string{"+#+"}}
	shipArt   = craft{art: []string{"..###..", "#######", "o.###.."}}
)

func craftOf(o *world.Object) craft {
	n := max(0, o.Count)
	switch {
	case o.Kind == world.KindShip:
		return shipArt
	case o.Body == nil:
		return stationArt[0]
	case o.Body.Kind == "mega":
		return megaArt[n%len(megaArt)]
	case o.Body.Kind == "module":
		return moduleArt
	case o.Body.Kind == "satellite":
		return swarmArt
	}
	return stationArt[n%len(stationArt)]
}

// craftR is how far a craft draws from its centre at scale s, in art pixels.
func craftR(c craft, s int) float64 {
	if c.ring > 0 {
		if c.solar {
			return float64((c.ring + 3) * s)
		}
		return float64(c.ring*s + s)
	}
	return float64(max(len(c.art[0]), len(c.art))/2*s + s)
}

// craftDraw paints craft c at x, y at scale s; on says whether its lights are lit.
func (p *pen) craftDraw(c craft, x, y float64, s int, w uint8, on bool) {
	hull, dim := p.pal.Mix(world.ToneWhite, w), p.pal.Mix(world.ToneWhite, w/2)
	panel := p.pal.Mix(world.ToneCyan, w)
	soft := p.pal.Mix(world.ToneWhite, w*7/10)
	lit := dim
	if on {
		lit = p.pal.Mix(world.ToneSun, 255)
	}
	xi, yi := int(math.Round(x)), int(math.Round(y))
	box := func(cx, cy int, col color.RGBA) {
		for dy := 0; dy < s; dy++ {
			for dx := 0; dx < s; dx++ {
				p.px(xi+cx*s+dx-s/2, yi+cy*s+dy-s/2, col)
			}
		}
	}
	if c.ring > 0 {
		R := float64(c.ring)
		for cy := -c.ring - 1; cy <= c.ring+1; cy++ {
			for cx := -c.ring - 1; cx <= c.ring+1; cx++ {
				d := math.Hypot(float64(cx), float64(cy))
				switch {
				case d >= R-.45 && d <= R+.45:
					box(cx, cy, soft)
				case d <= 1.2, c.spokes && (cx == 0 || cy == 0) && d < R:
					box(cx, cy, dim)
				}
			}
		}
		if c.solar {
			for cx := -c.ring - 3; cx <= c.ring+3; cx++ {
				if abs(cx) > c.ring+1 {
					box(cx, 0, panel)
					box(cx, 1, panel)
				}
			}
		}
		box(0, -c.ring, lit)
		return
	}
	h, wd := len(c.art), len(c.art[0])
	for r, row := range c.art {
		for q, ch := range row {
			var col color.RGBA
			switch ch {
			case '#':
				col = hull
			case '-':
				col = dim
			case '+':
				col = panel
			case 'o':
				col = lit
			default:
				continue
			}
			box(q-wd/2, r-h/2, col)
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// homeObject draws one thing of the Planet band at x, y; r is its radius in
// art pixels for the spheres.
func (p *pen) homeObject(o *world.Object, x, y, r float64) {
	switch o.Kind {
	case world.KindPlanet, world.KindMoon:
		if o.Body != nil {
			p.stamp(p.bod.get(p.pal, o, stepR(r), homeDir(p.f)), int(math.Round(x)), int(math.Round(y)))
		}
	case world.KindStation, world.KindShip:
		on := (p.tick+int(hashKey(o.Key)&1))&1 == 0
		p.craftDraw(craftOf(o), x, y, homeScale(p.ppu), o.Weight, on)
	}
}

// homeGuides draws the orbits under the traffic: the Moon's, a little
// brighter, and the stations' low ones, barely there.
func (p *pen) homeGuides() {
	for i := range p.f.Objects {
		if o := &p.f.Objects[i]; o.Body != nil && o.Body.Orbit > 0 {
			w := uint8(40)
			if o.Kind == world.KindMoon {
				w = 70
			}
			p.orbit(squeeze(p.f, o.Body.Orbit)*p.ppu, p.pal.Mix(world.ToneStar, w))
		}
	}
}
