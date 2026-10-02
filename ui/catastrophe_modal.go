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
// Succumb in the same epoch grants no new legacy or research bonus),
// researchNow is the current Succumb research bonus (e.g. 0.25), and eo is
// what Endure would cost with the Harbinger's Brace and the garrison counted
// (game.GameState.PendingEndure). st is the snapshot, for wall-clock durations.
func buildCatastropheModalLayout(epochKey string, alreadyLegacy bool, researchNow float64, eo game.EndureOutcome, st game.GameState) catastropheModalLayout {
	catName, catFlavor := config.CatastropheInfo(epochKey)
	ep := config.EpochByKey()[epochKey]
	inner := catastropheModalWidth - 2

	headerLines := []string{fmt.Sprintf("[red]☄ %s[-]", tview.Escape(catName))}
	for _, l := range tview.WordWrap(catFlavor, inner-2) {
		headerLines = append(headerLines, "[gray]"+tview.Escape(strings.TrimSpace(l))+"[-]")
	}

	endureLines := []string{
		"[white]── ENDURE: weather the catastrophe ──[-]",
		fmt.Sprintf("  [red]• %s of buildings destroyed (wonders and storage are spared)[-]", endurePct(eo.DestroyPct)),
		fmt.Sprintf("  [red]• All resources reduced to %s[-]", endurePct(eo.KeepFrac*100)),
		fmt.Sprintf("  [red]• %.0f%% of workers lost; workers of destroyed buildings go idle[-]", game.EndureWorkerLoss*100),
		fmt.Sprintf("  [red]• All production %.0f%% for %s, morale %+.0f points[-]",
			game.EndureDebuffProduction*100, formatTicks(game.EndureDebuffTicksIn(st.Age), st), game.EndureMoraleHit*100),
		"  [green]✓ Age, research, wonders and prestige preserved[-]",
		"  [green]✓ The epoch badge records that you endured[-]",
	}
	endureLines = append(endureLines, endureDefenseLines(eo)...)
	endure := strings.Join(endureLines, "\n")

	per := game.SuccumbResearchBonusPerEpoch
	researchLine := fmt.Sprintf("  [green]✓ Ancient Knowledge: research speed +%.0f%% (total +%.0f%%, permanent)[-]", per*100, (researchNow+per)*100)
	legacyLine := fmt.Sprintf("  [gold]✓ %s legacy: %s (permanent)[-]", ep.Name, legacyBonusText(epochKey))
	if alreadyLegacy {
		researchLine = fmt.Sprintf("  [gray]• Research bonus already earned here (stays +%.0f%%)[-]", researchNow*100)
		legacyLine = fmt.Sprintf("  [gray]• %s legacy already held; no new legacy bonus[-]", ep.Name)
	}
	succumb := strings.Join([]string{
		"[white]── SUCCUMB: let civilization fall ──[-]",
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

// buildLastPassageModalLayout assembles the text for the Last Passage, the
// Cosmic Era's catastrophe: prestige waits on it, Endure keeps a share of the
// run's points, Succumb trades them all for the Cosmic Legacy (closed once the
// player carries it).
func buildLastPassageModalLayout(lp game.LastPassageState) catastropheModalLayout {
	name, flavorText := config.LastPassageInfo()
	inner := catastropheModalWidth - 2

	headerLines := []string{fmt.Sprintf("[red]☄ %s[-]", tview.Escape(name))}
	for _, l := range tview.WordWrap(flavorText, inner-2) {
		headerLines = append(headerLines, "[gray]"+tview.Escape(strings.TrimSpace(l))+"[-]")
	}
	headerLines = append(headerLines, "[gray]Prestige waits on your answer. Nothing else does.[-]")

	keepLine := fmt.Sprintf("  [red]• You keep %d%% of this run's prestige points: %d of %d[-]", lp.KeepPct, lp.PointsIfEndured, lp.PointsNow)
	endureLines := []string{
		"[white]── ENDURE: pass through, diminished ──[-]",
		keepLine,
	}
	if lp.BraceLevel > 0 {
		endureLines = append(endureLines, fmt.Sprintf("  [green]✓ Braced (level %d): %d%% kept instead of %.0f%%[-]", lp.BraceLevel, lp.KeepPct, game.LastPassageKeep*100))
	}
	endureLines = append(endureLines, "  [green]✓ Level, upgrades, legacies and ruins carry over as always[-]")

	succumbLines := []string{
		"[white]── SUCCUMB: let it take the run ──[-]",
		"  [red]• Prestige completes with no points from this run (level still rises)[-]",
	}
	if lp.CosmicLegacy {
		succumbLines = append(succumbLines,
			"  [gray]• You already carry the Cosmic Legacy. Succumb is closed to you.[-]")
	} else {
		succumbLines = append(succumbLines,
			fmt.Sprintf("  [gold]✓ Cosmic Legacy: production +%.0f%%, permanent, through every prestige[-]", game.CosmicLegacyProductionBonus*100),
			"  [gold]✓ Earned once, kept forever[-]")
	}

	l := catastropheModalLayout{
		title:   " ✦ " + name + " ",
		header:  strings.Join(headerLines, "\n"),
		endure:  strings.Join(endureLines, "\n"),
		succumb: strings.Join(succumbLines, "\n"),
		hint:    "[gray]Esc: decide later · prestige waits · type 'catastrophe' to reopen[-]",
	}
	l.height = lineCount(l.header) + 1 + lineCount(l.endure) + 1 + lineCount(l.succumb) + 1 + 1 + 1 + 2
	return l
}

// pendingChoiceKey is the key of the choice the catastrophe modal waits on: the
// pending catastrophe's epoch key, config.LastPassageKey for a pending Last
// Passage, or "" when nothing waits.
func pendingChoiceKey(state game.GameState) string {
	switch {
	case state.PendingCatastrophe != "":
		return state.PendingCatastrophe
	case state.LastPassage.Pending:
		return config.LastPassageKey
	}
	return ""
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
// `catastrophe` command reopens it). key is the pending epoch key, or
// config.LastPassageKey for the Last Passage variant (prestige waits instead;
// Succumb is disabled once the Cosmic Legacy is held). Must be called from the
// UI goroutine.
func (d *Dashboard) showCatastropheModal(key string) {
	if d.pages.HasPage(catastrophePage) {
		d.pages.ShowPage(catastrophePage)
		d.pages.SendToFront(catastrophePage)
		if d.catFocus != nil {
			d.app.SetFocus(d.catFocus)
		}
		return
	}
	st := d.engine.GetState()
	endure, succumb := d.engine.Endure, d.engine.Succumb
	succumbOpen := true
	var l catastropheModalLayout
	if key == config.LastPassageKey {
		l = buildLastPassageModalLayout(st.LastPassage)
		endure, succumb = d.engine.EndureLastPassage, d.engine.SuccumbLastPassage
		succumbOpen = !st.LastPassage.CosmicLegacy
	} else {
		outcome := game.DefaultEndureOutcome()
		if st.PendingCatastrophe == key {
			outcome = st.PendingEndure
		}
		l = buildCatastropheModalLayout(key, st.LegacyBonuses[key], st.SuccumbResearchBonus, outcome, st)
	}

	btnEndure := tview.NewButton(tview.Escape("[E] ENDURE")).
		SetSelectedFunc(func() {
			if err := endure(); err != nil {
				d.engine.AddLog("error", "Endure failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
	styleFilledButton(btnEndure, theme.RoleNegative)

	btnSuccumb := tview.NewButton(tview.Escape("[S] SUCCUMB"))
	if succumbOpen {
		btnSuccumb.SetSelectedFunc(func() {
			if err := succumb(); err != nil {
				d.engine.AddLog("error", "Succumb failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
		styleFilledButton(btnSuccumb, theme.RoleNegative)
	} else {
		// Closed: drawn in the chip fill, never focused, never pressed.
		styleFilledButton(btnSuccumb, theme.RoleChip)
	}

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
	if !succumbOpen {
		focusOrder = focusOrder[:1]
	}
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
				if !succumbOpen {
					return nil
				}
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
// back (returnFocus: the icons window if open, else an open panel, else the
// command input). catModalShown stays set, so refresh() does not re-pop the
// modal for the same pending catastrophe; `catastrophe` reopens it.
func (d *Dashboard) closeCatastropheModal() {
	d.pages.RemovePage(catastrophePage)
	d.catFocus = nil
	d.returnFocus()
}

// reopenCatastropheModal shows the modal for the pending catastrophe or Last
// Passage, if any. Reports whether one was pending.
func (d *Dashboard) reopenCatastropheModal() bool {
	key := pendingChoiceKey(d.engine.GetState())
	if key == "" {
		return false
	}
	d.catModalShown = key
	d.showCatastropheModal(key)
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

// endurePct prints an Endure percentage: whole numbers plainly ("20%"), a
// garrison-softened share to one decimal ("16.5%").
func endurePct(p float64) string {
	if p == float64(int(p)) {
		return fmt.Sprintf("%d%%", int(p))
	}
	return fmt.Sprintf("%.1f%%", p)
}

// endureDefenseLines are the Endure lines about Brace and the garrison: what
// each already took off the numbers above, from the player's side.
func endureDefenseLines(o game.EndureOutcome) []string {
	var lines []string
	if o.BraceLevel > 0 {
		lines = append(lines, fmt.Sprintf("  [green]✓ Braced (level %d): %s fall, %s kept, before your garrison[-]",
			o.BraceLevel, endurePct(o.BracedDestroyPct), endurePct(o.BracedKeepFrac*100)))
	}
	if o.Garrison <= 0 {
		lines = append(lines, "  [gray]• No garrison: soldiers would soften this (see the Army panel)[-]")
		return lines
	}
	if o.BuildingsSaved > 0 {
		lines = append(lines, fmt.Sprintf("  [green]✓ Your garrison saves %d %s and keeps %s of stock, not %s[-]",
			o.BuildingsSaved, pluralize("building", o.BuildingsSaved), endurePct(o.KeepFrac*100), endurePct(o.BracedKeepFrac*100)))
	} else {
		lines = append(lines, fmt.Sprintf("  [green]✓ Your garrison keeps %s of stock, not %s, and softens the fall[-]",
			endurePct(o.KeepFrac*100), endurePct(o.BracedKeepFrac*100)))
	}
	if o.Capped {
		lines = append(lines, fmt.Sprintf("  [gray]• Brace and garrison together soften an Endure by at most %.0f%%[-]", config.EndureReductionCap*100))
	}
	return lines
}
