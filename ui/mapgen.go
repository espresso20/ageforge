package ui

import (
	"hash/fnv"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TerrainPalette defines the color palette for a terrain era
type TerrainPalette struct {
	Ground, GroundAlt            color.RGBA
	Water, WaterLight, WaterDeep color.RGBA
	Tree, TreeDark, TreeLight    color.RGBA
	Road, RoadEdge               color.RGBA
	Hill, HillLight              color.RGBA
	Farmland, FarmAlt            color.RGBA
	Accent1, Accent2             color.RGBA // era-specific accents
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
	domain   string
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
	"hut":             {ShapeCircle, c(160, 120, 70), c(190, 170, 60)},
	"house":           {ShapeSquare, c(140, 90, 50), c(170, 80, 50)},
	"manor":           {ShapeTriangle, c(140, 140, 130), c(130, 40, 35)},
	"apartment":       {ShapeTower, c(180, 165, 140), c(120, 85, 55)},
	"skyscraper":      {ShapeTower, c(130, 160, 190), c(220, 220, 225)},
	"neon_tower":      {ShapeTower, c(60, 20, 80), c(255, 50, 150)},
	"orbital_habitat": {ShapeCircle, c(180, 185, 190), c(140, 180, 220)},

	// ── Production (warm yellows/oranges/distinct) ──
	"gathering_camp":       {ShapeTriangle, c(100, 110, 50), c(110, 80, 40)},
	"woodcutter_camp":      {ShapeTriangle, c(120, 80, 40), c(40, 80, 30)},
	"stone_pit":            {ShapeWide, c(130, 130, 125), c(90, 90, 85)},
	"farm":                 {ShapeWide, c(140, 170, 50), c(190, 170, 40)},
	"lumber_mill":          {ShapeLShape, c(120, 80, 40), c(80, 55, 30)},
	"quarry":               {ShapeWide, c(120, 115, 100), c(100, 80, 55)},
	"mine":                 {ShapeDiamond, c(80, 80, 80), c(190, 150, 40)},
	"market":               {ShapeDiamond, c(200, 170, 40), c(180, 50, 40)},
	"coal_mine":            {ShapeDiamond, c(55, 50, 45), c(200, 120, 30)},
	"smithy":               {ShapeLShape, c(140, 40, 30), c(210, 140, 30)},
	"forum":                {ShapeSquare, c(220, 210, 180), c(200, 170, 40)},
	"aqueduct":             {ShapeWide, c(160, 155, 140), c(60, 100, 180)},
	"amphitheater":         {ShapeRing, c(190, 170, 130), c(170, 90, 55)},
	"cathedral":            {ShapeCross, c(160, 155, 140), c(200, 170, 40)},
	"art_studio":           {ShapeSquare, c(170, 150, 190), c(190, 120, 140)},
	"bank":                 {ShapeSquare, c(200, 170, 40), c(40, 80, 50)},
	"colony":               {ShapeSquare, c(170, 150, 110), c(80, 130, 60)},
	"port":                 {ShapeWide, c(110, 120, 135), c(100, 70, 40)},
	"plantation":           {ShapeWide, c(40, 100, 40), c(200, 180, 40)},
	"factory":              {ShapeLShape, c(150, 60, 40), c(110, 110, 110)},
	"oil_well":             {ShapeTower, c(30, 30, 30), c(210, 130, 20)},
	"power_grid":           {ShapeCross, c(200, 190, 40), c(130, 130, 130)},
	"telegraph":            {ShapeDish, c(110, 80, 40), c(170, 110, 50)},
	"clocktower":           {ShapeTower, c(150, 145, 130), c(200, 170, 40)},
	"electric_mill":        {ShapeLShape, c(140, 150, 160), c(170, 110, 50)},
	"train_station":        {ShapeWide, c(120, 80, 40), c(170, 40, 30)},
	"reactor":              {ShapeCircle, c(140, 140, 140), c(60, 180, 60)},
	"power_plant":          {ShapeLShape, c(160, 160, 155), c(200, 190, 40)},
	"server_farm":          {ShapeSquare, c(30, 50, 100), c(40, 160, 60)},
	"fiber_hub":            {ShapeHexagon, c(40, 140, 130), c(220, 220, 225)},
	"media_center":         {ShapeSquare, c(220, 220, 225), c(60, 100, 200)},
	"data_center":          {ShapeSquare, c(60, 70, 85), c(40, 190, 200)},
	"smart_grid":           {ShapeCross, c(170, 175, 180), c(40, 190, 200)},
	"augmentation_clinic":  {ShapeCross, c(220, 220, 225), c(40, 220, 80)},
	"black_market":         {ShapeDiamond, c(70, 70, 70), c(200, 40, 160)},
	"fusion_reactor":       {ShapeCircle, c(60, 100, 200), c(240, 230, 200)},
	"plasma_forge":         {ShapeLShape, c(220, 140, 30), c(240, 230, 200)},
	"maglev_station":       {ShapeWide, c(170, 175, 180), c(60, 100, 200)},
	"launch_pad":           {ShapeWide, c(160, 160, 155), c(220, 130, 30)},
	"warp_gate":            {ShapeRing, c(60, 40, 130), c(230, 230, 240)},
	"colony_ship":          {ShapeTriangle, c(170, 175, 180), c(60, 100, 200)},
	"star_forge":           {ShapeHexagon, c(220, 140, 30), c(240, 230, 200)},
	"galactic_hub":         {ShapeHexagon, c(200, 170, 40), c(60, 100, 200)},
	"antimatter_plant":     {ShapeCircle, c(60, 20, 80), c(220, 100, 160)},
	"megastructure":        {ShapeStar, c(170, 175, 180), c(200, 170, 40)},
	"reality_engine":       {ShapeHexagon, c(120, 50, 160), c(230, 230, 240)},
	"transcendence_beacon": {ShapeTower, c(210, 190, 50), c(240, 240, 240)},

	// ── Research (blues/cyans) ──
	"altar":              {ShapeDiamond, c(90, 85, 80), c(50, 70, 170)},
	"firepit":            {ShapeCircle, c(210, 130, 30), c(180, 40, 20)},
	"library":            {ShapeSquare, c(110, 70, 35), c(50, 70, 170)},
	"university":         {ShapeTriangle, c(150, 145, 130), c(40, 60, 160)},
	"observatory":        {ShapeDish, c(30, 40, 100), c(170, 175, 180)},
	"telephone_exchange": {ShapeSquare, c(110, 80, 40), c(170, 110, 50)},
	"research_lab":       {ShapeCross, c(220, 220, 225), c(40, 80, 200)},
	"space_station":      {ShapeRing, c(170, 175, 180), c(60, 100, 200)},
	"ai_lab":             {ShapeHexagon, c(25, 40, 100), c(40, 190, 200)},
	"quantum_computer":   {ShapeHexagon, c(60, 30, 120), c(140, 60, 200)},

	// ── Military (reds/dark) ──
	"barracks":     {ShapeLShape, c(120, 35, 30), c(100, 100, 100)},
	"castle":       {ShapeSquare, c(140, 135, 120), c(70, 65, 60)},
	"bunker":       {ShapeSquare, c(140, 140, 135), c(50, 70, 45)},
	"missile_silo": {ShapeTower, c(120, 120, 120), c(180, 40, 30)},

	// ── Storage (purples/grays) ──
	"stash":              {ShapeCircle, c(160, 130, 90), c(120, 85, 55)},
	"storage_pit":        {ShapeWide, c(120, 95, 60), c(100, 75, 45)},
	"warehouse":          {ShapeSquare, c(120, 80, 40), c(110, 110, 110)},
	"granary":            {ShapeCircle, c(200, 170, 50), c(120, 85, 55)},
	"classical_vault":    {ShapeSquare, c(200, 195, 185), c(200, 170, 40)},
	"keep":               {ShapeSquare, c(130, 125, 115), c(60, 55, 50)},
	"renaissance_vault":  {ShapeSquare, c(220, 210, 185), c(200, 170, 40)},
	"colonial_warehouse": {ShapeSquare, c(120, 80, 40), c(160, 40, 30)},
	"industrial_depot":   {ShapeLShape, c(120, 120, 120), c(100, 75, 45)},
	"victorian_vault":    {ShapeSquare, c(120, 35, 30), c(200, 170, 40)},
	"electric_warehouse": {ShapeSquare, c(140, 150, 160), c(200, 190, 40)},
	"atomic_vault":       {ShapeSquare, c(160, 160, 155), c(60, 150, 60)},
	"modern_depot":       {ShapeSquare, c(120, 120, 125), c(60, 100, 180)},
	"info_vault":         {ShapeSquare, c(30, 40, 90), c(100, 50, 140)},
	"digital_archive":    {ShapeHexagon, c(60, 70, 85), c(40, 190, 200)},
	"cyber_vault":        {ShapeDiamond, c(50, 50, 50), c(200, 40, 160)},
	"fusion_vault":       {ShapeCircle, c(60, 100, 200), c(230, 230, 240)},
	"orbital_depot":      {ShapeSquare, c(170, 175, 180), c(60, 100, 200)},
	"stellar_vault":      {ShapeDiamond, c(200, 170, 40), c(60, 100, 200)},
	"galactic_vault":     {ShapeHexagon, c(100, 50, 140), c(200, 170, 40)},
	"quantum_vault":      {ShapeDiamond, c(60, 30, 120), c(140, 60, 200)},

	// ── Wonders (vivid, unique) ──
	"sacred_grove":         {ShapeCircle, c(30, 120, 40), c(80, 200, 60)},
	"great_monolith":       {ShapeTower, c(140, 140, 135), c(180, 175, 165)},
	"stonehenge":           {ShapeRing, c(140, 140, 135), c(60, 80, 180)},
	"colosseum":            {ShapeRing, c(190, 170, 130), c(170, 50, 40)},
	"parthenon":            {ShapeTriangle, c(210, 205, 195), c(200, 170, 40)},
	"great_library":        {ShapeSquare, c(120, 80, 40), c(200, 170, 40)},
	"sistine_chapel":       {ShapeCross, c(200, 190, 170), c(180, 40, 50)},
	"grand_lighthouse":     {ShapeTower, c(220, 210, 180), c(240, 220, 80)},
	"crystal_palace":       {ShapeDiamond, c(180, 210, 230), c(220, 240, 255)},
	"eiffel_tower":         {ShapeTower, c(120, 110, 100), c(200, 180, 140)},
	"hoover_dam":           {ShapeWide, c(160, 160, 170), c(60, 140, 200)},
	"particle_accelerator": {ShapeRing, c(140, 150, 160), c(40, 170, 60)},
	"space_program":        {ShapeStar, c(230, 230, 235), c(220, 140, 30)},
	"global_network":       {ShapeHexagon, c(20, 60, 100), c(0, 200, 255)},
	"world_simulation":     {ShapeHexagon, c(15, 30, 70), c(0, 180, 220)},
	"neon_citadel":         {ShapeStar, c(40, 10, 60), c(255, 0, 180)},
	"stellar_cradle":       {ShapeCircle, c(220, 160, 40), c(255, 200, 80)},
	"dyson_scaffold":       {ShapeStar, c(210, 180, 40), c(220, 140, 30)},
	"warp_nexus":           {ShapeHexagon, c(80, 40, 160), c(160, 80, 255)},
	"cosmic_beacon":        {ShapeDish, c(100, 60, 180), c(200, 140, 255)},
	"reality_anchor":       {ShapeDiamond, c(60, 30, 120), c(180, 100, 255)},
	"singularity_core":     {ShapeStar, c(20, 20, 25), c(240, 240, 245)},
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
	riverBaseX := float64(w) * 0.46
	riverW := 5 + dl*4
	bankW := 2 + dl
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
	placements := collectBuildingPlacements(img, cfg, newRNG(seed), cx, cy, w, h, era, dl)

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
	drawBuildings(img, w, h, era, dl, pal, cfg.AgeKey, placements)

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

// ─── RNG helper ──────────────────────────────────────────
func newRNG(seed uint64) *rand.Rand {
	return rand.New(rand.NewSource(int64(seed))) //nolint:gosec
}

// ─── Plot grid (collision avoidance) ─────────────────────
type plotGrid struct {
	occupied map[[2]int]bool
	cellSize int
}

func newPlotGrid(cellSize int) *plotGrid {
	return &plotGrid{occupied: make(map[[2]int]bool), cellSize: cellSize}
}

func (g *plotGrid) claim(px, py, w, h int) {
	cx0 := px / g.cellSize
	cy0 := py / g.cellSize
	cx1 := (px + w) / g.cellSize
	cy1 := (py + h) / g.cellSize
	for cx := cx0; cx <= cx1; cx++ {
		for cy := cy0; cy <= cy1; cy++ {
			g.occupied[[2]int{cx, cy}] = true
		}
	}
}

func (g *plotGrid) isFree(px, py, w, h int) bool {
	cx0 := (px - g.cellSize*2) / g.cellSize
	cy0 := (py - g.cellSize*2) / g.cellSize
	cx1 := (px + w + g.cellSize*2) / g.cellSize
	cy1 := (py + h + g.cellSize*2) / g.cellSize
	if cx0 < 0 {
		cx0 = 0
	}
	if cy0 < 0 {
		cy0 = 0
	}
	for cx := cx0; cx <= cx1; cx++ {
		for cy := cy0; cy <= cy1; cy++ {
			if g.occupied[[2]int{cx, cy}] {
				return false
			}
		}
	}
	return true
}

// ─── Era name mapping ─────────────────────────────────────
func getEraName(ageKey string) string {
	switch ageKey {
	case "primitive_age":
		return "primitive"
	case "stone_age":
		return "stone"
	case "bronze_age":
		return "bronze"
	case "iron_age":
		return "iron"
	case "classical_age":
		return "classical"
	case "medieval_age":
		return "medieval"
	case "renaissance_age":
		return "renaissance"
	case "colonial_age":
		return "colonial"
	case "industrial_age", "victorian_age":
		return "industrial"
	case "electric_age":
		return "industrial"
	case "atomic_age":
		return "atomic"
	case "modern_age", "information_age":
		return "modern"
	case "digital_age":
		return "digital"
	case "cyberpunk_age", "fusion_age":
		return "nano"
	case "space_age", "interstellar_age":
		return "space"
	case "galactic_age", "quantum_age", "transcendent_age":
		return "galactic"
	default:
		return "stone"
	}
}

// ─── Sprite types ─────────────────────────────────────────
type spriteType int

const (
	spriteHut spriteType = iota
	spriteFarm
	spriteMill
	spriteLumberCamp
	spriteMine
	spriteFortress
	spriteBarracks
	spriteTemple
	spriteLibrary
	spriteMarket
	spriteFactory
	spriteWorkshop
	spritePalace
	spriteObservatory
	spriteDome
	spriteSkyscraper
	spriteServer
	spriteSpaceStation
	spriteWonder
	spriteShelter      // lean-to / open-sided structure
	spriteStakeHut     // raised platform hut
	spriteGranary      // rounded storage barn
	spriteStable       // wide low building with pitched roof
	spriteForge        // squat blocky furnace building
	spriteWell         // circular well
	spriteAqueduct     // arched bridge/aqueduct shape
	spriteCathederal   // tall spire church
	spriteArmory       // fortified rectangular storage
	spriteSiege        // siege weapon platform
	spriteMonastery    // cloistered courtyard shape
	spriteApothecary   // narrow tall shop
	spriteTavern       // wide building with sign-post
	spriteHarbour      // dock/pier extending out
	spriteMintHouse    // compact official building
	spriteGuildHall    // wide official hall
	spritePrintShop    // narrow two-story
	spriteGlassworks   // wide factory with tall chimney
	spriteCoalMine     // mine entrance with cart track
	spriteSteelMill    // large industrial with multiple chimneys
	spriteOilDerrick   // tall narrow derrick shape
	spritePowerPlant   // industrial + towers
	spriteResearchLab  // modern clean building
	spriteHospital     // wide building with cross marker
	spriteDataCenter   // dense rack-row building
	spriteAntenna      // tall narrow antenna/tower
	spriteReactor      // dome + cooling towers
	spriteSpaceDock    // horizontal launch pad
	spriteCrystalSpire // tall narrow crystalline
)

func getBuildingSprite(domain, buildingKey, eraName string) spriteType {
	// Per-building-key overrides first (takes priority over domain+era fallback)
	switch buildingKey {
	// ── Food domain ──
	case "gathering_camp":
		return spriteHut
	case "forager_post":
		return spriteShelter
	case "hunting_lodge":
		return spriteStakeHut
	case "farm", "estate_farm", "market_garden", "plantation", "industrial_farm",
		"mechanized_farm", "agricultural_complex", "agricultural_works", "agri_complex":
		return spriteFarm
	case "vat_farm", "nano_farm", "smart_farm", "hydroponic_bay", "protein_synthesizer",
		"bio_reactor_farm", "quantum_cultivator", "quantum_organic_extractor",
		"cosmic_organic_works":
		return spriteResearchLab
	case "granary":
		return spriteGranary

	// ── Lumber domain ──
	case "wood_camp", "woodcutter_camp", "lumber_camp", "logging_station":
		return spriteLumberCamp
	case "sawmill", "lumber_mill", "timber_yard", "mill":
		return spriteMill
	case "wood_workshop":
		return spriteWorkshop
	case "paper_mill":
		return spriteGlassworks

	// ── Masonry / Stone domain ──
	case "stone_camp", "stone_pit":
		return spriteMine
	case "quarry", "marble_quarry":
		return spriteMine
	case "stonemasons_guild":
		return spriteWorkshop
	case "kiln", "smelter", "bloomery", "smithy", "ironmonger", "ironworks",
		"iron_works", "forge":
		return spriteForge
	case "brickworks", "marble_works":
		return spriteFactory
	case "cement_plant", "bessemer_plant", "iron_works_complex", "steel_mill":
		return spriteSteelMill

	// ── Military domain ──
	case "barracks", "garrison", "fort", "war_camp", "field_works", "command_post":
		return spriteBarracks
	case "archery_range", "elders_hall":
		return spriteBarracks
	case "stable":
		return spriteStable
	case "siege_workshop":
		return spriteSiege
	case "fortress", "castle_keep", "bunker_complex", "legion_fort":
		return spriteFortress
	case "armory":
		return spriteArmory
	case "military_academy", "war_academy":
		return spriteGuildHall
	case "dockyard", "naval_yard":
		return spriteHarbour
	case "military_base", "space_force_base", "fleet_command", "stellar_armada_hq",
		"drone_warfare_center", "combat_aug_center", "special_ops_hq",
		"plasma_command", "omniversal_war_council", "probability_war_room":
		return spriteFortress

	// ── Knowledge domain ──
	case "altar", "story_circle", "standing_stones":
		return spriteTemple
	case "scriptorium", "natural_philosophy_hall":
		return spriteLibrary
	case "library", "monastery_library", "civilization_archive":
		return spriteLibrary
	case "scholars_hall", "guildhall", "great_hall":
		return spriteGuildHall
	case "observatory", "deep_space_observatory":
		return spriteObservatory
	case "academy":
		return spriteLibrary
	case "university", "think_tank", "theoretical_institute":
		return spriteGuildHall
	case "printing_press":
		return spritePrintShop
	case "research_institute", "research_campus", "innovation_hub", "physics_laboratory",
		"ai_research_lab", "neuro_research_center":
		return spriteResearchLab
	case "xenology_institute", "reality_academy":
		return spriteResearchLab
	case "neural_grid":
		return spriteDataCenter
	case "cosmic_research_station":
		return spriteResearchLab

	// ── Faith domain ──
	case "shrine", "firepit":
		return spriteTemple
	case "temple":
		return spriteTemple
	case "church", "mission", "meditation_center", "spiritual_center":
		return spriteTemple
	case "monastery":
		return spriteMonastery
	case "cathedral", "grand_cathedral", "basilica":
		return spriteCathederal
	case "seminary":
		return spriteApothecary
	case "pilgrimage_site":
		return spriteWell
	case "oracle_house":
		return spriteObservatory
	case "revival_hall":
		return spriteGuildHall
	case "neon_sanctuary", "cyber_shrine", "digital_temple":
		return spriteCrystalSpire
	case "void_monastery":
		return spriteMonastery
	case "orbital_sanctuary", "stellar_shrine":
		return spriteSpaceStation
	case "quantum_chapel":
		return spriteCrystalSpire
	case "astral_chapel":
		return spriteCrystalSpire

	// ── Trade domain ──
	case "market", "agora", "bazaar":
		return spriteMarket
	case "trading_post", "merchant_quarter":
		return spriteMintHouse
	case "bank", "investment_firm", "financial_district":
		return spriteMintHouse
	case "stock_exchange", "exchange", "crypto_exchange", "energy_exchange":
		return spriteGuildHall
	case "port", "harbour":
		return spriteHarbour
	case "venture_hub", "corporate_hq":
		return spriteGuildHall
	case "stellar_exchange", "omniversal_bazaar", "probability_market",
		"asteroid_market", "galactic_trade_hub":
		return spriteSkyscraper

	// ── Engineering domain ──
	case "workshop":
		return spriteWorkshop
	case "aqueduct":
		return spriteAqueduct
	case "waterwheel":
		return spriteAqueduct
	case "steam_works", "steam_turbine":
		return spriteFactory
	case "machine_shop", "assembly_plant", "electric_arc_furnace":
		return spriteFactory
	case "robotics_lab", "nanobot_vat":
		return spriteResearchLab
	case "nano_drill_complex", "molecular_synthesizer", "matter_converter",
		"exotic_matter_forge", "antimatter_forge", "reality_forge":
		return spriteReactor

	// ── Metallurgy domain ──
	case "foundry", "plasma_forge", "augmentation_foundry", "aerospace_foundry":
		return spriteSteelMill
	case "coal_mine", "steam_mine":
		return spriteCoalMine
	case "iron_mine", "deep_iron_mine", "ore_mine", "uranium_mine",
		"titanium_mine", "dark_crystal_mine", "neutron_star_mine",
		"asteroid_crystal_mine", "stellar_core_drill":
		return spriteMine
	case "alloy_plant", "advanced_alloy_plant", "nano_alloy_plant", "quantum_metal_works":
		return spriteGlassworks
	case "coal_works", "steel_era", "iron_era", "steel_works":
		return spriteSteelMill
	case "titanium_smelter", "uranium_processing_works":
		return spriteSteelMill
	case "geological_extraction", "precision_mine", "organic_extraction",
		"nuclear_extraction_plant", "dark_matter_refinery", "reality_excavator",
		"reality_matter_weaver":
		return spriteMine

	// ── Energy domain ──
	case "bonfire":
		return spriteWell
	case "coal_plant", "steam_coal_plant":
		return spriteFactory
	case "oil_derrick", "oil_field", "oil_platform":
		return spriteOilDerrick
	case "oil_refinery", "petroleum_refinery", "smart_refinery":
		return spriteOilDerrick
	case "power_station", "power_generator", "power_grid_hub", "smart_energy_grid",
		"smart_grid_node":
		return spritePowerPlant
	case "nuclear_plant", "nuclear_reactor":
		return spriteReactor
	case "solar_collector_array":
		return spriteDome
	case "fusion_reactor", "fusion_reactor_array":
		return spriteReactor
	case "quantum_battery_array", "zero_point_generator":
		return spriteReactor
	case "dark_energy_tap", "pulsar_tap", "quasar_tap":
		return spriteCrystalSpire
	case "dyson_assembly":
		return spriteWonder

	// ── Hacker domain ──
	case "terminal", "cyber_hub":
		return spriteServer
	case "server_farm", "quantum_server_farm":
		return spriteDataCenter
	case "dark_web_hub", "black_market":
		return spriteServer
	case "ai_core", "neural_art_complex":
		return spriteServer
	case "quantum_net", "galactic_network_node", "orbital_data_relay":
		return spriteCrystalSpire

	// ── Astronaut domain ──
	case "launch_complex", "launch_pad":
		return spriteSpaceDock
	case "space_station", "orbital_habitat", "orbital_platform", "habitat_ring",
		"orbital_refinery", "generation_ship":
		return spriteSpaceStation
	case "warp_gate", "warp_drive_plant":
		return spriteCrystalSpire
	case "dyson_sphere_habitat":
		return spriteWonder
	case "arcology_pod", "megaplex":
		return spriteDome
	}

	// Check wonders by building key
	wonderKeys := map[string]bool{
		"sacred_grove": true, "great_monolith": true, "stonehenge": true,
		"colosseum": true, "parthenon": true, "great_library": true,
		"sistine_chapel": true, "grand_lighthouse": true, "crystal_palace": true,
		"eiffel_tower": true, "hoover_dam": true, "particle_accelerator": true,
		"space_program": true, "global_network": true, "world_simulation": true,
		"neon_citadel": true, "stellar_cradle": true, "dyson_scaffold": true,
		"warp_nexus": true, "cosmic_beacon": true, "reality_anchor": true,
		"singularity_core": true,
	}
	if wonderKeys[buildingKey] {
		return spriteWonder
	}

	isEarlyEra := eraName == "primitive" || eraName == "stone" || eraName == "bronze" || eraName == "iron" || eraName == "classical"
	isLateEra := eraName == "space" || eraName == "galactic" || eraName == "nano"
	isDigitalEra := eraName == "digital" || eraName == "nano"

	switch domain {
	case "food":
		if isEarlyEra {
			return spriteHut
		}
		return spriteFarm
	case "lumber":
		return spriteLumberCamp
	case "masonry":
		return spriteMine
	case "military":
		if isEarlyEra {
			return spriteBarracks
		}
		return spriteFortress
	case "knowledge":
		if isLateEra {
			return spriteObservatory
		}
		if isDigitalEra {
			return spriteServer
		}
		return spriteLibrary
	case "faith":
		return spriteTemple
	case "trade":
		return spriteMarket
	case "engineering", "metallurgy":
		if isEarlyEra {
			return spriteWorkshop
		}
		return spriteFactory
	case "energy":
		if isLateEra {
			return spriteDome
		}
		return spriteWorkshop
	case "hacker":
		return spriteServer
	case "astronaut":
		return spriteSpaceStation
	}

	// Fallback by building key patterns
	switch buildingKey {
	case "skyscraper", "neon_tower", "apartment":
		return spriteSkyscraper
	case "observatory", "quantum_computer", "ai_lab":
		return spriteObservatory
	case "reactor", "fusion_reactor", "antimatter_plant":
		return spriteDome
	case "server_farm", "data_center", "fiber_hub":
		return spriteServer
	case "space_station", "orbital_habitat", "launch_pad", "warp_gate":
		return spriteSpaceStation
	}

	return spriteHut
}

// drawBuildingSprite renders a pixel-art sprite at (px, py). scale=1 for minimap, scale=2 for full map.
// Wonders use scale 3.
func drawBuildingSprite(img *image.RGBA, imgW, imgH, px, py int, stype spriteType, primary, accent color.RGBA, scale int) {
	// For minimap (scale <= 1), draw a solid 3×3 block — sprites are too small to matter.
	if scale <= 1 {
		for dy := 0; dy < 3; dy++ {
			for dx := 0; dx < 3; dx++ {
				x, y := px+dx, py+dy
				if x >= 0 && x < imgW && y >= 0 && y < imgH {
					img.SetRGBA(x, y, primary)
				}
			}
		}
		return
	}

	rows := spriteRows(stype)
	for row, line := range rows {
		for col, ch := range line {
			var cl color.RGBA
			switch ch {
			case 'P':
				cl = primary
			case 'A', 'I':
				cl = accent
			default:
				continue
			}
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					x := px + col*scale + sx
					y := py + row*scale + sy
					if x >= 0 && x < imgW && y >= 0 && y < imgH {
						img.SetRGBA(x, y, cl)
					}
				}
			}
		}
	}
}

