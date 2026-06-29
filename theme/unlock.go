package theme

// Milestone-gated unlock resolution (theming.md §5).
//
// Gated flavor themes declare their unlock condition as a milestone-or-chain key in
// the registry (Theme.UnlockMilestone / Theme.UnlockChain). The UI's only job is the
// reverse direction: "a milestone/chain just completed — does it grant a theme?"
// UnlockedBy answers that without the UI knowing any theme→key mapping itself, so the
// mapping stays owned here and theme keys never leak into engine/milestone code.
//
// theme remains a leaf package: this is pure string indexing over the registry, no
// game import.

// unlockIndex maps a completed milestone/chain key → the theme key it unlocks. Built
// once at init() from the registry's gated themes. A theme contributes its single
// UnlockKey() (milestone XOR chain); always-available themes contribute nothing.
//
// If two themes ever claimed the same unlock key it would be a registry bug (one key
// can't deterministically grant two themes); init() panics on that collision, the
// same fail-fast posture register() takes for duplicate theme keys.
var unlockIndex = map[string]string{}

// buildUnlockIndex (re)builds unlockIndex from the current registry. Called from a
// dedicated init() AFTER the themes_*.go init()s have registered every theme
// (Go runs init() in file-name order within a package; this file sorts after the
// themes_*.go files, and the index is also rebuilt lazily-safe below). It is
// idempotent: it clears and refills, so a test that registers a stub theme can call
// it again.
func buildUnlockIndex() {
	idx := make(map[string]string, len(registryOrder))
	for _, key := range registryOrder {
		t := registry[key]
		uk := t.UnlockKey()
		if uk == "" {
			continue // always-available (Accessible / Forge) — nothing to unlock
		}
		if existing, dup := idx[uk]; dup {
			panic("theme: unlock key " + uk + " maps to both " + existing + " and " + t.Key)
		}
		idx[uk] = t.Key
	}
	unlockIndex = idx
}

func init() { buildUnlockIndex() }

// UnlockedBy reverse-maps a completed milestone OR chain key to the theme key it
// unlocks. ok is false when the key unlocks no theme (the common case — most
// milestones grant no theme). This is the UI's unlock hook: on a newly-completed
// milestone/chain it asks UnlockedBy, and on ok it unlocks the mapped theme via the
// account layer (theming.md §5).
func UnlockedBy(completedKey string) (themeKey string, ok bool) {
	themeKey, ok = unlockIndex[completedKey]
	return
}

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
