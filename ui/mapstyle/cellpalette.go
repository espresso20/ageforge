package mapstyle

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/theme"
)

// The engine says what a cell is made of (a hue and a weight); this file turns
// that into a colour of the player's theme, the way the rest of the map gets
// its colours: a theme role leaning toward an identity hue, laid over the
// theme's own ground by the weight and clamped legible where a glyph is drawn.

// hueFrom is where a hue's identity colour comes from; nil leans on nothing.
type hueFrom func() tcell.Color

func fromMap(h theme.MapHue) hueFrom     { return func() tcell.Color { return theme.MapHueColor(h) } }
func fromCity(h theme.CityHue) hueFrom   { return func() tcell.Color { return theme.CityColor(h) } }
func fromSpace(h theme.SpaceHue) hueFrom { return func() tcell.Color { return theme.SpaceColor(h) } }
func fromEpoch(epoch int) hueFrom        { return fromMap(theme.EpochHue(epoch)) }

// recipe makes a hue: start from a theme role, lean toward a colour by mix.
type recipe struct {
	role theme.Role
	from hueFrom
	mix  float64
}

// recipes is every hue's recipe. The land leans on the map's identity hues,
// the city on the city's, the sky on the sky's.
var recipes = [mapgen.NumHues]recipe{
	mapgen.HDeep:     {theme.RoleLabel, fromMap(theme.HueWaterDeep), 0.85},
	mapgen.HWater:    {theme.RoleLabel, fromMap(theme.HueWater), 0.75},
	mapgen.HSand:     {theme.RoleHighlight, fromMap(theme.HueHill), 0.6},
	mapgen.HScrub:    {theme.RoleHighlight, fromMap(theme.HueGrass), 0.65},
	mapgen.HGrass:    {theme.RolePositive, fromMap(theme.HueGrass), 0.55},
	mapgen.HForest:   {theme.RolePositive, fromMap(theme.HueForest), 0.7},
	mapgen.HHill:     {theme.RoleDim, fromMap(theme.HueHill), 0.5},
	mapgen.HPeak:     {theme.RoleDim, fromMap(theme.HueRock), 0.4},
	mapgen.HEarth:    {theme.RoleDim, fromEpoch(0), 0.5},
	mapgen.HPath:     {theme.RoleDim, fromMap(theme.HueStone), 0.5},
	mapgen.HAsphalt:  {theme.RoleDim, fromCity(theme.CityAsphalt), 0.7},
	mapgen.HConcrete: {theme.RoleDim, fromCity(theme.CityConcrete), 0.7},
	mapgen.HLawn:     {theme.RolePositive, fromCity(theme.CityPark), 0.6},
	mapgen.HRubble:   {theme.RoleDim, fromCity(theme.CityLandfill), 0.5},
	mapgen.HCrop:     {theme.RoleHighlight, fromMap(theme.HueGrass), 0.6},
	mapgen.HThatch:   {theme.RoleHighlight, fromEpoch(0), 0.6},
	mapgen.HTimber:   {theme.RoleDim, fromEpoch(0), 0.4},
	mapgen.HStone:    {theme.RoleText, fromMap(theme.HueStone), 0.35},
	mapgen.HBrick:    {theme.RoleDim, fromCity(theme.CityRust), 0.6},
	mapgen.HTile:     {theme.RoleDim, fromCity(theme.CityRoof), 0.7},
	mapgen.HGlass:    {theme.RoleLabel, fromCity(theme.CityGlass), 0.7},
	mapgen.HSteel:    {theme.RoleDim, fromCity(theme.CitySteel), 0.6},
	mapgen.HGold:     {theme.RoleHighlight, fromMap(theme.HueGold), 0.5},
	mapgen.HNeon:     {theme.RoleText, fromCity(theme.CityNeonMagenta), 0.8},
	mapgen.HCyan:     {theme.RoleLabel, fromCity(theme.CityNeonCyan), 0.8},
	mapgen.HPlasma:   {theme.RoleText, fromCity(theme.CityPlasma), 0.6},
	mapgen.HFire:     {theme.RoleWarning, fromSpace(theme.SpaceSunCorona), 0.5},
	mapgen.HSmoke:    {theme.RoleDim, fromCity(theme.CitySteam), 0.3},
	mapgen.HLight:    {theme.RoleHighlight, fromCity(theme.CityScreen), 0.7},
	mapgen.HCool:     {theme.RoleLabel, fromCity(theme.CityScreenCool), 0.6},
	mapgen.HLife:     {theme.RoleBright, nil, 0},
	mapgen.HRed:      {theme.RoleNegative, fromCity(theme.CityTaillight), 0.6},
	mapgen.HWhite:    {theme.RoleBright, nil, 0},
	mapgen.HAmber:    {theme.RoleHighlight, fromSpace(theme.SpaceCityAmber), 0.6},
	mapgen.HVoid:     {theme.RoleBackground, fromSpace(theme.SpaceVoid), 0.85},
	mapgen.HStar:     {theme.RoleText, fromSpace(theme.SpaceStar), 0.5},
	mapgen.HNebula:   {theme.RoleDim, fromSpace(theme.SpaceNebulaViolet), 0.8},
	mapgen.HSun:      {theme.RoleHighlight, fromSpace(theme.SpaceSun), 0.8},
	mapgen.HSea:      {theme.RoleBackground, fromSpace(theme.SpacePlanetOcean), 0.8},
	mapgen.HLand:     {theme.RoleBackground, fromSpace(theme.SpacePlanetLand), 0.8},
	mapgen.HCloud:    {theme.RoleDim, fromSpace(theme.SpacePlanetCloud), 0.6},
	mapgen.HAir:      {theme.RoleLabel, fromSpace(theme.SpaceAtmosphere), 0.7},
	mapgen.HOrbit:    {theme.RoleDim, fromSpace(theme.SpaceOrbitLine), 0.7},
	mapgen.HEngine:   {theme.RoleLabel, fromSpace(theme.SpaceEngine), 0.7},
	mapgen.HPanel:    {theme.RoleLabel, fromSpace(theme.SpaceSolar), 0.8},
	mapgen.HHydro:    {theme.RolePositive, fromSpace(theme.SpaceHydro), 0.6},
	mapgen.HRock:     {theme.RoleDim, fromSpace(theme.SpaceRock), 0.6},
}

