package theme

import "github.com/gdamore/tcell/v2"

// maphues_skyline.go holds the skyline map's scene colours: building
// materials per period, the sky keyframes per epoch and the fixed nature and
// light colours. Like the other map hues they are never drawn as they are:
// the skyline lights them for the time of day, hazes them by depth, lifts
// them toward the page on light themes and folds them to a duotone on the
// monochrome ones. Keeping the tables here keeps "no raw colours outside
// theme/" true for ui/.

// SkyFamilies and SkyVariants size the material table: eleven period
// families (primitive … cosmic), three finishes each.
const (
	SkyFamilies = 11
	SkyVariants = 3
)

// SkyMaterial is one building finish.
type SkyMaterial struct {
	Wall, Roof, Trim, Metal, Glass, GlassHi, Neon1, Neon2, Neon3, Glow tcell.Color
}

type rawMaterial [10]int32

var skyMaterials = [SkyFamilies][SkyVariants]rawMaterial{
	// 0 primitive/stone: hide, wood, thatch
	{{0x7a5236, 0xa8843f, 0xd8c8a0, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a},
		{0x6a4630, 0xc19a4b, 0xcbb892, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a},
		{0x8a6242, 0x9a7a3a, 0xe0d2ae, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a}},
	// 1 bronze/iron: mudbrick, flat roofs
	{{0xb38456, 0x8a6a45, 0xe0cfa4, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a},
		{0xc49a6c, 0x7d5e3e, 0xeadcb8, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a},
		{0x9d9384, 0x7a6a55, 0xd8ccb0, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a}},
	// 2 classical: marble, terracotta
	{{0xe2dccb, 0xb4553a, 0xfff6e2, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050},
		{0xd6cdb8, 0xa04a34, 0xf4ecd8, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050},
		{0xcfc3a8, 0xc2653f, 0xfaf0dc, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050}},
	// 3 medieval: plaster and beams, rubble stone, slate
	{{0xd3c6a6, 0x6f3a2a, 0x4a3423, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050},
		{0x8f8a80, 0x4d5260, 0x5d5850, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050},
		{0x9c958a, 0x5a3a30, 0x6a6258, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050}},
	// 4 renaissance/colonial: ochre stucco, copper domes, whitewash
	{{0xd6a861, 0x9b4a2f, 0xefe3c4, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050},
		{0xe6dfd0, 0x4a5a6a, 0xffffff, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050},
		{0xc98f5f, 0x7a3a28, 0xf0e2c0, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050}},
	// 5 industrial/victorian: brick, soot, iron
	{{0x8f3c2a, 0x3c3838, 0xb8a48a, 0x55575c, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a},
		{0x6f3226, 0x2e2c2c, 0xa89880, 0x4a4c50, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a},
		{0x7c6a5a, 0x3a3a3e, 0xc0b098, 0x55575c, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a}},
	// 6 electric/atomic: limestone deco, concrete, chrome
	{{0xc9bfa6, 0x5e5e62, 0xc7a24a, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0},
		{0x9a9a95, 0x4e4e52, 0xd0d0c8, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0},
		{0xb0a48c, 0x5a5048, 0xe0c870, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0}},
	// 7 modern/information/digital: concrete and glass
	{{0x8e9499, 0x5a6066, 0xc8ced4, 0xb0b7bd, 0x3f6f96, 0x8fc0e0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff},
		{0x4a6a86, 0x3a4a5a, 0xa8c8e0, 0xb0b7bd, 0x2f5f86, 0x9fd0f0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff},
		{0x6e767e, 0x4a5056, 0xd8dde2, 0xb0b7bd, 0x35607a, 0x7fb0d0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff}},
	// 8 cyberpunk: dark towers, neon
	{{0x2c2a45, 0x1d1b2e, 0x4a4570, 0x5a5878, 0x3a2f63, 0x6a5aa3, 0xff3ea5, 0x29f0ff, 0xb4ff39, 0xff9ad5},
		{0x23213a, 0x16142a, 0x3d3a60, 0x5a5878, 0x2a3f63, 0x5a8ab3, 0x29f0ff, 0xff3ea5, 0xffe23a, 0x9af5ff},
		{0x34304a, 0x201c30, 0x524a78, 0x5a5878, 0x402a58, 0x8a5aa3, 0xb4ff39, 0xff3ea5, 0x29f0ff, 0xe0ffa0}},
	// 9 fusion/space: white composites, glass, cyan glow
	{{0xd9dee6, 0xaab4c2, 0xf4f8ff, 0x8a96a8, 0x4fa3c9, 0xaee8f8, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff},
		{0xc2cad6, 0x8894a6, 0xe8eef8, 0x8a96a8, 0x3f8fb9, 0x9edcf0, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff},
		{0xe6e2d8, 0xb0a894, 0xffffff, 0x8a96a8, 0x5fb3a9, 0xbef0e8, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff}},
	// 10 cosmic: iridescent alloys, violet and gold light
	{{0xe4e0f6, 0xb9b2e0, 0xfff4d0, 0x9a92c8, 0x5a4ab0, 0xc8b8ff, 0xb28cff, 0xffd479, 0x7ff0ff, 0xe8d8ff},
		{0xc9c2ec, 0x9a90d0, 0xffe8a8, 0x9a92c8, 0x4a3aa0, 0xb8a8ff, 0xffd479, 0xb28cff, 0x7ff0ff, 0xfff0c0},
		{0xf0ecff, 0xd0c8f0, 0xd8f8ff, 0x9a92c8, 0x3a6ab0, 0xa8d8ff, 0x7ff0ff, 0xb28cff, 0xffd479, 0xd8f8ff}},
}

