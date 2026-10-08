package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
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

	// cur is the save the main menu's Continue opens (game.CurrentGame), and
	// mainGame the save the player marked as the main game ("" for none).
	cur      game.CurrentGame
	hasCur   bool
	mainGame string

	// backPage is the page to return to on Back/Esc (e.g. "splash" from the menu,
	// "dashboard" when opened mid-game). startOnLoad gates engine.Start() after a
	// successful load: true from the splash (first start), false mid-game (the
	// engine is already running — a second Start() would double-run the ticker).
	backPage    string
	startOnLoad bool
}

// CreateLoadGamePage builds the full-screen Load Game browser and returns its
// root primitive. The caller registers it (e.g. pages.AddPage(loadGamePage, ...))
// and sets focus to the returned primitive's list — call SetFocus on the value
// returned by FocusTarget, or just rely on AddPage's focus when it is the only
// page shown. The screen is self-contained: all key handling lives on the list.
//
// backPage names the page to return to on Back/Esc; startOnLoad decides whether
// a successful load should start the engine (true from the splash for the first
// start, false mid-game where the engine is already running).
func CreateLoadGamePage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, backPage string, startOnLoad bool) tview.Primitive {
	b := &loadGameBrowser{
		app:         app,
		pages:       pages,
		engine:      engine,
		backPage:    backPage,
		startOnLoad: startOnLoad,
	}

	// ── Title ────────────────────────────────────────────────────────────────
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gold]═══ Load game ═══[-]")

	b.subtitle = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// ── Save list ────────────────────────────────────────────────────────────
	b.list = tview.NewList()
	b.list.SetBorder(false)
	// Selection backs light text, so it uses the Selection role (dark slate under
	// Forge), not Accent — white-on-gold is unreadable (theme/themes_forge.go).
	theme.Track(func() {
		b.list.SetSelectedBackgroundColor(theme.Color(theme.RoleSelection)).
			SetSelectedTextColor(theme.Color(theme.RoleText))
	})
	b.list.ShowSecondaryText(false)
	b.list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		b.updateDetail(index)
	})

	// ── Key / legend ─────────────────────────────────────────────────────────
	// Static box explaining the row symbols. Colours match the row tags exactly
	// (gold for ★, red for ⚠). Never changes, so no refresh wiring.
	legend := tview.NewTextView().
		SetDynamicColors(true).
		SetText(legendText())
	legend.SetBorder(true).
		SetTitle(" Key ")
	theme.Track(func() {
		legend.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent))
	})

	// ── Detail pane ──────────────────────────────────────────────────────────
	b.detail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	b.detail.SetBorder(true).
		SetTitle(" Details ")
	theme.Track(func() {
		b.detail.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent))
	})

	// ── Footer ───────────────────────────────────────────────────────────────
	// Written for the width it is drawn at, so no key is cut off at 80 columns.
	footer := newFitView(func(w, h int) string { return keyBar(loadGameKeys, w, h) })
	footer.SetTextAlign(tview.AlignCenter)
	footer.changed()

	// ── Layout ───────────────────────────────────────────────────────────────
	b.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(b.subtitle, 1, 0, false).
		AddItem(b.list, 0, 1, true).     // weighted — takes remaining space
		AddItem(legend, 7, 0, false).    // 5 content lines + border
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
		b.subtitle.SetText("[gray]" + savesDirLabel() + " · could not read saves[-]")
		b.detail.SetText(fmt.Sprintf("[red]Could not read saves: %v[-]", err))
		return
	}

	// Arrange the flat listing into a depth-first lineage forest. The rows carry
	// their tree-connector prefixes; b.saves stays 1:1 with rendered rows in this
	// same order so every handler still maps row i → b.saves[i].
	rows := buildSaveTree(saves)
	b.saves = make([]game.SaveInfo, len(rows))
	for i, r := range rows {
		b.saves[i] = r.Info
	}

	// Which save the main menu's Continue opens, and why: the subtitle says
	// it, and its row carries the mark.
	b.cur, b.hasCur = game.PickCurrentGame(saves, "", "")
	b.mainGame = ""
	if acct := b.engine.Account(); acct != nil {
		var last string
		b.mainGame, last = acct.GameRecord()
		b.cur, b.hasCur = game.PickCurrentGame(saves, b.mainGame, last)
	}
	b.subtitle.SetText(fmt.Sprintf("[gray]%s · %s%s[-]", savesDirLabel(), pluralSaves(len(saves)), b.continueNote()))

	b.list.Clear()
	if len(rows) == 0 {
		// Empty state — the action keys become no-ops (handleKey guards on len).
		b.detail.SetText("[gray]This account has no saved games yet.\n\nStart a new game to create one.[-]")
		return
	}

	// The active save is the one the autosave currently follows; mark its row.
	activeName := b.engine.ActiveSaveName()
	for _, r := range rows {
		b.list.AddItem(rowLabel(r.Info, r.Prefix, r.Info.Name == activeName)+b.currentTag(r.Info), "", 0, nil)
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

// continueNote is the end of the subtitle: which save Continue opens.
func (b *loadGameBrowser) continueNote() string {
	if !b.hasCur {
		return ""
	}
	return " · Continue opens " + b.cur.Save.Name
}

// currentTag is the trailing tag of the row Continue opens: the main game's
// mark when the player marked it, else a plain "continue".
func (b *loadGameBrowser) currentTag(s game.SaveInfo) string {
	switch {
	case s.Corrupt:
		return ""
	case s.Name == b.mainGame && b.mainGame != "":
		return "   [accent]◆ main game[-]"
	case b.hasCur && s.Name == b.cur.Save.Name:
		return "   [label]▸ continue[-]"
	}
	return ""
}

// currentLine is the detail pane's line for the save Continue opens, with
// the reason, or "" for any other save.
func (b *loadGameBrowser) currentLine(s game.SaveInfo) string {
	if !b.hasCur || s.Name != b.cur.Save.Name {
		return ""
	}
	switch b.cur.Why {
	case game.CurrentMain:
		return "[accent]◆ The main game:[-] [gray]Continue opens it. M unmarks it.[-]"
	case game.CurrentLast:
		return "[label]▸ Continue opens this game,[-] [gray]the one last played. M marks it as the main game.[-]"
	}
	return "[label]▸ Continue opens this game,[-] [gray]the newest save. M marks it as the main game.[-]"
}

// detailFor is the detail pane's text for a save: its facts, then the line
// about Continue when it is the one.
func (b *loadGameBrowser) detailFor(s game.SaveInfo) string {
	out := detailText(s, b.parentPresent(s), b.engine.AccountID())
	if line := b.currentLine(s); line != "" && !s.Corrupt {
		out += "\n" + line
	}
	return out
}

// updateDetail renders the detail pane for the save at index. Out-of-range or
// empty selections render nothing harmful.
func (b *loadGameBrowser) updateDetail(index int) {
	if index < 0 || index >= len(b.saves) {
		return
	}
	b.detail.SetText(b.detailFor(b.saves[index]))
}

// parentPresent reports whether s names a lineage parent that is itself among
// the loaded saves (so the detail pane can mark a missing parent "detached").
func (b *loadGameBrowser) parentPresent(s game.SaveInfo) bool {
	if s.ParentName == "" || s.ParentName == s.Name {
		return false
	}
	for _, other := range b.saves {
		if other.Name == s.ParentName {
			return true
		}
	}
	return false
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
		case 'm', 'M':
			b.doMarkMain()
			return nil
		case 'q', 'Q':
			b.back()
			return nil
		}
	}
	return event
}

