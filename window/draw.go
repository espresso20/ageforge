package window

import (
	"image/color"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/espresso20/ageforge/theme"
)

type advKey struct {
	f    *fontFile
	size float64
	r    rune
}

func toRGBA(c tcell.Color, def color.RGBA) color.RGBA {
	if !c.Valid() {
		return def
	}
	r, g, b := c.RGB()
	if r < 0 {
		return def
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func mix(fg, bg color.RGBA, cover float64) color.RGBA {
	m := func(a, b uint8) uint8 { return uint8(math.Round(float64(b) + (float64(a)-float64(b))*cover)) }
	return color.RGBA{m(fg.R, bg.R), m(fg.G, bg.G), m(fg.B, bg.B), 255}
}

// colors resolves a style to the colors to paint with, reverse and dim applied.
func (g *Game) colors(st tcell.Style) (fg, bg color.RGBA, attr tcell.AttrMask) {
	th := theme.Active()
	f, b, a := st.Decompose()
	fg = toRGBA(f, toRGBA(th.Color(theme.RoleText), color.RGBA{255, 255, 255, 255}))
	bg = toRGBA(b, toRGBA(th.Color(theme.RoleBackground), color.RGBA{A: 255}))
	if a&tcell.AttrReverse != 0 {
		fg, bg = bg, fg
	}
	if a&tcell.AttrDim != 0 {
		fg = mix(fg, bg, 0.6)
	}
	return fg, bg, a
}

func (g *Game) advance(f *fontFile, r rune) float64 {
	k := advKey{f, g.size, r}
	if a, ok := g.adv[k]; ok {
		return a
	}
	a := text.Advance(string(r), g.goFace(f, g.size))
	g.adv[k] = a
	return a
}

// drawCell paints one cell, or the two cells of a wide rune, onto the buffer.
func (g *Game) drawCell(x, y int, k cellKey) {
	cw := g.cw
	if k.wide {
		cw *= 2
	}
	px, py := float32(x*g.cw), float32(y*g.ch)
	fg, bg, attr := g.colors(k.st)
	vector.FillRect(g.buf, px, py, float32(cw), float32(g.ch), bg, false)
	if !k.covered && k.r != ' ' && k.r != 0 {
		if sh, ok := procedural(k.r, g.cw, g.ch); ok {
			g.drawShape(sh, px, py, fg, bg)
		} else {
			g.drawText(k, px, py, cw, fg, attr)
		}
	}
	if attr&tcell.AttrUnderline != 0 {
		t := float32(max(1, g.ch/16))
		vector.FillRect(g.buf, px, py+float32(g.ch)-2*t, float32(cw), t, fg, false)
	}
}

func (g *Game) drawShape(sh shape, px, py float32, fg, bg color.RGBA) {
	for _, t := range sh.tints {
		c := fg
		if t.cover < 1 {
			c = mix(fg, bg, t.cover)
		}
		r := t.r
		vector.FillRect(g.buf, px+float32(r.Min.X), py+float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c, false)
	}
	for _, a := range sh.arcs {
		var p vector.Path
		p.Arc(px+float32(a.cx), py+float32(a.cy), float32(a.rad), float32(a.from), float32(a.to), vector.Clockwise)
		o := &vector.DrawPathOptions{AntiAlias: true}
		o.ColorScale.ScaleWithColor(fg)
		vector.StrokePath(g.buf, &p, &vector.StrokeOptions{Width: float32(a.width)}, o)
	}
	for _, d := range sh.diags {
		vector.StrokeLine(g.buf, px+float32(d.x0), py+float32(d.y0), px+float32(d.x1), py+float32(d.y1), float32(d.width), fg, true)
	}
}

func (g *Game) drawText(k cellKey, px, py float32, cw int, fg color.RGBA, attr tcell.AttrMask) {
	g.mu.Lock()
	s := g.set.Font
	g.mu.Unlock()
	f := g.faces.pick(s, attr&tcell.AttrBold != 0, k.r)
	if f == nil {
		if f = g.faces.pick(s, false, '?'); f == nil {
			return
		}
		k.r, k.comb = '?', ""
	}
	face := g.goFace(f, g.size)
	adv := g.advance(f, k.r)
	op := &text.DrawOptions{}
	sx := 1.0
	if adv > float64(cw) && adv > 0 {
		sx = float64(cw) / adv // wider than its cell: squeeze to fit
		adv = float64(cw)
	}
	op.GeoM.Scale(sx, 1)
	op.GeoM.Translate(float64(px)+(float64(cw)-adv)/2, float64(py)+g.asc-face.Metrics().HAscent)
	op.ColorScale.ScaleWithColor(fg)
	if s == FontPixel {
		op.Filter = ebiten.FilterNearest
	}
	text.Draw(g.buf, string(k.r)+k.comb, face, op)
}
