package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// interactive.go: the prototype viewer. Keys:
//
//	← → / h l      scroll (shift or H L: a half screen)   Home End
//	[ ]            previous / next saved state (age)
//	i              inspect: a cursor; ← → move it, the status line names
//	               the building and its `build` key
//	n              time: auto (runs with the tick) / dawn / noon / dusk / night
//	w              weather: auto / clear / cloud / rain / storm
//	t              next theme      6  16-colour mode
//	c              "new since last visit" markers      m  compact mini view
//	space          pause           q  quit

type viewer struct {
	dir      string
	names    []string
	idx      int
	world    *World
	sceneH   int
	v        View
	themes   []theme.Theme
	ti       int
	timeMode int
	paused   bool
	compact  bool
	inspect  bool
}

var timeModes = []float64{-1, 0.26, 0.5, 0.755, 0.96}
var timeNames = []string{"auto", "dawn", "noon", "dusk", "night"}

func runInteractive(dir, age, themeKey string) error {
	vw := &viewer{dir: dir}
	for _, n := range availableStates(dir) {
		if !strings.HasPrefix(n, "prev_") {
			vw.names = append(vw.names, n)
		}
	}
	if len(vw.names) == 0 {
		return fmt.Errorf("no states in %s: run with -gen first", dir)
	}
	for i, n := range vw.names {
		if n == age {
			vw.idx = i
		}
	}
	vw.themes = theme.All()
	for i, t := range vw.themes {
		if t.Key == themeKey {
			vw.ti = i
		}
	}
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := scr.Init(); err != nil {
		return err
	}
	defer scr.Fini()
	vw.v = View{Cursor: -1, Weather: -1, TOD: -1, ShowNew: true, Theme: vw.themes[vw.ti]}
	w, h := scr.Size()
	if err := vw.load(w, h, true); err != nil {
		return err
	}

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
	tick := time.NewTicker(120 * time.Millisecond)
	defer tick.Stop()
	for {
		vw.draw(scr)
		select {
		case ev := <-events:
			switch e := ev.(type) {
			case *tcell.EventResize:
				w, h := scr.Size()
				_ = vw.load(w, h, false)
				scr.Sync()
			case *tcell.EventKey:
				if vw.key(e) {
					return nil
				}
			}
		case <-tick.C:
			if !vw.paused {
				vw.v.Frame++
			}
		}
	}
}

func (vw *viewer) load(w, h int, resetCam bool) error {
	st, err := loadState(vw.dir, vw.names[vw.idx])
	if err != nil {
		return err
	}
	var prev *game.GameState
	if p, err := loadState(vw.dir, "prev_"+vw.names[vw.idx]); err == nil {
		prev = &p
	}
	vw.v.W, vw.v.H = w, h
	vw.sceneH = sceneRows(h)
	vw.world = buildWorld(st, prev, vw.sceneH)
	if resetCam || vw.v.Cam > vw.world.W-w {
		vw.v.Cam = camFor(vw.world, w, "end")
	}
	return nil
}

func (vw *viewer) draw(scr tcell.Screen) {
	v := vw.v
	v.TOD = timeModes[vw.timeMode]
	if v.TOD < 0 {
		// auto: the game's clock, run fast so a day passes in minutes
		v.TOD = todFromTick(vw.world.St.Tick + v.Frame*3)
	}
	var fb *FB
	if vw.compact {
		cw, ch := min(40, v.W), min(15, v.H)
		cv := v
		cv.W, cv.H = cw, ch
		cw2 := buildWorld(vw.world.St, vw.world.Prev, ch)
		fb = render(vw.world, v)
		small := renderCompact(cw2, cv)
		// the mini view as it would sit in a sidebar: top right, boxed
		ox := v.W - cw - 1
		for y := 0; y < ch; y++ {
			for x := 0; x < cw; x++ {
				*fb.at(ox+x, 1+y) = *small.at(x, y)
			}
		}
	} else {
		fb = render(vw.world, v)
	}
	// a one-line hint of the viewer state, right-aligned in the header
	info := fmt.Sprintf(" %s · time %s · %s ", vw.names[vw.idx], timeNames[vw.timeMode], v.Theme.Name)
	x := v.W - len([]rune(info))
	if x > 0 {
		dim := fromTC(v.Theme.Color(theme.RoleDim))
		bg := fb.at(x, 0).Bg
		fb.text(x, 0, info, dim, bg)
	}
	fb.blitTo(scr)
	scr.Show()
}

// key handles one key; true means quit.
func (vw *viewer) key(e *tcell.EventKey) bool {
	step := 4
	big := vw.v.W / 2
	move := func(d int) {
		if vw.inspect {
			vw.v.Cursor += d
			if vw.v.Cursor < 4 {
				vw.v.Cam += vw.v.Cursor - 4
				vw.v.Cursor = 4
			}
			if vw.v.Cursor > vw.v.W-5 {
				vw.v.Cam += vw.v.Cursor - (vw.v.W - 5)
				vw.v.Cursor = vw.v.W - 5
			}
		} else {
			vw.v.Cam += d
		}
		vw.v.Cam = max(0, min(vw.world.W-vw.v.W, vw.v.Cam))
	}
	switch e.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return true
	case tcell.KeyLeft:
		if e.Modifiers()&tcell.ModShift != 0 {
			move(-big)
		} else {
			move(-step)
		}
	case tcell.KeyRight:
		if e.Modifiers()&tcell.ModShift != 0 {
			move(big)
		} else {
			move(step)
		}
	case tcell.KeyHome:
		vw.v.Cam = 0
	case tcell.KeyEnd:
		vw.v.Cam = camFor(vw.world, vw.v.W, "end")
	case tcell.KeyRune:
		switch e.Rune() {
		case 'q':
			return true
		case 'h':
			move(-step)
		case 'l':
			move(step)
		case 'H':
			move(-big)
		case 'L':
			move(big)
		case '[':
			if vw.idx > 0 {
				vw.idx--
				_ = vw.load(vw.v.W, vw.v.H, true)
			}
		case ']':
			if vw.idx < len(vw.names)-1 {
				vw.idx++
				_ = vw.load(vw.v.W, vw.v.H, true)
			}
		case 'i':
			vw.inspect = !vw.inspect
			if vw.inspect {
				vw.v.Cursor = vw.v.W / 2
			} else {
				vw.v.Cursor = -1
			}
		case 'n':
			vw.timeMode = (vw.timeMode + 1) % len(timeModes)
		case 'w':
			vw.v.Weather++
			if vw.v.Weather > 3 {
				vw.v.Weather = -1
			}
		case 't':
			vw.ti = (vw.ti + 1) % len(vw.themes)
			vw.v.Theme = vw.themes[vw.ti]
		case '6':
			if vw.v.Colors == 16 {
				vw.v.Colors = 0
			} else {
				vw.v.Colors = 16
			}
		case 'c':
			vw.v.ShowNew = !vw.v.ShowNew
		case 'm':
			vw.compact = !vw.compact
		case ' ':
			vw.paused = !vw.paused
		}
	}
	return false
}
