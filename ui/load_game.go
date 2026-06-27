package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// loadGamePage is the page name under which the Load Game browser is registered
// on the root Pages. It is added on demand (from the splash handler) and removed
// on Back/Load so the save list is always rebuilt fresh.
const loadGamePage = "load_game"

// loadGameBrowser holds the widgets and live state for the Load Game screen.
// Everything the input handlers and refresh logic touch hangs off this struct so
// the page can be rebuilt cheaply and selection restored after delete/rename/dup.
type loadGameBrowser struct {
	app    *tview.Application
	pages  *tview.Pages
	engine *game.GameEngine

	root     *tview.Flex
	list     *tview.List
	detail   *tview.TextView
	subtitle *tview.TextView

	saves []game.SaveInfo // current rows, sorted most-recent first
}

// CreateLoadGamePage builds the full-screen Load Game browser and returns its
// root primitive. The caller registers it (e.g. pages.AddPage(loadGamePage, ...))
// and sets focus to the returned primitive's list — call SetFocus on the value
// returned by FocusTarget, or just rely on AddPage's focus when it is the only
// page shown. The screen is self-contained: all key handling lives on the list.
func CreateLoadGamePage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine) tview.Primitive {
	b := &loadGameBrowser{
		app:    app,
		pages:  pages,
		engine: engine,
	}

	// ── Title ────────────────────────────────────────────────────────────────
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gold]═══ Load Game ═══[-]")

	b.subtitle = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// ── Save list ────────────────────────────────────────────────────────────
	b.list = tview.NewList()
	b.list.SetBorder(false)
	b.list.SetSelectedBackgroundColor(tcell.ColorGold)
	b.list.SetSelectedTextColor(tcell.ColorBlack)
	b.list.ShowSecondaryText(false)
	b.list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		b.updateDetail(index)
	})

	// ── Key / legend ─────────────────────────────────────────────────────────
	// Static box explaining the row symbols. Colours match the row tags exactly
	// (gold for ★/⭐, red for ⚠). Never changes, so no refresh wiring.
	legend := tview.NewTextView().
		SetDynamicColors(true).
		SetText(legendText())
	legend.SetBorder(true).
		SetBorderColor(tcell.ColorGold).
		SetTitle(" Key ").
		SetTitleColor(tcell.ColorGold)

	// ── Detail pane ──────────────────────────────────────────────────────────
	b.detail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	b.detail.SetBorder(true).
		SetBorderColor(tcell.ColorGold).
		SetTitle(" Details ").
		SetTitleColor(tcell.ColorGold)

	// ── Footer ───────────────────────────────────────────────────────────────
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(footerBar())

	// ── Layout ───────────────────────────────────────────────────────────────
	b.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(b.subtitle, 1, 0, false).
		AddItem(b.list, 0, 1, true).     // weighted — takes remaining space
		AddItem(legend, 6, 0, false).    // 4 content lines + border
		AddItem(b.detail, 10, 0, false). // fits 6 content lines + optional badge + border
		AddItem(footer, 1, 0, false)

	b.list.SetInputCapture(b.handleKey)

	b.refresh(0)
	return b.root
}

// refresh re-reads the save listing, re-sorts most-recent first, rebuilds the
// list rows, and restores a clamped selection. wantIdx is the index to select
// after the rebuild (clamped to range); pass the current index to keep position.
func (b *loadGameBrowser) refresh(wantIdx int) {
	saves, err := game.ListSaveDetails()
	if err != nil {
		// A listing error is rare (the saves dir not existing returns nil,nil).
		// Surface it in the detail pane rather than crashing, and show an empty list.
		b.saves = nil
		b.list.Clear()
		b.subtitle.SetText("[#8b949e]" + savesDirLabel() + " — could not read saves[-]")
		b.detail.SetText(fmt.Sprintf("[red]Could not read saves: %v[-]", err))
		return
	}

	// Sort most-recent first by Timestamp.
	sort.SliceStable(saves, func(i, j int) bool {
		return saves[i].Timestamp.After(saves[j].Timestamp)
	})
	b.saves = saves

	b.subtitle.SetText(fmt.Sprintf("[#8b949e]%s — %s[-]", savesDirLabel(), pluralSaves(len(saves))))

	b.list.Clear()
	if len(saves) == 0 {
		// Empty state — the action keys become no-ops (handleKey guards on len).
		b.detail.SetText("[#8b949e]No saved games found in " + savesDir() + ".\n\nStart a new game to create one.[-]")
		return
	}

	for _, s := range saves {
		b.list.AddItem(rowLabel(s), "", 0, nil)
	}

	// Restore a sensible selection (clamp to range).
	if wantIdx < 0 {
		wantIdx = 0
	}
	if wantIdx >= len(saves) {
		wantIdx = len(saves) - 1
	}
	b.list.SetCurrentItem(wantIdx)
	// SetCurrentItem fires SetChangedFunc only when the index actually changes;
	// force the detail pane to reflect the (possibly unchanged) selection.
	b.updateDetail(wantIdx)
}

