package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// drawMini is the sidebar form: the resource bus alone, each store with a
// tiny flow lane in front of it (packets crawl in at the inflow's speed,
// back up when the store is full, run backwards when it drains), then the
// most urgent alarm. It is the whole topology collapsed onto its spine, so
// it reads the same as the full view.
func (v *view) drawMini() {
	c := v.c
	st := v.m.st
	c.bgRect(0, 0, c.w, 1, bgChip)
	name := v.s.name
	if c.w < 44 {
		if i := strings.LastIndexByte(name, ' '); i > 0 {
			name = name[i+1:]
		}
	}
	x := c.textB(1, 0, "◆ "+name, cAccent, c.w-2, true)
	age := strings.TrimSuffix(st.AgeName, " Age")
	x += c.text(1+x, 0, " · "+age, cDim, c.w-2-x)
	warn := 0
	for _, n := range v.m.nodes {
		if n.sev >= sevWarn {
			warn++
		}
	}
	if warn > 0 || v.m.catast != "" {
		t := fmt.Sprintf("%c%d", v.s.alarm, len(v.m.alarms))
		c.textB(c.w-1-width(t), 0, t, cNeg, width(t), true)
	}

	// Inflow per port from the wires into it.
	in := map[string]float64{}
	for _, e := range v.m.edges {
		if strings.HasPrefix(e.to, "port:") && e.state != eSevered {
			in[e.to] += e.rate
		}
	}
	rows := c.h - 2
	ports := append([]*node{}, v.m.cols[kPort]...)
	if len(ports) > rows {
		pri := func(n *node) float64 {
			p := math.Abs(n.rate) + in[n.id]
			if n.id == "port:@pop" {
				p += 1e18
			}
			return p + float64(n.sev)*1e15
		}
		idx := make([]int, len(ports))
		for i := range idx {
			idx[i] = i
		}
		sort.SliceStable(idx, func(a, b int) bool { return pri(ports[idx[a]]) > pri(ports[idx[b]]) })
		keep := map[int]bool{}
		for _, i := range idx[:rows] {
			keep[i] = true
		}
		var kept []*node
		for i, p := range ports {
			if keep[i] {
				kept = append(kept, p)
			}
		}
		ports = kept
	}
	lane := 4
	nameW := 7
	if c.w >= 38 {
		nameW = 10
	}
	rateW := 7
	tagW := 0
	for _, p := range ports {
		if p.alert != "" {
			tagW = 5
		}
	}
	gw := c.w - 1 - lane - 1 - nameW - rateW - 1 - tagW
	for i, p := range ports {
		y := 1 + i
		// Flow lane.
		lk := cFaint
		for j := 0; j < lane; j++ {
			c.set(j, y, v.s.lines[mE|mW], lk)
		}
		switch {
		case p.alert == "FULL":
			for j := lane - 3; j < lane; j++ {
				c.set(j, y, v.s.head, cWarn)
			}
		case p.rate < -1e-9:
			pos := lane - 1 - (v.frame/2)%lane
			c.setB(pos, y, v.s.head, cNeg, true)
		case in[p.id] > 0 || p.rate > 0:
			sp := v.speed(math.Max(in[p.id], p.rate))
			pos := int(float64(v.frame)*sp+float64(i)) % (lane + 2)
			if pos < lane {
				c.setB(pos, y, v.s.head, cPos, true)
				if pos > 0 && v.s.trail != 0 {
					c.set(pos-1, y, v.s.trail, cPos)
				}
			}
		}
		c.set(lane, y, v.s.pinIn, cDim)
		x := lane + 1
		nk := cLabel
		if p.id == "port:@pop" {
			nk = cAccent
		}
		c.text(x, y, clip(p.title, nameW-1), nk, nameW-1)
		x += nameW
		gk := cPos
		switch p.sev {
		case sevWarn:
			gk = cWarn
		case sevCrit:
			gk = cNeg
		}
		if p.id == "port:@pop" {
			gk = cAccent
		}
		v.gauge(x, y, gw, p.gauge, gk)
		x += gw + 1
		rk := cDim
		switch {
		case p.rate > 0:
			rk = cPos
		case p.rate < 0:
			rk = cNeg
		}
		s := strings.TrimSuffix(p.stat, "/t")
		if p.id == "port:@pop" {
			rk, s = cText, short(p.amount)
		}
		c.text(x+rateW-1-width(s), y, clip(s, rateW-1), rk, rateW-1)
		x += rateW
		if p.alert != "" && tagW > 0 {
			k := cWarn
			if p.sev == sevCrit {
				k = cNeg
			}
			c.textB(x, y, clip(p.alert, tagW), k, tagW, true)
		}
	}
	y := c.h - 1
	c.bgRect(0, y, c.w, 1, bgChip)
	switch {
	case len(v.m.alarms) > 0:
		k := cWarn
		if v.m.catast != "" || strings.Contains(v.m.alarms[0], "WAR") {
			k = cNeg
		}
		c.setB(1, y, v.s.alarm, k, true)
		c.textB(3, y, clip(v.m.alarms[0], c.w-4), k, c.w-4, true)
	case v.m.fx.Prev != nil:
		c.text(1, y, "Δ "+clip(v.sinceLine(), c.w-4), cDim, c.w-2)
	default:
		c.text(1, y, "● nominal", cPos, c.w-2)
	}
}
