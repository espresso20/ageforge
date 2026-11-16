package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The plain-text badge list (the Stats panel and `account badges`) and the
// announcement of a badge just earned. Every test runs in a temp data root
// (mapTestDashboard), never data/.

// TestAccountBadgesCommand: `account badges` lists the account's badges
// family by family, earned and the first few still to earn, and shows
// nothing the spoiler rules hide: not a later age, not a secret's name, and
// not the heading of a family the player has not come to yet.
func TestAccountBadgesCommand(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
	_, sum := eng.Badges()

	res := HandleCommand("account badges", eng)
	if res.Type != "info" {
		t.Fatalf("account badges: %+v", res)
	}
	wantAll(t, "the list", res.Message,
		fmt.Sprintf("0 of %d earned", sum.Shown), "0 points", "??? hidden",
		"Rock Solid", "Reach the Stone Age.",
		"Faith Hobbyist", "(0 of 10)", "more to earn.",
		"First Prestige", "Hut Hoarder",
		"secret badges, each with a hint in the badge case.",
		"Ages", "Lineages", "Ladders", "Specials", "Payrolls", "Resources",
		"Type badges to open the badge case",
	)
	for _, hidden := range []string{
		"Iron", "Modern", "Liquidation", "Cookie", "Creative Accounting",
		// Families with nothing in sight yet are not named.
		"Civilizations", "Harbingers", "Catastrophes", "Awakenings",
	} {
		if strings.Contains(res.Message, hidden) {
			t.Errorf("the list shows %q, which a new account may not see:\n%s", hidden, res.Message)
		}
	}
	// A digest, not the catalog: a new account's list fits a screen or two.
	if n := strings.Count(res.Message, "\n"); n > 90 {
		t.Errorf("the list for a new account runs to %d lines", n)
	}

	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
	res = HandleCommand("account badges", eng)
	wantAll(t, "after earning a badge the list", res.Message, fmt.Sprintf("1 of %d earned", sum.Shown), "5 points", "★[-] Rock Solid")

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

// TestStatsPanelListsBadges: the Stats panel's Lifetime part has the number of
// badges earned and the title, and does not list locked badges.
func TestStatsPanelListsBadges(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
	eng.ReportForTest(config.BadgeEvPrestige, "modern_age")
	_, sum := eng.Badges()

	pg := buildStats(eng.GetState(), partLifetime, 120, false, false, 0)
	tile := func(name string) string {
		for _, tl := range pg.tiles {
			if tl.name == name {
				return tl.text
			}
		}
		t.Fatalf("no %q tile in %+v", name, pg.tiles)
		return ""
	}
	if got := tile(txtBadgesEarn.String()); got != fmt.Sprint(sum.Earned) || sum.Earned != 1 {
		t.Errorf("badges earned tile = %q, summary %+v", got, sum)
	}
	if got := tile(txtTitleOf.String()); got != "Settler" {
		t.Errorf("title tile = %q", got)
	}
	text := statsPageText(eng.GetState(), partLifetime, 120)
	if strings.Contains(text, "Achievements") || strings.Contains(text, "☆") || strings.Contains(text, "Rock Solid") {
		t.Errorf("the Lifetime part lists achievements or locked badges:\n%s", text)
	}
}

// TestDashboardAnnouncesBadges: a badge earned in play gets its card and one
// log line, once.
func TestDashboardAnnouncesBadges(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
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
	// The card, not a toast: the notice stays empty.
	if got := untag(d.toastMgr.GetCurrent()); got != "" {
		t.Errorf("the toast reads %q", got)
	}
	if len(d.moment.queue) != 1 {
		t.Errorf("the card has %d badges waiting, want 1", len(d.moment.queue))
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

// TestStatsPanelShowsCivilizationsStarted is B12 on screen: the Lifetime
// part shows Civilizations Started, which is now counted, and nothing
// about saves completed, which never was and is retired.
func TestStatsPanelShowsCivilizationsStarted(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
	pg := buildStats(eng.GetState(), partLifetime, 120, false, false, 0)
	found := false
	for _, tl := range pg.tiles {
		found = found || tl.name == txtCivs.String() && tl.text == "1"
	}
	if !found {
		t.Errorf("the Stats panel does not show one civilization started: %+v", pg.tiles)
	}
	if text := statsPageText(eng.GetState(), partLifetime, 120); strings.Contains(strings.ToLower(text), "saves completed") {
		t.Error("the Stats panel shows saves completed")
	}

	// The recover guard lists what an account holds: a civilization
	// started counts, a stored saves_completed does not.
	acct := eng.Account()
	if got := accountHoldings(acct); !strings.Contains(got, "1 civilization started") {
		t.Errorf("holdings %q do not mention the civilization started", got)
	}
	other, err := game.CreateAccount("Other Player")
	must(t, err)
	other.Stats.SavesCompleted = 3
	if got := accountHoldings(other); got != "" {
		t.Errorf("an account holding only a retired stat reads as holding %q", got)
	}
}
