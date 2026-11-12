package window

import (
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/mapview"
	"github.com/espresso20/ageforge/world"
)

// At the Galactic Age a click on Andromeda, made the way the window makes it
// (a pixel on screen, through the art scale and the canvas the last draw
// left), selects it and reports the lock that names the Quantum Age.
func TestGalacticClickOnAndromedaReportsTheLock(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "galactic_age", Seed: 7, Tick: 43200})
	w := world.Build(mapmodel.NewBuilder(mapmodel.NewCatalog()).Build(&st, nil), 7)
	pal := world.Palette{Bg: color.RGBA{R: 4, G: 4, B: 12, A: 255}}
	for i := range pal.Tones {
		pal.Tones[i] = color.RGBA{R: uint8(40 + i*12), G: uint8(255 - i*9), B: uint8(90 + i*7), A: 255}
	}
	s := newSky()
	var got mapstyle.MapState
	s.Watch(func(string) string { return "" }, func(m mapstyle.MapState) { got = m })
	s.Show(image.Rect(0, 0, 100, 40), w, 1, pal)

	px := image.Rect(0, 48, 1930, 1000) // as draw lays the map out
	n := mapview.ArtScale(px.Dx())
	s.px, s.n, s.w, s.h = px, n, (px.Dx()+n-1)/n, (px.Dy()+n-1)/n
	for i := 0; i < 40; i++ { // zoom all the way out, as the key does
		s.Command(mapstyle.MapCmd{Op: mapstyle.MapZoomOut})
		s.view.Update(time.Second)
	}
	s.view.Render(s.w, s.h)
	for cy := px.Min.Y; cy < px.Max.Y; cy += n {
		for cx := px.Min.X; cx < px.Max.X; cx += n {
			ax, ay, in := s.artAt(cx, cy)
			if !in {
				t.Fatalf("pixel %d,%d is outside the map", cx, cy)
			}
			if s.click(ax, ay); got.Sel != nil && got.Sel.Key == "andromeda" {
				if got.Lock != world.AgeQuantum {
					t.Fatalf("Andromeda reports lock %q, want %q", got.Lock, world.AgeQuantum)
				}
				return
			}
		}
	}
	t.Fatal("no click on the screen selected Andromeda")
}
