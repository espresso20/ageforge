package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// App is the running game's UI: the loop, the pages and the dashboard.
type App struct {
	loop      *appLoop
	pages     *nativePages
	engine    *game.GameEngine
	dashboard *Dashboard
	version   string
	menu      *mainMenu
}

// NewApp creates the UI application
func NewApp(engine *game.GameEngine, version string) *App {
	// Assert the active theme before any widget is built so the name-remap and
	// chrome defaults are in place for the first Draw, and every
	// theme.Track closure in setup() applies against a known palette. The theme
	// package init() already seeds Forge defensively; this makes the boot explicit
	// and survives a future import reshuffle. Phase 2 swaps DefaultKey for the
	// account's stored active theme (the theming design §6). Ignoring the error is fine —
	// DefaultKey is a registered built-in.
	_ = theme.SetActive(theme.DefaultKey)

	a := &App{
		loop:    newAppLoop(),
		pages:   newNativePages(),
		engine:  engine,
		version: version,
	}
	a.setup()
	return a
}

func (a *App) setup() {
	// Apply the account's persisted active theme (the theming design §6) before any widget
	// is built, so the splash and every theme.Track closure in the page constructors
	// below render in the chosen theme on the very first Draw. applyAccountTheme is
	// fully defensive: no engine/account, an empty stored key, or an unknown/locked
	// one all fall back to the default (Forge). NewApp already pinned Forge as a
	// floor; this layers the account's choice on top once the engine is in place.
	applyAccountTheme(a.engine)

	a.dashboard = NewDashboard(a.loop, a.engine, a.pages)
	// The arrival screen draws the player's town in their glyph set and theme: it
	// reads the game and the account's settings through the engine.
	a.dashboard.overlayMgr.engine = a.engine
	// Esc reaches the dashboard before it reaches an overlay, and the
	// dashboard answers it by closing the overlay at once. On the arrival
	// screen Esc is a key like any other (the first moves on to what the age
	// opens, the next closes), so that screen is offered it first. Before
	// either, Option+arrows become PgUp, PgDn, Home and End (mac_keys.go), so
	// every screen and panel sees keys a MacBook does not have.
	a.loop.SetInputCapture(func(k *kit.Key) *kit.Key {
		k = macKeys(k)
		if k.Code == kit.KeyEsc && a.dashboard.overlayMgr.arrivalKey() {
			return nil
		}
		return k
	})

	splash := CreateSplashPage(a.loop, a.pages, a.engine, a.version)
	a.menu, _ = splash.(*mainMenu)

	a.pages.AddPage("splash", splash, true, true)
	a.pages.AddPage("dashboard", a.dashboard.Root(), true, false)

	a.loop.SetRoot(a.pages, true)

	// First run: if no established (named) account is wired, prompt for the account
	// name as the very first thing the player sees — over the splash menu. The name
	// derives the identity (game.CreateAccount); the modal's Esc-accepts policy
	// guarantees we always come away with an account. Once confirmed, install it and
	// reveal the splash beneath.
	if acct := a.engine.Account(); acct == nil || !acct.Established() {
		showAccountNameModal(a.loop, a.pages, splash, func(name string) {
			newAcct, err := game.CreateAccount(name)
			if err != nil {
				// Account is non-critical: say why and fall through to the splash
				// accountless rather than block play.
				showUpdateMsg(a.loop, a.pages, err.Error())
				return
			}
			a.engine.SetAccount(newAcct)
			// A brand-new account starts on its own theme (ActiveTheme "" → Forge)
			// rather than inheriting whatever was previewed before it was installed
			// (the theming design §6). Re-resolve from the freshly-wired account.
			applyAccountTheme(a.engine)
		})
	}
}

// OnDiscordChange sets what is called when the player changes the Discord status setting,
// so the status can follow at once.
func (a *App) OnDiscordChange(fn func()) {
	if a.menu != nil {
		a.menu.onDiscord = fn
	}
}

// WindowHost is the window the game is drawn in, when there is one: the Settings
// page offers the font choice only then.
type WindowHost interface {
	FontChoice() string // "mono" or "pixel"
	NextFont()          // switch to the other font
	// CellSize is a cell of the grid in pixels (0, 0 until the window is
	// laid out), for pictures drawn pixel for pixel.
	CellSize() (w, h int)
}

// pictureCell is the window's cell size in pixels; nil when no window is open.
var pictureCell func() (w, h int)

// cellPixels is the size in pixels of a cell of the grid: the window's, or
// a screenshot's cell when no window is open (tests, recorded screens).
func cellPixels() (w, h int) {
	if pictureCell != nil {
		if w, h := pictureCell(); w > 0 && h > 0 {
			return w, h
		}
	}
	return shotCellW, shotCellH
}

// Resize, Key and Mouse are the window's events: the grid's size in cells,
// a key press, and the pointer at a cell. They are safe to call from the
// window's thread and never wait for the UI.
func (a *App) Resize(cols, rows int) { a.loop.Resize(cols, rows) }
func (a *App) Key(k kit.Key)         { a.loop.Key(k) }
func (a *App) Mouse(m kit.Mouse)     { a.loop.Mouse(m) }

// Frame copies the cells the game shows now into dst, for the window to draw.
func (a *App) Frame(dst *kit.Buffer) { a.loop.Frame(dst) }

// SetWindowHost tells the app it runs in a window (see WindowHost).
func (a *App) SetWindowHost(h WindowHost) {
	if a.menu != nil {
		a.menu.window = h
	}
	pictureCell = h.CellSize
}

// SetSpaceHost gives the Map panel the window's own sky (see mapstyle.SpaceHost).
// Without one space is drawn in cells.
func (a *App) SetSpaceHost(h mapstyle.SpaceHost) {
	a.dashboard.mapPanel.space = h
}

// Run runs the UI until Stop (it blocks).
func (a *App) Run() error {
	a.dashboard.StartUpdates()
	defer a.dashboard.StopUpdates()
	return a.loop.Run()
}

// Stop ends Run.
func (a *App) Stop() { a.loop.Stop() }
