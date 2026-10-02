package mapmodel

// glyphs.go is the map's symbol table, in three tiers. A terminal cannot
// bundle a font (the player's font draws every cell), so the maps pick
// symbols by what the player's font can show:
//
//   - TierASCII: plain printable ASCII, for poor fonts and dumb terminals.
//   - TierUnicode (default): box drawing, blocks, shades, geometric shapes,
//     arrows and the widely supported dingbats and symbols, every one single
//     width. Emoji-presentation code points (which draw double width) are
//     never used.
//   - TierNerd (opt-in): Nerd Font icons from the Private Use Area (the Font
//     Awesome set), for players with a Nerd Font installed. Every icon has a
//     Unicode fallback.
//
// Styles ask for a symbol by ID (Sym) and a tier; Fold maps any art rune a
// renderer draws itself (blocks, shades, box drawing) down to ASCII.

// GlyphTier is the symbol set a map draws with.
type GlyphTier uint8

const (
	TierUnicode GlyphTier = iota // the default
	TierASCII
	TierNerd
)

// TierNames lists the tiers in the order a setting offers them.
var TierNames = []string{"ascii", "unicode", "nerd"}

func (t GlyphTier) String() string {
	switch t {
	case TierASCII:
		return "ascii"
	case TierNerd:
		return "nerd"
	}
	return "unicode"
}

// ParseTier reads a tier name.
func ParseTier(s string) (GlyphTier, bool) {
	switch s {
	case "ascii":
		return TierASCII, true
	case "unicode", "":
		return TierUnicode, true
	case "nerd":
		return TierNerd, true
	}
	return TierUnicode, false
}

// Glyph is one symbol in every tier. Nerd is 0 when the symbol has no icon.
type Glyph struct{ ASCII, Unicode, Nerd rune }

// In returns the symbol for tier t, falling back nerd → unicode → ascii.
func (g Glyph) In(t GlyphTier) rune {
	switch t {
	case TierASCII:
		return g.ASCII
	case TierNerd:
		if g.Nerd != 0 {
			return g.Nerd
		}
	}
	return g.Unicode
}

// Sym names a symbol in the table.
type Sym uint16

const (
	SymNone Sym = iota
	// terrain
	SymDeep
	SymShallow
	SymRiver
	SymBeach
	SymGrass1
	SymGrass2
	SymGrass3
	SymForest
	SymForest2
	SymPine
	SymHills
	SymMountain
	SymStar
	SymStarBig
	// lineages (housing, fields and stores re-skin by epoch: see LineageSym)
	SymHut
	SymHouse
	SymHouseBlock
	SymTower
	SymHabitat
	SymField
	SymFieldPlough
	SymFieldGrid
	SymFieldHydro
	SymStore
	SymDataStore
	SymWood
	SymMine
	SymMetal
	SymEngineer
	SymEnergy
	SymHarbor
	SymHacker
	SymLaunch
	SymKnowledge
	SymFaith
	SymCulture
	SymMonument
	SymTrade
	SymMilitary
	SymDiplomacy
	SymWonder
	// the square, per epoch
	SymHearth
	SymHall
	SymTownHall
	SymPlazaDigital
	SymPlazaNeon
	SymPlazaCosmic
	// life and state
	SymWorker
	SymIdle
	SymCaravan
	SymScout
	SymRaider
	SymSmoke1
	SymSmoke2
	SymSmoke3
	SymRuin
	SymWar
	SymHarbinger
	SymHazard
	SymFire
	SymCiv
	SymFresh
	SymWarning
	SymCursor
	// weather and sky
	SymSun
	SymMoon
	SymCloud
	SymRain
	SymSnow
	SymBolt
	// vehicles (the skyline's traffic uses these as markers in the nerd tier)
	SymCart
	SymShip
	SymTrain
	SymCar
	SymPlane
	SymHeli
	SymRocket
	SymDrone
	SymSatellite
	// resources and gauges
	SymGold
	SymFood
	SymPop
	SymBuild
	// movers: the traffic each age adds (movers.go). Walkers, steam and
	// freight trains, steamships, planes, drones and satellites reuse the
	// symbols above.
	SymHunter
	SymOxCart
	SymRider
	SymRowboat
	SymWagon
	SymSailShip
	SymTram
	SymAuto
	SymTruck
	SymBoxShip
	SymMaglev
	SymHovercar
	SymShuttle
	SymOrbital
	// the rare visitor (never in a legend)
	SymUFO
	SymAlien
	numSyms
)

