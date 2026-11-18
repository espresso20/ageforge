package mapmodel

import "github.com/espresso20/ageforge/rowid"

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
var syllA = []rowid.Entry{
	rowid.E("mapmodel.world.syllA.01", "Ash"),
	rowid.E("mapmodel.world.syllA.02", "Bram"),
	rowid.E("mapmodel.world.syllA.03", "Cor"),
	rowid.E("mapmodel.world.syllA.04", "Dun"),
	rowid.E("mapmodel.world.syllA.05", "Eld"),
	rowid.E("mapmodel.world.syllA.06", "Fen"),
	rowid.E("mapmodel.world.syllA.07", "Gal"),
	rowid.E("mapmodel.world.syllA.08", "Hol"),
	rowid.E("mapmodel.world.syllA.09", "Ir"),
	rowid.E("mapmodel.world.syllA.10", "Kest"),
	rowid.E("mapmodel.world.syllA.11", "Lun"),
	rowid.E("mapmodel.world.syllA.12", "Mor"),
	rowid.E("mapmodel.world.syllA.13", "Nor"),
	rowid.E("mapmodel.world.syllA.14", "Ost"),
	rowid.E("mapmodel.world.syllA.15", "Rav"),
	rowid.E("mapmodel.world.syllA.16", "Sel"),
	rowid.E("mapmodel.world.syllA.17", "Thorn"),
	rowid.E("mapmodel.world.syllA.18", "Umb"),
	rowid.E("mapmodel.world.syllA.19", "Vel"),
	rowid.E("mapmodel.world.syllA.20", "Wyn"),
}

var syllB = []string{"vale", "ford", "hold", "mere", "wick", "stead", "haven", "gate", "moor", "crest", "fall", "reach"}

// TownNameOf is the name a run's seed gives its town, without laying the
// map out (the epoch screen names the town under the banner).
func TownNameOf(seed int64) string { return cityName(seed) }

// cityName is the name the town is given: a pure function of the run seed.
func cityName(seed int64) string {
	return syllA[Hash(seed, 5)%uint64(len(syllA))].String() + syllB[Hash(seed, 6)%uint64(len(syllB))]
}
