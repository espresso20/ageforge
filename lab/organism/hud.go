package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// labels: each limb is named at its outermost tip, biggest limbs first,
// as long as there is room. The label is the lineage, so the tree reads in
// monochrome.
func (r *renderer) labels() {
	type cand struct {
		lg  *limbGeom
		cnt int
	}
	var cs []cand
	for i := range r.limbs {
		cs = append(cs, cand{&r.limbs[i], r.o.Limbs[r.limbs[i].li].Count})
	}
	sort.SliceStable(cs, func(a, b int) bool { return cs[a].cnt > cs[b].cnt })
	limit := len(cs)
	if r.mini {
		limit = min(3, limit)
	}
	for i, cd := range cs {
		if i >= limit || !cd.lg.hasTip {
			continue
		}
		if r.v.Sel >= 0 && !cd.lg.isSel && i >= 3 {
			continue
		}
		l := r.o.Limbs[cd.lg.li]
		txt := fmt.Sprintf("%c %s %d", l.Spec.Leaf, l.Spec.Label, l.Count)
		if r.mini {
			txt = l.Spec.Label
		}
		n := len([]rune(txt))
		tx, ty := int(cd.lg.tip[0]), int(cd.lg.tip[1])
		right := cd.lg.tip[0] >= r.cx
		col := cd.lg.col
		if cd.lg.dimmed {
			col = theme.Mix(col, r.p.Bg, 0.4)
		}
		// Beside the tip's leaf cluster, then above or below it.
		ry := int(math.Ceil(math.Max(0.6, r.uy/15)))
		rx := int(math.Ceil(math.Max(0.6, r.uy/15) * 2.2))
		var spots [][2]int
		if right {
			spots = [][2]int{{tx + rx + 2, ty}, {tx + rx + 2, ty - 1}, {tx + rx + 2, ty + 1}, {tx - n/2, ty - ry - 1}, {tx + rx + 1, ty - ry - 1}, {tx + rx + 4, ty + 2}}
		} else {
			spots = [][2]int{{tx - rx - n - 1, ty}, {tx - rx - n - 1, ty - 1}, {tx - rx - n - 1, ty + 1}, {tx - n/2, ty - ry - 1}, {tx - rx - n, ty - ry - 1}, {tx - rx - n - 3, ty + 2}}
		}
		for _, sp := range spots {
			x := max(0, min(r.W-n, sp[0]))
			y := sp[1]
			if y < r.top || y >= r.ground {
				continue
			}
			if r.c.Free(x-1, y, n+2, 3) {
				att := tcell.AttrNone
				if cd.lg.isSel || i < 3 {
					att = tcell.AttrBold
				}
				r.c.Text(x, y, txt, col, 8, att)
				break
			}
		}
	}
}

