package mapgen

// Colour is described, not chosen. The engine says what a thing is made of
// (a Hue) and how strongly it shows (a Weight); the screen layer turns that
// into a colour of the player's theme, so a map reads in every theme.

// Hue names a colour of the map by what it is.
type Hue uint8

const (
	HNone Hue = iota // nothing: the theme's own ground
	// the land
	HDeep
	HWater
	HSand
	HScrub
	HGrass
	HForest
	HHill
	HPeak
	// made ground
	HEarth // packed earth, yards, dirt tracks
	HPath  // pale gravel and flagstones
	HAsphalt
	HConcrete
	HLawn // parks and gardens
	HRubble
	HCrop
	// materials
	HThatch
	HTimber
	HStone
	HBrick
	HTile
	HGlass
	HSteel
	HGold
	HNeon
	HCyan
	HPlasma
	HFire
	HSmoke
	// light
	HLight // a warm light: windows, lamps, hearths
	HCool  // a cold screen glow
	// life and signals
	HLife
	HRed
	HWhite
	HAmber
	// dark falling
	HNight
	// space
	HVoid
	HStar
	HNebula
	HSun
	HSea
	HLand
	HCloud
	HAir
	HOrbit
	HEngine
	HPanel
	HHydro
	HRock
	// the building lines come last, one hue each in the order of
	// mapmodel.LineageOrder
	HLine
	NumHues = HLine + numLines
)

// numLines is how many building lines there are, which the line hues follow
// (a test holds it to len(mapmodel.LineageOrder)).
const numLines = 17

// lineHue is the hue that tints the buildings of a line.
func lineHue(line string) Hue { return HLine + Hue(lineOf[line]) }

// Paint is a hue at a strength: Weight of 255 is the hue itself, less lets
// the theme's ground show through, none is the ground.
type Paint struct {
	Hue    Hue
	Weight uint8
}

// P is a paint of hue h at weight w of 255.
func P(h Hue, w int) Paint { return Paint{h, uint8(clamp(w, 0, 255))} }

// fade lets k/256 of a paint show: the rest is the ground.
func (p Paint) fade(k int) Paint { return P(p.Hue, int(p.Weight)*k/256) }

// vary moves a paint's weight up or down a little by h.
func (p Paint) vary(h uint64, amount int) Paint {
	return P(p.Hue, int(p.Weight)+int(h>>8)%(2*amount+1)-amount)
}
