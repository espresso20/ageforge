package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
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
	if sum.Title != "" {
		title := " [gold]Title:[-] " + sum.Title
		if sum.NextTitle != "" {
			title += fmt.Sprintf(" [gray](%s at %s points)[-]", sum.NextTitle, wholeNumber(sum.NextTitleAt))
		}
		lines = append(lines, title)
	}
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

// badgeToast is the one-row toast for a badge just earned, written for a
// bar w cells wide (0: any width): the badge in three cells, its emblem
// between two half discs in its tier's colours, then its name and tier,
// and its description when the bar has the room. The toast bar is one row,
// so this is the badge's art at its smallest.
func badgeToast(v game.BadgeView, tier mapmodel.GlyphTier, w int) string {
	pal := newGridPalette(0, nil)
	hex := func(k ink) string { return theme.HexTag(pal.inkOn(k, false)) }
	rim, core := inkCyan, inkMagenta
	if !v.Integrity {
		_, rim, core = tierInks(v.Level)
	}
	e, _ := emblemOf(v.Emblem)
	mark := []rune(toastMark(e.glyph.In(tier)))
	if tier == mapmodel.TierASCII {
		for i, r := range mark {
			if f, ok := badgeFold[r]; ok {
				mark[i] = f
			}
		}
	}
	const head = "Badge earned:"
	name, tierText, desc := v.Name, "", v.Desc
	if v.Tier != "" {
		tierText = " (" + v.Tier + ")"
	}
	// The mark, a space, the head, a space, the name, the tier, a full stop.
	fixed := len(mark) + 1 + len(head) + 1 + runeLen(tierText) + 1
	if w > 0 && fixed+runeLen(name)+1+runeLen(desc) > w {
		desc = ""
	}
	if room := w - fixed; w > 0 && runeLen(name) > room {
		name = clipRunes(name, max(room-1, 1)) + "…"
	}
	var sb strings.Builder
	sb.WriteString(hex(rim) + string(mark[0]) + hex(core) + lit(string(mark[1])) + hex(rim) + string(mark[2]) + "[-] ")
	sb.WriteString("[" + theme.TagName(theme.RoleAccent) + "::b]" + head + theme.Reset + " ")
	sb.WriteString("[" + theme.TagName(theme.RoleBright) + "::b]" + lit(name) + theme.Reset)
	if tierText != "" {
		sb.WriteString(theme.HexTag(theme.Legible(inkValue(rim, pal.light), pal.bg, 4.5)) + tierText + "[-]")
	}
	sb.WriteString(".")
	if desc != "" {
		sb.WriteString(" " + lit(desc))
	}
	return sb.String()
}
