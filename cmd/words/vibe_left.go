package main

// vibe_left.go: the texts of the code that are judged not to follow the
// switch, by row id, and why. vibe-wire leaves them as they are and
// vibe-export does not ask the script for them. The rest are decided from the
// syntax (see wireVerdict).
//
// A row id moves only with the text itself (a wired text keeps its id for
// good), so an entry here stays right until the text is wired or deleted.

func init() {
	leftAloneFor(leftState,
		// The history of dooms is kept in the save as the line it was written
		// as, with the names in it as keys (see doomMark and eraMark). It is
		// told in the script's words from that line when it is read
		// (vibeHistoryLine, the rows game.catastrophe.historyLines.*).
		"game.catastrophe.Endure.10", "game.catastrophe.Endure.11",
		"game.catastrophe.Succumb.02", "game.catastrophe.Succumb.03",
		"game.catastrophe.Succumb.04", "game.catastrophe.Succumb.05",
		"game.last_passage.recordLastPassageOutcome.01", "game.last_passage.recordLastPassageOutcome.02",
		"game.last_passage.recordLastPassageOutcome.03", "game.last_passage.recordLastPassageOutcome.04",
		// The description of a ruins offer written by hand is kept in the save as it was written.
		"game.mission_board.ruinsFoundPlain.01",
		// A road event's story is kept in the save in plain English with its
		// numbers written in, the span of time included (RoadTold); it is told
		// in the script's words from that when it is read (sayRoadDuration).
		"game.road_events.plainRoadDuration.01", "game.road_events.plainRoadDuration.02",
	)
	leftAloneFor(leftKey,
		"game.catastrophe.doomMark.01", "game.catastrophe.eraMark.01", // the markers in a history line
		"game.account.recoveryCodePrefix.01",                         // the first letters of a recovery code
		"game.backup.snapshotPreMigration.03",                        // a file name
		"game.save.DuplicateSave.01",                                 // a file name
		"game.updater.CheckLatest.02", "game.updater.CheckLatest.03", // HTTP headers
		"game.updater.updaterPlatformBinary.01",                    // the name of a release file
		"game.engine.articleName.01", "game.engine.articleName.02", // a or an, by the letter a name starts with
		"game.scholars.scholarTeachSource.01", // the name of a modifier source: the stats panel groups by it and the tests compare it
	)
	leftAloneFor(leftUnits,
		"pkg.textfmt.textfmt.units.01", "pkg.textfmt.textfmt.units.02", "pkg.textfmt.textfmt.units.03",
	)
	// "the " is put in front of an age's name unless the name begins with it
	// already, and " of " is where a name is cut to be made plural and where
	// it is joined again. Both are tested against the names of the ruleset,
	// whichever script they are in, so they are not the script's to change.
	leftAloneFor(leftGlue,
		"game.engine.DoPrestige.01", "game.names.pluralName.01",
	)
	leftAloneFor(leftSearch,
		"game.log_tags.logWords.01", "game.log_tags.logWords.02", "game.log_tags.logWords.03",
		"game.log_tags.logWords.04", "game.log_tags.logWords.05", "game.log_tags.logWords.06",
		"game.log_tags.logWords.07", "game.log_tags.logWords.08", "game.log_tags.logWords.09",
		"game.log_tags.logWords.10", "game.log_tags.logWords.11", "game.log_tags.logWords.12",
		"game.log_tags.logWords.13", "game.log_tags.logWords.14", "game.log_tags.logWords.15",
		"game.log_tags.logWords.16", "game.log_tags.logWords.17", "game.log_tags.logWords.18",
	)
	leftAloneFor(leftDev,
		"config.maker_rule.RawLeftOut.01", "config.maker_rule.RawLeftOut.02",
		"config.maker_rule.MakerHoles.01", "config.maker_rule.MakerHoles.02", "config.maker_rule.MakerHoles.03",
		"game.catastrophe.forceCatastrophe.01", "game.catastrophe.forceCatastrophe.02",
		"game.fate.forceFate.01", "game.fate.forceFate.02",
		"game.harbinger.SummonHarbingerForTest.01",
		"game.last_passage.forceLastPassage.01", "game.last_passage.forceLastPassage.02", "game.last_passage.forceLastPassage.03",
		"game.last_passage.ForceLastPassageForTest.01", "game.legacy.EnterAgeForTest.01",
		"game.engine.devTouchedLogLine.01",
	)
}
