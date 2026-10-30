package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The Primitive Age's ambient life. The model's schedule
// (mapmodel.Model.Dinos and AsteroidAt) says what is about; this layer says
// where. Everything is a function of the sighting and the animation frame
// alone, so any frame draws without the ones before it, and with the motion
// setting off whatever is on screen holds still. None of it is marked as a
// thing on the grid: nothing here can be inspected or clicked.
//
// The walkers keep to open ground (land that is not water or mountain, with
// nothing built on it) and the pterodactyl is not drawn over a building
// either. All of it is drawn in Unicode shapes that fold to ASCII.

// sprite is one dinosaur's look: rows of glyphs, ' ' being clear.
type sprite struct {
	Rows []string
	Fg   Paint
}

// The three of them, facing east (right). west mirrors one to face west.
var (
	sauropod = sprite{Rows: []string{"  ◢█◣╱•", " ◢████ "}, Fg: P(HBrick, 255)}
	raptor   = sprite{Rows: []string{"◣█•"}, Fg: P(HRed, 255)}
	// the pterodactyl flaps between two frames and faces neither way
	pterodactyl = [2]sprite{{Rows: []string{"╲•╱"}, Fg: P(HAmber, 255)}, {Rows: []string{"─•─"}, Fg: P(HAmber, 255)}}
)

// The colours are picked to stand apart from the greens of the forest, grass
// and scrub they walk in (a rust brown, a red and an amber), not to match
// them; a test holds it on every theme with colour. DinoHues lists them.
func DinoHues() []Hue {
	return []Hue{sauropod.Fg.Hue, raptor.Fg.Hue, pterodactyl[0].Fg.Hue}
}

// mirror swaps the glyphs that have a handedness.
var mirror = map[rune]rune{'◢': '◣', '◣': '◢', '◤': '◥', '◥': '◤', '╱': '╲', '╲': '╱', '▌': '▐', '▐': '▌', '►': '◄', '◄': '►'}

// west is the sprite turned to face west.
func (s sprite) west() sprite {
	out := sprite{Fg: s.Fg}
	for _, row := range s.Rows {
		rs := []rune(row)
		for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
			rs[i], rs[j] = rs[j], rs[i]
		}
		for i, r := range rs {
			if m, ok := mirror[r]; ok {
				rs[i] = m
			}
		}
		out.Rows = append(out.Rows, string(rs))
	}
	return out
}

// width is the widest row, in cells.
func (s sprite) width() int {
	w := 0
	for _, row := range s.Rows {
		w = max(w, len([]rune(row)))
	}
	return w
}

// paintDinos draws the dinosaurs that are about at the frame, and the
// asteroid.
func (f *frame) paintDinos() {
	if f.st.Scale > City {
		return
	}
	if f.clock > 0 && f.clock != f.anim {
		// held still: whoever is on the map on the world's clock stands
		// halfway through their walk
		for _, sg := range f.m.Dinos(f.clock) {
			f.paintDino(sg, sg.Frames/2)
		}
	} else {
		for _, sg := range f.m.Dinos(f.anim) {
			f.paintDino(sg, f.anim-sg.Start)
		}
	}
	if sg, ok := f.m.AsteroidAt(f.anim); ok {
		f.paintAsteroid(sg)
	}
}

// dinoPace says how a walker moves: a cell in every slow frames of per cells
// .
type dinoPace struct{ cells, frames int }

// paintDino draws one dinosaur at the frame.
func (f *frame) paintDino(sg mapmodel.Sighting, t int) {
	east := sg.Roll>>8&1 == 0
	switch sg.Kind {
	case mapmodel.SightSauropod:
		f.paintWalker(sg, sauropod, dinoPace{1, 24}, t, east)
	case mapmodel.SightRaptor:
		f.paintWalker(sg, raptor, dinoPace{1, 12}, t, east)
	case mapmodel.SightPterodactyl:
		f.paintFlyer(sg, t, east)
	}
}