// spriteRows returns the pixel-art row strings for a given sprite type.
// This is the single source of truth used by both drawBuildingSprite and
// the shadow-sizing helpers.
func spriteRows(stype spriteType) []string {
	switch stype {
	case spriteHut:
		return []string{"..P..", ".PPP.", "PPPPP", "A.A.A", "AAAAA"}
	case spriteFarm:
		return []string{"AAAAAAA", "PPPPPPP", "AAAAAAA", "PPPPPPP", "..A.A.."}
	case spriteLumberCamp:
		return []string{"..A..", "..A..", "PPPPP", "P.P.P", "PPPPP", ".PPP."}
	case spriteMine:
		return []string{"PPPPPP", "P.AAP.", "P....P", "AAAAAA"}
	case spriteBarracks:
		return []string{"PPPPPPP", "P.P.P.P", "PPPPPPP", "AAAAAAA"}
	case spriteFortress:
		return []string{"P.P.P.P", "PPPPPPP", "P.....P", "PPPPPPP", "A.AAA.A"}
	case spriteTemple:
		return []string{"..A..", ".AAA.", "AAAAA", "PPPPP", "P.P.P", "PPPPP", "AAAAA"}
	case spriteLibrary:
		return []string{"AAAAAAA", "PPPPPPP", "P.P.P.P", "PPPPPPP", "P.....P", "AAAAAAA"}
	case spriteMarket:
		return []string{".AAAAA.", "PPPPPPP", "P.P.P.P", "PPPPPPP", "AAAAAAA"}
	case spriteFactory:
		return []string{".A..A..", "AA..AA.", "AA..AA.", "PPPPPPPP", "PPPPPPPP", "P.PP.PP.", "AAAAAAAA"}
	case spriteWorkshop:
		return []string{".PPPP.", "PPPPPP", "P.PP.P", "PPPPPP", "AAAAAA"}
	case spriteObservatory:
		return []string{"..A..", ".AAA.", "AAAAA", "AAAAA", "PPPPP", "P...P", "PPPPP"}
	case spriteDome:
		return []string{".AAAA.", "AAAAAA", "AAAAAA", "PPPPPP", "PP..PP", "PPPPPP"}
	case spriteSkyscraper:
		return []string{"..A..", ".PAP.", ".PPP.", "PPPPP", "P.P.P", "PPPPP", "AAAAA"}
	case spriteServer:
		return []string{"PPPPP", "APPPA", "PPPPP", "APPPA", "PPPPP", "APPPA", "AAAAA"}
	case spriteSpaceStation:
		return []string{"....A....", "...AAA...", "AA.AAA.AA", "AA.AAA.AA", "...AAA...", "....A...."}
	case spriteWonder:
		return []string{"....AA....", "...AAAA...", "..AAAAAA..", "..PPPPPP..", ".PPPPPPPP.", ".PP....PP.", "PPPPPPPPPP", "PPPPPPPPPP", "A.AAAAAA.A", "AAAAAAAAAA"}
	case spriteMill:
		return []string{"..A..", ".AAA.", "PPPPP", "P.P.P", "PPPPP", "AAAAA"}
	case spriteShelter:
		return []string{"AAAAAA", "PPPPPP", "P....P", ".PPPP."}
	case spriteStakeHut:
		return []string{".PPP.", "PPPPP", "PAAAP", "..A..", "..A..", ".A.A."}
	case spriteGranary:
		return []string{".AAAA.", "AAAAAA", "PPPPPP", "PP..PP", "PPPPPP"}
	case spriteStable:
		return []string{"AAAAAAA", "PPPPPPP", "P.P.P.P", "PPPPPPP"}
	case spriteForge:
		return []string{"..A..", "PAAAP", "PPPPP", "AAPAA", "AAAAA"}
	case spriteWell:
		return []string{"A...A", "AAAAA", "A.P.A", "AAAAA", ".AAA."}
	case spriteAqueduct:
		return []string{"AAAAAAA", "AAAAAAA", "A.A.A.A", "P.P.P.P", "PPPPPPP"}
	case spriteCathederal:
		return []string{"..AA..", "..AA..", ".AAAA.", "PPPPPP", "P.PP.P", "PPPPPP", "P....P", "PPPPPP"}
	case spriteArmory:
		return []string{"AAPPAA", "PPPPPP", "P.PP.P", "PPPPPP", "AAAAAA"}
	case spriteSiege:
		return []string{"...A..", "AAAAAP", ".PPPPP", "PPPPPP", ".AA.AA"}
	case spriteMonastery:
		return []string{"PPPPPPP", "P.....P", "P..A..P", "P.AAA.P", "P.....P", "PPPPPPP"}
	case spriteApothecary:
		return []string{".AA.", "PPPP", "P..P", "PPPP", "P..P", "AAAA"}
	case spriteTavern:
		return []string{"AAAAAAA", "PPPPPPP", "P.PPP.P", "PPPPPPP", "AAAAAAA"}
	case spriteHarbour:
		return []string{"PP.....", "PP.....", "PPPAAAA", "PPPAAAA", "AAAAAAA", "AAAAAAA"}
	case spriteMintHouse:
		return []string{"AAAAA", "PPPPP", "P.A.P", "PPPPP", "AAAAA"}
	case spriteGuildHall:
		return []string{"AAAAAAAA", "PPPPPPPP", "P.P..P.P", "PPPPPPPP", "AAAAAAAA"}
	case spritePrintShop:
		return []string{"AAAAA", "P...P", "PAAAP", "PPPPP", "P.A.P", "AAAAA"}
	case spriteGlassworks:
		return []string{"......AA", "PPPPP.AA", "PPPPPAAA", "PPPPPPPP", "P.P.P.PP", "AAAAAAAA"}
	case spriteCoalMine:
		return []string{"AAAAAAA", "PP...PP", "P.AAA.P", "AAAAAAA", ".AA.AA."}
	case spriteSteelMill:
		return []string{"...A..A..", "PPPA.AAAP", "PPPAAAAPP", "PPPPPPPPP", "PPP...PPP", "PPPPPPPPP", "AAAAAAAAA"}
	case spriteOilDerrick:
		return []string{"..A..", "..A..", ".AAA.", "AAAAA", "A...A", "AAAAA", "A...A", "AAAAA"}
	case spritePowerPlant:
		return []string{"AA....AA", "AA....AA", "AAPPPPAA", "AAPPPPAA", "PPPPPPPP", "PP....PP", "PPPPPPPP"}
	case spriteResearchLab:
		return []string{"AAAAAAA", "PPPPPPP", "PA.A.AP", "PPPPPPP", "AAAAAAA"}
	case spriteHospital:
		return []string{"AAAAAAA", "PPPPPPP", "PP.A.PP", "PAAAAPP", "PP.A.PP", "PPPPPPP", "AAAAAAA"}
	case spriteDataCenter:
		return []string{"AAAAAA", "PPPPPP", "AAAAAA", "PPPPPP", "AAAAAA", "PPPPPP"}
	case spriteAntenna:
		return []string{".A.", ".A.", "AAA", ".A.", "AAA", "PAP", "PAP", "PPP"}
	case spriteReactor:
		return []string{"AA.AA...", "AA.AA...", ".AAAAA..", "AAAAAAAA", "PPPPPPPP", "PP....PP", "PPPPPPPP"}
	case spriteSpaceDock:
		return []string{"....A....", "...AAA...", "..AAAAA..", "AAAAAAAAA", "PPPPPPPPP", "PPPPPPPPP"}
	case spriteCrystalSpire:
		return []string{"..A..", ".AAA.", "AAAAA", ".AAA.", "AAPAA", "AAPAA", "AAPAA", "AAAAA"}
	default:
		return []string{"PPP", "PPP", "AAA"}
	}
}

