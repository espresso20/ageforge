package ui

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// TerrainPalette defines the color palette for a terrain era
type TerrainPalette struct {
	Ground, GroundAlt                      color.RGBA
	Water, WaterLight, WaterDeep           color.RGBA
	Tree, TreeDark, TreeLight              color.RGBA
	Road, RoadEdge                         color.RGBA
	Hill, HillLight                        color.RGBA
	Farmland, FarmAlt                      color.RGBA
	Accent1, Accent2                       color.RGBA // era-specific accents
}

// MapGenConfig holds parameters for map generation
type MapGenConfig struct {
	Width, Height int
	DetailLevel   int // 0=mini, 1=full
	Buildings     map[string]game.BuildingState
	AgeKey        string
}

func eraFromAge(ageKey string) int {
	ages := config.AgeByKey()
	if a, ok := ages[ageKey]; ok {
		o := a.Order
		switch {
		case o <= 1:
			return 0
		case o <= 4:
			return 1
		case o <= 7:
			return 2
		case o <= 10:
			return 3
		case o <= 13:
			return 4
		case o <= 15:
			return 5
		case o == 16:
			return 6
		case o <= 19:
			return 7
		default:
			return 8
		}
	}
	return 0
}

func c(r, g, b uint8) color.RGBA { return color.RGBA{r, g, b, 255} }

func getTerrainPalette(era int) TerrainPalette {
	switch era {
	case 0: // primitive — lush green wilderness
		return TerrainPalette{
			Ground: c(34, 85, 34), GroundAlt: c(44, 100, 42),
			Water: c(28, 55, 150), WaterLight: c(50, 80, 175), WaterDeep: c(18, 40, 120),
			Tree: c(20, 110, 30), TreeDark: c(12, 75, 18), TreeLight: c(35, 130, 45),
			Road: c(85, 65, 42), RoadEdge: c(65, 50, 32),
			Hill: c(50, 105, 50), HillLight: c(65, 120, 62),
			Farmland: c(90, 120, 40), FarmAlt: c(80, 110, 35),
			Accent1: c(180, 120, 40), Accent2: c(200, 100, 30), // campfire orange
		}
	case 1: // ancient — sandy fields and farms
		return TerrainPalette{
			Ground: c(110, 130, 55), GroundAlt: c(120, 140, 62),
			Water: c(30, 68, 140), WaterLight: c(45, 85, 162), WaterDeep: c(20, 50, 110),
			Tree: c(50, 100, 30), TreeDark: c(38, 78, 22), TreeLight: c(65, 118, 42),
			Road: c(145, 125, 82), RoadEdge: c(115, 98, 65),
			Hill: c(140, 120, 80), HillLight: c(160, 140, 95),
			Farmland: c(150, 170, 60), FarmAlt: c(140, 160, 50),
			Accent1: c(180, 160, 100), Accent2: c(160, 140, 80), // sandstone
		}
	case 2: // medieval — dark forest and stone
		return TerrainPalette{
			Ground: c(30, 60, 25), GroundAlt: c(38, 72, 32),
			Water: c(20, 48, 128), WaterLight: c(32, 65, 148), WaterDeep: c(12, 35, 100),
			Tree: c(15, 80, 20), TreeDark: c(8, 52, 12), TreeLight: c(25, 98, 30),
			Road: c(105, 95, 72), RoadEdge: c(80, 72, 55),
			Hill: c(45, 75, 40), HillLight: c(58, 90, 52),
			Farmland: c(70, 95, 30), FarmAlt: c(80, 105, 35),
			Accent1: c(140, 130, 110), Accent2: c(120, 110, 95), // stone walls
		}
	case 3: // industrial — smoke and brick
		return TerrainPalette{
			Ground: c(68, 68, 62), GroundAlt: c(78, 78, 72),
			Water: c(38, 58, 88), WaterLight: c(50, 70, 102), WaterDeep: c(28, 42, 68),
			Tree: c(48, 78, 38), TreeDark: c(35, 58, 28), TreeLight: c(58, 90, 48),
			Road: c(80, 75, 70), RoadEdge: c(60, 58, 55),
			Hill: c(85, 85, 78), HillLight: c(98, 98, 90),
			Farmland: c(75, 85, 55), FarmAlt: c(82, 92, 60),
			Accent1: c(140, 60, 40), Accent2: c(90, 90, 90), // brick + steel
		}
	case 4: // modern — asphalt and glass
		return TerrainPalette{
			Ground: c(52, 58, 62), GroundAlt: c(62, 68, 72),
			Water: c(28, 78, 148), WaterLight: c(42, 95, 168), WaterDeep: c(18, 58, 118),
			Tree: c(38, 88, 38), TreeDark: c(28, 68, 28), TreeLight: c(50, 105, 50),
			Road: c(50, 50, 52), RoadEdge: c(70, 70, 72),
			Hill: c(72, 78, 82), HillLight: c(85, 90, 95),
			Farmland: c(60, 80, 45), FarmAlt: c(68, 88, 52),
			Accent1: c(140, 180, 220), Accent2: c(200, 200, 210), // glass + white
		}
	case 5: // digital — server glow
		return TerrainPalette{
			Ground: c(10, 15, 38), GroundAlt: c(15, 22, 48),
			Water: c(18, 38, 118), WaterLight: c(28, 52, 138), WaterDeep: c(10, 25, 88),
			Tree: c(8, 28, 58), TreeDark: c(5, 18, 42), TreeLight: c(12, 38, 72),
			Road: c(0, 80, 130), RoadEdge: c(0, 60, 100),
			Hill: c(18, 25, 55), HillLight: c(25, 35, 68),
			Farmland: c(15, 35, 60), FarmAlt: c(18, 42, 72),
			Accent1: c(0, 200, 255), Accent2: c(0, 150, 200), // cyan glow
		}
	case 6: // cyberpunk — neon on black
		return TerrainPalette{
			Ground: c(8, 8, 12), GroundAlt: c(14, 12, 20),
			Water: c(78, 0, 118), WaterLight: c(100, 0, 148), WaterDeep: c(55, 0, 85),
			Tree: c(0, 38, 18), TreeDark: c(0, 22, 10), TreeLight: c(0, 55, 28),
			Road: c(30, 25, 35), RoadEdge: c(20, 18, 25),
			Hill: c(15, 12, 22), HillLight: c(22, 18, 32),
			Farmland: c(0, 50, 25), FarmAlt: c(0, 60, 30),
			Accent1: c(255, 0, 128), Accent2: c(0, 255, 80), // neon pink + green
		}
	case 7: // space — starfield and domes
		return TerrainPalette{
			Ground: c(5, 5, 15), GroundAlt: c(8, 8, 22),
			Water: c(15, 25, 78), WaterLight: c(22, 35, 98), WaterDeep: c(8, 15, 55),
			Tree: c(8, 8, 28), TreeDark: c(5, 5, 18), TreeLight: c(12, 12, 38),
			Road: c(58, 78, 138), RoadEdge: c(42, 58, 105),
			Hill: c(10, 10, 28), HillLight: c(15, 15, 38),
			Farmland: c(10, 18, 35), FarmAlt: c(12, 22, 42),
			Accent1: c(100, 160, 255), Accent2: c(200, 220, 255), // blue-white
		}
	case 8: // cosmic — transcendent void
		return TerrainPalette{
			Ground: c(3, 3, 8), GroundAlt: c(8, 5, 15),
			Water: c(38, 10, 78), WaterLight: c(58, 15, 98), WaterDeep: c(25, 5, 55),
			Tree: c(5, 5, 15), TreeDark: c(3, 3, 10), TreeLight: c(8, 8, 22),
			Road: c(78, 58, 18), RoadEdge: c(58, 42, 12),
			Hill: c(8, 5, 18), HillLight: c(12, 8, 25),
			Farmland: c(10, 8, 20), FarmAlt: c(12, 10, 25),
			Accent1: c(255, 200, 50), Accent2: c(200, 150, 255), // gold + violet
		}
	}
	return getTerrainPalette(0)
}

