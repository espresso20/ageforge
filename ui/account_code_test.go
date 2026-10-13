package ui

import (
	"strings"
	"testing"
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
		rows := logRows(t, d, sz[0], sz[1])
		if got := squash(rows); !strings.Contains(got, code) {
			t.Errorf("%dx%d: the log does not show the whole code %q:\n%s", sz[0], sz[1], code, strings.Join(rows, "\n"))
		}
	}
}

func TestAccountShowCodeReply(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	code := eng.Account().RecoveryCode()
	res := HandleCommand("account show code", eng)
	if res.Type != "info" || res.Message != code {
		t.Errorf("reply = %q (%s), want the code and nothing else", res.Message, res.Type)
	}
	if res := HandleCommand("account show", eng); res.Type != "error" || !strings.Contains(res.Message, "account show code") {
		t.Errorf("account show = %q (%s), want the usage", res.Message, res.Type)
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
