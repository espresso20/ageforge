package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// devConsole is a floating overlay the developer can toggle with backtick.
// It is only visible after the passphrase has been accepted.
type devConsole struct {
	pages  *tview.Pages
	engine *game.GameEngine
	app    *tview.Application

	panel  *tview.Flex
	output *tview.TextView
	input  *tview.InputField
	visible bool
}

const devConsolePage = "__dev_console__"

func newDevConsole(app *tview.Application, pages *tview.Pages, engine *game.GameEngine) *devConsole {
	dc := &devConsole{app: app, pages: pages, engine: engine}

	dc.output = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	dc.output.SetBorder(false)
	dc.output.SetBackgroundColor(tcell.ColorBlack)

	dc.input = tview.NewInputField().
		SetLabel("[red]/[-] ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorLime).
		SetLabelColor(tcell.ColorRed)

	dc.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		raw := strings.TrimSpace(dc.input.GetText())
		dc.input.SetText("")
		if raw == "" {
			return
		}
		if !strings.HasPrefix(raw, "/") {
			raw = "/" + raw
		}
		result := game.DevExecCommand(raw, dc.engine)
		dc.print("[yellow]> " + raw + "[-]")
		if result != "" {
			dc.print("[lime]" + result + "[-]")
		}
	})

	dc.panel = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(dc.output, 0, 1, false).
		AddItem(dc.input, 1, 0, true)
	dc.panel.SetBorder(true).
		SetBorderColor(tcell.ColorRed).
		SetTitle(" [red]DEV CONSOLE[-] — /ages /age /fill /give /techs /build /prestige /speed /god ").
		SetTitleColor(tcell.ColorRed).
		SetBackgroundColor(tcell.ColorBlack)

	dc.print("[red]DEV MODE ACTIVE[-] — type [lime]/ages[-] to list all age keys")
	return dc
}

func (dc *devConsole) print(line string) {
	dc.output.SetText(dc.output.GetText(false) + line + "\n")
	dc.output.ScrollToEnd()
}

func (dc *devConsole) toggle() {
	if dc.visible {
		dc.pages.RemovePage(devConsolePage)
		dc.visible = false
	} else {
		// Centred overlay: 60% wide, 40% tall
		overlay := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(dc.panel, 0, 2, true).
				AddItem(nil, 0, 1, false), 0, 3, true).
			AddItem(nil, 0, 1, false)
		dc.pages.AddPage(devConsolePage, overlay, true, true)
		dc.app.SetFocus(dc.input)
		dc.visible = true
	}
}
