package citymap

import (
	"container/heap"
	"math"
)

// pathfind.go routes terrain-aware roads. The old roads were straight Bresenham
// radii from each building to the center, which sliced through lakes and peaks and
// read as spokes, not streets. Instead we run A* over a downsampled cost grid built
// from the biome field: passable land is cheap, water/mountain is blocked, and a
// little per-cell noise makes even open-terrain routes wander rather than ruling a
// straight line. The result bends around obstacles and gently meanders, so it reads
// as a road. Pathfinding runs only when the cached image regenerates (age/size/
// buildings/theme change), so a per-render A* over the handful of roads is cheap.

// pfStep is the downsample factor: one path node per pfStep×pfStep px block. Coarse
// enough that A* over a big canvas stays cheap, fine enough that routes still hug
// the terrain. 3px ≈ 1.5 cells, a sensible road granularity at half-block res.
const pfStep = 3

// pfCostGrid is the node-resolution cost grid A* searches. cols/rows are node
// counts; cost[i] is the per-node entry cost (lower = preferred); blocked[i] marks
// impassable nodes (water/mountain) that A* may never enter. Node n maps back to a
// pixel center via nodePx.
type pfCostGrid struct {
	cols, rows int
	cost       []float64
	blocked    []bool
	w, h       int // source pixel dimensions (for clamping node centers)
}

// buildCostGrid downsamples the terrain field into a node cost grid. A node is
// blocked when its center pixel is impassable. Passable nodes get a base cost plus
// (1) a penalty that ramps up near blocked neighbours, so roads prefer to keep a
// little clearance from shorelines/cliffs instead of scraping along them, and (2) a
// small deterministic per-node noise jitter so open stretches still wobble. seed
// ties the jitter to the render so the same city always routes identically.
func buildCostGrid(f *terrainField, seed uint32) *pfCostGrid {
	cols := (f.w + pfStep - 1) / pfStep
	rows := (f.h + pfStep - 1) / pfStep
	g := &pfCostGrid{cols: cols, rows: rows, w: f.w, h: f.h}
	if cols <= 0 || rows <= 0 {
		return g
	}
	g.cost = make([]float64, cols*rows)
	g.blocked = make([]bool, cols*rows)

	for ny := 0; ny < rows; ny++ {
		for nx := 0; nx < cols; nx++ {
			px, py := g.nodePx(nx, ny)
			idx := ny*cols + nx
			if !f.passableAt(px, py) {
				g.blocked[idx] = true
				g.cost[idx] = math.Inf(1)
				continue
			}
			// Base travel cost.
			c := 1.0
			// Clearance penalty: look one node out in 8 directions; each blocked or
			// off-canvas neighbour adds a little cost, so the cheapest routes give
			// water/cliffs a berth and only hug them when boxed in.
			blockedNbr := 0
			for _, d := range pfNeighbors {
				ax, ay := g.nodePx(nx+d[0], ny+d[1])
				if !f.passableAt(ax, ay) {
					blockedNbr++
				}
			}
			c += 0.6 * float64(blockedNbr)
			// Organic meander: a per-node noise jitter in ~[0,0.5]. Hashing the node
			// coords keeps it deterministic and independent of the elevation field, so
			// even a wide-open plain gives a slightly wandering path rather than a ruler
			// line. Kept small so it bends the route without making it ramble.
			c += hashUnit(uint32(nx), uint32(ny), seed) * 0.5
			g.cost[idx] = c
		}
	}
	return g
}

// nodePx returns the pixel center of node (nx,ny), clamped to the canvas so an edge
// node still maps to a valid pixel.
func (g *pfCostGrid) nodePx(nx, ny int) (int, int) {
	px := nx*pfStep + pfStep/2
	py := ny*pfStep + pfStep/2
	return clampInt(px, 0, maxInt(g.w-1, 0)), clampInt(py, 0, maxInt(g.h-1, 0))
}

// pxToNode maps a pixel coordinate to its containing node, clamped into the grid.
func (g *pfCostGrid) pxToNode(px, py int) (int, int) {
	nx := clampInt(px/pfStep, 0, maxInt(g.cols-1, 0))
	ny := clampInt(py/pfStep, 0, maxInt(g.rows-1, 0))
	return nx, ny
}

// pfNeighbors is the 8-connected neighbour offset set (node coords).
var pfNeighbors = [8][2]int{
	{1, 0}, {-1, 0}, {0, 1}, {0, -1},
	{1, 1}, {1, -1}, {-1, 1}, {-1, -1},
}

// pfNode is a frontier entry for the A* priority queue.
type pfNode struct {
	idx   int     // node index (ny*cols+nx)
	fCost float64 // g + heuristic, the priority key
}

// pfHeap is a min-heap of pfNode by fCost.
type pfHeap []pfNode

func (h pfHeap) Len() int            { return len(h) }
func (h pfHeap) Less(i, j int) bool  { return h[i].fCost < h[j].fCost }
func (h pfHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *pfHeap) Push(x interface{}) { *h = append(*h, x.(pfNode)) }
func (h *pfHeap) Pop() interface{} {
	old := *h
	n := len(old)
	v := old[n-1]
	*h = old[:n-1]
	return v
}

