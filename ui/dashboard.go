package ui

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
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
type Dashboard struct {
	app    *tview.Application
	engine *game.GameEngine
	pages  *tview.Pages
	root   *tview.Flex

	// Tabs (only economy is permanent background)
	economyTab *EconomyTab

	// Sidebar
	sidebar       *tview.TextView
	workerMiniTV  *tview.TextView

	// Shared UI
	logTV    *tview.TextView
	statusTV *tview.TextView
	ageTV      *tview.TextView
	inputField *tview.InputField
	lastAge             string
	pendingAgeSplash    string // set by bus handler, consumed by refresh()
	pendingEpochChanged bool   // whether the pending age advance also crossed an epoch boundary
	toastMgr            *ToastManager
	toastTV     *tview.TextView
	contentArea *tview.Flex

	// Shame badge — set once on first load when CheaterBadge is true
	cheaterTV       *tview.TextView
	activeShameBadge string

	// catModalShown tracks the pending catastrophe key we have already shown a modal for,
	// so that choosing Defer does not immediately re-show the modal on the next refresh tick.
	// Reset to "" when PendingCatastrophe clears (player chose Endure or Succumb).
	catModalShown string

	// Command history is session-only (never persisted to disk). The slice is
	// append-only and capped at 50 entries. histIdx == -1 means not in history
	// navigation mode. histDraft stores whatever was in the input field before
	// the user started pressing Up.
	cmdHistory []string // append-only slice, capped at 50
	histIdx    int      // -1 = not navigating; 0 = most recent; len-1 = oldest
	histDraft  string   // draft text saved when user starts navigating history

	overlayMgr *OverlayManager
	lastState  *game.GameState

	stopCh chan struct{}
}

// NewDashboard creates the gameplay dashboard
func NewDashboard(app *tview.Application, engine *game.GameEngine, pages *tview.Pages) *Dashboard {
	d := &Dashboard{
		app:    app,
		engine: engine,
		pages:  pages,
		stopCh: make(chan struct{}),
		histIdx: -1,
	}
	d.build()
	d.overlayMgr = NewOverlayManager(d.pages, d.app, func() {
		d.updateSidebar("")
		d.app.SetFocus(d.inputField)
	})
	d.overlayMgr.Register("milestones", "Milestones", milestonesProvider)
	d.overlayMgr.Register("techs", "Research", researchProvider)
	d.overlayMgr.Register("army", "Military", militaryProvider)
	d.overlayMgr.Register("trade", "Trade", tradeProvider)
	d.overlayMgr.Register("stats", "Statistics", statsProvider)
	d.overlayMgr.Register("wonders", "Wonders", wondersProvider)
	d.overlayMgr.Register("workers", "Workers", workersProvider)
	d.overlayMgr.Register("logs", "Logs", logsProvider)
	d.overlayMgr.Register("epoch", "Epoch", epochProvider)
	d.overlayMgr.Register("history", "Civilization History", historyProvider)

	mv4 := NewMapV4()
	d.overlayMgr.RegisterWidget("map", "City Map", mv4.Build, mv4.Refresh, true)

	return d
}

