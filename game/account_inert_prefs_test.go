package game

import (
	"os"
	"path/filepath"
	"testing"
)

// inertPrefsAccountJSON is an account.json written while the map had a style
// and a zoom: its prefs name both. Its signature was made by the game that
// had them, and covers the prefs as written.
const inertPrefsAccountJSON = `{
  "version": 1,
  "account_id": "f16d05ec6b29248d2c61adb1e9263f78",
  "display_name": "Fixture",
  "created": "2026-09-01T12:00:00Z",
  "last_seen": "2026-09-01T12:00:00Z",
  "unlocks": {
    "themes": [
      "amber_crt"
    ]
  },
  "stats": {
    "total_prestiges": 2,
    "highest_age": "modern_age"
  },
  "achievements": [
    "first_prestige"
  ],
  "prefs": {
    "active_theme": "amber_crt",
    "map_style": "skyline",
    "map_glyphs": "ascii",
    "map_zoom": "far"
  },
  "_sig": "a7fc20bd1c2de94d64cfeb61e1d7f72a06dd793c84c4d4ffa6f1e60692505a4f"
}`

// TestAccountWithAMapStyleStillVerifies: nothing reads the map style and the
// map zoom any more, but the signature covers Prefs as written, so an
// account that holds them must still verify. Dropping the fields from
// AccountPrefs would flag every such account as modified.
func TestAccountWithAMapStyleStillVerifies(t *testing.T) {
	isolateAccountDir(t)
	noErr(t, os.MkdirAll(accountDir(fixtureAccountID), 0755))
	path := filepath.Join(accountDir(fixtureAccountID), accountFileName)
	noErr(t, os.WriteFile(path, []byte(inertPrefsAccountJSON), 0644))
	acct := slotAccount(t, fixtureAccountID)
	if acct.Prefs.MapStyle != "skyline" || acct.Prefs.MapZoom != "far" {
		t.Fatalf("the prefs lost the style or the zoom: %+v", acct.Prefs)
	}
	if acct.Tampered || !verifyAccount(acct) {
		t.Errorf("an account with a map style and zoom reads as modified (tampered=%v)", acct.Tampered)
	}
}
