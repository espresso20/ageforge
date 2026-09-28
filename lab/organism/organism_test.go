package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
)

const testStates = "states"

func load(t *testing.T, name string) (*Organism, *Organism) {
	t.Helper()
	o, prev, err := loadPair(testStates, name)
	if err != nil {
		t.Skipf("no saved state %s (run go run ./lab/organism -gen): %v", name, err)
	}
	return o, prev
}

func text(c *Canvas) string {
	var b strings.Builder
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			b.WriteRune(c.cells[y*c.W+x].ch)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Every age's save renders at every size without panicking, and the output
// is exactly the size asked for.
func TestRenderEveryAgeEverySize(t *testing.T) {
	_ = theme.SetActive("forge")
	names := append([]string{"primitive_seedling"}, config.AgeOrder()...)
	sizes := [][2]int{{200, 60}, {160, 48}, {100, 30}, {80, 24}, {40, 15}, {24, 8}, {10, 3}}
	for _, n := range names {
		o, prev := load(t, n)
		for _, sz := range sizes {
			for _, sel := range []int{-1, 0} {
				c := Render(o, prev, View{Sel: sel, Since: true, T: 1.5}, sz[0], sz[1])
				if c.W != sz[0] || c.H != sz[1] {
					t.Fatalf("%s %v: canvas is %dx%d", n, sz, c.W, c.H)
				}
			}
		}
	}
}

// A one-line summary of each saved state, for go test -v.
func TestStatesSummary(t *testing.T) {
	for _, n := range config.AgeOrder() {
		o, _ := load(t, n)
		war := 0
		for _, nb := range o.Neighbors {
			if nb.AtWar {
				war++
			}
		}
		t.Logf("%-17s limbs %2d  built %4d  routes %d  civs %2d  war %d  wonders %d  fossils %d  events %d  harbinger %v",
			n, len(o.Limbs), o.LimbTotal(), len(o.Routes), len(o.Neighbors), war, len(o.Wonders), len(o.Fossils), len(o.Events), o.Harbinger != nil)
	}
}

// The same state renders the same frame: placement is a function of the
// save, not of the process.
func TestDeterministic(t *testing.T) {
	_ = theme.SetActive("forge")
	o, prev := load(t, "industrial_age")
	a := text(Render(o, prev, View{Sel: -1, Since: true}, 160, 48))
	o2, prev2 := load(t, "industrial_age")
	b := text(Render(o2, prev2, View{Sel: -1, Since: true}, 160, 48))
	if a != b {
		t.Fatal("two loads of one save drew different trees")
	}
}

// Growth never reshuffles: a segment revealed for n buildings is laid out
// identically when the limb has more.
func TestGrowthIsPrefixStable(t *testing.T) {
	g := grow(42, 7)
	small := g.layout(reveal(10, len(g.segs)), 0, 0.2, 0.3, 0.4, 0.2, 0.12)
	big := g.layout(reveal(80, len(g.segs)), 0, 0.2, 0.3, 0.4, 0.2, 0.12)
	for i := range small {
		if small[i] != big[i] {
			t.Fatalf("segment %d moved when the limb grew", i)
		}
	}
}

func BenchmarkRender200x60(b *testing.B) {
	_ = theme.SetActive("forge")
	o, prev, err := loadPair(testStates, "transcendent_age")
	if err != nil {
		b.Skip(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Render(o, prev, View{Sel: -1, Since: true, T: float64(i) / 8}, 200, 60)
	}
}

// The interactive view runs on a tview application: tab inspects a limb,
// q quits.
func TestInteractive(t *testing.T) {
	load(t, "medieval_age")
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(120, 36)
	done := make(chan error, 1)
	go func() { done <- interactive(testStates, "medieval_age", "forge", s) }()
	time.Sleep(300 * time.Millisecond)
	s.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	time.Sleep(300 * time.Millisecond)
	cells, w, h := s.GetContents()
	var b strings.Builder
	for i := 0; i < w*h; i++ {
		if len(cells[i].Runes) > 0 {
			b.WriteRune(cells[i].Runes[0])
		}
	}
	if !strings.Contains(b.String(), "build ") {
		t.Error("tab did not open the inspect card")
	}
	s.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("q did not quit")
	}
}
