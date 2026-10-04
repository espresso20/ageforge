package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// themePickerPage is the page name under which the theme picker is registered on
// the root Pages. Added on demand (from the splash menu or a bare `theme`
// command) and removed on confirm/cancel so it's always rebuilt fresh.
const themePickerPage = "theme_picker"

// swatchRoles is the role set rendered as colored blocks in the detail pane, in
// display order. Selection backs a row rather than text, but we show it too so the
// player can see the highlight color a theme will use. Matches the §7 spec set.
var swatchRoles = []struct {
	role  theme.Role
	label string
}{
	{theme.RoleBackground, "Background"},
	{theme.RoleSurface, "Surface"},
	{theme.RoleText, "Text"},
	{theme.RoleAccent, "Accent"},
	{theme.RolePositive, "Positive"},
	{theme.RoleNegative, "Negative"},
	{theme.RoleHighlight, "Highlight"},
	{theme.RoleDim, "Dim"},
	{theme.RoleSelection, "Selection"},
}

// pickerRow is one list row: either a group heading (th == nil) or a theme.
type pickerRow struct {
	heading string
	th      *theme.Theme
}

// themePicker holds the widgets and live state for the theme picker screen. It is
// modeled on the load-game browser (load_game.go): a list on the left, detail and
// a live preview on the right, and live apply wired through SetChangedFunc.
type themePicker struct {
	app   *tview.Application
	pages *tview.Pages

	// engine is the bridge to the account layer (the theming design §6). It may be nil
	// (accountless play / tests); every account touch nil-guards both the engine
	// and engine.Account(). It also drives the unlock gate (themeAvailable) so
	// locked flavor themes can be previewed but not confirmed.
	engine *game.GameEngine

	root    *tview.Flex
	list    *tview.List
	detail  *tview.TextView
	preview *tview.Box

	rows      []pickerRow   // 1:1 with list items: group headings + themes
	themes    []theme.Theme // theme.All() order (kept for tests / lookups)
	lastIndex int           // previous list index, for skipping headings in the travel direction
	shown     theme.Theme   // theme the preview/detail currently describe

	// returnPage is the page to return to on confirm/cancel ("splash" from the
	// menu, "dashboard" when opened mid-game).
	returnPage string

	// originalKey is the theme that was active when the picker opened. Esc reverts
	// to it; Enter keeps whatever is currently previewed (it's already applied).
	originalKey string
}

// pickerRows groups themes under their picker section headings, in the order of
// theme.Groups, preserving theme.All() order within each group. Empty groups are
// omitted.
func pickerRows(all []theme.Theme) []pickerRow {
	var rows []pickerRow
	for _, g := range theme.Groups {
		first := true
		for i := range all {
			if all[i].Group() != g {
				continue
			}
			if first {
				rows = append(rows, pickerRow{heading: g.String()})
				first = false
			}
			th := all[i]
			rows = append(rows, pickerRow{th: &th})
		}
	}
	return rows
}