func hashKey(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}

func lerp(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 255,
	}
}

func noise2D(x, y int, seed uint64) float64 {
	h1 := hashKey(string(rune(seed)) + string(rune(x*7919+y*6271)))
	h2 := hashKey(string(rune(seed+77)) + string(rune((x/3)*4909+(y/3)*3571)))
	return float64(h1%10000)/10000.0*0.6 + float64(h2%10000)/10000.0*0.4
}

type bldInfo struct {
	key      string
	category string
	x, y     int
	size     int
}

// BuildingShape defines the visual footprint template for a building
type BuildingShape int

const (
	ShapeCircle   BuildingShape = iota // filled circle — huts, domes, reactors
	ShapeSquare                        // basic rectangle — houses, warehouses
	ShapeTriangle                      // peaked/tent — camps, manors, cathedrals
	ShapeDiamond                       // rotated square — altars, markets, gems
	ShapeCross                         // + shape — churches, labs, hospitals
	ShapeTower                         // tall narrow — skyscrapers, silos
	ShapeWide                          // wide & low — farms, stations
	ShapeLShape                        // L-footprint — factories, mills, barracks
	ShapeRing                          // hollow circle — arenas, accelerators
	ShapeStar                          // 4-point star — wonders, beacons
	ShapeHexagon                       // hexagonal — high-tech, hubs
	ShapeDish                          // half-dome/dish — observatories, antennas
)

// BuildingVisual maps a building to its unique shape and colors
type BuildingVisual struct {
	Shape   BuildingShape
	Primary color.RGBA
	Accent  color.RGBA
}

