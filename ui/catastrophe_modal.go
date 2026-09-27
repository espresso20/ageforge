package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// catastrophePage is the tview page name for the catastrophe choice overlay.
const catastrophePage = "catastrophe"

// catastropheModalWidth is the outer width of the catastrophe box, border
// included. Every body line below fits in width-2 cells without wrapping; the
// flavor line is word-wrapped explicitly so the box height is exact.
const catastropheModalWidth = 76

// catastropheModalLayout is the content of the catastrophe box, split out so
// tests can check the sizing without a screen.
type catastropheModalLayout struct {
	title   string
	header  string // pre-wrapped
	endure  string
	succumb string
	hint    string
	height  int // outer height, border included
}

// buildCatastropheModalLayout assembles the text for epochKey's catastrophe.
// alreadyLegacy says whether this epoch's legacy is already held (a repeat
// Succumb in the same epoch grants no new legacy or research bonus), and
// researchNow is the current Succumb research bonus (e.g. 0.25).
func buildCatastropheModalLayout(epochKey string, alreadyLegacy bool, researchNow float64) catastropheModalLayout {
	catName, catFlavor := config.CatastropheInfo(epochKey)
	ep := config.EpochByKey()[epochKey]
	inner := catastropheModalWidth - 2

	headerLines := []string{fmt.Sprintf("[red]☄ %s[-]", tview.Escape(catName))}
	for _, l := range tview.WordWrap(catFlavor, inner-2) {
		headerLines = append(headerLines, "[gray]"+tview.Escape(strings.TrimSpace(l))+"[-]")
	}

	endure := strings.Join([]string{
		"[white]── ENDURE — weather the catastrophe ──[-]",
		"  [red]• 20% of buildings destroyed (wonders are spared)[-]",
		"  [red]• All resources reduced to 15%[-]",
		"  [red]• 25% of workers lost; workers of destroyed buildings go idle[-]",
		"  [red]• Production -10% for 216 ticks, morale -10[-]",
		"  [green]✓ Age, research, wonders and prestige preserved[-]",
		"  [green]✓ Survived marker on the epoch badge[-]",
	}, "\n")

	per := game.SuccumbResearchBonusPerEpoch
	researchLine := fmt.Sprintf("  [green]✓ Ancient Knowledge: research speed +%.0f%% (total +%.0f%%, permanent)[-]", per*100, (researchNow+per)*100)
	legacyLine := fmt.Sprintf("  [gold]✓ %s legacy: %s (permanent)[-]", ep.Name, legacyBonusText(epochKey))
	if alreadyLegacy {
		researchLine = fmt.Sprintf("  [gray]• Research bonus already earned here (stays +%.0f%%)[-]", researchNow*100)
		legacyLine = fmt.Sprintf("  [gray]• %s legacy already held; no new legacy bonus[-]", ep.Name)
	}
	succumb := strings.Join([]string{
		"[white]── SUCCUMB — let civilization fall ──[-]",
		"  [red]• Full reset to the Primitive Age: buildings, resources, research[-]",
		"  [red]• No prestige points earned (level and upgrades are kept)[-]",
		fmt.Sprintf("  [green]✓ Up to %d buildings become ruins (50%% output, max %d ruins)[-]", game.SuccumbRuinCount, game.MaxRuins),
		researchLine,
		legacyLine,
	}, "\n")

	hint := "[gray]Esc: decide later · advancing waits · type 'catastrophe' to reopen[-]"

	l := catastropheModalLayout{
		title:   fmt.Sprintf(" ☄ %s Catastrophe ", ep.Name),
		header:  strings.Join(headerLines, "\n"),
		endure:  endure,
		succumb: succumb,
		hint:    hint,
	}
	// Rows: header, gap, endure, gap, succumb, gap, buttons, hint, + 2 border.
	l.height = lineCount(l.header) + 1 + lineCount(l.endure) + 1 + lineCount(l.succumb) + 1 + 1 + 1 + 2
	return l
}

func lineCount(s string) int { return strings.Count(s, "\n") + 1 }

// surfaceBox is an opaque filler painted with the modal surface color. tview
// only paints cells a primitive draws, so every row and column inside the box
// must belong to a clearing primitive or the dashboard shows through.
func surfaceBox() *tview.Box {
	return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleSurface))
}

func surfaceText(text string, align int) *tview.TextView {
	tv := tview.NewTextView().SetDynamicColors(true).SetTextAlign(align).SetWrap(false).SetText(text)
	tv.SetBackgroundColor(theme.Color(theme.RoleSurface))
	return tv
}

