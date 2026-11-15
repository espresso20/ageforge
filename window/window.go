// Package window shows the game in its own window. It is a thin host: the game
// is drawn by the kit into a buffer of cells, and this package paints
// those cells with Ebitengine and turns the window's keys into key
// events for it.
package window

import (
	"errors"
	"image"
	"log"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/espresso20/ageforge/platform/steam"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A Screen is the game as the window drives it: it is told the grid's size
// in cells, the keys and the pointer, and it hands back the cells to show.
// None of its methods may wait for the game.
type Screen interface {
	Resize(cols, rows int)
	Key(k kit.Key)
	Mouse(m kit.Mouse) // the pointer, at a cell of the grid
	// Frame copies the cells the game shows now into dst.
	Frame(dst *kit.Buffer)
}

// Options are what the window needs from main.
type Options struct {
	Screen  Screen
	Dir     string          // where the window's settings are kept
	Done    <-chan struct{} // closed when the game has ended; the window then closes
	OnClose func()          // the player closed the window: leave the game as Ctrl+C does
	Script  string          // dev only: a file of steps for a run nobody is watching (see script.go)
}

// Game is the Ebitengine game: it implements ebiten.Game.
type Game struct {
	o     Options
	faces *faces

	mu  sync.Mutex // guards set, which the game's goroutine reads through the Controls
	set Settings

	scale      float64 // device scale factor
	outW, outH float64 // the window, in device-independent pixels
	pw, ph     int     // the window, in device pixels
	size       float64 // the font's size in device pixels
	cw, ch     int     // one cell, in device pixels
	sky        *sky    // the map of space, drawn under the cells
	asc        float64 // top of the line box to the baseline
	cols, rows int
	buf        *ebiten.Image
	fr         kit.Buffer // the game's cells this frame
	prev       []kit.Cell // what is on the buffer image now
	closing    bool
	frame      int
	script     *script
	adv        map[advKey]float64
	tface      map[faceID]*text.GoTextFace
	ptr        cellPointer
	pointing   kit.Pointer                   // the shape the mouse pointer was last given
	pics       map[*image.RGBA]*ebiten.Image // the pictures on show, as the graphics card holds them
	cellPx     atomic.Uint64                 // the cell's width and height in pixels, for CellSize
	prevPics   []kit.Pic                     // the pictures of the frame before
}

// New is a window for the host, with the settings saved in dir.
func New(o Options) *Game {
	g := &Game{o: o, faces: newFaces(), sky: newSky(), adv: map[advKey]float64{}, tface: map[faceID]*text.GoTextFace{}, scale: 1}
	g.set = loadSettings(o.Dir).clean(1)
	if o.Script != "" {
		g.script = loadScript(o.Script)
	}
	return g
}

// Run opens the window and runs it until the game ends or the window is
// closed. It must be called on the main goroutine.
func (g *Game) Run() error {
	// Steam must start before the graphics device exists, for the overlay.
	if err := steam.Start(); err != nil {
		log.Printf("steam: %v", err)
	}
	// No display at all (every screen asleep, or a session with none): say
	// so, where asking a missing monitor for its size would crash.
	mon := ebiten.Monitor()
	if mon == nil {
		return errors.New("no display found: AgeForge needs a screen to open its window")
	}
	mw, mh := mon.Size()
	scale := mon.DeviceScaleFactor()
	g.scale = scale
	g.set = g.set.clean(scale)
	if g.set.Width == 0 {
		g.set.Width, g.set.Height = mw*8/10, mh*8/10
	}
	ebiten.SetWindowTitle("AgeForge")
	ebiten.SetWindowSize(g.set.Width, g.set.Height)
	ebiten.SetWindowSizeLimits(480, 300, -1, -1)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowClosingHandled(true)
	ebiten.SetFullscreen(g.set.Fullscreen)
	ebiten.SetTPS(60)
	err := ebiten.RunGame(g)
	g.save()
	return err
}

func (g *Game) save() {
	g.mu.Lock()
	s := g.set
	g.mu.Unlock()
	_ = saveSettings(g.o.Dir, s)
}

// Layout is not used (LayoutF is) but the interface wants it.
func (g *Game) Layout(w, h int) (int, int) { return w, h }

// LayoutF makes the screen as many device pixels as the window has, so a
// Retina display shows each glyph pixel for pixel.
func (g *Game) LayoutF(w, h float64) (float64, float64) {
	g.scale = ebiten.Monitor().DeviceScaleFactor()
	g.outW, g.outH = w, h
	g.pw, g.ph = int(math.Round(w*g.scale)), int(math.Round(h*g.scale))
	return float64(g.pw), float64(g.ph)
}

// fontMetrics sizes one cell for the font at the current size.
func (g *Game) fontMetrics() (size float64, cw, ch int, asc float64) {
	g.mu.Lock()
	s := g.set.clean(g.scale)
	g.mu.Unlock()
	f, file := s.Font, jbRegular
	size = float64(s.FontSize) * g.scale
	if f == FontPixel {
		file = departure
		size = float64(pixelEm * s.PixelScale)
	}
	file.load()
	face := g.goFace(file, size)
	m := face.Metrics()
	return size, max(1, int(math.Round(text.Advance("M", face)))), max(1, int(math.Round(m.HAscent+m.HDescent+m.HLineGap))), m.HAscent + m.HLineGap/2
}

type faceID struct {
	f    *fontFile
	size float64
}

func (g *Game) goFace(f *fontFile, size float64) *text.GoTextFace {
	id := faceID{f, size}
	if tf, ok := g.tface[id]; ok {
		return tf
	}
	tf := &text.GoTextFace{Source: f.src, Size: size}
	g.tface[id] = tf
	return tf
}

// relayout fits the grid to the window and the font, and tells the game when
// the grid changed.
func (g *Game) relayout() {
	size, cw, ch, asc := g.fontMetrics()
	cols, rows := gridSize(g.pw, g.ph, cw, ch)
	g.size, g.asc = size, asc
	if cw == g.cw && ch == g.ch && cols == g.cols && rows == g.rows && g.buf != nil {
		return
	}
	g.cw, g.ch = cw, ch
	g.cellPx.Store(uint64(cw)<<32 | uint64(ch))
	grew := cols != g.cols || rows != g.rows
	g.cols, g.rows = cols, rows
	if g.buf != nil {
		g.buf.Deallocate()
	}
	g.buf = ebiten.NewImage(cols*cw, rows*ch)
	g.prev = nil // paint every cell again
	if grew || g.frame == 0 {
		g.o.Screen.Resize(cols, rows)
	}
}

func (g *Game) holdFires(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || (d >= 30 && (d-30)%3 == 0)
}

// Update reads the keys and the window, once per tick.
func (g *Game) Update() error {
	steam.Frame()
	select {
	case <-g.o.Done:
		return ebiten.Termination
	default:
	}
	if ebiten.IsWindowBeingClosed() && !g.closing {
		g.closing = true
		go g.o.OnClose()
	}
	g.relayout()
	g.frame++
	if !ebiten.IsFullscreen() {
		g.mu.Lock()
		g.set.Width, g.set.Height = int(g.outW), int(g.outH)
		g.mu.Unlock()
	}
	g.mu.Lock()
	g.set.Fullscreen = ebiten.IsFullscreen()
	g.mu.Unlock()
	if g.script != nil && g.script.step(g) {
		return ebiten.Termination
	}
	g.keys()
	g.pointer()
	g.sky.mouse(time.Now())
	g.sky.update()
	return nil
}

func (g *Game) keys() {
	mac := runtime.GOOS == "darwin"
	m := mods{
		ctrl:  ebiten.IsKeyPressed(ebiten.KeyControl),
		alt:   ebiten.IsKeyPressed(ebiten.KeyAlt),
		shift: ebiten.IsKeyPressed(ebiten.KeyShift),
		meta:  ebiten.IsKeyPressed(ebiten.KeyMeta),
	}
	// The window's own keys come first and never reach the game.
	zoom := m.ctrl && !m.alt // Ctrl with Alt is AltGr, which types text
	if mac {
		zoom = m.meta
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyF11),
		m.alt && inpututil.IsKeyJustPressed(ebiten.KeyEnter),
		mac && m.meta && m.ctrl && inpututil.IsKeyJustPressed(ebiten.KeyF):
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		return
	case zoom && (g.holdFires(ebiten.KeyEqual) || g.holdFires(ebiten.KeyNumpadAdd)):
		g.zoomBy(1)
		return
	case zoom && (g.holdFires(ebiten.KeyMinus) || g.holdFires(ebiten.KeyNumpadSubtract)):
		g.zoomBy(-1)
		return
	case zoom && (inpututil.IsKeyJustPressed(ebiten.KeyDigit0) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad0)):
		g.zoomReset()
		return
	}
	game := g.o.Screen
	if (zoom && inpututil.IsKeyJustPressed(ebiten.KeyV)) || (m.shift && inpututil.IsKeyJustPressed(ebiten.KeyInsert)) {
		for _, p := range mapChars([]rune(clipboardText())) {
			game.Key(p.key)
		}
		return
	}
	for _, k := range pressedKeys {
		if g.holdFires(k) {
			if p, ok := mapKey(k, m); ok {
				game.Key(p.key)
			}
		}
	}
	if takesText(m) {
		for _, p := range mapChars(ebiten.AppendInputChars(nil)) {
			game.Key(p.key)
		}
	}
}