// CreateThemePickerPage builds the full-screen theme picker and returns its root
// primitive. The caller registers it (pages.AddPage(themePickerPage, ...)) and
// focuses the returned primitive. The screen is self-contained: all key handling
// lives on the list.
//
// Two previews, deliberately redundant:
//   - Live apply: moving the selection calls theme.SetActive for real, so the
//     whole UI behind the picker retints to the highlighted theme. Every theme
//     paints its own canvas (theme.WrapScreen), so Daylight looks like Daylight
//     even on a black terminal.
//   - A sample panel drawn with the candidate theme's OWN literal colors
//     (background, surface, border, text tiers, ±, selection, keycap, danger),
//     independent of the active theme, so the preview is self-contained.
//
// We capture the originally-active theme on open so Esc can revert it; Enter
// keeps the previewed theme (and persists it) if it is unlocked.
//
// engine bridges to the account layer for persistence + unlock gating (the theming design
// §6); it may be nil for accountless play/tests, in which case confirm simply
// doesn't persist and only the always-available themes are selectable.
func CreateThemePickerPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, returnPage string) tview.Primitive {
	all := theme.All()
	p := &themePicker{
		app:         app,
		pages:       pages,
		engine:      engine,
		returnPage:  returnPage,
		themes:      all,
		rows:        pickerRows(all),
		originalKey: theme.Active().Key,
	}

	// ── Title ────────────────────────────────────────────────────────────────
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gold]═══ Themes ═══[-]")

	subtitle := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]The preview is live. Pick a theme that reads well in your terminal; light or dark, every theme paints its own background.[-]")

	// ── Theme list ───────────────────────────────────────────────────────────
	p.list = tview.NewList()
	p.list.SetBorder(true).SetTitle(" Themes ")
	// Selection colors are late-bound Refs, so they follow the live preview with
	// no Track closure (theme/screen.go).
	p.list.SetSelectedBackgroundColor(theme.Ref(theme.RoleSelection)).
		SetSelectedTextColor(theme.Ref(theme.RoleSelectionText))
	p.list.ShowSecondaryText(false)
	for _, r := range p.rows {
		if r.th == nil {
			p.list.AddItem(themeGroupHeading(r.heading), "", 0, nil)
			continue
		}
		p.list.AddItem(themeRowLabel(*r.th, p.originalKey, themeAvailable(p.account(), *r.th)), "", 0, nil)
	}

	// ── Detail pane ──────────────────────────────────────────────────────────
	p.detail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	p.detail.SetBorder(true).
		SetTitle(" Details ")
	p.detail.SetBackgroundColor(theme.Ref(theme.RoleSurface))

	// ── Sample preview (candidate's own colors) ─────────────────────────────
	p.preview = tview.NewBox()
	p.preview.SetDrawFunc(func(screen tcell.Screen, x, y, w, h int) (int, int, int, int) {
		drawThemeSample(screen, x, y, w, h, p.shown)
		return x, y, w, h
	})

	// ── Footer ───────────────────────────────────────────────────────────────
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(themeFooterBar())

	// ── Layout ───────────────────────────────────────────────────────────────
	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p.detail, 0, 1, false).
		AddItem(p.preview, 9, 0, false) // sample panel: caption + bordered 7-row mock
	body := tview.NewFlex().
		AddItem(p.list, 0, 2, true).
		AddItem(right, 0, 3, false)
	p.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(footer, 1, 0, false)

	// Live preview: highlighting a theme applies it for real and retints the whole
	// UI. This callback runs on the tview MAIN goroutine, where tview redraws
	// automatically after the input event — so the live retint shows without an
	// explicit redraw. We must NOT call QueueUpdateDraw here: it blocks until the
	// main loop drains its update queue, but we ARE the main loop mid-callback, so
	// it would deadlock the entire app (frozen, no keys). This was the
	// freeze-on-arrow bug (Trello TCGiSWYX). Skipping a heading row uses
	// List.SetCurrentItem, which only re-enters this callback — no locks taken.
	p.list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		if index >= 0 && index < len(p.rows) && p.rows[index].th == nil {
			p.skipHeading(index)
			return
		}
		p.lastIndex = index
		p.applyAndDetail(index)
	})
	p.list.SetInputCapture(p.handleKey)

	// Start on the active theme's row so the picker opens on what's already in use
	// (no spurious preview-flicker to a different theme). We set this BEFORE the
	// detail seed so GetCurrentItem reflects the right row.
	start := p.currentIndex()
	p.lastIndex = start
	p.list.SetCurrentItem(start)
	// Seed the detail pane (and re-apply the active theme, a no-op) without queuing
	// a redraw: at construction the event loop may not be running yet, and
	// AddPage/focus triggers the first Draw. Forcing QueueUpdateDraw here would
	// block on an un-run app (and is redundant when it is running).
	p.applyAndDetail(p.list.GetCurrentItem())

	return p.root
}

// skipHeading moves the selection off a group heading at index, continuing in
// the direction of travel (and wrapping like tview.List does).
func (p *themePicker) skipHeading(index int) {
	n := len(p.rows)
	step := 1
	if index < p.lastIndex && !(p.lastIndex == n-1 && index == 0) {
		step = -1
	}
	next := index
	for i := 0; i < n; i++ {
		next = (next + step + n) % n
		if p.rows[next].th != nil {
			break
		}
	}
	p.list.SetCurrentItem(next)
}

// currentIndex returns the list index of the originally-active theme, or the
// first theme row.
func (p *themePicker) currentIndex() int {
	first := -1
	for i, r := range p.rows {
		if r.th == nil {
			continue
		}
		if first < 0 {
			first = i
		}
		if r.th.Key == p.originalKey {
			return i
		}
	}
	if first < 0 {
		return 0
	}
	return first
}

