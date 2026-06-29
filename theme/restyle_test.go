package theme

import "testing"

// TestTrack_AppliesNowAndEnrolls verifies the §3.3 registration discipline: Track
// runs the closure immediately AND re-runs it on Restyle, in one registration.
func TestTrack_AppliesNowAndEnrolls(t *testing.T) {
	resetRestylersForTest()
	t.Cleanup(resetRestylersForTest)

	calls := 0
	Track(func() { calls++ })
	if calls != 1 {
		t.Fatalf("Track did not apply immediately: calls = %d, want 1", calls)
	}

	Restyle()
	if calls != 2 {
		t.Fatalf("Restyle did not re-run tracked closure: calls = %d, want 2", calls)
	}

	Restyle()
	if calls != 3 {
		t.Fatalf("second Restyle: calls = %d, want 3", calls)
	}
}

func TestTrack_NilIsIgnored(t *testing.T) {
	resetRestylersForTest()
	t.Cleanup(resetRestylersForTest)
	Track(nil) // must not panic
	Restyle()  // must not panic on the nil
}

// TestRegistry_FourBuiltinThemes asserts the four Phase-1a themes registered and
// that Forge is the default/first.
func TestRegistry_FourBuiltinThemes(t *testing.T) {
	all := All()
	if len(all) < 4 {
		t.Fatalf("expected >= 4 registered themes, got %d", len(all))
	}
	if all[0].Key != DefaultKey {
		t.Errorf("first registered theme = %q, want default %q", all[0].Key, DefaultKey)
	}
	for _, key := range []string{"forge", "deuteranopia", "protanopia", "high_contrast"} {
		if _, ok := ByKey(key); !ok {
			t.Errorf("theme %q not registered", key)
		}
	}
	if _, ok := ByKey("does_not_exist"); ok {
		t.Errorf("ByKey returned ok for an unknown key")
	}
}

// TestSetActive verifies switching, the unknown-key error, and that Active reflects
// the switch.
func TestSetActive(t *testing.T) {
	t.Cleanup(func() { _ = SetActive(DefaultKey) })

	if err := SetActive("high_contrast"); err != nil {
		t.Fatalf("SetActive(high_contrast): %v", err)
	}
	if Active().Key != "high_contrast" {
		t.Errorf("Active().Key = %q, want high_contrast", Active().Key)
	}
	if Color(RoleBackground).Hex() != HighContrast.Color(RoleBackground).Hex() {
		t.Errorf("Color(RoleBackground) did not follow the active theme")
	}

	if err := SetActive("nope"); err == nil {
		t.Errorf("SetActive(unknown) returned nil error")
	}
	if Active().Key != "high_contrast" {
		t.Errorf("failed SetActive mutated the active theme to %q", Active().Key)
	}
}

// TestTag emits a hex tag for the active theme's role color.
func TestTag(t *testing.T) {
	t.Cleanup(func() { _ = SetActive(DefaultKey) })
	_ = SetActive(DefaultKey)
	// Forge accent is gold #ffd700.
	if got, want := Tag(RoleAccent), "[#ffd700]"; got != want {
		t.Errorf("Tag(RoleAccent) = %q, want %q", got, want)
	}
}
