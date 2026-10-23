package mapstyle

import (
	"slices"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// Palette resolves the model's colour classes against the active theme for
// one epoch. Build one per frame (it is a few dozen colour mixes) or cache
// it by (theme key, epoch).
type Palette struct {
	Bg      tcell.Color
	Fg      [mapmodel.NumClasses]tcell.Color
	Light   bool
	WaterBg tcell.Color // a faint water tint under sea cells
	FreshBg tcell.Color // under tiles new since the last visit
	FlowBg  tcell.Color // under tiles the flows overlay flags
	// Cursor is the inspect cursor: a highlighted cell (see cursor.go).
	// CursorAlt is the other half of its slow blink; draw it through
	// CursorStyle.
	Cursor, CursorAlt tcell.Style
	Chrome            tcell.Style // header and status lines
	Epoch             int
}

// NewPalette derives the class colours from theme roles: each class starts
// from its role, leans toward its identity hue and is clamped legible
// against the background, so light themes darken and dark themes brighten
// on their own.
func NewPalette(epoch int) *Palette {
	bg := theme.Color(theme.RoleBackground)
	p := &Palette{Bg: bg, Light: theme.IsLight(), Epoch: epoch}
	for c := mapmodel.Class(0); c < mapmodel.NumClasses; c++ {
		s := mapmodel.ClassSpecs[c]
		base := theme.Color(s.Role)
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
			col = theme.Mix(bg, theme.Color(theme.RoleDim), 0.5)
		case mapmodel.CStar:
			col = theme.Mix(bg, theme.Color(theme.RoleText), 0.55)
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
	p.FreshBg = theme.Mix(bg, theme.Color(theme.RolePositive), 0.25)
	p.FlowBg = theme.Mix(bg, theme.Color(theme.RoleWarning), 0.22)
	p.FitCursor()
	p.Chrome = tcell.StyleDefault.Foreground(theme.Color(theme.RoleText)).Background(theme.Color(theme.RoleSurface))
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
func (p *Palette) Style(c mapmodel.Class) tcell.Style {
	return tcell.StyleDefault.Foreground(p.Fg[c]).Background(p.Bg)
}

// On is class c on background bg, nudged legible against it.
func (p *Palette) On(c mapmodel.Class, bg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Foreground(theme.Legible(p.Fg[c], bg, 3)).Background(bg)
}

// Role is a theme role on the palette background.
func (p *Palette) Role(r theme.Role) tcell.Style {
	return tcell.StyleDefault.Foreground(theme.Legible(theme.Color(r), p.Bg, 3)).Background(p.Bg)
}
