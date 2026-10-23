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
		// as, with the names in it as keys (see doomMark and eraMark).
		"game.catastrophe.Endure.10", "game.catastrophe.Endure.11",
		"game.catastrophe.Succumb.02", "game.catastrophe.Succumb.03",
		"game.catastrophe.Succumb.04", "game.catastrophe.Succumb.05",
		"game.last_passage.recordLastPassageOutcome.01", "game.last_passage.recordLastPassageOutcome.02",
		"game.last_passage.recordLastPassageOutcome.03", "game.last_passage.recordLastPassageOutcome.04",
		// The name of a timed effect with no Src is kept in the save as it is
		// (ActiveEvent.Name); a name drawn from a ruleset is kept as a key.
		"game.catastrophe.startReconstruction.01", "game.defense.applyWarRaids.01",
		"game.engine.processDiplomacy.01", "game.engine.startChainBoost.01",
		"game.engine.DoFestival.03",
	)
	leftAloneFor(leftKey,
		"game.catastrophe.doomMark.01", "game.catastrophe.eraMark.01", // the markers in a history line
		"game.account.recoveryCodePrefix.01",                         // the first letters of a recovery code
		"game.backup.snapshotPreMigration.03",                        // a file name
		"game.save.DuplicateSave.01",                                 // a file name
		"game.updater.CheckLatest.02", "game.updater.CheckLatest.03", // HTTP headers
		"game.updater.updaterPlatformBinary.01",                    // the name of a release file
		"game.engine.articleName.01", "game.engine.articleName.02", // a or an, by the letter a name starts with
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
