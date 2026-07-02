package citymap

import (
	"container/heap"
	"math"
)

// pathfind.go routes terrain-aware roads. The old roads were straight Bresenham
// radii from each building to the center, which sliced through lakes and peaks and
// read as spokes, not streets. Instead we run A* over a downsampled cost grid built
// from the biome field: passable land is cheap, water/mountain is blocked. The
// earlier version added per-node noise jitter + a heavy clearance penalty to make
// roads "meander"; in practice that produced spaghetti — roads wandered all over open
// ground and bunched into parallel detours. We now keep the grid nearly flat so a road
// on open land runs essentially STRAIGHT and only bends where an impassable body forces
// a detour; a tiny clearance penalty keeps it from clipping a shoreline corner. The A*
// staircase is then collapsed to a few clean segments (Douglas–Peucker) and lightly
// smoothed, so the result reads as a road, not a jaggy per-node stair. Pathfinding runs
// only when the cached image regenerates (age/size/buildings/theme change), so a
// per-render A* over the handful of roads is cheap.

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

// pfClearancePenalty is the per-blocked-neighbour cost added to a passable node. Kept
// small (was 0.6): just enough that, given two equal-length routes, A* prefers the one
// a hair off the shoreline so a road doesn't scrape a water/cliff corner — but NOT so
// large that it detours wide of an obstacle into a parallel bundle. The base travel cost
// is 1.0, so a single blocked neighbour is a 15% nudge, not a wall.
const pfClearancePenalty = 0.15

