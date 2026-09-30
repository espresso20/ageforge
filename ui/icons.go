package ui

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/nerdfont"
	"github.com/espresso20/ageforge/theme"
)

// icons.go is the guided icons check (the icons command), in a window of its
// own over the dashboard. The map's nerd glyph tier needs a Nerd Font, which
// a terminal game cannot ship: the player's font draws every cell. The window
// walks through it:
//
//  1. Sample icons next to their Unicode fallbacks, and two buttons: I see
//     icons (sets map glyphs nerd) and I see boxes.
//  2. On boxes, the offer to install JetBrains Mono Nerd Font for the user.
//  3. The install: pkg/nerdfont downloads the pinned release, checks its
//     SHA-256 and copies the Mono TTFs, off the UI goroutine, with its
//     progress in the window. The window can close meanwhile; icons opens it
//     again, and the log says when the install is done.
//  4. The steps to pick the font in the player's terminal, detected from the
//     environment.
//  5. Restart the terminal, then run icons again to check.
//
// The window owns the keyboard while it is open: each button's letter
// presses it, Tab and the arrows move between buttons, Enter presses the
// focused one and Esc closes the window at any step.

// nerdFontsPage is where a player gets the font by hand.
const nerdFontsPage = "https://www.nerdfonts.com/font-downloads"

// iconsPage is the tview page name of the icons window.
const iconsPage = "icons"

type iconsStage int

const (
	iconsAskSee     iconsStage = iota // the samples: icons or boxes?
	iconsHaveIcons                    // answered icons: map glyphs nerd
	iconsAskInstall                   // the install offer
	iconsManual                       // declined: how to install by hand
	iconsInstalling                   // the download runs off the UI goroutine
	iconsInstalled                    // done: pick the font, then restart
	iconsFailed                       // the install failed; nothing changed
)

// maxIconsProgress is how many progress lines the window keeps.
const maxIconsProgress = 6

// fontInstaller is what the check needs from pkg/nerdfont (tests fake it).
type fontInstaller interface {
	Install(ctx context.Context) (nerdfont.Result, error)
}

// iconsFlow is the check's state. It lives on the UI goroutine; the install
// runs on its own goroutine and reports back through post.
type iconsFlow struct {
	stage iconsStage
	// open is set while the window shows. The install goes on without it.
	open bool
	// unseen: the install finished while the window was closed, so the next
	// icons shows its outcome instead of starting a new check.
	unseen   bool
	progress []string
	result   nerdfont.Result
	// err is the install's error (iconsFailed) or the glyph setting's
	// (iconsHaveIcons).
	err error

	log     func(kind, msg string)
	setNerd func() error
	// installer returns an installer whose progress lines go to progress
	// (called off the UI goroutine).
	installer func(progress func(string)) fontInstaller
	environ   func() []string
	goos      string
	// run starts the install off the UI goroutine; post runs a func back on
	// it. Tests control both.
	run  func(func())
	post func(func())
	// changed redraws the window after the flow moves (UI goroutine).
	changed func()
}

// newIconsFlow wires the check for the real environment. log must be safe
// on the UI goroutine (the engine log is).
func newIconsFlow(log func(kind, msg string), setNerd func() error, post func(func())) *iconsFlow {
	return &iconsFlow{
		log:     log,
		setNerd: setNerd,
		installer: func(progress func(string)) fontInstaller {
			in := nerdfont.New()
			in.Progress = progress
			return in
		},
		environ: os.Environ,
		goos:    runtime.GOOS,
		run:     func(f func()) { go f() },
		post:    post,
	}
}

func (f *iconsFlow) notify() {
	if f.changed != nil {
		f.changed()
	}
}

// start opens the window: on the running install, on an outcome the player
// has not seen yet, or on a fresh check.
func (f *iconsFlow) start() {
	switch {
	case f.stage == iconsInstalling:
	case f.unseen:
		f.unseen = false
	default:
		f.stage, f.err, f.progress = iconsAskSee, nil, nil
	}
	f.open = true
	f.notify()
}

// close hides the window. A running install carries on.
func (f *iconsFlow) close() {
	f.open = false
	f.notify()
}

// seeIcons is the I see icons answer: the map draws Nerd Font icons.
func (f *iconsFlow) seeIcons() {
	f.stage = iconsHaveIcons
	f.err = f.setNerd()
	if f.err == nil {
		f.log(game.LogRoutine, "Map glyphs set to nerd: your font shows Nerd Font icons.")
	}
	f.notify()
}

