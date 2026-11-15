package theme

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// DefaultKey is the theme applied when nothing else is selected. Forge is the
// shipped default (the theming design §4).
const DefaultKey = "forge"

// active is the package-global current theme: there is exactly one active
// theme per process. Guarded by mu.
//
// activeColors mirrors active.Colors behind an atomic pointer so the painter,
// which resolves colors for every cell of every frame, reads the palette
// without taking mu or copying the whole Theme.
var (
	mu           sync.RWMutex
	active       Theme
	activeColors atomic.Pointer[[numRoles]Color]
)

// init seeds the active theme to Forge and applies its remap so the very first
// Draw (splash) already wears a coherent palette. Themes register through
// package-level var initialization (themes_*.go), which Go completes before any
// init() runs, so the registry is fully populated here regardless of file order.
func init() {
	t, ok := ByKey(DefaultKey)
	if !ok {
		// Should never happen — Forge registers in themes_forge.go. If a future
		// refactor breaks that, don't panic the whole program over a theme; just
		// run with the zero value until SetActive is called.
		return
	}
	setActive(t)
}

func setActive(t Theme) {
	mu.Lock()
	active = t
	mu.Unlock()
	cols := t.Colors
	activeColors.Store(&cols)
}

// Active returns the currently active theme (a copy).
func Active() Theme {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// SetActive switches the active theme by key: it updates the package-global
// active theme and re-applies the restylable-widget registry (restyle.go) so
// live chrome re-pulls its colors.
//
// It does NOT trigger a redraw: the caller owns that, since only the UI layer
// holds the running app. Returns an error for unknown keys
// and leaves the active theme unchanged in that case.
func SetActive(key string) error {
	t, ok := ByKey(key)
	if !ok {
		return fmt.Errorf("theme: unknown theme %q", key)
	}
	setActive(t)

	Restyle()
	return nil
}

// Of returns the active theme's color for role, as it is now: a literal. Code
// that keeps a color across a theme switch wants Ref instead.
func Of(role Role) Color {
	if role < 0 || role >= numRoles {
		return ColorDefault
	}
	if cols := activeColors.Load(); cols != nil {
		return cols[role]
	}
	return ColorDefault
}

// IsLight reports whether the active theme has a light background.
func IsLight() bool {
	return RelativeLuminance(Of(RoleBackground)) >= lightLuminanceThreshold
}
