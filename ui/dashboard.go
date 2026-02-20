package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// Dashboard is the main gameplay screen with tabbed layout
type Dashboard struct {
	app    *tview.Application
	engine *game.GameEngine
	pages  *tview.Pages
	root   *tview.Flex

	// Tab system
	tabBar    *tview.TextView
	tabPages  *tview.Pages
	activeTab int
	tabNames  []string

	// Tabs
	economyTab  *EconomyTab
	researchTab *ResearchTab
	militaryTab *MilitaryTab
	statsTab    *StatsTab
	wikiTab     *WikiTab
	mapTab      *MapTab
	logsTab     *LogsTab

	// Shared UI
	logTV      *tview.TextView
	miniMap    *MiniMap
	statusTV   *tview.TextView
	ageTV      *tview.TextView
	inputField *tview.InputField
	lastAge     string
	toastMgr    *ToastManager
	toastTV     *tview.TextView
	contentArea *tview.Flex
	bottomArea  *tview.Flex

	stopCh chan struct{}
}

// NewDashboard creates the gameplay dashboard
func NewDashboard(app *tview.Application, engine *game.GameEngine, pages *tview.Pages) *Dashboard {
	d := &Dashboard{
		app:      app,
		engine:   engine,
		pages:    pages,
		stopCh:   make(chan struct{}),
		tabNames: []string{"Economy", "Research", "Military", "Stats", "Wiki", "Map", "Logs"},
	}
	d.build()
	return d
}

func (d *Dashboard) build() {
	// Create tabs
	d.economyTab = NewEconomyTab()
	d.researchTab = NewResearchTab()
	d.militaryTab = NewMilitaryTab()
	d.statsTab = NewStatsTab()
	d.wikiTab = NewWikiTab()
	d.mapTab = NewMapTab()
	d.logsTab = NewLogsTab()

	// Tab bar
	d.tabBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	d.updateTabBar()

	// Tab pages
	d.tabPages = tview.NewPages()
	d.tabPages.AddPage("Economy", d.economyTab.Root(), true, true)
	d.tabPages.AddPage("Research", d.researchTab.Root(), true, false)
	d.tabPages.AddPage("Military", d.militaryTab.Root(), true, false)
	d.tabPages.AddPage("Stats", d.statsTab.Root(), true, false)
	d.tabPages.AddPage("Wiki", d.wikiTab.Root(), true, false)
	d.tabPages.AddPage("Map", d.mapTab.Root(), true, false)
	d.tabPages.AddPage("Logs", d.logsTab.Root(), true, false)

	// Log panel
	d.logTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true).SetTitle(" Log ").SetTitleColor(ColorDim)

	// Mini-map panel (replaces Quick Reference)
	d.miniMap = NewMiniMap()

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
	d.engine.Bus.Subscribe(game.EventAgeAdvanced, func(e game.EventData) {
		if newAge, ok := e.Payload["new_age"].(string); ok {
			_ = newAge
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
			if text == "" {
				return
			}
			if strings.ToLower(strings.TrimSpace(text)) == "quit" {
				d.engine.SaveGame("autosave")
				d.app.Stop()
				return
			}
			result := HandleCommand(text, d.engine)
			if result.Message != "" {
				d.engine.AddLog(result.Type, result.Message)
			}
		}
	})

	// Bottom area: log + mini-map side by side
	d.bottomArea = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(d.logTV, 0, 1, false).
		AddItem(d.miniMap.Primitive(), 0, 1, false)

	// Main content area: tab content + bottom
	d.contentArea = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.tabPages, 0, 2, false).
		AddItem(d.bottomArea, 0, 1, false)

	// Root layout
	d.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.statusTV, 1, 0, false).
		AddItem(d.toastTV, 1, 0, false).
		AddItem(d.ageTV, 2, 0, false).
		AddItem(d.tabBar, 1, 0, false).
		AddItem(d.contentArea, 0, 1, false).
		AddItem(d.inputField, 1, 0, true)

	// Global key handling
	d.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			d.engine.SaveGame("autosave")
			d.engine.Stop()
			d.pages.SwitchToPage("splash")
			return nil
		case tcell.KeyF1:
			d.switchTab(0)
			return nil
		case tcell.KeyF2:
			d.switchTab(1)
			return nil
		case tcell.KeyF3:
			d.switchTab(2)
			return nil
		case tcell.KeyF4:
			d.switchTab(3)
			return nil
		case tcell.KeyF5:
			d.switchTab(4)
			return nil
		case tcell.KeyF6:
			d.switchTab(5)
			return nil
		case tcell.KeyF9:
			d.switchTab(6)
			return nil
		}

		// When logs tab is active, intercept navigation keys
		if d.activeTab == 6 {
			switch event.Key() {
			case tcell.KeyPgUp:
				d.logsTab.ScrollUp()
				return nil
			case tcell.KeyPgDn:
				d.logsTab.ScrollDown()
				return nil
			}
			if event.Rune() == 'v' {
				d.logsTab.ToggleVerbose()
				return nil
			}
		}

		// When wiki tab is active, intercept navigation keys
		if d.activeTab == 4 {
			switch event.Key() {
			case tcell.KeyUp:
				d.wikiTab.PrevPage()
				return nil
			case tcell.KeyDown:
				d.wikiTab.NextPage()
				return nil
			case tcell.KeyPgUp:
				d.wikiTab.ScrollUp()
				return nil
			case tcell.KeyPgDn:
				d.wikiTab.ScrollDown()
				return nil
			}
			// Number keys for quick nav
			if event.Rune() >= '1' && event.Rune() <= '9' {
				idx := int(event.Rune() - '1')
				d.wikiTab.GoToPage(idx)
				return nil
			}
		}

		// Always focus input field for typing (except wiki nav)
		if !d.inputField.HasFocus() {
			d.app.SetFocus(d.inputField)
		}
		return event
	})
}

