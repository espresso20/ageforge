package ui

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/nerdfont"
)

// icons.go is the guided icons check (the icons command). The map's nerd
// glyph tier needs a Nerd Font, which a terminal game cannot ship: the
// player's font draws every cell. So the check asks, in the log:
//
//  1. Sample icons next to their Unicode fallbacks: "Do you see icons, or
//     boxes/question marks? (icons/boxes)". icons sets map glyphs nerd.
//  2. On boxes: "Install JetBrains Mono Nerd Font for your user? (y/n)".
//     y downloads the pinned release, checks its SHA-256 and installs the
//     Mono TTFs per user (pkg/nerdfont), off the UI goroutine, with
//     progress in the log.
//  3. The exact steps to pick the font in the player's terminal, detected
//     from the environment.
//  4. "Restart your terminal, then run icons again to check."
//
// The player answers at the prompt: while a question is open, the next
// line is read as its answer. Anything else closes the check and runs as
// a command.

// nerdFontsPage is where a player gets the font by hand.
const nerdFontsPage = "https://www.nerdfonts.com/font-downloads"

type iconsStage int

const (
	iconsIdle       iconsStage = iota
	iconsAskSee                // "(icons/boxes)"
	iconsAskInstall            // "(y/n)"
	iconsInstalling            // the download runs; answers are not read
)

// fontInstaller is what the check needs from pkg/nerdfont (tests fake it).
type fontInstaller interface {
	Install(ctx context.Context) (nerdfont.Result, error)
}

// iconsFlow is the check's state. It lives on the UI goroutine; the
// install runs on its own goroutine and reports back through post.
type iconsFlow struct {
	stage iconsStage

	log     func(kind, msg string)
	setNerd func() error
	// installer returns an installer whose progress lines go to progress.
	installer func(progress func(string)) fontInstaller
	environ   func() []string
	goos      string
	// run starts the install off the UI goroutine; post runs a func back
	// on it. Tests run both inline.
	run  func(func())
	post func(func())
}

// newIconsFlow wires the check for the real environment. log must be safe
// off the UI goroutine (the engine log is).
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

// waiting reports whether the next prompt line is an answer.
func (f *iconsFlow) waiting() bool { return f.stage == iconsAskSee || f.stage == iconsAskInstall }

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

const (
	iconsQuestion  = "Do you see icons, or boxes/question marks? (icons/boxes)"
	installPrompt  = "Install JetBrains Mono Nerd Font for your user? (y/n)"
	restartMessage = "Restart your terminal, then run icons again to check."
)

// start asks the first question.
func (f *iconsFlow) start() {
	if f.stage == iconsInstalling {
		f.log("info", "The font install is still running. Its progress is in this log.")
		return
	}
	nerd, uni := iconSamples()
	f.log("info", "Nerd Font icons:   "+nerd)
	f.log("info", "Unicode fallbacks: "+uni)
	f.log("info", iconsQuestion)
	f.stage = iconsAskSee
}

// answer reads a prompt line while a question is open. It reports whether
// the line was an answer; if not, the check is closed and the caller runs
// the line as a command.
func (f *iconsFlow) answer(line string) bool {
	a := strings.ToLower(strings.TrimSpace(line))
	switch f.stage {
	case iconsAskSee:
		switch a {
		case "":
			f.log("info", iconsQuestion)
			return true
		case "icons", "icon", "i":
			f.stage = iconsIdle
			if err := f.setNerd(); err != nil {
				f.log("warning", "Your font shows the icons, but the setting could not be saved: "+err.Error()+". Type map glyphs nerd to try again.")
				return true
			}
			f.log("success", "Your font shows Nerd Font icons, so the map now draws them (map glyphs nerd).")
			return true
		case "boxes", "box", "b", "question marks":
			f.stage = iconsAskInstall
			f.log("info", installPrompt)
			return true
		}
	case iconsAskInstall:
		switch a {
		case "":
			f.log("info", installPrompt)
			return true
		case "y", "yes":
			f.install()
			return true
		case "n", "no":
			f.stage = iconsIdle
			f.log("info", "Nothing installed. To install a Nerd Font yourself, get JetBrainsMono from "+nerdFontsPage+", then pick it in your terminal:")
			f.terminalSteps()
			return true
		}
	default:
		return false
	}
	f.stage = iconsIdle
	f.log("info", "Icons check closed. Type icons to start it again.")
	return false
}

// install runs the installer off the UI goroutine.
func (f *iconsFlow) install() {
	f.stage = iconsInstalling
	f.log("info", "Installing JetBrains Mono Nerd Font "+nerdfont.Version+" for your user (about 130 MB to download). You can keep playing.")
	in := f.installer(func(s string) { f.log("info", s) })
	f.run(func() {
		res, err := in.Install(context.Background())
		f.post(func() { f.installed(res, err) })
	})
}

// installed reports the install's outcome (UI goroutine).
func (f *iconsFlow) installed(res nerdfont.Result, err error) {
	f.stage = iconsIdle
	if err != nil {
		f.log("error", "The font was not installed: "+err.Error()+". Nothing else was changed. You can get it from "+nerdFontsPage+".")
		return
	}
	if res.ManualStep != "" {
		f.log("warning", "One step is left for you: "+res.ManualStep)
	}
	f.log("success", "JetBrains Mono Nerd Font is installed. Now pick it in your terminal:")
	f.terminalSteps()
	f.log("info", restartMessage)
}

// terminalSteps logs how to pick the font in the detected terminal.
func (f *iconsFlow) terminalSteps() {
	name, steps := terminalFontSteps(f.environ(), f.goos)
	f.log("info", "In "+name+":")
	for i, s := range steps {
		f.log("info", "  "+strconv.Itoa(i+1)+". "+s)
	}
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

// startIcons begins the check (the icons command).
func (d *Dashboard) startIcons() { d.icons.start() }

// setMapGlyphsNerd is the check's "icons" answer: map glyphs nerd, saved to
// the account (or the session, with no account).
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
