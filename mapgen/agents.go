package mapgen

import "unicode/utf8"

// Agents are the traffic and the walkers on the roads. An agent has no
// stored state: where it is at a frame is a pure function of its number and
// the frame, so any frame draws without the ones before it. Each walks one
// laid road out to its end and back.

func (f *frame) paintAgents() {
	var routes []int
	for i, n := range f.shown {
		if n >= 4 {
			routes = append(routes, i)
		}
	}
	if len(routes) == 0 {
		return
	}
	pop := f.m.Workers.Pop
	for kind, k := range f.st.Agents {
		for id := 0; id < min(k.Max, pop/k.Per); id++ {
			f.paintAgent(k, kind, id, routes)
		}
	}
}

func (f *frame) paintAgent(k AgentKind, kind, id int, routes []int) {
	seed := f.m.Seed
	index := routes[roll(len(routes), seed, 70, int64(kind), int64(id))]
	cells := f.plan.Roads[index].Cells[:f.shown[index]]
	offset := roll(2*(len(cells)-1), seed, 71, int64(kind), int64(id))
	p, dx, dy := pace(cells, offset, f.anim, k.Slow)
	f.g.draw(p.X, p.Y, heading(k.Glyphs, dx, dy), k.Fg)
	f.g.mark(p.X, p.Y, Thing{Kind: KindAgent, Key: k.Name})
}

// pace is where a figure going back and forth along cells is at a frame, and
// the way it is heading: a cell every slow frames, from the cell its number
// (offset) puts it at, out to the end and back.
func pace(cells []Pt, offset, anim, slow int) (at Pt, dx, dy int) {
	span := 2 * (len(cells) - 1)
	pos := (offset + anim/max(slow, 1)) % span
	forward := pos < len(cells)
	i, next := pos, min(pos+1, len(cells)-1)
	if !forward {
		i = span - pos
		next = max(i-1, 0)
	}
	return cells[i], cells[next].X - cells[i].X, cells[next].Y - cells[i].Y
}

// heading picks the glyph for travelling by (dx, dy): a set of four is
// east, west, south, north; anything shorter serves every way.
func heading(glyphs string, dx, dy int) rune {
	runes := []rune(glyphs)
	if len(runes) < 4 || utf8.RuneCountInString(glyphs) < 4 {
		return runes[0]
	}
	switch {
	case dx > 0:
		return runes[0]
	case dx < 0:
		return runes[1]
	case dy > 0:
		return runes[2]
	}
	return runes[3]
}