func (d *Dashboard) switchTab(index int) {
	d.activeTab = index
	d.tabPages.SwitchToPage(d.tabNames[index])
	d.updateTabBar()

	// Map tab (index 5) is full-screen — hide bottom area
	if index == 5 {
		d.contentArea.RemoveItem(d.bottomArea)
	} else {
		// Re-add bottom area if not already present
		if d.contentArea.GetItemCount() < 2 {
			d.contentArea.AddItem(d.bottomArea, 0, 1, false)
		}
	}
}

func (d *Dashboard) updateTabBar() {
	var parts []string
	tabKeys := map[int]string{0: "F1", 1: "F2", 2: "F3", 3: "F4", 4: "F5", 5: "F6", 6: "F9"}
	for i, name := range d.tabNames {
		key := tabKeys[i]
		if i == d.activeTab {
			parts = append(parts, fmt.Sprintf(" [black:gold] %s %s [-:-] ", key, name))
		} else {
			parts = append(parts, fmt.Sprintf(" [gray]%s %s[-] ", key, name))
		}
	}
	d.tabBar.SetText(strings.Join(parts, "  "))
}

// Root returns the root primitive for page registration
func (d *Dashboard) Root() tview.Primitive {
	return d.root
}

// StartUpdates begins the UI refresh loop
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

// StopUpdates stops the UI refresh loop
func (d *Dashboard) StopUpdates() {
	select {
	case d.stopCh <- struct{}{}:
	default:
	}
}

func (d *Dashboard) refresh() {
	state := d.engine.GetState()

	if d.lastAge != state.Age {
		ApplyAgePalette(state.Age)
		d.lastAge = state.Age
	}

	d.refreshStatus(state)
	d.refreshAgeProgress(state)
	d.refreshLog(state)
	d.toastTV.SetText(d.toastMgr.GetCurrent())
	d.miniMap.UpdateState(state)

	// Only refresh the active tab
	switch d.activeTab {
	case 0:
		d.economyTab.Refresh(state)
	case 1:
		d.researchTab.Refresh(state)
	case 2:
		d.militaryTab.Refresh(state)
	case 3:
		d.statsTab.Refresh(state)
	case 4:
		d.wikiTab.Refresh(state)
	case 5:
		d.mapTab.Refresh(state)
	case 6:
		d.logsTab.Refresh(state)
	}
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
		speedStr = fmt.Sprintf("  [yellow]%.0fx[-]", state.SpeedMultiplier)
	}
	d.statusTV.SetText(fmt.Sprintf(
		"[gold]%s[-]%s  Tick: %d%s%s  |  Pop: %d/%d  |  [gray]F1-F6,F9=Tabs  ESC=Menu[-]",
		state.AgeName, prestigeStr, state.Tick, nextAgeStr, speedStr,
		state.Villagers.TotalPop, state.Villagers.MaxPop,
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

	d.ageTV.SetText(sb.String())
}

func (d *Dashboard) refreshLog(state game.GameState) {
	var sb strings.Builder
	start := 0
	if len(state.Log) > 20 {
		start = len(state.Log) - 20
	}
	for _, entry := range state.Log[start:] {
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
		}
		fmt.Fprintf(&sb, "[gray]T%d[-] [%s]%s[-]\n", entry.Tick, color, entry.Message)
	}
	d.logTV.SetText(sb.String())
	d.logTV.ScrollToEnd()
}