// back removes the load_game page and returns to the page it was opened from
// (splash from the menu, dashboard when opened mid-game).
func (b *loadGameBrowser) back() {
	b.pages.RemovePage(loadGamePage)
	b.pages.SwitchToPage(b.backPage)
}

// doLoad loads the selected save and transitions to the dashboard. Corrupt saves
// are refused with an inline message. Load errors are shown inline and stay.
func (b *loadGameBrowser) doLoad() {
	s, ok := b.selected()
	if !ok {
		return
	}
	if s.Corrupt {
		b.detail.SetText(b.detailFor(s) + "\n\n[red]Can't load a corrupt save.[-]")
		return
	}
	if err := b.engine.LoadGame(s.Name); err != nil {
		b.engine.AddLog("error", fmt.Sprintf("Load failed: %v", err))
		b.detail.SetText(b.detailFor(s) + fmt.Sprintf("\n\n[red]Load failed: %v[-]", err))
		return
	}
	b.engine.AddLog("success", "Game loaded.")
	b.pages.RemovePage(loadGamePage)
	b.pages.SwitchToPage("dashboard")
	// From the splash this performs the first engine start; mid-game the engine is
	// already running (LoadGame swapped state under the lock and the live ticker
	// picks it up), so a second Start() would double-run the ticker.
	if b.startOnLoad {
		go b.engine.Start()
	}
}

