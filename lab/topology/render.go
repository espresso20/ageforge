package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// view is everything one frame needs.
type view struct {
	m      *model
	L      *laid
	s      style
	c      *canvas
	frame  int
	sel    string // selected node id, "" for none
	inspect bool
	diff   bool // mark what changed since the last look
	staged string
	world  bool // zoomed out: the city as one node among the world's hosts
}

func newView(fx *Fixture, w, h int) *view {
	m := buildModel(fx)
	v := &view{m: m, s: styleFor(fx.State.Age), diff: fx.Prev != nil}
	v.resize(w, h)
	return v
}

func (v *view) resize(w, h int) {
	// Rebuild from the state: layout is a pure function of (state, size).
	v.m = buildModel(v.m.fx)
	v.L = layoutModel(v.m, w, h)
	route(v.m, v.L)
	v.c = newCanvas(w, h)
	if v.sel == "" || v.m.nodes[v.sel] == nil || v.m.nodes[v.sel].hidden {
		v.sel = ""
	}
}

// draw renders frame f.
func (v *view) draw() {
	c := v.c
	for i := range c.cells {
		c.cells[i] = cell{r: ' '}
	}
	if v.L.w < 60 || v.L.h < 16 {
		v.drawMini()
		return
	}
	if v.world {
		v.drawWorld()
		return
	}
	v.drawWires()
	v.drawParticles()
	v.drawChip()
	for _, n := range v.L.nodes {
		if n.kind == kPort {
			continue
		}
		v.drawNode(n)
	}
	v.drawHeader()
	v.drawFooter()
	if v.inspect && v.sel != "" {
		v.drawInspector()
	}
}

// ---- wires ------------------------------------------------------------

func (v *view) edgeClass(e *edge) cls {
	switch e.state {
	case eSevered, eStarve:
		return cNeg
	case eBackflow:
		return cWarn
	case eIdle:
		return cFaint
	}
	if v.alarmOn(v.m.nodes[e.from]) || v.alarmOn(v.m.nodes[e.to]) {
		return cNeg
	}
	return cWire
}

func (v *view) drawWires() {
	c := v.c
	masks := make([]uint8, c.w*c.h)
	klass := make([]cls, c.w*c.h)
	rank := func(k cls) int {
		switch k {
		case cNeg:
			return 4
		case cWarn:
			return 3
		case cWire:
			return 2
		case cAccent:
			return 5
		}
		return 1
	}
	for _, e := range v.m.edges {
		if len(e.path) < 2 {
			continue
		}
		k := v.edgeClass(e)
		if v.sel != "" && (e.from == v.sel || e.to == v.sel) {
			k = cAccent
		}
		for i, p := range e.path {
			if !c.in(p.x, p.y) {
				continue
			}
			var mk uint8
			link := func(q pt) {
				switch {
				case q.x == p.x+1:
					mk |= mE
				case q.x == p.x-1:
					mk |= mW
				case q.y == p.y+1:
					mk |= mS
				case q.y == p.y-1:
					mk |= mN
				}
			}
			if i > 0 {
				link(e.path[i-1])
			} else {
				link(pt{2*p.x - e.path[1].x, 2*p.y - e.path[1].y})
			}
			if i+1 < len(e.path) {
				link(e.path[i+1])
			} else {
				link(pt{2*p.x - e.path[i-1].x, 2*p.y - e.path[i-1].y})
			}
			j := p.y*c.w + p.x
			masks[j] |= mk
			if rank(k) > rank(klass[j]) {
				klass[j] = k
			}
		}
	}
	for j, mk := range masks {
		if mk == 0 {
			continue
		}
		c.cells[j].r = v.s.lines[mk]
		c.cells[j].fg = klass[j]
	}
	// Arrow heads where wires meet a node, and cuts on severed links.
	for _, e := range v.m.edges {
		n := len(e.path)
		if n < 2 {
			continue
		}
		head := e.path[n-1]
		k := v.edgeClass(e)
		if e.state == eSevered {
			mid := e.path[n/2]
			c.set(mid.x, mid.y, v.s.sever, cNeg)
			if n > 4 {
				a, b := e.path[n/2-1], e.path[n/2+1]
				c.set(a.x, a.y, ' ', cNeg)
				c.set(b.x, b.y, ' ', cNeg)
			}
			continue
		}
		prev := e.path[n-2]
		r := v.s.pinIn
		if prev.x > head.x {
			r = flipArrow(r)
		}
		if k == cFaint {
			k = cDim
		}
		c.set(head.x, head.y, r, k)
	}
}

