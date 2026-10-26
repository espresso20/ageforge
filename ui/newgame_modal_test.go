package ui

import (
	"testing"

	"github.com/rivo/tview"
)

// TestShowAccountNameConfirmModalAddsPage is the add-page contract for the first-run
// account confirmation step. Submitting a name in the entry modal opens this confirm
// modal before any account is created (the name is the identity, so a typo must be
// catchable). The create path (game.CreateAccount) is covered elsewhere; this is a
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