// glyphTable is the symbol table. Nerd code points are Font Awesome icons
// as Nerd Fonts ships them (nf-fa-*), all single width in a Nerd Font.
var glyphTable = [numSyms]Glyph{
	SymNone:     {' ', ' ', 0},
	SymDeep:     {'~', '≈', 0},
	SymShallow:  {'~', '~', 0},
	SymRiver:    {'~', '≈', 0},
	SymBeach:    {'.', '·', 0},
	SymGrass1:   {'.', '.', 0},
	SymGrass2:   {',', ',', 0},
	SymGrass3:   {'\'', '\'', 0},
	SymForest:   {'T', '♣', 0xf1bb},
	SymForest2:  {'t', 'τ', 0},
	SymPine:     {'A', '♠', 0},
	SymHills:    {'^', '^', 0},
	SymMountain: {'A', '▲', 0},
	SymStar:     {'.', '·', 0},
	SymStarBig:  {'*', '✦', 0xf005},

	SymHut:         {'n', '∩', 0xf015},
	SymHouse:       {'h', '⌂', 0xf015},
	SymHouseBlock:  {'H', '▪', 0xf0f7},
	SymTower:       {'#', '▓', 0xf1ad},
	SymHabitat:     {'O', '◘', 0xf1ad},
	SymField:       {'"', '"', 0xf06c},
	SymFieldPlough: {'=', '≡', 0xf06c},
	SymFieldGrid:   {'#', '▤', 0xf06c},
	SymFieldHydro:  {'#', '▦', 0xf06c},
	SymStore:       {'o', '□', 0xf187},
	SymDataStore:   {'o', '▣', 0xf1c0},
	SymWood:        {'Y', '♠', 0xf1bb},
	SymMine:        {'v', '▼', 0xf1b2},
	SymMetal:       {'*', '♦', 0xf0e3},
	SymEngineer:    {'%', '¤', 0xf085},
	SymEnergy:      {'&', '☼', 0xf0e7},
	SymHarbor:      {'J', '╤', 0xf13d},
	SymHacker:      {'L', 'λ', 0xf121},
	SymLaunch:      {'!', '↑', 0xf135},
	SymKnowledge:   {'S', '§', 0xf02d},
	SymFaith:       {'+', 'Ω', 0xf0f3},
	SymCulture:     {'m', '♫', 0xf001},
	SymMonument:    {'I', '◊', 0xf091},
	SymTrade:       {'$', '$', 0xf155},
	SymMilitary:    {'x', '†', 0xf132},
	SymDiplomacy:   {'=', '≡', 0xf2b5},
	SymWonder:      {'W', '★', 0xf219},

	SymHearth:       {'*', '*', 0xf06d},
	SymHall:         {'P', 'Π', 0xf19c},
	SymTownHall:     {'@', 'Φ', 0xf19c},
	SymPlazaDigital: {'O', '◘', 0xf0ac},
	SymPlazaNeon:    {'O', '◙', 0xf0ac},
	SymPlazaCosmic:  {'*', '☼', 0xf185},

	SymWorker:    {'o', '☺', 0xf183},
	SymIdle:      {'O', '☻', 0xf007},
	SymCaravan:   {'&', '&', 0xf0d1},
	SymScout:     {'@', '@', 0xf1e5},
	SymRaider:    {'>', '»', 0xf05b},
	SymSmoke1:    {'.', '∙', 0},
	SymSmoke2:    {'o', '°', 0},
	SymSmoke3:    {'\'', '˚', 0},
	SymRuin:      {'%', '%', 0xf127},
	SymWar:       {'X', '×', 0xf00d},
	SymHarbinger: {'Y', 'Ψ', 0xf06e},
	SymHazard:    {'!', '‼', 0xf071},
	SymFire:      {'^', '^', 0xf06d},
	SymCiv:       {'C', '▪', 0xf286},
	SymFresh:     {'+', '+', 0xf055},
	SymWarning:   {'!', '⚠', 0xf071},
	SymCursor:    {'X', '▸', 0xf0da},

	SymSun:   {'O', '☼', 0xf185},
	SymMoon:  {'C', '☾', 0xf186},
	SymCloud: {'~', '☁', 0xf0c2},
	SymRain:  {'/', '╱', 0xf043},
	SymSnow:  {'*', '❄', 0xf2dc},
	SymBolt:  {'Z', 'ϟ', 0xf0e7},

	SymCart:      {'o', '▭', 0xf0d1},
	SymShip:      {'v', '◒', 0xf21a},
	SymTrain:     {'=', '▬', 0xf238},
	SymCar:       {'o', '▭', 0xf1b9},
	SymPlane:     {'>', '✈', 0xf072},
	SymHeli:      {'+', '✢', 0xf072},
	SymRocket:    {'^', '▲', 0xf135},
	SymDrone:     {'x', '✕', 0xf1d8},
	SymSatellite: {'+', '✧', 0xf1eb},

	SymGold:  {'$', '¤', 0xf155},
	SymFood:  {'f', '♣', 0xf06c},
	SymPop:   {'p', '☺', 0xf0c0},
	SymBuild: {'#', '▲', 0xf275},

	// Movers that never share an age may share a shape (an ox cart and a
	// car are both a box on the street); movers that do share an age differ.
	SymHunter:   {'i', '♂', 0xf1b0}, // paw
	SymOxCart:   {'c', '▭', 0xf0d1}, // truck
	SymRider:    {'R', '♞', 0xf21c}, // motorcycle: Font Awesome 4 has no horse
	SymRowboat:  {'u', '◡', 0xf21a}, // ship
	SymWagon:    {'w', '▮', 0xf0d1},
	SymSailShip: {'A', '△', 0xf21a},
	SymTram:     {'B', '◫', 0xf207}, // bus
	SymAuto:     {'a', '▭', 0xf1b9}, // car
	SymTruck:    {'r', '▮', 0xf0d1},
	SymBoxShip:  {'D', '▥', 0xf21a},
	SymMaglev:   {'=', '▰', 0xf239}, // subway
	SymHovercar: {'z', '◇', 0xf1ba}, // taxi
	SymShuttle:  {'^', '▴', 0xf197}, // space shuttle
	SymOrbital:  {'0', '◎', 0xf1cd}, // life ring
	SymUFO:      {'O', '◉', 0xf192}, // dot in a circle
	SymAlien:    {'Q', '☿', 0xf21b}, // a figure in disguise
}

