package main

import (
	"fmt"
	"sort"
	"strings"
)

// The world view is the same grammar zoomed out: the whole city collapses
// into one node and the external hosts (every civilization, known or not,
// trade partners, expeditions, the harbinger) spread around it. It replaces
// the worldmap the way the city view replaces the citymap: one metaphor.

func (v *view) buildWorld() (*node, []*node, []*edge) {
	st := v.m.st
	city := &node{id: "world:city", kind: kService, title: "Your city", badge: st.AgeName, gauge: -1}
	var hosts []*node
	var edges []*edge
	link := func(h *node, rate float64, s edgeState, inbound bool) {
		e := &edge{id: "w>" + h.id, from: city.id, to: h.id, rate: rate, state: s}
		if inbound {
			e.from, e.to = h.id, city.id
		}
		edges = append(edges, e)
	}
	unmet := 0
	for _, key := range sortedMapKeys(st.Diplomacy.Factions) {
		f := st.Diplomacy.Factions[key]
		if !f.Discovered {
			unmet++
			continue
		}
		src := v.m.nodes["host:civ:"+key]
		h := *src
		hosts = append(hosts, &h)
		s := eIdle
		switch {
		case f.AtWar || f.Status == "embargo":
			s = eSevered
		case f.TradeCount > 0 || f.Status == "allied" || f.LentWorkers > 0:
			s = eFlow
		}
		link(&h, 0.2+0.05*float64(f.TradeCount), s, true)
	}
	for _, n := range v.m.cols[kHost] {
		if strings.HasPrefix(n.id, "host:civ:") || n.id == "host:next" {
			continue
		}
		h := *n
		hosts = append(hosts, &h)
		s := eFlow
		if n.sev == sevCrit {
			s = eSevered
		}
		inbound := n.id == "host:harbinger"
		if inbound {
			s = eIdle
		}
		link(&h, 0.3, s, inbound)
	}
	if unmet > 0 {
		// Civilizations not yet met: one quiet placeholder, no link.
		hosts = append(hosts, &node{id: "world:unmet", kind: kHost, title: "? uncharted",
			sub: fmt.Sprintf("%d civs not yet met", unmet), gauge: -1})
	}
	return city, hosts, edges
}

