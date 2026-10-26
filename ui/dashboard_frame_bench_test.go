package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// BenchmarkDashboardFrame is one frame of the dashboard as the app draws it:
// a refresh from the engine and a draw of the whole screen, at 160x48. It is
// the cost a player pays every half second, so what VibeMode adds to it (one
// atomic load for each text the frame writes) has to stay too small to see.
func BenchmarkDashboardFrame(b *testing.B) {
	for _, on := range []bool{false, true} {
		name := "off"
		if on {
			name = "on"
		}
		b.Run("vibemode_"+name, func(b *testing.B) {
			b.Cleanup(game.SetDataDirForTest(b.TempDir()))
			b.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
			was := vibe.On()
			b.Cleanup(func() { vibe.Set(was) })
			vibe.Set(false)
			d, pages := newMiniDashboard(b)
			sim := newSimScreen(b, 160, 48)
			scr := theme.WrapScreen(sim)
			pages.SetRect(0, 0, 160, 48)
			frame := func() {
				d.refresh()
				pages.Draw(scr)
				scr.Show()
			}
			frame()
			vibe.Set(on)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				frame()
			}
		})
	}
}
