package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// BenchmarkDashboardFrame is one frame of the dashboard as the app draws it:
// a refresh from the engine and a draw of the whole screen, at 160x48. It is
// the cost a player pays every half second.
func BenchmarkDashboardFrame(b *testing.B) {
	b.Cleanup(game.SetDataDirForTest(b.TempDir()))
	b.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	d, pages := newMiniDashboard(b)
	sim := newSimScreen(b, 160, 48)
	scr := sim
	pages.SetRect(0, 0, 160, 48)
	frame := func() {
		d.refresh()
		drawOn(scr, pages)
		scr.Show()
	}
	frame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		frame()
	}
}
