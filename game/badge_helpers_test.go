package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// Helpers for tests that work an account without an engine. Every test that
// uses them runs in a temp data root (isolateAccountDir), never data/.

// The seed catalog's keys the tests name.
const (
	badgePrestige1  = "ladder.prestiges.1"
	badgePrestige3  = "ladder.prestiges.2"
	badgePrestige10 = "ladder.prestiges.3"
	badgePrestige25 = "ladder.prestiges.4"
	badgeStone      = "age.stone_age"
	badgeIron       = "age.iron_age"
	badgeModern     = "age.modern_age"
	badgeHousing1   = "lineage.housing.1"
	badgeHousing2   = "lineage.housing.2"
	badgeCollector  = "special.collector" // 15 badges earned
	badgeHutHoarder = "special.hut_hoarder"
	badgeLate       = "special.fashionably_late"
	badgeSale       = "special.liquidation_sale"
	badgeCookieJar  = "special.hand_in_the_cookie_jar"
	badgeAccounting = "special.creative_accounting"
)

// coreBook is the core ruleset's badges, indexed.
func coreBook() *badgeBook { return newBadgeBook(rules.Core()) }

// hasBadge reports whether the account holds the badge.
func hasBadge(a *Account, key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.earnedLocked(key)
}

// recordPrestige is a prestige from age as the engine records it: the
// lifetime stat, then the report the badges are judged on.
func recordPrestige(a *Account, from string) {
	a.RecordPrestigeFrom(from)
	a.judge(coreBook(), Event{Kind: config.BadgeEvPrestige, Subject: from}, badgeCtx{age: from})
}

// recordAge is an age reached as the engine records it.
func recordAge(t *testing.T, a *Account, age string) {
	t.Helper()
	def, ok := rules.Core().Age(age)
	if !ok {
		t.Fatalf("unknown age %q", age)
	}
	a.RecordAgeReached(age, def.Order)
	a.judge(coreBook(), Event{Kind: config.BadgeEvAgeReached, Subject: age}, badgeCtx{age: age})
}

// badgeKeys returns the keys of the badges in views that are earned.
func earnedKeys(views []BadgeView) []string {
	var out []string
	for _, v := range views {
		if v.Earned {
			out = append(out, v.Key)
		}
	}
	return out
}

// viewOf returns the view of one badge, failing the test if it is not listed.
func viewOf(t *testing.T, views []BadgeView, key string) BadgeView {
	t.Helper()
	for _, v := range views {
		if v.Key == key {
			return v
		}
	}
	t.Fatalf("the badge %s is not listed", key)
	return BadgeView{}
}