// seeBoxes is the I see boxes answer: offer the install.
func (f *iconsFlow) seeBoxes() {
	f.stage = iconsAskInstall
	f.notify()
}

// decline turns the install down: the steps to do it by hand.
func (f *iconsFlow) decline() {
	f.stage = iconsManual
	f.notify()
}

// install runs the installer off the UI goroutine.
func (f *iconsFlow) install() {
	if f.stage == iconsInstalling {
		return
	}
	f.stage, f.err, f.progress = iconsInstalling, nil, nil
	f.notify()
	in := f.installer(func(s string) { f.post(func() { f.addProgress(s) }) })
	f.run(func() {
		res, err := in.Install(context.Background())
		f.post(func() { f.installed(res, err) })
	})
}

// addProgress shows one progress line (UI goroutine).
func (f *iconsFlow) addProgress(s string) {
	f.progress = append(f.progress, s)
	if len(f.progress) > maxIconsProgress {
		f.progress = f.progress[len(f.progress)-maxIconsProgress:]
	}
	f.notify()
}

// installed takes the install's outcome (UI goroutine). With the window
// closed the log says so, and the next icons shows it.
func (f *iconsFlow) installed(res nerdfont.Result, err error) {
	f.result, f.err = res, err
	f.stage = iconsInstalled
	if err != nil {
		f.stage = iconsFailed
	}
	switch {
	case f.open && err == nil:
		f.log(game.LogRoutine, "Installed JetBrains Mono Nerd Font "+nerdfont.Version+".")
	case f.open:
		f.log(game.LogRoutine, "The Nerd Font install failed: "+err.Error()+".")
	case err == nil:
		f.unseen = true
		f.log("success", "JetBrains Mono Nerd Font is installed. Type icons for the steps to pick it in your terminal.")
	default:
		f.unseen = true
		f.log("error", "The Nerd Font install failed: "+err.Error()+". Type icons for what to do next.")
	}
	f.notify()
}

// iconsButton is one of the window's buttons.
type iconsButton struct {
	key   rune // the letter that presses it; 0 for Close, which Esc presses
	label string
	fill  theme.Role
	press func()
}

// caption is the button's text: "[I] I see icons", "[Esc] Close".
func (b iconsButton) caption() string {
	k := "Esc"
	if b.key != 0 {
		k = strings.ToUpper(string(b.key))
	}
	return tview.Escape("[" + k + "] " + b.label)
}

