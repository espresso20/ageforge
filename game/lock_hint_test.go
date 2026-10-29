package game

import (
	"reflect"
	"testing"
)

// TestLockHintsNameTheTechAndItsChain: an altar of the Bronze Age waits for
// Calendar, which waits for Ritual, which waits for Language. The state names
// the tech and lists the research still to come nearest first, and the list
// shrinks as each is researched; a building whose tech is done carries no hint.
func TestLockHintsNameTheTechAndItsChain(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "bronze_age"
	ge.applyAgeUnlocks("bronze_age")

	altar := func() BuildingState {
		t.Helper()
		bs, ok := ge.GetState().Buildings["altar"]
		if !ok {
			t.Fatal("no altar in the state")
		}
		return bs
	}

	if got := altar(); got.LockTech != "Calendar" || !reflect.DeepEqual(got.LockTechPath, []string{"Ritual", "Language"}) {
		t.Fatalf("altar locked by %q after %v, want Calendar after [Ritual Language]", got.LockTech, got.LockTechPath)
	}

	ge.GrantTechsForTest("language")
	if got := altar(); got.LockTech != "Calendar" || !reflect.DeepEqual(got.LockTechPath, []string{"Ritual"}) {
		t.Fatalf("with Language: altar locked by %q after %v, want Calendar after [Ritual]", got.LockTech, got.LockTechPath)
	}

	ge.GrantTechsForTest("ritual")
	if got := altar(); got.LockTech != "Calendar" || len(got.LockTechPath) != 0 {
		t.Fatalf("with Ritual: altar locked by %q after %v, want Calendar with no path", got.LockTech, got.LockTechPath)
	}

	ge.GrantTechsForTest("calendar")
	if got := altar(); got.LockTech != "" || got.LockTechPath != nil || !got.Unlocked {
		t.Fatalf("with Calendar: altar locked by %q after %v (unlocked %v), want no hint and unlocked", got.LockTech, got.LockTechPath, got.Unlocked)
	}
}

// TestLockHintsOnlyForAgeUnlocked: a building its age has not unlocked yet
// carries no hint, so the panel never names a tech for a building the player
// has not been offered.
func TestLockHintsOnlyForAgeUnlocked(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "primitive_age"
	if got := ge.GetState().Buildings["altar"]; got.LockTech != "" || len(got.LockTechPath) != 0 {
		t.Fatalf("altar before its age: locked by %q after %v, want no hint", got.LockTech, got.LockTechPath)
	}
}

// TestLockTechNextListsOnlyResearchableTechs: the frontier lists only the
// techs on the path to a locked tech whose own prerequisites are all
// researched. The Altar waits for Calendar, which waits for Ritual, which
// waits for Language. At the start, only Language can be researched (no
// prerequisites). After Language, only Ritual. After Ritual, the path is
// empty so LockTechNext is empty (Calendar itself is researchable).
func TestLockTechNextListsOnlyResearchableTechs(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "bronze_age"
	ge.applyAgeUnlocks("bronze_age")

	altar := func() BuildingState {
		t.Helper()
		bs, ok := ge.GetState().Buildings["altar"]
		if !ok {
			t.Fatal("no altar in the state")
		}
		return bs
	}

	// At the start, only Language can be researched (it has no prerequisites)
	if got := altar(); !reflect.DeepEqual(got.LockTechNext, []string{"Language"}) {
		t.Fatalf("at start: next=%v, want [Language]", got.LockTechNext)
	}

	// After Language, only Ritual can be researched (its prereq Language is done)
	ge.GrantTechsForTest("language")
	if got := altar(); !reflect.DeepEqual(got.LockTechNext, []string{"Ritual"}) {
		t.Fatalf("after Language: next=%v, want [Ritual]", got.LockTechNext)
	}

	// After Ritual, the path is empty so LockTechNext is empty
	ge.GrantTechsForTest("ritual")
	if got := altar(); got.LockTechPath != nil || got.LockTechNext != nil {
		t.Fatalf("after Ritual: path=%v next=%v, want both empty", got.LockTechPath, got.LockTechNext)
	}

	// After Calendar, no lock at all
	ge.GrantTechsForTest("calendar")
	if got := altar(); got.LockTech != "" || got.LockTechNext != nil {
		t.Fatalf("after Calendar: tech=%q next=%v, want no lock", got.LockTech, got.LockTechNext)
	}
}
