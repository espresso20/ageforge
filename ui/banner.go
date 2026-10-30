package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// Public types and functions for banner generation.

// MCell is a cell in a menu grid.
type MCell struct {
	R     rune
	Fg    tcell.Color
	Bg    tcell.Color
	Bold  bool
	HasBg bool
}

// MGrid is a grid for menu rendering.
type MGrid = mGrid

// MenuPalette is the menu's color palette.
type MenuPalette struct {
	Key       string
	Bg        tcell.Color
	Ink       tcell.Color
	Dim       tcell.Color
	Faint     tcell.Color
	Accent    tcell.Color
	Band      tcell.Color
	OnBand    tcell.Color
	Ground    tcell.Color
	GroundInk tcell.Color
	GroundDim tcell.Color
	Label     tcell.Color
	Anvil     tcell.Color
	AnvilInk  tcell.Color
	Heat      [menuLUT]tcell.Color
	Fire      [menuLUT]tcell.Color
	HeatStops [6]tcell.Color
	FireStops [6]tcell.Color
}

// MenuStar is a star in the menu sky.
type MenuStar struct {
	X     int
	Y     int
	Phase float64
	R     rune
}

// MenuRand is a seeded random generator.
type MenuRand = menuRand

// MenuLayout defines the layout of a menu page.
type MenuLayout = menuLayout

// NewMGrid creates a new menu grid.
func NewMGrid(w, h int) *MGrid {
	return newMGrid(w, h)
}

// NewMenuPalette creates a menu palette for a theme.
func NewMenuPalette(th theme.Theme) *MenuPalette {
	p := newMenuPalette(th)
	// Convert private menuPalette to public MenuPalette
	return &MenuPalette{
		Key:       p.key,
		Bg:        p.bg,
		Ink:       p.ink,
		Dim:       p.dim,
		Faint:     p.faint,
		Accent:    p.accent,
		Band:      p.band,
		OnBand:    p.onBand,
		Ground:    p.ground,
		GroundInk: p.groundInk,
		GroundDim: p.groundDim,
		Label:     p.label,
		Anvil:     p.anvil,
		AnvilInk:  p.anvilInk,
		Heat:      p.heat,
		Fire:      p.fire,
		HeatStops: p.heatStops,
		FireStops: p.fireStops,
	}
}

// DrawWordmark draws the wordmark with heat gradient.
func DrawWordmark(g *MGrid, pal *MenuPalette, x0, y0, sx int, t, strike, uc float64, plain bool) {
	// Convert public MenuPalette back to private menuPalette
	privatePal := &menuPalette{
		key:       pal.Key,
		bg:        pal.Bg,
		ink:       pal.Ink,
		dim:       pal.Dim,
		faint:     pal.Faint,
		accent:    pal.Accent,
		band:      pal.Band,
		onBand:    pal.OnBand,
		ground:    pal.Ground,
		groundInk: pal.GroundInk,
		groundDim: pal.GroundDim,
		label:     pal.Label,
		anvil:     pal.Anvil,
		anvilInk:  pal.AnvilInk,
		heat:      pal.Heat,
		fire:      pal.Fire,
		heatStops: pal.HeatStops,
		fireStops: pal.FireStops,
	}
	drawWordmark(g, privatePal, x0, y0, sx, t, strike, uc, plain)
}

// DrawMenuStars draws stars in the menu sky.
func DrawMenuStars(g *MGrid, pal *MenuPalette, stars []MenuStar, t float64) {
	// Convert public MenuStar to private menuStar
	var privateStars []menuStar
	for _, s := range stars {
		privateStars = append(privateStars, menuStar{
			x:     s.X,
			y:     s.Y,
			phase: s.Phase,
			r:     s.R,
		})
	}
	// Convert public MenuPalette back to private menuPalette
	privatePal := &menuPalette{
		key:       pal.Key,
		bg:        pal.Bg,
		ink:       pal.Ink,
		dim:       pal.Dim,
		faint:     pal.Faint,
		accent:    pal.Accent,
		band:      pal.Band,
		onBand:    pal.OnBand,
		ground:    pal.Ground,
		groundInk: pal.GroundInk,
		groundDim: pal.GroundDim,
		label:     pal.Label,
		anvil:     pal.Anvil,
		anvilInk:  pal.AnvilInk,
		heat:      pal.Heat,
		fire:      pal.Fire,
		heatStops: pal.HeatStops,
		fireStops: pal.FireStops,
	}
	drawMenuStars(g, privatePal, privateStars, t)
}

// GetCell gets a cell from the grid (for reading).
func (g *MGrid) GetCell(x, y int) *MCell {
	if g.in(x, y) {
		c := g.c[y*g.w+x]
		return &MCell{
			R:     c.r,
			Fg:    c.fg,
			Bg:    c.bg,
			Bold:  c.bold,
			HasBg: c.hasBg,
		}
	}
	return nil
}

// Text writes text to the grid.
func (g *MGrid) Text(x, y int, s string, fg tcell.Color, bold bool) int {
	return g.text(x, y, s, fg, bold)
}

// GetBg returns the grid's background color (from the palette).
func (g *MGrid) GetBg() tcell.Color {
	return tcell.ColorDefault
}

// Properties of the wordmark.
const (
	MenuWordW = menuWordW
	MenuWordH = menuWordH
)

// NewMenuRand creates a new menu random generator.
func NewMenuRand(seed uint64) *MenuRand {
	return &MenuRand{s: seed}
}

// F generates a random float from 0 to 1.
func (r *MenuRand) F() float64 {
	return r.f()
}

// Next generates the next random value.
func (r *MenuRand) Next() uint64 {
	return r.next()
}
