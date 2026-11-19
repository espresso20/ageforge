package ui

import "github.com/espresso20/ageforge/rowid"

// Words said the same way on several screens, so each has one id.
var (
	txtExpeditions      = rowid.E("ui.badge_list.badgeFamilyTitles.19", "Expeditions")
	txtNext             = rowid.E("ui.badge_case.caseTabs.02", "Next")
	txtCountOfTotal     = rowid.E("ui.badge_case.drawTitle.02", "%d of %d")
	txtEarned           = rowid.E("ui.badge_case.drawCounts.02", " earned")
	txtRunsCommandBelow = rowid.E("ui.badge_case.drawKeys.15", "runs the command below")
	txtEscCloses        = rowid.E("ui.badge_case.drawCard.01", " Esc closes ")
)
