package ui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// menu.go is the main menu: the page the game opens on and returns to.
//
// An account with a game sees its own town behind the menu (the town
// page); an account without one sees the forge and the title page.
// menu_page.go draws both, menu_scene.go what moves on them, menu_town.go
// the town.
//
// The menu reads and never writes. Which game is current comes from the
// account's records (game.CurrentGame); the town is drawn from a look at
// that save (game.ViewSave), which loads nothing into the engine in play.
// It reads again each time it is returned to, and only when something it
// read from has changed.
//
// Its motion runs at the map's rate, from the clock, and only while the
// menu is the page in use: opening a page or a game stops it. With the
// motion setting off no clock runs and the page holds one still frame.

// eliteLines are shown on the menu to an account that holds a save with a
// forge master's proof: a mark, the words, a mark.
var eliteLines = [][3]string{
	{"⚡", "MASTER FORGER", "⚡"},
	{"{", "TOUCHED BY THE SOURCE", "}"},
	{"✦", "ARCHITECT OF THE FORGE", "✦"},
	{"<", "REALITY.EXE PATCHED", ">"},
	{"⚙", "THE FIRST MAKER", "⚙"},
}

// menuStrikeHold is how long a choice waits for its strike to be drawn
// before it acts: under one animation frame, so the flare is seen and the
// wait is not felt. With motion off there is no strike and no wait.
const menuStrikeHold = 90 * time.Millisecond

// menuCatchUp is the most animation frames one draw plays to catch the
// clock up; a longer gap is skipped, not replayed.
const menuCatchUp = 4

// mainMenu is the main menu's page.
type mainMenu struct {
	*tview.Box
	app     *tview.Application
	pages   *tview.Pages
	engine  *game.GameEngine
	version string
	reg     *mapstyle.Registry

	// What the page says, read by refresh.
	view   menuView
	cur    game.CurrentGame
	hasCur bool
	// contDetails and contExtras are the Continue entry's line and the
	// save's name beside it.
	contDetails, contExtras []string
	town                    *menuTown
	set                     mapSettings
	acct                    *game.Account
	read                    string // what refresh last read from (fingerprint)
	eliteMsg                [3]string
	update                  bool // a newer version is out

	// The scene, and the clock it runs on.
	scene     *menuScene
	sceneKey  [4]int
	pal       *menuPalette
	now       func() time.Time
	start     time.Time
	baseFrame int

	// active: the menu is the page in use (not under a page it opened).
	// stop ends the redraw clock; nil when none runs.
	active bool
	stop   chan struct{}
	// pending: a choice is waiting out its strike.
	pending bool
	// later runs fn after the strike's hold (tests run it at once).
	later func(fn func())
	// checkUpdates opens the update check (it asks the network).
	checkUpdates func()
}

// CreateSplashPage creates the main menu. A release build also asks, in
// the background, whether a newer version is out.
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string) tview.Primitive {
	m := newMainMenu(app, pages, engine, currentVersion)
	// Background update check: a no-op on dev builds or network errors.
	if currentVersion != "dev" {
		go func() {
			result, err := game.CheckLatest(currentVersion)
			if err != nil || !result.IsNewer {
				return
			}
			app.QueueUpdateDraw(func() {
				m.update = true
				m.buildView()
			})
		}()
	}
	return m
}

// newMainMenu makes the menu and reads what it says. It asks the network
// nothing.
func newMainMenu(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string) *mainMenu {
	m := &mainMenu{
		Box: tview.NewBox(), app: app, pages: pages, engine: engine, version: currentVersion,
		reg: all.Registry(), now: time.Now, active: true,
		eliteMsg: eliteLines[rand.Intn(len(eliteLines))],
	}
	m.start = m.now()
	m.later = func(fn func()) {
		time.AfterFunc(menuStrikeHold, func() { app.QueueUpdateDraw(fn) })
	}
	m.checkUpdates = func() { showUpdateCheck(app, pages, currentVersion) }
	m.refresh(true)
	return m
}

// ---- what the page says ----

