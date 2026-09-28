package main

// interactive.go runs the walk in a real terminal: arrow keys step, typed
// lines go through Exec, and a timer advances the frame so smoke rises and
// goods move along the roads. The game would host View as a panel instead.

import (
	"fmt"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

func runInteractive(states, age, since, themeKey string) error {
	if err := theme.SetActive(themeKey); err != nil {
		return err
	}
	st, err := loadState(states, age)
	if err != nil {
		return fmt.Errorf("loading %s: %w (run with -gen first)", age, err)
	}
	var prev *City
	if since != "" {
		ps, err := loadState(states, since)
		if err != nil {
			return err
		}
		prev = BuildCity(ps)
	}
	v := NewView(BuildCity(st), prev)

	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := scr.Init(); err != nil {
		return err
	}
	defer scr.Fini()
	mini := false

	events := make(chan tcell.Event, 16)
	go func() {
		for {
			ev := scr.PollEvent()
			if ev == nil {
				return
			}
			events <- ev
		}
	}()
	tick := time.NewTicker(700 * time.Millisecond)
	defer tick.Stop()
	for {
		if mini {
			v.DrawMini(scr)
		} else {
			v.Draw(scr)
		}
		scr.Show()
		select {
		case <-tick.C:
			v.Frame++
		case ev := <-events:
			switch e := ev.(type) {
			case *tcell.EventResize:
				scr.Sync()
			case *tcell.EventKey:
				switch e.Key() {
				case tcell.KeyCtrlC:
					return nil
				case tcell.KeyTab:
					mini = !mini
				case tcell.KeyEscape:
					if v.Input == "" {
						return nil
					}
					v.Input = ""
				case tcell.KeyEnter:
					line := v.Input
					v.Input = ""
					if line == "" {
						line = "look"
					}
					if !v.Exec(line) {
						return nil
					}
				case tcell.KeyBackspace, tcell.KeyBackspace2:
					if r := []rune(v.Input); len(r) > 0 {
						v.Input = string(r[:len(r)-1])
					}
				case tcell.KeyUp:
					v.Exec(v.City.DirWord(North))
				case tcell.KeyDown:
					v.Exec(v.City.DirWord(South))
				case tcell.KeyLeft:
					v.Exec(v.City.DirWord(West))
				case tcell.KeyRight:
					v.Exec(v.City.DirWord(East))
				case tcell.KeyPgDn:
					v.Scroll += 5
				case tcell.KeyPgUp:
					v.Scroll = max(0, v.Scroll-5)
				case tcell.KeyRune:
					v.Input += string(e.Rune())
				}
			}
		}
	}
}
