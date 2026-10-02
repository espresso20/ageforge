package mapmodel

// visitors.go schedules the rare visitor: from the Space Age on, now and
// then a saucer crosses the sky or hovers over the town, or a small figure
// walks a street, and is gone. Before the Space Age it is a long-odds joke.
// The schedule is a pure function of the save's seed, the age and the
// animation frame (the maps animate at about 8 frames a second on the wall
// clock), never of package math/rand, so captures and tests reproduce it.
// It never appears in a legend: the point is the surprise.

// SightingKind is how a visitor shows itself.
type SightingKind uint8

const (
	SightNone   SightingKind = iota
	SightFlyby               // a saucer crossing the sky
	SightHover               // a saucer hovering over the town, then gone
	SightWalker              // a small figure walking a street
)

// Sighting is one visit.
type Sighting struct {
	Kind   SightingKind
	Start  int    // first animation frame
	Frames int    // how long it lasts
	Roll   uint64 // a hash for the styles to place it with
}

// Phase is how far into the visit frame anim is, from 0 to 1.
func (s Sighting) Phase(anim int) float64 {
	if s.Frames <= 0 {
		return 0
	}
	return float64(anim-s.Start) / float64(s.Frames)
}

// The schedule: time is cut into blocks of SightingBlock frames (90
// minutes at 8 frames a second). From the Space Age every block holds one
// visit, placed in its middle stretch, so visits come about every 60 to 120
// minutes the map is open; before it, a block holds one with a chance of
// 150 in 1000 (one every 10 hours or so). A visit lasts 10 to 30 seconds.
const (
	SightingBlock   = 90 * 60 * 8
	sightingMin     = 10 * 8
	sightingMax     = 30 * 8
	sightingJokeIn  = 1000
	sightingJokeHit = 150
)

// SightingAt returns the visit showing at animation frame anim, if any.
// late is true from the Space Age on.
func SightingAt(seed int64, late bool, anim int) (Sighting, bool) {
	if anim < 0 {
		return Sighting{}, false
	}
	s, ok := sightingIn(seed, late, anim/SightingBlock)
	if !ok || anim < s.Start || anim >= s.Start+s.Frames {
		return Sighting{}, false
	}
	return s, true
}

// sightingIn is block b's visit, if it has one.
func sightingIn(seed int64, late bool, b int) (Sighting, bool) {
	const salt = 0x5a1e
	bb := int64(b)
	if !late && Hash(seed, salt, bb, 1)%sightingJokeIn >= sightingJokeHit {
		return Sighting{}, false
	}
	s := Sighting{Frames: sightingMin + int(Hash(seed, salt, bb, 2)%(sightingMax-sightingMin+1)),
		Roll: Hash(seed, salt, bb, 3)}
	lo, span := SightingBlock*15/100, SightingBlock*70/100-s.Frames
	s.Start = b*SightingBlock + lo + int(Hash(seed, salt, bb, 4)%uint64(span))
	switch k := Hash(seed, salt, bb, 5) % 20; {
	case k < 8:
		s.Kind = SightFlyby
	case k < 15:
		s.Kind = SightHover
	default:
		s.Kind = SightWalker
	}
	return s, true
}

// NextSighting returns the first visit that starts at or after frame anim
// (for captures and tests that want to look at one).
func NextSighting(seed int64, late bool, anim int) Sighting {
	for b := max(0, anim/SightingBlock); ; b++ {
		if s, ok := sightingIn(seed, late, b); ok && s.Start >= anim {
			return s
		}
	}
}

// SpaceAge is the index of the age the visitors come with, or -1.
func (c *Catalog) SpaceAge() int {
	if i, ok := c.AgeIdx["space_age"]; ok {
		return i
	}
	return -1
}

// SightingAt is the model's visit at animation frame anim, if any.
func (m *Model) SightingAt(anim int) (Sighting, bool) {
	late := false
	if m.Catalog != nil {
		sa := m.Catalog.SpaceAge()
		late = sa >= 0 && m.AgeIdx >= sa
	}
	return SightingAt(m.Seed, late, anim)
}