// lineRecipes are the hues of the building lines, in the order of
// mapmodel.LineageOrder. A line's hue is the age's wall colour leaning toward
// its recipe, so districts read apart in every age's material.
var lineRecipes = [...]recipe{
	{theme.RoleText, fromEpoch(0), 0.45},                   // housing
	{theme.RolePositive, fromMap(theme.HueGrass), 0.7},     // food
	{theme.RoleDim, fromEpoch(0), 0.8},                     // wood
	{theme.RoleDim, fromMap(theme.HueRock), 0.7},           // mines
	{theme.RoleWarning, fromEpoch(2), 0.7},                 // metal
	{theme.RoleDim, fromCity(theme.CitySteel), 0.8},        // engineering
	{theme.RoleHighlight, fromEpoch(3), 0.8},               // energy
	{theme.RoleLabel, fromMap(theme.HueWater), 0.6},        // harbor
	{theme.RoleText, fromCity(theme.CityNeonMagenta), 0.7}, // hacker
	{theme.RoleAccent, fromMap(theme.HueCivic), 0.6},       // knowledge
	{theme.RoleText, fromMap(theme.HueStone), 0.3},         // faith
	{theme.RoleText, fromCity(theme.CityNeonMagenta), 0.3}, // culture
	{theme.RoleHighlight, fromMap(theme.HueGold), 0.7},     // monuments
	{theme.RoleHighlight, fromMap(theme.HueGold), 0.5},     // trade
	{theme.RoleDim, fromMap(theme.HueHill), 0.6},           // storage
	{theme.RoleNegative, fromMap(theme.HueMilitary), 0.7},  // military
	{theme.RoleLabel, fromMap(theme.HueNeutral), 0.6},      // diplomacy
}

