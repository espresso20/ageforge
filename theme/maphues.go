package theme

import "github.com/gdamore/tcell/v2"

// maphues.go holds the maps' identity hues: the few fixed colours that make
// water read as water and war read as war on every theme. They are never
// drawn as they are. A map colour class starts from a theme role, leans
// toward one of these hues by a set amount, and is then clamped legible
// against the theme's background (see ui/mapstyle.Palette), so a theme
// switch retints the maps for free and light themes need no re-keying.
// Keeping the hues here keeps the "no raw colours outside theme/" rule.

// MapHue names a map identity hue.
type MapHue uint8

const (
	HueNone MapHue = iota
	HueGrass
	HueForest
	HueWater
	HueWaterDeep
	HueRock
	HueHill
	HueStone // walls, towers
	HueGold  // trade, stores, wonders
	HueCivic // knowledge, faith, culture
	HueMilitary
	HueIdle
	HueDanger
	HueAlly
	HueFriend
	HueNeutral
	HueRival
	HueWar
	HueFresh
	// production hue per epoch (Stone … Cosmic)
	HueEpochStone
	HueEpochIron
	HueEpochSteel
	HueEpochElectric
	HueEpochDigital
	HueEpochNeon
	HueEpochCosmic
	numMapHues
)

var mapHues = [numMapHues]int32{
	HueNone:          0x808080,
	HueGrass:         0x8a9a5a,
	HueForest:        0x3f9a4f,
	HueWater:         0x3a86c8,
	HueWaterDeep:     0x2a6ab0,
	HueRock:          0x9a8a78,
	HueHill:          0x9a8a5a,
	HueStone:         0xb0a898,
	HueGold:          0xe0b040,
	HueCivic:         0x8a7ae0,
	HueMilitary:      0xc05050,
	HueIdle:          0xe0a030,
	HueDanger:        0xff3a2a,
	HueAlly:          0x4fc060,
	HueFriend:        0xe0c040,
	HueNeutral:       0x6a90b8,
	HueRival:         0xe08030,
	HueWar:           0xe03030,
	HueFresh:         0x50c070,
	HueEpochStone:    0xc08a4a,
	HueEpochIron:     0xa8a29a,
	HueEpochSteel:    0xc0603a,
	HueEpochElectric: 0xe0b83a,
	HueEpochDigital:  0x3ab8d0,
	HueEpochNeon:     0xe040c0,
	HueEpochCosmic:   0x9a7ae0,
}

// MapHueColor returns a map identity hue.
func MapHueColor(h MapHue) tcell.Color {
	if h >= numMapHues {
		h = HueNone
	}
	return tcell.NewHexColor(mapHues[h])
}

// EpochHue is the production hue of an epoch (0 Stone … 6 Cosmic).
func EpochHue(epoch int) MapHue {
	if epoch < 0 {
		epoch = 0
	}
	if epoch > 6 {
		epoch = 6
	}
	return HueEpochStone + MapHue(epoch)
}
