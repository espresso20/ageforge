package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/nerdfont"
)

// fakeInstaller counts its runs, reports one progress line and returns a
// canned outcome. With release set it waits on it before returning, so a
// test can look at the window mid-install.
type fakeInstaller struct {
	ran      *int
	res      nerdfont.Result
	err      error
	progress func(string)
	release  chan struct{}
}

func (f fakeInstaller) Install(context.Context) (nerdfont.Result, error) {
	*f.ran++
	f.progress("Downloading... 50%")
	if f.release != nil {
		<-f.release
	}
	return f.res, f.err
}

type iconsHarness struct {
	flow    *iconsFlow
	logs    []string
	nerd    int
	install int
}

// newIconsHarness is a flow whose install runs inline, as if the UI
// goroutine and the install goroutine took turns.
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

// shown is the window's text and button captions as the player reads them.
func (h *iconsHarness) shown() string {
	body, keys := h.flow.view()
	parts := []string{guardRendered(body)}
	for _, k := range keys {
		parts = append(parts, guardRendered(k.caption()))
	}
	return strings.Join(parts, "\n")
}

func (h *iconsHarness) logged(s string) bool {
	for _, l := range h.logs {
		if strings.Contains(l, s) {
			return true
		}
	}
	return false
}

// press presses the button with key k (0: Close).
func (h *iconsHarness) press(t *testing.T, k rune) {
	t.Helper()
	_, keys := h.flow.view()
	for _, b := range keys {
		if b.key == k {
			b.press()
			return
		}
	}
	t.Fatalf("no button %q at stage %d:\n%s", k, h.flow.stage, h.shown())
}

func TestIconsFlowSeesIcons(t *testing.T) {
	h := newIconsHarness(nil, nerdfont.Result{}, nil)
	h.flow.start()
	txt := h.shown()
	nerd, uni := iconSamples()
	for _, want := range []string{nerd, uni, "[I] I see icons", "[B] I see boxes", "[Esc] Close"} {
		if want == "" || !strings.Contains(txt, want) {
			t.Errorf("the check lacks %q:\n%s", want, txt)
		}
	}
	h.press(t, 'i')
	if h.nerd != 1 || h.flow.stage != iconsHaveIcons || h.install != 0 {
		t.Errorf("after I see icons: nerd set %d times, stage %d, installs %d", h.nerd, h.flow.stage, h.install)
	}
	if !strings.Contains(h.shown(), "the map now draws them") {
		t.Errorf("no confirmation:\n%s", h.shown())
	}
}

func TestIconsFlowInstalls(t *testing.T) {
	h := newIconsHarness([]string{"TERM_PROGRAM=iTerm.app", "HOME=/x"}, nerdfont.Result{Dir: "/x/Library/Fonts", Files: []string{"a.ttf"}}, nil)
	h.flow.start()
	h.press(t, 'b')
	if txt := h.shown(); !strings.Contains(txt, "Install JetBrains Mono Nerd Font "+nerdfont.Version) || !strings.Contains(txt, "[Y] Install") {
		t.Fatalf("no install offer:\n%s", txt)
	}
	h.press(t, 'y')
	if h.install != 1 || h.flow.stage != iconsInstalled {
		t.Fatalf("installer ran %d times, stage %d", h.install, h.flow.stage)
	}
	txt := h.shown()
	for _, want := range []string{"is installed", "In iTerm2:", "1. Open iTerm2", nerdfont.FamilyName, "Restart your terminal, then type icons again to check."} {
		if !strings.Contains(txt, want) {
			t.Errorf("the window lacks %q:\n%s", want, txt)
		}
	}
	if h.nerd != 0 {
		t.Errorf("the install set the glyphs (%d)", h.nerd)
	}
}

func TestIconsFlowInstallFails(t *testing.T) {
	h := newIconsHarness(nil, nerdfont.Result{}, errors.New("the download could not start (are you online?)"))
	h.flow.start()
	h.press(t, 'b')
	h.press(t, 'y')
	txt := h.shown()
	if !strings.Contains(txt, "The font was not installed: the download could not start") || !strings.Contains(txt, nerdFontsPage) || !strings.Contains(txt, "[R] Try again") {
		t.Errorf("no failure report:\n%s", txt)
	}
	if strings.Contains(txt, "Restart your terminal") {
		t.Error("told to restart after a failed install")
	}
	h.press(t, 'r')
	if h.install != 2 {
		t.Errorf("try again ran the installer %d times in all, want 2", h.install)
	}
}

