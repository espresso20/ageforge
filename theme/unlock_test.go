package theme

import "testing"

// TestUnlockedByRoundTrip checks the reverse lookup: every gated theme's declared
// unlock key maps back to that exact theme, and an unmapped key returns ok=false.
func TestUnlockedByRoundTrip(t *testing.T) {
	gatedSeen := 0
	for _, th := range All() {
		if !th.Gated() || th.UnlockBadge != "" {
			continue // a badge's theme is unlocked by the engine, not through this index
		}
		gatedSeen++
		key := th.UnlockKey()
		gotTheme, ok := UnlockedBy(key)
		if !ok {
			t.Errorf("UnlockedBy(%q): ok=false, want the theme %q", key, th.Key)
			continue
		}
		if gotTheme != th.Key {
			t.Errorf("UnlockedBy(%q) = %q, want %q", key, gotTheme, th.Key)
		}
	}
	if gatedSeen == 0 {
		t.Fatal("no gated themes found — expected the 5 flavor themes to be gated")
	}

	// An unmapped key (a real milestone that grants no theme) returns ok=false.
	if got, ok := UnlockedBy("first_shelter"); ok {
		t.Errorf("UnlockedBy(\"first_shelter\") = (%q, true), want ok=false", got)
	}
	// A completely bogus key likewise.
	if got, ok := UnlockedBy("totally_not_a_key"); ok {
		t.Errorf("UnlockedBy(\"totally_not_a_key\") = (%q, true), want ok=false", got)
	}
	// Empty key never unlocks (un-gated themes contribute "" and must not be indexed).
	if got, ok := UnlockedBy(""); ok {
		t.Errorf("UnlockedBy(\"\") = (%q, true), want ok=false", got)
	}
}

// TestGatedThemeRegistryConsistency is the registry-consistency guard (the theming design
// §5): every gated theme (non-Accessible, non-default) declares EXACTLY ONE of
// UnlockMilestone / UnlockChain and a non-empty UnlockHint; every always-available
// theme (Accessible or the default Forge) declares NEITHER and no hint.
func TestGatedThemeRegistryConsistency(t *testing.T) {
	for _, th := range All() {
		alwaysAvailable := th.AlwaysAvailable()
		hasMilestone := th.UnlockMilestone != ""
		hasChain := th.UnlockChain != ""
		hasBadge := th.UnlockBadge != ""

		if alwaysAvailable {
			if hasMilestone || hasChain || hasBadge {
				t.Errorf("theme %q is always-available (accessible=%v, default=%v) but declares an unlock condition (milestone=%q chain=%q); it must declare neither",
					th.Key, th.Accessible, th.Key == DefaultKey, th.UnlockMilestone, th.UnlockChain)
			}
			if th.UnlockHint != "" {
				t.Errorf("theme %q is always-available but has UnlockHint=%q; it must be empty", th.Key, th.UnlockHint)
			}
			if th.Gated() {
				t.Errorf("theme %q is always-available but Gated() reports true", th.Key)
			}
			continue
		}

		// Gated theme: exactly one of milestone, chain and badge, and a hint.
		set := 0
		for _, has := range []bool{hasMilestone, hasChain, hasBadge} {
			if has {
				set++
			}
		}
		if set != 1 {
			t.Errorf("gated theme %q must set EXACTLY ONE of UnlockMilestone/UnlockChain/UnlockBadge; got milestone=%q chain=%q badge=%q",
				th.Key, th.UnlockMilestone, th.UnlockChain, th.UnlockBadge)
		}
		if hasBadge {
			if got, ok := UnlockedBy(th.UnlockBadge); ok {
				t.Errorf("badge theme %q is in the milestone unlock index (as %q): the engine unlocks it", th.Key, got)
			}
			switch th.Effect {
			case "", EffectRain, EffectGlitch:
			default:
				t.Errorf("theme %q carries the effect %q, which the UI does not draw", th.Key, th.Effect)
			}
		}
		if th.UnlockHint == "" {
			t.Errorf("gated theme %q has an empty UnlockHint; a locked theme must explain how to unlock it", th.Key)
		}
		if !th.Gated() {
			t.Errorf("gated theme %q reports Gated()=false", th.Key)
		}
		// UnlockKey must return the set key.
		if want := th.UnlockMilestone + th.UnlockChain; th.UnlockKey() != want {
			t.Errorf("theme %q UnlockKey()=%q, want %q", th.Key, th.UnlockKey(), want)
		}
	}
}

// TestUnlockHintFor checks the hint accessor for known/unknown keys.
func TestUnlockHintFor(t *testing.T) {
	// A gated theme's hint matches its struct field.
	if got := UnlockHintFor("cyberpunk"); got != Cyberpunk.UnlockHint {
		t.Errorf("UnlockHintFor(\"cyberpunk\") = %q, want %q", got, Cyberpunk.UnlockHint)
	}
	if got := UnlockHintFor("cyberpunk"); got == "" {
		t.Error("UnlockHintFor(\"cyberpunk\") is empty; want a hint")
	}
	// The default theme has no hint.
	if got := UnlockHintFor(DefaultKey); got != "" {
		t.Errorf("UnlockHintFor(%q) = %q, want empty (un-gated)", DefaultKey, got)
	}
	// Unknown key → "".
	if got := UnlockHintFor("no_such_theme"); got != "" {
		t.Errorf("UnlockHintFor(\"no_such_theme\") = %q, want empty", got)
	}
}
