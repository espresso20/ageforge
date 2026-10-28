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
	road, length := f.plan.Roads[index], f.shown[index]
	span := 2 * (length - 1)
	pos := (roll(span, seed, 71, int64(kind), int64(id)) + f.anim/max(k.Slow, 1)) % span
	forward := pos < length
	i := pos
	if !forward {
		i = span - pos
	}
	from, to := road.Cells[i], road.Cells[min(i+1, length-1)]
	if !forward {
		to = road.Cells[max(i-1, 0)]
	}
	f.g.draw(from.X, from.Y, heading(k.Glyphs, to.X-from.X, to.Y-from.Y), k.Fg)
	f.g.mark(from.X, from.Y, Thing{Kind: KindAgent, Key: k.Name})
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