// updateDetail renders the detail pane for the save at index. Out-of-range or
// empty selections render nothing harmful.
func (b *loadGameBrowser) updateDetail(index int) {
	if index < 0 || index >= len(b.saves) {
		return
	}
	b.detail.SetText(detailText(b.saves[index]))
}

// selected returns the currently selected SaveInfo and true, or a zero value and
// false when the list is empty / selection is out of range.
func (b *loadGameBrowser) selected() (game.SaveInfo, bool) {
	idx := b.list.GetCurrentItem()
	if idx < 0 || idx >= len(b.saves) {
		return game.SaveInfo{}, false
	}
	return b.saves[idx], true
}

// handleKey routes the action keys for the browser. tview.List handles ↑/↓
// natively; we intercept Enter/d/r/c/Esc/q.
func (b *loadGameBrowser) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		b.doLoad()
		return nil
	case tcell.KeyEsc:
		b.back()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'd', 'D':
			b.doDelete()
			return nil
		case 'r', 'R':
			b.doRename()
			return nil
		case 'c', 'C':
			b.doDuplicate()
			return nil
		case 'q', 'Q':
			b.back()
			return nil
		}
	}
	return event
}

// back removes the load_game page and returns to the splash.
func (b *loadGameBrowser) back() {
	b.pages.RemovePage(loadGamePage)
	b.pages.SwitchToPage("splash")
}

// doLoad loads the selected save and transitions to the dashboard. Corrupt saves
// are refused with an inline message. Load errors are shown inline and stay.
func (b *loadGameBrowser) doLoad() {
	s, ok := b.selected()
	if !ok {
		return
	}
	if s.Corrupt {
		b.detail.SetText(detailText(s) + "\n\n[red]Can't load a corrupt save.[-]")
		return
	}
	if err := b.engine.LoadGame(s.Name); err != nil {
		b.engine.AddLog("error", fmt.Sprintf("Load failed: %v", err))
		b.detail.SetText(detailText(s) + fmt.Sprintf("\n\n[red]Load failed: %v[-]", err))
		return
	}
	b.engine.AddLog("success", "Game loaded!")
	// Mirror splash.go's post-load transition exactly.
	b.pages.RemovePage(loadGamePage)
	b.pages.SwitchToPage("dashboard")
	go b.engine.Start()
}

