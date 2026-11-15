package mapstyle

import (
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// Palette resolves the model's colour classes against one theme for one
// epoch: the active theme's for the words and frame round the map
// (NewPalette), Forge's for the picture and its cursor (NewMapPalette).
// Build one per frame (it is a few dozen colour mixes) or cache it by
// (theme key, epoch).
type Palette struct {
	role func(theme.Role) theme.Color

	Bg      theme.Color
	Fg      [mapmodel.NumClasses]theme.Color
	Light   bool
	WaterBg theme.Color // a faint water tint under sea cells
	FreshBg theme.Color // under tiles new since the last visit
	FlowBg  theme.Color // under tiles the flows overlay flags
	// Cursor is the inspect cursor: a highlighted cell (see cursor.go).
	// CursorAlt is the other half of its slow blink; draw it through
	// CursorStyle.
	Cursor, CursorAlt kit.Style
	Chrome            kit.Style // header and status lines
	Epoch             int
}

// NewPalette is the class colours in the active theme, for the chrome: each
// class starts from its role, leans toward its identity hue and is clamped
// legible against the background, so light themes darken and dark themes
// brighten on their own. It has no cursor; the picture's is NewMapPalette's.
func NewPalette(epoch int) *Palette { return newPalette(epoch, theme.Active()) }

// NewMapPalette is the class colours in the picture's theme (Forge, whatever
// theme is active) with the inspect cursor fitted to them.
func NewMapPalette(epoch int) *Palette {
	p := newPalette(epoch, pictureTheme())
	p.FitCursor()
	return p
}

func newPalette(epoch int, th theme.Theme) *Palette {
	role := th.Color
	bg := role(theme.RoleBackground)
	p := &Palette{role: role, Bg: bg, Light: th.IsLight(), Epoch: epoch}
	for c := mapmodel.Class(0); c < mapmodel.NumClasses; c++ {
		s := mapmodel.ClassSpecs[c]
		base := role(s.Role)
		hue := s.Hue
		if s.EpochHue {
			hue = theme.EpochHue(epoch)
		}
		col := base
		if hue != theme.HueNone && s.Mix > 0 {
			col = theme.Mix(base, theme.MapHueColor(hue), s.Mix)
		}
		switch c {
		case mapmodel.CMemory:
			col = theme.Mix(bg, role(theme.RoleDim), 0.5)
		case mapmodel.CStar:
			col = theme.Mix(bg, role(theme.RoleText), 0.55)
		case mapmodel.CGround, mapmodel.CHill:
			col = theme.Mix(bg, col, 0.7)
		case mapmodel.CLegacy:
			col = theme.Mix(bg, col, 0.75)
		}
		if !p.Light && terrainDims(c) {
			col = theme.Mix(bg, col, terrainStep)
		}
		p.Fg[c] = theme.Legible(col, bg, MinContrast(c))
	}
	p.WaterBg = theme.Mix(bg, theme.MapHueColor(theme.HueWaterDeep), 0.10)
	p.FreshBg = theme.Mix(bg, role(theme.RolePositive), 0.25)
	p.FlowBg = theme.Mix(bg, role(theme.RoleWarning), 0.22)
	p.Chrome = kit.Style{}.Foreground(role(theme.RoleText)).Background(role(theme.RoleSurface))
	return p
}

// TerrainFloor is the contrast every map detail colour reads at least at
// against its background (WCAG 3.0:1, the large-text floor), on every theme.
const TerrainFloor = 3.0

// TerrainClasses are the map's ground and detail: the classes held to
// TerrainFloor. On dark themes the ground ones also sit a step quieter than
// the pins, routes and labels.
var TerrainClasses = []mapmodel.Class{mapmodel.CGround, mapmodel.CFlora, mapmodel.CWater,
	mapmodel.CRock, mapmodel.CHill, mapmodel.CRoad, mapmodel.CLegacy, mapmodel.CMemory, mapmodel.CStar}

// MinContrast is class c's floor on the map: its spec's, raised to
// TerrainFloor for the terrain and detail.
func MinContrast(c mapmodel.Class) float64 {
	m := mapmodel.ClassSpecs[c].MinContrast
	if slices.Contains(TerrainClasses, c) {
		m = max(m, TerrainFloor)
	}
	return m
}

// terrainStep is how much of a dark theme's ground colour is kept: one step
// quieter than the pins, routes and labels (the floor still holds).
const terrainStep = 0.8

// terrainDims reports whether class c is ground or detail that sits a step
// quieter on a dark theme: all the terrain but the sky dots.
func terrainDims(c mapmodel.Class) bool {
	return c != mapmodel.CStar && slices.Contains(TerrainClasses, c)
}

// Style is class c on the palette background.
func (p *Palette) Style(c mapmodel.Class) kit.Style {
	return kit.Style{}.Foreground(p.Fg[c]).Background(p.Bg)
}

// On is class c on background bg, nudged legible against it.
func (p *Palette) On(c mapmodel.Class, bg theme.Color) kit.Style {
	return kit.Style{}.Foreground(theme.Legible(p.Fg[c], bg, 3)).Background(bg)
}

// Role is a theme role on the palette background.
func (p *Palette) Role(r theme.Role) kit.Style {
	return kit.Style{}.Foreground(theme.Legible(p.role(r), p.Bg, 3)).Background(p.Bg)
}
