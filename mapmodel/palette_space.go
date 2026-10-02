package mapmodel

import "github.com/espresso20/ageforge/theme"

// palette_space.go gives each sky scene (sky.go) its palette, so the two
// styles agree on what an age looks like: black space and a blue planet in
// the Space Age, violet and teal in the Interstellar, gold, cyan and crimson
// in the Galactic, iridescence in the Quantum, white and gold light on
// indigo in the Transcendent. A palette names theme space hues by role (an
// ink); the colours themselves live in theme/maphues_space.go, and each
// style resolves them against the active theme (legible on the void, lifted
// toward the page on light themes, folded to a duotone on monochrome ones).

// SkyInk is a colour role in a sky scene.
type SkyInk uint8

const (
	InkVoid       SkyInk = iota // the background: space itself
	InkHaze                     // a faint lift of the void: the far glow, a nebula's floor
	InkStar                     // an ordinary star
	InkStarBright               // a bright star, a twinkle at its peak
	InkFrame                    // the main structure: stations, hulls, the gate, the starbase
	InkFrameDim                 // struts, scaffolds, older sections
	InkLight                    // lights: windows, city lights, beacons
	InkGlow                     // the age's signature glow
	InkAccent                   // first accent
	InkAccent2                  // second accent
	InkAccent3                  // third accent
	InkSurface                  // a planet's oceans, a colony's face
	InkSurface2                 // continents, a second colony face
	InkCloud                    // clouds, the first nebula
	InkCloud2                   // the second nebula
	InkRock                     // asteroids, rocky worlds
	InkRockDark                 // craters, shadowed rock
	InkMoon                     // the moon
	InkEcho                     // the faintest marks: time echoes, empty rings
	NumSkyInks
)

// SkyPalette is a sky scene's palette: a theme space hue per ink.
type SkyPalette [NumSkyInks]theme.SpaceHue

var skyPalettes = [NumSkyScenes]SkyPalette{
	SkyOrbit: {
		InkVoid: theme.SpaceVoid, InkHaze: theme.SpaceVoidHaze, InkStar: theme.SpaceStar,
		InkStarBright: theme.SpaceStarBright, InkFrame: theme.SpaceWhite, InkFrameDim: theme.SpaceSteel,
		InkLight: theme.SpaceCityAmber, InkGlow: theme.SpaceAtmosphere, InkAccent: theme.SpaceSolar,
		InkAccent2: theme.SpaceHydro, InkAccent3: theme.SpaceBeaconRed, InkSurface: theme.SpacePlanetOcean,
		InkSurface2: theme.SpacePlanetLand, InkCloud: theme.SpacePlanetCloud, InkCloud2: theme.SpaceAtmosphereHaze,
		InkRock: theme.SpaceRock, InkRockDark: theme.SpaceRockDark, InkMoon: theme.SpaceMoon,
		InkEcho: theme.SpaceSteelDark,
	},
	SkyDeep: {
		InkVoid: theme.SpaceVoidDeep, InkHaze: theme.SpaceVoidDeepHaze, InkStar: theme.SpaceStar,
		InkStarBright: theme.SpaceStarBright, InkFrame: theme.SpaceGateSteel, InkFrameDim: theme.SpaceOrbitLine,
		InkLight: theme.SpaceColonyBeacon, InkGlow: theme.SpaceEngine, InkAccent: theme.SpaceSun,
		InkAccent2: theme.SpaceGateField, InkAccent3: theme.SpaceSunCorona, InkSurface: theme.SpaceColonyBlue,
		InkSurface2: theme.SpaceColonyGreen, InkCloud: theme.SpaceNebulaViolet, InkCloud2: theme.SpaceNebulaTeal,
		InkRock: theme.SpaceColonyRust, InkRockDark: theme.SpaceRockDark, InkMoon: theme.SpaceArkHull,
		InkEcho: theme.SpaceEngineFade,
	},
	SkyGalaxy: {
		InkVoid: theme.SpaceVoidGalaxy, InkHaze: theme.SpaceSpiral, InkStar: theme.SpaceStar,
		InkStarBright: theme.SpaceStarBright, InkFrame: theme.SpaceStarbase, InkFrameDim: theme.SpaceStarbaseDim,
		InkLight: theme.SpaceGold, InkGlow: theme.SpaceWarp, InkAccent: theme.SpaceGold,
		InkAccent2: theme.SpaceCyan, InkAccent3: theme.SpaceCrimson, InkSurface: theme.SpaceNebulaBlue,
		InkSurface2: theme.SpaceNebulaGold, InkCloud: theme.SpaceNebulaMagenta, InkCloud2: theme.SpaceNebulaBlue,
		InkRock: theme.SpaceSpiralCore, InkRockDark: theme.SpaceSpiral, InkMoon: theme.SpaceTransporter,
		InkEcho: theme.SpaceSpiral,
	},
	SkyQuantum: {
		InkVoid: theme.SpaceVoidQuantum, InkHaze: theme.SpaceProbability, InkStar: theme.SpaceStar,
		InkStarBright: theme.SpaceStarBright, InkFrame: theme.SpaceIri3, InkFrameDim: theme.SpaceIri1,
		InkLight: theme.SpaceIri6, InkGlow: theme.SpaceIri0, InkAccent: theme.SpaceIri4,
		InkAccent2: theme.SpaceIri2, InkAccent3: theme.SpaceIri7, InkSurface: theme.SpaceProbability,
		InkSurface2: theme.SpaceProbabilityHi, InkCloud: theme.SpaceProbability, InkCloud2: theme.SpaceProbabilityHi,
		InkRock: theme.SpaceIri5, InkRockDark: theme.SpaceIri1, InkMoon: theme.SpaceIri6,
		InkEcho: theme.SpaceEcho,
	},
	SkyMandala: {
		InkVoid: theme.SpaceIndigo, InkHaze: theme.SpaceIndigoHaze, InkStar: theme.SpaceLightDim,
		InkStarBright: theme.SpaceLight, InkFrame: theme.SpaceLightGold, InkFrameDim: theme.SpaceLightDim,
		InkLight: theme.SpaceLight, InkGlow: theme.SpaceLightCore, InkAccent: theme.SpaceLightGold,
		InkAccent2: theme.SpaceLight, InkAccent3: theme.SpaceLightDim, InkSurface: theme.SpaceIndigoHaze,
		InkSurface2: theme.SpaceLightDim, InkCloud: theme.SpaceIndigoHaze, InkCloud2: theme.SpaceLightDim,
		InkRock: theme.SpaceLightGold, InkRockDark: theme.SpaceLightDim, InkMoon: theme.SpaceLight,
		InkEcho: theme.SpaceIndigoHaze,
	},
}

// SkyPaletteOf is a sky scene's palette (the ground has none: its colours
// come from the map classes).
func SkyPaletteOf(s SkyScene) SkyPalette {
	if s >= NumSkyScenes {
		return SkyPalette{}
	}
	return skyPalettes[s]
}

// SkyIridescent is the Quantum Age's shifting ramp: structures run through
// it cell by cell and frame by frame.
var SkyIridescent = [8]theme.SpaceHue{theme.SpaceIri0, theme.SpaceIri1, theme.SpaceIri2, theme.SpaceIri3,
	theme.SpaceIri4, theme.SpaceIri5, theme.SpaceIri6, theme.SpaceIri7}