var buildingVisuals = map[string]BuildingVisual{
	// ── Housing (greens/browns) ──
	"hut":              {ShapeCircle, c(160, 120, 70), c(190, 170, 60)},
	"house":            {ShapeSquare, c(140, 90, 50), c(170, 80, 50)},
	"manor":            {ShapeTriangle, c(140, 140, 130), c(130, 40, 35)},
	"apartment":        {ShapeTower, c(180, 165, 140), c(120, 85, 55)},
	"skyscraper":       {ShapeTower, c(130, 160, 190), c(220, 220, 225)},
	"neon_tower":       {ShapeTower, c(60, 20, 80), c(255, 50, 150)},
	"orbital_habitat":  {ShapeCircle, c(180, 185, 190), c(140, 180, 220)},

	// ── Production (warm yellows/oranges/distinct) ──
	"gathering_camp":   {ShapeTriangle, c(100, 110, 50), c(110, 80, 40)},
	"woodcutter_camp":  {ShapeTriangle, c(120, 80, 40), c(40, 80, 30)},
	"stone_pit":        {ShapeWide, c(130, 130, 125), c(90, 90, 85)},
	"farm":             {ShapeWide, c(140, 170, 50), c(190, 170, 40)},
	"lumber_mill":      {ShapeLShape, c(120, 80, 40), c(80, 55, 30)},
	"quarry":           {ShapeWide, c(120, 115, 100), c(100, 80, 55)},
	"mine":             {ShapeDiamond, c(80, 80, 80), c(190, 150, 40)},
	"market":           {ShapeDiamond, c(200, 170, 40), c(180, 50, 40)},
	"coal_mine":        {ShapeDiamond, c(55, 50, 45), c(200, 120, 30)},
	"smithy":           {ShapeLShape, c(140, 40, 30), c(210, 140, 30)},
	"forum":            {ShapeSquare, c(220, 210, 180), c(200, 170, 40)},
	"aqueduct":         {ShapeWide, c(160, 155, 140), c(60, 100, 180)},
	"amphitheater":     {ShapeRing, c(190, 170, 130), c(170, 90, 55)},
	"cathedral":        {ShapeCross, c(160, 155, 140), c(200, 170, 40)},
	"art_studio":       {ShapeSquare, c(170, 150, 190), c(190, 120, 140)},
	"bank":             {ShapeSquare, c(200, 170, 40), c(40, 80, 50)},
	"colony":           {ShapeSquare, c(170, 150, 110), c(80, 130, 60)},
	"port":             {ShapeWide, c(110, 120, 135), c(100, 70, 40)},
	"plantation":       {ShapeWide, c(40, 100, 40), c(200, 180, 40)},
	"factory":          {ShapeLShape, c(150, 60, 40), c(110, 110, 110)},
	"oil_well":         {ShapeTower, c(30, 30, 30), c(210, 130, 20)},
	"power_grid":       {ShapeCross, c(200, 190, 40), c(130, 130, 130)},
	"telegraph":        {ShapeDish, c(110, 80, 40), c(170, 110, 50)},
	"clocktower":       {ShapeTower, c(150, 145, 130), c(200, 170, 40)},
	"electric_mill":    {ShapeLShape, c(140, 150, 160), c(170, 110, 50)},
	"train_station":    {ShapeWide, c(120, 80, 40), c(170, 40, 30)},
	"reactor":          {ShapeCircle, c(140, 140, 140), c(60, 180, 60)},
	"power_plant":      {ShapeLShape, c(160, 160, 155), c(200, 190, 40)},
	"server_farm":      {ShapeSquare, c(30, 50, 100), c(40, 160, 60)},
	"fiber_hub":        {ShapeHexagon, c(40, 140, 130), c(220, 220, 225)},
	"media_center":     {ShapeSquare, c(220, 220, 225), c(60, 100, 200)},
	"data_center":      {ShapeSquare, c(60, 70, 85), c(40, 190, 200)},
	"smart_grid":       {ShapeCross, c(170, 175, 180), c(40, 190, 200)},
	"augmentation_clinic": {ShapeCross, c(220, 220, 225), c(40, 220, 80)},
	"black_market":     {ShapeDiamond, c(70, 70, 70), c(200, 40, 160)},
	"fusion_reactor":   {ShapeCircle, c(60, 100, 200), c(240, 230, 200)},
	"plasma_forge":     {ShapeLShape, c(220, 140, 30), c(240, 230, 200)},
	"maglev_station":   {ShapeWide, c(170, 175, 180), c(60, 100, 200)},
	"launch_pad":       {ShapeWide, c(160, 160, 155), c(220, 130, 30)},
	"warp_gate":        {ShapeRing, c(60, 40, 130), c(230, 230, 240)},
	"colony_ship":      {ShapeTriangle, c(170, 175, 180), c(60, 100, 200)},
	"star_forge":       {ShapeHexagon, c(220, 140, 30), c(240, 230, 200)},
	"galactic_hub":     {ShapeHexagon, c(200, 170, 40), c(60, 100, 200)},
	"antimatter_plant": {ShapeCircle, c(60, 20, 80), c(220, 100, 160)},
	"megastructure":    {ShapeStar, c(170, 175, 180), c(200, 170, 40)},
	"reality_engine":   {ShapeHexagon, c(120, 50, 160), c(230, 230, 240)},
	"transcendence_beacon": {ShapeTower, c(210, 190, 50), c(240, 240, 240)},

	// ── Research (blues/cyans) ──
	"altar":            {ShapeDiamond, c(90, 85, 80), c(50, 70, 170)},
	"firepit":          {ShapeCircle, c(210, 130, 30), c(180, 40, 20)},
	"library":          {ShapeSquare, c(110, 70, 35), c(50, 70, 170)},
	"university":       {ShapeTriangle, c(150, 145, 130), c(40, 60, 160)},
	"observatory":      {ShapeDish, c(30, 40, 100), c(170, 175, 180)},
	"telephone_exchange": {ShapeSquare, c(110, 80, 40), c(170, 110, 50)},
	"research_lab":     {ShapeCross, c(220, 220, 225), c(40, 80, 200)},
	"space_station":    {ShapeRing, c(170, 175, 180), c(60, 100, 200)},
	"ai_lab":           {ShapeHexagon, c(25, 40, 100), c(40, 190, 200)},
	"quantum_computer": {ShapeHexagon, c(60, 30, 120), c(140, 60, 200)},

	// ── Military (reds/dark) ──
	"barracks":         {ShapeLShape, c(120, 35, 30), c(100, 100, 100)},
	"castle":           {ShapeSquare, c(140, 135, 120), c(70, 65, 60)},
	"bunker":           {ShapeSquare, c(140, 140, 135), c(50, 70, 45)},
	"missile_silo":     {ShapeTower, c(120, 120, 120), c(180, 40, 30)},

	// ── Storage (purples/grays) ──
	"stash":            {ShapeCircle, c(160, 130, 90), c(120, 85, 55)},
	"storage_pit":      {ShapeWide, c(120, 95, 60), c(100, 75, 45)},
	"warehouse":        {ShapeSquare, c(120, 80, 40), c(110, 110, 110)},
	"granary":          {ShapeCircle, c(200, 170, 50), c(120, 85, 55)},
	"classical_vault":  {ShapeSquare, c(200, 195, 185), c(200, 170, 40)},
	"keep":             {ShapeSquare, c(130, 125, 115), c(60, 55, 50)},
	"renaissance_vault": {ShapeSquare, c(220, 210, 185), c(200, 170, 40)},
	"colonial_warehouse": {ShapeSquare, c(120, 80, 40), c(160, 40, 30)},
	"industrial_depot": {ShapeLShape, c(120, 120, 120), c(100, 75, 45)},
	"victorian_vault":  {ShapeSquare, c(120, 35, 30), c(200, 170, 40)},
	"electric_warehouse": {ShapeSquare, c(140, 150, 160), c(200, 190, 40)},
	"atomic_vault":     {ShapeSquare, c(160, 160, 155), c(60, 150, 60)},
	"modern_depot":     {ShapeSquare, c(120, 120, 125), c(60, 100, 180)},
	"info_vault":       {ShapeSquare, c(30, 40, 90), c(100, 50, 140)},
	"digital_archive":  {ShapeHexagon, c(60, 70, 85), c(40, 190, 200)},
	"cyber_vault":      {ShapeDiamond, c(50, 50, 50), c(200, 40, 160)},
	"fusion_vault":     {ShapeCircle, c(60, 100, 200), c(230, 230, 240)},
	"orbital_depot":    {ShapeSquare, c(170, 175, 180), c(60, 100, 200)},
	"stellar_vault":    {ShapeDiamond, c(200, 170, 40), c(60, 100, 200)},
	"galactic_vault":   {ShapeHexagon, c(100, 50, 140), c(200, 170, 40)},
	"quantum_vault":    {ShapeDiamond, c(60, 30, 120), c(140, 60, 200)},

	// ── Wonders (vivid, unique) ──
	"stonehenge":          {ShapeRing, c(140, 140, 135), c(60, 80, 180)},
	"colosseum":           {ShapeRing, c(190, 170, 130), c(170, 50, 40)},
	"parthenon":           {ShapeTriangle, c(210, 205, 195), c(200, 170, 40)},
	"great_library":       {ShapeSquare, c(120, 80, 40), c(200, 170, 40)},
	"space_program":       {ShapeStar, c(230, 230, 235), c(220, 140, 30)},
	"particle_accelerator": {ShapeRing, c(140, 150, 160), c(40, 170, 60)},
	"dyson_scaffold":      {ShapeStar, c(210, 180, 40), c(220, 140, 30)},
	"singularity_core":    {ShapeStar, c(20, 20, 25), c(240, 240, 245)},
}

