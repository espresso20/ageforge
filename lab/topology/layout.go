package main

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// Layout is a fixed four-column (three below 120 columns) layered graph.
// Order inside a column is a pure function of the state (canonical resource
// order, then lineage key), so nodes never reshuffle between refreshes; only
// their heights and the wires' tracks respond to size.

type laid struct {
	w, h     int
	top, bot int // diagram rows [top, bot]
	colX     [4]int
	colW     [4]int
	four     bool // hosts get their own column
	ports    []*node
	portRow  map[string]int // port id -> row
	chip     *node
	quiet    int // ports dropped for height
	world    *node // hosts folded into one list node (three-column mode)
	nodes    []*node
}

const (
	pinStat = 1 // producer boxes wire out of their first content row
)

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func layoutModel(m *model, W, H int) *laid {
	L := &laid{w: W, h: H, portRow: map[string]int{}}
	L.top, L.bot = 1, H-2
	if H >= 34 {
		L.top = 2
	}
	L.four = W >= 120
	if L.four {
		pw := clampi(W*17/100, 20, 28)
		rw := clampi(W*27/100, 30, 46)
		sw := clampi(W*17/100, 22, 28)
		hw := clampi(W*15/100, 20, 26)
		ch := (W - pw - rw - sw - hw - 2) / 3
		x := 1
		L.colX[0], L.colW[0] = x, pw
		x += pw + ch
		L.colX[1], L.colW[1] = x, rw
		x += rw + ch
		L.colX[2], L.colW[2] = x, sw
		x += sw + ch
		L.colX[3], L.colW[3] = x, W-1-x
	} else {
		pw := clampi(W*22/100, 16, 26)
		rw := clampi(W*31/100, 24, 32)
		sw := clampi(W*26/100, 19, 30)
		ch := (W - pw - rw - sw - 1) / 2
		x := 0
		L.colX[0], L.colW[0] = x, pw
		x += pw + ch
		L.colX[1], L.colW[1] = x, rw
		x += rw + ch
		L.colX[2], L.colW[2] = x, W-x
	}
	avail := L.bot - L.top + 1

	// ---- the chip (resource bus).
	ports := m.cols[kPort]
	maxPorts := avail - 2
	if len(ports) > maxPorts {
		// Keep alerts and the busiest ports; canonical order among the kept.
		idx := make([]int, len(ports))
		for i := range idx {
			idx[i] = i
		}
		pri := func(n *node) float64 {
			p := math.Abs(n.rate) + 1e-9
			if n.id == "port:@pop" {
				p += 1e18
			}
			return p + float64(n.sev)*1e15
		}
		sort.SliceStable(idx, func(a, b int) bool { return pri(ports[idx[a]]) > pri(ports[idx[b]]) })
		keep := map[int]bool{}
		for _, i := range idx[:maxPorts-1] {
			keep[i] = true
		}
		var kept []*node
		for i, n := range ports {
			if keep[i] {
				kept = append(kept, n)
			} else {
				n.hidden = true
			}
		}
		L.quiet = len(ports) - len(kept)
		ports = kept
	}
	L.ports = ports
	// Air between ports when there is room: the bus then spans about as much
	// height as the producers feeding it, so most wires run level.
	gap := 0
	if n := len(ports); n > 1 {
		// Up to three blank rows between ports, the bus never taller than
		// two thirds of the diagram.
		gap = clampi((avail*2/3-2-n)/(n-1), 0, 3)
	}
	chipH := len(ports) + 2 + (len(ports)-1)*gap
	if L.quiet > 0 {
		chipH++
	}
	chipY := L.top + (avail-chipH)/2
	L.chip = &node{id: "chip", kind: kPort, x: L.colX[1], y: chipY, w: L.colW[1], h: chipH}
	for i, p := range ports {
		p.x, p.y, p.w, p.h = L.chip.x, chipY+1+i*(1+gap), L.chip.w, 1
		L.portRow[p.id] = p.y
	}

	// ---- producers: sized by their share of the buildings (output is not
	// comparable across resources), aligned to the port they feed most.
	prods := m.cols[kProducer]
	total := 0.0
	for _, n := range prods {
		total += n.mass
	}
	for _, n := range prods {
		share := 0.0
		if total > 0 {
			share = n.mass / total
		}
		switch {
		case share >= 0.18:
			n.h = 5
		case share >= 0.10:
			n.h = 4
		case share >= 0.05 || n.alert != "":
			n.h = 3
		default:
			n.h = 1
		}
		if len(prods) <= 3 { // the first impression: few, generous boxes
			n.h = 5
		}
	}
	fitHeights(prods, avail)
	placeColumn(prods, L, func(n *node) int {
		row, best := -1, -1.0
		for _, e := range m.edges {
			if e.from == n.id && L.portRow[e.to] > 0 && e.rate > best {
				row, best = L.portRow[e.to], e.rate
			}
		}
		if row < 0 {
			return L.top
		}
		if n.h >= 3 {
			return row - pinStat
		}
		return row
	})
	for _, n := range prods {
		n.x, n.w = L.colX[0], L.colW[0]
	}

	// ---- services: aligned to the mean row of what feeds them.
	svcs := m.cols[kService]
	hosts := m.cols[kHost]
	if !L.four && len(hosts) > 0 {
		L.world = &node{id: "svc:world", kind: kService, title: "World", h: len(hosts) + 2}
		L.world.gauge = -1
		svcs = append(append([]*node{}, svcs...), L.world)
	}
	for _, n := range svcs {
		if n.h == 0 {
			n.h = 4
			if n.sub == "" {
				n.h = 3
			}
		}
	}
	fitHeights(svcs, avail)
	placeColumn(svcs, L, func(n *node) int {
		sum, k := 0, 0
		for _, e := range m.edges {
			if e.to == n.id || e.from == n.id {
				other := e.from
				if other == n.id {
					other = e.to
				}
				if r, ok := L.portRow[other]; ok {
					sum += r
					k++
				}
			}
		}
		if k == 0 {
			if n == L.world {
				return L.bot
			}
			return L.top + avail/2
		}
		return sum/k - 1
	})
	for _, n := range svcs {
		n.x, n.w = L.colX[2], L.colW[2]
	}

	// ---- hosts.
	if L.four {
		for _, n := range hosts {
			n.h = 3
		}
		fitHeights(hosts, avail)
		placeColumn(hosts, L, func(n *node) int {
			for _, e := range m.edges {
				other := ""
				if e.to == n.id {
					other = e.from
				} else if e.from == n.id {
					other = e.to
				}
				if s := m.nodes[other]; s != nil && s.kind == kService {
					return s.y + 1
				}
			}
			return L.top
		})
		for _, n := range hosts {
			n.x, n.w = L.colX[3], L.colW[3]
		}
	} else {
		for _, n := range hosts {
			n.hidden = true
		}
	}
	for c := 0; c < 4; c++ {
		for _, n := range m.cols[c] {
			if !n.hidden {
				L.nodes = append(L.nodes, n)
			}
		}
	}
	// Centre the drawing's bounding box vertically.
	lo, hi := L.chip.y, L.chip.y+L.chip.h-1
	all := append([]*node{}, L.nodes...)
	if L.world != nil {
		all = append(all, L.world)
	}
	for _, n := range all {
		if n.kind != kPort {
			lo, hi = min(lo, n.y), max(hi, n.y+n.h-1)
		}
	}
	if d := (L.top + L.bot - lo - hi) / 2; d != 0 && lo+d >= L.top && hi+d <= L.bot {
		L.chip.y += d
		for _, n := range all {
			n.y += d
		}
		for k := range L.portRow {
			L.portRow[k] += d
		}
	}
	if L.world != nil {
		L.nodes = append(L.nodes, L.world)
		m.nodes[L.world.id] = L.world
	}
	return L
}

