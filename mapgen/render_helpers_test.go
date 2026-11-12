package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Only the tests cut a window out of the world; the game draws whole scenes.

// Render draws the model as a w by h grid at animation frame anim. The
// picture is a window onto a world of fixed size, centred on the settlement.
func (gen *Generator) Render(m *mapmodel.Model, w, h, anim int) *Grid {
	f := gen.frame(m, anim)
	f.draw()
	x0, y0 := f.camera(w, h)
	return f.g.Window(x0, y0, w, h)
}

// camera is where the window's top-left corner goes: centred on the
// settlement, nudged a little (the nearest nudge that does best) to take in
// the most of what is built: the wonder being built first, then the rest.
func (f *frame) camera(w, h int) (x0, y0 int) {
	nudges := [][2]int{{0, 0}, {-8, 0}, {8, 0}, {0, -4}, {0, 4}, {-8, -4}, {8, -4}, {-8, 4}, {8, 4}, {-16, 0}, {16, 0}, {-16, -4}, {16, -4}, {-16, 4}, {16, 4}, {-24, 0}, {24, 0}}
	if f.st.Scale == Orbit || f.st.Scale == System {
		nudges = [][2]int{{0, 0}, {0, -12}, {0, -16}, {0, 12}, {0, 16}}
	} else if f.st.Scale > City {
		nudges = nudges[:1]
	}
	best := -1
	for _, n := range nudges {
		x := clamp(f.plan.Centre.X+n[0]-w/2, 0, f.g.W-w)
		y := clamp(f.plan.Centre.Y+n[1]-h/2, 0, f.g.H-h)
		score := f.inView(Rect{x, y, w, h})
		if score > best {
			if (f.st.Scale == Orbit || f.st.Scale == System) && best > 0 && n != [2]int{0, 0} {
				continue
			}
			x0, y0, best = x, y, score
		}
	}
	return x0, y0
}

// inView scores a window by what it takes in whole: one for each building,
// four for a wonder, a thousand for the wonder being built.
func (f *frame) inView(window Rect) int {
	inside := func(r Rect) bool {
		return r.X >= window.X && r.X+r.W <= window.X+window.W && r.Y >= window.Y && r.Y+r.H <= window.Y+window.H
	}
	score := 0
	for _, pl := range f.placed {
		if inside(f.plan.Lots[pl.lot].Rect) {
			score++
		}
	}
	for _, ws := range f.wonders() {
		switch {
		case inside(ws.plot.Rect) && ws.w.Current:
			score += 1000
		case inside(ws.plot.Rect):
			score += 4
		}
	}
	return score
}
