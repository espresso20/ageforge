package theme

import "testing"

// TestGatedThemeRegistryConsistency is the registry-consistency guard: every
// gated theme (not Accessible, not Standard) names the badge that gives it
// and carries a hint; every always-available theme names none and has no
// hint. An effect is one the UI draws.
func TestGatedThemeRegistryConsistency(t *testing.T) {
	gated := 0
	for _, th := range All() {
		if th.Gated() {
			gated++
		}
		if th.AlwaysAvailable() {
			if th.UnlockBadge != "" || th.UnlockHint.Plain() != "" || th.Gated() {
				t.Errorf("theme %q is always available but declares an unlock (badge %q, hint %q)", th.Key, th.UnlockBadge, th.UnlockHint.Plain())
			}
			if th.Effect != "" {
				t.Errorf("theme %q is always available and carries the effect %q: an accessible theme holds still", th.Key, th.Effect)
			}
			continue
		}
		if th.UnlockBadge == "" || !th.Gated() {
			t.Errorf("gated theme %q names no badge that gives it", th.Key)
		}
		if th.UnlockHint.Plain() == "" {
			t.Errorf("gated theme %q has an empty UnlockHint; a locked theme must explain how to unlock it", th.Key)
		}
		switch th.Effect {
		case "", EffectRain, EffectGlitch, EffectEmbers, EffectGreenbar, EffectPrism:
		default:
			t.Errorf("theme %q carries the effect %q, which the UI does not draw", th.Key, th.Effect)
		}
	}
	if gated != 10 {
		t.Errorf("%d gated themes, want the 10 flavor themes", gated)
	}
}