// spriteRowWidth returns the width (in pixels, pre-scale) of the widest row.
func spriteRowWidth(stype spriteType) string {
	rows := spriteRows(stype)
	if len(rows) == 0 {
		return ""
	}
	widest := rows[0]
	for _, r := range rows[1:] {
		if len(r) > len(widest) {
			widest = r
		}
	}
	return widest
}

// spriteRowCount returns the number of rows (height pre-scale).
func spriteRowCount(stype spriteType) int {
	return len(spriteRows(stype))
}

// ─── isWaterPixel checks if a position is water ───────────
func isWaterPixel(img *image.RGBA, x, y int, pal TerrainPalette) bool {
	if x < 0 || y < 0 || x >= img.Bounds().Max.X || y >= img.Bounds().Max.Y {
		return true
	}
	col := img.RGBAAt(x, y)
	return col == pal.Water || col == pal.WaterDeep || col == pal.WaterLight
}

// ─── Internal entry type for layout functions ─────────────
type layoutEntry struct {
	key      string
	category string
	domain   string
	count    int
	size     int
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ─── Building placement entry point ──────────────────────
// collectBuildingPlacements collects buildings and delegates to the
// era-appropriate layout function.
func collectBuildingPlacements(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h, era, dl int) []bldInfo {
	buildings := cfg.Buildings
	eraName := getEraName(cfg.AgeKey)

	// Determine base building size by era
	baseSize := 3
	switch {
	case era >= 6:
		baseSize = 5
	case era >= 3:
		baseSize = 4
	}
	if dl > 0 {
		baseSize += 2
	}

	var entries []layoutEntry
	var wonders []layoutEntry

	for key, bs := range buildings {
		if !bs.Unlocked || bs.Count == 0 {
			continue
		}
		sz := baseSize
		if bs.Category == "military" {
			sz++
		}
		// Show 1 map instance per 5 real buildings, minimum 1 if any exist
		mapCount := (bs.Count + 4) / 5
		if mapCount < 1 {
			mapCount = 1
		}
		if bs.Category == "wonder" {
			wsz := 7 + dl*3
			if era >= 4 {
				wsz += 2
			}
			wonders = append(wonders, layoutEntry{key, bs.Category, bs.WorkerDomain, mapCount, wsz})
		} else {
			entries = append(entries, layoutEntry{key, bs.Category, bs.WorkerDomain, mapCount, sz})
		}
	}

	// Sort entries deterministically by key for reproducibility
	sort.Slice(entries, func(i, j int) bool { return entries[i].key < entries[j].key })
	sort.Slice(wonders, func(i, j int) bool { return wonders[i].key < wonders[j].key })

	// Place city buildings using age-appropriate layout
	cityPlacements := placeAllBuildings(img, cfg, rng, cx, cy, w, h, eraName, entries)
	placements := append([]bldInfo{}, cityPlacements...)

	// Place wonders in outer zone
	maxDist := float64(min(w, h)) / 2.0
	for _, e := range wonders {
		bHash := hashKey(e.key)
		angle := float64(bHash%3600) / 3600.0 * 2.0 * math.Pi
		distRatio := 0.60 + float64(bHash%180)/1000.0
		dist := distRatio * maxDist
		bx := cx + int(math.Cos(angle)*dist)
		by := cy + int(math.Sin(angle)*dist*0.75)
		bx = clampInt(bx, 8, w-8)
		by = clampInt(by, 8, h-8)
		placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, e.size})
	}

	return placements
}