// doMarkMain marks the selected save as the main game, the one the main
// menu's Continue opens whatever was played last, or unmarks it when it is
// the main game already. The mark lives in the account's settings.
func (b *loadGameBrowser) doMarkMain() {
	s, ok := b.selected()
	if !ok {
		return
	}
	acct := b.engine.Account()
	switch {
	case s.Corrupt:
		b.detail.SetText(b.detailFor(s) + "\n\n[red]A corrupt save cannot be the main game.[-]")
		return
	case acct == nil:
		b.detail.SetText(b.detailFor(s) + "\n\n[red]The main game is kept on an account, and none is in use.[-]")
		return
	}
	name := s.Name
	if b.mainGame == s.Name {
		name = ""
	}
	if err := acct.SetMainGame(name); err != nil {
		b.detail.SetText(b.detailFor(s) + fmt.Sprintf("\n\n[red]The main game could not be saved to your account: %v[-]", err))
		return
	}
	b.refresh(b.list.GetCurrentItem())
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
			// The account's records of its current game let go of the name.
			if acct := b.engine.Account(); acct != nil {
				_ = acct.ForgetGame(s.Name)
			}
			// Keep the selection near where it was; clamp happens in refresh.
			b.refresh(curIdx)
		})
	// Transient confirm modal (rebuilt per delete): construction-read the danger
	// background from the Negative role so it tints with the theme.
	styleDangerModal(modal)
	b.pages.AddPage(page, modal, true, true)
}

// doRename shows an input modal prefilled with the current name. Errors from
// RenameSave (collision/invalid) surface inline in the dialog and let the player
// retry. On success the dialog closes, the list refreshes, and the renamed item
// stays selected.
//
// note: lineage is keyed by ParentName. RenameSave re-parents the children to the
// new name (and re-signs them), so the lineage follows the rename instead of
// detaching. We mirror that into the engine's in-memory active pointers below so
// the running session's next autosave doesn't re-orphan anything.
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
		// Keep the running session's lineage pointers consistent: if the renamed
		// save IS the active slot, point autosave at the new name; if it was this
		// run's parent, follow it so the next autosave doesn't re-orphan us.
		if b.engine.ActiveSaveName() == s.Name {
			b.engine.SetActiveSaveName(newName)
		}
		if b.engine.ActiveParentName() == s.Name {
			b.engine.SetActiveParentName(newName)
		}
		// And the account's records of its current game follow the name.
		if acct := b.engine.Account(); acct != nil {
			_ = acct.RenameGame(s.Name, newName)
		}
		b.pages.RemovePage(page)
		b.app.SetFocus(b.list)
		// Refresh, then re-select the renamed item by its new name.
		b.refresh(0)
		b.selectByName(newName)
	}

	okBtn := tview.NewButton("[ OK ]").SetSelectedFunc(submit)
	styleFilledButton(okBtn, theme.RoleAccent)
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
		SetTitle(" Rename save ").
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))

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