// fingerprint is everything refresh reads from, cheaply: the account, its
// records and display settings, and the saves' names, sizes and times.
func (m *mainMenu) fingerprint() string {
	var sb strings.Builder
	acct := m.engine.Account()
	set := resolveMapSettings(acct, m.reg)
	if acct != nil {
		mainGame, last := acct.GameRecord()
		fmt.Fprintf(&sb, "%s|%s|%s|", acct.AccountID, mainGame, last)
	}
	fmt.Fprintf(&sb, "%s|%d|%v|", set.Style, set.Tier, set.Motion)
	entries, _ := os.ReadDir(filepath.Join(game.DataDir(), "saves"))
	for _, e := range entries {
		if fi, err := e.Info(); err == nil {
			fmt.Fprintf(&sb, "%s:%d:%d|", e.Name(), fi.Size(), fi.ModTime().UnixNano())
		}
	}
	return sb.String()
}

// refresh reads what the page says: the account's settings, its current
// game and a look at it. It reads the saves only when something has
// changed since it last did (or force is set); the rest is cheap.
func (m *mainMenu) refresh(force bool) {
	m.acct = m.engine.Account()
	m.set = resolveMapSettings(m.acct, m.reg)
	if fp := m.fingerprint(); force || fp != m.read {
		m.read = fp
		// What the account has changed (a game was played, a save made or
		// deleted): the selection goes back to the first entry.
		m.view.sel, m.view.items = 0, nil
		m.town.close()
		m.town, m.hasCur = nil, false
		m.contDetails, m.contExtras, m.view.captions = nil, nil, nil
		if cur, ok := m.engine.CurrentGame(); ok {
			// A save that cannot be looked at cannot be opened either: the
			// menu then offers no Continue and shows its first page.
			if st, err := game.ViewSave(cur.Save.Name); err == nil {
				m.cur, m.hasCur = cur, true
				m.town = newMenuTown(&st, m.reg, m.set.Style)
				m.view.captions = menuCaptions(m.town, &st)
				m.continueDetails(&st)
			}
		}
		_, elite := game.SavesOnSplash()
		m.view.elite = [3]string{}
		if elite {
			m.view.elite = m.eliteMsg
		}
	}
	m.buildView()
}

// continueDetails works out the Continue entry's line: the town, its age
// and its people, with shorter forms for a narrow box.
func (m *mainMenu) continueDetails(st *game.GameState) {
	age := st.AgeName
	if age == "" {
		age = ageDisplay(st.Age)
	}
	people := textfmt.Count(st.Workers.TotalPop, "person", "people")
	if st.Workers.TotalPop >= 100000 {
		people = FormatNumber(float64(st.Workers.TotalPop)) + " people"
	}
	var details []string
	if m.town != nil && m.town.name != "" {
		details = append(details, m.town.name+" · "+age+" · "+people)
	}
	m.contDetails = append(details, age+" · "+people, age)
	// The save's own name stands at the right of the row when the player
	// gave it one.
	m.contExtras = nil
	if name := m.cur.Save.Name; name != game.AutosaveName {
		m.contExtras = []string{name, truncate(name, 20), truncate(name, 12)}
	}
}

// menuCaptions is the town page's caption, fullest first: the town, its
// age, and the prestige level when there is one.
func menuCaptions(town *menuTown, st *game.GameState) []string {
	age := st.AgeName
	if age == "" {
		age = ageDisplay(st.Age)
	}
	name := ""
	if town != nil {
		name = town.name
	}
	var out []string
	if name != "" && st.Prestige.Level > 0 {
		out = append(out, fmt.Sprintf("%s · %s · ★ Prestige level %d", name, age, st.Prestige.Level))
	}
	if name != "" {
		out = append(out, name+" · "+age)
	}
	return append(out, age)
}

// buildView puts the entries together from what refresh read. It keeps the
// selection on the entry it was on.
func (m *mainMenu) buildView() { m.buildViewWith(m.badgeExtras()) }

