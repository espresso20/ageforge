package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/pkg/nerdfont"
)

// fakeInstaller records that it ran and returns a canned outcome.
type fakeInstaller struct {
	ran      *int
	res      nerdfont.Result
	err      error
	progress func(string)
}

func (f fakeInstaller) Install(context.Context) (nerdfont.Result, error) {
	*f.ran++
	f.progress("Downloading... 50%")
	return f.res, f.err
}

type iconsHarness struct {
	flow    *iconsFlow
	logs    []string
	nerd    int
	install int
}

func newIconsHarness(env []string, res nerdfont.Result, err error) *iconsHarness {
	h := &iconsHarness{}
	h.flow = &iconsFlow{
		log:     func(kind, msg string) { h.logs = append(h.logs, kind+": "+msg) },
		setNerd: func() error { h.nerd++; return nil },
		installer: func(progress func(string)) fontInstaller {
			return fakeInstaller{ran: &h.install, res: res, err: err, progress: progress}
		},
		environ: func() []string { return env },
		goos:    "darwin",
		run:     func(f func()) { f() },
		post:    func(f func()) { f() },
	}
	return h
}

func (h *iconsHarness) logged(s string) bool {
	for _, l := range h.logs {
		if strings.Contains(l, s) {
			return true
		}
	}
	return false
}

func TestIconsFlowSeesIcons(t *testing.T) {
	h := newIconsHarness(nil, nerdfont.Result{}, nil)
	h.flow.start()
	if !h.logged(iconsQuestion) || !h.flow.waiting() {
		t.Fatalf("no question asked: %v", h.logs)
	}
	nerd, uni := iconSamples()
	if !h.logged(nerd) || !h.logged(uni) || nerd == "" || uni == "" {
		t.Errorf("samples missing: %v", h.logs)
	}
	if !h.flow.answer("Icons") {
		t.Fatal("icons was not taken as an answer")
	}
	if h.nerd != 1 || h.flow.waiting() || h.install != 0 {
		t.Errorf("after icons: nerd set %d times, waiting %v, installs %d", h.nerd, h.flow.waiting(), h.install)
	}
}

func TestIconsFlowInstalls(t *testing.T) {
	h := newIconsHarness([]string{"TERM_PROGRAM=iTerm.app", "HOME=/x"}, nerdfont.Result{Dir: "/x/Library/Fonts", Files: []string{"a.ttf"}}, nil)
	h.flow.start()
	h.flow.answer("boxes")
	if !h.logged(installPrompt) {
		t.Fatalf("no install question: %v", h.logs)
	}
	if !h.flow.answer("y") {
		t.Fatal("y was not an answer")
	}
	if h.install != 1 {
		t.Fatalf("installer ran %d times", h.install)
	}
	for _, want := range []string{"Downloading... 50%", "is installed", "In iTerm2:", "1. Open iTerm2", nerdfont.FamilyName, restartMessage} {
		if !h.logged(want) {
			t.Errorf("log lacks %q:\n%s", want, strings.Join(h.logs, "\n"))
		}
	}
	if h.nerd != 0 || h.flow.waiting() {
		t.Errorf("the install set glyphs (%d) or left a question open", h.nerd)
	}
}

func TestIconsFlowInstallFails(t *testing.T) {
	h := newIconsHarness(nil, nerdfont.Result{}, errors.New("the download could not start (are you online?)"))
	h.flow.start()
	h.flow.answer("boxes")
	h.flow.answer("yes")
	if !h.logged("error: The font was not installed") || !h.logged(nerdFontsPage) {
		t.Errorf("no failure report:\n%s", strings.Join(h.logs, "\n"))
	}
	if h.logged(restartMessage) {
		t.Error("told to restart after a failed install")
	}
	if h.flow.stage != iconsIdle {
		t.Errorf("stage %v after a failure", h.flow.stage)
	}
}

func TestIconsFlowDeclineAndOtherInput(t *testing.T) {
	h := newIconsHarness([]string{"WT_SESSION=abc"}, nerdfont.Result{}, nil)
	h.flow.start()
	h.flow.answer("b")
	h.flow.answer("n")
	if h.install != 0 || !h.logged("Nothing installed") || !h.logged("In Windows Terminal:") {
		t.Errorf("decline:\n%s", strings.Join(h.logs, "\n"))
	}
	h.flow.start()
	if h.flow.answer("build hut") {
		t.Error("a command was swallowed as an answer")
	}
	if h.flow.waiting() || !h.logged("Icons check closed") {
		t.Error("the check did not close on a command")
	}
}

func TestTerminalFontSteps(t *testing.T) {
	for _, c := range []struct {
		env  []string
		goos string
		want string
	}{
		{[]string{"TERM_PROGRAM=Apple_Terminal"}, "darwin", "Terminal"},
		{[]string{"TERM_PROGRAM=iTerm.app"}, "darwin", "iTerm2"},
		{[]string{"TERM_PROGRAM=vscode"}, "linux", "the VS Code terminal"},
		{[]string{"TERM_PROGRAM=WezTerm"}, "linux", "WezTerm"},
		{[]string{"WT_SESSION=1", "TERM_PROGRAM=vscode"}, "windows", "Windows Terminal"},
		{[]string{"KITTY_WINDOW_ID=1"}, "linux", "kitty"},
		{[]string{"ALACRITTY_SOCKET=/tmp/a"}, "linux", "Alacritty"},
		{[]string{"KONSOLE_VERSION=230400"}, "linux", "Konsole"},
		{[]string{"GNOME_TERMINAL_SCREEN=/org/x"}, "linux", "GNOME Terminal"},
		{nil, "linux", "your terminal"},
		{nil, "windows", "your terminal"},
	} {
		name, steps := terminalFontSteps(c.env, c.goos)
		if name != c.want {
			t.Errorf("%v: %q, want %q", c.env, name, c.want)
		}
		if len(steps) < 2 || len(steps) > 3 {
			t.Errorf("%s: %d steps", name, len(steps))
		}
		if !strings.Contains(strings.Join(steps, " "), nerdfont.FamilyName) {
			t.Errorf("%s: the steps never name the font", name)
		}
	}
}

// TestIconsCommandOnDashboard: icons at the prompt starts the check, and
// the answer at the prompt sets the account's glyphs.
func TestIconsCommandOnDashboard(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	d.runForTest("icons")
	if !d.icons.waiting() {
		t.Fatal("icons did not ask")
	}
	d.runForTest("icons")
	if _, g, _ := eng.Account().MapPrefs(); g != "nerd" {
		t.Errorf("account glyphs %q after answering icons", g)
	}
	if d.inputField.GetText() != "" {
		t.Errorf("the answer stayed in the prompt")
	}
}
