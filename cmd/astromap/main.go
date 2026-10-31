package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"
)

const W, H = 120, 32

// Cell represents a grid cell.
type Cell struct {
	R     rune
	Color int
	Bold  bool
	Prio  int
	Kind  string
}

// Map is the generated map.
type Map struct {
	Grid      [H][W]Cell
	RQ        int // rift quadrant
	SX        int // central star x
	SY        int // central star y
	GX        int // gas giant x
	GY        int // gas giant y
	Seed      int64
	Stations  []struct{ x, y int }
	LaneCells map[string]bool // for tracking lane cells
}

// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// hash combines x, y and salt to create a pseudorandom value.
func hash(x, y int, salt uint64) uint64 {
	h := uint64(x)*0x9E3779B97F4A7C15 ^ uint64(y)*0xC2B2AE3D27D4EB4F
	return h
}

// splitmix64 finalizer for hash output.
func splitmix64(z uint64) uint64 {
	z ^= z >> 30
	z *= 0xBF58476D1CE4E5B9
	z ^= z >> 27
	z *= 0x94D049BB133111EB
	z ^= z >> 31
	return z
}

// hash01 converts hash output to float64 in [0,1).
func hash01(x, y int, salt uint64) float64 {
	h := splitmix64(hash(x, y, salt))
	return float64(h>>11) / (1 << 53)
}

// set places a cell if prio is higher than existing.
func (m *Map) set(x, y int, r rune, color int, bold bool, prio int, kind string) {
	if x < 0 || x >= W || y < 0 || y >= H {
		return
	}
	if prio < m.Grid[y][x].Prio {
		return
	}
	m.Grid[y][x] = Cell{R: r, Color: color, Bold: bold, Prio: prio, Kind: kind}
}

// fbm computes fractional Brownian motion noise.
func fbm(x, y float64, seed int64, salt uint64) float64 {
	fx := x / 16
	fy := y / 8

	var result, amplitude float64 = 0, 1
	var maxAmp float64 = 0

	for octave := 0; octave < 4; octave++ {
		freq := math.Pow(2, float64(octave))
		sx := fx * freq
		sy := fy * freq

		ix := int(math.Floor(sx))
		iy := int(math.Floor(sy))

		tx := sx - float64(ix)
		ty := sy - float64(iy)

		// smoothstep
		tx = tx * tx * (3 - 2*tx)
		ty = ty * ty * (3 - 2*ty)

		// bilinear interpolation
		v00 := hash01(ix, iy, salt+uint64(octave))
		v10 := hash01(ix+1, iy, salt+uint64(octave))
		v01 := hash01(ix, iy+1, salt+uint64(octave))
		v11 := hash01(ix+1, iy+1, salt+uint64(octave))

		v0 := v00*(1-tx) + v10*tx
		v1 := v01*(1-tx) + v11*tx
		v := v0*(1-ty) + v1*ty

		result += v * amplitude
		maxAmp += amplitude
		amplitude *= 0.5
	}

	return result / maxAmp
}

