package mapmodel

import (
	"sync"

	"github.com/espresso20/ageforge/config"
)

// dinos.go schedules the Primitive Age's ambient life: three dinosaurs that
// roam the town map and, near the end of the Late Cretaceous, the asteroid.
// Like the visitors they are pure functions of the save's seed, the age and
// the animation frame (plus the gate share, which stands for the calendar),
// so captures and tests reproduce them. They change nothing in the game and
// are never saved.
//
// The dinosaurs thin out as the gate fills: at share 0 each kind comes about
// as often as the busiest visitor does today (one visit a block), and the
// chance falls in a straight line to nothing at dinoGateEnd.

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
	// dinoOdds is the chance scale: a kind shows in a block when its roll
	// out of dinoOdds falls under the chance for the share.
	dinoOdds = 1000

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

// dinoKinds lists the dinosaurs with how long each is about, in frames.
var dinoKinds = []struct {
	Kind     SightingKind
	Min, Max int
}{
	{SightSauropod, 30 * 8, 45 * 8},
	{SightRaptor, 6 * 8, 9 * 8},
	{SightPterodactyl, 12 * 8, 15 * 8},
}

// DinoChance is the chance, in thousandths, that a dinosaur kind is about in
// a block at a gate share: all of them at 0, falling in a straight line to
// none at dinoGateEnd.
func DinoChance(share float64) int {
	if share != share || share >= dinoGateEnd {
		return 0
	}
	return int((1 - max(share, 0)/dinoGateEnd) * dinoOdds)
}

// DinosAt lists the dinosaurs showing at animation frame anim.
func DinosAt(seed int64, share float64, anim int) []Sighting {
	if anim < 0 {
		return nil
	}
	chance := DinoChance(share)
	if chance == 0 {
		return nil
	}
	const salt = 0xd1a0
	b := int64(anim / SightingBlock)
	var out []Sighting
	for i, k := range dinoKinds {
		ki := int64(i)
		if int(Hash(seed, salt, b, ki, 1)%dinoOdds) >= chance {
			continue
		}
		s := Sighting{Kind: k.Kind, Frames: k.Min + int(Hash(seed, salt, b, ki, 2)%uint64(k.Max-k.Min+1)),
			Roll: Hash(seed, salt, b, ki, 3)}
		lo, span := SightingBlock*10/100, SightingBlock*80/100-s.Frames
		s.Start = int(b)*SightingBlock + lo + int(Hash(seed, salt, b, ki, 4)%uint64(span))
		if anim >= s.Start && anim < s.Start+s.Frames {
			out = append(out, s)
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
