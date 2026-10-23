package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
	"github.com/espresso20/ageforge/vibe"
)

// shameMessages are randomly chosen at session start when the cheater badge is active.
// Displayed in the top status bar — a deterrent against savegame manipulation.
var shameMessages = []string{
	"✦ FORGED SCROLLS ✦",
	"✦ ILLEGITIMATE EMPIRE ✦",
	"✦ COUNTERFEIT DYNASTY ✦",
	"✦ USURPER'S THRONE ✦",
	"✦ THE FABRICATED AGE ✦",
	"✦ DECREE OF DISHONOR ✦",
}

// Dashboard is the main gameplay screen. The economy tab is the always-visible
// background; named overlays (research, trade, military, etc.) are rendered on
// top of it via tview.Pages. Only one overlay can be visible at a time.
// promptRows is the command bar's height: one line inside its border. The
// Map panel leaves these rows free, so the prompt works while it is open.
const promptRows = 3

// The dashboard's other fixed sizes: the rows over the boxes (the badge
// line, the status bar, the toast line and the two of the Next Age strip)
// and the width of the right column (the Panels list and the People box).
const (
	dashHeaderRows = 1 + 1 + 1 + 2
	sidebarW       = 22
)

type Dashboard struct {
	app    *tview.Application
	engine *game.GameEngine
	pages  *tview.Pages
	root   *tview.Flex

	// helpTitle is the Help panel's title in the script now in force: the one
	// string of code VibeMode's hook (vibe.Line) reads as yet.
	helpTitle func() string
	// vibeSeen is the VibeMode switch as the dashboard last drew it.
	vibeSeen bool

	// Tabs (only economy is permanent background)
	economyTab *EconomyTab

	// Sidebar
	sidebar       *fitView
	sidebarActive string // the open panel, highlighted in the list
	workerMiniTV  *fitView
	peopleH       int  // the People box's height in the classic layout, border included: its lines
	workerDetail  bool // the terminal has the rows for the People box's food detail line

	// Shared UI
	logTV            *tview.TextView
	statusTV         *fitView
	ageTV            *fitView
	inputField       *commandInput
	lastAge          string
	pendingAgeSplash string // set by bus handler, consumed by refresh()
	// pendingEnding is the record of a run that just ended (a prestige, a
	// fall), set by the bus handler and taken by refresh(), which plays its
	// film. Atomic: the handler runs wherever the engine was called from.
	pendingEnding       atomic.Pointer[game.RunEnding]
	pendingEpochChanged bool // whether the pending age advance also crossed an epoch boundary
	toastMgr            *ToastManager
	toastTV             *tview.TextView
	toastW              int // the toast bar's width as last drawn; 0 before that
	contentArea         tview.Primitive

	// Shame badge — set once on first load when CheaterBadge is true
	cheaterTV        *tview.TextView
	activeShameBadge string

	// catModalShown tracks the pending catastrophe key we have already shown a modal for,
	// so that closing it with Esc does not re-pop it on the next refresh tick (the
	// `catastrophe` command reopens it). Reset to "" when PendingCatastrophe clears
	// (player chose Endure or Succumb) or when a save is loaded.
	catModalShown string
	// catReshow is set by the EventGameLoaded bus handler (which runs under the
	// engine lock, possibly off the UI goroutine) and consumed by refresh(): a
	// freshly loaded save with a pending catastrophe shows its modal again.
	catReshow atomic.Bool
	// catFocus is the catastrophe modal button that owns focus while it is open.
	catFocus tview.Primitive
	// memFocus is the Ancient Memory window's button that owns focus while
	// it is open.
	memFocus tview.Primitive

	// memoryModalShown guards the Ancient Memory offer modal the same way catModalShown
	// guards the catastrophe modal: it holds the offered tech key once shown so a Defer
	// (here: closing without choosing) doesn't re-pop it every refresh. Reset to "" when
	// PendingMemoryTech clears (accept/decline both clear it engine-side).
	memoryModalShown string

	// Command history is session-only (never persisted to disk). The slice is
	// append-only and capped at 50 entries. histIdx == -1 means not in history
	// navigation mode. histDraft stores whatever was in the input field before
	// the user started pressing Up.
	cmdHistory []string // append-only slice, capped at 50
	histIdx    int      // -1 = not navigating; 0 = most recent; len-1 = oldest
	histDraft  string   // draft text saved when user starts navigating history

	overlayMgr *OverlayManager
	lastState  *game.GameState

	// harbPanel is the Harbinger panel's UI state (feedback line, Invite
	// confirmation). Owned by the tview goroutine.
	harbPanel harbingerPanel
	// planPanel is the Plan panel's UI state (selection, feedback line,
	// Clear confirmation). Owned by the tview goroutine.
	planPanel planPanel
	// iconsWin is the icons check's window (icons.go), built on first open.
	// Owned by the tview goroutine.
	iconsWin *iconsWindow

	// The maps: the shared model builder and style registry, the Map panel
	// and the dashboard's mini map. UI goroutine only. mapLocal holds the
	// map settings when no account is loaded (session only).
	mapViews *mapViews
	mapPanel *mapPanel
	// researchPanel is the Research panel, the tech tree as a map
	// (research_panel.go).
	researchPanel *researchPanel
	// badgePanel is the badge case (badge_panel.go).
	badgePanel *badgePanel
	// expPanel is the Expeditions panel: the Ready Room and the Expedition
	// Report (expedition_panel.go).
	expPanel *expeditionPanel

	// peoplePanel is the People panel: Overview, Work, Army and Scholars.
	peoplePanel *peoplePanel

	// advancePanel is the Advance panel: the gate into the next age as four
	// pillars (advance_panel.go).
	advancePanel *advancePanel

	miniMap *miniMap
	// skyline is the skyline strip under the Buildings list (dash_skyline.go).
	skyline  *skylineStrip
	mapDock  *mapDock
	mapLocal *mapSettings
	// icons is the guided icons check (icons.go).
	icons *iconsFlow
	// mapOpen is set while the Map panel is open, so the refresh loop
	// redraws it at the animation rate. Read off the UI goroutine.
	mapOpen atomic.Bool
	// fxOn is set while the active theme's ambient effect is drawn
	// (theme_fx.go), for the same reason. rootFx is the root that draws
	// it and fxStart the effect's clock.
	fxOn    atomic.Bool
	rootFx  *fxRoot
	fxStart time.Time

	stopCh chan struct{}
}