// fitHeights shrinks the smallest nodes until the column fits avail rows
// (with a one-row gap between nodes when there is room for it).
func fitHeights(ns []*node, avail int) {
	steps := map[int]int{5: 4, 4: 3, 3: 1}
	sum := func() int {
		s := 0
		for _, n := range ns {
			s += n.h
		}
		return s + len(ns) - 1
	}
	order := append([]*node{}, ns...)
	sort.SliceStable(order, func(a, b int) bool { return order[a].mass < order[b].mass })
	for sum() > avail {
		changed := false
		// Shrink the tallest-but-smallest first so big producers keep their size.
		for lvl := 5; lvl >= 3 && !changed; lvl-- {
			for _, n := range order {
				if n.h == lvl {
					n.h = steps[lvl]
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}
}

// placeColumn assigns y to nodes in a fixed order, as close to each node's
// desired top as the column allows: overlap removal by merging clusters
// (each cluster sits at the mean of its members' wishes), then clamped.
func placeColumn(ns []*node, L *laid, want func(*node) int) {
	if len(ns) == 0 {
		return
	}
	type item struct {
		n *node
		d int
	}
	items := make([]item, len(ns))
	for i, n := range ns {
		items[i] = item{n, want(n)}
	}
	// The order is the wish order, ties by the model's stable order.
	sort.SliceStable(items, func(a, b int) bool { return items[a].d < items[b].d })
	total := 0
	for _, it := range items {
		total += it.n.h
	}
	gap := 1
	if total+len(items)-1 > L.bot-L.top+1 {
		gap = 0
	}
	type cluster struct {
		first, last int
		sumD, n     float64
		height      int
	}
	var cs []cluster
	for i, it := range items {
		c := cluster{first: i, last: i, sumD: float64(it.d), n: 1, height: it.n.h + gap}
		cs = append(cs, c)
		for len(cs) > 1 {
			a, b := cs[len(cs)-2], cs[len(cs)-1]
			as := a.sumD / a.n
			bs := b.sumD / b.n
			if as+float64(a.height) <= bs {
				break
			}
			merged := cluster{first: a.first, last: b.last, sumD: a.sumD + b.sumD - b.n*float64(a.height), n: a.n + b.n, height: a.height + b.height}
			cs = append(cs[:len(cs)-2], merged)
		}
	}
	y := L.top
	var pos []int
	for _, c := range cs {
		s := int(math.Round(c.sumD / c.n))
		if s < y {
			s = y
		}
		yy := s
		for i := c.first; i <= c.last; i++ {
			pos = append(pos, yy)
			yy += items[i].n.h + gap
		}
		y = yy
	}
	// Pull up from the bottom if the column ran past it.
	limit := L.bot + 1
	for i := len(items) - 1; i >= 0; i-- {
		if pos[i]+items[i].n.h > limit {
			pos[i] = limit - items[i].n.h
		}
		limit = pos[i] - gap
	}
	for i, it := range items {
		it.n.y = max(pos[i], L.top)
	}
}

// ---- wire routing ------------------------------------------------------

type wireEnd struct {
	x, y int
}

// route computes every visible edge's path. Wires run between adjacent
// columns through the channel between them: out of the left node's right
// edge, along a private vertical track, into the right node's left edge.
func route(m *model, L *laid) {
	type seg struct {
		e      *edge
		l, r   wireEnd
		rev    bool // particles run right to left
		netKey string
	}
	chans := map[int][]*seg{}
	colOf := func(n *node) int {
		if n.kind == kHost && !L.four {
			return 2
		}
		return int(n.kind)
	}
	outRows := map[string][]int{} // node -> used right pin rows
	inRows := map[string][]int{}
	pin := func(n *node, other int, used map[string][]int, right bool) wireEnd {
		x := n.x - 1
		if right {
			x = n.x + n.w
		}
		if n.kind == kPort || n.h == 1 {
			return wireEnd{x, n.y}
		}
		// Content rows: prefer the row level with the other end.
		lo, hi := n.y+1, n.y+n.h-2
		if n.kind == kProducer {
			lo = n.y + pinStat
		}
		y := clampi(other, lo, hi)
		used[n.id] = append(used[n.id], y)
		return wireEnd{x, y}
	}
	for _, e := range m.edges {
		a, b := m.nodes[e.from], m.nodes[e.to]
		if a == nil || b == nil {
			continue
		}
		if a.hidden && a.kind == kHost && L.world != nil {
			a = L.world
		}
		if b.hidden && b.kind == kHost && L.world != nil {
			b = L.world
		}
		if a.hidden || b.hidden || a == b {
			continue
		}
		ca, cb := colOf(a), colOf(b)
		rev := false
		if ca > cb {
			a, b, ca, cb = b, a, cb, ca
			rev = true
		}
		if cb != ca+1 {
			continue
		}
		// Anchor rows: the port's row when one end is a port.
		by := b.y + 1
		if b.kind == kPort {
			by = L.portRow[b.id]
		}
		le := pin(a, by, outRows, true)
		re := pin(b, le.y, inRows, false)
		s := &seg{e: e, l: le, r: re, rev: rev}
		if ca == 2 {
			s.netKey = a.id // service -> hosts fan out of one trunk
		} else {
			s.netKey = b.id // everything else collects into the target
		}
		if b.kind == kPort {
			s.netKey = a.id + "|" + b.id
		}
		chans[ca] = append(chans[ca], s)
	}

	for ch, segs := range chans {
		if len(segs) == 0 {
			continue
		}
		xa, xb := segs[0].l.x, segs[0].r.x
		// Nets: one vertical track each.
		type net struct {
			key      string
			lo, hi   int
			lrows    []int
			rrows    []int
			x        int
			straight bool
		}
		nets := map[string]*net{}
		var keys []string
		for _, s := range segs {
			if s.l.y == s.r.y && (nets[s.netKey] == nil) {
				// A level wire needs no track unless its net has others.
			}
			n := nets[s.netKey]
			if n == nil {
				n = &net{key: s.netKey, lo: s.l.y, hi: s.l.y}
				nets[s.netKey] = n
				keys = append(keys, s.netKey)
			}
			for _, y := range []int{s.l.y, s.r.y} {
				n.lo = min(n.lo, y)
				n.hi = max(n.hi, y)
			}
			n.lrows = append(n.lrows, s.l.y)
			n.rrows = append(n.rrows, s.r.y)
		}
		var bent []*net
		for _, k := range keys {
			n := nets[k]
			if n.lo == n.hi {
				n.straight = true
				continue
			}
			bent = append(bent, n)
		}
		// Order tracks to minimise crossings: hill-climb adjacent swaps from
		// a start sorted by span centre.
		sort.SliceStable(bent, func(i, j int) bool {
			return bent[i].lo+bent[i].hi < bent[j].lo+bent[j].hi
		})
		cross := func(order []*net) int {
			c := 0
			for i, a := range order {
				for j, b := range order {
					if i == j {
						continue
					}
					// a's right stubs pass b's track when b is right of a.
					if j > i {
						for _, y := range a.rrows {
							if y > b.lo && y < b.hi {
								c++
							}
						}
					} else {
						for _, y := range a.lrows {
							if y > b.lo && y < b.hi {
								c++
							}
						}
					}
				}
			}
			return c
		}
		best := cross(bent)
		for pass := 0; pass < 6; pass++ {
			improved := false
			for i := 0; i+1 < len(bent); i++ {
				bent[i], bent[i+1] = bent[i+1], bent[i]
				if c := cross(bent); c < best {
					best, improved = c, true
				} else {
					bent[i], bent[i+1] = bent[i+1], bent[i]
				}
			}
			if !improved {
				break
			}
		}
		inner := xb - xa - 1
		// A crowded channel bundles its nets onto a few shared trunks (a
		// bus, as a schematic would draw it) rather than a barcode of
		// adjacent verticals; each wire's packets still follow its own path.
		lanes := len(bent)
		if maxLanes := max(1, inner/3); lanes > maxLanes {
			lanes = maxLanes
		}
		for i, n := range bent {
			if inner <= 0 {
				n.x = xa
				continue
			}
			lane := i * lanes / max(1, len(bent))
			n.x = xa + 1 + (lane+1)*inner/(lanes+1)
			if n.x >= xb {
				n.x = xb - 1
			}
		}
		_ = ch
		for _, s := range segs {
			n := nets[s.netKey]
			var p []pt
			if n.straight || s.l.y == s.r.y && n.x == 0 {
				for x := s.l.x; x <= s.r.x; x++ {
					p = append(p, pt{x, s.l.y})
				}
			} else {
				for x := s.l.x; x <= n.x; x++ {
					p = append(p, pt{x, s.l.y})
				}
				step := 1
				if s.r.y < s.l.y {
					step = -1
				}
				for y := s.l.y + step; y != s.r.y+step; y += step {
					p = append(p, pt{n.x, y})
				}
				for x := n.x + 1; x <= s.r.x; x++ {
					p = append(p, pt{x, s.r.y})
				}
			}
			if s.rev {
				for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
					p[i], p[j] = p[j], p[i]
				}
			}
			s.e.path = p
		}
	}
}

// canonical index of a resource, for stable orders elsewhere.
var resIndex = func() map[string]int {
	m := map[string]int{"@pop": -1}
	for i, r := range config.BaseResources() {
		m[r.Key] = i
	}
	return m
}()
