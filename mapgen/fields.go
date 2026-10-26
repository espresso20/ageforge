package mapgen

import "slices"

// The world is a fixed grid of cells, the same for every terminal size, so a
// building never moves because the window did.
const (
	worldW = 240
	worldH = 96
)

// Biome is what covers the ground at a cell.
type Biome uint8

const (
	Deep Biome = iota
	Shallows
	Sand
	Scrub
	Grass
	Forest
	Hills
	Mountain
	River
	numBiomes
)

var biomeNames = [numBiomes]string{"deep water", "shallows", "sand", "scrub", "grass", "forest", "hills", "mountains", "river"}

// Water reports whether the biome is wet ground.
func (b Biome) Water() bool { return b <= Shallows || b == River }

// Height thresholds: the sea, its beach, and where the land turns to hills
// and mountains.
const (
	deepBelow = 270
	seaLevel  = 320
	beachTop  = 345
	hillsAt   = 520
	rangesAt  = 590
)

// biomeBands picks the biome from the height band (lowland, upland, peaks)
// and the moisture band (dry, mild, wet).
var biomeBands = [3][3]Biome{
	{Scrub, Grass, Forest},
	{Hills, Hills, Forest},
	{Mountain, Mountain, Hills},
}

// Fields is the land: height, moisture, biome, and the water's way down.
type Fields struct {
	Height, Moist []int
	// Ground is Height blurred, so the lie of the land (not each cell's
	// roughness) decides where building is too steep.
	Ground []int
	Biome  []Biome
	// WaterDist is the steps to the nearest water, capped at 255.
	WaterDist []uint8
}

func at(x, y int) int { return y*worldW + x }

func inWorld(x, y int) bool { return x >= 0 && y >= 0 && x < worldW && y < worldH }

// coasts are the eight directions the sea can lie in.
var coasts = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, 1}, {1, -1}, {-1, -1}}

// newFields grows the land for a seed: rolling noise tilted down toward one
// coast, rivers where the rain gathers, a biome for every cell.
func newFields(seed int64) *Fields {
	f := &Fields{
		Height:    make([]int, worldW*worldH),
		Moist:     make([]int, worldW*worldH),
		Ground:    make([]int, worldW*worldH),
		Biome:     make([]Biome, worldW*worldH),
		WaterDist: make([]uint8, worldW*worldH),
	}
	coast := coasts[roll(len(coasts), seed, 1)]
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			noise := fbm(seed, x*256/30, y*256/15, 4)
			reach := (x-worldW/2)*coast[0] + 2*(y-worldH/2)*coast[1]
			f.Height[at(x, y)] = clamp(noise*2/5+235-max(0, reach-46)*8, 0, noiseMax)
			f.Moist[at(x, y)] = fbm(seed+101, x*256/40, y*256/20, 4)
		}
	}
	f.blur()
	sea := make([]bool, len(f.Height))
	for i, h := range f.Height {
		sea[i] = h < seaLevel
	}
	order, parent := f.drain(sea)
	f.cutRivers(order, parent, sea)
	f.measureWater()
	f.settleBiomes()
	return f
}

// blur averages every cell with the 5 by 3 cells round it.
func (f *Fields) blur() {
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			sum, n := 0, 0
			for yy := y - 1; yy <= y+1; yy++ {
				for xx := x - 2; xx <= x+2; xx++ {
					if inWorld(xx, yy) {
						sum += f.Height[at(xx, yy)]
						n++
					}
				}
			}
			f.Ground[at(x, y)] = sum / n
		}
	}
}

// neighbours are the four steps, in a fixed order.
var neighbours = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// drain floods the land up from the sea, lowest ground first, so every cell
// learns the neighbour its water leaves by. A pit fills until it spills, so a
// river never dead-ends in a basin. It returns the cells in flood order and
// each cell's parent (-1 for the sea).
func (f *Fields) drain(sea []bool) (order, parent []int) {
	parent = make([]int, len(sea))
	seen := make([]bool, len(sea))
	levels := make([][]int, noiseMax+1)
	for i, wet := range sea {
		if wet {
			seen[i], parent[i] = true, -1
			levels[f.Height[i]] = append(levels[f.Height[i]], i)
		}
	}
	for level := range levels {
		for k := 0; k < len(levels[level]); k++ {
			c := levels[level][k]
			order = append(order, c)
			for _, d := range neighbours {
				x, y := c%worldW+d[0], c/worldW+d[1]
				if !inWorld(x, y) || seen[at(x, y)] {
					continue
				}
				n := at(x, y)
				seen[n], parent[n] = true, c
				up := max(f.Height[n], level)
				levels[up] = append(levels[up], n)
			}
		}
	}
	return order, parent
}

// cutRivers marks the land cells the most rain runs through. The rain of
// every cell flows down its parent chain, so the totals gather into streams
// that join; the busiest few percent of the land are rivers.
func (f *Fields) cutRivers(order, parent []int, sea []bool) {
	flow := make([]int, len(sea))
	for i := range flow {
		flow[i] = 1
	}
	for i := len(order) - 1; i >= 0; i-- {
		if p := parent[order[i]]; p >= 0 {
			flow[p] += flow[order[i]]
		}
	}
	var land []int
	for i, wet := range sea {
		if !wet {
			land = append(land, flow[i])
		}
	}
	slices.Sort(land)
	threshold := max(land[len(land)*97/100], 60)
	for i, wet := range sea {
		if !wet && flow[i] >= threshold {
			f.Biome[i] = River
		}
	}
}

// measureWater fills WaterDist by spreading outward from every wet cell.
func (f *Fields) measureWater() {
	var queue []int
	for i := range f.WaterDist {
		f.WaterDist[i] = 255
		if f.Height[i] < seaLevel || f.Biome[i] == River {
			f.WaterDist[i] = 0
			queue = append(queue, i)
		}
	}
	for head := 0; head < len(queue); head++ {
		c := queue[head]
		for _, d := range neighbours {
			x, y := c%worldW+d[0], c/worldW+d[1]
			if inWorld(x, y) && f.WaterDist[at(x, y)] == 255 {
				f.WaterDist[at(x, y)] = f.WaterDist[c] + 1
				queue = append(queue, at(x, y))
			}
		}
	}
}

// settleBiomes gives every dry cell its biome from height and moisture; land
// near water is damper.
func (f *Fields) settleBiomes() {
	for i, h := range f.Height {
		switch {
		case f.Biome[i] == River:
		case h < deepBelow:
			f.Biome[i] = Deep
		case h < seaLevel:
			f.Biome[i] = Shallows
		case h < beachTop:
			f.Biome[i] = Sand
		default:
			damp := f.Moist[i] + max(0, 6-int(f.WaterDist[i]))*25
			wet := clamp(damp*3/noiseMax, 0, 2)
			band := 0
			if h >= hillsAt {
				band = 1
			}
			if h >= rangesAt {
				band = 2
			}
			f.Biome[i] = biomeBands[band][wet]
		}
	}
}

// Land reports whether (x, y) is dry ground.
func (f *Fields) Land(x, y int) bool {
	return inWorld(x, y) && !f.Biome[at(x, y)].Water()
}

// Slope is the steepest drop to a neighbour over a two-cell stride, on the
// blurred ground, in height units.
func (f *Fields) Slope(x, y int) int {
	steep := 0
	for _, d := range neighbours {
		if inWorld(x+2*d[0], y+2*d[1]) {
			steep = max(steep, abs(f.Ground[at(x, y)]-f.Ground[at(x+2*d[0], y+2*d[1])]))
		}
	}
	return steep
}