// hud: one line on top, a lineage spectrum and a ticker at the bottom.
func (r *renderer) hud() {
	c, p, o := r.c, r.p, r.o
	for x := 0; x < r.W; x++ {
		c.Bg(x, 0, p.Surface)
		c.Glyph(x, 0, ' ', p.Text, 9) // the scene never draws over the header
	}
	x := c.Text(1, 0, o.AgeName, p.Accent, 9, tcell.AttrBold)
	if !r.mini {
		x = c.Text(x, 0, " · "+o.Epoch, p.Dim, 9, 0)
	}
	type stat struct {
		s   string
		col tcell.Color
	}
	idle := p.Dim
	if o.Idle > 0 {
		idle = p.Warning
	}
	food := p.Positive
	if o.FoodRate < 0 {
		food = p.Negative
	}
	var stats []stat
	if r.mini {
		stats = []stat{{fmt.Sprintf("pop %s", short(o.Pop)), p.Text}, {fmt.Sprintf("%d▲", o.LimbTotal()), p.Text}}
		if o.Idle > 0 {
			stats = append(stats, stat{fmt.Sprintf("%d idle", o.Idle), idle})
		}
	} else {
		stats = []stat{
			{fmt.Sprintf("pop %s/%s", short(o.Pop), short(o.MaxPop)), p.Text},
			{fmt.Sprintf("idle %d", o.Idle), idle},
			{"food " + rate(o.FoodRate), food},
			{fmt.Sprintf("morale %d%%", int(o.Morale*100+0.5)), p.Text},
			{fmt.Sprintf("%d built", o.Built), p.Text},
			{fmt.Sprintf("%d techs", o.Techs), p.Text},
		}
		if o.AgeReady && o.NextAgeName != "" {
			stats = append([]stat{{"ready: " + strings.ToLower(o.NextAgeName), p.Positive}}, stats...)
		}
	}
	w := 0
	for i, s := range stats {
		w += len([]rune(s.s))
		if i > 0 {
			w += 3
		}
	}
	// Stats are in priority order: drop from the end until the line fits.
	for len(stats) > 1 && x+2+w > r.W-1 {
		w -= len([]rune(stats[len(stats)-1].s)) + 3
		stats = stats[:len(stats)-1]
	}
	sx := r.W - 1 - w
	for i, s := range stats {
		if i > 0 {
			sx = c.Text(sx, 0, " · ", p.Dim, 9, 0)
		}
		sx = c.Text(sx, 0, s.s, s.col, 9, 0)
	}
	if r.mini {
		return
	}

	// Spectrum: the crown's make-up as one proportional rule, in crown
	// order, each lineage named where its share has room.
	y := r.H - 2
	total := o.LimbTotal()
	if total > 0 {
		x := 0
		acc := 0
		for i, l := range o.Limbs {
			acc += l.Count
			end := int(math.Round(float64(acc) / float64(total) * float64(r.W)))
			if i == len(o.Limbs)-1 {
				end = r.W
			}
			col := r.p.Limb[l.Index]
			if r.v.Sel >= 0 && o.Limbs[r.v.Sel].Index != l.Index {
				col = theme.Mix(col, p.Bg, 0.55)
			}
			for xx := x; xx < end; xx++ {
				c.Glyph(xx, y, '━', col, 9)
			}
			lbl := fmt.Sprintf(" %c %s %d ", l.Spec.Leaf, l.Spec.Label, l.Count)
			if end-x >= len([]rune(lbl))+2 {
				c.Text(x+1, y, lbl, col, 9, tcell.AttrBold)
			} else if end-x >= len(l.Spec.Label)+4 {
				c.Text(x+1, y, " "+l.Spec.Label+" ", col, 9, tcell.AttrBold)
			}
			x = end
		}
	}

	// Ticker: what changed since the last visit, else the newest headline.
	y = r.H - 1
	tx := c.Text(1, y, "» ", p.Accent, 9, tcell.AttrBold)
	var parts []string
	if r.prev != nil && r.v.Since {
		prev := map[int]int{}
		for _, l := range r.prev.Limbs {
			prev[l.Index] = l.Count
		}
		type delta struct {
			label string
			d     int
		}
		var ds []delta
		for _, l := range o.Limbs {
			if d := l.Count - prev[l.Index]; d != 0 {
				ds = append(ds, delta{l.Spec.Label, d})
			}
		}
		sort.SliceStable(ds, func(a, b int) bool { return ds[a].d > ds[b].d })
		var bits []string
		for _, d := range ds {
			bits = append(bits, fmt.Sprintf("%+d %s", d.d, d.label))
		}
		if len(bits) > 0 {
			parts = append(parts, "since your last visit: "+strings.Join(bits, " "))
		}
		if dp := o.Pop - r.prev.Pop; dp != 0 {
			parts = append(parts, fmt.Sprintf("%+d pop", dp))
		}
	}
	if len(o.Events) > 0 {
		parts = append(parts, "✧ "+strings.Join(o.Events, ", "))
	}
	if o.PendingCatastrophe != "" {
		parts = append([]string{"☄ catastrophe pending: type catastrophe to endure or succumb"}, parts...)
	}
	if len(parts) == 0 && len(o.Headlines) > 0 {
		parts = append(parts, o.Headlines[0].Message)
	}
	line := strings.Join(parts, " · ")
	if max(0, r.W-tx-1) < len([]rune(line)) {
		line = string([]rune(line)[:max(0, r.W-tx-2)]) + "…"
	}
	c.Text(tx, y, line, p.Text, 9, 0)
}

