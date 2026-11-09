// Package window shows the game in its own window. It is a thin host: the game
// is still drawn by tview into a simulation screen, and this package paints
// that screen's cells with Ebitengine and turns the window's keys into key
// events for it.
package window

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/theme"
)

// Options are what the window needs from main.
type Options struct {
	Host    *Host
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
	fr         frame
	prev, cur  []cellKey
	closing    bool
	frame      int
	script     *script
	adv        map[advKey]float64
	tface      map[faceID]*text.GoTextFace
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
	mw, mh := ebiten.Monitor().Size()
	scale := ebiten.Monitor().DeviceScaleFactor()
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
	grew := cols != g.cols || rows != g.rows
	g.cols, g.rows = cols, rows
	if g.buf != nil {
		g.buf.Deallocate()
	}
	g.buf = ebiten.NewImage(cols*cw, rows*ch)
	g.prev = nil // paint every cell again
	if grew || g.frame == 0 {
		g.o.Host.SetGrid(cols, rows)
	}
}

func (g *Game) holdFires(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || (d >= 30 && (d-30)%3 == 0)
}

// Update reads the keys and the window, once per tick.
func (g *Game) Update() error {
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
	host := g.o.Host
	if (zoom && inpututil.IsKeyJustPressed(ebiten.KeyV)) || (m.shift && inpututil.IsKeyJustPressed(ebiten.KeyInsert)) {
		for _, p := range mapChars([]rune(clipboardText())) {
			host.InjectKey(p.key, p.r, p.mod)
		}
		return
	}
	for _, k := range pressedKeys {
		if g.holdFires(k) {
			if p, ok := mapKey(k, m); ok {
				host.InjectKey(p.key, p.r, p.mod)
			}
		}
	}
	if takesText(m) {
		for _, p := range mapChars(ebiten.AppendInputChars(nil)) {
			host.InjectKey(p.key, p.r, p.mod)
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
	th := theme.Active()
	bg := toRGBA(th.Color(theme.RoleBackground), color.RGBA{A: 255})
	screen.Fill(bg)
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

// cursor draws the text cursor tview asks for, as a bar at the left of its
// cell, blinking.
func (g *Game) cursor(screen *ebiten.Image, ox, oy float64) {
	f := &g.fr
	if !f.cursorShowing || f.curX < 0 || f.curY < 0 || f.curX >= g.cols || f.curY >= g.rows || (g.frame/30)%2 == 1 {
		return
	}
	w := float32(max(2, g.cw/8))
	fg, _, _ := g.colors(f.cells[f.curY*f.cols+f.curX].Style)
	vector.FillRect(screen, float32(ox)+float32(f.curX*g.cw), float32(oy)+float32(f.curY*g.ch), w, float32(g.ch), fg, false)
}

// A cellKey is everything that decides how a cell looks, so two equal keys
// need no repaint.
type cellKey struct {
	r       rune
	comb    string
	st      tcell.Style
	wide    bool
	covered bool // the right half of a wide rune
}

func (g *Game) keyAt(i int, wasWide bool) cellKey {
	c := g.fr.cells[i]
	if wasWide {
		return cellKey{covered: true, st: c.Style}
	}
	k := cellKey{r: ' ', st: c.Style}
	if len(c.Runes) > 0 && c.Runes[0] != 0 {
		k.r = c.Runes[0]
		if len(c.Runes) > 1 {
			k.comb = string(c.Runes[1:])
		}
		if k.r >= 0x1100 && k.comb == "" && uniseg.StringWidth(string(k.r)) == 2 {
			k.wide = true
		}
	}
	return k
}

func (g *Game) paint() {
	g.o.Host.snapshot(&g.fr)
	if g.fr.cols != g.cols || g.fr.rows != g.rows {
		return // the game has not drawn at the new size yet
	}
	n := g.cols * g.rows
	if len(g.prev) != n {
		g.prev = make([]cellKey, n)
		for i := range g.prev {
			g.prev[i].r = -1 // matches nothing
		}
	}
	g.cur = append(g.cur[:0], make([]cellKey, n)...)
	for y := 0; y < g.rows; y++ {
		wasWide := false
		for x := 0; x < g.cols; x++ {
			i := y*g.cols + x
			k := g.keyAt(i, wasWide)
			g.cur[i] = k
			wasWide = k.wide
		}
	}
	for i, k := range g.cur {
		// The right half of a wide rune is painted with it, never on its own:
		// a fill there would wipe half the glyph.
		if k != g.prev[i] && !k.covered {
			g.drawCell(i%g.cols, i/g.cols, k)
		}
	}
	g.prev, g.cur = g.cur, g.prev
}
