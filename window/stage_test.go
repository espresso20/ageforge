package window

import (
	"image"
	"sync/atomic"
	"testing"
	"time"

	"github.com/espresso20/ageforge/ui/mapstyle"
)

// TestStageBuildsOncePerKey: a slot builds again only when its key or the cell
// size changes, never twice for the same job.
func TestStageBuildsOncePerKey(t *testing.T) {
	s := newStage()
	var n atomic.Int32
	cw, ch := 8, 16
	place := func(key uint64) {
		s.Place("map", mapstyle.Placement{Cells: image.Rect(0, 0, 4, 4), Key: key,
			Build: func(cw, ch int) *image.RGBA { n.Add(1); return image.NewRGBA(image.Rect(0, 0, 1, 1)) }})
	}
	settle := func(want int32) {
		t.Helper()
		for i := 0; i < 200 && len(s.poll(cw, ch)) == 0 && n.Load() < want; i++ {
			time.Sleep(5 * time.Millisecond)
		}
		for i := 0; i < 200; i++ {
			s.poll(cw, ch)
			s.mu.Lock()
			busy := s.slots["map"].building
			s.mu.Unlock()
			if !busy {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if got := n.Load(); got != want {
			t.Fatalf("%d builds, want %d", got, want)
		}
	}
	place(1)
	settle(1)
	for i := 0; i < 5; i++ {
		place(1)
		s.poll(cw, ch)
	}
	settle(1)
	place(2)
	settle(2)
	cw, ch = 10, 20
	settle(3)
}