// NewDashboard creates the gameplay dashboard
func NewDashboard(app *tview.Application, engine *game.GameEngine, pages *tview.Pages) *Dashboard {
	mv := newMapViews(all.Registry())
	d := &Dashboard{
		app:           app,
		engine:        engine,
		mapViews:      mv,
		mapPanel:      newMapPanel(mv),
		researchPanel: newResearchPanel(),
		badgePanel:    newBadgePanel(),
		expPanel:      newExpeditionPanel(),
		peoplePanel:   newPeoplePanel(),
		advancePanel:  newAdvancePanel(),
		miniMap:       newMiniMap(mv),
		skyline:       newSkylineStrip(mv),
		pages:         pages,
		stopCh:        make(chan struct{}),
		histIdx:       -1,
	}
	d.build()
	d.overlayMgr = NewOverlayManager(d.pages, d.app, func() {
		d.updateSidebar("")
		d.returnFocus()
	})
	d.overlayMgr.Register("milestones", "Milestones", milestonesProvider)
	// The Research panel: the tech tree as a map. Like the Map it leaves
	// the keyboard with the command bar.
	d.researchPanel.engine = d.engine
	d.researchPanel.tier = func() mapmodel.GlyphTier { return d.mapSettings().Tier }
	d.researchPanel.prompt = func() string { return d.inputField.GetText() }
	d.researchPanel.toPrompt = func(ev *tcell.EventKey) {
		d.overlayMgr.FocusOn(d.inputField)
		if h := d.inputField.InputHandler(); h != nil {
			h(ev, func(p tview.Primitive) { d.app.SetFocus(p) })
		}
	}
	d.overlayMgr.RegisterWidget("techs", "Research", d.researchPanel.open, d.researchPanel.update, true)
	// The badge case: the account's badges. It leaves the keyboard with the
	// command bar too.
	d.badgePanel.settings = d.mapSettings
	d.badgePanel.prompt = d.researchPanel.prompt
	d.badgePanel.toPrompt = d.researchPanel.toPrompt
	d.overlayMgr.RegisterWidget("badges", "Badges", d.badgePanel.open, d.badgePanel.update, true)
	d.overlayMgr.Register("army", "Army", militaryProvider)
	// The Expeditions panel leaves the keyboard with the command bar too.
	d.expPanel.engine = d.engine
	d.expPanel.settings = d.mapSettings
	d.expPanel.prompt = d.researchPanel.prompt
	d.expPanel.toPrompt = d.researchPanel.toPrompt
	d.overlayMgr.RegisterWidget("expedition", "Expeditions", d.expPanel.open, d.expPanel.update, true)
	d.overlayMgr.Register("trade", "Trade routes", tradeProvider)
	// The Factions panel, under two names pointing at one provider (the
	// citymap/map precedent below). "factions" is the primary — it is what the
	// sidebar lists, and the sidebar highlight matches on the registered name, so
	// the two strings have to agree exactly. Bare `diplomacy` is routed to the
	// primary name in input.go so it highlights too; "diplomacy" stays registered
	// so anything holding the old overlay name still resolves.
	d.overlayMgr.Register("factions", "Factions", factionsProvider)
	d.overlayMgr.Register("diplomacy", "Factions", factionsProvider)
	d.overlayMgr.Register("stats", "Statistics", statsProvider)
	d.overlayMgr.Register("wonders", "Wonders", wondersProvider)
	// The People panel, which `workers` opens too, leaves the keyboard with
	// the command bar as well.
	d.peoplePanel.settings = d.mapSettings
	d.peoplePanel.toPrompt = d.researchPanel.toPrompt
	d.overlayMgr.RegisterWidget("people", "People", d.peoplePanel.open, d.peoplePanel.update, true)
	d.registerAdvancePanel()
	d.overlayMgr.Register("logs", "Logs", logsProvider)
	d.overlayMgr.Register("epoch", "Epoch", epochProvider)
	d.overlayMgr.Register("harbinger", "Harbinger", d.harbPanel.provider)
	d.overlayMgr.SetKeyHandler("harbinger", func(event *tcell.EventKey) *tcell.EventKey {
		// tview goroutine, no engine lock held: the engine's action methods and
		// GetState take their own locks.
		if !d.harbPanel.handleKey(event, d.engine) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	d.overlayMgr.Register("plan", "Build Plan", d.planPanel.provider)
	d.overlayMgr.SetKeyHandler("plan", func(event *tcell.EventKey) *tcell.EventKey {
		// tview goroutine, no engine lock held (see the harbinger handler).
		if !d.planPanel.handleKey(event, d.engine) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	d.overlayMgr.Register("history", "Civilization History", historyProvider)
	d.overlayMgr.Register("buildings", "Buildings", buildingsProvider)
	d.helpTitle = func() string { return vibe.Line("ui.dashboard.NewDashboard.18", "Help") }
	d.vibeSeen = vibe.On()
	d.overlayMgr.Register("help", d.helpTitle(), helpProvider)

	// The Map panel: the active map style full screen (map, and its aliases
	// citymap and worldmap). Its settings travel with the account.
	d.mapPanel.settings = d.mapSettings
	d.mapPanel.hintShown = d.markMapHintShown
	d.mapPanel.stage = func(cmd string) {
		// tview goroutine (a key handler): stage the command in the prompt,
		// under the still-open map, for the player to run with Enter.
		d.inputField.SetText(cmd)
	}
	d.mapPanel.prompt = func() string { return d.inputField.GetText() }
	d.mapPanel.zoomed = func(level string) {
		s := d.mapSettings()
		s.Zoom = level
		d.saveMapSettings(s)
	}
	d.mapPanel.spotted = func(kind string) {
		// tview goroutine (a refresh or a key handler), outside the engine
		// lock: the account hears that its player looked at a visitor.
		d.engine.NoteVisitorInspected(kind)
	}
	d.mapPanel.toPrompt = func(ev *tcell.EventKey) {
		// The map itself had the keyboard: give it back to the prompt,
		// starting with this key.
		d.overlayMgr.FocusOn(d.inputField)
		if h := d.inputField.InputHandler(); h != nil {
			h(ev, func(p tview.Primitive) { d.app.SetFocus(p) })
		}
	}
	d.overlayMgr.RegisterWidget("map", "Map", d.mapPanel.open, d.mapPanel.update, true)

	// The icons check is a window of its own (icons.go). It logs through
	// the engine and brings the install's progress and outcome back to the
	// UI goroutine.
	d.icons = newIconsFlow(func(kind, msg string) { d.engine.AddLog(kind, msg) }, d.setMapGlyphsNerd,
		func(f func()) { d.app.QueueUpdateDraw(f) })
	d.icons.changed = d.refreshIconsWindow

	return d
}

func (d *Dashboard) build() {
	// Create permanent economy tab
	d.economyTab = NewEconomyTab()

	// Sidebar — command panel hints, laid out for the room it has: one
	// column when the names fit that way, two when the screen is short.
	d.sidebar = newFitView(func(w, h int) string { return sidebarText(d.sidebarActive, w, h) })
	d.sidebar.SetBorder(true).SetTitle(" Panels ")
	theme.Track(func() { d.sidebar.SetTitleColor(theme.Color(theme.RoleAccent)) })

	// Log panel
	d.logTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true).SetTitle(" Log ")
	theme.Track(func() { d.logTV.SetTitleColor(theme.Color(theme.RoleDim)) })

	// Shame badge bar (1 fixed line; text only shown when CheaterBadge is true)
	d.cheaterTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Status bar: written for the bar's width (statusLine).
	d.statusTV = newFitView(func(w, _ int) string {
		if d.lastState == nil {
			return ""
		}
		return statusLine(*d.lastState, w)
	})
	d.statusTV.SetTextAlign(tview.AlignCenter)

	// Age progress tracker: the next age's requirements, each kept whole on
	// its line.
	d.ageTV = newFitView(func(w, h int) string {
		if d.lastState == nil {
			return ""
		}
		return strings.Join(ageProgressLines(*d.lastState, w, h), "\n")
	})

	// Toast notification
	d.toastMgr = NewToastManager()
	d.toastTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// The bar's width, once it has been drawn: a toast written to fit
	// (a badge's) asks for it. Until then it is 0, any width.
	d.toastTV.SetDrawFunc(func(_ tcell.Screen, x, y, w, h int) (int, int, int, int) {
		d.toastW = w
		return x, y, w, h
	})

	// Subscribe to events for toasts
	// NOTE: Bus handlers run under the engine write lock. Never call GetState()
	// or any other lock-acquiring method inside these closures — use config.*ByKey()
	// (pure data lookups) for any data you need beyond the event payload.
	d.engine.Bus.Subscribe(game.EventAgeAdvanced, func(e game.EventData) {
		if newAge, ok := e.Payload["new_age"].(string); ok {
			// Store for splash — consumed in refresh() which runs in UI goroutine
			// (this handler runs under engine lock, so no GetState here!)
			d.pendingAgeSplash = newAge
			// Detect epoch transition using config only (no engine lock needed)
			oldEpoch := config.EpochForAge(d.lastAge)
			newEpoch := config.EpochForAge(newAge)
			d.pendingEpochChanged = (oldEpoch != newEpoch && d.lastAge != "")
		}
		newAge, _ := e.Payload["new_age"].(string)
		d.toastMgr.Show("Age advanced: "+game.AgeName(newAge), "gold", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventResearchDone, func(e game.EventData) {
		tech, _ := e.Payload["tech"].(string)
		d.toastMgr.Show("Research complete: "+game.TechName(tech), "cyan", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventBuildingBuilt, func(e game.EventData) {
		building, _ := e.Payload["building"].(string)
		// Only toast for wonders — look up from config, not engine state (avoids deadlock)
		if def, ok := config.BuildingByKey()[building]; ok && def.Category == "wonder" {
			d.toastMgr.Show(fmt.Sprintf("Wonder built: %s", def.Name), "green", 4*time.Second)
		}
	})
	d.engine.Bus.Subscribe(game.EventMilestoneCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		rewardText, _ := e.Payload["reward_text"].(string)
		// note: flavor quip rides in the log line, not the height-1 toast — see engine.checkMilestones.
		msg := "Milestone: " + name
		if rewardText != "" {
			msg += " " + rewardText
		}
		d.toastMgr.Show(msg, "gold", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventChainCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		title, _ := e.Payload["title"].(string)
		key, _ := e.Payload["key"].(string)
		boost, _ := e.Payload["boost_ticks"].(int)
		d.toastMgr.Show(chainToast(name, title, config.MilestoneChainByKey()[key], boost), "cyan", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventEpochAdvanced, func(e game.EventData) {
		epochName, _ := e.Payload["epoch_name"].(string)
		epochIcon, _ := e.Payload["epoch_icon"].(string)
		if epochIcon == "" {
			epochIcon = "✦"
		}
		d.toastMgr.Show(fmt.Sprintf("%s %s dawns", epochIcon, epochName), "gold", 6*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventRunEnded, func(e game.EventData) {
		// Runs under the engine write lock: the payload only.
		if end, ok := e.Payload["ending"].(game.RunEnding); ok {
			d.pendingEnding.Store(&end)
		}
	})
	d.engine.Bus.Subscribe(game.EventHarbingerArrived, func(e game.EventData) {
		// Runs under the engine write lock: payload and the toast queue only.
		// A toast is queued behind the age-advance one and never takes focus.
		name, _ := e.Payload["harbinger_name"].(string)
		verb := "has come"
		if handoff, _ := e.Payload["handoff"].(bool); handoff {
			verb = "takes up the warning"
		}
		d.toastMgr.Show(fmt.Sprintf("⚑ %s %s. Type harbinger to read the warning.", capFirstUI(name), verb), "warning", 8*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventGameLoaded, func(e game.EventData) {
		// Runs under the engine write lock: only flip the flag, never touch the engine.
		d.catReshow.Store(true)
	})
	d.engine.Bus.Subscribe(game.EventEpochEventFired, func(e game.EventData) {
		eventName, _ := e.Payload["event_name"].(string)
		eventType, _ := e.Payload["event_type"].(string)
		if eventType == "catastrophe" {
			d.toastMgr.Show(fmt.Sprintf("☄ Catastrophe: %s. Type catastrophe to choose.", eventName), "red", 8*time.Second)
			return
		}
		color := "cyan"
		if eventType == "bad_challenging" {
			color = "red"
		} else if eventType == "catastrophe" {
			color = "red"
		} else if eventType == "good_legendary" {
			color = "gold"
		}
		d.toastMgr.Show(fmt.Sprintf("Epoch event: %s", eventName), color, 6*time.Second)
	})

	// Command input: completions show as ghost text (command_input.go),
	// drawn from the registry and the state of the last refresh.
	d.inputField = newCommandInput(newCompleter(d.engine, func() *game.GameState { return d.lastState }))
	d.inputField.SetLabel("❯ ").SetFieldWidth(0)
	d.inputField.SetBorder(true).SetTitle(" Command ")
	theme.Track(func() {
		d.inputField.SetFieldBackgroundColor(theme.Color(theme.RoleBackground)).
			SetLabelColor(theme.Color(theme.RoleHighlight))
		// Frame the command bar so it reads as a first-class element, not an afterthought.
		d.inputField.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent))
	})

	d.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			d.submitInput()
		}
	})

	// Phase 16: history navigation via Up/Down arrow keys
	d.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if len(d.cmdHistory) == 0 {
				return nil // nothing to navigate
			}
			if d.histIdx == -1 {
				// Start navigating: save current draft, go to most recent
				d.histDraft = d.inputField.GetText()
				d.histIdx = len(d.cmdHistory) - 1
			} else if d.histIdx > 0 {
				d.histIdx--
			}
			// histIdx == 0: already at oldest, stay
			d.inputField.SetText(d.cmdHistory[d.histIdx])
			return nil // swallow key
		case tcell.KeyDown:
			if d.histIdx == -1 {
				return nil // not in history mode, no-op
			}
			if d.histIdx == len(d.cmdHistory)-1 {
				// Back to draft
				d.histIdx = -1
				d.inputField.SetText(d.histDraft)
			} else {
				d.histIdx++
				d.inputField.SetText(d.cmdHistory[d.histIdx])
			}
			return nil // swallow key
		default:
			// Any other key: exit history mode and update draft
			if d.histIdx != -1 {
				d.histIdx = -1
			}
			// Tab, Backtab and → (at the end of the line) take completions.
			if d.inputField.acceptKey(event) {
				return nil
			}
			// Keep draft in sync while user types normally
			// (draft is re-read from field on next Up press, so nothing extra needed)
			return event
		}
	})

	// The log goes at the foot of the economy tab's left column, under
	// Under construction (econColumn shares the height out).
	d.economyTab.AddToLeftColumn(d.logTV)

	// The mini map docks above the Buildings list and hides itself when the
	// terminal is too small for it (mapDock). So does the skyline strip,
	// under the list, which is half the Log's height.
	d.economyTab.WrapBuildings(func(list tview.Primitive) tview.Primitive {
		d.mapDock = newMapDock(d.miniMap, list)
		d.mapDock.strip, d.mapDock.logRows, d.mapDock.screen = d.skyline, d.economyTab.LogRows, d.screenSize
		return d.mapDock
	})

	// People box: the population and its four kinds. In the classic layout it
	// is below the Panels list in the right column, as tall as its lines; in
	// the rich layout it is the right half of the Under construction row
	// (dash_layout.go).
	d.workerMiniTV = newFitView(func(w, h int) string {
		if d.lastState == nil {
			return ""
		}
		if d.richLayout() {
			return peopleRichText(*d.lastState, w, h)
		}
		return peopleBoxText(*d.lastState, w, d.workerDetail)
	})
	d.workerMiniTV.SetBorder(true).SetTitle(" People ")
	theme.Track(func() { d.workerMiniTV.SetTitleColor(theme.Color(theme.RoleAccent)) })
	d.economyTab.AddPeopleBox(d.workerMiniTV, d.richLayout, func() int {
		_, th := d.screenSize()
		return richPeopleHeight(th)
	})

	// Main horizontal: the economy tab (the full height) and the Panels
	// column, with the People box under the list unless the layout is rich.
	d.peopleH = workerMiniBaseRows + 2
	mainHoriz := newDashBody(d.economyTab.Root(), d.sidebar, d.workerMiniTV, d.screenSize, func() int { return d.peopleH })
	mainHoriz.switched = func() { d.workerMiniTV.stale = true }

	// Content area is just mainHoriz: no bottom strip
	d.contentArea = mainHoriz

	// Root layout (no tab bar)
	d.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.cheaterTV, 1, 0, false).
		AddItem(d.statusTV, 1, 0, false).
		AddItem(d.toastTV, 1, 0, false).
		AddItem(d.ageTV, 2, 0, false).
		AddItem(d.contentArea, 0, 1, false).
		AddItem(d.inputField, promptRows, 0, true)

	// Global key handling
	d.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Ctrl+K — open passphrase modal (dev unlock)
		if event.Key() == tcell.KeyCtrlK {
			d.showDevUnlockModal()
			return nil
		}
		// The open Map panel takes the keys that print nothing; the rest
		// reach the command bar, which keeps the focus (map_panel.go).
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "map" && d.inputField.HasFocus() &&
			d.mapPanel.routeKey(event, d.inputField.GetText()) {
			return nil
		}
		// So does the open Research panel, and Esc closes its card first.
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "techs" && d.inputField.HasFocus() &&
			(d.researchPanel.routeKey(event, d.inputField.GetText()) || event.Key() == tcell.KeyEsc && d.researchPanel.closeCard()) {
			return nil
		}
		// And the open badge case, where Esc closes a badge's detail first.
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "badges" && d.inputField.HasFocus() &&
			(d.badgePanel.routeKey(event, d.inputField.GetText()) || event.Key() == tcell.KeyEsc && d.badgePanel.closeCard()) {
			return nil
		}
		// The open Expeditions panel, where Esc takes back the question to
		// send a party first.
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "expedition" && d.inputField.HasFocus() &&
			(d.expPanel.routeKey(event, d.inputField.GetText()) || event.Key() == tcell.KeyEsc && d.expPanel.closeConfirm()) {
			return nil
		}
		// The open People panel, which takes the arrows while the bar is empty.
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "people" && d.inputField.HasFocus() &&
			d.peoplePanel.routeKey(event, d.inputField.GetText()) {
			return nil
		}
		// The open Advance panel, which takes the arrows and Enter while the
		// bar is empty.
		if d.overlayMgr != nil && d.overlayMgr.ActiveName() == "charter" && d.inputField.HasFocus() &&
			d.advancePanel.routeKey(event, d.inputField.GetText()) {
			return nil
		}
		switch event.Key() {
		case resourcePageKey:
			// The Resources box's next page, when it has more rows than it
			// can show. With a panel open the box is behind it.
			if d.overlayMgr == nil || !d.overlayMgr.HasActive() {
				d.economyTab.NextResourcePage()
				return nil
			}
		case tcell.KeyEsc:
			if d.overlayMgr != nil && d.overlayMgr.HasActive() {
				d.overlayMgr.Hide()
				return nil
			}
			d.engine.SaveGame(d.engine.ActiveSaveName())
			d.engine.Stop()
			d.pages.SwitchToPage("splash")
			return nil
		// Economy tab scroll keys (always available since economy is permanent background)
		case tcell.KeyPgUp:
			if !d.overlayMgr.HasActive() {
				d.economyTab.ScrollUp()
				return nil
			}
		case tcell.KeyPgDn:
			if !d.overlayMgr.HasActive() {
				d.economyTab.ScrollDown()
				return nil
			}
		}

		// Always focus input field for typing.
		if !d.inputField.HasFocus() {
			d.app.SetFocus(d.inputField)
		}
		return event
	})
}

func (d *Dashboard) updateSidebar(activeOverlay string) {
	if d.sidebar != nil && d.sidebarActive != activeOverlay {
		d.sidebarActive = activeOverlay
		d.sidebar.changed()
	}
}

// sidebarPanels is the Panels list, in order.
var sidebarPanels = []string{"milestones", "badges", "research", "plan", "expedition", "army", "trade", "factions", "stats", "wonders", "people", "logs", "epoch", "harbinger", "history", "map", "help"}

// buildSidebarText is the Panels list in one column, the open panel
// highlighted: the list as it shows when the screen is tall enough for it.
func buildSidebarText(active string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, cmd := range sidebarPanels {
		if cmd == active {
			sb.WriteString(" " + theme.Selected(fmt.Sprintf(" %-10s ", cmd)) + "\n")
		} else {
			sb.WriteString(fmt.Sprintf(" [white]%-10s[-]\n", cmd))
		}
	}
	return sb.String()
}

// sidebarText is the Panels list for a box w cells wide and h rows tall
// inside: one column when every name gets a row (under a blank line when
// there is a row to spare), else two columns read down then across, so a
// short screen still lists every panel. When even two columns are too
// many rows (the smallest screen), the first column stays a name a row and
// the rest are packed beside it, as many to a row as fit. A box wide and
// tall enough for it (the rich layout's) lists the panels in groups, each
// with a line on what it is for (panelsRichText).
func sidebarText(active string, w, h int) string {
	if w >= panelsRichWidth() && h >= panelsRichRows() {
		return panelsRichText(active, w, h)
	}
	n := len(sidebarPanels)
	if h > n {
		return buildSidebarText(active)
	}
	if h == n {
		return strings.TrimPrefix(buildSidebarText(active), "\n")
	}
	cell := func(cmd string, width int) string {
		padded := cmd + strings.Repeat(" ", max(width-len(cmd), 0))
		if cmd == active {
			return theme.Selected(padded)
		}
		return "[white]" + padded + "[-]"
	}
	// pack writes names as rows of at most room cells, a space between two.
	pack := func(names []string, room int) [][]string {
		var rows [][]string
		used := 0
		for _, name := range names {
			if len(rows) == 0 || used+1+len(name) > room {
				rows = append(rows, nil)
				used = -1
			}
			rows[len(rows)-1] = append(rows[len(rows)-1], name)
			used += 1 + len(name)
		}
		return rows
	}
	join := func(names []string) string {
		cells := make([]string, len(names))
		for i, name := range names {
			cells[i] = cell(name, len(name))
		}
		return strings.Join(cells, " ")
	}
	var sb strings.Builder
	rows := (n + 1) / 2
	if rows > h && h > 0 {
		leftW := 0
		for _, cmd := range sidebarPanels[:h] {
			leftW = max(leftW, len(cmd))
		}
		right := pack(sidebarPanels[h:], w-leftW-1)
		if len(right) > h {
			// Not even that fits: every name, packed across the box.
			for _, row := range pack(sidebarPanels, w) {
				sb.WriteString(join(row) + "\n")
			}
			return sb.String()
		}
		for i := 0; i < h; i++ {
			sb.WriteString(cell(sidebarPanels[i], leftW))
			if i < len(right) {
				sb.WriteString(" " + join(right[i]))
			}
			sb.WriteString("\n")
		}
		return sb.String()
	}
	leftW, rightW := 0, 0
	for i, cmd := range sidebarPanels {
		if i < rows {
			leftW = max(leftW, len(cmd))
		} else {
			rightW = max(rightW, len(cmd))
		}
	}
	gap := max(1, min(2, w-leftW-rightW))
	for i := 0; i < rows; i++ {
		sb.WriteString(cell(sidebarPanels[i], leftW))
		if j := i + rows; j < n {
			sb.WriteString(strings.Repeat(" ", gap) + cell(sidebarPanels[j], rightW))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Root returns the root primitive for page registration: the layout, and
// over it the active theme's ambient effect.
func (d *Dashboard) Root() tview.Primitive {
	if d.rootFx == nil {
		d.rootFx = &fxRoot{Flex: d.root, after: d.drawEffect}
	}
	return d.rootFx
}

// drawEffect draws the active theme's ambient effect over the dashboard,
// when the theme has one and the motion setting is on.
func (d *Dashboard) drawEffect(scr tcell.Screen, x, y, w, h int) {
	effect := theme.Active().Effect
	if effect == "" {
		d.fxOn.Store(false)
		return
	}
	set := d.mapSettings()
	d.fxOn.Store(set.Motion)
	if !set.Motion {
		return
	}
	if d.fxStart.IsZero() {
		d.fxStart = time.Now()
	}
	// Not in the command bar: nothing moves where the player types.
	h -= promptRows
	drawThemeEffect(scr, x, y, w, h, effect, int(time.Since(d.fxStart)/mapAnimStep), set.Tier == mapmodel.TierASCII)
}

// StartUpdates begins the UI refresh loop, polling at 500 ms (2 fps).
// This is intentionally slower than the game tick rate (which can reach ~5 fps
// at 2x speed) to avoid burning CPU on terminal redraws.
// NOTE: app.QueueUpdateDraw is the only safe way to touch tview primitives from
// a background goroutine — it serialises with the tview event loop.
func (d *Dashboard) StartUpdates() {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		anim := time.NewTicker(mapAnimStep)
		defer anim.Stop()
		for {
			select {
			case <-ticker.C:
				d.app.QueueUpdateDraw(func() {
					d.refresh()
				})
			case <-anim.C:
				// The open Map panel animates between refreshes: a redraw
				// is enough, since its frame counter runs on the clock.
				// So does the badge case, while it shows a badge that moves,
				// and a theme with an ambient effect.
				if d.mapOpen.Load() || d.badgePanel.moving.Load() || d.expPanel.moving.Load() || d.fxOn.Load() {
					d.app.QueueUpdateDraw(func() {
						if d.overlayMgr.ActiveName() != "map" {
							d.mapOpen.Store(false)
						}
						if d.overlayMgr.ActiveName() != "badges" {
							d.badgePanel.moving.Store(false)
						}
						if d.overlayMgr.ActiveName() != "expedition" {
							d.expPanel.moving.Store(false)
						}
					})
				}
			case <-d.stopCh:
				return
			}
		}
	}()
}

// StopUpdates stops the UI refresh loop by sending on stopCh.
// The non-blocking send (default branch) prevents a hang if nobody is
// listening (e.g. StopUpdates called twice before the goroutine drains).
func (d *Dashboard) StopUpdates() {
	select {
	case d.stopCh <- struct{}{}:
	default:
	}
}

// syncVibe writes again the text the dashboard set once, when VibeMode has
// been switched since (by the command, or by an account that was switched to).
// The text the panels draw from the game's state follows on its own: the
// engine moved onto the other ruleset when the switch changed.
func (d *Dashboard) syncVibe() {
	if on := vibe.On(); on != d.vibeSeen {
		d.vibeSeen = on
		d.overlayMgr.Retitle("help", d.helpTitle())
	}
}

// refresh is the central UI update function, called every 500 ms from StartUpdates.
// It must only be called from within app.QueueUpdateDraw (i.e. on the tview goroutine).
// Order matters: catastrophe modal and age splash are checked first because they may
// swap the active page, then all content widgets are updated.
func (d *Dashboard) refresh() {
	d.syncVibe()
	// Check for pending age splash (set by bus handler under engine lock)
	if d.pendingAgeSplash != "" {
		newAge := d.pendingAgeSplash
		oldAge := d.lastAge
		epochChanged := d.pendingEpochChanged
		d.pendingAgeSplash = ""
		d.pendingEpochChanged = false

		state := d.engine.GetState()
		summary := state.LastAgeAdvanceSummary

		var epochEvent game.EpochEventRecord
		if epochChanged && len(state.EpochEventHistory) > 0 {
			epochEvent = state.EpochEventHistory[len(state.EpochEventHistory)-1]
		}

		ShowAgeSplashFull(d.overlayMgr, oldAge, newAge, summary, epochChanged, epochEvent)
	}
	// A run that just ended (a prestige, a fall): its film, before the
	// overlay manager is given the new run's state below, so that what it
	// last saw is still the town that was given up.
	if end := d.pendingEnding.Swap(nil); end != nil {
		ShowRunEnding(d.overlayMgr, *end)
	}

	state := d.engine.GetState()
	d.lastState = &state

	// Badges earned since the last refresh: a toast and a log line each, and a
	// line for the theme a badge gives. Here, in the UI goroutine: outside the
	// engine lock, never in a Bus handler. They wait for the film of a run's
	// ending, which covers the toast bar, and are told when it is over.
	if d.overlayMgr.ActiveName() != endingPageName {
		d.announceBadges()
	}

	// Phase 9: catastrophe modal — show once per new pending catastrophe; Esc hides it
	// until the player types `catastrophe` (or loads a save, which shows it again).
	// Never stack it on top of the age splash: an epoch-transition roll sets
	// PendingCatastrophe inside the same advance, so both surface in this refresh.
	// Stacked, the modal steals focus from the splash ("press any key" does nothing),
	// and the splash's 20s auto-dismiss (OverlayManager.Hide → onClose) then moves
	// focus to the input field underneath the still-visible modal, leaving it
	// unreachable by keyboard. Wait until the splash is gone; the next refresh shows it.
	if d.catReshow.Swap(false) {
		d.catModalShown = ""
	}
	// The Last Passage (a prestige from the Cosmic Era waiting on its choice)
	// uses the same modal and the same single-show rule; pendingChoiceKey names it.
	if pending := pendingChoiceKey(state); pending == "" {
		d.catModalShown = "" // reset so next catastrophe will show fresh
		if d.pages.HasPage(catastrophePage) {
			d.closeCatastropheModal() // e.g. a save without a pending catastrophe was loaded
		}
	} else if d.catModalShown != pending && !d.overlayMgr.FullScreenUp() {
		if d.pages.HasPage(catastrophePage) {
			d.pages.RemovePage(catastrophePage) // stale modal for a different epoch
		}
		d.catModalShown = pending
		d.showCatastropheModal(pending)
	}

	// Ancient Memory offer modal — same single-show pattern as the catastrophe modal.
	if state.PendingMemoryTech == "" {
		d.memoryModalShown = "" // reset so a future run's cache can show fresh
	} else if d.memoryModalShown == "" && !d.overlayMgr.FullScreenUp() {
		// Not over the film of the run that just ended: the new run's
		// first window waits for it.
		d.memoryModalShown = state.PendingMemoryTech
		d.showAncientMemoryModal(state.PendingMemoryTech, state.PendingMemoryTechName)
	}

	// Shame badge — pick once per session, never change after that
	if state.CheaterBadge && d.activeShameBadge == "" {
		d.activeShameBadge = shameMessages[rand.Intn(len(shameMessages))]
	}
	if d.activeShameBadge != "" {
		d.cheaterTV.SetText(fmt.Sprintf("[red]%s[-]", d.activeShameBadge))
	}

	if d.lastAge != state.Age {
		ApplyAgePalette(state.Age)
		d.lastAge = state.Age
	}

	d.refreshStatus(state)
	d.refreshAgeProgress(state)
	d.refreshLog(state)
	d.toastTV.SetText(safeTags(d.toastMgr.CurrentFor(d.toastW)))

	// Economy tab is always visible as the permanent background
	d.economyTab.Refresh(state)
	// The mini map's model, only while it is on and has room to show
	// (before the first layout the dock has no size yet, so build it
	// anyway).
	set := d.mapSettings()
	d.mapDock.off = !set.Minimap
	if d.mapDock.wantsModel() {
		d.miniMap.update(set, &state)
		d.miniMap.pointAt(d.mapPanel)
	}
	// The skyline strip's model, the same way.
	if d.mapDock.wantsStrip() {
		d.skyline.update(set, &state)
	}

	// Update overlay content and sidebar highlight
	d.overlayMgr.Refresh(state)
	d.maybeShowPrestigeCue(state)
	d.updateSidebar(d.overlayMgr.ActiveName())
	d.refreshWorkerMini(state)
}

// announceBadges says what the account earned since the last refresh. Badges
// its record already proved when it was loaded get one line between them.
// Then it drains the badges the account earned since the last refresh
// (the engine judged them under its lock and only queued them) and gives each a
// toast in the badge's own colours, written for the width of the toast bar, and
// one log line, with a second line for the theme it unlocks when it gives one. The line is fixed per badge, so earning one draws
// nothing from the run's random streams.
func (d *Dashboard) announceBadges() {
	if d.engine == nil {
		return
	}
	// Badges the account's own record already proved when it was loaded
	// (a ladder was lowered since): one line for all of them.
	if n := d.engine.DrainBadgeCatchUp(); n > 0 {
		line := game.BadgeCatchUpLine(n)
		d.toastMgr.Show(line, "green", 5*time.Second)
		d.engine.AddLog("success", line)
	}
	for _, v := range d.engine.DrainEarnedBadges() {
		d.toastMgr.ShowFit(func(w int) string { return badgeToast(v, d.mapSettings().Tier, w) }, 5*time.Second)
		d.engine.AddLog("success", game.BadgeLogLine(v))
		if t, ok := theme.ByKey(v.RewardTheme); ok {
			d.engine.AddLog("success", themeUnlockToast(t.Name))
		}
	}
}

func (d *Dashboard) refreshWorkerMini(game.GameState) {
	// Written from d.lastState for the box's width at its next draw
	// (peopleBoxText, peopleRichText).
	if d.lastState != nil {
		_, rootH := d.screenSize()
		if d.richLayout() {
			// The box has the Under construction row's height (the
			// economy tab sizes it), and its title carries the headcount
			// only when the box is too short for the row that does.
			d.workerDetail = true
			d.workerMiniTV.SetTitle(peopleRichTitle(*d.lastState, richPeopleHeight(rootH)-2))
		} else {
			// The box is as tall as its lines. The food detail line is the
			// one that goes first on a short terminal: at 24 rows the Panels
			// list needs every row the box leaves it.
			d.workerDetail = rootH >= workerDetailMinRows
			d.peopleH = strings.Count(peopleBoxText(*d.lastState, sidebarW-2, d.workerDetail), "\n") + 2
			d.workerMiniTV.SetTitle(peopleBoxTitle(*d.lastState, d.workerDetail))
		}
	}
	d.workerMiniTV.changed()
}

// screenSize is the terminal's size as the dashboard was last laid out in it.
func (d *Dashboard) screenSize() (w, h int) {
	if d.root == nil {
		return 0, 0
	}
	_, _, w, h = d.root.GetRect()
	return w, h
}

// richLayout says whether the dashboard is in the rich layout (dash_layout.go).
func (d *Dashboard) richLayout() bool { return dashRich(d.screenSize()) }

// workerDetailMinRows is the terminal height from which the People box adds
// the line that says how much food is spare and how many it feeds.
const workerDetailMinRows = 30

// workerMiniBaseRows is the People box's lines when nothing holds the
// population back: the headcount over the housing and the four kinds.
const workerMiniBaseRows = 5

// heldByLines says what holds the population back, for a box w cells wide (0:
// any width), as "Held by: food" and, for food, a second line with the net
// food and the villagers it feeds. Nothing when the population can grow.
func heldByLines(ws game.WorkerState, w int) []string {
	fits := func(plain string) bool { return w <= 0 || runeLen(plain) <= w }
	switch ws.HeldBy {
	case game.HeldHousing:
		return []string{"Held by: housing"}
	case game.HeldFood:
		x := textfmt.RateValue(ws.FoodNet) // "+0.45"
		feeds := fmt.Sprintf("feeds %d more", ws.FoodFeeds)
		for _, line := range []string{
			x + " food a tick " + feeds,
			x + "/tick " + feeds,
			x + " " + feeds,
			feeds,
		} {
			if fits(line) {
				return []string{"Held by: food", line}
			}
		}
		return []string{"Held by: food"}
	}
	return nil
}

// heldByText is heldByLines for one line (the roster command and panel): the
// food form joins its two parts, "Held by: food (+0.45 food a tick feeds 3
// more)".
func heldByText(ws game.WorkerState) string {
	lines := heldByLines(ws, 0)
	switch len(lines) {
	case 0:
		return ""
	case 1:
		return lines[0]
	}
	return lines[0] + " (" + lines[1] + ")"
}

// peopleBoxCompact says whether the People box is short a row: on a short
// terminal (detail unset) with something holding the population back, the
// Held-by line takes the row of the headcount, which moves to the title. The
// Panels list needs every row the box leaves it.
func peopleBoxCompact(state game.GameState, detail bool) bool {
	return !detail && state.Workers.HeldBy != game.HeldNone
}

// peopleBoxTitle is the People box's title: " People ", with the headcount
// over the housing in it when the box is compact and it fits.
func peopleBoxTitle(state game.GameState, detail bool) string {
	if !peopleBoxCompact(state, detail) {
		return " People "
	}
	ws := state.Workers
	for _, t := range []string{
		fmt.Sprintf(" People %d/%d ", ws.TotalPop, ws.MaxPop),
		fmt.Sprintf(" People %s/%s ", FormatNumber(float64(ws.TotalPop)), FormatNumber(float64(ws.MaxPop))),
	} {
		if runeLen(t) <= sidebarW-2 {
			return t
		}
	}
	return " People "
}

// peopleBoxText writes the People box for a box w cells wide (0: any width):
// the population over the housing, one line for each kind of person in this
// age's own word with its count (the soldiers carry "· N out" for those on
// missions when it fits), and what holds the population back (with the food
// detail line when detail is set). A count too long for its line is written
// short (51.8M for 51785785). The food use and net are on the People panel.
func peopleBoxText(state game.GameState, w int, detail bool) string {
	ws := state.Workers
	fits := func(plain string) bool { return w <= 0 || runeLen(plain) <= w }
	var sb strings.Builder

	// On a short terminal with a Held-by line up, the headcount moves to the
	// box's title (peopleBoxTitle) and its row goes to the Panels list.
	if !peopleBoxCompact(state, detail) {
		pop, housing := fmt.Sprintf("%d", ws.TotalPop), fmt.Sprintf("%d", ws.MaxPop)
		if !fits("Pop " + pop + "/" + housing) {
			pop, housing = FormatNumber(float64(ws.TotalPop)), FormatNumber(float64(ws.MaxPop))
		}
		popColor := "yellow"
		if ws.TotalPop > ws.MaxPop {
			popColor = "red" // the scholars moved in past the homes
		}
		fmt.Fprintf(&sb, "[gray]Pop[white] [%s]%s[white]/[green]%s[-]\n", popColor, pop, housing)
	}

	for _, k := range peopleKinds(state) {
		sb.WriteString(k.boxLine(w))
		sb.WriteString("\n")
	}
	if lines := heldByLines(ws, w); len(lines) > 0 {
		fmt.Fprintf(&sb, "[gray]%s[-]\n", lines[0])
		if len(lines) > 1 && detail {
			fmt.Fprintf(&sb, "[gray]%s[-]\n", lines[1])
		}
	}
	return sb.String()
}

func (d *Dashboard) refreshStatus(game.GameState) {
	// The line is written from d.lastState for the bar's width at its next
	// draw (statusLine), and in full until the bar has been drawn.
	d.statusTV.changed()
}

// statusLine writes the status bar for a bar w cells wide (0: as wide as
// it likes). When the whole line does not fit it gives up, in order: half
// the hint, then all of it, the wording of the catastrophe and harbinger
// badges (and long counts are written short), what morale does to
// production, the civilization's title, the account's name, the rest of the
// badges' words and the wide spacing. The age, the badges, the population
// and morale always show.
func statusLine(state game.GameState, w int) string {
	// note: the next age lives on the Next Age bar below, so the status bar
	// no longer repeats it (it used to print the raw key).
	prestigeStr := ""
	if state.Prestige.Level > 0 {
		prestigeStr = fmt.Sprintf("  [cyan]Prestige %d[-]", state.Prestige.Level)
	}
	devStr := ""
	if game.DevModeActive {
		devStr = "  [red]DEV[-]"
	}
	// Colour morale by the continuous production multiplier, not the raw percent:
	// green = bonus (mult>1.0), white = neutral (==1.0), red = penalty (<1.0).
	// computeMoraleBand is the shared source of truth (same as the workers panel).
	mBand := computeMoraleBand(state.Morale, state.MoraleMultiplier)

	// build writes the line at a level of brevity: each level gives up one
	// more thing.
	build := func(level int) string {
		titleStr := ""
		if state.Milestones.CurrentTitle != "" && level < 5 {
			titleStr = fmt.Sprintf("  [yellow]\"%s\"[-]", state.Milestones.CurrentTitle)
		}
		// Pending catastrophe badge: persistent until the player chooses, so a player
		// who closed the modal (or came back after a long idle) sees why advancing
		// is blocked and how to reopen the choice.
		catStr := ""
		badge := func(fg, bg theme.Role, full, brief, least string) {
			text := full
			if level >= 7 {
				text = least
			} else if level >= 3 {
				text = brief
			}
			catStr += fmt.Sprintf("  %s %s %s", theme.TagFgBg(fg, bg), text, theme.Reset)
		}
		if state.PendingCatastrophe != "" {
			badge(theme.RoleOnNegative, theme.RoleNegative, "☄ Catastrophe pending. Type catastrophe to choose.", "☄ Catastrophe: type catastrophe", "☄ Catastrophe")
		} else if state.LastPassage.Pending {
			badge(theme.RoleOnNegative, theme.RoleNegative, "☄ Last Passage. Type catastrophe to choose.", "☄ Last Passage: type catastrophe", "☄ Last Passage")
		}
		// Harbinger badge: present until the doom it warns of resolves. Nothing
		// expires, so the badge is the idle player's reminder that there is a
		// choice waiting (it never blocks anything).
		if state.Harbinger != nil {
			badge(theme.RoleOnAccent, theme.RoleAccent, "⚑ Harbinger. Type harbinger to read it.", "⚑ Harbinger: type harbinger", "⚑ Harbinger")
		}
		moraleDelta := ""
		if mBand.DeltaLabel != "" && level < 4 {
			moraleDelta = fmt.Sprintf(" (production [%s]%s[-])", mBand.Color, mBand.DeltaLabel)
		}
		moraleStr := fmt.Sprintf("  Morale [%s]%.0f%%[-]%s", mBand.Color, state.Morale*100, moraleDelta)
		// Leading account-name segment, when an account is wired. Truncate a long name so
		// the status line stays readable on narrow terminals.
		acctStr := ""
		if state.AccountStats != nil && state.AccountStats.DisplayName != "" && level < 6 {
			name := state.AccountStats.DisplayName
			if len(name) > 20 {
				name = name[:19] + "…"
			}
			acctStr = fmt.Sprintf("[gold]%s[-] · ", name)
			// The title the account wears, once it is not the one every
			// account starts with, and only on a bar with room.
			if sum := state.AccountStats.BadgeSummary; level == 0 && (sum.TitleRank > 0 || wornTitle(sum) != sum.Title) {
				acctStr = fmt.Sprintf("[gold]%s[-] [gray]%s[-] · ", name, wornTitle(sum))
			}
		}
		hint := ""
		switch level {
		case 0:
			hint = "  |  [gray]type a panel name to open it · Esc: close or menu[-]"
		case 1:
			hint = "  |  [gray]Esc: close or menu[-]"
		}
		pop := fmt.Sprintf("%d/%d", state.Workers.TotalPop, state.Workers.MaxPop)
		if level >= 3 {
			pop = FormatNumber(float64(state.Workers.TotalPop)) + "/" + FormatNumber(float64(state.Workers.MaxPop))
		}
		line := fmt.Sprintf("%s[gold]%s[-]%s%s%s%s  |  Pop: %s%s%s",
			acctStr, state.AgeName, prestigeStr, titleStr, catStr, devStr, pop, moraleStr, hint)
		if level >= 9 {
			line = strings.ReplaceAll(line, "  ", " ")
		}
		return line
	}
	line := build(0)
	for level := 1; level <= 9 && w > 0 && visibleLen(line) > w; level++ {
		line = build(level)
	}
	return line
}

func (d *Dashboard) refreshAgeProgress(game.GameState) {
	// The row is written from d.lastState at its next draw (ageProgressLines).
	d.ageTV.changed()
}

// ageGoal is one thing the next age asks for: its text ("Hut 7/10")
// and whether it is met.
type ageGoal struct {
	text string
	met  bool
}

// ageGoals lists what the next age asks for: the age's wonder and its
// keystone tech while they are still to do, then the buildings. The wonder
// leads so that a long list of buildings can never push it out of sight.
func ageGoals(state game.GameState) []ageGoal {
	var goals []ageGoal
	// CurrentAgeWonderKey is cleared once the wonder is built, so its presence == not yet built.
	if state.CurrentAgeWonderKey != "" {
		banked := wonderBankPercent(state, state.CurrentAgeWonderKey)
		goals = append(goals, ageGoal{fmt.Sprintf("Wonder: %s %.0f%% banked", state.CurrentAgeWonderName, banked), false})
		// Its keystone tech, while that is still to research.
		if tech := state.Buildings[state.CurrentAgeWonderKey].NeedsTech; tech != "" {
			goals = append(goals, ageGoal{"Keystone: " + game.TechName(tech), false})
		}
	}
	for _, key := range sortedKeysOf(state.NextAgeBldReqs) {
		req := state.NextAgeBldReqs[key]
		current := 0
		if bs, ok := state.Buildings[key]; ok {
			// A gate counts the building and the later rungs of its line.
			current = max(bs.Count, bs.HeldWithLater)
		}
		goals = append(goals, ageGoal{fmt.Sprintf("%s %d/%d", game.BuildingName(key), current, req), current >= req})
	}
	return goals
}

// ageProgressLines writes the Next Age row for a strip w cells wide and h
// rows tall: the age's name, then what it asks for (the wonder and its
// keystone first, then the buildings), each marked ✓ or ✗ and kept whole on
// its line. When the strip is too small for all of it, what is
// already met is counted instead of listed, and after that the list ends
// with how many more there are.
func ageProgressLines(state game.GameState, w, h int) []string {
	lines := ageStripLines(state, w, h)
	if state.NextAge == "" || len(lines) == 0 {
		return lines
	}
	// The strip is the short form of the Advance panel: say how to open it
	// when the last line has room for the words.
	last := len(lines) - 1
	if used := runeLen(plainLogText(lines[last])); used+2+runeLen(charterStripHint) <= w {
		lines[last] += "  [gray]" + charterStripHint + "[-]"
	}
	return lines
}

// charterStripHint ends the Next Age strip when it has the room.
const charterStripHint = "type charter"

// ageStripLines is the strip without its hint.
func ageStripLines(state game.GameState, w, h int) []string {
	if w < 1 || h < 1 {
		return nil
	}
	if state.NextAge == "" {
		return []string{" [gold]" + truncate("You have reached the final age.", w-1) + "[-]"}
	}
	type cell struct {
		tagged string
		width  int
	}
	title := truncate("Next Age: "+state.NextAgeName, w-1)
	goalCell := func(g ageGoal) cell {
		// The wonder and its keystone are red all through, as they were.
		if strings.HasPrefix(g.text, "Wonder: ") || strings.HasPrefix(g.text, "Keystone: ") {
			return cell{"[red]✗ " + g.text + "[-]", 2 + runeLen(g.text)}
		}
		if g.met {
			return cell{"[green]✓[-] " + g.text, 2 + runeLen(g.text)}
		}
		return cell{"[red]✗[-] " + g.text, 2 + runeLen(g.text)}
	}
	note := func(s string) cell { return cell{"[gray]" + s + "[-]", runeLen(s)} }
	// pack flows cells onto lines of at most w cells, two spaces apart,
	// after the title.
	pack := func(cells []cell) []string {
		lines := []string{" [gold]" + title + "[-]"}
		used := 1 + runeLen(title)
		for _, c := range cells {
			if used+2+c.width > w && used > 1 {
				lines = append(lines, " "+c.tagged)
				used = 1 + c.width
				continue
			}
			lines[len(lines)-1] += "  " + c.tagged
			used += 2 + c.width
		}
		return lines
	}
	goals := ageGoals(state)
	var all, open []cell
	met := 0
	for _, g := range goals {
		all = append(all, goalCell(g))
		if g.met {
			met++
		} else {
			open = append(open, goalCell(g))
		}
	}
	if lines := pack(all); len(lines) <= h {
		return lines
	}
	// What is met is counted, not listed.
	if met > 0 {
		open = append(open, note(fmt.Sprintf("✓ %d met", met)))
		if lines := pack(open); len(lines) <= h {
			return lines
		}
		open = open[:len(open)-1]
	}
	// As many of the open ones as fit, then how many more there are.
	for keep := len(open) - 1; keep >= 0; keep-- {
		rest := len(open) - keep + met
		if lines := pack(append(append([]cell(nil), open[:keep]...), note(fmt.Sprintf("+%d more", rest)))); len(lines) <= h {
			return lines
		}
	}
	return pack(nil)[:1]
}

func (d *Dashboard) refreshLog(state game.GameState) {
	var sb strings.Builder
	// Every line but debug ones, routine confirmations included, so the
	// player sees what each command did (log_routing.go).
	var visible []game.LogEntry
	for _, entry := range state.Log {
		if mainLogShows(entry) {
			visible = append(visible, entry)
		}
	}
	start := 0
	if len(visible) > 20 {
		start = len(visible) - 20
	}
	// No tick number: the line starts with what happened, and the logs panel
	// keeps the tick. The color keys off the entry type for scannability;
	// routine confirmations take the plain text color, so notable lines
	// still stand out among them.
	for _, entry := range visible[start:] {
		color := "white"
		switch entry.Type {
		case "success":
			color = "green"
		case "warning":
			color = "yellow"
		case "error":
			color = "red"
		case "event":
			color = "gold"
		case "info":
			color = "cyan"
		}
		fmt.Fprintf(&sb, "[%s]%s[-]\n", color, entry.Message)
	}
	d.logTV.SetText(safeTags(sb.String()))
	d.logTV.ScrollToEnd()
}

// showDevUnlockModal opens an unlabelled passphrase input modal.
// No hint text is shown — the existence of this modal is not advertised.
func (d *Dashboard) showDevUnlockModal() {
	if game.DevModeActive {
		return // already active, nothing to do
	}

	const devUnlockPage = "__dev_unlock__"
	// Transient dev-unlock modal: rebuilt each open and removed on close, so it
	// construction-reads theme.Color without enrolling in Track (the theming design §3.3).
	field := tview.NewInputField().
		SetLabel("").
		SetFieldWidth(40).
		SetMaskCharacter('·').
		SetFieldBackgroundColor(theme.Color(theme.RoleBackground)).
		SetFieldTextColor(theme.Color(theme.RoleText))

	field.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			d.pages.RemovePage(devUnlockPage)
			d.app.SetFocus(d.inputField)
			return
		}
		if key == tcell.KeyEnter {
			input := field.GetText()
			d.pages.RemovePage(devUnlockPage)
			if game.CheckDevKey(input) {
				d.engine.NoteDevUnlocked()
				d.engine.AddLog("info", "[red]Dev mode on.[-] Prefix commands with / (for example /god, /fill, /give wood 9999).")
				d.app.SetFocus(d.inputField)
			} else {
				d.app.SetFocus(d.inputField)
			}
		}
	})

	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(field, 1, 0, true)
	box.SetBorder(true).
		SetBorderColor(theme.Color(theme.RoleDim)).
		SetBackgroundColor(theme.Color(theme.RoleBackground))

	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(box, 3, 0, true).
			AddItem(nil, 0, 1, false), 44, 0, true).
		AddItem(nil, 0, 1, false)

	d.pages.AddPage(devUnlockPage, centered, true, true)
	d.app.SetFocus(field)
}

