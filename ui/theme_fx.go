package ui

import (
	"image"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// theme_fx.go draws a theme's ambient effect: a little motion in the cells
// of the dashboard that are otherwise empty (theme.Theme.Effect). Source
// has code falling down its empty columns; Glitch has a tear of static now
// and then; Ashfall has sparks rising; Ledger has the bars of green-bar
// paper; Prismatic has glints that turn through the spectrum.
//
// An effect is drawn after the dashboard, over it, and only into cells
// that hold nothing: a blank on the theme's own background, with blanks on
// both sides of it, so text keeps clear air round it. Each frame is a pure
// function of the frame number and the cell (a hash, see
// theme_fx_art.go), so nothing is stored and no random number is drawn.
// The motion setting turns it off: with motion off nothing here draws.

// fxRoot is the dashboard's root: its layout, and over it the active
// theme's effect.
type fxRoot struct {
	*nativeFlex
	// after draws over the layout once it is painted.
	after func(scr mapstyle.Cells, x, y, w, h int)
}

// newFxRoot is layout with after drawn over it. The layout's host shows the
// root from here on, so the effect is part of what it paints.
func newFxRoot(layout *nativeFlex, after func(scr mapstyle.Cells, x, y, w, h int)) *fxRoot {
	r := &fxRoot{nativeFlex: layout, after: after}
	return r
}

func (r *fxRoot) Paint(p *kit.Painter) {
	r.nativeFlex.Paint(p)
	if r.after != nil {
		x, y, w, h := r.GetRect()
		r.after(cellsFor(p, r, r.Bounds()), x, y, w, h)
	}
}

// covered: the effect may fall in any cell, so the whole root is copied.
func (r *fxRoot) covered() []image.Rectangle { return []image.Rectangle{r.Bounds()} }

// fxMargin is how many blank cells an effect keeps between itself and
// anything drawn, on either side.
const fxMargin = 3

// fxEmpty reports whether the cells of row y from x-fxMargin to x+fxMargin
// are all blank on bg: room for an effect at (x, y).
func fxEmpty(scr mapstyle.Cells, x, y, x0, x1 int, bg kit.Color) bool {
	for cx := x - fxMargin; cx <= x+fxMargin; cx++ {
		if cx < x0 || cx >= x1 {
			continue
		}
		r, _, st, _ := scr.GetContent(cx, y)
		if _, back, _ := st.Decompose(); r != ' ' || back != bg {
			return false
		}
	}
	return true
}

// drawThemeEffect draws effect into the empty cells of the box at (x, y),
// w by h, at frame n. plain draws it with the plain tier's glyphs.
func drawThemeEffect(scr mapstyle.Cells, x, y, w, h int, effect string, n int, plain bool) {
	if w <= 0 || h <= 0 {
		return // a window with no room: nothing to draw into
	}
	bg := theme.Of(theme.RoleBackground)
	style := func(c kit.Color, bold bool) kit.Style {
		return kit.Style{}.Background(bg).Foreground(theme.Legible(c, bg, 1.6)).Bold(bold)
	}
	// The cells that were empty before this frame drew anything: a glyph
	// drawn into one must not make its neighbour look taken.
	free := make([]bool, w*h)
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			free[r*w+c] = fxEmpty(scr, x+c, y+r, x, x+w, bg)
		}
	}
	light := theme.IsLight()
	switch effect {
	case theme.EffectRain:
		set := rainSet(plain)
		period := h + 14
		for c := 0; c < w; c++ {
			if artHash(c, 7, 7) >= 0.4 {
				continue // most columns stay dry
			}
			speed := 1 + int(artHash(c, 9, 2)*2)
			head := (n*speed/3 + int(artHash(c, 5, 6)*float64(period))) % period
			for d := 0; d < 5; d++ {
				r := head - d
				if r < 0 || r >= h || !free[r*w+c] {
					continue
				}
				scr.SetContent(x+c, y+r, pickRune(set, artHash((n>>1)+d, r, c)), nil,
					style(inkValue(rainInks[min(d+1, len(rainInks)-1)], light), d == 0))
			}
		}
	case theme.EffectGlitch:
		b := artBurst(n, 64, 4, 7)
		if b == 0 {
			return
		}
		noise := []rune("▚▞▀▄░▒")
		if plain {
			noise = []rune("#%=:-")
		}
		for r := 0; r < h; r++ {
			if artHash(n/2, r, 61) >= 0.10*b+0.02 {
				continue
			}
			for c := 0; c < w; c++ {
				if !free[r*w+c] || artHash(n, r*w+c, 63) >= 0.3 {
					continue
				}
				col := theme.Of(theme.RoleAccent)
				if artHash(n, r, c) < 0.5 {
					col = theme.Of(theme.RoleLabel)
				}
				scr.SetContent(x+c, y+r, pickRune(noise, artHash(n, c, r)), nil, style(col, false))
			}
		}
	case theme.EffectEmbers:
		// Sparks rise up some of the columns, a row every three frames,
		// bright and low at first, then small and dim.
		low, high := '·', '˙'
		if plain {
			low, high = '.', '\''
		}
		period := h + 6
		for c := 0; c < w; c++ {
			seed := artHash(c, 77, 3)
			if seed >= 0.3 {
				continue // most columns have none
			}
			for k := 0; k < 2; k++ {
				up := (n/3 + int(seed*400) + k*period/2) % period
				r := h - 1 - up
				if r < 0 || !free[r*w+c] {
					continue
				}
				glyph, col := low, theme.Of(theme.RoleHighlight)
				switch {
				case up > h*2/3:
					glyph, col = high, theme.Of(theme.RoleDim)
				case up > h/3:
					glyph, col = high, theme.Of(theme.RoleLabel)
				}
				scr.SetContent(x+c, y+r, glyph, nil, style(col, up <= h/3))
			}
		}
	case theme.EffectGreenbar:
		// Green-bar paper: every other row of the empty space is tinted. The
		// tint is drawn as ink, so it lies under nothing.
		bar := '█'
		if plain {
			bar = '.'
		}
		tint := kit.Style{}.Background(bg).Foreground(theme.Of(theme.RoleSelection))
		for r := 1; r < h; r += 2 {
			for c := 0; c < w; c++ {
				if free[r*w+c] {
					scr.SetContent(x+c, y+r, bar, nil, tint)
				}
			}
		}
	case theme.EffectPrism:
		// Glints here and there in the empty space, each turning through
		// the spectrum at its own place on the wheel.
		for r := 0; r < h; r++ {
			for c := 0; c < w; c++ {
				seed := artHash(c, r, 29)
				if seed >= 0.035 || !free[r*w+c] {
					continue
				}
				at := n/3 + int(seed*4000)
				glyph := glintRunes[at%4]
				if plain {
					glyph = badgeFold[glyph]
				}
				hue := float64((n*6 + int(seed*36000)) % 360)
				stop := stopPale
				if glyph == '·' || glyph == '.' {
					stop = stopBase
				}
				scr.SetContent(x+c, y+r, glyph, nil, style(inkValue(prismInk(hue, stop), light), false))
			}
		}
	}
}