func getBuildingVisual(key, category string) BuildingVisual {
	if vis, ok := buildingVisuals[key]; ok {
		return vis
	}
	// Fallback by category
	switch category {
	case "housing":
		return BuildingVisual{ShapeSquare, c(140, 120, 90), c(120, 80, 50)}
	case "production":
		return BuildingVisual{ShapeLShape, c(160, 130, 60), c(120, 100, 50)}
	case "research":
		return BuildingVisual{ShapeDiamond, c(80, 100, 160), c(60, 80, 140)}
	case "military":
		return BuildingVisual{ShapeLShape, c(130, 40, 35), c(90, 90, 90)}
	case "storage":
		return BuildingVisual{ShapeSquare, c(130, 120, 110), c(100, 90, 80)}
	case "wonder":
		return BuildingVisual{ShapeStar, c(200, 180, 60), c(220, 200, 80)}
	default:
		return BuildingVisual{ShapeSquare, c(140, 130, 120), c(110, 100, 90)}
	}
}

// GenerateMapImage creates a procedural pixel map as an image.RGBA
func GenerateMapImage(cfg MapGenConfig) *image.RGBA {
	w, h := cfg.Width, cfg.Height
	if w < 4 || h < 4 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	era := eraFromAge(cfg.AgeKey)
	pal := getTerrainPalette(era)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	seed := hashKey(cfg.AgeKey)
	cx, cy := w/2, h/2
	dl := cfg.DetailLevel

	// ═══════════════════════════════════════════
	// 1. BASE TERRAIN
	// ═══════════════════════════════════════════
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := noise2D(x, y, seed)
			elev := noise2D(x/4, y/4, seed+200)
			if elev > 0.65 {
				hillT := (elev - 0.65) / 0.35
				base := lerp(pal.Ground, pal.GroundAlt, n)
				img.SetRGBA(x, y, lerp(base, lerp(pal.Hill, pal.HillLight, n), hillT))
			} else {
				img.SetRGBA(x, y, lerp(pal.Ground, pal.GroundAlt, n))
			}
		}
	}

	// ═══════════════════════════════════════════
	// 2. ERA-SPECIFIC BACKGROUND FEATURES
	// ═══════════════════════════════════════════
	switch {
	case era >= 7: // space/cosmic — stars and nebulae
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n := noise2D(x, y, seed+99)
				if n < 0.03 {
					br := uint8(90 + int(n*5500))
					if era == 8 {
						img.SetRGBA(x, y, c(br, uint8(float64(br)*0.8), uint8(float64(br)*0.3)))
					} else {
						img.SetRGBA(x, y, c(br, br, uint8(min(255, int(float64(br)*1.2)))))
					}
				}
				// Nebula clouds
				if era == 8 {
					neb := noise2D(x/6, y/6, seed+500)
					if neb > 0.7 {
						t := (neb - 0.7) / 0.3 * 0.15
						existing := img.RGBAAt(x, y)
						img.SetRGBA(x, y, lerp(existing, pal.Accent2, t))
					}
				}
			}
		}
	case era == 6: // cyberpunk — grid lines on ground
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x%(w/12) == 0 || y%(h/12) == 0 {
					existing := img.RGBAAt(x, y)
					img.SetRGBA(x, y, lerp(existing, pal.Accent1, 0.08))
				}
			}
		}
	case era == 5: // digital — faint circuit traces
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x%(w/8) == 0 || y%(h/8) == 0 {
					existing := img.RGBAAt(x, y)
					img.SetRGBA(x, y, lerp(existing, pal.Accent1, 0.06))
				}
			}
		}
	}

	// ═══════════════════════════════════════════
	// 3. RIVER
	// ═══════════════════════════════════════════
	riverBaseX := float64(w) * 0.28
	riverW := 3 + dl*3
	bankW := 1 + dl
	// No river in space/cosmic
	if era < 7 {
		for y := 0; y < h; y++ {
			rx := riverBaseX + math.Sin(float64(y)*0.06)*float64(w)*0.10 + math.Sin(float64(y)*0.15)*float64(w)*0.03
			for dx := -bankW; dx < riverW+bankW; dx++ {
				px := int(rx) + dx
				if px < 0 || px >= w {
					continue
				}
				if dx < 0 || dx >= riverW {
					existing := img.RGBAAt(px, y)
					img.SetRGBA(px, y, lerp(existing, pal.Water, 0.3))
				} else {
					centerDist := math.Abs(float64(dx)-float64(riverW)/2.0) / (float64(riverW) / 2.0)
					wc := lerp(pal.WaterDeep, pal.WaterLight, centerDist)
					wc = lerp(wc, pal.Water, noise2D(px, y, seed+7)*0.2)
					img.SetRGBA(px, y, wc)
				}
			}
		}
	}

	// ═══════════════════════════════════════════
	// 4. VEGETATION (era-specific)
	// ═══════════════════════════════════════════
	drawVegetation(img, w, h, era, dl, pal, seed)

	// ═══════════════════════════════════════════
	// 5. COLLECT BUILDING PLACEMENTS
	// ═══════════════════════════════════════════
	placements := placeBuildingsRadial(cfg.Buildings, cx, cy, w, h, era, dl)

	// Sort: furthest first so close buildings draw on top
	sort.Slice(placements, func(i, j int) bool {
		di := (placements[i].x-cx)*(placements[i].x-cx) + (placements[i].y-cy)*(placements[i].y-cy)
		dj := (placements[j].x-cx)*(placements[j].x-cx) + (placements[j].y-cy)*(placements[j].y-cy)
		return di > dj
	})

	// ═══════════════════════════════════════════
	// 6. SURROUNDINGS (farmland, parking lots, etc)
	// ═══════════════════════════════════════════
	drawSurroundings(img, w, h, era, dl, pal, seed, placements)

	// ═══════════════════════════════════════════
	// 7. INFRASTRUCTURE (paths → roads → rails → highways → glowing lines)
	// ═══════════════════════════════════════════
	drawInfrastructure(img, w, h, era, dl, pal, cx, cy, placements)

	// ═══════════════════════════════════════════
	// 8. BUILDINGS (era-specific rendering)
	// ═══════════════════════════════════════════
	drawBuildings(img, w, h, era, dl, pal, placements)

	// ═══════════════════════════════════════════
	// 9. ERA DECORATIONS (smokestacks, power lines, neon, etc)
	// ═══════════════════════════════════════════
	drawDecorations(img, w, h, era, dl, pal, seed, placements)

	return img
}

