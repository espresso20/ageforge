package game

import "strings"

// DevConsoleCommand is the dev console's entry point: it handles the commands
// defined in this file and hands everything else to DevExecCommand
// (devmode.go). Like DevExecCommand it does nothing unless DevModeActive, so
// none of these commands exist for a normal player.
//
//	/catastrophe — make the current epoch's catastrophe pending now (testing)
func DevConsoleCommand(cmd string, ge *GameEngine) string {
	if !DevModeActive {
		return ""
	}
	parts := strings.Fields(cmd)
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/catastrophe" {
		if err := ge.forceCatastrophe(); err != nil {
			return "catastrophe refused: " + err.Error()
		}
		return "catastrophe forced for the current epoch"
	}
	return DevExecCommand(cmd, ge)
}
