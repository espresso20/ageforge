package ui

import (
	"fmt"
	"sort"
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
	{"swift", "Swift Eras"},
	{"wonder", "Wonders"},
	{"maximalist", "Maximalists"},
	{"techs", "Research"},
	{"lineage", "Lineages"},
	{"domain", "Payrolls"},
	{"resource", "Resources"},
	{"civ", "Civilizations"},
	{"harbinger", "Harbingers"},
	{"doom", "Catastrophes"},
	{"expedition", "Expeditions"},
	{"awakening", "Awakenings"},
	{"ladder", "Ladders"},
	{"theme", "Themes"},
	{"map", "Maps"},
	{"special", "Specials"},
}

// badgeListLocked is how many locked badges of a family the plain list
// writes out before it says how many more there are: the list is a digest,
// and the badge case is where every badge has its place.
const badgeListLocked = 4

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

// badgeInSight reports whether a badge has anything to show: it is earned,
// or locked in plain sight, or a secret with its hint. A family with none
// in sight is not named at all: its heading alone would say what is ahead.
func badgeInSight(v game.BadgeView) bool { return !v.Hidden || v.Secret }

// badgeFamilies sorts the badges in sight into their families, in the
// order the families list.
type badgeFamilyList struct {
	title          string
	earned, locked []game.BadgeView
	secrets        int
	// got of shown: the family's earned and its badges in plain sight.
	got, shown int
}

func badgeFamilies(views []game.BadgeView) []badgeFamilyList {
	by := map[string]*badgeFamilyList{}
	var order []string
	for _, v := range views {
		if !badgeInSight(v) {
			continue
		}
		f := by[v.Family]
		if f == nil {
			f = &badgeFamilyList{title: familyTitle(v.Family)}
			by[v.Family] = f
			order = append(order, v.Family)
		}
		switch {
		case v.Earned:
			f.earned = append(f.earned, v)
			if !v.Integrity {
				f.got++
				f.shown++
			}
		case v.Hidden:
			f.secrets++
		default:
			f.locked = append(f.locked, v)
			f.shown++
		}
	}
	var out []badgeFamilyList
	listed := map[string]bool{}
	add := func(key string) {
		if f := by[key]; f != nil && !listed[key] {
			listed[key] = true
			out = append(out, *f)
		}
	}
	for _, f := range badgeFamilyTitles {
		add(f.key)
	}
	for _, key := range order {
		add(key)
	}
	return out
}

// badgeHeadLines is the top of both lists: the count, and the title worn.
func badgeHeadLines(sum game.BadgeSummary) []string {
	lines := []string{"[gold]Badges:[-] " + badgeSummaryLine(sum)}
	if sum.Title != "" {
		title := " [gold]Title:[-] " + wornTitle(sum)
		if sum.NextTitle != "" {
			title += fmt.Sprintf(" [gray](%s at %s points)[-]", sum.NextTitle, wholeNumber(sum.NextTitleAt))
		}
		lines = append(lines, title)
	}
	return lines
}

const badgeCaseHint = " [gray]Type badges to open the badge case and see them all.[-]"

// badgeListLines is the account's badges as a list (`account badges`): a
// count line, then each family in sight under a heading with its own
// count, its earned badges, the first few it has still to earn and how
// many secrets it keeps. The badge case (`badges`) shows them all.
func badgeListLines(views []game.BadgeView, sum game.BadgeSummary) []string {
	lines := badgeHeadLines(sum)
	fams := badgeFamilies(views)
	for _, f := range fams {
		lines = append(lines, fmt.Sprintf(" [yellow]%s[-] [gray](%d of %d)[-]", f.title, f.got, f.shown))
		for _, v := range f.earned {
			if l, ok := badgeLine(v); ok {
				lines = append(lines, "   "+l)
			}
		}
		for i, v := range f.locked {
			if i == badgeListLocked {
				lines = append(lines, fmt.Sprintf("   [gray]and %s more to earn.[-]", wholeNumber(len(f.locked)-i)))
				break
			}
			if l, ok := badgeLine(v); ok {
				lines = append(lines, "   "+l)
			}
		}
		if f.secrets > 0 {
			lines = append(lines, "   [gray]? "+textfmt.Count(f.secrets, "secret badge", "secret badges")+", each with a hint in the badge case.[-]")
		}
	}
	if len(fams) == 0 {
		return append(lines, "   [gray]None to show yet.[-]")
	}
	return append(lines, badgeCaseHint)
}

// badgeDigestLines is the account's badges in a few lines, for the Stats
// panel: the count, the title, each family's count on a line, and the
// badges earned last.
func badgeDigestLines(views []game.BadgeView, sum game.BadgeSummary) []string {
	lines := badgeHeadLines(sum)
	fams := badgeFamilies(views)
	var latest []game.BadgeView
	for _, f := range fams {
		note := ""
		if f.secrets > 0 {
			note = " [gray]and " + textfmt.Count(f.secrets, "secret", "secrets") + "[-]"
		}
		lines = append(lines, fmt.Sprintf("   [yellow]%-14s[-] %d of %d%s", f.title, f.got, f.shown, note))
		for _, v := range f.earned {
			if !v.At.IsZero() {
				latest = append(latest, v)
			}
		}
	}
	if len(fams) == 0 {
		return append(lines, "   [gray]None to show yet.[-]")
	}
	sort.SliceStable(latest, func(a, b int) bool { return latest[a].At.After(latest[b].At) })
	for i, v := range latest {
		if i == 3 {
			break
		}
		head := " [gold]Latest:[-]"
		if i > 0 {
			head = "        "
		}
		if l, ok := badgeLine(v); ok {
			lines = append(lines, head+" "+l)
		}
	}
	return append(lines, badgeCaseHint)
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