// ─── Age-aware layout dispatcher ─────────────────────────
func placeAllBuildings(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, eraName string, entries []layoutEntry) []bldInfo {
	switch eraName {
	case "primitive", "stone":
		return placeBuildingsOrganic(img, cfg, rng, cx, cy, w, h, entries)
	case "bronze", "iron", "classical":
		return placeBuildingsVillage(img, cfg, rng, cx, cy, w, h, entries)
	case "medieval", "renaissance":
		return placeBuildingsMedieval(img, cfg, rng, cx, cy, w, h, entries)
	case "colonial", "industrial":
		return placeBuildingsIndustrial(img, cfg, rng, cx, cy, w, h, entries)
	case "atomic", "modern":
		return placeBuildingsModern(img, cfg, rng, cx, cy, w, h, entries)
	case "digital", "nano":
		return placeBuildingsCampus(img, cfg, rng, cx, cy, w, h, entries)
	case "space", "galactic":
		return placeBuildingsOrbital(img, cfg, rng, cx, cy, w, h, entries)
	default:
		return placeBuildingsVillage(img, cfg, rng, cx, cy, w, h, entries)
	}
}

// ─── Organic layout (primitive/stone) ────────────────────
func placeBuildingsOrganic(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(12)
	maxR := int(float64(min(w, h)) * 0.30)
	border := 8

	// Seed anchor points — small groups
	type anchor struct{ x, y int }
	var anchors []anchor

	for _, e := range entries {
		for i := 0; i < e.count; i++ {
			size := e.size
			var px, py int
			placed := false

			for attempt := 0; attempt < 30; attempt++ {
				if len(anchors) == 0 || (rng.Intn(3) == 0) {
					// Place near center with random walk
					ang := rng.Float64() * 2.0 * math.Pi
					dist := float64(rng.Intn(maxR))
					px = cx + int(math.Cos(ang)*dist)
					py = cy + int(math.Sin(ang)*dist*0.65)
				} else {
					// Walk from existing anchor
					anch := anchors[rng.Intn(len(anchors))]
					px = anch.x + rng.Intn(31) - 15
					py = anch.y + rng.Intn(21) - 10
				}
				px = clampInt(px, border, w-border)
				py = clampInt(py, border, h-border)
				if isWaterPixel(img, px, py, pal) {
					continue
				}
				if grid.isFree(px, py, size, size) {
					placed = true
					break
				}
			}
			if !placed {
				continue
			}
			grid.claim(px, py, size, size)
			anchors = append(anchors, anchor{px, py})
			placements = append(placements, bldInfo{e.key, e.category, e.domain, px, py, size})
		}
	}
	return placements
}

