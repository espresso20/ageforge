package main

import (
	"strings"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// runApp is the interactive atlas: a tview Box whose draw function renders
// the scene, as the game's overlays do, plus a slow ticker for the goods on
// the routes, the radar and the omens.
func runApp(lib *library, sc *Scene) error { return runAppOn(lib, sc, nil) }

// runAppOn runs the atlas on scr (a simulation screen in tests), or on the
// terminal when scr is nil.
func runAppOn(lib *library, sc *Scene, scr tcell.Screen) error {
	app := tview.NewApplication()
	if scr != nil {
		app.SetScreen(scr)
	}
	box := tview.NewBox()
	ageIdx := 0
	for i, a := range lib.ages {
		if a == sc.A.St.Age {
			ageIdx = i
		}
	}
	themes := theme.All()
	themeIdx := 0
	for i, t := range themes {
		if t.Key == theme.Active().Key {
			themeIdx = i
		}
	}
	box.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		cv := sc.Draw(w, h)
		for j := 0; j < cv.H; j++ {
			for i := 0; i < cv.W; i++ {
				c := cv.C[j*cv.W+i]
				screen.SetContent(x+i, y+j, c.R, nil, tcell.StyleDefault.Foreground(c.Fg).Background(c.Bg).Attributes(c.Attr))
			}
		}
		return x, y, w, h
	})
	switchAge := func(d int) {
		n := ageIdx + d
		if n < 0 || n >= len(lib.ages) {
			return
		}
		a, err := lib.atlas(lib.ages[n])
		if err != nil {
			return
		}
		ageIdx = n
		legend := sc.Legend
		*sc = *newScene(a)
		sc.Legend = legend
	}
	box.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		sc.Verb = ""
		in := sc.mapRect
		move := func(dx, dy int) {
			sc.CurX += dx
			sc.CurY += dy
			// the map scrolls when the cursor pushes at its edge
			if sc.CurX < 0 || sc.CurX >= in.W {
				sc.View.CX += float64(dx) * sc.View.S * float64(in.W) / 4
				sc.CurX -= dx * in.W / 4
			}
			if sc.CurY < 0 || sc.CurY >= in.H {
				sc.View.CY += float64(dy) * 2 * sc.View.S * float64(in.H) / 4
				sc.CurY -= dy * in.H / 4
			}
		}
		zoom := func(f float64) {
			// keep the point under the cursor fixed
			p := sc.toWorld(sc.CurX, sc.CurY, in)
			sc.View.S = clamp(sc.View.S*f, 0.2, 6)
			x, y := sc.toCell(p, in)
			sc.View.CX += float64(x-sc.CurX) * sc.View.S
			sc.View.CY += float64(y-sc.CurY) * 2 * sc.View.S
		}
		switch ev.Key() {
		case tcell.KeyLeft:
			move(-1, 0)
		case tcell.KeyRight:
			move(1, 0)
		case tcell.KeyUp:
			move(0, -1)
		case tcell.KeyDown:
			move(0, 1)
		case tcell.KeyTab:
			sc.Next(1)
		case tcell.KeyBacktab:
			sc.Next(-1)
		case tcell.KeyEnter:
			for _, l := range sc.inspect(newPal(sc.Plate)) {
				if strings.HasPrefix(l.s, "» ") {
					sc.Verb = strings.TrimPrefix(l.s, "» ")
					break
				}
			}
		case tcell.KeyEscape:
			app.Stop()
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				app.Stop()
			case 'h':
				move(-1, 0)
			case 'l':
				move(1, 0)
			case 'k':
				move(0, -1)
			case 'j':
				move(0, 1)
			case 'H', 'a':
				move(-in.W/4, 0)
			case 'L', 'd':
				move(in.W/4, 0)
			case 'K', 'w':
				move(0, -in.H/4)
			case 'J', 's':
				move(0, in.H/4)
			case '+', '=':
				zoom(0.7)
			case '-', '_':
				zoom(1 / 0.7)
			case '0':
				sc.fitted = false
			case '[', ']':
				// leaf back through the plates this atlas has earned
				n := plateNum(sc.Plate.Key) - 1
				if ev.Rune() == '[' && n > 0 {
					n--
				} else if ev.Rune() == ']' && n+1 < earnedPlates(sc.A.St.Age) {
					n++
				}
				sc.Plate = plates[plateOrder[n]]
				sc.fitted = false
			case '<', ',':
				switchAge(-1)
			case '>', '.':
				switchAge(1)
			case 't':
				themeIdx = (themeIdx + 1) % len(themes)
				_ = theme.SetActive(themes[themeIdx].Key)
				sc.Verb = "theme " + themes[themeIdx].Key
			case 'm':
				sc.Mono = !sc.Mono
			case 'g':
				sc.Legend = !sc.Legend
				sc.fitted = false
			}
		}
		return nil
	})
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(400 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				app.QueueUpdateDraw(func() { sc.Frame++ })
			}
		}
	}()
	err := app.SetRoot(box, true).Run()
	close(stop)
	return err
}
