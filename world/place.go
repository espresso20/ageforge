package world

// place.go: the places the Galaxy opens into. A star the people hold has a
// Region of its own (its neighbours in the arm round it), a neighbourhood
// holding the star, and its generated system with the planets' close-ups under
// that, the same shape as the home chain, hung on the star's key.

// Near is the neighbourhood frame of star key with its system under it: the
// home one for Sol, nil for a key that is not a star.
func (w *World) Near(key string) *Frame {
	if key == "sol" {
		return w.Root.Find(FrameNeighbourhood)
	}
	sys := w.System(key)
	if sys == nil {
		return nil
	}
	return w.cached("near "+key, func() *Frame {
		f := &Frame{Kind: FrameNeighbourhood, Key: key + "_neighbourhood", Unit: UnitLy, Anchor: key, Radius: 14, Band: Band{Min: 3, Max: 60},
			Children: []*Frame{sys}}
		f.Objects = []Object{GenStar(w.Seed, mustID(key), Vec{}, true)}
		return f
	})
}

// Place is the region frame about star key with its chain under it: the home
// region for Sol, nil for a key that is not a star.
func (w *World) Place(key string) *Frame {
	if key == "sol" {
		return w.Root.Find(FrameRegion)
	}
	near := w.Near(key)
	if near == nil {
		return nil
	}
	return w.cached("place "+key, func() *Frame {
		id := mustID(key)
		f := &Frame{Kind: FrameRegion, Key: key + "_region", Unit: UnitLy, Anchor: key, Radius: ArmReachLy, Band: Band{Min: 400, Max: 18000},
			Children: []*Frame{near}}
		f.Objects = append([]Object{GenStar(w.Seed, id, Vec{}, true)}, placeStars(w.Seed, id, placeStarCount)...)
		knots(f, w.Seed+int64(id), 11)
		return f
	})
}

// placeStarCount is how many neighbours a region shows.
const placeStarCount = 8

func mustID(key string) int {
	id, _ := StarID(key)
	return id
}

func (w *World) cached(key string, build func() *Frame) *Frame {
	w.mu.Lock()
	defer w.mu.Unlock()
	if f := w.sys[key]; f != nil {
		return f
	}
	if w.sys == nil {
		w.sys = map[string]*Frame{}
	}
	f := build()
	w.sys[key] = f
	return f
}