// inspect: the verb. The selected limb comes forward, the rest recede, and
// a card names its buildings by the keys the command line takes.
func (r *renderer) inspect() {
	c, p := r.c, r.p
	l := r.o.Limbs[r.v.Sel]
	col := r.p.Limb[l.Index]
	lines := []struct {
		s   string
		col tcell.Color
	}{
		{fmt.Sprintf("%d buildings · tier %d", l.Count, l.Tier), p.Text},
	}
	if l.WorkerCap > 0 {
		wc := p.Text
		if l.Vigor() < 0.8 {
			wc = p.Warning
		}
		lines = append(lines, struct {
			s   string
			col tcell.Color
		}{fmt.Sprintf("workers %d/%d (%d%%)", l.Workers, l.WorkerCap, int(l.Vigor()*100+0.5)), wc})
	}
	if l.Rate != 0 {
		lines = append(lines, struct {
			s   string
			col tcell.Color
		}{"output " + rate(l.Rate), p.Positive})
	}
	for i, b := range l.Buildings {
		if i >= 4 {
			lines = append(lines, struct {
				s   string
				col tcell.Color
			}{fmt.Sprintf("… %d more kinds", len(l.Buildings)-4), p.Dim})
			break
		}
		lines = append(lines, struct {
			s   string
			col tcell.Color
		}{fmt.Sprintf("%3d× build %s", b.Count, b.Key), p.Dim})
	}
	w := len(l.Spec.Label) + 6
	for _, ln := range lines {
		w = max(w, len([]rune(ln.s))+4)
	}
	h := len(lines) + 2
	x0, y0 := 1, r.top+1
	if r.mini {
		x0, y0 = 0, r.top
	}
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			c.Bg(x, y, p.Surface)
			c.Glyph(x, y, ' ', p.Text, 10)
		}
	}
	for x := x0; x < x0+w; x++ {
		c.Glyph(x, y0, '─', p.Border, 10)
		c.Glyph(x, y0+h-1, '─', p.Border, 10)
	}
	for y := y0; y < y0+h; y++ {
		c.Glyph(x0, y, '│', p.Border, 10)
		c.Glyph(x0+w-1, y, '│', p.Border, 10)
	}
	c.Glyph(x0, y0, '╭', p.Border, 10)
	c.Glyph(x0+w-1, y0, '╮', p.Border, 10)
	c.Glyph(x0, y0+h-1, '╰', p.Border, 10)
	c.Glyph(x0+w-1, y0+h-1, '╯', p.Border, 10)
	c.Text(x0+2, y0, " "+l.Spec.Label+" ", col, 10, tcell.AttrBold)
	for i, ln := range lines {
		c.Text(x0+2, y0+1+i, ln.s, ln.col, 10, 0)
	}
}

func short(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 10_000:
		return fmt.Sprintf("%.1fk", float64(n)/1e3)
	}
	return fmt.Sprint(n)
}

// rate formats a per-second rate compactly: +12.4/s, +2.6k/s, -1.1M/s.
func rate(v float64) string {
	a := math.Abs(v)
	switch {
	case a >= 1e9:
		return fmt.Sprintf("%+.1fG/s", v/1e9)
	case a >= 1e6:
		return fmt.Sprintf("%+.1fM/s", v/1e6)
	case a >= 1e4:
		return fmt.Sprintf("%+.1fk/s", v/1e3)
	}
	return fmt.Sprintf("%+.1f/s", v)
}
