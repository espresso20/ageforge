package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/rules"
)

// The age themes move from milestones to the badges of their ages. No
// account may lose a theme it holds, and an account that reached an age
// gets that age's theme.
//
// The fixtures under testdata/account_themes were written by master's own
// code from before the move (commit a58c4ee, run from a copy of that tree
// against a temp data root): accounts that played to an age through the
// engine, with themes unlocked the way that build's dashboard unlocked them
// (its evaluateThemeUnlock, a milestone at a time), saved with the run they
// played. Every test copies a fixture into a temp data root
// (isolateAccountDir) and never touches data/.
//
//   - milestones: reached the Information Age; Bronze, Parchment and
//     Monochrome unlocked on the way; Monochrome worn.
//   - ahead: holds Bronze, Cyberpunk and Cosmic with a record that stops at
//     the Bronze Age (themes an import merged in); Cosmic worn.
//   - reached: reached the Galactic Age and holds no theme (that build
//     unlocked themes from the dashboard, which never ran).
//   - fresh: a new account that has gone nowhere.

const themesDir = "testdata/account_themes"

// installThemes copies a theme fixture into its slot of the temp data
// root: account.json, the badge file and the run it played.
func installThemes(t *testing.T, name string) (id string, accountRaw, badgesRaw []byte) {
	t.Helper()
	read := func(file string) []byte {
		b, err := os.ReadFile(filepath.Join(themesDir, name, file))
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		return b
	}
	accountRaw, badgesRaw = read("account.json"), read("badges.json")
	var head struct {
		AccountID string `json:"account_id"`
	}
	if err := json.Unmarshal(accountRaw, &head); err != nil {
		t.Fatal(err)
	}
	id = head.AccountID
	if err := os.MkdirAll(filepath.Join(accountDir(id), "saves"), 0755); err != nil {
		t.Fatal(err)
	}
	for file, data := range map[string][]byte{
		accountFileName: accountRaw, badgeFileName: badgesRaw, filepath.Join("saves", "run.json"): read("run.json"),
	} {
		if err := os.WriteFile(filepath.Join(accountDir(id), file), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return id, accountRaw, badgesRaw
}

func TestAgeThemesMoveToBadgesWithNothingLost(t *testing.T) {
	// The theme each age's badge gives, from the ruleset.
	themeOfAge := map[string]string{}
	for _, def := range rules.Core().Badges() {
		if def.Family == "age" && def.Reward.Theme != "" {
			themeOfAge[def.Subject] = def.Reward.Theme
		}
	}
	if len(themeOfAge) != 5 {
		t.Fatalf("the age badges give %d themes, want the five age themes: %v", len(themeOfAge), themeOfAge)
	}
	cases := []struct {
		name   string
		reach  string   // the highest age on the record ("" for none)
		had    []string // the themes the fixture holds
		want   []string // the themes it holds afterwards
		active string
		// lacks: age badges the account must not be handed (its record
		// does not reach them), though it holds their themes.
		lacks []string
	}{
		{name: "milestones", reach: "information_age", had: []string{"bronze", "monochrome", "parchment"},
			want: []string{"bronze", "monochrome", "parchment"}, active: "monochrome"},
		{name: "ahead", reach: "bronze_age", had: []string{"bronze", "cosmic", "cyberpunk"},
			want: []string{"bronze", "cosmic", "cyberpunk"}, active: "cosmic",
			lacks: []string{"age.cyberpunk_age", "age.galactic_age"}},
		{name: "reached", reach: "galactic_age", had: nil,
			want: []string{"bronze", "cosmic", "cyberpunk", "monochrome", "parchment"}},
		{name: "fresh"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			isolateAccountDir(t)
			id, accountRaw, badgesRaw := installThemes(t, tc.name)
			var old Account
			if err := json.Unmarshal(accountRaw, &old); err != nil {
				t.Fatal(err)
			}
			if !verifyAccount(&old) || old.Stats.HighestAge != tc.reach || old.Prefs.ActiveTheme != tc.active {
				t.Fatalf("precondition: the fixture verifies %v, reached %q, wears %q", verifyAccount(&old), old.Stats.HighestAge, old.Prefs.ActiveTheme)
			}
			if got := sortedUnion(old.Unlocks.Themes, nil); !slices.Equal(got, sortedUnion(tc.had, nil)) {
				t.Fatalf("precondition: the fixture holds the themes %v, want %v", got, tc.had)
			}
			before := storedBadges(t, badgesRaw)

			ge, acct := openIn(t, id)
			if acct.Tampered || acct.BadgesTampered {
				t.Fatalf("the account reads as modified (account %v, badge file %v)", acct.Tampered, acct.BadgesTampered)
			}
			// Nothing lost: every theme it held, and the one it wears.
			for _, th := range tc.had {
				if !acct.HasTheme(th) {
					t.Errorf("the account lost the theme %s", th)
				}
			}
			if acct.ActiveTheme() != tc.active {
				t.Errorf("the account wears %q, was %q", acct.ActiveTheme(), tc.active)
			}
			if tc.active != "" && !acct.HasTheme(tc.active) {
				t.Errorf("the theme the account wears, %s, is locked", tc.active)
			}
			if got := acct.UnlockedThemes(); !slices.Equal(sortedUnion(got, nil), sortedUnion(tc.want, nil)) || len(got) != len(tc.want) {
				t.Errorf("themes %v, want %v", got, tc.want)
			}
			// Every age reached has its badge, and with it its theme.
			if tc.reach != "" {
				for _, key := range ageBadges(tc.reach) {
					if !hasBadge(acct, key) {
						t.Errorf("the account reached the %s and does not hold %s", tc.reach, key)
					}
					if th := themeOfAge[key[len("age."):]]; th != "" && !acct.HasTheme(th) {
						t.Errorf("%s is held and its theme %s is locked", key, th)
					}
				}
			}
			for _, key := range tc.lacks {
				if hasBadge(acct, key) {
					t.Errorf("the account was handed %s, an age its record does not reach", key)
				}
			}
			// What it had earned keeps its date and its run; what the record
			// proves comes undated; nothing is crossed or announced.
			for key, was := range before.Badges {
				if acct.Badges[key] != was {
					t.Errorf("%s changed: %+v, was %+v", key, acct.Badges[key], was)
				}
			}
			for key, e := range acct.Badges {
				if _, held := before.Badges[key]; !held && (e.At != 0 || e.Run != "" || e.Flags != 0) {
					t.Errorf("%s, granted from the record, is %+v; want undated and uncrossed", key, e)
				}
			}
			if earned := ge.DrainEarnedBadges(); len(earned) != 0 {
				t.Errorf("the move announced badges: %v", earnedKeysOf(earned))
			}

			// account.json keeps its shape and still verifies; it changes only
			// where a theme was gained, and then only in its unlocks.
			first := slotFile(t, id)
			assertOldShape(t, first)
			if disk := slotAccount(t, id); !verifyAccount(disk) || disk.Tampered || disk.BadgesTampered {
				t.Errorf("on disk: verifies %v, tampered %v, badge file tampered %v", verifyAccount(disk), disk.Tampered, disk.BadgesTampered)
			}
			if len(tc.want) == len(tc.had) {
				if string(first) != string(accountRaw) {
					t.Errorf("no theme was gained and account.json was rewritten:\nwas %s\nnow %s", accountRaw, first)
				}
			} else {
				assertOnlyUnlocksDiffer(t, accountRaw, first)
			}
			if bf := storedBadges(t, slotBadgeFile(t, id)); !verifyBadgeFile(&bf) || bf.Tampered {
				t.Errorf("the badge file: verifies %v, tampered %v", verifyBadgeFile(&bf), bf.Tampered)
			}

			// The run it played loads and plays on. It is where it already
			// was, so no age is told again and no theme is announced; what
			// the ticks themselves bring (a return, a civilization met) is
			// earned like anything else.
			setActiveAccountID(id) // saves are read from the active account's slot
			if err := ge.LoadGame("run"); err != nil {
				t.Fatalf("the fixture's run does not load: %v", err)
			}
			themesBefore := acct.UnlockedThemes()
			ge.StepTicks(2)
			for _, v := range ge.DrainEarnedBadges() {
				if v.Family == "age" || v.RewardTheme != "" {
					t.Errorf("loading the run announced %s (theme %q)", v.Key, v.RewardTheme)
				}
			}
			if got := acct.UnlockedThemes(); !slices.Equal(got, themesBefore) {
				t.Errorf("loading the run changed the themes: %v, were %v", got, themesBefore)
			}
			if err := acct.FlushIfDirty(); err != nil {
				t.Fatal(err)
			}

			// A second load changes nothing, in either file.
			_, again := openIn(t, id)
			if err := again.FlushIfDirty(); err != nil {
				t.Fatal(err)
			}
			if second := slotFile(t, id); string(second) != string(first) {
				t.Errorf("the second load rewrote account.json:\nfirst  %s\nsecond %s", first, second)
			}
			if !slices.Equal(again.EarnedBadges(), acct.EarnedBadges()) {
				t.Errorf("the second load changed the badges: %v then %v", acct.EarnedBadges(), again.EarnedBadges())
			}
			if got := again.UnlockedThemes(); !slices.Equal(sortedUnion(got, nil), sortedUnion(tc.want, nil)) {
				t.Errorf("themes after the second load %v, want %v", got, tc.want)
			}
		})
	}
}

// TestAgeBadgeUnlocksItsThemeInPlay: after the move a theme comes the
// moment its age's badge is earned, once, for the account the run belongs
// to and for no other.
func TestAgeBadgeUnlocksItsThemeInPlay(t *testing.T) {
	isolateAccountDir(t)
	ge, acct := devEngine(t, "Ada")
	advanceTo(ge, "stone_age")
	if acct.HasTheme("bronze") {
		t.Fatal("Bronze is unlocked before the Bronze Age")
	}
	advanceTo(ge, "bronze_age")
	if !acct.HasTheme("bronze") {
		t.Fatal("reaching the Bronze Age did not unlock Bronze")
	}
	var said []string
	for _, v := range ge.DrainEarnedBadges() {
		if v.RewardTheme != "" {
			said = append(said, v.RewardTheme)
		}
	}
	if !slices.Equal(said, []string{"bronze"}) {
		t.Errorf("badges that announce a theme: %v, want only Bronze", said)
	}
	advanceTo(ge, "bronze_age") // a later run reaching it again
	if got := acct.UnlockedThemes(); !slices.Equal(got, []string{"bronze"}) {
		t.Errorf("themes after reaching the age twice: %v", got)
	}
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	assertOldShape(t, slotFile(t, acct.AccountID))

	// Another account takes over the engine while the first one's run is
	// still in memory: that run earns the newcomer nothing.
	bea, err := CreateAccount("Bea")
	if err != nil {
		t.Fatal(err)
	}
	ge.SetAccount(bea)
	advanceTo(ge, "iron_age")
	if bea.HasTheme("bronze") || len(bea.EarnedBadges()) != 0 {
		t.Errorf("another account's run gave Bea %v and the themes %v", bea.EarnedBadges(), bea.UnlockedThemes())
	}
}