// view is the window's text and buttons at the current stage.
func (f *iconsFlow) view() (string, []iconsButton) {
	closeBtn := iconsButton{label: "Close", fill: theme.RoleChip, press: f.close}
	para := func(sb *strings.Builder, role theme.Role, s string) {
		sb.WriteString(theme.Paint(role, tview.Escape(s)) + "\n")
	}
	var sb strings.Builder
	switch f.stage {
	case iconsAskSee:
		nerd, uni := iconSamples()
		para(&sb, theme.RoleText, "The map can draw its symbols as Nerd Font icons, if your font has them.")
		sb.WriteString("\n")
		sb.WriteString("  " + theme.Paint(theme.RoleLabel, "Nerd Font icons    ") + theme.Paint(theme.RoleBright, nerd) + "\n")
		sb.WriteString("  " + theme.Paint(theme.RoleLabel, "Unicode fallbacks  ") + theme.Paint(theme.RoleBright, uni) + "\n\n")
		para(&sb, theme.RoleText, "Do the icons on the first line look like symbols, or like boxes and question marks?")
		return sb.String(), []iconsButton{
			{key: 'i', label: "I see icons", fill: theme.RoleAccent, press: f.seeIcons},
			{key: 'b', label: "I see boxes", fill: theme.RoleAccent, press: f.seeBoxes},
			closeBtn,
		}
	case iconsHaveIcons:
		if f.err != nil {
			para(&sb, theme.RoleWarning, "Your font shows the icons, but the setting could not be saved: "+f.err.Error()+". Type map glyphs nerd to try again.")
		} else {
			para(&sb, theme.RolePositive, "Your font shows Nerd Font icons, so the map now draws them (map glyphs nerd).")
		}
		return sb.String(), []iconsButton{closeBtn}
	case iconsAskInstall:
		para(&sb, theme.RoleText, "Your font has no Nerd Font icons, so the map keeps its Unicode symbols.")
		sb.WriteString("\n")
		para(&sb, theme.RoleText, "Install JetBrains Mono Nerd Font "+nerdfont.Version+" for your user? It downloads the official release from GitHub (about 130 MB), checks it, and copies the fonts into your user font folder. It needs no admin rights and changes nothing else.")
		return sb.String(), []iconsButton{
			{key: 'y', label: "Install", fill: theme.RoleAccent, press: f.install},
			{key: 'n', label: "No, show me how", fill: theme.RoleChip, press: f.decline},
			closeBtn,
		}
	case iconsManual:
		para(&sb, theme.RoleText, "Nothing installed. To install a Nerd Font yourself, get JetBrainsMono from "+nerdFontsPage+", then pick it in your terminal.")
		sb.WriteString("\n" + f.stepsText() + "\n")
		para(&sb, theme.RoleText, "Then restart your terminal and type icons again to check.")
		return sb.String(), []iconsButton{closeBtn}
	case iconsInstalling:
		para(&sb, theme.RoleText, "Installing JetBrains Mono Nerd Font "+nerdfont.Version+" for your user. You can close this window and keep playing: type icons to look in again.")
		sb.WriteString("\n")
		for _, p := range f.progress {
			sb.WriteString("  " + theme.Paint(theme.RoleDim, tview.Escape(p)) + "\n")
		}
		if len(f.progress) == 0 {
			sb.WriteString("  " + theme.Paint(theme.RoleDim, "Starting...") + "\n")
		}
		return sb.String(), []iconsButton{closeBtn}
	case iconsInstalled:
		para(&sb, theme.RolePositive, "JetBrains Mono Nerd Font is installed.")
		if f.result.ManualStep != "" {
			para(&sb, theme.RoleWarning, "One step is left for you: "+f.result.ManualStep)
		}
		para(&sb, theme.RoleText, "Now pick it in your terminal.")
		sb.WriteString("\n" + f.stepsText() + "\n")
		para(&sb, theme.RoleBright, "Restart your terminal, then type icons again to check.")
		return sb.String(), []iconsButton{closeBtn}
	case iconsFailed:
		msg := "The install failed."
		if f.err != nil {
			msg = "The font was not installed: " + f.err.Error() + "."
		}
		para(&sb, theme.RoleNegative, msg+" Nothing else was changed.")
		sb.WriteString("\n")
		para(&sb, theme.RoleText, "You can try again, or get it by hand from "+nerdFontsPage+".")
		return sb.String(), []iconsButton{
			{key: 'r', label: "Try again", fill: theme.RoleAccent, press: f.install},
			closeBtn,
		}
	}
	return "", []iconsButton{closeBtn}
}

// stepsText is how to pick the font in the detected terminal.
func (f *iconsFlow) stepsText() string {
	name, steps := terminalFontSteps(f.environ(), f.goos)
	var sb strings.Builder
	sb.WriteString(theme.Paint(theme.RoleAccent, tview.Escape("In "+name+":")) + "\n")
	for i, s := range steps {
		sb.WriteString("  " + strconv.Itoa(i+1) + ". " + tview.Escape(s) + "\n")
	}
	return sb.String()
}

// iconSamples is a few map symbols as Nerd Font icons and as their
// Unicode fallbacks.
func iconSamples() (nerd, uni string) {
	var n, u []string
	seen := map[rune]bool{}
	for _, g := range mapmodel.AllGlyphs() {
		if g.Nerd == 0 || seen[g.Nerd] || len(n) == 8 {
			continue
		}
		seen[g.Nerd] = true
		n = append(n, string(g.Nerd))
		u = append(u, string(g.Unicode))
	}
	return strings.Join(n, " "), strings.Join(u, " ")
}