// buildViewWith is buildView with the Badges entry's text given.
func (m *mainMenu) buildViewWith(badges []string) {
	was := ""
	if m.view.sel >= 0 && m.view.sel < len(m.view.items) {
		was = m.view.items[m.view.sel].id
	}
	var items []menuItem
	if m.hasCur {
		items = append(items, menuItem{id: miContinue, label: "Continue", key: 'c', extras: m.contExtras, details: m.contDetails})
	}
	items = append(items,
		menuItem{id: miNew, label: "New game", key: 'n'},
		menuItem{id: miLoad, label: "Load game", key: 'l'},
		menuItem{id: miBadges, label: "Badges", key: 'b', extras: badges},
		menuItem{id: miThemes, label: "Themes", key: 't'},
		menuItem{id: miAccounts, label: "Accounts", key: 'a'},
		menuItem{id: miUpdates, label: "Check for updates", key: 'u'},
		menuItem{id: miQuit, label: "Quit", key: 'q'},
	)
	m.view.items = items
	// Continue is selected by default; without it, New game, which is
	// then first. A selection made by hand stays where it was put.
	m.view.sel = 0
	for i, it := range items {
		if was != "" && it.id == was {
			m.view.sel = i
		}
	}
	m.view.forge = !m.hasCur || m.town == nil || m.town.failed
	m.view.plain = m.set.Tier == mapmodel.TierASCII
	m.view.ages = len(config.Ages())
	m.view.edition = menuEdition(m.version)
	m.view.versions = []string{m.version}
	if m.update {
		m.view.versions = []string{m.version + " · update available (u)", m.version + " · update (u)", m.version}
	}
}

// badgeExtras is the Badges entry's right-hand text: how many badges the
// account holds of those it can see, and the title it wears. The count is
// the badge case's own (hidden badges are not in it).
func (m *mainMenu) badgeExtras() []string {
	if m.acct == nil {
		return nil
	}
	_, sum := m.engine.Badges()
	count := fmt.Sprintf("%d of %d", sum.Earned, sum.Shown)
	var out []string
	if title := wornTitle(sum); title != "" {
		out = append(out, count+" · "+title)
	}
	return append(out, count, strconv.Itoa(sum.Earned))
}

// menuEdition names the edition a version is ("fourth" for v4.1.0), ""
// when the version does not start with a number.
func menuEdition(version string) string {
	v := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if i := strings.IndexAny(v, ".-+ "); i >= 0 {
		v = v[:i]
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return ""
	}
	return ordinalWords(n)
}

// ---- the page's life ----

// Focus is how the menu learns it is the page in use again: tview focuses
// it when a page it opened closes and when the game returns to it.
func (m *mainMenu) Focus(delegate func(p tview.Primitive)) {
	m.Box.Focus(delegate)
	m.enter()
}

// enter makes the menu the page in use: it reads what changed while it
// was not, and its clock starts again from where the scene stopped.
func (m *mainMenu) enter() {
	if !m.active {
		m.active = true
		m.rebase()
	}
	m.pending = false
	m.refresh(false)
}

// leave is called when the menu opens a page or a game: its motion stops.
func (m *mainMenu) leave() {
	m.active = false
	m.stopClock()
}

// rebase restarts the clock at the scene's present frame.
func (m *mainMenu) rebase() {
	m.start = m.now()
	m.baseFrame = 0
	if m.scene != nil {
		m.baseFrame = m.scene.frame
	}
}

// startClock starts the redraw clock, at the map's rate. Draw calls it,
// so nothing runs before the menu is on a screen.
func (m *mainMenu) startClock() {
	if m.stop != nil || m.app == nil {
		return
	}
	stop := make(chan struct{})
	m.stop = stop
	app := m.app
	go func() {
		tk := time.NewTicker(mapAnimStep)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				app.QueueUpdateDraw(func() {})
			case <-stop:
				return
			}
		}
	}()
}

func (m *mainMenu) stopClock() {
	if m.stop != nil {
		close(m.stop)
		m.stop = nil
	}
}

// moving reports whether the page's motion is running.
func (m *mainMenu) moving() bool { return m.active && m.set.Motion }

// Draw draws the page.
func (m *mainMenu) Draw(scr tcell.Screen) {
	if !m.active {
		return // under a page it opened, which covers the screen
	}
	// A new account (the first run names one over the menu) is a new page.
	if m.engine.Account() != m.acct {
		m.refresh(true)
	}
	x, y, w, h := m.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	grid, pal := m.frameGrid(w, h)
	grid.flush(scr, x, y, pal, m.view.plain)
	if m.moving() {
		m.startClock()
	} else {
		m.stopClock()
	}
}

