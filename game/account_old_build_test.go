package game

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Two builds on one machine: a build of the game from before badges and
// this one, taking turns with the same account. The older build must never
// be able to mark a healthy account as modified, and nothing it does may
// cost the account a badge.
//
// The fixtures under testdata/account_split are a real round trip:
//
//   - new/ was written by this code: account.json, the badge file beside
//     it, and the account's export.
//   - after_old_resave/account.json was written by master's own code from
//     before badges (commit 310f3a5, run from a copy of that tree against a
//     temp data root holding new/): it loaded the account and changed a
//     setting. That run itself failed if the old code read the account as
//     modified or changed or removed the badge file.
//
// These tests are the other half: this code loading what the old code left.
// Every test copies fixtures into a temp data root (isolateAccountDir) and
// never touches data/.

const splitDir = "testdata/account_split"

// installSplit puts an account.json fixture and the new-layout badge file
// into a slot of the temp data root, and returns the ID and both files'
// bytes.
func installSplit(t *testing.T, accountFile string) (id string, accountRaw, badgesRaw []byte) {
	t.Helper()
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(splitDir, name))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		return b
	}
	accountRaw = read(accountFile)
	var head struct {
		AccountID string `json:"account_id"`
	}
	if err := json.Unmarshal(accountRaw, &head); err != nil {
		t.Fatalf("fixture %s: %v", accountFile, err)
	}
	id = head.AccountID
	noErr(t, os.MkdirAll(accountDir(id), 0755))
	noErr(t, os.WriteFile(filepath.Join(accountDir(id), accountFileName), accountRaw, 0644))
	badgesRaw = read("new/badges.json")
	noErr(t, os.WriteFile(filepath.Join(accountDir(id), badgeFileName), badgesRaw, 0644))
	return id, accountRaw, badgesRaw
}

// storedBadges parses a badge file's bytes.
func storedBadges(t *testing.T, raw []byte) badgeFile {
	t.Helper()
	var bf badgeFile
	if err := json.Unmarshal(raw, &bf); err != nil {
		t.Fatalf("badge file: %v", err)
	}
	return bf
}

// TestNewLayoutFixtureIsCurrent: the new/ fixture is what this code still
// writes and reads. If the layout changes, the round trip must be run again
// with the old code, not patched by hand.
func TestNewLayoutFixtureIsCurrent(t *testing.T) {
	isolateAccountDir(t)
	id, accountRaw, badgesRaw := installSplit(t, "new/account.json")
	assertOldShape(t, accountRaw)
	acct := slotAccount(t, id)
	if acct.Tampered || acct.BadgesTampered {
		t.Fatalf("the fixture loads flagged: %v %v", acct.Tampered, acct.BadgesTampered)
	}
	// Saving it writes the same bytes.
	acct.mu.Lock()
	err := acct.Save()
	acct.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if string(slotFile(t, id)) != string(accountRaw) || string(slotBadgeFile(t, id)) != string(badgesRaw) {
		t.Error("this code writes the fixture account differently: run the round trip again with the old code")
	}
}

// TestOldBuildResaveKeepsBadges: the older build loaded the account,
// changed a setting and saved it, never seeing the badge file. Loaded here
// afterwards the account is unflagged and every badge is still earned and
// dated as before. The catalog has grown since the fixture was written, so
// the first load also grants what the record proves under it (the Bronze
// Age, between the two ages the fixture holds), undated. account.json is
// not rewritten, and a second load rewrites neither file.
func TestOldBuildResaveKeepsBadges(t *testing.T) {
	isolateAccountDir(t)
	id, accountRaw, badgesRaw := installSplit(t, "after_old_resave/account.json")
	before := storedBadges(t, badgesRaw)
	if len(before.Badges) < 4 {
		t.Fatalf("precondition: the fixture holds %d badges", len(before.Badges))
	}
	const picked = "age.bronze_age"
	wantBadges := maps.Clone(before.Badges)
	wantBadges[picked] = BadgeEarned{}
	wantCounters := maps.Clone(before.Counters)
	wantCounters[config.BadgeEvBadge] = float64(len(wantBadges))

	var firstBadges []byte
	for load := 1; load <= 2; load++ {
		ge, acct := openIn(t, id)
		if acct.Tampered || acct.BadgesTampered {
			t.Fatalf("load %d: the account reads as modified after the older build saved it (account %v, badge file %v)", load, acct.Tampered, acct.BadgesTampered)
		}
		if !reflect.DeepEqual(withoutCounts(acct.Badges), wantBadges) {
			t.Errorf("load %d: badges\n got %+v\nwant %+v", load, acct.Badges, wantBadges)
		}
		for key, e := range acct.Badges {
			if key == picked {
				continue
			}
			if e.At == 0 || e.Run == "" || e.Flags != 0 {
				t.Errorf("load %d: %s is %+v, want dated, with its run, uncrossed", load, key, e)
			}
		}
		if !reflect.DeepEqual(acct.Counters, wantCounters) || !slices.Equal(acct.Days, before.Days) {
			t.Errorf("load %d: counters %v days %v, want %v %v", load, acct.Counters, acct.Days, wantCounters, before.Days)
		}
		// What the older build changed is there.
		if glyphs, _ := acct.MapPrefs(); glyphs != "ascii" {
			t.Errorf("load %d: the older build's setting was lost: map glyphs %q", load, glyphs)
		}
		if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
			t.Errorf("load %d announced badges: %+v", load, earned)
		}
		noErr(t, acct.FlushIfDirty())
		if got := slotFile(t, id); string(got) != string(accountRaw) {
			t.Errorf("load %d rewrote account.json:\nwas %s\nnow %s", load, accountRaw, got)
		}
		got := slotBadgeFile(t, id)
		if load == 1 {
			firstBadges = got
			if bf := storedBadges(t, got); !verifyBadgeFile(&bf) || bf.Tampered {
				t.Errorf("the badge file after the first load: verifies %v, tampered %v", verifyBadgeFile(&bf), bf.Tampered)
			}
			continue
		}
		if string(got) != string(firstBadges) {
			t.Errorf("load %d rewrote the badge file:\nwas %s\nnow %s", load, firstBadges, got)
		}
	}
}

// withoutCounts is badges with the remembered counts left out: a rung from
// before they were kept settles by remembering today's bar, which these
// tests are not about.
func withoutCounts(badges map[string]BadgeEarned) map[string]BadgeEarned {
	out := make(map[string]BadgeEarned, len(badges))
	for k, e := range badges {
		e.Count = 0
		out[k] = e
	}
	return out
}
