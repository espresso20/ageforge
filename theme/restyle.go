package theme

import "sync"

// The restylable-widget registry routes Path B (direct widget chrome calls —
// SetBorderColor, SetBackgroundColor, list selection, …) through a live theme
// switch. Those setters apply once at construction, so a live switch needs each
// long-lived chrome widget to re-pull its colors. See theming.md §3.3.
//
// Registration discipline (§3.3): a registry someone must *remember* to populate
// will rot — some future widget ships unthemed because nobody touched the slice.
// Track makes registration the default path: it applies the closure *now* AND
// enrolls it, in one call, so a chrome widget cannot be created without being
// tracked. The UI registers widgets here in Phase 1b; this package only provides
// the mechanism.

var (
	restyleMu sync.Mutex
	restylers []func()
)

// Track runs apply() immediately (so the widget is styled at construction) and
// stores it for re-application on every later theme switch. This is the single
// combined "apply-now-and-enroll" operation that prevents untracked chrome.
//
// The apply closure should pull from theme.Color(role) each time it runs — never
// capture a tcell.Color once — so it reflects whatever theme is active when
// Restyle() re-invokes it. Typical shape:
//
//	theme.Track(func() {
//	    box.SetBorderColor(theme.Color(theme.RoleAccent)).
//	        SetTitleColor(theme.Color(theme.RoleAccent))
//	})
//
// Track is safe for concurrent use. A nil apply is ignored.
func Track(apply func()) {
	if apply == nil {
		return
	}
	apply()

	restyleMu.Lock()
	restylers = append(restylers, apply)
	restyleMu.Unlock()
}

// Restyle re-runs every tracked apply closure. SetActive calls it after applyRemap
// and before the caller's redraw, so existing chrome re-pulls the new palette.
// The closures run outside the lock so a closure may itself call Track without
// deadlocking (it just enrolls into the next pass).
func Restyle() {
	restyleMu.Lock()
	snapshot := make([]func(), len(restylers))
	copy(snapshot, restylers)
	restyleMu.Unlock()

	for _, apply := range snapshot {
		apply()
	}
}

// resetRestylersForTest clears the registry. Test-only: lets a test assert Track
// enrolls without leaking closures across tests in this package.
func resetRestylersForTest() {
	restyleMu.Lock()
	restylers = nil
	restyleMu.Unlock()
}