// doDelete shows a red confirm modal, then deletes on confirm and refreshes.
func (b *loadGameBrowser) doDelete() {
	s, ok := b.selected()
	if !ok {
		return
	}
	curIdx := b.list.GetCurrentItem()

	const page = "load_game_delete"
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Delete '%s'?\n\nThis can't be undone.", s.Name)).
		AddButtons([]string{"Cancel", "Delete"}).
		SetDoneFunc(func(_ int, label string) {
			b.pages.RemovePage(page)
			b.app.SetFocus(b.list)
			if label != "Delete" {
				return
			}
			if err := game.DeleteSave(s.Name); err != nil {
				b.engine.AddLog("error", fmt.Sprintf("Delete failed: %v", err))
				b.detail.SetText(fmt.Sprintf("[red]Delete failed: %v[-]", err))
				return
			}
			// Keep the selection near where it was; clamp happens in refresh.
			b.refresh(curIdx)
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	b.pages.AddPage(page, modal, true, true)
}

// doRename shows an input modal prefilled with the current name. Errors from
// RenameSave (collision/invalid) surface inline in the dialog and let the player
// retry. On success the dialog closes, the list refreshes, and the renamed item
// stays selected.
func (b *loadGameBrowser) doRename() {
	s, ok := b.selected()
	if !ok {
		return
	}
	if s.Corrupt {
		// A corrupt file can still be renamed at the FS level, but there's no
		// healthy use for it; keep behaviour simple and allow it — RenameSave
		// only touches the filename, not the bytes.
	}

	const page = "load_game_rename"

	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	input := tview.NewInputField().
		SetLabel("Name: ").
		SetText(s.Name).
		SetFieldWidth(40)

	close := func() {
		b.pages.RemovePage(page)
		b.app.SetFocus(b.list)
	}

	submit := func() {
		newName := strings.TrimSpace(input.GetText())
		if err := game.RenameSave(s.Name, newName); err != nil {
			errTV.SetText(fmt.Sprintf("[red]%v[-]", err))
			b.app.SetFocus(input)
			return
		}
		b.pages.RemovePage(page)
		b.app.SetFocus(b.list)
		// Refresh, then re-select the renamed item by its new name.
		b.refresh(0)
		b.selectByName(newName)
	}

	okBtn := tview.NewButton("[ OK ]").SetSelectedFunc(submit)
	okBtn.SetBackgroundColor(tcell.ColorGold)
	okBtn.SetLabelColor(tcell.ColorBlack)
	cancelBtn := tview.NewButton("[ Cancel ]").SetSelectedFunc(close)

	// Enter in the input field submits.
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			submit()
		}
	})

	btnRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(okBtn, 8, 0, false).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(cancelBtn, 12, 0, false).
		AddItem(tview.NewBox(), 0, 1, false)

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btnRow, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(" Rename Save ").
		SetTitleColor(tcell.ColorGold).
		SetBorderColor(tcell.ColorGold)

	// Tab cycles input → OK → Cancel; Esc cancels from anywhere on the modal.
	focusOrder := []tview.Primitive{input, okBtn, cancelBtn}
	focusIdx := 0
	modal := centeredModal(inner, 50, 9)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			close()
			return nil
		case tcell.KeyTab:
			focusIdx = (focusIdx + 1) % len(focusOrder)
			b.app.SetFocus(focusOrder[focusIdx])
			return nil
		case tcell.KeyBacktab:
			focusIdx = (focusIdx - 1 + len(focusOrder)) % len(focusOrder)
			b.app.SetFocus(focusOrder[focusIdx])
			return nil
		}
		return ev
	})

	b.pages.AddPage(page, modal, true, true)
	b.app.SetFocus(input)
}

// doDuplicate duplicates the selected save with no prompt, refreshes, and selects
// the new copy.
func (b *loadGameBrowser) doDuplicate() {
	s, ok := b.selected()
	if !ok {
		return
	}
	newName, err := game.DuplicateSave(s.Name)
	if err != nil {
		b.engine.AddLog("error", fmt.Sprintf("Duplicate failed: %v", err))
		b.detail.SetText(fmt.Sprintf("[red]Duplicate failed: %v[-]", err))
		return
	}
	b.refresh(0)
	b.selectByName(newName)
}

// selectByName moves the selection to the row whose save name matches, if present.
func (b *loadGameBrowser) selectByName(name string) {
	for i, s := range b.saves {
		if s.Name == name {
			b.list.SetCurrentItem(i)
			b.updateDetail(i)
			return
		}
	}
}

// ── Pure helpers (formatting) ───────────────────────────────────────────────

// rowLabel renders one list row: name, age, relative time, and a trailing tag.
// Corrupt rows are dimmed grey.
func rowLabel(s game.SaveInfo) string {
	if s.Corrupt {
		return fmt.Sprintf("[gray]%s   —   %s   ⚠ corrupt[-]", s.Name, relativeTime(s.Timestamp))
	}
	age := ageDisplay(s.Age)
	tag := rowTag(s)
	if tag != "" {
		tag = "   " + tag
	}
	return fmt.Sprintf("%s   [#8b949e]%s   %s[-]%s", s.Name, age, relativeTime(s.Timestamp), tag)
}

