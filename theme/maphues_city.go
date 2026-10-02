package theme

import "github.com/gdamore/tcell/v2"

// maphues_city.go holds the Earth arc's city colours, the Modern Age to the
// Fusion Age: concrete, glass and asphalt, the last parks, smog and
// landfill, the data age's screens and glow, the megacity's night and neon,
// its toxic water, and the powered city's plasma. Both map styles read them
// (the roguelike leans theme roles toward them and clamps them legible; the
// skyline lights and hazes them like its other scene colours), so like the
// rest of the map hues they are never drawn as they are. Keeping them here
// keeps "no raw colours outside theme/" true for ui/.

// CityHue names an Earth-arc city colour.
type CityHue uint8

const (
	CityConcrete     CityHue = iota // bare concrete, office blocks
	CityConcreteDark                // concrete in shadow, kerbs
	CityAsphalt                     // highways and streets
	CityLaneMark                    // highway lane paint
	CityGlass                       // glass towers
	CitySteel                       // steel trim, dishes
	CityRoof                        // suburban roofs
	CityPark                        // park and garden green
	CityHedge                       // park fences and hedges
	CitySmog                        // smog haze
	CityLandfill                    // landfill and rubbish
	CityRust                        // scrap and rust
	CityServer                      // server racks
	CityData                        // fibre and data light
	CityScreen                      // a warm screen in a window
	CityScreenCool                  // a cold screen in a window
	CityGlow                        // the data age's glow
	CityNight                       // the megacity's night
	CityTower                       // megacity tower roofs
	CityTowerLit                    // lit megacity windows and terraces
	CityNeonMagenta                 // neon
	CityNeonCyan
	CityNeonYellow
	CityToxic     // toxic canals and ground
	CityToxicDeep // the canals' depths
	CitySteam     // steam from the vents
	CityAcid      // acid rain
	CityPlasma    // white plasma
	CityElectric  // electric blue
	CityConduit   // power conduits
	CityReactor   // a reactor's ring
	CityTether    // the space elevator's tether
	CityHeadlight // headlights at night
	CityTaillight // tail lights at night
	CityCooling   // the powered city's cooling water
	CityReserve   // the reserve's old forest
	CityReserveWall
	numCityHues
)

var cityHues = [numCityHues]int32{
	CityConcrete:     0x9096a0,
	CityConcreteDark: 0x50545c,
	CityAsphalt:      0x3c3e44,
	CityLaneMark:     0xe8dc90,
	CityGlass:        0x6aa6dc,
	CitySteel:        0xb4c0cc,
	CityRoof:         0xc87a5a,
	CityPark:         0x4fa64f,
	CityHedge:        0x6a9a58,
	CitySmog:         0x8a7c66,
	CityLandfill:     0x8c7448,
	CityRust:         0xa8603a,
	CityServer:       0x3a5aa0,
	CityData:         0x40d8f0,
	CityScreen:       0xffc45a,
	CityScreenCool:   0x9ae4ff,
	CityGlow:         0x4ad4ff,
	CityNight:        0x0e0a18,
	CityTower:        0x2c2442,
	CityTowerLit:     0x6a5a9a,
	CityNeonMagenta:  0xff3ea5,
	CityNeonCyan:     0x29f0ff,
	CityNeonYellow:   0xffe23a,
	CityToxic:        0xd2c83c,
	CityToxicDeep:    0x2c2a16,
	CitySteam:        0xd0d0dc,
	CityAcid:         0xa8c070,
	CityPlasma:       0xeef8ff,
	CityElectric:     0x3a8cff,
	CityConduit:      0x5ab4ff,
	CityReactor:      0xa8ecff,
	CityTether:       0xdcf2ff,
	CityHeadlight:    0xfff2c4,
	CityTaillight:    0xff4a3a,
	CityCooling:      0x1e4a8a,
	CityReserve:      0x3a8a42,
	CityReserveWall:  0x8a9a88,
}

// CityColor returns an Earth-arc city colour.
func CityColor(h CityHue) tcell.Color {
	if h >= numCityHues {
		h = CityConcrete
	}
	return tcell.NewHexColor(cityHues[h])
}

// NumCityHues is the size of the city colour table (for caches and tests).
const NumCityHues = int(numCityHues)
