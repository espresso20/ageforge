package ui

import (
	"testing"

	"github.com/rivo/tview"
)

// TestShowNewGameNameModalAddsPage is a light lifecycle check: the modal must add
// its page to the supplied Pages under the expected name. The validation rule
// itself is covered by game.TestSanitizeSaveName (the modal calls the same
// game.ValidateSaveName wrapper), and the engine side is covered by
// game.TestStartNewNamedGame — a full focus/draw-loop UI test would be brittle, so
// we deliberately keep this to the add-page contract.
func TestShowNewGameNameModalAddsPage(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()

	if pages.HasPage(newGameNamePage) {
		t.Fatalf("page %q present before the modal was shown", newGameNamePage)
	}

	showNewGameNameModal(app, pages, nil, func(string) {
		t.Fatal("onConfirm should not fire when the modal merely opens")
	})

	if !pages.HasPage(newGameNamePage) {
		t.Errorf("showNewGameNameModal did not add page %q", newGameNamePage)
	}
}