// drawRoadLine draws a 1-2px road between two points using the palette road color
func drawRoadLine(img *image.RGBA, x0, y0, x1, y1, imgW, imgH, thickness int, col color.RGBA) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	steps := int(dist) + 1
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		px := x0 + int(dx*t)
		py := y0 + int(dy*t)
		for tw := -thickness / 2; tw <= thickness/2; tw++ {
			nx := px + int(float64(tw)*(-dy/dist))
			ny := py + int(float64(tw)*(dx/dist))
			if nx >= 0 && nx < imgW && ny >= 0 && ny < imgH {
				existing := img.RGBAAt(nx, ny)
				img.SetRGBA(nx, ny, lerp(existing, col, 0.65))
			}
		}
	}
}

// ─── Village layout (bronze/iron/classical) ───────────────
func placeBuildingsVillage(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(10)
	border := 8
	maxR := int(float64(min(w, h)) * 0.26)

	// Draw 6 radial spokes from center
	numSpokes := 6
	spokeColors := lerp(pal.Road, pal.Ground, 0.3)
	for s := 0; s < numSpokes; s++ {
		ang := float64(s) / float64(numSpokes) * 2.0 * math.Pi
		ex := cx + int(math.Cos(ang)*float64(maxR))
		ey := cy + int(math.Sin(ang)*float64(maxR)*0.70)
		drawRoadLine(img, cx, cy, ex, ey, w, h, 1, spokeColors)
	}

	// Place buildings along spokes
	idx := 0
	for _, e := range entries {
		for i := 0; i < e.count; i++ {
			size := e.size
			spoke := idx % numSpokes
			ang := float64(spoke)/float64(numSpokes)*2.0*math.Pi
			// Random distance along spoke
			dist := float64(8+rng.Intn(maxR-8))
			side := 1
			if rng.Intn(2) == 0 {
				side = -1
			}
			perpAng := ang + float64(side)*math.Pi/2.0
			offset := float64(3 + rng.Intn(8))
			bx := cx + int(math.Cos(ang)*dist+math.Cos(perpAng)*offset)
			by := cy + int(math.Sin(ang)*dist*0.70+math.Sin(perpAng)*offset)
			bx = clampInt(bx, border, w-border)
			by = clampInt(by, border, h-border)

			placed := false
			for attempt := 0; attempt < 20; attempt++ {
				if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, size, size) {
					placed = true
					break
				}
				bx = clampInt(bx+rng.Intn(11)-5, border, w-border)
				by = clampInt(by+rng.Intn(7)-3, border, h-border)
			}
			if placed {
				grid.claim(bx, by, size, size)
				placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, size})
			}
			idx++
		}
	}
	return placements
}

