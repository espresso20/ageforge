package game

import "testing"

// noErr fails the test at the caller's line when err is not nil.
func noErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// saveAndLoad saves ge and loads the save into a new engine: a second load.
func saveAndLoad(t *testing.T, ge *GameEngine, name string) *GameEngine {
	t.Helper()
	noErr(t, ge.SaveGame(name))
	again := NewGameEngine()
	noErr(t, again.LoadGame(name))
	if again.cheaterBadge {
		t.Fatalf("the re-saved %s failed its signature check", name)
	}
	return again
}