// terminalFontSteps names the terminal the game runs in, from its
// environment, and gives the steps to choose the Nerd Font there. It reads
// the environment only; it never opens a terminal's settings file.
func terminalFontSteps(environ []string, goos string) (name string, steps []string) {
	env := map[string]string{}
	alacritty := false
	for _, kv := range environ {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
		if strings.HasPrefix(k, "ALACRITTY_") {
			alacritty = true
		}
	}
	font := nerdfont.FamilyName
	switch {
	case env["WT_SESSION"] != "":
		return "Windows Terminal", []string{
			"Open Settings (Ctrl+,) and pick your profile under Profiles.",
			"Go to Appearance and set Font face to " + font + ".",
			"Click Save.",
		}
	case env["TERM_PROGRAM"] == "vscode":
		return "the VS Code terminal", []string{
			"Open Settings (Ctrl+, or Cmd+,) and search for terminal.integrated.fontFamily.",
			"Set it to '" + font + "'.",
		}
	case env["TERM_PROGRAM"] == "iTerm.app":
		return "iTerm2", []string{
			"Open iTerm2 > Settings > Profiles and pick your profile.",
			"On the Text tab, set Font to " + font + ".",
			"If Use a different font for non-ASCII text is ticked, set that font too.",
		}
	case env["TERM_PROGRAM"] == "Apple_Terminal":
		return "Terminal", []string{
			"Open Terminal > Settings > Profiles and pick your profile.",
			"On the Text tab, click Change under Font and choose " + font + ".",
		}
	case env["TERM_PROGRAM"] == "WezTerm":
		return "WezTerm", []string{
			"Open your WezTerm config (~/.wezterm.lua).",
			"Set config.font = wezterm.font('" + font + "') and save; WezTerm reloads it.",
		}
	case env["KITTY_WINDOW_ID"] != "":
		return "kitty", []string{
			"Open kitty.conf (Ctrl+Shift+F2, or Cmd+, on macOS).",
			"Set font_family " + font + " and save.",
			"Reload the config (Ctrl+Shift+F5, or Ctrl+Cmd+, on macOS).",
		}
	case alacritty:
		return "Alacritty", []string{
			"Open your Alacritty config (alacritty.toml).",
			"Under font.normal, set family = \"" + font + "\" and save; Alacritty reloads it.",
		}
	case env["KONSOLE_VERSION"] != "":
		return "Konsole", []string{
			"Open Settings > Edit Current Profile > Appearance.",
			"Click Choose next to the font and pick " + font + ".",
			"Click OK.",
		}
	case env["GNOME_TERMINAL_SCREEN"] != "" || env["GNOME_TERMINAL_SERVICE"] != "" || env["VTE_VERSION"] != "":
		return "GNOME Terminal", []string{
			"Open the menu > Preferences and pick your profile.",
			"Tick Custom font and choose " + font + ".",
		}
	}
	if goos == "windows" {
		return "your terminal", []string{
			"Open the terminal's settings (in Windows Terminal: Ctrl+,).",
			"Set the font of your profile to " + font + ".",
		}
	}
	return "your terminal", []string{
		"Open your terminal's preferences or profile settings.",
		"Find the font setting and choose " + font + ".",
		"If your terminal has a separate font for symbols or non-ASCII text, set that too.",
	}
}

// --- The window ---------------------------------------------------------------

// The window's size: it fits an 80x24 terminal.
const (
	iconsWindowWidth  = 72
	iconsWindowHeight = 20
)

// iconsWindow shows the flow in a box floating over the dashboard. Built on
// first open; its text and buttons are replaced each time the flow moves.
type iconsWindow struct {
	root    *tview.Flex // the page: the box, centered
	body    *tview.TextView
	btnRow  *tview.Flex
	buttons []*tview.Button
	keys    []iconsButton
	focus   int
}

// newIconsWindow builds the window. It reads theme colors at construction,
// like the other modals; it is rebuilt with each dashboard.
func (d *Dashboard) newIconsWindow() *iconsWindow {
	w := &iconsWindow{}
	w.body = tview.NewTextView().SetDynamicColors(true).SetWrap(true).SetWordWrap(true)
	w.body.SetTextColor(theme.Color(theme.RoleText)).SetBackgroundColor(theme.Color(theme.RoleSurface))
	w.btnRow = tview.NewFlex().SetDirection(tview.FlexColumn)
	hint := surfaceText(theme.Paint(theme.RoleDim, "Tab moves between buttons · Enter presses · Esc closes"), tview.AlignCenter)
	// A Flex paints nothing of its own, so the margins beside the text are
	// surface boxes: every cell inside the border belongs to something that
	// clears it, or the dashboard shows through.
	bodyRow := tview.NewFlex().
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(w.body, 0, 1, false).
		AddItem(surfaceBox(), 1, 0, false)
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bodyRow, 0, 1, false).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(w.btnRow, 1, 0, true).
		AddItem(hint, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(" Icons check ").
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent)).
		SetBackgroundColor(theme.Color(theme.RoleSurface))
	w.root = floatingModal(inner, iconsWindowWidth, iconsWindowHeight)
	w.root.SetInputCapture(d.iconsKey)
	return w
}