// ─── Medieval layout (medieval/renaissance) ───────────────
func placeBuildingsMedieval(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(10)
	border := 8
	wallR := 30

	// Draw castle walls (rough square)
	wallColor := lerp(pal.Accent1, pal.Ground, 0.4)
	for i := 0; i <= 360; i++ {
		ang := float64(i) * math.Pi / 180.0
		// Make it square-ish by clamping to the min of x/y components
		wx := math.Cos(ang)
		wy := math.Sin(ang)
		scale := float64(wallR) / math.Max(math.Abs(wx), math.Abs(wy))
		wpx := cx + int(wx*scale)
		wpy := cy + int(wy*scale*0.75)
		if wpx >= 0 && wpx < w && wpy >= 0 && wpy < h {
			img.SetRGBA(wpx, wpy, wallColor)
			if wpx+1 < w {
				img.SetRGBA(wpx+1, wpy, wallColor)
			}
		}
	}

	// Draw roads to 4 quarters
	roadCol := lerp(pal.Road, pal.Ground, 0.25)
	for _, off := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		ex := cx + off[0]*int(float64(min(w, h))*0.45)
		ey := cy + off[1]*int(float64(min(w, h))*0.35)
		drawRoadLine(img, cx, cy, ex, ey, w, h, 1, roadCol)
	}

	// Divide buildings into 4 quarter groups
	quarters := [4][]layoutEntry{}
	for i, e := range entries {
		q := i % 4
		quarters[q] = append(quarters[q], e)
	}

	quarterOffsets := [4][2]int{{1, -1}, {-1, -1}, {1, 1}, {-1, 1}}
	maxR := int(float64(min(w, h)) * 0.30)
	spacing := 10

	for q, qEntries := range quarters {
		qox := quarterOffsets[q][0]
		qoy := quarterOffsets[q][1]
		startX := cx + qox*(wallR+10)
		startY := cy + qoy*(wallR+10)
		col := 0
		row := 0
		for _, e := range qEntries {
			for i := 0; i < e.count; i++ {
				size := e.size
				bx := startX + col*spacing*qox
				by := startY + row*spacing*qoy

				// Clamp into bounds
				bx = clampInt(bx, border, w-border)
				by = clampInt(by, border, h-border)

				// Check we haven't gone too far
				if abs(bx-cx) > maxR || abs(by-cy) > int(float64(maxR)*0.75) {
					col = 0
					row = 0
					bx = startX
					by = startY
				}

				placed := false
				for attempt := 0; attempt < 20; attempt++ {
					if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, size, size) {
						placed = true
						break
					}
					bx = clampInt(bx+rng.Intn(9)-4, border, w-border)
					by = clampInt(by+rng.Intn(7)-3, border, h-border)
				}
				if placed {
					grid.claim(bx, by, size, size)
					placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, size})
				}
				col++
				if col > 3 {
					col = 0
					row++
				}
			}
		}
	}
	return placements
}