// ─── Vegetation ──────────────────────────────────────────
func drawVegetation(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64) {
	// Primitive/ancient/medieval: dense forests
	// Industrial: sparse remaining trees
	// Modern+: decorative parks only
	// Digital/cyber: none
	// Space/cosmic: none
	if era >= 5 {
		return
	}

	density := 0.10
	switch era {
	case 0:
		density = 0.14
	case 1:
		density = 0.08
	case 2:
		density = 0.10
	case 3:
		density = 0.025
	case 4:
		density = 0.015
	}
	treeR := 2 + dl
	step := treeR + 1

	for y := treeR; y < h-treeR; y += step {
		for x := treeR; x < w-treeR; x += step {
			n := noise2D(x, y, seed+42)
			if n >= density {
				continue
			}
			// Check not in water
			px := img.RGBAAt(x, y)
			if px == pal.Water || px == pal.WaterLight || px == pal.WaterDeep {
				continue
			}
			// Draw canopy
			for dy := -treeR; dy <= treeR; dy++ {
				for dx := -treeR; dx <= treeR; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(treeR) {
						continue
					}
					tx, ty := x+dx, y+dy
					if tx < 0 || tx >= w || ty < 0 || ty >= h {
						continue
					}
					shade := float64(dy+treeR) / float64(treeR*2)
					tc := lerp(pal.TreeLight, pal.TreeDark, shade)
					edgeFade := d / float64(treeR)
					existing := img.RGBAAt(tx, ty)
					img.SetRGBA(tx, ty, lerp(existing, tc, 1.0-edgeFade*0.4))
				}
			}
			img.SetRGBA(x, y, pal.TreeDark)
		}
	}
}

// ─── Building placement ──────────────────────────────────
func placeBuildingsRadial(buildings map[string]game.BuildingState, cx, cy, w, h, era, dl int) []bldInfo {
	var placements []bldInfo
	maxDist := float64(min(w, h)) / 2.0

	for key, bs := range buildings {
		if !bs.Unlocked || bs.Count == 0 {
			continue
		}
		ringMin, ringMax := 0.08, 0.35
		switch bs.Category {
		case "housing":
			ringMin, ringMax = 0.05, 0.22
		case "production":
			ringMin, ringMax = 0.12, 0.32
		case "military":
			ringMin, ringMax = 0.22, 0.42
		case "wonder":
			ringMin, ringMax = 0.03, 0.16
		case "research":
			ringMin, ringMax = 0.08, 0.26
		case "storage":
			ringMin, ringMax = 0.06, 0.20
		}

		for i := 0; i < bs.Count; i++ {
			bHash := hashKey(key + string(rune(i)))
			angle := float64(bHash%3600) / 3600.0 * 2.0 * math.Pi
			distRatio := ringMin + float64(bHash%1000)/1000.0*(ringMax-ringMin)
			dist := distRatio * maxDist
			bx := cx + int(math.Cos(angle)*dist)
			by := cy + int(math.Sin(angle)*dist*0.7)

			// Building size scales with era
			size := 3
			switch {
			case era >= 6: // cyber+ = mega towers
				size = 5
			case era >= 4: // modern = skyscrapers
				size = 4
			case era >= 3: // industrial = bigger
				size = 4
			}
			if bs.Category == "wonder" {
				size += 3
			} else if bs.Category == "military" {
				size += 1
			}
			if dl > 0 {
				size += 2
			}

			placements = append(placements, bldInfo{key, bs.Category, bx, by, size})
		}
	}
	return placements
}