// generate creates a new map with the given seed.
func generate(seed int64) *Map {
	m := &Map{
		Seed:      seed,
		LaneCells: make(map[string]bool),
	}
	rng := rand.New(rand.NewPCG(uint64(seed), 0x9E3779B97F4A7C15))

	// Initialize grid
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			m.Grid[y][x] = Cell{R: ' ', Color: 0, Bold: false, Prio: 0, Kind: "Empty space"}
		}
	}

	// Layer 1: cosmic strata
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			g := fbm(float64(x), float64(y), seed, 100)
			h := fbm(float64(x), float64(y), seed, 200)
			d := fbm(float64(x), float64(y), seed, 300)

			// Hue from h
			var hue int
			if h < 0.40 {
				hue = 37 // cyan
				if g > 0.70 {
					hue = 51
				} else if g > 0.62 {
					hue = 44
				}
			} else if h < 0.60 {
				hue = 25 // blue
				if g > 0.70 {
					hue = 33
				} else if g > 0.62 {
					hue = 27
				}
			} else {
				hue = 90 // magenta
				if g > 0.70 {
					hue = 165
				} else if g > 0.62 {
					hue = 127
				}
			}

			// Nebula glyphs
			if g > 0.70 {
				m.set(x, y, '▓', hue, false, 1, "Nebula (cyan|blue|magenta)")
			} else if g > 0.62 {
				m.set(x, y, '▒', hue, false, 1, "Nebula (cyan|blue|magenta)")
			} else if g > 0.54 {
				m.set(x, y, '░', hue, false, 1, "Nebula (cyan|blue|magenta)")
			} else if g > 0.48 && hash01(x, y, 101) < 0.35 {
				m.set(x, y, '~', 238, false, 1, "Nebula wisp")
			}

			// Dust lane
			if math.Abs(d-0.5) < 0.03 {
				m.set(x, y, '·', 236, false, 2, "Dust lane")
			}
		}
	}

	// Layer 2: stellar field
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			c := &m.Grid[y][x]
			if c.Prio > 0 && c.R != '░' && c.R != '~' {
				continue // skip non-empty or dust cells
			}
			if c.Kind == "Dust lane" {
				continue // skip dust
			}

			r := hash01(x, y, 400)
			if r < 0.030 {
				m.set(x, y, '.', 240, false, 3, "Faint star")
			} else if r < 0.045 {
				m.set(x, y, '·', 244, false, 3, "Faint star")
			} else if r < 0.053 {
				glyph := '*'
				if splitmix64(hash(x, y, 0))%2 == 0 {
					glyph = '∙'
				}
				m.set(x, y, glyph, 250, false, 3, "Star")
			} else if r < 0.057 {
				glyph := '✦'
				if splitmix64(hash(x, y, 0))%2 == 0 {
					glyph = '✧'
				}
				m.set(x, y, glyph, 231, true, 3, "Bright star")
			}
		}
	}

	// Layer 3: celestial bodies
	m.RQ = rng.IntN(4)

	// Central star
	m.SX = 60 + rng.IntN(31) - 15
	m.SY = 16 + rng.IntN(7) - 3

	// Helper: aspect distance
	ad := func(dx, dy int) float64 {
		return math.Sqrt(math.Pow(float64(dx)/2, 2) + math.Pow(float64(dy), 2))
	}

	// Place central star and corona
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			dx, dy := x-m.SX, y-m.SY
			e := ad(dx, dy)

			if e < 0.5 {
				m.set(x, y, '☼', 226, true, 9, "Central star")
			} else if e < 1.6 {
				m.set(x, y, '●', 220, true, 9, "Central star")
			} else if e < 2.6 {
				m.set(x, y, '▓', 214, false, 9, "Corona")
			} else if e < 3.6 {
				m.set(x, y, '▒', 208, false, 9, "Corona")
			} else if e < 4.6 && hash01(x, y, 500) < 0.6 {
				m.set(x, y, '░', 202, false, 9, "Corona")
			}
		}
	}

	// Asteroid belt
	a := 24 + rng.IntN(10)
	b := float64(a) / 2.4

	numSteps := int(math.Ceil(2*math.Pi/0.004)) + 1
	for step := 0; step < numSteps; step++ {
		t := 0.004 * float64(step)
		jitter := (hash01(step, 0, 600) - 0.5) * 2.0

		x := int(math.Round(float64(m.SX) + (float64(a)+jitter)*math.Cos(t)))
		y := int(math.Round(float64(m.SY) + (b+jitter/2)*math.Sin(t)))

		if x >= 0 && x < W && y >= 0 && y < H {
			if hash01(x, y, 601) < 0.30 {
				r := hash01(x, y, 602)
				glyph := '•'
				if r < 0.8 {
					glyph = '°'
				} else {
					glyph = '*'
				}

				color := 137
				if splitmix64(hash(x, y, 0))%2 == 0 {
					color = 180
				}
				m.set(x, y, glyph, color, false, 4, "Asteroid belt")
			}
		}
	}

	// Gas giant
	quadrants := [4][2]int{{30, 8}, {90, 8}, {30, 24}, {90, 24}}
	var qx, qy int
	var gq int
	for i := 0; i < 4; i++ {
		if i != m.RQ {
			gq = i
			break
		}
	}
	qx = quadrants[gq][0]
	qy = quadrants[gq][1]

	gx := qx + rng.IntN(17) - 8
	gy := qy + rng.IntN(5) - 2
	m.GX = gx
	m.GY = gy

	// Push away if too close to central star
	if ad(gx-m.SX, gy-m.SY) < 30 {
		for ad(gx-m.SX, gy-m.SY) >= 30 {
			if gx < m.SX {
				gx--
			} else {
				gx++
			}
		}
	}
	if gx < 12 {
		gx = 12
	}
	if gx > 107 {
		gx = 107
	}
	if gy < 4 {
		gy = 4
	}
	if gy > 27 {
		gy = 27
	}

	// Gas giant body
	rx, ry := 5, 2.5
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			dx, dy := float64(x-gx), float64(y-gy)
			if (dx/float64(rx))*(dx/float64(rx))+(dy/ry)*(dy/ry) <= 1 {
				glyph := '▓'
				if (y-gy)%2 == 1 {
					glyph = '▒'
				}
				color := 172
				if (y-gy)%2 == 1 {
					color = 136
				}
				m.set(x, y, glyph, color, false, 8, "Gas giant")
			}
		}
	}

	// Planetary ring
	rx2, ry2 := 11.0, 1.6
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			dx, dy := float64(x-gx), float64(y-gy)
			e2 := (dx/rx2)*(dx/rx2) + (dy/ry2)*(dy/ry2)

			// Check if inside body
			inBody := (dx/float64(rx))*(dx/float64(rx))+(dy/ry)*(dy/ry) <= 1

			if e2 >= 0.75 && e2 <= 1.25 && !inBody {
				glyph := '─'
				if math.Abs(dy) == 0 {
					glyph = '─'
				} else if dx*dy > 0 {
					glyph = '╱'
				} else {
					glyph = '╲'
				}
				m.set(x, y, glyph, 223, false, 7, "Planetary ring")
			}
		}
	}

	// Layer 5: harbinger rift
	var rfQx, rfQy, rfQw, rfQh int
	switch m.RQ {
	case 0: // NW
		rfQx, rfQy, rfQw, rfQh = 0, 0, 60, 16
	case 1: // NE
		rfQx, rfQy, rfQw, rfQh = 60, 0, 60, 16
	case 2: // SW
		rfQx, rfQy, rfQw, rfQh = 0, 16, 60, 16
	case 3: // SE
		rfQx, rfQy, rfQw, rfQh = 60, 16, 60, 16
	}

	x0 := rfQx + 10 + rng.IntN(40)
	x := x0
	var riftCells []struct{ x, y int }

	for step := 0; step < rfQh; step++ {
		y := rfQy + step

		dx := rng.IntN(5) - 2

		// Bias toward middle
		if x-rfQx < 6 && dx < 0 {
			dx = 1
		}
		if x-rfQx > rfQw-6 && dx > 0 {
			dx = -1
		}

		x += dx

		// Core
		glyph := '█'
		if dx > 0 {
			glyph = '╲'
		} else if dx < 0 {
			glyph = '╱'
		}

		if x >= 0 && x < W {
			m.set(x, y, glyph, 196, true, 10, "Harbinger rift")
			riftCells = append(riftCells, struct{ x, y int }{x, y})
		}

		// Intermediate
		if math.Abs(float64(dx)) == 2 {
			ix := x - dx/2
			if ix >= 0 && ix < W {
				m.set(ix, y, glyph, 196, true, 10, "Harbinger rift")
				riftCells = append(riftCells, struct{ x, y int }{ix, y})
			}
		}

		// Halo
		if x-1 >= 0 && x-1 < W {
			if hash01(x-1, y, 701) < 0.6 {
				m.set(x-1, y, '▓', 129, false, 10, "Harbinger rift")
			} else {
				m.set(x-1, y, '▒', 93, false, 10, "Harbinger rift")
			}
		}
		if x+1 >= 0 && x+1 < W {
			if hash01(x+1, y, 701) < 0.6 {
				m.set(x+1, y, '▓', 129, false, 10, "Harbinger rift")
			} else {
				m.set(x+1, y, '▒', 93, false, 10, "Harbinger rift")
			}
		}
	}

	// Sparks around rift cells - only check cells near rift
	for _, rc := range riftCells {
		// Check cells in a box around this rift cell
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				sx, sy := rc.x+dx, rc.y+dy
				if sx >= 0 && sx < W && sy >= 0 && sy < H {
					if ad(sx-rc.x, sy-rc.y) <= 3 && hash01(sx, sy, 702) < 0.10 {
						m.set(sx, sy, 'ϟ', 201, true, 10, "Rift spark")
					}
				}
			}
		}
	}

	// Layer 4: civilisation
	// Stations
	n := 5 + rng.IntN(4)
	type Station struct {
		x, y      int
		isFoundry bool
	}
	var stations []Station
	stationAttempts := 0
	for len(stations) < n && stationAttempts < 20000 {
		sx := 2 + rng.IntN(113)
		sy := 1 + rng.IntN(30)

		valid := true
		for _, st := range stations {
			if ad(sx-st.x, sy-st.y) < 14 {
				valid = false
				break
			}
		}

		if valid {
			for dy := 0; dy < 3; dy++ {
				for dx := -1; dx <= 1; dx++ {
					cx, cy := sx+dx, sy+dy
					if cx >= 0 && cx < W && cy >= 0 && cy < H {
						if m.Grid[cy][cx].Prio >= 7 {
							valid = false
							break
						}
					}
				}
				if !valid {
					break
				}
			}
		}

		if valid {
			// Check if in rift quadrant
			if sx >= rfQx && sx < rfQx+rfQw && sy >= rfQy && sy < rfQy+rfQh {
				valid = false
			}
		}

		if valid {
			stations = append(stations, Station{x: sx, y: sy, isFoundry: rng.IntN(2) == 0})
		}
		stationAttempts++
	}

	m.Stations = make([]struct{ x, y int }, len(stations))
	for i, st := range stations {
		m.Stations[i] = struct{ x, y int }{st.x, st.y}
	}

	// Place station tokens
	for _, st := range stations {
		left := '['
		center := '◈'
		right := ']'
		color := 51

		if st.isFoundry {
			center = '⬡'
			color = 214
		}

		m.set(st.x-1, st.y, left, 250, false, 11, "Station: habitat")
		m.set(st.x, st.y, center, color, true, 11, "Station: habitat")
		if st.isFoundry {
			m.Grid[st.y][st.x].Kind = "Station: foundry"
		}
		m.set(st.x+1, st.y, right, 250, false, 11, "Station: habitat")
	}

	// Jump gate
	gx_gate := 2 + rng.IntN(13)
	if rng.IntN(2) == 0 {
		gx_gate = 103 + rng.IntN(12)
	}

	gateFound := false
	var gy_gate int
	for tries := 0; tries < 32 && !gateFound; tries++ {
		// Try sequential positions
		gy_gate = 1 + tries

		valid := true
		// Check the 3 cells where the gate will be placed
		for dx := -1; dx <= 1; dx++ {
			cx, cy := gx_gate+dx, gy_gate
			if cx >= 0 && cx < W && cy >= 0 && cy < H {
				// Only block on star or harbinger (prio >= 9)
				if m.Grid[cy][cx].Prio >= 9 {
					valid = false
					break
				}
			}
		}

		if valid {
			// Check if in rift quadrant
			if gx_gate >= rfQx && gx_gate < rfQx+rfQw && gy_gate >= rfQy && gy_gate < rfQy+rfQh {
				valid = false
			}
		}

		if valid {
			m.set(gx_gate-1, gy_gate, '[', 250, false, 11, "Jump gate")
			m.set(gx_gate, gy_gate, '⟐', 201, true, 11, "Jump gate")
			m.set(gx_gate+1, gy_gate, ']', 250, false, 11, "Jump gate")
			gateFound = true
		}
	}

	// Lanes - MST and non-tree edges
	type Node struct {
		x, y int
		name string
	}

	var nodes []Node
	for i, st := range stations {
		nodes = append(nodes, Node{st.x + 1, st.y, fmt.Sprintf("S%d", i)})
	}
	if gateFound {
		nodes = append(nodes, Node{gx_gate + 1, gy_gate, "Gate"})
	}
	nodes = append(nodes, Node{m.SX, m.SY, "Star"})

	// Prim's MST
	if len(nodes) > 1 {
		inMST := make([]bool, len(nodes))
		edges := make([]struct{ a, b int; dist float64 }, 0)

		inMST[0] = true
		mstEdgeCount := 0

		for mstEdgeCount < len(nodes)-1 {
			minDist := math.MaxFloat64
			minA, minB := -1, -1

			for i := 0; i < len(nodes); i++ {
				if !inMST[i] {
					continue
				}
				for j := 0; j < len(nodes); j++ {
					if inMST[j] {
						continue
					}
					dx, dy := nodes[i].x-nodes[j].x, nodes[i].y-nodes[j].y
					dist := ad(dx, dy)
					if dist < minDist {
						minDist = dist
						minA, minB = i, j
					}
				}
			}

			if minA != -1 {
				inMST[minB] = true
				edges = append(edges, struct{ a, b int; dist float64 }{minA, minB, minDist})
				mstEdgeCount++
			}
		}

		// Add 2 shortest non-tree edges where both stations exist
		type EdgeCandidate struct {
			a, b int
			dist float64
		}
		var candidates []EdgeCandidate
		for i := 0; i < len(nodes); i++ {
			for j := i + 1; j < len(nodes); j++ {
				isInTree := false
				for _, e := range edges {
					if (e.a == i && e.b == j) || (e.a == j && e.b == i) {
						isInTree = true
						break
					}
				}
				if isInTree {
					continue
				}
				dx, dy := nodes[i].x-nodes[j].x, nodes[i].y-nodes[j].y
				dist := ad(dx, dy)
				if dist < 45 {
					candidates = append(candidates, EdgeCandidate{i, j, dist})
				}
			}
		}
		// Sort by distance (simple bubble sort)
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				if candidates[j].dist < candidates[i].dist {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
		addCount := 0
		for _, cand := range candidates {
			if addCount >= 2 {
				break
			}
			edges = append(edges, struct{ a, b int; dist float64 }{cand.a, cand.b, cand.dist})
			addCount++
		}

		// Draw lanes for each edge
		for _, edge := range edges {
			a, b := edge.a, edge.b
			xa, ya := nodes[a].x, nodes[a].y
			xb, yb := nodes[b].x, nodes[b].y

			dx, dy := xb-xa, yb-ya
			m_dist := abs(dx)
			n_dist := abs(dy)

			// Determine colors and kind based on node types
			isGateLink := (nodes[a].name == "Gate") || (nodes[b].name == "Gate")
			isConduit := (nodes[a].name == "Star") || (nodes[b].name == "Star")

			color := 39
			hGlyph := '─'
			kind := "Warp lane"
			if isGateLink {
				color = 51
				hGlyph = '═'
				kind = "Gate link"
			} else if isConduit {
				color = 220
				hGlyph = '═'
				kind = "Gravity conduit"
			}

			// Helper to check if cell should be skipped
			skipCell := func(cx, cy int) bool {
				for _, st := range stations {
					if cx >= st.x-1 && cx <= st.x+1 && cy == st.y {
						return true
					}
				}
				if gateFound {
					if cx >= gx_gate-1 && cx <= gx_gate+1 && cy == gy_gate {
						return true
					}
				}
				return false
			}

			// Route and draw the lane
			x, y := xa, ya

			if m_dist >= n_dist {
				// Horizontal first, then diagonal, then horizontal
				hSteps := (m_dist - n_dist) / 2
				signX := 1
				if dx < 0 {
					signX = -1
				}
				signY := 1
				if dy < 0 {
					signY = -1
				}

				// First horizontal segment
				for i := 0; i < hSteps; i++ {
					x += signX
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, hGlyph, color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}

				// Diagonal segment
				diagGlyph := '╱'
				if signX*signY > 0 {
					diagGlyph = '╱'
				} else {
					diagGlyph = '╲'
				}
				for i := 0; i < n_dist; i++ {
					x += signX
					y += signY
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, diagGlyph, color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}

				// Second horizontal segment
				for i := 0; i < hSteps; i++ {
					x += signX
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, hGlyph, color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}
			} else {
				// Vertical first, then diagonal, then vertical
				vSteps := (n_dist - m_dist) / 2
				signY := 1
				if dy < 0 {
					signY = -1
				}
				signX := 1
				if dx < 0 {
					signX = -1
				}

				// First vertical segment
				for i := 0; i < vSteps; i++ {
					y += signY
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, '│', color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}

				// Diagonal segment
				diagGlyph := '╱'
				if signX*signY > 0 {
					diagGlyph = '╱'
				} else {
					diagGlyph = '╲'
				}
				for i := 0; i < m_dist; i++ {
					x += signX
					y += signY
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, diagGlyph, color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}

				// Second vertical segment
				for i := 0; i < vSteps; i++ {
					y += signY
					if x >= 0 && x < W && y >= 0 && y < H && !skipCell(x, y) {
						m.set(x, y, '│', color, false, 5, kind)
						cellKey := fmt.Sprintf("%d,%d", x, y)
						m.LaneCells[cellKey] = true
					}
				}
			}
		}
	}

	// Freighters
	if len(m.LaneCells) > 0 {
		k := 3 + rng.IntN(4)
		laneCellsSlice := make([]string, 0, len(m.LaneCells))
		for cellKey := range m.LaneCells {
			laneCellsSlice = append(laneCellsSlice, cellKey)
		}

		// Sort to ensure deterministic ordering
		sort.Strings(laneCellsSlice)

		// Pick k cells with spacing of at least 4 cells
		selected := make([]string, 0, k)
		lastX, lastY := -10, -10

		for i := 0; i < len(laneCellsSlice) && len(selected) < k; i++ {
			cellKey := laneCellsSlice[i]
			parts := strings.Split(cellKey, ",")
			cx, _ := strconv.Atoi(parts[0])
			cy, _ := strconv.Atoi(parts[1])

			// Require spacing of at least 4 cells
			if abs(cx-lastX)+abs(cy-lastY) >= 4 {
				selected = append(selected, cellKey)
				lastX, lastY = cx, cy
			}
		}

		// Place freighters
		for _, cellKey := range selected {
			parts := strings.Split(cellKey, ",")
			fx, _ := strconv.Atoi(parts[0])
			fy, _ := strconv.Atoi(parts[1])

			if fx >= 0 && fx < W && fy >= 0 && fy < H {
				cellRune := m.Grid[fy][fx].R
				glyph := rune('▲')
				if cellRune == '─' || cellRune == '═' {
					glyph = '►'
				}
				m.set(fx, fy, glyph, 231, true, 6, "Freighter")
			}
		}
	}

	return m
}

