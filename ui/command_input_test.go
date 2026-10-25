package ui

import (
	"strings"
	"testing"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// drawPrompt draws a focused command input holding text on a w-wide
// simulated line, through theme.WrapScreen like the real app, and returns
// the line.
func drawPrompt(t *testing.T, eng *game.GameEngine, text string, w int) frame {
	t.Helper()
	in := newCommandInput(newCompleter(eng, nil))
	in.SetLabel("❯ ").SetFieldWidth(0).
		SetFieldBackgroundColor(theme.Color(theme.RoleBackground)).
		SetLabelColor(theme.Color(theme.RoleHighlight))
	in.SetText(text)
	in.Focus(func(tview.Primitive) {})
	sim := newSimScreen(t, w, 1)
	screen := theme.WrapScreen(sim)
	in.SetRect(0, 0, w, 1)
	in.Draw(screen)
	screen.Show()
	return simCells(sim)
}

// dimRun is the text drawn in the Dim role's colour.
func dimRun(f frame) string {
	dim := theme.Color(theme.RoleDim).Hex()
	var sb strings.Builder
	for _, c := range f.Cells() {
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

	f := drawPrompt(t, eng, "adv", 40)
	if got := f.Text(); got != "❯ advance" {
		t.Errorf("prompt line = %q, want %q", got, "❯ advance")
	}
	if got := dimRun(f); got != "ance" {
		t.Errorf("dim text = %q, want %q", got, "ance")
	}
	// The typed part is not dim.
	if fg, _, _ := f.Cells()[2].Style.Decompose(); fg.Hex() == theme.Color(theme.RoleDim).Hex() {
		t.Error("the typed text is drawn dim")
	}

	for _, text := range []string{"advance", "research list", "zzqx", "build nosuchbuilding", ""} {
		f := drawPrompt(t, eng, text, 40)
		if got := dimRun(f); got != "" {
			t.Errorf("%q: dim text %q, want none", text, got)
		}
		if got := f.Text(); got != strings.TrimRight("❯ "+text, " ") {
			t.Errorf("%q: prompt line = %q", text, got)
		}
	}

	// An argument completes too, and the ghost is clipped at the field's end.
	if got := dimRun(drawPrompt(t, eng, "plan bu", 40)); got != "ild" {
		t.Errorf("plan bu: dim text %q, want %q", got, "ild")
	}
	if got := drawPrompt(t, eng, "adv", 7).Text(); got != "❯ advan" {
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
		must(t, theme.SetActive(th.Key))
		f := drawPrompt(t, eng, "adv", 40)
		if got := dimRun(f); got != "ance" {
			t.Errorf("%s: dim text %q, want %q", th.Key, got, "ance")
		}
		for i := 5; i < 9; i++ {
			fg, bg, _ := f.Cells()[i].Style.Decompose()
			if fg.Hex() == bg.Hex() {
				t.Errorf("%s: ghost cell %d is invisible (fg == bg == %06x)", th.Key, i, fg.Hex())
			}
		}
	}
}
