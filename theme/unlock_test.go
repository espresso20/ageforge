package theme

import "testing"

// TestGivenByRoundTrip checks the reverse lookup: every gated theme is given
// by the badge it names, and a badge that gives no theme gives none.
func TestGivenByRoundTrip(t *testing.T) {
	gated := 0
	for _, th := range All() {
		if !th.Gated() {
			continue
		}
		gated++
		if got := GivenBy(th.UnlockBadge); len(got) != 1 || got[0] != th.Key {
			t.Errorf("GivenBy(%q) = %v, want only %q: one badge gives one theme", th.UnlockBadge, got, th.Key)
		}
	}
	if gated != 10 {
		t.Errorf("%d gated themes, want the 10 flavor themes", gated)
	}
	for _, key := range []string{"", "age.stone_age", "totally_not_a_key"} {
		if got := GivenBy(key); len(got) != 0 {
			t.Errorf("GivenBy(%q) = %v, want none", key, got)
		}
	}
}

// TestGatedThemeRegistryConsistency is the registry-consistency guard: every
// gated theme (not Accessible, not Standard) names the badge that gives it
// and carries a hint; every always-available theme names none and has no
// hint. An effect is one the UI draws.
func TestGatedThemeRegistryConsistency(t *testing.T) {
	for _, th := range All() {
		if th.AlwaysAvailable() {
			if th.UnlockBadge != "" || th.UnlockHint != "" || th.Gated() {
				t.Errorf("theme %q is always available but declares an unlock (badge %q, hint %q)", th.Key, th.UnlockBadge, th.UnlockHint)
			}
			if th.Effect != "" {
				t.Errorf("theme %q is always available and carries the effect %q: an accessible theme holds still", th.Key, th.Effect)
			}
			continue
		}
		if th.UnlockBadge == "" || !th.Gated() {
			t.Errorf("gated theme %q names no badge that gives it", th.Key)
		}
		if th.UnlockHint == "" {
			t.Errorf("gated theme %q has an empty UnlockHint; a locked theme must explain how to unlock it", th.Key)
		}
		switch th.Effect {
		case "", EffectRain, EffectGlitch, EffectEmbers, EffectGreenbar, EffectPrism:
		default:
			t.Errorf("theme %q carries the effect %q, which the UI does not draw", th.Key, th.Effect)
		}
		if th.Effect == EffectGreenbar && EffectMoves(th.Effect) {
			t.Error("green-bar paper holds still: the motion setting must leave it on")
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
