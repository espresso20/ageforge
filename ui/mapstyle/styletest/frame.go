package styletest

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// FrameBudget draws one frame of model in style at w by h, then n more, and
// fails the test when a frame takes longer than budget on average. It logs
// the time; label names the case in the log and the error.
func FrameBudget(t testing.TB, label string, style mapstyle.Style, model *mapmodel.Model, w, h, n int, budget time.Duration) {
	t.Helper()
	scr := capture.NewScreen(w, h)
	defer scr.Fini()
	r := mapstyle.Rect{W: w, H: h}
	f := mapstyle.Frame{Model: model}
	style.Draw(scr, r, f)
	start := time.Now()
	for i := 0; i < n; i++ {
		f.Anim = i
		style.Draw(scr, r, f)
	}
	per := time.Since(start) / time.Duration(n)
	t.Logf("%s: a %dx%d frame takes %v", label, w, h, per)
	if per > budget {
		t.Errorf("%s: a %dx%d frame takes %v, over %v", label, w, h, per, budget)
	}
}
