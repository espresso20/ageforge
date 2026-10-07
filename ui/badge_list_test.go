package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The plain-text badge list (the Stats panel and `account badges`) and the
// announcement of a badge just earned. Every test runs in a temp data root
// (mapTestDashboard), never data/.

// TestAccountBadgesCommand: `account badges` lists the account's badges,
// earned and locked, and shows nothing the spoiler rules hide.
func TestAccountBadgesCommand(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}

	res := HandleCommand("account badges", eng)
	if res.Type != "info" {
		t.Fatalf("account badges: %+v", res)
	}
	for _, want := range []string{
		"0 of 9 earned", "0 points", "??? hidden",
		"Rock Solid", "Reach the Stone Age.",
		"Housing Hobbyist", "(0 of 14)",
		"First Prestige", "Hut Hoarder", "Fashionably Late",
		"Something about a clearance.",
		"Ages", "Lineages", "Ladders", "Specials",
	} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("the list does not have %q:\n%s", want, res.Message)
		}
	}
	for _, hidden := range []string{"Iron", "Modern", "Liquidation", "Cookie", "Creative Accounting"} {
		if strings.Contains(res.Message, hidden) {
			t.Errorf("the list shows %q, which a new account may not see:\n%s", hidden, res.Message)
		}
	}

	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
	res = HandleCommand("account badges", eng)
	for _, want := range []string{"1 of 9 earned", "5 points", "★[-] Rock Solid"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("after earning a badge the list does not have %q:\n%s", want, res.Message)
		}
	}

	// The bare `account` reply points at it.
	if res := HandleCommand("account", eng); !strings.Contains(res.Message, "account badges") {
		t.Errorf("the account reply does not mention account badges:\n%s", res.Message)
	}
}

// TestAccountBadgesWithoutAnAccount: no account, a plain refusal.
func TestAccountBadgesWithoutAnAccount(t *testing.T) {
	_, eng := mapTestDashboard(t, false)
	if res := HandleCommand("account badges", eng); res.Type != "warning" {
		t.Errorf("account badges with no account: %+v", res)
	}
}

// TestStatsPanelListsBadges: the Stats panel's Lifetime section lists the
// badges, where it used to list four achievement names.
func TestStatsPanelListsBadges(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	eng.ReportForTest(config.BadgeEvPrestige, "modern_age")

	text := statsProvider(eng.GetState(), 0)
	for _, want := range []string{"Lifetime (account)", "Badges:", "1 of 9 earned", "★[-] First Prestige", "☆[-] Rock Solid"} {
		if !strings.Contains(text, want) {
			t.Errorf("the Stats panel does not have %q", want)
		}
	}
	if strings.Contains(text, "Achievements") {
		t.Error("the Stats panel still has an Achievements heading")
	}
}

// TestDashboardAnnouncesBadges: a badge earned in play gets a toast and one
// log line, once.
func TestDashboardAnnouncesBadges(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")

	const line = "Badge earned: Rock Solid (bronze). Reach the Stone Age."
	count := func() int {
		n := 0
		for _, e := range eng.GetLogs() {
			if e.Message == line {
				n++
			}
		}
		return n
	}
	if count() != 0 {
		t.Fatal("the log line was written under the engine's lock, before the dashboard drained it")
	}
	d.refresh()
	if got := d.toastMgr.GetCurrent(); !strings.Contains(got, "Badge earned: Rock Solid (bronze)") {
		t.Errorf("the toast reads %q", got)
	}
	if count() != 1 {
		t.Errorf("the badge's log line appears %d times, want once", count())
	}
	d.refresh()
	d.refresh()
	if count() != 1 {
		t.Errorf("after more refreshes the log line appears %d times", count())
	}
}

// TestBadgeLines: each form a badge takes in the list.
func TestBadgeLines(t *testing.T) {
	earned := game.BadgeView{Name: "Rock Solid", Desc: "Reach the Stone Age.", Tier: "bronze", Earned: true, At: time.Unix(1, 0)}
	if line, ok := badgeLine(earned); !ok || !strings.Contains(line, "★") || !strings.Contains(line, "(bronze)") {
		t.Errorf("an earned badge: %q %v", line, ok)
	}
	crossed := earned
	crossed.Crossed = true
	if line, _ := badgeLine(crossed); !strings.Contains(line, "no points") {
		t.Errorf("a crossed badge does not say it adds no points: %q", line)
	}
	locked := game.BadgeView{Name: "Housing Hobbyist", Desc: "Build 14 housing buildings.", Tier: "bronze", Progress: 3, Target: 14}
	if line, ok := badgeLine(locked); !ok || !strings.Contains(line, "☆") || !strings.Contains(line, "(3 of 14)") {
		t.Errorf("a locked badge with a count: %q %v", line, ok)
	}
	secret := game.BadgeView{Name: game.BadgeHiddenName, Desc: "Something about huts.", Tier: "gold", Hidden: true, Secret: true}
	if line, ok := badgeLine(secret); !ok || !strings.Contains(line, "???") || !strings.Contains(line, "Something about huts.") {
		t.Errorf("a secret badge: %q %v", line, ok)
	}
	silhouette := game.BadgeView{Name: game.BadgeHiddenName, Tier: "gold", Hidden: true}
	if line, ok := badgeLine(silhouette); ok {
		t.Errorf("a silhouette with nothing to say was listed: %q", line)
	}
	integrity := game.BadgeView{Name: "Hand in the Cookie Jar", Desc: "Unlock the developer console.", Earned: true, Integrity: true}
	if line, ok := badgeLine(integrity); !ok || strings.Contains(line, "()") {
		t.Errorf("an integrity badge has no tier to print: %q %v", line, ok)
	}

	// The count of hidden badges stays "???" until the account has seen the last age.
	sum := game.BadgeSummary{Earned: 2, Shown: 9, Hidden: 3, Points: 30}
	if got := badgeSummaryLine(sum); got != "2 of 9 earned, 30 points, ??? hidden" {
		t.Errorf("summary %q", got)
	}
	sum.HiddenCounted = true
	if got := badgeSummaryLine(sum); got != "2 of 9 earned, 30 points, 3 hidden" {
		t.Errorf("summary with the count shown %q", got)
	}
	sum.Hidden, sum.Points, sum.Earned = 0, 5, 1
	if got := badgeSummaryLine(sum); got != "1 of 9 earned, 5 points" {
		t.Errorf("summary with nothing hidden %q", got)
	}
}
