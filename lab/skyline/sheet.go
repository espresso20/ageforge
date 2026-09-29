package main

import (
	"os"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// sheet renders every archetype (per period) and every wonder on a plain
// sky, for reviewing the grammar in one place: -sheet out.html
func writeSheet(path string, night bool, lo, hi int) error {
	th, _ := theme.ByKey("forge")
	tod := 0.5
	if night {
		tod = 0.96
	}
	p := newPal(Dials{AgeIdx: 10, Epoch: "digital_era", TOD: tod, Theme: th})
	W, H := 200, 150
	fb := newFB(W, H)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			fb.fill(x, y, p.SkyAt(20, 30))
		}
	}
	x, y, rowH := 1, 0, 0
	place := func(s *Sprite, fam int) {
		if x+s.W+2 > W {
			x = 1
			y += rowH + 2
			rowH = 0
		}
		for sy := 0; sy < s.H; sy++ {
			for sx := 0; sx < s.W; sx++ {
				c := s.C[sy*s.W+sx]
				if c.Ch == 0 {
					continue
				}
				fg := p.Col(fam, 0, c.Fg, 0)
				if c.Fg == SWin && night {
					fg = p.Emit(cWinLitA, 0)
				}
				if c.Bg == SNone {
					if c.Ch == '█' {
						fb.set(x+sx, y+sy, '█', fg, fg)
					} else {
						fb.fg(x+sx, y+sy, c.Ch, fg)
					}
				} else {
					fb.set(x+sx, y+sy, c.Ch, fg, p.Col(fam, 0, c.Bg, 0))
				}
			}
		}
		for _, b := range s.Blades {
			fb.fg(x+b.X, y+b.Y, '•', hex(0xffffff))
		}
		x += s.W + 2
		rowH = max(rowH, s.H)
	}
	for a := range config.AgeOrder() {
		if a < lo || a > hi {
			continue
		}
		for _, d := range ageTypes()[a] {
			fm := formFor(d, a)
			if d.Category == "wonder" {
				place(wonderSprite(d.Key, 1), familyOf(a))
				continue
			}
			place(fm.fn(newRnd(len(d.Key), a), int(fm.height)), familyOf(a))
		}
	}
	sim := tcell.NewSimulationScreen("UTF-8")
	_ = sim.Init()
	sim.SetSize(W, min(H, y+rowH+2))
	fb.H = min(H, y+rowH+2)
	fb.blitTo(sim)
	sim.Show()
	return os.WriteFile(path, []byte(wrapHTML("skyline sprite sheet", "#101014", "#ccc", []string{screenHTML(sim)}, 1)), 0o644)
}
