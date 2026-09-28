package main

import (
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// Class is a glyph's colour class. The whole map uses these few classes, and
// each one is derived from a theme role, so a theme switch is free and a
// light theme needs no re-keying.
type Class uint8

const (
	CNone   Class = iota
	CGround       // grass stipple, beach
	CFlora        // forest, fields
	CWater
	CRock   // mountains, ruins
	CHill   // hills: quieter than mountains
	CRoad   // streets, trails
	CWall   // walls, towers, gates
	CHouse  // housing
	CWork   // production buildings (the epoch hue)
	CCivic  // knowledge, faith, culture, monuments
	CWealth // trade, storage, wonders
	CMil    // military
	CLife   // workers, caravans, scouts
	CIdle   // idle workers: a bottleneck signal
	CDanger // war, catastrophe, ruins on fire
	CMemory // fogged-but-remembered terrain
	CStar   // cosmic sky
	CLabel  // names on the grid
	numClasses
)

// Dial is everything that changes between epochs. There is one rendering
// grammar; an epoch only turns these dials.
type Dial struct {
	Epoch     string
	Road      int // 0 trodden dots, 1 light box, 2 heavy box, 3 double box
	Wall      int // 0 none, 1 palisade, 2 stone wall, 3 ring boulevard
	House     rune
	Field     rune
	Store     rune
	Centre    rune
	CentreHot bool  // flickers like a fire
	Hue       int32 // 0xRRGGBB tint for production buildings
	Night     bool  // terrain dimmed so lit structures glow (neon)
	Space     bool  // the region view shows the planet in space
	Wonder    [2]string
	Smoke     []rune // puff sequence rising from works
	Pollution float64
}

var dials = map[string]Dial{
	"stone_era": {Epoch: "stone_era", Road: 0, Wall: 1, House: '∩', Field: '"', Store: '□', Centre: '*', CentreHot: true,
		Hue: 0xc08a4a, Wonder: [2]string{" ▲ ", "▲█▲"}, Smoke: []rune{'∙', '°', '˚'}},
	"iron_era": {Epoch: "iron_era", Road: 1, Wall: 2, House: '⌂', Field: '"', Store: '□', Centre: 'Π',
		Hue: 0xa8a29a, Wonder: [2]string{"╓╥╖", "╨╨╨"}, Smoke: []rune{'∙', '°', '˚'}},
	"steel_era": {Epoch: "steel_era", Road: 1, Wall: 2, House: '⌂', Field: '≡', Store: '□', Centre: 'Φ',
		Hue: 0xc0603a, Wonder: [2]string{"┌╬┐", "▐█▌"}, Smoke: []rune{'°', '○', '∙'}, Pollution: 0.2},
	"electric_era": {Epoch: "electric_era", Road: 2, Wall: 3, House: '▪', Field: '≡', Store: '□', Centre: 'Φ',
		Hue: 0xe0b83a, Wonder: [2]string{" ▲ ", "▐▓▌"}, Smoke: []rune{'°', '○', '∙'}, Pollution: 0.35},
	"digital_era": {Epoch: "digital_era", Road: 2, Wall: 3, House: '▓', Field: '≡', Store: '□', Centre: '◘',
		Hue: 0x3ab8d0, Wonder: [2]string{" ║ ", "▐█▌"}, Smoke: []rune{'∙', '·'}},
	"neon_era": {Epoch: "neon_era", Road: 3, Wall: 3, House: '▓', Field: '▤', Store: '□', Centre: '◙',
		Hue: 0xe040c0, Night: true, Wonder: [2]string{" ◊ ", "▀█▀"}, Smoke: []rune{'·', '∙'}},
	"cosmic_era": {Epoch: "cosmic_era", Road: 3, Wall: 3, House: '◘', Field: '▦', Store: '□', Centre: '☼',
		Hue: 0x9a7ae0, Night: true, Space: true, Wonder: [2]string{"(◙)", " ║ "}, Smoke: []rune{'·'}},
}

func dialFor(epoch string) Dial {
	if d, ok := dials[epoch]; ok {
		return d
	}
	return dials["stone_era"]
}

// lineageGlyph is the one glyph a lineage keeps across every age, so a
// player learns the map once: § is always knowledge, $ is always trade.
// Housing, fields and storage re-skin per epoch through the Dial.
func lineageGlyph(lin string, d Dial) (rune, Class) {
	switch lin {
	case "housing":
		return d.House, CHouse
	case "food":
		return d.Field, CFlora
	case "organic_extraction":
		return '♠', CWork
	case "geological_extraction":
		return '▼', CWork
	case "metallurgy":
		return '♦', CWork
	case "engineering":
		return '¤', CWork
	case "energy":
		return '☼', CWork
	case "harbor":
		return '╤', CWork
	case "hacker":
		return 'λ', CWork
	case "astronaut":
		return '↑', CWork
	case "knowledge":
		return '§', CCivic
	case "faith":
		return 'Ω', CCivic
	case "culture_arts":
		return '♫', CCivic
	case "monument":
		return '◊', CCivic
	case "trade":
		return '$', CWealth
	case "storage":
		return d.Store, CWealth
	case "military":
		return '†', CMil
	}
	return '■', CWork
}

var lineageNames = map[string]string{
	"housing": "housing", "food": "farms & food", "organic_extraction": "wood & fibre",
	"geological_extraction": "quarries & mines", "metallurgy": "metalworks", "engineering": "engineering",
	"energy": "power", "harbor": "harbor", "hacker": "hacker dens", "astronaut": "launch & orbit",
	"knowledge": "knowledge", "faith": "faith", "culture_arts": "culture", "monument": "monuments",
	"trade": "trade", "storage": "storage", "military": "military", "wonder": "wonder",
}

// smoky lineages puff when staffed.
func smoky(lin string) bool {
	return lin == "metallurgy" || lin == "energy" || lin == "engineering" || lin == "geological_extraction"
}

// Palette is the resolved colour of each class for one theme and dial.
type Palette struct {
	Bg                 tcell.Color
	Fg                 [numClasses]tcell.Color
	WaterBg            tcell.Color
	FreshBg            tcell.Color
	CursorBg, CursorFg tcell.Color
	Dim                tcell.Color // understaffed buildings
}

func rgb(h int32) tcell.Color { return tcell.NewHexColor(h) }

// NewPalette derives the class colours from theme roles. Each class starts
// from a role and leans toward an identity hue just enough to read as
// water/forest/stone, then is clamped legible against the background.
func NewPalette(d Dial) Palette {
	bg := theme.Color(theme.RoleBackground)
	light := theme.IsLight()
	role := theme.Color
	mix := func(r theme.Role, hue int32, t, min float64) tcell.Color {
		return theme.Legible(theme.Mix(role(r), rgb(hue), t), bg, min)
	}
	p := Palette{Bg: bg}
	p.Fg[CGround] = theme.Legible(theme.Mix(bg, theme.Mix(role(theme.RoleDim), rgb(0x8a9a5a), 0.5), 0.55), bg, 1.6)
	p.Fg[CFlora] = mix(theme.RolePositive, 0x3f9a4f, 0.55, 3.0)
	p.Fg[CWater] = mix(theme.RoleLabel, 0x3a86c8, 0.7, 3.0)
	p.Fg[CRock] = mix(theme.RoleDim, 0x9a8a78, 0.4, 2.8)
	p.Fg[CHill] = theme.Legible(theme.Mix(bg, mix(theme.RoleDim, 0x9a8a5a, 0.5, 2.0), 0.8), bg, 1.9)
	p.Fg[CRoad] = mix(theme.RoleDim, d.Hue, 0.25, 2.6)
	p.Fg[CWall] = mix(theme.RoleText, 0xb0a898, 0.35, 4.0)
	p.Fg[CHouse] = mix(theme.RoleText, d.Hue, 0.30, 4.5)
	p.Fg[CWork] = mix(theme.RoleHighlight, d.Hue, 0.65, 4.0)
	p.Fg[CCivic] = mix(theme.RoleAccent, 0x8a7ae0, 0.25, 4.0)
	p.Fg[CWealth] = mix(theme.RoleHighlight, 0xe0b040, 0.45, 4.0)
	p.Fg[CMil] = mix(theme.RoleNegative, 0xc05050, 0.3, 4.0)
	p.Fg[CLife] = theme.Legible(role(theme.RoleBright), bg, 7)
	p.Fg[CIdle] = mix(theme.RoleWarning, 0xe0a030, 0.3, 4.5)
	p.Fg[CDanger] = mix(theme.RoleNegative, 0xff3a2a, 0.4, 4.5)
	p.Fg[CMemory] = theme.Mix(bg, role(theme.RoleDim), 0.45)
	p.Fg[CStar] = theme.Mix(bg, role(theme.RoleText), 0.55)
	p.Fg[CLabel] = theme.Legible(role(theme.RoleLabel), bg, 4.5)
	p.WaterBg = theme.Mix(bg, rgb(0x2a6ab0), 0.10)
	if light {
		p.WaterBg = theme.Mix(bg, rgb(0x3a86c8), 0.12)
	}
	p.FreshBg = theme.Mix(bg, role(theme.RolePositive), 0.28)
	p.CursorBg = role(theme.RoleSelection)
	p.CursorFg = role(theme.RoleSelectionText)
	p.Dim = theme.Mix(bg, role(theme.RoleDim), 0.75)
	if d.Night {
		// Night: the land recedes so lit structures glow.
		for _, c := range []Class{CGround, CFlora, CRock, CHill, CWater} {
			p.Fg[c] = theme.Mix(p.Fg[c], bg, 0.35)
		}
		p.Fg[CGround] = theme.Legible(p.Fg[CGround], bg, 1.4)
		p.Fg[CRoad] = mix(theme.RoleAccent, d.Hue, 0.5, 3.0)
	}
	return p
}
