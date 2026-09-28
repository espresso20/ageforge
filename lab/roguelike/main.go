// Command roguelike is the Map Lab "glyph world" prototype: AgeForge's map
// as a living roguelike map of coloured glyphs, drawn from real game states
// produced by the smoke bot.
//
//	go run ./lab/roguelike -age medieval_age          # interactive
//	go run ./lab/roguelike -age galactic_age -theme daylight
//	go run ./lab/roguelike -captures                  # rewrite lab/roguelike/captures
//	go run ./lab/roguelike -gen medieval_age -seed 11 # new state via the smoke bot
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// World size in tiles. The region view fits it to the terminal.
const worldW, worldH = 232, 84

func main() {
	gen := flag.String("gen", "", "comma-separated ages to (re)generate with the smoke bot, e.g. medieval_age,galactic_age")
	seed := flag.Int64("seed", 7, "run seed (picks the state file and the world)")
	linger := flag.Duration("linger", 20*time.Minute, "simulated time to keep playing inside the target age (-gen)")
	dir := flag.String("states", "lab/roguelike/states", "directory of cached MapView snapshots")
	age := flag.String("age", "medieval_age", "age to show")
	themeKey := flag.String("theme", "forge", "theme key (forge, daylight, parchment, cyberpunk, ...)")
	zoom := flag.Int("zoom", ZSettlement, "start zoom: 0 region, 1 settlement, 2 district")
	cata := flag.Bool("catastrophe", false, "force the epoch's catastrophe overlay on")
	captures := flag.Bool("captures", false, "write every capture into lab/roguelike/captures and exit")
	flag.Parse()

	if *gen != "" {
		for _, a := range strings.Split(*gen, ",") {
			t := time.Now()
			v, err := generate(a, *seed, *linger)
			if err != nil {
				fmt.Fprintln(os.Stderr, a, err)
				continue
			}
			if err := saveView(*dir, v, *seed); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			fmt.Printf("%s seed %d: %d building types, pop %d, %d expeditions (%s)\n",
				a, *seed, len(v.Buildings), v.Pop, v.Expeditions, time.Since(t).Round(time.Millisecond))
		}
		return
	}
	if err := theme.SetActive(*themeKey); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *captures {
		if err := writeAllCaptures(*dir, "lab/roguelike/captures"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	mv, err := loadView(*dir, *age, *seed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no state for %s (seed %d): %v\nrun with -gen %s first\n", *age, *seed, err, *age)
		os.Exit(1)
	}
	w := NewWorld(*seed, worldW, worldH)
	v := NewView(Build(w, NewPlan(w), mv, *cata))
	v.Zoom = *zoom
	if err := interactive(v); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func interactive(v *View) error {
	scr, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := scr.Init(); err != nil {
		return err
	}
	defer scr.Fini()
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
	tick := time.NewTicker(280 * time.Millisecond)
	defer tick.Stop()
	draw := func() {
		w, h := scr.Size()
		v.Draw(scr, w, h)
		scr.Show()
	}
	draw()
	for {
		select {
		case <-tick.C:
			v.Frame++
			draw()
		case ev := <-events:
			switch e := ev.(type) {
			case *tcell.EventResize:
				scr.Sync()
			case *tcell.EventKey:
				step := 1
				if e.Modifiers()&tcell.ModShift != 0 {
					step = 8
				}
				if v.Zoom == ZRegion {
					step *= v.scale
				}
				switch e.Key() {
				case tcell.KeyEscape, tcell.KeyCtrlC:
					return nil
				case tcell.KeyUp:
					v.CurY -= step
				case tcell.KeyDown:
					v.CurY += step
				case tcell.KeyLeft:
					v.CurX -= step
				case tcell.KeyRight:
					v.CurX += step
				case tcell.KeyTab:
					v.NextBuilding(1)
				case tcell.KeyBacktab:
					v.NextBuilding(-1)
				case tcell.KeyRune:
					switch e.Rune() {
					case 'q':
						return nil
					case 'k':
						v.CurY -= step
					case 'j':
						v.CurY += step
					case 'h':
						v.CurX -= step
					case 'l':
						v.CurX += step
					case 'K':
						v.CurY -= 8
					case 'J':
						v.CurY += 8
					case 'H':
						v.CurX -= 16
					case 'L':
						v.CurX += 16
					case 'z', '+', '=':
						if v.Zoom < ZDistrict {
							v.Zoom++
						}
					case 'x', '-':
						if v.Zoom > ZRegion {
							v.Zoom--
						}
					case 'c':
						v.CurX, v.CurY = v.S.W.CX, v.S.W.CY
					case '?':
						v.Legend = !v.Legend
					}
				}
				v.CurX = clamp(v.CurX, 0, v.S.W.W-1)
				v.CurY = clamp(v.CurY, 0, v.S.W.H-1)
			}
			draw()
		}
	}
}