// setFocus focuses button i (wrapping around).
func (w *iconsWindow) setFocus(app *tview.Application, i int) {
	if len(w.buttons) == 0 {
		return
	}
	w.focus = (i%len(w.buttons) + len(w.buttons)) % len(w.buttons)
	app.SetFocus(w.buttons[w.focus])
}

// startIcons opens the window (the icons command).
func (d *Dashboard) startIcons() { d.icons.start() }

// refreshIconsWindow shows the flow's current stage, or removes the window
// once the flow is closed (UI goroutine: the flow's changed hook).
func (d *Dashboard) refreshIconsWindow() {
	if !d.icons.open {
		if d.pages.HasPage(iconsPage) {
			d.pages.RemovePage(iconsPage)
			d.returnFocus()
		}
		return
	}
	if d.iconsWin == nil {
		d.iconsWin = d.newIconsWindow()
	}
	w := d.iconsWin
	body, keys := d.icons.view()
	w.body.SetText(body).ScrollToBeginning()
	w.keys = keys
	w.buttons = w.buttons[:0]
	w.btnRow.Clear()
	w.btnRow.AddItem(surfaceBox(), 0, 1, false)
	for i, k := range keys {
		b := tview.NewButton(k.caption()).SetSelectedFunc(k.press)
		styleFilledButton(b, k.fill)
		if i > 0 {
			w.btnRow.AddItem(surfaceBox(), 2, 0, false)
		}
		w.btnRow.AddItem(b, tview.TaggedStringWidth(k.caption())+4, 0, i == 0)
		w.buttons = append(w.buttons, b)
	}
	w.btnRow.AddItem(surfaceBox(), 0, 1, false)
	if !d.pages.HasPage(iconsPage) {
		d.pages.AddPage(iconsPage, w.root, true, true)
	}
	// Take the keyboard only when nothing sits over the window: a splash
	// that opened on top keeps it until it closes, and returnFocus then
	// hands it back.
	if front, _ := d.pages.GetFrontPage(); front == iconsPage {
		w.setFocus(d.app, 0)
	}
}

// iconsKey is the window's key handling: Esc closes, Tab and the arrows move
// between buttons, a button's letter presses it. Other letters do nothing;
// Enter reaches the focused button.
func (d *Dashboard) iconsKey(event *tcell.EventKey) *tcell.EventKey {
	w := d.iconsWin
	switch event.Key() {
	case tcell.KeyEsc:
		d.icons.close()
		return nil
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyDown:
		w.setFocus(d.app, w.focus+1)
		return nil
	case tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyUp:
		w.setFocus(d.app, w.focus-1)
		return nil
	case tcell.KeyRune:
		r := unicode.ToLower(event.Rune())
		for _, k := range w.keys {
			if k.key != 0 && k.key == r {
				k.press()
				return nil
			}
		}
		return nil
	}
	return event
}

// returnFocus gives the keyboard back when a window over the dashboard
// closes: to the icons window if it is still open, else to the open panel,
// else to the prompt. Without it a splash closing over the icons window
// would leave the prompt focused under it, where Esc leaves the game.
func (d *Dashboard) returnFocus() {
	switch {
	case d.icons != nil && d.icons.open && d.iconsWin != nil && d.pages.HasPage(iconsPage):
		d.pages.SendToFront(iconsPage)
		d.iconsWin.setFocus(d.app, d.iconsWin.focus)
	case d.overlayMgr != nil && d.overlayMgr.Focus():
	default:
		d.app.SetFocus(d.inputField)
	}
}

// setMapGlyphsNerd is the check's I see icons answer: map glyphs nerd, saved
// to the account (or the session, with no account).
func (d *Dashboard) setMapGlyphsNerd() error {
	s := d.mapSettings()
	s.Tier = mapmodel.TierNerd
	if d.engine != nil {
		if acct := d.engine.Account(); acct != nil {
			return acct.SetMapGlyphs(s.Tier.String())
		}
	}
	d.mapLocal = &s
	return nil
}