// ─── Industrial layout (colonial/industrial) ──────────────
func placeBuildingsIndustrial(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(14)
	border := 8
	cellSpacing := 14

	// Separate into production (left) and residential (right)
	var prodEntries, resEntries []layoutEntry
	prodDomains := map[string]bool{
		"food": true, "lumber": true, "masonry": true, "metallurgy": true, "energy": true,
	}
	for _, e := range entries {
		if prodDomains[e.domain] || e.category == "production" {
			prodEntries = append(prodEntries, e)
		} else {
			resEntries = append(resEntries, e)
		}
	}

	// Draw road grid
	roadCol := lerp(pal.Road, c(150, 145, 135), 0.4)
	maxR := int(float64(min(w, h)) * 0.32)
	// Horizontal roads every 3 cells
	for ry := cy - maxR; ry <= cy+maxR; ry += cellSpacing * 3 {
		drawRoadLine(img, cx-maxR, ry, cx+maxR, ry, w, h, 1, roadCol)
	}
	// Vertical roads every 4 cells
	for rx := cx - maxR; rx <= cx+maxR; rx += cellSpacing * 4 {
		drawRoadLine(img, rx, cy-maxR, rx, cy+maxR, w, h, 1, roadCol)
	}

	// Place production on left half
	placeGrid := func(ents []layoutEntry, startX, startY, dirX int) {
		col := 0
		row := 0
		for _, e := range ents {
			for i := 0; i < e.count; i++ {
				size := e.size
				bx := startX + col*cellSpacing*dirX
				by := startY + row*cellSpacing
				bx = clampInt(bx, border, w-border)
				by = clampInt(by, border, h-border)

				placed := false
				for attempt := 0; attempt < 20; attempt++ {
					if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, size, size) {
						placed = true
						break
					}
					bx = clampInt(bx+rng.Intn(5)-2, border, w-border)
					by = clampInt(by+rng.Intn(5)-2, border, h-border)
				}
				if placed {
					grid.claim(bx, by, size, size)
					placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, size})
				}
				col++
				if col > 4 {
					col = 0
					row++
				}
			}
		}
	}

	startY := cy - int(float64(min(w, h))*0.3)
	placeGrid(prodEntries, cx-cellSpacing, startY, -1)
	placeGrid(resEntries, cx+cellSpacing, startY, 1)

	return placements
}

// ─── Modern layout (modern/atomic) ───────────────────────
func placeBuildingsModern(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(12)
	border := 8
	blockW := 6
	blockH := 4
	cellSize := 10
	streetW := 3

	blockTotalW := blockW*cellSize + streetW
	blockTotalH := blockH*cellSize + streetW

	// Start from top-left of city area
	cityR := int(float64(min(w, h)) * 0.32)
	startX := cx - cityR + border
	startY := cy - cityR + border
	streetColor := lerp(pal.Road, c(80, 80, 85), 0.5)

	bldIdx := 0
	// Flatten all buildings into a single list
	var allBlds []layoutEntry
	for _, e := range entries {
		for i := 0; i < e.count; i++ {
			allBlds = append(allBlds, e)
		}
	}

	blockRow := 0
	blockCol := 0
	for bldIdx < len(allBlds) {
		blockX := startX + blockCol*blockTotalW
		blockY := startY + blockRow*blockTotalH

		if blockX+blockTotalW > cx+cityR || blockY+blockTotalH > cy+cityR {
			break
		}

		// Draw streets around this block
		drawRoadLine(img, blockX-streetW, blockY, blockX+blockTotalW, blockY, w, h, streetW, streetColor)
		drawRoadLine(img, blockX, blockY-streetW, blockX, blockY+blockTotalH, w, h, streetW, streetColor)

		// Fill block with buildings
		for row := 0; row < blockH && bldIdx < len(allBlds); row++ {
			for col := 0; col < blockW && bldIdx < len(allBlds); col++ {
				e := allBlds[bldIdx]
				bldIdx++
				bx := blockX + col*cellSize + cellSize/2
				by := blockY + row*cellSize + cellSize/2
				bx = clampInt(bx, border, w-border)
				by = clampInt(by, border, h-border)
				if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, e.size, e.size) {
					grid.claim(bx, by, e.size, e.size)
					placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, e.size})
				}
			}
		}

		blockCol++
		if blockX+blockTotalW*2 > cx+cityR {
			blockCol = 0
			blockRow++
		}
	}
	_ = rng // suppress unused
	return placements
}

// ─── Campus layout (digital/nano) ────────────────────────
func placeBuildingsCampus(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(8)
	border := 8

	// Flatten all buildings
	var allBlds []layoutEntry
	for _, e := range entries {
		for i := 0; i < e.count; i++ {
			allBlds = append(allBlds, e)
		}
	}

	if len(allBlds) == 0 {
		return placements
	}

	// Arrange campus cluster centers in hexagonal pattern
	clusterSize := 8 // buildings per campus
	numClusters := (len(allBlds) + clusterSize - 1) / clusterSize
	if numClusters < 1 {
		numClusters = 1
	}

	clusterR := int(float64(min(w, h)) * 0.32)
	pathColor := lerp(pal.Road, pal.Accent1, 0.3)

	type clusterCenter struct{ x, y int }
	var clusterCenters []clusterCenter

	for ci := 0; ci < numClusters; ci++ {
		var ccx, ccy int
		if ci == 0 {
			ccx, ccy = cx, cy
		} else {
			// Hexagonal arrangement
			ring := 1
			pos := ci - 1
			hexDir := [][2]float64{
				{1, 0}, {0.5, 0.866}, {-0.5, 0.866},
				{-1, 0}, {-0.5, -0.866}, {0.5, -0.866},
			}
			hexIdx := pos % 6
			ang := math.Atan2(hexDir[hexIdx][1], hexDir[hexIdx][0])
			_ = ring
			ccx = cx + int(math.Cos(ang)*float64(clusterR))
			ccy = cy + int(math.Sin(ang)*float64(clusterR)*0.70)
		}
		ccx = clampInt(ccx, border+20, w-border-20)
		ccy = clampInt(ccy, border+20, h-border-20)
		clusterCenters = append(clusterCenters, clusterCenter{ccx, ccy})
	}

	// Draw winding paths between cluster centers
	for i := 1; i < len(clusterCenters); i++ {
		a, b := clusterCenters[i-1], clusterCenters[i]
		drawRoadLine(img, a.x, a.y, b.x, b.y, w, h, 1, pathColor)
	}

	// Place buildings in each cluster
	bldIdx := 0
	for ci, cc := range clusterCenters {
		_ = ci
		clusterCellSize := 8
		for i := 0; i < clusterSize && bldIdx < len(allBlds); i++ {
			e := allBlds[bldIdx]
			bldIdx++
			ang := float64(i) / float64(clusterSize) * 2.0 * math.Pi
			dist := float64(5 + i%3*clusterCellSize)
			bx := cc.x + int(math.Cos(ang)*dist)
			by := cc.y + int(math.Sin(ang)*dist*0.75)
			bx = clampInt(bx, border, w-border)
			by = clampInt(by, border, h-border)

			placed := false
			for attempt := 0; attempt < 20; attempt++ {
				if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, e.size, e.size) {
					placed = true
					break
				}
				bx = clampInt(bx+rng.Intn(9)-4, border, w-border)
				by = clampInt(by+rng.Intn(7)-3, border, h-border)
			}
			if placed {
				grid.claim(bx, by, e.size, e.size)
				placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, e.size})
			}
		}
	}
	return placements
}

