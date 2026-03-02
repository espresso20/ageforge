package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// DevTab is a hidden dashboard tab unlocked by the developer passphrase.
// It provides a command log + input for dev commands.
// The normal game command input at the bottom still works as usual.
type DevTab struct {
	engine *game.GameEngine
	root   *tview.Flex
	output *tview.TextView
	input  *tview.InputField
}

func newDevTab(engine *game.GameEngine) *DevTab {
	dt := &DevTab{engine: engine}

	dt.output = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	dt.output.SetBorder(false)
	dt.output.SetBackgroundColor(tcell.ColorBlack)

	dt.input = tview.NewInputField().
		SetLabel("[red]dev>[-] ").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorLime).
		SetLabelColor(tcell.ColorRed)

	dt.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		raw := strings.TrimSpace(dt.input.GetText())
		dt.input.SetText("")
		if raw == "" {
			return
		}
		if !strings.HasPrefix(raw, "/") {
			raw = "/" + raw
		}
		dt.print("[yellow]" + raw + "[-]")
		result := game.DevExecCommand(raw, dt.engine)
		if result != "" {
			dt.print("[lime]→ " + result + "[-]")
		}
	})

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[red]  DEV MODE[-]  [gray]Commands: /ages /age <key> /fill /give <res> <n> /techs /build <key> /prestige <n> /speed <n> /god[-]")

	dt.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(dt.output, 0, 1, false).
		AddItem(dt.input, 1, 0, true)

	dt.print("[red]DEV MODE ACTIVE[-]  type [lime]/ages[-] to list all age keys, [lime]/god[-] to toggle free costs")
	return dt
}

func (dt *DevTab) print(line string) {
	dt.output.SetText(dt.output.GetText(false) + line + "\n")
	dt.output.ScrollToEnd()
}

// Primitive returns the root view for registration as a tab page.
func (dt *DevTab) Primitive() tview.Primitive {
	return dt.root
}

// FocusInput sets focus to the dev command input.
func (dt *DevTab) FocusInput() *tview.InputField {
	return dt.input
}
