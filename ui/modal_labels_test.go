package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// renderPages draws pages onto a fresh w×h SimulationScreen and returns the
// screen as plain text, one line per row.
func renderPages(t *testing.T, pages *tview.Pages, w, h int) string {
	t.Helper()
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatalf("sim.Init: %v", err)
	}
	defer sim.Fini()
	sim.SetSize(w, h)
	pages.SetRect(0, 0, w, h)
	pages.Draw(sim)
	sim.Show()

	cells, cw, ch := sim.GetContents()
	var sb strings.Builder
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			b := cells[y*cw+x].Bytes
			if len(b) == 0 {
				b = []byte{' '}
			}
			sb.Write(b)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func newLabelTestDashboard() (*Dashboard, *tview.Pages) {
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), game.NewGameEngine(), pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	return d, pages
}

// Button labels like "[ENDURE]" parse as tview style tags and render blank
// unless escaped. These check what actually lands on screen, shortcuts included.
func TestCatastropheModalButtonLabelsVisible(t *testing.T) {
	d, pages := newLabelTestDashboard()
	d.showCatastropheModal("stone_era")
	screen := renderPages(t, pages, 160, 50)
	for _, want := range []string{"[E] ENDURE", "[S] SUCCUMB", "[D] Defer — Decide Later"} {
		if !strings.Contains(screen, want) {
			t.Errorf("catastrophe modal: button label %q not on screen\n%s", want, screen)
		}
	}
}

func TestAncientMemoryModalButtonLabelsVisible(t *testing.T) {
	d, pages := newLabelTestDashboard()
	d.showAncientMemoryModal("fire_mastery", "")
	screen := renderPages(t, pages, 160, 50)
	for _, want := range []string{"[A] ACCEPT", "[D] Decline"} {
		if !strings.Contains(screen, want) {
			t.Errorf("ancient memory modal: button label %q not on screen\n%s", want, screen)
		}
	}
}
