package game

import "testing"

// TestApplyExport_PrefsAndStatsMerge covers how an import folds preferences and lifetime
// stats into the account: with merge, a setting the account already has stays and an empty
// one adopts the import's, the hint flag is set if either has it, and each best is the larger
// of the two; without merge, the import's preferences and stats replace the account's.
func TestApplyExport_PrefsAndStatsMerge(t *testing.T) {
	local := func() *Account {
		return &Account{
			Prefs: AccountPrefs{ActiveTheme: "ember"},
			Stats: AccountStats{TotalPrestiges: 5, CivilizationsStarted: 2, SavesCompleted: 9,
				PrestigesByAge: map[string]int{"stone_age": 3}},
		}
	}
	exp := &progressExport{
		BadgeStore: &badgeFile{},
		Prefs:      AccountPrefs{ActiveTheme: "frost", MapGlyphs: "ascii", Minimap: "on", MapIconsHint: true},
		Stats: AccountStats{TotalPrestiges: 4, CivilizationsStarted: 7, SavesCompleted: 9,
			PrestigesByAge: map[string]int{"stone_age": 1, "bronze_age": 2}},
	}

	a := local()
	a.applyExportLocked(exp, true)
	want := AccountPrefs{ActiveTheme: "ember", MapGlyphs: "ascii", Minimap: "on", MapIconsHint: true}
	if a.Prefs != want {
		t.Errorf("merged prefs = %+v, want %+v", a.Prefs, want)
	}
	if a.Stats.TotalPrestiges != 5 || a.Stats.CivilizationsStarted != 7 || a.Stats.SavesCompleted != 9 ||
		a.Stats.PrestigesByAge["stone_age"] != 3 || a.Stats.PrestigesByAge["bronze_age"] != 2 {
		t.Errorf("merged stats = %+v, want each the larger of the two", a.Stats)
	}

	b := local()
	b.Prefs.MapIconsHint = true
	b.applyExportLocked(&progressExport{BadgeStore: &badgeFile{}}, true)
	if !b.Prefs.MapIconsHint {
		t.Error("an import without the hint flag cleared the account's")
	}

	r := local()
	r.applyExportLocked(exp, false)
	if r.Prefs != exp.Prefs || r.Stats.TotalPrestiges != 4 {
		t.Errorf("replaced: prefs %+v stats %+v, want the import's", r.Prefs, r.Stats)
	}
}

// TestAccountHasTheme covers the theme membership check on both the locked and unlocked sides.
func TestAccountHasTheme(t *testing.T) {
	a := &Account{Unlocks: AccountUnlocks{Themes: []string{"ember", "frost"}}}
	if !a.HasTheme("frost") || a.HasTheme("moss") || a.HasTheme("") {
		t.Error("HasTheme answers wrongly for a held, an absent and an empty key")
	}
	if newly, _ := (&Account{AccountID: "not an id", Unlocks: a.Unlocks}).UnlockTheme("frost"); newly {
		t.Error("UnlockTheme reported a held theme as new")
	}
}
