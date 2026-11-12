package world

// opens is the age that opens each frame of the ladder to the player, by
// FrameKind. This table is the whole rule: the map reads it and nothing else
// says what an age may see. A kind with no age is always open (the planet).
var opens = [...]string{
	FrameWeb:           AgeTranscendent,
	FrameLocalGroup:    AgeQuantum,
	FrameGalaxy:        AgeGalactic,
	FrameRegion:        AgeGalactic,
	FrameNeighbourhood: AgeInterstellar,
	FrameSystem:        AgeSpace,
	FramePlanet:        "",
}

// OpensAt is the age that opens frames of kind k ("" for always).
func OpensAt(k FrameKind) string { return opens[k] }

// Gate is what one age lets the map show: Open is the outermost frame the
// player can enter, and Locked says the frame just outside it is drawn greyed
// (its objects can be selected, not entered). Nothing outside the locked frame
// is drawn at all.
type Gate struct {
	Open   FrameKind
	Locked bool
}

// Top is the outermost frame drawn: the locked one, else the open one.
func (g Gate) Top() FrameKind {
	if g.Locked {
		return g.Open - 1
	}
	return g.Open
}

// GateOf is the gate of an age. The ground ages have no sky, so the planet is
// all they could show and nothing is locked.
func GateOf(age string) Gate {
	at := -1
	for i, a := range SpaceAges {
		if a == age {
			at = i
		}
	}
	g := Gate{Open: FramePlanet}
	if at < 0 {
		return g
	}
	for k := FramePlanet; k > FrameWeb; k-- {
		if !ageAtLeast(opens[k-1], at) {
			break
		}
		g.Open = k - 1
	}
	g.Locked = g.Open > FrameWeb
	return g
}

// ageAtLeast says whether the age at index at of SpaceAges has reached age a.
func ageAtLeast(a string, at int) bool {
	for i, s := range SpaceAges {
		if s == a {
			return i <= at
		}
	}
	return false
}
