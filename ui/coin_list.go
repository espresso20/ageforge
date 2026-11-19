package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// coin_list.go writes the account's coins as plain text for `account badges`:
// the badge count and title, then a line for each age the player may see
// that holds a coin.

// wholeNumber writes a count in full, its thousands set off with commas
// ("1,245"): a score is exact, and reads wrong rounded to "1.25K".
func wholeNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}

// badgeSummaryLine is the header count: "3 of 9 earned, 15 points, ???
// hidden". Hidden badges are not in the "of" number, and how many there
// are stays "???" until the account has seen the last age.
func badgeSummaryLine(sum game.BadgeSummary) string {
	parts := []string{
		vibe.Sprintf("ui.badge_list.badgeSummaryLine.01", "%d of %d earned", sum.Earned, sum.Shown),
		textfmt.Count(sum.Points, vibe.Line("ui.badge_list.badgeSummaryLine.02", "point"), vibe.Line("ui.badge_list.badgeSummaryLine.03", "points")),
	}
	if sum.Hidden > 0 {
		hidden := game.BadgeHiddenName
		if sum.HiddenCounted {
			hidden = textfmt.Int(sum.Hidden)
		}
		parts = append(parts, hidden+vibe.Line("ui.badge_list.badgeSummaryLine.04", " hidden"))
	}
	return strings.Join(parts, ", ")
}

var coinListHint = vibe.E("ui.badge_list.badgeCaseHint.01", " [gray]Type badges to open the coin cabinet and see them all.[-]")

// coinListLines is the list: the count, the title worn, then each age in
// sight with the pieces its drawer holds and its finest level.
func coinListLines(state game.GameState, acct *game.Account, sum game.BadgeSummary) []string {
	lines := []string{vibe.Line("ui.badge_list.badgeHeadLines.01", "[gold]Badges:[-] ") + badgeSummaryLine(sum)}
	if sum.Title != "" {
		title := vibe.Line("ui.badge_list.badgeHeadLines.02", " [gold]Title:[-] ") + wornTitle(sum)
		if sum.NextTitle != "" {
			title += vibe.Sprintf("ui.badge_list.badgeHeadLines.03", " [gray](%s at %s points)[-]", sum.NextTitle, wholeNumber(sum.NextTitleAt))
		}
		lines = append(lines, title)
	}
	sight := game.SightOf(&state)
	for _, a := range config.Ages() {
		held, best := 0, 0
		for _, typ := range append([]string{config.CoinFeat}, config.CoinTypes...) {
			if b := acct.CoinBest(typ, a.Key); b > 0 {
				held, best = held+1, max(best, b)
			}
		}
		if held > 0 && sight.Age(a.Key) {
			lines = append(lines, vibe.Sprintf("ui.badge_list.coinListLines.01", " [yellow]%s[-] [gray](%d of %d pieces, finest level %d)[-]", a.Name, held, len(config.CoinTypes)+1, best))
		}
	}
	return append(lines, coinListHint.String())
}