func TestIconsFlowDecline(t *testing.T) {
	h := newIconsHarness([]string{"WT_SESSION=abc"}, nerdfont.Result{}, nil)
	h.flow.start()
	h.press(t, 'b')
	h.press(t, 'n')
	txt := h.shown()
	if h.install != 0 || !strings.Contains(txt, "Nothing installed") || !strings.Contains(txt, "In Windows Terminal:") || !strings.Contains(txt, nerdFontsPage) {
		t.Errorf("decline:\n%s", txt)
	}
	h.press(t, 0)
	if h.flow.open {
		t.Error("Close left the window open")
	}
	h.flow.start()
	if h.flow.stage != iconsAskSee {
		t.Errorf("icons again opened on stage %d, want a fresh check", h.flow.stage)
	}
}

// The install runs off the UI goroutine: its progress and its outcome come
// back through post, and the window shows each.
func TestIconsInstallRunsOffTheUIGoroutine(t *testing.T) {
	h := newIconsHarness([]string{"KITTY_WINDOW_ID=1"}, nerdfont.Result{}, nil)
	release := make(chan struct{})
	posted := make(chan func(), 4)
	h.flow.installer = func(progress func(string)) fontInstaller {
		return fakeInstaller{ran: &h.install, progress: progress, release: release}
	}
	h.flow.run = func(f func()) { go f() }
	h.flow.post = func(f func()) { posted <- f }

	h.flow.start()
	h.press(t, 'b')
	h.press(t, 'y')
	if h.flow.stage != iconsInstalling {
		t.Fatalf("stage %d after Install, want installing", h.flow.stage)
	}
	(<-posted)() // the progress line, run on "the UI goroutine"
	if txt := h.shown(); !strings.Contains(txt, "Downloading... 50%") || !strings.Contains(txt, "keep playing") {
		t.Errorf("no progress in the window:\n%s", txt)
	}
	close(release)
	(<-posted)() // the outcome
	if h.flow.stage != iconsInstalled || !strings.Contains(h.shown(), "In kitty:") {
		t.Errorf("after the install: stage %d\n%s", h.flow.stage, h.shown())
	}
}