func (v *view) drawWorld() {
	c := v.c
	city, hosts, edges := v.buildWorld()
	top, bot := v.L.top, v.L.bot
	avail := bot - top + 1
	cw := clampi(c.w/4, 26, 40)
	hw := clampi(c.w/5, 20, 30)
	city.w = cw
	city.x = (c.w - cw) / 2
	// Hosts alternate left and right in a stable order.
	var left, right []*node
	for i, h := range hosts {
		if i%2 == 0 {
			left = append(left, h)
		} else {
			right = append(right, h)
		}
	}
	place := func(col []*node, x int) {
		n := len(col)
		if n == 0 {
			return
		}
		h := 3
		if n*4-1 > avail {
			h = 1
		}
		step := avail / n
		for i, nd := range col {
			nd.h, nd.w, nd.x = h, hw, x
			nd.y = top + i*step + (step-h)/2
		}
	}
	place(left, 1)
	place(right, c.w-1-hw)
	// The city: tall enough to give every link its own pin row.
	city.h = clampi(max(len(left), len(right))+4, 9, avail)
	city.y = top + (avail-city.h)/2

	// Wires: from each host to the city's facing side, bending once in the
	// channel between them.
	byID := map[string]*node{}
	for _, h := range hosts {
		byID[h.id] = h
	}
	// Linked hosts per side, top to bottom, each with its own pin row on
	// the city and its own bend column: the host furthest (vertically)
	// from its pin bends nearest the city, so no two wires cross.
	linked := func(col []*node) []*node {
		var out []*node
		for _, h := range col {
			for _, e := range edges {
				if e.from == h.id || e.to == h.id {
					out = append(out, h)
					break
				}
			}
		}
		return out
	}
	pins := map[*node]int{}
	bend := map[*node]int{}
	for _, col := range [][]*node{linked(left), linked(right)} {
		for i, h := range col {
			pins[h] = city.y + 2
			if len(col) > 1 {
				pins[h] = city.y + 2 + i*(city.h-5)/(len(col)-1)
			}
		}
		order := append([]*node{}, col...)
		dist := func(h *node) int {
			d := h.y + h.h/2 - pins[h]
			if d < 0 {
				d = -d
			}
			return d
		}
		sort.SliceStable(order, func(a, b int) bool { return dist(order[a]) > dist(order[b]) })
		for r, h := range order {
			bend[h] = r
		}
	}
	for _, e := range edges {
		hid := e.to
		if hid == city.id {
			hid = e.from
		}
		h := byID[hid]
		if h == nil {
			continue
		}
		var p []pt
		isLeft := h.x < city.x
		k := len(linked(left))
		if !isLeft {
			k = len(linked(right))
		}
		hy := h.y + h.h/2
		cy := pins[h]
		if isLeft {
			x0, x1 := h.x+h.w, city.x-1
			mid := x1 - (x1-x0)*(bend[h]+1)/(k+1)
			p = orth(x0, hy, mid, x1, cy)
		} else {
			x0, x1 := city.x+city.w, h.x-1
			mid := x0 + (x1-x0)*(bend[h]+1)/(k+1)
			p = orth(x0, cy, mid, x1, hy)
		}
		// Paths run city→host; flip for flows into the city.
		flowIn := e.to == city.id
		if isLeft != flowIn {
			for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
				p[i], p[j] = p[j], p[i]
			}
		}
		e.path = p
	}
	saved := v.m.edges
	v.m.edges = edges
	v.drawWires()
	v.drawParticles()
	v.m.edges = saved
	for _, h := range hosts {
		if h.id == "world:unmet" {
			v.box(h.x, h.y, h.w, h.h, cFaint, v.s.boxes)
			if h.h == 1 {
				c.text(h.x+1, h.y, clip("? "+h.sub, h.w-2), cFaint, h.w-2)
				continue
			}
			c.text(h.x+2, h.y, " ? uncharted ", cDim, h.w-4)
			c.text(h.x+2, h.y+1, clip(h.sub, h.w-4), cFaint, h.w-4)
			continue
		}
		v.drawNode(h)
	}
	v.drawCity(city)
	v.drawHeader()
	v.drawFooter()
}

func orth(x0, y0, mid, x1, y1 int) []pt {
	var p []pt
	for x := x0; x <= mid; x++ {
		p = append(p, pt{x, y0})
	}
	step := 1
	if y1 < y0 {
		step = -1
	}
	if y1 != y0 {
		for y := y0 + step; y != y1+step; y += step {
			p = append(p, pt{mid, y})
		}
	}
	for x := mid + 1; x <= x1; x++ {
		p = append(p, pt{x, y1})
	}
	return p
}

// drawCity is the city folded into one node: its busiest stores and people.
func (v *view) drawCity(n *node) {
	c := v.c
	k := cAccent
	if v.m.catast != "" && v.frame%12 < 3 {
		k = cNeg
	}
	v.box(n.x, n.y, n.w, n.h, k, v.s.boxes)
	v.title(n, n.x, n.y, n.w, k)
	iw := n.w - 4
	y := n.y + 1
	w := v.m.st.Workers
	c.textB(n.x+2, y, clip(fmt.Sprintf("%s citizens · %d idle", short(float64(w.TotalPop)), w.TotalIdle), iw), cText, iw, true)
	y++
	ports := append([]*node{}, v.m.cols[kPort]...)
	sort.SliceStable(ports, func(i, j int) bool { return ports[i].rate > ports[j].rate })
	for _, p := range ports {
		if y >= n.y+n.h-2 {
			break
		}
		if p.id == "port:@pop" {
			continue
		}
		gw := iw - 10 - 8
		c.text(n.x+2, y, clip(p.title, 9), cLabel, 9)
		v.gauge(n.x+12, y, gw, p.gauge, cPos)
		st := strings.TrimSuffix(p.stat, "/t")
		c.text(n.x+2+iw-width(st), y, st, cPos, width(st))
		y++
	}
	met := 0
	for _, f := range v.m.st.Diplomacy.Factions {
		if f.Discovered {
			met++
		}
	}
	c.text(n.x+2, n.y+n.h-2, clip(fmt.Sprintf("%d of %d civs met", met, len(v.m.st.Diplomacy.Factions)), iw), cDim, iw)
}
