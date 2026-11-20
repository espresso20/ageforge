package game

import (
	"sync"
	"time"
)

// hold.go holds the simulation for the pause card. The tick loop keeps its
// timer but runs no tick and no autosave while the engine is held; the time
// held is added to the run's start and to the autosave clock, so it is neither
// time played nor time until the next autosave.

// holdState is the held flag and the wall-clock marks the loop works from.
type holdState struct {
	mu       sync.Mutex
	at       time.Time // when the hold began or was last settled; zero when not held
	autosave time.Time // the last autosave, or the loop's start
	saved    time.Time // the last save or load
}

// with runs f on the state under its lock.
func (ge *GameEngine) with(f func(h *holdState)) {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	f(&ge.hold)
}

// settle moves the hold's mark to now (clearing it when end is set), pushes the
// autosave clock on by the time held since the mark, and moves the run's start
// on by the same, so play time stands still. A no-op when not held.
func (ge *GameEngine) settle(end bool) {
	var d time.Duration
	ge.with(func(h *holdState) {
		if h.at.IsZero() {
			return
		}
		now := time.Now()
		d, h.at = now.Sub(h.at), now
		if end {
			h.at = time.Time{}
		}
		h.autosave = h.autosave.Add(d)
	})
	if d > 0 {
		ge.mu.Lock()
		if ge.Stats != nil {
			ge.Stats.GameStarted = ge.Stats.GameStarted.Add(d)
		}
		ge.mu.Unlock()
	}
}

// Hold stops the simulation until Release. Holding again changes nothing.
func (ge *GameEngine) Hold() {
	ge.with(func(h *holdState) {
		if h.at.IsZero() {
			h.at = time.Now()
		}
	})
}

// Release lets the simulation go on; the time held is not time played.
func (ge *GameEngine) Release() { ge.settle(true) }

// Held reports whether the simulation is held.
func (ge *GameEngine) Held() (held bool) {
	ge.with(func(h *holdState) { held = !h.at.IsZero() })
	return
}

// heldTick is the loop's check for each timer: when held it settles and
// reports true, and the loop skips the tick.
func (ge *GameEngine) heldTick() bool {
	if !ge.Held() {
		return false
	}
	ge.settle(false)
	return true
}

// SavedAt is when the game was last saved or loaded; zero before either.
func (ge *GameEngine) SavedAt() (t time.Time) {
	ge.with(func(h *holdState) { t = h.saved })
	return
}

// NextAutosave is how long until the loop's next autosave; zero when the loop
// is not running.
func (ge *GameEngine) NextAutosave() (d time.Duration) {
	ge.with(func(h *holdState) {
		if !h.autosave.IsZero() {
			d = max(0, AutosaveInterval-time.Since(h.autosave))
		}
	})
	return
}

func (ge *GameEngine) noteSaved(t time.Time) { ge.with(func(h *holdState) { h.saved = t }) }

func (ge *GameEngine) autosaveDue() (due bool) {
	ge.with(func(h *holdState) { due = time.Since(h.autosave) >= AutosaveInterval })
	return
}

func (ge *GameEngine) markAutosave() { ge.with(func(h *holdState) { h.autosave = time.Now() }) }
