package mapmodel

// glyphs_city.go is the Earth arc's symbol block, the Modern Age to the
// Fusion Age: the movers those ages add and the city they build (suburbs,
// parks, glass towers, server farms, dishes, gardens, landfill, arcologies,
// neon, vents, scrap, reactors, launch towers). It sits past the core table
// in glyphs.go, from symCityBase, so the two grow apart: G reads a symbol
// past the core table through extGlyph. Every symbol has the three tiers,
// and cityFold folds every art rune the Earth arc draws to ASCII.

// symCityBase is the first symbol of the city block.
const symCityBase Sym = 0x400

const (
	// movers (movers.go)
	SymNewsHeli Sym = symCityBase + iota
	SymSkyTrain
	SymCrowd
	SymClimber
	// the city
	SymSuburb
	SymPark
	SymGlassTower
	SymServerFarm
	SymDish
	SymGarden
	SymLandfill
	SymArcology
	SymNeonSign
	SymVent
	SymScrap
	SymReactor
	SymLaunchTower
	symCityEnd
)

// cityGlyphs is the city block. Nerd code points are Font Awesome icons as
// Nerd Fonts ships them (nf-fa-*), single width in a Nerd Font.
var cityGlyphs = [symCityEnd - symCityBase]Glyph{
	SymNewsHeli - symCityBase:    {'H', '✢', 0xf03d}, // video camera: a news crew aloft
	SymSkyTrain - symCityBase:    {'T', '▬', 0xf238}, // train
	SymCrowd - symCityBase:       {':', '∷', 0xf0c0}, // users
	SymClimber - symCityBase:     {'E', '◘', 0xf077}, // chevron up
	SymSuburb - symCityBase:      {'h', '⌂', 0xf015}, // home
	SymPark - symCityBase:        {'T', '♣', 0xf1bb}, // tree
	SymGlassTower - symCityBase:  {'#', '▦', 0xf1ad}, // building
	SymServerFarm - symCityBase:  {':', '▤', 0xf233}, // server
	SymDish - symCityBase:        {'d', '◎', 0xf09e}, // feed
	SymGarden - symCityBase:      {'"', '♧', 0xf06c}, // leaf
	SymLandfill - symCityBase:    {'%', '▒', 0xf1f8}, // trash
	SymArcology - symCityBase:    {'A', '◈', 0xf1ad}, // building
	SymNeonSign - symCityBase:    {'Y', '¥', 0xf0eb}, // light bulb
	SymVent - symCityBase:        {'o', '○', 0xf10c}, // circle
	SymScrap - symCityBase:       {'%', '‰', 0xf1f8}, // trash
	SymReactor - symCityBase:     {'@', '◉', 0xf0e7}, // bolt
	SymLaunchTower - symCityBase: {'#', '╫', 0xf135}, // rocket
}

// extGlyph is a symbol past the core table: the city block's, or the empty
// symbol.
func extGlyph(s Sym) Glyph {
	if s >= symCityBase && s < symCityEnd {
		return cityGlyphs[s-symCityBase]
	}
	return glyphTable[SymNone]
}

// extGlyphs lists the symbols past the core table, for AllGlyphs.
func extGlyphs() []Glyph { return cityGlyphs[:] }

// cityFold folds the Earth arc's art runes to ASCII: the neon signs'
// letters, crowds, scrap, gardens and the arcology's crown. Fold reads it
// through asciiFold, which init extends; a rune the core table already
// folds keeps the core's fold.
var cityFold = map[rune]rune{
	'∷': ':', '‰': '%', '♧': '"', '◈': 'A', '◎': 'o', 'ƒ': 'f', 'Σ': 'E', '∞': '8',
	'⌂': 'h', '▦': '#', '▤': '#', '✢': '+', '◘': 'O', '◉': '@', '╫': '#', '╪': '#',
	'┄': '-', '┆': '|', '┈': '-', '┊': '|', '╍': '-', '╏': '|', '▔': '-', '¦': '|',
	'◦': 'o', '⁘': ':', '∴': ':', '⋅': '.', '∘': 'o', '€': 'E', '▴': '^', '◊': 'I',
}

func init() {
	add := func(r, a rune) {
		if _, ok := asciiFold[r]; ok || r < 0x80 || coreFolds(r) {
			return
		}
		asciiFold[r] = a
	}
	for r, a := range cityFold {
		add(r, a)
	}
	for _, g := range cityGlyphs {
		add(g.Unicode, g.ASCII)
		add(g.Nerd, g.ASCII)
	}
}

// coreFolds reports whether the core table already folds r (Fold scans it
// after asciiFold), so the city block never changes a core symbol's fold.
func coreFolds(r rune) bool {
	for _, g := range glyphTable {
		if g.Unicode == r || g.Nerd == r {
			return true
		}
	}
	return false
}
