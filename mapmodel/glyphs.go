package mapmodel

// glyphs.go is the map's symbol table, in two tiers:
//
//   - TierASCII: plain printable ASCII, for poor fonts and dumb terminals.
//   - TierUnicode (default): box drawing, blocks, shades, geometric shapes,
//     arrows and the widely supported dingbats and symbols, every one single
//     width. Emoji-presentation code points (which draw double width) are
//     never used.
//
// Styles ask for a symbol by ID (Sym) and a tier; Fold maps any art rune a
// renderer draws itself (blocks, shades, box drawing) down to ASCII.

// GlyphTier is the symbol set a map draws with.
type GlyphTier uint8

const (
	TierUnicode GlyphTier = iota // the default
	TierASCII
)

// TierNames lists the tiers in the order a setting offers them.
var TierNames = []string{"ascii", "unicode"}

func (t GlyphTier) String() string {
	switch t {
	case TierASCII:
		return "ascii"
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
	}
	return TierUnicode, false
}

// Glyph is one symbol in every tier.
type Glyph struct{ ASCII, Unicode rune }

// In returns the symbol for tier t.
func (g Glyph) In(t GlyphTier) rune {
	if t == TierASCII {
		return g.ASCII
	}
	return g.Unicode
}

// Sym names a symbol in the table.
type Sym uint16

const (
	SymNone Sym = iota
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
	// state
	SymRuin
	SymWar
	SymHarbinger
	SymHazard
	SymCiv
	// weather and sky
	SymSun
	SymMoon
	SymCloud
	SymRain
	SymSnow
	SymBolt
	// resources
	SymGold
	SymFood
	// the rare visitor (never in a legend)
	SymUFO
	numSyms
)

// glyphTable is the symbol table.
var glyphTable = [numSyms]Glyph{
	SymNone:    {' ', ' '},
	SymStarBig: {'*', '✦'},

	SymHut:         {'n', '∩'},
	SymHouse:       {'h', '⌂'},
	SymHouseBlock:  {'H', '▪'},
	SymTower:       {'#', '▓'},
	SymHabitat:     {'O', '◘'},
	SymField:       {'"', '"'},
	SymFieldPlough: {'=', '≡'},
	SymFieldGrid:   {'#', '▤'},
	SymFieldHydro:  {'#', '▦'},
	SymStore:       {'o', '□'},
	SymDataStore:   {'o', '▣'},
	SymWood:        {'Y', '♠'},
	SymMine:        {'v', '▼'},
	SymMetal:       {'*', '♦'},
	SymEngineer:    {'%', '¤'},
	SymEnergy:      {'&', '☼'},
	SymHarbor:      {'J', '╤'},
	SymHacker:      {'L', 'λ'},
	SymKnowledge:   {'S', '§'},
	SymFaith:       {'+', 'Ω'},
	SymCulture:     {'m', '♫'},
	SymMonument:    {'I', '◊'},
	SymTrade:       {'$', '$'},
	SymMilitary:    {'x', '†'},
	SymDiplomacy:   {'=', '≡'},
	SymWonder:      {'W', '★'},

	SymHearth:       {'*', '*'},
	SymHall:         {'P', 'Π'},
	SymTownHall:     {'@', 'Φ'},
	SymPlazaDigital: {'O', '◘'},
	SymPlazaNeon:    {'O', '◙'},
	SymPlazaCosmic:  {'*', '☼'},

	SymRuin:      {'%', '%'},
	SymWar:       {'X', '×'},
	SymHarbinger: {'Y', 'Ψ'},
	SymHazard:    {'!', '‼'},
	SymCiv:       {'C', '▪'},

	SymSun:   {'O', '☼'},
	SymMoon:  {'C', '☾'},
	SymCloud: {'~', '☁'},
	SymRain:  {'/', '╱'},
	SymSnow:  {'*', '❄'},
	SymBolt:  {'Z', 'ϟ'},

	SymGold: {'$', '¤'},
	SymFood: {'f', '♣'},

	SymUFO: {'O', '◉'}, // dot in a circle
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

// Fold maps an art rune to what tier t can show. The Unicode tier
// keeps it; the ASCII tier folds blocks, shades, box drawing and the symbols
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
		if g.Unicode == r {
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
	'↗': '/', '↖': '\\', '↘': '\\', '↙': '/',
	'▸': '>', '◂': '<', '▴': '^', '▾': 'v', '○': 'o', '●': '@', '◘': 'O', '◙': 'O',
	'◌': 'o', '□': 'o', '■': '#', '▪': '#', '▫': '.', '▭': '=', '▬': '=', '▣': '#', '▤': '#',
	'▥': '#', '▦': '#', '▧': '#', '▨': '#', '▩': '#', '◒': 'v', '◓': 'A', '◊': 'I',
	'†': '+', '‡': '+', '§': 'S', '¤': '%', '¥': 'Y', '£': 'L', '¢': 'c', 'Ω': 'O',
	'Π': 'P', 'Φ': '@', 'Ψ': 'Y', 'λ': 'L', 'τ': 't', 'ϟ': 'Z', '‼': '!', '⚠': '!',
	'✈': '>', '✢': '+', '✕': 'x', '☁': '~', '❄': '*', '♫': 'm', '♪': 'm', '»': '>',
	'«': '<', '¦': '|', '˙': '.', '¨': '"', '⁺': '+', '¹': '1', '²': '2', '³': '3',
	'⁴': '4', '⁵': '5', '⁶': '6', '⁷': '7', '⁸': '8', '⁹': '9', '⁰': '0',
	'‾': '-', '◠': 'n', 'ƒ': 'f', 'Σ': 'E', '‰': '%', '⁘': ':', '€': 'E', '∘': 'o', '∞': '8', '∴': ':',
	'∷': ':', '⊕': '@', '⋅': '.', '╪': '#', '╫': '#', '▮': 'w', '▰': '=', '△': 'A', '◈': 'A', '◎': '0',
	'◡': 'u', '◦': 'o', '◫': 'B', '☿': 'Q', '♂': 'i', '♞': 'R', '♧': '"',
	'◖': '(', '◗': ')',
	'⬢': '#', '⬡': 'o', '◉': 'O', 'ᗒ': '>', 'ᗑ': 'v', 'ᗕ': '<', 'ᗐ': '^',
	'⌾': 'o', '✺': '*', '⊛': '*',
	'✜': '+', '⊚': 'O', '⊝': 'O', '⫷': '<', '⫸': '>',
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
