package ui

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// The plain-text badge list (the Stats panel and `account badges`) and the
// announcement of a badge just earned. Every test runs in a temp data root
// (mapTestDashboard), never data/.

// TestAccountBadgesCommand: `account badges` gives the count, the title and
// a line for each age in sight that holds a coin, and shows nothing the
// spoiler rules hide: not a later age.
func TestAccountBadgesCommand(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
	_, sum := eng.Badges()

	res := HandleCommand("account badges", eng)
	if res.Type != "info" {
		t.Fatalf("account badges: %+v", res)
	}
	wantAll(t, "the list", res.Message,
		fmt.Sprintf("0 of %d earned", sum.Shown), "0 points", "??? hidden", "Type badges to open the coin cabinet")
	if strings.Contains(res.Message, "pieces") {
		t.Errorf("a new account lists a coin:\n%s", res.Message)
	}

	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
	res = HandleCommand("account badges", eng)
	wantAll(t, "after earning a badge the list", res.Message, fmt.Sprintf("1 of %d earned", sum.Shown), "5 points")

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
	eng.ReportForTest(config.BadgeEvAgeReached, "stone_age")
	_, sum := eng.Badges()

	pg := buildStats(eng.GetState(), partLifetime, 120, 0, false, false, 0)
	tile := func(name string) string {
		for _, tl := range pg.tiles {
			if tl.name == name {
				return tl.text
			}
		}
		t.Fatalf("no %q tile in %+v", name, pg.tiles)
		return ""
	}
	if got := tile(txtBadgesEarn.String()); got != fmt.Sprintf("%d/%d", sum.Earned, sum.Shown) || sum.Earned != 1 {
		t.Errorf("badges earned tile = %q, summary %+v", got, sum)
	}
	if pl := pg.tiles[0].plaque; pl == nil || pl.title != "Settler" {
		t.Errorf("plaque = %+v", pl)
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
	pg := buildStats(eng.GetState(), partLifetime, 120, 0, false, false, 0)
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

// A family's coin on the Stats page is the finest one held in it, drawn by
// the mint; a family holding none, or one whose money is not drawn, is the
// empty recess.
func TestStatsFamilyCoin(t *testing.T) {
	views := []game.BadgeView{
		{Key: "age.a", Family: "age", Earned: true, Coin: game.CoinCopy{Type: "age", Age: "classical_age", Level: 2, Seed: 1}},
		{Key: "age.b", Family: "age", Earned: true, Coin: game.CoinCopy{Type: "age", Age: "classical_age", Level: 4, Seed: 2}},
		{Key: "wonder.a", Family: "wonder", Earned: true, Coin: game.CoinCopy{Type: "wonder", Age: "no_such_age", Level: 3}},
	}
	fams := statsFamilies(views, false)
	if len(fams) != 2 || fams[0].coin.Level != 4 || fams[1].coin.Level != 3 {
		t.Fatalf("families %+v", fams)
	}
	in := newStatsInk(newExpPalette())
	held, empty := statsCoinPic(fams[0].coin, 70, 45, in), statsCoinPic(fams[1].coin, 70, 45, in)
	if coinLevelWord(fams[1].coin.Level) != "Level 3 of 6" {
		t.Errorf("level line %q", coinLevelWord(fams[1].coin.Level))
	}
	if held.Rect != empty.Rect || bytes.Equal(held.Pix, empty.Pix) {
		t.Error("the held coin is not drawn over the recess")
	}
}
