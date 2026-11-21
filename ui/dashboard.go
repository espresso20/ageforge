package ui

import (
	"fmt"
	"image"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/window/kit"
)

// shameMessages are randomly chosen at session start when the cheater badge is active.
// Displayed in the top status bar — a deterrent against savegame manipulation.
var shameMessages = []rowid.Entry{
	rowid.E("ui.dashboard.shameMessages.01", "✦ FORGED SCROLLS ✦"),
	rowid.E("ui.dashboard.shameMessages.02", "✦ ILLEGITIMATE EMPIRE ✦"),
	rowid.E("ui.dashboard.shameMessages.03", "✦ COUNTERFEIT DYNASTY ✦"),
	rowid.E("ui.dashboard.shameMessages.04", "✦ USURPER'S THRONE ✦"),
	rowid.E("ui.dashboard.shameMessages.05", "✦ THE FABRICATED AGE ✦"),
	rowid.E("ui.dashboard.shameMessages.06", "✦ DECREE OF DISHONOR ✦"),
}

// Dashboard is the main gameplay screen. The economy tab is the always-visible
// background; named overlays (research, trade, military, etc.) are rendered on
// top of it via the page stack. Only one overlay can be visible at a time.
// The command bar is barRows tall: a typed line at twice the size between
// blank rows inside its border (6), or one line at normal size (3) when the
// window is under barBigMinCols by barBigMinRows. Panels leave these rows
// free, so the prompt works while one is open.
const (
	barBigMinCols = 120
	barBigMinRows = 39
	barBigRows    = 6
	barSmallRows  = 3
)

// barRows is the command bar's height in a w by h window.
func barRows(w, h int) int {
	if w >= barBigMinCols && h >= barBigMinRows {
		return barBigRows
	}
	return barSmallRows
}

// bar is the command bar's height in the window as it was last laid out.
func (d *Dashboard) bar() int { return barRows(d.screenSize()) }

// The dashboard's other fixed sizes: the rows over the boxes on the smallest
// terminals (the header, dash_header.go; a bigger terminal gives it one
// more) and the width of the right column (the Panels list and the People
// box).
const (
	dashHeaderRows = headerSmallRows
	sidebarW       = 22
)

type Dashboard struct {
	app    *appLoop
	engine *game.GameEngine
	pages  *nativePages
	root   *nativeFlex

	// Tabs (only economy is permanent background)
	economyTab *EconomyTab

	// The Next card and the Panels list (the right column's top box)
	nextBox      *fitView
	workerMiniTV *fitView
	peopleH      int  // the People box's height in the classic layout, border included: its lines
	workerDetail bool // the terminal has the rows for the People box's food detail line

	// Shared UI
	logTV *nativeText
	// headerTV is the framed header at the top (dash_header.go) and ledger
	// the gate's facts it draws under the age.
	headerTV         *headerBox
	ledger           ledger
	inputField       *commandInput
	lastAge          string
	pendingAgeSplash string // set by bus handler, consumed by refresh()
	// pendingEnding is the record of a run that just ended (a prestige, a
	// fall), set by the bus handler and taken by refresh(), which plays its
	// film. Atomic: the handler runs wherever the engine was called from.
	pendingEnding       atomic.Pointer[game.RunEnding]
	pendingEpochChanged bool // whether the pending age advance also crossed an epoch boundary
	toastMgr            *ToastManager
	badgeQ              []game.BadgeView // earned, notice not shown yet
	badgeNote           bool
	nextHeld            bool // the command line holds the Next step
	badgeUntil          time.Time
	contentArea         Primitive

	// Shame badge — set once on first load when CheaterBadge is true; the
	// header shows it as an alert.
	shame *rowid.Entry // the message picked for the session; read as it is now each frame

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
	catFocus Primitive
	// memFocus is the Ancient Memory window's button that owns focus while
	// it is open.
	memFocus Primitive

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
	// next is the next step, worked out from lastState once a refresh.
	next nextStep
	// panels is what the account has found and opened (panels_found.go).
	panels panelView

	// The lessons: the band above the Command box, what was learned and put
	// off, and the counts some lessons read (lessons.go).
	lessons                   lessonBook
	bandBox                   *fitView
	lessonCard, lessonPanel   string
	lessonTyped, lessonListed int

	// harbPanel is the Harbinger panel's UI state (feedback line, Invite
	// confirmation). Owned by the UI goroutine.
	harbPanel harbingerPanel
	// planPanel is the Plan panel's UI state (selection, feedback line,
	// Clear confirmation). Owned by the UI goroutine.
	planPanel *planPanel

	// The maps: the shared model builder, the Map panel and the dashboard's
	// mini map. UI goroutine only. mapLocal holds the
	// map settings when no account is loaded (session only).
	mapViews  *mapViews
	mapPanel  *mapPanel
	wonders   *wondersView
	factions  *factionsView
	epochHall *epochHall
	// researchPanel is the Research panel, the tech tree as a map
	// (research_panel.go).
	researchPanel *researchPanel
	// treePanel is the whole tech tree as a map, one key from the board.
	treePanel *treePanel
	// badgePanel is the Ranks screen (ranks_panel.go).
	badgePanel *ranksPanel
	// expPanel is the Expeditions panel: the Ready Room and the Expedition
	// Report (expedition_panel.go).
	expPanel *expeditionPanel

	// peoplePanel is the People panel: Overview, Work, Army and Scholars.
	peoplePanel *peoplePanel
	// statsPanel is the Stats panel: the record board of the run.
	statsPanel *statsPanel
	// historyPanel is the History panel: the metrics of the run.
	historyPanel *historyPanel
	// logsPanel is the Logs panel: the roll and the slate.
	logsPanel *logsPanel

	// decks are the panels drawn as decks of cards (deck_panel.go), by
	// overlay name.
	decks map[string]*deckPanel

	// advancePanel is the Advance panel: the gate into the next age as four
	// pillars (advance_panel.go).
	advancePanel *advancePanel

	// barOverlays are the panels that leave the keyboard with the command
	// bar, by overlay name (bar_panel.go).
	barOverlays map[string]barOverlay

	miniMap  *miniMap
	body     *dashBody
	mapLocal *mapSettings
	// window is the window the game is in (nil in the terminal), for the
	// pause card's display settings.
	window WindowHost
	// mapOpen is set while the Map panel is open, so the refresh loop
	// redraws it at the animation rate. Read off the UI goroutine.
	mapOpen atomic.Bool
	// fxOn is set while the active theme's ambient effect is drawn
	// (theme_fx.go), for the same reason. rootFx is the root that draws
	// it and fxStart the effect's clock.
	fxOn    atomic.Bool
	rootFx  *fxRoot
	fxStart time.Time
	// fxNow is the effect's clock; nil means time.Now (a test pins it).
	fxNow func() time.Time

	stopCh chan struct{}
}