// G returns a symbol's glyph record.
func G(s Sym) Glyph {
	if s >= numSyms {
		return glyphTable[SymNone]
	}
	return glyphTable[s]
}

// R returns a symbol's rune in tier t.
func R(s Sym, t GlyphTier) rune { return G(s).In(t) }

// AllGlyphs returns every symbol in the table (for tests and legends).
func AllGlyphs() []Glyph {
	out := make([]Glyph, 0, numSyms)
	for s := Sym(1); s < numSyms; s++ {
		out = append(out, glyphTable[s])
	}
	return out
}

// LineageSym is the one symbol a lineage keeps across every age, so a player
// learns the map once: § is always knowledge, $ is always trade. Housing,
// fields and stores re-skin by epoch.
func LineageSym(lin string, epoch int) Sym {
	switch lin {
	case LinHousing:
		return pickEpoch(epoch, SymHut, SymHouse, SymHouse, SymHouseBlock, SymTower, SymTower, SymHabitat)
	case LinFood:
		return pickEpoch(epoch, SymField, SymField, SymFieldPlough, SymFieldPlough, SymFieldPlough, SymFieldGrid, SymFieldHydro)
	case LinStorage:
		return pickEpoch(epoch, SymStore, SymStore, SymStore, SymStore, SymDataStore, SymDataStore, SymDataStore)
	case LinWood:
		return SymWood
	case LinMines:
		return SymMine
	case LinMetal:
		return SymMetal
	case LinEngineer:
		return SymEngineer
	case LinEnergy:
		return SymEnergy
	case LinHarbor:
		return SymHarbor
	case LinHacker:
		return SymHacker
	case LinKnowledge:
		return SymKnowledge
	case LinFaith:
		return SymFaith
	case LinCulture:
		return SymCulture
	case LinMonument:
		return SymMonument
	case LinTrade:
		return SymTrade
	case LinMilitary:
		return SymMilitary
	case LinDiplomacy:
		return SymDiplomacy
	case LinWonder:
		return SymWonder
	}
	return SymEngineer
}