// accountNoticePage is the page of the notice leaveToMenu shows over the main menu.
const accountNoticePage = "account_notice"

// leaveToMenu takes the player from the dashboard to the main menu after a command
// ended the game in progress: an account switch or recovery, where the engine already
// stopped the run and saved it to the account it belongs to, so unlike Esc this does
// not save. msg, the command's reply, is shown in a notice over the menu, since the
// game log it would otherwise go to belongs to the run that just ended.
func (d *Dashboard) leaveToMenu(msg string) {
	if d.overlayMgr != nil && d.overlayMgr.HasActive() {
		d.overlayMgr.Hide()
	}
	d.pages.SwitchToPage("splash")
	notice := tview.NewModal().
		SetText(safeTags(msg)).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			d.pages.RemovePage(accountNoticePage)
		})
	d.pages.AddPage(accountNoticePage, notice, true, true)
}

// codeWindowPage is the page of the small window a command's Window text is
// shown in.
const codeWindowPage = "code_window"

// showWindow shows text in a small window over the dashboard, titled with the
// recovery code's label, and closes it on Esc or Enter. The text lives in the
// window only: it is never added to the game log.
func (d *Dashboard) showWindow(text string) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	tv.SetBorder(true).SetTitle(" Recovery code ")
	tv.SetText("\n" + tview.Escape(text) + "\n\n" + theme.KeycapButton("Esc", "Close"))
	done := func() {
		d.pages.RemovePage(codeWindowPage)
		d.app.SetFocus(d.inputField)
	}
	tv.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyEnter {
			done()
			return nil
		}
		return nil
	})
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tv, 7, 0, true).
			AddItem(nil, 0, 1, false), 49, 0, true).
		AddItem(nil, 0, 1, false)
	d.pages.AddPage(codeWindowPage, centered, true, true)
	d.app.SetFocus(tv)
}

