package game

import (
	"sync"
	"time"
)

// hold.go holds the simulation for the pause card. The tick loop keeps its
// timer but runs no tick, no autosave and no catch-up while the engine is
// held; the time held is added to the run's start and to the autosave clock,
// so it is neither time played nor time until the next autosave.

// holdState is the held flag and the wall-clock marks the loop works from.
type holdState struct {
	mu       sync.Mutex
	at       time.Time // when the hold began or was last settled; zero when not held
	autosave time.Time // the last autosave, or the loop's start
	saved    time.Time // the last successful save
}

// settle returns the time held since the last settle and moves the mark to
// now (or clears it when end is set). It is zero when not held.
func (h *holdState) settle(now time.Time, end bool) time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.at.IsZero() {
		return 0
	}
	d := now.Sub(h.at)
	if end {
		h.at = time.Time{}
	} else {
		h.at = now
	}
	h.autosave = h.autosave.Add(d)
	return d
}

// Hold stops the simulation until Release. Calling it again changes nothing.
func (ge *GameEngine) Hold() {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	if ge.hold.at.IsZero() {
		ge.hold.at = time.Now()
	}
}

// Release lets the simulation go on. The time held does not count as time
// played.
func (ge *GameEngine) Release() { ge.shiftStart(ge.hold.settle(time.Now(), true)) }

// Held reports whether the simulation is held.
func (ge *GameEngine) Held() bool {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	return !ge.hold.at.IsZero()
}

// heldTick is called by the tick loop for each timer: when held it keeps the
// time held out of the play time and reports true, and the loop skips the tick.
func (ge *GameEngine) heldTick() bool {
	if !ge.Held() {
		return false
	}
	ge.shiftStart(ge.hold.settle(time.Now(), false))
	return true
}

// shiftStart moves the run's start on by d.
func (ge *GameEngine) shiftStart(d time.Duration) {
	if d <= 0 {
		return
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.Stats != nil {
		ge.Stats.GameStarted = ge.Stats.GameStarted.Add(d)
	}
}

// SavedAt is when the game was last saved or loaded; zero before either.
func (ge *GameEngine) SavedAt() time.Time {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	return ge.hold.saved
}

// NextAutosave is how long until the loop's next autosave; zero when the loop
// is not running.
func (ge *GameEngine) NextAutosave() time.Duration {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	if ge.hold.autosave.IsZero() {
		return 0
	}
	return max(0, AutosaveInterval-time.Since(ge.hold.autosave))
}

func (ge *GameEngine) noteSaved() { ge.noteSavedAt(time.Now()) }

func (ge *GameEngine) noteSavedAt(t time.Time) {
	ge.hold.mu.Lock()
	ge.hold.saved = t
	ge.hold.mu.Unlock()
}

func (ge *GameEngine) autosaveDue() bool {
	ge.hold.mu.Lock()
	defer ge.hold.mu.Unlock()
	return time.Since(ge.hold.autosave) >= AutosaveInterval
}

func (ge *GameEngine) markAutosave() {
	ge.hold.mu.Lock()
	ge.hold.autosave = time.Now()
	ge.hold.mu.Unlock()
}
