package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// TestMapCommandsAreRegistryCommands: every command a map can put on its
// inspect line is a whole command to the registry, so Enter would run it
// as typed. It covers the model's command vocabulary for every age, and
// what each style's cursor actually reports as it steps through targets.
func TestMapCommandsAreRegistryCommands(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	c := newCompleter(game.NewGameEngine(), nil)
	b := mapmodel.NewBuilder(nil)
	check := func(where, cmd string) {
		if cmd == "" {
			return
		}
		if !c.complete(cmd) {
			t.Errorf("%s: %q is not a whole registry command", where, cmd)
		}
	}
	for _, age := range config.AgeOrder() {
		for _, o := range []fixture.Options{
			{Age: age, Seed: 7, Wars: 1, Harbinger: true, Idle: 5},
			{Age: age, Seed: 8, Catastrophe: true, Routes: 3},
		} {
			st := fixture.State(o)
			m := b.Build(&st, nil)
			for _, cmd := range m.AllCommands() {
				check(age, cmd)
			}
		}
	}

	reg := all.Registry()
	scr := capture.NewScreen(160, 48)
	defer scr.Fini()
	for _, age := range []string{"primitive_age", "medieval_age", "industrial_age", "cyberpunk_age", "galactic_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 7, Wars: 1, Harbinger: true, Idle: 5})
		m := b.Build(&st, nil)
		f := mapstyle.Frame{Model: m}
		for _, name := range reg.Names() {
			s, _ := reg.New(name)
			s.SetOption(mapstyle.OptInspect, true)
			s.Draw(scr, mapstyle.Rect{W: 160, H: 48}, f)
			seen := 0
			for i := 0; i < 60; i++ {
				if in, ok := s.Inspect(f); ok {
					check(name+"/"+age+" "+in.Title, in.Command)
					if in.Command != "" {
						seen++
					}
				}
				s.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
				s.Draw(scr, mapstyle.Rect{W: 160, H: 48}, f)
			}
			if seen == 0 {
				t.Errorf("%s/%s: the cursor never reported a command", name, age)
			}
		}
	}
}
