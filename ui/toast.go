package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Toast represents a single toast notification
type Toast struct {
	Message  string
	Color    string
	Duration time.Duration
	Expiry   time.Time
	// Fit, when set, writes the toast for a bar w cells wide (0: any
	// width), with its own colour tags. It is asked again whenever the
	// toast is read, so a toast follows the bar through a resize.
	Fit func(w int) string
}

// ToastManager manages toast notifications with a queue
type ToastManager struct {
	mu      sync.Mutex
	current *Toast
	queue   []Toast
}

// NewToastManager creates a new toast manager
func NewToastManager() *ToastManager {
	return &ToastManager{}
}

// Show queues a toast notification (thread-safe)
func (tm *ToastManager) Show(message, color string, duration time.Duration) {
	tm.show(Toast{Message: message, Color: color, Duration: duration})
}

// ShowFit queues a toast that is written for the width of the toast bar
// and carries its own colour tags (thread-safe).
func (tm *ToastManager) ShowFit(fit func(w int) string, duration time.Duration) {
	tm.show(Toast{Fit: fit, Duration: duration})
}

func (tm *ToastManager) show(toast Toast) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	toast.Expiry = time.Now().Add(toast.Duration)
	if tm.current == nil || time.Now().After(tm.current.Expiry) {
		tm.current = &toast
	} else {
		tm.queue = append(tm.queue, toast)
	}
}

// GetCurrent returns the current toast text or empty string if none active
func (tm *ToastManager) GetCurrent() string { return tm.CurrentFor(0) }

// CurrentFor is GetCurrent for a toast bar w cells wide (0: any width).
func (tm *ToastManager) CurrentFor(w int) string {
	text, fit := tm.onShow()
	if fit != nil {
		// Written outside the manager's lock: a toast that fits itself to
		// the bar reads the theme and the account's settings.
		return fit(w)
	}
	return text
}

// onShow is the toast on show: its text, or the function that writes it.
func (tm *ToastManager) onShow() (text string, fit func(w int) string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()

	// Check if current is expired, promote from queue
	for tm.current != nil && now.After(tm.current.Expiry) {
		if len(tm.queue) > 0 {
			next := tm.queue[0]
			next.Expiry = now.Add(next.Duration)
			tm.current = &next
			tm.queue = tm.queue[1:]
		} else {
			tm.current = nil
		}
	}

	if tm.current == nil {
		return "", nil
	}
	if tm.current.Fit != nil {
		return "", tm.current.Fit
	}
	return fmt.Sprintf("[%s]%s[-]", tm.current.Color, tm.current.Message), nil
}

// chainToast is the milestone-chain toast: "Chain complete: Settlement Chain.
// Title: The Founders. Game speed +300% for ~1m 30s." The boost's size comes
// from the chain's config and its length in ticks from the event (boostTicks:
// the engine stretches it for the age it lands in); 0 falls back to the
// config's base-curve length. The duration assumes speed 1x, since the bus
// handler that calls this cannot read the engine's speed (hence the "~").
func chainToast(name, title string, chain config.MilestoneChainDef, boostTicks int) string {
	msg := fmt.Sprintf("Chain complete: %s. Title: %s.", name, title)
	if boostTicks <= 0 {
		boostTicks = chain.BoostDuration
	}
	if chain.BoostValue <= 0 || boostTicks <= 0 {
		return msg
	}
	interval := time.Duration(float64(game.BaseTickInterval) / (1 + chain.BoostValue))
	if interval < game.MinTickInterval {
		interval = game.MinTickInterval
	}
	return msg + fmt.Sprintf(" Game speed %s for %s.",
		textfmt.SignedPercent(chain.BoostValue), game.DurationText(boostTicks, interval))
}