// ─── Orbital layout (space/galactic) ─────────────────────
func placeBuildingsOrbital(img *image.RGBA, cfg MapGenConfig, rng *rand.Rand, cx, cy, w, h int, entries []layoutEntry) []bldInfo {
	var placements []bldInfo
	pal := getTerrainPalette(eraFromAge(cfg.AgeKey))
	grid := newPlotGrid(10)
	border := 8
	_ = rng

	// Flatten buildings
	var allBlds []layoutEntry
	for _, e := range entries {
		for i := 0; i < e.count; i++ {
			allBlds = append(allBlds, e)
		}
	}

	// Central hub: most-assigned building
	if len(allBlds) > 0 {
		hub := allBlds[0]
		grid.claim(cx, cy, hub.size, hub.size)
		placements = append(placements, bldInfo{hub.key, hub.category, hub.domain, cx, cy, hub.size})
		allBlds = allBlds[1:]
	}

	// 3 orbital rings
	ringRadii := []int{25, 45, 70}
	orbitColors := []color.RGBA{
		lerp(pal.Accent1, pal.Ground, 0.5),
		lerp(pal.Accent2, pal.Ground, 0.5),
		lerp(pal.Accent1, pal.Accent2, 0.5),
	}

	// Draw dotted rings
	for ri, r := range ringRadii {
		ringC := orbitColors[ri]
		circumference := int(2.0 * math.Pi * float64(r))
		for i := 0; i < circumference; i += 3 {
			ang := float64(i) / float64(circumference) * 2.0 * math.Pi
			rx := cx + int(math.Cos(ang)*float64(r))
			ry := cy + int(math.Sin(ang)*float64(r)*0.70)
			if rx >= 0 && rx < w && ry >= 0 && ry < h {
				existing := img.RGBAAt(rx, ry)
				img.SetRGBA(rx, ry, lerp(existing, ringC, 0.6))
			}
		}
	}

	// Distribute remaining buildings across rings
	bldIdx := 0
	for ri, r := range ringRadii {
		if bldIdx >= len(allBlds) {
			break
		}
		// How many buildings fit in this ring?
		ringCount := len(allBlds) / (len(ringRadii) - ri)
		if ri == len(ringRadii)-1 {
			ringCount = len(allBlds) - bldIdx
		}
		if ringCount < 1 {
			ringCount = 1
		}

		for i := 0; i < ringCount && bldIdx < len(allBlds); i++ {
			e := allBlds[bldIdx]
			bldIdx++
			total := ringCount
			if total < 1 {
				total = 1
			}
			ang := float64(i) / float64(total) * 2.0 * math.Pi
			bx := cx + int(math.Cos(ang)*float64(r))
			by := cy + int(math.Sin(ang)*float64(r)*0.70)
			bx = clampInt(bx, border, w-border)
			by = clampInt(by, border, h-border)

			placed := false
			for attempt := 0; attempt < 20; attempt++ {
				if !isWaterPixel(img, bx, by, pal) && grid.isFree(bx, by, e.size, e.size) {
					placed = true
					break
				}
				// Try slight angular offset
				offAng := ang + float64(attempt)*0.2
				bx = cx + int(math.Cos(offAng)*float64(r))
				by = cy + int(math.Sin(offAng)*float64(r)*0.70)
				bx = clampInt(bx, border, w-border)
				by = clampInt(by, border, h-border)
			}
			if placed {
				grid.claim(bx, by, e.size, e.size)
				placements = append(placements, bldInfo{e.key, e.category, e.domain, bx, by, e.size})
			}
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
func drawBuildings(img *image.RGBA, w, h, era, dl int, pal TerrainPalette, ageKey string, placements []bldInfo) {
	eraName := getEraName(ageKey)
	for _, b := range placements {
		bx, by, size := b.x, b.y, b.size

		// Determine scale: wonders get scale 2, full-map gets 1, minimap gets 1
		scale := 1
		if dl > 0 {
			if b.category == "wonder" {
				scale = 2
			} else {
				scale = 1
			}
		}

		// Look up colors from the existing visual map (keeps all 284 mappings)
		vis := getBuildingVisual(b.key, b.category)
		primary := vis.Primary
		accent := vis.Accent

		// Determine sprite type from domain + era
		stype := getBuildingSprite(b.domain, b.key, eraName)
		if b.category == "wonder" {
			stype = spriteWonder
		}

		// Shadow under the sprite.
		// At scale<=1 drawBuildingSprite renders a plain 3×3 block, so the
		// shadow must also be 3×3 — using the full sprite dimensions would
		// leave dark artifact pixels visible around the tiny solid block.
		shadowOff := 1 + dl
		var sprW, sprH int
		if scale <= 1 {
			sprW, sprH = 3, 3
		} else {
			sprW = len([]rune(spriteRowWidth(stype))) * scale
			sprH = spriteRowCount(stype) * scale
		}
		for dy := 0; dy < sprH; dy++ {
			for dx := 0; dx < sprW; dx++ {
				px := bx - sprW/2 + dx + shadowOff
				py := by - sprH/2 + dy + shadowOff
				if px >= 0 && px < w && py >= 0 && py < h {
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, c(0, 0, 0), 0.30))
				}
			}
		}

		// Draw the sprite centred on (bx, by)
		drawBuildingSprite(img, w, h,
			bx-sprW/2, by-sprH/2,
			stype, primary, accent, scale)

		// Wonder glow corona (retained from old system)
		if b.category == "wonder" {
			glowR := size + 7 + dl*4
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
					img.SetRGBA(px, py, lerp(existing, gc, fade*0.55))
				}
			}
			innerR := size/2 + 2 + dl
			for dy := -innerR; dy <= innerR; dy++ {
				for dx := -innerR; dx <= innerR; dx++ {
					d := math.Sqrt(float64(dx*dx + dy*dy))
					if d < float64(size/2+1) || d > float64(innerR) {
						continue
					}
					px, py := bx+dx, by+dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					ringFade := 1.0 - math.Abs(d-float64(size/2+2))/float64(innerR-size/2)
					gc := pal.Accent2
					if era >= 6 {
						gc = pal.Accent1
					}
					existing := img.RGBAAt(px, py)
					img.SetRGBA(px, py, lerp(existing, gc, ringFade*0.75))
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
	peakH := r + r/3  // height of peak region
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

func settlementLabel(buildingCount int, ageKey string) string {
	size := ""
	switch {
	case buildingCount <= 5:
		size = "Village"
	case buildingCount <= 20:
		size = "Town"
	case buildingCount <= 50:
		size = "City"
	default:
		size = "Metropolis"
	}
	eraName := getEraName(ageKey)
	if eraName == "" {
		return size
	}
	// Capitalise the era name for display
	if len(eraName) > 0 {
		eraName = string(eraName[0]-32) + eraName[1:]
	}
	return eraName + " " + size
}