// frameGrid brings the scene to the clock's frame and draws the page at
// w by h.
func (m *mainMenu) frameGrid(w, h int) (*mGrid, *menuPalette) {
	if th := theme.Active(); m.pal == nil || m.pal.key != th.Key {
		m.pal = newMenuPalette(th)
	}
	forge := 0
	if m.view.forge {
		forge = 1
	}
	if key := [4]int{w, h, forge, len(m.view.items)}; m.scene == nil || key != m.sceneKey {
		m.scene, m.sceneKey = newMenuScene(menuLayoutFor(w, h, &m.view)), key
		m.rebase()
	}
	since := time.Duration(0)
	if m.moving() {
		since = m.now().Sub(m.start)
		target := m.baseFrame + int(since/mapAnimStep)
		for i := 0; m.scene.frame < target && i < menuCatchUp; i++ {
			m.scene.step()
		}
		if m.scene.frame < target {
			m.scene.frame = target
		}
	}
	return m.scene.render(m.pal, &m.view, m.town, mapFrame(nil, m.set, since)), m.pal
}

// ---- keys ----

// InputHandler takes the menu's keys: the arrows (and Tab) move, Enter
// opens, and each entry has a letter.
func (m *mainMenu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		if m.pending || len(m.view.items) == 0 {
			return
		}
		n := len(m.view.items)
		switch ev.Key() {
		case tcell.KeyUp, tcell.KeyBacktab:
			m.view.sel = (m.view.sel + n - 1) % n
		case tcell.KeyDown, tcell.KeyTab:
			m.view.sel = (m.view.sel + 1) % n
		case tcell.KeyHome, tcell.KeyPgUp:
			m.view.sel = 0
		case tcell.KeyEnd, tcell.KeyPgDn:
			m.view.sel = n - 1
		case tcell.KeyEnter:
			m.choose(m.view.sel)
		case tcell.KeyRune:
			r := ev.Rune()
			if r == ' ' {
				m.choose(m.view.sel)
				return
			}
			if r >= 'A' && r <= 'Z' {
				r += 'a' - 'A'
			}
			for i, it := range m.view.items {
				if it.key == r {
					m.view.sel = i
					m.choose(i)
					return
				}
			}
		}
	})
}

// choose opens entry i. With motion on the title is struck first and the
// entry opens a moment later, once the flare has been drawn; with motion
// off it opens at once and nothing flares.
func (m *mainMenu) choose(i int) {
	if i < 0 || i >= len(m.view.items) || m.pending {
		return
	}
	id := m.view.items[i].id
	if !m.moving() || m.scene == nil {
		m.open(id)
		return
	}
	m.scene.hit()
	m.pending = true
	m.later(func() {
		if !m.pending {
			return // the menu was left or entered again in the meantime
		}
		m.pending = false
		m.open(id)
	})
}

// open does what an entry does.
func (m *mainMenu) open(id string) {
	switch id {
	case miContinue:
		m.continueGame()
	case miNew:
		// The name first; the game starts only when it is confirmed. Cancel
		// returns to the menu, which never stopped.
		showNewGameNameModal(m.app, m.pages, m, func(name string) {
			m.leave()
			m.engine.StartNewNamedGame(name)
			m.pages.SwitchToPage("dashboard")
			go m.engine.Start()
		})
	case miLoad:
		// Built fresh each time, so the list is always current.
		m.leave()
		page := CreateLoadGamePage(m.app, m.pages, m.engine, "splash", true)
		m.pages.AddPage(loadGamePage, page, true, true)
		m.app.SetFocus(page)
	case miBadges:
		m.leave()
		openMenuBadges(m.app, m.pages, m.engine, m.reg, "splash")
	case miThemes:
		m.leave()
		page := CreateThemePickerPage(m.app, m.pages, m.engine, "splash")
		m.pages.AddPage(themePickerPage, page, true, true)
		m.app.SetFocus(page)
	case miAccounts:
		m.leave()
		page := CreateAccountsPage(m.app, m.pages, m.engine, m.version, "splash")
		m.pages.AddPage(accountsPage, page, true, true)
		m.app.SetFocus(page)
	case miUpdates:
		m.checkUpdates()
	case miQuit:
		m.leave()
		m.app.Stop()
	}
}

// continueGame opens the current game, as the Load Game browser would.
func (m *mainMenu) continueGame() {
	if !m.hasCur {
		return
	}
	name := m.cur.Save.Name
	if err := m.engine.LoadGame(name); err != nil {
		// It was readable a moment ago: say what happened and read again.
		showUpdateMsg(m.app, m.pages, fmt.Sprintf("Could not open '%s'.\n\n%v", name, err))
		m.refresh(true)
		return
	}
	m.leave()
	m.engine.AddLog("success", "Game loaded.")
	m.pages.SwitchToPage("dashboard")
	go m.engine.Start()
}
