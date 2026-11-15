package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// shameMessages are randomly chosen at session start when the cheater badge is active.
// Displayed in the top status bar — a deterrent against savegame manipulation.
var shameMessages = []vibe.Entry{
	vibe.E("ui.dashboard.shameMessages.01", "✦ FORGED SCROLLS ✦"),
	vibe.E("ui.dashboard.shameMessages.02", "✦ ILLEGITIMATE EMPIRE ✦"),
	vibe.E("ui.dashboard.shameMessages.03", "✦ COUNTERFEIT DYNASTY ✦"),
	vibe.E("ui.dashboard.shameMessages.04", "✦ USURPER'S THRONE ✦"),
	vibe.E("ui.dashboard.shameMessages.05", "✦ THE FABRICATED AGE ✦"),
	vibe.E("ui.dashboard.shameMessages.06", "✦ DECREE OF DISHONOR ✦"),
}

// Dashboard is the main gameplay screen. The economy tab is the always-visible
// background; named overlays (research, trade, military, etc.) are rendered on
// top of it via the page stack. Only one overlay can be visible at a time.
// promptRows is the command bar's height: one line inside its border. The
// Map panel leaves these rows free, so the prompt works while it is open.
const promptRows = 3

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
	contentArea         Primitive

	// Shame badge — set once on first load when CheaterBadge is true; the
	// header shows it as an alert.
	shame *vibe.Entry // the message picked for the session; read as it is now each frame

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

	// harbPanel is the Harbinger panel's UI state (feedback line, Invite
	// confirmation). Owned by the UI goroutine.
	harbPanel harbingerPanel
	// planPanel is the Plan panel's UI state (selection, feedback line,
	// Clear confirmation). Owned by the UI goroutine.
	planPanel planPanel

	// The maps: the shared model builder, the Map panel and the dashboard's
	// mini map. UI goroutine only. mapLocal holds the
	// map settings when no account is loaded (session only).
	mapViews  *mapViews
	mapPanel  *mapPanel
	wonders   *wondersView
	epochHall *epochHall
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
		badgePanel:    newBadgePanel(),
		expPanel:      newExpeditionPanel(),
		peoplePanel:   newPeoplePanel(),
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
		d.updateSidebar("")
		d.returnFocus()
	})
	// The text panels' frames fold to ASCII with the map glyphs.
	d.overlayMgr.plain = func() bool { return d.mapSettings().Tier == mapmodel.TierASCII }
	d.registerDeck("milestones", vibe.E("ui.dashboard.NewDashboard.01", "Milestones"), milestonesPage).draw = renderMileRoads
	// The Research panel: the tech tree as a map. Like the Map it leaves
	// the keyboard with the command bar.
	d.researchPanel.engine = d.engine
	d.researchPanel.tier = func() mapmodel.GlyphTier { return d.mapSettings().Tier }
	d.researchPanel.prompt = func() string { return d.inputField.GetText() }
	d.researchPanel.toPrompt = d.handToPrompt
	d.overlayMgr.RegisterWidget("techs", vibe.E("ui.dashboard.NewDashboard.02", "Research"), d.researchPanel.open, d.researchPanel.update)
	// The badge case: the account's badges. It leaves the keyboard with the
	// command bar too.
	d.badgePanel.settings = d.mapSettings
	d.badgePanel.prompt = d.researchPanel.prompt
	d.badgePanel.toPrompt = d.handToPrompt
	d.badgePanel.buttons.press = d.app.Key
	d.overlayMgr.RegisterWidget("badges", vibe.E("ui.dashboard.NewDashboard.03", "Badges"), d.badgePanel.open, d.badgePanel.update)
	d.registerDeck("army", vibe.E("ui.dashboard.NewDashboard.04", "Army"), armyPage).draw = renderArmyField
	// The Expeditions panel leaves the keyboard with the command bar too.
	d.expPanel.engine = d.engine
	d.expPanel.settings = d.mapSettings
	d.expPanel.prompt = d.researchPanel.prompt
	d.expPanel.toPrompt = d.handToPrompt
	d.expPanel.buttons.press = d.app.Key
	d.overlayMgr.RegisterWidget("expedition", txtExpeditions, d.expPanel.open, d.expPanel.update)
	d.registerDeck("trade", vibe.E("ui.dashboard.NewDashboard.06", "Trade routes"), tradePage).draw = renderTradeRoads
	// The Factions panel, under two names pointing at one provider (the
	// citymap/map precedent below). "factions" is the primary — it is what the
	// sidebar lists, and the sidebar highlight matches on the registered name, so
	// the two strings have to agree exactly. Bare `diplomacy` is routed to the
	// primary name in input.go so it highlights too; "diplomacy" stays registered
	// so anything holding the old overlay name still resolves.
	d.overlayMgr.Register("factions", vibe.E("ui.dashboard.NewDashboard.07", "Factions"), factionsProvider)
	d.overlayMgr.Register("diplomacy", vibe.E("ui.dashboard.NewDashboard.08", "Factions"), factionsProvider)
	d.overlayMgr.Register("stats", vibe.E("ui.dashboard.NewDashboard.09", "Statistics"), statsProvider)
	d.wonders = &wondersView{}
	d.overlayMgr.Register("wonders", vibe.E("ui.dashboard.NewDashboard.10", "Wonders"), d.wonders.provider)
	d.overlayMgr.SetOpenHandler("wonders", d.wonders.open)
	d.overlayMgr.SetBackHandler("wonders", d.wonders.back)
	d.overlayMgr.SetFooter("wonders", d.wonders.footer)
	d.overlayMgr.SetKeyHandler("wonders", func(event *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held (see the harbinger handler).
		if !d.wonders.handleKey(event) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	// The People panel, which `workers` opens too, leaves the keyboard with
	// the command bar as well.
	d.peoplePanel.settings = d.mapSettings
	d.peoplePanel.toPrompt = d.handToPrompt
	d.peoplePanel.buttons.press = d.app.Key
	d.overlayMgr.RegisterWidget("people", vibe.E("ui.dashboard.NewDashboard.11", "People"), d.peoplePanel.open, d.peoplePanel.update)
	d.registerAdvancePanel()
	d.overlayMgr.Register("logs", vibe.E("ui.dashboard.NewDashboard.12", "Logs"), logsProvider)
	d.epochHall = &epochHall{
		tier:   func() mapmodel.GlyphTier { return d.mapSettings().Tier },
		height: func() int { return d.overlayMgr.screenH },
		motion: func() bool { return d.mapSettings().Motion },
	}
	d.epochHall.state = func() game.GameState { return d.engine.GetState() }
	d.epochHall.card = func(st game.GameState) (epochCardResult, error) { return makeEpochCard(st, copyImageToClipboard) }
	d.overlayMgr.Register("epoch", vibe.E("ui.dashboard.NewDashboard.13", "Epoch"), d.epochHall.provider)
	// The Badges panel and the epoch screen draw their coins from one set
	// of pictures.
	coins := &coinCache{}
	d.badgePanel.setCoins(coins)
	d.epochHall.pics.cache = coins
	d.overlayMgr.SetOver("epoch", d.epochHall.paintCoins)
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
	d.overlayMgr.Register("harbinger", vibe.E("ui.dashboard.NewDashboard.14", "Harbinger"), d.harbPanel.provider)
	d.overlayMgr.SetFooter("harbinger", harbingerFooter)
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
	d.planPanel.height = func() int { return d.overlayMgr.screenH }
	d.overlayMgr.Register("plan", txtBuildPlan, d.planPanel.provider)
	d.overlayMgr.SetFooter("plan", planFooter)
	d.overlayMgr.SetKeyHandler("plan", func(event *kit.Key) *kit.Key {
		// UI goroutine, no engine lock held (see the harbinger handler).
		if !d.planPanel.handleKey(event, d.engine) {
			return event
		}
		d.overlayMgr.Refresh(d.engine.GetState())
		return nil
	})
	d.overlayMgr.Register("history", vibe.E("ui.dashboard.NewDashboard.16", "Civilization History"), historyProvider)
	d.overlayMgr.Register("buildings", txtBuildings, buildingsProvider)
	d.vibeSeen = vibe.On()
	d.overlayMgr.Register("help", vibe.E("ui.dashboard.NewDashboard.18", "Help"), helpProvider)

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
	d.overlayMgr.RegisterWidget("map", vibe.E("ui.dashboard.NewDashboard.19", "Map"), d.mapPanel.open, d.mapPanel.update)
	d.barOverlays = d.newBarOverlays()

	return d
}

func (d *Dashboard) build() {
	// Create permanent economy tab
	d.economyTab = NewEconomyTab()
	d.economyTab.plain = func() bool { return d.mapSettings().Tier == mapmodel.TierASCII }

	// Sidebar — command panel hints, laid out for the room it has: one
	// column when the names fit that way, two when the screen is short.
	d.sidebar = newFitView(func(w, h int) string { return sidebarText(d.sidebarActive, w, h) })
	// A click on a panel's name opens it, as typing the name does.
	d.sidebar.body.OnRegion = func(name string, _ int) { d.runLine(name) }
	// A click on a building writes its build command in the prompt, and a
	// double click runs it; a click on the mini map opens the Map panel.
	d.economyTab.buildingTV.body.OnRegion = d.clickLine
	d.miniMap.onClick = func() { d.runLine("map") }
	d.sidebar.SetBorder(true)
	theme.Track(func() { d.sidebar.SetTitleColor(theme.Of(theme.RoleAccent)) })

	// Log panel
	d.logTV = newNativeText().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true)
	// A click on the Log opens the full record.
	d.logTV.click = func() { d.runLine("logs") }
	theme.Track(func() { d.logTV.SetTitleColor(theme.Of(theme.RoleDim)) })

	// The header at the top: the account, the age, the people, the way to
	// the next age (headerView).
	d.toastMgr = NewToastManager()
	d.headerTV = newHeaderBox(d.headerView, func() bool { return headerPlain(d.mapSettings()) }, d.runLine, d.app.Key)

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
		d.toastMgr.Show(vibe.Line("ui.dashboard.build.03", "Age advanced: ")+game.AgeName(newAge), "gold", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventResearchDone, func(e game.EventData) {
		tech, _ := e.Payload["tech"].(string)
		d.toastMgr.Show(vibe.Line("ui.dashboard.build.04", "Research complete: ")+game.TechName(tech), "cyan", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventBuildingBuilt, func(e game.EventData) {
		building, _ := e.Payload["building"].(string)
		// Only toast for wonders — look up from config, not engine state (avoids deadlock)
		if def, ok := config.BuildingByKey()[building]; ok && def.Category == "wonder" {
			d.toastMgr.Show(vibe.Sprintf("ui.dashboard.build.05", "Wonder built: %s", def.Name), "green", 4*time.Second)
		}
	})
	d.engine.Bus.Subscribe(game.EventMilestoneCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		rewardText, _ := e.Payload["reward_text"].(string)
		// note: flavor quip rides in the log line, not the height-1 toast — see engine.checkMilestones.
		msg := vibe.Line("ui.dashboard.build.06", "Milestone: ") + name
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
		d.toastMgr.Show(vibe.Sprintf("ui.dashboard.build.07", "%s %s dawns", epochIcon, epochName), "gold", 6*time.Second)
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
		verb := vibe.Line("ui.dashboard.build.08", "has come")
		if handoff, _ := e.Payload["handoff"].(bool); handoff {
			verb = vibe.Line("ui.dashboard.build.09", "takes up the warning")
		}
		d.toastMgr.Show(vibe.Sprintf("ui.dashboard.build.10", "⚑ %s %s. Type harbinger to read the warning.", textfmt.Capitalize(name), verb), "warning", 8*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventGameLoaded, func(e game.EventData) {
		// Runs under the engine write lock: only flip the flag, never touch the engine.
		d.catReshow.Store(true)
	})
	d.engine.Bus.Subscribe(game.EventEpochEventFired, func(e game.EventData) {
		eventName, _ := e.Payload["event_name"].(string)
		eventType, _ := e.Payload["event_type"].(string)
		if eventType == "catastrophe" {
			d.toastMgr.Show(vibe.Sprintf("ui.dashboard.build.11", "☄ Catastrophe: %s. Type catastrophe to choose.", eventName), "red", 8*time.Second)
			return
		}
		color := "cyan"
		if eventType == "bad_challenging" {
			color = "red"
		} else if eventType == "good_legendary" {
			color = "gold"
		}
		d.toastMgr.Show(vibe.Sprintf("ui.dashboard.build.12", "Epoch event: %s", eventName), color, 6*time.Second)
	})

	// Command input: completions show as ghost text (command_input.go),
	// drawn from the registry and the state of the last refresh.
	d.inputField = newCommandInput(newCompleter(d.engine, func() *game.GameState { return d.lastState }))
	d.inputField.SetLabel("❯ ").SetFieldWidth(0)
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
	// A click on the People box opens the People panel, and one on Under
	// construction the build plan.
	d.workerMiniTV.click = func() { d.runLine("people") }
	d.economyTab.constructionTV.click = func() { d.runLine("plan") }
	d.workerMiniTV.SetBorder(true)
	theme.Track(func() { d.workerMiniTV.SetTitleColor(theme.Of(theme.RoleAccent)) })
	d.economyTab.AddPeopleBox(d.workerMiniTV, d.richLayout, func() int {
		_, th := d.screenSize()
		return richPeopleHeight(th)
	})

	// Main horizontal: the economy tab (the full height) and the Panels
	// column, with the People box under the list unless the layout is rich.
	d.peopleH = workerMiniBaseRows + 2
	mainHoriz := newDashBody(d.economyTab.Root(), d.sidebar, d.workerMiniTV, d.miniMap, d.screenSize, func() int { return d.peopleH })
	mainHoriz.modelReady = func() bool { return d.miniMap.model != nil }
	d.body = mainHoriz
	mainHoriz.switched = func() { d.workerMiniTV.stale = true }

	// Content area is just mainHoriz: no bottom strip
	d.contentArea = mainHoriz

	// Root layout (no tab bar)
	d.root = newNativeFlex().SetDirection(flexRow).
		AddItem(d.headerTV, headerSmallRows, 0, false).
		AddItem(d.contentArea, 0, 1, false).
		AddItem(d.inputField, promptRows, 0, true)
	// The header is as tall as the terminal calls for, chosen before the
	// rows are shared out, so a resize shows the right one at once.
	d.root.SetDrawFunc(func(_, _, w, h int) { d.root.ResizeItem(d.headerTV, headerHeight(w, h), 0) })

	// Global key handling
	d.root.SetInputCapture(func(event *kit.Key) *kit.Key {
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
			d.engine.SaveGame(d.engine.ActiveSaveName())
			d.engine.Stop()
			d.pages.SwitchToPage("splash")
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

func (d *Dashboard) updateSidebar(activeOverlay string) {
	if d.sidebar != nil && d.sidebarActive != activeOverlay {
		d.sidebarActive = activeOverlay
		d.sidebar.changed()
	}
}

// panelLink is a panel's name as the Panels list shows it, in a region named
// for the panel: a click on it opens the panel.
func panelLink(name, shown string) string { return clickRegion(name, shown) }

// clickRegion is shown in a click region called name: a click on it tells
// the box's OnRegion the name.
func clickRegion(name, shown string) string { return `["` + name + `"]` + shown + `[""]` }

// buildLine is the command line that builds one of a building.
func buildLine(key string) string { return "build " + key }

// sidebarPanels is the Panels list, in order.
var sidebarPanels = []string{"milestones", "badges", "research", "plan", "charter", "expedition", "army", "trade", "factions", "stats", "wonders", "people", "logs", "epoch", "harbinger", "history", "map", "help"}

// buildSidebarText is the Panels list in one column, the open panel
// highlighted: the list as it shows when the screen is tall enough for it.
func buildSidebarText(active string) string {
	var sb strings.Builder
	sb.WriteString("\n")
	for _, cmd := range sidebarPanels {
		if cmd == active {
			sb.WriteString(" " + panelLink(cmd, theme.Selected(fmt.Sprintf(" %-10s ", cmd))) + "\n")
		} else {
			sb.WriteString(" " + panelLink(cmd, fmt.Sprintf("[white]%-10s[-]", cmd)) + "\n")
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
			return panelLink(cmd, theme.Selected(padded))
		}
		return panelLink(cmd, "[white]"+padded+"[-]")
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
func (d *Dashboard) Root() Primitive {
	if d.rootFx == nil {
		d.rootFx = newFxRoot(d.root, d.drawEffect)
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
	now := time.Now
	if d.fxNow != nil {
		now = d.fxNow
	}
	if d.fxStart.IsZero() {
		d.fxStart = now()
	}
	// Not in the command bar: nothing moves where the player types.
	h -= promptRows
	drawThemeEffect(scr, x, y, w, h, effect, int(now().Sub(d.fxStart)/mapAnimStep), set.Tier == mapmodel.TierASCII)
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
		// The animation ticker runs at the rate of the badge case's coins,
		// twice the rate of everything else that moves, which redraws on
		// every other tick.
		anim := time.NewTicker(coinStep)
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
				// So does the badge case, while it shows a badge that moves,
				// and a theme with an ambient effect.
				other = !other
				if d.badgePanel.moving.Load() || d.epochHall.pics.moving.Load() || other && (d.mapOpen.Load() || d.expPanel.moving.Load() || d.fxOn.Load()) {
					d.app.QueueUpdateDraw(func() {
						if d.overlayMgr.ActiveName() != "map" {
							d.mapOpen.Store(false)
							d.mapPanel.hide()
						}
						if d.overlayMgr.ActiveName() != "badges" {
							d.badgePanel.moving.Store(false)
						}
						if d.overlayMgr.ActiveName() != "epoch" {
							d.epochHall.pics.moving.Store(false)
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
// been switched since (by the command, or by an account that was switched to):
// the titles on the borders of its boxes and of the panels. The text the panels
// draw from the game's state follows on its own: the engine moved onto the other
// ruleset when the switch changed, and what the panels write each frame reads
// the script then.
func (d *Dashboard) syncVibe() {
	if on := vibe.On(); on != d.vibeSeen {
		d.vibeSeen = on
		d.retitle()
		d.economyTab.retitle()
		// A box that keeps the text it wrote for its size writes it again.
		for _, v := range []*fitView{d.sidebar, d.workerMiniTV,
			d.economyTab.resourceTV, d.economyTab.buildingTV, d.economyTab.constructionTV} {
			v.changed()
		}
		d.headerTV.changed()
	}
}

// retitle writes the titles on the borders of the dashboard's own boxes, in the
// script now in force. The People box's title is also written each frame, with
// the headcount in it when there is room.
func (d *Dashboard) retitle() {
	d.sidebar.SetTitle(vibe.Line("ui.dashboard.build.01", " Panels "))
	d.logTV.SetTitle(vibe.Line("ui.dashboard.build.02", " Log "))
	d.inputField.SetTitle(vibe.Line("ui.dashboard.build.13", " Command "))
	d.workerMiniTV.SetTitle(vibe.Line("ui.dashboard.build.14", " People "))
}

// refresh is the central UI update function, called every 500 ms from StartUpdates.
// It must only be called from within app.QueueUpdateDraw (i.e. on the UI goroutine).
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
// toast in the badge's own colours, written for the room the header's notice has, and
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
			d.engine.AddLog("success", themeUnlockToast(t.Name.String()))
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
		return []string{vibe.Line("ui.dashboard.heldByLines.01", "Held by: housing")}
	case game.HeldFood:
		x := textfmt.RateValue(ws.FoodNet) // "+0.45"
		feeds := vibe.Sprintf("ui.dashboard.heldByLines.02", "feeds %d more", ws.FoodFeeds)
		for _, line := range []string{
			x + vibe.Line("ui.dashboard.heldByLines.03", " food a tick ") + feeds,
			x + vibe.Line("ui.dashboard.heldByLines.04", "/tick ") + feeds,
			x + " " + feeds,
			feeds,
		} {
			if fits(line) {
				return []string{vibe.Line("ui.dashboard.heldByLines.05", "Held by: food"), line}
			}
		}
		return []string{vibe.Line("ui.dashboard.heldByLines.06", "Held by: food")}
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

// txtPeopleBoxCount is said the same way on several screens, so it has one id.
var txtPeopleBoxCount = vibe.E("ui.dashboard.peopleBoxTitle.03", " People %s/%s ")

// peopleBoxTitle is the People box's title: " People ", with the headcount
// over the housing in it when the box is compact and it fits.
func peopleBoxTitle(state game.GameState, detail bool) string {
	if !peopleBoxCompact(state, detail) {
		return vibe.Line("ui.dashboard.peopleBoxTitle.01", " People ")
	}
	ws := state.Workers
	for _, t := range []string{
		vibe.Sprintf("ui.dashboard.peopleBoxTitle.02", " People %d/%d ", ws.TotalPop, ws.MaxPop),
		fmt.Sprintf(txtPeopleBoxCount.String(), textfmt.Number(float64(ws.TotalPop)), textfmt.Number(float64(ws.MaxPop))),
	} {
		if runeLen(t) <= sidebarW-2 {
			return t
		}
	}
	return vibe.Line("ui.dashboard.peopleBoxTitle.04", " People ")
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
			return vibe.Sprintf("ui.dashboard.peopleBoxText.02", "[gray]Pop[white] [%s]%s[white]/[green]%s[-]\n", popColor, pop, housing)
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
	if lines := heldByLines(ws, w); len(lines) > 0 {
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
	// A toast written to fit (a badge's) is asked for the room the notice has.
	v.Notice = strings.TrimSpace(kit.Strip(safeTags(d.toastMgr.CurrentFor(max(w-31, 0)))))
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

// plainFact is the text a vibe line wrote, without its colour tags and the
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
		v.Prestige = plainFact(vibe.Sprintf("ui.dashboard.statusLine.01", "  [cyan]Prestige %d[-]", state.Prestige.Level))
	}
	if game.DevModeActive {
		v.Dev = plainFact(vibe.Line("ui.dashboard.statusLine.02", "  [red]DEV[-]"))
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
		v.Alerts = append(v.Alerts, headerAlert{vibe.Line("ui.dashboard.statusLine.04", "☄ Catastrophe: type catastrophe"), vibe.Line("ui.dashboard.statusLine.05", "☄ Catastrophe"), "catastrophe", xChipMil})
	} else if state.LastPassage.Pending {
		v.Alerts = append(v.Alerts, headerAlert{vibe.Line("ui.dashboard.statusLine.07", "☄ Last Passage: type catastrophe"), vibe.Line("ui.dashboard.statusLine.08", "☄ Last Passage"), "catastrophe", xChipMil})
	}
	// Harbinger badge: present until the doom it warns of resolves. Nothing
	// expires, so the badge is the idle player's reminder that there is a
	// choice waiting (it never blocks anything).
	if state.Harbinger != nil {
		v.Alerts = append(v.Alerts, headerAlert{vibe.Line("ui.dashboard.statusLine.10", "⚑ Harbinger: type harbinger"), vibe.Line("ui.dashboard.statusLine.11", "⚑ Harbinger"), "harbinger", xChipLegend})
	}
	if mBand.DeltaLabel != "" {
		v.Delta = headFact{plainFact(vibe.Sprintf("ui.dashboard.statusLine.12", " (production [%s]%s[-])", mBand.Color, mBand.DeltaLabel)), mBand.DeltaLabel}
	}
	pct := fmt.Sprintf("%.0f%%", state.Morale*100)
	v.Morale = headFact{plainFact(vibe.Sprintf("ui.dashboard.statusLine.13", "  Morale [%s]%.0f%%[-]%s", mBand.Color, state.Morale*100, "")), pct}
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
		v.Year = vibe.Sprintf("ui.dashboard.statusLine.19", "Year %s", year)
	}
	pop := groupedCount(state.Workers.TotalPop) + "/" + groupedCount(state.Workers.MaxPop)
	short := textfmt.Number(float64(state.Workers.TotalPop)) + "/" + textfmt.Number(float64(state.Workers.MaxPop))
	v.People = headFact{vibe.Sprintf("ui.dashboard.statusLine.17", "People %s", pop), pop}
	v.PeopleShort = headFact{strings.Replace(v.People.Text, pop, short, 1), short}
	return v
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
	d.ledger.update(state, d.engine)
	d.headerTV.changed()
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
				d.engine.AddLog("info", vibe.Line("ui.dashboard.showDevUnlockModal.01", "[red]Dev mode on.[-] Prefix commands with / (for example /god, /fill, /give wood 9999)."))
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
var txtOK = vibe.E("ui.dashboard.leaveToMenu.02", "OK")

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
	tv.SetBorder(true).SetTitle(vibe.Line("ui.dashboard.showWindow.01", " Recovery code "))
	tv.SetText("\n" + kit.Escape(text) + "\n\n" + theme.KeycapButton("Esc", vibe.Line("ui.dashboard.showWindow.03", "Close")))
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
	d.runLine(text)
}

// clickLine writes a command line in the prompt and gives the prompt the
// keyboard, for a click on something that stands for the line; the second
// click of a double click runs it, as Enter would.
func (d *Dashboard) clickLine(line string, clicks int) {
	d.inputField.SetText(line)
	d.app.SetFocus(d.inputField)
	if clicks >= 2 {
		d.submitInput()
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
	if result.Vibe {
		d.syncVibe()
		d.refresh()
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
		d.updateSidebar(result.OverlayName)
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
		d.engine.AddLog(result.Type, result.Message)
	}
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
