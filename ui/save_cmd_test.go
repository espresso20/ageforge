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

// TestCmdLoadNoArgsDoesNotLoad verifies a bare `load` no longer silently loads
// the "autosave" slot. Instead it returns an informational usage message and
// leaves the engine's active save untouched — the dashboard intercepts the bare
// command to open the browser, so this fallback must never load a slot by
// surprise.
func TestCmdLoadNoArgsDoesNotLoad(t *testing.T) {
	name := "Cmd Load Current"
	t.Cleanup(func() {
		_ = game.DeleteSave(name)
	})

	engine := game.NewGameEngine()
	if err := engine.StartNewNamedGame(name); err != nil {
		t.Fatalf("StartNewNamedGame(%q) failed: %v", name, err)
	}
	wantActive := engine.ActiveSaveName()

	res := cmdLoad(nil, engine)
	if res.Type != "info" {
		t.Errorf("bare cmdLoad type = %q, want %q", res.Type, "info")
	}
	if !strings.Contains(res.Message, "load <name>") {
		t.Errorf("bare cmdLoad message = %q, want it to mention the 'load <name>' usage", res.Message)
	}
	if strings.HasPrefix(res.Message, "Game loaded") {
		t.Errorf("bare cmdLoad returned a success-style message %q; it must not load a slot", res.Message)
	}
	// The active slot must be unchanged — no load happened.
	if got := engine.ActiveSaveName(); got != wantActive {
		t.Errorf("bare cmdLoad changed active slot to %q, want %q (unchanged)", got, wantActive)
	}
}

// TestCmdLoadWithNameLoads verifies that `load <name>` still loads the named
// save directly through the engine.
func TestCmdLoadWithNameLoads(t *testing.T) {
	current := "Cmd Load Running"
	target := "Cmd Load Target"
	t.Cleanup(func() {
		_ = game.DeleteSave(current)
		_ = game.DeleteSave(target)
	})

	engine := game.NewGameEngine()
	if err := engine.StartNewNamedGame(current); err != nil {
		t.Fatalf("StartNewNamedGame(%q) failed: %v", current, err)
	}
	// Branch a second save so there is a distinct slot to load.
	if res := cmdSave([]string{target}, engine); res.Type == "error" {
		t.Fatalf("cmdSave(%q) returned error: %s", target, res.Message)
	}
	// Switch the active slot back so loading the target is an observable change.
	if res := cmdLoad([]string{current}, engine); res.Type == "error" {
		t.Fatalf("cmdLoad(%q) returned error: %s", current, res.Message)
	}
	if engine.ActiveSaveName() != current {
		t.Fatalf("after load %q, ActiveSaveName() = %q, want %q", current, engine.ActiveSaveName(), current)
	}

	res := cmdLoad([]string{target}, engine)
	if res.Type == "error" {
		t.Fatalf("cmdLoad(%q) returned error: %s", target, res.Message)
	}
	if !strings.Contains(res.Message, target) {
		t.Errorf("cmdLoad(%q) message = %q, want it to name the loaded save", target, res.Message)
	}
	if engine.ActiveSaveName() != target {
		t.Errorf("after load %q, ActiveSaveName() = %q, want %q", target, engine.ActiveSaveName(), target)
	}
}