// SkyMaterialFor returns a finish (family and variant are clamped).
func SkyMaterialFor(family, variant int) SkyMaterial {
	family = clampIdx(family, SkyFamilies)
	variant = clampIdx(variant, SkyVariants)
	r := skyMaterials[family][variant]
	h := tcell.NewHexColor
	return SkyMaterial{h(r[0]), h(r[1]), h(r[2]), h(r[3]), h(r[4]), h(r[5]), h(r[6]), h(r[7]), h(r[8]), h(r[9])}
}

// SkyGradient is a four-stop sky, zenith to horizon.
type SkyGradient [4]tcell.Color

// SkyKeys are an epoch's sky keyframes.
type SkyKeys struct{ Day, Dusk, Dawn, Night SkyGradient }

var skyKeys = [7][4][4]int32{
	{{0x2f6fd1, 0x5a9be6, 0x9fcdf2, 0xdcefff}, {0x1c1b4a, 0x5b3072, 0xd0585a, 0xffb257},
		{0x243a78, 0x6a5a9a, 0xe0889a, 0xffc890}, {0x03060f, 0x0a1330, 0x16224a, 0x27305c}},
	{{0x2a68c8, 0x5898e0, 0xa0cdf0, 0xe8f0f0}, {0x1c1b4a, 0x5b3072, 0xd8604a, 0xffba60},
		{0x243a78, 0x6a5a9a, 0xe0889a, 0xffc890}, {0x03060f, 0x0a1330, 0x16224a, 0x2a3058}},
	{{0x5a7896, 0x7f93a6, 0xb2b0a4, 0xd8c8a8}, {0x2a2240, 0x6a3a5a, 0xc0604a, 0xe89a58},
		{0x3a4060, 0x7a6a80, 0xc08a80, 0xe0b080}, {0x06070c, 0x121624, 0x262430, 0x4a3428}},
	{{0x3f6fa8, 0x6a92bc, 0xa8bccc, 0xdcd6c4}, {0x201c46, 0x60306a, 0xd05a50, 0xffa860},
		{0x2a3a70, 0x6a5a90, 0xd08890, 0xffc088}, {0x04060e, 0x0e1428, 0x1e2440, 0x42384a}},
	{{0x2a6ad8, 0x4f94ec, 0x98caf6, 0xe0f2ff}, {0x1a1a50, 0x582f78, 0xe0585e, 0xffb060},
		{0x243a80, 0x6a5aa0, 0xe888a0, 0xffc890}, {0x02040c, 0x0a1230, 0x16204a, 0x2a3a66}},
	{{0x1f5d7a, 0x3f8a98, 0x88b8b0, 0xd7e6c8}, {0x1a0c3a, 0x5a1a6a, 0xd0306a, 0xff8a50},
		{0x1a1a50, 0x4a2a78, 0xb04a8a, 0xff9a80}, {0x07031a, 0x1c0838, 0x40104f, 0x8a1a60}},
	{{0x14103a, 0x3a3a9a, 0x8a8fe0, 0xe0d8ff}, {0x0e0826, 0x3a1a60, 0xa04a90, 0xffa080},
		{0x0e0e30, 0x3a3a80, 0x9a70c0, 0xffc0a0}, {0x020108, 0x0a0620, 0x1a0e3a, 0x3a1e5a}},
}

