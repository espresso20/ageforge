package mapmodel

// glyphs_space.go is the sky arc's part of the symbol table: what the maps
// draw from the Space Age on (sky.go), when both styles leave the ground.
// The symbols number from skySymBase, well past the ground table's numSyms,
// so the two tables grow independently; G, AllGlyphs and Fold cover both.
// Every symbol has an ASCII, a Unicode and (where Font Awesome has one) a
// Nerd Font form, and no Unicode form is an emoji-presentation code point.

const skySymBase Sym = 200

const (
	// structures and places
	SymSkyHub      Sym = skySymBase + iota // the heart of a sky scene: the station hub, the starbase core
	SymSkySolar                            // a solar panel or a stellar collector
	SymSkyYard                             // a hull on the slip
	SymSkyCity                             // city lights on the night side
	SymSkyAsteroid                         // a rock in the belt
	SymSkyColony                           // a colony world
	SymSkyGate                             // the warp gate
	// movers (movers.go): the sky arc's traffic (the tether's climber is the
	// Earth arc's SymClimber)
	SymMiningDrone // a mining drone
	SymGenShip     // a generation ship
	SymStarship    // a starship
	SymPhaseShip   // a ship that tunnels
	SymMote        // a mote of light
	skySymEnd
)

// skyGlyphs is the sky arc's table, indexed from skySymBase.
var skyGlyphs = [skySymEnd - skySymBase]Glyph{
	SymSkyHub - skySymBase:      {'@', '⊕', 0xf140}, // bullseye
	SymSkySolar - skySymBase:    {'#', '▤', 0xf185}, // sun
	SymSkyYard - skySymBase:     {'#', '▓', 0xf0ad}, // wrench
	SymSkyCity - skySymBase:     {'*', '•', 0xf0eb}, // lightbulb
	SymSkyAsteroid - skySymBase: {'o', '◦', 0},
	SymSkyColony - skySymBase:   {'O', '●', 0xf111}, // circle
	SymSkyGate - skySymBase:     {'O', '○', 0xf10c}, // circle outline

	// Movers that share an age differ in every tier (TestMoverGlyphs).
	SymMiningDrone - skySymBase: {'m', '▫', 0xf1b2}, // cube
	SymGenShip - skySymBase:     {'G', '►', 0xf135}, // rocket
	SymStarship - skySymBase:    {'A', '●', 0xf0fb}, // fighter jet
	SymPhaseShip - skySymBase:   {'%', '◈', 0xf219}, // diamond
	SymMote - skySymBase:        {'.', '∘', 0xf005}, // star
}

// skyGlyph is the sky table's record for s, if s is one of its symbols.
func skyGlyph(s Sym) (Glyph, bool) {
	if s >= skySymBase && s < skySymEnd {
		return skyGlyphs[s-skySymBase], true
	}
	return Glyph{}, false
}

// skySyms lists the sky arc's symbols in order.
func skySyms() []Sym {
	out := make([]Sym, 0, skySymEnd-skySymBase)
	for s := skySymBase; s < skySymEnd; s++ {
		out = append(out, s)
	}
	return out
}
