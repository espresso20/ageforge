package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// Toast represents a single toast notification
type Toast struct {
	Message  string
	Color    string
	Duration time.Duration
	Expiry   time.Time
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

// Current is the toast on show as coloured text, or "" when none is showing
// (thread-safe).
func (tm *ToastManager) Current() string {
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
		return ""
	}
	return fmt.Sprintf("[%s]%s[-]", tm.current.Color, tm.current.Message)
}

// chainToast is the milestone-chain toast: "Chain complete: Settlement Chain.
// Title: The Founders. Game speed +300% for ~1m 30s." The boost's size comes
// from the chain's config and its length in ticks from the event (boostTicks:
// the engine stretches it for the age it lands in); 0 falls back to the
// config's base-curve length. The duration assumes speed 1x, since the bus
// handler that calls this cannot read the engine's speed (hence the "~").
func chainToast(name, title string, chain config.MilestoneChainDef, boostTicks int) string {
	msg := rowid.Sprintf("ui.toast.chainToast.01", "Chain complete: %s. Title: %s.", name, title)
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
	return msg + rowid.Sprintf("ui.toast.chainToast.02", " Game speed %s for %s.",
		textfmt.SignedPercent(chain.BoostValue), game.DurationText(boostTicks, interval))
}
