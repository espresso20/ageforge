package mapgen

import (
	"maps"
	"strings"
)

// An age style is one row of data: what the ground, the roads and the
// buildings of an age look like. Adding an age is adding a row; no code
// changes. A row serves every age from its From up to the next row's.

// Scale is how far the picture is pulled back: the ground, the city, orbit.
type Scale uint8

const (
	Ground Scale = iota // open country and villages
	City                // streets and blocks
	Orbit               // a starfield, a planet and ring roads
)

// Look is how one sort of cell is drawn: a glyph picked by hash from Glyphs
// in Dense percent of cells (all when 0), over a background.
type Look struct {
	Glyphs string
	Fg, Bg Paint
	Dense  int
}

// RoadLook draws a road. Joins is sixteen glyphs indexed by which
// neighbours are road (N=1, E=2, S=4, W=8); a shorter string is a set to
// pick from by hash, for roads that do not join up (a dirt track).
type RoadLook struct {
	Joins  string
	Fg, Bg Paint
}

// Art is a picture in characters; a space shows what is below.
type Art []string

func art(rows string) Art { return strings.Split(rows, "\n") }

// Family groups the building lines that share a silhouette.
type Family uint8

const (
	Home  Family = iota // housing, food
	Work                // production, power, storage
	Civic               // knowledge, faith, trade, defence
	numFamilies
)

const numTiers = 4

// Tiers holds the art of one family by size tier (hut, house, block, tower),
// a few variants of each.
type Tiers [numTiers][]Art

// Scatter strews detail over a biome: Chance percent of its spots get one of
// Glyphs.
type Scatter struct {
	On     Biome
	Glyphs string
	Fg     Paint
	Chance int
}

// AgentKind is traffic: things that move along the roads, one per Per
// people up to Max, shown as Glyphs (east, west, south, north in turn; one
// glyph serves all four), moving a cell every Slow frames.
type AgentKind struct {
	Name   string
	Glyphs string
	Fg     Paint
	Per    int
	Max    int
	Slow   int
}

// Night is how dark falls: colours slide toward the Tint hue by Dim (of 256)
// at full night, lit windows and lamps glow in the Light hue.
type Night struct {
	Tint  Hue
	Dim   int
	Light Hue
	Lamps bool
}

// AgeStyle is one age's look.
type AgeStyle struct {
	From  string // the first age the row serves
	Scale Scale
	// Land is the ground by biome. In orbit, Deep is the planet's sea, Grass
	// its land, Hills its cloud and Sand its limb of air.
	Land   [numBiomes]Look
	Roads  [3]RoadLook // by road tier; in orbit, the rings (0), spokes (1)
	Deck   Paint       // a bridge's boards
	Yard   Look        // the lot a building stands on
	Crops  Look        // a farm's lot
	Plaza  Look        // an unbuilt wonder plot
	Rubble Look        // a lost building's lot
	Shapes [numFamilies]Tiers
	// Filler is the art of ordinary houses: one lot in the town's middle is
	// given to one for every FillPer people, where no building stands.
	Filler  []Art
	FillPer int
	// Wall colours every building rune not in Ink; Windows lists the runes
	// that light up at night.
	Wall    Hue
	Ink     map[rune]Paint
	Windows string
	// Chimney is the art rune that smokes, and Smoke how its smoke looks.
	Chimney rune
	Smoke   Look
	Scatter []Scatter
	Agents  []AgentKind
	Night   Night
	// Tones dresses wonder templates: the material letter of a template
	// cell, in this age's colours.
	Tones map[byte]Paint
	// Sky is the void and its stars (orbit only).
	Sky Look
	// Pad is the art of one launch pad, and Pads how many the spaceport has:
	// none, no spaceport.
	Pad  Art
	Pads int
}

// AgeRow is one age: the era whose look it shares, and what sets it apart. A
// field left empty keeps the era's.
type AgeRow struct {
	Age    string
	Era    *AgeStyle
	Roads  [3]RoadLook        // the age's road kinds
	Wall   Hue                // the age's material
	Top    [numFamilies][]Art // its tallest buildings, replacing the era's
	Ink    map[rune]Paint     // colours added to the era's
	Tones  map[byte]Paint     // wonder materials added to the era's
	Filler []Art
	Agents []AgentKind
	Night  Night
	Pads   int // launch pads of the spaceport, when there is one
}

// compose makes the style of an age row: the era's, with the row's changes.
func compose(row AgeRow) AgeStyle {
	st := *row.Era
	st.From = row.Age
	if row.Roads[0].Joins != "" {
		st.Roads = row.Roads
	}
	if row.Wall != HNone {
		st.Wall = row.Wall
	}
	for fam, art := range row.Top {
		if len(art) > 0 {
			st.Shapes[fam][numTiers-1] = art
		}
	}
	st.Ink = mergeInk(st.Ink, row.Ink)
	st.Tones = mergeInk(st.Tones, row.Tones)
	if len(row.Filler) > 0 {
		st.Filler = row.Filler
	}
	if len(row.Agents) > 0 {
		st.Agents = row.Agents
	}
	if row.Night.Dim > 0 {
		st.Night = row.Night
	}
	st.Pads = row.Pads
	return st
}

// mergeInk is the map a plus the entries of b, a map of its own.
func mergeInk[K comparable](a, b map[K]Paint) map[K]Paint {
	out := make(map[K]Paint, len(a)+len(b))
	maps.Copy(out, a)
	maps.Copy(out, b)
	return out
}

// composeAges makes the styles of the rows, in order.
func composeAges(rows []AgeRow) []AgeStyle {
	out := make([]AgeStyle, len(rows))
	for i, row := range rows {
		out[i] = compose(row)
	}
	return out
}

// styleFor is the row that serves an age index; the first row serves any age
// before it.
func styleFor(ageIdx func(string) int, age int) *AgeStyle {
	row := &styles[0]
	for i := range styles {
		if ageIdx(styles[i].From) <= age {
			row = &styles[i]
		}
	}
	return row
}

// pick chooses one rune of s by h.
func pick(s string, h uint64) rune {
	runes := []rune(s)
	return runes[h%uint64(len(runes))]
}

// at draws a cell of the look at (x, y); salt moves the pattern (the
// animation phase, or a layer's number).
func (l Look) at(x, y int, salt int64) Cell {
	h := hash(salt, int64(x), int64(y))
	c := Cell{R: ' ', Fg: l.Fg, Bg: l.Bg}
	if l.Glyphs != "" && (l.Dense == 0 || int(h>>40)%100 < l.Dense) {
		c.R = pick(l.Glyphs, h>>20)
		c.Fg = l.Fg.vary(h, 18)
	}
	return c
}

// joinRune is the glyph of a road or river cell whose neighbours N, E, S, W
// are marked in mask.
func joinRune(joins string, mask int, h uint64) rune {
	runes := []rune(joins)
	if len(runes) == 16 {
		return runes[mask]
	}
	return runes[h%uint64(len(runes))]
}

// Common join sets.
const (
	joinsLight   = "·╵╶└╷│┌├╴┘─┴┐┤┬┼"
	joinsHeavy   = "·╹╺┗╻┃┏┣╸┛━┻┓┫┳╋"
	joinsDouble  = "·║═╚║║╔╠═╝═╩╗╣╦╬"
	joinsRounded = "·╵╶╰╷│╭├╴╯─┴╮┤┬┼"
)
