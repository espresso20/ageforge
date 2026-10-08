package theme

// A gated theme is the reward of an account badge (Theme.UnlockBadge). The
// engine unlocks it when the badge is earned, so there is no index here from
// a key to a theme: the theme only says which badge gives it, and how a
// locked theme's condition reads.
//
// theme remains a leaf package: no game import.

// UnlockHintFor returns the human-readable unlock condition for a theme key (e.g.
// "Reach the Cyberpunk Age"), or "" for an unknown key or an un-gated theme. Used by
// the picker detail pane and `theme list` to show LOCKED themes' conditions.
func UnlockHintFor(themeKey string) string {
	t, ok := ByKey(themeKey)
	if !ok {
		return ""
	}
	return t.UnlockHint
}

// GivenBy returns the keys of the themes the badge gives, in registry order.
func GivenBy(badge string) []string {
	var out []string
	if badge == "" {
		return nil
	}
	for _, t := range All() {
		if t.UnlockBadge == badge {
			out = append(out, t.Key)
		}
	}
	return out
}
