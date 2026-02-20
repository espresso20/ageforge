package ui

import "github.com/gdamore/tcell/v2"

// Color theme
var (
	ColorBg        = tcell.ColorDefault
	ColorFg        = tcell.ColorWhite
	ColorTitle     = tcell.ColorGold
	ColorAccent    = tcell.ColorDodgerBlue
	ColorSuccess   = tcell.ColorGreen
	ColorWarning   = tcell.ColorYellow
	ColorError     = tcell.ColorRed
	ColorDim       = tcell.ColorGray
	ColorResource  = tcell.ColorTeal
	ColorBuilding  = tcell.ColorOrangeRed
	ColorVillager  = tcell.ColorPurple
	ColorAge       = tcell.ColorGold
)

// AgePalette defines the color theme for an age era
type AgePalette struct {
	Title    tcell.Color
	Accent   tcell.Color
	Resource tcell.Color
	Building tcell.Color
	Dim      tcell.Color
}

// AgePalettes maps age keys to their color palette
var AgePalettes = map[string]AgePalette{
	// Primitive/Stone: earthy greens and browns
	"primitive_age": {tcell.ColorDarkGreen, tcell.ColorOlive, tcell.ColorTeal, tcell.ColorSaddleBrown, tcell.ColorDimGray},
	"stone_age":     {tcell.ColorDarkGreen, tcell.ColorOlive, tcell.ColorTeal, tcell.ColorSienna, tcell.ColorDimGray},
	// Bronze/Iron: warm bronze and metallic
	"bronze_age": {tcell.ColorGold, tcell.ColorDarkGoldenrod, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	"iron_age":   {tcell.ColorSilver, tcell.ColorSteelBlue, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	// Classical: marble white and royal blue
	"classical_age": {tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.ColorCadetBlue, tcell.ColorCoral, tcell.ColorLightGray},
	// Medieval: dark purple and stone
	"medieval_age":    {tcell.ColorDarkMagenta, tcell.ColorMediumPurple, tcell.ColorDarkCyan, tcell.ColorFireBrick, tcell.ColorDimGray},
	"renaissance_age": {tcell.ColorGold, tcell.ColorMediumOrchid, tcell.ColorDarkCyan, tcell.ColorOrangeRed, tcell.ColorGray},
	"colonial_age":    {tcell.ColorNavajoWhite, tcell.ColorBurlyWood, tcell.ColorTeal, tcell.ColorSienna, tcell.ColorGray},
	// Industrial: dark grays and orange
	"industrial_age": {tcell.ColorDarkOrange, tcell.ColorOrange, tcell.ColorDarkSlateGray, tcell.ColorFireBrick, tcell.ColorDarkGray},
	"victorian_age":  {tcell.ColorRosyBrown, tcell.ColorDarkKhaki, tcell.ColorSlateGray, tcell.ColorBrown, tcell.ColorDimGray},
	"electric_age":   {tcell.ColorYellow, tcell.ColorGold, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	// Modern: clean blue and white
	"atomic_age":  {tcell.ColorLimeGreen, tcell.ColorGreen, tcell.ColorDarkCyan, tcell.ColorRed, tcell.ColorDarkGray},
	"modern_age":  {tcell.ColorDodgerBlue, tcell.ColorSteelBlue, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	"information_age": {tcell.ColorDeepSkyBlue, tcell.ColorCornflowerBlue, tcell.ColorMediumAquamarine, tcell.ColorOrangeRed, tcell.ColorLightGray},
	// Digital: blue/cyan tech
	"digital_age":   {tcell.ColorDarkCyan, tcell.ColorDodgerBlue, tcell.ColorMediumAquamarine, tcell.ColorDeepPink, tcell.ColorDarkSlateGray},
	"cyberpunk_age": {tcell.ColorHotPink, tcell.ColorDarkMagenta, tcell.ColorAqua, tcell.ColorLime, tcell.ColorDarkSlateGray},
	// Fusion/Space: blue and white
	"fusion_age":        {tcell.ColorAquaMarine, tcell.ColorDarkCyan, tcell.ColorTurquoise, tcell.ColorOrangeRed, tcell.ColorSlateGray},
	"space_age":         {tcell.ColorSteelBlue, tcell.ColorLightSkyBlue, tcell.ColorLightCyan, tcell.ColorOrangeRed, tcell.ColorSlateGray},
	"interstellar_age":  {tcell.ColorMediumPurple, tcell.ColorSlateBlue, tcell.ColorLightBlue, tcell.ColorGold, tcell.ColorDimGray},
	// Cosmic: deep purple and gold
	"galactic_age":      {tcell.ColorBlueViolet, tcell.ColorMediumPurple, tcell.ColorLavender, tcell.ColorGold, tcell.ColorDimGray},
	"quantum_age":       {tcell.ColorMediumOrchid, tcell.ColorOrchid, tcell.ColorPlum, tcell.ColorGold, tcell.ColorDarkSlateGray},
	"transcendent_age":  {tcell.ColorGold, tcell.ColorWhite, tcell.ColorLightGoldenrodYellow, tcell.ColorGold, tcell.ColorLightGray},
}

// ApplyAgePalette sets global color variables based on the current age
func ApplyAgePalette(ageKey string) {
	p, ok := AgePalettes[ageKey]
	if !ok {
		return
	}
	ColorTitle = p.Title
	ColorAccent = p.Accent
	ColorResource = p.Resource
	ColorBuilding = p.Building
	ColorDim = p.Dim
}

// ASCII art for splash screen
const SplashArt = `
 █████╗  ██████╗ ███████╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔══██╗██╔════╝ ██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
███████║██║  ███╗█████╗  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██╔══██║██║   ██║██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
██║  ██║╚██████╔╝███████╗██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝`

const SplashTagline = "Forge Your Empire Through the Ages"
