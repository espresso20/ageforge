package ui

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/gdamore/tcell/v2"
)

// The recovery code (card G9Bx4UFM): the bare `account` reply is long and the
// log keeps its tail, so the code used to scroll out of view. It now has its
// own short command. Every test runs in a temp data root.

// logRows is the text inside the Log box, one string per drawn row.
func logRows(t *testing.T, d *Dashboard, w, h int) []string {
	t.Helper()
	d.refresh()
	cells := drawnDashboard(t, d, d.pages, w, h)
	x, y, bw, bh := d.logTV.GetInnerRect()
	return rectText(cells, w, x, y, bw, bh)
}

// squash joins the drawn rows, dropping blanks, so a code or a command that
// the box wrapped reads as it was written.
func squash(rows []string) string {
	return strings.Join(strings.Fields(strings.Join(rows, " ")), "")
}

func TestAccountShowCodeIsVisible(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		d, eng := mapTestDashboard(t, true)
		code := eng.Account().RecoveryCode()
		d.runForTest("account show code")
		if front, _ := d.pages.GetFrontPage(); front != codeWindowPage {
			t.Fatalf("front page %q, want the code window", front)
		}
		scr := screenAt(t, d.pages, sz[0], sz[1])
		if !strings.Contains(scr, code) {
			t.Errorf("%dx%d: the window does not show the whole code %q:\n%s", sz[0], sz[1], code, scr)
		}
	}
}

// TestAccountShowCodeIsNotKept: the code is shown in a window and nowhere
// else: not in the game log, the state snapshot, the Logs panel's text, a
// dump, or the bytes of a save written afterwards. Esc closes the window.
func TestAccountShowCodeIsNotKept(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	code := eng.Account().RecoveryCode()
	must(t, eng.StartNewNamedGame("Rome"))
	d.runForTest("account show code")
	d.refresh()
	for _, e := range eng.GetLogs() {
		if strings.Contains(e.Message, code) {
			t.Fatalf("the game log holds the code: %q", e.Message)
		}
	}
	for _, e := range eng.GetState().Log {
		if strings.Contains(e.Message, code) {
			t.Fatalf("the state snapshot's log holds the code: %q", e.Message)
		}
	}
	if strings.Contains(fmt.Sprintf("%+v", eng.GetState()), code) {
		t.Error("the state snapshot holds the code")
	}
	d.pages.RemovePage(codeWindowPage)
	if scr := screenAt(t, d.pages, 100, 30); strings.Contains(scr, code) {
		t.Errorf("the code is on screen after its window closed:\n%s", scr)
	}
	must(t, eng.SaveGame("after-code"))
	files, _ := filepath.Glob(filepath.Join(game.DataDir(), "saves", "*.json"))
	if len(files) == 0 {
		t.Fatal("no save was written")
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		must(t, err)
		if bytes.Contains(raw, []byte(code)) {
			t.Errorf("%s holds the code", f)
		}
	}
}

// TestAccountShowCodeWindowCloses: Esc takes the window away.
func TestAccountShowCodeWindowCloses(t *testing.T) {
	d, _ := mapTestDashboard(t, true)
	d.runForTest("account show code")
	pressKey(d, d.pages, tcell.KeyEscape, 0)
	if front, _ := d.pages.GetFrontPage(); front == codeWindowPage {
		t.Error("Esc left the code window open")
	}
}

func TestAccountShowCodeReply(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	code := eng.Account().RecoveryCode()
	res := HandleCommand("account show code", eng)
	if res.Type != "info" || res.Window != code || res.Message != "" {
		t.Errorf("reply = %+v, want the code in a window and nothing logged", res)
	}
	// Only the exact form shows it; anything else is the usage line.
	for _, line := range []string{"account show", "account show code anything", "account show codes", "account show code code", "account show nope"} {
		res := HandleCommand(line, eng)
		if res.Window != "" || res.Type != "error" || !strings.Contains(res.Message, "account show code") || strings.Contains(res.Message, code) {
			t.Errorf("%q = %+v, want the usage line", line, res)
		}
	}
}

func TestAccountShowCodeWithoutAccount(t *testing.T) {
	_, eng := mapTestDashboard(t, false)
	res := HandleCommand("account show code", eng)
	if res.Type != "warning" || !strings.Contains(res.Message, "Accounts are unavailable") {
		t.Errorf("reply = %q (%s), want the neighbors' no-account answer", res.Message, res.Type)
	}
}

// TestAccountBareNamesTheCommand: the bare reply no longer carries the code,
// and what stays in view at 80x24 and 100x30 names the command.
func TestAccountBareNamesTheCommand(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {100, 30}} {
		d, eng := mapTestDashboard(t, true)
		code := eng.Account().RecoveryCode()
		if msg := HandleCommand("account", eng).Message; strings.Contains(msg, code) {
			t.Fatalf("the bare account reply still prints the code")
		}
		d.runForTest("account")
		rows := logRows(t, d, sz[0], sz[1])
		if got := squash(rows); !strings.Contains(got, "accountshowcode") {
			t.Errorf("%dx%d: the visible log does not name account show code:\n%s", sz[0], sz[1], strings.Join(rows, "\n"))
		}
	}
}

func TestAccountShowCodeCompletes(t *testing.T) {
	if got := helpRow("account show code"); got.Form != "account show code" || got.Text == "" {
		t.Errorf("help row = %+v", got)
	}
	if lookup(lookup(registry(), "account").Subs, "show") == nil {
		t.Error("account has no show subcommand")
	}
}
