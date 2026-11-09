package window

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/espresso20/ageforge/ui/mapstyle"
)

// stage holds the map images the game asks for (mapstyle.Stage). The game
// places them from its goroutine; the window builds each one off the draw
// thread, at the size of its cells, and keeps showing the last one until the
// next is ready.
type stage struct {
	mu    sync.Mutex
	slots map[string]*slot
}

type slot struct {
	want     mapstyle.Placement
	have     bool // the slot is placed
	asked    job  // what the build in flight, or the last one done, was for
	building bool // a build is running: never two at once
	done     *built

	// owned by the draw thread
	img   *ebiten.Image
	cells image.Rectangle
}

type job struct {
	key    uint64
	cw, ch int
}

type built struct {
	rgba *image.RGBA
	job  job
}

func newStage() *stage { return &stage{slots: map[string]*slot{}} }

func (s *stage) Place(name string, p mapstyle.Placement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl := s.slots[name]
	if sl == nil {
		sl = &slot{}
		s.slots[name] = sl
	}
	sl.want, sl.have = p, true
}

func (s *stage) Clear(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sl := s.slots[name]; sl != nil {
		sl.have = false
	}
}

// poll starts the builds that are due for cells of cw by ch pixels and hands
// back the images that finished since the last call, for the caller to upload.
// A slot builds again only when its key or the cell size is not what it last
// built for, and never while a build of its own is running.
func (s *stage) poll(cw, ch int) []finished {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []finished
	for name, sl := range s.slots {
		if sl.done != nil {
			out = append(out, finished{name, sl, sl.done})
			sl.done = nil
		}
		if j := (job{sl.want.Key, cw, ch}); sl.have && !sl.building && j != sl.asked {
			sl.asked, sl.building = j, true
			go func(sl *slot, build func(cw, ch int) *image.RGBA, j job) {
				img := build(j.cw, j.ch)
				s.mu.Lock()
				sl.building, sl.done = false, &built{img, j}
				s.mu.Unlock()
			}(sl, sl.want.Build, j)
		}
	}
	return out
}

type finished struct {
	name string
	sl   *slot
	b    *built
}

// show lists the slots that are placed and have an image, with where each goes.
func (s *stage) show() []*slot {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*slot
	for _, sl := range s.slots {
		if sl.have && sl.img != nil {
			sl.cells = sl.want.Cells
			out = append(out, sl)
		}
	}
	return out
}

// upload puts the finished images on the GPU. It runs on the draw thread.
func (s *stage) upload(done []finished) {
	for _, f := range done {
		if f.b.rgba == nil {
			continue
		}
		b := f.b.rgba.Bounds()
		if f.sl.img == nil || f.sl.img.Bounds().Size() != b.Size() {
			if f.sl.img != nil {
				f.sl.img.Deallocate()
			}
			f.sl.img = ebiten.NewImage(b.Dx(), b.Dy())
		}
		f.sl.img.WritePixels(f.b.rgba.Pix)
	}
}

// draw lays the images on screen, each scaled to its cells (an image built for
// another cell size stays on show until the new one is ready), ox and oy being
// where the cell grid starts.
func (s *stage) draw(screen *ebiten.Image, cw, ch int, ox, oy float64) {
	for _, sl := range s.show() {
		w, h := sl.cells.Dx()*cw, sl.cells.Dy()*ch
		b := sl.img.Bounds()
		if w <= 0 || h <= 0 {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(w)/float64(b.Dx()), float64(h)/float64(b.Dy()))
		op.GeoM.Translate(ox+float64(sl.cells.Min.X*cw), oy+float64(sl.cells.Min.Y*ch))
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(sl.img, op)
	}
}
