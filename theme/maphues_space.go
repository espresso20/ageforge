package theme

import "github.com/gdamore/tcell/v2"

// maphues_space.go holds the sky arc's colours: the palettes the maps draw
// from the Space Age on, when both styles leave the ground (mapmodel/sky.go
// and mapmodel/palette_space.go say which age uses which). Like the other
// map hues they are never drawn as they are: the roguelike makes each one
// legible on the scene's void and the skyline lights and hazes it, both lift
// the scene toward the page on light themes and fold it to a duotone on the
// monochrome ones. Keeping the table here keeps "no raw colours outside
// theme/" true for ui/.

// SpaceHue names a sky arc colour.
type SpaceHue uint8

const (
	SpaceNone SpaceHue = iota

	// the Space Age: black space, a blue planet, steel stations, amber lights
	SpaceVoid
	SpaceVoidHaze
	SpaceStar
	SpaceStarBright
	SpaceStarWarm
	SpacePlanetOcean
	SpacePlanetLand
	SpacePlanetCloud
	SpaceAtmosphere
	SpaceAtmosphereHaze
	SpaceCityAmber
	SpaceCityGold
	SpaceWhite
	SpaceSteel
	SpaceSteelDark
	SpaceSolar
	SpaceHydro
	SpaceBeaconRed
	SpaceRock
	SpaceRockDark
	SpaceMoon
	SpaceMoonDark

	// the Interstellar Age: the first colour in space, violet and teal
	SpaceVoidDeep
	SpaceVoidDeepHaze
	SpaceSun
	SpaceSunCorona
	SpaceOrbitLine
	SpaceNebulaViolet
	SpaceNebulaTeal
	SpaceEngine
	SpaceEngineFade
	SpaceGateSteel
	SpaceGateField
	SpaceColonyBeacon
	SpaceArkHull
	SpaceColonyBlue
	SpaceColonyRust
	SpaceColonyGreen

	// the Galactic Age: rich gold, cyan and crimson
	SpaceVoidGalaxy
	SpaceSpiral
	SpaceSpiralCore
	SpaceGold
	SpaceCyan
	SpaceCrimson
	SpaceStarbase
	SpaceStarbaseDim
	SpaceWarp
	SpaceNebulaMagenta
	SpaceNebulaBlue
	SpaceNebulaGold
	SpaceAlienGreen
	SpaceAlienViolet
	SpaceAlienAmber
	SpaceTransporter

	// the Quantum Age: shifting iridescence over a dark field
	SpaceVoidQuantum
	SpaceIri0
	SpaceIri1
	SpaceIri2
	SpaceIri3
	SpaceIri4
	SpaceIri5
	SpaceIri6
	SpaceIri7
	SpaceEcho
	SpaceProbability
	SpaceProbabilityHi

	// the Transcendent Age: white and gold light on deep indigo
	SpaceIndigo
	SpaceIndigoHaze
	SpaceLight
	SpaceLightGold
	SpaceLightDim
	SpaceLightCore

	numSpaceHues
)

// NumSpaceHues is the size of the sky arc's colour table.
const NumSpaceHues = int(numSpaceHues)

var spaceHues = [numSpaceHues]int32{
	SpaceNone: 0x808080,

	SpaceVoid: 0x04060c, SpaceVoidHaze: 0x0c1528, SpaceStar: 0xb8c6e6, SpaceStarBright: 0xffffff,
	SpaceStarWarm: 0xffe6b8, SpacePlanetOcean: 0x0b2648, SpacePlanetLand: 0x1b3226, SpacePlanetCloud: 0x46628a,
	SpaceAtmosphere: 0x4fb0ff, SpaceAtmosphereHaze: 0x1f5a9a, SpaceCityAmber: 0xffb347, SpaceCityGold: 0xffdc8a,
	SpaceWhite: 0xeef3fa, SpaceSteel: 0x9aa8bc, SpaceSteelDark: 0x58667a, SpaceSolar: 0x3f78d8,
	SpaceHydro: 0x7fd88e, SpaceBeaconRed: 0xff4a3a, SpaceRock: 0x8f8172, SpaceRockDark: 0x5c5248,
	SpaceMoon: 0xc4c0b6, SpaceMoonDark: 0x7e7b74,

	SpaceVoidDeep: 0x08040f, SpaceVoidDeepHaze: 0x170b29, SpaceSun: 0xffd27a, SpaceSunCorona: 0xff9440,
	SpaceOrbitLine: 0x4c4676, SpaceNebulaViolet: 0x8a46d0, SpaceNebulaTeal: 0x2ec0b0, SpaceEngine: 0x78f0ff,
	SpaceEngineFade: 0x23707e, SpaceGateSteel: 0xc6cee2, SpaceGateField: 0x5cf4dc, SpaceColonyBeacon: 0xa6ff8e,
	SpaceArkHull: 0xdbe2ec, SpaceColonyBlue: 0x4c74a4, SpaceColonyRust: 0xa4704c, SpaceColonyGreen: 0x5c9a64,

	SpaceVoidGalaxy: 0x03050f, SpaceSpiral: 0x3c3468, SpaceSpiralCore: 0x8a78c8, SpaceGold: 0xffc94a,
	SpaceCyan: 0x40e8ff, SpaceCrimson: 0xec3a54, SpaceStarbase: 0xdfe3ee, SpaceStarbaseDim: 0x8c93a8,
	SpaceWarp: 0xa8f6ff, SpaceNebulaMagenta: 0xc63e90, SpaceNebulaBlue: 0x2e60c8, SpaceNebulaGold: 0xc4902e,
	SpaceAlienGreen: 0x6cff6c, SpaceAlienViolet: 0xcc84ff, SpaceAlienAmber: 0xffa444, SpaceTransporter: 0xc4f8ff,

	SpaceVoidQuantum: 0x060509, SpaceIri0: 0xff5ad8, SpaceIri1: 0xb46cff, SpaceIri2: 0x5c8eff,
	SpaceIri3: 0x3ce2ff, SpaceIri4: 0x4cffb4, SpaceIri5: 0xcaff4c, SpaceIri6: 0xffda4c, SpaceIri7: 0xff8c5c,
	SpaceEcho: 0x34344a, SpaceProbability: 0x3c2c62, SpaceProbabilityHi: 0x6a58a6,

	SpaceIndigo: 0x150d38, SpaceIndigoHaze: 0x241a56, SpaceLight: 0xfff6e0, SpaceLightGold: 0xffd479,
	SpaceLightDim: 0x8c7ccc, SpaceLightCore: 0xffffff,
}

// SpaceColor returns a sky arc colour.
func SpaceColor(h SpaceHue) tcell.Color {
	if h >= numSpaceHues {
		h = SpaceNone
	}
	return tcell.NewHexColor(spaceHues[h])
}
