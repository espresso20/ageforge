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
	// the Transcendent mandala (ui/mapstyle/mandala.go): each era's bead,
	// Stone to Cosmic (EraSym), the crown's petal and the core
	SymEraStone
	SymEraIron
	SymEraSteel
	SymEraElectric
	SymEraDigital
	SymEraNeon
	SymEraCosmic
	SymPetal
	SymCore
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

	// The mandala's beads: one per era, each the same mirrored left to
	// right and top to bottom, so the rings stay symmetric.
	SymEraStone - skySymBase:    {'o', '•', 0},
	SymEraIron - skySymBase:     {'x', '×', 0},
	SymEraSteel - skySymBase:    {'+', '¤', 0},
	SymEraElectric - skySymBase: {'=', '≡', 0},
	SymEraDigital - skySymBase:  {'#', '▣', 0},
	SymEraNeon - skySymBase:     {'@', '◉', 0},
	SymEraCosmic - skySymBase:   {'*', '✧', 0},
	SymPetal - skySymBase:       {'%', '◆', 0},
	SymCore - skySymBase:        {'O', '✦', 0},
}

// EraSym is the bead of era e (0 Stone … 6 Cosmic) on the mandala.
func EraSym(e int) Sym {
	e = max(0, min(e, int(SymEraCosmic-SymEraStone)))
	return SymEraStone + Sym(e)
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
