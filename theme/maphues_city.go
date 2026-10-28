package theme

import "github.com/gdamore/tcell/v2"

// maphues_city.go holds the Earth arc's city colours, the Modern Age to the
// Fusion Age: concrete, glass and asphalt, the last parks, landfill, the
// data age's screens, the megacity's neon and the powered city's plasma. The
// map leans theme roles toward them and clamps them legible, so like the rest
// of the map hues they are never drawn as they are. Keeping them here keeps
// "no raw colours outside theme/" true for ui/.

// CityHue names an Earth-arc city colour.
type CityHue uint8

const (
	CityConcrete    CityHue = iota // bare concrete, office blocks
	CityAsphalt                    // highways and streets
	CityGlass                      // glass towers
	CitySteel                      // steel trim, dishes
	CityRoof                       // suburban roofs
	CityPark                       // park and garden green
	CityLandfill                   // landfill and rubbish
	CityRust                       // scrap and rust
	CityScreen                     // a warm screen in a window
	CityScreenCool                 // a cold screen in a window
	CityNeonMagenta                // neon
	CityNeonCyan
	CitySteam     // steam from the vents
	CityPlasma    // white plasma
	CityTaillight // tail lights at night
	numCityHues
)

var cityHues = [numCityHues]int32{
	CityConcrete:    0x9096a0,
	CityAsphalt:     0x3c3e44,
	CityGlass:       0x6aa6dc,
	CitySteel:       0xb4c0cc,
	CityRoof:        0xc87a5a,
	CityPark:        0x4fa64f,
	CityLandfill:    0x8c7448,
	CityRust:        0xa8603a,
	CityScreen:      0xffc45a,
	CityScreenCool:  0x9ae4ff,
	CityNeonMagenta: 0xff3ea5,
	CityNeonCyan:    0x29f0ff,
	CitySteam:       0xd0d0dc,
	CityPlasma:      0xeef8ff,
	CityTaillight:   0xff4a3a,
}

// CityColor returns an Earth-arc city colour.
func CityColor(h CityHue) tcell.Color {
	if h >= numCityHues {
		h = CityConcrete
	}
	return tcell.NewHexColor(cityHues[h])
}