// rowTag returns the trailing status tag for a (non-corrupt) save row, or "".
// Precedence: autosave first, then modified, then elite.
// footerButton renders one keycap-style action button: the hotkey on a gold
// cap fused to its label on a dark chip — e.g. a gold "Enter" beside "Load".
func footerButton(key, label string) string {
	return fmt.Sprintf("[black:gold:b] %s [white:#30363d:b] %s [-:-:-]", key, label)
}

// footerBar is the Load Game action bar: a keycap button per action so the
// player can see at a glance which key triggers what.
func footerBar() string {
	return "  " + strings.Join([]string{
		footerButton("↑↓", "Navigate"),
		footerButton("Enter", "Load"),
		footerButton("D", "Delete"),
		footerButton("R", "Rename"),
		footerButton("C", "Duplicate"),
		footerButton("Esc", "Back"),
	}, "  ") + "  "
}

// legendText returns the static Key-box markup: one row symbol per line with a
// short explanation. Colours match the row tags exactly (gold for ★/⭐, red for
// ⚠) so the box reads as a direct legend for what's shown in the list.
func legendText() string {
	return strings.Join([]string{
		"[gold]★ auto[-]      automatic save slot (overwritten on autosave)",
		"[gold]⭐ elite[-]     an elite save",
		"[red]⚠ modified[-]  save file edited outside the game",
		"[red]⚠ corrupt[-]   file could not be read — cannot be loaded",
	}, "\n")
}

func rowTag(s game.SaveInfo) string {
	if s.Name == game.AutosaveName {
		return "[gold]★ auto[-]"
	}
	if s.Modified {
		return "[red]⚠ modified[-]"
	}
	if s.Elite {
		return "[gold]⭐ elite[-]"
	}
	return ""
}

// detailSep is the gold middle-dot separator between detail-pane segments.
const detailSep = " [gold]·[-] "

// detailText renders the detail pane for a save. Healthy saves show a rich stat
// block (identity, population/structures, progress, prestige/morale, an optional
// catastrophe warning, and save metadata); corrupt saves show only the
// unloadable notice + file time.
func detailText(s game.SaveInfo) string {
	if s.Corrupt {
		return fmt.Sprintf(
			"[red]⚠ Corrupt save — cannot be loaded[-]\n[#8b949e]File time: %s[-]",
			s.Timestamp.Format("Jan 2, 2006 3:04 PM"),
		)
	}

	var lines []string

	// Line 1 — identity: "Title" · Age · Epoch. Drop the quoted title when empty
	// so the line starts cleanly with the Age (no empty quotes).
	var id []string
	if s.Title != "" {
		id = append(id, fmt.Sprintf("[gold]\"%s\"[-]", s.Title))
	}
	id = append(id, fmt.Sprintf("[white]%s[-]", ageDisplay(s.Age)))
	if s.Epoch != "" {
		id = append(id, fmt.Sprintf("[#8b949e]%s[-]", s.Epoch))
	}
	lines = append(lines, strings.Join(id, detailSep))

	// Line 2 — civilisation footprint.
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[#8b949e]Population[-] [white]%s[-]", commafy(s.Population)),
		fmt.Sprintf("[#8b949e]Buildings[-] [white]%s[-]", commafy(s.Buildings)),
		fmt.Sprintf("[#8b949e]Wonders[-] [white]%s[-]", commafy(s.Wonders)),
	}, detailSep))

	// Line 3 — progress markers. Milestones show "done/total" only when the total
	// is known (config accessor available).
	milestones := commafy(s.MilestonesDone)
	if s.MilestonesTotal > 0 {
		milestones = fmt.Sprintf("%s/%s", commafy(s.MilestonesDone), commafy(s.MilestonesTotal))
	}
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[#8b949e]Milestones[-] [white]%s[-]", milestones),
		fmt.Sprintf("[#8b949e]Techs[-] [white]%s[-]", commafy(s.Techs)),
		fmt.Sprintf("[#8b949e]Soldiers[-] [white]%s[-]", commafy(s.Soldiers)),
	}, detailSep))

	// Line 4 — prestige + morale.
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[#8b949e]Prestige[-] [white]Lv %s[-] [#8b949e](%s pts)[-]", commafy(s.PrestigeLevel), commafy(s.PrestigeTotal)),
		fmt.Sprintf("[#8b949e]Morale[-] [white]%.0f%%[-]", s.Morale*100),
	}, detailSep))

	// Line 5 — looming catastrophe warning (omitted entirely when none pending).
	if s.PendingCatastrophe != "" {
		lines = append(lines, fmt.Sprintf("[red]⚠ Pending: %s[-]", s.PendingCatastrophe))
	}

	// Line 6 — save metadata.
	lines = append(lines, fmt.Sprintf(
		"[#8b949e]Saved[-] %s [gold]·[-] [#8b949e]%s ticks[-]",
		s.Timestamp.Format("Jan 2, 2006 3:04 PM"), commafy(s.Tick),
	))

	out := strings.Join(lines, "\n")
	if badges := detailBadges(s); badges != "" {
		out += "\n" + badges
	}
	return out
}