// SkyKeysFor returns an epoch's sky keyframes (0 Stone … 6 Cosmic).
func SkyKeysFor(epoch int) SkyKeys {
	k := skyKeys[clampIdx(epoch, len(skyKeys))]
	g := func(v [4]int32) SkyGradient {
		return SkyGradient{tcell.NewHexColor(v[0]), tcell.NewHexColor(v[1]), tcell.NewHexColor(v[2]), tcell.NewHexColor(v[3])}
	}
	return SkyKeys{Day: g(k[0]), Dusk: g(k[1]), Dawn: g(k[2]), Night: g(k[3])}
}

// SkyHue names a fixed skyline scene colour.
type SkyHue uint8

const (
	SkyLeaf SkyHue = iota
	SkyLeafDark
	SkyTrunk
	SkyField1
	SkyField2
	SkyRock
	SkyRockDark
	SkyInk
	SkyWater
	SkyWaterDeep
	SkySmoke
	SkySmokeHeavy
	SkySmokeBlack
	SkyFurnace
	SkyFire
	SkyFireHot
	SkyFireDeep
	SkyFireGold
	SkyBeacon
	SkyWinWarm
	SkyWinPale
	SkyWinNeon
	SkyWhite
	SkyLightNight
	SkyLightDay
	SkyLightDusk
	SkyBlueHour
	SkyBruise
	SkyBruiseHigh
	SkyStarWarm
	SkyStarCool
	SkySun
	SkySunLow
	SkyMoon
	SkyMoonPale
	SkyCloud
	SkyCloudDusk
	SkyRidge
	SkyRidgeCosmic
	SkySnowCap
	SkyHillYoung
	SkyHillIndustrial
	SkyHillModern
	SkyForest
	SkyTownFar
	SkyRidgeTown
	SkyNeonPink
	SkyNeonCyan
	SkyNeonMagenta
	SkyGroundGrass
	SkyGroundSoil
	SkyGroundCobble
	SkyGroundCobbleDark
	SkyGroundBrick
	SkyGroundBrickDark
	SkyGroundSlab
	SkyGroundSlabDark
	SkyRoadDirt
	SkyRoadDirtMark
	SkyRoadCobble
	SkyRoadCobbleMark
	SkyRoadRail
	SkyRoadRailMark
	SkyRoadAsphalt
	SkyRoadAsphaltMark
	SkyRoadNeon
	SkyRoadNeonMark
	SkyRoadLight
	SkyRoadLightMark
	SkyVerge
	SkyVergeDark
	SkyVergeUrban
	SkyVergeUrbanDark
	SkyQuay
	SkyLampPost
	SkyLamp
	SkyHeadlight
	SkyCrane
	SkyScaffold
	SkyRain
	SkyHarbinger
	SkyHarbingerPale
	SkyHarbingerHolo
	SkyHarbingerEyes
	SkyCrack
	SkyPlanet
	SkyPlanetRing
	SkyOrbitalRing
	SkyOrbitalLight
	SkyTradeGold
	SkyWarRed
	SkyHull
	SkySail
	SkyHullSteam
	SkyFunnel
	SkySteam
	SkyHullModern
	SkyCargoRed
	SkyCargoBlue
	SkyCargoGold
	SkyCargoGreen
	SkyNavLight
	SkyAircraft
	SkyAircraftLate
	SkyBird
	SkyWonderGold
	SkyOx
	SkyHorse
	SkyCanvas
	SkyBalloon
	SkyZeppelin
	SkyTram
	SkyCarRed
	SkyCarBlue
	SkyCarWhite
	SkyCarYellow
	SkyRail
	SkyWarp
	SkyExhaust
	numSkyHues
)