// zoomBy changes the text size one step: a point for the smooth font, one
// pixel of the pixel font (a whole multiple of its own) for that one.
func (g *Game) zoomBy(d int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	s := g.set.clean(g.scale)
	if s.Font == FontPixel {
		s.PixelScale += d
	} else {
		s.FontSize += d
	}
	g.set = s.clean(g.scale)
}

func (g *Game) zoomReset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.set.FontSize, g.set.PixelScale = defaultFontSize, pixelDefault(g.scale)
}

// FontChoice is the font now in use, "mono" or "pixel". It is safe to call from
// the game's goroutine.
func (g *Game) FontChoice() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return string(g.set.Font)
}

// CellSize is a cell of the grid in pixels, so a picture can be drawn pixel
// for pixel; 0, 0 before the window has been laid out.
func (g *Game) CellSize() (w, h int) {
	v := g.cellPx.Load()
	return int(v >> 32), int(v & 0xffffffff)
}

// NextFont switches to the other font and keeps the choice for the next run.
func (g *Game) NextFont() {
	g.mu.Lock()
	if g.set.Font == FontPixel {
		g.set.Font = FontMono
	} else {
		g.set.Font = FontPixel
	}
	g.mu.Unlock()
	g.save()
}

// Draw paints the cells that changed, then the buffer.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(theme.RGBA(theme.RoleBackground))
	if g.buf == nil {
		return
	}
	g.paint()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64((g.pw-g.cols*g.cw)/2), float64((g.ph-g.rows*g.ch)/2))
	// Under the cells, so the sky shows where a cell is clear.
	g.sky.draw(screen, int(op.GeoM.Element(0, 2)), int(op.GeoM.Element(1, 2)), g.cw, g.ch)
	screen.DrawImage(g.buf, op)
	g.cursor(screen, op.GeoM.Element(0, 2), op.GeoM.Element(1, 2))
	if g.script != nil {
		g.script.capture(screen)
	}
}