// themeAt returns the theme at a list index, or ok=false for a heading or an
// out-of-range index.
func (p *themePicker) themeAt(index int) (theme.Theme, bool) {
	if index < 0 || index >= len(p.rows) || p.rows[index].th == nil {
		return theme.Theme{}, false
	}
	return *p.rows[index].th, true
}

// applyAndDetail applies the theme at index for real (remap + restyle) and
// refreshes the detail pane and sample. It deliberately does NOT redraw: both
// callers run on the main goroutine (the construction-time seed before the loop,
// and the SetChangedFunc live preview during it), and tview redraws automatically
// after the input event — so an explicit QueueUpdateDraw here would be redundant
// at best and a whole-app deadlock at worst (Trello TCGiSWYX). Headings and
// out-of-range indices are a no-op.
func (p *themePicker) applyAndDetail(index int) {
	th, ok := p.themeAt(index)
	if !ok {
		return
	}
	// SetActive sets the active theme + applies the name-remap + re-runs the
	// restyle registry; the caller owns any redraw.
	_ = theme.SetActive(th.Key)
	p.shown = th
	p.detail.SetText(themeDetailText(th, themeAvailable(p.account(), th), p.state()))
}

// handleKey routes the picker's action keys. tview.List handles ↑/↓ natively; we
// intercept Enter (confirm) and Esc/q (cancel/revert). No lock acquisition here.
func (p *themePicker) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		p.confirm()
		return nil
	case tcell.KeyEsc:
		p.cancel()
		return nil
	case tcell.KeyRune:
		if r := event.Rune(); r == 'q' || r == 'Q' {
			p.cancel()
			return nil
		}
	}
	return event
}

// state is the game snapshot the unlock hints are worded for, or nil with no
// engine (tests).
func (p *themePicker) state() *game.GameState {
	if p.engine == nil {
		return nil
	}
	st := p.engine.GetState()
	return &st
}

// account returns the picker's account, or nil when accountless. Centralizes the
// engine→account nil-guarding the persist/gate paths share.
func (p *themePicker) account() *game.Account {
	if p.engine == nil {
		return nil
	}
	return p.engine.Account()
}

// confirm keeps the currently-previewed theme and persists it account-wide
// (the theming design §6). The previewed theme is already applied process-locally via the
// live preview; here we make it durable.
//
// Unlock gate (the theming design §4/§5): if the highlighted theme isn't available to this
// account (a locked flavor theme), confirming it would be wrong, so we revert to
// the open-time theme instead of keeping/persisting the locked preview.
// Preview-on-highlight is still allowed to show it; only Enter is gated.
func (p *themePicker) confirm() {
	if th, ok := p.themeAt(p.list.GetCurrentItem()); ok {
		if !themeAvailable(p.account(), th) {
			p.cancel()
			return
		}
		// Persist the chosen theme account-wide. Nil-guarded for accountless play;
		// the Save error is non-fatal (theme is a preference, not empire state) so we
		// keep the applied theme regardless and let the account layer log/return it.
		if acct := p.account(); acct != nil {
			_ = acct.SetActiveTheme(theme.Active().Key)
		}
	}
	p.close()
}

// cancel reverts to the theme that was active when the picker opened (undoing any
// live preview) and returns to the page the picker was opened from.
func (p *themePicker) cancel() {
	if theme.Active().Key != p.originalKey {
		_ = theme.SetActive(p.originalKey)
		// No QueueUpdateDraw: cancel() runs from the key handler on the main
		// goroutine, where QueueUpdateDraw deadlocks (see SetChangedFunc above /
		// Trello TCGiSWYX). close() switches pages and tview redraws after the event,
		// so the reverted theme paints automatically.
	}
	p.close()
}

// close removes the picker page and returns to the page it was opened from.
func (p *themePicker) close() {
	p.pages.RemovePage(themePickerPage)
	p.pages.SwitchToPage(p.returnPage)
}

// ── Pure helpers (formatting) ───────────────────────────────────────────────

// themeGroupHeading renders a non-selectable section heading row.
func themeGroupHeading(name string) string {
	return "[dim]── " + name + " ──[-]"
}

