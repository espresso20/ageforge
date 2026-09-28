package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

// interactive runs the organism full screen, as a tview primitive the way
// the game's panels are.
//
//	← →  / tab   inspect the previous / next limb     esc  stop inspecting
//	[ ]          previous / next age's save          v    since-last-visit marks
//	t            next theme                          space  pause the wind
//	q            quit
func interactive(states, age, themeKey string, screen tcell.Screen) error {
	if err := theme.SetActive(themeKey); err != nil {
		return err
	}
	ages := []string{"primitive_seedling"}
	ages = append(ages, config.AgeOrder()...)
	ai := 0
	for i, a := range ages {
		if a == age {
			ai = i
		}
	}
	o, prev, err := loadPair(states, ages[ai])
	if err != nil {
		return err
	}
	themes := theme.All()
	ti := 0
	for i, t := range themes {
		if t.Key == themeKey {
			ti = i
		}
	}
	v := View{Sel: -1, Since: true}
	start := time.Now()
	paused := false
	var frozen float64

	app := tview.NewApplication()
	if screen != nil {
		app.SetScreen(screen) // tests drive a SimulationScreen
	}
	box := tview.NewBox()
	box.SetDrawFunc(func(s tcell.Screen, x, y, w, h int) (int, int, int, int) {
		if paused {
			v.T = frozen
		} else {
			v.T = 1 + time.Since(start).Seconds()
		}
		Render(o, prev, v, w, h).Blit(s, x, y)
		return x, y, w, h
	})
	box.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		n := len(o.Limbs)
		switch {
		case ev.Key() == tcell.KeyRight || ev.Key() == tcell.KeyTab:
			if n > 0 {
				v.Sel = (v.Sel + 1) % n
			}
		case ev.Key() == tcell.KeyLeft || ev.Key() == tcell.KeyBacktab:
			if n > 0 {
				if v.Sel <= 0 {
					v.Sel = n - 1
				} else {
					v.Sel--
				}
			}
		case ev.Key() == tcell.KeyEsc:
			v.Sel = -1
		case ev.Rune() == '[' || ev.Rune() == ']':
			d := 1
			if ev.Rune() == '[' {
				d = -1
			}
			for k := 0; k < len(ages); k++ {
				ai = (ai + d + len(ages)) % len(ages)
				if no, np, err := loadPair(states, ages[ai]); err == nil {
					o, prev = no, np
					v.Sel = -1
					break
				}
			}
		case ev.Rune() == 'v':
			v.Since = !v.Since
		case ev.Rune() == 't':
			ti = (ti + 1) % len(themes)
			_ = theme.SetActive(themes[ti].Key)
		case ev.Rune() == ' ':
			paused = !paused
			frozen = v.T
		case ev.Rune() == 'q':
			app.Stop()
		}
		return nil
	})
	go func() {
		tk := time.NewTicker(125 * time.Millisecond)
		defer tk.Stop()
		for range tk.C {
			app.Draw()
		}
	}()
	return app.SetRoot(box, true).Run()
}