// cursor draws the text cursor the game asks for, as a bar at the left of
// its cell, blinking.
func (g *Game) cursor(screen *ebiten.Image, ox, oy float64) {
	f := &g.fr
	at := f.Cursor
	if !f.CursorOn || at.X < 0 || at.Y < 0 || at.X >= g.cols || at.Y >= g.rows || f.W != g.cols || (g.frame/30)%2 == 1 {
		return
	}
	w := float32(max(2, g.cw/8))
	fg, _ := cellColors(f.At(at.X, at.Y))
	vector.FillRect(screen, float32(ox)+float32(at.X*g.cw), float32(oy)+float32(at.Y*g.ch), w, float32(g.ch), fg, false)
}

// paint takes the game's cells and draws the ones that changed onto the
// buffer image. Two equal cells look the same, so an unchanged cell needs no
// repaint.
func (g *Game) paint() {
	g.o.Screen.Frame(&g.fr)
	if g.fr.Pointer != g.pointing {
		g.pointing = g.fr.Pointer
		ebiten.SetCursorShape(pointerShapes[g.pointing])
	}
	if g.fr.W != g.cols || g.fr.H != g.rows {
		return // the game has not drawn at the new size yet
	}
	n := g.cols * g.rows
	if len(g.prev) != n {
		g.prev = make([]kit.Cell, n)
		for i := range g.prev {
			g.prev[i].R = -1 // matches nothing
		}
	}
	g.syncPics()
	for i, c := range g.fr.Cells {
		// The right half of a wide rune is painted with it, never on its own:
		// a fill there would wipe half the glyph.
		if c != g.prev[i] && !c.Cont {
			g.drawCell(i%g.cols, i/g.cols, c)
		}
	}
	copy(g.prev, g.fr.Cells)
}
