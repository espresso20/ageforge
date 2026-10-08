package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
)

// theme_fx.go draws a theme's ambient effect: a little motion in the cells
// of the dashboard that are otherwise empty (theme.Theme.Effect). Source
// has code falling down its empty columns; Glitch has a tear of static now
// and then.
//
// An effect is drawn after the dashboard, over it, and only into cells
// that hold nothing: a blank on the theme's own background, with blanks on
// both sides of it, so text keeps clear air round it. Each frame is a pure
// function of the frame number and the cell (the same hash as the
// hand-drawn badges), so nothing is stored and no random number is drawn.
// The motion setting turns it off: with motion off nothing here draws.

// fxRoot is the dashboard's root: its layout, and over it the active
// theme's effect.
type fxRoot struct {
	*tview.Flex
	// after draws over the layout once it is drawn.
	after func(scr tcell.Screen, x, y, w, h int)
}

func (r *fxRoot) Draw(scr tcell.Screen) {
	r.Flex.Draw(scr)
	if r.after != nil {
		x, y, w, h := r.GetRect()
		r.after(scr, x, y, w, h)
	}
}

// fxMargin is how many blank cells an effect keeps between itself and
// anything drawn, on either side.
const fxMargin = 3

// fxEmpty reports whether the cells of row y from x-fxMargin to x+fxMargin
// are all blank on bg: room for an effect at (x, y).
func fxEmpty(scr tcell.Screen, x, y, x0, x1 int, bg tcell.Color) bool {
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
func drawThemeEffect(scr tcell.Screen, x, y, w, h int, effect string, n int, plain bool) {
	if w <= 0 || h <= 0 {
		return // a window with no room: nothing to draw into
	}
	bg := theme.Color(theme.RoleBackground)
	style := func(c tcell.Color, bold bool) tcell.Style {
		return tcell.StyleDefault.Background(bg).Foreground(theme.Legible(c, bg, 1.6)).Bold(bold)
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
				col := theme.Color(theme.RoleAccent)
				if artHash(n, r, c) < 0.5 {
					col = theme.Color(theme.RoleLabel)
				}
				scr.SetContent(x+c, y+r, pickRune(noise, artHash(n, c, r)), nil, style(col, false))
			}
		}
	}
}