// buildCostGrid downsamples the terrain field into a node cost grid. A node is
// blocked when its center pixel is impassable. Passable nodes get a base cost of 1.0
// plus a small clearance penalty that ramps up near blocked neighbours, so — all else
// equal — roads keep a hair of clearance from shorelines/cliffs instead of scraping
// along them. There is deliberately NO per-node jitter anymore: the old wobble made
// open-ground roads wander into spaghetti. With a near-flat grid, A* rules a straight
// line across open land and only bends to route around impassable water/mountain, which
// is the clean look we want. seed is retained for signature/compat but no longer feeds
// randomness into the grid.
func buildCostGrid(f *terrainField, seed uint32) *pfCostGrid {
	_ = seed // no jitter anymore — grid is deterministic from the terrain field alone
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
			// Base travel cost — flat across all open land so roads run straight.
			c := 1.0
			// Clearance penalty: look one node out in 8 directions; each blocked or
			// off-canvas neighbour adds a small cost, so between two equal routes A*
			// prefers the one a hair off the shoreline/cliff. Small on purpose — a big
			// penalty is what used to push roads into wide parallel detours.
			blockedNbr := 0
			for _, d := range pfNeighbors {
				ax, ay := g.nodePx(nx+d[0], ny+d[1])
				if !f.passableAt(ax, ay) {
					blockedNbr++
				}
			}
			c += pfClearancePenalty * float64(blockedNbr)
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

// pxBlocked reports whether the pixel (px,py) falls on a blocked node or off the grid.
// Unlike pxToNode it does NOT clamp: an off-canvas pixel is blocked (the edge is a wall),
// so the chord-clearance test can't approve a segment that leaves the map or grazes
// water. Used by path simplification to keep simplified chords on passable ground.
func (g *pfCostGrid) pxBlocked(px, py int) bool {
	if g == nil || g.cols <= 0 || g.rows <= 0 {
		return true
	}
	nx, ny := px/pfStep, py/pfStep
	if px < 0 || py < 0 || nx < 0 || ny < 0 || nx >= g.cols || ny >= g.rows {
		return true
	}
	return g.blocked[ny*g.cols+nx]
}

// segClear reports whether the straight pixel segment a→b stays entirely on passable
// terrain (no sampled pixel hits a blocked node). It walks the segment at ~1px steps via
// the segment's dominant axis. This is the guard that keeps Douglas–Peucker from
// collapsing a detour into a chord that cuts across the very water the detour avoided.
func (g *pfCostGrid) segClear(a, b [2]int) bool {
	dx := b[0] - a[0]
	dy := b[1] - a[1]
	steps := absInt(dx)
	if absInt(dy) > steps {
		steps = absInt(dy)
	}
	if steps == 0 {
		return !g.pxBlocked(a[0], a[1])
	}
	for i := 0; i <= steps; i++ {
		x := a[0] + dx*i/steps
		y := a[1] + dy*i/steps
		if g.pxBlocked(x, y) {
			return false
		}
	}
	return true
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

// simplifyPath runs a terrain-aware Douglas–Peucker on the raw A* polyline: it drops a
// waypoint only when it lies within eps pixels of the straight chord between the kept
// points AND that chord stays on passable ground (grid.segClear). Collapsing the coarse-
// grid staircase this way yields a few clean segments; the segClear guard is what stops
// DP from cutting a corner across the water the detour just avoided. On open ground the
// whole path reduces to a single start→goal segment (dead straight); around an obstacle
// it keeps the corner vertices of the detour. Endpoints are always preserved. eps ~1.5px
// is well under a cell, so genuine bends survive while the 45°/90° stair noise is gone.
// A nil grid disables the clearance guard (pure geometric DP) for callers without one.
func simplifyPath(pts [][2]int, eps float64, grid *pfCostGrid) [][2]int {
	if len(pts) < 3 {
		return pts
	}
	keep := make([]bool, len(pts))
	keep[0] = true
	keep[len(pts)-1] = true
	var rec func(lo, hi int)
	rec = func(lo, hi int) {
		if hi <= lo+1 {
			return
		}
		// Farthest point from the chord pts[lo]→pts[hi].
		maxD, maxI := -1.0, -1
		for i := lo + 1; i < hi; i++ {
			d := perpDist(pts[i], pts[lo], pts[hi])
			if d > maxD {
				maxD, maxI = d, i
			}
		}
		// Split when a point bows past eps, OR when the chord would cut across blocked
		// terrain (force-keep the farthest point so the detour's corner survives).
		mustSplit := maxD > eps
		if !mustSplit && grid != nil && !grid.segClear(pts[lo], pts[hi]) {
			mustSplit = true
		}
		if mustSplit && maxI > 0 {
			keep[maxI] = true
			rec(lo, maxI)
			rec(maxI, hi)
		}
	}
	rec(0, len(pts)-1)
	out := make([][2]int, 0, len(pts))
	for i, k := range keep {
		if k {
			out = append(out, pts[i])
		}
	}
	return out
}

// perpDist is the perpendicular distance from point p to the line through a and b (a==b
// degrades to the point distance). Used by Douglas–Peucker to score how far a waypoint
// bows off its chord.
func perpDist(p, a, b [2]int) float64 {
	ax, ay := float64(a[0]), float64(a[1])
	bx, by := float64(b[0]), float64(b[1])
	px, py := float64(p[0]), float64(p[1])
	dx, dy := bx-ax, by-ay
	denom := dx*dx + dy*dy
	if denom == 0 {
		return math.Hypot(px-ax, py-ay)
	}
	// |cross(b-a, p-a)| / |b-a|.
	cross := dx*(py-ay) - dy*(px-ax)
	return math.Abs(cross) / math.Sqrt(denom)
}

// ---- road network (MST) -----------------------------------------------------
//
// The old road layer routed one independent A* path from EACH building straight to the
// City Center. Every path shared the center endpoint, so they converged on the hub as a
// bundle of parallel, doubled-back segments — the tangle. buildRoadNetwork replaces that
// with a shared network: it builds a minimum-spanning-tree over {City Center + every
// building} and routes only the tree edges. Because each node connects to its NEAREST
// tree neighbour (not always to the center), roads chain building→building→trunk→center
// and MERGE instead of running parallel into the hub — a clean little street network.
//
// The MST edge metric is straight-line distance (cheap, O(N²) over a handful of nodes);
// terrain-awareness comes from routing each of the N-1 chosen edges with A* on the cost
// grid, so a road that must cross a strait still bends around it. Shared trunk cells are
// naturally single because the tree has no redundant edges; we additionally de-duplicate
// identical segments so a coincidental overlap isn't painted twice.

// buildRoadNetwork returns the road network connecting center + buildings as a set of
// routed, simplified segments (single-pixel, ready for drawRoad). It computes a Euclidean
// MST over the node set, then A*-routes and smooths each tree edge on the terrain grid;
// an edge whose A* fails (endpoint walled off) falls back to a straight segment so the
// tree stays connected. Nodes: index 0 is always the City Center; 1..n are buildings.
// Deterministic (no randomness). Returns nil for <2 nodes (nothing to connect).
func buildRoadNetwork(grid *pfCostGrid, center [2]int, buildings [][2]int) []roadSeg {
	nodes := make([][2]int, 0, len(buildings)+1)
	nodes = append(nodes, center)
	nodes = append(nodes, buildings...)
	n := len(nodes)
	if n < 2 {
		return nil
	}

	// Prim's MST with straight-line weights. inTree[i] marks placed nodes; best[i] is the
	// cheapest edge from the current tree to node i, parent[i] its other endpoint.
	inTree := make([]bool, n)
	best := make([]float64, n)
	parent := make([]int, n)
	for i := range best {
		best[i] = math.Inf(1)
		parent[i] = -1
	}
	best[0] = 0 // start the tree at the City Center
	edges := make([][2]int, 0, n-1)
	for k := 0; k < n; k++ {
		// Pick the un-placed node with the cheapest connecting edge.
		u, bd := -1, math.Inf(1)
		for i := 0; i < n; i++ {
			if !inTree[i] && best[i] < bd {
				bd, u = best[i], i
			}
		}
		if u < 0 {
			break // remaining nodes unreachable (shouldn't happen with a finite metric)
		}
		inTree[u] = true
		if parent[u] >= 0 {
			edges = append(edges, [2]int{parent[u], u})
		}
		// Relax the frontier: nearer-to-u nodes may now have a cheaper connection.
		for v := 0; v < n; v++ {
			if inTree[v] {
				continue
			}
			d := nodeDist(nodes[u], nodes[v])
			if d < best[v] {
				best[v] = d
				parent[v] = u
			}
		}
	}

	// Route each tree edge on the terrain grid and collect its simplified segments,
	// de-duplicating so a shared cell run isn't painted twice.
	seen := map[roadSeg]bool{}
	out := make([]roadSeg, 0, (n-1)*2)
	addSeg := func(s roadSeg) {
		// Normalize orientation so A→B and B→A dedupe to one entry.
		if s.x0 > s.x1 || (s.x0 == s.x1 && s.y0 > s.y1) {
			s = roadSeg{s.x1, s.y1, s.x0, s.y0}
		}
		if s.x0 == s.x1 && s.y0 == s.y1 {
			return // zero-length
		}
		if seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, e := range edges {
		a, b := nodes[e[0]], nodes[e[1]]
		if pts, ok := grid.findPath(a[0], a[1], b[0], b[1]); ok && len(pts) >= 2 {
			sm := smoothPath(pts, grid)
			for i := 0; i+1 < len(sm); i++ {
				addSeg(roadSeg{sm[i][0], sm[i][1], sm[i+1][0], sm[i+1][1]})
			}
			continue
		}
		// A* couldn't reach (endpoint boxed in by water) — keep the tree connected with a
		// straight fallback so the network never fragments.
		addSeg(roadSeg{a[0], a[1], b[0], b[1]})
	}
	return out
}

// nodeDist is the Euclidean distance between two pixel nodes — the MST edge metric.
func nodeDist(a, b [2]int) float64 {
	return math.Hypot(float64(a[0]-b[0]), float64(a[1]-b[1]))
}

// smoothPath first simplifies the A* polyline to its essential vertices (terrain-aware
// Douglas–Peucker), then applies ONE light 3-point averaging pass to the surviving
// interior corners so a bend around an obstacle reads as a gentle curve rather than a
// hard elbow. Endpoints are pinned so the road still meets the building and the center
// exactly. On open ground simplify already collapses the path to a straight segment, so
// there is nothing left to smooth and the road is a clean straight line. A smoothed
// vertex is reverted to its pre-smooth position if the nudge would push it (or either
// adjoining segment) onto a blocked cell, so smoothing can never move a road onto water.
func smoothPath(pts [][2]int, grid *pfCostGrid) [][2]int {
	pts = simplifyPath(pts, 1.5, grid)
	if len(pts) < 3 {
		return pts
	}
	out := make([][2]int, len(pts))
	out[0] = pts[0]
	out[len(pts)-1] = pts[len(pts)-1]
	for i := 1; i < len(pts)-1; i++ {
		ax := (pts[i-1][0] + 2*pts[i][0] + pts[i+1][0]) / 4
		ay := (pts[i-1][1] + 2*pts[i][1] + pts[i+1][1]) / 4
		cand := [2]int{ax, ay}
		// Keep the original vertex if the smoothed one (or the segments into/out of it)
		// would cross blocked terrain — smoothing must never relocate a road onto water.
		if grid != nil {
			prev := out[i-1]
			if grid.pxBlocked(cand[0], cand[1]) ||
				!grid.segClear(prev, cand) || !grid.segClear(cand, pts[i+1]) {
				cand = pts[i]
			}
		}
		out[i] = cand
	}
	return out
}
