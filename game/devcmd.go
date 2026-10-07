package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// DevConsoleCommand is the dev console's entry point: it handles the commands
// defined in this file and hands everything else to DevExecCommand
// (devmode.go). Like DevExecCommand it does nothing unless DevModeActive, so
// none of these commands exist for a normal player. Each one that succeeds
// marks the run dev-touched, as DevExecCommand's do.
//
//	/catastrophe — make the current epoch's catastrophe pending now (testing)
//	/harbinger   — bring the current age's harbinger now (testing)
//	/lastpassage — make the Last Passage pending now, final epoch only (testing)
//	/mastery <age|all> <0-10> — set Era Mastery for one age or every age
//	/record <age> — set the record (the deepest age ever entered)
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
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/mastery" {
		if len(parts) != 3 {
			return "usage: /mastery <age|all> <0-10>"
		}
		m, err := strconv.Atoi(parts[2])
		if err != nil || m < 0 || m > config.MasteryCap {
			return fmt.Sprintf("mastery must be 0 to %d", config.MasteryCap)
		}
		if err := ge.devSetMastery(strings.ToLower(parts[1]), m); err != nil {
			return "mastery refused: " + err.Error()
		}
		ge.markDevTouched()
		return fmt.Sprintf("mastery %d set for %s", m, parts[1])
	}
	if len(parts) > 0 && strings.ToLower(parts[0]) == "/record" {
		if len(parts) != 2 {
			return "usage: /record <age>"
		}
		if err := ge.devSetRecord(strings.ToLower(parts[1])); err != nil {
			return "record refused: " + err.Error()
		}
		ge.markDevTouched()
		return "record set to " + parts[1]
	}
	return DevExecCommand(cmd, ge)
}

// devSetMastery sets age's mastery (or every age's, for "all") and reruns
// the rates at the new speed (a drop applies the grace rule).
func (ge *GameEngine) devSetMastery(age string, m int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if age == "all" {
		for _, a := range ge.rules.AgeKeys() {
			ge.Prestige.SetMastery(a, m)
		}
	} else if _, ok := ge.rules.Index(age); ok {
		ge.Prestige.SetMastery(age, m)
	} else {
		return fmt.Errorf("unknown age %q", age)
	}
	ge.recalculateRates()
	return nil
}

// devSetRecord sets the record and reruns the rates (catch-up may change).
func (ge *GameEngine) devSetRecord(age string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.rules.Index(age); !ok {
		return fmt.Errorf("unknown age %q", age)
	}
	ge.Prestige.SetRecord(age)
	ge.recalculateRates()
	return nil
}
