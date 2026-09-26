package theme

import "github.com/gdamore/tcell/v2"

// screen.go is the late-binding half of the background decision (theming.md
// §3.7). tview widgets capture their colors once, at construction, from
// tview.Styles. Before the overhaul those were concrete RGB values, so a widget
// built under Forge kept Forge's canvas after a live switch to a light theme —
// dark boxes with dark ink on them. Rather than enrolling every widget in the
// restyle registry, chrome now carries Ref(role) sentinels and the screen itself
// resolves them against the ACTIVE theme on every SetContent. A theme switch is
// then just "redraw": nothing to re-apply, nothing to forget.
//
// WrapScreen also resolves tcell.ColorDefault — fg to Text, bg to Background —
// so code that draws with tcell.StyleDefault (splash canvas, map overlays) paints
// the theme canvas instead of the terminal's own default. That is what makes
// every theme paint an explicit background: Daylight looks like Daylight on a
// black terminal, and Forge looks like Forge on a white one.

// refBase is the first sentinel. Sentinels are ColorValid palette indices far
// beyond the 256-color range, never IsRGB, so they cannot collide with a real
// theme color (all themes are RGB) or a real palette slot.
const refBase = tcell.ColorValid | 0xA0000

// Ref returns a late-bound color for role. It is only meaningful on a screen
// wrapped by WrapScreen (the app installs one before Run); tview.Styles is
// populated with Refs by applyRemap.
func Ref(role Role) tcell.Color {
	if role < 0 || role >= numRoles {
		return tcell.ColorDefault
	}
	return refBase + tcell.Color(role)
}

// RefRole reports the role a sentinel stands for, and ok=false for any other
// color (including real RGB colors and tcell.ColorDefault).
func RefRole(c tcell.Color) (Role, bool) {
	if c < refBase || c >= refBase+tcell.Color(numRoles) {
		return 0, false
	}
	return Role(c - refBase), true
}

// Resolve maps a sentinel to the active theme's concrete color; anything else is
// returned unchanged.
func Resolve(c tcell.Color) tcell.Color {
	if r, ok := RefRole(c); ok {
		return Color(r)
	}
	return c
}

// ResolveStyle resolves both ends of a style against the active theme: sentinels
// become their role color, ColorDefault fg becomes Text and ColorDefault bg
// becomes Background. Attributes and URLs are preserved.
func ResolveStyle(st tcell.Style) tcell.Style {
	return resolveStyleFor(st, Active())
}

func resolveStyleFor(st tcell.Style, t Theme) tcell.Style {
	fg, bg, _ := st.Decompose()
	nfg, nbg := fg, bg
	if r, ok := RefRole(fg); ok {
		nfg = t.Color(r)
	} else if fg == tcell.ColorDefault {
		nfg = t.Color(RoleText)
	}
	if r, ok := RefRole(bg); ok {
		nbg = t.Color(r)
	} else if bg == tcell.ColorDefault {
		nbg = t.Color(RoleBackground)
	}
	if nfg != fg {
		st = st.Foreground(nfg)
	}
	if nbg != bg {
		st = st.Background(nbg)
	}
	return st
}

// ThemedScreen wraps a tcell.Screen and resolves theme sentinels on every write.
// Every drawing entry point tcell exposes is overridden so nothing can bypass
// resolution; everything else is delegated untouched.
type ThemedScreen struct {
	tcell.Screen

	initDone bool
	initErr  error
}

// Init initializes the underlying screen once. It is idempotent so the caller
// can Init (and check the error) before handing the screen to
// tview.Application.SetScreen, which calls Init again and discards the result.
func (s *ThemedScreen) Init() error {
	if s.initDone {
		return s.initErr
	}
	s.initDone = true
	s.initErr = s.Screen.Init()
	return s.initErr
}

// WrapScreen returns s wrapped so theme sentinels and default colors resolve to
// the active theme. Wrapping an already-wrapped screen returns it as-is.
func WrapScreen(s tcell.Screen) *ThemedScreen {
	if ts, ok := s.(*ThemedScreen); ok {
		return ts
	}
	return &ThemedScreen{Screen: s}
}

// Unwrap returns the underlying screen (tests use it to reach a SimulationScreen).
func (s *ThemedScreen) Unwrap() tcell.Screen { return s.Screen }

func (s *ThemedScreen) canvas() tcell.Style {
	t := Active()
	return tcell.StyleDefault.Foreground(t.Color(RoleText)).Background(t.Color(RoleBackground))
}

// Clear fills the screen with the active theme's canvas instead of the
// terminal's default colors, and makes that the underlying default style.
func (s *ThemedScreen) Clear() {
	c := s.canvas()
	s.Screen.SetStyle(c)
	s.Screen.Fill(' ', c)
}

// Fill resolves the style before filling.
func (s *ThemedScreen) Fill(r rune, st tcell.Style) { s.Screen.Fill(r, ResolveStyle(st)) }

// SetStyle resolves the style before installing it as the default.
func (s *ThemedScreen) SetStyle(st tcell.Style) { s.Screen.SetStyle(ResolveStyle(st)) }

// SetContent is the path tview uses for every cell.
func (s *ThemedScreen) SetContent(x, y int, primary rune, combining []rune, st tcell.Style) {
	s.Screen.SetContent(x, y, primary, combining, ResolveStyle(st))
}

// SetCell resolves the style before writing.
func (s *ThemedScreen) SetCell(x, y int, st tcell.Style, ch ...rune) {
	s.Screen.SetCell(x, y, ResolveStyle(st), ch...)
}

// Put resolves the style before writing.
func (s *ThemedScreen) Put(x, y int, str string, st tcell.Style) (string, int) {
	return s.Screen.Put(x, y, str, ResolveStyle(st))
}

// PutStr writes with the (resolved) default style.
func (s *ThemedScreen) PutStr(x, y int, str string) {
	s.Screen.PutStrStyled(x, y, str, ResolveStyle(tcell.StyleDefault))
}

// PutStrStyled resolves the style before writing.
func (s *ThemedScreen) PutStrStyled(x, y int, str string, st tcell.Style) {
	s.Screen.PutStrStyled(x, y, str, ResolveStyle(st))
}
