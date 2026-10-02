package game

import "strings"

// DevConsoleCommand is the dev console's entry point: it handles the commands
// defined in this file and hands everything else to DevExecCommand
// (devmode.go). Like DevExecCommand it does nothing unless DevModeActive, so
// none of these commands exist for a normal player. Each one that succeeds
// marks the run dev-touched, as DevExecCommand's do.
//
//	/catastrophe — make the current epoch's catastrophe pending now (testing)
//	/harbinger   — bring the current age's harbinger now (testing)
//	/lastpassage — make the Last Passage pending now, final epoch only (testing)
func DevConsoleCommand(cmd string, ge *GameEngine) string {
	if !DevModeActive {
		return ""
	}
	parts := strings.Fields(cmd)
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/catastrophe" {
		if err := ge.forceCatastrophe(); err != nil {
			return "catastrophe refused: " + err.Error()
		}
		ge.markDevTouched()
		return "catastrophe forced for the current epoch"
	}
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/harbinger" {
		if err := ge.summonHarbinger(); err != nil {
			return "harbinger refused: " + err.Error()
		}
		ge.markDevTouched()
		return "harbinger summoned for the current age"
	}
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/lastpassage" {
		if err := ge.forceLastPassage(); err != nil {
			return "last passage refused: " + err.Error()
		}
		ge.markDevTouched()
		return "the Last Passage is pending"
	}
	return DevExecCommand(cmd, ge)
}