func (d *Dashboard) build() {
	// Create permanent economy tab
	d.economyTab = NewEconomyTab()

	// Sidebar — command panel hints
	d.sidebar = tview.NewTextView().
		SetDynamicColors(true).
		SetText(buildSidebarText(""))
	d.sidebar.SetBorder(true).SetTitle(" Panels ").SetTitleColor(tcell.ColorGold)

	// Log panel
	d.logTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true).SetTitle(" Log ").SetTitleColor(ColorDim)

	// Shame badge bar (1 fixed line; text only shown when CheaterBadge is true)
	d.cheaterTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Status bar
	d.statusTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Age progress tracker
	d.ageTV = tview.NewTextView().
		SetDynamicColors(true)

	// Toast notification
	d.toastMgr = NewToastManager()
	d.toastTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

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
		d.toastMgr.Show("AGE ADVANCED!", "gold", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventResearchDone, func(e game.EventData) {
		tech, _ := e.Payload["tech"].(string)
		d.toastMgr.Show(fmt.Sprintf("Research Complete: %s", tech), "cyan", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventBuildingBuilt, func(e game.EventData) {
		building, _ := e.Payload["building"].(string)
		// Only toast for wonders — look up from config, not engine state (avoids deadlock)
		if def, ok := config.BuildingByKey()[building]; ok && def.Category == "wonder" {
			d.toastMgr.Show(fmt.Sprintf("Wonder Built: %s", def.Name), "green", 4*time.Second)
		}
	})
	d.engine.Bus.Subscribe(game.EventMilestoneCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		rewardText, _ := e.Payload["reward_text"].(string)
		msg := fmt.Sprintf("Milestone: %s!", name)
		if rewardText != "" {
			msg += " " + rewardText
		}
		d.toastMgr.Show(msg, "gold", 4*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventChainCompleted, func(e game.EventData) {
		name, _ := e.Payload["name"].(string)
		title, _ := e.Payload["title"].(string)
		d.toastMgr.Show(fmt.Sprintf("Chain Complete: %s! Title: %s — Speed Boost!", name, title), "cyan", 5*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventEpochAdvanced, func(e game.EventData) {
		epochName, _ := e.Payload["epoch_name"].(string)
		epochIcon, _ := e.Payload["epoch_icon"].(string)
		d.toastMgr.Show(fmt.Sprintf("✦ The %s %s Dawns!", epochIcon, epochName), "gold", 6*time.Second)
	})
	d.engine.Bus.Subscribe(game.EventEpochEventFired, func(e game.EventData) {
		eventName, _ := e.Payload["event_name"].(string)
		eventType, _ := e.Payload["event_type"].(string)
		color := "cyan"
		if eventType == "bad_challenging" {
			color = "red"
		} else if eventType == "catastrophe" {
			color = "red"
		} else if eventType == "good_legendary" {
			color = "gold"
		}
		d.toastMgr.Show(fmt.Sprintf("Epoch Event: %s", eventName), color, 6*time.Second)
	})

	// Command input
	d.inputField = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(ColorAccent)

	// Wire up autocomplete
	d.inputField.SetAutocompleteFunc(NewAutoCompleter(d.engine))
	d.inputField.SetAutocompletedFunc(func(text string, index, source int) bool {
		d.inputField.SetText(text + " ")
		return true
	})

	d.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := d.inputField.GetText()
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
				d.engine.SaveGame("autosave")
				d.app.Stop()
				return
			}
			// Route /commands to dev exec when dev mode is active
			if game.DevModeActive && strings.HasPrefix(cmd, "/") {
				result := game.DevExecCommand(cmd, d.engine)
				if result != "" {
					d.engine.AddLog("info", "[lime]dev → "+result+"[-]")
				}
				return
			}
			result := HandleCommand(text, d.engine)
			if result.OverlayName != "" {
				state := d.engine.GetState()
				d.overlayMgr.Show(result.OverlayName, state)
				d.updateSidebar(result.OverlayName)
			}
			if result.Message != "" && result.Type != "success" {
				d.engine.AddLog(result.Type, result.Message)
			}
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
			// Keep draft in sync while user types normally
			// (draft is re-read from field on next Up press, so nothing extra needed)
			return event
		}
	})

	// Inject log panel into the economy tab's left column (below Under Construction)
	d.economyTab.AddToLeftColumn(d.logTV, 0, 1)

	// Mini worker summary box — sits below the sidebar in the right column
	d.workerMiniTV = tview.NewTextView().SetDynamicColors(true)
	d.workerMiniTV.SetBorder(true).SetTitle(" Workers ").SetTitleColor(ColorVillager)

	rightCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.sidebar, 0, 1, false).
		AddItem(d.workerMiniTV, 8, 0, false)

	// Main horizontal: economy (permanent, full height) + right column (sidebar + worker mini)
	mainHoriz := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(d.economyTab.Root(), 0, 1, false).
		AddItem(rightCol, 22, 0, false)

	// Content area is just mainHoriz — no bottom strip
	d.contentArea = mainHoriz

	// Root layout (no tab bar)
	d.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.cheaterTV, 1, 0, false).
		AddItem(d.statusTV, 1, 0, false).
		AddItem(d.toastTV, 1, 0, false).
		AddItem(d.ageTV, 2, 0, false).
		AddItem(d.contentArea, 0, 1, false).
		AddItem(d.inputField, 1, 0, true)

	// Global key handling
	d.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Ctrl+K — open passphrase modal (dev unlock)
		if event.Key() == tcell.KeyCtrlK {
			d.showDevUnlockModal()
			return nil
		}
		switch event.Key() {
		case tcell.KeyEsc:
			if d.overlayMgr != nil && d.overlayMgr.HasActive() {
				d.overlayMgr.Hide()
				return nil
			}
			d.engine.SaveGame("autosave")
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
	if d.sidebar != nil {
		d.sidebar.SetText(buildSidebarText(activeOverlay))
	}
}

