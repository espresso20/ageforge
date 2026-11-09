package window

import (
	"image"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The mouse on the sky: the wheel zooms about the cursor, a drag pans, a click
// selects and a double click selects and focuses. Only presses that start in
// the map's rectangle count; the rest of the window is the cells' business.

const (
	dragSlop    = 4                      // device pixels the pointer may wander and still be a click
	doubleClick = 400 * time.Millisecond // two clicks this close are a double click
)

// pointer is what the sky remembers of the mouse between frames.
type pointer struct {
	down       bool
	moved      bool
	sx, sy     int // where the press began
	lx, ly     int // where the pointer was last frame
	clickAt    time.Time
	clx, cly   int
	wheelAccum float64
}

// artAt is the art pixel under device pixel cx, cy, and whether it is in the map.
func (s *sky) artAt(cx, cy int) (ax, ay int, in bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.n == 0 || !image.Pt(cx, cy).In(s.px) {
		return 0, 0, false
	}
	return (cx - s.px.Min.X) / s.n, (cy - s.px.Min.Y) / s.n, true
}

// mouse reads the pointer for this frame.
func (s *sky) mouse(now time.Time) {
	cx, cy := ebiten.CursorPosition()
	if _, wy := ebiten.Wheel(); wy != 0 {
		if ax, ay, in := s.artAt(cx, cy); in {
			s.wheel(ax, ay, math.Max(-3, math.Min(3, wy)))
		}
	}
	p := &s.ptr
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		if _, _, in := s.artAt(cx, cy); in {
			*p = pointer{down: true, sx: cx, sy: cy, lx: cx, ly: cy, clickAt: p.clickAt, clx: p.clx, cly: p.cly}
		}
	case p.down && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
		if !p.moved && math.Hypot(float64(cx-p.sx), float64(cy-p.sy)) > dragSlop {
			p.moved = true
		}
		if p.moved {
			s.pan(float64(cx-p.lx), float64(cy-p.ly))
		}
		p.lx, p.ly = cx, cy
	case p.down:
		p.down = false
		if p.moved {
			return
		}
		ax, ay, _ := s.artAt(p.sx, p.sy)
		if now.Sub(p.clickAt) < doubleClick && math.Hypot(float64(p.sx-p.clx), float64(p.sy-p.cly)) <= 2*dragSlop {
			p.clickAt = time.Time{}
			s.focus(ax, ay)
			return
		}
		p.clickAt, p.clx, p.cly = now, p.sx, p.sy
		s.click(ax, ay)
	}
}

// size is the canvas in art pixels, as the last draw made it.
func (s *sky) size() (w, h, n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w, s.h, s.n
}

func (s *sky) wheel(ax, ay int, n float64) {
	if w, h, _ := s.size(); w > 0 {
		s.view.Wheel(ax, ay, w, h, n)
	}
}

func (s *sky) click(ax, ay int) {
	if w, h, _ := s.size(); w > 0 {
		s.view.Click(ax, ay, w, h)
	}
}

func (s *sky) focus(ax, ay int) {
	if w, h, _ := s.size(); w > 0 {
		s.view.Focus(ax, ay, w, h)
	}
}

// pan moves the sky by dx, dy device pixels.
func (s *sky) pan(dx, dy float64) {
	if w, _, n := s.size(); w > 0 {
		s.view.Pan(dx/float64(n), dy/float64(n), w)
	}
}