// NewDashboard creates the gameplay dashboard
func NewDashboard(app *appLoop, engine *game.GameEngine, pages *nativePages) *Dashboard {
	mv := newMapViews()
	d := &Dashboard{
		app:           app,
		engine:        engine,
		mapViews:      mv,
		mapPanel:      newMapPanel(mv),
		researchPanel: newResearchPanel(),
		treePanel:     newTreePanel(),
		badgePanel:    newRanksPanel(),
		expPanel:      newExpeditionPanel(),
		peoplePanel:   newPeoplePanel(),
		statsPanel:    newStatsPanel(),
		historyPanel:  newHistoryPanel(),
		planPanel:     newPlanPanel(),
		logsPanel:     newLogsPanel(),
		decks:         map[string]*deckPanel{},
		advancePanel:  newAdvancePanel(),
		miniMap:       newMiniMap(mv),
		pages:         pages,
		stopCh:        make(chan struct{}),
		histIdx:       -1,
	}
	d.build()
	d.retitle()
	d.overlayMgr = NewOverlayManager(d.pages, d.app, func() {
		d.returnFocus()
	})
	d.registerDeck("milestones", rowid.E("ui.dashboard.NewDashboard.01", "Milestones"), milestonesPage).draw = renderMileRoads
	// The Research panel: the tech tree as a map. Like the Map it leaves
	// the keyboard with the command bar.
	d.researchPanel.engine = d.engine
	d.researchPanel.prompt = func() string { return d.inputField.GetText() }
	d.researchPanel.toPrompt = d.handToPrompt
	d.treePanel.wireButtons(d.app)
	d.overlayMgr.RegisterWidget("techs", rowid.E("ui.dashboard.NewDashboard.02", "Research"), d.researchPanel.open, d.researchPanel.update)
	d.researchPanel.toTree = func(key string) { d.switchOverlay("tree", func() { d.treePanel.show(key) }) }
	d.treePanel.engine = d.engine
	d.treePanel.prompt = d.researchPanel.prompt
	d.treePanel.toPrompt = d.handToPrompt
	d.treePanel.back = func() { d.switchOverlay("techs", nil) }
	d.overlayMgr.RegisterWidget("tree", rowid.E("ui.dashboard.NewDashboard.tree", "Research tree"), d.treePanel.open, d.treePanel.update)
	// The Ranks screen: the account's badges. It leaves the keyboard with
	// the command bar too.
	d.badgePanel.prompt = d.researchPanel.prompt
	d.badgePanel.toPrompt = d.handToPrompt
	d.badgePanel.buttons.wire(d.app)
	d.researchPanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget("badges", rowid.E("ui.dashboard.NewDashboard.03", "Badges"), d.badgePanel.open, d.badgePanel.update)
	d.registerDeck("army", rowid.E("ui.dashboard.NewDashboard.04", "Army"), armyPage).draw = renderArmyField
	// The Expeditions panel leaves the keyboard with the command bar too.
	d.expPanel.engine = d.engine
	d.expPanel.settings = d.mapSettings
	d.expPanel.prompt = d.researchPanel.prompt
	d.expPanel.toPrompt = d.handToPrompt
	d.expPanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget("expedition", txtExpeditions, d.expPanel.open, d.expPanel.update)
	d.registerDeck("trade", rowid.E("ui.dashboard.NewDashboard.06", "Trade routes"), tradePage).draw = renderTradeBoard
	d.factions = &factionsView{}
	d.overlayMgr.Register("factions", rowid.E("ui.dashboard.NewDashboard.07", "Factions"), d.factions.provider)
	d.factions.wire(d.overlayMgr)
	d.wonders = &wondersView{}
	d.overlayMgr.Register("wonders", rowid.E("ui.dashboard.NewDashboard.10", "Wonders"), d.wonders.provider)
	d.overlayMgr.SetOpenHandler("wonders", d.wonders.open)
	d.overlayMgr.SetBackHandler("wonders", d.wonders.back)
	d.overlayMgr.SetFooter("wonders", d.wonders.footer)
	d.wonders.wire(d.overlayMgr, func() { d.overlayMgr.Refresh(d.engine.GetState()) })
	// The People panel, which `workers` opens too, leaves the keyboard with
	// the command bar as well.
	d.statsPanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget("stats", rowid.E("ui.dashboard.NewDashboard.09", "Statistics"), d.statsPanel.open, d.statsPanel.update)
	d.peoplePanel.toPrompt = d.handToPrompt
	d.peoplePanel.wire(d.app)
	d.overlayMgr.RegisterWidget("people", rowid.E("ui.dashboard.NewDashboard.11", "People"), d.peoplePanel.open, d.peoplePanel.update)
	d.registerAdvancePanel()
	d.logsPanel.prompt = d.researchPanel.prompt
	d.logsPanel.toPrompt = d.handToPrompt
	d.logsPanel.buttons.wire(d.app)
	d.logsPanel.focus = func() { d.overlayMgr.FocusOn(d.logsPanel) }
	d.overlayMgr.RegisterWidget("logs", txtLogsTitle, d.logsPanel.open, d.logsPanel.update)
	d.epochHall = &epochHall{
		height: func() int {
			om := d.overlayMgr
			return om.screenH - barRows(om.screenW, om.screenH) - d.bandRows(om.screenH)
		},
		motion: func() bool { return d.mapSettings().Motion },
	}
	d.epochHall.state = func() game.GameState { return d.engine.GetState() }
	d.epochHall.card = func(st game.GameState) (epochCardResult, error) { return makeEpochCard(st, copyImageToClipboard) }
	d.overlayMgr.Register("epoch", rowid.E("ui.dashboard.NewDashboard.13", "Epoch"), d.epochHall.provider)
	d.overlayMgr.SetNotice("epoch", d.epochHall.notice)
	d.overlayMgr.SetOpenHandler("epoch", d.epochHall.open)
	d.overlayMgr.SetBackHandler("epoch", d.epochHall.back)
	d.overlayMgr.SetFooter("epoch", d.epochHall.footer)
	d.overlayMgr.SetKeyHandler("epoch", func(event *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held (see the harbinger handler).
		if !d.epochHall.handleKey(event) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	d.overlayMgr.Register("harbinger", rowid.E("ui.dashboard.NewDashboard.14", "Harbinger"), d.harbPanel.provider)
	d.harbPanel.view.attach(d.overlayMgr, "harbinger", d.harbPanel.view.layer)
	d.overlayMgr.SetFooter("harbinger", func() [][2]string { return harbingerFooter(d.engine.GetState().Harbinger != nil) })
	d.overlayMgr.SetKeyHandler("harbinger", func(event *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held: the engine's action methods and
		// GetState take their own locks.
		if !d.harbPanel.handleKey(event, d.engine) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	d.planPanel.engine = d.engine
	d.planPanel.prompt = d.researchPanel.prompt
	d.planPanel.toPrompt = d.handToPrompt
	d.planPanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget("plan", txtBuildPlan, d.planPanel.open, d.planPanel.update)
	d.peoplePanel.hist = d.engine.History
	d.historyPanel.src = func() (*game.RunHistory, game.RunArchiveView, string) {
		return d.engine.History(), d.engine.RunArchive(), d.engine.ActiveSaveName()
	}
	d.historyPanel.prompt = d.researchPanel.prompt
	d.historyPanel.toPrompt = d.handToPrompt
	d.historyPanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget("history", txtHistoryTitle, d.historyPanel.open, d.historyPanel.update)
	d.overlayMgr.Register("buildings", txtBuildings, buildingsProvider)
	d.overlayMgr.Register("help", rowid.E("ui.dashboard.NewDashboard.18", "Help"), func(st game.GameState, w int) string {
		return helpProvider(d.lessonKeys(), st, w)
	})

	// The Map panel: the map full screen (map, and its aliases citymap and
	// worldmap). Its settings travel with the account.
	d.mapPanel.settings = d.mapSettings
	d.mapPanel.stage = func(cmd string) {
		// UI goroutine (a key handler): stage the command in the prompt,
		// under the still-open map, for the player to run with Enter.
		d.inputField.SetText(cmd)
	}
	d.mapPanel.prompt = d.researchPanel.prompt
	d.mapPanel.spotted = func(kind string) {
		// UI goroutine (a refresh or a key handler), outside the engine
		// lock: the account hears that its player looked at a visitor.
		d.engine.NoteVisitorInspected(kind)
	}
	d.mapPanel.settle = func(ids []string) { d.engine.SettlePlaces(ids) } // same thread and lock rules as spotted
	d.mapPanel.toPrompt = d.handToPrompt
	d.overlayMgr.RegisterWidget("map", rowid.E("ui.dashboard.NewDashboard.19", "Map"), d.mapPanel.open, d.mapPanel.update)
	d.barOverlays = d.newBarOverlays()

	return d
}

func (d *Dashboard) build() {
	// Create permanent economy tab
	d.economyTab = NewEconomyTab(d.app)
	d.economyTab.buildBox.offer = d.engine.UpgradeOffer

	// Sidebar — command panel hints, laid out for the room it has: one
	// column when the names fit that way, two when the screen is short.
	d.bandBox = newFitView(func(w, _ int) string { return d.bandText(w) })
	d.bandBox.body.OnRegion = func(name string, _ int) { d.bandClick(name) }
	d.nextBox = newFitView(func(w, h int) string {
		s := d.next
		s.onLine = d.inputField != nil && s.cmd != "" && d.inputField.GetText() == s.cmd
		d.nextHeld = s.onLine
		return nextBoxText(s, w, h, d.richLayout())
	})
	// A click on a panel's name opens it, as typing the name does.
	d.nextBox.body.OnRegion = func(name string, _ int) {
		switch {
		case name == regionShow:
			d.showMe(d.next.card, d.next.panel)
		case strings.HasPrefix(name, regionThen):
			if i := int(name[len(regionThen)] - '0'); i < len(d.next.then) {
				d.showMe(d.next.then[i].card, d.next.then[i].panel)
			}
		default:
			d.lessonListed++
			d.pressLine(name)
		}
	}
	// The pointer on a panel's name shows the line a click runs.
	shown := ""
	d.nextBox.body.OnHover = func(name string) {
		if name == regionShow || strings.HasPrefix(name, regionThen) {
			name = ""
		}
		d.app.feel.rest(name, shown)
		shown = name
	}
	// A click on the mini map opens the Map panel.
	d.miniMap.onClick = func() { d.pressLine("map") }
	d.nextBox.SetBorder(true)
	theme.Track(func() { d.nextBox.SetTitleColor(theme.Of(theme.RoleAccent)) })

	// Log panel
	d.logTV = newNativeText().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true)
	// A click on the Log opens the full record.
	d.logTV.click = func() { d.pressLine("logs") }
	theme.Track(func() { d.logTV.SetTitleColor(theme.Of(theme.RoleDim)) })

	// The header at the top: the account, the age, the people, the way to
	// the next age (headerView).
	d.toastMgr = NewToastManager()
	d.headerTV = newHeaderBox(d.headerView, d.pressLine, d.app)

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
		d.toastMgr.Show(rowid.Line("ui.dashboard.build.03", "Age advanced: ")+game.AgeName(newAge), "gold", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventResearchDone, func(e game.EventData) {
		tech, _ := e.Payload["tech"].(string)
		d.toastMgr.Show(rowid.Line("ui.dashboard.build.04", "Research complete: ")+game.TechName(tech), "cyan", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventBuildingBuilt, func(e game.EventData) {
		building, _ := e.Payload["building"].(string)
		// Only toast for wonders — look up from config, not engine state (avoids deadlock)
		if def, ok := config.BuildingByKey()[building]; ok && def.Category == "wonder" {
			d.toastMgr.Show(rowid.Sprintf("ui.dashboard.build.05", "Wonder built: %s", def.Name), "green", 4*time.Second)
		}
	})
	d.engine.Bus.Subscribe(game.EventMilestoneCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		rewardText, _ := e.Payload["reward_text"].(string)
		// note: flavor quip rides in the log line, not the height-1 toast — see engine.checkMilestones.
		msg := rowid.Line("ui.dashboard.build.06", "Milestone: ") + name
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
		d.toastMgr.Show(rowid.Sprintf("ui.dashboard.build.07", "%s %s dawns", epochIcon, epochName), "gold", 6*time.Second)
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
		verb := rowid.Line("ui.dashboard.build.08", "has come")
		if handoff, _ := e.Payload["handoff"].(bool); handoff {
			verb = rowid.Line("ui.dashboard.build.09", "takes up the warning")
		}
		d.toastMgr.Show(rowid.Sprintf("ui.dashboard.build.10", "⚑ %s %s. Type harbinger to read the warning.", textfmt.Capitalize(name), verb), "warning", 8*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventGameLoaded, func(e game.EventData) {
		// Runs under the engine write lock: only flip the flag, never touch the engine.
		d.catReshow.Store(true)
	})
	d.engine.Bus.Subscribe(game.EventEpochEventFired, func(e game.EventData) {
		eventName, _ := e.Payload["event_name"].(string)
		eventType, _ := e.Payload["event_type"].(string)
		if eventType == "catastrophe" {
			d.toastMgr.Show(rowid.Sprintf("ui.dashboard.build.11", "☄ Catastrophe: %s. Type catastrophe to choose.", eventName), "red", 8*time.Second)
			return
		}
		color := "cyan"
		if eventType == "bad_challenging" {
			color = "red"
		} else if eventType == "good_legendary" {
			color = "gold"
		}
		d.toastMgr.Show(rowid.Sprintf("ui.dashboard.build.12", "Epoch event: %s", eventName), color, 6*time.Second)
	})

	// Command input: completions show as ghost text (command_input.go),
	// drawn from the registry and the state of the last refresh.
	d.inputField = newCommandInput(newCompleter(d.engine, func() *game.GameState { return d.lastState }))
	d.inputField.preview = func() string {
		if p := d.app.feel.preview; p != "" {
			return p
		}
		return d.cardLine()
	}
	d.app.line, d.app.confirm = d.pressLine, d.askLine
	d.app.feel.tip.Cell = d.tipCell
	d.inputField.SetLabel("❯ ").SetFieldWidth(0)
	// The card's key hint reads Enter while the line holds the step.
	d.inputField.SetChangedFunc(func(text string) {
		if held := text != "" && text == d.next.cmd; held != d.nextHeld {
			d.nextHeld = held
			d.nextBox.changed()
		}
	})
	d.inputField.SetBorder(true)
	theme.Track(func() {
		d.inputField.SetFieldBackgroundColor(theme.Of(theme.RoleBackground)).
			SetLabelColor(theme.Of(theme.RoleHighlight))
		// Frame the command bar so it reads as a first-class element, not an afterthought.
		d.inputField.SetBorderColor(theme.Of(theme.RoleAccent)).
			SetTitleColor(theme.Of(theme.RoleAccent))
	})

	d.inputField.SetDoneFunc(func(key kit.KeyCode) {
		if key == kit.KeyEnter {
			d.submitInput()
		}
	})

	// Phase 16: history navigation via Up/Down arrow keys
	d.inputField.SetInputCapture(func(event *kit.Key) *kit.Key {
		switch event.Code {
		case kit.KeyUp:
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
		case kit.KeyDown:
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
			// Tab on an empty line of the main view writes the next step's
			// command; the panels keep Tab for themselves.
			if event.Code == kit.KeyTab && d.takeNext() {
				return nil
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

	// People box: the population and its four kinds. In the classic layout it
	// is below the Panels list in the right column, as tall as its lines; in
	// the rich layout it is the left of the middle column's foot strip
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
	// A click on the People box opens the People panel, and one on Under
	// construction the build plan.
	d.workerMiniTV.click = func() { d.pressLine("people") }
	d.economyTab.constructionTV.click = func() { d.pressLine("plan") }
	d.workerMiniTV.SetBorder(true)
	theme.Track(func() { d.workerMiniTV.SetTitleColor(theme.Of(theme.RoleAccent)) })
	d.economyTab.AddPeopleBox(d.workerMiniTV, d.richLayout, func() int {
		_, th := d.screenSize()
		return richPeopleHeight(th)
	})
	d.economyTab.richLeft = func() int {
		w, _ := d.screenSize()
		left, _, _ := richCols(w)
		return left
	}

	// Main horizontal: the economy tab (the full height) and the Panels
	// column, with the People box under the list unless the layout is rich.
	d.peopleH = workerMiniBaseRows + 2
	mainHoriz := newDashBody(d.economyTab.Root(), d.nextBox, d.workerMiniTV, d.miniMap, d.screenSize, func() int { return d.peopleH })
	mainHoriz.modelReady = func() bool { return d.miniMap.model != nil }
	d.body = mainHoriz
	mainHoriz.switched = func() { d.workerMiniTV.stale = true; d.titleNext() }

	// Content area is just mainHoriz: no bottom strip
	d.contentArea = mainHoriz

	// Root layout (no tab bar)
	d.root = newNativeFlex().SetDirection(flexRow).
		AddItem(d.headerTV, headerSmallRows, 0, false).
		AddItem(d.contentArea, 0, 1, false).
		AddItem(d.bandBox, 1, 0, false).
		AddItem(d.inputField, barSmallRows, 0, true)
	// The header is as tall as the terminal calls for, chosen before the
	// rows are shared out, so a resize shows the right one at once.
	d.pages.inset = func(name string) int {
		if d.overlayMgr != nil && name == d.overlayMgr.ActiveName() {
			return d.bandRows(d.overlayMgr.screenH)
		}
		return 0
	}
	d.root.SetDrawFunc(func(_, _, w, h int) {
		d.root.ResizeItem(d.headerTV, headerHeight(w, h), 0)
		n := barRows(w, h)
		d.root.ResizeItem(d.inputField, n, 0)
		d.root.ResizeItem(d.bandBox, d.bandRows(h), 0)
		d.inputField.SetBig(n == barBigRows)
	})

	// Global key handling
	d.root.SetInputCapture(func(event *kit.Key) *kit.Key {
		d.closeBadgeNote()
		// Ctrl+K — open passphrase modal (dev unlock)
		if event.Ctrl('k') {
			d.showDevUnlockModal()
			return nil
		}
		// An open panel that leaves the keyboard with the command bar takes
		// the keys that print nothing; the rest reach the bar, which keeps
		// the focus. Esc closes the panel's card first, where it has one.
		if d.overlayMgr != nil {
			if o, ok := d.barOverlays[d.overlayMgr.ActiveName()]; ok && d.inputField.HasFocus() &&
				(o.routeKey(event, d.inputField.GetText()) || event.Code == kit.KeyEsc && o.closeFirst != nil && o.closeFirst()) {
				return nil
			}
		}
		// The open text panels (Milestones, Stats, Help and the rest) scroll
		// with the keys that print nothing while the bar is empty.
		if d.overlayMgr != nil && d.inputField.HasFocus() && d.overlayMgr.scrollKey(event, d.inputField.GetText()) {
			return nil
		}
		switch event.Code {
		case kit.KeyChord:
			// The Resources box's next page, when it has more rows than it
			// can show. With a panel open the box is behind it.
			if event.Ctrl(resourcePageKey) && (d.overlayMgr == nil || !d.overlayMgr.HasActive()) {
				d.economyTab.NextResourcePage()
				return nil
			}
		case kit.KeyEsc:
			if d.overlayMgr != nil && d.overlayMgr.HasActive() {
				d.overlayMgr.Hide()
				return nil
			}
			if d.inputField.GetText() == "" && d.economyTab.buildBox.closeOlder() {
				return nil
			}
			d.openPause()
			return nil
		// Economy tab scroll keys (always available since economy is permanent background)
		case kit.KeyPgUp:
			if !d.overlayMgr.HasActive() {
				d.economyTab.ScrollUp()
				return nil
			}
		case kit.KeyPgDn:
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

// switchOverlay swaps the open bar panel for another, which keeps the
// keyboard with the command bar as it does when a command opens it. prep
// readies the panel first.
func (d *Dashboard) switchOverlay(name string, prep func()) {
	if prep != nil {
		prep()
	}
	d.overlayMgr.Show(name, d.engine.GetState())
	d.overlayMgr.FocusOn(d.inputField)
}

// panelLink is a panel's name as the Panels list shows it, in a region named
// for the panel: a click on it opens the panel.
func panelLink(name, shown string) string { return clickRegion(name, shown) }

// clickRegion is shown in a click region called name: a click on it tells
// the box's OnRegion the name.
func clickRegion(name, shown string) string { return `["` + name + `"]` + shown + `[""]` }

// buildLine is the command line that builds one of a building.
func buildLine(key string) string { return "build " + key }

// Root returns the root primitive for page registration: the layout, and
// over it the active theme's ambient effect.
func (d *Dashboard) Root() Primitive {
	if d.rootFx == nil {
		d.rootFx = newFxRoot(d.root, d.drawEffect, d.drawOver)
	}
	return d.rootFx
}

// drawEffect draws the active theme's ambient effect over the dashboard,
// when the theme has one and the motion setting is on.
func (d *Dashboard) drawEffect(scr mapstyle.Cells, x, y, w, h int) {
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
	now := d.clock()
	if d.fxStart.IsZero() {
		d.fxStart = now()
	}
	// Not in the command bar: nothing moves where the player types.
	h -= d.bar()
	drawThemeEffect(scr, x, y, w, h, effect, int(now().Sub(d.fxStart)/mapAnimStep))
}

// clock is the dashboard's time: time.Now, or the one a test pinned.
func (d *Dashboard) clock() func() time.Time {
	if d.fxNow != nil {
		return d.fxNow
	}
	return time.Now
}

// tipCell is the cell size the tooltip's picture is drawn at: none with no
// window.
func (d *Dashboard) tipCell() (int, int) {
	if pictureCell == nil {
		return 0, 0
	}
	return pictureCell()
}

// drawOver draws what goes over the dashboard: the tooltip, which is the last
// thing painted. A panel over the dashboard takes the tooltip away.
func (d *Dashboard) drawOver(p *kit.Painter, bounds image.Rectangle) {
	if d.overlayMgr.HasActive() {
		d.app.feel.tip.Clear()
		return
	}
	d.app.feel.tip.Paint(p, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y-d.bar()))
}

// StartUpdates begins the UI refresh loop, polling at 500 ms (2 fps).
// This is intentionally slower than the game tick rate (which can reach ~5 fps
// at 2x speed) to avoid burning CPU on redraws.
// NOTE: app.QueueUpdateDraw is the only safe way to touch the UI's primitives from
// a background goroutine: it serialises with the app loop.
func (d *Dashboard) StartUpdates() {
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		// The animation ticker runs at the rate of a key button's sink,
		// twice the rate of everything else that moves, which redraws on
		// every other tick.
		anim := time.NewTicker(animStep)
		defer anim.Stop()
		other := false
		for {
			select {
			case <-ticker.C:
				d.app.QueueUpdateDraw(func() {
					d.refresh()
				})
			case <-anim.C:
				// The open Map panel animates between refreshes: a redraw
				// is enough, since its frame counter runs on the clock.
				// So does a theme with an ambient effect, and a key button that is sinking.
				other = !other
				if d.app.feel.busy() || d.app.feel.tip.Waiting() || d.inputField.echoing() || other && (d.mapOpen.Load() || d.expPanel.moving.Load() || d.decks["trade"].moving.Load() || d.fxOn.Load()) {
					d.app.QueueUpdateDraw(func() {
						if d.overlayMgr.ActiveName() != "map" {
							d.mapOpen.Store(false)
							d.mapPanel.hide()
						}
						if d.overlayMgr.ActiveName() != "trade" {
							d.decks["trade"].moving.Store(false)
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

// titleNext titles the right column's top box: Next beside the card, Panels
// where the list is alone.
func (d *Dashboard) titleNext() {
	if d.richLayout() {
		d.nextBox.SetTitle(txtNextTitle.String())
	} else {
		d.nextBox.SetTitle(txtPanels.String())
	}
}

// retitle writes the titles on the borders of the dashboard's own boxes, in the
// script now in force. The People box's title is also written each frame, with
// the headcount in it when there is room.
func (d *Dashboard) retitle() {
	d.titleNext()
	d.logTV.SetTitle(rowid.Line("ui.dashboard.build.02", " Log "))
	d.inputField.SetTitle(rowid.Line("ui.dashboard.build.13", " Command "))
	d.workerMiniTV.SetTitle(rowid.Line("ui.dashboard.build.14", " People "))
}

// refresh is the central UI update function, called every 500 ms from StartUpdates.
// It must only be called from within app.QueueUpdateDraw (i.e. on the UI goroutine).
// Order matters: catastrophe modal and age splash are checked first because they may
// swap the active page, then all content widgets are updated.
func (d *Dashboard) refresh() {
	// The pause card owns the screen and the keyboard: a splash or modal opened
	// now would take the focus from it, and its close would hand the focus to the
	// Command box behind the dimmed town. What is pending shows after Resume.
	if d.engine.Held() {
		return
	}
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
	d.notePanels(panelsReached(state), false)
	d.next = computeNext(state, d.panels)
	d.stepLessons(state)
	if d.economyTab.buildBox.nextKey != d.next.card {
		d.economyTab.buildBox.resort = true
	}
	d.economyTab.buildBox.nextKey = d.next.card
	d.economyTab.buildBox.nextCmd = d.next.cmd
	d.nextBox.changed()

	// Badges earned since the last refresh: a card and a log line each, and a
	// line for the theme a badge gives. Here, in the UI goroutine: outside the
	// engine lock, never in a Bus handler. They wait for the film of a run's
	// ending, which covers the header, and are told when it is over.
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
	} else if d.catModalShown != pending && !d.overlayMgr.FullScreenUp() && d.inputField.GetText() == "" {
		// Not over a half-typed command: it waits for the prompt to be empty.
		if d.pages.HasPage(catastrophePage) {
			d.pages.RemovePage(catastrophePage) // stale modal for a different epoch
		}
		d.catModalShown = pending
		d.showCatastropheModal(pending)
	}

	// Ancient Memory offer modal — same single-show pattern as the catastrophe modal.
	if state.PendingMemoryTech == "" {
		d.memoryModalShown = "" // reset so a future run's cache can show fresh
	} else if d.memoryModalShown == "" && !d.overlayMgr.FullScreenUp() && d.inputField.GetText() == "" {
		// Not over the film of the run that just ended: the new run's
		// first window waits for it.
		d.memoryModalShown = state.PendingMemoryTech
		d.showAncientMemoryModal(state.PendingMemoryTech, state.PendingMemoryTechName)
	}

	// Shame badge — pick once per session, never change after that
	if state.CheaterBadge && d.shame == nil {
		pick := shameMessages[rand.Intn(len(shameMessages))]
		d.shame = &pick
	}

	if d.lastAge != state.Age {
		d.lastAge = state.Age
	}

	d.refreshAgeProgress(state)
	d.refreshLog(state)

	// Economy tab is always visible as the permanent background
	d.economyTab.Refresh(state)
	// The mini map's model, only while it is on and has room to show
	// (before the first layout the dock has no size yet, so build it
	// anyway).
	set := d.mapSettings()
	d.body.miniOff = !set.Minimap
	if d.body.wantsModel() {
		d.miniMap.update(set, &state)
		d.miniMap.pointAt(d.mapPanel)
	}

	// Update overlay content
	d.overlayMgr.Refresh(state)
	d.maybeShowPrestigeCue(state)
	d.refreshWorkerMini(state)
}

// announceBadges says what the account earned since the last refresh. Badges
// its record already proved when it was loaded get one line between them.
// Then it drains the badges the account earned since the last refresh
// (the engine judged them under its lock and only queued them) and gives each
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
		d.engine.AddLog(game.OfficeWorld, game.LevelHeadline, "success", line)
	}
	for _, v := range d.engine.DrainEarnedBadges() {
		d.engine.AddLog(game.OfficeWorld, game.LevelHeadline, "success", game.BadgeLogLine(v))
		d.badgeQ = append(d.badgeQ, v)
		if t, ok := theme.ByKey(v.RewardTheme); ok {
			d.engine.AddLog(game.OfficeWorld, game.LevelHeadline, "success", themeUnlockToast(t.Name.String()))
		}
	}
	d.pumpBadgeNote()
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
func heldByLines(ws game.WorkerState, tickMs, w int) []string {
	fits := func(plain string) bool { return w <= 0 || runeLen(plain) <= w }
	switch ws.HeldBy {
	case game.HeldHousing:
		return []string{rowid.Line("ui.dashboard.heldByLines.01", "Held by: housing")}
	case game.HeldWork:
		return []string{rowid.Line("ui.dashboard.heldByLines.07", "Held by: no work for more people")}
	case game.HeldFood:
		x, unit := rateFigure(ws.FoodNet, tickMs) // "+0.45", "/s"
		feeds := rowid.Sprintf("ui.dashboard.heldByLines.02", "feeds %d more", ws.FoodFeeds)
		for _, line := range []string{
			x + rowid.Line("ui.dashboard.heldByLines.03", " food") + unit + " " + feeds,
			x + unit + " " + feeds,
			x + " " + feeds,
			feeds,
		} {
			if fits(line) {
				return []string{rowid.Line("ui.dashboard.heldByLines.05", "Held by: food"), line}
			}
		}
		return []string{rowid.Line("ui.dashboard.heldByLines.06", "Held by: food")}
	}
	return nil
}

// heldByText is heldByLines for one line (the roster command and panel): the
// food form joins its two parts, "Held by: food (+0.45 food/s feeds 3
// more)".
func heldByText(ws game.WorkerState, tickMs int) string {
	lines := heldByLines(ws, tickMs, 0)
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

// txtPeopleBoxCount is said the same way on several screens, so it has one id.
var txtPeopleBoxCount = rowid.E("ui.dashboard.peopleBoxTitle.03", " People %s/%s ")

// peopleBoxTitle is the People box's title: " People ", with the headcount
// over the housing in it when the box is compact and it fits.
func peopleBoxTitle(state game.GameState, detail bool) string {
	if !peopleBoxCompact(state, detail) {
		return rowid.Line("ui.dashboard.peopleBoxTitle.01", " People ")
	}
	ws := state.Workers
	for _, t := range []string{
		rowid.Sprintf("ui.dashboard.peopleBoxTitle.02", " People %d/%d ", ws.TotalPop, ws.MaxPop),
		fmt.Sprintf(txtPeopleBoxCount.String(), textfmt.Number(float64(ws.TotalPop)), textfmt.Number(float64(ws.MaxPop))),
	} {
		if runeLen(t) <= sidebarW-2 {
			return t
		}
	}
	return rowid.Line("ui.dashboard.peopleBoxTitle.04", " People ")
}

// peopleBoxText writes the People box for a box w cells wide (0: any width):
// the population over the housing, one line for each kind of person in this
// age's own word with its count (the soldiers carry "· N out" for those on
// missions when it fits), and what holds the population back (with the food
// detail line when detail is set). A count too long for its line is written
// short (51.8M for 51785785). The food use and net are on the People panel.
func peopleBoxText(state game.GameState, w int, detail bool) string {
	ws := state.Workers
	var sb strings.Builder

	// On a short terminal with a Held-by line up, the headcount moves to the
	// box's title (peopleBoxTitle) and its row goes to the Panels list.
	if !peopleBoxCompact(state, detail) {
		pop, housing := fmt.Sprintf("%d", ws.TotalPop), fmt.Sprintf("%d", ws.MaxPop)
		popColor := "yellow"
		if ws.TotalPop > ws.MaxPop {
			popColor = "red" // the scholars moved in past the homes
		}
		// The line is measured as it is drawn, in whichever script it is written.
		popLine := func() string {
			return rowid.Sprintf("ui.dashboard.peopleBoxText.02", "[gray]Pop[white] [%s]%s[white]/[green]%s[-]\n", popColor, pop, housing)
		}
		if w > 0 && visibleLen(strings.TrimRight(popLine(), "\n")) > w {
			pop, housing = textfmt.Number(float64(ws.TotalPop)), textfmt.Number(float64(ws.MaxPop))
		}
		fmt.Fprint(&sb, popLine())
	}

	for _, k := range peopleKinds(state) {
		sb.WriteString(k.boxLine(w))
		sb.WriteString("\n")
	}
	if lines := heldByLines(ws, state.TickIntervalMs, w); len(lines) > 0 {
		fmt.Fprintf(&sb, "[gray]%s[-]\n", lines[0])
		if len(lines) > 1 && detail {
			fmt.Fprintf(&sb, "[gray]%s[-]\n", lines[1])
		}
	}
	return sb.String()
}

// headerView reads what the header shows from the last refresh, for a header
// w cells wide: the facts of statusLine, the cheater badge before any alert,
// the toast on show as the notice, and the Ledger's rows under them.
func (d *Dashboard) headerView(w int) headerView {
	if d.lastState == nil {
		return headerView{}
	}
	year := ""
	if d.ledger.ready {
		year = d.ledger.data.v.YearText
	}
	v := statusLine(*d.lastState, year)
	if d.shame != nil {
		msg := kit.Strip(safeTags(d.shame.String()))
		v.Alerts = append([]headerAlert{{text: msg, short: fitText(msg, 14), k: xBad}}, v.Alerts...)
	}
	v.Notice = strings.TrimSpace(kit.Strip(safeTags(d.toastMgr.Current())))
	if d.ledger.ready {
		v.Led = &d.ledger.data
	}
	return v
}

// moraleKind is the kind a morale band is drawn in: green for a bonus, red
// for a penalty, plain for neither.
func moraleKind(b moraleBand) xk {
	switch {
	case b.Bonus:
		return xGood
	case b.Penalty:
		return xBad
	}
	return xText
}

// plainFact is the text a rowid line wrote, without its colour tags and the
// spaces that set it off from what was beside it.
func plainFact(s string) string { return strings.TrimSpace(kit.Strip(s)) }

// statusLine is the facts of the header for a snapshot: the account's name
// and titles, the age, the calendar year (year; "" when there is nothing to
// say), Prestige, DEV, the people, morale and any alert, each in the lengths
// the header gives up in (dash_header.go).
func statusLine(state game.GameState, year string) headerView {
	var v headerView
	v.Age = state.AgeName
	if state.Prestige.Level > 0 {
		v.Prestige = plainFact(rowid.Sprintf("ui.dashboard.statusLine.01", "  [cyan]Prestige %d[-]", state.Prestige.Level))
	}
	if game.DevModeActive {
		v.Dev = plainFact(rowid.Line("ui.dashboard.statusLine.02", "  [red]DEV[-]"))
	}
	// Colour morale by the continuous production multiplier, not the raw percent:
	// green = bonus (mult>1.0), white = neutral (==1.0), red = penalty (<1.0).
	// computeMoraleBand is the shared source of truth (same as the workers panel).
	mBand := computeMoraleBand(state.Morale, state.MoraleMultiplier)
	v.MoraleK = moraleKind(mBand)
	// Pending catastrophe badge: persistent until the player chooses, so a player
	// who closed the modal (or came back after a long idle) sees why advancing
	// is blocked and how to reopen the choice.
	if state.PendingCatastrophe != "" {
		v.Alerts = append(v.Alerts, headerAlert{rowid.Line("ui.dashboard.statusLine.04", "☄ Catastrophe: type catastrophe"), rowid.Line("ui.dashboard.statusLine.05", "☄ Catastrophe"), "catastrophe", xChipMil})
	} else if state.LastPassage.Pending {
		v.Alerts = append(v.Alerts, headerAlert{rowid.Line("ui.dashboard.statusLine.07", "☄ Last Passage: type catastrophe"), rowid.Line("ui.dashboard.statusLine.08", "☄ Last Passage"), "catastrophe", xChipMil})
	}
	// Harbinger badge: present until the doom it warns of resolves. Nothing
	// expires, so the badge is the idle player's reminder that there is a
	// choice waiting (it never blocks anything).
	if state.Harbinger != nil {
		v.Alerts = append(v.Alerts, headerAlert{rowid.Line("ui.dashboard.statusLine.10", "⚑ Harbinger: type harbinger"), rowid.Line("ui.dashboard.statusLine.11", "⚑ Harbinger"), "harbinger", xChipLegend})
	}
	if mBand.DeltaLabel != "" {
		v.Delta = headFact{plainFact(rowid.Sprintf("ui.dashboard.statusLine.12", " (production [%s]%s[-])", mBand.Color, mBand.DeltaLabel)), mBand.DeltaLabel}
	}
	v.Festival = festivalChip(state)
	pct := fmt.Sprintf("%.0f%%", state.Morale*100)
	v.Morale = headFact{plainFact(rowid.Sprintf("ui.dashboard.statusLine.13", "  Morale [%s]%.0f%%[-]%s", mBand.Color, state.Morale*100, "")), pct}
	// The name is cut to what a bar can hold; the title the account wears is
	// shown once it is not the one every account starts with.
	if st := state.AccountStats; st != nil && st.DisplayName != "" {
		v.Name = st.DisplayName
		if r := []rune(v.Name); len(r) > 20 {
			v.Name = string(r[:19]) + "…"
		}
		if sum := st.BadgeSummary; sum.TitleRank > 0 || wornTitle(sum) != sum.Title {
			v.Worn = wornTitle(sum)
		}
	}
	v.Honour = state.Milestones.CurrentTitle
	if year != "" {
		v.Year = rowid.Sprintf("ui.dashboard.statusLine.19", "Year %s", year)
	}
	pop := groupedCount(state.Workers.TotalPop) + "/" + groupedCount(state.Workers.MaxPop)
	short := textfmt.Number(float64(state.Workers.TotalPop)) + "/" + textfmt.Number(float64(state.Workers.MaxPop))
	v.People = headFact{rowid.Sprintf("ui.dashboard.statusLine.17", "People %s", pop), pop}
	v.PeopleShort = headFact{strings.Replace(v.People.Text, pop, short, 1), short}
	return v
}

// txtExpeditions is said the same way on several screens, so it has one id.
var txtExpeditions = rowid.E("ui.badge_list.badgeFamilyTitles.19", "Expeditions")

// wholeNumber writes a count in full, its thousands set off with commas
// ("1,245"): a score is exact, and reads wrong rounded to "1.25K".
func wholeNumber(n int) string {
	s := strconv.Itoa(max(n, -n))
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if n < 0 {
		return "-" + s
	}
	return s
}

// groupedCount writes a whole count with a comma between its thousands
// (1,094), and a count of a million or more as a short number (41.2M).
func groupedCount(n int) string {
	if n < 1000 || n >= 1_000_000 {
		return textfmt.Number(float64(n))
	}
	return wholeNumber(n)
}

func (d *Dashboard) refreshAgeProgress(state game.GameState) {
	// The Ledger is worked out here, once, from the snapshot: the header
	// reads its year, its bar and what it waits on, and is written from it at
	// its next layout. (The notice is the toast on show, which a refresh
	// also moves on.)
	d.ledger.update(state, d.engine, d.next.card)
	d.headerTV.changed()
}

func (d *Dashboard) refreshLog(state game.GameState) {
	var sb strings.Builder
	// The town level of the log plus the game office's replies, folded, with
	// a count after a line that happened more than once.
	// In the first age the routine lines show too: a gather, a build and a
	// research each leave a trace while the log is quiet.
	i, ok := state.Ruleset().Index(state.Age)
	firstAge := ok && i == 0
	var visible []game.LogEntry
	for _, entry := range state.Log {
		if !entry.Seam && entry.Type != "debug" && (entry.Level <= game.LevelTown || entry.Office == game.OfficeGame || (firstAge && entry.Type == game.LogRoutine)) {
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
		times := ""
		if entry.Count > 1 {
			times = fmt.Sprintf(" ×%d", entry.Count)
		}
		fmt.Fprintf(&sb, "[%s]%s%s[-]\n", color, entry.Message, times)
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
	field := newNativeInput().
		SetLabel("").
		SetFieldWidth(40).
		SetMaskCharacter('·').
		SetFieldBackgroundColor(theme.Of(theme.RoleBackground)).
		SetFieldTextColor(theme.Of(theme.RoleText))

	field.SetDoneFunc(func(key kit.KeyCode) {
		if key == kit.KeyEsc {
			d.pages.RemovePage(devUnlockPage)
			d.app.SetFocus(d.inputField)
			return
		}
		if key == kit.KeyEnter {
			input := field.GetText()
			d.pages.RemovePage(devUnlockPage)
			if game.CheckDevKey(input) {
				d.engine.NoteDevUnlocked()
				d.engine.AddLog(game.OfficeGame, game.LevelEverything, "info", rowid.Line("ui.dashboard.showDevUnlockModal.01", "[red]Dev mode on.[-] Prefix commands with / (for example /god, /fill, /give wood 9999)."))
				d.app.SetFocus(d.inputField)
			} else {
				d.app.SetFocus(d.inputField)
			}
		}
	})

	box := newNativeFlex().SetDirection(flexRow).
		AddItem(field, 1, 0, true)
	box.SetBorder(true).
		SetBorderColor(theme.Of(theme.RoleDim)).
		SetBackgroundColor(theme.Of(theme.RoleBackground))

	d.pages.AddPage(devUnlockPage, floatingModal(box, 44, 3), true, true)
	d.app.SetFocus(field)
}

// accountNoticePage is the page of the notice leaveToMenu shows over the main menu.
const accountNoticePage = "account_notice"

// txtOK is said the same way on several screens, so it has one id.
var txtOK = rowid.E("ui.dashboard.leaveToMenu.02", "OK")

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
	showButtonModal(d.app, d.pages, accountNoticePage, safeTags(msg), []string{txtOK.String()}, func(int, string) {})
}

// codeWindowPage is the page of the small window a command's Window text is
// shown in.
const codeWindowPage = "code_window"

// showWindow shows text in a small window over the dashboard, titled with the
// recovery code's label, and closes it on Esc or Enter. The text lives in the
// window only: it is never added to the game log.
func (d *Dashboard) showWindow(text string) {
	tv := newNativeText().
		SetDynamicColors(true).
		SetTextAlign(kit.AlignCenter)
	tv.SetBorder(true).SetTitle(rowid.Line("ui.dashboard.showWindow.01", " Recovery code "))
	tv.SetText("\n" + kit.Escape(text) + "\n\n" + theme.KeycapButton("Esc", rowid.Line("ui.dashboard.showWindow.03", "Close")))
	done := func() {
		d.pages.RemovePage(codeWindowPage)
		d.app.SetFocus(d.inputField)
	}
	tv.SetInputCapture(func(ev *kit.Key) *kit.Key {
		if ev.Code == kit.KeyEsc || ev.Code == kit.KeyEnter {
			done()
			return nil
		}
		return nil
	})
	d.pages.AddPage(codeWindowPage, floatingModal(tv, 49, 7), true, true)
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
		// Enter on an empty prompt runs what the prompt shows dim: the picked
		// card's command, unless the pointer is on a button whose own command
		// is shown in its place.
		if line := d.cardLine(); line != "" && d.app.feel.preview == "" {
			d.pressLine(line)
		}
		return
	}
	d.remember(text)
	d.lessonTyped++
	d.runLine(text)
}

// cardLine is the command of the picked build card, the line an empty prompt
// shows dim and Enter runs; "" when no card is picked or one cannot be built,
// and while a panel covers the dashboard.
func (d *Dashboard) cardLine() string {
	if d.overlayMgr != nil && d.overlayMgr.HasActive() {
		return ""
	}
	return d.economyTab.buildBox.pickedLine()
}

// remember records a non-empty trimmed command in history (cap at 50).
func (d *Dashboard) remember(text string) {
	if cmd := strings.TrimSpace(text); cmd != "" {
		if n := len(d.cmdHistory); n > 0 && d.cmdHistory[n-1] == cmd {
			return // a line run again is one place in the history
		}
		if len(d.cmdHistory) >= 50 {
			d.cmdHistory = d.cmdHistory[1:] // drop oldest
		}
		d.cmdHistory = append(d.cmdHistory, cmd)
	}
}

// pressLine runs a line a button stands for, as typing it and pressing Enter
// would: the command bar shows it for a moment, it enters the history, and it
// runs through runLine. It is the one way a click or a panel's key runs a
// command.
func (d *Dashboard) pressLine(line string) {
	d.inputField.echoLine(line)
	d.remember(line)
	d.runLine(line)
}

var txtCancel = rowid.E("ui.dashboard.askLine.01", "Cancel")

// askLine puts a question over the dashboard with a way out and a button
// worded yes, which runs the line as a click on a button does. For a line the
// command table marks Dangerous the yes is the danger button and the keyboard
// starts on the way out; for any other the yes is the primary button, which
// Enter presses. Either answer gives the keyboard back.
func (d *Dashboard) askLine(question, yes, line string) {
	const page = "ask_line"
	danger := lineDangerous(line)
	buttons, run := []string{yes, txtCancel.String()}, 0
	if danger {
		buttons, run = []string{txtCancel.String(), yes}, 1
	}
	m := showButtonModal(d.app, d.pages, page, question, buttons, func(i int, _ string) {
		if i == run {
			d.pressLine(line)
		}
		d.returnFocus()
	})
	if danger {
		styleDangerModal(m)
	}
}

// runLine does what a command line asks, as Enter on it does. It leaves the
// prompt and its history alone: they are the typist's.
func (d *Dashboard) runLine(text string) {
	cmd := strings.TrimSpace(text)
	if strings.ToLower(cmd) == "quit" {
		// The same way out as Ctrl+C and a signal: the game in play is
		// stopped and saved to its own save. (A failure has nowhere to be
		// shown: the screen goes with the app.)
		_, _ = d.engine.SaveOnExit()
		d.app.Stop()
		return
	}
	if d.screenLine(cmd) {
		return
	}
	// Route /commands to dev exec when dev mode is active
	if game.DevModeActive && strings.HasPrefix(cmd, "/") {
		result := game.DevConsoleCommand(cmd, d.engine)
		if result != "" {
			d.engine.AddLog(game.OfficeGame, game.LevelEverything, "info", "[positive]dev → "+result+"[-]")
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
	if fields := strings.Fields(cmd); len(fields) > 0 {
		if name := panelNamed(fields[0]); name != "" {
			d.notePanels([]string{name}, true)
		}
	}
	result := HandleCommand(text, d.engine)
	if result.EpochCard {
		result = epochCardReply(d.epochHall.card(d.engine.GetState()))
	}
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
	if result.OverlayName == "map" {
		d.mapPanel.world = result.MapWorld
		d.mapOpen.Store(true)
	}
	if result.OverlayName == "tree" {
		d.treePanel.request(result.ResearchZoom, result.ResearchCard)
	}
	if result.OverlayName == "logs" {
		d.logsPanel.request(result.Logs)
	}
	if result.OverlayName == "badges" {
		d.badgePanel.request(result.Badges)
	}
	if result.MapPref.Key != "" {
		d.applyMapPref(result.MapPref)
		d.body.miniOff = !d.mapSettings().Minimap
	}
	if result.MapFlows != "" {
		result.Message = flowsReply(d.mapPanel.setFlows(result.MapFlows))
	}
	if result.OverlayName != "" {
		state := d.engine.GetState()
		d.overlayMgr.Show(result.OverlayName, state)
		if _, ok := d.barOverlays[result.OverlayName]; ok || d.overlayMgr.keepsBar(result.OverlayName) {
			// The command bar keeps the keyboard while the map is open.
			d.overlayMgr.FocusOn(d.inputField)
		}
	} else if o, ok := d.barOverlays[d.overlayMgr.ActiveName()]; ok {
		// A command typed over one of those panels: say what happened on
		// its key bar where the log is behind it, and show the change now
		// rather than at the next refresh.
		if o.reply != nil {
			o.reply(cmd, result)
		}
		o.update(d.engine.GetState())
	}
	if result.Message != "" && result.Type != "success" {
		// A badge this command earned is announced first, so the reply, which the
		// player asked for, is the last line in the log and stays in view.
		d.announceBadges()
		d.engine.AddLog(game.OfficeGame, game.LevelEverything, result.Type, result.Message)
	}
}

// screenLine runs the commands that are the dashboard's own (buy, pause,
// settings, menu, new) and reports whether line was one.
func (d *Dashboard) screenLine(line string) bool {
	f := strings.Fields(strings.ToLower(line))
	if len(f) == 0 {
		return false
	}
	switch f[0] {
	case "buy":
		if len(f) == 2 {
			if i := slices.Index(buyArgs[:], f[1]); i >= 0 {
				d.economyTab.buildBox.setBuy(buyChoice(i))
				return true
			}
		}
		d.engine.AddLog(game.OfficeGame, game.LevelEverything, "error", usageFor("buy"))
	case "pause", "settings":
		if d.pages.HasPage(pausePage) {
			break
		}
		if p := d.openPause(); f[0] == "settings" {
			p.go2(pauseDisplay)
		}
	case "menu", "new":
		d.toMenu()
	default:
		return false
	}
	return true
}

// returnFocus gives the keyboard back when a window over the dashboard
// closes: to a window that waits on a choice (the catastrophe's, an Ancient
// Memory's), else to the open panel, else to the prompt. An arrival that
// came up over an Ancient Memory left unanswered (a planned advance goes by
// itself) would otherwise leave that window in front with no key able to
// reach it.
func (d *Dashboard) returnFocus() {
	switch {
	case d.catFocus != nil && d.pages.HasPage(catastrophePage):
		d.pages.SendToFront(catastrophePage)
		d.app.SetFocus(d.catFocus)
	case d.memFocus != nil && d.pages.HasPage(ancientMemoryPage):
		d.pages.SendToFront(ancientMemoryPage)
		d.app.SetFocus(d.memFocus)
	case d.overlayMgr != nil && d.overlayMgr.Focus():
	default:
		d.app.SetFocus(d.inputField)
	}
}

// takeNext writes the next step's command on the empty line as typed text,
// and reports whether it did. Nothing runs until Enter.
func (d *Dashboard) takeNext() bool {
	if d.inputField.GetText() != "" || d.next.cmd == "" || (d.overlayMgr != nil && d.overlayMgr.HasActive()) {
		return false
	}
	d.inputField.setText(d.next.cmd)
	return true
}

// showMe takes the player to a step: the card of a building, picked and in
// view as after a click, or a panel, opened. It builds and starts nothing.
func (d *Dashboard) showMe(card, panel string) {
	if card != "" {
		d.economyTab.buildBox.show(card)
		return
	}
	if panel != "" {
		d.pressLine(panel)
	}
}