func buildSidebarText(active string) string {
	commands := []string{"milestones", "research", "army", "trade", "stats", "wonders", "workers", "logs", "epoch", "history", "map"}
	var sb strings.Builder
	sb.WriteString("\n")
	for _, cmd := range commands {
		if cmd == active {
			sb.WriteString(fmt.Sprintf(" [black:gold] %-10s [-:-]\n", cmd))
		} else {
			sb.WriteString(fmt.Sprintf(" [white]%-10s[-]\n", cmd))
		}
	}
	return sb.String()
}

// Root returns the root primitive for page registration
func (d *Dashboard) Root() tview.Primitive {
	return d.root
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
		for {
			select {
			case <-ticker.C:
				d.app.QueueUpdateDraw(func() {
					d.refresh()
				})
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

// refresh is the central UI update function, called every 500 ms from StartUpdates.
// It must only be called from within app.QueueUpdateDraw (i.e. on the tview goroutine).
// Order matters: catastrophe modal and age splash are checked first because they may
// swap the active page, then all content widgets are updated.
func (d *Dashboard) refresh() {
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

	state := d.engine.GetState()
	d.lastState = &state

	// Phase 9: catastrophe modal — show once per new pending catastrophe; Defer hides it
	if state.PendingCatastrophe == "" {
		d.catModalShown = "" // reset so next catastrophe will show fresh
	} else if d.catModalShown == "" {
		d.catModalShown = state.PendingCatastrophe
		d.showCatastropheModal(state.PendingCatastrophe)
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
	d.toastTV.SetText(d.toastMgr.GetCurrent())

	// Economy tab is always visible as the permanent background
	d.economyTab.Refresh(state)

	// Update overlay content and sidebar highlight
	d.overlayMgr.Refresh(state)
	d.updateSidebar(d.overlayMgr.ActiveName())
	d.refreshWorkerMini(state)
}

func (d *Dashboard) refreshWorkerMini(state game.GameState) {
	w := state.Workers
	food, hasFoodRS := state.Resources["food"]
	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]%d[white]/[green]%d[-]  [gray]Idle:[white] %d[-]\n", w.TotalPop, w.MaxPop, w.TotalIdle)
	fmt.Fprintf(&sb, "[gray]House left:[white] %d[-]\n", w.MaxPop-w.TotalPop)
	fmt.Fprintf(&sb, "[gray]Drain:[red] %.2f/t[-]\n", w.FoodDrain)
	if hasFoodRS {
		netColor := "green"
		prefix := "+"
		if food.Rate < 0 {
			netColor = "red"
			prefix = ""
		}
		fmt.Fprintf(&sb, "[gray]Net:[%s] %s%.2f/t[-]\n", netColor, prefix, food.Rate)
	}
	d.workerMiniTV.SetText(sb.String())
}

func (d *Dashboard) refreshStatus(state game.GameState) {
	nextAgeStr := ""
	if state.NextAge != "" {
		nextAgeStr = fmt.Sprintf("  [gray]Next: %s[-]", state.NextAge)
	}
	prestigeStr := ""
	if state.Prestige.Level > 0 {
		prestigeStr = fmt.Sprintf("  [cyan]P%d[-]", state.Prestige.Level)
	}
	speedStr := ""
	if state.SpeedMultiplier > 1 {
		speedStr = fmt.Sprintf("  [yellow]%.1fx[-]", state.SpeedMultiplier)
	}
	titleStr := ""
	if state.Milestones.CurrentTitle != "" {
		titleStr = fmt.Sprintf("  [yellow]\"%s\"[-]", state.Milestones.CurrentTitle)
	}
	devStr := ""
	if game.DevModeActive {
		devStr = "  [red]DEV[-]"
	}
	// Epoch badge
	epochStr := ""
	if state.EpochKey != "" {
		survivedMark := ""
		if state.EpochSurvived {
			survivedMark = " ·Survived"
		}
		epochStr = fmt.Sprintf("  [%s]%s %s%s[-]", state.EpochColor, state.EpochIcon, state.EpochName, survivedMark)
	}
	moraleColor := "green"
	if state.Morale < 0.50 {
		moraleColor = "red"
	} else if state.Morale < 0.80 {
		moraleColor = "yellow"
	}
	moraleStr := fmt.Sprintf("  Morale: [%s]%.0f%%[-]", moraleColor, state.Morale*100)
	d.statusTV.SetText(fmt.Sprintf(
		"[gold]%s[-]%s%s%s  Tick: %d%s%s%s  |  Pop: %d/%d%s  |  [gray]type panel name to open  ESC=close/menu[-]",
		state.AgeName, prestigeStr, titleStr, epochStr, state.Tick, nextAgeStr, speedStr, devStr,
		state.Workers.TotalPop, state.Workers.MaxPop, moraleStr,
	))
}

func (d *Dashboard) refreshAgeProgress(state game.GameState) {
	if state.NextAge == "" {
		d.ageTV.SetText(" [gold]You have reached the final age![-]")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, " [gold]Next Age: %s[-]  ", state.NextAgeName)

	// Resource requirements
	resKeys := make([]string, 0, len(state.NextAgeResReqs))
	for k := range state.NextAgeResReqs {
		resKeys = append(resKeys, k)
	}
	sort.Strings(resKeys)
	for _, key := range resKeys {
		req := state.NextAgeResReqs[key]
		current := 0.0
		if rs, ok := state.Resources[key]; ok {
			current = rs.Amount
		}
		color := "red"
		if current >= req {
			color = "green"
		}
		bar := ProgressBar(current, req, 8)
		fmt.Fprintf(&sb, "[%s]%s:%.0f/%.0f %s[-]  ", color, key, current, req, bar)
	}

	// Building requirements
	bldKeys := make([]string, 0, len(state.NextAgeBldReqs))
	for k := range state.NextAgeBldReqs {
		bldKeys = append(bldKeys, k)
	}
	sort.Strings(bldKeys)
	if len(bldKeys) > 0 {
		sb.WriteString(" ")
		for _, key := range bldKeys {
			req := state.NextAgeBldReqs[key]
			current := 0
			if bs, ok := state.Buildings[key]; ok {
				current = bs.Count
			}
			color := "red"
			if current >= req {
				color = "green"
			}
			bar := ProgressBar(float64(current), float64(req), 8)
			fmt.Fprintf(&sb, "[%s]%s:%d/%d %s[-]  ", color, key, current, req, bar)
		}
	}

	// Wonder gate: show if current age wonder must still be completed
	if state.CurrentAgeWonderKey != "" {
		fmt.Fprintf(&sb, "[red]✗ Wonder required: %s[-]  ", state.CurrentAgeWonderName)
	}

	d.ageTV.SetText(sb.String())
}

func (d *Dashboard) refreshLog(state game.GameState) {
	var sb strings.Builder
	// Filter to only user-facing entries (skip debug)
	var visible []game.LogEntry
	for _, entry := range state.Log {
		if entry.Type != "debug" {
			visible = append(visible, entry)
		}
	}
	start := 0
	if len(visible) > 20 {
		start = len(visible) - 20
	}
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
		fmt.Fprintf(&sb, "[gray]T%d[-] [%s]%s[-]\n", entry.Tick, color, entry.Message)
	}
	d.logTV.SetText(sb.String())
	d.logTV.ScrollToEnd()
}


// showDevUnlockModal opens an unlabelled passphrase input modal.
// No hint text is shown — the existence of this modal is not advertised.
func (d *Dashboard) showDevUnlockModal() {
	if game.DevModeActive {
		return // already active, nothing to do
	}

	const devUnlockPage = "__dev_unlock__"
	field := tview.NewInputField().
		SetLabel("").
		SetFieldWidth(40).
		SetMaskCharacter('·').
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)

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
				d.engine.AddLog("info", "[red]DEV MODE ACTIVE[-] — prefix commands with / (e.g. /god, /fill, /give wood 9999)")
				d.app.SetFocus(d.inputField)
			} else {
				d.app.SetFocus(d.inputField)
			}
		}
	})

	box := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(field, 1, 0, true)
	box.SetBorder(true).
		SetBorderColor(tcell.ColorDarkGray).
		SetBackgroundColor(tcell.ColorBlack)

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