// rowLabel renders one list row: an optional tree-connector prefix, the name,
// age, relative time, and trailing tags. Corrupt rows are dimmed grey. When
// active is true (the save the autosave follows) a distinct aqua "● active" tag
// is appended, separate from the gold ★ auto tag.
func rowLabel(s game.SaveInfo, prefix string, active bool) string {
	if s.Corrupt {
		return fmt.Sprintf("[gray]%s%s   ·   %s   ⚠ corrupt[-]", prefix, s.Name, relativeTime(s.Timestamp))
	}
	age := ageDisplay(s.Age)
	tag := rowTag(s)
	if active {
		if tag != "" {
			tag += " "
		}
		tag += "[label]● active[-]"
	}
	if tag != "" {
		tag = "   " + tag
	}
	return fmt.Sprintf("%s%s   [gray]%s   %s[-]%s", prefix, s.Name, age, relativeTime(s.Timestamp), tag)
}

// rowTag returns the trailing status tag for a (non-corrupt) save row, or "".
// Precedence: autosave first, then modified, then elite.
// footerButton renders one keycap-style action button: the hotkey on a gold
// cap fused to its label on a dark chip — e.g. a gold "Enter" beside "Load".
func footerButton(key, label string) string {
	return theme.KeycapButton(key, label)
}

// footerKey is one action on a key bar: its key, its name, and a shorter
// name for a bar with little room ("" when the name is short already).
type footerKey struct{ key, label, short string }

// keyBar lays a key bar out for a box w cells wide and rows lines tall. It
// writes the fullest form that fits on one line: every name in full, then
// the shorter names, then those set closer. A bar still too long for one
// line is set in full over two when the box has them.
func keyBar(keys []footerKey, w, rows int) string {
	line := func(ks []footerKey, short bool, gap string) string {
		parts := make([]string, len(ks))
		for i, k := range ks {
			label := k.label
			if short && k.short != "" {
				label = k.short
			}
			parts[i] = footerButton(k.key, label)
		}
		return strings.Join(parts, gap)
	}
	for _, form := range []struct {
		short bool
		gap   string
	}{{false, "  "}, {true, "  "}, {true, " "}} {
		if l := line(keys, form.short, form.gap); visibleLen(l) <= w {
			return l
		}
	}
	if rows >= 2 {
		// Two lines, as even as the buttons allow.
		for _, short := range []bool{false, true} {
			for cut := (len(keys) + 1) / 2; cut < len(keys); cut++ {
				a, b := line(keys[:cut], short, "  "), line(keys[cut:], short, "  ")
				if visibleLen(a) <= w && visibleLen(b) <= w {
					return a + "\n" + b
				}
			}
		}
	}
	return line(keys, true, " ")
}

// loadGameKeys are the Load Game browser's actions.
var loadGameKeys = []footerKey{
	{"↑↓", "Navigate", "Move"},
	{"Enter", "Load", ""},
	{"D", "Delete", ""},
	{"R", "Rename", ""},
	{"C", "Duplicate", "Copy"},
	{"M", "Main game", "Main"},
	{"Esc", "Back", ""},
}

// footerBar is the Load Game action bar with every name in full: a keycap
// button per action, so the player can see at a glance which key triggers
// what. The page fits it to its width (keyBar).
func footerBar() string {
	return keyBar(loadGameKeys, 1<<16, 1)
}

