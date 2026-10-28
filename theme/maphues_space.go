package theme

import "github.com/gdamore/tcell/v2"

// maphues_space.go holds the sky arc's colours: the palettes the map draws
// from the Space Age on, when it leaves the ground. Like the other map hues
// they are never drawn as they are: the map makes each one legible on the
// scene's void. Keeping the table here keeps "no raw colours outside theme/"
// true for ui/.

// SpaceHue names a sky arc colour.
type SpaceHue uint8

const (
	SpaceNone SpaceHue = iota

	// the Space Age: black space, a blue planet, steel stations, amber lights
	SpaceVoid
	SpaceStar
	SpacePlanetOcean
	SpacePlanetLand
	SpacePlanetCloud
	SpaceAtmosphere
	SpaceCityAmber
	SpaceSolar
	SpaceHydro
	SpaceRock

	// the Interstellar Age: the first colour in space, violet and teal
	SpaceSun
	SpaceSunCorona
	SpaceOrbitLine
	SpaceNebulaViolet
	SpaceEngine

	numSpaceHues
)

var spaceHues = [numSpaceHues]int32{
	SpaceNone:         0x808080,
	SpaceVoid:         0x04060c,
	SpaceStar:         0xb8c6e6,
	SpacePlanetOcean:  0x0b2648,
	SpacePlanetLand:   0x1b3226,
	SpacePlanetCloud:  0x46628a,
	SpaceAtmosphere:   0x4fb0ff,
	SpaceCityAmber:    0xffb347,
	SpaceSolar:        0x3f78d8,
	SpaceHydro:        0x7fd88e,
	SpaceRock:         0x8f8172,
	SpaceSun:          0xffd27a,
	SpaceSunCorona:    0xff9440,
	SpaceOrbitLine:    0x4c4676,
	SpaceNebulaViolet: 0x8a46d0,
	SpaceEngine:       0x78f0ff,
}

// SpaceColor returns a sky arc colour.
func SpaceColor(h SpaceHue) tcell.Color {
	if h >= numSpaceHues {
		h = SpaceNone
	}
	return tcell.NewHexColor(spaceHues[h])
}
