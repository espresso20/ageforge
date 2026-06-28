package game

import (
	"testing"
)

// TestExportImportRoundTripRestoresWipedUnlocks covers the headline §3.6 promise:
// export a backup, lose the unlocks locally, then import(merge) puts them back.
func TestExportImportRoundTripRestoresWipedUnlocks(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}

	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Wipe the unlocks on the live account, then re-import the backup.
	acct.Unlocks.Themes = nil
	if acct.HasTheme("amber_crt") || acct.HasTheme("obsidian") {
		t.Fatal("expected unlocks wiped before import")
	}

	if err := acct.ImportProgress(blob, true); err != nil {
		t.Fatalf("ImportProgress(merge): %v", err)
	}
	if !acct.HasTheme("amber_crt") || !acct.HasTheme("obsidian") {
		t.Fatalf("expected both themes restored, got %v", acct.UnlockedThemes())
	}
}

// TestImportMergeDoesNotDropNewerUnlocks covers the "re-importing an old backup must
// not remove a theme you've since unlocked" invariant: export {A}, add {B} locally,
// import(merge) → {A,B}.
func TestImportMergeDoesNotDropNewerUnlocks(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}

	blob, err := acct.ExportProgress() // backup carries only {A}
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Locally unlock a newer theme the backup predates.
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}

	if err := acct.ImportProgress(blob, true); err != nil {
		t.Fatalf("ImportProgress(merge): %v", err)
	}
	got := acct.UnlockedThemes() // sorted
	if len(got) != 2 || got[0] != "amber_crt" || got[1] != "obsidian" {
		t.Fatalf("expected union {amber_crt, obsidian}, got %v", got)
	}
}

// TestImportReplaceWholesale covers merge=false: the DATA fields are replaced by the
// blob's, so a local-only theme not in the blob is dropped.
func TestImportReplaceWholesale(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}

	blob, err := acct.ExportProgress() // backup carries {amber_crt}
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Local-only theme that the blob does NOT have.
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}

	if err := acct.ImportProgress(blob, false); err != nil {
		t.Fatalf("ImportProgress(replace): %v", err)
	}
	got := acct.UnlockedThemes()
	if len(got) != 1 || got[0] != "amber_crt" {
		t.Fatalf("expected wholesale replace to {amber_crt}, got %v", got)
	}
	if acct.HasTheme("obsidian") {
		t.Fatal("replace should have dropped the local-only theme")
	}
}

// TestImportMergeMaxesNumericStats covers the lifetime-best rule: merge takes the MAX
// per numeric stat so bests never regress, even when the blob's value is lower.
func TestImportMergeMaxesNumericStats(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	acct.Stats.TotalPrestiges = 5
	acct.Stats.CivilizationsStarted = 3

	blob, err := acct.ExportProgress() // backup: prestiges=5, civs=3
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Local has progressed past the backup on one stat, fallen on another scenario
	// is impossible for monotonic counters; cover both directions explicitly:
	acct.Stats.TotalPrestiges = 9       // local higher than blob (5) → keep 9
	acct.Stats.CivilizationsStarted = 1 // local lower than blob (3) → take blob's 3

	if err := acct.ImportProgress(blob, true); err != nil {
		t.Fatalf("ImportProgress(merge): %v", err)
	}
	if acct.Stats.TotalPrestiges != 9 {
		t.Fatalf("expected max(9,5)=9 prestiges, got %d", acct.Stats.TotalPrestiges)
	}
	if acct.Stats.CivilizationsStarted != 3 {
		t.Fatalf("expected max(1,3)=3 civilizations, got %d", acct.Stats.CivilizationsStarted)
	}
}

// TestImportTamperGuard covers the integrity guard: a single flipped byte in the blob
// must make ImportProgress return an error AND leave the account unchanged.
func TestImportTamperGuard(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}

	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Flip a byte in the data region (the theme key 'a' in "amber_crt") so the payload
	// no longer matches the signature.
	tampered := make([]byte, len(blob))
	copy(tampered, blob)
	idx := indexOfByte(tampered, 'a')
	if idx < 0 {
		t.Fatal("test fixture: expected an 'a' byte to flip in the blob")
	}
	tampered[idx] = 'z'

	before := acct.UnlockedThemes()
	if err := acct.ImportProgress(tampered, true); err == nil {
		t.Fatal("expected ImportProgress to reject a tampered blob, got nil error")
	}
	after := acct.UnlockedThemes()
	if len(after) != len(before) || (len(after) == 1 && after[0] != before[0]) {
		t.Fatalf("tampered import must not change the account: before %v, after %v", before, after)
	}
}

// indexOfByte returns the index of the first occurrence of b in data, or -1.
func indexOfByte(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}
	return -1
}