// themeRowLabel renders one list row: the theme name, a "(current)" marker on the
// theme that was active when the picker opened, a Light/Dark tag, an accessible
// tag if so, and a "locked" marker when the theme isn't available to this
// account. The current marker tracks the open-time active theme (not the live
// preview) so the row stays stable as the player arrows through previews.
//
// available routes through themeAvailable (the theming design §4/§5): locked flavor themes
// still LIST (and can preview on highlight) but are visually marked and refused on
// confirm.
func themeRowLabel(t theme.Theme, activeKey string, available bool) string {
	label := "  " + t.Name
	if t.Key == activeKey {
		label = "[gold]● " + t.Name + " (current)[-]"
	}
	label += "  [dim]" + strings.ToLower(t.Variant()) + "[-]"
	if t.Accessible {
		label += "  [cyan]accessible[-]"
	}
	if !available {
		label += "  [gray]🔒 locked[-]"
	}
	return label
}

// themeDetailText renders the detail pane for a theme: name, Light/Dark and
// group, blurb, an accessible note (with the gain/loss glyphs so the redundant ±
// encoding is visible), the unlock condition when locked, and the palette swatch
// rows.
//
// available reports whether this theme is unlocked for the current account
// (themeAvailable). A LOCKED theme shows a "🔒 Locked — <UnlockHint>" line above the
// swatches (the theming design §7); the swatches and sample still render as a preview of
// what the player will get, so the locked theme is enticing rather than blank. st,
// when set, words the hint for that game (themeUnlockHint: no age the player
// cannot see yet); nil shows it as written.
func themeDetailText(t theme.Theme, available bool, st *game.GameState) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("[gold]%s[-]  [dim]%s theme · %s[-]", t.Name, t.Variant(), t.Group()))
	if t.Blurb != "" {
		lines = append(lines, "[gray]"+t.Blurb+"[-]")
	}
	if t.Accessible {
		note := "[cyan]Accessible:[-] [gray]colorblind-safe or high-contrast, always unlocked[-]"
		if t.GainGlyph != "" || t.LossGlyph != "" {
			// Show the signed glyphs so the shape-based ± encoding is visible in the
			// picker itself (the theming design §7).
			note += fmt.Sprintf(" [gray](gain %s / loss %s)[-]", t.GainGlyph, t.LossGlyph)
		}
		lines = append(lines, note)
	}
	if !available {
		// Locked flavor theme: surface the unlock condition. Fall back to a generic
		// line if the theme somehow has no hint (the registry-consistency test makes
		// that impossible for gated themes, but the detail pane shouldn't render an
		// empty "Locked —" tail if it ever happens).
		hint := t.UnlockHint
		if st != nil {
			hint = themeUnlockHint(t, *st) // never an age the player cannot see yet
		}
		if hint == "" {
			hint = "unlocked by a milestone"
		}
		lines = append(lines, fmt.Sprintf("[red]🔒 Locked:[-] [gray]%s[-]", hint))
	}
	lines = append(lines, "") // blank spacer before the swatch block
	lines = append(lines, themeSwatches(t))
	return strings.Join(lines, "\n")
}

// themeSwatches renders the theme's palette as one labeled colored-block row per
// role. Each block uses the theme's OWN literal color (not the active remap) via a
// hex tag, so the swatches always show the true palette. The block sits on a rim
// color mixed toward black/white, so a swatch that matches the pane behind it
// (Background, Surface) still reads as a framed chip rather than vanishing.
func themeSwatches(t theme.Theme) string {
	rows := make([]string, 0, len(swatchRoles))
	for _, sr := range swatchRoles {
		c := t.Color(sr.role)
		rim := theme.Mix(c, theme.BestOn(c), 0.45)
		rows = append(rows, fmt.Sprintf("%s▐███▌[-:-]  [gray]%s[-]", theme.HexTagFgBg(c, rim), sr.label))
	}
	return strings.Join(rows, "\n")
}

// colorHexTag renders a tcell.Color as a tview "[#rrggbb]" inline tag from its
// literal RGB (theme.HexTag). Kept as the picker's name for it.
func colorHexTag(c tcell.Color) string {
	return theme.HexTag(c)
}

// themeFooterBar is the picker action bar: keycap buttons matching the load-game
// footer idiom (footerButton lives in load_game.go).
func themeFooterBar() string {
	return "  " + strings.Join([]string{
		footerButton("↑↓", "Preview"),
		footerButton("Enter", "Keep"),
		footerButton("Esc", "Cancel"),
	}, "  ") + "  "
}

