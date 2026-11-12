package world

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
)

// quantum.go: what the Quantum Age opens. The Local Group's galaxies can be
// entered, each with its own shape and stars made from the save's seed, and
// everything made up there is uncertain: it has two or three candidate
// versions and shows them in turn until the player settles it. Which version
// a settled place keeps follows from the seed and the place's id alone.

// UncertainLabel is the label key of a place that is not settled: the window
// writes it by the object instead of its name.
const UncertainLabel = "map_uncertain"

// galStride is the room each galaxy's star ids have: the Milky Way's are below
// it, galaxy g's from g*galStride, so no two galaxies share a key.
const galStride = 1_000_000

// galStars is how many stars a galaxy shows when entered.
const galStars = 24

// galBand is the zoom band of an entered galaxy (the home galaxy's).
var galBand = Band{Min: 15000, Max: 240000}

// galIndex numbers the Local Group galaxies that open as galaxies of their own, or -1: the Milky Way and its satellites do not.
func galIndex(key string) int {
	switch key {
	case "andromeda":
		return 1
	case "triangulum":
		return 2
	} // the Magellanic Clouds stay the Milky Way's satellites: the galaxy frame draws them close up

	if n, err := strconv.Atoi(strings.TrimPrefix(key, "far_")); err == nil && strings.HasPrefix(key, "far_") {
		return 3 + n
	}
	return -1
}

// doubtOf is how many versions star id has: none inside the Milky Way, two or
// three in any other galaxy.
func doubtOf(seed int64, id int) uint8 {
	if id < galStride {
		return 0
	}
	return uint8(2 + Roll(2, seed, 9600, int64(id)))
}

// vseed is the seed of version n of the place called key.
func vseed(seed int64, key string, n int) int64 {
	return int64(mapmodel.Hash(seed, mapmodel.HashStr(key), int64(n)))
}

// PickVersion is the version a settled place keeps.
func PickVersion(seed int64, o *Object) int {
	if o.Doubt == 0 {
		return 0
	}
	return Roll(int(o.Doubt), seed, 9601, mapmodel.HashStr(o.Key))
}

// Version is the object as version n of it looks. Version 0 is the object as
// it stands; the others re-roll its look from the seed and its key.
func Version(seed int64, o *Object, n int) Object {
	c := *o
	if n <= 0 || o.Doubt == 0 {
		return c
	}
	switch {
	case o.Kind == KindStar && o.Star != nil:
		id, _ := StarID(o.Key)
		l := starLookV(seed, id, n)
		if o.Weight < 200 { // seen from outside, a star keeps its dimness
			l.Weight = 120
		}
		c.Star, c.Tone, c.Weight = &l, l.Tone, uint8(l.Weight)
	case o.Shape != nil:
		s := PickShape(vseed(seed, o.Key, n))
		c.Shape, c.Seed = &s, vseed(seed, o.Key, n)
	case o.Galaxy != nil:
		l := *o.Galaxy
		l.Phi = Roll(Turn/2, seed, 9602, mapmodel.HashStr(o.Key), int64(n))
		if !l.Irr && l.Arms > 0 {
			l.Arms, l.Spin = 2+n%2, -l.Spin
		} else if !l.Irr {
			l.Arms, l.Spin, l.Core = 2, 256, 120 // an elliptical, seen as a spiral
		}
		c.Galaxy, c.Seed = &l, vseed(seed, o.Key, n)
	}
	return c
}

// openLocalGroup makes the Local Group's galaxies other than the Milky Way
// places to enter: pickable, uncertain, the two big spirals drawn from a
// seeded shape like the home galaxy.
func openLocalGroup(f *Frame, seed int64) {
	for i := range f.Objects {
		o := &f.Objects[i]
		gi := galIndex(o.Key)
		if o.Kind != KindGalaxy || gi < 0 {
			continue
		}
		o.Pick, o.Doubt = true, uint8(2+Roll(2, seed, 9603, int64(gi)))
		if gi > 2 { // a far galaxy has no name of its own until it is settled
			k := StarKey(seed, gi)
			o.Label, o.Inspect = k, k
		}
		if gi <= 2 { // the two big spirals have arms the seed winds
			s := PickShape(vseed(seed, o.Key, 0))
			o.Shape, o.Seed, o.Galaxy = &s, vseed(seed, o.Key, 0), nil
		}
	}
}

// Galaxy is the frame of the Local Group galaxy key with its first star's
// place under it: the home galaxy's frame for the Milky Way, nil for anything
// that cannot be entered (the Local Group is not open, or key is no galaxy).
func (w *World) Galaxy(key string) *Frame {
	if w.Root == nil {
		return nil
	}
	if key == "milky_way" {
		return w.Root.Find(FrameGalaxy)
	}
	if w.Root.Kind != FrameLocalGroup || galIndex(key) < 0 {
		return nil
	}
	var o *Object
	for i := range w.Root.Objects {
		if c := &w.Root.Objects[i]; c.Key == key && c.Kind == KindGalaxy && c.Doubt > 0 {
			o = c
		}
	}
	if o == nil {
		return nil
	}
	base := galIndex(key) * galStride
	ids := starIDs(w.Seed, base, base, galStars)
	place := w.Place(StarKey(w.Seed, ids[0])) // before cached: Place takes the lock itself
	return w.cached("galaxy "+key, func() *Frame {
		f := &Frame{Kind: FrameGalaxy, Key: key, Unit: UnitLy, Anchor: key, Pos: o.Pos, Radius: galRadiusLy, Band: galBand,
			Children: []*Frame{place}}
		disc := *o
		disc.Pos, disc.Size, disc.Weight, disc.Pick = Vec{}, galRadiusLy, 255, false
		f.Objects = append(f.Objects, disc)
		for j, id := range ids {
			a := Roll(Turn, w.Seed, 9604, int64(id))
			r := galRadiusLy * float64(15+Roll(75, w.Seed, 9605, int64(id), int64(j))) / 100
			f.Objects = append(f.Objects, GenStar(w.Seed, id, polar(r, a), true))
		}
		return f
	})
}