func flipArrow(r rune) rune {
	switch r {
	case '▶':
		return '◀'
	case '◀':
		return '▶'
	case '▸':
		return '◂'
	case '>':
		return '<'
	}
	return r
}

// ---- particles -------------------------------------------------------

// speed in cells per frame for a rate: log-scaled so a trickle crawls and a
// flood races, whatever the age's magnitudes.
func (v *view) speed(rate float64) float64 {
	if rate <= 0 {
		return 0
	}
	s := 0.18 + 0.12*math.Log10(1+rate*20)
	return math.Min(1.4, s) * v.s.speed
}

func (v *view) drawParticles() {
	c := v.c
	for ei, e := range v.m.edges {
		n := len(e.path)
		if n < 3 || e.state == eSevered || e.state == eIdle {
			continue
		}
		sp := v.speed(e.rate)
		if sp == 0 {
			continue
		}
		fg := cPos
		if e.state == eBackflow {
			fg = cWarn
		}
		if v.alarmOn(v.m.nodes[e.from]) || v.alarmOn(v.m.nodes[e.to]) {
			fg = cNeg
		}
		spacing := 7.0
		if v.s.fibre || v.s.quantum {
			spacing = 11
		}
		off := float64(ei*5%int(spacing)) + float64(v.frame)*sp
		body := e.path[:n-1] // keep the arrow head clear
		limit := len(body)
		if e.state == eBackflow {
			// Pile-up: packets queue against the full store.
			q := min(3, limit-1)
			for i := 0; i < q; i++ {
				p := body[limit-1-i]
				c.set(p.x, p.y, v.s.head, cWarn)
			}
			limit -= q
		}
		for base := math.Mod(off, spacing); base < float64(limit); base += spacing {
			i := int(base)
			if i <= 0 || i >= limit {
				continue
			}
			p := body[i]
			if v.s.fibre {
				// A lit pulse: three cells of heavy line, brightest at the head.
				for k := 0; k < 3 && i-k > 0; k++ {
					q := body[i-k]
					r := litLine(c.at(q.x, q.y).r, body, i-k)
					cl := fg
					if k == 0 {
						cl = cBright
					}
					c.setB(q.x, q.y, r, cl, k == 0)
				}
				continue
			}
			c.setB(p.x, p.y, v.s.head, fg, true)
			if i > 1 && v.s.trail != 0 {
				q := body[i-1]
				c.set(q.x, q.y, v.s.trail, fg)
			}
			if v.s.quantum && (v.frame/2+ei)%2 == 0 {
				// Superposition: the same packet also further along.
				j := i + int(spacing/2)
				if j < limit {
					q := body[j]
					c.set(q.x, q.y, '◇', cHi)
				}
			}
		}
	}
}

// litLine turns a light wire glyph into its heavy twin for a fibre pulse;
// a cell already lit (or overdrawn) takes the heavy glyph of its own path.
func litLine(r rune, path []pt, i int) rune {
	for mk, g := range linesLight {
		if g == r && mk != 0 {
			return linesHeavy[mk]
		}
	}
	var mk uint8
	p := path[i]
	for _, j := range []int{i - 1, i + 1} {
		if j < 0 || j >= len(path) {
			continue
		}
		q := path[j]
		switch {
		case q.x > p.x:
			mk |= mE
		case q.x < p.x:
			mk |= mW
		case q.y > p.y:
			mk |= mS
		case q.y < p.y:
			mk |= mN
		}
	}
	if mk == mN || mk == mS {
		mk = mN | mS
	}
	if mk == mE || mk == mW || mk == 0 {
		mk = mE | mW
	}
	return linesHeavy[mk]
}