// detailBadges returns the badge line for the detail pane, or "".
func detailBadges(s game.SaveInfo) string {
	var parts []string
	if s.Name == game.AutosaveName {
		parts = append(parts, "[gold]★ autosave[-]")
	}
	if s.Elite {
		parts = append(parts, "[gold]⭐ elite[-]")
	}
	if s.Modified {
		parts = append(parts, "[red]⚠ modified[-]")
	}
	return strings.Join(parts, "   ")
}

// ageDisplay maps an age key to its display name (e.g. "stone_age" → "Stone Age").
// Falls back to the raw key, or an em-dash when empty.
func ageDisplay(key string) string {
	if key == "" {
		return "—"
	}
	if def, ok := config.AgeByKey()[key]; ok {
		return def.Name
	}
	return key
}

// relativeTime renders a coarse human-friendly age for a timestamp, e.g.
// "just now", "5m ago", "2h ago", "yesterday", "3 days ago". A zero timestamp
// (possible on a corrupt file with no readable mtime) renders "unknown".
func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < 0:
		// Clock skew / future-stamped save — treat as just now.
		return "just now"
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d / time.Minute)
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d / time.Hour)
		return fmt.Sprintf("%dh ago", h)
	case d < 48*time.Hour:
		return "yesterday"
	default:
		days := int(d / (24 * time.Hour))
		return fmt.Sprintf("%d days ago", days)
	}
}

// commafy formats a non-negative int with thousands separators (e.g. 12400 →
// "12,400"). Negative inputs are formatted with a leading minus.
func commafy(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var out strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		out.WriteString(s[:pre])
		if len(s) > pre {
			out.WriteByte(',')
		}
	}
	for i := pre; i < len(s); i += 3 {
		out.WriteString(s[i : i+3])
		if i+3 < len(s) {
			out.WriteByte(',')
		}
	}
	if neg {
		return "-" + out.String()
	}
	return out.String()
}

// pluralSaves renders the save-count label ("1 save" / "N saves").
func pluralSaves(n int) string {
	if n == 1 {
		return "1 save"
	}
	return fmt.Sprintf("%d saves", n)
}

// savesDir returns the player-facing saves folder label used in messages.
func savesDir() string {
	return "./data/saves/"
}

// savesDirLabel is the subtitle prefix (same folder, kept as its own helper so
// the call sites read clearly).
func savesDirLabel() string {
	return savesDir()
}

// centeredModal wraps an inner primitive in nested Flex spacers so it floats at a
// fixed width×height in the centre of the screen — the same idiom as
// catastrophe_modal.go. width/height are fixed cell counts.
func centeredModal(inner tview.Primitive, width, height int) *tview.Flex {
	return tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(inner, height, 0, true).
			AddItem(tview.NewBox(), 0, 1, false),
			width, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)
}
