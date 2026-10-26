package mapgen

// Detail scatter: one candidate spot per small block of the world, jittered
// inside it by hash, so trees and rocks sit apart without lining up. Which
// spots get anything depends on the biome and the age's scatter rows.
const (
	spotW, spotH = 4, 2
)

func (f *frame) paintDetail() {
	if f.st.Scene == Orbit {
		return
	}
	seed := f.m.Seed
	for by := 0; by < worldH/spotH; by++ {
		for bx := 0; bx < worldW/spotW; bx++ {
			x := bx*spotW + roll(spotW, seed, 60, int64(bx), int64(by))
			y := by*spotH + roll(spotH, seed, 61, int64(bx), int64(by))
			if f.busy[at(x, y)] {
				continue
			}
			biome := f.fields.Biome[at(x, y)]
			for _, s := range f.st.Scatter {
				if s.On == biome && chance(s.Chance, seed, 62, int64(x), int64(y)) {
					f.g.draw(x, y, pick(s.Glyphs, hash(seed, 63, int64(x), int64(y))), s.Fg)
					break
				}
			}
		}
	}
}

// paintSmoke lets each chimney breathe out a puff that climbs and drifts. Its
// place is a function of the frame and the chimney alone.
func (f *frame) paintSmoke() {
	for _, v := range f.vents {
		beat := (f.anim/5 + roll(9, f.m.Seed, 64, int64(v.X), int64(v.Y))) % 9
		if beat >= 3 {
			continue
		}
		x, y := v.X+beat/2*(2*roll(2, f.m.Seed, 65, int64(v.X), int64(v.Y))-1), v.Y-1-beat
		if !f.g.in(x, y) || f.g.ThingAt(x, y).Kind == KindBuilding {
			continue
		}
		f.g.draw(x, y, pick(f.st.Smoke.Glyphs, hash(f.m.Seed, 66, int64(beat))), mix(f.g.At(x, y).Bg, f.st.Smoke.Fg, 256-beat*60))
	}
}
