//go:build smoke

package ui

// UI smoke sweep, run by `make smoke` (go test -tags smoke -run TestSmokeUISweep ./ui).
//
// Boots the real App on a tcell SimulationScreen, with no dev mode, and under
// EVERY theme walks the screens a player can reach: the splash's theme
// picker, load game browser and accounts panel, then a new game and every
// panel in the sidebar list plus help and the other overlays, each opened by
// typing its command into the real input field. Under the default theme it
// also runs the harmless status/list commands. After every action it checks
// that the tview event loop still answers within a timeout (a freeze fails
// the step with the screen text and a goroutine dump). A panic anywhere
// kills the test binary, which fails the run.
//
// Nothing is written outside a temp data directory, and no screenshots are
// kept; a failing step prints the screen as text.

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// sweepTimeout bounds how long the event loop may take to answer after an action.
const sweepTimeout = 4 * time.Second

// sweepOverlays are the panel commands a player can type, and the overlay each
// should open. The first block mirrors the sidebar list (buildSidebarText).
var sweepOverlays = []struct{ cmd, overlay string }{
	{"milestones", "milestones"},
	{"research", "techs"},
	{"expedition", "expedition"},
	{"army", "army"},
	{"trade", "trade"},
	{"factions", "factions"},
	{"stats", "stats"},
	{"wonders", "wonders"},
	{"workers", "workers"},
	{"logs", "logs"},
	{"epoch", "epoch"},
	{"history", "history"},
	{"citymap", "citymap"},
	{"worldmap", "worldmap"},
	{"help", "help"},
	// Not in the sidebar, still reachable by command.
	{"buildings", "buildings"},
	{"diplomacy", "factions"},
	{"map", "map"},
	{"techs", "techs"},
}

// sweepCommands only read state (bare forms print a status or a list).
var sweepCommands = []string{
	"status", "rates", "build", "upgrade", "wonder", "festival", "blackmarket",
	"prestige", "prestige shop", "speed", "catastrophe", "theme", "theme list",
	"saves", "account",
}

type sweeper struct {
	*reproHarness
	step string
}

// screenText renders the simulated screen as text. The event loop draws into
// the same cells, so call it from the loop (liveScreen) unless the loop is
// already wedged, as in fail.
func (s *sweeper) screenText() string {
	cells, w, h := s.sim.GetContents()
	var sb strings.Builder
	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) == 0 {
				line.WriteByte(' ')
			} else {
				line.WriteRune(c.Runes[0])
			}
		}
		sb.WriteString(strings.TrimRight(line.String(), " "))
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n")
}

// liveScreen is screenText read on the event loop.
func (s *sweeper) liveScreen() string {
	var txt string
	s.ui(func() { txt = s.screenText() })
	return txt
}

func (s *sweeper) fail(format string, args ...interface{}) {
	s.t.Helper()
	s.t.Fatalf("step %q: %s\n\n=== SCREEN ===\n%s\n\n=== GOROUTINES ===\n%s",
		s.step, fmt.Sprintf(format, args...), s.screenText(), dumpGoroutines())
}

// ping proves the event loop is alive: a queued update must run in time.
func (s *sweeper) ping() {
	s.t.Helper()
	done := make(chan struct{})
	go s.a.tviewApp.QueueUpdateDraw(func() { close(done) })
	select {
	case <-done:
	case <-time.After(sweepTimeout):
		s.fail("event loop did not answer within %s (frozen?)", sweepTimeout)
	}
}

// ui runs f on the event loop, failing the step if the loop is wedged.
func (s *sweeper) ui(f func()) {
	s.t.Helper()
	done := make(chan struct{})
	go s.a.tviewApp.QueueUpdate(func() { f(); close(done) })
	select {
	case <-done:
	case <-time.After(sweepTimeout):
		s.fail("event loop did not answer within %s (frozen?)", sweepTimeout)
	}
}

