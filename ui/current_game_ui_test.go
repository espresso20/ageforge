package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// current_game_ui_test.go: the current game as the player sees and sets it,
// in the Load Game browser and in the Accounts panel, and how the main
// menu follows. Every test runs in a temp data root.

// browserFor opens the Load Game browser of a rig.
func browserFor(r *menuRig) *loadGameBrowser {
	b := newLoadGameBrowser(r.app, r.pages, r.eng, "splash", true)
	r.pages.AddPage(loadGamePage, b.root, true, true)
	r.app.SetFocus(b.root)
	return b
}

// rowOf is the browser's row for a save, as text.
func rowOf(b *loadGameBrowser, name string) string {
	for i, s := range b.saves {
		if s.Name == name {
			text, _ := b.list.GetItemText(i)
			return untag(text)
		}
	}
	return ""
}

// TestLoadBrowserMarksTheMainGame: with one save, many saves, a marked
// save and a marked save that is deleted, the browser shows which save
// Continue opens, M marks and unmarks the main game, and the main menu
// follows.
func TestLoadBrowserMarksTheMainGame(t *testing.T) {
	// One save: Continue opens it, as the game last played.
	r := newMenuRig(t, "ashford")
	b := browserFor(r)
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "1 save") || !strings.Contains(sub, "Continue opens ashford") {
		t.Errorf("one save: the subtitle is %q", sub)
	}
	if row := rowOf(b, "ashford"); !strings.Contains(row, "▸ continue") {
		t.Errorf("one save: its row is %q", row)
	}
	if d := untag(b.detail.GetText(false)); !strings.Contains(d, "Continue opens this game") || !strings.Contains(d, "last played") {
		t.Errorf("one save: the detail is %q", d)
	}
	b.back()

	// Many saves: the one played last.
	r.newGame("brackwater")
	r.newGame("corvale")
	b = browserFor(r)
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "3 saves") || !strings.Contains(sub, "Continue opens corvale") {
		t.Errorf("many saves: the subtitle is %q", sub)
	}
	for _, name := range []string{"ashford", "brackwater"} {
		if row := rowOf(b, name); strings.Contains(row, "continue") || strings.Contains(row, "main game") {
			t.Errorf("many saves: the row of %s is marked: %q", name, row)
		}
	}

	// M on another save marks it as the main game.
	b.selectByName("ashford")
	r.key(tcell.KeyRune, 'm')
	if mainGame, _ := r.acct.GameRecord(); mainGame != "ashford" {
		t.Fatalf("M did not mark ashford: the main game is %q", mainGame)
	}
	if row := rowOf(b, "ashford"); !strings.Contains(row, "◆ main game") {
		t.Errorf("a marked save: its row is %q", row)
	}
	if row := rowOf(b, "corvale"); strings.Contains(row, "continue") {
		t.Errorf("a marked save: the game last played still says continue: %q", row)
	}
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "Continue opens ashford") {
		t.Errorf("a marked save: the subtitle is %q", sub)
	}
	if got, _ := b.selected(); got.Name != "ashford" {
		t.Errorf("after M the selection moved to %q", got.Name)
	}
	if d := untag(b.detail.GetText(false)); !strings.Contains(d, "The main game") {
		t.Errorf("a marked save: the detail is %q", d)
	}
	// The main menu follows when the browser closes.
	r.key(tcell.KeyEsc, 0)
	if r.front() != "splash" || r.m.cur.Save.Name != "ashford" || r.m.cur.Why != game.CurrentMain {
		t.Errorf("back at the menu: front %q, Continue opens %q (%s)", r.front(), r.m.cur.Save.Name, r.m.cur.Why)
	}
	if got := r.m.view.items[0]; got.id != miContinue || got.extras[0] != "ashford" {
		t.Errorf("the menu's Continue names %q", got.extras)
	}

	// Playing another game does not unseat the main game.
	b = browserFor(r)
	b.selectByName("brackwater")
	r.key(tcell.KeyEnter, 0)
	waitRunning(t, r.eng)
	r.eng.Stop()
	r.pages.SwitchToPage("splash")
	if r.m.cur.Save.Name != "ashford" {
		t.Errorf("after playing brackwater Continue opens %q, want the main game", r.m.cur.Save.Name)
	}

	// A rename carries the mark.
	if err := game.RenameSave("ashford", "ashford_ii"); err != nil {
		t.Fatal(err)
	}
	_ = r.acct.RenameGame("ashford", "ashford_ii") // the browser's rename does this
	// M on the main game unmarks it: the game last played is current again.
	b = browserFor(r)
	b.selectByName("ashford_ii")
	r.key(tcell.KeyRune, 'M')
	if mainGame, _ := r.acct.GameRecord(); mainGame != "" {
		t.Errorf("M on the main game left %q marked", mainGame)
	}
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "Continue opens brackwater") {
		t.Errorf("unmarked: the subtitle is %q", sub)
	}

	// A marked save is deleted from the browser: the mark goes with it.
	r.key(tcell.KeyRune, 'm') // mark ashford_ii again
	if mainGame, _ := r.acct.GameRecord(); mainGame != "ashford_ii" {
		t.Fatalf("the main game is %q", mainGame)
	}
	r.key(tcell.KeyRune, 'd')
	if r.front() != "load_game_delete" {
		t.Fatalf("D opens %q", r.front())
	}
	r.key(tcell.KeyTab, 0) // Cancel, then Delete
	r.key(tcell.KeyEnter, 0)
	if game.SaveExists("ashford_ii") {
		t.Fatal("the save was not deleted")
	}
	if mainGame, _ := r.acct.GameRecord(); mainGame != "" {
		t.Errorf("the deleted save is still on record as the main game: %q", mainGame)
	}
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "2 saves") || !strings.Contains(sub, "Continue opens brackwater") {
		t.Errorf("after the delete: the subtitle is %q", sub)
	}
	r.key(tcell.KeyEsc, 0)
	if r.m.cur.Save.Name != "brackwater" || r.m.cur.Why != game.CurrentLast {
		t.Errorf("back at the menu Continue opens %q (%s)", r.m.cur.Save.Name, r.m.cur.Why)
	}
}

