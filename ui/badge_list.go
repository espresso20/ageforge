package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// badge_list.go writes the account's badges as plain text: the Stats
// panel's Lifetime section and the `account badges` command. The views
// arrive with the spoiler rules already applied (game.BadgeView): a
// silhouette has no name or description in it, so nothing here decides
// what to hide.

// badgeFamilyTitles are the headings of the badge families, in the order
// they list. A family not named here lists last, under its own key.
var badgeFamilyTitles = []struct{ key, title string }{
	{"age", "Ages"},
	{"lineage", "Lineages"},
	{"ladder", "Ladders"},
	{"special", "Specials"},
}

// badgeSummaryLine is the header count: "3 of 9 earned, 15 points, ???
// hidden". Hidden badges are not in the "of" number, and how many there
// are stays "???" until the account has seen the last age.
func badgeSummaryLine(sum game.BadgeSummary) string {
	parts := []string{
		fmt.Sprintf("%d of %d earned", sum.Earned, sum.Shown),
		textfmt.Count(sum.Points, "point", "points"),
	}
	if sum.Hidden > 0 {
		hidden := game.BadgeHiddenName
		if sum.HiddenCounted {
			hidden = textfmt.Int(sum.Hidden)
		}
		parts = append(parts, hidden+" hidden")
	}
	return strings.Join(parts, ", ")
}

// badgeLine is one badge on one line. ok is false for a silhouette with
// nothing to say (a hidden badge that is not a secret): those are only
// counted.
func badgeLine(v game.BadgeView) (line string, ok bool) {
	tier := ""
	if v.Tier != "" {
		tier = " [gray](" + v.Tier + ")[-]"
	}
	switch {
	case v.Earned:
		line = fmt.Sprintf("[green]★[-] %s%s  %s", lit(v.Name), tier, lit(v.Desc))
		if v.Crossed {
			line += " [red]Earned in a modified game: no points.[-]"
		}
		return line, true
	case v.Hidden:
		if v.Desc == "" {
			return "", false
		}
		// One color for the whole line: a nested tag would end it early.
		plain := v.Name
		if v.Tier != "" {
			plain += " (" + v.Tier + ")"
		}
		return fmt.Sprintf("[gray]? %s  %s[-]", plain, lit(v.Desc)), true
	}
	line = fmt.Sprintf("[gray]☆[-] %s%s  [gray]%s[-]", lit(v.Name), tier, lit(v.Desc))
	if v.Target > 0 {
		line += fmt.Sprintf(" [gray](%s of %s)[-]", FormatNumber(v.Progress), FormatNumber(v.Target))
	}
	return line, true
}

// badgeListLines is the whole list: a count line, then each family's
// badges under a heading, earned and locked.
func badgeListLines(views []game.BadgeView, sum game.BadgeSummary) []string {
	lines := []string{"[gold]Badges:[-] " + badgeSummaryLine(sum)}
	byFamily := map[string][]string{}
	var order []string
	for _, v := range views {
		line, ok := badgeLine(v)
		if !ok {
			continue
		}
		if _, seen := byFamily[v.Family]; !seen {
			order = append(order, v.Family)
		}
		byFamily[v.Family] = append(byFamily[v.Family], line)
	}
	listed := map[string]bool{}
	family := func(key, title string) {
		if listed[key] || len(byFamily[key]) == 0 {
			return
		}
		listed[key] = true
		lines = append(lines, " [yellow]"+title+"[-]")
		for _, l := range byFamily[key] {
			lines = append(lines, "   "+l)
		}
	}
	for _, f := range badgeFamilyTitles {
		family(f.key, f.title)
	}
	for _, key := range order {
		family(key, textfmt.Capitalize(strings.ReplaceAll(key, "_", " ")))
	}
	if len(lines) == 1 {
		lines = append(lines, "   [gray]None to show yet.[-]")
	}
	return lines
}

// badgeToast is the one-line toast for a badge just earned.
func badgeToast(v game.BadgeView) string {
	if v.Tier == "" {
		return fmt.Sprintf("Badge earned: %s", lit(v.Name))
	}
	return fmt.Sprintf("Badge earned: %s (%s)", lit(v.Name), v.Tier)
}
