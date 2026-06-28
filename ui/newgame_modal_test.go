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

// TestShowAccountNameConfirmModalAddsPage is the add-page contract for the first-run
// account confirmation step. Submitting a name in the entry modal opens this confirm
// modal before any account is created (the name is the identity, so a typo must be
// catchable). The create path (game.CreateNamedAccount) is covered elsewhere; this is a
// presentation-only step, so we keep the check to the page-registration contract and
// assert neither callback fires merely on open.
func TestShowAccountNameConfirmModalAddsPage(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()

	if pages.HasPage(accountNameConfirmPage) {
		t.Fatalf("page %q present before the confirm modal was shown", accountNameConfirmPage)
	}

	showAccountNameConfirmModal(app, pages, "Imperium",
		func(string) { t.Fatal("onCreate should not fire when the confirm modal merely opens") },
		func() { t.Fatal("onRetype should not fire when the confirm modal merely opens") },
	)

	if !pages.HasPage(accountNameConfirmPage) {
		t.Errorf("showAccountNameConfirmModal did not add page %q", accountNameConfirmPage)
	}
}
