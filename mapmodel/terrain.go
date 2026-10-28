package mapmodel

import "github.com/espresso20/ageforge/vibe"

// Terrain is a tile's ground type, as the map names it.
type Terrain uint8

const (
	TDeep Terrain = iota
	TShallow
	TRiver
	TBeach
	TGrass
	TForest
	THills
	TMountain
)

// Pt is a tile coordinate.
type Pt struct{ X, Y int }

// The syllables a town's name is put together from.
var syllA = []vibe.Entry{
	vibe.E("mapmodel.world.syllA.01", "Ash"),
	vibe.E("mapmodel.world.syllA.02", "Bram"),
	vibe.E("mapmodel.world.syllA.03", "Cor"),
	vibe.E("mapmodel.world.syllA.04", "Dun"),
	vibe.E("mapmodel.world.syllA.05", "Eld"),
	vibe.E("mapmodel.world.syllA.06", "Fen"),
	vibe.E("mapmodel.world.syllA.07", "Gal"),
	vibe.E("mapmodel.world.syllA.08", "Hol"),
	vibe.E("mapmodel.world.syllA.09", "Ir"),
	vibe.E("mapmodel.world.syllA.10", "Kest"),
	vibe.E("mapmodel.world.syllA.11", "Lun"),
	vibe.E("mapmodel.world.syllA.12", "Mor"),
	vibe.E("mapmodel.world.syllA.13", "Nor"),
	vibe.E("mapmodel.world.syllA.14", "Ost"),
	vibe.E("mapmodel.world.syllA.15", "Rav"),
	vibe.E("mapmodel.world.syllA.16", "Sel"),
	vibe.E("mapmodel.world.syllA.17", "Thorn"),
	vibe.E("mapmodel.world.syllA.18", "Umb"),
	vibe.E("mapmodel.world.syllA.19", "Vel"),
	vibe.E("mapmodel.world.syllA.20", "Wyn"),
}

var syllB = []string{"vale", "ford", "hold", "mere", "wick", "stead", "haven", "gate", "moor", "crest", "fall", "reach"}

// cityName is the name the town is given: a pure function of the run seed.
func cityName(seed int64) string {
	return syllA[Hash(seed, 5)%uint64(len(syllA))].String() + syllB[Hash(seed, 6)%uint64(len(syllB))]
}
