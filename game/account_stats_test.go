package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// hasAchievement reports whether key is present in the account's achievement set.
// A small test helper so each assertion reads as intent rather than a loop.
func hasAchievement(a *Account, key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.hasAchievementLocked(key)
}

// TestRecordPrestigeIncrementsAndUnlocks covers the prestige hook: the lifetime count
// climbs with each call, "first_prestige" unlocks at 1, and the higher "prestige_x10"
// tier unlocks only once the threshold is crossed — never before.
func TestRecordPrestigeIncrementsAndUnlocks(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	acct.RecordPrestige()
	if acct.Stats.TotalPrestiges != 1 {
		t.Fatalf("TotalPrestiges = %d, want 1", acct.Stats.TotalPrestiges)
	}
	if !hasAchievement(acct, "first_prestige") {
		t.Fatalf("expected first_prestige unlocked at 1 prestige")
	}
	if hasAchievement(acct, "prestige_x10") {
		t.Fatalf("prestige_x10 unlocked too early (1 prestige)")
	}
	if !acct.dirty {
		t.Fatalf("expected dirty=true after RecordPrestige")
	}

	// Climb to 10; prestige_x10 fires exactly at the threshold, not before.
	for acct.Stats.TotalPrestiges < 9 {
		acct.RecordPrestige()
		if hasAchievement(acct, "prestige_x10") {
			t.Fatalf("prestige_x10 unlocked early at %d prestiges", acct.Stats.TotalPrestiges)
		}
	}
	acct.RecordPrestige() // -> 10
	if acct.Stats.TotalPrestiges != 10 {
		t.Fatalf("TotalPrestiges = %d, want 10", acct.Stats.TotalPrestiges)
	}
	if !hasAchievement(acct, "prestige_x10") {
		t.Fatalf("expected prestige_x10 unlocked at 10 prestiges")
	}
}

// TestRecordAgeReachedRanksByOrder covers the age hook: HighestAge advances only when
// a strictly higher order is reached, a lower order does NOT regress it, and the age
// achievements fire at their threshold orders.
func TestRecordAgeReachedRanksByOrder(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	ages := config.AgeByKey()
	iron := ages["iron_age"]     // Order 3
	bronze := ages["bronze_age"] // Order 2
	modern := ages["modern_age"] // Order 12

	acct.RecordAgeReached(iron.Key, iron.Order)
	if acct.Stats.HighestAge != "iron_age" {
		t.Fatalf("HighestAge = %q, want iron_age", acct.Stats.HighestAge)
	}
	if !hasAchievement(acct, "reached_iron") {
		t.Fatalf("expected reached_iron unlocked at iron_age")
	}
	if hasAchievement(acct, "reached_modern") {
		t.Fatalf("reached_modern unlocked too early (iron_age)")
	}

	// A LOWER age must not regress the lifetime best.
	acct.RecordAgeReached(bronze.Key, bronze.Order)
	if acct.Stats.HighestAge != "iron_age" {
		t.Fatalf("HighestAge regressed to %q after lower age; want iron_age", acct.Stats.HighestAge)
	}

	// A higher age advances it and unlocks the modern achievement.
	acct.RecordAgeReached(modern.Key, modern.Order)
	if acct.Stats.HighestAge != "modern_age" {
		t.Fatalf("HighestAge = %q, want modern_age", acct.Stats.HighestAge)
	}
	if !hasAchievement(acct, "reached_modern") {
		t.Fatalf("expected reached_modern unlocked at modern_age")
	}
}

// TestFlushIfDirtyPersistsAndClears covers the write-debounce lifecycle: a Record* call
// marks the account dirty; FlushIfDirty Saves and clears the flag; a fresh LoadOrCreate
// reflects the persisted stats (survives a "restart"); a flush when clean is a no-op.
func TestFlushIfDirtyPersistsAndClears(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	acct.RecordPrestige()
	acct.RecordAgeReached("iron_age", config.AgeByKey()["iron_age"].Order)
	if !acct.dirty {
		t.Fatalf("expected dirty after Record* calls")
	}

	// This call must NOT hang — FlushIfDirty holds a.mu across Save(), which does not
	// re-acquire a.mu (self-deadlock guard). If the lock discipline were wrong this
	// test would deadlock rather than fail.
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatalf("FlushIfDirty: %v", err)
	}
	if acct.dirty {
		t.Fatalf("expected dirty cleared after FlushIfDirty")
	}

	// A second flush when clean is a pure no-op (no error, stays clean).
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatalf("FlushIfDirty (clean no-op): %v", err)
	}

	// Persisted across a "restart": a fresh LoadOrCreate reads the same data root.
	reloaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (reload): %v", err)
	}
	if reloaded.Stats.TotalPrestiges != 1 {
		t.Fatalf("reloaded TotalPrestiges = %d, want 1", reloaded.Stats.TotalPrestiges)
	}
	if reloaded.Stats.HighestAge != "iron_age" {
		t.Fatalf("reloaded HighestAge = %q, want iron_age", reloaded.Stats.HighestAge)
	}
	if !hasAchievement(reloaded, "first_prestige") || !hasAchievement(reloaded, "reached_iron") {
		t.Fatalf("reloaded account missing expected achievements: %v", reloaded.Achievements)
	}
	if reloaded.Tampered {
		t.Fatalf("reloaded account flagged tampered — sign/round-trip broke")
	}
}

// TestLifetimeStatsReturnsCopy covers the snapshot accessor: it returns the current
// stats plus a COPY of the achievements slice that the caller cannot use to mutate the
// account's backing state.
func TestLifetimeStatsReturnsCopy(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	acct.RecordPrestige()

	stats, ach := acct.LifetimeStats()
	if stats.TotalPrestiges != 1 {
		t.Fatalf("LifetimeStats TotalPrestiges = %d, want 1", stats.TotalPrestiges)
	}
	if len(ach) == 0 {
		t.Fatalf("expected at least first_prestige in returned achievements")
	}
	// Mutating the returned slice must not affect the account.
	ach[0] = "tampered_key"
	if hasAchievement(acct, "tampered_key") {
		t.Fatalf("LifetimeStats leaked its backing slice — caller mutation reached the account")
	}
}
