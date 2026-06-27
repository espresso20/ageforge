package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestCmdSaveBranches verifies the explicit `save <name>` path forks a new save:
// it branches off the active run, switches the active slot to the new name, and
// records the old slot as the parent. A second branch to the same name is
// rejected with an error message. The bare `save` (no args) path overwrites the
// active slot without branching. Save-dir isolation/cleanup mirrors the
// game-package save tests.
func TestCmdSaveBranches(t *testing.T) {
	parent := "Cmd Save Parent"
	child := "Cmd Save Child"
	t.Cleanup(func() {
		_ = game.DeleteSave(parent)
		_ = game.DeleteSave(child)
	})

	engine := game.NewGameEngine()
	if err := engine.StartNewNamedGame(parent); err != nil {
		t.Fatalf("StartNewNamedGame(%q) failed: %v", parent, err)
	}

	// save <name> → branch.
	res := cmdSave([]string{child}, engine)
	if res.Type == "error" {
		t.Fatalf("cmdSave(%q) returned error: %s", child, res.Message)
	}
	if engine.ActiveSaveName() != child {
		t.Errorf("after branch, ActiveSaveName() = %q, want %q", engine.ActiveSaveName(), child)
	}
	if engine.ActiveParentName() != parent {
		t.Errorf("after branch, ActiveParentName() = %q, want %q", engine.ActiveParentName(), parent)
	}
	if !game.SaveExists(child) {
		t.Errorf("branch save %q not written", child)
	}

	// save <existing-name> → error (already exists).
	res = cmdSave([]string{child}, engine)
	if res.Type != "error" {
		t.Errorf("cmdSave to existing name = %q type, want error", res.Type)
	}
	if !strings.Contains(res.Message, "already exists") {
		t.Errorf("cmdSave to existing name message = %q, want it to mention 'already exists'", res.Message)
	}

	// Bare save (no args) → overwrite the active slot (now the child), no branch.
	res = cmdSave(nil, engine)
	if res.Type == "error" {
		t.Fatalf("bare cmdSave returned error: %s", res.Message)
	}
	if engine.ActiveSaveName() != child {
		t.Errorf("bare save changed active slot to %q, want %q", engine.ActiveSaveName(), child)
	}
}
