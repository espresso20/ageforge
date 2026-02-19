package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

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
	logsTab     *LogsTab

	// Shared UI
	logTV      *tview.TextView
	helpTV     *tview.TextView
	statusTV   *tview.TextView
	ageTV      *tview.TextView
	inputField *tview.InputField

	stopCh chan struct{}
}

// NewDashboard creates the gameplay dashboard
func NewDashboard(app *tview.Application, engine *game.GameEngine, pages *tview.Pages) *Dashboard {
	d := &Dashboard{
		app:      app,
		engine:   engine,
		pages:    pages,
		stopCh:   make(chan struct{}),
		tabNames: []string{"Economy", "Research", "Military", "Stats", "Wiki", "Logs"},
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
	d.tabPages.AddPage("Logs", d.logsTab.Root(), true, false)

	// Log panel
	d.logTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(100)
	d.logTV.SetBorder(true).SetTitle(" Log ").SetTitleColor(ColorDim)

	// Help/tips panel
	d.helpTV = tview.NewTextView().
		SetDynamicColors(true)
	d.helpTV.SetBorder(true).SetTitle(" Quick Reference ").SetTitleColor(ColorTitle)
	d.helpTV.SetText(helpText())

	// Status bar
	d.statusTV = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// Age progress tracker
	d.ageTV = tview.NewTextView().
		SetDynamicColors(true)

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

	// Bottom area: log + help side by side
	bottomArea := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(d.logTV, 0, 1, false).
		AddItem(d.helpTV, 0, 1, false)

	// Main content area: tab content + bottom
	contentArea := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.tabPages, 0, 2, false).
		AddItem(bottomArea, 0, 1, false)

	// Root layout
	d.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.statusTV, 1, 0, false).
		AddItem(d.ageTV, 2, 0, false).
		AddItem(d.tabBar, 1, 0, false).
		AddItem(contentArea, 0, 1, false).
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
		case tcell.KeyF9:
			d.switchTab(5)
			return nil
		}

		// When logs tab is active, intercept navigation keys
		if d.activeTab == 5 {
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
}

func (d *Dashboard) updateTabBar() {
	var parts []string
	tabKeys := map[int]string{0: "F1", 1: "F2", 2: "F3", 3: "F4", 4: "F5", 5: "F9"}
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
	d.refreshStatus(state)
	d.refreshAgeProgress(state)
	d.refreshLog(state)

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
	d.statusTV.SetText(fmt.Sprintf(
		"[gold]%s[-]%s  Tick: %d%s  |  Pop: %d/%d  |  [gray]F1-F5,F9=Tabs  ESC=Menu[-]",
		state.AgeName, prestigeStr, state.Tick, nextAgeStr,
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
		fmt.Fprintf(&sb, "[%s]%s:%.0f/%.0f[-]  ", color, key, current, req)
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
			fmt.Fprintf(&sb, "[%s]%s:%d/%d[-]  ", color, key, current, req)
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

func helpText() string {
	return ` [gold]Commands[-]
 [cyan]gather[-] <resource> [n]    Gather by hand
 [cyan]build[-] <building>         Build a structure
 [cyan]recruit[-] <type> [n]       Recruit villagers
 [cyan]assign[-] <type> <res> [n]  Put to work
 [cyan]unassign[-] <type> <res> [n] Remove from work
 [cyan]research[-] <tech>          Research tech
 [cyan]expedition[-] <key>         Launch expedition
 [cyan]prestige[-] [shop|buy|confirm] Prestige system
 [cyan]status[-]                   Detailed overview
 [cyan]save[-] / [cyan]load[-] [name]       Save or load game

 [gold]Shortcuts[-]
 g=gather  b=build  r=recruit
 a=assign  u=unassign  s=status
 res=research  exp=expedition

 [gold]Navigation[-]
 F1-F5,F9       Switch tabs
 Tab            Autocomplete
 ESC            Save & menu

 [gold]Villager Types[-]
 [mediumpurple]Worker[-]    Gathers resources
 [mediumpurple]Scholar[-]   Generates knowledge
 [mediumpurple]Soldier[-]   Military (Bronze+)
 [mediumpurple]Merchant[-]  Earns gold (Medieval)`
}
