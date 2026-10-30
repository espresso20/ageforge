package mapmodel

import (
	"sync"

	"github.com/espresso20/ageforge/config"
)

// dinos.go schedules the Primitive Age's ambient life: dinosaurs that roam
// the town map and, near the end of the Late Cretaceous, the asteroid. Like
// the visitors they are pure functions of the save's seed, the age and the
// animation frame (plus the gate share, which stands for the calendar), so
// captures and tests reproduce them. They change nothing in the game and are
// never saved.
//
// They roam: at share 0 about two sauropods and two or three raptors are on
// the map at any moment, each walking for a few minutes before it leaves and
// another comes, and a pterodactyl crosses every minute or so. Each walker
// has a slot that is always on its next walk; the number of slots on falls
// in a straight line with the gate share (rounded down) to none at
// dinoGateEnd, and the pterodactyl's chance falls the same way.

// The ambient kinds, after the visitor's.
const (
	SightSauropod    SightingKind = iota + SightWalker + 1 // a long-necked giant, slow
	SightRaptor                                            // a small fast walker
	SightPterodactyl                                       // a flyby across the upper rows
	SightAsteroid                                          // the fireball
)

const (
	// dinoGateEnd is the gate share from which no dinosaur shows.
	dinoGateEnd = 0.95

	// asteroidWindow is how far either side of its share the asteroid shows,
	// asteroidSlot the frames between streaks (about 10 seconds) and
	// asteroidLife how long one lasts.
	asteroidWindow = 0.01
	asteroidSlot   = 10 * 8
	asteroidLife   = 24

	// asteroidYear is the year the asteroid strikes: the calendar's end of
	// the Late Cretaceous.
	asteroidYear = -66000000
)

// dinoKinds lists the dinosaurs: how many slots each has at share 0, how long
// a visit lasts (min to max, in frames) and how long a slot's cycle is. A
// slot's walk starts somewhere in its cycle, so a walk is a little under
// the cycle and the next begins soon after. The pterodactyl has one slot
// that is on with a chance instead of in a count.
var dinoKinds = []struct {
	Kind          SightingKind
	Slots         int
	Min, Max, Per int
}{
	{SightSauropod, 2, 180 * 8, 210 * 8, 215 * 8},
	{SightRaptor, 3, 90 * 8, 135 * 8, 140 * 8},
	{SightPterodactyl, 1, 12 * 8, 15 * 8, 75 * 8},
}

// DinoFactor is how much of the dinosaurs is left at a gate share: 1 at 0,
// falling in a straight line to 0 at dinoGateEnd.
func DinoFactor(share float64) float64 {
	if share != share || share >= dinoGateEnd {
		return 0
	}
	return 1 - max(share, 0)/dinoGateEnd
}

// DinoSlots is how many slots of kind k (an index of dinoKinds) are on at a
// gate share: the count at share 0 scaled by DinoFactor, rounded down.
func DinoSlots(k int, share float64) int {
	return int(float64(dinoKinds[k].Slots)*DinoFactor(share) + 1e-9)
}

// DinosAt lists the dinosaurs on the map at animation frame anim.
func DinosAt(seed int64, share float64, anim int) []Sighting {
	if anim < 0 {
		return nil
	}
	f := DinoFactor(share)
	const salt = 0xd1a0
	var out []Sighting
	for i, k := range dinoKinds {
		ki := int64(i)
		for slot := 0; slot < k.Slots; slot++ {
			si := int64(slot)
			if k.Kind == SightPterodactyl {
				// one slot, on in a period with a chance of f
				if slot > 0 {
					break
				}
			} else if slot >= DinoSlots(i, share) {
				continue
			}
			off := int(Hash(seed, salt, ki, si, 0) % uint64(k.Per))
			n := int64((anim + off) / k.Per)
			if k.Kind == SightPterodactyl && float64(Hash(seed, salt, ki, si, n, 1)%1000) >= f*1000 {
				continue
			}
			s := Sighting{Kind: k.Kind, Frames: k.Min + int(Hash(seed, salt, ki, si, n, 2)%uint64(k.Max-k.Min+1)),
				Roll: Hash(seed, salt, ki, si, n, 3)}
			s.Start = int(n)*k.Per - off + int(Hash(seed, salt, ki, si, n, 4)%uint64(k.Per-s.Frames+1))
			if anim >= s.Start && anim < s.Start+s.Frames {
				out = append(out, s)
			}
		}
	}
	return out
}

// primitive says whether the model is in the Primitive Age.
func (m *Model) primitive() bool {
	return m.Catalog != nil && m.AgeIdx >= 0 && m.AgeIdx < len(m.Catalog.Ages) && m.Catalog.Ages[m.AgeIdx] == "primitive_age"
}

// Dinos lists the dinosaurs showing at animation frame anim: none outside the
// Primitive Age.
func (m *Model) Dinos(anim int) []Sighting {
	if !m.primitive() {
		return nil
	}
	return DinosAt(m.Seed, m.GateShare, anim)
}

var (
	asteroidOnce  sync.Once
	asteroidAtVal float64
)

// AsteroidShare is the gate share at which the Primitive Age's calendar
// first reaches the year of the asteroid.
func AsteroidShare() float64 {
	asteroidOnce.Do(func() {
		years := config.AgeStartYears([]string{"primitive_age", "stone_age"})
		for th := 0; th <= 1000; th++ {
			if config.YearBetween(years[0], years[1], float64(th)/1000) >= asteroidYear {
				asteroidAtVal = float64(th) / 1000
				return
			}
		}
		asteroidAtVal = 1
	})
	return asteroidAtVal
}

// AsteroidAt is the asteroid's streak showing at animation frame anim, if
// any: only in the Primitive Age and only while the gate share is within
// asteroidWindow of AsteroidShare; one streak a slot.
func (m *Model) AsteroidAt(anim int) (Sighting, bool) {
	if !m.primitive() || anim < 0 {
		return Sighting{}, false
	}
	if d := m.GateShare - AsteroidShare(); d > asteroidWindow || d < -asteroidWindow {
		return Sighting{}, false
	}
	slot := int64(anim / asteroidSlot)
	s := Sighting{Kind: SightAsteroid, Frames: asteroidLife, Roll: Hash(m.Seed, 0xa57e, slot, 1)}
	s.Start = int(slot)*asteroidSlot + int(Hash(m.Seed, 0xa57e, slot, 2)%uint64(asteroidSlot-asteroidLife))
	if anim < s.Start || anim >= s.Start+s.Frames {
		return Sighting{}, false
	}
	return s, true
}
