package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// Milestone-gated theme unlocking — the UI side of theming.md §5.
//
// WHY THIS LIVES IN THE UI GOROUTINE, NOT THE ENGINE:
// account.UnlockTheme persists the account (file I/O via Save). The documented
// Bus-deadlock rule (CLAUDE.md / MEMORY.md) forbids file I/O — and any
// lock-acquiring engine call — inside a Bus handler or under the engine write lock.
// So the unlock check runs in the dashboard's per-refresh path, which executes
// inside app.QueueUpdateDraw on the tview goroutine and does NOT hold ge.mu (it
// reads a GetState() snapshot). That makes the Save here safe.
//
// FIRST-SYNC / RETROACTIVE UNLOCKS:
// On the first refresh after a load, every already-completed milestone/chain looks
// "new to this process," so we must unlock their mapped themes (a player who reached
// the Cyberpunk Age before this update still earns Cyberpunk) WITHOUT firing a toast
// per retroactive unlock — that would spam the log on every launch. The mechanism:
// the dashboard's themeSyncDone flag is false on the first pass, so unlocks happen
// silently; it flips true afterward, and only unlocks during live play toast.

// themeUnlockResult is the decision for one completed milestone/chain key.
type themeUnlockResult struct {
	// Toast is true only when a genuinely-new unlock happened during live play (not
	// the first retroactive sync). The caller AddLogs iff Toast.
	Toast bool
	// ThemeName is the unlocked theme's display name, for the toast text. Set only
	// when Toast is true.
	ThemeName string
}

// evaluateThemeUnlock decides what to do with a single completed milestone/chain key
// (theming.md §5). It is the pure-ish core, split out from the dashboard method so it
// can be unit-tested without a tview app:
//
//   - If completedKey maps to no theme → no-op (the common case).
//   - If it maps to a theme and acct != nil → unlock it via account.UnlockTheme.
//     Toast iff the unlock was genuinely NEW (newly==true) AND this isn't the
//     first-sync pass (firstSync==false). UnlockTheme is idempotent, so replaying a
//     key returns newly==false and never re-toasts.
//   - acct == nil (accountless play) → nothing to unlock; no toast.
//
// The account Save inside UnlockTheme is acceptable here: callers invoke this from
// the UI goroutine, never under the engine lock or in a Bus handler.
func evaluateThemeUnlock(acct *game.Account, completedKey string, firstSync bool) themeUnlockResult {
	themeKey, ok := theme.UnlockedBy(completedKey)
	if !ok {
		return themeUnlockResult{} // this milestone/chain grants no theme
	}
	if acct == nil {
		// Accountless: no store to own the unlock. The flavor theme stays locked until
		// an account exists; nothing to toast.
		return themeUnlockResult{}
	}
	newly, _ := acct.UnlockTheme(themeKey)
	// UnlockTheme's error is the account Save error — non-fatal here (the unlock is
	// already in memory and the theme is cosmetic), so we don't surface it: a write
	// hiccup must not block play, and the next FlushIfDirty/Save retries it.
	if !newly || firstSync {
		// Already owned (replayed check), or a retroactive first-sync unlock we grant
		// silently to avoid a toast-per-already-earned-theme on every launch.
		return themeUnlockResult{}
	}
	name := themeKey
	if th, found := theme.ByKey(themeKey); found {
		name = th.Name
	}
	return themeUnlockResult{Toast: true, ThemeName: name}
}

// themeUnlockToast renders the unlock notification text (theming.md §7). The accent
// color comes from the active theme's role tag so the toast reads in whatever theme
// is live. Kept separate (and pure) so the message format is testable.
func themeUnlockToast(themeName string) string {
	return fmt.Sprintf("%s🎨 New theme unlocked: %s! Type 'theme' to use it.[-]",
		theme.Tag(theme.RoleAccent), themeName)
}

// completedUnlockKeys gathers every completed milestone key AND completed chain key
// from a GameState snapshot — the universe of keys that might grant a theme. It does
// not touch the account or the registry; it's just the snapshot read, isolated so
// the dashboard hook (and its test) share one definition of "completed."
func completedUnlockKeys(ms game.MilestoneState) []string {
	keys := make([]string, 0, len(ms.Milestones)+len(ms.Chains))
	for key, info := range ms.Milestones {
		if info.Completed {
			keys = append(keys, key)
		}
	}
	for _, ch := range ms.Chains {
		if ch.Complete {
			keys = append(keys, ch.Key)
		}
	}
	return keys
}