// TestLoadBrowserWithNoSaves: nothing to continue, and M does nothing.
func TestLoadBrowserWithNoSaves(t *testing.T) {
	r := newMenuRig(t)
	b := browserFor(r)
	if sub := untag(b.subtitle.GetText(false)); !strings.Contains(sub, "0 saves") || strings.Contains(sub, "Continue") {
		t.Errorf("no saves: the subtitle is %q", sub)
	}
	r.key(tcell.KeyRune, 'm')
	if mainGame, last := r.acct.GameRecord(); mainGame != "" || last != "" {
		t.Errorf("no saves: M recorded %q and %q", mainGame, last)
	}
}

// pressModal answers the dialog in front: the button at index.
func pressModal(t *testing.T, r *menuRig, index int) {
	t.Helper()
	_, front := r.pages.GetFrontPage()
	if _, ok := front.(*tview.Modal); !ok {
		t.Fatalf("the page in front (%q) is not a dialog", r.front())
	}
	for i := 0; i < index; i++ {
		r.key(tcell.KeyTab, 0)
	}
	r.key(tcell.KeyEnter, 0)
}

// TestAccountsPanelDeletesTheSelectedAccountsSaves: "Delete all saves" is
// in the Accounts panel now, keeps its confirmation, and acts on the
// account that is selected, the current one or another. The account
// itself stays.
func TestAccountsPanelDeletesTheSelectedAccountsSaves(t *testing.T) {
	r := newMenuRig(t, "ashford", "brackwater")
	accounts := filepath.Dir(game.DataDir())
	first := r.acct
	_ = first.SetMainGame("ashford")

	// A second account with a save of its own; the first stays current.
	second, err := game.CreateAccount("Second Player")
	if err != nil {
		t.Fatal(err)
	}
	mk := game.NewGameEngine()
	mk.SetAccount(second)
	if err := mk.StartNewNamedGame("delmere"); err != nil {
		t.Fatal(err)
	}
	if _, err := game.SwitchAccount(first.AccountID); err != nil {
		t.Fatal(err)
	}
	savesOf := func(id string) int {
		m, _ := filepath.Glob(filepath.Join(accounts, id, "saves", "*.json"))
		return len(m)
	}
	if savesOf(first.AccountID) != 2 || savesOf(second.AccountID) != 1 {
		t.Fatalf("staging: %d and %d saves", savesOf(first.AccountID), savesOf(second.AccountID))
	}

	r.press('a')
	if r.front() != accountsPage {
		t.Fatalf("a opens %q", r.front())
	}
	p := newAccountsPanel(r.app, r.pages, r.eng, "dev", "splash")
	r.pages.AddPage(accountsPage, p.root, true, true)
	r.app.SetFocus(p.root)
	if bar := untag(keyBar(accountsKeys, 200, 2)); !strings.Contains(bar, "x") || !strings.Contains(bar, "Delete saves") {
		t.Errorf("the panel's keys do not list deleting saves: %q", bar)
	}

	// The current account, and Cancel: nothing is deleted.
	p.selectByID(first.AccountID)
	r.key(tcell.KeyRune, 'x')
	if r.front() != accountsSavesConfirm {
		t.Fatalf("x opens %q, want the confirmation", r.front())
	}
	pressModal(t, r, 0)
	if savesOf(first.AccountID) != 2 {
		t.Fatal("Cancel deleted saves")
	}

	// The other account, confirmed: its saves go, the current one's stay.
	p.selectByID(second.AccountID)
	r.key(tcell.KeyRune, 'x')
	pressModal(t, r, 1)
	if savesOf(second.AccountID) != 0 || savesOf(first.AccountID) != 2 {
		t.Fatalf("deleting the second account's saves left %d and %d", savesOf(second.AccountID), savesOf(first.AccountID))
	}
	if status := untag(p.status.GetText(false)); !strings.Contains(status, "Deleted the 1 save of Second Player") {
		t.Errorf("the status line says %q", status)
	}
	if n := len(r.eng.ListAccounts()); n != 2 {
		t.Errorf("%d accounts left, want both", n)
	}

	// The current account, confirmed.
	p.selectByID(first.AccountID)
	r.key(tcell.KeyRune, 'X')
	pressModal(t, r, 1)
	if savesOf(first.AccountID) != 0 {
		t.Fatalf("the current account still has %d saves", savesOf(first.AccountID))
	}
	if status := untag(p.status.GetText(false)); !strings.Contains(status, "Deleted the 2 saves of Menu Player") {
		t.Errorf("the status line says %q", status)
	}
	if mainGame, last := first.GameRecord(); mainGame != "" || last != "" {
		t.Errorf("the records outlived the saves: %q and %q", mainGame, last)
	}
	if acct := r.eng.Account(); acct == nil || acct.AccountID != first.AccountID {
		t.Error("deleting the saves let go of the account")
	}
	// Once more: nothing left to delete, and it says so.
	r.key(tcell.KeyRune, 'x')
	pressModal(t, r, 1)
	if status := untag(p.status.GetText(false)); !strings.Contains(status, "no saves to delete") {
		t.Errorf("with nothing to delete the status line says %q", status)
	}

	// Back at the menu: an account with no game sees the first page.
	r.key(tcell.KeyEsc, 0)
	if r.front() != "splash" || r.m.hasCur || !r.m.view.forge || r.m.view.items[0].id != miNew {
		t.Errorf("back at the menu: front %q, a current game %v, forge page %v, first entry %q", r.front(), r.m.hasCur, r.m.view.forge, r.m.view.items[0].id)
	}
}
