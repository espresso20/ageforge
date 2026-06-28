package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// sentinel is an arbitrary RGB that no shipped theme uses, so a match proves the
// remap drove resolution rather than a coincidence.
var sentinel = tcell.NewRGBColor(0x12, 0x34, 0x56)

// restoreColorNames snapshots and restores the tcell.ColorNames keys this package
// owns, so the remap tests don't leave global state poisoned for other tests.
func restoreColorNames(t *testing.T) {
	t.Helper()
	saved := map[string]tcell.Color{}
	for _, m := range remappedNames {
		saved[m.name] = tcell.ColorNames[m.name]
	}
	t.Cleanup(func() {
		for name, c := range saved {
			tcell.ColorNames[name] = c
		}
	})
}

// TestRemap_MapDrivesResolution is the floor assertion (§3.6): after remapping
// ColorNames["gold"], tcell.GetColor("gold") — the SAME lookup tview's tag parser
// uses — returns the sentinel. If this fails, the name-remap strategy is dead at
// the root.
func TestRemap_MapDrivesResolution(t *testing.T) {
	restoreColorNames(t)

	tcell.ColorNames["gold"] = sentinel
	got := tcell.GetColor("gold")
	if got.Hex() != sentinel.Hex() {
		t.Fatalf("GetColor(\"gold\") = %06x, want sentinel %06x; ColorNames does not drive resolution",
			got.Hex(), sentinel.Hex())
	}
}

// TestRemap_RenderedCellRetints is the real §3.6 guard: render a "[gold]X" TextView
// to a SimulationScreen with gold remapped to the sentinel, then read back the
// rendered cell and assert its foreground IS the sentinel. tview re-resolves named
// tags through tcell.ColorNames on every Draw, so this passes today. If a future
// tview caches resolved colors at parse time, this FAILS loudly — signaling the §9
// semantic-token fallback.
func TestRemap_RenderedCellRetints(t *testing.T) {
	restoreColorNames(t)
	tcell.ColorNames["gold"] = sentinel

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("SimulationScreen.Init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(10, 1)

	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetText("[gold]X")
	// No border, origin at 0,0 so 'X' lands at cell (0,0).
	tv.SetRect(0, 0, 10, 1)
	tv.Draw(screen)
	screen.Show()

	cells, w, _ := screen.GetContents()
	if w < 1 || len(cells) == 0 {
		t.Fatalf("empty simulation screen contents (w=%d, cells=%d)", w, len(cells))
	}

	// Find the cell carrying 'X'. With no border it's cell 0, but scan to be robust
	// against tview padding tweaks across versions.
	var found bool
	var fg tcell.Color
	for _, c := range cells {
		if len(c.Runes) == 1 && c.Runes[0] == 'X' {
			fg, _, _ = c.Style.Decompose()
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("did not find rendered 'X' cell in simulation output")
	}
	if fg.Hex() != sentinel.Hex() {
		t.Fatalf("rendered [gold]X foreground = %06x, want sentinel %06x; "+
			"tview no longer re-resolves named tags through tcell.ColorNames — "+
			"switch to the §9 semantic-token fallback",
			fg.Hex(), sentinel.Hex())
	}
}

// TestApplyRemap_OverwritesFullSet asserts applyRemap rewrites the entire owned set
// (no name left stale) and pushes chrome defaults into tview.Styles.
func TestApplyRemap_OverwritesFullSet(t *testing.T) {
	restoreColorNames(t)

	// Poison every owned key first; applyRemap must overwrite all of them.
	for _, m := range remappedNames {
		tcell.ColorNames[m.name] = sentinel
	}

	applyRemap(Forge)

	for _, m := range remappedNames {
		want := Forge.Color(m.role)
		got := tcell.ColorNames[m.name]
		if got.Hex() != want.Hex() {
			t.Errorf("ColorNames[%q] = %06x, want Forge %s %06x",
				m.name, got.Hex(), m.role, want.Hex())
		}
	}

	if tview.Styles.BorderColor.Hex() != Forge.Color(RoleAccent).Hex() {
		t.Errorf("tview.Styles.BorderColor not set to Accent")
	}
	if tview.Styles.PrimitiveBackgroundColor.Hex() != Forge.Color(RoleBackground).Hex() {
		t.Errorf("tview.Styles.PrimitiveBackgroundColor not set to Background")
	}
}