// legendText returns the static Key-box markup: one row symbol per line with a
// short explanation. Colours match the row tags exactly (gold for ★, red for
// ⚠) so the box reads as a direct legend for what's shown in the list.
func legendText() string {
	return strings.Join([]string{
		"[gold]★ auto[-]      automatic save slot (overwritten on autosave)",
		"[label]● active[-]    the save your game is autosaving into",
		"[accent]◆ main game[-] Continue on the main menu opens it (M marks one)",
		"[red]⚠ modified[-]  save file edited outside the game",
		"[red]⚠ corrupt[-]   file could not be read, so it cannot load",
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
func detailText(s game.SaveInfo, parentPresent bool, currentAccountID string) string {
	if s.Corrupt {
		return fmt.Sprintf(
			"[red]⚠ Corrupt save. It cannot be loaded.[-]\n[gray]File time: %s[-]",
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
		id = append(id, fmt.Sprintf("[gray]%s[-]", s.Epoch))
	}
	lines = append(lines, strings.Join(id, detailSep))

	// Line 2 — civilisation footprint.
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[gray]Population[-] [white]%s[-]", textfmt.Int(s.Population)),
		fmt.Sprintf("[gray]Buildings[-] [white]%s[-]", textfmt.Int(s.Buildings)),
		fmt.Sprintf("[gray]Wonders[-] [white]%s[-]", textfmt.Int(s.Wonders)),
	}, detailSep))

	// Line 3 — progress markers. Milestones show "done/total" only when the total
	// is known (config accessor available).
	milestones := textfmt.Int(s.MilestonesDone)
	if s.MilestonesTotal > 0 {
		milestones = fmt.Sprintf("%s/%s", textfmt.Int(s.MilestonesDone), textfmt.Int(s.MilestonesTotal))
	}
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[gray]Milestones[-] [white]%s[-]", milestones),
		fmt.Sprintf("[gray]Techs[-] [white]%s[-]", textfmt.Int(s.Techs)),
		fmt.Sprintf("[gray]Soldiers[-] [white]%s[-]", textfmt.Int(s.Soldiers)),
	}, detailSep))

	// Line 4 — prestige + morale.
	lines = append(lines, strings.Join([]string{
		fmt.Sprintf("[gray]Prestige[-] [white]level %s[-] [gray](%s points)[-]", textfmt.Int(s.PrestigeLevel), textfmt.Int(s.PrestigeTotal)),
		fmt.Sprintf("[gray]Morale[-] [white]%.0f%%[-]", s.Morale*100),
	}, detailSep))

	// Line 4b — account attribution. Pre-account/legacy saves show a dash; saves
	// from the current account are tagged "(this account)", everything else
	// "(another account)". Only the short id prefix is shown — never the full id.
	switch {
	case s.AccountID == "":
		lines = append(lines, "[gray]Account:[-] [gray]none (pre-account save)[-]")
	case s.AccountID == currentAccountID && currentAccountID != "":
		lines = append(lines, fmt.Sprintf("[gray]Account:[-] [white]%s[-] [gray](this account)[-]", shortAccountID(s.AccountID)))
	default:
		lines = append(lines, fmt.Sprintf("[gray]Account:[-] [white]%s[-] [yellow](another account)[-]", shortAccountID(s.AccountID)))
	}

	// Line 5 — looming catastrophe warning (omitted entirely when none pending).
	if s.PendingCatastrophe != "" {
		lines = append(lines, fmt.Sprintf("[red]⚠ Pending: %s[-]", s.PendingCatastrophe))
	}

	// Line 6 — save metadata.
	// note: the raw tick count used to ride here; ticks mean nothing to a player.
	lines = append(lines, fmt.Sprintf(
		"[gray]Saved[-] %s",
		s.Timestamp.Format("Jan 2, 2006 3:04 PM"),
	))

	// Line 7 — lineage. Shown only for branched saves. A parent that is no longer
	// present (deleted/renamed) is marked "(detached)" rather than implying a link.
	if s.ParentName != "" && s.ParentName != s.Name {
		if parentPresent {
			lines = append(lines, fmt.Sprintf("[gray]Branched from: %s[-]", s.ParentName))
		} else {
			lines = append(lines, fmt.Sprintf("[gray]Branched from: %s (detached)[-]", s.ParentName))
		}
	}

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
// Falls back to the raw key, or "none" when empty.
func ageDisplay(key string) string {
	if key == "" {
		return "none"
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

// pluralSaves renders the save-count label ("1 save" / "N saves").
func pluralSaves(n int) string {
	if n == 1 {
		return "1 save"
	}
	return fmt.Sprintf("%d saves", n)
}

// savesDirLabel is the subtitle prefix. Saves live in the active account's
// slot (data/accounts/<id>/saves/), so the label names the account rather than
// a folder path that would be wrong for every account.
func savesDirLabel() string {
	return "This account"
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