// lineWallMix is how far a line's hue leans from the wall toward its recipe.
const lineWallMix = 0.65

func (r recipe) colour() tcell.Color {
	base := theme.Color(r.role)
	if r.from == nil {
		return base
	}
	return theme.Mix(base, r.from(), r.mix)
}

// paintKey is one cell's colours: what the style cache is keyed by.
type paintKey struct {
	fg, bg mapgen.Paint
	dim    uint8
	blank  bool
}

// cellPalette is the colours of one theme and one age's walls. It is rebuilt when
// either changes.
type cellPalette struct {
	theme string
	wall  mapgen.Hue
	tint  mapgen.Hue
	bg    tcell.Color
	night tcell.Color
	hue   [mapgen.NumHues]tcell.Color
	cache map[paintKey]tcell.Style
}

// cellPaletteFor returns p when it still fits the theme and the grid, else a new
// one.
func cellPaletteFor(p *cellPalette, g *mapgen.Grid) *cellPalette {
	key := theme.Active().Key
	if p != nil && p.theme == key && p.wall == g.Wall && p.tint == g.Tint {
		return p
	}
	p = &cellPalette{theme: key, wall: g.Wall, tint: g.Tint, bg: theme.Color(theme.RoleBackground), cache: map[paintKey]tcell.Style{}}
	for h := range p.hue {
		p.hue[h] = recipes[h].colour()
	}
	wall := p.hue[g.Wall]
	for i, r := range lineRecipes {
		p.hue[mapgen.HLine+mapgen.Hue(i)] = theme.Mix(wall, r.colour(), lineWallMix)
	}
	// The dark falls toward a deep blue-black, whatever the theme: a light
	// theme's night is dark too, with its own text colour as the black.
	black := p.bg
	if theme.IsLight() {
		black = theme.Color(theme.RoleText)
	}
	p.night = theme.Mix(black, theme.MapHueColor(theme.HueWaterDeep), 0.2)
	return p
}

// over is a paint laid over the theme's ground.
func (p *cellPalette) over(pt mapgen.Paint) tcell.Color {
	if pt.Hue == mapgen.HNone || pt.Weight == 0 {
		return p.bg
	}
	return theme.Mix(p.bg, p.hue[pt.Hue], float64(pt.Weight)/255)
}

// Contrast floors for a glyph on its ground: by day, and in the dim of night.
const (
	dayContrast   = 3.0
	nightContrast = 2.2
)

// style is the style of a cell: its paints over the ground, pulled toward the
// night by its dim, the glyph kept legible on the ground it stands on.
func (p *cellPalette) style(c mapgen.Cell) tcell.Style {
	k := paintKey{c.Fg, c.Bg, c.Dim, c.R == ' ' || c.R == 0}
	if st, ok := p.cache[k]; ok {
		return st
	}
	fg, bg := p.over(c.Fg), p.over(c.Bg)
	floor := dayContrast
	if c.Dim > 0 {
		t := float64(c.Dim) / 255
		fg, bg = theme.Mix(fg, p.night, t*0.75), theme.Mix(bg, p.night, t)
		floor = nightContrast
	}
	if !k.blank {
		fg = theme.Legible(fg, bg, floor)
	}
	st := tcell.StyleDefault.Foreground(fg).Background(bg)
	p.cache[k] = st
	return st
}

// colour is the colour of a hue at full weight, for chrome that borrows it
// (the legend's signs).
func (p *cellPalette) colour(h mapgen.Hue) tcell.Color {
	return theme.Legible(p.hue[h], p.bg, dayContrast)
}
