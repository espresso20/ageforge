package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// drawPrompt draws a focused command input holding text on a w-wide
// simulated line, through theme.WrapScreen like the real app, and returns
// the line's cells.
func drawPrompt(t *testing.T, eng *game.GameEngine, text string, w int) []tcell.SimCell {
	t.Helper()
	in := newCommandInput(newCompleter(eng, nil))
	in.SetLabel("❯ ").SetFieldWidth(0).
		SetFieldBackgroundColor(theme.Color(theme.RoleBackground)).
		SetLabelColor(theme.Color(theme.RoleHighlight))
	in.SetText(text)
	in.Focus(func(tview.Primitive) {})
	sim := tcell.NewSimulationScreen("UTF-8")
	if err := sim.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sim.Fini)
	sim.SetSize(w, 1)
	screen := theme.WrapScreen(sim)
	in.SetRect(0, 0, w, 1)
	in.Draw(screen)
	screen.Show()
	cells, _, _ := sim.GetContents()
	return cells
}

// lineText is the cells as text, trailing blanks trimmed.
func lineText(cells []tcell.SimCell) string {
	var sb strings.Builder
	for _, c := range cells {
		if len(c.Runes) == 0 {
			sb.WriteByte(' ')
		} else {
			sb.WriteRune(c.Runes[0])
		}
	}
	return strings.TrimRight(sb.String(), " ")
}

// dimRun is the text drawn in the Dim role's colour.
func dimRun(cells []tcell.SimCell) string {
	dim := theme.Color(theme.RoleDim).Hex()
	var sb strings.Builder
	for _, c := range cells {
		fg, _, _ := c.Style.Decompose()
		if len(c.Runes) > 0 && c.Runes[0] != ' ' && fg.Hex() == dim {
			sb.WriteRune(c.Runes[0])
		}
	}
	return sb.String()
}

func TestCommandInputGhostText(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()

	cells := drawPrompt(t, eng, "adv", 40)
	if got := lineText(cells); got != "❯ advance" {
		t.Errorf("prompt line = %q, want %q", got, "❯ advance")
	}
	if got := dimRun(cells); got != "ance" {
		t.Errorf("dim text = %q, want %q", got, "ance")
	}
	// The typed part is not dim.
	if fg, _, _ := cells[2].Style.Decompose(); fg.Hex() == theme.Color(theme.RoleDim).Hex() {
		t.Error("the typed text is drawn dim")
	}

	for _, text := range []string{"advance", "research list", "zzqx", "build nosuchbuilding", ""} {
		cells := drawPrompt(t, eng, text, 40)
		if got := dimRun(cells); got != "" {
			t.Errorf("%q: dim text %q, want none", text, got)
		}
		if got := lineText(cells); got != strings.TrimRight("❯ "+text, " ") {
			t.Errorf("%q: prompt line = %q", text, got)
		}
	}

	// An argument completes too, and the ghost is clipped at the field's end.
	if got := dimRun(drawPrompt(t, eng, "plan bu", 40)); got != "ild" {
		t.Errorf("plan bu: dim text %q, want %q", got, "ild")
	}
	if got := lineText(drawPrompt(t, eng, "adv", 7)); got != "❯ advan" {
		t.Errorf("narrow field: %q, want the ghost clipped to %q", got, "❯ advan")
	}
}

// TestCommandInputGhostEveryTheme: the ghost text is visible (its colour is
// not its background) under every theme.
func TestCommandInputGhostEveryTheme(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	restoreForge(t)
	eng := game.NewGameEngine()
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		cells := drawPrompt(t, eng, "adv", 40)
		if got := dimRun(cells); got != "ance" {
			t.Errorf("%s: dim text %q, want %q", th.Key, got, "ance")
		}
		for i := 5; i < 9; i++ {
			fg, bg, _ := cells[i].Style.Decompose()
			if fg.Hex() == bg.Hex() {
				t.Errorf("%s: ghost cell %d is invisible (fg == bg == %06x)", th.Key, i, fg.Hex())
			}
		}
	}
}
