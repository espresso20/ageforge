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
	{theme.RoleText, "Text"},
	{theme.RoleAccent, "Accent"},
	{theme.RolePositive, "Positive"},
	{theme.RoleNegative, "Negative"},
	{theme.RoleHighlight, "Highlight"},
	{theme.RoleDim, "Dim"},
	{theme.RoleSelection, "Selection"},
}

// themePicker holds the widgets and live state for the theme picker screen. It is
// modeled on the load-game browser (load_game.go): a list on the left, a detail
// pane on the right, and live preview wired through SetChangedFunc.
type themePicker struct {
	app   *tview.Application
	pages *tview.Pages

	// engine is the bridge to the account layer (theming.md §6). It may be nil
	// (accountless play / tests); every account touch nil-guards both the engine
	// and engine.Account(). It also drives the unlock gate (themeAvailable) so
	// Phase-3 locked flavor themes can't be previewed/confirmed.
	engine *game.GameEngine

	root   *tview.Flex
	list   *tview.List
	detail *tview.TextView

	themes []theme.Theme // rows, 1:1 with list items (theme.All() order)

	// returnPage is the page to return to on confirm/cancel ("splash" from the
	// menu, "dashboard" when opened mid-game).
	returnPage string

	// originalKey is the theme that was active when the picker opened. Esc reverts
	// to it; Enter keeps whatever is currently previewed (it's already applied).
	originalKey string
}

// CreateThemePickerPage builds the full-screen theme picker and returns its root
// primitive. The caller registers it (pages.AddPage(themePickerPage, ...)) and
// focuses the returned primitive. The screen is self-contained: all key handling
// lives on the list.
//
// Live preview is apply-on-highlight: moving the selection calls theme.SetActive
// for real, so the whole UI (the picker plus whatever's behind it) retints to the
// highlighted theme instantly. We capture the originally-active theme on open so
// Esc can revert it; Enter keeps the previewed theme.
//
// returnPage names the page to return to on confirm (Enter) or cancel (Esc).
//
// engine bridges to the account layer for persistence + unlock gating (theming.md
// §6); it may be nil for accountless play/tests, in which case confirm simply
// doesn't persist and every theme in the always-available set stays selectable.
func CreateThemePickerPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, returnPage string) tview.Primitive {
	p := &themePicker{
		app:         app,
		pages:       pages,
		engine:      engine,
		returnPage:  returnPage,
		themes:      theme.All(),
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
		SetText("[gray]Preview applies live — pick a theme that reads well in your terminal.[-]")

	// ── Theme list ───────────────────────────────────────────────────────────
	p.list = tview.NewList()
	p.list.SetBorder(false)
	// Selection backs light text, so it uses the Selection role (like load_game),
	// not Accent — white-on-gold is unreadable (theme/themes_forge.go).
	theme.Track(func() {
		p.list.SetSelectedBackgroundColor(theme.Color(theme.RoleSelection)).
			SetSelectedTextColor(theme.Color(theme.RoleText))
	})
	p.list.ShowSecondaryText(false)
	for _, th := range p.themes {
		p.list.AddItem(themeRowLabel(th, p.originalKey, themeAvailable(p.account(), th)), "", 0, nil)
	}

	// ── Detail pane ──────────────────────────────────────────────────────────
	p.detail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	p.detail.SetBorder(true).
		SetTitle(" Details ")
	theme.Track(func() {
		p.detail.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent))
	})

	// ── Footer ───────────────────────────────────────────────────────────────
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(themeFooterBar())

	// ── Layout ───────────────────────────────────────────────────────────────
	p.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(p.list, 0, 1, true).    // weighted — takes remaining space
		AddItem(p.detail, 13, 0, false). // blurb + accessible note + 8 swatch rows + border
		AddItem(footer, 1, 0, false)

	// Live preview: highlighting a theme applies it for real and retints the whole
	// UI. SetChangedFunc fires on every interactive selection move, so it both
	// applies the theme AND forces a redraw so the player sees the retint at once.
	p.list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		p.applyAndDetail(index)
		if p.app != nil {
			// Retint the whole UI live — the picker is itself a sample of the running
			// theme (theming.md §7). QueueUpdateDraw is safe from the tview goroutine.
			p.app.QueueUpdateDraw(func() {})
		}
	})
	p.list.SetInputCapture(p.handleKey)

	// Start on the active theme's row so the picker opens on what's already in use
	// (no spurious preview-flicker to a different theme). We set this BEFORE the
	// detail seed so GetCurrentItem reflects the right row.
	p.list.SetCurrentItem(p.currentIndex())
	// Seed the detail pane (and re-apply the active theme, a no-op) without queuing
	// a redraw: at construction the event loop may not be running yet, and
	// AddPage/focus triggers the first Draw. Forcing QueueUpdateDraw here would
	// block on an un-run app (and is redundant when it is running).
	p.applyAndDetail(p.list.GetCurrentItem())

	return p.root
}

// currentIndex returns the list index of the originally-active theme, or 0.
func (p *themePicker) currentIndex() int {
	for i, th := range p.themes {
		if th.Key == p.originalKey {
			return i
		}
	}
	return 0
}

