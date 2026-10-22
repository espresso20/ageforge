package ui

import (
	"hash/fnv"
	"math"
	"strconv"

	"github.com/espresso20/ageforge/game"
)

// expedition_map.go is the Expeditions panel's map: terrain drawn from a
// seeded noise of the run's seed and the age, and the pins an offer or a
// mission sets on it. Nothing here is random: the same run, age, size and
// board give the same map every time it is drawn.

// mix64 is a 64-bit mixer (the splitmix64 finaliser).
func mix64(v uint64) uint64 {
	v += 0x9e3779b97f4a7c15
	v = (v ^ (v >> 30)) * 0xbf58476d1ce4e5b9
	v = (v ^ (v >> 27)) * 0x94d049bb133111eb
	return v ^ (v >> 31)
}

// unit is the low 16 bits of v as a number from 0 to 1.
func unit(v uint64) float64 { return float64(v&0xffff) / 65535 }

// pinHash is the stable number an offer's place on the map comes from: its
// title and the tick it expires at, neither of which changes while it is
// on the board or out on its mission.
func pinHash(title string, expiry int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(title))
	h.Write([]byte{'|'})
	h.Write([]byte(strconv.Itoa(expiry)))
	return mix64(h.Sum64())
}

// terrainSeed is the number a map's land comes from: the run's seed and
// the age, so it holds while the age lasts and changes with the next.
func terrainSeed(seed int64, age string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(age))
	h.Write([]byte{'|'})
	h.Write([]byte(strconv.FormatInt(seed, 10)))
	return mix64(h.Sum64())
}

func expNoise(s uint64, x, y int) float64 {
	v := mix64(s ^ uint64(int64(x))*0x9e3779b97f4a7c15 ^ uint64(int64(y))*0xc2b2ae3d27d4eb4f)
	return float64(v>>11) / float64(1<<53)
}

// expTerrain is the land of a map w by h cells.
type expTerrain struct {
	w, h  int
	r     []rune
	k     []xk
	water []bool
}

// buildTerrain draws the land: a coast on one side (which side comes from
// the seed), hills and mountains where one noise runs high, meadow where it
// runs low.
func buildTerrain(seed int64, age string, w, h int) *expTerrain {
	s := terrainSeed(seed, age)
	t := &expTerrain{w: w, h: h, r: make([]rune, w*h), k: make([]xk, w*h), water: make([]bool, w*h)}
	left := (s>>7)&1 == 1
	block := func(x, y int) float64 { return expNoise(s, x/4, y/3) }
	for y := 0; y < h; y++ {
		coast := int(float64(w)*0.80) + int(math.Round(math.Sin(float64(y)*.7)*2+expNoise(s, 0, y)*2))
		for x := 0; x < w; x++ {
			i := y*w + x
			t.r[i], t.k[i] = ' ', xFaint
			edge := x
			if left {
				edge = w - 1 - x
			}
			if edge > coast {
				t.r[i], t.k[i], t.water[i] = '≈', xWater, true
				continue
			}
			v := block(x, y)*.6 + block(x+2, y+1)*.4
			r := expNoise(s, x+1000, y+1000)
			switch {
			case v > .8 && r > .35:
				t.r[i], t.k[i] = '∩', xHill
				if r > .6 {
					t.r[i] = '^'
				}
			case v < .2 && r > .3:
				t.r[i], t.k[i] = '\'', xGrass
				if r > .6 {
					t.r[i] = '"'
				}
			case r > .9:
				t.r[i] = '.'
			}
		}
	}
	return t
}

// isWater reports whether the cell is sea (cells off the map count as sea).
func (t *expTerrain) isWater(x, y int) bool {
	return x < 0 || y < 0 || x >= t.w || y >= t.h || t.water[y*t.w+x]
}

// pt is a cell of the map.
type pt struct{ x, y int }

// townAt is where the town stands on a map w by h.
func townAt(w, h int) pt { return pt{w / 2, h / 2} }

