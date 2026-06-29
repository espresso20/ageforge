package theme

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// DefaultKey is the theme applied when nothing else is selected. Forge is the
// shipped default (theming.md §4).
const DefaultKey = "forge"

// active is the package-global current theme. It is process-wide state by design:
// the name-remap in remap.go mutates global tcell.ColorNames, so there is exactly
// one active theme per process. Guarded by mu.
var (
	mu     sync.RWMutex
	active Theme
)

// init seeds the active theme to Forge and applies its remap so the very first
// Draw (splash) already wears a coherent palette. The themes_*.go init()s run
// before this (same package, registration order is forge → accessible), so the
// registry is populated by now; we fall back to a zero Theme defensively.
func init() {
	t, ok := ByKey(DefaultKey)
	if !ok {
		// Should never happen — Forge registers in themes_forge.go init(). If a
		// future refactor breaks that, don't panic the whole program over a theme;
		// just run with the zero value until SetActive is called.
		return
	}
	active = t
	applyRemap(t)
}

// Active returns the currently active theme (a copy).
func Active() Theme {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// SetActive switches the active theme by key: it updates the package-global
// active theme, rewrites the tcell.ColorNames remap (remap.go), and re-applies the
// restylable-widget registry (restyle.go) so live chrome re-pulls its colors.
//
// It does NOT trigger a redraw — the caller owns app.Draw / QueueUpdateDraw, since
// only the UI layer holds the tview.Application. Returns an error for unknown keys
// and leaves the active theme unchanged in that case.
func SetActive(key string) error {
	t, ok := ByKey(key)
	if !ok {
		return fmt.Errorf("theme: unknown theme %q", key)
	}
	mu.Lock()
	active = t
	mu.Unlock()

	// Order matters: remap the named tags first (so any restyle closure that pulls
	// a color sees the new palette), then re-run the widget restyle pass. Both are
	// idempotent. The caller redraws afterward.
	applyRemap(t)
	Restyle()
	return nil
}

// Color returns the active theme's color for role. Path B (direct widget chrome)
// routes through here instead of tcell color literals (theming.md §3.3).
func Color(role Role) tcell.Color {
	mu.RLock()
	defer mu.RUnlock()
	return active.Color(role)
}

// Tag returns the active theme's color for role as a tview inline color tag, e.g.
// "[#ffd700]". This is the §3.2 fallback / future semantic-token form (§9): a
// helper that emits the active hex at format time, independent of the ColorNames
// remap. Unused by Phase 1a UI but provided so the fallback path exists.
func Tag(role Role) string {
	return tagFor(Color(role))
}

// tagFor renders a tcell.Color as a tview "[#rrggbb]" tag. Colors carry true RGB
// (themes are built with NewRGBColor), so RGB() yields the literal components.
func tagFor(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("[#%02x%02x%02x]", r, g, b)
}
