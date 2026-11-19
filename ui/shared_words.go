package ui

import "github.com/espresso20/ageforge/vibe"

// Words said the same way on several screens, so each has one id.
var (
	txtExpeditions      = vibe.E("ui.badge_list.badgeFamilyTitles.19", "Expeditions")
	txtNext             = vibe.E("ui.badge_case.caseTabs.02", "Next")
	txtCountOfTotal     = vibe.E("ui.badge_case.drawTitle.02", "%d of %d")
	txtEarned           = vibe.E("ui.badge_case.drawCounts.02", " earned")
	txtRunsCommandBelow = vibe.E("ui.badge_case.drawKeys.15", "runs the command below")
	txtEscCloses        = vibe.E("ui.badge_case.drawCard.01", " Esc closes ")
)
