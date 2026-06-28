package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// This file is the Phase-2 bridge between the leaf `theme` package (pure
// presentation, no game dependency) and the `game` account layer (theming.md §6,
// §10). The bridge lives in `ui` precisely so `theme` stays importable without
// pulling in the engine: ui imports both and wires them together here.
//
// Two concerns live here:
//   1. applyAccountTheme — read the account's persisted active theme at boot and
//      apply it (routed through the unlock gate) before the first Draw.
//   2. themeAvailable — the unlock-gating predicate. It is a no-op gate today (all
//      shipped themes are accessible or Forge), but the picker, boot-apply, and the
//      `theme <key>` command all route through it so Phase 3's milestone-gated
//      flavor themes gate correctly with no further wiring.

// themeUnavailableMsg is the player-facing refusal when a theme isn't unlocked yet
// (a locked flavor theme in Phase 3). No theme is unavailable today.
const themeUnavailableMsg = "that theme isn't unlocked yet"

// themeAvailable reports whether theme t may be selected/applied for the given
// account. The policy (theming.md §4/§5/§6):
//   - Accessible themes are NEVER gated — always available.
//   - The default theme (Forge) is always available.
//   - Any other (flavor) theme is available only if the account has unlocked it.
//
// acct may be nil — the game must run accountless. With no account, only the
// always-available set (Accessible + Forge) is available; flavor themes stay locked
// until an account exists to own the unlock. Since every theme shipped today is
// Accessible or Forge, this returns true for all of them now.
func themeAvailable(acct *game.Account, t theme.Theme) bool {
	if t.Accessible || t.Key == theme.DefaultKey {
		return true
	}
	return acct != nil && acct.HasTheme(t.Key)
}

// applyAccountTheme resolves and applies the account's persisted active theme at
// startup (theming.md §6): it must run once, before the first Draw, so the splash
// already wears the chosen theme.
//
// Resolution is defensive — any of these fall back to the default (Forge) without
// erroring, so a brand-new/wiped account (ActiveTheme == "") or a stale/unknown
// stored key never crashes and never strands the player on an inapplicable theme:
//   - no engine / no account → leave the default in place.
//   - stored key is empty → leave the default in place.
//   - stored key is unknown to the registry → fall back to default.
//   - stored theme exists but is NOT available to this account (a locked flavor
//     theme; impossible today) → fall back to default.
//
// It applies via theme.SetActive + theme.Restyle. Any redraw is the caller's job;
// at boot the first Draw happens after setup, so no explicit redraw is needed here.
func applyAccountTheme(engine *game.GameEngine) {
	if engine == nil {
		applyDefaultTheme()
		return
	}
	acct := engine.Account()
	if acct == nil {
		applyDefaultTheme()
		return
	}
	key := acct.ActiveTheme()
	if key == "" {
		applyDefaultTheme()
		return
	}
	t, ok := theme.ByKey(key)
	if !ok || !themeAvailable(acct, t) {
		applyDefaultTheme()
		return
	}
	_ = theme.SetActive(t.Key)
	theme.Restyle()
}

// applyDefaultTheme pins the active theme to the default (Forge). Used as the
// fallback when no account theme applies, and to reset a freshly-created/wiped
// account onto its own (empty → default) theme rather than inheriting a previewed
// or previous one (theming.md §6).
func applyDefaultTheme() {
	_ = theme.SetActive(theme.DefaultKey)
	theme.Restyle()
}
