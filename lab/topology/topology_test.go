package main

import (
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// The interactive loop: select, inspect, stage a command, zoom out, quit.
func TestInteractive(t *testing.T) {
	loadAll(t)
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatal(err)
	}
	scr.SetSize(160, 48)
	names := listFixtures("fixtures")
	ai := 0
	for i, n := range names {
		if n == "medieval_age" {
			ai = i
		}
	}
	go func() {
		for _, k := range []tcell.Key{tcell.KeyDown, tcell.KeyDown, tcell.KeyEnter} {
			scr.InjectKey(k, 0, tcell.ModNone)
		}
		scr.InjectKey(tcell.KeyRune, '1', tcell.ModNone)
		scr.InjectKey(tcell.KeyRune, 'w', tcell.ModNone)
		scr.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	}()
	var last *view
	sawInspect, staged := false, ""
	err := loop(scr, "fixtures", names, ai, theme.All(), 0, func(v *view) {
		last = v
		if v.inspect && v.sel != "" {
			sawInspect = true
		}
		if v.staged != "" {
			staged = v.staged
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !sawInspect || staged == "" || !last.world {
		t.Fatalf("inspect=%v staged=%q world=%v", sawInspect, staged, last.world)
	}
	t.Logf("staged %q on %s", staged, last.sel)
}

func loadAll(t testing.TB) []*Fixture {
	ms, _ := filepath.Glob("fixtures/*.json.gz")
	if len(ms) == 0 {
		t.Skip("no fixtures (run -gen)")
	}
	var out []*Fixture
	for _, m := range ms {
		fx, err := loadFixture(m)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, fx)
	}
	return out
}

// Every fixture renders at every size without panicking, inside the frame,
// and the same state always lays out the same way.
func TestRenderSizes(t *testing.T) {
	for _, fx := range loadAll(t) {
		for _, sz := range [][2]int{{200, 60}, {160, 48}, {120, 36}, {100, 30}, {80, 24}, {40, 15}, {30, 10}} {
			for _, world := range []bool{false, true} {
				a := newView(fx, sz[0], sz[1])
				a.world = world
				a.draw()
				b := newView(fx, sz[0], sz[1])
				b.world = world
				b.draw()
				for i := range a.c.cells {
					if a.c.cells[i] != b.c.cells[i] {
						t.Fatalf("%s %v: render not deterministic at cell %d", fx.Age, sz, i)
					}
				}
			}
		}
	}
}

// BenchmarkFrame is one refresh at 160×48: rebuild the model, lay it out,
// route the wires and draw a frame (the production widget would cache the
// layout between state changes and only redraw).
func BenchmarkFrame(b *testing.B) {
	fxs := loadAll(b)
	fx := fxs[len(fxs)-1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := newView(fx, 160, 48)
		v.frame = i
		v.draw()
	}
}

// BenchmarkAnimate is an animation frame on a cached layout.
func BenchmarkAnimate(b *testing.B) {
	fxs := loadAll(b)
	v := newView(fxs[len(fxs)-1], 160, 48)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.frame = i
		v.draw()
	}
}
