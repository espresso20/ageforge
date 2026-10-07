package game

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestRecordPrestigeIncrementsAndUnlocks covers the prestige record: the lifetime count
// climbs with each prestige, the first rung of the prestige ladder is earned at 1, and
// the rung at 10 (the old prestige_x10) only once the threshold is crossed, never before.
func TestRecordPrestigeIncrementsAndUnlocks(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	recordPrestige(acct, "modern_age")
	if acct.Stats.TotalPrestiges != 1 {
		t.Fatalf("TotalPrestiges = %d, want 1", acct.Stats.TotalPrestiges)
	}
	if !hasBadge(acct, badgePrestige1) {
		t.Fatalf("expected %s earned at 1 prestige", badgePrestige1)
	}
	if hasBadge(acct, badgePrestige10) {
		t.Fatalf("%s earned too early (1 prestige)", badgePrestige10)
	}
	if !acct.dirty {
		t.Fatalf("expected dirty=true after a prestige")
	}

	// Climb to 10; the rung is earned exactly at the threshold, not before.
	for acct.Stats.TotalPrestiges < 9 {
		recordPrestige(acct, "modern_age")
		if hasBadge(acct, badgePrestige10) {
			t.Fatalf("%s earned early at %d prestiges", badgePrestige10, acct.Stats.TotalPrestiges)
		}
	}
	if !hasBadge(acct, badgePrestige3) {
		t.Errorf("%s not earned by 9 prestiges", badgePrestige3)
	}
	recordPrestige(acct, "modern_age") // -> 10
	if acct.Stats.TotalPrestiges != 10 {
		t.Fatalf("TotalPrestiges = %d, want 10", acct.Stats.TotalPrestiges)
	}
	if !hasBadge(acct, badgePrestige10) {
		t.Fatalf("expected %s earned at 10 prestiges", badgePrestige10)
	}
	if got := acct.Counters[config.BadgeEvPrestige]; got != 10 {
		t.Errorf("the prestige counter is %v, want 10", got)
	}
	// The two rungs that were account achievements keep their place in that list,
	// which is what a build from before badges shows.
	if !slices.Equal(acct.Achievements, []string{"first_prestige", "prestige_x10"}) {
		t.Errorf("the achievements list is %v, want first_prestige and prestige_x10", acct.Achievements)
	}
}

