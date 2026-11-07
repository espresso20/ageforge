package termimg

import (
	"fmt"
	"image"

	"github.com/gdamore/tcell/v2"
)

// Screen wraps a tcell screen and keeps one picture over a rectangle of its
// cells. A draw asks for the picture with Place (the way it asks for a cell
// with SetContent); Show then writes it straight to the terminal once tcell
// has flushed its cells, and only when the picture is not the one already
// there.
//
// tcell does not know about the picture, so the cells under it are locked
// (tcell never redraws them) for as long as it is up. The picture is taken
// down, and the cells released to be redrawn, in the frame that no longer
// asks for it, that draws something over it (a pop-up) or that moves it.
type Screen struct {
	tcell.Screen
	mode  Mode
	cellW int // pixels per cell when the terminal does not say
	cellH int

	// this frame
	want     *request
	obscured bool

	// on screen
	shown   *request
	payload []byte
	termW   int // the screen size the picture was placed at
	termH   int
}

type request struct {
	x, y, w, h int
	key        uint64
	render     func(cellW, cellH int) *image.RGBA
	cw, ch     int
}

// Wrap makes a Screen over s showing pictures by mode. With None it is a plain
// pass-through. cell is the cell size in pixels the terminal reported at
// startup (0 if it did not), used when the window size has none.
func Wrap(s tcell.Screen, mode Mode, cell [2]int) *Screen {
	if cell[0] <= 0 || cell[1] <= 0 {
		cell = [2]int{8, 16}
	}
	return &Screen{Screen: s, mode: mode, cellW: cell[0], cellH: cell[1]}
}

// Mode is how the screen shows pictures.
func (s *Screen) Mode() Mode { return s.mode }

func (s *Screen) tty() (tcell.Tty, bool) {
	if s.mode == None {
		return nil, false
	}
	t, ok := s.Screen.Tty()
	return t, ok && t != nil
}

// CellPixels is the size of one cell in pixels, and whether pictures can be
// shown at all.
func (s *Screen) CellPixels() (w, h int, ok bool) {
	t, ok := s.tty()
	if !ok {
		return 0, 0, false
	}
	w, h = s.cellW, s.cellH
	if ws, err := t.WindowSize(); err == nil {
		if cw, ch := ws.CellDimensions(); cw > 0 && ch > 0 {
			w, h = cw, ch
		}
	}
	return w, h, true
}

// Place asks for the picture over the w by h cells at (x, y). key says what the
// picture is of: while it stays the same the picture is not drawn again. render
// makes the pixels, at the given cell size, and runs only when needed.
func (s *Screen) Place(x, y, w, h int, key uint64, render func(cellW, cellH int) *image.RGBA) {
	cw, ch, ok := s.CellPixels()
	if !ok || w <= 0 || h <= 0 {
		return
	}
	s.want = &request{x: x, y: y, w: w, h: h, key: key, render: render, cw: cw, ch: ch}
	s.obscured = false
}

// touch notes a write of n cells from (x, y): over the picture, it takes it down.
func (s *Screen) touch(x, y, n int) {
	if r := s.want; r != nil && y >= r.y && y < r.y+r.h && x < r.x+r.w && x+n > r.x {
		s.obscured = true
	}
}

func (s *Screen) Clear() {
	s.want, s.obscured = nil, false
	s.Screen.Clear()
}

func (s *Screen) Fill(r rune, st tcell.Style) {
	if s.want != nil {
		s.obscured = true
	}
	s.Screen.Fill(r, st)
}

func (s *Screen) SetContent(x, y int, primary rune, combining []rune, st tcell.Style) {
	s.touch(x, y, 1)
	s.Screen.SetContent(x, y, primary, combining, st)
}

func (s *Screen) SetCell(x, y int, st tcell.Style, ch ...rune) {
	s.touch(x, y, 1)
	s.Screen.SetCell(x, y, st, ch...)
}

func (s *Screen) Put(x, y int, str string, st tcell.Style) (string, int) {
	s.touch(x, y, len(str))
	return s.Screen.Put(x, y, str, st)
}

func (s *Screen) PutStr(x, y int, str string) {
	s.touch(x, y, len(str))
	s.Screen.PutStr(x, y, str)
}

func (s *Screen) PutStrStyled(x, y int, str string, st tcell.Style) {
	s.touch(x, y, len(str))
	s.Screen.PutStrStyled(x, y, str, st)
}

// Show flushes the cells, then puts the picture up if this frame asked for it.
func (s *Screen) Show() {
	want, obscured := s.want, s.obscured
	s.want, s.obscured = nil, false
	if obscured {
		want = nil
	}
	tw, th := s.Screen.Size()
	if s.shown != nil && (want == nil || !sameBox(want, s.shown) || want.cw != s.shown.cw || want.ch != s.shown.ch || tw != s.termW || th != s.termH) {
		s.retire()
	}
	s.Screen.Show()
	if want == nil || s.shown != nil && s.shown.key == want.key {
		return
	}
	img := want.render(want.cw, want.ch)
	if img == nil {
		return
	}
	switch s.mode {
	case Sixel:
		s.payload = EncodeSixel(img)
	case Kitty:
		s.payload = EncodeKitty(img, want.w, want.h)
	}
	s.shown, s.termW, s.termH = want, tw, th
	s.Screen.LockRegion(want.x, want.y, want.w, want.h, true)
	s.emit()
}

func sameBox(a, b *request) bool { return a.x == b.x && a.y == b.y && a.w == b.w && a.h == b.h }

// emit writes the picture, keeping the cursor where tcell left it.
func (s *Screen) emit() {
	t, ok := s.tty()
	if !ok || s.shown == nil {
		return
	}
	r := s.shown
	buf := make([]byte, 0, len(s.payload)+32)
	buf = append(buf, "\x1b7"...)
	buf = fmt.Appendf(buf, "\x1b[%d;%dH", r.y+1, r.x+1)
	buf = append(buf, s.payload...)
	buf = append(buf, "\x1b8"...)
	_, _ = t.Write(buf)
}

// retire takes the picture down: the cells under it are released and will be
// redrawn in the Show that follows (which covers a sixel); a kitty image is
// deleted.
func (s *Screen) retire() {
	r := s.shown
	if r == nil {
		return
	}
	s.shown, s.payload = nil, nil
	s.Screen.LockRegion(r.x, r.y, r.w, r.h, false)
	if s.mode == Kitty {
		if t, ok := s.tty(); ok {
			_, _ = t.Write(KittyDelete())
		}
	}
}

// Sync redraws everything, which wipes the terminal under the picture, so the
// picture goes up again.
func (s *Screen) Sync() {
	s.Screen.Sync()
	s.emit()
}

// Fini takes the picture down before the screen is given back.
func (s *Screen) Fini() {
	s.retire()
	s.Screen.Fini()
}