// ---- nodes -------------------------------------------------------------

// alarmOn is the catastrophe cascade: an alarm wave sweeps the graph from
// the resource bus outward, column by column, and repeats.
func (v *view) alarmOn(n *node) bool {
	if v.m.catast == "" || n == nil {
		return false
	}
	d := 0
	switch n.kind {
	case kPort:
		d = 0
	case kProducer, kService:
		d = 1
	case kHost:
		d = 2
	}
	// A pulse three frames wide leaves the bus and rolls outward.
	ph := ((v.frame-d*3)%12 + 12) % 12
	return ph < 3
}

func (v *view) borderClass(n *node) cls {
	switch {
	case v.alarmOn(n):
		return cNeg
	case n.id == v.sel:
		return cAccent
	case n.sev == sevCrit:
		return cNeg
	case n.sev == sevWarn:
		return cWarn
	}
	return cBorder
}

func (v *view) box(x, y, w, h int, k cls, set lineSet) {
	c := v.c
	for xx := x + 1; xx < x+w-1; xx++ {
		c.set(xx, y, set[mE|mW], k)
		c.set(xx, y+h-1, set[mE|mW], k)
	}
	for yy := y + 1; yy < y+h-1; yy++ {
		c.set(x, yy, set[mN|mS], k)
		c.set(x+w-1, yy, set[mN|mS], k)
	}
	c.set(x, y, set[mE|mS], k)
	c.set(x+w-1, y, set[mW|mS], k)
	c.set(x, y+h-1, set[mN|mE], k)
	c.set(x+w-1, y+h-1, set[mN|mW], k)
	for yy := y + 1; yy < y+h-1; yy++ {
		for xx := x + 1; xx < x+w-1; xx++ {
			c.set(xx, yy, ' ', cText)
		}
	}
	c.bgRect(x, y, w, h, bgSurface)
}