var skyHues = [numSkyHues]int32{
	SkyLeaf: 0x3f7a3a, SkyLeafDark: 0x2a5530, SkyTrunk: 0x5a3d26, SkyField1: 0x9fae4a, SkyField2: 0xc9a84a,
	SkyRock: 0x8a8580, SkyRockDark: 0x5e5a58, SkyInk: 0x141218, SkyWater: 0x2d5f8a, SkyWaterDeep: 0x16304a,
	SkySmoke: 0xb8b4b0, SkySmokeHeavy: 0x6a6460, SkySmokeBlack: 0x1e1a1a, SkyFurnace: 0xff7a30,
	SkyFire: 0xff7a1a, SkyFireHot: 0xffd23a, SkyFireDeep: 0xff5a1a, SkyFireGold: 0xffb03a,
	SkyBeacon: 0xff3030, SkyWinWarm: 0xffd27a, SkyWinPale: 0xfff0c0, SkyWinNeon: 0x9af5ff, SkyWhite: 0xffffff,
	SkyLightNight: 0x5a6690, SkyLightDay: 0xfffaf0, SkyLightDusk: 0xe8a080, SkyBlueHour: 0x8a90bc,
	SkyBruise: 0xa0302a, SkyBruiseHigh: 0x6a2030, SkyStarWarm: 0xfff6e0, SkyStarCool: 0x9ab8ff,
	SkySun: 0xfffbe8, SkySunLow: 0xffa040, SkyMoon: 0xe8ecf8, SkyMoonPale: 0xfaf8f0,
	SkyCloud: 0xf4f4f8, SkyCloudDusk: 0xff9a80, SkyRidge: 0x4a5a7e, SkyRidgeCosmic: 0x5a4a8a, SkySnowCap: 0xf0f4ff,
	SkyHillYoung: 0x2f5234, SkyHillIndustrial: 0x4a4e4a, SkyHillModern: 0x3a4a5a, SkyForest: 0x1e3a26,
	SkyTownFar: 0x404858, SkyRidgeTown: 0x3a4668, SkyNeonPink: 0xff3ea5, SkyNeonCyan: 0x29f0ff, SkyNeonMagenta: 0xff6ac8,
	SkyGroundGrass: 0x5d8a3a, SkyGroundSoil: 0x6a4a2e, SkyGroundCobble: 0x8a8478, SkyGroundCobbleDark: 0x5a5448,
	SkyGroundBrick: 0x6a5a50, SkyGroundBrickDark: 0x3e3834, SkyGroundSlab: 0x9aa0a8, SkyGroundSlabDark: 0x4a4e56,
	SkyRoadDirt: 0x4f7a32, SkyRoadDirtMark: 0x7a9a4a, SkyRoadCobble: 0x8a7050, SkyRoadCobbleMark: 0xb09a70,
	SkyRoadRail: 0x4a4440, SkyRoadRailMark: 0x8a8a8a, SkyRoadAsphalt: 0x3a3c40, SkyRoadAsphaltMark: 0xe0d890,
	SkyRoadNeon: 0x24222e, SkyRoadNeonMark: 0xff3ea5, SkyRoadLight: 0x2a2a40, SkyRoadLightMark: 0x7ff0ff,
	SkyVerge: 0x3f6a2e, SkyVergeDark: 0x2a4a22, SkyVergeUrban: 0x8a8e94, SkyVergeUrbanDark: 0x5a5e64,
	SkyQuay: 0x6a6258, SkyLampPost: 0x3a3a40, SkyLamp: 0xffe0a0, SkyHeadlight: 0xfff0c0,
	SkyCrane: 0xe0b020, SkyScaffold: 0x8a8070, SkyRain: 0xa8b8d0,
	SkyHarbinger: 0x0c0a10, SkyHarbingerPale: 0x1a1420, SkyHarbingerHolo: 0x7ff0ff, SkyHarbingerEyes: 0xff4a3a,
	SkyCrack: 0xf0e8ff, SkyPlanet: 0xc8a0e0, SkyPlanetRing: 0xf0d8a0, SkyOrbitalRing: 0xb8c8e0, SkyOrbitalLight: 0xfff0a0,
	SkyTradeGold: 0xe0b020, SkyWarRed: 0xff2a2a, SkyHull: 0x4a3020, SkySail: 0xf0ead8, SkyHullSteam: 0x2a2a2e,
	SkyFunnel: 0xa03020, SkySteam: 0x9a9a9a, SkyHullModern: 0x3a4a5a, SkyCargoRed: 0xc04a2a, SkyCargoBlue: 0x2a7ac0,
	SkyCargoGold: 0xe0b020, SkyCargoGreen: 0x3a9a4a, SkyNavLight: 0x60ff60, SkyAircraft: 0xd8dce4,
	SkyAircraftLate: 0xe8f8ff, SkyBird: 0x2a2a30, SkyWonderGold: 0xffd479, SkyOx: 0x6a4a30, SkyHorse: 0x5a3a24,
	SkyCanvas: 0xe8dcc0, SkyBalloon: 0xd0503a, SkyZeppelin: 0xb0b4b8, SkyTram: 0x3a7a5a, SkyCarRed: 0xc03a30,
	SkyCarBlue: 0x3a6ac0, SkyCarWhite: 0xe0e4e8, SkyCarYellow: 0xe0c040, SkyRail: 0x6a6e78, SkyWarp: 0xd8f0ff,
	SkyExhaust: 0xffc070,
}

// SkyColor returns a fixed skyline scene colour.
func SkyColor(h SkyHue) tcell.Color {
	if h >= numSkyHues {
		h = SkyWhite
	}
	return tcell.NewHexColor(skyHues[h])
}

// Shade scales a colour's channels by k (a light level: 0.7 is a shaded
// face, 1.15 a sunlit one), clamped to the channel range.
func Shade(c tcell.Color, k float64) tcell.Color {
	r, g, b := c.RGB()
	if r < 0 {
		return c
	}
	s := func(v int32) int32 {
		x := int32(float64(float64(v)*k) + 0.5)
		if x < 0 {
			return 0
		}
		if x > 255 {
			return 255
		}
		return x
	}
	return tcell.NewRGBColor(s(r), s(g), s(b))
}

// Tint multiplies a colour channel by channel by a light colour (white
// leaves it unchanged, a blue night light cools it).
func Tint(c, light tcell.Color) tcell.Color {
	r, g, b := c.RGB()
	lr, lg, lb := light.RGB()
	if r < 0 || lr < 0 {
		return c
	}
	return tcell.NewRGBColor(r*lr/255, g*lg/255, b*lb/255)
}

func clampIdx(i, n int) int {
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}