// ─── Surroundings ────────────────────────────────────────
func drawSurroundings(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64, placements []bldInfo) {
	for _, b := range placements {
		radius := b.size + 4 + dl*2

		switch {
		case era <= 2 && b.category == "production":
			// Farmland with crop rows
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fc := pal.Farmland
					if (px+py)%4 < 2 {
						fc = pal.FarmAlt
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(fc, existing, fade*0.5))
				}
			}

		case era >= 4 && era <= 5 && b.category == "housing":
			// Parking lot / pavement around modern buildings
			asphalt := c(45, 45, 48)
			lineC := c(220, 220, 50)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					pc := asphalt
					if dx%6 == 0 && dy > 0 {
						pc = lineC // parking lines
					}
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(pc, existing, fade*0.6))
				}
			}

		case era == 3 && b.category == "production":
			// Soot-stained ground around factories
			soot := c(40, 38, 35)
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d > float64(radius) || d < float64(b.size) {
						continue
					}
					px, py := b.x+dx, b.y+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := (d - float64(b.size)) / float64(radius-b.size)
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(soot, existing, fade*0.4+noise2D(px, py, seed+300)*0.3))
				}
			}
		}
	}
}

// ─── Infrastructure ──────────────────────────────────────
func drawInfrastructure(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, cx, cy int, placements []bldInfo) {
	if len(placements) == 0 {
		return
	}

	switch {
	case era == 0:
		// Dirt footpaths — thin, irregular
		for _, b := range placements {
			drawPath(img, cx, cy, b.x, b.y, w, h, pal.Road, 0)
		}

	case era <= 2:
		// Stone/cobble roads
		for _, b := range placements {
			drawRoad(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
		}

	case era == 3:
		// Railways: dark track + sleeper ties
		trackC := c(60, 55, 50)
		tieC := c(90, 70, 45)
		for _, b := range placements {
			drawRailway(img, cx, cy, b.x, b.y, w, h, trackC, tieC, dl)
		}
		// Also some roads
		for i, b := range placements {
			if i%3 == 0 {
				drawRoad(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
			}
		}

	case era <= 5:
		// Multi-lane highways with lane markings
		for _, b := range placements {
			drawHighway(img, cx, cy, b.x, b.y, w, h, pal.Road, pal.RoadEdge, dl)
		}

	case era == 6:
		// Neon light trails
		for _, b := range placements {
			drawNeonTrail(img, cx, cy, b.x, b.y, w, h, pal.Accent1, pal.Accent2, dl)
		}

	case era >= 7:
		// Energy conduits — glowing blue/gold lines
		for _, b := range placements {
			drawNeonTrail(img, cx, cy, b.x, b.y, w, h, pal.Accent1, pal.Accent2, dl)
		}
	}
}

// ─── Building rendering ──────────────────────────────────
func drawBuildings(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, placements []bldInfo) {
	for _, b := range placements {
		bx, by, size := b.x, b.y, b.size

		// Shadow
		shadowOff := 1 + dl
		for dy := 0; dy < size; dy++ {
			for dx := 0; dx < size; dx++ {
				px := bx - size/2 + dx + shadowOff
				py := by - size/2 + dy + shadowOff
				if px >= 0 && px < w && py >= 0 && py < h {
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, c(0, 0, 0), 0.35))
				}
			}
		}

		vis := getBuildingVisual(b.key, b.category)
		drawBuildingShape(img, w, h, b, vis, era, dl)

		// Wonder glow (all eras)
		if b.category == "wonder" {
			glowR := size + 3 + dl*3
			for dy := -glowR; dy <= glowR; dy++ {
				for dx := -glowR; dx <= glowR; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d <= float64(size/2) || d >= float64(glowR) {
						continue
					}
					px, py := bx+dx, by+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					fade := 1.0 - (d-float64(size/2))/float64(glowR-size/2)
					gc := pal.Accent1
					if era >= 6 {
						gc = pal.Accent2
					}
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, gc, fade*0.25))
				}
			}
		}
	}
}

// ─── Shape templates ─────────────────────────────────────
// Each returns true if (dx,dy) relative to center is inside the shape,
// and whether it's an "accent" region (roof/trim/detail).

func shapeCircle(dx, dy, r int) (inside, accent bool) {
	d := math.Sqrt(float64(dx*dx + dy*dy))
	if d > float64(r) {
		return false, false
	}
	return true, dy < 0 // top half = accent
}

func shapeSquare(dx, dy, r int) (inside, accent bool) {
	if dx < -r || dx > r || dy < -r || dy > r {
		return false, false
	}
	return true, dy < -r/3 // top third = accent (roof)
}

func shapeTriangle(dx, dy, r int) (inside, accent bool) {
	if dx < -r || dx > r || dy < -r || dy > r {
		return false, false
	}
	// Bottom 60% is rectangular body
	if dy > -r/3 {
		return true, false
	}
	// Top 40% is peaked triangle
	mid := 0
	peakH := r + r/3 // height of peak region
	fromTop := dy + r // distance from top
	span := r * fromTop / peakH
	if dx >= mid-span && dx <= mid+span {
		return true, true
	}
	return false, false
}

func shapeDiamond(dx, dy, r int) (inside, accent bool) {
	if abs(dx)+abs(dy) > r {
		return false, false
	}
	return true, dy < 0
}

func shapeCross(dx, dy, r int) (inside, accent bool) {
	armW := max(1, r/3)
	inVert := dx >= -armW && dx <= armW && dy >= -r && dy <= r
	inHoriz := dy >= -armW && dy <= armW && dx >= -r && dx <= r
	if !inVert && !inHoriz {
		return false, false
	}
	// Center overlap = accent
	return true, inVert && inHoriz
}

func shapeTower(dx, dy, r int) (inside, accent bool) {
	halfW := max(1, r/2)
	if dx < -halfW || dx > halfW || dy < -r || dy > r {
		return false, false
	}
	return true, dy < -r/2 // top portion = accent
}

func shapeWide(dx, dy, r int) (inside, accent bool) {
	halfH := max(1, r/2)
	if dx < -r || dx > r || dy < -halfH || dy > halfH {
		return false, false
	}
	return true, dy < 0 // top stripe = accent
}

func shapeLShape(dx, dy, r int) (inside, accent bool) {
	// Vertical part (left half, full height)
	leftW := max(1, r/2)
	inLeft := dx >= -r && dx <= -r+leftW && dy >= -r && dy <= r
	// Horizontal part (bottom half, full width)
	botH := max(1, r/2)
	inBot := dx >= -r && dx <= r && dy >= r-botH && dy <= r
	if !inLeft && !inBot {
		return false, false
	}
	return true, inLeft && dy < 0 // upper-left part = accent
}