// tierBands is how far from the town a tier's pins stand, as a share of
// the way to the map's edge: short near, long in the middle, grand far.
var tierBands = map[string][2]float64{
	game.TierShort: {0.20, 0.42},
	game.TierLong:  {0.44, 0.68},
	game.TierGrand: {0.70, 0.98},
}

// pinSpec is what the placement needs to know of a pin.
type pinSpec struct {
	hash uint64
	tier string
}

// placePins sets every spec on the map and returns its cell, in order. A
// pin stands on land, not on the town and not on or beside another pin;
// its first choice comes from its hash and the tier it is, and a pin that
// finds the cell taken tries the next one its hash points to. ok[i] is
// false for a pin that could find no room, on a map too small for them.
func placePins(t *expTerrain, specs []pinSpec) (at []pt, ok []bool) {
	town := townAt(t.w, t.h)
	at, ok = make([]pt, len(specs)), make([]bool, len(specs))
	taken := func(p pt, upto int) bool {
		if t.isWater(p.x, p.y) || p.x < 1 || p.x > t.w-2 {
			return true
		}
		if abs(p.x-town.x)+2*abs(p.y-town.y) < 4 {
			return true
		}
		for j := 0; j < upto; j++ {
			if ok[j] && at[j].y == p.y && abs(at[j].x-p.x) <= 3 {
				return true
			}
		}
		return false
	}
	maxRx, maxRy := float64(town.x-1), float64(town.y)
	for i, sp := range specs {
		band := tierBands[sp.tier]
		if band == [2]float64{} {
			band = tierBands[game.TierShort]
		}
		for attempt := 0; attempt < 600 && !ok[i]; attempt++ {
			h := sp.hash
			if attempt > 0 {
				h = mix64(sp.hash + uint64(attempt)*0x9e3779b97f4a7c15)
			}
			lo, hi := band[0], band[1]
			if attempt > 40 { // the band is full: look further afield
				lo, hi = 0.15, 1.0
			}
			f := lo + (hi-lo)*unit(h>>16)
			th := 2 * math.Pi * unit(h>>32)
			p := pt{town.x + int(math.Round(f*maxRx*math.Cos(th))), town.y + int(math.Round(f*maxRy*math.Sin(th)))}
			p.x, p.y = clamp(p.x, 1, t.w-2), clamp(p.y, 0, t.h-1)
			if !taken(p, i) {
				at[i], ok[i] = p, true
			}
		}
		for y := 0; y < t.h && !ok[i]; y++ { // the last resort: the first free cell
			for x := 1; x < t.w-1 && !ok[i]; x++ {
				if p := (pt{x, y}); !taken(p, i) {
					at[i], ok[i] = p, true
				}
			}
		}
	}
	return at, ok
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi int) int { return max(lo, min(hi, v)) }

// routeCells is the cells of a straight route from a to b, ends included
// (Bresenham's line).
func routeCells(a, b pt) []pt {
	var out []pt
	dx, dy := abs(b.x-a.x), -abs(b.y-a.y)
	sx, sy := 1, 1
	if a.x > b.x {
		sx = -1
	}
	if a.y > b.y {
		sy = -1
	}
	e := dx + dy
	for {
		out = append(out, a)
		if a == b {
			return out
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			a.x += sx
		}
		if e2 <= dx {
			e += dx
			a.y += sy
		}
	}
}

// hopTo picks the pin to move to from pin cur in direction (dx, dy): the
// nearest one lying that way, counting a step down as two across because
// a cell is twice as tall as it is wide, and a pin off the line as worse.
// -1 when nothing lies that way.
func hopTo(pins []pt, cur int, dx, dy int) int {
	if cur < 0 || cur >= len(pins) {
		return -1
	}
	p := pins[cur]
	best, bestScore := -1, math.MaxInt
	for i, q := range pins {
		if i == cur {
			continue
		}
		ddx, ddy := q.x-p.x, (q.y-p.y)*2
		along := ddx*dx + ddy*dy
		if along <= 0 {
			continue
		}
		across := abs(ddx*dy) + abs(ddy*dx)
		if score := along + across*2; score < bestScore {
			best, bestScore = i, score
		}
	}
	return best
}