// walkable says whether a dinosaur may stand on a cell: dry open land with
// no building, wonder or scenery on it. A road or someone on it is open.
func (f *frame) walkable(x, y int) bool {
	if !inWorld(x, y) {
		return false
	}
	switch f.g.ThingAt(x, y).Kind {
	case KindTerrain, KindRoad, KindAgent:
	default:
		return false
	}
	b := f.fields.Biome[at(x, y)]
	return !b.Water() && b != Mountain
}

// paintWalker walks a sprite along a row of open ground: the sweep it makes
// from where it starts to where it ends is picked, among the places the
// visit's roll offers near the square, as the first that is open all along;
// when the town leaves none, it does not come.
func (f *frame) paintWalker(sg mapmodel.Sighting, s sprite, pace dinoPace, t int, east bool) {
	if !east {
		s = s.west()
	}
	w, h := s.width(), len(s.Rows)
	run := sg.Frames * pace.cells / pace.frames // the cells it walks in all
	steps := t * pace.cells / pace.frames
	c := f.plan.Centre
	for try := int64(0); try < 120; try++ {
		x0 := c.X - 65 + roll(130-run, int64(sg.Roll), 20, try)
		y0 := c.Y - 10 + roll(21, int64(sg.Roll), 21, try)
		open := true
		for y := y0; y < y0+h && open; y++ {
			for x := x0; x < x0+run+w; x++ {
				if !f.walkable(x, y) {
					open = false
					break
				}
			}
		}
		if !open {
			continue
		}
		x := x0 + steps
		if !east {
			x = x0 + run - steps
		}
		f.drawSprite(Pt{x, y0}, s)
		return
	}
}

// paintFlyer flies the pterodactyl across the upper rows of the town, flapping
// every few frames. It is not drawn on a frame where it would cover a building.
func (f *frame) paintFlyer(sg mapmodel.Sighting, t int, east bool) {
	const reach = 60
	c := f.plan.Centre
	x := c.X - reach + t*2*reach/sg.Frames
	if !east {
		x = c.X + reach - t*2*reach/sg.Frames
	}
	y := c.Y - 14 + roll(5, int64(sg.Roll), 22) + t/10%2
	f.drawSprite(Pt{x - 1, y}, pterodactyl[t/3%2])
}

// drawSprite draws a sprite with its top-left corner at p, all or nothing: a
// sprite that would run off the map or over a building is not drawn.
func (f *frame) drawSprite(p Pt, s sprite) {
	for dy, row := range s.Rows {
		for dx, r := range []rune(row) {
			if r != ' ' && (!inWorld(p.X+dx, p.Y+dy) || f.solid(p.X+dx, p.Y+dy)) {
				return
			}
		}
	}
	for dy, row := range s.Rows {
		for dx, r := range []rune(row) {
			if r != ' ' {
				f.g.draw(p.X+dx, p.Y+dy, r, s.Fg)
			}
		}
	}
}

// paintAsteroid streaks the fireball down across the sky over the town: a
// white head and a tail that goes from orange to red as it fades. The ground
// map has no sky of its own, so it runs over the rows above the square and
// down toward it, behind the buildings.
func (f *frame) paintAsteroid(sg mapmodel.Sighting) {
	c := f.plan.Centre
	x0, y0 := c.X-30+int(sg.Roll>>4%61), c.Y-19
	vx := 3
	if sg.Roll&1 == 1 {
		vx = -3
	}
	f.streak(x0, y0, vx, 2, f.anim-sg.Start, 9,
		func(x, y int) bool { return inWorld(x, y) && !f.solid(x, y) },
		func(k int) (rune, Paint) {
			switch {
			case k == 0:
				return '✦', P(HWhite, 255)
			case k == 1:
				return '●', P(HAmber, 255)
			case k < 4:
				return '∙', P(HFire, 240)
			case k < 7:
				return '·', P(HRed, 200)
			}
			return '·', P(HRed, 120)
		})
}