func shapeRing(dx, dy, r int) (inside, accent bool) {
	d := math.Sqrt(float64(dx*dx + dy*dy))
	inner := float64(r) * 0.5
	if d > float64(r) || d < inner {
		return false, false
	}
	return true, d > float64(r)*0.8 // outer rim = accent
}

func shapeStar(dx, dy, r int) (inside, accent bool) {
	// 4-pointed star = diamond + cross overlap
	armW := max(1, r/3)
	inCross := (dx >= -armW && dx <= armW && dy >= -r && dy <= r) ||
		(dy >= -armW && dy <= armW && dx >= -r && dx <= r)
	inDiamond := abs(dx)+abs(dy) <= r
	if !inCross && !inDiamond {
		return false, false
	}
	// Points/tips = accent
	return true, !inDiamond && inCross
}

func shapeHexagon(dx, dy, r int) (inside, accent bool) {
	// Hex: |dy| <= r, |dx| <= r - |dy|/2
	adx := abs(dx)
	ady := abs(dy)
	if ady > r {
		return false, false
	}
	maxX := r - ady/2
	if adx > maxX {
		return false, false
	}
	return true, ady < r/3 // middle band = accent
}

func shapeDish(dx, dy, r int) (inside, accent bool) {
	// Top half: half-circle dome
	if dy < 0 {
		d := math.Sqrt(float64(dx*dx + dy*dy))
		if d <= float64(r) {
			return true, true // dome = accent
		}
		return false, false
	}
	// Bottom: small base rect
	baseW := max(1, r*2/3)
	baseH := max(1, r/2)
	if dx >= -baseW && dx <= baseW && dy <= baseH {
		return true, false
	}
	return false, false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// testShape dispatches to the correct shape function
func testShape(shape BuildingShape, dx, dy, r int) (inside, accent bool) {
	switch shape {
	case ShapeCircle:
		return shapeCircle(dx, dy, r)
	case ShapeSquare:
		return shapeSquare(dx, dy, r)
	case ShapeTriangle:
		return shapeTriangle(dx, dy, r)
	case ShapeDiamond:
		return shapeDiamond(dx, dy, r)
	case ShapeCross:
		return shapeCross(dx, dy, r)
	case ShapeTower:
		return shapeTower(dx, dy, r)
	case ShapeWide:
		return shapeWide(dx, dy, r)
	case ShapeLShape:
		return shapeLShape(dx, dy, r)
	case ShapeRing:
		return shapeRing(dx, dy, r)
	case ShapeStar:
		return shapeStar(dx, dy, r)
	case ShapeHexagon:
		return shapeHexagon(dx, dy, r)
	case ShapeDish:
		return shapeDish(dx, dy, r)
	default:
		return shapeSquare(dx, dy, r)
	}
}

// drawBuildingShape renders a building using its visual definition with era-specific effects
func drawBuildingShape(img *image.RGBA, w, h int, b bldInfo, vis BuildingVisual, era, dl int) {
	r := b.size / 2
	if r < 1 {
		r = 1
	}

	primary := vis.Primary
	accent := vis.Accent

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			inside, isAccent := testShape(vis.Shape, dx, dy, r)
			if !inside {
				continue
			}
			px, py := b.x+dx, b.y+dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}

			// Base color
			bc := primary
			if isAccent {
				bc = accent
			}

			// Edge detection (darken outline pixels)
			isEdge := false
			if r > 1 {
				for _, dd := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					ni, _ := testShape(vis.Shape, dx+dd[0], dy+dd[1], r)
					if !ni {
						isEdge = true
						break
					}
				}
			}

			// Apply era-specific material effects
			switch {
			case era <= 1: // primitive/ancient: soften, organic
				// Slight color noise for organic feel
				n := noise2D(px, py, 555)
				bc = lerp(bc, lerp(primary, accent, 0.5), n*0.15)
				if isEdge {
					bc = lerp(bc, c(60, 45, 25), 0.3)
				}

			case era == 2: // medieval: stone texture, dark outlines
				if !isAccent && (dx+dy)%3 == 0 {
					bc = lerp(bc, c(0, 0, 0), 0.08) // stone mortar lines
				}
				if isEdge {
					bc = lerp(bc, c(0, 0, 0), 0.45)
				}

			case era == 3: // industrial: brick pattern, soot
				if !isAccent && (dx+dy)%3 == 0 {
					bc = lerp(bc, c(0, 0, 0), 0.10) // brick mortar
				}
				// Soot overlay
				bc = lerp(bc, c(40, 38, 35), 0.08)
				if isEdge {
					bc = lerp(bc, c(0, 0, 0), 0.4)
				}

			case era <= 5: // modern/digital: clean edges, window grid
				if !isAccent && r > 2 && dx%3 == 0 && dy%3 == 0 {
					bc = lerp(bc, c(40, 45, 55), 0.35) // window grid
				}
				if isEdge {
					bc = lerp(bc, c(30, 35, 40), 0.5)
				}
				// Glass sheen on accent
				if isAccent {
					highlight := 1.0 - float64(dy+r)/float64(r*2)
					bc = lerp(bc, c(200, 210, 230), highlight*0.2)
				}

			case era == 6: // cyberpunk: neon outline, dark body
				if !isEdge {
					bc = lerp(bc, c(10, 10, 15), 0.55) // darken body
					// Scattered lit windows
					if (dx*3+dy*7)%11 < 2 {
						bc = lerp(bc, accent, 0.5)
					}
				} else {
					// Neon glow outline
					bc = accent
					if (dx+dy)%4 < 2 {
						bc = lerp(accent, c(255, 255, 255), 0.3)
					}
				}

			default: // space/cosmic (era 7-8): dome highlight, energy glow
				// Dome-like highlight from top
				t := float64(dy+r) / float64(r*2)
				bc = lerp(bc, c(200, 210, 240), (1.0-t)*0.25)
				if isEdge {
					// Energy glow rim
					bc = lerp(accent, c(255, 255, 255), 0.2)
				}
			}

			img.SetRGBA(px, py, bc)
		}
	}
}