func (s *sweeper) press(k tcell.Key, r rune) {
	s.t.Helper()
	done := make(chan struct{})
	go func() { s.sim.InjectKey(k, r, tcell.ModNone); close(done) }()
	select {
	case <-done:
	case <-time.After(sweepTimeout):
		s.fail("key injection blocked: event queue full, loop wedged")
	}
	time.Sleep(3 * time.Millisecond)
}

func (s *sweeper) wait(what string, cond func() bool) {
	s.t.Helper()
	deadline := time.Now().Add(sweepTimeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	s.fail("timed out waiting for %s\n%s", what, s.describeUI())
}

func (s *sweeper) front() string {
	var name string
	s.ui(func() { name, _ = s.a.pages.GetFrontPage() })
	return name
}

func (s *sweeper) activeOverlay() string {
	var name string
	s.ui(func() { name = s.a.dashboard.overlayMgr.ActiveName() })
	return name
}

func (s *sweeper) inputFocused() bool {
	var ok bool
	s.ui(func() { ok = s.a.tviewApp.GetFocus() == s.a.dashboard.inputField })
	return ok
}

func (s *sweeper) inputText() string {
	var txt string
	s.ui(func() { txt = s.a.dashboard.inputField.GetText() })
	return txt
}

// submit types cmd into the command input and presses Enter, taking a second
// Enter when the first only accepted an autocomplete suggestion.
func (s *sweeper) submit(cmd string) {
	s.t.Helper()
	s.wait("input focus", s.inputFocused)
	for _, r := range cmd {
		s.press(tcell.KeyRune, r)
	}
	s.wait("typed text", func() bool { return s.inputText() == cmd })
	s.press(tcell.KeyEnter, 0)
	s.wait("Enter processed", func() bool { t := s.inputText(); return t == "" || strings.HasPrefix(t, cmd) && t != cmd })
	if s.inputText() != "" {
		s.press(tcell.KeyEnter, 0)
		s.wait("command submitted", func() bool { return s.inputText() == "" })
	}
	s.ping()
}

// splashPage opens a splash menu entry by its shortcut and backs out with Esc.
func (s *sweeper) splashPage(key rune, page string, exercise func()) {
	s.t.Helper()
	s.step = fmt.Sprintf("splash '%c' -> %s", key, page)
	s.wait("splash in front", func() bool { return s.front() == "splash" })
	s.press(tcell.KeyRune, key)
	s.wait(page+" in front", func() bool { return s.front() == page })
	s.ping()
	if exercise != nil {
		exercise()
	}
	s.press(tcell.KeyEsc, 0)
	s.wait("back to splash", func() bool { return s.front() == "splash" })
	s.ping()
}

// bootSweeper starts the real App on a w x h simulated screen, in a temp
// data dir with a named account, and waits for the splash menu.
func bootSweeper(t *testing.T, w, h int) (*sweeper, *game.GameEngine) {
	t.Helper()
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	t.Cleanup(func() { _ = theme.SetActive(theme.DefaultKey) })
	acct, err := game.CreateNamedAccount("Smoke Sweep")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	eng := game.NewGameEngine()
	eng.SetAccount(acct)
	a := NewApp(eng, "dev") // "dev" skips the network update check
	sim := tcell.NewSimulationScreen("UTF-8")
	a.SetScreen(sim)
	sim.SetSize(w, h)
	hr := &reproHarness{t: t, eng: eng, a: a, sim: sim, runErr: make(chan error, 1)}
	go func() { hr.runErr <- a.Run() }()
	// Registered after SetDataDirForTest, so (cleanups run last in, first
	// out) it runs before the data dir is restored.
	t.Cleanup(hr.teardown)
	s := &sweeper{reproHarness: hr, step: "boot"}
	s.wait("splash menu", func() bool { return s.front() == "splash" })
	return s, eng
}

// sweepThemes picks the themes for the small-terminal sweep from
// SMOKE_UI_THEMES: "all", or a comma list of keys where "default" is the
// default theme and "light" the first light (Standard, non-default) one.
// Unset means default and light.
func sweepThemes(t *testing.T) []string {
	spec := os.Getenv("SMOKE_UI_THEMES")
	if spec == "" {
		spec = "default,light"
	}
	var out []string
	for _, k := range strings.Split(spec, ",") {
		switch k = strings.TrimSpace(k); k {
		case "all":
			for _, th := range theme.All() {
				out = append(out, th.Key)
			}
		case "default":
			out = append(out, theme.DefaultKey)
		case "light":
			out = append(out, "daylight")
		case "":
		default:
			out = append(out, k)
		}
	}
	for _, k := range out {
		found := false
		for _, th := range theme.All() {
			found = found || th.Key == k
		}
		if !found {
			t.Fatalf("SMOKE_UI_THEMES names unknown theme %q", k)
		}
	}
	return out
}

// sweepSizes parses SMOKE_UI_SIZES ("80x24,100x30"; unset means both).
func sweepSizes(t *testing.T) [][2]int {
	spec := os.Getenv("SMOKE_UI_SIZES")
	if spec == "" {
		spec = "80x24,100x30"
	}
	var out [][2]int
	for _, s := range strings.Split(spec, ",") {
		var w, h int
		if _, err := fmt.Sscanf(strings.TrimSpace(s), "%dx%d", &w, &h); err != nil || w <= 0 || h <= 0 {
			t.Fatalf("bad SMOKE_UI_SIZES entry %q", s)
		}
		out = append(out, [2]int{w, h})
	}
	return out
}

// TestSmokeUISmallTerminals opens the dashboard and every overlay on small
// terminals (SMOKE_UI_SIZES, default 80x24 and 100x30) under the themes in
// SMOKE_UI_THEMES, then resizes the live dashboard to each other size. A
// panic, a frozen event loop or a blank screen fails it.
func TestSmokeUISmallTerminals(t *testing.T) {
	sizes := sweepSizes(t)
	themes := sweepThemes(t)
	steps := 0
	for _, sz := range sizes {
		for _, key := range themes {
			sz, key := sz, key
			t.Run(fmt.Sprintf("%dx%d/%s", sz[0], sz[1], key), func(t *testing.T) {
				s, _ := bootSweeper(t, sz[0], sz[1])
				s.step = "switch theme to " + key
				s.ui(func() {
					if err := theme.SetActive(key); err != nil {
						t.Errorf("SetActive(%q): %v", key, err)
					}
				})
				s.step = fmt.Sprintf("new game at %dx%d", sz[0], sz[1])
				s.press(tcell.KeyRune, 'n')
				s.wait("new game name prompt", func() bool { return s.front() == newGameNamePage })
				s.press(tcell.KeyEnter, 0)
				s.wait("dashboard with input focus", func() bool { return s.front() == "dashboard" && s.inputFocused() })
				s.ping()
				if strings.TrimSpace(s.liveScreen()) == "" {
					s.fail("blank dashboard at %dx%d", sz[0], sz[1])
				}
				for _, o := range sweepOverlays {
					s.step = fmt.Sprintf("[%dx%d %s] %s", sz[0], sz[1], key, o.cmd)
					s.submit(o.cmd)
					s.wait("overlay "+o.overlay, func() bool { return s.activeOverlay() == o.overlay })
					s.ping()
					if strings.TrimSpace(s.liveScreen()) == "" {
						s.fail("blank screen with overlay %s open", o.overlay)
					}
					s.press(tcell.KeyEsc, 0)
					s.wait("overlay closed", func() bool { return s.activeOverlay() == "" && s.inputFocused() })
					steps++
				}
				for _, other := range append(sizes, [2]int{sz[0], sz[1]}) {
					other := other
					s.step = fmt.Sprintf("resize to %dx%d", other[0], other[1])
					// The simulation screen resizes silently; post the event a
					// real terminal would send.
					s.sim.SetSize(other[0], other[1])
					if err := s.sim.PostEvent(tcell.NewEventResize(other[0], other[1])); err != nil {
						s.fail("posting the resize event: %v", err)
					}
					time.Sleep(120 * time.Millisecond) // tview throttles resize redraws
					s.ping()
					if strings.TrimSpace(s.liveScreen()) == "" {
						s.fail("blank dashboard after resizing to %dx%d", other[0], other[1])
					}
				}
				for _, l := range s.eng.GetLogs() {
					if strings.Contains(l.Message, "recovered from panic") {
						t.Errorf("engine recovered a panic: %s", l.Message)
					}
				}
			})
		}
	}
	t.Logf("small terminals: %d size(s) x %d theme(s), %d overlay steps", len(sizes), len(themes), steps)
}

func TestSmokeUISweep(t *testing.T) {
	s, eng := bootSweeper(t, 180, 56)
	h := s.reproHarness

	steps := 0
	for i, th := range theme.All() {
		key := th.Key
		s.step = "switch theme to " + key
		s.ui(func() {
			if err := theme.SetActive(key); err != nil {
				t.Errorf("SetActive(%q): %v", key, err)
			}
		})
		s.ping()

		s.splashPage('t', themePickerPage, func() {
			// Arrowing previews each highlighted theme live; Esc must revert.
			for j := 0; j < 3; j++ {
				s.press(tcell.KeyDown, 0)
				s.ping()
			}
		})
		if got := theme.Active().Key; got != key {
			t.Errorf("theme picker Esc left theme %q active, want %q", got, key)
		}
		s.splashPage('l', loadGamePage, nil)
		s.splashPage('a', accountsPage, nil)

		s.step = "new game (" + key + ")"
		s.press(tcell.KeyRune, 'n')
		s.wait("new game name prompt", func() bool { return s.front() == newGameNamePage })
		s.press(tcell.KeyEnter, 0)
		s.wait("dashboard with input focus", func() bool { return s.front() == "dashboard" && s.inputFocused() })
		s.ping()

		for _, o := range sweepOverlays {
			s.step = fmt.Sprintf("[%s] %s", key, o.cmd)
			s.submit(o.cmd)
			s.wait("overlay "+o.overlay, func() bool { return s.activeOverlay() == o.overlay })
			s.ping()
			if strings.TrimSpace(s.liveScreen()) == "" {
				s.fail("blank screen with overlay %s open", o.overlay)
			}
			s.press(tcell.KeyEsc, 0)
			s.wait("overlay closed", func() bool { return s.activeOverlay() == "" && s.inputFocused() })
			steps++
		}
		if i == 0 {
			for _, cmd := range sweepCommands {
				s.step = fmt.Sprintf("[%s] %s", key, cmd)
				s.submit(cmd)
				if front := s.front(); front != "dashboard" {
					// Some bare commands open a page (bare `theme` opens the
					// picker); it must close with Esc like any other.
					s.press(tcell.KeyEsc, 0)
					s.wait("back to dashboard from "+front, func() bool { return s.front() == "dashboard" && s.inputFocused() })
				}
				if ov := s.activeOverlay(); ov != "" {
					s.press(tcell.KeyEsc, 0)
					s.wait("overlay closed", func() bool { return s.activeOverlay() == "" })
				}
				steps++
			}
		}

		// Esc on the dashboard saves, stops the engine and returns to the splash.
		s.step = "dashboard -> splash (" + key + ")"
		s.press(tcell.KeyEsc, 0)
		s.wait("splash after Esc", func() bool { return s.front() == "splash" })
		s.ping()
	}

	for _, l := range eng.GetLogs() {
		if strings.Contains(l.Message, "recovered from panic") {
			t.Errorf("engine recovered a panic during the sweep: %s", l.Message)
		}
	}
	select {
	case err := <-h.runErr:
		t.Fatalf("app exited during the sweep: %v", err)
	default:
	}
	t.Logf("UI sweep: %d themes, %d overlay/command steps", len(theme.All()), steps)
}