// title writes " title " into a top border with a right-aligned badge.
func (v *view) title(n *node, x, y, w int, k cls) {
	c := v.c
	t := n.title
	glyph := ""
	switch {
	case n.sev == sevCrit:
		glyph = string(v.s.alarm) + " "
	case n.sev == sevWarn:
		glyph = string(v.s.warn) + " "
	}
	badge := n.badge
	if d := v.delta(n); d != "" {
		badge = strings.TrimSpace(badge + " " + d)
	}
	room := w - 4
	bw := 0
	if badge != "" && room > width(t)+width(badge)+4 {
		bw = width(badge) + 2
	}
	tt := clip(glyph+t, room-bw)
	tk := cAccent
	if k == cNeg || k == cWarn {
		tk = k
	}
	c.set(x+1, y, ' ', k)
	c.textB(x+2, y, tt, tk, room-bw, true)
	c.set(x+2+width(tt), y, ' ', k)
	if bw > 0 {
		bx := x + w - 2 - bw
		c.set(bx, y, ' ', k)
		dx := bx + 1
		base := n.badge
		c.text(dx, y, base, cLabel, bw)
		if d := v.delta(n); d != "" {
			c.text(dx+width(base)+boolInt(base != ""), y, d, cHi, bw)
		}
		c.set(x+w-2, y, ' ', k)
	}
	if v.isNew(n) && w > width(tt)+12 {
		c.textB(x+3+width(tt), y, "NEW", cHi, 3, true)
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (v *view) drawNode(n *node) {
	c := v.c
	k := v.borderClass(n)
	if n.h == 1 {
		v.drawPill(n, k)
		return
	}
	v.box(n.x, n.y, n.w, n.h, k, v.s.boxes)
	v.title(n, n.x, n.y, n.w, k)
	if n.id == v.sel {
		c.bgRect(n.x, n.y, n.w, n.h, bgSel)
	}
	iw := n.w - 4
	ix := n.x + 2
	rows := n.h - 2
	if n == v.L.world {
		v.drawWorldList(n)
		return
	}
	var lines []func(y int)
	statLine := func(y int) {
		if n.alert != "" && n.kind != kService && n.kind != kHost {
			tag := string(v.s.warn) + " " + n.alert
			k := cWarn
			if n.sev == sevCrit {
				tag, k = string(v.s.alarm)+" "+n.alert, cNeg
			}
			c.textB(ix, y, tag, k, iw, true)
			sw := iw - width(tag) - 1
			if sw > 4 {
				c.text(ix+iw-width(n.stat), y, n.stat, cHi, width(n.stat))
			}
			return
		}
		st := n.stat
		sw := iw - width(st) - 1
		if len(n.spark) > 1 && sw >= 4 {
			sw = min(sw, 12)
			v.spark(ix, y, n.spark, sw, cPos)
		}
		c.textB(ix+iw-width(st), y, st, cHi, width(st), true)
	}
	gaugeLine := func(y int) {
		lbl := n.gaugeLabel
		gw := iw - width(lbl) - 1
		v.gauge(ix, y, gw, n.gauge, gaugeClass(n))
		c.text(ix+iw-width(lbl), y, lbl, cLabel, width(lbl))
	}
	subLine := func(y int) { c.text(ix, y, clip(n.sub, iw), cDim, iw) }
	alertLine := func(y int) {
		tag := string(v.s.warn) + " " + n.alert
		k := cWarn
		if n.sev == sevCrit {
			tag, k = string(v.s.alarm)+" "+n.alert, cNeg
		} else if n.sev == sevInfo {
			tag, k = "• "+n.alert, cHi
		}
		c.textB(ix, y, clip(tag, iw), k, iw, true)
	}
	switch n.kind {
	case kProducer:
		lines = append(lines, statLine)
		if n.gauge >= 0 {
			lines = append(lines, gaugeLine)
		}
		if n.sub != "" {
			lines = append(lines, subLine)
		}
	default:
		if n.gauge >= 0 {
			lines = append(lines, gaugeLine)
		}
		if n.alert != "" {
			lines = append(lines, alertLine)
		}
		if n.sub != "" {
			lines = append(lines, subLine)
		}
		if n.stat != "" {
			lines = append(lines, func(y int) { c.text(ix, y, clip(n.stat, iw), cLabel, iw) })
		}
	}
	for i := 0; i < rows && i < len(lines); i++ {
		lines[i](n.y + 1 + i)
	}
}

func gaugeClass(n *node) cls {
	switch n.sev {
	case sevWarn:
		return cWarn
	case sevCrit:
		return cNeg
	}
	return cPos
}

func (v *view) drawPill(n *node, k cls) {
	c := v.c
	x, y, w := n.x, n.y, n.w
	c.set(x, y, v.s.pillL, k)
	c.set(x+w-1, y, v.s.pillR, k)
	c.bgRect(x, y, w, 1, bgSurface)
	if n.id == v.sel {
		c.bgRect(x, y, w, 1, bgSel)
	}
	iw := w - 2
	t := n.title
	if n.sev >= sevWarn {
		g := v.s.warn
		if n.sev == sevCrit {
			g = v.s.alarm
		}
		t = string(g) + t
	}
	st := n.stat
	if n.alert != "" && n.kind == kHost {
		st = n.alert
	}
	tw := iw - width(st) - 1
	used := c.textB(x+1, y, clip(t, tw), cLabel, tw, n.sev >= sevWarn)
	if n.badge != "" && used+width(n.badge)+1 < tw {
		c.text(x+2+used, y, n.badge, cDim, tw-used-1)
	}
	sk := cHi
	if n.alert != "" && n.kind == kHost {
		sk = cNeg
	}
	c.text(x+1+iw-width(st), y, st, sk, width(st))
}

// drawChip draws the resource bus: one bordered component, a row per port.
func (v *view) drawChip() {
	c := v.c
	ch := v.L.chip
	k := cBorder
	for _, p := range v.L.ports {
		if v.alarmOn(p) {
			k = cNeg
		}
	}
	v.box(ch.x, ch.y, ch.w, ch.h, k, v.s.boxes)
	tn := &node{title: v.s.chip, badge: fmt.Sprintf("%d stores", len(v.L.ports)-1)}
	v.title(tn, ch.x, ch.y, ch.w, k)
	iw := ch.w - 3
	anyTag := false
	for _, p := range v.L.ports {
		if p.alert != "" {
			anyTag = true
		}
	}
	for _, p := range v.L.ports {
		y := p.y
		x := ch.x + 2
		if p.id == v.sel {
			c.bgRect(ch.x+1, y, ch.w-2, 1, bgSel)
		}
		nameW := 4
		for _, q := range v.L.ports {
			nameW = max(nameW, width(q.title)+1)
		}
		nameW = min(nameW, 10)
		if iw < 30 {
			nameW = min(nameW, 8)
		}
		nk := cLabel
		if p.id == "port:@pop" {
			nk = cAccent
		}
		c.textB(x, y, clip(p.title, nameW-1), nk, nameW-1, p.id == "port:@pop")
		rest := iw - nameW
		tagW := 0
		if anyTag {
			tagW = 6 // "EMPTY" plus a space
		}
		rateW := 8
		amtW := 0
		if rest-tagW-rateW >= 11 {
			amtW = 6
		}
		gw := rest - tagW - rateW - amtW - 1
		gx := x + nameW
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
		v.gauge(gx, y, gw, p.gauge, gk)
		ax := gx + gw + 1
		if amtW > 0 && p.id == "port:@pop" {
			ax += amtW
		} else if amtW > 0 {
			a := short(p.amount)
			c.text(ax+amtW-1-width(a), y, a, cText, amtW)
			ax += amtW
		}
		st := strings.TrimSuffix(p.stat, "/t")
		rk := cDim
		switch {
		case p.rate > 0:
			rk = cPos
		case p.rate < 0:
			rk = cNeg
		}
		if p.id == "port:@pop" {
			rk = cText
			w := rateW + amtW // the pop row has no separate amount
			c.text(ax+rateW-min(width(st), w), y, clip(st, w), rk, w)
		} else {
			c.text(ax+rateW-width(st), y, clip(st, rateW), rk, rateW)
		}
		if p.alert != "" {
			tag := p.alert
			tk := cWarn
			if p.sev == sevCrit {
				tk = cNeg
			} else if p.sev == sevInfo {
				tk = cHi
			}
			if width(tag) > tagW-1 {
				tag = clip(tag, tagW-1)
			}
			c.textB(ax+rateW+1, y, tag, tk, tagW-1, true)
		}
	}
	if v.L.quiet > 0 {
		c.text(ch.x+2, ch.y+ch.h-2, fmt.Sprintf("+%d quiet stores", v.L.quiet), cFaint, iw)
	}
}

func (v *view) drawWorldList(n *node) {
	c := v.c
	iw := n.w - 4
	y := n.y + 1
	for _, h := range v.m.cols[kHost] {
		if y >= n.y+n.h-1 {
			break
		}
		g, k := "·", cLabel
		st := h.gaugeLabel
		switch {
		case h.sev == sevCrit:
			g, k, st = string(v.s.sever), cNeg, h.alert
		case h.sev == sevWarn:
			g, k, st = string(v.s.warn), cWarn, h.alert
		case strings.HasPrefix(h.id, "host:route"):
			g, st = "≡", h.badge
		case h.id == "host:next":
			g = "→"
		}
		c.text(n.x+2, y, g, k, 1)
		tw := iw - 2 - width(st) - 1
		t := strings.TrimPrefix(h.title, "→ ")
		c.text(n.x+4, y, clip(t, tw), cText, tw)
		c.text(n.x+2+iw-width(st), y, st, k, width(st))
		y++
	}
}

// ---- small instruments --------------------------------------------------

func (v *view) gauge(x, y, w int, f float64, k cls) {
	if w <= 0 || f < 0 {
		return
	}
	c := v.c
	f = math.Max(0, math.Min(1, f))
	full := f * float64(w)
	for i := 0; i < w; i++ {
		switch {
		case float64(i+1) <= full:
			c.set(x+i, y, v.s.gaugeOn, k)
		case v.s.gaugeFrac != nil && float64(i) < full:
			part := full - float64(i)
			j := int(part * float64(len(v.s.gaugeFrac)))
			j = clampi(j, 0, len(v.s.gaugeFrac)-1)
			c.set(x+i, y, v.s.gaugeFrac[j], k)
		default:
			c.set(x+i, y, v.s.gaugeOff, cFaint)
		}
	}
}

func (v *view) spark(x, y int, data []float64, w int, k cls) {
	if len(data) == 0 || w <= 0 {
		return
	}
	if len(data) > w {
		data = data[len(data)-w:]
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, d := range data {
		lo = math.Min(lo, d)
		hi = math.Max(hi, d)
	}
	lv := v.s.spark
	// A steady rate (under 1% of swing) is a low steady line; otherwise the
	// range is padded below so the minimum still shows as a stub, not a hole.
	flat := hi-lo <= 0.01*math.Max(math.Abs(hi), 1e-9)
	lo -= (hi - lo) * 0.4
	for i, d := range data {
		j := 1
		if !flat {
			j = int(math.Round((d - lo) / (hi - lo) * float64(len(lv)-1)))
		}
		j = clampi(j, 0, len(lv)-1)
		v.c.set(x+w-len(data)+i, y, lv[j], k)
	}
}

// ---- diff since the last look ------------------------------------------

func (v *view) delta(n *node) string {
	if !v.diff || v.m.fx.Prev == nil || n.kind != kProducer {
		return ""
	}
	lin := strings.TrimPrefix(n.id, "prod:")
	now, was := lineageCount(v.m.st, lin), lineageCount(*v.m.fx.Prev, lin)
	if now > was && was > 0 {
		return fmt.Sprintf("+%d", now-was)
	}
	return ""
}

func (v *view) isNew(n *node) bool {
	if !v.diff || v.m.fx.Prev == nil {
		return false
	}
	p := v.m.fx.Prev
	switch {
	case strings.HasPrefix(n.id, "prod:"):
		return lineageCount(*p, strings.TrimPrefix(n.id, "prod:")) == 0
	case strings.HasPrefix(n.id, "host:civ:"):
		f, ok := p.Diplomacy.Factions[strings.TrimPrefix(n.id, "host:civ:")]
		return !ok || !f.Discovered
	case strings.HasPrefix(n.id, "host:route:"):
		k := strings.TrimPrefix(n.id, "host:route:")
		for _, r := range p.Trade.ActiveRoutes {
			if r.Key == k {
				return false
			}
		}
		return true
	}
	return false
}

func lineageCount(st game.GameState, lin string) int {
	n := 0
	for key, b := range st.Buildings {
		if d, ok := defs()[key]; ok && lineageOf(d) == lin {
			n += b.Count
		}
	}
	return n
}

// sinceLine summarises what changed since the last look.
func (v *view) sinceLine() string {
	p := v.m.fx.Prev
	if p == nil {
		return ""
	}
	st := v.m.st
	var parts []string
	if d := st.Stats.TotalBuilt - p.Stats.TotalBuilt; d > 0 {
		parts = append(parts, fmt.Sprintf("+%d built", d))
	}
	if d := st.Research.TotalResearched - p.Research.TotalResearched; d > 0 {
		parts = append(parts, fmt.Sprintf("+%d tech", d))
	}
	if d := st.Workers.TotalPop - p.Workers.TotalPop; d != 0 {
		parts = append(parts, fmt.Sprintf("%+d pop", d))
	}
	civs := 0
	for k, f := range st.Diplomacy.Factions {
		if f.Discovered && !p.Diplomacy.Factions[k].Discovered {
			civs++
		}
	}
	if civs > 0 {
		parts = append(parts, fmt.Sprintf("+%d civ", civs))
	}
	if len(parts) == 0 {
		return "quiet since last look"
	}
	mins := (st.Tick - p.Tick) * 2 / 60
	return fmt.Sprintf("last %dm: %s", mins, strings.Join(parts, " "))
}

// ---- chrome ---------------------------------------------------------------

func (v *view) drawHeader() {
	c := v.c
	st := v.m.st
	c.bgRect(0, 0, c.w, 1, bgChip)
	x := 1
	x += c.textB(x, 0, "◆ "+v.s.name, cAccent, c.w, true)
	sep := func() {
		x += c.text(x, 0, "  │  ", cFaint, 5)
	}
	sep()
	ep := config.EpochByKey()[st.EpochKey].Name
	x += c.textB(x, 0, st.AgeName, cText, c.w-x, true)
	x += c.text(x, 0, " · "+ep, cDim, c.w-x)
	right := fmt.Sprintf("tick %d", st.Tick)
	if v.diff {
		if s := v.sinceLine(); s != "" && v.L.w >= 110 {
			right = "Δ " + s + "   " + right
		}
	}
	rx := c.w - 1 - width(right)
	if rx > x+2 {
		c.text(rx, 0, right, cDim, width(right))
		if strings.HasPrefix(right, "Δ") {
			c.set(rx, 0, 'Δ', cHi)
		}
	}
	if v.L.top >= 2 {
		// A quiet legend row under the header.
		leg := v.legend()
		c.text(1, 1, clip(leg, c.w-2), cFaint, c.w-2)
	}
}

func (v *view) legend() string {
	s := v.s
	flow := string(s.head)
	if s.trail != 0 {
		flow = string(s.trail) + flow
	}
	if s.quantum {
		flow += "◇"
	}
	return fmt.Sprintf("%s flow   %c%c backlog   %c cut link   %c alarm   %c warning   rates per tick · wires run producer → store → use, imports run back",
		flow, s.head, s.head, s.sever, s.alarm, s.warn)
}

func (v *view) drawFooter() {
	c := v.c
	y := c.h - 1
	c.bgRect(0, y, c.w, 1, bgChip)
	keys := "←↑↓→ select  ⏎ inspect  w world  d diff  t theme  a age  q quit"
	if c.w < 130 {
		keys = "⏎ inspect  q quit"
	}
	if v.staged != "" {
		keys = "› " + v.staged
	}
	kx := c.w - 1 - width(keys)
	alarms := v.m.alarms
	x := 1
	if len(alarms) == 0 {
		c.text(x, y, "● all systems nominal", cPos, kx-x-1)
	} else {
		// A ticker: most severe first; it scrolls when it does not fit.
		tk := strings.Join(alarms, "  ·  ")
		room := kx - x - 3
		if room < 10 {
			room = c.w - 4
			kx = c.w
		}
		k := cWarn
		if v.m.catast != "" || strings.Contains(tk, "WAR") || strings.Contains(tk, "EMPTY") {
			k = cNeg
		}
		g := v.s.alarm
		if k == cNeg && v.m.catast != "" && v.frame%4 >= 2 {
			g = ' '
		}
		c.setB(x, y, g, k, true)
		txt := tk
		if width(tk) > room {
			loop := tk + "   ·   "
			rs := []rune(loop)
			off := max(0, v.frame-6) % len(rs) // hold, then scroll
			txt = string(append(rs[off:], rs[:off]...))
		}
		c.textB(x+2, y, clip(txt, room), k, room, true)
		if v.m.catast != "" {
			c.bgRect(0, y, kx-1, 1, bgAlarm)
		}
	}
	if kx > 20 {
		c.text(kx, y, keys, cDim, width(keys))
		if v.staged != "" {
			c.textB(kx, y, keys, cHi, width(keys), true)
		}
	}
}
