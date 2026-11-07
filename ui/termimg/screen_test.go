package termimg

import (
	"bytes"
	"image"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

type fakeTty struct{ out bytes.Buffer }

func (f *fakeTty) Start() error                { return nil }
func (f *fakeTty) Stop() error                 { return nil }
func (f *fakeTty) Drain() error                { return nil }
func (f *fakeTty) NotifyResize(func())         {}
func (f *fakeTty) Read([]byte) (int, error)    { return 0, nil }
func (f *fakeTty) Close() error                { return nil }
func (f *fakeTty) Write(b []byte) (int, error) { return f.out.Write(b) }
func (f *fakeTty) WindowSize() (tcell.WindowSize, error) {
	return tcell.WindowSize{Width: 80, Height: 24, PixelWidth: 800, PixelHeight: 480}, nil
}

type ttyScreen struct {
	tcell.Screen
	t *fakeTty
}

func (s ttyScreen) Tty() (tcell.Tty, bool) { return s.t, true }

func newTest(t *testing.T, m Mode) (*Screen, *fakeTty) {
	sim := tcell.NewSimulationScreen("")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	sim.SetSize(80, 24)
	ft := &fakeTty{}
	return Wrap(ttyScreen{sim, ft}, m, [2]int{}), ft
}

func frame(s *Screen, key uint64, renders *int) {
	s.Clear()
	s.Place(2, 1, 10, 5, key, func(cw, ch int) *image.RGBA {
		*renders++
		if cw != 10 || ch != 20 {
			panic("cell size not from the window")
		}
		return image.NewRGBA(image.Rect(0, 0, 10*cw, 5*ch))
	})
	s.Show()
}

func TestNoUploadWhenUnchanged(t *testing.T) {
	for _, m := range []Mode{Sixel, Kitty} {
		s, ft := newTest(t, m)
		n := 0
		frame(s, 1, &n)
		first := ft.out.Len()
		if n != 1 || first == 0 {
			t.Fatalf("%v: renders %d, bytes %d", m, n, first)
		}
		frame(s, 1, &n)
		frame(s, 1, &n)
		if n != 1 || ft.out.Len() != first {
			t.Errorf("%v: same key drew again: renders %d, bytes %d -> %d", m, n, first, ft.out.Len())
		}
		frame(s, 2, &n)
		if n != 2 || ft.out.Len() <= first {
			t.Errorf("%v: new key did not draw", m)
		}
	}
}

func TestRetireAndObscure(t *testing.T) {
	s, ft := newTest(t, Kitty)
	n := 0
	frame(s, 1, &n)
	ft.out.Reset()
	// a frame with no Place takes the picture down
	s.Clear()
	s.Show()
	if !strings.Contains(ft.out.String(), string(KittyDelete())) {
		t.Errorf("no delete: %q", ft.out.String())
	}
	// a write over the box in the same frame takes it down too
	frame(s, 1, &n)
	ft.out.Reset()
	s.Clear()
	s.Place(2, 1, 10, 5, 1, func(int, int) *image.RGBA { return image.NewRGBA(image.Rect(0, 0, 100, 100)) })
	s.SetContent(5, 3, 'x', nil, tcell.StyleDefault)
	s.Show()
	if ft.out.Len() == 0 || strings.Contains(ft.out.String(), "a=T") {
		t.Errorf("obscured picture was drawn: %q", ft.out.String())
	}
	// Fini takes it down
	frame(s, 1, &n)
	ft.out.Reset()
	s.Fini()
	if !strings.Contains(ft.out.String(), string(KittyDelete())) {
		t.Error("Fini left the picture")
	}
}

func TestNoneIsPassThrough(t *testing.T) {
	s, ft := newTest(t, None)
	n := 0
	frame(s, 1, &n)
	if n != 0 || ft.out.Len() != 0 {
		t.Errorf("none mode drew: %d %d", n, ft.out.Len())
	}
	if _, _, ok := s.CellPixels(); ok {
		t.Error("none mode offers pictures")
	}
}