// Closing the window mid-install lets the install finish; the log says so,
// and icons then opens on the outcome once before checking afresh.
func TestIconsWindowClosedDuringInstall(t *testing.T) {
	h := newIconsHarness(nil, nerdfont.Result{}, nil)
	release := make(chan struct{})
	posted := make(chan func(), 4)
	h.flow.installer = func(progress func(string)) fontInstaller {
		return fakeInstaller{ran: &h.install, progress: progress, release: release}
	}
	h.flow.run = func(f func()) { go f() }
	h.flow.post = func(f func()) { posted <- f }

	h.flow.start()
	h.press(t, 'b')
	h.press(t, 'y')
	h.press(t, 0) // Close
	(<-posted)()
	close(release)
	(<-posted)()
	if h.flow.open || !h.logged("success: JetBrains Mono Nerd Font is installed. Type icons") {
		t.Fatalf("open %v, log %q", h.flow.open, h.logs)
	}
	h.flow.start()
	if h.flow.stage != iconsInstalled || !strings.Contains(h.shown(), "Restart your terminal") {
		t.Errorf("icons after the install opened on stage %d:\n%s", h.flow.stage, h.shown())
	}
	h.press(t, 0)
	h.flow.start()
	if h.flow.stage != iconsAskSee {
		t.Errorf("the next icons opened on stage %d, want a fresh check", h.flow.stage)
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

// iconsTestDashboard is a dashboard on a fresh game with a named account,
// whose icons check installs through in, inline.
func iconsTestDashboard(t *testing.T, env []string, in func(progress func(string)) fontInstaller) (*Dashboard, *game.GameEngine, *tview.Pages) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	eng := game.NewGameEngine()
	acct, err := game.CreateNamedAccount("Icons Tester")
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(acct)
	pages := tview.NewPages()
	d := NewDashboard(tview.NewApplication(), eng, pages)
	pages.AddPage("dashboard", d.Root(), true, true)
	d.icons.installer = in
	d.icons.environ = func() []string { return env }
	d.icons.goos = "darwin"
	d.icons.run = func(f func()) { f() }
	d.icons.post = func(f func()) { f() }
	return d, eng, pages
}

// pressKey sends a key to whatever holds the keyboard, as the event loop does.
func pressKey(d *Dashboard, pages *tview.Pages, k tcell.Key, r rune) {
	pages.InputHandler()(tcell.NewEventKey(k, r, tcell.ModNone), func(p tview.Primitive) { d.app.SetFocus(p) })
}

// screenAt renders pages at w×h on a SimulationScreen.
func screenAt(t *testing.T, pages *tview.Pages, w, h int) string {
	t.Helper()
	var rows []string
	for _, r := range screenGrid(t, pages, w, h) {
		rows = append(rows, strings.TrimRight(string(r), " "))
	}
	return strings.Join(rows, "\n")
}

// TestIconsWindowOnScreen: icons at the prompt opens the window over the
// dashboard; its own keys walk the check through a (fake) install to the
// terminal steps, and Esc closes it back to the prompt, not to the menu.
func TestIconsWindowOnScreen(t *testing.T) {
	installs := 0
	d, eng, pages := iconsTestDashboard(t, []string{"TERM_PROGRAM=iTerm.app"}, func(progress func(string)) fontInstaller {
		return fakeInstaller{ran: &installs, res: nerdfont.Result{Dir: "/x"}, progress: progress}
	})
	d.inputField.SetText("icons")
	d.submitInput()
	if front, _ := pages.GetFrontPage(); front != iconsPage {
		t.Fatalf("front page %q after icons, want the icons window", front)
	}
	scr := screenAt(t, pages, 80, 24)
	for _, want := range []string{" Icons check ", "[I] I see icons", "[B] I see boxes", "[Esc] Close", "Nerd Font icons"} {
		if !strings.Contains(scr, want) {
			t.Errorf("80x24 screen lacks %q:\n%s", want, scr)
		}
	}
	if strings.Contains(scr, "[accent]") || strings.Contains(scr, "[text]") {
		t.Errorf("a color tag leaked onto the screen:\n%s", scr)
	}

	pressKey(d, pages, tcell.KeyRune, 'B')
	if scr := screenAt(t, pages, 100, 30); !strings.Contains(scr, "[Y] Install") || !strings.Contains(scr, "No, show me how") {
		t.Fatalf("no install offer after B:\n%s", scr)
	}
	pressKey(d, pages, tcell.KeyRune, 'x') // not a button: nothing happens
	if d.icons.stage != iconsAskInstall {
		t.Fatalf("a stray key moved the check to stage %d", d.icons.stage)
	}
	pressKey(d, pages, tcell.KeyTab, 0)     // focus No, show me how
	pressKey(d, pages, tcell.KeyBacktab, 0) // back to Install
	pressKey(d, pages, tcell.KeyEnter, 0)   // press it
	if installs != 1 {
		t.Fatalf("Enter on Install ran the installer %d times", installs)
	}
	scr = screenAt(t, pages, 100, 30)
	for _, want := range []string{"is installed", "In iTerm2:", "Restart your terminal, then type icons again to check."} {
		if !strings.Contains(scr, want) {
			t.Errorf("screen lacks %q:\n%s", want, scr)
		}
	}

	pressKey(d, pages, tcell.KeyEsc, 0)
	if pages.HasPage(iconsPage) || !d.inputField.HasFocus() {
		t.Errorf("after Esc: window still there %v, prompt focused %v", pages.HasPage(iconsPage), d.inputField.HasFocus())
	}
	if front, _ := pages.GetFrontPage(); front != "dashboard" {
		t.Errorf("Esc in the window left front page %q", front)
	}

	// The answer in the window sets the account's glyphs.
	d.startIcons()
	pressKey(d, pages, tcell.KeyRune, 'i')
	if _, g, _ := eng.Account().MapPrefs(); g != "nerd" {
		t.Errorf("account glyphs %q after I see icons", g)
	}
}

// A panel that opens over the window and closes again hands the keyboard
// back to the window, not to the prompt under it (where Esc would leave the
// game).
func TestIconsWindowGetsTheKeyboardBack(t *testing.T) {
	d, eng, pages := iconsTestDashboard(t, nil, nil)
	d.startIcons()
	if !d.overlayMgr.Show("help", eng.GetState()) {
		t.Fatal("no help panel")
	}
	d.overlayMgr.Hide()
	if front, _ := pages.GetFrontPage(); front != iconsPage || !d.iconsWin.buttons[d.iconsWin.focus].HasFocus() {
		t.Fatalf("after a panel closed over it: front %q, window focused %v", front, d.iconsWin.buttons[0].HasFocus())
	}
	pressKey(d, pages, tcell.KeyEsc, 0)
	if pages.HasPage(iconsPage) || !d.inputField.HasFocus() {
		t.Error("Esc did not close the window back to the prompt")
	}
}

// The Ancient Memory offer gives the keyboard back when it closes: to the
// prompt, not to a button that is no longer on screen (where every key was
// lost until something else took focus).
func TestAncientMemoryModalGivesTheKeyboardBack(t *testing.T) {
	d, _, pages := iconsTestDashboard(t, nil, nil)
	d.showAncientMemoryModal("tool_making", "Tool Making")
	pressKey(d, pages, tcell.KeyRune, 'd') // Decline
	if pages.HasPage(ancientMemoryPage) || !d.inputField.HasFocus() {
		t.Errorf("after Decline: modal still there %v, prompt focused %v", pages.HasPage(ancientMemoryPage), d.inputField.HasFocus())
	}
}