// findPath runs A* on the cost grid from the node containing (x0,y0) to the node
// containing (x1,y1) and returns the route as a polyline of PIXEL waypoints (node
// centers), start→goal inclusive. Returns (nil,false) when no route exists (the
// goal is walled off by water/mountain) so the caller can fall back to a direct
// line. 8-connected; diagonal moves cost √2× the entered node so the metric is
// Euclidean and paths don't favour staircases.
func (g *pfCostGrid) findPath(x0, y0, x1, y1 int) ([][2]int, bool) {
	if g == nil || g.cols <= 0 || g.rows <= 0 {
		return nil, false
	}
	sx, sy := g.pxToNode(x0, y0)
	gx, gy := g.pxToNode(x1, y1)
	start := sy*g.cols + sx
	goal := gy*g.cols + gx

	// If either endpoint sits on a blocked node (a building nudged to the very edge
	// of a lake, or a center that landed on water before its own nudge), snap it to
	// the nearest passable node so A* has a valid entry/exit. If none is found nearby
	// the route is genuinely impossible → signal a fallback.
	if g.blocked[start] {
		if n, ok := g.nearestOpen(sx, sy); ok {
			start = n
		} else {
			return nil, false
		}
	}
	if g.blocked[goal] {
		if n, ok := g.nearestOpen(gx, gy); ok {
			goal = n
		} else {
			return nil, false
		}
	}
	if start == goal {
		px, py := g.nodePx(sx, sy)
		return [][2]int{{px, py}}, true
	}

	n := g.cols * g.rows
	gScore := make([]float64, n)
	for i := range gScore {
		gScore[i] = math.Inf(1)
	}
	came := make([]int, n)
	for i := range came {
		came[i] = -1
	}
	closed := make([]bool, n)

	heur := func(a int) float64 {
		ax, ay := a%g.cols, a/g.cols
		dx := float64(ax - goal%g.cols)
		dy := float64(ay - goal/g.cols)
		return math.Hypot(dx, dy)
	}

	gScore[start] = 0
	open := &pfHeap{{idx: start, fCost: heur(start)}}
	heap.Init(open)

	for open.Len() > 0 {
		cur := heap.Pop(open).(pfNode)
		ci := cur.idx
		if ci == goal {
			return g.reconstruct(came, goal), true
		}
		if closed[ci] {
			continue // a stale (higher-cost) duplicate left in the heap
		}
		closed[ci] = true
		cx, cy := ci%g.cols, ci/g.cols
		for _, d := range pfNeighbors {
			nxn, nyn := cx+d[0], cy+d[1]
			if nxn < 0 || nyn < 0 || nxn >= g.cols || nyn >= g.rows {
				continue
			}
			ni := nyn*g.cols + nxn
			if g.blocked[ni] || closed[ni] {
				continue
			}
			// Step cost = the entered node's cost, scaled √2 for diagonal moves so the
			// metric stays Euclidean.
			step := g.cost[ni]
			if d[0] != 0 && d[1] != 0 {
				step *= math.Sqrt2
			}
			tentative := gScore[ci] + step
			if tentative < gScore[ni] {
				gScore[ni] = tentative
				came[ni] = ci
				heap.Push(open, pfNode{idx: ni, fCost: tentative + heur(ni)})
			}
		}
	}
	return nil, false // goal unreachable — caller falls back to a direct line
}

// reconstruct walks the came-from chain back from the goal and returns the pixel
// waypoints in start→goal order.
func (g *pfCostGrid) reconstruct(came []int, goal int) [][2]int {
	var rev []int
	for n := goal; n != -1; n = came[n] {
		rev = append(rev, n)
	}
	out := make([][2]int, 0, len(rev))
	for i := len(rev) - 1; i >= 0; i-- {
		nx, ny := rev[i]%g.cols, rev[i]/g.cols
		px, py := g.nodePx(nx, ny)
		out = append(out, [2]int{px, py})
	}
	return out
}

// nearestOpen ring-searches outward from node (nx,ny) for the closest non-blocked
// node, returning its index. Used to snap a blocked endpoint onto walkable ground
// before A* so a road can still reach a building parked at a lake's edge. Bounded
// radius — a building truly drowned in open water returns (0,false) → fallback.
func (g *pfCostGrid) nearestOpen(nx, ny int) (int, bool) {
	maxR := g.cols
	if g.rows > maxR {
		maxR = g.rows
	}
	for r := 1; r <= maxR; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if absInt(dx) != r && absInt(dy) != r {
					continue // ring perimeter only
				}
				ax, ay := nx+dx, ny+dy
				if ax < 0 || ay < 0 || ax >= g.cols || ay >= g.rows {
					continue
				}
				i := ay*g.cols + ax
				if !g.blocked[i] {
					return i, true
				}
			}
		}
	}
	return 0, false
}

// smoothPath averages each interior waypoint with its neighbours (a single
// 3-point box pass) so the A* polyline reads as a gentle curve rather than a chain
// of 45°/90° node steps. Endpoints are pinned so the road still meets the building
// and the center exactly. With ≥3 points this softens the staircase A* produces on
// the coarse grid into the meander we want.
func smoothPath(pts [][2]int) [][2]int {
	if len(pts) < 3 {
		return pts
	}
	out := make([][2]int, len(pts))
	out[0] = pts[0]
	out[len(pts)-1] = pts[len(pts)-1]
	for i := 1; i < len(pts)-1; i++ {
		ax := (pts[i-1][0] + pts[i][0] + pts[i+1][0]) / 3
		ay := (pts[i-1][1] + pts[i][1] + pts[i+1][1]) / 3
		out[i] = [2]int{ax, ay}
	}
	return out
}