// render generates the output strings for the map.
func render(m *Map, color bool, ix, iy int) []string {
	var lines []string

	// Top frame
	title := fmt.Sprintf("AGEFORGE · DEEP-SPACE CANVAS · seed %d", m.Seed)
	topLine := fmt.Sprintf("┌─ %s ", title)
	titleWidth := uniseg.StringWidth(topLine)
	for titleWidth < 121 {
		topLine += "─"
		titleWidth++
	}
	topLine += "┐"
	lines = append(lines, topLine)

	// Grid rows
	for y := 0; y < H; y++ {
		line := "│"
		var lastColor *int
		var lastBold *bool
		for x := 0; x < W; x++ {
			c := &m.Grid[y][x]
			if color {
				needsChange := lastColor == nil || *lastColor != c.Color || lastBold == nil || *lastBold != c.Bold
				if needsChange {
					if lastColor != nil || lastBold != nil {
						line += "\x1b[0m"
					}
					if c.Color != 0 {
						line += fmt.Sprintf("\x1b[38;5;%dm", c.Color)
					}
					if c.Bold {
						line += "\x1b[1m"
					}
					lastColor = &c.Color
					lastBold = &c.Bold
				}
			}
			line += string(c.R)
		}
		if color {
			line += "\x1b[0m"
		}
		line += "│"
		lines = append(lines, line)
	}

	// Middle frame
	midLine := "├" + strings.Repeat("─", 120) + "┤"
	lines = append(lines, midLine) // width = 1 + 120 + 1 = 122

	// Metrics
	var starCount, stationCount, laneCount, freighterCount int
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			k := m.Grid[y][x].Kind
			if k == "Faint star" || k == "Star" || k == "Bright star" {
				starCount++
			}
			if strings.Contains(k, "lane") || strings.Contains(k, "conduit") || strings.Contains(k, "link") {
				laneCount++
			}
			if k == "Freighter" {
				freighterCount++
			}
		}
	}
	stationCount = len(m.Stations)

	quadrantNames := [4]string{"NW", "NE", "SW", "SE"}
	metricsLine := fmt.Sprintf("│ stars %d · stations %d · lanes %d · freighters %d · rift %s",
		starCount, stationCount, laneCount, freighterCount, quadrantNames[m.RQ])
	for uniseg.StringWidth(metricsLine) < 121 {
		metricsLine += " "
	}
	metricsLine += "│"
	lines = append(lines, metricsLine)

	// Inspect
	var inspectText string
	if ix >= 0 && ix < W && iy >= 0 && iy < H {
		c := &m.Grid[iy][ix]
		inspectText = fmt.Sprintf("│ inspect (%d,%d): %s '%c'", ix, iy, c.Kind, c.R)
	} else {
		inspectText = fmt.Sprintf("│ inspect (%d,%d): out of bounds", ix, iy)
	}
	for uniseg.StringWidth(inspectText) < 121 {
		inspectText += " "
	}
	inspectText += "│"
	lines = append(lines, inspectText)

	// Bottom frame
	botLine := "└" + strings.Repeat("─", 120) + "┘"
	lines = append(lines, botLine) // width = 1 + 120 + 1 = 122

	return lines
}

func main() {
	seed := flag.Int64("seed", 42, "random seed")
	nocolor := flag.Bool("nocolor", false, "disable ANSI color codes")
	inspect := flag.String("inspect", "", "cell to inspect (x,y)")

	flag.Parse()

	m := generate(*seed)

	// Parse inspect
	ix, iy := m.SX, m.SY // central star default
	if *inspect != "" {
		parts := strings.Split(*inspect, ",")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "astromap: invalid inspect format\n")
			os.Exit(2)
		}
		var err1, err2 error
		ix, err1 = strconv.Atoi(parts[0])
		iy, err2 = strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			fmt.Fprintf(os.Stderr, "astromap: invalid inspect coordinates\n")
			os.Exit(2)
		}
	}

	// Width guard
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if uniseg.StringWidth(string(m.Grid[y][x].R)) != 1 {
				fmt.Fprintf(os.Stderr, "astromap: cell (%d,%d) rune %q is not one cell wide\n", x, y, m.Grid[y][x].R)
				os.Exit(1)
			}
		}
	}

	lines := render(m, !*nocolor, ix, iy)
	for _, line := range lines {
		fmt.Println(line)
	}
}
