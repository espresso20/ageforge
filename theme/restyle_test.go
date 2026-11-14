package theme

import (
	"testing"
)

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

// TestRegistry_DefaultFirst asserts the default theme is registered first and an
// unknown key is not found.
func TestRegistry_DefaultFirst(t *testing.T) {
	all := All()
	if len(all) == 0 || all[0].Key != DefaultKey {
		t.Errorf("first registered theme is not the default %q", DefaultKey)
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
	if Of(RoleBackground).Hex() != HighContrast.Color(RoleBackground).Hex() {
		t.Errorf("Of(RoleBackground) did not follow the active theme")
	}

	if err := SetActive("nope"); err == nil {
		t.Errorf("SetActive(unknown) returned nil error")
	}
	if Active().Key != "high_contrast" {
		t.Errorf("failed SetActive mutated the active theme to %q", Active().Key)
	}
}

// TestTag emits a late-bound NAMED tag for a role (never a frozen hex literal),
// and that name resolves to the active theme's role color when it is painted.
func TestTag(t *testing.T) {
	t.Cleanup(func() { _ = SetActive(DefaultKey) })
	_ = SetActive(DefaultKey)
	if got, want := Tag(RoleAccent), "[accent]"; got != want {
		t.Errorf("Tag(RoleAccent) = %q, want %q", got, want)
	}
	role, ok := TagRole("accent")
	if !ok || role != RoleAccent {
		t.Fatalf("the name accent is role %v (known %v), want Accent", role, ok)
	}
	hex := func() int32 {
		c, _ := Ref(role).RGBA()
		return int32(c.R)<<16 | int32(c.G)<<8 | int32(c.B)
	}
	// Forge accent is gold #ffd700.
	if got := hex(); got != 0xffd700 {
		t.Errorf("accent paints %06x under Forge, want ffd700", got)
	}
	_ = SetActive("parchment")
	if got, want := hex(), Parchment.Color(RoleAccent).Hex(); got != want {
		t.Errorf("accent paints %06x under Parchment, want %06x (the tag must retint)", got, want)
	}
}