// CentreSym is the square's landmark in an epoch.
func CentreSym(epoch int) Sym {
	return pickEpoch(epoch, SymHearth, SymHall, SymTownHall, SymTownHall, SymPlazaDigital, SymPlazaNeon, SymPlazaCosmic)
}

func pickEpoch(epoch int, s ...Sym) Sym {
	if epoch < 0 {
		epoch = 0
	}
	if epoch >= len(s) {
		epoch = len(s) - 1
	}
	return s[epoch]
}

// Fold maps an art rune to what tier t can show. Unicode and nerd tiers
// keep it; the ASCII tier folds blocks, shades, box drawing and the symbols
// the maps draw to their nearest ASCII shape, and anything else non-ASCII
// to '?'.
func Fold(r rune, t GlyphTier) rune {
	if t != TierASCII || r < 0x80 {
		return r
	}
	if a, ok := asciiFold[r]; ok {
		return a
	}
	for _, g := range glyphTable {
		if g.Unicode == r || g.Nerd == r {
			return g.ASCII
		}
	}
	switch {
	case r >= 0x2500 && r <= 0x257F: // box drawing
		return boxFold(r)
	case r >= 0x2580 && r <= 0x259F: // blocks and shades
		return '#'
	case r >= 0x2800 && r <= 0x28FF: // braille
		return ':'
	}
	return '?'
}

var asciiFold = map[rune]rune{
	'█': '#', '▓': '#', '▒': ':', '░': '.', '▀': '"', '▄': '_', '▌': '[', '▐': ']',
	'▁': '_', '▂': '_', '▃': '_', '▅': '=', '▆': '=', '▇': '#', '▔': '-', '▕': '|', '▏': '|',
	'◢': '/', '◣': '\\', '◤': '/', '◥': '\\', '▲': 'A', '▼': 'v', '◆': '*', '◇': 'o',
	'·': '.', '∙': '.', '•': 'o', '°': 'o', '˚': '\'', '…': '.', '─': '-', '│': '|',
	'═': '=', '║': '|', '╱': '/', '╲': '\\', '╳': 'X', '≈': '~', '≡': '=', '∩': 'n',
	'⌂': 'h', '♣': 'T', '♠': 'Y', '♦': '*', '♥': 'v', '☺': 'o', '☻': 'O', '☼': '*',
	'☾': 'C', '★': '*', '☆': '*', '✦': '*', '✧': '+', '×': 'x', '÷': '/', '±': '+',
	'←': '<', '→': '>', '↑': '^', '↓': 'v', '◄': '<', '►': '>', '◀': '<', '▶': '>',
	'▸': '>', '◂': '<', '▴': '^', '▾': 'v', '○': 'o', '●': '@', '◘': 'O', '◙': 'O',
	'□': 'o', '■': '#', '▪': '#', '▫': '.', '▭': '=', '▬': '=', '▣': '#', '▤': '#',
	'▥': '#', '▦': '#', '▧': '#', '▨': '#', '▩': '#', '◒': 'v', '◓': 'A', '◊': 'I',
	'†': '+', '‡': '+', '§': 'S', '¤': '%', '¥': 'Y', '£': 'L', '¢': 'c', 'Ω': 'O',
	'Π': 'P', 'Φ': '@', 'Ψ': 'Y', 'λ': 'L', 'τ': 't', 'ϟ': 'Z', '‼': '!', '⚠': '!',
	'✈': '>', '✢': '+', '✕': 'x', '☁': '~', '❄': '*', '♫': 'm', '♪': 'm', '»': '>',
	'«': '<', '¦': '|', '˙': '.', '¨': '"', '⁺': '+', '¹': '1', '²': '2', '³': '3',
	'⁴': '4', '⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9', '⁰': '0',
}

func boxFold(r rune) rune {
	switch r {
	case '─', '━', '═', '┄', '┅', '┈', '┉', '╌', '╍':
		return '-'
	case '│', '┃', '║', '┆', '┇', '┊', '┋', '╎', '╏':
		return '|'
	case '╱':
		return '/'
	case '╲':
		return '\\'
	case '╳':
		return 'X'
	}
	return '+'
}