// floatingModal centers inner at width×height over whatever page is beneath.
// The spacers are nil flex items, which draw nothing, so the dashboard stays
// visible around the box (unlike centeredModal, whose Box spacers clear it).
func floatingModal(inner tview.Primitive, width, height int) *tview.Flex {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(inner, height, 0, true).
			AddItem(nil, 0, 1, false),
			width, 0, true).
		AddItem(nil, 0, 1, false)
}

// showCatastropheModal displays the Endure / Succumb choice for the pending
// catastrophe as a box floating over the dashboard. Esc closes it without
// choosing (the catastrophe stays pending and keeps blocking advancement; the
// `catastrophe` command reopens it). Must be called from the UI goroutine.
func (d *Dashboard) showCatastropheModal(epochKey string) {
	if d.pages.HasPage(catastrophePage) {
		d.pages.ShowPage(catastrophePage)
		d.pages.SendToFront(catastrophePage)
		if d.catFocus != nil {
			d.app.SetFocus(d.catFocus)
		}
		return
	}
	st := d.engine.GetState()
	l := buildCatastropheModalLayout(epochKey, st.LegacyBonuses[epochKey], st.SuccumbResearchBonus)

	btnEndure := tview.NewButton(tview.Escape("[E] ENDURE")).
		SetSelectedFunc(func() {
			if err := d.engine.Endure(); err != nil {
				d.engine.AddLog("error", "Endure failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
	styleFilledButton(btnEndure, theme.RoleNegative)

	btnSuccumb := tview.NewButton(tview.Escape("[S] SUCCUMB")).
		SetSelectedFunc(func() {
			if err := d.engine.Succumb(); err != nil {
				d.engine.AddLog("error", "Succumb failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
	styleFilledButton(btnSuccumb, theme.RoleNegative)

	btnRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(surfaceBox(), 0, 1, false).
		AddItem(btnEndure, 14, 0, true).
		AddItem(surfaceBox(), 4, 0, false).
		AddItem(btnSuccumb, 14, 0, false).
		AddItem(surfaceBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(surfaceText(l.header, tview.AlignCenter), lineCount(l.header), 0, false).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(surfaceText(l.endure, tview.AlignLeft), lineCount(l.endure), 0, false).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(surfaceText(l.succumb, tview.AlignLeft), lineCount(l.succumb), 0, false).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(btnRow, 1, 0, true).
		AddItem(surfaceText(l.hint, tview.AlignCenter), 1, 0, false)
	inner.SetBorder(true).
		SetTitle(l.title).
		SetTitleColor(theme.Color(theme.RoleNegative)).
		SetBorderColor(theme.Color(theme.RoleNegative)).
		SetBackgroundColor(theme.Color(theme.RoleSurface))

	modal := floatingModal(inner, catastropheModalWidth, l.height)

	focusOrder := []*tview.Button{btnEndure, btnSuccumb}
	focusIdx := 0
	setFocus := func(i int) {
		focusIdx = i
		d.catFocus = focusOrder[i]
		d.app.SetFocus(focusOrder[i])
	}
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			d.closeCatastropheModal()
			return nil
		case tcell.KeyTab, tcell.KeyRight:
			setFocus((focusIdx + 1) % len(focusOrder))
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			setFocus((focusIdx - 1 + len(focusOrder)) % len(focusOrder))
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e', 'E':
				setFocus(0)
				btnEndure.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			case 's', 'S':
				setFocus(1)
				btnSuccumb.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			}
		}
		return event
	})

	d.pages.AddPage(catastrophePage, modal, true, true)
	setFocus(0)
}

// closeCatastropheModal removes the catastrophe overlay page and hands focus
// back to the command input. catModalShown stays set, so refresh() does not
// re-pop the modal for the same pending catastrophe; `catastrophe` reopens it.
func (d *Dashboard) closeCatastropheModal() {
	d.pages.RemovePage(catastrophePage)
	d.catFocus = nil
	if d.overlayMgr == nil || d.overlayMgr.ActiveName() == "" {
		d.app.SetFocus(d.inputField)
	}
}

// reopenCatastropheModal shows the modal for the pending catastrophe, if any.
// Reports whether one was pending.
func (d *Dashboard) reopenCatastropheModal() bool {
	st := d.engine.GetState()
	if st.PendingCatastrophe == "" {
		return false
	}
	d.catModalShown = st.PendingCatastrophe
	d.showCatastropheModal(st.PendingCatastrophe)
	return true
}

// legacyBonusText returns a short summary of the epoch legacy bonus, in sorted
// resource order.
func legacyBonusText(epochKey string) string {
	bonuses := config.LegacyBonusForEpoch(epochKey)
	if len(bonuses) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(bonuses))
	for res := range bonuses {
		keys = append(keys, res)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, res := range keys {
		parts = append(parts, fmt.Sprintf("%s +%.0f%%", res, bonuses[res]*100))
	}
	return strings.Join(parts, ", ")
}