// submitInput runs the command in the input field, clears it and records it
// in history. Called on Enter. A whole command runs as typed; otherwise the
// ghost completion runs when it makes a whole command, except a Dangerous
// one, which is put in the field for a second Enter (completer.enterLine).
func (d *Dashboard) submitInput() {
	text, run := d.inputField.comp.enterLine(d.inputField.GetText())
	if !run {
		d.inputField.SetText(text)
		return
	}
	d.inputField.SetText("")
	// Reset history navigation state
	d.histIdx = -1
	d.histDraft = ""
	if text == "" {
		return
	}
	// Record non-empty trimmed commands in history (cap at 50)
	cmd := strings.TrimSpace(text)
	if cmd != "" {
		if len(d.cmdHistory) >= 50 {
			d.cmdHistory = d.cmdHistory[1:] // drop oldest
		}
		d.cmdHistory = append(d.cmdHistory, cmd)
	}
	if strings.ToLower(cmd) == "quit" {
		// The same way out as Ctrl+C and a signal: the game in play is
		// stopped and saved to its own save. (A failure has nowhere to be
		// shown: the screen goes with the app.)
		_, _ = d.engine.SaveOnExit()
		d.app.Stop()
		return
	}
	// Route /commands to dev exec when dev mode is active
	if game.DevModeActive && strings.HasPrefix(cmd, "/") {
		result := game.DevConsoleCommand(cmd, d.engine)
		if result != "" {
			d.engine.AddLog("info", "[positive]dev → "+result+"[-]")
		}
		return
	}
	if strings.ToLower(cmd) == "save" { // bare save — no name
		d.showSaveChoiceModal()
		return
	}
	if strings.ToLower(cmd) == "load" { // bare load — open the browser instead of assuming a slot
		page := CreateLoadGamePage(d.app, d.pages, d.engine, "dashboard", false)
		d.pages.AddPage(loadGamePage, page, true, true)
		d.app.SetFocus(page)
		return
	}
	if strings.ToLower(cmd) == "theme" { // bare theme — open the live picker (theme list / theme <key> fall through)
		page := CreateThemePickerPage(d.app, d.pages, d.engine, "dashboard")
		d.pages.AddPage(themePickerPage, page, true, true)
		d.app.SetFocus(page)
		return
	}
	result := HandleCommand(text, d.engine)
	if result.ToMenu {
		d.leaveToMenu(result.Message)
		return
	}
	if result.Window != "" {
		d.showWindow(result.Window)
		return
	}
	if result.OpenCatastrophe {
		d.reopenCatastropheModal()
	}
	if result.OverlayName == "harbinger" {
		d.harbPanel.reset()
	}
	if result.OverlayName == "plan" {
		d.planPanel.reset()
	}
	if result.OverlayName == "map" {
		d.mapPanel.world = result.MapWorld
		d.mapOpen.Store(true)
	}
	if result.OverlayName == "techs" {
		d.researchPanel.request(result.ResearchZoom, result.ResearchCard)
	}
	if result.OverlayName == "badges" {
		d.badgePanel.request(result.Badges)
	}
	if result.Icons {
		d.startIcons()
	}
	if result.Vibe {
		d.syncVibe()
		d.refresh()
	}
	if result.MapPref.Key != "" {
		d.applyMapPref(result.MapPref)
		d.mapDock.off = !d.mapSettings().Minimap
	}
	if result.MapFlows != "" {
		result.Message = flowsReply(d.mapPanel.setFlows(result.MapFlows))
	}
	if result.OverlayName != "" {
		state := d.engine.GetState()
		d.overlayMgr.Show(result.OverlayName, state)
		d.updateSidebar(result.OverlayName)
		if result.OverlayName == "map" || result.OverlayName == "techs" || result.OverlayName == "badges" || result.OverlayName == "expedition" || result.OverlayName == "people" || result.OverlayName == "charter" {
			// The command bar keeps the keyboard while the map is open.
			d.overlayMgr.FocusOn(d.inputField)
		}
	} else if d.overlayMgr.ActiveName() == "badges" {
		// A command typed over the case (map glyphs, motion): show what it
		// changed now rather than at the next refresh.
		d.badgePanel.update(d.engine.GetState())
	} else if d.overlayMgr.ActiveName() == "people" {
		// A command typed over the panel (roster, recruit): show what it
		// changed now rather than at the next refresh.
		d.peoplePanel.update(d.engine.GetState())
	} else if d.overlayMgr.ActiveName() == "charter" {
		// A command typed over the panel (build, advance): show what it
		// changed now rather than at the next refresh.
		d.advancePanel.update(d.engine.GetState())
	} else if d.overlayMgr.ActiveName() == "expedition" {
		// An offer sent by name: the map shows it gone now.
		d.expPanel.update(d.engine.GetState())
	} else if d.overlayMgr.ActiveName() == "techs" {
		// A command typed over the tree (research <tech>, plan research):
		// show what it changed now rather than at the next refresh.
		d.researchPanel.update(d.engine.GetState())
	} else if d.overlayMgr.ActiveName() == "map" {
		// The log is behind the map: say what happened on its key bar, and
		// show the change now rather than at the next refresh.
		d.mapPanel.reply(cmd, result)
		d.mapPanel.update(d.engine.GetState())
	}
	if result.Message != "" && result.Type != "success" {
		// A badge this command earned is announced first, so the reply, which the
		// player asked for, is the last line in the log and stays in view.
		d.announceBadges()
		d.engine.AddLog(result.Type, result.Message)
	}
}
