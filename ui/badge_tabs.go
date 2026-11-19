package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// The tabs that are not a family, and the size of a small coin in cells.
const (
	caseTabAll  = ""
	caseTabNext = "next"

	coinCols, coinRows = 7, 3
)

// Words said the same way on several screens, so each has one id.
var (
	txtNext             = vibe.E("ui.badge_case.caseTabs.02", "Next")
	txtCountOfTotal     = vibe.E("ui.badge_case.drawTitle.02", "%d of %d")
	txtEarned           = vibe.E("ui.badge_case.drawCounts.02", " earned")
	txtRunsCommandBelow = vibe.E("ui.badge_case.drawKeys.15", "runs the command below")
	txtEscCloses        = vibe.E("ui.badge_case.drawCard.01", " Esc closes ")
)

// caseTab is a tab and its family's counts.
type caseTab struct {
	key, title            string
	earned, shown, hidden int
	// sight is how many of the family's badges have something to show
	// (badgeInSight).
	sight int
}

// familyTitle is a family's heading.
func familyTitle(key string) string {
	for _, f := range badgeFamilyTitles {
		if f.key == key {
			return f.title.String()
		}
	}
	return textfmt.Capitalize(strings.ReplaceAll(key, "_", " "))
}

// caseTabs lists the tabs: All, each family in sight in the families' own
// order, and Next. A family with nothing in sight has no tab and no
// heading: its name alone would say what is ahead. Once the account may
// know how many badges are withheld (counted), every family is listed.
func caseTabs(views []game.BadgeView, counted bool) []caseTab {
	stat := map[string]*caseTab{}
	var seen []string
	for _, v := range views {
		t := stat[v.Family]
		if t == nil {
			t = &caseTab{key: v.Family, title: familyTitle(v.Family)}
			stat[v.Family] = t
			seen = append(seen, v.Family)
		}
		if badgeInSight(v) {
			t.sight++
		}
		switch {
		case v.Integrity:
		case v.Earned:
			t.earned++
			t.shown++
		case v.Hidden:
			t.hidden++
		default:
			t.shown++
		}
	}
	tabs := []caseTab{{key: caseTabAll, title: vibe.Line("ui.badge_case.caseTabs.01", "All")}}
	listed := map[string]bool{}
	add := func(key string) {
		if t := stat[key]; t != nil && !listed[key] && (t.sight > 0 || counted) {
			listed[key] = true
			tabs = append(tabs, *t)
		}
	}
	for _, f := range badgeFamilyTitles {
		add(f.key)
	}
	for _, key := range seen {
		add(key)
	}
	return append(tabs, caseTab{key: caseTabNext, title: txtNext.String()})
}