// applyAndDetail applies the theme at index for real (remap + restyle) and
// refreshes the detail pane. It does NOT queue a redraw — that's the caller's job,
// because the construction-time seed runs before the event loop and would block on
// QueueUpdateDraw, while the interactive SetChangedFunc path does want a redraw.
// Out-of-range is a no-op.
func (p *themePicker) applyAndDetail(index int) {
	if index < 0 || index >= len(p.themes) {
		return
	}
	th := p.themes[index]
	// SetActive sets the active theme + applies the name-remap + re-runs the
	// restyle registry; the caller owns any redraw.
	_ = theme.SetActive(th.Key)
	theme.Restyle()
	p.detail.SetText(themeDetailText(th, themeAvailable(p.account(), th)))
}

// handleKey routes the picker's action keys. tview.List handles ↑/↓ natively; we
// intercept Enter (confirm) and Esc/q (cancel/revert).
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

// account returns the picker's account, or nil when accountless. Centralizes the
// engine→account nil-guarding the persist/gate paths share.
func (p *themePicker) account() *game.Account {
	if p.engine == nil {
		return nil
	}
	return p.engine.Account()
}

// confirm keeps the currently-previewed theme and persists it account-wide
// (theming.md §6). The previewed theme is already applied process-locally via the
// live preview; here we make it durable.
//
// Unlock gate (theming.md §4/§5): if the highlighted theme isn't available to this
// account (a locked flavor theme — impossible today since all shipped themes are
// Accessible/Forge), confirming it would be wrong, so we revert to the open-time
// theme instead of keeping/persisting the locked preview. Preview-on-highlight is
// still allowed to show it; only Enter is gated.
func (p *themePicker) confirm() {
	idx := p.list.GetCurrentItem()
	if idx >= 0 && idx < len(p.themes) {
		th := p.themes[idx]
		if !themeAvailable(p.account(), th) {
			// Locked: don't keep the preview. Fall back to the open-time theme and
			// don't persist. (No theme is locked today; this is the Phase-3 path.)
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
		theme.Restyle()
		if p.app != nil {
			p.app.QueueUpdateDraw(func() {})
		}
	}
	p.close()
}

// close removes the picker page and returns to the page it was opened from.
func (p *themePicker) close() {
	p.pages.RemovePage(themePickerPage)
	p.pages.SwitchToPage(p.returnPage)
}

// ── Pure helpers (formatting) ───────────────────────────────────────────────

// themeRowLabel renders one list row: the theme name, a "(current)" marker on the
// theme that was active when the picker opened, an accessible tag if so, and a
// "locked" marker when the theme isn't available to this account. The current
// marker tracks the open-time active theme (not the live preview) so the row stays
// stable as the player arrows through previews.
//
// available routes through themeAvailable (theming.md §4/§5): locked flavor themes
// still LIST (and can preview on highlight) but are visually marked and refused on
// confirm. No theme is locked today, so this renders the plain row for all of them.
func themeRowLabel(t theme.Theme, activeKey string, available bool) string {
	label := t.Name
	if t.Key == activeKey {
		label = "[gold]● " + t.Name + " (current)[-]"
	}
	if t.Accessible {
		label += "   [cyan]accessible[-]"
	}
	if !available {
		label += "   [gray]🔒 locked[-]"
	}
	return label
}

// themeDetailText renders the detail pane for a theme: name, blurb, an accessible
// note (with the gain/loss glyphs so the redundant ± encoding is visible), and the
// palette swatch rows.
//
// available reports whether this theme is unlocked for the current account
// (themeAvailable). A LOCKED theme shows a "🔒 Locked — <UnlockHint>" line above the
// swatches (theming.md §7); the swatches still render as a preview of what the player
// will get, so the locked theme is enticing rather than blank.
func themeDetailText(t theme.Theme, available bool) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("[gold]%s[-]", t.Name))
	if t.Blurb != "" {
		lines = append(lines, "[gray]"+t.Blurb+"[-]")
	}
	if t.Accessible {
		note := "[cyan]Accessible[-] [gray]— colorblind-safe / high-contrast, always unlocked[-]"
		if t.GainGlyph != "" || t.LossGlyph != "" {
			// Show the signed glyphs so the shape-based ± encoding is visible in the
			// picker itself (theming.md §7).
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
		if hint == "" {
			hint = "unlock via a milestone"
		}
		lines = append(lines, fmt.Sprintf("[red]🔒 Locked[-] [gray]— %s[-]", hint))
	}
	lines = append(lines, "") // blank spacer before the swatch block
	lines = append(lines, themeSwatches(t))
	return strings.Join(lines, "\n")
}

// themeSwatches renders the theme's palette as one labeled colored-block row per
// role. Each block uses the theme's OWN literal color (not the active remap) via a
// hex tag, so every row in the list shows its true palette even though only the
// highlighted theme is the live-applied one.
func themeSwatches(t theme.Theme) string {
	rows := make([]string, 0, len(swatchRoles))
	for _, sr := range swatchRoles {
		hex := colorHexTag(t.Color(sr.role))
		// ███ block in the role's literal color, then a gray label.
		rows = append(rows, fmt.Sprintf("%s███[-]  [gray]%s[-]", hex, sr.label))
	}
	return strings.Join(rows, "\n")
}

// colorHexTag renders a tcell.Color as a tview "[#rrggbb]" inline tag from its
// literal RGB. Themes are built with NewRGBColor, so Hex() yields the true RGB;
// an invalid/default color (Hex() < 0) falls back to a neutral gray tag so the
// swatch never emits a malformed tag.
func colorHexTag(c tcell.Color) string {
	h := c.Hex()
	if h < 0 {
		return "[#808080]"
	}
	return fmt.Sprintf("[#%06x]", h)
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
