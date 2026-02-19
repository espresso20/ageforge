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

// ASCII art for splash screen
const SplashArt = `
 █████╗  ██████╗ ███████╗███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔══██╗██╔════╝ ██╔════╝██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
███████║██║  ███╗█████╗  █████╗  ██║   ██║██████╔╝██║  ███╗█████╗
██╔══██║██║   ██║██╔══╝  ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝
██║  ██║╚██████╔╝███████╗██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝`

const SplashTagline = "Forge Your Empire Through the Ages"