// TestRecordAgeReachedRanksByOrder covers the age record: HighestAge advances only when
// a strictly higher order is reached, a lower order does NOT regress it, and each age's
// badge is earned on reaching that age.
func TestRecordAgeReachedRanksByOrder(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	recordAge(t, acct, "iron_age")
	if acct.Stats.HighestAge != "iron_age" {
		t.Fatalf("HighestAge = %q, want iron_age", acct.Stats.HighestAge)
	}
	if !hasBadge(acct, badgeIron) {
		t.Fatalf("expected %s earned at the Iron Age", badgeIron)
	}
	if hasBadge(acct, badgeModern) {
		t.Fatalf("%s earned too early (the Iron Age)", badgeModern)
	}

	// A LOWER age must not regress the lifetime best.
	recordAge(t, acct, "bronze_age")
	if acct.Stats.HighestAge != "iron_age" {
		t.Fatalf("HighestAge regressed to %q after lower age; want iron_age", acct.Stats.HighestAge)
	}

	// A higher age advances it and earns the Modern Age badge.
	recordAge(t, acct, "modern_age")
	if acct.Stats.HighestAge != "modern_age" {
		t.Fatalf("HighestAge = %q, want modern_age", acct.Stats.HighestAge)
	}
	if !hasBadge(acct, badgeModern) {
		t.Fatalf("expected %s earned at the Modern Age", badgeModern)
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

	recordPrestige(acct, "modern_age")
	recordAge(t, acct, "iron_age")
	if !acct.dirty {
		t.Fatalf("expected dirty after the records")
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
	if !hasBadge(reloaded, badgePrestige1) || !hasBadge(reloaded, badgeIron) {
		t.Fatalf("reloaded account missing expected badges: %v", reloaded.Badges)
	}
	if e := reloaded.Badges[badgePrestige1]; e.At == 0 || e.Flags != 0 {
		t.Errorf("a badge earned in play came back undated or flagged: %+v", e)
	}
	if reloaded.Tampered {
		t.Fatalf("reloaded account flagged tampered — sign/round-trip broke")
	}
}

// TestLifetimeStatsReturnsCopy covers the snapshot accessor: it returns the current
// stats plus the earned badge keys in a slice of its own, which the caller cannot use to
// mutate the account's backing state.
func TestLifetimeStatsReturnsCopy(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	recordPrestige(acct, "modern_age")

	stats, earned := acct.LifetimeStats()
	if stats.TotalPrestiges != 1 {
		t.Fatalf("LifetimeStats TotalPrestiges = %d, want 1", stats.TotalPrestiges)
	}
	if len(earned) != 1 || earned[0] != badgePrestige1 {
		t.Fatalf("LifetimeStats badges = %v, want [%s]", earned, badgePrestige1)
	}
	// Mutating the returned slice must not affect the account.
	earned[0] = "tampered_key"
	if hasBadge(acct, "tampered_key") || !hasBadge(acct, badgePrestige1) {
		t.Fatalf("LifetimeStats leaked its backing data: caller mutation reached the account")
	}
}

// TestCivilizationsStartedCountsEveryRun is B12, the first half: the account's
// Civilizations Started was stored, merged and exported but never counted. A civilization
// starts with every new game and with the run that follows each prestige and each Succumb.
// Loading a game starts nothing. A run the developer console has changed counts like any
// other.
func TestCivilizationsStartedCountsEveryRun(t *testing.T) {
	isolateAccountDir(t)
	acct, err := CreateAccount("Ada")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(acct)
	started := func() int {
		s, _ := acct.LifetimeStats()
		return s.CivilizationsStarted
	}
	if started() != 0 {
		t.Fatalf("a new account has started %d civilizations", started())
	}

	if err := ge.StartNewNamedGame("first"); err != nil {
		t.Fatal(err)
	}
	if started() != 1 {
		t.Fatalf("after a new game: %d, want 1", started())
	}
	if err := ge.SaveGame("first"); err != nil {
		t.Fatal(err)
	}
	if err := ge.LoadGame("first"); err != nil {
		t.Fatal(err)
	}
	if started() != 1 {
		t.Errorf("loading a game counted a civilization: %d", started())
	}

	prestigeNow(t, ge)
	if started() != 2 {
		t.Errorf("after a prestige: %d, want 2", started())
	}

	ge.mu.Lock()
	ge.age = "iron_age"
	ge.currentEpoch = config.EpochForAge("iron_age")
	ge.pendingCatastrophe = "iron_era"
	ge.mu.Unlock()
	if err := ge.Succumb(); err != nil {
		t.Fatalf("Succumb: %v", err)
	}
	if started() != 3 {
		t.Errorf("after a Succumb: %d, want 3", started())
	}

	// It is saved with the account, in a key every build signs.
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if disk := slotAccount(t, acct.AccountID); disk.Stats.CivilizationsStarted != 3 || disk.Tampered {
		t.Errorf("on disk: %d civilizations started, tampered %v", disk.Stats.CivilizationsStarted, disk.Tampered)
	}

	// A run the developer console changed is a civilization like any other.
	withDevMode(t)
	if out := DevExecCommand("/give food 5", ge); !strings.Contains(out, "gave") {
		t.Fatalf("/give: %q", out)
	}
	prestigeNow(t, ge)
	if started() != 4 {
		t.Errorf("after a dev-touched run's prestige: %d, want 4", started())
	}
}

// TestSavesCompletedIsKeptButNotCounted is B12, the second half. Saves Completed was never
// counted and cannot be: a save is never completed (the game has no ending, and a prestige
// is Total Prestiges already). It is retired: nothing counts it and no view carries it. The
// key stays in the file format, so an account or an export that holds one still verifies
// and keeps it.
func TestSavesCompletedIsKeptButNotCounted(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	prestigeNow(t, ge)
	if s, _ := acct.LifetimeStats(); s.SavesCompleted != 0 {
		t.Errorf("something counted Saves Completed: %d", s.SavesCompleted)
	}

	// A file that carries the key (written by hand here, through the signed path) loads
	// unflagged and keeps it through a save and an export.
	acct.mu.Lock()
	acct.Stats.SavesCompleted = 2
	err := acct.Save()
	acct.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	disk := slotAccount(t, acct.AccountID)
	if disk.Tampered || disk.Stats.SavesCompleted != 2 {
		t.Errorf("an account.json with saves_completed: tampered %v, value %d", disk.Tampered, disk.Stats.SavesCompleted)
	}
	blob, err := disk.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	isolateAccountDir(t)
	back, err := ImportAccountExport(blob, true)
	if err != nil || back.Stats.SavesCompleted != 2 {
		t.Errorf("the export round trip: %v, value %d", err, back.Stats.SavesCompleted)
	}
}