// drawThemeSample paints a miniature of the game UI in th's own colors: the
// candidate's canvas, a Surface panel with its Border and Accent title, the text
// tiers, ± deltas, a selected row, a keycap and a danger chip. Every color is the
// candidate's literal RGB (never a Ref), so the sample is correct regardless of
// which theme is active.
func drawThemeSample(screen tcell.Screen, x, y, w, h int, th theme.Theme) {
	if w <= 0 || h <= 0 || th.Key == "" {
		return
	}
	col := th.Color
	fill := func(x0, y0, x1, y1 int, bg tcell.Color) {
		st := tcell.StyleDefault.Background(bg).Foreground(col(theme.RoleText))
		for yy := y0; yy < y1; yy++ {
			for xx := x0; xx < x1; xx++ {
				screen.SetContent(xx, yy, ' ', nil, st)
			}
		}
	}
	put := func(xx, yy int, s string, fg, bg tcell.Color, bold bool) int {
		st := tcell.StyleDefault.Foreground(fg).Background(bg).Bold(bold)
		for _, r := range s {
			if xx >= x+w-1 {
				break
			}
			screen.SetContent(xx, yy, r, nil, st)
			xx++
		}
		return xx
	}

	// Canvas.
	bg := col(theme.RoleBackground)
	fill(x, y, x+w, y+h, bg)
	put(x+1, y, "Sample: "+th.Name+" ("+th.Variant()+")", col(theme.RoleDim), bg, false)

	// Surface panel with a border.
	px0, py0, px1, py1 := x+1, y+1, x+w-1, y+h
	if px1-px0 < 12 || py1-py0 < 4 {
		return
	}
	surf := col(theme.RoleSurface)
	fill(px0, py0, px1, py1, surf)
	bst := tcell.StyleDefault.Foreground(col(theme.RoleBorder)).Background(surf)
	for xx := px0; xx < px1; xx++ {
		screen.SetContent(xx, py0, '─', nil, bst)
		screen.SetContent(xx, py1-1, '─', nil, bst)
	}
	for yy := py0; yy < py1; yy++ {
		screen.SetContent(px0, yy, '│', nil, bst)
		screen.SetContent(px1-1, yy, '│', nil, bst)
	}
	screen.SetContent(px0, py0, '┌', nil, bst)
	screen.SetContent(px1-1, py0, '┐', nil, bst)
	screen.SetContent(px0, py1-1, '└', nil, bst)
	screen.SetContent(px1-1, py1-1, '┘', nil, bst)
	put(px0+2, py0, " Resources ", col(theme.RoleAccent), surf, true)

	gain, loss := "+", "-"
	if th.GainGlyph != "" {
		gain = th.GainGlyph
	}
	if th.LossGlyph != "" {
		loss = th.LossGlyph
	}
	lines := []func(yy int){
		func(yy int) {
			xx := put(px0+2, yy, "Food ", col(theme.RoleLabel), surf, false)
			xx = put(xx, yy, "1.2K ", col(theme.RoleHighlight), surf, false)
			put(xx, yy, gain+"12.5/s", col(theme.RolePositive), surf, false)
		},
		func(yy int) {
			xx := put(px0+2, yy, "Wood ", col(theme.RoleLabel), surf, false)
			xx = put(xx, yy, "340 ", col(theme.RoleHighlight), surf, false)
			put(xx, yy, loss+"3.1/s", col(theme.RoleNegative), surf, false)
		},
		func(yy int) {
			xx := put(px0+2, yy, "Body text ", col(theme.RoleText), surf, false)
			put(xx, yy, "and a dim hint", col(theme.RoleDim), surf, false)
		},
		func(yy int) {
			sel := col(theme.RoleSelection)
			xx := put(px0+2, yy, " ▸ selected row ", col(theme.RoleSelectionText), sel, false)
			_ = xx
		},
		func(yy int) {
			xx := put(px0+2, yy, " Enter ", col(theme.RoleOnAccent), col(theme.RoleAccent), true)
			xx = put(xx, yy, " Keep ", col(theme.RoleText), col(theme.RoleChip), true)
			xx = put(xx, yy, "  ", col(theme.RoleText), surf, false)
			put(xx, yy, " danger ", col(theme.RoleOnNegative), col(theme.RoleNegative), true)
		},
	}
	for i, ln := range lines {
		yy := py0 + 1 + i
		if yy >= py1-1 {
			break
		}
		ln(yy)
	}
}
