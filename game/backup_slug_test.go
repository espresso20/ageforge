package game

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBackupNameSlug pins how a display name becomes part of a backup folder name: letters,
// digits, dashes and underscores stay, anything else is a dash, runs of dashes become one,
// and none is left at either end.
func TestBackupNameSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Ada Lovelace":     "Ada-Lovelace",
		"  spaced   out  ": "spaced-out",
		"a/b\\c..d":        "a-b-c-d",
		"--x--":            "x",
		"日本":               "",
		"snake_case-ok":    "snake_case-ok",
		"é x":              "x",
		"":                 "",
		"  ":               "",
		"Bob's Camp!":      "Bob-s-Camp",
		"a - b":            "a-b",
		"x\xffy":           "x-y",
	} {
		if got := backupNameSlug(in); got != want {
			t.Errorf("backupNameSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestBackupAccountCopiesSettingsWhenThereAreSome covers the settings file riding along
// beside account.json, and a slot without one backing up fine.
func TestBackupAccountCopiesSettingsWhenThereAreSome(t *testing.T) {
	isolateAccountDir(t)
	acct, err := CreateAccount("Setter")
	if err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(accountDir(acct.AccountID), settingsFileName)
	os.Remove(settings)

	without, err := BackupAccount(acct.AccountID)
	if err != nil {
		t.Fatalf("BackupAccount without settings: %v", err)
	}
	if _, err := os.Stat(filepath.Join(without, settingsFileName)); !os.IsNotExist(err) {
		t.Errorf("a backup of a slot with no settings file has one: %v", err)
	}

	const body = `{"version": 1}`
	if err := os.WriteFile(settings, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	with, err := BackupAccount(acct.AccountID)
	if err != nil {
		t.Fatalf("BackupAccount with settings: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(with, settingsFileName)); err != nil || string(got) != body {
		t.Errorf("the backup's settings file = %q, %v; want %q", got, err, body)
	}
}
