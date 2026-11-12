package window

import (
	"image"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/mapview"
	"github.com/espresso20/ageforge/world"
)

// A sky is the window's own map of space: it draws the world model on a pixel
// canvas under the cells of the Map panel, which are clear where it shows. The
// ui goroutine feeds it (Show, Hide, Command); the draw thread paints it.
type sky struct {
	mu   sync.Mutex
	rect image.Rectangle // the map, in screen cells; empty when hidden
	view *mapview.View
	img  *ebiten.Image
	// where the last draw put the picture: the pixels it covers, the art
	// scale, and the canvas size in art pixels
	px   image.Rectangle
	n    int
	w, h int
	ptr  pointer
}

func newSky() *sky {
	s := &sky{view: mapview.New()}
	s.view.SetFace(newFace(departure))
	return s
}

// SpaceHost is the window's sky, for the Map panel to hand the world to.
func (g *Game) SpaceHost() mapstyle.SpaceHost { return g.sky }

func (s *sky) Show(r image.Rectangle, w *world.World, key uint64, pal world.Palette) {
	s.mu.Lock()
	s.rect = r
	s.mu.Unlock()
	s.view.Set(w, key, pal)
}

// Watch hands the view the words for names and forwards its state to fn.
func (s *sky) Watch(names func(string) string, fn func(mapstyle.MapState)) {
	s.view.SetNames(names)
	s.view.OnChange(func(m mapview.State) {
		fn(mapstyle.MapState{Frame: m.Frame, SelFrame: m.SelFrame, Sel: m.Sel})
	})
}

func (s *sky) Hide() {
	s.mu.Lock()
	s.rect = image.Rectangle{}
	s.mu.Unlock()
	s.view.Clear()
}

// ops maps the ui's map keys to the camera's.
var ops = map[mapstyle.MapOp]mapview.Op{
	mapstyle.MapPanLeft: mapview.PanLeft, mapstyle.MapPanRight: mapview.PanRight, mapstyle.MapPanUp: mapview.PanUp,
	mapstyle.MapPanDown: mapview.PanDown, mapstyle.MapZoomIn: mapview.ZoomIn, mapstyle.MapZoomOut: mapview.ZoomOut,
	mapstyle.MapNext: mapview.Next, mapstyle.MapPrev: mapview.Prev, mapstyle.MapHome: mapview.Home,
}

func (s *sky) Command(c mapstyle.MapCmd) { s.view.Do(mapview.Cmd{Op: ops[c.Op], Fast: c.Fast}) }

// update moves the camera on by one tick.
func (s *sky) update() { s.view.Update(time.Second / time.Duration(ebiten.TPS())) }

// draw paints the sky into the map's rectangle. ox, oy is where the cell grid
// starts on screen and cw, ch one cell, all in device pixels.
func (s *sky) draw(screen *ebiten.Image, ox, oy, cw, ch int) {
	s.mu.Lock()
	r := s.rect
	s.mu.Unlock()
	if r.Empty() {
		return
	}
	px := image.Rect(ox+r.Min.X*cw, oy+r.Min.Y*ch, ox+r.Max.X*cw, oy+r.Max.Y*ch)
	n := mapview.ArtScale(px.Dx())
	w, h := (px.Dx()+n-1)/n, (px.Dy()+n-1)/n
	s.mu.Lock()
	s.px, s.n, s.w, s.h = px, n, w, h
	s.mu.Unlock()
	pix := s.view.Render(w, h)
	if pix == nil {
		return
	}
	if s.img == nil || s.img.Bounds().Dx() != w || s.img.Bounds().Dy() != h {
		s.img = ebiten.NewImage(w, h)
	}
	s.img.WritePixels(pix.Pix)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(n), float64(n))
	op.GeoM.Translate(float64(px.Min.X), float64(px.Min.Y))
	screen.SubImage(px).(*ebiten.Image).DrawImage(s.img, op)
}