// ─── Infrastructure drawing helpers ──────────────────────

func drawPath(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px >= 0 && px < w && py >= 0 && py < h {
			existing := img.RGBAAt(px, py)
			img.SetRGBA(px, py, lerp(existing, roadC, 0.5))
		}
	}
}

func drawRoad(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC, edgeC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	roadW := 1 + dl
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for rw := -roadW; rw <= roadW; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx < 0 || rpx >= w || rpy < 0 || rpy >= h {
				continue
			}
			if rw == -roadW || rw == roadW {
				existing := img.RGBAAt(rpx, rpy)
				img.SetRGBA(rpx, rpy, lerp(existing, edgeC, 0.5))
			} else {
				img.SetRGBA(rpx, rpy, roadC)
			}
		}
	}
}

func drawRailway(img *image.RGBA, x0, y0, x1, y1, w, h int, trackC, tieC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px < 0 || px >= w || py < 0 || py >= h {
			continue
		}
		// Rails
		img.SetRGBA(px, py, trackC)
		// Sleeper ties every few pixels perpendicular
		if i%4 == 0 {
			for rw := -2; rw <= 2; rw++ {
				rpx := px + int(float64(rw)*(-dy/dist))
				rpy := py + int(float64(rw)*(dx/dist))
				if rpx >= 0 && rpx < w && rpy >= 0 && rpy < h {
					img.SetRGBA(rpx, rpy, tieC)
				}
			}
		}
	}
}

func drawHighway(img *image.RGBA, x0, y0, x1, y1, w, h int, roadC, edgeC color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	roadW := 2 + dl
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for rw := -roadW; rw <= roadW; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx < 0 || rpx >= w || rpy < 0 || rpy >= h {
				continue
			}
			if rw == -roadW || rw == roadW {
				img.SetRGBA(rpx, rpy, edgeC)
			} else if rw == 0 && i%6 < 3 {
				// Dashed center line
				img.SetRGBA(rpx, rpy, c(220, 220, 50))
			} else {
				img.SetRGBA(rpx, rpy, roadC)
			}
		}
	}
}

func drawNeonTrail(img *image.RGBA, x0, y0, x1, y1, w, h int, neon1, neon2 color.RGBA, dl int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		if px < 0 || px >= w || py < 0 || py >= h {
			continue
		}
		// Core bright line
		nc := neon1
		if i%8 < 4 {
			nc = neon2
		}
		img.SetRGBA(px, py, nc)
		// Glow around the line
		for rw := -1; rw <= 1; rw++ {
			rpx := px + int(float64(rw)*(-dy/dist))
			rpy := py + int(float64(rw)*(dx/dist))
			if rpx >= 0 && rpx < w && rpy >= 0 && rpy < h && rw != 0 {
				existing := img.RGBAAt(rpx, rpy)
				img.SetRGBA(rpx, rpy, lerp(existing, nc, 0.25))
			}
		}
	}
}

// ─── Decorations ─────────────────────────────────────────
func drawDecorations(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, seed uint64, placements []bldInfo) {
	switch {
	case era == 3:
		// Smokestacks above production buildings
		smokeC := c(100, 100, 105)
		for _, b := range placements {
			if b.category != "production" {
				continue
			}
			// Chimney
			chimneyX := b.x + b.size/4
			for cy := b.y - b.size/2 - 4; cy < b.y-b.size/2; cy++ {
				if cy >= 0 && cy < h && chimneyX >= 0 && chimneyX < w {
					img.SetRGBA(chimneyX, cy, c(80, 50, 35))
					if chimneyX+1 < w {
						img.SetRGBA(chimneyX+1, cy, c(80, 50, 35))
					}
				}
			}
			// Smoke puff
			for dy := -3; dy <= 0; dy++ {
				for dx := -2; dx <= 2; dx++ {
					px, py := chimneyX+dx, b.y-b.size/2-5+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						existing := img.RGBAAt(px, py)
						img.SetRGBA(px, py, lerp(existing, smokeC, 0.3))
					}
				}
			}
		}

	case era == 4 || era == 5:
		// Power lines between some buildings
		lineC := c(60, 60, 65)
		for i := 0; i+1 < len(placements); i += 2 {
			a, b := placements[i], placements[i+1]
			drawPowerLine(img, a.x, a.y-a.size/2, b.x, b.y-b.size/2, w, h, lineC)
		}

	case era == 6:
		// Holographic billboards above some buildings
		for _, b := range placements {
			if b.category != "housing" && b.category != "production" {
				continue
			}
			bHash := hashKey(b.key + "holo")
			if bHash%3 != 0 {
				continue
			}
			holoW, holoH := 4+dl*2, 2+dl
			holoC := pal.Accent1
			if bHash%2 == 0 {
				holoC = pal.Accent2
			}
			for dy := 0; dy < holoH; dy++ {
				for dx := 0; dx < holoW; dx++ {
					px := b.x - holoW/2 + dx
					py := b.y - b.size - 2 + dy
					if px >= 0 && px < w && py >= 0 && py < h {
						existing := img.RGBAAt(px, py)
						img.SetRGBA(px, py, lerp(existing, holoC, 0.6))
					}
				}
			}
		}
	}
}

func drawPowerLine(img *image.RGBA, x0, y0, x1, y1, w, h int, lineC color.RGBA) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		// Slight sag (catenary)
		sag := math.Sin(t*math.Pi) * 3.0
		py := y0 + int(dy*t+sag)
		if px >= 0 && px < w && py >= 0 && py < h {
			img.SetRGBA(px, py, lineC)
		}
	}
}

func settlementLabel(buildingCount int) string {
	switch {
	case buildingCount <= 5:
		return "Village"
	case buildingCount <= 20:
		return "Town"
	case buildingCount <= 50:
		return "City"
	default:
		return "Metropolis"
	}
}
